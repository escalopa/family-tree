package handler

import (
	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type treeHandler struct {
	treeUseCase       TreeUseCase
	familyTreeUseCase FamilyTreeUseCase
}

func NewTreeHandler(treeUseCase TreeUseCase, familyTreeUseCase FamilyTreeUseCase) *treeHandler {
	return &treeHandler{treeUseCase: treeUseCase, familyTreeUseCase: familyTreeUseCase}
}

func (h *treeHandler) GetTree(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var query dto.TreeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		delivery.Error(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.familyTreeUseCase.EnsureAccess(c.Request.Context(), uri.TreeID, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	// Check style
	if query.Style == "list" {
		members, err := h.treeUseCase.List(c.Request.Context(), uri.TreeID, query.RootID, userRole)
		if err != nil {
			delivery.Error(c, err)
			return
		}

		// Get user's preferred language from middleware
		preferredLang := middleware.GetPreferredLanguage(c)

		var response []dto.MemberListItem
		for _, m := range members {
			response = append(response, dto.MemberListItem{
				MemberID:    m.MemberID,
				Name:        extractName(m.Names, preferredLang),
				Gender:      m.Gender,
				Picture:     m.Picture,
				DateOfBirth: dto.FromTimePtr(m.DateOfBirth),
				DateOfDeath: dto.FromTimePtr(m.DateOfDeath),
				IsMarried:   m.IsMarried,
			})
		}

		delivery.SuccessWithData(c, response)
		return
	}

	tree, err := h.treeUseCase.Get(c.Request.Context(), uri.TreeID, query.RootID, userRole)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	if tree == nil {
		delivery.SuccessWithData(c, nil)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)
	delivery.SuccessWithData(c, h.convertToTreeResponse(tree, preferredLang))
}

func (h *treeHandler) GetRelation(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var query dto.RelationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		delivery.Error(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.familyTreeUseCase.EnsureAccess(c.Request.Context(), uri.TreeID, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	tree, err := h.treeUseCase.GetRelation(c.Request.Context(), uri.TreeID, query.Member1ID, query.Member2ID, userRole)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	if tree == nil {
		delivery.SuccessWithData(c, nil)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)
	delivery.SuccessWithData(c, h.convertToTreeResponse(tree, preferredLang))
}

func (h *treeHandler) convertToTreeResponse(node *domain.MemberTreeNode, preferredLang string) *dto.TreeNodeResponse {
	if node == nil {
		return nil
	}

	spousesDTO := make([]dto.SpouseInfo, len(node.Spouses))
	for i, spouse := range node.Spouses {
		spousesDTO[i] = dto.SpouseInfo{
			SpouseID:     spouse.SpouseID,
			MemberID:     spouse.MemberID,
			Name:         extractName(spouse.Names, preferredLang),
			Names:        spouse.Names,
			Gender:       spouse.Gender,
			Picture:      spouse.Picture,
			MarriageDate: dto.FromTimePtr(spouse.MarriageDate),
			DivorceDate:  dto.FromTimePtr(spouse.DivorceDate),
			MarriedYears: dto.CalculateMarriedYears(spouse.MarriageDate, spouse.DivorceDate),
		}
	}

	response := &dto.TreeNodeResponse{
		Member: dto.MemberResponse{
			MemberID:        node.MemberID,
			Name:            extractName(node.Names, preferredLang),
			Names:           node.Names,
			FullName:        extractName(node.FullNames, preferredLang),
			FullNames:       node.FullNames,
			Gender:          node.Gender,
			Picture:         node.Picture,
			DateOfBirth:     dto.FromTimePtr(node.DateOfBirth),
			DateOfDeath:     dto.FromTimePtr(node.DateOfDeath),
			FatherID:        node.FatherID,
			MotherID:        node.MotherID,
			Nicknames:       node.Nicknames,
			Profession:      node.Profession,
			Version:         node.Version,
			Age:             node.Age,
			GenerationLevel: node.GenerationLevel,
			IsMarried:       node.IsMarried,
			Spouses:         spousesDTO,
		},
		IsInPath: node.IsInPath,
	}

	for _, child := range node.Children {
		response.Children = append(response.Children, h.convertToTreeResponse(child, preferredLang))
	}

	return response
}
