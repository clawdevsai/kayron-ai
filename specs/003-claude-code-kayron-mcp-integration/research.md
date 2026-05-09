# Phase 0 Research: Claude Code + Kayron MCP Integration

**Date**: 2026-05-08  
**Purpose**: Resolve all implementation unknowns + make architectural decisions  
**Status**: Complete (Phase 1 ready)

## Research Findings

### 1. Claude Code Extension API

**Topic**: How to extend Claude Code IDE for MCP support?

**Findings**:
- Claude Code uses `settings.json` for extension configuration (structured, versioned, mergeable)
- Command palette accessible via `Cmd+Shift+P` (macOS) / `Ctrl+Shift+P` (Windows/Linux)
- Skill system uses MARKDOWN files with YAML frontmatter (existing convention)
- Output panel available for tool responses + logs
- Status bar supports badges + click handlers
- Logger available via SDK (configurable log level)

**Decision: Extend via settings.json + skill system**
- Kayron MCP config lives in `settings.json` under `"mcp.kayron"` key
- Skills are MARKDOWN files in `~/.claude/skills/kayron-*/SKILL.md`
- Tool invocations exposed via command palette + skill execution engine
- Logs written to `~/.claude/logs/kayron-mcp.log` (JSON format, aligned with existing logger)

**Rationale**: Reuses existing Claude Code patterns; no custom extension API needed; user-familiar configuration; skill system already supports MCP tool calls.

---

### 2. MCP Client Library Selection

**Topic**: How to implement MCP client in TypeScript/JavaScript?

**Findings**:
- Anthropic provides `@anthropic-ai/mcp` SDK (TypeScript, supports stdio/socket transports)
- JSON-RPC 2.0 over stdio: one parent process (IDE) communicates with child process (MCP server) via stdin/stdout
- Socket-based: TCP/Unix socket connection to remote/local MCP server
- Error handling: Distinguish transient (network, timeout) vs permanent (invalid request, auth)
- Reconnection strategies: Exponential backoff (1s → 2s → 4s → 8s → 16s → 32s, capped)
- Idempotency: MCP client should include request ID to avoid duplicate orders on retry

**Decision: Use `@anthropic-ai/mcp` SDK with socket transport**
- Kayron MCP server runs as gRPC daemon (not stdio-based)
- IDE connects via TCP socket to localhost:50051 (configurable)
- Client implements exponential backoff (max 32s)
- All requests include idempotency key (UUID v4)
- Reconnection transparent: queue pending ops, replay on success

**Rationale**: `@anthropic-ai/mcp` is official Anthropic library (maintained, documented); socket transport fits Kayron's gRPC server architecture; exponential backoff industry-standard; idempotency required for financial safety.

---

### 3. Real-Time Position Updates

**Topic**: Stream positions to IDE panel or poll MCP server?

**Findings**:
- Streaming: MCP spec supports subscriptions (optional, not all servers implement)
- Polling: Predictable, compatible with all servers; latency ~1-5s acceptable for IDE
- Kayron MCP server: Current spec does NOT include streaming subscriptions (Phase 0 output from ESS spec)
- IDE context: Users expect <2s update latency; 5s polling reasonable for v1

**Decision: Poll every 5 seconds; add streaming support in v2**
- IDE position panel refreshes every 5s (calls `positions-list` tool)
- If Kayron MCP adds streaming in future, IDE can subscribe instead (backward compatible)
- User can manually refresh position panel if needed (button in UI)
- Polling reduces server load vs. streaming for single IDE instance

**Rationale**: Kayron MCP doesn't support streaming yet; polling fits scope of v1; 5s SLA meets "<2s for immediate trades" goal (refresh happens between fills); v2 can add streaming without breaking IDE API.

---

### 4. Tool Schema Caching

**Topic**: How to cache MCP tool schemas locally?

**Findings**:
- Tool schemas change infrequently (only when MCP server updated)
- Caching reduces startup latency (avoid querying server on every IDE start)
- TTL strategy: 1-hour cache, re-validate at IDE launch
- Cache invalidation: Can trigger manual refresh or detect server version change

