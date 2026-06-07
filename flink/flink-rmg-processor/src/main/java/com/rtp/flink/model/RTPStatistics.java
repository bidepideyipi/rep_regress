package com.rtp.flink.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

/**
 * RTP统计数据模型
 * 用于存储全局、终端、玩家三个维度的RTP统计结果
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class RTPStatistics {
    /** 统计维度类型: global/terminal/player */
    private String dimensionType;

    /** 维度键值(终端ID或玩家ID) */
    private String dimensionKey;

    /** 集成商ID */
    private String integratorId;

    /** 游戏ID */
    private Integer gameId;

    /** 总下注金额 */
    @Builder.Default
    private BigDecimal totalBetAmount = BigDecimal.ZERO;

    /** 总赔付金额 */
    @Builder.Default
    private BigDecimal totalWinAmount = BigDecimal.ZERO;

    /** 游戏次数 */
    @Builder.Default
    private Long gameCount = 0L;

    /** RTP百分比: totalWin/totalBet*100 */
    @Builder.Default
    private BigDecimal rtp = BigDecimal.ZERO;

    /** 窗口开始时间(毫秒时间戳) */
    private Long windowStart;

    /** 窗口结束时间(毫秒时间戳) */
    private Long windowEnd;

    /** 统计时间(毫秒时间戳) */
    private Long statisticsTime;

    /**
     * 累加游戏数据
     */
    public void accumulate(GameLogDetail log) {
        this.totalBetAmount = this.totalBetAmount.add(log.getBetAmount() != null ? log.getBetAmount() : BigDecimal.ZERO);
        this.totalWinAmount = this.totalWinAmount.add(log.getWinAmount() != null ? log.getWinAmount() : BigDecimal.ZERO);
        this.gameCount = this.gameCount + 1;

        // 计算RTP
        if (this.totalBetAmount.compareTo(BigDecimal.ZERO) > 0) {
            this.rtp = this.totalWinAmount
                    .divide(this.totalBetAmount, 4, java.math.RoundingMode.HALF_UP)
                    .multiply(new BigDecimal("100"))
                    .setScale(2, java.math.RoundingMode.HALF_UP);
        }
    }

    /**
     * 合并两个RTP统计
     */
    public void merge(RTPStatistics other) {
        this.totalBetAmount = this.totalBetAmount.add(other.getTotalBetAmount() != null ? other.getTotalBetAmount() : BigDecimal.ZERO);
        this.totalWinAmount = this.totalWinAmount.add(other.getTotalWinAmount() != null ? other.getTotalWinAmount() : BigDecimal.ZERO);
        this.gameCount = this.gameCount + (other.getGameCount() != null ? other.getGameCount() : 0L);

        // 重新计算RTP
        if (this.totalBetAmount.compareTo(BigDecimal.ZERO) > 0) {
            this.rtp = this.totalWinAmount
                    .divide(this.totalBetAmount, 4, java.math.RoundingMode.HALF_UP)
                    .multiply(new BigDecimal("100"))
                    .setScale(2, java.math.RoundingMode.HALF_UP);
        }
    }
}
