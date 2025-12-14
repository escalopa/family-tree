package usecase

import (
	"context"
	"encoding/json"
	"fmt"

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

	// Ensure member1_id < member2_id for consistency
	if spouse.Member1ID > spouse.Member2ID {
		spouse.Member1ID, spouse.Member2ID = spouse.Member2ID, spouse.Member1ID
	}

	// Create spouse relationship
	if err := uc.spouseRepo.Create(ctx, spouse); err != nil {
		return fmt.Errorf("failed to create spouse relationship: %w", err)
	}

	// Record history for both members
	spouseJSON, _ := json.Marshal(spouse)
	history1 := &domain.History{
		MemberID:      spouse.Member1ID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeAddSpouse,
		OldValues:     nil,
		NewValues:     spouseJSON,
		MemberVersion: 0,
	}
	_ = uc.historyRepo.Create(ctx, history1)

	history2 := &domain.History{
		MemberID:      spouse.Member2ID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeAddSpouse,
		OldValues:     nil,
		NewValues:     spouseJSON,
		MemberVersion: 0,
	}
	_ = uc.historyRepo.Create(ctx, history2)

	// Record scores for both members
	score1 := &domain.Score{
		UserID:        userID,
		MemberID:      spouse.Member1ID,
		FieldName:     "spouse",
		Points:        domain.PointsSpouse,
		MemberVersion: 0,
	}
	_ = uc.scoreRepo.Create(ctx, score1)

	score2 := &domain.Score{
		UserID:        userID,
		MemberID:      spouse.Member2ID,
		FieldName:     "spouse",
		Points:        domain.PointsSpouse,
		MemberVersion: 0,
	}
	_ = uc.scoreRepo.Create(ctx, score2)

	return nil
}

func (uc *spouseUseCase) UpdateSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error {
	// Validate dates
	if !validator.ValidateDateOrder(spouse.MarriageDate, spouse.DivorceDate) {
		return fmt.Errorf("divorce date must be after marriage date")
	}

	// Ensure member1_id < member2_id for consistency
	if spouse.Member1ID > spouse.Member2ID {
		spouse.Member1ID, spouse.Member2ID = spouse.Member2ID, spouse.Member1ID
	}

	// Get old values
	oldSpouse, err := uc.spouseRepo.Get(ctx, spouse.Member1ID, spouse.Member2ID)
	if err != nil {
		return fmt.Errorf("failed to get spouse relationship: %w", err)
	}

	// Update spouse relationship
	if err := uc.spouseRepo.Update(ctx, spouse); err != nil {
		return fmt.Errorf("failed to update spouse relationship: %w", err)
	}

	// Record history
	oldJSON, _ := json.Marshal(oldSpouse)
	newJSON, _ := json.Marshal(spouse)
	history1 := &domain.History{
		MemberID:      spouse.Member1ID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeUpdateSpouse,
		OldValues:     oldJSON,
		NewValues:     newJSON,
		MemberVersion: 0,
	}
	_ = uc.historyRepo.Create(ctx, history1)

	history2 := &domain.History{
		MemberID:      spouse.Member2ID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeUpdateSpouse,
		OldValues:     oldJSON,
		NewValues:     newJSON,
		MemberVersion: 0,
	}
	_ = uc.historyRepo.Create(ctx, history2)

	return nil
}

func (uc *spouseUseCase) RemoveSpouse(ctx context.Context, member1ID, member2ID, userID int) error {
	// Get old values for history
	oldSpouse, err := uc.spouseRepo.Get(ctx, member1ID, member2ID)
	if err != nil {
		return fmt.Errorf("failed to get spouse relationship: %w", err)
	}

	// Delete spouse relationship
	if err := uc.spouseRepo.Delete(ctx, member1ID, member2ID); err != nil {
		return fmt.Errorf("failed to delete spouse relationship: %w", err)
	}

	// Record history
	oldJSON, _ := json.Marshal(oldSpouse)
	history1 := &domain.History{
		MemberID:      oldSpouse.Member1ID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeRemoveSpouse,
		OldValues:     oldJSON,
		NewValues:     nil,
		MemberVersion: 0,
	}
	_ = uc.historyRepo.Create(ctx, history1)

	history2 := &domain.History{
		MemberID:      oldSpouse.Member2ID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeRemoveSpouse,
		OldValues:     oldJSON,
		NewValues:     nil,
		MemberVersion: 0,
	}
	_ = uc.historyRepo.Create(ctx, history2)

	return nil
}
