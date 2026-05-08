# MT5 MCP Evolution Plan: 10 New Features

**Status**: Planning Phase  
**Date**: 2026-05-08  
**Current**: 6 tools implemented (account-info, quote, place-order, close-position, orders-list, get-candles)  
**Target**: 16 total tools (6 core + 10 evolution)

---

## Evolution Roadmap

### Phase 1: Core Tools ✅ COMPLETE
- ✅ account-info
- ✅ quote
- ✅ place-order
- ✅ close-position
- ✅ orders-list
- ✅ get-candles

### Phase 2: Order Management (Priority: HIGH)
Low complexity, direct MT5 mapping, foundation for advanced features.

| Feature | Complexity | Dependencies | Est. Tasks |
|---------|-----------|--------------|-----------|
| **1. modify-order** | Medium | MT5 /order/{id} endpoint | 8 |
| **2. pending-order-details** | Medium | Filtered query logic | 7 |
| **3. symbol-properties** | Low | MT5 /symbols endpoint | 5 |

**Rationale**: Enable order lifecycle management (create → modify → close). Symbol properties needed for validation across all tools.

**MVP Acceptance**:
- modify-order: Change pending order price/SL/TP without closing
- pending-order-details: Filter orders by symbol, status, date range
- symbol-properties: Return pip value, lot min/max, trading hours per symbol

---

### Phase 3: Account Analytics (Priority: MEDIUM)
Medium complexity, builds on account-info core.

| Feature | Complexity | Dependencies | Est. Tasks |
|---------|-----------|--------------|-----------|
| **4. margin-calculator** | Low | Decimal math | 4 |
| **5. position-details** | Low | Enhance account-info | 6 |
| **6. account-equity-history** | Medium | SQLite query → internal storage | 8 |

**Rationale**: Risk management layer. Traders need margin-before-order-placement calculation. Equity history enables drawdown analysis.

**MVP Acceptance**:
- margin-calculator: Return margin % required for 0.1 lot EURUSD at current price
- position-details: Return P&L, swap, duration for open positions
- account-equity-history: Query equity by date range (daily snapshots)

---

### Phase 4: Advanced Analytics (Priority: MEDIUM)
Medium-high complexity, optional for MVP but high value.

| Feature | Complexity | Dependencies | Est. Tasks |
|---------|-----------|--------------|-----------|
| **7. balance-drawdown** | Medium | Equity history + math | 6 |
| **8. order-fill-analysis** | Medium | Order execution data | 7 |
| **9. market-hours** | Low | Symbol properties + trading cal | 4 |
| **10. tick-data** | High | gRPC streaming, real-time MT5 | 12 |

**Rationale**: Performance tracking (drawdown, slippage). Market hours for automation rules. Tick data for high-frequency strategies.

**MVP Acceptance**:
- balance-drawdown: Return max % drawdown since account creation
- order-fill-analysis: Return slippage, fill time, latency per order
- market-hours: Return open/close times for symbol
- tick-data: Stream bid/ask ticks every 100ms via gRPC

---

## Architecture Pattern (All 10 Features)

All features follow same 3-layer stack:

```
JSON-RPC /rpc (HTTP)
    ↓
MCP Tool (internal/services/mcp/{feature}_tool.go)
    ↓
gRPC Handler (internal/services/daemon/{feature}_service.go)
    ↓
MT5 Service (internal/services/mt5/{feature}_service.go)
    ↓
MT5 HTTP API or SQLite local cache
```

### Shared Infrastructure (No new code):
- Decimal precision handling (shopspring/decimal)
- Portuguese error messages (internal/services/pt_br.go)
- Latency logging + audit trail (internal/logger.go)
- TLS, auth, health checks (existing)
- SQLite queue persistence (existing)

### New Infrastructure (Needed for Phase 3+):
- **Equity snapshot store**: SQLite table `account_equity_history(account_id, timestamp, equity, balance)` for Phase 3
- **Tick buffer**: In-memory ring buffer for last 1000 ticks per symbol (Phase 4)
- **Trade execution cache**: Order fill details (slippage, time) for Phase 4

---

## Implementation Strategy

### Immediate (This Sprint):
- Phase 2 complete: modify-order, pending-order-details, symbol-properties
- All 3 features in bin/mcp-mt5-server
- Test against mock MT5 or real MT5 if available

### Next Sprint:
- Phase 3: margin-calculator, position-details, account-equity-history
- Add SQLite migrations for equity history
- Integration test with 10-order concurrent load

### Future:
- Phase 4: Advanced analytics + tick streaming
- Tick buffer requires gRPC streaming (upgrade from unary calls)
- Order-fill-analysis requires capturing fill data in queue handler

