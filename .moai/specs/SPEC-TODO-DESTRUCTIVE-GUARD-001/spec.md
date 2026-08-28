---
id: SPEC-TODO-DESTRUCTIVE-GUARD-001
title: "Reversibility for `moai todo done` — an additive archive, its restore verb, and a landing-predicate seam"
version: "0.2.2"
status: completed
created: 2026-08-27
updated: 2026-08-28
author: manager-spec (card t330)
priority: P2
phase: "v3.1.4 target"
module: "internal/kanban, internal/cli, .claude/skills/moai/workflows/todo.md, internal/template/templates/.claude/skills/moai/workflows/todo.md"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, cli, reversibility, archive, sqlite, additive-schema, guard"
tier: M
related_specs:
  - SPEC-KANBAN-TODO-CLI-001
  - SPEC-TODO-SQLITE-001
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
---

# SPEC: Reversibility for `moai todo done`

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-27 | Initial plan-phase authoring (card t330). Decision 1 ruled from the storage measurements; Decision 2 boundary settled against t331 after the existing landed primitive was measured against the motivating incident. |
| 0.2.2 | 2026-08-28 | §A.4 strengthened with one measured observation and one correction, both re-measured on this tree. Correction: the earliest of the 13 ref-corrected matches is the plan-phase artifacts commit `3030df58b`, not the run merge `3cb258d62` (tenth of 13) — the prior wording named the wrong commit as earliest. Strengthening: the grep-plus-nonempty shape (`prlink_landed.go:39-48`, `:80`) cannot discriminate commit kind, while the card-id-in-every-commit traceability discipline generates the matched signal unconditionally, so under the ref correction the predicate is structurally always-true for any card that has reached the integration branch. No REQ or AC added (16/16 unchanged); Decision 2's opt-in ruling unchanged and further justified. |
| 0.2.1 | 2026-08-28 | plan-audit iter2 debt closure (PASS-WITH-DEBT 0.9375). S1: the two `acceptance.md` citations the D7 sweep missed (`todo.go:341`→347, `352-354`→351-353) and the false "four refreshed" claim in progress.md. N1: AC-TDG-015 now captures stdout and stderr separately, with the reason recorded. N2: AC-TDG-007 asserts the exact `why` output. N3: `move`'s flag set corrected to four. N4: the "only point we control" claim softened so the in-artifact warning carrier stays open. Budget exhaustion recorded as plan.md §F.4. |
| 0.2.0 | 2026-08-28 | plan-audit iter1 delta (FAIL 0.75). §A.4 rewritten to state both failure modes of the landed predicate with their ref conditions (D1). Decision 3 added: the archive is deliberately included in `export-json` with a downgrade disclosure — REQ-TDG-015, and the D4 downgrade-loss gap closed with it (D2/D4). REQ-TDG-006, AC-TDG-006 and §A.3 rewritten to the reachable single-engine configuration (D3). Verb surface corrected to 15, re-derived from `AddCommand` (D5). `--expect` holder set corrected to `next`/`edit`/`drop`/`undrop` (D6). Four citations refreshed and the t306 commit count corrected 10→13 (D7). REQ-TDG-016 added: restore empties the archive entry (D8). Decision 1 unchanged. |

---

## §A Context

### A.1 The verb surface, and the asymmetry in it

`moai todo` exposes fifteen verbs. Read from the `AddCommand` call at `internal/cli/todo.go:137-141` — which is the authority, not the doctrine table and not any summary of it:

```
add  list  done  next  unpick  edit  move  drop  undrop
analyze  relate  unrelate  why  pr  export-json
```

Two destructive verbs already have exact inverses:

| Destructive verb | Inverse | Mechanism |
|---|---|---|
| `drop` | `undrop` | the card stays in the record; only `state` changes (`internal/cli/todo_drop.go`) |
| `next` (pick) | `unpick` | same — a state transition, nothing removed |
| `done` | **none** | the row is spliced out of `rec.Items` and cannot be recovered |

`done` is the only destructive verb with no way back.

### A.2 What `done` destroys

Measured at `internal/cli/todo.go:332` (`newTodoDoneCmd`), the mutation does **two** things under one lock:

