# Data Model: Claude Code + Kayron MCP Integration

**Phase**: 1 (Design)  
**Date**: 2026-05-08  
**Purpose**: Define entities, relationships, and state transitions

## Entities

### MCP Connection

Represents authenticated session between Claude Code IDE and Kayron AI MCP server.

**Fields**:
- `id`: UUID (session ID)
- `host`: string (default: "localhost")
- `port`: number (default: 50051)
- `apiKey`: string (user credential, sourced from settings.json or env var)
- `status`: enum ("connected" | "disconnected" | "reconnecting")
- `lastConnectTime`: ISO 8601 timestamp
- `lastDisconnectTime`: ISO 8601 timestamp (null if never disconnected)
- `reconnectAttempts`: number (reset on successful connection)

**Validation**:
- `port` must be 1024-65535
- `host` must be valid hostname or IP address
- `apiKey` must not be empty or null
- `status` must be one of enum values

**Relationships**:
- Owns: pending operations queue, tool schema cache

---

### Tool Schema

Represents MCP tool metadata (input/output schema, documentation).

**Fields**:
- `id`: UUID (tool ID)
- `name`: string (e.g., "place-order", "get-quote")
- `description`: string (user-readable)
- `inputSchema`: JSON Schema (input parameters spec)
- `outputSchema`: JSON Schema (response spec)
- `version`: string (semantic version, e.g., "1.0.0")
- `createdAt`: ISO 8601 timestamp

**Validation**:
- `name` must match `/^[a-z0-9-]+$/` (lowercase, hyphens, alphanumeric)
- `inputSchema` and `outputSchema` must be valid JSON Schema
- `version` must follow semantic versioning (MAJOR.MINOR.PATCH)

**Relationships**:
- Referenced by: Tool Execution, Skill definition

---

### Tool Execution

Represents a single invocation of an MCP tool.

