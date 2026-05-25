---
id: SPEC-V3R6-TEST-REFACTOR-001
title: "Go test suite refactor — implementation plan"
version: "0.1.3"
status: completed
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0 follow-up"
module: "internal/template, internal/skills, internal/harness, internal/statusline"
lifecycle: spec-anchored
tags: "test-refactor, plan, atr-001-debt-discharge"
depends_on: [SPEC-V3R6-AGENT-TEAM-REBUILD-001]
tier: M
---

# SPEC-V3R6-TEST-REFACTOR-001 — Implementation Plan

## Section A — Milestone breakdown

The work decomposes into 6 milestones (M1–M6) organized by package + verification. Tier M default delegation: manager-develop receives this plan.md verbatim Section A–E and executes M1–M6 sequentially with a single commit per milestone (SHOULD-1 of spec.md).

### M1 — Ground truth re-measurement + draft→in-progress transition

**Scope**: Single orchestrator-direct edit; no production / test code modification.

**Deliverables**:
- Re-run `go test ./... 2>&1 | grep "^--- FAIL"` and verify exact match against §A.4 (15 rows). If drift detected (e.g., new failures appeared, existing failures resolved, or classification needs revision), return blocker report to orchestrator via L66 working-tree race protocol — do NOT proceed to M2.
- Transition spec.md / plan.md / acceptance.md / progress.md frontmatter `status: draft → in-progress`.
- Update spec.md / plan.md / acceptance.md / progress.md HISTORY with v0.1.0 → v0.1.1 (status transition entry).
- Update progress.md §B (Run-phase milestone log) with M1 entry.

**Files modified** (4 SPEC artifact frontmatter + HISTORY only):
- `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/spec.md` (frontmatter status + HISTORY)
- `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/plan.md` (frontmatter status + HISTORY)
- `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/acceptance.md` (frontmatter status + HISTORY)
- `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/progress.md` (frontmatter status + HISTORY + §B M1 entry)

**Verification command**: `go test ./... 2>&1 | grep -c "^--- FAIL"` returns `15`.

**Risk**: Low — no production code touched. If ground truth drift is detected, halt and return blocker; do not attempt scope expansion in this SPEC.

**Commit message**: `chore(SPEC-V3R6-TEST-REFACTOR-001): M1 ground truth re-measurement + status:in-progress`

---

### M2 — `internal/template` 11 failure fixes (largest milestone)

**Scope**: Address 11 of 15 baseline failures in `internal/template` package. This is the largest milestone by file count and represents the majority of the architectural-pivot debt.

**Failing tests in M2 scope** (11 tests):
- TestContractSchemaVerification (likely in `internal/template/contract_schema_test.go`)
- TestBackwardCompatibility (likely in `internal/template/contract_schema_test.go`)
- TestContractAssertionsNaturalLanguage (likely in `internal/template/contract_schema_test.go`)
- TestAgentFrontmatterAudit (`internal/template/agent_frontmatter_audit_test.go`)
- TestTemplateAgentsStructure
- TestEmbeddedTemplates_AgentDefinitions
- TestLoadCatalog
- TestAllAgentsInCatalog
- TestLoadEmbeddedCatalog_Success
- TestRuleTemplateMirrorDrift
- TestRetirementCompletenessAssertion (pre-existing path drift — `.claude/agents/moai/manager-develop.md` → `.claude/agents/core/manager-develop.md` per AGENT-FOLDER-SPLIT-001)

**Fix strategy** (per-test, in order):

