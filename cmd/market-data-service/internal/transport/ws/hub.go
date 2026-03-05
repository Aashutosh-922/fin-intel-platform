package ws

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]*Client
}

type Client struct {
	conn              *websocket.Conn
	writeMu           sync.Mutex
	symbols           map[string]struct{}
	topics            map[string]struct{}
	allSymbolsAllowed bool
	allTopicsAllowed  bool
}

type Envelope struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]*Client),
	}
}

func (h *Hub) Register(conn *websocket.Conn, symbols []string, topics []string) {
	client := &Client{
		conn:              conn,
		symbols:           make(map[string]struct{}),
		topics:            make(map[string]struct{}),
		allSymbolsAllowed: len(symbols) == 0,
		allTopicsAllowed:  len(topics) == 0,
	}
	for _, s := range symbols {
		client.symbols[strings.ToUpper(strings.TrimSpace(s))] = struct{}{}
	}
	for _, t := range topics {
		client.topics[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}

	h.mu.Lock()
	h.clients[conn] = client
	h.mu.Unlock()
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	_ = conn.Close()
}

func (h *Hub) Broadcast(topic string, payload []byte) {
	msg, err := json.Marshal(Envelope{
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		return
	}

	symbol := extractSymbol(payload)
	normalizedTopic := strings.ToLower(strings.TrimSpace(topic))

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for _, c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		if !c.acceptsTopic(normalizedTopic) || !c.acceptsSymbol(symbol) {
			continue
		}
		c.writeMu.Lock()
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		c.writeMu.Unlock()
		if err != nil {
			go h.Unregister(c.conn)
		}
	}
}

func (c *Client) acceptsTopic(topic string) bool {
	if c.allTopicsAllowed {
		return true
	}
	_, ok := c.topics[topic]
	return ok
}

func (c *Client) acceptsSymbol(symbol string) bool {
	if symbol == "" || c.allSymbolsAllowed {
		return true
	}
	_, ok := c.symbols[symbol]
	return ok
}

func extractSymbol(payload []byte) string {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return ""
	}
	raw, ok := m["symbol"]
	if !ok {
		return ""
	}
	symbol, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(symbol))
}
