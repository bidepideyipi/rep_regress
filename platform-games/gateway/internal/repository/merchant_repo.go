package repository

import (
	"context"
	"fmt"

	"github.com/platform-games/gateway/internal/models"
)

type MerchantRepository struct {
	cache *Cache
}

func NewMerchantRepository(cache *Cache) *MerchantRepository {
	return &MerchantRepository{cache: cache}
}

func (r *MerchantRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.MerchantCache, error) {
	merchant, err := r.cache.GetMerchant(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	if merchant == nil {
		return nil, fmt.Errorf("merchant not found with API key: %s", apiKey)
	}

	return merchant, nil
}

func (r *MerchantRepository) GetByID(ctx context.Context, merchantID string) (*models.MerchantCache, error) {
	merchant, err := r.cache.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	if merchant == nil {
		return nil, fmt.Errorf("merchant not found: %s", merchantID)
	}

	return merchant, nil
}

func (r *MerchantRepository) ValidateAccess(ctx context.Context, merchantID, gameID string) (bool, error) {
	merchant, err := r.GetByID(ctx, merchantID)
	if err != nil {
		return false, err
	}

	if !merchant.IsActive {
		return false, nil
	}

	for _, allowedGame := range merchant.AllowedGames {
		if allowedGame == gameID {
			return true, nil
		}
	}

	return false, nil
}
