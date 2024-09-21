package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
)

type refreshTokenRepositoryPostgres struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepositoryPostgres{db: db}
}

func (r *refreshTokenRepositoryPostgres) Create(ctx context.Context, userID string, jti uuid.UUID, expiresAt time.Time) error {
	query := `
        INSERT INTO refresh_tokens (user_id, jti, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.ExecContext(ctx, query, userID, jti, expiresAt)
	return err
}

func (r *refreshTokenRepositoryPostgres) GetByJTI(ctx context.Context, jti uuid.UUID) (*models.RefreshToken, error) {
	query := `
        SELECT id, user_id, jti, expires_at, created_at
        FROM refresh_tokens
        WHERE jti = $1
    `
	refreshToken := &models.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, jti).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.JTI,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return refreshToken, nil
}

func (r *refreshTokenRepositoryPostgres) Delete(ctx context.Context, jti uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE jti = $1`
	_, err := r.db.ExecContext(ctx, query, jti)
	return err
}

func (r *refreshTokenRepositoryPostgres) DeleteAllForUser(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
