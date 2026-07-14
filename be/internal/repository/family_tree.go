package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FamilyTreeRepository struct {
	db *pgxpool.Pool
}

func NewFamilyTreeRepository(db *pgxpool.Pool) *FamilyTreeRepository {
	return &FamilyTreeRepository{db: db}
}

func (r *FamilyTreeRepository) Create(ctx context.Context, tree *domain.FamilyTree) error {
	return doWithQuerier(ctx, r.db, func(txCtx context.Context) error {
		querier := getQuerier(txCtx, r.db)
		query := `
			INSERT INTO family_trees (name, description, owner_user_id)
			VALUES ($1, $2, $3)
			RETURNING tree_id, created_at, updated_at
		`
		if err := querier.QueryRow(txCtx, query, tree.Name, tree.Description, tree.OwnerUserID).
			Scan(&tree.TreeID, &tree.CreatedAt, &tree.UpdatedAt); err != nil {
			return domain.NewDatabaseError(err)
		}

		membershipQuery := `
			INSERT INTO family_tree_memberships (tree_id, user_id, role)
			VALUES ($1, $2, $3)
		`
		if _, err := querier.Exec(txCtx, membershipQuery, tree.TreeID, tree.OwnerUserID, domain.TreeRoleOwner); err != nil {
			return domain.NewDatabaseError(err)
		}
		tree.UserRole = domain.TreeRoleOwner
		return nil
	})
}

