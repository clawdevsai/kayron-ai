# Tasks: Python + gRPC MT5 MCP Implementation

**Feature**: Python + gRPC MCP Migration  
**Branch**: `004-python-grpc-mcp`  
**Status**: Ready for implementation  
**Task Count**: 24 tasks across 5 phases

---

## Overview

Implementation organized by user story priority (P1 → P2). Each phase is independently testable and can be demoed to stakeholders.

**MVP Scope** (recommended): Complete Phase 1 + Phase 3 (US1) for core multi-agent MT5 access via gRPC.

---

## Execution Model

### Parallelization Opportunities

- **Phase 1 (Setup)**: Tasks T001-T003 parallelizable (independent file creation)
- **Phase 2 (Foundational)**: Tasks T004-T008 depend on Phase 1; T004-T005 parallelizable (separate modules)
- **Phase 3 (US1)**: Tasks T009-T016 mostly parallelizable after T008 (protobuf codegen)
  - T009-T010 (proto + models) → feed T011-T016 (service, pool, queue, callbacks)
  - All service modules (T011-T014) can develop in parallel
- **Phase 4 (US2)**: Tasks T017-T019 parallelizable (independent middleware/auth)
- **Phase 5 (US3)**: Tasks T020-T024 sequential (integration, testing, deployment)

### Independent Test Criteria

- **US1 Complete**: gRPC service accepts requests, routes to MT5, returns results via bidirectional stream
- **US2 Complete**: Only authenticated agents (API key) can access; unauthenticated requests rejected
- **US3 Complete**: Operations queue persists to SQLite; server restart recovers queued operations

---

## Phase 1: Setup

Project structure, dependencies, proto codegen setup.

- [x] T001 Create project directory structure (mcp-server/{proto,src,tests,examples,contracts})
- [x] T002 [P] Initialize requirements.txt with gRPC, metatrader5, asyncio, pydantic, sqlalchemy deps
- [x] T003 [P] Create setup.py and .gitignore for Python gRPC project

---

## Phase 2: Foundational (Blocking Prerequisites)

Proto definitions, core models, database schema, dependency injection.

- [x] T004 Generate proto Python code: Run protoc on mt5_messages.proto + mt5_service.proto → mcp-server/src/pb/
- [x] T005 [P] Create pydantic models (src/models.py): AgentSession, QueuedOperation, MT5Connection, CallbackStream, OperationLog per data-model.md
- [x] T006 [P] Initialize SQLite schema (src/db_schema.py): tables for api_keys, queued_operations, operation_logs
- [x] T007 [P] Create configuration loader (src/config.py): Load config.yaml, validate MT5 settings, API keys
- [x] T008 Create logging config (src/logging_config.py): JSON structured logs per research.md decision

---

## Phase 3: User Story 1 - Multi-Agent MT5 Access via gRPC (P1)

Enable trading agents to access MT5 operations via gRPC with bidirectional streaming callbacks.

**Story Goal**: Deploy gRPC server that exposes all MT5 operations. Agents connect with API key, send operation request, receive QUEUED/EXECUTING/COMPLETED status updates via stream.

**Independent Test**: Launch server. Connect Python agent. Place order. Verify status callbacks received. Order executes on MT5.

### Foundational Tasks (for this story)

- [x] T009 Implement MT5 SDK adapter (src/mt5_adapter.py): Thread-safe wrapper around metatrader5 package per research.md concurrency decision
- [x] T010 [P] Implement connection pool (src/connection_pool.py): Manage single MT5 connection, queue operations, enforce single-connection constraint

### Service Implementation (parallelizable)

- [x] T011 [P] [US1] Implement operation queue (src/operation_queue.py): SQLite persistence, retry logic (exponential backoff per research.md), durability
- [x] T012 [P] [US1] Implement callback manager (src/callback_manager.py): Track bidirectional streams, push operation status updates
- [x] T013 [P] [US1] Implement gRPC service (src/service.py MT5ServiceImpl): Route all RPC calls (ExecuteOrderOperation, GetAccountInfo, etc.) to MT5 adapter + queue
- [x] T014 [P] [US1] Implement gRPC server (src/server.py): Initialize service, bind to port 50051, start async loop
- [x] T015 [US1] Create example Python client (examples/client_python.py): Demonstrate PlaceOrder + GetAccountInfo with streaming callbacks
- [x] T016 [US1] Create example Go + Node.js clients (examples/client_go.md, examples/client_node.md): Setup instructions

---

## Phase 4: User Story 2 - Session & Connection Management (P1)

Implement queuing, connection pooling, operation durability per clarification Q3→A (queue + auto-retry on reconnect).

**Story Goal**: When MT5 is offline, operations queue in SQLite. When MT5 reconnects, queued operations auto-execute with exponential backoff retry.

**Independent Test**: Kill MT5 process. Submit operation (verify QUEUED). Restart MT5. Verify operation auto-executes and completes.

### Service Implementation (parallelizable)

- [x] T017 [P] [US2] Implement API key authentication middleware (src/middleware/auth.py): Validate api-key header, reject unauthenticated requests with gRPC UNAUTHENTICATED
- [x] T018 [P] [US2] Implement agent session management (src/session_manager.py): Track active sessions, timeout cleanup, per-session operation queues
- [x] T019 [US2] Implement operation recovery on server restart (src/recovery.py): Load QUEUED operations from SQLite on startup, begin processing

