#!/usr/bin/env bash
# session_drain_test.sh — fixture-based test for the session-start drain wrapper
# `session_drain.sh` (SPEC-LSEL-DRAIN-STALL-001 M1; AC-LDS-001..006 and the
# AC-LDS-005 / AC-LDS-010 mutant probe).
#
# Paths covered:
#   1. normal drain             (AC-LDS-001) — offset advance + candidates + one-line status
#   2. lock contention          (AC-LDS-002) — skip + exit 0 + contention notice, drain NOT run
#   3. archive-before-overwrite (AC-LDS-003) — prior staged candidates preserved to clusters-history
#   4. no-op                    (AC-LDS-004) — offset==tail: exit 0, offset unchanged, no-op status
#   5. fail-open                (AC-LDS-006) — internal failure still exits 0 with a stderr notice
#   6. mutant probe             (AC-LDS-005) — an offset-only-advance fake FAILS the verification
#      predicate (a check whose red has never been seen proves nothing).
#
# Fixture discipline (drain_test.sh pattern): everything under mktemp -d with trap
# cleanup; the REAL .moai/state/lsel and the live inbox are never touched (M1
# boundary — live wiring and the bulk backlog drain are M2).
set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
WRAPPER="$HERE/session_drain.sh"

fail() { echo "FAIL: $*" >&2; exit 1; }
ok()   { echo "ok - $*"; }

command -v jq >/dev/null || fail "jq is required for session_drain_test.sh"

if [[ ! -x "$WRAPPER" ]]; then
  # session_drain.sh does not exist yet (RED), or is not executable. Surface a clear failure.
  fail "session_drain.sh not found or not executable at $WRAPPER (expected during RED phase; this line is the RED failure)"
fi

FIX_TMP="$(mktemp -d)"
trap 'rm -rf "$FIX_TMP"' EXIT

# ---------------------------------------------------------------------------
# Fixture inbox: 18 stubs (same composition as drain_test.sh) —
#   noise 7     (5x Bash:UnknownFailure, 1x Bash:TimeoutError, 1x MCP timeout)
#   signal 10   (3x Bash:ExitError, 4x Agent:UnknownFailure, 3x Write:PermissionDenied)
#   singleton 1 (DesignSync)
# Expected real-drain result: offset 0->18, exactly 3 candidate clusters.
# ---------------------------------------------------------------------------
build_inbox() {
  local f="$1" i
  : >"$f"
  for i in 1 2 3 4 5; do
    printf '%s\n' '{"timestamp":"2026-08-04T01:00:00Z","event_key":"tool_failure:Bash:UnknownFailure","summary":"noise","source":"tool:Bash"}' >>"$f"
  done
  printf '%s\n' '{"timestamp":"2026-08-04T02:00:00Z","event_key":"tool_failure:Bash:TimeoutError","summary":"noise","source":"tool:Bash"}' >>"$f"
  printf '%s\n' '{"timestamp":"2026-08-04T02:05:00Z","event_key":"tool_failure:mcp__web_search_prime__w:TimeoutError","summary":"noise","source":"tool:mcp__web_search_prime__w"}' >>"$f"
  for i in 1 2 3; do
    printf '%s\n' '{"timestamp":"2026-08-04T03:00:00Z","event_key":"tool_failure:Bash:ExitError","summary":"signal","source":"tool:Bash"}' >>"$f"
  done
  for i in 1 2 3 4; do
    printf '%s\n' '{"timestamp":"2026-08-04T04:00:00Z","event_key":"tool_failure:Agent:UnknownFailure","summary":"signal","source":"tool:Agent"}' >>"$f"
  done
  for i in 1 2 3; do
    printf '%s\n' '{"timestamp":"2026-08-04T05:00:00Z","event_key":"tool_failure:Write:PermissionDenied","summary":"signal","source":"tool:Write"}' >>"$f"
  done
  printf '%s\n' '{"timestamp":"2026-08-04T06:00:00Z","event_key":"tool_failure:DesignSync:UnknownFailure","summary":"singleton","source":"tool:DesignSync"}' >>"$f"
  [[ "$(wc -l <"$f" | tr -d ' ')" == "18" ]] || fail "fixture inbox line count mismatch (expected 18)"
}

# The AC-LDS-010 verification predicate — IDENTICAL to the hns-lsel-curator
# SKILL.md Verification recipe: offset==live AND candidates>=1 AND self-consistency
# (total_read == live - offset_before). Judged on the ARCHIVED clusters-history copy.
verify_predicate() { # $1=json file  $2=live  $3=offset_before -> prints true/false (bare boolean)
  jq --argjson live "$2" --argjson before "$3" \
     '(.offset_after == $live) and ((.candidates // []) | length >= 1) and (.total_read == ($live - $before))' \
     "$1"
}

run_wrapper() { # $1=inbox $2=state-dir $3=stdout-file $4=stderr-file -> sets RUN_RC
  RUN_RC=0
  "$WRAPPER" --inbox "$1" --state-dir "$2" >"$3" 2>"$4" || RUN_RC=$?
}

