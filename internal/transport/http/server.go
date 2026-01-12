package http

import (
	"net/http"

	"github.com/Aashutosh-922/fin-intel-platform/internal/config"
	"log/slog"
)

type Server struct {
	cfg    config.Config
	logger *slog.Logger
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

	s.logger.Info("http server started", "port", s.cfg.HTTPPort)
	return http.ListenAndServe(":"+s.cfg.HTTPPort, mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
