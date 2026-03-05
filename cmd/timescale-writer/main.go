package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
	events "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/application/events"
	consumer "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/consumer"
	timescale "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/infrastructure/timescale"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup("timeline-group"),
		kgo.ConsumeTopics("risk-decisions", "ai-insights"),
	)
	if err != nil {
		log.Fatal(err)
	}

	repo := timescale.NewEventRepository(db)
	service := events.New(repo)

	ctx := context.Background()
	consumer.NewDecisionConsumer(client, service).Start(ctx)
}