echo "=== SPEC-LSEL-DRAIN-STALL-001 M1: session_drain wrapper fixture ==="

# --- path 1: normal drain (AC-LDS-001) ---------------------------------------
INBOX1="$FIX_TMP/inbox1.jsonl"; build_inbox "$INBOX1"
SD1="$FIX_TMP/state1"
run_wrapper "$INBOX1" "$SD1" "$FIX_TMP/p1.out" "$FIX_TMP/p1.err"
[[ "$RUN_RC" -eq 0 ]] || { cat "$FIX_TMP/p1.err"; fail "path1: wrapper exited $RUN_RC"; }
[[ "$(jq -r '.offset' "$SD1/drain-offset.json")" == "18" ]] \
  || fail "path1: offset expected 18, got $(jq -r '.offset' "$SD1/drain-offset.json")"
[[ "$(jq '.candidates | length' "$SD1/clusters.json")" == "3" ]] \
  || fail "path1: expected 3 candidates, got $(jq '.candidates | length' "$SD1/clusters.json")"
grep -q '^session_drain: read=18 candidates=3 offset=18$' "$FIX_TMP/p1.err" \
  || fail "path1: one-line status missing; stderr: $(cat "$FIX_TMP/p1.err")"
[[ ! -d "$SD1/clusters-history" ]] || fail "path1: fresh first run must not create an archive"
ok "path1 normal drain: offset 0->18, 3 candidates, one-line status emitted"

# --- path 2: lock contention (AC-LDS-002) -------------------------------------
INBOX2="$FIX_TMP/inbox2.jsonl"; build_inbox "$INBOX2"
SD2="$FIX_TMP/state2"; mkdir -p "$SD2"
mkdir "$SD2/drain.lock"   # the TEST plays the concurrent session holding the lock
run_wrapper "$INBOX2" "$SD2" "$FIX_TMP/p2.out" "$FIX_TMP/p2.err"
[[ "$RUN_RC" -eq 0 ]] || { cat "$FIX_TMP/p2.err"; fail "path2: wrapper exited $RUN_RC"; }
grep -q 'contention' "$FIX_TMP/p2.err" || fail "path2: contention notice missing; stderr: $(cat "$FIX_TMP/p2.err")"
[[ ! -f "$SD2/clusters.json" ]]     || fail "path2: clusters.json written despite lock contention (drain ran!)"
[[ ! -f "$SD2/drain-offset.json" ]] || fail "path2: offset written despite lock contention (drain ran!)"
[[ ! -d "$SD2/clusters-history" ]]  || fail "path2: archive ran despite lock contention"
rmdir "$SD2/drain.lock"   # release the test-held lock
ok "path2 lock contention: skip + exit 0 + notice; drain/archive NOT executed"

# --- path 3: archive-before-overwrite (AC-LDS-003) ----------------------------
INBOX3="$FIX_TMP/inbox3.jsonl"; build_inbox "$INBOX3"
SD3="$FIX_TMP/state3"; mkdir -p "$SD3"
# prior staged state: clusters.json with ONE candidate (what a PROPOSE pass would consume)
jq -n '{drained_at:"2026-08-20T00:00:00Z",offset_before:0,offset_after:0,total_read:0,
        noise_discarded:0,singletons_discarded:0,
        candidates:[{event_key:"tool_failure:Fake:Prior",frequency:2,importance:2}]}' \
  >"$SD3/clusters.json"
run_wrapper "$INBOX3" "$SD3" "$FIX_TMP/p3.out" "$FIX_TMP/p3.err"
[[ "$RUN_RC" -eq 0 ]] || { cat "$FIX_TMP/p3.err"; fail "path3: wrapper exited $RUN_RC"; }
ARCHIVES3="$(ls "$SD3/clusters-history" 2>/dev/null | wc -l | tr -d ' ')"
[[ "$ARCHIVES3" == "1" ]] || fail "path3: expected exactly 1 archived copy, got $ARCHIVES3"
grep -q 'tool_failure:Fake:Prior' "$SD3/clusters-history/"*.json \
  || fail "path3: archived copy does not contain the prior staged candidate"
[[ "$(jq '.candidates | length' "$SD3/clusters.json")" == "3" ]] \
  || fail "path3: new clusters.json is not the fresh drain result"
jq -e --arg k 'tool_failure:Fake:Prior' '([.candidates[] | select(.event_key == $k)] | length) == 0' \
  "$SD3/clusters.json" >/dev/null || fail "path3: prior candidate leaked into the new drain result"
ok "path3 archive-before-overwrite: prior candidate preserved to clusters-history, then overwritten"

