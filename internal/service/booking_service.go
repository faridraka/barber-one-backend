package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"barber-one/backend/internal/dto"
	"barber-one/backend/internal/repository"
)

// BookingService handles booking business logic
type BookingService struct {
	repo     *repository.Repository
	midtrans *MidtransService
}

func NewBookingService(repo *repository.Repository, midtrans *MidtransService) *BookingService {
	return &BookingService{repo: repo, midtrans: midtrans}
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

	if slotHasPassed(req.BookingDate, req.BookingTime) {
		return nil, errors.New("jam booking sudah lewat. Silakan pilih jam lain")
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
	available, _ := barber["available"].(bool)
	if !available {
		return nil, errors.New("barber sedang tidak tersedia")
	}

	// Check slot availability
	duration := asInt(service["duration_minutes"])
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

	// Generate ticket code (bookings.ticket_code is VARCHAR(20))
	ticketCode, err := newTicketCode()
	if err != nil {
		return nil, errors.New("gagal membuat kode tiket")
	}

	// Create booking
	notes := req.Notes
	bookingID, err := s.repo.CreateBooking(ctx, ticketCode, &userID, req.CustomerName, req.CustomerPhone, req.ServiceID, req.BarberID, req.BookingDate, req.BookingTime, &notes)
	if err != nil {
		log.Printf("create booking failed: %v", err)
		return nil, errors.New("gagal membuat booking")
	}

	amount, _ := service["price"].(float64)
	serviceName, _ := service["name"].(string)
	_, _, gatewayRedirectURL, _, err := s.startPayment(ctx, bookingID, ticketCode, req.CustomerName, req.CustomerPhone, amount, serviceName)
	if err != nil {
		log.Printf("start payment for booking %d failed: %v", bookingID, err)
		// Free the slot again: a booking without a payment session cannot be paid.
		if delErr := s.repo.DeleteBooking(ctx, bookingID); delErr != nil {
			log.Printf("rollback booking %d failed: %v", bookingID, delErr)
		}
		return nil, err
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

// newTicketCode returns a code like "BK-3F9A1C2E" (11 chars).
func newTicketCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "BK-" + strings.ToUpper(hex.EncodeToString(b)), nil
}

func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func (s *BookingService) startPayment(ctx context.Context, bookingID int64, ticketCode, name, phone string, amount float64, serviceName string) (string, string, string, time.Time, error) {
	if s.midtrans == nil {
		return "", "", "", time.Time{}, errors.New("pembayaran belum tersedia: konfigurasi MIDTRANS_SERVER_KEY belum diatur")
	}
	startedAt := time.Now()
	orderID := fmt.Sprintf("%s-%d", ticketCode, startedAt.UnixNano())
	result, err := s.midtrans.CreateSnap(ctx, orderID, name, phone, amount, serviceName, "", startedAt)
	if err != nil {
		return "", "", "", time.Time{}, fmt.Errorf("gagal membuat sesi pembayaran Midtrans: %w", err)
	}
	expiresAt := startedAt.Add(5 * time.Minute)
	if _, err := s.repo.CreateTransaction(ctx, bookingID, amount, "Midtrans", orderID, result["token"], result["redirect_url"], expiresAt); err != nil {
		return "", "", "", time.Time{}, errors.New("gagal menyimpan transaksi pembayaran")
	}
	return orderID, result["token"], result["redirect_url"], expiresAt, nil
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
			ID:                 b["id"].(int64),
			TicketCode:         b["ticket_code"].(string),
			Status:             b["status"].(string),
			CustomerName:       b["customer_name"].(string),
			CustomerPhone:      b["customer_phone"].(string),
			ServiceID:          b["service_id"].(int64),
			ServiceName:        b["service_name"].(string),
			Price:              b["price"].(float64),
			DurationMinutes:    asInt(b["duration_minutes"]),
			BarberID:           b["barber_id"].(int64),
			BarberName:         b["barber_name"].(string),
			BookingDate:        b["booking_date"].(string),
			BookingTime:        b["booking_time"].(string),
			Notes:              b["notes"].(string),
			TotalAmount:        b["total_amount"].(float64),
			PaymentMethod:      b["payment_method"].(*string),
			PaymentStatus:      b["payment_status"].(*string),
			PaymentRedirectURL: b["payment_redirect_url"].(*string),
			PaymentExpiresAt:   b["expires_at"].(*time.Time),
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
		ID:                 booking["id"].(int64),
		TicketCode:         booking["ticket_code"].(string),
		Status:             booking["status"].(string),
		CustomerName:       booking["customer_name"].(string),
		CustomerPhone:      booking["customer_phone"].(string),
		ServiceID:          booking["service_id"].(int64),
		ServiceName:        booking["service_name"].(string),
		Price:              booking["price"].(float64),
		DurationMinutes:    asInt(booking["duration_minutes"]),
		BarberID:           booking["barber_id"].(int64),
		BarberName:         booking["barber_name"].(string),
		BookingDate:        booking["booking_date"].(string),
		BookingTime:        booking["booking_time"].(string),
		ChairNumber:        asInt(booking["chair_number"]),
		Notes:              booking["notes"].(string),
		TotalAmount:        booking["total_amount"].(float64),
		PaymentMethod:      booking["payment_method"].(*string),
		PaymentStatus:      booking["payment_status"].(*string),
		PaymentRedirectURL: booking["payment_redirect_url"].(*string),
		PaymentExpiresAt:   booking["expires_at"].(*time.Time),
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

	duration := asInt(booking["duration_minutes"])
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

// bookingStartTimes must match allBookingSlots in the mobile DateTimeSelector.
var bookingStartTimes = []string{"10:00", "10:45", "11:30", "12:15", "13:00", "13:45", "14:30", "15:15", "16:00", "16:45", "17:30", "18:15", "19:00", "19:45"}

const closingTime = "21:00"

var wib = time.FixedZone("WIB", 7*60*60)

// clockMinutes converts "HH:MM" to minutes since midnight.
func clockMinutes(hhmm string) int {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return -1
	}
	return t.Hour()*60 + t.Minute()
}

// slotHasPassed reports whether date+time (WIB) is already in the past.
func slotHasPassed(date, hhmm string) bool {
	start, err := time.ParseInLocation("2006-01-02 15:04", date+" "+hhmm, wib)
	return err != nil || !start.After(time.Now())
}

// CheckAvailability returns the start times still bookable for a barber, service, and date:
// not in the past (WIB), not overlapping another active booking, and ending before closing time.
func (s *BookingService) CheckAvailability(ctx context.Context, barberID, serviceID int64, bookingDate string) ([]string, error) {
	if _, err := time.Parse("2006-01-02", bookingDate); err != nil {
		return nil, errors.New("format tanggal harus YYYY-MM-DD")
	}
	service, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, errors.New("layanan tidak ditemukan")
	}
	barber, err := s.repo.GetBarberByID(ctx, barberID)
	if err != nil {
		return nil, errors.New("barber tidak ditemukan")
	}
	if available, _ := barber["available"].(bool); !available {
		return []string{}, nil
	}

	booked, err := s.repo.GetBookedSlots(ctx, barberID, bookingDate)
	if err != nil {
		log.Printf("get booked slots failed: %v", err)
		return nil, errors.New("gagal memeriksa slot")
	}

	duration := asInt(service["duration_minutes"])
	closing := clockMinutes(closingTime)
	slots := []string{}
	for _, slot := range bookingStartTimes {
		start := clockMinutes(slot)
		end := start + duration
		if end > closing || slotHasPassed(bookingDate, slot) {
			continue
		}
		conflict := false
		for _, b := range booked {
			bStart := clockMinutes(b["booking_time"].(string))
			bEnd := bStart + asInt(b["duration_minutes"])
			if start < bEnd && bStart < end {
				conflict = true
				break
			}
		}
		if !conflict {
			slots = append(slots, slot)
		}
	}
	return slots, nil
}

// CreatePayment creates a new payment transaction
func (s *BookingService) CreatePayment(ctx context.Context, bookingID, userID int64) (string, error) {
	// Get booking details
	booking, err := s.repo.GetBookingByID(ctx, bookingID, userID)
	if err != nil {
		return "", errors.New("booking tidak ditemukan")
	}

	if status, _ := booking["status"].(string); status == "Batal" || status == "Selesai" {
		return "", errors.New("booking ini sudah tidak aktif")
	}

	paymentStatus := booking["payment_status"].(*string)
	if paymentStatus != nil && *paymentStatus == "Lunas" {
		return "", errors.New("booking ini sudah dibayar")
	}

	// Reuse the open Snap session. Once it expires the payment expiry worker
	// cancels the booking, so a new session would be cancelled with it.
	if paymentStatus != nil && *paymentStatus == "Menunggu" {
		if expiresAt := booking["expires_at"].(*time.Time); expiresAt != nil && !time.Now().Before(*expiresAt) {
			return "", errors.New("sesi pembayaran sudah kedaluwarsa. Silakan buat booking baru")
		}
		redirectURL := booking["payment_redirect_url"].(*string)
		if redirectURL != nil && *redirectURL != "" {
			return *redirectURL, nil
		}
	}
	ticketCode, _ := booking["ticket_code"].(string)
	name, _ := booking["customer_name"].(string)
	phone, _ := booking["customer_phone"].(string)
	amount, _ := booking["total_amount"].(float64)
	serviceName, _ := booking["service_name"].(string)
	_, _, redirectURL, _, err := s.startPayment(ctx, bookingID, ticketCode, name, phone, amount, serviceName)
	return redirectURL, err
}

// GetMyTickets returns the user's active ticket, visit history, and stats
func (s *BookingService) GetMyTickets(ctx context.Context, userID int64) (*dto.MyTicketsResponse, error) {
	profile, err := s.repo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, errors.New("gagal mengambil profil")
	}
	rows, err := s.repo.GetMyTickets(ctx, userID)
	if err != nil {
		log.Printf("get my tickets failed: %v", err)
		return nil, errors.New("gagal mengambil tiket")
	}

	resp := &dto.MyTicketsResponse{
		User: dto.RegisteredUser{
			ID:    profile["id"].(int64),
			Name:  profile["name"].(string),
			Phone: profile["phone"].(string),
			Email: profile["email"].(*string),
			Role:  profile["role"].(string),
		},
		History: []dto.TicketBooking{},
	}

	barberVisits := map[string]int{}
	// Rows are ordered by schedule ascending: the first open booking is the active one.
	for _, t := range rows {
		ticket := dto.TicketBooking{
			ID:                 t["id"].(int64),
			TicketCode:         t["ticket_code"].(string),
			Status:             t["status"].(string),
			BookingDate:        t["booking_date"].(string),
			BookingTime:        t["booking_time"].(string),
			ChairNumber:        asInt(t["chair_number"]),
			Notes:              t["notes"].(string),
			ServiceName:        t["service_name"].(string),
			ServiceDesc:        t["service_description"].(*string),
			DurationMinutes:    asInt(t["duration_minutes"]),
			Price:              t["price"].(float64),
			TotalAmount:        t["total_amount"].(float64),
			BarberName:         t["barber_name"].(string),
			BarberRating:       t["barber_rating"].(float64),
			PaymentMethod:      t["payment_method"].(*string),
			PaymentStatus:      t["payment_status"].(*string),
			PaymentRedirectURL: t["payment_redirect_url"].(*string),
			PaymentExpiresAt:   t["expires_at"].(*time.Time),
		}
		failed := ticket.PaymentStatus != nil && *ticket.PaymentStatus == "Gagal"
		switch {
		case ticket.Status == "Selesai":
			resp.Stats.Visits++
			barberVisits[ticket.BarberName]++
			resp.History = append(resp.History, ticket)
		case ticket.Status == "Batal" || failed:
			resp.History = append(resp.History, ticket)
		case resp.Active == nil:
			active := ticket
			resp.Active = &active
		}
	}

	// History newest first
	for i, j := 0, len(resp.History)-1; i < j; i, j = i+1, j-1 {
		resp.History[i], resp.History[j] = resp.History[j], resp.History[i]
	}
	for name, count := range barberVisits {
		if count > barberVisits[resp.Stats.FavoriteBarber] {
			resp.Stats.FavoriteBarber = name
		}
	}
	return resp, nil
}

