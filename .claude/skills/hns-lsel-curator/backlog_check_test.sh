#!/usr/bin/env bash
# backlog_check_test.sh — AC-LSEL-007 fixture test for the SessionStart backlog check
# (SPEC-LSEL-LOCAL-EVOLUTION-001 M2).
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$PROJECT_ROOT"

FIXTURE="$(mktemp -d)"
trap 'rm -rf "$FIXTURE"' EXIT
SCRIPT=".claude/skills/hns-lsel-curator/backlog_check.sh"

FAIL=0
log() { printf '%s\n' "$*"; }
fail() { log "FAIL: $*"; FAIL=1; }
pass() { log "PASS: $*"; }

log "=== AC-LSEL-007 SessionStart backlog-check fixture ==="

# Fixture: inbox with 40 stubs, offset at 0, threshold 25 → overflow (40 > 0+25)
INBOX="$FIXTURE/inbox.jsonl"
STATE="$FIXTURE/state"
mkdir -p "$STATE"
printf '{"event_key":"x"}\n%.0s' {1..40} > "$INBOX"
echo '{"offset":0}' > "$STATE/drain-offset.json"

OUT=$(bash "$SCRIPT" --inbox "$INBOX" --state-dir "$STATE" --threshold 25 2>&1 || true)
if echo "$OUT" | grep -q "lsel-backlog: 40 unread stubs"; then
  pass "system-reminder emitted on overflow (40 > 0+25)"
else
  fail "no system-reminder on overflow; output: $OUT"
fi

# Fixture: inbox with 10 stubs, offset at 0, threshold 25 → below threshold (silent)
printf '{"event_key":"x"}\n%.0s' {1..10} > "$INBOX"
OUT=$(bash "$SCRIPT" --inbox "$INBOX" --state-dir "$STATE" --threshold 25 2>&1 || true)
if [[ -z "$OUT" ]]; then
  pass "silent below threshold (advisory non-blocking, no reminder)"
else
  fail "non-silent below threshold; output: $OUT"
fi

# Fixture: offset advanced to 35, 40 stubs, threshold 25 → backlog=5 ≤ 25 (silent)
echo '{"offset":35}' > "$STATE/drain-offset.json"
OUT=$(bash "$SCRIPT" --inbox "$INBOX" --state-dir "$STATE" --threshold 25 2>&1 || true)
if [[ -z "$OUT" ]]; then
  pass "silent after drain (offset advanced; backlog below threshold)"
else
  fail "non-silent after drain; output: $OUT"
fi

if [[ "$FAIL" -ne 0 ]]; then exit 1; fi
log "backlog_check_test: PASS"
exit 0
