# Implementation Plan — SPEC-QUEUE-UPGRADE-PROOF-001

Card: `t470` · Tier M · Branch `WT-queue-upgrade-proof` · Base `4e4607abe`

## §A Resolved clarifications

Both markers raised at `v0.1.0` are now settled by the dispatcher's ruling on
card `t470`. Each record below keeps the original question, states the answer,
names its source, and states the consequence for this SPEC. Nothing here is
deleted: a reader must be able to see what was asked and how it was settled.

### RESOLVED — G2 definition

**What was asked.** The dispatch named "G2: close alongside G1" and never said
what G2 is. This SPEC refused to guess, wrote no requirement for it, and
recorded that if no definition arrived the card would close with G2 marked
unstarted.

**The answer.** The definition existed on the `t470` card body all along; the
dispatch simply did not carry it. Verbatim:

> **G2** = the chained case is unverified — the existing tests hand-plant their
> layouts and measure each conversion separately, and none measures the path
> where, from a real v3.1.2 layout, the rename and the SQLite conversion fire
> one after the other.

**Source.** The dispatcher's ruling on card `t470`, relayed with the card-body
definition above.

**The ruling.** G2 is **ABSORBED by this SPEC's corrected G1**, not unstarted.
Fixture F1 — the genuine three-field v3.1.2 record planted in
`.moai/state/kanban/` (§B below) — makes a single `moai todo` invocation fire
the directory relocation and then the JSON-to-SQLite conversion in sequence,
which is exactly the chain G2 asked for.

**Consequence for this SPEC.** No new requirement and no new acceptance
criterion is added; G2's content is already carried by criteria this SPEC
already states. The criteria that carry it:

| Chain step G2 names | Criterion that measures it |
|---|---|
| Starting from a real v3.1.2 layout (not a hand-planted develop-schema one) | `AC-QUP-006` (fixture fidelity: `version`, `last_seq`, `items` and nothing else) |
| The rename fires | `AC-QUP-002` (legacy directory relocated to the current name, sentinel arrives) |
| The SQLite conversion fires **after** it, in the same invocation | `AC-QUP-003` (legacy document quarantined as `.migrated`) and `AC-QUP-004` (`backlog.db` present) |
| The chain preserves the queue end to end | `AC-QUP-001a` / `AC-QUP-001b` |
| The chain is entered once, through the CLI, not per-layer through the API | `AC-QUP-005` (§C test-entry design) |

The earlier contingency wording — that G2 would close as unstarted if no
definition arrived — is **withdrawn**: that branch did not happen. G2 closes
with G1.

### RESOLVED — downgrade intent vs quarantine rename (and the earlier mechanism was WRONG)

**What was asked.** `state_dir.go`'s comment on `backlogFileName` says the
queue document keeps the name `backlog.json` so "an older binary reads only
this file". `migrateLegacyBacklog` renames that file to
`backlog.json.migrated`, and the directory itself moves from
`.moai/state/kanban/` to `.moai/state/todo/`. Read together, an older binary
after an upgrade would find neither the directory nor the file. Is the comment
wrong, or is the rename wrong?

**The mechanism as previously written here was incorrect.** The `.migrated`
rename does **not** contradict the downgrade intent, and the rename was never
the tension. Measured at this tree:

1. `internal/cli/todo_export.go:34-41` — `moai todo export-json` exists
   precisely as the downgrade route: "Write the live queue out as a
   legacy-format backlog.json beside the database. Use this before downgrading
   to a release that predates the SQLite queue store."
2. `internal/kanban/backlog_migrate.go:530-532` — a `backlog.json` sitting
   beside a database with no in-flight marker "is not pre-cutover legacy — it
   is an export written for a downgrade — and it is left untouched."
3. `internal/kanban/backlog_migrate.go:604-606` — that export "IS the legacy
   artifact a downgrade-then-upgrade cycle migrates back".

So the file is **meant to be re-created** by `export-json`, and keeping the
name `backlog.json` is exactly the choice that makes that work. The quarantine
rename moves the *consumed* legacy document out of the way; the *downgrade*
artifact is written fresh.

**The real hole is the DIRECTORY, not the filename.** `runTodoExportJSON`
writes to `store.Path()` (`internal/cli/todo_export.go:74`) — the resolved,
i.e. NEW, directory `.moai/state/todo/`. A `v3.1.2` binary reads
`.moai/state/kanban/` (measured:
`git show v3.1.2:internal/kanban/backlog_store.go`, `BacklogPathForRoot`). The
export therefore lands where the old binary will not look; the user must move
the file by hand, and nothing on any surface says so. The comment at
`internal/kanban/state_dir.go:143-146` is right about the filename and silent
about the directory.

**Source.** The dispatcher's ruling on card `t470`, with the four citations
above verified against this tree before being recorded.

