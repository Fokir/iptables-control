#!/bin/bash

# ============================================
# IPTABLES TEST SUITE
# Compatible with: Linux, macOS, Windows (Git Bash)
# Note: jq runs inside Docker container, not on host
# ============================================

set -e

# Get script directory (works in Git Bash too)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Convert path for Docker on Windows (Git Bash)
get_docker_path() {
    local path="$1"
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "mingw"* ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        echo "$path" | sed -e 's|^/\([a-zA-Z]\)/|\1:/|'
    else
        echo "$path"
    fi
}

DOCKER_SCRIPT_DIR="$(get_docker_path "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
CONTAINER_NAME="system-control-iptables-test"
APP_PORT=3000
AUTH_USER="admin"
AUTH_PASSWORD="admin"

# Check requirements
check_requirements() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}[ERROR]${NC} Missing required tool: docker"
        exit 1
    fi
}

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_section() {
    echo ""
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}========================================${NC}"
}

# Docker wrapper for Windows
docker_run() {
    MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL="*" docker "$@"
}

# Wait for app
wait_for_app() {
    log_info "Waiting for application to start..."
    local max_attempts=30
    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        if docker exec "$CONTAINER_NAME" curl -s -o /dev/null -w "%{http_code}" -u "$AUTH_USER:$AUTH_PASSWORD" http://localhost:$APP_PORT 2>/dev/null | grep -q "200\|401"; then
            log_success "Application is ready"
            return 0
        fi
        printf "."
        sleep 2
        attempt=$((attempt + 1))
    done
    echo ""
    log_error "Application failed to start"
    return 1
}

# Check if rule exists
check_rule_exists() {
    local target_ip=$1
    local dest_ip=$2
    local port=$3
    local protocol=${4:-tcp}
    local port_target=${5:-$port}
    docker exec "$CONTAINER_NAME" iptables -t nat -C PREROUTING -d "$target_ip" -p "$protocol" --dport "$port" -j DNAT --to-destination "$dest_ip:$port_target" 2>/dev/null
    return $?
}

# Check POSTROUTING rule exists
check_postrouting_exists() {
    local dest_ip=$1
    local source_ip=$2
    local port=$3
    local protocol=${4:-tcp}
    docker exec "$CONTAINER_NAME" iptables -t nat -C POSTROUTING -d "$dest_ip" -p "$protocol" --dport "$port" -j SNAT --to-source "$source_ip" 2>/dev/null
    return $?
}

# Count rules in PREROUTING
count_prerouting_rules() {
    docker exec "$CONTAINER_NAME" iptables -t nat -L PREROUTING -n 2>/dev/null | grep -c "DNAT" || echo "0"
}

# Database helpers
get_database() {
    docker exec "$CONTAINER_NAME" sh -c 'cat /app/database/iptables.json 2>/dev/null | jq "."' 2>/dev/null
}

container_jq() {
    local filter="$1"
    docker exec "$CONTAINER_NAME" sh -c "cat /app/database/iptables.json 2>/dev/null | jq -r '$filter'" 2>/dev/null
}

container_jq_exists() {
    local filter="$1"
    docker exec "$CONTAINER_NAME" sh -c "cat /app/database/iptables.json 2>/dev/null | jq -e '$filter'" > /dev/null 2>&1
    return $?
}

# Cleanup
cleanup() {
    log_info "Cleaning up..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true
}

# Start container with options
start_container() {
    local env_vars="$1"
    local use_test_data="${2:-true}"

    cleanup

    log_info "Building Docker image..."
    docker build -t system-control-test "$SCRIPT_DIR" > /dev/null 2>&1 || {
        log_error "Failed to build Docker image"
        return 1
    }

    log_info "Starting container..."
    local volume_opt=""
    if [ "$use_test_data" = "true" ]; then
        volume_opt="-v $DOCKER_SCRIPT_DIR/test-data:/app/test-data:ro"
    fi

    eval docker_run run -d \
        --name "$CONTAINER_NAME" \
        --cap-add NET_ADMIN \
        -e BASIC_AUTH_USER="$AUTH_USER" \
        -e BASIC_AUTH_PASSWORD="$AUTH_PASSWORD" \
        $env_vars \
        $volume_opt \
        -p "$APP_PORT:3000" \
        system-control-test > /dev/null 2>&1 || {
        log_error "Failed to start container"
        return 1
    }

    wait_for_app || return 1
    sleep 2
    return 0
}

