package settler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/poly-predict/backend/pkg/model"
)

// Settler performs periodic settlement of resolved prediction-market events.
type Settler struct {
	pool *pgxpool.Pool
}

// New creates a new Settler backed by the given connection pool.
func New(pool *pgxpool.Pool) *Settler {
	return &Settler{pool: pool}
}

// Run executes a single settlement cycle: find all resolved-but-unsettled
// events, settle each one, then recalculate the global rankings.
func (s *Settler) Run(ctx context.Context) error {
	start := time.Now()
	log.Info().Msg("settlement cycle started")

	// 1. Find resolved events that have not been settled yet.
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.polymarket_event_id, e.slug, e.question, e.description,
		       e.category, e.image_url, e.outcomes, e.outcome_prices,
		       e.clob_token_ids, e.status, e.resolved_outcome, e.resolved_at,
		       e.volume, e.volume_24h, e.liquidity, e.end_date,
		       e.created_at, e.updated_at, e.synced_at
		FROM events e
		WHERE e.status = 'resolved'
		  AND e.resolved_outcome IS NOT NULL
		  AND NOT EXISTS (SELECT 1 FROM settlements s WHERE s.event_id = e.id)
	`)
	if err != nil {
		return fmt.Errorf("query unsettled events: %w", err)
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		var ev model.Event
		if err := rows.Scan(
			&ev.ID, &ev.PolymarketEventID, &ev.Slug, &ev.Question, &ev.Description,
			&ev.Category, &ev.ImageURL, &ev.Outcomes, &ev.OutcomePrices,
			&ev.ClobTokenIDs, &ev.Status, &ev.ResolvedOutcome, &ev.ResolvedAt,
			&ev.Volume, &ev.Volume24h, &ev.Liquidity, &ev.EndDate,
			&ev.CreatedAt, &ev.UpdatedAt, &ev.SyncedAt,
		); err != nil {
			return fmt.Errorf("scan event row: %w", err)
		}
		events = append(events, &ev)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate event rows: %w", err)
	}

	if len(events) == 0 {
		log.Info().Dur("elapsed", time.Since(start)).Msg("no unsettled events found")
		return nil
	}

	log.Info().Int("count", len(events)).Msg("found unsettled events")

	// 2. Settle each event independently; one failure must not block the rest.
	settled := 0
	for _, ev := range events {
		if err := s.settleEvent(ctx, ev); err != nil {
			log.Error().Err(err).Str("event_id", ev.ID).Msg("failed to settle event")
			continue
		}
		settled++
	}

	// 3. Recalculate global rankings.
	if err := s.recalculateRankings(ctx); err != nil {
		log.Error().Err(err).Msg("failed to recalculate rankings")
	}

	log.Info().
		Int("total_events", len(events)).
		Int("settled", settled).
		Dur("elapsed", time.Since(start)).
		Msg("settlement cycle completed")

	return nil
}

// settleEvent settles a single resolved event inside one atomic transaction.
// It is idempotent: if a settlement record already exists the call is a no-op.
func (s *Settler) settleEvent(ctx context.Context, event *model.Event) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// Ensure rollback on any non-committed exit path.
	defer tx.Rollback(ctx) //nolint:errcheck

	// Idempotency check: bail out if a settlement already exists.
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM settlements WHERE event_id = $1)`, event.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if exists {
		log.Debug().Str("event_id", event.ID).Msg("event already settled, skipping")
		return nil
	}

	// Lock all pending bets for this event to prevent concurrent modifications.
	betRows, err := tx.Query(ctx, `
		SELECT id, user_id, event_id, outcome, amount, locked_odds,
		       potential_payout, status, payout, settled_at, created_at
		FROM bets
		WHERE event_id = $1 AND status = 'pending'
		FOR UPDATE
	`, event.ID)
	if err != nil {
		return fmt.Errorf("lock bets: %w", err)
	}
	defer betRows.Close()

	var bets []*model.Bet
	for betRows.Next() {
		var b model.Bet
		if err := betRows.Scan(
			&b.ID, &b.UserID, &b.EventID, &b.Outcome, &b.Amount,
			&b.LockedOdds, &b.PotentialPayout, &b.Status, &b.Payout,
			&b.SettledAt, &b.CreatedAt,
		); err != nil {
			return fmt.Errorf("scan bet row: %w", err)
		}
		bets = append(bets, &b)
	}
	if err := betRows.Err(); err != nil {
		return fmt.Errorf("iterate bet rows: %w", err)
	}

	resolvedOutcome := ""
	if event.ResolvedOutcome != nil {
		resolvedOutcome = *event.ResolvedOutcome
	}

	var totalPayouts int64
	betCount := len(bets)

	for _, bet := range bets {
		won := strings.EqualFold(bet.Outcome, resolvedOutcome)

		if won {
			// --- Winner path ---
			if err := settleWinningBet(ctx, tx, bet, event); err != nil {
				return fmt.Errorf("settle winning bet %s: %w", bet.ID, err)
			}
			totalPayouts += bet.PotentialPayout
		} else {
			// --- Loser path ---
			if err := settleLosingBet(ctx, tx, bet, event); err != nil {
				return fmt.Errorf("settle losing bet %s: %w", bet.ID, err)
			}
		}
	}

	// Record the settlement.
	_, err = tx.Exec(ctx, `
		INSERT INTO settlements (event_id, resolved_outcome, total_bets, total_payouts)
		VALUES ($1, $2, $3, $4)
	`, event.ID, resolvedOutcome, betCount, totalPayouts)
	if err != nil {
		return fmt.Errorf("insert settlement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Info().
		Str("event_id", event.ID).
		Str("question", event.Question).
		Str("resolved_outcome", resolvedOutcome).
		Int("bet_count", betCount).
		Int64("total_payouts", totalPayouts).
		Msg("event settled")

	return nil
}

// settleWinningBet marks a bet as won, credits the user, records a
// credit_transaction, and updates the user's streak.
func settleWinningBet(ctx context.Context, tx pgx.Tx, bet *model.Bet, event *model.Event) error {
	// Mark the bet as won.
	_, err := tx.Exec(ctx, `
		UPDATE bets SET status = 'won', payout = $1, settled_at = NOW() WHERE id = $2
	`, bet.PotentialPayout, bet.ID)
	if err != nil {
		return fmt.Errorf("update bet: %w", err)
	}

	// Credit the user: release frozen balance, add winnings, bump counters.
	var newBalance int64
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET frozen_balance = frozen_balance - $1,
		    balance         = balance + $2,
		    total_wins      = total_wins + 1,
		    total_bets      = total_bets + 1,
		    current_streak  = current_streak + 1,
		    max_streak      = GREATEST(max_streak, current_streak + 1),
		    updated_at      = NOW()
		WHERE id = $3
		RETURNING balance
	`, bet.Amount, bet.PotentialPayout, bet.UserID).Scan(&newBalance)
	if err != nil {
		return fmt.Errorf("update user balance: %w", err)
	}

	// Ledger entry.
	desc := "Won bet on: " + event.Question
	_, err = tx.Exec(ctx, `
		INSERT INTO credit_transactions (user_id, type, amount, balance_after, reference_id, description)
		VALUES ($1, 'bet_won', $2, $3, $4, $5)
	`, bet.UserID, bet.PotentialPayout, newBalance, bet.ID, desc)
	if err != nil {
		return fmt.Errorf("insert credit_transaction: %w", err)
	}

	return nil
}

// settleLosingBet marks a bet as lost, adjusts the user's frozen balance,
// records a credit_transaction, and resets the user's streak.
func settleLosingBet(ctx context.Context, tx pgx.Tx, bet *model.Bet, event *model.Event) error {
	// Mark the bet as lost.
	_, err := tx.Exec(ctx, `
		UPDATE bets SET status = 'lost', payout = 0, settled_at = NOW() WHERE id = $1
	`, bet.ID)
	if err != nil {
		return fmt.Errorf("update bet: %w", err)
	}

	// Deduct frozen balance, bump total_bets, reset streak.
	var currentBalance int64
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET frozen_balance = frozen_balance - $1,
		    total_bets     = total_bets + 1,
		    current_streak = 0,
		    updated_at     = NOW()
		WHERE id = $2
		RETURNING balance
	`, bet.Amount, bet.UserID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("update user balance: %w", err)
	}

	// Ledger entry.
	desc := "Lost bet on: " + event.Question
	_, err = tx.Exec(ctx, `
		INSERT INTO credit_transactions (user_id, type, amount, balance_after, reference_id, description)
		VALUES ($1, 'bet_lost', 0, $2, $3, $4)
	`, bet.UserID, currentBalance, bet.ID, desc)
	if err != nil {
		return fmt.Errorf("insert credit_transaction: %w", err)
	}

	return nil
}

// recalculateRankings rebuilds the all_time rankings table from current user
// stats. It deletes old rows and reinserts in a single statement.
func (s *Settler) recalculateRankings(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM rankings WHERE period = 'all_time' AND category IS NULL;

		INSERT INTO rankings (
			user_id, period, total_assets, total_profit,
			win_count, loss_count, win_rate, roi,
			consecutive_wins, rank_position
		)
		SELECT
			u.id,
			'all_time',
			u.balance + u.frozen_balance,
			(u.balance + u.frozen_balance) - 1000000,
			u.total_wins,
			u.total_bets - u.total_wins,
			CASE WHEN u.total_bets > 0 THEN u.total_wins::numeric / u.total_bets ELSE 0 END,
			CASE WHEN u.total_bets > 0 THEN ((u.balance + u.frozen_balance) - 1000000)::numeric / 1000000 ELSE 0 END,
			u.current_streak,
			ROW_NUMBER() OVER (ORDER BY (u.balance + u.frozen_balance) DESC)
		FROM users u
		WHERE u.total_bets > 0;
	`)
	if err != nil {
		return fmt.Errorf("recalculate rankings: %w", err)
	}

	log.Info().Msg("rankings recalculated")
	return nil
}
