# Plan — SPEC-BACKLOG-JSON-DISCLOSURE-001

Milestones are ordered by decision-reversibility: the shape decisions that are
expensive to change late come first, the mechanical mirror sweep last.

## §A Context

Card t395, re-aimed by the lead on 2026-09-02 after its dispatched premise was
disproven. The evidence base is three reports already landed on this branch:
`.moai/reports/t395/premise-verdict.md` (the premise reversal and the
unidentified-writer Gap), `.moai/reports/t395/reader-surfaces.md` (the four
reader sites and the reuse candidate), `.moai/reports/t395/r1-repro.md` (the
observed dead monitor, with a control arm).

Worktree `.claude/worktrees/t395`, branch `WT-stale-backlog-json`, based on
`origin/develop@ad272be20`.

## §B Known issues carried in

- The writer of the present `backlog.json` is **unidentified**. This is a
  recorded Gap, not a loose end this SPEC closes — see `spec.md` §A.3 for why
  disclosure rather than cleanup follows from it.
- `.moai/reports/t395/r1-repro.md` observed Case A only. Case B is promoted to
  AC-BJD-009 rather than assumed.
- The control arm in that report showed a `backlog.db` cksum moving on a
  mutation, but SQLite WAL means that is not guaranteed in general. AC-BJD-010
  is the criterion that decides whether the naive cksum-on-`backlog.db` repair is
  sufficient or whether the watch target must cover the WAL.
- **The deferred state has to be built on purpose, and building it is not
  obvious.** `wal_autocheckpoint` is per-connection and unset anywhere in
  `internal/kanban`, so a side connection cannot restrain the one `moai todo`
  opens; and a short-lived `moai todo` process may fold the WAL back on exit,
  which would explain why the control arm saw the cksum move at all. Neither the
  autocheckpoint default nor any checkpoint-on-close behaviour has been measured
  in this tree — both are candidate explanations the run-phase settles. AC-BJD-010
  therefore requires the Given be **evidenced**, not assumed, and treats a failure
  to build it as a Gap.

## §C Pre-flight

1. `git -C <worktree> rev-parse --short HEAD` and `git branch --show-current`
   re-read immediately before any commit (shared-checkout staleness rule).
2. Confirm the fixture convention: no acceptance work touches
   `/Users/goos/MoAI/moai-adk-go/.moai/state/todo/`.
3. Confirm `moai` on PATH is a post-SQLite build, since M2's fixture mutates a
   queue through the real write path.

## §D Constraints

- **Reuse, not invention.** `InspectBacklogArchiveVouch` is extended in place
  (REQ-BJD-006). `inspectBacklogLayout` already computes `jsonExists`; the fact
  is measured today and discarded, so the change is a field and a branch, not a
  new probe.
- **Read-only.** No migration, no DDL, no lock on the disclosure path.
- **stderr only.** stdout is a machine surface for `moai todo` and must stay
  byte-identical.
- **Template-First** (`CLAUDE.local.md` §2). `internal/template/templates/` is
  the source of truth for distributed files; `make build` re-embeds.
- **No engine, migration, schema, or downgrade-route change** (`spec.md` §C).
- Verification is package-scoped locally; the full-suite verdict is CI's.

### §D.1 Disclosure breadth — DECIDED

**Decided 2026-09-02 by the operator, relayed through the lead. Read verbs in
full: every verb that reads the queue and prints, not the `todo history`
precedent alone.**

**The reason is evidence, not preference, which is why it is recorded here.** The
lead walked into this defect through `moai todo` — the list verb — on the day of
the decision: the output ran to 107KB, was truncated, and the lead fell back to
reading `sqlite3 backlog.db` directly. Disclosing on `history` alone would leave
silent the one verb people actually reach for, which is the exact incident shape
this SPEC exists to prevent. A precedent-shaped scope would have shipped a
disclosure nobody encounters.

- **The criteria are unchanged, and that is the point.** REQ-BJD-002 already
  bound "a `moai todo` **read** command"; the decision names which read verbs,
  not a different surface. AC-BJD-002 verifies a read command and needs no
  rewrite. Nothing widened because the floor and the answer coincide — the
  question was never read-vs-write, it was precedent-vs-full-read-surface.
- **Write verbs stay out.** They hold the queue lock and carry a different
  stdout contract; the decision did not reach them, and this SPEC does not.
