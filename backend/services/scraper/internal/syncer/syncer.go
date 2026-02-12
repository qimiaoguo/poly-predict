package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/poly-predict/backend/services/scraper/internal/polymarket"
)

// Syncer orchestrates the synchronization of Polymarket data into the database.
type Syncer struct {
	pool   *pgxpool.Pool
	gamma  *polymarket.GammaClient
	clob   *polymarket.CLOBClient
}

// New creates a new Syncer.
func New(pool *pgxpool.Pool, gamma *polymarket.GammaClient, clob *polymarket.CLOBClient) *Syncer {
	return &Syncer{
		pool:  pool,
		gamma: gamma,
		clob:  clob,
	}
}

// syncStats tracks statistics for a sync run.
type syncStats struct {
	Total    int
	New      int
	Updated  int
	Resolved int
	Errors   int
}

// SyncAll performs a full sync of all active markets from Polymarket.
func (s *Syncer) SyncAll(ctx context.Context) error {
	startTime := time.Now()
	log.Info().Msg("starting market sync")

	// 1. Fetch all active markets from Gamma.
	markets, err := s.gamma.FetchMarkets(ctx)
	if err != nil {
		return fmt.Errorf("fetching markets: %w", err)
	}

	log.Info().Int("market_count", len(markets)).Msg("fetched markets from Gamma API")

	// 2. Collect the set of condition IDs we see from the API for resolution detection.
	activeConditionIDs := make(map[string]struct{}, len(markets))
	stats := syncStats{Total: len(markets)}

	// 3. Upsert each market and record price history.
	for _, market := range markets {
		if market.ConditionID == "" {
			log.Warn().
				Str("market_id", market.ID).
				Str("question", market.Question).
				Msg("skipping market with empty conditionId")
			stats.Errors++
			continue
		}

		activeConditionIDs[market.ConditionID] = struct{}{}

		isNew, err := s.upsertMarket(ctx, market)
		if err != nil {
			log.Error().
				Err(err).
				Str("condition_id", market.ConditionID).
				Str("question", market.Question).
				Msg("failed to upsert market, continuing")
			stats.Errors++
			continue
		}

		if isNew {
			stats.New++
		} else {
			stats.Updated++
		}

		// Record price history for each outcome with a CLOB token.
		if err := s.recordPriceHistory(ctx, market); err != nil {
			log.Error().
				Err(err).
				Str("condition_id", market.ConditionID).
				Msg("failed to record price history, continuing")
		}
	}

	// 4. Detect resolutions: markets where one outcome price is "1" and another is "0".
	resolved := s.detectResolutions(ctx, markets)
	stats.Resolved = resolved

	elapsed := time.Since(startTime)
	log.Info().
		Int("total", stats.Total).
		Int("new", stats.New).
		Int("updated", stats.Updated).
		Int("resolved", stats.Resolved).
		Int("errors", stats.Errors).
		Dur("elapsed", elapsed).
		Msg("market sync completed")

	return nil
}

// upsertMarket inserts or updates a single market in the events table.
// It returns true if the row was newly inserted, false if it was updated.
func (s *Syncer) upsertMarket(ctx context.Context, m polymarket.GammaMarket) (bool, error) {
	// Marshal slices to JSON for DB storage.
	outcomesJSON, err := json.Marshal(m.Outcomes)
	if err != nil {
		return false, fmt.Errorf("marshaling outcomes: %w", err)
	}

	pricesJSON, err := json.Marshal(m.OutcomePrices)
	if err != nil {
		return false, fmt.Errorf("marshaling outcome prices: %w", err)
	}

	tokenIDsJSON, err := json.Marshal(m.ClobTokenIDs)
	if err != nil {
		return false, fmt.Errorf("marshaling clob token ids: %w", err)
	}

	// Parse numeric fields.
	volume, _ := strconv.ParseFloat(m.Volume, 64)
	volume24h, _ := strconv.ParseFloat(m.Volume24hr, 64)
	liquidity, _ := strconv.ParseFloat(m.Liquidity, 64)

	// Parse end date.
	var endDate *time.Time
	if m.EndDate != "" {
		t, err := time.Parse(time.RFC3339, m.EndDate)
		if err != nil {
			log.Warn().
				Err(err).
				Str("end_date", m.EndDate).
				Str("condition_id", m.ConditionID).
				Msg("failed to parse end date")
		} else {
			endDate = &t
		}
	}

	// Optional string fields.
	description := nilIfEmpty(m.Description)
	category := nilIfEmpty(m.Category)
	imageURL := nilIfEmpty(m.Image)

	query := `
		INSERT INTO events (
			id, polymarket_event_id, slug, question, description, category, image_url,
			outcomes, outcome_prices, clob_token_ids, status, volume, volume_24h,
			liquidity, end_date, synced_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, 'open', $11, $12,
			$13, $14, NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			outcome_prices = EXCLUDED.outcome_prices,
			volume = EXCLUDED.volume,
			volume_24h = EXCLUDED.volume_24h,
			liquidity = EXCLUDED.liquidity,
			synced_at = NOW(),
			updated_at = NOW()
		RETURNING (xmax = 0) AS is_new
	`

	var isNew bool
	err = s.pool.QueryRow(ctx, query,
		m.ConditionID,   // $1  id
		m.EventID,       // $2  polymarket_event_id
		m.Slug,          // $3  slug
		m.Question,      // $4  question
		description,     // $5  description
		category,        // $6  category
		imageURL,        // $7  image_url
		outcomesJSON,    // $8  outcomes
		pricesJSON,      // $9  outcome_prices
		tokenIDsJSON,    // $10 clob_token_ids
		volume,          // $11 volume
		volume24h,       // $12 volume_24h
		liquidity,       // $13 liquidity
		endDate,         // $14 end_date
	).Scan(&isNew)
	if err != nil {
		return false, fmt.Errorf("upserting event %s: %w", m.ConditionID, err)
	}

	return isNew, nil
}

