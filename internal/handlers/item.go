package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"

  "github.com/artnikel/marketplace/internal/service"
  "github.com/artnikel/marketplace/internal/models"
)

type ItemsHandler struct {
  Svc *service.ItemsService
}

func NewAdsHandler(svc *service.ItemsService) *ItemsHandler {
  return &ItemsHandler{Svc: svc}
}

func (h *ItemsHandler) CreateAd(w http.ResponseWriter, r *http.Request) {
  var req struct {
    Title       string  `json:"title"`
    Description string  `json:"description"`
    ImageURL    string  `json:"image_url"`
    Price       float64 `json:"price"`
  }
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
    return
  }
  uid, _ := strconv.Atoi(r.Header.Get("User-ID"))
  login := r.Header.Get("User-Login")

  ad := &models.Item{
    Title:       req.Title,
    Description: req.Description,
    ImageURL:    req.ImageURL,
    Price:       req.Price,
    AuthorID:    uid,
    AuthorLogin: login,
  }
  out, err := h.Svc.CreateAd(r.Context(), ad)
  if err != nil {
    http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
    return
  }
  json.NewEncoder(w).Encode(out)
}

func (h *ItemsHandler) GetAds(w http.ResponseWriter, r *http.Request) {
  page, _ := strconv.Atoi(r.URL.Query().Get("page"))
  if page < 1 { page = 1 }

  limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
  if limit < 1 { limit = 10 }

  minPrice, _ := strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
  maxPrice, _ := strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)

  ads, err := h.Svc.ListAds(r.Context(), page, limit, minPrice, maxPrice)
  if err != nil {
    http.Error(w, `{"error":"failed to list ads"}`, http.StatusInternalServerError)
    return
  }
  json.NewEncoder(w).Encode(ads)
}
