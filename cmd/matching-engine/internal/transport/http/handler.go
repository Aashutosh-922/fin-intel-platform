package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/application"
)

type Handler struct {
	matcher *application.Matcher
}

func NewHandler(m *application.Matcher) *Handler {
	return &Handler{matcher: m}
}

func (h *Handler) GetOrderBook(w http.ResponseWriter, r *http.Request) {

	symbol := chi.URLParam(r, "symbol")

	snapshot := h.matcher.GetSnapshot(symbol, 10)

	json.NewEncoder(w).Encode(snapshot)
}