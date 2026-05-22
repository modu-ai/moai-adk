---
spec_id: SPEC-V3R6-GEARS-MIGRATION-001
created: 2026-05-22
updated: 2026-05-22
---

# Implementation Plan — SPEC-V3R6-GEARS-MIGRATION-001

## 0. Executive Summary

Tier M, 4-milestone plan. Total estimated change: ~300 LOC source + ~600 LOC fixtures + 4 × ~100 lines locale docs = ~1300 lines net. Touches `internal/spec/lint.go` (1 file edit + 1 new finding code), `internal/spec/lint_test.go` (test additions), `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-plan.md` (4 locale additions; final path resolved at run-phase entry per Q4), and creates `test-fixtures/legacy-ears-sample.md` + `test-fixtures/gears-sample.md` under this SPEC's directory.

**Delegation pattern**: full `manager-develop` Tier M template (Sections A-E mandatory per `manager-develop-prompt-template.md`). Tier M SPECs require the 5-section delegation prompt; orchestrator-direct execution is anti-pattern for code-touching SPECs of this surface area.

**Run-phase entry hard gate**: M1 (web-research validation) MUST land in main before M2 (lint.go edit) is attempted. If M1 verdict is MISMATCH, this SPEC re-enters plan-phase amendment before M2 proceeds.

## 1. Approach + Strategy

### 1.1 Single-axis sequencing (M1 → M2 → M3 → M4)

The 4 milestones are strictly sequential, NOT parallel:

- **M1 must precede M2**: The lint.go edit (M2) encodes the canonical GEARS keyword set from the paper (M1 output). Skipping M1 risks shipping a keyword definition the source paper does not endorse.
- **M2 must precede M3**: The 4-locale docs guide (M3) must cite the actual `LegacyEARSKeyword` finding message text + the lint code identifier. Drafting M3 against a phantom message produces docs drift on day-one.
- **M3 must precede M4**: The CI regression guards (M4) must verify the 4-locale docs URL appears in the finding message; without M3 landed, M4's grep test fails.

No parallel acceleration via Agent Teams is appropriate here — the surface is small and the sequential dependency is real.

### 1.2 Late-Branch + Hybrid Trunk fit

- **Tier M**, ~1300 lines total → exceeds Tier S threshold; **PR + branch required** (NOT main direct push per CLAUDE.local.md §23 Hybrid Trunk).
- Branch: `feat/SPEC-V3R6-GEARS-MIGRATION-001`.
- 4 milestone commits → 1 squash merge at PR completion.
- Cross-Wave caveat: this SPEC plans before Wave 1-5 complete (user opt-in 2026-05-22). Plan PR may land independently; run PR blocks on cross-Wave compatibility (AC-GM-007 second clause).

### 1.3 manager-develop delegation prompt requirements

Per `manager-develop-prompt-template.md` (Tier M REQUIRED):

- Section A (Context): path + branch + plan artifact line counts + plan-auditor verdict + PRESERVE list
- Section B (Known Issues): inject B2 (cross-SPEC scan against in-flight Wave 2-5), B3 (subagent boundary in `internal/spec/`), B4 (frontmatter canonical), B5 (CI 3-tier), B6 (Out of Scope h3), B8 (working tree hygiene). B1 (cross-platform) optional — pure Go file edits.
- Section C (Pre-flight): cross-platform build, lint baseline, retired/superseded scan in `internal/spec/`.
- Section D (Constraints): PRESERVE = 88 existing SPECs + all untracked files. EXTEND = lint.go (additive only, ModalityMalformed code unchanged) + lint_test.go + new fixtures dir + 4 locale docs.
- Section E (Self-Verification): all 8 ACs + cross-platform + coverage + C-HRA-008 grep + lint baseline before/after + new test PASS.

## 2. Milestones

