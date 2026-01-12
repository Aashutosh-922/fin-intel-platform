package postgres

import (
	"context"
	"database/sql"

	"github.com/Aashutosh-922/fin-intel-platform/internal/domain/transaction"
)

type TransactionRepo struct {
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) Save(ctx context.Context, tx transaction.Transaction) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO transactions
		(id, user_id, amount, currency, country, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`,
		tx.ID,
		tx.UserID,
		tx.Amount,
		tx.Currency,
		tx.Country,
		tx.Status,
		tx.CreatedAt,
	)
	return err
}

func (r *TransactionRepo) GetByID(
	ctx context.Context,
	id string,
) (transaction.Transaction, error) {

	var tx transaction.Transaction

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, amount, currency, country, status, created_at
		FROM transactions
		WHERE id = $1
	`, id).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.Amount,
		&tx.Currency,
		&tx.Country,
		&tx.Status,
		&tx.CreatedAt,
	)

	return tx, err
}
