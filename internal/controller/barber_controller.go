package controller

import (
	"net/http"

	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// BarberController handles barber requests
type BarberController struct {
	barberService *service.BarberService
}

func NewBarberController(barberService *service.BarberService) *BarberController {
	return &BarberController{barberService: barberService}
}

// ListBarbers handles GET /barbers
// @Summary      Daftar barber
// @Tags         barbers
// @Produce      json
// @Success      200  {object}  map[string][]map[string]any
// @Failure      500  {object}  map[string]string
// @Router       /barbers [get]
func (c *BarberController) ListBarbers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	barbers, err := c.barberService.ListBarbers(r.Context())
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal mengambil daftar barber."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string][]map[string]any{"barbers": barbers})
}
