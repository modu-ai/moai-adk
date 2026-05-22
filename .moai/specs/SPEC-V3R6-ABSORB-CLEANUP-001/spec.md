---
id: SPEC-V3R6-ABSORB-CLEANUP-001
title: "Wave 1 Foundation Cleanup — baseline sentinel insertion + catalog reconciliation"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: GOOS Kim
priority: P1
phase: "v3.0.0"
module: ".claude/skills, internal/template"
lifecycle: spec-anchored
tier: S
tags: "cleanup, baseline, sentinel, catalog, ssot, v3, wave1"
depends_on: ["SPEC-V3R6-CATALOG-SSOT-001"]
---

# SPEC-V3R6-ABSORB-CLEANUP-001: Wave 1 Foundation Cleanup

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0   | 2026-05-22 | GOOS Kim | Initial draft — resolve 3 baseline test failures verified zero-new-regression during CATALOG-SSOT-001 |

## 1. Problem Statement

Three test failures persist on `main` HEAD `7892b412b` that were **verified zero-new-regression** during SPEC-V3R6-CATALOG-SSOT-001 run-phase (i.e., they were already failing on `origin/main` before that SPEC). They represent legitimate baseline drift between source-of-truth files and the assertions that guard them:

1. **`TestImplementationSkillsContainPipelineRejectionSentinel`** (`internal/template/agentless_audit_test.go:169`) — asserts the literal string `MODE_PIPELINE_ONLY_UTILITY` is present in `.claude/skills/moai/workflows/plan.md` and `.claude/skills/moai/workflows/sync.md`. Currently absent in both. The sentinel **is** present in `.claude/skills/moai/workflows/design.md:30` per REQ-WF003-016 ↔ REQ-WF004-014, but the test expects the same enforcement guard in `plan.md` and `sync.md`.

2. **`TestRunDesignSkillsContainModeUnknownSentinel`** (`internal/template/agentless_audit_test.go:211`) — asserts the literal string `MODE_UNKNOWN` is present in `.claude/skills/moai/workflows/run.md` and `.claude/skills/moai/workflows/design.md`. Currently present in `design.md:29` but **absent in `run.md`**.

3. **`TestAllAgentsInCatalog`** (`internal/template/catalog_tier_audit_test.go:198`) — asserts disk agent count equals `expectedAgentCount = 20`. Disk currently has 19 agents (`builder-harness, claude-code-guide, evaluator-active, expert-{backend,devops,frontend,performance,refactoring,security}, manager-{brain,develop,docs,git,project,quality,spec,strategy}, plan-auditor, researcher`). Comment claims "14 active system + 4 my-harness + 2 evaluator-family" = 20, but no `my-harness-*` agents exist on disk — comment is stale post-Bundle C purge (8 zombie agents removed per Workflow audit 2026-05-16). The constant must be reconciled with disk reality.

These three failures must be resolved on `main` before Wave 1 sync-phase batch PR so that the merged baseline is clean.

## 2. Goals

- G1: Restore sentinel coverage so `plan.md`, `run.md`, `sync.md` all satisfy their CI guard tests.
- G2: Reconcile `expectedAgentCount` with disk truth (19) and update the explanatory comment to reflect actual category breakdown.
- G3: Stay within Tier S budget (≤5 files affected, single PR) to enable inclusion in Wave 1 sync-phase batch.

## 3. EARS Requirements

- REQ-ACL-001 (Ubiquitous): The system shall ensure that `.claude/skills/moai/workflows/plan.md` contains the literal string `MODE_PIPELINE_ONLY_UTILITY` within a CI-guard sentence that explicitly references `internal/template/agentless_audit_test.go`.

- REQ-ACL-002 (Ubiquitous): The system shall ensure that `.claude/skills/moai/workflows/sync.md` contains the literal string `MODE_PIPELINE_ONLY_UTILITY` within a CI-guard sentence that explicitly references `internal/template/agentless_audit_test.go`.

- REQ-ACL-003 (Ubiquitous): The system shall ensure that `.claude/skills/moai/workflows/run.md` contains the literal string `MODE_UNKNOWN` within a CI-guard sentence that explicitly references `internal/template/agentless_audit_test.go`.

- REQ-ACL-004 (Ubiquitous): The system shall declare `expectedAgentCount = 19` in `internal/template/catalog_tier_audit_test.go` to match the actual count of `.md` files under `.claude/agents/moai/`.

- REQ-ACL-005 (Ubiquitous): The system shall update the explanatory comment preceding `expectedAgentCount` in `internal/template/catalog_tier_audit_test.go` to accurately describe the 19 agents by category (e.g., "8 manager + 6 expert + 1 builder + 1 evaluator + 1 plan-auditor + 1 researcher + 1 claude-code-guide = 19").

- REQ-ACL-006 (Unwanted): The system shall not modify the sentinel detection logic in `agentless_audit_test.go` (test contract is fixed; only the source files under test are edited).

- REQ-ACL-007 (Unwanted): The system shall not add any new `MODE_*` sentinel categories beyond the two already enforced (`MODE_PIPELINE_ONLY_UTILITY`, `MODE_UNKNOWN`); scope is restoration, not extension.

## 4. Binary Acceptance Criteria

