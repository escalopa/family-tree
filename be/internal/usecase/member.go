package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
)

type memberUseCase struct {
	memberRepo  MemberRepository
	spouseRepo  SpouseRepository
	historyRepo HistoryRepository
	scoreRepo   ScoreRepository
	s3Client    S3Client
}

func NewMemberUseCase(
	memberRepo MemberRepository,
	spouseRepo SpouseRepository,
	historyRepo HistoryRepository,
	scoreRepo ScoreRepository,
	s3Client S3Client,
) *memberUseCase {
	return &memberUseCase{
		memberRepo:  memberRepo,
		spouseRepo:  spouseRepo,
		historyRepo: historyRepo,
		scoreRepo:   scoreRepo,
		s3Client:    s3Client,
	}
}

func (uc *memberUseCase) CreateMember(ctx context.Context, member *domain.Member, userID int) error {
	// Create member
	if err := uc.memberRepo.Create(ctx, member); err != nil {
		return fmt.Errorf("create member: %w", err)
	}

	// Record history
	newValuesJSON, _ := json.Marshal(member)
	history := &domain.History{
		MemberID:      member.MemberID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeInsert,
		OldValues:     nil,
		NewValues:     newValuesJSON,
		MemberVersion: member.Version,
	}
	if err := uc.historyRepo.Create(ctx, history); err != nil {
		return fmt.Errorf("record history: %w", err)
	}

	// Calculate and record scores
	if err := uc.calculateAndRecordScores(ctx, member, userID); err != nil {
		return fmt.Errorf("record scores: %w", err)
	}

	return nil
}

func (uc *memberUseCase) UpdateMember(ctx context.Context, member *domain.Member, expectedVersion, userID int) error {
	// Get old member
	oldMember, err := uc.memberRepo.GetByID(ctx, member.MemberID)
	if err != nil {
		return fmt.Errorf("get old member: %w", err)
	}

	// Update member
	if err := uc.memberRepo.Update(ctx, member, expectedVersion); err != nil {
		return fmt.Errorf("update member: %w", err)
	}

	// Record history
	oldValuesJSON, _ := json.Marshal(oldMember)
	newValuesJSON, _ := json.Marshal(member)
	history := &domain.History{
		MemberID:      member.MemberID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeUpdate,
		OldValues:     oldValuesJSON,
		NewValues:     newValuesJSON,
		MemberVersion: member.Version,
	}
	if err := uc.historyRepo.Create(ctx, history); err != nil {
		return fmt.Errorf("record history: %w", err)
	}

	// Update scores
	if err := uc.updateScores(ctx, oldMember, member, userID); err != nil {
		return fmt.Errorf("update scores: %w", err)
	}

	return nil
}

func (uc *memberUseCase) DeleteMember(ctx context.Context, memberID, userID int) error {
	// Get old member
	oldMember, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}

	// Soft delete
	if err := uc.memberRepo.SoftDelete(ctx, memberID); err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	// Record history
	oldValuesJSON, _ := json.Marshal(oldMember)
	history := &domain.History{
		MemberID:      memberID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeDelete,
		OldValues:     oldValuesJSON,
		NewValues:     nil,
		MemberVersion: oldMember.Version + 1,
	}
	if err := uc.historyRepo.Create(ctx, history); err != nil {
		return fmt.Errorf("record history: %w", err)
	}

	return nil
}

func (uc *memberUseCase) GetMemberByID(ctx context.Context, memberID int) (*domain.Member, error) {
	return uc.memberRepo.GetByID(ctx, memberID)
}

func (uc *memberUseCase) SearchMembers(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.memberRepo.Search(ctx, filter, cursor, limit)
}

func (uc *memberUseCase) GetMemberHistory(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.historyRepo.GetByMemberID(ctx, memberID, cursor, limit)
}

func (uc *memberUseCase) UploadPicture(ctx context.Context, memberID int, data []byte, filename string, userID int) (string, error) {
	// Upload to S3
	url, err := uc.s3Client.UploadImage(ctx, data, filename)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	// Get old member to delete old picture if exists
	oldMember, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return "", fmt.Errorf("get member: %w", err)
	}

	// Update member picture
	if err := uc.memberRepo.UpdatePicture(ctx, memberID, url); err != nil {
		// Rollback: delete uploaded image
		_ = uc.s3Client.DeleteImage(ctx, url)
		return "", fmt.Errorf("update member picture: %w", err)
	}

	// Delete old picture from S3 if exists
	if oldMember.Picture != nil && *oldMember.Picture != "" {
		_ = uc.s3Client.DeleteImage(ctx, *oldMember.Picture)
	}

	// Record score if first time adding picture
	if oldMember.Picture == nil || *oldMember.Picture == "" {
		score := &domain.Score{
			UserID:        userID,
			MemberID:      memberID,
			FieldName:     "picture",
			Points:        domain.PointsPicture,
			MemberVersion: oldMember.Version + 1,
		}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	return url, nil
}

