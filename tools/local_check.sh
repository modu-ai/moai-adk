#!/usr/bin/env bash
# The local check set (#2906 S5a, AC20): the check steps of the CI `check` job
# (.github/workflows/ci.yml, "Check set"), in the same order with the same arguments. The
# CI job runs this file for those steps, so the two cannot drift apart;
# tests/local_check_test.py pins the step list.
#
#   tools/local_check.sh                       run every step, stop at the first failure
#   LOCAL_CHECK_DRY_RUN=1 tools/local_check.sh  print the steps and the final line, run nothing
#
# Steps: fetch, test (cargo-nextest), doctest, clippy, fmt, fmt-ops (rustfmt on src/ops/*.rs,
# which `cargo fmt` does not reach), unittest, diff-check.
# Skipped, because they set up the CI machine rather than check the tree: checkout, the
# docs-only scope step (tools/ci_scope.py; a local run always takes the full set), Python
# 3.12 + PyYAML, ripgrep, the Rust toolchain (rust-toolchain.toml), the dependency cache,
# the cargo-nextest download, and the step-summary header and macOS-only case table that
# the workflow writes itself. Locally these must already be installed.
#
# Two modes (#3066 P2). CI checks the merge of the pull request with its base branch; a
# local run used to check only its own head, so a head that passes alone could fail once
# merged (#3062). The default mode therefore checks a CANDIDATE:
#
#   candidate (default)  the HEAD merged with the latest tip of the base branch. The script
#                        fetches origin's <base branch> (LOCAL_CHECK_BASE_BRANCH, default
#                        v2026.11.0) into refs/remotes/origin/<branch> and pins its commit,
#                        builds the merge with `git merge-tree --write-tree` and
#                        `git commit-tree` (parents: base, head), checks that commit out as a
#                        detached `git worktree` under a `mktemp -d` folder, runs the steps
#                        there, and removes the worktree on exit (`git worktree remove`, never
#                        `rm -rf`). The working folder, its index, its branches and the stash
#                        are not touched; the only ref that changes is
#                        refs/remotes/origin/<branch>. Only the committed HEAD is tested, so
#                        uncommitted changes make the result fail at once. The candidate
#                        folder is new every run, so cargo builds there from scratch unless
#                        CARGO_TARGET_DIR points elsewhere.
#   in-place             LOCAL_CHECK_MODE=in-place (CI only): the steps run in this working
#                        folder, which already is the merge ref, and the short line is printed.
#
# The last line of output is machine-readable, exactly one line. Candidate mode prints the
# long form the gate reads (src/gate.rs `parse_local_check_evidence`), fields in this order:
#   local-check: head=<40hex> result=<pass|fail> base=<40hex|none> tree=<40hex|none> script=<40hex> rustc=<token> run=<12hex> cmd=tools/local_check.sh at=<UTC>
# base is the pinned tip the head was merged onto and tree is the tree of that merge, both
# 40 lowercase hex; `none` appears only on a fail (no base: the fetch or the branch name
# failed, the tree had uncommitted changes, or a dry run; a base without a tree: a merge
# conflict or a merge error). script is `git hash-object` of this file, rustc the version
# word of `rustc --version` read inside the candidate (1-32 of [A-Za-z0-9._+-]; `unknown`
# on a fail when it cannot be read), run the first 12 hex of the sha1 of
# "head base tree at pid". A fetch that fails never falls back to an older base: the line
# is result=fail base=none tree=none. In this mode LOCAL_CHECK_BASE and
# LOCAL_CHECK_DIFF_HEAD are ignored (a stderr notice says so); diff-check reads
# <base>...<head>.
# In-place mode prints the short form:
#   local-check: head=<40-hex HEAD> result=pass|fail [dirty=1 ]cmd=tools/local_check.sh at=<UTC>
# A tree with uncommitted changes (tracked or untracked, not ignored) at the start or the end
# is not the HEAD being checked, so the result is fail, even when every step passed (in
# place the line says dirty=1; the candidate line has no dirty field, because the gate would
# not read a line with one). A dry run checked nothing and always says result=fail. The exit
# code is 0 only for result=pass.
#
# Environment:
#   LOCAL_CHECK_MODE        candidate (default) or in-place (CI only); anything else exits 2
#   LOCAL_CHECK_BASE_BRANCH candidate base branch on origin (default v2026.11.0); only
#                           [A-Za-z0-9._/-], no leading `-`, no `..`
#   LOCAL_CHECK_BASE        in-place only: diff-check range start (default origin/v2026.11.0,
#                           else HEAD~1)
#   LOCAL_CHECK_DIFF_HEAD   in-place only: diff-check range end (default HEAD)
#   LOCAL_CHECK_KEEP_GOING=1  run every step and fail at the end (CI uses it, so each step's
#                           exit is recorded as before); the default stops at the first failure
#   LOCAL_CHECK_LEG         the CI check-matrix leg (SPEC-20261119-ci-check-set-legs): unset or
#                           empty keeps today's single full set; `nextest` runs exactly fetch,
#                           test, doctest; `rest` exactly fetch, clippy, fmt, fmt-ops, unittest,
#                           diff-check; any other value exits 2 before any step runs, and so
#                           does any set value outside CI — candidate mode, or in-place without
#                           GITHUB_ACTIONS=true (the full set is the only local evidence; legs
#                           are CI's in-place parallelism)
#   GITHUB_STEP_SUMMARY     when set, one `| step | command | exit |` row per step is appended
#   MOAI_HEAVY_SLOT_MAX (default 5; the slot command takes only a value below the machine's
#   limits.json), MOAI_HEAVY_SLOT_MIN_FREE_GB (default 50),
#   MOAI_HEAVY_SLOT_NAME (default local-check): when `moai heavy-slot-acquire` exists, the
#                           heavy cargo steps (test, doctest, clippy) each run in one of the
#                           machine's heavy slots (leader decision 2026-11-14 12:37Z; #2930
#                           replaces these with a configured value). Without it they run directly.
#   MOAI_PYTHON_SLOT_MAX (default 2), MOAI_PYTHON_SLOT_DIR (default
#                           ${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/python-slots): the unittest
#                           step runs in a python slot, a slot folder apart from the cargo slots,
#                           so a whole discover never holds a cargo slot. Without `moai
#                           heavy-slot-acquire` it runs directly, as before.
#   CARGO_INCREMENTAL is always 0 here (card C-82: incremental output was 2.9 of a 4.2 GiB test
#                           build); CARGO_TARGET_DIR is the caller's, normally the worktree's
#                           build folder from `tools/build_folder.py env`.
set -eu
export CARGO_INCREMENTAL=0

