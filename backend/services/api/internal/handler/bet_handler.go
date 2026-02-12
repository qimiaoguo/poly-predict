package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/api/internal/service"
)

// placeBetRequest is the JSON body for placing a bet.
type placeBetRequest struct {
	EventID string `json:"event_id" binding:"required"`
	Outcome string `json:"outcome" binding:"required"`
	Amount  int64  `json:"amount" binding:"required,gt=0"`
}

// BetHandler handles bet-related HTTP requests.
type BetHandler struct {
	service *service.BetService
}

// NewBetHandler creates a new BetHandler.
func NewBetHandler(service *service.BetService) *BetHandler {
	return &BetHandler{service: service}
}

// PlaceBet handles POST /api/v1/bets
func (h *BetHandler) PlaceBet(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req placeBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request: "+err.Error())
		return
	}

	bet, err := h.service.PlaceBet(c.Request.Context(), userID, req.EventID, req.Outcome, req.Amount)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(c, bet)
}

// ListBets handles GET /api/v1/bets
func (h *BetHandler) ListBets(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	status := c.Query("status")

	bets, total, err := h.service.ListByUser(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list bets")
		return
	}

	if bets == nil {
		bets = []model.Bet{}
	}

	response.Paginated(c, bets, total, page, pageSize)
}

// GetBet handles GET /api/v1/bets/:id
func (h *BetHandler) GetBet(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")

	bet, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get bet")
		return
	}

	if bet == nil {
		response.Error(c, http.StatusNotFound, "bet not found")
		return
	}

	// Ensure the bet belongs to the requesting user.
	if bet.UserID != userID {
		response.Error(c, http.StatusNotFound, "bet not found")
		return
	}

	response.Success(c, bet)
}
