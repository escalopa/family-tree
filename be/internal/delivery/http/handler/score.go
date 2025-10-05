package handler

import (
	"github.com/escalopa/family-tree-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type ScoreHandler struct {
	scoreUseCase usecase.ScoreUseCase
}

func NewScoreHandler(scoreUseCase usecase.ScoreUseCase) *ScoreHandler {
	return &ScoreHandler{
		scoreUseCase: scoreUseCase,
	}
}

func (h *ScoreHandler) GetLeaderboard(c *gin.Context) {
	// TODO: Implementation
}
