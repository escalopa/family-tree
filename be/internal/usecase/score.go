package usecase

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
)

type ScoreUseCase interface {
	GetUserScores(ctx context.Context, userID int) ([]*domain.UserScoreWithDetails, int, error)
	GetLeaderboard(ctx context.Context, page, limit int) ([]*domain.LeaderboardEntry, int, error)
}

type ScoreRepo interface {
	GetByUserID(ctx context.Context, userID int) ([]*domain.UserScoreWithDetails, error)
	GetLeaderboard(ctx context.Context, page, limit int) ([]*domain.LeaderboardEntry, int, error)
}

type scoreUseCase struct {
	scoreRepo ScoreRepo
}

func NewScoreUseCase(
	scoreRepo ScoreRepo,
) ScoreUseCase {
	return &scoreUseCase{
		scoreRepo: scoreRepo,
	}
}

func (uc *scoreUseCase) GetUserScores(ctx context.Context, userID int) ([]*domain.UserScoreWithDetails, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}

func (uc *scoreUseCase) GetLeaderboard(ctx context.Context, page, limit int) ([]*domain.LeaderboardEntry, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}
