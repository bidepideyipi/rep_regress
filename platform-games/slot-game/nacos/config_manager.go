package nacos

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"platform-games/slot-game/models"
)

// ClickHouseConfig ClickHouse配置
type ClickHouseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// ConfigManager Nacos配置管理器
type ConfigManager struct {
	client    config_client.IConfigClient
	gameID    string
	config    *models.GameConfig
	namespace string
	group     string
	dataID    string
}

// NewConfigManager 创建配置管理器
func NewConfigManager(serverAddr string, gameID, namespace, group, dataID string) (*ConfigManager, error) {
	// Nacos客户端配置
	clientConfig := constant.NewClientConfig(
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("logs/nacos"),
		constant.WithCacheDir("cache/nacos"),
		constant.WithLogLevel("info"),
	)

	// Nacos服务器配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: serverAddr,
			Port:   8848,
		},
	}

	// 创建配置客户端
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建Nacos客户端失败: %w", err)
	}

	return &ConfigManager{
		client:    client,
		gameID:    gameID,
		namespace: namespace,
		group:     group,
		dataID:    dataID,
	}, nil
}

// LoadConfig 从Nacos加载配置
func (cm *ConfigManager) LoadConfig() (*models.GameConfig, error) {
	content, err := cm.client.GetConfig(vo.ConfigParam{
		DataId: cm.dataID,
		Group:  cm.group,
	})
	if err != nil {
		return nil, fmt.Errorf("从Nacos获取配置失败: %w", err)
	}

	if content == "" {
		return nil, fmt.Errorf("配置内容为空")
	}

	// 解析JSON配置
	var config models.GameConfig
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	cm.config = &config
	log.Printf("成功加载游戏配置: %s (版本: %s)", config.Config.GameID, config.Config.Version)
	
	return &config, nil
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *models.GameConfig {
	return cm.config
}

// GetRawConfig 获取Nacos原始配置内容
func (cm *ConfigManager) GetRawConfig(dataID, group string) (string, error) {
	content, err := cm.client.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	if err != nil {
		return "", fmt.Errorf("从Nacos获取配置失败: %w", err)
	}
	return content, nil
}

// WatchConfig 监听配置变化
func (cm *ConfigManager) WatchConfig(onChange func(*models.GameConfig)) error {
	err := cm.client.ListenConfig(vo.ConfigParam{
		DataId: cm.dataID,
		Group:  cm.group,
		OnChange: func(namespace, group, dataID, data string) {
			log.Printf("配置变更通知: %s/%s/%s", namespace, group, dataID)
			
			// 解析新配置
			var newConfig models.GameConfig
			if err := json.Unmarshal([]byte(data), &newConfig); err != nil {
				log.Printf("解析新配置失败: %v", err)
				return
			}

			// 检查游戏ID是否匹配
			if newConfig.Config.GameID != cm.gameID {
				log.Printf("游戏ID不匹配，忽略配置: 期望=%s, 实际=%s", cm.gameID, newConfig.Config.GameID)
				return
			}

			cm.config = &newConfig
			log.Printf("配置已更新: %s (版本: %s)", newConfig.Config.GameID, newConfig.Config.Version)
			
			// 调用回调函数
			if onChange != nil {
				onChange(&newConfig)
			}
		},
	})

	return err
}

// StartAutoRefresh 启动自动刷新
func (cm *ConfigManager) StartAutoRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("开始自动刷新配置...")
		if _, err := cm.LoadConfig(); err != nil {
			log.Printf("自动刷新配置失败: %v", err)
		}
	}
}

// UpdateConfig 更新配置到Nacos
func (cm *ConfigManager) UpdateConfig(config *models.GameConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	success, err := cm.client.PublishConfig(vo.ConfigParam{
		DataId:  cm.dataID,
		Group:   cm.group,
		Content: string(data),
	})
	if err != nil {
		return fmt.Errorf("发布配置失败: %w", err)
	}

	if !success {
		return fmt.Errorf("发布配置失败")
	}

	log.Printf("配置已更新到Nacos: %s", cm.dataID)
	return nil
}

