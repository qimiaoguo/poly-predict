package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Middleware provides JWT authentication middleware for Gin.
type Middleware struct {
	jwtSecret []byte
}

// NewMiddleware creates a new auth Middleware with the given JWT secret.
func NewMiddleware(jwtSecret string) *Middleware {
	return &Middleware{
		jwtSecret: []byte(jwtSecret),
	}
}

// RequireAuth returns a Gin middleware that requires a valid Supabase JWT.
// It extracts the user ID from the "sub" claim and sets it in the Gin context as "user_id".
// Returns 401 if the token is missing or invalid.
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := extractBearerToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"message": "missing or malformed authorization header"},
			})
			return
		}

		userID, err := m.validateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"message": "invalid or expired token"},
			})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// OptionalAuth returns a Gin middleware that extracts user_id from a valid JWT if present,
// but does not block the request if the token is missing or invalid.
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := extractBearerToken(c)
		if !ok {
			c.Next()
			return
		}

		userID, err := m.validateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// extractBearerToken extracts the token string from the Authorization header.
func extractBearerToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

// validateToken parses and validates a JWT token using HMAC HS256 and returns the user ID from the "sub" claim.
func (m *Middleware) validateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return "", jwt.ErrTokenInvalidClaims
	}

	return sub, nil
}
