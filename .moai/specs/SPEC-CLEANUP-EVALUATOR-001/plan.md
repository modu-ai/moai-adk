---
id: SPEC-CLEANUP-EVALUATOR-001
title: "Implementation plan — remove orphaned internal/evaluator package"
version: "0.1.0"
status: draft
created: 2026-06-22
updated: 2026-06-22
author: Goos Kim
---

## §A. Context

This plan implements the removal of the orphaned `internal/evaluator/` package and the
synchronization of its 2 codemaps references, per `spec.md` REQ-CLEANUP-EVAL-001
through REQ-CLEANUP-EVAL-007.

The package is a TDD RED-phase scaffold (`SPEC-EVAL-001 M1`) that never advanced to
GREEN. The feature was delivered as the `sync-auditor` agent definition instead. The
scaffold has zero importers and contains only non-functional stubs.

## §B. Tier Justification — Tier S (minimal)

This SPEC is classified **Tier S (minimal)** for the following reasons:

- **Single-unit change**: one directory removal (one file) plus 2 documentation-line
  edits in the same codemaps cohort. No production source is added or modified.
- **Zero blast radius**: the package has zero importers (verified:
  `grep -rn 'internal/evaluator' --include='*.go'` returns no matches outside the
  package). Removal cannot break a build or a passing real test — the 5 stub tests
  only pass or skip.
- **No design surface**: there is no architecture decision, no API contract, no new
  abstraction. A `design.md` / `tasks.md` 4-file structure is NOT warranted; the 3-file
  structure (spec/plan/acceptance) suffices.
- **Mechanical verification**: the acceptance criteria are all mechanically checkable
  with `go build`, `go test`, `go list`, and `grep` — no judgment-heavy review needed.

Tier S minimal harness routing applies: fast validation, no sync-auditor 4-dimension
scoring required for the change itself (the orchestrator may still run standard gates).

## §C. Pre-flight (verified grounding)

The orchestrator verified the following this session (re-confirmable at run-phase):

- `ls internal/evaluator/` → exactly one file: `evaluator_test.go` (49 lines).
- No production `.go` file in the package (package clause: `package evaluator_test`).
- `grep -rn 'internal/evaluator' --include='*.go' .` → no matches (zero importers).
- `go test ./internal/evaluator/` → ok (stubs pass/skip).
- `go vet ./internal/evaluator/` → clean.
- `grep -rn 'internal/evaluator' .moai/project/codemaps/` → 2 references:
  - `.moai/project/codemaps/modules.md:235`
  - `.moai/project/codemaps/overview.md:8`

## §D. Constraints

- No production Go source added or modified.
- Only the 2 named codemaps references are edited (no full codemaps regeneration).
- No file under the §E Exclusions list (spec.md) is touched.

## §E. Self-Verification

### §E.1 Plan-phase Audit-Ready Signal

- SPEC ID validated against canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (PASS).
- Frontmatter 12-field schema validated for spec.md (PASS).
- Out of Scope section present with 5 `### Out of Scope — <topic>` H3 sub-headings,
  each carrying `-` bullets (satisfies `OutOfScopeRule`).
- All requirements expressed in GEARS notation (Ubiquitous / State-driven /
  Event-detected).
- Tier S justification recorded (§B).
- Grounding pre-verified and re-confirmable (§C).

### §E.2 Run-phase Evidence

_<pending run-phase>_

### §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Remove the package (Priority High)

- Delete the `internal/evaluator/` directory and its sole file `evaluator_test.go`.
- Satisfies REQ-CLEANUP-EVAL-001.

### M2 — Synchronize codemaps references (Priority High)

- Update `.moai/project/codemaps/modules.md` (remove or amend the `internal/evaluator`
  line in the modules list).
- Update `.moai/project/codemaps/overview.md` (decrement the test-only package count
  from 2 to 1, remove the `internal/evaluator` parenthetical, and adjust the total
  `internal` directory count for arithmetic consistency).
- Satisfies REQ-CLEANUP-EVAL-006 and REQ-CLEANUP-EVAL-007.

### M3 — Verify (Priority High)

- `go build ./...` → success (REQ-CLEANUP-EVAL-002).
- `go test ./...` → success (REQ-CLEANUP-EVAL-003).
- `grep -rn 'internal/evaluator' --include='*.go'` → empty (REQ-CLEANUP-EVAL-004).
- `go list ./internal/evaluator/` → package absent (REQ-CLEANUP-EVAL-005).
- Confirm no §E-Exclusions path was modified (`git status` review).

## §G. Anti-Patterns to Avoid

- Removing or editing any SPEC-EVAL-001 / SPEC-EVALLIB-001 history artifact.
- Touching `.claude/agents/moai/sync-auditor.md` or `.moai/config/evaluator-profiles/`.
- Confusing `internal/evaluator` (this target) with `internal/skills` (a separate
  ACTIVE test-fixture package that MUST be preserved).
- Regenerating the entire codemaps when only 2 references need a consistency fixup.
- Leaving the overview.md package-count arithmetic internally inconsistent after the
  removal.

## §H. Cross-References

- `spec.md` §C Requirements (REQ-CLEANUP-EVAL-001..007).
- `acceptance.md` AC matrix.
- `.claude/agents/moai/sync-auditor.md` — the active delivery that superseded the Go
  scaffold (read-only context; NOT a modification target).
- `internal/evaluator/evaluator_test.go` — the removal target.
