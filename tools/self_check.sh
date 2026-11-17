#!/usr/bin/env bash
# The machine items of the pre-PR self-check (card C-77,
# skills/moai/references/internal/self-check.md) from one command; tests/self_check_test.py
# pins the item list.
#
#   tools/self_check.sh                       run every item in this working folder
#   SELF_CHECK_DRY_RUN=1 tools/self_check.sh  print the items and the last line, run nothing
#
# Items, in order: bin-moai, bin-moai-cli (whole `cargo test --bin`, never filtered: the help
# rules live there), clippy, fmt, git-env-scrub, home-isolation, ac-command-lint,
# package-inspect, package-size (no file under skills/moai over 65536 bytes, the LIMIT of
# src/profile.rs that install, update and package-inspect read each package file through),
# manifest (the source manifest check).
#
# It is not the merge evidence: tools/local_check.sh checks the committed head merged with the
# base tip and prints the line the merge gate reads. This one runs in place, before or after a
# commit, keeps going past a failure and prints one line per item:
#   self-check: <item> exit=<code> log=<file> cmd=<command>
#   self-check: <item> skip (no <path>)        an input the item needs is missing
# package-size prints `exit=<0|1> limit=65536 largest=<bytes> <path> over=<n> log=<file>`; the
# files over the bound are listed in its log. Each item's output goes to its log
# (SELF_CHECK_LOG_DIR, default target/self-check, relative to the repository root). The last
# line is:
#   self-check: head=<40-hex|unknown> result=<pass|fail> [dirty=1 ][dry-run=1 ]failed=<items|none> logs=<dir>
# dirty=1 says the folder held uncommitted changes, so the run checked more than the head; it
# does not fail the run. A dry run checked nothing and says result=fail. Exit 0 only for
# result=pass, 1 otherwise, 2 for a usage error.
#
# MOAI_HEAVY_SLOT_MAX (default 3), MOAI_HEAVY_SLOT_MIN_FREE_GB (default 50) and
# MOAI_HEAVY_SLOT_NAME (default self-check): when `moai heavy-slot-acquire` exists, the heavy
# cargo items run in one of the machine's heavy slots, as in tools/local_check.sh.
set -u

if [ "$#" -ne 0 ]; then
  echo "usage: tools/self_check.sh (no arguments; SELF_CHECK_DRY_RUN=1 lists the items)" >&2
  exit 2
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root" || exit 2

limit=65536 # src/profile.rs LIMIT
dry_run="${SELF_CHECK_DRY_RUN:-0}"
log_dir="${SELF_CHECK_LOG_DIR:-target/self-check}"
slot_max="${MOAI_HEAVY_SLOT_MAX:-3}"
slot_min_free="${MOAI_HEAVY_SLOT_MIN_FREE_GB:-50}"
slot_name="${MOAI_HEAVY_SLOT_NAME:-self-check}"
head="$(git rev-parse --verify -q 'HEAD^{commit}' 2>/dev/null || echo unknown)"
failed=""

trap 'exit 130' INT
trap 'exit 143' TERM

use_slot=0
if command -v moai >/dev/null 2>&1 && moai heavy-slot-acquire --help >/dev/null 2>&1; then
  use_slot=1
fi

if [ "$dry_run" != 1 ] && ! mkdir -p "$log_dir"; then
  echo "tools/self_check.sh: cannot create the log folder '$log_dir'" >&2
  exit 2
fi

fail() {
  failed="${failed:+$failed,}$1"
}

# item <label> <heavy|light> <needed path> <command>...
item() {
  local label="$1" weight="$2" need="$3" log code=0
  shift 3
  # Every cargo item needs a cargo project first.
  if [ "$1" = cargo ] && [ ! -e Cargo.toml ]; then need=Cargo.toml; fi
  if [ ! -e "$need" ]; then
    echo "self-check: $label skip (no $need)"
    return 0
  fi
  if [ "$dry_run" = 1 ]; then
    echo "self-check: $label dry-run cmd=$*"
    return 0
  fi
  log="$log_dir/$label.log"
  if [ "$weight" = heavy ] && [ "$use_slot" = 1 ]; then
    moai heavy-slot-acquire --max "$slot_max" --min-free-gb "$slot_min_free" \
      --name "$slot_name" -- "$@" >"$log" 2>&1 || code=$?
  else
    "$@" >"$log" 2>&1 || code=$?
  fi
  echo "self-check: $label exit=$code log=$log cmd=$*"
  [ "$code" -eq 0 ] || fail "$label"
}

# No file under skills/moai may be larger than the bound; a larger one stops install, update
# and package-inspect.
package_size() {
  local label=package-size need=skills/moai/SKILL.md log code=0 largest over
  if [ ! -e "$need" ]; then
    echo "self-check: $label skip (no $need)"
    return 0
  fi
  if [ "$dry_run" = 1 ]; then
    echo "self-check: $label dry-run limit=$limit"
    return 0
  fi
  log="$log_dir/$label.log"
  if ! find skills/moai -type f -size +"$limit"c >"$log" 2>&1; then code=1; fi
  over="$(grep -c . "$log")"
  [ "$over" -eq 0 ] || code=1
  largest="$(find skills/moai -type f -exec wc -c {} + 2>/dev/null | grep -v ' total$' \
    | sort -n | tail -n 1 | sed 's/^ *//')"
  echo "self-check: $label exit=$code limit=$limit largest=${largest:-none} over=$over log=$log"
  [ "$code" -eq 0 ] || fail "$label"
}

item bin-moai heavy Cargo.toml cargo test --offline --locked --bin moai
item bin-moai-cli heavy Cargo.toml cargo test --offline --locked --bin moai-cli
item clippy heavy Cargo.toml cargo clippy --offline --locked --all-targets -- -D warnings
item fmt light Cargo.toml cargo fmt --check
item git-env-scrub heavy tests/git_env_scrub.rs cargo test --offline --locked --test git_env_scrub
item home-isolation heavy tests/home_isolation_scan.rs \
  cargo test --offline --locked --test home_isolation_scan
item ac-command-lint heavy tests/ac_command_lint.rs \
  cargo test --offline --locked --test ac_command_lint
item package-inspect light skills/moai/SKILL.md \
  cargo run --offline --locked --quiet --bin moai -- package-inspect skills/moai
package_size
item manifest light tools/source_manifest.py python3 -I tools/source_manifest.py check

result=pass
[ -z "$failed" ] || result=fail
[ "$dry_run" != 1 ] || result=fail
flags=""
if [ -n "$(git status --porcelain --untracked-files=normal 2>/dev/null)" ]; then
  flags="dirty=1 "
fi
[ "$dry_run" != 1 ] || flags="${flags}dry-run=1 "
echo "self-check: head=${head} result=${result} ${flags}failed=${failed:-none} logs=${log_dir}"
[ "$result" = pass ]
