#!/usr/bin/env bash
# apply_test.sh — SPEC-LSEL-LOCAL-EVOLUTION-001 M3 APPLY characterization test.
#
# Exercises the REAL lsel-apply.sh + frozen-allowlist.json mechanism inside an
# hermetic temp git repo (no pollution of the host worktree). Asserts:
#   AC-LSEL-001  frozen-path hard-reject (reject-log row + no write)
#   AC-LSEL-002  allowlist lives OUTSIDE the 6 evolvable surfaces
#   AC-LSEL-005  execution-meta without synchronous-approval marker REFUSED;
#                with marker proceeds
#   AC-LSEL-008  playback-only (no forbidden primitives in source)
#   AC-LSEL-013  routine apply → file written + ledger row + lsel-* commit
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
APPLIER="$PROJECT_ROOT/.moai/hooks/lsel-apply.sh"
ALLOWLIST="$PROJECT_ROOT/.claude/lsel/frozen-allowlist.json"

FAIL=0
log(){ printf '%s\n' "$*"; }
fail(){ log "FAIL: $*"; FAIL=1; }
pass(){ log "PASS: $*"; }

# --- RED guard: mechanism must exist ---
if [[ ! -f "$APPLIER" ]]; then
  log "RED: lsel-apply.sh absent at $APPLIER (mechanism not built yet)"
  exit 1
fi
if [[ ! -f "$ALLOWLIST" ]]; then
  log "RED: frozen-allowlist.json absent at $ALLOWLIST (mechanism not built yet)"
  exit 1
fi

# --- hermetic temp repo ---
REPO="$(mktemp -d)"
trap 'rm -rf "$REPO"' EXIT
cd "$REPO"
git init -q
git config user.email "lsel@test.local"
git config user.name "lsel-test"
mkdir -p .claude/lsel .claude/rules/moai .claude/skills/hns-lsel-applier \
         .moai/hooks .moai/state/lsel .moai/logs memory
cp "$ALLOWLIST" .claude/lsel/frozen-allowlist.json
cp "$APPLIER" .moai/hooks/lsel-apply.sh
chmod +x .moai/hooks/lsel-apply.sh
export LSEL_PROJECT_ROOT="$REPO"
echo "# frozen doctrine file" > .claude/rules/moai/test.md
echo "# applier skill body" > .claude/skills/hns-lsel-applier/SKILL.md
git add -A
git commit -qm "init"

run_applier(){ bash "$REPO/.moai/hooks/lsel-apply.sh" "$1"; }

log "=== AC-LSEL-002 allowlist location (outside the 6 evolvable surfaces) ==="
# The 6 evolvable surfaces per spec.md §B.3:
#   CLAUDE.local.md, .claude/settings.local.json, memory/,
#   .claude/agents/harness/** + hns-* skills, .moai/lessons-inbox.jsonl,
#   .moai/state/ (+.moai/state/lsel/)
for s in CLAUDE.local.md settings.local.json memory lessons-inbox.jsonl state; do
  found=$(find "$REPO" -path "*$s*frozen-allowlist.json" 2>/dev/null | head -1 || true)
  if [[ -n "$found" ]]; then fail "allowlist leaked under evolvable surface $s ($found)"; fi
done
if [[ -f "$REPO/.claude/lsel/frozen-allowlist.json" ]]; then
  pass "allowlist at .claude/lsel/ (outside all 6 evolvable surfaces)"
else
  fail "allowlist not at canonical .claude/lsel/ path"
fi

log "=== AC-LSEL-008 playback-only (no forbidden primitives) ==="
if grep -nE 'propose|self-approve|new-proposal' "$REPO/.moai/hooks/lsel-apply.sh"; then
  fail "forbidden primitive found in lsel-apply.sh source"
else
  pass "no forbidden primitives (playback-only)"
fi
if grep -q 'decision\.json' "$REPO/.moai/hooks/lsel-apply.sh" && \
   grep -qi 'approv' "$REPO/.moai/hooks/lsel-apply.sh"; then
  pass "script reads an approved decision.json"
else
  fail "script does not reference decision.json + approval"
fi

log "=== AC-LSEL-001 frozen-path hard-reject ==="
mkdir -p "$REPO/p-frozen"
cat > "$REPO/p-frozen/decision.json" <<'JSON'
{"proposal_id":"lsel-frozen-001","target_surface":".claude/rules/moai/test.md","synchronous_approval":null}
JSON
cat > "$REPO/p-frozen/diff.patch" <<'PATCH'
diff --git a/.claude/rules/moai/test.md b/.claude/rules/moai/test.md
--- a/.claude/rules/moai/test.md
+++ b/.claude/rules/moai/test.md
@@ -1 +1,2 @@
 # frozen doctrine file
+injected-by-lsel
PATCH
if run_applier "$REPO/p-frozen/decision.json" 2>/dev/null; then
  fail "frozen target was NOT rejected"
else
  pass "frozen target REFUSED (non-zero exit)"
fi
if grep -q "lsel-frozen-001" "$REPO/.moai/logs/lsel-reject.log" 2>/dev/null && \
   grep -q "frozen-path" "$REPO/.moai/logs/lsel-reject.log" 2>/dev/null; then
  pass "reject-log row appended naming the frozen target"
