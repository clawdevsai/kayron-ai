"""End-to-end integration tests for MT5 gRPC service"""
import pytest
import asyncio
from unittest.mock import Mock, patch, AsyncMock
import sys
import os

# Add src to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../src"))

from operation_queue import OperationQueue
from callback_manager import CallbackManager
from session_manager import SessionManager
from service import MT5Service
from connection_pool import ConnectionPool
from db_schema import DatabaseSchema


@pytest.fixture
def test_db():
    """Create test database"""
    db = DatabaseSchema("test_mcp.db")
    db.connect()
    db.initialize()
    yield db
    db.close()
    # Cleanup
    import os

    if os.path.exists("test_mcp.db"):
        os.remove("test_mcp.db")


@pytest.fixture
def operation_queue(test_db):
    """Create operation queue"""
    return OperationQueue("test_mcp.db")


@pytest.fixture
def callback_manager():
    """Create callback manager"""
    return CallbackManager()


@pytest.fixture
def session_manager(test_db):
    """Create session manager"""
    manager = SessionManager("test_mcp.db")
    manager.load_sessions_from_db()
    return manager


@pytest.fixture
def connection_pool():
    """Create connection pool"""
    return ConnectionPool()


@pytest.fixture
def service(connection_pool, operation_queue, callback_manager):
    """Create MT5 service"""
    return MT5Service(connection_pool, operation_queue, callback_manager)


class TestServiceE2E:
    """End-to-end service tests"""

    @pytest.mark.asyncio
    async def test_health_check_when_disconnected(self, service):
        """Test health check when MT5 is disconnected"""
        health = await service.health_check()
        assert health["status"] == "UNHEALTHY"
        assert not service.pool.is_connected()

    @pytest.mark.asyncio
    async def test_get_account_info_without_connection(self, service):
        """Test get account info fails when disconnected"""
        result = await service.get_account_info("session1", "agent1", "stream1")
        assert result["status"] == "FAILED"

    @pytest.mark.asyncio
    async def test_execute_order_without_connection(self, service):
        """Test order execution fails when disconnected"""
        result = await service.execute_order_operation(
            "session1",
            "agent1",
            "stream1",
            "EURUSD",
            "BUY",
            1.0,
            1.0950,
        )
        assert result["status"] == "FAILED"

    def test_operation_queueing(self, operation_queue):
        """Test operation queueing"""
        op_id = operation_queue.enqueue(
            "session1",
            "agent1",
            "PlaceOrder",
            {"symbol": "EURUSD", "volume": 1.0},
        )

        assert op_id is not None
        queued = operation_queue.get_queued()
        assert len(queued) > 0
        assert queued[0]["operation_id"] == op_id

    def test_operation_status_update(self, operation_queue):
        """Test operation status updates"""
        op_id = operation_queue.enqueue(
            "session1",
            "agent1",
            "PlaceOrder",
            {"symbol": "EURUSD"},
        )

        # Update status
        operation_queue.update_status(
            op_id, "COMPLETED", {"order_id": 12345}, None, None
        )

        op = operation_queue.get_operation(op_id)
        assert op["status"] == "COMPLETED"
        assert op["result_data"]["order_id"] == 12345

    def test_callback_stream_creation(self, callback_manager):
        """Test callback stream creation"""
        stream_id = callback_manager.create_stream("session1", "agent1")
        assert stream_id is not None

        stream = callback_manager.get_stream(stream_id)
        assert stream is not None
        assert stream.is_active

    def test_callback_stream_update(self, callback_manager):
        """Test pushing updates to callback stream"""
        stream_id = callback_manager.create_stream("session1", "agent1")
        stream = callback_manager.get_stream(stream_id)

        updates = []

        def capture_update(update):
            updates.append(update)

        stream.register_callback(capture_update)

        # Push update
        stream.push_update("op1", "QUEUED")

        assert len(updates) == 1
        assert updates[0]["status"] == "QUEUED"

    def test_session_creation(self, session_manager):
        """Test session creation"""
        session_id = session_manager.create_session("agent1", "key1")
        assert session_id is not None

        session = session_manager.get_session(session_id)
        assert session is not None
        assert session["is_active"]

    def test_session_activity_tracking(self, session_manager):
        """Test session activity tracking"""
        session_id = session_manager.create_session("agent1", "key1")
        session = session_manager.get_session(session_id)

        original_activity = session["last_activity_at"]

        # Update activity
        session_manager.update_activity(session_id)
        session = session_manager.get_session(session_id)

        assert session["last_activity_at"] > original_activity

    def test_retry_logic(self, operation_queue):
        """Test operation retry logic"""
        op_id = operation_queue.enqueue(
            "session1",
            "agent1",
            "PlaceOrder",
            {"symbol": "EURUSD"},
        )

        # Mark as executing then failed
        operation_queue.mark_executing(op_id)
        operation_queue.update_status(op_id, "FAILED", None, "ORDER_FAILED", "error")

        # Increment retry
        retry_count = operation_queue.increment_retry(op_id)

        op = operation_queue.get_operation(op_id)
        assert op["status"] == "QUEUED"  # Back to queued for retry
        assert op["retry_count"] == 1

    def test_operation_cleanup(self, operation_queue):
        """Test operation cleanup"""
        op_id = operation_queue.enqueue(
            "session1", "agent1", "PlaceOrder", {"symbol": "EURUSD"}
        )

        operation_queue.mark_executing(op_id)
        operation_queue.update_status(op_id, "COMPLETED")

        # Cleanup should not affect recent operations
        deleted = operation_queue.cleanup_completed(days=7)
        # Should be 0 since operation was just created


class TestServiceIntegration:
    """Integration tests for service components"""

    @pytest.mark.asyncio
    async def test_full_operation_lifecycle(self, service, operation_queue):
        """Test complete operation lifecycle"""
        # Queue operation
        op_id = operation_queue.enqueue(
            "session1",
            "agent1",
            "PlaceOrder",
            {"symbol": "EURUSD", "volume": 1.0},
        )

        # Get queued operations
        queued = operation_queue.get_queued()
        assert len(queued) > 0

        # Update to executing
        operation_queue.mark_executing(op_id)

        # Complete operation
        operation_queue.update_status(
            op_id, "COMPLETED", {"order_id": 123, "price": 1.0950}
        )

        # Verify completion
        op = operation_queue.get_operation(op_id)
        assert op["status"] == "COMPLETED"
        assert op["result_data"]["order_id"] == 123


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
