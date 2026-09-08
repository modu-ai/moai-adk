#!/usr/bin/env bash
# backlog_check.sh — AC-LSEL-007 SessionStart backlog-check (SPEC-LSEL-LOCAL-EVOLUTION-001 M2).
#
# Emits a system-reminder referencing the LSEL drain when the lessons-inbox
# line count exceeds (drain-offset.json offset + N). Designed to run as an
# advisory check on SessionStart, wired alongside the session_drain.sh wrapper
# (the durable mechanical trigger — SPEC-LSEL-DRAIN-STALL-001). Advisory only —
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

# t459: count rotated archive generations. The collector-side cap
# (SPEC-INBOX-DRAIN-GAP-001) rotates the live inbox into lessons-inbox.jsonl.N
# whenever the LSEL marker (.moai/state/lsel/) is absent, and drain.sh reads
# only --inbox — it carries zero references to any .N archive, so archived
# stubs are permanently out of the drain's reach. Rotation also collapses
# LINES, which would otherwise make this advisory go silent at exactly the
# moment stubs left reach. Archived lines count as undrained by the drain's own
# accounting: a rotation implies the marker was absent, and the offset file
# lives inside that marker directory, so the offset read 0.
#
# The generation set is discovered by glob, never by restating the retention
# count — that constant lives in Go (config.DefaultInboxArchiveGenerations) and
# must not be duplicated here (CLAUDE.local.md sec 14).
ARCHIVED=0
for gen in "$INBOX".[0-9]*; do
  [[ -f "$gen" ]] || continue
  ARCHIVED=$(( ARCHIVED + $(wc -l < "$gen" | tr -d ' ') ))
done

if [[ "$BACKLOG" -le "$THRESHOLD" && "$ARCHIVED" -eq 0 ]]; then
  exit 0   # advisory silent — below threshold and nothing rotated out of reach
fi

# Emit a system-reminder to stderr (advisory; the orchestrator reads stderr reminders)
ARCHIVED_NOTE=""
if [[ "$ARCHIVED" -gt 0 ]]; then
  ARCHIVED_NOTE="
lsel-rotation: $ARCHIVED archived stubs in $INBOX.N are OUT OF THE DRAIN'S REACH — drain.sh reads only the live inbox and references no .N generation. A rotation fires only while the marker dir ($STATE_DIR) is absent, so restore it (session_drain.sh recreates it) before further appends; the archives are then frozen but bounded, and each further rotation evicts the oldest. Recovering them needs a SEPARATE --state-dir: the shared drain-offset.json is a line index into the live file and reusing it would skip live stubs."
fi

cat >&2 <<EOF
<system-reminder>
lsel-backlog: $BACKLOG unread stubs in $INBOX (offset=$OFFSET, threshold=$THRESHOLD).$ARCHIVED_NOTE
Run the LSEL drain via the wrapper: .claude/skills/hns-lsel-curator/session_drain.sh --inbox $INBOX --state-dir $STATE_DIR
Then draft shadow proposals from the archived candidates in $STATE_DIR/clusters-history/ (hns-lsel-curator PROPOSE; the live clusters.json is ephemeral under per-session-start drains).
</system-reminder>
EOF
exit 0
