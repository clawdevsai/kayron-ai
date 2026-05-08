# gRPC Services Documentation

## Overview

The MT5 MCP integration exposes 5 core trading operations via gRPC with Portuguese (pt-BR) error messages and comprehensive error handling.

## Service Definition

```protobuf
service MT5TradingService {
  rpc AccountInfo(AccountRequest) returns (AccountResponse);
  rpc GetQuote(QuoteRequest) returns (QuoteResponse);
  rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
  rpc ClosePosition(ClosePositionRequest) returns (ClosePositionResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

## 1. AccountInfo RPC

**Purpose**: Retrieve current account information

### Request

```protobuf
message AccountRequest {
  string account_id = 1;
}
```

### Response

```protobuf
message AccountResponse {
  string account_id = 1;
  double balance = 2;
  double equity = 3;
  double free_margin = 4;
  double margin_used = 5;
  double margin_level = 6;
  string currency = 7;
  int64 timestamp = 8;
}
```

### Error Codes

| gRPC Code | pt-BR Message | Scenario |
|-----------|---------------|----------|
| `UNAVAILABLE` | "Terminal desconectado" | Terminal not connected |
| `DEADLINE_EXCEEDED` | "Tempo limite excedido" | RPC timeout |
| `UNAUTHENTICATED` | "Credenciais inválidas" | Invalid credentials |
| `INTERNAL` | "Erro interno do terminal" | Terminal error |

### Example

```bash
grpcurl -plaintext \
  -d '{"account_id": "12345"}' \
  localhost:50051 \
  mt5.MT5TradingService/AccountInfo
```

---

## 2. GetQuote RPC

**Purpose**: Retrieve current market quote for a symbol

### Request

```protobuf
message QuoteRequest {
  string symbol = 1;
}
```

### Response

```protobuf
message QuoteResponse {
  string symbol = 1;
  double bid = 2;
  double ask = 3;
  int64 timestamp = 4;
}
```

### Error Codes

| gRPC Code | pt-BR Message | Scenario |
|-----------|---------------|----------|
| `NOT_FOUND` | "Símbolo não encontrado" | Invalid symbol |
| `UNAVAILABLE` | "Cotação indisponível" | No quote data |
| `DEADLINE_EXCEEDED` | "Tempo limite excedido" | RPC timeout |

### Example

```bash
grpcurl -plaintext \
  -d '{"symbol": "EURUSD"}' \
  localhost:50051 \
  mt5.MT5TradingService/GetQuote
```

---

## 3. PlaceOrder RPC

**Purpose**: Place a new trading order

### Request

```protobuf
message PlaceOrderRequest {
  string account_id = 1;
  string symbol = 2;
  string order_type = 3; // "BUY" or "SELL"
  double volume = 4;
  double stop_loss = 5;
  double take_profit = 6;
  string idempotency_key = 7;
}
```

### Response

```protobuf
message PlaceOrderResponse {
  int64 order_ticket = 1;
  string symbol = 2;
  double volume = 3;
  double entry_price = 4;
  string status = 5;
  int64 timestamp = 6;
}
```

### Error Codes

| gRPC Code | pt-BR Message | Scenario |
|-----------|---------------|----------|
| `INVALID_ARGUMENT` | "Saldo insuficiente para a margem" | Insufficient margin |
| `FAILED_PRECONDITION` | "Abertura de preço detectada" | Price gapping |
| `INTERNAL` | "Pedido rejeitado" | Order rejected by MT5 |

### Idempotency

Requests with `idempotency_key` are guaranteed to be processed exactly-once. Duplicate requests return the original result.

### Example

```bash
grpcurl -plaintext \
  -d '{
    "account_id": "12345",
    "symbol": "EURUSD",
    "order_type": "BUY",
    "volume": 1.0,
    "idempotency_key": "order-001"
  }' \
  localhost:50051 \
  mt5.MT5TradingService/PlaceOrder
```

---

## 4. ClosePosition RPC

**Purpose**: Close an open position

### Request

```protobuf
message ClosePositionRequest {
  string account_id = 1;
  int64 position_ticket = 2;
}
```

### Response

```protobuf
message ClosePositionResponse {
  int64 position_ticket = 1;
  double close_price = 2;
  double profit_loss = 3;
  int64 timestamp = 4;
}
```

### Error Codes

| gRPC Code | pt-BR Message | Scenario |
|-----------|---------------|----------|
| `NOT_FOUND` | "Posição não encontrada" | Position doesn't exist |
| `FAILED_PRECONDITION` | "Posição já foi fechada" | Position already closed |
| `INTERNAL` | "Não é possível fechar a posição" | Close operation failed |

### Example

```bash
grpcurl -plaintext \
  -d '{
    "account_id": "12345",
    "position_ticket": 67890
  }' \
  localhost:50051 \
  mt5.MT5TradingService/ClosePosition
```

---

## 5. ListOrders RPC

**Purpose**: List all open orders and positions

### Request

```protobuf
message ListOrdersRequest {
  string account_id = 1;
}
```

### Response

```protobuf
message ListOrdersResponse {
  repeated Order orders = 1;
}

message Order {
  int64 ticket = 1;
  string symbol = 2;
  string type = 3;
  double volume = 4;
  double entry_price = 5;
  double current_price = 6;
  double profit_loss = 7;
}
```

### Example

```bash
grpcurl -plaintext \
  -d '{"account_id": "12345"}' \
  localhost:50051 \
  mt5.MT5TradingService/ListOrders
```

---

## Error Handling

### gRPC Status Codes Mapping

```
MT5 Error Type          → gRPC Status Code
─────────────────────────────────────────
Disconnect             → UNAVAILABLE
Timeout                → DEADLINE_EXCEEDED
Invalid Credentials    → UNAUTHENTICATED
Insufficient Margin    → INVALID_ARGUMENT
Symbol Not Found       → NOT_FOUND
Price Gapping          → FAILED_PRECONDITION
Quote Unavailable      → UNAVAILABLE
Order Rejected         → FAILED_PRECONDITION
```

### Error Response Format

All errors include Portuguese (pt-BR) error messages:

```
{
  "code": 14,  // UNAVAILABLE
  "message": "Terminal desconectado"
}
```

---

## Performance SLAs

| Operation | Target Latency | P95 | P99 |
|-----------|----------------|-----|-----|
| AccountInfo | < 2s | < 1.5s | < 2.5s |
| GetQuote | < 500ms | < 400ms | < 600ms |
| PlaceOrder | < 5s | < 4s | < 6s |
| ClosePosition | < 5s | < 4s | < 6s |
| ListOrders | < 2s | < 1.5s | < 2.5s |

Latency metrics are exposed via `/health` endpoint.

---

## Health Check Endpoint

```
GET /health
```

Returns:

```json
{
  "status": "ok",
  "terminal_connected": true,
  "last_heartbeat": "2026-05-08T10:30:45Z",
  "metrics": {
    "account_info": {"p50_ms": 150, "p95_ms": 1200, "p99_ms": 1800},
    "get_quote": {"p50_ms": 200, "p95_ms": 400, "p99_ms": 550},
    "place_order": {"p50_ms": 2000, "p95_ms": 3500, "p99_ms": 4800}
  }
}
```
