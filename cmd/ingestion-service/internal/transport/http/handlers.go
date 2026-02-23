package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Aashutosh-922/fin-intel-platform/internal/events"
)

type TransactionRequest struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var req TransactionRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// =========================
// 	// ✅ STRICT VALIDATION
// 	// =========================

// 	req.UserID = strings.TrimSpace(req.UserID)
// 	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))

// 	if req.UserID == "" {
// 		http.Error(w, "user_id is required", http.StatusBadRequest)
// 		return
// 	}

// 	if req.Amount <= 0 {
// 		http.Error(w, "amount must be > 0", http.StatusBadRequest)
// 		return
// 	}

// 	allowedCurrencies := map[string]bool{
// 		"INR": true,
// 		"USD": true,
// 		"EUR": true,
// 	}

// 	if !allowedCurrencies[req.Currency] {
// 		http.Error(w, "unsupported currency", http.StatusBadRequest)
// 		return
// 	}

// 	// =========================
// 	// ✅ CANONICAL EVENT
// 	// =========================

// 	event := events.TransactionEvent{
// 		Version:    "v1",
// 		EventID:    uuid.NewString(),
// 		UserID:     req.UserID,
// 		Amount:     req.Amount,
// 		Currency:   req.Currency,
// 		OccurredAt: time.Now().UTC(),
// 		Source:     "api",
// 	}

// 	// if err := s.producer.Publish(r.Context(), event); err != nil {
// 	// 	http.Error(w, "failed to publish transaction", http.StatusInternalServerError)
// 	// 	return
// 	// }
// 	payload, _ := json.Marshal(event)

// _, err := s.db.ExecContext(r.Context(), `
// 	INSERT INTO outbox_events (id, topic, payload)
// 	VALUES ($1, $2, $3)
// `,
// 	event.EventID,
// 	"transactions",
// 	payload,
// )

// if err != nil {
// 	http.Error(w, "failed to store transaction", http.StatusInternalServerError)
// 	return
// }

// 	w.WriteHeader(http.StatusAccepted)
// 	w.Write([]byte("accepted"))
// }

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))

	if req.UserID == "" || req.Amount <= 0 {
		http.Error(w, "invalid transaction fields", http.StatusBadRequest)
		return
	}

	evt := events.TransactionEvent{
		Version:    "v1",
		EventID:    uuid.NewString(),
		UserID:     req.UserID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		OccurredAt: time.Now().UTC(),
		Source:     "api",
	}

	// -----------------------------
	// 1️⃣ STORE IN DB FIRST
	// -----------------------------
	_, err := s.db.ExecContext(
		context.Background(),
		`INSERT INTO transactions (id, user_id, amount, currency, status)
		 VALUES ($1,$2,$3,$4,'RECEIVED')`,
		evt.EventID,
		evt.UserID,
		evt.Amount,
		evt.Currency,
	)
	if err != nil {
		s.logger.Error("db insert failed", "err", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// -----------------------------
	// 2️⃣ THEN PUBLISH TO KAFKA
	// -----------------------------
	if err := s.producer.Publish(r.Context(), evt); err != nil {
		s.logger.Error("kafka publish failed", "err", err)
		http.Error(w, "failed to publish transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("accepted"))
}