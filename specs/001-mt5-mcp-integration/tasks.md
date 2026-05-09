---
description: "MT5 MCP Integration - Complete implementation tasks for all 6 user stories"
---

# Tasks: MT5 MCP Integration (001-mt5-mcp-integration)

**Input**: Design documents from `/specs/001-mt5-mcp-integration/`  
**Prerequisites**: plan.md ✅ (required), spec.md ✅ (required), research.md ✅, data-model.md ✅, contracts/ ✅  
**Status**: Ready for Phase 2 Implementation  

**TDD Requirement**: Write tests BEFORE implementation (non-negotiable for Phase 5).  
**Decimal Precision**: All financial calculations use `shopspring/decimal` — zero float64.  
**Real MT5 Data**: All integration tests fetch REAL data from MT5 terminal. Configuration loaded from `config-real-ftmo.yaml` (path, login, password, ports, symbols, etc.). Unit tests use mocks only.

## Format: `- [ ] [ID] [P?] [Story?] Description with file path`

- **[P]**: Parallelizable (different files, no dependencies)
- **[Story]**: User story label ([US1] through [US6])
- **File paths**: Exact locations for each task
- **[Test]**: Test task — must FAIL before implementation

---

## Phase 1: Setup & Project Initialization

**Purpose**: Go module, proto pipeline, package structure.

- [ ] T001 [P] Create go.mod with dependencies in kayron-ai/ (gRPC, protobuf, shopspring/decimal, sqlite3)
- [ ] T002 [P] Create Protocol Buffer compilation target in Makefile (protoc → pb/ directory)
- [ ] T003 [P] Create internal package structure in kayron-ai/internal/{models,services,logger,errors,contracts}
- [ ] T004 [P] Create test directory structure in kayron-ai/tests/{unit,integration,contract}
- [ ] T005 [P] Verify go.mod dependencies resolve: gRPC middleware, decimal, sqlite3

---

## Phase 2: Foundational Infrastructure (CRITICAL BLOCKER)

**Purpose**: gRPC daemon, MT5 connection, auto-reconnect, idempotency, logging, errors.  
**⚠️ BLOCKS ALL USER STORIES — MUST COMPLETE FIRST.**

- [ ] T006 [P] Create config loader in internal/config/config.go (load YAML from config-real-ftmo.yaml, parse mt5/http/grpc/trading/logging/security sections)
- [ ] T006a [P] Create gRPC daemon server scaffold in cmd/mcp-mt5-server/main.go (load config, listen on grpc.port from config, health check)
- [ ] T007 [P] Create MT5 terminal connection interface in internal/services/mt5/terminal.go (use mt5.path, mt5.server, mt5.login, mt5.timeout from config)
- [ ] T008 [P] Create MT5 WebAPI HTTP client in internal/services/mt5/client.go (basic auth with login/password from config, timeout from config)
- [ ] T009 [P] Create idempotency cache interface in internal/services/cache/idempotency.go
- [ ] T010 [P] Create SQLite persistence layer in internal/storage/db.go (queue + idempotency cache)
- [ ] T011 [P] Create auto-reconnect mechanism in internal/services/daemon/reconnect.go (<10s detection)
- [ ] T012 [P] Create decimal utilities in internal/math/decimal.go (parse, multiply, compare)
- [ ] T013 [P] Create JSON logger in internal/logger/logger.go (structured logs, latency metrics)
- [ ] T014 [P] Create Portuguese error messages in internal/errors/pt_br.go ([CONNECTION_FAILED], [MARGIN_INSUFFICIENT], etc.)
- [ ] T015 [P] Create error wrapper in internal/errors/errors.go (codes, messages, translations)
- [ ] T016 [P] Implement health check endpoint in internal/services/health/health.go (terminal status)
- [ ] T017 [Test] Integration test for daemon startup in tests/integration/test_daemon_startup.go
- [ ] T018 Create MCP JSON-RPC 2.0 server scaffold in internal/services/mcp/server.go (tool registry)
- [ ] T019 Create models in internal/models/{account,quote,order,position,candle}.go
- [ ] T020 Create RiskPolicy model in internal/models/risk_policy.go (max_volume, max_positions, max_drawdown_percent)

**Config File**: All tasks load configuration from `config-real-ftmo.yaml` (path, login, password, ports, max_volume, allowed_symbols, logging, security).  
**Checkpoint**: Foundation complete — user stories can proceed in parallel.

---

## Phase 3: User Story 1 - Query MT5 Account Status (Priority: P1) 🎯 MVP

**Goal**: Return account balance, equity, margin, free margin within 2 seconds.  
**Independent Test**: Invoke account-info tool, verify balance/equity/margin.

### Tests for US1

