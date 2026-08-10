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
| M1 + M2 | `710a735b6` | `RecordIfAbsent` / `Annotate` / `RoutingPatch`, identity + outcome markers, `ledger annotate` verb |
| M3 | `947cee983` | `hook.LogBashEvidence` + the Bash carve-out in `runHarnessObserve` |
| M4 | `7939997cb` | the three seams, gate deduplication into package `hook`, handler wiring |

### AC PASS/FAIL matrix

Every row was decided by running its own command in this worktree at `7939997cb`. `ok <pkg> <t>` is the verbatim `go test` tail.

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

### Two pre-existing statusline hangs — and a correction to the §D gate row

The §D table above records `go test ./... -count=1` as exit 0 with zero `FAIL` lines. **That was observed at `7939997cb`, and a later re-run at `b9202951c` — byte-identical code — came back exit 1.** Both results are real; the suite is non-deterministic on this machine. The correction and its attribution:

```
FULL_SUITE_EXIT=1
FAIL  github.com/modu-ai/moai-adk/internal/cli          601.215s
FAIL  github.com/modu-ai/moai-adk/internal/statusline   605.114s

panic: test timed out after 10m0s
    running tests:
        TestRunStatusline_NilDeps (9m50s)      <- internal/cli
        TestBuilder_Build_FullData (10m0s)     <- internal/statusline
```

`grep -E "^--- FAIL"` over the whole log returns nothing: **zero assertions failed.** Both are wall-clock hangs, and both hanging tests are statusline tests.

Both legs are proven pre-existing on the untouched base commit `5a929480a`, by two independent measurements:

- **Mine**, `git archive 5a929480a` into a scratch tree, per-package: `internal/cli` timed out with `TestRunStatusline_NilDeps (9m48s)`; `internal/statusline` `FAIL … 600.642s` with `TestBuilder_Build_FullData (10m0s)`.
- **The orchestrator's**, a dedicated baseline worktree (`.claude/worktrees/hle-baseline`), no coverage instrumentation, `go test ./internal/cli/... ./internal/statusline/...`:

```
FAIL	github.com/modu-ai/moai-adk/internal/cli          607.369s
FAIL	github.com/modu-ai/moai-adk/internal/statusline   608.341s

panic: test timed out after 10m0s
		TestRunStatusline_NilDeps (9m50s)       <- internal/cli
panic: test timed out after 10m0s
		TestBuilder_Build_FullData (10m0s)      <- internal/statusline
```

Evidence file: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/26bbfcfb-4b2e-4578-b5ef-8391eb56fac1/baseline.txt`.

| Test | At HEAD `b9202951c` | At base `5a929480a` |
|---|---|---|
| `internal/cli` `TestRunStatusline_NilDeps` | timeout, hung 9m50s | timeout, hung **9m48s / 9m50s** — same test, both measurements |
| `internal/statusline` `TestBuilder_Build_FullData` | timeout, hung 10m0s | timeout **600.642s / 608.341s** — same test, both measurements |

The orchestrator also ran the full suite independently at `7939997cb` and got exit 1 with the same two packages timing out — the run I recorded as exit 0. So the non-determinism reproduces on both sides, and the disagreement between a 244s pass and a 601s hang on identical code is that non-determinism, not a discrepancy between observers.

The same code passed both earlier in the same session (`internal/statusline ok 30.847s`, `internal/cli ok 244.804s`), so the variable is the environment, not the tree. The likely mechanism is visible in `internal/statusline/usage.go`: it issues HTTP requests to the Anthropic API and shells out to `security find-generic-password` for the macOS keychain, and `TestBuilder_Build_FullData` calls `Build(context.Background(), …)` with no deadline — so an unavailable network or a prompting keychain blocks indefinitely.

Neither hang is attributable to this SPEC: nothing in the diff touches `internal/statusline`, and this SPEC's own `internal/cli` tests still pass in 1.879s under the very conditions where the package as a whole hangs —
`go test ./internal/cli/ -count=1 -run 'TestRoutingLedger_TerminalCloseEndToEnd|TestHarnessObserve_Bash|TestHarnessLedgerAnnotateCmd|TestLedgerVerbs|TestHarnessObserveStop'` → `ok …/internal/cli 1.879s`.

**Consequence for the §D gate — stated plainly rather than smoothed over.** The local full suite is **non-deterministic on this machine at a 10-minute timeout**, which means it **cannot distinguish a real regression from a hang**. The §D full-suite row must therefore be read as green-with-caveat, and it would be wrong to report that the local suite passed: it passed once and failed once on identical code, and both observations are recorded above.

**CI is the arbiter for PR #1420**, not this local run. Re-measuring `internal/cli` coverage was blocked by the same hang (`-timeout=20m` → `FAIL … 1200.889s`; a `-skip TestRunStatusline_NilDeps` attempt also exceeded 10 minutes), so the 76.5% figure above is the value observed at `7939997cb` and was not re-confirmed at HEAD.

### R1 discharged — the live dogfood ran

The R1 gap below was closed by the orchestrator after this run-phase report was first written. The check was performed against a throwaway project under the scratchpad — its own `system.yaml` with `hook.opt_in.enabled: true`, a `.claude/settings.json` pointing at this worktree's `bin/moai` by absolute path, and a minimal Go module so a real `go test` would execute — so neither the primary checkout nor the global binary was touched.

Final ledger row from a live `claude -p` session:

```json
{"session_id":"69f536d5","request_class":"other",
 "delegations":[{"agent":"Explore","outcome":"unknown","blocker":null}],
 "outcome":"success",
 "evidence_refs":[{"kind":"gate_exit","value":"0","ref":"session test evidence","terminal":true}]}
