CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY,
    display_name    VARCHAR(100) NOT NULL,
    avatar_url      TEXT,
    balance         BIGINT NOT NULL DEFAULT 10000,
    frozen_balance  BIGINT NOT NULL DEFAULT 0,
    level           INTEGER NOT NULL DEFAULT 1,
    xp              INTEGER NOT NULL DEFAULT 0,
    current_streak  INTEGER NOT NULL DEFAULT 0,
    max_streak      INTEGER NOT NULL DEFAULT 0,
    total_bets      INTEGER NOT NULL DEFAULT 0,
    total_wins      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT balance_non_negative CHECK (balance >= 0),
    CONSTRAINT frozen_balance_non_negative CHECK (frozen_balance >= 0)
);

CREATE INDEX IF NOT EXISTS idx_users_balance ON users(balance DESC);
