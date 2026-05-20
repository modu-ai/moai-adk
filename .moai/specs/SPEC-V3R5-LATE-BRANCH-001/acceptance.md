# SPEC-V3R5-LATE-BRANCH-001 — Acceptance Criteria

> Plan-phase artifact. Companion to `spec.md` (WHAT/WHY) and `plan.md` (HOW BUILT). This document covers HOW VERIFIED (binary ACs, edge cases, Given-When-Then scenarios, Definition of Done).

## 1. Binary Acceptance Criteria

All ACs are executable as a single shell command. No manual judgment is required. Run during run-phase REFACTOR and again at sync-phase pre-PR check.

### AC-LB-001 — Config switch persisted in git-strategy.yaml `team` section
**Given** SPEC-V3R5-LATE-BRANCH-001 is implemented
**When** `yq` queries the `team` section of `.moai/config/sections/git-strategy.yaml`
**Then** the 4 keys return the expected values

| Key | Expected value |
|---|---|
| `git_strategy.team.automation.auto_branch` | `false` |
| `git_strategy.team.automation.auto_pr` | `false` |
| `git_strategy.team.branch_creation.auto_enabled` | `false` |
| `git_strategy.team.branch_creation.prompt_always` | `true` |

**Verification command** (binary PASS = all 4 yq queries return expected values):
```bash
test "$(yq '.git_strategy.team.automation.auto_branch' .moai/config/sections/git-strategy.yaml)" = "false" && \
test "$(yq '.git_strategy.team.automation.auto_pr' .moai/config/sections/git-strategy.yaml)" = "false" && \
test "$(yq '.git_strategy.team.branch_creation.auto_enabled' .moai/config/sections/git-strategy.yaml)" = "false" && \
test "$(yq '.git_strategy.team.branch_creation.prompt_always' .moai/config/sections/git-strategy.yaml)" = "true" && \
echo "AC-LB-001 PASS" || echo "AC-LB-001 FAIL"
```

**Maps to**: REQ-LB-001, REQ-LB-002, REQ-LB-007

### AC-LB-002 — Skill body Phase 3 conditional present
**Given** the skill body modification per D2 is complete
**When** grep searches `.claude/skills/moai/workflows/plan/spec-assembly.md`
**Then** the conditional clause and Late-branch display mode are present

**Verification command**:
```bash
test "$(grep -c 'auto_enabled' .claude/skills/moai/workflows/plan/spec-assembly.md)" -ge 1 && \
test "$(grep -c 'Late-branch' .claude/skills/moai/workflows/plan/spec-assembly.md)" -ge 1 && \
echo "AC-LB-002 PASS" || echo "AC-LB-002 FAIL"
```

**Maps to**: REQ-LB-004

### AC-LB-003 — Agent body Personal Mode option + Invocation Pattern subsection
**Given** the agent body modification per D3 is complete
**When** grep searches `.claude/agents/moai/manager-git.md`
**Then** the `main_late_branch` row and "Late-Branch Invocation Pattern" subsection are present, and all 4 phases (A/B/C/D) are individually labeled

**Verification command**:
```bash
test "$(grep -c 'main_late_branch' .claude/agents/moai/manager-git.md)" -ge 1 && \
test "$(grep -c 'Late-Branch Invocation Pattern' .claude/agents/moai/manager-git.md)" -ge 1 && \
test "$(grep -cE '^(Phase A|Phase B|Phase C|Phase D)' .claude/agents/moai/manager-git.md)" -ge 4 && \
echo "AC-LB-003 PASS" || echo "AC-LB-003 FAIL"
```

**Maps to**: REQ-LB-005, REQ-LB-003

### AC-LB-004 — Rule update covers Step 1 entry + Step 4 cleanup
**Given** the rule update per D4 is complete
**When** grep searches `.claude/rules/moai/workflow/spec-workflow.md`
**Then** the `main checkout` precondition and the `git reset --hard origin/main` cleanup procedure are present

