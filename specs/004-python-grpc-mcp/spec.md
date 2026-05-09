# Feature Specification: Python + gRPC MCP Migration

**Feature Branch**: `004-python-grpc-mcp`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: Migrate MT5 MCP to Python + gRPC, enable any agent to use the service

## Clarifications

### Session 2026-05-09

- Q: Symbol-level serialization strategy for concurrent orders? → A: Allow concurrent execution; rely on MT5 SDK thread-safety
- Q: Operation queueing/async model? → A: Server push/callback with bidirectional gRPC streaming
- Q: MT5 offline behavior? → A: Queue operations and auto-retry when reconnected
- Q: Scope of MT5 operations v1? → A: All operations documented in MT5 SDK

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Multi-Agent MT5 Access via gRPC (Priority: P1)

**Actor**: Trading agents, automation platforms, third-party services

Agents need standardized, language-agnostic access to MetaTrader 5 trading operations. Currently tightly coupled to specific platforms; need decoupled gRPC service accessible from any agent (Python, Go, Node.js, Java, etc.).

**Why this priority**: Enables ecosystem of integrations. Core business value—agents can connect to live trading data and execute operations without reimplementing MT5 bindings.

**Independent Test**: Deploy gRPC server. Connect agent in different language. Execute buy/sell order, fetch account info, verify results. Full trading workflow end-to-end.

**Acceptance Scenarios**:

1. **Given** gRPC server running with bidirectional stream, **When** agent sends `GetAccountInfo` request, **Then** server immediately pushes current balance, equity, margin data to callback
2. **Given** agent authenticated and server at capacity, **When** sending `PlaceOrder` request, **Then** order queued and server pushes execution result via callback once MT5 connection available
3. **Given** multiple agents connected, **When** all send concurrent orders for same symbol, **Then** all execute via MT5 SDK concurrently with consistent results
4. **Given** connection drops mid-operation, **When** agent reconnects, **Then** queued operation auto-executes and result pushed to agent; no data loss

---

### User Story 2 - Session & Connection Management (Priority: P1)

**Actor**: MCP server, agents

Server must manage MT5 connections—only one active connection per MT5 terminal instance. Agents may be short-lived; server maintains long-lived pool.

**Why this priority**: MT5 only allows single active connection. Without pooling, agents fight for exclusive access. Server must enforce this transparently.

**Independent Test**: Launch server with MT5 connection pool. Connect 10 simultaneous agents. Issue commands in round-robin. Verify all commands execute without `"connection already in use"` errors.

**Acceptance Scenarios**:

1. **Given** agent requests trading operation, **When** pool has available MT5 connection, **Then** operation queues and executes
2. **Given** connection is held by agent A, **When** agent B requests operation, **Then** queued until connection available
3. **Given** agent disconnects, **When** connection released to pool, **Then** next queued agent gets connection
4. **Given** MT5 connection stale, **When** reuse attempt, **Then** reconnect automatically before operation

---

### User Story 3 - Error Resilience & Logging (Priority: P2)

**Actor**: Operations team, agent developers

Errors from MT5 must be surfaced clearly (network, auth, platform errors). Server logs all operations for debugging, compliance, audit trail.

**Why this priority**: Production reliability. Developers need clear error messages. Compliance requires operation audit trail. Reduces MTTR (mean time to repair).

**Independent Test**: Cause failures (invalid login, network disconnect, malformed requests). Verify error codes/messages are actionable. Check logs contain operation history and timestamp.

**Acceptance Scenarios**:

1. **Given** invalid MT5 credentials, **When** agent sends operation, **Then** receive `UNAUTHENTICATED` error code with reason
2. **Given** network glitch during operation, **When** gRPC connection lost, **Then** server reconnects and retries idempotent ops
3. **Given** operation executed, **When** checked in logs, **Then** log entry shows timestamp, agent ID, operation type, result

---

### Edge Cases

- MT5 terminal offline: Server queues operations and auto-retries when reconnected (at-least-once semantics)
- Invalid/malformed gRPC message: Server gracefully rejects with clear error code and message
- Agent connection not properly closed: Server-side timeout triggers cleanup and resource release
- Multiple agents request same symbol concurrently: Operations execute concurrently; MT5 SDK thread-safety ensures consistency

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: MCP server MUST expose all MT5 trading operations documented in MT5 SDK via gRPC endpoints (all account, trading, symbol, order operations)
- **FR-002**: Server MUST authenticate agents (API key, OAuth, or other scheme) before allowing operations
- **FR-003**: Server MUST maintain a connection pool to MT5 terminal(s), enforcing single-connection-per-terminal constraint
- **FR-004**: Server MUST support concurrent execution of operations; rely on MT5 SDK thread-safety to prevent race conditions
- **FR-004b**: Server MUST support bidirectional gRPC streaming for operation callbacks—when queued operation completes, server pushes result to agent
- **FR-004c**: Server MUST queue operations when MT5 connection unavailable and auto-execute on reconnect (at-least-once delivery)
- **FR-005**: Agents MUST NOT directly import or depend on MT5 bindings—all access MUST go through gRPC service
- **FR-006**: Server MUST log all operations (timestamp, agent identity, operation type, result, errors) for audit and debugging
- **FR-007**: Server MUST expose health check endpoint (readiness, liveness) for orchestration/monitoring
- **FR-008**: Server MUST handle connection losses and automatically reconnect without manual intervention
- **FR-009**: Agents MUST receive structured error responses with error codes (e.g., `UNAUTHENTICATED`, `UNAVAILABLE`, `INVALID_ARGUMENT`)
- **FR-010**: Server MUST support graceful shutdown—complete in-flight operations, reject new requests

### Key Entities

- **MT5 Connection**: Managed resource representing active session to MT5 terminal
- **Agent Session**: Authenticated agent's logical session; maps to shared MT5 connection pool
- **Trading Operation**: Request (place order, cancel, query) from agent to execute on MT5
- **Operation Log**: Audit record containing operation details, timestamp, result

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Agents can access MT5 operations through gRPC within 100ms latency (p95) for standard queries
- **SC-002**: Server handles 50+ concurrent agents without degradation or dropped requests
- **SC-003**: Operation success rate ≥ 99.5% (excluding intentional user errors like invalid credentials)
- **SC-004**: Server maintains ≤ 5-minute recovery time from MT5 connection loss
- **SC-005**: All operations audited in logs with 100% capture rate (no silent failures)
- **SC-006**: Agents implemented in ≥ 3 different languages (Python, Go, Node.js) can successfully integrate with server
- **SC-007**: Server shutdown completes within 30 seconds; all in-flight operations complete or fail cleanly
- **SC-008**: Bidirectional streaming callbacks deliver results to agents within 1 second of MT5 operation completion
- **SC-009**: Queued operations survive server restart; 100% recovery on reconnect (durability)

## Assumptions

- MT5 terminal already installed and accessible on server machine with valid license
- Single MT5 terminal instance per server deployment (not multi-terminal setup)
- Agents authenticate using pre-shared API keys (more sophisticated auth deferred to v2)
- Network connectivity between agents and server is stable; transient failures handled via gRPC retries
- Python 3.8+ available on server
- gRPC client libraries available for agents' target languages (standard, no custom codegen)
- Logging infrastructure (file/syslog/ELK) available; server writes to stdout/stderr
