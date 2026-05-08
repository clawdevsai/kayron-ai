<!-- Sync Impact Report -->
<!-- Version: 0.1.0 → 1.0.0 (initial constitution) -->
<!-- Modified: N/A (new) -->
<!-- Added: Core Principles I-V, Development Workflow, Security Requirements, Governance -->
<!-- Removed: none -->
<!-- Templates: ✅ plan-template.md aligned / ⚠ spec-template.md pending review / ⚠ tasks-template.md pending review -->
<!-- Deferred: none -->

# Kayron AI Constitution

## Core Principles

### I. MCP Protocol Compliance
Kayron AI exposes functionality exclusively via MCP (Model Context Protocol). All tools MUST conform to MCP JSON-RPC 2.0 specification. No ad-hoc HTTP endpoints or non-MCP interfaces.

### II. Go + gRPC First
Core services use Go. Inter-service communication uses gRPC with Protocol Buffers. No alternative RPC frameworks unless explicitly approved by governance.

### III. MetaTrader 5 Integration Safety
MCP tools for MT5 MUST validate all inputs, handle terminal disconnections gracefully, and never expose raw market data without proper error handling. Financial calculations require decimal precision — never use floating point for currency values.

### IV. Test-Driven Development (NON-NEGOTIABLE)
Tests MUST be written before implementation for new MCP tools and critical paths. Red-Green-Refactor cycle enforced. Integration tests verify gRPC service contracts and MT5 terminal responses.

### V. Observability
Structured logging required for all MCP tool invocations. Metrics for latency, error rates, and MT5 terminal health. Logs in JSON format for machine parsing.

### VI. Versioning & Compatibility
MCP tool schemas use semantic versioning. Breaking changes require MAJOR version bump and deprecation warning. Client compatibility MUST be maintained within same major version.

### VII. Security by Default
All gRPC services use TLS. MT5 credential handling follows secrets management best practices. No hardcoded credentials. Input validation on all boundaries.

## Development Workflow

### Development Standards
- All MCP tools documented with JSON schema
- gRPC services follow buf schema validation
- Go modules use semantic versioning (MAJOR.MINOR.PATCH)
- All PRs require review before merge

### Code Review Requirements
- Minimum 1 reviewer for MCP tool changes
- Security-sensitive code (MT5 integration, credential handling) requires 2 reviewers
- Reviewer must verify tests pass locally before approval

### Quality Gates
- `go vet` clean
- `go test ./...` passing
- Protocol buffer schemas validated with `buf lint`
- No new dependencies without justification in PR

## Security Requirements

### Secrets Management
- MT5 credentials stored via environment variables or secrets manager
- No credentials in source code or config files
- gRPC channel credentials required for production

### Input Validation
- All MCP tool inputs validated against JSON schema
- MT5 terminal responses sanitized before processing
- Buffer overflow protection in Go code

## Governance

This constitution supersedes all other practices. Amendments require:
1. PR with constitution changes
2. Review and approval
3. Migration plan for existing MCP tool consumers
4. Version bump (MAJOR for breaking, MINOR for additions, PATCH for clarifications)

All PRs and reviews must verify compliance with these principles. Complexity must be justified — simple solutions preferred.

**Version**: 1.0.0 | **Ratified**: 2026-05-08 | **Last Amended**: 2026-05-08