package repository

import (
	"context"

	"github.com/joakimcarlsson/yaas/internal/models"
)

type RoleRepository interface {
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByID(ctx context.Context, id int) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, id int) error
}
