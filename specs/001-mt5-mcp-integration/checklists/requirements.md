# Checklist: MT5 MCP Integration Requirements Quality

**Purpose**: Validate specification completeness and quality for MT5 MCP Integration feature
**Created**: 2026-05-08
**Feature**: [spec.md](../spec.md)
**Focus**: Full integration test coverage, data consistency edge cases, order rejection scenarios, formal performance requirements

---

## Requirement Completeness

- [ ] CHK001 - Are all 5 MCP tools (account-info, quote, place-order, close-position, orders-list) specified with complete input/output schemas? [Completeness, Spec §User Scenarios]
- [ ] CHK002 - Are order type enums (buy, sell, buy_limit, sell_limit, buy_stop, sell_stop) explicitly defined in requirements? [Completeness, Spec §User Story 3]
- [ ] CHK003 - Are volume validation rules (minimum lot increment, maximum lot) specified for order placement? [Gap]
- [ ] CHK004 - Is the price gapping detection threshold specified for order rejection? [Gap, Spec §Edge Cases]
- [ ] CHK005 - Are duplicate ticket detection and resolution rules fully specified? [Completeness, Spec §Edge Cases]
- [ ] CHK006 - Are the exact timeout thresholds for each tool operation documented? [Gap]
- [ ] CHK007 - Is reconnect behavior (number of retries, backoff strategy) specified for terminal disconnection? [Gap, Spec §Edge Cases]
- [ ] CHK008 - Are health check endpoint requirements fully specified including what "healthy" means? [Completeness, Spec §FR-010]
- [ ] CHK009 - Is TLS configuration for gRPC specified (TLS version, cipher suites)? [Completeness, Spec §FR-007]
- [ ] CHK010 - Are re-login detection and authentication error recovery fully specified? [Completeness, Spec §Edge Cases]

---

## Requirement Clarity

- [ ] CHK011 - Is "sufficient margin" quantified with specific calculation formula? [Clarity, Spec §User Story 3]
- [ ] CHK012 - Is "market open" vs "market closed" defined with specific trading session hours? [Ambiguity, Spec §User Story 3]
- [ ] CHK013 - Are decimal precision requirements for financial values explicitly stated (number of decimal places)? [Clarity, Spec §FR-003]
- [ ] CHK014 - Is "clear disconnection message" defined with specific required content fields? [Clarity, Spec §FR-002]
- [ ] CHK015 - Is the concurrent invocation limit (10 per SC-004) technically justified? [Clarity, Spec §SC-004]
- [ ] CHK016 - Are error code names (MT5_TERMINAL_DISCONNECTED, etc.) documented with exact string values? [Clarity, Spec §Error Model]
- [ ] CHK017 - Is "normal conditions" for SC-001 (2-second account info) defined with specific load/criteria? [Ambiguity, Spec §SC-001]
- [ ] CHK018 - Is "stale quote" threshold defined with specific time delta? [Gap, Spec §Edge Cases]

---

## Requirement Consistency

- [ ] CHK019 - Are error response format requirements consistent across all tools (code + message + data)? [Consistency, Spec §FR-004]
- [ ] CHK020 - Do SC-006 (10s disconnection detection) and timeout error for order placement (5s per SC-003) align without conflict? [Consistency, Spec §SC-003/SC-006]
- [ ] CHK021 - Is Portuguese (pt-BR) language requirement consistent for all user-facing error messages? [Consistency, Spec §SC-005]
- [ ] CHK022 - Are decimal field types (balance, equity, margin, etc.) consistently specified across all entities? [Consistency, Spec §Key Entities]
- [ ] CHK023 - Do input validation requirements (FR-009) align with the contract schemas in contracts/? [Consistency, Spec §FR-009 vs contracts/]

---

## Acceptance Criteria Quality

- [ ] CHK024 - Can SC-001 (account info <2s) be objectively measured without implementation details? [Measurability, Spec §SC-001]
- [ ] CHK025 - Can SC-002 (quote <500ms) be objectively measured across all instruments? [Measurability, Spec §SC-002]
- [ ] CHK026 - Can SC-003 (order placement <5s) be objectively verified? [Measurability, Spec §SC-003]
- [ ] CHK027 - Can SC-004 (10 concurrent invocations) be formally verified? [Measurability, Spec §SC-004]
- [ ] CHK028 - Can SC-006 (disconnection detected <10s) be objectively verified? [Measurability, Spec §SC-006]
- [ ] CHK029 - Is SC-007 (zero hardcoded credentials) verifiable via code scanning? [Measurability, Spec §SC-007]
- [ ] CHK030 - Is SC-008 (integration tests for all MCP tools) formally tracked as acceptance criterion? [Acceptance Criteria, Spec §SC-008]