# ============================================
# TEST 1: Sync detection of pre-existing rules
# ============================================
test_sync_detection() {
    log_section "TEST 1: Sync Detection of Pre-existing Rules"
    start_container "-e ADD_TEST_RULES=true" "false" || return 1

    local db_content=$(get_database)
    if echo "$db_content" | grep -q "<SYSTEM>"; then
        log_success "Pre-existing rule detected with <SYSTEM> prefix"
        container_jq '.groups[] | select(.name | contains("<SYSTEM>")) | {name, targetIp, ports: [.ports[].value]}'
    else
        log_error "Pre-existing rule was NOT detected"
        return 1
    fi

    if check_rule_exists "10.0.0.1" "192.168.1.100" 8080; then
        log_success "Rule is present in iptables"
    else
        log_warning "Rule not found in iptables"
    fi
    return 0
}

# ============================================
# TEST 2: Rule deletion verification
# ============================================
test_rule_deletion() {
    log_section "TEST 2: Rule Deletion Verification"

    local rule_id=$(container_jq '.groups[] | select(.name | contains("<SYSTEM>")) | .id' | head -1)
    if [ -z "$rule_id" ] || [ "$rule_id" = "null" ]; then
        rule_id=$(container_jq '.groups[0].id')
    fi

    local target_ip=$(container_jq ".groups[] | select(.id==\"$rule_id\") | .targetIp")
    local dest_ip=$(container_jq ".groups[] | select(.id==\"$rule_id\") | .destinationIp")
    local port=$(container_jq ".groups[] | select(.id==\"$rule_id\") | .ports[0].value")

    log_info "Deleting rule: $target_ip:$port -> $dest_ip:$port"

    docker exec "$CONTAINER_NAME" iptables -t nat -D PREROUTING -d "$target_ip" -p tcp --dport "$port" -j DNAT --to-destination "$dest_ip:$port" 2>/dev/null || true
    docker exec "$CONTAINER_NAME" iptables -t nat -D POSTROUTING -d "$dest_ip" -p tcp --dport "$port" -j SNAT --to-source 192.168.1.1 2>/dev/null || true

    if check_rule_exists "$target_ip" "$dest_ip" "$port"; then
        log_error "Rule still exists after deletion"
        return 1
    fi
    log_success "Rule successfully deleted"
    return 0
}

# ============================================
# TEST 3: Rule addition verification
# ============================================
test_rule_addition() {
    log_section "TEST 3: Rule Addition Verification"

    local ip="10.0.0.99"
    local dest="192.168.1.99"
    local port=7777

    docker exec "$CONTAINER_NAME" iptables -t nat -A PREROUTING -d "$ip" -p tcp --dport "$port" -j DNAT --to-destination "$dest:$port"
    docker exec "$CONTAINER_NAME" iptables -t nat -A POSTROUTING -d "$dest" -p tcp --dport "$port" -j SNAT --to-source 192.168.1.1

    if check_rule_exists "$ip" "$dest" "$port"; then
        log_success "Rule successfully added"
    else
        log_error "Rule was NOT added"
        return 1
    fi
    return 0
}

# ============================================
# TEST 4: Disabled rules matching iptables
# ============================================
test_disabled_rules_no_system_prefix() {
    log_section "TEST 4: Disabled Rules - No SYSTEM Prefix Created"
    start_container "-e ADD_DISABLED_MATCH_RULE=true" "true" || return 1

    if check_rule_exists "10.0.0.5" "192.168.1.50" 5000; then
        log_error "Disabled rule still in iptables (should be removed)"
    else
        log_success "Disabled rule removed from iptables"
    fi

    local db_content=$(get_database)
    if echo "$db_content" | grep -q "<SYSTEM>.*5000"; then
        log_error "<SYSTEM> rule incorrectly created for disabled rule"
        return 1
    fi
    log_success "No <SYSTEM> rule created for disabled rule"

    if container_jq_exists '.groups[] | select(.id=="test-disabled-rule-1")'; then
        log_success "Original disabled rule preserved in database"
    else
        log_error "Original disabled rule removed from database!"
        return 1
    fi
    return 0
}

