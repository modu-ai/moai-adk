# Implementation Plan — SPEC-KANBAN-QUEUE-PR-SYNC-001

Milestones are ordered by decision-reversibility: the decisions most likely to
change come first, mechanical work last.

## A. Context

Derived from `spec.md` §A. Two evidence bases, and neither is re-derived during
run-phase:

- `.moai/reports/t210/measurement.md` — M1..M6.
- `.moai/reports/t210/verdict.md` — audit iteration 1 (FAIL), D1-D18.

**Sequencing against the sibling SPEC.** `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`
carries the doctrine changes (pre-dispatch cross-check, [HARD] PR-title clause)
and lands **first**. That ordering is deliberate: the title clause raises the
Q1 title carrier from 64% recall toward 100%, which is what turns REQ-1.2's
`exact` path from a minority case into the common one. This SPEC ships the
tooling the doctrine relies on; the two are sequenced, not merged.

## B. Known issues in the current tree

- No card↔PR mapping exists (**M4**). Every piece of REQ-1 is new code.
- `gh` is invoked from four places today; none of them share a PR-listing helper
  that returns title and body together. Check whether `internal/github/gh.go`
  already exposes a suitable list call before adding one.
- **Fixture pinning is mandatory.** **M2**'s carrier statistics and the AC
  fixtures were measured against a live PR set that changes continuously — the
  open set has already grown since M2 was taken. Transcribe the fixtures from
  `acceptance.md`'s pinned block; do **not** re-fetch at test time. A test that
  queries live GitHub is not a test.
- **The `-E` trap is a live hazard, not a hypothetical** (audit D17). `\b` is
  not POSIX ERE and git fails silently. Any landed-check code path must pass
  `--perl-regexp` explicitly, and the AC-011 positive control is what stops a
  regression from passing as a clean result.

## C. Pre-flight

- Confirm `internal/github/gh.go`'s existing surface: does it expose a PR list
  carrying `title` and `body`, or only state?
- Confirm the `moai todo` subverb registration point in `internal/cli/todo.go`.
  `internal/cli/todo_why.go` is the smallest existing example of a read-only
  subverb.
- Confirm `internal/kanban/backlog_store.go` exposes a read path that does not
  take the write lock — REQ-2.2 forbids taking one.
- Confirm both mirror targets exist:
  `internal/template/templates/.claude/skills/moai/workflows/todo.md` and
  `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`
  (the latter belongs to the sibling SPEC).
- Record the AC-013 zero baseline on both `todo.md` copies before M4.

## D. Constraints

- [HARD] Zero writes anywhere under `.moai/state/kanban/`. This is the
  load-bearing constraint; AC-004 enforces it at directory granularity, not
  file granularity.
- Fail-open on every `gh` failure path. No new error exit.
- The landed check is local git and must keep working when `gh` does not.
- Template-First: the `todo.md` mirror lands in the same commit as the local
  edit, then `make build`.
- No new dependency.

## E. Self-verification

Every milestone's exit is a command whose output is cited, per
`.claude/rules/moai/core/verification-claim-integrity.md`. AC-011 additionally
requires its controls to be cited, not merely asserted — a landed-check result
with no positive control is a gap, not a pass.

## F. Milestones

The former M1 (doctrine change) has moved to the sibling SPEC, so this plan
starts at the resolver.

### M1 — The resolver, including the two-question split (REQ-1)

Most reversible remaining decision: the outcome taxonomy (`linked` / `ambiguous`
/ `landed` / `no-link`) and the confidence labels are a user-visible contract,
and the Q1/Q2 carrier split is the design ruling a reviewer is most likely to
want to revisit.

- New package function taking `(cardID, prs []PRRecord, landed LandedQuerier)`
  and returning one outcome record.
- Pure — no network, no filesystem, no repo. `LandedQuerier` is an interface so
  the landed path is fixture-driven in tests.
- Q1 (title → body) and Q2 (landed) are separate code paths with separate tests;
  do not fold them into one scan.
- Fixture set transcribed from `acceptance.md`'s pinned block, covering #1600,
  #1601, #1611, #1612, #1614.
- Exit: AC-001, AC-002, AC-003, AC-006, AC-007, AC-008, AC-012 green.

### M2 — The landed check and its controls (REQ-1.9, REQ-1.10)

Separated from M1 because its failure mode is silent and its guard is the thing
that catches it.

- Implement the `LandedQuerier` against `git log origin/main --perl-regexp
  --grep=…`, returning a boolean and nothing else.
- Write the AC-011 controls **first** — the positive control, the negative
  control, and the `-E` tripwire — then the implementation.
- Exit: AC-011 green, with the positive control's non-empty result cited.

### M3 — The read surface (REQ-2)

- New `moai todo pr [<id>]` verb with `--json`. The dedicated-verb choice is
  settled by **M5** (0.878s per `gh pr list`); do not reopen it as a
  `todo list` column during implementation, and do not add a `--pr` flag —
  §H records why it is out of scope.
- Exactly one `gh pr list --state open --json number,title,body,state` call per
  invocation. No per-card network query.
- Fail-open wrapper around the `gh` call; the landed path stays live when `gh`
  is absent.
- Exit: AC-004 (recursive directory digest), AC-005, AC-009, AC-010 green.

### M4 — Mechanical wiring and the mirror

- Help text; the `moai todo pr` row in the `.claude/skills/moai/workflows/todo.md`
  command table; the template mirror of that file; `make build`.
- Exit: AC-013 green (both mechanical greps, plus the reviewer-judgement half
  recorded as a judgement); `go vet ./...`, `golangci-lint run`, and the
  affected-package tests clean.

## G. Anti-patterns to avoid

- **Re-fetching fixtures from live GitHub at test time.** The measured PR set
  has already changed; a live query makes the suite non-deterministic and
  silently invalidates the AC-002 / AC-003 fixtures the way the v0.1.1 draft was
  invalidated.
- **Using `-E` instead of `--perl-regexp`.** Returns empty, looks clean, reports
  every card as not-landed.
- **Returning the first matching commit as "the delivering commit."** REQ-1.10
  forbids it; §C.2 records the mis-attribution it produces.
- **Extending REQ-1.5's commit-message ban to the landed check.** That
  over-generalization is the defect audit D16 found; the ban binds Q1 only.
- **Adding the PR column to `moai todo list`, or adding a `--pr` flag.** Both
  are out of scope; the first is rejected in REQ-2.5, the second in §H.
- **Writing the resolved outcome into `findings[]` "because the array already
  exists."** The array holds *pair relations between cards*, and writing to it is
  still a write.
- **Asserting only `backlog.json`'s bytes.** AC-004 is a directory digest for a
  reason: a sidecar, a lock, or an mtime touch all pass a single-file check.

## H. Cross-references

- `spec.md` §B (the read-only ruling), §C (the two questions and their carriers)
- `acceptance.md` (the 13 criteria and the pinned fixture block)
- `SPEC-KANBAN-PR-CARD-TRACEABILITY-001` (the sibling doctrine SPEC, lands first)
- `.moai/reports/t210/measurement.md`, `.moai/reports/t210/verdict.md`
