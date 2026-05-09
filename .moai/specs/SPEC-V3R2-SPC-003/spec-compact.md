---
id: SPEC-V3R2-SPC-003
document: spec-compact
version: "0.1.0"
status: backfilled
created: 2026-05-10
updated: 2026-05-10
author: manager-spec (Batch 3 backfill)
related_spec: SPEC-V3R2-SPC-003
phase: plan
language: ko
---

# SPEC-V3R2-SPC-003 Compact Reference

> Auto-extract from `spec.md` v0.1.1 — REQs + ACs + Files + Exclusions only.
> Use this file for fast plan-auditor scans and cross-SPEC referencing.
> **Status: implemented** (PR #745 merged Wave 5, main commit `03146d1ae`).

---

## Goal (one sentence)

Provide a deterministic, CI-integrable `moai spec lint` subcommand that validates every SPEC document under `.moai/specs/SPEC-*/spec.md` against the EARS format, hierarchical acceptance schema (SPC-001), frontmatter schema (CON-001), and inter-SPEC dependency DAG (no cycles).

---

## Scope

### In Scope (10 items)

1. `moai spec lint` CLI subcommand (single file or full tree)
2. EARS compliance: 5 modality forms + REQ ID regex `^REQ-[A-Z]{2,5}-\d{3}-\d{3}$`
3. REQ ID uniqueness within SPEC
4. AC→REQ coverage ≥ 100%
5. Frontmatter schema validation (14 required fields)
6. Dependency DAG cycle detection (Tarjan SCC)
7. Hierarchical AC structural validation (delegated to SPC-001 parser)
8. Zone registry cross-reference (CON-001)
9. Out-of-Scope subsection presence
10. Output modes: table (default) / JSON / SARIF 2.1.0

### Out of Scope

- @MX TAG validation (→ SPEC-V3R2-SPC-002)
- Subjective content quality (grammar, ambiguity)
- Test execution against implementation (evaluator-active scope)
- `--fix` auto-correction (intentionally deferred)
- Non-SPEC artifact validation (reports, decisions)

---

## Requirements (EARS, 18 total)

### Ubiquitous (10)

- **REQ-SPC-003-001**: `moai spec lint` SHALL accept zero-or-more SPEC paths; no args → lint all `.moai/specs/SPEC-*/spec.md`.
- **REQ-SPC-003-002**: SHALL exit 0 iff no errors; warnings/info SHALL NOT affect exit by default.
- **REQ-SPC-003-003**: Every requirement SHALL use exactly one EARS modality form (Ubiquitous / Event / State / Optional / Unwanted / Complex).
- **REQ-SPC-003-004**: Every REQ ID SHALL match `^REQ-[A-Z]{2,5}-\d{3}-\d{3}$` and be unique within SPEC.
- **REQ-SPC-003-005**: Every REQ SHALL appear in ≥1 leaf AC's `(maps REQ-...)` tail; uncovered REQs → `CoverageIncomplete` error.
- **REQ-SPC-003-006**: 14 frontmatter fields SHALL be present with correct types.
- **REQ-SPC-003-007**: Every `dependencies:` entry SHALL be a valid SPEC ID and exist on disk.
- **REQ-SPC-003-008**: SHALL detect cycles in dependency DAG; cycles → `DependencyCycle` error.
- **REQ-SPC-003-009**: Every spec.md SHALL contain `Out of Scope` subsection with ≥1 entry; violations → `MissingExclusions`.
- **REQ-SPC-003-010**: Every `related_rule: [CONST-V3R2-NNN]` SHALL exist in CON-001 zone registry; dangling → `DanglingRuleReference` warning.

### Event-Driven (3)

- **REQ-SPC-003-020**: WHEN `--json` invoked, SHALL emit JSON array of findings to stdout.
- **REQ-SPC-003-021**: WHEN `--sarif` invoked, SHALL emit SARIF 2.1.0-conformant JSON.
- **REQ-SPC-003-022**: WHEN unparseable file encountered, SHALL emit `ParseFailure` error and continue with remaining files.

### State-Driven (2)

- **REQ-SPC-003-030**: WHILE `--strict` set, WARNINGS SHALL be promoted to ERRORS and cause non-zero exit.
- **REQ-SPC-003-031**: WHILE two SPECs declare same `id:`, SHALL report `DuplicateSPECID` error citing both paths.

### Optional (2)

- **REQ-SPC-003-040**: WHERE frontmatter `lint.skip: [CODE]` present, SHALL suppress matching findings for that SPEC only.
- **REQ-SPC-003-041**: WHERE `--format table` specified, default human-readable output is explicitly selected.

### Complex (3)

- **REQ-SPC-003-050**: WHILE requirement starts with "WHEN" but lacks "SHALL", SHALL report `ModalityMalformed`.
- **REQ-SPC-003-051**: WHEN AC tail references nonexistent REQ → `UnmappedRequirement`; ELSE if REQ exists but no AC → `CoverageIncomplete`.
- **REQ-SPC-003-052**: IF `breaking: true` AND `bc_id: []` → `BreakingChangeMissingID` error; IF `breaking: false` AND `bc_id` non-empty → `OrphanBCID` warning.

---

## Acceptance Criteria Index (16 total)

| AC ID | Subject | Test Function | REQ Map |
|-------|---------|---------------|---------|
| AC-V3R2-SPC-003-01 | Happy path: 0 findings, exit 0 | `TestLinter_AC01_HappyPath` | 001, 002, 005 |
| AC-V3R2-SPC-003-02 | CoverageIncomplete: REQ unmapped | `TestLinter_AC02_CoverageIncomplete` | 005 |
| AC-V3R2-SPC-003-03 | ModalityMalformed: WHEN missing SHALL | `TestLinter_AC03_ModalityMalformed` | 003, 050 |
| AC-V3R2-SPC-003-04 | DependencyCycle: A↔B | `TestLinter_AC04_DependencyCycle` | 008 |
| AC-V3R2-SPC-003-05 | DuplicateREQID within SPEC | `TestLinter_AC05_DuplicateREQID` | 004 |
| AC-V3R2-SPC-003-06 | MissingExclusions: no Out-of-Scope | `TestLinter_AC06_MissingExclusions` | 009 |
| AC-V3R2-SPC-003-07 | MissingDependency: SPEC-NONEXISTENT | `TestLinter_AC07_MissingDependency` | 007 |
| AC-V3R2-SPC-003-08 | DanglingRuleReference: CONST-V3R2-999 | `TestLinter_AC08_DanglingRuleReference` | 010 |
| AC-V3R2-SPC-003-09 | JSON output is valid array | `TestLinter_AC09_JSONOutput` | 020 |
| AC-V3R2-SPC-003-10 | SARIF 2.1.0 schema conformance | `TestLinter_AC10_SARIFOutput` | 021 |
| AC-V3R2-SPC-003-11 | --strict promotes warning → error | `TestLinter_AC11_StrictMode` | 030 |
| AC-V3R2-SPC-003-12 | DuplicateSPECID across two SPECs | `TestLinter_AC12_DuplicateSPECID` | 031 |
| AC-V3R2-SPC-003-13 | lint.skip suppresses matched code | `TestLinter_AC13_LintSkip` | 040 |
| AC-V3R2-SPC-003-14 | BreakingChangeMissingID: breaking + empty bc_id | `TestLinter_AC14_BreakingChangeMissingID` | 052 |
| AC-V3R2-SPC-003-15 | ParseFailure: malformed YAML, continue | `TestLinter_AC15_ParseFailure` | 022, 006 |
| AC-V3R2-SPC-003-16 | Hierarchical AC coverage: leaf maps parent | `TestLinter_AC16_HierarchicalACCoverage` | 005 |

REQ coverage: 18 REQs → 16 ACs (some REQs share ACs). 100% REQ→AC mapping.

---

## Files to Modify (as-merged)

### Created (5 + testdata)

- `internal/spec/lint.go` (816 LOC)
- `internal/spec/lint_test.go` (572 LOC)
- `internal/spec/dag.go` (100 LOC)
- `internal/spec/sarif.go` (165 LOC)
- `internal/cli/spec_lint.go` (164 LOC)
- `internal/spec/testdata/` (11 fixture SPEC directories)

### Modified

- `internal/cli/spec.go` (`newSpecLintCmd()` registration, ~1 LOC)

Total: 1245 source LOC + 572 test LOC + 11 fixture dirs.

---

## Exclusions (Out of Scope — What NOT to Build)

Per `spec.md` §2.2:

1. @MX TAG validation (delegated to SPEC-V3R2-SPC-002)
2. SPEC content quality checks (grammar, clarity, ambiguity) — too subjective
3. Test execution against implementation (evaluator-active's job)
4. `moai spec lint --fix` auto-correction (deferred; linter is report-only)
5. Non-SPEC artifact validation (reports under `.moai/reports/`, decisions under `.moai/decisions/`)

---

## Implementation Status

- [x] All 16 ACs PASS (`go test ./internal/spec/...`)
- [x] internal/spec coverage 86.6% (target ≥85%)
- [x] `go vet ./...` 0 issues
- [x] `golangci-lint run ./...` 0 issues
- [x] `go test -race ./...` no data races
- [x] PR #745 merged to main as commit `03146d1ae` (2026-04-30, Wave 5)

---

End of spec-compact.md (backfill v0.1.0).
