#!/bin/bash
# @CODE:DOC-TAG-004 | Component 1: Pre-commit hook installer
#
# This script installs the TAG validation pre-commit hook into .git/hooks/
#
# Usage:
#   ./install.sh              # Install hook
#   ./install.sh --uninstall  # Remove hook

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Get repository root
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo ".")
HOOK_SOURCE="${REPO_ROOT}/.moai/hooks/pre-commit.sh"
HOOK_TARGET="${REPO_ROOT}/.git/hooks/pre-commit"

# Function to install hook
install_hook() {
    echo "🔧 Installing TAG validation pre-commit hook..."

    # Check if source exists
    if [ ! -f "$HOOK_SOURCE" ]; then
        echo -e "${RED}Error: Hook source not found at $HOOK_SOURCE${NC}"
        exit 1
    fi

    # Check if target already exists
    if [ -f "$HOOK_TARGET" ]; then
        echo -e "${YELLOW}Warning: Pre-commit hook already exists.${NC}"
        echo "Backing up existing hook to ${HOOK_TARGET}.backup"
        cp "$HOOK_TARGET" "${HOOK_TARGET}.backup"
    fi

    # Create .git/hooks directory if it doesn't exist
    mkdir -p "$(dirname "$HOOK_TARGET")"

    # Copy hook
    cp "$HOOK_SOURCE" "$HOOK_TARGET"
    chmod +x "$HOOK_TARGET"

    echo -e "${GREEN}✓ Pre-commit hook installed successfully!${NC}"
    echo ""
    echo "The hook will now validate TAG annotations on every commit."
    echo ""
    echo "To uninstall: $0 --uninstall"
}

# Function to uninstall hook
uninstall_hook() {
    echo "🔧 Uninstalling TAG validation pre-commit hook..."

    if [ ! -f "$HOOK_TARGET" ]; then
        echo -e "${YELLOW}No pre-commit hook installed.${NC}"
        exit 0
    fi

    # Check if backup exists
    if [ -f "${HOOK_TARGET}.backup" ]; then
        echo "Restoring backup..."
        mv "${HOOK_TARGET}.backup" "$HOOK_TARGET"
        echo -e "${GREEN}✓ Backup restored.${NC}"
    else
        rm "$HOOK_TARGET"
        echo -e "${GREEN}✓ Pre-commit hook removed.${NC}"
    fi
}

# Parse command line arguments
if [ "$1" = "--uninstall" ]; then
    uninstall_hook
else
    install_hook
fi
