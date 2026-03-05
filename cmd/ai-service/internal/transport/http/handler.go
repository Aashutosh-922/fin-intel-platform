package http

import (
	"encoding/json"
	"net/http"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) AnalyzeTransaction(w http.ResponseWriter, r *http.Request) {
	var event domain.RiskDecisionEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.service.Process(r.Context(), event)
	if err != nil {
		http.Error(w, "analysis failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
