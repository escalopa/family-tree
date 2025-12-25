package usecase

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/pkg/i18n"
)

type (
	languageUseCaseRepo struct {
		lang     LanguageRepository
		langPref UserLanguagePreferenceRepository
	}

	languageUseCase struct {
		repo languageUseCaseRepo
	}
)

func NewLanguageUseCase(
	langRepo LanguageRepository,
	langPrefRepo UserLanguagePreferenceRepository,
) *languageUseCase {
	return &languageUseCase{
		repo: languageUseCaseRepo{
			lang:     langRepo,
			langPref: langPrefRepo,
		},
	}
}

func (uc *languageUseCase) Get(ctx context.Context, code string) (*domain.Language, error) {
	return uc.repo.lang.GetByCode(ctx, code)
}

func (uc *languageUseCase) List(ctx context.Context, activeOnly bool) ([]*domain.Language, error) {
	filter := domain.LanguageFilter{}
	if activeOnly {
		active := true
		filter.IsActive = &active
	}
	return uc.repo.lang.GetAll(ctx, filter)
}

func (uc *languageUseCase) ToggleActive(ctx context.Context, code string, isActive bool) error {
	if !i18n.IsSupported(code) {
		return domain.NewNotFoundError("language")
	}

	return uc.repo.lang.ToggleActive(ctx, code, isActive)
}

func (uc *languageUseCase) UpdatePreference(ctx context.Context, pref *domain.UserLanguagePreference) error {
	if !i18n.IsSupported(pref.PreferredLanguage) {
		return domain.NewNotFoundError("language")
	}

	lang, err := uc.repo.lang.GetByCode(ctx, pref.PreferredLanguage)
	if err != nil {
		return err
	}
	if !lang.IsActive {
		return domain.NewValidationError("error.language.not_active", nil)
	}

	if err := uc.repo.langPref.Upsert(ctx, pref); err != nil {
		return err
	}

	return nil
}

func (uc *languageUseCase) UpdateDisplayOrder(ctx context.Context, orders map[string]int) error {
	for code := range orders {
		if !i18n.IsSupported(code) {
			return domain.NewNotFoundError("language")
		}
	}

	return uc.repo.lang.UpdateDisplayOrder(ctx, orders)
}
