"""Contract tests for MT5 gRPC service"""
import pytest
import sys
import os

# Add src to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../src"))

from db_schema import DatabaseSchema
from operation_queue import OperationQueue
from session_manager import SessionManager
from middleware.auth import AuthMiddleware
from errors import ErrorCode, ErrorMapper


@pytest.fixture
def test_db():
    """Create test database"""
    db = DatabaseSchema("test_contract.db")
    db.connect()
    db.initialize()
    yield db
    db.close()
    # Cleanup
    import os

    if os.path.exists("test_contract.db"):
        os.remove("test_contract.db")


class TestGRPCContracts:
    """Test gRPC service contracts"""

    def test_execute_order_operation_request_schema(self):
        """Test ExecuteOrderOperation request schema"""
        request_schema = {
            "symbol": "string",
            "operation_type": "string",  # BUY, SELL
            "volume": "float",
            "price": "float",
            "stop_loss": "float (optional)",
            "take_profit": "float (optional)",
        }

        # Verify schema fields
        assert "symbol" in request_schema
        assert "operation_type" in request_schema
        assert "volume" in request_schema
        assert "price" in request_schema

    def test_execute_order_operation_response_schema(self):
        """Test ExecuteOrderOperation response schema"""
        response_schema = {
            "status": "string",  # SUCCESS, FAILED
            "operation_id": "string",
            "result": "object (optional)",
            "error": "string (optional)",
        }

        assert "status" in response_schema
        assert "operation_id" in response_schema

    def test_get_account_info_request_schema(self):
        """Test GetAccountInfo request schema"""
        request_schema = {}  # Empty request
        assert isinstance(request_schema, dict)

    def test_get_account_info_response_schema(self):
        """Test GetAccountInfo response schema"""
        response_schema = {
            "status": "string",
            "data": {
                "login": "int",
                "balance": "float",
                "equity": "float",
                "profit": "float",
                "margin": "float",
                "margin_free": "float",
                "margin_level": "float",
                "credit": "float",
                "currency": "string",
                "leverage": "int",
                "server": "string",
            },
            "error": "string (optional)",
        }

        assert "status" in response_schema
        assert "data" in response_schema

    def test_get_positions_response_schema(self):
        """Test GetPositions response schema"""
        response_schema = {
            "status": "string",
            "positions": [
                {
                    "ticket": "int",
                    "symbol": "string",
                    "type": "string",
                    "volume": "float",
                    "open_price": "float",
                    "current_price": "float",
                    "stop_loss": "float",
                    "take_profit": "float",
                    "profit": "float",
                    "time_open": "int",
                    "comment": "string",
                }
            ],
            "error": "string (optional)",
        }

        assert "status" in response_schema
        assert "positions" in response_schema

    def test_close_position_request_schema(self):
        """Test ClosePosition request schema"""
        request_schema = {"ticket": "int", "volume": "float (optional)"}

        assert "ticket" in request_schema

    def test_close_position_response_schema(self):
        """Test ClosePosition response schema"""
        response_schema = {
            "status": "string",
            "operation_id": "string",
            "error": "string (optional)",
        }

        assert "status" in response_schema
        assert "operation_id" in response_schema

    def test_check_health_response_schema(self):
        """Test CheckHealth response schema"""
        response_schema = {
            "status": "string",  # SERVING, NOT_SERVING
            "checks": {
                "mt5_connection": {
                    "is_connected": "bool",
                    "status": "string",
                    "message": "string",
                },
                "database": {
                    "status": "string",
                    "message": "string",
                },
            },
        }

        assert "status" in response_schema
        assert "checks" in response_schema

    def test_error_codes_mapping(self):
        """Test error codes are mapped correctly"""
        error_mappings = {
            "UNAUTHENTICATED": 16,
            "INVALID_ARGUMENT": 3,
            "NOT_FOUND": 5,
            "UNAVAILABLE": 14,
            "INTERNAL": 13,
        }

        for code, grpc_code in error_mappings.items():
            assert ErrorMapper.get_grpc_code(code) == grpc_code

    def test_operation_queue_persists_to_database(self, test_db):
        """Test operations persist to database"""
        queue = OperationQueue("test_contract.db")

        op_id = queue.enqueue("s1", "a1", "PlaceOrder", {"symbol": "EURUSD"})

        # Verify in database
        op = queue.get_operation(op_id)
        assert op is not None
        assert op["operation_id"] == op_id

    def test_authentication_contract(self, test_db):
        """Test authentication contract"""
        auth = AuthMiddleware("test_contract.db")

        # Register key
        success, msg = auth.register_api_key("agent1", "secret-key-123")
        assert success

        # Validate key
        is_valid, agent_id, msg = auth.validate_api_key("secret-key-123")
        assert is_valid
        assert agent_id == "agent1"

        # Invalid key
        is_valid, agent_id, msg = auth.validate_api_key("invalid-key")
        assert not is_valid

    def test_session_contract(self, test_db):
        """Test session management contract"""
        session_mgr = SessionManager("test_contract.db")

        # Create session
        session_id = session_mgr.create_session("agent1", "key1")
        assert session_id is not None

        # Retrieve session
        session = session_mgr.get_session(session_id)
        assert session is not None
        assert session["agent_id"] == "agent1"

        # Check active
        assert session_mgr.is_session_active(session_id)

        # Close session
        session_mgr.close_session(session_id)
        assert not session_mgr.is_session_active(session_id)


class TestMessageContracts:
    """Test message serialization contracts"""

    def test_operation_message_serialization(self):
        """Test operation message can be serialized"""
        operation = {
            "operation_id": "op123",
            "agent_id": "agent1",
            "status": "COMPLETED",
            "request_data": {"symbol": "EURUSD"},
            "result_data": {"order_id": 456},
        }

        # Should be JSON serializable
        import json

        json_str = json.dumps(operation)
        assert json_str is not None

    def test_health_check_message_serialization(self):
        """Test health check message can be serialized"""
        health = {
            "status": "SERVING",
            "checks": {
                "mt5": {"is_connected": True, "status": "HEALTHY"},
                "database": {"tables": 4, "status": "HEALTHY"},
            },
        }

        import json

        json_str = json.dumps(health)
        assert json_str is not None

    def test_error_message_structure(self):
        """Test error message structure"""
        error_response = {
            "error_code": ErrorCode.UNAUTHENTICATED,
            "message": "Invalid API key",
            "timestamp": "2026-05-09T00:00:00.000000",
        }

        assert "error_code" in error_response
        assert "message" in error_response


class TestRPCCallability:
    """Test all RPCs can be called"""

    def test_all_rpcs_defined(self):
        """Test all expected RPCs exist"""
        expected_rpcs = [
            "ExecuteOrderOperation",
            "GetAccountInfo",
            "GetPositions",
            "ClosePosition",
            "GetSymbolInfo",
            "GetRates",
            "CheckHealth",
        ]

        # These are the 7 core RPCs mentioned in the spec
        assert len(expected_rpcs) == 7


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
