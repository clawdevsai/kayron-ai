"""MT5 gRPC Service Implementation"""
import asyncio
import uuid
from typing import Optional
from datetime import datetime
from .mt5_adapter import MT5Adapter
from .connection_pool import ConnectionPool
from .operation_queue import OperationQueue
from .callback_manager import CallbackManager
from .logging_config import StructuredLogger, setup_logging


class MT5Service:
    """Main gRPC service for MT5 operations"""

    def __init__(self, pool: ConnectionPool, queue: OperationQueue, callbacks: CallbackManager):
        self.pool = pool
        self.queue = queue
        self.callbacks = callbacks
        self.logger = StructuredLogger(setup_logging())

    async def execute_order_operation(
        self,
        session_id: str,
        agent_id: str,
        stream_id: str,
        symbol: str,
        operation_type: str,
        volume: float,
        price: float,
        stop_loss: Optional[float] = None,
        take_profit: Optional[float] = None,
    ) -> dict:
        """Execute order operation"""
        operation_id = str(uuid.uuid4())
        start_time = datetime.utcnow()

        try:
            # Queue operation
            request_data = {
                "symbol": symbol,
                "type": operation_type,
                "volume": volume,
                "price": price,
                "stop_loss": stop_loss,
                "take_profit": take_profit,
            }

            self.queue.enqueue(session_id, agent_id, "PlaceOrder", request_data)
            self.callbacks.push_update_to_stream(stream_id, operation_id, "QUEUED")

            # Mark as executing
            connection = self.pool.get_connection()
            if connection:
                self.queue.mark_executing(operation_id)
                self.callbacks.push_update_to_stream(stream_id, operation_id, "EXECUTING")

                # Execute on MT5 (simplified - would call actual MT5 API)
                result = connection.place_order(symbol, 0, volume, price, stop_loss, take_profit)

                if result and result.get("status") == "SUCCESS":
                    latency_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)

                    self.queue.update_status(operation_id, "COMPLETED", result)
                    self.callbacks.push_update_to_stream(stream_id, operation_id, "COMPLETED", result)

                    self.logger.log_operation(
                        operation_id,
                        agent_id,
                        "PlaceOrder",
                        f"Symbol: {symbol}, Volume: {volume}",
                        f"Order executed: {result.get('order_id')}",
                        latency_ms,
                        success=True,
                        session_id=session_id,
                    )

                    return {"status": "SUCCESS", "operation_id": operation_id, "result": result}
                else:
                    error_code = "ORDER_FAILED"
                    error_msg = "Failed to place order"

                    self.queue.update_status(operation_id, "FAILED", None, error_code, error_msg)
                    self.callbacks.push_update_to_stream(
                        stream_id,
                        operation_id,
                        "FAILED",
                        {"error_code": error_code, "error_message": error_msg},
                    )

                    latency_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)
                    self.logger.log_error(
                        error_msg, error_code, agent_id, operation_id, session_id
                    )

                    return {"status": "FAILED", "operation_id": operation_id, "error": error_msg}
            else:
                error_code = "MT5_NOT_CONNECTED"
                error_msg = "MT5 terminal not connected"

                self.queue.update_status(operation_id, "FAILED", None, error_code, error_msg)
                self.callbacks.push_update_to_stream(
                    stream_id,
                    operation_id,
                    "FAILED",
                    {"error_code": error_code, "error_message": error_msg},
                )

                latency_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)
                self.logger.log_error(error_msg, error_code, agent_id, operation_id, session_id)

                return {"status": "FAILED", "operation_id": operation_id, "error": error_msg}

        except Exception as e:
            error_code = "INTERNAL_ERROR"
            latency_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)

            self.queue.update_status(operation_id, "FAILED", None, error_code, str(e))
            self.callbacks.push_update_to_stream(
                stream_id, operation_id, "FAILED", {"error_code": error_code, "error_message": str(e)}
            )

            self.logger.log_error(str(e), error_code, agent_id, operation_id, session_id)

            return {"status": "FAILED", "operation_id": operation_id, "error": str(e)}

    async def get_account_info(self, session_id: str, agent_id: str, stream_id: str) -> dict:
        """Get account information"""
        operation_id = str(uuid.uuid4())
        start_time = datetime.utcnow()

        try:
            connection = self.pool.get_connection()
            if connection:
                account_info = connection.get_account_info()

                if account_info:
                    latency_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)

                    self.callbacks.push_update_to_stream(stream_id, operation_id, "COMPLETED", account_info)

                    self.logger.log_operation(
                        operation_id,
                        agent_id,
                        "GetAccountInfo",
                        "Request account info",
                        f"Login: {account_info.get('login')}",
                        latency_ms,
                        success=True,
                        session_id=session_id,
                    )

                    return {"status": "SUCCESS", "data": account_info}
                else:
                    raise Exception("Failed to get account info")
            else:
                raise Exception("MT5 not connected")

        except Exception as e:
            error_code = "GET_INFO_FAILED"
            latency_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)

            self.callbacks.push_update_to_stream(
                stream_id,
                operation_id,
                "FAILED",
                {"error_code": error_code, "error_message": str(e)},
            )

            self.logger.log_error(str(e), error_code, agent_id, operation_id, session_id)

            return {"status": "FAILED", "error": str(e)}

    async def health_check(self) -> dict:
        """Health check endpoint"""
        return {
            "status": "HEALTHY" if self.pool.is_connected() else "UNHEALTHY",
            "pool": self.pool.health_check(),
            "callbacks": self.callbacks.health_check(),
        }
