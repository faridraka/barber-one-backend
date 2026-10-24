package controller

import (
	"net/http"

	"barber-one/backend/internal/middleware"
)

// TicketController handles ticket requests
type TicketController struct {
	sessionStore *middleware.SessionStore
}

func NewTicketController(sessionStore *middleware.SessionStore) *TicketController {
	return &TicketController{sessionStore: sessionStore}
}

// MyTickets handles GET /my-tickets
// @Summary      Tiket aktif, riwayat kunjungan, dan statistik user yang login
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Router       /my-tickets [get]
func (c *TicketController) MyTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	userID, ok := c.sessionStore.UserIDFromRequest(r)
	if !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Sesi tidak valid atau sudah kedaluwarsa. Silakan login kembali."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":    userID,
			"name":  "User Name",
			"phone": "081234567890",
			"email": nil,
			"role":  "user",
		},
		"stats": map[string]any{
			"visits":         5,
			"favorite_barber": "Barber 1",
		},
		"active": map[string]any{
			"id":           1,
			"ticket_code":  "BK-ABC123",
			"status":       "Dikonfirmasi",
			"booking_date": "2026-10-25",
			"booking_time": "10:00",
		},
		"history": []map[string]any{},
	})
}
