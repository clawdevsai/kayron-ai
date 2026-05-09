# Feature Specification: Claude Code + Kayron AI MCP Integration

**Feature Branch**: `003-claude-code-kayron-mcp-integration`  
**Created**: 2026-05-08  
**Status**: Draft  
**Input**: User description: "O que falta para Claude Code usar esse MCP Kayron AI?"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Discover & Configure Kayron AI MCP in Claude Code (Priority: P1)

Developer opens Claude Code IDE and wants to connect to running Kayron AI MCP server. Should auto-discover via localhost or allow manual config in settings.json.

**Why this priority**: Configuration is gateway. Without setup, IDE cannot access any tools.

**Independent Test**: Can be fully tested by starting Claude Code, detecting MCP server availability, and listing available tools in IDE sidebar.

**Acceptance Scenarios**:

1. **Given** Kayron AI MCP running on localhost:50051, **When** Claude Code starts, **Then** auto-detects server, loads tool schemas, displays "Connected ✓" in status bar within 3 seconds
2. **Given** MCP server on non-standard port (e.g., 9999), **When** dev adds `"mcp.kayron.port": 9999` to settings.json, **Then** IDE reconnects and displays new tools after file save (no IDE restart needed)
3. **Given** MCP server unavailable, **When** Claude Code starts, **Then** displays "MCP Offline" badge in status bar + clear config instructions (don't error/crash)

---

### User Story 2 - Browse & Execute Trading Tools as IDE Skills (Priority: P1)

Developer wants to invoke Kayron AI tools (place-order, get-quote, close-position) via Claude Code command palette (`Cmd+Shift+P`/`Ctrl+Shift+P`) without leaving IDE.

**Why this priority**: IDE integration enables hands-on trading automation within development workflow.

**Independent Test**: Can be fully tested by opening command palette, searching "Kayron", executing a tool (e.g., "place order"), and verifying response in output panel.

**Acceptance Scenarios**:

1. **Given** IDE connected to Kayron MCP, **When** dev opens command palette + types "kayron place", **Then** autocomplete shows "Kayron: Place Order" + schema hints (inputs required)
2. **Given** "Place Order" tool selected, **When** dev enters parameters (symbol=EURUSD, volume=0.1, type=BUY), **Then** tool executes, returns ticket number in output panel within 5s
3. **Given** order fails (insufficient margin), **When** response displayed, **Then** error is highlighted + actionable message ("Need 500 USD margin, have 200 USD")

---

### User Story 3 - Create Reusable Trading Skill from CLI (Priority: P2)

Developer wants to create a custom skill that bundles repeated Kayron trading sequences (e.g., "close-all-eurusd-positions", "check-account-health") for quick access.

**Why this priority**: Custom skills reduce repetition and enable domain-specific workflows. Increases productivity.

**Independent Test**: Can be fully tested by creating skill file, loading in IDE, and invoking it successfully.

**Acceptance Scenarios**:

1. **Given** skill template exists, **When** dev creates `~/.claude/skills/kayron-close-eurusd/SKILL.md` with MCP tool calls, **Then** skill auto-registers in IDE + appears in command palette
2. **Given** skill file has syntax error, **When** IDE loads skill, **Then** error reported in IDE's skill browser (doesn't crash, points to line)
3. **Given** skill invokes multiple Kayron tools sequentially, **When** skill executed, **Then** each tool call logged with timestamp + result in IDE output panel (audit trail visible)

---

### User Story 4 - View Real-Time Position Updates in IDE (Priority: P2)

Developer wants side panel showing live position list + P&L, updating as trades execute (optional streaming, fallback to poll).

**Why this priority**: Real-time visibility reduces need to switch contexts (IDE ↔ MT5 terminal). Improves trading experience.

**Independent Test**: Can be fully tested by opening position panel, placing order, and verifying panel updates within SLA.

**Acceptance Scenarios**:

1. **Given** position panel open, **When** order fills, **Then** new position appears in panel with entry price, volume, current P&L within 2 seconds
2. **Given** 5+ positions open, **When** panel displays all, **Then** table sorted by symbol + P&L sortable (click header) + search filter works
3. **Given** panel refreshing too frequently (>1Hz), **When** user hovers over position, **Then** tooltip shows "Last updated: 2s ago" (no flicker)

---

### User Story 5 - Hotkeys & Quick Commands for Common Operations (Priority: P3)

Developer configures hotkeys (e.g., `Cmd+K M` = place market order, `Cmd+K C` = close all) for rapid trading.

**Why this priority**: Power-user feature. Speeds up experienced traders but optional for beginners.

**Independent Test**: Can be fully tested by binding hotkey, triggering it, and verifying command executes with correct parameters.

**Acceptance Scenarios**:

