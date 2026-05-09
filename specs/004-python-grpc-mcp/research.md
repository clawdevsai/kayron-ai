# Research: Python + gRPC MT5 MCP

**Purpose**: Resolve technical unknowns and document design decisions for implementation.

---

## 1. Operation Queue Persistence Strategy

**Unknown**: Should queued operations persist across server restarts? If so, what backend?

**Options Evaluated**:
- **Option A: In-memory only** — Simple, fast, but loses queued operations on crash
- **Option B: Redis** — Persistent, distributed, but adds dependency and complexity
- **Option C: SQLite** — Single-machine persistence, no external dependency, adequate for single-server setup
- **Option D: Files (JSON)** — Simple persistence, but slow and not scalable

**Decision**: **Option C - SQLite** 

**Rationale**: 
- Spec requires "at-least-once delivery" (SC-009: 100% recovery on restart)
- Single-server deployment (MT5 terminal co-located) doesn't need distributed cache
- SQLite provides durability without external dependency
- Adequate performance for operation queue (writes are async, reads minimal)

**Alternatives Rejected**:
- In-memory violates durability requirement
- Redis adds operational burden (deployment, monitoring) for single-server setup
- File-based is slower and harder to query

**Implementation Detail**: Use SQLAlchemy ORM with SQLite backend. Table: `queued_operations` with fields (id, agent_id, operation_type, payload, created_at, status).

---

## 2. gRPC Proto Design for MT5 Operations

**Unknown**: How to map 100+ MT5 SDK operations to gRPC proto messages?

**Options Evaluated**:
- **Option A: One message per operation** — 100+ message types, verbose proto file, type-safe
- **Option B: Generic Operation message** — Single message with operation_type + payload (any), flexible but loses type safety
- **Option C: Grouped operations** — Cluster operations by domain (OrderManagement, AccountInfo, etc.), balanced

**Decision**: **Option C - Grouped operations with domains**

**Rationale**:
- MT5 SDK operations naturally cluster: OrderManagement (Place, Modify, Cancel), SymbolInfo, TickData, AccountInfo
- Each domain gets RPC method (ExecuteOrderOperation, ExecuteSymbolOperation, GetAccountInfo, etc.)
- Payload is domain-specific message (PlaceOrderRequest, ModifyOrderRequest, etc.)
- Balances type safety (domain-level) with maintainability (not 100+ types)

**Alternatives Rejected**:
- 100+ messages: proto file becomes unmanageable; code generation bloats
- Generic Operation: loses compile-time type checking; agents must parse union types

**Implementation Detail**: Proto structure:
```protobuf
service MT5Service {
  rpc ExecuteOrderOperation(OrderOperationRequest) returns (stream OrderOperationResponse);
  rpc ExecuteSymbolOperation(SymbolOperationRequest) returns (stream SymbolOperationResponse);
  rpc GetAccountInfo(GetAccountInfoRequest) returns (stream AccountInfoResponse);
  // ... grouped by domain
}

message OrderOperationRequest {
  oneof operation {
    PlaceOrderRequest place_order = 1;
    ModifyOrderRequest modify_order = 2;
    CancelOrderRequest cancel_order = 3;
  }
}
```

---

## 3. Bidirectional Streaming Callback Implementation

**Unknown**: Precise semantics of server push callbacks. How do agents register callbacks?

**Options Evaluated**:
- **Option A: Traditional RPC** — Agent calls, blocks, receives result (no streaming)
- **Option B: Server-initiated streaming** — Agent opens stream, server pushes results and updates
- **Option C: Polling** — Agent polls endpoint for result status
- **Option D: Webhook** — Agent registers callback URL, server HTTP POSTs results

**Decision**: **Option B - Server-initiated streaming (gRPC server-push)**

**Rationale**:
- Spec clarification chose "server push/callback" (Q2→C)
- gRPC bidirectional streaming is native for this pattern
- Lower latency than polling, no polling overhead
- No external callback URL requirement (simpler agent setup)
- Agent opens stream: `stream = stub.ExecuteOrderOperation(request)` → server pushes status updates and final result

