package repository

import (
	"context"
	"errors"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO user_sessions (session_id, user_id, issued_at, expires_at, revoked)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, session.SessionID, session.UserID, session.IssuedAt, session.ExpiresAt, session.Revoked)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *SessionRepository) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	query := `
		SELECT session_id, user_id, issued_at, expires_at, revoked
		FROM user_sessions
		WHERE session_id = $1
	`
	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&session.SessionID, &session.UserID, &session.IssuedAt, &session.ExpiresAt, &session.Revoked,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return session, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, sessionID string) error {
	query := `UPDATE user_sessions SET revoked = true WHERE session_id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *SessionRepository) RevokeAllByUser(ctx context.Context, userID int) error {
	query := `UPDATE user_sessions SET revoked = true WHERE user_id = $1 AND revoked = false`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *SessionRepository) CleanExpired(ctx context.Context) error {
	query := `DELETE FROM user_sessions WHERE expires_at < $1`
	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}
