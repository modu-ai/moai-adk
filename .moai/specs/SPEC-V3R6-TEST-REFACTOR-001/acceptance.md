---
id: SPEC-V3R6-TEST-REFACTOR-001
title: "Go test suite refactor — acceptance criteria"
version: "0.1.3"
status: completed
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0 follow-up"
module: "internal/template, internal/skills, internal/harness, internal/statusline"
lifecycle: spec-anchored
tags: "test-refactor, acceptance, atr-001-debt-discharge"
depends_on: [SPEC-V3R6-AGENT-TEAM-REBUILD-001]
tier: M
---

# SPEC-V3R6-TEST-REFACTOR-001 — Acceptance Criteria

## Section A — AC Matrix

All 14 AC rows below are MUST-PASS for Definition of Done (spec.md §D #1).

| # | AC ID | Severity | Description | REQ refs | Milestone |
|---|-------|----------|-------------|----------|-----------|
| 1 | AC-TST-001 | MUST | `go test ./...` returns exit code 0 with zero `FAIL` lines | REQ-TST-001 | M6 |
| 2 | AC-TST-002 | MUST | TestRetirementCompletenessAssertion passes — path literal updated from `.claude/agents/moai/` to `.claude/agents/core/` per AGENT-FOLDER-SPLIT-001 | REQ-TST-001, REQ-TST-010 | M2 |
| 3 | AC-TST-003 | MUST | All 9 architectural-pivot-consequence tests in `internal/template` pass (TestContractSchemaVerification, TestBackwardCompatibility, TestContractAssertionsNaturalLanguage, TestAgentFrontmatterAudit, TestTemplateAgentsStructure, TestEmbeddedTemplates_AgentDefinitions, TestLoadCatalog, TestAllAgentsInCatalog, TestLoadEmbeddedCatalog_Success) | REQ-TST-001, REQ-TST-011 | M2 |
| 4 | AC-TST-004 | MUST | TestRuleTemplateMirrorDrift passes — mirror parity expected file list updated to current `internal/template/templates/.claude/rules/moai/**` reality | REQ-TST-001 | M2 |
| 5 | AC-TST-005 | MUST | TestTemplateMirrorParity passes — `internal/skills` skill mirror parity reconciled | REQ-TST-001 | M3 |
| 6 | AC-TST-006 | MUST | TestSubSkillLOCCeiling passes — classified at M3 entry as (a) test fixture threshold drift OR (b) escalation blocker; if (b), return blocker report instead of expanding scope | REQ-TST-001, REQ-TST-004 | M3 |
| 7 | AC-TST-007 | MUST | TestSubagentBoundary_NoAskUserQuestion passes — scan paths extended to include ATR-001's 3 new hook scripts (status-transition-ownership.sh, sync-phase-quality-gate.sh, team-ac-verify.sh) | REQ-TST-001, REQ-TST-011 | M4 |
| 8 | AC-TST-008 | MUST | TestRenderPRSegment_Absence passes — pre-existing classification verified via git blame; fix annotated `(pre-existing per ATR-001 §F.2.8)` in commit body | REQ-TST-001, REQ-TST-010 | M5 |
| 9 | AC-TST-009 | MUST | `golangci-lint run --timeout=2m` returns no NEW lint issues compared to baseline at HEAD `e7b119924` | REQ-TST-001 | M6 |
| 10 | AC-TST-010 | MUST | No predecessor SPEC body modified — `git diff origin/main -- .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/ .moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/` returns empty after 4-phase close | REQ-TST-003, HARD-2 | M6 |
| 11 | AC-TST-011 | MUST | Every milestone commit's `git diff --cached --name-only` matches the milestone's declared file scope per plan.md §A — no cross-milestone scope bleed | REQ-TST-009, HARD-1 | M1–M6 |
| 12 | AC-TST-012 | MUST | Every milestone commit's pre-commit and post-push `git rev-list --count --left-right origin/main...HEAD` returns `0 0` (or documents L52 race absorption with verbatim push range) | REQ-TST-009, HARD-4 | M1–M6 |
| 13 | AC-TST-013 | MUST | No `t.Skip()` or `testing.Short()` added to mask failures; no failing test deleted without REQ-TST-013 obsolescence justification + retained-catalog assertion replacement | REQ-TST-012, REQ-TST-013 | M2–M5 |
| 14 | AC-TST-014 | MUST | All catalog edits routed through `make build` — `internal/template/embedded.go` and `internal/template/catalog.go`-generated artifacts MUST NOT be hand-edited; verified by checking commits include corresponding `make build` regeneration evidence | REQ-TST-005, REQ-TST-007, HARD-3 | M2 |

## Section B — Detailed AC bodies

### AC-TST-001 — `go test ./...` exit 0

**Verification command**:
```bash
go test ./...
echo "exit code: $?"
```

**Expected output pattern**:
- Final line includes `ok` or empty for all packages
- No `FAIL` line anywhere in stdout
- exit code: 0

**Negative test**: Currently fails (15 FAIL lines at HEAD `e7b119924`).

---

### AC-TST-002 — TestRetirementCompletenessAssertion path-drift fix

**Verification command**:
```bash
go test -run TestRetirementCompletenessAssertion ./internal/template/...
```

**Expected output pattern**:
- `PASS` or `ok github.com/modu-ai/moai-adk/internal/template`
- All subtests including `manager-tdd_replacement_manager-develop_must_exist` and `manager-ddd_replacement_manager-develop_must_exist` PASS

**Evidence**: `grep -rn ".claude/agents/moai/manager-develop.md" internal/template/` returns 0 matches (literal updated to `.claude/agents/core/manager-develop.md`).

**Commit annotation requirement**: Commit body MUST include `(pre-existing per ATR-001 §F.2.8)` per REQ-TST-010.

---

### AC-TST-003 — 9 architectural-pivot consequences in `internal/template`

**Verification command**:
```bash
go test -run "TestContractSchemaVerification|TestBackwardCompatibility|TestContractAssertionsNaturalLanguage|TestAgentFrontmatterAudit|TestTemplateAgentsStructure|TestEmbeddedTemplates_AgentDefinitions|TestLoadCatalog|TestAllAgentsInCatalog|TestLoadEmbeddedCatalog_Success" ./internal/template/...
```

**Expected output pattern**:
- 9 separate `--- PASS:` lines
- Package-level `ok github.com/modu-ai/moai-adk/internal/template`

**Evidence**:
- Test fixtures / expected-agent-count constants updated from 17 → 7 retained MoAI-custom agents
- Archived-agent enumerations replaced with retained-catalog equivalents per REQ-TST-013
- Commit body references ATR-001 archive list per REQ-TST-011

---

### AC-TST-004 — TestRuleTemplateMirrorDrift NEW failure resolution

**Verification command**:
```bash
go test -run TestRuleTemplateMirrorDrift ./internal/template/...
```

**Expected output pattern**:
- `--- PASS: TestRuleTemplateMirrorDrift`

**Evidence**: Mirror parity expected file list updated to current reality. Investigated at M2 entry — likely cascade from post-ATR-001 template rule churn (NOTICE.md / agent-common-protocol.md / spec-frontmatter-schema.md / archived-agent-rejection.md / orchestration-mode-selection.md / CLAUDE.md / 3 new hook scripts).

---

### AC-TST-005 — TestTemplateMirrorParity

**Verification command**:
```bash
go test -run TestTemplateMirrorParity ./internal/skills/...
```

**Expected output pattern**:
- `--- PASS: TestTemplateMirrorParity`

**Evidence**: Skill mirror parity expected list reconciled with current `internal/template/templates/.claude/skills/**`. No regression of LNCO-001 `.claude/agents/local/` namespace protection per REQ-TST-002.

---

### AC-TST-006 — TestSubSkillLOCCeiling

**Verification command**:
```bash
go test -run TestSubSkillLOCCeiling ./internal/skills/...
```

**Expected output pattern**:
- `--- PASS: TestSubSkillLOCCeiling`

**Evidence**: Classified at M3 entry. If (a) test fixture threshold drift, fix is test-only per HARD-1. If (b) real skill body LOC violation, return blocker report and propose follow-up SPEC — do NOT expand this SPEC's scope per REQ-TST-004.

---

### AC-TST-007 — TestSubagentBoundary_NoAskUserQuestion

**Verification command**:
```bash
go test -run TestSubagentBoundary_NoAskUserQuestion ./internal/harness/...
```

**Plus orchestrator-side independent verification**:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ internal/template/templates/.claude/hooks/moai/ \
  | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" | grep -v "^[^:]*:[0-9]*:[ \t]*#"
```

**Expected output pattern**:
- `--- PASS: TestSubagentBoundary_NoAskUserQuestion`
- grep returns 0 matches (3 new ATR-001 hook scripts comply with C-HRA-008 sentinel per REQ-TST-011)

**Evidence**: Test scan paths extended to include ATR-001 hook directory. Commit body references ATR-001 anchor per REQ-TST-011.

---

### AC-TST-008 — TestRenderPRSegment_Absence (pre-existing)

**Verification command**:
```bash
go test -run TestRenderPRSegment_Absence ./internal/statusline/...
```

**Expected output pattern**:
- `--- PASS: TestRenderPRSegment_Absence`

**Pre-existing classification verification**:
```bash
git log --oneline -p internal/statusline/renderer_test.go | head -50
```

Verify the failing assertion predates ATR-001 plan-phase commit `b957a4d04`.

**Commit annotation requirement**: Commit body MUST include `(pre-existing per ATR-001 §F.2.8)` per REQ-TST-010.

---

### AC-TST-009 — Lint baseline preservation

**Verification command**:
```bash
golangci-lint run --timeout=2m
```

**Expected output pattern**:
- Zero NEW issues compared to baseline at HEAD `e7b119924`. Pre-existing lint debt may remain (documented as out-of-scope per SHOULD-3).

**Evidence**: Lint output compared milestone-by-milestone. If a milestone introduces NEW lint issues, fix before commit per plan.md §B M6 risk mitigation.

---

### AC-TST-010 — Predecessor SPEC body preservation (L48 SSOT)

**Verification command** (run at M6 exit):
```bash
git diff origin/main -- \
  .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ \
  .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/ \
  .moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/ \
  .moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/
```

**Expected output**: Empty (no diff).

**Evidence**: L48 SSOT preservation absolute throughout this SPEC's 4-phase lifecycle. Any predecessor SPEC body diff is a HARD-2 violation requiring revert.

---

### AC-TST-011 — Per-milestone path-specific staging (L46)

**Verification command** (run at every milestone commit, pre-commit):
```bash
git diff --cached --name-only
```

**Expected output pattern**: Files match milestone's declared scope per plan.md §A.

**Per-milestone expected scopes**:
- M1: 4 SPEC artifact files only
- M2: `internal/template/*_test.go` (4–8 files) + optionally `internal/template/embedded.go` + catalog generated artifacts (via `make build` only)
- M3: `internal/skills/*_test.go` (2 files)
- M4: `internal/harness/*_test.go` (1 file)
- M5: `internal/statusline/renderer_test.go` (1 file)
- M6: `CHANGELOG.md` + 4 SPEC artifact frontmatter + HISTORY

**No cross-milestone scope bleed** per HARD-1.

---

### AC-TST-012 — Pre/post fetch HARD discipline (L44 38x)

**Verification command** (run pre-commit and post-push at every milestone):
```bash
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
```

**Expected output**: `0 0`

**L52 race absorption documentation**: If a parallel session commit lands between fetch and push, document the push range in the commit body (e.g., `(L52 case 27 absorbed: <other-commit-sha> origin-ahead at <timestamp>, scope-disjoint clean FF)`).

---

### AC-TST-013 — No skip / no delete

**Verification commands**:
```bash
git diff origin/main -- internal/ | grep -E "^\+[ \t]*(t\.Skip|testing\.Short)" | head -10
```

**Expected output**: Empty.

```bash
git diff origin/main -- internal/ | grep -E "^-[ \t]*func Test" | head -10
```

**Expected output**: If non-empty, each deleted test MUST have a corresponding entry in the commit body justifying obsolescence per REQ-TST-013 (e.g., "TestArchivedAgentX deleted because subject was archived per ATR-001 archive list; replaced with TestRetainedAgentX equivalent assertion").

---

### AC-TST-014 — Catalog regen via `make build` only

**Verification commands**:
```bash
# Inspect catalog regen evidence in commit
git log --oneline --all --grep="make build" -10
```

```bash
# Inspect generated artifacts for hand-edit signatures
git diff origin/main -- internal/template/embedded.go | head -50
```

**Expected output**:
- If `internal/template/embedded.go` is modified in any milestone, the commit body MUST include a `make build` execution note (e.g., "Catalog regenerated via `make build` post-edit of catalog.yaml").
- If catalog.yaml is modified, embedded.go and catalog.go-generated artifacts MUST be regenerated in the same commit via `make build`.
- No partial regeneration (catalog.yaml without embedded.go refresh) allowed.

## Section C — REQ ↔ AC Bidirectional Traceability

### REQ → AC

| REQ ID | Type | AC refs |
|--------|------|---------|
| REQ-TST-001 | Ubiquitous | AC-TST-001, AC-TST-002, AC-TST-003, AC-TST-004, AC-TST-005, AC-TST-006, AC-TST-007, AC-TST-008, AC-TST-009 |
| REQ-TST-002 | Ubiquitous | AC-TST-005 (LNCO-001 namespace), AC-TST-007 (boundary preservation) |
| REQ-TST-003 | Ubiquitous | AC-TST-010 |
| REQ-TST-004 | Ubiquitous | AC-TST-006 (scope discipline at TestSubSkillLOCCeiling escalation) |
| REQ-TST-005 | Event-driven | AC-TST-014 |
| REQ-TST-006 | Event-driven | AC-TST-001 (7-item batch at M6) |
| REQ-TST-007 | Event-driven | AC-TST-014 |
| REQ-TST-008 | State-driven | AC-TST-010, AC-TST-011 |
| REQ-TST-009 | State-driven | AC-TST-011, AC-TST-012 |
| REQ-TST-010 | Where-capability | AC-TST-002 (pre-existing annotation), AC-TST-008 (pre-existing annotation) |
| REQ-TST-011 | Where-capability | AC-TST-003 (ATR-001 anchor in commit body), AC-TST-007 (ATR-001 anchor) |
| REQ-TST-012 | Unwanted | AC-TST-013 (no skip / no Short) |
| REQ-TST-013 | Unwanted | AC-TST-013 (no delete without justification) |

### AC → REQ

| AC ID | REQ refs |
|-------|----------|
| AC-TST-001 | REQ-TST-001, REQ-TST-006 |
| AC-TST-002 | REQ-TST-001, REQ-TST-010 |
| AC-TST-003 | REQ-TST-001, REQ-TST-011 |
| AC-TST-004 | REQ-TST-001 |
| AC-TST-005 | REQ-TST-001, REQ-TST-002 |
| AC-TST-006 | REQ-TST-001, REQ-TST-004 |
| AC-TST-007 | REQ-TST-001, REQ-TST-002, REQ-TST-011 |
| AC-TST-008 | REQ-TST-001, REQ-TST-010 |
| AC-TST-009 | REQ-TST-001 |
| AC-TST-010 | REQ-TST-003, REQ-TST-008 |
| AC-TST-011 | REQ-TST-008, REQ-TST-009 |
| AC-TST-012 | REQ-TST-009 |
| AC-TST-013 | REQ-TST-012, REQ-TST-013 |
| AC-TST-014 | REQ-TST-005, REQ-TST-007 |

**Coverage**: Every REQ has ≥1 AC, every AC has ≥1 REQ. 100% bidirectional traceability.

## Section D — Definition of Done

Per spec.md §D, this SPEC is COMPLETE when all 8 conditions hold simultaneously:

1. ✅ All 14 MUST-PASS AC rows verified PASS (AC-TST-001 through AC-TST-014)
2. ✅ `go test ./...` exit code 0
3. ✅ `golangci-lint run --timeout=2m` no NEW issues vs baseline at HEAD `e7b119924`
4. ✅ CHANGELOG.md `[Unreleased]` entry summarizes the discharge
5. ✅ Frontmatter status: `draft → in-progress → implemented` (sync may further transition to `completed`)
6. ✅ Every milestone commit's staging matches milestone scope (L46 path-specific)
7. ✅ Mx-phase Step C executes + 4-phase close marker emitted
8. ✅ Predecessor SPEC bodies unmodified (AC-TST-010 verifies)

## HISTORY

### v0.1.3 (2026-05-25) — 4-phase close terminator

- Frontmatter status transition: `implemented → completed` (orchestrator-direct chore — 4-phase close marker completion).
- Version bump 0.1.2 → 0.1.3.

### v0.1.1 (2026-05-25) — sync-phase status:in-progress → implemented

- Sync-phase status transition: `in-progress → implemented` per Status Transition Ownership Matrix (manager-docs owned).
- AC matrix verification: all 14 MUST-PASS AC rows verified PASS (progress.md §E.2).
- 100% REQ↔AC bidirectional traceability preserved.

### v0.1.0 (2026-05-25) — initial draft

- 14 MUST-PASS AC matrix authored covering all 15 baseline failures (TestRenderPRSegment_Absence is row 4 of §A.4 mapped to AC-TST-008; the other 14 baseline failures map to AC-TST-002, AC-TST-003 (9 tests), AC-TST-004, AC-TST-005, AC-TST-006, AC-TST-007 = 13 tests + AC-TST-008 = 14 tests; the 15th coverage is AC-TST-001 aggregate `go test ./...` exit 0).
- 100% REQ↔AC bidirectional traceability across 13 REQ + 14 AC.
- Per-AC verification commands + expected output patterns + evidence requirements.
- Definition of Done explicitly enumerates 8 conditions.
