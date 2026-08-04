#!/usr/bin/env bash
# drain_test.sh — fixture-based characterization test for drain.sh (the hns-lsel-curator mechanical drain engine).
#
# SPEC-LSEL-LOCAL-EVOLUTION-001 M1, TDD RED→GREEN harness.
# Covers AC-LSEL-009 (drain produces candidates, offset advances, zero memory/ writes)
# and AC-LSEL-010 (drain-side severity filter excludes Bash-timeout/sandbox noise BEFORE clustering).
#
# Builds a synthetic lessons-inbox with known noise + signal stubs, runs drain.sh against
# a temp state-dir, and asserts the drain semantics via jq. No real inbox is touched.
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
DRAIN="$HERE/drain.sh"

fail() { echo "FAIL: $*" >&2; exit 1; }
ok()   { echo "ok - $*"; }

command -v jq >/dev/null || fail "jq is required for drain_test.sh"

if [[ ! -x "$DRAIN" ]]; then
  # drain.sh does not exist yet (RED), or is not executable. Surface a clear failure.
  fail "drain.sh not found or not executable at $DRAIN (expected during RED phase; this line is the RED failure)"
fi

# ---------------------------------------------------------------------------
# Fixture: 12 stubs spanning noise + signal classes.
#   noise (must be discarded by severity filter):
#     - 5x tool_failure:Bash:UnknownFailure   (the dominant ~65% timeout/sandbox bucket)
#     - 1x tool_failure:Bash:TimeoutError
#     - 1x tool_failure:mcp__web_search_prime__w:TimeoutError
#   signal (must survive the filter and be clustered):
#     - 3x tool_failure:Bash:ExitError         (Bash signal cluster, freq=3)
#     - 4x tool_failure:Agent:UnknownFailure   (non-Bash signal cluster, freq=4)
#     - 3x tool_failure:Write:PermissionDenied (non-Bash signal cluster, freq=3)
#   singleton signal (must be discarded by the freq>=2 gate):
#     - 1x tool_failure:DesignSync:UnknownFailure (freq=1)
#
# Expected post-drain candidate set: 3 clusters {Bash:ExitError, Agent:UnknownFailure, Write:PermissionDenied}.
#   Bash:* share of accepted topics = 1/3 = 33%... wait — the <30% rule is an integration
#   property over the real inbox; this fixture intentionally keeps the test deterministic
#   rather than chasing the integration ratio. The noise-filter assertion (noise event_keys
#   absent from candidates) is the load-bearing AC-LSEL-010 check here.
# ---------------------------------------------------------------------------
FIX_TMP="$(mktemp -d)"
trap 'rm -rf "$FIX_TMP"' EXIT

INBOX="$FIX_TMP/lessons-inbox.jsonl"
STATE_DIR="$FIX_TMP/lsel-state"
mkdir -p "$STATE_DIR"

emit() { printf '%s\n' "$1" >>"$INBOX"; }

# noise — Bash:UnknownFailure x5
for _ in 1 2 3 4 5; do
  emit '{"timestamp":"2026-08-04T01:00:00Z","event_key":"tool_failure:Bash:UnknownFailure","summary":"cmd timed out / sandbox opaque","source":"tool:Bash"}'
done
# noise — Bash:TimeoutError x1
emit '{"timestamp":"2026-08-04T02:00:00Z","event_key":"tool_failure:Bash:TimeoutError","summary":"bash 120s ceiling","source":"tool:Bash"}'
# noise — MCP timeout x1
emit '{"timestamp":"2026-08-04T02:05:00Z","event_key":"tool_failure:mcp__web_search_prime__w:TimeoutError","summary":"mcp no response 328s","source":"tool:mcp__web_search_prime__w"}'
# signal — Bash:ExitError x3
for _ in 1 2 3; do
  emit '{"timestamp":"2026-08-04T03:00:00Z","event_key":"tool_failure:Bash:ExitError","summary":"exit code 1 — real failure signal","source":"tool:Bash"}'
done
# signal — Agent:UnknownFailure x4
for _ in 1 2 3 4; do
  emit '{"timestamp":"2026-08-04T04:00:00Z","event_key":"tool_failure:Agent:UnknownFailure","summary":"agent opaque failure","source":"tool:Agent"}'
done
# signal — Write:PermissionDenied x3
for _ in 1 2 3; do
  emit '{"timestamp":"2026-08-04T05:00:00Z","event_key":"tool_failure:Write:PermissionDenied","summary":"write blocked","source":"tool:Write"}'
done
# singleton signal — DesignSync freq=1 (discarded)
emit '{"timestamp":"2026-08-04T06:00:00Z","event_key":"tool_failure:DesignSync:UnknownFailure","summary":"singleton","source":"tool:DesignSync"}'

TOTAL_STUBS=18  # 5 Bash:UnknownFailure + 1 Bash:TimeoutError + 1 MCP timeout + 3 Bash:ExitError + 4 Agent + 3 Write + 1 DesignSync singleton

# ---------------------------------------------------------------------------
# Pre-flight: drain-offset.json absent → drain must start at offset 0.
# Confirm the fixture inbox line count matches expectation.
# ---------------------------------------------------------------------------
[[ "$(wc -l <"$INBOX" | tr -d ' ')" == "$TOTAL_STUBS" ]] || fail "fixture inbox line count mismatch"

