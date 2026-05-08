# MT5 MCP Evolution: Task Breakdown (10 Features)

**Status**: Ready for execution  
**Total Tasks**: 75  
**Phases**: 4 (Setup + Phase 2/3/4 per evolution-plan.md)  
**MVP Scope**: Phase 2 complete (3 features, 24 tasks)  
**Test Strategy**: Unit + integration per feature  

---

## Phase 0: Infrastructure (Setup)

Shared foundation for all 10 features. Complete before Phase 2.

- [ ] T001 Create internal/models/modify_order.go with ModifyOrder struct (price, stopLoss, takeProfit fields)
- [ ] T002 Create internal/models/market_hours.go with MarketHours struct (symbol, openTime, closeTime)
- [ ] T003 [P] Create internal/models/tick.go with Tick struct (timestamp, bid, ask)
- [ ] T004 [P] Create internal/models/equity_snapshot.go with EquitySnapshot struct (timestamp, equity, balance)
- [ ] T005 Add SQLite migration: CREATE TABLE account_equity_history in queue_schema.sql
- [ ] T006 Add SQLite migration: CREATE TABLE order_fills in queue_schema.sql
- [ ] T007 Create internal/services/cache/tick_buffer.go with circular buffer for 1000 ticks per symbol
- [ ] T008 Update api/mt5.pb.go with ModifyOrder* and MarketHours* message types
- [ ] T009 Update api/mt5_grpc.pb.go with GetCandles and 10 new methods in MT5ServiceServer interface
- [ ] T010 Build + verify compilation (all models + API changes)

---

## Phase 2: Order Management (Priority: HIGH)

**Goal**: Enable modify-order capabilities + query filters. Foundation for position management.

**MVP Criteria**:
- All 3 features callable via /rpc endpoint
- Concurrent 10 orders test passes
- Error messages in Portuguese
- Decimal precision maintained

### Feature 2.1: modify-order

- [ ] T011 [US1] Create internal/services/mt5/modify_order_service.go with GetModifyOrder() calling MT5 PATCH /order/{ticket}
- [ ] T012 [US1] Create internal/services/daemon/modify_order_service.go with ModifyOrderServiceHandler (validation + gRPC)
- [ ] T013 [US1] Create internal/services/mcp/modify_order_tool.go JSON-RPC wrapper
- [ ] T014 [US1] [P] Update cmd/mcp-mt5-server/main.go: initialize ModifyOrderService, handler, tool
- [ ] T015 [US1] [P] Register "modify-order" in tool registry + add handleModifyOrder method
- [ ] T016 [US1] Create test_modify_order.go with unit tests (validation, decimal handling)
- [ ] T017 [US1] Create integration test: test_modify_order_integration.go (real MT5 or mock)
- [ ] T018 [US1] Build + test: curl request to /rpc with modify-order params

### Feature 2.2: pending-order-details

- [ ] T019 [US2] Create internal/services/mt5/pending_order_details_service.go with GetPendingOrderDetails() (filter by symbol/status/date)
- [ ] T020 [US2] Create internal/services/daemon/pending_order_details_service.go with PendingOrderDetailsServiceHandler
- [ ] T021 [US2] Create internal/services/mcp/pending_order_details_tool.go JSON-RPC wrapper
- [ ] T022 [US2] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "pending-order-details"
- [ ] T023 [US2] [P] Add handlePendingOrderDetails method
- [ ] T024 [US2] Create test_pending_order_details.go (filter logic, empty list)
- [ ] T025 [US2] Create integration test: test_pending_order_details_integration.go
- [ ] T026 [US2] Build + test: curl with filters (symbol, status, date range)

### Feature 2.3: symbol-properties

- [ ] T027 [US3] Create internal/services/mt5/symbol_properties_service.go with GetSymbolProperties() calling MT5 GET /symbols/{symbol}
- [ ] T028 [US3] Create internal/services/daemon/symbol_properties_service.go with SymbolPropertiesServiceHandler
- [ ] T029 [US3] Create internal/services/mcp/symbol_properties_tool.go JSON-RPC wrapper
- [ ] T030 [US3] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "symbol-properties"
- [ ] T031 [US3] [P] Add handleSymbolProperties method
- [ ] T032 [US3] Create test_symbol_properties.go (parsing, decimal tick value)
- [ ] T033 [US3] Create integration test: test_symbol_properties_integration.go
- [ ] T034 [US3] Build + test: curl EURUSD, GBPUSD, verify response

