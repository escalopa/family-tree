package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
)

type familyTreeUseCaseRepo struct {
	tree FamilyTreeRepository
	user UserRepository
}

type familyTreeUseCase struct {
	repo familyTreeUseCaseRepo
}

func NewFamilyTreeUseCase(treeRepo FamilyTreeRepository, userRepo UserRepository) *familyTreeUseCase {
	return &familyTreeUseCase{
		repo: familyTreeUseCaseRepo{
			tree: treeRepo,
			user: userRepo,
		},
	}
}

func (uc *familyTreeUseCase) Create(ctx context.Context, tree *domain.FamilyTree, userID int) error {
	tree.Name = strings.TrimSpace(tree.Name)
	if tree.Name == "" {
		return domain.NewValidationError("error.validation.name_required")
	}
	tree.OwnerUserID = userID
	return uc.repo.tree.Create(ctx, tree)
}

func (uc *familyTreeUseCase) List(ctx context.Context, userID int) ([]*domain.FamilyTree, error) {
	return uc.repo.tree.ListForUser(ctx, userID)
}

func (uc *familyTreeUseCase) Get(ctx context.Context, treeID, userID int) (*domain.FamilyTree, error) {
	return uc.repo.tree.GetForUser(ctx, treeID, userID)
}

func (uc *familyTreeUseCase) EnsureAccess(ctx context.Context, treeID, userID int) error {
	ok, err := uc.repo.tree.HasAccess(ctx, treeID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.NewNotFoundError("family_tree")
	}
	return nil
}

func (uc *familyTreeUseCase) Invite(ctx context.Context, treeID, inviterUserID int, inviteeEmail string, message *string, expiresAt *time.Time) (*domain.FamilyTreeInvitation, error) {
	if err := uc.EnsureAccess(ctx, treeID, inviterUserID); err != nil {
		return nil, err
	}

	inviteeEmail = strings.ToLower(strings.TrimSpace(inviteeEmail))
	if inviteeEmail == "" {
		return nil, domain.NewValidationError("error.validation.email_required")
	}

	invitee, err := uc.repo.user.GetByEmail(ctx, inviteeEmail)
	if err != nil {
		if domain.IsDomainError(err, domain.ErrCodeNotFound) {
			return nil, domain.NewNotFoundError("user")
		}
		return nil, err
	}

	if expiresAt == nil {
		defaultExpiry := time.Now().Add(14 * 24 * time.Hour)
		expiresAt = &defaultExpiry
	}

	invitation := &domain.FamilyTreeInvitation{
		TreeID:        treeID,
		InviterUserID: inviterUserID,
		InviteeUserID: &invitee.UserID,
		InviteeEmail:  invitee.Email,
		Message:       message,
		ExpiresAt:     expiresAt,
	}
	if err := uc.repo.tree.CreateInvitation(ctx, invitation); err != nil {
		return nil, err
	}
	return invitation, nil
}

func (uc *familyTreeUseCase) ListTreeInvitations(ctx context.Context, treeID, userID int) ([]*domain.FamilyTreeInvitation, error) {
	return uc.repo.tree.ListTreeInvitations(ctx, treeID, userID)
}

func (uc *familyTreeUseCase) ListMyInvitations(ctx context.Context, userID int) ([]*domain.FamilyTreeInvitation, error) {
	return uc.repo.tree.ListPendingInvitationsForUser(ctx, userID)
}

func (uc *familyTreeUseCase) AcceptInvitation(ctx context.Context, invitationID, userID int) error {
	return uc.repo.tree.RespondToInvitation(ctx, invitationID, userID, true)
}

func (uc *familyTreeUseCase) DeclineInvitation(ctx context.Context, invitationID, userID int) error {
	return uc.repo.tree.RespondToInvitation(ctx, invitationID, userID, false)
}

func (uc *familyTreeUseCase) CreateShareLink(ctx context.Context, treeID, userID int, expiresAt *time.Time, maxVisits *int) (*domain.FamilyTreeShareLink, error) {
	if err := uc.EnsureAccess(ctx, treeID, userID); err != nil {
		return nil, err
	}
	if maxVisits != nil && *maxVisits <= 0 {
		return nil, domain.NewValidationError("error.validation.max_visits_positive")
	}
	link := &domain.FamilyTreeShareLink{
		TreeID:    treeID,
		CreatedBy: userID,
		ExpiresAt: expiresAt,
		MaxVisits: maxVisits,
	}
	if err := uc.repo.tree.CreateShareLink(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (uc *familyTreeUseCase) ListShareLinks(ctx context.Context, treeID, userID int) ([]*domain.FamilyTreeShareLink, error) {
	return uc.repo.tree.ListShareLinks(ctx, treeID, userID)
}

func (uc *familyTreeUseCase) RevokeShareLink(ctx context.Context, treeID, shareID, userID int) error {
	return uc.repo.tree.RevokeShareLink(ctx, treeID, shareID, userID)
}

func (uc *familyTreeUseCase) ConsumeShareLink(ctx context.Context, token string) (*domain.FamilyTreeShareLink, error) {
	return uc.repo.tree.ConsumeShareLink(ctx, token)
}
