# Lab 9 — Concurrent TCP Server in Go

A concurrent TCP echo server and load-testing client demonstrating goroutine-based concurrency, connection timeouts, input validation, and resource exhaustion analysis.

---

## Files

| File | Description |
|------|-------------|
| `server.go` | Concurrent TCP server with timeouts, connection cap, logging, metrics |
| `client.go` | Configurable load-testing client |
| `report.md` | Crash analysis, debugging evidence, metrics, and mitigation |

---

## Requirements

- Go 1.18 or later (`go version` to check)
- Linux or macOS recommended (`ss`, `lsof` available)

---

## How to Run

### 1. Start the Server

```bash
go run server.go
```

Or build and run:

```bash
go build -o server server.go
./server
```

Expected output:
```
[SERVER] Listening on 127.0.0.1:8080 | timeout=10s max_conns=500
[METRICS] goroutines=3 active_conns=0 total_conns=0
```

### 2. Run the Load Testing Client

Open a second terminal:

```bash
go run client.go
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-clients` | 100 | Number of concurrent goroutines/connections |
| `-msg` | `LOAD_TEST_PING` | Message sent by each client |
| `-timeout` | 5s | Per-client dial and read timeout |
| `-waves` | 1 | Number of sequential load waves |

**Examples:**

```bash
# Default: 100 concurrent clients, 1 wave
go run client.go

# Stress test: 300 clients across 3 waves
go run client.go -clients 300 -waves 3

# Custom message and timeout
go run client.go -clients 50 -msg "HELLO" -timeout 3s
```

### 3. Test a Single Connection (Manual)

```bash
echo "Hello Server" | nc 127.0.0.1 8080
```

---

## Expected Behavior Under Load

### Normal Load (< 200 clients)

- Server accepts all connections
- Each connection handled in its own goroutine
- Stats log shows stable goroutine count (~client count + 3 base)
- Client reports: all successes, low latency

```
[METRICS] goroutines=103 active_conns=100 total_conns=100
========== LOAD TEST RESULTS ==========
Total clients  : 100
Successes      : 100
Failures       : 0
Timeouts       : 0
Throughput     : ~800 req/s
```

### High Load (approaching MAX_CONNS = 500)

- Server logs `[WARN]` when goroutine count exceeds 300
- Some clients may hit the connection cap and receive `ERROR: server at capacity`
- Goroutine count grows proportionally to active connections, then stabilises

### Timeout Behavior

Clients that connect but stall receive no data from the server. After 10 seconds (`CONN_TIMEOUT`), the server closes the connection automatically:

```
[DISCONNECT] 127.0.0.1:XXXXX | i/o timeout
```

This prevents Slowloris-style resource exhaustion.

### Connection Cap Behavior

When `active_conns >= 500`, new connections are immediately rejected:

```
[LIMIT] Cap reached (500), rejecting 127.0.0.1:XXXXX
```

---

## Observability Commands

While the server is running, in another terminal:

```bash
# View open sockets
ss -tuln | grep 8080

# Count file descriptors used by server process
lsof -i :8080 | wc -l

# Watch goroutine/connection metrics live (server logs to stdout)
```

---

## Stopping the Server

Press `Ctrl+C` — the server shuts down gracefully:

```
[SERVER] Shutting down...
```
