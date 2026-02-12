package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/admin/internal/service"
)

// DashboardHandler handles admin dashboard endpoints.
type DashboardHandler struct {
	svc *service.DashboardService
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// GetDashboard returns aggregate admin dashboard statistics.
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load dashboard stats")
		return
	}

	response.Success(c, stats)
}
