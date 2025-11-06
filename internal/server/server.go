package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/TheRealSaiTama/event-fanout/internal/store"
	"github.com/TheRealSaiTama/event-fanout/internal/ws"
)

type Server struct {
	addr      string
	mux       *http.ServeMux
	httpSrv   *http.Server
	hub       *ws.Hub
	redisAddr string
	redisPass string
}

func New(addr, redisAddr, redisPass string) *Server {
	mux := http.NewServeMux()
	hub := ws.NewHub()
	redis := store.NewRedis(redisAddr, redisPass)
	ps := store.NewPubSub(redis, hub)
	go ps.Run()

	s := &Server{
		addr:      addr,
		mux:       mux,
		hub:       hub,
		redisAddr: redisAddr,
		redisPass: redisPass,
	}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := store.Health(r.Context(), redis); err != nil { w.WriteHeader(http.StatusServiceUnavailable); w.Write([]byte("redis down")); return }
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/metrics", s.metricsHandler)
	mux.Handle("/ws", ws.Handler(hub, ps))

	s.httpSrv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Start() error { return s.httpSrv.ListenAndServe() }

func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("stopping hub...")
	s.hub.Stop()
	log.Printf("closing http server...")
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	m := s.hub.Snapshot()
	fmt.Fprintf(w, "rooms %d\nclients %d\ndropped_messages %d\n", m.Rooms, m.Clients, m.Dropped)
}
