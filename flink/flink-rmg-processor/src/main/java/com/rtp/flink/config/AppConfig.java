package com.rtp.flink.config;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;

import java.util.List;

/**
 * 应用配置类
 * 对应 rtp-processor 中的 Config 结构
 * 支持 Nacos 配置中心的动态加载
 *
 * Nacos 配置格式（与 rtp-processor 统一）：
 * <pre>
 * {
 *   "clickhouse": {
 *     "host": "127.0.0.1",
 *     "port": 8123,
 *     "username": "default",
 *     "password": "",
 *     "database": "rtp_game"
 *   },
 *   "rocket_mq": {
 *     "name_servers": ["127.0.0.1:9876"],
 *     "producer": {
 *       "group_name": "game_producer_group",
 *       "topic": "game_log_topic"
 *     },
 *     "consumer": {
 *       "group_name": "flink_rtp_consumer_group",
 *       "topic": "game_log_topic",
 *       "batch_size": 100
 *     }
 *   },
 *   "mysql": {
 *     "host": "127.0.0.1",
 *     "port": 3306,
 *     "username": "root",
 *     "password": "",
 *     "database": "rtp_game"
 *   },
 *   "aggregate_user_interval": "5m",
 *   "aggregate_game_interval": "5m",
 *   "alert_interval": "10m"
 * }
 * </pre>
 */
@Data
public class AppConfig {

    /**
     * ClickHouse 配置
     */
    @JSONField(name = "clickhouse")
    private ClickHouseConfig clickhouse = new ClickHouseConfig();

    /**
     * RocketMQ 配置
     */
    @JSONField(name = "rocket_mq")
    private RocketMQConfig rocketMq = new RocketMQConfig();

    /**
     * MySQL 配置
     */
    @JSONField(name = "mysql")
    private MySQLConfig mysql = new MySQLConfig();

    /**
     * 用户聚合间隔
     */
    @JSONField(name = "aggregate_user_interval")
    private String aggregateUserInterval = "5m";

    /**
     * 游戏聚合间隔
     */
    @JSONField(name = "aggregate_game_interval")
    private String aggregateGameInterval = "5m";

    /**
     * 告警间隔
     */
    @JSONField(name = "alert_interval")
    private String alertInterval = "10m";

    /**
     * ClickHouse 配置
     */
    @Data
    public static class ClickHouseConfig {
        @JSONField(name = "host")
        private String host = "127.0.0.1";

        @JSONField(name = "port")
        private int port = 8123;

        @JSONField(name = "username")
        private String username = "default";

        @JSONField(name = "password")
        private String password = "";

        @JSONField(name = "database")
        private String database = "rtp_game";

        /** JDBC URL */
        public String getJdbcUrl() {
            return String.format("jdbc:clickhouse://%s:%d/%s", host, port, database);
        }

        /** HTTP URL */
        public String getHttpUrl() {
            return String.format("http://%s:%d", host, port);
        }
    }

    /**
     * RocketMQ 配置
     */
    @Data
    public static class RocketMQConfig {
        @JSONField(name = "name_servers")
        private List<String> nameServers;

        @JSONField(name = "producer")
        private ProducerConfig producer = new ProducerConfig();

        @JSONField(name = "consumer")
        private ConsumerConfig consumer = new ConsumerConfig();

        /** 获取NameServer地址字符串 */
        public String getNameServersStr() {
            if (nameServers == null || nameServers.isEmpty()) {
                return "127.0.0.1:9876";
            }
            return String.join(";", nameServers);
        }

        @Data
        public static class ProducerConfig {
            @JSONField(name = "group_name")
            private String groupName = "game_producer_group";

            @JSONField(name = "topic")
            private String topic = "game_log_topic";
        }

        @Data
        public static class ConsumerConfig {
            @JSONField(name = "group_name")
            private String groupName = "flink_rtp_consumer_group";

            @JSONField(name = "topic")
            private String topic = "game_log_topic";

            @JSONField(name = "batch_size")
            private int batchSize = 100;
        }
    }

    /**
     * MySQL 配置
     */
    @Data
    public static class MySQLConfig {
        @JSONField(name = "host")
        private String host = "127.0.0.1";

        @JSONField(name = "port")
        private int port = 3306;

        @JSONField(name = "username")
        private String username = "root";

        @JSONField(name = "password")
        private String password = "";

        @JSONField(name = "database")
        private String database = "rtp_game";

        /** JDBC URL */
        public String getJdbcUrl() {
            return String.format("jdbc:mysql://%s:%d/%s", host, port, database);
        }
    }

    /**
     * 获取用户聚合间隔（秒）
     */
    public int getAggregateUserIntervalSeconds() {
        return parseIntervalToSeconds(aggregateUserInterval, 300);
    }

    /**
     * 获取游戏聚合间隔（秒）
     */
    public int getAggregateGameIntervalSeconds() {
        return parseIntervalToSeconds(aggregateGameInterval, 300);
    }

    /**
     * 获取告警间隔（秒）
     */
    public int getAlertIntervalSeconds() {
        return parseIntervalToSeconds(alertInterval, 600);
    }

    /**
     * 解析间隔字符串为秒数
     * 支持：30s, 5m, 1h 等
     */
    private int parseIntervalToSeconds(String interval, int defaultSeconds) {
        if (interval == null || interval.isEmpty()) {
            return defaultSeconds;
        }
        try {
            if (interval.endsWith("s")) {
                return Integer.parseInt(interval.substring(0, interval.length() - 1));
            } else if (interval.endsWith("m")) {
                return Integer.parseInt(interval.substring(0, interval.length() - 1)) * 60;
            } else if (interval.endsWith("h")) {
                return Integer.parseInt(interval.substring(0, interval.length() - 1)) * 3600;
            }
            return Integer.parseInt(interval);
        } catch (NumberFormatException e) {
            return defaultSeconds;
        }
    }
}
