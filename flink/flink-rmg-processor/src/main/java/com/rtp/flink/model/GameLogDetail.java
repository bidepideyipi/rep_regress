package com.rtp.flink.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

/**
 * 游戏日志明细数据模型
 * 对应 rtp-processor 中的 GameLogDetail 结构
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class GameLogDetail {
    /** 主键ID */
    private Long id;

    /** 游戏局唯一标识 */
    private String gameId;

    /** 集成商ID */
    private String integratorId;

    /** 玩家ID */
    private String playerId;

    /** 终端类型: web/mobile/app */
    private String terminal;

    /** 游戏ID */
    private Integer gameIdInt;

    /** 房间ID */
    private Integer roomId;

    /** 下注金额 */
    private BigDecimal betAmount;

    /** 赔付金额 */
    private BigDecimal winAmount;

    /** RTP计算值: win/bet*100 */
    private BigDecimal rtp;

    /** 游戏状态: playing/completed/failed */
    private String status;

    /** 游戏结果数据(JSON格式) */
    private String resultData;

    /** 创建时间 */
    private Long createdAt;

    /** 更新时间 */
    private Long updatedAt;

    /** 扩展字段1 */
    private String ext1;

    /** 扩展字段2 */
    private String ext2;
}
