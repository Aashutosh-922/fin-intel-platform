-- App database schema (postgres://fintech:fintech@postgres:5432/fintech)

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'RECEIVED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS positions (
    user_id TEXT,
    symbol TEXT,
    quantity INT,
    avg_price DOUBLE PRECISION,
    realized_pnl DOUBLE PRECISION,
    unrealized_pnl DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, symbol)
);

ALTER TABLE positions
    ADD COLUMN IF NOT EXISTS unrealized_pnl DOUBLE PRECISION NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS market_prices (
    symbol TEXT PRIMARY KEY,
    price DOUBLE PRECISION,
    updated_at BIGINT
);
