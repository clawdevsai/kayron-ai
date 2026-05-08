# Data Model: MT5 MCP Integration

**Feature**: 001-mt5-mcp-integration
**Date**: 2026-05-08

---

## Core Entities

### MT5Terminal

Represents the MetaTrader 5 terminal instance.

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| `path` | string | non-empty, valid Windows path | Path to MT5 terminal .exe |
| `host` | string | valid hostname or IP | MT5 WebAPI host (default localhost:8228) |
| `connected` | bool | — | Runtime connection state |
| `lastHeartbeat` | timestamp | — | Last successful health check |

### TradingAccount

MT5 trading account associated with the terminal.

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| `accountNumber` | int64 | > 0 | MT5 account number |
| `balance` | decimal.Decimal | ≥ 0 | Account balance in account currency |
| `equity` | decimal.Decimal | ≥ 0 | Current equity |
| `margin` | decimal.Decimal | ≥ 0 | Used margin |
| `freeMargin` | decimal.Decimal | ≥ 0 | Available margin |
| `currency` | string | non-empty | Account currency (e.g., "USD") |
| `server` | string | non-empty | MT5 server name |

### Instrument

Tradeable symbol on MT5.

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| `symbol` | string | non-empty, uppercase | Trading symbol (e.g., "EURUSD") |
| `description` | string | — | Human-readable name |
| `digits` | int | 2–6 | Price decimal places |
| `tickValue` | decimal.Decimal | > 0 | Value of one tick |

### Quote

Real-time bid/ask price.

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| `symbol` | string | valid instrument | Symbol this quote is for |
| `bid` | decimal.Decimal | > 0 | Bid price |
| `ask` | decimal.Decimal | > bid | Ask price |
| `timestamp` | timestamp | ≤ now + 1s | Quote timestamp |

### Order

Trading order (market or pending).

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| `ticket` | int64 | > 0 | MT5 order ticket (unique per terminal) |
| `symbol` | string | valid instrument | Order symbol |
| `type` | enum | buy\|sell\|buy_limit\|sell_limit\|buy_stop\|sell_stop | Order type |
| `volume` | decimal.Decimal | > 0, ≤ 100 | Lot size |
| `price` | decimal.Decimal | > 0 | Limit/stop price (0 for market) |
| `stopLoss` | decimal.Decimal | ≥ 0 | Stop loss price (0 if none) |
| `takeProfit` | decimal.Decimal | ≥ 0 | Take profit price (0 if none) |
| `status` | enum | submitted\|filled\|cancelled\|rejected | Order status |
| `fillPrice` | decimal.Decimal | > 0, present if filled | Actual fill price |
| `comment` | string | max 100 chars | Order comment |

### Position

Open trading position.

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| `ticket` | int64 | > 0 | MT5 position ticket |
| `symbol` | string | valid instrument | Position symbol |
| `type` | enum | buy\|sell | Position type |
| `volume` | decimal.Decimal | > 0 | Position size |
| `openPrice` | decimal.Decimal | > 0 | Entry price |
| `currentPrice` | decimal.Decimal | > 0 | Current market price |
| `profit` | decimal.Decimal | — | Floating P/L in account currency |
| `stopLoss` | decimal.Decimal | ≥ 0 | SL price (0 if none) |
| `takeProfit` | decimal.Decimal | ≥ 0 | TP price (0 if none) |

---

## Relationships

```
MT5Terminal 1────1 TradingAccount  (one account per terminal)
MT5Terminal 1────* Position         (multiple open positions)
MT5Terminal 1────* Order           (multiple pending orders)
MT5Terminal 1────* Instrument      (instruments available for trading)
TradingAccount *────* Position      (account's open positions)
Instrument    1────* Quote          (real-time quotes per instrument)
```

---

## State Transitions

### Order States

```
submitted ──► filled      (when market order executes)
submitted ──► rejected   (when MT5 rejects order)
submitted ──► cancelled  (when client cancels pending order)
```

### Terminal States

```
disconnected ──► connecting ──► connected ──► reconnecting
                                                    │
                                          (fails) ───┴──► disconnected
```

---

## Error Model

All MCP tool errors follow consistent structure:

```json
{
  "error": {
    "code": "MT5_TERMINAL_DISCONNECTED",
    "message": "Conexão com terminal MT5 perdida. Verifique a conexão.",
    "data": {
      "terminal": "mt5-terminal-1",
      "lastHeartbeat": "2026-05-08T10:30:00Z"
    }
  }
}
```

### Error Codes

| Code | HTTP-equivalent | Description |
|------|----------------|--------------|
| `MT5_TERMINAL_DISCONNECTED` | 503 | Terminal unreachable |
| `MT5_AUTH_FAILED` | 401 | Login credentials invalid |
| `MT5_ORDER_REJECTED` | 422 | Order rejected by MT5 (reason in message) |
| `MT5_INSUFFICIENT_MARGIN` | 422 | Not enough margin for order |
| `MT5_SYMBOL_NOT_FOUND` | 404 | Instrument not in market watch |
| `MT5_POSITION_NOT_FOUND` | 404 | Position ticket not found |
| `MT5_TIMEOUT` | 504 | MT5 WebAPI timeout (>5s) |
| `INPUT_VALIDATION_ERROR` | 400 | Invalid tool input parameters |

---

## Validation Rules

| Entity | Rule |
|--------|------|
| Order.volume | Must be multiple of MT5 minimum lot (e.g., 0.01 for most symbols) |
| Order.price | Must be within symbol-specific range (no gapping beyond allowed distance) |
| Quote.bid | Must be < Quote.ask (spread must be positive) |
| Terminal.host | Must be reachable on port 8228 |
| All decimal fields | Use shops/decimal — never float64 for currency values |
