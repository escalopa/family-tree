package repository

import (
	"context"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScoreRepository struct {
	db *pgxpool.Pool
}

func NewScoreRepository(db *pgxpool.Pool) *ScoreRepository {
	return &ScoreRepository{db: db}
}

func (r *ScoreRepository) Create(ctx context.Context, scores ...domain.Score) error {
	if len(scores) == 0 {
		return nil
	}

	querier := getQuerier(ctx, r.db)
	batch := &pgx.Batch{}
	query := `
		INSERT INTO user_scores (user_id, member_id, field_name, points, member_version)
		VALUES ($1, $2, $3, $4, $5)
	`
	for i := range scores {
		batch.Queue(query,
			scores[i].UserID,
			scores[i].MemberID,
			scores[i].FieldName,
			scores[i].Points,
			scores[i].MemberVersion,
		)
	}

	br := querier.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		return domain.NewDatabaseError(err)
	}

	return nil
}

func (r *ScoreRepository) GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error) {
	query := `
		SELECT us.user_id, us.member_id, us.field_name, us.points, us.member_version, us.created_at
		FROM user_scores us
		JOIN members m ON us.member_id = m.member_id
		WHERE us.user_id = $1
		  AND (($2::timestamp IS NULL) OR us.created_at < $2)
		ORDER BY us.created_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, userID, cursor, limit)
	if err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var scores []*domain.ScoreHistory
	var memberIDs []int
	for rows.Next() {
		s := &domain.ScoreHistory{}
		err := rows.Scan(
			&s.UserID, &s.MemberID, &s.FieldName, &s.Points, &s.MemberVersion, &s.CreatedAt,
		)
		if err != nil {
			return nil, nil, domain.NewDatabaseError(err)
		}
		scores = append(scores, s)
		memberIDs = append(memberIDs, s.MemberID)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}

	if len(scores) == 0 {
		return nil, nil, nil
	}

	namesQuery := `
			SELECT member_id, language_code, name
			FROM member_names
			WHERE member_id = ANY($1)
		`
	nameRows, err := r.db.Query(ctx, namesQuery, memberIDs)
	if err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}
	defer nameRows.Close()

	namesMap := make(map[int]map[string]string)
	for nameRows.Next() {
		var (
			mid            int
			langCode, name string
		)
		if err := nameRows.Scan(&mid, &langCode, &name); err != nil {
			return nil, nil, domain.NewDatabaseError(err)
		}
		if namesMap[mid] == nil {
			namesMap[mid] = make(map[string]string)
		}
		namesMap[mid][langCode] = name
	}

	for i := range scores {
		if names, ok := namesMap[scores[i].MemberID]; ok {
			scores[i].MemberNames = names
		} else {
			scores[i].MemberNames = make(map[string]string)
		}
	}

	var nextCursor *string
	if len(scores) == limit {
		lastCreatedAt := scores[len(scores)-1].CreatedAt.Format(time.RFC3339Nano)
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
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var leaderboard []*domain.UserScore
	for rows.Next() {
		s := &domain.UserScore{}
		err := rows.Scan(&s.UserID, &s.FullName, &s.Avatar, &s.TotalScore, &s.Rank)
		if err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		leaderboard = append(leaderboard, s)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return leaderboard, nil
}

func (r *ScoreRepository) GetTotalByUserID(ctx context.Context, userID int) (int, error) {
	query := `SELECT COALESCE(SUM(points), 0) FROM user_scores WHERE user_id = $1`
	var total int
	err := r.db.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, domain.NewDatabaseError(err)
	}
	return total, nil
}

// DeleteByMemberAndField removes scores for a specific member field (used when updating)
func (r *ScoreRepository) DeleteByMemberAndField(ctx context.Context, memberID int, fieldName string, memberVersion int) error {
	query := `DELETE FROM user_scores WHERE member_id = $1 AND field_name = $2 AND member_version = $3`
	_, err := r.db.Exec(ctx, query, memberID, fieldName, memberVersion)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}
