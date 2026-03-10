#!/bin/bash
set -e

APP_NAME="system-control"
REPO="Fokir/iptables-control"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/system-control"
CONFIG_DIR="/etc/system-control"
SERVICE_FILE="/etc/systemd/system/system-control.service"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[+]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[x]${NC} $1"; exit 1; }

# ─────────────────────────────────────────────
# Checks
# ─────────────────────────────────────────────

if [ "$EUID" -ne 0 ]; then
    error "Please run as root: sudo bash install.sh"
fi

if ! grep -qiE 'debian|ubuntu' /etc/os-release 2>/dev/null; then
    error "This script supports only Debian/Ubuntu"
fi

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  BINARY_ARCH="amd64" ;;
    aarch64) BINARY_ARCH="arm64" ;;
    *)       error "Unsupported architecture: $ARCH" ;;
esac

log "OS: $(. /etc/os-release && echo "$PRETTY_NAME"), Arch: $ARCH ($BINARY_ARCH)"

# ─────────────────────────────────────────────
# Install system dependencies
# ─────────────────────────────────────────────

log "Updating package lists..."
apt-get update -qq

log "Installing dependencies..."
DEPS=(
    nginx                   # reverse proxy & domain management
    certbot                 # Let's Encrypt SSL certificates
    python3-certbot-nginx   # certbot nginx plugin
    nftables                # packet filtering (replaces iptables)
    curl                    # downloading binary
    wireguard-tools         # WireGuard VPN (if not already present)
)

for pkg in "${DEPS[@]}"; do
    if dpkg -s "$pkg" &>/dev/null; then
        log "  $pkg — already installed"
    else
        log "  $pkg — installing..."
        apt-get install -y -qq "$pkg" > /dev/null 2>&1
    fi
done

# ─────────────────────────────────────────────
# Enable IP forwarding (required for NAT)
# ─────────────────────────────────────────────

log "Enabling IP forwarding..."
sysctl -w net.ipv4.ip_forward=1 > /dev/null
if ! grep -q "net.ipv4.ip_forward=1" /etc/sysctl.d/99-system-control.conf 2>/dev/null; then
    echo "net.ipv4.ip_forward=1" > /etc/sysctl.d/99-system-control.conf
fi

# ─────────────────────────────────────────────
# Enable and start nftables
# ─────────────────────────────────────────────

log "Enabling nftables..."
systemctl enable nftables 2>/dev/null || true
systemctl start nftables 2>/dev/null || true

# ─────────────────────────────────────────────
# Remove default nginx site (conflicts with default_server)
# ─────────────────────────────────────────────

if [ -f /etc/nginx/sites-enabled/default ]; then
    log "Removing default nginx site..."
    rm -f /etc/nginx/sites-enabled/default
fi

# ─────────────────────────────────────────────
# Create directories
# ─────────────────────────────────────────────

mkdir -p "$DATA_DIR" "$CONFIG_DIR"

# ─────────────────────────────────────────────
# Download binary
# ─────────────────────────────────────────────

log "Downloading latest release from github.com/$REPO..."

LATEST_TAG=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i '^location:' | sed 's|.*/||' | tr -d '\r\n')

if [ -z "$LATEST_TAG" ]; then
    error "Could not determine latest release. Check https://github.com/$REPO/releases"
fi

log "Latest version: $LATEST_TAG"

DOWNLOAD_URL="https://github.com/$REPO/releases/download/${LATEST_TAG}/${APP_NAME}-linux-${BINARY_ARCH}"

if ! curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/$APP_NAME"; then
    error "Download failed from $DOWNLOAD_URL"
fi

chmod +x "$INSTALL_DIR/$APP_NAME"
log "Installed binary to $INSTALL_DIR/$APP_NAME"

# ─────────────────────────────────────────────
# Configuration
# ─────────────────────────────────────────────

if [ ! -f "$CONFIG_DIR/.env" ]; then
    log "Creating configuration..."

    read -p "Admin username [admin]: " ADMIN_USER
    ADMIN_USER=${ADMIN_USER:-admin}

    while true; do
        read -sp "Admin password: " ADMIN_PASSWORD
        echo
        if [ -n "$ADMIN_PASSWORD" ]; then
            break
        fi
        warn "Password cannot be empty, try again"
    done

    read -p "HTTP port [8080]: " PORT
    PORT=${PORT:-8080}

    cat > "$CONFIG_DIR/.env" <<EOF
PORT=$PORT
ADMIN_USER=$ADMIN_USER
ADMIN_PASSWORD=$ADMIN_PASSWORD
DB_PATH=$DATA_DIR/system-control.db
NGINX_SITES_DIR=/etc/nginx/sites-enabled
SESSION_MAX_AGE=604800
EOF

    chmod 600 "$CONFIG_DIR/.env"
    log "Configuration saved to $CONFIG_DIR/.env"
else
    warn "Configuration already exists at $CONFIG_DIR/.env, skipping"
    PORT=$(grep '^PORT=' "$CONFIG_DIR/.env" | cut -d= -f2)
    PORT=${PORT:-8080}
fi

# ─────────────────────────────────────────────
# Systemd service
# ─────────────────────────────────────────────

log "Creating systemd service..."

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=System Control - Network Management
After=network.target nginx.service nftables.service
Wants=nftables.service

[Service]
Type=simple
ExecStart=$INSTALL_DIR/$APP_NAME
EnvironmentFile=$CONFIG_DIR/.env
WorkingDirectory=$DATA_DIR
Restart=always
RestartSec=5
LimitNOFILE=65536

# Security hardening
NoNewPrivileges=no
ProtectSystem=strict
ReadWritePaths=$DATA_DIR /etc/nginx/sites-enabled
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable system-control

# Stop existing service if running (for upgrades)
systemctl stop system-control 2>/dev/null || true
systemctl start system-control

log "Service started"

# ─────────────────────────────────────────────
# Nginx reverse proxy for the panel itself
# ─────────────────────────────────────────────

NGINX_SC_CONF="/etc/nginx/sites-enabled/system-control-panel.conf"
if [ ! -f "$NGINX_SC_CONF" ]; then
    log "Setting up nginx reverse proxy..."

    cat > "$NGINX_SC_CONF" <<'NGINX_EOF'
server {
    listen 80 default_server;
    server_name _;

    location / {
NGINX_EOF

    cat >> "$NGINX_SC_CONF" <<EOF
        proxy_pass http://127.0.0.1:$PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

    if nginx -t 2>/dev/null; then
        systemctl reload nginx
        log "Nginx reverse proxy configured"
    else
        warn "Nginx config test failed, please check manually"
    fi
else
    warn "Nginx config already exists, skipping"
fi

# ─────────────────────────────────────────────
# Done
# ─────────────────────────────────────────────

IP_ADDR=$(hostname -I | awk '{print $1}')

echo ""
echo -e "${GREEN}══════════════════════════════════════════${NC}"
echo -e "${GREEN} Installation complete!${NC}"
echo -e "${GREEN}══════════════════════════════════════════${NC}"
echo ""
echo -e "  Access:    http://${IP_ADDR}"
echo -e "  Config:    $CONFIG_DIR/.env"
echo -e "  Database:  $DATA_DIR/system-control.db"
echo -e "  Service:   systemctl status system-control"
echo -e "  Logs:      journalctl -u system-control -f"
echo ""
echo -e "  ${YELLOW}To upgrade later:${NC}"
echo -e "  curl -sL https://raw.githubusercontent.com/$REPO/master/scripts/install.sh | sudo bash"
echo ""
