
# store package

Thin Redis wrapper + Pub/Sub loop. Depends on a **Broadcaster** interface:

```go
type Broadcaster interface { Broadcast(room string, msg []byte) }
```

This avoids an import cycle between `store` and `ws`.
