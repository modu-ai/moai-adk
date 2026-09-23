# Research — SPEC-DUAL-HARNESS-HOOK-PARITY-001

All measurements in §R1 were taken by manager-spec on 2026-09-23 in the worktree
`.claude/worktrees/dual-harness-parity-rebuild`, branch `WT-dual-harness-parity-rebuild`,
HEAD `530d8cc067765a3cf6ac49a76954d34a2193c7d0`. Figures from the handoff were re-measured, not
carried over. Static readings are labelled "static"; they establish what the code says, not what a
host does at runtime.

## R0. Baseline

```text
$ git rev-parse HEAD
530d8cc067765a3cf6ac49a76954d34a2193c7d0
$ git branch --show-current
WT-dual-harness-parity-rebuild
$ git rev-list --count --left-right origin/develop...HEAD
0	25
$ codex --version
codex-cli 0.155.1
$ claude --version
2.1.280 (Claude Code)
$ go version
go version go1.26.8 darwin/arm64
$ go test ./internal/codexadapter/... ./internal/codexwiring/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/codexadapter	0.619s
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.657s
```

## R1. Re-measured anchor facts

### R1.1 Event table — 8 of 12 adapted (handoff claim confirmed)

```text
$ grep -c '", true}' internal/codexadapter/events.go
8
$ grep -c '", false}' internal/codexadapter/events.go
4
```

Adapted: PreToolUse, PostToolUse, SessionStart, SessionEnd, Stop, UserPromptSubmit,
SubagentStart, SubagentStop. Held back: PreCompact (`compact`), PostCompact (`post-compact`),
PermissionRequest (`permission-request`), Interrupt (no dispatcher arg). The held-back reasons in
the `events.go` doc comment are trigger-not-achieved (compaction, approval request) and
no-dispatcher-counterpart (Interrupt), on the codex-cli 0.153.4 basis.

### R1.2 Claude Stop chain — 8 handlers (static)

```text
$ python3 (parse internal/template/templates/.claude/settings.json.tmpl, collect hooks/moai/*.sh per event)
Stop ['handle-stop.sh', 'sync-phase-quality-gate.sh', 'handle-stop-goal.sh', 'handle-security-turn.sh',
      'handle-security-commit.sh', 'handle-codex-review-gate.sh', 'handle-multi-review-gate.sh',
      'handle-harness-observe-stop.sh']
PreCompact ['handle-compact.sh']
PostCompact ['handle-post-compact.sh']
PermissionRequest ['handle-permission-request.sh']
```

What each Stop wrapper invokes (static grep of the wrapper bodies):

| Wrapper | Invokes | Claude timeout | Proposed class |
|---|---|---|---|
| handle-stop.sh | `moai hook stop` | 5 | advisory (plus factory continuation) |
| sync-phase-quality-gate.sh | pure shell, 778 lines; runs compile/vet; blocking by default, `MOAI_SYNC_GATE_BLOCKING` opt-out | 60 | required-gate |
| handle-stop-goal.sh | `moai hook stop-goal` | 120 | goal |
| handle-security-turn.sh | `moai hook security-turn` (async) | 5 | advisory |
| handle-security-commit.sh | `moai hook security-commit` (async) | 5 | advisory |
| handle-codex-review-gate.sh | `moai hook codex-review-gate` | — | required-gate (config-conditional) |
| handle-multi-review-gate.sh | `moai hook multi-review-gate` | — | required-gate (config-conditional) |
| handle-harness-observe-stop.sh | `moai hook harness-observe-stop` | — | advisory |

The class column is a proposal for REQ-HPR-001; run-phase M2d confirms it against each member's
actual blocking behaviour.

### R1.3 Codex Stop path is advisory-allow (handoff claim confirmed, static)

```text
$ grep -n "if !row.Adapted" internal/codexwiring/hooks.go
98:		if !row.Adapted {
$ grep -n '"moai hook " + row.DispatcherArg' internal/codexwiring/hooks.go
104:				Command: "moai hook " + row.DispatcherArg + harnessCodexSuffix,
$ grep -n "return &HookOutput{}, nil\|factoryHookBatch\|runEvidenceGate" internal/hook/stop.go
46:		return &HookOutput{}, nil
79:	runEvidenceGate(projectDir, input.SessionID)
80:	if factoryCtx, shouldContinue, _ := factoryHookBatch(ctx, input, EventStop); shouldContinue {
87:	return &HookOutput{}, nil
$ printf '%s' '{}' | go run ./cmd/moai hook stop ; echo "exit=$?"
{}
exit=0
```

