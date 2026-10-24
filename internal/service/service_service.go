package service

import (
	"context"

	"barber-one/backend/internal/repository"
)

// ServiceService handles service business logic
type ServiceService struct {
	repo *repository.Repository
}

func NewServiceService(repo *repository.Repository) *ServiceService {
	return &ServiceService{repo: repo}
}

// ListServices returns all available services
func (s *ServiceService) ListServices(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListServices(ctx)
}

// GetServiceByID returns a specific service by ID
func (s *ServiceService) GetServiceByID(ctx context.Context, id int64) (map[string]any, error) {
	return s.repo.GetServiceByID(ctx, id)
}
