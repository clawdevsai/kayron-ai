# Feature Specification: MT5 MCP Integration

**Feature Branch**: `001-mt5-mcp-integration`
**Created**: 2026-05-08
**Status**: Draft
**Input**: User description: "MCP tools for MetaTrader 5 integration using Go and gRPC"

## Clarifications

### Session 2026-05-08

- Q: How should the MCP server **connect to MT5 terminal**? → A: **gRPC daemon local** (desacoplado, resiliente a crashes MT5)
- Q: Terminal desconecta durante operação — comportamento esperado? → A: **Auto-reconectar com fila pending** (operações bufferizadas, transparente ao cliente)
- Q: Múltiplas ordens simultâneas (concurrent) — conflito de preço? → A: **Processamento independente, erro MT5 propagado** (cliente responsável por retry + lógica)
- Q: Credenciais MT5 em dev/test → armazenamento? → A: **Env vars (dev) + Secrets Manager (prod)**, rotação automática
- Q: Escopo explícito — o que está IN-SCOPE? → A: Account info, quotes, place orders, close positions, query pending orders, MCP JSON-RPC 2.0, gRPC+TLS, health check, error messages pt-BR, concurrent invocations (≥10)
- Q: Persistência & limites fila auto-reconnect? → A: **SQLite local, durável cross-restart, máx unlimited, FIFO priority**
- Q: Idempotência & sequenciamento ordens concurrent? → A: **Idempotency key obrigatória (UUID), FIFO sequencing por account, 1 fill garantido**
- Q: Estratégia testes — MT5 real vs mock? → A: **MT5 real para integration, mock para unit, CI/CD skip integration (manual)**

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Query MT5 Account Status (Priority: P1)

Trading AI agent needs to check account balance, equity, margin before making decisions.

**Why this priority**: Core information needed for risk management and trade decisions. Without account visibility, no automated trading can occur.

**Independent Test**: Can be fully tested by invoking account-info tool and verifying correct balance/equity/margin values returned from MT5 terminal.

**Acceptance Scenarios**:

1. **Given** MT5 terminal is connected and logged in, **When** account-info tool is called, **Then** return account balance, equity, margin, and free margin within 2 seconds
2. **Given** MT5 terminal is disconnected, **When** account-info tool is called, **Then** return error with clear disconnection message

---

### User Story 2 - Get Market Quotes (Priority: P1)

Trading AI agent needs current bid/ask prices for trading instruments.

**Why this priority**: Price data is essential for any trading decision. Without quotes, no order placement possible.

**Independent Test**: Can be fully tested by invoking quote tool with instrument symbol and verifying bid/ask prices are returned.

**Acceptance Scenarios**:

1. **Given** MT5 terminal connected, **When** quote tool called with "EURUSD", **Then** return current bid/ask with timestamp within 500ms
2. **Given** instrument not in market watch, **When** quote tool called, **Then** return error indicating instrument not available

---

### User Story 3 - Place Trading Orders (Priority: P1)

Trading AI agent needs to execute buy/sell orders through MT5.

**Why this priority**: Core trading functionality. Enables automated order execution based on AI decisions.

**Independent Test**: Can be fully tested by placing market order and verifying order ticket returned from MT5.

**Acceptance Scenarios**:

1. **Given** sufficient margin and market open, **When** buy order placed for 0.1 lot EURUSD, **Then** return order ticket and fill price
2. **Given** insufficient margin, **When** order placed, **Then** return error with margin requirement details
3. **Given** market closed, **When** order placed, **Then** return error indicating market unavailable

---

### User Story 4 - Manage Positions (Priority: P2)

Trading AI agent needs to close or modify existing positions.

**Why this priority**: Position management enables stop-loss/take-profit automation and portfolio rebalancing.

**Independent Test**: Can be fully tested by closing a position and verifying position removed from MT5.

**Acceptance Scenarios**:

1. **Given** open position exists, **When** close-position tool called with ticket, **Then** position closed and profit/loss returned
2. **Given** position already closed, **When** close-position tool called, **Then** return error indicating position not found

---

### User Story 5 - Query Open Orders (Priority: P2)

Trading AI agent needs to monitor pending orders and their status.

**Why this priority**: Visibility into pending orders necessary for order management and conflict prevention.

**Independent Test**: Can be fully tested by querying orders and verifying returned list matches actual pending orders in MT5.

**Acceptance Scenarios**:

1. **Given** pending orders exist, **When** orders-list tool called, **Then** return list of all pending orders with ticket, type, volume, price
2. **Given** no pending orders, **When** orders-list tool called, **Then** return empty list

---

### User Story 6 - Get Historical Candles (Priority: P2)

Trading AI agent needs historical OHLC data for technical analysis and signal generation.

**Why this priority**: Enables AI to analyze price action, support/resistance levels, trend confirmation before trading decisions.

**Independent Test**: Can be fully tested by querying candles for symbol and verifying returned OHLC matches MT5 chart data.

**Acceptance Scenarios**:

1. **Given** EURUSD H1 timeframe, **When** get-candles called with count=100, **Then** return 100 candles with open/high/low/close/volume/timestamp
2. **Given** 5m timeframe requested, **When** get-candles called, **Then** return 5-minute candles (M5, M15, H1, D, W supported)
3. **Given** instrument has no history, **When** get-candles called, **Then** return empty candle list or error

