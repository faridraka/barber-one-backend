package repository

import (
	"context"
	"time"
)

//go:generate mockgen -source=database.go -destination=mocks/mock_database.go -package=mocks

type Database interface {
	GetUserBookings(ctx context.Context, userID int64) ([]map[string]any, error)
	GetBookingByID(ctx context.Context, bookingID, userID int64) (map[string]any, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string) (int64, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (map[string]any, error)
	GetUserByID(ctx context.Context, id int64) (map[string]any, error)
	ListServices(ctx context.Context) ([]map[string]any, error)
	ListBarbers(ctx context.Context) ([]map[string]any, error)
	CreateBooking(ctx context.Context, ticketCode string, userID *int64, customerName, customerPhone string, serviceID, barberID int64, bookingDate, bookingTime string, notes *string) (int64, error)
	CreateTransaction(ctx context.Context, bookingID int64, amount float64, paymentMethod, gatewayOrderID, gatewayToken, gatewayRedirectURL string, expiresAt time.Time) (int64, error)
	UpdateTransactionPaymentStatus(ctx context.Context, gatewayOrderID, paymentStatus string, paidAt time.Time) (int64, error)
	GetTransactionByGatewayOrderID(ctx context.Context, gatewayOrderID string) (map[string]any, error)
	GetServiceByID(ctx context.Context, id int64) (map[string]any, error)
	GetBarberByID(ctx context.Context, id int64) (map[string]any, error)
}

type BookingRepository struct {
	db interface{}
}

func NewBookingRepository(db interface{}) *BookingRepository {
	return &BookingRepository{db: db}
}
