#!/bin/bash

# MoAI Hook Wrapper - handle-codex-review-gate.sh
# Forwards stdin JSON to `moai hook codex-review-gate` (the codex review-gate
# Stop-hook handler). Stop hooks COMPOSE — this wrapper is a SEPARATE entry
# alongside handle-stop.sh, handle-stop-goal.sh, and sync-phase-quality-gate.sh;
# it does NOT replace them. Each wrapper reads the same stdin JSON independently.
#
# The handler is opt-in (workflow.codex.review_gate.enabled, default OFF) and
# self-gates to ALLOW on a no-edit / loop-prevention / disabled turn; it runs a
# codex review otherwise and BLOCKs on a fail verdict. Fail-open ALLOW on a
# missing or inconclusive codex.
#
# Capture stdin once (Stop hooks may have multiple composited readers).
INPUT=$(cat)

# Resolve the moai binary (3-tier: $CLAUDE_PROJECT_DIR-relative, PATH, $HOME).
MOAI_BIN=""
if [ -n "$CLAUDE_PROJECT_DIR" ] && [ -x "$CLAUDE_PROJECT_DIR/../../moai" ]; then
	# Repo-relative build (dev): internal/ is two levels above .claude/hooks/moai/.
	MOAI_BIN="$CLAUDE_PROJECT_DIR/../../moai"
fi
if [ -z "$MOAI_BIN" ]; then
	if command -v moai >/dev/null 2>&1; then
		MOAI_BIN="$(command -v moai)"
	elif [ -x "$HOME/go/bin/moai" ]; then
		MOAI_BIN="$HOME/go/bin/moai"
	fi
fi

# Fail-open: if the moai binary is unavailable, ALLOW (never break the Stop
# pipeline). The handler itself is fail-open, so an absent binary is a no-op.
if [ -z "$MOAI_BIN" ]; then
	exit 0
fi

printf '%s' "$INPUT" | "$MOAI_BIN" hook codex-review-gate
# Always exit 0: the BLOCK decision rides the JSON Decision field on stdout
# (exit 2 would discard stdout JSON per Claude Code semantics).
exit 0
