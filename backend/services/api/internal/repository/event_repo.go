package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// EventFilters holds the filter and pagination parameters for listing events.
type EventFilters struct {
	Status   string
	Category string
	Search   string
	Sort     string
	Page     int
	PageSize int
}

// CategoryCount holds a category name and its event count.
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// EventRepository provides database access for events.
type EventRepository struct {
	pool *pgxpool.Pool
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

// List retrieves a paginated list of events with optional filters.
func (r *EventRepository) List(ctx context.Context, filters EventFilters) ([]model.Event, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}

	if filters.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filters.Category)
		argIdx++
	}

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("question ILIKE $%d", argIdx))
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events %s", whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	// Sort clause.
	orderClause := "ORDER BY CASE WHEN end_date > NOW() THEN 0 ELSE 1 END, volume_24h DESC"
	switch filters.Sort {
	case "trending":
		orderClause = "ORDER BY CASE WHEN end_date > NOW() THEN 0 ELSE 1 END, volume_24h DESC"
	case "volume":
		orderClause = "ORDER BY volume DESC"
	case "volume_24h":
		orderClause = "ORDER BY volume_24h DESC"
	case "liquidity":
		orderClause = "ORDER BY liquidity DESC"
	case "newest":
		orderClause = "ORDER BY created_at DESC"
	case "ending_soon":
		orderClause = "ORDER BY CASE WHEN end_date > NOW() THEN 0 ELSE 1 END, end_date ASC NULLS LAST"
	}

	offset := (filters.Page - 1) * filters.PageSize
	dataQuery := fmt.Sprintf(
		`SELECT id, polymarket_event_id, slug, question, description, category, image_url,
		        outcomes, outcome_prices, clob_token_ids, status, resolved_outcome, resolved_at,
		        volume, volume_24h, liquidity, end_date, created_at, updated_at, synced_at
		 FROM events %s %s LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIdx, argIdx+1,
	)
	args = append(args, filters.PageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		err := rows.Scan(
			&e.ID, &e.PolymarketEventID, &e.Slug, &e.Question, &e.Description, &e.Category,
			&e.ImageURL, &e.Outcomes, &e.OutcomePrices, &e.ClobTokenIDs, &e.Status,
			&e.ResolvedOutcome, &e.ResolvedAt, &e.Volume, &e.Volume24h, &e.Liquidity,
			&e.EndDate, &e.CreatedAt, &e.UpdatedAt, &e.SyncedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate events: %w", err)
	}

	return events, total, nil
}

// GetByID retrieves a single event by its ID.
func (r *EventRepository) GetByID(ctx context.Context, id string) (*model.Event, error) {
	query := `SELECT id, polymarket_event_id, slug, question, description, category, image_url,
	                  outcomes, outcome_prices, clob_token_ids, status, resolved_outcome, resolved_at,
	                  volume, volume_24h, liquidity, end_date, created_at, updated_at, synced_at
	           FROM events WHERE id = $1`

	var e model.Event
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.PolymarketEventID, &e.Slug, &e.Question, &e.Description, &e.Category,
		&e.ImageURL, &e.Outcomes, &e.OutcomePrices, &e.ClobTokenIDs, &e.Status,
		&e.ResolvedOutcome, &e.ResolvedAt, &e.Volume, &e.Volume24h, &e.Liquidity,
		&e.EndDate, &e.CreatedAt, &e.UpdatedAt, &e.SyncedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get event by id: %w", err)
	}

	return &e, nil
}

// GetPriceHistory retrieves the price history for an event, filtered by period.
// Period can be: 1h, 6h, 24h, 7d, 30d.
func (r *EventRepository) GetPriceHistory(ctx context.Context, eventID string, period string) ([]model.PriceHistory, error) {
	interval := "24 hours" // default
	switch period {
	case "1h":
		interval = "1 hour"
	case "6h":
		interval = "6 hours"
	case "24h":
		interval = "24 hours"
	case "7d":
		interval = "7 days"
	case "30d":
		interval = "30 days"
	}

	query := `SELECT id, event_id, outcome_label, price, recorded_at
	          FROM price_history
	          WHERE event_id = $1 AND recorded_at >= NOW() - $2::interval
	          ORDER BY recorded_at ASC`

	rows, err := r.pool.Query(ctx, query, eventID, interval)
	if err != nil {
		return nil, fmt.Errorf("get price history: %w", err)
	}
	defer rows.Close()

	var history []model.PriceHistory
	for rows.Next() {
		var ph model.PriceHistory
		err := rows.Scan(&ph.ID, &ph.EventID, &ph.OutcomeLabel, &ph.Price, &ph.RecordedAt)
		if err != nil {
			return nil, fmt.Errorf("scan price history: %w", err)
		}
		history = append(history, ph)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate price history: %w", err)
	}

	return history, nil
}

// GetCategories retrieves all distinct categories with their event counts.
func (r *EventRepository) GetCategories(ctx context.Context) ([]CategoryCount, error) {
	query := `SELECT category, COUNT(*) as count
	          FROM events
	          WHERE category IS NOT NULL
	          GROUP BY category
	          ORDER BY count DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	defer rows.Close()

	var categories []CategoryCount
	for rows.Next() {
		var cc CategoryCount
		err := rows.Scan(&cc.Category, &cc.Count)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, cc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}
