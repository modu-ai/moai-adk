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
| handle-codex-review-gate.sh | `moai hook codex-review-gate` | 900 | required-gate (config-conditional) |
| handle-multi-review-gate.sh | `moai hook multi-review-gate` | 900 | required-gate (config-conditional) |
| handle-harness-observe-stop.sh | `moai hook harness-observe-stop` (inside `{{ if .HookOptIn.Enabled }}`) | 5 | advisory |

Correction (v0.2.0, plan-audit iter-1 D2): the v0.1.0 table recorded "—" for the last three
timeouts. Re-measured, the Stop array's timeouts in order are:

```text
$ python3 (parse the Stop block of settings.json.tmpl, collect "timeout" values in order)
['5', '60', '120', '5', '5', '900', '900', '5']
```

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
condition 90 seconds and the Claude wrapper 120. **Correction (v0.2.0, D2):** the 10 s is MoAI's
own render constant, not a host limit. `codexwiring.go:58–60` reads "defaultHandlerTimeout is the
table constant for every non-SessionEnd handler (D1; matches the t83 observed-behavior case)", and
`moaiHandlerTimeout` (`hooks.go:33–39`) already renders SessionEnd differently. The largest Stop
timeout Codex accepts is unmeasured (§R4). Registering stop-goal on Codex unchanged would let
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
There is no cancellation status distinct from `cleared`.

`StatusCleared` is defined and read but never written by non-test code, and `moai goal clear`
removes the goal file instead of writing a status:

```text
$ grep -rn "StatusCleared" internal cmd pkg | grep -v "_test.go"
internal/goal/schema.go:72:	// StatusCleared: the goal was explicitly cleared (moai goal clear).
internal/goal/schema.go:73:	StatusCleared Status = "cleared"
internal/goal/evaluate.go:294:	if g == nil || g.Status == StatusCleared || g.Status == StatusSatisfied {
$ grep -n "func ClearGoal\|os.Remove" internal/goal/state.go
120:func ClearGoal(projectRoot, sessionID string) error {
127:		if err := os.Remove(p); err != nil {
```

The three non-test lines are two definition lines and one reader (`evaluate.go:294`), not three
readers.

Correction (v0.3.0, plan-audit iter-2 N3): the v0.2.0 grep was scoped to `internal/goal internal/cli`
and missed three sites in `internal/hook`. Re-run over all three packages:

```text
$ grep -rn "goal\.Status\|g\.Status\|StatusArmed\|StatusSatisfied\|StatusCeilingExit\|StatusUnsatisfiable\|StatusCleared\|ClearGoal(" internal/goal internal/cli internal/hook | grep -v _test.go
(schema.go:65–78 and :130, dashboard.go:141, state.go:120, evaluate.go:294/306/325/340/404/435 as before, plus:)
internal/cli/launcher_blockcap_infinite.go:68:	if g.Status != goal.StatusArmed || g.Ceiling.MaxTurns != 0 {
internal/cli/handoff.go:85:	... g.Status == goalpkg.StatusArmed {
internal/cli/goal.go:1282:	... g.SessionID, g.Status, g.Goal)
internal/cli/goal.go:1313:	if err := goal.ClearGoal(root, sessionID); err != nil {
internal/cli/goal.go:1331:	_, _ = fmt.Fprintf(out, "status:     %s\n", g.Status)
internal/hook/session_start_compact.go:88:	if g.Status != goal.StatusArmed {
internal/hook/handoff_inject.go:185:		Status:          goal.StatusArmed,
internal/hook/stop_failure.go:111:	if err := goal.ClearGoal(root, input.SessionID); err != nil {
```

The same grep also hit four `binlag.Status*` cases in `internal/cli/doctor.go:606–623`. Those are a
different type and are not goal-status sites. Goal-status sites:

