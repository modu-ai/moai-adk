# SPEC-WEB-ANCHOR-SCOPE-001 — plan.md

## A. Context

Card t527 (t509-derived sweep, lane-12, worktree `WT-anchor-scope-sweep`). The codex
mirror (SPEC-WEB-CODEX-PANEL-001) re-renders mirrored-field chips at tab position 8,
ahead of the MCP tab at 11. Page-wide first-occurrence anchors can therefore resolve to
the mirror's copy — the t509 incident. This SPEC's first deliverable is the
CLASSIFICATION TABLE (research.md §5, 60 rows, complete at plan phase), not a repair.
Defect family: "the judged target exists — but it is a different one" (siblings t506,
t509; this is the third variant).

## B. Known Issues

- B.1 The mirror's field inventory is DERIVED by predicate (`codexmirror.go`): a future
  codex field is mirrored with no code change, silently widening the duplication
  inventory. The discriminator (research.md §4) keys on the predicate families for this
  reason; it must be re-run, not remembered, whenever a mirror field is added.
- B.2 Numerical drift hazard: card figures (34/15 @ 0b1e27877, 60/19 lane-1) are tree-
  specific. Any re-measurement cites its own tree — never a carried-over count.

## C. Pre-flight

1. `git rev-parse --short HEAD` → confirm worktree tree; record in evidence.
2. `git merge-base --is-ancestor 1aaf4951f HEAD && echo MIRROR-PRESENT` → must print
   MIRROR-PRESENT (REQ-WAS-004).
3. Re-run the superset measurement (`grep -c 'strings\.Index(' internal/web/*_test.go`);
   if the count differs from 60, re-classify the delta rows through research.md §4
   before any repair.

## D. Constraints

- Repairs only on (c)=TRUE rows (REQ-WAS-002); zero-repair is a valid close (REQ-WAS-006).
- Repair shape: scope narrowing only, `panelHTML` pattern or equivalent already-scoped
  slice (REQ-WAS-003). Production/render/save code untouched.
- All regression measurement under mirror presence (REQ-WAS-004).
- Verification scope: affected package `./internal/web/` only — no full-suite local run.

## E. Self-Verification

Each milestone closes with verbatim command output recorded in progress.md §E.2:
build (`go build ./internal/web/`), affected tests (`go test ./internal/web/`),
and the mutant check (AC-WAS-005) with its observed RED output.

## F. Milestones

Ordered by decision-reversibility: the classification (highest change-likelihood — it
decides everything downstream) leads; mechanical verification trails.

### M1 — Classification table currency check (Priority High)

Re-run the pre-flight measurement on the run-phase tree; diff against research.md §5's
60 rows; route any new/changed site through the §4 discriminator and append rows.
**Exit**: row count == measured site count; every row carries (a)(b)(c).

### M2 — Discriminator falsification (mutant check) (Priority High)

Temporarily revert mcp_console_test.go:115 to the t509 shape
(`strings.Index(body, chip)` on the full renderConsolePage body), run
`go test ./internal/web/ -run TestMCPConsoleWriteCapableTextDistinction`, observe RED,
restore, observe GREEN. **Exit**: RED output captured (AC-WAS-005). Do NOT land the
mutant — restore is mandatory before milestone exit.

### M3 — Repairs per table (Priority High; expected zero rows)

For each (c)=TRUE row from M1: apply the t509 scope-narrowing shape; one commit per
logical group; diff traceable row-by-row to the table. If the set is empty, close with
no code change (REQ-WAS-006). **Exit**: zero unclassified (c)=TRUE rows; any repairs'
diffs cite their rows.

### M4 — Verification (Priority Medium)

`go build ./internal/web/ && go test ./internal/web/` green on the mirror-present tree
(AC-WAS-004/006); golangci-lint on touched packages if any code changed.

## G. Anti-Patterns

- Bulk sed-style replacement of the superset (forbidden by REQ-WAS-002).
- "Fixing" (c)=false rows for consistency — a divergence-free anchor is not debt.
- Measuring regression on a mirror-absent tree.
- Weakening an assertion while narrowing its scope (narrowing is not weakening — t509
  precedent: the window must still assert the same property, just on the right region).

## H. Cross-References

- research.md §4 — the discriminator (durable).
- research.md §5 — the classification table (gating artifact).
- SPEC-WEB-CODEX-PANEL-001 — the mirror; its close commit 1aaf4951f is the
  mirror-presence sentinel.
- `.moai/reports/t527/` — card evidence path (lane-owned at close).

### Gate-round resolution (2026-09-07)

The zero-repair clarification above was taken to the operator gate round and RESOLVED:

- **Decision 1 — zero-repair close via REQ-WAS-006.** The table + mutant evidence
  (AC-WAS-005) are the card's deliverable; defensive hardening of the remaining exposed
  full-body anchors (e.g. converting them to `panelHTML` without a (c)=TRUE row) is NOT
  taken, and stays closed by REQ-WAS-002 as written — no change to the requirement.
- **Decision 2 — run-phase entry APPROVED** (Implementation Kickoff passed).
- **Decision 3 — autonomous continuous progression**: run → sync → lead-window request
  in this session.

Milestone consequence: the run phase executes M1 (re-measure + table currency) and M2
(mutation evidence) only; M3 executes over an expected-empty (c)=TRUE set and closes with
no code change per REQ-WAS-006; M4 unchanged.
