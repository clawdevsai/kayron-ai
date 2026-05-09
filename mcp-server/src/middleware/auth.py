"""API key authentication middleware for gRPC service"""
import sqlite3
from typing import Optional, Tuple
from datetime import datetime


class AuthMiddleware:
    """API key authentication and validation"""

    def __init__(self, db_path: str = "mcp-server.db"):
        self.db_path = db_path

    def register_api_key(self, agent_id: str, api_key: str) -> Tuple[bool, str]:
        """Register new API key for agent"""
        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            # Check if agent already has key
            cursor.execute("SELECT api_key FROM api_keys WHERE agent_id = ?", (agent_id,))
            existing = cursor.fetchone()

            if existing:
                return False, "Agent already has API key"

            # Insert new key
            cursor.execute(
                """
                INSERT INTO api_keys (key_id, api_key, agent_id, is_active)
                VALUES (?, ?, ?, 1)
                """,
                (agent_id, api_key, agent_id),
            )

            conn.commit()
            conn.close()

            return True, "API key registered successfully"

        except Exception as e:
            return False, f"Registration error: {str(e)}"

    def validate_api_key(self, api_key: str) -> Tuple[bool, Optional[str], str]:
        """Validate API key and return agent_id"""
        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            # Check key existence and active status
            cursor.execute(
                """
                SELECT agent_id FROM api_keys
                WHERE api_key = ? AND is_active = 1
                """,
                (api_key,),
            )

            result = cursor.fetchone()

            # Update last_used_at
            if result:
                cursor.execute(
                    "UPDATE api_keys SET last_used_at = ? WHERE api_key = ?",
                    (datetime.utcnow().isoformat(), api_key),
                )
                conn.commit()

            conn.close()

            if result:
                return True, result[0], "Valid API key"
            else:
                return False, None, "Invalid or inactive API key"

        except Exception as e:
            return False, None, f"Validation error: {str(e)}"

    def revoke_api_key(self, api_key: str) -> Tuple[bool, str]:
        """Revoke API key"""
        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()

            cursor.execute("UPDATE api_keys SET is_active = 0 WHERE api_key = ?", (api_key,))

            conn.commit()
            conn.close()

            if cursor.rowcount > 0:
                return True, "API key revoked"
            else:
                return False, "API key not found"

        except Exception as e:
            return False, f"Revocation error: {str(e)}"

    def get_agent_by_key(self, api_key: str) -> Optional[str]:
        """Get agent_id for API key"""
        is_valid, agent_id, _ = self.validate_api_key(api_key)
        return agent_id if is_valid else None

    def list_keys(self, agent_id: Optional[str] = None) -> list:
        """List API keys (optionally filtered by agent)"""
        try:
            conn = sqlite3.connect(self.db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()

            if agent_id:
                cursor.execute(
                    """
                    SELECT key_id, agent_id, is_active, created_at, last_used_at
                    FROM api_keys WHERE agent_id = ?
                    """,
                    (agent_id,),
                )
            else:
                cursor.execute(
                    """
                    SELECT key_id, agent_id, is_active, created_at, last_used_at
                    FROM api_keys
                    """
                )

            rows = cursor.fetchall()
            conn.close()

            return [dict(row) for row in rows]

        except Exception as e:
            return []


class AuthInterceptor:
    """gRPC interceptor for authentication"""

    def __init__(self, auth: AuthMiddleware):
        self.auth = auth

    def intercept(self, continuation, client_call_details):
        """Intercept RPC call and validate API key"""
        # Extract API key from metadata
        metadata = dict(client_call_details.metadata) if client_call_details.metadata else {}
        api_key = metadata.get("api-key", "")

        if not api_key:
            # Return UNAUTHENTICATED error
            raise Exception("UNAUTHENTICATED: No API key provided")

        # Validate key
        is_valid, agent_id, message = self.auth.validate_api_key(api_key)

        if not is_valid:
            raise Exception(f"UNAUTHENTICATED: {message}")

        # Continue with valid key
        return continuation(client_call_details)
