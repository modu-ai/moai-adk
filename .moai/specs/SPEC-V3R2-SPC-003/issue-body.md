## SPEC-V3R2-SPC-003 — `moai spec lint` (RETROACTIVE plan-phase BACKFILL)

> **Phase**: v3.0.0 — Phase 7 — Extension
> **Module**: `internal/spec/`, `internal/cli/spec_lint.go`
> **Priority**: P1 High
> **Breaking**: false
> **Lifecycle**: spec-anchored
> **Status**: implemented (PR #745 merged 2026-04-30, Wave 5)

## Summary

This PR back-fills the **plan-phase artifacts** (`research.md`, `plan.md`, `acceptance.md`, `tasks.md`, `spec-compact.md`, `issue-body.md`) for SPEC-V3R2-SPC-003 (`moai spec lint`). The implementation has already been merged to `main` via **PR #745** as part of the Wave 5 batch on 2026-04-30. Plan-phase normalization was missed at the time. This is **SPEC-First retroactive normalization** — no code changes are made.

## Why backfill?

MoAI-ADK enforces a SPEC-First discipline where every implementation MUST be backed by a complete plan-phase artifact set under `.moai/specs/SPEC-XXX/`. SPC-003 was authored only with `spec.md` + `progress.md` and was merged in Wave 5 without the standard 6-file companion set. Batch 3 of the SPEC-First normalization sweep (2026-05-10) brings SPC-003 into compliance:

- Establishes traceability between commit history (`44019ec4d` RED → `4d5699f27` GREEN → `1587e6e3a` REFACTOR → `03146d1ae` merge) and plan-phase artifacts.
- Provides plan-auditor-readable AC index that maps 1:1 to existing `internal/spec/lint_test.go` test functions.
- Updates `spec.md` frontmatter `status: draft → implemented` and `version: 0.1.0 → 0.1.1`.
- Documents external research citations (EARS, Tarjan SCC, SARIF 2.1.0, eslint/golangci-lint UX) post-hoc.

## Changes

### Files added (6 new artifacts)

- `.moai/specs/SPEC-V3R2-SPC-003/research.md` — External references (EARS Mavin 2009, Tarjan 1972, SARIF OASIS 2.1.0, golangci-lint conventions) + sister-SPEC dependencies (CON-001, SPC-001) + 9-rule-vs-10-rule reconciliation (progress.md said 9, code has 10 = 8 per-SPEC + 2 cross-SPEC).
- `.moai/specs/SPEC-V3R2-SPC-003/plan.md` — Retroactive M1-M4 milestones reflecting actual Wave 5 commit chain.
- `.moai/specs/SPEC-V3R2-SPC-003/acceptance.md` — 16 ACs in Given-When-Then format. Each AC maps 1:1 to a `TestLinter_AC{NN}_*` test in `internal/spec/lint_test.go`.
- `.moai/specs/SPEC-V3R2-SPC-003/tasks.md` — 14 tasks (T-SPC003-01..14) reflecting actual Wave 5 work, each task references the actual commit SHA.
- `.moai/specs/SPEC-V3R2-SPC-003/spec-compact.md` — Compact reference for plan-auditor scans and cross-SPEC referencing.
- `.moai/specs/SPEC-V3R2-SPC-003/issue-body.md` — This file.

### Files modified (2)

- `.moai/specs/SPEC-V3R2-SPC-003/spec.md` — frontmatter `status: draft → implemented`, `version: 0.1.0 → 0.1.1`, `updated: 2026-04-23 → 2026-05-10`. HISTORY entry added. Body text unchanged.
- `.moai/specs/SPEC-V3R2-SPC-003/progress.md` — HISTORY section added. AC table extended with RED/GREEN commit SHA columns. Backfill note appended.

### Files NOT modified (FROZEN post-merge)

- `internal/spec/lint.go` (816 LOC)
- `internal/spec/lint_test.go` (572 LOC)
- `internal/spec/dag.go` (100 LOC)
- `internal/spec/sarif.go` (165 LOC)
- `internal/cli/spec_lint.go` (164 LOC)
- `internal/spec/testdata/` (11 fixture SPEC directories)
- `internal/cli/spec.go` (registration line)

These files are FROZEN as they were merged via PR #745. This backfill PR touches `.moai/specs/SPEC-V3R2-SPC-003/` only.

## AC Index (16 ACs, all PASS)

| AC ID | Test Function | Status |
|-------|---------------|--------|
| AC-V3R2-SPC-003-01 | `TestLinter_AC01_HappyPath` | PASS |
| AC-V3R2-SPC-003-02 | `TestLinter_AC02_CoverageIncomplete` | PASS |
| AC-V3R2-SPC-003-03 | `TestLinter_AC03_ModalityMalformed` | PASS |
| AC-V3R2-SPC-003-04 | `TestLinter_AC04_DependencyCycle` | PASS |
| AC-V3R2-SPC-003-05 | `TestLinter_AC05_DuplicateREQID` | PASS |
| AC-V3R2-SPC-003-06 | `TestLinter_AC06_MissingExclusions` | PASS |
| AC-V3R2-SPC-003-07 | `TestLinter_AC07_MissingDependency` | PASS |
| AC-V3R2-SPC-003-08 | `TestLinter_AC08_DanglingRuleReference` | PASS |
| AC-V3R2-SPC-003-09 | `TestLinter_AC09_JSONOutput` | PASS |
| AC-V3R2-SPC-003-10 | `TestLinter_AC10_SARIFOutput` | PASS |
| AC-V3R2-SPC-003-11 | `TestLinter_AC11_StrictMode` | PASS |
| AC-V3R2-SPC-003-12 | `TestLinter_AC12_DuplicateSPECID` | PASS |
| AC-V3R2-SPC-003-13 | `TestLinter_AC13_LintSkip` | PASS |
| AC-V3R2-SPC-003-14 | `TestLinter_AC14_BreakingChangeMissingID` | PASS |
| AC-V3R2-SPC-003-15 | `TestLinter_AC15_ParseFailure` | PASS |
| AC-V3R2-SPC-003-16 | `TestLinter_AC16_HierarchicalACCoverage` | PASS |

## Implementation Already Merged

- **Original PR**: #745 (Wave 5 batch)
- **Merge commit**: `03146d1ae feat(spec): SPEC-V3R2-SPC-003 — moai spec lint CLI (Wave 5) (#745)`
- **Merge date**: 2026-04-30 17:51 KST
- **Pre-squash commit chain**:
  - `44019ec4d test(spec): SPEC-V3R2-SPC-003 RED phase — failing lint tests`
  - `4d5699f27 feat(spec): SPEC-V3R2-SPC-003 GREEN — moai spec lint linter engine`
  - `1587e6e3a refactor(spec): SPEC-V3R2-SPC-003 REFACTOR — Karpathy simplicity + MX tags`
- **Coverage**: `internal/spec` package 86.6% (target ≥85%)
- **Lint gates**: `go vet` 0 issues, `golangci-lint run` 0 issues, `go test -race` clean

## Verification

After merging this backfill PR, the following commands verify SPEC-First compliance:

```bash
# Status check
head -15 .moai/specs/SPEC-V3R2-SPC-003/spec.md | grep status
# expected: status: implemented

# 16 ACs in acceptance.md
grep -cE "AC-V3R2-SPC-003-[0-9]+" .moai/specs/SPEC-V3R2-SPC-003/acceptance.md
# expected: ≥16

# 1:1 mapping check (acceptance.md AC IDs ↔ lint_test.go test functions)
grep -oE "TestLinter_AC[0-9]+_\w+" internal/spec/lint_test.go | sort -u | wc -l
# expected: 16

# Self-lint: spec.md should pass moai spec lint
moai spec lint .moai/specs/SPEC-V3R2-SPC-003/spec.md
# expected: 0 errors
```

## Reference

- **PR #745 (original implementation)**: Wave 5 batch merge.
- **Sister SPECs**:
  - SPEC-V3R2-CON-001 (Constitution Zone Registry, consumed by `ZoneRegistryRule`)
  - SPEC-V3R2-SPC-001 (Hierarchical Acceptance parser, source of canonical AC tree)
  - SPEC-V3R2-SPC-002 (@MX TAG validator, NOT this linter's scope)
- **Memory**: `project_wave5_complete.md` (Wave 5 batch context).
- **External**:
  - Mavin et al. 2009 (EARS), Tarjan 1972 (SCC), Sedgewick *Algorithms* 4e, CLRS 3e §22.5
  - OASIS SARIF 2.1.0: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
  - ESLint and golangci-lint UX conventions (severity tiers, code identifiers, suppress mechanisms)

## Test Plan

- [x] `head -15 .moai/specs/SPEC-V3R2-SPC-003/spec.md | grep status` returns `status: implemented`
- [x] `grep -cE "AC-V3R2-SPC-003-[0-9]+" .moai/specs/SPEC-V3R2-SPC-003/acceptance.md` ≥ 16
- [x] Every AC ID in `acceptance.md` corresponds to a `TestLinter_AC{NN}_*` function in `internal/spec/lint_test.go`
- [x] No code or test files modified (verified via `git diff --stat`)
- [x] CI all-green (Lint / Test ubuntu/macos/windows / Build 5 / CodeQL — backfill is docs-only, should not impact)

🗿 MoAI <email@mo.ai.kr>
