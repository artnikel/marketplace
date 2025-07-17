// Package repository provides data access to the database
package repository

import (
	"context"
	"errors"

	"github.com/artnikel/marketplace/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo handles database operations related to users
type UserRepo struct {
	DB *pgxpool.Pool
}

// NewUserRepo creates a new instance of UserRepo
func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{DB: db}
}

// Create inserts a new user into the database
func (r *UserRepo) Create(ctx context.Context, login, hash string) (*models.User, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int
	err := r.DB.QueryRow(ctx, query, login, hash).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:    id,
		Login: login,
		Hash:  hash,
	}, nil
}

// GetByLogin retrieves a user by their login
func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`

	row := r.DB.QueryRow(ctx, query, login)

	var user models.User
	err := row.Scan(&user.ID, &user.Login, &user.Hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
