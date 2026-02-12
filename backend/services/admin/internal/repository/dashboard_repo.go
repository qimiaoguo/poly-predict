package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// DashboardStats holds aggregate statistics for the admin dashboard.
type DashboardStats struct {
	TotalUsers   int64      `json:"total_users"`
	TotalBets    int64      `json:"total_bets"`
	ActiveEvents int64      `json:"active_events"`
	TotalVolume  int64      `json:"total_volume"`
	RecentBets   []model.Bet `json:"recent_bets"`
}

// DashboardRepository handles database operations for admin dashboard data.
type DashboardRepository struct {
	pool *pgxpool.Pool
}

// NewDashboardRepository creates a new DashboardRepository.
func NewDashboardRepository(pool *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{pool: pool}
}

// GetStats returns aggregate statistics and the 10 most recent bets.
func (r *DashboardRepository) GetStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Total users.
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Total bets.
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM bets`).Scan(&stats.TotalBets); err != nil {
		return nil, fmt.Errorf("failed to count bets: %w", err)
	}

	// Active (open) events.
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE status = 'open'`).Scan(&stats.ActiveEvents); err != nil {
		return nil, fmt.Errorf("failed to count active events: %w", err)
	}

	// Total volume (sum of bet amounts).
	if err := r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM bets`).Scan(&stats.TotalVolume); err != nil {
		return nil, fmt.Errorf("failed to sum bet volume: %w", err)
	}

	// Recent 10 bets.
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, event_id, outcome, amount, locked_odds,
			potential_payout, status, payout, settled_at, created_at
		 FROM bets
		 ORDER BY created_at DESC
		 LIMIT 10`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent bets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b model.Bet
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.EventID, &b.Outcome, &b.Amount, &b.LockedOdds,
			&b.PotentialPayout, &b.Status, &b.Payout, &b.SettledAt, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bet: %w", err)
		}
		stats.RecentBets = append(stats.RecentBets, b)
	}

	if stats.RecentBets == nil {
		stats.RecentBets = []model.Bet{}
	}

	return stats, nil
}
