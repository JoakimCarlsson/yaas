package postgres

import (
	"context"
	"database/sql"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
)

type actionRepositoryPostgres struct {
	db *sql.DB
}

func NewActionRepository(db *sql.DB) repository.ActionRepository {
	return &actionRepositoryPostgres{db: db}
}

func (r *actionRepositoryPostgres) CreateAction(ctx context.Context, action *models.Action) error {
	query := `
		INSERT INTO actions (name, type, code, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query, action.Name, action.Type, action.Code, action.IsActive).
		Scan(&action.ID, &action.CreatedAt, &action.UpdatedAt)
}

func (r *actionRepositoryPostgres) GetActionByID(ctx context.Context, id int) (*models.Action, error) {
	query := `
		SELECT id, name, type, code, is_active, created_at, updated_at
		FROM actions WHERE id = $1
	`
	action := &models.Action{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&action.ID, &action.Name, &action.Type, &action.Code, &action.IsActive, &action.CreatedAt, &action.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return action, nil
}

func (r *actionRepositoryPostgres) GetActionsByType(ctx context.Context, actionType string) ([]*models.Action, error) {
	query := `
		SELECT id, name, type, code, is_active, created_at, updated_at
		FROM actions WHERE type = $1 AND is_active = true
	`
	rows, err := r.db.QueryContext(ctx, query, actionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []*models.Action
	for rows.Next() {
		action := &models.Action{}
		if err := rows.Scan(&action.ID, &action.Name, &action.Type, &action.Code, &action.IsActive, &action.CreatedAt, &action.UpdatedAt); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func (r *actionRepositoryPostgres) UpdateAction(ctx context.Context, action *models.Action) error {
	query := `
		UPDATE actions
		SET name = $1, type = $2, code = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, action.Name, action.Type, action.Code, action.IsActive, action.ID)
	return err
}

func (r *actionRepositoryPostgres) DeleteAction(ctx context.Context, id int) error {
	query := `DELETE FROM actions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *actionRepositoryPostgres) GetAllActions(ctx context.Context) ([]*models.Action, error) {
	query := `
		SELECT id, name, type, code, is_active, created_at, updated_at
		FROM actions
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []*models.Action
	for rows.Next() {
		action := &models.Action{}
		if err := rows.Scan(&action.ID, &action.Name, &action.Type, &action.Code, &action.IsActive, &action.CreatedAt, &action.UpdatedAt); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}
