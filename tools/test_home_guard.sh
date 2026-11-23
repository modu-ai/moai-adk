#!/bin/sh
# nextest wrapper (SPEC-20261123-test-home-isolation). nextest calls it with the test binary and its
# arguments; it runs them under an empty temporary HOME that holds decoy files, then reports a test
# that read or changed that HOME with a line that starts with `HOME-GUARD:`.
#
#   MOAI_HOME_GUARD      enforce (default): exit 1 when the test itself exited 0 and touched its HOME.
#                        observe: print the line and keep the test's exit code.
#   MOAI_HOME_GUARD_LOG  a file that receives each line too (best effort, never fails a test).
#   MOAI_HOME_GUARD_BASELINE  the baseline file; default tests/fixtures/test_home_guard/baseline.tsv
#                        next to this script's folder. A listed test that touched only the decoy
#                        user-assets.json (and its folders) gets a `tolerated` line and no failure.
#
# Exit codes: the test's own; 1 for an enforced touch; 70 when the folder cannot be made (the test
# does not run); 64 when called without a test. A read is seen through the access time of the decoys,
# a change through the modification time of any entry: both are compared with a reference file made
# after the decoys, so a host that keeps no access time does not see reads (the repository test of
# this wrapper reads a decoy and fails there).

[ "$#" -gt 0 ] || exit 64

t="${TMPDIR:-/tmp}"
t="${t%/}"
d="$t/moai-test-home.$$"
ref="$t/moai-test-home.$$.ref"

# Only these two paths, built above from TMPDIR and this process id, are ever removed.
trap 'rm -rf "$d" "$ref"' EXIT
trap 'exit 143' HUP INT TERM

mkdir -m 700 "$d" "$d/.moai" "$d/.claude" 2>/dev/null || exit 70
: > "$d/.moai/user-assets.json" || exit 70
: > "$d/.moai/language.yaml" || exit 70
: > "$d/.claude/settings.json" || exit 70
touch -a -t 200001010000 "$d/.moai/user-assets.json" "$d/.moai/language.yaml" \
  "$d/.claude/settings.json" "$d/.moai" "$d/.claude" "$d" || exit 70
: > "$ref" || exit 70

# cargo and rustup keep finding their toolchain where they found it before HOME was replaced.
if [ -n "${HOME:-}" ]; then
  ch="${CARGO_HOME:-$HOME/.cargo}"
  rh="${RUSTUP_HOME:-$HOME/.rustup}"
else
  ch="${CARGO_HOME:-}"
  rh="${RUSTUP_HOME:-}"
fi
[ -n "$ch" ] && export CARGO_HOME="$ch"
[ -n "$rh" ] && export RUSTUP_HOME="$rh"

HOME="$d" USERPROFILE="$d" "$@"
rc=$?

hits=$(find "$d" \( -newer "$ref" -o -anewer "$ref" \) 2>/dev/null)
if [ -n "$hits" ]; then
  # A test listed in the baseline (`<binary><TAB><test name>` lines) is tolerated when it touched
  # nothing but the decoy user-assets.json and the two folders above it. nextest names the test
  # after `--exact`; the binary is its file name without the 16-digit hash.
  tolerated=0
  bin="${1##*/}"
  bin="${bin%-????????????????}"
  name=""
  prev=""
  for a in "$@"; do
    [ "$prev" = --exact ] && name="$a"
    prev="$a"
  done
  here="${0%/*}"
  [ "$here" = "$0" ] && here=.
  list="${MOAI_HOME_GUARD_BASELINE:-$here/../tests/fixtures/test_home_guard/baseline.tsv}"
  if [ -n "$name" ] && [ -f "$list" ]; then
    key=$(printf '%s\t%s' "$bin" "$name")
    if grep -qxF -- "$key" "$list" 2>/dev/null; then
      # The first find already listed the folders it read itself, so the hits are filtered, not
      # looked up again: a second find would see its own reads as the test's.
      extra=$(printf '%s\n' "$hits" | grep -vxF -e "$d" -e "$d/.moai" -e "$d/.moai/user-assets.json")
      [ -z "$extra" ] && tolerated=1
    fi
  fi
  hits=$(printf '%s' "$hits" | tr '\n' ' ')
  if [ "$tolerated" -eq 1 ]; then
    line="HOME-GUARD: tolerated (baseline): $* -> $hits"
  else
    line="HOME-GUARD: touched its HOME: $* -> $hits"
  fi
  printf '%s\n' "$line" >&2
  if [ -n "${MOAI_HOME_GUARD_LOG:-}" ]; then
    (exec 2>/dev/null; printf '%s\n' "$line" >> "$MOAI_HOME_GUARD_LOG") || :
  fi
  if [ "$tolerated" -eq 0 ] && [ "$rc" -eq 0 ] && [ "${MOAI_HOME_GUARD:-enforce}" != observe ]; then
    exit 1
  fi
fi
exit "$rc"
