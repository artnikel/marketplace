package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/artnikel/marketplace/internal/models"
)

// MockItemRepo is a mock implementation of ItemRepo
type MockItemRepo struct {
	mock.Mock
}

func (m *MockItemRepo) Create(ctx context.Context, item *models.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepo) List(ctx context.Context, offset, limit int, filters *models.ItemFilters) ([]*models.Item, error) {
	args := m.Called(ctx, offset, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

func TestItemsService_CreateItem(t *testing.T) {
	tests := []struct {
		name      string
		input     *models.Item
		setupMock func(*MockItemRepo)
		wantErr   bool
		wantMsg   string
	}{
		{
			name: "successful item creation",
			input: &models.Item{
				Title:       "Test Item",
				Description: "Test Description",
				Price:       99.99,
				AuthorID:    1,
				AuthorLogin: "testuser",
			},
			setupMock: func(m *MockItemRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Item")).
					Return(nil).
					Run(func(args mock.Arguments) {
						item := args.Get(1).(*models.Item)
						item.ID = 1
						item.CreatedAt = time.Now()
					})
			},
			wantErr: false,
		},
		{
			name: "empty title",
			input: &models.Item{
				Title:       "",
				Description: "Test Description",
				Price:       99.99,
				AuthorID:    1,
				AuthorLogin: "testuser",
			},
			setupMock: func(_ *MockItemRepo) {},
			wantErr:   true,
			wantMsg:   "title, description and positive price are required",
		},
		{
			name: "empty description",
			input: &models.Item{
				Title:       "Test Item",
				Description: "",
				Price:       99.99,
				AuthorID:    1,
				AuthorLogin: "testuser",
			},
			setupMock: func(_ *MockItemRepo) {},
			wantErr:   true,
			wantMsg:   "title, description and positive price are required",
		},
		{
			name: "zero price",
			input: &models.Item{
				Title:       "Test Item",
				Description: "Test Description",
				Price:       0,
				AuthorID:    1,
				AuthorLogin: "testuser",
			},
			setupMock: func(_ *MockItemRepo) {},
			wantErr:   true,
			wantMsg:   "title, description and positive price are required",
		},
		{
			name: "negative price",
			input: &models.Item{
				Title:       "Test Item",
				Description: "Test Description",
				Price:       -10.50,
				AuthorID:    1,
				AuthorLogin: "testuser",
			},
			setupMock: func(_ *MockItemRepo) {},
			wantErr:   true,
			wantMsg:   "title, description and positive price are required",
		},
		{
			name: "database error",
			input: &models.Item{
				Title:       "Test Item",
				Description: "Test Description",
				Price:       99.99,
				AuthorID:    1,
				AuthorLogin: "testuser",
			},
			setupMock: func(m *MockItemRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Item")).
					Return(errors.New("database connection failed"))
			},
			wantErr: true,
			wantMsg: "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockItemRepo := new(MockItemRepo)
			mockUserRepo := new(MockUserRepo)
			tt.setupMock(mockItemRepo)

			service := NewItemsService(mockItemRepo, mockUserRepo)
			result, err := service.CreateItem(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Title, result.Title)
				assert.Equal(t, tt.input.Description, result.Description)
				assert.Equal(t, tt.input.Price, result.Price)
			}

			mockItemRepo.AssertExpectations(t)
		})
	}
}