**Fields**:
- `id`: UUID (execution ID, request ID for MCP)
- `toolName`: string (foreign key: Tool Schema name)
- `inputParams`: object (validated against tool's inputSchema)
- `status`: enum ("pending" | "success" | "error" | "timeout")
- `result`: object | null (validated against tool's outputSchema, null if pending/error)
- `error`: object | null (error code + message, null if success)
- `createdAt`: ISO 8601 timestamp
- `startedAt`: ISO 8601 timestamp | null
- `completedAt`: ISO 8601 timestamp | null
- `durationMs`: number | null (elapsed time)
- `retryCount`: number (0-5)
- `idempotencyKey`: UUID (for duplicate prevention)

**Validation**:
- `inputParams` must conform to tool's inputSchema
- `status` must be one of enum values
- `durationMs` must be non-negative (if set)
- `retryCount` must be 0-5

**State Transitions**:
```
pending → success → [end]
pending → error → [end]
pending → timeout → error (retry) → pending
pending → [end after retry limit exceeded]
```

**Relationships**:
- References: Tool Schema (via toolName)
- References: Pending Operations Queue (if status = pending)
- References: Execution Log (audit trail)

---

### Position

Represents an open trading position in MT5.

**Fields**:
- `ticket`: number (MT5 order ticket)
- `symbol`: string (e.g., "EURUSD")
- `type`: enum ("buy" | "sell")
- `volume`: decimal string (e.g., "0.1") — stored as string for precision
- `entryPrice`: decimal string
- `currentPrice`: decimal string (updated on panel refresh)
- `pnl`: decimal string (unrealized P&L)
- `pnlPercent`: decimal string (as percentage)
- `openTime`: ISO 8601 timestamp
- `lastUpdateTime`: ISO 8601 timestamp

**Validation**:
- `ticket` must be positive integer
- `symbol` must be 6-7 character forex pair
- `type` must be "buy" or "sell"
- `volume` must be positive decimal (≥0.01)
- `entryPrice`, `currentPrice`, `pnl` must be valid decimal strings (no floating point)
- `openTime` must be valid ISO 8601

**Relationships**:
- Referenced by: Position Panel (displayed in real-time)
- Referenced by: Execution Log (audit trail of position creation/closure)

---

### Execution Log

Represents audit trail of all MCP tool invocations (for compliance, debugging).

**Fields**:
- `id`: UUID (log entry ID)
- `timestamp`: ISO 8601 timestamp
- `toolName`: string
- `inputParams`: object (stringified JSON)
- `output`: object | string (stringified JSON)
- `error`: object | string | null (stringified JSON)
- `executionDurationMs`: number
- `retryCount`: number
- `idempotencyKey`: UUID
- `userId`: string (from settings.json, for multi-user audit)

**Validation**:
- `timestamp` must be ISO 8601
- `executionDurationMs` must be non-negative
- All JSON fields must be valid JSON strings

**Format**: JSONL (one JSON object per line, appended to `~/.claude/logs/kayron-mcp.log`)

**Retention**: Kept indefinitely (user responsible for log rotation/cleanup)

**Relationships**:
- Created from: Tool Execution (denormalized for audit purposes)

---

### Skill Definition

Represents a reusable trading skill (MARKDOWN file with embedded MCP tool calls).

**Fields**:
- `id`: UUID (skill ID, derived from file path)
- `name`: string (e.g., "close-all-eurusd-positions")
- `description`: string (markdown)
- `skillPath`: string (absolute path, e.g., `~/.claude/skills/kayron-close-eurusd/SKILL.md`)
- `content`: string (MARKDOWN + YAML frontmatter)
- `toolDependencies`: [string] (list of MCP tools invoked, e.g., ["positions-list", "close-position"])
- `createdAt`: ISO 8601 timestamp
- `modifiedAt`: ISO 8601 timestamp
- `enabled`: boolean

**Validation**:
- `name` must match `/^[a-z0-9-]+$/`
- `skillPath` must start with `~/.claude/skills/` and end with `SKILL.md`
- `toolDependencies` must reference existing tools
- Skill content must be valid MARKDOWN with YAML frontmatter

**Relationships**:
- References: Tool Schema (via toolDependencies)
- Executes: Tool Execution (when skill invoked)

---

### Settings

Represents user configuration in `settings.json` (Kayron MCP section).

**Fields**:
- `mcp.kayron.enabled`: boolean (default: true)
- `mcp.kayron.host`: string (default: "localhost")
- `mcp.kayron.port`: number (default: 50051)
- `mcp.kayron.apiKey`: string (required, from env var or settings)
- `mcp.kayron.cacheTtlMinutes`: number (default: 60)
- `mcp.kayron.logLevel`: enum ("debug" | "info" | "warn" | "error", default: "info")
- `mcp.kayron.hotkeys`: object (hotkey bindings, see research.md)
- `mcp.kayron.reconnectMaxRetries`: number (default: 5)
- `mcp.kayron.reconnectBackoffMs`: number (default: 1000, exponential)

**Schema**: Documented in IDE settings schema (enforced by IDE)

**Relationships**:
- Used by: MCP Connection (during initialization)

---

## State Machines

### MCP Connection Lifecycle

```
[Disconnected]
    ↓
[Connecting] → [Connected]
    ↓              ↓
[Error]     [Connected] (heartbeat OK)
    ↓              ↓
[Retrying]   [Disconnected] (heartbeat failed or explicit disconnect)
    ↓
[Connected]
```

### Tool Execution Lifecycle (with Retry)

```
[Pending]
    ↓
[Executing]
    ├→ [Success] → [Completed]
    ├→ [Error (transient)] → [Retrying] → [Pending] (retry count < 5)
    ├→ [Error (permanent)] → [Failed] → [Completed]
    └→ [Timeout] → [Retrying] → [Pending] (retry count < 5)
```

### Position Lifecycle

```
[New (order placed)]
    ↓
[Open] → [Closing] → [Closed]
         (manual or SL/TP)
```

---

## Relationships & Constraints

- **MCP Connection** owns **Pending Operations Queue**: Each connection has associated queue (durable, persisted locally)
- **Tool Schema** ← **Tool Execution**: Execution must reference valid schema (validated on invocation)
- **Skill Definition** → **Tool Schema**: Skill's toolDependencies must exist (validated on skill load)
- **Position** ← **Tool Execution** (place-order): Creating position via tool invocation
- **Position** ← **Tool Execution** (close-position): Closing position via tool invocation
- **Execution Log** ← **Tool Execution**: Every execution logged (audit trail)

---

## Next Steps

Phase 1 continues with contract definitions + quickstart guide.