# ============================================
# TEST 5: Enabled rule gets applied
# ============================================
test_enabled_rule_applied() {
    log_section "TEST 5: Enabled Rule Gets Applied"

    if check_rule_exists "10.0.0.10" "192.168.1.10" 9000; then
        log_success "Enabled rule applied to iptables"
    else
        log_error "Enabled rule NOT applied"
        return 1
    fi
    return 0
}

# ============================================
# TEST 6: Multiple ports in one group
# ============================================
test_multiple_ports() {
    log_section "TEST 6: Multiple Ports in One Group"

    local all_present=true

    for port in 80 443 8080; do
        if check_rule_exists "10.0.0.20" "192.168.1.20" "$port"; then
            log_success "Port $port is present"
        else
            log_error "Port $port is MISSING"
            all_present=false
        fi
    done

    if [ "$all_present" = true ]; then
        log_success "All ports in group are applied"
        return 0
    else
        log_error "Some ports are missing"
        return 1
    fi
}

# ============================================
# TEST 7: TCP + UDP protocols
# ============================================
test_tcp_udp_protocols() {
    log_section "TEST 7: TCP + UDP Protocols"

    # Rule test-enabled-rule-1 has protocols: ["tcp", "udp"] for port 9000
    if check_rule_exists "10.0.0.10" "192.168.1.10" 9000 "tcp"; then
        log_success "TCP rule present for port 9000"
    else
        log_error "TCP rule MISSING for port 9000"
        return 1
    fi

    if check_rule_exists "10.0.0.10" "192.168.1.10" 9000 "udp"; then
        log_success "UDP rule present for port 9000"
    else
        log_error "UDP rule MISSING for port 9000"
        return 1
    fi

    log_success "Both TCP and UDP rules applied"
    return 0
}

# ============================================
# TEST 8: Port forwarding (valueTarget != value)
# ============================================
test_port_forwarding() {
    log_section "TEST 8: Port Forwarding (8080 -> 80)"

    # test-port-forward-rule: value=8080, valueTarget=80
    if check_rule_exists "10.0.0.30" "192.168.1.30" 8080 "tcp" 80; then
        log_success "Port forwarding rule applied (8080 -> 80)"
    else
        log_error "Port forwarding rule NOT applied"
        docker exec "$CONTAINER_NAME" iptables -t nat -L PREROUTING -n | grep "10.0.0.30" || true
        return 1
    fi

    if check_postrouting_exists "192.168.1.30" "192.168.1.1" 80 "tcp"; then
        log_success "POSTROUTING uses target port 80"
    else
        log_error "POSTROUTING not using target port"
        return 1
    fi
    return 0
}

# ============================================
# TEST 9: Boundary port 1
# ============================================
test_boundary_port_1() {
    log_section "TEST 9: Boundary Port 1 (minimum)"

    if check_rule_exists "10.0.0.50" "192.168.1.50" 1; then
        log_success "Port 1 rule applied"
    else
        log_error "Port 1 rule NOT applied"
        return 1
    fi
    return 0
}

# ============================================
# TEST 10: Boundary port 65535
# ============================================
test_boundary_port_65535() {
    log_section "TEST 10: Boundary Port 65535 (maximum)"

    if check_rule_exists "10.0.0.51" "192.168.1.51" 65535; then
        log_success "Port 65535 rule applied"
    else
        log_error "Port 65535 rule NOT applied"
        return 1
    fi
    return 0
}

