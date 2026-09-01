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
- Measure WAL behaviour (AC-BJD-010). If a committed-but-uncheckpointed write is
  invisible to the chosen target, widen the target to cover the deferral. Do not
  widen the observation window to make a marginal case pass.

Criteria: AC-BJD-008 (BLOCKING), AC-BJD-009, AC-BJD-010.

### M3 — the disclosure surface

- Emit the disclosure from the `moai todo` command surface, on stderr, following
  the shape `internal/cli/todo_history.go:99` already established. Which verbs
  carry it is a scoping decision to state explicitly in the run-phase report; the
  criteria are written against "a `moai todo` read command", so at minimum the
  read surface an operator reaches for must carry it.
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

- Mirror the M2 and M4 edits into the four `internal/template/templates/` copies.
- `make build`.
- Assert the mirror is complete by the grep of AC-BJD-015 and the embedded-parity
  check of AC-BJD-016.

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
