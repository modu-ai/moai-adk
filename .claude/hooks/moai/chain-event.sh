#!/bin/bash
# chain-event hook wrapper — appends completion edge to the chain ledger.
#
# Fires on SubagentStop. Reads the hook payload from stdin and passes it to
# `moai hook chain-event`, which appends a completion-edge event to the chain
# ledger. Fail-open: any error is silently swallowed (exit 0).
#
# The hook is registered in .claude/settings.json under the SubagentStop event.

# Determine the moai binary path. Priority:
# 1. CLAUDE_PROJECT_DIR/.moai/bin/moai (project-local)
# 2. $HOME/go/bin/moai (default Go bin)
# 3. PATH lookup
MOAI_BIN=""
if [ -n "$CLAUDE_PROJECT_DIR" ] && [ -x "$CLAUDE_PROJECT_DIR/.moai/bin/moai" ]; then
    MOAI_BIN="$CLAUDE_PROJECT_DIR/.moai/bin/moai"
elif [ -x "$HOME/go/bin/moai" ]; then
    MOAI_BIN="$HOME/go/bin/moai"
elif command -v moai >/dev/null 2>&1; then
    MOAI_BIN="moai"
else
    # moai binary not found — fail-open (exit 0, do not block).
    exit 0
fi

# Read stdin and pass to the Go handler.
INPUT=$(cat)
"$MOAI_BIN" hook chain-event <<< "$INPUT" 2>/dev/null || true

# Always exit 0 — the hook is advisory and non-blocking.
exit 0
