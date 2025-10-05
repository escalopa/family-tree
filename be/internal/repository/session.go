package repository

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.UserSession) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID int) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository struct {
	db *pgx.Conn
}

func NewSessionRepository(db *pgx.Conn) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.UserSession) error {
	// TODO: Implementation
	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	// TODO: Implementation
	return nil
}

func (r *sessionRepository) DeleteAllByUserID(ctx context.Context, userID int) error {
	// TODO: Implementation
	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	// TODO: Implementation
	return nil
}
