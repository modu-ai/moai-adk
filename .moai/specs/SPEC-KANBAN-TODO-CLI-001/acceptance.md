---
id: SPEC-KANBAN-TODO-CLI-001
title: "Acceptance — moai todo CLI subcommand with lock-guarded backlog store"
version: "0.1.1"
status: in-progress
created: 2026-08-14
updated: 2026-08-14
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, cli, backlog, concurrency, acceptance"
tier: M
---

# acceptance.md — SPEC-KANBAN-TODO-CLI-001

Verification layer. Each entry is an `AC-XXX` labeled **Given-When-Then** scenario, binary-testable. The GEARS obligation lives in the requirement layer (`spec.md` §C `REQ-TODO-001`…`REQ-TODO-016`); nothing here restates a GEARS requirement, and no Given-When-Then scenario is presented as one.

## §A Coverage matrix (REQ → AC)

16 REQs covered by 15 ACs (Tier M AC ceiling). Grouping is by shared verification command; every REQ maps to at least one AC. REQ numbering per spec.md §F History v0.1.1 consolidation map.

| REQ | AC | REQ | AC | REQ | AC |
|-----|----|-----|----|-----|----|
| 001 | AC-TODO-001 | 007 | AC-TODO-007 | 013 | AC-TODO-013 |
| 002 | AC-TODO-001 | 008 | AC-TODO-008 | 014 | AC-TODO-014 |
| 003 | AC-TODO-003 | 009 | AC-TODO-008 | 015 | AC-TODO-015 |
| 004 | AC-TODO-004 | 010 | AC-TODO-009 | 016 | AC-TODO-016 |
| 005 | AC-TODO-005, AC-TODO-006 | 011 | AC-TODO-010 | | |
| 006 | AC-TODO-007 | 012 | AC-TODO-011, AC-TODO-012 | | |

## §B Acceptance criteria

### AC-TODO-001 — concurrent `add` loses nothing (REQ-TODO-001, REQ-TODO-002)

**Given** a temp-dir backlog with the `moai todo` command wired, **When** 8 concurrent `moai todo add "card N"` processes run simultaneously, **Then** a subsequent `moai todo list --json` reports exactly 8 items, each carrying the text it was added with, and each `add` printed its issued id and queue position.

### AC-TODO-003 — `list` is lock-free and structured (REQ-TODO-003)

**Given** a backlog holding queued and picked items, **When** `moai todo list` runs while `backlog.lock` is held by another process, **Then** the command succeeds and renders the queue without acquiring the lock; **and when** `moai todo list --json` runs, **Then** stdout is valid JSON carrying the full record shape.

### AC-TODO-004 — `done` removes by id under lock (REQ-TODO-004)

**Given** a backlog holding item `t3`, **When** `moai todo done 3` runs, **Then** the row `t3` is absent from the file after the command and the exit code is 0; **and when** `moai todo done t3` runs against a queue lacking `t3`, **Then** the command exits non-zero with the miss reported.

### AC-TODO-005 — bare `next` is read-only (REQ-TODO-005)

**Given** a backlog holding items added in known order, **When** `moai todo next` runs, **Then** the items are printed oldest-first and the backlog file is byte-identical before and after (no state change, no `picked` transition).

### AC-TODO-006 — `next <n> [--spec]` is one locked write (REQ-TODO-005)

**Given** a backlog holding queued item `t2`, **When** `moai todo next 2 --spec SPEC-X-001` runs, **Then** exactly item `t2` carries `state: "picked"` and `spec_id: "SPEC-X-001"` after the command, and no other row changed — a single mutation observable in one file write.

### AC-TODO-007 — lock serialization, bounded retry, named-path error (REQ-TODO-006, REQ-TODO-007)

**Given** `backlog.lock` held by a foreign process, **When** `moai todo add "x"` runs, **Then** the command retries within the bounded window (~1 s: 25 ms × 40) and then exits non-zero with an error message containing the literal lock artifact path `.moai/state/kanban/backlog.lock` (or the temp-dir equivalent); **and when** two mutating verbs run concurrently, **Then** both complete with correct results — the loser serialized by the retry, not corrupted (composes with AC-TODO-001).

### AC-TODO-008 — ids unique, monotonic, never reused (REQ-TODO-008, REQ-TODO-009)

**Given** 8 concurrent adds (AC-TODO-001 fixture), **When** ids are collected, **Then** all 8 are distinct and drawn from one monotonic `t<n>` sequence; **given** a version-1 file with no high-water field and max id `t19`, **When** the store first loads and then writes, **Then** the next issued id is `t20` and the persisted file carries the derived high-water mark; **and given** `t20` is subsequently removed via `done`, **When** the next `add` runs, **Then** the issued id is `t21`, not `t20` — removal never enables reuse.

