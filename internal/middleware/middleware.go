// Package middleware provides HTTP middleware functions for logging, CORS, and authentication
package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/artnikel/marketplace/internal/service"
)

// contextKey is a custom type used for storing values in context
type contextKey string

// Keys used to store user information in context
const (
	UserIDKey    contextKey = "userID"
	UserLoginKey contextKey = "userLogin"
)

// CORSMiddleware adds CORS headers to the response
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, User-Agent, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
// LoggingMiddleware logs each incoming HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT token and injects user info into the request context
func AuthMiddleware(authService service.AuthServiceInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"authorization token required"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := authService.ParseToken(token)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserLoginKey, claims.Login)

			r.Header.Set("User-ID", strconv.Itoa(claims.UserID))
			r.Header.Set("User-Login", claims.Login)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the request context
func GetUserID(r *http.Request) int {
	if id, ok := r.Context().Value(UserIDKey).(int); ok {
		return id
	}
	return 0
}

// GetUserLogin extracts the user login from the request context
func GetUserLogin(r *http.Request) string {
	if login, ok := r.Context().Value(UserLoginKey).(string); ok {
		return login
	}
	return ""
}
