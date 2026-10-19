#!/usr/bin/env bash
# conditions.sh <N> — synthetic reproduction of drain failure condition N (1..18) on the LSEL drain pair
# (drain.sh + session_drain.sh). SPEC-LESSONS-INBOX-SAFE-CLEANUP-001 RED/GREEN cell driver.
#
# Conditions 1..9 are the nine ADK conditions (card t1901 body); 10 is this SPEC's own addition
# (a replaced inbox file is consumed as if it were the same lineage).
#
# Safety: every fixture lives in a fresh mktemp -d (honours TMPDIR) and is removed on exit. No real
# inbox or checkpoint path is ever an argument or a default; this script never writes outside $FIX.
#
# Output: key=value observation lines on stdout (paths sanitised). Exit code: 0 when the harness ran
# to completion (the observation, not the exit code, is the verdict), 64 usage, 65 curator scripts missing.
#
# CURATOR_DIR overrides the directory holding drain.sh / session_drain.sh (default: this tree's
# .claude/skills/hns-lsel-curator, resolved from this script's location).
set -uo pipefail

N="${1:-}"
case "$N" in 1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16|17|18) ;; *) echo "usage: conditions.sh <1..18> [crash-point [lost|tamper]] for 15" >&2; exit 64;; esac

HERE="$(cd "$(dirname "$0")" && pwd)"
CUR="${CURATOR_DIR:-$HERE/../../../../.claude/skills/hns-lsel-curator}"
[[ -x "$CUR/drain.sh" && -x "$CUR/session_drain.sh" ]] || { echo "curator scripts not found under $CUR" >&2; exit 65; }

REAL_JQ="$(command -v jq)"; REAL_WC="$(command -v wc)"
export REAL_JQ REAL_WC

FIX="$(mktemp -d)"
trap 'rm -rf "$FIX"' EXIT
INBOX="$FIX/inbox.jsonl"
SD="$FIX/state"
mkdir -p "$SD"
RC=0

add() { # add <file> <event_key> <from> <to> — append lines from..to (summary/timestamp carry the index)
  local f="$1" k="$2" i
  for ((i = $3; i <= $4; i++)); do
    printf '{"timestamp":"2026-10-19T00:00:%02dZ","event_key":"%s","summary":"s%s","source":"tool:X","v":1}\n' "$i" "$k" "$i" >>"$f"
  done
}

KA="tool_failure:Bash:ExitError"
KB="tool_failure:Write:PermissionDenied"
KC="tool_failure:Agent:ContextCancelled"

run_drain()   { "$CUR/drain.sh" --inbox "$INBOX" --state-dir "$SD" >"$FIX/out.txt" 2>"$FIX/err.txt"; RC=$?; }
run_wrapper() { "$CUR/session_drain.sh" --inbox "$INBOX" --state-dir "$SD" >"$FIX/out.txt" 2>"$FIX/err.txt"; RC=$?; }

report() { # report <prefix>
  local p="$1"
  echo "${p}exit=$RC"
  if [[ -s "$SD/clusters.json" ]]; then
    echo "${p}read=$("$REAL_JQ" -r '.total_read' "$SD/clusters.json")"
    echo "${p}candidates=$("$REAL_JQ" -r '.candidates | length' "$SD/clusters.json")"
    echo "${p}candidate_keys=$("$REAL_JQ" -c '[.candidates[].event_key]' "$SD/clusters.json")"
    echo "${p}singletons_discarded=$("$REAL_JQ" -r '.singletons_discarded' "$SD/clusters.json")"
    echo "${p}clusters_offset_after=$("$REAL_JQ" -r '.offset_after' "$SD/clusters.json")"
  elif [[ -e "$SD/clusters.json" ]]; then
    echo "${p}clusters=EMPTY-FILE bytes=0"
  else
    echo "${p}clusters=absent"
  fi
  if [[ -s "$SD/drain-offset.json" ]]; then
    echo "${p}checkpoint_offset=$("$REAL_JQ" -r '(.offset_lines // .offset)' "$SD/drain-offset.json")"
  elif [[ -e "$SD/drain-offset.json" ]]; then
    echo "${p}checkpoint=EMPTY-FILE bytes=0"
  else
    echo "${p}checkpoint=absent"
  fi
  if [[ -s "$SD/consumer-status.json" ]]; then
    echo "${p}status_state=$("$REAL_JQ" -r '.state' "$SD/consumer-status.json")"
  else
    echo "${p}status_state=absent"
  fi
  echo "${p}inbox_lines=$("$REAL_WC" -l <"$INBOX" | tr -d ' ')"
  echo "${p}stderr=$(head -c 240 "$FIX/err.txt" | tr '\n' '|' | sed "s#$FIX#<fx>#g")"
}