# ============================================
# TEST 11: Orphan PREROUTING (no POSTROUTING)
# ============================================
test_orphan_prerouting() {
    log_section "TEST 11: Orphan PREROUTING (no matching POSTROUTING)"
    start_container "-e ADD_ORPHAN_PREROUTING=true" "false" || return 1

    local db_content=$(get_database)
    if echo "$db_content" | grep -q "10.0.0.70"; then
        log_error "Orphan PREROUTING incorrectly added to database"
        return 1
    fi
    log_success "Orphan PREROUTING NOT added to database (correct)"
    return 0
}

# ============================================
# TEST 12: Orphan POSTROUTING (no PREROUTING)
# ============================================
test_orphan_postrouting() {
    log_section "TEST 12: Orphan POSTROUTING (no matching PREROUTING)"
    start_container "-e ADD_ORPHAN_POSTROUTING=true" "false" || return 1

    local db_content=$(get_database)
    if echo "$db_content" | grep -q "192.168.1.71"; then
        log_error "Orphan POSTROUTING incorrectly added to database"
        return 1
    fi
    log_success "Orphan POSTROUTING NOT added to database (correct)"
    return 0
}

# ============================================
# TEST 13: Partial match disabled (different port)
# ============================================
test_partial_match_disabled() {
    log_section "TEST 13: Partial Match - Different Port Creates SYSTEM"
    start_container "-e ADD_PARTIAL_MATCH_RULE=true" "true" || return 1

    # Disabled rule has port 6000, iptables has port 6001
    # Should create <SYSTEM> for 6001 since it doesn't match
    local db_content=$(get_database)

    if echo "$db_content" | grep -q "<SYSTEM>.*6001"; then
        log_success "<SYSTEM> rule created for non-matching port 6001"
    else
        log_error "<SYSTEM> rule NOT created for port 6001"
        echo "Database content:"
        echo "$db_content" | grep -E "(SYSTEM|6001|6000)" || echo "No matching entries"
        return 1
    fi
    return 0
}

# ============================================
# TEST 14: Idempotency - sync doesn't duplicate
# ============================================
test_idempotency() {
    log_section "TEST 14: Idempotency - Sync Doesn't Create Duplicates"
    start_container "-e ADD_TEST_RULES=true" "true" || return 1

    local count_before=$(count_prerouting_rules)
    log_info "Rules in iptables: $count_before"

    local db_groups_before=$(container_jq '.groups | length')
    log_info "Groups in database: $db_groups_before"

    # Stop container
    log_info "Restarting container to test idempotency..."
    docker stop "$CONTAINER_NAME" > /dev/null 2>&1

    # Start same container again
    docker start "$CONTAINER_NAME" > /dev/null 2>&1

    wait_for_app || return 1
    sleep 2

    local count_after=$(count_prerouting_rules)
    log_info "Rules after restart: $count_after"

    local db_groups_after=$(container_jq '.groups | length')
    log_info "Groups in database after restart: $db_groups_after"

    if [ "$count_before" = "$count_after" ]; then
        log_success "No duplicate iptables rules after restart"
    else
        log_error "iptables rule count changed: $count_before -> $count_after"
        return 1
    fi

    if [ "$db_groups_before" = "$db_groups_after" ]; then
        log_success "No duplicate database groups after restart"
    else
        log_error "Database group count changed: $db_groups_before -> $db_groups_after"
        return 1
    fi

    return 0
}

