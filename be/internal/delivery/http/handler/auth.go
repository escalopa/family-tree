package handler

import (
	"net/http"

	httpErrors "github.com/escalopa/family-tree/internal/delivery/http"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

type authHandler struct {
	authUseCase AuthUseCase
}

func NewAuthHandler(authUseCase AuthUseCase) *authHandler {
	return &authHandler{authUseCase: authUseCase}
}

// GetAuthURL returns the OAuth authorization URL for any provider
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

// HandleCallback handles the OAuth callback for any provider
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

	// Set cookies
	c.SetCookie("auth_token", tokens.AccessToken, 3600, "/", "", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*3600, "/", "", false, true)
	c.SetCookie("session_id", tokens.SessionID, 7*24*3600, "/", "", false, true)

	response := dto.AuthResponse{}
	response.User.UserID = user.UserID
	response.User.FullName = user.FullName
	response.User.Email = user.Email
	response.User.Avatar = user.Avatar
	response.User.RoleID = user.RoleID
	response.User.IsActive = user.IsActive

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *authHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	tokens, err := h.authUseCase.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	// Set new cookies
	c.SetCookie("auth_token", tokens.AccessToken, 3600, "/", "", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*3600, "/", "", false, true)

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "tokens refreshed"})
}

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

	// Clear cookies
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "logged out"})
}

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

	// Clear cookies
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "logged out from all devices"})
}