seed_checkpoint() { printf '{"offset":%s,"updated":"2026-10-15T04:30:35Z"}\n' "$1" >"$SD/drain-offset.json"; }

# --- named test seams (fixtures 13..17) ---------------------------------------------------------
# The consumer contract (design.md section 2.6) defines named points. When LSEL_HOOK_DIR is set and
# $LSEL_HOOK_DIR/<point> is executable, the consumer appends <point> to <state-dir>/hook-trace.log and
# runs it: exit 0 continues, exit 1 fails that step, exit 99 makes the consumer SIGKILL itself. The
# trace is written by the consumer, not by the hook, so "the boundary was reached" is observed
# independently of what the hook does. Points: after-snapshot-copy, temps-written, after-prepared,
# between-renames, after-renames, after-committed. A tree without the seam never reaches any point.
mkhook() { # mkhook <point> <body line>
  mkdir -p "$FIX/hooks"
  printf '%s\n' '#!/bin/bash' "$2" >"$FIX/hooks/$1"
  chmod +x "$FIX/hooks/$1"
  export LSEL_HOOK_DIR="$FIX/hooks"
}
trace_of() { # trace_of <prefix> <point>
  if [[ -s "$SD/hook-trace.log" ]] && grep -qx "$2" "$SD/hook-trace.log"; then echo "${1}hook_reached=true"; else echo "${1}hook_reached=false"; fi
}
pair_state() { # pair_state <prefix> — are clusters and checkpoint a coherent pair?
  local c="absent" k="absent"
  if [[ -s "$SD/clusters.json" ]]; then c="$("$REAL_JQ" -r '.offset_after' "$SD/clusters.json")"; elif [[ -e "$SD/clusters.json" ]]; then c="EMPTY"; fi
  if [[ -s "$SD/drain-offset.json" ]]; then k="$("$REAL_JQ" -r '(.offset_lines // .offset)' "$SD/drain-offset.json")"; elif [[ -e "$SD/drain-offset.json" ]]; then k="EMPTY"; fi
  local ahead=false
  if [[ "$k" =~ ^[0-9]+$ ]]; then
    if [[ ! "$c" =~ ^[0-9]+$ ]] || (( k > c )); then ahead=true; fi
  fi
  echo "${1}clusters_offset=$c"
  echo "${1}checkpoint_offset=$k"
  echo "${1}pair_consistent=$([[ "$c" == "$k" ]] && echo true || echo false)"
  echo "${1}checkpoint_ahead_of_clusters=$ahead"
  echo "${1}zero_byte_final=$([[ "$c" == "EMPTY" || "$k" == "EMPTY" ]] && echo true || echo false)"
}
status_of() { # status_of <prefix>
  if [[ -s "$SD/consumer-status.json" ]]; then echo "${1}status_state=$("$REAL_JQ" -r '.state' "$SD/consumer-status.json")"; else echo "${1}status_state=absent"; fi
}

