#!/usr/bin/env bash
# rollback_rehearsal_test.sh — AC-LSEL-014 SHIP GATE (SPEC-LSEL-LOCAL-EVOLUTION-001 M3).
#
# Builds a fixture CLAUDE.local.md history (≥1 lsel-* commit interleaved with ≥2
# manual edits) in an hermetic temp git repo, then proves `git revert <lsel-tag>`
# lands clean (exit 0) AND the post-revert state matches the pre-lsel state for
# the affected lines. If this test FAILS, M3 does NOT ship (plan.md §F.3).
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
APPLIER="$PROJECT_ROOT/.moai/hooks/lsel-apply.sh"
ALLOWLIST="$PROJECT_ROOT/.claude/lsel/frozen-allowlist.json"

FAIL=0
log(){ printf '%s\n' "$*"; }
fail(){ log "FAIL: $*"; FAIL=1; }
pass(){ log "PASS: $*"; }

if [[ ! -f "$APPLIER" ]] || [[ ! -f "$ALLOWLIST" ]]; then
  log "RED: mechanism absent — lsel-apply.sh or frozen-allowlist.json missing"
  exit 1
fi

REPO="$(mktemp -d)"
trap 'rm -rf "$REPO"' EXIT
cd "$REPO"
git init -q
git config user.email "lsel@test.local"
git config user.name "lsel-test"
mkdir -p .claude/lsel .moai/hooks .moai/state/lsel .moai/logs
cp "$ALLOWLIST" .claude/lsel/frozen-allowlist.json
cp "$APPLIER" .moai/hooks/lsel-apply.sh
chmod +x .moai/hooks/lsel-apply.sh
export LSEL_PROJECT_ROOT="$REPO"

# --- seed CLAUDE.local.md (manual edit #1) ---
cat > CLAUDE.local.md <<'MD'
# CLAUDE.local.md fixture

## Section A
manual line A1
manual line A2

## Section B
manual line B1
MD
git add -A; git commit -qm "manual edit 1: seed CLAUDE.local.md"

# --- manual edit #2: touch Section A (NOT the lsel block region) ---
cat > CLAUDE.local.md <<'MD'
# CLAUDE.local.md fixture

## Section A
manual line A1
manual line A2
manual line A3 (edit 2)

## Section B
manual line B1
MD
git add -A; git commit -qm "manual edit 2: extend Section A"

# --- lsel-* apply: append a uniquely-marked block at the END (Section C) ---
PRE_LSEL_MD="$(cat CLAUDE.local.md)"
mkdir -p "$REPO/p-lsel"
cat > "$REPO/p-lsel/decision.json" <<'JSON'
{"proposal_id":"lsel-rehearse-001","target_surface":"CLAUDE.local.md","synchronous_approval":null}
JSON
# Generate a correct unified diff by staging the intended change, capturing the
# diff, then restoring the pre-state so lsel-apply.sh plays it back fresh.
cat >> CLAUDE.local.md <<'MD'

## Section C (lsel)
lsel-block-start
lsel-applied-content
lsel-block-end
lsel-block-mark
MD
git diff -- CLAUDE.local.md > "$REPO/p-lsel/diff.patch"
git checkout -- CLAUDE.local.md
if [[ ! -s "$REPO/p-lsel/diff.patch" ]]; then
  fail "generated diff.patch is empty (staged change did not produce a diff)"
  log "rollback_rehearsal_test: FAILED"; exit 1
fi
if ! bash "$REPO/.moai/hooks/lsel-apply.sh" "$REPO/p-lsel/decision.json"; then
  fail "lsel-apply.sh failed to apply the rehearsal proposal"
  log "rollback_rehearsal_test: FAILED"; exit 1
fi
LSEL_SHA=$(git rev-parse HEAD)
if ! git log --grep 'lsel-rehearse-001' --oneline | grep -q 'lsel-rehearse-001'; then
  fail "lsel-rehearse-001 commit did not land"
  log "rollback_rehearsal_test: FAILED"; exit 1
fi
pass "lsel-* commit landed in the mixed history"

# --- manual edit #3: modify an EXISTING Section A line (NOT the EOF lsel region).
# Touching a non-adjacent region keeps the lsel block's EOF context stable so the
# revert lands clean.
python3 - <<'PY'
import pathlib
p = pathlib.Path("CLAUDE.local.md")
s = p.read_text()
s = s.replace("manual line A3 (edit 2)", "manual line A3 (edit 2, revised by edit 3)")
p.write_text(s)
PY
git add -A; git commit -qm "manual edit 3: revise Section A line (non-adjacent to lsel block)"

# Sanity: history now has ≥1 lsel-* commit interleaved with ≥2 manual edits.
MANUAL_COUNT=$(git log --oneline | grep -c '^.*manual edit' || true)
if [[ "$MANUAL_COUNT" -lt 2 ]]; then
  fail "fixture history lacks ≥2 manual edits (found $MANUAL_COUNT)"
fi
pass "fixture history: $MANUAL_COUNT manual edits interleaved with the lsel-* commit"

# --- AC-LSEL-014: git revert <lsel-tag> lands clean (exit 0) ---
set +e
git revert --no-edit "$LSEL_SHA" > /tmp/lsel-revert.out 2>&1
RC=$?
set -e
if [[ $RC -ne 0 ]]; then
  fail "git revert $LSEL_SHA exited non-zero ($RC) — conflict on mixed history"
  cat /tmp/lsel-revert.out
  log "rollback_rehearsal_test: FAILED — M3 MUST NOT ship"
  exit 1
fi
pass "git revert <lsel-tag> landed clean (exit 0)"

# --- post-revert state matches pre-lsel state for the affected (Section C) lines ---
if grep -q 'lsel-block-start\|lsel-applied-content\|lsel-block-end\|lsel-block-mark' CLAUDE.local.md; then
  fail "post-revert CLAUDE.local.md still contains lsel-block lines (revert incomplete)"
else
  pass "post-revert CLAUDE.local.md: lsel-block lines removed (matches pre-lsel state for affected lines)"
fi
# Sections A (untouched by the lsel commit) must survive the revert.
if grep -q 'manual line A3 (edit 2, revised by edit 3)' CLAUDE.local.md; then
  pass "manual edits outside the lsel block survived the revert"
else
  fail "revert clobbered manual edits outside the lsel block"
fi

if [[ "$FAIL" -ne 0 ]]; then
  log "rollback_rehearsal_test: FAILED"
  exit 1
fi
log "rollback_rehearsal_test: PASS"
exit 0
