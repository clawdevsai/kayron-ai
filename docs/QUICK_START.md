# Quick Start Guide

## Installation

### Prerequisites
- Go 1.21+
- Protocol Buffers compiler (protoc)
- MT5 Terminal with WebAPI enabled
- Git

### Build

```bash
# Clone repository
git clone https://github.com/lukeware/kayron-ai.git
cd kayron-ai

# Install dependencies
go mod download

# Build daemon
go build -o bin/mcp-mt5-server ./cmd/mcp-mt5-server

# Build tests
go test ./... -v
```

### Configuration

```bash
# Set environment variables
export MT5_SERVER=localhost:7788
export MT5_LOGIN=12345
export MT5_PASSWORD=your_password
export MT5_SERVER_NAME=DemoServer
export LOG_FILE=logs/app.json
export DEBUG=false

# For production TLS
export TLS_CERT_FILE=/path/to/cert.pem
export TLS_KEY_FILE=/path/to/key.pem
```

### Run

```bash
# Start gRPC daemon
./bin/mcp-mt5-server

# With debug logging
./bin/mcp-mt5-server -debug

# Check health
curl http://localhost:50051/health | jq
```

---

## Using the 5 Tools

### 1. Account Info

Get current account balance, equity, and margin.

```bash
grpcurl -plaintext \
  -d '{"account_id": "12345"}' \
  localhost:50051 \
  mt5.MT5TradingService/AccountInfo
```

Response:
```json
{
  "account_id": "12345",
  "balance": 10000.50,
  "equity": 9850.25,
  "free_margin": 4925.12,
  "margin_level": 200.5,
  "currency": "USD"
}
```

### 2. Get Quote

Get bid/ask for a symbol.

```bash
grpcurl -plaintext \
  -d '{"symbol": "EURUSD"}' \
  localhost:50051 \
  mt5.MT5TradingService/GetQuote
```

Response:
```json
{
  "symbol": "EURUSD",
  "bid": 1.08550,
  "ask": 1.08560
}
```

### 3. Place Order

Submit a buy/sell order.

```bash
grpcurl -plaintext \
  -d '{
    "account_id": "12345",
    "symbol": "EURUSD",
    "order_type": "BUY",
    "volume": 1.0,
    "stop_loss": 1.08400,
    "take_profit": 1.08700,
    "idempotency_key": "order-001"
  }' \
  localhost:50051 \
  mt5.MT5TradingService/PlaceOrder
```

Response:
```json
{
  "order_ticket": 12345678,
  "symbol": "EURUSD",
  "volume": 1.0,
  "entry_price": 1.08560,
  "status": "FILLED"
}
```

### 4. List Orders

List all open orders.

```bash
grpcurl -plaintext \
  -d '{"account_id": "12345"}' \
  localhost:50051 \
  mt5.MT5TradingService/ListOrders
```

Response:
```json
{
  "orders": [
    {
      "ticket": 12345678,
      "symbol": "EURUSD",
      "type": "BUY",
      "volume": 1.0,
      "entry_price": 1.08560,
      "current_price": 1.08600,
      "profit_loss": 60.00
    }
  ]
}
```

### 5. Close Position

Close an open position.

```bash
grpcurl -plaintext \
  -d '{
    "account_id": "12345",
    "position_ticket": 12345678
  }' \
  localhost:50051 \
  mt5.MT5TradingService/ClosePosition
```

Response:
```json
{
  "position_ticket": 12345678,
  "close_price": 1.08600,
  "profit_loss": 60.00
}
```

---

## Testing

### Run All Tests

```bash
go test ./tests/... -v
```

### Load Test (10 concurrent)

```bash
go test -run TestLoadConcurrentToolInvocations ./tests/load/ -v
```

### Performance SLA Verification

```bash
go test -run TestPerformanceSLAVerification ./tests/performance/ -v
```

### Reconnect Test

```bash
go test -run TestAutoReconnectDetection ./tests/integration/ -v
```

---

## Monitoring

