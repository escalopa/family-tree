package validator

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/usecase"
)

// MarriageValidator implements marriage prohibition validation
type MarriageValidator struct {
	memberRepo usecase.MemberRepository
	spouseRepo usecase.SpouseRepository
}

// NewMarriageValidator creates a new marriage validator
func NewMarriageValidator(memberRepo usecase.MemberRepository, spouseRepo usecase.SpouseRepository) usecase.MarriageValidator {
	return &MarriageValidator{
		memberRepo: memberRepo,
		spouseRepo: spouseRepo,
	}
}

// Create validates that two people can be married according to Islamic rules
func (v *MarriageValidator) Create(ctx context.Context, memberAID, memberBID int) error {
	// Fetch both members
	personA, err := v.memberRepo.Get(ctx, memberAID)
	if err != nil {
		return err
	}
	personB, err := v.memberRepo.Get(ctx, memberBID)
	if err != nil {
		return err
	}

	// 1. Blood Relationships (Nasab) - Permanent prohibitions
	if err := v.validateBloodRelationships(ctx, personA, personB); err != nil {
		return err
	}

	// 2. In-Law Relationships (Marriage-based) - Permanent prohibitions
	if err := v.validateInLawRelationships(ctx, personA, personB); err != nil {
		return err
	}

	// 3. Marriage state - Temporary prohibition
	if err := v.validateMarriageState(ctx, personA, personB); err != nil {
		return err
	}

	return nil
}

// validateBloodRelationships checks all blood-based prohibitions
func (v *MarriageValidator) validateBloodRelationships(ctx context.Context, personA, personB *domain.Member) error {
	// 1. Ancestor/Descendant check
	isAncestor, err := v.isAncestor(ctx, personA.MemberID, personB.MemberID)
	if err != nil {
		return err
	}
	if isAncestor {
		return domain.NewValidationError("error.spouse.ancestor_descendant", nil)
	}

	isAncestor, err = v.isAncestor(ctx, personB.MemberID, personA.MemberID)
	if err != nil {
		return err
	}
	if isAncestor {
		return domain.NewValidationError("error.spouse.ancestor_descendant", nil)
	}

	// 2. Siblings check (full or half siblings)
	if v.areSiblings(personA, personB) {
		return domain.NewValidationError("error.spouse.siblings", nil)
	}

	// 3. Aunt/Niece check (parent's sibling or sibling's child)
	isAuntNiece, err := v.isAuntNieceRelationship(ctx, personA, personB)
	if err != nil {
		return err
	}
	if isAuntNiece {
		return domain.NewValidationError("error.spouse.aunt_niece", nil)
	}

	return nil
}

// validateInLawRelationships checks all marriage-based prohibitions
func (v *MarriageValidator) validateInLawRelationships(ctx context.Context, personA, personB *domain.Member) error {
	// 1. Spouse's parents or children
	isInLaw, err := v.isDirectInLaw(ctx, personA.MemberID, personB.MemberID)
	if err != nil {
		return err
	}
	if isInLaw {
		return domain.NewValidationError("error.spouse.in_law", nil)
	}

	// 2. Parent's spouse or child's spouse
	isStepRelation, err := v.isStepRelationship(ctx, personA, personB)
	if err != nil {
		return err
	}
	if isStepRelation {
		return domain.NewValidationError("error.spouse.step_relation", nil)
	}

	return nil
}

// validateMarriageState checks if either person is currently married
func (v *MarriageValidator) validateMarriageState(ctx context.Context, personA, personB *domain.Member) error {
	isMarried, err := v.isCurrentlyMarried(ctx, personA.MemberID)
	if err != nil {
		return err
	}
	if isMarried {
		return domain.NewValidationError("error.spouse.already_married", map[string]string{"person": personA.Gender})
	}

	isMarried, err = v.isCurrentlyMarried(ctx, personB.MemberID)
	if err != nil {
		return err
	}
	if isMarried {
		return domain.NewValidationError("error.spouse.already_married", map[string]string{"person": personB.Gender})
	}

	return nil
}

// areSiblings checks if two people share at least one parent
func (v *MarriageValidator) areSiblings(personA, personB *domain.Member) bool {
	if personA.FatherID != nil && personB.FatherID != nil && *personA.FatherID == *personB.FatherID {
		return true
	}
	if personA.MotherID != nil && personB.MotherID != nil && *personA.MotherID == *personB.MotherID {
		return true
	}
	return false
}

// isAuntNieceRelationship checks if one person is aunt/uncle or niece/nephew of the other
func (v *MarriageValidator) isAuntNieceRelationship(ctx context.Context, personA, personB *domain.Member) (bool, error) {
	// Check if A is aunt/uncle of B (A is sibling of B's parent)
	if personB.FatherID != nil {
		father, err := v.memberRepo.Get(ctx, *personB.FatherID)
		if err == nil && v.areSiblings(personA, father) {
			return true, nil
		}
	}
	if personB.MotherID != nil {
		mother, err := v.memberRepo.Get(ctx, *personB.MotherID)
		if err == nil && v.areSiblings(personA, mother) {
			return true, nil
		}
	}

	// Check if B is aunt/uncle of A (B is sibling of A's parent)
	if personA.FatherID != nil {
		father, err := v.memberRepo.Get(ctx, *personA.FatherID)
		if err == nil && v.areSiblings(personB, father) {
			return true, nil
		}
	}
	if personA.MotherID != nil {
		mother, err := v.memberRepo.Get(ctx, *personA.MotherID)
		if err == nil && v.areSiblings(personB, mother) {
			return true, nil
		}
	}

	return false, nil
}

