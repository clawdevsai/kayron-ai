# Phase 0 Artifacts - Python 3.14 Migration

## Templates Criados (Pronto para Uso)

### Estrutura & Build
- **PROJECT_STRUCTURE.md** — Layout minimalista (services/, shared/, tests/, bin/, docker/)
- **pyproject-template.toml** — Deps otimizadas (grpcio, fastapi, sqlalchemy 2.0, structlog)
- **Makefile-template** — Build rápido (proto, install, test, docker targets)
- **Dockerfile-template** — Alpine multi-stage (<150MB imagem)

### Código (Skeletons)
- **main-fastapi-skeleton.py** — FastAPI gateway minimalista (gRPC client pool, structlog)
- **daemon-grpc-skeleton.py** — gRPC daemon (MT5 pool, async-first, graceful shutdown)

## Decisões Phase 0 (Pesquisa)

| Aspecto | Decisão | Status |
|---------|---------|--------|
| **gRPC Framework** | grpcio + grpc.aio | ✅ PESQUISADO |
| **Web Framework** | FastAPI | ✅ PESQUISADO |
| **Build System** | Hybrid: Poetry + Make | ✅ PESQUISADO |
| **MT5 Binding** | Aguardando... | ⏳ PESQUISANDO |
| **Database Pool** | SQLAlchemy 2.0 async | ⏳ PESQUISANDO |
| **Performance** | 2-4x Go (aceitável) | ✅ PESQUISADO |

## Pesquisa Agentes Ativas

1. ✅ **aiogrpc vs grpcio**: grpcio VENCE (grpc.aio nativo, 10x mais rápido)
2. ✅ **Poetry vs Make**: Hybrid (Poetry deps + Make protoc)
3. ✅ **FastAPI+aiogrpc benchmarks**: 2-4x Go (aceitável), 5-10x FastAPI (OK)
4. ⏳ **MT5 Python bindings**: Pesquisando melhor wrapper
5. ⏳ **SQLAlchemy async pool**: Pesquisando config otimizada

## Próximos Passos (Após Phase 0 Completa)

### Phase 1: Design & Contracts
- [ ] Consolidar research.md (decisões + rationale)
- [ ] Gerar data-model.md (entidades, relacionamentos)
- [ ] Gerar contracts/mt5.proto
- [ ] Gerar contracts/api.proto
- [ ] Gerar quickstart.md (setup + run local)
- [ ] Atualizar CLAUDE.md agent context

### Phase 2: Tasks
- [ ] /speckit-tasks para gerar tasks.md (implementation steps)
- [ ] /speckit-taskstoissues para sync com GitHub

## Critérios Sucesso

- **Custo**: Zero licenses, minimal infra
- **Performance**: 2-4x Go baseline (p99 <50ms)
- **Código**: <30 linhas per function, zero dead code, PEP 8 strict
- **Deploy**: Alpine images <150MB, startup <1s
- **Tests**: 100% async/await patterns, contract tests pass
