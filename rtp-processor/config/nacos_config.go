package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type Config struct {
	ClickHouse struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"clickhouse"`
	RocketMQ struct {
		NameServers []string `json:"name_servers"`
		Producer    struct {
			GroupName string `json:"group_name"`
			Topic     string `json:"topic"`
		} `json:"producer"`
		Consumer struct {
			GroupName string `json:"group_name"`
			Topic     string `json:"topic"`
			BatchSize int    `json:"batch_size"`
		} `json:"consumer"`
	} `json:"rocket_mq"`
	AggregateInterval string `json:"aggregate_interval"`
	AlertInterval     string `json:"alert_interval"`
}

func LoadConfigFromNacos(nacosServerAddr, namespace, group, dataId string) (*Config, error) {
	clientConfig := constant.NewClientConfig(
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("logs/nacos"),
		constant.WithCacheDir("cache/nacos"),
		constant.WithLogLevel("info"),
		constant.WithNamespaceId(namespace),
	)

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: nacosServerAddr,
			Port:   8848,
		},
	}

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Nacos 客户端失败: %v", err)
	}

	content, err := client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return nil, fmt.Errorf("从 Nacos 获取配置失败: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("解析 Nacos 配置失败: %v", err)
	}

	return &cfg, nil
}

func (c *Config) GetAggregateInterval() time.Duration {
	d, err := time.ParseDuration(c.AggregateInterval)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

func (c *Config) GetAlertInterval() time.Duration {
	d, err := time.ParseDuration(c.AlertInterval)
	if err != nil {
		return 10 * time.Minute
	}
	return d
}

func (c *Config) ClickHouseAddr() string {
	return fmt.Sprintf("%s:%d", c.ClickHouse.Host, c.ClickHouse.Port)
}
