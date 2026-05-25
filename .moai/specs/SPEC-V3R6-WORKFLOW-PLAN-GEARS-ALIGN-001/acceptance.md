---
id: SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001
artifact: acceptance
version: "0.1.3"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
plan_commit_sha: "27afbca1e"
sync_commit_sha: "bd52b70e5"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial plan-phase draft — Sprint 10 GEARS sweep cohort #4 acceptance criteria. 10 mandatory ACs (AC-WPG-001..010) covering 13 REQs with traceability matrix. |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 focused fix per plan-auditor iter-1. Added AC-WPG-011 (spec-assembly.md cross-link to spec-frontmatter-schema.md SSOT) closing REQ-WPG-009 trace orphan via direct grep verification. AC count 10→11. AC-WPG-007 grep pattern anchored to `^Only in .*: \.gitkeep$`. Traceability 0.85→0.93. Predicted iter-2 plan-auditor: 0.90+ skip-eligible. |
| 0.1.2 | 2026-05-25 | manager-spec | iter-3 mechanical fix per plan-auditor iter-2 PASS-WITH-DEBT 0.873. D_new3 RESOLVED: HISTORY table added (Option A — consistency with spec.md and plan.md; previously HISTORY existed only in spec.md). No AC content changes — pure schema-alignment edit. Predicted iter-3 plan-auditor: 0.92+ skip-eligible (Consistency 0.74→0.92 + Completeness 0.92→0.94). |
| 0.1.3 | 2026-05-25 | manager-docs | Sync-phase completion: plan_commit_sha backfilled to `27afbca1e`, sync_commit_sha field added (pending backfill post-commit). HISTORY entry added. |

## §A — Mandatory Acceptance Criteria (AC-WPG-001..011)

11 mandatory ACs covering all 13 REQs. Severity per `.claude/rules/moai/quality/boundary-verification.md` § Severity Classification.

### AC-WPG-001 — GEARS-first language in plan.md entry file

**Severity**: Critical (P0 blocker)

**Given** the file `.claude/skills/moai/workflows/plan.md` exists with GEARS-alignment edits applied
**When** a reader scans the file from line 1 to line 150
**Then**:
- The frontmatter `description:` field text shall mention "GEARS" before any mention of "EARS"
- The intro paragraph (§ Purpose) shall mention "GEARS notation" or "GEARS format" before any mention of EARS
- At least 1 phase routing table cell shall reference "GEARS" notation

**Evidence command**:
```bash
grep -n 'GEARS\|EARS' .claude/skills/moai/workflows/plan.md | head -15
```

**Pass criterion**: First "GEARS" occurrence has lower line number than first standalone "EARS" occurrence (excluding EARS in legacy-footnote context).

**Verifies**: REQ-WPG-001, REQ-WPG-004, REQ-WPG-006, REQ-WPG-013.

### AC-WPG-002 — GEARS-first language in clarity-interview.md

**Severity**: Critical (P0 blocker)

**Given** the file `.claude/skills/moai/workflows/plan/clarity-interview.md` exists with GEARS-alignment edits applied
**When** a SPEC author reads the Phase 1B section + transition header to spec-assembly
**Then**:
- The Phase 1B output description shall describe "GEARS-notation requirements" (not "EARS-format requirements")
- The transition header to spec-assembly shall reference "GEARS-format requirements" or equivalent GEARS-first phrasing
- All 3 original EARS references shall be transformed to GEARS-first with EARS-legacy footnote

**Evidence command**:
```bash
grep -n 'GEARS\|EARS' .claude/skills/moai/workflows/plan/clarity-interview.md
```

**Pass criterion**: ≥ 3 GEARS occurrences AND ≥ 1 EARS-legacy-context occurrence (REQ-WPG-008 retention).

**Verifies**: REQ-WPG-002, REQ-WPG-005, REQ-WPG-008.

### AC-WPG-003 — GEARS-first language in spec-assembly.md

**Severity**: Critical (P0 blocker)

**Given** the file `.claude/skills/moai/workflows/plan/spec-assembly.md` exists with GEARS-alignment edits applied
**When** a SPEC author reads the Phase 2 SPEC document creation section, frontmatter checklist, AC traceability summary, and quality gate criteria
**Then**:
- "GEARS structure" phrasing shall replace "EARS structure" wherever it appears in the body
- "GEARS ↔ AC coverage" shall replace "EARS ↔ AC coverage" in the quality gate section
- "GEARS-format requirements" shall replace "EARS-format requirements" wherever it appears
- The Phase 1B JWT example (line 530) shall reference GEARS notation as the requirement form

**Evidence command**:
```bash
grep -nE 'GEARS|EARS' .claude/skills/moai/workflows/plan/spec-assembly.md
```

