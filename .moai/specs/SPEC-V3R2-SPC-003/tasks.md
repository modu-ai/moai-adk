# SPEC-V3R2-SPC-003 Task Breakdown (BACKFILL)

> Granular task decomposition matching the as-merged Wave 5 implementation (PR #745).
> Companion to `spec.md` v0.1.1, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-10 | manager-spec (Batch 3 backfill) | Retroactive task breakdown — 14 tasks (T-SPC003-01..14) reflecting actual Wave 5 commits |

---

## Task ID Convention

- ID format: `T-SPC003-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `expert-backend` (Go), `manager-git` (commit/PR boundary)
- Phase mapping: M1=RED, M2=GREEN, M3=REFACTOR, M4=Merge
- Each task references the actual Wave 5 commit SHA where the work was performed

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies + commit traceability only.

---

## M1: RED phase (failing tests) — Priority P0

Goal: Author 16 failing AC tests + fixture data per TDD discipline. Per `spec-workflow.md` TDD: write failing tests first.

Anchor commit: **`44019ec4d test(spec): SPEC-V3R2-SPC-003 RED phase — failing lint tests`**

| ID | Subject | Owner role | File:line target | Dependency | TDD alignment |
|----|---------|-----------|-------------------|------------|---------------|
| T-SPC003-01 | Create `internal/spec/lint_test.go` skeleton with `TestLinter_AC01_HappyPath` ~ `TestLinter_AC16_HierarchicalACCoverage` (16 table-driven test functions). All reference yet-undefined `Linter`, `LinterOptions`, `NewLinter`, `Report`, `Finding` symbols. | expert-backend | `internal/spec/lint_test.go` (new, 572 LOC) | none | RED — fails today (undefined symbols) |
| T-SPC003-02 | Create `internal/spec/testdata/` with 11 fixture SPEC directories covering: happy-path, coverage-incomplete, modality-malformed, dependency-cycle (2 specs), duplicate-req-id, missing-exclusions, missing-dependency, dangling-rule, duplicate-spec-id (2 specs), lint-skip, breaking-change-missing-id, parse-failure (broken YAML), hierarchical-ac. | expert-backend | `internal/spec/testdata/` (new, 11 dirs) | T-SPC003-01 | RED — fixtures consumed by tests |
| T-SPC003-03 | Run `go test ./internal/spec/...` and confirm 16/16 RED state (compile error or run failure). Existing tests in other packages remain GREEN (regression sentinel). | manager-tdd | n/a (verification only) | T-SPC003-01, T-SPC003-02 | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones. TDD discipline.

---

## M2: GREEN phase (linter engine implementation) — Priority P0

Goal: Implement `Linter`, 10 rules, parser, 3 output formats so all 16 RED tests turn GREEN.

Anchor commit: **`4d5699f27 feat(spec): SPEC-V3R2-SPC-003 GREEN — moai spec lint linter engine`**

| ID | Subject | Owner role | File:line target | Dependency | TDD alignment |
|----|---------|-----------|-------------------|------------|---------------|
| T-SPC003-04 | Create `internal/spec/lint.go` core types: `Severity` (string enum), `Finding` struct (`File, Line, Severity, Code, Message`), `Report` struct (`Findings`, `HasErrors()`, `ToJSON()`, `ToSARIF()` methods), `LinterOptions`, `Linter` struct, `NewLinter()`. | expert-backend | `internal/spec/lint.go:20-135` | T-SPC003-03 | GREEN T-SPC003-01 (compile) |
| T-SPC003-05 | Implement `Linter.Lint(paths []string) (*Report, error)` orchestration: discovery → per-SPEC parse → per-SPEC rules → cross-SPEC rules → lint.skip filter → return Report. | expert-backend | `internal/spec/lint.go:137-225` | T-SPC003-04 | GREEN AC-01 |
| T-SPC003-06 | Implement parser: `parseSPECDoc()`, `extractFrontmatter()` (gopkg.in/yaml.v3), `parseREQs()`, `collectAllREQIDs()`, `SPECFrontmatter` struct (with validator/v10 tags), `REQEntry` struct, `SPECDoc` struct. | expert-backend | `internal/spec/lint.go:255-390` | T-SPC003-04 | GREEN AC-01,15 (parsing) |
| T-SPC003-07 | Implement 8 per-SPEC rules: `EARSModalityRule`, `REQIDUniquenessRule`, `CoverageRule`, `FrontmatterSchemaRule`, `DependencyExistsRule`, `OutOfScopeRule`, `BreakingChangeIDRule`, `ZoneRegistryRule`. Each implements `Rule` interface (`Code() string`, `Check(doc, allDocs) []Finding`). | expert-backend | `internal/spec/lint.go:394-728` | T-SPC003-06 | GREEN AC-02,03,05,06,07,08,14,16 |
| T-SPC003-08 | Implement `internal/spec/dag.go`: Tarjan SCC iterative algorithm. `DependencyCycleRule` (cross-SPEC) calls into dag.go to detect cycles. | expert-backend | `internal/spec/dag.go` (new, 100 LOC) + `lint.go:730-784` | T-SPC003-06 | GREEN AC-04 |
| T-SPC003-09 | Implement `DuplicateSPECIDRule` (cross-SPEC): scans all loaded `SPECDoc.Frontmatter.ID` and reports duplicates with both file paths. | expert-backend | `internal/spec/lint.go:786-816` | T-SPC003-06 | GREEN AC-12 |
| T-SPC003-10 | Implement `internal/spec/sarif.go`: SARIF 2.1.0 writer. `Report.ToSARIF()` builds `runs[0].tool.driver` + `runs[0].results[]` with `ruleId`, `level`, `message`, `locations[]`. JSON-marshalled. | expert-backend | `internal/spec/sarif.go` (new, 165 LOC) | T-SPC003-04 | GREEN AC-10 |
| T-SPC003-11 | Implement `internal/cli/spec_lint.go`: cobra `newSpecLintCmd()` with flags `--json`, `--sarif`, `--strict`, `--format=table`. Wire to `internal/cli/spec.go` via `newSpecLintCmd()` registration (one-line addition). | expert-backend | `internal/cli/spec_lint.go` (new, 164 LOC) + `internal/cli/spec.go` (1-line edit) | T-SPC003-04 | GREEN AC-09,11 (CLI flags) |
| T-SPC003-12 | Run `go test ./internal/spec/... ./internal/cli/...` and confirm 16/16 GREEN. Run `go test ./...` regression check. | manager-tdd | n/a (verification only) | T-SPC003-04..11 | GREEN gate verification |

**M2 priority: P0** — implements all 10 rules + CLI subcommand.

---

## M3: REFACTOR phase (Karpathy simplicity + MX tags) — Priority P1

Goal: Remove dead code, fix lint warnings, inject @MX annotations, satisfy quality gate.

Anchor commit: **`1587e6e3a refactor(spec): SPEC-V3R2-SPC-003 REFACTOR — Karpathy simplicity + MX tags`**

| ID | Subject | Owner role | File:line target | Dependency | TDD alignment |
|----|---------|-----------|-------------------|------------|---------------|
| T-SPC003-13a | Remove unused `collectLeafREQIDs()` function (Karpathy "delete unused code" principle). Verify no callers via grep. | expert-backend | `internal/spec/lint.go` (deletion) | T-SPC003-12 | REFACTOR — behavior-preserving |
| T-SPC003-13b | Fix 7 `errcheck` violations: `fmt.Fprintf` / `fmt.Fprintln` return values. Add `_, _ =` discard or proper error handling. | expert-backend | `internal/spec/lint.go`, `internal/spec/sarif.go` (multi-line edits) | T-SPC003-12 | REFACTOR — lint clean |
| T-SPC003-13c | Rename `sarif.go` `max()` → `positiveLineNum()` to avoid shadowing Go 1.21+ built-in `max()`. | expert-backend | `internal/spec/sarif.go` (rename) | T-SPC003-12 | REFACTOR — name conflict resolution |
| T-SPC003-13d | Inject @MX annotations: `@MX:NOTE` for intent disclosure on `Linter.Lint()`, `Rule` interface, and `parseSPECDoc()`; `@MX:ANCHOR` for high-fan-in functions (`parseREQs()`, `extractFrontmatter()`). | expert-backend | `internal/spec/lint.go` (comments) | T-SPC003-13a..c | REFACTOR — observability |
| T-SPC003-14 | Update `progress.md` with final status: 16/16 ACs PASS, coverage 86.6%, files created/modified inventory, drift analysis, quality gate results, TDD cycle summary. | manager-tdd | `.moai/specs/SPEC-V3R2-SPC-003/progress.md` | T-SPC003-13a..d | M3 gate completion |

**M3 priority: P1** — quality gate + observability. Must run `golangci-lint run` 0 issues + `go vet` clean before proceeding.

---

## M4: PR + Wave 5 squash merge — Priority P0

Goal: Open PR for Wave 5 batch, get reviewer approval, squash merge to main.

Anchor commit: **`03146d1ae feat(spec): SPEC-V3R2-SPC-003 — moai spec lint CLI (Wave 5) (#745)`** (squashed M1+M2+M3)

| ID | Subject | Owner role | Action | Dependency |
|----|---------|-----------|--------|------------|
| (post-M3) | Push branch + open PR #745 to `main` with M1/M2/M3 commits. CI all-green. Squash merge into main as Wave 5 batch entry. | manager-git | `gh pr create` + `gh pr merge --squash` | T-SPC003-14 |

> Note: M4 is a process step (not a code task) and is not assigned a `T-SPC003-NN` ID. Reference commit `03146d1ae` is the squash-merge result of M1+M2+M3 commits. This task list intentionally has 14 task IDs (T-SPC003-01 through T-SPC003-14) since post-merge process is owned by manager-git and tracked in PR #745 metadata.

---

## Coverage Summary

| Phase | Tasks | Commit | Status |
|-------|-------|--------|--------|
| M1 (RED) | T-SPC003-01..03 | `44019ec4d` | Completed pre-merge |
| M2 (GREEN) | T-SPC003-04..12 | `4d5699f27` | Completed pre-merge |
| M3 (REFACTOR) | T-SPC003-13a..d, T-SPC003-14 | `1587e6e3a` | Completed pre-merge |
| M4 (Merge) | (process) | `03146d1ae` | Merged 2026-04-30 |

**Total tasks: 14** (T-SPC003-01..14, with T-SPC003-13 split into 4 sub-letters).
**Status (as of 2026-05-10)**: All tasks completed. Implementation FROZEN post-PR-#745 merge.

---

## Files Created/Modified (실측, post-merge)

### New files
- `internal/spec/lint.go` (816 LOC after REFACTOR)
- `internal/spec/lint_test.go` (572 LOC)
- `internal/spec/dag.go` (100 LOC)
- `internal/spec/sarif.go` (165 LOC)
- `internal/cli/spec_lint.go` (164 LOC)
- `internal/spec/testdata/` (11 fixture SPEC directories)

### Modified files
- `internal/cli/spec.go` (newSpecLintCmd registration, ~1 LOC)

---

End of tasks.md (backfill v0.1.0).
