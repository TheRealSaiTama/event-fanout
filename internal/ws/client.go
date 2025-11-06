package ws

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/TheRealSaiTama/event-fanout/internal/store"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn   *websocket.Conn
	room   string
	out    chan []byte
	closed atomic.Bool
	leave  func(string, *Client)
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
	outBuf     = 256
)

func newClient(conn *websocket.Conn, room string, leave func(string,*Client)) *Client {
	return &Client{ conn: conn, room: room, out: make(chan []byte, outBuf), leave: leave }
}

func (c *Client) TrySend(b []byte) bool {
	select {
	case c.out <- b:
		return true
	default:
		return false
	}
}

func (c *Client) Close() {
	if c.closed.CompareAndSwap(false, true) {
		_ = c.conn.Close()
		c.leave(c.room, c)
	}
}

func (c *Client) readPump(ps *store.PubSub, from string) {
	defer c.Close()
	c.conn.SetReadLimit(1 << 20)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil
	})
	for {
		t, msg, err := c.conn.ReadMessage()
		if err != nil { return }
		if t == websocket.TextMessage || t == websocket.BinaryMessage {
			ps.Publish(context.Background(), store.Message{Room: c.room, From: from, Data: string(msg)})
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func(){ ticker.Stop(); c.Close() }()
	for {
		select {
		case b, ok := <-c.out:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { return }
			if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil { return }
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil { return }
		}
	}
}
