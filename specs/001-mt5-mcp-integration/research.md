# Research: MT5 MCP Integration

**Feature**: 001-mt5-mcp-integration
**Date**: 2026-05-08
**Status**: Partial — 3 unknowns remain

---

## 1. MT5 Go Integration Mechanism

### Decision

MT5 WebAPI (HTTP/JSON) is the recommended integration path. Direct DLL/COM or cgo bindings are complex and Windows-specific.

### Rationale

- MT5 provides official WebAPI (port 8228) — HTTP/JSON interface
- No official Go SDK; third-party MT4/MT5 Go libraries are scarce and unmaintained
- WebAPI avoids cgo complexity, works from Linux to Windows MT5 terminal
- Alternative: MT5 COM API via `github.com/go-ole/go-ole` (Windows only, heavier)

### Alternatives Considered

| Alternative | Status | Why Rejected |
|-------------|--------|--------------|
| Direct MT5 DLL via cgo | Complex | Requires cgo + Windows build environment; fragile ABI |
| MT5 COM API via go-ole | Windows only | Adds COM dependency; cross-compilation harder |
| Third-party MT4 Go lib | Unmaintained | No reliable MT5 Go library found |
| MT5 WebAPI | **Selected** | HTTP/JSON, cross-platform, simplest path |

---

## 2. Go Decimal Library for Financial Calculations

### Decision

Use `github.com/shopsam/decimal` (shops/decimal) — mature, widely-used, precise.

### Rationale

- Go has no built-in decimal type; float64 is unsuitable for financial values
- `shops/decimal` is the de-facto standard (12k+ stars, active maintenance)
- Alternatives: `github.com/ericlagerlöf/decimal` — similar API, fewer downloads

### Alternatives Considered

| Alternative | Downloads | Status |
|-------------|-----------|--------|
| shops/decimal | ~50M/yr | **Selected** — standard choice |
| ericlagerlöf/decimal | ~5M/yr | Viable alternative |
| go-sql-driver decimal | N/A | For DB only, not general arithmetic |

---

## 3. MCP + gRPC Bridge Architecture

### Decision

MCP server as Go library + CLI. MCP handles JSON-RPC 2.0 over stdio; internal gRPC for service decomposition.

### Rationale

- MCP spec: JSON-RPC 2.0 over stdio (server-initiated notifications supported)
- gRPC: internal service calls (not exposed to MCP clients)
- MCP tools delegate to gRPC services internally
- Bridge pattern: MCP handler → gRPC client → MT5 WebAPI

### Architecture

```
MCP Client (stdio)
    │
    ▼
MCP Handler (Go)
    │
    ├─► AccountService (gRPC)
    │       └─► MT5 WebAPI Client
    │
    ├─► QuoteService (gRPC)
    │       └─► MT5 WebAPI Client
    │
    └─► OrderService (gRPC)
            └─► MT5 WebAPI Client
```

---

## 4. TDD Confirmation (RESOLVED)

**Decision**: Option B — tests written alongside implementation, still required before PR merge.

### Updated Summary

| Unknown | Decision | Confidence |
|---------|----------|------------|
| MT5 Go integration | MT5 WebAPI (HTTP/JSON) | High |
| Decimal library | shops/decimal | High |
| MCP-gRPC bridge | Library + CLI, bridge pattern | High |
| TDD confirmation | Option B (tests with implementation) | **Resolved** |

---

*Research ready. Proceed to Phase 1 (Design) once TDD clarification resolved.*
