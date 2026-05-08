# Implementation Plan: MT5 MCP Integration

**Branch**: `001-mt5-mcp-integration` | **Date**: 2026-05-08 | **Spec**: `specs/001-mt5-mcp-integration/spec.md`
**Input**: Feature specification from `/specs/001-mt5-mcp-integration/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

Build MCP server exposing MetaTrader 5 trading operations via JSON-RPC 2.0 tools. Core features: account-info, market quotes, place/close orders, query pending orders. Arch: Go + gRPC daemon (local MT5 connection), auto-reconnect with SQLite-persisted pending queue, idempotent order processing (UUID key), FIFO sequencing per account. Tech: MT5 WebAPI (HTTP/JSON), shops/decimal for financial calcs, Protocol Buffers for gRPC contracts. Clarifications: Q1-Q8 resolved (gRPC daemon, auto-reconnect persistent, concurrent orders idempotent, credential mgmt, scope explicit, queue persistence, idempotency key, testing strategy).

## Clarifications (Session 2026-05-08)

- Q1: Terminal connection → **gRPC daemon local**
- Q2: Terminal disconnect → **Auto-reconnect + fila pending**
- Q3: Concurrent orders → **Independent, erro propagado**
- Q4: Credenciais → **Env vars + Secrets Manager**
- Q5: Escopo IN-SCOPE → **5 core tools + MCP + gRPC + health + pt-BR**
- Q6: Fila persistência → **SQLite, durável cross-restart, unlimited**
- Q7: Idempotência → **UUID key, FIFO sequencing, exactly-once fill**
- Q8: Testing → **MT5 real integration, mock unit, CI/CD manual**

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.21+  
**Primary Dependencies**: gRPC, Protocol Buffers, MT5 WebAPI (HTTP/JSON client), shops/decimal library  
**Storage**: N/A (stateless MCP server, no persistence)  
**Testing**: Go testing + integration tests with MT5 terminal (fixtures required)  
**Target Platform**: Windows server (MT5 terminal accessible)  
**Project Type**: MCP server / library  
**Performance Goals**: account-info <2s, quote <500ms, order placement <5s, concurrent handling ≥10  
**Constraints**: pt-BR error messages, zero hardcoded credentials, TLS in prod, decimal precision (no floats for currency), gRPC daemon auto-reconnect <10s  
**Scale/Scope**: Single MT5 terminal per MCP instance, 5 core MCP tools

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate | Status |
|-----------|------|--------|
| I. MCP Protocol Compliance | Tools MUST conform to MCP JSON-RPC 2.0 | ✅ PASS (spec FR-001) |
| II. Go + gRPC First | Services in Go, inter-service comms via gRPC + Protobuf | ✅ PASS (spec + clarification Q1) |
| III. MT5 Integration Safety | Input validation, graceful disconnect handling, decimal precision | ✅ PASS (spec FR-002/003/009, Q2 auto-reconnect) |
| IV. Test-Driven Development | Tests written before impl, Red-Green-Refactor, integration tests | ✅ PASS (spec SC-008) |
| V. Observability | Structured logging (JSON), latency metrics, MT5 health | ✅ PASS (spec FR-005) |
| VI. Versioning & Compatibility | Semantic versioning, backward compatibility | ⚠️ DEFERRED (v1.0.0, no breaking changes required yet) |

**Gate Result**: ✅ **PASS — Proceed to Phase 0**

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
├── mcp-mt5-server/
│   └── main.go              # MCP server entry point + gRPC daemon launcher

internal/
├── models/                  # Data entities (Account, Order, Position, Quote, etc.)
├── services/
│   ├── mt5/                 # MT5 WebAPI client + gRPC service implementations
│   ├── mcp/                 # MCP tool handlers (account-info, quote, place-order, etc.)
│   └── daemon/              # gRPC daemon + auto-reconnect logic
├── contracts/               # gRPC service definitions (*.proto files)
└── logger/                  # Structured JSON logging

tests/
├── integration/             # Integration tests vs MT5 terminal (fixtures)
└── unit/                    # Unit tests for services, models

go.mod / go.sum             # Go module dependencies
Dockerfile                   # Container build for Windows server
```

**Structure Decision**: Single Go project (Option 1). MCP tools handler in `internal/services/mcp/`, MT5 integration in `internal/services/mt5/`, gRPC daemon in `internal/services/daemon/`. Integration tests use real MT5 fixtures (or mock for CI).

## Phase 0: Research

**Status**: ✅ **COMPLETE**

**Completed artifacts:**
- `research.md` — MT5 WebAPI, shops/decimal, gRPC daemon lifecycle, MCP health check, pt-BR translations

---

## Phase 1: Design & Contracts

**Status**: ✅ **COMPLETE**

**Completed deliverables:**
1. ✅ `data-model.md` — Entity definitions (MT5Terminal, TradingAccount, Instrument, Quote, Order, Position) with validation rules
2. ✅ `contracts/mcp-tools.md` — gRPC service + MCP tool contracts (account-info, quote, place-order, close-position, orders-list)
3. ✅ `quickstart.md` — Setup + local dev guide

**Outputs generated**: data-model.md, contracts/mcp-tools.md, quickstart.md

---

## Phase 2: Implementation Tasks

**Status**: PENDING (generated by `/speckit-tasks`)

This phase (task breakdown + implementation sequencing) will be generated by the `/speckit-tasks` command after Phase 1 design completes. Expected deliverable: `tasks.md`
