package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// jwksKey represents a single key from a JWKS response.
type jwksKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

// jwksResponse represents a JWKS endpoint response.
type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

// Middleware provides JWT authentication middleware for Gin.
type Middleware struct {
	jwtSecret []byte
	jwksURL   string

	mu      sync.RWMutex
	keys    map[string]*ecdsa.PublicKey
	fetched time.Time
}

// NewMiddleware creates a new auth Middleware.
// jwtSecret is used for HS256 tokens (legacy).
// supabaseURL is used to fetch JWKS for ES256 tokens.
func NewMiddleware(jwtSecret string, supabaseURL string) *Middleware {
	m := &Middleware{
		jwtSecret: []byte(jwtSecret),
		keys:      make(map[string]*ecdsa.PublicKey),
	}
	if supabaseURL != "" {
		m.jwksURL = strings.TrimRight(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"
	}
	return m
}

// RequireAuth returns a Gin middleware that requires a valid Supabase JWT.
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

// OptionalAuth returns a Gin middleware that extracts user_id if a valid JWT is present.
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

// validateToken tries ES256 (via JWKS) first, then falls back to HS256.
func (m *Middleware) validateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodECDSA:
			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, fmt.Errorf("missing kid in token header")
			}
			key, err := m.getECDSAKey(kid)
			if err != nil {
				return nil, err
			}
			return key, nil
		case *jwt.SigningMethodHMAC:
			return m.jwtSecret, nil
		default:
			return nil, jwt.ErrSignatureInvalid
		}
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

// getECDSAKey returns the ECDSA public key for the given kid, fetching JWKS if needed.
func (m *Middleware) getECDSAKey(kid string) (*ecdsa.PublicKey, error) {
	m.mu.RLock()
	key, ok := m.keys[kid]
	fetched := m.fetched
	m.mu.RUnlock()

	if ok {
		return key, nil
	}

	// Refresh JWKS if stale (> 5 min) or key not found.
	if m.jwksURL == "" {
		return nil, fmt.Errorf("JWKS URL not configured")
	}

	if time.Since(fetched) < 5*time.Second {
		return nil, fmt.Errorf("key %s not found in JWKS", kid)
	}

	if err := m.fetchJWKS(); err != nil {
		return nil, fmt.Errorf("fetching JWKS: %w", err)
	}

	m.mu.RLock()
	key, ok = m.keys[kid]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("key %s not found in JWKS", kid)
	}

	return key, nil
}

// fetchJWKS fetches the JWKS from Supabase and caches the EC public keys.
func (m *Middleware) fetchJWKS() error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(m.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	keys := make(map[string]*ecdsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "EC" || k.Crv != "P-256" {
			continue
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			continue
		}

		keys[k.Kid] = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
	}

	m.mu.Lock()
	m.keys = keys
	m.fetched = time.Now()
	m.mu.Unlock()

	log.Info().Int("keys", len(keys)).Msg("JWKS keys loaded")

	return nil
}
