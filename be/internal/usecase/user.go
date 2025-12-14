package usecase

import (
	"context"
	"fmt"

	"github.com/escalopa/family-tree/internal/domain"
)

type userUseCase struct {
	userRepo    UserRepository
	scoreRepo   ScoreRepository
	historyRepo HistoryRepository
}

func NewUserUseCase(
	userRepo UserRepository,
	scoreRepo ScoreRepository,
	historyRepo HistoryRepository,
) *userUseCase {
	return &userUseCase{
		userRepo:    userRepo,
		scoreRepo:   scoreRepo,
		historyRepo: historyRepo,
	}
}

func (uc *userUseCase) GetUserByID(ctx context.Context, userID int) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

func (uc *userUseCase) GetUserWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error) {
	return uc.userRepo.GetWithScore(ctx, userID)
}

func (uc *userUseCase) ListUsers(ctx context.Context, cursor *string, limit int) ([]*domain.User, *string, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.userRepo.List(ctx, cursor, limit)
}

func (uc *userUseCase) UpdateUserRole(ctx context.Context, userID, newRoleID, changedBy int) error {
	// Get current user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	// Update role
	if err := uc.userRepo.UpdateRole(ctx, userID, newRoleID); err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	// Note: Role history tracking could be implemented here if needed
	_ = user
	_ = changedBy

	return nil
}

func (uc *userUseCase) UpdateUserActive(ctx context.Context, userID int, isActive bool) error {
	return uc.userRepo.UpdateActive(ctx, userID, isActive)
}

func (uc *userUseCase) GetLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error) {
	if limit <= 0 {
		limit = 10
	}
	return uc.scoreRepo.GetLeaderboard(ctx, limit)
}

func (uc *userUseCase) GetScoreHistory(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.scoreRepo.GetByUserID(ctx, userID, cursor, limit)
}

func (uc *userUseCase) GetUserChanges(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.historyRepo.GetByUserID(ctx, userID, cursor, limit)
}
