---
id: SPEC-KANBAN-RECORD-SESSION-KEY-001
title: "Key the kanban record by the session it describes, and record that session's lane and card"
version: "0.2.0"
status: in-progress
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: kanban, session, record, factory, lane
era: V3R6
tier: M
related_specs: [SPEC-WEB-CONSOLE-015, SPEC-SESSION-TELEMETRY-001]
---

# SPEC-KANBAN-RECORD-SESSION-KEY-001 — The kanban record's session key

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.2.0 | 2026-08-24 | Iteration-2 revision closing the plan-phase audit FAIL at 0.750 (`.moai/reports/t207/plan-audit-kanban-record-key-iter1.md`). Every measurement re-taken at `3c3a6fbf8`; the record directory is re-attributed to the project root, mutable file counts are replaced by the property they stood for, the card-identifier derivation gains a worktree-containment test so its anti-guess clause is reachable, sibling citations are retargeted to REQ ids, and the backend claim is scoped to the command that supports it. |
| 0.1.0 | 2026-08-24 | Initial draft. Split from SPEC-WEB-CONSOLE-015 per the ratified split design (`.moai/reports/t207/spec-split-design.md` Appendix 2, decision D-6), which measured that the record is keyed by the launching session's identifier and therefore that the parent's lane join does not close. |

## §A Background

Ground truth re-measured in worktree `.claude/worktrees/t207` at `3c3a6fbf8` (version 0.1.0 measured
at `dfbf828a6`; every citation below was re-run there, not carried over). Two scoping facts version
0.1.0 left implicit:

