CREATE TYPE bet_status AS ENUM ('pending', 'won', 'lost', 'cancelled');

CREATE TABLE bets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    event_id        VARCHAR(100) NOT NULL REFERENCES events(id),
    outcome         VARCHAR(10) NOT NULL,
    amount          BIGINT NOT NULL,
    locked_odds     NUMERIC(10,6) NOT NULL,
    potential_payout BIGINT NOT NULL,
    status          bet_status NOT NULL DEFAULT 'pending',
    payout          BIGINT DEFAULT 0,
    settled_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT amount_positive CHECK (amount > 0),
    CONSTRAINT odds_valid CHECK (locked_odds > 0 AND locked_odds <= 1)
);

CREATE INDEX idx_bets_user ON bets(user_id, created_at DESC);
CREATE INDEX idx_bets_event ON bets(event_id);
CREATE INDEX idx_bets_pending ON bets(event_id, status) WHERE status = 'pending';

CREATE TABLE settlements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        VARCHAR(100) NOT NULL UNIQUE REFERENCES events(id),
    resolved_outcome VARCHAR(100) NOT NULL,
    total_bets      INTEGER NOT NULL DEFAULT 0,
    total_payouts   BIGINT NOT NULL DEFAULT 0,
    settled_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE credit_transactions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    type            VARCHAR(50) NOT NULL,
    amount          BIGINT NOT NULL,
    balance_after   BIGINT NOT NULL,
    reference_id    UUID,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_tx_user ON credit_transactions(user_id, created_at DESC);

CREATE TABLE rankings (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    period          VARCHAR(20) NOT NULL DEFAULT 'all_time',
    category        VARCHAR(100),
    total_assets    BIGINT NOT NULL DEFAULT 0,
    total_profit    BIGINT NOT NULL DEFAULT 0,
    win_count       INTEGER NOT NULL DEFAULT 0,
    loss_count      INTEGER NOT NULL DEFAULT 0,
    win_rate        NUMERIC(5,4) NOT NULL DEFAULT 0,
    roi             NUMERIC(8,4) NOT NULL DEFAULT 0,
    consecutive_wins INTEGER NOT NULL DEFAULT 0,
    rank_position   INTEGER,
    calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT rankings_unique UNIQUE(user_id, period, category)
);

CREATE INDEX idx_rankings_leaderboard ON rankings(period, category, rank_position ASC);
