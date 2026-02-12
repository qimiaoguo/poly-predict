package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/api/internal/service"
)

// RankingHandler handles ranking-related HTTP requests.
type RankingHandler struct {
	service *service.RankingService
}

// NewRankingHandler creates a new RankingHandler.
func NewRankingHandler(service *service.RankingService) *RankingHandler {
	return &RankingHandler{service: service}
}

// GetRankings handles GET /api/v1/rankings
func (h *RankingHandler) GetRankings(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	period := c.Query("period")
	category := c.Query("category")
	sortBy := c.Query("sort_by")

	rankings, total, err := h.service.GetRankings(c.Request.Context(), period, category, sortBy, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get rankings")
		return
	}

	if rankings == nil {
		rankings = []model.Ranking{}
	}

	response.Paginated(c, rankings, total, page, pageSize)
}
