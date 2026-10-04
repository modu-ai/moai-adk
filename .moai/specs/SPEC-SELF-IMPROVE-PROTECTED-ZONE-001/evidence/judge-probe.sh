#!/usr/bin/env bash
# judge-probe.sh — the verdict-bearing half of the probe pair. probe-protected-zone.sh
# REPORTS what the real PreToolUse handler decided; this script JUDGES those rows against
# the post-implementation expectation table below and exits non-zero on any divergence.
#
# Two forms, and they are not interchangeable:
#   LIVE    judge-probe.sh [-o ID-REGEX] [moai-binary]
#           runs the probe against the tree (builds ./cmd/moai when no binary is given) and
#           judges what it just observed. THIS is the form whose output changes when the
#           implementation lands, and the form every green-path cell names.
#   REPLAY  judge-probe.sh [-o ID-REGEX] <recorded.tsv>
#           judges a recorded probe output. It is evidence of what the base tree did and can
#           never flip; it carries the RED record, not the green.
#
#   -o ID-REGEX  judge only the case ids matching the (extended) regex, e.g. '^R' or '^(P|S)'
#   A non-executable regular file argument selects REPLAY; anything else selects LIVE.
#
# Exit codes:  0 every expectation met and the swept set is exactly the expected size
#              1 at least one divergence, an unexpected/missing id, or a wrong swept count
#              2 the probe itself could not run
#
# Output: one `FAIL <id> ...` row per divergence, then `JUDGE swept=<n> expected=<n> fail=<n>`.
set -u
HERE="$(cd "$(dirname "$0")" && pwd -P)"
ONLY='.'
if [ "${1:-}" = "-o" ]; then ONLY="${2:?-o needs a regex}"; shift 2; fi
ARG="${1:-}"
TSV="$(mktemp)"
trap 'rm -f "$TSV"' EXIT
if [ -n "$ARG" ] && [ -f "$ARG" ] && ! [ -x "$ARG" ]; then
  cp "$ARG" "$TSV"
else
  bash "$HERE/probe-protected-zone.sh" ${ARG:+"$ARG"} > "$TSV" || { echo "PROBE_FAILED"; exit 2; }
fi

# Expectation rows: id  want-decision  want-sentinel  want-reason-substrings (comma list; "-" = none)
PZ=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
RT="route=human,category=,next=return-blocker-report"
EXPECT="C1 deny HARNESS_FROZEN_HOOK_VIOLATION -
C2 deny HARNESS_FROZEN_RULE_VIOLATION -
C3 deny HARNESS_FROZEN_INSTRUCTION_VIOLATION -
C4 deny HARNESS_FROZEN_AGENT_VIOLATION -"
for i in $(seq 1 21); do EXPECT="$EXPECT
R$i deny $PZ $RT"; done
# R21 (.claude/skills/moai/) is a baseline-list member; it keeps a legacy sentinel when the
# compiled floor covers it and the new sentinel otherwise — either is a deny with a route.
EXPECT="$EXPECT
P1 deny HARNESS_FROZEN_HOOK_VIOLATION -
P2 deny HARNESS_FROZEN_HOOK_VIOLATION -
P3 deny HARNESS_FROZEN_HOOK_VIOLATION -
P4 deny HARNESS_FROZEN_HOOK_VIOLATION -
P5 deny HARNESS_FROZEN_HOOK_VIOLATION -
P6 deny HARNESS_FROZEN_HOOK_VIOLATION -
S1 deny HARNESS_FROZEN_HOOK_VIOLATION -
LP1 deny $PZ $RT
LP2 deny $PZ $RT"
for i in $(seq 1 13); do EXPECT="$EXPECT
B$i deny $PZ $RT"; done
EXPECT="$EXPECT
MS1 deny $PZ manifest=invalid,route=human
MS2 deny HARNESS_FROZEN_HOOK_VIOLATION -
MS3 allow - -
MS4 deny $PZ category=probe_docs,route=human
MS5 allow - -
MS6 deny $PZ category=probe_docs,route=human
MS7 deny $PZ category=probe_base,route=human
MS8 deny $PZ manifest=invalid,route=human
E1 deny $PZ $RT
E2 deny $PZ $RT
E3 deny $PZ $RT"
for i in 1 2 3 4 5 6 7 8 9; do EXPECT="$EXPECT
N$i allow - -"; done

# R21 is judged on the decision and a route marker only (sentinel may be legacy or new).
fail=0; swept=0; expected=0
SEEN=" "   # space-delimited id list (portable: macOS ships bash 3.2, no associative arrays)
while IFS=$'\t' read -r id tool agent dec sent path reason; do
  case "$id" in SWEPT=*) continue;; esac
  [ -z "$id" ] && continue
  printf '%s\n' "$id" | grep -Eq "$ONLY" || continue
  swept=$((swept+1)); SEEN="$SEEN$id "
  line="$(printf '%s\n' "$EXPECT" | awk -v i="$id" '$1==i {print $2" "$3" "$4}')"
  if [ -z "$line" ]; then
    echo "FAIL $id unexpected-id (no expectation row)"; fail=$((fail+1)); continue
  fi
  set -- $line
  want_dec="$1"; want_sent="$2"; want_reason="$3"
  if [ "$id" = "R21" ]; then
    if [ "$dec" != "deny" ]; then echo "FAIL R21 got=$dec/$sent want=deny/<any sentinel>"; fail=$((fail+1)); fi
    continue
  fi
  if [ "$dec" != "$want_dec" ] || [ "$sent" != "$want_sent" ]; then
    echo "FAIL $id got=$dec/$sent want=$want_dec/$want_sent"
    fail=$((fail+1)); continue
  fi
  if [ "$want_reason" != "-" ]; then
    oldifs="$IFS"; IFS=','
    for sub in $want_reason; do
      case "$reason" in
        *"$sub"*) ;;
        *) echo "FAIL $id reason-missing '$sub' (reason: $reason)"; fail=$((fail+1));;
      esac
    done
    IFS="$oldifs"
  fi
done < "$TSV"
while read -r eid _; do
  [ -z "$eid" ] && continue
  printf '%s\n' "$eid" | grep -Eq "$ONLY" || continue
  expected=$((expected+1))
  case "$SEEN" in *" $eid "*) ;; *) echo "FAIL $eid missing-from-probe"; fail=$((fail+1));; esac
done <<EOF
$EXPECT
EOF
[ "$swept" -ne "$expected" ] && { echo "FAIL swept-count swept=$swept expected=$expected"; fail=$((fail+1)); }
[ "$expected" -eq 0 ] && { echo "FAIL empty-selection (the -o regex matched no expectation row)"; fail=$((fail+1)); }
echo "JUDGE swept=$swept expected=$expected fail=$fail"
[ "$fail" -eq 0 ] && exit 0 || exit 1
