# Plan — SPEC-TODO-ARCHIVE-QUERY-001

Milestones are ordered by decision-reversibility: the output contract first
(hardest to change once an operator scripts against it), then the invariant
guards, then the mechanical work.

## §A Context

Read `spec.md` §A first. The one-line summary: the archive exists, is loaded into
memory on every read, and has no reader. This plan adds the reader.

Tree this plan was written against: `origin/develop@2c18091d1`, worktree
`.claude/worktrees/t394`, branch `WT-todo-done-history`.

## §B Known issues carried in from the measurement

1. **~289 issued ids were destroyed in the pre-archive era and are held by no
   store.** The binary installed until `2026-09-01 06:37 KST` — a `v3.1.2`
   main-era build — deleted on `done`; under it the database carried no archive
   tables. Measured at the plan-phase measurement turn
   `2026-09-01T~04:11+09:00`: `meta.last_seq` = 401, `count(*) FROM items` =
   113, `.tables` = `findings items meta` → 288 unaccounted. Re-measured
   `2026-09-01T04:30:18+09:00`: `402` / `113` → **289**. Both figures are
   snapshots of live counters taken in the **primary checkout** under the
   then-installed `v3.1.2` — the paths do not exist in this worktree.
   **Corrected 2026-09-01 (lead re-measurement `16:20+09:00`)**: the `06:37`
   develop install archives instead of deleting — the queue now carries
   `archived_items` with 11 rows, and the corrected arithmetic
   (`408 − 108 − 11`) still gives **289**, so no card has been destroyed since
   the install. The 289 is a **pre-archive-era estimate**, frozen as of that
   install — it no longer grows on its own. Full coordinates, including the
   post-install snapshot, are in `spec.md` §A.4 under
   `[MEASUREMENT PROVENANCE]`; do not carry the digits forward without
   re-measuring. The verb still ships into a world holding ~289 destroyed
   cards it cannot name individually, which is why REQ-TAQ-004's disclosure is
   a requirement and not a nicety — without it the verb answers `absent` for
   every pre-archive-era id and reads as authoritative. Note the disclosure is
   keyed on `last_seq`, **not** on archive emptiness: the DDL's `IF NOT EXISTS`
   creates the archive on the first open by a newer binary
   (`internal/kanban/backlog_sqlite.go:92-94`), so an emptiness test would go
   silent after the first `done` while the destroyed cards stay destroyed.
2. **There is no queue-path divergence. The earlier claim here was wrong and is
   withdrawn.** An earlier draft recorded `.moai/state/todo` vs
   `.moai/state/kanban` as a live migration risk, citing
   `backlog_store.go:250`. Re-measured at `2c18091d1`: `BacklogPathForRoot` is at
   `internal/kanban/state_dir.go:129`; `state_dir.go:37` is
   `const stateDirName = "todo"` and `state_dir.go:43` is
   `const legacyStateDirName = "kanban"` — `todo` is canonical, `kanban` is the
   retired name, the inverse of the draft's claim. `resolveStateDir`
   (`state_dir.go:79-104`) returns the canonical directory unconditionally once
   it exists. Nothing here couples this SPEC to a path change, because there is
   no path change. See `spec.md` §E.
3. **A stale pre-archive `backlog.json` sits beside the database.** Measured at
   the plan-phase measurement turn `2026-09-01T~04:11+09:00`, primary checkout,
   then-installed `v3.1.2` (`spec.md` §A.4 provenance):
   `.moai/state/todo/` holds `backlog.db` (09-01 02:37), `backlog.json`
   (08-31 21:13) and `backlog.json.migrated` (08-27 23:01); the stale
   `backlog.json` parses with no `archived` key and 109 items against the
   database's 113 at that instant. The durable fact is the missing `archived`
   key; the counts and mtimes move, and the gap widens on its own because the
   JSON is frozen while the database is not. Every read path prefers the
   database when it exists
   (`backlog_store.go:411-427`, `:551-568`), so the stale file cannot answer
   today — but where the database is absent it answers authoritatively with no
   archive at all. That is what REQ-TAQ-013's second clause exists for. Card
   **t395** owns removing or reconciling the file; this SPEC does not touch it.
