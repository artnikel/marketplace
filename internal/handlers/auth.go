package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artnikel/marketplace/internal/service"
)

type AuthHandler struct {
	AuthService *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: s}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
		return
	}

	token, err := h.AuthService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
		return
	}

	token, err := h.AuthService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
