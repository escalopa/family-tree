package repository

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserLanguagePreferenceRepository struct {
	db *pgxpool.Pool
}

func NewUserLanguagePreferenceRepository(db *pgxpool.Pool) *UserLanguagePreferenceRepository {
	return &UserLanguagePreferenceRepository{db: db}
}

func (r *UserLanguagePreferenceRepository) Upsert(ctx context.Context, pref *domain.UserLanguagePreference) error {
	query := `
		INSERT INTO user_language_preferences (user_id, preferred_language, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id)
		DO UPDATE SET
			preferred_language = EXCLUDED.preferred_language,
			updated_at = CURRENT_TIMESTAMP
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		pref.UserID, pref.PreferredLanguage,
	).Scan(&pref.CreatedAt, &pref.UpdatedAt)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}