// isDirectInLaw checks if one person is spouse's parent or spouse's child
func (v *MarriageValidator) isDirectInLaw(ctx context.Context, personAID, personBID int) (bool, error) {
	// Get all spouses of A
	spousesA, err := v.spouseRepo.GetByMemberID(ctx, personAID)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return false, err
	}

	for _, spouseInfo := range spousesA {
		spouse, err := v.memberRepo.Get(ctx, spouseInfo.MemberID)
		if err != nil {
			continue
		}

		// Check if B is parent of A's spouse
		if (spouse.FatherID != nil && *spouse.FatherID == personBID) ||
			(spouse.MotherID != nil && *spouse.MotherID == personBID) {
			return true, nil
		}

		// Check if B is child of A's spouse
		children, err := v.memberRepo.GetChildrenByParentID(ctx, spouse.MemberID)
		if err == nil {
			for _, child := range children {
				if child.MemberID == personBID {
					return true, nil
				}
			}
		}
	}

	// Get all spouses of B and check the reverse
	spousesB, err := v.spouseRepo.GetByMemberID(ctx, personBID)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return false, err
	}

	for _, spouseInfo := range spousesB {
		spouse, err := v.memberRepo.Get(ctx, spouseInfo.MemberID)
		if err != nil {
			continue
		}

		// Check if A is parent of B's spouse
		if (spouse.FatherID != nil && *spouse.FatherID == personAID) ||
			(spouse.MotherID != nil && *spouse.MotherID == personAID) {
			return true, nil
		}

		// Check if A is child of B's spouse
		children, err := v.memberRepo.GetChildrenByParentID(ctx, spouse.MemberID)
		if err == nil {
			for _, child := range children {
				if child.MemberID == personAID {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// isStepRelationship checks if one person is spouse of the other's parent or child
func (v *MarriageValidator) isStepRelationship(ctx context.Context, personA, personB *domain.Member) (bool, error) {
	// Check if A is spouse of B's parent (stepparent)
	if personB.FatherID != nil {
		spouses, err := v.spouseRepo.GetByMemberID(ctx, *personB.FatherID)
		if err == nil {
			for _, spouse := range spouses {
				if spouse.MemberID == personA.MemberID {
					return true, nil
				}
			}
		}
	}
	if personB.MotherID != nil {
		spouses, err := v.spouseRepo.GetByMemberID(ctx, *personB.MotherID)
		if err == nil {
			for _, spouse := range spouses {
				if spouse.MemberID == personA.MemberID {
					return true, nil
				}
			}
		}
	}

	// Check if A is spouse of B's child (child-in-law)
	children, err := v.memberRepo.GetChildrenByParentID(ctx, personB.MemberID)
	if err == nil {
		for _, child := range children {
			spouses, err := v.spouseRepo.GetByMemberID(ctx, child.MemberID)
			if err == nil {
				for _, spouse := range spouses {
					if spouse.MemberID == personA.MemberID {
						return true, nil
					}
				}
			}
		}
	}

	// Check reverse (B is spouse of A's parent or child)
	if personA.FatherID != nil {
		spouses, err := v.spouseRepo.GetByMemberID(ctx, *personA.FatherID)
		if err == nil {
			for _, spouse := range spouses {
				if spouse.MemberID == personB.MemberID {
					return true, nil
				}
			}
		}
	}
	if personA.MotherID != nil {
		spouses, err := v.spouseRepo.GetByMemberID(ctx, *personA.MotherID)
		if err == nil {
			for _, spouse := range spouses {
				if spouse.MemberID == personB.MemberID {
					return true, nil
				}
			}
		}
	}

	children, err = v.memberRepo.GetChildrenByParentID(ctx, personA.MemberID)
	if err == nil {
		for _, child := range children {
			spouses, err := v.spouseRepo.GetByMemberID(ctx, child.MemberID)
			if err == nil {
				for _, spouse := range spouses {
					if spouse.MemberID == personB.MemberID {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

// isCurrentlyMarried checks if a person has an active marriage
func (v *MarriageValidator) isCurrentlyMarried(ctx context.Context, memberID int) (bool, error) {
	spouses, err := v.spouseRepo.GetByMemberID(ctx, memberID)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return false, err
	}

	for _, spouse := range spouses {
		if spouse.DivorceDate == nil {
			return true, nil
		}
	}

	return false, nil
}

// isAncestor checks if potentialAncestor is an ancestor of descendant
func (v *MarriageValidator) isAncestor(ctx context.Context, potentialAncestorID, descendantID int) (bool, error) {
	visited := make(map[int]bool)
	return v.checkAncestorRecursive(ctx, potentialAncestorID, descendantID, visited, 0, 100)
}

// checkAncestorRecursive recursively traverses up the family tree
func (v *MarriageValidator) checkAncestorRecursive(ctx context.Context, potentialAncestorID, currentID int, visited map[int]bool, depth, maxDepth int) (bool, error) {
	if depth > maxDepth || visited[currentID] {
		return false, nil
	}
	visited[currentID] = true

	current, err := v.memberRepo.Get(ctx, currentID)
	if err != nil {
		return false, err
	}

	if current.FatherID != nil {
		if *current.FatherID == potentialAncestorID {
			return true, nil
		}
		if isAncestor, err := v.checkAncestorRecursive(ctx, potentialAncestorID, *current.FatherID, visited, depth+1, maxDepth); err != nil {
			return false, err
		} else if isAncestor {
			return true, nil
		}
	}

	if current.MotherID != nil {
		if *current.MotherID == potentialAncestorID {
			return true, nil
		}
		if isAncestor, err := v.checkAncestorRecursive(ctx, potentialAncestorID, *current.MotherID, visited, depth+1, maxDepth); err != nil {
			return false, err
		} else if isAncestor {
			return true, nil
		}
	}

	return false, nil
}
