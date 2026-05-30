package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
	Security SecurityConfig `mapstructure:"security"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type AuthConfig struct {
	JWT               JWTConfig `mapstructure:"jwt"`
	TokenExpiresIn    int64      `mapstructure:"token_expires_in"`
	RefreshExpiresIn  int64      `mapstructure:"refresh_expires_in"`
	TimestampTTL      int64      `mapstructure:"timestamp_ttl"`
	Issuer            string     `mapstructure:"issuer"`
}

type JWTConfig struct {
	PrivateKeyPath string `mapstructure:"private_key_path"`
	PublicKeyPath  string `mapstructure:"public_key_path"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime  int    `mapstructure:"conn_max_lifetime"`
}

type SecurityConfig struct {
	RateLimit      RateLimitConfig      `mapstructure:"rate_limit"`
	IPWhitelist    []string             `mapstructure:"ip_whitelist"`
	EnableReplayProtection bool        `mapstructure:"enable_replay_protection"`
	ReplayTTL      int64                `mapstructure:"replay_ttl"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

var cfg *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.read_timeout", 10*time.Second)
	viper.SetDefault("server.write_timeout", 10*time.Second)
	viper.SetDefault("server.shutdown_timeout", 10*time.Second)

	viper.SetDefault("auth.token_expires_in", 1800)
	viper.SetDefault("auth.refresh_expires_in", 604800)
	viper.SetDefault("auth.timestamp_ttl", 300)
	viper.SetDefault("auth.issuer", "platform.com")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	viper.SetDefault("security.rate_limit.requests_per_minute", 100)
	viper.SetDefault("security.rate_limit.burst", 20)
	viper.SetDefault("security.enable_replay_protection", true)
	viper.SetDefault("security.replay_ttl", 300)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output_path", "stdout")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func Get() *Config {
	return cfg
}
