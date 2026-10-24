package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"barber-one/backend/internal/dto"
	"barber-one/backend/internal/repository"
)

// BookingService handles booking business logic
type BookingService struct {
	repo *repository.Repository
}

func NewBookingService(repo *repository.Repository) *BookingService {
	return &BookingService{repo: repo}
}

// CreateBooking handles creating a new booking
func (s *BookingService) CreateBooking(ctx context.Context, req dto.CreateBookingRequest, userID int64) (*dto.CreateBookingResponse, error) {
	// Validate input
	if req.ServiceID < 1 || req.BarberID < 1 || req.CustomerName == "" || req.CustomerPhone == "" {
		return nil, errors.New("service_id, barber_id, customer_name, dan customer_phone wajib diisi")
	}
	if len(req.CustomerName) > 100 || len(req.CustomerPhone) > 20 {
		return nil, errors.New("nama maksimal 100 karakter dan nomor telepon maksimal 20 karakter")
	}

	// Parse date and time
	if _, err := time.Parse("2006-01-02", req.BookingDate); err != nil {
		return nil, errors.New("format tanggal harus YYYY-MM-DD")
	}
	if _, err := time.Parse("15:04", req.BookingTime); err != nil {
		return nil, errors.New("format jam harus HH:MM")
	}

	// Get service details
	service, err := s.repo.GetServiceByID(ctx, req.ServiceID)
	if err != nil {
		return nil, errors.New("layanan tidak ditemukan")
	}

	// Get barber details
	barber, err := s.repo.GetBarberByID(ctx, req.BarberID)
	if err != nil {
		return nil, errors.New("barber tidak ditemukan")
	}

	// Check if barber is available
	if !barber["available"].(bool) {
		return nil, errors.New("barber sedang tidak tersedia")
	}

	// Check slot availability
	duration := int(service["duration_minutes"].(int64))
	hasConflict, err := s.repo.CheckSlotAvailability(ctx, req.BarberID, req.BookingDate, req.BookingTime, duration)
	if err != nil {
		return nil, errors.New("gagal memeriksa slot")
	}
	if hasConflict {
		return nil, errors.New("slot ini sudah dipesan. Silakan pilih jam lain")
	}

	// Check if time is within business hours
	bookingStartTime, _ := time.Parse("15:04", req.BookingTime)
	bookingEndTime := bookingStartTime.Add(time.Duration(duration) * time.Minute)
	closeTime, _ := time.Parse("15:04", "21:00")
	if bookingEndTime.After(closeTime) {
		return nil, errors.New("durasi layanan melewati jam operasional")
	}

	// Generate ticket code
	ticketCode := fmt.Sprintf("BK-%d", time.Now().UnixNano())

	// Create booking
	notes := req.Notes
	if notes == "" {
		notes = ""
	}
	bookingID, err := s.repo.CreateBooking(ctx, ticketCode, &userID, req.CustomerName, req.CustomerPhone, req.ServiceID, req.BarberID, req.BookingDate, req.BookingTime, &notes)
	if err != nil {
		return nil, errors.New("gagal membuat booking")
	}

	// Create transaction
	amount := service["price"].(float64)
	orderID := fmt.Sprintf("%s-%d", ticketCode, time.Now().Unix())
	_ = orderID // For future use
	paymentStartedAt := time.Now()
	expiresAt := paymentStartedAt.Add(5 * time.Minute)

	gatewayOrderID := orderID
	gatewayToken := "dummy_token"
	gatewayRedirectURL := "http://localhost:3000/api/payments/midtrans/notification"

	_, err = s.repo.CreateTransaction(ctx, bookingID, amount, "Midtrans", gatewayOrderID, gatewayToken, gatewayRedirectURL, expiresAt)
	if err != nil {
		return nil, errors.New("gagal membuat transaksi")
	}

	return &dto.CreateBookingResponse{
		Message: "Booking dibuat. Selesaikan pembayaran untuk mengonfirmasi jadwal.",
		Booking: dto.CreatedBooking{
			ID:                 bookingID,
			TicketCode:         ticketCode,
			Status:             "Menunggu",
			PaymentStatus:      "Menunggu",
			PaymentRedirectURL: gatewayRedirectURL,
		},
	}, nil
}

// ListUserBookings returns all bookings for a user
func (s *BookingService) ListUserBookings(ctx context.Context, userID int64) ([]dto.UserBooking, error) {
	bookings, err := s.repo.GetUserBookings(ctx, userID)
	if err != nil {
		return nil, errors.New("gagal mengambil daftar booking")
	}

	result := make([]dto.UserBooking, 0, len(bookings))
	for _, b := range bookings {
		result = append(result, dto.UserBooking{
			ID:              b["id"].(int64),
			TicketCode:      b["ticket_code"].(string),
			Status:          b["status"].(string),
			CustomerName:    b["customer_name"].(string),
			CustomerPhone:   b["customer_phone"].(string),
			ServiceID:       b["service_id"].(int64),
			ServiceName:     b["service_name"].(string),
			Price:           b["price"].(float64),
			DurationMinutes: int(b["duration_minutes"].(int64)),
			BarberID:        b["barber_id"].(int64),
			BarberName:      b["barber_name"].(string),
			BookingDate:     b["booking_date"].(string),
			BookingTime:     b["booking_time"].(string),
			Notes:           b["notes"].(string),
			TotalAmount:     b["total_amount"].(float64),
			PaymentMethod:   b["payment_method"].(*string),
			PaymentStatus:   b["payment_status"].(*string),
			PaymentRedirectURL: b["payment_redirect_url"].(*string),
			PaymentExpiresAt: b["expires_at"].(*time.Time),
		})
	}
	return result, nil
}

