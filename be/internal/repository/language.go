package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/pkg/i18n"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LanguageRepository struct {
	db *pgxpool.Pool
}

func NewLanguageRepository(db *pgxpool.Pool) *LanguageRepository {
	return &LanguageRepository{db: db}
}

func (r *LanguageRepository) GetByCode(ctx context.Context, code string) (*domain.Language, error) {
	query := `
		SELECT language_code, is_active, display_order, created_at
		FROM languages
		WHERE language_code = $1
	`
	language := &domain.Language{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&language.LanguageCode, &language.IsActive, &language.DisplayOrder, &language.CreatedAt,
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
		SELECT language_code, is_active, display_order, created_at
		FROM languages
		WHERE ($1::boolean IS NULL OR is_active = $1)
		ORDER BY display_order
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
			&language.LanguageCode, &language.IsActive, &language.DisplayOrder, &language.CreatedAt,
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

func (r *LanguageRepository) ToggleActive(ctx context.Context, code string, isActive bool) error {
	query := `
		UPDATE languages
		SET is_active = $1
		WHERE language_code = $2
	`
	result, err := r.db.Exec(ctx, query, isActive, code)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		slog.Warn("LanguageRepository.ToggleActive: language not found", "code", code)
		return domain.NewNotFoundError("language")
	}
	return nil
}

func (r *LanguageRepository) UpdateDisplayOrder(ctx context.Context, orders map[string]int) error {
	batch := &pgx.Batch{}
	query := `
		UPDATE languages
		SET display_order = $1
		WHERE language_code = $2
	`

	for code, order := range orders {
		batch.Queue(query, order, code)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for code := range orders {
		result, err := br.Exec()
		if err != nil {
			return domain.NewDatabaseError(err)
		}
		if result.RowsAffected() == 0 {
			slog.Warn("LanguageRepository.UpdateDisplayOrder: language not found", "code", code)
			return domain.NewNotFoundError("language")
		}
	}

	return nil
}

func (r *LanguageRepository) InitializeLanguages(ctx context.Context) error {
	slog.Info("Initializing languages from i18n translations")

	supportedLangs := i18n.GetSupportedLanguages()
	for i, langCode := range supportedLangs {
		query := `
			INSERT INTO languages (language_code, is_active, display_order)
			VALUES ($1, $2, $3)
			ON CONFLICT (language_code) DO NOTHING
		`

		isActive := false
		displayOrder := i + 1

		_, err := r.db.Exec(ctx, query,
			langCode,
			isActive,
			displayOrder,
		)
		if err != nil {
			slog.Error("Failed to initialize language", "code", langCode, "error", err)
			return err
		}

		slog.Debug("Language initialized", "code", langCode, "active", isActive)
	}

	slog.Info("Languages initialized successfully", "count", len(supportedLangs))
	return nil
}
