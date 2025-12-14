package repository

import (
	"context"
	"fmt"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HistoryRepository struct {
	db *pgxpool.Pool
}

func NewHistoryRepository(db *pgxpool.Pool) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func (r *HistoryRepository) Create(ctx context.Context, history *domain.History) error {
	query := `
		INSERT INTO members_history (member_id, user_id, change_type, old_values, new_values, member_version)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING history_id, changed_at
	`
	return r.db.QueryRow(ctx, query,
		history.MemberID, history.UserID, history.ChangeType,
		history.OldValues, history.NewValues, history.MemberVersion,
	).Scan(&history.HistoryID, &history.ChangedAt)
}

func (r *HistoryRepository) GetByMemberID(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	query := `
		SELECT h.history_id, h.member_id, h.user_id, h.changed_at, h.change_type,
		       h.old_values, h.new_values, h.member_version, u.full_name, u.email
		FROM members_history h
		JOIN users u ON h.user_id = u.user_id
		WHERE h.member_id = $1
	`
	var args []interface{}
	args = append(args, memberID)
	argCount := 2

	// Apply cursor-based pagination (using changed_at as cursor)
	if cursor != nil && *cursor != "" {
		query += fmt.Sprintf(" AND h.changed_at < $%d", argCount)
		args = append(args, *cursor)
		argCount++
	}

	query += " ORDER BY h.changed_at DESC"

	// Fetch one extra to determine if there's a next page
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var histories []*domain.HistoryWithUser
	for rows.Next() {
		h := &domain.HistoryWithUser{}
		err := rows.Scan(
			&h.HistoryID, &h.MemberID, &h.UserID, &h.ChangedAt, &h.ChangeType,
			&h.OldValues, &h.NewValues, &h.MemberVersion, &h.UserFullName, &h.UserEmail,
		)
		if err != nil {
			return nil, nil, err
		}
		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Determine next cursor
	var nextCursor *string
	if len(histories) > limit {
		// Remove the extra history and set cursor
		histories = histories[:limit]
		lastChangedAt := histories[len(histories)-1].ChangedAt.Format("2006-01-02T15:04:05.999999Z07:00")
		nextCursor = &lastChangedAt
	}

	return histories, nextCursor, nil
}

func (r *HistoryRepository) GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	query := `
		SELECT h.history_id, h.member_id, h.user_id, h.changed_at, h.change_type,
		       h.old_values, h.new_values, h.member_version, u.full_name, u.email
		FROM members_history h
		JOIN users u ON h.user_id = u.user_id
		WHERE h.user_id = $1
	`
	var args []interface{}
	args = append(args, userID)
	argCount := 2

	// Apply cursor-based pagination (using changed_at as cursor)
	if cursor != nil && *cursor != "" {
		query += fmt.Sprintf(" AND h.changed_at < $%d", argCount)
		args = append(args, *cursor)
		argCount++
	}

	query += " ORDER BY h.changed_at DESC"

	// Fetch one extra to determine if there's a next page
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var histories []*domain.HistoryWithUser
	for rows.Next() {
		h := &domain.HistoryWithUser{}
		err := rows.Scan(
			&h.HistoryID, &h.MemberID, &h.UserID, &h.ChangedAt, &h.ChangeType,
			&h.OldValues, &h.NewValues, &h.MemberVersion, &h.UserFullName, &h.UserEmail,
		)
		if err != nil {
			return nil, nil, err
		}
		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Determine next cursor
	var nextCursor *string
	if len(histories) > limit {
		// Remove the extra history and set cursor
		histories = histories[:limit]
		lastChangedAt := histories[len(histories)-1].ChangedAt.Format("2006-01-02T15:04:05.999999Z07:00")
		nextCursor = &lastChangedAt
	}

	return histories, nextCursor, nil
}
