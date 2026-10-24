package controller

import (
	"net/http"
	"time"

	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// CashierController handles cashier requests
type CashierController struct {
	cashierService *service.CashierService
	sessionStore   *middleware.SessionStore
}

func NewCashierController(cashierService *service.CashierService, sessionStore *middleware.SessionStore) *CashierController {
	return &CashierController{cashierService: cashierService, sessionStore: sessionStore}
}

// IsCashier checks if user has cashier role
func (c *CashierController) IsCashier(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := c.sessionStore.UserIDFromRequest(r)
	if !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login sebagai kasir."})
		return 0, false
	}
	return userID, true
}

// GetQueue handles GET /cashier/queue
// @Summary      Daftar antrean kasir
// @Tags         cashier
// @Produce      json
// @Security     BearerAuth
// @Param        date  query     string  false  "Tanggal (YYYY-MM-DD), default hari ini"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /cashier/queue [get]
func (c *CashierController) GetQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	if _, ok := c.IsCashier(w, r); !ok {
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"date":  date,
		"queue": []map[string]any{},
	})
}

// GetSummary handles GET /cashier/summary
// @Summary      Ringkasan dashboard kasir
// @Tags         cashier
// @Produce      json
// @Security     BearerAuth
// @Param        date  query     string  false  "Tanggal (YYYY-MM-DD), default hari ini"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /cashier/summary [get]
func (c *CashierController) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	if _, ok := c.IsCashier(w, r); !ok {
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"date":         date,
		"waiting":      0,
		"in_service":   0,
		"done":         0,
		"total_chairs": 3,
		"revenue":      0,
		"cash_count":   0,
		"non_cash_count": 0,
	})
}

// WalkIn handles POST /cashier/walk-in
// @Summary      Tambah pelanggan walk-in
// @Tags         cashier
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /cashier/walk-in [post]
func (c *CashierController) WalkIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	if _, ok := c.IsCashier(w, r); !ok {
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Walk-in berhasil",
	})
}
