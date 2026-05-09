"""Operation queue with SQLite persistence and retry logic"""
import sqlite3
import json
import uuid
import time
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum
import threading


class RetryStrategy:
    """Exponential backoff retry strategy"""

    def __init__(self, max_retries: int = 3, base_delay: float = 1.0):
        self.max_retries = max_retries
        self.base_delay = base_delay

    def get_delay(self, retry_count: int) -> float:
        """Calculate exponential backoff delay"""
        if retry_count >= self.max_retries:
            return -1  # Max retries exceeded
        return self.base_delay * (2 ** retry_count)

    def should_retry(self, retry_count: int) -> bool:
        """Check if operation should be retried"""
        return retry_count < self.max_retries


class OperationQueue:
    """Queue for operations with SQLite persistence"""

    def __init__(self, db_path: str = "mcp-server.db"):
        self.db_path = db_path
        self.lock = threading.RLock()
        self.retry_strategy = RetryStrategy()

    def enqueue(
        self,
        session_id: str,
        agent_id: str,
        operation_type: str,
        request_data: dict,
    ) -> str:
        """Add operation to queue"""
        operation_id = str(uuid.uuid4())

        with self.lock:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                INSERT INTO queued_operations
                (operation_id, session_id, agent_id, operation_type, status, request_data, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    operation_id,
                    session_id,
                    agent_id,
                    operation_type,
                    "QUEUED",
                    json.dumps(request_data),
                    datetime.utcnow().isoformat(),
                ),
            )

            conn.commit()
            conn.close()

        return operation_id

    def get_queued(self, limit: int = 10) -> List[Dict[str, Any]]:
        """Get queued operations ready for execution"""
        with self.lock:
            conn = sqlite3.connect(self.db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()

            cursor.execute(
                """
                SELECT * FROM queued_operations
                WHERE status = 'QUEUED'
                ORDER BY created_at ASC
                LIMIT ?
                """,
                (limit,),
            )

            rows = cursor.fetchall()
            conn.close()

            operations = []
            for row in rows:
                operations.append(
                    {
                        "operation_id": row["operation_id"],
                        "session_id": row["session_id"],
                        "agent_id": row["agent_id"],
                        "operation_type": row["operation_type"],
                        "request_data": json.loads(row["request_data"]),
                        "retry_count": row["retry_count"],
                        "max_retries": row["max_retries"],
                    }
                )

            return operations

    def update_status(
        self,
        operation_id: str,
        status: str,
        result_data: Optional[dict] = None,
        error_code: Optional[str] = None,
        error_message: Optional[str] = None,
    ) -> bool:
        """Update operation status"""
        with self.lock:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                UPDATE queued_operations
                SET status = ?, result_data = ?, error_code = ?, error_message = ?, completed_at = ?
                WHERE operation_id = ?
                """,
                (
                    status,
                    json.dumps(result_data) if result_data else None,
                    error_code,
                    error_message,
                    datetime.utcnow().isoformat() if status in ["COMPLETED", "FAILED"] else None,
                    operation_id,
                ),
            )

            conn.commit()
            conn.close()

        return True

    def mark_executing(self, operation_id: str) -> bool:
        """Mark operation as executing"""
        with self.lock:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                UPDATE queued_operations
                SET status = 'EXECUTING', started_at = ?
                WHERE operation_id = ?
                """,
                (datetime.utcnow().isoformat(), operation_id),
            )

            conn.commit()
            conn.close()

        return True

    def increment_retry(self, operation_id: str) -> int:
        """Increment retry count for failed operation"""
        with self.lock:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                UPDATE queued_operations
                SET retry_count = retry_count + 1, status = 'QUEUED'
                WHERE operation_id = ?
                """,
                (operation_id,),
            )

            cursor.execute("SELECT retry_count FROM queued_operations WHERE operation_id = ?", (operation_id,))
            result = cursor.fetchone()
            conn.commit()
            conn.close()

        return result[0] if result else -1

    def get_operation(self, operation_id: str) -> Optional[Dict[str, Any]]:
        """Get specific operation"""
        with self.lock:
            conn = sqlite3.connect(self.db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()

            cursor.execute("SELECT * FROM queued_operations WHERE operation_id = ?", (operation_id,))

            row = cursor.fetchone()
            conn.close()

            if row:
                return {
                    "operation_id": row["operation_id"],
                    "session_id": row["session_id"],
                    "agent_id": row["agent_id"],
                    "operation_type": row["operation_type"],
                    "status": row["status"],
                    "request_data": json.loads(row["request_data"]),
                    "result_data": json.loads(row["result_data"]) if row["result_data"] else None,
                    "error_code": row["error_code"],
                    "error_message": row["error_message"],
                    "retry_count": row["retry_count"],
                    "max_retries": row["max_retries"],
                }

            return None

    def cleanup_completed(self, days: int = 7) -> int:
        """Clean up completed operations older than N days"""
        with self.lock:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            # Delete completed operations older than N days
            cursor.execute(
                """
                DELETE FROM queued_operations
                WHERE status IN ('COMPLETED', 'FAILED')
                AND completed_at < datetime('now', '-' || ? || ' days')
                """,
                (days,),
            )

            deleted = cursor.rowcount
            conn.commit()
            conn.close()

        return deleted
