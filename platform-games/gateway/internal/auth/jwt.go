package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/platform-games/gateway/internal/models"
)

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
}

func NewJWTManager(privateKeyPath, publicKeyPath, issuer string) (*JWTManager, error) {
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
	}, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPrivateKeyFromPEM(data)
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(data)
}

func (j *JWTManager) GenerateToken(claims *models.JWTClaims) (string, error) {
	now := time.Now()
	if claims.IssuedAt == 0 {
		claims.IssuedAt = now.Unix()
	}
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = now.Add(30 * time.Minute).Unix()
	}
	claims.Issuer = j.issuer

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":         claims.Issuer,
		"sub":         claims.Subject,
		"aud":         claims.Audience,
		"exp":         claims.ExpiresAt,
		"iat":         claims.IssuedAt,
		"jti":         claims.JTI,
		"merchant_id": claims.MerchantID,
		"user_id":     claims.UserID,
		"permissions": claims.Permissions,
	})

	return token.SignedString(j.privateKey)
}

func (j *JWTManager) VerifyToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return extractClaims(claims)
}

func (j *JWTManager) DecodeToken(tokenString string) (*models.JWTClaims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return extractClaims(claims)
}

func extractClaims(claims jwt.MapClaims) (*models.JWTClaims, error) {
	jwtClaims := &models.JWTClaims{
		Issuer:     getStringClaim(claims, "iss"),
		Subject:    getStringClaim(claims, "sub"),
		Audience:   getStringClaim(claims, "aud"),
		IssuedAt:   getInt64Claim(claims, "iat"),
		ExpiresAt:  getInt64Claim(claims, "exp"),
		JTI:        getStringClaim(claims, "jti"),
		MerchantID: getStringClaim(claims, "merchant_id"),
		UserID:     getStringClaim(claims, "user_id"),
	}

	if permissions, ok := claims["permissions"].([]interface{}); ok {
		for _, p := range permissions {
			if s, ok := p.(string); ok {
				jwtClaims.Permissions = append(jwtClaims.Permissions, s)
			}
		}
	}

	return jwtClaims, nil
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func getInt64Claim(claims jwt.MapClaims, key string) int64 {
	switch val := claims[key].(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	}
	return 0
}
