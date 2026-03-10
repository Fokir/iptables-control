#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[+]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[x]${NC} $1"; exit 1; }
info() { echo -e "${CYAN}[i]${NC} $1"; }

# ─────────────────────────────────────────────
# Checks
# ─────────────────────────────────────────────

command -v git >/dev/null 2>&1 || err "git is not installed"
command -v gh >/dev/null 2>&1  || warn "gh (GitHub CLI) not found — release will be created via git tag only"

git rev-parse --is-inside-work-tree >/dev/null 2>&1 || err "Not inside a git repository"

BRANCH=$(git rev-parse --abbrev-ref HEAD)
DEFAULT_BRANCH=$(git remote show origin 2>/dev/null | grep 'HEAD branch' | awk '{print $NF}')
DEFAULT_BRANCH=${DEFAULT_BRANCH:-main}

info "Current branch: $BRANCH"
info "Default branch: $DEFAULT_BRANCH"

# ─────────────────────────────────────────────
# Determine next version
# ─────────────────────────────────────────────

LATEST_TAG=$(git tag -l 'v*' --sort=-v:refname | head -1)

if [ -z "$LATEST_TAG" ]; then
    CURRENT="0.0.0"
    info "No existing tags found, starting from v0.0.0"
else
    CURRENT="${LATEST_TAG#v}"
    info "Latest release: $LATEST_TAG"
fi

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

NEXT_PATCH="v${MAJOR}.${MINOR}.$((PATCH + 1))"
NEXT_MINOR="v${MAJOR}.$((MINOR + 1)).0"
NEXT_MAJOR="v$((MAJOR + 1)).0.0"

echo ""
echo -e "  ${CYAN}1)${NC} patch  → ${GREEN}$NEXT_PATCH${NC}"
echo -e "  ${CYAN}2)${NC} minor  → ${GREEN}$NEXT_MINOR${NC}"
echo -e "  ${CYAN}3)${NC} major  → ${GREEN}$NEXT_MAJOR${NC}"
echo -e "  ${CYAN}4)${NC} custom"
echo ""

read -p "Select version type [1]: " VERSION_TYPE
VERSION_TYPE=${VERSION_TYPE:-1}

case "$VERSION_TYPE" in
    1) NEXT_VERSION="$NEXT_PATCH" ;;
    2) NEXT_VERSION="$NEXT_MINOR" ;;
    3) NEXT_VERSION="$NEXT_MAJOR" ;;
    4)
        read -p "Enter version (e.g. v1.2.3): " NEXT_VERSION
        if [[ ! "$NEXT_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            err "Invalid version format. Expected: v1.2.3"
        fi
        ;;
    *) err "Invalid selection" ;;
esac

info "Next version: $NEXT_VERSION"

# ─────────────────────────────────────────────
# Show changes since last release
# ─────────────────────────────────────────────

echo ""
info "Changes since ${LATEST_TAG:-beginning}:"
echo ""

if [ -n "$LATEST_TAG" ]; then
    git log "${LATEST_TAG}..HEAD" --oneline --no-decorate
else
    git log --oneline --no-decorate -20
fi

echo ""

# ─────────────────────────────────────────────
# Check for uncommitted changes
# ─────────────────────────────────────────────

if [ -n "$(git status --porcelain)" ]; then
    warn "There are uncommitted changes:"
    git status --short
    echo ""
    read -p "Commit all changes before release? [y/N]: " DO_COMMIT
    if [[ "$DO_COMMIT" =~ ^[Yy]$ ]]; then
        read -p "Commit message [release: $NEXT_VERSION]: " COMMIT_MSG
        COMMIT_MSG=${COMMIT_MSG:-"release: $NEXT_VERSION"}
        git add -A
        git commit -m "$COMMIT_MSG"
        log "Changes committed"
    else
        err "Aborting — commit or stash your changes first"
    fi
fi

# ─────────────────────────────────────────────
# Confirm
# ─────────────────────────────────────────────

echo ""
echo -e "  Version:  ${GREEN}$NEXT_VERSION${NC}"
echo -e "  Branch:   ${CYAN}$BRANCH${NC}"
echo -e "  Push to:  origin/$BRANCH"
echo -e "  Tag:      $NEXT_VERSION → triggers CI build & release"
echo ""

read -p "Proceed? [y/N]: " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    err "Aborted"
fi

# ─────────────────────────────────────────────
# Push changes
# ─────────────────────────────────────────────

log "Pushing branch $BRANCH..."
git push origin "$BRANCH"

# ─────────────────────────────────────────────
# Create and push tag
# ─────────────────────────────────────────────

log "Creating tag $NEXT_VERSION..."
git tag -a "$NEXT_VERSION" -m "Release $NEXT_VERSION"

log "Pushing tag..."
git push origin "$NEXT_VERSION"

# ─────────────────────────────────────────────
# Wait for CI and show status
# ─────────────────────────────────────────────

echo ""
log "Release $NEXT_VERSION initiated!"

if command -v gh >/dev/null 2>&1; then
    REPO=$(gh repo view --json nameWithOwner -q '.nameWithOwner' 2>/dev/null || echo "")

    if [ -n "$REPO" ]; then
        echo ""
        info "GitHub Actions will build binaries and create the release."
        info "Monitor: https://github.com/$REPO/actions"
        echo ""

        read -p "Wait for CI to finish? [y/N]: " WAIT_CI
        if [[ "$WAIT_CI" =~ ^[Yy]$ ]]; then
            log "Waiting for CI run to appear..."
            sleep 5

            for i in $(seq 1 10); do
                RUN_ID=$(gh run list --repo "$REPO" --limit 1 --json databaseId,headBranch,status -q ".[0].databaseId" 2>/dev/null)
                if [ -n "$RUN_ID" ]; then
                    break
                fi
                sleep 3
            done

            if [ -n "$RUN_ID" ]; then
                log "Watching CI run #$RUN_ID..."
                gh run watch "$RUN_ID" --repo "$REPO" && log "CI completed successfully!" || warn "CI run failed. Check: https://github.com/$REPO/actions/runs/$RUN_ID"

                echo ""
                RELEASE_URL="https://github.com/$REPO/releases/tag/$NEXT_VERSION"
                info "Release: $RELEASE_URL"
            else
                warn "Could not find CI run. Check manually: https://github.com/$REPO/actions"
            fi
        fi
    fi
fi

echo ""
log "Done! Release $NEXT_VERSION"
