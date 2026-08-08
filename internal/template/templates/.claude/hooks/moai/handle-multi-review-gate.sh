#!/bin/bash

# MoAI Hook Wrapper - handle-multi-review-gate.sh
# Forwards stdin JSON to `moai hook multi-review-gate` (the multi-model
# review-gate Stop-hook handler). Stop hooks COMPOSE — this wrapper is a
# SEPARATE entry alongside handle-stop.sh, handle-stop-goal.sh,
# sync-phase-quality-gate.sh, and handle-codex-review-gate.sh; it does NOT
# replace them. Each wrapper reads the same stdin JSON independently.
#
# The handler is opt-in (workflow.multi_review_gate.enabled, default OFF) and
# self-gates to ALLOW on a no-edit / loop-prevention / disabled turn. Otherwise
# it reads the most recent multi-model convergence result and BLOCKs only on an
# unresolved required-backend FAIL. Disagreement among advisory-only backends
# NEVER blocks. Fail-open ALLOW on a missing or malformed convergence result: a
# missing optional backend is evidence-of-absence, not evidence-of-failure.
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

printf '%s' "$INPUT" | "$MOAI_BIN" hook multi-review-gate
# Always exit 0: the BLOCK decision rides the JSON Decision field on stdout
# (exit 2 would discard stdout JSON per Claude Code semantics).
exit 0