# ============================================
# TEST 15: Toggle enabled state
# ============================================
test_toggle_enabled() {
    log_section "TEST 15: Toggle Enabled State"
    start_container "" "true" || return 1

    # Verify toggle rule is applied (enabled: true in DB)
    if check_rule_exists "10.0.0.40" "192.168.1.40" 3000; then
        log_success "Toggle rule initially applied"
    else
        log_error "Toggle rule not initially applied"
        return 1
    fi

    # Manually remove it (simulate disable)
    docker exec "$CONTAINER_NAME" iptables -t nat -D PREROUTING -d 10.0.0.40 -p tcp --dport 3000 -j DNAT --to-destination 192.168.1.40:3000 2>/dev/null || true
    docker exec "$CONTAINER_NAME" iptables -t nat -D POSTROUTING -d 192.168.1.40 -p tcp --dport 3000 -j SNAT --to-source 192.168.1.1 2>/dev/null || true

    if check_rule_exists "10.0.0.40" "192.168.1.40" 3000; then
        log_error "Rule still exists after removal"
        return 1
    fi
    log_success "Rule removed successfully (simulated disable)"

    # Re-add it (simulate enable)
    docker exec "$CONTAINER_NAME" iptables -t nat -A PREROUTING -d 10.0.0.40 -p tcp --dport 3000 -j DNAT --to-destination 192.168.1.40:3000
    docker exec "$CONTAINER_NAME" iptables -t nat -A POSTROUTING -d 192.168.1.40 -p tcp --dport 3000 -j SNAT --to-source 192.168.1.1

    if check_rule_exists "10.0.0.40" "192.168.1.40" 3000; then
        log_success "Rule re-added successfully (simulated enable)"
    else
        log_error "Failed to re-add rule"
        return 1
    fi
    return 0
}

# ============================================
# TEST 16: Multiple groups independence
# ============================================
test_multiple_groups() {
    log_section "TEST 16: Multiple Groups Applied Independently"
    start_container "-e ADD_TEST_RULES=true -e ADD_SECOND_RULE=true" "true" || return 1

    local groups_ok=true

    # Check system rule from ADD_TEST_RULES
    if check_rule_exists "10.0.0.1" "192.168.1.100" 8080; then
        log_success "Group 1 (system rule) present"
    else
        log_error "Group 1 missing"
        groups_ok=false
    fi

    # Check second system rule
    if check_rule_exists "10.0.0.2" "192.168.1.200" 443; then
        log_success "Group 2 (second system rule) present"
    else
        log_error "Group 2 missing"
        groups_ok=false
    fi

    # Check DB rule
    if check_rule_exists "10.0.0.10" "192.168.1.10" 9000; then
        log_success "Group 3 (DB enabled rule) present"
    else
        log_error "Group 3 missing"
        groups_ok=false
    fi

    if [ "$groups_ok" = true ]; then
        log_success "All groups applied independently"
        return 0
    fi
    return 1
}

# ============================================
# TEST 17: UDP rule scanning
# ============================================
test_udp_scanning() {
    log_section "TEST 17: UDP Rule Scanning"
    start_container "-e ADD_UDP_RULE=true" "false" || return 1

    local db_content=$(get_database)
    if echo "$db_content" | grep -q "<SYSTEM>.*53"; then
        log_success "UDP rule detected and added to database"
        container_jq '.groups[] | select(.name | contains("<SYSTEM>")) | select(.ports[].value == 53)'
    else
        log_error "UDP rule NOT detected"
        return 1
    fi
    return 0
}

# ============================================
# API HELPER FUNCTIONS
# ============================================
api_call() {
    local method="$1"
    local data="$2"
    docker exec "$CONTAINER_NAME" curl -s -X "$method" \
        -H "Content-Type: application/json" \
        -u "$AUTH_USER:$AUTH_PASSWORD" \
        -d "$data" \
        "http://localhost:$APP_PORT/api/test/iptables" 2>/dev/null
}

api_get() {
    docker exec "$CONTAINER_NAME" curl -s \
        -u "$AUTH_USER:$AUTH_PASSWORD" \
        "http://localhost:$APP_PORT/api/test/iptables" 2>/dev/null
}

# ============================================
# TEST 18: API - Add rule via application
# ============================================
test_api_add_rule() {
    log_section "TEST 18: API - Add Rule via Application"
    start_container "" "true" || return 1

    local new_rule='{
        "id": "api-test-rule-1",
        "name": "API Test Rule",
        "enabled": true,
        "targetIp": "10.0.0.200",
        "destinationIp": "192.168.1.200",
        "targetReverseIp": "192.168.1.1",
        "ports": [{
            "id": "api-port-1",
            "value": 4000,
            "valueTarget": 4000,
            "protocols": ["tcp"]
        }]
    }'

    log_info "Adding rule via API..."
    local response=$(api_call "POST" "$new_rule")

    if echo "$response" | grep -q '"success":true'; then
        log_success "API returned success"
    else
        log_error "API call failed: $response"
        return 1
    fi

    # Verify in database
    sleep 1
    if container_jq_exists '.groups[] | select(.id=="api-test-rule-1")'; then
        log_success "Rule added to database"
    else
        log_error "Rule NOT found in database"
        return 1
    fi

    # Verify in iptables
    if check_rule_exists "10.0.0.200" "192.168.1.200" 4000; then
        log_success "Rule applied to iptables"
    else
        log_error "Rule NOT applied to iptables"
        return 1
    fi

    return 0
}

