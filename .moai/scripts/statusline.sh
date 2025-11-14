#!/usr/bin/env bash
# MoAI-ADK Statusline for Claude Code
#
# This shell wrapper is CRITICAL for proper Claude Code statusline integration.
# Claude Code sends JSON via STDIN, which MUST be forwarded to the Python module.
# Direct "uv run" does not properly pipe STDIN, so this wrapper is required.
#
# Features:
# - Receives JSON input from Claude Code via STDIN
# - Pipes JSON to Python module for processing
# - Works in local development and package distribution scenarios
# - Graceful fallback to config-based display on errors
#
# Usage (called by Claude Code):
#   echo '{"model": {...}, ...}' | .moai/scripts/statusline.sh
#
# Or from shell directly (for debugging):
#   .moai/scripts/statusline.sh

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'  # No Color

# Read JSON input from Claude Code (CRITICAL: read once and reuse)
INPUT_JSON=$(cat)

# Get working directory (priority: CLAUDE_PROJECT_DIR > argument > current dir)
WORKING_DIR="${CLAUDE_PROJECT_DIR:-${1:-.}}"

# Script location for finding project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Debug mode (set MOAI_DEBUG=1 to enable)
DEBUG=${MOAI_DEBUG:-0}

debug_log() {
    if [ "$DEBUG" = "1" ]; then
        echo "[DEBUG] $*" >&2
    fi
}

# Verify we have the necessary tools
check_dependencies() {
    if ! command -v uv >/dev/null 2>&1; then
        debug_log "uv command not found"
        return 1
    fi
    return 0
}

# Try to run statusline via installed moai-adk package
run_via_package() {
    debug_log "Attempting to run via installed package..."

    if echo "$INPUT_JSON" | uv run python -m moai_adk.statusline.main 2>/dev/null; then
        debug_log "Successfully ran via installed package"
        return 0
    else
        debug_log "Failed to run via installed package"
        return 1
    fi
}

# Try to run statusline via local project
run_via_project() {
    debug_log "Attempting to run via local project..."

    if [ -f "$PROJECT_ROOT/pyproject.toml" ] && \
       grep -q "name.*=.*\"moai-adk\"" "$PROJECT_ROOT/pyproject.toml" 2>/dev/null; then
        debug_log "Found MoAI-ADK project at: $PROJECT_ROOT"

        if echo "$INPUT_JSON" | uv run --project "$PROJECT_ROOT" python -m moai_adk.statusline.main 2>/dev/null; then
            debug_log "Successfully ran via local project"
            return 0
        fi
    fi

    debug_log "Could not run via local project"
    return 1
}

# Fallback to config-based display
fallback_display() {
    debug_log "Using fallback display from config"

    # Look for config in working directory
    CONFIG_PATH="$WORKING_DIR/.moai/config/config.json"

    # Also try project root
    if [ ! -f "$CONFIG_PATH" ]; then
        CONFIG_PATH="$PROJECT_ROOT/.moai/config/config.json"
    fi

    if [ -f "$CONFIG_PATH" ]; then
        debug_log "Found config at: $CONFIG_PATH"

        # Try to extract version and project name using basic JSON parsing
        VERSION=$(python3 -c "import json; print(json.load(open('$CONFIG_PATH')).get('moai', {}).get('version', 'unknown'))" 2>/dev/null || echo "unknown")
        PROJECT=$(python3 -c "import json; print(json.load(open('$CONFIG_PATH')).get('project', {}).get('name', 'MoAI-ADK'))" 2>/dev/null || echo "MoAI-ADK")

        echo "📦 Version: $VERSION (fallback mode)  🏗️  Project: $PROJECT"
        return 0
    else
        debug_log "Config file not found at: $CONFIG_PATH"
        echo "⚠️  MoAI-ADK not configured - run moai-adk init first"
        return 1
    fi
}

# Main execution flow
main() {
    debug_log "=== Statusline Execution ==="
    debug_log "Working directory: $WORKING_DIR"
    debug_log "Project root: $PROJECT_ROOT"
    debug_log "Input JSON length: ${#INPUT_JSON}"

    # Check for required tools
    if ! check_dependencies; then
        debug_log "Dependencies check failed"
        fallback_display
        return $?
    fi

    # Try execution paths in order of preference

    # 1. Try package installation (faster, works when moai-adk is installed globally)
    if run_via_package; then
        debug_log "Success via package"
        return 0
    fi

    # 2. Try local project (works in development)
    if run_via_project; then
        debug_log "Success via project"
        return 0
    fi

    # 3. Fall back to config-based display
    debug_log "All execution attempts failed, using fallback"
    fallback_display
    return $?
}

# Run main function
main "$@"
