#!/bin/bash
# Hook: team-ac-verify
# Purpose: Team-mode AC verification (TaskCompleted event); dormant by default unless team mode is enabled
# Trigger: TaskCompleted event when team.enabled: true in workflow.yaml
#
# Dormant behavior: this hook exits 0 immediately unless workflow.yaml declares
# team.enabled: true. This avoids overhead in solo-mode sessions.
#
# Reject path (--reject stub): emits exit 2 with a ledger_note JSON field.
# This is the REQ-LEDGER-002 reject path (SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001).
# The trigger is a MINIMAL STUB (--reject test flag) — full AC-verification logic
# (parsing acceptance.md, running evidence commands, emitting exit 2 on AC failure)
# is explicitly OUT OF SCOPE and deferred to a follow-up SPEC. Exit-code semantics
# are unchanged: exit 2 still = reject completion (CON-1).
#
# Manual smoke test:
#   echo '{"task":{"metadata":{}}}' | bash .claude/hooks/moai/team-ac-verify.sh
# Expected: {"hook":"team-ac-verify","decision":"dormant",...} when team mode disabled.
#
# Reject-path smoke test:
#   bash .claude/hooks/moai/team-ac-verify.sh --reject
# Expected: exit 2 + JSON with "decision":"reject" and a "ledger_note" field.

set -e

# Opt-out flag
if [ "$1" = "--skip-hook" ]; then
    echo "{\"skipped\": true, \"reason\": \"--skip-hook flag\"}" >&2
    mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [team-ac-verify] skipped via --skip-hook" \
        >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log"
    exit 0
fi

# REQ-LEDGER-002 reject-path stub (SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001).
# Minimal trigger: explicit --reject test flag. The orchestrator injects the
# emitted ledger_note as the ledger-closing artifact for the rejected task.
# Full AC-verification logic (parse acceptance.md, run evidence commands) is
# deferred to a follow-up SPEC — see spec.md §X.1.
if [ "$1" = "--reject" ]; then
    LEDGER_NOTE="task rejected via --reject stub: AC verification not yet implemented (deferred to follow-up SPEC-V3R6-TEAM-AC-VERIFY-FULL-001)"
    cat <<EOF
{
  "hook": "team-ac-verify",
  "decision": "reject",
  "ledger_note": "$LEDGER_NOTE",
  "reason": "reject path triggered by --reject stub (REQ-LEDGER-002)"
}
EOF
    exit 2
fi

# Dormant capability gate (team-mode opt-in)
WORKFLOW_CONFIG="${CLAUDE_PROJECT_DIR:-$PWD}/.moai/config/sections/workflow.yaml"
if [ ! -f "$WORKFLOW_CONFIG" ]; then
    echo "{\"hook\":\"team-ac-verify\",\"decision\":\"dormant\",\"reason\":\"workflow.yaml absent\"}"
    exit 0
fi

# Detect team.enabled: true in workflow.yaml
# Simple YAML scan; tolerates inline comments and whitespace variation
TEAM_ENABLED=$(awk '
/^team:/ { in_team = 1; next }
in_team && /^[a-zA-Z]/ && !/^  / { in_team = 0 }
in_team && /^  enabled:/ {
    val = $2
    gsub(/[",]/, "", val)
    print val
    exit
}
' "$WORKFLOW_CONFIG")

if [ "$TEAM_ENABLED" != "true" ]; then
    echo "{\"hook\":\"team-ac-verify\",\"decision\":\"dormant\",\"reason\":\"team mode disabled (team.enabled != true)\"}"
    exit 0
fi

# Graceful degradation: jq required for active verification
if ! command -v jq >/dev/null 2>&1; then
    echo "{\"hook\":\"team-ac-verify\",\"decision\":\"allow\",\"warning\":\"jq absent — hook degraded to allow-all\"}"
    exit 0
fi

# Active team mode — verify AC reference in task metadata
INPUT=$(cat)
TASK_SUBJECT=$(echo "$INPUT" | jq -r '.task.subject // ""')
TASK_AC_REF=$(echo "$INPUT" | jq -r '.task.metadata.acceptance_criteria // .task.metadata.ac_ref // ""')

if [ -z "$TASK_AC_REF" ]; then
    echo "{\"hook\":\"team-ac-verify\",\"decision\":\"allow\",\"reason\":\"no AC reference in task.metadata; advisory only\",\"task_subject\":\"$TASK_SUBJECT\"}"
    exit 0
fi

# Try to verify AC: expect format like "SPEC-XXX-001#AC-FOO-003"
# This is a stub for future expansion; full AC verification would parse acceptance.md
# and run the AC's evidence command. For M4 baseline, log the reference and allow.
mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [team-ac-verify] task=\"$TASK_SUBJECT\" ac_ref=\"$TASK_AC_REF\"" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/team-ac-verify.log"

cat <<EOF
{
  "hook": "team-ac-verify",
  "decision": "allow",
  "task_subject": "$TASK_SUBJECT",
  "ac_ref": "$TASK_AC_REF",
  "note": "AC reference recorded for audit; active verification logic deferred to follow-up SPEC"
}
EOF
exit 0