1. **Given** hotkey `Cmd+K M` configured for place-order, **When** dev presses hotkey, **Then** quick-input dialog appears (preset symbol "EURUSD", request volume + type only)
2. **Given** quick-input submitted, **When** tool executed, **Then** order placed + ticket shown in status bar (compact feedback)
3. **Given** hotkey triggered but MCP offline, **When** dev presses hotkey, **Then** status bar shows "MCP offline" (don't execute, clear feedback)

---

### Edge Cases

- What if user changes settings.json while IDE is running (add/remove MCP config)?
- What if Kayron MCP tool schema changes mid-session (new parameter added)?
- What if user invokes tool with params from previous session (cached values)?
- What if network latency causes tool response to arrive after user closes output panel?
- What if IDE crashes during pending tool invocation — what happens to order state?
- What if user opens multiple IDE windows with same Kayron MCP connection?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Claude Code MUST auto-discover Kayron AI MCP server on startup (check localhost:50051 by default, configurable via settings.json)
- **FR-002**: Claude Code MUST cache MCP tool schemas locally (JSON file in `~/.claude/cache/kayron-tools.json`) and validate cache freshness every session (TTL 1 hour)
- **FR-003**: Claude Code MUST expose all Kayron MCP tools as "Kayron" commands in IDE command palette (filterable, autocomplete-enabled)
- **FR-004**: Claude Code MUST display MCP connection status in IDE status bar (icon + "Connected" or "Offline" badge)
- **FR-005**: Claude Code MUST support skill files (MARKDOWN + frontmatter) that bundle MCP tool calls (via skill execution system)
- **FR-006**: Claude Code MUST log all MCP tool invocations with timestamp, tool name, input params, output, and error (if any) to `~/.claude/logs/kayron-mcp.log` (JSON format)
- **FR-007**: Claude Code MUST display tool responses in IDE output panel with syntax highlighting (JSON/table format based on output type)
- **FR-008**: Claude Code MUST support hotkey binding for common operations (configurable in `settings.json`, e.g., `"kayron.hotkeys": { "place-order": "cmd+k m" }`)
- **FR-009**: Claude Code MUST include side panel (optional toggle) showing open positions with live P&L updates (poll every 5s or subscribe if MCP supports streaming)
- **FR-010**: Claude Code MUST handle MCP reconnection transparently: detect disconnect, queue pending commands, replay on reconnect (at-least-once semantics)
- **FR-011**: Claude Code MUST provide settings schema for Kayron MCP configuration (documented in settings.json with defaults)
- **FR-012**: Claude Code MUST include example skill file (`~/.claude/skills/kayron-demo/SKILL.md`) demonstrating tool usage + best practices

### Key Entities

- **MCP Connection**: Authenticated session between Claude Code and Kayron AI MCP server (host, port, credentials)
- **Tool Schema Cache**: Local JSON file containing all available Kayron tools + input/output schemas
- **Skill**: Reusable workflow file (MARKDOWN frontmatter + commands) that invokes one or more Kayron tools
- **Output Panel**: IDE pane displaying tool responses + logs
- **Status Badge**: IDE status bar indicator showing MCP connection state
- **Position Panel**: Optional side panel showing live trade positions + P&L

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Claude Code developer can connect to Kayron AI MCP, discover tools, and execute first order within 5 minutes of reading docs (no prior MCP knowledge required)
- **SC-002**: MCP tool execution latency: p50 <1s, p95 <3s, p99 <5s (excluding network round-trip to MT5 server)
- **SC-003**: Tool schema cache reduces repeated MCP queries by ≥90% (verified via log analysis: tool discovery calls drop from 10/session to 1/session)
- **SC-004**: Zero orders lost due to IDE crash or MCP disconnect: pending queue durably persisted (verified by killing IDE mid-operation + verifying order state in MT5)
- **SC-005**: Position panel updates within ≤2 seconds of trade fill (verified by triggering order → checking panel timestamp vs order execution timestamp)
- **SC-006**: Hotkey execution completes in <500ms from key-press to status bar feedback (excludes MCP server latency)
- **SC-007**: 95% of tool invocations succeed on first attempt (excluding deliberate errors like "insufficient margin")

## Assumptions

- **Claude Code IDE**: Running latest version (supports settings.json, command palette, side panels, output panel, hotkey system)
- **Kayron MCP server**: Running + available on localhost:50051 (or user-configured host/port)
- **Network**: IDE and MCP server in same LAN (low latency, stable connectivity)
- **User permissions**: Developer has read/write access to `~/.claude/` directory for cache + logs
- **File format**: Skill files use existing Claude Code MARKDOWN + YAML frontmatter convention (not custom syntax)
- **Order execution**: IDE is advisory tool; MT5 terminal remains source-of-truth (IDE does not override/force orders)
- **Streaming API**: Optional; MCP server may not support streaming updates (fallback to polling acceptable)
- **Settings precedence**: User settings.json overrides defaults; env vars override settings.json
- **Error tolerance**: Transient network errors <3% of the time (not resilience to permanently offline server)
- **Performance baseline**: Assume Kayron MCP server responds to tool calls within 2-5 seconds (does not over-promise IDE latency)
