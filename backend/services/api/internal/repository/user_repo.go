package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// UserRepository provides database access for users.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// GetByID retrieves a user by their ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, display_name, avatar_url, balance, frozen_balance, level, xp,
	                 current_streak, max_streak, total_bets, total_wins, created_at, updated_at
	          FROM users WHERE id = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.DisplayName, &u.AvatarURL, &u.Balance, &u.FrozenBalance,
		&u.Level, &u.XP, &u.CurrentStreak, &u.MaxStreak, &u.TotalBets,
		&u.TotalWins, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, display_name, balance, created_at, updated_at)
	          VALUES ($1, $2, $3, NOW(), NOW())`

	_, err := r.pool.Exec(ctx, query, user.ID, user.DisplayName, user.Balance)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// UpdateDisplayName updates the user's display name and returns the updated user.
func (r *UserRepository) UpdateDisplayName(ctx context.Context, id, name string) (*model.User, error) {
	query := `UPDATE users SET display_name = $2, updated_at = NOW() WHERE id = $1
	          RETURNING id, display_name, avatar_url, balance, frozen_balance, level, xp,
	                    current_streak, max_streak, total_bets, total_wins, created_at, updated_at`

	var u model.User
	err := r.pool.QueryRow(ctx, query, id, name).Scan(
		&u.ID, &u.DisplayName, &u.AvatarURL, &u.Balance, &u.FrozenBalance,
		&u.Level, &u.XP, &u.CurrentStreak, &u.MaxStreak, &u.TotalBets,
		&u.TotalWins, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update display name: %w", err)
	}

	return &u, nil
}

// GetOrCreate retrieves a user by ID, or creates one with default values if not found.
// Uses INSERT ... ON CONFLICT DO NOTHING followed by a SELECT to handle the upsert.
func (r *UserRepository) GetOrCreate(ctx context.Context, id, displayName string) (*model.User, error) {
	const defaultBalance int64 = 10000 // 10,000 credits

	upsertQuery := `INSERT INTO users (id, display_name, balance, created_at, updated_at)
	                VALUES ($1, $2, $3, NOW(), NOW())
	                ON CONFLICT (id) DO NOTHING`

	_, err := r.pool.Exec(ctx, upsertQuery, id, displayName, defaultBalance)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return r.GetByID(ctx, id)
}

// GetTransactions retrieves a paginated list of credit transactions for a user.
func (r *UserRepository) GetTransactions(ctx context.Context, userID string, page, pageSize int) ([]model.CreditTransaction, int64, error) {
	// Count.
	var total int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM credit_transactions WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	// Data.
	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, type, amount, balance_after, reference_id, description, created_at
	          FROM credit_transactions
	          WHERE user_id = $1
	          ORDER BY created_at DESC
	          LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []model.CreditTransaction
	for rows.Next() {
		var t model.CreditTransaction
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Type, &t.Amount, &t.BalanceAfter,
			&t.ReferenceID, &t.Description, &t.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate transactions: %w", err)
	}

	return transactions, total, nil
}
