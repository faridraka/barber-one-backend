package dto

import (
	"time"
)

type CreateBookingRequest struct {
	ServiceID     int64  `json:"service_id"`
	BarberID      int64  `json:"barber_id"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	BookingDate   string `json:"booking_date"`
	BookingTime   string `json:"booking_time"`
	Notes         string `json:"notes"`
}

type CreatedBooking struct {
	ID                 int64  `json:"id"`
	TicketCode         string `json:"ticket_code"`
	Status             string `json:"status"`
	PaymentStatus      string `json:"payment_status"`
	PaymentRedirectURL string `json:"payment_redirect_url"`
}

type CreateBookingResponse struct {
	Message string         `json:"message"`
	Booking CreatedBooking `json:"booking"`
}

type UserBooking struct {
	ID                 int64       `json:"id"`
	TicketCode         string      `json:"ticket_code"`
	Status             string      `json:"status"`
	CustomerName       string      `json:"customer_name"`
	CustomerPhone      string      `json:"customer_phone"`
	ServiceID          int64       `json:"service_id"`
	ServiceName        string      `json:"service_name"`
	Price              float64     `json:"price"`
	DurationMinutes    int         `json:"duration_minutes"`
	BarberID           int64       `json:"barber_id"`
	BarberName         string      `json:"barber_name"`
	BookingDate        string      `json:"booking_date"`
	BookingTime        string      `json:"booking_time"`
	ChairNumber        int         `json:"chair_number"`
	Notes              string      `json:"notes"`
	TotalAmount        float64     `json:"total_amount"`
	PaymentMethod      *string     `json:"payment_method,omitempty"`
	PaymentStatus      *string     `json:"payment_status,omitempty"`
	PaymentRedirectURL *string     `json:"payment_redirect_url,omitempty"`
	PaymentExpiresAt   *time.Time  `json:"payment_expires_at,omitempty"`
}

type RescheduleBookingRequest struct {
	BookingDate string `json:"booking_date"`
	BookingTime string `json:"booking_time"`
}

type ReviewRequest struct {
	BookingID int64  `json:"booking_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}

type CreateReviewResponse struct {
	Message string `json:"message"`
	Review  Review `json:"review"`
}

type ReviewListItem struct {
	ID           int64     `json:"id"`
	BookingID    int64     `json:"booking_id"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"created_at"`
	CustomerName string    `json:"customer_name"`
	ServiceName  string    `json:"service_name"`
	BarberName   string    `json:"barber_name"`
}

type AvailabilityResponse struct {
	BarberID       int64    `json:"barber_id"`
	ServiceID      int64    `json:"service_id"`
	Date           string   `json:"date"`
	AvailableSlots []string `json:"available_slots"`
}

type CashierQueueItem struct {
	ID              int64      `json:"id"`
	TicketCode      string     `json:"ticket_code"`
	Status          string     `json:"status"`
	IsWalkIn        bool       `json:"is_walk_in"`
	CustomerName    string     `json:"customer_name"`
	CustomerPhone   string     `json:"customer_phone"`
	ServiceName     string     `json:"service_name"`
	DurationMinutes int        `json:"duration_minutes"`
	BarberID        int64      `json:"barber_id"`
	BarberName      string     `json:"barber_name"`
	ChairNumber     int        `json:"chair_number"`
	BookingDate     string     `json:"booking_date"`
	BookingTime     string     `json:"booking_time"`
	TotalAmount     float64    `json:"total_amount"`
	PaymentMethod   *string    `json:"payment_method,omitempty"`
	PaymentStatus   *string    `json:"payment_status,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
}

type CashierQueueResponse struct {
	Date       string             `json:"date"`
	ServerTime time.Time          `json:"server_time"`
	Queue      []CashierQueueItem `json:"queue"`
}

type CashierSummaryResponse struct {
	Date         string  `json:"date"`
	Waiting      int     `json:"waiting"`
	InService    int     `json:"in_service"`
	Done         int     `json:"done"`
	TotalChairs  int     `json:"total_chairs"`
	Revenue      float64 `json:"revenue"`
	CashCount    int     `json:"cash_count"`
	NonCashCount int     `json:"non_cash_count"`
}

type WalkInRequest struct {
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	ServiceID     int64  `json:"service_id"`
	BarberID      int64  `json:"barber_id"`
	PaymentMethod string `json:"payment_method"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type RegisterRequest struct {
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    *string `json:"email,omitempty"`
	Password string  `json:"password"`
}

type MidtransNotification struct {
	MerchantID        string `json:"merchant_id"`
	OrderID           string `json:"order_id"`
	TransactionID     string `json:"transaction_id"`
	TransactionTime   string `json:"transaction_time"`
	SettlementTime    string `json:"settlement_time"`
	PaymentType       string `json:"payment_type"`
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	GrossAmount       string `json:"gross_amount"`
	Currency          string `json:"currency"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

type TicketBooking struct {
	ID                 int64       `json:"id"`
	TicketCode         string      `json:"ticket_code"`
	Status             string      `json:"status"`
	BookingDate        string      `json:"booking_date"`
	BookingTime        string      `json:"booking_time"`
	ChairNumber        int         `json:"chair_number"`
	Notes              string      `json:"notes"`
	ServiceName        string      `json:"service_name"`
	ServiceDesc        *string     `json:"service_description,omitempty"`
	DurationMinutes    int         `json:"duration_minutes"`
	Price              float64     `json:"price"`
	TotalAmount        float64     `json:"total_amount"`
	BarberName         string      `json:"barber_name"`
	BarberRating       float64     `json:"barber_rating"`
	PaymentMethod      *string     `json:"payment_method,omitempty"`
	PaymentStatus      *string     `json:"payment_status,omitempty"`
	PaymentRedirectURL *string     `json:"payment_redirect_url,omitempty"`
	PaymentExpiresAt   *time.Time  `json:"payment_expires_at,omitempty"`
}

type TicketStats struct {
	Visits         int    `json:"visits"`
	FavoriteBarber string `json:"favorite_barber"`
}

type MyTicketsResponse struct {
	User    RegisteredUser `json:"user"`
	Stats   TicketStats    `json:"stats"`
	Active  *TicketBooking `json:"active,omitempty"`
	History []TicketBooking `json:"history"`
}

type Review struct {
	ID        int64     `json:"id"`
	BookingID int64     `json:"booking_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisteredUser struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Phone string  `json:"phone"`
	Email *string `json:"email,omitempty"`
	Role  string  `json:"role"`
}
