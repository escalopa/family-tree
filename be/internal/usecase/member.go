package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"
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
	// Validate parent relationships
	if err := uc.validateParentRelationships(ctx, member.MemberID, member.FatherID, member.MotherID); err != nil {
		return err
	}

	// Create member
	if err := uc.memberRepo.Create(ctx, member); err != nil {
		return fmt.Errorf("create member: %w", err)
	}

	// Ensure spouse relationship exists between parents
	if err := uc.ensureParentSpouseRelationship(ctx, member.FatherID, member.MotherID); err != nil {
		return fmt.Errorf("ensure parent spouse relationship: %w", err)
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

	// Validate parent relationships
	if err := uc.validateParentRelationships(ctx, member.MemberID, member.FatherID, member.MotherID); err != nil {
		return err
	}

	// Update member
	if err := uc.memberRepo.Update(ctx, member, expectedVersion); err != nil {
		return fmt.Errorf("update member: %w", err)
	}

	// Ensure spouse relationship exists between parents
	if err := uc.ensureParentSpouseRelationship(ctx, member.FatherID, member.MotherID); err != nil {
		return fmt.Errorf("ensure parent spouse relationship: %w", err)
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
	oldMember, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}

	children, err := uc.memberRepo.GetChildrenByParentID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("check children: %w", err)
	}
	if len(children) > 0 {
		return domain.NewValidationError("cannot delete member: this member has children")
	}

	if err := uc.memberRepo.SoftDelete(ctx, memberID); err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

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
		slog.Error("failed to record delete history", "error", err, "member_id", memberID)
	}

	return nil
}

func (uc *memberUseCase) GetMemberByID(ctx context.Context, memberID int) (*domain.Member, error) {
	return uc.memberRepo.GetByID(ctx, memberID)
}

func (uc *memberUseCase) GetChildrenByParentID(ctx context.Context, parentID int) ([]*domain.Member, error) {
	return uc.memberRepo.GetChildrenByParentID(ctx, parentID)
}

func (uc *memberUseCase) GetSiblingsByMemberID(ctx context.Context, memberID int) ([]*domain.Member, error) {
	return uc.memberRepo.GetSiblingsByMemberID(ctx, memberID)
}

func (uc *memberUseCase) SearchMembers(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error) {

	return uc.memberRepo.Search(ctx, filter, cursor, limit)
}

func (uc *memberUseCase) GetMemberHistory(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.historyRepo.GetByMemberID(ctx, memberID, cursor, limit)
}

func (uc *memberUseCase) UploadPicture(ctx context.Context, memberID int, data []byte, filename string, userID int) (string, error) {
	oldMember, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return "", fmt.Errorf("get member: %w", err)
	}

	newPictureURL, err := uc.s3Client.UploadImage(ctx, data, filename)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	if err := uc.memberRepo.UpdatePicture(ctx, memberID, newPictureURL); err != nil {
		if deleteErr := uc.s3Client.DeleteImage(ctx, newPictureURL); deleteErr != nil {
			slog.Error("failed to rollback S3 upload after DB error",
				"error", deleteErr,
				"member_id", memberID,
				"picture_url", newPictureURL,
				"original_error", err)
		}
		return "", fmt.Errorf("update member picture: %w", err)
	}

	if oldMember.Picture != nil && *oldMember.Picture != "" {
		if err := uc.s3Client.DeleteImage(ctx, *oldMember.Picture); err != nil {
			slog.Warn("failed to delete old picture from S3",
				"error", err,
				"member_id", memberID,
				"old_picture", *oldMember.Picture)
		}
	}

	updatedMember, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return "", fmt.Errorf("get updated member: %w", err)
	}

	oldValuesJSON, _ := json.Marshal(map[string]any{"picture": oldMember.Picture})
	newValuesJSON, _ := json.Marshal(map[string]any{"picture": updatedMember.Picture})
	history := &domain.History{
		MemberID:      memberID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeAddPicture,
		OldValues:     oldValuesJSON,
		NewValues:     newValuesJSON,
		MemberVersion: updatedMember.Version,
	}
	if err := uc.historyRepo.Create(ctx, history); err != nil {
		slog.Error("failed to create history for picture upload", "error", err, "member_id", memberID)
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

	return newPictureURL, nil
}

