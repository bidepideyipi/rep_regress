package models

import (
	"time"
)

// UserInfo 用户信息 (对应 user_info 表)
type UserInfo struct {
	UserID               string     `json:"user_id" db:"user_id"`
	IntegratorID         string     `json:"integrator_id" db:"integrator_id"`
	UserName             *string    `json:"user_name,omitempty" db:"user_name"`
	Nickname             *string    `json:"nickname,omitempty" db:"nickname"`
	Email                *string    `json:"email,omitempty" db:"email"`
	Phone                *string    `json:"phone,omitempty" db:"phone"`
	RTPtoleranceThreshold float64   `json:"rtp_tolerance_threshold" db:"rtp_tolerance_threshold"`
	Balance              float64    `json:"balance" db:"balance"`
	VIPLevel             int        `json:"vip_level" db:"vip_level"`
	Status               int        `json:"status" db:"status"`
	LastLoginTime        *time.Time `json:"last_login_time,omitempty" db:"last_login_time"`
	LastLoginIP          *string    `json:"last_login_ip,omitempty" db:"last_login_ip"`
	CreateTime           time.Time  `json:"create_time" db:"create_time"`
	UpdateTime           time.Time  `json:"update_time" db:"update_time"`
}

// TransactionRecord 交易记录 (对应 transaction_records 表)
type TransactionRecord struct {
	ID             int64      `json:"id" db:"id"`
	TransactionID  string     `json:"transaction_id" db:"transaction_id"`
	UserID         string     `json:"user_id" db:"user_id"`
	IntegratorID   string     `json:"integrator_id" db:"integrator_id"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"`
	Amount         float64    `json:"amount" db:"amount"`
	BalanceBefore  float64    `json:"balance_before" db:"balance_before"`
	BalanceAfter   float64    `json:"balance_after" db:"balance_after"`
	GameSessionID  string     `json:"game_session_id,omitempty" db:"game_session_id"`
	ReferenceID    string     `json:"reference_id,omitempty" db:"reference_id"`
	Status         int        `json:"status" db:"status"`
	ErrorMessage   string     `json:"error_message,omitempty" db:"error_message"`
	CreateTime     time.Time  `json:"create_time" db:"create_time"`
}

// TransactionType 交易类型常量
const (
	TransactionTypeBet     = "bet"     // 下注扣减
	TransactionTypeWin     = "win"     // 中奖增加
	TransactionTypeRefund  = "refund"  // 退款
	TransactionTypeAdjust  = "adjust"  // 调整
)

// TransactionStatus 交易状态常量
const (
	TransactionStatusFailed    = 0 // 失败
	TransactionStatusSuccess   = 1 // 成功
	TransactionStatusProcessing = 2 // 处理中
)

// UserStatus 用户状态常量
const (
	UserStatusDisabled = 0 // 禁用
	UserStatusEnabled  = 1 // 启用
	UserStatusFrozen   = 2 // 冻结
)
