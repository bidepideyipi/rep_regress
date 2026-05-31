package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

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

/**
 * @brief 获取用户聚合间隔
 * @return time.Duration 聚合间隔
 * @note 如果配置中没有指定聚合间隔，默认返回 5 分钟
 */
func (c *Config) GetAggregateUserInterval() time.Duration {
	d, err := time.ParseDuration(c.AggregateUserInterval)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

/**
 * @brief 获取游戏聚合间隔
 * @return time.Duration 游戏聚合间隔
 * @note 如果配置中没有指定游戏聚合间隔，默认返回 5 分钟
 */
func (c *Config) GetAggregateGameInterval() time.Duration {
	d, err := time.ParseDuration(c.AggregateGameInterval)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

/**
 * @brief 获取告警间隔
 * @return time.Duration 告警间隔
 * @note 如果配置中没有指定告警间隔，默认返回 10 分钟
 */
func (c *Config) GetAlertInterval() time.Duration {
	d, err := time.ParseDuration(c.AlertInterval)
	if err != nil {
		return 10 * time.Minute
	}
	return d
}

/**
 * @brief 获取 ClickHouse 连接地址
 * @return string ClickHouse 连接地址
 */
func (c *Config) ClickHouseAddr() string {
	return fmt.Sprintf("%s:%d", c.ClickHouse.Host, c.ClickHouse.Port)
}
