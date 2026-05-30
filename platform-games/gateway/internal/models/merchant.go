package models

type Merchant struct {
	MerchantID    string   `json:"merchant_id" db:"merchant_id" gorm:"primaryKey"`
	MerchantName  string   `json:"merchant_name" db:"merchant_name"`
	APIKey        string   `json:"-" db:"api_key"`
	APISecret     string   `json:"-" db:"api_secret"`
	AllowedGames  []string `json:"allowed_games" db:"allowed_games" gorm:"type:json"`
	IPWhitelist   []string `json:"ip_whitelist" db:"ip_whitelist" gorm:"type:json"`
	IsActive      bool     `json:"is_active" db:"is_active"`
	CreatedAt     int64    `json:"created_at" db:"created_at"`
	UpdatedAt     int64    `json:"updated_at" db:"updated_at"`
}

type MerchantCache struct {
	MerchantID   string
	APIKey       string
	APISecret    string
	AllowedGames []string
	IPWhitelist  []string
	IsActive     bool
}
