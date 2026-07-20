#!/bin/bash
# Hook: team-ac-verify
# Purpose: Team-mode AC verification (TaskCompleted event); dormant by default unless team mode is enabled
# Trigger: TaskCompleted event when team.enabled: true in workflow.yaml
#
# Dormant behavior: this hook exits 0 immediately unless workflow.yaml declares
# team.enabled: true. This avoids overhead in solo-mode sessions.
#
# Delivery channel: a completion rejection is signaled via stdout JSON
# {"continue":false,"stopReason":"AC verification failed: ...","ledger_note":"..."}
# + exit 0. Per Claude Code hook semantics, stdout JSON is honored only on exit
# 0 — on exit 2 stdout is discarded and only stderr is surfaced — so the reject
# decision MUST ride the exit-0 stdout channel. The "decision" field is NOT
# used here because it is documented only for PostToolUse/Stop/SubagentStop/
# UserPromptSubmit/ConfigChange/PreCompact/PostToolBatch — NOT TaskCompleted.
# The official TaskCompleted reject contract is {"continue":false,"stopReason":...}.
#
# Reject path (--reject stub): emits the continue:false + stopReason form above
# with a ledger_note sidecar field; the orchestrator injects the ledger_note as
# the ledger-closing artifact for the rejected task (see agent-common-protocol.md
# § Ledger Closure). The trigger is a MINIMAL STUB (--reject test flag) — full
# AC-verification logic (parsing acceptance.md, running evidence commands,
# blocking on AC failure) is deferred to a follow-up.
#
# Manual smoke test:
#   echo '{"task":{"metadata":{}}}' | bash .claude/hooks/moai/team-ac-verify.sh
# Expected: {"hook":"team-ac-verify","status":"dormant",...} when team mode disabled.
#
# Reject-path smoke test:
#   bash .claude/hooks/moai/team-ac-verify.sh --reject
# Expected: exit 0 + JSON with "continue":false, "stopReason", and a "ledger_note" field.

set -e

# Opt-out flag
if [ "$1" = "--skip-hook" ]; then
    echo "{\"skipped\": true, \"reason\": \"--skip-hook flag\"}" >&2
    mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [team-ac-verify] skipped via --skip-hook" \
        >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log"
    exit 0
fi

# Reject-path stub: minimal trigger via explicit --reject test flag (static
# JSON — no interpolation, safe without jq).
if [ "$1" = "--reject" ]; then
    printf '{"continue":false,"stopReason":"AC verification failed: task completion rejected via --reject stub (full AC verification deferred to a follow-up)","ledger_note":"task rejected via --reject stub: AC verification not yet implemented (full AC verification deferred to a follow-up)"}\n'
    exit 0
fi

# Dormant capability gate (team-mode opt-in)
WORKFLOW_CONFIG="${CLAUDE_PROJECT_DIR:-$PWD}/.moai/config/sections/workflow.yaml"
if [ ! -f "$WORKFLOW_CONFIG" ]; then
    echo "{\"hook\":\"team-ac-verify\",\"status\":\"dormant\",\"reason\":\"workflow.yaml absent\"}"
    exit 0
fi

# Detect workflow.team.enabled: true in workflow.yaml.
# The shipped schema nests team: under the top-level workflow: block
# (workflow: > team: > enabled:, 4-space indent), so team: is NOT at column 0.
# Indentation-aware scan: find the team: mapping at any indent, then match its
# DIRECT child enabled: (one 4-space level deeper) — this avoids false matches
# on other enabled: keys at the same depth (e.g. auto_clear.enabled).
TEAM_ENABLED=$(awk '
/^[[:space:]]*#/ { next }
/^[[:space:]]*team:[[:space:]]*$/ { in_team = 1; team_indent = match($0, /[^[:space:]]/); next }
in_team && NF > 0 {
    indent = match($0, /[^[:space:]]/)
    if (indent <= team_indent) { in_team = 0 }
}
in_team && $1 == "enabled:" && match($0, /[^[:space:]]/) == team_indent + 4 {
    val = $2
    gsub(/[",]/, "", val)
    print val
    exit
}
' "$WORKFLOW_CONFIG")

if [ "$TEAM_ENABLED" != "true" ]; then
    echo "{\"hook\":\"team-ac-verify\",\"status\":\"dormant\",\"reason\":\"team mode disabled (team.enabled != true)\"}"
    exit 0
fi

# Graceful degradation: jq required for active verification
if ! command -v jq >/dev/null 2>&1; then
    echo "{\"hook\":\"team-ac-verify\",\"status\":\"allow\",\"warning\":\"jq absent — hook degraded to allow-all\"}"
    exit 0
fi

# Active team mode — verify AC reference in task metadata
INPUT=$(cat)
TASK_SUBJECT=$(echo "$INPUT" | jq -r '.task.subject // ""')
TASK_AC_REF=$(echo "$INPUT" | jq -r '.task.metadata.acceptance_criteria // .task.metadata.ac_ref // ""')

if [ -z "$TASK_AC_REF" ]; then
    # jq -n builds the JSON so an embedded quote in the task subject cannot
    # produce malformed output. No "decision" key: informational allow only.
    jq -cn --arg subject "$TASK_SUBJECT" \
        '{hook: "team-ac-verify", status: "allow", reason: "no AC reference in task.metadata; advisory only", task_subject: $subject}'
    exit 0
fi

# Log the AC reference for audit; active verification logic (parse
# acceptance.md, run the AC's evidence command) is deferred to a follow-up.
mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [team-ac-verify] task=\"$TASK_SUBJECT\" ac_ref=\"$TASK_AC_REF\"" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/team-ac-verify.log"

jq -cn --arg subject "$TASK_SUBJECT" --arg ac_ref "$TASK_AC_REF" \
    '{hook: "team-ac-verify", status: "allow", task_subject: $subject, ac_ref: $ac_ref, note: "AC reference recorded for audit; active verification logic deferred to a follow-up"}'
exit 0
