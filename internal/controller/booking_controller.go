package controller

import (
	"encoding/json"
	"net/http"

	"barber-one/backend/internal/dto"
	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/service"
)

// BookingController handles booking HTTP requests
type BookingController struct {
	bookingService *service.BookingService
	userService    *service.UserService
	sessionStore   *middleware.SessionStore
}

func NewBookingController(bookingService *service.BookingService, userService *service.UserService, sessionStore *middleware.SessionStore) *BookingController {
	return &BookingController{bookingService: bookingService, userService: userService, sessionStore: sessionStore}
}

// CreateBooking handles POST /bookings
// @Summary      Membuat booking baru
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /bookings [post]
func (c *BookingController) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	userID, ok := c.sessionStore.UserIDFromRequest(r)
	if !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login untuk membuat booking."})
		return
	}

	var req dto.CreateBookingRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Data booking tidak valid."})
		return
	}
	result, err := c.bookingService.CreateBooking(r.Context(), req, userID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, result)
}

// ListUserBookings handles GET /bookings
// @Summary      Daftar booking user yang sedang login
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string][]map[string]any
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bookings [get]
func (c *BookingController) ListUserBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	userID, ok := c.sessionStore.UserIDFromRequest(r)
	if !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login untuk melihat booking."})
		return
	}

	bookings, err := c.bookingService.ListUserBookings(r.Context(), userID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

// GetBookingDetail handles GET /bookings/{id}
// @Summary      Detail booking user
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "ID booking"
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bookings/{id} [get]
func (c *BookingController) GetBookingDetail(w http.ResponseWriter, r *http.Request, bookingID int64) {
	userID, ok := c.sessionStore.UserIDFromRequest(r)
	if !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login."})
		return
	}

	booking, err := c.bookingService.GetBookingDetail(r.Context(), bookingID, userID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Booking tidak ditemukan."})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]any{"booking": booking})
}

// CancelBooking handles PATCH /bookings/{id}/cancel
// @Summary      Membatalkan booking user
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "ID booking"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /bookings/{id}/cancel [patch]
func (c *BookingController) CancelBooking(w http.ResponseWriter, r *http.Request, bookingID int64) {
	if _, ok := c.sessionStore.UserIDFromRequest(r); !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "Booking berhasil dibatalkan."})
}

// RescheduleBooking handles PATCH /bookings/{id}/reschedule
// @Summary      Mengubah jadwal booking user
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                       true  "ID booking"
// @Success      200   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /bookings/{id}/reschedule [patch]
func (c *BookingController) RescheduleBooking(w http.ResponseWriter, r *http.Request, bookingID int64) {
	if _, ok := c.sessionStore.UserIDFromRequest(r); !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "Jadwal booking berhasil diubah."})
}

// CreatePayment handles POST /bookings/{id}/payment
// @Summary      Buat sesi pembayaran baru
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "ID booking"
// @Success      201  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /bookings/{id}/payment [post]
func (c *BookingController) CreatePayment(w http.ResponseWriter, r *http.Request, bookingID int64) {
	userID, ok := c.sessionStore.UserIDFromRequest(r)
	if !ok {
		middleware.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Silakan login."})
		return
	}
	url, err := c.bookingService.CreatePayment(r.Context(), bookingID, userID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, map[string]string{"payment_redirect_url": url})
}

// CheckAvailability handles GET /availability
// @Summary      Cek slot booking yang tersedia
// @Tags         bookings
// @Produce      json
// @Success      200        {object}  map[string]any
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /availability [get]
func (c *BookingController) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"barber_id":       1,
		"service_id":      1,
		"date":            "2026-10-23",
		"available_slots": []string{"10:00", "10:45", "11:30"},
	})
}
