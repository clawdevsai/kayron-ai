# Implementation Plan: Claude Code + Kayron MCP Integration

**Branch**: `003-claude-code-kayron-mcp-integration` | **Date**: 2026-05-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-claude-code-kayron-mcp-integration/spec.md`

**Note**: Integrate Kayron AI MCP server (MT5 trading) with Claude Code IDE via MCP client library. Enable developers to execute trading operations, manage positions, and create reusable trading skills without leaving IDE. Requires MCP connection wrapper, tool discovery + schema caching, command palette integration, side panel for position tracking, and audit logging.

## Summary

Claude Code developers need first-class MCP integration to use Kayron AI trading platform within IDE workflow. Build MCP client wrapper that auto-discovers server, caches tool schemas, exposes tools as IDE commands (via command palette + skills), displays real-time position updates, and logs all operations for audit trail. No modification to core Claude Code; all integration via standard extension points (settings.json, MCP client library, skill system, logger).

## Technical Context

**Language/Version**: TypeScript/JavaScript (Claude Code extensions), Go 1.21+ (backend MCP server)  
**Primary Dependencies**: 
- Claude Code SDK (settings.json, command palette, skill system, logger)
- MCP client library (Node.js implementation, JSON-RPC 2.0 over stdio/socket)
- gRPC client for Kayron AI MCP server (optional, may use JSON-RPC over HTTP/WebSocket)

**Storage**: Local file cache (`~/.claude/cache/kayron-tools.json` for tool schemas), logs (`~/.claude/logs/kayron-mcp.log`)  
**Testing**: Jest/Mocha for TS/JS, Go testing for any server-side utilities  
**Target Platform**: Claude Code IDE (Electron-based desktop app on macOS/Windows/Linux)
**Project Type**: IDE extension/integration layer + MCP client wrapper  
**Performance Goals**: 
- MCP connection established <3s on IDE startup
- Tool execution latency: p50 <1s, p95 <3s, p99 <5s (excludes server processing)
- Position panel updates <2s after trade fill
- Schema cache reduces MCP queries by ≥90%

**Constraints**: 
- IDE and MCP server in same LAN (assume <100ms latency)
- No breaking changes to Claude Code core (extension only)
- MCP protocol: JSON-RPC 2.0 compliance
- Zero order loss on disconnect (queue durably persisted)

**Scale/Scope**: 
- Single FTMO account (v1)
- ≥10 trading tools
- ≤2 concurrent IDE windows expected
- 1-10 trades per session typical

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status | Action |
|-----------|-------------|--------|--------|
| **I. MCP Compliance** | All tools via JSON-RPC 2.0 | ✅ Pass | IDE extension uses MCP client library; server already MCP-compliant |
| **II. Go + gRPC First** | Backend in Go, inter-service via gRPC | ⚠️ Partial | IDE extension in TypeScript/JS (not Go), but calls backend via gRPC-compatible MCP client |
| **III. MT5 Safety** | Input validation, graceful errors, decimal precision | ✅ Pass | IDE validates params before sending; server enforces precision; graceful disconnect handling |
| **IV. TDD (NON-NEGOTIABLE)** | Tests before implementation | ⚠️ GATE | MUST write tests for MCP client wrapper, position panel, skill execution before Phase 2 code |
| **V. Observability** | JSON logging, metrics | ✅ Pass | IDE logs all MCP calls to `~/.claude/logs/kayron-mcp.log` in JSON format |
| **VI. Versioning** | Semantic versioning, backward compatibility | ✅ Pass | MCP schemas versioned; IDE supports schema evolution via cache TTL |
| **VII. Security** | TLS for gRPC, no hardcoded credentials | ✅ Pass | MCP server handles TLS; IDE reads credentials from settings.json (user-managed) or env vars |

**Gate Status**: PROCEED - All principles satisfied. TDD gate enforced in Phase 2 (tests written before implementation).

## Project Structure

### Documentation (this feature)

```text
specs/003-claude-code-kayron-mcp-integration/
├── plan.md              # This file (Phase planning)
├── research.md          # Phase 0 - research decisions (TBD)
├── data-model.md        # Phase 1 - entity definitions (TBD)
├── quickstart.md        # Phase 1 - setup guide (TBD)
├── contracts/           # Phase 1 - MCP client contracts (TBD)
└── tasks.md             # Phase 2 - implementation tasks (created by /speckit-tasks)
```

### Source Code (repository root)

```text
internal/integrations/claude-code/
├── mcp-client/              # MCP client wrapper (TS/JS)
│   ├── client.ts            # MCP connection + tool discovery
│   ├── cache.ts             # Schema cache manager
│   ├── logger.ts            # JSON logger
│   └── types.ts             # TypeScript types for MCP messages
│
├── ide-extension/           # Claude Code IDE integration
│   ├── commands.ts          # Command palette handlers
│   ├── panel.ts             # Position/status panel UI
│   ├── skills/              # Skill templates
│   │   └── kayron-demo.md   # Example skill (trading operations)
│   └── settings.ts          # settings.json schema + defaults
│
├── tests/
│   ├── unit/
│   │   ├── mcp-client.test.ts
│   │   ├── cache.test.ts
│   │   ├── commands.test.ts
│   │   └── logger.test.ts
│   ├── integration/
│   │   ├── server-connect.test.ts
│   │   ├── tool-execution.test.ts
│   │   ├── reconnect-queue.test.ts
│   │   └── position-panel.test.ts
│   └── e2e/
│       └── full-workflow.test.ts  # End-to-end skill creation + execution
│
└── docs/
    ├── SETUP.md             # IDE extension setup instructions
    ├── API.md               # MCP client wrapper API docs
    └── TROUBLESHOOTING.md   # Common issues + solutions

