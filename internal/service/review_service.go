package service

import (
	"context"

	"barber-one/backend/internal/repository"
)

// ReviewService handles review business logic
type ReviewService struct {
	repo *repository.Repository
}

func NewReviewService(repo *repository.Repository) *ReviewService {
	return &ReviewService{repo: repo}
}

// CreateReview creates a new review
func (s *ReviewService) CreateReview(ctx context.Context, bookingID int64, rating int, comment *string, userID int64) (int64, error) {
	// Check if booking belongs to user
	// Check if booking status is "Selesai"
	// TODO: Implement validation
	return s.repo.CreateReview(ctx, bookingID, rating, comment)
}

// ListReviews returns all reviews
func (s *ReviewService) ListReviews(ctx context.Context, barberID *int64) ([]map[string]any, error) {
	return s.repo.ListReviews(ctx, barberID)
}
