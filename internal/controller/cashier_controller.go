package controller

import (
	"log"
	"net/http"
	"time"

	"barber-one/backend/internal/dto"
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

	date, ok := cashierDate(w, r)
	if !ok {
		return
	}

	queue, err := c.cashierService.GetQueue(r.Context(), date)
	if err != nil {
		log.Printf("cashier queue failed: %v", err)
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal mengambil antrean."})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, dto.CashierQueueResponse{Date: date, ServerTime: time.Now(), Queue: queue})
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

	date, ok := cashierDate(w, r)
	if !ok {
		return
	}

	summary, err := c.cashierService.GetSummary(r.Context(), date)
	if err != nil {
		log.Printf("cashier summary failed: %v", err)
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal mengambil ringkasan."})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, summary)
}

// cashierDate reads ?date=YYYY-MM-DD, defaulting to today in WIB (the server may run in UTC).
func cashierDate(w http.ResponseWriter, r *http.Request) (string, bool) {
	date := r.URL.Query().Get("date")
	if date == "" {
		return time.Now().In(time.FixedZone("WIB", 7*60*60)).Format("2006-01-02"), true
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format tanggal harus YYYY-MM-DD."})
		return "", false
	}
	return date, true
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
