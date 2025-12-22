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

func (uc *spouseUseCase) AddSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error {
	// Validate dates
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return fmt.Errorf("divorce date must be after marriage date")
	}

	// Validate that father is male and mother is female
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

	// Create spouse relationship
	if err := uc.spouseRepo.Create(ctx, spouse); err != nil {
		return fmt.Errorf("create spouse relationship: %w", err)
	}

	// Record history for both members
	spouseJSON, _ := json.Marshal(spouse)
	history1 := &domain.History{
		MemberID:      spouse.FatherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeAddSpouse,
		OldValues:     nil,
		NewValues:     spouseJSON,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history1); err != nil {
		slog.Error("failed to create history for father", "error", err, "father_id", spouse.FatherID)
	}

	history2 := &domain.History{
		MemberID:      spouse.MotherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeAddSpouse,
		OldValues:     nil,
		NewValues:     spouseJSON,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history2); err != nil {
		slog.Error("failed to create history for mother", "error", err, "mother_id", spouse.MotherID)
	}

	// Record scores for both members
	score1 := &domain.Score{
		UserID:        userID,
		MemberID:      spouse.FatherID,
		FieldName:     "spouse",
		Points:        domain.PointsSpouse,
		MemberVersion: 0,
	}
	_ = uc.scoreRepo.Create(ctx, score1)

	score2 := &domain.Score{
		UserID:        userID,
		MemberID:      spouse.MotherID,
		FieldName:     "spouse",
		Points:        domain.PointsSpouse,
		MemberVersion: 0,
	}
	_ = uc.scoreRepo.Create(ctx, score2)

	return nil
}

func (uc *spouseUseCase) UpdateSpouseByID(ctx context.Context, spouse *domain.Spouse, userID int) error {
	// Validate dates
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return fmt.Errorf("divorce date must be after marriage date")
	}

	// Get old values
	oldSpouse, err := uc.spouseRepo.GetByID(ctx, spouse.SpouseID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	// Update spouse relationship
	if err := uc.spouseRepo.UpdateByID(ctx, spouse); err != nil {
		return fmt.Errorf("update spouse relationship: %w", err)
	}

	// Record history
	oldJSON, _ := json.Marshal(oldSpouse)
	newJSON, _ := json.Marshal(spouse)
	history1 := &domain.History{
		MemberID:      oldSpouse.FatherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeUpdateSpouse,
		OldValues:     oldJSON,
		NewValues:     newJSON,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history1); err != nil {
		slog.Error("failed to create update history for father", "error", err, "father_id", oldSpouse.FatherID)
	}

	history2 := &domain.History{
		MemberID:      oldSpouse.MotherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeUpdateSpouse,
		OldValues:     oldJSON,
		NewValues:     newJSON,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history2); err != nil {
		slog.Error("failed to create update history for mother", "error", err, "mother_id", oldSpouse.MotherID)
	}

	return nil
}

func (uc *spouseUseCase) UpdateSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error {
	// Validate dates
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return fmt.Errorf("divorce date must be after marriage date")
	}

	// Validate that father is male and mother is female
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

	// Get old values
	oldSpouse, err := uc.spouseRepo.Get(ctx, spouse.FatherID, spouse.MotherID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	// Update spouse relationship
	if err := uc.spouseRepo.Update(ctx, spouse); err != nil {
		return fmt.Errorf("update spouse relationship: %w", err)
	}

	// Record history
	oldJSON, _ := json.Marshal(oldSpouse)
	newJSON, _ := json.Marshal(spouse)
	history1 := &domain.History{
		MemberID:      spouse.FatherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeUpdateSpouse,
		OldValues:     oldJSON,
		NewValues:     newJSON,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history1); err != nil {
		slog.Error("failed to create update history for father", "error", err, "father_id", spouse.FatherID)
	}

	history2 := &domain.History{
		MemberID:      spouse.MotherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeUpdateSpouse,
		OldValues:     oldJSON,
		NewValues:     newJSON,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history2); err != nil {
		slog.Error("failed to create update history for mother", "error", err, "mother_id", spouse.MotherID)
	}

	return nil
}

func (uc *spouseUseCase) RemoveSpouse(ctx context.Context, fatherID, motherID, userID int) error {
	// Get old values for history
	oldSpouse, err := uc.spouseRepo.Get(ctx, fatherID, motherID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	// Check if there are children with this spouse relationship
	hasChildren, err := uc.memberRepo.HasChildrenWithParents(ctx, fatherID, motherID)
	if err != nil {
		return fmt.Errorf("check for children: %w", err)
	}
	if hasChildren {
		return domain.NewValidationError("cannot delete spouse relationship: there are children with both parents")
	}

	// Delete spouse relationship
	if err := uc.spouseRepo.Delete(ctx, fatherID, motherID); err != nil {
		return fmt.Errorf("delete spouse relationship: %w", err)
	}

	// Record history
	oldJSON, _ := json.Marshal(oldSpouse)
	history1 := &domain.History{
		MemberID:      oldSpouse.FatherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeRemoveSpouse,
		OldValues:     oldJSON,
		NewValues:     nil,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history1); err != nil {
		slog.Error("failed to create remove history for father", "error", err, "father_id", oldSpouse.FatherID)
	}

	history2 := &domain.History{
		MemberID:      oldSpouse.MotherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeRemoveSpouse,
		OldValues:     oldJSON,
		NewValues:     nil,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history2); err != nil {
		slog.Error("failed to create remove history for mother", "error", err, "mother_id", oldSpouse.MotherID)
	}

	return nil
}

func (uc *spouseUseCase) RemoveSpouseByID(ctx context.Context, spouseID, userID int) error {
	// Get old values for history
	oldSpouse, err := uc.spouseRepo.GetByID(ctx, spouseID)
	if err != nil {
		return fmt.Errorf("get spouse relationship: %w", err)
	}

	// Check if there are children with this spouse relationship
	hasChildren, err := uc.memberRepo.HasChildrenWithParents(ctx, oldSpouse.FatherID, oldSpouse.MotherID)
	if err != nil {
		return fmt.Errorf("check for children: %w", err)
	}
	if hasChildren {
		return domain.NewValidationError("cannot delete spouse relationship: there are children with both parents")
	}

	// Delete spouse relationship
	if err := uc.spouseRepo.DeleteByID(ctx, spouseID); err != nil {
		return fmt.Errorf("delete spouse relationship: %w", err)
	}

	// Record history
	oldJSON, _ := json.Marshal(oldSpouse)
	history1 := &domain.History{
		MemberID:      oldSpouse.FatherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeRemoveSpouse,
		OldValues:     oldJSON,
		NewValues:     nil,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history1); err != nil {
		slog.Error("failed to create remove history for father", "error", err, "father_id", oldSpouse.FatherID)
	}

	history2 := &domain.History{
		MemberID:      oldSpouse.MotherID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeRemoveSpouse,
		OldValues:     oldJSON,
		NewValues:     nil,
		MemberVersion: 0,
	}
	if err := uc.historyRepo.Create(ctx, history2); err != nil {
		slog.Error("failed to create remove history for mother", "error", err, "mother_id", oldSpouse.MotherID)
	}

	return nil
}
