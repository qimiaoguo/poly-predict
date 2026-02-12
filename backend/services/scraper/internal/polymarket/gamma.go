package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	gammaBaseURL     = "https://gamma-api.polymarket.com"
	gammaPageLimit   = 100
	gammaRateDelay   = 200 * time.Millisecond
	gammaReqTimeout  = 30 * time.Second
)

// stringSlice is a []string that can unmarshal from a JSON-encoded string
// (e.g. "[\"Yes\",\"No\"]") or a regular JSON array.
type stringSlice []string

func (s *stringSlice) UnmarshalJSON(data []byte) error {
	// Try normal array first.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}
	// Otherwise it's a JSON string wrapping an array.
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	return json.Unmarshal([]byte(raw), (*[]string)(s))
}

// flexString accepts both JSON strings and numbers, storing the value as a string.
type flexString string

func (f *flexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = flexString(n.String())
		return nil
	}
	*f = ""
	return nil
}

func (f flexString) String() string { return string(f) }

// GammaMarket represents a market from the Gamma API response.
type GammaMarket struct {
	ID             string      `json:"id"`
	Question       string      `json:"question"`
	Slug           string      `json:"slug"`
	Description    string      `json:"description"`
	Category       string      `json:"category"`
	Image          string      `json:"image"`
	Outcomes       stringSlice `json:"outcomes"`
	OutcomePrices  stringSlice `json:"outcomePrices"`
	ClobTokenIDs   stringSlice `json:"clobTokenIds"`
	Volume         flexString  `json:"volume"`
	Volume24hr     flexString  `json:"volume24hr"`
	Liquidity      flexString  `json:"liquidity"`
	EndDate        string      `json:"endDate"`
	Closed         bool        `json:"closed"`
	ConditionID    string      `json:"conditionId"`
	EventID        string      `json:"eventId"`
	Active         bool        `json:"active"`
}

// GammaClient is an HTTP client for the Polymarket Gamma API.
type GammaClient struct {
	client *http.Client
}

// NewGammaClient creates a new GammaClient.
func NewGammaClient() *GammaClient {
	return &GammaClient{
		client: &http.Client{
			Timeout: gammaReqTimeout,
		},
	}
}

// FetchMarkets fetches all active markets from the Gamma API, paginating until
// the response is empty.
func (g *GammaClient) FetchMarkets(ctx context.Context) ([]GammaMarket, error) {
	var allMarkets []GammaMarket
	offset := 0

	for {
		url := fmt.Sprintf("%s/markets?closed=false&limit=%d&offset=%d", gammaBaseURL, gammaPageLimit, offset)

		log.Debug().
			Int("offset", offset).
			Int("limit", gammaPageLimit).
			Msg("fetching markets page from Gamma API")

		markets, err := g.fetchPage(ctx, url)
		if err != nil {
			return allMarkets, fmt.Errorf("fetching page at offset %d: %w", offset, err)
		}

		if len(markets) == 0 {
			break
		}

		allMarkets = append(allMarkets, markets...)
		offset += gammaPageLimit

		log.Debug().
			Int("page_count", len(markets)).
			Int("total_so_far", len(allMarkets)).
			Msg("fetched markets page")

		// Simple rate limiting between pages.
		select {
		case <-ctx.Done():
			return allMarkets, ctx.Err()
		case <-time.After(gammaRateDelay):
		}
	}

	log.Info().
		Int("total_markets", len(allMarkets)).
		Msg("finished fetching all markets from Gamma API")

	return allMarkets, nil
}

// fetchPage fetches a single page of markets from the given URL.
func (g *GammaClient) fetchPage(ctx context.Context, url string) ([]GammaMarket, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var markets []GammaMarket
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return markets, nil
}
