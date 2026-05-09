"""
gRPC Daemon (MT5 Adapter) - Minimal, async-first, zero-overhead
- Connection pool (singleton pattern)
- Structured logging
- Graceful shutdown
- Protocol buffer stubs (generated)
"""

import asyncio
import signal
import sys
from typing import AsyncGenerator

import grpc
import structlog
from grpc import aio

# Import generated protobuf stubs
# from . import mt5_pb2, mt5_pb2_grpc

log = structlog.get_logger()


class MT5Pool:
    """Reusable MT5 connection pool (singleton)"""

    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self, host: str = "localhost", port: int = 5000, pool_size: int = 10):
        if self._initialized:
            return
        self._initialized = True

        self.host = host
        self.port = port
        self.pool_size = pool_size
        self.pool: list = []
        self._lock = asyncio.Lock()

    async def connect_all(self):
        """Pre-warm connection pool"""
        log.msg("pool_init", size=self.pool_size)
        for _ in range(self.pool_size):
            conn = await self._create_connection()
            self.pool.append(conn)

    async def _create_connection(self):
        """Create single MT5 connection (stub implementation)"""
        # TODO: Replace with actual MT5 WebAPI/ctypes/mt5-async wrapper
        return {"id": id({}), "connected": True}

    async def get(self) -> AsyncGenerator:
        """Get connection from pool (context manager)"""
        async with self._lock:
            if not self.pool:
                log.msg("pool_exhausted")
                raise RuntimeError("No available MT5 connections")
            conn = self.pool.pop()

        try:
            yield conn
        finally:
            async with self._lock:
                self.pool.append(conn)

    async def close_all(self):
        """Close all pooled connections"""
        log.msg("pool_close", count=len(self.pool))
        for conn in self.pool:
            # TODO: Close connection gracefully
            pass
        self.pool.clear()


class MT5Adapter:
    """gRPC service implementation"""

    def __init__(self):
        self.pool = MT5Pool()

    async def PlaceTrade(self, request, context):
        """Example: Place trade order via MT5

        Args:
            request: mt5_pb2.TradeRequest
            context: grpc.ServicerContext

        Returns:
            mt5_pb2.TradeResponse
        """
        try:
            async with self.pool.get() as conn:
                log.msg(
                    "trade_place",
                    symbol=request.symbol,
                    volume=request.volume,
                    pool_id=conn["id"],
                )

                # TODO: Call actual MT5 WebAPI
                # result = await mt5_adapter.place_order(...)

                # Stub response
                return {  # mt5_pb2.TradeResponse
                    "order_id": 12345,
                    "status": "ACCEPTED",
                    "message": "Order queued",
                }

        except Exception as e:
            log.msg("trade_error", error=str(e))
            await context.abort(grpc.StatusCode.INTERNAL, f"MT5 error: {e}")

    async def GetBalance(self, request, context):
        """Example: Get account balance"""
        try:
            async with self.pool.get() as conn:
                log.msg("balance_query", pool_id=conn["id"])

                # TODO: Call actual MT5 WebAPI
                return {  # mt5_pb2.BalanceResponse
                    "balance": 100000.00,
                    "equity": 95000.00,
                    "margin_free": 50000.00,
                }

        except Exception as e:
            log.msg("balance_error", error=str(e))
            await context.abort(grpc.StatusCode.INTERNAL, f"MT5 error: {e}")


async def serve(host: str = "0.0.0.0", port: int = 50051):
    """Start gRPC server"""

    # Setup signal handlers
    loop = asyncio.get_event_loop()

    def signal_handler(sig):
        log.msg("signal_received", signal=sig)
        asyncio.create_task(shutdown(server))

    for sig in (signal.SIGTERM, signal.SIGINT):
        loop.add_signal_handler(sig, lambda s=sig: signal_handler(s))

    # Initialize pool
    pool = MT5Pool()
    await pool.connect_all()

    # Create gRPC server
    server = aio.server(
        options=[
            ("grpc.max_concurrent_streams", 100),
            ("grpc.max_receive_message_length", 4 * 1024 * 1024),  # 4MB
        ]
    )

    # Register service
    adapter = MT5Adapter()
    # mt5_pb2_grpc.add_MT5AdapterServicer_to_server(adapter, server)

    # Bind address
    server.add_insecure_port(f"{host}:{port}")

    log.msg("server_startup", host=host, port=port)
    await server.start()

    return server, pool


async def shutdown(server):
    """Graceful shutdown"""
    log.msg("shutdown_initiated")
    await server.stop(grace=10)
    log.msg("server_stopped")


async def main():
    """Entry point"""
    structlog.configure(
        processors=[
            structlog.processors.JSONRenderer(),
        ],
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
    )

    try:
        server, pool = await serve()

        # Keep server running
        await asyncio.sleep(float("inf"))

    except KeyboardInterrupt:
        log.msg("interrupted")
        await pool.close_all()
        sys.exit(0)

    except Exception as e:
        log.msg("fatal_error", error=str(e), exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
