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
	pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:5432/adsdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepo(pool)
	adRepo := repository.NewItemRepo(pool)

	authSvc := service.NewAuthService(userRepo)
	adsSvc := service.NewItemsService(adRepo, userRepo)

	authH := handlers.NewAuthHandler(authSvc)
	adsH := handlers.NewAdsHandler(adsSvc)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)

	r.HandleFunc("/auth/register", authH.Register).Methods("POST")
	r.HandleFunc("/auth/login", authH.Login).Methods("POST")

	r.HandleFunc("/items", middleware.AuthMiddleware(authSvc)(adsH.CreateAd)).Methods("POST")
	r.HandleFunc("/items", adsH.GetAds).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server running at :8080")
	log.Fatal(srv.ListenAndServe())
}
