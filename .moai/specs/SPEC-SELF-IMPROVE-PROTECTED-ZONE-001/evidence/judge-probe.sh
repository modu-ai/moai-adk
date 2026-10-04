#!/usr/bin/env bash
# judge-probe.sh — the verdict-bearing half of the probe pair. probe-protected-zone.sh
# REPORTS what the real PreToolUse handler decided; this script JUDGES those rows against
# the post-implementation expectation table below and exits non-zero on any divergence.
#
# Usage (from the repository root):
#   bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/judge-probe.sh [-o ID-REGEX] [probe.tsv | moai-binary]
#     -o ID-REGEX   judge only the case ids matching the (extended) regex, e.g. '^R' or '^(P|S)'
#     - an existing non-executable file -> judged as a recorded probe output
#     - anything else / none            -> the probe is run first (a binary path is passed
#                                          through; with no argument the probe builds one
#                                          from ./cmd/moai)
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

# id  want-decision  want-sentinel  want-reason-substring   ("-" = no constraint / none)
PZ=HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION
EXPECT="C1 deny HARNESS_FROZEN_HOOK_VIOLATION -
C2 deny HARNESS_FROZEN_RULE_VIOLATION -
C3 deny HARNESS_FROZEN_INSTRUCTION_VIOLATION -
C4 deny HARNESS_FROZEN_AGENT_VIOLATION -
R1 deny $PZ route=human
R2 deny $PZ route=human
R3 deny $PZ route=human
R4 deny $PZ route=human
R5 deny $PZ route=human
R6 deny $PZ route=human
R7 deny $PZ route=human
R8 deny $PZ route=human
R9 deny $PZ route=human
R10 deny $PZ route=human
R11 deny $PZ route=human
R12 deny $PZ route=human
R13 deny $PZ route=human
R14 deny $PZ route=human
R15 deny $PZ route=human
R16 deny $PZ route=human
R17 deny $PZ route=human
R18 deny $PZ route=human
R19 deny $PZ route=human
R20 deny $PZ route=human
P1 deny HARNESS_FROZEN_HOOK_VIOLATION -
P2 deny HARNESS_FROZEN_HOOK_VIOLATION -
P3 deny HARNESS_FROZEN_HOOK_VIOLATION -
P4 deny HARNESS_FROZEN_HOOK_VIOLATION -
S1 deny HARNESS_FROZEN_HOOK_VIOLATION -
B1 deny $PZ route=human
B2 deny $PZ route=human
B3 deny $PZ route=human
M1 deny $PZ manifest=invalid
M2 deny HARNESS_FROZEN_HOOK_VIOLATION -
M3 allow - -
M4 deny $PZ category=probe_docs
M5 allow - -
M6 deny $PZ category=probe_docs
E1 deny $PZ route=human
N1 allow - -
N2 allow - -
N3 allow - -
N4 allow - -
N5 allow - -
N6 allow - -
N7 allow - -
N8 allow - -"

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
  if [ "$dec" != "$want_dec" ] || [ "$sent" != "$want_sent" ]; then
    echo "FAIL $id got=$dec/$sent want=$want_dec/$want_sent"
    fail=$((fail+1)); continue
  fi
  if [ "$want_reason" != "-" ]; then
    case "$reason" in
      *"$want_reason"*) ;;
      *) echo "FAIL $id reason-missing '$want_reason' (reason: $reason)"; fail=$((fail+1));;
    esac
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
