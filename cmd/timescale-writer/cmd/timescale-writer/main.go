package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"

	"github.com/Aashutosh-922/fin-intel-platform/internal/application/events"
	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
	"github.com/Aashutosh-922/fin-intel-platform/internal/infrastructure/timescale"
)

type KafkaEvent struct {
	EventVersion  int                    `json:"event_version"`
	TransactionID string                 `json:"transaction_id"`
	EventType     string                 `json:"event_type"`
	EventTime     string                 `json:"event_time"`
	Metadata      map[string]interface{} `json:"metadata"`
}

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.PostgresDSN) // Timescale DSN
	if err != nil {
		log.Fatal(err)
	}

	repo := timescale.NewEventRepository(db)
	service := events.New(repo)

	// Kafka consumer loop (pseudo)
	for {
		msg := consumeKafkaMessage() // replace with real client

		var evt KafkaEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			continue
		}

		eventTime, _ := time.Parse(time.RFC3339, evt.EventTime)

		service.Record(context.Background(), events.Event{
			TransactionID: evt.TransactionID,
			Type:          evt.EventType,
			Time:          eventTime,
			Metadata:      evt.Metadata,
		})
	}
}