**The ruling.** **OUT OF SCOPE for `t470`.** This card proves the UPGRADE
direction; the export-directory hole is on the DOWNGRADE direction, the
opposite one. It is a **separate-card candidate**.

**Consequence for this SPEC.** No requirement, no acceptance criterion, and no
change of any kind is proposed on its account. `spec.md §E` carries the
corrected mechanism in its exclusions list so a later reader finds it recorded
rather than re-deriving it.

### RESOLVED — G4, newly supplied and out of scope

**What was asked.** Nothing: the dispatch omitted G4 entirely, and `v0.1.0`
recorded that no G4 was ever mentioned. The dispatcher has now supplied it.

**The definition.** Verbatim:

> **G4** = split-brain is guarded by notice rather than prevention —
> `export-json` re-creates `backlog.json` at the canonical path, and while the
> store prefers the database, any consumer that bypasses `BacklogStore` and
> reads the file directly (a human's `cat`, an agent, `backlog_check.sh`) gets
> a stale answer.

**Source.** The dispatcher's ruling on card `t470`.

**The ruling.** **OUT OF SCOPE for `t470`.** Prevention-versus-notice is a
design decision and belongs to its own card.

**Consequence for this SPEC.** Recorded in `spec.md §E` alongside G3 and G5, so
the G-numbering is complete (G1 delivered, G2 absorbed into G1, G3/G4/G5
excluded) and no later reader wonders which items were considered.

## §B Fixture design — the highest-change-likelihood decision

The fixture shape is the decision most likely to be wrong, so it leads.

### F1 — the genuine v3.1.2 record (mandatory, G1 uses this)

`v3.1.2`'s `BacklogRecord` (`internal/kanban/backlog_store.go:75-78` at that
tag) carries exactly three fields:

```go
type BacklogRecord struct {
    Version int           `json:"version"`
    LastSeq int           `json:"last_seq"`
    Items   []BacklogItem `json:"items"`
}
```

It has NO `findings` and NO `archived`. Those two fields exist only on develop
(`backlog_store.go:187-192`). Therefore a fixture described as "v3.1.2 layout
containing archived and findings" describes a state **no real upgrading user
can be in** — it is a develop-schema record planted in the old directory.

F1 is consequently `{version, last_seq, items}` and nothing else, written at
`<root>/.moai/state/kanban/backlog.json`, carrying several items spanning the
three `BacklogState` values that exist on both sides (`queued`, `picked`,
`dropped` — verified present in the v3.1.2 enum) and a `last_seq` above the
highest item id, so continuity is observable rather than coincidental.

### F2 — the forward-compatibility record (optional)

A record carrying `findings` and `archived`, planted in the LEGACY directory.
Reachable only by someone who ran a development build before the release. It
exercises parity of those two fields through the composed path. Scoped as an
OPTIONAL acceptance criterion and labelled with that reachability caveat, so a
reader does not mistake it for a user-facing scenario.

### Schema version — settled, not a decision

`backlogVersion` is `const backlogVersion = 1` on BOTH `v3.1.2`
(`backlog_store.go:47`) and develop (`backlog_store.go:48`). The schema is
additive, so there is no version bump for the migration to negotiate and no
version-mismatch branch to test. Recorded here so the question is not re-opened.

## §C Test-entry design — the second reversible decision

The proof must enter through the CLI, not through `BacklogStore`. Two facts
shape how:

- `internal/cli/todo.go:57` — `todoBacklogPath` calls
  `kanban.BacklogPathForRootAdopting(root)`. This is the ONLY production caller
  of the adopting form, which makes it the composed path's real entrance.
- `internal/cli/todo.go` — `resolveTodoQueueRoot()` calls
  `kanban.ResolveTodoQueueRootAdopting(resolveProjectDir())`, which resolves
  the repository's PRIMARY checkout via git, and falls back to a home-based
  root when git cannot answer.

The second fact is a trap: a bare `t.TempDir()` is not a git repository, so the
queue root would resolve to the home-based fallback and the fixture planted in
the temp project would never be read. The existing suite already solves this —
`internal/cli/todo_queue_root_test.go:89-91`
(`TestResolveTodoQueueRoot_PrimaryIsItself`) uses `initGitRepo(t, primary)` at
L89 plus `t.Setenv("CLAUDE_PROJECT_DIR", primary)` at L91 — the repo-root form.
(The neighbouring `todo_queue_root_test.go:100`
`TestResolveTodoQueueRoot_SubdirectoryResolvesToRepoRoot` calls the same
`initGitRepo` but points `CLAUDE_PROJECT_DIR` at `sub`, not `primary`; it is the
subdirectory variant, not the pattern to copy.) And
`todo_queue_root_test.go:153` seeds a queue document directly and then exercises
the command path. The new test reuses that pattern verbatim, changing exactly
one thing: it seeds under `LegacyStateDirForRoot(root)` rather than
`StateDirForRoot(root)`.

