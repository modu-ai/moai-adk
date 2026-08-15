---
id: SPEC-KANBAN-TODO-CLI-001
title: "moai todo CLI subcommand with lock-guarded backlog store"
version: "0.1.1"
status: in-progress
created: 2026-08-14
updated: 2026-08-14
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: kanban, cli, backlog, concurrency
tier: M
---

# SPEC-KANBAN-TODO-CLI-001 — moai todo CLI subcommand with lock-guarded backlog store

## §A Background

`/moai todo` is the operator's entry mechanism into the kanban board (`.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act). Today no `moai todo` CLI subcommand exists (verified: `moai todo` → `Unknown command "todo"`; zero `todo*.go` files in `internal/cli/`). The skill document (`.claude/skills/moai/workflows/todo.md`) therefore instructs the model to Read and Write `.moai/state/kanban/backlog.json` directly.

With five concurrent sessions on one checkout, the gap between a session's Read and its Write is measured in minutes. Every concurrent writer resolves to last-writer-wins. On 2026-08-14 this produced a real incident: 3 cards lost and an ID collision cluster (t4 absent from the queue; visible non-monotonic id sequence t1, t2, t10, t5, t8, t6, t7, t11, t12, t13–t19). The skill's "temp file + rename" instruction protects only against a crash mid-write; it does nothing about the read-modify-write race.

The fix is not a new design: the repository already owns the pattern. `internal/kanban` (SPEC-KANBAN-BOARD-001) carries a path-parameterized cross-process lock substrate (flock on Unix, atomic-create on Windows) and a guarded write path (lock → load → mutate → atomic replace); `internal/session/registry.go` `Registry.withLock` is the same shape. This SPEC applies that substrate to the backlog queue and replaces the skill's model-driven Read→Write cycle with a real `moai todo` CLI.

## §B Goals

- Eliminate the backlog read-modify-write race across concurrent sessions.
- Issue item ids under the lock so concurrent adds cannot collide and removed ids are never reused.
- Keep the on-disk backlog record load-compatible with the existing version-1 file.
- Reduce the todo skill to a "run the command" form and mirror the delta to the template source.

## §C Requirements (GEARS)

### C.1 Command surface

- **REQ-TODO-001** (Ubiquitous) The `moai todo` command shall expose the verbs `add`, `list`, `done`, and `next` operating on the backlog queue at `.moai/state/kanban/backlog.json`.
- **REQ-TODO-002** (Event-driven) **When** the operator runs `moai todo add "<text>"`, the command shall append the item to the queue under the lock and print the issued id and its queue position.
- **REQ-TODO-003** (Event-driven) **When** the operator runs `moai todo list`, the command shall render the queue without taking the lock and shall support a `--json` flag emitting the structured records.
- **REQ-TODO-004** (Event-driven) **When** the operator runs `moai todo done <n>`, the command shall remove the addressed row from the queue under the lock; a bare `<n>` argument shall be normalized to the item id `t<n>` (explicit id addressing is the preferred form because queue positions move under concurrent adds).
- **REQ-TODO-005** (Event-driven) **When** the operator runs `moai todo next` with no argument, the command shall print the queued items oldest-first as read-only candidates — the selection among them remains the operator's act performed through the lead session's question channel; **and when** the operator runs `moai todo next <n> [--spec <SPEC-ID>]`, the command shall mark the addressed item `picked` (attaching `spec_id` when given) as a single locked write.

### C.2 Store and concurrency

- **REQ-TODO-006** (Ubiquitous) The backlog store shall serialize each entire read-modify-write of the backlog file under a cross-process lock spanning load, mutate, and atomic write.
- **REQ-TODO-007** (Ubiquitous) The backlog store shall reuse the `internal/kanban` lock substrate with a sibling lock artifact at `.moai/state/kanban/backlog.lock`, acquired with a bounded retry window (~1 s) before contention surfaces as an explicit error naming the lock artifact path.
- **REQ-TODO-008** (Ubiquitous) The backlog store shall issue every new item id under the lock from a persisted high-water mark, such that ids are unique under concurrent adds and never reused after a row's removal (`done` removes rows, so max-present-id derivation alone would regress).
- **REQ-TODO-009** (Capability gate) **Where** the backlog file predates the high-water field, the store shall derive the initial high-water mark from the maximum existing item id, preserving load compatibility with version-1 files.
- **REQ-TODO-010** (Ubiquitous) The backlog store shall persist writes through same-directory temp file plus atomic rename (the `internal/atomicfile` substrate).
- **REQ-TODO-011** (Ubiquitous) The backlog store shall not apply the board's lead-role guard: the backlog is the operator's queue and any session may write it. (Deliberate contrast with `internal/kanban/board_store.go` `requireLeadRole`, which exists because the board has exactly one writer; the backlog has many.)

