package service

import (
	"context"
	"errors"
	"strings"

	"github.com/artnikel/marketplace/internal/models"
	"github.com/artnikel/marketplace/internal/repository"
)

type ItemsService struct {
	ItemRepo *repository.ItemRepo
	UserRepo *repository.UserRepo
}

func NewItemsService(itemRepo *repository.ItemRepo, userRepo *repository.UserRepo) *ItemsService {
	return &ItemsService{ItemRepo: itemRepo, UserRepo: userRepo}
}

func (s *ItemsService) CreateItem(ctx context.Context, input *models.Item) (*models.Item, error) {
	if input.Title == "" || input.Description == "" || input.Price <= 0 {
		return nil, errors.New("title, description and positive price are required")
	}
	if err := s.ItemRepo.Create(ctx, input); err != nil {
		return nil, err
	}
	return input, nil
}

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
