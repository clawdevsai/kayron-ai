# Backend Requirements Quality Checklist: Python + gRPC MCP

**Purpose**: Validate requirement quality for gRPC server backend (concurrency, resilience, audit, error handling)  
**Created**: 2026-05-09  
**Feature**: [spec.md](../spec.md)  
**Focus**: Concurrency & Resilience + Audit & Compliance (Reviewer gate)

---

## Concurrency Safety & Connection Pooling

- [ ] CHK001 - Is single-connection-per-terminal constraint explicitly defined and enforced in requirements? [Completeness, Spec §FR-003]
- [ ] CHK002 - Are concurrent operation execution requirements quantified (max concurrent ops, threading model)? [Clarity, Spec §FR-004]
- [ ] CHK003 - Is MT5 SDK thread-safety assumption documented and validated in requirements? [Completeness, Assumption, Spec §FR-004]
- [ ] CHK004 - Are race condition scenarios (concurrent orders same symbol) addressed in requirements? [Coverage, Spec §US-1 Scenario 3]
- [ ] CHK005 - Is connection pool reuse behavior consistent across all operation types? [Consistency, Spec §FR-003]
- [ ] CHK006 - Can "concurrent execution without degradation" (SC-002: 50+ agents) be objectively measured? [Measurability, Spec §SC-002]

## Operation Queueing & Resilience

- [ ] CHK007 - Are queue semantics (FIFO, priority, max queue size) explicitly defined? [Clarity, Gap]
- [ ] CHK008 - Is "at-least-once delivery" defined with timeout/retry semantics for queued operations? [Completeness, Spec §FR-004c]
- [ ] CHK009 - Are durability requirements specified (survival across server restart)? [Completeness, Spec §SC-009]
- [ ] CHK010 - Is MT5 reconnection behavior defined (retry backoff, max attempts, timeout)? [Clarity, Spec §FR-008]
- [ ] CHK011 - Are recovery time requirements (≤5 min recovery) measurable and traceable? [Measurability, Spec §SC-004]
- [ ] CHK012 - Is "auto-retry when reconnected" behavior consistent across all operation types? [Consistency, Spec §FR-004c]

## Bidirectional Streaming & Callbacks

- [ ] CHK013 - Are callback delivery semantics defined (timing, order, failure modes)? [Completeness, Spec §FR-004b]
- [ ] CHK014 - Is callback latency requirement (≤1 sec) measurable and testable? [Measurability, Spec §SC-008]
- [ ] CHK015 - Are scenarios for slow/non-responsive agents addressed in requirements? [Coverage, Gap]
- [ ] CHK016 - Is bidirectional stream error handling (connection drop, timeout) defined? [Completeness, Spec §FR-004b]
- [ ] CHK017 - Are callback ordering guarantees specified for operations on same agent? [Clarity, Gap]

## Audit, Logging & Traceability

- [ ] CHK018 - Are operation log contents (timestamp, agent ID, op type, result, errors) explicitly enumerated? [Completeness, Spec §FR-006]
- [ ] CHK019 - Is "100% operation capture rate" (SC-005) measurable (how verified, sampling strategy)? [Measurability, Spec §SC-005]
- [ ] CHK020 - Are silent failure prevention requirements defined with detection/alerting? [Completeness, Spec §SC-005]
- [ ] CHK021 - Is log format/structure defined (JSON, fields, encoding) for machine parsing? [Clarity, Gap]
- [ ] CHK022 - Are log retention/archival requirements specified? [Completeness, Gap]
- [ ] CHK023 - Is audit trail tamper-evidence or immutability required for compliance? [Completeness, Gap]

## Error Handling & Status Codes

- [ ] CHK024 - Are gRPC error codes (UNAUTHENTICATED, UNAVAILABLE, INVALID_ARGUMENT) mapped to specific failure modes? [Completeness, Spec §FR-009]
- [ ] CHK025 - Is error message structure (code + reason text) defined with examples? [Clarity, Spec §FR-009]
- [ ] CHK026 - Are recoverable vs. non-recoverable error scenarios defined in requirements? [Completeness, Gap]
- [ ] CHK027 - Is handling of invalid/malformed gRPC messages (acceptance criteria) explicitly defined? [Completeness, Spec §Edge Case]
- [ ] CHK028 - Are error scenarios from each user story (invalid creds, network glitch, etc.) consistently addressed? [Consistency, Spec §US-3]

## Health & Lifecycle Management

- [ ] CHK029 - Are health check endpoint requirements (readiness vs. liveness) explicitly defined? [Completeness, Spec §FR-007]
- [ ] CHK030 - Is graceful shutdown behavior (in-flight completion, timeout, new request rejection) specified? [Completeness, Spec §FR-010]
- [ ] CHK031 - Is shutdown completion SLA (≤30 sec) measurable and traceable? [Measurability, Spec §SC-007]
- [ ] CHK032 - Are in-flight operation states defined (queued, executing, completed, failed)? [Completeness, Gap]

## Agent Session & Authentication

- [ ] CHK033 - Is agent authentication scheme (API key, OAuth, etc.) defined in requirements? [Completeness, Spec §FR-002, Assumption]
- [ ] CHK034 - Is session lifecycle (creation, validation, timeout, cleanup) defined? [Completeness, Spec §FR-002]
- [ ] CHK035 - Is stale/abandoned session cleanup behavior (timeout, resource release) specified? [Completeness, Spec §Edge Case]
- [ ] CHK036 - Are multi-agent isolation requirements (session data, operation atomicity) defined? [Completeness, Gap]

## Non-Functional Requirements & Success Criteria

- [ ] CHK037 - Are latency targets (100ms p95 query, 1 sec callback) defined with measurement methodology? [Measurability, Spec §SC-001, SC-008]
- [ ] CHK038 - Is success rate target (99.5%) defined with exclusion criteria (intentional user errors)? [Clarity, Spec §SC-003]
- [ ] CHK039 - Are cross-language agent compatibility requirements (≥3 languages) measurable? [Measurability, Spec §SC-006]
- [ ] CHK040 - Can all 9 success criteria be objectively verified without ambiguity? [Measurability, Spec §Success Criteria]

## Assumptions & Dependencies

- [ ] CHK041 - Is single MT5 terminal per deployment assumption validated (no multi-terminal plans)? [Completeness, Assumption]
- [ ] CHK042 - Are Python 3.8+ version requirements tied to specific feature dependencies? [Completeness, Assumption]
- [ ] CHK043 - Are external dependencies (logging infrastructure, gRPC libraries) documented in requirements? [Completeness, Assumption]
- [ ] CHK044 - Is network stability assumption (transient failures handled via gRPC retries) quantified? [Clarity, Assumption]

## Ambiguities & Conflicts

- [ ] CHK045 - Does "at-least-once delivery" conflict with "exactly-once operation semantics"? [Conflict, Spec §FR-004c]
- [ ] CHK046 - Is idempotency requirement defined for operations that auto-retry? [Completeness, Gap]
- [ ] CHK047 - Do concurrent order requirements (FR-004, Scenario 3) conflict with single-connection constraint? [Consistency, Spec §FR-003, FR-004]

---

## Notes

- All items test requirement QUALITY, not implementation
- Focus: Concurrency/resilience (P1) + audit/compliance + error handling
- Traceability: ≥80% items reference spec sections or mark gaps
- Next: Append items for proto contracts, API design, or testing strategy if needed
