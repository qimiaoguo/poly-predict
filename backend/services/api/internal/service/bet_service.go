package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
	"github.com/poly-predict/backend/services/api/internal/repository"
)

// BetService handles bet business logic.
type BetService struct {
	pool     *pgxpool.Pool
	betRepo  *repository.BetRepository
	userRepo *repository.UserRepository
}

// NewBetService creates a new BetService.
func NewBetService(pool *pgxpool.Pool, betRepo *repository.BetRepository, userRepo *repository.UserRepository) *BetService {
	return &BetService{
		pool:     pool,
		betRepo:  betRepo,
		userRepo: userRepo,
	}
}

// PlaceBet creates a new bet atomically within a database transaction.
// It verifies the user has sufficient balance, the event is open, and locks odds at the time of placement.
func (s *BetService) PlaceBet(ctx context.Context, userID, eventID, outcome string, amount int64) (*model.Bet, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock user row and check balance.
	var balance int64
	err = tx.QueryRow(ctx, "SELECT balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("get user balance: %w", err)
	}

	if balance < amount {
		return nil, fmt.Errorf("insufficient balance: have %d, need %d", balance, amount)
	}

	// 2. Get event and verify it is open.
	var status model.EventStatus
	var outcomes json.RawMessage
	var outcomePrices json.RawMessage
	err = tx.QueryRow(ctx,
		"SELECT status, outcomes, outcome_prices FROM events WHERE id = $1",
		eventID,
	).Scan(&status, &outcomes, &outcomePrices)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	if status != model.EventStatusOpen {
		return nil, fmt.Errorf("event is not open for betting")
	}

	// 3. Parse outcomes and prices to find the odds for the chosen outcome.
	var outcomeLabels []string
	if err := json.Unmarshal(outcomes, &outcomeLabels); err != nil {
		return nil, fmt.Errorf("parse outcomes: %w", err)
	}

	var prices []string
	if err := json.Unmarshal(outcomePrices, &prices); err != nil {
		return nil, fmt.Errorf("parse outcome prices: %w", err)
	}

	outcomeIdx := findOutcomeIndex(outcomeLabels, outcome)
	if outcomeIdx == -1 || outcomeIdx >= len(prices) {
		return nil, fmt.Errorf("invalid outcome: %s", outcome)
	}

	// Parse the price string to float64.
	var lockedOdds float64
	_, err = fmt.Sscanf(prices[outcomeIdx], "%f", &lockedOdds)
	if err != nil || lockedOdds <= 0 {
		return nil, fmt.Errorf("invalid odds for outcome: %s", outcome)
	}

	// 4. Calculate potential payout: amount / lockedOdds (as int64).
	potentialPayout := int64(float64(amount) / lockedOdds)

	// 5. Update user balances and increment total_bets.
	_, err = tx.Exec(ctx,
		"UPDATE users SET balance = balance - $1, frozen_balance = frozen_balance + $1, total_bets = total_bets + 1, updated_at = NOW() WHERE id = $2",
		amount, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("update user balance: %w", err)
	}

	// 6. Insert bet.
	betID := uuid.New().String()
	now := time.Now()
	bet := &model.Bet{
		ID:              betID,
		UserID:          userID,
		EventID:         eventID,
		Outcome:         outcome,
		Amount:          amount,
		LockedOdds:      lockedOdds,
		PotentialPayout: potentialPayout,
		Status:          model.BetStatusPending,
		CreatedAt:       now,
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO bets (id, user_id, event_id, outcome, amount, locked_odds, potential_payout, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		bet.ID, bet.UserID, bet.EventID, bet.Outcome, bet.Amount,
		bet.LockedOdds, bet.PotentialPayout, bet.Status, bet.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert bet: %w", err)
	}

	// 7. Insert credit transaction.
	newBalance := balance - amount
	desc := fmt.Sprintf("Bet placed on event %s", eventID)
	_, err = tx.Exec(ctx,
		`INSERT INTO credit_transactions (user_id, type, amount, balance_after, reference_id, description, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, "bet_placed", -amount, newBalance, &betID, &desc, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert credit transaction: %w", err)
	}

	// 8. Commit.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return bet, nil
}

// ListByUser retrieves a paginated list of bets for the given user.
func (s *BetService) ListByUser(ctx context.Context, userID string, status string, page, pageSize int) ([]model.Bet, int64, error) {
	return s.betRepo.ListByUser(ctx, userID, status, page, pageSize)
}

// GetByID retrieves a single bet by ID.
func (s *BetService) GetByID(ctx context.Context, id string) (*model.Bet, error) {
	return s.betRepo.GetByID(ctx, id)
}

// findOutcomeIndex returns the index of outcome in labels using case-insensitive
// comparison, or -1 if not found. This allows the API to accept lowercase outcomes
// (as defined in the spec) while the database stores capitalized labels from Polymarket.
func findOutcomeIndex(labels []string, outcome string) int {
	for i, label := range labels {
		if strings.EqualFold(label, outcome) {
			return i
		}
	}
	return -1
}
