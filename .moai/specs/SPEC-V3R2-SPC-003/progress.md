# SPEC-V3R2-SPC-003 Implementation Progress

## Status: COMPLETED

## Completed ACs

| AC | Test Name | Status |
|----|-----------|--------|
| AC-SPC-003-01 | TestLinter_AC01_HappyPath | PASS |
| AC-SPC-003-02 | TestLinter_AC02_CoverageIncomplete | PASS |
| AC-SPC-003-03 | TestLinter_AC03_ModalityMalformed | PASS |
| AC-SPC-003-04 | TestLinter_AC04_DependencyCycle | PASS |
| AC-SPC-003-05 | TestLinter_AC05_DuplicateREQID | PASS |
| AC-SPC-003-06 | TestLinter_AC06_MissingExclusions | PASS |
| AC-SPC-003-07 | TestLinter_AC07_MissingDependency | PASS |
| AC-SPC-003-08 | TestLinter_AC08_DanglingRuleReference | PASS |
| AC-SPC-003-09 | TestLinter_AC09_JSONOutput | PASS |
| AC-SPC-003-10 | TestLinter_AC10_SARIFOutput | PASS |
| AC-SPC-003-11 | TestLinter_AC11_StrictMode | PASS |
| AC-SPC-003-12 | TestLinter_AC12_DuplicateSPECID | PASS |
| AC-SPC-003-13 | TestLinter_AC13_LintSkip | PASS |
| AC-SPC-003-14 | TestLinter_AC14_BreakingChangeMissingID | PASS |
| AC-SPC-003-15 | TestLinter_AC15_ParseFailure | PASS |
| AC-SPC-003-16 | TestLinter_AC16_HierarchicalACCoverage | PASS |

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
