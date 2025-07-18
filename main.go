// Package main is an entry point to application
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artnikel/marketplace/internal/config"
	"github.com/artnikel/marketplace/internal/constants"
	"github.com/artnikel/marketplace/internal/handlers"
	"github.com/artnikel/marketplace/internal/logging"
	"github.com/artnikel/marketplace/internal/middleware"
	"github.com/artnikel/marketplace/internal/repository"
	"github.com/artnikel/marketplace/internal/service"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.Database.Connection = dbURL
	}

	logger, err := logging.NewLogger(cfg.Logging.Path)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Database.Connection)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	//nolint:gocritic
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping database:", err)
	}
	log.Println("database connection established")

	userRepo := repository.NewUserRepo(pool)
	itemRepo := repository.NewItemRepo(pool)

	authSvc := service.NewAuthService(userRepo, cfg)
	itemsSvc := service.NewItemsService(itemRepo, userRepo)

	authH := handlers.NewAuthHandler(authSvc, logger)
	itemsH := handlers.NewItemsHandler(itemsSvc, logger)

	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.LoggingMiddleware)

	// Public routes
	r.HandleFunc("/auth/register", authH.Register).Methods("POST")
	r.HandleFunc("/auth/login", authH.Login).Methods("POST")
	r.HandleFunc("/items", itemsH.GetItems).Methods("GET")

	// Protected POST route
	r.Handle("/items", middleware.AuthMiddleware(authSvc)(http.HandlerFunc(itemsH.CreateItem))).Methods("POST")

	// Serve frontend
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("web"))))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		ReadTimeout:  constants.ServerTimeout,
		WriteTimeout: constants.ServerTimeout,
	}

	log.Println("Server running at :" + strconv.Itoa(cfg.Server.Port))
	log.Fatal(srv.ListenAndServe())
}
