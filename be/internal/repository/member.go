package repository

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/jackc/pgx/v5"
)

type MemberFilter struct {
	ArabicName  *string
	EnglishName *string
	Gender      *string
	Married     *bool
}

type Member struct {
	db *pgx.Conn
}

func NewMember(db *pgx.Conn) *Member {
	return &Member{db: db}
}

func (r *Member) Create(ctx context.Context, member *domain.Member) error {
	// TODO: Implementation
	return nil
}

func (r *Member) GetByID(ctx context.Context, memberID int) (*domain.MemberWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *Member) Update(ctx context.Context, member *domain.Member) error {
	// TODO: Implementation
	return nil
}

func (r *Member) SoftDelete(ctx context.Context, memberID int) error {
	// TODO: Implementation
	return nil
}

func (r *Member) Search(ctx context.Context, filter MemberFilter) ([]*domain.MemberWithDetails, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *Member) UpdatePicture(ctx context.Context, memberID int, picture []byte) error {
	// TODO: Implementation
	return nil
}

func (r *Member) DeletePicture(ctx context.Context, memberID int) error {
	// TODO: Implementation
	return nil
}

func (r *Member) GetPicture(ctx context.Context, memberID int) ([]byte, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *Member) GetPictures(ctx context.Context, memberIDs []int) (map[int][]byte, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *Member) IncrementRevision(ctx context.Context, memberID int) (int, error) {
	// TODO: Implementation
	return 0, nil
}

func (r *Member) GetRevision(ctx context.Context, memberID int) (int, error) {
	// TODO: Implementation
	return 0, nil
}
