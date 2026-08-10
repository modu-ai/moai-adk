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

# --- shell-layer self-gate (the handle-stop-goal.sh precondition pattern) ---
# The gate ships OFF, so being registered in the Stop array must NOT add a moai
# cold start to every turn-end for every user: a hook that ships OFF should
# cost nothing per turn. Read the opt-in here, in pure shell, and exit 0 before
# any binary resolution unless it is explicitly true.
#
# The parse is deliberately conservative: awk walks the nested
# workflow: -> codex: -> review_gate: -> enabled: indentation and accepts only a
# literal true (the values the Go yaml decoder accepts for a bool). Anything it
# cannot read — an unusual layout, a flow mapping, an absent awk — counts as
# OFF. The failure direction is the point: being fooled into OFF costs one
# maintainer an opt-in to debug, while being fooled into ON costs every user a
# cold start and a possible block. Same fail-CLOSED direction as the Go reader.
# No jq/yq dependency (per hook independence: a precondition MUST NOT introduce
# a new shared failure mode).
CODEX_GATE_PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
CODEX_GATE_WORKFLOW_YAML="$CODEX_GATE_PROJECT_ROOT/.moai/config/sections/workflow.yaml"

[ -f "$CODEX_GATE_WORKFLOW_YAML" ] || exit 0
command -v awk >/dev/null 2>&1 || exit 0

CODEX_GATE_ENABLED=$(awk -v gate="codex" '
{
    line = $0
    sub(/#.*/, "", line)
    if (line ~ /^[[:space:]]*$/) next
    match(line, /^ */); ind = RLENGTH
    key = substr(line, ind + 1)
    sub(/[[:space:]]+$/, "", key)

    # Leaving a block: a line at or left of its indent closes it.
    if (in_r && ind <= r_ind) in_r = 0
    if (in_g && ind <= g_ind) { in_g = 0; in_r = 0 }
    if (in_w && ind <= w_ind) { in_w = 0; in_g = 0; in_r = 0 }

    if (!in_w) { if (ind == 0 && key == "workflow:") { in_w = 1; w_ind = ind } next }
    if (!in_g) { if (key == gate ":") { in_g = 1; g_ind = ind } next }
    if (!in_r) { if (key == "review_gate:") { in_r = 1; r_ind = ind } next }
    if (key ~ /^enabled:[[:space:]]*(true|True|TRUE)$/) { print "true"; exit }
}
' "$CODEX_GATE_WORKFLOW_YAML" 2>/dev/null)

# Gate off (or unreadable) ⇒ ALLOW with zero moai cold-start.
[ "$CODEX_GATE_ENABLED" = "true" ] || exit 0

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
