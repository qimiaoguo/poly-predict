package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// EventRepository handles database operations for events.
type EventRepository struct {
	pool *pgxpool.Pool
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

// List returns a paginated list of events with optional filters.
func (r *EventRepository) List(ctx context.Context, status, category, search string, page, pageSize int) ([]model.Event, int64, error) {
	offset := (page - 1) * pageSize

	var conditions []string
	var args []interface{}
	argIdx := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("question ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total.
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events %s", whereClause)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Fetch page.
	dataQuery := fmt.Sprintf(
		`SELECT id, polymarket_event_id, slug, question, description, category,
			image_url, outcomes, outcome_prices, clob_token_ids, status,
			resolved_outcome, resolved_at, volume, volume_24h, liquidity,
			end_date, created_at, updated_at, synced_at
		FROM events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(
			&e.ID, &e.PolymarketEventID, &e.Slug, &e.Question, &e.Description, &e.Category,
			&e.ImageURL, &e.Outcomes, &e.OutcomePrices, &e.ClobTokenIDs, &e.Status,
			&e.ResolvedOutcome, &e.ResolvedAt, &e.Volume, &e.Volume24h, &e.Liquidity,
			&e.EndDate, &e.CreatedAt, &e.UpdatedAt, &e.SyncedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	if events == nil {
		events = []model.Event{}
	}

	return events, total, nil
}

// GetByID returns a single event by ID.
func (r *EventRepository) GetByID(ctx context.Context, id string) (*model.Event, error) {
	e := &model.Event{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, polymarket_event_id, slug, question, description, category,
			image_url, outcomes, outcome_prices, clob_token_ids, status,
			resolved_outcome, resolved_at, volume, volume_24h, liquidity,
			end_date, created_at, updated_at, synced_at
		FROM events
		WHERE id = $1`, id,
	).Scan(
		&e.ID, &e.PolymarketEventID, &e.Slug, &e.Question, &e.Description, &e.Category,
		&e.ImageURL, &e.Outcomes, &e.OutcomePrices, &e.ClobTokenIDs, &e.Status,
		&e.ResolvedOutcome, &e.ResolvedAt, &e.Volume, &e.Volume24h, &e.Liquidity,
		&e.EndDate, &e.CreatedAt, &e.UpdatedAt, &e.SyncedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}
	return e, nil
}

// Update performs a partial update of an event's category and/or status.
func (r *EventRepository) Update(ctx context.Context, id string, category, status *string) (*model.Event, error) {
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *category)
		argIdx++
	}

	if status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *status)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf(
		`UPDATE events SET %s WHERE id = $%d
		 RETURNING id, polymarket_event_id, slug, question, description, category,
			image_url, outcomes, outcome_prices, clob_token_ids, status,
			resolved_outcome, resolved_at, volume, volume_24h, liquidity,
			end_date, created_at, updated_at, synced_at`,
		strings.Join(setClauses, ", "), argIdx,
	)
	args = append(args, id)

	e := &model.Event{}
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&e.ID, &e.PolymarketEventID, &e.Slug, &e.Question, &e.Description, &e.Category,
		&e.ImageURL, &e.Outcomes, &e.OutcomePrices, &e.ClobTokenIDs, &e.Status,
		&e.ResolvedOutcome, &e.ResolvedAt, &e.Volume, &e.Volume24h, &e.Liquidity,
		&e.EndDate, &e.CreatedAt, &e.UpdatedAt, &e.SyncedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return e, nil
}
