"""Agent session management"""
import sqlite3
import uuid
from typing import Optional, List, Dict, Any
from datetime import datetime, timedelta
import threading


class SessionManager:
    """Manage agent sessions and track activity"""

    def __init__(self, db_path: str = "mcp-server.db", timeout_minutes: int = 30):
        self.db_path = db_path
        self.timeout = timedelta(minutes=timeout_minutes)
        self.lock = threading.RLock()
        self.sessions: Dict[str, dict] = {}  # In-memory session cache

    def create_session(self, agent_id: str, api_key_id: str) -> str:
        """Create new session for agent"""
        session_id = str(uuid.uuid4())

        with self.lock:
            # Store in database
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                INSERT INTO agent_sessions
                (session_id, agent_id, api_key_id, created_at, last_activity_at, is_active)
                VALUES (?, ?, ?, ?, ?, 1)
                """,
                (session_id, agent_id, api_key_id, datetime.utcnow().isoformat(), datetime.utcnow().isoformat()),
            )

            conn.commit()
            conn.close()

            # Cache in memory
            self.sessions[session_id] = {
                "session_id": session_id,
                "agent_id": agent_id,
                "api_key_id": api_key_id,
                "created_at": datetime.utcnow(),
                "last_activity_at": datetime.utcnow(),
                "is_active": True,
            }

        return session_id

    def update_activity(self, session_id: str) -> bool:
        """Update last activity timestamp"""
        with self.lock:
            if session_id not in self.sessions:
                return False

            now = datetime.utcnow()
            self.sessions[session_id]["last_activity_at"] = now

            # Update database
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                "UPDATE agent_sessions SET last_activity_at = ? WHERE session_id = ?",
                (now.isoformat(), session_id),
            )

            conn.commit()
            conn.close()

        return True

    def get_session(self, session_id: str) -> Optional[Dict[str, Any]]:
        """Get session details"""
        with self.lock:
            return self.sessions.get(session_id)

    def list_sessions(self, agent_id: Optional[str] = None) -> List[Dict[str, Any]]:
        """List active sessions"""
        with self.lock:
            sessions = list(self.sessions.values())

            if agent_id:
                sessions = [s for s in sessions if s["agent_id"] == agent_id]

            return sessions

    def is_session_active(self, session_id: str) -> bool:
        """Check if session is active and not timed out"""
        with self.lock:
            session = self.sessions.get(session_id)

            if not session or not session["is_active"]:
                return False

            # Check timeout
            last_activity = session["last_activity_at"]
            if isinstance(last_activity, str):
                last_activity = datetime.fromisoformat(last_activity)

            if datetime.utcnow() - last_activity > self.timeout:
                # Mark as inactive
                session["is_active"] = False
                return False

            return True

    def close_session(self, session_id: str) -> bool:
        """Close session"""
        with self.lock:
            if session_id not in self.sessions:
                return False

            session = self.sessions[session_id]
            session["is_active"] = False

            # Update database
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute("UPDATE agent_sessions SET is_active = 0 WHERE session_id = ?", (session_id,))

            conn.commit()
            conn.close()

        return True

    def cleanup_expired_sessions(self) -> int:
        """Clean up expired sessions"""
        with self.lock:
            expired_ids = []

            for session_id, session in self.sessions.items():
                last_activity = session["last_activity_at"]
                if isinstance(last_activity, str):
                    last_activity = datetime.fromisoformat(last_activity)

                if datetime.utcnow() - last_activity > self.timeout:
                    expired_ids.append(session_id)

            # Remove expired sessions
            for sid in expired_ids:
                self.close_session(sid)

            # Clean up database
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute(
                """
                DELETE FROM agent_sessions
                WHERE is_active = 0 AND last_activity_at < datetime('now', '-' || ? || ' minutes')
                """,
                (int(self.timeout.total_seconds() / 60),),
            )

            deleted = cursor.rowcount
            conn.commit()
            conn.close()

        return len(expired_ids)

    def load_sessions_from_db(self) -> int:
        """Load active sessions from database on startup"""
        with self.lock:
            conn = sqlite3.connect(self.db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()

            cursor.execute(
                """
                SELECT * FROM agent_sessions WHERE is_active = 1
                """
            )

            rows = cursor.fetchall()
            conn.close()

            for row in rows:
                session_id = row["session_id"]
                self.sessions[session_id] = {
                    "session_id": session_id,
                    "agent_id": row["agent_id"],
                    "api_key_id": row["api_key_id"],
                    "created_at": datetime.fromisoformat(row["created_at"]),
                    "last_activity_at": datetime.fromisoformat(row["last_activity_at"]),
                    "is_active": bool(row["is_active"]),
                }

            return len(rows)

    def health_check(self) -> Dict[str, Any]:
        """Health check for session manager"""
        with self.lock:
            active_count = sum(1 for s in self.sessions.values() if s["is_active"])
            return {
                "total_sessions": len(self.sessions),
                "active_sessions": active_count,
                "session_timeout_minutes": int(self.timeout.total_seconds() / 60),
            }
