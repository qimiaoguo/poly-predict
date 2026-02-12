CREATE TYPE event_status AS ENUM ('open', 'closed', 'resolved');

CREATE TABLE events (
    id                  VARCHAR(100) PRIMARY KEY,
    polymarket_event_id VARCHAR(100),
    slug                VARCHAR(500) NOT NULL,
    question            TEXT NOT NULL,
    description         TEXT,
    category            VARCHAR(100),
    image_url           TEXT,

    outcomes            JSONB NOT NULL DEFAULT '["Yes", "No"]',
    outcome_prices      JSONB NOT NULL DEFAULT '["0.5", "0.5"]',
    clob_token_ids      JSONB NOT NULL DEFAULT '[]',

    status              event_status NOT NULL DEFAULT 'open',
    resolved_outcome    VARCHAR(100),
    resolved_at         TIMESTAMPTZ,

    volume              NUMERIC(20,2) DEFAULT 0,
    volume_24h          NUMERIC(20,2) DEFAULT 0,
    liquidity           NUMERIC(20,2) DEFAULT 0,
    end_date            TIMESTAMPTZ,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    synced_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_category ON events(category);
CREATE INDEX idx_events_slug ON events(slug);
CREATE INDEX idx_events_volume_24h ON events(volume_24h DESC);

CREATE TABLE price_history (
    id              BIGSERIAL PRIMARY KEY,
    event_id        VARCHAR(100) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    outcome_label   VARCHAR(100) NOT NULL,
    price           NUMERIC(10,6) NOT NULL,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_price_history_event_time ON price_history(event_id, recorded_at DESC);
