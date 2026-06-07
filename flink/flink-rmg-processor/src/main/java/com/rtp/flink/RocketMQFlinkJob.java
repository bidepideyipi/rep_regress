package com.rtp.flink;

import com.rtp.flink.aggregator.RTPAggregateFunction;
import com.rtp.flink.clickhouse.ClickHouseSink;
import com.rtp.flink.clickhouse.GameLogDetailSink;
import com.rtp.flink.config.AppConfig;
import com.rtp.flink.config.NacosConfigLoader;
import com.rtp.flink.deserializer.GameLogDeserializer;
import com.rtp.flink.function.RTPWindowFunction;
import com.rtp.flink.model.GameLogDetail;
import com.rtp.flink.model.RTPStatistics;
import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.java.tuple.Tuple3;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.windowing.assigners.TumblingEventTimeWindows;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.streaming.connectors.rocketmq.FlinkRocketMQConsumer;
import org.apache.flink.streaming.connectors.rocketmq.config.ConsumerConfig;
import org.apache.flink.streaming.connectors.rocketmq.config.RocketMQConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Properties;

/**
 * Flink + RocketMQ RTP计算作业
 *
 * 数据流处理：
 * 1. 消费 RocketMQ 游戏日志
 * 2. 分支1：直接写入 game_log_detail_local（明细）
 * 3. 分支2：窗口聚合后写入 rtp_statistics_local（统计）
 *
 * Nacos配置格式（与rtp-processor统一）：
 * <pre>
 * {
 *   "clickhouse": {...},
 *   "rocket_mq": {
 *     "name_servers": ["127.0.0.1:9876"],
 *     "consumer": {
 *       "group_name": "flink_rtp_consumer_group",
 *       "topic": "game_log_topic",
 *       "batch_size": 100
 *     }
 *   }
 * }
 * </pre>
 */
public class RocketMQFlinkJob {

    private static final Logger LOG = LoggerFactory.getLogger(RocketMQFlinkJob.class);

    private static AppConfig appConfig;
    private static ClickHouseSink rtpSink;
    private static GameLogDetailSink gameLogSink;

    // 窗口配置（秒）
    private static final int GLOBAL_RTP_WINDOW = 60;
    private static final int TERMINAL_RTP_WINDOW = 60;
    private static final int PLAYER_RTP_WINDOW = 300;
    private static final int PARALLELISM = 2;
    private static final long CHECKPOINT_INTERVAL = 60000;

    public static void main(String[] args) {
        try {
            loadConfig();

            LOG.info("========== Flink RTP计算作业启动 ==========");
            LOG.info("RocketMQ: {}", appConfig.getRocketMq().getNameServersStr());
            LOG.info("Topic: {}", appConfig.getRocketMq().getConsumer().getTopic());
            LOG.info("ClickHouse: {}/{}", appConfig.getClickhouse().getHost(), appConfig.getClickhouse().getDatabase());

            StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();
            env.enableCheckpointing(CHECKPOINT_INTERVAL);
            env.setParallelism(PARALLELISM);

            // 创建RocketMQ数据源
            DataStream<GameLogDetail> gameLogStream = env
                    .addSource(createRocketMQSource())
                    .name("RocketMQ-Source")
                    .uid("rocketmq-source");

            // 分支1：写入游戏日志明细
            gameLogStream
                    .filter(log -> log != null)
                    .addSink(gameLogSink)
                    .name("GameLog-Sink")
                    .uid("gamelog-sink");

            // 分支2：RTP统计计算
            calculateGlobalRTP(gameLogStream);
            calculateTerminalRTP(gameLogStream);
            calculatePlayerRTP(gameLogStream);

            // 关闭钩子
            Runtime.getRuntime().addShutdownHook(new Thread(() -> {
                LOG.info("关闭Sink...");
                if (gameLogSink != null) gameLogSink.close();
                if (rtpSink != null) rtpSink.close();
            }));

            env.execute("Flink-RTP-Calculation");

        } catch (Exception e) {
            LOG.error("作业启动失败: {}", e.getMessage(), e);
            System.exit(1);
        }
    }

