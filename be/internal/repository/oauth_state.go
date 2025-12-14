package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OAuthStateRepository struct {
	db *pgxpool.Pool
}

func NewOAuthStateRepository(db *pgxpool.Pool) *OAuthStateRepository {
	return &OAuthStateRepository{db: db}
}

func (r *OAuthStateRepository) Create(ctx context.Context, state *domain.OAuthState) error {
	query := `
		INSERT INTO oauth_states (state, provider, created_at, expires_at, used)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, state.State, state.Provider, state.CreatedAt, state.ExpiresAt, state.Used)
	if err != nil {
		slog.Error("OAuthStateRepository.Create: insert OAuth state", "error", err)
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *OAuthStateRepository) Get(ctx context.Context, stateStr string) (*domain.OAuthState, error) {
	query := `
		SELECT state, provider, created_at, expires_at, used
		FROM oauth_states
		WHERE state = $1
	`
	state := &domain.OAuthState{}
	err := r.db.QueryRow(ctx, query, stateStr).Scan(
		&state.State, &state.Provider, &state.CreatedAt, &state.ExpiresAt, &state.Used,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewInvalidOAuthStateError()
	}
	if err != nil {
		slog.Error("OAuthStateRepository.Get: query OAuth state", "error", err)
		return nil, domain.NewDatabaseError(err)
	}
	return state, nil
}

func (r *OAuthStateRepository) MarkUsed(ctx context.Context, stateStr string) error {
	query := `UPDATE oauth_states SET used = true WHERE state = $1`
	_, err := r.db.Exec(ctx, query, stateStr)
	if err != nil {
		slog.Error("OAuthStateRepository.MarkUsed: update OAuth state", "error", err)
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *OAuthStateRepository) CleanExpired(ctx context.Context) error {
	query := `DELETE FROM oauth_states WHERE expires_at < NOW() OR used = true`
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		slog.Error("OAuthStateRepository.CleanExpired: delete expired states", "error", err)
		return domain.NewDatabaseError(err)
	}
	return nil
}
