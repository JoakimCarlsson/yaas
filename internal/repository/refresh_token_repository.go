package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joakimcarlsson/yaas/internal/models"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID string, jti uuid.UUID, expiresAt time.Time) error
	GetByJTI(ctx context.Context, jti uuid.UUID) (*models.RefreshToken, error)
	Delete(ctx context.Context, jti uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID string) error
}
