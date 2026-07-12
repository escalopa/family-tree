package handler

import (
	"fmt"
	"strings"

	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type familyTreeHandler struct {
	treeUseCase        FamilyTreeUseCase
	treeDiagramUseCase TreeUseCase
}

func NewFamilyTreeHandler(treeUseCase FamilyTreeUseCase, treeDiagramUseCase TreeUseCase) *familyTreeHandler {
	return &familyTreeHandler{treeUseCase: treeUseCase, treeDiagramUseCase: treeDiagramUseCase}
}

func (h *familyTreeHandler) Create(c *gin.Context) {
	var req dto.CreateFamilyTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	tree := &domain.FamilyTree{Name: req.Name, Description: req.Description}
	if err := h.treeUseCase.Create(c.Request.Context(), tree, middleware.GetUserID(c)); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.SuccessWithData(c, toFamilyTreeResponse(tree))
}

func (h *familyTreeHandler) List(c *gin.Context) {
	trees, err := h.treeUseCase.List(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		delivery.Error(c, err)
		return
	}
	response := dto.FamilyTreeListResponse{Trees: make([]dto.FamilyTreeResponse, 0, len(trees))}
	for _, tree := range trees {
		response.Trees = append(response.Trees, toFamilyTreeResponse(tree))
	}
	delivery.SuccessWithData(c, response)
}

func (h *familyTreeHandler) Get(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	tree, err := h.treeUseCase.Get(c.Request.Context(), uri.TreeID, middleware.GetUserID(c))
	if err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.SuccessWithData(c, toFamilyTreeResponse(tree))
}

func (h *familyTreeHandler) ListMyInvitations(c *gin.Context) {
	invitations, err := h.treeUseCase.ListMyInvitations(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.SuccessWithData(c, toInvitationListResponse(invitations))
}

func (h *familyTreeHandler) Invite(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	var req dto.InviteToTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}
	invitation, err := h.treeUseCase.Invite(
		c.Request.Context(),
		uri.TreeID,
		middleware.GetUserID(c),
		req.Email,
		req.Message,
		req.ExpiresAt.ToTimePtr(),
	)
	if err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.SuccessWithData(c, toInvitationResponse(invitation))
}

func (h *familyTreeHandler) ListInvitations(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	invitations, err := h.treeUseCase.ListTreeInvitations(c.Request.Context(), uri.TreeID, middleware.GetUserID(c))
	if err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.SuccessWithData(c, toInvitationListResponse(invitations))
}

func (h *familyTreeHandler) AcceptInvitation(c *gin.Context) {
	var uri dto.InvitationIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	if err := h.treeUseCase.AcceptInvitation(c.Request.Context(), uri.InvitationID, middleware.GetUserID(c)); err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.Success(c, "success.invitation.accepted", nil)
}

func (h *familyTreeHandler) DeclineInvitation(c *gin.Context) {
	var uri dto.InvitationIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	if err := h.treeUseCase.DeclineInvitation(c.Request.Context(), uri.InvitationID, middleware.GetUserID(c)); err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.Success(c, "success.invitation.declined", nil)
}

func (h *familyTreeHandler) CreateShareLink(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	var req dto.CreateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}
	link, err := h.treeUseCase.CreateShareLink(c.Request.Context(), uri.TreeID, middleware.GetUserID(c), req.ExpiresAt.ToTimePtr(), req.MaxVisits)
	if err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.SuccessWithData(c, toShareLinkResponse(link, publicShareBaseURL(c)))
}

func (h *familyTreeHandler) ListShareLinks(c *gin.Context) {
	var uri dto.TreeIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	links, err := h.treeUseCase.ListShareLinks(c.Request.Context(), uri.TreeID, middleware.GetUserID(c))
	if err != nil {
		delivery.Error(c, err)
		return
	}
	response := dto.FamilyTreeShareLinkListResponse{ShareLinks: make([]dto.FamilyTreeShareLinkResponse, 0, len(links))}
	baseURL := publicShareBaseURL(c)
	for _, link := range links {
		response.ShareLinks = append(response.ShareLinks, toShareLinkResponse(link, baseURL))
	}
	delivery.SuccessWithData(c, response)
}

