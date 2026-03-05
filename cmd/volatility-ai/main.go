package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/volatility-ai/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/volatility-ai/internal/infrastructure/kafka"
)

func main() {

	ctx := context.Background()

	brokerEnv := os.Getenv("KAFKA_BROKERS")
	if brokerEnv == "" {
		brokerEnv = "localhost:9092"
	}
	brokers := strings.Split(brokerEnv, ",")

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8071"
	}

	detector := application.NewDetector(30)
	service := application.NewService(detector)

	producer := kafka.NewProducer(brokers, "market-alerts")

	consumer := kafka.NewConsumer(
		brokers,
		"volatility-ai-group",
		"trade-executed",
		service,
		producer,
	)

	log.Println("Volatility AI service started...")
	go consumer.Start(ctx)

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	http.HandleFunc("/alerts", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(service.GetAlerts())
	})

	log.Printf("Volatility AI HTTP listening on :%s\n", httpPort)
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
		log.Fatal(err)
	}
}
