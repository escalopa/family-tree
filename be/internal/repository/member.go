package repository

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberRepository struct {
	db *pgxpool.Pool
}

func NewMemberRepository(db *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{
		db: db,
	}
}

func (r *MemberRepository) GetMemberNames(ctx context.Context, memberID int) (map[string]string, error) {
	query := `
		SELECT language_code, name
		FROM member_names
		WHERE member_id = $1
	`
	rows, err := r.db.Query(ctx, query, memberID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	names := make(map[string]string)
	for rows.Next() {
		var langCode, name string
		if err := rows.Scan(&langCode, &name); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		names[langCode] = name
	}

	return names, nil
}

func (r *MemberRepository) GetMemberNamesByIDs(ctx context.Context, memberIDs []int) (map[int]map[string]string, error) {
	if len(memberIDs) == 0 {
		return make(map[int]map[string]string), nil
	}

	query := `
		SELECT member_id, language_code, name
		FROM member_names
		WHERE member_id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, memberIDs)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	namesMap := make(map[int]map[string]string)
	for rows.Next() {
		var memberID int
		var langCode, name string
		if err := rows.Scan(&memberID, &langCode, &name); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		if namesMap[memberID] == nil {
			namesMap[memberID] = make(map[string]string)
		}
		namesMap[memberID][langCode] = name
	}

	return namesMap, nil
}

func (r *MemberRepository) Create(ctx context.Context, member *domain.Member) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO members (gender, picture, date_of_birth, date_of_death,
		                     father_id, mother_id, nicknames, profession, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1)
		RETURNING member_id, version
	`
	err = tx.QueryRow(ctx, query,
		member.Gender, member.Picture,
		member.DateOfBirth, member.DateOfDeath, member.FatherID, member.MotherID,
		member.Nicknames, member.Profession,
	).Scan(&member.MemberID, &member.Version)
	if err != nil {
		return domain.NewDatabaseError(err)
	}

	batch := &pgx.Batch{}
	nameQuery := `
			INSERT INTO member_names (member_id, language_code, name, created_at, updated_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
	for langCode, name := range member.Names {
		batch.Queue(nameQuery, member.MemberID, langCode, name)
	}

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		return domain.NewDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.NewDatabaseError(err)
	}

	return nil
}

func (r *MemberRepository) GetByID(ctx context.Context, memberID int) (*domain.Member, error) {
	query := `
		SELECT member_id, gender, picture, date_of_birth, date_of_death,
		       father_id, mother_id, nicknames, profession, version, deleted_at
		FROM members
		WHERE member_id = $1 AND deleted_at IS NULL
	`
	member := &domain.Member{}
	err := r.db.QueryRow(ctx, query, memberID).Scan(
		&member.MemberID, &member.Gender,
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

	names, err := r.GetMemberNames(ctx, memberID)
	if err != nil {
		return nil, err
	}
	member.Names = names

	return member, nil
}

func (r *MemberRepository) Update(ctx context.Context, member *domain.Member, expectedVersion int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE members
		SET gender = $1, picture = $2, date_of_birth = $3,
		    date_of_death = $4, father_id = $5, mother_id = $6, nicknames = $7, profession = $8,
		    version = version + 1
		WHERE member_id = $9 AND version = $10 AND deleted_at IS NULL
		RETURNING version
	`
	err = tx.QueryRow(ctx, query,
		member.Gender, member.Picture,
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

	batch := &pgx.Batch{}
	nameQuery := `
			INSERT INTO member_names (member_id, language_code, name, created_at, updated_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (member_id, language_code)
			DO UPDATE SET
				name = EXCLUDED.name,
				updated_at = CURRENT_TIMESTAMP
		`
	for langCode, name := range member.Names {
		batch.Queue(nameQuery, member.MemberID, langCode, name)
	}

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		return domain.NewDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.NewDatabaseError(err)
	}

	return nil
}

// Delete performs a soft delete and cleans up all member data
// Returns the picture URL if one exists, for cleanup from storage
func (r *MemberRepository) Delete(ctx context.Context, memberID int) (*string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer tx.Rollback(ctx)

	var pictureURL *string
	deleteQuery := `
		UPDATE members
		SET deleted_at = NOW()
		WHERE member_id = $1 AND deleted_at IS NULL
		RETURNING picture
	`
	err = tx.QueryRow(ctx, deleteQuery, memberID).Scan(&pictureURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("MemberRepository.Delete: member not found or already deleted", "member_id", memberID)
			return nil, domain.NewNotFoundError("member")
		}
		return nil, domain.NewDatabaseError(err)
	}

	namesQuery := `DELETE FROM member_names WHERE member_id = $1`
	_, err = tx.Exec(ctx, namesQuery, memberID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	spouseQuery := `
		UPDATE members_spouse
		SET deleted_at = NOW()
		WHERE (father_id = $1 OR mother_id = $1)
		AND deleted_at IS NULL
	`
	_, err = tx.Exec(ctx, spouseQuery, memberID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	return pictureURL, nil
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
		SELECT DISTINCT m.member_id, m.gender, m.picture, m.date_of_birth,
		       m.date_of_death, m.father_id, m.mother_id, m.nicknames, m.profession, m.version, m.deleted_at,
		       CASE WHEN COUNT(ms.spouse_id) > 0 THEN true ELSE false END as is_married
		FROM members m
		LEFT JOIN members_spouse ms ON (m.member_id = ms.father_id OR m.member_id = ms.mother_id) AND ms.deleted_at IS NULL
		LEFT JOIN member_names mn ON m.member_id = mn.member_id
		WHERE m.deleted_at IS NULL
		  AND (($1::text IS NULL) OR m.member_id > $1::int)
		  AND (($2::text IS NULL) OR (mn.name ILIKE '%' || $2 || '%'))
		  AND (($3::text IS NULL) OR m.gender = $3)
		  AND (($4::boolean IS NULL) OR (
		    CASE
		      WHEN $4 = true THEN (ms.father_id IS NOT NULL OR ms.mother_id IS NOT NULL)
		      ELSE (ms.father_id IS NULL AND ms.mother_id IS NULL)
		    END
		  ))
		GROUP BY m.member_id, m.gender, m.picture, m.date_of_birth,
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
		limit,
	)
	if err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var members []*domain.Member
	var memberIDs []int
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath, &member.FatherID,
			&member.MotherID, &member.Nicknames, &member.Profession, &member.Version, &member.DeletedAt,
			&member.IsMarried,
		)
		if err != nil {
			return nil, nil, domain.NewDatabaseError(err)
		}
		members = append(members, member)
		memberIDs = append(memberIDs, member.MemberID)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}

	var nextCursor *string
	if len(members) == limit {
		lastMemberID := strconv.Itoa(members[len(members)-1].MemberID)
		nextCursor = &lastMemberID
	}

	namesMap, err := r.GetMemberNamesByIDs(ctx, memberIDs)
	if err != nil {
		return nil, nil, err
	}
	for _, member := range members {
		member.Names = namesMap[member.MemberID]
	}

	return members, nextCursor, nil
}

