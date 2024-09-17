package repository

import "context"

type User struct {
	ID         string
	Email      string
	Password   string
	FirstName  string
	LastName   string
	IsActive   bool
	IsVerified bool
	Provider   string
	ProviderID *string
	LastLogin  *string
	CreatedAt  string
	UpdatedAt  string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
}
