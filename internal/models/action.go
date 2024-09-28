package models

import "time"

type Action struct {
	ID        int
	Name      string
	Type      string
	Code      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
