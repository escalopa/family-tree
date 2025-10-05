package repository

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/jackc/pgx/v5"
)

type History struct {
	db *pgx.Conn
}

func NewHistory(db *pgx.Conn) *History {
	return &History{db: db}
}

func (r *History) Create(ctx context.Context, history *domain.MemberHistory) error {
	// TODO: Implementation
	return nil
}

func (r *History) GetByMemberID(ctx context.Context, memberID int, page, limit int) ([]*domain.MemberHistoryWithUser, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}

func (r *History) GetByRevision(ctx context.Context, memberID, revision int) (*domain.MemberHistory, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *History) GetRecentActivities(ctx context.Context, page, limit int) ([]*domain.Activity, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}

func (r *History) GetUserRecentActivities(ctx context.Context, userID int, page, limit int) ([]*domain.Activity, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}