**Pass criterion**: ≥ 6 GEARS occurrences (covering the 6 original EARS edit zones) AND ≥ 1 EARS-legacy-context occurrence.

**Verifies**: REQ-WPG-003, REQ-WPG-005, REQ-WPG-006, REQ-WPG-008.

### AC-WPG-004 — Cross-link to canonical GEARS SSOT

**Severity**: Major (P1)

**Given** any of the 3 in-scope content files (plan.md, clarity-interview.md, spec-assembly.md) has been GEARS-aligned
**When** a SPEC author follows the first "GEARS" reference in each file
**Then** at least 1 of the 3 files shall contain an explicit cross-link to `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format

**Evidence command**:
```bash
grep -rn 'moai-workflow-spec.*GEARS\|GEARS.*moai-workflow-spec' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/
```

**Pass criterion**: ≥ 1 match across the 4 in-scope files.

**Verifies**: REQ-WPG-004.

### AC-WPG-005 — EARS legacy reference retention (REQ-WPG-008)

**Severity**: Critical (P0 blocker)

**Given** the 4 in-scope local files have been GEARS-aligned
**When** a reader scans each file for EARS notation references
**Then** each of the 3 content files (plan.md, clarity-interview.md, spec-assembly.md) shall retain at least 1 explicit EARS reference (in legacy-footnote context — "EARS retained as legacy reference" or equivalent)

**Evidence command**:
```bash
grep -c 'EARS' .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/plan/clarity-interview.md .claude/skills/moai/workflows/plan/spec-assembly.md
```

**Pass criterion**: Each file count ≥ 1 (the 6-month backward-compat window mandates EARS retention until 2026-11-22).

**Verifies**: REQ-WPG-008.

### AC-WPG-006 — No IF/THEN deprecated modality introduced

**Severity**: Critical (P0 blocker)

**Given** the 4 in-scope local files + their 4 template mirror counterparts have been GEARS-aligned
**When** the sentinel grep runs against all 8 .md files
**Then** zero (0) occurrences of the deprecated `IF .* THEN` pattern shall be present in the modified files

**Evidence command**:
```bash
grep -rE 'IF .* THEN' \
  .claude/skills/moai/workflows/plan.md \
  .claude/skills/moai/workflows/plan/ \
  internal/template/templates/.claude/skills/moai/workflows/plan.md \
  internal/template/templates/.claude/skills/moai/workflows/plan/ \
  2>&1 | grep -v '^Binary' | wc -l
