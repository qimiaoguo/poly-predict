package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/poly-predict/backend/pkg/model"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// List returns a paginated list of users, optionally filtered by display_name.
func (r *UserRepository) List(ctx context.Context, search, sortBy string, page, pageSize int) ([]model.User, int64, error) {
	offset := (page - 1) * pageSize

	// Validate sort column.
	allowedSorts := map[string]string{
		"created_at":   "created_at DESC",
		"balance":      "balance DESC",
		"display_name": "display_name ASC",
		"total_bets":   "total_bets DESC",
	}
	orderClause := "created_at DESC"
	if s, ok := allowedSorts[sortBy]; ok {
		orderClause = s
	}

	var total int64
	var users []model.User

	if search != "" {
		searchPattern := "%" + search + "%"

		err := r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE display_name ILIKE $1`, searchPattern,
		).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count users: %w", err)
		}

		rows, err := r.pool.Query(ctx,
			fmt.Sprintf(`SELECT id, display_name, avatar_url, balance, frozen_balance,
				level, xp, current_streak, max_streak, total_bets, total_wins,
				created_at, updated_at
			FROM users
			WHERE display_name ILIKE $1
			ORDER BY %s
			LIMIT $2 OFFSET $3`, orderClause),
			searchPattern, pageSize, offset,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query users: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var u model.User
			if err := rows.Scan(
				&u.ID, &u.DisplayName, &u.AvatarURL, &u.Balance, &u.FrozenBalance,
				&u.Level, &u.XP, &u.CurrentStreak, &u.MaxStreak, &u.TotalBets, &u.TotalWins,
				&u.CreatedAt, &u.UpdatedAt,
			); err != nil {
				return nil, 0, fmt.Errorf("failed to scan user: %w", err)
			}
			users = append(users, u)
		}
	} else {
		err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count users: %w", err)
		}

		rows, err := r.pool.Query(ctx,
			fmt.Sprintf(`SELECT id, display_name, avatar_url, balance, frozen_balance,
				level, xp, current_streak, max_streak, total_bets, total_wins,
				created_at, updated_at
			FROM users
			ORDER BY %s
			LIMIT $1 OFFSET $2`, orderClause),
			pageSize, offset,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query users: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var u model.User
			if err := rows.Scan(
				&u.ID, &u.DisplayName, &u.AvatarURL, &u.Balance, &u.FrozenBalance,
				&u.Level, &u.XP, &u.CurrentStreak, &u.MaxStreak, &u.TotalBets, &u.TotalWins,
				&u.CreatedAt, &u.UpdatedAt,
			); err != nil {
				return nil, 0, fmt.Errorf("failed to scan user: %w", err)
			}
			users = append(users, u)
		}
	}

	if users == nil {
		users = []model.User{}
	}

	return users, total, nil
}

// GetByID returns a single user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, display_name, avatar_url, balance, frozen_balance,
			level, xp, current_streak, max_streak, total_bets, total_wins,
			created_at, updated_at
		FROM users
		WHERE id = $1`, id,
	).Scan(
		&u.ID, &u.DisplayName, &u.AvatarURL, &u.Balance, &u.FrozenBalance,
		&u.Level, &u.XP, &u.CurrentStreak, &u.MaxStreak, &u.TotalBets, &u.TotalWins,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

// AdjustBalance atomically adjusts a user's balance and logs a credit_transaction.
func (r *UserRepository) AdjustBalance(ctx context.Context, id string, adjustment int64) (*model.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update balance atomically.
	u := &model.User{}
	err = tx.QueryRow(ctx,
		`UPDATE users
		 SET balance = balance + $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, display_name, avatar_url, balance, frozen_balance,
			level, xp, current_streak, max_streak, total_bets, total_wins,
			created_at, updated_at`,
		adjustment, id,
	).Scan(
		&u.ID, &u.DisplayName, &u.AvatarURL, &u.Balance, &u.FrozenBalance,
		&u.Level, &u.XP, &u.CurrentStreak, &u.MaxStreak, &u.TotalBets, &u.TotalWins,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust balance: %w", err)
	}

	// Log credit transaction.
	txType := "admin_adjustment"
	description := "Admin balance adjustment"
	_, err = tx.Exec(ctx,
		`INSERT INTO credit_transactions (user_id, type, amount, balance_after, description)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, txType, adjustment, u.Balance, description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to log credit transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return u, nil
}
