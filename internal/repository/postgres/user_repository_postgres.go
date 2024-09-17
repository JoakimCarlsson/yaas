package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joakimcarlsson/yaas/internal/repository"
)

type userRepositoryPostgres struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepositoryPostgres{db: db}
}

func (r *userRepositoryPostgres) CreateUser(ctx context.Context, user *repository.User) error {
	query := `
        INSERT INTO users (
            email, password, first_name, last_name, is_active,
            is_verified, provider, provider_id, last_login, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        ) RETURNING id
    `
	err := r.db.QueryRowContext(ctx, query,
		user.Email, user.Password, user.FirstName, user.LastName,
		user.IsActive, user.IsVerified, user.Provider, user.ProviderID, user.LastLogin,
	).Scan(&user.ID)
	return err
}

func (r *userRepositoryPostgres) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	query := `
        SELECT id, email, password, first_name, last_name, is_active,
               is_verified, provider, provider_id, last_login, created_at, updated_at
        FROM users WHERE id = $1
    `
	user := &repository.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.IsActive, &user.IsVerified, &user.Provider, &user.ProviderID, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *userRepositoryPostgres) GetUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	query := `
        SELECT id, email, password, first_name, last_name, is_active,
               is_verified, provider, provider_id, last_login, created_at, updated_at
        FROM users WHERE email = $1
    `
	user := &repository.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.IsActive, &user.IsVerified, &user.Provider, &user.ProviderID, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *userRepositoryPostgres) UpdateUser(ctx context.Context, user *repository.User) error {
	query := `
        UPDATE users SET
            email = $1,
            password = $2,
            first_name = $3,
            last_name = $4,
            is_active = $5,
            is_verified = $6,
            provider = $7,
            provider_id = $8,
            last_login = $9,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $10
    `
	_, err := r.db.ExecContext(ctx, query,
		user.Email, user.Password, user.FirstName, user.LastName,
		user.IsActive, user.IsVerified, user.Provider, user.ProviderID, user.LastLogin, user.ID,
	)
	return err
}

func (r *userRepositoryPostgres) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
