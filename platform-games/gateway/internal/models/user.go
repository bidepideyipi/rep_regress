package models

type User struct {
	UserID        string         `json:"user_id" db:"user_id" gorm:"primaryKey"`
	MerchantID    string         `json:"merchant_id" db:"merchant_id"`
	Username      string         `json:"username" db:"username"`
	AvatarURL     string         `json:"avatar_url" db:"avatar_url"`
	Balance       float64        `json:"balance" db:"balance"`
	Currency      string         `json:"currency" db:"currency"`
	VIPLevel      int            `json:"vip_level" db:"vip_level"`
	TotalPlayTime int64          `json:"total_play_time" db:"total_play_time"`
	LastPlayTime  int64          `json:"last_play_time" db:"last_play_time"`
	Preferences  UserPreferences `json:"preferences" db:"preferences" gorm:"type:json"`
	CreatedAt     int64          `json:"created_at" db:"created_at"`
	UpdatedAt     int64          `json:"updated_at" db:"updated_at"`
}

type UserPreferences struct {
	Language      string `json:"language"`
	SoundEnabled  bool   `json:"sound_enabled"`
	MusicEnabled  bool   `json:"music_enabled"`
}

type UserCache struct {
	UserID        string
	MerchantID    string
	Username      string
	AvatarURL     string
	Balance       float64
	Currency      string
	VIPLevel      int
	Preferences   UserPreferences
}