| ID | Title | Files touched | Estimated LOC | Blocker on prior |
|----|-------|---------------|--------------:|------------------|
| M1 | GEARS paper validation (run-phase first step) | `.moai/research/gears-paper-validation.md` (NEW, ~80 lines) | 80 | — |
| M2 | Lint.go GEARS support + new finding code | `internal/spec/lint.go` (edit ~30 LOC), `internal/spec/lint_test.go` (add ~120 LOC), `test-fixtures/legacy-ears-sample.md` + `test-fixtures/gears-sample.md` (NEW, ~150 lines each = 300 total) | ~450 | M1 |
| M3 | 4-locale docs-site GEARS migration guide | `docs-site/content/{ko,en,ja,zh}/<resolved-path>` (4 files, ~100 lines each = 400 lines) | 400 | M2 |
| M4 | CI regression guards + chore commit | New test cases in `lint_test.go` for AC-GM-002/003/008 + chore PR description + status frontmatter v0.1.0 → v0.2.0 | ~80 | M3 |

Total estimate: 80 + 450 + 400 + 80 = **~1010 lines net** (within Tier M range 300-1000 LOC source + reasonable fixture overhead).

### M1 — GEARS Paper Validation (orchestrator-direct, NOT manager-develop)

**Owner**: orchestrator (web-research is orchestration responsibility, not implementation).

**Deliverables**:

1. WebSearch query: `"GEARS notation" requirements syntax "Easy Approach" 2026` and variants.
2. WebFetch top candidates (target: official EARS guide successor or DEV 2026 paper PDF/HTML).
3. Write `.moai/research/gears-paper-validation.md` with:
   - Verified URL(s)
   - Verbatim GEARS keyword definitions
   - Side-by-side diff against this SPEC's spec.md §1 GEARS table
   - **Verdict line**: `Verdict: MATCH` OR `Verdict: MISMATCH — re-open SPEC for amendment, see appendix`

**Pass criteria**:

- AC-GM-005 PASS: file exists, verdict line present.
- If `Verdict: MATCH`: proceed to M2.
- If `Verdict: MISMATCH`: HALT M2. Plan-phase amendment commits required first.

**Rationale for orchestrator-direct**: Pure web-research + markdown write. Delegating to manager-develop adds overhead with no quality benefit. Matches V3R6-CATALOG-SSOT-001 + V3R6-CONFIG-DEAD-CLEANUP-001 precedent.

### M2 — Lint.go GEARS Support + Finding Code

**Owner**: manager-develop (Tier M REQUIRED template).

**Deliverables**:

1. Edit `internal/spec/lint.go` `EARSModalityRule.Check()`:
   - Add new method `isLegacyEARSPattern(text string) bool` that returns true ONLY for `IF ... THEN` REQs.
   - When `isLegacyEARSPattern(req.Text) == true`: emit additional `Finding{Code: "LegacyEARSKeyword", Severity: SeverityWarning, Message: "REQ %s: GEARS migration: replace IF/THEN with WHEN/event normalization; see https://adk.mo.ai.kr/<locale>/workflow-commands/moai-plan/#gears-notation"}`.
   - Preserve existing `ModalityMalformed` check unchanged (additive, no breaking change).
   - Add a 12-line source comment block explaining the 6-month window (AC-GM-009).
2. Add test cases in `internal/spec/lint_test.go`:
   - `TestEARSModalityRule_LegacyEARSKeyword_IFThen`: IF/THEN REQ → exactly 1 `LegacyEARSKeyword` finding, severity warning.
   - `TestEARSModalityRule_GEARSWellFormed`: WHEN/WHILE/WHERE/Ubiquitous REQs → zero findings.
   - `TestEARSModalityRule_LegacyEARSKeyword_StrictExitCode`: --strict flag escalates warning to exit-1.
   - `TestEARSModalityRule_MessageContainsDocsURL`: finding message contains `adk.mo.ai.kr` substring.
3. Create fixtures (acceptance.md AC-GM-001/003 requirement):
   - `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md`
   - `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/gears-sample.md`
   - Both with full 12-field frontmatter so they parse cleanly through `parseSPECDoc`.

**Pre-flight verification (manager-develop runs first)**:

```bash
# Baseline: zero LegacyEARSKeyword findings exist anywhere
grep -rn 'LegacyEARSKeyword' internal/spec/  # empty
golangci-lint run --timeout=2m internal/spec/...  # baseline count
go test ./internal/spec/...  # baseline PASS count
```

