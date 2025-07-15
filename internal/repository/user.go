package repository

import (
	"context"
	"errors"

	"github.com/artnikel/marketplace/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	DB *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{DB: db}
}

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