| Location | Use |
|---|---|
| `internal/goal/evaluate.go:294` | early return for cleared/satisfied |
| `internal/goal/evaluate.go:306, 325, 340, 404, 435` | writes ceiling-exit / unsatisfiable / satisfied |
| `internal/goal/schema.go:130` | new goal starts `armed` |
| `internal/goal/dashboard.go:141` | dashboard status string |
| `internal/cli/launcher_blockcap_infinite.go:68` | acts only on `armed` |
| `internal/cli/handoff.go:85` | embeds only an `armed` goal |
| `internal/cli/goal.go:1282, 1331` | `goal status` output |
| `internal/cli/goal.go:1313` | `moai goal clear` → `ClearGoal` |
| `internal/hook/session_start_compact.go:88` | reader, acts only on `armed` |
| `internal/hook/handoff_inject.go:185` | writer of `armed` |
| `internal/hook/stop_failure.go:111` | `ClearGoal` on unrecoverable StopFailure |

design.md §D6 assigns the `cancelled` handling for each row.

### R1.11 Review-gate budgets and the receipt-shaped gate (static)

```text
$ grep -n -i 'timeout' internal/cli/codex_review_gate.go
83:	// The 900s override (config.DefaultCodexReviewGateTimeout) is pinned in the
87:	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultCodexReviewGateTimeout)
$ grep -rn "DefaultCodexReviewGateTimeout *=" internal/config/
internal/config/defaults.go:436:var DefaultCodexReviewGateTimeout = 900 * time.Second
$ grep -n "^func \|audit-multi" internal/cli/multi_review_gate.go
65:func HandleMultiReviewGate(input *hook.HookInput, enabled bool, projectDir, sessionID string) (*hook.HookOutput, error) {
113:func loadConvergenceResult(projectDir, sessionID string) (ConvergenceResult, bool) {
117:	path := filepath.Join(projectDir, ".moai", "state", "audit-multi", sessionID+".json")
```

The codex review gate runs a live RPC of up to 900 s inside the hook. The multi-review gate only
reads a result the `audit_multi` MCP tool persisted during the turn, so it is receipt-shaped today.

### R1.12 Non-Stop multi-handler chains (static)

```text
$ python3 (parse settings.json.tmpl, collect hooks/moai/*.sh per event)
PreToolUse ['handle-pre-tool.sh', 'handle-pre-tool.sh', 'handle-pre-tool.sh', 'handle-pre-tool.sh']
PostToolUse ['handle-post-tool.sh', 'status-transition-ownership.sh', 'status-transition-ownership.sh', 'status-transition-ownership.sh', 'handle-harness-observe.sh']
SessionStart ['handle-session-start.sh', 'handle-session-start-navigator.sh']
SubagentStop ['handle-subagent-stop.sh', 'handle-harness-observe-subagent-stop.sh', 'chain-event.sh']
UserPromptSubmit ['handle-user-prompt-submit.sh', 'handle-harness-observe-user-prompt-submit.sh']
```

These chains are excluded from this SPEC (spec.md §F) and registered as `unverified`.

### R1.13 Tests that pin behaviour this SPEC changes (static)

- `internal/codexadapter/events_test.go:67` `TestAdaptedRowCount` pins `wantAdapted = 8`. M2e sets
  it to **12**: PreCompact, PostCompact, and PermissionRequest are adapted, and Interrupt gains the
  dispatcher arg `interrupt` (design.md §D7).
- `internal/codexwiring/hooks_test.go:73` `TestRenderHooks_InterruptNeverInstalled` asserts the
  rendered `hooks.json` carries no Interrupt key (SPEC-CODEX-EVENT-COVERAGE-001 REQ-CEV-004 /
  AC-CEV-005). Once Interrupt is adapted, `RenderHooks` installs it, so M2e inverts this test to
  assert the Interrupt handler **is** rendered.
- `internal/codexadapter/output_test.go:278` `TestPreToolUseAskDropped` asserts `ask` → `{}`; M2c
  inverts it.

