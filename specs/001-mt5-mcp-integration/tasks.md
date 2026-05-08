# Tasks: MT5 MCP Integration

**Input**: Design documents from `specs/001-mt5-mcp-integration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests written alongside implementation (TDD: B) — included in each user story phase

**Organization**: Tasks grouped by user story for independent implementation and testing

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1-US5)
- Exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, Go module, proto generation, directory structure

- [ ] T001 Create Go module `mt5-mcp` with `go.mod`
- [ ] T002 Create directory structure (`cmd/server`, `internal/mt5`, `internal/mcp`, `internal/decimal`, `api/proto`, `tests/unit`, `tests/integration`)
- [ ] T003 [P] Initialize `buf.yaml` for gRPC proto management
- [ ] T004 [P] Create `api/proto/mt5.proto` with MT5 gRPC services
- [ ] T005 [P] Add dependencies: gRPC, shops/decimal, standard library
- [ ] T006 Create `Makefile` with `build`, `test`, `lint`, `proto` targets

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure — WebAPI client, decimal utils, MCP handler base, error types. ALL user stories depend on this.

**⚠️ CRITICAL**: No user story work until this phase is complete

- [ ] T007 Create `internal/mt5/types.go` with domain types (MT5Terminal, TradingAccount, Instrument, Quote, Order, Position)
- [ ] T008 Create `internal/decimal/decimal.go` using shops/decimal
- [ ] T009 Create `internal/mt5/webapi.go` — MT5 WebAPI HTTP client with reconnect logic
- [ ] T010 [P] Create `internal/mt5/errors.go` — error codes (MT5_TERMINAL_DISCONNECTED, MT5_AUTH_FAILED, etc.)
- [ ] T011 [P] Create `internal/mt5/terminal.go` — terminal connection state + heartbeat
- [ ] T012 Create `internal/mt5/account.go` — account info queries
- [ ] T013 [P] Create `internal/mt5/quote.go` — quote fetching
- [ ] T014 [P] Create `internal/mt5/order.go` — order placement, position closing, orders listing
- [ ] T015 Create `internal/mcp/handler.go` — MCP JSON-RPC 2.0 handler over stdio
- [ ] T016 Create `internal/mcp/tools.go` — MCP tool definitions and schema registration
- [ ] T017 Create `cmd/server/main.go` — CLI entry point, environment variable loading

**Checkpoint**: Foundation ready — MCP server runs, terminal connects. User story implementation can begin.

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
