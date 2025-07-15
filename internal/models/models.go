package models

import "time"

type User struct {
	ID    int
	Login string
	Hash  string
}

type Item struct {
	ID          int
	Title       string
	Description string
	ImageURL    string
	Price       float64
	AuthorID    int
	AuthorLogin string
	CreatedAt   time.Time
}