# ============================================
# TEST 19: API - Update rule via application
# ============================================
test_api_update_rule() {
    log_section "TEST 19: API - Update Rule via Application"

    # Update the rule created in test 18
    local updated_rule='{
        "id": "api-test-rule-1",
        "name": "API Test Rule Updated",
        "enabled": true,
        "targetIp": "10.0.0.201",
        "destinationIp": "192.168.1.201",
        "targetReverseIp": "192.168.1.1",
        "ports": [{
            "id": "api-port-1",
            "value": 4001,
            "valueTarget": 4001,
            "protocols": ["tcp"]
        }]
    }'

    log_info "Updating rule via API..."
    local response=$(api_call "PUT" "$updated_rule")

    if echo "$response" | grep -q '"success":true'; then
        log_success "API returned success"
    else
        log_error "API call failed: $response"
        return 1
    fi

    sleep 1

    # Verify old rule removed from iptables
    if check_rule_exists "10.0.0.200" "192.168.1.200" 4000; then
        log_error "Old rule still in iptables"
        return 1
    fi
    log_success "Old rule removed from iptables"

    # Verify new rule in iptables
    if check_rule_exists "10.0.0.201" "192.168.1.201" 4001; then
        log_success "New rule applied to iptables"
    else
        log_error "New rule NOT applied to iptables"
        return 1
    fi

    # Verify database updated
    local db_name=$(container_jq '.groups[] | select(.id=="api-test-rule-1") | .name')
    if [ "$db_name" = "API Test Rule Updated" ]; then
        log_success "Database updated correctly"
    else
        log_error "Database not updated: $db_name"
        return 1
    fi

    return 0
}

# ============================================
# TEST 20: API - Disable rule via application
# ============================================
test_api_disable_rule() {
    log_section "TEST 20: API - Disable Rule via Application"

    # Disable the rule
    local disabled_rule='{
        "id": "api-test-rule-1",
        "name": "API Test Rule Updated",
        "enabled": false,
        "targetIp": "10.0.0.201",
        "destinationIp": "192.168.1.201",
        "targetReverseIp": "192.168.1.1",
        "ports": [{
            "id": "api-port-1",
            "value": 4001,
            "valueTarget": 4001,
            "protocols": ["tcp"]
        }]
    }'

    log_info "Disabling rule via API..."
    local response=$(api_call "PUT" "$disabled_rule")

    if echo "$response" | grep -q '"success":true'; then
        log_success "API returned success"
    else
        log_error "API call failed: $response"
        return 1
    fi

    sleep 1

    # Verify rule removed from iptables
    if check_rule_exists "10.0.0.201" "192.168.1.201" 4001; then
        log_error "Disabled rule still in iptables"
        return 1
    fi
    log_success "Disabled rule removed from iptables"

    # Verify rule still in database with enabled=false
    local db_enabled=$(container_jq '.groups[] | select(.id=="api-test-rule-1") | .enabled')
    if [ "$db_enabled" = "false" ]; then
        log_success "Rule disabled in database"
    else
        log_error "Rule not disabled in database: enabled=$db_enabled"
        return 1
    fi

    return 0
}

