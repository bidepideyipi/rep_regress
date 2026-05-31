package rocketmq

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"platform-games/slot-game/config"
)

// Producer RocketMQ生产者
type Producer struct {
	producer rocketmq.Producer
	once     sync.Once
	topic    string
}

var (
	globalProducer *Producer
	producerOnce  sync.Once
)

// GetProducer 获取全局生产者实例
func GetProducer() (*Producer, error) {
	var err error
	producerOnce.Do(func() {
		globalProducer, err = NewProducer()
	})
	return globalProducer, err
}

// NewProducer 创建生产者
func NewProducer() (*Producer, error) {
	p := &Producer{
		topic: config.GetRocketMQTopic(),
	}

	// 创建生产者
	namesrv, err := primitive.NewNamesrvAddr(config.GetRocketMQNameSrv())
	if err != nil {
		return nil, err
	}
	prod, err := rocketmq.NewProducer(
		producer.WithGroupName(config.GetRocketMQProducerGroup()),
		producer.WithNameServer(namesrv),
	)
	if err != nil {
		return nil, err
	}

	// 启动生产者
	if err := prod.Start(); err != nil {
		return nil, err
	}

	p.producer = prod
	log.Printf("[RocketMQ] 生产者启动成功, topic=%s, group=%s", p.topic, config.GetRocketMQProducerGroup())

	return p, nil
}

// SendGameLogAsync 异步发送游戏日志
func (p *Producer) SendGameLogAsync(logData interface{}) error {
	// 序列化数据
	body, err := json.Marshal(logData)
	if err != nil {
		return err
	}

	log.Printf("[RocketMQ] 准备发送消息, topic=%s, body_size=%d", p.topic, len(body))

	// 构建消息
	msg := &primitive.Message{
		Topic: p.topic,
		Body:  body,
	}

	// 异步发送
	err = p.producer.SendAsync(context.Background(), func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			log.Printf("[RocketMQ] 发送失败: %v", err)
		} else {
			log.Printf("[RocketMQ] 发送成功, msg_id=%s", result.MsgID)
		}
	}, msg)

	return err
}

// SendGameLogSync 同步发送游戏日志
func (p *Producer) SendGameLogSync(logData interface{}) error {
	// 序列化数据
	body, err := json.Marshal(logData)
	if err != nil {
		return err
	}

	// 构建消息
	msg := &primitive.Message{
		Topic: p.topic,
		Body:  body,
	}

	// 同步发送
	result, err := p.producer.SendSync(context.Background(), msg)
	if err != nil {
		return err
	}

	log.Printf("[RocketMQ] 发送成功, msgID=%s", result.MsgID)
	return nil
}

// Shutdown 关闭生产者
func (p *Producer) Shutdown() error {
	if p.producer != nil {
		return p.producer.Shutdown()
	}
	return nil
}
