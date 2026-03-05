package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/infrasturcture/kafka"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/infrasturcture/repository"
	httptransport "github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/transport/http"
)

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	producer := kafka.NewProducer(brokers, "orders")
	defer producer.Close()
	idRepo := repository.NewMemoryRepo()
	service := application.NewService(producer, idRepo)
	handler := httptransport.NewHandler(service, producer)
	router := httptransport.NewRouter(handler)

	log.Println("order-service listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
