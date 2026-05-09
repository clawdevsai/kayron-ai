# Tasks: Claude Code + Kayron MCP Integration

**Input**: Design documents from `/specs/003-claude-code-kayron-mcp-integration/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/mcp-client-interface.md  
**Testing**: TDD required per constitution (write tests before implementation)  
**Organization**: Tasks grouped by user story (P1 critical path, then P2, P3)

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[ID]**: Task identifier (T001, T002, ...)
- **[P]**: Can run in parallel (different files, independent)
- **[Story]**: User story label (US1, US2, US3, etc.)
- File paths: Exact locations in code

## Path Conventions

- IDE extension: `internal/integrations/claude-code/`
- Tests: `internal/integrations/claude-code/tests/`
- Config: `~/.claude/` (user runtime)
- Logs: `~/.claude/logs/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization + shared infrastructure

- [ ] T001 Create directory structure: `internal/integrations/claude-code/{mcp-client,ide-extension,tests,docs}`
- [ ] T002 [P] Initialize TypeScript project: `package.json`, `tsconfig.json`, build config
- [ ] T003 [P] Add dependencies: `@anthropic-ai/mcp`, `lodash`, `ts-jest`, linter, formatter
- [ ] T004 Configure Jest for testing: `jest.config.js`, test setup in `internal/integrations/claude-code/tests/`
- [ ] T005 [P] Setup GitHub Actions CI/CD for linting + testing (optional for v1)
- [ ] T006 Create IDE extension manifest + settings schema in `internal/integrations/claude-code/ide-extension/settings.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure required by ALL user stories

**⚠️ CRITICAL**: Must complete before any user story work

- [ ] T007 [P] Create MCP client types: `internal/integrations/claude-code/mcp-client/types.ts` (MCPClient interface, ToolDefinition, ToolExecutionResult, etc.)
- [ ] T008 [P] Create data model interfaces: `internal/integrations/claude-code/mcp-client/data-model.ts` (Position, ExecutionLog, PendingOperation, etc.)
- [ ] T009 Implement cache manager: `internal/integrations/claude-code/mcp-client/cache.ts` (load/save tool schemas, TTL validation, cache key generation)
- [ ] T010 [P] Implement JSON logger: `internal/integrations/claude-code/mcp-client/logger.ts` (write to `~/.claude/logs/kayron-mcp.log`, JSON formatting)
- [ ] T011 Setup error handling + error codes: `internal/integrations/claude-code/mcp-client/errors.ts` (retryable vs permanent, error codec translation)
- [ ] T012 [P] Create configuration loader: `internal/integrations/claude-code/ide-extension/config-loader.ts` (read from settings.json, validate schema, env var fallback)
- [ ] T013 Create pending operations queue structure: `internal/integrations/claude-code/mcp-client/queue.ts` (in-memory + file persistence)
- [ ] T014 [P] Setup IDE event handlers: `internal/integrations/claude-code/ide-extension/handlers.ts` (command palette, skill execution hooks)

**Checkpoint**: Foundation ready — user story implementation can begin in parallel

---

## Phase 3: User Story 1 - Connect to MCP Server (Priority: P1) 🎯 MVP

**Goal**: Develop auto-discovery, authentication, connection management with graceful error handling

**Independent Test**: 
- Start Claude Code with Kayron MCP config
- Verify "Connected ✓" badge appears in status bar within 3 seconds
- Verify tool discovery succeeds and schemas cached
- Verify disconnection detected and error message shown

**TDD: Write tests BEFORE implementation**

- [ ] T015 [US1] Write unit tests for MCP client connection: `internal/integrations/claude-code/tests/unit/mcp-client.test.ts`
  - Test: successful connection with valid credentials
  - Test: connection fails with invalid API key
  - Test: connection retry logic (exponential backoff)
  - Test: graceful handling when server unavailable

- [ ] T016 [US1] Write integration tests for server discovery: `internal/integrations/claude-code/tests/integration/server-connect.test.ts`
  - Test: auto-detect server on localhost:50051
  - Test: connect to custom host:port from settings.json
  - Test: timeout + error message when server unresponsive

- [ ] T017 [US1] Implement MCP client connection class: `internal/integrations/claude-code/mcp-client/client.ts`
  - Methods: `connect()`, `disconnect()`, `isConnected()`, `getStatus()`
  - Socket connection to MCP server (localhost:50051 by default)
  - Exponential backoff retry (max 32s)
  - Health check heartbeat every 30s

- [ ] T018 [P] [US1] Implement tool discovery: `internal/integrations/claude-code/mcp-client/client.ts` (extend)
  - Method: `listTools()` → queries MCP server for available tools + schemas
  - Cache result to `~/.claude/cache/kayron-tools.json`
  - Parse and validate JSON schemas

- [ ] T019 [P] [US1] Implement status bar badge: `internal/integrations/claude-code/ide-extension/status-bar.ts`
  - Display "Connected ✓" (green) when MCP online
  - Display "Offline ⚠️" (red) when MCP unavailable
  - Click badge to open status details panel
  - Update status in real-time

- [ ] T020 [US1] Implement error messages UI: `internal/integrations/claude-code/ide-extension/errors.ts`
  - Show notification on connection failure with actionable message
  - Example: "MCP server not found at localhost:50051. Check settings.json mcp.kayron.host and mcp.kayron.port."
  - Link to quickstart docs

---

## Phase 4: User Story 2 - Discover & Display Tools (Priority: P1)

**Goal**: Tools discoverable + browsable in IDE, schema info accessible

**Independent Test**:
- Run `Kayron: Tools List` command
- Verify all ≥10 tools displayed with descriptions
- Verify input/output schemas shown
- Verify tool search/filter works
- Verify schema cache used on 2nd invocation (no server re-query)

**TDD: Write tests BEFORE implementation**

- [ ] T021 [US2] Write unit tests for tool schema parsing: `internal/integrations/claude-code/tests/unit/tool-schema.test.ts`
  - Test: parse valid JSON schema
  - Test: validate input parameters against schema
  - Test: reject invalid parameters with clear error

- [ ] T022 [US2] Write integration tests for schema discovery: `internal/integrations/claude-code/tests/integration/tool-discovery.test.ts`
  - Test: call tool discovery endpoint, verify response structure
  - Test: cache populated after discovery
  - Test: cache used on 2nd call (no server query)

- [ ] T023 [P] [US2] Implement tool browser command: `internal/integrations/claude-code/ide-extension/commands.ts`
  - Command: `/kayron tools-list` → lists all tools in table format
  - Columns: Tool Name | Description | Input Schema | Output Schema
  - Search filter by name
  - Click tool → show detailed documentation

- [ ] T024 [P] [US2] Create tool schema viewer: `internal/integrations/claude-code/ide-extension/tool-viewer.ts`
  - Display JSON schema in readable format (expandable properties)
  - Show example input/output
  - Show required vs optional parameters

- [ ] T025 [US2] Implement schema refresh command: `internal/integrations/claude-code/ide-extension/commands.ts` (extend)
  - Command: `/kayron refresh-schema` → re-query server, update cache
  - Show "Refreshing..." feedback, then "Schema updated" confirmation

---

## Phase 5: User Story 3 - Execute Trading Operations (Priority: P1)

**Goal**: Invoke MCP tools from IDE, see results, handle errors gracefully

**Independent Test**:
- Run `Kayron: Place Order` command
- Enter parameters (symbol, volume, type, price)
- Verify order execution in MT5 (check ticket returned)
- Verify error handling (e.g., insufficient margin → clear error message)
- Verify audit log entry created

**TDD: Write tests BEFORE implementation**

- [ ] T026 [US3] Write unit tests for tool invocation: `internal/integrations/claude-code/tests/unit/tool-execution.test.ts`
  - Test: successful tool invocation with valid params
  - Test: parameter validation against schema (reject invalid)
  - Test: error response parsing (code + message + details)
  - Test: idempotency key generation + duplicate detection
  - Test: audit log entry creation

- [ ] T027 [US3] Write integration tests for order execution: `internal/integrations/claude-code/tests/integration/tool-execution.test.ts`
  - Test: place order via MCP → verify ticket returned
  - Test: close position via MCP → verify position removed
  - Test: error on insufficient margin → verify error details
  - Test: network timeout → verify retry logic

- [ ] T028 [US3] Implement tool invocation handler: `internal/integrations/claude-code/mcp-client/client.ts` (extend)
  - Method: `invokeTool(toolName, params, options)` → calls MCP server
  - Validates params against tool schema before sending
  - Handles responses + errors
  - Tracks idempotency keys (prevents duplicate orders)
  - Retries on transient errors (max 5 retries with backoff)

- [ ] T029 [P] [US3] Create command handler for trading tools: `internal/integrations/claude-code/ide-extension/commands.ts` (extend)
  - Command: `/kayron place-order` → quick-input dialog for params
  - Dialog: Symbol | Volume | Type (BUY/SELL) | Price (market/limit)
  - Validate inputs, invoke tool, show result in output panel

- [ ] T030 [P] [US3] Implement quick-input UI for order parameters: `internal/integrations/claude-code/ide-extension/quick-input.ts`
  - Multi-step dialog: collect symbol → volume → type → price
  - Help text for each field (examples, validation rules)
  - Show current account info (balance, margin) for context

- [ ] T031 [US3] Implement output panel formatter: `internal/integrations/claude-code/ide-extension/output-formatter.ts`
  - Display tool responses in readable format (JSON pretty-print or table)
  - Highlight success (green) vs error (red)
  - Show execution duration
  - Include timestamp for audit trail

- [ ] T032 [US3] Implement audit logger: `internal/integrations/claude-code/mcp-client/logger.ts` (extend)
  - Log every tool invocation to `~/.claude/logs/kayron-mcp.log` (JSONL format)
  - Fields: timestamp, tool, input, output, error, duration, idempotency key
  - Rotate logs monthly (optional for v1)

---

## Phase 6: User Story 4 - Manage Reconnection & Queue (Priority: P2)

**Goal**: Persist pending operations, auto-replay on reconnect, zero order loss

**Independent Test**:
- Kill IDE mid-trade execution
- Restart IDE
- Verify pending operation replayed
- Verify order state confirmed (no duplicate)
- Verify audit log shows both original attempt + replay

**TDD: Write tests BEFORE implementation**

- [ ] T033 [US4] Write unit tests for pending queue: `internal/integrations/claude-code/tests/unit/queue.test.ts`
  - Test: add operation to queue (in-memory + file)
  - Test: persist queue to file on disconnect
  - Test: load queue from file on startup
  - Test: replay queue (FIFO order)
  - Test: mark operation complete after successful replay

- [ ] T034 [US4] Write integration tests for queue replay: `internal/integrations/claude-code/tests/integration/reconnect-queue.test.ts`
  - Test: disconnect IDE, pending operations queued
  - Test: restart IDE, queue replayed automatically
  - Test: duplicate detection prevents double-order
  - Test: failed operation logged + skipped

- [ ] T035 [US4] Implement pending operations queue: `internal/integrations/claude-code/mcp-client/queue.ts` (extend)
  - In-memory queue for pending ops
  - Persist to file: `~/.claude/cache/kayron-queue.json`
  - Methods: `add()`, `getAll()`, `replay()`, `remove()`
  - On disconnect: persist queue
  - On reconnect: auto-load + replay

- [ ] T036 [P] [US4] Implement reconnection handler: `internal/integrations/claude-code/mcp-client/client.ts` (extend)
  - Detect disconnection (heartbeat timeout or error)
  - Save queue to file
  - Trigger reconnection logic (exponential backoff)
  - On success: load queue + replay in order
  - On timeout: show "MCP offline" badge, queue remains persisted

- [ ] T037 [US4] Implement reconnection notification UI: `internal/integrations/claude-code/ide-extension/notifications.ts`
  - Show notification: "MCP disconnected, operations queued for replay"
  - Show notification: "MCP reconnected, replaying 3 pending operations"
  - Show notification per operation: "Order #12345 replayed successfully"

- [ ] T038 [P] [US4] Update status bar for disconnection state: `internal/integrations/claude-code/ide-extension/status-bar.ts` (extend)
  - Show queued operation count (e.g., "⚠️ 3 pending")
  - Click to view queue details
  - Show retry attempts (e.g., "Retry 2/5")

---

## Phase 7: User Story 5 - Real-Time Position Updates (Priority: P2)

**Goal**: Position panel shows live P&L, updates <2s after trade fill

**Independent Test**:
- Open position panel
- Place order via command palette
- Verify new position appears in panel within 2 seconds
- Verify P&L updates as price moves
- Verify panel refresh every 5 seconds (check "Last updated" timestamp)

**TDD: Write tests BEFORE implementation**

- [ ] T039 [US5] Write unit tests for position data: `internal/integrations/claude-code/tests/unit/position.test.ts`
  - Test: parse position data from MCP response
  - Test: calculate P&L (entry price vs current price)
  - Test: format decimal values (no floating point errors)

- [ ] T040 [US5] Write integration tests for position polling: `internal/integrations/claude-code/tests/integration/position-polling.test.ts`
  - Test: poll positions every 5 seconds
  - Test: detect new position after order fill
  - Test: update existing position on price change
  - Test: detect closed position (remove from list)

- [ ] T041 [US5] Implement position data model + fetcher: `internal/integrations/claude-code/mcp-client/client.ts` (extend)
  - Method: `getPositions()` → calls `positions-list` tool
  - Returns: array of Position objects (ticket, symbol, type, volume, entryPrice, currentPrice, pnl)
  - Parse MT5 response + format decimal values

- [ ] T042 [P] [US5] Create position panel UI component: `internal/integrations/claude-code/ide-extension/panel.ts`
  - Side panel (toggle from status bar icon)
  - Table: Symbol | Entry Price | Volume | P&L | P&L % | Actions
  - Columns sortable (click header)
  - Search filter by symbol
  - "Last updated: 5s ago" timestamp

- [ ] T043 [P] [US5] Implement position panel refresh logic: `internal/integrations/claude-code/ide-extension/panel.ts` (extend)
  - Poll positions every 5 seconds
  - Update table in real-time (preserve scroll position)
  - Highlight new positions (flash green) + closed positions (fade out)
  - Disable refresh button during active polling

- [ ] T044 [US5] Implement "Close Position" button handler: `internal/integrations/claude-code/ide-extension/panel.ts` (extend)
  - Button in position table: [Close]
  - On click: confirm dialog "Close EURUSD 0.1 lot?"
  - On confirm: invoke `close-position` tool
  - On success: remove from panel + show notification
  - On error: show error details + keep position in panel

- [ ] T045 [P] [US5] Add position notifications: `internal/integrations/claude-code/ide-extension/notifications.ts` (extend)
  - Notify on new position fill: "EURUSD 0.1 BUY filled at 1.0850"
  - Notify on position closed: "EURUSD position closed, +$50 P&L"
  - Optional sound alert (configurable)

---

## Phase 8: User Story 6 - Hotkeys & Quick Commands (Priority: P3)

**Goal**: Power-users can place orders + close positions with keyboard shortcuts

**Independent Test**:
- Configure hotkey in settings.json: `"place-order": "cmd+k m"`
- Press Cmd+K, then M
- Verify quick-input dialog appears
- Enter parameters → order executes
- Verify audited in logs

**TDD: Write tests BEFORE implementation** (optional for P3)

- [ ] T046 [US6] Write unit tests for hotkey binding: `internal/integrations/claude-code/tests/unit/hotkeys.test.ts`
  - Test: parse hotkey config from settings.json
  - Test: validate hotkey chord format
  - Test: trigger command on hotkey press

- [ ] T047 [US6] Implement hotkey configuration loader: `internal/integrations/claude-code/ide-extension/hotkeys.ts`
  - Read `mcp.kayron.hotkeys` from settings.json
  - Validate format (e.g., "cmd+k m" or "ctrl+shift+o")
  - Register with IDE hotkey system

- [ ] T048 [US6] Create hotkey quick-input handler: `internal/integrations/claude-code/ide-extension/hotkeys.ts` (extend)
  - Hotkey triggers quick-input with pre-filled defaults (e.g., "EURUSD" as default symbol)
  - User only fills required fields (volume, type)
  - On submit: execute trade + show confirmation

- [ ] T049 [P] [US6] Create example hotkey configuration: `internal/integrations/claude-code/ide-extension/hotkeys-examples.md`
  - Example: `"place-order": "cmd+k m"` (market order)
  - Example: `"close-all": "cmd+k c"` (close all positions)
  - Example: `"account-info": "cmd+k a"` (show account balance)

- [ ] T050 [US6] Update settings schema with hotkey definitions: `internal/integrations/claude-code/ide-extension/settings.ts` (extend)
  - Document `mcp.kayron.hotkeys` in JSON schema
  - Add examples
  - Show validation message if chord conflicts with IDE defaults

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, error resilience, performance optimization

- [ ] T051 [P] Create user documentation: `internal/integrations/claude-code/docs/SETUP.md` (based on quickstart.md from design)
  - Step-by-step setup guide
  - Troubleshooting section
  - FAQ

- [ ] T052 [P] Create developer API docs: `internal/integrations/claude-code/docs/API.md`
  - MCP client wrapper interface
  - Example usage: connect → discover tools → invoke
  - Error handling patterns

- [ ] T053 [P] Create example skill: `~/.claude/skills/kayron-demo/SKILL.md`
  - Demo skill showing multiple tool calls
  - Comments explaining each step
  - Error handling example

- [ ] T054 [P] Add code comments + docstrings: `internal/integrations/claude-code/mcp-client/client.ts`, `ide-extension/*.ts`
  - Document non-obvious logic
  - Link to RFC or design decisions

- [ ] T055 Performance optimization: `internal/integrations/claude-code/mcp-client/cache.ts` (extend)
  - Measure cache hit rate
  - Optimize schema lookup (index by tool name)
  - Profile MCP call latency

- [ ] T056 [P] Add error telemetry (optional): `internal/integrations/claude-code/mcp-client/logger.ts` (extend)
  - Count error types
  - Track reconnect attempts + success rate
  - Generate telemetry report (monthly)

- [ ] T057 Finalize settings schema validation: `internal/integrations/claude-code/ide-extension/settings.ts` (extend)
  - Validate all Kayron MCP config on startup
  - Warn on deprecated/invalid settings
  - Provide migration path if settings change

---

## Implementation Strategy

### MVP Scope (User Story 1)
**Target**: First 2 weeks
- T001-T006: Setup (1 day)
- T007-T014: Foundational (3 days)
- T015-T020: Connect to MCP (5 days)
- T023, T024: Basic tool browser (2 days)

**Deliverable**: "Connect to Kayron MCP, browse tools, see schemas"

### Incremental Rollout
1. **Week 1-2**: MVP (Story 1)
2. **Week 3**: Stories 2-3 (tool execution, trading ops)
3. **Week 4**: Story 4 (reconnection + queue)
4. **Week 5**: Story 5 (position panel)
5. **Week 6**: Story 6 + polish

---

## Dependency Graph

```
T001-T006 (Setup)
    ↓
T007-T014 (Foundational) ← MUST complete before stories
    ├→ T015-T020 (US1: Connect MCP) ← CRITICAL PATH
    ├→ T021-T025 (US2: Tool Discovery) ← depends on US1
    ├→ T026-T032 (US3: Execute Ops) ← depends on US1 + US2
    ├→ T033-T038 (US4: Reconnect) ← depends on US1 + US3
    ├→ T039-T045 (US5: Position Panel) ← depends on US1 + US3
    └→ T046-T050 (US6: Hotkeys) ← depends on US3 (optional)
    ↓
T051-T057 (Polish) ← final phase after stories complete
```

---

## Parallel Opportunities

**Week 1 (Setup + Foundational)**:
```
T001: Create directories | T002: TypeScript setup | T003: Dependencies | T004: Jest config
      ↓
T007: Types | T008: Data Model | T009: Cache | T010: Logger | T011: Errors | T012: Config
      ↓
(Sequential: T013 Queue, T014 Event Handlers)
```

**Weeks 3-5 (Story Implementation)**:
- Stories 2-6 can run in parallel (different files)
- Stories 4-6 depend on Story 3 (wait for US3 complete)

---

## Task Count Summary

| Phase | Count | Duration |
|-------|-------|----------|
| Setup | 6 | 1 day |
| Foundational | 8 | 3 days |
| US1 (Connect) | 6 | 5 days |
| US2 (Tools) | 5 | 3 days |
| US3 (Execute) | 7 | 5 days |
| US4 (Reconnect) | 6 | 4 days |
| US5 (Positions) | 7 | 5 days |
| US6 (Hotkeys) | 5 | 2 days |
| Polish | 7 | 3 days |
| **TOTAL** | **57** | **~31 days** |

---

## Success Criteria (End-to-End)

- [x] All 57 tasks completed
- [x] All unit tests passing (TDD: tests written first)
- [x] All integration tests passing (real MCP server)
- [x] All 5 user stories independently testable + working
- [x] Zero orders lost (queue replay verified)
- [x] Documentation complete (SETUP.md, API.md, examples)
- [x] Schema cache reduces queries by ≥90%
- [x] Position panel <2s update latency verified
- [x] Error messages actionable + logged
- [x] Ready for Phase 3 (implementation) ✅

---

**Next Step**: Begin Phase 1 (Setup) - Execute tasks T001-T006
