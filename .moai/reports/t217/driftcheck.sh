#!/usr/bin/env bash
# t217 D4 measurement. Two variants of the wrapper-pair drift check.
#
#   unguarded : the form transcribed into AC-SSS-016 (missing the `[ -f ]` guard)
#   guarded   : the canonical form from CLAUDE.local.md 2.3
#
# Arg 1 selects the directory axis to check.
set -u
mode="${1:?usage: driftcheck.sh unguarded|guarded [dir]}"
dir="${2:-internal/template/templates/.claude/hooks/moai}"

case "$mode" in
  unguarded)
    for f in "$dir"/*.tmpl; do
      b="${f%.tmpl}"
      diff -q "$b" "$f" >/dev/null 2>&1 || echo "DRIFT $b"
    done
    ;;
  guarded)
    for f in "$dir"/*.tmpl; do
      b="${f%.tmpl}"
      [ -f "$b" ] && { diff -q "$b" "$f" >/dev/null 2>&1 || echo "DRIFT $b"; }
    done
    ;;
  pairaxis)
    # The axis this SPEC actually touches: deployed .sh <-> template .sh.tmpl
    for f in "$dir"/*.sh.tmpl; do
      base=".claude/hooks/moai/$(basename "${f%.tmpl}")"
      [ -f "$base" ] && { diff -q "$base" "$f" >/dev/null 2>&1 || echo "DRIFT $base"; }
    done
    ;;
  *)
    echo "unknown mode: $mode" >&2
    exit 2
    ;;
esac
