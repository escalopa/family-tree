package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LanguageRepository struct {
	db *pgxpool.Pool
}

func NewLanguageRepository(db *pgxpool.Pool) *LanguageRepository {
	return &LanguageRepository{db: db}
}

func (r *LanguageRepository) Create(ctx context.Context, language *domain.Language) error {
	query := `
		INSERT INTO languages (language_code, language_name, is_active, display_order)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`
	err := r.db.QueryRow(ctx, query,
		language.LanguageCode, language.LanguageName, language.IsActive, language.DisplayOrder,
	).Scan(&language.CreatedAt)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *LanguageRepository) GetByCode(ctx context.Context, code string) (*domain.Language, error) {
	query := `
		SELECT language_code, language_name, is_active, display_order, created_at
		FROM languages
		WHERE language_code = $1
	`
	language := &domain.Language{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&language.LanguageCode, &language.LanguageName,
		&language.IsActive, &language.DisplayOrder, &language.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn("LanguageRepository.GetByCode: language not found", "code", code)
		return nil, domain.NewNotFoundError("language")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return language, nil
}

func (r *LanguageRepository) GetAll(ctx context.Context, filter domain.LanguageFilter) ([]*domain.Language, error) {
	query := `
		SELECT language_code, language_name, is_active, display_order, created_at
		FROM languages
		WHERE ($1::boolean IS NULL OR is_active = $1)
		ORDER BY display_order, language_code
	`
	rows, err := r.db.Query(ctx, query, filter.IsActive)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var languages []*domain.Language
	for rows.Next() {
		language := &domain.Language{}
		err := rows.Scan(
			&language.LanguageCode, &language.LanguageName,
			&language.IsActive, &language.DisplayOrder, &language.CreatedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		languages = append(languages, language)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return languages, nil
}

func (r *LanguageRepository) Update(ctx context.Context, language *domain.Language) error {
	query := `
		UPDATE languages
		SET language_name = $1, is_active = $2, display_order = $3
		WHERE language_code = $4
	`
	result, err := r.db.Exec(ctx, query,
		language.LanguageName, language.IsActive, language.DisplayOrder, language.LanguageCode,
	)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		slog.Warn("LanguageRepository.Update: language not found", "code", language.LanguageCode)
		return domain.NewNotFoundError("language")
	}
	return nil
}
