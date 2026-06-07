package com.rtp.flink.aggregator;

import com.rtp.flink.model.GameLogDetail;
import com.rtp.flink.model.RTPStatistics;
import org.apache.flink.api.common.functions.AggregateFunction;

/**
 * RTP聚合函数
 * 用于在时间窗口内聚合游戏日志，计算RTP统计数据
 *
 * 参考 rtp-processor 中的批量处理逻辑
 */
public class RTPAggregateFunction implements AggregateFunction<GameLogDetail, RTPStatistics, RTPStatistics> {

    @Override
    public RTPStatistics createAccumulator() {
        return RTPStatistics.builder()
                .dimensionType("")
                .dimensionKey("")
                .totalBetAmount(java.math.BigDecimal.ZERO)
                .totalWinAmount(java.math.BigDecimal.ZERO)
                .gameCount(0L)
                .rtp(java.math.BigDecimal.ZERO)
                .build();
    }

    @Override
    public RTPStatistics add(GameLogDetail log, RTPStatistics accumulator) {
        accumulator.accumulate(log);
        return accumulator;
    }

    @Override
    public RTPStatistics getResult(RTPStatistics accumulator) {
        return accumulator;
    }

    @Override
    public RTPStatistics merge(RTPStatistics a, RTPStatistics b) {
        a.merge(b);
        return a;
    }
}
