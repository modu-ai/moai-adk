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
# does not run) or, in enforce mode, when the comparison afterwards could not run (the reference
# file or the HOME vanished) while the test itself passed; 64 when called without a test. A read is seen through the access time of the decoys,
# a change through the modification time of any entry: both are compared with a reference file made
# after the decoys, so a host that keeps no access time does not see reads (the repository test of
# this wrapper reads a decoy and fails there).

[ "$#" -gt 0 ] || exit 64

t="${TMPDIR:-/tmp}"
t="${t%/}"
trap 'exit 143' HUP INT TERM

# The folder and the reference file are made by mktemp under names nobody can predict, and a trap
# that removes one is set only after mktemp made it: a folder or a link that already carries a
# name (a leftover of an earlier process with the same id) is never created over, written through
# or removed. Only the two paths mktemp printed are ever removed.
d=$(mktemp -d "$t/moai-test-home.XXXXXX") || exit 70
trap 'rm -rf "$d"' EXIT

mkdir -m 700 "$d/.moai" "$d/.claude" 2>/dev/null || exit 70
: > "$d/.moai/user-assets.json" || exit 70
: > "$d/.moai/language.yaml" || exit 70
: > "$d/.claude/settings.json" || exit 70
touch -a -t 200001010000 "$d/.moai/user-assets.json" "$d/.moai/language.yaml" \
  "$d/.claude/settings.json" "$d/.moai" "$d/.claude" "$d" || exit 70
ref=$(mktemp "$t/moai-test-home-ref.XXXXXX") || exit 70
trap 'rm -rf "$d"; rm -f "$ref"' EXIT

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

# Evidence for a `touched` verdict (SPEC-20261124-home-guard-flake). It runs only after that verdict,
# so a test that touched nothing pays for none of it, and it never changes the verdict: when
# MOAI_HOME_GUARD_LOG names a file it appends `HOME-GUARD-DETAIL:` lines, and a failure to write
# them is ignored.
#   kind=hit    one per hit entry: its path relative to the HOME, its access and modification time,
#               the modification time of the reference file (made just before the test started) and
#               the time of the comparison, in seconds with a fraction (whole seconds without perl).
#   kind=peer   one per other `moai-test-home.*` folder under TMPDIR: the access time of its
#               user-assets.json, read with stat only; nothing inside that folder is listed or opened.
#   kind=proc   one per running process whose command file name starts with `moai`: its id and that
#               file name, never its arguments or its folder (`kind=procs procs=unavailable` when
#               the process table cannot be read).
st_time() {
  case "$gnu:$frac:$1" in
    0:1:a) stat -f %Fa "$2" ;;
    0:1:m) stat -f %Fm "$2" ;;
    0:0:a) stat -f %a "$2" ;;
    0:0:m) stat -f %m "$2" ;;
    1:1:a) stat -c %.9X "$2" ;;
    1:1:m) stat -c %.9Y "$2" ;;
    1:0:a) stat -c %X "$2" ;;
    1:0:m) stat -c %Y "$2" ;;
  esac 2>/dev/null
}

detail_body() {
  gnu=0
  stat --version > /dev/null 2>&1 && gnu=1
  frac=0
  if now=$(perl -MTime::HiRes=time -e 'printf "%.6f\n", time' 2> /dev/null) && [ -n "$now" ]; then
    frac=1
  else
    now=$(date +%s)
  fi
  refm=$(st_time m "$ref")
  refm=${refm:-none}
  key="$bin:$name"
  printf '%s\n' "$hit_list" | while IFS= read -r h; do
    [ -n "$h" ] || continue
    rel=${h#"$d"}
    rel=${rel#/}
    [ -n "$rel" ] || rel=.
    at=$(st_time a "$h")
    mt=$(st_time m "$h")
    printf 'HOME-GUARD-DETAIL: kind=hit test=%s path=%s atime=%s mtime=%s ref=%s end=%s\n' \
      "$key" "$rel" "${at:-none}" "${mt:-none}" "$refm" "$now"
  done
  for p in "$t"/moai-test-home.*; do
    [ -d "$p" ] || continue
    [ "$p" = "$d" ] && continue
    pa=$(st_time a "$p/.moai/user-assets.json")
    printf 'HOME-GUARD-DETAIL: kind=peer test=%s folder=%s atime=%s\n' "$key" "${p##*/}" "${pa:-none}"
  done
  if procs=$(ps -axo pid=,comm= 2> /dev/null) && [ -n "$procs" ]; then
    printf '%s\n' "$procs" | while read -r pid comm; do
      n=${comm##*/}
      case "$n" in
        moai*) printf 'HOME-GUARD-DETAIL: kind=proc test=%s pid=%s name=%s\n' "$key" "$pid" "$n" ;;
      esac
    done
  else
    printf 'HOME-GUARD-DETAIL: kind=procs test=%s procs=unavailable\n' "$key"
  fi
}

HOME="$d" USERPROFILE="$d" "$@"
rc=$?

hits=$(find "$d" \( -newer "$ref" -o -anewer "$ref" \) 2>/dev/null)
frc=$?
if [ "$frc" -ne 0 ]; then
  # The comparison itself failed (the reference file or the HOME is gone): an empty answer would
  # read as "touched nothing", so the failure is reported and, in enforce mode, stops a test that
  # passed. A test that failed keeps its own exit code.
  line="HOME-GUARD: could not compare its HOME with the reference (find exit $frc): $*"
  printf '%s\n' "$line" >&2
  if [ -n "${MOAI_HOME_GUARD_LOG:-}" ]; then
    (exec 2>/dev/null; printf '%s\n' "$line" >> "$MOAI_HOME_GUARD_LOG") || :
  fi
  if [ "$rc" -eq 0 ] && [ "${MOAI_HOME_GUARD:-enforce}" != observe ]; then
    exit 70
  fi
  exit "$rc"
fi
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
  hit_list=$hits
  hits=$(printf '%s' "$hits" | tr '\n' ' ')
  if [ "$tolerated" -eq 1 ]; then
    line="HOME-GUARD: tolerated (baseline): $* -> $hits"
  else
    line="HOME-GUARD: touched its HOME: $* -> $hits"
  fi
  printf '%s\n' "$line" >&2
  if [ -n "${MOAI_HOME_GUARD_LOG:-}" ]; then
    (exec 2>/dev/null; printf '%s\n' "$line" >> "$MOAI_HOME_GUARD_LOG") || :
    if [ "$tolerated" -eq 0 ]; then
      (exec 2>/dev/null; detail_body >> "$MOAI_HOME_GUARD_LOG") || :
    fi
  fi
  if [ "$tolerated" -eq 0 ] && [ "$rc" -eq 0 ] && [ "${MOAI_HOME_GUARD:-enforce}" != observe ]; then
    exit 1
  fi
fi
exit "$rc"
