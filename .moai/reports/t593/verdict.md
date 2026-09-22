# t593 Verdict — Claude Code binary pin (GH #1697)

## Claim

1. A configured pin (`MOAI_CLAUDE_BIN` env var, or the `llm.claude_bin` config key in `.moai/config/sections/llm.yaml`) is what the launcher hands to the exec seam — not a PATH lookup. Resolution order: env var > config key > `exec.LookPath("claude")` (fallback unchanged).
2. With no pin on either surface, behavior is byte-identical to pre-change (PATH-resolved `claude`).
3. A configured-but-invalid pin (missing path / directory / non-executable) fails the launch with an actionable error and never silently falls back to PATH.

## Evidence

All commands run in this worktree, Go 1.26.4, darwin/arm64.

### New + adjacent launcher tests (verbatim)

```
$ go test ./internal/cli/ -run 'TestValidateClaudeBinaryPin|TestResolveLaunchClaudeBinary|TestLaunchClaudeDefault_LaunchesPinnedBinary|TestLaunchClaudeDefault_PinErrorBlocksLaunch|TestLaunchClaudeDefault_UnsetPinFallsBackToPATH|TestExecOrSpawnClaude_PosixBuildTagGate' -count=1 -v
=== RUN   TestValidateClaudeBinaryPin_AcceptsExecutable
--- PASS: TestValidateClaudeBinaryPin_AcceptsExecutable (0.00s)
=== RUN   TestValidateClaudeBinaryPin_MissingPath
--- PASS: TestValidateClaudeBinaryPin_MissingPath (0.00s)
=== RUN   TestValidateClaudeBinaryPin_NotExecutable
--- PASS: TestValidateClaudeBinaryPin_NotExecutable (0.00s)
=== RUN   TestValidateClaudeBinaryPin_Directory
--- PASS: TestValidateClaudeBinaryPin_Directory (0.00s)
=== RUN   TestResolveLaunchClaudeBinary_EnvPinWinsOverPATH
--- PASS: TestResolveLaunchClaudeBinary_EnvPinWinsOverPATH (0.00s)
=== RUN   TestResolveLaunchClaudeBinary_ConfigPinWinsOverPATH
--- PASS: TestResolveLaunchClaudeBinary_ConfigPinWinsOverPATH (0.00s)
=== RUN   TestResolveLaunchClaudeBinary_EnvOverridesConfig
--- PASS: TestResolveLaunchClaudeBinary_EnvOverridesConfig (0.00s)
=== RUN   TestResolveLaunchClaudeBinary_UnsetFallsBackToPATH
--- PASS: TestResolveLaunchClaudeBinary_UnsetFallsBackToPATH (0.00s)
=== RUN   TestResolveLaunchClaudeBinary_MissingPinFailsLoud
--- PASS: TestResolveLaunchClaudeBinary_MissingPinFailsLoud (0.00s)
=== RUN   TestLaunchClaudeDefault_LaunchesPinnedBinary
--- PASS: TestLaunchClaudeDefault_LaunchesPinnedBinary (0.00s)
=== RUN   TestLaunchClaudeDefault_PinErrorBlocksLaunch
--- PASS: TestLaunchClaudeDefault_PinErrorBlocksLaunch (0.00s)
=== RUN   TestLaunchClaudeDefault_UnsetPinFallsBackToPATH
--- PASS: TestLaunchClaudeDefault_UnsetPinFallsBackToPATH (0.00s)
=== RUN   TestExecOrSpawnClaude_PosixBuildTagGate
--- PASS: TestExecOrSpawnClaude_PosixBuildTagGate (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.829s
```

