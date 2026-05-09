"""Custom exceptions and error mapping for gRPC service"""
from enum import Enum


class ErrorCode(str, Enum):
    """gRPC error codes"""

    OK = "OK"
    UNAUTHENTICATED = "UNAUTHENTICATED"
    PERMISSION_DENIED = "PERMISSION_DENIED"
    INVALID_ARGUMENT = "INVALID_ARGUMENT"
    NOT_FOUND = "NOT_FOUND"
    ALREADY_EXISTS = "ALREADY_EXISTS"
    ABORTED = "ABORTED"
    UNAVAILABLE = "UNAVAILABLE"
    INTERNAL = "INTERNAL"
    UNKNOWN = "UNKNOWN"


class MT5Error(Exception):
    """Base exception for MT5 errors"""

    def __init__(self, message: str, error_code: str = ErrorCode.INTERNAL):
        self.message = message
        self.error_code = error_code
        super().__init__(self.message)


class AuthenticationError(MT5Error):
    """Authentication failure"""

    def __init__(self, message: str = "Authentication failed"):
        super().__init__(message, ErrorCode.UNAUTHENTICATED)


class ValidationError(MT5Error):
    """Request validation error"""

    def __init__(self, message: str = "Invalid argument"):
        super().__init__(message, ErrorCode.INVALID_ARGUMENT)


class ConnectionError(MT5Error):
    """MT5 connection error"""

    def __init__(self, message: str = "MT5 connection unavailable"):
        super().__init__(message, ErrorCode.UNAVAILABLE)


class OperationError(MT5Error):
    """Operation execution error"""

    def __init__(self, message: str = "Operation failed"):
        super().__init__(message, ErrorCode.INTERNAL)


class SessionError(MT5Error):
    """Session management error"""

    def __init__(self, message: str = "Session error"):
        super().__init__(message, ErrorCode.ABORTED)


class TimeoutError(MT5Error):
    """Operation timeout"""

    def __init__(self, message: str = "Operation timeout"):
        super().__init__(message, ErrorCode.UNAVAILABLE)


class DuplicateError(MT5Error):
    """Duplicate resource error"""

    def __init__(self, message: str = "Resource already exists"):
        super().__init__(message, ErrorCode.ALREADY_EXISTS)


class NotFoundError(MT5Error):
    """Resource not found error"""

    def __init__(self, message: str = "Resource not found"):
        super().__init__(message, ErrorCode.NOT_FOUND)


class ErrorMapper:
    """Map exceptions to gRPC error responses"""

    ERROR_MAPPING = {
        AuthenticationError: ("UNAUTHENTICATED", "Authentication failed"),
        ValidationError: ("INVALID_ARGUMENT", "Invalid request"),
        ConnectionError: ("UNAVAILABLE", "MT5 service unavailable"),
        OperationError: ("INTERNAL", "Operation failed"),
        SessionError: ("ABORTED", "Session error"),
        TimeoutError: ("UNAVAILABLE", "Operation timeout"),
        DuplicateError: ("ALREADY_EXISTS", "Resource already exists"),
        NotFoundError: ("NOT_FOUND", "Resource not found"),
    }

    @staticmethod
    def to_error_response(exception: Exception) -> dict:
        """Convert exception to error response"""
        if isinstance(exception, MT5Error):
            return {
                "error_code": exception.error_code,
                "message": exception.message,
            }

        # Check mapping
        for exc_type, (code, message) in ErrorMapper.ERROR_MAPPING.items():
            if isinstance(exception, exc_type):
                return {
                    "error_code": code,
                    "message": f"{message}: {str(exception)}",
                }

        # Default error
        return {
            "error_code": ErrorCode.INTERNAL,
            "message": f"Unexpected error: {str(exception)}",
        }

    @staticmethod
    def get_grpc_code(error_code: str) -> int:
        """Get gRPC status code from error code"""
        grpc_codes = {
            "OK": 0,
            "CANCELLED": 1,
            "UNKNOWN": 2,
            "INVALID_ARGUMENT": 3,
            "DEADLINE_EXCEEDED": 4,
            "NOT_FOUND": 5,
            "ALREADY_EXISTS": 6,
            "PERMISSION_DENIED": 7,
            "RESOURCE_EXHAUSTED": 8,
            "FAILED_PRECONDITION": 9,
            "ABORTED": 10,
            "OUT_OF_RANGE": 11,
            "UNIMPLEMENTED": 12,
            "INTERNAL": 13,
            "UNAVAILABLE": 14,
            "DATA_LOSS": 15,
            "UNAUTHENTICATED": 16,
        }
        return grpc_codes.get(error_code, 2)  # Default to UNKNOWN


class ErrorHandler:
    """Global error handler for service"""

    def __init__(self, logger):
        self.logger = logger

    def handle(self, exception: Exception, context: dict = None) -> dict:
        """Handle exception and log"""
        error_response = ErrorMapper.to_error_response(exception)

        # Log error
        if context:
            self.logger.log_error(
                error_response["message"],
                error_response["error_code"],
                agent_id=context.get("agent_id"),
                operation_id=context.get("operation_id"),
                session_id=context.get("session_id"),
            )

        return error_response