`RenderHooks` emits exactly one handler per adapted row, so Codex Stop gets only
`moai hook stop --harness codex`. The stop handler returns `{}` on every path except a factory
continuation; its evidence gate is advisory-only by its own comment. None of stop-goal, the sync
gate, or the two review gates is registered on Codex.

### R1.4 PreToolUse `ask` is dropped to no-opinion (static)

```text
$ grep -n '"ask": *true' internal/codexadapter/output.go
40:	"ask":   true,
```

`preToolUseDropDecisions` drops `allow` (without updatedInput), `ask`, and `defer`; the
`normalizePreToolUseDecision` comment states the drop "hands the choice to Codex's own approval
flow". **Hypothesis (not measured):** under an approval policy that never prompts, that hand-off
resolves as allow, which would violate REQ-HPR-007. AC-HPR-007 measures it.

### R1.5 Timeout budgets (static)

```text
$ grep -n "defaultHandlerTimeout = \|sessionEndTimeoutCeiling = " internal/codexwiring/codexwiring.go
56:	sessionEndTimeoutCeiling = 3
60:	defaultHandlerTimeout = 10
$ grep -n "stopGoalHookTimeout = " internal/cli/hook_stop_goal.go
20:const stopGoalHookTimeout = 90 * time.Second
```

Codex handlers are registered with a 10-second timeout; the goal evaluator allows each mechanical
condition 90 seconds and the Claude wrapper 120. Registering stop-goal on Codex unchanged would let
Codex kill it mid-evaluation. This is the concrete basis for REQ-HPR-018/019 and design.md §D3.

### R1.6 No obligation registry exists

```text
$ ls internal/template/catalog
ls: internal/template/catalog: No such file or directory
$ ls .moai/policies .moai/workflows
ls: .moai/policies: No such file or directory
ls: .moai/workflows: No such file or directory
```

`internal/template/templates/.moai/policies/` holds only a README. `features.yaml` from design §07
does not exist. AC-POL-01 therefore needs a registry built by this SPEC (design.md §D4).

### R1.7 Measured-basis drift

The adapter's measurement basis is codex-cli 0.147.0 / 0.153.4 (events.go header); the installed
binary is 0.155.1 (R0). The t496 campaign record cited by events.go is absent from this tree:

```text
$ ls .moai/reports/t496/
ls: .moai/reports/t496/: No such file or directory
```

Every live verdict in this SPEC is re-measured on the installed version.

### R1.8 gpt profile hides `.claude/` (static)

```text
internal/template/harness_fs.go:112:	if h.hideClaude && (name == ".claude" || strings.HasPrefix(name, ".claude/")) {
```

Under the `gpt` profile the 778-line `sync-phase-quality-gate.sh` is not deployed, so it cannot be
the Codex path for the required gate (REQ-HPR-003).

### R1.9 Reusable receipt primitives (static)

```text
$ grep -n "func Key" internal/verify/key.go
39:func Key(ctx context.Context, repoDir string) (string, error) {
$ grep -n "DefaultTTL = " internal/verify/freshness.go
7:const DefaultTTL = 10 * time.Minute
```

`verify.Key` already binds HEAD + porcelain-v2 digest + `git diff HEAD` hash + untracked-content
hash. The sync gate already keeps `<head-sha> <outcome> <worktree-content-id>` in
`.moai/state/sync-quality-gate.last`. Both cover two of the five receipt fields (HEAD, tree
digest); configuration digest, command, and tool version are missing.

### R1.10 Goal status vocabulary (static)

`internal/goal/schema.go` defines `armed`, `satisfied`, `ceiling-exit`, `cleared`, `unsatisfiable`.
There is no cancellation status distinct from `cleared` (explicit `moai goal clear`).

## R2. Existing live-test conventions

Live Codex tests gate on env switches (`MOAI_CODEX_LIVE_PROBE`, `MOAI_CODEX_LIVE_BIN`,
`MOAI_SKIP_LIVE_CODEX`) and are catalogued by `internal/cli/codex_live_axis_declaration_test.go`.
They `t.Skip` when unset — which is exactly why AC rule P rejects a skipped run as evidence. The
status report records four factory live tests that exited 0 while all four were `SKIP`.

## R3. Open questions carried to plan.md

Hypotheses that only run-phase measurement can settle:

- H1: Codex resolves the no-opinion PreToolUse output as allow under a non-prompting policy (R1.4).
- H2: Codex treats a hook timeout / non-contract exit on PreToolUse as proceed (fail-open).
- H3: Codex merges multiple handlers on one event with "any block wins".
- H4: PreCompact / PostCompact / PermissionRequest can be triggered on 0.155.1 in some run mode.
