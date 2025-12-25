package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/pkg/validator"
)

type (
	spouseUseCaseRepo struct {
		spouse  SpouseRepository
		member  MemberRepository
		history HistoryRepository
		score   ScoreRepository
	}

	spouseUseCaseValidator struct {
		marriage MarriageValidator
	}

	spouseUseCase struct {
		repo      spouseUseCaseRepo
		validator spouseUseCaseValidator
	}
)

func NewSpouseUseCase(
	spouseRepo SpouseRepository,
	memberRepo MemberRepository,
	historyRepo HistoryRepository,
	scoreRepo ScoreRepository,
	marriageValidator MarriageValidator,
) *spouseUseCase {
	return &spouseUseCase{
		repo: spouseUseCaseRepo{
			spouse:  spouseRepo,
			member:  memberRepo,
			history: historyRepo,
			score:   scoreRepo,
		},
		validator: spouseUseCaseValidator{
			marriage: marriageValidator,
		},
	}
}

func (uc *spouseUseCase) recordSpouseHistory(
	ctx context.Context,
	fatherID, motherID int,
	changeType string,
	oldValues, newValues json.RawMessage,
	userID int,
) {
	fatherVersion := 0
	if father, err := uc.repo.member.Get(ctx, fatherID); err != nil {
		slog.Error("get father for history", "error", err, "father_id", fatherID)
	} else {
		fatherVersion = father.Version
	}

	motherVersion := 0
	if mother, err := uc.repo.member.Get(ctx, motherID); err != nil {
		slog.Error("get mother for history", "error", err, "mother_id", motherID)
	} else {
		motherVersion = mother.Version
	}

	histories := []*domain.History{
		{
			MemberID:      fatherID,
			UserID:        userID,
			ChangeType:    changeType,
			OldValues:     oldValues,
			NewValues:     newValues,
			MemberVersion: fatherVersion,
		},
		{
			MemberID:      motherID,
			UserID:        userID,
			ChangeType:    changeType,
			OldValues:     oldValues,
			NewValues:     newValues,
			MemberVersion: motherVersion,
		},
	}

	if err := uc.repo.history.CreateBatch(ctx, histories...); err != nil {
		slog.Error("create batch history for spouse", "error", err, "father_id", fatherID, "mother_id", motherID, "change_type", changeType)
	}
}

func (uc *spouseUseCase) Create(ctx context.Context, spouse *domain.Spouse, userID int) error {
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return domain.NewValidationError("error.spouse.invalid_marriage_date", nil)
	}

	existingSpouse, err := uc.repo.spouse.GetByParents(ctx, spouse.FatherID, spouse.MotherID)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return err
	}
	if existingSpouse != nil {
		return domain.NewConflictError("error.spouse.already_exists", nil)
	}

	father, err := uc.repo.member.Get(ctx, spouse.FatherID)
	if err != nil {
		return err
	}
	if father.Gender != "M" {
		return domain.NewValidationError("error.member.invalid_parent", map[string]string{"parent": "father"})
	}

	mother, err := uc.repo.member.Get(ctx, spouse.MotherID)
	if err != nil {
		return err
	}
	if mother.Gender != "F" {
		return domain.NewValidationError("error.member.invalid_parent", map[string]string{"parent": "mother"})
	}

	// Validate Islamic marriage prohibitions
	if err := uc.validator.marriage.Create(ctx, spouse.FatherID, spouse.MotherID); err != nil {
		return err
	}

	if err := uc.validateMarriageDateAgainstBirth(father, mother, spouse.MarriageDate); err != nil {
		return err
	}

	if err := uc.validateMarriageDateAgainstChildren(ctx, spouse.FatherID, spouse.MotherID, spouse.MarriageDate); err != nil {
		return err
	}

	if err := uc.repo.spouse.Create(ctx, spouse); err != nil {
		return err
	}

	newValues, _ := json.Marshal(spouse)
	uc.recordSpouseHistory(ctx, spouse.FatherID, spouse.MotherID, domain.ChangeTypeAddSpouse, nil, newValues, userID)

	scores := []domain.Score{
		{
			UserID:        userID,
			MemberID:      spouse.FatherID,
			FieldName:     "spouse",
			Points:        domain.PointsSpouse,
			MemberVersion: father.Version,
		},
		{
			UserID:        userID,
			MemberID:      spouse.MotherID,
			FieldName:     "spouse",
			Points:        domain.PointsSpouse,
			MemberVersion: mother.Version,
		},
	}

	return uc.repo.score.Create(ctx, scores...)
}

