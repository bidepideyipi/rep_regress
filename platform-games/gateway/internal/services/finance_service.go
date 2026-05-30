package services

import (
	"context"
	"fmt"

	"github.com/platform-games/gateway/internal/models"
)

type FinanceService struct {
	repo FinanceRepository
}

type FinanceRepository interface {
	GetBalance(ctx context.Context, userID string) (float64, error)
	GetTransactions(ctx context.Context, userID string, limit, offset int) ([]models.Transaction, error)
}

func NewFinanceService(repo FinanceRepository) *FinanceService {
	return &FinanceService{repo: repo}
}

func (s *FinanceService) GetBalance(ctx context.Context, userID string) (float64, string, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, "CNY", nil
}

func (s *FinanceService) GetTransactions(ctx context.Context, userID string, limit, offset int) ([]models.Transaction, int64, error) {
	transactions, err := s.repo.GetTransactions(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	return transactions, int64(len(transactions)), nil
}
