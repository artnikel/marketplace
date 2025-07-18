package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/artnikel/marketplace/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

var db *pgxpool.Pool
var userRepo *UserRepo
var itemRepo *ItemRepo
var pool *dockertest.Pool
var resource *dockertest.Resource

func TestMain(m *testing.M) {
	var err error

	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}

	resource, err = pool.Run("postgres", "15", []string{
		"POSTGRES_USER=postgres",
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=testdb",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %v", err)
	}

	err = pool.Retry(func() error {
		connStr := fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))
		db, err = pgxpool.New(context.Background(), connStr)
		if err != nil {
			return err
		}
		return db.Ping(context.Background())
	})
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	createTables()

	userRepo = NewUserRepo(db)
	itemRepo = NewItemRepo(db)

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %v", err)
	}

	os.Exit(code)
}

func createTables() {
	_, err := db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		image_url TEXT NOT NULL,
		price NUMERIC(10,2) NOT NULL,
		author_id INT NOT NULL,
		author_login TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	`)
	if err != nil {
		log.Fatalf("Could not create tables: %v", err)
	}
}

func cleanTables(t *testing.T) {
	_, err := db.Exec(context.Background(), "DELETE FROM items")
	assert.NoError(t, err)
	_, err = db.Exec(context.Background(), "DELETE FROM users")
	assert.NoError(t, err)
}

func TestUserRepo_CreateAndGetByLogin(t *testing.T) {
	cleanTables(t)

	ctx := context.Background()

	user, err := userRepo.Create(ctx, "testuser", "hashedpass")
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "hashedpass", user.Hash)

	gotUser, err := userRepo.GetByLogin(ctx, "testuser")
	assert.NoError(t, err)
	assert.NotNil(t, gotUser)
	assert.Equal(t, user.ID, gotUser.ID)
	assert.Equal(t, user.Login, gotUser.Login)
	assert.Equal(t, user.Hash, gotUser.Hash)

	noUser, err := userRepo.GetByLogin(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, noUser)
}

func TestItemRepo_CreateAndList(t *testing.T) {
	cleanTables(t)

	ctx := context.Background()

	item := &models.Item{
		Title:       "Test item",
		Description: "Desc",
		ImageURL:    "http://image.url",
		Price:       123.45,
		AuthorID:    1,
		AuthorLogin: "author1",
	}

	err := itemRepo.Create(ctx, item)
	assert.NoError(t, err)
	assert.NotZero(t, item.ID)
	assert.False(t, item.CreatedAt.IsZero())

	items, err := itemRepo.List(ctx, 0, 10, &models.ItemFilters{})
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, item.Title, items[0].Title)

	item2 := &models.Item{
		Title:       "Second item",
		Description: "Other desc",
		ImageURL:    "http://image2.url",
		Price:       200,
		AuthorID:    2,
		AuthorLogin: "author2",
	}
	err = itemRepo.Create(ctx, item2)
	assert.NoError(t, err)

	filteredItems, err := itemRepo.List(ctx, 0, 10, &models.ItemFilters{
		MinPrice: 150,
	})
	assert.NoError(t, err)
	assert.Len(t, filteredItems, 1)
	assert.Equal(t, item2.Title, filteredItems[0].Title)
}
