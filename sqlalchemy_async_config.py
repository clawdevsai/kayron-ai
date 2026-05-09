"""
SQLAlchemy 2.0 Async Configuration for High-Throughput Microservices
Target: 1000+ req/s, minimal latency, robust connection pooling
"""

from sqlalchemy.ext.asyncio import (
    create_async_engine,
    AsyncSession,
    async_scoped_session,
    async_sessionmaker,
)
from sqlalchemy.pool import QueuePool, NullPool
from contextlib import asynccontextmanager
import contextvars
import asyncio

# ============================================================================
# CONFIGURATION PROFILES
# ============================================================================

class AsyncPoolConfig:
    """Production-grade async pool configurations."""

    # PROFILE 1: Microservices (8-16 cores, 1000+ req/s)
    MICROSERVICE_8CORE = {
        "poolclass": QueuePool,
        "pool_size": 20,
        "max_overflow": 10,
        "pool_recycle": 3600,
        "pool_pre_ping": True,
        "echo_pool": False,
        "timeout": 30,
    }

    # PROFILE 2: High-Traffic (16+ cores, heavy load)
    HIGHTRAFFIC_16CORE = {
        "poolclass": QueuePool,
        "pool_size": 32,
        "max_overflow": 16,
        "pool_recycle": 3600,
        "pool_pre_ping": True,
        "echo_pool": False,
        "timeout": 30,
    }

    # PROFILE 3: Lightweight (single container, 4 cores)
    LIGHTWEIGHT_4CORE = {
        "poolclass": QueuePool,
        "pool_size": 8,
        "max_overflow": 4,
        "pool_recycle": 3600,
        "pool_pre_ping": True,
        "echo_pool": False,
        "timeout": 30,
    }

    # PROFILE 4: Serverless/FaaS (no persistent pool)
    SERVERLESS = {
        "poolclass": NullPool,
        "echo_pool": False,
    }


# ============================================================================
# FACTORY FUNCTION
# ============================================================================

def create_async_db_engine(
    db_url: str,
    profile: dict = None,
    **kwargs
):
    """
    Create optimized async engine for high-throughput scenarios.

    Args:
        db_url: Connection string (e.g., 'postgresql+asyncpg://...')
        profile: Pool configuration dict (uses MICROSERVICE_8CORE if None)
        **kwargs: Additional engine options

    Returns:
        AsyncEngine with optimized pooling

    Example:
        engine = create_async_db_engine(
            "postgresql+asyncpg://user:pass@localhost/db",
            profile=AsyncPoolConfig.MICROSERVICE_8CORE
        )
    """
    if profile is None:
        profile = AsyncPoolConfig.MICROSERVICE_8CORE

    # Extract pool-specific args
    pool_config = {k: v for k, v in profile.items() if k != "poolclass"}
    poolclass = profile.get("poolclass", QueuePool)

    # Engine options
    engine_kwargs = {
        "echo": False,  # Disable SQL logging in production
        "future": True,  # SQLAlchemy 2.0 style
        "poolclass": poolclass,
        **pool_config,
        **kwargs,
    }

    return create_async_engine(db_url, **engine_kwargs)


# ============================================================================
# ASYNC SESSION FACTORY (Request-scoped)
# ============================================================================

_session_context_var = contextvars.ContextVar(
    "async_session",
    default=None
)


async def get_async_session_factory(engine):
    """
    Create async_sessionmaker with recommended async settings.

    Critical settings:
    - expire_on_commit=False: Prevents lazy loading post-commit (blocks in async)
    - autoflush=False: Explicit control in async context
    - autocommit=False: Standard transactional behavior
    """
    return async_sessionmaker(
        engine,
        class_=AsyncSession,
        expire_on_commit=False,  # CRITICAL for async
        autoflush=False,
        autocommit=False,
    )


async def get_scoped_session(session_factory):
    """
    Return async_scoped_session bound to context var.
    Use this for request-scoped session management (FastAPI, etc).
    """
    return async_scoped_session(
        session_factory,
        scopefunc=lambda: _session_context_var.get(),
    )


# ============================================================================
# CONTEXT MANAGER (Per-request pattern)
# ============================================================================

@asynccontextmanager
async def get_db_session(session_factory):
    """
    Per-request session context manager.

    Usage (FastAPI):
        @app.get("/items")
        async def list_items(session = Depends(get_db_session)):
            result = await session.execute(select(Item))
            return result.scalars().all()
    """
    session = session_factory()
    try:
        yield session
        await session.commit()
    except Exception:
        await session.rollback()
        raise
    finally:
        await session.close()


# ============================================================================
# MONITORING & DIAGNOSTICS
# ============================================================================

def get_pool_stats(engine):
    """
    Inspect pool health and utilization.

    Returns:
        dict with:
        - checkedout: Active connections
        - size: Total pooled connections
        - overflow: Temporary connections
        - checkedin: Available for reuse
    """
    pool = engine.pool
    return {
        "checkedout": pool.checkedout(),
        "size": pool.size(),
        "overflow": pool.overflow(),
        "checkedin": pool.checkedin(),
        "total_capacity": pool.size() + pool.overflow(),
    }


async def health_check(engine):
    """
    Validate database connectivity.
    Returns True if connection successful, False otherwise.
    """
    try:
        async with engine.connect() as conn:
            result = await conn.execute("SELECT 1")
            return result.scalar() == 1
    except Exception as e:
        print(f"Health check failed: {e}")
        return False


# ============================================================================
# EXAMPLE USAGE (FastAPI)
# ============================================================================

"""
# In your FastAPI app:

from fastapi import FastAPI, Depends
from sqlalchemy.future import select

app = FastAPI()

# Initialize
engine = create_async_db_engine(
    "postgresql+asyncpg://user:pass@localhost/mydb",
    profile=AsyncPoolConfig.MICROSERVICE_8CORE
)

SessionLocal = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)


async def get_session():
    async with AsyncSession(engine) as session:
        yield session


@app.get("/items")
async def list_items(session: AsyncSession = Depends(get_session)):
    result = await session.execute(select(Item))
    return result.scalars().all()


@app.on_event("startup")
async def startup():
    # Verify DB connectivity
    is_healthy = await health_check(engine)
    if not is_healthy:
        raise RuntimeError("Database unreachable")


@app.on_event("shutdown")
async def shutdown():
    await engine.dispose()  # Drain connections gracefully
"""
