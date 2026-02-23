// package query

// import (
// 	"encoding/json"
// 	"net/http"

// 	"github.com/go-chi/chi/v5"
// )

// type Handler struct {
// 	svc *Service
// }

// func NewHandler(svc *Service) *Handler {
// 	return &Handler{svc: svc}
// }

// func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "id")

// 	events, err := h.svc.Timeline(r.Context(), id)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}

// 	resp := map[string]interface{}{
// 		"transaction_id": id,
// 		"events":         events,
// 	}

// 	json.NewEncoder(w).Encode(resp)
// }
package query

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.svc.GetTimeline(r.Context(), id)
	if err != nil {
		http.Error(w, "timeline not found", 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}