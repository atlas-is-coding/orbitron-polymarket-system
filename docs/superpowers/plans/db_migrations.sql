-- Order History Schema for polytrade-bot
-- Stores complete order and trade history with wallet statistics and notifications

-- Main orders table: stores all orders placed by the bot
CREATE TABLE IF NOT EXISTS orders_history (
    id TEXT PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    condition_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    side TEXT NOT NULL,                    -- 'BUY' or 'SELL'
    order_type TEXT NOT NULL,              -- 'LIMIT', 'MARKET', etc.
    price REAL NOT NULL,
    size REAL NOT NULL,
    status TEXT NOT NULL,                  -- 'PENDING', 'OPEN', 'FILLED', 'CANCELED', 'EXPIRED'
    expires_at TEXT,                       -- GTD expiration timestamp (ISO 8601)
    created_at TEXT NOT NULL,              -- ISO 8601
    updated_at TEXT NOT NULL,              -- ISO 8601
    synced_at TEXT,                        -- Last sync with API

    -- Indexing
    FOREIGN KEY (wallet_address) REFERENCES wallets(address)
);

CREATE INDEX IF NOT EXISTS idx_orders_wallet_status
    ON orders_history(wallet_address, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_condition
    ON orders_history(condition_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_expires
    ON orders_history(expires_at)
    WHERE expires_at IS NOT NULL;


-- Trades table: stores fills/trades from orders
CREATE TABLE IF NOT EXISTS trades_history (
    id TEXT PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    order_id TEXT NOT NULL,
    trade_id TEXT NOT NULL,
    condition_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    side TEXT NOT NULL,
    price REAL NOT NULL,
    size REAL NOT NULL,
    fee REAL NOT NULL DEFAULT 0,
    timestamp TEXT NOT NULL,               -- ISO 8601

    FOREIGN KEY (wallet_address) REFERENCES wallets(address),
    FOREIGN KEY (order_id) REFERENCES orders_history(id)
);

CREATE INDEX IF NOT EXISTS idx_trades_wallet_time
    ON trades_history(wallet_address, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_trades_condition
    ON trades_history(condition_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_trades_order
    ON trades_history(order_id);


-- Wallet statistics: snapshots of wallet balance and P&L
CREATE TABLE IF NOT EXISTS wallet_statistics (
    wallet_address TEXT NOT NULL,
    fetched_at INTEGER NOT NULL,           -- Unix timestamp
    balance_usd REAL NOT NULL DEFAULT 0,
    pnl_usd REAL NOT NULL DEFAULT 0,
    win_rate REAL DEFAULT 0,               -- Percentage of winning trades
    total_trades INTEGER DEFAULT 0,
    total_volume REAL DEFAULT 0,

    PRIMARY KEY (wallet_address, fetched_at),
    FOREIGN KEY (wallet_address) REFERENCES wallets(address)
);

CREATE INDEX IF NOT EXISTS idx_wallet_stats_recent
    ON wallet_statistics(wallet_address, fetched_at DESC);


-- Notifications queue: for delivery guarantees with retry logic
CREATE TABLE IF NOT EXISTS notifications_queue (
    id TEXT PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    event_type TEXT NOT NULL,              -- 'ORDER_PLACED', 'ORDER_FILLED', 'ORDER_CANCELED', etc.
    payload TEXT NOT NULL,                 -- JSON
    status TEXT NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'SENT', 'FAILED', 'DELIVERED'
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    next_retry_at INTEGER,                 -- Unix timestamp
    created_at TEXT NOT NULL,              -- ISO 8601
    updated_at TEXT NOT NULL,              -- ISO 8601

    FOREIGN KEY (wallet_address) REFERENCES wallets(address)
);

CREATE INDEX IF NOT EXISTS idx_notifications_pending
    ON notifications_queue(wallet_address, status, next_retry_at);


-- Wallets table: reference for foreign keys (if not already defined elsewhere)
CREATE TABLE IF NOT EXISTS wallets (
    address TEXT PRIMARY KEY,
    created_at INTEGER NOT NULL            -- Unix timestamp
);


-- Optional: Order performance metrics table
CREATE TABLE IF NOT EXISTS order_metrics (
    id TEXT PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    condition_id TEXT NOT NULL,
    avg_entry_price REAL,
    avg_exit_price REAL,
    total_quantity REAL,
    realized_pnl REAL DEFAULT 0,
    unrealized_pnl REAL DEFAULT 0,
    win_count INTEGER DEFAULT 0,
    loss_count INTEGER DEFAULT 0,
    computed_at TEXT NOT NULL,             -- ISO 8601

    FOREIGN KEY (wallet_address) REFERENCES wallets(address)
);

CREATE INDEX IF NOT EXISTS idx_metrics_wallet
    ON order_metrics(wallet_address, computed_at DESC);
