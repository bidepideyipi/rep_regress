package com.rtp.flink.sink;

import com.rtp.flink.model.RTPStatistics;
import org.apache.flink.streaming.api.functions.sink.SinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * RTP统计结果输出Sink
 * 将计算结果输出到控制台，后续可扩展为写入ClickHouse
 *
 * 参考 rtp-processor 中的 ClickHouse 批量写入逻辑
 */
public class RTPStatisticsSink implements SinkFunction<RTPStatistics> {

    private static final Logger LOG = LoggerFactory.getLogger(RTPStatisticsSink.class);

    @Override
    public void invoke(RTPStatistics statistics, Context context) throws Exception {
        // 格式化输出RTP统计结果
        LOG.info("========== RTP统计结果 ==========");
        LOG.info("维度类型: {}", statistics.getDimensionType());
        LOG.info("维度键值: {}", statistics.getDimensionKey());
        LOG.info("集成商ID: {}", statistics.getIntegratorId());
        LOG.info("游戏ID: {}", statistics.getGameId());
        LOG.info("游戏次数: {}", statistics.getGameCount());
        LOG.info("总下注: {}", statistics.getTotalBetAmount());
        LOG.info("总赔付: {}", statistics.getTotalWinAmount());
        LOG.info("RTP百分比: {}%", statistics.getRtp());
        LOG.info("窗口时间: {} - {}", formatTime(statistics.getWindowStart()), formatTime(statistics.getWindowEnd()));
        LOG.info("==================================");
    }

    /**
     * 格式化时间戳为可读字符串
     */
    private String formatTime(Long timestamp) {
        if (timestamp == null) {
            return "N/A";
        }
        java.time.Instant instant = java.time.Instant.ofEpochMilli(timestamp);
        return instant.toString();
    }
}
