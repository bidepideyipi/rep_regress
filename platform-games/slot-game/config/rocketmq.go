package config

import (
	"fmt"
)

// RocketMQConfig RocketMQ配置
type RocketMQConfig struct {
	NameServers []string `yaml:"name_servers" mapstructure:"name_servers"`
	GroupName   string   `yaml:"group_name" mapstructure:"group_name"`
	// 生产者配置
	Producer struct {
		GroupName string `yaml:"group_name" mapstructure:"group_name"`
		Topic     string `yaml:"topic" mapstructure:"topic"`
	} `yaml:"producer" mapstructure:"producer"`
	// 消费者配置
	Consumer struct {
		GroupName string `yaml:"group_name" mapstructure:"group_name"`
		Topic     string `yaml:"topic" mapstructure:"topic"`
		BatchSize int    `yaml:"batch_size" mapstructure:"batch_size"`
	} `yaml:"consumer" mapstructure:"consumer"`
}

// GetRocketMQNameSrv 获取NameServer地址
func GetRocketMQNameSrv() string {
	cfg := GetConfig()
	if len(cfg.RocketMQ.NameServers) == 0 {
		return "127.0.0.1:9876"
	}
	return fmt.Sprintf("%s", cfg.RocketMQ.NameServers[0])
}

// GetRocketMQProducerGroup 获取生产者组名
func GetRocketMQProducerGroup() string {
	cfg := GetConfig()
	if cfg.RocketMQ.Producer.GroupName == "" {
		return "slot_game_producer_group"
	}
	return cfg.RocketMQ.Producer.GroupName
}

// GetRocketMQTopic 获取Topic名称
func GetRocketMQTopic() string {
	cfg := GetConfig()
	if cfg.RocketMQ.Producer.Topic == "" {
		return "game_log_topic"
	}
	return cfg.RocketMQ.Producer.Topic
}

// GetRocketMQConsumerGroup 获取消费者组名
func GetRocketMQConsumerGroup() string {
	cfg := GetConfig()
	if cfg.RocketMQ.Consumer.GroupName == "" {
		return "slot_game_consumer_group"
	}
	return cfg.RocketMQ.Consumer.GroupName
}

// GetRocketMQConsumerBatchSize 获取消费者批量大小
func GetRocketMQConsumerBatchSize() int {
	cfg := GetConfig()
	if cfg.RocketMQ.Consumer.BatchSize == 0 {
		return 100
	}
	return cfg.RocketMQ.Consumer.BatchSize
}
