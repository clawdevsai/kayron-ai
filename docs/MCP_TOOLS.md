# MCP Tools Documentation

## Overview

The MT5 MCP integration exposes 5 tools via JSON-RPC 2.0 for use with Claude and other AI agents.

---

## 1. account_info Tool

Retrieve current account information including balance, equity, and margin levels.

### JSON-RPC Request

```json
{
  "jsonrpc": "2.0",
  "id": "req-001",
  "method": "account_info",
  "params": {
    "account_id": "12345"
  }
}
```

### Success Response

```json
{
  "jsonrpc": "2.0",
  "id": "req-001",
  "result": {
    "account_id": "12345",
    "balance": 10000.50,
    "equity": 9850.25,
    "free_margin": 4925.12,
    "margin_used": 4925.13,
    "margin_level": 200.5,
    "currency": "USD",
    "timestamp": 1715164245000
  }
}
```

### Error Response (pt-BR)

```json
{
  "jsonrpc": "2.0",
  "id": "req-001",
  "error": {
    "code": -1,
    "message": "Terminal desconectado"
  }
}
```

---

## 2. get_quote Tool

Retrieve current market quote (bid/ask) for a trading symbol.

### JSON-RPC Request

```json
{
  "jsonrpc": "2.0",
  "id": "req-002",
  "method": "get_quote",
  "params": {
    "symbol": "EURUSD"
  }
}
```

### Success Response

```json
{
  "jsonrpc": "2.0",
  "id": "req-002",
  "result": {
    "symbol": "EURUSD",
    "bid": 1.08550,
    "ask": 1.08560,
    "timestamp": 1715164250000
  }
}
```

### Common Errors

```json
{
  "jsonrpc": "2.0",
  "id": "req-002",
  "error": {
    "code": -1,
    "message": "Símbolo não encontrado"
  }
}
```

pt-BR messages:
- "Símbolo não encontrado" - Symbol not found
- "Cotação indisponível" - Quote unavailable
- "Tempo limite excedido" - Request timeout

---

## 3. place_order Tool

Place a new trading order with optional stop-loss and take-profit levels.

### JSON-RPC Request

```json
{
  "jsonrpc": "2.0",
  "id": "req-003",
  "method": "place_order",
  "params": {
    "account_id": "12345",
    "symbol": "EURUSD",
    "order_type": "BUY",
    "volume": 1.5,
    "stop_loss": 1.08400,
    "take_profit": 1.08700,
    "idempotency_key": "order-unique-key-001"
  }
}
```

### Success Response

```json
{
  "jsonrpc": "2.0",
  "id": "req-003",
  "result": {
    "order_ticket": 12345678,
    "symbol": "EURUSD",
    "volume": 1.5,
    "entry_price": 1.08560,
    "status": "FILLED",
    "timestamp": 1715164255000
  }
}
```

### Error Response Examples

```json
{
  "jsonrpc": "2.0",
  "id": "req-003",
  "error": {
    "code": -1,
    "message": "Saldo insuficiente para a margem"
  }
}
```

pt-BR messages:
- "Saldo insuficiente para a margem" - Insufficient margin
- "Abertura de preço detectada" - Price gapping
- "Pedido rejeitado" - Order rejected

### Idempotency

The `idempotency_key` ensures exactly-once semantics. Duplicate requests with the same key return the original result.

---

## 4. close_position Tool

Close an open trading position.

### JSON-RPC Request

```json
{
  "jsonrpc": "2.0",
  "id": "req-004",
  "method": "close_position",
  "params": {
    "account_id": "12345",
    "position_ticket": 12345678
  }
}
```

### Success Response

```json
{
  "jsonrpc": "2.0",
  "id": "req-004",
  "result": {
    "position_ticket": 12345678,
    "close_price": 1.08600,
    "profit_loss": 60.00,
    "timestamp": 1715164260000
  }
}
```

### Error Response

```json
{
  "jsonrpc": "2.0",
  "id": "req-004",
  "error": {
    "code": -1,
    "message": "Posição não encontrada"
  }
}
```

pt-BR messages:
- "Posição não encontrada" - Position not found
- "Posição já foi fechada" - Position already closed
- "Não é possível fechar a posição" - Cannot close position

---

## 5. list_orders Tool

List all open orders and positions for an account.

### JSON-RPC Request

```json
{
  "jsonrpc": "2.0",
  "id": "req-005",
  "method": "list_orders",
  "params": {
    "account_id": "12345"
  }
}
```

### Success Response

```json
{
  "jsonrpc": "2.0",
  "id": "req-005",
  "result": {
    "orders": [
      {
        "ticket": 12345678,
        "symbol": "EURUSD",
        "type": "BUY",
        "volume": 1.5,
        "entry_price": 1.08560,
        "current_price": 1.08600,
        "profit_loss": 60.00
      },
      {
        "ticket": 12345679,
        "symbol": "GBPUSD",
        "type": "SELL",
        "volume": 1.0,
        "entry_price": 1.27300,
        "current_price": 1.27250,
        "profit_loss": 50.00
      }
    ]
  }
}
```

---

## Error Handling

### Common Error Codes & pt-BR Messages

| Message (pt-BR) | Meaning | Scenario |
|-----------------|---------|----------|
| Terminal desconectado | Terminal disconnected | Not connected to MT5 |
| Tempo limite excedido | Timeout exceeded | RPC took too long |
| Credenciais inválidas | Invalid credentials | Auth failed |
| Saldo insuficiente | Insufficient balance | Not enough funds |
| Símbolo não encontrado | Symbol not found | Invalid symbol |
| Cotação indisponível | Quote unavailable | No market data |
| Abertura de preço | Price gap detected | Market gapped |
| Pedido rejeitado | Order rejected | MT5 rejected order |
| Posição não encontrada | Position not found | Invalid ticket |
| Posição já foi fechada | Position closed | Already closed |

### Retry Logic

Transient errors (disconnect, timeout) should trigger retry with exponential backoff:

```
Attempt 1: Wait 100ms
Attempt 2: Wait 200ms
Attempt 3: Wait 400ms
Attempt 4: Wait 800ms
Max: 3 retries
```

---

## Performance Metrics

Each tool logs latency metrics to `/health` endpoint:

```json
{
  "metrics": {
    "account_info": {
      "count": 150,
      "avg_ms": 450,
      "p50_ms": 250,
      "p95_ms": 1200,
      "p99_ms": 1800,
      "min_ms": 50,
      "max_ms": 2100
    }
  }
}
```

---

## Example: Complete Trading Flow

```json
[
  {
    "method": "account_info",
    "params": {"account_id": "12345"}
  },
  {
    "method": "get_quote",
    "params": {"symbol": "EURUSD"}
  },
  {
    "method": "place_order",
    "params": {
      "account_id": "12345",
      "symbol": "EURUSD",
      "order_type": "BUY",
      "volume": 1.0,
      "idempotency_key": "trade-001"
    }
  },
  {
    "method": "list_orders",
    "params": {"account_id": "12345"}
  },
  {
    "method": "close_position",
    "params": {
      "account_id": "12345",
      "position_ticket": 12345678
    }
  }
]
```

All error messages returned in Portuguese (pt-BR) for Brazilian users.