func TestItemsService_ListItems(t *testing.T) {
	mockItems := []*models.Item{
		{
			ID:          1,
			Title:       "Item 1",
			Description: "Description 1",
			Price:       100.00,
			AuthorID:    1,
			AuthorLogin: "user1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			Title:       "Item 2",
			Description: "Description 2",
			Price:       200.00,
			AuthorID:    2,
			AuthorLogin: "user2",
			CreatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name        string
		page        int
		limit       int
		filters     *models.ItemFilters
		setupMock   func(*MockItemRepo)
		wantErr     bool
		wantMsg     string
		wantItems   int
		wantOffset  int
		wantLimit   int
		wantFilters *models.ItemFilters
	}{
		{
			name:  "successful list with default pagination",
			page:  1,
			limit: 10,
			filters: &models.ItemFilters{
				MinPrice: 0,
				MaxPrice: 0,
			},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.AnythingOfType("*models.ItemFilters")).
					Return(mockItems, nil)
			},
			wantErr:     false,
			wantItems:   2,
			wantOffset:  0,
			wantLimit:   10,
			wantFilters: &models.ItemFilters{MinPrice: 0, MaxPrice: 0},
		},
		{
			name:  "page 2 with limit 5",
			page:  2,
			limit: 5,
			filters: &models.ItemFilters{
				MinPrice: 50,
				MaxPrice: 150,
			},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 5, 5, mock.AnythingOfType("*models.ItemFilters")).
					Return(mockItems[:1], nil)
			},
			wantErr:     false,
			wantItems:   1,
			wantOffset:  5,
			wantLimit:   5,
			wantFilters: &models.ItemFilters{MinPrice: 50, MaxPrice: 150},
		},
		{
			name:    "invalid page number (0)",
			page:    0,
			limit:   10,
			filters: &models.ItemFilters{},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.AnythingOfType("*models.ItemFilters")).
					Return(mockItems, nil)
			},
			wantErr:    false,
			wantItems:  2,
			wantOffset: 0,
			wantLimit:  10,
		},
		{
			name:    "invalid page number (negative)",
			page:    -1,
			limit:   10,
			filters: &models.ItemFilters{},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.AnythingOfType("*models.ItemFilters")).
					Return(mockItems, nil)
			},
			wantErr:    false,
			wantItems:  2,
			wantOffset: 0,
			wantLimit:  10,
		},
		{
			name:    "invalid limit (0)",
			page:    1,
			limit:   0,
			filters: &models.ItemFilters{},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.AnythingOfType("*models.ItemFilters")).
					Return(mockItems, nil)
			},
			wantErr:    false,
			wantItems:  2,
			wantOffset: 0,
			wantLimit:  10,
		},
		{
			name:    "invalid limit (over 100)",
			page:    1,
			limit:   150,
			filters: &models.ItemFilters{},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.AnythingOfType("*models.ItemFilters")).
					Return(mockItems, nil)
			},
			wantErr:    false,
			wantItems:  2,
			wantOffset: 0,
			wantLimit:  10,
		},
		{
			name:  "min price greater than max price",
			page:  1,
			limit: 10,
			filters: &models.ItemFilters{
				MinPrice: 200,
				MaxPrice: 100,
			},
			setupMock: func(_ *MockItemRepo) {},
			wantErr:   true,
			wantMsg:   "min_price cannot be greater than max_price",
		},
		{
			name:    "database error",
			page:    1,
			limit:   10,
			filters: &models.ItemFilters{},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.AnythingOfType("*models.ItemFilters")).
					Return(nil, errors.New("database connection failed"))
			},
			wantErr: true,
			wantMsg: "database connection failed",
		},
		{
			name:  "with title filter",
			page:  1,
			limit: 10,
			filters: &models.ItemFilters{
				Title: "  Test Item  ",
			},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.MatchedBy(func(f *models.ItemFilters) bool {
					return f.Title == "Test Item"
				})).Return(mockItems, nil)
			},
			wantErr:   false,
			wantItems: 2,
		},
		{
			name:  "with description filter",
			page:  1,
			limit: 10,
			filters: &models.ItemFilters{
				Description: "  Test Description  ",
			},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.MatchedBy(func(f *models.ItemFilters) bool {
					return f.Description == "Test Description"
				})).Return(mockItems, nil)
			},
			wantErr:   false,
			wantItems: 2,
		},
		{
			name:  "negative prices are normalized",
			page:  1,
			limit: 10,
			filters: &models.ItemFilters{
				MinPrice: -50,
				MaxPrice: -10,
			},
			setupMock: func(m *MockItemRepo) {
				m.On("List", mock.Anything, 0, 10, mock.MatchedBy(func(f *models.ItemFilters) bool {
					return f.MinPrice == 0 && f.MaxPrice == 0
				})).Return(mockItems, nil)
			},
			wantErr:   false,
			wantItems: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepo)
			tt.setupMock(mockRepo)

			service := &ItemsService{
				ItemRepo: mockRepo,
			}

			items, err := service.ListItems(context.Background(), tt.page, tt.limit, tt.filters)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMsg)
				assert.Nil(t, items)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, items)
				assert.Len(t, items, tt.wantItems)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
