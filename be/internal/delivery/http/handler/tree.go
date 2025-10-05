package handler

import (
	"github.com/escalopa/family-tree-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type TreeHandler struct {
	treeUseCase usecase.TreeUseCase
}

func NewTreeHandler(treeUseCase usecase.TreeUseCase) *TreeHandler {
	return &TreeHandler{
		treeUseCase: treeUseCase,
	}
}

func (h *TreeHandler) GetTree(c *gin.Context) {
	// TODO: Implementation
}

func (h *TreeHandler) GetRelation(c *gin.Context) {
	// TODO: Implementation
}
