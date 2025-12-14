package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0)
		RETURNING member_id, version
	`
	return r.db.QueryRow(ctx, query,
		member.ArabicName, member.EnglishName, member.Gender, member.Picture,
		member.DateOfBirth, member.DateOfDeath, member.FatherID, member.MotherID,
		member.Nicknames, member.Profession,
	).Scan(&member.MemberID, &member.Version)
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
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("member not found")
	}
	return member, err
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
	if err == pgx.ErrNoRows {
		return fmt.Errorf("version conflict or member not found")
	}
	return err
}

func (r *MemberRepository) SoftDelete(ctx context.Context, memberID int) error {
	query := `UPDATE members SET deleted_at = NOW() WHERE member_id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, memberID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found or already deleted")
	}
	return nil
}

func (r *MemberRepository) UpdatePicture(ctx context.Context, memberID int, pictureURL string) error {
	query := `UPDATE members SET picture = $1, version = version + 1 WHERE member_id = $2 AND deleted_at IS NULL RETURNING version`
	var version int
	err := r.db.QueryRow(ctx, query, pictureURL, memberID).Scan(&version)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("member not found")
	}
	return err
}

func (r *MemberRepository) DeletePicture(ctx context.Context, memberID int) error {
	query := `UPDATE members SET picture = NULL, version = version + 1 WHERE member_id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, memberID)
	return err
}

func (r *MemberRepository) Search(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error) {
	query := `
		SELECT DISTINCT m.member_id, m.arabic_name, m.english_name, m.gender, m.picture, m.date_of_birth,
		       m.date_of_death, m.father_id, m.mother_id, m.nicknames, m.profession, m.version, m.deleted_at
		FROM members m
		LEFT JOIN members_spouse ms ON m.member_id = ms.member1_id OR m.member_id = ms.member2_id
		WHERE m.deleted_at IS NULL
	`
	var args []interface{}
	argCount := 1

	// Apply cursor-based pagination
	if cursor != nil && *cursor != "" {
		query += fmt.Sprintf(" AND m.member_id > $%d", argCount)
		args = append(args, *cursor)
		argCount++
	}

	if filter.ArabicName != nil {
		query += fmt.Sprintf(" AND m.arabic_name LIKE $%d", argCount)
		args = append(args, *filter.ArabicName+"%")
		argCount++
	}
	if filter.EnglishName != nil {
		query += fmt.Sprintf(" AND m.english_name LIKE $%d", argCount)
		args = append(args, *filter.EnglishName+"%")
		argCount++
	}
	if filter.Gender != nil {
		query += fmt.Sprintf(" AND m.gender = $%d", argCount)
		args = append(args, *filter.Gender)
		argCount++
	}
	if filter.IsMarried != nil {
		if *filter.IsMarried {
			query += " AND (ms.member1_id IS NOT NULL OR ms.member2_id IS NOT NULL)"
		} else {
			query += " AND ms.member1_id IS NULL AND ms.member2_id IS NULL"
		}
	}

	query += " ORDER BY m.member_id"

	// Fetch one extra to determine if there's a next page
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
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
			return nil, nil, err
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Determine next cursor
	var nextCursor *string
	if len(members) > limit {
		// Remove the extra member and set cursor
		members = members[:limit]
		lastMemberID := fmt.Sprintf("%d", members[len(members)-1].MemberID)
		nextCursor = &lastMemberID
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
		return nil, err
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
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (r *MemberRepository) GetByIDs(ctx context.Context, memberIDs []int) ([]*domain.Member, error) {
	if len(memberIDs) == 0 {
		return []*domain.Member{}, nil
	}

	placeholders := make([]string, len(memberIDs))
	args := make([]interface{}, len(memberIDs))
	for i, id := range memberIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT member_id, arabic_name, english_name, gender, picture, date_of_birth, date_of_death,
		       father_id, mother_id, nicknames, profession, version, deleted_at
		FROM members
		WHERE member_id IN (%s) AND deleted_at IS NULL
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}
