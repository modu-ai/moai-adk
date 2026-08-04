#!/usr/bin/env bash
# verify_test.sh — SPEC-LSEL-LOCAL-EVOLUTION-001 M4 VERIFY characterization test.
#
# Exercises the REAL verify.sh mechanism inside an hermetic temp git repo (no
# pollution of the host worktree). Asserts AC-LSEL-015:
#   - both verify_command AND /moai gate named (gate is MANDATORY, not optional)
#   - timeout-class failure on first VERIFY run → retries exactly once
#   - second non-timeout failure → git revert lsel-<id> auto-fires + verified:false
#   - a passing verify_command → verified:true (no revert)
#   - the applier SKILL body names /moai gate as MANDATORY (grep AC)
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
APPLIER_HOOK="$PROJECT_ROOT/.moai/hooks/lsel-apply.sh"
ALLOWLIST="$PROJECT_ROOT/.claude/lsel/frozen-allowlist.json"
VERIFY="$PROJECT_ROOT/.claude/skills/hns-lsel-applier/verify.sh"
SKILL="$PROJECT_ROOT/.claude/skills/hns-lsel-applier/SKILL.md"

FAIL=0
log(){ printf '%s\n' "$*"; }
fail(){ log "FAIL: $*"; FAIL=1; }
pass(){ log "PASS: $*"; }

# --- RED guard: mechanism must exist ---
if [[ ! -f "$VERIFY" ]]; then
  log "RED: verify.sh absent at $VERIFY (mechanism not built yet)"
  exit 1
fi
if [[ ! -f "$APPLIER_HOOK" ]] || [[ ! -f "$ALLOWLIST" ]]; then
  log "RED: M3 foundation (lsel-apply.sh / frozen-allowlist.json) missing"
  exit 1
fi

setup_repo() {
  # $1 = repo dir to initialize
  local repo="$1"
  git -C "$repo" init -q
  git -C "$repo" config user.email "lsel@test.local"
  git -C "$repo" config user.name "lsel-test"
  mkdir -p "$repo/.claude/lsel" "$repo/.moai/hooks" "$repo/.moai/state/lsel" \
           "$repo/.moai/logs" "$repo/memory"
  cp "$ALLOWLIST" "$repo/.claude/lsel/frozen-allowlist.json"
  cp "$APPLIER_HOOK" "$repo/.moai/hooks/lsel-apply.sh"
  chmod +x "$repo/.moai/hooks/lsel-apply.sh"
  echo "# seed" > "$repo/seed.txt"
  git -C "$repo" add -A
  git -C "$repo" commit -qm "init"
}

apply_proposal() {
  # $1 repo  $2 proposal-dir  -> runs the M3 applier, returns lsel commit sha
  local repo="$1" pdir="$2"
  LSEL_PROJECT_ROOT="$repo" bash "$repo/.moai/hooks/lsel-apply.sh" "$pdir/decision.json" >/dev/null 2>&1
  git -C "$repo" log --grep "lsel-" --format='%h' | head -1
}

# =====================================================================
log "=== AC-LSEL-015 grep: applier SKILL names /moai gate as MANDATORY ==="
if grep -qi '/moai gate' "$SKILL" && grep -qi 'MANDATORY' "$SKILL"; then
  pass "applier SKILL body names /moai gate + MANDATORY"
else
  fail "applier SKILL body does NOT name /moai gate as MANDATORY (circular-verify hazard, report §11 mustFix B#6)"
fi

