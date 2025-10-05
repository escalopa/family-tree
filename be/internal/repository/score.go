package repository

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/jackc/pgx/v5"
)

type Score struct {
	db *pgx.Conn
}

func NewScore(db *pgx.Conn) *Score {
	return &Score{db: db}
}

func (r *Score) Upsert(ctx context.Context, score *domain.UserScore) error {
	// TODO: Implementation
	return nil
}

func (r *Score) GetByUserID(ctx context.Context, userID int) ([]*domain.UserScoreWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *Score) GetLeaderboard(ctx context.Context, page, limit int) ([]*domain.LeaderboardEntry, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}

func (r *Score) GetUserRank(ctx context.Context, userID int) (int, error) {
	// TODO: Implementation
	return 0, nil
}

func (r *Score) CalculateScore(ctx context.Context, userID, memberID int) (int, error) {
	// TODO: Implementation
	return 0, nil
}