---

## Success Criteria (All 10 Features)

| Criterion | Target |
|-----------|--------|
| **Tool invocation latency** | <5s for all tools (MT5 API timeout) |
| **Decimal precision** | All currency values as strings, no float rounding |
| **Error messages** | Portuguese (pt-BR) + technical detail |
| **Concurrent load** | 10 concurrent tool calls, FIFO queue |
| **Persistence** | SQLite survives daemon restart (Phase 3+) |
| **Test coverage** | Unit + integration per tool |

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| MT5 API endpoint changes | Contract tests catch breaking changes early |
| Equity history data explosion | Compress old snapshots to daily → weekly after 30 days |
| Tick data high bandwidth | Sample every 100ms, circular buffer (max 10MB) |
| Order fill capture complexity | Capture in daemon queue handler before MT5 response |

---

## Data Model Additions (Phases 2-4)

### Tables
```sql
-- Phase 3
CREATE TABLE account_equity_history (
  account_id INTEGER,
  timestamp INTEGER,    -- Unix epoch
  equity DECIMAL(20,2),
  balance DECIMAL(20,2),
  PRIMARY KEY (account_id, timestamp)
);

-- Phase 4 (optional)
CREATE TABLE order_fills (
  ticket INTEGER PRIMARY KEY,
  symbol TEXT,
  fill_time INTEGER,      -- Unix epoch
  fill_price DECIMAL(10,5),
  slippage DECIMAL(10,5),
  execution_latency_ms INTEGER
);
```

### In-Memory
```go
// Phase 4: tick buffer per symbol
type TickBuffer struct {
  symbol string
  ticks  []*Tick  // ring buffer, size 1000
  mutex  sync.RWMutex
}

type Tick struct {
  timestamp int64
  bid       decimal.Decimal
  ask       decimal.Decimal
}
```

---

## Contracts (API Signatures)

### Phase 2 Examples

**modify-order**
```json
{
  "method": "modify-order",
  "params": {
    "ticket": 12345,
    "price": 1.0950,
    "stop_loss": 1.0900,
    "take_profit": 1.1000
  }
}
```

**pending-order-details**
```json
{
  "method": "pending-order-details",
  "params": {
    "symbol": "EURUSD",
    "status": "pending",
    "created_after": 1715000000
  }
}
```

**symbol-properties**
```json
{
  "method": "symbol-properties",
  "params": {
    "symbol": "EURUSD"
  }
}
```

### Phase 3 Examples

**margin-calculator**
```json
{
  "method": "margin-calculator",
  "params": {
    "symbol": "EURUSD",
    "volume": 0.1,
    "price": 1.0950
  }
}
```

**account-equity-history**
```json
{
  "method": "account-equity-history",
  "params": {
    "from_timestamp": 1714000000,
    "to_timestamp": 1715000000,
    "granularity": "daily"
  }
}
```

### Phase 4 Examples

**balance-drawdown**
```json
{
  "method": "balance-drawdown",
  "params": {
    "since_timestamp": null
  }
}
```

**tick-data**
```json
{
  "method": "tick-data",
  "params": {
    "symbol": "EURUSD",
    "duration_seconds": 60
  }
}
```
(Returns stream of ticks, client subscribes via gRPC streaming)

---

## Tasks Decomposition (Per Feature)

### Phase 2 (modify-order) — 8 tasks
1. Update MT5 service with PATCH endpoint call
2. Create daemon handler with validation
3. Create MCP tool wrapper
4. Register in main.go tool registry
5. Write unit tests
6. Write integration tests
7. Build + verify compilation
8. Test with curl against mock/real MT5

### (Repeat pattern for pending-order-details, symbol-properties, etc.)

---

## Rollout Plan

**Week 1** (this sprint):
- Phase 2 all 3 features complete
- Binary rebuilt with 9 tools
- Integration test passes

**Week 2**:
- Phase 3 started
- Equity history SQLite migration complete
- margin-calculator + position-details ready

**Week 3**:
- Phase 3 complete
- 12 tools in production
- Load test: 10 concurrent orders + analytics queries

**Week 4+**:
- Phase 4 (advanced, optional)
- Tick streaming if high-frequency trading needed

---

## Definition of Done (Per Feature)

- [ ] Code compiles without warnings
- [ ] Unit tests pass (mock MT5)
- [ ] Integration test passes (real MT5 or mock)
- [ ] Tool callable via `/rpc` endpoint
- [ ] Error messages in Portuguese
- [ ] Latency <5s measured
- [ ] Commit message references feature + ticket
- [ ] Decimal precision verified (no float rounding)
- [ ] Concurrent load test (10 calls) passes