- [ ] T021 [P] [US1] [Test] Contract test for account-info in tests/contract/test_account_info.go
- [ ] T022 [P] [US1] [Test] Unit test for AccountService in tests/unit/test_account_service.go (mock MT5 client)
- [ ] T023 [US1] [Test] Integration test for account-info end-to-end in tests/integration/test_account_info_live.go (load config-real-ftmo.yaml, connect to REAL MT5, fetch actual balance/equity)

### Implementation for US1

- [ ] T024 [P] [US1] Extend MT5 client with GetAccountInfo in internal/services/mt5/client.go
- [ ] T025 [P] [US1] Create AccountService in internal/services/daemon/account_service.go
- [ ] T026 [US1] Create account-info MCP tool in internal/services/mcp/tools/account_info.go
- [ ] T027 [US1] Wire account-info into MCP server in internal/services/mcp/server.go
- [ ] T028 [US1] Add latency metrics for account-info in internal/logger/logger.go

**Checkpoint**: US1 complete — account-info tool <2s. Testable independently.

---

## Phase 4: User Story 2 - Get Market Quotes (Priority: P1)

**Goal**: Return bid/ask prices within 500ms.  
**Independent Test**: Invoke get-quote tool with symbol "EURUSD".

### Tests for US2

- [ ] T029 [P] [US2] [Test] Contract test for get-quote in tests/contract/test_get_quote.go
- [ ] T030 [P] [US2] [Test] Unit test for QuoteService in tests/unit/test_quote_service.go
- [ ] T031 [US2] [Test] Integration test for get-quote end-to-end in tests/integration/test_get_quote_live.go

### Implementation for US2

- [ ] T032 [P] [US2] Extend MT5 client with GetQuote in internal/services/mt5/client.go
- [ ] T033 [P] [US2] Create QuoteService in internal/services/daemon/quote_service.go (decimal precision)
- [ ] T034 [US2] Create get-quote MCP tool in internal/services/mcp/tools/get_quote.go
- [ ] T035 [US2] Wire get-quote into MCP server in internal/services/mcp/server.go
- [ ] T036 [US2] Add input validation for get-quote (symbol, verify instrument exists)

**Checkpoint**: US1 & US2 complete. Both tools <2s independently.

---

## Phase 5: User Story 3 - Place Trading Orders (Priority: P1)

**Goal**: Execute orders with margin validation, position limits, kill switch. Real MT5 integration.  
**Independent Test**: Place BUY order, verify ticket returned <5s.

**⚠️ TDD NON-NEGOTIABLE: All tests must be written FIRST and FAIL before implementation.**

### Tests for US3 (WRITE THESE FIRST)

- [ ] T037 [P] [US3] [Test] Unit test RiskManager.CheckOrderRisk in tests/unit/test_risk_manager.go (margin, positions, kill switch)
- [ ] T038 [P] [US3] [Test] Unit test margin edge cases in tests/unit/test_risk_manager_margin.go
- [ ] T039 [P] [US3] [Test] Unit test position limits in tests/unit/test_risk_manager_positions.go
- [ ] T040 [P] [US3] [Test] Unit test kill switch in tests/unit/test_risk_manager_killswitch.go
- [ ] T041 [P] [US3] [Test] Contract test for place-order in tests/contract/test_place_order.go
- [ ] T042 [US3] [Test] Integration test place-order with margin in tests/integration/test_place_order_margin.go
- [ ] T043 [US3] [Test] Integration test idempotency (same key = cached) in tests/integration/test_place_order_idempotency.go
- [ ] T044 [US3] [Test] Integration test concurrent orders (FIFO, 10 concurrent) in tests/integration/test_concurrent_orders.go

### Implementation for US3

- [ ] T045 [P] [US3] Create RiskManager in internal/services/mt5/risk_manager.go with decimal precision
  - CheckOrderRisk(symbol, volume, side, policy) → OK/ERROR
  - Calculate required margin (MT5 formula)
  - Validate free_margin >= required
  - Check positions <= max_positions
  - Check volume <= max_volume
  - Kill switch: if loss/balance > max_drawdown, reject
- [ ] T046 [P] [US3] Extend MT5 client with PlaceOrder in internal/services/mt5/client.go (real WebAPI call)
- [ ] T047 [US3] Implement idempotency in OrderService in internal/services/daemon/order_service.go (UUID keys, 24h cache)
- [ ] T048 [US3] Create OrderService with RiskManager integration in internal/services/daemon/order_service.go
- [ ] T049 [US3] Create place-order MCP tool in internal/services/mcp/tools/place_order.go
- [ ] T050 [US3] Wire place-order into MCP server in internal/services/mcp/server.go (concurrent safe)
- [ ] T051 [US3] Add input validation (symbol, volume, side, optional UUID) in place-order
- [ ] T052 [US3] Implement FIFO sequencing in internal/services/daemon/order_sequencer.go (per account)
- [ ] T053 [US3] Add Portuguese errors ([MARGIN_INSUFFICIENT], [MAX_POSITIONS_REACHED], [KILL_SWITCH_ACTIVE], [ORDER_REJECTED])

