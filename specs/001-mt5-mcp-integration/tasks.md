# Tasks: MT5 MCP Integration

**Feature**: 001-mt5-mcp-integration  
**Branch**: `001-mt5-mcp-integration`  
**Date Generated**: 2026-05-08  
**Spec**: `specs/001-mt5-mcp-integration/spec.md`

## Phase 1: Setup & Project Initialization

**Purpose**: Initialize Go project, module structure, and development environment  
**Estimated**: 2-3 hours  
**Success Criteria**: Project builds, tests run, dev environment ready

- [ ] T001 [P] Initialize Go module and project structure in cmd/mcp-mt5-server
- [ ] T002 [P] Create go.mod with dependencies (gRPC, protobuf, shops/decimal, SQLite)
- [ ] T003 [P] Setup Protocol Buffer compilation pipeline with Makefile or build script
- [ ] T004 [P] Create internal/ package structure (models, services, contracts, logger)
- [ ] T005 [P] Setup tests/ directory with integration/ and contract/ subdirectories
- [ ] T006 [P] Configure .gitignore for Go project (binaries, vendor/, *.pb.go)
- [ ] T007 Setup development environment documentation in docs/DEVELOPMENT.md

**Checkpoint**: Go project compiles, protoc installed, test framework ready

---

## Phase 2: Foundational Infrastructure (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story  
**Estimated**: 8-10 hours  
**Success Criteria**: gRPC daemon operational, MT5 WebAPI client working, queue schema ready, error handling in place

- [ ] T008 [P] Design and create gRPC service definition (mt5.proto) with 5 tool schemas
- [ ] T009 [P] Generate gRPC Go stubs from protobuf contracts in internal/contracts/
- [ ] T010 [P] Implement MT5 WebAPI HTTP client in internal/services/mt5/client.go with auth handling
- [ ] T011 [P] Create gRPC daemon skeleton in internal/services/daemon/daemon.go with auto-reconnect logic
- [ ] T012 [P] Implement SQLite schema for pending operations queue in internal/models/queue_schema.sql
- [ ] T013 [P] Create queue persistence layer in internal/models/queue.go (enqueue, dequeue, FIFO)
- [ ] T014 Implement structured logger in internal/logger/logger.go (JSON format, latency metrics)
- [ ] T015 [P] Create error types and handling in internal/errors/errors.go (MCP-compliant, pt-BR messages)
- [ ] T016 [P] Setup environment configuration in internal/config/config.go (env vars, secrets mgmt)
- [ ] T017 Setup health check endpoint in internal/services/health/health.go
- [ ] T018 [P] Create MCP server skeleton in cmd/mcp-mt5-server/main.go (JSON-RPC 2.0 handler)
- [ ] T019 Implement auto-reconnect mechanism in internal/services/daemon/reconnect.go (<10s detection)
- [ ] T020 [P] Create base test utilities in tests/test_helpers.go (fixtures, mocks)
- [ ] T021 [P] Setup CI/CD test runner configuration (.github/workflows/test.yml or equivalent)

**Checkpoint**: gRPC daemon compiles, MT5 client authenticates, queue table created, error handling works, health check responds

---

## Phase 3: User Story 1 - Get Account Information (Priority: P1)

**Goal**: Retrieve trading account balance, equity, margin, and account metadata from MT5  
**Independent Test**: Query account via account-info tool; verify balance, equity, margin returned within 2 seconds

### Tests for User Story 1

- [ ] T022 [P] [US1] Contract test for account-info tool in tests/contract/test_account_info.go
- [ ] T023 [P] [US1] Integration test for account retrieval from MT5 in tests/integration/test_account_info.go
- [ ] T024 [US1] Test error handling for disconnected terminal in tests/integration/test_account_info.go

### Implementation for User Story 1

- [ ] T025 [P] [US1] Create TradingAccount model in internal/models/account.go (balance, equity, margin, currency)
- [ ] T026 [P] [US1] Implement MT5 account query in internal/services/mt5/account_service.go
- [ ] T027 [P] [US1] Implement gRPC AccountInfo service in internal/services/daemon/account_service.go
- [ ] T028 [US1] Create account-info MCP tool handler in internal/services/mcp/account_tool.go
- [ ] T029 [US1] Wire account-info tool into MCP server in cmd/mcp-mt5-server/main.go
- [ ] T030 [US1] Add account-info integration test with real MT5 terminal in tests/integration/

