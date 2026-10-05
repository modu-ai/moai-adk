# SPEC-WEB-ANCHOR-SCOPE-001 — acceptance.md

## D. AC Matrix

| AC | Maps to | Subject | Severity |
|---|---|---|---|
| AC-WAS-001 | REQ-WAS-001 | classification table completeness | must-pass |
| AC-WAS-002 | REQ-WAS-001 | per-row (a)(b)(c) evidence | must-pass |
| AC-WAS-003 | REQ-WAS-002 | no bulk substitution | must-pass |
| AC-WAS-004 | REQ-WAS-003 | repair shape + production untouched | must-pass |
| AC-WAS-005 | REQ-WAS-007 | mutant falsification of the discriminator | must-pass |
| AC-WAS-006 | REQ-WAS-004/006 | mirror-present green baseline / valid zero-repair close | must-pass |
| AC-WAS-007 | REQ-WAS-005 | discriminator durable in research.md | must-pass |

## D.1 AC-WAS-001 — Table completeness

**Given** the raw-superset measurement on the run-phase tree
**When** the classification table row count is compared to the measured `strings.Index(`
site count in `internal/web/*_test.go`
**Then** the counts are equal, and the measurement commands + verbatim outputs are
recorded in progress.md §E.2 with the measured HEAD SHA.

## D.2 AC-WAS-002 — Per-row evidence

**Given** every row of research.md §5
**When** inspected
**Then** each row carries (a) intended scope, (b) actual anchor resolution under
mirror-present rendering (exposure + duplication + order), and (c) a three-valued
verdict (TRUE / false / NOT-EXPOSED), with the REPAIRED-BY-T509 row explicitly marked.

## D.3 AC-WAS-003 — No bulk substitution

**Given** the run phase's diff
**When** each changed test line is traced
**Then** every change maps to a (c)=TRUE row of the table (zero rows ⇒ zero diff on
`internal/web/*_test.go`), and no mechanical mass rewrite of the superset exists.

## D.4 AC-WAS-004 — Repair shape

**Given** any landed repair
**When** inspected
**Then** the receiver is narrowed via `panelHTML(t, renderConsolePage(t), "<panel>")`
or an equivalent already-scoped slice, the asserted property is unchanged (narrowing is
not weakening), and `git diff --stat` shows zero changed lines under non-test
`internal/web/*.go` files.

## D.5 AC-WAS-005 — Mutant falsification (the discriminator has teeth)

**Given** the mirror-present tree
**When** mcp_console_test.go:115 is temporarily reverted to the pre-t509 body-wide
anchor and `go test ./internal/web/ -run TestMCPConsoleWriteCapableTextDistinction` runs
**Then** the test FAILS with the mirror-misanchor failure mode, the RED output is
recorded, the mutant is restored, and the restored test passes. A discriminator adopted
without this RED cannot claim it would catch the defect family (absence-guard rule:
RED-now evidence or no adoption).

## D.6 AC-WAS-006 — Mirror-present regression baseline

**Given** the run-phase tree with `git merge-base --is-ancestor 1aaf4951f HEAD` exit 0
**When** `go build ./internal/web/` and `go test ./internal/web/` run
**Then** build exits 0 and tests report `ok`, with verbatim output recorded. On the
zero-repair path this AC plus AC-WAS-005 constitute the card's deliverable (REQ-WAS-006).

## D.7 AC-WAS-007 — Discriminator durability

**Given** the SPEC directory
**When** research.md §4 is read
**Then** the discriminator is stated as a three-condition rule (EXPOSURE ∧ DUPLICATION ∧
ORDER) keyed to the derived mirror predicate families, applicable to future tests
without this card's context.

## Edge Cases

- A NEW mirrored field lands between plan and run → M1 re-classification catches the
  widened inventory (B.1); delta rows go through the §4 discriminator.
- A site whose needle is duplicated by the mirror but whose owning panel precedes the
  mirror (audit-tab needles) → (c)=false by the ORDER condition; recorded, not repaired.
- Receiver is an already-scoped slice that happens to start before the codex panel and
  extend past it (none in the current corpus) → EXPOSED for the slice's span; classify
  on the slice, not the original body.

## Quality Gate Criteria

- Affected package build + tests green (D.6); lint clean on touched files.
- No time estimates anywhere; priority labels only.
- Every count in progress.md §E.2 carries tree SHA + command + verbatim output.

## Definition of Done

All ACs pass; progress.md §E.1/§E.2 populated by run phase; zero unexplained diff
surface; the classification table is current with the closing tree.
