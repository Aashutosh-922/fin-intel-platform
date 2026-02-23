// package query

// import (
// 	"context"
// 	"database/sql"
// )

// type Event struct {
// 	Type string `json:"type"`
// 	Time string `json:"time"`
// }

// type Repository struct {
// 	db *sql.DB
// }

// func NewRepository(db *sql.DB) *Repository {
// 	return &Repository{db: db}
// }

// func (r *Repository) GetTimeline(ctx context.Context, txnID string) ([]Event, error) {
// 	rows, err := r.db.QueryContext(ctx, `
// 		SELECT event_type, created_at
// 		FROM transaction_events
// 		WHERE transaction_id=$1
// 		ORDER BY created_at ASC
// 	`, txnID)

// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var events []Event

// 	for rows.Next() {
// 		var e Event
// 		if err := rows.Scan(&e.Type, &e.Time); err != nil {
// 			return nil, err
// 		}
// 		events = append(events, e)
// 	}

// 	return events, nil
// }

package query

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTimeline(ctx context.Context, txnID string) ([]TimelineEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT event_type, created_at
		FROM transaction_events
		WHERE transaction_id = $1
		ORDER BY created_at ASC
	`, txnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []TimelineEvent

	for rows.Next() {
		var e TimelineEvent
		if err := rows.Scan(&e.EventType, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}