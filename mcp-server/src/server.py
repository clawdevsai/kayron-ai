"""MT5 gRPC Server"""
import asyncio
import signal
import logging
from typing import Optional

from .config import load_config, validate_config
from .db_schema import init_db
from .connection_pool import PoolManager
from .operation_queue import OperationQueue
from .callback_manager import CallbackManager
from .service import MT5Service
from .logging_config import setup_logging


class MT5GRPCServer:
    """Main gRPC server for MT5 operations"""

    def __init__(self, config_path: Optional[str] = None):
        self.config = load_config(config_path)
        validate_config(self.config)

        # Initialize components
        self.logger = setup_logging(
            self.config.logging.level, self.config.logging.log_file, self.config.logging.format
        )

        self.db_schema = init_db(self.config.database.db_path)
        self.pool_manager = PoolManager(self.config.mt5.terminal_path)
        self.operation_queue = OperationQueue(self.config.database.db_path)
        self.callback_manager = CallbackManager()
        self.service = MT5Service(
            self.pool_manager.get_pool(), self.operation_queue, self.callback_manager
        )

        self.server = None
        self.running = False

    async def initialize(self) -> bool:
        """Initialize server and connect to MT5"""
        try:
            self.logger.info("Initializing MT5 gRPC server...")

            # Connect to MT5
            if not self.pool_manager.initialize(
                self.config.mt5.login, self.config.mt5.password, self.config.mt5.server
            ):
                self.logger.error("Failed to connect to MT5 terminal")
                return False

            self.logger.info(f"Connected to MT5 at {self.config.mt5.server}")

            # Recover queued operations
            self._recover_queued_operations()

            self.running = True
            return True

        except Exception as e:
            self.logger.error(f"Initialization error: {e}")
            return False

    def _recover_queued_operations(self):
        """Recover queued operations from database on startup"""
        try:
            queued = self.operation_queue.get_queued(limit=100)
            self.logger.info(f"Recovered {len(queued)} queued operations")

            for op in queued:
                self.logger.info(f"Operation {op['operation_id']} ready for retry")

        except Exception as e:
            self.logger.error(f"Error recovering queued operations: {e}")

    async def start(self) -> None:
        """Start the gRPC server"""
        try:
            self.logger.info(f"Starting gRPC server on {self.config.server.host}:{self.config.server.port}")

            # Note: This is a simplified server. Real implementation would use:
            # import grpc
            # grpc.aio.server() with proper service implementation

            # For now, just log that server is starting
            self.logger.info("gRPC server initialized (awaiting proper gRPC implementation)")

            # Add signal handlers
            loop = asyncio.get_event_loop()
            loop.add_signal_handler(signal.SIGINT, self._handle_shutdown)
            loop.add_signal_handler(signal.SIGTERM, self._handle_shutdown)

            # Keep server running
            while self.running:
                await asyncio.sleep(1)

        except Exception as e:
            self.logger.error(f"Server error: {e}")
            await self.shutdown()

    def _handle_shutdown(self):
        """Handle graceful shutdown"""
        self.logger.info("Shutdown signal received")
        self.running = False

    async def shutdown(self) -> None:
        """Shutdown server and clean up resources"""
        try:
            self.logger.info("Shutting down gRPC server...")

            # Close all callback streams
            closed = self.callback_manager.close_session_streams("")
            self.logger.info(f"Closed {closed} callback streams")

            # Disconnect from MT5
            self.pool_manager.shutdown()
            self.logger.info("Disconnected from MT5")

            # Close database
            if self.db_schema:
                self.db_schema.close()

            self.running = False
            self.logger.info("gRPC server shutdown complete")

        except Exception as e:
            self.logger.error(f"Error during shutdown: {e}")

    async def health_check(self) -> dict:
        """Health check endpoint"""
        return await self.service.health_check()


async def main():
    """Main entry point"""
    server = MT5GRPCServer()

    if await server.initialize():
        await server.start()
    else:
        print("Failed to initialize server")
        exit(1)


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("Server stopped")
