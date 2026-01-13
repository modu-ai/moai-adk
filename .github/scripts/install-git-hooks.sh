#!/bin/bash

################################################################################
# Git Hooks Installation Script
#
# Installs pre-push hook from .github/scripts/git-hooks/ to .git/hooks/
# This enables local type checking before push to prevent CI failures.
################################################################################

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔧 Installing Git Hooks...${NC}"

# Project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source and destination
SOURCE_HOOK="$PROJECT_ROOT/.github/scripts/git-hooks/pre-push"
DEST_HOOK="$PROJECT_ROOT/.git/hooks/pre-push"

# Check if source exists
if [ ! -f "$SOURCE_HOOK" ]; then
    echo -e "${YELLOW}⚠️  Source hook not found: $SOURCE_HOOK${NC}"
    echo "Skipping git hooks installation..."
    exit 0
fi

# Create backup if existing hook
if [ -f "$DEST_HOOK" ]; then
    BACKUP="$DEST_HOOK.backup.$(date +%s)"
    echo -e "${YELLOW}📦 Backing up existing hook to: $BACKUP${NC}"
    cp "$DEST_HOOK" "$BACKUP"
fi

# Copy hook
echo "📋 Installing pre-push hook..."
cp "$SOURCE_HOOK" "$DEST_HOOK"

# Make executable
chmod +x "$DEST_HOOK"

echo -e "${GREEN}✅ Git hooks installed successfully!${NC}"
echo ""
echo "📝 Installed hooks:"
echo "  • pre-push: Security detection + GitFlow rules + Type checking"
echo ""
echo "💡 To skip checks temporarily: git push --no-verify"
echo ""
