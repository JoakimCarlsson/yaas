package models

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	JTI       string
	ExpiresAt time.Time
	CreatedAt time.Time
}