### Health Check

```bash
curl http://localhost:50051/health | jq
```

Shows:
- Terminal connection status
- Last heartbeat timestamp
- Latency metrics (p50, p95, p99)

### View Logs

```bash
# Follow real-time logs
tail -f logs/app.json | jq

# Filter by tool
tail -f logs/app.json | jq 'select(.tool_name == "place_order")'

# Filter errors only
tail -f logs/app.json | jq 'select(.error != null)'

# Audit logs
tail -f logs/audit.json | jq
```

### Latency Metrics

```bash
# Get latency stats from health endpoint
curl http://localhost:50051/health | jq '.metrics'
```

---

## Common Errors

### "Terminal desconectado"
Terminal is not connected. Check:
1. MT5 terminal is running
2. WebAPI is enabled
3. `MT5_SERVER` env var is correct

```bash
curl http://localhost:7788/health  # Should return 200
```

### "Saldo insuficiente"
Insufficient margin. Solution:
1. Deposit more funds
2. Reduce order volume
3. Close losing positions

### "Símbolo não encontrado"
Invalid symbol. Check:
1. Symbol name spelling (EURUSD, not EUR/USD)
2. Symbol is available on broker account
3. Symbol is enabled in MT5 Market Watch

### "Tempo limite excedido"
Request timeout. Check:
1. Terminal is responsive
2. System CPU/memory usage
3. Reduce concurrent requests

---

## Performance Targets

| Operation | Target | p95 | p99 |
|-----------|--------|-----|-----|
| AccountInfo | <2s | <1.5s | <2.5s |
| GetQuote | <500ms | <400ms | <600ms |
| PlaceOrder | <5s | <4s | <6s |
| ClosePosition | <5s | <4s | <6s |
| ListOrders | <2s | <1.5s | <2.5s |

---

## Security

### Enable TLS (Production)

```bash
export TLS_CERT_FILE=/path/to/cert.pem
export TLS_KEY_FILE=/path/to/key.pem
export ENV=production
./bin/mcp-mt5-server
```

### Credential Security

Never:
- Hardcode credentials in code
- Pass credentials in URLs
- Log credentials
- Commit .env files

Always:
- Use environment variables
- Use AWS Secrets Manager (when available)
- Audit credential access
- Rotate credentials regularly

---

## Troubleshooting

### Check Daemon Status

```bash
ps aux | grep mcp-mt5-server
```

### Restart Daemon

```bash
pkill mcp-mt5-server
sleep 2
./bin/mcp-mt5-server &
```

### Clear Stuck Queues

```bash
# View queue
sqlite3 queue.db "SELECT * FROM orders WHERE status='PENDING';"

# Clear old pending orders
sqlite3 queue.db "DELETE FROM orders WHERE created_at < datetime('now', '-1 day');"
```

### Enable Debug Mode

```bash
./bin/mcp-mt5-server -debug
# Shows all gRPC requests/responses and MT5 API calls
```

---

## Documentation

- **gRPC Services**: `docs/GRPC_SERVICES.md` - Full API documentation
- **MCP Tools**: `docs/MCP_TOOLS.md` - JSON-RPC 2.0 tool guide
- **Troubleshooting**: `docs/TROUBLESHOOTING.md` - Common errors & solutions
- **Architecture**: `docs/ARCHITECTURE.md` - System design
- **Release Notes**: `CHANGELOG.md` - v1.0.0 features and improvements

---

## Support

For issues:
1. Check logs: `tail -f logs/app.json`
2. Check health: `curl http://localhost:50051/health`
3. Review troubleshooting guide: `docs/TROUBLESHOOTING.md`
4. Collect debug logs: `./bin/mcp-mt5-server -debug`

---

## Next Steps

1. ✅ Build and start daemon
2. ✅ Verify health endpoint
3. ✅ Test account_info tool
4. ✅ Test place_order tool
5. ✅ Run integration tests
6. ✅ Deploy to production

Congratulations! Your MT5 MCP integration is ready.
