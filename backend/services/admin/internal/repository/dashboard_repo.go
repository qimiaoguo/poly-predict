package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// RecentBet extends model.Bet with joined user/event info for the dashboard.
type RecentBet struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	EventID         string          `json:"event_id"`
	Outcome         string          `json:"outcome"`
	Amount          int64           `json:"amount"`
	LockedOdds      float64         `json:"locked_odds"`
	PotentialPayout int64           `json:"potential_payout"`
	Status          model.BetStatus `json:"status"`
	Payout          *int64          `json:"payout"`
	SettledAt       *time.Time      `json:"settled_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UserDisplayName string          `json:"user_display_name"`
	EventQuestion   string          `json:"event_question"`
}

// DashboardStats holds aggregate statistics for the admin dashboard.
type DashboardStats struct {
	TotalUsers   int64       `json:"total_users"`
	TotalBets    int64       `json:"total_bets"`
	ActiveEvents int64       `json:"active_events"`
	TotalVolume  int64       `json:"total_volume"`
	RecentBets   []RecentBet `json:"recent_bets"`
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

	// Recent 10 bets with user display name and event question.
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.user_id, b.event_id, b.outcome, b.amount, b.locked_odds,
			b.potential_payout, b.status, b.payout, b.settled_at, b.created_at,
			COALESCE(u.display_name, 'Unknown') AS user_display_name,
			COALESCE(e.question, 'Unknown') AS event_question
		 FROM bets b
		 LEFT JOIN users u ON u.id = b.user_id
		 LEFT JOIN events e ON e.id = b.event_id
		 ORDER BY b.created_at DESC
		 LIMIT 10`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent bets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rb RecentBet
		if err := rows.Scan(
			&rb.ID, &rb.UserID, &rb.EventID, &rb.Outcome, &rb.Amount, &rb.LockedOdds,
			&rb.PotentialPayout, &rb.Status, &rb.Payout, &rb.SettledAt, &rb.CreatedAt,
			&rb.UserDisplayName, &rb.EventQuestion,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bet: %w", err)
		}
		stats.RecentBets = append(stats.RecentBets, rb)
	}

	if stats.RecentBets == nil {
		stats.RecentBets = []RecentBet{}
	}

	return stats, nil
}