**Verification command**:
```bash
test "$(grep -c 'main checkout' .claude/rules/moai/workflow/spec-workflow.md)" -ge 1 && \
test "$(grep -c 'git reset --hard origin/main' .claude/rules/moai/workflow/spec-workflow.md)" -ge 1 && \
echo "AC-LB-004 PASS" || echo "AC-LB-004 FAIL"
```

**Maps to**: REQ-LB-006

### AC-LB-005 — Template mirror parity (4 files)
**Given** the template mirror updates per D5 are complete
**When** `diff` compares local sources to their template counterparts (after stripping `.tmpl` template variable expansions, which are absent in the affected sections of these specific files)
**Then** all 4 file pairs show zero drift, and the existing `internal/template/rule_template_mirror_test.go` PASSES

**Verification command** (for the 3 `.md` mirror pairs — literal byte-equivalence expected since affected sections contain no Go template variables; for the `.yaml.tmpl` pair, compare semantic content):
```bash
DRIFT_COUNT=0
# .md mirror pairs (literal diff)
for pair in \
  ".claude/skills/moai/workflows/plan/spec-assembly.md:internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md" \
  ".claude/agents/moai/manager-git.md:internal/template/templates/.claude/agents/moai/manager-git.md" \
  ".claude/rules/moai/workflow/spec-workflow.md:internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md"; do
  src="${pair%%:*}"; tgt="${pair##*:}"
  if ! diff -q "$src" "$tgt" >/dev/null 2>&1; then
    DRIFT_COUNT=$((DRIFT_COUNT + 1))
    echo "DRIFT: $src vs $tgt"
  fi
done
# .yaml.tmpl: 4 modified keys must exist in both files with same values
for key in 'auto_branch: false' 'auto_pr: false' 'auto_enabled: false' 'prompt_always: true'; do
  src_count=$(grep -c "$key" .moai/config/sections/git-strategy.yaml)
  tgt_count=$(grep -c "$key" internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl)
  if [ "$src_count" -lt 1 ] || [ "$tgt_count" -lt 1 ]; then
    DRIFT_COUNT=$((DRIFT_COUNT + 1))
    echo "YAML drift on key: $key (src=$src_count, tgt=$tgt_count)"
  fi
done
# Existing CI test must PASS
go test ./internal/template/ -run TestRuleTemplateMirror -count=1 || DRIFT_COUNT=$((DRIFT_COUNT + 1))
test "$DRIFT_COUNT" -eq 0 && echo "AC-LB-005 PASS" || echo "AC-LB-005 FAIL ($DRIFT_COUNT drifts)"
```

**Maps to**: REQ-LB-008, plan.md M5

### AC-LB-006 — E2E Late-branch scenario (dogfooding + scripted)
**Given** the 4-phase Late-branch procedure is documented (M3) and the config is in Late-branch default (M1)
**When** the maintainer (or test script) executes Phase A → B → C → D against a test repo
**Then** Phase D completes with `git status --porcelain` returning empty and `git rev-parse main == git rev-parse origin/main`

**Verification path (primary — dogfooding)**: This SPEC's own plan-PR + run-PR + sync-PR cycle is the first concrete E2E run. If all 3 PRs squash-merge cleanly and the post-merge `git reset --hard origin/main` leaves a clean working tree on `main`, AC-LB-006 PASS is empirically achieved on the maintainer's machine.

