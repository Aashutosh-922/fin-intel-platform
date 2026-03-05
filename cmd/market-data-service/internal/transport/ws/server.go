package ws

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

func StartServer(httpPort string, hub *Hub) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	mux.HandleFunc("/ws/market", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		symbols := parseCSVQuery(r, "symbols")
		topics := parseCSVQuery(r, "topics")
		hub.Register(conn, symbols, topics)

		go func() {
			defer hub.Unregister(conn)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
	})

	log.Printf("market-data websocket server listening on :%s\n", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatal(err)
	}
}

func parseCSVQuery(r *http.Request, key string) []string {
	values := r.URL.Query()[key]
	out := make([]string, 0)
	for _, v := range values {
		for _, item := range strings.Split(v, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}
