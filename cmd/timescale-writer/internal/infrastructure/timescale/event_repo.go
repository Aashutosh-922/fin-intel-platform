package timescale

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Event struct {
	TransactionID string
	EventType     string
	EventTime     time.Time
	Metadata      map[string]interface{}
}

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Insert(
	ctx context.Context,
	e Event,
) error {
	meta, err := json.Marshal(e.Metadata)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO transaction_events
		(transaction_id, event_type, event_time, metadata)
		VALUES ($1, $2, $3, $4)
	`,
		e.TransactionID,
		e.EventType,
		e.EventTime,
		meta,
	)

	return err
}
