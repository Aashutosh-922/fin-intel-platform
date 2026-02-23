package http

import (
	"context"
	"net/http"
	"database/sql"

	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ingestion-service/internal/kafka"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// type Server struct {
//     cfg      config.Config
//     logger   Logger
//     http     *http.Server
//     producer *kafka.Producer
// }
type Server struct {
	cfg      config.Config
	logger   Logger
	producer *kafka.Producer
	db       *sql.DB
	http     *http.Server
}


func NewServer(
	cfg config.Config,
	logger Logger,
	producer *kafka.Producer,
	db *sql.DB,
) *Server {
	s := &Server{
		cfg:      cfg,
		logger:   logger,
		producer: producer,
		db: db,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/transactions", s.createTransaction)

	s.http = &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down ingestion http server")
	return s.http.Shutdown(ctx)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}