#!/usr/bin/env bash
# backlog_check.sh — AC-LSEL-007 SessionStart backlog-check (SPEC-LSEL-LOCAL-EVOLUTION-001 M2).
#
# Emits a system-reminder referencing the LSEL drain when the lessons-inbox
# line count exceeds (drain-offset.json offset + N). Designed to run as an
# advisory check on SessionStart (see CLAUDE.local.md §28). Advisory only —
# never blocks session start (advisory-check discipline, coding-standards.md).
#
# Usage: ./backlog_check.sh [--inbox <path>] [--state-dir <path>] [--threshold N]
set -euo pipefail

INBOX=".moai/lessons-inbox.jsonl"
STATE_DIR=".moai/state/lsel"
THRESHOLD="${LSEL_BACKLOG_THRESHOLD:-25}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --inbox) INBOX="$2"; shift 2 ;;
    --state-dir) STATE_DIR="$2"; shift 2 ;;
    --threshold) THRESHOLD="$2"; shift 2 ;;
    *) shift ;;
  esac
done

OFFSET_FILE="$STATE_DIR/drain-offset.json"
OFFSET=0
if [[ -f "$OFFSET_FILE" ]]; then
  OFFSET=$(grep -o '"offset":[[:space:]]*[0-9]*' "$OFFSET_FILE" | head -1 | grep -o '[0-9]*' || echo 0)
  OFFSET="${OFFSET:-0}"
fi

if [[ ! -f "$INBOX" ]]; then
  echo "lsel-backlog: inbox absent; no-op"
  exit 0
fi

LINES=$(wc -l < "$INBOX" | tr -d ' ')
BACKLOG=$(( LINES - OFFSET ))

if [[ "$BACKLOG" -le "$THRESHOLD" ]]; then
  exit 0   # advisory silent — below threshold
fi

# Emit a system-reminder to stderr (advisory; the orchestrator reads stderr reminders)
cat >&2 <<EOF
<system-reminder>
lsel-backlog: $BACKLOG unread stubs in $INBOX (offset=$OFFSET, threshold=$THRESHOLD).
Run the LSEL drain: .claude/skills/hns-lsel-curator/drain.sh --inbox $INBOX --state-dir $STATE_DIR
Then draft shadow proposals from .moai/state/lsel/clusters.json (hns-lsel-curator PROPOSE).
Per CLAUDE.local.md §28 (LSEL operating instructions).
</system-reminder>
EOF
exit 0
