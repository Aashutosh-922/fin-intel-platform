package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/domain"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/infrasturcture/kafka"
	otrace "github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/trace"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service  *application.Service
	producer *kafka.Producer
}

type createOrderRequest struct {
	OrderID  string  `json:"order_id"`
	UserID   string  `json:"user_id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Type     string  `json:"type"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

func NewHandler(service *application.Service, producer *kafka.Producer) *Handler {
	return &Handler{
		service:  service,
		producer: producer,
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := domain.Order{
		OrderID:  strings.TrimSpace(req.OrderID),
		UserID:   strings.TrimSpace(req.UserID),
		Symbol:   strings.ToUpper(strings.TrimSpace(req.Symbol)),
		Side:     strings.ToUpper(strings.TrimSpace(req.Side)),
		Type:     domain.OrderType(strings.ToUpper(strings.TrimSpace(req.Type))),
		Price:    req.Price,
		Quantity: req.Quantity,
	}

	if order.OrderID == "" || order.UserID == "" || order.Symbol == "" || order.Side == "" || order.Type == "" {
		http.Error(w, "missing required order fields", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateOrder(withTraceID(r), order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {

	orderID := chi.URLParam(r, "order_id")
	symbol := r.URL.Query().Get("symbol")
	if strings.TrimSpace(orderID) == "" || strings.TrimSpace(symbol) == "" {
		http.Error(w, "order_id and symbol are required", http.StatusBadRequest)
		return
	}

	event := domain.OrderCancel{
		OrderID: strings.TrimSpace(orderID),
		Symbol:  strings.ToUpper(strings.TrimSpace(symbol)),
	}

	if err := h.producer.PublishCancel(withTraceID(r), event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func withTraceID(r *http.Request) context.Context {
	traceID := strings.TrimSpace(r.Header.Get("X-Trace-ID"))
	if traceID == "" {
		traceID = uuid.NewString()
	}
	return otrace.WithTraceID(r.Context(), traceID)
}
