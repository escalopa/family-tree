package handler

import (
	"net/http"

	httpErrors "github.com/escalopa/family-tree/internal/delivery/http"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

type authHandler struct {
	authUseCase   AuthUseCase
	cookieManager CookieManager
}

func NewAuthHandler(authUseCase AuthUseCase, cookieManager CookieManager) *authHandler {
	return &authHandler{
		authUseCase:   authUseCase,
		cookieManager: cookieManager,
	}
}

// GetAuthURL godoc
// @Summary Get OAuth authentication URL
// @Description Returns the OAuth authentication URL for the specified provider
// @Tags auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (e.g., google)"
// @Param state query string false "State parameter for OAuth flow"
// @Success 200 {object} dto.Response{data=dto.AuthURLResponse}
// @Failure 400 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/{provider} [get]
func (h *authHandler) GetAuthURL(c *gin.Context) {
	provider := c.Param("provider")
	state := c.Query("state")

	url, err := h.authUseCase.GetAuthURL(provider, state)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.AuthURLResponse{
			URL:      url,
			Provider: provider,
		},
	})
}

// HandleCallback godoc
// @Summary Handle OAuth callback
// @Description Handles the OAuth callback and creates a user session
// @Tags auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (e.g., google)"
// @Param code query string true "OAuth authorization code"
// @Success 200 {object} dto.Response{data=dto.AuthResponse}
// @Failure 400 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/{provider}/callback [get]
func (h *authHandler) HandleCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "missing code"})
		return
	}

	user, tokens, err := h.authUseCase.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
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
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid session"})
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), sessionID); err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	h.cookieManager.ClearAuthCookies(c)

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "logged out"})
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
	if userID == 0 {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid user"})
		return
	}

	if err := h.authUseCase.LogoutAll(c.Request.Context(), userID); err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	h.cookieManager.ClearAuthCookies(c)

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "logged out from all devices"})
}
