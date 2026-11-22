#!/usr/bin/env bash
# The machine items of the pre-PR self-check (card C-77,
# skills/moai/references/internal/self-check.md) from one command; tests/self_check_test.py
# pins the item list.
#
#   tools/self_check.sh                       run every item in this working folder
#   tools/self_check.sh --full-python         the same, with CI's whole Python unittest discover
#   tools/self_check.sh --pre-ready           only the whole-tree checks a head passes before READY
#   SELF_CHECK_DRY_RUN=1 tools/self_check.sh  print the items and the last line, run nothing
#
# Items, in order: bin-moai, bin-moai-cli (whole `cargo test --bin`, never filtered: the help
# rules live there), clippy, fmt, fmt-ops (`rustfmt` on src/ops/*.rs: build.rs generates their
# module list, so `cargo fmt` never sees them), git-env-scrub, home-isolation,
# ac-command-lint, package-inspect, package-size (no file under skills/moai over 65536 bytes,
# the LIMIT of src/profile.rs that install, update and package-inspect read each package file
# through), manifest (the source manifest check), diff-check (`git diff --check` from the merge
# base to the working folder, as CI's diff-check step), python (the unittest files the change
# touched: a changed tests/*_test.py, and tests/<name>_test.py of a changed tools/<name>.py or
# .sh, plus tests/skill_references_test.py; with --full-python, CI's
# `python3 -m unittest discover -s tests -p '*_test.py'` instead).
#
# --pre-ready runs only these items, fast ones first, the whole-tree checks that held pull
# requests back before READY: diff-check, fmt, fmt-ops, manifest, spec-check (`spec check
# --closure` of the built moai-cli for each SPEC the change touched; a draft SPEC's
# `closure_status` alone passes), py-source-manifest (`tests.source_manifest_test`),
# py-repo-scans (`tests.repo_scans_test`) and skill-package (`cargo test --test product_cli
# skill_package`). The two items that read the built moai-cli (`MOAI_CLI_BIN`, else
# `${CARGO_TARGET_DIR:-target}/debug/moai-cli`) are not_run while it is missing or older than
# `src/`, `build.rs`, `Cargo.toml` or `Cargo.lock`: build it first with `cargo build --offline
# --locked --bin moai-cli`. It cannot be combined with --full-python. A Windows type check is
# not an item: see docs/ci-rc-ci-spec.md.
#
# Whole Python runs (--full-python and the py-* items) run in a python slot of their own, a
# slot folder `MOAI_PYTHON_SLOT_DIR` (default `${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/
# python-slots`) with at most `MOAI_PYTHON_SLOT_MAX` (default 2) runs at a time, so a 38-minute
# discover never holds one of the cargo slots.
#
# The base of diff-check and python is SELF_CHECK_BASE (default: the local branch v2026.11.0,
# not its remote copy, which stays behind until someone pushes); the change
# is everything from its merge base with HEAD to the working folder, untracked files included.
# A base that does not resolve makes diff-check and python not_run (with --pre-ready:
# diff-check and spec-check).
#
# It is not the merge evidence: tools/local_check.sh checks the committed head merged with the
# base tip and prints the line the merge gate reads. This one runs in place, before or after a
# commit, keeps going past a failure and prints one line per item:
#   self-check: <item> exit=<code> log=<file> cmd=<command>
#   self-check: <item> skip (no <path>)               the item does not apply here
#   self-check: <item> not_run (<reason>)             the item applies but could not run
#   self-check: <item> exit=75 slot-refused (not run) log=<file>
# The last form is the heavy slot refusing to start (exit 75, too little disk or too much
# swap): the item did not run, so it is not_run, not a failure. package-size prints
# `exit=<0|1> limit=65536 largest=<bytes> <path> over=<n> log=<file>`; the files over the
# bound are listed in its log. Each item's output goes to its log (SELF_CHECK_LOG_DIR, default
# target/self-check, relative to the repository root). The last line is:
#   self-check: head=<40-hex|unknown> result=<pass|fail|not_run> [dirty=1 ][dry-run=1 ]failed=<items|none> not_run=<items|none> logs=<dir>
# result is fail when any item failed, else not_run when any item could not run, else pass.
# dirty=1 says the folder held uncommitted changes, so the run checked more than the head; it
# does not fail the run. A dry run checked nothing and says result=fail. Exit 0 only for
# result=pass, 1 otherwise, 2 for a usage error.
#
# MOAI_HEAVY_SLOT_MAX (default 3), MOAI_HEAVY_SLOT_MIN_FREE_GB (default 50) and
# MOAI_HEAVY_SLOT_NAME (default self-check-<worktree folder>-<pid>, one name per run so runs do
# not share a queue name): when `moai heavy-slot-acquire` exists, the heavy cargo items run in
# one of the machine's heavy slots, as in tools/local_check.sh.
set -u
# Card C-82: incremental output is off for every build here; CARGO_TARGET_DIR is the caller's,
# normally the worktree's build folder from `tools/build_folder.py env`.
export CARGO_INCREMENTAL=0

