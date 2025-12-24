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

func (uc *memberUseCase) Create(ctx context.Context, member *domain.Member, userID int) error {
	if len(member.Names) == 0 {
		return domain.NewValidationError("at least one name must be provided")
	}

	if err := uc.validateParentRelationships(ctx, member.MemberID, member.FatherID, member.MotherID); err != nil {
		return err
	}

	if err := uc.memberRepo.Create(ctx, member); err != nil {
		return fmt.Errorf("create member: %w", err)
	}

	if err := uc.ensureParentSpouseRelationship(ctx, member.FatherID, member.MotherID, userID); err != nil {
		return fmt.Errorf("ensure parent spouse relationship: %w", err)
	}

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

	if err := uc.calculateAndRecordScores(ctx, member, userID); err != nil {
		return fmt.Errorf("record scores: %w", err)
	}

	return nil
}

func (uc *memberUseCase) Update(ctx context.Context, member *domain.Member, expectedVersion, userID int) error {
	oldMember, err := uc.memberRepo.Get(ctx, member.MemberID)
	if err != nil {
		return fmt.Errorf("get old member: %w", err)
	}

	if len(member.Names) == 0 {
		return domain.NewValidationError("at least one name must be provided")
	}

	if oldMember.Gender != member.Gender {
		spouses, err := uc.spouseRepo.GetByMemberID(ctx, member.MemberID)
		if err != nil {
			return fmt.Errorf("check for spouses: %w", err)
		}
		if len(spouses) > 0 {
			return domain.NewValidationError("cannot change gender: member has spouse relationships")
		}
	}

	if err := uc.validateParentRelationships(ctx, member.MemberID, member.FatherID, member.MotherID); err != nil {
		return err
	}

	if err := uc.memberRepo.Update(ctx, member, expectedVersion); err != nil {
		return fmt.Errorf("update member: %w", err)
	}

	if err := uc.ensureParentSpouseRelationship(ctx, member.FatherID, member.MotherID, userID); err != nil {
		return fmt.Errorf("ensure parent spouse relationship: %w", err)
	}

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

	if err := uc.updateScores(ctx, oldMember, member, userID); err != nil {
		return fmt.Errorf("update scores: %w", err)
	}

	return nil
}

func (uc *memberUseCase) Delete(ctx context.Context, memberID, userID int) error {
	oldMember, err := uc.memberRepo.Get(ctx, memberID)
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
		slog.Error("record delete history", "error", err, "member_id", memberID)
	}

	pictureURL, err := uc.memberRepo.Delete(ctx, memberID)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	if pictureURL != nil && *pictureURL != "" {
		if err := uc.s3Client.DeleteImage(ctx, *pictureURL); err != nil {
			slog.Error("delete member picture from storage", "error", err, "member_id", memberID, "picture", *pictureURL)
		}
	}

	return nil
}

func (uc *memberUseCase) Get(ctx context.Context, memberID int) (*domain.Member, error) {
	return uc.memberRepo.Get(ctx, memberID)
}

func (uc *memberUseCase) ListChildren(ctx context.Context, parentID int) ([]*domain.Member, error) {
	return uc.memberRepo.GetChildrenByParentID(ctx, parentID)
}

func (uc *memberUseCase) ListSiblings(ctx context.Context, memberID int) ([]*domain.Member, error) {
	return uc.memberRepo.GetSiblingsByMemberID(ctx, memberID)
}

func (uc *memberUseCase) List(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error) {
	return uc.memberRepo.List(ctx, filter, cursor, limit)
}

func (uc *memberUseCase) ListHistory(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	return uc.historyRepo.GetByMemberID(ctx, memberID, cursor, limit)
}

