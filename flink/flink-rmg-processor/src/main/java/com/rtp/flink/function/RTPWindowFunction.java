package com.rtp.flink.function;

import com.rtp.flink.model.GameLogDetail;
import com.rtp.flink.model.RTPStatistics;
import org.apache.flink.api.java.tuple.Tuple3;
import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.util.Collector;

import java.math.BigDecimal;

/**
 * RTP窗口处理函数
 * 支持全局、终端、玩家三个维度的RTP计算
 *
 * 窗口键: Tuple3<维度类型, 维度键, 集成商ID>
 * - 维度类型: "global"(全局), "terminal"(终端), "player"(玩家)
 * - 维度键: 终端ID或玩家ID，全局维度为"all"
 *
 * 参考 rtp-processor 中的批量处理逻辑和 ClickHouse 写入
 */
public class RTPWindowFunction extends ProcessWindowFunction<
        GameLogDetail,
        RTPStatistics,
        Tuple3<String, String, String>,
        TimeWindow> {

    @Override
    public void process(
            Tuple3<String, String, String> key,
            Context context,
            Iterable<GameLogDetail> logs,
            Collector<RTPStatistics> out) throws Exception {

        String dimensionType = key.f0;
        String dimensionKey = key.f1;
        String integratorId = key.f2;

        // 初始化统计数据
        BigDecimal totalBetAmount = BigDecimal.ZERO;
        BigDecimal totalWinAmount = BigDecimal.ZERO;
        long gameCount = 0L;

        // 遍历窗口内所有日志进行聚合
        for (GameLogDetail log : logs) {
            if (log.getBetAmount() != null) {
                totalBetAmount = totalBetAmount.add(log.getBetAmount());
            }
            if (log.getWinAmount() != null) {
                totalWinAmount = totalWinAmount.add(log.getWinAmount());
            }
            gameCount++;
        }

        // 计算RTP
        BigDecimal rtp = BigDecimal.ZERO;
        if (totalBetAmount.compareTo(BigDecimal.ZERO) > 0) {
            rtp = totalWinAmount
                    .divide(totalBetAmount, 4, java.math.RoundingMode.HALF_UP)
                    .multiply(new BigDecimal("100"))
                    .setScale(2, java.math.RoundingMode.HALF_UP);
        }

        // 获取窗口时间
        long windowStart = context.window().getStart();
        long windowEnd = context.window().getEnd();

        // 构建RTP统计结果
        RTPStatistics statistics = RTPStatistics.builder()
                .dimensionType(dimensionType)
                .dimensionKey(dimensionKey)
                .integratorId(integratorId)
                .totalBetAmount(totalBetAmount)
                .totalWinAmount(totalWinAmount)
                .gameCount(gameCount)
                .rtp(rtp)
                .windowStart(windowStart)
                .windowEnd(windowEnd)
                .statisticsTime(System.currentTimeMillis())
                .build();

        // 输出结果
        out.collect(statistics);
    }
}
