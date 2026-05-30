package services

import (
	"context"
	"fmt"
	"time"

	"github.com/platform-games/gateway/internal/auth"
	"github.com/platform-games/gateway/internal/models"
	"github.com/platform-games/gateway/internal/repository"
	"github.com/platform-games/gateway/internal/security"
)

type AuthService struct {
	merchantRepo   *repository.MerchantRepository
	hmacSigner     *auth.HMACSigner
	tokenManager   *auth.TokenManager
	replayProtect  *security.ReplayProtection
	validator      *security.Validator
	timestampTTL   int64
}

func NewAuthService(
	merchantRepo *repository.MerchantRepository,
	hmacSigner *auth.HMACSigner,
	tokenManager *auth.TokenManager,
	replayProtect *security.ReplayProtection,
	validator *security.Validator,
	timestampTTL int64,
) *AuthService {
	return &AuthService{
		merchantRepo:  merchantRepo,
		hmacSigner:    hmacSigner,
		tokenManager:  tokenManager,
		replayProtect: replayProtect,
		validator:     validator,
		timestampTTL:  timestampTTL,
	}
}

func (s *AuthService) ProcessGameAccess(ctx context.Context, req *models.GameAccessRequest) (*models.GameAccessResponse, error) {
	if err := s.validateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	merchant, err := s.merchantRepo.GetByID(ctx, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	if !merchant.IsActive {
		return nil, fmt.Errorf("merchant is not active")
	}

	signer := auth.NewHMACSigner(merchant.APISecret)
	if !signer.Verify(map[string]string{
		"merchant_id": req.MerchantID,
		"user_id":     req.UserID,
		"game_id":     req.GameID,
	}, req.Timestamp, req.Nonce, req.Signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(req.UserID, req.MerchantID, req.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	sessionID := security.GenerateSessionID()

	gameURL, err := s.tokenManager.GenerateGameURL(accessToken, req.GameID, map[string]string{
		"language":   req.Language,
		"currency":   req.Currency,
		"return_url": req.ReturnURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate game URL: %w", err)
	}

	return &models.GameAccessResponse{
		GameURL:     gameURL,
		AccessToken: accessToken,
		ExpiresIn:   1800,
		SessionID:   sessionID,
		Timestamp:   time.Now().Unix(),
	}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, token string) (*models.VerifyTokenResponse, error) {
	claims, err := s.tokenManager.VerifyAccessToken(token)
	if err != nil {
		return &models.VerifyTokenResponse{
			IsValid:   false,
			ExpiresAt: 0,
		}, nil
	}

	return &models.VerifyTokenResponse{
		IsValid: true,
		UserInfo: models.UserInfo{
			UserID:      claims.UserID,
			MerchantID:  claims.MerchantID,
			Permissions: claims.Permissions,
		},
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

func (s *AuthService) validateRequest(ctx context.Context, req *models.GameAccessRequest) error {
	if err := s.validator.ValidateMerchantID(req.MerchantID); err != nil {
		return err
	}

	if err := s.validator.ValidateUserID(req.UserID); err != nil {
		return err
	}

	if err := s.validator.ValidateGameID(req.GameID); err != nil {
		return err
	}

	if err := s.validator.ValidateNonce(req.Nonce); err != nil {
		return err
	}

	if err := s.validator.ValidateTimestamp(req.Timestamp, s.timestampTTL); err != nil {
		return err
	}

	requestID := fmt.Sprintf("%s:%s:%s", req.MerchantID, req.Timestamp, req.Nonce)
	isReplay, err := s.replayProtect.IsReplayRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to check replay attack: %w", err)
	}

	if isReplay {
		return fmt.Errorf("replay attack detected")
	}

	return nil
}

func GenerateSessionID() string {
	return ""
}
