package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/admin/internal/service"
)

// EventHandler handles admin event management endpoints.
type EventHandler struct {
	eventSvc      *service.EventService
	settlementSvc *service.SettlementService
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(eventSvc *service.EventService, settlementSvc *service.SettlementService) *EventHandler {
	return &EventHandler{
		eventSvc:      eventSvc,
		settlementSvc: settlementSvc,
	}
}

// ListEvents returns a paginated list of events with optional filters.
func (h *EventHandler) ListEvents(c *gin.Context) {
	page, pageSize := parsePagination(c)
	status := c.Query("status")
	category := c.Query("category")
	search := c.Query("search")

	events, total, err := h.eventSvc.List(c.Request.Context(), status, category, search, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list events")
		return
	}

	response.Paginated(c, events, total, page, pageSize)
}

type patchEventRequest struct {
	Category *string `json:"category"`
	Status   *string `json:"status"`
}

// PatchEvent performs a partial update on an event.
func (h *EventHandler) PatchEvent(c *gin.Context) {
	id := c.Param("id")

	var req patchEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body")
		return
	}

	if req.Category == nil && req.Status == nil {
		response.ValidationError(c, "at least one field (category, status) is required")
		return
	}

	event, err := h.eventSvc.Update(c.Request.Context(), id, req.Category, req.Status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update event")
		return
	}

	response.Success(c, event)
}

type settleEventRequest struct {
	Outcome string `json:"outcome" binding:"required"`
}

// SettleEvent force-settles an event with the given outcome.
func (h *EventHandler) SettleEvent(c *gin.Context) {
	id := c.Param("id")

	var req settleEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "outcome is required")
		return
	}

	settlement, err := h.settlementSvc.ForceSettle(c.Request.Context(), id, req.Outcome)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(c, settlement)
}
