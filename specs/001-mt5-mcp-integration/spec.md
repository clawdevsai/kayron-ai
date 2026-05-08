# Feature Specification: MT5 MCP Integration

**Feature Branch**: `001-mt5-mcp-integration`
**Created**: 2026-05-08
**Status**: Draft
**Input**: User description: "MCP tools for MetaTrader 5 integration using Go and gRPC"

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

### Edge Cases

- MT5 terminal disconnects during order placement → return error, do not leave order in ambiguous state
- MT5 returns duplicate ticket numbers → use most recent, log warning
- Network latency exceeds 5 seconds → return timeout error
- MT5 requires re-login → detect and return authentication error
- Order rejected by MT5 due to price gapping → return rejection reason to client

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

### Key Entities

- **MT5Terminal**: Represents MetaTrader 5 terminal instance. Attributes: connection state, last heartbeat, terminal path.
- **TradingAccount**: MT5 trading account. Attributes: account number, balance, equity, margin, free margin, currency.
- **Instrument**: Tradeable symbol (EURUSD, etc.). Attributes: symbol, description, digits, tick value.
- **Quote**: Real-time price data. Attributes: symbol, bid, ask, timestamp.
- **Order**: Trading order (market or pending). Attributes: ticket, type (buy/sell), volume, price, stop loss, take profit, status.
- **Position**: Open trading position. Attributes: ticket, symbol, type, volume, open price, current price, profit.

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

## Assumptions

- MT5 terminal runs on Windows server accessible from MCP server
- Users are automated trading systems (not manual traders)
- MT5 API (MT5 DLL or COM) available for Go integration
- Single MT5 terminal per MCP server instance (no distributed terminals)
- Trading account has sufficient permissions for all operations
- Portuguese (pt-BR) is the primary language for user-facing messages