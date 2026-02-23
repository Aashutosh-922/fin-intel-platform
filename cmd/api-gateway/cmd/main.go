// package main

// import (
// 	"log"
// 	"net/http"
// )

// import (
// 	"log"
// 	"net/http"
//     "database/sql"
//     _ "github.com/lib/pq"

//     timelineQuery "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/query"
// )

// func main() {
// 	router := buildRouter()

// 	log.Println("api-gateway listening on :8080")
// 	if err := http.ListenAndServe(":8080", router); err != nil {
// 		log.Fatal(err)
// 	}

// 	tsdb, err := sql.Open("postgres",
// 	"postgres://timescale:timescale@timescaledb:5432/events?sslmode=disable")
// 	if err != nil { log.Fatal(err) }

// 	timelineRepo := timelineQuery.NewRepository(tsdb)
//     timelineSvc := timelineQuery.NewService(timelineRepo)
//     timelineHandler := timelineQuery.NewHandler(timelineSvc)
	
	
// }

package main

import (
	"log"
	"net/http"
)

func main() {
	router := buildRouter()

	log.Println("api-gateway listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}