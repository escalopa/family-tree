package handler

import (
	"github.com/escalopa/family-tree-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase    usecase.UserUseCase
	historyUseCase usecase.HistoryUseCase
	scoreUseCase   usecase.ScoreUseCase
}

func NewUserHandler(
	userUseCase usecase.UserUseCase,
	historyUseCase usecase.HistoryUseCase,
	scoreUseCase usecase.ScoreUseCase,
) *UserHandler {
	return &UserHandler{
		userUseCase:    userUseCase,
		historyUseCase: historyUseCase,
		scoreUseCase:   scoreUseCase,
	}
}

func (h *UserHandler) List(c *gin.Context) {
	// TODO: Implementation
}

func (h *UserHandler) GetByID(c *gin.Context) {
	// TODO: Implementation
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// TODO: Implementation
}

func (h *UserHandler) GetRecentActivities(c *gin.Context) {
	// TODO: Implementation
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	// TODO: Implementation
}

func (h *UserHandler) UpdateActiveStatus(c *gin.Context) {
	// TODO: Implementation
}

func (h *UserHandler) AdminLogout(c *gin.Context) {
	// TODO: Implementation
}

func (h *UserHandler) GetUserScores(c *gin.Context) {
	// TODO: Implementation
}
