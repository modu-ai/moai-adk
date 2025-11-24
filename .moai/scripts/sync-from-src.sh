#!/bin/bash

# Sync script: Manual file copy (Safe synchronization)
# Source: src/moai_adk/ → Local: ./
# Updated: 2025-11-24

set -e  # Exit on error

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"
SRC_DIR="$PROJECT_ROOT/src/moai_adk"

echo "🔄 Starting manual sync from $SRC_DIR"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter
COPIED=0
SKIPPED=0

# Function to copy file
copy_file() {
    local src="$1"
    local dest="$2"

    if [ -f "$src" ]; then
        mkdir -p "$(dirname "$dest")"
        cp "$src" "$dest"
        echo -e "${GREEN}✓${NC} $(basename "$dest")"
        ((COPIED++))
    else
        echo -e "${YELLOW}⊘${NC} $(basename "$src") not found"
        ((SKIPPED++))
    fi
}

# Function to copy directory
copy_dir() {
    local src="$1"
    local dest="$2"
    local name="$3"

    if [ -d "$src" ]; then
        mkdir -p "$dest"
        cp -r "$src"/* "$dest/" 2>/dev/null || true
        echo -e "${GREEN}✓${NC} $name/"
        ((COPIED++))
    else
        echo -e "${YELLOW}⊘${NC} $name/ not found"
        ((SKIPPED++))
    fi
}

echo "📌 Syncing .claude/agents/moai/ (Agent definitions)..."
copy_dir "$SRC_DIR/.claude/agents/moai" "$PROJECT_ROOT/.claude/agents/moai" ".claude/agents/moai"
echo ""

echo "📌 Syncing .claude/skills/ (Skill definitions)..."
copy_dir "$SRC_DIR/.claude/skills" "$PROJECT_ROOT/.claude/skills" ".claude/skills"
echo ""

echo "📌 Syncing .claude/commands/ (Custom commands)..."
if [ -d "$SRC_DIR/.claude/commands" ]; then
    copy_dir "$SRC_DIR/.claude/commands" "$PROJECT_ROOT/.claude/commands" ".claude/commands"
else
    echo -e "${YELLOW}⊘${NC} .claude/commands/ not found (optional)"
fi
echo ""

echo "📌 Syncing .moai/memory/ (Reference documentation)..."
copy_dir "$SRC_DIR/.moai/memory" "$PROJECT_ROOT/.moai/memory" ".moai/memory"
echo ""

echo "📌 Syncing .moai/project/ (Project structure)..."
if [ -d "$SRC_DIR/.moai/project" ]; then
    copy_dir "$SRC_DIR/.moai/project" "$PROJECT_ROOT/.moai/project" ".moai/project"
fi
echo ""

echo "📌 Syncing CLAUDE.md (Main directives)..."
copy_file "$SRC_DIR/CLAUDE.md" "$PROJECT_ROOT/CLAUDE.md"
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Sync Complete${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Files copied: $COPIED"
echo "Skipped: $SKIPPED"
echo ""
echo "📍 Next step: Review changes and commit if needed"
echo "   git status"
echo "   git add ..."
echo "   git commit -m 'chore: Sync from src/moai_adk/'"
