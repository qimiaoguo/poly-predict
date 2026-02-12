package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// SettlementRepository handles database operations for settlements.
type SettlementRepository struct {
	pool *pgxpool.Pool
}

// NewSettlementRepository creates a new SettlementRepository.
func NewSettlementRepository(pool *pgxpool.Pool) *SettlementRepository {
	return &SettlementRepository{pool: pool}
}

// ForceSettle atomically settles an event with the given outcome.
// It updates the event, resolves all pending bets, adjusts user balances,
// logs credit transactions, and inserts a settlement record -- all within
// a single database transaction.
func (r *SettlementRepository) ForceSettle(ctx context.Context, eventID, outcome string) (*model.Settlement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Check event exists and is not already resolved.
	var currentStatus model.EventStatus
	err = tx.QueryRow(ctx,
		`SELECT status FROM events WHERE id = $1 FOR UPDATE`, eventID,
	).Scan(&currentStatus)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}
	if currentStatus == model.EventStatusResolved {
		return nil, fmt.Errorf("event is already resolved")
	}

	// 2. Update event to resolved.
	now := time.Now()
	_, err = tx.Exec(ctx,
		`UPDATE events
		 SET status = 'resolved', resolved_outcome = $1, resolved_at = $2, updated_at = $2
		 WHERE id = $3`,
		outcome, now, eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update event status: %w", err)
	}

	// 3. Get all pending bets for this event (lock rows).
	rows, err := tx.Query(ctx,
		`SELECT id, user_id, outcome, amount, potential_payout
		 FROM bets
		 WHERE event_id = $1 AND status = 'pending'
		 FOR UPDATE`, eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending bets: %w", err)
	}

	type betRow struct {
		ID              string
		UserID          string
		Outcome         string
		Amount          int64
		PotentialPayout int64
	}

	var bets []betRow
	for rows.Next() {
		var b betRow
		if err := rows.Scan(&b.ID, &b.UserID, &b.Outcome, &b.Amount, &b.PotentialPayout); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan bet: %w", err)
		}
		bets = append(bets, b)
	}
	rows.Close()

	// 4. Resolve each bet.
	var totalPayouts int64
	for _, b := range bets {
		if b.Outcome == outcome {
			// Winner: set status=won, payout=potential_payout, credit user balance.
			payout := b.PotentialPayout
			totalPayouts += payout

			_, err = tx.Exec(ctx,
				`UPDATE bets SET status = 'won', payout = $1, settled_at = $2 WHERE id = $3`,
				payout, now, b.ID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update winning bet: %w", err)
			}

			// Release frozen balance and add payout to available balance.
			var balanceAfter int64
			err = tx.QueryRow(ctx,
				`UPDATE users
				 SET balance = balance + $1,
				     frozen_balance = frozen_balance - $2,
				     total_wins = total_wins + 1,
				     updated_at = NOW()
				 WHERE id = $3
				 RETURNING balance`,
				payout, b.Amount, b.UserID,
			).Scan(&balanceAfter)
			if err != nil {
				return nil, fmt.Errorf("failed to credit winner balance: %w", err)
			}

			// Log credit transaction for the win.
			refID := b.ID
			desc := "Bet won - settlement payout"
			_, err = tx.Exec(ctx,
				`INSERT INTO credit_transactions (user_id, type, amount, balance_after, reference_id, description)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				b.UserID, "bet_win", payout, balanceAfter, refID, desc,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to log winner credit transaction: %w", err)
			}
		} else {
			// Loser: set status=lost, payout=0, release frozen balance.
			var zeroPayout int64
			_, err = tx.Exec(ctx,
				`UPDATE bets SET status = 'lost', payout = $1, settled_at = $2 WHERE id = $3`,
				zeroPayout, now, b.ID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update losing bet: %w", err)
			}

			// Release frozen balance (the amount was already deducted from available balance
			// when the bet was placed, so we just remove it from frozen).
			var balanceAfter int64
			err = tx.QueryRow(ctx,
				`UPDATE users
				 SET frozen_balance = frozen_balance - $1,
				     updated_at = NOW()
				 WHERE id = $2
				 RETURNING balance`,
				b.Amount, b.UserID,
			).Scan(&balanceAfter)
			if err != nil {
				return nil, fmt.Errorf("failed to release frozen balance: %w", err)
			}

			// Log credit transaction for the loss.
			refID := b.ID
			desc := "Bet lost - settlement"
			_, err = tx.Exec(ctx,
				`INSERT INTO credit_transactions (user_id, type, amount, balance_after, reference_id, description)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				b.UserID, "bet_loss", -b.Amount, balanceAfter, refID, desc,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to log loser credit transaction: %w", err)
			}
		}
	}

	// 5. Insert settlement record.
	settlement := &model.Settlement{}
	err = tx.QueryRow(ctx,
		`INSERT INTO settlements (event_id, resolved_outcome, total_bets, total_payouts, settled_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, event_id, resolved_outcome, total_bets, total_payouts, settled_at`,
		eventID, outcome, len(bets), totalPayouts, now,
	).Scan(
		&settlement.ID, &settlement.EventID, &settlement.ResolvedOutcome,
		&settlement.TotalBets, &settlement.TotalPayouts, &settlement.SettledAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert settlement record: %w", err)
	}

	// 6. Recalculate rankings (all_time, weekly, monthly).
	_, err = tx.Exec(ctx, `
		DELETE FROM rankings WHERE category IS NULL;

		-- All time rankings from user stats
		INSERT INTO rankings (
			user_id, period, total_assets, total_profit,
			win_count, loss_count, win_rate, roi,
			consecutive_wins, rank_position
		)
		SELECT
			u.id, 'all_time',
			u.balance + u.frozen_balance,
			(u.balance + u.frozen_balance) - 10000,
			u.total_wins,
			u.total_bets - u.total_wins,
			CASE WHEN u.total_bets > 0 THEN u.total_wins::numeric / u.total_bets ELSE 0 END,
			CASE WHEN u.total_bets > 0 THEN ((u.balance + u.frozen_balance) - 10000)::numeric / 10000 ELSE 0 END,
			u.current_streak,
			ROW_NUMBER() OVER (ORDER BY (u.balance + u.frozen_balance) DESC)
		FROM users u
		WHERE u.total_bets > 0;

		-- Weekly rankings from bets settled in the last 7 days
		INSERT INTO rankings (
			user_id, period, total_assets, total_profit,
			win_count, loss_count, win_rate, roi,
			consecutive_wins, rank_position
		)
		SELECT
			b.user_id, 'weekly',
			COALESCE(SUM(CASE WHEN b.status = 'won' THEN b.payout ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN b.status = 'won' THEN b.payout - b.amount ELSE -b.amount END), 0),
			COUNT(*) FILTER (WHERE b.status = 'won'),
			COUNT(*) FILTER (WHERE b.status = 'lost'),
			CASE WHEN COUNT(*) > 0 THEN COUNT(*) FILTER (WHERE b.status = 'won')::numeric / COUNT(*) ELSE 0 END,
			CASE WHEN SUM(b.amount) > 0 THEN
				SUM(CASE WHEN b.status = 'won' THEN b.payout - b.amount ELSE -b.amount END)::numeric / SUM(b.amount)
			ELSE 0 END,
			0,
			ROW_NUMBER() OVER (ORDER BY SUM(CASE WHEN b.status = 'won' THEN b.payout - b.amount ELSE -b.amount END) DESC)
		FROM bets b
		WHERE b.status IN ('won', 'lost') AND b.settled_at >= NOW() - INTERVAL '7 days'
		GROUP BY b.user_id;

		-- Monthly rankings from bets settled in the last 30 days
		INSERT INTO rankings (
			user_id, period, total_assets, total_profit,
			win_count, loss_count, win_rate, roi,
			consecutive_wins, rank_position
		)
		SELECT
			b.user_id, 'monthly',
			COALESCE(SUM(CASE WHEN b.status = 'won' THEN b.payout ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN b.status = 'won' THEN b.payout - b.amount ELSE -b.amount END), 0),
			COUNT(*) FILTER (WHERE b.status = 'won'),
			COUNT(*) FILTER (WHERE b.status = 'lost'),
			CASE WHEN COUNT(*) > 0 THEN COUNT(*) FILTER (WHERE b.status = 'won')::numeric / COUNT(*) ELSE 0 END,
			CASE WHEN SUM(b.amount) > 0 THEN
				SUM(CASE WHEN b.status = 'won' THEN b.payout - b.amount ELSE -b.amount END)::numeric / SUM(b.amount)
			ELSE 0 END,
			0,
			ROW_NUMBER() OVER (ORDER BY SUM(CASE WHEN b.status = 'won' THEN b.payout - b.amount ELSE -b.amount END) DESC)
		FROM bets b
		WHERE b.status IN ('won', 'lost') AND b.settled_at >= NOW() - INTERVAL '30 days'
		GROUP BY b.user_id;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to recalculate rankings: %w", err)
	}

	// 7. Commit transaction.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit settlement transaction: %w", err)
	}

	return settlement, nil
}

// List returns a paginated list of settlement records.
func (r *SettlementRepository) List(ctx context.Context, page, pageSize int) ([]model.Settlement, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM settlements`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count settlements: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, event_id, resolved_outcome, total_bets, total_payouts, settled_at
		 FROM settlements
		 ORDER BY settled_at DESC
		 LIMIT $1 OFFSET $2`, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query settlements: %w", err)
	}
	defer rows.Close()

	settlements, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Settlement, error) {
		var s model.Settlement
		err := row.Scan(&s.ID, &s.EventID, &s.ResolvedOutcome, &s.TotalBets, &s.TotalPayouts, &s.SettledAt)
		return s, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to collect settlements: %w", err)
	}

	if settlements == nil {
		settlements = []model.Settlement{}
	}

	return settlements, total, nil
}
