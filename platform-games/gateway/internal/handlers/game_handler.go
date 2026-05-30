package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GameHandler struct {
	logger *zap.Logger
}

func NewGameHandler(logger *zap.Logger) *GameHandler {
	return &GameHandler{logger: logger}
}

func (h *GameHandler) GameAction(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Game action - to be implemented"})
}

func (h *GameHandler) GameSync(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Game sync - to be implemented"})
}
