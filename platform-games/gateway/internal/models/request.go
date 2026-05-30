package models

type GameAccessRequest struct {
	MerchantID  string                 `json:"merchant_id"`
	UserID      string                 `json:"user_id"`
	GameID      string                 `json:"game_id"`
	Language    string                 `json:"language,omitempty"`
	Currency    string                 `json:"currency,omitempty"`
	ReturnURL   string                 `json:"return_url,omitempty"`
	CustomData  map[string]interface{} `json:"custom_data,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
	Nonce       string                 `json:"nonce"`
	Signature   string                 `json:"signature"`
}

type GameAccessResponse struct {
	GameURL     string `json:"game_url"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	SessionID   string `json:"session_id"`
	Timestamp   int64  `json:"timestamp"`
}

type VerifyTokenRequest struct {
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
}

type VerifyTokenResponse struct {
	IsValid   bool       `json:"is_valid"`
	UserInfo  UserInfo   `json:"user_info"`
	ExpiresAt int64      `json:"expires_at"`
}

type UserInfo struct {
	UserID      string   `json:"user_id"`
	MerchantID  string   `json:"merchant_id"`
	Permissions []string `json:"permissions"`
}

type UpdateUserPreferencesRequest struct {
	Preferences UserPreferences `json:"preferences"`
	Timestamp  int64            `json:"timestamp"`
	Signature  string           `json:"signature"`
}

type GameActionRequest struct {
	Action    string                 `json:"action"`
	GameID    string                 `json:"game_id"`
	GameData  map[string]interface{} `json:"game_data"`
	Timestamp int64                  `json:"timestamp"`
	Signature string                 `json:"signature"`
}

type GameActionResponse struct {
	Action       string     `json:"action"`
	Result       GameResult `json:"result"`
	Timestamp    int64      `json:"timestamp"`
}

type GameResult struct {
	GameResult   interface{} `json:"game_result"`
	BalanceChange float64    `json:"balance_change"`
	Balance      float64     `json:"balance"`
	WinAmount    float64     `json:"win_amount"`
	Timestamp    int64       `json:"timestamp"`
}

type GameSyncRequest struct {
	GameID    string                 `json:"game_id"`
	SessionID string                 `json:"session_id"`
	GameState map[string]interface{} `json:"game_state"`
	Timestamp int64                  `json:"timestamp"`
	Signature string                 `json:"signature"`
}

type Transaction struct {
	TransactionID  string  `json:"transaction_id" db:"transaction_id" gorm:"primaryKey"`
	UserID         string  `json:"user_id" db:"user_id"`
	MerchantID     string  `json:"merchant_id" db:"merchant_id"`
	Type           string  `json:"type" db:"type"`
	Amount         float64 `json:"amount" db:"amount"`
	BalanceBefore  float64 `json:"balance_before" db:"balance_before"`
	BalanceAfter   float64 `json:"balance_after" db:"balance_after"`
	Description    string  `json:"description" db:"description"`
	CreatedAt      int64   `json:"timestamp" db:"created_at"`
}

type RefreshKeyRequest struct {
	MerchantID string `json:"merchant_id"`
	KeyType    string `json:"key_type"`
	Reason     string `json:"reason"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
}