# =====================================================================
log "=== AC-LSEL-015 fixture A: passing verify_command → verified:true, no revert ==="
REPO_A="$(mktemp -d)"; trap 'rm -rf "$REPO_A" "$REPO_B" "$REPO_C"' EXIT
setup_repo "$REPO_A"
mkdir -p "$REPO_A/pA"
cat > "$REPO_A/pA/decision.json" <<'JSON'
{"proposal_id":"lsel-verify-pass","target_surface":"memory/feedback_pass.md","synchronous_approval":null}
JSON
cat > "$REPO_A/pA/diff.patch" <<'PATCH'
diff --git a/memory/feedback_pass.md b/memory/feedback_pass.md
new file mode 100644
--- /dev/null
+++ b/memory/feedback_pass.md
@@ -0,0 +1 @@
+pass-marker
PATCH
cat > "$REPO_A/pA/proposal.md" <<'MD'
---
proposal_id: lsel-verify-pass
verify_command: test -f memory/feedback_pass.md
---
pass fixture
MD
SHA_A="$(apply_proposal "$REPO_A" "$REPO_A/pA")" || { fail "fixture A apply failed"; }
LSEL_PROJECT_ROOT="$REPO_A" bash "$VERIFY" --proposal-dir "$REPO_A/pA" --repo-root "$REPO_A" --timeout 5 >/tmp/vA.out 2>&1 || true
if grep -q '"verified":true\|"verified": true' "$REPO_A/.moai/state/lsel/apply-ledger.jsonl" 2>/dev/null; then
  pass "fixture A: ledger row verified:true"
else
  fail "fixture A: no verified:true ledger row ($(cat /tmp/vA.out))"
fi
# pipe-free check: `git log | grep -q` races with SIGPIPE under `set -o pipefail`
# (grep exits early → git gets SIGPIPE 141 → pipeline nonzero → false negative).
LOG_A="$(git -C "$REPO_A" log --oneline)"
if [[ "$LOG_A" == *"$SHA_A"* ]]; then
  pass "fixture A: lsel commit survived (no revert) — verified:true, no revert"
else
  fail "fixture A: lsel commit was reverted despite a passing verify (should NOT revert)"
fi

# =====================================================================
log "=== AC-LSEL-015 fixture B: timeout-class on first run → retries exactly once → pass ==="
REPO_B="$(mktemp -d)"; setup_repo "$REPO_B"
mkdir -p "$REPO_B/pB"
cat > "$REPO_B/pB/decision.json" <<'JSON'
{"proposal_id":"lsel-verify-flaky","target_surface":"memory/feedback_flaky.md","synchronous_approval":null}
JSON
cat > "$REPO_B/pB/diff.patch" <<'PATCH'
diff --git a/memory/feedback_flaky.md b/memory/feedback_flaky.md
new file mode 100644
--- /dev/null
+++ b/memory/feedback_flaky.md
@@ -0,0 +1 @@
+flaky-marker
PATCH
# flaky verifier: first call sleeps past the timeout (killed → TIMEOUT), second call passes.
cat > "$REPO_B/flaky.sh" <<'SH'
#!/usr/bin/env bash
n=$(cat .flaky-counter 2>/dev/null || echo 0)
echo $((n+1)) > .flaky-counter
if [ "$n" -eq 0 ]; then
  sleep 30   # first invocation: will be killed by the timeout wrapper
else
  exit 0     # second invocation: pass
fi
SH
chmod +x "$REPO_B/flaky.sh"
cat > "$REPO_B/pB/proposal.md" <<'MD'
---
proposal_id: lsel-verify-flaky
verify_command: bash flaky.sh
---
flaky fixture
MD
SHA_B="$(apply_proposal "$REPO_B" "$REPO_B/pB")" || { fail "fixture B apply failed"; }
LSEL_PROJECT_ROOT="$REPO_B" bash "$VERIFY" --proposal-dir "$REPO_B/pB" --repo-root "$REPO_B" --timeout 2 >/tmp/vB.out 2>&1 || true
# The counter file proves how many attempts ran.
ATTEMPTS_B="$(cat "$REPO_B/.flaky-counter" 2>/dev/null || echo 0)"
if [[ "$ATTEMPTS_B" -eq 2 ]]; then
  pass "fixture B: retried exactly once after timeout (attempts=$ATTEMPTS_B)"
else
  fail "fixture B: expected exactly 2 attempts (timeout-retry-once), got $ATTEMPTS_B"
fi
if grep -q '"verified":true\|"verified": true' "$REPO_B/.moai/state/lsel/apply-ledger.jsonl" 2>/dev/null; then
  pass "fixture B: ledger row verified:true (flaky tolerated)"