1. splices the addressed item out of `rec.Items`;
2. calls `rec.RemoveFindingsNaming(id)` (`internal/cli/todo.go:347`), which drops **every** findings row naming that card (`internal/kanban/backlog_store.go:201`).

The second is easy to miss. A restore that recovered only the card row would silently lose every relation the operator or the analyser had recorded about it — `near-duplicate`, `contains`, `absorbs`, `replaces`, `conflicts`. Those relations are operator judgment, and losing them quietly is a worse failure than losing the card, because nothing signals the loss.

### A.3 The motivating incident

The kanban lead ran `done` on card t306 **before its sync commit had landed**. The run-phase landing (`3cb258d62`) is loud — a lane reports it, ancestry checks run, the integration branch moves — while the sync commit stays quiet in the lane's worktree. Reading the loud signal as the terminal one is a natural error.

It was not reversible. Re-issuing the card under a new id would split one card's history across two ids, which is worse than the original mistake: the commits, the evidence path, and the dispatch record all name t306.

### A.4 Why the obvious guard does not work — a measurement that refutes it

The natural response is "refuse `done` when the card's last step has not landed", reusing the existing landed primitive (`internal/kanban/prlink_landed.go`, `GitLandedQuerier.Landed`). It does not work — and it fails in **two different ways** depending on which ref the predicate reads. Both must be stated, because each alone gives a misleading picture.

Measured at `812ee01fc`:

```
$ git log origin/main    --perl-regexp --grep='\bt306\b' --oneline | wc -l
       0
$ git log origin/develop --perl-regexp --grep='\bt306\b' --oneline | wc -l
      13
```

**Mode 1 — as shipped.** `LandedRef = "origin/main"` (`internal/kanban/prlink_landed.go:28`), while the project's integration branch is now `develop`. `origin/main` names t306 in **zero** commits, so the predicate answers **false** — and would answer false for every develop-integrated card. A default-on refusal in this mode blocks `done` on essentially the entire queue. It would have "caught" the t306 incident only in the sense that a check refusing everything catches everything.

**Mode 2 — after the obvious ref correction.** Pointed at the real integration branch, the predicate answers **true**: `origin/develop` names t306 in 13 commits. The run-phase integration merge `3cb258d62` is the tenth of them and landed long before the sync commit — so in the mode a maintainer would naturally "fix" it into, the predicate is satisfied at the exact moment of the premature `done`, and a default-on refusal passes silently while reporting that a check ran and was satisfied.

**Mode 2 is satisfied earlier still — at plan-phase.** The same ref-corrected measurement puts the *earliest* of the 13 at `3030df58b`, `docs(SPEC-TODO-SQLITE-001): plan-phase artifacts … (t306)`, which precedes the run merge by about three hours. The predicate cannot tell the two apart: `LandedGrepArgs` (`internal/kanban/prlink_landed.go:39-48`) builds `git log <ref> --perl-regexp --grep='\b<id>\b' --oneline`, and `Landed` decides by `strings.TrimSpace(out) != ""` (`:80`) — a non-empty commit set, whatever kind of commit produced it. Meanwhile the signal it greps for is generated unconditionally by a *different* discipline: this project requires the card id in **every** commit message on the card's branch (`AGENTS.md` §3, `.claude/rules/moai/workflow/kanban-dispatch.md`), precisely because the `WT-<slug>` branch name no longer identifies the card. So the plan-phase commit alone satisfies the predicate, before a line of implementation exists. Under the ref correction, the check is structurally always-true for any card that has reached the integration branch at all — it is not merely permissive in this instance, it has no discriminating power to lose.

Neither mode is a guard. The primitive answers *"has anything naming this card landed"*, not *"has this card's last step landed"*, and only the first question is answerable from a commit-message grep. Mode 1 is uselessly strict, Mode 2 is dangerously permissive, and correcting the ref moves the failure from the first to the second rather than removing it.

The opt-in ruling in §B.2 therefore rests on **the predicate answering the wrong question**, not on which ref it reads. Correcting the ref does not unlock a default-on check.

Distinguishing the last step requires knowing which phase the card is in — the persisted landing-state field, which is **card t331's scope**, and t331 is recorded as containing t330.

> Any restatement of the Mode 2 form — "the run commit satisfies it", or its stronger version "even the plan commit satisfies it" — is incomplete without its ref-correction condition. As shipped, against `origin/main`, both are false.

