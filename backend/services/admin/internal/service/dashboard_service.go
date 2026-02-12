package service

import (
	"context"

	"github.com/poly-predict/backend/services/admin/internal/repository"
)

// DashboardService wraps the DashboardRepository.
type DashboardService struct {
	repo *repository.DashboardRepository
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(repo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

// GetStats returns aggregate dashboard statistics.
func (s *DashboardService) GetStats(ctx context.Context) (*repository.DashboardStats, error) {
	return s.repo.GetStats(ctx)
}
