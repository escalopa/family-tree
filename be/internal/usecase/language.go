package usecase

import (
	"context"
	"fmt"

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
func (uc *languageUseCase) CreateLanguage(ctx context.Context, language *domain.Language) error {
	// Check if language already exists
	existing, _ := uc.langRepo.GetByCode(ctx, language.LanguageCode)
	if existing != nil {
		return domain.NewValidationError("language with this code already exists")
	}

	if err := uc.langRepo.Create(ctx, language); err != nil {
		return fmt.Errorf("create language: %w", err)
	}

	return nil
}

// GetLanguage gets a language by code
func (uc *languageUseCase) GetLanguage(ctx context.Context, code string) (*domain.Language, error) {
	return uc.langRepo.GetByCode(ctx, code)
}

// GetAllLanguages gets all languages (optionally filtered by active status)
func (uc *languageUseCase) GetAllLanguages(ctx context.Context, activeOnly bool) ([]*domain.Language, error) {
	filter := domain.LanguageFilter{}
	if activeOnly {
		active := true
		filter.IsActive = &active
	}
	return uc.langRepo.GetAll(ctx, filter)
}

// UpdateLanguage updates a language
func (uc *languageUseCase) UpdateLanguage(ctx context.Context, language *domain.Language) error {
	// Check if language exists
	_, err := uc.langRepo.GetByCode(ctx, language.LanguageCode)
	if err != nil {
		return err
	}

	if err := uc.langRepo.Update(ctx, language); err != nil {
		return fmt.Errorf("update language: %w", err)
	}

	return nil
}

// UpdateUserLanguagePreference updates a user's language preference
func (uc *languageUseCase) UpdateUserLanguagePreference(ctx context.Context, pref *domain.UserLanguagePreference) error {
	// Validate that language exists and is active
	lang, err := uc.langRepo.GetByCode(ctx, pref.PreferredLanguage)
	if err != nil {
		return fmt.Errorf("invalid preferred language: %w", err)
	}
	if !lang.IsActive {
		return domain.NewValidationError("preferred language is not active")
	}

	if err := uc.langPrefRepo.Upsert(ctx, pref); err != nil {
		return fmt.Errorf("update user language preference: %w", err)
	}

	return nil
}
