package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Aashutosh-922/fin-intel-platform/internal/application/ingest"
	"github.com/Aashutosh-922/fin-intel-platform/internal/domain/transaction"
	"github.com/google/uuid"
)

type IngestionHandler struct {
	service *ingest.Service
}

func NewIngestionHandler(service *ingest.Service) *IngestionHandler {
	return &IngestionHandler{service: service}
}

func (h *IngestionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	idempoKey := r.Header.Get("Idempotency-Key")
	if idempoKey == "" {
		http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Amount <= 0 {
		http.Error(w, "invalid request fields", http.StatusBadRequest)
		return
	}

	tx := transaction.Transaction{
		ID:        uuid.NewString(),
		UserID:    req.UserID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Country:   req.Country,
		Status:    "RECEIVED",
		CreatedAt: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resultTx, err := h.service.Ingest(ctx, idempoKey, tx)
	if err != nil {
		http.Error(w, "failed to ingest transaction", http.StatusInternalServerError)
		return
	}

	resp := CreateTransactionResponse{
		TransactionID: resultTx.ID,
		Status:        resultTx.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}
