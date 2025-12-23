package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (full_name, email, avatar, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id, created_at
	`
	err := r.db.QueryRow(ctx, query, user.FullName, user.Email, user.Avatar, user.RoleID, user.IsActive).
		Scan(&user.UserID, &user.CreatedAt)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID int) (*domain.User, error) {
	query := `
		SELECT user_id, full_name, email, avatar, role_id, is_active, created_at
		FROM users
		WHERE user_id = $1
	`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID, &user.FullName, &user.Email, &user.Avatar,
		&user.RoleID, &user.IsActive, &user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("user")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT user_id, full_name, email, avatar, role_id, is_active, created_at
		FROM users
		WHERE email = $1
	`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.UserID, &user.FullName, &user.Email, &user.Avatar,
		&user.RoleID, &user.IsActive, &user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET full_name = $1, avatar = $2, role_id = $3, is_active = $4
		WHERE user_id = $5
	`
	_, err := r.db.Exec(ctx, query, user.FullName, user.Avatar, user.RoleID, user.IsActive, user.UserID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID, roleID int) error {
	query := `UPDATE users SET role_id = $1 WHERE user_id = $2`
	_, err := r.db.Exec(ctx, query, roleID, userID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) UpdateActive(ctx context.Context, userID int, isActive bool) error {
	query := `UPDATE users SET is_active = $1 WHERE user_id = $2`
	_, err := r.db.Exec(ctx, query, isActive, userID)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, filter domain.UserFilter, cursor *string, limit int) ([]*domain.User, *string, error) {
	query := `
		SELECT user_id, full_name, email, avatar, role_id, is_active, created_at
		FROM users
		WHERE (($1::text IS NULL) OR user_id > $1::int)
		  AND (($2::text IS NULL) OR (full_name ILIKE '%' || $2 || '%' OR email ILIKE '%' || $2 || '%'))
		  AND (($3::int IS NULL) OR role_id = $3)
		  AND (($4::boolean IS NULL) OR is_active = $4)
		ORDER BY user_id
		LIMIT $5
	`

	rows, err := r.db.Query(ctx, query, cursor, filter.Search, filter.RoleID, filter.IsActive, limit)
	if err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.UserID, &user.FullName, &user.Email, &user.Avatar,
			&user.RoleID, &user.IsActive, &user.CreatedAt,
		)
		if err != nil {
			return nil, nil, domain.NewDatabaseError(err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, domain.NewDatabaseError(err)
	}

	var nextCursor *string
	if len(users) == limit {
		lastUserID := strconv.Itoa(users[len(users)-1].UserID)
		nextCursor = &lastUserID
	}

	return users, nextCursor, nil
}

func (r *UserRepository) GetWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error) {
	query := `
		SELECT u.user_id, u.full_name, u.email, u.avatar, u.role_id, u.is_active, u.created_at,
		       COALESCE(SUM(us.points), 0) as total_score
		FROM users u
		LEFT JOIN user_scores us ON u.user_id = us.user_id
		WHERE u.user_id = $1
		GROUP BY u.user_id
	`
	user := &domain.UserWithScore{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID, &user.FullName, &user.Email, &user.Avatar,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.TotalScore,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("user")
	}
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	return user, nil
}

func (r *UserRepository) CreateRoleHistory(ctx context.Context, userID, oldRoleID, newRoleID, changedBy int, actionType string) error {
	query := `
		INSERT INTO user_role_history (user_id, old_role_id, new_role_id, changed_by, action_type)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, userID, oldRoleID, newRoleID, changedBy, actionType)
	if err != nil {
		return domain.NewDatabaseError(err)
	}
	return nil
}
