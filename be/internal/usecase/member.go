package usecase

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/escalopa/family-tree-api/internal/repository"
)

type MemberUseCase interface {
	Create(ctx context.Context, member *domain.Member, userID int) (*domain.MemberWithDetails, error)
	GetByID(ctx context.Context, memberID int, userRoleID int) (*domain.MemberWithDetails, error)
	Update(ctx context.Context, member *domain.Member, userID int) (*domain.MemberWithDetails, error)
	Delete(ctx context.Context, memberID int, userID int) error
	Search(ctx context.Context, filter MemberFilter) ([]*domain.MemberWithDetails, error)
	UploadPicture(ctx context.Context, memberID int, picture []byte, revision int, userID int) (int, error)
	DeletePicture(ctx context.Context, memberID int, revision int, userID int) error
	GetPicture(ctx context.Context, memberID int, userRoleID int, memberGender string) ([]byte, error)
	GetPictures(ctx context.Context, memberIDs []int, userRoleID int) (map[int]string, error)
	Rollback(ctx context.Context, memberID, revision int, userID int) (*domain.MemberWithDetails, error)
}

type memberUseCase struct {
	memberRepo  MemberRepo
	historyRepo HistoryRepo
	scoreRepo   ScoreRepo
	spouseRepo  SpouseRepo
}

// Local interfaces for repositories used by this use case
type MemberRepo interface {
	Create(ctx context.Context, member *domain.Member) error
	GetByID(ctx context.Context, memberID int) (*domain.MemberWithDetails, error)
	Update(ctx context.Context, member *domain.Member) error
	SoftDelete(ctx context.Context, memberID int) error
	Search(ctx context.Context, filter MemberFilter) ([]*domain.MemberWithDetails, error)
	UpdatePicture(ctx context.Context, memberID int, picture []byte) error
	DeletePicture(ctx context.Context, memberID int) error
	GetPicture(ctx context.Context, memberID int) ([]byte, error)
	GetPictures(ctx context.Context, memberIDs []int) (map[int][]byte, error)
	IncrementRevision(ctx context.Context, memberID int) (int, error)
	GetRevision(ctx context.Context, memberID int) (int, error)
}

type MemberFilter struct {
	ArabicName  *string
	EnglishName *string
	Gender      *string
	Married     *bool
}

type HistoryRepo interface {
	Create(ctx context.Context, history *domain.MemberHistory) error
}

type SpouseRepo interface {
	Exists(ctx context.Context, member1ID, member2ID int) (bool, error)
}

func NewMemberUseCase(
	memberRepo MemberRepo,
	historyRepo HistoryRepo,
	scoreRepo ScoreRepo,
	spouseRepo SpouseRepo,
) MemberUseCase {
	return &memberUseCase{
		memberRepo:  memberRepo,
		historyRepo: historyRepo,
		scoreRepo:   scoreRepo,
		spouseRepo:  spouseRepo,
	}
}

func (uc *memberUseCase) Create(ctx context.Context, member *domain.Member, userID int) (*domain.MemberWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *memberUseCase) GetByID(ctx context.Context, memberID int, userRoleID int) (*domain.MemberWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *memberUseCase) Update(ctx context.Context, member *domain.Member, userID int) (*domain.MemberWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *memberUseCase) Delete(ctx context.Context, memberID int, userID int) error {
	// TODO: Implementation
	return nil
}

func (uc *memberUseCase) Search(ctx context.Context, filter repository.MemberFilter) ([]*domain.MemberWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *memberUseCase) UploadPicture(ctx context.Context, memberID int, picture []byte, revision int, userID int) (int, error) {
	// TODO: Implementation
	return 0, nil
}

func (uc *memberUseCase) DeletePicture(ctx context.Context, memberID int, revision int, userID int) error {
	// TODO: Implementation
	return nil
}

func (uc *memberUseCase) GetPicture(ctx context.Context, memberID int, userRoleID int, memberGender string) ([]byte, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *memberUseCase) GetPictures(ctx context.Context, memberIDs []int, userRoleID int) (map[int]string, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *memberUseCase) Rollback(ctx context.Context, memberID, revision int, userID int) (*domain.MemberWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *memberUseCase) CreateSpouse(ctx context.Context, spouse *domain.MemberSpouse, userID int) error {
	// TODO: Implementation
	return nil
}

func (uc *memberUseCase) UpdateSpouse(ctx context.Context, spouse *domain.MemberSpouse, userID int) error {
	// TODO: Implementation
	return nil
}

func (uc *memberUseCase) DeleteSpouse(ctx context.Context, member1ID, member2ID int, userID int) error {
	// TODO: Implementation
	return nil
}
