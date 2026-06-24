---
id: SPEC-CLEANUP-EVALUATOR-001
title: "Acceptance criteria — remove orphaned internal/evaluator package"
version: "0.1.0"
status: draft
created: 2026-06-22
updated: 2026-06-22
author: Goos Kim
---

## §D. Acceptance Criteria Matrix

| AC ID | Requirement | Verification command | Pass condition |
|-------|-------------|----------------------|----------------|
| AC-CLEANUP-EVAL-001 | REQ-CLEANUP-EVAL-001 | `ls internal/evaluator/ 2>/dev/null; echo $?` | Directory absent (non-zero exit / no listing) |
| AC-CLEANUP-EVAL-002 | REQ-CLEANUP-EVAL-002 | `go build ./...` | Exit 0, no errors |
| AC-CLEANUP-EVAL-003 | REQ-CLEANUP-EVAL-003 | `go test ./...` | Exit 0, no new failures |
| AC-CLEANUP-EVAL-004 | REQ-CLEANUP-EVAL-004 | `grep -rn 'internal/evaluator' --include='*.go' .` | No matches (empty output) |
| AC-CLEANUP-EVAL-005 | REQ-CLEANUP-EVAL-005 | `go list ./internal/evaluator/ 2>&1` | Package reported absent (error / no package) |
| AC-CLEANUP-EVAL-006a | REQ-CLEANUP-EVAL-006 | `grep -n 'internal/evaluator' .moai/project/codemaps/modules.md` | No stale `internal/evaluator` modules-list entry remains |
| AC-CLEANUP-EVAL-006b | REQ-CLEANUP-EVAL-006 | `grep -n 'internal/evaluator' .moai/project/codemaps/overview.md` | No stale `internal/evaluator` parenthetical remains |
| AC-CLEANUP-EVAL-007 | REQ-CLEANUP-EVAL-007 | manual read of `overview.md` package-count line | Test-only count reads 1 (only `internal/skills`); total `internal` count decremented consistently |
| AC-CLEANUP-EVAL-008 | §E Exclusions | `git status --porcelain` | No modification to any §E-Exclusions path |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Package removed, build and tests still pass

```
GIVEN the repository with internal/evaluator/ present (one file, zero importers)
WHEN the internal/evaluator/ directory is removed
 AND go build ./... is executed
 AND go test ./... is executed
THEN go build ./... exits 0 with no errors
 AND go test ./... exits 0 with no new failures
 AND go list ./internal/evaluator/ reports the package as absent
```

### Scenario 2 — No dangling references remain after removal

```
GIVEN the internal/evaluator/ directory has been removed
WHEN grep -rn 'internal/evaluator' --include='*.go' . is executed over the repository
THEN the command returns no matches (the import path is referenced nowhere in Go source)
```

### Scenario 3 — Codemaps references synchronized and arithmetically consistent

```
GIVEN .moai/project/codemaps/modules.md and overview.md each reference internal/evaluator
WHEN the 2 references are updated to reflect the removal
THEN the modules.md modules list no longer carries a live internal/evaluator entry
 AND the overview.md "test-only 패키지" count reads 1 (only internal/skills)
 AND the overview.md total internal directory count is decremented for consistency
```

### Scenario 4 — Exclusions untouched (negative scenario)

```
GIVEN the removal and codemaps edits are complete
WHEN git status --porcelain is inspected
THEN no path under .moai/specs/SPEC-EVAL-001/, .moai/specs/SPEC-EVALLIB-001/,
     .claude/agents/moai/sync-auditor.md, .moai/config/evaluator-profiles/,
     or internal/skills/ appears as modified
```

## §D.2 Edge Cases

- **Stale build cache**: if `go list ./internal/evaluator/` is cached, run with
  `go list -e ./internal/evaluator/` or after `go clean -cache` to confirm absence.
- **codemaps regenerated this session**: the 5 codemaps files are uncommitted in the
  working tree; the 2 edits must be applied on top of the regenerated content, not
  reverted onto a stale baseline.
- **grep false-positive guard**: AC-CLEANUP-EVAL-004 uses `--include='*.go'` so that
  the SPEC's own documentation references to `internal/evaluator` (in spec.md /
  plan.md / acceptance.md) do NOT count as dangling Go-source references.

## §D.3 Quality Gate Criteria

- All 9 AC rows PASS.
- LSP / lint baseline: zero new errors introduced (none expected — no production
  source changes).
- No path under the §E Exclusions list modified.

## §D.4 Definition of Done

- [ ] `internal/evaluator/` directory removed.
- [ ] `go build ./...` succeeds.
- [ ] `go test ./...` succeeds.
- [ ] `grep -rn 'internal/evaluator' --include='*.go'` returns empty.
- [ ] `go list ./internal/evaluator/` reports the package absent.
- [ ] `.moai/project/codemaps/modules.md` reference updated.
- [ ] `.moai/project/codemaps/overview.md` reference + count arithmetic updated.
- [ ] `git status` confirms no §E-Exclusions path was touched.