func (uc *memberUseCase) UploadPicture(ctx context.Context, memberID int, data []byte, filename string, userID int) (string, error) {
	oldMember, err := uc.memberRepo.Get(ctx, memberID)
	if err != nil {
		return "", fmt.Errorf("get member: %w", err)
	}

	newPictureURL, err := uc.s3Client.UploadImage(ctx, data, filename)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	if err := uc.memberRepo.UpdatePicture(ctx, memberID, newPictureURL); err != nil {
		if deleteErr := uc.s3Client.DeleteImage(ctx, newPictureURL); deleteErr != nil {
			slog.Error("rollback S3 upload after DB error",
				"error", deleteErr,
				"member_id", memberID,
				"picture_url", newPictureURL,
				"original_error", err)
		}
		return "", fmt.Errorf("update member picture: %w", err)
	}

	if oldMember.Picture != nil && *oldMember.Picture != "" {
		if err := uc.s3Client.DeleteImage(ctx, *oldMember.Picture); err != nil {
			slog.Warn("delete old picture from S3",
				"error", err,
				"member_id", memberID,
				"old_picture", *oldMember.Picture)
		}
	}

	updatedMember, err := uc.memberRepo.Get(ctx, memberID)
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
		slog.Error("create history for picture upload", "error", err, "member_id", memberID)
	}

	if oldMember.Picture == nil || *oldMember.Picture == "" {
		scores := []domain.Score{
			{
				UserID:        userID,
				MemberID:      memberID,
				FieldName:     "picture",
				Points:        domain.PointsPicture,
				MemberVersion: oldMember.Version + 1,
			},
		}
		_ = uc.scoreRepo.Create(ctx, scores...)
	}

	return newPictureURL, nil
}

func (uc *memberUseCase) DeletePicture(ctx context.Context, memberID int, userID int) error {
	member, err := uc.memberRepo.Get(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}

	oldPictureURL := member.Picture

	if err := uc.memberRepo.DeletePicture(ctx, memberID); err != nil {
		return fmt.Errorf("delete picture from database: %w", err)
	}

	if oldPictureURL != nil && *oldPictureURL != "" {
		if err := uc.s3Client.DeleteImage(ctx, *oldPictureURL); err != nil {
			slog.Warn("delete picture from S3 after DB update",
				"error", err,
				"member_id", memberID,
				"picture_url", *oldPictureURL)
		}
	}

	updatedMember, err := uc.memberRepo.Get(ctx, memberID)
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
		slog.Error("create history for picture deletion", "error", err, "member_id", memberID)
	}

	return nil
}

func (uc *memberUseCase) GetPicture(ctx context.Context, memberID int) ([]byte, string, error) {
	member, err := uc.memberRepo.Get(ctx, memberID)
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
	if fatherID != nil && *fatherID == memberID {
		slog.Warn("memberUseCase.validateParentRelationships: member cannot be their own father", "member_id", memberID, "father_id", *fatherID)
		return domain.NewValidationError("member cannot be their own father")
	}
	if motherID != nil && *motherID == memberID {
		slog.Warn("memberUseCase.validateParentRelationships: member cannot be their own mother", "member_id", memberID, "mother_id", *motherID)
		return domain.NewValidationError("member cannot be their own mother")
	}

	if fatherID != nil {
		if err := uc.checkCircularRelationship(ctx, memberID, *fatherID); err != nil {
			return fmt.Errorf("invalid father relationship: %w", err)
		}
	}

	if motherID != nil {
		if err := uc.checkCircularRelationship(ctx, memberID, *motherID); err != nil {
			return fmt.Errorf("invalid mother relationship: %w", err)
		}
	}

	return nil
}

// ensureParentSpouseRelationship creates a spouse relationship between father and mother if both exist
func (uc *memberUseCase) ensureParentSpouseRelationship(ctx context.Context, fatherID, motherID *int, userID int) error {
	if fatherID == nil || motherID == nil {
		return nil
	}

	_, err := uc.spouseRepo.GetByParents(ctx, *fatherID, *motherID)
	if err == nil {
		return nil
	}

	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrCodeNotFound {
		return fmt.Errorf("check spouse relationship: %w", err)
	}

	spouse := &domain.Spouse{
		FatherID:     *fatherID,
		MotherID:     *motherID,
		MarriageDate: nil,
		DivorceDate:  nil,
	}

	if err := uc.spouseRepo.Create(ctx, spouse); err != nil {
		return fmt.Errorf("create spouse relationship: %w", err)
	}

	newValuesJSON, _ := json.Marshal(spouse)
	uc.recordSpouseHistory(ctx, *fatherID, *motherID, domain.ChangeTypeAddSpouse, nil, newValuesJSON, userID)

	return nil
}