# --- path 4: no-op (AC-LDS-004) — continues on SD1 (offset 18 == tail 18) ------
run_wrapper "$INBOX1" "$SD1" "$FIX_TMP/p4.out" "$FIX_TMP/p4.err"
[[ "$RUN_RC" -eq 0 ]] || { cat "$FIX_TMP/p4.err"; fail "path4: wrapper exited $RUN_RC"; }
[[ "$(jq -r '.offset' "$SD1/drain-offset.json")" == "18" ]] \
  || fail "path4: offset moved on no-op: $(jq -r '.offset' "$SD1/drain-offset.json")"
grep -q '^session_drain: no-op read=0' "$FIX_TMP/p4.err" \
  || fail "path4: no-op status line missing; stderr: $(cat "$FIX_TMP/p4.err")"
# The no-op overwrite WIPES the live candidates (spec section B.5 ephememerality) —
# and the archive taken BEFORE the overwrite preserves the path-1 bulk result:
[[ "$(jq '.candidates | length' "$SD1/clusters.json")" == "0" ]] \
  || fail "path4: expected live clusters.json wiped to 0 candidates by the no-op overwrite"
ARCH4="$SD1/clusters-history/$(ls "$SD1/clusters-history" | head -1)"
[[ "$(jq '.candidates | length' "$ARCH4")" == "3" ]] \
  || fail "path4: archived copy does not hold the 3-candidate bulk result"
grep -q 'tool_failure:Agent:UnknownFailure' "$ARCH4" \
  || fail "path4: archived copy missing a real signal cluster"
ok "path4 no-op: exit 0, offset unchanged, no-op status; bulk result archived before the wipe"

# --- path 5: fail-open (AC-LDS-006) --------------------------------------------
# 5a. inbox absent (fresh clone shape)
run_wrapper "$FIX_TMP/does-not-exist.jsonl" "$FIX_TMP/state5a" "$FIX_TMP/p5a.out" "$FIX_TMP/p5a.err"
[[ "$RUN_RC" -eq 0 ]] || fail "path5a: inbox-absent must exit 0, got $RUN_RC"
grep -q '^session_drain:' "$FIX_TMP/p5a.err" || fail "path5a: stderr notice missing"
# 5b. state dir uncreatable (a FILE blocks the mkdir -p path)
: >"$FIX_TMP/blocker-file"
run_wrapper "$INBOX1" "$FIX_TMP/blocker-file/sub" "$FIX_TMP/p5b.out" "$FIX_TMP/p5b.err"
[[ "$RUN_RC" -eq 0 ]] || fail "path5b: uncreatable state dir must exit 0, got $RUN_RC"
grep -q '^session_drain:' "$FIX_TMP/p5b.err" || fail "path5b: stderr notice missing"
ok "path5 fail-open: inbox-absent and uncreatable state dir both exit 0 with a stderr notice"

# --- path 6: mutant probe (AC-LDS-005 / the AC-LDS-010 guard) ------------------
INBOX6="$FIX_TMP/inbox6.jsonl"; build_inbox "$INBOX6"
# REAL: bulk drain via the wrapper, then a second (no-op) run whose unconditional
# archive preserves the bulk result — the archived copy is what verification judges.
SD6R="$FIX_TMP/state6r"
run_wrapper "$INBOX6" "$SD6R" "$FIX_TMP/p6r1.out" "$FIX_TMP/p6r1.err"
run_wrapper "$INBOX6" "$SD6R" "$FIX_TMP/p6r2.out" "$FIX_TMP/p6r2.err"
ARCH_R="$SD6R/clusters-history/$(ls -t "$SD6R/clusters-history" | head -1)"
VERDICT_REAL="$(verify_predicate "$ARCH_R" 18 0)"
[[ "$VERDICT_REAL" == "true" ]] \
  || fail "mutant probe: predicate rejected the REAL bulk drain ($VERDICT_REAL) — predicate too strict"
# MUTANT INJECTION: what an offset-only-advance implementation would leave behind —
# offset pushed to live, clusters.json written WITHOUT real clustering (candidates: []).
SD6M="$FIX_TMP/state6m"; mkdir -p "$SD6M"
jq -n --argjson off 18 '{offset:$off, updated:"2099-01-01T00:00:00Z"}' >"$SD6M/drain-offset.json"
jq -n --argjson off 18 '{drained_at:"2099-01-01T00:00:00Z",offset_before:0,offset_after:$off,
                         total_read:$off,noise_discarded:0,singletons_discarded:0,candidates:[]}' \
  >"$SD6M/clusters.json"
run_wrapper "$INBOX6" "$SD6M" "$FIX_TMP/p6m.out" "$FIX_TMP/p6m.err"
ARCH_M="$SD6M/clusters-history/$(ls -t "$SD6M/clusters-history" | head -1)"
VERDICT_MUTANT="$(verify_predicate "$ARCH_M" 18 0)"
[[ "$VERDICT_MUTANT" == "false" ]] \
  || fail "mutant probe FAILED: predicate accepted the offset-only-advance mutant — the guard is hollow"
ok "mutant probe: predicate accepts the real bulk drain (true) and REJECTS the offset-only-advance mutant (false)"

echo
echo "ALL SESSION DRAIN TESTS PASSED"
