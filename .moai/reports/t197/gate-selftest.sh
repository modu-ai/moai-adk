#!/bin/bash
# Self-test for citation-sweep.sh — the step that was missing last round.
#
#   bash .moai/reports/t197/gate-selftest.sh
#
# A gate that has never been observed failing is indistinguishable from a
# report. This drives four inputs that MUST make the gate exit non-zero, one
# per check, and fails if any of them slips through. It restores every file it
# touches from a backup taken here, and verifies the restore byte-for-byte.
#
# rc 0 = the gate rejects everything it should. rc 1 = a check is asleep.

set -uo pipefail

GATE=.moai/reports/t197/citation-sweep.sh
MEAS=.moai/reports/t197/measurement.md
MAN=.moai/reports/t197/citation-manifest.txt

BK="$(mktemp -d)" || exit 2
cp "$MEAS" "$BK/measurement.md"
cp "$MAN" "$BK/manifest.txt"
restore() { cp "$BK/measurement.md" "$MEAS"; cp "$BK/manifest.txt" "$MAN"; }
trap 'restore; rm -rf "$BK"' EXIT

MISSES=0

expect_fail() { # <label> <grep-pattern the failure output must contain>
  local label="$1" want="$2" out rc
  out=$(bash "$GATE" 2>&1); rc=$?
  restore
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

echo "baseline (unmodified tree) must PASS:"
if bash "$GATE" > /dev/null 2>&1; then
  echo "        ok: rc=0"
else
  echo "MISS: the gate rejects the current tree — fix that before trusting this self-test"
  MISSES=$((MISSES + 1))
fi
echo

# C1 — a pin that disagrees with the transcript's own recorded HEAD.
perl -pi -e 's/1ed61e4ac/deadbeef1/ if $. == 20' "$MEAS"
expect_fail "C1 stale pin" "cites deadbeef1"

# C2 — a citation past the end of the transcript.
printf '\n측정 근거 (L9990-9999).\n' >> "$MEAS"
expect_fail "C2 out-of-bounds range" "out of bounds"

# C3 — a citation nothing in the manifest declares.
printf '\n측정 근거 (L36-54) 를 새 주장에 갖다 붙인다.\n' >> "$MEAS"
expect_fail "C3 undeclared citation" "no manifest row"

# C4 — the iteration-7 shape: a valid range that does not show the claim.
#      L211-214 is the path-constant grep; it contains no write call.
printf '%s\n' '.moai/specs/SPEC-CODEX-INIT-001/spec.md ~ L211-214 ~ 경로 상수 ~ writeAtomic\(' >> "$MAN"
expect_fail "C4 unsupported citation" "does NOT contain"

# C5 — a manifest row whose citation is gone.
printf '%s\n' '.moai/reports/t197/measurement.md ~ L271-272 ~ 없는주장 ~ rc=0' >> "$MAN"
expect_fail "C5 stale manifest row" "stale row"

restore
echo
echo "restore check:"
if cmp -s "$BK/measurement.md" "$MEAS" && cmp -s "$BK/manifest.txt" "$MAN"; then
  echo "        ok: both files byte-identical to the backup"
else
  echo "MISS: restore left a modified file behind"
  MISSES=$((MISSES + 1))
fi

echo
if [ "$MISSES" = 0 ]; then
  echo "SELFTEST PASS — every injected fault was rejected"
  exit 0
fi
echo "SELFTEST FAIL — $MISSES input(s) slipped through"
exit 1
