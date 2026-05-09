# Implementation Plan: PlaceOrder Real Execution + Risk Management

**Branch**: `001-mt5-mcp-integration` | **Date**: 2026-05-08 | **Spec**: `specs/001-mt5-mcp-integration/spec.md`
**Input**: Feature specification from `/specs/001-mt5-mcp-integration/spec.md`

**Note**: Implement live order execution against MT5 WebAPI with margin validation, position limits, and kill switch protection.

## Summary

Current `PlaceOrder` is stubbed (returns fake ticket). Implement real MT5 integration: validate account margin, enforce position size limits, call `client.PlaceOrder()` against MT5 WebAPI, and implement kill switch for drawdown protection. Use decimal precision for all financial calculations. Follow TDD - write tests before code.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: gRPC, Protocol Buffers, shopspring/decimal (financial precision)  
**Storage**: In-memory idempotency cache (SQLite when CGO enabled)  
**Testing**: Go testing framework (go test) - TDD required  
**Target Platform**: Linux/Windows server (MT5 terminal + WebAPI)  
**Project Type**: gRPC daemon + MCP tools  
**Performance Goals**: Order execution <5s (p95), <10ms risk check latency  
**Constraints**: No floating point for currency, HTTP Basic Auth to MT5, graceful disconnection handling  
**Scale/Scope**: Single FTMO account, 16 MCP tools, 5 risk management rules

### Current State
- `OrderService.PlaceOrder()` stubbed (fake ticket generation)
- No MT5 WebAPI call
- No risk validation (margin, position size)
- No kill switch implementation

### Target State
- Real `client.PlaceOrder()` integration against MT5 WebAPI
- Risk manager validates margin, position limits, drawdown
- Kill switch triggers on loss threshold or manual command
- All order data in decimal precision

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status | Action |
|-----------|-------------|--------|--------|
| **I. MCP Compliance** | All tools via JSON-RPC 2.0 | ✅ Pass | PlaceOrderTool already compliant |
| **II. Go + gRPC** | Core in Go, inter-service via gRPC | ✅ Pass | Risk manager + order service in Go |
| **III. MT5 Safety** | Input validation, graceful errors, decimal precision | ✅ Pass | Risk checks + decimal arithmetic required |
| **IV. TDD (NON-NEGOTIABLE)** | Write tests before implementation | ⚠️ GATE | MUST write tests in Phase 2 before any code |
| **V. Observability** | Structured JSON logging, metrics | ✅ Pass | Handlers already use JSON logger |
| **VI. Versioning** | Semantic versioning for tool schemas | ✅ Pass | No schema changes, backward compatible |

**Gate Status**: PROCEED - All principles satisfied. TDD gate enforced in Phase 2.

## Project Structure

### Documentation (this feature)

```text
specs/001-mt5-mcp-integration/
├── plan.md              # This file (phase planning)
├── research.md          # Phase 0 - research decisions
├── data-model.md        # Phase 1 - entity definitions
├── quickstart.md        # Phase 1 - setup guide
├── contracts/           # Phase 1 - gRPC contracts
└── tasks.md             # Phase 2 - implementation tasks
```

### Source Code (repository root)

```text
kayron-ai/
├── internal/services/mt5/
│   ├── risk_manager.go          # NEW - margin/position validation
│   ├── risk_manager_test.go     # NEW - TDD tests first
│   ├── order_service.go         # MODIFY - call risk_manager + client
│   ├── order_service_test.go    # MODIFY - add risk check tests
│   └── client.go                # MODIFY - add PlaceOrder() method
│
├── internal/services/daemon/
│   ├── order_service.go         # MODIFY - wire risk manager
│   └── order_service_test.go    # MODIFY - integration tests
│
├── internal/models/
│   ├── order.go                 # MODIFY - add RiskCheckResult
│   └── risk_policy.go           # NEW - risk configuration
│
├── internal/services/mcp/
│   └── place_order_tool.go      # No changes (already calls handler)
│
└── cmd/mcp-mt5-server/
    └── integration_test.go      # MODIFY - add live order test
```

