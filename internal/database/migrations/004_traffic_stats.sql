-- Raw traffic data points (short retention, ~2 hours)
CREATE TABLE IF NOT EXISTS traffic_raw (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id INTEGER,
    interface_name TEXT,
    bytes_in INTEGER NOT NULL DEFAULT 0,
    bytes_out INTEGER NOT NULL DEFAULT 0,
    collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_traffic_raw_time ON traffic_raw(collected_at);
CREATE INDEX IF NOT EXISTS idx_traffic_raw_node ON traffic_raw(node_id, collected_at);
CREATE INDEX IF NOT EXISTS idx_traffic_raw_iface ON traffic_raw(interface_name, collected_at);

-- Aggregated traffic data (rollups for longer retention)
CREATE TABLE IF NOT EXISTS traffic_agg (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id INTEGER,
    interface_name TEXT,
    bytes_in INTEGER NOT NULL DEFAULT 0,
    bytes_out INTEGER NOT NULL DEFAULT 0,
    bucket_start DATETIME NOT NULL,
    bucket_seconds INTEGER NOT NULL,
    sample_count INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_traffic_agg_bucket ON traffic_agg(bucket_start, bucket_seconds);
CREATE INDEX IF NOT EXISTS idx_traffic_agg_node ON traffic_agg(node_id, bucket_seconds, bucket_start);
CREATE INDEX IF NOT EXISTS idx_traffic_agg_iface ON traffic_agg(interface_name, bucket_seconds, bucket_start);