```

One row, `delegations` populated with the correct agent identity, terminal outcome `success`, pending file cleared. That is precisely the shape the original four-row ledger never produced (`spec.md` §A.1: every row `delegations: []`, outcomes only `abort`/`reroute`).

The M3 fix is confirmed independently of the ledger: the throwaway project's telemetry recorded `{"is_test_pass":true,"outcome":"success"}`. Against a production corpus of 37,107 records containing **zero** such signals (`spec.md` §A.6), this is the first one — the `buildBashRecord` path that §A.6 measured as unreachable now executes under real dispatch.

**What this proves:** seams A, B, C and the M3 Bash-evidence path all fire from real Claude Code hook dispatch, not from a test harness. AC-HLE-014's live-invocation caveat is satisfied.

**What it does not prove:** the anomaly recorded below remains open, and the falsification review still needs 50 rows.

#### Observed hook invocation order — and why the create-if-absent split was load-bearing

Captured by wrapping each hook in a tracing script:

```
10:15:51 user-prompt-submit    UserPromptSubmit  sid=e7cf92fc
10:15:55 harness-observe       PostToolUse
10:16:00 harness-observe-stop  Stop                        <- Stop fires MID-SESSION
10:16:03 harness-observe       PostToolUse  agent=Explore
10:16:06 subagent-stop         SubagentStop agent=Explore
10:16:06 user-prompt-submit    UserPromptSubmit            <- A fires a SECOND time
10:16:10 harness-observe       PostToolUse
10:16:13 harness-observe-stop  Stop
```

A single `claude -p` invocation fires **UserPromptSubmit and Stop twice each**. This is the runtime behavior that decides whether the SPEC works, and it was not known when the design was chosen.

`RecordIfAbsent` absorbed it exactly as intended: the second UserPromptSubmit was a no-op against the existing row, so the `Explore` delegation survived and the second Stop closed one complete row. Had seam A used `Store.Record`, the second UserPromptSubmit would have rerouted the first row closed, dropped the delegation, and opened a fresh one — regenerating the reroute-only ledger this SPEC exists to repair, while row count rose and the failure looked like success. The B1 trap in `plan.md` §B is real at runtime, not merely theoretical, and the refusal to collapse `RecordIfAbsent` back into `Record` is what held under it.

#### Hook wrapper names — a distinction worth stating explicitly

Seam C lives in `runHarnessObserveStop` (`internal/cli/hook.go:733`), which is the **`moai hook harness-observe-stop`** subcommand — **not** `moai hook stop`. Wiring `stop` instead produces a silent no-op: the handler never runs, nothing is written, and no error surfaces. The orchestrator hit exactly this during the dogfood before correcting the wiring.

The four dispatch points, verified against the code:

| Seam | Handler | Dispatched by |
|---|---|---|
| A — pending row | `userPromptSubmitHandler.Handle` (`internal/hook/user_prompt_submit.go`) | `moai hook user-prompt-submit` |
| B — delegation | `subagentStopHandler.Handle` (`internal/hook/subagent_stop.go`) | `moai hook subagent-stop` |
| C — terminal evidence | `runHarnessObserveStop` (`internal/cli/hook.go`) | `moai hook harness-observe-stop` |
| M3 — Bash evidence | `runHarnessObserve` (`internal/cli/hook.go`) | `moai hook harness-observe` |

#### Open anomaly — delegation/outcome ordering (new residual risk R8)

An earlier dogfood run (same setup, untraced) produced **two** ledger rows for one session, both with `delegations: []` and `outcome: "success"`. It did **not** reproduce across two subsequent traced runs, each of which produced a single correct row.

Stated as a hypothesis, because it was not reproduced and therefore not confirmed: in that run `go test` executed *before* the `Explore` call, so a terminal signal already existed at the first (mid-session) Stop. The row closed early carrying no delegation, and the later delegation landed on a freshly created row.

If that reading is right, it is an ordering hazard rather than a coding error — every seam did exactly what it is specified to do — but it matters downstream, and it is filed as **R8** in `acceptance.md` §F and cross-referenced into `SPEC-HARNESS-LEARNING-EVO-002` §B. Not fixed in 001, by decision.

### Gaps — what was NOT verified

- ~~**The live hook-dispatch link (R1).**~~ **DISCHARGED** — see § R1 discharged above. Still true: **no CI-runnable assertion covers this link**, so it remains a manual check rather than a gate.

  Correcting the reason this section originally gave: it stated the dogfood was not performed *because* enabling a fail-closed gate lies outside this worktree's write boundary. That rationale was wrong. The boundary held and was never crossed — the orchestrator ran the check against a throwaway project with its own `system.yaml`, its own `.claude/settings.json` pointing at the worktree's `bin/moai` by absolute path, and a minimal Go module, touching neither the primary checkout nor the global binary. The check was reachable inside the boundary all along; what was missing was the throwaway-project route, not permission. The distinction matters because "blocked by a boundary" would justify leaving R1 permanently open, and it does not.
- **A live `is_test_fail` signal end to end.** The dogfood observed `is_test_pass` under real dispatch; the **failing** direction was not exercised live. It is covered by unit tests driving the production path with a fixture (`TestHarnessObserve_BashEvidenceTestFail`), but no observed live run has produced `is_test_fail` through the restored carve-out.
- **The two-row anomaly was not reproduced**, so its stated cause is a hypothesis rather than a diagnosis. Recorded as R8.
- **The falsification review (`spec.md` §E F1/F2).** Needs 50 finalized rows, which do not exist at merge time. Scheduled, not performed — as `acceptance.md` §E already states.

### Residual risk

R1-R6 stand as written in `acceptance.md` §F. One is added by the M0 measurement:

- **R7 — the terminal signal rests on an output-text heuristic, not a structured exit code.** The observe channel's `tool_response` carries no `exit_code`, so `deriveFromExitCode` cannot fire and pass/fail detection falls to `deriveFromOutputText`'s marker matching (`ok  \t`, `--- FAIL`, `test result:`, and a `passed`/`failed` word count). A test runner whose output matches none of those yields no signal, the row stays pending, and it closes later as `abort` via the stale sweep rather than as `fail`. The live yield of terminal `success`/`fail` rows may therefore be lower than the code path suggests, and that will only be measurable once rows accumulate.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-10
run_commit_sha: 7939997cb
run_status: audit-ready
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 15
m1_to_mN_commit_strategy: "four milestones in three commits — M1+M2 share 710a735b6 because the identity markers live in the same file as the store types; M3 is 947cee983; M4 is 7939997cb. M0 produced no commit: it is a probe, and its result is recorded in §E.2 rather than in code."
```