**Verification path (secondary — scripted, in /tmp)**:
```bash
set -e
TESTDIR=$(mktemp -d)
cd "$TESTDIR"
git init -b main
git commit --allow-empty -m "initial"
git update-ref refs/remotes/origin/main HEAD

# Phase A
echo "spec" > spec.md; git add . && git commit -m "spec(SPEC-TEST): plan"
# Phase B
echo "red" > impl.go; git add . && git commit -m "RED"
echo "green" > impl.go; git add . && git commit -m "GREEN"
# Phase C (simulate squash merge)
git switch -c feat/SPEC-TEST
git checkout main
git merge --squash feat/SPEC-TEST
git commit -m "feat: SPEC-TEST (squash)"
git update-ref refs/remotes/origin/main HEAD
git branch -D feat/SPEC-TEST
# Phase D
git reset --hard refs/remotes/origin/main
git pull . main 2>/dev/null || true   # fetch is no-op for self-remote

# Verify
PORCELAIN=$(git status --porcelain)
LOCAL_HEAD=$(git rev-parse main)
ORIGIN_HEAD=$(git rev-parse refs/remotes/origin/main)
test -z "$PORCELAIN" && test "$LOCAL_HEAD" = "$ORIGIN_HEAD" && \
  echo "AC-LB-006 PASS" || echo "AC-LB-006 FAIL"
cd / && rm -rf "$TESTDIR"
```

**Maps to**: REQ-LB-003, REQ-LB-006 (end-to-end integration)

### AC-LB-007 — No automatic GitHub Issue creation in /moai plan workflow (added v0.1.1)
**Given** REQ-LB-009 (no-auto-issue policy) is implemented
**When** grep searches the `spec-assembly.md` skill body for `gh issue create` invocations
**Then** either zero occurrences exist, OR every occurrence is gated by an explicit `--issue` flag check

**Verification command** (binary PASS = condition met):
```bash
# Primary check: zero unguarded gh issue create calls
UNGUARDED=$(grep -c "gh issue create" .claude/skills/moai/workflows/plan/spec-assembly.md || echo 0)

if [ "$UNGUARDED" -eq 0 ]; then
  echo "AC-LB-007 PASS (zero gh issue create occurrences)"
else
  # Secondary check: every occurrence must be gated by --issue flag
  # Approach: grep -B 5 'gh issue create' and look for --issue flag check within 5 lines above
  GATED=$(grep -B 5 "gh issue create" .claude/skills/moai/workflows/plan/spec-assembly.md | grep -c "\-\-issue")
  if [ "$GATED" -ge "$UNGUARDED" ]; then
    echo "AC-LB-007 PASS (all $UNGUARDED gh issue create calls gated by --issue flag)"
  else
    echo "AC-LB-007 FAIL ($UNGUARDED unguarded gh issue create calls)"
  fi
fi
```

**Maps to**: REQ-LB-009

---

## 2. Edge Cases

### EC-LB-001 — Concurrent SPEC work conflict
**Scenario**: User runs `/moai plan SPEC-V3R5-LATE-BRANCH-001` while uncommitted work for `SPEC-V3R5-OTHER-001` exists on `main`.

**Expected behavior**: BODP gate detects dirty working tree in Phase 3.0 and surfaces blocker report to orchestrator. Orchestrator invokes AskUserQuestion offering: (a) commit SPEC-V3R5-OTHER-001 first, (b) stash and proceed, (c) abort.

**Rationale**: Late-branch explicitly restricts one-SPEC-at-a-time per C-LB-002. Parallel SPECs require the worktree pattern, which is out of scope (EXCL-LB-005).

**Test**: Manual scenario during run-phase. No automated test required since BODP gate logic is unchanged by this SPEC.

### EC-LB-002 — Accidental `git push origin main` during Phase B
**Scenario**: User attempts `git push origin main` during Phase B (accumulating implementation commits on main).

**Expected behavior**: GitHub branch protection (4 required checks) rejects the push because main is protected. User receives a `! [remote rejected]` message.

**Recovery**: Orchestrator (next `/moai` invocation) detects via `git log origin/main..main --count > 0` and surfaces Phase C `git switch -c feat/SPEC-*` recommendation through AskUserQuestion.

**Test**: Documented in R-LB-001 (plan.md §5). No automated test in this SPEC.

### EC-LB-003 — Phase D omitted before next SPEC
**Scenario**: User completes squash merge of SPEC-X but skips `git reset --hard origin/main`, then runs `/moai plan SPEC-Y`.

**Expected behavior**: Phase A precondition check (`git rev-parse --abbrev-ref HEAD == main && git status --porcelain == empty && git rev-parse main == git rev-parse origin/main`) catches the divergence. BODP gate surfaces blocker report with recommended `git fetch origin && git reset --hard origin/main` command.

