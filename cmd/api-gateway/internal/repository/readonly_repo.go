package repository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
)

type ReadOnlyRepo struct {
	appDB   *sql.DB
	eventDB *sql.DB
}

func NewReadOnlyRepo(appDB, eventDB *sql.DB) *ReadOnlyRepo {
	return &ReadOnlyRepo{
		appDB:   appDB,
		eventDB: eventDB,
	}
}

func (r *ReadOnlyRepo) GetTransaction(id string) (handlers.Transaction, error) {
	var tx handlers.Transaction
	row := r.appDB.QueryRow(
		`SELECT id, amount, status
		 FROM transactions
		 WHERE id = $1`,
		id,
	)
	if err := row.Scan(&tx.ID, &tx.Amount, &tx.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.Transaction{}, sql.ErrNoRows
		}
		return handlers.Transaction{}, err
	}
	return tx, nil
}

func (r *ReadOnlyRepo) GetRisk(id string) (handlers.Risk, error) {
	type riskPayload struct {
		RiskScore float64 `json:"risk_score"`
	}
	var raw string
	row := r.eventDB.QueryRow(
		`SELECT payload
		 FROM transaction_events
		 WHERE transaction_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		id,
	)
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.Risk{}, sql.ErrNoRows
		}
		return handlers.Risk{}, err
	}

	var payload riskPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return handlers.Risk{}, err
	}

	return handlers.Risk{Score: payload.RiskScore}, nil
}

func (r *ReadOnlyRepo) GetExplanation(id string) (handlers.AIExplanation, error) {
	var eventType string
	err := r.eventDB.QueryRow(
		`SELECT event_type
		 FROM transaction_events
		 WHERE transaction_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		id,
	).Scan(&eventType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.AIExplanation{}, sql.ErrNoRows
		}
		return handlers.AIExplanation{}, err
	}

	return handlers.AIExplanation{
		Text: "Decision from risk-engine: " + eventType,
	}, nil
}

func (r *ReadOnlyRepo) GetEvents(id string) ([]handlers.Event, error) {
	rows, err := r.eventDB.Query(
		`SELECT event_type, created_at::text
		 FROM transaction_events
		 WHERE transaction_id = $1
		 ORDER BY created_at ASC`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]handlers.Event, 0, 8)
	for rows.Next() {
		var e handlers.Event
		if err := rows.Scan(&e.Type, &e.Data); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
