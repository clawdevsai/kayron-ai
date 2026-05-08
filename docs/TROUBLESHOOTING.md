# Troubleshooting Guide

## Common Errors & Solutions

### 1. "Terminal desconectado" (Terminal Disconnected)

**Error Message**: Terminal desconectado

**Symptoms**:
- All RPC calls fail with "Terminal desconectado"
- Health check shows `terminal_connected: false`
- gRPC status: `UNAVAILABLE`

**Debug Steps**:

1. Check MT5 terminal is running:
```bash
curl http://localhost:7788/health
# Should return 200 OK if MT5 WebAPI is running
```

2. Verify gRPC daemon is running:
```bash
ps aux | grep mcp-mt5-server
```

3. Check logs for connection errors:
```bash
tail -f logs/app.json | grep "terminal_connection"
```

4. Verify environment variables:
```bash
echo $MT5_SERVER  # Should be "localhost:7788" or IP:port
echo $MT5_LOGIN   # Should be populated
```

5. Restart connection:
```bash
# Kill daemon
pkill mcp-mt5-server
sleep 2

# Restart
./bin/mcp-mt5-server
```

**Solutions**:
- [ ] Start MT5 terminal and ensure WebAPI is enabled
- [ ] Verify `MT5_SERVER` env var points to correct MT5 WebAPI
- [ ] Check network connectivity between daemon and MT5
- [ ] Review MT5 terminal logs for connection errors
- [ ] Restart both MT5 and gRPC daemon

---

### 2. "Saldo insuficiente" (Insufficient Balance/Margin)

**Error Message**: Saldo insuficiente para a margem

**Symptoms**:
- PlaceOrder fails with margin error
- Account equity is too low for requested volume
- gRPC status: `INVALID_ARGUMENT`

**Debug Steps**:

1. Check account equity:
```bash
grpcurl -plaintext \
  -d '{"account_id": "12345"}' \
  localhost:50051 \
  mt5.MT5TradingService/AccountInfo
```

2. Calculate required margin:
```
Required Margin = Volume × (Bid Price × Lot Size) / Leverage
Example: 1.0 lot × 1.08550 × 100,000 / 100 = $1,085.50
```

3. Check margin level:
```bash
# From AccountInfo response
margin_level = (equity / margin_used) × 100
# Should be > 100% to open new positions
```

**Solutions**:
- [ ] Deposit more funds to the account
- [ ] Reduce order volume/lot size
- [ ] Close existing losing positions to free margin
- [ ] Check leverage setting (may need to increase)
- [ ] Wait for profitable positions to improve margin

---

### 3. "Símbolo não encontrado" (Symbol Not Found)

**Error Message**: Símbolo não encontrado

**Symptoms**:
- GetQuote/PlaceOrder fails for valid symbols
- Symbol name typo or incorrect format
- gRPC status: `NOT_FOUND`

**Debug Steps**:

1. Verify symbol format:
```bash
# Correct formats:
EURUSD, GBPUSD, USDJPY, XAUUSD
# Incorrect:
EUR/USD, EUR-USD, eurusd (case matters on some brokers)
```

2. Check available symbols:
```bash
# MT5 terminal: View → Market Watch
# or make quote request for known symbol first
grpcurl -plaintext \
  -d '{"symbol": "EURUSD"}' \
  localhost:50051 \
  mt5.MT5TradingService/GetQuote
```

3. Check symbol enabled on account:
```bash
# In MT5 terminal: right-click symbol → Properties
# Verify "Trade" is enabled
```

**Solutions**:
- [ ] Verify symbol name (case-sensitive, no spaces)
- [ ] Check symbol is available on broker's account
- [ ] Ensure symbol is not disabled/restricted
- [ ] Use Market Watch list to copy exact symbol name
- [ ] Verify correct market hours (some symbols inactive after hours)

---

### 4. "Tempo limite excedido" (Timeout/Deadline Exceeded)

**Error Message**: Tempo limite excedido

**Symptoms**:
- RPC calls take > 10 seconds and fail
- High latency (p95 > 5000ms)
- gRPC status: `DEADLINE_EXCEEDED`

**Debug Steps**:

1. Check gRPC daemon health:
```bash
curl http://localhost:50051/grpc.health.v1.Health/Check
```

2. Check MT5 terminal responsiveness:
```bash
# Try simple AccountInfo call - should be <2s
time grpcurl -plaintext \
  -d '{"account_id": "12345"}' \
  localhost:50051 \
  mt5.MT5TradingService/AccountInfo
```

3. Check system resources:
```bash
# Check CPU/Memory usage
top -n 1 | grep mcp-mt5-server
free -h  # Check available memory
```

4. Review metrics:
```bash
curl http://localhost:50051/health | jq '.metrics'
# Check p95_ms and p99_ms values
```

5. Check logs for slow operations:
```bash
tail -f logs/app.json | jq 'select(.latency_ms > 2000)'
```

**Solutions**:
- [ ] Restart gRPC daemon
- [ ] Increase deadline timeout (default 10s)
- [ ] Check MT5 terminal CPU/Memory usage
- [ ] Reduce concurrent requests (< 10 concurrent)
- [ ] Move daemon to faster machine
- [ ] Optimize PlaceOrder queue processing

