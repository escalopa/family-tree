package usecase

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
)

type SpouseUseCase interface {
	Create(ctx context.Context, spouse *domain.MemberSpouse, userID int) error
	Update(ctx context.Context, spouse *domain.MemberSpouse, userID int) error
	Delete(ctx context.Context, member1ID, member2ID int, userID int) error
}

type SpouseRepo interface {
	Create(ctx context.Context, spouse *domain.MemberSpouse) error
	Update(ctx context.Context, spouse *domain.MemberSpouse) error
	Delete(ctx context.Context, member1ID, member2ID int) error
}

type MemberRepo interface {
	GetByID(ctx context.Context, memberID int) (*domain.MemberWithDetails, error)
}

type HistoryRepo interface {
	Create(ctx context.Context, history *domain.MemberHistory) error
}

type ScoreRepo interface {
	CalculateScore(ctx context.Context, userID, memberID int) (int, error)
}

type spouseUseCase struct {
	spouseRepo  SpouseRepo
	memberRepo  MemberRepo
	historyRepo HistoryRepo
	scoreRepo   ScoreRepo
}

func NewSpouseUseCase(
	spouseRepo SpouseRepo,
	memberRepo MemberRepo,
	historyRepo HistoryRepo,
	scoreRepo ScoreRepo,
) SpouseUseCase {
	return &spouseUseCase{
		spouseRepo:  spouseRepo,
		memberRepo:  memberRepo,
		historyRepo: historyRepo,
		scoreRepo:   scoreRepo,
	}
}
