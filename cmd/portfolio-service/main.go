package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/infrastructure/kafka"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/infrastructure/repository"
)

func main() {

	ctx := context.Background()
	brokerEnv := os.Getenv("KAFKA_BROKERS")
	if brokerEnv == "" {
		brokerEnv = "localhost:9092"
	}
	brokers := strings.Split(brokerEnv, ",")

	repo := repository.NewPostgresRepo(initDB())
	service := application.NewService(repo)

	go kafka.StartTradeConsumer(ctx, brokers, service)
	go kafka.StartMarketConsumer(ctx, brokers, service)

	select {}
}

func initDB() *sql.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://fintech:fintech@localhost:5432/fintech?sslmode=disable"
	}

	for i := 0; i < 30; i++ {
		db, err := sql.Open("postgres", dsn)
		if err == nil {
			if err := db.Ping(); err == nil {
				return db
			}
			_ = db.Close()
		}
		log.Println("waiting for postgres...")
		time.Sleep(2 * time.Second)
	}

	log.Fatal("failed to connect to postgres for portfolio-service")
	return nil
}
