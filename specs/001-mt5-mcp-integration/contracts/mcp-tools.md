# Contracts: MT5 MCP Integration

**Feature**: 001-mt5-mcp-integration
**Date**: 2026-05-08

---

## MCP Tool Schemas

### account-info

```json
{
  "name": "account-info",
  "description": "Retorna informações da conta MT5: saldo, equity, margem e margem livre.",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "accountNumber": { "type": "integer", "description": "Número da conta MT5" },
      "balance": { "type": "string", "description": "Saldo em decimal (ex: '10000.50')" },
      "equity": { "type": "string", "description": "Equity em decimal" },
      "margin": { "type": "string", "description": "Margem utilizada em decimal" },
      "freeMargin": { "type": "string", "description": "Margem livre em decimal" },
      "currency": { "type": "string", "description": "Moeda da conta (ex: 'USD')" },
      "server": { "type": "string", "description": "Servidor MT5" }
    }
  }
}
```

### quote

```json
{
  "name": "quote",
  "description": "Retorna cotação atual (bid/ask) para um instrumento.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "symbol": {
        "type": "string",
        "description": "Símbolo do instrumento (ex: 'EURUSD')",
        "pattern": "^[A-Z]{3}[A-Z]{3}$"
      }
    },
    "required": ["symbol"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "symbol": { "type": "string" },
      "bid": { "type": "string", "description": "Preço bid em decimal" },
      "ask": { "type": "string", "description": "Preço ask em decimal" },
      "timestamp": { "type": "string", "format": "date-time", "description": "Timestamp da cotação" }
    }
  }
}
```

### place-order

```json
{
  "name": "place-order",
  "description": "Coloca ordem de compra ou venda no MT5.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "symbol": { "type": "string", "description": "Símbolo (ex: 'EURUSD')" },
      "type": { "type": "string", "enum": ["buy", "sell", "buy_limit", "sell_limit", "buy_stop", "sell_stop"] },
      "volume": { "type": "string", "description": "Volume em lotes (ex: '0.1')" },
      "price": { "type": "string", "description": "Preço para ordens pendentes (0 para market)" },
      "stopLoss": { "type": "string", "description": "Preço stop loss (0 para nenhum)" },
      "takeProfit": { "type": "string", "description": "Preço take profit (0 para nenhum)" },
      "comment": { "type": "string", "maxLength": 100 }
    },
    "required": ["symbol", "type", "volume"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "ticket": { "type": "integer", "description": "Número do ticket MT5" },
      "status": { "type": "string" },
      "fillPrice": { "type": "string", "description": "Preço de execução (se preenchida)" }
    }
  }
}
```

### close-position

```json
{
  "name": "close-position",
  "description": "Fecha uma posição aberta no MT5.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "ticket": { "type": "integer", "description": "Número do ticket da posição" }
    },
    "required": ["ticket"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "ticket": { "type": "integer" },
      "profit": { "type": "string", "description": "Lucro/prejuízo realizado em decimal" },
      "status": { "type": "string" }
    }
  }
}
```

### orders-list

```json
{
  "name": "orders-list",
  "description": "Lista todas as ordens pendentes na conta MT5.",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "orders": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "ticket": { "type": "integer" },
            "symbol": { "type": "string" },
            "type": { "type": "string" },
            "volume": { "type": "string" },
            "price": { "type": "string" }
          }
        }
      }
    }
  }
}
```

---

## Error Response Contract

All errors follow this structure:

```json
{
  "error": {
    "code": "MT5_<ERROR_CODE>",
    "message": "Mensagem em português (pt-BR)",
    "data": {}
  }
}
```

**Error codes**: MT5_TERMINAL_DISCONNECTED | MT5_AUTH_FAILED | MT5_ORDER_REJECTED | MT5_INSUFFICIENT_MARGIN | MT5_SYMBOL_NOT_FOUND | MT5_POSITION_NOT_FOUND | MT5_TIMEOUT | INPUT_VALIDATION_ERROR

---

## gRPC Service Contract

```protobuf
syntax = "proto3";

package mt5.v1;

service MT5Terminal {
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
  rpc GetAccountInfo(GetAccountInfoRequest) returns (GetAccountInfoResponse);
  rpc GetQuote(GetQuoteRequest) returns (GetQuoteResponse);
  rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
  rpc ClosePosition(ClosePositionRequest) returns (ClosePositionResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message HealthCheckRequest {}
message HealthCheckResponse { bool connected = 1; string last_heartbeat = 2; }

message GetAccountInfoRequest {}
message GetAccountInfoResponse {
  int64 account_number = 1;
  string balance = 2;
  string equity = 3;
  string margin = 4;
  string free_margin = 5;
  string currency = 6;
  string server = 7;
}

message GetQuoteRequest { string symbol = 1; }
message GetQuoteResponse {
  string symbol = 1;
  string bid = 2;
  string ask = 3;
  string timestamp = 4;
}

message PlaceOrderRequest {
  string symbol = 1;
  OrderType type = 2;
  string volume = 3;
  string price = 4;
  string stop_loss = 5;
  string take_profit = 6;
  string comment = 7;
}
message PlaceOrderResponse { int64 ticket = 1; OrderStatus status = 2; string fill_price = 3; }

message ClosePositionRequest { int64 ticket = 1; }
message ClosePositionResponse { int64 ticket = 1; string profit = 2; OrderStatus status = 3; }

message ListOrdersRequest {}
message ListOrdersResponse { repeated Order orders = 1; }

enum OrderType { BUY = 0; SELL = 1; BUY_LIMIT = 2; SELL_LIMIT = 3; BUY_STOP = 4; SELL_STOP = 5; }
enum OrderStatus { SUBMITTED = 0; FILLED = 1; CANCELLED = 2; REJECTED = 3; }

message Order {
  int64 ticket = 1;
  string symbol = 2;
  OrderType type = 3;
  string volume = 4;
  string price = 5;
}
```
