package handler

import (
	"github.com/escalopa/family-tree-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type HistoryHandler struct {
	historyUseCase usecase.HistoryUseCase
}

func NewHistoryHandler(historyUseCase usecase.HistoryUseCase) *HistoryHandler {
	return &HistoryHandler{
		historyUseCase: historyUseCase,
	}
}

func (h *HistoryHandler) GetRecentActivities(c *gin.Context) {
	// TODO: Implementation
}
