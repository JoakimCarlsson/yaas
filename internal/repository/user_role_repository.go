package repository

import "context"

type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, userID string, roleID int) error
	RemoveRoleFromUser(ctx context.Context, userID string, roleID int) error
	GetRolesByUserID(ctx context.Context, userID string) ([]*Role, error)
}
