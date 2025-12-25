package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
)

type (
	memberUseCaseValidator struct {
		marriage     MarriageValidator
		birthDate    BirthDateValidator
		relationship RelationshipValidator
	}

	memberUseCaseRepo struct {
		member  MemberRepository
		spouse  SpouseRepository
		history HistoryRepository
		score   ScoreRepository
	}

	memberUseCase struct {
		repo      memberUseCaseRepo
		validator memberUseCaseValidator
		s3Client  S3Client
	}
)

func NewMemberUseCase(
	memberRepo MemberRepository,
	spouseRepo SpouseRepository,
	historyRepo HistoryRepository,
	scoreRepo ScoreRepository,
	s3Client S3Client,
	marriageValidator MarriageValidator,
	birthDateValidator BirthDateValidator,
	relationshipValidator RelationshipValidator,
) *memberUseCase {
	return &memberUseCase{
		repo:      memberUseCaseRepo{memberRepo, spouseRepo, historyRepo, scoreRepo},
		validator: memberUseCaseValidator{marriageValidator, birthDateValidator, relationshipValidator},
		s3Client:  s3Client,
	}
}

func (uc *memberUseCase) Create(ctx context.Context, member *domain.Member, userID int) error {
	if len(member.Names) == 0 {
		return domain.
			NewValidationError("error.validation.names_required").
			WithParams(map[string]string{"language": "all", "code": "all"})
	}

	if err := uc.validator.relationship.CheckParents(ctx, member.MemberID, member.FatherID, member.MotherID); err != nil {
		return err
	}

	if err := uc.validator.birthDate.Create(ctx, member.DateOfBirth, member.FatherID, member.MotherID); err != nil {
		return err
	}

	if err := uc.repo.member.Create(ctx, member); err != nil {
		return err
	}

	if err := uc.ensureParentSpouseRelationship(ctx, member.FatherID, member.MotherID, userID); err != nil {
		return err
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
	if err := uc.repo.history.Create(ctx, history); err != nil {
		return err
	}

	if err := uc.calculateAndRecordScores(ctx, member, userID); err != nil {
		return err
	}

	return nil
}

func (uc *memberUseCase) Update(ctx context.Context, member *domain.Member, expectedVersion, userID int) error {
	oldMember, err := uc.repo.member.Get(ctx, member.MemberID)
	if err != nil {
		return err
	}

	if len(member.Names) == 0 {
		return domain.
			NewValidationError("error.validation.names_required").
			WithParams(map[string]string{"language": "all", "code": "all"})
	}

	if oldMember.Gender != member.Gender {
		spouses, err := uc.repo.spouse.GetByMemberID(ctx, member.MemberID)
		if err != nil {
			return err
		}
		if len(spouses) > 0 {
			return domain.NewValidationError("error.validation.invalid_gender")
		}
	}

	if err := uc.validator.relationship.CheckParents(ctx, member.MemberID, member.FatherID, member.MotherID); err != nil {
		return err
	}

	if err := uc.validator.birthDate.Update(ctx, member.MemberID, member.DateOfBirth); err != nil {
		return err
	}

	if err := uc.validator.birthDate.Create(ctx, member.DateOfBirth, member.FatherID, member.MotherID); err != nil {
		return err
	}

	if err := uc.repo.member.Update(ctx, member, expectedVersion); err != nil {
		return err
	}

	if err := uc.ensureParentSpouseRelationship(ctx, member.FatherID, member.MotherID, userID); err != nil {
		return err
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
	if err := uc.repo.history.Create(ctx, history); err != nil {
		return err
	}

	if err := uc.updateScores(ctx, oldMember, member, userID); err != nil {
		return err
	}

	return nil
}

func (uc *memberUseCase) Delete(ctx context.Context, memberID, userID int) error {
	oldMember, err := uc.repo.member.Get(ctx, memberID)
	if err != nil {
		return err
	}

	children, err := uc.repo.member.GetChildrenByParentID(ctx, memberID)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return domain.NewConflictError("error.member.has_children", map[string]string{"count": fmt.Sprintf("%d", len(children))})
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
	if err := uc.repo.history.Create(ctx, history); err != nil {
		slog.Error("record delete history", "error", err, "member_id", memberID)
	}

	pictureURL, err := uc.repo.member.Delete(ctx, memberID)
	if err != nil {
		return err
	}

	if pictureURL != nil && *pictureURL != "" {
		if err := uc.s3Client.DeleteImage(ctx, *pictureURL); err != nil {
			slog.Error("delete member picture from storage", "error", err, "member_id", memberID, "picture", *pictureURL)
		}
	}

	return nil
}

func (uc *memberUseCase) Get(ctx context.Context, memberID int) (*domain.Member, error) {
	return uc.repo.member.Get(ctx, memberID)
}

func (uc *memberUseCase) ListChildren(ctx context.Context, parentID int) ([]*domain.Member, error) {
	return uc.repo.member.GetChildrenByParentID(ctx, parentID)
}

func (uc *memberUseCase) ListSiblings(ctx context.Context, memberID int) ([]*domain.Member, error) {
	return uc.repo.member.GetSiblingsByMemberID(ctx, memberID)
}

func (uc *memberUseCase) List(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error) {
	return uc.repo.member.List(ctx, filter, cursor, limit)
}

func (uc *memberUseCase) ListHistory(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	return uc.repo.history.GetByMemberID(ctx, memberID, cursor, limit)
}

func (uc *memberUseCase) UploadPicture(ctx context.Context, memberID int, data []byte, filename string, userID int) (string, error) {
	oldMember, err := uc.repo.member.Get(ctx, memberID)
	if err != nil {
		return "", err
	}

	newPictureURL, err := uc.s3Client.UploadImage(ctx, data, filename)
	if err != nil {
		return "", err
	}

	if err := uc.repo.member.UpdatePicture(ctx, memberID, newPictureURL); err != nil {
		if deleteErr := uc.s3Client.DeleteImage(ctx, newPictureURL); deleteErr != nil {
			slog.Error("rollback S3 upload after DB error",
				"error", deleteErr,
				"member_id", memberID,
				"picture_url", newPictureURL,
				"original_error", err)
		}
		return "", err
	}

	if oldMember.Picture != nil && *oldMember.Picture != "" {
		if err := uc.s3Client.DeleteImage(ctx, *oldMember.Picture); err != nil {
			slog.Warn("delete old picture from S3",
				"error", err,
				"member_id", memberID,
				"old_picture", *oldMember.Picture)
		}
	}

	updatedMember, err := uc.repo.member.Get(ctx, memberID)
	if err != nil {
		return "", err
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
	if err := uc.repo.history.Create(ctx, history); err != nil {
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
		_ = uc.repo.score.Create(ctx, scores...)
	}

	return newPictureURL, nil
}

func (uc *memberUseCase) DeletePicture(ctx context.Context, memberID int, userID int) error {
	member, err := uc.repo.member.Get(ctx, memberID)
	if err != nil {
		return err
	}

	oldPictureURL := member.Picture

	if err := uc.repo.member.DeletePicture(ctx, memberID); err != nil {
		return err
	}

	if oldPictureURL != nil && *oldPictureURL != "" {
		if err := uc.s3Client.DeleteImage(ctx, *oldPictureURL); err != nil {
			slog.Warn("delete picture from S3 after DB update",
				"error", err,
				"member_id", memberID,
				"picture_url", *oldPictureURL)
		}
	}

	updatedMember, err := uc.repo.member.Get(ctx, memberID)
	if err != nil {
		return err
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
	if err := uc.repo.history.Create(ctx, history); err != nil {
		slog.Error("create history for picture deletion", "error", err, "member_id", memberID)
	}

	return nil
}

func (uc *memberUseCase) GetPicture(ctx context.Context, memberID int) ([]byte, string, error) {
	member, err := uc.repo.member.Get(ctx, memberID)
	if err != nil {
		return nil, "", err
	}

	if member.Picture == nil || *member.Picture == "" {
		slog.Warn("memberUseCase.GetPicture: picture not found", "member_id", memberID)
		return nil, "", domain.NewNotFoundError("picture")
	}

	imageData, err := uc.s3Client.GetImage(ctx, *member.Picture)
	if err != nil {
		return nil, "", err
	}

	contentType := mime.TypeByExtension(filepath.Ext(*member.Picture))

	return imageData, contentType, nil
}

func (uc *memberUseCase) ensureParentSpouseRelationship(ctx context.Context, fatherID, motherID *int, userID int) error {
	if fatherID == nil || motherID == nil {
		return nil
	}

	existingSpouse, err := uc.repo.spouse.GetByParents(ctx, *fatherID, *motherID)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return err
	}
	if existingSpouse != nil {
		return nil
	}

	spouse := &domain.Spouse{
		FatherID:     *fatherID,
		MotherID:     *motherID,
		MarriageDate: nil,
		DivorceDate:  nil,
	}

	if err := uc.repo.spouse.Create(ctx, spouse); err != nil {
		return err
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

	return uc.repo.score.Create(ctx, scores...)

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

	return uc.repo.score.Create(ctx, scores...)
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

	spouses, _ := uc.repo.spouse.GetByMemberID(ctx, member.MemberID)
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
		maxDepth := 100 // Prevent infinite loops
		depth := 0

		for currentFatherID != nil && depth < maxDepth {
			father, err := uc.repo.member.Get(ctx, *currentFatherID)
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
