---
id: SPEC-BACKLOG-JSON-DISCLOSURE-001
title: "A backlog.json at the canonical path is not the queue — disclose it, and stop reading it"
version: "0.2.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec (card t395)
priority: P2
phase: "v3.1.5 target"
module: "internal/kanban, internal/cli, .claude/skills/moai/SKILL.md, .claude/skills/moai/workflows/todo.md, .claude/skills/moai-kanban-foreman/SKILL.md, internal/template/templates"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, sqlite, disclosure, foreman, template-mirror, stale-read"
tier: M
related_specs:
  - SPEC-TODO-SQLITE-001
  - SPEC-TODO-ARCHIVE-QUERY-001
  - SPEC-KANBAN-TODO-CLI-001
  - SPEC-TODO-LANDING-STATE-001
---

# SPEC: A backlog.json at the canonical path is not the queue

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.2.0 | 2026-09-02 | Plan-audit iter-1 repair (`.moai/reports/t395/plan-audit.md`, PASS-WITH-DEBT 0.85); five blocking-class defects closed, wording only, no redesign. **D5**: AC-BJD-015's single regex structurally could not see the fourth defect site — `~/.moai/todo/<project-key>/backlog.json` does not match `state/todo/backlog\.json`, so the template mirror's only BLOCKING completeness check passed while `todo.md:21` stayed unrepaired; the criterion now enumerates its sites and runs two greps. **D2**: AC-BJD-007 delegated a universal negative to human judgement; replaced with two runnable commands and their expected values. **D4**: AC-BJD-010's Given had no construction technique, so it could be satisfied by measuring an already-checkpointed state; the Given is now constructible and self-evidencing, and a construction that fails yields a Gap rather than a pass. **D3**: REQ-BJD-002 / AC-BJD-002 / plan M3 split three ways on verb breadth; unified at the read-surface floor, with the breadth recorded as an open operator decision that widens rather than contradicts. **D1**: AC-BJD-008's rebinding named a single `f=` variable, which the multi-target repair AC-BJD-010 permits would invalidate; rebinding is now directory-based. Optional **D6** (four sites live in three files; REQ-BJD-003 mislabelled `Where` for what is runtime disk state), **D7** (§A.2 inherited an R5 citation whose grep did not cover the site it was cited for — conclusion true, attribution wrong), and **D8** (an existing REQ-TAQ-013 stderr line could make "exactly one" ambiguous) also closed. |
| 0.1.0 | 2026-09-02 | Initial plan-phase authoring (card t395), written against the card **as re-aimed by the lead** after its original premise was disproven. The dispatching card said "the SQLite migration failed to remove the original `backlog.json`". Measurement (`.moai/reports/t395/premise-verdict.md`) showed the original *was* quarantined correctly on 2026-08-27 and that the present file was **created separately on 2026-08-31**, by a writer the investigation could not identify. The SPEC therefore follows the measured damage (a non-authoritative file that answers every direct read silently, and repository instructions that tell readers to read it) rather than the dispatched cause (migration cleanup). |

## §A Context

### A.1 What is authoritative today

The queue `moai todo` operates on lives in one SQLite database,
`.moai/state/todo/backlog.db` (SPEC-TODO-SQLITE-001). `.moai/docs/todo-queue-storage.md:20`
states this correctly, including that a `backlog.json` at the same path is
present "only if you exported one, or if the queue has not been moved onto the
database yet".

The tool path is already safe. `internal/kanban/backlog_store.go:594-604`
resolves State D (database and JSON both present) in favour of the database, and
`internal/kanban/backlog_migrate.go:556-566` `quarantineLegacyBacklog` renames
rather than deletes, by design (REQ-TOSQ-014). Neither is a defect and neither is
touched by this SPEC.

### A.2 The measured damage

The damage is on the **reading** side, and it has two shapes.

**A machine reader that was pointed at the wrong file.**
`.claude/skills/moai-kanban-foreman/SKILL.md:95` arms a persistent Monitor whose
watch target is `.moai/state/todo/backlog.json`. On a migrated project it never
fires. This was **observed, not inferred** (`.moai/reports/t395/r1-repro.md`): the
shipped loop, run verbatim as a Monitor against an isolated fixture queue for 45
seconds spanning a real `moai todo add`, produced **zero events**, while a control
arm watching `backlog.db` in the same window on the same mutation fired. The
unattended foreman's queue watch is dead, and it is dead silently — the shape of
a vacuous green.

