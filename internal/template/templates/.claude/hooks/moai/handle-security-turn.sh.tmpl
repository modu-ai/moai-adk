#!/bin/bash

MOAI_HOOK_STDERR_LOG="${MOAI_HOOK_STDERR_LOG:-$HOME/.moai/logs/hook-stderr.log}"
# Allowlist: if MOAI_HOOK_STDERR_LOG is set but does not start with an allowed
# prefix ($HOME/.moai/logs or $CLAUDE_PROJECT_DIR/.moai/logs), reset to default.
case "$MOAI_HOOK_STDERR_LOG" in
    "$HOME/.moai/logs/"*|"$CLAUDE_PROJECT_DIR/.moai/logs/"*|/dev/null) ;;
    *) MOAI_HOOK_STDERR_LOG="$HOME/.moai/logs/hook-stderr.log" ;;
esac

mkdir -p "$(dirname "$MOAI_HOOK_STDERR_LOG")" 2>/dev/null || true
# MoAI Security Guardian - Layer 2 (Stop turn-diff review) wrapper.
# Forwards stdin JSON to `moai hook security-turn`, which runs the regex engine
# over the turn's working-tree diff and surfaces high-severity findings via
# systemMessage. Advisory by default; blocking is opt-in via MOAI_SECURITY_BLOCKING.
# A `--skip-hook` first argument bypasses the gate (audit-logged). Thin forwarder.

# Single-level rotation at 10MB (best-effort, non-blocking)
if [ -f "$MOAI_HOOK_STDERR_LOG" ]; then
    hook_log_size=$(stat -f%z "$MOAI_HOOK_STDERR_LOG" 2>/dev/null || stat -c%s "$MOAI_HOOK_STDERR_LOG" 2>/dev/null || echo 0)
    if [ "$hook_log_size" -gt 10485760 ]; then
        mv -f "$MOAI_HOOK_STDERR_LOG" "${MOAI_HOOK_STDERR_LOG}.1" 2>/dev/null || true
    fi
fi

# Try moai command in PATH
if command -v moai &> /dev/null; then
    exec moai hook security-turn "$@" 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try default ~/go/bin/moai
if [ -f "$HOME/go/bin/moai" ]; then
    exec "$HOME/go/bin/moai" hook security-turn "$@" 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try ~/.local/bin/moai (Linux install location)
if [ -f "$HOME/.local/bin/moai" ]; then
    exec "$HOME/.local/bin/moai" hook security-turn "$@" 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Not found - exit silently (Claude Code handles missing hooks gracefully)
exit 0
