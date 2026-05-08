-- MT5 MCP Operation Queue Schema

CREATE TABLE IF NOT EXISTS pending_operations (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    payload TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    status TEXT DEFAULT 'PENDING',
    attempts INTEGER DEFAULT 0
);

-- Index on account_id for efficient queries
CREATE INDEX IF NOT EXISTS idx_pending_operations_account_id
ON pending_operations(account_id);

-- Index on created_at for FIFO ordering
CREATE INDEX IF NOT EXISTS idx_pending_operations_created_at
ON pending_operations(created_at);

-- Index on status for filtering by status
CREATE INDEX IF NOT EXISTS idx_pending_operations_status
ON pending_operations(status);

-- Combined index for efficient queue processing
CREATE INDEX IF NOT EXISTS idx_pending_operations_status_created_at
ON pending_operations(status, created_at);
