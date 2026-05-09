"""SQLite schema initialization for MT5 gRPC service"""
import sqlite3
from pathlib import Path
from typing import Optional


class DatabaseSchema:
    def __init__(self, db_path: str = "mcp-server.db"):
        self.db_path = db_path
        self.connection: Optional[sqlite3.Connection] = None

    def connect(self):
        """Connect to SQLite database"""
        self.connection = sqlite3.connect(self.db_path)
        self.connection.row_factory = sqlite3.Row
        return self.connection

    def close(self):
        """Close database connection"""
        if self.connection:
            self.connection.close()

    def initialize(self):
        """Create all required tables"""
        if not self.connection:
            self.connect()

        cursor = self.connection.cursor()

        # API Keys table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS api_keys (
                key_id TEXT PRIMARY KEY,
                api_key TEXT UNIQUE NOT NULL,
                agent_id TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                is_active BOOLEAN DEFAULT 1,
                last_used_at TIMESTAMP
            )
        """)

        # Queued Operations table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS queued_operations (
                operation_id TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                agent_id TEXT NOT NULL,
                operation_type TEXT NOT NULL,
                status TEXT DEFAULT 'QUEUED',
                request_data TEXT NOT NULL,
                result_data TEXT,
                error_code TEXT,
                error_message TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                started_at TIMESTAMP,
                completed_at TIMESTAMP,
                retry_count INTEGER DEFAULT 0,
                max_retries INTEGER DEFAULT 3,
                FOREIGN KEY (agent_id) REFERENCES api_keys(agent_id)
            )
        """)

        # Operation Logs table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS operation_logs (
                log_id TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                agent_id TEXT NOT NULL,
                operation_type TEXT NOT NULL,
                operation_id TEXT,
                request_summary TEXT,
                result_summary TEXT,
                latency_ms INTEGER,
                success BOOLEAN,
                error_code TEXT,
                timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (operation_id) REFERENCES queued_operations(operation_id)
            )
        """)

        # Agent Sessions table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS agent_sessions (
                session_id TEXT PRIMARY KEY,
                agent_id TEXT NOT NULL,
                api_key_id TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                is_active BOOLEAN DEFAULT 1,
                FOREIGN KEY (api_key_id) REFERENCES api_keys(key_id)
            )
        """)

        # Create indexes for common queries
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_queued_ops_status
            ON queued_operations(status)
        """)
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_queued_ops_agent
            ON queued_operations(agent_id, created_at)
        """)
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_logs_timestamp
            ON operation_logs(timestamp DESC)
        """)
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_sessions_agent
            ON agent_sessions(agent_id)
        """)

        self.connection.commit()

    def drop_all(self):
        """Drop all tables (for testing)"""
        if not self.connection:
            self.connect()

        cursor = self.connection.cursor()
        cursor.execute("DROP TABLE IF EXISTS operation_logs")
        cursor.execute("DROP TABLE IF EXISTS agent_sessions")
        cursor.execute("DROP TABLE IF EXISTS queued_operations")
        cursor.execute("DROP TABLE IF EXISTS api_keys")
        self.connection.commit()


def init_db(db_path: str = "mcp-server.db") -> DatabaseSchema:
    """Initialize database with schema"""
    schema = DatabaseSchema(db_path)
    schema.connect()
    schema.initialize()
    return schema
