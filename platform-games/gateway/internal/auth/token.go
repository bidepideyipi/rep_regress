package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/platform-games/gateway/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type TokenManager struct {
	jwtManager  *JWTManager
	aesCipher   *AESCipher
	tokenExpiry int64
}

func NewTokenManager(jwtManager *JWTManager, aesCipher *AESCipher, tokenExpiry int64) *TokenManager {
	return &TokenManager{
		jwtManager:  jwtManager,
		aesCipher:   aesCipher,
		tokenExpiry: tokenExpiry,
	}
}

func (tm *TokenManager) GenerateAccessToken(userID, merchantID, gameID string) (string, error) {
	jti := uuid.New().String()
	now := time.Now()

	claims := &models.JWTClaims{
		Issuer:     tm.jwtManager.issuer,
		Subject:    userID,
		Audience:   gameID,
		ExpiresAt: now.Add(time.Duration(tm.tokenExpiry) * time.Second).Unix(),
		IssuedAt:   now.Unix(),
		JTI:        jti,
		MerchantID: merchantID,
		UserID:     userID,
		Permissions: []string{"game:play", "game:bet", "game:collect"},
	}

	return tm.jwtManager.GenerateToken(claims)
}

func (tm *TokenManager) VerifyAccessToken(tokenString string) (*models.JWTClaims, error) {
	claims, err := tm.jwtManager.VerifyToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

func (tm *TokenManager) GenerateGameURL(accessToken, gameID string, params map[string]string) (string, error) {
	urlData := map[string]string{
		"token":   accessToken,
		"game_id": gameID,
	}

	for k, v := range params {
		urlData[k] = v
	}

	dataJSON := mapToJSON(urlData)
	encryptedData, err := tm.aesCipher.Encrypt(dataJSON)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %w", err)
	}

	timestamp := time.Now().Unix()
	signature := tm.generateURLSignature(encryptedData, timestamp)

	return fmt.Sprintf("https://games.platform.com/%s?data=%s&timestamp=%d&signature=%s",
		gameID, encryptedData, timestamp, signature), nil
}

func (tm *TokenManager) VerifyGameURL(data, timestampStr, signature string) error {
	return nil
}

func (tm *TokenManager) generateURLSignature(data string, timestamp int64) string {
	sig := fmt.Sprintf("%s|%d", data, timestamp)
	return base64.StdEncoding.EncodeToString([]byte(sig))
}

func mapToJSON(m map[string]string) string {
	var result string
	for k, v := range m {
		if result != "" {
			result += "&"
		}
		result += k + "=" + v
	}
	return result
}

func ConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
