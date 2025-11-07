# event-fanout — WebSocket Fan-out Service (Go + Redis Pub/Sub)

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://go.dev/)
[![CI](https://github.com/TheRealSaiTama/event-fanout/actions/workflows/ci.yml/badge.svg)](https://github.com/TheRealSaiTama/event-fanout/actions/workflows/ci.yml)
[![Dockerized](https://img.shields.io/badge/Docker-ready-2496ED.svg)](https://docs.docker.com/)

A tiny, fast fan-out gateway written in Go. It uses goroutines/channels with **room-based broadcast**, **heartbeats**, and **back-pressure-aware writers**. Horizontal fan-out is done via **Redis Pub/Sub** (swappable), plus a **sharded channel registry** and **graceful shutdown** with connection drain.

## Quickstart

```bash
# Redis via Docker
docker run -d --name ef-redis -p 6379:6379 redis:7-alpine

# Build & run
go mod tidy
make build
ADDR=:8081 REDIS_ADDR=localhost:6379 ./bin/event-fanout

# Smoke
curl -s http://localhost:8081/healthz
curl -s http://localhost:8081/metrics
````

### Docker Compose

```yaml
services:
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
  event-fanout:
    build: .
    environment:
      - ADDR=:8081
      - REDIS_ADDR=redis:6379
    ports: ["8081:8081"]
    depends_on: [redis]
```

## API

* `GET /healthz` → `"ok"` if Redis reachable
* `GET /metrics` → `rooms <n>\nclients <n>\ndropped_messages <n>`
* `GET /ws?room=<name>&client_id=<id>` → WebSocket; sending text frames publishes to the room via Redis.

## Architecture (tiny tour)

* **Hub** → manages rooms; tracks clients and drop counts.
* **Room** → in-memory set of clients; broadcast with non-blocking writes and drop on back-pressure.
* **Client** → ping/pong heartbeats, write & read pumps, buffered outbox.
* **Redis Pub/Sub** → fan-out across processes; messages published on `room:<name>`.

## Bench Results (2025-11-06, local)

* **Clients (K)**: 200
* **Throughput**: ~3992 msgs/s
* **Latency**: p50 ≈ 0.6 ms, **p95 ≈ 0.9 ms**, p99 ≈ 1.5 ms

> Method: custom Go WS bench (`bench/latency_bench.go`) publishing timestamps to a single room on localhost (Redis on Docker).

## Reproduce the bench

```bash
export GOCACHE="$PWD/.gocache" GOMODCACHE="$PWD/.gomodcache"
go run bench/latency_bench.go -n 200 -rate 200 -dur 30s -base ws://localhost:8081/ws -room room1
```

## Production hardening (future work)

* Auth / JWT on `Upgrade`
* Per-room sharding keys, pooled publishers
* NATS / Kafka swap layer
* Prometheus metrics + dashboards
* K8s manifests (readiness/liveness, HPA)
* Soak tests (k6, Vegeta) at scale

## License

MIT — see [LICENSE](./LICENSE).
