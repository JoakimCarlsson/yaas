package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
)

type roleRepositoryPostgres struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) repository.RoleRepository {
	return &roleRepositoryPostgres{db: db}
}

func (r *roleRepositoryPostgres) CreateRole(ctx context.Context, role *models.Role) error {
	query := `
        INSERT INTO roles (
            name, description, created_at, updated_at
        ) VALUES (
            $1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        ) RETURNING id
    `
	err := r.db.QueryRowContext(ctx, query, role.Name, role.Description).Scan(&role.ID)
	return err
}

func (r *roleRepositoryPostgres) GetRoleByID(ctx context.Context, id int) (*models.Role, error) {
	query := `
        SELECT id, name, description, created_at, updated_at
        FROM roles WHERE id = $1
    `
	role := &models.Role{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("role not found")
	}
	return role, err
}

func (r *roleRepositoryPostgres) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	query := `
        SELECT id, name, description, created_at, updated_at
        FROM roles WHERE name = $1
    `
	role := &models.Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("role not found")
	}
	return role, err
}

func (r *roleRepositoryPostgres) UpdateRole(ctx context.Context, role *models.Role) error {
	query := `
        UPDATE roles SET
            name = $1,
            description = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
    `
	_, err := r.db.ExecContext(ctx, query, role.Name, role.Description, role.ID)
	return err
}

func (r *roleRepositoryPostgres) DeleteRole(ctx context.Context, id int) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
