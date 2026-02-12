package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/api/internal/repository"
	"github.com/poly-predict/backend/services/api/internal/service"
)

// EventHandler handles event-related HTTP requests.
type EventHandler struct {
	service *service.EventService
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(service *service.EventService) *EventHandler {
	return &EventHandler{service: service}
}

// ListEvents handles GET /api/v1/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := repository.EventFilters{
		Status:   c.Query("status"),
		Category: c.Query("category"),
		Search:   c.Query("search"),
		Sort:     c.Query("sort"),
		Page:     page,
		PageSize: pageSize,
	}

	events, total, err := h.service.List(c.Request.Context(), filters)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list events")
		return
	}

	if events == nil {
		events = []model.Event{}
	}

	response.Paginated(c, events, total, page, pageSize)
}

// GetEvent handles GET /api/v1/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	id := c.Param("id")

	event, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get event")
		return
	}

	if event == nil {
		response.Error(c, http.StatusNotFound, "event not found")
		return
	}

	response.Success(c, event)
}

// GetPriceHistory handles GET /api/v1/events/:id/price-history
func (h *EventHandler) GetPriceHistory(c *gin.Context) {
	eventID := c.Param("id")
	period := c.Query("period")
	if period == "" {
		period = "24h"
	}

	history, err := h.service.GetPriceHistory(c.Request.Context(), eventID, period)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get price history")
		return
	}

	if history == nil {
		history = []model.PriceHistory{}
	}

	response.Success(c, history)
}

// GetCategories handles GET /api/v1/events/categories
func (h *EventHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get categories")
		return
	}

	if categories == nil {
		categories = []repository.CategoryCount{}
	}

	response.Success(c, categories)
}
