package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/artnikel/marketplace/internal/service"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	userLoginKey contextKey = "userLogin"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(authService *service.AuthService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"authorization token required"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := authService.ParseToken(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userLoginKey, claims.Login)

			next(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) int {
	if id, ok := r.Context().Value(userIDKey).(int); ok {
		return id
	}
	return 0
}

func GetUserLogin(r *http.Request) string {
	if login, ok := r.Context().Value(userLoginKey).(string); ok {
		return login
	}
	return ""
}
