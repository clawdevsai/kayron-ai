# Data Model: Python + gRPC MT5 MCP

**Purpose**: Define entities, relationships, and state transitions for the implementation.

---

## Entity: AgentSession

**Represents**: Authenticated agent connection and state.

**Fields**:
- `session_id` (UUID, PK): Unique session identifier
- `agent_id` (string): Agent name/identifier (from API key)
- `api_key_hash` (string): Hash of API key for validation
- `connected_at` (ISO 8601 timestamp): When agent first connected
- `last_activity_at` (ISO 8601 timestamp): Last operation timestamp
- `stream_active` (boolean): Whether bidirectional stream is open
- `queued_operation_count` (integer): Operations pending execution
- `status` (enum): ACTIVE, DISCONNECTED, REVOKED

**Lifecycle**:
```
[New] --authenticate--> ACTIVE --last_activity--> [Timeout after 1h] --> DISCONNECTED
                          |
                          +--[gRPC disconnect]--> DISCONNECTED
                          |
                          +--[revoked key]--> REVOKED
```

**Validation Rules**:
- `session_id` must be unique
- `agent_id` cannot be empty
- `last_activity_at` must be >= `connected_at`
- Status transitions: ACTIVE → {DISCONNECTED, REVOKED} only

---

## Entity: QueuedOperation

**Represents**: Pending operation waiting for MT5 connection availability.

**Fields**:
- `operation_id` (UUID, PK): Unique operation identifier
- `session_id` (FK to AgentSession): Which agent requested
- `operation_type` (enum): PlaceOrder, ModifyOrder, CancelOrder, GetPositions, etc.
- `request_payload` (JSON): Operation parameters (compressed, encrypted if sensitive)
- `created_at` (ISO 8601): When operation was queued
- `execute_after_at` (ISO 8601, nullable): Retry-after timestamp (for backoff)
- `retry_count` (integer): Attempts so far (0-10)
- `status` (enum): QUEUED, EXECUTING, COMPLETED, FAILED, EXPIRED
- `result` (JSON, nullable): MT5 operation result or error
- `completed_at` (ISO 8601, nullable): When operation finished

**Lifecycle**:
```
[Submitted] --> QUEUED --[MT5 available]--> EXECUTING --[success]--> COMPLETED
                  |
                  +--[MT5 unavailable, retry < max] --> QUEUED (execute_after_at = now + backoff)
                  |
                  +--[retry max exceeded or terminal error] --> FAILED
                  |
                  +--[timeout > 5 min] --> EXPIRED
```

**Validation Rules**:
- `operation_id` must be unique
- `session_id` must exist in AgentSession
- `retry_count` must be 0-10
- `execute_after_at` must be in future if set
- Result JSON populated only when status = COMPLETED or FAILED
- `completed_at` required when status in {COMPLETED, FAILED, EXPIRED}

---

## Entity: MT5Connection

**Represents**: Managed connection to MT5 terminal.

**Fields**:
- `connection_id` (UUID, PK): Unique connection identifier
- `terminal_name` (string): MT5 terminal identifier (hostname:port or path)
- `is_active` (boolean): Currently connected to terminal
- `last_connected_at` (ISO 8601, nullable): Last successful connection
- `last_error` (string, nullable): Last connection error message
- `thread_id` (integer, nullable): OS thread running MT5 session
- `operation_in_progress_id` (FK to QueuedOperation, nullable): Currently executing operation

**Lifecycle**:
```
[Init] --> [INACTIVE] --[connect]--> ACTIVE --[operation]--> [in-progress] --> ACTIVE
                         |                         |
                         +--[error]--> INACTIVE <--+--[connection lost]--> INACTIVE
```

**Validation Rules**:
- `terminal_name` must be non-empty and match configured terminal
- `last_connected_at` can only be set when `is_active` = true
- `operation_in_progress_id` non-null only when operation actively executing
- Single MT5Connection per terminal (enforced at application level)

---

## Entity: CallbackStream

**Represents**: Bidirectional gRPC stream for pushing operation results to agent.

**Fields**:
- `stream_id` (UUID, PK): Unique stream identifier
- `session_id` (FK to AgentSession): Associated agent session
- `operation_id` (FK to QueuedOperation, nullable): Currently tracked operation
- `is_open` (boolean): Stream actively receiving/sending
- `opened_at` (ISO 8601): When stream was established
- `closed_at` (ISO 8601, nullable): When stream closed or timed out
- `messages_sent` (integer): Count of callback messages pushed

**Relationships**:
- One CallbackStream per AgentSession (1:1)
- CallbackStream may track multiple operations sequentially (1:N with QueuedOperation)

**Lifecycle**:
```
[Agent connects] --> OPEN --[push result]--> [wait for next op] --> [repeat]
                      |
                      +--[agent disconnect or timeout > 30s]--> CLOSED
```

**Validation Rules**:
- `stream_id` must be unique
- `session_id` must exist
- `closed_at` only set when `is_open` = false
- `messages_sent` must be >= 0

---

## Entity: OperationLog

**Represents**: Audit record of all operations for compliance and debugging.

**Fields**:
- `log_id` (UUID, PK): Unique log entry identifier
- `timestamp` (ISO 8601): When operation occurred
- `agent_id` (string): Which agent (from AgentSession)
- `operation_type` (enum): Operation executed
- `request_summary` (string): Truncated request details (PII safe)
- `result_summary` (string): Truncated result (PII safe)
- `latency_ms` (integer): Time from submit to complete
- `success` (boolean): Operation succeeded or failed
- `error_code` (string, nullable): gRPC error code if failed
- `error_message` (string, nullable): Error details (redacted if sensitive)

**Validation Rules**:
- `log_id` must be unique
- `timestamp` must be reasonable (within last 24h for current session)
- `latency_ms` must be >= 0
- `error_code` only non-null when `success` = false
- All PII (account details, order IDs if sensitive) redacted from summaries

---

## Relationships

```
AgentSession (1) ──< (N) QueuedOperation
  └─ api_key_id → APIKey.key_id

AgentSession (1) ──< (1) CallbackStream
  └─ stream_id → CallbackStream.stream_id

QueuedOperation (N) ──> (1) MT5Connection
  └─ depends on: is_active, thread_id for execution

CallbackStream (1) ──< (N) QueuedOperation (sequential)
  └─ pushes results for: operation_id

OperationLog (N) ← all entities
  └─ audit trail of: AgentSession.agent_id, QueuedOperation operations
```

---

## State Transition Rules

1. **AgentSession ACTIVE → DISCONNECTED**: Triggers cleanup of CallbackStream (close) and orphaned QueuedOperations (mark as FAILED or EXPIRED)

2. **QueuedOperation EXECUTING → COMPLETED**: Automatically pushed to agent via CallbackStream

3. **QueuedOperation retry**: Exponential backoff; retry_count incremented each attempt; `execute_after_at` set for retry delay

4. **MT5Connection reconnect**: Trigger processing of QUEUED operations from head of queue

---

## Concurrency Notes

- AgentSession: Accessed during RPC handler setup and stream management (read-heavy, low contention)
- QueuedOperation: High contention; queue processing, updates to status/result
  - Use row-level locking (SQLite WAL mode) or async queue with atomic updates
  - Batch retrieval of QUEUED operations for MT5 worker thread
  
- MT5Connection: Single instance, protected by lock; accessed only by queue processor thread (no contention)
- CallbackStream: Per-session stream object; accessed only within session's RPC handler (no contention)