else
  fail "fixture B: no verified:true after tolerated timeout ($(cat /tmp/vB.out))"
fi
LOG_B="$(git -C "$REPO_B" log --oneline)"
if [[ "$LOG_B" == *"$SHA_B"* ]]; then
  pass "fixture B: lsel commit survived (no revert — flaky tolerated)"
else
  fail "fixture B: lsel commit reverted despite a tolerated-timeout pass"
fi

# =====================================================================
log "=== AC-LSEL-015 fixture C: second non-timeout failure → git revert + verified:false ==="
REPO_C="$(mktemp -d)"; setup_repo "$REPO_C"
mkdir -p "$REPO_C/pC"
cat > "$REPO_C/pC/decision.json" <<'JSON'
{"proposal_id":"lsel-verify-failtwice","target_surface":"memory/feedback_failtwice.md","synchronous_approval":null}
JSON
cat > "$REPO_C/pC/diff.patch" <<'PATCH'
diff --git a/memory/feedback_failtwice.md b/memory/feedback_failtwice.md
new file mode 100644
--- /dev/null
+++ b/memory/feedback_failtwice.md
@@ -0,0 +1 @@
+failtwice-marker
PATCH
cat > "$REPO_C/pC/proposal.md" <<'MD'
---
proposal_id: lsel-verify-failtwice
verify_command: exit 1
---
failtwice fixture (deterministic non-timeout failure)
MD
SHA_C="$(apply_proposal "$REPO_C" "$REPO_C/pC")" || { fail "fixture C apply failed"; }
# The proposal's associated feedback file (to be marked verified:false).
mkdir -p "$REPO_C/memory"
cat > "$REPO_C/memory/feedback_failtwice.md.lsel" <<'MD'
# feedback_failtwice (pre-existing body)
MD
# verify.sh --feedback-file points at the proposal's feedback topic file.
LSEL_PROJECT_ROOT="$REPO_C" bash "$VERIFY" --proposal-dir "$REPO_C/pC" --repo-root "$REPO_C" \
  --timeout 5 --feedback-file "$REPO_C/memory/feedback_lsel-verify-failtwice.md" >/tmp/vC.out 2>&1 || true
if grep -q '"verified":false\|"verified": false' "$REPO_C/.moai/state/lsel/apply-ledger.jsonl" 2>/dev/null; then
  pass "fixture C: ledger row verified:false"
else
  fail "fixture C: no verified:false ledger row ($(cat /tmp/vC.out))"
fi
# git revert adds an INVERSE commit (it does not erase history), so the right
# signal is a landed revert commit + the applied change undone — not the
# original commit vanishing from the log.
LOG_C="$(git -C "$REPO_C" log --oneline)"
if printf '%s\n' "$LOG_C" | grep -qi 'Revert "feat(lsel-'; then
  pass "fixture C: git revert landed an inverse commit (2nd non-timeout failure → auto-revert fired)"
else
  fail "fixture C: no revert commit in log (git revert did NOT fire on 2nd non-timeout failure)"
fi
# The applied file (memory/feedback_failtwice.md) must be undone by the revert.
if [[ ! -f "$REPO_C/memory/feedback_failtwice.md" ]] || \
   ! grep -q 'failtwice-marker' "$REPO_C/memory/feedback_failtwice.md" 2>/dev/null; then
  pass "fixture C: applied change undone by revert (target file content reverted)"
else
  fail "fixture C: applied change still present after revert (revert ineffective)"
fi
if grep -qi 'verified: false\|verified:false' "$REPO_C/memory/feedback_lsel-verify-failtwice.md" 2>/dev/null; then
  pass "fixture C: feedback_*.md marked verified:false"
else
  fail "fixture C: feedback_*.md not marked verified:false"
fi

if [[ "$FAIL" -ne 0 ]]; then
  log "verify_test: FAILED"
  exit 1
fi
log "verify_test: PASS"
exit 0