**Pass criteria**:

- All 4 new tests PASS.
- `go test ./internal/spec/...` exit 0.
- Coverage on lint.go remains >= 85%.
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

### M3 — 4-Locale docs-site GEARS Migration Guide

**Owner**: manager-develop (continues from M2 commit).

**Deliverables**:

1. Resolve Open Question Q4 at orchestrator pre-flight (default: `workflow-commands/moai-plan.md`).
2. Edit 4 locale files (ko/en/ja/zh) with **identical structure** + locale-appropriate translations:
   - New H2 section: `## GEARS 표기법 (v3.0.0+)` (ko) / `## GEARS notation (v3.0.0+)` (en) / `## GEARS 表記法 (v3.0.0+)` (ja) / `## GEARS 表示法 (v3.0.0+)` (zh).
   - 5-pattern table identical to spec.md §1.
   - 6-month window explanation (AC-GM-009).
   - "For tool authors" subsection (downstream impact).
   - Cross-link to lint rule code identifier (`LegacyEARSKeyword`).
3. Verify Hugo build PASS from `docs-site/` with `hugo --gc --minify`.

**i18n discipline** (per `.moai/docs/docs-site-i18n-rules.md` §17.3):

- ko as the original, en/ja/zh as translations.
- Same H2 anchor across 4 locales (Hugo anchor generation deterministic).
- Translation parity ratio (max/min line count) <= 1.20 per AC-GM-004.

**Pass criteria**:

- AC-GM-004 verification commands all return expected outputs.
- Hugo build exit 0.

### M4 — CI Regression Guards + Chore Commit

**Owner**: manager-develop (chore-only commit, M2/M3 changes already landed).

**Deliverables**:

1. Verify all 8 ACs (AC-GM-001..009 minus vacant AC-GM-006) via the commands in acceptance.md.
2. Update SPEC frontmatter: `status: draft → implemented`, `version: 0.1.0 → 0.2.0`, `updated: <new date>`.
3. Create `progress.md` with run-phase evidence:
   - All AC verification command outputs.
   - Cross-platform build evidence.
   - Coverage delta evidence.
   - Baseline lint comparison (before/after pre-existing failure delta).
4. Hybrid Trunk Tier M: `feat/SPEC-V3R6-GEARS-MIGRATION-001` branch → squash PR → admin merge.

**Pass criteria**:

- progress.md complete with 8 AC outputs.
- PR description includes provenance (this SPEC ID + cross-Wave compatibility check from AC-GM-007 second clause).

## 3. Detailed Technical Approach (M2 Focus)

### 3.1 `internal/spec/lint.go` Diff (planned, not authoritative)

```go
// EARSModalityRule checks REQ text for EARS modality compliance
// AND emits GEARS migration warnings for legacy IF/THEN patterns.
//
// GEARS migration policy (SPEC-V3R6-GEARS-MIGRATION-001):
//   - WHEN/WHILE/WHERE/Ubiquitous "The system shall" remain canonical.
//   - IF ... THEN is deprecated; flagged with LegacyEARSKeyword (warning).
//   - Backward-compat window: 6 months from v3.0.0 OR until
//     SPEC-V3R6-GEARS-SWEEP-001 bulk-rewrites all 88 existing SPECs,
//     whichever comes first. After window: SPEC-V3R6-V3-CUTOVER-001
//     promotes the warning to error.
//   - --strict mode escalates warnings to errors immediately (existing
//     Report.HasErrors() behavior; opt-in for CI authors).
//
// Implements REQ-SPC-003-003, REQ-SPC-003-050, REQ-GM-002, REQ-GM-006, REQ-GM-009
type EARSModalityRule struct{}

func (r *EARSModalityRule) Code() string { return "ModalityMalformed" }

func (r *EARSModalityRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
    var findings []Finding
    for _, req := range doc.REQs {
        // Existing legacy check (unchanged)
        if isModalityMalformed(req.Text) {
            findings = append(findings, Finding{
                File: doc.Path, Line: req.Line,
                Severity: SeverityError, Code: "ModalityMalformed",
                Message: fmt.Sprintf("REQ %s: EARS modality violation — SHALL missing or format mismatch: %q", req.ID, req.Text),
            })
        }
        // NEW: GEARS migration warning for legacy IF/THEN patterns
        if isLegacyEARSPattern(req.Text) {
            findings = append(findings, Finding{
                File: doc.Path, Line: req.Line,
                Severity: SeverityWarning, Code: "LegacyEARSKeyword",
                Message: fmt.Sprintf("REQ %s: GEARS migration: replace IF/THEN with WHEN/event normalization; see https://adk.mo.ai.kr/en/workflow-commands/moai-plan/#gears-notation", req.ID),
            })
        }
    }
    return findings
}

// isLegacyEARSPattern returns true ONLY for IF ... THEN REQs.
// Other EARS keywords (WHEN/WHILE/WHERE/Ubiquitous) are GEARS-compatible.
// SPEC-V3R6-GEARS-MIGRATION-001 REQ-GM-002.
func isLegacyEARSPattern(text string) bool {
    upper := strings.ToUpper(text)
    return strings.HasPrefix(upper, "IF ") && strings.Contains(upper, " THEN ")
}
```

