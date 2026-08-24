---
id: SPEC-KANBAN-RECORD-SESSION-KEY-001
title: "Key the kanban record by the session it describes, and record that session's lane and card"
version: "0.1.0"
status: draft
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
| 0.1.0 | 2026-08-24 | Initial draft. Split from SPEC-WEB-CONSOLE-015 per the ratified split design (`.moai/reports/t207/spec-split-design.md` Appendix 2, decision D-6), which measured that the record is keyed by the launching session's identifier and therefore that the parent's lane join does not close. |

## §A Background

Ground truth measured in worktree `.claude/worktrees/t207` at `dfbf828a6`. Every claim names the
command that produced it; nothing below is carried over from the split design's citations.

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
the launcher. The record is therefore filed under the **parent's** identifier, always, and the
role written into it is the **child's** role.

### A.2 The consequence, measured on this machine today

| Claim | Evidence |
|---|---|
| Three sessions are live and registered | `.moai/state/active-sessions.json` holds `2beac221…` (pid 15207, cwd `…/worktrees/t219`), `c15d8434…` (pid 51045, `…/t210`), `3db058e1…` (pid 36912, `…/t207`) |
| None of the three has a record of its own | `ls .moai/state/kanban/{2beac221…,c15d8434…,3db058e1…}.json` → `No such file or directory` ×3 |
| A record exists under a fourth identifier, and it describes a different session than it names | `cat .moai/state/kanban/d281730e-a47e-4f82-878e-5fd0ddc4dcb9.json` → `{"session_id":"d281730e-…","spec_id":"","role":"lane","backend":"claude","entered_at":"2026-08-23T17:47:22Z","deepscan_dir":"","verify_reentries":0}`. `d281730e…` is a **lead** session; the record's `role` reads `lane` |
| The single slot presently names this session | `cat .moai/state/current-session-id.txt` → `3db058e1-…`, whose SessionStart wrote it last |

One mistake breaks both ends at once: the lane has no record, and the lead's identifier carries a
lane's record. There is no reading of this data under which a consumer recovers the truth, because
the file name and the file's contents describe different sessions and nothing on disk says so.

### A.3 What this blocks

`SPEC-WEB-CONSOLE-015` §A.5 asserts that
`workers.json[lane-N].PID → active-sessions entry → session_id → kanban.Record` "is a join that
closes on today's data with no new state file". The first two hops hold — `FactoryWorkerEntry.PID`
(`internal/kanban/factory_slots.go:38`) and `Entry.PID` (`internal/session/registry.go:92`) both
exist and the registry carries the runtime identifier. The **third hop is where it breaks**: the
registry's identifier is the session's own, the record's is its parent's, so the lookup returns
either nothing or another session's record. Every view built on that join renders misattributed
rows or no rows.

### A.4 Two things the record cannot say, and one it says wrongly

| Missing | Evidence |
|---|---|
| **The lane number.** `internal/kanban/role.go:42` defines `RoleLane = "lane"` as a bare constant, so every factory lane writes the same value and N lanes are mutually indistinguishable. Measured: `grep -h '"role"' .moai/state/kanban/*.json \| sort \| uniq -c` over the 84 record files returns **34** `"role": "lane"` rows and no lane number anywhere — `grep -l '"lane"\s*:' .moai/state/kanban/*.json` returns **0** files |
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
| **Backend** | **No** | `grep -rn "BackendGLM\|BackendClaude" internal/ \| grep -v _test` shows the value exists only as a literal argument at the eight `recordKanbanSession` call sites (`cc.go:161,175,192,208`, `glm.go:224,237,250,264`). No environment key names it: `grep -rn 'BACKEND' internal/config/envkeys.go` returns nothing |
| **Card identifier** | **No** | No environment key exists for it (same grep), and no per-lane card identifier exists on disk anywhere |

So role and lane number are already there; **backend, SPEC identifier for companions and lanes, and
the card-identifier override are not**, and carrying them is part of this SPEC rather than an
assumption it rests on (REQ-KRS-006). The card identifier itself is otherwise derivable inside the
session — the dispatch protocol fixes a card's worktree at `.claude/worktrees/<card-id>`, so the
basename of `git rev-parse --show-toplevel` is the card identifier wherever a card-carrying session
stands. Measured here: that command returns
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`, basename `t207`, which is this card.

### A.6 Why the schema may only grow

`internal/kanban/record.go:45` carries an `@MX:ANCHOR` whose stated reason is that the launcher,
the orchestrator, and the sync-phase dedup gate all bind to these JSON keys, so a renamed key
breaks readers the package cannot see. Every field this SPEC adds is therefore additive and
`omitempty`, and the 84 existing record files stay readable.

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

- **REQ-KRS-004** — **Where** a session is a factory lane, its record shall carry the lane's number
  as data distinct from the role value, as a non-pointer integer whose zero value means "not a
  lane".
- **REQ-KRS-005** — A record shall carry the card identifier the session is working, as a field
  distinct from the SPEC identifier and populated for cards that never acquire a SPEC. The value
  shall be taken from an explicit override where one is supplied, and otherwise from the basename
  of the session's worktree root. **When** neither source yields a value, the field shall be left
  empty rather than guessed.
- **REQ-KRS-006** — Every launch fact a record carries that the launched session cannot observe for
  itself — the backend it runs on, the SPEC the launch targets, and the card-identifier override
  where one is supplied — shall be conveyed to that session through its launch environment.

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

Six to eight files across three packages, no always-loaded doctrine, no published documentation:
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

The 84 files under `.moai/state/kanban/` are left exactly as they are. They are neither migrated
nor deleted: the directory is gitignored runtime state, each file is a few hundred bytes, and the
misattributed ones cannot be repaired — the parent identifier they are named for does not encode
which child they described, so a migration would have to guess. They age out as their sessions do,
and REQ-KRS-007 keeps them readable in the meantime. Introducing a reaper is a separate change with
its own liveness question.

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

- Migrating, repairing, or deleting the 84 files already under `.moai/state/kanban/`. The
  disposition chosen is to leave them (§C.4); a reaper or a migration is a separate change.

### Out of Scope — the role vocabulary

- Adding, removing, or renaming a role value, and widening the role setter to admit arbitrary
  launch-label text. The lane number is new data beside the role, not a new role.
