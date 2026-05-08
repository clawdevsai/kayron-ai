# Development Guide: MT5 MCP Integration

## Setup

1. Install Go 1.21+ and protoc
2. Clone repo and cd to project root
3. Run `make proto` to generate gRPC stubs
4. Run `go mod download` to fetch dependencies

## Building

```bash
make build
```

## Testing

```bash
make test
```

## Project Structure

- `cmd/mcp-mt5-server/` — MCP server entry point
- `internal/models/` — Data models (Account, Order, Position, Quote)
- `internal/services/` — Business logic (MT5 client, gRPC services, daemon)
- `internal/contracts/` — Protocol Buffer definitions
- `internal/logger/` — Logging utilities
- `tests/` — Unit, integration, contract tests

## Development Notes

- Use `shops/decimal` for all financial calculations (no float64)
- All error messages should be in pt-BR
- gRPC services must handle disconnection gracefully
- Tests should use mock MT5 client for unit tests