**Size estimate**: +30 LOC source (existing `isModalityMalformed` unchanged, new `isLegacyEARSPattern` 5 LOC, new finding-emit block in Check() 10 LOC, comment block 15 LOC).

### 3.2 Finding code addition (additive, non-breaking)

| Code | Severity | Introduced by | Behavior change |
|------|----------|---------------|-----------------|
| `ModalityMalformed` | Error | Pre-existing | UNCHANGED |
| `LegacyEARSKeyword` | Warning | This SPEC | NEW — emitted for IF/THEN REQs |

`Report.HasErrors()` logic is unchanged; warnings count toward errors only when `Strict == true` (existing `--strict` behavior, lines 49-60 of lint.go).

### 3.3 Fixture file design

**`legacy-ears-sample.md`** (AC-GM-001 + AC-GM-002 + AC-GM-008 fixture):

```yaml
---
id: SPEC-FIXTURE-LEGACY-001
title: "Legacy EARS sample"
version: "0.0.1"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: GEARS-MIGRATION-001-fixture
priority: P3
phase: "v3.0.0"
module: "test-fixtures"
lifecycle: spec-anchored
tags: "fixture, ears, legacy"
---

## Requirements

### REQ-LEG-001
The system shall log every request.

### REQ-LEG-002
WHEN a user logs in, the system shall record the timestamp.

### REQ-LEG-003
WHILE the cache is warming, the system shall return stale data.

### REQ-LEG-004
WHERE multi-tenancy is enabled, the system shall partition data per tenant.

### REQ-LEG-005
IF input is null THEN the system shall return an error.

### 1.1 Out of Scope

- Database migration handling.
```

Expected lint output:

- 0 `ModalityMalformed` findings (all REQs well-formed per legacy rule).
- 1 `LegacyEARSKeyword` finding (REQ-LEG-005 IF/THEN).
- Exit 0 non-strict; exit 1 with `--strict`.

**`gears-sample.md`** (AC-GM-003 fixture):

```yaml
---
id: SPEC-FIXTURE-GEARS-001
title: "GEARS canonical sample"
version: "0.0.1"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: GEARS-MIGRATION-001-fixture
priority: P3
phase: "v3.0.0"
module: "test-fixtures"
lifecycle: spec-anchored
tags: "fixture, gears, canonical"
---

## Requirements

### REQ-GRS-001
The system shall log every request.

### REQ-GRS-002
WHEN a user logs in, the system shall record the timestamp.

### REQ-GRS-003
WHILE the cache is warming, the system shall return stale data.

### REQ-GRS-004
WHERE multi-tenancy is enabled, the system shall partition data per tenant.

### 1.1 Out of Scope

- Database migration handling.
```

Expected lint output: 0 findings total. Exit 0 in both non-strict and `--strict`.

