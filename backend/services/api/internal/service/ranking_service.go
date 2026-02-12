package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RankingEntry is the enriched ranking returned by the API.
type RankingEntry struct {
	Rank        int       `json:"rank"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	TotalProfit int64     `json:"total_profit"`
	WinRate     float64   `json:"win_rate"`
	ROI         float64   `json:"roi"`
	TotalBets   int       `json:"total_bets"`
	WinCount    int       `json:"win_count"`
	LossCount   int       `json:"loss_count"`
	CalculatedAt time.Time `json:"calculated_at"`
}

// RankingService handles ranking queries.
type RankingService struct {
	pool *pgxpool.Pool
}

// NewRankingService creates a new RankingService.
func NewRankingService(pool *pgxpool.Pool) *RankingService {
	return &RankingService{pool: pool}
}

// GetRankings retrieves a paginated list of rankings with optional filters.
func (s *RankingService) GetRankings(ctx context.Context, period, category, sortBy string, page, pageSize int) ([]RankingEntry, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if period != "" {
		whereClause += fmt.Sprintf(" AND r.period = $%d", argIdx)
		args = append(args, period)
		argIdx++
	}

	if category != "" {
		whereClause += fmt.Sprintf(" AND r.category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	// Count.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rankings r %s", whereClause)
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count rankings: %w", err)
	}

	// Sort.
	orderClause := "ORDER BY r.rank_position ASC"
	switch sortBy {
	case "total_assets":
		orderClause = "ORDER BY r.total_assets DESC"
	case "total_profit":
		orderClause = "ORDER BY r.total_profit DESC"
	case "win_rate":
		orderClause = "ORDER BY r.win_rate DESC"
	case "roi":
		orderClause = "ORDER BY r.roi DESC"
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(
		`SELECT r.rank_position, r.user_id, COALESCE(u.display_name, 'Unknown'),
		        u.avatar_url, r.total_profit, r.win_rate, r.roi,
		        r.win_count, r.loss_count, r.win_count + r.loss_count, r.calculated_at
		 FROM rankings r
		 LEFT JOIN users u ON u.id = r.user_id
		 %s %s LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := s.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list rankings: %w", err)
	}
	defer rows.Close()

	var rankings []RankingEntry
	for rows.Next() {
		var r RankingEntry
		var totalBets int
		err := rows.Scan(
			&r.Rank, &r.UserID, &r.DisplayName,
			&r.AvatarURL, &r.TotalProfit, &r.WinRate, &r.ROI,
			&r.WinCount, &r.LossCount, &totalBets, &r.CalculatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan ranking: %w", err)
		}
		r.TotalBets = totalBets
		rankings = append(rankings, r)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rankings: %w", err)
	}

	return rankings, total, nil
}
