package handler

import (
	"github.com/escalopa/family-tree-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	memberUseCase  usecase.MemberUseCase
	historyUseCase usecase.HistoryUseCase
}

func NewMemberHandler(
	memberUseCase usecase.MemberUseCase,
	historyUseCase usecase.HistoryUseCase,
) *MemberHandler {
	return &MemberHandler{
		memberUseCase:  memberUseCase,
		historyUseCase: historyUseCase,
	}
}

func (h *MemberHandler) Search(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) Create(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) GetByID(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) Update(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) Patch(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) Delete(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) GetPicture(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) UploadPicture(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) DeletePicture(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) GetPictures(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) Rollback(c *gin.Context) {
	// TODO: Implementation
}

func (h *MemberHandler) GetHistory(c *gin.Context) {
	// TODO: Implementation
}
