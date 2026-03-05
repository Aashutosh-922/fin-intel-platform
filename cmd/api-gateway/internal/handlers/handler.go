// package handlers

// import (
// 	"context"

// 	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/clients"
// )

// /*
//    ===== Minimal DTOs (Handler-layer only) =====
//    These are NOT domain models.
//    They only shape API responses.
// */

// type Transaction struct {
// 	ID     string
// 	Amount float64
// 	Status string
// }

// type Risk struct {
// 	Score int
// }

// type AIExplanation struct {
// 	Text string
// }

// type Event struct {
// 	Type string
// 	Time string
// }

// type AIResponse struct {
// 	Text string
// }

// /*
//    ===== Interfaces =====
// */

// type IngestionClient interface {
// 	Forward(ctx context.Context, payload []byte) ([]byte, error)
// }

// type ReadOnlyRepository interface {
// 	GetTransaction(id string) (*Transaction, error)
// 	GetRisk(id string) (*Risk, error)
// 	GetExplanation(id string) (*AIExplanation, error)
// 	GetEvents(id string) ([]Event, error)
// }

// type AIClient interface {
// 	Query(ctx context.Context, question string) (*AIResponse, error)
// }

// /*
//    ===== Handler =====
// */

// type Handler struct {
// 	ingestionClient *clients.IngestionClient
// 	Repo            ReadOnlyRepository
// 	AiClient        AIClient
// }

// func NewHandler(
// 	ingestion IngestionClient,
// 	repo ReadOnlyRepository,
// 	ai AIClient,
// ) *Handler {
// 	return &Handler{
// 		IngestionClient: ingestion,
// 		Repo:            repo,
// 		AIClient:        ai,
// 	}
// }

// package handlers

// import "context"

// /*
//    Interfaces keep api-gateway decoupled.
//    These are MOCKABLE and SAFE.
// */

// type IngestionClient interface {
// 	CreateTransaction(ctx context.Context, req CreateTransactionRequest) error
// }

// type ReadOnlyRepository interface {
// 	GetTransaction(id string) (Transaction, error)
// 	GetRisk(id string) (Risk, error)
// 	GetExplanation(id string) (AIExplanation, error)
// 	GetEvents(id string) ([]Event, error)
// }

// type AIClient interface {
// 	Query(ctx context.Context, question string) (AIResponse, error)
// }

// type Handler struct {
// 	IngestionClient IngestionClient
// 	Repo            ReadOnlyRepository
// 	AIClient        AIClient
// }

// package handlers

// import "github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/clients"

// type Handler struct {
// 	IngestionClient *clients.IngestionClient
// 	Repo            *repository.ReadOnlyRepo
// 	AIClient        AIClient
// }

// func NewHandler(
// 	ingestion *clients.IngestionClient,
// 	repo ReadOnlyRepository,
// 	ai AIClient,
// ) *Handler {
// 	return &Handler{
// 		IngestionClient: ingestion,
// 		Repo:            repo,
// 		AIClient:        ai,
// 	}
// }

package handlers

import (
	"context"

	timelineQuery "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/query"
)

// ---------- Interfaces (DECLARE ONCE ONLY) ----------

// IngestionClient sends transactions to ingestion service
type IngestionClient interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) error
}

// ReadOnlyRepository fetches data for reads
type ReadOnlyRepository interface {
	GetTransaction(id string) (Transaction, error)
	GetRisk(id string) (Risk, error)
	GetExplanation(id string) (AIExplanation, error)
	GetEvents(id string) ([]Event, error)
}

// AIClient handles AI queries
type AIClient interface {
	Query(ctx context.Context, q AIQuery) (AIResponse, error)
}

// ---------- Handler ----------

type Handler struct {
	ingestion   IngestionClient
	repo        ReadOnlyRepository
	ai          AIClient
	timelineSvc *timelineQuery.Service
}

func NewHandler(
	ingestion IngestionClient,
	repo ReadOnlyRepository,
	ai AIClient,
	timelineSvc *timelineQuery.Service,
) *Handler {
	return &Handler{
		ingestion:   ingestion,
		repo:        repo,
		ai:          ai,
		timelineSvc: timelineSvc,
	}
}
