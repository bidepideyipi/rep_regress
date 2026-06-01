package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

var config *AppConfig

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*AppConfig, error) {
	v := viper.New()
	
	// 设置配置文件路径
	v.SetConfigFile(configPath)
	
	// 设置配置文件类型（支持yaml, json等）
	v.SetConfigType("yaml")
	
	// 读取环境变量
	v.AutomaticEnv()
	v.SetEnvPrefix("SLOT")
	v.SetEnvKeyReplacer(strings.NewReplacer("_", "."))
	
	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	// 解析配置
	config = &AppConfig{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	
	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}
	
	log.Printf("配置加载成功: server port=%d, nacos addr=%s", 
		config.Server.Port, config.Nacos.ServerAddr)
	
	return config, nil
}

// GetConfig 获取配置
func GetConfig() *AppConfig {
	return config
}

// validateConfig 验证配置
func validateConfig(cfg *AppConfig) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", cfg.Server.Port)
	}
	
	if cfg.Nacos.ServerAddr == "" {
		return fmt.Errorf("Nacos服务器地址不能为空")
	}
	
	if cfg.Nacos.GameConfig.DataID == "" {
		return fmt.Errorf("游戏配置DataID不能为空")
	}
	
	if cfg.Database.MySQL.Host == "" {
		return fmt.Errorf("MySQL主机地址不能为空")
	}
	
	if cfg.Database.ClickHouse.Host == "" {
		return fmt.Errorf("ClickHouse主机地址不能为空")
	}
	
	return nil
}

// GetMySQLDSN 获取MySQL连接字符串
func GetMySQLDSN() string {
	cfg := GetConfig()
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Database.MySQL.Username,
		cfg.Database.MySQL.Password,
		cfg.Database.MySQL.Host,
		cfg.Database.MySQL.Port,
		cfg.Database.MySQL.Database,
		cfg.Database.MySQL.Charset,
	)
}

// GetClickHouseDSN 获取ClickHouse连接字符串
func GetClickHouseDSN() string {
	cfg := GetConfig()
	return fmt.Sprintf("tcp://%s:%d/%s?username=%s&password=%s",
		cfg.Database.ClickHouse.Host,
		cfg.Database.ClickHouse.Port,
		cfg.Database.ClickHouse.Database,
		cfg.Database.ClickHouse.Username,
		cfg.Database.ClickHouse.Password,
	)
}

// GetServerAddr 获取服务器地址
func GetServerAddr() string {
	cfg := GetConfig()
	return fmt.Sprintf(":%d", cfg.Server.Port)
}

// GetRocketMQNameSrv 获取NameServer地址
func GetRocketMQNameSrv() string {
	cfg := GetConfig()
	if len(cfg.RocketMQ.NameServers) == 0 {
		return "127.0.0.1:9876"
	}
	return cfg.RocketMQ.NameServers[0]
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