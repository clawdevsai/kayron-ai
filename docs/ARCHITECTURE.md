# MT5 MCP Integration Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Claude AI / Client Applications                                 │
└──────────────────────────────┬──────────────────────────────────┘
                               │ JSON-RPC 2.0
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│  MCP Tool Handler (JSON-RPC Router)                             │
│  - account_info, get_quote, place_order, close_position, etc.  │
└──────────────────────────────┬──────────────────────────────────┘
                               │ gRPC (port 50051)
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│  gRPC Daemon (mcp-mt5-server)                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ MT5 Trading Service Implementation                      │   │
│  │ - Account service                                       │   │
│  │ - Quote service                                         │   │
│  │ - Order service                                         │   │
│  │ - Position service                                      │   │
│  └────────────────────────┬─────────────────────────────────┘   │
│                           │                                      │
│  ┌────────────────────────▼─────────────────────────────────┐   │
│  │ Queue Manager                                           │   │
│  │ - Pending order queue (SQLite)                         │   │
│  │ - FIFO sequencing                                      │   │
│  │ - Idempotency cache                                    │   │
│  │ - Replay on reconnect                                  │   │
│  └────────────────────────┬─────────────────────────────────┘   │
│                           │                                      │
│  ┌────────────────────────▼─────────────────────────────────┐   │
│  │ Reconnect Handler                                       │   │
│  │ - Health monitoring (<10s heartbeat)                    │   │
│  │ - Auto-reconnect on disconnect                         │   │
│  │ - Queue replay after reconnect                         │   │
│  └────────────────────────┬─────────────────────────────────┘   │
│                           │ HTTP Client                          │
└───────────────────────────┼──────────────────────────────────────┘
                            │ (localhost:7788)
                            ▼
                  ┌──────────────────────┐
                  │  MT5 WebAPI Server   │
                  │  (MT5 Terminal)      │
                  └──────────────────────┘
```

## Component Details

### 1. MCP Tool Handler (Client Entry Point)

**Role**: Converts JSON-RPC 2.0 requests to gRPC calls

**Tools**:
- `account_info` → `MT5TradingService.AccountInfo`
- `get_quote` → `MT5TradingService.GetQuote`
- `place_order` → `MT5TradingService.PlaceOrder`
- `close_position` → `MT5TradingService.ClosePosition`
- `list_orders` → `MT5TradingService.ListOrders`

**Features**:
- Error translation (gRPC → JSON-RPC)
- Portuguese (pt-BR) error messages
- Request validation
- Latency tracking

---

### 2. gRPC Daemon

**Package**: `cmd/mcp-mt5-server/main.go`

**Responsibilities**:
- gRPC server (port 50051)
- Service orchestration
- Health check endpoint (`GET /health`)
- Metrics exposure
- Graceful shutdown

**Startup Sequence**:
```
1. Load config (env vars, TLS)
2. Load secrets (MT5 credentials)
3. Initialize services:
   - MT5 HTTP client
   - Health monitor
   - Queue manager
   - Latency tracker
4. Start gRPC server
5. Register health check
6. Start reconnect goroutine
```

---

### 3. MT5 Trading Service Implementation

**Location**: `internal/services/daemon/`

**Services**:

#### a. Account Service
- `AccountInfo()`: Fetch account balance, equity, margin
- Caches account state for 5 seconds
- Falls back to last good state on error

#### b. Quote Service
- `GetQuote()`: Fetch bid/ask for symbol
- Validates symbol exists
- Returns timestamp with quote

#### c. Order Service
- `PlaceOrder()`: Submit market order
- Enforces idempotency (exact-once delivery)
- Routes through queue for FIFO ordering
- Returns order ticket + entry price

#### d. Position Service
- `ClosePosition()`: Close open position by ticket
- Validates position exists
- Returns close price + P&L

#### e. Orders Service
- `ListOrders()`: Enumerate all open orders/positions
- Refreshes from MT5 on each call
- Groups by symbol for clarity

---

### 4. Queue Manager

**Location**: `internal/models/queue.go`

**Database**: SQLite (`queue.db`)

**Tables**:
```sql
CREATE TABLE orders (
  id TEXT PRIMARY KEY,
  idempotency_key TEXT UNIQUE,
  account_id TEXT,
  symbol TEXT,
  order_type TEXT,
  volume REAL,
  status TEXT,
  created_at TIMESTAMP,
  submitted_at TIMESTAMP,
  result_ticket INTEGER,
  error_message TEXT
);

CREATE TABLE processing_log (
  id INTEGER PRIMARY KEY,
  order_id TEXT,
  status TEXT,
  attempt INT,
  timestamp TIMESTAMP
);
```

**FIFO Ordering**:
- Orders processed in `created_at` sequence
- Mutex-protected for concurrency safety
- Atomic status transitions

**Idempotency**:
- Unique constraint on `idempotency_key`
- If duplicate received: return cached result
- Cache stored in same database, persisted across restarts

**Replay Logic**:
```go
func (q *Queue) ReplayPending() {
  pending := q.GetOrdersByStatus("PENDING")
  for _, order := range pending {
    result := q.SubmitToMT5(order)
    q.UpdateStatus(order.ID, result.Status)
  }
}
```

---

### 5. Reconnect Handler

**Location**: `internal/services/daemon/reconnect.go`

**Monitoring**:
- Heartbeat interval: 5 seconds
- Timeout threshold: 10 seconds
- Auto-reconnect: enabled

**State Machine**:
```
CONNECTED
    ↓ (no heartbeat for 10s)
DISCONNECTED → RECONNECTING → CONNECTED
    ↓                              ↑
    └──────────────────────────────┘
    (on successful reconnect)
