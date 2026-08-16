---
id: SPEC-FACTORY-WORKER-FANOUT-001
title: "Factory Mode -f <N>: numbered worker fan-out on cc/glm"
version: "1.0.0"
status: implemented
created: 2026-08-17
updated: 2026-08-17
author: kanban lane t68 (run tjv7iy; operator-priority card, direct lane implementation)
priority: P1
phase: "v3.1.x"
module: "internal/cli, internal/kanban, internal/hook, internal/config"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "factory-mode, worker-fanout, cc-flag, glm-flag, cross-session-dispatch, name-bump"
---

# SPEC-FACTORY-WORKER-FANOUT-001 — Factory Mode `-f <N>` (numbered worker fan-out)

Kanban card t68 (operator priority 2026-08-16: needed immediately in another
project). Integration vehicle: first merge of release/v3.1.1.

## GENEALOGY (binding)

The pre-3.1 "factory" flag (`-f` / `--factory`) was RENAMED to `-k` /
`--kanban` in #1513 (7f61332ef) and is the direct predecessor of today's
four-role kanban chain (SPEC-FACTORY-BOOTSTRAP-001 / SPEC-FACTORY-MODE-001
kept their historical names through that rename). **The `-f` this SPEC
introduces is a NEW feature — a numbered worker fan-out — and shares nothing
with that predecessor beyond the recycled letter.** Both launchers' help text
carries this note verbatim (`TestFactoryGenealogyInHelp` is the AC).

## What Factory Mode is

A factory run is one LEAD session plus N numbered WORKERS:

- `moai cc -f <N>` (or `moai glm -f <N>`) enters as the **lead**. The
  SessionStart notice prints the run id, the `lead-<run-id>` session name, N
  worker launch lines (`moai cc -f <N> --name worker-<i>`), the GLM
  substitute guidance (`moai glm -f <N> --name ...`), and the leader socket
  path — the same notice shape and dual-channel emission (agent English /
  operator conversation_language) as the kanban bootstrap.
- `moai cc -f <N> --name worker-<i>` enters as **worker i**. The count
  travels in the worker's own `-f <N>` token, which is why the worker launch
  command carries it.
- Card distribution is the lead's cross-session SendMessage dispatch to the
  worker names; reference implementation is the 2026-08-16 tjv7iy 4-lane run.

## Requirements

- **REQ-FF-001 (parse)**: `-f`/`--factory` accepts a REQUIRED count N ≥ 1 in
  the forms `-f N`, `-f=N`, `--factory N`, `--factory=N`; anything else is an
  error naming the expected form. No upper bound (v1): runtime concurrency
  caps govern. The `--` pass-through discipline matches every other launcher
  parser.
- **REQ-FF-002 (branch truth table)**: `-f N` + worker-shape `--name` →
  worker; `-f N` otherwise → lead; no `-f` → no-op regardless of name (a
  `worker-<i>` name alone joins nothing — symmetric with kanban's
  REQ-FB-001).
- **REQ-FF-003 (bootstrap reuse, mode branch)**: the lead reuses the kanban
  bootstrap's run-id minting/adoption (`leadRunID`), `lead-<run-id>` name
  injection, leader-socket path, transient `--settings` injection, autonomy
  tier seed, and session record (role `lead`); workers record role `worker`.
  The kanban chain variables (`MOAI_KANBAN`, `MOAI_KANBAN_LABEL`) are never
  set — the factory discriminator is `MOAI_FACTORY_WORKERS` (both branches)
  plus `MOAI_FACTORY_WORKER` (worker only).
- **REQ-FF-004 (name uniqueness / conflict bump)**: a worker number whose
  label is registered to a LIVE pid is bumped to the next free number before
  the backend argv is built (`replaceNamedLabel` carries the bumped value
  into the session name — the name is the dispatch address). Claims live in
  `.moai/state/factory/workers.json` keyed by project root; dead claims are
  pruned and free the name (a relaunch reuses its number); every registry
  failure is fail-open to the label as supplied. The launcher exec's without
  forking, so the recorded pid IS the session pid — that is what makes the
  probe valid.
