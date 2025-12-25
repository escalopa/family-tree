package handler

import (
	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type authHandler struct {
	authUseCase   AuthUseCase
	userUseCase   UserUseCase
	cookieManager CookieManager
}

func NewAuthHandler(authUseCase AuthUseCase, userUseCase UserUseCase, cookieManager CookieManager) *authHandler {
	return &authHandler{
		authUseCase:   authUseCase,
		userUseCase:   userUseCase,
		cookieManager: cookieManager,
	}
}

// GetAuthURL godoc
// @Summary Get OAuth authentication URL
// @Description Returns the OAuth authentication URL for the specified provider with a generated state token
// @Tags auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (e.g., google)"
// @Success 200 {object} dto.Response{data=dto.AuthURLResponse}
// @Failure 400 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/{provider} [get]
func (h *authHandler) GetAuthURL(c *gin.Context) {
	var uri dto.ProviderUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	url, err := h.authUseCase.GetURL(c.Request.Context(), uri.Provider)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	data := dto.AuthURLResponse{
		URL:      url,
		Provider: uri.Provider,
	}

	delivery.SuccessWithData(c, data)
}

// HandleCallback godoc
// @Summary Handle OAuth callback
// @Description Handles the OAuth callback and creates a user session
// @Tags auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (e.g., google)"
// @Param code query string true "OAuth authorization code"
// @Param state query string true "OAuth state token for CSRF protection"
// @Success 200 {object} dto.Response{data=dto.AuthResponse}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/{provider}/callback [get]
func (h *authHandler) HandleCallback(c *gin.Context) {
	var uri dto.ProviderUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var params dto.CallbackQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		delivery.Error(c, err)
		return
	}

	user, tokens, err := h.authUseCase.HandleCallback(c.Request.Context(), uri.Provider, params.Code, params.State)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	h.cookieManager.SetAuthCookies(c, tokens.AccessToken, tokens.RefreshToken, tokens.SessionID)

	response := dto.AuthResponse{}
	response.User.UserID = user.UserID
	response.User.FullName = user.FullName
	response.User.Email = user.Email
	response.User.Avatar = user.Avatar
	response.User.RoleID = user.RoleID
	response.User.IsActive = user.IsActive

	delivery.SuccessWithData(c, response)
}

// Logout godoc
// @Summary Logout user
// @Description Invalidates the current user session
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/auth/logout [post]
func (h *authHandler) Logout(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)

	if err := h.authUseCase.Logout(c.Request.Context(), sessionID); err != nil {
		delivery.Error(c, err)
		return
	}

	h.cookieManager.ClearAuthCookies(c)

	delivery.Success(c, "success.auth.logout", nil)
}

// LogoutAll godoc
// @Summary Logout from all devices
// @Description Invalidates all user sessions across all devices
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=string}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/auth/logout-all [post]
func (h *authHandler) LogoutAll(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.authUseCase.LogoutAll(c.Request.Context(), userID); err != nil {
		delivery.Error(c, err)
		return
	}

	h.cookieManager.ClearAuthCookies(c)

	delivery.Success(c, "success.auth.logout_all", nil)
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Returns the currently authenticated user's information
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.AuthResponse}
// @Failure 401 {object} dto.Response
// @Router /api/auth/me [get]
func (h *authHandler) GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		delivery.Error(c, domain.NewUnauthorizedError("error.unauthorized", nil))
		return
	}

	user, err := h.userUseCase.Get(c.Request.Context(), userID)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	response := dto.AuthResponse{}
	response.User.UserID = user.UserID
	response.User.FullName = user.FullName
	response.User.Email = user.Email
	response.User.Avatar = user.Avatar
	response.User.RoleID = user.RoleID
	response.User.IsActive = user.IsActive

	delivery.SuccessWithData(c, response)
}

// GetProviders godoc
// @Summary Get available OAuth providers
// @Description Returns the list of enabled OAuth providers in order
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.ProvidersResponse}
// @Router /auth/providers [get]
func (h *authHandler) GetProviders(c *gin.Context) {
	providers := h.authUseCase.ListProviders(c.Request.Context())

	data := dto.ProvidersResponse{
		Providers: providers,
	}

	delivery.SuccessWithData(c, data)
}