**Checkpoint**: US1, 2, 3 complete. MVP ready — place-order validates margin, enforces limits, handles concurrency.

---

## Phase 6: User Story 4 - Manage Positions (Priority: P2)

**Goal**: Close positions, return profit/loss.  
**Independent Test**: Close open position, verify removed from MT5.

### Tests for US4

- [ ] T054 [P] [US4] [Test] Unit test PositionService.ClosePosition in tests/unit/test_position_service.go
- [ ] T055 [P] [US4] [Test] Contract test for close-position in tests/contract/test_close_position.go
- [ ] T056 [US4] [Test] Integration test for close-position in tests/integration/test_close_position_live.go

### Implementation for US4

- [ ] T057 [P] [US4] Extend MT5 client with ClosePosition in internal/services/mt5/client.go
- [ ] T058 [P] [US4] Create PositionService in internal/services/daemon/position_service.go (decimal P&L)
- [ ] T059 [US4] Create close-position MCP tool in internal/services/mcp/tools/close_position.go
- [ ] T060 [US4] Wire close-position into MCP server in internal/services/mcp/server.go
- [ ] T061 [US4] Add input validation (ticket, verify exists) in close-position
- [ ] T062 [US4] Add Portuguese errors ([POSITION_NOT_FOUND], [CLOSE_FAILED])

**Checkpoint**: US4 complete — close-position works independently.

---

## Phase 7: User Story 5 - Query Open Orders (Priority: P2)

**Goal**: Return list of pending orders.  
**Independent Test**: Query pending orders, verify matches MT5.

### Tests for US5

- [ ] T063 [P] [US5] [Test] Unit test GetPendingOrders in tests/unit/test_get_pending_orders.go
- [ ] T064 [P] [US5] [Test] Contract test for get-orders in tests/contract/test_get_orders.go
- [ ] T065 [US5] [Test] Integration test for get-orders in tests/integration/test_get_orders_live.go

### Implementation for US5

- [ ] T066 [P] [US5] Extend MT5 client with GetPendingOrders in internal/services/mt5/client.go
- [ ] T067 [P] [US5] Create GetPendingOrders in OrderService in internal/services/daemon/order_service.go
- [ ] T068 [US5] Create get-orders MCP tool in internal/services/mcp/tools/get_orders.go
- [ ] T069 [US5] Wire get-orders into MCP server in internal/services/mcp/server.go
- [ ] T070 [US5] Add latency logging for get-orders in internal/logger/logger.go

**Checkpoint**: US5 complete — get-orders returns list independently.

---

## Phase 8: User Story 6 - Get Historical Candles (Priority: P2)

**Goal**: Return OHLC data for technical analysis.  
**Independent Test**: Query 100 H1 candles, verify matches MT5 chart, <1s.

### Tests for US6

- [ ] T071 [P] [US6] [Test] Unit test CandleService.GetCandles in tests/unit/test_candle_service.go
- [ ] T072 [P] [US6] [Test] Contract test for get-candles in tests/contract/test_get_candles.go (M5/M15/H1/D/W)
- [ ] T073 [US6] [Test] Integration test for get-candles in tests/integration/test_get_candles_live.go (100 candles)
- [ ] T074 [US6] [Test] Boundary test (no history, unknown symbol) in tests/integration/test_get_candles_errors.go

### Implementation for US6

- [ ] T075 [P] [US6] Extend MT5 client with GetCandles in internal/services/mt5/client.go
- [ ] T076 [P] [US6] Create CandleService in internal/services/daemon/candle_service.go (decimal OHLC)
- [ ] T077 [US6] Create get-candles MCP tool in internal/services/mcp/tools/get_candles.go
- [ ] T078 [US6] Wire get-candles into MCP server in internal/services/mcp/server.go
- [ ] T079 [US6] Add input validation (symbol, timeframe in [M5, M15, H1, D, W], count >0 ≤1000)
- [ ] T080 [US6] Add Portuguese errors ([SYMBOL_NOT_FOUND], [INVALID_TIMEFRAME], [NO_CANDLE_DATA])

**Checkpoint**: All 6 user stories complete. All tools functional independently.

---

## Phase 9: Polish & Production Readiness

**Purpose**: Documentation, full integration, performance, security hardening.

