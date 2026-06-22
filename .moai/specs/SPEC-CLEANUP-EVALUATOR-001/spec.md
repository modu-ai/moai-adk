---
id: SPEC-CLEANUP-EVALUATOR-001
title: "Remove orphaned internal/evaluator Go package (abandoned TDD RED scaffold)"
version: "0.1.0"
status: completed
created: 2026-06-22
updated: 2026-06-22
author: Goos Kim
priority: P3
phase: "v3.0.0"
module: "internal/evaluator"
lifecycle: spec-anchored
tier: S
tags: "cleanup, housekeeping, dead-code, evaluator, sync-auditor"
related_specs: [SPEC-EVAL-001, SPEC-EVALLIB-001]
---

## HISTORY

- 2026-06-22 — draft created (manager-spec). Plan-phase authoring of a Tier S cleanup SPEC for the orphaned `internal/evaluator` package.

## §A. Context and Motivation

`internal/evaluator/` contains a single file, `evaluator_test.go` (49 lines, package
`evaluator_test`), with NO production `.go` file. Its header comment records its
origin: "SPEC-EVAL-001 M1 RED phase: sync-auditor agent scaffolding". The package
holds 5 test functions, all non-functional stubs:

- `TestEvaluatorAgent_ScoreDimensions` — a string-only table that asserts nothing real.
- `TestEvaluatorAgent_ScoreRange` — `t.Skip("M2: implement score range validation")`.
- `TestEvaluatorAgent_MustPassCriteria` — `t.Skip("M3: implement must-pass firewall")`.
- `TestEvaluatorAgent_SprintContract` — `t.Skip("M4: implement sprint contract protocol")`.
- `TestEvaluatorAgent_DebiasAnchoring` — `t.Skip("M5: implement rubric anchoring")`.

The feature this scaffold targeted (SPEC-EVAL-001, "evaluator-active Agent & Sprint
Contract", `status: implemented`, `phase: "v2.x - Legacy"`) was delivered as an AGENT
DEFINITION, not as Go code. The active implementation is the `sync-auditor` agent at
`.claude/agents/moai/sync-auditor.md`, retained in the 8-agent catalog. The Go RED
scaffold never advanced to GREEN (M2-M5 were never implemented) and is superseded by
the agent-based delivery.

The package has ZERO importers. Nothing in the module references `internal/evaluator`
outside the package itself. Removing it cannot break a build or a passing real test —
the stubs only pass or skip. The package is dead scaffolding occupying directory space
and adding a misleading "test-only package" entry to the codemaps.

## §B. Objective

Remove the orphaned `internal/evaluator/` directory and synchronize the 2 codemaps
references so the project's documented package inventory reflects the removal. No
behavior changes anywhere; no other package touched.

## §C. Requirements (GEARS)

### REQ-CLEANUP-EVAL-001 — Package removal (Ubiquitous)

The cleanup change **shall** remove the `internal/evaluator/` directory in its entirety,
including its sole file `evaluator_test.go`.

### REQ-CLEANUP-EVAL-002 — Build integrity preserved (State-driven)

**While** the module is built after removal, `go build ./...` **shall** complete
successfully with no new errors attributable to the removal.

### REQ-CLEANUP-EVAL-003 — Test suite integrity preserved (State-driven)

**While** the test suite is executed after removal, `go test ./...` **shall** complete
successfully with no new failures attributable to the removal.

### REQ-CLEANUP-EVAL-004 — No dangling references (Ubiquitous)

After removal, the codebase **shall** contain no Go-source reference to the removed
import path: `grep -rn 'internal/evaluator' --include='*.go'` over the repository
**shall** return no matches.

### REQ-CLEANUP-EVAL-005 — Package no longer enumerated (Event-detected)

**When** `go list ./internal/evaluator/` is invoked after removal, the Go toolchain
**shall** report the package as absent (non-zero exit / no listed package).

### REQ-CLEANUP-EVAL-006 — Codemaps reference synchronization (Ubiquitous)

The cleanup change **shall** update the 2 codemaps references that name
`internal/evaluator` so the documented "test-only packages" note reflects only
`internal/skills` after removal (or notes that `internal/evaluator` was removed):

- `.moai/project/codemaps/modules.md` (the `internal/evaluator` line in the modules list)
- `.moai/project/codemaps/overview.md` (the "test-only 패키지" count and parenthetical)

### REQ-CLEANUP-EVAL-007 — Codemaps count consistency (State-driven)

**While** `.moai/project/codemaps/overview.md` is updated, the package-count
arithmetic in the overview **shall** remain internally consistent (the test-only
package count decremented from 2 to 1, and the total `internal` directory count
decremented accordingly).

## §D. Constraints

- Documentation-only and dead-code-only: no production Go source is added or modified.
- The removal MUST NOT alter the behavior, public surface, or test coverage of any
  other package.
- The codemaps edits are consistency fixups, not regeneration — only the 2 named
  references are touched.

## §E. Exclusions

The following are explicitly out of scope. Each excluded item is something the eventual
implementation MUST NOT touch.

### Out of Scope — SPEC history artifacts

- `.moai/specs/SPEC-EVAL-001/` — `status: implemented`, `phase: "v2.x - Legacy"` SPEC
  history; immutable. MUST NOT be modified.
- `.moai/specs/SPEC-EVALLIB-001/` — related legacy SPEC history; immutable. MUST NOT
  be modified.

### Out of Scope — the active sync-auditor implementation

- `.claude/agents/moai/sync-auditor.md` — the actual active delivery of the evaluator
  feature, retained in the 8-agent catalog. MUST NOT be touched.

### Out of Scope — active evaluator scoring profiles

- `.moai/config/evaluator-profiles/` (`default.md`, `strict.md`, `lenient.md`,
  `frontend.md`) — ACTIVE sync-auditor scoring profiles, unrelated to the Go package.
  MUST NOT be touched.

### Out of Scope — the separate internal/skills test-fixture package

- `internal/skills` — a separate, ACTIVE test-fixture package (skill LOC-ceiling /
  template-mirror-parity audit). NOT in scope for this SPEC; MUST NOT be removed or
  modified.

### Out of Scope — any unrelated refactor

- No refactor of any other package; no behavior change anywhere else in the module.
- No regeneration of the full codemaps (only the 2 named references are edited).