---

### Edge Cases

- MT5 terminal disconnects during order placement → MCP auto-reconnects, queues order to SQLite, retries on reconnection FIFO; client receives pending status; orders survive daemon crash
- Concurrent orders on same symbol (price gap risk) → orders processed FIFO per account; idempotency key prevents duplicates; MT5 enforces sequencing via timestamp
- Client retries order with same idempotency key within 24h → MCP returns cached fill result (exactly-once semantics); no duplicate fill on MT5
- Pending queue accumulates >1000 orders (large disconnect) → process FIFO, no arbitrary limit; SQLite handles persistence
- MT5 returns duplicate ticket numbers → use most recent, log warning
- Network latency exceeds 5 seconds → return timeout error
- MT5 requires re-login → detect and return authentication error
- Order rejected by MT5 due to price gapping → return rejection reason to client; client can retry with new idempotency key

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: MCP server MUST expose tools following MCP JSON-RPC 2.0 specification
- **FR-002**: Tools MUST handle MT5 terminal disconnection gracefully with clear error messages
- **FR-003**: All financial calculations MUST use decimal precision — no floating point for currency values
- **FR-004**: Tools MUST return structured JSON responses with consistent error format
- **FR-005**: System MUST log all tool invocations with latency metrics for observability
- **FR-006**: MCP server MUST support concurrent tool invocations without race conditions
- **FR-007**: gRPC services MUST use TLS in production environments
- **FR-008**: MT5 credentials MUST be retrieved from environment variables or secrets manager
- **FR-009**: Tools MUST validate input parameters against defined schemas before passing to MT5
- **FR-010**: MCP server MUST expose health check endpoint for terminal connection status
- **FR-011**: Terminal connection MUST use gRPC daemon (local, decoupled from MT5 process) for resilience
- **FR-012**: MCP server MUST auto-reconnect on terminal disconnect and buffer pending operations in queue until reconnected
- **FR-013**: Scope explicitly includes: account-info, market quotes, place market orders, close positions, query pending orders, historical candles (OHLC). Excludes: technical analysis indicators (SMA/RSI/MACD), backtesting, copytrading, tick-by-tick data, modify pending orders (only close supported)
- **FR-014**: Pending operations queue MUST be persisted to local SQLite; survive daemon restart; reprocess FIFO on reconnection within 60s window; max unlimited entries
- **FR-015**: Place-order tool MUST accept optional `idempotency_key` (UUID); deduplicate by key; guarantee exactly-once fill semantics; reject duplicate keys with cached response
- **FR-016**: Order sequencing MUST be FIFO per trading account; prevent margin/position race conditions; timestamp enforcement via MT5

### Key Entities

- **MT5Terminal**: Represents MetaTrader 5 terminal instance. Attributes: connection state, last heartbeat, terminal path.
- **TradingAccount**: MT5 trading account. Attributes: account number, balance, equity, margin, free margin, currency.
- **Instrument**: Tradeable symbol (EURUSD, etc.). Attributes: symbol, description, digits, tick value.
- **Quote**: Real-time price data. Attributes: symbol, bid, ask, timestamp.
- **Order**: Trading order (market or pending). Attributes: ticket, type (buy/sell), volume, price, stop loss, take profit, status.
- **Position**: Open trading position. Attributes: ticket, symbol, type, volume, open price, current price, profit.
- **Candle**: OHLC price bar. Attributes: symbol, timeframe (M1/M5/M15/H1/D/W), open, high, low, close, volume, timestamp.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Account info tool returns data within 2 seconds under normal conditions
- **SC-002**: Quote tool returns prices within 500ms for available instruments
- **SC-003**: Order placement returns ticket or error within 5 seconds
- **SC-004**: MCP server handles 10 concurrent tool invocations without errors
- **SC-005**: All error responses include clear human-readable messages in Portuguese (pt-BR)
- **SC-006**: Terminal disconnection detected and reported within 10 seconds
- **SC-007**: Zero hardcoded credentials in codebase
- **SC-008**: All MCP tools have corresponding integration tests with MT5
- **SC-009**: Credentials rotated automatically per Secrets Manager policy (≤90 days); zero hardcoded in code or logs
- **SC-010**: gRPC daemon reconnect latency <10 seconds; queued operations processed FIFO on reconnect within 60s window
- **SC-011**: Pending queue persisted to SQLite; zero orders lost on daemon crash/restart; verified via integration test (crash sim)
- **SC-012**: Idempotency key (optional UUID) prevents duplicate fills; same key within 24h returns cached response; verified via concurrent order test
- **SC-013**: Order sequencing FIFO per account enforced; no margin race conditions detected in stress test (10 concurrent orders, margin edge case)
- **SC-014**: Integration tests use real MT5 (demo account); unit tests use mock MT5; CI/CD skips integration (manual trigger only)
- **SC-015**: Get-candles tool returns 100 H1 candles within 1 second; supports M5/M15/H1/D/W timeframes; validates symbol exists before querying

## Assumptions

- MT5 terminal runs on Windows server accessible from MCP server
- Users are automated trading systems (not manual traders)
- MT5 API (MT5 DLL or COM) available for Go integration
- Single MT5 terminal per MCP server instance (no distributed terminals)
- Trading account has sufficient permissions for all operations
- Portuguese (pt-BR) is the primary language for user-facing messages