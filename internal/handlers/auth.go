package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/artnikel/marketplace/internal/logging"
	"github.com/artnikel/marketplace/internal/service"
)

// AuthHandler handles authentication-related endpoints like login and register
type AuthHandler struct {
	AuthService *service.AuthService
	logger      *logging.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(s *service.AuthService, logger *logging.Logger) *AuthHandler {
	return &AuthHandler{AuthService: s, logger: logger}
}

// Register handles POST /auth/register — user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error.Println("invalid request body:", err)
		http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	req.Password = strings.TrimSpace(req.Password)

	if req.Login == "" || req.Password == "" {
		http.Error(w, `{"error":"login and password are required"}`, http.StatusBadRequest)
		return
	}

	user, token, err := h.AuthService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Error.Println("error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// Login handles POST /auth/login — user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error.Println("invalid request format:", err)
		http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	req.Password = strings.TrimSpace(req.Password)

	if req.Login == "" || req.Password == "" {
		http.Error(w, `{"error":"login and password are required"}`, http.StatusBadRequest)
		return
	}

	user, token, err := h.AuthService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Error.Println("error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user,
		"token": token,
	})
}
