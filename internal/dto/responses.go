package dto

import (
	"time"
)

type BookingSummary struct {
	ID              int64       `json:"id"`
	TicketCode      string      `json:"ticket_code"`
	Status          string      `json:"status"`
	CustomerName    string      `json:"customer_name"`
	CustomerPhone   string      `json:"customer_phone"`
	ServiceName     string      `json:"service_name"`
	DurationMinutes int         `json:"duration_minutes"`
	BarberName      string      `json:"barber_name"`
	TotalAmount     float64     `json:"total_amount"`
	PaymentMethod   *string     `json:"payment_method,omitempty"`
	PaymentStatus   *string     `json:"payment_status,omitempty"`
	BookingDate     string      `json:"booking_date"`
	BookingTime     string      `json:"booking_time"`
	StartedAt       *time.Time  `json:"started_at,omitempty"`
}

type ServiceSummary struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	Price           float64 `json:"price"`
	DurationMinutes int     `json:"duration_minutes"`
}

type BarberSummary struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Rating    float64 `json:"rating"`
	Available bool    `json:"available"`
}
