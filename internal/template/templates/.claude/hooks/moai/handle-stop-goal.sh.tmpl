#!/bin/bash

# MoAI Hook Wrapper - handle-stop-goal.sh
# Forwards stdin JSON to `moai hook stop-goal` (the goal-engine Stop-hook
# evaluator). Stop hooks COMPOSE — this wrapper is registered as a SEPARATE
# entry alongside handle-stop.sh and sync-phase-quality-gate.sh; it does NOT
# replace them. Both wrappers read the same stdin JSON independently.
#
# SPEC-STOPCHAIN-TRIM-001 (A10 / AC-001): shell-layer precondition. A session
# with NO armed goal pays ZERO moai-binary cold-starts at turn-end. Before any
# binary resolution, the wrapper (1) captures stdin once, (2) extracts
# session_id from the Stop hook JSON at the shell layer (no jq — the hook may
# run where jq is absent; session_id is a flat top-level field), (3) checks
# whether .moai/state/goal/<session-id>.json exists, and (4) `exit 0`s
# immediately when it does not. Only when a goal IS armed does the wrapper
# proceed to the 3-tier moai-binary resolution and re-feed the captured stdin.
# This collapses the per-turn cold-start tax from "always" to "conditional on
# an armed goal" for goal-less sessions.
#
# K-3 resolution: the Stop hook stdin DOES carry session_id (HookInput.SessionID
# json:"session_id,omitempty" — internal/hook/types.go), so the shell-layer
# path is constructible without a runtime contract change.

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

# --- SPEC-STOPCHAIN-TRIM-001 A10: capture stdin + shell precondition ---
# Read the Stop hook JSON ONCE so it can be re-fed to the moai binary later
# (exec inherits stdin; once we consume it here, the binary would see EOF
# unless we replay it).
INPUT=$(cat)

# Extract session_id at the shell layer. Best-effort flat-field extraction
# (session_id is a top-level field in the Stop hook payload; no nested-object
# traversal needed). grep+sed — no jq dependency (per hook-independence
# doctrine: a precondition MUST NOT introduce a new shared failure mode).
SESSION_ID=$(printf '%s' "$INPUT" | grep -o '"session_id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -n 1 | sed 's/.*"session_id"[[:space:]]*:[[:space:]]*"//; s/"[[:space:]]*$//')

PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
STATE_FILE="$PROJECT_ROOT/.moai/state/goal/${SESSION_ID}.json"

# Goal-less session: short-circuit at the shell layer. No moai binary exec,
# no cold-start tax. (AC-001)
if [ -z "$SESSION_ID" ] || [ ! -f "$STATE_FILE" ]; then
    exit 0
fi

# A goal IS armed — proceed to the 3-tier moai-binary resolution, replaying
# the captured stdin so the evaluator sees the full Stop payload.
# Try moai command in PATH
if command -v moai &> /dev/null; then
    printf '%s' "$INPUT" | exec moai hook stop-goal 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try default ~/go/bin/moai
if [ -f "$HOME/go/bin/moai" ]; then
    printf '%s' "$INPUT" | exec "$HOME/go/bin/moai" hook stop-goal 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try ~/.local/bin/moai (Linux install location)
if [ -f "$HOME/.local/bin/moai" ]; then
    printf '%s' "$INPUT" | exec "$HOME/.local/bin/moai" hook stop-goal 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Not found - exit silently (Claude Code handles missing hooks gracefully)
exit 0
