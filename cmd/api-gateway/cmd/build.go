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
	"net/http"
	"log"
    "database/sql"
    _ "github.com/lib/pq"


	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/clients"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/repository"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/server"
	timelineQuery "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/query"

)

func buildRouter() http.Handler {
	ingestion := clients.NewIngestionClient("http://ingestion:8080")
	repo := repository.NewReadOnlyRepo()
	ai := clients.NewAIClient()

	    // ---- Timescale DB ----
		tsdb, err := sql.Open("postgres",
        "postgres://timescale:timescale@timescaledb:5432/events?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }

    timelineRepo := timelineQuery.NewRepository(tsdb)
    timelineSvc := timelineQuery.NewService(timelineRepo)

	h := handlers.NewHandler(ingestion, repo, ai, timelineSvc)

	return server.NewRouter(h)
}
