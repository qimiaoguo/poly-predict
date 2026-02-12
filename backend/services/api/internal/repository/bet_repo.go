package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// BetRepository provides database access for bets.
type BetRepository struct {
	pool *pgxpool.Pool
}

// NewBetRepository creates a new BetRepository.
func NewBetRepository(pool *pgxpool.Pool) *BetRepository {
	return &BetRepository{pool: pool}
}

// Create inserts a new bet into the database.
func (r *BetRepository) Create(ctx context.Context, bet *model.Bet) error {
	query := `INSERT INTO bets (id, user_id, event_id, outcome, amount, locked_odds, potential_payout, status, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		bet.ID, bet.UserID, bet.EventID, bet.Outcome, bet.Amount,
		bet.LockedOdds, bet.PotentialPayout, bet.Status, bet.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create bet: %w", err)
	}

	return nil
}

// ListByUser retrieves a paginated list of bets for a given user, optionally filtered by status.
func (r *BetRepository) ListByUser(ctx context.Context, userID string, status string, page, pageSize int) ([]model.Bet, int64, error) {
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIdx := 2

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	// Count.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bets %s", whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count bets: %w", err)
	}

	// Data.
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(
		`SELECT id, user_id, event_id, outcome, amount, locked_odds, potential_payout,
		        status, payout, settled_at, created_at
		 FROM bets %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list bets: %w", err)
	}
	defer rows.Close()

	var bets []model.Bet
	for rows.Next() {
		var b model.Bet
		err := rows.Scan(
			&b.ID, &b.UserID, &b.EventID, &b.Outcome, &b.Amount,
			&b.LockedOdds, &b.PotentialPayout, &b.Status, &b.Payout,
			&b.SettledAt, &b.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan bet: %w", err)
		}
		bets = append(bets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate bets: %w", err)
	}

	return bets, total, nil
}

// GetByID retrieves a single bet by its ID.
func (r *BetRepository) GetByID(ctx context.Context, id string) (*model.Bet, error) {
	query := `SELECT id, user_id, event_id, outcome, amount, locked_odds, potential_payout,
	                 status, payout, settled_at, created_at
	          FROM bets WHERE id = $1`

	var b model.Bet
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.UserID, &b.EventID, &b.Outcome, &b.Amount,
		&b.LockedOdds, &b.PotentialPayout, &b.Status, &b.Payout,
		&b.SettledAt, &b.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get bet by id: %w", err)
	}

	return &b, nil
}