- **Source paths** are relative to the worktree root, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`.
- **Runtime state paths are not.** `.moai/state/` belongs to the **project root**: this worktree
  carries no `.moai/state/kanban/` at all (`ls .moai/state/` returns `.gitkeep`,
  `config-cache.json`, `context-usage.json` and nothing else). Every `.moai/state/…` path below is
  therefore read under `/Users/goos/MoAI/moai-adk-go/`.

### A.1 The record is keyed by the wrong session — the chain

`kanban.Record` is persisted at `.moai/state/kanban/<session-id>.json`
(`internal/kanban/record.go` `RecordPath`), and the identifier that names the file is resolved
like this:

| Step | Evidence |
|---|---|
| The launcher writes the record | `internal/cli/kanban.go:472` `func recordKanbanSession(specID, backend, role string)`, `:478` `kanban.WriteBestEffort(projectRoot, kanban.NewRecord(sessionID, specID, backend).WithRole(role))` |
| …keyed by whatever the launch-time resolver returns | `internal/cli/kanban.go:474` `sessionID := resolveLaunchSessionID("")` |
| …which, with no override, reads a single project-wide file | `internal/cli/launcher_blockcap_infinite.go:126` `func resolveLaunchSessionID`, `:130` `if id, _, ok := resolveCurrentSessionID(); ok` |
| …that file being one slot per project root | `internal/session/registry.go:52` `const CurrentSideChannelFile = ".moai/state/current-session-id.txt"` |
| …rewritten by every session's SessionStart | `internal/hook/session_start.go:314` writes `input.SessionID` to that path; its own doc at `registry.go:45` says the file "is overwritten on every SessionStart" |

Two independent faults compose. The slot is single, so the last session to start wins it. And the
launcher runs **before the session it launches exists**, so at that instant the slot cannot hold
the child's identifier under any circumstances — it holds the identifier of the session that ran
the launcher. The record is therefore filed under **whichever session wrote that slot last** — in
practice the launching session, and never the launched one by design. It is not *always* the
parent's identifier: measured on this machine, two of four live sessions carry a record under their
own identifier, because the slot happened to hold it when some launcher ran. That is coincidence,
not correctness, and it is the worse case rather than the better one — a consumer that finds a
record cannot tell whether it describes the session it asked about. (One measured instance: lane-5's
session carries a record whose `entered_at` is lane-10's registration instant, and lane-10's session
carries none.) The key is not a function of the session the record describes, and the
role written into it is the **child's** role.

### A.2 The consequence, measured on this machine today

Re-measured 2026-08-24 under `/Users/goos/MoAI/moai-adk-go/`. Version 0.1.0 cited three session
identifiers that no longer exist — those sessions ended — so the rows below are a fresh instance of
the same shape rather than the same rows.

| Claim | Evidence |
|---|---|
| Two sessions are live and registered | `.moai/state/active-sessions.json` holds `5d3be9b8-be19-42ab-8be1-7cb40b29c456` (cwd `…/.claude/worktrees/t210`) and `e46fcfef-1f5c-4f9c-beff-2ada72e26eb5` (cwd `/Users/goos/moai/moai-adk-go`) |
| Neither has a record of its own | `ls .moai/state/kanban/5d3be9b8-….json .moai/state/kanban/e46fcfef-….json` → `No such file or directory` ×2 |
| A record exists under an identifier the registry does not carry, and it names a role its session did not have | `cat .moai/state/kanban/d281730e-a47e-4f82-878e-5fd0ddc4dcb9.json` → `{"session_id":"d281730e-…","spec_id":"","role":"lane","backend":"claude","entered_at":"2026-08-23T17:47:22Z","deepscan_dir":"","verify_reentries":0}`. the record's `role` reads `lane` while the registry carries no entry for that identifier at all. That the session was a **lead** is a side observation, attributed to its own statusline payload (`session_name: "lead"`, captured 2026-08-24) rather than to any file in the tree, and it is not re-derivable from a command here — nothing in this SPEC rests on it. The load-bearing half is the one that is: a record exists under an identifier no registered session carries |
| The single slot presently names the live session that has no record | `cat .moai/state/current-session-id.txt` → `e46fcfef-…`, whose SessionStart wrote it last. Every record a launcher writes between now and the next SessionStart is filed under `e46fcfef…`, whatever session it launches |

One mistake breaks both ends at once: the lane has no record, and the lead's identifier carries a
lane's record. There is no reading of this data under which a consumer recovers the truth, because
the file name and the file's contents describe different sessions and nothing on disk says so.

The identifiers above will age out as version 0.1.0's did — and the population they describe drifts
faster than that. A later measurement on the same machine found four live sessions of which **two
did** carry a record under their own identifier, so the strongest phrasing this evidence supports is
NOT "a registered session has no record of its own". Stating it that way would make the SPEC
falsifiable by one lucky session, and it was: lane-5's session had a record, whose `entered_at`
(`09:25:29Z`) is lane-10's registration instant rather than its own (`09:22:12Z`).

What does not drift is the property this SPEC is written against: **a record's key is not a function
of the session the record describes.** It is a function of which session's SessionStart last wrote
the single slot, which sometimes coincides with the launched session and usually does not. The
coincidence is the dangerous case, not the reassuring one — a lookup that returns a record gives a
consumer no way to tell whether the record is about the session it asked for.

### A.3 What this blocks

`SPEC-WEB-CONSOLE-015` **REQ-WC15-043** resolves each registered factory lane to a session by
joining the factory registry's recorded process identifier to the active-sessions entry carrying it,
and thence to that session's record; **REQ-WC15-044** presents the lane number and card identifier
that join is supposed to deliver. Version 0.1.0 of that SPEC **asserted** the chain
`workers.json[lane-N].PID → active-sessions entry → session_id → kanban.Record` "is a join that
closes on today's data with no new state file". That assertion has since been **withdrawn there**,
and this section is the measurement the withdrawal rests on. (Sibling citations in this SPEC name
REQ ids rather than section numbers: the parent was rewritten and its sections moved; its REQ ids
did not.)

Both PID fields exist — `FactoryWorkerEntry.PID` (`internal/kanban/factory_slots.go:38`) and
`Entry.PID` (`internal/session/registry.go:92`) — and the registry carries the runtime identifier.
That is what is measured here, and it is less than saying the first two hops *hold*: a process
identifier is not unique in the registry, because `Registry.Register`
(`internal/session/registry.go:166-199`) deduplicates by session id alone, so two entries can carry
one live PID. That collision is the registry-side instance of the same shape
`SPEC-WEB-CONSOLE-015` **REQ-WC15-047** already covers on the `workers.json` side, and it belongs
to that consumer rather than to this SPEC: nothing here changes how sessions register.

The **third hop is the one this SPEC repairs**, and it is where the join breaks outright: the
registry's identifier is the session's own, the record's is its parent's, so the lookup returns
either nothing or another session's record. Every view built on that join renders misattributed
rows or no rows.

### A.4 Two things the record cannot say, and one it says wrongly

| Missing | Evidence |
|---|---|
| **The lane number.** `internal/kanban/role.go:42` defines `RoleLane = "lane"` as a bare constant, so every factory lane writes the same value and N lanes are mutually indistinguishable. The property, which does not move: **no record file carries a lane number, and every lane writes the same role value.** Measured 2026-08-24 under `/Users/goos/MoAI/moai-adk-go/.moai/state/kanban/`, whose file set grows while this SPEC is open — `grep -h '"role"' *.json \| sort \| uniq -c` returns **34** `"role": "lane"` rows, and `grep -l '"lane"[[:space:]]*:' *.json` returns **0** files |
| **The card identifier.** `Record.SpecID` (`record.go:61`) is a SPEC identifier; Class A and Class B cards never acquire one. `MOAI_KANBAN_ID` is not a substitute — `internal/config/envkeys.go:167-173` documents it as the **run** identifier, generated once per run at `internal/cli/kanban.go:173` and `internal/cli/factory.go:255` |
| **The role, when the label is a lane label.** `WithRole` (`record.go:116-130`) admits only `RoleLead`, `RoleLane`, and the companion roles, and silently drops anything else — so a `lane-3` value passed through it would vanish without an error |

### A.5 Which launch facts the launched session can already see

The record must be written by the session it describes, because that is the first actor that holds
the child's identifier. So the question is which launch facts reach that session. The launcher sets
its variables with `os.Setenv` before exec'ing the backend, and the SessionStart hook — a
descendant of that process — already reads them:

| Fact | Reachable today? | Evidence |
|---|---|---|
| The session's own identifier | **Yes** | `internal/hook/session_start.go:302` gates on `input.SessionID` |
| Role — kanban lead | **Yes** | `MOAI_KANBAN` set at `internal/cli/kanban.go:171`, read at `internal/hook/session_start_kanban.go:53` |
| Role — kanban companion, with its label | **Yes** | `MOAI_KANBAN_LABEL` set at `internal/cli/kanban.go:315`, read at `session_start_kanban.go:50` |
| Role — factory lead | **Yes** | `MOAI_FACTORY_WORKERS` set at `internal/cli/factory.go:253`, read at `internal/hook/session_start_factory.go:51` |
| Role — factory lane, **and its lane number** | **Yes** | `MOAI_FACTORY_WORKER` carries the `lane-<n>` label; set at `internal/cli/factory.go:289`, read at `session_start_factory.go:48` |
| SPEC identifier | **Partly** | `MOAI_KANBAN_SPEC` is set only inside `enterKanbanMode` (`kanban.go:174-176`, the kanban lead). `enterKanbanCompanionMode` (`kanban.go:305-322`) and `enterFactoryWorkerMode` (`factory.go:283-298`) set no SPEC variable |
| **Backend** | **No** | The record's backend reaches the write only as a literal argument at the eight `recordKanbanSession` call sites: `grep -rn 'BackendGLM\|BackendClaude' internal/cli/cc.go internal/cli/glm.go` returns `cc.go:161,175,192,208` and `glm.go:224,237,250,264` — eight lines, every one of them that argument. (A tree-wide grep additionally returns the constant declarations and their doc comment at `record.go:23,24,75`, plus nine hits on an unrelated same-named constant pair in `internal/cli/mcp_convergence.go`. Neither set is the record's backend, which is why the claim is scoped to the two launcher files rather than to the tree.) No environment key names it: `grep -rn 'BACKEND' internal/config/envkeys.go` returns nothing, `rc=1` |
| **Card identifier** | **No** | No environment key exists for it (same grep), and no per-lane card identifier exists on disk anywhere |

So role and lane number are already there; **backend, SPEC identifier for companions and lanes, and
the card-identifier override are not**, and carrying them is part of this SPEC rather than an
assumption it rests on (REQ-KRS-006).

The card identifier is otherwise derivable inside the session, but **only where the session stands
in a card worktree**, and that condition is load-bearing rather than incidental. The dispatch
protocol fixes a card's worktree directory at `.claude/worktrees/<card-id>`, so the basename of
`git rev-parse --show-toplevel` is the card identifier **when that root's parent directory is named
`worktrees`**, and is not a card identifier otherwise. Both cases are live right now:

| Session's worktree root | Parent directory | Basename | A card identifier? |
|---|---|---|---|
| `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207` — what `git rev-parse --show-toplevel` returns in this tree | `…/.claude/worktrees` | `t207` | **Yes**, this card |
| `/Users/goos/moai/moai-adk-go` — session `e46fcfef…`, its `cwd` in `active-sessions.json` | `/Users/goos/moai` | `moai-adk-go` | **No**, a checkout name |

An unconditional basename would file `moai-adk-go` as this second session's card identifier, and the
consumer — `SPEC-WEB-CONSOLE-015` REQ-WC15-044, which asks for a card identifier — would render it
as one. The containment test is what separates the two cases, and REQ-KRS-005 carries it.

### A.6 Why the schema may only grow

`internal/kanban/record.go:45` carries an `@MX:ANCHOR` whose stated reason is that the launcher,
the orchestrator, and the sync-phase dedup gate all bind to these JSON keys, so a renamed key
breaks readers the package cannot see. Every field this SPEC adds is therefore additive and
`omitempty`, and **every record file already on disk stays readable** — stated as that property
rather than as a file count, because the directory is live state that grows while this SPEC is open
(79 files carried a `session_id` when this was measured on 2026-08-24; the count moves, the
obligation does not).

## §B Requirements (GEARS)

### B.1 The key and the writer

- **REQ-KRS-001** — A session's kanban record shall be keyed by, and shall carry, the identifier
  that session's own runtime delivered to it, and shall not be keyed by any project-wide
  single-slot identifier file.
- **REQ-KRS-002** — The record describing a session shall be written from within that session, and
  the process that launches a session shall not write a record on the launched session's behalf.
- **REQ-KRS-003** — The role carried by a record shall be the role of the session the record
  describes.

### B.2 Lane and card identity

- **REQ-KRS-004** — **While** a session is a factory lane, its record shall carry the lane's number
  as data distinct from the role value, as a non-pointer integer whose zero value means "not a
  lane".
- **REQ-KRS-005** — A record shall carry the card identifier the session is working, as a field
  distinct from the SPEC identifier and populated for cards that never acquire a SPEC. The value
  shall be taken from an explicit override where one is supplied, and otherwise from the basename
  of the session's worktree root **only where that root's parent directory is named `worktrees`**.
  **When** neither source yields a value — an absent override together with a worktree root that
  fails that containment test, or no resolvable worktree root at all — the field shall be left empty
  rather than guessed.
- **REQ-KRS-006** — Every launch fact a record carries that the launched session cannot observe for
  itself — the backend it runs on, the SPEC the launch targets, and the card-identifier override
  where one is supplied — shall be observable by that session at its own start. The means is
  **deliberately constrained** rather than left to the implementation: the fact shall be conveyed by
  the launcher through the launch environment, and shall not be inferred from any other signal the
  session can see. (The rejected inference, and why it is a guess rather than a measurement, is
  recorded in `plan.md` §F.)

### B.3 Compatibility and failure

- **REQ-KRS-007** — Every field added shall be additive and shall be omitted from the encoded
  record when empty, and no existing key shall be renamed or removed. **When** a record written by
  a build predating this schema is read, every reader shall present the absent fields as not
  recorded and shall not fail, and rewriting such a record shall not alter its pre-existing keys.
- **REQ-KRS-008** — **When** a launch fact the record would carry is unavailable, or the record
  cannot be persisted at all, the session shall record what it has and shall start normally; the
  absence of a record shall not fail, delay, or alter the session.

## §C Constraints

### C.1 Blast radius

| Surface | Change |
|---|---|
| `internal/kanban/record.go` | two additive fields (lane number, card identifier); the constructor and the role setter grow to carry them |
| `internal/cli/kanban.go:472-479` | `recordKanbanSession` stops writing a record and becomes the place the launch facts are exported instead |
| `internal/cli/cc.go:161,175,192,208` · `internal/cli/glm.go:224,237,250,264` | the eight call sites, which today pass a backend literal that must now travel in the environment |
| `internal/config/envkeys.go` | the new launch-fact keys (REQ-KRS-006) |
| `internal/hook/session_start.go` (+ its kanban/factory siblings) | the record write, keyed by `input.SessionID` |
| package tests in `internal/kanban`, `internal/cli`, `internal/hook` | fixtures asserting the launcher-written record move to the session-written one |

Six to nine files across four packages, no always-loaded doctrine, no published documentation:
Tier M.

### C.2 Fail-open is preserved verbatim

`WriteBestEffort` (`record.go`) discards every failure by design, and its `@MX:NOTE` states that
the absent error return **is** the guarantee rather than an oversight: a launch path that cannot
observe a failure cannot block on one. Moving the write into the session preserves that property
rather than inheriting it by accident — REQ-KRS-008 is its restatement at the new site, because a
hook that fails loudly is a session that fails to start.

### C.3 Template-First does not apply

Go source under `internal/kanban`, `internal/cli`, `internal/config`, and `internal/hook` has no
mirror under `internal/template/templates/`, so no mirror pair is touched. Hook **wrapper** scripts
are a separate surface and are not changed by this SPEC.

### C.4 The existing record files

Every file already under `.moai/state/kanban/` is left exactly as it is — neither migrated,
repaired, nor deleted. The commitment is stated as that property and not as a count, because the
directory is live runtime state that grows while this SPEC is open: 79 record files carried a
`session_id` when this was measured on 2026-08-24, and one of them had been created that same
afternoon. A count quoted here would be a different number by the time anyone checked it, and a
reviewer could not tell drift from a violation.

The misattributed files cannot be repaired in any case: the parent identifier a file is named for
does not encode which child it described, so a migration would have to guess. They are gitignored
runtime state, each a few hundred bytes; they age out as their sessions do, and REQ-KRS-007 keeps
them readable in the meantime. Introducing a reaper is a separate change with its own liveness
question.

One measurement note, carried into `acceptance.md`: `.moai/state/kanban/*.json` is **not** a record
glob. `backlog.json` and `leads.json` sit in the same directory and carry no `session_id`
(`grep -L '"session_id"' .moai/state/kanban/*.json` returns exactly those two), so any evidence
about records excludes them.

## §D Exclusions

Explicitly out of scope. Each may be taken up separately.

### Out of Scope — the single-slot session-id sidecar

- Repairing `.moai/state/current-session-id.txt` itself, which is one slot per project root and is
  overwritten by every session's SessionStart. That surface is card **t221** and is deliberately
  not merged here: t221 is about the slot; this SPEC is about `kanban.Record` being keyed *from*
  it. This SPEC stops reading the slot for this purpose and fixes nothing about the slot — every
  other consumer of it (`moai session current`, handoff `source_session_id` attribution) is
  unaffected and remains as exposed as it is today.

### Out of Scope — anything the console renders

- The telemetry cells, the factory lane section, and every other view built on the corrected join.
  Owned by `SPEC-WEB-CONSOLE-015`, which depends on this SPEC.

### Out of Scope — recorded stage transitions

- Replacing heartbeat stage estimation with lead-recorded transitions. This SPEC records identity,
  not progress.

### Out of Scope — session telemetry values

- Model, effort, and context-window usage. Those ride the per-session statusline snapshot and are
  owned by `SPEC-SESSION-TELEMETRY-001`; this SPEC neither reads nor writes that record.

### Out of Scope — migrating the existing record files

- Migrating, repairing, or deleting any file already under `.moai/state/kanban/`. The disposition
  chosen is to leave every one of them untouched (§C.4); a reaper or a migration is a separate
  change.

### Out of Scope — the role vocabulary

- Adding, removing, or renaming a role value, and widening the role setter to admit arbitrary
  launch-label text. The lane number is new data beside the role, not a new role.
