package models

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
