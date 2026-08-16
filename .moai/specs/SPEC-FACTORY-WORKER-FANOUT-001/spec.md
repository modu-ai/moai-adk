---
id: SPEC-FACTORY-WORKER-FANOUT-001
title: "Factory Mode -k <N>: numbered worker fan-out on cc/glm"
version: "1.2.0"
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

# SPEC-FACTORY-WORKER-FANOUT-001 — Factory Mode `-k <N>` (numbered worker fan-out)

Kanban card t68 (operator priority 2026-08-16: needed immediately in another
project). Integration vehicle: first merge of release/v3.1.1. Reworked in
place by card t85 (2026-08-17, v1.2.0): entry unified on `-k`, see
 Amendments below.

## GENEALOGY (binding)

The pre-3.1 "factory" flag (`-f` / `--factory`) was RENAMED to `-k` /
`--kanban` in #1513 (7f61332ef) and is the direct predecessor of today's
four-role kanban chain (SPEC-FACTORY-BOOTSTRAP-001 / SPEC-FACTORY-MODE-001
kept their historical names through that rename). **The `-f` this SPEC
originally introduced (v1.0.0, a numbered worker fan-out) was RETIRED the
same day (v1.2.0): the factory is entered with the SAME `-k` token carrying a
worker count, and any `-f`/`--factory` token is an explicit error
(`rejectRetiredFactoryFlag`).** Both launchers' help text carries this note
(`TestFactoryGenealogyInHelp` is the AC).

## What Factory Mode is

A factory run is one LEAD session plus N numbered WORKERS, entered through
the SAME `-k` token as Kanban Mode (v1.2.0 unification — one entry flag, the
launcher selects the backend, `-k` selects the role/shape):

- `moai cc -k <N>` (or `moai glm -k <N>`) enters as the **lead**. The
  SessionStart notice prints the run id, the `lead-<run-id>` session name, N
  worker launch lines (`moai cc -k <N> --name worker-<i>`), the GLM
  substitute guidance (`moai glm -k <N> --name ...`), the dispatch
  discipline (card-class routing, stagger, no model override), the free-slot
  line, and the leader socket path — the same notice shape and dual-channel
  emission (agent English / operator conversation_language) as the kanban
  bootstrap.
- `moai cc -k <N> --name worker-<i>` enters as **worker i**. The count
  travels in the worker's own `-k <N>` token, which is why the worker launch
  command carries it; a bare `-k --name worker-<i>` (no N) also enters as a
  worker with the default count (REQ-FF-001).
- Card routing is the lead's discipline: the queue is polled and cards picked
  by the operator or the kanban foreman loop (t96), and the lead routes
  PICKED cards to worker lanes by class (REQ-FF-008); reference
  implementation is the 2026-08-16 tjv7iy 4-lane run.

## Requirements

- **REQ-FF-001 (parse — unified -k, v1.2.0)**: one `-k`/`--kanban` token
  selects either shape. `-k N` / `-k=N` / `--kanban N` / `--kanban=N`
  (integer N ≥ 1) select the FACTORY with N workers; a non-numeric
  positional is the kanban SPEC identifier (bare `-k` or `-k SPEC-ID` is the
  kanban chain — unchanged); `-k --name worker-<i>` with no positional
  selects the factory WORKER entry with the operator-decided default count
  `config.DefaultFactoryWorkers` (8) — a count-less factory LEAD does not
  exist, because a bare `-k` is already the kanban lead. A SUPPLIED value
  that is numeric but < 1 (or `=`-joined and non-numeric) is an error naming
  every accepted shape (`kanbanFlagUsageError`). The numeric-positional
  discriminator is unambiguous: a SPEC identifier is never a bare integer.
  The retired `-f`/`--factory` token is an explicit error
  (`rejectRetiredFactoryFlag`) on cc / glm / cg. No upper bound (v1): runtime
  concurrency caps govern. The `--` pass-through discipline matches every
  other launcher parser.
- **REQ-FF-002 (branch truth table — v1.2.0)**: `-k N` + worker-shape
  `--name` → factory worker; `-k N` otherwise → factory lead; `-k` (no
  numeric positional) + worker-shape `--name` → factory worker at the
  default count; `-k` without either → kanban lead/companion (the §A.2 truth
  table, unchanged); no `-k` → no-op regardless of name.
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
  probe valid. The registry cluster (entry shape, load/save, prune,
  free-slot picker) lives in `internal/kanban/factory_slots.go` (v1.2.0
  move) so the SessionStart hook shares it — `internal/hook` cannot import
  `internal/cli`.
- **REQ-FF-005 (notice)**: lead notice emits exactly N launch lines carrying
  `-k <N>`, the GLM substitute line, the dispatch discipline block
  (REQ-FF-008 routing + the stagger and no-model-override rules), the
  free-slot line (from the worker registry under the project root,
  fail-open), and the socket path, on the kanban notice's startup-only
  re-entry gate; worker notice is a single join line naming the (possibly
  bumped) label — the reliable surface for a bump, since the launcher's
  stderr note is overwritten when the TUI takes the screen. Four conversation
  locales; protocol tokens verbatim in every locale. The notice does NOT
  teach queue polling — that loop is the kanban foreman's (t96: `.claude/`
  `loop.md` + the moai-kanban-foreman skill), and a second polling protocol
  in the notice would conflict with it.
- **REQ-FF-006 (block cap)**: factory sessions (lead and worker) take the
  raised Stop-hook block cap via `MOAI_FACTORY_WORKERS`, on the same
  unconditional branch as kanban.
