package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"database/sql"
	_ "github.com/lib/pq"

	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
	"github.com/Aashutosh-922/fin-intel-platform/internal/observability"

	httptransport "github.com/Aashutosh-922/fin-intel-platform/cmd/ingestion-service/internal/transport/http"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ingestion-service/internal/kafka"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ingestion-service/internal/storage"

)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.ServiceName)

	// brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	// producer := kafka.NewProducer(brokers, "transactions")
	brokersEnv := os.Getenv("KAFKA_BROKERS")
    logger.Info("kafka brokers env", "value", brokersEnv)

    brokers := strings.Split(brokersEnv, ",")
	producer := kafka.NewProducer(brokers, "transactions", "transactions-dlq")


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

// 	db, err := sql.Open("postgres", os.Getenv("POSTGRES_DSN"))
// if err != nil {
// 	logger.Error("db connection failed", "err", err)
// 	os.Exit(1)
// }
db, err := sql.Open("postgres", cfg.PostgresDSN)
if err != nil {
	logger.Error("db connection failed", "err", err)
	panic(err)
}

if err := storage.Migrate(db); err != nil {
	logger.Error("migration failed", "err", err)
	panic(err)
}

logger.Info("database migrated")



server := httptransport.NewServer(cfg, logger, producer, db)

	go func() {
		logger.Info("starting ingestion service", "port", cfg.HTTPPort)
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server error", "err", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	server.Shutdown(shutdownCtx)
}
