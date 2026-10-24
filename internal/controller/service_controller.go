package controller

import (
	"net/http"

	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// ServiceController handles service requests
type ServiceController struct {
	serviceService *service.ServiceService
}

func NewServiceController(serviceService *service.ServiceService) *ServiceController {
	return &ServiceController{serviceService: serviceService}
}

// ListServices handles GET /services
// @Summary      Daftar layanan aktif
// @Tags         services
// @Produce      json
// @Success      200  {object}  map[string][]map[string]any
// @Failure      500  {object}  map[string]string
// @Router       /services [get]
func (c *ServiceController) ListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	services, err := c.serviceService.ListServices(r.Context())
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal mengambil daftar layanan."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string][]map[string]any{"services": services})
}
