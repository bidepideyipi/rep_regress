package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FinanceHandler struct {
	logger *zap.Logger
}

func NewFinanceHandler(logger *zap.Logger) *FinanceHandler {
	return &FinanceHandler{logger: logger}
}

func (h *FinanceHandler) GetBalance(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get balance - to be implemented"})
}

func (h *FinanceHandler) GetTransactions(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get transactions - to be implemented"})
}
