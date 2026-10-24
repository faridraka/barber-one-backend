package service

import (
	"context"

	"barber-one/backend/internal/repository"
)

// BarberService handles barber business logic
type BarberService struct {
	repo *repository.Repository
}

func NewBarberService(repo *repository.Repository) *BarberService {
	return &BarberService{repo: repo}
}

// ListBarbers returns all barbers
func (s *BarberService) ListBarbers(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListBarbers(ctx)
}

// GetBarberByID returns a specific barber by ID
func (s *BarberService) GetBarberByID(ctx context.Context, id int64) (map[string]any, error) {
	return s.repo.GetBarberByID(ctx, id)
}
