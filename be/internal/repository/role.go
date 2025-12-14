package repository

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetByID(ctx context.Context, roleID int) (*domain.Role, error) {
	query := `SELECT role_id, name FROM roles WHERE role_id = $1`
	role := &domain.Role{}
	err := r.db.QueryRow(ctx, query, roleID).Scan(&role.RoleID, &role.Name)
	return role, err
}

func (r *RoleRepository) GetAll(ctx context.Context) ([]*domain.Role, error) {
	query := `SELECT role_id, name FROM roles ORDER BY role_id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.RoleID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}


