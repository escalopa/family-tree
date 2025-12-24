package repository

import (
	"context"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
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
	err := r.db.QueryRow(ctx, query,
		history.MemberID, history.UserID, history.ChangeType,
		history.OldValues, history.NewValues, history.MemberVersion,
	).Scan(&history.HistoryID, &history.ChangedAt)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *HistoryRepository) CreateBatch(ctx context.Context, histories ...*domain.History) error {
	if len(histories) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO members_history (member_id, user_id, change_type, old_values, new_values, member_version)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING history_id, changed_at
	`

	for _, history := range histories {
		batch.Queue(query,
			history.MemberID, history.UserID, history.ChangeType,
			history.OldValues, history.NewValues, history.MemberVersion,
		)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	// Scan the results back into the history objects
	for i, history := range histories {
		err := results.QueryRow().Scan(&history.HistoryID, &history.ChangedAt)
		if err != nil {
			return domain.NewDatabaseError(err)
		}
		_ = i // Avoid unused variable warning
	}

	return nil
}

func (r *HistoryRepository) GetByMemberID(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	query := `
		SELECT h.history_id, h.member_id, h.user_id, h.changed_at, h.change_type,
		       h.old_values, h.new_values, h.member_version, u.full_name, u.email,
		       COALESCE(
			       (SELECT jsonb_object_agg(mn.language_code, mn.name)
			        FROM member_names mn
			        WHERE mn.member_id = h.member_id),
			       '{}'::jsonb
		       ) as member_names
		FROM members_history h
		JOIN users u ON h.user_id = u.user_id
		WHERE h.member_id = $1
		  AND (($2::timestamp IS NULL) OR h.changed_at < $2)
		ORDER BY h.changed_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, memberID, cursor, limit)
	if err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var histories []*domain.HistoryWithUser
	for rows.Next() {
		h := &domain.HistoryWithUser{}
		err := rows.Scan(
			&h.HistoryID, &h.MemberID, &h.UserID, &h.ChangedAt, &h.ChangeType,
			&h.OldValues, &h.NewValues, &h.MemberVersion, &h.UserFullName, &h.UserEmail,
			&h.MemberNames,
		)
		if err != nil {
			return nil, nil, domain.NewDatabaseError(err)
		}
		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}

	var nextCursor *string
	if len(histories) == limit {
		lastChangedAt := histories[len(histories)-1].ChangedAt.Format(time.RFC3339Nano)
		nextCursor = &lastChangedAt
	}

	return histories, nextCursor, nil
}

func (r *HistoryRepository) GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error) {
	query := `
		SELECT h.history_id, h.member_id, h.user_id, h.changed_at, h.change_type,
		       h.old_values, h.new_values, h.member_version, u.full_name, u.email,
		       COALESCE(
			       (SELECT jsonb_object_agg(mn.language_code, mn.name)
			        FROM member_names mn
			        WHERE mn.member_id = h.member_id),
			       '{}'::jsonb
		       ) as member_names
		FROM members_history h
		JOIN users u ON h.user_id = u.user_id
		WHERE h.user_id = $1
		  AND (($2::timestamp IS NULL) OR h.changed_at < $2)
		ORDER BY h.changed_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, userID, cursor, limit)
	if err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var histories []*domain.HistoryWithUser
	for rows.Next() {
		h := &domain.HistoryWithUser{}
		err := rows.Scan(
			&h.HistoryID, &h.MemberID, &h.UserID, &h.ChangedAt, &h.ChangeType,
			&h.OldValues, &h.NewValues, &h.MemberVersion, &h.UserFullName, &h.UserEmail,
			&h.MemberNames,
		)
		if err != nil {
			return nil, nil, domain.NewDatabaseError(err)
		}
		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}

	var nextCursor *string
	if len(histories) == limit && limit > 0 {
		lastChangedAt := histories[len(histories)-1].ChangedAt.Format(time.RFC3339Nano)
		nextCursor = &lastChangedAt
	}

	return histories, nextCursor, nil
}
