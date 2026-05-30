package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger *zap.Logger
}

func NewUserHandler(logger *zap.Logger) *UserHandler {
	return &UserHandler{logger: logger}
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get user info - to be implemented"})
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update user info - to be implemented"})
}
