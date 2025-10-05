package usecase

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
)

type HistoryUseCase interface {
	GetMemberHistory(ctx context.Context, memberID, page, limit int) ([]*domain.MemberHistoryWithUser, int, error)
	GetRecentActivities(ctx context.Context, page, limit int) ([]*domain.Activity, int, error)
	GetUserRecentActivities(ctx context.Context, userID, page, limit int) ([]*domain.Activity, int, error)
}

type historyUseCase struct {
	historyRepo HistoryRepository
}

func NewHistoryUseCase(
	historyRepo HistoryRepository,
) HistoryUseCase {
	return &historyUseCase{
		historyRepo: historyRepo,
	}
}

func (uc *historyUseCase) GetMemberHistory(ctx context.Context, memberID, page, limit int) ([]*domain.MemberHistoryWithUser, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}

func (uc *historyUseCase) GetRecentActivities(ctx context.Context, page, limit int) ([]*domain.Activity, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}

func (uc *historyUseCase) GetUserRecentActivities(ctx context.Context, userID, page, limit int) ([]*domain.Activity, int, error) {
	// TODO: Implementation
	return nil, 0, nil
}