### 3.4 `discoverSPECs` interaction with fixtures

The `internal/spec/lint.go` `discoverSPECs(baseDir)` function (line 224) walks `.moai/specs/` recursively. Fixtures under `test-fixtures/` subdirectory WILL be picked up by `go run ./cmd/moai spec lint .moai/specs/`.

**Mitigation strategy** (decision at M2):

- Option A (preferred, low-risk): Fixtures use frontmatter with `tags: "fixture, ..."` and include proper `### N.Y Out of Scope` so they lint clean. Picked up but produce zero findings (good — proves the lint engine is consistent on canonical fixtures).
- Option B (riskier): Modify `discoverSPECs` to skip `test-fixtures/` directory. Adds 5 LOC + risk of regressing existing SPEC discovery.

Plan defaults to **Option A**.

## 4. Files Touched Summary

| Path | Operation | LOC | Owner |
|------|-----------|----:|-------|
| `.moai/research/gears-paper-validation.md` | NEW | ~80 | orchestrator (M1) |
| `internal/spec/lint.go` | EDIT (additive) | ~30 | manager-develop (M2) |
| `internal/spec/lint_test.go` | EDIT (4 new tests) | ~120 | manager-develop (M2) |
| `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md` | NEW | ~150 | manager-develop (M2) |
| `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/gears-sample.md` | NEW | ~150 | manager-develop (M2) |
| `docs-site/content/ko/workflow-commands/moai-plan.md` | EDIT (append) | ~100 | manager-develop (M3) |
| `docs-site/content/en/workflow-commands/moai-plan.md` | EDIT (append) | ~100 | manager-develop (M3) |
| `docs-site/content/ja/workflow-commands/moai-plan.md` | EDIT (append) | ~100 | manager-develop (M3) |
| `docs-site/content/zh/workflow-commands/moai-plan.md` | EDIT (append) | ~100 | manager-develop (M3) |
| `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/spec.md` | EDIT (status, version) | ~5 | orchestrator (M4 chore) |
| `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/progress.md` | NEW | ~200 | orchestrator (M4 chore) |

**Total**: 11 file operations across 4 milestones.

## 5. EARS → GEARS Keyword Mapping Table (Reference)

For ready use during M3 docs authoring. Source: this SPEC's spec.md §1 + verified via M1 web-research.

| Notation | EARS (legacy) | GEARS (canonical) | Lint behavior on EARS legacy |
|----------|---------------|-------------------|------------------------------|
| Ubiquitous | `The system shall <action>` | Same | Unchanged (`ModalityMalformed` only on SHALL missing) |
| Event-driven | `WHEN <event>, the system shall <action>` | Same | Unchanged |
| State-driven | `WHILE <state>, the system shall <action>` | Same | Unchanged |
| Precondition / capability | `WHERE <feature-exists>, the system shall <action>` | `WHERE <precondition>, the system shall <action>` (reframed) | Unchanged at lint layer; reframing is documentation-only |
| Negative trigger | `IF <undesired-event>, THEN the system shall <action>` | **DEPRECATED** — use `WHEN <event-detected>, the system shall <action>` | NEW: `LegacyEARSKeyword` warning emitted |

## 6. Risks + Mitigations

| ID | Risk | Probability | Impact | Mitigation |
|----|------|:-----------:|:------:|------------|
| R1 | GEARS paper notation mismatches §1 Table | M | H | M1 (AC-GM-005) is run-phase first step + HALT gate. Plan-phase amendment if mismatch. |
| R2 | 88-SPEC backward-compat regression — accidental lint behavior change breaks `moai spec lint .moai/specs/` baseline | L | H | AC-GM-001 + AC-GM-007 explicit diff verification. Baseline captured pre-edit, compared post-edit. |
| R3 | 4-locale translation drift — Hugo build PASS but content semantically diverged across locales | M | M | AC-GM-004 line-count parity (<=1.20) + table-row grep count check. `.moai/docs/docs-site-i18n-rules.md` §17.3 HARD obligation. |
| R4 | Downstream tool breakage from new `LegacyEARSKeyword` finding code — third-party tools assuming only `ModalityMalformed` exists | L | L | `LegacyEARSKeyword` is additive (not replacement). Severity warning (non-blocking by default). Migration guide §"For tool authors" provides upgrade path. |
| R5 | Cross-Wave ordering — this SPEC plans before Wave 2-5 complete; in-flight Wave 2-5 SPECs may reach implementation status with REQ patterns this SPEC then flags | M | M | AC-GM-007 second clause: run-phase orchestrator checks all Wave 2-5 SPEC statuses. If any reached `in-progress` between plan-merge and run-entry, plan-auditor cross-Wave compatibility check runs first. |
| R6 | Fixture files picked up by recursive `moai spec lint .moai/specs/` scan and break the 88-SPEC baseline | L | M | Option A in §3.4: fixtures are valid SPECs with `### Out of Scope`. Verified by AC-GM-001 (baseline lint exit code unchanged after fixtures added). |