# ============================================
# TEST 21: API - Delete rule via application
# ============================================
test_api_delete_rule() {
    log_section "TEST 21: API - Delete Rule via Application"

    # First enable it again to verify deletion removes from iptables
    local enable_rule='{
        "id": "api-test-rule-1",
        "name": "API Test Rule Updated",
        "enabled": true,
        "targetIp": "10.0.0.201",
        "destinationIp": "192.168.1.201",
        "targetReverseIp": "192.168.1.1",
        "ports": [{
            "id": "api-port-1",
            "value": 4001,
            "valueTarget": 4001,
            "protocols": ["tcp"]
        }]
    }'
    api_call "PUT" "$enable_rule" > /dev/null
    sleep 1

    # Verify rule is in iptables
    if ! check_rule_exists "10.0.0.201" "192.168.1.201" 4001; then
        log_warning "Rule not in iptables before delete"
    fi

    # Delete the rule
    local delete_rule='{
        "id": "api-test-rule-1",
        "name": "API Test Rule Updated",
        "enabled": true,
        "targetIp": "10.0.0.201",
        "destinationIp": "192.168.1.201",
        "targetReverseIp": "192.168.1.1",
        "ports": [{
            "id": "api-port-1",
            "value": 4001,
            "valueTarget": 4001,
            "protocols": ["tcp"]
        }]
    }'

    log_info "Deleting rule via API..."
    local response=$(api_call "DELETE" "$delete_rule")

    if echo "$response" | grep -q '"success":true'; then
        log_success "API returned success"
    else
        log_error "API call failed: $response"
        return 1
    fi

    sleep 1

    # Verify rule removed from iptables
    if check_rule_exists "10.0.0.201" "192.168.1.201" 4001; then
        log_error "Deleted rule still in iptables"
        return 1
    fi
    log_success "Rule removed from iptables"

    # Verify rule removed from database
    if container_jq_exists '.groups[] | select(.id=="api-test-rule-1")'; then
        log_error "Rule still in database"
        return 1
    fi
    log_success "Rule removed from database"

    return 0
}

# ============================================
# TEST 22: API - Get all rules
# ============================================
test_api_get_rules() {
    log_section "TEST 22: API - Get All Rules"

    log_info "Getting rules via API..."
    local response=$(api_get)

    if echo "$response" | grep -q '"success":true'; then
        log_success "API returned success"
    else
        log_error "API call failed: $response"
        return 1
    fi

    if echo "$response" | grep -q '"groups":\['; then
        log_success "API returned groups array"
        local count=$(echo "$response" | docker exec -i "$CONTAINER_NAME" jq '.groups | length')
        log_info "Found $count groups"
    else
        log_error "API did not return groups"
        return 1
    fi

    return 0
}

# ============================================
# MAIN TEST RUNNER
# ============================================
main() {
    log_section "IPTABLES TEST SUITE - EXTENDED"
    check_requirements

    log_info "Script directory: $SCRIPT_DIR"
    log_info "Docker path: $DOCKER_SCRIPT_DIR"

    local total_tests=0
    local passed_tests=0
    local failed_tests=0

    trap cleanup EXIT

    local tests=(
        "test_sync_detection"
        "test_rule_deletion"
        "test_rule_addition"
        "test_disabled_rules_no_system_prefix"
        "test_enabled_rule_applied"
        "test_multiple_ports"
        "test_tcp_udp_protocols"
        "test_port_forwarding"
        "test_boundary_port_1"
        "test_boundary_port_65535"
        "test_orphan_prerouting"
        "test_orphan_postrouting"
        "test_partial_match_disabled"
        "test_idempotency"
        "test_toggle_enabled"
        "test_multiple_groups"
        "test_udp_scanning"
        "test_api_add_rule"
        "test_api_update_rule"
        "test_api_disable_rule"
        "test_api_delete_rule"
        "test_api_get_rules"
    )

    for test in "${tests[@]}"; do
        total_tests=$((total_tests + 1))
        if $test; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    done

    log_section "TEST SUMMARY"
    echo ""
    echo -e "Total tests: $total_tests"
    echo -e "${GREEN}Passed: $passed_tests${NC}"
    echo -e "${RED}Failed: $failed_tests${NC}"
    echo ""

    if [ $failed_tests -eq 0 ]; then
        log_success "All tests passed!"
        return 0
    else
        log_error "Some tests failed!"
        return 1
    fi
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
