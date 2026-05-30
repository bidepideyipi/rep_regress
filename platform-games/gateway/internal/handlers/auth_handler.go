package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/platform-games/gateway/internal/models"
	"github.com/platform-games/gateway/pkg/response"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger *zap.Logger
}

func NewAuthHandler(logger *zap.Logger) *AuthHandler {
	return &AuthHandler{logger: logger}
}

func (h *AuthHandler) GameAccess(c *gin.Context) {
	var req models.GameAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	h.logger.Info("Game access request",
		zap.String("merchant_id", req.MerchantID),
		zap.String("user_id", req.UserID),
		zap.String("game_id", req.GameID),
	)

	resp := &models.GameAccessResponse{
		GameURL:     "https://games.platform.com/slot-game?token=placeholder",
		AccessToken: "placeholder_token",
		ExpiresIn:   1800,
		SessionID:   "session_123456",
		Timestamp:   0,
	}

	response.Success(c, resp)
}

func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var req models.VerifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	h.logger.Info("Verify token request")

	resp := &models.VerifyTokenResponse{
		IsValid:   true,
		ExpiresAt: 0,
	}

	response.Success(c, resp)
}
