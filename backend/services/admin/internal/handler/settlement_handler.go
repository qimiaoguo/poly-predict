package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/admin/internal/service"
)

// SettlementHandler handles admin settlement endpoints.
type SettlementHandler struct {
	svc *service.SettlementService
}

// NewSettlementHandler creates a new SettlementHandler.
func NewSettlementHandler(svc *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{svc: svc}
}

// ListSettlements returns a paginated list of settlement records.
func (h *SettlementHandler) ListSettlements(c *gin.Context) {
	page, pageSize := parsePagination(c)

	settlements, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list settlements")
		return
	}

	response.Paginated(c, settlements, total, page, pageSize)
}
