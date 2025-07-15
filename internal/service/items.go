package service

import (
  "context"
  "errors"
  "github.com/artnikel/marketplace/internal/models"
  "github.com/artnikel/marketplace/internal/repository"
)

type ItemsService struct {
  ItemRepo   *repository.ItemRepo
  UserRepo *repository.UserRepo
}

func NewItemsService(itemRepo *repository.ItemRepo, userRepo *repository.UserRepo) *ItemsService {
  return &ItemsService{ItemRepo: itemRepo, UserRepo: userRepo}
}

func (s *ItemsService) CreateAd(ctx context.Context, input *models.Item) (*models.Item, error) {
  if input.Title == "" || input.Description == "" || input.Price <= 0 {
    return nil, errors.New("title, description and positive price are required")
  }
  if err := s.ItemRepo.Create(ctx, input); err != nil {
    return nil, err
  }
  return input, nil
}

func (s *ItemsService) ListAds(ctx context.Context, page, limit int, minPrice, maxPrice float64) ([]*models.Item, error) {
  offset := (page - 1) * limit
  return s.ItemRepo.List(ctx, offset, limit, minPrice, maxPrice)
}
