package com.rtp.flink.util;

import com.alibaba.fastjson2.JSON;
import com.rtp.flink.model.GameLogDetail;
import org.apache.rocketmq.client.producer.DefaultMQProducer;
import org.apache.rocketmq.common.message.Message;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.math.BigDecimal;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

/**
 * 测试消息生成器
 * 用于生成游戏日志并发送到RocketMQ，测试Flink作业
 */
public class TestMessageGenerator {

    private static final Logger LOG = LoggerFactory.getLogger(TestMessageGenerator.class);

    private static final String NAME_SERVER_ADDRESS = "127.0.0.1:9876";
    private static final String PRODUCER_GROUP = "test_producer_group";
    private static final String TOPIC = "game_log_topic";

    public static void main(String[] args) {
        try {
            // 创建生产者
            DefaultMQProducer producer = new DefaultMQProducer(PRODUCER_GROUP);
            producer.setNamesrvAddr(NAME_SERVER_ADDRESS);
            producer.start();

            LOG.info("测试消息生成器已启动，开始发送消息...");

            // 模拟终端类型
            String[] terminals = {"web", "mobile", "app"};
            // 模拟玩家ID
            String[] playerIds = {"player-001", "player-002", "player-003", "player-004", "player-005"};
            // 集成商ID
            String integratorId = "integrator-001";

            // 持续发送消息
            int messageCount = 0;
            long lastReportTime = System.currentTimeMillis();

            while (true) {
                try {
                    // 随机生成游戏日志
                    GameLogDetail log = generateRandomLog(integratorId, terminals, playerIds);
                    String json = JSON.toJSONString(log);

                    Message message = new Message(TOPIC, json.getBytes());
                    producer.send(message);

                    messageCount++;

                    // 每100条消息报告一次
                    if (messageCount % 100 == 0) {
                        long elapsed = System.currentTimeMillis() - lastReportTime;
                        LOG.info("已发送 {} 条消息，耗时 {} ms", messageCount, elapsed);
                        lastReportTime = System.currentTimeMillis();
                    }

                    // 控制发送频率
                    Thread.sleep(100);
                } catch (Exception e) {
                    LOG.error("发送消息失败", e);
                    Thread.sleep(1000);
                }
            }
        } catch (Exception e) {
            LOG.error("测试消息生成器启动失败", e);
        }
    }

    /**
     * 生成随机游戏日志
     * 模拟真实的游戏结算数据
     */
    private static GameLogDetail generateRandomLog(String integratorId, String[] terminals, String[] playerIds) {
        // 随机选择终端和玩家
        String terminal = terminals[(int) (Math.random() * terminals.length)];
        String playerId = playerIds[(int) (Math.random() * playerIds.length)];

        // 生成下注金额（10-1000）
        BigDecimal betAmount = new BigDecimal(10 + (int) (Math.random() * 990));

        // 计算赔付金额（模拟RTP在85%-98%之间）
        double rtp = 0.85 + (Math.random() * 0.13);  // 85% - 98%
        BigDecimal winAmount = betAmount.multiply(new BigDecimal(rtp))
                .setScale(2, java.math.RoundingMode.HALF_UP);

        return GameLogDetail.builder()
                .gameId(UUID.randomUUID().toString().substring(0, 20))
                .integratorId(integratorId)
                .playerId(playerId)
                .terminal(terminal)
                .gameIdInt(1)
                .roomId(1)
                .betAmount(betAmount)
                .winAmount(winAmount)
                .rtp(winAmount.divide(betAmount, 4, java.math.RoundingMode.HALF_UP)
                        .multiply(new BigDecimal("100"))
                        .setScale(2, java.math.RoundingMode.HALF_UP))
                .status("completed")
                .createdAt(System.currentTimeMillis())
                .updatedAt(System.currentTimeMillis())
                .build();
    }
}
