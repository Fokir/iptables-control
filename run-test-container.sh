#!/bin/bash

# ============================================
# Quick script to run the test container interactively
# Compatible with: Linux, macOS, Windows (Git Bash)
# ============================================

set -e

# Get script directory (works in Git Bash too)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Convert path for Docker on Windows (Git Bash)
get_docker_path() {
    local path="$1"
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "mingw"* ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        # Convert /c/Users/... to C:/Users/... for Docker on Windows
        echo "$path" | sed -e 's|^/\([a-zA-Z]\)/|\1:/|'
    else
        echo "$path"
    fi
}

DOCKER_SCRIPT_DIR="$(get_docker_path "$SCRIPT_DIR")"

echo "Building Docker image..."
docker build -t system-control-test "$SCRIPT_DIR"

echo ""
echo "Starting container with iptables support..."
echo "You can access the app at: http://localhost:3000"
echo "Login: admin / admin"
echo ""

# Use MSYS_NO_PATHCONV to prevent path conversion issues on Windows
MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL="*" docker run -it --rm \
    --name system-control-interactive \
    --cap-add NET_ADMIN \
    -e ADD_TEST_RULES=true \
    -e ADD_SECOND_RULE=true \
    -e ADD_DISABLED_MATCH_RULE=true \
    -e BASIC_AUTH_USER=admin \
    -e BASIC_AUTH_PASSWORD=admin \
    -v "$DOCKER_SCRIPT_DIR/test-data:/app/test-data:ro" \
    -p 3000:3000 \
    system-control-test
