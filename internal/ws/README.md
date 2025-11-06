
# ws package

Primitives:

* **Hub**: room registry + global metrics
* **Room**: client set, broadcast, drop-on-backpressure
* **Client**: WS conn, bounded outbox, `readPump`/`writePump`
* **Handler**: `GET /ws?room=&client_id=`

Back-pressure: non-blocking `TrySend` → queue full ⇒ `Close()` & drop counted.
