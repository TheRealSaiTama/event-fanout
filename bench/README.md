# Bench Suite

## Go Latency Bench

```bash
# from repo root
export GOCACHE="$PWD/.gocache" GOMODCACHE="$PWD/.gomodcache"
go run bench/latency_bench.go -n 200 -rate 200 -dur 30s -base ws://localhost:8081/ws -room room1 | tee bench.out
```

Sample result:

```
=== BENCH RESULTS ===
clients=200, sent=6000, received=119800
throughput=3992 msgs/s, p50=0.6 ms, p95=0.9 ms, p99=1.5 ms
```

## k6 WebSocket Broadcast

```bash
make bench   # requires k6 installed
```

### Tips

* Increase file limits for >1k clients: `ulimit -n 65535`
* Tune `outBuf` in `internal/ws/client.go` for back-pressure experiments
