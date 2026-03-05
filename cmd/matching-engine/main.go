package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/infrastructure/kafka"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/infrastructure/state"
	httpTransport "github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/transport/http"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/transport/ws"
)

func main() {

	ctx := context.Background()
	brokerEnv := os.Getenv("KAFKA_BROKERS")
	if brokerEnv == "" {
		brokerEnv = "localhost:9092"
	}
	brokers := strings.Split(brokerEnv, ",")

	// 🔥 Create WebSocket Hub
	hub := ws.NewHub()

	// 🔥 Create Matcher with Hub injected
	matcher := application.NewMatcher(hub)
	stateFile := os.Getenv("ORDERBOOK_STATE_FILE")
	if stateFile == "" {
		stateFile = "/tmp/orderbook-state.json"
	}
	store := state.NewFileStore(stateFile)
	if snapshots, err := store.Load(); err == nil {
		matcher.RestoreFromSnapshots(snapshots)
		log.Printf("restored %d symbols from orderbook state\n", len(snapshots))
	} else {
		log.Printf("orderbook restore skipped: %v\n", err)
	}

	// Kafka producer for trades
	tradeProducer := kafka.NewProducer(brokers, "trade-executed")
	snapshotProducer := kafka.NewProducer(brokers, "orderbook-snapshots")
	deltaProducer := kafka.NewProducer(brokers, "orderbook-deltas")
	defer tradeProducer.Close()
	defer snapshotProducer.Close()
	defer deltaProducer.Close()

	// Kafka consumer for orders
	consumer := kafka.NewConsumer(
		brokers,
		"matching-engine-group",
		"orders",
		matcher,
		tradeProducer,
		snapshotProducer,
		deltaProducer,
		store,
	)

	// 🔥 Start Kafka consumer in goroutine
	go consumer.Start(ctx)
	go kafka.StartCancelConsumer(ctx, brokers, matcher)

	// HTTP + WebSocket Handlers
	httpHandler := httpTransport.NewHandler(matcher)
	wsHandler := ws.NewHandler(matcher, hub)

	r := chi.NewRouter()

	// REST snapshot endpoint
	r.Get("/orderbook/{symbol}", httpHandler.GetOrderBook)

	// WebSocket endpoint
	r.Get("/ws/orderbook/{symbol}", wsHandler.OrderBookStream)

	r.Get("/ws/trades/{symbol}", wsHandler.TradeStream)

	log.Println("Matching engine running on :8090")

	// 🔥 Start HTTP server (blocking)
	if err := http.ListenAndServe(":8090", r); err != nil {
		log.Fatal(err)
	}
}
