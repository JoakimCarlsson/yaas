package models

import "time"

type User struct {
	ID         string
	Email      string
	Password   string
	IsActive   bool
	IsVerified bool
	Provider   string
	ProviderID *string
	LastLogin  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
