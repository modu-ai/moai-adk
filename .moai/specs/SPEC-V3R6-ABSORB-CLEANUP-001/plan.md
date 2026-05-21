# Plan — SPEC-V3R6-ABSORB-CLEANUP-001

## 1. Implementation Strategy

Tier S LEAN: single fix commit, 4 files affected, no design phase needed (mechanical baseline restoration).

### 1.1 File-by-file edit table

| # | File | Operation | Anchor | Change |
|---|------|-----------|--------|--------|
| 1 | `.claude/skills/moai/workflows/plan.md` | Edit (insert) | End of file or after a relevant prose section | Add sentence: "CI guards in `internal/template/agentless_audit_test.go` enforce the literal `MODE_PIPELINE_ONLY_UTILITY` sentinel remains present in this skill body (REQ-WF003-016 ↔ REQ-WF004-014, shared with `design.md`)." Single line, no new heading. |
| 2 | `.claude/skills/moai/workflows/sync.md` | Edit (insert) | Same pattern | Add sentence: "CI guards in `internal/template/agentless_audit_test.go` enforce the literal `MODE_PIPELINE_ONLY_UTILITY` sentinel remains present in this skill body (REQ-WF003-016 ↔ REQ-WF004-014, shared with `design.md`)." |
| 3 | `.claude/skills/moai/workflows/run.md` | Edit (insert) | Same pattern | Add sentence: "CI guards in `internal/template/agentless_audit_test.go` enforce the literal `MODE_UNKNOWN` sentinel remains present in this skill body (REQ-WF003-010, shared with `design.md`)." |
| 4 | `internal/template/catalog_tier_audit_test.go` | Edit (replace) | Line ~196-198 | Change `const expectedAgentCount = 20` to `const expectedAgentCount = 19`. Replace prior comment block `// Workflow audit 2026-05-16 Bundle C / F-003: 8 zombie agents purged.` and `// Expected: 28 − 8 = 20 (14 active system + 4 my-harness + 2 evaluator-family).` with updated breakdown: `// Workflow audit 2026-05-16 Bundle C / F-003: 8 zombie agents purged.` then `// Wave 1 ABSORB-CLEANUP-001 reconciliation 2026-05-22: my-harness category` and `// removed from disk; actual breakdown is 8 manager + 6 expert + 1 builder + 1` and `// evaluator-active + 1 plan-auditor + 1 researcher + 1 claude-code-guide = 19.` |

Files affected: **4 total**. Within Tier S budget (≤5).

### 1.2 Sentinel sentence template (canonical form)

To minimize risk R-ACL-001 (drift), the insertion text follows the exact form already proven in `.claude/skills/moai/workflows/design.md:27-30`. Both new sentences use the structure:

```
CI guards in `internal/template/agentless_audit_test.go` enforce the literal `<SENTINEL>` sentinel remains present in this skill body (<REQ-references>, shared with `design.md`).
```

Variables: `<SENTINEL>` = `MODE_PIPELINE_ONLY_UTILITY` (for plan.md, sync.md) or `MODE_UNKNOWN` (for run.md). `<REQ-references>` matches each sentinel's owning SPEC requirement IDs.

## 2. Verification Commands

Per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution, the following 6-command batch runs as a single parallel orchestrator turn:

```bash
# 1. Three sentinel tests (combined regex)
go test -count=1 -run "TestImplementationSkillsContainPipelineRejectionSentinel|TestRunDesignSkillsContainModeUnknownSentinel|TestAllAgentsInCatalog" ./internal/template/...

# 2. Full template package suite (catch cascading failures)
go test -count=1 ./internal/template/...

# 3. Sentinel literal presence (positive — must return >= 1)
grep -c "MODE_PIPELINE_ONLY_UTILITY" .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/sync.md
grep -c "MODE_UNKNOWN" .claude/skills/moai/workflows/run.md

# 4. Sentinel inventory (negative — REQ-ACL-007: no new MODE_* introduced)
grep -hoE "MODE_[A-Z_]+" .claude/skills/moai/workflows/*.md | sort -u

# 5. Test file integrity (negative — REQ-ACL-006: agentless_audit_test.go unchanged)
git diff --stat internal/template/agentless_audit_test.go

# 6. Cross-platform build (Section B1)
GOOS=windows GOARCH=amd64 go build ./...
```

Expected outcomes:
- (1): `ok` with `--- PASS` for all 3 tests
- (2): `ok` across package (no cascading regression)
- (3): each `grep -c` returns `1` or higher; aggregate ≥ 3 hits
- (4): output is exactly `MODE_PIPELINE_ONLY_UTILITY` and `MODE_UNKNOWN` (two lines, no third)
- (5): empty output (no modifications)
- (6): exit 0

## 3. Commit Plan

Single fix commit on `main` (Late-Branch per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005, no branch creation at plan-phase or run-phase):

