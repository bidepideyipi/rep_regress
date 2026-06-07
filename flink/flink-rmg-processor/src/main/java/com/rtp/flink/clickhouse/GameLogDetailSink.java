package com.rtp.flink.clickhouse;

import com.alibaba.fastjson2.JSON;
import com.rtp.flink.config.AppConfig;
import com.rtp.flink.model.GameLogDetail;
import org.apache.flink.streaming.api.functions.sink.SinkFunction;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicLong;

/**
 * 游戏日志明细写入 Sink
 * 将原始游戏日志批量写入 ClickHouse game_log_detail_local 表
 *
 * 参考 rtp-processor 中的 ClickHouse 批量写入实现
 */
public class GameLogDetailSink implements SinkFunction<GameLogDetail> {

    private static final Logger LOG = LoggerFactory.getLogger(GameLogDetailSink.class);
    private static final String GAME_LOG_TABLE = "game_log_detail_local";

    private final AppConfig.ClickHouseConfig chConfig;
    private final CloseableHttpClient httpClient;

    // 批量写入缓冲区
    private final List<GameLogDetail> buffer = new ArrayList<>();
    private final int batchSize;

    // 统计信息
    private final AtomicLong successCount = new AtomicLong(0);
    private final AtomicLong errorCount = new AtomicLong(0);

    public GameLogDetailSink(AppConfig.ClickHouseConfig chConfig, int batchSize) {
        this.chConfig = chConfig;
        this.batchSize = batchSize;
        this.httpClient = HttpClients.createDefault();

        initTable();
    }

    /**
     * 初始化 ClickHouse 表
     */
    private void initTable() {
        String createTableSQL = String.format(
                "CREATE TABLE IF NOT EXISTS %s (" +
                        "id UInt64, " +
                        "game_id String, " +
                        "integrator_id String, " +
                        "player_id String, " +
                        "terminal String, " +
                        "game_id_int Nullable(Int32), " +
                        "room_id Nullable(Int32), " +
                        "bet_amount Decimal(18,2), " +
                        "win_amount Decimal(18,2), " +
                        "rtp Nullable(Decimal(10,2)), " +
                        "status String, " +
                        "result_data Nullable(String), " +
                        "created_at DateTime64(3), " +
                        "updated_at DateTime64(3), " +
                        "ext1 Nullable(String), " +
                        "ext2 Nullable(String)" +
                        ") ENGINE = MergeTree() " +
                        "PARTITION BY toYYYYMM(created_at) " +
                        "ORDER BY (game_id, created_at)",
                GAME_LOG_TABLE
        );

        try {
            execSQL(createTableSQL);
            LOG.info("[ClickHouse] 游戏日志表初始化成功: {}", GAME_LOG_TABLE);
        } catch (Exception e) {
            LOG.error("[ClickHouse] 游戏日志表初始化失败: {}", e.getMessage(), e);
        }
    }

    @Override
    public void invoke(GameLogDetail log, Context context) throws Exception {
        synchronized (buffer) {
            buffer.add(log);

            if (buffer.size() >= batchSize) {
                flush();
            }
        }
    }

    private void flush() {
        if (buffer.isEmpty()) {
            return;
        }

        try {
            batchInsert(new ArrayList<>(buffer));
            successCount.addAndGet(buffer.size());
            LOG.debug("[ClickHouse] 游戏日志写入成功: {} 条", buffer.size());
            buffer.clear();
        } catch (Exception e) {
            errorCount.addAndGet(buffer.size());
            LOG.error("[ClickHouse] 游戏日志写入失败: {}", e.getMessage(), e);
            buffer.clear();
        }
    }

    private void batchInsert(List<GameLogDetail> logs) throws Exception {
        if (logs.isEmpty()) {
            return;
        }

        List<String> rows = new ArrayList<>();
        for (GameLogDetail log : logs) {
            rows.add(JSON.toJSONString(log));
        }

        String body = String.join("\n", rows);

        String url = String.format("%s/?database=%s&query=INSERT INTO %s FORMAT JSONEachRow",
                chConfig.getHttpUrl(),
                chConfig.getDatabase(),
                GAME_LOG_TABLE
        );

        HttpPost httpPost = new HttpPost(url);
        httpPost.setHeader("Content-Type", "application/json");
        httpPost.setEntity(new StringEntity(body, StandardCharsets.UTF_8));

        if (chConfig.getPassword() != null && !chConfig.getPassword().isEmpty()) {
            setBasicAuth(httpPost);
        }

        httpClient.execute(httpPost, response -> {
            int statusCode = response.getStatusLine().getStatusCode();
            if (statusCode != 200) {
                String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                throw new RuntimeException(String.format(
                        "ClickHouse写入失败: HTTP %d, %s", statusCode, responseBody
                ));
            }
            return null;
        });
    }

    private void execSQL(String sql) throws Exception {
        String url = String.format("%s/?database=%s&query=%s",
                chConfig.getHttpUrl(),
                chConfig.getDatabase(),
                java.net.URLEncoder.encode(sql, StandardCharsets.UTF_8)
        );

        org.apache.http.client.methods.HttpGet httpGet = new org.apache.http.client.methods.HttpGet(url);

        if (chConfig.getPassword() != null && !chConfig.getPassword().isEmpty()) {
            setBasicAuth(httpGet);
        }

        httpClient.execute(httpGet, response -> {
            int statusCode = response.getStatusLine().getStatusCode();
            if (statusCode != 200) {
                String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                throw new RuntimeException(String.format(
                        "ClickHouse执行SQL失败: HTTP %d, %s", statusCode, responseBody
                ));
            }
            return null;
        });
    }

    private void setBasicAuth(org.apache.http.client.methods.HttpRequestBase request) {
        String auth = chConfig.getUsername() + ":" + chConfig.getPassword();
        String encodedAuth = java.util.Base64.getEncoder().encodeToString(
                auth.getBytes(StandardCharsets.UTF_8)
        );
        request.setHeader("Authorization", "Basic " + encodedAuth);
    }

    public void close() {
        synchronized (buffer) {
            if (!buffer.isEmpty()) {
                try {
                    flush();
                } catch (Exception e) {
                    LOG.error("[ClickHouse] 关闭前刷新失败: {}", e.getMessage(), e);
                }
            }
        }

        try {
            httpClient.close();
        } catch (Exception e) {
            LOG.error("[ClickHouse] 关闭HTTP客户端失败: {}", e.getMessage(), e);
        }

        LOG.info("[ClickHouse] 游戏日志写入统计 - 成功: {}, 失败: {}",
                successCount.get(), errorCount.get());
    }
}
