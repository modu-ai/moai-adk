# SPEC-HARNESS-LEARNING-EVO-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` authored at `status: draft`, Tier M.
- Scope: **L1 only** (instrumentation repair). The L2 analyzer was split out to `SPEC-HARNESS-LEARNING-EVO-002`; L3 (application to `delegation.yaml`) remains an explicit non-goal with its three-surface rationale.
- Split rationale: the v0.1.0 SPEC carried 33 requirements and 36 acceptance criteria, over the ceiling at Tier M (16/16) and Tier L (25/25). Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, exceeding either ceiling is a signal to split, not to relax the budget.
- Requirements: 16 GEARS requirements (REQ-HLE-001..016); 16 acceptance criteria with 100% requirement coverage (`acceptance.md` §G). Both counts are exactly at the Tier M ceiling.
- Material changes on new measurement: agent identity now reads `agent_type` rather than the derived `subject` (`spec.md` §A.5), and the terminal-signal source changed after the root cause of the empty signal population was determined to be hook registration (`spec.md` §A.6, `plan.md` §E D2).
- Open decisions carried into audit: the `matched_subcommand` write policy (`plan.md` §E D1 — first-writer-wins) and the terminal-signal source (`plan.md` §E D2 — option (c), gated by the M0 probe with option (a) as the declared fallback).

_<pending plan-audit>_

## §E.2 Run-phase Evidence

### M0 — the terminal-signal probe (plan.md §E D2 decision gate)

**Verdict: option (c). The premise held; no `settings.json` or template edit was needed.**

The primary checkout's wrapper could not be instrumented (that would write outside this worktree), so the probe used a throwaway project with a `.claude/settings.json` registering a single PostToolUse entry with **no matcher** — structurally identical to how `handle-harness-observe.sh` is registered — pointing at a capture script that appends raw stdin to a file.

Command (Claude Code 2.1.226):

```
claude -p "Run exactly this shell command with the Bash tool and nothing else: echo M0PROBE-OK" \
  --permission-mode acceptEdits --model claude-haiku-4-5-20251001
```

Observed payload, verbatim (one JSONL line, pretty-printed here):

```json
{
  "session_id": "f7a781f1-728c-478e-aa35-bdce34fb406c",
  "cwd": "/private/tmp/.../m0probe",
  "permission_mode": "acceptEdits",
  "hook_event_name": "PostToolUse",
  "tool_name": "Bash",
  "tool_input": { "command": "echo M0PROBE-OK", "description": "Run the specified echo command" },
  "tool_response": { "stdout": "M0PROBE-OK", "stderr": "", "interrupted": false, "isImage": false, "noOutputExpected": false },
  "tool_use_id": "toolu_01NpG6muqBX19PUbqpookSFz",
  "duration_ms": 462
}
```

Both fields the gate asks about — `tool_input.command` and `tool_response` — are present and non-empty. `runHarnessObserve` discards them at the **handler**, exactly as plan.md §E D2's refutation of option (b) states; the wrapper does receive them.

Two observations the plan did not anticipate, recorded because they change what the code can rely on:

1. **The payload arrives flat snake_case, not nested camelCase.** `normalizeHookInput` returns unchanged when `session_id` is present, so this is the already-flat pass-through path. plan.md §A.2's implication that camelCase is the native shape does not match 2.1.226 on this channel. The observed shape is pinned in `internal/cli/hook_routing_ledger_test.go` `bashPostToolPayload`.
2. **`tool_response` carries no `exit_code`.** A second probe on a failing in-workspace command (`grep NOPE_ZZZ data.txt`) returned `{"stdout":"","stderr":"","interrupted":false,"isImage":false,"returnCodeInterpretation":"No matches found","noOutputExpected":false}`. `classifyTestCommand` tries `deriveFromExitCode` first and falls back to `deriveFromOutputText`; with no exit-code key the structured path cannot fire on this channel, so pass/fail rests on the output-text heuristic. This does not block option (c) — REQ-HLE-010's "absence appends nothing" already covers the ambiguous case — but it is recorded as residual risk R7 below.

Third probe, recorded so a later reader does not misread an empty capture as a falsified premise: a Bash call **blocked by permissions** fires no PostToolUse at all (the capture file was never created).

### Milestone commits

| Milestone | Commit | Content |
|---|---|---|
| M1 + M2 | `038bcc4fc` | `RecordIfAbsent` / `Annotate` / `RoutingPatch`, identity + outcome markers, `ledger annotate` verb |
| M3 | `631c9aa1c` | `hook.LogBashEvidence` + the Bash carve-out in `runHarnessObserve` |
| M4 | `637f9a65f` | the three seams, gate deduplication into package `hook`, handler wiring |

### AC PASS/FAIL matrix

