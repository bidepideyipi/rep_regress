package config

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
	MySQL struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"mysql"`
	AggregateUserInterval string `json:"aggregate_user_interval"`
	AggregateGameInterval string `json:"aggregate_game_interval"`
	AlertInterval         string `json:"alert_interval"`
	// Dependencies 依赖配置
	Dependencies struct {
		ClickHouseRequired bool `json:"clickhouse_required"` // ClickHouse是否必需
		MySQLRequired       bool `json:"mysql_required"`      // MySQL是否必需
		JackpotRequired     bool `json:"jackpot_required"`    // Jackpot是否必需
	} `json:"dependencies"`
}