echo "condition=$N"
case "$N" in
  1) # normal control: offset 0, 3 lines
    add "$INBOX" "$KA" 1 3
    run_drain; report "";;
  2) # offset 5 beyond a 3-line file
    add "$INBOX" "$KA" 1 3; seed_checkpoint 5
    run_drain; report "";;
  3) # condition 2, then lines 4-5 are appended
    add "$INBOX" "$KA" 1 3; seed_checkpoint 5
    run_drain; add "$INBOX" "$KB" 4 5; run_drain; report "";;
  4) # condition 3, then lines 6-7 are appended
    add "$INBOX" "$KA" 1 3; seed_checkpoint 5
    run_drain; add "$INBOX" "$KB" 4 5; run_drain; add "$INBOX" "$KC" 6 7; run_drain; report ""
    if [[ -s "$SD/clusters.json" ]]; then
      echo "lines_4_5_key_in_candidates=$("$REAL_JQ" -r --arg k "$KB" '[.candidates[].event_key] | index($k) != null' "$SD/clusters.json")"
    else
      echo "lines_4_5_key_in_candidates=n/a (no clusters.json)"
    fi;;
  5) # existing consumer lock
    add "$INBOX" "$KA" 1 3; seed_checkpoint 0; mkdir "$SD/drain.lock"
    run_wrapper; report "";;
  6) # wc counts 2 lines, a third is appended before tail
    add "$INBOX" "$KA" 1 2
    SHIM="$FIX/shim"; mkdir -p "$SHIM"
    printf '%s\n' '#!/bin/bash' 'out="$("$REAL_WC" "$@")"' 'if [[ ! -e "$SHIM_MARK" ]]; then : >"$SHIM_MARK"; printf "%s\n" "$SHIM_LINE" >>"$SHIM_INBOX"; fi' 'printf "%s\n" "$out"' >"$SHIM/wc"
    chmod +x "$SHIM/wc"
    export SHIM_MARK="$FIX/shim.mark" SHIM_INBOX="$INBOX"
    export SHIM_LINE='{"timestamp":"2026-10-19T00:00:03Z","event_key":"tool_failure:Bash:ExitError","summary":"s3","source":"tool:X","v":1}'
    PATH="$SHIM:$PATH" run_drain; report "";;
  7) # condition 6, then the next run
    add "$INBOX" "$KA" 1 2
    SHIM="$FIX/shim"; mkdir -p "$SHIM"
    printf '%s\n' '#!/bin/bash' 'out="$("$REAL_WC" "$@")"' 'if [[ ! -e "$SHIM_MARK" ]]; then : >"$SHIM_MARK"; printf "%s\n" "$SHIM_LINE" >>"$SHIM_INBOX"; fi' 'printf "%s\n" "$out"' >"$SHIM/wc"
    chmod +x "$SHIM/wc"
    export SHIM_MARK="$FIX/shim.mark" SHIM_INBOX="$INBOX"
    export SHIM_LINE='{"timestamp":"2026-10-19T00:00:03Z","event_key":"tool_failure:Bash:ExitError","summary":"s3","source":"tool:X","v":1}'
    PATH="$SHIM:$PATH" run_drain; report "run1_"
    run_drain; report "run2_";;
  8|9) # checkpoint write fails after clusters are written (8); next run on the empty checkpoint (9)
    add "$INBOX" "$KA" 1 3
    SHIM="$FIX/shim"; mkdir -p "$SHIM"
    printf '%s\n' '#!/bin/bash' 'for a in "$@"; do case "$a" in *"{offset: \$off, updated: \$ts}"*) exit 1;; esac; done' 'exec "$REAL_JQ" "$@"' >"$SHIM/jq"
    chmod +x "$SHIM/jq"
    PATH="$SHIM:$PATH" run_wrapper; report "run1_"
    echo "run1_checkpoint_bytes=$("$REAL_WC" -c <"$SD/drain-offset.json" | tr -d ' ')"
    if [[ "$N" == "9" ]]; then run_wrapper; report "run2_"; echo "run2_checkpoint_bytes=$("$REAL_WC" -c <"$SD/drain-offset.json" | tr -d ' ')"; fi;;
  10) # the inbox file is replaced by a different file of at least the checkpoint length
    add "$INBOX" "$KA" 1 3
    run_drain; report "run1_"
    NEW="$FIX/inbox.new"; add "$NEW" "$KB" 1 5; mv "$NEW" "$INBOX"
    run_drain; report "run2_"
    echo "run2_lineage_check=none-performed";;
  11) # stress: a background appender runs while the consumer drains repeatedly.
      # Invariant (implementation-agnostic): sum of rows clustered == final checkpoint offset == final line count.
    add "$INBOX" "$KA" 1 20
    ( for ((i = 21; i <= 220; i++)); do add "$INBOX" "$KA" "$i" "$i"; sleep 0.001; done ) &
    APP=$!
    SUM=0
    for r in $(seq 1 25); do
      run_drain
      if [[ -s "$SD/clusters.json" ]]; then SUM=$((SUM + $("$REAL_JQ" -r '.total_read // 0' "$SD/clusters.json"))); fi
      sleep 0.01
    done
    wait "$APP"
    run_drain
    if [[ -s "$SD/clusters.json" ]]; then SUM=$((SUM + $("$REAL_JQ" -r '.total_read // 0' "$SD/clusters.json"))); fi
    FINAL_OFF="n/a"; [[ -s "$SD/drain-offset.json" ]] && FINAL_OFF="$("$REAL_JQ" -r '(.offset_lines // .offset)' "$SD/drain-offset.json")"
    FINAL_LINES="$("$REAL_WC" -l <"$INBOX" | tr -d ' ')"
    echo "sum_read=$SUM"
    echo "final_checkpoint_offset=$FINAL_OFF"
    echo "final_inbox_lines=$FINAL_LINES"
    echo "invariant_sum_eq_offset_eq_lines=$([[ "$SUM" == "$FINAL_OFF" && "$FINAL_OFF" == "$FINAL_LINES" ]] && echo true || echo false)"
    echo "status_state=$([[ -s "$SD/consumer-status.json" ]] && "$REAL_JQ" -r '.state' "$SD/consumer-status.json" || echo absent)";;
  12) # a drain succeeds (clusters.json now holds candidates), then the checkpoint becomes a zero-byte file
      # and one more row arrives: the next run must not destroy the staged candidates.
    add "$INBOX" "$KA" 1 3
    run_wrapper; report "run1_"
    CL_BEFORE="$(shasum -a 256 "$SD/clusters.json" | cut -d' ' -f1)"
    : >"$SD/drain-offset.json"
    add "$INBOX" "$KB" 4 4
    run_wrapper; report "run2_"
    CL_AFTER="absent"; [[ -e "$SD/clusters.json" ]] && CL_AFTER="$(shasum -a 256 "$SD/clusters.json" | cut -d' ' -f1)"
    echo "clusters_unchanged_across_run2=$([[ "$CL_BEFORE" == "$CL_AFTER" ]] && echo true || echo false)"
    echo "clusters_bytes_after_run2=$([[ -e "$SD/clusters.json" ]] && "$REAL_WC" -c <"$SD/clusters.json" | tr -d ' ' || echo absent)";;
  13) # race through the seam: a row is appended right after the snapshot copy.
      # Non-vacuous GREEN: hook_reached=true AND run1 clusters exactly the 2 rows copied AND run2 picks up the third.
    add "$INBOX" "$KA" 1 2
    export HOOK_MARK="$FIX/hook.mark" HOOK_INBOX="$INBOX"
    export HOOK_LINE='{"timestamp":"2026-10-19T00:00:03Z","event_key":"tool_failure:Bash:ExitError","summary":"s3","source":"tool:X","v":1}'
    mkhook after-snapshot-copy 'if [[ ! -e "$HOOK_MARK" ]]; then : >"$HOOK_MARK"; printf "%s\n" "$HOOK_LINE" >>"$HOOK_INBOX"; fi'
    run_drain; trace_of "run1_" after-snapshot-copy; echo "run1_read=$([[ -s "$SD/clusters.json" ]] && "$REAL_JQ" -r '.total_read' "$SD/clusters.json" || echo none)"; pair_state "run1_"
    run_drain; echo "run2_read=$([[ -s "$SD/clusters.json" ]] && "$REAL_JQ" -r '.total_read' "$SD/clusters.json" || echo none)"; pair_state "run2_"; status_of "run2_"
    echo "inbox_lines=$("$REAL_WC" -l <"$INBOX" | tr -d ' ')";;
  14) # commit failure at the rename boundary (the hook fails the between-renames step)
    add "$INBOX" "$KA" 1 3
    mkhook between-renames 'exit 1'
    run_wrapper; trace_of "" between-renames; echo "wrapper_exit=$RC"; pair_state ""; status_of ""
    echo "checkpoint_zero_byte=$([[ -e "$SD/drain-offset.json" && ! -s "$SD/drain-offset.json" ]] && echo true || echo false)";;
  15) # crash matrix with an existing committed pair. A first clean run commits the pair at offset 3, rows
      # 4-6 arrive, the consumer is SIGKILLed at the named point while committing offset 6, then a clean
      # run recovers. Optional third argument: lost | tamper (the checkpoint temporary is deleted or
      # altered by the crash hook before the kill). The seam after-recovery fires between recovery and
      # consumption, so the pair is observed BEFORE anything is consumed again.
    POINT="${2:-}"; MODE="${3:-plain}"
    case "$POINT" in temps-written|after-prepared|between-renames|after-renames|after-committed) ;; *) echo "usage: conditions.sh 15 <temps-written|after-prepared|between-renames|after-renames|after-committed> [lost|tamper]" >&2; exit 64;; esac
    case "$MODE" in plain|lost|tamper) ;; *) echo "usage: third argument plain|lost|tamper" >&2; exit 64;; esac
    add "$INBOX" "$KA" 1 3
    run_wrapper
    add "$INBOX" "$KB" 4 6
    export HOOK_REC="$FIX/at-recovery.txt"
    # The hook observes the pair as recovery left it (line 1) and compares the recovery record's
    # recorded `after` hashes against the files actually present at the seam (line 2) — recovery
    # wrote the record immediately before this seam, and the inbox has not been read yet.
    mkhook after-recovery 'printf "clusters=%s checkpoint=%s\n" "$(jq -r .offset_after "$2/clusters.json" 2>/dev/null || echo absent)" "$(jq -r "(.offset_lines // .offset)" "$2/drain-offset.json" 2>/dev/null || echo absent)" >"$HOOK_REC"; h="$2/recovery-record.json"; if [[ -s "$h" ]]; then c="$(jq -r ".after.clusters_sha256 // empty" "$h" 2>/dev/null)"; k="$(jq -r ".after.checkpoint_sha256 // empty" "$h" 2>/dev/null)"; a="$(shasum -a 256 "$2/clusters.json" 2>/dev/null | cut -d" " -f1)"; b="$(shasum -a 256 "$2/drain-offset.json" 2>/dev/null | cut -d" " -f1)"; if [[ -n "$c" && "$c" == "$a" && -n "$k" && "$k" == "$b" ]]; then echo "recovery_hashes_match=true" >>"$HOOK_REC"; else echo "recovery_hashes_match=false" >>"$HOOK_REC"; fi; else echo "recovery_hashes_match=absent" >>"$HOOK_REC"; fi'
    case "$MODE" in
      plain)  CRASH_BODY='exit 99';;
      lost)   CRASH_BODY='rm -f "$2"/drain-offset.json.*.tmp; exit 99';;
      tamper) CRASH_BODY='for f in "$2"/drain-offset.json.*.tmp; do printf "x" >>"$f"; done; exit 99';;
    esac
    mkhook "$POINT" "$CRASH_BODY"
    run_wrapper; trace_of "crash_" "$POINT"; pair_state "crash_"
    rm -f "$FIX/hooks/$POINT"
    run_wrapper
    echo "after_recovery_seam_count=$([[ -s "$SD/hook-trace.log" ]] && grep -cx after-recovery "$SD/hook-trace.log" || echo 0)"
    echo "at_recovery=$([[ -s "$HOOK_REC" ]] && head -n 1 "$HOOK_REC" || echo none)"
    echo "recovery_hashes_match=$([[ -s "$HOOK_REC" ]] && sed -n 2p "$HOOK_REC" || echo none)"
    echo "recovery_branch=$([[ -s "$SD/recovery-record.json" ]] && "$REAL_JQ" -r '.branch' "$SD/recovery-record.json" || echo none)"
    pair_state "final_"; status_of "final_"
    if [[ -s "$SD/commits.jsonl" ]]; then
      echo "slices=$("$REAL_JQ" -s -r 'map("\(.from)-\(.to)") | join(",")' "$SD/commits.jsonl")"
      echo "slices_tile=$("$REAL_JQ" -s -r 'reduce .[] as $r ({e:0,ok:true}; if $r.from == .e then .e = $r.to else .ok = false end) | "\(.ok) end=\(.e)"' "$SD/commits.jsonl")"
    else
      echo "slices=none"
      echo "slices_tile=none"
    fi;;
  16) # the inbox is swapped between the snapshot copy and the post-check
    add "$INBOX" "$KA" 1 3
    run_drain
    add "$INBOX" "$KB" 4 5
    NEW="$FIX/inbox.swap"; add "$NEW" "$KC" 1 6
    export HOOK_MARK="$FIX/hook.mark" HOOK_INBOX="$INBOX" HOOK_NEW="$NEW"
    mkhook after-snapshot-copy 'if [[ ! -e "$HOOK_MARK" ]]; then : >"$HOOK_MARK"; mv "$HOOK_NEW" "$HOOK_INBOX"; fi'
    run_wrapper; trace_of "" after-snapshot-copy; status_of ""; pair_state ""
    echo "swapped_file_keys_in_candidates=$([[ -s "$SD/clusters.json" ]] && "$REAL_JQ" -r --arg k "$KC" '[.candidates[].event_key] | index($k) != null' "$SD/clusters.json" || echo n/a)";;
  17) # a fully drained, lineage-verified inbox: the next run is NOOP_AT_TAIL
    add "$INBOX" "$KA" 1 3
    run_drain; report "run1_"
    run_drain; report "run2_";;
  18) # a legacy offset-only checkpoint (offset 5) meets its FIRST consumer run only after the inbox has
      # already grown past the offset (7 lines): no earlier run ever observed a mismatch.
    add "$INBOX" "$KA" 1 3
    add "$INBOX" "$KB" 4 5
    seed_checkpoint 5
    add "$INBOX" "$KC" 6 7
    run_drain; report ""
    echo "rows_6_7_key_in_candidates=$([[ -s "$SD/clusters.json" ]] && "$REAL_JQ" -r --arg k "$KC" '[.candidates[].event_key] | index($k) != null' "$SD/clusters.json" || echo n/a)";;
esac
exit 0
