package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheRealSaiTama/event-fanout/internal/server"
)

func main() {
	addr := getenv("ADDR", ":8081")
	redisURL := getenv("REDIS_ADDR", "redis:6379")
	redisPass := os.Getenv("REDIS_PASSWORD")

	srv := server.New(addr, redisURL, redisPass)
	go func() {
		log.Printf("event-fanout listening on %s", addr)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()
	log.Printf("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Printf("bye")
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" { return v }
	return d
}