### Phase 2 Integration

- [ ] T035 Concurrent load test: 10 orders modify-order + pending-order-details queries (no race conditions)
- [ ] T036 Error handling test: invalid ticket, symbol not found, malformed params
- [ ] T037 Portuguese error messages audit: all errors contain pt-BR text
- [ ] T038 Latency measurement: all 3 tools < 5 seconds (record baseline)
- [ ] T039 Commit Phase 2: all 3 tools + tests + infrastructure

---

## Phase 3: Account Analytics (Priority: MEDIUM)

**Goal**: Risk management + position tracking. Enable margin checks before orders.

**Dependencies**: Phase 2 complete  
**MVP Criteria**:
- margin-calculator callable for hypothetical trades
- position-details returns P&L per open position
- equity-history queries historical snapshots

### Feature 3.1: margin-calculator

- [ ] T040 [US4] Create internal/services/mt5/margin_calculator_service.go with CalculateMargin(symbol, volume, price) → margin%
- [ ] T041 [US4] Create internal/services/daemon/margin_calculator_service.go with MarginCalculatorServiceHandler
- [ ] T042 [US4] Create internal/services/mcp/margin_calculator_tool.go JSON-RPC wrapper
- [ ] T043 [US4] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "margin-calculator"
- [ ] T044 [US4] [P] Add handleMarginCalculator method
- [ ] T045 [US4] Create test_margin_calculator.go (decimal math, edge cases: 100% margin, 0 volume)
- [ ] T046 [US4] Create integration test: test_margin_calculator_integration.go

### Feature 3.2: position-details

- [ ] T047 [US5] Create internal/services/mt5/position_details_service.go with GetPositionDetails() querying open positions
- [ ] T048 [US5] Create internal/services/daemon/position_details_service.go with PositionDetailsServiceHandler
- [ ] T049 [US5] Create internal/services/mcp/position_details_tool.go JSON-RPC wrapper
- [ ] T050 [US5] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "position-details"
- [ ] T051 [US5] [P] Add handlePositionDetails method
- [ ] T052 [US5] Create test_position_details.go (empty positions, multiple positions, P&L calculation)
- [ ] T053 [US5] Create integration test: test_position_details_integration.go

### Feature 3.3: account-equity-history

- [ ] T054 [US6] Create internal/services/mt5/equity_history_service.go with QueryEquityHistory(from, to) querying SQLite table
- [ ] T055 [US6] [P] Implement SQLite migration runner: execute account_equity_history schema on startup
- [ ] T056 [US6] [P] Create background job: snap account equity hourly → SQLite account_equity_history table
- [ ] T057 [US6] Create internal/services/daemon/equity_history_service.go with EquityHistoryServiceHandler
- [ ] T058 [US6] Create internal/services/mcp/equity_history_tool.go JSON-RPC wrapper
- [ ] T059 [US6] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "account-equity-history"
- [ ] T060 [US6] [P] Add handleEquityHistory method + start background snapshot job
- [ ] T061 [US6] Create test_equity_history.go (SQLite queries, date range, empty history)
- [ ] T062 [US6] Create integration test: test_equity_history_integration.go

### Phase 3 Integration

- [ ] T063 Equity snapshot background job test: 5 snapshots in 5 hours, SQLite persists across restart
- [ ] T064 Margin-calculator + position-details combined: verify margin % accurate for hypothetical trade on open position
- [ ] T065 Concurrent load test: 10 simultaneous margin-calculator + equity-history queries
- [ ] T066 Latency baseline: all 3 tools <5s (margin-calc <1s, position <2s, equity-history <3s)
- [ ] T067 Commit Phase 3: all 3 features + background job + SQLite schema + tests

---

## Phase 4: Advanced Analytics (Priority: MEDIUM, Optional for MVP)

**Goal**: Performance tracking + real-time data. Enable high-frequency strategies.

**Dependencies**: Phase 3 complete  
**Note**: tick-data requires gRPC streaming (optional, MVP skips this)

### Feature 4.1: balance-drawdown