---

## Scenario Coverage

- [ ] CHK031 - Are primary flow requirements complete for all 5 user stories? [Coverage, Spec §User Scenarios]
- [ ] CHK032 - Are alternate flow requirements defined when MT5 terminal is disconnected mid-operation? [Coverage, Spec §Edge Cases]
- [ ] CHK033 - Are exception flow requirements defined for order rejection scenarios (gap, margin, market closed)? [Coverage, Spec §User Story 3 + Edge Cases]
- [ ] CHK034 - Are recovery flow requirements defined for terminal reconnection after disconnection? [Coverage, Gap, Spec §Edge Cases]
- [ ] CHK035 - Are requirements for zero-state scenarios defined (no pending orders, no open positions)? [Coverage, Spec §User Story 4/5]
- [ ] CHK036 - Are concurrent user interaction scenarios addressed (multiple simultaneous tool invocations)? [Coverage, Spec §FR-006 + SC-004]

---

## Edge Case Coverage

- [ ] CHK037 - Are duplicate ticket number handling rules fully specified? [Edge Case, Spec §Edge Cases]
- [ ] CHK038 - Are stale quote detection and handling requirements specified? [Edge Case, Gap]
- [ ] CHK039 - Are "ghost position" detection and resolution requirements specified? [Edge Case, Gap]
- [ ] CHK040 - Is partial order fill handling specified (quantity filled vs requested)? [Edge Case, Gap]
- [ ] CHK041 - Is order state ambiguity resolution specified for terminal disconnection during order placement? [Edge Case, Spec §Edge Cases]
- [ ] CHK042 - Are price gapping scenarios (market gap between request and fill) specified with rejection criteria? [Edge Case, Spec §Edge Cases]

---

## Non-Functional Requirements

- [ ] CHK043 - Are performance requirements (SC-001, SC-002, SC-003, SC-004) formally specified as binding requirements? [NFR, Spec §Success Criteria]
- [ ] CHK044 - Is observability requirement (structured JSON logging with latency metrics) specified with log format schema? [NFR, Spec §FR-005]
- [ ] CHK045 - Are security requirements (TLS, no hardcoded credentials, input validation) complete and traceable? [NFR, Spec §FR-007/FR-008/FR-009]
- [ ] CHK046 - Is concurrency safety requirement (no race conditions per FR-006) formally verifiable? [NFR, Spec §FR-006]

---

## Dependencies & Assumptions

- [ ] CHK047 - Is the assumption "MT5 WebAPI available on port 8228" validated against actual MT5 deployment? [Assumption, Spec §Assumptions]
- [ ] CHK048 - Is MT5 COM/DLL availability assumption still valid given research decision for WebAPI? [Assumption, Spec §Assumptions vs research.md]
- [ ] CHK049 - Are single terminal per MCP instance constraints documented and enforced? [Assumption, Spec §Assumptions]
- [ ] CHK050 - Is the Windows server accessibility for MT5 terminal validated? [Assumption, Gap]

---

## Integration Test Requirements

- [ ] CHK051 - Are integration test requirements complete for all 5 MCP tools? [Integration Testing, Spec §SC-008]
- [ ] CHK052 - Is integration test scope for contract tests vs full MT5 terminal tests differentiated? [Integration Testing, Gap]
- [ ] CHK053 - Are test fixtures and mock strategies for MT5 WebAPI specified? [Integration Testing, Gap]
- [ ] CHK054 - Are error injection scenarios (disconnect, timeout, auth failure) required for integration tests? [Integration Testing, Gap]
- [ ] CHK055 - Is test environment isolation requirement documented (dev vs staging vs production MT5)? [Integration Testing, Gap]

---

## Ambiguities & Conflicts

- [ ] CHK056 - Is "instrument not in market watch" definition for quote tool clarified? [Ambiguity, Spec §User Story 2]
- [ ] CHK057 - Is the behavior when position ticket does not exist but close is attempted consistently defined? [Ambiguity, Spec §User Story 4]
- [ ] CHK058 - Do margin calculation requirements (for insufficient margin error) align between spec and data-model? [Conflict, Spec §User Story 3 vs data-model.md]
- [ ] CHK059 - Is "within 500ms" for quote tool measured from request receipt or from MT5 WebAPI response? [Ambiguity, Spec §SC-002]
- [ ] CHK060 - Are stop-loss/take-profit validation requirements consistent across order placement and position modification? [Consistency, Gap]
