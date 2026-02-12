package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	clobBaseURL    = "https://clob.polymarket.com"
	clobReqTimeout = 30 * time.Second
)

// midpointResponse represents the JSON response from the CLOB midpoint endpoint.
type midpointResponse struct {
	Mid string `json:"mid"`
}

// CLOBClient is an HTTP client for the Polymarket CLOB API.
type CLOBClient struct {
	client *http.Client
}

// NewCLOBClient creates a new CLOBClient.
func NewCLOBClient() *CLOBClient {
	return &CLOBClient{
		client: &http.Client{
			Timeout: clobReqTimeout,
		},
	}
}

// GetMidpoint fetches the midpoint price for a single CLOB token.
func (c *CLOBClient) GetMidpoint(ctx context.Context, tokenID string) (float64, error) {
	endpoint := fmt.Sprintf("%s/midpoint?token_id=%s", clobBaseURL, url.QueryEscape(tokenID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result midpointResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	mid, err := strconv.ParseFloat(result.Mid, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing midpoint %q: %w", result.Mid, err)
	}

	return mid, nil
}

// GetMidpoints fetches midpoint prices for multiple CLOB tokens. It returns
// a map from token ID to its midpoint price. If fetching an individual token
// fails, that token is skipped and the error is logged.
func (c *CLOBClient) GetMidpoints(ctx context.Context, tokenIDs []string) (map[string]float64, error) {
	midpoints := make(map[string]float64, len(tokenIDs))

	for _, tokenID := range tokenIDs {
		if tokenID == "" {
			continue
		}

		mid, err := c.GetMidpoint(ctx, tokenID)
		if err != nil {
			log.Warn().
				Err(err).
				Str("token_id", tokenID).
				Msg("failed to fetch midpoint, skipping token")
			continue
		}

		midpoints[tokenID] = mid
	}

	return midpoints, nil
}
