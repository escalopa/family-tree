package repository

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID int) (*domain.UserWithRole, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserWithRole, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateRole(ctx context.Context, userID, newRoleID int) error
	UpdateActiveStatus(ctx context.Context, userID int, isActive bool) error
	List(ctx context.Context) ([]*domain.UserWithRole, error)
	GetTotalScore(ctx context.Context, userID int) (int, error)
}

type userRepository struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// TODO: Implementation
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userID int) (*domain.UserWithRole, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.UserWithRole, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	// TODO: Implementation
	return nil
}

func (r *userRepository) UpdateRole(ctx context.Context, userID, newRoleID int) error {
	// TODO: Implementation
	return nil
}

func (r *userRepository) UpdateActiveStatus(ctx context.Context, userID int, isActive bool) error {
	// TODO: Implementation
	return nil
}

func (r *userRepository) List(ctx context.Context) ([]*domain.UserWithRole, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *userRepository) GetTotalScore(ctx context.Context, userID int) (int, error) {
	// TODO: Implementation
	return 0, nil
}
