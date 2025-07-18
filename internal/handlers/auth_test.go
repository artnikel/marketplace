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

	"github.com/artnikel/marketplace/internal/logging"
	"github.com/artnikel/marketplace/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, login, password string) (*models.User, string, error) {
	args := m.Called(ctx, login, password)
	user := args.Get(0)
	if user == nil {
		return nil, "", args.Error(2)
	}
	return user.(*models.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) Login(ctx context.Context, login, password string) (*models.User, string, error) {
	args := m.Called(ctx, login, password)
	user := args.Get(0)
	if user == nil {
		return nil, "", args.Error(2)
	}
	return user.(*models.User), args.String(1), args.Error(2)
}

func TestAuthHandler_Register(t *testing.T) {
	mockLogger := log.New(io.Discard, "", 0)
	logger := &logging.Logger{
		Error: mockLogger,
	}

	tests := []struct {
		name           string
		body           string
		setupMock      func(m *MockAuthService)
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "successful registration",
			body: `{"login":"testuser","password":"password123"}`,
			setupMock: func(m *MockAuthService) {
				m.On("Register", mock.Anything, "testuser", "password123").
					Return(&models.User{ID: 1, Login: "testuser"}, "token123", nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `"token":"token123"`,
		},
		{
			name:           "invalid JSON",
			body:           `{invalid json}`,
			setupMock:      func(_ *MockAuthService) {},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `invalid request format`,
		},
		{
			name:           "empty login",
			body:           `{"login":"", "password":"password"}`,
			setupMock:      func(_ *MockAuthService) {},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `login and password are required`,
		},
		{
			name: "service error",
			body: `{"login":"testuser","password":"password123"}`,
			setupMock: func(m *MockAuthService) {
				m.On("Register", mock.Anything, "testuser", "password123").
					Return(nil, "", errors.New("user already exists"))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `user already exists`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := new(MockAuthService)
			tt.setupMock(mockAuth)

			handler := &AuthHandler{
				AuthService: mockAuth,
				logger:      logger,
			}
			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			bodyBytes := new(bytes.Buffer)
			bodyBytes.ReadFrom(resp.Body)
			bodyStr := bodyBytes.String()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.Contains(t, bodyStr, tt.wantBody)

			mockAuth.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	mockLogger := log.New(io.Discard, "", 0)
	logger := &logging.Logger{
		Error: mockLogger,
	}

	tests := []struct {
		name           string
		body           string
		setupMock      func(m *MockAuthService)
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "successful login",
			body: `{"login":"testuser","password":"password123"}`,
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "testuser", "password123").
					Return(&models.User{ID: 1, Login: "testuser"}, "token123", nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `"token":"token123"`,
		},
		{
			name:           "invalid JSON",
			body:           `{invalid json}`,
			setupMock:      func(_ *MockAuthService) {},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `invalid request format`,
		},
		{
			name:           "empty password",
			body:           `{"login":"testuser","password":""}`,
			setupMock:      func(_ *MockAuthService) {},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `login and password are required`,
		},
		{
			name: "invalid credentials",
			body: `{"login":"testuser","password":"wrongpass"}`,
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "testuser", "wrongpass").
					Return(nil, "", errors.New("invalid credentials"))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `invalid credentials`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := new(MockAuthService)
			tt.setupMock(mockAuth)

			handler := NewAuthHandler(mockAuth, logger)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			bodyBytes := new(bytes.Buffer)
			bodyBytes.ReadFrom(resp.Body)
			bodyStr := bodyBytes.String()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.Contains(t, bodyStr, tt.wantBody)

			mockAuth.AssertExpectations(t)
		})
	}
}
