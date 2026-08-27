# Implementation Plan — SPEC-TODO-DESTRUCTIVE-GUARD-001

Tier M. Milestones are ordered by decision-reversibility: the storage shape and the guard semantics lead, the mechanical wiring and doctrine follow.

---

## §A Tier classification and its justification

**Tier M.** Measured against the taxonomy (5-15 files, 300-1000 LOC):

| File | Nature |
|---|---|
| `internal/kanban/backlog_store.go` | archive fields on `BacklogRecord`, archive/restore methods |
| `internal/kanban/backlog_sqlite.go` | two `IF NOT EXISTS` tables in `backlogDDL`, load/save |
| `internal/kanban/backlog_migrate.go` | record↔database mapping for the new tables |
| `internal/cli/todo.go` | `done` rewired to archive; `--expect`, `--require-landed` |
| `internal/cli/todo_undone.go` | new verb (new file) |
| `internal/cli/todo_export.go` | downgrade disclosure on stderr (REQ-TDG-015) |
| `internal/kanban/backlog_archive_test.go` | storage round-trip (new file) |
| `internal/cli/todo_undone_test.go` | CLI round-trip and refusals (new file) |
| `.claude/skills/moai/workflows/todo.md` | doctrine |
| `internal/template/templates/.claude/skills/moai/workflows/todo.md` | template mirror |

Ten files, none of them constitutional. Estimated 400-700 LOC including tests. Tier M PASS threshold: 0.80. Requirement count 16 and acceptance-criterion count 16 both sit **at** the Tier M ceiling of 16 — any further growth is a signal to split, not to relax the budget.

Not Tier S: ten files exceeds the `< 5 files` guidance, and the change reaches the storage layer, the CLI, and the downgrade route. Not Tier L: no constitutional surface, no cross-subsystem redesign, and the storage change is additive by ruling.

---

## §B Decision 1 — why an archive rather than a fourth state

Stated in full at `spec.md` §B.1. The short form, from the measurements rather than from symmetry with `drop`:

1. `backlog_sqlite.go:100` bakes `CHECK (state IN ('queued','picked','dropped'))` into a `CREATE TABLE IF NOT EXISTS`. Existing operator databases keep the old CHECK, and SQLite cannot `ALTER` one. A fourth state costs a **table-rebuild migration on every existing queue in the field**.
2. `backlog_sqlite.go:232-235` runs the entire `IF NOT EXISTS` DDL on **every open**. A new table therefore lands on every existing database for free, with no migration code.

That asymmetry — additive tables free, changed CHECK expensive — is what decided it. The `findings` precedent (`backlog_store.go:145-157`: a second additive top-level field, no version bump, per-item contract untouched) confirms the record-level shape without being the reason.

**If a reviewer prefers the fourth-state design anyway**, the cost is concrete and must be stated rather than discovered: a rebuild migration touching every operator queue, plus an audit of every state-reading path — `backlog_analysis.go:139` skips only `dropped`; `backlog_migrate.go:252-258` counts only `picked` and `queued`; every listing filter besides. The archive has none of that leak surface by construction.

**The `schema_version` must not be bumped.** `backlogSchemaVersion = "1"` and `ensureSchema` aborts with `ErrBacklogCorrupt` semantics on any mismatch, a *newer* stamp included. Bumping to `"2"` would make an older binary refuse to open the queue outright. Holding at `"1"` keeps both directions working. This is REQ-TDG-005 and it is easy to violate by reflex.

---

## §C Decision 2 — the scope boundary against t331

Stated in full at `spec.md` §B.2, with the two failure modes at §A.4. The load-bearing finding is that the existing landed primitive fails in **both** of its reachable configurations, and correcting the ref moves the failure rather than removing it:

| Mode | `LandedRef` | Answer for t306 | Failure |
|---|---|---|---|
| As shipped | `origin/main` (`prlink_landed.go:28`) | **false** — 0 matching commits | default-on refusal blocks every develop-integrated card |
| After the obvious ref correction | integration branch | **true** — 13 matching commits, earliest the run commit `3cb258d62` | default-on refusal passes the premature `done` silently |

Measured at `812ee01fc`:

```
$ git log origin/main    --perl-regexp --grep='\bt306\b' --oneline | wc -l
       0
$ git log origin/develop --perl-regexp --grep='\bt306\b' --oneline | wc -l
      13
```

`origin/main` is live (tip `48239c7dc`, and it names older cards such as t230), so the zero is a genuine "not on main", not a missing ref.

Consequence for the plan: **do not ship a default-on landing refusal**, and do not treat the ref correction as a prerequisite that would unlock one. t330 ships the seam plus an opt-in `--require-landed` whose help text names the ref its answer is about. t331 supplies the persisted landing-state field that lets the predicate answer the right question; flipping to default-on is t331's call.

