package repository

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/artnikel/marketplace/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepo struct {
  DB *pgxpool.Pool
}

func NewItemRepo(db *pgxpool.Pool) *ItemRepo {
  return &ItemRepo{DB: db}
}

func (r *ItemRepo) Create(ctx context.Context, item *models.Item) error {
  q := `
    INSERT INTO items (title, description, image_url, price, author_id, author_login, created_at)
    VALUES ($1,$2,$3,$4,$5,$6,$7)
    RETURNING id, created_at
  `
  return r.DB.QueryRow(ctx, q,
    item.Title, item.Description, item.ImageURL, item.Price,
    item.AuthorID, item.AuthorLogin, time.Now(),
  ).Scan(&item.ID, &item.CreatedAt)
}

func (r *ItemRepo) List(ctx context.Context, offset, limit int, filters *models.ItemFilters) ([]*models.Item, error) {
	args := []interface{}{limit, offset}
	argIndex := 3 // Начинаем с 3, так как limit=$1, offset=$2

	q := `
		SELECT id, title, description, image_url, price, author_id, author_login, created_at
		FROM items
	`

	var conditions []string

	if filters.MinPrice > 0 {
		conditions = append(conditions, "price >= "+pgxPlaceholder(argIndex))
		args = append(args, filters.MinPrice)
		argIndex++
	}

	if filters.MaxPrice > 0 {
		conditions = append(conditions, "price <= "+pgxPlaceholder(argIndex))
		args = append(args, filters.MaxPrice)
		argIndex++
	}

	if filters.Title != "" {
		conditions = append(conditions, "title ILIKE "+pgxPlaceholder(argIndex))
		args = append(args, "%"+filters.Title+"%")
		argIndex++
	}

	if filters.Description != "" {
		conditions = append(conditions, "description ILIKE "+pgxPlaceholder(argIndex))
		args = append(args, "%"+filters.Description+"%")
		argIndex++
	}

	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}

	q += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item := &models.Item{}
		if err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.ImageURL,
			&item.Price, &item.AuthorID, &item.AuthorLogin, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func pgxPlaceholder(n int) string {
  return "$" + strconv.Itoa(n)
}