4. **The dispatch's queue figures do not reproduce — and neither will these.**
   Measured 112 live cards and 156,468 bytes of `list` output at the plan-phase
   measurement turn `2026-09-01T~04:11+09:00` (primary checkout,
   then-installed `v3.1.2`), against the dispatch's 109 and ~70KB. But the reason the dispatch's
   figures did not reproduce is the reason these will not either: **the queue is
   a live counter**, and this item existed to warn about carrying figures forward
   while doing exactly that. `112` and the `113` in item 1 above are two instants
   of the same population, not a disagreement — `list` applies no state filter
   (`internal/cli/todo.go:310-311` @ `2c18091d1`), and at
   `2026-09-01T04:30:18+09:00` both commands returned `113` within one second
   (`spec.md` §A.4 snapshot B). Anyone re-deriving the default-output argument
   re-measures; the argument rests on the order of magnitude, which is stable,
   not on the digits, which are not.

## §C Pre-flight

Before the first edit of the run phase:

```
git rev-parse --short HEAD
git branch --show-current
git fetch origin develop && git rev-list --count --left-right origin/develop...HEAD
```

Proceed on `0 0` or `0 N`. On `N 0` or `N M`, resolve before editing.

**`origin/develop` has already moved.** At plan-audit iteration 1,
`git rev-list --count --left-right origin/develop...HEAD` read `21 0` — develop
was 21 commits ahead of `2c18091d1`, the tree every citation in these artifacts
was measured against. The artifacts deliberately keep citing `2c18091d1` rather
than a tree nobody measured. The run phase therefore **must re-run this §C batch
against the then-current develop and re-verify, at that head, every `file:line`
these artifacts carry** before the first edit — specifically
`internal/kanban/state_dir.go:37,43,79-104,129`,
`internal/kanban/backlog_sqlite.go:92-94,113,124,133`,
`internal/kanban/backlog_store.go:14-17,223,411-427,551-568,772-778`,
`internal/kanban/backlog_migrate.go:102,411`,
`internal/cli/todo.go:148-152,310-311,409`, `internal/cli/todo_why.go:34-37`,
`internal/cli/todo_export.go:75,100-110`, `internal/cli/todo_undone.go:68`. A
citation that no longer resolves is refreshed at the new head and the drift is
recorded in `progress.md` §E.2; a citation whose **meaning** changed is a
blocker, not a refresh.

## §D Constraints

- **No storage change.** REQ-TAQ-012. If a milestone appears to need one, the
  milestone is wrong, not the constraint.
- **No change to default reads.** REQ-TAQ-011, guarded by golden files.
- **AC-TAQ-011 clause 1 depends on the merge method, and the dependency is
  recorded here because it did not exist before this SPEC created it.** Clause 1
  asserts that the goldens' introducing commit `C` is a *proper ancestor* of
  `HEAD` whose tree lacks the verb — a statement about M0 preceding M1. A
  **squash** collapses the two into one commit whose tree holds both, so `C`
  becomes `HEAD` and the clause fails permanently on the branch it lands on
  (measured against a throwaway repository: a single commit holding goldens and
  verb resolves `C == HEAD`, and `git cat-file -e "$C:<verb>"` exits `0`, so the
  negated form fails). A guard red forever is a guard that gets deleted.

  Two consequences, in order of who owns them:

  1. **Clause 1 is scoped to the pre-integration branch** — this card's branch
     and its pull request — and `acceptance.md` now says so. It is not asserted
     on `develop` or `main`, because M0-before-M1 is a within-card property and
     nothing on an integration branch can restore it once history is rewritten.
  2. **Today's merge path happens to preserve it, but that is a fact about the
     process, not a guarantee this SPEC may lean on.** Card branches reach
     `develop` by a local `git merge --no-ff` (CLAUDE.local.md §4.1), and
     `develop` reaches `main` by a `release/*` pull request, which
     `.github/workflows/auto-merge.yml:262-269` merges with `--merge` rather than
     `--squash` (`--squash` is the default for every non-release branch). Both
     preserve M0 and M1 as distinct commits. If this card is ever taken to an
     integration branch by a squash instead, clause 1 must be dropped there
     rather than repaired — its content has been destroyed, not broken.
