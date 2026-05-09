"""Bidirectional streaming callback manager"""
import threading
import uuid
from typing import Optional, Callable, List, Dict, Any
from datetime import datetime
from collections import defaultdict


class CallbackStream:
    """Represents a bidirectional callback stream for an agent session"""

    def __init__(self, stream_id: str, session_id: str, agent_id: str):
        self.stream_id = stream_id
        self.session_id = session_id
        self.agent_id = agent_id
        self.created_at = datetime.utcnow()
        self.is_active = True
        self.callbacks: List[Callable] = []
        self.lock = threading.RLock()

    def register_callback(self, callback: Callable) -> None:
        """Register a callback function"""
        with self.lock:
            self.callbacks.append(callback)

    def unregister_callback(self, callback: Callable) -> None:
        """Unregister a callback function"""
        with self.lock:
            if callback in self.callbacks:
                self.callbacks.remove(callback)

    def push_update(self, operation_id: str, status: str, data: Optional[Dict[str, Any]] = None) -> None:
        """Push operation status update to all registered callbacks"""
        with self.lock:
            if not self.is_active:
                return

            update = {
                "operation_id": operation_id,
                "status": status,
                "timestamp": datetime.utcnow().isoformat(),
                "data": data or {},
            }

            for callback in self.callbacks:
                try:
                    callback(update)
                except Exception as e:
                    print(f"Error executing callback: {e}")

    def close(self) -> None:
        """Close the stream"""
        with self.lock:
            self.is_active = False
            self.callbacks.clear()


class CallbackManager:
    """Manages bidirectional callback streams for agent sessions"""

    def __init__(self):
        self.streams: Dict[str, CallbackStream] = {}
        self.session_streams: Dict[str, List[str]] = defaultdict(list)
        self.lock = threading.RLock()

    def create_stream(self, session_id: str, agent_id: str) -> str:
        """Create a new callback stream"""
        stream_id = str(uuid.uuid4())

        with self.lock:
            stream = CallbackStream(stream_id, session_id, agent_id)
            self.streams[stream_id] = stream
            self.session_streams[session_id].append(stream_id)

        return stream_id

    def get_stream(self, stream_id: str) -> Optional[CallbackStream]:
        """Get callback stream by ID"""
        with self.lock:
            return self.streams.get(stream_id)

    def get_session_streams(self, session_id: str) -> List[CallbackStream]:
        """Get all streams for a session"""
        with self.lock:
            stream_ids = self.session_streams.get(session_id, [])
            return [self.streams[sid] for sid in stream_ids if sid in self.streams]

    def close_stream(self, stream_id: str) -> bool:
        """Close a callback stream"""
        with self.lock:
            if stream_id in self.streams:
                stream = self.streams[stream_id]
                stream.close()

                # Remove from session mapping
                for session_id, stream_ids in self.session_streams.items():
                    if stream_id in stream_ids:
                        stream_ids.remove(stream_id)

                del self.streams[stream_id]
                return True

            return False

    def close_session_streams(self, session_id: str) -> int:
        """Close all streams for a session"""
        closed_count = 0

        with self.lock:
            stream_ids = self.session_streams.get(session_id, []).copy()

            for stream_id in stream_ids:
                if self.close_stream(stream_id):
                    closed_count += 1

            # Clear session mapping
            if session_id in self.session_streams:
                del self.session_streams[session_id]

        return closed_count

    def push_update_to_stream(
        self, stream_id: str, operation_id: str, status: str, data: Optional[Dict[str, Any]] = None
    ) -> bool:
        """Push update to a specific stream"""
        stream = self.get_stream(stream_id)
        if stream:
            stream.push_update(operation_id, status, data)
            return True
        return False

    def push_update_to_session(
        self, session_id: str, operation_id: str, status: str, data: Optional[Dict[str, Any]] = None
    ) -> int:
        """Push update to all streams in a session"""
        streams = self.get_session_streams(session_id)
        for stream in streams:
            stream.push_update(operation_id, status, data)
        return len(streams)

    def broadcast_update(
        self, operation_id: str, status: str, data: Optional[Dict[str, Any]] = None
    ) -> int:
        """Broadcast update to all active streams"""
        with self.lock:
            count = 0
            for stream in self.streams.values():
                if stream.is_active:
                    stream.push_update(operation_id, status, data)
                    count += 1
        return count

    def health_check(self) -> Dict[str, Any]:
        """Get health status of callback manager"""
        with self.lock:
            return {
                "total_streams": len(self.streams),
                "active_sessions": len(self.session_streams),
                "streams_by_session": {
                    sid: len(stream_ids) for sid, stream_ids in self.session_streams.items()
                },
            }


class StreamWriter:
    """Helper to write updates to a stream"""

    def __init__(self, stream: CallbackStream):
        self.stream = stream

    def write_queued(self, operation_id: str) -> None:
        """Write QUEUED status"""
        self.stream.push_update(operation_id, "QUEUED")

    def write_executing(self, operation_id: str) -> None:
        """Write EXECUTING status"""
        self.stream.push_update(operation_id, "EXECUTING")

    def write_completed(self, operation_id: str, result: Dict[str, Any]) -> None:
        """Write COMPLETED status with result"""
        self.stream.push_update(operation_id, "COMPLETED", result)

    def write_failed(self, operation_id: str, error_code: str, error_message: str) -> None:
        """Write FAILED status with error"""
        self.stream.push_update(
            operation_id, "FAILED", {"error_code": error_code, "error_message": error_message}
        )
