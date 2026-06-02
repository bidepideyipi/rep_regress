package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// GameLogDetail 游戏日志详情 - 对应rtp_analytics.game_log_detail_local表
type GameLogDetail struct {
	LogID           string          `ch:"log_id" json:"log_id"`
	GameSessionID   string          `ch:"game_session_id" json:"game_session_id"`
	IntegratorID    string          `ch:"integrator_id" json:"integrator_id"`
	UserID          string          `ch:"user_id" json:"user_id"`
	GameID          string          `ch:"game_id" json:"game_id"`
	BetAmount       decimal.Decimal `ch:"bet_amount" json:"bet_amount"`
	WinAmount       decimal.Decimal `ch:"win_amount" json:"win_amount"`
	NetResult       decimal.Decimal `ch:"net_result" json:"net_result"`
	BetLines        uint16          `ch:"bet_lines" json:"bet_lines"`
	BetPerLine      decimal.Decimal `ch:"bet_per_line" json:"bet_per_line"`
	IsFreeSpin      uint8           `ch:"is_free_spin" json:"is_free_spin"`
	BonusFeature    string          `ch:"bonus_feature" json:"bonus_feature"`
	DeviceType      string          `ch:"device_type" json:"device_type"`
	DeviceOS        string          `ch:"device_os" json:"device_os"`
	BrowserType     string          `ch:"browser_type" json:"browser_type"`
	IPAddress       uint32          `ch:"ip_address" json:"ip_address"`
	IPRegion        string          `ch:"ip_region" json:"ip_region"`
	IPCountry       string          `ch:"ip_country" json:"ip_country"`
	SessionID       string          `ch:"session_id" json:"session_id"`
	ServerID        string          `ch:"server_id" json:"server_id"`
	ProcessingTime  uint32          `ch:"processing_time_ms" json:"processing_time_ms"`
	ErrorCode       uint16          `ch:"error_code" json:"error_code"`
	ErrorMessage    string          `ch:"error_message" json:"error_message"`
	GameResultJSON  string          `ch:"game_result_json" json:"game_result_json"`
	ReelResult      [][]string      `ch:"reel_result" json:"reel_result"`
	WinLines        [][]interface{} `ch:"win_lines" json:"win_lines"` // Tuple(UInt16, Decimal(18,4), Array(String), UInt8)
	UserAgent       string          `ch:"user_agent" json:"user_agent"`
	LogTime         time.Time       `ch:"log_time" json:"log_time"`
}
