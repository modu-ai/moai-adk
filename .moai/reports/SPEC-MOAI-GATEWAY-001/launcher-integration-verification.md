# Gateway launcher integration verification

## Claim

The launcher now has a bounded conversation-family integration seam. When an
approved `gatewaySessionOptions.Family` is supplied, the seam creates an
explicit native session descriptor before child startup, preserves the actual
working directory separately from the project index key, carries the selected
secure-storage namespace without copying credentials, and replaces ambiguous
resume/continue/fork input with manager-verified native arguments. Existing
launch behavior remains unchanged when the family manager is absent.

## Evidence

The following focused checks were run against the current worktree:

```text
$ go test ./internal/cli -run '^TestGatewayConversationNewUsesActualCWDAndPrivateSecureNamespace$' -count=1 -timeout=30s
ok  github.com/modu-ai/moai-adk/internal/cli  0.983s

$ go test ./internal/cli -run '^TestGatewaySession' -count=1 -timeout=25s -v
--- PASS: TestGatewaySessionStartsPrivateChildAndKeepsSecretsOutOfOverlay
--- PASS: TestGatewaySessionFailureAlwaysStopsStartedChild
--- PASS: TestGatewaySessionRealSupervisorHandoffAndCleanup
--- PASS: TestGatewaySessionSettingsCannotRedirectSelectedProfile
--- PASS: TestGatewaySessionDisablesInheritedFallbackWithoutMutatingCaller
--- PASS: TestGatewaySessionRejectsExplicitFallbackBeforeChildStart
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  0.954s

$ go test ./internal/gateway/conversation ./internal/gateway/opaque
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation  0.416s
ok  github.com/modu-ai/moai-adk/internal/gateway/opaque  0.583s

$ go vet ./internal/cli ./internal/gateway/conversation ./internal/gateway/opaque
(exit 0; no output)
```

The focused implementation is in `internal/cli/gateway_session.go` and
`internal/cli/gateway_launcher.go`. `internal/cli/launcher.go` supplies the
actual process CWD, project key, and explicit namespace-presence bits at the
existing binding seam. `internal/gateway/catalog.go` carries the verified native
argument descriptor through `LaunchPlan.Args`.

## Baseline-attribution

All commands above ran on branch `WT-unified-gateway`, current worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, after the
integration edits. The test package output is the observed output for this
tree, not a CI or prior-session result.

## Gaps

The stale `TestRunCC_WithTeamModeMessage` fixture was updated from retired
`team_mode: cg` to supported `team_mode: glm`. The production guard remains
unchanged: a real legacy `cg` configuration is rejected before launch. The
root-error wrapper in `guardCGLaunchMode` preserves the established
`find project root` error contract when that early guard runs.

Final regression readback:

```text
$ go test ./internal/cli -run 'TestRunCC_(NoProjectRoot|FindProjectRootFails|WithTeamModeMessage)' -count=1 -v
=== RUN   TestRunCC_NoProjectRoot
--- PASS: TestRunCC_NoProjectRoot (0.00s)
=== RUN   TestRunCC_WithTeamModeMessage
Team mode disabled (was: glm)
Launching Claude Code...
--- PASS: TestRunCC_WithTeamModeMessage (0.00s)
=== RUN   TestRunCC_FindProjectRootFails
--- PASS: TestRunCC_FindProjectRootFails (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  1.621s

$ go test ./internal/cli -run '^TestCGEntryGuardRunsBeforeLaunchAndSpawn$' -count=1 -v
=== RUN   TestCGEntryGuardRunsBeforeLaunchAndSpawn
--- PASS: TestCGEntryGuardRunsBeforeLaunchAndSpawn (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  0.945s
```

- No production factory registration was added. `newGatewayChildCommand(nil)`
  remains closed and no unverified OAuth or transport dependency is activated.
- The current CLI parser does not expose a separate public `--resume` or fork
  descriptor field; the seam recognizes the native pass-through forms when an
  approved family manager is supplied.
- Receipt authorization, opaque envelope projection, terminal persistence,
  and native transcript completion are not yet called from the HTTP request
  path. This slice only carries the pre-launch conversation descriptor.
- No Windows runtime execution or GitHub Actions run was observed here.
- The full `internal/cli` suite was not used as a completion claim because the
  existing repository contains long-running integration tests; the focused
  gateway session suite and package tests above are the scoped evidence.

## Residual-risk

The family manager's `ConfigDir` is used as the private native
`CLAUDE_CONFIG_DIR` only for the explicitly supplied family binding. The
production caller still has to construct that binding after verifying the
factory, receipt, codec, and launcher lifecycle contracts. Until that caller
exists and is tested through `moai gpt`, the end-to-end gateway goal remains
incomplete.