full_python=0
pre_ready=0
for arg in "$@"; do
  case "$arg" in
    --full-python) full_python=1 ;;
    --pre-ready) pre_ready=1 ;;
    *)
      echo "usage: tools/self_check.sh [--full-python | --pre-ready] (SELF_CHECK_DRY_RUN=1 lists the items)" >&2
      exit 2
      ;;
  esac
done
if [ "$full_python" = 1 ] && [ "$pre_ready" = 1 ]; then
  echo "usage: tools/self_check.sh [--full-python | --pre-ready] (SELF_CHECK_DRY_RUN=1 lists the items)" >&2
  exit 2
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root" || exit 2

limit=65536 # src/profile.rs LIMIT
slot_refused=75 # src/heavy_slot.rs BUSY
dry_run="${SELF_CHECK_DRY_RUN:-0}"
log_dir="${SELF_CHECK_LOG_DIR:-target/self-check}"
slot_max="${MOAI_HEAVY_SLOT_MAX:-3}"
slot_min_free="${MOAI_HEAVY_SLOT_MIN_FREE_GB:-50}"
folder="$(printf '%s' "$(basename "$root")" | tr -c 'A-Za-z0-9._-' '_')"
slot_name="${MOAI_HEAVY_SLOT_NAME:-self-check-${folder}-$$}"
# Whole Python runs (a discover over every test file, 38 minutes under load) wait in a slot
# folder of their own, so one of them never holds one of the cargo slots (SPEC-20261119-python-slot-pre-ready).
py_slot_max="${MOAI_PYTHON_SLOT_MAX:-2}"
py_slot_dir="${MOAI_PYTHON_SLOT_DIR:-}"
if [ -z "$py_slot_dir" ]; then
  if [ -n "${XDG_CACHE_HOME:-}" ]; then
    py_slot_dir="$XDG_CACHE_HOME/moai-cli/python-slots"
  elif [ -n "${HOME:-}" ]; then
    py_slot_dir="$HOME/.cache/moai-cli/python-slots"
  fi
fi
base_ref="${SELF_CHECK_BASE:-v2026.11.0}"
head="$(git rev-parse --verify -q 'HEAD^{commit}' 2>/dev/null || echo unknown)"
failed=""
not_run=""

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

# The merge base of the base and HEAD, or empty when either does not resolve.
merge_base=""
case "$base_ref" in
  "" | -*) ;;
  *) merge_base="$(git merge-base "$base_ref" HEAD 2>/dev/null)" || merge_base="" ;;
esac

fail() {
  failed="${failed:+$failed,}$1"
}

