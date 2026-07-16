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

func (h *treeHandler) GetGraph(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.familyTreeUseCase.EnsureAccess(c.Request.Context(), uri.TreeID, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	graph, err := h.treeUseCase.GetGraph(c.Request.Context(), uri.TreeID, userRole)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)
	delivery.SuccessWithData(c, h.convertToGraphResponse(graph, preferredLang))
}

func (h *treeHandler) GetRelationGraph(c *gin.Context) {
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

	graph, err := h.treeUseCase.GetRelationGraph(c.Request.Context(), uri.TreeID, query.Member1ID, query.Member2ID, userRole)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)
	delivery.SuccessWithData(c, h.convertToGraphResponse(graph, preferredLang))
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

func (h *treeHandler) convertToGraphResponse(graph *domain.FamilyGraph, preferredLang string) dto.FamilyGraphResponse {
	if graph == nil {
		return dto.FamilyGraphResponse{}
	}

	response := dto.FamilyGraphResponse{
		People:            make([]dto.FamilyGraphPersonResponse, 0, len(graph.People)),
		FamilyUnits:       make([]dto.FamilyGraphUnitResponse, 0, len(graph.FamilyUnits)),
		Edges:             make([]dto.FamilyGraphEdgeResponse, 0, len(graph.Edges)),
		References:        make([]dto.FamilyGraphReferenceResponse, 0, len(graph.References)),
		PathPersonIDs:     graph.PathPersonIDs,
		PathFamilyUnitIDs: graph.PathFamilyUnitIDs,
	}

	for _, person := range graph.People {
		response.People = append(response.People, dto.FamilyGraphPersonResponse{
			Member: dto.MemberResponse{
				MemberID:        person.MemberID,
				TreeID:          person.TreeID,
				Name:            extractName(person.Names, preferredLang),
				Names:           person.Names,
				FullName:        extractName(person.FullNames, preferredLang),
				FullNames:       person.FullNames,
				Gender:          person.Gender,
				Picture:         person.Picture,
				DateOfBirth:     dto.FromTimePtr(person.DateOfBirth),
				DateOfDeath:     dto.FromTimePtr(person.DateOfDeath),
				FatherID:        person.FatherID,
				MotherID:        person.MotherID,
				Nicknames:       person.Nicknames,
				Profession:      person.Profession,
				Version:         person.Version,
				Age:             person.Age,
				GenerationLevel: person.GenerationLevel,
				IsMarried:       person.IsMarried,
			},
			ParentFamilyUnitIDs:  person.ParentFamilyUnitIDs,
			PartnerFamilyUnitIDs: person.PartnerFamilyUnitIDs,
			IsReferenceCandidate: person.IsReferenceCandidate,
			IsInPath:             person.IsInPath,
		})
	}

	for _, unit := range graph.FamilyUnits {
		response.FamilyUnits = append(response.FamilyUnits, dto.FamilyGraphUnitResponse{
			FamilyUnitID:     unit.FamilyUnitID,
			TreeID:           unit.TreeID,
			RelationshipType: unit.RelationshipType,
			Status:           unit.Status,
			StartDate:        dto.FromTimePtr(unit.StartDate),
			EndDate:          dto.FromTimePtr(unit.EndDate),
			PartnerIDs:       unit.PartnerIDs,
			ChildIDs:         unit.ChildIDs,
		})
	}

	for _, edge := range graph.Edges {
		response.Edges = append(response.Edges, dto.FamilyGraphEdgeResponse{
			EdgeID:       edge.EdgeID,
			SourceID:     edge.SourceID,
			TargetID:     edge.TargetID,
			Type:         edge.Type,
			RelationType: edge.RelationType,
			Status:       edge.Status,
			IsInPath:     edge.IsInPath,
		})
	}

	for _, ref := range graph.References {
		response.References = append(response.References, dto.FamilyGraphReferenceResponse{
			ReferenceID:  ref.ReferenceID,
			PersonID:     ref.PersonID,
			FamilyUnitID: ref.FamilyUnitID,
			Reason:       ref.Reason,
		})
	}

	return response
}
