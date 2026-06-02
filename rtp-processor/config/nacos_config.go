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

	// 先解析为 map[string]interface{} 以支持多种数据类型
	var rawCfg map[string]interface{}
	if err := json.Unmarshal([]byte(content), &rawCfg); err != nil {
		return nil, fmt.Errorf("解析 Nacos 配置失败: %v", err)
	}

	cfg := &Config{}

	// 解析 ClickHouse 配置
	if ch, ok := rawCfg["clickhouse"].(map[string]interface{}); ok {
		if host, ok := ch["host"].(string); ok {
			cfg.ClickHouse.Host = host
		}
		cfg.ClickHouse.Port = parseInt(ch["port"])
		if username, ok := ch["username"].(string); ok {
			cfg.ClickHouse.Username = username
		}
		if password, ok := ch["password"].(string); ok {
			cfg.ClickHouse.Password = password
		}
		if database, ok := ch["database"].(string); ok {
			cfg.ClickHouse.Database = database
		}
	}

	// 解析 MySQL 配置
	if mysql, ok := rawCfg["mysql"].(map[string]interface{}); ok {
		if host, ok := mysql["host"].(string); ok {
			cfg.MySQL.Host = host
		}
		cfg.MySQL.Port = parseInt(mysql["port"])
		if username, ok := mysql["username"].(string); ok {
			cfg.MySQL.Username = username
		}
		if password, ok := mysql["password"].(string); ok {
			cfg.MySQL.Password = password
		}
		if database, ok := mysql["database"].(string); ok {
			cfg.MySQL.Database = database
		}
	}

	// 解析 RocketMQ 配置
	if rmq, ok := rawCfg["rocket_mq"].(map[string]interface{}); ok {
		if nameServers, ok := rmq["name_servers"].([]interface{}); ok {
			for _, ns := range nameServers {
				if nsStr, ok := ns.(string); ok {
					cfg.RocketMQ.NameServers = append(cfg.RocketMQ.NameServers, nsStr)
				}
			}
		}
		if producer, ok := rmq["producer"].(map[string]interface{}); ok {
			if groupName, ok := producer["group_name"].(string); ok {
				cfg.RocketMQ.Producer.GroupName = groupName
			}
			if topic, ok := producer["topic"].(string); ok {
				cfg.RocketMQ.Producer.Topic = topic
			}
		}
		if consumer, ok := rmq["consumer"].(map[string]interface{}); ok {
			if groupName, ok := consumer["group_name"].(string); ok {
				cfg.RocketMQ.Consumer.GroupName = groupName
			}
			if topic, ok := consumer["topic"].(string); ok {
				cfg.RocketMQ.Consumer.Topic = topic
			}
			cfg.RocketMQ.Consumer.BatchSize = parseInt(consumer["batch_size"])
		}
	}

	// 解析聚合间隔配置
	if interval, ok := rawCfg["aggregate_user_interval"].(string); ok {
		cfg.AggregateUserInterval = interval
	}
	if interval, ok := rawCfg["aggregate_game_interval"].(string); ok {
		cfg.AggregateGameInterval = interval
	}
	if interval, ok := rawCfg["alert_interval"].(string); ok {
		cfg.AlertInterval = interval
	}

	return cfg, nil
}

// parseInt 从 interface{} 中解析 int，支持 float64 和 string 类型
func parseInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	default:
		return 0
	}
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
