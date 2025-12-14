package repository

import (
	"context"
	"errors"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SpouseRepository struct {
	db *pgxpool.Pool
}

func NewSpouseRepository(db *pgxpool.Pool) *SpouseRepository {
	return &SpouseRepository{db: db}
}

func (r *SpouseRepository) Create(ctx context.Context, spouse *domain.Spouse) error {
	query := `
		INSERT INTO members_spouse (member1_id, member2_id, marriage_date, divorce_date)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, spouse.Member1ID, spouse.Member2ID, spouse.MarriageDate, spouse.DivorceDate)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *SpouseRepository) Get(ctx context.Context, member1ID, member2ID int) (*domain.Spouse, error) {
	query := `
		SELECT member1_id, member2_id, marriage_date, divorce_date
		FROM members_spouse
		WHERE (member1_id = $1 AND member2_id = $2) OR (member1_id = $2 AND member2_id = $1)
	`
	spouse := &domain.Spouse{}
	err := r.db.QueryRow(ctx, query, member1ID, member2ID).Scan(
		&spouse.Member1ID, &spouse.Member2ID, &spouse.MarriageDate, &spouse.DivorceDate,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("spouse relationship")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return spouse, nil
}

func (r *SpouseRepository) Update(ctx context.Context, spouse *domain.Spouse) error {
	query := `
		UPDATE members_spouse
		SET marriage_date = $1, divorce_date = $2
		WHERE (member1_id = $3 AND member2_id = $4) OR (member1_id = $4 AND member2_id = $3)
	`
	result, err := r.db.Exec(ctx, query, spouse.MarriageDate, spouse.DivorceDate, spouse.Member1ID, spouse.Member2ID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		return domain.NewNotFoundError("spouse relationship")
	}
	return nil
}

func (r *SpouseRepository) Delete(ctx context.Context, member1ID, member2ID int) error {
	query := `
		DELETE FROM members_spouse
		WHERE (member1_id = $1 AND member2_id = $2) OR (member1_id = $2 AND member2_id = $1)
	`
	result, err := r.db.Exec(ctx, query, member1ID, member2ID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		return domain.NewNotFoundError("spouse relationship")
	}
	return nil
}

func (r *SpouseRepository) GetSpousesByMemberID(ctx context.Context, memberID int) ([]int, error) {
	query := `
		SELECT
			CASE
				WHEN member1_id = $1 THEN member2_id
				ELSE member1_id
			END as spouse_id
		FROM members_spouse
		WHERE member1_id = $1 OR member2_id = $1
	`
	rows, err := r.db.Query(ctx, query, memberID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var spouseIDs []int
	for rows.Next() {
		var spouseID int
		if err := rows.Scan(&spouseID); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		spouseIDs = append(spouseIDs, spouseID)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return spouseIDs, nil
}

func (r *SpouseRepository) GetAllSpouses(ctx context.Context) (map[int][]int, error) {
	query := `SELECT member1_id, member2_id FROM members_spouse`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	spouseMap := make(map[int][]int)
	for rows.Next() {
		var member1ID, member2ID int
		if err := rows.Scan(&member1ID, &member2ID); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		spouseMap[member1ID] = append(spouseMap[member1ID], member2ID)
		spouseMap[member2ID] = append(spouseMap[member2ID], member1ID)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return spouseMap, nil
}