```

**On Disconnect**:
1. Set `terminal_connected = false`
2. Return `UNAVAILABLE` for new requests
3. Accumulate new requests in queue
4. Attempt reconnect every 5 seconds

**On Reconnect**:
1. Set `terminal_connected = true`
2. Verify connection with health check
3. Replay all pending orders (FIFO)
4. Resume normal operation

**Implementation**:
```go
type ReconnectManager struct {
  lastHeartbeat     time.Time
  heartbeatInterval time.Duration
  timeoutThreshold  int64 // seconds
  mtClient          *mt5.Client
}

func (rm *ReconnectManager) Monitor() {
  ticker := time.NewTicker(5 * time.Second)
  for range ticker.C {
    if time.Since(rm.lastHeartbeat).Seconds() > float64(rm.timeoutThreshold) {
      rm.Reconnect() // Attempt reconnect
      rm.ReplayQueue() // Process pending orders
    }
  }
}
```

---

### 6. Error Handling & Observability

**Error Types**: `internal/errors/mt5_errors.go`

**gRPC Status Mapping**:
```
MT5 Error          → gRPC Code              → pt-BR Message
─────────────────────────────────────────────────────────────
Disconnect         → UNAVAILABLE            → Terminal desconectado
Timeout            → DEADLINE_EXCEEDED      → Tempo limite excedido
Invalid Creds      → UNAUTHENTICATED        → Credenciais inválidas
Margin Error       → INVALID_ARGUMENT       → Saldo insuficiente
Symbol Not Found   → NOT_FOUND              → Símbolo não encontrado
Price Gapping      → FAILED_PRECONDITION    → Abertura de preço
```

**Logging**:

JSON structured logging to `logs/app.json`:
```json
{
  "timestamp": "2026-05-08T10:30:45Z",
  "level": "INFO",
  "tool_name": "place_order",
  "account_id": "12345",
  "input": {...},
  "output": {...},
  "latency_ms": 1234,
  "error": null
}
```

Audit logging to `logs/audit.json`:
```json
{
  "timestamp": "2026-05-08T10:30:45Z",
  "event_type": "login_attempt",
  "account_id": "12345",
  "status": "success",
  "source": "env_var"
}
```

**Metrics**: `internal/logger/latency.go`

Exposed via `/health`:
```json
{
  "metrics": {
    "account_info": {
      "count": 150,
      "p50_ms": 250,
      "p95_ms": 1200,
      "p99_ms": 1800
    }
  }
}
```

---

### 7. Configuration & Security

**Config Loading**: `internal/config/config.go`

**Environment Variables**:
```
MT5_SERVER=localhost:7788
MT5_LOGIN=12345
MT5_PASSWORD=secure_password
MT5_SERVER_NAME=DemoServer
TLS_CERT_FILE=/path/to/cert.pem
TLS_KEY_FILE=/path/to/key.pem
LOG_FILE=logs/app.json
DEBUG=false
```

**Secrets Management**: `internal/config/secrets.go`

- Load from env vars (default)
- Load from AWS Secrets Manager (optional)
- Credential rotation support
- Never log credentials

**TLS Configuration**: `internal/config/tls.go`

- TLS disabled in dev mode
- TLS required in production
- Min TLS 1.2
- Strong cipher suites only

**Security Scanning**: `internal/security/cred_scan.go`

- Regex patterns for hardcoded credentials
- Scans source files in CI/CD
- Fails build if credentials found

---

## Data Flow Example: PlaceOrder

```
1. Client sends JSON-RPC request
   {
     "method": "place_order",
     "params": {
       "account_id": "12345",
       "symbol": "EURUSD",
       "idempotency_key": "order-001"
     }
   }

2. MCP Handler validates request
   - Extracts parameters
   - Validates idempotency key format

3. MCP Handler calls gRPC
   client.PlaceOrder(ctx, &PlaceOrderRequest{...})

4. gRPC Daemon receives request
   - Checks idempotency cache
   - If found: return cached result
   - If new: queue order

5. Queue Manager processes
   - Create database record
   - Assign sequence number
   - Log to processing_log

6. Order Service submits to MT5
   mt5.PlaceOrder(symbol, volume, ...)
   
7. MT5 returns ticket + price

8. Update queue record with result
   status = "COMPLETED"
   result_ticket = 12345678

9. Log to JSON logger
   latency_ms = 1234
   
10. Return gRPC response
    {
      "order_ticket": 12345678,
      "entry_price": 1.08560,
      "status": "FILLED"
    }

11. MCP Handler converts to JSON-RPC
    {
      "result": {
        "order_ticket": 12345678,
        ...
      }
    }

12. Send to client
```

---

## Performance Characteristics

**Latency SLAs**:
- AccountInfo: < 2s (p95: < 1.5s)
- GetQuote: < 500ms (p95: < 400ms)
- PlaceOrder: < 5s (includes queue processing)
- ClosePosition: < 5s
- ListOrders: < 2s

**Throughput**:
- ~100 concurrent requests (load tested)
- Queue can hold 1000+ pending orders
- FIFO ordering preserved under all loads

**Reliability**:
- Auto-reconnect on disconnect (< 15s recovery)
- Queue persistence (SQLite)
- Idempotency guarantees
- No orders lost on daemon restart

---

## Deployment Topology

**Single MT5 Terminal** (current):
- One daemon ↔ one MT5 terminal
- No clustering
- Simple scaling: run multiple daemons per terminal (not recommended)

**Future: Multi-Terminal** (v1.1):
- Multiple daemons per broker account
- Load balancing by symbol
- Account replication

