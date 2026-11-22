#!/bin/bash
# The build-folder reuse run of SPEC-20261122-ops-registry-path-free (AC-4).
#
# Two worktrees of one commit share a build folder the way a warmed slot does: A is built, its
# target/ is cloned into B (APFS `cp -c`, timestamps kept), A is removed, and B is built after a
# change outside src/ops/, so the build script does not run again and the library is compiled
# with the registry that A's build wrote. Before the change that registry declares A's
# operation files by absolute path; after it, the registry names no path.
#
# usage: reuse-run.sh <repository> <commit> <label>
# Prints one STEP line per step with its exit code, the build lines of B, the registry lines
# that name A, and the first help line of one operation from B's binary. Worktrees go under the
# repository's .claude/worktrees/; A is removed by the run, B is left for reading and is removed
# by the caller with `git worktree remove`. macOS only (`cp -c`).
# No compiler cache: a cached compile could hide which sources were compiled.
set -u
repo=$1
commit=$2
label=$3
a="$repo/.claude/worktrees/reuse-$label-a"
b="$repo/.claude/worktrees/reuse-$label-b"
export CARGO_BUILD_JOBS=5 CARGO_INCREMENTAL=0
unset RUSTC_WRAPPER CARGO_TARGET_DIR CLAUDE_CODE_SESSION_ID

build() {
    (cd "$1" && CARGO_TARGET_DIR="$1/target" moai heavy-slot-acquire --min-free-gb 50 \
        --name "reuse-$label" -- cargo build --offline --locked --bin moai-cli)
}

# Both worktrees exist before A is built, so B's files are not newer than A's build.
git -C "$repo" worktree add -q --detach "$a" "$commit" </dev/null
echo "STEP worktree-a exit=$?"
git -C "$repo" worktree add -q --detach "$b" "$commit" </dev/null
echo "STEP worktree-b exit=$?"
build "$a" </dev/null
echo "STEP build-a exit=$?"
cp -c -R -p "$a/target" "$b/target"
echo "STEP clone-target exit=$?"
git -C "$repo" worktree remove "$a" </dev/null
echo "STEP remove-a exit=$?"
test -e "$a"
echo "STEP a-still-there exit=$? (1 means gone)"
# A change outside src/ops/: the library compiles again, the build script stays fresh.
touch "$b/src/main.rs"
build "$b" </dev/null 2>&1 | grep -E "^error|file not found|Finished|Compiling moai-cli" | head -6
echo "STEP build-b exit=${PIPESTATUS[0]}"
registry=$(find "$b/target/debug/build" -path '*/out/ops_registry.rs' -newer "$b/Cargo.toml" | head -1)
echo "REGISTRY $registry"
echo "A_PATH_LINES $(grep -c -F "$a" "$registry")"
if [ -x "$b/target/debug/moai-cli" ]; then
    "$b/target/debug/moai-cli" agents-apply --help </dev/null | head -1
    echo "STEP help-b exit=${PIPESTATUS[0]}"
fi
