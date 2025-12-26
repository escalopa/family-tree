package usecase

import (
	"context"
	"encoding/json"
	"log/slog"

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
		tx        TransactionManager
	}
)

func NewSpouseUseCase(
	spouseRepo SpouseRepository,
	memberRepo MemberRepository,
	historyRepo HistoryRepository,
	scoreRepo ScoreRepository,
	txManager TransactionManager,
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
		tx: txManager,
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
		return domain.NewValidationError("error.spouse.invalid_marriage_date")
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
		return domain.NewValidationError("error.member.invalid_father")
	}

	mother, err := uc.repo.member.Get(ctx, spouse.MotherID)
	if err != nil {
		return err
	}
	if mother.Gender != "F" {
		return domain.NewValidationError("error.member.invalid_mother")
	}

	if err := uc.validator.marriage.Create(ctx, spouse.FatherID, spouse.MotherID); err != nil {
		return err
	}

	if err := uc.validator.marriage.MarriageDate(ctx, spouse.FatherID, spouse.MotherID, spouse.MarriageDate); err != nil {
		return err
	}

	return uc.tx.Do(ctx, func(txCtx context.Context) error {
		return uc.createTx(txCtx, spouse, father, mother, userID)
	})
}

func (uc *spouseUseCase) createTx(ctx context.Context, spouse *domain.Spouse, father, mother *domain.Member, userID int) error {
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
		return domain.NewValidationError("error.spouse.invalid_marriage_date")
	}

	oldSpouse, err := uc.repo.spouse.Get(ctx, spouse.SpouseID)
	if err != nil {
		return err
	}

	spouse.FatherID = oldSpouse.FatherID
	spouse.MotherID = oldSpouse.MotherID

	if err := uc.validator.marriage.MarriageDate(ctx, spouse.FatherID, spouse.MotherID, spouse.MarriageDate); err != nil {
		return err
	}

	return uc.tx.Do(ctx, func(txCtx context.Context) error {
		return uc.updateTx(txCtx, spouse, oldSpouse, userID)
	})
}

func (uc *spouseUseCase) updateTx(ctx context.Context, spouse, oldSpouse *domain.Spouse, userID int) error {
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

	return uc.tx.Do(ctx, func(txCtx context.Context) error {
		return uc.deleteTx(txCtx, spouseID, oldSpouse, userID)
	})
}

func (uc *spouseUseCase) deleteTx(ctx context.Context, spouseID int, oldSpouse *domain.Spouse, userID int) error {
	if err := uc.repo.spouse.Delete(ctx, spouseID); err != nil {
		return err
	}

	oldValues, _ := json.Marshal(oldSpouse)
	uc.recordSpouseHistory(ctx, oldSpouse.FatherID, oldSpouse.MotherID, domain.ChangeTypeRemoveSpouse, oldValues, nil, userID)

	return nil
}
