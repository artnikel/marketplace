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

	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Printf("failed to write health response: %v", err)
		}
	}).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Public routes
	api.HandleFunc("/auth/register", authH.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", authH.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/items", itemsH.GetItems).Methods("GET", "OPTIONS")

	// Protected routes
	api.Handle("/items", middleware.AuthMiddleware(authSvc)(http.HandlerFunc(itemsH.CreateItem))).Methods("POST", "OPTIONS")

	// Fallback for old API paths (без /api prefix)
	r.HandleFunc("/auth/register", authH.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/login", authH.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/items", itemsH.GetItems).Methods("GET", "OPTIONS")
	r.Handle("/items", middleware.AuthMiddleware(authSvc)(http.HandlerFunc(itemsH.CreateItem))).Methods("POST", "OPTIONS")

	// Serve frontend
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("web"))))

	port := cfg.Server.Port
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + strconv.Itoa(port),
		ReadTimeout:  constants.ServerTimeout,
		WriteTimeout: constants.ServerTimeout,
	}

	log.Printf("Server running on 0.0.0.0:%d", port)
	log.Fatal(srv.ListenAndServe())
}