script_path="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

mode="${LOCAL_CHECK_MODE:-candidate}"
case "$mode" in
  candidate | in-place) ;;
  *)
    echo "tools/local_check.sh: LOCAL_CHECK_MODE must be candidate or in-place, not '$mode'" >&2
    exit 2
    ;;
esac

# SPEC-20261119-ci-check-set-legs (C-20): the CI check matrix runs the set as two parallel
# legs of the same target. Unset or empty is the local default, today's full set; any value
# but the two leg names refuses before a single stage runs (a CI typo must be loud, never a
# silent second full set). The refusal is before the traps, so no result line is printed.
leg="${LOCAL_CHECK_LEG:-}"
case "$leg" in
  "" | nextest | rest) ;;
  *)
    echo "tools/local_check.sh: LOCAL_CHECK_LEG must be unset, 'nextest' or 'rest', not '$leg'" >&2
    exit 2
    ;;
esac

# Leader repair N1 (2026-11-19): a leg narrows the set, so it may never produce the
# candidate-mode evidence line a gate reads as a full-set pass — a one-leg run must not be
# readable as a pass for everything. Slice 4 (audit record 1): in-place mode keeps legs on
# CI only (GITHUB_ACTIONS=true) — a local in-place run also prints a pass line, so a
# one-leg short pass must not appear there either. Refused before the traps: no result line.
if [ -n "$leg" ]; then
  if [ "$mode" = candidate ]; then
    echo "tools/local_check.sh: LOCAL_CHECK_LEG is a CI leg for in-place mode; candidate mode always checks the full set, unset LOCAL_CHECK_LEG" >&2
    exit 2
  fi
  if [ "${GITHUB_ACTIONS:-}" != true ]; then
    echo "tools/local_check.sh: LOCAL_CHECK_LEG is for CI (in-place under GITHUB_ACTIONS); a local run always checks the full set, unset LOCAL_CHECK_LEG" >&2
    exit 2
  fi
