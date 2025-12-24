package handler

import (
	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type spouseHandler struct {
	spouseUseCase SpouseUseCase
}

func NewSpouseHandler(spouseUseCase SpouseUseCase) *spouseHandler {
	return &spouseHandler{spouseUseCase: spouseUseCase}
}

func (h *spouseHandler) Create(c *gin.Context) {
	var req dto.CreateSpouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	spouse := &domain.Spouse{
		FatherID:     req.FatherID,
		MotherID:     req.MotherID,
		MarriageDate: req.MarriageDate.ToTimePtr(),
		DivorceDate:  req.DivorceDate.ToTimePtr(),
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.Create(c.Request.Context(), spouse, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.spouse.created", nil)
}

func (h *spouseHandler) Update(c *gin.Context) {
	var req dto.UpdateSpouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	spouse := &domain.Spouse{
		SpouseID:     req.SpouseID,
		MarriageDate: req.MarriageDate.ToTimePtr(),
		DivorceDate:  req.DivorceDate.ToTimePtr(),
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.Update(c.Request.Context(), spouse, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.spouse.updated", nil)
}

func (h *spouseHandler) Delete(c *gin.Context) {
	var req dto.DeleteSpouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.Delete(c.Request.Context(), req.SpouseID, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.spouse.deleted", nil)
}
