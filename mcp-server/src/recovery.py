"""Operation recovery on server restart"""
import sqlite3
import json
import asyncio
from typing import List, Dict, Any
from datetime import datetime
from .operation_queue import OperationQueue, RetryStrategy
from .logging_config import StructuredLogger, setup_logging


class RecoveryManager:
    """Manages recovery of queued operations on server startup"""

    def __init__(self, operation_queue: OperationQueue, logger: Optional[StructuredLogger] = None):
        self.operation_queue = operation_queue
        self.logger = logger or StructuredLogger(setup_logging())
        self.retry_strategy = RetryStrategy()

    def recover_operations(self) -> List[Dict[str, Any]]:
        """Load and recover queued operations from database"""
        try:
            # Get all queued operations
            queued_ops = self.operation_queue.get_queued(limit=1000)

            if not queued_ops:
                self.logger.logger.info("No queued operations to recover")
                return []

            self.logger.logger.info(f"Recovering {len(queued_ops)} operations from database")

            # Categorize operations by retry count
            recoverable = []
            unrecoverable = []

            for op in queued_ops:
                retry_count = op.get("retry_count", 0)
                max_retries = op.get("max_retries", 3)

                if retry_count < max_retries:
                    recoverable.append(op)
                else:
                    unrecoverable.append(op)
                    # Mark as permanently failed
                    self.operation_queue.update_status(
                        op["operation_id"],
                        "FAILED",
                        None,
                        "MAX_RETRIES_EXCEEDED",
                        f"Operation exceeded max retries ({max_retries})",
                    )

            self.logger.logger.info(
                f"Recovered {len(recoverable)} operations, "
                f"marked {len(unrecoverable)} as permanently failed"
            )

            return recoverable

        except Exception as e:
            self.logger.logger.error(f"Error during recovery: {e}")
            return []

    def schedule_recovery_execution(self, operations: List[Dict[str, Any]]) -> None:
        """Schedule recovered operations for execution"""
        for op in operations:
            operation_id = op["operation_id"]
            retry_count = op.get("retry_count", 0)

            # Calculate backoff delay
            delay = self.retry_strategy.get_delay(retry_count)

            self.logger.logger.info(
                f"Scheduling operation {operation_id} for retry "
                f"(attempt {retry_count + 1}, delay {delay}s)"
            )

    async def process_recovered_operations(
        self, operations: List[Dict[str, Any]], execution_handler
    ) -> Dict[str, Any]:
        """Process recovered operations asynchronously"""
        results = {"total": len(operations), "successful": 0, "failed": 0, "scheduled": 0}

        for op in operations:
            operation_id = op["operation_id"]
            retry_count = op.get("retry_count", 0)

            try:
                # Get retry delay
                delay = self.retry_strategy.get_delay(retry_count)

                if delay < 0:
                    # Max retries exceeded
                    results["failed"] += 1
                    continue

                # Wait before retry (exponential backoff)
                if retry_count > 0:
                    await asyncio.sleep(delay)

                # Execute operation
                self.logger.logger.info(f"Executing recovered operation: {operation_id}")

                # Call execution handler
                if execution_handler:
                    success = await execution_handler(op)

                    if success:
                        results["successful"] += 1
                        self.logger.logger.info(f"Operation {operation_id} recovered successfully")
                    else:
                        results["failed"] += 1
                        # Increment retry count for next attempt
                        self.operation_queue.increment_retry(operation_id)
                        self.logger.logger.warning(f"Operation {operation_id} failed, will retry later")
                else:
                    results["scheduled"] += 1

            except Exception as e:
                results["failed"] += 1
                self.logger.logger.error(f"Error processing operation {operation_id}: {e}")
                self.operation_queue.increment_retry(operation_id)

        return results

    def get_recovery_status(self) -> Dict[str, Any]:
        """Get status of queued operations ready for recovery"""
        try:
            conn = sqlite3.connect(self.operation_queue.db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()

            # Count operations by status
            cursor.execute(
                """
                SELECT status, COUNT(*) as count FROM queued_operations
                GROUP BY status
                """
            )

            status_counts = {row["status"]: row["count"] for row in cursor.fetchall()}

            # Get operations by retry count
            cursor.execute(
                """
                SELECT retry_count, COUNT(*) as count FROM queued_operations
                WHERE status = 'QUEUED'
                GROUP BY retry_count
                ORDER BY retry_count ASC
                """
            )

            retry_counts = {f"retry_{row['retry_count']}": row["count"] for row in cursor.fetchall()}

            conn.close()

            return {
                "by_status": status_counts,
                "by_retry_count": retry_counts,
                "total_queued": status_counts.get("QUEUED", 0),
            }

        except Exception as e:
            self.logger.logger.error(f"Error getting recovery status: {e}")
            return {}


class RecoveryWorker:
    """Background worker for ongoing operation recovery"""

    def __init__(self, recovery_manager: RecoveryManager, execution_handler=None):
        self.recovery_manager = recovery_manager
        self.execution_handler = execution_handler
        self.running = False

    async def start(self, check_interval_seconds: int = 60) -> None:
        """Start recovery worker"""
        self.running = True

        while self.running:
            try:
                # Check for operations that need recovery
                operations = self.recovery_manager.recover_operations()

                if operations:
                    # Process them
                    results = await self.recovery_manager.process_recovered_operations(
                        operations, self.execution_handler
                    )

                    self.recovery_manager.logger.logger.info(f"Recovery cycle results: {results}")

            except Exception as e:
                self.recovery_manager.logger.logger.error(f"Recovery worker error: {e}")

            # Wait before next check
            await asyncio.sleep(check_interval_seconds)

    def stop(self) -> None:
        """Stop recovery worker"""
        self.running = False


# Import Optional for type hints
from typing import Optional
