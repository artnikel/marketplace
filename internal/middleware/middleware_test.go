package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artnikel/marketplace/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

type mockAuthService struct{}

func (m *mockAuthService) ParseToken(token string) (*jwt.Claims, error) {
	if token == "valid-token" {
		return &jwt.Claims{UserID: 42, Login: "user42"}, nil
	}
	return nil, errors.New("invalid token")
}

func TestCORSMiddleware(t *testing.T) {
	nextCalled := false
	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		nextCalled = true
	})

	handler := CORSMiddleware(nextHandler)

	t.Run("OPTIONS request", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("OPTIONS", "/", http.NoBody)
		req.Header.Set("Origin", "http://example.com")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.False(t, nextCalled)

		assert.Equal(t, "http://example.com", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
	})

	t.Run("GET request", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("GET", "/", http.NoBody)
		req.Header.Set("Origin", "http://example.org")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, nextCalled)

		assert.Equal(t, "http://example.org", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
	})
}

func TestLoggingMiddleware(t *testing.T) {
	nextCalled := false
	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		nextCalled = true
	})
	handler := LoggingMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.True(t, nextCalled)
}

func TestAuthMiddleware(t *testing.T) {
	mockSvc := &mockAuthService{}

	nextCalled := false
	var userIDInCtx int
	var userLoginInCtx string
	var userIDHeader, userLoginHeader string

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		userIDInCtx = GetUserID(r)
		userLoginInCtx = GetUserLogin(r)
		userIDHeader = r.Header.Get("User-ID")
		userLoginHeader = r.Header.Get("User-Login")
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(mockSvc)(nextHandler)

	req := httptest.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	req = httptest.NewRequest("GET", "/", http.NoBody)
	req.Header.Set("Authorization", "invalidtoken")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	req = httptest.NewRequest("GET", "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer badtoken")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	req = httptest.NewRequest("GET", "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-token")
	w = httptest.NewRecorder()
	nextCalled = false

	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "42", userIDHeader)
	assert.Equal(t, "user42", userLoginHeader)
	assert.Equal(t, 42, userIDInCtx)
	assert.Equal(t, "user42", userLoginInCtx)
}

func TestGetUserIDAndLogin(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey, 100)
	ctx = context.WithValue(ctx, UserLoginKey, "testuser")

	req := httptest.NewRequest("GET", "/", http.NoBody).WithContext(ctx)

	assert.Equal(t, 100, GetUserID(req))
	assert.Equal(t, "testuser", GetUserLogin(req))

	req = httptest.NewRequest("GET", "/", http.NoBody)
	assert.Equal(t, 0, GetUserID(req))
	assert.Equal(t, "", GetUserLogin(req))
}