### AC-TODO-009 — writes are atomic temp+rename (REQ-TODO-010)

**Given** any mutating verb, **When** it completes, **Then** the backlog file is valid parseable JSON ending in a newline, the write went through `internal/atomicfile` (same-directory temp + rename — verified by the store's unit test asserting the `atomicfile.Replace` call path, not by grep of production source), and no `.tmp` residue remains in the directory.

### AC-TODO-010 — no lead-role guard (REQ-TODO-011)

**Given** a process holding no lead role / kanban identity, **When** it runs `moai todo add "x"` against a temp-dir backlog, **Then** the write succeeds — the backlog applies no `requireLeadRole`-equivalent gate (explicit contrast test; board files remain untouched by this SPEC).

### AC-TODO-011 — missing file is an empty queue (REQ-TODO-012)

**Given** no backlog file exists, **When** `moai todo list`, `moai todo next`, or `moai todo add "first"` run, **Then** the first two report an empty queue and exit 0, and the third creates a valid version-1 file containing exactly one item — none of the three errors.

### AC-TODO-012 — malformed file reported, never reset (REQ-TODO-012)

**Given** a backlog file containing invalid JSON, **When** each verb runs, **Then** every verb exits non-zero with the parse error reported, and the file is byte-identical (checksummed before and after) — no silent recovery path exists.

### AC-TODO-013 — version-1 record shape preserved, change additive only (REQ-TODO-013)

**Given** a production-shaped version-1 file (`{"version":1,"items":[{"id","text","added_at","spec_id","state"}]}`, states `queued`/`picked`/`dropped`), **When** the store round-trips it (load → any mutation → write), **Then** every pre-existing item retains all five fields with unchanged values, `version` remains `1`, no per-item field is added, and the only top-level addition is the high-water mark.

### AC-TODO-014 — CLI no-prompt static guard (REQ-TODO-014)

**Given** the built `internal/cli` package, **When** the todo-command static guard test runs (the `TestNew_NoAskUserQuestion` grep-based pattern scoped to the todo command surface), **Then** it passes — zero `AskUserQuestion`/interactive-prompt references in the todo command path, and the guard fails if one is introduced (negative control asserted in test).

### AC-TODO-015 — skill shrunk, twins delta-matched, template neutral (REQ-TODO-015)

**Given** the run-phase edits, **When** the shrunk local skill is inspected, **Then** it retains the queue philosophy, the boundaries, the kanban-dispatch cross-reference, and the [HARD] operator-selection rule, and no longer contains Read→Write file-manipulation or temp-file/rename instructions (absence verified by grep for the removed instruction tokens); **and when** the template twin is diffed against the local twin, **Then** the same delta is present on both; **and when** `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` runs, **Then** it exits 0 — no SPEC IDs and no internal dates in the template twin; **and** `make build` regenerated `catalog.yaml` is committed (mirror-parity CI green).

### AC-TODO-016 — Windows compile gate covers the test layer (REQ-TODO-016)

**Given** the completed implementation, **When** `GOOS=windows GOARCH=amd64 go build ./...` AND `GOOS=windows GOARCH=amd64 go vet ./...` run, **Then** both exit 0 with exit codes cited verbatim in the run-phase evidence — the vet result is the load-bearing half because `go build` never compiles `_test.go`.

## §C Edge cases (binary coverage demanded by §B)

- Empty-text add (`moai todo add ""`) → rejected with usage error, exit non-zero, no file write.
- `done`/`next <n>` with an out-of-range `<n>` → non-zero, miss reported, file untouched.
- `--spec` value not matching `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → accepted verbatim (store is not a SPEC registry; validation is out of scope) but recorded as-is — no normalization.
- Backlog file with `last_seq` smaller than a present id (hand-edited) → next issued id is still greater than every present id (high-water = max(persisted, max-present)).

## §D Quality gates

- Package coverage ≥ 85% for the new store and CLI files (`go test -cover ./internal/kanban/... ./internal/cli/...`, cited output).
- Full-suite rule: `go test ./...` green (affected-package-only self-report insufficient — template-mirror CI precedent).
- Lint: zero NEW findings vs the §C pre-flight baseline (`golangci-lint run`).
- MX tags: new exported store functions are `@MX:ANCHOR` candidates per plan.md §F; the locked-mutation path carries `@MX:WARN` (concurrency pattern).

## §E Definition of Done

All 15 ACs PASS with attributed evidence (command + verbatim output + tree SHA), the four milestones in plan.md §F complete, frontmatter transitions owned per the Status Transition Ownership Matrix, and no production file `.moai/state/kanban/backlog.json` was mutated by tests or run-phase commands.