- [ ] T081 [P] Complete README.md with architecture and usage in README.md
- [ ] T082 [P] Create DEVELOPMENT.md with local setup and debugging in docs/DEVELOPMENT.md
- [ ] T083 [P] Create OPERATIONS.md with credentials rotation, kill switch, monitoring in docs/OPERATIONS.md
- [ ] T084 Create full end-to-end integration test in tests/integration/test_e2e_full_scenario.go
- [ ] T085 Performance validation: order <5s (p95), quote <500ms, account <2s
- [ ] T086 Concurrent invocation safety: go run -race ./cmd/mcp-mt5-server
- [ ] T087 Security audit: zero hardcoded credentials, all env vars validated, TLS documented
- [ ] T088 Full test suite: go test ./... (all unit, integration, contract pass)
- [ ] T089 Code cleanup: remove dead code, no TODO/FIXME without issues
- [ ] T090 Validate quickstart.md — user can build and run in <5 minutes (document config-real-ftmo.yaml setup, MT5 path, credentials, WebAPI enablement)
- [ ] T091 [Test] Stress test: 1000+ queued orders, verify FIFO processed without duplicates
- [ ] T092 Documentation sync: update plan.md, verify all docs match code

**Checkpoint**: Production ready — 6 user stories shipped, tested, documented, secure.

---

## Dependencies & Execution

| Phase | Depends On | Can Parallel With |
|-------|-----------|-------------------|
| Phase 1 | None | N/A |
| Phase 2 | Phase 1 | N/A (blocks all stories) |
| Phase 3 (US1) | Phase 2 | US2, US3, US4, US5, US6 |
| Phase 4 (US2) | Phase 2 | US1, US3, US4, US5, US6 |
| Phase 5 (US3) | Phase 2 | US1, US2, US4, US5, US6 |
| Phase 6 (US4) | Phase 2 | US1, US2, US3, US5, US6 |
| Phase 7 (US5) | Phase 2 | US1, US2, US3, US4, US6 |
| Phase 8 (US6) | Phase 2 | US1, US2, US3, US4, US5 |
| Phase 9 | Phases 3-8 | N/A (final stage) |

### Critical Path
```
Phase 1 (Setup) → Phase 2 (Foundation) ← BLOCKER
                      ↓
        Phases 3-8 (US1-6) all in parallel
                      ↓
                Phase 9 (Polish)
```

---

## Parallel Examples

### Setup Phase (All [P] tasks simultaneous)
```
T001, T002, T003, T004, T005 → Run in parallel
```

### Foundation Phase (All [P] tasks simultaneous)
```
T006-T016 → Run in parallel
T017 (test) → Sequential after foundation
```

### User Story Phases
After Phase 2, all 6 stories run in parallel:
```
Dev A: Phase 3 (US1)
Dev B: Phase 4 (US2)
Dev C: Phase 5 (US3) ← TDD-heavy, most complex
Dev D: Phase 6 (US4)
Dev E: Phase 7 (US5)
Dev F: Phase 8 (US6)
```

---

## MVP First Strategy

**Minimum viable:** Phase 1 + Phase 2 + Phase 3 + Phase 5 (account-info + place-order)

**Timeline**: ~4-5 days  
**Scope**: Core risk management + order execution

**Next increment**: Add Phases 4, 6, 7, 8 for full feature set.

---

## Key Requirements (Non-Negotiable)

✅ **TDD**: Phase 5 tests BEFORE implementation (violation = plan rejection)  
✅ **Decimal Precision**: All currency uses `shopspring/decimal` (zero float64)  
✅ **Portuguese Errors**: All user-facing messages in pt-BR  
✅ **Idempotency**: place-order accepts UUID, prevents duplicate fills (exactly-once)  
✅ **FIFO Sequencing**: Orders processed in order per account (prevents margin race)  
✅ **Auto-Reconnect**: <10s detection, reprocess FIFO, persist to SQLite  
✅ **Concurrent Safety**: 10+ concurrent invocations, no race conditions  
✅ **Performance**: Order <5s, quote <500ms, account <2s  
✅ **Real MT5 Data**: Integration tests connect to REAL MT5 terminal via `MT5_PATH` env var (C:/Program Files/FTMO Global Markets MT5 Terminal/terminal64.exe). Unit tests only use mocks.

---

## Total Tasks: 92

**By Phase**:
- Phase 1: 5 tasks
- Phase 2: 15 tasks
- Phase 3: 8 tasks
- Phase 4: 8 tasks
- Phase 5: 20 tasks
- Phase 6: 9 tasks
- Phase 7: 8 tasks
- Phase 8: 10 tasks
- Phase 9: 12 tasks

**Ready to ship**: Phase 9 completion.