### A.5 The safety net this SPEC actually delivers

Reversibility. `done` becomes an act the operator can take back, restoring the card **and its findings** to the bytes they had. The landing predicate becomes a seam with an explicit opt-in, honestly documented as to what it can and cannot answer, so that t331 can make it answer the right question without re-opening this design.

---

## §B Decisions

### B.1 Decision 1 — how `done` is reversed

Ruled: **an additive archive**. `done` moves the card row and its findings rows into archive storage; `undone` moves them back. Rejected: a fourth `done` value on `BacklogState` plus an `undone` verb.

The ruling rests on three measurements, not on symmetry with `drop`.

**M1 — the state CHECK cannot be altered in place.** `internal/kanban/backlog_sqlite.go:100`:

```sql
state TEXT NOT NULL CHECK (state IN ('queued','picked','dropped'))
```

The table is created with `CREATE TABLE IF NOT EXISTS`, so a database already present on an operator's machine **keeps the old CHECK**. SQLite cannot `ALTER` a CHECK constraint. Admitting a fourth state value requires a table-rebuild migration on every existing queue in the field.

**M2 — a new table is free, on every existing database.** `ensureSchema` (`internal/kanban/backlog_sqlite.go:232-235`) executes the whole `backlogDDL` on **every open**, and every statement in it is `IF NOT EXISTS`. A new table therefore lands automatically on every existing database with no migration code at all. This is the decisive asymmetry: additive *tables* are free; a changed *CHECK* is not.

**M3 — the record schema has a standing additive precedent, and a standing freeze.** `internal/kanban/backlog_store.go:44-47` and `145-157`:

- the schema is ADDITIVE within version 1;
- `last_seq` was appended as a top-level field with no version bump;
- `findings` followed the same precedent — "a second top-level field, no version bump, and the five-field per-item contract untouched";
- **no per-item field may ever be added** (REQ-TODO-013).

An archive expressed as top-level fields is the third application of a precedent this record already carries twice.

Two further consequences follow, and both are requirements rather than notes:

- **No `schema_version` bump.** `backlogSchemaVersion = "1"` (`backlog_sqlite.go:50`), and `ensureSchema` aborts with `ErrBacklogCorrupt` semantics on **any** version mismatch, including a *newer* stamp. Bumping to `"2"` would make an older binary refuse to open the queue at all. Holding at `"1"` keeps both directions working: a new binary creates the archive tables on an old database, and an old binary ignores them.
- **Zero leak surface.** A fourth state value would oblige an audit of every state-reading path — `ClassifyCardText` skips only `dropped` (`internal/kanban/backlog_analysis.go:139`), the state counts tally only `picked` and `queued` (`internal/kanban/backlog_migrate.go:252-258`), and every listing filter besides. Archived rows live in their own tables and their own top-level fields, and are invisible to all of those **by construction**. `done` also keeps its plain meaning: the card leaves the queue.

### B.2 Decision 2 — the confirmation, and the boundary against t331

Ruled: **t330 lands the reversal and the refusal seam; t331 lands the persisted landing-state field that a default-on judgment would read.**

Three constraints shape the confirmation's form:

**C1 — it must never prompt.** The CLI is invoked by agents, not only humans; a prompt would hang an unattended lane. The package carries this as a stated discipline — `internal/cli/todo.go:20` (`SUBAGENT BOUNDARY (C-HRA-008 / REQ-TODO-014): this command never prompts`) and `internal/cli/todo_pr.go:15`. This SPEC preserves it and says so.

**C2 — the guard convention already exists.** Four verbs take `--expect <prefix>`, refusing when the addressed card's text does not match and leaving the file byte-identical: `next` (`internal/cli/todo.go:441`), `edit` (`internal/cli/todo_edit_move.go:91`), `drop` and `undrop` (`internal/cli/todo_drop.go:133, 190`). `move` does **not** carry it — its four flags are `--top`, `--bottom`, `--before`, `--after` (`internal/cli/todo_edit_move.go:149-152`), none of them a text guard — so the convention covers every verb that changes a card's identity or lifecycle, and skips the one that changes only its position. `done` is the destructive verb that lacks it. Adding it is convention-following, not invention, and it directly addresses mis-addressing — the other half of the incident class.

