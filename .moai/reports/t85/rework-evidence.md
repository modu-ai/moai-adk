# t85 rework run evidence — entry unification on `-k N` (commit a34494665, base 9794f293d)

Persisted per review-hub request (inline-message evidence gap). Scope correction source:
operator final decision (§6) — `-f`/`--factory` entry flag retired; factory entered via
`-k N` (`moai cc -k N` / `moai glm -k N` — launcher selects backend, `-k` selects role).

## Claim

All rework items delivered on top of 9794f293d (which already includes t85 v1
59dd62d02 + t94 4c5994237 + t96 5c3141372): `-f` retired with explicit error; unified
`-k N` parse (default-8 placed on the count-less worker entry — documented choice);
card-class path split codified (A/B wholesale, C serial 3-stage) in 4 locales + SPEC;
t96 de-duplication executed (notice loop block stripped to factory-specific
discipline, dead cli helpers removed, shared kanban shapes kept); stagger scoped to
fan-out only with `CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS` excluded; glm_task guard
untouched; SPEC v1.2.0; t94 file (both copies) updated to the `-k <N>` surface.

## Files (13, +629/−566)

- internal/cli/kanban.go — `kanbanEntryParse` struct + error; numeric-positional
  discrimination; `=`-joined count forms; worker-name-default resolution;
  `rejectKanbanOnCG` passes factory shapes through
- internal/cli/factory.go — `parseFactoryFlag`/flag consts/`rejectConflictingModes`/
  cli wrappers `factoryFreeSlots`+`factoryQueuedCards` DELETED;
  `rejectRetiredFactoryFlag` added; `rejectFactoryOnCG` rewritten on unified parse
- internal/cli/{cc,glm,cg}.go — routing on unified parse; retired-flag guard; help
  texts (factory folds into `-k` section; genealogy rewritten with RETIRED)
- internal/cli/{factory,cc}_test.go — `TestParseKanbanFlagUnifiedEntry` (16-case
  truth table), `TestParseKanbanFlagPassThroughBoundary`, `TestRejectRetiredFactoryFlag`,
  `TestRejectKanbanOnCGLeavesFactoryForms`; `TestFactoryGenealogyInHelp` gains
  RETIRED/`-k <N>` markers + asserts no `-f` entry form
- internal/hook/session_start_factory{,_i18n,_test}.go — launch lines `-k <N>`;
  loop block → `leadClasses`/`leadStagger`/`leadNoOverride` (4 locales);
  queued-count line removed; free-slot line kept
- .claude/rules/moai/workflow/orchestration-mode-selection.md + template mirror —
  line 171 `moai glm -f` → `moai glm -k <N>`; copies verified identical via diff
- .moai/specs/SPEC-FACTORY-WORKER-FANOUT-001/spec.md — v1.2.0 amendments

## De-dup decision + evidence (t96 absorbs / t85 keeps)

- Absorbed by t96 (evidence: `.claude/loop.md` + `.claude/skills/moai-kanban-foreman/SKILL.md`
  §The iteration): queue watching (Monitor cksum poll), queue reading
  (`moai todo list --json`), card picking, dispatch mechanics, collection-on-evidence.
  The v1.1.0 notice loop block's polling sentence taught a SECOND, conflicting
  protocol — removed, replaced by a one-line handoff ("the queue is polled and cards
  are picked by the operator or the kanban foreman loop (bare `/loop`); this factory
  routes PICKED cards to worker lanes"), matching the foreman skill's own
  "Factory seam (reserved, not implemented)" reservation.
- Kept (factory-specific): class routing, fan-out-only stagger (excluding
  `CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS`), no-model-override, N launch lines
  (`-k <N>`), glm_task Go guard.
- Data lines: free-slot line KEPT (the stagger rule names "free-slot workers";
  rendered from `kanban.FactoryFreeSlots`). Queued-count line REMOVED (its consumer
  was the polling instruction; the foreman reads the queue itself).
- Go helpers: cli wrappers removed (nothing Go-side consumed them — dead-code
  discipline). Shared shapes in internal/kanban KEPT with live consumers:
  `FactoryFreeSlots`/`PruneFactoryDeadClaims` (hook notice + `resolveFactoryWorkerName`),
  `BacklogStore.QueuedCount`/`QueuedBacklogCountForRoot`/`BacklogPathForRoot`
  (kanban lead notice).

## RED (pre-rework, against the -f surface @ 9794f293d)

- `go test ./internal/cli/ -run 'TestRejectFactoryOnCG|TestRejectKanbanOnCGLeavesFactoryForms' -count=1`
  → FAIL: "worker-name factory form on cg must carry the sentinel, got <nil>" +
  "invalid factory count must surface the parse error, got <nil>" +
  "[-k 4] is the factory rejection's, not kanban's, got KANBAN_MODE_UNSUPPORTED_BACKEND"
- `go test ./internal/hook/ -run 'TestFactoryLeadNoticeCarriesWorkerLines...|...WorkerCountDrivesLineCount' -count=1`
  → FAIL ("N=1 launch line count = 0, want 1" — notice still printed `-f`)

## GREEN (this run, tree @ a34494665)

- `go build ./...` → exit 0; `go vet ./internal/cli/ ./internal/kanban/ ./internal/hook/` → exit 0
- `go test ./internal/cli/ -run 'Kanban|Factory|GLMTask|TestParse' -count=1` → ok 2.651s
- renamed-test batch → ok 1.225s
- `go test ./internal/kanban/ -count=1` → ok 14.233s
- `go test ./internal/hook/ -run 'TestFactory|TestKanban' -count=1` → ok 2.831s
- `golangci-lint run internal/cli/... internal/kanban/... internal/hook/...` → 0 issues
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `make build` → succeeds; catalog.yaml byte-unchanged
- `go test ./internal/template/ -count=1` → ok 34.460s (mirror parity + neutrality)
- Snapshot recorded: key HEAD:a34494665

## Orchestrator independent re-verification (post-commit, direct re-run)

- `go build ./...` → BUILD OK
- `go test ./internal/cli/ -run 'Kanban|Factory|GLMTask|TestParse' -count=1` → ok 1.425s
- `go test ./internal/kanban/ -count=1` → ok 24.541s
- `go test ./internal/hook/ -run 'TestFactory|TestKanban' -count=1` → ok 0.559s
- t94 pair diff → IDENTICAL

## Default-8 placement (material ambiguity, resolved and documented)

Bare `-k` is already the kanban lead, so a count-less factory LEAD cannot exist. The
operator's N=8 default lives on the only count-less context that unambiguously
selects the factory: the worker-shape name (`-k --name worker-<i>` → worker, default
8). Documented in SPEC REQ-FF-001 + the t94 file edit. A factory lead always
declares N explicitly.

## Gaps

- `parseKanbanFlag` signature changed; all 6 in-tree call sites migrated, build green.
- CLAUDE.md / template CLAUDE.md checked — no `-f` factory references existed.
- Full `go test ./internal/cli/` not run locally (card-scoped discipline; CI owns it).
- `-k=SPEC` form invalid by design (usage error names all shapes).

## Residual risks

- `moai cc -k 4` shadows what was previously "spec id 4" — no valid SPEC identifier
  is a bare integer (discriminator sound); no in-repo numeric-positional caller found.
- The foreman handoff line is doctrine-level: the foreman skill's factory seam is
  "reserved, not implemented" — a future t96 follow-up grows lane routing.
- `-k` + `--name worker-<i>` + explicit SPEC-ID rides the id into the session record
  (pre-existing behavior class; untested combination).