- **REQ-FF-007 (exclusions — v1.2.0)**: the -k/-f mutual-exclusion clause is
  RETIRED with `-f` itself. `moai cg -k <N>` (and the count-less worker
  form) is rejected with sentinel `FACTORY_MODE_UNSUPPORTED_BACKEND`; the
  plain kanban shapes of `-k` on cg keep the kanban sentinel
  (`rejectKanbanOnCG` passes factory shapes through to the factory
  rejection); a malformed count on cg surfaces the parse error; the retired
  `-f` token on cg surfaces the retirement error.
- **REQ-FF-008 (card-class routing — v1.2.0)**: the lead routes PICKED
  cards by the kanban-dispatch card classes. A-class and B-class cards are
  fanned out WHOLESALE — the whole card goes to one worker lane. C-class
  cards (design changes) run the serial 3-stage path (`plan -> run -> sync`,
  one stage completing before the next begins) and are never fanned out.
  The stagger rule governs FACTORY fan-out only: activate the first worker,
  wait for its first output, then the remaining free-slot workers
  (cache-aware-execution directive 2) — the workflow runtime staggers itself
  (`CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS`) and is not governed by this
  rule. Dispatches carry NO model override: the GLM tier mapping rides the
  `ANTHROPIC_DEFAULT_*_MODEL` slot env, and `glm_task` enforces the guard
  mechanically when `MOAI_FACTORY_WORKERS` is in the server env (the
  override is ignored with a note).

## v1 ceilings (deliberate, named)

- Workers carry no run id (no `worker-<run-id>` form): a worker is addressed
  by name alone and replies ride the dispatch's reply address. Two concurrent
  factory runs in ONE project therefore share the worker namespace — the bump
  keeps session names unique but cannot disambiguate runs; the kanban-style
  run-id suffix (t56 territory for `-k`) is the upgrade path.
- The registry is liveness-only. A same-second relaunch after an unclean exit
  reuses the number once the dead pid is observed dead.
- No worker-count ceiling and no workload balancing: the lead routes; the
  runtime's concurrency cap bounds what is useful.

## Acceptance criteria → evidence

All commands run in the implementation worktree (branch
`worktree-factory-mode`, base `origin/main` @ d6b80a01c) on 2026-08-17:

- **AC-1** parse/branch/env/bump/reject/help tables green:
  `go test ./internal/cli/ -run 'TestParseKanbanFlag|TestResolveFactoryBranch|TestParseFactoryWorkerLabel|TestEnterFactory|TestResolveFactoryWorkerName|TestReplaceNamedLabel|TestRejectRetiredFactoryFlag|TestRejectFactoryOnCG|TestRejectKanbanOnCGLeavesFactoryForms|TestFactoryGenealogyInHelp|TestFactoryRaisesBlockCap' -v`
  → `ok github.com/modu-ai/moai-adk/internal/cli` (v1.2.0 test set; the
  v1.0.0 names `TestParseFactoryFlag` / `TestRejectConflictingModes` were
  retired with the -f surface).
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

## Amendments

### v1.2.0 (2026-08-17, card t85 scope correction — same-day in-place rework)

Operator finalized the §6 entry-surface decision: the `-f`/`--factory` entry
flag is RETIRED; the factory is entered via `-k <N>` (one entry token; the
launcher selects the backend, `-k` selects the shape). Changed in place on
branch WT-t85 (commit after 9794f293d):

- REQ-FF-001 rewritten for the unified `-k` parse (`parseKanbanFlag` now
  returns a `kanbanEntryParse` and owns the whole truth table). The v1.1.0
  bare-`-f` default became the count-less WORKER entry
  (`-k --name worker-<i>` → default 8); a count-less factory LEAD does not
  exist — a bare `-k` is the kanban lead. This is the documented
  minimal-coherence placement of the operator's N=8 default.
- REQ-FF-002 truth table updated; REQ-FF-007 rewritten (the -k/-f exclusion
  died with `-f`; the cg rejection split by shape); new REQ-FF-008
  (card-class routing: A/B wholesale, C serial 3-stage, fan-out-only stagger
  excluding `CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS`, glm_task no-model-
  override guard kept from v1.1.0).
- t96 de-duplication: queue polling / free-slot picking / fan-out mechanics
  are the kanban foreman loop's (`.claude/loop.md` +
  `moai-kanban-foreman`). The lead notice's v1.1.0 loop block was stripped to
  the factory-specific discipline (class routing + stagger + no-override +
  foreman handoff line); the queued-count data line went with the polling
  instruction (the foreman reads the queue itself), the free-slot line stays
  (the stagger rule names free-slot workers). The cli wrappers
  `factoryFreeSlots`/`factoryQueuedCards` were removed (nothing Go-side
  consumed them); the shared shapes stay in `internal/kanban`
  (`FactoryFreeSlots` consumed by the hook notice;
  `BacklogStore.QueuedCount`/`QueuedBacklogCountForRoot` consumed by the
  kanban notice).
- t94 interplay: `orchestration-mode-selection.md` (both copies) updated
  from the `-f` reference to the `-k <N>` surface.
- Registry/name machinery, block cap, session records: unchanged — only the
  entry token changed.

## Close plan

Sync close (3-phase contract) rides the release vehicle: lead merges this
branch into release/v3.1.1 as its first merge, and the release PR carries the
`implemented → completed` transition + `sync_commit_sha`. README/docs-site
coverage is the t61 sync lane's scope, not this card's.
