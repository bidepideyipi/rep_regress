package com.rtp.flink.clickhouse;

import com.alibaba.fastjson2.JSON;
import com.rtp.flink.config.AppConfig;
import com.rtp.flink.model.RTPStatistics;
import org.apache.flink.streaming.api.functions.sink.SinkFunction;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicLong;

/**
 * ClickHouse RTP统计结果写入Sink
 * 将Flink计算出的RTP统计结果批量写入ClickHouse
 *
 * 参考 rtp-processor 中的 ClickHouse 批量写入实现：
 * - 使用HTTP接口写入
 * - JSONEachRow格式
 * - 支持批量写入
 */
public class ClickHouseSink implements SinkFunction<RTPStatistics> {

    private static final Logger LOG = LoggerFactory.getLogger(ClickHouseSink.class);

    private final AppConfig.ClickHouseConfig chConfig;
    private final CloseableHttpClient httpClient;

    /** RTP统计表名 - 与rtp-processor保持一致 */
    private static final String RTP_STATS_TABLE = "rtp_statistics_local";

    // 批量写入缓冲区
    private final List<RTPStatistics> buffer = new ArrayList<>();
    private final int batchSize;

    // 统计信息
    private final AtomicLong successCount = new AtomicLong(0);
    private final AtomicLong errorCount = new AtomicLong(0);

    /**
     * 构造函数
     */
    public ClickHouseSink(AppConfig.ClickHouseConfig chConfig, int batchSize) {
        this.chConfig = chConfig;
        this.batchSize = batchSize;

        this.httpClient = HttpClients.createDefault();

        // 初始化表结构
        initTable();
    }

    /**
     * 初始化ClickHouse表
     */
    private void initTable() {
        String createTableSQL = String.format(
                "CREATE TABLE IF NOT EXISTS %s (" +
                        "dimension_type String, " +
                        "dimension_key String, " +
                        "integrator_id String, " +
                        "game_id Nullable(Int32), " +
                        "total_bet_amount Decimal(18,2), " +
                        "total_win_amount Decimal(18,2), " +
                        "game_count UInt64, " +
                        "rtp Decimal(10,2), " +
                        "window_start DateTime64(3), " +
                        "window_end DateTime64(3), " +
                        "statistics_time DateTime64(3)" +
                        ") ENGINE = MergeTree() " +
                        "PARTITION BY toYYYYMM(window_start) " +
                        "ORDER BY (dimension_type, dimension_key, window_start)",
                RTP_STATS_TABLE
        );

        try {
            execSQL(createTableSQL);
            LOG.info("[ClickHouse] 表初始化成功: {}", rtpStatsTable);
        } catch (Exception e) {
            LOG.error("[ClickHouse] 表初始化失败: {}", e.getMessage(), e);
        }
    }

    @Override
    public void invoke(RTPStatistics statistics, Context context) throws Exception {
        synchronized (buffer) {
            buffer.add(statistics);

            // 达到批次大小时写入
            if (buffer.size() >= batchSize) {
                flush();
            }
        }
    }

    /**
     * 刷新缓冲区，将数据写入ClickHouse
     */
    private void flush() {
        if (buffer.isEmpty()) {
            return;
        }

        try {
            batchInsert(new ArrayList<>(buffer));
            successCount.addAndGet(buffer.size());
            LOG.info("[ClickHouse] 批量写入成功: {} 条", buffer.size());
            buffer.clear();
        } catch (Exception e) {
            errorCount.addAndGet(buffer.size());
            LOG.error("[ClickHouse] 批量写入失败: {}", e.getMessage(), e);
            buffer.clear();
        }
    }

    /**
     * 批量插入数据
     */
    private void batchInsert(List<RTPStatistics> stats) throws Exception {
        if (stats.isEmpty()) {
            return;
        }

        // 构建JSONEachRow格式的数据
        List<String> rows = new ArrayList<>();
        for (RTPStatistics stat : stats) {
            rows.add(JSON.toJSONString(stat));
        }

        String body = String.join("\n", rows);

        // 构建请求URL
        String url = String.format("%s/?database=%s&query=INSERT INTO %s FORMAT JSONEachRow",
                chConfig.getHttpUrl(),
                chConfig.getDatabase(),
                RTP_STATS_TABLE
        );

        HttpPost httpPost = new HttpPost(url);
        httpPost.setHeader("Content-Type", "application/json");
        httpPost.setEntity(new StringEntity(body, StandardCharsets.UTF_8));

        // 认证
        if (chConfig.getPassword() != null && !chConfig.getPassword().isEmpty()) {
            setBasicAuth(httpPost);
        }

        // 执行请求
        httpClient.execute(httpPost, response -> {
            int statusCode = response.getStatusLine().getStatusCode();
            String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);

            if (statusCode != 200) {
                throw new RuntimeException(String.format(
                        "ClickHouse写入失败: HTTP %d, %s", statusCode, responseBody
                ));
            }
            return null;
        });
    }

    /**
     * 执行SQL语句
     */
    private void execSQL(String sql) throws Exception {
        String url = String.format("%s/?database=%s&query=%s",
                chConfig.getHttpUrl(),
                chConfig.getDatabase(),
                java.net.URLEncoder.encode(sql, StandardCharsets.UTF_8)
        );

        HttpGet httpGet = new HttpGet(url);

        // 认证
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

    /**
     * 设置Basic认证
     */
    private void setBasicAuth(org.apache.http.client.methods.HttpRequestBase request) {
        String auth = chConfig.getUsername() + ":" + chConfig.getPassword();
        String encodedAuth = java.util.Base64.getEncoder().encodeToString(
                auth.getBytes(StandardCharsets.UTF_8)
        );
        request.setHeader("Authorization", "Basic " + encodedAuth);
    }

    /**
     * 关闭Sink
     */
    public void close() {
        // 刷新剩余数据
        synchronized (buffer) {
            if (!buffer.isEmpty()) {
                try {
                    flush();
                } catch (Exception e) {
                    LOG.error("[ClickHouse] 关闭前刷新失败: {}", e.getMessage(), e);
                }
            }
        }

        // 关闭HTTP客户端
        try {
            httpClient.close();
        } catch (Exception e) {
            LOG.error("[ClickHouse] 关闭HTTP客户端失败: {}", e.getMessage(), e);
        }

        // 输出统计信息
        LOG.info("[ClickHouse] 写入统计 - 成功: {}, 失败: {}",
                successCount.get(), errorCount.get());
    }

    /**
     * 获取统计信息
     */
    public long getSuccessCount() {
        return successCount.get();
    }

    public long getErrorCount() {
        return errorCount.get();
    }
}
