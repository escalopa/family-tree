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
			response = append(response, dto.MemberResponse{
				MemberID:    m.MemberID,
				ArabicName:  m.ArabicName,
				EnglishName: m.EnglishName,
				Gender:      m.Gender,
				Picture:     m.Picture,
				DateOfBirth: m.DateOfBirth,
				DateOfDeath: m.DateOfDeath,
				FatherID:    m.FatherID,
				MotherID:    m.MotherID,
				Nicknames:   m.Nicknames,
				Profession:  m.Profession,
				Version:     m.Version,
				IsMarried:   m.IsMarried,
				Spouses:     m.Spouses,
			})
		}

		c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
		return
	}

	// Default: tree view
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
	path, err := h.treeUseCase.GetRelationPath(c.Request.Context(), query.Member1ID, query.Member2ID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var response []dto.MemberResponse
	for _, m := range path {
		response = append(response, dto.MemberResponse{
			MemberID:    m.MemberID,
			ArabicName:  m.ArabicName,
			EnglishName: m.EnglishName,
			Gender:      m.Gender,
			Picture:     m.Picture,
			DateOfBirth: m.DateOfBirth,
			DateOfDeath: m.DateOfDeath,
			FatherID:    m.FatherID,
			MotherID:    m.MotherID,
			Nicknames:   m.Nicknames,
			Profession:  m.Profession,
			Version:     m.Version,
			IsMarried:   m.IsMarried,
			Spouses:     m.Spouses,
		})
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *treeHandler) convertToTreeResponse(node *domain.MemberTreeNode) *dto.TreeNodeResponse {
	if node == nil {
		return nil
	}

	response := &dto.TreeNodeResponse{
		Member: dto.MemberResponse{
			MemberID:        node.MemberID,
			ArabicName:      node.ArabicName,
			EnglishName:     node.EnglishName,
			Gender:          node.Gender,
			Picture:         node.Picture,
			DateOfBirth:     node.DateOfBirth,
			DateOfDeath:     node.DateOfDeath,
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
			Spouses:         node.Spouses,
		},
	}

	for _, child := range node.Children {
		response.Children = append(response.Children, h.convertToTreeResponse(child))
	}

	return response
}
