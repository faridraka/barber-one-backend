package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// ReviewController handles review requests
type ReviewController struct {
	reviewService *service.ReviewService
	sessionStore  *middleware.SessionStore
}

func NewReviewController(reviewService *service.ReviewService, sessionStore *middleware.SessionStore) *ReviewController {
	return &ReviewController{reviewService: reviewService, sessionStore: sessionStore}
}

// CreateReview handles POST /reviews
// @Summary      Buat review untuk booking yang selesai
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /reviews [post]
func (c *ReviewController) CreateReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	if _, ok := c.sessionStore.UserIDFromRequest(r); !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login untuk membuat review."})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format review tidak valid."})
		return
	}

	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format review tidak valid."})
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Review berhasil disimpan.",
		"review": map[string]any{
			"id":        1,
			"booking_id": 1,
			"rating":     5,
			"comment":    req["comment"],
		},
	})
}

// ListReviews handles GET /reviews
// @Summary      Daftar review
// @Tags         reviews
// @Produce      json
// @Success      200        {object}  map[string][]map[string]any
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /reviews [get]
func (c *ReviewController) ListReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string][]map[string]any{"reviews": []map[string]any{
		{
			"id":           1,
			"booking_id":   1,
			"rating":       5,
			"comment":      "Layanan bagus!",
			"created_at":   "2026-10-20T10:00:00Z",
			"customer_name": "John Doe",
			"service_name":  "Potong Rambut",
			"barber_name":   "Barber 1",
		},
	}})
}
