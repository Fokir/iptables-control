#!/bin/bash

# ============================================
# Script to check iptables rules inside running container
# Compatible with: Linux, macOS, Windows (Git Bash)
# ============================================

CONTAINER_NAME="${1:-system-control-interactive}"

echo "=========================================="
echo "Checking iptables rules in container: $CONTAINER_NAME"
echo "=========================================="

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: Container '$CONTAINER_NAME' is not running."
    echo ""
    echo "Running containers:"
    docker ps --format "  {{.Names}}"
    exit 1
fi

echo ""
echo "NAT PREROUTING rules:"
docker exec "$CONTAINER_NAME" iptables -t nat -L PREROUTING -n -v --line-numbers

echo ""
echo "NAT POSTROUTING rules:"
docker exec "$CONTAINER_NAME" iptables -t nat -L POSTROUTING -n -v --line-numbers

echo ""
echo "=========================================="
echo "Raw iptables-save output (NAT table):"
echo "=========================================="
docker exec "$CONTAINER_NAME" iptables-save -t nat

echo ""
echo "=========================================="
echo "Database content:"
echo "=========================================="
docker exec "$CONTAINER_NAME" cat /app/database/iptables.json 2>/dev/null | jq '.' || echo "Database not found or empty"
