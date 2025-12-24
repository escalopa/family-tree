package dto

type LanguageResponse struct {
	LanguageCode string `json:"language_code"`
	LanguageName string `json:"language_name"`
	IsActive     bool   `json:"is_active"`
	DisplayOrder int    `json:"display_order"`
}

type CreateLanguageRequest struct {
	LanguageCode string `json:"language_code" binding:"required,min=2,max=10"`
	LanguageName string `json:"language_name" binding:"required,min=1,max=50"`
	DisplayOrder int    `json:"display_order" binding:"omitempty,min=0"`
}

type UpdateLanguageRequest struct {
	LanguageName string `json:"language_name" binding:"required,min=1,max=50"`
	IsActive     bool   `json:"is_active"`
	DisplayOrder int    `json:"display_order" binding:"min=0"`
}

type UserLanguagePreferenceResponse struct {
	PreferredLanguage string `json:"preferred_language"`
}

type UpdateUserLanguagePreferenceRequest struct {
	PreferredLanguage string `json:"preferred_language" binding:"required,min=2,max=10"`
}
