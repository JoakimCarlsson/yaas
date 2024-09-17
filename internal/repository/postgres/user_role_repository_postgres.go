package postgres

import (
	"context"
	"database/sql"

	"github.com/joakimcarlsson/yaas/internal/repository"
)

type userRoleRepositoryPostgres struct {
	db *sql.DB
}

func NewUserRoleRepository(db *sql.DB) repository.UserRoleRepository {
	return &userRoleRepositoryPostgres{db: db}
}

func (r *userRoleRepositoryPostgres) AssignRoleToUser(ctx context.Context, userID string, roleID int) error {
	query := `
        INSERT INTO user_roles (user_id, role_id, created_at)
        VALUES ($1, $2, CURRENT_TIMESTAMP)
        ON CONFLICT DO NOTHING
    `
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *userRoleRepositoryPostgres) RemoveRoleFromUser(ctx context.Context, userID string, roleID int) error {
	query := `
        DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *userRoleRepositoryPostgres) GetRolesByUserID(ctx context.Context, userID string) ([]*repository.Role, error) {
	query := `
        SELECT r.id, r.name, r.description, r.created_at, r.updated_at
        FROM roles r
        INNER JOIN user_roles ur ON ur.role_id = r.id
        WHERE ur.user_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := []*repository.Role{}
	for rows.Next() {
		role := &repository.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