func (r *MemberRepository) GetAll(ctx context.Context) ([]*domain.Member, error) {
	query := `
		SELECT member_id, gender, picture, date_of_birth, date_of_death,
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
	var memberIDs []int
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath, &member.FatherID,
			&member.MotherID, &member.Nicknames, &member.Profession, &member.Version, &member.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		members = append(members, member)
		memberIDs = append(memberIDs, member.MemberID)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	namesMap, err := r.GetMemberNamesByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		member.Names = namesMap[member.MemberID]
	}

	return members, nil
}

func (r *MemberRepository) GetByIDs(ctx context.Context, memberIDs []int) ([]*domain.Member, error) {
	if len(memberIDs) == 0 {
		return []*domain.Member{}, nil
	}

	query := `
		SELECT member_id, gender, picture, date_of_birth, date_of_death,
		       father_id, mother_id, nicknames, profession, version, deleted_at
		FROM members
		WHERE member_id = ANY($1) AND deleted_at IS NULL
	`

	rows, err := r.db.Query(ctx, query, memberIDs)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.Gender,
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

	namesMap, err := r.GetMemberNamesByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		member.Names = namesMap[member.MemberID]
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
		SELECT member_id, gender, picture, date_of_birth, date_of_death,
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
	var memberIDs []int
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath,
			&member.FatherID, &member.MotherID, &member.Nicknames, &member.Profession,
			&member.Version, &member.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		children = append(children, member)
		memberIDs = append(memberIDs, member.MemberID)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	namesMap, err := r.GetMemberNamesByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	for _, member := range children {
		member.Names = namesMap[member.MemberID]
	}

	return children, nil
}

func (r *MemberRepository) GetSiblingsByMemberID(ctx context.Context, memberID int) ([]*domain.Member, error) {
	query := `
		SELECT DISTINCT m.member_id, m.gender, m.picture,
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
	var memberIDs []int
	for rows.Next() {
		member := &domain.Member{}
		err := rows.Scan(
			&member.MemberID, &member.Gender,
			&member.Picture, &member.DateOfBirth, &member.DateOfDeath,
			&member.FatherID, &member.MotherID, &member.Nicknames, &member.Profession,
			&member.Version, &member.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		siblings = append(siblings, member)
		memberIDs = append(memberIDs, member.MemberID)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	namesMap, err := r.GetMemberNamesByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	for _, member := range siblings {
		member.Names = namesMap[member.MemberID]
	}

	return siblings, nil
}
