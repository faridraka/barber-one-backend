package controller

import (
	"net/http"

	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// TicketController handles ticket requests
type TicketController struct {
	bookingService *service.BookingService
	sessionStore   *middleware.SessionStore
}

func NewTicketController(bookingService *service.BookingService, sessionStore *middleware.SessionStore) *TicketController {
	return &TicketController{bookingService: bookingService, sessionStore: sessionStore}
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

	tickets, err := c.bookingService.GetMyTickets(r.Context(), userID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tickets)
}