```

**Pass criterion**: Output = 0.

**Verifies**: REQ-WPG-011.

### AC-WPG-007 — Template mirror byte-for-byte parity

**Severity**: Critical (P0 blocker)

**Given** M1 (local edits) and M2 (template mirror sync) have both completed
**When** the orchestrator runs `diff -q` across all 4 file pairs and `diff -r` on the plan/ directory pair
**Then**:
- Each of the 4 file pairs shall be byte-for-byte identical (zero `diff -q` output)
- The plan/ directory diff shall report only the `.gitkeep` template-only file as a difference

**Evidence command**:
```bash
diff -q .claude/skills/moai/workflows/plan.md internal/template/templates/.claude/skills/moai/workflows/plan.md
diff -r .claude/skills/moai/workflows/plan/ internal/template/templates/.claude/skills/moai/workflows/plan/ | grep -v "^Only in .*: \.gitkeep$"
```

**Pass criterion**: Both commands produce empty output. The anchored pattern `^Only in .*: \.gitkeep$` matches only the `.gitkeep` template-only divergence line; any other unmatched output (e.g., genuine byte divergence in a tracked file) will pass through and fail the check.

**Verifies**: REQ-WPG-007, REQ-WPG-012.

### AC-WPG-008 — Status transition + frontmatter schema compliance

**Severity**: Critical (P0 blocker)

**Given** M4 (status transition) has completed
**When** the orchestrator inspects the spec.md frontmatter
**Then**:
- `status:` shall equal `in-progress` (transition from `draft` per Status Transition Ownership Matrix; manager-develop owns this transition)
- `updated:` shall equal today's ISO date (2026-05-25 or later)
- All 12 canonical frontmatter fields shall remain present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags)

**Evidence command**:
```bash
grep -E '^(status|updated|id|title|version|created|author|priority|phase|module|lifecycle|tags):' .moai/specs/SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001/spec.md | wc -l
grep -E '^status: in-progress$' .moai/specs/SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001/spec.md
```

**Pass criterion**: 12 canonical fields present AND `status: in-progress` present.

**Verifies**: REQ-WPG-007, REQ-WPG-013.

### AC-WPG-009 — Lint regression check (zero new LegacyEARSKeyword findings)

**Severity**: Major (P1)

**Given** M3 (sentinel verification + lint regression check) has completed
**When** `go run ./cmd/moai spec lint --json` runs against all SPECs in `.moai/specs/`
**Then** the count of `LegacyEARSKeyword` findings shall not exceed the pre-edit baseline (no regression; new SPECs in this edit cycle introduce 0 NEW LegacyEARSKeyword findings)

**Evidence command**:
```bash
go run ./cmd/moai spec lint --json 2>&1 | jq '[.findings[] | select(.rule == "LegacyEARSKeyword")] | length'
```

**Pass criterion**: Output count ≤ pre-edit baseline (capture baseline at M1 start; verify at M3 finish; delta ≤ 0).

**Verifies**: REQ-WPG-010, REQ-WPG-011, REQ-WPG-012.

### AC-WPG-010 — Pre-commit staging area scope discipline (L59)

**Severity**: Major (P1)

**Given** M4 (pre-commit staging assertion) is about to execute
**When** the orchestrator runs `git diff --cached --name-only | sort -u`
**Then** the staged paths shall be limited to:
- 4 local files (`.claude/skills/moai/workflows/plan*.md`, `.claude/skills/moai/workflows/plan/*.md`)
- 4 template mirror files (`internal/template/templates/.claude/skills/moai/workflows/plan*.md`, `internal/template/templates/.claude/skills/moai/workflows/plan/*.md`)
- spec.md (frontmatter status transition)
- progress.md (run-phase audit-ready signal)

Total expected: 10 paths (8 .md + spec.md + progress.md). If M1+M2+M3+M4 split across multiple commits, each commit asserts its own subset count.

**Evidence command**:
```bash
git diff --cached --name-only | sort -u | wc -l
git diff --cached --name-only | sort -u
```

**Pass criterion**: Staging set is a subset of the expected 10-path enumeration. Zero out-of-scope paths.

**Verifies**: L46 attribution discipline + L48 SSOT discipline + L59 pre-commit staging scope.

### AC-WPG-011 — spec-assembly.md cross-link to canonical frontmatter schema SSOT

**Severity**: Major (P1)

**Given** M1 has completed and `.claude/skills/moai/workflows/plan/spec-assembly.md` has been GEARS-aligned
**When** the orchestrator scans `spec-assembly.md` for a cross-link to the canonical frontmatter schema SSOT
**Then** the file shall contain at least 1 explicit reference to `spec-frontmatter-schema.md` in the Phase 2 SPEC document creation section, replacing or supplementing any inline restatement of the 12 canonical frontmatter fields

**Evidence command**:
```bash
grep -c 'spec-frontmatter-schema.md' .claude/skills/moai/workflows/plan/spec-assembly.md
```

**Pass criterion**: Output ≥ 1 (cross-link present, reduces drift risk per §22 SSOT discipline).

**Verifies**: REQ-WPG-009 (cross-link instead of inline 12-field restatement).

## §B — Traceability Matrix (REQ-WPG ↔ AC-WPG)

| REQ ID | AC IDs | Coverage |
|--------|--------|----------|
| REQ-WPG-001 (Ubiquitous: plan.md GEARS) | AC-WPG-001 | Full |
| REQ-WPG-002 (Ubiquitous: clarity-interview.md GEARS) | AC-WPG-002 | Full |
| REQ-WPG-003 (Ubiquitous: spec-assembly.md GEARS) | AC-WPG-003 | Full |
| REQ-WPG-004 (Ubiquitous: cross-link SSOT) | AC-WPG-001, AC-WPG-004 | Full |
| REQ-WPG-005 (Event-driven: manager-spec output) | AC-WPG-002, AC-WPG-003 | Full |
| REQ-WPG-006 (Event-driven: GEARS appears first) | AC-WPG-001, AC-WPG-003 | Full |
| REQ-WPG-007 (Event-driven: mirror parity sync) | AC-WPG-007, AC-WPG-008 | Full |
| REQ-WPG-008 (State-driven: EARS legacy retention) | AC-WPG-002, AC-WPG-003, AC-WPG-005 | Full |
| REQ-WPG-009 (State-driven: frontmatter cross-link) | AC-WPG-011 | Full (direct: grep verifies cross-link presence in spec-assembly.md) |
| REQ-WPG-010 (Capability-gate: no EARS warning) | AC-WPG-009 | Full |
| REQ-WPG-011 (Event-detected: no IF/THEN intro) | AC-WPG-006, AC-WPG-009 | Full |
| REQ-WPG-012 (Event-detected: mirror parity halt) | AC-WPG-007, AC-WPG-009 | Full |
| REQ-WPG-013 (Compound) | AC-WPG-001, AC-WPG-008 | Full |
| **Coverage check** | **13/13 REQs covered** | **100%** |

## §C — Edge Cases

### C.1 .gitkeep presence in template mirror

The `internal/template/templates/.claude/skills/moai/workflows/plan/.gitkeep` is a 0-byte placeholder file. It exists ONLY in the template mirror (not in the local skills tree). It has no notation content and shall be preserved untouched. `diff -r` will report `Only in internal/template/templates/.claude/skills/moai/workflows/plan: .gitkeep` — this is the expected and ONLY allowable divergence in the M2 mirror parity check.

### C.2 Parallel session race during M1-M5

If a parallel Sprint 10 cohort session (e.g., DOCS-SITE-FULL run-phase) commits scope-disjoint changes during this SPEC's M1-M5 phases, the orchestrator pre-spawn fetch (per `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check) detects `N 0` divergence and absorbs the parallel commits via fast-forward fetch. AC-WPG-010 path-specific staging assertion prevents cross-attribution leakage.

### C.3 Plan-auditor score below 0.85 PASS

Phase 0.5 plan-auditor independent verification may return < 0.85. If so, orchestrator returns to manager-spec with a focused-fix delegation (per FOUNDATION-CORE-GEARS-ALIGN-001 D1 discovery precedent) targeting the lowest-scoring MP dimension. Max iterations: 2 (per CONST-V3R5-026 + Tier M conventions). If iter-2 still < 0.85, escalate to user via AskUserQuestion blocker report from orchestrator.

### C.4 lint baseline drift

If `go run ./cmd/moai spec lint --json` baseline at M1 start differs from baseline previously observed (e.g., a parallel cohort closed an unrelated `LegacyEARSKeyword` finding between cohort #3 close and cohort #4 start), AC-WPG-009 uses the M1-start baseline (re-measured fresh) as the no-regression reference. This prevents stale-baseline false-positive AC failure.

## §D — Quality Gate Criteria + Definition of Done

### D.1 Plan-phase Definition of Done

- [x] spec.md created with 12 canonical frontmatter fields + 13 REQ-WPG entries (≥80% GEARS notation; this SPEC achieves 100%)
- [x] plan.md created with 6-milestone decomposition + lifecycle table + verification batch
- [x] acceptance.md created with 11 mandatory ACs + traceability matrix + edge cases
- [x] progress.md created with plan_commit_sha placeholder
- [ ] Phase 0.5 plan-auditor independent verification ≥ 0.85 PASS (target 0.90 skip-eligible)

### D.2 Run-phase Definition of Done

- [ ] M1: 4 local files edited per §D Milestone Decomposition (including spec-assembly.md cross-link to spec-frontmatter-schema.md SSOT per REQ-WPG-009 / AC-WPG-011)
- [ ] M2: 4 template mirror files byte-for-byte parity (AC-WPG-007 PASS)
- [ ] M3: Sentinel verification batch + lint regression check (AC-WPG-006, AC-WPG-009 PASS) + spec-assembly.md cross-link grep ≥ 1 (AC-WPG-011 PASS)
- [ ] M4: spec.md frontmatter `status: draft → in-progress` (AC-WPG-008 PASS) + pre-commit staging scope exactly 10 paths per single commit (AC-WPG-010 PASS)
- [ ] 11/11 ACs PASS verified inline in progress.md §E.2

### D.3 Sync-phase Definition of Done

- [ ] M5: 4-artifact `sync_commit_sha` atomic backfill (L60 discipline)
- [ ] CHANGELOG entry under appropriate version section
- [ ] spec.md frontmatter `status: in-progress → implemented` (manager-docs owns)
- [ ] progress.md §E.4 Sync-phase Audit-Ready Signal

### D.4 Mx-phase Definition of Done

- [ ] M6: Orchestrator Mx Step C judge → SKIP-eligible verdict expected (markdown-only, 0 .go files, 0 @MX delta)
- [ ] progress.md §E.5 Mx-phase Audit-Ready Signal
- [ ] 4-phase close marker (plan + run + sync + mx) in progress.md

### D.5 Predecessor pattern fidelity check

This SPEC follows the cohort-3 FOUNDATION-CORE-GEARS-ALIGN-001 precedent:
- Tier M, 1-pass run-phase target (6 commits acceptable: M1+M2+M3+M4 split, M5 sync, M6 Mx)
- Plan-auditor target 0.87+ (PASS), 0.90+ (skip-eligible)
- 13 REQs + 11 ACs (cohort-3 had 12 REQs + 9 ACs; this is appropriately scaled to 4-file scope; +1 AC vs iter-1 closes D2 REQ-WPG-009 trace orphan)
- 100% GEARS self-dogfood (cohort-3 achieved similar)
- L46 + L48 + L59 attribution + SSOT + staging discipline maintained
- Template mirror parity enforced via `diff -q` in M2 + AC-WPG-007
