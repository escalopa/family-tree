package usecase

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
)

type UserUseCase interface {
	GetByID(ctx context.Context, userID int) (*domain.UserProfile, error)
	GetCurrentUser(ctx context.Context, userID int) (*domain.UserProfile, error)
	List(ctx context.Context) ([]*domain.UserWithRole, error)
	UpdateRole(ctx context.Context, targetUserID, newRoleID, actorUserID int) error
	UpdateActiveStatus(ctx context.Context, targetUserID int, isActive bool) error
	AdminLogout(ctx context.Context, targetUserID int) error
}

// Local interfaces (subset) for the repositories used here
type UserRepo interface {
	GetByID(ctx context.Context, userID int) (*domain.UserWithRole, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserWithRole, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateRole(ctx context.Context, userID, newRoleID int) error
	UpdateActiveStatus(ctx context.Context, userID int, isActive bool) error
	List(ctx context.Context) ([]*domain.UserWithRole, error)
	GetTotalScore(ctx context.Context, userID int) (int, error)
}

type ScoreRepo interface {
	GetByUserID(ctx context.Context, userID int) ([]*domain.UserScoreWithDetails, error)
}

type SessionRepo interface {
	DeleteAllByUserID(ctx context.Context, userID int) error
}

type userUseCase struct {
	userRepo    UserRepo
	scoreRepo   ScoreRepo
	sessionRepo SessionRepo
}

func NewUserUseCase(
	userRepo UserRepo,
	scoreRepo ScoreRepo,
	sessionRepo SessionRepo,
) UserUseCase {
	return &userUseCase{
		userRepo:    userRepo,
		scoreRepo:   scoreRepo,
		sessionRepo: sessionRepo,
	}
}

func (uc *userUseCase) GetByID(ctx context.Context, userID int) (*domain.UserProfile, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *userUseCase) GetCurrentUser(ctx context.Context, userID int) (*domain.UserProfile, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *userUseCase) List(ctx context.Context) ([]*domain.UserWithRole, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *userUseCase) UpdateRole(ctx context.Context, targetUserID, newRoleID, actorUserID int) error {
	// TODO: Implementation
	return nil
}

func (uc *userUseCase) UpdateActiveStatus(ctx context.Context, targetUserID int, isActive bool) error {
	// TODO: Implementation
	return nil
}

func (uc *userUseCase) AdminLogout(ctx context.Context, targetUserID int) error {
	// TODO: Implementation
	return nil
}
