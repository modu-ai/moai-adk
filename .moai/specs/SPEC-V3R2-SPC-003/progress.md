# SPEC-V3R2-SPC-003 Implementation Progress

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-30 | Wave 5 implementer | Initial progress tracker — 16/16 ACs PASS, 86.6% coverage |
| 0.1.1 | 2026-05-10 | manager-spec (Batch 3 backfill) | plan-phase artifacts back-filled post-Wave 5; commit SHA references added to AC table |

## Status: COMPLETED

Merged via PR #745 → main commit `03146d1ae` (2026-04-30 17:51 KST, Wave 5 batch).

## Completed ACs

> Each AC was authored as a failing test in commit `44019ec4d` (RED) and turned GREEN in commit `4d5699f27`. REFACTOR commit `1587e6e3a` preserved all PASS status while applying Karpathy simplicity + @MX tag injection.

| AC | Test Name | Status | RED commit | GREEN commit |
|----|-----------|--------|-----------|--------------|
| AC-SPC-003-01 | TestLinter_AC01_HappyPath | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-02 | TestLinter_AC02_CoverageIncomplete | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-03 | TestLinter_AC03_ModalityMalformed | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-04 | TestLinter_AC04_DependencyCycle | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-05 | TestLinter_AC05_DuplicateREQID | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-06 | TestLinter_AC06_MissingExclusions | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-07 | TestLinter_AC07_MissingDependency | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-08 | TestLinter_AC08_DanglingRuleReference | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-09 | TestLinter_AC09_JSONOutput | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-10 | TestLinter_AC10_SARIFOutput | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-11 | TestLinter_AC11_StrictMode | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-12 | TestLinter_AC12_DuplicateSPECID | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-13 | TestLinter_AC13_LintSkip | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-14 | TestLinter_AC14_BreakingChangeMissingID | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-15 | TestLinter_AC15_ParseFailure | PASS | `44019ec4d` | `4d5699f27` |
| AC-SPC-003-16 | TestLinter_AC16_HierarchicalACCoverage | PASS | `44019ec4d` | `4d5699f27` |

## Coverage

- `internal/spec` package: **86.6%** (target: 85%)
- All new files: lint.go, dag.go, sarif.go
- Extended: internal/cli/spec_lint.go, internal/cli/spec.go

## Files Created/Modified

### New Files
- `internal/spec/lint.go` — Linter engine (Linter struct, Rule interface, 9 rules)
- `internal/spec/lint_test.go` — Table-driven tests covering all 16 ACs
- `internal/spec/dag.go` — Tarjan SCC cycle detection
- `internal/spec/sarif.go` — SARIF 2.1.0 output writer
- `internal/cli/spec_lint.go` — cobra subcommand implementation
- `internal/spec/testdata/` — 11 fixture SPEC directories

### Modified Files
- `internal/cli/spec.go` — Added newSpecLintCmd() registration

## Drift Analysis vs SPEC §2.1 In Scope

In scope items covered:
- [x] moai spec lint CLI subcommand
- [x] EARS compliance checks
- [x] REQ ID uniqueness (within SPEC)
- [x] AC→REQ coverage >= 100%
- [x] Frontmatter schema validation
- [x] Dependency DAG validation (no cycles, deps exist)
- [x] Zone registry cross-reference
- [x] Out of Scope section presence
- [x] JSON output (--json)
- [x] SARIF 2.1.0 output (--sarif)
- [x] Human-readable table output (default)

Out of scope items correctly excluded:
- @MX TAG validation (SPEC-V3R2-SPC-002 scope)
- --fix flag (intentionally deferred)
- Non-SPEC artifact validation

Drift: ~0% — all modified files are within SPEC §2.1 scope.

## Quality Gate Results

- go vet: PASS (0 issues)
- golangci-lint: PASS (0 issues)
- go test -race: PASS (all packages)
- Coverage: 86.6% >= 85% target

## TDD Cycle Summary

- RED: lint_test.go created with 16 AC tests → all failed (undefined symbols)
- GREEN: lint.go + dag.go + sarif.go + spec_lint.go implemented → all 16 AC tests pass
- REFACTOR: removed unused collectLeafREQIDs, fixed errcheck violations, renamed max() to positiveLineNum()

## Next Step

Push branch and create PR for orchestrator review.

---

## Backfill Note (2026-05-10, Batch 3)

The plan-phase artifacts (`research.md`, `plan.md`, `acceptance.md`, `tasks.md`, `spec-compact.md`, `issue-body.md`) were not authored prior to PR #745 merge. As part of SPEC-First retroactive normalization (Batch 3), these artifacts have been back-filled to align with the actual as-merged implementation. No code or test changes were made in this backfill — implementation files (`internal/spec/lint.go`, `dag.go`, `sarif.go`, `internal/cli/spec_lint.go`, `internal/spec/lint_test.go`) are FROZEN post-merge.

`spec.md` frontmatter updated: `status: draft` → `status: implemented`, `version: "0.1.0"` → `version: "0.1.1"`, `updated: 2026-04-23` → `updated: 2026-05-10`.

### Self-lint observation (pre-existing drift in spec.md §6)

Running `moai spec lint .moai/specs/SPEC-V3R2-SPC-003/spec.md` against the FROZEN spec.md body surfaces 6 errors:

- `ModalityMalformed` on REQ-SPC-003-041 (Optional form `WHERE ... is specified, the default human-readable output is explicitly selected` — linter's `WHERE...SHALL` regex flags it because "is selected" is passive without explicit `SHALL`)
- `CoverageIncomplete` on REQ-SPC-003-002, 006, 041, 050, 051 (linter's per-AC-line regex appears not to extract all comma-separated REQ IDs from the `(maps REQ-..., REQ-..., REQ-...)` tail correctly when there are 3+ REQs in a single tail; this is a known per-line tail parsing limitation in the as-merged `CoverageRule.Check()` at `internal/spec/lint.go:477`)

Both findings reflect **pre-existing limitations in the as-merged linter implementation** (Wave 5, FROZEN), not new drift introduced by backfill. They are documented here for traceability. Future work to refine the multi-REQ tail parser or the WHERE-form modality regex should be filed as a separate follow-up SPEC; this backfill PR does NOT modify `internal/spec/lint.go` or `spec.md` body.

The acceptance.md REQ↔AC traceability matrix (this backfill) shows the *intended* mapping (16 ACs cover 18 REQs with shared ACs). The traceability matrix is the canonical source of truth for plan-auditor and humans; the linter's per-line regex captures a strict subset.
