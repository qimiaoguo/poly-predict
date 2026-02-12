package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// RankingService handles ranking queries.
type RankingService struct {
	pool *pgxpool.Pool
}

// NewRankingService creates a new RankingService.
func NewRankingService(pool *pgxpool.Pool) *RankingService {
	return &RankingService{pool: pool}
}

// GetRankings retrieves a paginated list of rankings with optional filters.
func (s *RankingService) GetRankings(ctx context.Context, period, category, sortBy string, page, pageSize int) ([]model.Ranking, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if period != "" {
		whereClause += fmt.Sprintf(" AND period = $%d", argIdx)
		args = append(args, period)
		argIdx++
	}

	if category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	// Count.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rankings %s", whereClause)
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count rankings: %w", err)
	}

	// Sort.
	orderClause := "ORDER BY rank_position ASC"
	switch sortBy {
	case "total_assets":
		orderClause = "ORDER BY total_assets DESC"
	case "total_profit":
		orderClause = "ORDER BY total_profit DESC"
	case "win_rate":
		orderClause = "ORDER BY win_rate DESC"
	case "roi":
		orderClause = "ORDER BY roi DESC"
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(
		`SELECT id, user_id, period, category, total_assets, total_profit,
		        win_count, loss_count, win_rate, roi, consecutive_wins,
		        rank_position, calculated_at
		 FROM rankings %s %s LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := s.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list rankings: %w", err)
	}
	defer rows.Close()

	var rankings []model.Ranking
	for rows.Next() {
		var r model.Ranking
		err := rows.Scan(
			&r.ID, &r.UserID, &r.Period, &r.Category, &r.TotalAssets,
			&r.TotalProfit, &r.WinCount, &r.LossCount, &r.WinRate, &r.ROI,
			&r.ConsecutiveWins, &r.RankPosition, &r.CalculatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan ranking: %w", err)
		}
		rankings = append(rankings, r)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rankings: %w", err)
	}

	return rankings, total, nil
}
