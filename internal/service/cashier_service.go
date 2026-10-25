package service

import (
	"context"
	"errors"
	"time"

	"barber-one/backend/internal/dto"
	"barber-one/backend/internal/repository"
)

// CashierService handles cashier business logic
type CashierService struct {
	repo *repository.Repository
}

func NewCashierService(repo *repository.Repository) *CashierService {
	return &CashierService{repo: repo}
}

// GetQueue returns the current queue for cashier
func (s *CashierService) GetQueue(ctx context.Context, date string) ([]dto.CashierQueueItem, error) {
	queue, err := s.repo.GetCashierQueue(ctx, date)
	if err != nil {
		return nil, err
	}
	return convertToCashierQueueItem(queue), nil
}

// GetSummary returns cashier summary
func (s *CashierService) GetSummary(ctx context.Context, date string) (dto.CashierSummaryResponse, error) {
	waiting, err := s.repo.GetCashierSummary(ctx, date)
	if err != nil {
		return dto.CashierSummaryResponse{}, err
	}

	revenue, err := s.repo.GetRevenueSummary(ctx, date)
	if err != nil {
		return dto.CashierSummaryResponse{}, err
	}

	return dto.CashierSummaryResponse{
		Date:         date,
		Waiting:      asInt(waiting["waiting"]),
		InService:    asInt(waiting["in_service"]),
		Done:         asInt(waiting["done"]),
		TotalChairs:  totalChairs,
		Revenue:      revenue["revenue"].(float64),
		CashCount:    asInt(revenue["cash_count"]),
		NonCashCount: asInt(revenue["non_cash_count"]),
	}, nil
}

// WalkIn handles walk-in booking
func (s *CashierService) WalkIn(ctx context.Context, req dto.WalkInRequest, cashierID int64) (dto.CashierQueueItem, error) {
	// TODO: Implement walk-in booking logic
	return dto.CashierQueueItem{}, nil
}

// ErrBookingNotFound is returned when the booking ID does not exist
var ErrBookingNotFound = errors.New("booking tidak ditemukan")

// CashierConflictError means the booking is not in a state that allows the action
type CashierConflictError struct{ Message string }

func (e *CashierConflictError) Error() string { return e.Message }

// StartService seats a paid booking in a free chair (triggered manually by the cashier)
func (s *CashierService) StartService(ctx context.Context, bookingID int64) (int, error) {
	chair, reason, err := s.repo.StartBookingService(ctx, bookingID, totalChairs)
	if err != nil {
		return 0, err
	}
	switch reason {
	case "":
		return chair, nil
	case "not_found":
		return 0, ErrBookingNotFound
	case "status":
		return 0, &CashierConflictError{"Booking ini tidak sedang menunggu layanan."}
	case "unpaid":
		return 0, &CashierConflictError{"Booking ini belum lunas."}
	case "barber_busy":
		return 0, &CashierConflictError{"Barber ini sedang melayani pelanggan lain."}
	default: // "no_chair"
		return 0, &CashierConflictError{"Semua kursi sedang terisi."}
	}
}

// FinishService marks a service as finished
func (s *CashierService) FinishService(ctx context.Context, bookingID int64) error {
	return s.endService(ctx, bookingID, "Selesai")
}

// CancelService cancels a service
func (s *CashierService) CancelService(ctx context.Context, bookingID int64) error {
	return s.endService(ctx, bookingID, "Batal")
}

func (s *CashierService) endService(ctx context.Context, bookingID int64, status string) error {
	found, updated, err := s.repo.EndBookingService(ctx, bookingID, status)
	if err != nil {
		return err
	}
	if !found {
		return ErrBookingNotFound
	}
	if !updated {
		return &CashierConflictError{"Booking ini tidak sedang dilayani."}
	}
	return nil
}

const totalChairs = 3

func convertToCashierQueueItem(items []map[string]any) []dto.CashierQueueItem {
	result := make([]dto.CashierQueueItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.CashierQueueItem{
			ID:              int64(item["id"].(int64)),
			TicketCode:      item["ticket_code"].(string),
			Status:          item["status"].(string),
			IsWalkIn:        item["is_walk_in"].(bool),
			CustomerName:    item["customer_name"].(string),
			CustomerPhone:   item["customer_phone"].(string),
			ServiceName:     item["service_name"].(string),
			DurationMinutes: asInt(item["duration_minutes"]),
			BarberID:        int64(item["barber_id"].(int64)),
			BarberName:      item["barber_name"].(string),
			ChairNumber:     asInt(item["chair_number"]),
			BookingDate:     item["booking_date"].(string),
			BookingTime:     item["booking_time"].(string),
			TotalAmount:     item["total_amount"].(float64),
			PaymentMethod:   item["payment_method"].(*string),
			PaymentStatus:   item["payment_status"].(*string),
			StartedAt:       item["started_at"].(*time.Time),
		})
	}
	return result
}