mark_not_run() {
  not_run="${not_run:+$not_run,}$1"
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
  if [ "$weight" = python ] && [ "$use_slot" = 1 ] && [ -z "$py_slot_dir" ]; then
    # No folder can be named (no MOAI_PYTHON_SLOT_DIR, XDG_CACHE_HOME or HOME): never fall back
    # to the cargo slot folder; say so and run directly.
    echo "tools/self_check.sh: no python slot folder (MOAI_PYTHON_SLOT_DIR, XDG_CACHE_HOME and HOME are all empty); $label runs without a slot" >&2
    "$@" >"$log" 2>&1 || code=$?
  elif [ "$use_slot" = 1 ] && { [ "$weight" = heavy ] || [ "$weight" = python ]; }; then
    if [ "$weight" = python ]; then
      moai heavy-slot-acquire --dir "$py_slot_dir" --max "$py_slot_max" \
        --min-free-gb "$slot_min_free" --name "${slot_name}-py" -- "$@" >"$log" 2>&1 || code=$?
    else
      moai heavy-slot-acquire --max "$slot_max" --min-free-gb "$slot_min_free" \
        --name "$slot_name" -- "$@" >"$log" 2>&1 || code=$?
    fi
    if [ "$code" -eq "$slot_refused" ]; then
      echo "self-check: $label exit=$code slot-refused (not run) log=$log"
      mark_not_run "$label"
      return 0
    fi
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

# The operation files are not in `cargo fmt`'s module tree.
fmt_ops() {
  set -- src/ops/*.rs
  if [ ! -e "$1" ]; then
    echo "self-check: fmt-ops skip (no src/ops/*.rs)"
    return 0
  fi
  item fmt-ops light "$1" rustfmt --check --edition 2021 "$@"
}

no_base() {
  echo "self-check: $1 not_run (no base: '$base_ref' has no merge base with HEAD; set SELF_CHECK_BASE)"
  mark_not_run "$1"
}

diff_check() {
  if [ -z "$merge_base" ]; then
    no_base diff-check
    return 0
  fi
  item diff-check light . git diff --check "$merge_base"
}

# The unittest files of the change, one per line, sorted: changed tests/*_test.py and the
# tests/<name>_test.py of a changed tools/<name>.py or tools/<name>.sh, when the file exists.
touched_python_tests() {
  { git diff --name-only "$merge_base" -- 2>/dev/null
    git ls-files --others --exclude-standard 2>/dev/null; } | while IFS= read -r p; do
    case "$p" in
      tests/*_test.py) t="$p" ;;
      tools/*.py | tools/*.sh)
        t="${p#tools/}"
        t="tests/${t%.*}_test.py"
        ;;
      *) continue ;;
    esac
    [ -f "$t" ] && printf '%s\n' "$t"
  done | sort -u
}

python_item() {
  local list t
  if [ "$full_python" = 1 ]; then
    item python python tests python3 -m unittest discover -s tests -p '*_test.py'
    return 0
  fi
  if [ -z "$merge_base" ]; then
    no_base python
    return 0
  fi
  list=""
  [ -f tests/skill_references_test.py ] && list="tests/skill_references_test.py"
  for t in $(touched_python_tests); do
    [ "$t" = tests/skill_references_test.py ] || list="${list:+$list }$t"
  done
  if [ -z "$list" ]; then
    echo "self-check: python skip (no changed Python test and no tests/skill_references_test.py)"
    return 0
  fi
  # shellcheck disable=SC2086 # test paths hold no spaces
  set -- $list
  item python light "$1" python3 -m unittest "$@"
}

# The SPEC folders the change touched, one id per line, sorted: SPEC-*/spec.md changed from the
# merge base to the working folder (untracked included), when the file still exists.
touched_specs() {
  { git diff --name-only "$merge_base" -- 2>/dev/null
    git ls-files --others --exclude-standard 2>/dev/null; } | while IFS= read -r p; do
    case "$p" in
      .moai/specs/SPEC-*/spec.md) [ -f "$p" ] && basename "$(dirname "$p")" ;;
    esac
  done | sort -u
}

# needs_bin <label>: sets cli_bin to the built moai-cli and returns 0, or prints a not_run line
# and returns 1 when it is missing or older than the sources it was built from. A dry run
# only names the path.
cli_bin=""
needs_bin() {
  local label="$1" newer
  cli_bin="${MOAI_CLI_BIN:-${CARGO_TARGET_DIR:-target}/debug/moai-cli}"
  [ "$dry_run" != 1 ] || return 0
  if [ ! -x "$cli_bin" ]; then
    echo "self-check: $label not_run (no built moai-cli: $cli_bin)"
    mark_not_run "$label"
    return 1
  fi
  newer="$(find src build.rs Cargo.toml Cargo.lock -type f -newer "$cli_bin" 2>/dev/null | head -n 1)"
  if [ -n "$newer" ]; then
    echo "self-check: $label not_run (stale moai-cli: $cli_bin is older than $newer)"
    mark_not_run "$label"
    return 1
  fi
  return 0
}

# spec-check: `spec check --closure` of the built moai-cli for each SPEC the change touched.
# A draft SPEC always answers `closure_status` (approval is the user's), so that code alone
# passes; any other code (a requirement that does not parse, a missing `## Result`) fails.
spec_check() {
  local label=spec-check ids id log code=0 out codes
  if [ -z "$merge_base" ]; then
    no_base "$label"
    return 0
  fi
  ids="$(touched_specs)"
  if [ -z "$ids" ]; then
    echo "self-check: $label skip (no changed SPEC)"
    return 0
  fi
  needs_bin "$label" || return 0
  if [ "$dry_run" = 1 ]; then
    echo "self-check: $label dry-run cmd=$cli_bin spec check <id> --closure --project . ($(echo $ids))"
    return 0
  fi
  log="$log_dir/$label.log"
  : >"$log"
  for id in $ids; do
    out="$("$cli_bin" spec check "$id" --closure --project . 2>&1)" && continue
    printf '%s\n' "$out" >>"$log"
    codes="$(printf '%s\n' "$out" | grep -oE 'spec\.md(:[0-9]+)? [a-z_]+' | awk '{print $2}' | sort -u)"
    if [ -z "$codes" ] || [ -n "$(printf '%s\n' "$codes" | grep -v '^closure_status$')" ]; then
      code=1
    fi
  done
  echo "self-check: $label exit=$code log=$log cmd=$cli_bin spec check <id> --closure --project . ($(echo $ids))"
  [ "$code" -eq 0 ] || fail "$label"
}

# A whole-file Python test item that reads the built moai-cli.
bin_python_item() {
  local label="$1" file="$2" module="$3"
  if [ ! -e "$file" ]; then
    echo "self-check: $label skip (no $file)"
    return 0
  fi
  needs_bin "$label" || return 0
  item "$label" python "$file" python3 -m unittest "$module"
}

# --pre-ready: the whole-tree checks that held pull requests back before READY, fast ones
# first. The cargo build the two binary items need is the caller's (`cargo build --bin moai-cli`).
if [ "$pre_ready" = 1 ]; then
  diff_check
  item fmt light Cargo.toml cargo fmt --check
  fmt_ops
  item manifest light tools/source_manifest.py python3 -I tools/source_manifest.py check
  spec_check
  item py-source-manifest python tests/source_manifest_test.py \
    python3 -m unittest tests.source_manifest_test
  bin_python_item py-repo-scans tests/repo_scans_test.py tests.repo_scans_test
  item skill-package heavy tests/product_cli.rs \
    cargo test --offline --locked --test product_cli skill_package
else
  item bin-moai heavy Cargo.toml cargo test --offline --locked --bin moai
  item bin-moai-cli heavy Cargo.toml cargo test --offline --locked --bin moai-cli
  item clippy heavy Cargo.toml cargo clippy --offline --locked --all-targets -- -D warnings
  item fmt light Cargo.toml cargo fmt --check
  fmt_ops
  item git-env-scrub heavy tests/git_env_scrub.rs cargo test --offline --locked --test git_env_scrub
  item home-isolation heavy tests/home_isolation_scan.rs \
    cargo test --offline --locked --test home_isolation_scan
  item ac-command-lint heavy tests/ac_command_lint.rs \
    cargo test --offline --locked --test ac_command_lint
  item package-inspect light skills/moai/SKILL.md \
    cargo run --offline --locked --quiet --bin moai -- package-inspect skills/moai
  package_size
  item manifest light tools/source_manifest.py python3 -I tools/source_manifest.py check
  diff_check
  python_item
fi

result=pass
[ -z "$not_run" ] || result=not_run
[ -z "$failed" ] || result=fail
[ "$dry_run" != 1 ] || result=fail
flags=""
if [ -n "$(git status --porcelain --untracked-files=normal 2>/dev/null)" ]; then
  flags="dirty=1 "
fi
[ "$dry_run" != 1 ] || flags="${flags}dry-run=1 "
echo "self-check: head=${head} result=${result} ${flags}failed=${failed:-none} not_run=${not_run:-none} logs=${log_dir}"
[ "$result" = pass ]
