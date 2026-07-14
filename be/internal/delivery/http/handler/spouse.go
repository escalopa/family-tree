package handler

import (
	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type spouseHandler struct {
	spouseUseCase     SpouseUseCase
	memberUseCase     MemberUseCase
	familyTreeUseCase FamilyTreeUseCase
}

func NewSpouseHandler(spouseUseCase SpouseUseCase, memberUseCase MemberUseCase, familyTreeUseCase FamilyTreeUseCase) *spouseHandler {
	return &spouseHandler{
		spouseUseCase:     spouseUseCase,
		memberUseCase:     memberUseCase,
		familyTreeUseCase: familyTreeUseCase,
	}
}

func (h *spouseHandler) requireTreeAccess(c *gin.Context) (int, bool) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return 0, false
	}
	if err := h.familyTreeUseCase.EnsureAccess(c.Request.Context(), uri.TreeID, middleware.GetUserID(c)); err != nil {
		delivery.Error(c, err)
		return 0, false
	}
	return uri.TreeID, true
}

func (h *spouseHandler) requireMemberPairInTree(c *gin.Context, treeID, fatherID, motherID int) bool {
	father, err := h.memberUseCase.Get(c.Request.Context(), fatherID)
	if err != nil {
		delivery.Error(c, err)
		return false
	}
	mother, err := h.memberUseCase.Get(c.Request.Context(), motherID)
	if err != nil {
		delivery.Error(c, err)
		return false
	}
	if father.TreeID != treeID || mother.TreeID != treeID {
		delivery.Error(c, domain.NewNotFoundError("member"))
		return false
	}
	return true
}

func (h *spouseHandler) requireSpouseInTree(c *gin.Context, treeID, spouseID int) (*domain.Spouse, bool) {
	spouse, err := h.spouseUseCase.Get(c.Request.Context(), spouseID)
	if err != nil {
		delivery.Error(c, err)
		return nil, false
	}
	if !h.requireMemberPairInTree(c, treeID, spouse.FatherID, spouse.MotherID) {
		return nil, false
	}
	return spouse, true
}

func (h *spouseHandler) Create(c *gin.Context) {
	treeID, ok := h.requireTreeAccess(c)
	if !ok {
		return
	}

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

	if !h.requireMemberPairInTree(c, treeID, spouse.FatherID, spouse.MotherID) {
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.Create(c.Request.Context(), spouse, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.spouse.created", nil)
}

func (h *spouseHandler) Update(c *gin.Context) {
	treeID, ok := h.requireTreeAccess(c)
	if !ok {
		return
	}

	var uri dto.SpouseIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	if _, ok := h.requireSpouseInTree(c, treeID, uri.SpouseID); !ok {
		return
	}

	var req dto.UpdateSpouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	spouse := &domain.Spouse{
		SpouseID:     uri.SpouseID,
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
	treeID, ok := h.requireTreeAccess(c)
	if !ok {
		return
	}

	var uri dto.SpouseIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	if _, ok := h.requireSpouseInTree(c, treeID, uri.SpouseID); !ok {
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.spouseUseCase.Delete(c.Request.Context(), uri.SpouseID, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.spouse.deleted", nil)
}
