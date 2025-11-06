package ws

import (
	"sync"
	"sync/atomic"
)

type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]*Room
	closing chan struct{}
	metrics struct {
		rooms   int64
		clients int64
		dropped uint64
	}
}

func NewHub() *Hub {
	return &Hub{ rooms: make(map[string]*Room), closing: make(chan struct{}) }
}

func (h *Hub) Room(name string) *Room {
	h.mu.Lock(); defer h.mu.Unlock()
	r, ok := h.rooms[name]
	if !ok {
		r = NewRoom(name, func(n uint64){ atomic.AddUint64(&h.metrics.dropped, n) })
		h.rooms[name] = r
		atomic.StoreInt64(&h.metrics.rooms, int64(len(h.rooms)))
		go r.Run()
	}
	return r
}

func (h *Hub) Join(room string, c *Client) {
	h.Room(room).Join(c)
	atomic.AddInt64(&h.metrics.clients, 1)
}

func (h *Hub) Leave(room string, c *Client) {
	if r := h.get(room); r != nil { r.Leave(c) }
	atomic.AddInt64(&h.metrics.clients, -1)
}

func (h *Hub) Broadcast(room string, msg []byte) {
	if r := h.get(room); r != nil { r.Broadcast(msg) }
}

func (h *Hub) Stop() {
	close(h.closing)
	h.mu.Lock()
	for _, r := range h.rooms { r.Stop() }
	h.mu.Unlock()
}

func (h *Hub) get(name string) *Room {
	h.mu.RLock(); defer h.mu.RUnlock()
	return h.rooms[name]
}

type SnapshotMetrics struct {
	Rooms   int
	Clients int
	Dropped uint64
}
func (h *Hub) Snapshot() SnapshotMetrics {
	return SnapshotMetrics{
		Rooms:   int(atomic.LoadInt64(&h.metrics.rooms)),
		Clients: int(atomic.LoadInt64(&h.metrics.clients)),
		Dropped: atomic.LoadUint64(&h.metrics.dropped),
	}
}
