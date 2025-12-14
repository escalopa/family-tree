package repository

import (
	"context"
	"fmt"

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
	return r.db.QueryRow(ctx, query, user.FullName, user.Email, user.Avatar, user.RoleID, user.IsActive).
		Scan(&user.UserID, &user.CreatedAt)
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
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
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
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET full_name = $1, avatar = $2, role_id = $3, is_active = $4
		WHERE user_id = $5
	`
	_, err := r.db.Exec(ctx, query, user.FullName, user.Avatar, user.RoleID, user.IsActive, user.UserID)
	return err
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID, roleID int) error {
	query := `UPDATE users SET role_id = $1 WHERE user_id = $2`
	_, err := r.db.Exec(ctx, query, roleID, userID)
	return err
}

func (r *UserRepository) UpdateActive(ctx context.Context, userID int, isActive bool) error {
	query := `UPDATE users SET is_active = $1 WHERE user_id = $2`
	_, err := r.db.Exec(ctx, query, isActive, userID)
	return err
}

func (r *UserRepository) List(ctx context.Context, cursor *string, limit int) ([]*domain.User, *string, error) {
	query := `
		SELECT user_id, full_name, email, avatar, role_id, is_active, created_at
		FROM users
	`
	var args []interface{}
	argCount := 1

	// Apply cursor-based pagination
	if cursor != nil && *cursor != "" {
		query += fmt.Sprintf(" WHERE user_id > $%d", argCount)
		args = append(args, *cursor)
		argCount++
	}

	query += " ORDER BY user_id"

	// Fetch one extra to determine if there's a next page
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
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
			return nil, nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Determine next cursor
	var nextCursor *string
	if len(users) > limit {
		// Remove the extra user and set cursor
		users = users[:limit]
		lastUserID := fmt.Sprintf("%d", users[len(users)-1].UserID)
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
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}


