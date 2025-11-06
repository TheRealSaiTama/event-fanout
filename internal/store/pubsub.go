package store

import (
	"context"
	"encoding/json"
	"log"
)

type Broadcaster interface {
	Broadcast(room string, msg []byte)
}

type PubSub struct {
	r *Redis
	b Broadcaster
}

func NewPubSub(r *Redis, b Broadcaster) *PubSub { return &PubSub{r: r, b: b} }

type Message struct {
	Room string `json:"room"`
	From string `json:"from"`
	Data string `json:"data"`
}

func (ps *PubSub) Publish(ctx context.Context, m Message) {
	b, _ := json.Marshal(m)
	if err := ps.r.Publish(ctx, RoomChannel(m.Room), b).Err(); err != nil {
		log.Printf("redis publish error: %v", err)
	}
}

func (ps *PubSub) Run() {
	ctx := context.Background()
	pubsub := ps.r.PSubscribe(ctx, "room:*")
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil { log.Printf("redis psubscribe error: %v", err); continue }
		var m Message
		if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil { continue }
		room := ParseRoom(msg.Channel)
		if room == "" { room = m.Room }
		if room == "" { continue }
		ps.b.Broadcast(room, []byte(m.Data))
	}
}