- **Template-First.** The skill doc lives in two places; the template mirror is
  edited in `internal/template/templates/`, then `make build`.
- **Read-only verb.** No `Mutate`, no lock.
- Local verification is scoped to `./internal/cli/...` and `./internal/kanban/...`.
  Do not run the full suite locally; push and read CI for the full verdict.

## §E Self-verification

Each milestone below closes only when its named ACs pass with the command that
decided them cited. A milestone whose AC was not run is not complete.

## §F Milestones

### M0 — capture and commit the goldens, before any verb code exists

**This milestone is first for a mechanical reason, not a stylistic one.**
AC-TAQ-011 clause 1 requires the goldens' introducing commit `C` to be a proper
ancestor of `HEAD` whose tree contains neither `internal/cli/todo_history.go` nor
the symbol `newTodoHistoryCmd`. A golden captured after the verb is written
cannot satisfy that, so the capture cannot be deferred to M4 — it must land in
its own commit before M1 writes a line.

Build the current binary, run `list`, `list --json`, `next`, `why <id>`,
`analyze` and the state counts against the `FIXTURE`, write each stream to
`internal/cli/testdata/golden/live-readers/`, and commit **that directory alone**
with no other change. Record `C = git rev-parse HEAD` in `progress.md` §E.2.

Delivers nothing user-facing. Enables AC-TAQ-011.

### M1 — the output contract and the four fates

**Highest change-likelihood: this is the decision an operator scripts against.**

Decide and implement the line shape for both forms. The shape must let a reader
distinguish all four fates of `spec.md` §A.3 without a second call, and must place
the card text LAST so a consumer reading the tail is unaffected if a column is
ever added — the convention `pr` already holds
(`.claude/skills/moai/workflows/todo.md` line 51).

Proposed shape, tab-separated, to be confirmed in review:

```
<id>	live	queued|picked|dropped	<text>
<id>	archived	<state-at-archive>	<text>
<id>	absent
```

Delivers REQ-TAQ-001, 002, 003, 005, 006, 009.
Closes AC-TAQ-001, 002, 003, 005, 006, 009.

New file: `internal/cli/todo_history.go`. Registration edit:
`internal/cli/todo.go:148-152`.

**Naming contract — the constructor is `newTodoHistoryCmd`.** This is not a style
preference; two acceptance criteria are decided by grepping for that exact
symbol, and no requirement mandates it, so it is fixed here instead:

- **AC-TAQ-011 clause 1** asserts `! git grep -q "newTodoHistoryCmd" "$C" -- internal/cli/`.
  Under a different constructor name that grep finds nothing at `C` *and*
  nothing anywhere, so the clause passes **vacuously** on a tree where the verb
  exists — the failure mode is silent, which is the dangerous direction.
- **AC-TAQ-014** locates its scan subject with
  `SRC=$(git grep -l "newTodoHistoryCmd" …)` and then asserts `test -n "$SRC"`.
  A different name makes that AC fail for a naming reason rather than a
  prompting one — noisy, and therefore the safer of the two.

The name follows the existing convention rather than inventing one: all sixteen
registered `todo` verbs are built by `newTodoXxxCmd()` constructors
(`internal/cli/todo.go:148-152` @ `2c18091d1`). If the run phase has cause to
name it otherwise, the AC commands must be changed **in the same commit** — the
alternative, matching on the Cobra command's `Use` string instead of the Go
symbol, is available and equally mechanical.

### M2 — the honest limits

The two stderr disclosures and the degraded path. These carry the SPEC's honesty
obligations and are separated from M1 because each is a judgement about what the
operator must be told, not a rendering detail.

- REQ-TAQ-004 — `absent` is qualified for an id at or below `last_seq`. Read
  `rec.LastSeq`, which `readRecord` already populates
  (`internal/kanban/backlog_migrate.go:106-110`); parse the looked-up id's
  ordinal with the same helper `normalizeBacklogRecord` uses
  (`parseBacklogSeq`). No new read, no new field.
