package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/poly-predict/backend/pkg/response"
	"github.com/poly-predict/backend/services/admin/internal/auth"
)

// AuthHandler handles admin authentication endpoints.
type AuthHandler struct {
	adminAuth *auth.AdminAuth
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(adminAuth *auth.AdminAuth) *AuthHandler {
	return &AuthHandler{adminAuth: adminAuth}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates an admin user and returns a JWT token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request: email and password are required")
		return
	}

	token, admin, err := h.adminAuth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"admin": admin,
	})
}