### C.3 File contracts (preserved verbatim from the todo skill)

- **REQ-TODO-012** (Event-driven) **When** the backlog file is absent, every verb shall treat the queue as empty — `list` and `next` report an empty queue, and `add` creates a version-1 file — never an error; **and when** the backlog file is malformed, every verb shall report the error and leave the file untouched, never silently resetting it, because the operator's queued intent is the one thing here that cannot be regenerated.

### C.4 Record shape

- **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record shape — `{"version":1,"items":[{"id","text","added_at","spec_id","state"}]}` with `state ∈ {queued, picked, dropped}` — changing it only additively (the high-water mark, per REQ-TODO-009).

### C.5 CLI discipline

- **REQ-TODO-014** (Ubiquitous) The `moai todo` command shall not prompt: it is headless-safe, asks no questions, and follows the `internal/cli` conventions (structured stdout for `--json`, human-readable stderr, exit 0/1/2, no AskUserQuestion anywhere in the package).

### C.6 Skill and template surfaces

- **REQ-TODO-015** (Ubiquitous) The todo skill (`.claude/skills/moai/workflows/todo.md`) shall be reduced to a "run the command" form — the queue philosophy, the boundaries, the kanban-dispatch cross-reference, and the [HARD] operator-selection rule are preserved, and the direct Read→Write instructions and the temp-file/rename guidance are removed — and the same skill delta shall be applied to the template twin (`internal/template/templates/.claude/skills/moai/workflows/todo.md`) and re-embedded via `make build`, keeping template neutrality (no SPEC ids, no internal dates).

### C.7 Platform

- **REQ-TODO-016** (Ubiquitous) The backlog store shall compile and behave on Windows through the lock substrate's platform split (flock on Unix, atomic-create on Windows), verified by `GOOS=windows` build and vet.

## §D Constraints

- No new module dependencies; the substrate is in-package (`internal/kanban`) plus `internal/atomicfile`.
- The kanban board state store (`board.json`, `WriteBoardState`, `requireLeadRole`) is NOT touched by this SPEC.
- Environment variable names, if any, must be constants in `internal/config/envkeys.go` per repo hardcoding policy.

## §E Out of Scope

### Out of Scope — Stale-lock automatic clearing

- Automatic stale-lock detection/clearing for `backlog.lock` is NOT built here. Contention beyond the bounded retry window surfaces as an explicit error naming the lock path; clearing a stale artifact is a deliberate operator-visible act (the board's `ClearStaleBoardLock` pattern shows the shape a future SPEC would reuse).
- On Unix the flock substrate releases automatically when the holder process dies, so the stale artifact case is a Windows-substrate concern deferred with this boundary.

### Out of Scope — Kanban board state store changes

- `internal/kanban` board surfaces (`board.go`, `board_store.go`, `board_recover.go`, WIP limits, role guards) are untouched. The backlog queue is a separate file with separate ownership rules (REQ-TODO-011).

### Out of Scope — Backlog schema redesign

- No version bump and no new per-item fields. The only schema change is the additive high-water mark (REQ-TODO-009); everything else stays load-compatible with the existing production file.

### Out of Scope — Skill workflow redesign

- The todo skill is shrunk, not redesigned. Queue philosophy, boundaries, and the operator-selection [HARD] rule survive verbatim in intent; no new workflow behavior is added to the skill.

## §F History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-14 | manager-spec | Initial plan-phase draft (kanban card t12). |
| 0.1.1 | 2026-08-14 | manager-spec | Plan-audit D2 consolidation: 19 → 16 REQs. Merged pairs: 005+006 → 005 (two When-branches, `next` read-only + `next <n>` pick), 013+014 → 012 (two When-branches, missing/malformed file contract), 017+018 → 015 (skill shrink + template mirror, both surfaces in one REQ). Renumbered: 007-012 → 006-011, 015-016 → 013-014, 019 → 016. All behaviors preserved verbatim; AC set unchanged (15). |

## §G Cross-References

- Kanban card t12 (`.moai/state/kanban/backlog.json`) — motivating incident and scope source.
- SPEC-KANBAN-BOARD-001 — the lock substrate and guarded-write precedent this SPEC reuses.
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board — the operator-act doctrine the verbs serve.
- `.claude/skills/moai/workflows/todo.md` — the surface being shrunk (current local and template twins are byte-identical, verified by diff).