**Checkpoint**: User Story 1 is independently functional. Account-info tool returns structured data within 2s.

---

## Phase 4: User Story 2 - Get Market Quotes (Priority: P1)

**Goal**: Fetch real-time bid/ask prices and instrument metadata from MT5  
**Independent Test**: Query quotes for symbol via quote tool; verify bid < ask, timestamp valid

### Tests for User Story 2

- [ ] T031 [P] [US2] Contract test for quote tool in tests/contract/test_quote.go
- [ ] T032 [P] [US2] Integration test for quote retrieval from MT5 in tests/integration/test_quote.go
- [ ] T033 [US2] Test quote caching/stale data handling in tests/integration/test_quote.go

### Implementation for User Story 2

- [ ] T034 [P] [US2] Create Instrument and Quote models in internal/models/quote.go (bid, ask, spread validation)
- [ ] T035 [P] [US2] Implement MT5 symbol/quote query in internal/services/mt5/quote_service.go
- [ ] T036 [P] [US2] Implement gRPC GetQuote service in internal/services/daemon/quote_service.go
- [ ] T037 [US2] Create quote MCP tool handler in internal/services/mcp/quote_tool.go
- [ ] T038 [US2] Wire quote tool into MCP server in cmd/mcp-mt5-server/main.go
- [ ] T039 [US2] Add quote integration test with real MT5 terminal in tests/integration/

**Checkpoint**: User Stories 1 AND 2 both work independently. Quote tool returns bid/ask within 500ms.

---

## Phase 5: User Story 3 - Place Trading Orders (Priority: P1)

**Goal**: Submit market/pending orders with decimal precision, handle idempotency, prevent duplicates  
**Independent Test**: Place order, verify ticket assigned; retry with same idempotency key returns cached result

### Tests for User Story 3

- [ ] T040 [P] [US3] Contract test for place-order tool in tests/contract/test_place_order.go
- [ ] T041 [P] [US3] Integration test for order placement in tests/integration/test_place_order.go
- [ ] T042 [P] [US3] Test idempotency key deduplication in tests/integration/test_place_order.go
- [ ] T043 [US3] Test order validation (volume, price, symbol) in tests/contract/test_place_order.go
- [ ] T044 [US3] Test concurrent orders FIFO sequencing in tests/integration/test_place_order.go

### Implementation for User Story 3

- [ ] T045 [P] [US3] Create Order model in internal/models/order.go (ticket, type, volume, price, status, profit/loss)
- [ ] T046 [P] [US3] Implement idempotency key cache in internal/models/idempotency_cache.go (UUID deduplication, 24h TTL)
- [ ] T047 [P] [US3] Implement MT5 order placement in internal/services/mt5/order_service.go (shops/decimal precision)
- [ ] T048 [US3] Implement gRPC PlaceOrder service in internal/services/daemon/order_service.go (idempotency handling, FIFO queue)
- [ ] T049 [US3] Create place-order MCP tool handler in internal/services/mcp/place_order_tool.go (validation + error responses)
- [ ] T050 [US3] Wire place-order tool into MCP server in cmd/mcp-mt5-server/main.go
- [ ] T051 [US3] Implement pending order queue persistence in internal/models/queue.go (SQLite FIFO)
- [ ] T052 [US3] Add order placement integration test with real MT5 terminal in tests/integration/

**Checkpoint**: User Stories 1, 2, AND 3 work independently. Orders placed with exactly-once semantics via idempotency key.

---

## Phase 6: User Story 4 - Manage Positions (Priority: P2)

**Goal**: Close open positions, return profit/loss; support position modification (stop-loss/take-profit)  
**Independent Test**: Open position, close via close-position tool, verify removed from MT5; profit/loss calculated

### Tests for User Story 4