That one-character-of-intent difference is the whole gap. Measured: `grep -rn
"LegacyStateDirForRoot\|state.*kanban" internal/cli/` returns no test seed, so
no existing CLI test plants the legacy directory.

Test file placement: `internal/cli/`, because that is where the entry point and
its isolation helpers (`initGitRepo`, `userHomeDirFn`) live. Placing it in
`internal/kanban/` would force it back onto the API path this card exists to
step outside of.

## §D Milestones

Ordered by decision-reversibility: the shape decisions first, the mechanical
steps last.

### M1 — G1, the composed-path proof (Priority: High)

Add one CLI-layer test that materializes the F1 layout, runs the first
`moai todo` invocation against it, and asserts the composed outcome:

- every card, its state, and `last_seq` survive the composition (REQ-QUP-001)
- the legacy directory is gone and the current-name directory holds the queue
  (REQ-QUP-002)
- the legacy document is present under `backlog.json.migrated` (REQ-QUP-003)
- the SQLite artifact `backlog.db` exists beside it (REQ-QUP-004)

Each assertion is independently checkable by a command, per the AC matrix in
`acceptance.md`.

### M2 — F2, the forward-compatibility criterion (Priority: Low, OPTIONAL)

Extend the proof with a second fixture carrying `findings` and `archived` in
the legacy directory, asserting both survive. Labelled optional in
`acceptance.md` with its reachability caveat. Skipping M2 does not fail the
card.

### M3 — record what was measured (Priority: Medium)

Write the run-phase evidence at `.moai/reports/t470/verdict.md`: the commands
run, their verbatim output, the baseline attribution, and the gaps.

The G-numbering is now complete (§A), so the verdict states it in full: G1
delivered; **G2 absorbed into G1** — covered by the criteria named in §A's
resolution table, therefore NOT a gap; G3 (cross-process concurrency), G4
(split-brain guarded by notice), and G5 (`moai doctor` check) excluded and
named as not covered.

## §E Technical approach

- **Cycle**: TDD. The test is the deliverable, so RED is the natural starting
  state — write the composed-path assertions first, confirm they exercise a
  path no existing test reaches, then confirm GREEN against unmodified
  production code. A GREEN-on-first-run result is expected here and is not a
  vacuous pass **only if** RED was established first by a deliberate mutation
  (see §F).
- **No production edits.** If any assertion fails, the finding is reported as a
  blocker, not repaired inside this card. A failing composed path is a defect
  discovery that changes the card's premise and belongs to the dispatcher.

## §F Risks

| Risk | Consequence | Mitigation |
|---|---|---|
| Vacuous green — the test passes without ever entering the composed path | The proof asserts nothing | Establish RED first by the mutation `AC-QUP-010` names — pre-create an EMPTY `<root>/.moai/state/todo/` alongside the seeded legacy directory, which makes `resolveStateDir` take the stale-copy branch (`state_dir.go:81-88`) so no relocation runs and `AC-QUP-002` fails with the legacy directory still present. Record the mutation and its verbatim failure output in the verdict. Note the mutation explicitly REJECTED there: seeding under the current directory name instead of the legacy one produces GREEN on every criterion. |
| `resolveTodoQueueRoot` resolving to the real repository | The test reads or mutates this repository's live queue | `initGitRepo(t, t.TempDir())` + `t.Setenv("CLAUDE_PROJECT_DIR", …)` + `userHomeDirFn` override, exactly as `todo_queue_root_test.go` does. Assert the resolved root is inside the temp dir before proceeding. |
| `./internal/cli/` exceeding the default Bash timeout | The verification appears to fail when it merely ran long | Run the package with an explicit timeout at or above 600s, and record the elapsed time alongside the result. |
| Scope creep into repair | The card stops being a proof | REQ-QUP-009 and the `§E Exclusions` section of `spec.md`. Any behavior-changing impulse goes to the exclusions list with its reason. |

## §G Anti-patterns

- Inventing a defect in a sound mechanism in order to have something to fix.
- Writing the F1 fixture with `findings` or `archived` — that record shape is
  unreachable for a real v3.1.2 user and would make G1 prove a fiction.
- Asserting only "the cards survived" and skipping the side effects: the
  relocation, the quarantine, and the artifact are the composition's evidence.
- Running the full local suite to verify (C-2).

## §H Cross-references

- `internal/kanban/state_dir.go` — directory layer, `resolveStateDir`
- `internal/kanban/backlog_migrate.go` — storage layer, `migrateLegacyBacklog`
- `internal/cli/todo.go` — the composed path's CLI entrance
- `internal/cli/todo_queue_root_test.go` — the isolation pattern to reuse
- `internal/kanban/backlog_concurrency_test.go:135` — the in-process stress
  test whose existence is why G3 is framed as a cross-PROCESS gap