Added in v0.4.0 (plan-audit iter-3 R3), from a grep of every `_test.go` under
`internal/codexadapter`, `internal/codexwiring`, `internal/cli`, and `internal/template` for
`Interrupt|PreCompact|PostCompact|PermissionRequest|permission-request|post-compact|wantAdapted|Unadapted`:

- `internal/codexadapter/events_test.go:24–42` `TestEventTableMapping` pins `{"compact", false}`,
  `{"post-compact", false}`, `{"permission-request", false}` (`:37–39`) and
  `"Interrupt": {"", false}` (`:42`) — SPEC-CODEX-EVENT-COVERAGE-001 REQ-CEV-001. M2e rewrites the
  four rows as adapted, with Interrupt carrying `interrupt`.
- `events_test.go:111` `TestResolveRecognizedButUnadapted` iterates PreCompact, PostCompact,
  PermissionRequest, and Interrupt (`:114`) expecting refusal — REQ-CEV-003. M2e rewrites it; the
  unknown-vs-unadapted distinction it guards is kept by `TestResolveUnknownEvent` (`:158`).
- `events_test.go:134` `TestResolveInterruptNoCounterpart` expects Interrupt to be refused with no
  dispatcher-arg claim — REQ-CEV-003. M2e rewrites it to expect `interrupt`.
- `internal/cli/hook_harness_codex_test.go:305` `TestHarnessCodexUnadaptedSubcommandRejected` uses
  `compact` as its unadapted example. M2e retargets or rewrites it.
- Conditional: `internal/codexadapter/stderr_test.go:78` `TestExcludedEventsHaveNoClass` pins no
  stderr classification for PreCompact, PostCompact, and PermissionRequest (`:82–84`). It is
  amended only if M2e gives those events a stderr class.

Examined and not amended:

- `internal/codexadapter/dispatcher_registration_test.go` skips rows with an empty `DispatcherArg`
  (`:40–47`). Once Interrupt carries `interrupt`, the test requires that subcommand to be
  registered in `internal/cli/hook.go` — a check the M2e `moai hook interrupt` subcommand
  satisfies, not an amendment.
- The remaining grep hits (`hook_e2e_test.go`, `hook_test.go`, `codex_job_control_test.go`,
  `codex_live_protocol_probe_test.go`, `preference/m4_crash_repro_test.go`,
  `internal/template/settings_test.go`) concern the Claude-side dispatcher, the codex app-server
  `turn/interrupt` RPC, or the Claude settings template, none of which this SPEC changes.

All listed amendments are intentional (plan.md M2c, M2e), not regressions. Adapting the four rows
reverses SPEC-CODEX-EVENT-COVERAGE-001 REQ-CEV-001, REQ-CEV-003, and REQ-CEV-004 (spec.md §D).

## R2. Existing live-test conventions

Live Codex tests gate on env switches (`MOAI_CODEX_LIVE_PROBE`, `MOAI_CODEX_LIVE_BIN`,
`MOAI_SKIP_LIVE_CODEX`) and are catalogued by `internal/cli/codex_live_axis_declaration_test.go`.
They `t.Skip` when unset — which is exactly why AC rule P rejects a skipped run as evidence. The
status report records four factory live tests that exited 0 while all four were `SKIP`.

## R4. Unmeasured: Codex's largest accepted Stop timeout

No measurement exists in this tree of the largest Stop hook timeout Codex accepts and honours. The
value MoAI renders today is its own constant (R1.5). AC-HPR-021 defines the probe. Under operator
decision Q5 (no live runs in this SPEC) it stays `NOT_RUN`, and `T_stop` stays at 10 s
(design.md §D3.5).

## R3. Open questions carried to plan.md

Hypotheses that only run-phase measurement can settle:

- H1: Codex resolves the no-opinion PreToolUse output as allow under a non-prompting policy (R1.4).
- H2: Codex treats a hook timeout / non-contract exit on PreToolUse as proceed (fail-open).
- H3: Codex merges multiple handlers on one event with "any block wins".
- H4: PreCompact / PostCompact / PermissionRequest can be triggered on 0.155.1 in some run mode.