- **Run-phase enumerates; it does not choose.** The concrete verb list is
  derived from the code at run-phase and its derivation basis recorded in M3.
  Naming the verbs here from memory would substitute recall for measurement —
  the same unverified-list mistake this SPEC has already corrected twice, in a
  new place.

No clarification-gate marker is placed on this section: the decision has landed,
so there is nothing left to clarify.

(The marker's literal bracketed token is deliberately not written anywhere in
this file: the plan-auditor's clarification-gate check greps for that token, so
writing it even to say it is absent would trip the gate it describes.)

## §E Self-verification

Before the run-phase reports any milestone complete, each of its criteria is
cited with the command run and its verbatim output, with the measurement
attributed to this tree and this run.

## §F Milestones

### M1 — the disclosure fact and its shape *(highest reversibility cost)*

The one decision here that is expensive to revisit is **the return type's
shape**: `BacklogArchiveVouch` is already consumed by `internal/cli/todo_history.go`,
so the new fact must be added without disturbing that consumer's semantics.

- Extend `BacklogArchiveVouch` with a field reporting that a non-authoritative
  `backlog.json` sits beside the answering SQLite store (`backlog_archive_vouch.go:41`).
  `InspectBacklogArchiveVouch` populates it from the `layout.jsonExists` value it
  already has at `:57` — the SQLite branch currently discards it.
- The field is meaningful **only** on the SQLite branch. On
  `BacklogStoreLegacyJSON` the JSON *is* the answering store; on
  `BacklogStoreNone` there is nothing. Both keep the field false, and that is
  asserted, not left implicit.
- Existing `todo_history.go` behaviour is unchanged by this milestone.

Criteria: AC-BJD-001, AC-BJD-006, AC-BJD-007.

### M2 — the queue watch repair *(behavioural; decides the watch target)*

The second irreversible-ish decision: **what the watch actually observes**. The
naive answer (`cksum backlog.db`) is what the r1-repro control arm exercised, but
its Residual-risk names WAL deferral as an unverified assumption. Settle
AC-BJD-010 before committing to a target.

- Establish the red first: run the shipped script verbatim against a fixture
  queue and observe zero events across a real mutation. This reproduces
  `r1-repro.md` §B and is what makes AC-BJD-008's falsifiability condition
  demonstrable rather than asserted.
- Repair `.claude/skills/moai-kanban-foreman/SKILL.md:95` to watch the
  authoritative store.
- Run Case B (AC-BJD-009) — the gap `r1-repro.md` recorded and did not close.
- Measure WAL behaviour (AC-BJD-010). **Build the deferred state first and prove
  it holds** — `backlog.db-wal` non-empty and `backlog.db` cksum unmoved — before
  arming the watch; a run that skips this measures an already-checkpointed state
  and passes having tested nothing. If a committed-but-uncheckpointed write is
  invisible to the chosen target, widen the target to cover the deferral. Do not
  widen the observation window, and do not report a pass whose Given never held.

Criteria: AC-BJD-008 (BLOCKING), AC-BJD-009, AC-BJD-010.

### M3 — the disclosure surface

- Emit the disclosure from the `moai todo` **read** surface in full, on stderr,
  following the shape `internal/cli/todo_history.go:99` already established.
  §D.1 decided this scope; the run-phase does not re-open it and does not widen
  past it.
- **Enumerate the read verbs from the code, and record the basis.** Derive the
  list from the verb registration site rather than from recall or from this
  plan's prose, and write into this milestone's report both the enumeration and
  how it was derived — the command run and what it returned. A verb list that
  cannot say where it came from is an unattributed claim, and this SPEC has
  already been bitten twice by lists that looked complete.
- Keep the fixture's `backlog.db` archive tables present, so the existing
  REQ-TAQ-013 line at `todo_history.go:99-107` cannot co-fire and make
  AC-BJD-002's count ambiguous.
- Assert stdout parity against a no-JSON layout (AC-BJD-003) rather than merely
  asserting the line lands on stderr.
- Assert the file is untouched by sha256 and mtime (AC-BJD-005).

Criteria: AC-BJD-002 (BLOCKING), AC-BJD-003 (BLOCKING), AC-BJD-004, AC-BJD-005
(BLOCKING).

### M4 — the reader instructions

