"""Structured JSON logging configuration"""
import logging
import json
from datetime import datetime
from typing import Optional


class JSONFormatter(logging.Formatter):
    """Custom JSON formatter for structured logging"""

    def format(self, record: logging.LogRecord) -> str:
        log_data = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }

        # Add exception info if present
        if record.exc_info:
            log_data["exception"] = self.formatException(record.exc_info)

        # Add extra fields from logging context
        if hasattr(record, "session_id"):
            log_data["session_id"] = record.session_id
        if hasattr(record, "agent_id"):
            log_data["agent_id"] = record.agent_id
        if hasattr(record, "operation_id"):
            log_data["operation_id"] = record.operation_id
        if hasattr(record, "operation_type"):
            log_data["operation_type"] = record.operation_type
        if hasattr(record, "latency_ms"):
            log_data["latency_ms"] = record.latency_ms
        if hasattr(record, "error_code"):
            log_data["error_code"] = record.error_code

        return json.dumps(log_data)


def setup_logging(
    log_level: str = "INFO",
    log_file: Optional[str] = None,
    format_type: str = "json",
) -> logging.Logger:
    """Configure logging with JSON format"""
    logger = logging.getLogger("mt5-grpc")
    logger.setLevel(getattr(logging, log_level))

    # Remove existing handlers
    for handler in logger.handlers[:]:
        logger.removeHandler(handler)

    # Console handler
    console_handler = logging.StreamHandler()
    console_handler.setLevel(getattr(logging, log_level))

    if format_type == "json":
        formatter = JSONFormatter()
    else:
        formatter = logging.Formatter(
            "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )

    console_handler.setFormatter(formatter)
    logger.addHandler(console_handler)

    # File handler (if specified)
    if log_file:
        file_handler = logging.FileHandler(log_file)
        file_handler.setLevel(getattr(logging, log_level))
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)

    return logger


class StructuredLogger:
    """Wrapper for structured logging with context"""

    def __init__(self, logger: logging.Logger):
        self.logger = logger

    def log_operation(
        self,
        operation_id: str,
        agent_id: str,
        operation_type: str,
        request_summary: str,
        result_summary: Optional[str] = None,
        latency_ms: int = 0,
        success: bool = True,
        error_code: Optional[str] = None,
        session_id: Optional[str] = None,
    ):
        """Log operation with structured context"""
        level = logging.INFO if success else logging.ERROR

        extra = {
            "operation_id": operation_id,
            "agent_id": agent_id,
            "operation_type": operation_type,
            "latency_ms": latency_ms,
        }

        if session_id:
            extra["session_id"] = session_id
        if error_code:
            extra["error_code"] = error_code

        message = (
            f"Operation {operation_id}: {operation_type} - "
            f"Request: {request_summary}, Result: {result_summary or 'N/A'}"
        )

        # Create a new LogRecord with extra fields
        record = self.logger.makeRecord(
            name=self.logger.name,
            level=level,
            fn=__file__,
            lno=0,
            msg=message,
            args=(),
            exc_info=None,
        )

        # Attach extra fields to record
        for key, value in extra.items():
            setattr(record, key, value)

        self.logger.handle(record)

    def log_error(
        self,
        message: str,
        error_code: str,
        agent_id: Optional[str] = None,
        operation_id: Optional[str] = None,
        session_id: Optional[str] = None,
    ):
        """Log error with structured context"""
        extra = {"error_code": error_code}

        if agent_id:
            extra["agent_id"] = agent_id
        if operation_id:
            extra["operation_id"] = operation_id
        if session_id:
            extra["session_id"] = session_id

        record = self.logger.makeRecord(
            name=self.logger.name,
            level=logging.ERROR,
            fn=__file__,
            lno=0,
            msg=message,
            args=(),
            exc_info=None,
        )

        for key, value in extra.items():
            setattr(record, key, value)

        self.logger.handle(record)

    def log_connection(self, message: str, is_connected: bool, terminal_version: Optional[str] = None):
        """Log connection status"""
        extra = {}
        if terminal_version:
            extra["terminal_version"] = terminal_version

        level = logging.INFO if is_connected else logging.WARNING

        record = self.logger.makeRecord(
            name=self.logger.name,
            level=level,
            fn=__file__,
            lno=0,
            msg=message,
            args=(),
            exc_info=None,
        )

        for key, value in extra.items():
            setattr(record, key, value)

        self.logger.handle(record)
