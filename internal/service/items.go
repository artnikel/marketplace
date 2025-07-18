// Package service contains business logic for handling items
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/artnikel/marketplace/internal/models"
)

// ItemRepository is an interface that contains item repository methods
type ItemRepository interface {
	Create(ctx context.Context, item *models.Item) error
	List(ctx context.Context, offset, limit int, filters *models.ItemFilters) ([]*models.Item, error)
}

// ItemsService provides methods for managing items
type ItemsService struct {
	ItemRepo ItemRepository
	UserRepo UserRepository
}

// NewItemsService creates a new instance of ItemsService
func NewItemsService(itemRepo ItemRepository, userRepo UserRepository) *ItemsService {
	return &ItemsService{ItemRepo: itemRepo, UserRepo: userRepo}
}

// CreateItem validates and creates a new item
func (s *ItemsService) CreateItem(ctx context.Context, input *models.Item) (*models.Item, error) {
	if input.Title == "" || input.Description == "" || input.Price <= 0 {
		return nil, errors.New("title, description and positive price are required")
	}
	if err := s.ItemRepo.Create(ctx, input); err != nil {
		return nil, err
	}
	return input, nil
}

// ListItems returns a paginated list of items based on filters
func (s *ItemsService) ListItems(ctx context.Context, page, limit int, filters *models.ItemFilters) ([]*models.Item, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	if filters == nil {
		filters = &models.ItemFilters{}
	}

	filters.Title = strings.TrimSpace(filters.Title)
	filters.Description = strings.TrimSpace(filters.Description)

	if filters.MinPrice < 0 {
		filters.MinPrice = 0
	}
	if filters.MaxPrice < 0 {
		filters.MaxPrice = 0
	}
	if filters.MaxPrice > 0 && filters.MinPrice > filters.MaxPrice {
		return nil, errors.New("min_price cannot be greater than max_price")
	}

	return s.ItemRepo.List(ctx, offset, limit, filters)
}
