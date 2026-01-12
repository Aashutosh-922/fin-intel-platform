package main

import (
	"log"

	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
	"github.com/Aashutosh-922/fin-intel-platform/internal/observability"
	httptransport "github.com/Aashutosh-922/fin-intel-platform/internal/transport/http"
)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.ServiceName)

	logger.Info("starting service")

	server := httptransport.NewServer(cfg, logger)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
