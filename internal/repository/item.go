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

func (r *ItemRepo) List(ctx context.Context, offset, limit int, minPrice, maxPrice float64) ([]*models.Item, error) {
  args := []interface{}{limit, offset}
  q := `
    SELECT id, title, description, image_url, price, author_id, author_login, created_at
    FROM items
  `
  filters := []string{}
  if minPrice > 0 {
    filters = append(filters, `price >= `+pgxPlaceholder(len(args)+1))
    args = append(args, minPrice)
  }
  if maxPrice > 0 {
    filters = append(filters, `price <= `+pgxPlaceholder(len(args)+1))
    args = append(args, maxPrice)
  }
  if len(filters) > 0 {
    q += " WHERE " + strings.Join(filters, " AND ")
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