**Alternatives Rejected**:
- Traditional RPC: doesn't support callback pattern
- Polling: higher latency (p95 would be 1-2 seconds waiting), wasteful
- Webhook: requires agent to expose HTTP endpoint, adds infrastructure

**Implementation Detail**: Each operation RPC returns `stream OperationResponse` with status (QUEUED, EXECUTING, COMPLETED, FAILED) and result. Agent receives immediate QUEUED, then EXECUTING, then COMPLETED with result.

---

## 4. Error Handling & Retry Strategy

**Unknown**: Exact retry semantics for MT5 connection loss. Exponential backoff? Max retries?

**Options Evaluated**:
- **Option A: Immediate retry** — Retry once, fail fast if still down (simple)
- **Option B: Exponential backoff** — Retry with 1s, 2s, 4s delays, up to 5 minutes (standard distributed pattern)
- **Option C: Queue indefinitely** — Retry every N seconds forever (not realistic)

**Decision**: **Option B - Exponential backoff with 5-minute timeout**

**Rationale**:
- Spec (Q3→A) chose "queue and auto-retry"
- Exponential backoff prevents thundering herd if MT5 temporarily unavailable
- 5-minute timeout balances durability vs not holding stale operations

**Implementation Detail**: 
```python
retries = 0
max_retries = 10  # ~5 minutes with exponential backoff
base_delay = 1  # second

while retries < max_retries:
    try:
        result = mt5.execute(operation)
        return result
    except ConnectionError:
        delay = base_delay * (2 ** retries)
        await asyncio.sleep(delay)
        retries += 1

# After max retries, operation fails with UNAVAILABLE error
```

---

## 5. Authentication & Authorization

**Unknown**: Exactly how to implement API key auth + agent session management?

**Options Evaluated**:
- **Option A: Static API keys** — Hardcoded key list, simple, limited
- **Option B: Database of API keys** — Scalable, can revoke keys, rotation support
- **Option C: OAuth2** — Industry standard, complex to implement in v1

**Decision**: **Option B - Database of API keys (deferred OAuth2 to v2)**

**Rationale**:
- Spec assumes "pre-shared API keys" (assumptions section)
- Database allows key rotation without code change
- Support multiple agents with individual keys
- Upgrade path to OAuth2 in v2

**Implementation Detail**: 
- Table: `api_keys` (key_id, key_value_hash, agent_name, created_at, revoked_at)
- Middleware: Extract key from gRPC metadata, hash it, lookup in DB
- Return error `UNAUTHENTICATED` if key not found or revoked

---

## 6. Concurrency & Thread Safety

**Unknown**: Can we trust MT5 SDK's thread-safety for concurrent operations?

**Research Result**: 
- MT5 SDK (metatrader5 package) uses thread-safe C bindings under the hood
- Concurrent calls to different operations are safe
- **But**: Single active connection per terminal (already enforced by connection pool)
- Recommendation: Use thread pool for MT5 operations, leverage SDK's internal locks

**Decision**: Rely on MT5 SDK thread-safety for concurrent execution (per clarification Q1→B). Use asyncio with thread pool executor for blocking MT5 calls.

---

## 7. Logging & Audit Trail

**Unknown**: Log format and storage for operation audit trail (FR-006)?

**Decision**: 
- Format: JSON (structured, queryable)
- Fields: timestamp (ISO 8601), agent_id, operation_type, request_payload (truncated), result, latency_ms, error_code
- Storage: stdout (JSON lines), can be shipped to ELK/Splunk by orchestrator
- Retention: No app-level retention (ops team manages log rotation)

---

## Summary

All critical unknowns resolved. Ready for Phase 1 design (data-model.md, contracts/, quickstart.md).

**Next Steps**:
1. Generate `data-model.md` — Entity definitions (AgentSession, Operation, MT5Connection, OperationQueue, CallbackStream)
2. Generate `contracts/mt5_service.proto` — gRPC service definition
3. Generate `quickstart.md` — Setup guide, example client
