// package timescale

// import (
// 	"context"
// 	"database/sql"

// 	"github.com/Aashutosh-922/fin-intel-platform/internal/timeline/application/events"
// )

// type EventRepository struct {
// 	db *sql.DB
// }

// func NewEventRepository(db *sql.DB) *EventRepository {
// 	return &EventRepository{db: db}
// }

// func (r *EventRepository) Insert(ctx context.Context, e events.Event) error {
// 	_, err := r.db.ExecContext(ctx, `
// 	INSERT INTO transaction_events (transaction_id, event_type, payload, created_at)
// 	VALUES ($1,$2,$3,now())
// 	`,
// 		e.TransactionID,
// 		e.Type,
// 		e.Payload,
// 	)
// 	return err
// }

package timescale

import (
	"context"
	"database/sql"

	events "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/application/events"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// func (r *EventRepository) Insert(ctx context.Context, e events.Event) error {
// 	_, err := r.db.ExecContext(ctx, `
// 	INSERT INTO transaction_events
// 	(event_id, transaction_id, event_type, payload, created_at)
// 	VALUES ($1,$2,$3,$4,$5)
// 	`,
// 		e.ID,
// 		e.TransactionID,
// 		e.Type,
// 		e.Payload,
// 		e.CreatedAt,
// 	)
// 	return err
// }

func (r *EventRepository) Insert(ctx context.Context, e events.Event) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO transaction_events
		(event_id, transaction_id, event_type, payload, created_at)
		VALUES ($1,$2,$3,$4,$5)
	`,
		e.ID,
		e.TransactionID,
		e.Type,
		e.Payload,
		e.CreatedAt,
	)

	return err
}