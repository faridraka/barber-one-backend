package model

import(
	"time"
)

type User struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Phone     string `json:"phone" db:"phone"`
	Email     *string `json:"email,omitempty" db:"email"`
	Role      string `json:"role" db:"role"`
	Password  string `json:"-" db:"password_hash"`
}

type RegisteredUser struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Phone string  `json:"phone"`
	Email *string `json:"email,omitempty"`
	Role  string  `json:"role"`
}

type Service struct {
	ID              int64   `json:"id" db:"id"`
	Name            string  `json:"name" db:"name"`
	Description     *string `json:"description,omitempty" db:"description"`
	Price           float64 `json:"price" db:"price"`
	DurationMinutes int     `json:"duration_minutes" db:"duration_minutes"`
}

type Barber struct {
	ID        int64   `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	Rating    float64 `json:"rating" db:"rating"`
	Available bool    `json:"available" db:"available"`
}

type Booking struct {
	ID               int64      `json:"id" db:"id"`
	TicketCode       string     `json:"ticket_code" db:"ticket_code"`
	UserID           *int64     `json:"-" db:"user_id"`
	CustomerName     string     `json:"customer_name" db:"customer_name"`
	CustomerPhone    string     `json:"customer_phone" db:"customer_phone"`
	ServiceID        int64      `json:"service_id" db:"service_id"`
	BarberID         int64      `json:"barber_id" db:"barber_id"`
	BookingDate      string     `json:"booking_date" db:"booking_date"`
	BookingTime      string     `json:"booking_time" db:"booking_time"`
	Status           string     `json:"status" db:"status"`
	Notes            *string    `json:"notes,omitempty" db:"notes"`
	ChairNumber      *int       `json:"chair_number,omitempty" db:"chair_number"`
	StartedAt        *time.Time `json:"started_at,omitempty" db:"started_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type Transaction struct {
	ID               int64       `json:"id" db:"id"`
	BookingID        int64       `json:"booking_id" db:"booking_id"`
	TotalAmount      float64     `json:"total_amount" db:"total_amount"`
	PaymentMethod    string      `json:"payment_method" db:"payment_method"`
	PaymentStatus    string      `json:"payment_status" db:"payment_status"`
	GatewayOrderID   string      `json:"gateway_order_id" db:"gateway_order_id"`
	GatewayToken     string      `json:"-" db:"gateway_token"`
	GatewayRedirectURL *string   `json:"payment_redirect_url,omitempty" db:"gateway_redirect_url"`
	ExpiresAt        *time.Time  `json:"expires_at,omitempty" db:"expires_at"`
	PaidAt           *time.Time  `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
}

type Review struct {
	ID        int64     `json:"id" db:"id"`
	BookingID int64     `json:"booking_id" db:"booking_id"`
	Rating    int       `json:"rating" db:"rating"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