func (uc *spouseUseCase) Update(ctx context.Context, spouse *domain.Spouse, userID int) error {
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return domain.NewValidationError("error.spouse.invalid_marriage_date", nil)
	}

	oldSpouse, err := uc.repo.spouse.Get(ctx, spouse.SpouseID)
	if err != nil {
		return err
	}

	spouse.FatherID = oldSpouse.FatherID
	spouse.MotherID = oldSpouse.MotherID

	father, err := uc.repo.member.Get(ctx, spouse.FatherID)
	if err != nil {
		return err
	}

	mother, err := uc.repo.member.Get(ctx, spouse.MotherID)
	if err != nil {
		return err
	}

	if err := uc.validateMarriageDateAgainstBirth(father, mother, spouse.MarriageDate); err != nil {
		return err
	}

	if err := uc.validateMarriageDateAgainstChildren(ctx, spouse.FatherID, spouse.MotherID, spouse.MarriageDate); err != nil {
		return err
	}

	if err := uc.repo.spouse.Update(ctx, spouse); err != nil {
		return err
	}

	oldValues, _ := json.Marshal(oldSpouse)
	newValues, _ := json.Marshal(spouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeUpdateSpouse, oldValues, newValues, userID)

	return nil
}

func (uc *spouseUseCase) Delete(ctx context.Context, spouseID, userID int) error {
	oldSpouse, err := uc.repo.spouse.Get(ctx, spouseID)
	if err != nil {
		return err
	}

	hasChildren, err := uc.repo.member.HasChildrenWithParents(ctx, oldSpouse.FatherID, oldSpouse.MotherID)
	if err != nil {
		return err
	}
	if hasChildren {
		return domain.NewConflictError("error.spouse.has_children", nil)
	}

	if err := uc.repo.spouse.Delete(ctx, spouseID); err != nil {
		return err
	}

	oldValues, _ := json.Marshal(oldSpouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeRemoveSpouse, oldValues, nil, userID)

	return nil
}

// validateParentAges ensures both parents exist (already validated by caller)
func (uc *spouseUseCase) validateParentAges(ctx context.Context, father, mother *domain.Member) error {
	// Both members already validated to exist and have correct genders
	return nil
}

// validateMarriageDateAgainstBirth ensures marriage date is after both parents' birth dates
func (uc *spouseUseCase) validateMarriageDateAgainstBirth(father, mother *domain.Member, marriageDate *time.Time) error {
	if marriageDate == nil {
		return nil
	}

	if father.DateOfBirth != nil && marriageDate.Before(*father.DateOfBirth) {
		return domain.NewValidationError("error.spouse.marriage_before_father_birth", nil)
	}

	if mother.DateOfBirth != nil && marriageDate.Before(*mother.DateOfBirth) {
		return domain.NewValidationError("error.spouse.marriage_before_mother_birth", nil)
	}

	return nil
}

func (uc *spouseUseCase) validateMarriageDateAgainstChildren(ctx context.Context, fatherID, motherID int, marriageDate *time.Time) error {
	if marriageDate == nil {
		return nil
	}

	children, err := uc.repo.member.GetChildrenByParents(ctx, fatherID, motherID)
	if err != nil {
		return err
	}

	for _, child := range children {
		if child.DateOfBirth != nil && marriageDate.After(*child.DateOfBirth) {
			return domain.NewValidationError("error.spouse.marriage_after_child_birth", nil)
		}
	}

	return nil
}
