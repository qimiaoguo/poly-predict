package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/pkg/response"
)

// AdminAuth provides authentication functionality for admin users.
type AdminAuth struct {
	pool      *pgxpool.Pool
	jwtSecret string
}

// NewAdminAuth creates a new AdminAuth instance.
func NewAdminAuth(pool *pgxpool.Pool, jwtSecret string) *AdminAuth {
	return &AdminAuth{
		pool:      pool,
		jwtSecret: jwtSecret,
	}
}

// Login verifies admin credentials and returns a signed JWT token.
func (a *AdminAuth) Login(ctx context.Context, email, password string) (string, *model.AdminUser, error) {
	admin := &model.AdminUser{}
	err := a.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at
		 FROM admin_users
		 WHERE email = $1`, email,
	).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Role, &admin.CreatedAt)
	if err != nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT.
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   admin.ID,
		"email": admin.Email,
		"role":  admin.Role,
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, admin, nil
}

// Middleware returns a Gin middleware that verifies JWT tokens from the
// Authorization header and sets admin_id in the request context.
func (a *AdminAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, http.StatusUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "invalid token claims")
			c.Abort()
			return
		}

		adminID, ok := claims["sub"].(string)
		if !ok || adminID == "" {
			response.Error(c, http.StatusUnauthorized, "invalid token subject")
			c.Abort()
			return
		}

		c.Set("admin_id", adminID)
		c.Next()
	}
}

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}
