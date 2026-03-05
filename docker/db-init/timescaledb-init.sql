-- Event timeline schema (postgres://timescale:timescale@timescaledb:5432/events)

CREATE EXTENSION IF NOT EXISTS timescaledb;

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
