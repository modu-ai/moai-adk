# t85 run evidence — factory lead loop (commit 59dd62d02, base 7588f0529)

> Persisted by card t94 at the review hub's request: t85's verification previously existed only as inline
> cross-session messages. This file is the persisted record of the t85 run that preceded t94 on this tree;
> the observations below were reported by the t85 implementation and orchestrator turns, not re-executed by t94.

## Implementation agent RED captures (pre-implementation)

- `go test ./internal/hook/ -run 'TestFactoryLeadNoticeCarriesLoopDiscipline' -count=1`
  → FAIL 0.640s — notice missing all 7 loop tokens (`.moai/state/kanban/backlog.json`, `moai todo list`,
  `SendMessage`, `started producing output`, `cache-aware-execution directive 2`, `No model override`,
  `ANTHROPIC_DEFAULT_*_MODEL`)
- `go test ./internal/cli/` → FAIL [build failed] — undefined: `config.DefaultFactoryWorkers` /
  `factoryFreeSlots` / `factoryQueuedCards` / `kanban.BacklogPathForRoot`

## Implementation agent GREEN (this run, tree @ 59dd62d02)

- `go build ./...` → exit 0
- `go vet ./internal/cli/ ./internal/kanban/ ./internal/hook/` → exit 0
- `go test ./internal/cli/ -run 'Factory|GLMTask' -count=1` → ok 1.061s
- `go test ./internal/cli/ -run 'TestParseFactoryFlag|TestResolveFactoryBranch|TestEnterFactory|TestReplaceNamedLabel|TestReject|TestFactoryGenealogy' -count=1` → ok 5.265s (11 subtests)
- `go test ./internal/kanban/ -count=1` → ok 20.203s
- `go test ./internal/hook/ -run 'TestFactory|TestKanban' -count=1` → ok 2.004s
- `golangci-lint run internal/cli/... internal/kanban/... internal/hook/...` → 0 issues
- `GOOS=windows GOARCH=amd64 go build ./...` + vet → exit 0
- verify snapshot recorded: key HEAD:59dd62d02

## Orchestrator independent re-verification (post-commit, direct re-run)

- `git show --stat 59dd62d02` → 18 files +852/−158; boundary grep → t94/t96 surfaces NOT in commit
- `go build ./...` → BUILD OK
- `go test ./internal/cli/ -run 'Factory|GLMTask' -count=1` → ok 1.800s
- `go test ./internal/kanban/ -count=1` → ok 15.043s
- `go test ./internal/hook/ -run 'TestFactory|TestKanban' -count=1` → ok 0.629s

## Gaps (carried from the implementation report)

- glm_task env-inheritance final link (claude→mcp-server stdio child) verified in code chain + production
  notice corroboration, not executed directly
- full `go test ./internal/cli/` not run locally (card-scoped discipline; CI owns the full matrix)
