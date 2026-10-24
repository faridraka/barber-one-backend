package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// PaymentController handles payment notifications
type PaymentController struct {
	midtransService *service.MidtransService
}

func NewPaymentController(midtransService *service.MidtransService) *PaymentController {
	return &PaymentController{midtransService: midtransService}
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

	var notification map[string]any
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&notification); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Notifikasi tidak valid."})
		return
	}

	orderID := notification["order_id"].(string)
	statusCode := notification["status_code"].(string)
	grossAmount := notification["gross_amount"].(string)
	signatureKey := notification["signature_key"].(string)

	if !c.midtransService.VerifySignature(orderID, statusCode, grossAmount, signatureKey) {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Signature notifikasi tidak valid."})
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
