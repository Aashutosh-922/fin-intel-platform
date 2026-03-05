-- Enable extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- TIMESCALE EVENTS TABLE
CREATE TABLE IF NOT EXISTS transaction_events (
    event_id TEXT NOT NULL,
    transaction_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

SELECT create_hypertable('transaction_events','created_at', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_txn_time 
ON transaction_events (transaction_id, created_at DESC);


-- PORTFOLIO TABLE
CREATE TABLE IF NOT EXISTS positions (
    user_id TEXT,
    symbol TEXT,
    quantity INT,
    avg_price DOUBLE PRECISION,
    realized_pnl DOUBLE PRECISION,
    PRIMARY KEY (user_id, symbol)
);

-- MARKET PRICES TABLE
CREATE TABLE IF NOT EXISTS market_prices (
    symbol TEXT PRIMARY KEY,
    price DOUBLE PRECISION,
    updated_at BIGINT
);