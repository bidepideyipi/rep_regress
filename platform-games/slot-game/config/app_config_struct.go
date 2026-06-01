package config

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
