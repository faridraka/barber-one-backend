package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"barber-one/backend/internal/dto"
	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// PaymentController handles payment notifications
type PaymentController struct {
	bookingService *service.BookingService
}

func NewPaymentController(bookingService *service.BookingService) *PaymentController {
	return &PaymentController{bookingService: bookingService}
}

// MidtransNotification handles POST /payments/midtrans/notification
// @Summary      Terima notifikasi pembayaran Midtrans
// @Tags         payments
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      405 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Failure      503 {object} map[string]string
// @Router       /payments/midtrans/notification [post]
func (c *PaymentController) MidtransNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	var notification dto.MidtransNotification
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&notification); err != nil || notification.OrderID == "" {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Notifikasi tidak valid."})
		return
	}

	if err := c.bookingService.HandleMidtransNotification(r.Context(), notification); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSignature):
			middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrPaymentUnavailable):
			middleware.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
		default:
			middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "Notifikasi pembayaran diproses."})
}

// Status handles GET /status and GET /
// @Summary      Status API
// @Description  Memeriksa bahwa API Barber One aktif.
// @Tags         system
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /status [get]
func (c *PaymentController) Status(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/api/status" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "Barber One API aktif"})
}