fi

dry_run="${LOCAL_CHECK_DRY_RUN:-0}"
keep_going="${LOCAL_CHECK_KEEP_GOING:-0}"
base_branch="${LOCAL_CHECK_BASE_BRANCH:-v2026.11.0}"
remote=origin
head="$(git rev-parse HEAD 2>/dev/null || echo unknown)"

if [ "$mode" = in-place ]; then
  if [ -n "${LOCAL_CHECK_BASE:-}" ]; then
    base="$LOCAL_CHECK_BASE"
  elif git rev-parse --verify -q 'origin/v2026.11.0^{commit}' >/dev/null; then
    base=origin/v2026.11.0
  else
    base=HEAD~1
  fi
  diff_head="${LOCAL_CHECK_DIFF_HEAD:-HEAD}"
  diff_range="${base}...${diff_head}"
  if [ "${GITHUB_ACTIONS:-}" != true ]; then
    echo "tools/local_check.sh: LOCAL_CHECK_MODE=in-place checks only this working folder and prints the short line; it is for CI, a local run should use the default candidate mode" >&2
  fi
else
  if [ -n "${LOCAL_CHECK_BASE:-}" ] || [ -n "${LOCAL_CHECK_DIFF_HEAD:-}" ]; then
    echo "tools/local_check.sh: candidate mode ignores LOCAL_CHECK_BASE and LOCAL_CHECK_DIFF_HEAD; it checks the head merged with the latest ${remote}/${base_branch}" >&2
  fi
  diff_range="<base>...<head>"
fi
slot_max="${MOAI_HEAVY_SLOT_MAX:-5}"
slot_min_free="${MOAI_HEAVY_SLOT_MIN_FREE_GB:-50}"
slot_name="${MOAI_HEAVY_SLOT_NAME:-local-check}"
# The whole Python discover (38 minutes under load) waits in a slot folder of its own, so it
# never holds one of the cargo slots (SPEC-20261119-python-slot-pre-ready).
py_slot_max="${MOAI_PYTHON_SLOT_MAX:-2}"
py_slot_dir="${MOAI_PYTHON_SLOT_DIR:-}"
if [ -z "$py_slot_dir" ]; then
  if [ -n "${XDG_CACHE_HOME:-}" ]; then
    py_slot_dir="$XDG_CACHE_HOME/moai-cli/python-slots"
  elif [ -n "${HOME:-}" ]; then
    py_slot_dir="$HOME/.cache/moai-cli/python-slots"
  fi
fi

use_slot=0
if command -v moai >/dev/null 2>&1 && moai heavy-slot-acquire --help >/dev/null 2>&1; then
  use_slot=1
fi

dirty=0
failed=0
stopped=0
# Candidate-mode evidence: `none` until the run knows the value.
base_sha=none
tree_sha=none
rustc_tok=unknown
tmp_parent=""
cand=""
cand_added=0

tree_is_dirty() {
  [ -n "$(git -C "$root" status --porcelain --untracked-files=normal 2>/dev/null || echo error)" ]
}

valid_branch() {
  case "$1" in
    "" | -* | *..* | *[!A-Za-z0-9._/-]*) return 1 ;;
  esac
  return 0
}

is_hex40() {
  case "$1" in
    *[!0-9a-f]*) return 1 ;;
  esac
  [ "${#1}" -eq 40 ]
}

