// package handlers

// import (
// 	"encoding/json"
// 	"net/http"
// )

// func (h *Handler) AIQuery(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		Question string `json:"question"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "invalid request", http.StatusBadRequest)
// 		return
// 	}

// 	resp, err := h.AIClient.Query(r.Context(), req.Question)
// 	if err != nil {
// 		http.Error(w, "ai failed", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type AIRequest struct {
	TransactionID string `json:"transaction_id"`
	Question      string `json:"question"`
}

func (h *Handler) AIQuery(w http.ResponseWriter, r *http.Request) {
	var req AIRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(req.TransactionID); err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	resp, err := h.ai.Query(r.Context(), AIQuery{
		TransactionID: req.TransactionID,
		Question:      req.Question,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "ai insight not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