- **REQ-FF-005 (notice)**: lead notice emits exactly N launch lines carrying
  `-f <N>`, the GLM substitute line, and the socket path, on the kanban
  notice's startup-only re-entry gate; worker notice is a single join line
  naming the (possibly bumped) label — the reliable surface for a bump, since
  the launcher's stderr note is overwritten when the TUI takes the screen.
  Four conversation locales; protocol tokens verbatim in every locale.
- **REQ-FF-006 (block cap)**: factory sessions (lead and worker) take the
  raised Stop-hook block cap via `MOAI_FACTORY_WORKERS`, on the same
  unconditional branch as kanban.
- **REQ-FF-007 (exclusions)**: `-k` + `-f` together is rejected before any
  environment mutation; `moai cg -f` is rejected with sentinel
  `FACTORY_MODE_UNSUPPORTED_BACKEND` (mixed backend contradicts
  one-session/one-backend), and a malformed count on cg surfaces the parse
  error.

## v1 ceilings (deliberate, named)

- Workers carry no run id (no `worker-<run-id>` form): a worker is addressed
  by name alone and replies ride the dispatch's reply address. Two concurrent
  factory runs in ONE project therefore share the worker namespace — the bump
  keeps session names unique but cannot disambiguate runs; the kanban-style
  run-id suffix (t56 territory for `-k`) is the upgrade path.
- The registry is liveness-only. A same-second relaunch after an unclean exit
  reuses the number once the dead pid is observed dead.
- No worker-count ceiling and no workload balancing: the lead dispatches; the
  runtime's concurrency cap bounds what is useful.

## Acceptance criteria → evidence

All commands run in the implementation worktree (branch
`worktree-factory-mode`, base `origin/main` @ d6b80a01c) on 2026-08-17:

- **AC-1** parse/branch/env/bump/reject/help tables green:
  `go test ./internal/cli/ -run 'TestParseFactoryFlag|TestResolveFactoryBranch|TestParseFactoryWorkerLabel|TestEnterFactory|TestResolveFactoryWorkerName|TestReplaceNamedLabel|TestRejectConflictingModes|TestRejectFactoryOnCG|TestFactoryGenealogyInHelp|TestFactoryRaisesBlockCap' -v`
  → `ok github.com/modu-ai/moai-adk/internal/cli` (13 tests PASS).
- **AC-2** label vocabulary + record admission:
  `go test ./internal/kanban/` → `ok ... 25.442s` (whole package; includes
  the new factory_label_test.go).
- **AC-3** notices (lines/socket/GLM/startup-gate/suppression/i18n):
  `go test ./internal/hook/ -run 'TestFactory|TestKanban' -v` → 28 PASS.
- **AC-4** regression — kanban surfaces unchanged: full
  `go test ./internal/cli/` shows only the pre-existing
  `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` failure, which
  fails identically on the PRISTINE worktree (verified via stash) — local
  codex-auth environment, not this card. Full `go test ./internal/hook/`
  likewise shows only the pre-existing `TestPreTool_AstGrepSkipReasonSurfaces`
  failure (local-only ast-grep ruleset absent in the worktree; also verified
  via stash on the pristine tree).
- **AC-5** cross-platform: `GOOS=windows GOARCH=amd64 go vet ./internal/cli/
  ./internal/kanban/ ./internal/hook/` and `GOOS=windows GOARCH=amd64 go
  build ./...` both clean (the windows liveness twin compiles).
- **AC-6** binary: `make build` succeeds; catalog.yaml byte-unchanged (no
  template files touched by this card).

## Close plan

Sync close (3-phase contract) rides the release vehicle: lead merges this
branch into release/v3.1.1 as its first merge, and the release PR carries the
`implemented → completed` transition + `sync_commit_sha`. README/docs-site
coverage is the t61 sync lane's scope, not this card's.