- [ ] T068 [US7] Create internal/services/mt5/balance_drawdown_service.go with CalculateDrawdown(since_timestamp) querying equity_history table
- [ ] T069 [US7] Create internal/services/daemon/balance_drawdown_service.go with BalanceDrawdownServiceHandler
- [ ] T070 [US7] Create internal/services/mcp/balance_drawdown_tool.go JSON-RPC wrapper
- [ ] T071 [US7] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "balance-drawdown"
- [ ] T072 [US7] [P] Add handleBalanceDrawdown method
- [ ] T073 [US7] Create test_balance_drawdown.go (math: max equity - current / max equity, edge cases)
- [ ] T074 [US7] Create integration test: test_balance_drawdown_integration.go

### Feature 4.2: order-fill-analysis

- [ ] T075 [US8] Create internal/services/mt5/order_fill_analysis_service.go with GetFillAnalysis(ticket) querying order_fills table
- [ ] T076 [US8] [P] Extend daemon queue handler: capture fill details (slippage, fill_time) → SQLite order_fills on success
- [ ] T077 [US8] Create internal/services/daemon/order_fill_analysis_service.go with OrderFillAnalysisServiceHandler
- [ ] T078 [US8] Create internal/services/mcp/order_fill_analysis_tool.go JSON-RPC wrapper
- [ ] T079 [US8] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "order-fill-analysis"
- [ ] T080 [US8] [P] Add handleOrderFillAnalysis method
- [ ] T081 [US8] Create test_order_fill_analysis.go (slippage calc, execution latency, missing orders)
- [ ] T082 [US8] Create integration test: test_order_fill_analysis_integration.go

### Feature 4.3: market-hours

- [ ] T083 [US9] Create internal/services/mt5/market_hours_service.go with GetMarketHours(symbol) from trading calendar
- [ ] T084 [US9] Create internal/services/daemon/market_hours_service.go with MarketHoursServiceHandler
- [ ] T085 [US9] Create internal/services/mcp/market_hours_tool.go JSON-RPC wrapper
- [ ] T086 [US9] [P] Update cmd/mcp-mt5-server/main.go: initialize + register "market-hours"
- [ ] T087 [US9] [P] Add handleMarketHours method
- [ ] T088 [US9] Create test_market_hours.go (forex, stocks, weekend closures)
- [ ] T089 [US9] Create integration test: test_market_hours_integration.go

### Feature 4.4: tick-data (Optional, requires gRPC streaming)

- [ ] T090 [US10] Extend tick_buffer.go: implement gRPC stream writer for tick updates
- [ ] T091 [US10] Create internal/services/mt5/tick_data_service.go with TickSubscription(symbol, duration)
- [ ] T092 [US10] Create internal/services/daemon/tick_data_service.go with TickDataServiceHandler (gRPC streaming)
- [ ] T093 [US10] [P] Update cmd/mcp-mt5-server/main.go: initialize + register gRPC streaming endpoint
- [ ] T094 [US10] [P] Start background tick listener: fetch MT5 ticks every 100ms → tick_buffer
- [ ] T095 [US10] Create test_tick_data.go (stream format, sampling rate, buffer overflow)
- [ ] T096 [US10] Create integration test: test_tick_data_integration.go (gRPC streaming client)

### Phase 4 Integration

- [ ] T097 Fill analysis + order execution: place order → capture fill → query fill-analysis → verify slippage
- [ ] T098 Market hours + margin calculator: prevent orders outside trading hours
- [ ] T099 Balance drawdown + equity history: verify drawdown % correct vs actual snapshots
- [ ] T100 Concurrent stress test: 10 orders + fill-analysis + drawdown + tick-data stream (Phase 4 complete)

---

## Phase 5: Polish & Cross-Cutting Concerns

**Goal**: Production readiness. Complete after all 4 features per phase.

- [ ] T101 Latency audit: record /rpc response times for all 10 tools (p50, p95, p99)
- [ ] T102 Decimal precision audit: all currency values verified as strings, no float rounding
- [ ] T103 Portuguese error messages audit: 100% coverage, no English text in errors
- [ ] T104 Concurrent stress test: 20 simultaneous tools, FIFO queue behavior verified
- [ ] T105 SQLite persistence test: crash daemon + restart, verify equity history + order fills intact
- [ ] T106 TLS production readiness: gRPC TLS config tested
- [ ] T107 Secrets management audit: no hardcoded credentials, env vars only
- [ ] T108 Documentation: update README with all 10 tool signatures + examples
- [ ] T109 Build final binary: bin/mcp-mt5-server with all 16 tools
- [ ] T110 Full integration test suite: all tools, happy path + error cases

