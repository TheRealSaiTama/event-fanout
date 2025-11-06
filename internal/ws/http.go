package ws

import (
	"log"
	"net/http"

	"github.com/TheRealSaiTama/event-fanout/internal/store"
)

// Handler upgrades to WebSocket and joins a room.
// GET /ws?room=<name>&client_id=<id>
func Handler(h *Hub, ps *store.PubSub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		if room == "" {
			http.Error(w, "missing room", http.StatusBadRequest)
			return
		}
		clientID := r.URL.Query().Get("client_id")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return
		}

		c := newClient(conn, room, h.Leave)
		h.Join(room, c)
		go c.writePump()
		go c.readPump(ps, clientID)
	})
}