| AC | Verification Command | Expected Output |
|----|---------------------|-----------------|
| AC-ACL-001 | `grep -c "MODE_PIPELINE_ONLY_UTILITY" .claude/skills/moai/workflows/plan.md` | `1` or higher (literal present) |
| AC-ACL-002 | `grep -c "MODE_PIPELINE_ONLY_UTILITY" .claude/skills/moai/workflows/sync.md` | `1` or higher (literal present) |
| AC-ACL-003 | `grep -c "MODE_UNKNOWN" .claude/skills/moai/workflows/run.md` | `1` or higher (literal present) |
| AC-ACL-004 | `grep -E "^[[:space:]]*const expectedAgentCount = 19$" internal/template/catalog_tier_audit_test.go \| wc -l` | `1` (exact constant declaration) |
| AC-ACL-005 | `go test -count=1 -run "TestImplementationSkillsContainPipelineRejectionSentinel" ./internal/template/...` | `ok` (PASS, no FAIL lines) |
| AC-ACL-006 | `go test -count=1 -run "TestRunDesignSkillsContainModeUnknownSentinel" ./internal/template/...` | `ok` (PASS, no FAIL lines) |
| AC-ACL-007 | `go test -count=1 -run "TestAllAgentsInCatalog" ./internal/template/...` | `ok` (PASS, no FAIL lines) |

## 5. Out of Scope

### 5.1 Out of Scope

- EXCL-ACL-001: GEARS migration is out of scope (deferred to `SPEC-V3R6-GEARS-MIGRATION-001` in Wave 6 per v3-redesign-blueprint-2026-05-22.md update).
- EXCL-ACL-002: Wave 2 folder restructure SPECs (HARNESS-RENAME-001, AGENT-FOLDER-SPLIT-001) are not touched.
- EXCL-ACL-003: Statusline parallel session workstream files (`internal/statusline/memory.go`, `renderer.go`, `renderer_test.go`, `stdinfields_test.go` — currently dirty in working tree) are preserved verbatim.
- EXCL-ACL-004: Runtime-managed `.moai/harness/usage-log.jsonl` is not modified (per §B8 working tree hygiene rule).
- EXCL-ACL-005: `internal/hook/.moai/` capture-path artifact cleanup (§B7 observer.go working-dir leak) is deferred to a future Tier S SPEC; this SPEC's 5-file budget is reserved for the 4 sentinel/catalog fixes.
- EXCL-ACL-006: Adding new agents to disk (alternative resolution to AC-ACL-004) is out of scope — design/scope of any new agent belongs to a dedicated SPEC.
- EXCL-ACL-007: PR #1037 absorbed implementation files (`.claude/commands/99-release.md`, `.claude/skills/moai/workflows/release.md`, `internal/cli/init_layout.go`, `internal/cli/wizard/{fullscreen,review}.go`) are confirmed legitimate work product — not residuals to clean.

## 6. Risks

- R-ACL-001 (Low / Low impact): Sentinel sentence wording in `plan.md`/`sync.md`/`run.md` could drift from the established `design.md` pattern. Mitigation: copy the exact form `CI guards in internal/template/agentless_audit_test.go enforce the literal MODE_XXX sentinel remains present in this skill body.` to ensure stylistic consistency.

- R-ACL-002 (Medium / Low impact): Lowering `expectedAgentCount` from 20 to 19 could mask a future agent-addition regression. Mitigation: the explanatory comment under REQ-ACL-005 enumerates each agent category, making any future drift (e.g., adding 1 agent without updating constant) obvious during code review.

- R-ACL-003 (Low / Low impact): `plan.md`/`sync.md` text addition could trigger ancillary spec-lint or markdown-lint checks beyond the sentinel tests. Mitigation: addition is a single sentence inserted in existing prose paragraph; no new section headings.

- R-ACL-004 (Low / Medium impact): If `.claude/skills/moai/workflows/{plan,run,sync}.md` are templates synced via `make build`, the embedded copy (`internal/template/embedded.go`) must also be regenerated. Mitigation: skill files in `.claude/skills/moai/workflows/` are NOT under `internal/template/templates/` — they are project-local consumed directly by test FS walker; `make build` is **not** required (verified in pre-flight: agentless_audit_test reads from `EmbeddedTemplates()` but the test source list in §1 points to project-relative paths via `fs.ReadFile`; the embedded FS path `.claude/skills/moai/workflows/` is resolved through `internal/template/templates/` mirroring). Pre-flight will confirm by running `make build && go test ./internal/template/...` once.

- R-ACL-005 (Low / Low impact): Late-Branch workflow per REQ-LB-005 means this SPEC commits land on local `main` and accumulate alongside CATALOG-SSOT-001 (`617b4a76a + 7892b412b`) for batch sync. Mitigation: single fix commit with explicit `🗿 MoAI` trailer; sync-phase Phase C handles branch creation + cherry-pick.

## 7. REQ ↔ AC Traceability

| REQ | AC | Status |
|-----|----|--------|
| REQ-ACL-001 | AC-ACL-001, AC-ACL-005 | covered |
| REQ-ACL-002 | AC-ACL-002, AC-ACL-005 | covered |
| REQ-ACL-003 | AC-ACL-003, AC-ACL-006 | covered |
| REQ-ACL-004 | AC-ACL-004, AC-ACL-007 | covered |
| REQ-ACL-005 | AC-ACL-007 (comment update verified implicitly via test PASS + manual review) | covered |
| REQ-ACL-006 | (negative — verified by `git diff agentless_audit_test.go = empty`) | covered (verification step in plan §2) |
| REQ-ACL-007 | (negative — verified by `grep -E "MODE_[A-Z_]+" .claude/skills/moai/workflows/*.md` returning only the 2 sanctioned sentinels) | covered (verification step in plan §2) |
