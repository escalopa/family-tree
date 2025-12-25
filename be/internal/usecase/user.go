package usecase

import (
	"context"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
)

type (
	userUseCaseRepo struct {
		user    UserRepository
		score   ScoreRepository
		history HistoryRepository
	}

	userUseCase struct {
		repo userUseCaseRepo
	}
)

func NewUserUseCase(
	userRepo UserRepository,
	scoreRepo ScoreRepository,
	historyRepo HistoryRepository,
) *userUseCase {
	return &userUseCase{
		repo: userUseCaseRepo{
			user:    userRepo,
			score:   scoreRepo,
			history: historyRepo,
		},
	}
}

func (uc *userUseCase) Get(ctx context.Context, userID int) (*domain.User, error) {
	return uc.repo.user.Get(ctx, userID)
}

func (uc *userUseCase) GetWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error) {
	return uc.repo.user.GetWithScore(ctx, userID)
}

func (uc *userUseCase) List(ctx context.Context, filter domain.UserFilter, cursor *string, limit int) ([]*domain.User, *string, error) {
	return uc.repo.user.List(ctx, filter, cursor, limit)
}

func (uc *userUseCase) determineRoleActionType(oldRoleID, newRoleID int) string {
	if newRoleID < oldRoleID {
		return "REVOKE"
	}
	return "GRANT"
}

func (uc *userUseCase) Update(ctx context.Context, userID int, roleID *int, isActive *bool, changedBy int) error {
	var oldRoleID int
	var recordRoleHistory bool

	if roleID != nil {
		user, err := uc.repo.user.Get(ctx, userID)
		if err != nil {
			return err
		}
		oldRoleID = user.RoleID
		recordRoleHistory = true
	}

	if err := uc.repo.user.Update(ctx, userID, roleID, isActive); err != nil {
		return err
	}

	if recordRoleHistory {
		actionType := uc.determineRoleActionType(oldRoleID, *roleID)
		if err := uc.repo.user.CreateRoleHistory(ctx, userID, oldRoleID, *roleID, changedBy, actionType); err != nil {
			slog.Error("record role change history",
				"error", err,
				"user_id", userID,
				"old_role_id", oldRoleID,
				"new_role_id", *roleID,
				"changed_by", changedBy,
				"action_type", actionType,
			)
		}
	}

	return nil
}

func (uc *userUseCase) ListLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error) {
	return uc.repo.score.GetLeaderboard(ctx, limit)
}

func (uc *userUseCase) ListScoreHistory(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error) {
	return uc.repo.score.GetByUserID(ctx, userID, cursor, limit)
}

func (uc *userUseCase) ListChanges(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	return uc.repo.history.GetByUserID(ctx, userID, cursor, limit)
}
