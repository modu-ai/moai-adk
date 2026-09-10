#!/bin/bash
# Self-test for citation-sweep.sh — the step that was missing before iteration 8.
#
#   bash .moai/reports/t197/gate-selftest.sh
#
# A gate that has never been observed failing is indistinguishable from a
# report. This drives five inputs that MUST make the gate exit non-zero, one per
# check, and fails if any of them slips through.
#
# It runs the gate against a THROWAWAY COPY of the corpus (MOAI_CITATION_ROOT),
# never against the working tree. The previous version injected faults into the
# tracked files themselves and restored them from a backup — which looked fine
# in isolation and was not: an audit ran it while another session ran it too,
# the second run backed up the first run's injected state, and the tree was left
# with a modified citation-manifest.txt that made the gate return rc 1. The
# restore check missed it because it compared the restored files against that
# same poisoned backup, so it could only ever pass.
#
# Two lessons are wired in below rather than written down: the sandbox removes
# the shared mutable state instead of sequencing access to it, and the final
# check asks git about the WHOLE repository rather than asking the script about
# the files it remembers touching.
#
# rc 0 = the gate rejects everything it should, and the working tree is untouched.

set -uo pipefail

GATE=.moai/reports/t197/citation-sweep.sh
MEAS=.moai/reports/t197/measurement.md
MAN=.moai/reports/t197/citation-manifest.txt

for f in "$GATE" "$MEAS" "$MAN"; do
  [ -r "$f" ] || { echo "FATAL: cannot read $f — run from the repository root"; exit 2; }
done

MISSES=0

# The tree state before anything runs. The final check compares against this
# rather than against "clean", so the self-test is honest about a tree that was
# already dirty when it started instead of blaming its own run for it.
BEFORE=$(git status --porcelain)

SB="$(mktemp -d)" || exit 2
trap 'rm -rf "$SB"' EXIT

mkdir -p "$SB/.moai/specs" "$SB/.moai/reports"
cp -R .moai/specs/SPEC-CODEX-LAUNCHER-001 "$SB/.moai/specs/"
cp -R .moai/specs/SPEC-CODEX-INIT-001 "$SB/.moai/specs/"
cp -R .moai/reports/t197 "$SB/.moai/reports/"

SB_MEAS="$SB/$MEAS"
SB_MAN="$SB/$MAN"

# A pristine second copy, so an injection is undone by restoring from something
# the run itself can never have modified.
PRISTINE="$SB/.pristine"
mkdir -p "$PRISTINE"
cp "$MEAS" "$PRISTINE/measurement.md"
cp "$MAN" "$PRISTINE/manifest.txt"
reset_sandbox() {
  cp "$PRISTINE/measurement.md" "$SB_MEAS"
  cp "$PRISTINE/manifest.txt" "$SB_MAN"
}

run_gate() { MOAI_CITATION_ROOT="$SB" bash "$GATE" 2>&1; }

expect_fail() { # <label> <pattern the failure output must contain>
  local label="$1" want="$2" out rc
  out=$(run_gate); rc=$?
  reset_sandbox
  if [ "$rc" = 0 ]; then
    echo "MISS: [$label] gate exited 0 — this input should have been rejected"
    MISSES=$((MISSES + 1))
    return
  fi
  if printf '%s' "$out" | grep -q "$want"; then
    echo "caught: [$label] rc=$rc"
    printf '%s\n' "$out" | grep '^FAIL' | sed 's/^/        /'
  else
    echo "MISS: [$label] gate exited $rc but not for the expected reason"
    printf '%s\n' "$out" | grep '^FAIL' | sed 's/^/        /'
    MISSES=$((MISSES + 1))
  fi
}

echo "sandbox: $SB"
echo "baseline (unmodified copy) must PASS:"
if run_gate > /dev/null 2>&1; then
  echo "        ok: rc=0"
else
  echo "MISS: the gate rejects the current corpus — fix that before trusting this self-test"
  run_gate | grep '^FAIL' | sed 's/^/        /'
  MISSES=$((MISSES + 1))
fi
echo

# C1 — a pin that disagrees with the transcript's own recorded HEAD.
perl -pi -e 's/1ed61e4ac/deadbeef1/ if $. == 20' "$SB_MEAS"
expect_fail "C1 stale pin" "cites deadbeef1"

# C2 — a citation past the end of the transcript.
printf '\n측정 근거 (L9990-9999).\n' >> "$SB_MEAS"
expect_fail "C2 out-of-bounds range" "out of bounds"

# C3 — a citation nothing in the manifest declares.
printf '\n측정 근거 (L36-54) 를 새 주장에 갖다 붙인다.\n' >> "$SB_MEAS"
expect_fail "C3 undeclared citation" "no manifest row"

# C4 — the iteration-7 shape: a valid range that does not show the claim.
#      L211-214 is the path-constant grep; it contains no write call.
printf '%s\n' '.moai/specs/SPEC-CODEX-INIT-001/spec.md ~ L211-214 ~ 경로 상수 ~ writeAtomic\(' >> "$SB_MAN"
expect_fail "C4 unsupported citation" "does NOT contain"

# C5 — a manifest row whose citation is gone.
printf '%s\n' '.moai/reports/t197/measurement.md ~ L271-272 ~ 없는주장 ~ rc=0' >> "$SB_MAN"
expect_fail "C5 stale manifest row" "stale row"

echo
echo "working-tree check (git, whole repository — not a list this script keeps):"
AFTER=$(git status --porcelain)
if [ "$BEFORE" = "$AFTER" ]; then
  echo "        ok: git reports no change caused by this run"
else
  echo "MISS: this run changed the working tree"
  diff <(printf '%s\n' "$BEFORE") <(printf '%s\n' "$AFTER") | sed 's/^/        /'
  MISSES=$((MISSES + 1))
fi

echo
if [ "$MISSES" = 0 ]; then
  echo "SELFTEST PASS — every injected fault was rejected, working tree untouched"
  exit 0
fi
echo "SELFTEST FAIL — $MISSES check(s) did not hold"
exit 1