func (uc *memberUseCase) recordSpouseHistory(
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

func (uc *memberUseCase) checkCircularRelationship(ctx context.Context, memberID, parentID int) error {
	parent, err := uc.memberRepo.Get(ctx, parentID)
	if err != nil {
		return fmt.Errorf("get parent: %w", err)
	}

	if parent.FatherID != nil && *parent.FatherID == memberID {
		slog.Warn("memberUseCase.checkCircularRelationship: circular relationship detected - parent's father cannot be their child", "member_id", memberID, "parent_id", parentID)
		return domain.NewValidationError("circular relationship detected: parent's father cannot be their child")
	}
	if parent.MotherID != nil && *parent.MotherID == memberID {
		slog.Warn("memberUseCase.checkCircularRelationship: circular relationship detected - parent's mother cannot be their child", "member_id", memberID, "parent_id", parentID)
		return domain.NewValidationError("circular relationship detected: parent's mother cannot be their child")
	}

	visited := make(map[int]bool)
	return uc.checkAncestors(ctx, parentID, memberID, visited, 0)
}

func (uc *memberUseCase) checkAncestors(ctx context.Context, currentID, targetID int, visited map[int]bool, depth int) error {
	if depth > 50 {
		slog.Warn("memberUseCase.checkAncestors: family tree depth limit exceeded", "current_id", currentID, "target_id", targetID, "depth", depth)
		return domain.NewValidationError("family tree depth limit exceeded")
	}
	if visited[currentID] {
		return nil
	}
	visited[currentID] = true

	current, err := uc.memberRepo.Get(ctx, currentID)
	if err != nil {
		return nil // If member not found, skip
	}

	if current.FatherID != nil {
		if *current.FatherID == targetID {
			slog.Warn("memberUseCase.checkAncestors: circular relationship detected in ancestry (father)", "current_id", currentID, "target_id", targetID, "father_id", *current.FatherID, "depth", depth)
			return domain.NewValidationError("circular relationship detected in ancestry")
		}
		if err := uc.checkAncestors(ctx, *current.FatherID, targetID, visited, depth+1); err != nil {
			return err
		}
	}

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

	for langCode := range member.Names {
		scores = append(scores, domain.Score{
			UserID:        userID,
			MemberID:      member.MemberID,
			FieldName:     "name_" + langCode,
			Points:        domain.PointsName,
			MemberVersion: member.Version,
		})
	}
	scores = append(scores, domain.Score{UserID: userID, MemberID: member.MemberID, FieldName: "gender", Points: domain.PointsGender, MemberVersion: member.Version})

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

	return uc.scoreRepo.Create(ctx, scores...)

}

func (uc *memberUseCase) updateScores(ctx context.Context, oldMember, newMember *domain.Member, userID int) error {
	scores := []domain.Score{}

	for langCode, newName := range newMember.Names {
		oldName, exists := oldMember.Names[langCode]
		if newName != "" && (!exists || oldName == "") {
			scores = append(scores, domain.Score{
				UserID:        userID,
				MemberID:      newMember.MemberID,
				FieldName:     "name_" + langCode,
				Points:        domain.PointsName,
				MemberVersion: newMember.Version,
			})
		}
	}

	if (oldMember.Picture == nil || *oldMember.Picture == "") && newMember.Picture != nil && *newMember.Picture != "" {
		scores = append(scores, domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "picture", Points: domain.PointsPicture, MemberVersion: newMember.Version})
	}

	if oldMember.DateOfBirth == nil && newMember.DateOfBirth != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "date_of_birth", Points: domain.PointsDateOfBirth, MemberVersion: newMember.Version})
	}

	if oldMember.DateOfDeath == nil && newMember.DateOfDeath != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "date_of_death", Points: domain.PointsDateOfDeath, MemberVersion: newMember.Version})
	}

	if oldMember.FatherID == nil && newMember.FatherID != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "father_id", Points: domain.PointsFather, MemberVersion: newMember.Version})
	}

	if oldMember.MotherID == nil && newMember.MotherID != nil {
		scores = append(scores, domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "mother_id", Points: domain.PointsMother, MemberVersion: newMember.Version})
	}

	if len(oldMember.Nicknames) == 0 && len(newMember.Nicknames) > 0 {
		scores = append(scores, domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "nicknames", Points: domain.PointsNicknames, MemberVersion: newMember.Version})
	}

	if (oldMember.Profession == nil || *oldMember.Profession == "") && newMember.Profession != nil && *newMember.Profession != "" {
		scores = append(scores, domain.Score{UserID: userID, MemberID: newMember.MemberID, FieldName: "profession", Points: domain.PointsProfession, MemberVersion: newMember.Version})
	}

	return uc.scoreRepo.Create(ctx, scores...)
}