func (uc *memberUseCase) DeletePicture(ctx context.Context, memberID int, userID int) error {
	member, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}

	oldPictureURL := member.Picture

	if err := uc.memberRepo.DeletePicture(ctx, memberID); err != nil {
		return fmt.Errorf("delete picture from database: %w", err)
	}

	if oldPictureURL != nil && *oldPictureURL != "" {
		if err := uc.s3Client.DeleteImage(ctx, *oldPictureURL); err != nil {
			slog.Warn("failed to delete picture from S3 after DB update",
				"error", err,
				"member_id", memberID,
				"picture_url", *oldPictureURL)
		}
	}

	updatedMember, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get updated member: %w", err)
	}

	oldValuesJSON, _ := json.Marshal(map[string]any{"picture": oldPictureURL})
	newValuesJSON, _ := json.Marshal(map[string]any{"picture": nil})
	history := &domain.History{
		MemberID:      memberID,
		UserID:        userID,
		ChangeType:    domain.ChangeTypeDeletePicture,
		OldValues:     oldValuesJSON,
		NewValues:     newValuesJSON,
		MemberVersion: updatedMember.Version,
	}
	if err := uc.historyRepo.Create(ctx, history); err != nil {
		slog.Error("failed to create history for picture deletion", "error", err, "member_id", memberID)
	}

	return nil
}

