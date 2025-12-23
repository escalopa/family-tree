package handler

import (
	"net/http"

	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type treeHandler struct {
	treeUseCase TreeUseCase
}

func NewTreeHandler(treeUseCase TreeUseCase) *treeHandler {
	return &treeHandler{treeUseCase: treeUseCase}
}

func (h *treeHandler) GetTree(c *gin.Context) {
	var query dto.TreeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	userRole := middleware.GetUserRole(c)

	// Check style
	if query.Style == "list" {
		members, err := h.treeUseCase.GetListView(c.Request.Context(), query.RootID, userRole)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
			return
		}

		var response []dto.MemberResponse
		for _, m := range members {
			spousesDTO := make([]dto.SpouseInfo, len(m.Spouses))
			for i, spouse := range m.Spouses {
				spousesDTO[i] = dto.SpouseInfo{
					SpouseID:     spouse.SpouseID,
					MemberID:     spouse.MemberID,
					ArabicName:   spouse.ArabicName,
					EnglishName:  spouse.EnglishName,
					Gender:       spouse.Gender,
					Picture:      spouse.Picture,
					MarriageDate: dto.FromTimePtr(spouse.MarriageDate),
					DivorceDate:  dto.FromTimePtr(spouse.DivorceDate),
					MarriedYears: dto.CalculateMarriedYears(spouse.MarriageDate, spouse.DivorceDate),
				}
			}

			response = append(response, dto.MemberResponse{
				MemberID:    m.MemberID,
				ArabicName:  m.ArabicName,
				EnglishName: m.EnglishName,
				Gender:      m.Gender,
				Picture:     m.Picture,
				DateOfBirth: dto.FromTimePtr(m.DateOfBirth),
				DateOfDeath: dto.FromTimePtr(m.DateOfDeath),
				FatherID:    m.FatherID,
				MotherID:    m.MotherID,
				Nicknames:   m.Nicknames,
				Profession:  m.Profession,
				Version:     m.Version,
				IsMarried:   m.IsMarried,
				Spouses:     spousesDTO,
			})
		}

		c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
		return
	}

	tree, err := h.treeUseCase.GetTree(c.Request.Context(), query.RootID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: h.convertToTreeResponse(tree)})
}

func (h *treeHandler) GetRelation(c *gin.Context) {
	var query dto.RelationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	userRole := middleware.GetUserRole(c)

	tree, err := h.treeUseCase.GetRelationTree(c.Request.Context(), query.Member1ID, query.Member2ID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: h.convertToTreeResponse(tree)})
}

func (h *treeHandler) convertToTreeResponse(node *domain.MemberTreeNode) *dto.TreeNodeResponse {
	if node == nil {
		return nil
	}

	// Convert spouse information to DTO
	spousesDTO := make([]dto.SpouseInfo, len(node.Spouses))
	for i, spouse := range node.Spouses {
		spousesDTO[i] = dto.SpouseInfo{
			SpouseID:     spouse.SpouseID,
			MemberID:     spouse.MemberID,
			ArabicName:   spouse.ArabicName,
			EnglishName:  spouse.EnglishName,
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
			ArabicName:      node.ArabicName,
			EnglishName:     node.EnglishName,
			Gender:          node.Gender,
			Picture:         node.Picture,
			DateOfBirth:     dto.FromTimePtr(node.DateOfBirth),
			DateOfDeath:     dto.FromTimePtr(node.DateOfDeath),
			FatherID:        node.FatherID,
			MotherID:        node.MotherID,
			Nicknames:       node.Nicknames,
			Profession:      node.Profession,
			Version:         node.Version,
			ArabicFullName:  node.ArabicFullName,
			EnglishFullName: node.EnglishFullName,
			Age:             node.Age,
			GenerationLevel: node.GenerationLevel,
			IsMarried:       node.IsMarried,
			Spouses:         spousesDTO,
		},
		IsInPath: node.IsInPath,
	}

	for _, child := range node.Children {
		response.Children = append(response.Children, h.convertToTreeResponse(child))
	}

	for _, spouse := range node.SpouseNodes {
		response.SpouseNodes = append(response.SpouseNodes, h.convertToTreeResponse(spouse))
	}

	for _, sibling := range node.SiblingNodes {
		response.SiblingNodes = append(response.SiblingNodes, h.convertToTreeResponse(sibling))
	}

	return response
}
