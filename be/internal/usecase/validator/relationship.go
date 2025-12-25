package validator

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/usecase"
)

type RelationshipValidator struct {
	memberRepo usecase.MemberRepository
	spouseRepo usecase.SpouseRepository
}

func NewRelationshipValidator(memberRepo usecase.MemberRepository, spouseRepo usecase.SpouseRepository) *RelationshipValidator {
	return &RelationshipValidator{
		memberRepo: memberRepo,
		spouseRepo: spouseRepo,
	}
}

// CheckParents checks for circular relationships in the family tree
func (v *RelationshipValidator) CheckParents(ctx context.Context, memberID int, fatherID, motherID *int) error {
	if fatherID != nil && *fatherID == memberID {
		return domain.NewValidationError("error.member.circular_relationship")
	}
	if motherID != nil && *motherID == memberID {
		return domain.NewValidationError("error.member.circular_relationship")
	}

	if fatherID != nil {
		if err := v.checkCircularRelationship(ctx, memberID, *fatherID); err != nil {
			return err
		}
	}

	if motherID != nil {
		if err := v.checkCircularRelationship(ctx, memberID, *motherID); err != nil {
			return err
		}
	}

	return nil
}

// checkCircularRelationship checks if parentID is a descendant of memberID (which would create a loop)
func (v *RelationshipValidator) checkCircularRelationship(ctx context.Context, memberID, parentID int) error {
	visited := make(map[int]bool)
	if err := v.checkAncestors(ctx, parentID, memberID, visited, 0); err != nil {
		return err
	}

	return nil
}

// checkAncestors recursively checks if targetID appears in the ancestry of currentID
func (v *RelationshipValidator) checkAncestors(ctx context.Context, currentID, targetID int, visited map[int]bool, depth int) error {
	const maxDepth = 100
	if depth > maxDepth {
		return domain.NewValidationError("error.member.depth_limit")
	}

	if visited[currentID] {
		return nil
	}
	visited[currentID] = true

	current, err := v.memberRepo.Get(ctx, currentID)
	if err != nil {
		return nil // If member not found, skip
	}

	// Check if either parent is the target (circular relationship detected)
	if current.FatherID != nil && *current.FatherID == targetID {
		return domain.NewValidationError("error.member.circular_relationship")
	}
	if current.MotherID != nil && *current.MotherID == targetID {
		return domain.NewValidationError("error.member.circular_relationship")
	}

	// Recursively check ancestors
	if current.FatherID != nil {
		if err := v.checkAncestors(ctx, *current.FatherID, targetID, visited, depth+1); err != nil {
			return err
		}
	}
	if current.MotherID != nil {
		if err := v.checkAncestors(ctx, *current.MotherID, targetID, visited, depth+1); err != nil {
			return err
		}
	}

	return nil
}