1. **TestRetirementCompletenessAssertion (pre-existing path drift)** — first fix. Update test fixture / expected path literal from `.claude/agents/moai/` to `.claude/agents/core/`. Commit message annotation: `(pre-existing per ATR-001 §F.2.8)` per REQ-TST-010.
2. **TestContractSchemaVerification / TestBackwardCompatibility / TestContractAssertionsNaturalLanguage** — investigate together. ATR-001 M8 reported the first as `/manager-quality.md` subtest specifically; these likely assert on archived agents (manager-quality, manager-brain, etc.) and need updating to retained-agent contract assertions per REQ-TST-013 ("replaced with an equivalent retained-catalog assertion"). Reference ATR-001 archive list (12 archived: manager-strategy, manager-quality, manager-brain, manager-project, claude-code-guide, researcher, expert-backend, expert-frontend, expert-security, expert-devops, expert-performance, expert-refactoring) inline in commit body.
3. **TestAgentFrontmatterAudit / TestTemplateAgentsStructure / TestEmbeddedTemplates_AgentDefinitions** — these likely enumerate expected agents from the catalog. Update enumerations to the 7 retained MoAI-custom agents (manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, evaluator-active, builder-harness) excluding the Anthropic built-in Explore (which has no MoAI file).
4. **TestLoadCatalog / TestAllAgentsInCatalog / TestLoadEmbeddedCatalog_Success** — these load catalog.yaml. If catalog.yaml is up to date (per ATR-001 M8), the test expected-count constants may need updating from 17 → 7 retained. If catalog.yaml needs regen, route through `make build` per HARD-3.
5. **TestRuleTemplateMirrorDrift** — investigate at M2 entry. NEW failure since ATR-001 M8 — likely caused by post-ATR-001 template rule churn (NOTICE.md / agent-common-protocol.md / spec-frontmatter-schema.md / archived-agent-rejection.md / orchestration-mode-selection.md / CLAUDE.md / 3 new hook scripts). Fix: update mirror parity expected file list to match current `internal/template/templates/.claude/rules/moai/**` reality.

**Files modified** (estimated 4–8 test files + 1 catalog regen via `make build`):
- `internal/template/*_test.go` (8 test files maximum)
- `internal/template/embedded.go` (regenerated via `make build` if catalog edited)
- `internal/template/catalog.go`-generated entries (regenerated via `make build`)
- (NO hand-edits of generated artifacts per HARD-1 + HARD-3)

**Verification commands** (M2 exit):
```bash
go test ./internal/template/...
go test -run "TestContractSchemaVerification|TestBackwardCompatibility|TestContractAssertionsNaturalLanguage|TestAgentFrontmatterAudit|TestTemplateAgentsStructure|TestEmbeddedTemplates_AgentDefinitions|TestLoadCatalog|TestAllAgentsInCatalog|TestLoadEmbeddedCatalog_Success|TestRuleTemplateMirrorDrift|TestRetirementCompletenessAssertion" ./internal/template/...
```

Both MUST exit 0.

**Risk**: Medium — largest milestone, touches catalog state. Risk mitigations: (a) route catalog regen through `make build`; (b) verify `make build` produces a deterministic embedded.go (no incidental diff outside expected enumeration changes); (c) M2 commit MUST NOT touch any other package.

**Commit message**: `fix(SPEC-V3R6-TEST-REFACTOR-001): M2 internal/template 11 test fixes (10 architectural-pivot + 1 pre-existing path drift)`

---

### M3 — `internal/skills` 2 failure fixes

**Scope**: 2 tests in `internal/skills` (TestTemplateMirrorParity + TestSubSkillLOCCeiling).

**Fix strategy**:
1. **TestTemplateMirrorParity** — likely a cascade from ATR-001 archive (12 archived agents have NO skills, but if the parity check enumerates expected mirror skills, it needs updating). Investigate at M3 entry. Likely fix: update expected skill mirror list to match current `internal/template/templates/.claude/skills/**` reality.
2. **TestSubSkillLOCCeiling** — needs-investigation per §A.4. Run the test in isolation, inspect failure message, classify as: (a) LOC budget drift in `moai-foundation-*` / `moai-workflow-*` skills (production fix: split skill body per `.claude/rules/moai/development/skill-authoring.md`), or (b) ceiling threshold drift in test fixture (test fix: adjust threshold). Prefer test-fix path per HARD-1 unless production skill is genuinely over budget.

**Files modified** (estimated 2 test files):
- `internal/skills/*_test.go` (2 test files)

**Verification command**:
```bash
go test ./internal/skills/...
```

MUST exit 0.

**Risk**: Low–Medium — TestSubSkillLOCCeiling could escalate to skill body refactor if real LOC budget violation. If escalation needed, return blocker report and propose follow-up SPEC `SPEC-V3R6-SKILL-LOC-COMPRESS-XXX` rather than expanding this SPEC's scope.

**Commit message**: `fix(SPEC-V3R6-TEST-REFACTOR-001): M3 internal/skills 2 test fixes`

---

### M4 — `internal/harness` 1 failure fix

