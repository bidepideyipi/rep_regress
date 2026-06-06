package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/shopspring/decimal"
)

type GameLogDetail struct {
	LogID          string          `json:"log_id"`
	GameSessionID  string          `json:"game_session_id"`
	IntegratorID   string          `json:"integrator_id"`
	UserID         string          `json:"user_id"`
	GameID         string          `json:"game_id"`
	BetAmount      decimal.Decimal `json:"bet_amount"`
	WinAmount      decimal.Decimal `json:"win_amount"`
	NetResult      decimal.Decimal `json:"net_result"`
	BetLines       uint16          `json:"bet_lines"`
	BetPerLine     decimal.Decimal `json:"bet_per_line"`
	IsFreeSpin     uint8           `json:"is_free_spin"`
	BonusFeature   string          `json:"bonus_feature"`
	DeviceType     string          `json:"device_type"`
	DeviceOS       string          `json:"device_os"`
	BrowserType    string          `json:"browser_type"`
	IPAddress      uint32          `json:"ip_address"`
	IPRegion       string          `json:"ip_region"`
	IPCountry      string          `json:"ip_country"`
	SessionID      string          `json:"session_id"`
	ServerID       string          `json:"server_id"`
	ProcessingTime uint32          `json:"processing_time_ms"`
	ErrorCode      uint16          `json:"error_code"`
	ErrorMessage   string          `json:"error_message"`
	GameResultJSON string          `json:"game_result_json"`
	ReelResult     [][]string      `json:"reel_result"`
	WinLines       [][]interface{} `json:"win_lines"`
	UserAgent      string          `json:"user_agent"`
	LogTime        time.Time       `json:"log_time"`
}

var (
	nameSrv   = flag.String("namesrv", "127.0.0.1:9876", "RocketMQ NameServer地址")
	topic     = flag.String("topic", "game_log_topic", "Topic名称")
	group     = flag.String("group", "bench_producer_group", "生产者组名")
	count     = flag.Int("count", 1000, "发送消息数量")
	rate      = flag.Int("rate", 100, "每秒发送速率（0=无限制）")
	batchSize = flag.Int("batch", 1, "批量发送大小")
)

func main() {
	flag.Parse()

	log.Printf("=== RocketMQ 消息堆积工具 ===")
	log.Printf("NameServer: %s", *nameSrv)
	log.Printf("Topic: %s", *topic)
	log.Printf("发送数量: %d", *count)
	log.Printf("发送速率: %d msg/s", *rate)
	log.Printf("批量大小: %d", *batchSize)

	namesrv, err := primitive.NewNamesrvAddr(*nameSrv)
	if err != nil {
		log.Fatalf("创建NameServer失败: %v", err)
	}

	prod, err := rocketmq.NewProducer(
		producer.WithGroupName(*group),
		producer.WithNameServer(namesrv),
		producer.WithRetry(3),
	)
	if err != nil {
		log.Fatalf("创建生产者失败: %v", err)
	}

	if err := prod.Start(); err != nil {
		log.Fatalf("启动生产者失败: %v", err)
	}
	defer prod.Shutdown()

	log.Printf("生产者启动成功，开始发送消息...")

	var successCount, failCount int64
	startTime := time.Now()
	msgID := time.Now().UnixNano()

	// 速率控制
	var rateInterval time.Duration
	if *rate > 0 {
		rateInterval = time.Second / time.Duration(*rate)
	}

	for i := 0; i < *count; i++ {
		// 速率限制
		if rateInterval > 0 {
			time.Sleep(rateInterval)
		}

		entry := generateMessage(msgID + int64(i))
		body, _ := json.Marshal(entry)

		msg := &primitive.Message{
			Topic: *topic,
			Body:  body,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = prod.SendSync(ctx, msg)
		cancel()

		if err != nil {
			atomic.AddInt64(&failCount, 1)
			if failCount <= 10 {
				log.Printf("发送失败: %v", err)
			}
		} else {
			atomic.AddInt64(&successCount, 1)
		}

		if (i+1)%100 == 0 {
			elapsed := time.Since(startTime).Seconds()
			currentRate := float64(i+1) / elapsed
			log.Printf("进度: %d/%d, 成功: %d, 失败: %d, 速率: %.0f msg/s",
				i+1, *count, successCount, failCount, currentRate)
		}
	}

	elapsed := time.Since(startTime)

	log.Printf("\n=== 发送完成 ===")
	log.Printf("总耗时: %v", elapsed.Round(time.Millisecond))
	log.Printf("成功: %d", successCount)
	log.Printf("失败: %d", failCount)
	log.Printf("平均速率: %.0f msg/s", float64(successCount)/elapsed.Seconds())
}

func generateMessage(id int64) GameLogDetail {
	now := time.Now()
	// 确保 bet_amount >= 1，避免除零错误
	betVal := float64(1 + int(id%10))
	if betVal < 1 {
		betVal = 1.0
	}

	//模拟96%的RTP
	bet := decimal.NewFromFloat(betVal * 100)
	win := decimal.NewFromFloat(betVal * 96)

	//log.Printf("bet = %s, win= %s\n", bet.String(), win.String())

	return GameLogDetail{
		LogID:          fmt.Sprintf("log_%d_user_%d", id, id%100),
		GameSessionID:  fmt.Sprintf("session_%d", id/100),
		IntegratorID:   "integrator_001",
		UserID:         fmt.Sprintf("user_%d", id%1000),
		GameID:         "game_001",
		BetAmount:      bet,
		WinAmount:      win,
		NetResult:      bet.Sub(win),
		BetLines:       5,
		BetPerLine:     decimal.NewFromFloat(0.1),
		IsFreeSpin:     uint8(id % 5),
		BonusFeature:   "",
		DeviceType:     "mobile",
		DeviceOS:       "ios",
		BrowserType:    "safari",
		IPAddress:      3232235521 + uint32(id%1000), // 192.168.1.1 + offset
		IPRegion:       "guangdong",
		IPCountry:      "CN",
		SessionID:      fmt.Sprintf("sess_%d", id%500),
		ServerID:       "server_001",
		ProcessingTime: uint32(50 + (id % 200)),
		ErrorCode:      0,
		ErrorMessage:   "",
		GameResultJSON: fmt.Sprintf(`{"spin_id":%d}`, id),
		ReelResult:     [][]string{{"cherry", "lemon", "orange"}, {"cherry", "lemon", "orange"}, {"cherry", "lemon", "orange"}},
		WinLines:       [][]interface{}{},
		UserAgent:      "Mozilla/5.0",
		LogTime:        now,
	}
}
