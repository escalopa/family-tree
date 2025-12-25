package validator

import (
	"context"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/usecase"
)

type BirthDateValidator struct {
	memberRepo usecase.MemberRepository
	spouseRepo usecase.SpouseRepository
}

// NewBirthDateValidator creates a new birth date validator
func NewBirthDateValidator(memberRepo usecase.MemberRepository, spouseRepo usecase.SpouseRepository) *BirthDateValidator {
	return &BirthDateValidator{
		memberRepo: memberRepo,
		spouseRepo: spouseRepo,
	}
}

// Update validates birth date changes for a member
func (v *BirthDateValidator) Update(ctx context.Context, memberID int, newBirthDate *time.Time) error {
	if newBirthDate == nil {
		return nil
	}

	// 1. Member birth date must be before all their marriage dates
	if err := v.validateBirthBeforeMarriages(ctx, memberID, newBirthDate); err != nil {
		return err
	}

	// 2. Member birth date must be after parent birth dates
	member, err := v.memberRepo.Get(ctx, memberID)
	if err != nil {
		return err
	}

	if err := v.validateBirthAfterParents(ctx, member, newBirthDate); err != nil {
		return err
	}

	// 3. Member birth date must be before children birth dates
	if err := v.validateBirthBeforeChildren(ctx, memberID, newBirthDate); err != nil {
		return err
	}

	return nil
}

// Create validates birth date when creating/updating a child
func (v *BirthDateValidator) Create(ctx context.Context, childBirth *time.Time, fatherID, motherID *int) error {
	if childBirth == nil {
		return nil
	}

	// 1. Child birth must be after parent birth dates
	if fatherID != nil {
		father, err := v.memberRepo.Get(ctx, *fatherID)
		if err != nil {
			return err
		}
		if father.DateOfBirth != nil && !father.DateOfBirth.Before(*childBirth) {
			return domain.NewValidationError("error.member.parent_born_after_child", map[string]string{"parent": father.Gender})
		}
	}

	if motherID != nil {
		mother, err := v.memberRepo.Get(ctx, *motherID)
		if err != nil {
			return err
		}
		if mother.DateOfBirth != nil && !mother.DateOfBirth.Before(*childBirth) {
			return domain.NewValidationError("error.member.parent_born_after_child", map[string]string{"parent": mother.Gender})
		}
	}

	// 2. Child birth must be after parents' marriage date (if both parents set)
	if fatherID != nil && motherID != nil {
		if err := v.validateBirthAfterParentsMarriage(ctx, childBirth, *fatherID, *motherID); err != nil {
			return err
		}
	}

	return nil
}

// validateBirthBeforeMarriages ensures member's birth date is before all their marriage dates
func (v *BirthDateValidator) validateBirthBeforeMarriages(ctx context.Context, memberID int, birthDate *time.Time) error {
	spouses, err := v.spouseRepo.GetByMemberID(ctx, memberID)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return err
	}

	for _, spouse := range spouses {
		if spouse.MarriageDate != nil && !birthDate.Before(*spouse.MarriageDate) {
			return domain.NewValidationError("error.member.birth_after_marriage", nil)
		}
	}

	return nil
}

// validateBirthAfterParents ensures member's birth date is after parent birth dates
func (v *BirthDateValidator) validateBirthAfterParents(ctx context.Context, member *domain.Member, birthDate *time.Time) error {
	if member.FatherID != nil {
		father, err := v.memberRepo.Get(ctx, *member.FatherID)
		if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
			return err
		}
		if err == nil && father.DateOfBirth != nil && !father.DateOfBirth.Before(*birthDate) {
			return domain.NewValidationError("error.member.parent_born_after_child", map[string]string{"parent": father.Gender})
		}
	}

	if member.MotherID != nil {
		mother, err := v.memberRepo.Get(ctx, *member.MotherID)
		if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
			return err
		}
		if err == nil && mother.DateOfBirth != nil && !mother.DateOfBirth.Before(*birthDate) {
			return domain.NewValidationError("error.member.parent_born_after_child", map[string]string{"parent": mother.Gender})
		}
	}

	return nil
}

// validateBirthBeforeChildren ensures member's birth date is before children birth dates
func (v *BirthDateValidator) validateBirthBeforeChildren(ctx context.Context, memberID int, birthDate *time.Time) error {
	children, err := v.memberRepo.GetChildrenByParentID(ctx, memberID)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return err
	}

	for _, child := range children {
		if child.DateOfBirth != nil && !birthDate.Before(*child.DateOfBirth) {
			return domain.NewValidationError("error.member.parent_born_after_child", nil)
		}
	}

	return nil
}

// validateBirthAfterParentsMarriage ensures child's birth is after parents' marriage date
func (v *BirthDateValidator) validateBirthAfterParentsMarriage(ctx context.Context, childBirth *time.Time, fatherID, motherID int) error {
	spouse, err := v.spouseRepo.GetByParents(ctx, fatherID, motherID)
	if err != nil {
		if domain.IsDomainError(err, domain.ErrCodeNotFound) {
			return nil
		}
		return err
	}

	if spouse.MarriageDate != nil && childBirth.Before(*spouse.MarriageDate) {
		return domain.NewValidationError("error.member.birth_before_parents_marriage", nil)
	}

	return nil
}
