"""
FastAPI Gateway - Minimal, async-first
- gRPC client pool (aiogrpc)
- Structured logging (structlog)
- Health checks
- Zero abstractions
"""

import asyncio
import json
from contextlib import asynccontextmanager

import grpc
import structlog
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel

# Logging setup
structlog.configure(
    processors=[
        structlog.processors.JSONRenderer(),
    ],
    context_class=dict,
    logger_factory=structlog.PrintLoggerFactory(),
)
log = structlog.get_logger()

# gRPC client pool (example)
class GRPCClientPool:
    def __init__(self, host: str = "mt5-adapter:50051"):
        self.host = host
        self.channel: grpc.aio.Channel | None = None

    async def connect(self):
        self.channel = grpc.aio.secure_channel(
            self.host,
            grpc.ssl_channel_credentials(),
        )
        log.msg("grpc_connected", host=self.host)

    async def close(self):
        if self.channel:
            await self.channel.close()
            log.msg("grpc_disconnected")

    async def get_channel(self):
        if not self.channel:
            await self.connect()
        return self.channel


# App lifecycle
pool = GRPCClientPool()

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await pool.connect()
    log.msg("app_startup")
    yield
    # Shutdown
    await pool.close()
    log.msg("app_shutdown")


app = FastAPI(lifespan=lifespan)


# Routes
@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "ok"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics (stub)"""
    return {"requests": 0, "errors": 0}


class TradeRequest(BaseModel):
    symbol: str
    volume: float
    price: float


@app.post("/api/v1/trade")
async def place_trade(req: TradeRequest):
    """Example: route trade request to MT5 adapter via gRPC"""
    try:
        channel = await pool.get_channel()

        # TODO: Call MT5 adapter gRPC method
        # stub = mt5_pb2_grpc.MT5AdapterStub(channel)
        # response = await stub.PlaceTrade(...)

        log.msg("trade_request", symbol=req.symbol, volume=req.volume)
        return {"order_id": 12345, "status": "pending"}

    except grpc.RpcError as e:
        log.msg("grpc_error", code=e.code(), details=e.details())
        raise HTTPException(status_code=503, detail="MT5 adapter unavailable")

    except Exception as e:
        log.msg("unexpected_error", error=str(e))
        raise HTTPException(status_code=500, detail="Internal error")


@app.middleware("http")
async def log_requests(request: Request, call_next):
    """Log all requests (structured)"""
    try:
        response = await call_next(request)
        log.msg(
            "http_request",
            method=request.method,
            path=request.url.path,
            status=response.status_code,
        )
        return response
    except Exception as e:
        log.msg("http_error", method=request.method, path=request.url.path, error=str(e))
        raise


if __name__ == "__main__":
    import uvicorn

    # Production: 4+ workers, SO_REUSEADDR
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8000,
        workers=4,
        access_log=False,  # structlog handles logging
    )
