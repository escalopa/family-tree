package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type MemberRepository struct {
	db *pgxpool.Pool
}

func NewMemberRepository(db *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(ctx context.Context, member *domain.Member) error {
	query := `
		INSERT INTO members (arabic_name, english_name, gender, picture, date_of_birth, date_of_death,
		                     father_id, mother_id, nicknames, profession, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1)
		RETURNING member_id, version
	`
	err := r.db.QueryRow(ctx, query,
		member.ArabicName, member.EnglishName, member.Gender, member.Picture,
		member.DateOfBirth, member.DateOfDeath, member.FatherID, member.MotherID,
		member.Nicknames, member.Profession,
	).Scan(&member.MemberID, &member.Version)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *MemberRepository) GetByID(ctx context.Context, memberID int) (*domain.Member, error) {
	query := `
		SELECT member_id, arabic_name, english_name, gender, picture, date_of_birth, date_of_death,
		       father_id, mother_id, nicknames, profession, version, deleted_at
		FROM members
		WHERE member_id = $1 AND deleted_at IS NULL
	`
	member := &domain.Member{}
	err := r.db.QueryRow(ctx, query, memberID).Scan(
		&member.MemberID, &member.ArabicName, &member.EnglishName, &member.Gender,
		&member.Picture, &member.DateOfBirth, &member.DateOfDeath, &member.FatherID,
		&member.MotherID, &member.Nicknames, &member.Profession, &member.Version, &member.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn("MemberRepository.GetByID: member not found", "member_id", memberID)
		return nil, domain.NewNotFoundError("member")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return member, nil
}

func (r *MemberRepository) Update(ctx context.Context, member *domain.Member, expectedVersion int) error {
	query := `
		UPDATE members
		SET arabic_name = $1, english_name = $2, gender = $3, picture = $4, date_of_birth = $5,
		    date_of_death = $6, father_id = $7, mother_id = $8, nicknames = $9, profession = $10,
		    version = version + 1
		WHERE member_id = $11 AND version = $12 AND deleted_at IS NULL
		RETURNING version
	`
	err := r.db.QueryRow(ctx, query,
		member.ArabicName, member.EnglishName, member.Gender, member.Picture,
		member.DateOfBirth, member.DateOfDeath, member.FatherID, member.MotherID,
		member.Nicknames, member.Profession, member.MemberID, expectedVersion,
	).Scan(&member.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn("MemberRepository.Update: version conflict", "member_id", member.MemberID, "expected_version", expectedVersion)
		return domain.NewVersionConflictError()
	}
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *MemberRepository) SoftDelete(ctx context.Context, memberID int) error {
	query := `UPDATE members SET deleted_at = NOW() WHERE member_id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, memberID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		slog.Warn("MemberRepository.SoftDelete: member not found", "member_id", memberID)
		return domain.NewNotFoundError("member")
	}
	return nil
}

func (r *MemberRepository) UpdatePicture(ctx context.Context, memberID int, pictureURL string) error {
	query := `UPDATE members SET picture = $1, version = version + 1 WHERE member_id = $2 AND deleted_at IS NULL RETURNING version`
	var version int
	err := r.db.QueryRow(ctx, query, pictureURL, memberID).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn("MemberRepository.UpdatePicture: member not found", "member_id", memberID)
		return domain.NewNotFoundError("member")
	}
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *MemberRepository) DeletePicture(ctx context.Context, memberID int) error {
	query := `UPDATE members SET picture = NULL, version = version + 1 WHERE member_id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, memberID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *MemberRepository) Search(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error) {
	query := `
		SELECT DISTINCT m.member_id, m.arabic_name, m.english_name, m.gender, m.picture, m.date_of_birth,
		       m.date_of_death, m.father_id, m.mother_id, m.nicknames, m.profession, m.version, m.deleted_at,
		       CASE WHEN COUNT(ms.spouse_id) > 0 THEN true ELSE false END as is_married
		FROM members m
		LEFT JOIN members_spouse ms ON (m.member_id = ms.father_id OR m.member_id = ms.mother_id) AND ms.deleted_at IS NULL
		WHERE m.deleted_at IS NULL
		  AND (($1::text IS NULL) OR m.member_id > $1::int)
		  AND (($2::text IS NULL) OR (m.arabic_name ILIKE '%' || $2 || '%' OR m.english_name ILIKE '%' || $2 || '%'))
		  AND (($3::text IS NULL) OR m.gender = $3)
		  AND (($4::boolean IS NULL) OR (
		    CASE
		      WHEN $4 = true THEN (ms.father_id IS NOT NULL OR ms.mother_id IS NOT NULL)
		      ELSE (ms.father_id IS NULL AND ms.mother_id IS NULL)
		    END
		  ))
		GROUP BY m.member_id, m.arabic_name, m.english_name, m.gender, m.picture, m.date_of_birth,
		         m.date_of_death, m.father_id, m.mother_id, m.nicknames, m.profession, m.version, m.deleted_at
		ORDER BY m.member_id
		LIMIT $5
	`

	var cursorValue *string
	if cursor != nil && *cursor != "" {
		cursorValue = cursor
	}

	rows, err := r.db.Query(ctx, query,
		cursorValue,
		filter.Name,
		filter.Gender,
		filter.IsMarried,
		limit+1,
	)
	if err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.ArabicName, &member.EnglishName, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath, &member.FatherID,
			&member.MotherID, &member.Nicknames, &member.Profession, &member.Version, &member.DeletedAt,
			&member.IsMarried,
		)
		if err != nil {
			return nil, nil, domain.NewDatabaseError(err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}

	// Determine next cursor
	var nextCursor *string
	if len(members) > limit {
		// Remove the extra member and set cursor
		members = members[:limit]
		if len(members) > 0 {
			lastMemberID := fmt.Sprintf("%d", members[len(members)-1].MemberID)
			nextCursor = &lastMemberID
		}
	}

	return members, nextCursor, nil
}

func (r *MemberRepository) GetAll(ctx context.Context) ([]*domain.Member, error) {
	query := `
		SELECT member_id, arabic_name, english_name, gender, picture, date_of_birth, date_of_death,
		       father_id, mother_id, nicknames, profession, version, deleted_at
		FROM members
		WHERE deleted_at IS NULL
		ORDER BY member_id
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.ArabicName, &member.EnglishName, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath, &member.FatherID,
			&member.MotherID, &member.Nicknames, &member.Profession, &member.Version, &member.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return members, nil
}

func (r *MemberRepository) GetByIDs(ctx context.Context, memberIDs []int) ([]*domain.Member, error) {
	if len(memberIDs) == 0 {
		return []*domain.Member{}, nil
	}

	query := `
		SELECT member_id, arabic_name, english_name, gender, picture, date_of_birth, date_of_death,
		       father_id, mother_id, nicknames, profession, version, deleted_at
		FROM members
		WHERE member_id IN ($1) AND deleted_at IS NULL
	`

	rows, err := r.db.Query(ctx, query, pq.Array(memberIDs))
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.ArabicName, &member.EnglishName, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath, &member.FatherID,
			&member.MotherID, &member.Nicknames, &member.Profession, &member.Version, &member.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return members, nil
}

func (r *MemberRepository) HasChildrenWithParents(ctx context.Context, fatherID, motherID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM members
			WHERE deleted_at IS NULL
			  AND father_id = $1 AND mother_id = $2
		)
	`
	var hasChildren bool
	err := r.db.QueryRow(ctx, query, fatherID, motherID).Scan(&hasChildren)
	if err != nil {
		return false, domain.NewDatabaseError(err)
	}
	return hasChildren, nil
}

func (r *MemberRepository) GetChildrenByParentID(ctx context.Context, parentID int) ([]*domain.Member, error) {
	query := `
		SELECT member_id, arabic_name, english_name, gender, picture, date_of_birth, date_of_death,
		       father_id, mother_id, nicknames, profession, version, deleted_at
		FROM members
		WHERE deleted_at IS NULL
		  AND (father_id = $1 OR mother_id = $1)
		ORDER BY date_of_birth ASC NULLS LAST, member_id ASC
	`
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var children []*domain.Member
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.ArabicName, &member.EnglishName, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath,
			&member.FatherID, &member.MotherID, &member.Nicknames, &member.Profession,
			&member.Version, &member.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		children = append(children, member)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	return children, nil
}

func (r *MemberRepository) GetSiblingsByMemberID(ctx context.Context, memberID int) ([]*domain.Member, error) {
	query := `
		SELECT DISTINCT m.member_id, m.arabic_name, m.english_name, m.gender, m.picture,
		       m.date_of_birth, m.date_of_death, m.father_id, m.mother_id, m.nicknames,
		       m.profession, m.version, m.deleted_at
		FROM members m
		WHERE m.deleted_at IS NULL
		  AND m.member_id != $1
		  AND (
		    (m.father_id IS NOT NULL AND m.father_id IN (
		      SELECT father_id FROM members WHERE member_id = $1 AND father_id IS NOT NULL
		    ))
		    OR
		    (m.mother_id IS NOT NULL AND m.mother_id IN (
		      SELECT mother_id FROM members WHERE member_id = $1 AND mother_id IS NOT NULL
		    ))
		  )
		ORDER BY m.date_of_birth ASC NULLS LAST, m.member_id ASC
	`
	rows, err := r.db.Query(ctx, query, memberID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var siblings []*domain.Member
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.ArabicName, &member.EnglishName, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath,
			&member.FatherID, &member.MotherID, &member.Nicknames, &member.Profession,
			&member.Version, &member.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		siblings = append(siblings, member)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	return siblings, nil
}
