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

# t459: post-rotation silence. The collector-side cap (SPEC-INBOX-DRAIN-GAP-001)
# rotates the live inbox into lessons-inbox.jsonl.1 whenever the LSEL marker
# (.moai/state/lsel/) is absent. Rotation collapses the LIVE line count, so the
# backlog advisory goes quiet at exactly the moment stubs left the drain's reach
# — drain.sh reads only --inbox and carries zero references to any .N archive
# (measured: .moai/reports/t459/r2-drain-reach.log). Archived lines are undrained
# by the drain's own accounting: rotation implies the marker was absent, and the
# offset file lives inside that marker directory, so it read 0.
ROT_INBOX="$FIXTURE/rotated.jsonl"
ROT_STATE="$FIXTURE/rot-state"   # deliberately NOT created — marker absent
printf '{"event_key":"live"}\n' > "$ROT_INBOX"
printf '{"event_key":"archived"}\n%.0s' {1..40} > "$ROT_INBOX.1"

OUT=$(bash "$SCRIPT" --inbox "$ROT_INBOX" --state-dir "$ROT_STATE" --threshold 25 2>&1 || true)
if echo "$OUT" | grep -q "40 archived"; then
  pass "t459: rotation-archived stubs reported (advisory does not go silent)"
else
  fail "t459: advisory silent after rotation — 40 archived stubs unreported; output: $OUT"
fi

# t1249: stdout is SessionStart hook context — every notice must stay off it.
# Streams are captured SEPARATELY; merging them (2>&1) cannot see a stdout leak.
bash "$SCRIPT" --inbox "$FIXTURE/absent.jsonl" --state-dir "$STATE" --threshold 25 \
  >"$FIXTURE/abs.out" 2>"$FIXTURE/abs.err" || true
if [[ ! -s "$FIXTURE/abs.out" ]] && grep -q "inbox absent" "$FIXTURE/abs.err"; then
  pass "t1249: inbox-absent notice on stderr, stdout empty"
else
  fail "t1249: inbox-absent leaked to stdout; stdout: $(cat "$FIXTURE/abs.out") stderr: $(cat "$FIXTURE/abs.err")"
fi
printf '{"event_key":"x"}\n%.0s' {1..40} > "$INBOX"
echo '{"offset":0}' > "$STATE/drain-offset.json"
bash "$SCRIPT" --inbox "$INBOX" --state-dir "$STATE" --threshold 25 \
  >"$FIXTURE/ovf.out" 2>"$FIXTURE/ovf.err" || true
if [[ ! -s "$FIXTURE/ovf.out" ]] && grep -q "lsel-backlog: 40 unread stubs" "$FIXTURE/ovf.err"; then
  pass "t1249: overflow reminder on stderr, stdout empty"
else
  fail "t1249: overflow reminder leaked to stdout; stdout: $(cat "$FIXTURE/ovf.out")"
fi

if [[ "$FAIL" -ne 0 ]]; then exit 1; fi
log "backlog_check_test: PASS"
exit 0
