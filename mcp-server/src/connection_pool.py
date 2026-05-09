"""MT5 connection pool with single connection constraint"""
import threading
import uuid
from typing import Optional
from datetime import datetime
from .mt5_adapter import MT5Adapter


class ConnectionPool:
    """Manages single MT5 connection and enforces connection constraint"""

    def __init__(self, terminal_path: str = ""):
        self.connection: Optional[MT5Adapter] = None
        self.connection_id = str(uuid.uuid4())
        self.lock = threading.RLock()
        self.is_active = False
        self.thread_id: Optional[int] = None
        self.connected_at: Optional[datetime] = None
        self.terminal_path = terminal_path
        self.operation_queue = []
        self.queue_lock = threading.Lock()

    def connect(self, login: int, password: str, server: str) -> bool:
        """Connect to MT5 (enforce single connection)"""
        with self.lock:
            if self.is_active:
                return True  # Already connected

            try:
                self.connection = MT5Adapter(self.terminal_path)
                if self.connection.initialize(login, password, server):
                    self.is_active = True
                    self.thread_id = threading.current_thread().ident
                    self.connected_at = datetime.utcnow()
                    return True
                else:
                    self.connection = None
                    return False
            except Exception as e:
                print(f"Connection error: {e}")
                self.connection = None
                self.is_active = False
                return False

    def disconnect(self) -> bool:
        """Disconnect from MT5"""
        with self.lock:
            if self.connection:
                result = self.connection.shutdown()
                self.connection = None
                self.is_active = False
                return result
            return True

    def get_connection(self) -> Optional[MT5Adapter]:
        """Get active connection (thread-safe)"""
        with self.lock:
            if self.is_active and self.connection:
                return self.connection
            return None

    def is_connected(self) -> bool:
        """Check if connected"""
        with self.lock:
            return self.is_active and self.connection is not None

    def queue_operation(self, operation: dict) -> str:
        """Queue operation for execution"""
        operation_id = str(uuid.uuid4())
        operation["id"] = operation_id
        operation["queued_at"] = datetime.utcnow().isoformat()

        with self.queue_lock:
            self.operation_queue.append(operation)

        return operation_id

    def get_queued_operations(self) -> list:
        """Get all queued operations"""
        with self.queue_lock:
            return self.operation_queue.copy()

    def remove_operation(self, operation_id: str) -> bool:
        """Remove operation from queue"""
        with self.queue_lock:
            self.operation_queue = [
                op for op in self.operation_queue if op.get("id") != operation_id
            ]
        return True

    def clear_queue(self) -> int:
        """Clear all queued operations"""
        with self.queue_lock:
            count = len(self.operation_queue)
            self.operation_queue.clear()
        return count

    def health_check(self) -> dict:
        """Health check for connection pool"""
        with self.lock:
            check = {
                "pool_id": self.connection_id,
                "is_connected": self.is_active,
                "thread_id": self.thread_id,
                "connected_at": self.connected_at.isoformat() if self.connected_at else None,
                "queued_operations": len(self.operation_queue),
            }

            if self.connection and self.is_active:
                mt5_health = self.connection.health_check()
                check["mt5_health"] = mt5_health

            return check


class PoolManager:
    """Manages connection pool lifecycle"""

    def __init__(self, terminal_path: str = ""):
        self.pool = ConnectionPool(terminal_path)
        self.lock = threading.RLock()

    def initialize(self, login: int, password: str, server: str) -> bool:
        """Initialize connection pool"""
        with self.lock:
            return self.pool.connect(login, password, server)

    def shutdown(self) -> bool:
        """Shutdown connection pool"""
        with self.lock:
            return self.pool.disconnect()

    def get_pool(self) -> ConnectionPool:
        """Get connection pool"""
        return self.pool

    def health_check(self) -> dict:
        """Get pool health status"""
        return self.pool.health_check()