// GetBookingDetail returns a specific booking by ID
func (s *BookingService) GetBookingDetail(ctx context.Context, bookingID, userID int64) (*dto.UserBooking, error) {
	booking, err := s.repo.GetBookingByID(ctx, bookingID, userID)
	if err != nil {
		return nil, errors.New("gagal mengambil booking")
	}

	return &dto.UserBooking{
		ID:              booking["id"].(int64),
		TicketCode:      booking["ticket_code"].(string),
		Status:          booking["status"].(string),
		CustomerName:    booking["customer_name"].(string),
		CustomerPhone:   booking["customer_phone"].(string),
		ServiceID:       booking["service_id"].(int64),
		ServiceName:     booking["service_name"].(string),
		Price:           booking["price"].(float64),
		DurationMinutes: int(booking["duration_minutes"].(int64)),
		BarberID:        booking["barber_id"].(int64),
		BarberName:      booking["barber_name"].(string),
		BookingDate:     booking["booking_date"].(string),
		BookingTime:     booking["booking_time"].(string),
		ChairNumber:     int(booking["chair_number"].(int64)),
		Notes:           booking["notes"].(string),
		TotalAmount:     booking["total_amount"].(float64),
		PaymentMethod:   booking["payment_method"].(*string),
		PaymentStatus:   booking["payment_status"].(*string),
		PaymentRedirectURL: booking["payment_redirect_url"].(*string),
		PaymentExpiresAt: booking["expires_at"].(*time.Time),
	}, nil
}

// CancelBooking handles booking cancellation
func (s *BookingService) CancelBooking(ctx context.Context, bookingID, userID int64) error {
	booking, err := s.repo.GetBookingByID(ctx, bookingID, userID)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}

	status := booking["status"].(string)
	if status == "Batal" || status == "Selesai" || status == "Sedang Dilayani" {
		return errors.New("status booking tidak dapat diubah")
	}

	_, err = s.repo.UpdateBookingStatus(ctx, bookingID, "Batal")
	if err != nil {
		return errors.New("gagal membatalkan booking")
	}

	return nil
}

// RescheduleBooking handles rescheduling a booking
func (s *BookingService) RescheduleBooking(ctx context.Context, req dto.RescheduleBookingRequest, bookingID, userID int64) error {
	booking, err := s.repo.GetBookingByID(ctx, bookingID, userID)
	if err != nil {
		return errors.New("booking tidak ditemukan")
	}

	status := booking["status"].(string)
	if status == "Batal" || status == "Selesai" || status == "Sedang Dilayani" {
		return errors.New("status booking tidak dapat diubah")
	}

	// Parse new date and time
	if _, err := time.Parse("2006-01-02", req.BookingDate); err != nil {
		return errors.New("format tanggal harus YYYY-MM-DD")
	}
	if _, err := time.Parse("15:04", req.BookingTime); err != nil {
		return errors.New("format jam harus HH:MM")
	}

	duration := int(booking["duration_minutes"].(int64))
	hasConflict, err := s.repo.CheckSlotAvailability(ctx, booking["barber_id"].(int64), req.BookingDate, req.BookingTime, duration)
	if err != nil {
		return errors.New("gagal memeriksa slot")
	}
	if hasConflict {
		return errors.New("slot ini sudah dipesan. Silakan pilih jam lain")
	}

	_, err = s.repo.UpdateBooking(ctx, bookingID, req.BookingDate, req.BookingTime)
	if err != nil {
		return errors.New("gagal mengubah jadwal")
	}

	return nil
}

// CheckAvailability checks if a slot is available
func (s *BookingService) CheckAvailability(ctx context.Context, barberID, serviceID int64, bookingDate, bookingTime string) ([]string, error) {
	service, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, errors.New("layanan tidak ditemukan")
	}

	duration := int(service["duration_minutes"].(int64))
	hasConflict, err := s.repo.CheckSlotAvailability(ctx, barberID, bookingDate, bookingTime, duration)
	if err != nil {
		return nil, errors.New("gagal memeriksa slot")
	}

	if hasConflict {
		return []string{}, nil
	}

	// Return available times (mock implementation)
	return []string{"10:00", "10:45", "11:30"}, nil
}

// CreatePayment creates a new payment transaction
func (s *BookingService) CreatePayment(ctx context.Context, bookingID, userID int64) (string, error) {
	// Get booking details
	booking, err := s.repo.GetBookingByID(ctx, bookingID, userID)
	if err != nil {
		return "", errors.New("booking tidak ditemukan")
	}

	paymentStatus := booking["payment_status"].(*string)
	if paymentStatus != nil && *paymentStatus == "Lunas" {
		return "", errors.New("booking ini sudah dibayar")
	}

	if paymentStatus != nil && *paymentStatus == "Menunggu" {
		redirectURL := booking["payment_redirect_url"].(*string)
		if redirectURL != nil && *redirectURL != "" {
			return *redirectURL, nil
		}
		return "", errors.New("status pembayaran sedang diperbarui")
	}

	// Create new payment (mock)
	return "http://localhost:3000/payment/redirect", nil
}
