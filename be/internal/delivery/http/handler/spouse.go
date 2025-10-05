package handler

import (
	"github.com/escalopa/family-tree-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type SpouseHandler struct {
	spouseUseCase usecase.SpouseUseCase
}

func NewSpouseHandler(spouseUseCase usecase.SpouseUseCase) *SpouseHandler {
	return &SpouseHandler{
		spouseUseCase: spouseUseCase,
	}
}

func (h *SpouseHandler) Create(c *gin.Context) {
	// TODO: Implementation
}

func (h *SpouseHandler) Update(c *gin.Context) {
	// TODO: Implementation
}

func (h *SpouseHandler) Delete(c *gin.Context) {
	// TODO: Implementation
}
