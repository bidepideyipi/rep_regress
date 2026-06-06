package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/rtp-processor/db"
	"github.com/rtp-processor/jackpot"
	"github.com/rtp-processor/models"
)

// MessageListener 消息监听器（独立业务逻辑层）
type MessageListener struct {
	chWriter       *db.ClickHouseWriter
	jackpotManager *jackpot.Manager
}

// NewMessageListener 创建消息监听器
func NewMessageListener(chWriter *db.ClickHouseWriter, jackpotMgr *jackpot.Manager) *MessageListener {
	return &MessageListener{
		chWriter:       chWriter,
		jackpotManager: jackpotMgr,
	}
}

// Consume 消费消息（实现RocketMQ监听器接口）
func (l *MessageListener) Consume(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	log.Printf("[MessageListener] 收到 %d 条消息", len(msgs))

	// 解析并收集所有消息
	entries := make([]models.GameLogDetail, 0, len(msgs))
	for _, msg := range msgs {
		var logEntry models.GameLogDetail
		if err := json.Unmarshal(msg.Body, &logEntry); err != nil {
			log.Printf("[MessageListener] JSON解析失败: %v", err)
			continue
		}
		entries = append(entries, logEntry)

		// 计算 Jackpot 累计奖池
		if l.jackpotManager != nil {
			l.jackpotManager.CalculateJackpot(&logEntry)
		}
	}

	// 批量写入 ClickHouse
	if len(entries) > 0 && l.chWriter != nil {
		if err := l.chWriter.BatchWrite(entries); err != nil {
			log.Printf("[MessageListener] 批量写入失败: %v", err)
			return consumer.ConsumeRetryLater, err
		}
	}

	return consumer.ConsumeSuccess, nil
}
