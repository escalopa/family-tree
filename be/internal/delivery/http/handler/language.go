package handler

import (
	"net/http"
	"strconv"

	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
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
// @Description Get all supported languages (optionally filter by active status)
// @Tags languages
// @Accept json
// @Produce json
// @Param active query boolean false "Filter by active status"
// @Success 200 {array} dto.LanguageResponse
// @Failure 500 {object} dto.Response
// @Router /languages [get]
func (h *LanguageHandler) GetLanguages(c *gin.Context) {
	activeOnly, _ := strconv.ParseBool(c.Query("active")) // default to false

	languages, err := h.languageUC.GetAllLanguages(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Error: err.Error()})
		return
	}

	response := make([]dto.LanguageResponse, len(languages))
	for i, lang := range languages {
		response[i] = dto.LanguageResponse{
			LanguageCode: lang.LanguageCode,
			LanguageName: lang.LanguageName,
			IsActive:     lang.IsActive,
			DisplayOrder: lang.DisplayOrder,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetLanguage godoc
// @Summary Get a language by code
// @Description Get details of a specific language
// @Tags languages
// @Accept json
// @Produce json
// @Param code path string true "Language Code"
// @Success 200 {object} dto.LanguageResponse
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /languages/{code} [get]
func (h *LanguageHandler) GetLanguage(c *gin.Context) {
	code := c.Param("code")

	language, err := h.languageUC.GetLanguage(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{Error: err.Error()})
		return
	}

	response := dto.LanguageResponse{
		LanguageCode: language.LanguageCode,
		LanguageName: language.LanguageName,
		IsActive:     language.IsActive,
		DisplayOrder: language.DisplayOrder,
	}

	c.JSON(http.StatusOK, response)
}

// CreateLanguage godoc
// @Summary Create a new language (Super Admin only)
// @Description Create a new supported language
// @Tags languages
// @Accept json
// @Produce json
// @Param request body dto.CreateLanguageRequest true "Create Language Request"
// @Success 201 {object} dto.LanguageResponse
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /languages [post]
// @Security BearerAuth
func (h *LanguageHandler) CreateLanguage(c *gin.Context) {
	var req dto.CreateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Error: err.Error()})
		return
	}

	language := &domain.Language{
		LanguageCode: req.LanguageCode,
		LanguageName: req.LanguageName,
		IsActive:     true,
		DisplayOrder: req.DisplayOrder,
	}

	if err := h.languageUC.CreateLanguage(c.Request.Context(), language); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Error: err.Error()})
		return
	}

	response := dto.LanguageResponse{
		LanguageCode: language.LanguageCode,
		LanguageName: language.LanguageName,
		IsActive:     language.IsActive,
		DisplayOrder: language.DisplayOrder,
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateLanguage godoc
// @Summary Update a language (Super Admin only)
// @Description Update an existing language
// @Tags languages
// @Accept json
// @Produce json
// @Param code path string true "Language Code"
// @Param request body dto.UpdateLanguageRequest true "Update Language Request"
// @Success 200 {object} dto.LanguageResponse
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /languages/{code} [put]
// @Security BearerAuth
func (h *LanguageHandler) UpdateLanguage(c *gin.Context) {
	code := c.Param("code")

	var req dto.UpdateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Error: err.Error()})
		return
	}

	language := &domain.Language{
		LanguageCode: code,
		LanguageName: req.LanguageName,
		IsActive:     req.IsActive,
		DisplayOrder: req.DisplayOrder,
	}

	if err := h.languageUC.UpdateLanguage(c.Request.Context(), language); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Error: err.Error()})
		return
	}

	response := dto.LanguageResponse{
		LanguageCode: language.LanguageCode,
		LanguageName: language.LanguageName,
		IsActive:     language.IsActive,
		DisplayOrder: language.DisplayOrder,
	}

	c.JSON(http.StatusOK, response)
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
func (h *LanguageHandler) GetUserLanguagePreference(c *gin.Context) {
	// Get preferred language from middleware (already loaded with user)
	preferredLang := middleware.GetPreferredLanguage(c)

	response := dto.UserLanguagePreferenceResponse{
		PreferredLanguage: preferredLang,
	}

	c.JSON(http.StatusOK, response)
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
func (h *LanguageHandler) UpdateUserLanguagePreference(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req dto.UpdateUserLanguagePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Error: err.Error()})
		return
	}

	pref := &domain.UserLanguagePreference{
		UserID:            userID,
		PreferredLanguage: req.PreferredLanguage,
	}

	if err := h.languageUC.UpdateUserLanguagePreference(c.Request.Context(), pref); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Error: err.Error()})
		return
	}

	response := dto.UserLanguagePreferenceResponse{
		PreferredLanguage: pref.PreferredLanguage,
	}

	c.JSON(http.StatusOK, response)
}
