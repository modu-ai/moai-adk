# t68 Factory Mode -f N — verification evidence

Worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/factory-mode
Branch: worktree-factory-mode | Base: origin/main @ d6b80a01c | HEAD: b56f5dbc2
Date: 2026-08-17 (KST) | Card: t68 | SPEC: SPEC-FACTORY-WORKER-FANOUT-001

## Claim / Evidence (verbatim observed outputs)

### internal/kanban (whole package, includes new factory_label_test.go)
$ go test ./internal/kanban/
ok  	github.com/modu-ai/moai-adk/internal/kanban	25.442s

### internal/config (envkeys.go touched)
$ go test ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/config	1.565s

### internal/cli — new factory tests, verbose
$ go test ./internal/cli/ -run 'TestFactory|TestParseFactory|TestResolveFactory|TestReplaceNamedLabel|TestRejectConflictingModes|TestRejectFactoryOnCG|TestParseFactoryWorkerLabel|TestEnterFactory' -v
13 × --- PASS (TestParseFactoryFlag, TestParseFactoryFlagStopsAtPassThroughMarker,
TestResolveFactoryBranch, TestParseFactoryWorkerLabelRecognizesWithoutConsuming,
TestEnterFactoryLeadModeEnv, TestEnterFactoryLeadModeMintsRunID,
TestEnterFactoryWorkerModeEnv, TestResolveFactoryWorkerName ×4 subtests,
TestReplaceNamedLabel ×7 subtests, TestRejectConflictingModes,
TestRejectFactoryOnCG, TestFactoryGenealogyInHelp, TestFactoryRaisesBlockCap)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.396s

### internal/hook — factory + kanban notices, verbose
$ go test ./internal/hook/ -run 'TestFactory|TestKanban' -v
28 × --- PASS, ok github.com/modu-ai/moai-adk/internal/hook 0.670s

### Regression — full package runs
$ go test ./internal/cli/
FAIL — exactly ONE failure: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey
$ go test ./internal/hook/
FAIL — exactly ONE failure: TestPreTool_AstGrepSkipReasonSurfaces

Both verified PRE-EXISTING on the pristine worktree via `git stash -u` →
rerun → same FAIL → `git stash pop` (local codex-auth / local-only ast-grep
ruleset absent in the worktree; unrelated to this card).

$ go test ./internal/cli/ -run 'TestKanban|TestParseKanbanFlag|TestResolveKanbanBranch|TestParseCompanionLabel|TestGLM.*Kanban|TestLeadName|TestCC'
ok  	github.com/modu-ai/moai-adk/internal/cli	2.778s

### Cross-platform (windows liveness twin compiles)
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/ ./internal/kanban/ ./internal/hook/   → clean
$ GOOS=windows GOARCH=amd64 go build ./...                                              → clean

### Lint
$ golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/hook/... ./internal/config/...
→ 1 issue (errcheck, factory.go note write) → fixed with `_, _ =` → rerun: 0 issues.

### Build + binary smoke
$ make build → ok; catalog.yaml byte-unchanged (no template files touched).
$ ./bin/moai cc --help → Factory Mode section + genealogy block rendered.
$ ./bin/moai cg -f 3 → FACTORY_MODE_UNSUPPORTED_BACKEND (title-cased by the
  error renderer; raw error string carries the sentinel, test-asserted).
$ ./bin/moai cc -k -f 3 → "-k/--kanban and -f/--factory are mutually exclusive".

## Gaps
- golangci-lint full-repo run not performed (lane discipline: changed
  packages only; CI lint covers the PR).
- No live multi-terminal factory run executed (would exec into claude
  sessions); behavior verified at unit + binary-surface level.

## Residual risk
- The bump registry is liveness-only; two factory runs in ONE project share
  the worker namespace (v1 ceiling, documented in the SPEC).
- The `moai cc -f N` lead path itself was not exec'd end-to-end in this
  lane (it replaces the process); the branch logic, env contract, notice
  emission, and block cap are all covered by the tests above.