// StopWatchConfig 停止监听配置
func (cm *ConfigManager) StopWatchConfig(dataID, group string) error {
	return cm.client.CancelListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
}

// Close 关闭配置管理器
func (cm *ConfigManager) Close() {
	// Nacos SDK会自动处理连接关闭
	log.Println("配置管理器已关闭")
}

// LoadAppConfig 从Nacos加载应用配置（app-config）
func (cm *ConfigManager) LoadAppConfig(dataID, group string) (map[string]string, error) {
	content, err := cm.client.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	if err != nil {
		return nil, fmt.Errorf("从Nacos获取应用配置失败: %w", err)
	}

	if content == "" {
		return nil, fmt.Errorf("应用配置内容为空")
	}

	// 解析properties格式
	config := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	log.Printf("成功加载应用配置: %s/%s", group, dataID)
	return config, nil
}

// GetClickHouseConfig 从应用配置中获取ClickHouse配置
func (cm *ConfigManager) GetClickHouseConfig(dataID, group string) (*ClickHouseConfig, error) {
	appConfig, err := cm.LoadAppConfig(dataID, group)
	if err != nil {
		return nil, err
	}

	// 从properties配置中提取ClickHouse配置
	chConfig := &ClickHouseConfig{
		Host:     getOrDefault(appConfig, "db.clickhouse.url", "127.0.0.1"),
		Username: getOrDefault(appConfig, "db.clickhouse.user", "default"),
		Password: getOrDefault(appConfig, "db.clickhouse.pwd", ""),
	}

	// 解析端口
	if portStr, ok := appConfig["db.clickhouse.port"]; ok {
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
			chConfig.Port = port
		} else {
			chConfig.Port = 9000
		}
	} else {
		chConfig.Port = 9000
	}

	log.Printf("ClickHouse配置: host=%s, port=%d, user=%s",
		chConfig.Host, chConfig.Port, chConfig.Username)

	return chConfig, nil
}

// getOrDefault 从配置map中获取值，不存在则返回默认值
func getOrDefault(config map[string]string, key, defaultValue string) string {
	if value, ok := config[key]; ok {
		return value
	}
	return defaultValue
}

// GetConfigInfo 获取配置信息
func (cm *ConfigManager) GetConfigInfo() map[string]interface{} {
	if cm.config == nil {
		return nil
	}

	// 构建符号赔率表
	symbolPaytable := make([]map[string]interface{}, 0)
	for _, symbol := range cm.config.Symbols {
		multipliers := make([]map[string]interface{}, 0)
		for _, m := range symbol.Multipliers {
			multipliers = append(multipliers, map[string]interface{}{
				"match_count": m.MatchCount,
				"multiplier":  m.Multiplier,
				"is_bet_line": m.IsBetLine,
			})
		}

		symbolPaytable = append(symbolPaytable, map[string]interface{}{
			"symbol_id":     symbol.SymbolID,
			"symbol_name":   symbol.SymbolName,
			"symbol_type":   symbol.SymbolType,
			"is_active":     symbol.IsActive,
			"multipliers":   multipliers,
		})
	}

	return map[string]interface{}{
		"game_id":       cm.config.Config.GameID,
		"game_name":     cm.config.Config.GameName,
		"version":       cm.config.Config.Version,
		"description":   cm.config.Config.Description,
		"last_updated":  cm.config.Config.LastUpdated,
		"symbols_count": len(cm.config.Symbols),
		"reels_count":   len(cm.config.Reels),
		"pay_lines":     cm.config.PayTable.PayLineCount,
		"min_bet":       cm.config.GameSettings.MinBet,
		"max_bet":       cm.config.GameSettings.MaxBet,
		"symbol_paytable": symbolPaytable,
	}
}