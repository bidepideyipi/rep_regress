package models

import "time"

type JWTClaims struct {
	Issuer      string   `json:"iss"`
	Subject     string   `json:"sub"`
	Audience    string   `json:"aud"`
	ExpiresAt   int64    `json:"exp"`
	IssuedAt    int64    `json:"iat"`
	JTI         string   `json:"jti"`
	MerchantID  string   `json:"merchant_id"`
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
}

type TokenInfo struct {
	JTI         string
	UserID      string
	MerchantID  string
	GameID      string
	ExpiresAt   time.Time
	Revoked     bool
}
