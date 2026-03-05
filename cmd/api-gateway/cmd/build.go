// package main

// import (
// 	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
// 	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/server"
// )

// func Build(cfg config.Config) http.Handler {
// 	ingestion := clients.NewIngestionClient(cfg)
// 	repo := repository.NewReadOnlyRepo(cfg)
// 	ai := clients.NewAIClient(cfg)

// 	h := handlers.NewHandler(ingestion, repo, ai)
// 	return server.NewRouter(h)
// }

// func buildRouter() http.Handler {
// 	h := &handlers.Handler{} // dependencies can be injected later
// 	return server.NewRouter(h)
// }

// package main

// import (
// 	"net/http"

// 	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/clients"
// 	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
// 	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/repository"
// 	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/server"
// )

// func buildRouter() http.Handler {
// 	ingestionClient := clients.NewIngestionClient()
// 	repo := &repository.ReadOnlyRepo{}

// 	h := &handlers.Handler{
// 		IngestionClient: ingestionClient, // ✅ THIS WAS NIL BEFORE
// 		Repo:            repo,
// 	}

// 	return server.NewRouter(h)
// }

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/clients"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/repository"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/server"
	timelineQuery "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/query"
)

func buildRouter() http.Handler {
	ingestion := clients.NewIngestionClient("http://ingestion:8080")

	appDSN := os.Getenv("APP_POSTGRES_DSN")
	if appDSN == "" {
		appDSN = "postgres://fintech:fintech@postgres:5432/fintech?sslmode=disable"
	}
	eventDSN := os.Getenv("EVENTS_POSTGRES_DSN")
	if eventDSN == "" {
		eventDSN = "postgres://timescale:timescale@timescaledb:5432/events?sslmode=disable"
	}

	appDB, err := sql.Open("postgres", appDSN)
	if err != nil {
		log.Fatal(err)
	}
	eventDB, err := sql.Open("postgres", eventDSN)
	if err != nil {
		log.Fatal(err)
	}

	ai := clients.NewAIClient(eventDB)
	repo := repository.NewReadOnlyRepo(appDB, eventDB)
	timelineRepo := timelineQuery.NewRepository(eventDB)
	timelineSvc := timelineQuery.NewService(timelineRepo)

	h := handlers.NewHandler(ingestion, repo, ai, timelineSvc)

	return server.NewRouter(h)
}
