# Implementation Plan — SPEC-TODO-LANDING-EVIDENCE-001

Ordered by **decision reversibility**: the decisions most likely to change on review come first —
the stored shape, the record type, the new verb's surface — and the mechanical work (parity,
downgrade proof, doctrine mirroring) comes last. Review attention belongs at the top.

---

## §A Context

Half A (`SPEC-TODO-LANDING-STATE-001`, card t331) is landed and `completed`; its §B.2 handed this
axis over and named the dependency direction. Preconditions measured in this worktree at HEAD
`e50964ad3` — see `research.md` §R.1.

Route: Tier L ⇒ **Route B (PR route)** per `spec-workflow.md` § SPEC Phase Discipline. Under the
repo-local all-tier PR policy the card's branch integrates through the lead's window regardless; the
lane does not push `develop` and does not open the PR itself.

Development mode: this is new behaviour on an existing store, so **TDD** for the new surfaces
(record type, verb, render) and **characterization-first** for the migration path, where existing
rows must be shown unchanged before the `ALTER` is introduced.

---

## §B Known issues carried in

1. **The schema-freeze guard is column-blind** (measured, `spec.md` §A.5). This SPEC's own column
   would land silently. M1 closes it, and the guard extension is sequenced **before** the column is
   added so the guard's new assertion is written against the old column set and then updated
   deliberately — otherwise the extension is written to match whatever was already built.
2. **`ALTER TABLE` is new to this codebase** (`grep -rn "ALTER TABLE" internal/kanban/ internal/cli/`
   → `rc=1`). There is no in-tree idempotence pattern to copy; M1 establishes one.
3. **The seventh column is the second contract change on this surface** and external consumers still
   cannot be enumerated. Inherited from half A, recorded in `spec.md` §G, not closed here. M4's
   criterion (AC-TLE-015) therefore pins all seven fields, not just the two it adds — leaving
   fields 1-5 free would let a reorder ride along with the insertion.
4. **`moai todo pr` shells out to `git` in the working directory**
   (`internal/kanban/prlink_landed.go:150-154` via the `todoRunCommand` seam at
   `internal/cli/todo_pr.go:57-65`), and git may write inside `.git/` during a read. Any
   byte-identity assertion over the project root must exclude `.git/` or it flakes for reasons
   unrelated to the property (AC-TLE-014).
5. **`--json` currently marshals `PRLinkOutcome` alone** (`internal/cli/todo_pr.go:167`), which the
   render does not merge with queue state. M4 must decide whether the evidence rides on the outcome
   struct or on a render-time wrapper; the wrapper is preferred because `PRLinkOutcome` is the
   resolver's type and REQ-TLE-011 keeps the resolver clean of stored data.

---

## §C Pre-flight

- [ ] `git rev-parse --show-toplevel` reports the t359 worktree; `git branch --show-current` reports
      `WT-landing-evidence`.
- [ ] `go test ./internal/kanban/... ./internal/cli/...` is green **before** any edit, so a later red
      is attributable.
- [ ] `SPEC-TODO-LANDING-STATE-001` still reads `status: completed` and
      `git merge-base --is-ancestor c9f712232 develop` still returns rc=0 (the precondition decays;
      re-read it at run-phase entry rather than trusting `research.md`).

---

## §D Constraints

