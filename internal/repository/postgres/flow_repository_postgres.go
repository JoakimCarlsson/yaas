package postgres

import (
	"context"
	"database/sql"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
	"time"
)

type flowRepositoryPostgres struct {
	db *sql.DB
}

func NewFlowRepository(db *sql.DB) repository.FlowRepository {
	return &flowRepositoryPostgres{db: db}
}

func (r *flowRepositoryPostgres) CreateFlow(ctx context.Context, flow *models.Flow) error {
	query := `
        INSERT INTO flows (id, type, state, expires_at, issued_at, request_url, errors)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.ExecContext(ctx, query, flow.ID, flow.Type, flow.State, flow.ExpiresAt, flow.IssuedAt, flow.RequestURL, nil)
	return err
}

func (r *flowRepositoryPostgres) GetFlowByID(ctx context.Context, id string) (*models.Flow, error) {
	query := `
        SELECT id, type, state, expires_at, issued_at, request_url, errors
        FROM flows WHERE id = $1
    `
	flow := &models.Flow{}
	var errorsData []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&flow.ID,
		&flow.Type,
		&flow.State,
		&flow.ExpiresAt,
		&flow.IssuedAt,
		&flow.RequestURL,
		&errorsData,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Optionally unmarshal errorsData into flow.Errors
	return flow, nil
}

func (r *flowRepositoryPostgres) UpdateFlow(ctx context.Context, flow *models.Flow) error {
	query := `
        UPDATE flows SET state = $1, errors = $2, request_url = $3 WHERE id = $4
    `
	// Optionally marshal flow.Errors into a JSON or other format
	_, err := r.db.ExecContext(ctx, query, flow.State, nil, flow.RequestURL, flow.ID)
	return err
}

func (r *flowRepositoryPostgres) DeleteFlow(ctx context.Context, id string) error {
	query := `DELETE FROM flows WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *flowRepositoryPostgres) CleanupExpiredFlows(ctx context.Context) error {
	query := `DELETE FROM flows WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