---

### 5. "Abertura de preço" (Price Gap/Slippage)

**Error Message**: Abertura de preço detectada

**Symptoms**:
- PlaceOrder fails even with sufficient margin
- Price changed between quote and order
- gRPC status: `FAILED_PRECONDITION`

**Debug Steps**:

1. Check quote freshness:
```bash
grpcurl -plaintext \
  -d '{"symbol": "EURUSD"}' \
  localhost:50051 \
  mt5.MT5TradingService/GetQuote
# Note timestamp and compare to current time
```

2. Check market hours:
```bash
# Verify symbol is trading (not pre/post-market)
# Check timezone - ensure market is open
```

3. Review order parameters:
```bash
# In PlaceOrder request:
# - Volume should be within broker limits (e.g., 0.01 - 100)
# - SL/TP should be reasonable distance from entry
```

**Solutions**:
- [ ] Add slippage tolerance to quotes (refresh before order)
- [ ] Use limit orders instead of market orders
- [ ] Verify market is in normal trading hours
- [ ] Check broker's max spread tolerance
- [ ] Retry order with fresh quote

---

### 6. "Cotação indisponível" (Quote Unavailable)

**Error Message**: Cotação indisponível

**Symptoms**:
- GetQuote returns no data
- No bid/ask available for symbol
- gRPC status: `UNAVAILABLE`

**Debug Steps**:

1. Check market status:
```bash
# Verify trading hours for symbol's exchange
# Check if market is closed (weekends, holidays)
```

2. Verify symbol liquidity:
```bash
# Check in MT5: View → Market Watch
# Symbols without quotes are illiquid or closed
```

3. Check MT5 data connection:
```bash
# MT5 should show "data connection" status
# Check Tools → Options → Data (verify data is feeding)
```

**Solutions**:
- [ ] Ensure market is open (check trading hours)
- [ ] Verify symbol has sufficient liquidity
- [ ] Check MT5 data connection is active
- [ ] Retry during normal trading hours
- [ ] Switch to liquid symbols (EURUSD, GBPUSD)

---

## Health Check Monitoring

### Check Terminal Connection Status

```bash
curl http://localhost:50051/health | jq '.'
```

Response should show:

```json
{
  "terminal_connected": true,
  "last_heartbeat": "2026-05-08T10:30:45Z",
  "seconds_since_heartbeat": 2,
  "is_healthy": true
}
```

**If `is_healthy: false`**:
- Terminal is not responding
- Last heartbeat was > 10 seconds ago
- Daemon will attempt auto-reconnect

### Check Queue Status

```bash
# View queued orders
tail -f logs/app.json | jq 'select(.event_type == "queue")'
```

Queued orders will be reprocessed automatically after reconnect.

---

## Debug Logging

### Enable Debug Mode

```bash
./mcp-mt5-server -debug
```

Outputs:
- All gRPC requests/responses
- All MT5 API calls
- Detailed error information
- Queue processing steps

### View Audit Logs

```bash
tail -f logs/audit.json | jq '.'
```

Shows:
- Login attempts
- Credential rotations
- Connection state changes
- Access attempts

---

## Performance Diagnosis

### Latency Analysis

```bash
# Get latency metrics
curl http://localhost:50051/health | jq '.metrics'

# Real-time latency monitoring
tail -f logs/app.json | jq '{tool: .tool_name, latency_ms: .latency_ms}'
```

**SLA Targets**:
- `account_info`: < 2s (p95: < 1.5s)
- `get_quote`: < 500ms (p95: < 400ms)
- `place_order`: < 5s (p95: < 4s)

If exceeding SLA:
1. Check daemon CPU/memory
2. Check concurrent request count
3. Check MT5 terminal performance
4. Consider load balancing

---

## Recovery Procedures

### Restart gRPC Daemon

```bash
pkill mcp-mt5-server
sleep 2
./bin/mcp-mt5-server &
```

### Clear Stuck Queues

```bash
# View queue database
sqlite3 queue.db ".schema"

# Clear stuck orders (use caution!)
sqlite3 queue.db "DELETE FROM orders WHERE status='PENDING' AND created_at < datetime('now', '-1 day');"
```

### Reset Idempotency Cache

```bash
rm -f idempotency_cache.db
# Daemon will recreate on restart
```

### Full Reset (Data Loss - Caution!)

```bash
rm -f queue.db idempotency_cache.db
pkill mcp-mt5-server
sleep 2
./bin/mcp-mt5-server &
```

---

## Getting Help

If issue persists after troubleshooting:

1. **Collect logs** (last 1 hour):
```bash
tail -1000 logs/app.json > debug_logs.json
tail -1000 logs/audit.json >> debug_logs.json
```

2. **Provide**:
   - Debug logs
   - MT5 terminal version
   - Broker name
   - Account type (demo/live)
   - Exact error message

3. **Run diagnostics**:
```bash
./bin/mcp-mt5-server -health-check
```

