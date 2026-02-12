package service

import (
	"context"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/services/admin/internal/repository"
)

// SettlementService wraps the SettlementRepository.
type SettlementService struct {
	repo *repository.SettlementRepository
}

// NewSettlementService creates a new SettlementService.
func NewSettlementService(repo *repository.SettlementRepository) *SettlementService {
	return &SettlementService{repo: repo}
}

// ForceSettle atomically settles an event with the given outcome.
func (s *SettlementService) ForceSettle(ctx context.Context, eventID, outcome string) (*model.Settlement, error) {
	return s.repo.ForceSettle(ctx, eventID, outcome)
}

// List returns a paginated list of settlements.
func (s *SettlementService) List(ctx context.Context, page, pageSize int) ([]model.Settlement, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}