// HandleMidtransNotification verifies a Midtrans notification and stores the payment result
func (s *BookingService) HandleMidtransNotification(ctx context.Context, n dto.MidtransNotification) error {
	if s.midtrans == nil {
		return ErrPaymentUnavailable
	}
	if !s.midtrans.VerifySignature(n.OrderID, n.StatusCode, n.GrossAmount, n.SignatureKey) {
		return ErrInvalidSignature
	}

	var paymentStatus string
	switch n.TransactionStatus {
	case "settlement":
		paymentStatus = "Lunas"
	case "capture":
		if n.FraudStatus == "" || n.FraudStatus == "accept" {
			paymentStatus = "Lunas"
		} else {
			paymentStatus = "Menunggu"
		}
	case "pending":
		paymentStatus = "Menunggu"
	case "deny", "cancel", "expire", "failure":
		paymentStatus = "Gagal"
	default:
		// refund, partial_refund, etc. are not handled by this app
		return nil
	}

	found, err := s.repo.ApplyPaymentNotification(ctx, n.OrderID, paymentStatus)
	if err != nil {
		log.Printf("apply midtrans notification %s failed: %v", n.OrderID, err)
		return errors.New("gagal memproses notifikasi")
	}
	if !found {
		log.Printf("midtrans notification for unknown order %s", n.OrderID)
	}
	return nil
}

var (
	ErrPaymentUnavailable = errors.New("pembayaran belum dikonfigurasi")
	ErrInvalidSignature   = errors.New("signature notifikasi tidak valid")
)
