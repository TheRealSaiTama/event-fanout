.PHONY: build run bench

build:
	go build -o bin/event-fanout ./cmd/server

run:
	ADDR=:8081 REDIS_ADDR=localhost:6379 go run ./cmd/server

bench:
	k6 run bench/ws_broadcast.js