**Scope**: TestSubagentBoundary_NoAskUserQuestion — architectural-pivot consequence from ATR-001 hook directory expansion (3 new hook scripts: status-transition-ownership.sh, sync-phase-quality-gate.sh, team-ac-verify.sh).

**Fix strategy**: The test enforces the C-HRA-008 sentinel that hook scripts MUST NOT invoke `AskUserQuestion` or `mcp__askuser`. The 3 new ATR-001 hook scripts (now in PRESERVE list as `M internal/template/templates/.claude/hooks/moai/*.sh`) need to be added to the test's scanned-paths list, OR the test's scan glob needs updating to include the new hooks directory structure. Inspect test source at M4 entry to determine exact fix.

Verification grep (per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution canonical example):
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ internal/template/templates/.claude/hooks/moai/ \
  | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" | grep -v "^[^:]*:[0-9]*:[ \t]*#"
```

Expected: no matches (3 new hook scripts comply with subagent boundary).

**Files modified** (estimated 1 test file):
- `internal/harness/subagent_boundary_test.go` (or equivalent — locate at M4 entry)

**Verification command**:
```bash
go test ./internal/harness/...
```

MUST exit 0.

**Risk**: Low — test-only fix, no production code modification expected.

**Commit message**: `fix(SPEC-V3R6-TEST-REFACTOR-001): M4 internal/harness subagent-boundary test extension for ATR-001 hook directory`

---

### M5 — `internal/statusline` 1 failure fix

**Scope**: TestRenderPRSegment_Absence at `internal/statusline/renderer_test.go:1429` — classified as pre-existing per ATR-001 §F.2.8.

**Fix strategy**:
1. Verify the pre-existing classification at M5 entry: run `git log -p internal/statusline/renderer_test.go` and confirm the failing assertion predates ATR-001 plan-phase commit `b957a4d04`.
2. If verified pre-existing: fix the test (the failure is unrelated to ATR-001 but still blocks `go test ./...` exit 0 required by DoD #2). Inspect the failure at M5 entry and determine the minimal fix.
3. Commit message annotation: `(pre-existing per ATR-001 §F.2.8)` per REQ-TST-010.

**Files modified** (estimated 1 test file):
- `internal/statusline/renderer_test.go` (line ~1429)

**Verification command**:
```bash
go test ./internal/statusline/...
```

MUST exit 0.

**Risk**: Low — single test, pre-existing classification.

**Commit message**: `fix(SPEC-V3R6-TEST-REFACTOR-001): M5 internal/statusline render-pr-segment-absence (pre-existing per ATR-001 §F.2.8)`

---

### M6 — Verification batch + CHANGELOG + frontmatter status:implemented

**Scope**: Final verification + documentation transition. This milestone is owned by manager-develop for the verification batch + CHANGELOG entry; manager-docs ownership begins at sync-phase.

**Deliverables**:
- Execute the 7-item read-only verification batch (per REQ-TST-006 + `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution):
  1. `go test ./...` exit 0
  2. `go test -coverprofile=cover.out ./internal/template/... ./internal/skills/... ./internal/harness/... ./internal/statusline/...`
  3. `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ internal/template/templates/.claude/hooks/moai/` (sentinel; M4 verification)
  4. `grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ | head -20` (sentinel-key audit)
  5. `go run ./cmd/moai --version` (CLI smoke)
  6. `go test -bench=. -benchmem -run=^$ ./internal/template/...` (optional benchmark, skip if no benchmarks defined)
  7. `golangci-lint run --timeout=2m` (lint baseline — zero NEW issues vs baseline)
- Update CHANGELOG.md `[Unreleased]` entry from stub to final entry summarizing the discharge.
- Transition spec.md / plan.md / acceptance.md / progress.md frontmatter `status: in-progress → implemented` (note: sync-phase manager-docs may further transition to `completed` at Mx-phase per ownership matrix in `.claude/rules/moai/development/spec-frontmatter-schema.md`).

**Files modified**:
- `CHANGELOG.md` (stub → final entry)
- 4 SPEC artifact frontmatter status transitions + HISTORY v0.1.x

**Verification commands**: the 7-item batch above, all green.

**Risk**: Low — verification-only milestone.

