package model

import (
	"encoding/json"
	"time"
)

// EventStatus represents the status of an event.
type EventStatus string

const (
	EventStatusOpen     EventStatus = "open"
	EventStatusClosed   EventStatus = "closed"
	EventStatusResolved EventStatus = "resolved"
)

// Event represents a prediction market event.
type Event struct {
	ID                string          `json:"id" db:"id"`
	PolymarketEventID string          `json:"polymarket_event_id" db:"polymarket_event_id"`
	Slug              string          `json:"slug" db:"slug"`
	Question          string          `json:"question" db:"question"`
	Description       *string         `json:"description" db:"description"`
	Category          *string         `json:"category" db:"category"`
	ImageURL          *string         `json:"image_url" db:"image_url"`
	Outcomes          json.RawMessage `json:"outcomes" db:"outcomes"`
	OutcomePrices     json.RawMessage `json:"outcome_prices" db:"outcome_prices"`
	ClobTokenIDs      json.RawMessage `json:"clob_token_ids" db:"clob_token_ids"`
	Status            EventStatus     `json:"status" db:"status"`
	ResolvedOutcome   *string         `json:"resolved_outcome" db:"resolved_outcome"`
	ResolvedAt        *time.Time      `json:"resolved_at" db:"resolved_at"`
	Volume            float64         `json:"volume" db:"volume"`
	Volume24h         float64         `json:"volume_24h" db:"volume_24h"`
	Liquidity         float64         `json:"liquidity" db:"liquidity"`
	EndDate           *time.Time      `json:"end_date" db:"end_date"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
	SyncedAt          time.Time       `json:"synced_at" db:"synced_at"`
}

// PriceHistory represents a historical price point for an event outcome.
type PriceHistory struct {
	ID           int64     `json:"id" db:"id"`
	EventID      string    `json:"event_id" db:"event_id"`
	OutcomeLabel string    `json:"outcome_label" db:"outcome_label"`
	Price        float64   `json:"price" db:"price"`
	RecordedAt   time.Time `json:"recorded_at" db:"recorded_at"`
}
