// Package handlers contains HTTP handlers for item management
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/artnikel/marketplace/internal/logging"
	"github.com/artnikel/marketplace/internal/middleware"
	"github.com/artnikel/marketplace/internal/models"
)

// ItemsService is an interface that contains items service methods
type ItemsService interface {
	CreateItem(ctx context.Context, input *models.Item) (*models.Item, error)
	ListItems(ctx context.Context, page, limit int, filters *models.ItemFilters) ([]*models.Item, error)
}

// ItemsHandler handles item-related HTTP requests
type ItemsHandler struct {
	Svc    ItemsService
	logger *logging.Logger
}

// NewItemsHandler creates a new ItemsHandler instance
func NewItemsHandler(svc ItemsService, logger *logging.Logger) *ItemsHandler {
	return &ItemsHandler{Svc: svc, logger: logger}
}

// CreateItem handles POST /items — creates a new item
func (h *ItemsHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		ImageURL    string  `json:"image_url"`
		Price       float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error.Println("invalid request body:", err)
		http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r)
	userLogin := middleware.GetUserLogin(r)

	if userID == 0 || userLogin == "" {
		http.Error(w, `{"error":"user not authenticated"}`, http.StatusUnauthorized)
		return
	}

	item := &models.Item{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		AuthorID:    userID,
		AuthorLogin: userLogin,
	}

	out, err := h.Svc.CreateItem(r.Context(), item)
	if err != nil {
		h.logger.Error.Println("error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetItems handles GET /items — lists items with optional filters and pagination
func (h *ItemsHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	minPrice, _ := strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)

	titleFilter := strings.TrimSpace(r.URL.Query().Get("title"))
	descriptionFilter := strings.TrimSpace(r.URL.Query().Get("description"))

	var currentUserID int
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		currentUserID = middleware.GetUserID(r)
	}

	filters := &models.ItemFilters{
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		Title:       titleFilter,
		Description: descriptionFilter,
	}

	items, err := h.Svc.ListItems(r.Context(), page, limit, filters)
	if err != nil {
		h.logger.Error.Println("error:", err)
		http.Error(w, `{"error":"failed to list items"}`, http.StatusInternalServerError)
		return
	}

	response := make([]map[string]interface{}, len(items))
	for i, item := range items {
		response[i] = map[string]interface{}{
			"id":           item.ID,
			"title":        item.Title,
			"description":  item.Description,
			"image_url":    item.ImageURL,
			"price":        item.Price,
			"author_id":    item.AuthorID,
			"author_login": item.AuthorLogin,
			"created_at":   item.CreatedAt,
			"is_mine":      currentUserID > 0 && item.AuthorID == currentUserID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