**Recovery**: User runs Phase D, then retries `/moai plan SPEC-Y`.

**Test**: Verifiable via the existing BODP gate logic; the precondition extension in REQ-LB-001 + Step 1 documentation update (M4) makes this case explicit.

### EC-LB-004 — Branch protection blocks squash merge during Phase C
**Scenario**: PR is opened in Phase C but CI fails (e.g., lint baseline violation). Squash merge is blocked.

**Expected behavior**: User can either (a) push fixup commits to `feat/SPEC-*` branch until CI passes, then squash-merge; or (b) abort the PR, return to `main`, accumulate additional fix commits, force-update `feat/SPEC-*` via `git switch feat/SPEC-* && git reset --hard main && git push --force-with-lease`.

**Rationale**: Late-branch does not change CI gate behavior — the branch protection ruleset is enforced regardless of when the branch is created. This is by design (`mode: team` preserved).

**Test**: Pattern (b) is the dogfooding equivalent of admin-squash-override observed in V3R5-CONSTITUTION-DUAL-001 (chicken-and-egg lint baseline). No new test surface in this SPEC.

---

## 3. Definition of Done

This SPEC is COMPLETE when ALL of the following are true:

- [ ] All 7 binary ACs (AC-LB-001 through AC-LB-007) return PASS
- [ ] All 4 edge cases (EC-LB-001 through EC-LB-004) have documented expected behavior in this file
- [ ] All 9 EARS REQs (REQ-LB-001 through REQ-LB-009, minus REQ-LB-008 if mirror test already covers) are mapped to ≥1 AC each
- [ ] `moai spec lint --strict` returns 0 findings on this SPEC
- [ ] Plan-auditor verdict on plan-phase artifacts is PASS ≥ 0.85
- [ ] All 5 source files (`.moai/`/`.claude/`) and 5 template mirrors (`internal/template/templates/`) are byte-equivalent (modulo template variables) — `go test ./internal/template/...` PASSES
- [ ] Dogfooding succeeded: this SPEC's plan/run/sync PRs all squash-merged cleanly with Phase D reset performed by maintainer
- [ ] v3.5.0 release notes draft includes the breaking-default callout per D1 Option (a) AND the no-auto-issue policy callout per D2 / REQ-LB-009
- [ ] No regression in existing `moai init` flow (verified by running `moai init /tmp/test-init-late-branch && grep -c 'auto_enabled: false' /tmp/test-init-late-branch/.moai/config/sections/git-strategy.yaml`)
- [ ] No regression in existing `moai update` flow (existing user `git-strategy.yaml` files are preserved verbatim AND existing SPEC `issue_number` fields retained per EXCL-LB-008)

## 4. Quality Gate Mapping (TRUST 5)

| Pillar | Coverage in this SPEC |
|---|---|
| **Tested** | 6 binary ACs + 4 edge cases + dogfooding E2E |
| **Readable** | 4-phase procedure documented in 3 places (memory, plan.md, manager-git.md) |
| **Unified** | All config keys flow through single source (`git-strategy.yaml`); template mirror parity enforced |
| **Secured** | REQ-LB-007 prohibits accidental `main` push; pre-push hook recommended as follow-on (EXCL-LB-002) |
| **Trackable** | SPEC commits land directly on main as dogfooding; PR-squash provides trackable history |

## 5. References

- spec.md (this SPEC) — WHAT/WHY
- plan.md (this SPEC) — HOW BUILT
- `feedback_late_branch_workflow.md` memory — workflow lessons + 4-phase procedure
- `project_v3r5_late_branch_decision.md` memory — decision rationale + paste-ready resume
- `.claude/rules/moai/workflow/spec-workflow.md` — Step 1 entry / Step 4 cleanup canonical reference (post-M4)
- `.claude/agents/moai/manager-git.md` — Late-Branch Invocation Pattern (post-M3)