func (uc *memberUseCase) Compute(ctx context.Context, member *domain.Member, userRole int) *domain.MemberWithComputed {
	computed := &domain.MemberWithComputed{
		Member: *member,
	}

	computed.FullNames = uc.buildFullNamesForAllLanguages(ctx, member)

	if member.DateOfBirth != nil {
		endDate := time.Now()
		if member.DateOfDeath != nil {
			endDate = *member.DateOfDeath
		}
		age := int(endDate.Sub(*member.DateOfBirth).Hours() / 24 / 365)
		computed.Age = &age
	}

	spouses, _ := uc.spouseRepo.GetByMemberID(ctx, member.MemberID)
	computed.IsMarried = len(spouses) > 0
	computed.Spouses = spouses

	if member.Gender == "F" {
		if userRole < domain.RoleAdmin {
			computed.Picture = nil
		}
		if userRole < domain.RoleSuperAdmin {
			if computed.DateOfBirth != nil {
				hiddenDate := time.Date(1, computed.DateOfBirth.Month(), computed.DateOfBirth.Day(), 0, 0, 0, 0, time.UTC)
				computed.DateOfBirth = &hiddenDate
			}
			if computed.DateOfDeath != nil {
				hiddenDate := time.Date(1, computed.DateOfDeath.Month(), computed.DateOfDeath.Day(), 0, 0, 0, 0, time.UTC)
				computed.DateOfDeath = &hiddenDate
			}
			computed.Age = nil
		}
	}

	return computed
}

// buildFullNamesForAllLanguages builds full names in all available languages by tracing through the father's lineage
// Returns: map[languageCode]fullName
// Example: {"ar": "محمد أحمد علي", "en": "Muhammad Ahmad Ali", "ru": "Мухаммад Ахмад Али"}
func (uc *memberUseCase) buildFullNamesForAllLanguages(ctx context.Context, member *domain.Member) map[string]string {
	namesPerLanguage := make(map[string][]string)
	for langCode, name := range member.Names {
		namesPerLanguage[langCode] = []string{name}
	}

	if member.FatherID != nil {
		currentFatherID := member.FatherID
		maxDepth := 10 // Prevent infinite loops
		depth := 0

		for currentFatherID != nil && depth < maxDepth {
			father, err := uc.memberRepo.Get(ctx, *currentFatherID)
			if err != nil {
				break
			}

			for langCode, name := range father.Names {
				namesPerLanguage[langCode] = append(namesPerLanguage[langCode], name)
			}

			currentFatherID = father.FatherID
			depth++
		}
	}

	fullNames := make(map[string]string)
	for langCode, names := range namesPerLanguage {
		fullNames[langCode] = strings.Join(names, " ")
	}

	return fullNames
}