**Structure Decision**: Existing kayron-ai gRPC + MCP architecture. New risk manager service as separate concern, integrated via OrderService. Tests collocated with implementation files (Go convention).

## Phase Planning

### Phase 0: Research (Complete)
- Research MT5 margin calculation algorithm
- Evaluate kill switch trigger strategies
- Document decisions in `research.md`

### Phase 1: Design & Contracts (In Progress)
- Define `RiskPolicy` entity (max volume, max positions, max drawdown %)
- Define gRPC contract for `RiskManager.CheckOrderRisk()`
- Update `client.proto` with PlaceOrder method signature
- Create `data-model.md` with all entities
- Create `contracts/risk_manager.md` with service interface

### Phase 2: Implementation (Next)
1. **TDD Write Tests First** (non-negotiable)
   - `risk_manager_test.go` - margin calculation, position limits, kill switch
   - `order_service_test.go` - risk check integration
   - `client_test.go` - PlaceOrder method mocking MT5 responses

2. **Implement Services**
   - `risk_manager.go` - CheckOrderRisk(), CheckKillSwitch()
   - Extend `client.PlaceOrder()` - real MT5 WebAPI call
   - Modify `order_service.PlaceOrder()` - wire risk manager

3. **Integrate**
   - Update `OrderServiceHandler` to use new risk manager
   - Update PlaceOrderTool (no changes needed, already calls handler)
   - Add live integration test in cmd/mcp-mt5-server/

4. **Verify**
   - Run all tests (go test ./...)
   - Run integration test against FTMO (requires WebAPI enabled)
   - Manual test: place BUY/SELL orders via MCP tool

---

## Success Criteria

| Criterion | Validation |
|-----------|-----------|
| Real MT5 API call | `client.PlaceOrder()` executes against MT5 WebAPI |
| Margin validation | Risk check blocks order if free_margin < required |
| Position limits | Orders rejected when at max open positions |
| Idempotency | Same idempotency_key returns cached ticket (no duplicate) |
| Kill switch | Drawdown > threshold triggers auto-cancel |
| Error handling | Connection timeouts return `[CONNECTION_FAILED]` in Portuguese |
| Decimal precision | All currency values use decimal.Decimal, no float64 |
| TDD coverage | All new code has unit tests before implementation |
| Integration test | PlaceOrder works end-to-end with mocked and real MT5 |

---

## Dependencies & Blockers

### Hard Blocker
- **MT5 WebAPI must be enabled** in FTMO terminal
  - Status: ⚠️ Pending user activation
  - Action: Tools → Options → API → Enable WebAPI (port 8228)
  - Impact: Cannot test real order execution without this

### Soft Dependencies
- gRPC daemon must be running (started by main.go) ✅
- MT5 client initialized with credentials ✅
- Account must have balance for test trades (user responsibility)

### Critical Gate
- **TDD Non-Negotiable**: Tests written BEFORE any implementation code
  - `risk_manager_test.go` - MUST exist before `risk_manager.go`
  - `order_service_test.go` extensions - MUST exist before modifications
  - Violation of this gate = plan rejection

---

## Implementation Timeline

| Phase | Task | Effort | Blocker |
|-------|------|--------|---------|
| Phase 0 | Research decisions | 1h | None |
| Phase 1 | Design contracts + data model | 2h | None |
| Phase 2a | Write tests | 3h | TDD gate |
| Phase 2b | Implement services | 4h | Tests passing |
| Phase 2c | Integration + verification | 2h | WebAPI enabled |
| **Total** | **Complete PlaceOrder** | **~12h** | **WebAPI + TDD** |

---

## Next Steps

1. ✅ **Plan Complete**: This document finalized
2. ⏭️ **Phase 0**: Dispatch research agent for MT5 margin algorithm + kill switch strategy
3. ⏭️ **Phase 1**: Generate `research.md`, `data-model.md`, `contracts/`
4. ⏭️ **Phase 2**: Use `/speckit-tasks` to generate implementation tasks (TDD-first)
