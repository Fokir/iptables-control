#!/bin/bash
set -e

APP_NAME="system-control"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/system-control"
CONFIG_DIR="/etc/system-control"
SERVICE_FILE="/etc/systemd/system/system-control.service"
NGINX_SC_CONF="/etc/nginx/sites-enabled/system-control-panel.conf"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[+]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[x]${NC} $1"; exit 1; }

if [ "$EUID" -ne 0 ]; then
    error "Please run as root: sudo bash uninstall.sh"
fi

echo -e "${YELLOW}This will completely remove System Control from this server.${NC}"
echo ""
echo "  The following will be deleted:"
echo "    - Binary:   $INSTALL_DIR/$APP_NAME"
echo "    - Config:   $CONFIG_DIR/"
echo "    - Database: $DATA_DIR/"
echo "    - Service:  $SERVICE_FILE"
echo "    - Nginx:    $NGINX_SC_CONF"
echo ""
read -p "Continue? [y/N]: " CONFIRM < /dev/tty
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

# Stop and disable service
if systemctl is-active --quiet system-control 2>/dev/null; then
    log "Stopping service..."
    systemctl stop system-control
fi

if systemctl is-enabled --quiet system-control 2>/dev/null; then
    log "Disabling service..."
    systemctl disable system-control
fi

# Remove service file
if [ -f "$SERVICE_FILE" ]; then
    log "Removing systemd service..."
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
fi

# Remove nginx config
if [ -f "$NGINX_SC_CONF" ]; then
    log "Removing nginx config..."
    rm -f "$NGINX_SC_CONF"
    if nginx -t 2>/dev/null; then
        systemctl reload nginx 2>/dev/null || true
    fi
fi

# Remove binary
if [ -f "$INSTALL_DIR/$APP_NAME" ]; then
    log "Removing binary..."
    rm -f "$INSTALL_DIR/$APP_NAME"
fi

# Remove config
if [ -d "$CONFIG_DIR" ]; then
    log "Removing configuration..."
    rm -rf "$CONFIG_DIR"
fi

# Remove data
if [ -d "$DATA_DIR" ]; then
    log "Removing database and data..."
    rm -rf "$DATA_DIR"
fi

echo ""
echo -e "${GREEN}══════════════════════════════════════════${NC}"
echo -e "${GREEN} System Control has been removed.${NC}"
echo -e "${GREEN}══════════════════════════════════════════${NC}"
echo ""
echo -e "  ${YELLOW}Note:${NC} System packages (nginx, certbot, nftables, wireguard-tools)"
echo -e "  were NOT removed. Remove them manually if no longer needed."
echo ""