# Guard: a stale memory/ dir under STATE_DIR would mask the no-memory-write assertion.
# (drain.sh never touches memory/ — the M1 invariant — but verify by asserting the
# dir is absent after drain rather than asserting it was never created, since STATE_DIR
# is a fresh tmp.)

# ---------------------------------------------------------------------------
# Run the drain.
# ---------------------------------------------------------------------------
"$DRAIN" --inbox "$INBOX" --state-dir "$STATE_DIR" >/tmp/drain-stdout.txt 2>/tmp/drain-stderr.txt
DRAIN_RC=$?
[[ $DRAIN_RC -eq 0 ]] || { echo "--- drain stdout ---"; cat /tmp/drain-stdout.txt; echo "--- drain stderr ---"; cat /tmp/drain-stderr.txt; fail "drain.sh exited non-zero ($DRAIN_RC)"; }

CLUSTERS="$STATE_DIR/clusters.json"
OFFSET="$STATE_DIR/drain-offset.json"
[[ -f "$CLUSTERS" ]] || fail "clusters.json not produced at $CLUSTERS"
[[ -f "$OFFSET"  ]] || fail "drain-offset.json not produced at $OFFSET"

# ---------------------------------------------------------------------------
# Assertion 1 — AC-LSEL-009: offset advanced by the number of stubs read.
# ---------------------------------------------------------------------------
OFFSET_AFTER="$(jq -r '.offset' "$OFFSET")"
[[ "$OFFSET_AFTER" == "$TOTAL_STUBS" ]] \
  || fail "offset did not advance: expected $TOTAL_STUBS, got $OFFSET_AFTER"
ok "offset advanced to $OFFSET_AFTER (read all $TOTAL_STUBS stubs)"

# ---------------------------------------------------------------------------
# Assertion 2 — AC-LSEL-010: noise event_keys absent from accepted candidates.
# ---------------------------------------------------------------------------
NOISE_KEYS=("tool_failure:Bash:UnknownFailure" "tool_failure:Bash:TimeoutError" "tool_failure:mcp__web_search_prime__w:TimeoutError")
for nk in "${NOISE_KEYS[@]}"; do
  cnt="$(jq --arg k "$nk" '[.candidates[] | select(.event_key == $k)] | length' "$CLUSTERS")"
  [[ "$cnt" == "0" ]] || fail "noise key $nk leaked into candidates ($cnt occurrence(s))"
done
ok "all three noise event_keys excluded from candidates (severity filter fired pre-cluster)"

# ---------------------------------------------------------------------------
# Assertion 3 — AC-LSEL-009: exactly the 3 signal clusters survived (singletons discarded).
# ---------------------------------------------------------------------------
CAND_COUNT="$(jq '.candidates | length' "$CLUSTERS")"
[[ "$CAND_COUNT" == "3" ]] \
  || fail "expected 3 candidate clusters, got $CAND_COUNT (singleton gate or clustering wrong)"
ok "3 candidate clusters emitted (singleton discarded)"

# ---------------------------------------------------------------------------
# Assertion 4 — candidate frequencies are correct (freq proxy for the 1-10 importance gate).
# ---------------------------------------------------------------------------
check_freq() {
  local key="$1" want="$2"
  local got; got="$(jq -r --arg k "$key" '[.candidates[] | select(.event_key == $k) | .frequency] | .[0]' "$CLUSTERS")"
  [[ "$got" == "$want" ]] || fail "frequency for $key: expected $want, got $got"
}
check_freq "tool_failure:Bash:ExitError"         "3"
check_freq "tool_failure:Agent:UnknownFailure"   "4"
check_freq "tool_failure:Write:PermissionDenied" "3"
ok "candidate frequencies correct (3, 4, 3)"

# ---------------------------------------------------------------------------
# Assertion 5 — each candidate carries a Generative-Agents-style 1-10 importance score.
# ---------------------------------------------------------------------------
while IFS=$'\t' read -r key importance; do
  [[ "$importance" =~ ^[1-9]$ || "$importance" == "10" ]] \
    || fail "importance for $key out of [1,10] range: $importance"
  [[ -n "$importance" ]] || fail "importance missing for $key"
done < <(jq -r '.candidates[] | [.event_key, (.importance|tostring)] | @tsv' "$CLUSTERS")
ok "every candidate carries a 1-10 importance score"

# ---------------------------------------------------------------------------
# Assertion 6 — AC-LSEL-009: zero memory/ writes (M1 produces candidate topics only).
# ---------------------------------------------------------------------------
if [[ -d "$STATE_DIR/memory" ]] || find "$STATE_DIR" -name 'feedback_*' -print -quit | grep -q .; then
  fail "drain wrote to memory/ or emitted feedback_* files (M1 invariant violated)"
fi
ok "no memory/ writes (M1 produces candidate topics in clusters.json only)"

# ---------------------------------------------------------------------------
# Assertion 7 — idempotency: a second drain with no new stubs is a no-op.
# ---------------------------------------------------------------------------
"$DRAIN" --inbox "$INBOX" --state-dir "$STATE_DIR" >/dev/null 2>&1 || fail "second drain (no new stubs) exited non-zero"
OFFSET_AFTER_2="$(jq -r '.offset' "$OFFSET")"
[[ "$OFFSET_AFTER_2" == "$TOTAL_STUBS" ]] \
  || fail "second drain moved offset on an empty delta: $OFFSET_AFTER_2"
ok "idempotent re-drain is a no-op (offset stable, empty-delta handled)"

echo
echo "ALL DRAIN TESTS PASSED"
