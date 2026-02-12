package model

import "time"

// User represents a user account in the system.
type User struct {
	ID             string    `json:"id" db:"id"`
	DisplayName    string    `json:"display_name" db:"display_name"`
	AvatarURL      *string   `json:"avatar_url" db:"avatar_url"`
	Balance        int64     `json:"balance" db:"balance"`
	FrozenBalance  int64     `json:"frozen_balance" db:"frozen_balance"`
	Level          int       `json:"level" db:"level"`
	XP             int       `json:"xp" db:"xp"`
	CurrentStreak  int       `json:"current_streak" db:"current_streak"`
	MaxStreak      int       `json:"max_streak" db:"max_streak"`
	TotalBets      int       `json:"total_bets" db:"total_bets"`
	TotalWins      int       `json:"total_wins" db:"total_wins"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
