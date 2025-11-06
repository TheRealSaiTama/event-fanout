package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	base := flag.String("base", "ws://localhost:8081/ws", "WS base URL")
	room := flag.String("room", "bench", "room name")
	n := flag.Int("n", 200, "number of clients")
	rate := flag.Int("rate", 200, "messages per second from publisher")
	dur := flag.Duration("dur", 30*time.Second, "test duration")
	flag.Parse()

	url := fmt.Sprintf("%s?room=%s&client_id=", *base, *room)

	// Connect clients
	type client struct{ c *websocket.Conn }
	clients := make([]client, *n)
	var wg sync.WaitGroup
	dial := func(id int) *websocket.Conn {
		u := url + fmt.Sprintf("c%d", id)
		c, _, err := websocket.DefaultDialer.Dial(u, nil)
		if err != nil { log.Fatalf("dial %d: %v", id, err) }
		return c
	}
	for i := 0; i < *n; i++ {
		clients[i] = client{c: dial(i)}
	}

	// Collector
	lats := make([]float64, 0, 100000)
	var mu sync.Mutex
	addLat := func(ms float64) { mu.Lock(); lats = append(lats, ms); mu.Unlock() }

	// Readers
	stop := make(chan struct{})
	for i := range clients {
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_, msg, err := c.ReadMessage()
					if err != nil { return }
					// payload is unix ns timestamp in text (int64)
					var ts int64
					_, err = fmt.Sscanf(string(msg), "%d", &ts)
					if err == nil {
						lat := float64(time.Since(time.Unix(0, ts)).Microseconds()) / 1000.0
						if !math.IsNaN(lat) && !math.IsInf(lat, 0) { addLat(lat) }
					}
				}
			}
		}(clients[i].c)
	}

	// Publisher
	pub, _, err := websocket.DefaultDialer.Dial(url+"pub", nil)
	if err != nil { log.Fatalf("dial pub: %v", err) }
	ticker := time.NewTicker(time.Second / time.Duration(*rate))
	defer ticker.Stop()

	start := time.Now()
	sendCount := 0
	recvTicker := time.NewTicker(1 * time.Second)
	defer recvTicker.Stop()
	_ = pub.SetWriteDeadline(time.Now().Add(3 * time.Second))

	for time.Since(start) < *dur {
		select {
		case <-ticker.C:
			ts := time.Now().UnixNano()
			_ = pub.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d", ts)))
			sendCount++
		case <-recvTicker.C:
			// keep loop responsive
		}
	}

	// Stop
	close(stop)
	for _, cl := range clients { _ = cl.c.Close() }
	_ = pub.Close()
	wg.Wait()

	// Stats
	mu.Lock()
	sort.Float64s(lats)
	total := len(lats)
	p := func(q float64) float64 {
		if total == 0 { return math.NaN() }
		i := int(math.Ceil(q*float64(total))) - 1
		if i < 0 { i = 0 }
		if i >= total { i = total-1 }
		return lats[i]
	}
	p50, p95, p99 := p(0.50), p(0.95), p(0.99)
	elapsed := time.Since(start).Seconds()
	throughput := float64(total) / elapsed
	fmt.Printf("\n=== BENCH RESULTS ===\n")
	fmt.Printf("clients=%d, sent=%d, received=%d\n", *n, sendCount, total)
	fmt.Printf("throughput=%.0f msgs/s, p50=%.1f ms, p95=%.1f ms, p99=%.1f ms\n", throughput, p50, p95, p99)
	mu.Unlock()
}