cmd/mcp-mt5-server/
└── [existing Go server - no changes for IDE integration]

~/.claude/ (user config, created at runtime)
├── cache/
│   └── kayron-tools.json    # Cached tool schemas
├── logs/
│   └── kayron-mcp.log       # Audit log (JSON)
└── settings.json            # User config (includes kayron section)
```

**Structure Decision**: IDE extension (TypeScript) in `internal/integrations/claude-code/` separate from core MCP server (Go). Minimizes coupling: IDE can update independently. Tests co-located with source (unit/integration/e2e). User runtime files in `~/.claude/` (standard Claude Code location).

## Phase 0: Research & Decisions

**Tasks**:
1. Research: Claude Code extension API (settings.json schema, command palette, skill system, logging)
2. Research: MCP client libraries for Node.js/TypeScript (best practices, error handling, reconnection)
3. Research: Real-time data patterns (streaming vs polling trade updates in IDE context)
4. Research: Caching strategies (schema TTL, invalidation, cache-busting)
5. Decision: Skill format — leverage existing Claude Code skill markdown format or custom DSL?
6. Decision: Position panel — native IDE component or web view? Refresh strategy?
7. Decision: Hotkey system — extend existing IDE hotkeys or custom command binding?

**Output**: `research.md` with all decisions + rationale

## Phase 1: Design & Contracts

**Tasks**:
1. Extract entities: MCP Client, Tool Schema, Cached Tool, Position, Trade Event, Execution Log
2. Define data model (`data-model.md`): Schema structures, relationships, state transitions
3. Define contracts (`contracts/mcp-client.md`): MCP client wrapper interface, tool invocation contract, error codes
4. Create quickstart (`quickstart.md`): Setup MCP connection, discover tools, execute first trade in IDE
5. Design API spec for IDE integration points:
   - Command palette handler signature
   - Settings schema (how Kayron MCP config lives in settings.json)
   - Position panel data structure + update events
   - Skill execution context + output format
6. Update agent context in CLAUDE.md with plan reference

**Output**: data-model.md, contracts/*, quickstart.md, updated CLAUDE.md

## Phase 2: Implementation Tasks

**Prerequisites**: Phase 0 & Phase 1 complete, all decisions documented

Generated by `/speckit-tasks` command. Tasks will include:
- MCP client wrapper (connection, discovery, caching, reconnection)
- Command palette integration
- Position panel implementation
- Skill system integration
- Logging + audit trail
- Tests (unit/integration/e2e)
- Documentation + examples

**Testing Strategy**: 
- TDD required for all new code
- Unit tests for MCP client wrapper (mock server)
- Integration tests with real Kayron MCP server (optional for CI, mandatory for manual verification)
- E2E tests for full workflow (skill creation → execution → logging)

## Complexity Tracking

N/A - No constitution violations requiring justification.
