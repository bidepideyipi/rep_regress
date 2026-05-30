package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/platform-games/gateway/pkg/response"
)

type AuthMiddleware struct {
	jwtManager JWTManager
}

type JWTManager interface {
	VerifyToken(tokenString string) (interface{}, error)
}

func NewAuthMiddleware(jwtManager JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, response.ErrInvalidToken)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, response.ErrInvalidToken)
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := m.jwtManager.VerifyToken(token)
		if err != nil {
			response.Unauthorized(c, response.ErrExpiredToken)
			c.Abort()
			return
		}

		c.Set("user_claims", claims)
		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]

			claims, err := m.jwtManager.VerifyToken(token)
			if err == nil {
				c.Set("user_claims", claims)
			}
		}

		c.Next()
	}
}
