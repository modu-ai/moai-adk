# Implementation Plan — SPEC-QUEUE-UPGRADE-PROOF-001

Card: `t470` · Tier M · Branch `WT-queue-upgrade-proof` · Base `4e4607abe`

## §A Open clarifications

These block Implementation Kickoff Approval. They are stated as questions for
the dispatcher, not as decisions this SPEC makes.

### [NEEDS CLARIFICATION: G2 definition]

The dispatch named "G2: close alongside G1" and never said what G2 is. No G4
was mentioned at any point. This SPEC does not guess: G2 is left undefined, no
requirement is written for it, and no acceptance criterion covers it. The
orchestrator has queried the dispatcher; the definition will be injected before
run-phase entry, at which point G2's requirements and ACs are appended.

Consequence if it arrives unresolved: the run phase delivers G1 (plus the
optional F2 criterion) and nothing else, and the card closes with G2 recorded
as unstarted rather than silently dropped.

### [NEEDS CLARIFICATION: downgrade intent vs quarantine rename]

Two mechanical facts, both measured at tree `4e4607abe`:

1. `internal/kanban/state_dir.go`, on the `backlogFileName` constant: the queue
   document "stays `backlog.json` after the storage swap … keeping this name is
   what makes the downgrade story literally true — an older binary reads only
   this file and ignores the rest."
2. On a successful migration, `migrateLegacyBacklog` calls
   `quarantineLegacyBacklog`, which renames that file to
   `backlog.json.migrated` (`backlogMigratedSuffix`,
   `internal/kanban/backlog_migrate.go:41`). Separately, the directory itself
   has moved from `.moai/state/kanban/` to `.moai/state/todo/`.

Taken literally together, an older binary after an upgrade would find neither
the directory nor the file. The comment may carry a narrower meaning than it
reads — for instance that the name is preserved during the pre-migration
window, or that it is about the sibling `.db` artifact not colliding.

**This SPEC does not resolve the tension and proposes no change on its
account.** If the tension is real it is larger than this card; if it is not, it
is a comment-wording matter. Either way the judgment is the dispatcher's. It is
recorded here so that a future reader who notices the same thing finds it
already logged rather than re-deriving it.

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
`internal/cli/todo_queue_root_test.go:100` uses `initGitRepo(t, primary)` plus
`t.Setenv("CLAUDE_PROJECT_DIR", primary)`, and
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
run, their verbatim output, the baseline attribution, and the gaps —
explicitly naming G2 (undefined), G3 (excluded), and G5 (excluded) as not
covered.

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
| Vacuous green — the test passes without ever entering the composed path | The proof asserts nothing | Establish RED first by a deliberate mutation (e.g. temporarily seed under the current directory name instead of the legacy one, and confirm the relocation assertion fails). Record the mutation and its observed failure in the verdict. |
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