The core proof is `TestLaunchClaudeDefault_LaunchesPinnedBinary`: it runs `launchClaudeDefault` end-to-end in a `t.TempDir()` fake project with PATH pointed at a directory holding NO `claude`, pins `MOAI_CLAUDE_BIN` to a fake executable, and captures what the launch hands to `execOrSpawnClaudeFunc` (the exec seam; on POSIX the production default is `syscall.Exec`, which cannot run inside a test). Captured binary == the pin. `TestLaunchClaudeDefault_UnsetPinFallsBackToPATH` is the regression guard: no pin → captured binary == the PATH-resolved fake `claude`.

### Broader launch-seam regression filter (verbatim)

```
$ go test ./internal/cli/ -run 'Launch|Claude|Glm|Exec' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	30.301s
```

### Touched config package, full (verbatim)

```
$ go test ./internal/config/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	2.275s
```

### Template guards (verbatim)

```
$ go test ./internal/template/ -run 'Leak|Neutrality' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	1.410s
$ go test ./internal/template/ -run 'TestEmbedded|TestLLM|TestProfileMatrix|TestModelPolicy|TestGLMEffort' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.253s
```

Shipped-key anti-rot guard (REQ-CKH-008) — passes with the new `llm.claude_bin` inventory entry (W/reader); the key classifies direct-live (production reader in internal/cli/claude_binary.go):

```
$ go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	1.000s
```

### Static checks

```
$ go build ./internal/cli/ ./internal/config/ && go vet ./internal/cli/ ./internal/config/   → exit 0
$ gofmt -l <touched .go files>                                                              → no output
```

### Template embed regeneration

```
$ make build   → exit 0 (agents-emit-check passed read-only; catalog.yaml regenerated with no diff; binary rebuilt with embedded templates)
```

## Baseline-attribution

- Worktree: `.claude/worktrees/agent-a5da1905f2b02bc0c` (isolated, runtime-provisioned), branch `WT-claude-binary-pin`.
- All measurements above taken in this run against base commit `04de513e4` + the working-tree changes listed below (tree was dirty during measurement — pre-commit state).
- Files changed: `internal/cli/claude_binary.go` (new), `internal/cli/claude_binary_test.go` (new), `internal/cli/launcher.go`, `internal/cli/launch_exec_test.go`, `internal/config/envkeys.go` (`EnvClaudeBin` = `MOAI_CLAUDE_BIN`), `internal/config/types.go` (`LLMConfig.ClaudeBin`, yaml `claude_bin,omitempty`), `internal/config/testdata/shipped_key_inventory.yaml` (+`llm.claude_bin`, W/reader), `internal/template/templates/.moai/config/sections/llm.yaml` (shipped key + guidance comment).

## Gaps

- README (4-locale sync obligation) and docs-site NOT updated — deliberate per card scope; doc follow-up owed: document `MOAI_CLAUDE_BIN` / `llm.claude_bin` in user docs.
- Full test suite NOT run locally (load discipline; CI owns the whole-suite verdict). internal/cli was covered by a targeted filter (`Launch|Claude|Glm|Exec`), not the entire package.
- No real-process launch verified: the end-to-end proof captures the exec seam because a production POSIX launch `syscall.Exec`s and would replace the test process. Windows spawn path is compile-checked only (no windows run in this environment).
- CHANGELOG not touched (sync-phase concern).
- The config-key surface is read from the resolved project root; launching outside any `.moai` project skips the config key (env var still honored) — config is per-project by construction.

## Residual-risk

- Existing deployments keep working unchanged (no shipped key → PATH fallback); the key reaches users only via the next template redeploy.
- On Windows the executable-bit check degrades to existence + not-a-directory (documented in code); a pin pointing at a non-executable file type would fail later at process spawn, not at pin validation.
- A pin that later breaks (binary deleted/upgraded away) now fails the launch loudly by design — an operator with a stale pin must clear or fix it; the error names both configuration surfaces and the fallback.
- `saveLLMSection` round-trips the whole LLMConfig on a future team-mode change; `claude_bin` survives the round-trip (omitempty preserves non-empty values), but a rewrite drops the template's hand-written comments — pre-existing persist-path behavior, unchanged by this card.
