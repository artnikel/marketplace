package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/artnikel/marketplace/internal/logging"
	"github.com/artnikel/marketplace/internal/middleware"
	"github.com/artnikel/marketplace/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockItemsService struct {
	mock.Mock
}

func (m *MockItemsService) CreateItem(ctx context.Context, input *models.Item) (*models.Item, error) {
	args := m.Called(ctx, input)
	item, _ := args.Get(0).(*models.Item)
	return item, args.Error(1)
}

func (m *MockItemsService) ListItems(ctx context.Context, page, limit int, filters *models.ItemFilters) ([]*models.Item, error) {
	args := m.Called(ctx, page, limit, filters)
	items, _ := args.Get(0).([]*models.Item)
	return items, args.Error(1)
}

func setUserContext(r *http.Request, id int, login string) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, id))
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserLoginKey, login))
	return r
}

func TestItemsHandler_CreateItem(t *testing.T) {
	mockLogger := log.New(io.Discard, "", 0)
	logger := &logging.Logger{
		Error: mockLogger,
	}
	mockSvc := new(MockItemsService)
	handler := NewItemsHandler(mockSvc, logger)

	validItem := &models.Item{
		ID:          1,
		Title:       "Test Title",
		Description: "Test Desc",
		ImageURL:    "http://example.com/image.jpg",
		Price:       10.5,
		AuthorID:    123,
		AuthorLogin: "user123",
	}

	tests := []struct {
		name           string
		body           string
		userID         int
		userLogin      string
		setupMock      func()
		wantStatusCode int
		wantContains   string
	}{
		{
			name:      "successful create item",
			body:      `{"title":"Test Title","description":"Test Desc","image_url":"http://example.com/image.jpg","price":10.5}`,
			userID:    123,
			userLogin: "user123",
			setupMock: func() {
				mockSvc.On("CreateItem", mock.Anything, mock.MatchedBy(func(item *models.Item) bool {
					return item.Title == "Test Title" && item.AuthorID == 123
				})).Return(validItem, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantContains:   `"title":"Test Title"`,
		},
		{
			name:           "invalid json body",
			body:           `{"title":"Test Title",`,
			userID:         123,
			userLogin:      "user123",
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantContains:   "invalid request format",
		},
		{
			name:           "unauthenticated user",
			body:           `{"title":"Test Title","description":"Test Desc","image_url":"http://example.com/image.jpg","price":10.5}`,
			userID:         0,
			userLogin:      "",
			setupMock:      func() {},
			wantStatusCode: http.StatusUnauthorized,
			wantContains:   "user not authenticated",
		},
		{
			name:      "service error on create",
			body:      `{"title":"Test Title","description":"Test Desc","image_url":"http://example.com/image.jpg","price":10.5}`,
			userID:    123,
			userLogin: "user123",
			setupMock: func() {
				mockSvc.On("CreateItem", mock.Anything, mock.Anything).
					Return(nil, errors.New("some error")).Once()
			},
			wantStatusCode: http.StatusBadRequest,
			wantContains:   "some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc.ExpectedCalls = nil
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(tt.body))
			if tt.userID != 0 {
				req = setUserContext(req, tt.userID, tt.userLogin)
			}
			w := httptest.NewRecorder()

			handler.CreateItem(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body := new(bytes.Buffer)
			_, _ = body.ReadFrom(resp.Body)

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.Contains(t, body.String(), tt.wantContains)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestItemsHandler_GetItems(t *testing.T) {
	mockLogger := log.New(io.Discard, "", 0)
	logger := &logging.Logger{
		Error: mockLogger,
	}
	mockSvc := new(MockItemsService)
	handler := NewItemsHandler(mockSvc, logger)

	mockItems := []*models.Item{
		{
			ID:          1,
			Title:       "Item 1",
			Description: "Desc 1",
			ImageURL:    "http://example.com/1.jpg",
			Price:       100,
			AuthorID:    123,
			AuthorLogin: "user1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			Title:       "Item 2",
			Description: "Desc 2",
			ImageURL:    "http://example.com/2.jpg",
			Price:       200,
			AuthorID:    456,
			AuthorLogin: "user2",
			CreatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name           string
		query          string
		authHeader     string
		userID         int
		userLogin      string
		setupMock      func()
		wantStatusCode int
		wantContains   []string
	}{
		{
			name:       "successful get items",
			query:      "?page=1&limit=2",
			authHeader: "Bearer token",
			userID:     123,
			userLogin:  "user123",
			setupMock: func() {
				mockSvc.On("ListItems", mock.Anything, 1, 2, mock.Anything).
					Return(mockItems, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantContains:   []string{`"title":"Item 1"`, `"title":"Item 2"`, `"is_mine":true`, `"is_mine":false`},
		},
		{
			name:       "service error",
			query:      "?page=1&limit=10",
			authHeader: "",
			setupMock: func() {
				mockSvc.On("ListItems", mock.Anything, 1, 10, mock.Anything).
					Return(nil, errors.New("db error")).Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantContains:   []string{"failed to list items"},
		},
		{
			name:       "no auth header",
			query:      "?page=1&limit=10",
			authHeader: "",
			setupMock: func() {
				mockSvc.On("ListItems", mock.Anything, 1, 10, mock.Anything).
					Return(mockItems, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantContains:   []string{`"is_mine":false`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc.ExpectedCalls = nil
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/items"+tt.query, http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.userID != 0 {
				req = setUserContext(req, tt.userID, tt.userLogin)
			}

			w := httptest.NewRecorder()
			handler.GetItems(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body := new(bytes.Buffer)
			_, _ = body.ReadFrom(resp.Body)

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			for _, substr := range tt.wantContains {
				assert.Contains(t, body.String(), substr)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
