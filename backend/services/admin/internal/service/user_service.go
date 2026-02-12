package service

import (
	"context"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/services/admin/internal/repository"
)

// UserService wraps the UserRepository.
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// List returns a paginated list of users.
func (s *UserService) List(ctx context.Context, search, sortBy string, page, pageSize int) ([]model.User, int64, error) {
	return s.repo.List(ctx, search, sortBy, page, pageSize)
}

// GetByID returns a single user.
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

// AdjustBalance atomically adjusts a user's balance.
func (s *UserService) AdjustBalance(ctx context.Context, id string, adjustment int64) (*model.User, error) {
	return s.repo.AdjustBalance(ctx, id, adjustment)
}
