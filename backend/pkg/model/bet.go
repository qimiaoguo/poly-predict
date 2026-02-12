package model

import "time"

// BetStatus represents the status of a bet.
type BetStatus string

const (
	BetStatusPending   BetStatus = "pending"
	BetStatusWon       BetStatus = "won"
	BetStatusLost      BetStatus = "lost"
	BetStatusCancelled BetStatus = "cancelled"
)

// Bet represents a user's wager on an event outcome.
type Bet struct {
	ID              string     `json:"id" db:"id"`
	UserID          string     `json:"user_id" db:"user_id"`
	EventID         string     `json:"event_id" db:"event_id"`
	Outcome         string     `json:"outcome" db:"outcome"`
	Amount          int64      `json:"amount" db:"amount"`
	LockedOdds      float64    `json:"locked_odds" db:"locked_odds"`
	PotentialPayout int64      `json:"potential_payout" db:"potential_payout"`
	Status          BetStatus  `json:"status" db:"status"`
	Payout          *int64     `json:"payout" db:"payout"`
	SettledAt       *time.Time `json:"settled_at" db:"settled_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// Settlement represents the settlement record for an event.
type Settlement struct {
	ID              string    `json:"id" db:"id"`
	EventID         string    `json:"event_id" db:"event_id"`
	ResolvedOutcome string    `json:"resolved_outcome" db:"resolved_outcome"`
	TotalBets       int       `json:"total_bets" db:"total_bets"`
	TotalPayouts    int64     `json:"total_payouts" db:"total_payouts"`
	SettledAt       time.Time `json:"settled_at" db:"settled_at"`
}

// CreditTransaction represents a ledger entry for credit movements.
type CreditTransaction struct {
	ID           int64     `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Type         string    `json:"type" db:"type"`
	Amount       int64     `json:"amount" db:"amount"`
	BalanceAfter int64     `json:"balance_after" db:"balance_after"`
	ReferenceID  *string   `json:"reference_id" db:"reference_id"`
	Description  *string   `json:"description" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Ranking represents a user's ranking entry for a given period and category.
type Ranking struct {
	ID               int64     `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	Period           string    `json:"period" db:"period"`
	Category         string    `json:"category" db:"category"`
	TotalAssets      int64     `json:"total_assets" db:"total_assets"`
	TotalProfit      int64     `json:"total_profit" db:"total_profit"`
	WinCount         int       `json:"win_count" db:"win_count"`
	LossCount        int       `json:"loss_count" db:"loss_count"`
	WinRate          float64   `json:"win_rate" db:"win_rate"`
	ROI              float64   `json:"roi" db:"roi"`
	ConsecutiveWins  int       `json:"consecutive_wins" db:"consecutive_wins"`
	RankPosition     int       `json:"rank_position" db:"rank_position"`
	CalculatedAt     time.Time `json:"calculated_at" db:"calculated_at"`
}