func (uc *memberUseCase) GetPicture(ctx context.Context, memberID int) ([]byte, string, error) {
	member, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, "", fmt.Errorf("get member: %w", err)
	}

	if member.Picture == nil || *member.Picture == "" {
		slog.Warn("memberUseCase.GetPicture: picture not found", "member_id", memberID)
		return nil, "", domain.NewNotFoundError("picture")
	}

	imageData, err := uc.s3Client.GetImage(ctx, *member.Picture)
	if err != nil {
		return nil, "", fmt.Errorf("get image: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(*member.Picture))

	return imageData, contentType, nil
}

// validateParentRelationships checks for circular relationships in the family tree
func (uc *memberUseCase) validateParentRelationships(ctx context.Context, memberID int, fatherID, motherID *int) error {
	// Check if member is trying to be their own parent
	if fatherID != nil && *fatherID == memberID {
		slog.Warn("memberUseCase.validateParentRelationships: member cannot be their own father", "member_id", memberID, "father_id", *fatherID)
		return domain.NewValidationError("member cannot be their own father")
	}
	if motherID != nil && *motherID == memberID {
		slog.Warn("memberUseCase.validateParentRelationships: member cannot be their own mother", "member_id", memberID, "mother_id", *motherID)
		return domain.NewValidationError("member cannot be their own mother")
	}

	// Check if father is trying to be a child of this member
	if fatherID != nil {
		if err := uc.checkCircularRelationship(ctx, memberID, *fatherID); err != nil {
			return fmt.Errorf("invalid father relationship: %w", err)
		}
	}

	// Check if mother is trying to be a child of this member
	if motherID != nil {
		if err := uc.checkCircularRelationship(ctx, memberID, *motherID); err != nil {
			return fmt.Errorf("invalid mother relationship: %w", err)
		}
	}

	return nil
}

// ensureParentSpouseRelationship creates a spouse relationship between father and mother if both exist
func (uc *memberUseCase) ensureParentSpouseRelationship(ctx context.Context, fatherID, motherID *int) error {
	// Only proceed if both parents are set
	if fatherID == nil || motherID == nil {
		return nil
	}

	// Check if spouse relationship already exists
	_, err := uc.spouseRepo.Get(ctx, *fatherID, *motherID)
	if err == nil {
		// Relationship already exists
		return nil
	}

	// If error is not "not found", return it
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrCodeNotFound {
		return fmt.Errorf("check spouse relationship: %w", err)
	}

	// Create spouse relationship (not found, so we create it)
	spouse := &domain.Spouse{
		FatherID:     *fatherID,
		MotherID:     *motherID,
		MarriageDate: nil, // Marriage date is unknown when auto-creating
		DivorceDate:  nil,
	}

	if err := uc.spouseRepo.Create(ctx, spouse); err != nil {
		return fmt.Errorf("create spouse relationship: %w", err)
	}

	return nil
}

// checkCircularRelationship checks if parentID is a descendant of memberID
func (uc *memberUseCase) checkCircularRelationship(ctx context.Context, memberID, parentID int) error {
	// Get the potential parent
	parent, err := uc.memberRepo.GetByID(ctx, parentID)
	if err != nil {
		return fmt.Errorf("get parent: %w", err)
	}

	// Check if the parent's father or mother is the member (direct circular)
	if parent.FatherID != nil && *parent.FatherID == memberID {
		slog.Warn("memberUseCase.checkCircularRelationship: circular relationship detected - parent's father cannot be their child", "member_id", memberID, "parent_id", parentID)
		return domain.NewValidationError("circular relationship detected: parent's father cannot be their child")
	}
	if parent.MotherID != nil && *parent.MotherID == memberID {
		slog.Warn("memberUseCase.checkCircularRelationship: circular relationship detected - parent's mother cannot be their child", "member_id", memberID, "parent_id", parentID)
		return domain.NewValidationError("circular relationship detected: parent's mother cannot be their child")
	}

	// Recursively check ancestors (prevent deep circular relationships)
	visited := make(map[int]bool)
	return uc.checkAncestors(ctx, parentID, memberID, visited, 0)
}

// checkAncestors recursively checks if targetID appears in the ancestry of currentID
func (uc *memberUseCase) checkAncestors(ctx context.Context, currentID, targetID int, visited map[int]bool, depth int) error {
	// Prevent infinite loops and limit depth
	if depth > 50 {
		slog.Warn("memberUseCase.checkAncestors: family tree depth limit exceeded", "current_id", currentID, "target_id", targetID, "depth", depth)
		return domain.NewValidationError("family tree depth limit exceeded")
	}
	if visited[currentID] {
		return nil
	}
	visited[currentID] = true

	// Get current member
	current, err := uc.memberRepo.GetByID(ctx, currentID)
	if err != nil {
		return nil // If member not found, skip
	}

	// Check father's lineage
	if current.FatherID != nil {
		if *current.FatherID == targetID {
			slog.Warn("memberUseCase.checkAncestors: circular relationship detected in ancestry (father)", "current_id", currentID, "target_id", targetID, "father_id", *current.FatherID, "depth", depth)
			return domain.NewValidationError("circular relationship detected in ancestry")
		}
		if err := uc.checkAncestors(ctx, *current.FatherID, targetID, visited, depth+1); err != nil {
			return err
		}
	}

	// Check mother's lineage
	if current.MotherID != nil {
		if *current.MotherID == targetID {
			slog.Warn("memberUseCase.checkAncestors: circular relationship detected in ancestry (mother)", "current_id", currentID, "target_id", targetID, "mother_id", *current.MotherID, "depth", depth)
			return domain.NewValidationError("circular relationship detected in ancestry")
		}
		if err := uc.checkAncestors(ctx, *current.MotherID, targetID, visited, depth+1); err != nil {
			return err
		}
	}

	return nil
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

	// Compute full names by building lineage through father's line
	computed.ArabicFullName, computed.EnglishFullName = uc.buildFullNames(ctx, member)

	// Compute age
	if member.DateOfBirth != nil {
		endDate := time.Now()
		if member.DateOfDeath != nil {
			endDate = *member.DateOfDeath
		}
		age := int(endDate.Sub(*member.DateOfBirth).Hours() / 24 / 365)
		computed.Age = &age
	}

	// Fetch spouse information
	spouses, _ := uc.spouseRepo.GetSpousesWithMemberInfo(ctx, member.MemberID)
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

// buildFullNames builds both Arabic and English full names by tracing through the father's lineage in a single pass
// Returns: (arabicFullName, englishFullName)
// Example: (محمد أحمد علي, Muhammad Ahmad Ali)
func (uc *memberUseCase) buildFullNames(ctx context.Context, member *domain.Member) (string, string) {
	arabicNames := []string{member.ArabicName}
	englishNames := []string{member.EnglishName}

	// If member has a father, trace through father's lineage
	if member.FatherID != nil {
		currentFatherID := member.FatherID
		maxDepth := 10 // Prevent infinite loops
		depth := 0

		for currentFatherID != nil && depth < maxDepth {
			father, err := uc.memberRepo.GetByID(ctx, *currentFatherID)
			if err != nil {
				break
			}

			arabicNames = append(arabicNames, father.ArabicName)
			englishNames = append(englishNames, father.EnglishName)

			currentFatherID = father.FatherID
			depth++
		}
	}

	return strings.Join(arabicNames, " "), strings.Join(englishNames, " ")
}
