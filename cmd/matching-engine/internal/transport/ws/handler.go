package ws

import (
	"encoding/json"
	"net/http"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type SnapshotProvider interface {
	GetSnapshot(symbol string, depth int) domain.OrderBookSnapshot
}

type Handler struct {
	matcher SnapshotProvider
	hub     *Hub
}

func NewHandler(m SnapshotProvider, h *Hub) *Handler {
	return &Handler{matcher: m, hub: h}
}

func (h *Handler) OrderBookStream(w http.ResponseWriter, r *http.Request) {

	symbol := chi.URLParam(r, "symbol")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.Register(symbol, conn)

	// Send initial snapshot
	snapshot := h.matcher.GetSnapshot(symbol, 10)
	data, _ := json.Marshal(snapshot)
	conn.WriteMessage(websocket.TextMessage, data)

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			h.hub.Unregister(symbol, conn)
			break
		}
	}
}
