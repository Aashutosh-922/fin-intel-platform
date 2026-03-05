package repository

import (
	"context"
	"database/sql"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) GetPosition(ctx context.Context, userID, symbol string) (*domain.Position, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT quantity, avg_price, realized_pnl, unrealized_pnl
		 FROM positions 
		 WHERE user_id=$1 AND symbol=$2`,
		userID, symbol,
	)

	var pos domain.Position
	pos.UserID = userID
	pos.Symbol = symbol

	err := row.Scan(&pos.Quantity, &pos.AvgPrice, &pos.RealizedPnL, &pos.UnrealizedPnL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pos, err
}

func (r *PostgresRepo) SavePosition(ctx context.Context, pos *domain.Position) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO positions (user_id, symbol, quantity, avg_price, realized_pnl, unrealized_pnl)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (user_id, symbol)
		 DO UPDATE SET
		   quantity=EXCLUDED.quantity,
		   avg_price=EXCLUDED.avg_price,
		   realized_pnl=EXCLUDED.realized_pnl,
		   unrealized_pnl=EXCLUDED.unrealized_pnl`,
		pos.UserID, pos.Symbol, pos.Quantity,
		pos.AvgPrice, pos.RealizedPnL, pos.UnrealizedPnL,
	)
	return err
}

func (r *PostgresRepo) GetAllBySymbol(ctx context.Context, symbol string) ([]domain.Position, error) {

	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, quantity, avg_price, realized_pnl, unrealized_pnl
		 FROM positions
		 WHERE symbol=$1`,
		symbol,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []domain.Position

	for rows.Next() {
		var p domain.Position
		p.Symbol = symbol

		if err := rows.Scan(
			&p.UserID,
			&p.Quantity,
			&p.AvgPrice,
			&p.RealizedPnL,
			&p.UnrealizedPnL,
		); err != nil {
			return nil, err
		}

		positions = append(positions, p)
	}

	return positions, nil
}
