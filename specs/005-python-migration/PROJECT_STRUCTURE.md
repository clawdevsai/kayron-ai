# Project Structure: Python 3.14 Migration

## Diretórios (minimalista)

```
kayron-ai/
├── services/                    # Microservices migradas
│   ├── mt5-adapter/            # MT5 connection pool + gRPC daemon
│   │   ├── src/
│   │   │   ├── __init__.py
│   │   │   ├── daemon.py       # gRPC server (aiogrpc)
│   │   │   ├── mt5_client.py   # MT5 connection wrapper
│   │   │   ├── pool.py         # Connection pool
│   │   │   └── observability.py # structlog + JSON output
│   │   ├── proto/              # Protocol buffers
│   │   │   └── mt5.proto
│   │   ├── tests/
│   │   │   ├── test_daemon.py
│   │   │   └── test_pool.py
│   │   └── pyproject.toml
│   │
│   ├── api-gateway/            # FastAPI (ASGI)
│   │   ├── src/
│   │   │   ├── __init__.py
│   │   │   ├── main.py         # FastAPI app
│   │   │   ├── routes/         # API endpoints (minimal)
│   │   │   ├── models/         # Pydantic schemas only
│   │   │   ├── grpc_clients.py # aiogrpc stubs
│   │   │   └── observability.py
│   │   ├── proto/
│   │   └── tests/
│   │   └── pyproject.toml
│   │
│   └── worker/                 # Async worker (optional)
│       └── [similar structure]
│
├── shared/                      # Code compartilhado (minimal)
│   ├── proto/                  # Proto definitions centralizadas
│   │   ├── mt5.proto
│   │   ├── api.proto
│   │   └── common.proto
│   └── observability/          # Shared logging config
│       └── __init__.py
│
├── tests/                       # Integration tests
│   ├── contract/               # API contract tests
│   ├── integration/            # End-to-end
│   └── conftest.py
│
├── bin/                         # Scripts utilitários
│   ├── build.sh               # Build ALL services + protos
│   ├── run-local.sh           # Dev server (uvicorn + aiogrpc)
│   └── bench.py               # Performance baseline
│
├── docker/
│   ├── Dockerfile.mt5-adapter  # Multi-stage, alpine
│   └── Dockerfile.api-gateway
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── DEPLOYMENT.md
│   └── PERFORMANCE.md
│
└── pyproject.toml             # Root workspace (optional)
```

## Critérios Otimização

| Aspecto | Escolha | Razão |
|---------|---------|-------|
| **Framework gRPC** | aiogrpc | Async nativo, ~5KB, zero-copy |
| **Web API** | FastAPI | ASGI, ~15KB core, async built-in |
| **Async runtime** | asyncio | Stdlib, sem overhead |
| **Connection pool** | SQLAlchemy 2.0 async | Eficiente, battle-tested |
| **Logging** | structlog + stdout JSON | Leve, machine-parseable, zero-disk |
| **Dependency mgmt** | Poetry | Lock file deterministico |
| **Build** | Makefile lean + poetry | Paralelo, cache-aware |
| **Container** | Alpine + multi-stage | <100MB imagem, fast startup |
| **Test** | pytest + pytest-asyncio | Built-in async support |

## Regras Código

1. **Async-first**: `async def`, `await`, pooled connections
2. **Zero abstractions**: Direct gRPC, no facade layers
3. **Single responsibility**: Funções <30 linhas
4. **No dead code**: Deletar imports/funções não usadas
5. **PEP 8 strict**: Black formatter, 88 char line
6. **Type hints**: `from typing import ...` (Python 3.14 native)
7. **Error handling**: Catch specific exceptions, log structured
8. **Connection pooling**: NEVER new conn per request

## Performance Targets

- **Startup time**: <1s
- **Latency p99**: <50ms (vs Go ~10ms acceptable)
- **Throughput**: 500+ req/s per service
- **Memory**: <200MB per service
- **Docker image**: <150MB (multi-stage)
