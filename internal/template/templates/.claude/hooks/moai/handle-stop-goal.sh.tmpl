#!/bin/bash

# MoAI Hook Wrapper - handle-stop-goal.sh
# Forwards stdin JSON to `moai hook stop-goal` (the goal-engine Stop-hook
# evaluator). Stop hooks COMPOSE — this wrapper is registered as a SEPARATE
# entry alongside handle-stop.sh and sync-phase-quality-gate.sh; it does NOT
# replace them. Both wrappers read the same stdin JSON independently.

MOAI_HOOK_STDERR_LOG="${MOAI_HOOK_STDERR_LOG:-$HOME/.moai/logs/hook-stderr.log}"
case "$MOAI_HOOK_STDERR_LOG" in
    "$HOME/.moai/logs/"*|"$CLAUDE_PROJECT_DIR/.moai/logs/"*|/dev/null) ;;
    *) MOAI_HOOK_STDERR_LOG="$HOME/.moai/logs/hook-stderr.log" ;;
esac

mkdir -p "$(dirname "$MOAI_HOOK_STDERR_LOG")" 2>/dev/null || true

# Single-level rotation at 10MB (best-effort, non-blocking)
if [ -f "$MOAI_HOOK_STDERR_LOG" ]; then
    hook_log_size=$(stat -f%z "$MOAI_HOOK_STDERR_LOG" 2>/dev/null || stat -c%s "$MOAI_HOOK_STDERR_LOG" 2>/dev/null || echo 0)
    if [ "$hook_log_size" -gt 10485760 ]; then
        mv -f "$MOAI_HOOK_STDERR_LOG" "${MOAI_HOOK_STDERR_LOG}.1" 2>/dev/null || true
    fi
fi

# Try moai command in PATH
if command -v moai &> /dev/null; then
    exec moai hook stop-goal 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try default ~/go/bin/moai
if [ -f "$HOME/go/bin/moai" ]; then
    exec "$HOME/go/bin/moai" hook stop-goal 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try ~/.local/bin/moai (Linux install location)
if [ -f "$HOME/.local/bin/moai" ]; then
    exec "$HOME/.local/bin/moai" hook stop-goal 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Not found - exit silently (Claude Code handles missing hooks gracefully)
exit 0
