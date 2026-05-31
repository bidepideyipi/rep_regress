package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig 应用配置
type AppConfig struct {
	Server struct {
		Port            int    `mapstructure:"port"`
		Mode            string `mapstructure:"mode"`
		ReadTimeout     int    `mapstructure:"read_timeout"`
		WriteTimeout    int    `mapstructure:"write_timeout"`
		ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
	} `mapstructure:"server"`

	Nacos struct {
		ServerAddr string `mapstructure:"server_addr"`
		Namespace  string `mapstructure:"namespace"`
		Group      string `mapstructure:"group"`
		GameConfig struct {
			DataID string `mapstructure:"data_id"`
			GameID string `mapstructure:"game_id"`
		} `mapstructure:"game_config"`
		Timeout int `mapstructure:"timeout"`
	} `mapstructure:"nacos"`

	Database struct {
		MySQL struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Database string `mapstructure:"database"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			Charset  string `mapstructure:"charset"`
			MaxIdle  int    `mapstructure:"max_idle"`
			MaxOpen  int    `mapstructure:"max_open"`
		} `mapstructure:"mysql"`
		ClickHouse struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Database string `mapstructure:"database"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		} `mapstructure:"clickhouse"`
	} `mapstructure:"database"`

	Log struct {
		Level      string `mapstructure:"level"`
		Format     string `mapstructure:"format"`
		Output     string `mapstructure:"output"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxAge     int    `mapstructure:"max_age"`
		Compress   bool   `mapstructure:"compress"`
	} `mapstructure:"log"`

	Monitoring struct {
		Enabled bool   `mapstructure:"enabled"`
		Port    int    `mapstructure:"port"`
		Path    string `mapstructure:"path"`
	} `mapstructure:"monitoring"`

	RocketMQ struct {
		NameServers []string `mapstructure:"name_servers"`
		Producer    struct {
			GroupName string `mapstructure:"group_name"`
			Topic     string `mapstructure:"topic"`
		} `mapstructure:"producer"`
		Consumer struct {
			GroupName string `mapstructure:"group_name"`
			Topic     string `mapstructure:"topic"`
			BatchSize int    `mapstructure:"batch_size"`
		} `mapstructure:"consumer"`
	} `mapstructure:"rocketmq"`
}

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