---

## Phase 5: User Story 3 - Error Resilience & Logging (P2)

Comprehensive error handling, structured logging, observability per research.md decisions.

**Story Goal**: All operations logged in JSON format with timestamp, agent_id, operation_type, result, latency. Errors surfaced with clear codes (UNAUTHENTICATED, UNAVAILABLE, INVALID_ARGUMENT).

**Independent Test**: Submit invalid request. Verify error code returned. Check logs for entry. Submit valid order. Verify latency_ms recorded.

### Implementation

- [x] T020 Implement structured error handling (src/errors.py): Define custom exceptions, map to gRPC error codes per FR-009
- [x] T021 [P] [US3] Implement audit logger (src/audit_logger.py): Log all operations (timestamp, agent_id, operation_type, request_summary, result_summary, latency_ms, success, error_code)
- [x] T022 [US3] Create health check endpoint (src/health_check.py): Implement CheckHealth RPC, return SERVING/NOT_SERVING + MT5 status message
- [x] T023 [US3] Write integration tests (tests/integration/test_service_e2e.py): End-to-end flow (auth → queue → execute → complete) for all 7 RPCs
- [x] T024 [US3] Create contract tests (tests/contract/test_grpc_contract.py): Validate proto messages match expected schema, all RPCs callable

---

## Phase 6: Polish & Deployment

Docker, Kubernetes, documentation, performance tuning.

- [x] T025 [P] Create Dockerfile (Dockerfile): Multi-stage build, minimal runtime, expose port 50051
- [x] T026 [P] Create Kubernetes manifests (k8s/deployment.yaml, k8s/service.yaml): Deployment, Service, ConfigMap, PersistentVolumeClaim
- [x] T027 [P] Create deployment guide (docs/DEPLOYMENT.md): Docker, K8s, config management, troubleshooting
- [x] T028 Create performance benchmarks (tests/benchmarks/latency.py): Measure p95 latency for standard operations, validate SC-001
- [x] T029 Create stress test (tests/stress/concurrent_agents.py): Simulate 50+ concurrent agents, validate SC-002 (no degradation)

---

## Dependencies & Story Sequencing

```
Phase 1 (Setup)
  ↓
Phase 2 (Foundational: Proto + Models + DB + Config)
  ↓
Phase 3 (US1: Multi-Agent Access)  ←─────────┐
  ↓                                           │
Phase 4 (US2: Queuing + Reconnect) ←─── Depends on US1
  ↓
Phase 5 (US3: Observability)  ←─────── Depends on US1 + US2
  ↓
Phase 6 (Deployment)
```

**Parallel Opportunities**:
- All Phase 1 tasks (T001-T003)
- Phase 2 tasks T004-T008 (after Phase 1 completes)
- Phase 3 tasks T011-T016 (after T009-T010 complete)
- Phase 4 tasks T017-T019 (after Phase 3 complete; can start while Phase 3 still in progress)
- Phase 5 tasks T020-T024 (after Phase 4; T023-T024 tests can run parallel with Phase 5 implementation)
- Phase 6 tasks T025-T029 (final polish, can parallelize after Phase 5)

---

## Implementation Strategy

### MVP (Minimum Viable Product)
**Target**: Phase 1 + Phase 2 + Phase 3 (15 tasks, ~5 days of dev)
- **Deliverable**: gRPC server with multi-agent MT5 access, bidirectional streaming
- **Demo**: Connect Python agent, place order, see status updates via stream
- **Test**: Manual testing with real MT5, happy path only

### V1 (Production Ready)
**Target**: Phases 1-5 (24 tasks, ~10-12 days of dev)
- **Deliverable**: Full feature with queuing, error handling, observability
- **Demo**: Multi-agent access, MT5 offline recovery, operation audit trail
- **Test**: Integration tests for all 7 RPCs, stress tests for 50 agents, contract tests

### V1.1 (Operations Ready)
**Target**: Phase 6 (5 tasks, ~2 days)
- **Deliverable**: Docker image, K8s manifests, performance benchmarks
- **Demo**: Deploy via Docker or K8s, performance metrics
- **Test**: Performance benchmarks (p95 latency), stress test under load

---

## Task Tracking

Track completion by phase:

- [x] **Phase 1**: T001-T003 (3/3) 100% complete
- [x] **Phase 2**: T004-T008 (5/5) 100% complete  
- [x] **Phase 3**: T009-T016 (8/8) 100% complete
- [x] **Phase 4**: T017-T019 (3/3) 100% complete
- [x] **Phase 5**: T020-T024 (5/5) 100% complete
- [x] **Phase 6**: T025-T029 (5/5) 100% complete

**Total**: 29 tasks — **✓ ALL COMPLETE**

---

## Notes

- **Code Review**: Each phase completes with CR before proceeding to next phase
- **Testing**: Minimum integration test per user story (T023 covers all 7 RPCs for US1-US3)
- **Documentation**: Quickstart.md already provides setup + examples; deployments guide added in Phase 6
- **Performance**: Benchmarks in Phase 6 validate success criteria (SC-001, SC-002)
