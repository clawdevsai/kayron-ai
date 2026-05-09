# SQLAlchemy 2.0 Async Configuration Guide for High-Throughput Microservices

## Executive Summary

For microservices handling 1000+ req/s with minimal latency:

| Setting | Value | Rationale |
|---------|-------|-----------|
| **Pool Type** | QueuePool | Connection reuse, async-safe |
| **pool_size** | 20-32 | 2-4x CPU cores |
| **max_overflow** | 10-16 | Burst capacity without queueing |
| **pool_recycle** | 3600 | Prevent stale connections (< 8h DB timeout) |
| **pool_pre_ping** | True | Lightweight health check (~1ms) |
| **timeout** | 30s | Surface bottlenecks early |
| **expire_on_commit** | False | CRITICAL: Prevents lazy-load blocks |

---

## 1. Connection Pool Architecture

### QueuePool (Recommended)

**When to use:** General-purpose async workloads, high-throughput microservices

- Maintains a queue of reusable connections
- Checks out connections on demand, returns them to pool after use
- Optimal for request-per-connection patterns (web APIs)

**Configuration:**
```python
from sqlalchemy.ext.asyncio import create_async_engine

engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    poolclass=QueuePool,
    pool_size=20,           # Baseline concurrent connections
    max_overflow=10,        # Temporary connections for bursts
    pool_recycle=3600,      # Recycle every hour
    pool_pre_ping=True,     # Validate before use
    timeout=30,             # Queue wait timeout
)
```

### Pool Size Calculation

**Formula:** `pool_size = 2 to 4 × CPU_cores`

| CPU Cores | pool_size | max_overflow | Use Case |
|-----------|-----------|--------------|----------|
| 4 | 8-16 | 4-8 | Small container |
| 8 | 16-32 | 8-16 | Standard microservice |
| 16 | 32-64 | 16-32 | High-traffic service |
| 32 | 64-128 | 32-64 | Bulk processing |

**Reasoning:**
- Each database connection consumes memory (~1-5 MB)
- Each CPU core can safely manage 2-4 concurrent connections
- Too few: tail latency spike (requests queue)
- Too many: memory overhead, connection pool exhaustion on DB side

---

## 2. Overflow Strategy

### Without `max_overflow`
```
Incoming Requests (1000 req/s)
        |
    Pool (size=20)
        |
   Queue (unlimited)
        |
    Result: Unbounded latency, p99 > 1s
```

### With `max_overflow=10`
```
Incoming Requests (1000 req/s)
        |
    Pool (size=20) + Overflow (max=10)
        |
   Available connections: 30 total
        |
    Result: Burst absorbed, p99 < 100ms
```

**Trade-offs:**
- Temporary connections use OS file descriptors
- Each connection: ~1-5 MB memory
- Benefit: Prevents latency spikes during traffic bursts

**Recommendation:** `max_overflow = pool_size / 2`

---

## 3. Connection Recycling

### pool_recycle (seconds)

Database connections idle indefinitely become "stale" when the database closes them due to timeout.

```python
pool_recycle=3600  # Recycle every 1 hour
```

**Why it matters:**
- PostgreSQL default idle timeout: ~8-9 hours
- MySQL default: similar
- Issue: After timeout, connection still in pool but broken on DB side
- Result: "Connection lost" error on next query

**Best value:** 3600-7200 seconds (1-2 hours), less than DB timeout

### pool_pre_ping (boolean)

Validate connection before checking out.

```python
pool_pre_ping=True  # Lightweight SELECT 1
```

**Overhead:** ~1ms per checkout (negligible vs. 100+ ms queries)

**Benefit:** Catches broken connections early, prevents cascading failures

---

## 4. Async Session Configuration (CRITICAL)

### expire_on_commit=False

This is the single most important setting for async code.

```python
# WRONG (will cause deadlocks in async)
async_sessionmaker(engine, expire_on_commit=True)

# CORRECT (prevents lazy-load blocks)
async_sessionmaker(
    engine,
    expire_on_commit=False,  # Don't reset lazy loaders
    autoflush=False,          # Explicit flush control
    autocommit=False,         # Standard transaction behavior
)
```

**Why?**
- `expire_on_commit=True`: Objects cleared after commit, lazy loads block event loop
- `expire_on_commit=False`: Objects remain accessible, eager-load only what you need

**Pattern:**
```python
from sqlalchemy.orm import selectinload
from sqlalchemy.future import select

# Eager load relationships upfront
result = await session.execute(
    select(User).options(selectinload(User.posts))
)
users = result.scalars().unique().all()

# No lazy loading needed post-commit
```

---

## 5. Async Concurrency Patterns

### asyncio.Semaphore for Rate Limiting

```python
import asyncio

pool_semaphore = asyncio.Semaphore(20)  # Match pool_size

async def query_with_rate_limit(session, query):
    async with pool_semaphore:
        return await session.execute(query)
```

### Context-Scoped Sessions

Use `async_scoped_session` with `contextvars` for per-request isolation:

