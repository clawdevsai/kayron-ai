# Implementation Plan: MT5 MCP Integration

**Branch**: `001-mt5-mcp-integration` | **Date**: 2026-05-08 | **Spec**: [spec.md](./spec.md)

## Summary

MCP server exposing Go+gRPC tools for MetaTrader 5 trading automation. Tools: account-info, quote, place-order, close-position, orders-list. Terminal runs on Windows; MCP server on Linux. Decimal precision for all financial values.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: gRPC, Protocol Buffers (buf), MT5 API (MT5 DLL/COM via cgo or COM binding), decimal arithmetic library  
**Storage**: N/A (stateless MCP server; MT5 terminal is source of truth)  
**Testing**: `go test`, integration tests with MT5 terminal, `buf lint` for schema validation  
**Target Platform**: Linux server (MCP server) + Windows server (MT5 terminal)  
**Project Type**: MCP server / library (CLI + library embedding)  
**Performance Goals**: account-info <2s, quote <500ms, order placement <5s, 10 concurrent invocations  
**Constraints**: TLS in production, decimal for currency values, no hardcoded credentials  
**Scale/Scope**: Single MT5 terminal per MCP server instance; single trading account per terminal  

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. MCP Protocol Compliance | ✅ PASS | All tools conform to MCP JSON-RPC 2.0 (FR-001, FR-004) |
| II. Go + gRPC First | ✅ PASS | Go with gRPC + Protobuf |
| III. MT5 Integration Safety | ✅ PASS | Input validation (FR-009), disconnection handling (FR-002), decimal precision (FR-003) |
| IV. TDD (NON-NEGOTIABLE) | ✅ PASS | User chose Option B: tests written alongside implementation (still required before PR merge per Quality Gates) |
| V. Observability | ✅ PASS | Structured logging with latency metrics (FR-005) |
| VI. Versioning & Compatibility | ✅ PASS | Semantic versioning for MCP tool schemas |
| VII. Security by Default | ✅ PASS | TLS (FR-007), env credentials (FR-008), input validation (FR-009) |

**Gate Result**: ✅ All gates pass. TDD resolved (Option B).

## Project Structure

### Documentation (this feature)

```text
specs/001-mt5-mcp-integration/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
mt5-mcp/
├── cmd/
│   └── server/
│       └── main.go           # MCP server entry point
├── internal/
│   ├── mt5/
│   │   ├── terminal.go      # MT5 terminal connection
│   │   ├── account.go       # Account queries
│   │   ├── quote.go         # Market quotes
│   │   ├── order.go         # Order placement/management
│   │   └── types.go         # MT5 domain types
│   ├── mcp/
│   │   ├── tools.go         # MCP tool definitions
│   │   └── handler.go       # MCP request handler
│   └── decimal/
│       └── decimal.go       # Decimal arithmetic for financial values
├── api/
│   └── proto/
│       └── mt5.proto        # gRPC service definition
├── tests/
│   ├── unit/
│   ├── integration/
│   └── contract/
├── Makefile
└── go.mod
```

**Structure Decision**: Single Go module `mt5-mcp`. MCP server embedded as library for programmatic use. CLI wrapper for standalone execution.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None yet | — | — |

## Phase 0: Research

### Unknowns to Resolve

1. **MT5 Go integration mechanism** — MT5 DLL/COM? Third-party library? Custom cgo binding?
2. **gRPC + MCP interaction pattern** — MCP is JSON-RPC 2.0 over stdio; gRPC for internal services. Need bridge design.
3. **Decimal library choice** — shops/decimal vs. ericlagerlöf/decimal (Go standard has no decimal type)
4. **TDD confirmation** — does user want tests BEFORE implementation per constitution IV?

### Research Tasks

- Task: "MT5 API integration options for Go (DLL/COM/third-party)"
- Task: "MCP server architecture patterns (JSON-RPC 2.0 bridge to gRPC)"
- Task: "Go decimal library benchmark for financial calculations"

**Output**: research.md

## Phase 1: Design & Contracts

### Entities

From spec: MT5Terminal, TradingAccount, Instrument, Quote, Order, Position

### Interface Contracts

- MCP tools: account-info, quote, place-order, close-position, orders-list
- gRPC service: MT5Trading service with Terminal, Account, Order services
- CLI: `mt5-mcp serve --terminal <path>`

### Agent Context Update

Update AGENTS.md `<!-- SPECKIT START -->` → `<!-- SPECKIT END -->` to point to `specs/001-mt5-mcp-integration/plan.md`

**Output**: data-model.md, contracts/, quickstart.md

---

*Plan ready for Phase 0. Await user confirmation on TDD requirement.*
