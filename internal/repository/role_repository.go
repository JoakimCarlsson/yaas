package repository

import "context"

type Role struct {
	ID          int
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

type RoleRepository interface {
	CreateRole(ctx context.Context, role *Role) error
	GetRoleByID(ctx context.Context, id int) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id int) error
}
