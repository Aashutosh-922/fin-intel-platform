package storage

import "database/sql"

func Migrate(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'RECEIVED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
	_, err := db.Exec(query)
	return err
}