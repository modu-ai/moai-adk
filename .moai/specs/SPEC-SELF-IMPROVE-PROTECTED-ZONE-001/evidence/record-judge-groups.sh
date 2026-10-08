#!/usr/bin/env bash
# record-judge-groups.sh — runs judge-probe.sh once per acceptance-criterion group and prints, for
# each, the command, its verbatim stdout and its exit code, in the ledger layout used by
# acceptance.md §E. The input is either a recorded TSV (REPLAY: the RED record) or a moai binary
# (LIVE: the form that flips). Usage (from the repository root):
#   bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/record-judge-groups.sh <recorded.tsv | moai-binary>
set -u
HERE="$(cd "$(dirname "$0")" && pwd -P)"
IN="${1:?usage: record-judge-groups.sh <recorded.tsv | moai-binary>}"
if [ -f "$IN" ] && ! [ -x "$IN" ]; then shown="<recorded.tsv>"; mode=REPLAY; else shown="<moai-binary>"; mode=LIVE; fi
echo "# mode: $mode"
for g in '^R' '^(P|S)' '^LP' '^B' '^(MS|E)' '^(C|N)'; do
  echo "## cmd: bash judge-probe.sh -o '$g' $shown"
  bash "$HERE/judge-probe.sh" -o "$g" "$IN"
  echo "## exit: $?"
  echo
done
