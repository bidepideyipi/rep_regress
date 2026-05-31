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
	AggregateUserInterval string `json:"aggregate_user_interval"`
	AggregateGameInterval string `json:"aggregate_game_interval"`
	AlertInterval         string `json:"alert_interval"`
}