# The version word of `rustc --version` (`rustc 1.91.0 (...)` -> `1.91.0`), or failure.
read_rustc() {
  local out tok
  out="$(rustc --version 2>/dev/null)" || return 1
  tok="$(printf '%s\n' "$out" | awk 'NR == 1 { print $2 }')"
  if printf '%s' "$tok" | grep -Eq '^[A-Za-z0-9._+-]{1,32}$'; then
    rustc_tok="$tok"
    return 0
  fi
  return 1
}

# Removes the candidate with `git worktree remove` (never `rm -rf`) and its mktemp parent
# with `rmdir`. A leftover is reported and makes the run fail.
cleanup_candidate() {
  [ -n "$tmp_parent" ] || return 0
  cd "$root" 2>/dev/null || true
  if [ "$cand_added" = 1 ]; then
    case "$cand" in
      "$tmp_parent"/*) ;;
      *)
        echo "tools/local_check.sh: refusing to remove '$cand': it is not under '$tmp_parent'" >&2
        failed=1
        return 0
        ;;
    esac
    if git -C "$root" worktree remove --force "$cand" >/dev/null 2>&1; then
      cand_added=0
    else
      echo "tools/local_check.sh: could not remove the candidate worktree '$cand'; remove it with 'git worktree remove --force' and then 'rmdir' its folder" >&2
      failed=1
      return 0
    fi
  fi
  if rmdir "$tmp_parent" 2>/dev/null; then
    tmp_parent=""
  else
    echo "tools/local_check.sh: could not remove the temporary folder '$tmp_parent'" >&2
    failed=1
  fi
  return 0
}

finish() {
  cleanup_candidate
  if tree_is_dirty; then dirty=1; fi
  # A commit made while the checks ran means the checked tree is not the reported head
  # (#2951 audit N1): fail, and keep the head the run started from.
  if [ "$(git -C "$root" rev-parse HEAD 2>/dev/null || echo unknown)" != "$head" ]; then failed=1; fi
  result=pass
  if [ "$failed" -ne 0 ] || [ "$dirty" -ne 0 ] || [ "$dry_run" = 1 ]; then result=fail; fi
  at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  if [ "$mode" = in-place ]; then
    flag=""
    if [ "$dirty" -ne 0 ]; then flag="dirty=1 "; fi
    echo "local-check: head=${head} result=${result} ${flag}cmd=tools/local_check.sh at=${at}"
  else
    # A pass needs a real base, a real tree and a readable toolchain; never print one without.
    if [ "$result" = pass ]; then
      if [ "$base_sha" = none ] || [ "$tree_sha" = none ] || [ "$rustc_tok" = unknown ]; then
        result=fail
      fi
    fi
    script_hash="$(git -C "$root" hash-object -- "$script_path" 2>/dev/null || echo none)"
    run_id="$(printf '%s %s %s %s %s' "$head" "$base_sha" "$tree_sha" "$at" "$$" \
      | git -C "$root" hash-object --stdin 2>/dev/null | cut -c1-12 || true)"
    if ! is_hex40 "$script_hash" || [ "${#run_id}" -ne 12 ]; then
      echo "tools/local_check.sh: could not compute the script or run id" >&2
      script_hash=0000000000000000000000000000000000000000
      run_id=000000000000
      result=fail
    fi
    echo "local-check: head=${head} result=${result} base=${base_sha} tree=${tree_sha} script=${script_hash} rustc=${rustc_tok} run=${run_id} cmd=tools/local_check.sh at=${at}"
  fi
  if [ "$result" = pass ]; then exit 0; fi
  exit 1
}
# Any unexpected abort still ends with the line, as a fail.
trap 'failed=1; finish' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# Candidate mode: fetch the base tip, merge it with the head in the object database only,
# check the merge out as a detached worktree and enter it. Returns 1 when it cannot; the
# evidence variables then tell what is known. This runs in an `if`, where `set -e` is off,
# so every command is checked by hand.
candidate_setup() {
  local mt_out mt_code cand_commit leftover
  if ! valid_branch "$base_branch"; then
    echo "tools/local_check.sh: LOCAL_CHECK_BASE_BRANCH '$base_branch' is not a branch name this script accepts ([A-Za-z0-9._/-], no leading '-', no '..')" >&2
    return 1
  fi
  if [ "$head" = unknown ]; then
    echo "tools/local_check.sh: cannot read HEAD" >&2
    return 1
  fi
  # The same refspec style as `moai gate-loop` (src/gate_loop.rs `fetch`): FETCH_HEAD is not
  # written, nothing else is fetched, and the tip is pinned to its SHA at once.
  if ! GIT_TERMINAL_PROMPT=0 git -c gc.auto=0 -c maintenance.auto=false fetch --quiet \
    --no-write-fetch-head --no-tags "$remote" \
    "+refs/heads/${base_branch}:refs/remotes/${remote}/${base_branch}" </dev/null; then
    echo "tools/local_check.sh: fetching ${remote}/${base_branch} failed; the base would be stale, so the result is fail" >&2
    return 1
  fi
  base_sha="$(git rev-parse --verify -q "refs/remotes/${remote}/${base_branch}^{commit}")" || base_sha=none
  if ! is_hex40 "$base_sha"; then
    echo "tools/local_check.sh: ${remote}/${base_branch} is not a commit after the fetch" >&2
    base_sha=none
    return 1
  fi
  mt_code=0
  mt_out="$(git merge-tree --write-tree "$base_sha" "$head")" || mt_code=$?
  if [ "$mt_code" -eq 1 ]; then
    echo "tools/local_check.sh: the head conflicts with ${remote}/${base_branch} (${base_sha}); merge-tree says:" >&2
    printf '%s\n' "$mt_out" | sed 1d >&2
    return 1
  elif [ "$mt_code" -ne 0 ]; then
    echo "tools/local_check.sh: git merge-tree failed (exit ${mt_code})" >&2
    return 1
  fi
  tree_sha="$(printf '%s\n' "$mt_out" | head -n 1)"
  if ! is_hex40 "$tree_sha"; then
    echo "tools/local_check.sh: git merge-tree did not print a tree" >&2
    tree_sha=none
    return 1
  fi
  if ! cand_commit="$(GIT_AUTHOR_NAME=local-check GIT_AUTHOR_EMAIL=local-check@invalid \
    GIT_COMMITTER_NAME=local-check GIT_COMMITTER_EMAIL=local-check@invalid \
    git -c commit.gpgSign=false commit-tree "$tree_sha" -p "$base_sha" -p "$head" \
    -m "local-check candidate: ${head} on ${base_sha}")" || ! is_hex40 "$cand_commit"; then
    echo "tools/local_check.sh: could not create the candidate commit" >&2
    tree_sha=none
    return 1
  fi
  # Folders a killed run left behind are only reported: another run may be using one.
  for leftover in "${TMPDIR:-/tmp}"/local-check-cand.*; do
    if [ -d "$leftover" ]; then
      echo "tools/local_check.sh: an earlier candidate folder exists: '$leftover' (not deleted; remove it yourself if no run uses it)" >&2
    fi
  done
  if ! tmp_parent="$(mktemp -d "${TMPDIR:-/tmp}/local-check-cand.XXXXXX")"; then
    tmp_parent=""
    echo "tools/local_check.sh: could not create a temporary folder" >&2
    tree_sha=none
    return 1
  fi
  cand="${tmp_parent}/candidate"
  if ! git worktree add --detach --quiet "$cand" "$cand_commit" >/dev/null 2>&1; then
    echo "tools/local_check.sh: could not check the candidate out into '$cand'" >&2
    tree_sha=none
    return 1
  fi
  cand_added=1
  cd "$cand" || return 1
  echo "candidate: head ${head} merged onto ${remote}/${base_branch} ${base_sha}: tree ${tree_sha}, commit ${cand_commit}, folder ${cand}"
  if ! read_rustc; then
    echo "tools/local_check.sh: cannot read 'rustc --version' inside the candidate; the result is fail" >&2
    return 1
  fi
  diff_range="${base_sha}...${head}"
  return 0
}

if tree_is_dirty; then
  dirty=1
  if [ "$mode" = in-place ]; then
    echo "tools/local_check.sh: the tree has uncommitted changes; the result will be fail (dirty=1)" >&2
  else
    echo "tools/local_check.sh: the tree has uncommitted changes and candidate mode checks only the committed HEAD; commit or stash them first. The result is fail." >&2
    stopped=1
  fi
fi

if [ "$mode" = candidate ] && [ "$stopped" = 0 ]; then
  if [ "$dry_run" = 1 ]; then
    read_rustc || true
  elif ! candidate_setup; then
    failed=1
    stopped=1
  fi
fi

# step heavy|python|light <label> <command>...
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
  elif [ "$weight" = python ] && [ "$use_slot" = 1 ] && [ -z "$py_slot_dir" ]; then
    # No folder can be named: never fall back to the cargo slot folder; say so and run directly.
    echo "tools/local_check.sh: no python slot folder (MOAI_PYTHON_SLOT_DIR, XDG_CACHE_HOME and HOME are all empty); step $label runs without a slot" >&2
    "$@" || code=$?
  elif [ "$weight" = python ] && [ "$use_slot" = 1 ]; then
    moai heavy-slot-acquire --dir "$py_slot_dir" --max "$py_slot_max" \
      --min-free-gb "$slot_min_free" --name "${slot_name}-py" -- "$@" || code=$?
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

# SPEC-20261119-ci-check-set-legs: which steps this run owns. `on_leg <list>` answers
# whether the selected leg is in the list; the unset leg (the local default full set)
# matches every list, so today's order and stage set are exactly today's for it.
on_leg() {
  [ -z "$leg" ] && return 0
  case " $1 " in
    *" $leg "*) return 0 ;;
    *) return 1 ;;
  esac
}

# Fetch once against Cargo.lock, then run the set offline. Both legs fetch: each CI leg is
# its own job with its own target folder, and a local run always wants the fetch first.
if on_leg "nextest rest"; then
  step light fetch cargo fetch --locked
fi
# #2452: test threads get 8 MiB (debug `Cli::command()` needs about 2000 KiB of the 2 MiB
# default). #2554 2-1: the test binaries run in parallel under cargo-nextest; nextest does
# not run doc-tests, so `cargo test --doc` runs them right after. The logs (ignored by
# .gitignore `*.log`) feed the workflow's macOS-only case table.
if on_leg "nextest"; then
  step heavy test bash -c 'RUST_MIN_STACK=8388608 cargo nextest run --offline --locked --no-fail-fast 2>&1 | tee ci-test.log; exit "${PIPESTATUS[0]}"'
fi
if on_leg "nextest"; then
  step heavy doctest bash -c 'RUST_MIN_STACK=8388608 cargo test --offline --locked --doc 2>&1 | tee ci-doctest.log; exit "${PIPESTATUS[0]}"'
fi
if on_leg "rest"; then
  step heavy clippy cargo clippy --offline --locked --all-targets -- -D warnings
fi
if on_leg "rest"; then
  step light fmt cargo fmt --check
fi
# C-75 S0: build.rs includes src/ops/*.rs through `#[path]` in a file it writes into
# OUT_DIR, so `cargo fmt` never reaches them; rustfmt checks them here with the crate's
# edition (Cargo.toml). With no file the pattern reaches rustfmt as is, and that fails.
if on_leg "rest"; then
  step light fmt-ops bash -c 'rustfmt --check --edition 2021 src/ops/*.rs'
fi
if on_leg "rest"; then
  step python unittest bash -c "python3 tools/unittest_parallel.py -v -s tests -p '*_test.py' 2>&1 | tee ci-unittest.log; exit \"\${PIPESTATUS[0]}\""
fi
if on_leg "rest"; then
  step light diff-check git diff --check "${diff_range}"
fi

trap - EXIT
finish