// recordPriceHistory fetches midpoints for each outcome's CLOB token and
// inserts a price_history row.
func (s *Syncer) recordPriceHistory(ctx context.Context, m polymarket.GammaMarket) error {
	if len(m.ClobTokenIDs) == 0 || len(m.Outcomes) == 0 {
		return nil
	}

	midpoints, err := s.clob.GetMidpoints(ctx, m.ClobTokenIDs)
	if err != nil {
		return fmt.Errorf("fetching midpoints: %w", err)
	}

	query := `
		INSERT INTO price_history (event_id, outcome_label, price, recorded_at)
		VALUES ($1, $2, $3, NOW())
	`

	for i, tokenID := range m.ClobTokenIDs {
		if tokenID == "" {
			continue
		}

		mid, ok := midpoints[tokenID]
		if !ok {
			continue
		}

		outcomeLabel := "Unknown"
		if i < len(m.Outcomes) {
			outcomeLabel = m.Outcomes[i]
		}

		_, err := s.pool.Exec(ctx, query, m.ConditionID, outcomeLabel, mid)
		if err != nil {
			log.Error().
				Err(err).
				Str("event_id", m.ConditionID).
				Str("outcome", outcomeLabel).
				Float64("price", mid).
				Msg("failed to insert price history row")
		}
	}

	return nil
}

// detectResolutions checks if any markets have been resolved by looking for
// outcome prices where one is "1" and another is "0". It updates their status
// to 'resolved' in the database.
func (s *Syncer) detectResolutions(ctx context.Context, markets []polymarket.GammaMarket) int {
	resolved := 0

	for _, m := range markets {
		if m.ConditionID == "" || len(m.OutcomePrices) == 0 || len(m.Outcomes) == 0 {
			continue
		}

		// Check if there is an outcome with price "1" (the winner).
		winnerIdx := -1
		hasZero := false
		for i, price := range m.OutcomePrices {
			if price == "1" {
				winnerIdx = i
			}
			if price == "0" {
				hasZero = true
			}
		}

		if winnerIdx < 0 || !hasZero {
			continue
		}

		winnerOutcome := "Unknown"
		if winnerIdx < len(m.Outcomes) {
			winnerOutcome = m.Outcomes[winnerIdx]
		}

		query := `
			UPDATE events
			SET status = 'resolved',
				resolved_outcome = $2,
				resolved_at = NOW(),
				updated_at = NOW()
			WHERE id = $1
			  AND status = 'open'
		`

		tag, err := s.pool.Exec(ctx, query, m.ConditionID, winnerOutcome)
		if err != nil {
			log.Error().
				Err(err).
				Str("condition_id", m.ConditionID).
				Str("winner", winnerOutcome).
				Msg("failed to update resolved event")
			continue
		}

		if tag.RowsAffected() > 0 {
			resolved++
			log.Info().
				Str("condition_id", m.ConditionID).
				Str("resolved_outcome", winnerOutcome).
				Msg("event resolved")
		}
	}

	return resolved
}

// nilIfEmpty returns a pointer to s if s is non-empty, otherwise nil.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