else
  fail "no reject-log row for frozen target"
fi
if grep -q "injected-by-lsel" "$REPO/.claude/rules/moai/test.md" 2>/dev/null; then
  fail "frozen target WAS written (allowlist failed to protect)"
else
  pass "no write to frozen target"
fi

log "=== AC-LSEL-013 routine apply (evolvable target) ==="
mkdir -p "$REPO/p-routine"
cat > "$REPO/p-routine/decision.json" <<'JSON'
{"proposal_id":"lsel-routine-001","target_surface":"memory/feedback_routine.md","synchronous_approval":null}
JSON
cat > "$REPO/p-routine/diff.patch" <<'PATCH'
diff --git a/memory/feedback_routine.md b/memory/feedback_routine.md
new file mode 100644
--- /dev/null
+++ b/memory/feedback_routine.md
@@ -0,0 +1 @@
+lsel-applied-line
PATCH
if run_applier "$REPO/p-routine/decision.json"; then
  pass "routine proposal applied (exit 0)"
else
  fail "routine proposal failed"
fi
if [[ -f "$REPO/memory/feedback_routine.md" ]] && \
   grep -q "lsel-applied-line" "$REPO/memory/feedback_routine.md"; then
  pass "diff applied to evolvable target"
else
  fail "diff NOT applied"
fi
if grep -q '"proposal_id":"lsel-routine-001"' "$REPO/.moai/state/lsel/apply-ledger.jsonl" 2>/dev/null && \
   grep -q '"result":"applied"' "$REPO/.moai/state/lsel/apply-ledger.jsonl" 2>/dev/null; then
  pass "apply-ledger.jsonl row appended (proposal_id + result:applied)"
else
  fail "no ledger row"
fi
if git -C "$REPO" log --grep 'lsel-routine-001' --oneline | grep -q 'lsel-routine-001'; then
  pass "lsel-routine-001 tagged commit landed"
else
  fail "no lsel-tagged commit"
fi
# no reject-log row for this proposal (allowlist validated the target)
if grep -q "lsel-routine-001" "$REPO/.moai/logs/lsel-reject.log" 2>/dev/null; then
  fail "reject-log row exists for a valid routine proposal"
else
  pass "no reject-log row for routine proposal (allowlist validated target)"
fi

log "=== AC-LSEL-005 execution-meta forced-gate (D3 mechanical refusal) ==="
# execution-meta category: an applier/curator skill body
mkdir -p "$REPO/p-em-wo"
cat > "$REPO/p-em-wo/decision.json" <<'JSON'
{"proposal_id":"lsel-em-wo","target_surface":".claude/skills/hns-lsel-applier/SKILL.md","synchronous_approval":null}
JSON
cat > "$REPO/p-em-wo/diff.patch" <<'PATCH'
diff --git a/.claude/skills/hns-lsel-applier/SKILL.md b/.claude/skills/hns-lsel-applier/SKILL.md
--- a/.claude/skills/hns-lsel-applier/SKILL.md
+++ b/.claude/skills/hns-lsel-applier/SKILL.md
@@ -1 +1,2 @@
 # applier skill body
+without-marker-attempt
PATCH
if run_applier "$REPO/p-em-wo/decision.json" 2>/dev/null; then
  fail "execution-meta WITHOUT marker was NOT refused"
else
  pass "execution-meta WITHOUT marker REFUSED"
fi
if grep -q "lsel-em-wo" "$REPO/.moai/logs/lsel-reject.log" 2>/dev/null && \
   grep -q "execution-meta" "$REPO/.moai/logs/lsel-reject.log" 2>/dev/null; then
  pass "reject-log row appended naming execution-meta category"
else
  fail "no reject-log row for execution-meta refusal"
fi

# WITH synchronous-approval marker → proceeds
mkdir -p "$REPO/p-em-w"
cat > "$REPO/p-em-w/decision.json" <<'JSON'
{"proposal_id":"lsel-em-w","target_surface":".claude/skills/hns-lsel-applier/SKILL.md","synchronous_approval":{"decision":"approved","channel":"orchestrator-synchronous-user-gate","ts":"2026-08-04T00:00:00Z"}}
JSON
cat > "$REPO/p-em-w/diff.patch" <<'PATCH'
diff --git a/.claude/skills/hns-lsel-applier/SKILL.md b/.claude/skills/hns-lsel-applier/SKILL.md
--- a/.claude/skills/hns-lsel-applier/SKILL.md
+++ b/.claude/skills/hns-lsel-applier/SKILL.md
@@ -1 +1,2 @@
 # applier skill body
+with-marker-applied
PATCH
if run_applier "$REPO/p-em-w/decision.json"; then
  pass "execution-meta WITH marker proceeds (refusal keyed on absent marker, not category)"
else
  fail "execution-meta WITH marker was refused (gate over-fired)"
fi

if [[ "$FAIL" -ne 0 ]]; then
  log "apply_test: FAILED"
  exit 1
fi
log "apply_test: PASS"
exit 0
