package repository

import (
	"context"

	"github.com/joakimcarlsson/yaas/internal/models"
)

type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, userID string, roleID int) error
	RemoveRoleFromUser(ctx context.Context, userID string, roleID int) error
	GetRolesByUserID(ctx context.Context, userID string) ([]*models.Role, error)
}
