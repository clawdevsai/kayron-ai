"""Audit logger for operation tracking"""
import sqlite3
import uuid
import json
from datetime import datetime
from typing import Optional, Dict, Any
from .logging_config import StructuredLogger, setup_logging


class AuditLogger:
    """Logs all operations to database and structured logs"""

    def __init__(self, db_path: str = "mcp-server.db"):
        self.db_path = db_path
        self.logger = StructuredLogger(setup_logging())

    def log_operation(
        self,
        agent_id: str,
        operation_type: str,
        request_summary: str,
        result_summary: Optional[str] = None,
        latency_ms: int = 0,
        success: bool = True,
        error_code: Optional[str] = None,
        session_id: Optional[str] = None,
        operation_id: Optional[str] = None,
    ) -> str:
        """Log operation to database and structured logs"""
        log_id = str(uuid.uuid4())

        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                INSERT INTO operation_logs
                (log_id, session_id, agent_id, operation_type, operation_id,
                 request_summary, result_summary, latency_ms, success,
                 error_code, timestamp)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    log_id,
                    session_id,
                    agent_id,
                    operation_type,
                    operation_id,
                    request_summary,
                    result_summary or "",
                    latency_ms,
                    1 if success else 0,
                    error_code,
                    datetime.utcnow().isoformat(),
                ),
            )

            conn.commit()
            conn.close()

            # Also log to structured logger
            self.logger.log_operation(
                operation_id or log_id,
                agent_id,
                operation_type,
                request_summary,
                result_summary or "N/A",
                latency_ms,
                success,
                error_code,
                session_id,
            )

            return log_id

        except Exception as e:
            self.logger.logger.error(f"Error logging operation: {e}")
            return log_id

    def log_authentication(
        self, agent_id: str, session_id: str, success: bool, reason: Optional[str] = None
    ) -> str:
        """Log authentication event"""
        log_id = str(uuid.uuid4())

        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            result_summary = f"Authentication {'succeeded' if success else 'failed'}"
            if reason:
                result_summary += f": {reason}"

            cursor.execute(
                """
                INSERT INTO operation_logs
                (log_id, session_id, agent_id, operation_type, operation_id,
                 request_summary, result_summary, latency_ms, success,
                 error_code, timestamp)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    log_id,
                    session_id,
                    agent_id,
                    "Authentication",
                    None,
                    "API key validation",
                    result_summary,
                    0,
                    1 if success else 0,
                    "UNAUTHENTICATED" if not success else None,
                    datetime.utcnow().isoformat(),
                ),
            )

            conn.commit()
            conn.close()

            level = "INFO" if success else "WARNING"
            self.logger.logger.log(
                getattr(__import__("logging"), level),
                f"Authentication event: agent={agent_id}, success={success}",
            )

            return log_id

        except Exception as e:
            self.logger.logger.error(f"Error logging authentication: {e}")
            return log_id

    def get_operation_logs(
        self,
        agent_id: Optional[str] = None,
        operation_type: Optional[str] = None,
        limit: int = 100,
        offset: int = 0,
    ) -> list:
        """Retrieve operation logs"""
        try:
            conn = sqlite3.connect(self.db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()

            query = "SELECT * FROM operation_logs WHERE 1=1"
            params = []

            if agent_id:
                query += " AND agent_id = ?"
                params.append(agent_id)

            if operation_type:
                query += " AND operation_type = ?"
                params.append(operation_type)

            query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
            params.extend([limit, offset])

            cursor.execute(query, params)

            rows = cursor.fetchall()
            conn.close()

            return [dict(row) for row in rows]

        except Exception as e:
            self.logger.logger.error(f"Error retrieving logs: {e}")
            return []

    def get_agent_statistics(self, agent_id: str) -> Dict[str, Any]:
        """Get statistics for an agent"""
        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            # Total operations
            cursor.execute("SELECT COUNT(*) as count FROM operation_logs WHERE agent_id = ?", (agent_id,))
            total = cursor.fetchone()[0]

            # Successful operations
            cursor.execute(
                "SELECT COUNT(*) as count FROM operation_logs WHERE agent_id = ? AND success = 1",
                (agent_id,),
            )
            successful = cursor.fetchone()[0]

            # Failed operations
            failed = total - successful

            # Average latency
            cursor.execute(
                "SELECT AVG(latency_ms) as avg_latency FROM operation_logs WHERE agent_id = ? AND success = 1",
                (agent_id,),
            )
            avg_latency = cursor.fetchone()[0] or 0

            # Operations by type
            cursor.execute(
                """
                SELECT operation_type, COUNT(*) as count
                FROM operation_logs
                WHERE agent_id = ?
                GROUP BY operation_type
                """,
                (agent_id,),
            )

            by_type = {row[0]: row[1] for row in cursor.fetchall()}

            # Most common errors
            cursor.execute(
                """
                SELECT error_code, COUNT(*) as count
                FROM operation_logs
                WHERE agent_id = ? AND error_code IS NOT NULL
                GROUP BY error_code
                ORDER BY count DESC
                LIMIT 5
                """,
                (agent_id,),
            )

            errors = {row[0]: row[1] for row in cursor.fetchall()}

            conn.close()

            return {
                "agent_id": agent_id,
                "total_operations": total,
                "successful": successful,
                "failed": failed,
                "success_rate": (successful / total * 100) if total > 0 else 0,
                "average_latency_ms": round(avg_latency, 2),
                "by_operation_type": by_type,
                "common_errors": errors,
            }

        except Exception as e:
            self.logger.logger.error(f"Error getting statistics: {e}")
            return {}

    def cleanup_old_logs(self, days: int = 30) -> int:
        """Delete operation logs older than N days"""
        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                DELETE FROM operation_logs
                WHERE timestamp < datetime('now', '-' || ? || ' days')
                """,
                (days,),
            )

            deleted = cursor.rowcount
            conn.commit()
            conn.close()

            self.logger.logger.info(f"Deleted {deleted} operation logs older than {days} days")

            return deleted

        except Exception as e:
            self.logger.logger.error(f"Error cleaning up logs: {e}")
            return 0
