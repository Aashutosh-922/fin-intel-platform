// package http

// import (
// 	"context"
// 	"net/http"

// 	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
// 	"log/slog"
// )

// type Server struct {
// 	cfg    config.Config
// 	logger *slog.Logger
// 	http   *http.Server
// }

// func NewServer(cfg config.Config, logger *slog.Logger) *Server {
// 	return &Server{
// 		cfg:    cfg,
// 		logger: logger,
// 	}
// }

// func (s *Server) Start() error {
// 	mux := http.NewServeMux()

// 	// routes
// 	mux.HandleFunc("/health", s.health)
// 	mux.HandleFunc("/transactions", s.createTransaction)

// 	s.http = &http.Server{
// 		Addr:    ":" + s.cfg.HTTPPort,
// 		Handler: mux,
// 	}

// 	return s.http.ListenAndServe()
// }

// func (s *Server) Shutdown(ctx context.Context) error {
// 	s.logger.Info("shutting down http server")
// 	return s.http.Shutdown(ctx)
// }

// func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("ok"))
// }

package http

import (
	"context"
	"net/http"

	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
	"log/slog"
)

type Server struct {
	cfg    config.Config
	logger *slog.Logger
	http   *http.Server
}

func NewServer(cfg config.Config, logger *slog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/transactions", s.createTransaction)

	s.http = &http.Server{
		Addr:    ":" + s.cfg.HTTPPort,
		Handler: mux,
	}

	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.http.Shutdown(ctx)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("accepted"))
}
