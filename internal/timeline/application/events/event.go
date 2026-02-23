// package events

// import "time"

// type Event struct {
// 	EventID       string
// 	TransactionID string
// 	Type          string
// 	Payload       string
// 	CreatedAt     time.Time
// }

// package events

// import "time"

// type Event struct {
// 	ID            string
// 	TransactionID string
// 	Type          string
// 	Payload       string
// 	CreatedAt     time.Time
// }

package events

import "time"

type Event struct {
	ID            string    `db:"event_id"`       // maps to DB column
	TransactionID string    `db:"transaction_id"`
	Type          string    `db:"event_type"`
	Payload       string    `db:"payload"`
	CreatedAt     time.Time `db:"created_at"`
}