- [HARD] No fourth `items.state` value; the CHECK is byte-identical after this change.
- [HARD] No `schema_version` bump.
- [HARD] `moai todo pr` writes nothing (`SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-2.1).
- [HARD] The resolver names no delivering commit (same SPEC, REQ-1.10) — unamended.
- [HARD] Evidence never transitions a card.
- No new network call or per-card file read on any read path.
- Template-First: every `.claude/` change lands in `internal/template/templates/` too, then
  `make build`.

---

## §E Self-verification

Run in this worktree, scoped to the touched packages (never the full local suite — CI owns the
full-suite verdict):

```
go test ./internal/kanban/... -count=1
go test ./internal/cli/... -count=1 -timeout 600s
go vet ./internal/kanban/... ./internal/cli/...
```

Each milestone records its command and verbatim tail in `progress.md` §E.2. The five planted mutants
(`acceptance.md` §D.2) record before/after file hashes.

---

## §F Milestones

### M1 — The stored shape *(most reversible decision: review this first)*

The one decision a reviewer is most likely to want changed: **one nullable TEXT column holding a
record**, versus four scalar columns, versus a fifth table. `spec.md` §B.1 argues it; if the review
overturns it, everything below changes shape and nothing below has been built yet.

1. Extend `TestTodoHistoryAddsNoSchemaChange` with exact ordered
   `(name, type, notnull, dflt_value)` tuple assertions for `items` **and** `archived_items`,
   asserted per table — written against the **current** tuples, so it passes before the column
   exists and fails the moment one is added, retyped, or made non-nullable without updating it
   (AC-TLE-019a/b/c). Both tables move in this step: a guard extended for `items` alone leaves half
   of REQ-TLE-019 unbuilt, which AC-TLE-019b exists to catch.
2. Add `landing TEXT` to `items` and `archived_items` via an idempotent `ALTER`, decided from
   `pragma_table_info`, executed at engine open immediately after the DDL and **before** the
   `schema_version` switch (AC-TLE-001/002/003).
3. Update the guard's expected column tuples in the same commit as the `ALTER`, so the two move
   together and the update is a visible act.
4. Assert the CHECK survives the ALTER and existing rows are untouched (AC-TLE-003). **AC-TLE-004
   is NOT claimed here**: its Given-When requires the `moai todo landed` verb, which does not exist
   until M3, so the criterion belongs wholly to M3. (`spec.md` §E maps REQ-TLE-004 → M3 for the same
   reason.)

Deliverables: `internal/kanban/backlog_sqlite.go`, `internal/kanban/backlog_schema_freeze_test.go`,
new `internal/kanban/backlog_landing_test.go`.

### M2 — The record type and the attribution boundary *(second-most reversible)*

The field set (`spec.md` §B.4) is the second decision a review may move. Build it before anything
depends on its encoding.

1. Define the evidence record type with the six fields, its JSON encoding, and its decode
   (AC-TLE-005).
2. Absence is NULL, never a zero-valued record (AC-TLE-006).
3. Distinct keys for the operator-asserted SHA and the observed ref position (AC-TLE-013).
4. Lock REQ-1.10 in with a test over the resolver's rendered outcome against a three-match fixture
   ref, and observe the planted-mutant RED (AC-TLE-011).

Deliverables: `internal/kanban/landing_evidence.go` + tests; `internal/kanban/prlink*_test.go`
addition. **No change to `prlink.go` / `prlink_landed.go` production code** — M2 constrains them, it
does not edit them.

### M3 — The recording verb *(user-facing surface)*

1. `moai todo landed <id> [--sha <sha>] [--ref <ref>] | --clear`, single locked write, no prompt
   (AC-TLE-007).
2. Whole-queue immutability assertion + planted mutant (AC-TLE-008); state untouched (AC-TLE-004).
3. Replace-and-clear semantics (AC-TLE-009).
4. SPEC-status read at record time, unknown when unreadable, never defaulted (AC-TLE-010).
5. No SHA is ever derived from the grep predicate (AC-TLE-012, planted mutant).
6. **`--sha` referential-integrity validation** (AC-TLE-020): resolve with
   `git rev-parse --verify <sha>^{commit}` (existence + full SHA) then
   `git merge-base --is-ancestor <resolved> <ref>` (reachability); store the resolved full SHA;
   exit 1 naming which check failed otherwise, writing nothing. **An unrunnable check — no git, or a
   ref that resolves to nothing — is also exit 1**, deliberately opposite to `todo pr`'s fail-open
   degradation on the same condition: a read that cannot answer stays permissive, a write that
   cannot validate refuses (AC-TLE-020 case (d)). The card id is passed to neither command — that is
   what keeps this a referential-integrity check rather than attribution (`spec.md` §B.3.1), and
   AC-TLE-020 asserts it by recording one `--sha` against **two different card ids**, not by
   renaming one card: the predicate keys on the id
   (`internal/kanban/prlink_landed.go:96-108`), which no `todo` verb changes.

Deliverables: `internal/cli/todo_landed.go` + tests; registration on the `todo` command.

### M4 — The read surfaces *(user-facing surface)*

1. Seventh column, card text stays last; `--json` gains the record under its own key (AC-TLE-015).
   Prefer a render-time wrapper over widening `PRLinkOutcome` (§B.4 above).
2. Assertion-versus-observation marker survives SHA substitution (AC-TLE-016).
3. Project-root-wide byte-identity across `todo pr`, with the mutant planted **outside** the queue
   directory (AC-TLE-014).

Deliverables: `internal/cli/todo_pr.go` + tests.

### M5 — Compatibility and doctrine *(mechanical; least likely to change)*

1. Carry `landing` through the JSON⇄SQLite round trip and into the parity comparison; observe the
   parity check **failing** under the drop mutant (AC-TLE-017).
   Sites: `internal/kanban/backlog_migrate.go:60` (SELECT), `:276` (INSERT), `:585-630` (parity),
   plus the archived read at `:121-123` and the archived write at `:196-201`.
2. Downgrade proof: pre-change statements verbatim against a post-change database; `schema_version`
   still `"1"` (AC-TLE-018).
3. `.claude/skills/moai/workflows/todo.md` — the verb row in the table, the seven-column contract in
   the `todo pr` row (`:57`), and one sentence placing the verb under the operator-act rule
   (`:59-63`). Mirror to `internal/template/templates/.claude/skills/moai/workflows/todo.md`, then
   `make build`.
4. Mirror-parity + stated-column-count criterion (AC-TLE-021). The parity half guards the drift this
   repository has already paid for once in the `.sh` / `.sh.tmpl` hook-wrapper pair; the count half
   ties the prose to what M4 actually emits.

---

## §G Anti-patterns to refuse during implementation

- **Auto-closing a card on a recorded landing.** The most plausible slip on this axis; half A named
  it and AC-TLE-008 catches it.
- **Filling the SHA from the grep's first match "because it's usually right".** REQ-1.10's grounds
  say it is not; AC-TLE-012 catches it.
- **Recording on the read path** because "the observation is already in hand". REQ-2.1 forbids it.
- **A queue-directory-scoped byte-identity assertion.** Half A measured that form passing against a
  cache planted outside it.
- **Bumping `schema_version` "to be safe".** It is the one change that breaks every older binary.
- **Writing the doctrine sentence and treating it as the criterion.** No AC here is satisfied by a
  doc grep.

---

## §H Cross-references

`spec.md` §B (decisions), §D (exclusions), §G (gaps); `acceptance.md` §D.2 (mutant register);
`design.md` (the record encoding and the ALTER placement); `research.md` (the measurements every
claim above rests on).
