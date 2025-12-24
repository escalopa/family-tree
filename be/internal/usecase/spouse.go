package usecase

import (
	"context"
	"encoding/json"
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
	if father, err := uc.memberRepo.Get(ctx, fatherID); err != nil {
		slog.Error("get father for history", "error", err, "father_id", fatherID)
	} else {
		fatherVersion = father.Version
	}

	motherVersion := 0
	if mother, err := uc.memberRepo.Get(ctx, motherID); err != nil {
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

	if err := uc.historyRepo.CreateBatch(ctx, histories...); err != nil {
		slog.Error("create batch history for spouse", "error", err, "father_id", fatherID, "mother_id", motherID, "change_type", changeType)
	}
}

func (uc *spouseUseCase) Create(ctx context.Context, spouse *domain.Spouse, userID int) error {
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return domain.NewValidationError("error.spouse.invalid_marriage_date", nil)
	}

	father, err := uc.memberRepo.Get(ctx, spouse.FatherID)
	if err != nil {
		return err
	}
	if father.Gender != "M" {
		return domain.NewValidationError("error.member.invalid_parent", map[string]string{"parent": "father"})
	}

	mother, err := uc.memberRepo.Get(ctx, spouse.MotherID)
	if err != nil {
		return err
	}
	if mother.Gender != "F" {
		return domain.NewValidationError("error.member.invalid_parent", map[string]string{"parent": "mother"})
	}

	if err := uc.spouseRepo.Create(ctx, spouse); err != nil {
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

	return uc.scoreRepo.Create(ctx, scores...)
}

func (uc *spouseUseCase) Update(ctx context.Context, spouse *domain.Spouse, userID int) error {
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return domain.NewValidationError("error.spouse.invalid_marriage_date", nil)
	}

	oldSpouse, err := uc.spouseRepo.Get(ctx, spouse.SpouseID)
	if err != nil {
		return err
	}

	spouse.FatherID = oldSpouse.FatherID
	spouse.MotherID = oldSpouse.MotherID

	if err := uc.spouseRepo.Update(ctx, spouse); err != nil {
		return err
	}

	oldValues, _ := json.Marshal(oldSpouse)
	newValues, _ := json.Marshal(spouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeUpdateSpouse, oldValues, newValues, userID)

	return nil
}

func (uc *spouseUseCase) Delete(ctx context.Context, spouseID, userID int) error {
	oldSpouse, err := uc.spouseRepo.Get(ctx, spouseID)
	if err != nil {
		return err
	}

	hasChildren, err := uc.memberRepo.HasChildrenWithParents(ctx, oldSpouse.FatherID, oldSpouse.MotherID)
	if err != nil {
		return err
	}
	if hasChildren {
		return domain.NewConflictError("error.spouse.has_children", nil)
	}

	if err := uc.spouseRepo.Delete(ctx, spouseID); err != nil {
		return err
	}

	oldValues, _ := json.Marshal(oldSpouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeRemoveSpouse, oldValues, nil, userID)

	return nil
}