**Human and agent readers that were told the wrong location.** Three sites assert
that queue state lives in `backlog.json`:

- `.claude/skills/moai/SKILL.md:170` — "State: `.moai/state/todo/backlog.json` — project-local, not committed, atomic writes."
- `.claude/skills/moai/workflows/todo.md:17` — "State lives at `.moai/state/todo/backlog.json` of the PRIMARY checkout"
- `.claude/skills/moai/workflows/todo.md:21` — the home-fallback form, `~/.moai/todo/<project-key>/backlog.json`

All three are false after the SQLite cutover, and all three contradict
`.moai/docs/todo-queue-storage.md`. Which one a reader meets first decides
whether they read the queue or a snapshot. This is the instruction that produced
the lane-10 stale read.

**Four defect sites, three files, and a control that must not be edited.** The
four sites above live in three files — `workflows/todo.md` carries two of them.
Two further hits share the string but are **correct** and are the control:
`.moai/docs/todo-queue-storage.md:20` (the JSON is present only if exported or
not yet migrated) and `:55` (`export-json` writes the canonical path, which is
designed behaviour). Neither is edited by this SPEC.

All four defect sites are mirrored under `internal/template/templates/`, so
distributed users receive the same dead monitor and the same false assertion.
The attribution for that claim is **two** greps, not one, because the sites do
not share a path shape:

```
$ grep -rn 'state/todo/backlog\.json' internal/template/templates/
  .../.claude/skills/moai/SKILL.md:170 · .../workflows/todo.md:17 · .../moai-kanban-foreman/SKILL.md:95
  (plus .../.moai/docs/todo-queue-storage.md:55 — the export-json control)
$ grep -rn 'moai/todo/<project-key>/backlog\.json' internal/template/templates/
  .../.claude/skills/moai/workflows/todo.md:21
```

`reader-surfaces.md` R5 originally transcribed the first grep's four hits as
"four defect sites", which counted the `todo-queue-storage.md:55` control as a
defect and missed `todo.md:21` entirely. Its conclusion was true and its
attribution was not; the report now carries that correction, and the two greps
above are what this SPEC stands on. AC-BJD-015 runs both.

### A.3 Why disclosure is the only durable defence

The investigation **could not identify what wrote** the present `backlog.json`
(recorded Gap in `premise-verdict.md`): its record shape predates the `archived`
field, so the current binary's only writer to that path — `moai todo export-json`
(`internal/cli/todo_export.go:74`) — did not produce it. A one-time cleanup
cannot prevent recurrence by an unidentified writer. Making the reading side
able to tell that the file is not authoritative is defence that holds whoever the
writer turns out to be.

### A.4 The mechanism already exists

`internal/kanban/backlog_archive_vouch.go:46` `InspectBacklogArchiveVouch`
already resolves **which store answers reads** and returns it by name
(`BacklogStoreSQLite` / `BacklogStoreLegacyJSON` / `BacklogStoreNone`),
read-only on every branch — no migration, no DDL, no lock. It reads
`inspectBacklogLayout` (`backlog_migrate.go:392`), which already computes both
`dbExists` and `jsonExists`, so the State D fact this SPEC needs is **already
measured and then discarded**.

`internal/cli/todo_history.go:84,99` consumes it to name the answering store on
stderr, so that `absent` is not misread as authoritative — the same class of
defence this card needs, one surface over. This SPEC extends that mechanism. It
does not introduce a second one.

## §B Requirements

### B.1 Disclosure

- **REQ-BJD-001** (Ubiquitous) — The store-identity inspector shall report, as a
  fact about the queue layout, whether a `backlog.json` exists at the canonical
  queue path **while** the SQLite store is the one answering reads.

