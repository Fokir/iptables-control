# System Control

Network management panel for Linux servers. Manages **nftables NAT rules**, **nginx domains with SSL**, and **network nodes** through a clean web interface.

Single binary. One command to install. Zero dependencies to manage.

---

## Quick Install

```bash
curl -sL https://raw.githubusercontent.com/Fokir/iptables-control/main/scripts/install.sh | sudo bash
```

The installer will interactively ask for:

| Prompt | Default | Description |
|--------|---------|-------------|
| **Admin username** | `admin` | Login for the web panel |
| **Admin password** | — | Required, cannot be empty |
| **HTTP port** | `8080` | Internal port (nginx proxies from 80) |

Credentials are saved to `/etc/system-control/.env` (mode `600`, root-only).

### What gets installed

- Binary at `/usr/local/bin/system-control`
- Config at `/etc/system-control/.env`
- Database at `/var/lib/system-control/system-control.db`
- Systemd service `system-control`
- Nginx reverse proxy on port 80

System dependencies installed automatically: `nginx`, `certbot`, `nftables`, `wireguard-tools`.

### Requirements

- **Debian** or **Ubuntu**
- **Root** access
- `x86_64` or `arm64` architecture

---

## Features

### NAT Rules (nftables)
Create and manage Source NAT / Masquerade rules organized in groups. Rules are applied atomically via netlink — no shell commands, no iptables legacy.

### Nginx Domains
Add domains, generate nginx configs, obtain Let's Encrypt SSL certificates via certbot. Configs are validated with `nginx -t` before reload.

### Network Nodes
Track servers and network devices with metadata.

### Audit Log
Every change is logged with user, action, and timestamp. Auto-cleanup after 90 days.

---

## Authentication

Session-based authentication with security hardened defaults:

- Passwords hashed with **bcrypt**
- Sessions stored in SQLite with **UUID** identifiers
- **HttpOnly** + **SameSite=Strict** cookies (XSS/CSRF protection)
- **Sliding window** session renewal (default lifetime: 7 days)
- Expired sessions cleaned up automatically every hour

On first startup, the admin account is created from `ADMIN_USER` / `ADMIN_PASSWORD` environment variables. After that, login through the web UI at `http://<server-ip>`.

---

## Architecture

```
┌─────────────────────────────────────────────┐
│              Single Go Binary               │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐ │
│  │ React SPA│  │ Chi REST │  │  SQLite   │ │
│  │ (embedded)│  │   API    │  │    DB     │ │
│  └──────────┘  └────┬─────┘  └───────────┘ │
│                     │                       │
│         ┌───────────┼───────────┐           │
│         ▼           ▼           ▼           │
│    ┌─────────┐ ┌─────────┐ ┌─────────┐     │
│    │nftables │ │  nginx  │ │  audit  │     │
│    │ netlink │ │ configs │ │   log   │     │
│    └─────────┘ └─────────┘ └─────────┘     │
└─────────────────────────────────────────────┘
```

**Backend:** Go + Chi router, layered architecture (Handler → Service → Repository → Engine)

**Frontend:** React + Vite + Tailwind CSS v4, embedded into the binary at build time

**Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGO)

---

## Configuration

Environment variables (set in `/etc/system-control/.env`):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `ADMIN_USER` | `admin` | Initial admin username |
| `ADMIN_PASSWORD` | — | Admin password (**required**) |
| `DB_PATH` | `system-control.db` | SQLite database path |
| `NGINX_SITES_DIR` | `/etc/nginx/sites-enabled` | Nginx configs directory |
| `SESSION_MAX_AGE` | `604800` | Session lifetime in seconds (7 days) |

---

## Upgrade

Re-run the same command — the script detects existing installation, skips setup, and only updates the binary:

```bash
curl -sL https://raw.githubusercontent.com/Fokir/iptables-control/main/scripts/install.sh | sudo bash
```

If already on the latest version, exits immediately with no changes.

---

## Development

```bash
# Prerequisites: Go 1.23+, Node.js 20+

# Start dev servers (Go backend + Vite frontend)
make dev

# Run tests
make test

# Build production binary
make build

# Cross-compile for Linux
make build-linux-amd64
make build-linux-arm64
```

### Create a release

```bash
./scripts/release.sh
```

Tags trigger GitHub Actions to build binaries and publish a release automatically.

---

## Service Management

```bash
# Status
systemctl status system-control

# Logs
journalctl -u system-control -f

# Restart
systemctl restart system-control
```

---

## License

MIT
