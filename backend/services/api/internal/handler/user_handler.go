package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/api/internal/repository"
)

// updateProfileRequest is the JSON body for updating a user profile.
type updateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	repo *repository.UserRepository
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// GetProfile handles GET /api/v1/users/me
// Auto-creates the user on first authenticated request with 10,000 credits.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Auto-create user with default display name if they don't exist yet.
	user, err := h.repo.GetOrCreate(c.Request.Context(), userID, "Player")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get user profile")
		return
	}

	if user == nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.Success(c, user)
}

// UpdateProfile handles PATCH /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request: "+err.Error())
		return
	}

	user, err := h.repo.UpdateDisplayName(c.Request.Context(), userID, req.DisplayName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update profile")
		return
	}

	if user == nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.Success(c, user)
}

// GetTransactions handles GET /api/v1/users/me/transactions
func (h *UserHandler) GetTransactions(c *gin.Context) {
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

	transactions, total, err := h.repo.GetTransactions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get transactions")
		return
	}

	if transactions == nil {
		transactions = []model.CreditTransaction{}
	}

	response.Paginated(c, transactions, total, page, pageSize)
}