func (r *FamilyTreeRepository) ListForUser(ctx context.Context, userID int) ([]*domain.FamilyTree, error) {
	query := `
		SELECT ft.tree_id, ft.name, ft.description, ft.owner_user_id,
		       owner.full_name, owner.email, ftm.role,
		       COUNT(m.member_id) FILTER (WHERE m.deleted_at IS NULL) AS member_count,
		       ft.created_at, ft.updated_at
		FROM family_trees ft
		JOIN family_tree_memberships ftm ON ftm.tree_id = ft.tree_id
		JOIN users owner ON owner.user_id = ft.owner_user_id
		LEFT JOIN members m ON m.tree_id = ft.tree_id
		WHERE ftm.user_id = $1
		GROUP BY ft.tree_id, owner.full_name, owner.email, ftm.role
		ORDER BY ft.updated_at DESC, ft.tree_id DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var trees []*domain.FamilyTree
	for rows.Next() {
		tree := &domain.FamilyTree{}
		if err := rows.Scan(
			&tree.TreeID,
			&tree.Name,
			&tree.Description,
			&tree.OwnerUserID,
			&tree.OwnerName,
			&tree.OwnerEmail,
			&tree.UserRole,
			&tree.MemberCount,
			&tree.CreatedAt,
			&tree.UpdatedAt,
		); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		trees = append(trees, tree)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return trees, nil
}

func (r *FamilyTreeRepository) GetForUser(ctx context.Context, treeID, userID int) (*domain.FamilyTree, error) {
	query := `
		SELECT ft.tree_id, ft.name, ft.description, ft.owner_user_id,
		       owner.full_name, owner.email, ftm.role,
		       COUNT(m.member_id) FILTER (WHERE m.deleted_at IS NULL) AS member_count,
		       ft.created_at, ft.updated_at
		FROM family_trees ft
		JOIN family_tree_memberships ftm ON ftm.tree_id = ft.tree_id
		JOIN users owner ON owner.user_id = ft.owner_user_id
		LEFT JOIN members m ON m.tree_id = ft.tree_id
		WHERE ft.tree_id = $1 AND ftm.user_id = $2
		GROUP BY ft.tree_id, owner.full_name, owner.email, ftm.role
	`
	tree := &domain.FamilyTree{}
	err := r.db.QueryRow(ctx, query, treeID, userID).Scan(
		&tree.TreeID,
		&tree.Name,
		&tree.Description,
		&tree.OwnerUserID,
		&tree.OwnerName,
		&tree.OwnerEmail,
		&tree.UserRole,
		&tree.MemberCount,
		&tree.CreatedAt,
		&tree.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("family_tree")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return tree, nil
}

func (r *FamilyTreeRepository) HasAccess(ctx context.Context, treeID, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM family_tree_memberships WHERE tree_id = $1 AND user_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, treeID, userID).Scan(&exists); err != nil {
		return false, domain.NewDatabaseError(err)
	}
	return exists, nil
}

func (r *FamilyTreeRepository) CreateInvitation(ctx context.Context, invitation *domain.FamilyTreeInvitation) error {
	query := `
		INSERT INTO family_tree_invitations (
			tree_id, inviter_user_id, invitee_user_id, invitee_email, message, expires_at
		)
		VALUES ($1, $2, $3, LOWER($4), $5, $6)
		RETURNING invitation_id, status, created_at
	`
	if err := r.db.QueryRow(ctx, query,
		invitation.TreeID,
		invitation.InviterUserID,
		invitation.InviteeUserID,
		strings.TrimSpace(invitation.InviteeEmail),
		invitation.Message,
		invitation.ExpiresAt,
	).Scan(&invitation.InvitationID, &invitation.Status, &invitation.CreatedAt); err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *FamilyTreeRepository) ListTreeInvitations(ctx context.Context, treeID, userID int) ([]*domain.FamilyTreeInvitation, error) {
	if ok, err := r.HasAccess(ctx, treeID, userID); err != nil {
		return nil, err
	} else if !ok {
		return nil, domain.NewNotFoundError("family_tree")
	}

	query := `
		SELECT i.invitation_id, i.tree_id, ft.name, i.inviter_user_id, inviter.full_name,
		       i.invitee_user_id, i.invitee_email, i.status, i.message,
		       i.created_at, i.expires_at, i.responded_at
		FROM family_tree_invitations i
		JOIN family_trees ft ON ft.tree_id = i.tree_id
		JOIN users inviter ON inviter.user_id = i.inviter_user_id
		WHERE i.tree_id = $1
		ORDER BY i.created_at DESC
	`
	return r.scanInvitations(ctx, query, treeID)
}

func (r *FamilyTreeRepository) ListPendingInvitationsForUser(ctx context.Context, userID int) ([]*domain.FamilyTreeInvitation, error) {
	query := `
		SELECT i.invitation_id, i.tree_id, ft.name, i.inviter_user_id, inviter.full_name,
		       i.invitee_user_id, i.invitee_email, i.status, i.message,
		       i.created_at, i.expires_at, i.responded_at
		FROM family_tree_invitations i
		JOIN family_trees ft ON ft.tree_id = i.tree_id
		JOIN users inviter ON inviter.user_id = i.inviter_user_id
		WHERE i.invitee_user_id = $1
		  AND i.status = 'pending'
		  AND (i.expires_at IS NULL OR i.expires_at > NOW())
		ORDER BY i.created_at DESC
	`
	return r.scanInvitations(ctx, query, userID)
}

func (r *FamilyTreeRepository) scanInvitations(ctx context.Context, query string, args ...any) ([]*domain.FamilyTreeInvitation, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var invitations []*domain.FamilyTreeInvitation
	for rows.Next() {
		invitation := &domain.FamilyTreeInvitation{}
		if err := rows.Scan(
			&invitation.InvitationID,
			&invitation.TreeID,
			&invitation.TreeName,
			&invitation.InviterUserID,
			&invitation.InviterName,
			&invitation.InviteeUserID,
			&invitation.InviteeEmail,
			&invitation.Status,
			&invitation.Message,
			&invitation.CreatedAt,
			&invitation.ExpiresAt,
			&invitation.RespondedAt,
		); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		invitations = append(invitations, invitation)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return invitations, nil
}

func (r *FamilyTreeRepository) RespondToInvitation(ctx context.Context, invitationID, userID int, accept bool) error {
	status := domain.InvitationStatusDeclined
	if accept {
		status = domain.InvitationStatusAccepted
	}

	return doWithQuerier(ctx, r.db, func(txCtx context.Context) error {
		querier := getQuerier(txCtx, r.db)
		var treeID int
		query := `
			UPDATE family_tree_invitations
			SET status = $1, responded_at = NOW()
			WHERE invitation_id = $2
			  AND invitee_user_id = $3
			  AND status = 'pending'
			  AND (expires_at IS NULL OR expires_at > NOW())
			RETURNING tree_id
		`
		if err := querier.QueryRow(txCtx, query, status, invitationID, userID).Scan(&treeID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.NewNotFoundError("invitation")
			}
			return domain.NewDatabaseError(err)
		}

		if !accept {
			return nil
		}

		membershipQuery := `
			INSERT INTO family_tree_memberships (tree_id, user_id, role)
			VALUES ($1, $2, $3)
			ON CONFLICT (tree_id, user_id) DO NOTHING
		`
		if _, err := querier.Exec(txCtx, membershipQuery, treeID, userID, domain.TreeRoleEditor); err != nil {
			return domain.NewDatabaseError(err)
		}
		return nil
	})
}

func (r *FamilyTreeRepository) CreateShareLink(ctx context.Context, link *domain.FamilyTreeShareLink) error {
	token, err := randomShareToken()
	if err != nil {
		return domain.NewInternalError(err)
	}
	link.Token = token

	query := `
		INSERT INTO family_tree_share_links (tree_id, token, created_by, expires_at, max_visits)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING share_id, created_at, visit_count
	`
	if err := r.db.QueryRow(ctx, query, link.TreeID, link.Token, link.CreatedBy, link.ExpiresAt, link.MaxVisits).
		Scan(&link.ShareID, &link.CreatedAt, &link.VisitCount); err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *FamilyTreeRepository) ListShareLinks(ctx context.Context, treeID, userID int) ([]*domain.FamilyTreeShareLink, error) {
	if ok, err := r.HasAccess(ctx, treeID, userID); err != nil {
		return nil, err
	} else if !ok {
		return nil, domain.NewNotFoundError("family_tree")
	}

	query := `
		SELECT share_id, tree_id, token, created_by, created_at, expires_at, max_visits, visit_count, revoked_at
		FROM family_tree_share_links
		WHERE tree_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, treeID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var links []*domain.FamilyTreeShareLink
	for rows.Next() {
		link := &domain.FamilyTreeShareLink{}
		if err := rows.Scan(&link.ShareID, &link.TreeID, &link.Token, &link.CreatedBy, &link.CreatedAt, &link.ExpiresAt, &link.MaxVisits, &link.VisitCount, &link.RevokedAt); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return links, nil
}

func (r *FamilyTreeRepository) RevokeShareLink(ctx context.Context, treeID, shareID, userID int) error {
	if ok, err := r.HasAccess(ctx, treeID, userID); err != nil {
		return err
	} else if !ok {
		return domain.NewNotFoundError("family_tree")
	}

	query := `
		UPDATE family_tree_share_links
		SET revoked_at = NOW()
		WHERE tree_id = $1 AND share_id = $2 AND revoked_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, treeID, shareID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		return domain.NewNotFoundError("share_link")
	}
	return nil
}

func (r *FamilyTreeRepository) ConsumeShareLink(ctx context.Context, token string) (*domain.FamilyTreeShareLink, error) {
	query := `
		UPDATE family_tree_share_links
		SET visit_count = visit_count + 1
		WHERE token = $1
		  AND revoked_at IS NULL
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND (max_visits IS NULL OR visit_count < max_visits)
		RETURNING share_id, tree_id, token, created_by, created_at, expires_at, max_visits, visit_count, revoked_at
	`
	link := &domain.FamilyTreeShareLink{}
	err := r.db.QueryRow(ctx, query, token).Scan(&link.ShareID, &link.TreeID, &link.Token, &link.CreatedBy, &link.CreatedAt, &link.ExpiresAt, &link.MaxVisits, &link.VisitCount, &link.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("share_link")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return link, nil
}

func randomShareToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func DefaultInvitationExpiry() *time.Time {
	expiresAt := time.Now().Add(14 * 24 * time.Hour)
	return &expiresAt
}
