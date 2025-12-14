package repository

import (
	"context"
	"fmt"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScoreRepository struct {
	db *pgxpool.Pool
}

func NewScoreRepository(db *pgxpool.Pool) *ScoreRepository {
	return &ScoreRepository{db: db}
}

func (r *ScoreRepository) Create(ctx context.Context, score *domain.Score) error {
	query := `
		INSERT INTO user_scores (user_id, member_id, field_name, points, member_version)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`
	return r.db.QueryRow(ctx, query,
		score.UserID, score.MemberID, score.FieldName, score.Points, score.MemberVersion,
	).Scan(&score.CreatedAt)
}

func (r *ScoreRepository) GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error) {
	query := `
		SELECT us.user_id, us.member_id, us.field_name, us.points, us.member_version, us.created_at,
		       m.arabic_name, m.english_name
		FROM user_scores us
		JOIN members m ON us.member_id = m.member_id
		WHERE us.user_id = $1
	`
	var args []interface{}
	args = append(args, userID)
	argCount := 2

	// Apply cursor-based pagination (using created_at as cursor)
	if cursor != nil && *cursor != "" {
		query += fmt.Sprintf(" AND us.created_at < $%d", argCount)
		args = append(args, *cursor)
		argCount++
	}

	query += " ORDER BY us.created_at DESC"

	// Fetch one extra to determine if there's a next page
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var scores []*domain.ScoreHistory
	for rows.Next() {
		s := &domain.ScoreHistory{}
		err := rows.Scan(
			&s.UserID, &s.MemberID, &s.FieldName, &s.Points, &s.MemberVersion, &s.CreatedAt,
			&s.MemberArabicName, &s.MemberEnglishName,
		)
		if err != nil {
			return nil, nil, err
		}
		scores = append(scores, s)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Determine next cursor
	var nextCursor *string
	if len(scores) > limit {
		// Remove the extra score and set cursor
		scores = scores[:limit]
		lastCreatedAt := scores[len(scores)-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
		nextCursor = &lastCreatedAt
	}

	return scores, nextCursor, nil
}

func (r *ScoreRepository) GetLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error) {
	query := `
		SELECT u.user_id, u.full_name, u.avatar, COALESCE(SUM(us.points), 0) as total_score,
		       RANK() OVER (ORDER BY COALESCE(SUM(us.points), 0) DESC) as rank
		FROM users u
		LEFT JOIN user_scores us ON u.user_id = us.user_id
		WHERE u.is_active = true
		GROUP BY u.user_id, u.full_name, u.avatar
		ORDER BY total_score DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []*domain.UserScore
	for rows.Next() {
		s := &domain.UserScore{}
		err := rows.Scan(&s.UserID, &s.FullName, &s.Avatar, &s.TotalScore, &s.Rank)
		if err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, s)
	}
	return leaderboard, rows.Err()
}

func (r *ScoreRepository) GetTotalByUserID(ctx context.Context, userID int) (int, error) {
	query := `SELECT COALESCE(SUM(points), 0) FROM user_scores WHERE user_id = $1`
	var total int
	err := r.db.QueryRow(ctx, query, userID).Scan(&total)
	return total, err
}

// DeleteByMemberAndField removes scores for a specific member field (used when updating)
func (r *ScoreRepository) DeleteByMemberAndField(ctx context.Context, memberID int, fieldName string, memberVersion int) error {
	query := `DELETE FROM user_scores WHERE member_id = $1 AND field_name = $2 AND member_version = $3`
	_, err := r.db.Exec(ctx, query, memberID, fieldName, memberVersion)
	return err
}