- **REQ-BJD-002** (Event-driven) — **When** a `moai todo` **read** command runs
  against a queue layout in which both `backlog.db` and `backlog.json` are
  present, the CLI shall emit one disclosure line naming the SQLite store as the
  store that answered and naming the `backlog.json` as not authoritative.

  **Breadth is an open operator decision, and the read surface is the floor.**
  Whether the disclosure also rides the write verbs was escalated to the operator
  and is unanswered. The read surface is normative now and implementable now, so
  nothing is blocked: a later answer **widens** this requirement rather than
  contradicting it, and a widening changes no criterion below. The floor is
  chosen rather than the ceiling because the write verbs already take the queue
  lock and carry a different stdout contract, so extending to them is a decision
  with consequences the read surface does not have. See `plan.md` §D.1.

- **REQ-BJD-003** (State-driven) — **While** the queue layout has no
  `backlog.json` beside the database, the CLI shall emit no disclosure line.

- **REQ-BJD-004** (Unwanted) — The disclosure shall not be written to stdout, so
  that a machine consuming a `moai todo` command's stdout is unaffected.

- **REQ-BJD-005** (Unwanted) — The inspector and the disclosure shall not delete,
  truncate, rename, move, or rewrite `backlog.json`, and shall not run a
  migration, DDL, or take the queue lock.

- **REQ-BJD-006** (Ubiquitous) — The disclosure shall be carried by the existing
  `InspectBacklogArchiveVouch` store-identity surface in
  `internal/kanban/backlog_archive_vouch.go`, extended in place; a second,
  parallel disclosure mechanism shall not be introduced.

### B.2 The foreman queue watch

- **REQ-BJD-007** (Ubiquitous) — The kanban foreman's queue watch shall observe
  the authoritative store, `.moai/state/todo/backlog.db`.

- **REQ-BJD-008** (Event-driven) — **When** the queue is mutated while the
  foreman queue watch is armed, the watch shall emit a change event.

- **REQ-BJD-009** (Unwanted) — The foreman queue watch shall not report the queue
  as unchanged on the basis of a watch target that is permanently absent or
  permanently static.

- **REQ-BJD-010** (State-driven) — **While** the SQLite store defers a committed
  write to `backlog.db-wal`, the queue watch shall still detect the mutation.

### B.3 The reader instructions

- **REQ-BJD-011** (Ubiquitous) — `.claude/skills/moai/SKILL.md` and
  `.claude/skills/moai/workflows/todo.md` shall name
  `.moai/state/todo/backlog.db` as the location of queue state.

- **REQ-BJD-012** (Unwanted) — Those files shall not assert that
  `.moai/state/todo/backlog.json`, or the home-fallback
  `~/.moai/todo/<project-key>/backlog.json`, is where queue state lives.

- **REQ-BJD-013** (Ubiquitous) — The corrected assertions shall be consistent
  with `.moai/docs/todo-queue-storage.md`, which is already correct and is not
  rewritten by this SPEC.

### B.4 The distributed copy

- **REQ-BJD-014** (Event-driven) — **When** any of the four `.claude/` sites named
  in §A.2 is edited, its `internal/template/templates/` mirror shall be edited to
  the same effect in the same change, and the binary shall be re-embedded with
  `make build`.

## §C Exclusions

### Out of Scope — removing or cleaning up `backlog.json`

- Deleting, truncating, or automatically cleaning the present `backlog.json`.
  The writer that produced it was not identified (§A.3), so a cleanup cannot
  prevent recurrence, and the file is also the deliberate target of the
  downgrade route (`moai todo export-json`).
- Changing `quarantineLegacyBacklog` or the State D resolution in
  `backlog_store.go`. Both were measured to behave as designed; neither is a
  defect.

### Out of Scope — storage and migration

- Changing the storage engine, the JSON-to-SQLite migration, or the downgrade
  route. `export-json` writing to the canonical path stays deliberate.
- Any change to the queue's SQLite schema. The items-table landed-evidence
  column axis belongs to card t359; this SPEC owns disk-file identity only.

### Out of Scope — unbounded reader surfaces

- Making the `backlog.json` file itself self-describing on disk (an embedded
  marker, a rename convention, a sentinel key). Disclosure here protects readers
  that go through the tool; a reader that runs `cat` on the file is not reachable
  by this SPEC and is not claimed to be.
- Auditing past reports, agent transcripts, or operator habits for prior stale
  reads.

### Out of Scope — foreman iteration beyond the watch

- The remaining steps of the `moai-kanban-foreman` iteration. Only the queue
  watch step is in scope; the rest of the skill body is untouched.