Every row was decided by running its own command in this worktree at `637f9a65f`. `ok <pkg> <t>` is the verbatim `go test` tail.

| AC | Status | Deciding command | Observed |
|---|---|---|---|
| AC-HLE-001 | PASS | `go test ./internal/harness/routing/ -run 'TestRecordIfAbsent_Lifecycle\|TestRecord_StillReroutesSelf' -count=1` | `ok …/internal/harness/routing 0.344s` |
| AC-HLE-002 | PASS | `go test ./internal/harness/routing/ -run TestRecordIfAbsent_NoSweepOnCreatePath -count=1` ; `go test ./internal/telemetry/ -run TestLoadBySession_TwoDayWindow -count=1` | `ok …/routing 0.345s` ; `ok …/telemetry 0.343s` |
| AC-HLE-003 | PASS | `go test ./internal/harness/routing/ -run TestSweepStale_AgeAndLivenessGuards -count=1` | `ok …/routing 0.329s` |
| AC-HLE-004 | PASS | `go test ./internal/harness/routing/ -run TestAnnotate -count=1` ; `go test ./internal/cli/ -run TestHarnessLedgerAnnotateCmd -count=1` | `ok …/routing 0.335s` ; `ok …/cli 1.420s` |
| AC-HLE-005 | PASS | `go test ./internal/harness/routing/ -run 'TestSchemaVersionStable\|TestReader' -count=1` | `ok …/routing 0.343s` |
| AC-HLE-006 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_UserPromptSubmit_CreatesPending -count=1` | `ok …/hook 0.570s` |
| AC-HLE-007 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_NoRawPromptPersisted -count=1` | `ok …/hook 0.760s` |
| AC-HLE-008 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_MultiTurnNoReroute -count=1` | `ok …/hook 0.653s` |
| AC-HLE-009 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_LiteralSubcommandFirstWriterWins -count=1` | `ok …/hook 0.676s` |
| AC-HLE-010 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_SubagentStop_AgentTypeVerbatim -count=1` | `ok …/hook 0.671s` |
| AC-HLE-011 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_AbsentAgentTypeMarker -count=1` | `ok …/hook 0.745s` |
| AC-HLE-012 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_UnknownOutcomeNotSuccess -count=1` | `ok …/hook 0.575s` |
| AC-HLE-013 | PASS | `go test ./internal/cli/ -run TestHarnessObserve_BashEvidenceRecorded -count=1` ; `go test ./internal/hook/ -run TestEvidenceRecord_NoDoubleWriteOnEdit -count=1` | `ok …/cli 0.947s` ; `ok …/hook 0.622s` |
| AC-HLE-014 | PASS (bounded — see R1) | `go test ./internal/cli/ -run TestRoutingLedger_TerminalCloseEndToEnd -count=1` | `ok …/cli 1.091s` |
| AC-HLE-015 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_FailOpen -count=1` | `ok …/hook 0.651s` |
| AC-HLE-016 | PASS | `go test ./internal/hook/ -run TestRoutingSeam_GatedOffWritesNothing -count=1` | `ok …/hook 0.648s` |

**AC-HLE-014 is bounded, not unconditional.** It drives the production handler functions against a temp root and proves the seam sequence closes a row end to end. It does NOT prove Claude Code invokes those handlers in a live session — the R1 gap in `acceptance.md` §F. That link is unverified here; see the Gaps section below.

### §D quality gate

| Check | Command | Observed |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Full suite | `go test ./... -count=1` | exit 0, 0 `FAIL` lines |
| Vet | `go vet ./...` | exit 0, no output |
| Lint | `golangci-lint run --timeout=6m` | `0 issues.` (pre-change baseline was also `0 issues.`) |
| SPEC lint | `moai spec lint …/spec.md --strict` | `✓ No findings — all SPEC documents are valid` |
| No template / frozen-skill / frozen-allowlist path | `git diff --name-only 5a929480a...HEAD \| grep -cE '^(internal/template/templates/\|\.claude/skills/\|\.claude/lsel/)'` | `0` |
| No delegation.yaml reader or writer | `grep -rn 'delegation.yaml' internal/ --include='*.go' \| grep -v _test` | no output |
| No AskUserQuestion in the seam file | `grep -rn 'AskUserQuestion' internal/hook/routing_ledger.go` | no output |

Coverage on the touched packages: `internal/harness/routing` 88.5%, `internal/hook` 83.9%, `internal/telemetry` 82.7%, `internal/cli` 76.5%. The first three clear the 85% package target or sit near it; `internal/cli` is a large pre-existing package whose baseline this SPEC did not move materially.

`go test -race ./internal/harness/routing/ ./internal/telemetry/ -count=1` also passes, which is the meaningful race surface here — the new store operations are the only concurrency-adjacent code this SPEC adds.

### One observed failure, diagnosed and attributed

