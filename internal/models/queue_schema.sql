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

-- Account equity history (for Phase 3 analytics)
CREATE TABLE IF NOT EXISTS account_equity_history (
    account_id INTEGER NOT NULL,
    timestamp INTEGER NOT NULL,
    equity TEXT NOT NULL,
    balance TEXT NOT NULL,
    PRIMARY KEY (account_id, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_account_equity_history_account_id
ON account_equity_history(account_id);

CREATE INDEX IF NOT EXISTS idx_account_equity_history_timestamp
ON account_equity_history(timestamp);

-- Order fill analysis (for Phase 4 analytics)
CREATE TABLE IF NOT EXISTS order_fills (
    ticket INTEGER PRIMARY KEY,
    symbol TEXT NOT NULL,
    fill_time INTEGER NOT NULL,
    fill_price TEXT NOT NULL,
    slippage TEXT NOT NULL,
    execution_latency_ms INTEGER,
    created_at INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_order_fills_symbol
ON order_fills(symbol);

CREATE INDEX IF NOT EXISTS idx_order_fills_fill_time
ON order_fills(fill_time);