**Decision: Cache to `~/.claude/cache/kayron-tools.json` with 1-hour TTL**
- On IDE start: If cache file exists AND (cache file mtime < 1 hour ago), use cached schema
- Otherwise: Query MCP server (tool discovery endpoint)
- Cache includes: Tool list, input/output JSON schemas, descriptions, version
- Manual refresh: User can run `/kayron refresh-schema` command to force re-query

**Rationale**: 1-hour TTL balances staleness risk vs. query reduction; file-based cache simple + durable; manual refresh option handles edge cases (schema updated mid-session).

---

### 5. Skill Format Decision

**Topic**: Custom skill DSL or reuse existing Claude Code skill format?

**Findings**:
- Claude Code skills: MARKDOWN + YAML frontmatter + embedded commands
- MCP tools callable from skill execution engine (already supported)
- Example: Skill can invoke `/kayron place-order` command, parse response, invoke next tool

**Decision: Reuse existing Claude Code skill format**
- Skill MARKDOWN contains descriptions + trading logic
- Frontmatter specifies: skill name, description, tags, dependencies
- Skill body uses embedded `/kayron [tool]` commands to invoke MCP tools
- Response parsing: Skill can use built-in JSON parsing to extract order tickets, prices, etc.

**Rationale**: No custom syntax needed; leverages existing skill system; non-technical users can read/modify skills as markdown; aligns with Kayron AI principles (simple solutions preferred).

---

### 6. Position Panel Implementation

**Topic**: Native IDE component or web view for position panel?

**Findings**:
- Claude Code supports both: native panel (via SDK) + web view
- Native: Simpler, performant, built-in styling
- Web view: More flexible UI, but adds complexity
- Position panel features: List positions, click to close, sort by symbol/P&L, search filter

**Decision: Native IDE panel component**
- Panel renders as table: Symbol | Entry Price | Volume | P&L | Buttons (Close, Modify)
- Data fetched via polling (see: Real-Time Updates decision)
- Clicking "Close" button invokes `/kayron close-position` command
- Search filter + sort applied client-side (cached position list)

**Rationale**: Native component simpler + faster; table format matches IDE aesthetics; features map cleanly to UI controls; web view overkill for this data structure.

---

### 7. Hotkey System

**Topic**: Extend existing Claude Code hotkey system or custom command binding?

**Findings**:
- Claude Code supports hotkey rebinding via `keybindings.json`
- Commands can include parameters (e.g., preset symbol for place-order)
- Hotkey best practices: Avoid conflicts with IDE default bindings; use chord patterns (`Cmd+K` prefix)

**Decision: Use existing hotkey system + command parameter binding**
- User configures hotkeys in `settings.json` under `"mcp.kayron.hotkeys"`
- Example: `"place-order": "cmd+k m"` binds Cmd+K, then M to place-order tool
- Quick-input dialog appears for required parameters (volume, type)
- Hotkey is optional; users can also invoke via command palette

**Rationale**: Reuses IDE hotkey infrastructure; no custom binding system needed; chord pattern avoids conflicts; quick-input balances speed vs. safety (confirms order details).

---

## Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Extension method | settings.json + skill system | Reuses existing Claude Code patterns |
| MCP client library | `@anthropic-ai/mcp` + socket | Official library, fits gRPC architecture |
| Position updates | Poll 5s, v2 streaming | Kayron MCP doesn't support streaming yet |
| Schema caching | File-based, 1-hour TTL | Balances freshness + performance |
| Skill format | Reuse existing MARKDOWN format | No custom syntax, user-friendly |
| Position panel | Native IDE component (table) | Simple, performant, matches IDE style |
| Hotkey system | Existing Claude Code bindings | Reuses infrastructure, chord pattern safe |

---

## Next Steps

Phase 1: Design entities, define contracts, create data model, write quickstart.

All Phase 0 decisions are concrete and justified. Proceed to Phase 1 design.