**Commit message**: `feat(SPEC-V3R6-TEST-REFACTOR-001): M6 verification batch + CHANGELOG + status:implemented`

## Section B — Risk matrix per milestone

| Milestone | Risk | Likelihood | Impact | Mitigation |
|-----------|------|------------|--------|------------|
| M1 | Ground truth drift detected | Low | Medium | Halt + blocker report; do not expand scope |
| M2 | catalog.yaml hand-edit instead of `make build` | Low | High | HARD-3 enforcement + post-edit verification of embedded.go diff cleanliness |
| M2 | TestRuleTemplateMirrorDrift fix requires broader template rule churn | Medium | Medium | Time-box investigation; if scope creep emerges, return blocker for orchestrator decision |
| M3 | TestSubSkillLOCCeiling escalates to real skill body refactor | Medium | High | If real LOC violation, return blocker; propose follow-up SPEC `SPEC-V3R6-SKILL-LOC-COMPRESS-XXX` |
| M3 | TestTemplateMirrorParity touches `moai update` namespace protection | Low | Medium | Cross-reference §24 CLAUDE.local.md + UNP-001 contract; verify no behavior drift |
| M4 | Subagent-boundary test scan expands to 3 NEW hook scripts | Low | Low | Inspect test source at M4 entry; minimal fix |
| M5 | Pre-existing classification wrong | Low | Low | Verify via git log; if not pre-existing, document classification correction in commit body |
| M6 | golangci-lint NEW issues introduced by fix milestones | Medium | Medium | Run lint after each milestone (M2–M5) before commit; do not defer to M6 |

## Section C — Dependencies

### C.1 Predecessor SPEC dependencies (REFERENCE ONLY — NEVER MODIFY)

- **SPEC-V3R6-AGENT-TEAM-REBUILD-001** (anchor): PROCEED-WITH-DEBT directive at progress.md §F.2.8. This SPEC discharges that directive. ATR-001 spec.md / plan.md / acceptance.md / progress.md MUST remain unmodified throughout this SPEC's 4-phase lifecycle.
- **SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001** (cohort): AC-CHR-008 ORPHAN purge post-condition + REQ-CHR-009 catalog excludes ORPHAN. This SPEC's HARD-3 catalog regen routing aligns with CATALOG-HASH-REGRESSION-CLEANUP-001 mechanism.
- **SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001** (cohort): `.claude/agents/local/` namespace + `moai update` PRESERVE contract. M3 TestTemplateMirrorParity fix MUST NOT regress LNCO-001 namespace protection.
- **SPEC-V3R6-AGENT-FOLDER-SPLIT-001** (anchor for M2 row 15): `.claude/agents/moai/` → `.claude/agents/core/` path split that caused TestRetirementCompletenessAssertion pre-existing failure.

### C.2 Plan-phase outputs feeding run-phase

- spec.md §B requirements REQ-TST-001..013 (run-phase verifies each as part of M6 verification).
- spec.md §A.4 ground truth table (run-phase M1 verifies re-measurement against this snapshot).
- spec.md §D Definition of Done (run-phase M6 verifies all 8 conditions).
- acceptance.md §A AC matrix (run-phase M2–M5 commits each AC PASS evidence).

### C.3 Sync-phase / Mx-phase dependencies

- Sync-phase (manager-docs): CHANGELOG.md final entry from M6 stub → polished form; README update if needed (no user-facing API change expected, so README update likely unnecessary); status `implemented → completed` (optional, ownership per `.claude/rules/moai/development/spec-frontmatter-schema.md`).
- Mx-phase: EVALUATE Step C per coverage delta. Expected `EVALUATE-EXECUTE` per cohort precedent (test-heavy SPEC may exceed coverage delta threshold to trigger MX tag deep-inspection on test files). If actual coverage delta is below threshold, expected `EVALUATE-SKIP` per CATALOG-HASH-REGRESSION-CLEANUP-001 precedent.

## Section D — Validation strategy

### D.1 Per-milestone validation

Each milestone exit MUST satisfy:
1. The milestone's package-level `go test ./internal/<package>/...` returns exit 0.
2. `git diff --cached --name-only` matches the milestone's declared file scope.
3. Pre-commit `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` returns `0 0` (L44 HARD).
4. Post-commit + post-push (when push happens) re-fetch returns `0 0` or documents L52 race absorption.

