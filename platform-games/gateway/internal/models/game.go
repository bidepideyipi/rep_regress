package models

type Game struct {
	GameID             string   `json:"game_id" db:"game_id" gorm:"primaryKey"`
	GameName           string   `json:"game_name" db:"game_name"`
	GameType           string   `json:"game_type" db:"game_type"`
	GameURL            string   `json:"game_url" db:"game_url"`
	MinBet             float64  `json:"min_bet" db:"min_bet"`
	MaxBet             float64  `json:"max_bet" db:"max_bet"`
	RTP                float64  `json:"rtp" db:"rtp"`
	IsActive           bool     `json:"is_active" db:"is_active"`
	SupportedLanguages []string `json:"supported_languages" db:"supported_languages" gorm:"type:json"`
	SupportedCurrencies []string `json:"supported_currencies" db:"supported_currencies" gorm:"type:json"`
}

type GameSession struct {
	SessionID    string `json:"session_id" db:"session_id" gorm:"primaryKey"`
	UserID       string `json:"user_id" db:"user_id"`
	MerchantID   string `json:"merchant_id" db:"merchant_id"`
	GameID       string `json:"game_id" db:"game_id"`
	CurrentBet   float64 `json:"current_bet" db:"current_bet"`
	AutoSpin     bool   `json:"auto_spin" db:"auto_spin"`
	SoundEnabled bool   `json:"sound_enabled" db:"sound_enabled"`
	GameState    string `json:"game_state" db:"game_state" gorm:"type:json"`
	StartedAt    int64  `json:"started_at" db:"started_at"`
	LastActiveAt int64  `json:"last_active_at" db:"last_active_at"`
}

type SpinResult struct {
	ReelResult  [][]string       `json:"reel_result"`
	WinLines    []WinLine        `json:"win_lines"`
	TotalWin    float64          `json:"total_win"`
	BalanceChange float64        `json:"balance_change"`
	NewBalance   float64         `json:"new_balance"`
}

type WinLine struct {
	LineID      string  `json:"line_id"`
	SymbolID    string  `json:"symbol_id"`
	MatchCount  int     `json:"match_count"`
	WinAmount   float64 `json:"win_amount"`
	Multiplier  float64 `json:"multiplier"`
	Positions   []int   `json:"positions"`
}