```
fix(SPEC-V3R6-ABSORB-CLEANUP-001): restore baseline sentinels + reconcile agent count

3 baseline test failures on main HEAD 7892b412b verified zero-new-regression
during SPEC-V3R6-CATALOG-SSOT-001 run-phase. Wave 1 sync-phase batch PR
requires clean baseline.

Changes:
- .claude/skills/moai/workflows/plan.md: add MODE_PIPELINE_ONLY_UTILITY sentinel sentence
- .claude/skills/moai/workflows/sync.md: add MODE_PIPELINE_ONLY_UTILITY sentinel sentence
- .claude/skills/moai/workflows/run.md: add MODE_UNKNOWN sentinel sentence
- internal/template/catalog_tier_audit_test.go: expectedAgentCount 20 → 19,
  update breakdown comment (my-harness category removed from disk post-Bundle C)

Verification:
- TestImplementationSkillsContainPipelineRejectionSentinel: PASS (3 subtests)
- TestRunDesignSkillsContainModeUnknownSentinel: PASS (2 subtests)
- TestAllAgentsInCatalog: PASS
- internal/template/... full suite: PASS

🗿 MoAI <email@mo.ai.kr>
```

A separate plan-phase chore commit precedes this fix commit:

```
plan(SPEC-V3R6-ABSORB-CLEANUP-001): Wave 1 foundation cleanup (Tier S)

Tier S LEAN spec + plan (2 artifacts):
- 7 EARS REQs (5 Ubiquitous + 2 Unwanted)
- 7 binary ACs (single-command verifiable)
- 5 Risks + 7 Exclusions + REQ↔AC traceability

Scope: resolve 3 baseline test failures verified zero-new-regression
during CATALOG-SSOT-001 — sentinel restoration (plan.md/sync.md/run.md)
+ catalog reconciliation (expectedAgentCount 20→19).

4 files affected, within Tier S budget (≤5).

depends_on: SPEC-V3R6-CATALOG-SSOT-001

🗿 MoAI <email@mo.ai.kr>
```

## 4. Rollback Procedure

If verification fails post-fix-commit on local main (Late-Branch — no remote push yet):

```bash
# Identify the fix commit (most recent)
git log --oneline -3

# Soft reset to drop the fix commit while keeping plan commit
git reset --soft HEAD~1

# Or, if both plan + fix need to be undone:
git reset --hard <SHA-before-plan-commit>
```

Recovery does not affect any pushed branches (Late-Branch keeps everything local until sync-phase). SPEC artifacts remain in working tree for re-edit.

## 5. Brownfield Strategy

### PRESERVE list (MUST NOT modify)

Runtime-managed:
- `.moai/harness/usage-log.jsonl` (per §B8)

Parallel session workstream (completely unrelated, currently dirty):
- `internal/statusline/memory.go`
- `internal/statusline/renderer.go`
- `internal/statusline/renderer_test.go`
- `internal/statusline/stdinfields_test.go`

§B7 capture-path artifact (deferred per EXCL-ACL-005):
- `internal/hook/.moai/` (entire subtree — out of scope this SPEC)

PR #1037 legitimate work product (NOT residuals per EXCL-ACL-007):
- `.claude/commands/99-release.md`
- `.claude/skills/moai/workflows/release.md`
- `internal/cli/init_layout.go`
- `internal/cli/wizard/fullscreen.go`
- `internal/cli/wizard/review.go`

Test contract (REQ-ACL-006):
- `internal/template/agentless_audit_test.go` (sentinel detection logic frozen — only `catalog_tier_audit_test.go` constant is edited)

Other workflow skills not named in §1.1:
- `.claude/skills/moai/workflows/design.md` (already has both sentinels — no change)
- All other `.claude/skills/moai/workflows/*.md` files

Other catalog entries:
- `internal/template/catalog.yaml` is NOT modified — only the test constant guarding it changes.

Blueprint:
- `.moai/research/v3-redesign-blueprint-2026-05-22.md` (FROZEN per §0 USER-CONFIRMED)

### EXTEND list (MAY modify)

Only the 4 files in §1.1.

## 6. Risk Mitigation Reference

| Risk | Mitigation | Verification |
|------|-----------|--------------|
| R-ACL-001 | Use canonical sentinel sentence template from §1.2 | Sentence body matches `design.md:27-30` structure |
| R-ACL-002 | Comment under `expectedAgentCount` enumerates each category | `grep -A2 expectedAgentCount internal/template/catalog_tier_audit_test.go` shows breakdown |
| R-ACL-003 | No new heading; insertion is mid-paragraph | Verification §2 step 2 (full package test) catches markdown-lint regressions |
| R-ACL-004 | Pre-flight `make build && go test ./internal/template/...` confirms templating path | Run before declaring complete |
| R-ACL-005 | Single fix commit + plan commit, both local main | `git log --oneline origin/main..HEAD` shows expected 2 new commits atop CATALOG-SSOT-001's commits |