**Pre-flight for M4.** Verify `LandedRef`'s value at implementation time before wiring the flag, and make the flag's help text and the doctrine state plainly which ref the answer comes from. Do **not** correct the constant here — it is shared with `moai todo pr`, so changing it alters that verb's answers (recorded as out of scope in `spec.md` §D). Raise it to the operator as a candidate follow-up card.

**A caution for whoever writes this up.** The Mode 2 form — "the run commit satisfies it, so it was already true" — is false as shipped and must never be stated without its ref-correction condition. That condition was dropped once already between the measurement and the first draft of this SPEC.

---

## §D Constraints

- **Never prompts.** `internal/cli/todo.go:20` and `todo_pr.go:15` carry `SUBAGENT BOUNDARY` as a package discipline. Every path added here — success, refusal, and error — respects it. An agent-invoked lane must never block.
- **Refused mutation writes nothing.** `Mutate` already leaves the file byte-identical when the callback returns an error (`internal/cli/todo.go:351-353`). Guards must refuse *inside* the callback so they inherit this, not before it.
- **One live engine, two serializations.** Express the archive at the `BacklogRecord` level so it rides the existing `Mutate` seam. There is no live JSON engine to support: `openEngine` (`backlog_store.go:437-455`) migrates a JSON-only queue under the lock and then falls through to `openBacklogEngine(backlogSQLitePath(...))` on every path, and `todo_export.go:3-11` records that the swap is one-way with no engine-selecting knob. The record-level shape matters because the JSON **file format** is the downgrade artifact, not because JSON ever backs a running queue.
- **Per-item contract frozen.** No new field on `BacklogItem` — REQ-TODO-013, restated at `backlog_store.go:44-47`.
- **`moai todo pr` is not modified.** Read-only by ruling.

---

## §E Milestones

Ordered most-reversible-decision first.

### M1 — archive storage shape (highest change likelihood)

Add `Archived []BacklogItem` and `ArchivedFindings []BacklogFinding` as top-level `BacklogRecord` fields, following the `findings` precedent exactly — always rendered as arrays, never null, never omitted, absent in older files. Add the two `IF NOT EXISTS` tables to `backlogDDL` and their load/save mapping. **Do not touch `backlogSchemaVersion`.**

Exit: a database created by the previous binary opens under the new one and gains the tables; a database containing archived rows opens under the previous binary without error.

### M2 — `done` archives; `undone` restores

Rewire `newTodoDoneCmd` to move the row and its findings into the archive instead of discarding them, and add `newTodoUndoneCmd` to move them back. The restore must return **both** — the card and every findings row that named it (§A.2). Refuse a restore whose id has been reissued (REQ-TDG-013).

Exit: `done` then `undone` leaves the queue record byte-identical, asserted the way the `drop`/`undrop` round-trip is (`internal/cli/todo_drop_test.go:69`).

### M3 — `--expect` guard on `done`

Follow the existing convention verbatim (`todo_drop.go:104-106`): compare against the card's current text prefix, refuse inside the `Mutate` callback, report the observed prefix in the error.

### M4 — `--require-landed` seam (read §C pre-flight first)

Opt-in only. Refuse on positive evidence of not-landed; **proceed on inconclusive** — no runner, git unavailable, or query error — matching the repository's fail-open guard doctrine. Absent the flag, run no landing query at all (REQ-TDG-010).

### M5 — the visibility split: invisible to readers, present in the export

Two directions, and they are **opposite on purpose**. This milestone budgets work for both; neither is "true by construction".

- **Invisible to live-queue readers** — `list`, `next`, `why`, `analyze`/`ClassifyCardText`, and the state counts (REQ-TDG-007). With the archive in separate tables and separate record fields this should hold structurally; M5 asserts it.
- **Present in `export-json`, with a disclosure** (REQ-TDG-015). Inclusion is what `json.MarshalIndent(rec, …)` (`internal/cli/todo_export.go:69`) already does with any top-level field, so the *inclusion* costs nothing — but it is a deliberate ruling (`spec.md` §C.5 Decision 3), not an accident to be tidied away. The work here is the **stderr disclosure**: when the exported record carries archived rows, say that a release predating the archive discards them on its first write.

The disclosure is the whole answer to the downgrade-loss problem. Nothing in this SPEC runs inside the older binary, so the loss cannot be prevented — only made loud. The stderr line satisfies REQ-TDG-015 and is the cheapest carrier; the exported artifact is equally ours, and a top-level warning string beside the archive would ride the same marshal for free and outlive scrollback for anyone who opens the file before downgrading. Either carrier, or both, is acceptable here.

