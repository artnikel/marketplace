package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artnikel/marketplace/internal/handlers"
	"github.com/artnikel/marketplace/internal/middleware"
	"github.com/artnikel/marketplace/internal/repository"
	"github.com/artnikel/marketplace/internal/service"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://user:password@postgres:5432/marketplacedb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Database connection established")

	userRepo := repository.NewUserRepo(pool)
	itemRepo := repository.NewItemRepo(pool)

	authSvc := service.NewAuthService(userRepo)
	itemsSvc := service.NewItemsService(itemRepo, userRepo)

	authH := handlers.NewAuthHandler(authSvc)
	itemsH := handlers.NewItemsHandler(itemsSvc)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.LoggingMiddleware)

	r.HandleFunc("/auth/register", authH.Register).Methods("POST")
	r.HandleFunc("/auth/login", authH.Login).Methods("POST")
	r.HandleFunc("/items", itemsH.GetItems).Methods("GET")

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(authSvc))
	protected.HandleFunc("/items", itemsH.CreateItem).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("Server running at :8080")
	log.Fatal(srv.ListenAndServe())
}
