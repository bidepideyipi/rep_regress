package services

import (
	"context"
	"fmt"

	"github.com/platform-games/gateway/internal/models"
	"github.com/platform-games/gateway/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUserInfo(ctx context.Context, userID, merchantID string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.MerchantID != merchantID {
		return nil, fmt.Errorf("user does not belong to merchant")
	}

	return user, nil
}

func (s *UserService) UpdatePreferences(ctx context.Context, userID string, prefs models.UserPreferences) error {
	return s.userRepo.UpdatePreferences(ctx, userID, prefs)
}