func (h *familyTreeHandler) RevokeShareLink(c *gin.Context) {
	var uri dto.ShareIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	if err := h.treeUseCase.RevokeShareLink(c.Request.Context(), uri.TreeID, uri.ShareID, middleware.GetUserID(c)); err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.Success(c, "success.share_link.revoked", nil)
}

func (h *familyTreeHandler) GetPublicTree(c *gin.Context) {
	var uri dto.PublicShareUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}
	link, err := h.treeUseCase.ConsumeShareLink(c.Request.Context(), uri.Token)
	if err != nil {
		delivery.Error(c, err)
		return
	}
	tree, err := h.treeDiagramUseCase.Get(c.Request.Context(), link.TreeID, nil, domain.RoleGuest)
	if err != nil {
		delivery.Error(c, err)
		return
	}
	delivery.SuccessWithData(c, dto.PublicTreeResponse{
		Share: toShareLinkResponse(link, publicShareBaseURL(c)),
		Tree:  h.convertToTreeResponse(tree, "en"),
	})
}

func (h *familyTreeHandler) convertToTreeResponse(node *domain.MemberTreeNode, preferredLang string) *dto.TreeNodeResponse {
	treeHandler := &treeHandler{}
	return treeHandler.convertToTreeResponse(node, preferredLang)
}

func toFamilyTreeResponse(tree *domain.FamilyTree) dto.FamilyTreeResponse {
	return dto.FamilyTreeResponse{
		TreeID:      tree.TreeID,
		Name:        tree.Name,
		Description: tree.Description,
		OwnerUserID: tree.OwnerUserID,
		OwnerName:   tree.OwnerName,
		OwnerEmail:  tree.OwnerEmail,
		UserRole:    tree.UserRole,
		MemberCount: tree.MemberCount,
		CreatedAt:   tree.CreatedAt,
		UpdatedAt:   tree.UpdatedAt,
	}
}

func toInvitationListResponse(invitations []*domain.FamilyTreeInvitation) dto.FamilyTreeInvitationListResponse {
	response := dto.FamilyTreeInvitationListResponse{Invitations: make([]dto.FamilyTreeInvitationResponse, 0, len(invitations))}
	for _, invitation := range invitations {
		response.Invitations = append(response.Invitations, toInvitationResponse(invitation))
	}
	return response
}

func toInvitationResponse(invitation *domain.FamilyTreeInvitation) dto.FamilyTreeInvitationResponse {
	return dto.FamilyTreeInvitationResponse{
		InvitationID:  invitation.InvitationID,
		TreeID:        invitation.TreeID,
		TreeName:      invitation.TreeName,
		InviterUserID: invitation.InviterUserID,
		InviterName:   invitation.InviterName,
		InviteeUserID: invitation.InviteeUserID,
		InviteeEmail:  invitation.InviteeEmail,
		Status:        invitation.Status,
		Message:       invitation.Message,
		CreatedAt:     invitation.CreatedAt,
		ExpiresAt:     invitation.ExpiresAt,
		RespondedAt:   invitation.RespondedAt,
	}
}

func toShareLinkResponse(link *domain.FamilyTreeShareLink, baseURL string) dto.FamilyTreeShareLinkResponse {
	return dto.FamilyTreeShareLinkResponse{
		ShareID:    link.ShareID,
		TreeID:     link.TreeID,
		Token:      link.Token,
		URL:        fmt.Sprintf("%s/public/trees/%s", strings.TrimRight(baseURL, "/"), link.Token),
		CreatedBy:  link.CreatedBy,
		CreatedAt:  link.CreatedAt,
		ExpiresAt:  link.ExpiresAt,
		MaxVisits:  link.MaxVisits,
		VisitCount: link.VisitCount,
		RevokedAt:  link.RevokedAt,
	}
}

func publicShareBaseURL(c *gin.Context) string {
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
		if c.Request.TLS != nil {
			proto = "https"
		}
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return proto + "://" + host
}