    private static void loadConfig() {
        String serverAddr = System.getProperty("nacos.serverAddr", "127.0.0.1:8848");
        String namespace = System.getProperty("nacos.namespace", "");
        String dataId = System.getProperty("nacos.dataId", "app-config");
        String group = System.getProperty("nacos.group", "DEFAULT_GROUP");

        try {
            NacosConfigLoader loader = new NacosConfigLoader(serverAddr, namespace);
            appConfig = loader.loadConfig(dataId, group);
            loader.close();

            // 初始化Sink
            int batchSize = appConfig.getRocketMq().getConsumer().getBatchSize();
            gameLogSink = new GameLogDetailSink(appConfig.getClickhouse(), batchSize);
            rtpSink = new ClickHouseSink(appConfig.getClickhouse(), batchSize);

        } catch (Exception e) {
            LOG.error("加载Nacos配置失败: {}", e.getMessage(), e);
            throw new RuntimeException("请确保Nacos已启动并配置正确", e);
        }
    }

    private static FlinkRocketMQConsumer<GameLogDetail> createRocketMQSource() {
        Properties props = new Properties();
        props.setProperty(RocketMQConfig.NAME_SERVER_ADDR, appConfig.getRocketMq().getNameServersStr());

        var consumerConfig = appConfig.getRocketMq().getConsumer();

        ConsumerConfig config = new ConsumerConfig.Builder()
                .setConsumerGroup(consumerConfig.getGroupName())
                .setTopic(consumerConfig.getTopic())
                .setMessageModel(ConsumerConfig.MessageModel.CLUSTERING)
                .setPullBatchSize(consumerConfig.getBatchSize())
                .setPullInterval(0)
                .build();

        FlinkRocketMQConsumer<GameLogDetail> source = new FlinkRocketMQConsumer<>(
                new GameLogDeserializer(), config, props
        );
        source.setParallelism(PARALLELISM);
        return source;
    }

    private static void calculateGlobalRTP(DataStream<GameLogDetail> stream) {
        stream.filter(log -> log != null && log.getBetAmount() != null
                        && log.getBetAmount().compareTo(java.math.BigDecimal.ZERO) > 0)
                .keyBy(log -> new Tuple3<>("global", "all",
                        log.getIntegratorId() != null ? log.getIntegratorId() : "default"))
                .window(TumblingEventTimeWindows.of(Time.seconds(GLOBAL_RTP_WINDOW)))
                .aggregate(new RTPAggregateFunction(), new RTPWindowFunction())
                .addSink(rtpSink)
                .name("Global-RTP");
    }

    private static void calculateTerminalRTP(DataStream<GameLogDetail> stream) {
        stream.filter(log -> log != null && log.getBetAmount() != null
                        && log.getBetAmount().compareTo(java.math.BigDecimal.ZERO) > 0
                        && log.getTerminal() != null && !log.getTerminal().isEmpty())
                .keyBy(log -> new Tuple3<>("terminal", log.getTerminal(),
                        log.getIntegratorId() != null ? log.getIntegratorId() : "default"))
                .window(TumblingEventTimeWindows.of(Time.seconds(TERMINAL_RTP_WINDOW)))
                .aggregate(new RTPAggregateFunction(), new RTPWindowFunction())
                .addSink(rtpSink)
                .name("Terminal-RTP");
    }

    private static void calculatePlayerRTP(DataStream<GameLogDetail> stream) {
        stream.filter(log -> log != null && log.getBetAmount() != null
                        && log.getBetAmount().compareTo(java.math.BigDecimal.ZERO) > 0
                        && log.getPlayerId() != null && !log.getPlayerId().isEmpty())
                .keyBy(log -> new Tuple3<>("player", log.getPlayerId(),
                        log.getIntegratorId() != null ? log.getIntegratorId() : "default"))
                .window(TumblingEventTimeWindows.of(Time.seconds(PLAYER_RTP_WINDOW)))
                .aggregate(new RTPAggregateFunction(), new RTPWindowFunction())
                .addSink(rtpSink)
                .name("Player-RTP");
    }
}
