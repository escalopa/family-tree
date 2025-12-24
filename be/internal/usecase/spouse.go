package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/pkg/validator"
)

type spouseUseCase struct {
	spouseRepo  SpouseRepository
	memberRepo  MemberRepository
	historyRepo HistoryRepository
	scoreRepo   ScoreRepository
}

func NewSpouseUseCase(
	spouseRepo SpouseRepository,
	memberRepo MemberRepository,
	historyRepo HistoryRepository,
	scoreRepo ScoreRepository,
) *spouseUseCase {
	return &spouseUseCase{
		spouseRepo:  spouseRepo,
		memberRepo:  memberRepo,
		historyRepo: historyRepo,
		scoreRepo:   scoreRepo,
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
	if father, err := uc.memberRepo.GetByID(ctx, fatherID); err != nil {
		slog.Error("get father for history", "error", err, "father_id", fatherID)
	} else {
		fatherVersion = father.Version
	}

	motherVersion := 0
	if mother, err := uc.memberRepo.GetByID(ctx, motherID); err != nil {
		slog.Error("get mother for history", "error", err, "mother_id", motherID)
	} else {
		motherVersion = mother.Version
	}

	historyFather := &domain.History{
		MemberID:      fatherID,
		UserID:        userID,
		ChangeType:    changeType,
		OldValues:     oldValues,
		NewValues:     newValues,
		MemberVersion: fatherVersion,
	}
	if err := uc.historyRepo.Create(ctx, historyFather); err != nil {
		slog.Error("create history for father", "error", err, "father_id", fatherID, "change_type", changeType)
	}

	historyMother := &domain.History{
		MemberID:      motherID,
		UserID:        userID,
		ChangeType:    changeType,
		OldValues:     oldValues,
		NewValues:     newValues,
		MemberVersion: motherVersion,
	}
	if err := uc.historyRepo.Create(ctx, historyMother); err != nil {
		slog.Error("create history for mother", "error", err, "mother_id", motherID, "change_type", changeType)
	}
}

func (uc *spouseUseCase) AddSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error {
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return fmt.Errorf("divorce date must be after marriage date")
	}

	father, err := uc.memberRepo.GetByID(ctx, spouse.FatherID)
	if err != nil {
		return fmt.Errorf("get father: %w", err)
	}
	if father.Gender != "M" {
		return domain.NewValidationError("father must be male")
	}

	mother, err := uc.memberRepo.GetByID(ctx, spouse.MotherID)
	if err != nil {
		return fmt.Errorf("get mother: %w", err)
	}
	if mother.Gender != "F" {
		return domain.NewValidationError("mother must be female")
	}

	if err := uc.spouseRepo.Create(ctx, spouse); err != nil {
		return fmt.Errorf("create spouse relationship: %w", err)
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

	return uc.scoreRepo.Create(ctx, scores...)
}

func (uc *spouseUseCase) UpdateSpouseByID(ctx context.Context, spouse *domain.Spouse, userID int) error {
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return fmt.Errorf("divorce date must be after marriage date")
	}

	oldSpouse, err := uc.spouseRepo.GetByID(ctx, spouse.SpouseID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	if err := uc.spouseRepo.UpdateByID(ctx, spouse); err != nil {
		return fmt.Errorf("update spouse relationship: %w", err)
	}

	oldValues, _ := json.Marshal(oldSpouse)
	newValues, _ := json.Marshal(spouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeUpdateSpouse, oldValues, newValues, userID)

	return nil
}

func (uc *spouseUseCase) UpdateSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error {
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return fmt.Errorf("divorce date must be after marriage date")
	}

	father, err := uc.memberRepo.GetByID(ctx, spouse.FatherID)
	if err != nil {
		return fmt.Errorf("get father: %w", err)
	}
	if father.Gender != "M" {
		return domain.NewValidationError("father must be male")
	}

	mother, err := uc.memberRepo.GetByID(ctx, spouse.MotherID)
	if err != nil {
		return fmt.Errorf("get mother: %w", err)
	}
	if mother.Gender != "F" {
		return domain.NewValidationError("mother must be female")
	}

	oldSpouse, err := uc.spouseRepo.Get(ctx, spouse.FatherID, spouse.MotherID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	if err := uc.spouseRepo.Update(ctx, spouse); err != nil {
		return fmt.Errorf("update spouse relationship: %w", err)
	}

	oldValues, _ := json.Marshal(oldSpouse)
	newValues, _ := json.Marshal(spouse)
	uc.recordSpouseHistory(ctx, spouse.FatherID, spouse.MotherID, domain.ChangeTypeUpdateSpouse, oldValues, newValues, userID)

	return nil
}

func (uc *spouseUseCase) RemoveSpouse(ctx context.Context, fatherID, motherID, userID int) error {
	oldSpouse, err := uc.spouseRepo.Get(ctx, fatherID, motherID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	hasChildren, err := uc.memberRepo.HasChildrenWithParents(ctx, fatherID, motherID)
	if err != nil {
		return fmt.Errorf("check for children: %w", err)
	}
	if hasChildren {
		return domain.NewValidationError("cannot delete spouse relationship: there are children with both parents")
	}

	if err := uc.spouseRepo.Delete(ctx, fatherID, motherID); err != nil {
		return fmt.Errorf("delete spouse relationship: %w", err)
	}

	oldValues, _ := json.Marshal(oldSpouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeRemoveSpouse, oldValues, nil, userID)

	return nil
}

func (uc *spouseUseCase) RemoveSpouseByID(ctx context.Context, spouseID, userID int) error {
	oldSpouse, err := uc.spouseRepo.GetByID(ctx, spouseID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	hasChildren, err := uc.memberRepo.HasChildrenWithParents(ctx, oldSpouse.FatherID, oldSpouse.MotherID)
	if err != nil {
		return fmt.Errorf("check for children: %w", err)
	}
	if hasChildren {
		return domain.NewValidationError("cannot delete spouse relationship: there are children with both parents")
	}

	if err := uc.spouseRepo.DeleteByID(ctx, spouseID); err != nil {
		return fmt.Errorf("delete spouse relationship: %w", err)
	}

	oldValues, _ := json.Marshal(oldSpouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeRemoveSpouse, oldValues, nil, userID)

	return nil
}
