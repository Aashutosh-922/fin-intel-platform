// package handlers

// import (
// 	"encoding/json"
// 	"net/http"

// 	"github.com/go-chi/chi/v5"
// )

// func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
// 	reqBody, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "invalid body", http.StatusBadRequest)
// 		return
// 	}

// 	resp, err := h.IngestionClient.Forward(r.Context(), reqBody)
// 	if err != nil {
// 		http.Error(w, "ingestion failed", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusAccepted)
// 	w.Write(resp)
// }

// func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	var req CreateTransactionRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "invalid request", http.StatusBadRequest)
// 		return
// 	}

// 	// 🔴 THIS CALL MUST NOT BLOCK
// 	if err := h.IngestionClient.CreateTransaction(ctx, req); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadGateway)
// 		return
// 	}

// 	// 🔴 THIS LINE IS MANDATORY
// 	w.WriteHeader(http.StatusAccepted)
// }

// func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "id")

// 	tx, _ := h.Repo.GetTransaction(id)
// 	risk, _ := h.Repo.GetRisk(id)
// 	ai, _ := h.Repo.GetExplanation(id)

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"id":          tx.ID,
// 		"amount":      tx.Amount,
// 		"status":      tx.Status,
// 		"risk_score":  risk.Score,
// 		"explanation": ai.Text,
// 	})
// }

// func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "id")

// 	events, _ := h.Repo.GetEvents(id)

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(events)
// }

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.ingestion.CreateTransaction(r.Context(), req); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	tx, err := h.repo.GetTransaction(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch transaction", http.StatusInternalServerError)
		return
	}
	risk, err := h.repo.GetRisk(id)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "failed to fetch risk", http.StatusInternalServerError)
		return
	}
	ai, err := h.repo.GetExplanation(id)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "failed to fetch explanation", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          tx.ID,
		"amount":      tx.Amount,
		"status":      tx.Status,
		"risk_score":  risk.Score,
		"explanation": ai.Text,
	})
}

func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	result, err := h.timelineSvc.GetTimeline(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
