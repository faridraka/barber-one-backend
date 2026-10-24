package service

import (
	"context"
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
		Waiting:      int(waiting["waiting"].(int64)),
		InService:    int(waiting["in_service"].(int64)),
		Done:         int(waiting["done"].(int64)),
		TotalChairs:  3,
		Revenue:      revenue["revenue"].(float64),
		CashCount:    int(revenue["cash_count"].(int64)),
		NonCashCount: int(revenue["non_cash_count"].(int64)),
	}, nil
}

// WalkIn handles walk-in booking
func (s *CashierService) WalkIn(ctx context.Context, req dto.WalkInRequest, cashierID int64) (dto.CashierQueueItem, error) {
	// TODO: Implement walk-in booking logic
	return dto.CashierQueueItem{}, nil
}

// FinishService marks a service as finished
func (s *CashierService) FinishService(ctx context.Context, bookingID int64) error {
	// TODO: Implement finish service logic
	return nil
}

// CancelService cancels a service
func (s *CashierService) CancelService(ctx context.Context, bookingID int64) error {
	// TODO: Implement cancel service logic
	return nil
}

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
			DurationMinutes: int(item["duration_minutes"].(int64)),
			BarberID:        int64(item["barber_id"].(int64)),
			BarberName:      item["barber_name"].(string),
			ChairNumber:     int(item["chair_number"].(int64)),
			BookingDate:     item["booking_date"].(string),
			BookingTime:     item["booking_time"].(string),
			TotalAmount:     item["total_amount"].(float64),
			PaymentMethod:   func() *string { s := item["payment_method"].(string); return &s }(),
			PaymentStatus:   func() *string { s := item["payment_status"].(string); return &s }(),
			StartedAt:       func() *time.Time { t := item["started_at"].(time.Time); return &t }(),
		})
	}
	return result
}
