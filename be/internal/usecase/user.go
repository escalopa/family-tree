package usecase

import (
	"context"
	"fmt"
	"log/slog"

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

func (uc *userUseCase) ListUsers(ctx context.Context, filter domain.UserFilter, cursor *string, limit int) ([]*domain.User, *string, error) {
	return uc.userRepo.List(ctx, filter, cursor, limit)
}

func (uc *userUseCase) determineRoleActionType(oldRoleID, newRoleID int) string {
	if newRoleID < oldRoleID {
		return "REVOKE"
	}
	return "GRANT"
}

func (uc *userUseCase) UpdateUserRole(ctx context.Context, userID, newRoleID, changedBy int) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	oldRoleID := user.RoleID

	if err := uc.userRepo.UpdateRole(ctx, userID, newRoleID); err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	actionType := uc.determineRoleActionType(oldRoleID, newRoleID)

	if err := uc.userRepo.CreateRoleHistory(ctx, userID, oldRoleID, newRoleID, changedBy, actionType); err != nil {
		slog.Error("failed to record role change history",
			"error", err,
			"user_id", userID,
			"old_role_id", oldRoleID,
			"new_role_id", newRoleID,
			"changed_by", changedBy,
			"action_type", actionType,
		)
	}

	return nil
}

func (uc *userUseCase) UpdateUserActive(ctx context.Context, userID int, isActive bool) error {
	return uc.userRepo.UpdateActive(ctx, userID, isActive)
}

func (uc *userUseCase) GetLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error) {
	return uc.scoreRepo.GetLeaderboard(ctx, limit)
}

func (uc *userUseCase) GetScoreHistory(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error) {
	return uc.scoreRepo.GetByUserID(ctx, userID, cursor, limit)
}

func (uc *userUseCase) GetUserChanges(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	return uc.historyRepo.GetByUserID(ctx, userID, cursor, limit)
}
