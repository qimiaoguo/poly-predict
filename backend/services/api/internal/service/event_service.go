package service

import (
	"context"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/services/api/internal/repository"
)

// EventService wraps EventRepository methods.
type EventService struct {
	repo *repository.EventRepository
}

// NewEventService creates a new EventService.
func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

// List retrieves a paginated list of events with filters.
func (s *EventService) List(ctx context.Context, filters repository.EventFilters) ([]model.Event, int64, error) {
	return s.repo.List(ctx, filters)
}

// GetByID retrieves a single event by ID.
func (s *EventService) GetByID(ctx context.Context, id string) (*model.Event, error) {
	return s.repo.GetByID(ctx, id)
}

// GetPriceHistory retrieves price history for an event filtered by period.
func (s *EventService) GetPriceHistory(ctx context.Context, eventID string, period string) ([]model.PriceHistory, error) {
	return s.repo.GetPriceHistory(ctx, eventID, period)
}

// GetCategories retrieves all categories with their event counts.
func (s *EventService) GetCategories(ctx context.Context) ([]repository.CategoryCount, error) {
	return s.repo.GetCategories(ctx)
}
