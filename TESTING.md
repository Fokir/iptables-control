# Testing iptables Functionality

This document describes how to test the iptables management features of the application.

## Prerequisites

- Docker Desktop installed and running
- Bash shell (Git Bash on Windows, or native on Linux/macOS)
- `jq` command-line JSON processor
  - Windows (via scoop): `scoop install jq`
  - Windows (via choco): `choco install jq`
  - macOS: `brew install jq`
  - Linux: `apt install jq` or `yum install jq`

## Windows (Git Bash) Notes

All scripts are compatible with Git Bash on Windows. The scripts automatically:
- Convert paths from `/c/Users/...` to `C:/Users/...` format for Docker
- Use `MSYS_NO_PATHCONV=1` to prevent path mangling
- Handle Windows-specific Docker Desktop requirements

**Important:** Make sure Docker Desktop is running before executing tests.

## Quick Start

### 1. Build and Run Interactive Container

```bash
./run-test-container.sh
```

This will:
- Build the Docker image
- Start a container with NET_ADMIN capability (required for iptables)
- Add test iptables rules before application starts
- Start the application on port 3000

Access the app at: http://localhost:3000 (login: admin/admin)

### 2. Run Automated Test Suite

```bash
./test-iptables.sh
```

This runs all test scenarios automatically and reports results.

### 3. Check iptables Status

While the container is running:

```bash
./check-iptables.sh [container-name]
```

## Test Scenarios

### Test 1: Sync Detection of Pre-existing Rules

**Scenario:** Rules exist in iptables before application starts.

**Expected behavior:**
- Application scans iptables at startup
- Finds rules that don't exist in database
- Creates new database entries with `<SYSTEM>` prefix

**Test variables:**
```bash
ADD_TEST_RULES=true  # Adds rule 10.0.0.1:8080 -> 192.168.1.100:8080
```

### Test 2: Rule Deletion Verification

**Scenario:** Delete a rule and verify it's removed from iptables.

**Expected behavior:**
- Rule is removed from iptables
- Rule is removed from database
- `iptables -C` command returns error (rule not found)

### Test 3: Rule Addition Verification

**Scenario:** Add a new rule and verify it appears in iptables.

**Expected behavior:**
- Rule is added to database
- Rule is added to iptables (PREROUTING + POSTROUTING)
- `iptables -C` command returns success

### Test 4: Disabled Rules Matching iptables

**Scenario:** Database has a disabled rule, and iptables has a matching rule.

**Expected behavior:**
1. `execSyncGroup` runs first and removes the rule from iptables (because it's disabled in DB)
2. `scanSystemRules` runs after and doesn't find the rule
3. No `<SYSTEM>` rule is created
4. Original disabled rule remains in database unchanged

**Test variables:**
```bash
ADD_DISABLED_MATCH_RULE=true  # Adds rule 10.0.0.5:5000 matching disabled DB entry
```

**Test database (`test-data/iptables.json`):**
```json
{
  "groups": [
    {
      "id": "test-disabled-rule-1",
      "name": "Disabled test rule (matching iptables)",
      "enabled": false,
      "targetIp": "10.0.0.5",
      "destinationIp": "192.168.1.50",
      "targetReverseIp": "192.168.1.1",
      "ports": [{"value": 5000, "protocols": ["tcp"]}]
    }
  ]
}
```

### Test 5: Enabled Rule Gets Applied

**Scenario:** Database has an enabled rule that doesn't exist in iptables.

**Expected behavior:**
- `execSyncGroup` adds the rule to iptables
- Rule can be verified with `iptables -C`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ADD_TEST_RULES` | Add test rule 10.0.0.1:8080 at startup | false |
| `ADD_SECOND_RULE` | Add additional test rule 10.0.0.2:443 | false |
| `ADD_DISABLED_MATCH_RULE` | Add rule matching disabled DB entry | false |
| `BASIC_AUTH_USER` | Basic auth username | admin |
| `BASIC_AUTH_PASSWORD` | Basic auth password | admin |

## Docker Commands Reference

### Build image
```bash
docker build -t system-control-test .
```

### Run with all test rules
```bash
docker run -it --rm \
    --cap-add NET_ADMIN \
    -e ADD_TEST_RULES=true \
    -e ADD_SECOND_RULE=true \
    -e ADD_DISABLED_MATCH_RULE=true \
    -v "$(pwd)/test-data:/app/test-data:ro" \
    -p 3000:3000 \
    system-control-test
```

### Check iptables inside container
```bash
docker exec <container> iptables -t nat -L PREROUTING -n -v
docker exec <container> iptables -t nat -L POSTROUTING -n -v
docker exec <container> iptables-save -t nat
```

### Check database inside container
```bash
docker exec <container> cat /app/database/iptables.json | jq .
```

### Execute bash inside container
```bash
docker exec -it <container> /bin/bash
```

## Sync Logic Explained

The application follows this sequence at startup (`src/instrumentation.ts`):

```
1. Load database
2. For each group in database:
   └── execSyncGroup(group)
       ├── If enabled=true AND not in iptables: ADD to iptables
       ├── If enabled=true AND in iptables: SKIP
       ├── If enabled=false AND in iptables: REMOVE from iptables
       └── If enabled=false AND not in iptables: SKIP
3. Scan iptables for existing rules (scanSystemRules)
4. For each system rule found:
   ├── If matches any DB group (by IPs and ports): SKIP
   └── If NOT in DB: ADD to DB with <SYSTEM> prefix
5. Save database if changed
6. Save iptables rules
```

**Key insight:** The sync removes disabled rules from iptables BEFORE scanning, so disabled rules won't create `<SYSTEM>` duplicates.