Notes for the auditor:

- **The run-phase code is merged; this documentation is not.** PR #1420 squash-merged to `main` as `8ed9622f3` with all CI checks green. Because it was a squash, none of the branch commits are ancestors of `main`, but the code content did land — verified by `git cat-file -e origin/main:internal/hook/routing_ledger.go` and by `func LogBashEvidence` being present in `origin/main:internal/hook/evidence_writer.go`. What did NOT land is the R1 discharge and the R8 filing: `git diff --name-only origin/main..HEAD` returns only the three SPEC documents, and no code file. So the follow-up carries documentation exclusively — a docs-only delta over the merged implementation.

- **CI settles the local-suite question.** Every check on PR #1420 passed, including builds on five platforms and integration tests on three operating systems. That is independent confirmation that the two statusline timeouts recorded above are a local-environment condition, not a property of the tree — which is what the green-with-caveat §D row already assumed but could not prove from this machine.

- **SHA citations were rewritten once.** The branch was rebased onto `85693f3bf` mid-run (the orchestrator updated it for PR #1420), which rewrote every run-phase commit. The SHAs above and throughout §E.2 are the **post-rebase** values and are all ancestors of the current HEAD; the pre-rebase values cited in earlier reports (`038bcc4fc` / `631c9aa1c` / `637f9a65f` / `f83583d39`) are no longer on the branch. The evidence itself is unaffected — the trees are identical — but a reader chasing an old SHA would find nothing on the branch, so the citations were corrected rather than left dangling.

- **M0 changed nothing in the plan.** The probe confirmed option (c), so the declared option (a) fallback — a `settings.json.tmpl` edit plus `make build`, plus the §25 neutrality re-check — was never entered. No template file, frozen skill, or frozen allowlist was touched.
- **One deviation from `plan.md` §I's file inventory.** The inventory anticipated edits to `internal/cli/hook_test.go` and `internal/telemetry/recorder_test.go`; the CLI-side tests landed in a new file `internal/cli/hook_routing_ledger_test.go` instead, because `hook_test.go` is a small unrelated file and appending an unrelated 300-line fixture set to it would have obscured both. The telemetry edit landed as planned.
- **One addition beyond the inventory**, made to avoid a defect rather than for tidiness: the two observation gates existed only in `internal/cli`, which `internal/hook` cannot import. Copying them would have created two truth tables that could drift — and a drift between them is exactly the class of failure this SPEC is repairing elsewhere. The implementations moved down into `internal/hook` (`HarnessLearningEnabled`, `HookObserveOptInEnabled`) and the CLI functions became thin delegating wrappers, so every existing call site and test is unchanged.
- **`Store.LoadPending` was exported**, which `plan.md` §I did not list. The first-writer-wins policy (REQ-HLE-006, §E D1) requires the seam to ask "is this field still empty?", and §E D1 explicitly places that policy at the seam rather than in the store — so the seam needs a read primitive. Baking the policy into `Annotate` was the alternative and was rejected for that reason.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