- REQ-TAQ-008 — truncation states the withheld count.
- REQ-TAQ-013 — a store that cannot vouch for an archive says which store
  answered rather than failing. Two reachable inputs: archive tables dropped, and
  a legacy `backlog.json` with no `backlog.db`. The store already knows which it
  is — `inspectBacklogLayout` (`internal/kanban/backlog_store.go:411`,
  `:552`) — so this is a rendering of an existing fact.

  **Ordering coupling to check in implementation**: `readRecord` calls
  `readArchive` at `backlog_migrate.go:102` and `readLastSeq` at `:106`. If a
  missing-archive error returns before `:106`, the degraded path loses
  `rec.LastSeq` and REQ-TAQ-004 goes silent exactly where it is most needed.
  The degraded path must still obtain `last_seq`.

All three write to stderr only, so a machine reading stdout is unaffected.

Closes AC-TAQ-004, 008, 013.

### M3 — the bound

`--limit`, default 20, `0` meaning unbounded. Deliberately after M2: the bound is
mechanical, but its truncation notice (M2) is the part that keeps a bounded read
from being mistaken for a complete one.

Delivers REQ-TAQ-007. Closes AC-TAQ-007.

### M4 — the invariant guards

The goldens were captured and committed in M0; this milestone writes the
assertions over them — clause 1 (provenance **and integrity**: the singleness of
the adding commit, its ancestry, the absence of the verb in its tree, and that
the goldens' bytes have not moved since), clause 3 (post-change comparison),
plus the schema assertion and the read-only assertion. Clause 2 (regeneration
from `C`) is run locally here and its diff output cited in `progress.md` §E.2.
Do NOT re-capture the goldens in this milestone: a re-capture from this tree is
exactly the self-certification AC-TAQ-011 exists to prevent.

These are written LAST in ordering but must pass before the change is pushed —
they are what proves the change took nothing away.

Delivers REQ-TAQ-010, 011, 012, 014. Closes AC-TAQ-010, 011, 012, 014.

### M5 — documentation, both surfaces

Add the verb's row to `.claude/skills/moai/workflows/todo.md` and its template
mirror, then `make build`. The mirror carries no SPEC ID, REQ token, date, or SHA.

Delivers REQ-TAQ-015. Closes AC-TAQ-015.

## §G Anti-patterns

- **Adding `--archived` to `list`.** It looks smaller than a new verb and is
  larger: it puts the archive one flag away from every existing caller of the
  queue's cheapest read, and REQ-TDG-007 then has to be re-argued per caller.
- **Reaching for `export-json` in the implementation.** The data is already in
  `rec.Archived` after `Load()`; calling the downgrade route would write a file.
- **Treating `dropped` as a kind of archived.** It is a live row. Conflating them
  is the defect this SPEC exists to close, reintroduced in the fix.
- **Reporting `absent` without the empty-archive qualifier.** The verb's most
  dangerous output is a confident false negative.
- **Skipping the golden capture because the change "only adds a verb".** A shared
  render helper touched in M1 is exactly how a default read grows by accident.
- **Capturing the goldens after writing the verb.** The comparison then has one
  side, and a broken default read certifies itself. M0 exists to make this
  mechanically impossible; AC-TAQ-011 clause 1 is what catches it if M0 is
  skipped.
- **Keying REQ-TAQ-004's disclosure on the archive being empty.** It reads as the
  obvious test and self-disables after the first `done`, while the destroyed
  cards stay destroyed. AC-TAQ-004's second clause fails an implementation that
  does this.

## §H Cross-references

- `spec.md` — requirements and scope boundaries.
- `acceptance.md` — the AC enumeration and the command deciding each.
- `.moai/reports/t394/premise-check.md` — the nine measurements this plan rests on.
- `.moai/specs/SPEC-TODO-DESTRUCTIVE-GUARD-001/spec.md` — REQ-TDG-003/004/005/007/015.
- Card **t395** — the stale sibling `backlog.json` inside `.moai/state/todo/`.
  A dependency note only: this SPEC discloses the consequence at read time
  (REQ-TAQ-013) and absorbs none of the repair. The queue-path divergence an
  earlier draft attributed to t395 does not exist (§B.2).