- [ ] T053 [P] [US4] Contract test for close-position tool in tests/contract/test_close_position.go
- [ ] T054 [P] [US4] Integration test for position closure in tests/integration/test_close_position.go
- [ ] T055 [US4] Test error on non-existent position in tests/integration/test_close_position.go

### Implementation for User Story 4

- [ ] T056 [P] [US4] Create Position model in internal/models/position.go (ticket, symbol, type, volume, entry price, current price, profit)
- [ ] T057 [P] [US4] Implement MT5 position query in internal/services/mt5/position_service.go
- [ ] T058 [P] [US4] Implement MT5 position closure in internal/services/mt5/position_service.go (profit calculation)
- [ ] T059 [US4] Implement gRPC ClosePosition service in internal/services/daemon/position_service.go
- [ ] T060 [US4] Create close-position MCP tool handler in internal/services/mcp/close_position_tool.go
- [ ] T061 [US4] Wire close-position tool into MCP server in cmd/mcp-mt5-server/main.go
- [ ] T062 [US4] Add position closure integration test with real MT5 terminal in tests/integration/

**Checkpoint**: User Story 4 is independently functional. Positions closed with profit/loss returned.

---

## Phase 7: User Story 5 - Query Pending Orders (Priority: P2)

**Goal**: List all pending orders (unfilled, not yet executed) per account  
**Independent Test**: Place limit order, query via orders-list tool, verify order in list with correct status

### Tests for User Story 5

- [ ] T063 [P] [US5] Contract test for orders-list tool in tests/contract/test_orders_list.go
- [ ] T064 [P] [US5] Integration test for pending orders query in tests/integration/test_orders_list.go
- [ ] T065 [US5] Test empty orders list (no pending) in tests/integration/test_orders_list.go

### Implementation for User Story 5

- [ ] T066 [P] [US5] Implement MT5 pending orders query in internal/services/mt5/orders_service.go
- [ ] T067 [P] [US5] Implement gRPC GetOrders service in internal/services/daemon/orders_service.go (list + filter)
- [ ] T068 [US5] Create orders-list MCP tool handler in internal/services/mcp/orders_list_tool.go
- [ ] T069 [US5] Wire orders-list tool into MCP server in cmd/mcp-mt5-server/main.go
- [ ] T070 [US5] Add orders list integration test with real MT5 terminal in tests/integration/

**Checkpoint**: All user stories (US1-5) are independently functional. Five MCP tools exposed.

---

## Phase 8: Polish & Production Readiness

**Purpose**: Error handling, observability, concurrent testing, documentation, security hardening  
**Estimated**: 6-8 hours  
**Success Criteria**: Error messages in pt-BR, structured logging with latency, 10 concurrent tools, zero hardcoded credentials, TLS configured, docs complete

### Error Handling & Observability

- [ ] T071 [P] Add comprehensive error handling for all MT5 failure modes (disconnect, timeout, invalid credentials, margin error, symbol not found)
- [ ] T072 [P] Implement structured JSON logging for all tool invocations (timestamp, tool name, input, output, latency_ms, error)
- [ ] T073 [P] Add latency metrics collection in internal/logger/latency.go (p50, p95, p99 percentiles per tool)
- [ ] T074 Translate all error messages to Portuguese (pt-BR) in internal/errors/pt_br.go
- [ ] T075 [P] Add terminal connection health monitoring in internal/services/health/health.go (heartbeat detection <10s)
- [ ] T076 Setup debug logging mode with -debug flag in cmd/mcp-mt5-server/main.go

### Concurrent Testing & Scalability

- [ ] T077 [P] Create load test in tests/load/load_test.go (10 concurrent tool invocations, measure latency)
- [ ] T078 [P] Test race conditions with concurrent orders on same symbol in tests/integration/test_concurrent.go
- [ ] T079 Test queue handling under high volume (1000+ pending orders) in tests/integration/test_queue_load.go
- [ ] T080 [P] Verify FIFO sequencing with concurrent requests in tests/integration/test_fifo_ordering.go

### Security & Credentials

