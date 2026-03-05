package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// type Hub struct {
// 	clients map[string]map[*websocket.Conn]bool // symbol -> connections
// 	mu      sync.RWMutex
// }

type Hub struct {
	orderBookClients map[string]map[*websocket.Conn]bool
	tradeClients     map[string]map[*websocket.Conn]bool
	mu               sync.RWMutex
}

//	func NewHub() *Hub {
//		return &Hub{
//			clients: make(map[string]map[*websocket.Conn]bool),
//		}
//	}
func NewHub() *Hub {
	return &Hub{
		orderBookClients: make(map[string]map[*websocket.Conn]bool),
		tradeClients:     make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *Hub) Register(symbol string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.orderBookClients[symbol] == nil {
		h.orderBookClients[symbol] = make(map[*websocket.Conn]bool)
	}

	h.orderBookClients[symbol][conn] = true
}

func (h *Hub) Unregister(symbol string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.orderBookClients[symbol], conn)
	conn.Close()
}

func (h *Hub) Broadcast(symbol string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.orderBookClients[symbol] {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (h *Hub) RegisterTrade(symbol string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.tradeClients[symbol] == nil {
		h.tradeClients[symbol] = make(map[*websocket.Conn]bool)
	}

	h.tradeClients[symbol][conn] = true
}

func (h *Hub) BroadcastTrade(symbol string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.tradeClients[symbol] {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}