## 7. plan-auditor Pre-Submission Checklist

Per `.claude/agents/meta/plan-auditor.md` § Tier M threshold (0.80):

| Dimension | Score target | Self-assessment evidence |
|-----------|-------------:|--------------------------|
| D1 (REQ↔AC traceability + EARS modality) | 0.85+ | acceptance.md Traceability Matrix shows 9/9 REQs → 9 ACs (100% coverage). REQs §3 dogfood GEARS notation as ubiquitous demonstration. |
| D2 (technical feasibility + risk surface) | 0.85+ | M1 web-research first-step blocker mitigates highest-impact risk (R1). M2 source-edit is additive 30-LOC; M3 i18n surface is contained. R1-R6 all have explicit mitigations. |
| D3 (artifact structure + spec-lint compliance) | 0.85+ | 3 artifacts (spec.md / acceptance.md / plan.md) per Tier M default. spec.md has 12-field frontmatter (verified against SSOT) + 3× `### N.Y Out of Scope` h3 subsections. acceptance.md has 8 active binary ACs. |
| D4 (cross-SPEC alignment + scope hygiene) | 0.80+ | Cross-Wave caveat explicitly documented in §1.2 + §4 + R5 + AC-GM-007. SPEC-V3R6-V3-CUTOVER-001 dependency named (post-window cutover). SPEC-V3R6-GEARS-SWEEP-001 forward-named (out of this SPEC's scope). |

Expected aggregate: ~0.84 (above Tier M 0.80 threshold).

## 8. Open Questions (Plan-Phase Discoverable)

These were raised in spec.md §5 and recorded here as planning-time open questions. Defaults are applied unless user overrides at run-phase entry:

| ID | Question | Default | Run-phase action if default |
|----|----------|---------|----------------------------|
| Q1 | Bulk-rewrite scope (this SPEC vs SWEEP-001) | Defer to SWEEP-001 | Proceed M1-M4 with 88 SPECs untouched. AC-GM-007 verifies. |
| Q2 | `--strict` enforcement timing | Immediate (v3.0.0 release) | M2 source comment documents. No timed-rollout code. |
| Q3 | GEARS paper URL | Mandatory M1 web-research | M1 first-step. HALT if MISMATCH. |
| Q4 | docs-site target file | `workflow-commands/moai-plan.md` × 4 | M3 amends existing file (4-locale parity preserved). |

Any user override at run-phase requires plan-phase amendment commit BEFORE M1 proceeds.

## 9. Cross-References

- spec.md (this SPEC)
- acceptance.md (this SPEC)
- `.moai/research/v3-redesign-blueprint-2026-05-22.md` line 16, 24, 174, 200-206
- `internal/spec/lint.go` lines 404-447
- `internal/spec/lint_test.go` (existing test patterns for `EARSModalityRule`)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12-field SSOT)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (Section A-E template, Tier M REQUIRED)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier M = 3 artifacts, 0.80 PASS threshold)
- `.moai/docs/docs-site-i18n-rules.md` §17.3 (4-locale HARD obligation)
- Reference SPEC structures: `.moai/specs/SPEC-V3R6-HOOK-CONTRACT-FIX-001/`, `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/`