func (uc *memberUseCase) DeletePicture(ctx context.Context, memberID int) error {
	member, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}

	if member.Picture != nil && *member.Picture != "" {
		// Delete from S3
		if err := uc.s3Client.DeleteImage(ctx, *member.Picture); err != nil {
			return fmt.Errorf("delete image: %w", err)
		}
	}

	// Update member
	return uc.memberRepo.DeletePicture(ctx, memberID)
}

func (uc *memberUseCase) calculateAndRecordScores(ctx context.Context, member *domain.Member, userID int) error {
	scores := []domain.Score{}

	// Mandatory fields (always present)
	scores = append(scores,
		domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "arabic_name", Points: domain.PointsArabicName, MemberVersion: member.Version},
		domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "english_name", Points: domain.PointsEnglishName, MemberVersion: member.Version},
		domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "gender", Points: domain.PointsGender, MemberVersion: member.Version},
	)

	// Optional fields
	if member.Picture != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "picture", Points: domain.PointsPicture, MemberVersion: member.Version})
	}
	if member.DateOfBirth != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "date_of_birth", Points: domain.PointsDateOfBirth, MemberVersion: member.Version})
	}
	if member.DateOfDeath != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "date_of_death", Points: domain.PointsDateOfDeath, MemberVersion: member.Version})
	}
	if member.FatherID != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "father_id", Points: domain.PointsFather, MemberVersion: member.Version})
	}
	if member.MotherID != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "mother_id", Points: domain.PointsMother, MemberVersion: member.Version})
	}
	if len(member.Nicknames) > 0 {
		scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "nicknames", Points: domain.PointsNicknames, MemberVersion: member.Version})
	}
	if member.Profession != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "profession", Points: domain.PointsProfession, MemberVersion: member.Version})
	}

	// Record all scores
	for _, score := range scores {
		if err := uc.scoreRepo.Create(ctx, &score); err != nil {
			return err
		}
	}

	return nil
}

func (uc *memberUseCase) updateScores(ctx context.Context, oldMember, newMember *domain.Member, userID int) error {
	// Check which fields changed and award points accordingly
	// For simplicity, we'll check optional fields that might be newly filled

	// Picture
	if (oldMember.Picture == nil || *oldMember.Picture == "") && newMember.Picture != nil && *newMember.Picture != "" {
		score := &domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "picture", Points: domain.PointsPicture, MemberVersion: newMember.Version}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	// Date of birth
	if oldMember.DateOfBirth == nil && newMember.DateOfBirth != nil {
		score := &domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "date_of_birth", Points: domain.PointsDateOfBirth, MemberVersion: newMember.Version}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	// Date of death
	if oldMember.DateOfDeath == nil && newMember.DateOfDeath != nil {
		score := &domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "date_of_death", Points: domain.PointsDateOfDeath, MemberVersion: newMember.Version}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	// Father
	if oldMember.FatherID == nil && newMember.FatherID != nil {
		score := &domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "father_id", Points: domain.PointsFather, MemberVersion: newMember.Version}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	// Mother
	if oldMember.MotherID == nil && newMember.MotherID != nil {
		score := &domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "mother_id", Points: domain.PointsMother, MemberVersion: newMember.Version}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	// Nicknames
	if len(oldMember.Nicknames) == 0 && len(newMember.Nicknames) > 0 {
		score := &domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "nicknames", Points: domain.PointsNicknames, MemberVersion: newMember.Version}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	// Profession
	if (oldMember.Profession == nil || *oldMember.Profession == "") && newMember.Profession != nil && *newMember.Profession != "" {
		score := &domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "profession", Points: domain.PointsProfession, MemberVersion: newMember.Version}
		_ = uc.scoreRepo.Create(ctx, score)
	}

	return nil
}

func (uc *memberUseCase) ComputeMemberWithExtras(ctx context.Context, member *domain.Member, userRole int) *domain.MemberWithComputed {
	computed := &domain.MemberWithComputed{
		Member: *member,
	}

	// Compute full names (simplified - in production would build full lineage)
	computed.ArabicFullName = member.ArabicName
	computed.EnglishFullName = member.EnglishName

	// Compute age
	if member.DateOfBirth != nil {
		endDate := time.Now()
		if member.DateOfDeath != nil {
			endDate = *member.DateOfDeath
		}
		age := int(endDate.Sub(*member.DateOfBirth).Hours() / 24 / 365)
		computed.Age = &age
	}

	// Check if married
	spouses, _ := uc.spouseRepo.GetSpousesByMemberID(ctx, member.MemberID)
	computed.IsMarried = len(spouses) > 0
	computed.Spouses = spouses

	// Apply privacy rules
	if member.Gender == "F" {
		if userRole < domain.RoleAdmin {
			// Hide picture for non-admins
			computed.Picture = nil
		}
		if userRole < domain.RoleSuperAdmin {
			// Hide birth/death dates for non-super-admins
			computed.DateOfBirth = nil
			computed.DateOfDeath = nil
			computed.Age = nil
		}
	}

	return computed
}
