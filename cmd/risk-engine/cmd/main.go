package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"net"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/consumer"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/producer"
	"github.com/Aashutosh-922/fin-intel-platform/internal/observability"
)


func main() {
	logger := observability.NewLogger("risk-engine")

	brokerEnv := os.Getenv("KAFKA_BROKERS")
if brokerEnv == "" {
	log.Fatal("KAFKA_BROKERS not set")
}

brokers := strings.Split(brokerEnv, ",")


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("waiting for kafka to be ready...")

	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", brokers[0], 2*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}	


	// Kafka client
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	
		// VERY IMPORTANT ↓↓↓
		kgo.ConsumerGroup("risk-engine-group"),
		kgo.ConsumeTopics("transactions", "transactions-retry"),
			
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		panic(err)
	}
	
	defer client.Close()

	// Producer for risk events
	eventProducer := producer.New(client)

	// Consumer (core engine)
	engine := consumer.New(
		client,
		eventProducer,
		logger,
	)

	// Start consuming
	go engine.Start(ctx)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	logger.Info("shutting down risk-engine")
	cancel()
}