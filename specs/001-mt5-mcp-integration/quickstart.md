# Quickstart: MT5 MCP Integration

**Feature**: 001-mt5-mcp-integration
**Date**: 2026-05-08

---

## Prerequisites

- Go 1.21+
- MT5 terminal with WebAPI enabled (port 8228)
- MT5 trading account with credentials

---

## Setup

### 1. Environment Variables

```bash
export MT5_HOST=192.168.1.100:8228
export MT5_LOGIN=12345678
export MT5_PASSWORD=your_password
export MT5_SERVER=MT5Server
```

### 2. Build

```bash
go build -o mt5-mcp ./cmd/server
```

### 3. Run

```bash
./mt5-mcp serve --terminal /path/to/mt5terminal.exe
```

---

## Usage

### Start MCP Server

```bash
mt5-mcp serve
```

Server starts on stdio (MCP protocol). Connect via MCP client.

### Available Tools

| Tool | Description |
|------|-------------|
| `account-info` | Account balance, equity, margin |
| `quote` | Current bid/ask for symbol |
| `place-order` | Place buy/sell or pending order |
| `close-position` | Close open position by ticket |
| `orders-list` | List pending orders |

### Example: Account Info

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"account-info","arguments":{}}}' | mt5-mcp serve
```

### Example: Get Quote

```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"quote","arguments":{"symbol":"EURUSD"}}}' | mt5-mcp serve
```

### Example: Place Order

```bash
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"place-order","arguments":{"symbol":"EURUSD","type":"buy","volume":"0.1","stopLoss":"1.0800","takeProfit":"1.0900"}}}' | mt5-mcp serve
```

---

## Health Check

```bash
echo '{"jsonrpc":"2.0","id":0,"method":"tools/call","params":{"name":"mt5-health","arguments":{}}}' | mt5-mcp serve
```

Returns terminal connection status and last heartbeat.

---

## Error Handling

All errors return Portuguese (pt-BR) messages:

```json
{
  "error": {
    "code": "MT5_TERMINAL_DISCONNECTED",
    "message": "Conexão com terminal MT5 perdida. Verifique a conexão."
  }
}
```

---

## Performance

| Operation | Target |
|-----------|--------|
| account-info | < 2s |
| quote | < 500ms |
| place-order | < 5s |
| Concurrent invocations | 10 simultaneous |

---

## Development

### Run Tests

```bash
go test ./...
```

### Lint Proto

```bash
buf lint api/proto/mt5.proto
```

### Regenerate Proto

```bash
buf generate api/proto/mt5.proto
```
