package models

import "time"

type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Hash  string `json:"-"` 
}

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

type ItemFilters struct {
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
}