`internal/cli/wizard` `TestUnifiedForm_ManualModeSkipsConditionals` failed **once**, during the multi-package coverage run (`go test -cover ./internal/harness/... ./internal/hook/... ./internal/cli/...`). It is not attributable to this SPEC, and the reason is structural rather than a judgement call:

- It does not reproduce. `-count=5`, `-count=20`, and `-race` on the package all pass, as does the package alone under `-cover`.
- The package is `internal/cli/wizard`, which this SPEC does not touch and which does not reach the routing seams.
- The mechanism is a fixed wall-clock deadline in the test's own driver: `execBoundedCmd` (`internal/cli/wizard/unified_form_test.go:60`) waits **50 ms** for a bubbletea command and returns `nil` on timeout, silently dropping the message. Under the contention of a multi-package coverage run a command can exceed 50 ms, a state transition is lost, and the form never reaches `StateCompleted` — exactly the assertion that failed.
- The same multi-package coverage command run against the pre-SPEC base commit `5a929480a` (extracted via `git archive` into a scratch tree) also reported `FAIL github.com/modu-ai/moai-adk/internal/cli`, so the baseline is not clean under that load either.

The full suite without coverage instrumentation — `go test ./... -count=1`, the §D gate command — exits 0 with zero `FAIL` lines.

### Gaps — what was NOT verified

- **The live hook-dispatch link (R1).** No CI-runnable assertion covers "Claude Code actually invokes these handlers in a real session". The M0 probe proves the *payload* reaches a matcher-null PostToolUse wrapper; it does not exercise the UserPromptSubmit or SubagentStop wrappers, nor the seams behind them. The manual dogfood check named in `plan.md` §F M5 — enable the opt-in, run one real `/moai` dispatch with a subagent, read the resulting row — was **not performed**, because it requires enabling a fail-closed gate in the primary checkout, which is outside this worktree's write boundary.
- **A live `is_test_fail` signal end to end.** The failing direction is covered by unit tests driving the production path with a fixture, but no observed live run has produced `is_test_fail` through the restored carve-out.
- **The falsification review (`spec.md` §E F1/F2).** Needs 50 finalized rows, which do not exist at merge time. Scheduled, not performed — as `acceptance.md` §E already states.

### Residual risk

R1-R6 stand as written in `acceptance.md` §F. One is added by the M0 measurement:

- **R7 — the terminal signal rests on an output-text heuristic, not a structured exit code.** The observe channel's `tool_response` carries no `exit_code`, so `deriveFromExitCode` cannot fire and pass/fail detection falls to `deriveFromOutputText`'s marker matching (`ok  \t`, `--- FAIL`, `test result:`, and a `passed`/`failed` word count). A test runner whose output matches none of those yields no signal, the row stays pending, and it closes later as `abort` via the stale sweep rather than as `fail`. The live yield of terminal `success`/`fail` rows may therefore be lower than the code path suggests, and that will only be measurable once rows accumulate.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-10
run_commit_sha: 637f9a65f
run_status: audit-ready
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 15
m1_to_mN_commit_strategy: "four milestones in three commits — M1+M2 share 038bcc4fc because the identity markers live in the same file as the store types; M3 is 631c9aa1c; M4 is 637f9a65f. M0 produced no commit: it is a probe, and its result is recorded in §E.2 rather than in code."
```

Notes for the auditor:

- **M0 changed nothing in the plan.** The probe confirmed option (c), so the declared option (a) fallback — a `settings.json.tmpl` edit plus `make build`, plus the §25 neutrality re-check — was never entered. No template file, frozen skill, or frozen allowlist was touched.
- **One deviation from `plan.md` §I's file inventory.** The inventory anticipated edits to `internal/cli/hook_test.go` and `internal/telemetry/recorder_test.go`; the CLI-side tests landed in a new file `internal/cli/hook_routing_ledger_test.go` instead, because `hook_test.go` is a small unrelated file and appending an unrelated 300-line fixture set to it would have obscured both. The telemetry edit landed as planned.
- **One addition beyond the inventory**, made to avoid a defect rather than for tidiness: the two observation gates existed only in `internal/cli`, which `internal/hook` cannot import. Copying them would have created two truth tables that could drift — and a drift between them is exactly the class of failure this SPEC is repairing elsewhere. The implementations moved down into `internal/hook` (`HarnessLearningEnabled`, `HookObserveOptInEnabled`) and the CLI functions became thin delegating wrappers, so every existing call site and test is unchanged.
- **`Store.LoadPending` was exported**, which `plan.md` §I did not list. The first-writer-wins policy (REQ-HLE-006, §E D1) requires the seam to ask "is this field still empty?", and §E D1 explicitly places that policy at the seam rather than in the store — so the seam needs a read primitive. Baking the policy into `Annotate` was the alternative and was rejected for that reason.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
