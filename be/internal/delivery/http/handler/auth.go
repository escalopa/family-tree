package handler

import (
	"github.com/escalopa/family-tree-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	// TODO: Implementation
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// TODO: Implementation
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// TODO: Implementation
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Implementation
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// TODO: Implementation
}