Note the stream: the disclosure goes to **stderr**. `internal/cli/todo.go:20-22` contracts one structured stdout line, and `internal/cli/todo_export.go:81-82` is that line — the surface agents parse. AC-TDG-015 captures the streams separately for exactly this reason.

Exit: AC-TDG-007 and AC-TDG-015 both pass in the same isolated repository, demonstrating the split is deliberate.

### M6 — doctrine and template mirror (mechanical)

See §F.

---

## §F Run-phase obligations

### F.1 Template-First [HARD]

The todo doctrine exists at **both** paths, byte-identical (13709 bytes, `cmp` clean at `812ee01fc`):

- `.claude/skills/moai/workflows/todo.md`
- `internal/template/templates/.claude/skills/moai/workflows/todo.md`

Any doctrine edit describing `undone` or the new flags must land in **both**, followed by `make build` to regenerate the embedded template. Verify parity with `cmp` before committing; a diverged pair is a silent template regression.

### F.2 Verification isolation [HARD]

The live queue in this repository is in active use by six concurrent lanes plus the lead. **No acceptance criterion may be verified against it.** Every runnable check runs in an isolated repository — `mktemp -d`, `git init`, a `.moai/` skeleton. This is a standing precondition, restated at the head of `acceptance.md`.

### F.3 Exit-code reading [HARD]

Any check asserting a non-zero exit reads it as:

```bash
out=$(cmd 2>&1); rc=$?
```

Never `$?` after a pipe — that reports the pipe's status and has already produced one false `rc=0` reading in this project.

### F.4 REQ/AC budget is exhausted [HARD]

REQ and AC counts both sit at **16/16**, exactly the Tier M ceiling. The run phase inherits no headroom.

If a fix genuinely requires a 17th requirement or a 17th acceptance criterion, that is the **split signal firing** — split the SPEC rather than relaxing the ceiling. An over-budget SPEC is the over-formalization failure the tier taxonomy exists to prevent, and it lands hardest on the auditor, which must hold every requirement and criterion in view at once. Sharpening an existing criterion costs nothing against this budget; adding one costs the whole margin.

### F.5 Verification scope

Run the affected packages (`go test ./internal/kanban/... ./internal/cli/...`), then push and read CI for the full-suite verdict. Do **not** run `go test ./...` locally — concurrent lanes doing so drove machine load to 413 (CLAUDE.local.md §4). Note `internal/cli` needs a timeout floor of 600s.

---

## §G Risks

| Risk | Mitigation |
|---|---|
| Reflexive `schema_version` bump breaks downgrade | REQ-TDG-005; M1 exit criterion tests the old binary explicitly |
| Restore recovers the card but not its findings | §A.2 makes findings restoration a first-class requirement; AC-TDG-002 asserts byte-identity, which fails if findings are lost |
| `--require-landed` wired to a stale `LandedRef` | §C pre-flight: verify the constant before wiring; document which ref the answer is about |
| Doctrine lands in one of the two todo.md paths | §F.1: `cmp` parity check before commit |
| Archived rows leak into a live-queue reader | M5 sweep; separate tables make it structural rather than filter-dependent |
| The M5 export inclusion is "tidied away" as a leak by a later reader | `spec.md` §C.5 Decision 3 states it as a ruling with its reason; AC-TDG-015 asserts it positively, so removing it fails a test rather than passing silently |
| Downgrade silently discards the archive | Cannot be prevented — the older binary is not ours. REQ-TDG-015 makes it loud at export time, the only controllable point |
| Unbounded archive growth | Accepted and declared out of scope; retention needs operator input |

---

## §H Cross-references

- `spec.md` §A.4 — the measurement refuting the obvious guard
- `spec.md` §B.1 — the three storage measurements behind Decision 1
- `internal/kanban/backlog_store.go:44-47, 145-157, 201` — additive precedent, record shape, findings removal
- `internal/kanban/backlog_sqlite.go:50, 100, 232-235, 251-253` — schema version, state CHECK, DDL on every open, the version-mismatch abort
- `internal/kanban/backlog_store.go:437-455` — `openEngine`; every path falls through to the SQLite engine
- `internal/kanban/prlink_landed.go:28, 55-60` — `LandedRef`, the no-network querier
- `internal/cli/todo.go:20, 137-141, 332, 347, 441` — subagent boundary, the 15-verb registration, the `done` command and its findings removal, `next --expect`
- `internal/cli/todo_drop.go:104-106, 133, 169-171, 190` — the `--expect` convention on `drop`/`undrop`
- `internal/cli/todo_edit_move.go:91` — `--expect` on `edit` (note: `move` does **not** carry it)
- `internal/cli/todo_export.go:1-11, 69` — the downgrade-route ruling and the whole-record marshal
- `internal/cli/todo_pr.go:1-15` — the read-only ruling and the dedicated-verb cost rationale
