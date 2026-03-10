# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Build and Development Commands

```bash
make dev              # Start Go backend + Vite frontend in parallel
make dev-backend      # Go backend only (port 8080)
make dev-frontend     # Vite dev server only (port 5173, proxies /api to 8080)
make build            # Build frontend + Go binary with embedded SPA
make build-linux-amd64  # Cross-compile for Linux amd64
make build-linux-arm64  # Cross-compile for Linux arm64
make test             # Run Go tests
make clean            # Clean build artifacts
```

## Architecture Overview

Go application for managing network routing (nftables NAT), nginx domains, and network nodes via web UI. Single binary with embedded React SPA.

### Backend (Go + Chi)

Layered architecture per module: **Handler → Service → Repository (+ Engine for side-effects)**

- `cmd/server/main.go` — entry point, Chi router, embedded SPA serving
- `internal/config/` — env-based configuration
- `internal/database/` — SQLite (modernc.org/sqlite, pure Go), embedded migrations
- `internal/auth/` — session-based auth (bcrypt + cookie), middleware
- `internal/nftables/` — NAT group CRUD + nftables engine (google/nftables via netlink)
  - `engine_linux.go` — real nftables implementation
  - `engine_stub.go` — no-op for non-Linux
- `internal/network_nodes/` — network node CRUD
- `internal/nginx/` — domain management, config generation, certbot SSL
- `internal/audit/` — audit logging with pagination
- `internal/pkg/httputil/` — JSON response helpers
- `internal/pkg/validate/` — IP, port, domain validation

### Frontend (React + Vite + Tailwind v4)

- `web/src/pages/` — LoginPage, NatRulesPage, NetworkNodesPage, NginxDomainsPage, AuditLogPage
- `web/src/components/` — UI components (Button, Input, Modal, Toggle, Sidebar)
- `web/src/api/client.ts` — fetch wrapper with 401 redirect
- `web/src/types/` — TypeScript interfaces matching Go models
- Built with Vite, output goes to `web/dist/`, copied to `cmd/server/frontend/` for embedding

### Key Data Flow

1. **Startup**: apply all enabled NAT groups via nftables engine
2. **Auth**: session cookie, validated on every /api request
3. **NAT changes**: update DB → atomically rebuild all nftables rules
4. **Nginx changes**: write config file → `nginx -t` → `systemctl reload nginx`

## Environment Variables

- `PORT` — HTTP port (default: 8080)
- `ADMIN_USER` — initial admin username (default: admin)
- `ADMIN_PASSWORD` — admin password (required)
- `DB_PATH` — SQLite database path (default: system-control.db)
- `NGINX_SITES_DIR` — nginx sites-enabled path (default: /etc/nginx/sites-enabled)
- `SESSION_MAX_AGE` — session lifetime in seconds (default: 604800 = 7 days)

## Deployment

Single binary, installed via `scripts/install.sh` on Debian/Ubuntu. Creates systemd service + nginx reverse proxy.