---

## Dependency Graph

```
Phase 0: Infrastructure (T001-T010) ← foundation for all
    ↓
Phase 2: modify-order (T011-T018)
Phase 2: pending-order-details (T019-T026) ← independent, can run parallel
Phase 2: symbol-properties (T027-T034) ← independent, can run parallel
    ↓ (all complete)
Phase 2 Integration (T035-T039)
    ↓
Phase 3: margin-calculator (T040-T046) ← can start after Phase 2 integration
Phase 3: position-details (T047-T053) ← independent
Phase 3: equity-history (T054-T062) ← requires SQLite setup (Phase 0 done)
    ↓ (all complete)
Phase 3 Integration (T063-T067)
    ↓
Phase 4: balance-drawdown (T068-T074) ← requires equity history
Phase 4: order-fill-analysis (T075-T082) ← requires order_fills table
Phase 4: market-hours (T083-T089) ← independent
Phase 4: tick-data (T090-T096) ← independent (optional)
    ↓ (all complete)
Phase 4 Integration (T097-T100)
    ↓
Phase 5: Polish (T101-T110)
```

## Parallel Execution Map

**Phase 2** (24 tasks, ~6 tasks per feature):
- T011-T018 (modify-order) → 1 person
- T019-T026 (pending-order-details) → 1 person  
- T027-T034 (symbol-properties) → 1 person
- **Parallel window**: All 3 people work simultaneously T011-T034 while Phase 0 complete
- **Time**: ~2-3 days per person (1 day setup, 2 days per feature)

**Phase 3** (23 tasks, ~7 tasks per feature):
- T040-T046 (margin-calculator) → 1 person
- T047-T053 (position-details) → 1 person
- T054-T062 (equity-history) → 1 person (requires SQLite background job)
- **Parallel window**: All 3 simultaneous
- **Time**: ~3 days per person (equity-history takes 1 extra day for background job)

**Phase 4** (28 tasks, ~7 tasks per feature):
- T068-T074 (balance-drawdown) → 1 person
- T075-T082 (order-fill-analysis) → 1 person
- T083-T089 (market-hours) → 1 person
- T090-T096 (tick-data) → 1 person (optional)
- **Parallel window**: All 4 simultaneous
- **Time**: ~3 days per person (tick-data optional, adds 2 extra days)

## MVP Scope

**Start here**: Phase 0 + Phase 2 (28 tasks, ~1 week for 1 person)

```
- [ ] Phase 0: Infrastructure (10 tasks) — 1 day
- [ ] Phase 2.1: modify-order (8 tasks) — 2 days
- [ ] Phase 2.2: pending-order-details (8 tasks) — 2 days
- [ ] Phase 2.3: symbol-properties (8 tasks) — 1 day
- [ ] Phase 2 integration + tests — 1 day
Total: ~1 week for single developer
```

**MVP Success Criteria**:
- All 3 tools in bin/mcp-mt5-server
- 9 tools total (6 core from Phase 1 + 3 Phase 2)
- <5s latency per tool
- 10 concurrent orders test passes
- All errors in Portuguese
- All tests passing

**Next**: Phase 3 (1 more week) → Phase 4 (1.5 weeks) → Phase 5 (polish, 2-3 days)

---

## Format Validation Checklist

- ✅ All tasks follow: `- [ ] [TaskID] [P?] [Story?] Description with file path`
- ✅ Task IDs sequential T001-T110
- ✅ [P] marker only on parallelizable tasks (different files, no wait dependencies)
- ✅ [US1]-[US10] story labels on Phase 2-4 tasks only
- ✅ File paths exact (internal/services/mt5/*.go, etc.)
- ✅ Dependencies documented (Phase 0 → 2 → 3 → 4 → 5)
- ✅ Parallel execution examples per phase
- ✅ Test criteria per phase included
- ✅ MVP scope clearly marked