```python
from sqlalchemy.ext.asyncio import async_scoped_session
import contextvars

request_context = contextvars.ContextVar("request_id")

AsyncSessionLocal = async_scoped_session(
    async_sessionmaker(engine, expire_on_commit=False),
    scopefunc=lambda: request_context.get(),
)

# In FastAPI middleware:
@app.middleware("http")
async def set_request_context(request, call_next):
    token = request.headers.get("x-request-id", "default")
    request_context.set(token)
    return await call_next(request)
```

---

## 6. Monitoring Pool Health

### Pool Statistics

```python
def get_pool_stats(engine):
    pool = engine.pool
    return {
        "active": pool.checkedout(),      # Currently in use
        "idle": pool.checkedin(),          # Available to checkout
        "size": pool.size(),               # Total pooled connections
        "overflow": pool.overflow(),       # Temporary connections
        "capacity": pool.size() + pool.overflow(),
    }

# Example: {"active": 15, "idle": 5, "size": 20, "overflow": 0, "capacity": 30}
```

### Metrics to Monitor

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| Pool utilization | 60-80% | > 90% |
| Queue wait time | < 5ms | > 50ms |
| Active connections | < pool_size | == pool_size (bottleneck) |
| Connection errors | 0 | > 0 |

---

## 7. Production Configuration Checklist

- [ ] Use `pool_size = 2-4 × CPU_cores`
- [ ] Set `max_overflow = pool_size / 2`
- [ ] Enable `pool_pre_ping=True`
- [ ] Set `pool_recycle < DB_idle_timeout`
- [ ] Set `timeout=30` (queue wait limit)
- [ ] Disable `echo=False, echo_pool=False`
- [ ] Use `expire_on_commit=False` in async_sessionmaker
- [ ] Eager-load relationships (selectinload, joinedload)
- [ ] Implement per-request session pattern
- [ ] Monitor pool metrics (active, idle, overflow)
- [ ] Test under 1000+ req/s load
- [ ] Set up connection pooling proxy (PgBouncer, ProxySQL) for multi-region

---

## 8. Anti-Patterns to Avoid

```python
# WRONG: Defeats pooling benefit
create_async_engine(
    url,
    pool_size=1,
    max_overflow=100,  # Creates 100 connections for single overflow
)

# WRONG: Stale connections
create_async_engine(url, pool_recycle=None)

# WRONG: Lazy loading in async (DEADLOCK)
async_sessionmaker(engine, expire_on_commit=True)
result = await session.execute(select(User))
user = result.scalar()
print(user.posts)  # BLOCKS EVENT LOOP

# WRONG: ThreadLocalRegistry in async
from sqlalchemy.orm import scoped_session
Session = scoped_session(sessionmaker(engine))  # NOT async-safe

# WRONG: No connection validation
create_async_engine(url, pool_pre_ping=False)  # Stale connections slip through
```

---

## 9. Reference Configurations

### Microservice (8-core, 1000+ req/s)
```python
engine = create_async_engine(
    "postgresql+asyncpg://...",
    poolclass=QueuePool,
    pool_size=20,
    max_overflow=10,
    pool_recycle=3600,
    pool_pre_ping=True,
    timeout=30,
    echo=False,
)
```

### High-Traffic (16-core, 10,000+ req/s)
```python
engine = create_async_engine(
    "postgresql+asyncpg://...",
    poolclass=QueuePool,
    pool_size=32,
    max_overflow=16,
    pool_recycle=3600,
    pool_pre_ping=True,
    timeout=30,
    echo=False,
)
```

### Lightweight (4-core, development)
```python
engine = create_async_engine(
    "postgresql+asyncpg://...",
    poolclass=QueuePool,
    pool_size=8,
    max_overflow=4,
    pool_recycle=3600,
    pool_pre_ping=True,
    timeout=30,
    echo=True,  # OK for dev
)
```

### Serverless/Lambda (no persistent pool)
```python
engine = create_async_engine(
    "postgresql+asyncpg://...",
    poolclass=NullPool,  # New connection per request
    echo=False,
)
```

---

## 10. Load Testing & Tuning

### Test Script
```python
import asyncio
import time
from sqlalchemy.future import select

async def load_test(engine, num_tasks=1000, duration_sec=60):
    session_factory = async_sessionmaker(
        engine,
        expire_on_commit=False,
    )
    
    start = time.time()
    tasks = []
    
    async def query_task():
        async with session_factory() as session:
            await session.execute(select(1))
    
    while time.time() - start < duration_sec:
        for _ in range(min(100, num_tasks)):
            tasks.append(query_task())
        await asyncio.gather(*tasks)
        tasks.clear()
    
    stats = get_pool_stats(engine)
    print(f"Pool stats: {stats}")

# Run: asyncio.run(load_test(engine, num_tasks=1000))
```

### Tuning Steps
1. Start with `pool_size = 2 × CPU_cores`
2. Run 1000+ req/s load test
3. Monitor: active connections, queue wait time
4. If active > 90% of pool_size: increase pool_size
5. If queue wait > 50ms: increase max_overflow
6. Verify p99 latency < 100ms

---

## References

- SQLAlchemy 2.0 Async Documentation: https://docs.sqlalchemy.org/en/20/orm/extensions/asyncio.html
- asyncio Best Practices: https://docs.python.org/3/library/asyncio-task.html
- Connection Pool Tuning: https://wiki.postgresql.org/wiki/Number_Of_Database_Connections
