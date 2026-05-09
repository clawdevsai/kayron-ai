# Implementation Plan: Python + gRPC MCP Migration

**Branch**: `004-python-grpc-mcp` | **Date**: 2026-05-09 | **Spec**: [spec.md](spec.md)
**Input**: Migrate MT5 MCP to Python + gRPC, enable any agent to use via bidirectional streaming callbacks

## Summary

Build gRPC service that exposes all MT5 operations to multi-language agents. Implement bidirectional streaming for operation callbacks. Queue operations when MT5 offline with auto-retry on reconnect. Support 50+ concurrent agents via connection pooling and concurrent execution (MT5 SDK thread-safety).

## Technical Context

**Language/Version**: Python 3.8+ (per assumptions)  
**Primary Dependencies**: gRPC (proto3), metatrader5 SDK, asyncio, grpcio-tools, pydantic  
**Storage**: In-memory operation queue with optional persistence (Redis/SQLite for durability NEEDS CLARIFICATION)  
**Testing**: pytest, grpcio.testing, pytest-asyncio  
**Target Platform**: Linux/Windows server (co-located with MT5 terminal)
**Project Type**: gRPC microservice  
**Performance Goals**: 100ms p95 latency (standard queries), 50+ concurrent agents, 99.5% success rate  
**Constraints**: Single active MT5 connection per terminal, bidirectional streaming, at-least-once operation delivery  
**Scale/Scope**: All MT5 SDK operations exposed, multi-language client support (Python, Go, Node.js, Java)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Status**: ✅ No constitution violations identified.

- Single service tier (gRPC server) ✓
- Standard Python + gRPC tech stack ✓
- No multi-repo/workspace expansion ✓
- Integrates with existing MT5 terminal (no new infra) ✓

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
mcp-server/
├── proto/
│   ├── mt5_service.proto          # gRPC service definition (operations, callbacks)
│   └── mt5_messages.proto         # Message types (Order, Position, Account, etc.)
├── src/
│   ├── server.py                  # gRPC server entry point
│   ├── service.py                 # MT5ServiceImpl (main request handler)
│   ├── connection_pool.py         # MT5ConnectionPool (single connection + queue)
│   ├── operation_queue.py         # OperationQueue (durability, at-least-once)
│   ├── callback_manager.py        # CallbackManager (bidirectional stream callbacks)
│   ├── models.py                  # Pydantic models (AgentSession, Operation, etc.)
│   ├── mt5_adapter.py             # MT5 SDK wrapper (thread-safe operations)
│   └── logging_config.py          # Audit logging setup
├── tests/
│   ├── unit/
│   │   ├── test_connection_pool.py
│   │   ├── test_operation_queue.py
│   │   └── test_mt5_adapter.py
│   ├── integration/
│   │   └── test_service_e2e.py
│   └── contract/
│       └── test_grpc_contract.py
├── examples/
│   ├── client_python.py           # Example Python agent
│   └── client_go.md               # Example Go agent setup
├── Dockerfile
├── requirements.txt
└── README.md
```

**Structure Decision**: Single gRPC service with clear separation: proto definitions, core service logic, MT5 adapter, operation queue, callback management, and comprehensive testing.

## Phase 0: Research (Complete)

**Output**: `research.md`

**Resolved Unknowns**:
1. ✅ Operation queue persistence → SQLite with at-least-once durability
2. ✅ gRPC proto design → Grouped operations by domain (Order, Account, Symbol, etc.)
3. ✅ Bidirectional streaming → Server-push callbacks via gRPC streaming
4. ✅ Error handling & retry → Exponential backoff (1s, 2s, 4s...) with 5-min timeout
5. ✅ Authentication → API keys in SQLite, validated via gRPC metadata middleware
6. ✅ Concurrency model → Rely on MT5 SDK thread-safety; use asyncio + thread pool
7. ✅ Logging & audit → JSON structured logs, stdout, shipped by orchestrator

---

## Phase 1: Design (Complete)

**Outputs**:
- `data-model.md` — Entities: AgentSession, QueuedOperation, MT5Connection, CallbackStream, OperationLog
- `contracts/mt5_messages.proto` — Message definitions (Order, Position, Account, Symbol, Tick types)
- `contracts/mt5_service.proto` — Service definition (ExecuteOrderOperation, GetAccountInfo, GetPositions, etc.)
- `quickstart.md` — Setup guide, example clients (Python, Go, Node.js), Docker, Kubernetes

---

## Complexity Tracking

> No constitution violations. Single service architecture, standard Python + gRPC stack, no multi-repo expansion.

---

## Phase 2: Task Breakdown (Next: `/speckit-tasks`)

**To generate detailed task list**: Run `/speckit-tasks` to decompose design into:

- **Core service** (server.py, service.py): gRPC service entry point, request routing
- **Connection pool** (connection_pool.py, mt5_adapter.py): MT5 session management, thread-safe adapter
- **Operation queue** (operation_queue.py): SQLite persistence, retry logic, durability
- **Callback manager** (callback_manager.py): Bidirectional stream management, push notifications
- **Proto codegen & models** (pydantic models, proto compilation)
- **Middleware & auth** (API key validation, gRPC metadata handlers)
- **Logging & observability** (JSON structured logs, audit trail)
- **Testing** (unit, integration, contract tests)
- **Documentation** (API docs, deployment guides)
- **Docker & K8s** (Dockerfile, K8s manifests)

---

## References

- **Spec**: `spec.md` (user requirements, acceptance criteria, success metrics)
- **Research**: `research.md` (design decisions, rationales, alternatives)
- **Data Model**: `data-model.md` (entities, relationships, state machines)
- **Contracts**: `contracts/mt5_*.proto` (gRPC service & message definitions)
- **Quickstart**: `quickstart.md` (setup, examples, deployment)

---

**Status**: ✅ Phase 0-1 complete. Ready for Phase 2 task decomposition and implementation.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
