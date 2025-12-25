package handler

import (
	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/pkg/i18n"
	"github.com/gin-gonic/gin"
)

type LanguageHandler struct {
	languageUC LanguageUseCase
}

func NewLanguageHandler(languageUC LanguageUseCase) *LanguageHandler {
	return &LanguageHandler{
		languageUC: languageUC,
	}
}

// GetLanguages godoc
// @Summary Get all languages
// @Description Get all supported languages (optionally filter by active status). Language names are resolved based on interface language.
// @Tags languages
// @Accept json
// @Produce json
// @Param active query boolean false "Filter by active status"
// @Success 200 {array} dto.LanguageResponse
// @Failure 500 {object} dto.Response
// @Router /languages [get]
func (h *LanguageHandler) List(c *gin.Context) {
	var query dto.ListLanguagesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		delivery.Error(c, err)
		return
	}

	languages, err := h.languageUC.List(c.Request.Context(), query.Active)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	response := make([]dto.LanguageResponse, len(languages))
	for i, lang := range languages {
		response[i] = dto.LanguageResponse{
			LanguageCode: lang.LanguageCode,
			LanguageName: i18n.GetLanguageName(lang.LanguageCode, middleware.GetInterfaceLanguage(c)),
			IsActive:     lang.IsActive,
			DisplayOrder: lang.DisplayOrder,
		}
	}

	delivery.SuccessWithData(c, response)
}

// GetLanguage godoc
// @Summary Get a language by code
// @Description Get details of a specific language. Language name is resolved based on interface language.
// @Tags languages
// @Accept json
// @Produce json
// @Param code path string true "Language Code"
// @Success 200 {object} dto.LanguageResponse
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /languages/{code} [get]
func (h *LanguageHandler) Get(c *gin.Context) {
	var uri dto.LanguageURI
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	language, err := h.languageUC.Get(c.Request.Context(), uri.Code)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	response := dto.LanguageResponse{
		LanguageCode: language.LanguageCode,
		LanguageName: i18n.GetLanguageName(language.LanguageCode, middleware.GetInterfaceLanguage(c)),
		IsActive:     language.IsActive,
		DisplayOrder: language.DisplayOrder,
	}

	delivery.SuccessWithData(c, response)
}

// ToggleLanguageActive godoc
// @Summary Toggle language active status (Super Admin only)
// @Description Enable or disable a language. Languages cannot be created or deleted, only toggled.
// @Tags languages
// @Accept json
// @Produce json
// @Param code path string true "Language Code"
// @Param request body dto.ToggleLanguageActiveRequest true "Toggle Active Request"
// @Success 200 {object} dto.LanguageResponse
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /languages/{code}/toggle [patch]
// @Security BearerAuth
func (h *LanguageHandler) ToggleActive(c *gin.Context) {
	var uri dto.LanguageURI
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var req dto.ToggleLanguageActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	if err := h.languageUC.ToggleActive(c.Request.Context(), uri.Code, req.IsActive); err != nil {
		delivery.Error(c, err)
		return
	}

	// Get the updated language
	language, err := h.languageUC.Get(c.Request.Context(), uri.Code)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	// Get interface language from middleware
	interfaceLang := middleware.GetInterfaceLanguage(c)

	// Resolve language name from i18n based on interface language
	languageName := i18n.GetLanguageName(language.LanguageCode, interfaceLang)

	response := dto.LanguageResponse{
		LanguageCode: language.LanguageCode,
		LanguageName: languageName,
		IsActive:     language.IsActive,
		DisplayOrder: language.DisplayOrder,
	}

	delivery.SuccessWithData(c, response)
}

// GetUserLanguagePreference godoc
// @Summary Get user's language preference
// @Description Get the authenticated user's language preference from middleware (no database query)
// @Tags languages
// @Accept json
// @Produce json
// @Success 200 {object} dto.UserLanguagePreferenceResponse
// @Failure 401 {object} dto.Response
// @Router /users/me/preferences/languages [get]
// @Security BearerAuth
func (h *LanguageHandler) GetPreference(c *gin.Context) {
	// Get preferred language from middleware (already loaded with user)
	preferredLang := middleware.GetPreferredLanguage(c)

	response := dto.UserLanguagePreferenceResponse{
		PreferredLanguage: preferredLang,
	}

	delivery.SuccessWithData(c, response)
}

// UpdateUserLanguagePreference godoc
// @Summary Update user's language preference
// @Description Update the authenticated user's language preference
// @Tags languages
// @Accept json
// @Produce json
// @Param request body dto.UpdateUserLanguagePreferenceRequest true "Update Preference"
// @Success 200 {object} dto.UserLanguagePreferenceResponse
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /users/me/preferences/languages [put]
// @Security BearerAuth
func (h *LanguageHandler) UpdatePreference(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req dto.UpdateUserLanguagePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	pref := &domain.UserLanguagePreference{
		UserID:            userID,
		PreferredLanguage: req.PreferredLanguage,
	}

	if err := h.languageUC.UpdatePreference(c.Request.Context(), pref); err != nil {
		delivery.Error(c, err)
		return
	}

	response := dto.UserLanguagePreferenceResponse{
		PreferredLanguage: pref.PreferredLanguage,
	}

	delivery.SuccessWithData(c, response)
}

// UpdateLanguageOrder godoc
// @Summary Update language display order (Super Admin only)
// @Description Update the display order of multiple languages in a single request
// @Tags languages
// @Accept json
// @Produce json
// @Param request body dto.UpdateLanguageOrderRequest true "Update Order Request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /languages/order [put]
// @Security BearerAuth
func (h *LanguageHandler) UpdateOrder(c *gin.Context) {
	var req dto.UpdateLanguageOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	// Convert array to map for easier handling
	orders := make(map[string]int, len(req.Languages))
	for _, item := range req.Languages {
		orders[item.LanguageCode] = item.DisplayOrder
	}

	if err := h.languageUC.UpdateDisplayOrder(c.Request.Context(), orders); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.language.order_updated", nil)
}