**C3 — the landing predicate cannot yet answer the right question.** Per §A.4. Therefore the landing check ships **opt-in** (`--require-landed`), documented precisely as to its limit, rather than default-on. Shipping it default-on would satisfy the letter of the card while manufacturing the false confidence §A.4 identifies.

The boundary, stated so t331's author inherits an unambiguous contract:

| Owned by t330 | Owned by t331 |
|---|---|
| the archive storage and `undone` | the persisted landing-state field on the card |
| `--expect` guard on `done` | the phase-aware predicate that reads it |
| the landing-predicate **seam** + opt-in `--require-landed` | flipping the check to default-on, if warranted |
| documenting the predicate's limit | replacing the predicate with one that has no such limit |

`moai todo pr` is **not** modified. It is read-only by ruling (`internal/cli/todo_pr.go:1-15`) and stays so.

---

## §C Requirements

### C.1 Reversal

- **REQ-TDG-001** — The queue shall provide a verb that restores a card removed by `done`, returning the card row and every findings row that named it.
- **REQ-TDG-002** — When a card is restored, the queue record shall be byte-identical to its state immediately before the `done` that removed it.
- **REQ-TDG-003** — When `done` removes a card, the queue shall retain the card row and every findings row naming it in archive storage, rather than discarding them.

### C.2 Storage

- **REQ-TDG-004** — The archive shall be additive: new top-level record fields and new database tables, with the five-field per-item contract and the three-value `BacklogState` enum unchanged.
- **REQ-TDG-005** — The stored `schema_version` shall remain `"1"`, so that a binary predating the archive continues to open a database containing archived rows.
- **REQ-TDG-006** — The archive shall be expressed on `BacklogRecord`, so that it rides the existing `Mutate` seam and is carried by both the SQLite tables backing the live engine and the JSON *file format* used for export and downgrade artifacts.

  > There is exactly one live engine. `openEngine` (`internal/kanban/backlog_store.go:437-455`) migrates a JSON-only queue under the lock and then falls through to `openBacklogEngine(backlogSQLitePath(...))` on every path; `internal/cli/todo_export.go:3-11` states the swap is one-way and that no config knob selects an engine. So "both backends" names a *format* obligation, not two live engines: JSON must carry the archive because a downgrade artifact is JSON, not because a JSON engine ever serves it.

- **REQ-TDG-007** — Archived rows shall not appear in any listing, count, candidate set, or duplicate-analysis input that reports on the **live queue** — specifically `list`, `next`, `why`, `analyze`, and the state counts. This requirement does not constrain `export-json`, which is governed by REQ-TDG-015.

### C.3 Guards

- **REQ-TDG-008** — Where `--expect <prefix>` is supplied to `done`, the command shall refuse when the addressed card's text does not start with the prefix.
- **REQ-TDG-009** — Where `--require-landed` is supplied to `done`, the command shall refuse when the landing predicate reports the card as not landed, and shall proceed when the predicate is inconclusive.
- **REQ-TDG-010** — While `--require-landed` is absent, `done` shall run no landing query and incur no subprocess cost beyond its existing budget.
- **REQ-TDG-011** — When any guard refuses a command, the queue record shall be left byte-identical.

### C.4 Boundaries

- **REQ-TDG-012** — The `done` and `undone` commands shall not prompt, on any path, including refusal paths.
- **REQ-TDG-013** — The queue shall not restore a card whose id has since been reissued to a different card; such a restore shall be refused with the collision named.
- **REQ-TDG-014** — The new verb and the new flags shall be documented in the todo workflow doctrine, in both the working copy and the template source.

### C.5 Downgrade and archive lifecycle

- **REQ-TDG-015** — The archive shall be **included** in `export-json` output, and where the exported record carries archived rows the command shall disclose on stderr that a binary predating the archive will discard them on its first write.
- **REQ-TDG-016** — When a card is restored, its archive entry shall be **removed** rather than retained, so that each row has exactly one home and no card is simultaneously live and archived.

Requirement count: 16 (Tier M ceiling: 16).

#### Why the export includes the archive (Decision 3)

`export-json` is *the deliberate downgrade route* (`internal/cli/todo_export.go:1-11`): it writes the live queue as a legacy-format `backlog.json`, and "that older binary reads only the json filename", so the exported file is exactly what the previous release serves. Two consequences follow, and they point the same way:

- **Inclusion is the correct default, not an accident of `json.MarshalIndent(rec, …)` marshalling the whole record** (`internal/cli/todo_export.go:69`). A downgrade that silently dropped every archived card would destroy exactly the rows this SPEC exists to preserve. Excluding them would also cost a custom marshaller — real work to make the product worse.
- **The residual loss cannot be prevented, only disclosed.** An older binary unmarshals the export into a struct with no archive field, drops the unknown key, and re-serializes without it on its first write. Nothing in this SPEC runs inside that binary. What *is* in reach is the export itself — so REQ-TDG-015 converts a silent loss into a loud one, which is the standard §A.2 asks for.

  > The stderr disclosure is the cheapest carrier, not the only one. The **artifact** is also ours: a top-level warning string written beside the archive would ride the same marshal for free and outlive terminal scrollback for anyone who opens `backlog.json` before downgrading. Not built here — REQ-TDG-015 is satisfied by the stderr line — but a run-phase implementer who prefers the in-artifact form, or both, is not foreclosed.

REQ-TDG-016 is forced by REQ-TDG-002: if a restore left the archive entry in place, the record would not return to its pre-`done` bytes, and the card could be restored twice.

---

## §D Exclusions

This SPEC deliberately does not build the following. Each is out of scope for a stated reason, not by oversight.

### Out of Scope — the already-paired verbs

- `drop` / `undrop` — already an exact inverse pair with a round-trip test; untouched.
- `next` (pick) / `unpick` — already paired; untouched.
- Any change to the `BacklogState` enum or its CHECK constraint (§B.1 M1).

### Out of Scope — the persisted landing state

- The persisted landing-state field on the card. That is card t331, which is recorded as containing t330.
- Any phase-aware landing predicate — one that distinguishes a run landing from a sync landing (§A.4). t330 ships the seam and an honestly-limited default.
- Flipping `--require-landed` to default-on. That decision belongs with the predicate that makes it defensible.

### Out of Scope — authority and the read path

- Any change to **who** may issue a destructive verb. The [HARD] doctrine stands unchanged: the drop/done decision is the **operator's**, and agents never drop or close cards on their own initiative. This SPEC makes a destructive act recoverable; it does not widen who may take it.
- Any modification to `moai todo pr`. It is read-only by ruling and stays read-only.
- Any default-on network cost on `todo list`, `todo next`, or any other read path (`internal/cli/todo_pr.go:8-13`).

### Out of Scope — storage rework

- Any `schema_version` bump, table rebuild, or migration of existing operator queues (§B.1 M2, REQ-TDG-005).
- Archive pruning, retention limits, or compaction. The archive grows without bound in this SPEC; a retention policy is a separate decision requiring operator input.
- Any new per-item field (frozen by REQ-TODO-013).
- Any attempt to make a pre-archive binary preserve the archive across a downgrade. That code is not ours to change; REQ-TDG-015 discloses the loss instead.
- Correcting `LandedRef` (`internal/kanban/prlink_landed.go:28`) from `origin/main` to the current integration branch. The constant is shared with `moai todo pr`, so changing it alters that verb's answers — a separate decision with its own blast radius. §A.4 shows the correction would not unlock a default-on check here, so this SPEC does not depend on it.

---

## §E Traceability

| Requirement | Acceptance criteria |
|---|---|
| REQ-TDG-001, 002 | AC-TDG-001, AC-TDG-002 |
| REQ-TDG-003 | AC-TDG-003 |
| REQ-TDG-004, 005 | AC-TDG-004, AC-TDG-005 |
| REQ-TDG-006 | AC-TDG-006 |
| REQ-TDG-007 | AC-TDG-007 |
| REQ-TDG-008 | AC-TDG-008 |
| REQ-TDG-009, 010 | AC-TDG-009, AC-TDG-010 |
| REQ-TDG-011 | AC-TDG-011 |
| REQ-TDG-012 | AC-TDG-012 |
| REQ-TDG-013 | AC-TDG-013 |
| REQ-TDG-014 | AC-TDG-014 |
| REQ-TDG-015 | AC-TDG-015 |
| REQ-TDG-016 | AC-TDG-016 |

Full criteria: `acceptance.md`.