- [ ] T081 [P] Verify zero hardcoded credentials in all source files (scan for hardcoded strings)
- [ ] T082 [P] Implement TLS configuration for gRPC in production mode (internal/config/tls.go)
- [ ] T083 [P] Setup Secrets Manager integration for credential rotation in internal/config/secrets.go
- [ ] T084 [P] Add credential audit logging (login attempts, rotations) without exposing secrets
- [ ] T085 Verify no credentials in logs or error messages (redact sensitive data)

### Documentation & Release

- [ ] T086 Write gRPC service documentation in docs/GRPC_SERVICES.md (proto definitions, request/response examples)
- [ ] T087 Write MCP tool documentation in docs/MCP_TOOLS.md (tool names, inputs, outputs, errors, pt-BR examples)
- [ ] T088 Create troubleshooting guide in docs/TROUBLESHOOTING.md (common errors, terminal disconnection, credential issues)
- [ ] T089 Write architecture documentation in docs/ARCHITECTURE.md (gRPC daemon, queue persistence, auto-reconnect)
- [ ] T090 Update README.md with setup instructions, tech stack, and quick start
- [ ] T091 Create CHANGELOG.md with v1.0.0 release notes
- [ ] T092 [P] Generate API documentation from protobuf (godoc comments, proto annotations)

### Final Integration & Verification

- [ ] T093 [P] Run full integration test suite with real MT5 terminal (all 5 tools end-to-end)
- [ ] T094 [P] Verify all error messages are in pt-BR and human-readable
- [ ] T095 [P] Performance verification (all tools respond within SLA: account <2s, quote <500ms, order <5s)
- [ ] T096 Verify terminal auto-reconnect within 10 seconds in tests/integration/test_reconnect.go
- [ ] T097 Verify queue persistence survives daemon restart in tests/integration/test_queue_persistence.go
- [ ] T098 Code review by second pair (security, gRPC patterns, error handling)
- [ ] T099 Security audit (credentials, TLS, input validation, SQL injection prevention)
- [ ] T100 Final documentation review and link verification

**Checkpoint**: Production-ready MCP server with full error handling, observability, security, and documentation.

---

## Summary & Execution

### Task Distribution by Phase

| Phase | Tasks | Hours | Status |
|-------|-------|-------|--------|
| Phase 1: Setup | T001-T007 | 2-3h | Pending |
| Phase 2: Foundational | T008-T021 | 8-10h | Pending |
| Phase 3: US1 (P1) | T022-T030 | 6-8h | Pending |
| Phase 4: US2 (P1) | T031-T039 | 6-8h | Pending |
| Phase 5: US3 (P1) | T040-T052 | 12-14h | Pending |
| Phase 6: US4 (P2) | T053-T062 | 8-10h | Pending |
| Phase 7: US5 (P2) | T063-T070 | 6-8h | Pending |
| Phase 8: Polish | T071-T100 | 6-8h | Pending |
| **TOTAL** | **100 tasks** | **54-69h** | **Ready** |

### Parallel Execution Strategy

**Phase 1 (Setup)**: All [P] tasks can run in parallel → 2-3 hours  
**Phase 2 (Foundational)**: All [P] tasks can run in parallel → 6-8 hours (gRPC, client, queue, logging)  
**Phase 3-7 (User Stories)**: Can run in parallel per story (3 P1 stories + 2 P2 stories) → 12-20 hours total with team  
**Phase 8 (Polish)**: All [P] tasks run in parallel → 4-6 hours  

**Optimal Staffing**: 3-5 developers  
**Sequential Timeline**: ~12 weeks (1 person)  
**Parallel Timeline**: ~4-5 weeks (3 developers, organized by phase)

---

## Task Checklist Format Validation

OK Format Rules Verified:
- Task IDs sequential (T001-T100)
- Priority markers: [P] = parallelizable
- US markers: [US1], [US2], [US3], [US4], [US5]
- File paths: exact internal/, cmd/, tests/ locations
- Phase headers with estimates and success criteria
- Phase checkpoints after each major milestone
- Dependency documentation in Phase Dependencies section
- Parallel opportunities clearly marked

---

## Notes for Implementation

