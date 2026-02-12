package service

import (
	"context"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/services/admin/internal/repository"
)

// EventService wraps the EventRepository.
type EventService struct {
	repo *repository.EventRepository
}

// NewEventService creates a new EventService.
func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

// List returns a paginated list of events.
func (s *EventService) List(ctx context.Context, status, category, search string, page, pageSize int) ([]model.Event, int64, error) {
	return s.repo.List(ctx, status, category, search, page, pageSize)
}

// GetByID returns a single event.
func (s *EventService) GetByID(ctx context.Context, id string) (*model.Event, error) {
	return s.repo.GetByID(ctx, id)
}

// Update performs a partial update on an event.
func (s *EventService) Update(ctx context.Context, id string, category, status *string) (*model.Event, error) {
	return s.repo.Update(ctx, id, category, status)
}
