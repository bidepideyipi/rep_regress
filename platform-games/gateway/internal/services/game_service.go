package services

import (
	"context"
	"fmt"

	"github.com/platform-games/gateway/internal/models"
)

type GameService struct {
	repo GameRepository
}

type GameRepository interface {
	GetGame(ctx context.Context, gameID string) (*models.Game, error)
	GetSession(ctx context.Context, sessionID string) (*models.GameSession, error)
	CreateSession(ctx context.Context, session *models.GameSession) error
	UpdateSession(ctx context.Context, session *models.GameSession) error
}

func NewGameService(repo GameRepository) *GameService {
	return &GameService{repo: repo}
}

func (s *GameService) ProcessAction(ctx context.Context, userID, gameID string, action string, gameData map[string]interface{}) (*models.GameActionResponse, error) {
	session, err := s.repo.GetSession(ctx, fmt.Sprintf("session_%s_%s", userID, gameID))
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	result, err := s.executeGameAction(session, action, gameData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute action: %w", err)
	}

	return &models.GameActionResponse{
		Action:    action,
		Result:    *result,
		Timestamp: 0,
	}, nil
}

func (s *GameService) SyncState(ctx context.Context, sessionID string, gameState map[string]interface{}) error {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.GameState = fmt.Sprintf("%v", gameState)
	return s.repo.UpdateSession(ctx, session)
}

func (s *GameService) executeGameAction(session *models.GameSession, action string, gameData map[string]interface{}) (*models.GameResult, error) {
	switch action {
	case "spin":
		return s.processSpin(session, gameData)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (s *GameService) processSpin(session *models.GameSession, gameData map[string]interface{}) (*models.GameResult, error) {
	betAmount := gameData["bet_amount"].(float64)

	result := &models.GameResult{
		GameResult: models.SpinResult{
			ReelResult: [][]string{
				{"cherry", "lemon", "orange"},
				{"plum", "grape", "watermelon"},
				{"bell", "seven", "wild"},
			},
			WinLines: []models.WinLine{
				{
					LineID:      "1",
					SymbolID:    "cherry",
					MatchCount:  3,
					WinAmount:   10.0,
					Multiplier:  10.0,
					Positions:   []int{0, 3, 6},
				},
			},
			TotalWin: 10.0,
		},
		BalanceChange: 9.0,
		Balance:       1000.0,
		WinAmount:     10.0,
		Timestamp:     0,
	}

	return result, nil
}