1. **MT5 Terminal Setup**: Ensure MT5 terminal runs locally on port 8228 before starting Phase 2
2. **Decimal Precision**: Use `github.com/shopspring/decimal` for all financial calculations (no float64)
3. **Idempotency**: UUID key format: `{account_id}_{timestamp}_{nonce}` for 24-hour deduplication
4. **FIFO Queue**: SQLite table schema in T012 must include (id, account_id, operation, payload, created_at, status)
5. **pt-BR Messages**: All error strings in internal/errors/pt_br.go (e.g., "Saldo insuficiente")
6. **TLS in Prod**: gRPC daemon must use TLS when `ENV=production` (generate certs with mkcert or Let's Encrypt)
7. **Health Check**: Expose HTTP health endpoint at localhost:8229/health (status, terminal_connected, queue_length)
8. **Testing**: MT5 test fixtures required (dummy account, symbols); mock layer for unit tests

---

**Generated**: 2026-05-08  
**Template**: .specify/templates/tasks-template.md  
**Spec**: specs/001-mt5-mcp-integration/spec.md

---

## Phase 3: User Story 1 — Query MT5 Account Status (Priority: P1) 🎯 MVP

**Goal**: `account-info` tool returns balance, equity, margin, free margin

**Independent Test**: Invoke `account-info` tool → verify structured response with all account fields

### Implementation

- [ ] T018 [US1] Add `AccountInfo` method to `internal/mt5/account.go`
- [ ] T019 [US1] Register `account-info` tool in `internal/mcp/tools.go` with input/output schemas
- [ ] T020 [US1] Wire `account-info` handler in `internal/mcp/handler.go` → calls MT5 WebAPI
- [ ] T021 [US1] Add unit tests in `tests/unit/account_test.go` (mock MT5 WebAPI responses)
- [ ] T022 [US1] Add integration test in `tests/integration/test_account_info.go` (if MT5 terminal available)

**Checkpoint**: `account-info` tool works independently. Returns balance/equity/margin within 2s.

---

## Phase 4: User Story 2 — Get Market Quotes (Priority: P1)

**Goal**: `quote` tool returns bid/ask for instrument symbol

**Independent Test**: Invoke `quote` tool with "EURUSD" → verify bid < ask, valid timestamp

### Implementation

- [ ] T023 [US2] Add `GetQuote` method to `internal/mt5/quote.go`
- [ ] T024 [US2] Register `quote` tool in `internal/mcp/tools.go` with symbol input schema
- [ ] T025 [US2] Wire `quote` handler in `internal/mcp/handler.go`
- [ ] T026 [US2] Add unit tests in `tests/unit/quote_test.go` (mock WebAPI quote response)
- [ ] T027 [US2] Add integration test in `tests/integration/test_quote.go`

**Checkpoint**: `quote` tool works. Returns bid/ask within 500ms for available instruments.

---

## Phase 5: User Story 3 — Place Trading Orders (Priority: P1)

**Goal**: `place-order` tool executes buy/sell/pending orders

**Independent Test**: Place market order → verify ticket returned or error returned for invalid margin

### Implementation

- [ ] T028 [US3] Add `PlaceOrder` method to `internal/mt5/order.go`
- [ ] T029 [US3] Register `place-order` tool in `internal/mcp/tools.go` with full input schema (symbol, type, volume, price, stopLoss, takeProfit, comment)
- [ ] T030 [US3] Wire `place-order` handler in `internal/mcp/handler.go`
- [ ] T031 [US3] Add input validation: volume multiple of minimum lot, price within range
- [ ] T032 [US3] Add unit tests in `tests/unit/order_test.go`
- [ ] T033 [US3] Add integration test in `tests/integration/test_place_order.go`

**Checkpoint**: `place-order` tool works. Order placement returns ticket within 5s or clear error.

---

## Phase 6: User Story 4 — Manage Positions (Priority: P2)

**Goal**: `close-position` tool closes open position by ticket

**Independent Test**: Close open position → verify profit returned, position no longer exists

### Implementation

- [ ] T034 [US4] Add `ClosePosition` method to `internal/mt5/order.go`
- [ ] T035 [US4] Register `close-position` tool in `internal/mcp/tools.go`
- [ ] T036 [US4] Wire `close-position` handler in `internal/mcp/handler.go`
- [ ] T037 [US4] Add unit tests in `tests/unit/position_test.go`
- [ ] T038 [US4] Add integration test in `tests/integration/test_close_position.go`

**Checkpoint**: `close-position` tool works independently.

---

## Phase 7: User Story 5 — Query Open Orders (Priority: P2)

**Goal**: `orders-list` tool returns all pending orders

**Independent Test**: Query `orders-list` → verify returned list matches actual pending orders in MT5

### Implementation

- [ ] T039 [US5] Add `ListOrders` method to `internal/mt5/order.go`
- [ ] T040 [US5] Register `orders-list` tool in `internal/mcp/tools.go`
- [ ] T041 [US5] Wire `orders-list` handler in `internal/mcp/handler.go`
- [ ] T042 [US5] Add unit tests in `tests/unit/orders_list_test.go`
- [ ] T043 [US5] Add integration test in `tests/integration/test_orders_list.go`

**Checkpoint**: `orders-list` tool works independently. Returns empty list when no pending orders.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Health check, observability, TLS, final integration tests

- [ ] T044 [P] Implement `mt5-health` tool (terminal connection status + last heartbeat) per FR-010
- [ ] T045 [P] Add structured JSON logging with latency metrics per FR-005
- [ ] T046 [P] Implement TLS configuration for gRPC channel per FR-007
- [ ] T047 Add `buf lint` validation to Makefile (proto schema validation)
- [ ] T048 Verify `go vet` clean, all tests pass
- [ ] T049 Run `quickstart.md` validation — verify all examples work
- [ ] T050 Update README with tool reference and architecture diagram

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Stories (Phase 3-7)**: Depend on Foundational — can proceed in parallel after Phase 2
- **Polish (Phase 8)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Starts after Phase 2 — no dependencies on other stories
- **US2 (P1)**: Starts after Phase 2 — no dependencies on US1
- **US3 (P1)**: Starts after Phase 2 — no dependencies on US1/US2
- **US4 (P2)**: Starts after Phase 2 — no dependencies on US1/US2/US3
- **US5 (P2)**: Starts after Phase 2 — no dependencies on US1/US2/US3/US4

### Within Each User Story

- Domain types (already in Phase 2) before services
- Services before MCP tool registration
- Implementation before tests
- Story complete before next priority

### Parallel Opportunities

- Setup tasks T003/T004/T005 marked [P] can run in parallel
- Foundational tasks T010/T011/T013/T014 marked [P] can run in parallel
- All 5 user stories can start in parallel after Phase 2
- Polish tasks T044/T045/T046 marked [P] can run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 + 2 + 3)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3 + 4 + 5: Account info, Quote, Place-order (core trading flow)
4. **STOP and VALIDATE**: Core trading operations work
5. Deploy MVP if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 + US2 → Account + Quote → Deploy/demo
3. Add US3 → Place-order → Deploy/demo
4. Add US4 → Close-position → Deploy/demo
5. Add US5 → Orders-list → Deploy/demo
6. Polish → Health check, logging, TLS

---

## Task Summary

| Phase | Tasks | User Stories |
|-------|-------|--------------|
| Phase 1: Setup | T001-T006 (6 tasks) | — |
| Phase 2: Foundational | T007-T017 (11 tasks) | — |
| Phase 3: US1 | T018-T022 (5 tasks) | account-info |
| Phase 4: US2 | T023-T027 (5 tasks) | quote |
| Phase 5: US3 | T028-T033 (6 tasks) | place-order |
| Phase 6: US4 | T034-T038 (5 tasks) | close-position |
| Phase 7: US5 | T039-T043 (5 tasks) | orders-list |
| Phase 8: Polish | T044-T050 (7 tasks) | — |

**Total: 45 tasks**

**MVP Scope**: Phase 1 + 2 + 3 (US1 account-info) — 22 tasks to first working increment

**Suggested MVP**: Core trading — account-info + quote + place-order (US1 + US2 + US3, Phases 1-5)
