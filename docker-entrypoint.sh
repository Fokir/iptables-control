#!/bin/bash
set -e

echo "=========================================="
echo "IPTABLES TEST ENVIRONMENT ENTRYPOINT"
echo "=========================================="

# Function to show current iptables NAT rules
show_nat_rules() {
    echo ""
    echo "Current NAT PREROUTING rules:"
    iptables -t nat -L PREROUTING -n -v --line-numbers 2>/dev/null || echo "No PREROUTING rules"
    echo ""
    echo "Current NAT POSTROUTING rules:"
    iptables -t nat -L POSTROUTING -n -v --line-numbers 2>/dev/null || echo "No POSTROUTING rules"
    echo ""
}

# Function to add a complete NAT rule pair
add_nat_rule() {
    local target_ip=$1
    local dest_ip=$2
    local source_ip=$3
    local port=$4
    local port_target=${5:-$port}
    local protocol=${6:-tcp}

    iptables -t nat -A PREROUTING -d "$target_ip" -p "$protocol" --dport "$port" -j DNAT --to-destination "$dest_ip:$port_target" || true
    iptables -t nat -A POSTROUTING -d "$dest_ip" -p "$protocol" --dport "$port_target" -j SNAT --to-source "$source_ip" || true
    echo "[SETUP] Added rule: $target_ip:$port -> $dest_ip:$port_target ($protocol) (SNAT: $source_ip)"
}

# Check if we should add pre-existing rules (for testing sync)
if [ "$ADD_TEST_RULES" = "true" ]; then
    echo "[SETUP] Adding pre-existing test iptables rules..."

    # Test rule 1: A complete NAT rule pair (for TEST 1: sync detection)
    add_nat_rule "10.0.0.1" "192.168.1.100" "192.168.1.1" 8080

    echo "[SETUP] Basic test rules added"
fi

# Test rule 2: Second rule for multiple groups testing
if [ "$ADD_SECOND_RULE" = "true" ]; then
    add_nat_rule "10.0.0.2" "192.168.1.200" "192.168.1.1" 443
    echo "[SETUP] Second test rule added"
fi

# Test rule 3: Rule that matches disabled rule in database
if [ "$ADD_DISABLED_MATCH_RULE" = "true" ]; then
    add_nat_rule "10.0.0.5" "192.168.1.50" "192.168.1.1" 5000
    echo "[SETUP] Disabled matching rule added"
fi

# Test rule 4: Orphan PREROUTING only (no matching POSTROUTING)
if [ "$ADD_ORPHAN_PREROUTING" = "true" ]; then
    iptables -t nat -A PREROUTING -d 10.0.0.70 -p tcp --dport 7000 -j DNAT --to-destination 192.168.1.70:7000 || true
    echo "[SETUP] Added orphan PREROUTING rule (no POSTROUTING): 10.0.0.70:7000"
fi

# Test rule 5: Orphan POSTROUTING only (no matching PREROUTING)
if [ "$ADD_ORPHAN_POSTROUTING" = "true" ]; then
    iptables -t nat -A POSTROUTING -d 192.168.1.71 -p tcp --dport 7100 -j SNAT --to-source 192.168.1.1 || true
    echo "[SETUP] Added orphan POSTROUTING rule (no PREROUTING): 192.168.1.71:7100"
fi

# Test rule 6: Partial match rule (same IP as disabled, different port)
if [ "$ADD_PARTIAL_MATCH_RULE" = "true" ]; then
    # This rule has same IPs as test-partial-match-disabled but different port (6001 vs 6000)
    add_nat_rule "10.0.0.60" "192.168.1.60" "192.168.1.1" 6001
    echo "[SETUP] Partial match rule added (different port from disabled)"
fi

# Test rule 7: Duplicate rule for idempotency testing
if [ "$ADD_DUPLICATE_RULE" = "true" ]; then
    # Add the same rule twice to test duplicate handling
    add_nat_rule "10.0.0.80" "192.168.1.80" "192.168.1.1" 8000
    add_nat_rule "10.0.0.80" "192.168.1.80" "192.168.1.1" 8000
    echo "[SETUP] Duplicate rules added for idempotency testing"
fi

# Test rule 8: UDP rule
if [ "$ADD_UDP_RULE" = "true" ]; then
    add_nat_rule "10.0.0.90" "192.168.1.90" "192.168.1.1" 53 53 "udp"
    echo "[SETUP] UDP rule added"
fi

# Test rule 9: Port forwarding rule (different source and target ports)
if [ "$ADD_PORT_FORWARD_RULE" = "true" ]; then
    add_nat_rule "10.0.0.100" "192.168.1.100" "192.168.1.1" 8888 80
    echo "[SETUP] Port forwarding rule added (8888 -> 80)"
fi

echo ""
echo "[INFO] Initial iptables state before application start:"
show_nat_rules

# Create database directory if not exists
mkdir -p /app/database

# If test database is provided, use it
if [ -f "/app/test-data/iptables.json" ]; then
    echo "[SETUP] Using test database from /app/test-data/iptables.json"
    cp /app/test-data/iptables.json /app/database/iptables.json
fi

if [ -f "/app/test-data/network-nodes.json" ]; then
    echo "[SETUP] Using test network nodes from /app/test-data/network-nodes.json"
    cp /app/test-data/network-nodes.json /app/database/network-nodes.json
fi

echo ""
echo "=========================================="
echo "Starting application..."
echo "=========================================="

# Execute the main command
exec "$@"
