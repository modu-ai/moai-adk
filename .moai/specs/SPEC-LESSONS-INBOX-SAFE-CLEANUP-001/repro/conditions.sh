#!/usr/bin/env bash
# conditions.sh <N> — synthetic reproduction of drain failure condition N (1..12) on the LSEL drain pair
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
case "$N" in 1|2|3|4|5|6|7|8|9|10|11|12) ;; *) echo "usage: conditions.sh <1..12>" >&2; exit 64;; esac

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
esac
exit 0
