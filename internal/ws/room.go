package ws

import "sync"

type Room struct {
	name    string
	mu      sync.RWMutex
	clients map[*Client]struct{}
	closing chan struct{}
	onDrop  func(uint64)
}

func NewRoom(name string, onDrop func(uint64)) *Room {
	return &Room{name: name, clients: make(map[*Client]struct{}), closing: make(chan struct{}), onDrop: onDrop}
}

func (r *Room) Run() { <-r.closing }

func (r *Room) Join(c *Client) {
	r.mu.Lock(); r.clients[c] = struct{}{}; r.mu.Unlock()
}

func (r *Room) Leave(c *Client) {
	r.mu.Lock(); delete(r.clients, c); r.mu.Unlock()
}

func (r *Room) Broadcast(msg []byte) {
	r.mu.RLock()
	toDrop := make([]*Client, 0, 8)
	for c := range r.clients {
		if !c.TrySend(msg) { toDrop = append(toDrop, c) }
	}
	r.mu.RUnlock()

	if len(toDrop) > 0 {
		r.mu.Lock()
		for _, c := range toDrop {
			if _, ok := r.clients[c]; ok {
				c.Close()
				delete(r.clients, c)
			}
		}
		r.mu.Unlock()
		if r.onDrop != nil { r.onDrop(uint64(len(toDrop))) }
	}
}

func (r *Room) Stop() { close(r.closing) }
