package postgres

import (
	"context"
	"database/sql"
)

type IdempotencyRepo struct {
	db *sql.DB
}

func NewIdempotencyRepo(db *sql.DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

// Returns existing transaction_id if key exists
func (r *IdempotencyRepo) Get(
	ctx context.Context,
	key string,
) (string, bool, error) {
	var txID string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT transaction_id FROM idempotency_keys WHERE key = $1`,
		key,
	).Scan(&txID)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return txID, true, nil
}

// Stores key → transaction mapping
func (r *IdempotencyRepo) Save(
	ctx context.Context,
	key string,
	transactionID string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO idempotency_keys (key, transaction_id)
		 VALUES ($1, $2)`,
		key,
		transactionID,
	)
	return err
}
