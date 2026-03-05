package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/market-data-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/market-data-service/internal/infrastucture/kafka"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/market-data-service/internal/transport/ws"
	"github.com/Aashutosh-922/fin-intel-platform/internal/contracts"
	kafkaGo "github.com/segmentio/kafka-go"
)

func main() {

	brokerEnv := os.Getenv("KAFKA_BROKERS")
	if brokerEnv == "" {
		brokerEnv = "localhost:9092"
	}
	brokers := strings.Split(brokerEnv, ",")
	ctx := context.Background()
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8070"
	}

	symbols := []string{
		"RELIANCE",
		"TCS",
		"INFY",
		"HDFCBANK",
	}

	generator := application.NewGenerator(symbols)
	producer := kafka.NewProducer(brokers, "market-ticks")
	defer producer.Close()

	hub := ws.NewHub()
	go ws.StartServer(httpPort, hub)

	tradeConsumer := kafka.NewTopicConsumer(
		brokers,
		"market-data-broadcast-trades",
		"trade-executed",
		func(msg kafkaGo.Message) {
			if contracts.ValidateTopicPayload("trade-executed", msg.Value) == nil {
				hub.Broadcast("trade-executed", msg.Value)
			}
		},
	)

	snapshotConsumer := kafka.NewTopicConsumer(
		brokers,
		"market-data-broadcast-orderbook",
		"orderbook-snapshots",
		func(msg kafkaGo.Message) {
			if contracts.ValidateTopicPayload("orderbook-snapshots", msg.Value) == nil {
				hub.Broadcast("orderbook-snapshots", msg.Value)
			}
		},
	)

	deltaConsumer := kafka.NewTopicConsumer(
		brokers,
		"market-data-broadcast-orderbook-deltas",
		"orderbook-deltas",
		func(msg kafkaGo.Message) {
			if contracts.ValidateTopicPayload("orderbook-deltas", msg.Value) == nil {
				hub.Broadcast("orderbook-deltas", msg.Value)
			}
		},
	)

	alertConsumer := kafka.NewTopicConsumer(
		brokers,
		"market-data-broadcast-volatility-alerts",
		"market-alerts",
		func(msg kafkaGo.Message) {
			if contracts.ValidateTopicPayload("market-alerts", msg.Value) == nil {
				hub.Broadcast("market-alerts", msg.Value)
			}
		},
	)

	riskRejectionConsumer := kafka.NewTopicConsumer(
		brokers,
		"market-data-broadcast-risk-decisions",
		"risk-decisions",
		func(msg kafkaGo.Message) {
			if contracts.ValidateTopicPayload("risk-decisions", msg.Value) != nil {
				return
			}
			var evt struct {
				Decision string `json:"decision"`
			}
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				return
			}
			if strings.EqualFold(evt.Decision, "APPROVED") {
				return
			}
			hub.Broadcast("risk-rejections", msg.Value)
		},
	)

	go tradeConsumer.Start(ctx)
	go snapshotConsumer.Start(ctx)
	go deltaConsumer.Start(ctx)
	go alertConsumer.Start(ctx)
	go riskRejectionConsumer.Start(ctx)

	log.Println("Market data service started...")

	for {
		for _, symbol := range symbols {

			tick := generator.NextTick(symbol)

			if err := producer.Publish(ctx, tick); err != nil {
				log.Println("publish error:", err)
			}

			if payload, err := json.Marshal(tick); err == nil {
				hub.Broadcast("market-ticks", payload)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
