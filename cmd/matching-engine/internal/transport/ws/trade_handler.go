package ws

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) TradeStream(w http.ResponseWriter, r *http.Request) {

	symbol := chi.URLParam(r, "symbol")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.RegisterTrade(symbol, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			conn.Close()
			break
		}
	}
}
