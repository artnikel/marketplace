// Package models provides the data models used in the application
package models

import "time"

// User entity
type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Hash  string `json:"-"`
}

// Item entity
type Item struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price"`
	AuthorID    int       `json:"author_id"`
	AuthorLogin string    `json:"author_login"`
	CreatedAt   time.Time `json:"created_at"`
}

// ItemFilters for filtering items by fields
type ItemFilters struct {
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
}
