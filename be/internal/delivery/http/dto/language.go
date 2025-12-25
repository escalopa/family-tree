package dto

type LanguageResponse struct {
	LanguageCode string `json:"language_code"`
	LanguageName string `json:"language_name"`
	IsActive     bool   `json:"is_active"`
	DisplayOrder int    `json:"display_order"`
}

type LanguageURI struct {
	Code string `uri:"code" binding:"required,min=2,max=10"`
}

type ListLanguagesQuery struct {
	Active bool `form:"active"`
}

type UserLanguagePreferenceResponse struct {
	PreferredLanguage string `json:"preferred_language"`
}

type ToggleLanguageActiveRequest struct {
	IsActive bool `json:"is_active"`
}

type UpdateUserLanguagePreferenceRequest struct {
	PreferredLanguage string `json:"preferred_language" binding:"required,min=2,max=10"`
}

type LanguageOrderItem struct {
	LanguageCode string `json:"language_code" binding:"required,min=2,max=10"`
	DisplayOrder int    `json:"display_order" binding:"min=0"`
}

type UpdateLanguageOrderRequest struct {
	Languages []LanguageOrderItem `json:"languages" binding:"required,min=1,dive"`
}