Three sentence-level edits across two files, correcting the state location to
`backlog.db`. Mechanical, but it is the edit that fixes the *cause* of the
lane-10 stale read, so its check (AC-BJD-012) is written to reject a residual
mention that is not explicitly labelled export-or-legacy.

`.moai/docs/todo-queue-storage.md` is **not** edited — it is already correct, and
AC-BJD-014 asserts it stays byte-identical.

Criteria: AC-BJD-012 (BLOCKING), AC-BJD-013, AC-BJD-014.

### M5 — the template mirror *(mechanical; lowest reversibility cost)*

- Mirror the M2 and M4 edits into the **three** `internal/template/templates/`
  copies carrying the four sites (`workflows/todo.md` holds two of them).
- `make build`.
- Assert the mirror is complete with **both** greps of AC-BJD-015 and the
  embedded-parity check of AC-BJD-016. The second grep is not optional: the
  home-fallback site's path shape does not match the first pattern, so a
  single-grep run leaves `todo.md:21` unrepaired and still reports complete.

Criteria: AC-BJD-011, AC-BJD-015 (BLOCKING), AC-BJD-016.

## §G Anti-patterns

- **Cleaning up the file.** Deleting or truncating `backlog.json` is out of scope
  and does not survive the unidentified writer (`spec.md` §A.3).
- **A second disclosure mechanism.** A new inspector, a new probe, or a
  disclosure computed independently of `InspectBacklogArchiveVouch` violates
  REQ-BJD-006 and is rejected regardless of how clean it reads.
- **A watch-target check that passes both ways.** AC-BJD-008 is satisfied only
  when reverting to `backlog.json` turns it red. A static grep that merely asserts
  the string `backlog.db` appears somewhere in the block is not that check unless
  it demonstrably fails on the shipped block.
- **Widening the window to pass AC-BJD-010.** A timing-sensitive pass is an
  unobserved claim about WAL behaviour.
- **Passing AC-BJD-010 without its Given ever holding.** The more likely vacuity
  path, and distinct from the one above: `moai todo add` followed by a watch may
  measure an already-checkpointed state, so the criterion goes green having tested
  nothing about deferral. The two evidence values (`backlog.db-wal` non-empty,
  `backlog.db` cksum unmoved) are what establish the Given; without both observed
  the result is a Gap.
- **Checking the mirror with one grep.** The four sites do not share a path
  shape. A single `state/todo/backlog\.json` pattern cannot see the home-fallback
  site, so it reports completeness over three of four. A completeness check blind
  to an item in its own scope is the defect, not the wording around it.
- **Picking the disclosure breadth in run-phase.** Still forbidden — the breadth
  is now fixed in §D.1 by operator decision, so the run-phase **enumerates
  against that decision** rather than choosing. Narrowing to the `history`
  precedent and widening to the write verbs are both out of bounds.
- **Enumerating the read verbs from memory.** The list is derived from the code
  and its derivation recorded (M3). An enumeration that cannot cite where it came
  from is the same unverified-list defect in a new place.
- **Editing `.moai/docs/todo-queue-storage.md:20` or `:55`.** Both are correct
  and are the control. They share a string with the defect sites, which is
  exactly why a careless sweep reaches them.
- **Editing `.claude/` without the template mirror**, or the mirror without
  `make build`. Both leave the distributed user with the defect this SPEC exists
  to remove.
- **Touching the live queue.** Any acceptance work against
  `/Users/goos/MoAI/moai-adk-go/.moai/state/todo/` is a defect in the
  verification, not a shortcut.

## §H Cross-references

- `.moai/reports/t395/premise-verdict.md` — premise reversal, unidentified-writer Gap
- `.moai/reports/t395/reader-surfaces.md` — the four reader sites, R4 reuse candidate, R5 template mirror
- `.moai/reports/t395/r1-repro.md` — the observed dead monitor and its control arm
- `.moai/docs/todo-queue-storage.md` — the correct storage document (unmodified)
- `internal/kanban/backlog_archive_vouch.go` — the mechanism being extended
- `internal/cli/todo_history.go:84,99` — the disclosure precedent
- `CLAUDE.local.md` §2 — Template-First mirror obligation
- SPEC-TODO-SQLITE-001 — the cutover this SPEC's reader surfaces failed to follow
- SPEC-TODO-ARCHIVE-QUERY-001 — REQ-TAQ-013, the sibling disclosure this extends
