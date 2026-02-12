package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/admin/internal/service"
)

// UserHandler handles admin user management endpoints.
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// ListUsers returns a paginated list of users.
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePagination(c)
	search := c.Query("search")
	sortBy := c.Query("sort_by")

	users, total, err := h.svc.List(c.Request.Context(), search, sortBy, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	response.Paginated(c, users, total, page, pageSize)
}

// GetUser returns a single user by ID.
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.Success(c, user)
}

type patchUserRequest struct {
	BalanceAdjustment *int64 `json:"balance_adjustment"`
}

// PatchUser updates a user's balance.
func (h *UserHandler) PatchUser(c *gin.Context) {
	id := c.Param("id")

	var req patchUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body")
		return
	}

	if req.BalanceAdjustment == nil {
		response.ValidationError(c, "balance_adjustment is required")
		return
	}

	user, err := h.svc.AdjustBalance(c.Request.Context(), id, *req.BalanceAdjustment)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to adjust balance")
		return
	}

	response.Success(c, user)
}

// parsePagination extracts page and page_size from query parameters with defaults.
func parsePagination(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20

	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	return page, pageSize
}
