"""Health check endpoint for MT5 gRPC service"""
from enum import Enum
from typing import Dict, Any, Optional
from datetime import datetime
from .connection_pool import ConnectionPool
from .operation_queue import OperationQueue
from .callback_manager import CallbackManager
from .session_manager import SessionManager


class HealthStatus(str, Enum):
    """Health status"""

    SERVING = "SERVING"
    NOT_SERVING = "NOT_SERVING"
    UNKNOWN = "UNKNOWN"


class HealthCheck:
    """Health check for gRPC service"""

    def __init__(
        self,
        pool: ConnectionPool,
        queue: OperationQueue,
        callbacks: CallbackManager,
        session_manager: SessionManager,
    ):
        self.pool = pool
        self.queue = queue
        self.callbacks = callbacks
        self.session_manager = session_manager

    def check(self) -> Dict[str, Any]:
        """Perform health check"""
        is_serving = self._is_serving()

        return {
            "status": HealthStatus.SERVING if is_serving else HealthStatus.NOT_SERVING,
            "timestamp": datetime.utcnow().isoformat(),
            "checks": {
                "mt5_connection": self._check_mt5_connection(),
                "database": self._check_database(),
                "queues": self._check_queues(),
                "sessions": self._check_sessions(),
                "callbacks": self._check_callbacks(),
            },
            "summary": self._get_summary(is_serving),
        }

    def _check_mt5_connection(self) -> Dict[str, Any]:
        """Check MT5 connection health"""
        is_connected = self.pool.is_connected()

        return {
            "is_connected": is_connected,
            "status": "HEALTHY" if is_connected else "UNHEALTHY",
            "pool": self.pool.health_check() if is_connected else {},
            "message": "MT5 connection active" if is_connected else "MT5 connection unavailable",
        }

    def _check_database(self) -> Dict[str, Any]:
        """Check database health"""
        try:
            import sqlite3

            conn = sqlite3.connect(self.queue.db_path)
            cursor = conn.cursor()

            # Check if tables exist
            cursor.execute(
                """
                SELECT COUNT(*) FROM sqlite_master
                WHERE type='table' AND name IN
                ('queued_operations', 'operation_logs', 'api_keys', 'agent_sessions')
                """
            )

            table_count = cursor.fetchone()[0]
            conn.close()

            is_healthy = table_count == 4

            return {
                "status": "HEALTHY" if is_healthy else "UNHEALTHY",
                "tables_found": table_count,
                "expected_tables": 4,
                "message": "Database schema intact" if is_healthy else "Database schema incomplete",
            }

        except Exception as e:
            return {
                "status": "UNHEALTHY",
                "error": str(e),
                "message": f"Database error: {e}",
            }

    def _check_queues(self) -> Dict[str, Any]:
        """Check operation queues"""
        try:
            queued_ops = self.queue.get_queued(limit=1)
            queue_size_estimate = len(self.queue.get_queued(limit=1000))

            return {
                "status": "HEALTHY",
                "queued_operations_estimate": queue_size_estimate,
                "message": f"Queue has ~{queue_size_estimate} pending operations",
            }

        except Exception as e:
            return {
                "status": "UNHEALTHY",
                "error": str(e),
                "message": f"Queue error: {e}",
            }

    def _check_sessions(self) -> Dict[str, Any]:
        """Check session management"""
        try:
            stats = self.session_manager.health_check()

            return {
                "status": "HEALTHY",
                "total_sessions": stats.get("total_sessions", 0),
                "active_sessions": stats.get("active_sessions", 0),
                "timeout_minutes": stats.get("session_timeout_minutes", 30),
                "message": f"{stats.get('active_sessions', 0)} active sessions",
            }

        except Exception as e:
            return {
                "status": "UNHEALTHY",
                "error": str(e),
                "message": f"Session manager error: {e}",
            }

    def _check_callbacks(self) -> Dict[str, Any]:
        """Check callback streams"""
        try:
            stats = self.callbacks.health_check()

            return {
                "status": "HEALTHY",
                "total_streams": stats.get("total_streams", 0),
                "active_sessions": stats.get("active_sessions", 0),
                "message": f"{stats.get('total_streams', 0)} callback streams active",
            }

        except Exception as e:
            return {
                "status": "UNHEALTHY",
                "error": str(e),
                "message": f"Callback manager error: {e}",
            }

    def _is_serving(self) -> bool:
        """Determine if service is ready to serve requests"""
        checks = {
            "mt5": self._check_mt5_connection(),
            "database": self._check_database(),
        }

        # Service is serving if MT5 is connected AND database is healthy
        return (
            checks["mt5"].get("status") == "HEALTHY"
            and checks["database"].get("status") == "HEALTHY"
        )

    def _get_summary(self, is_serving: bool) -> str:
        """Get human-readable summary"""
        if is_serving:
            return "Service is healthy and ready to serve requests"
        else:
            return "Service is degraded or unavailable"


class HealthCheckServer:
    """Standalone health check server"""

    def __init__(self, health_check: HealthCheck, port: int = 50052):
        self.health_check = health_check
        self.port = port

    async def handle_health_check(self) -> Dict[str, Any]:
        """Handle health check request"""
        return self.health_check.check()

    def get_status(self) -> str:
        """Get simple status string"""
        check = self.health_check.check()
        return check["status"]

    def is_ready(self) -> bool:
        """Check if service is ready"""
        check = self.health_check.check()
        return check["status"] == HealthStatus.SERVING
