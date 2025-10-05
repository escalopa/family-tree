package repository

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/jackc/pgx/v5"
)

type SpouseRepository interface {
	Create(ctx context.Context, spouse *domain.MemberSpouse) error
	Update(ctx context.Context, spouse *domain.MemberSpouse) error
	Delete(ctx context.Context, member1ID, member2ID int) error
	GetByMemberID(ctx context.Context, memberID int) ([]*domain.SpouseInfo, error)
	Exists(ctx context.Context, member1ID, member2ID int) (bool, error)
}

type spouseRepository struct {
	db *pgx.Conn
}

func NewSpouseRepository(db *pgx.Conn) SpouseRepository {
	return &spouseRepository{db: db}
}

func (r *spouseRepository) Create(ctx context.Context, spouse *domain.MemberSpouse) error {
	// TODO: Implementation
	return nil
}

func (r *spouseRepository) Update(ctx context.Context, spouse *domain.MemberSpouse) error {
	// TODO: Implementation
	return nil
}

func (r *spouseRepository) Delete(ctx context.Context, member1ID, member2ID int) error {
	// TODO: Implementation
	return nil
}

func (r *spouseRepository) GetByMemberID(ctx context.Context, memberID int) ([]*domain.SpouseInfo, error) {
	// TODO: Implementation
	return nil, nil
}

func (r *spouseRepository) Exists(ctx context.Context, member1ID, member2ID int) (bool, error) {
	// TODO: Implementation
	return false, nil
}