### D.2 Full verification batch (M6 only)

Per REQ-TST-006 + `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution, the 7-item batch executes as a single orchestrator parallel turn:

```bash
# Single turn, 7 parallel Bash calls
go test ./...                                                                              # (1)
go test -coverprofile=cover.out ./internal/template/... ./internal/skills/... ./internal/harness/... ./internal/statusline/...  # (2)
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ internal/template/templates/.claude/hooks/moai/ | grep -v _test.go | grep -v '^[^:]*:[0-9]*:[ \t]*//' | grep -v '^[^:]*:[0-9]*:[ \t]*#'  # (3)
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ | head -20                            # (4)
go run ./cmd/moai --version                                                                # (5)
go test -bench=. -benchmem -run=^$ ./internal/template/...                                 # (6 optional)
golangci-lint run --timeout=2m                                                             # (7)
```

All 7 MUST satisfy exit conditions documented in spec.md §D.

### D.3 Trust-but-verify discipline (L49)

The orchestrator independently re-runs each milestone's verification command after manager-develop reports completion. Discrepancies between manager-develop's reported state and orchestrator's verified state are surfaced via AskUserQuestion before proceeding to the next milestone.

## Section E — Delegation prompt template for manager-develop

When `/moai run SPEC-V3R6-TEST-REFACTOR-001` is invoked, the orchestrator MUST spawn manager-develop with the following prompt structure:

```
You are manager-develop for SPEC-V3R6-TEST-REFACTOR-001 Tier M run-phase.

Anchor: ATR-001 PROCEED-WITH-DEBT discharge (progress.md §F.2.8 verbatim).
Predecessor SSOT (L48): .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/** and 2 other predecessor SPECs remain UNMODIFIED.

Milestones to execute sequentially (one commit each):
M1: ground truth re-measurement + status:in-progress
M2: internal/template 11 test fixes (largest)
M3: internal/skills 2 test fixes
M4: internal/harness 1 test fix
M5: internal/statusline 1 test fix
M6: verification batch + CHANGELOG + status:implemented

HARD obligations (zero tolerance):
- L44 pre/post fetch HARD on every commit (expect 0 0)
- L46 path-specific staging per milestone (NEVER git add . or git add -A)
- L48 SSOT preservation absolute
- L58 subagent boundary — return structured blocker report instead of AskUserQuestion
- HARD-3 catalog regen via make build only

Verification on each milestone exit:
- Package-level go test exit 0
- git diff --cached --name-only matches milestone scope

Return after each milestone:
- Commit sha
- Pre/post fetch verbatim output
- Staged files list
- Package-level test result (PASS count / FAIL count)
- L52 race signal (if any commit landed between fetch and push)

Halt + blocker conditions (any of): ground truth drift detected at M1 entry; TestSubSkillLOCCeiling escalates to real skill body refactor at M3 entry; any milestone scope expansion beyond plan.
```

## HISTORY

### v0.1.3 (2026-05-25) — 4-phase close terminator

- Frontmatter status transition: `implemented → completed` (orchestrator-direct chore — manager-docs Mx attempt at `23f91adf5` claimed close but did NOT actually transition status).
- Version bump 0.1.2 → 0.1.3 (covers gap in HISTORY v0.1.2 entry).

### v0.1.1 (2026-05-25) — sync-phase status:in-progress → implemented

- Sync-phase status transition: `in-progress → implemented` per Status Transition Ownership Matrix (manager-docs owned).
- CHANGELOG.md replaced plan-phase stub with final-form run-phase discharge narrative.
- All M1–M6 milestones executed; manager-develop verification batch (7 items) PASS; 14/14 AC verified PASS; zero PASS-WITH-DEBT debt.

### v0.1.0 (2026-05-25) — initial draft

- M1–M6 milestone breakdown authored to discharge ATR-001 PROCEED-WITH-DEBT directive.
- M2 identified as largest milestone (11 of 15 test fixes in `internal/template`).
- HARD-1..5 + SHOULD-1..3 inherited from spec.md §C.
- Tier M default delegation (manager-develop receives plan.md verbatim §A–E).
- Hybrid Trunk main-direct routing — no `--pr` flag default (per §23.7 of CLAUDE.local.md + §23.9 Tier-based routing — Tier M = main direct).
