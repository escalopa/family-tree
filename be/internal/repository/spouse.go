package repository

import (
	"context"
	"errors"
	"log/slog"

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
		INSERT INTO members_spouse (father_id, mother_id, marriage_date, divorce_date)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (father_id, mother_id)
		DO UPDATE SET
			marriage_date = EXCLUDED.marriage_date,
			divorce_date = EXCLUDED.divorce_date,
			deleted_at = NULL
		RETURNING spouse_id
	`
	err := r.db.QueryRow(ctx, query, spouse.FatherID, spouse.MotherID, spouse.MarriageDate, spouse.DivorceDate).Scan(&spouse.SpouseID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *SpouseRepository) Get(ctx context.Context, spouseID int) (*domain.Spouse, error) {
	query := `
		SELECT spouse_id, father_id, mother_id, marriage_date, divorce_date, deleted_at
		FROM members_spouse
		WHERE spouse_id = $1 AND deleted_at IS NULL
	`
	spouse := &domain.Spouse{}
	err := r.db.QueryRow(ctx, query, spouseID).Scan(
		&spouse.SpouseID, &spouse.FatherID, &spouse.MotherID, &spouse.MarriageDate, &spouse.DivorceDate, &spouse.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn("SpouseRepository.Get: spouse relationship not found", "spouse_id", spouseID)
		return nil, domain.NewNotFoundError("spouse relationship")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return spouse, nil
}

func (r *SpouseRepository) GetByParents(ctx context.Context, fatherID, motherID int) (*domain.Spouse, error) {
	query := `
		SELECT spouse_id, father_id, mother_id, marriage_date, divorce_date, deleted_at
		FROM members_spouse
		WHERE father_id = $1 AND mother_id = $2 AND deleted_at IS NULL
	`
	spouse := &domain.Spouse{}
	err := r.db.QueryRow(ctx, query, fatherID, motherID).Scan(
		&spouse.SpouseID, &spouse.FatherID, &spouse.MotherID, &spouse.MarriageDate, &spouse.DivorceDate, &spouse.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn("SpouseRepository.Get: spouse relationship not found", "father_id", fatherID, "mother_id", motherID)
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
		WHERE spouse_id = $3 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, spouse.MarriageDate, spouse.DivorceDate, spouse.SpouseID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		slog.Warn("SpouseRepository.Update: spouse relationship not found", "spouse_id", spouse.SpouseID)
		return domain.NewNotFoundError("spouse relationship")
	}
	return nil
}

func (r *SpouseRepository) Delete(ctx context.Context, spouseID int) error {
	query := `
		UPDATE members_spouse
		SET deleted_at = NOW()
		WHERE spouse_id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, spouseID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		slog.Warn("SpouseRepository.Delete: spouse relationship not found", "spouse_id", spouseID)
		return domain.NewNotFoundError("spouse relationship")
	}
	return nil
}

func (r *SpouseRepository) GetAllSpouses(ctx context.Context) (map[int][]int, error) {
	query := `
		SELECT ms.father_id, ms.mother_id
		FROM members_spouse ms
		JOIN members m1 ON m1.member_id = ms.father_id
		JOIN members m2 ON m2.member_id = ms.mother_id
		WHERE ms.deleted_at IS NULL
		  AND m1.deleted_at IS NULL
		  AND m2.deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	spouseMap := make(map[int][]int)
	for rows.Next() {
		var fatherID, motherID int
		if err := rows.Scan(&fatherID, &motherID); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		spouseMap[fatherID] = append(spouseMap[fatherID], motherID)
		spouseMap[motherID] = append(spouseMap[motherID], fatherID)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return spouseMap, nil
}

func (r *SpouseRepository) GetByMemberID(ctx context.Context, memberID int) ([]domain.SpouseWithMemberInfo, error) {
	query := `
		SELECT
			ms.spouse_id,
			m.member_id,
			m.gender,
			m.picture,
			ms.marriage_date,
			ms.divorce_date
		FROM members_spouse ms
		JOIN members m ON (
			(ms.father_id = $1 AND m.member_id = ms.mother_id) OR
			(ms.mother_id = $1 AND m.member_id = ms.father_id)
		)
		WHERE ms.deleted_at IS NULL AND m.deleted_at IS NULL
		ORDER BY ms.marriage_date ASC NULLS LAST, m.date_of_birth ASC NULLS LAST, m.member_id ASC
	`
	rows, err := r.db.Query(ctx, query, memberID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var spouses []domain.SpouseWithMemberInfo
	var memberIDs []int
	for rows.Next() {
		var spouse domain.SpouseWithMemberInfo
		if err := rows.Scan(
			&spouse.SpouseID,
			&spouse.MemberID,
			&spouse.Gender,
			&spouse.Picture,
			&spouse.MarriageDate,
			&spouse.DivorceDate,
		); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		spouses = append(spouses, spouse)
		memberIDs = append(memberIDs, spouse.MemberID)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	// Fetch names for all spouses
	if len(memberIDs) > 0 {
		namesQuery := `
			SELECT member_id, language_code, name
			FROM member_names
			WHERE member_id = ANY($1)
		`
		nameRows, err := r.db.Query(ctx, namesQuery, memberIDs)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		defer nameRows.Close()

		namesMap := make(map[int]map[string]string)
		for nameRows.Next() {
			var mid int
			var langCode, name string
			if err := nameRows.Scan(&mid, &langCode, &name); err != nil {
				return nil, domain.NewDatabaseError(err)
			}
			if namesMap[mid] == nil {
				namesMap[mid] = make(map[string]string)
			}
			namesMap[mid][langCode] = name
		}

		// Assign names to spouses
		for i := range spouses {
			if names, ok := namesMap[spouses[i].MemberID]; ok {
				spouses[i].Names = names
			} else {
				spouses[i].Names = make(map[string]string)
			}
		}
	}

	return spouses, nil
}
