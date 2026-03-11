CREATE TABLE IF NOT EXISTS wireguard_peers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    public_key TEXT NOT NULL,
    preshared_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    allowed_ips TEXT NOT NULL,
    address TEXT NOT NULL,
    dns TEXT NOT NULL DEFAULT '1.1.1.1, 1.0.0.1',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
