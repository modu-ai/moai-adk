#!/usr/bin/env bash
# The local check set (#2906 S5a, AC20): the check steps of the CI `check` job
# (.github/workflows/ci.yml, "Check set"), in the same order with the same arguments. The
# CI job runs this file for those steps, so the two cannot drift apart;
# tests/local_check_test.py pins the step list.
#
#   tools/local_check.sh                       run every step, stop at the first failure
#   LOCAL_CHECK_DRY_RUN=1 tools/local_check.sh  print the steps and the final line, run nothing
#
# Steps: fetch, test (cargo-nextest), doctest, clippy, fmt, unittest, diff-check.
# Skipped, because they set up the CI machine rather than check the tree: checkout, the
# docs-only scope step (tools/ci_scope.py; a local run always takes the full set), Python
# 3.12 + PyYAML, ripgrep, the Rust toolchain (rust-toolchain.toml), the dependency cache,
# the cargo-nextest download, and the step-summary header and macOS-only case table that
# the workflow writes itself. Locally these must already be installed.
#
# The last line of output is machine-readable, exactly one line:
#   local-check: head=<40-hex HEAD> result=pass|fail [dirty=1 ]cmd=tools/local_check.sh at=<UTC>
# A tree with uncommitted changes (tracked or untracked, not ignored) at the start or the end
# is not the HEAD being checked, so the result is fail with dirty=1, even when every step
# passed. A dry run checked nothing and always says result=fail. The exit code is 0 only
# for result=pass.
#
# Environment:
#   LOCAL_CHECK_BASE        diff-check range start (default origin/v2026.11.0, else HEAD~1)
#   LOCAL_CHECK_DIFF_HEAD   diff-check range end (default HEAD)
#   LOCAL_CHECK_KEEP_GOING=1  run every step and fail at the end (CI uses it, so each step's
#                           exit is recorded as before); the default stops at the first failure
#   GITHUB_STEP_SUMMARY     when set, one `| step | command | exit |` row per step is appended
#   MOAI_HEAVY_SLOT_MAX (default 3), MOAI_HEAVY_SLOT_MIN_FREE_GB (default 50),
#   MOAI_HEAVY_SLOT_NAME (default local-check): when `moai heavy-slot-acquire` exists, the
#                           heavy cargo steps (test, doctest, clippy) each run in one of the
#                           machine's heavy slots (leader decision 2026-11-14 12:37Z; #2930
#                           replaces these with a configured value). Without it they run directly.
set -eu

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

dry_run="${LOCAL_CHECK_DRY_RUN:-0}"
keep_going="${LOCAL_CHECK_KEEP_GOING:-0}"
if [ -n "${LOCAL_CHECK_BASE:-}" ]; then
  base="$LOCAL_CHECK_BASE"
elif git rev-parse --verify -q 'origin/v2026.11.0^{commit}' >/dev/null; then
  base=origin/v2026.11.0
else
  base=HEAD~1
fi
diff_head="${LOCAL_CHECK_DIFF_HEAD:-HEAD}"
slot_max="${MOAI_HEAVY_SLOT_MAX:-3}"
slot_min_free="${MOAI_HEAVY_SLOT_MIN_FREE_GB:-50}"
slot_name="${MOAI_HEAVY_SLOT_NAME:-local-check}"

use_slot=0
if command -v moai >/dev/null 2>&1 && moai heavy-slot-acquire --help >/dev/null 2>&1; then
  use_slot=1
fi

head="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
dirty=0
failed=0
stopped=0

tree_is_dirty() {
  [ -n "$(git status --porcelain --untracked-files=normal 2>/dev/null || echo error)" ]
}

finish() {
  if tree_is_dirty; then dirty=1; fi
  # A commit made while the checks ran means the checked tree is not the reported head
  # (#2951 audit N1): fail, and keep the head the run started from.
  if [ "$(git rev-parse HEAD 2>/dev/null || echo unknown)" != "$head" ]; then failed=1; fi
  result=pass
  if [ "$failed" -ne 0 ] || [ "$dirty" -ne 0 ] || [ "$dry_run" = 1 ]; then result=fail; fi
  flag=""
  if [ "$dirty" -ne 0 ]; then flag="dirty=1 "; fi
  echo "local-check: head=${head} result=${result} ${flag}cmd=tools/local_check.sh at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  if [ "$result" = pass ]; then exit 0; fi
  exit 1
}
# Any unexpected abort still ends with the line, as a fail.
trap 'failed=1; finish' EXIT

if tree_is_dirty; then
  dirty=1
  echo "tools/local_check.sh: the tree has uncommitted changes; the result will be fail (dirty=1)" >&2
fi

# step heavy|light <label> <command>...
step() {
  local weight="$1" label="$2"
  shift 2
  [ "$stopped" = 0 ] || return 0
  echo "step $label: $*"
  [ "$dry_run" = 1 ] && return 0
  local code=0
  if [ "$weight" = heavy ] && [ "$use_slot" = 1 ]; then
    moai heavy-slot-acquire --max "$slot_max" --min-free-gb "$slot_min_free" \
      --name "$slot_name" -- "$@" || code=$?
  else
    "$@" || code=$?
  fi
  if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    echo "| $label | \`$*\` | $code |" >> "$GITHUB_STEP_SUMMARY"
  fi
  if [ "$code" -ne 0 ]; then
    echo "step $label: exit $code" >&2
    failed=1
    [ "$keep_going" = 1 ] || stopped=1
  fi
  return 0
}

# Fetch once against Cargo.lock, then run the set offline.
step light fetch cargo fetch --locked
# #2452: test threads get 8 MiB (debug `Cli::command()` needs about 2000 KiB of the 2 MiB
# default). #2554 2-1: the test binaries run in parallel under cargo-nextest; nextest does
# not run doc-tests, so `cargo test --doc` runs them right after. The logs (ignored by
# .gitignore `*.log`) feed the workflow's macOS-only case table.
step heavy test bash -c 'RUST_MIN_STACK=8388608 cargo nextest run --offline --locked --no-fail-fast 2>&1 | tee ci-test.log; exit "${PIPESTATUS[0]}"'
step heavy doctest bash -c 'RUST_MIN_STACK=8388608 cargo test --offline --locked --doc 2>&1 | tee ci-doctest.log; exit "${PIPESTATUS[0]}"'
step heavy clippy cargo clippy --offline --locked --all-targets -- -D warnings
step light fmt cargo fmt --check
step light unittest bash -c "python3 -m unittest discover -v -s tests -p '*_test.py' 2>&1 | tee ci-unittest.log; exit \"\${PIPESTATUS[0]}\""
step light diff-check git diff --check "${base}...${diff_head}"

trap - EXIT
finish
