package handler

import (
	"net/http"

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

func (h *spouseHandler) AddSpouse(c *gin.Context) {
	var req dto.CreateSpouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	spouse := &domain.Spouse{
		FatherID:     req.FatherID,
		MotherID:     req.MotherID,
		MarriageDate: req.MarriageDate.ToTimePtr(),
		DivorceDate:  req.DivorceDate.ToTimePtr(),
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.AddSpouse(c.Request.Context(), spouse, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{Success: true, Data: "spouse relationship created"})
}

func (h *spouseHandler) UpdateSpouse(c *gin.Context) {
	var req dto.UpdateSpouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	spouse := &domain.Spouse{
		FatherID:     req.FatherID,
		MotherID:     req.MotherID,
		MarriageDate: req.MarriageDate.ToTimePtr(),
		DivorceDate:  req.DivorceDate.ToTimePtr(),
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.UpdateSpouse(c.Request.Context(), spouse, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "spouse relationship updated"})
}

func (h *spouseHandler) UpdateSpouseByID(c *gin.Context) {
	var req dto.UpdateSpouseByMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	spouse := &domain.Spouse{
		SpouseID:     req.SpouseID,
		MarriageDate: req.MarriageDate.ToTimePtr(),
		DivorceDate:  req.DivorceDate.ToTimePtr(),
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.UpdateSpouseByID(c.Request.Context(), spouse, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "spouse relationship updated"})
}

func (h *spouseHandler) RemoveSpouse(c *gin.Context) {
	var req struct {
		SpouseID int `json:"spouse_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.RemoveSpouseByID(c.Request.Context(), req.SpouseID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "spouse relationship removed"})
}
