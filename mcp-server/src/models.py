from datetime import datetime
from typing import Optional, List
from enum import Enum
from pydantic import BaseModel, Field


class OperationStatus(str, Enum):
    QUEUED = "QUEUED"
    EXECUTING = "EXECUTING"
    COMPLETED = "COMPLETED"
    FAILED = "FAILED"


class OperationType(str, Enum):
    PLACE_ORDER = "PlaceOrder"
    CLOSE_POSITION = "ClosePosition"
    GET_ACCOUNT_INFO = "GetAccountInfo"
    GET_POSITIONS = "GetPositions"
    GET_SYMBOLS = "GetSymbols"
    GET_RATES = "GetRates"
    GET_TICKS = "GetTicks"


class AgentSession(BaseModel):
    """Agent session management"""
    session_id: str
    agent_id: str
    api_key_id: str
    created_at: datetime = Field(default_factory=datetime.utcnow)
    last_activity_at: datetime = Field(default_factory=datetime.utcnow)
    is_active: bool = True

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}


class QueuedOperation(BaseModel):
    """Operation queued for MT5 execution"""
    operation_id: str
    session_id: str
    agent_id: str
    operation_type: OperationType
    status: OperationStatus = OperationStatus.QUEUED
    request_data: dict
    result_data: Optional[dict] = None
    error_code: Optional[str] = None
    error_message: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    retry_count: int = 0
    max_retries: int = 3

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}


class MT5Connection(BaseModel):
    """MT5 terminal connection state"""
    connection_id: str
    is_active: bool = False
    thread_id: Optional[int] = None
    connected_at: Optional[datetime] = None
    last_heartbeat: Optional[datetime] = None
    terminal_version: Optional[str] = None

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}


class CallbackStream(BaseModel):
    """Bidirectional callback stream for operation updates"""
    stream_id: str
    session_id: str
    agent_id: str
    created_at: datetime = Field(default_factory=datetime.utcnow)
    is_active: bool = True

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}


class OperationLog(BaseModel):
    """Audit log for all operations"""
    log_id: str
    session_id: str
    agent_id: str
    operation_type: OperationType
    operation_id: str
    request_summary: str
    result_summary: str
    latency_ms: int
    success: bool
    error_code: Optional[str] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}
