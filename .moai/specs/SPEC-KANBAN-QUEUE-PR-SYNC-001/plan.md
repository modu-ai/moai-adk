# Implementation Plan — SPEC-KANBAN-QUEUE-PR-SYNC-001

Milestones are ordered by decision-reversibility: the decisions most likely to
change come first, mechanical work last.

## A. Context

Derived from `spec.md` §A. The measurement record `.moai/reports/t210/measurement.md`
is the evidence base; do not re-derive its numbers during run-phase.

## B. Known issues in the current tree

- No card↔PR mapping exists (**M4**). Every piece of REQ-1 is new code.
- `gh` is invoked from four places today; none of them share a PR-listing
  helper that returns title + body together. The run phase must check whether
  `internal/github/gh.go` already exposes a suitable list call before adding one.
- **M2**'s carrier statistics were computed against a live PR set that will have
  changed by run-phase. The fixture set in the tests must be pinned from the
  measurement record's table, not re-fetched.

## C. Pre-flight

- Confirm `internal/github/gh.go`'s existing surface: does it expose a PR list
  with `title` and `body` fields, or only state?
- Confirm the `moai todo` command registration point in `internal/cli/todo.go`
  and how subverbs are added (`todo_why.go` is the smallest existing example).
- Read `internal/kanban/backlog_store.go` to confirm a read path exists that
  does not take the write lock.

## D. Constraints

- [HARD] Zero writes to `backlog.json`. This is the load-bearing constraint;
  see AC-004.
- Fail-open on every `gh` failure path. No new error exit.
- Template-First: the doctrine edit lands in `internal/template/templates/…`
  in the same commit, then `make build`.
- No new dependency.

## E. Self-verification

Every milestone's exit is a command whose output is cited, per
`.claude/rules/moai/core/verification-claim-integrity.md`.

## F. Milestones

### M1 — The doctrine change (REQ-3)

Highest reversibility risk: it is a [HARD] behavioural rule on the lead and a
naming obligation on every card-delivering PR. It goes first so the decision is
reviewed before any code is written to serve it. Note when writing the clause
that **M6** makes it codification rather than imposition — 8 of 15 merged PR
titles already carry the token, and most of the remainder deliver no card.

- Add the pre-dispatch PR cross-check to `kanban-dispatch.md`, sited next to
  § Entry into the board is an operator act.
- Add the [HARD] PR-title card-id clause, with the explicit non-contradiction
  note against the branch-naming rule (REQ-3.3).
- Mirror both into `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`,
  neutrality-checked.
- Exit: `go test ./internal/template/...` green; mirror parity check green.

### M2 — The resolver (REQ-1)

Second-most reversible: the confidence taxonomy (`exact`/`inferred`/`ambiguous`)
is a user-visible contract, and the no-commit-messages ruling is a decision a
reviewer may want to revisit.

- New package function taking `(cardIDs []string, prs []PRRecord) []Link`.
- Pure — no network, no filesystem.
- Fixture PR set transcribed from the measurement record's M2 table, so the
  known-hard cases (#1600, #1614, #1611, #1612) are all covered.
- Exit: table-driven tests asserting the confidence label for each fixture row.

### M3 — The read surface (REQ-2)

- New `moai todo pr [<id>]` verb with `--json`. The dedicated-verb choice is
  settled by **M5** (0.878s per `gh pr list`), so do not reopen it as a
  `todo list` column during implementation.
- Exactly one `gh pr list --state open --json number,title,body,state` call per
  invocation; feed the resolver; render. No per-card query.
- Fail-open wrapper around the `gh` call.
- Exit: byte-identity assertion on `backlog.json` (AC-004); fail-open test with
  `gh` forced absent via `PATH`.

### M4 — Mechanical wiring

- Help text, `--json` schema doc in `.claude/skills/moai/workflows/todo.md`
  command table, template mirror of that skill file.
- Exit: `go vet ./...`, `golangci-lint run`, affected-package tests.

## G. Anti-patterns to avoid

- Adding the PR column to `moai todo list` unconditionally — rejected in
  REQ-2.5; it makes every queue read a network read.
- "Helpfully" resolving an `ambiguous` link to the highest-numbered PR.
- Writing the resolved link back into `findings[]` "because the array already
  exists". The array exists for *pair relations between cards*, and writing to
  it is still a write.
- Scanning commit messages "as a tiebreaker". M2 rules the carrier out entirely.

## H. Cross-references

- `spec.md` §B (the read-only ruling), §C (the carrier problem)
- `acceptance.md` (AC matrix)
- `.moai/reports/t210/measurement.md`
