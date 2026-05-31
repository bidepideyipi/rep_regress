package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"platform-games/slot-game/rtp"
)

func main() {
	// 初始化 RTP 服务
	rtpService, err := rtp.NewRTPService("127.0.0.1", 9000, "default", "", "rtp_analytics")
	if err != nil {
		log.Fatal("初始化 RTP 服务失败:", err)
	}
	defer rtpService.Close()

	ctx := context.Background()

	// 示例1: 查询用户实时 RTP
	fmt.Println("========== 用户实时 RTP ========== ")
	userMetrics, err := rtpService.GetUserRealtimeRTP(ctx, "user_001", "game_001", 30)
	if err != nil {
		log.Printf("查询用户实时 RTP 失败: %v", err)
	} else {
		for _, m := range userMetrics {
			fmt.Printf("时间窗口: %s\n", m.TimeWindow.Format(time.RFC3339))
			fmt.Printf("总投注: %s, 总赢取: %s, RTP: %.4f\n", m.TotalBet.String(), m.TotalWin.String(), m.RTP.InexactFloat64())
			fmt.Printf("旋转次数: %d, 平均投注: %s, 最大赢取: %s\n", m.SpinCount, m.AvgBet.String(), m.MaxWin.String())
			fmt.Printf("获胜率: %.4f, 获胜次数: %d, 失败次数: %d\n", m.WinRate.InexactFloat64(), m.WinSpinCount, m.LossSpinCount)
			fmt.Println("---")
		}
	}

	// 示例2: 查询游戏整体 RTP
	fmt.Println("\n========== 游戏整体 RTP ========== ")
	gameMetrics, err := rtpService.GetGameHourlyRTP(ctx, "game_001", 24)
	if err != nil {
		log.Printf("查询游戏整体 RTP 失败: %v", err)
	} else {
		for _, m := range gameMetrics {
			fmt.Printf("时间窗口: %s\n", m.TimeWindow.Format(time.RFC3339))
			fmt.Printf("总投注: %s, 总赢取: %s, RTP: %.4f\n", m.TotalBet.String(), m.TotalWin.String(), m.RTP.InexactFloat64())
			fmt.Printf("旋转次数: %d, 活跃用户: %d\n", m.SpinCount, m.ActiveUsers)
			fmt.Printf("波动率: %.4f, 获胜率: %.4f\n", m.Volatility.InexactFloat64(), m.WinRate.InexactFloat64())
			fmt.Println("---")
		}
	}

	// 示例3: 获取活跃告警
	fmt.Println("\n========== 活跃告警 ========== ")
	alerts, err := rtpService.GetActiveAlerts(ctx, 10)
	if err != nil {
		log.Printf("获取活跃告警失败: %v", err)
	} else {
		if len(alerts) == 0 {
			fmt.Println("当前没有活跃告警")
		} else {
			for _, a := range alerts {
				fmt.Printf("告警ID: %s\n", a.AlertID)
				fmt.Printf("告警类型: %s, 严重程度: %s\n", a.AlertType, a.Severity)
				fmt.Printf("用户ID: %s, 游戏ID: %s\n", a.UserID, a.GameID)
				fmt.Printf("RTP: %.4f, 旋转次数: %d\n", a.RTP.InexactFloat64(), a.SpinCount)
				fmt.Printf("检测时间: %s\n", a.DetectedTime.Format(time.RFC3339))
				fmt.Printf("备注: %s\n", a.Notes)
				fmt.Println("---")
			}
		}
	}

	// 示例4: 手动检测异常 RTP
	fmt.Println("\n========== 异常 RTP 检测 ========== ")
	abnormalUsers, err := rtpService.DetectAbnormalRTP(ctx, 0.98, 0.50, 20)
	if err != nil {
		log.Printf("检测异常 RTP 失败: %v", err)
	} else {
		if len(abnormalUsers) == 0 {
			fmt.Println("没有检测到异常 RTP 用户")
		} else {
			for _, u := range abnormalUsers {
				fmt.Printf("用户ID: %s, 游戏ID: %s\n", u.UserID, u.GameID)
				fmt.Printf("RTP: %.4f (异常)\n", u.RTP.InexactFloat64())
				fmt.Printf("总投注: %s, 总赢取: %s\n", u.TotalBet.String(), u.TotalWin.String())
				fmt.Printf("旋转次数: %d\n", u.SpinCount)
				fmt.Println("---")
			}
		}
	}

	// 示例5: 高 RTP 用户排行
	fmt.Println("\n========== 高 RTP 用户排行 ========== ")
	topUsers, err := rtpService.GetTopHighRTPUsers(ctx, 1, 10, 20)
	if err != nil {
		log.Printf("获取高 RTP 用户排行失败: %v", err)
	} else {
		for i, u := range topUsers {
			fmt.Printf("排名 #%d\n", i+1)
			fmt.Printf("用户ID: %s, 游戏ID: %s\n", u.UserID, u.GameID)
			fmt.Printf("RTP: %.4f\n", u.RTP.InexactFloat64())
			fmt.Printf("总投注: %s, 总赢取: %s\n", u.TotalBet.String(), u.TotalWin.String())
			fmt.Printf("旋转次数: %d\n", u.SpinCount)
			fmt.Println("---")
		}
	}

	fmt.Println("\n========== 查询完成 ========== ")
}