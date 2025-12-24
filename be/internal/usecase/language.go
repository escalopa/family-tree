package usecase

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
)

type languageUseCase struct {
	langRepo     LanguageRepository
	langPrefRepo UserLanguagePreferenceRepository
}

func NewLanguageUseCase(
	langRepo LanguageRepository,
	langPrefRepo UserLanguagePreferenceRepository,
) *languageUseCase {
	return &languageUseCase{
		langRepo:     langRepo,
		langPrefRepo: langPrefRepo,
	}
}

// CreateLanguage creates a new language (Super Admin only)
func (uc *languageUseCase) Create(ctx context.Context, language *domain.Language) error {
	// Check if language already exists
	existing, _ := uc.langRepo.GetByCode(ctx, language.LanguageCode)
	if existing != nil {
		return domain.NewAlreadyExistsError("language")
	}

	if err := uc.langRepo.Create(ctx, language); err != nil {
		return err
	}

	return nil
}

// GetLanguage gets a language by code
func (uc *languageUseCase) Get(ctx context.Context, code string) (*domain.Language, error) {
	return uc.langRepo.GetByCode(ctx, code)
}

// GetAllLanguages gets all languages (optionally filtered by active status)
func (uc *languageUseCase) List(ctx context.Context, activeOnly bool) ([]*domain.Language, error) {
	filter := domain.LanguageFilter{}
	if activeOnly {
		active := true
		filter.IsActive = &active
	}
	return uc.langRepo.GetAll(ctx, filter)
}

// UpdateLanguage updates a language
func (uc *languageUseCase) Update(ctx context.Context, language *domain.Language) error {
	// Check if language exists
	_, err := uc.langRepo.GetByCode(ctx, language.LanguageCode)
	if err != nil {
		return err
	}

	if err := uc.langRepo.Update(ctx, language); err != nil {
		return err
	}

	return nil
}

func (uc *languageUseCase) UpdatePreference(ctx context.Context, pref *domain.UserLanguagePreference) error {
	lang, err := uc.langRepo.GetByCode(ctx, pref.PreferredLanguage)
	if err != nil {
		return err
	}
	if !lang.IsActive {
		return domain.NewValidationError("error.language.not_active", nil)
	}

	if err := uc.langPrefRepo.Upsert(ctx, pref); err != nil {
		return err
	}

	return nil
}
