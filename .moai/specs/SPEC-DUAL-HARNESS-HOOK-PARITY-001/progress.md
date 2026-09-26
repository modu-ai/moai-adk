# Progress — SPEC-DUAL-HARNESS-HOOK-PARITY-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L), progress.md
- SPEC ID check: `[[ "SPEC-DUAL-HARNESS-HOOK-PARITY-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`
- Baseline: HEAD `530d8cc06`, branch `WT-dual-harness-parity-rebuild`
- REQ count 25 (REQ-HPR-001..025), AC count 22 (AC-HPR-001..022)
- v0.2.0 (plan-audit iter-1 revision): Q1–Q6 resolved, decision record in plan.md §C; no open clarification markers
- v0.3.0 (plan-audit iter-2 revision): N1–N6, N8, N9 addressed; Q1/Q2/Q6 to be confirmed by the operator at Implementation Kickoff (N7)
- v0.4.0 (plan-audit iter-3 revision, delta re-audit authorized for R1–R4): member 6 on the receipt method, fail-closed when codex is installed (R1); sync-gate self-gate before the receipt (R2); intentional test-amendment list completed (R3); plan-phase `live-uncertified` tag + closure-mode HISTORY line written, sync close limited to §E.4 + CHANGELOG (R4); A2, A4
- Completion condition: closes as partial (live-uncertified) per operator decision Q5 (spec.md §E)

## §E.2 Run-phase Evidence

### Run-phase entry (2026-09-23)

- Implementation Kickoff Approval: granted by the operator on 2026-09-23. Progression mode: step-wise
  (the run stops after each milestone for operator review).
- Operator confirmation of the Jev-sourced decisions: **Q1** (whole-catalog obligation registry, M1
  rows `blocked:M1`), **Q2** (Codex `needs_input` → fail-closed deny, surfaced visibly), and **Q6**
  (SPEC-CODEX-HOOK-ADAPTER-001 REQ-7 kept — nothing under `internal/hook` changes) were confirmed by
  the operator at kickoff (plan.md §C, plan-audit iter-2 N7). They are now operator decisions.
- Status transition `draft → in-progress` on spec.md (the only artifact carrying frontmatter).
- Run baseline: HEAD `9e92fbb88` on `WT-dual-harness-parity-rebuild` (develop `533929f2b`
  absorbed; the SPEC last changed at `b10042d04`).

### M2a — decision and state models (2026-09-23)

Commits: `ce15e08b8` (goal `cancelled`), `ac02f2d61` (decision table), `db1273e5e` (obligation
registry). Measured on Darwin arm64, go1.26.8, tree = those commits. `internal/hook` is unchanged:
`git diff --stat -- internal/hook` printed nothing.

Re-measured anchors (all matched design.md §D6 except where noted): `evaluate.go:294` early return,
writers `:306/325/340/404/435`, `launcher_blockcap_infinite.go:68`, `handoff.go:85`,
`goal.go:1282/1313/1331`, `dashboard.go:141`, `session_start_compact.go:88`,
`handoff_inject.go:185`, `stop_failure.go:111`, `output.go:40` (`"ask": true`). The type-checked
scan found no site outside the design table; it also records three declaration-level uses of the
`Status` type in `schema.go` (the const block, `type Goal`, the new `IsKnownStatus`) that a grep for
`g.Status` cannot see, and lists them.

**AC-HPR-013 — consumer leg** (`TestGoalStatusConsumersHandleCancelled`, `./internal/goal/`)

- RED (compile): `internal/goal/status_consumers_test.go:268:6: undefined: StatusCancelled` …
  `v.Diagnostic undefined` → `FAIL github.com/modu-ai/moai-adk/internal/goal [build failed]`
- RED (runtime, constant added, evaluator unchanged):
  `site internal/goal/evaluate.go|(*Eval).Evaluate references Goal.Status,StatusCeilingExit,StatusCleared,StatusSatisfied,StatusUnsatisfiable, table lists …StatusCancelled…`;
  `unmet condition: cancelled goal blocked`; `turn ceiling: status overwritten to "ceiling-exit"`;
  `status "paused" / unmet condition: silent block`; `status "" / satisfied condition: mapped to satisfied`
- GREEN: `go test -json -count=1 -run '^TestGoalStatusConsumersHandleCancelled$' ./internal/goal/ | jq …`
  → `7 pass`, `7 run`, no `skip`, no `fail`
- Mutation 1 (source, reverted): drop `StatusCancelled` from the early-return case →
  `unmet condition: cancelled goal produced output {… Diagnostic:stop-goal: unrecognised goal status "cancelled" …}` → FAIL
- Mutation 2 (source, reverted): make an unknown status fall through to evaluation →
  `status "paused" / unmet condition: silent block` → FAIL
- Mutation 3 (in-test): drop one row from the table → the diff names the dropped site

**AC-HPR-013 — internal/cli readers** (`TestGoalCancelledStatusReaders`, `./internal/cli/`)

- RED: `goal_cancelled_readers_test.go:64: stderr does not surface the unrecognised status: ""`
  (the other five legs are characterization assertions of the "none" rows of §D6 and passed on the
  first run)
- GREEN: `go test -json -count=1 -timeout 25m -run '^TestGoalCancelledStatusReaders$' ./internal/cli/ | jq …`
  → `7 pass`, `7 run`, no `skip`, no `fail`
- Not in M2a: the precedence leg `TestGoalCancellationPrecedence` (Interrupt producer, M2e/M2f).

**AC-HPR-006 — table leg** (`TestDecisionTranslationNeverLoosens`, `./internal/codexadapter/`)

- RED (compile): `decision_test.go:14:25: undefined: Translation` … `too many errors`
- GREEN: `go test -json -count=1 -run '^TestDecisionTranslationNeverLoosens' ./internal/codexadapter/ | jq …`
  → `7 pass`, `7 run`, no `skip`, no `fail`
- Mutation (source, reverted): Codex PreToolUse `needs_input` → `no-opinion` (the restored drop) →
  `codex/PreToolUse/needs_input → no-opinion is in the host-resolves-as-allow set`,
  `… is the empty no-opinion object`, `codex/PreToolUse/needs_input rendered the empty object` → FAIL.
  The same mutation also runs in-test (`the checker names a restored ask drop`).
- Not in M2a: the live `MapOutput` path still carries the card-t590 `ask` drop (`output.go:40`);
  switching it onto the table is M2c, and AC-HPR-006 is not closed until then.

**AC-HPR-018 — schema leg** (`TestObligationCoverage`, `TestObligationCoverageSchema`, `./internal/template/`)

- RED (compile): `obligations_test.go:66:54: undefined: ObligationRegistry` … `undefined: CheckObligationCoverage`
- GREEN: `go test -json -count=1 -run '^TestObligationCoverage' ./internal/template/ | jq …`
  → `18 pass`, `18 run`, no `skip`, no `fail`
- Mutations (in-test): remove the Codex path, the Claude path, the check; name a missing check; name
  an unresolvable Codex path → each yields exactly one violation naming the obligation id.
  Source mutation (reverted): skip the path-resolution test →
  `want one violation for claude-interrupt-cancellation mentioning "moai hook no-such", got []` → FAIL
- Not in M2a: the embedded whole-catalog registry file and the production path resolver (M2h);
  `<registry-pkg>` resolves to `internal/template` (design.md §D4 default).

**Package runs and static checks**

- `go test -count=1 ./internal/goal/ ./internal/codexadapter/` → `ok … goal 8.642s`, `ok … codexadapter 0.639s`
- `go test -count=1 -timeout 25m ./internal/template/` → `ok … template 176.909s`
- `go test -count=1 -timeout 25m -run 'Goal|StopGoal|Handoff|BlockCap|Launcher' ./internal/cli/` → `ok … cli 12.719s`
- `go vet ./internal/goal/ ./internal/codexadapter/ ./internal/template/ ./internal/cli/` → no output (exit 0)
- `golangci-lint run ./internal/goal/... ./internal/codexadapter/...` → `0 issues.`;
  `golangci-lint run ./internal/template/ ./internal/cli/` → `0 issues.`

Residual risk: a goal file with an empty `status` now reads as unrecognised (diagnostic, no block)
instead of being evaluated. Every writer in the tree sets a status, so no such file is expected, but
a hand-edited or pre-schema file would stop blocking.

### M2b — receipt contract and Stop budget (2026-09-23)

Commits: `1be628d9f` (verify receipt contract) and `dba21895e` (Codex Stop-chain budget
declarations). Measured on Darwin arm64, go1.26.8, against the tree at those commits.
`git diff --stat HEAD -- internal/hook` printed nothing.

Re-measured anchors: `defaultHandlerTimeout = 10` (`codexwiring.go:60`), `moaiHandlerTimeout`
(`hooks.go:33–39`), the Stop row `{hook.EventStop, "stop", true}` (`events.go:74`), `verify.Key`
(`key.go:39`), `Fresh` (`freshness.go:17`), and `CheckEntry` (`schema.go:34`). Ledger L-06 is
closed: `ToolVersion` now appears in `internal/verify`.

**AC-HPR-017 — receipt fields** (`TestCheckReceipt*`, `./internal/verify/`)

- RED (compile): `receipt_test.go:12:24: undefined: Receipt` … `undefined: CheckReceipt` → build failed
- RED (runtime, stub = today's snapshot semantics, key + command only):
  `TestCheckReceiptFieldMismatchIsNotRun/config_digest: config_digest differs but the receipt was accepted`;
  `…/tool_version: tool_version differs but the receipt was accepted`;
  `TestCheckReceiptLegacyEntryIsNotRun: a legacy entry without the receipt fields must not be accepted`;
  `…/head: reason must name head, got "stale key"`
- GREEN: `go test -json -count=1 -run '^TestCheckReceipt' ./internal/verify/` (output kept at
  `.moai/state/verify/m2b/receipt.json`), then `jq -r 'select(.Test!=null) | .Action' | sort | uniq -c`
  → `21 pass`, `21 run`. No `skip` and no `fail`.
- Mutations (source, each reverted): dropping one field from the comparison, for each of the five
  fields, turned red the matching subtests. For example, dropping `head` failed
  `TestCheckReceiptStoredFieldMismatchIsNotRun/head` and `TestCheckReceiptFieldMismatchIsNotRun/head`.
  Dropping `tool_version` also failed `TestCheckReceiptEmptyFieldIsNotRun/tool_version`.
  An absent receipt treated as run →
  `absent receipt must read as not run, got {Run:true …}`, plus the round-trip and truncation
  tests fail. A truncated snapshot synthesized as a receipt →
  `a truncated receipt must load as absent, got &{…}`.
- Rule: every one of the five fields must be non-empty on both sides and equal. An absent receipt,
  a truncated file, an unbound field, or a record older than the TTL (`DefaultTTL`, kept as an
  extra bound) reads as not run.

**Backward compatibility** (`TestSnapshotRecordShapeBackwardCompatible`, characterization)

- It passed on the pre-change shape, before any schema edit, and passes after. A legacy snapshot
  still loads and still serves `Source.Lookup`. Re-saving it writes no `config_digest`,
  `tool_version`, or `verdict` key.
- Mutation (reverted): dropping `omitempty` from `tool_version` →
  `re-saved legacy snapshot gained key "tool_version"` → FAIL.
- `go test -count=1 -timeout 25m -run 'Verify|StopGoal|MCP' ./internal/cli/` → `ok … cli 22.743s`
  (the `moai verify` readers and the stop-goal lookup).

**AC-HPR-016 — declaration and sum leg** (`TestStopChainAggregateBudgetFitsTimeout`,
`TestStopChainReceiptPlacement`, `TestStopUnmeasuredCapDeclared`, `./internal/codexwiring/`)

- RED (compile): `stop_budget_test.go:48:9: undefined: StopChainMembers` …
- RED (runtime, the Claude registration timeouts carried over):
  `member 3 (moai hook stop-goal) budget 1m30s exceeds T_stop 10s (REQ-HPR-018)`;
  `Σ member budgets 32m50s + chain_overhead 1s = 32m51s exceeds T_stop 10s`;
  `member 6 … receipt placement = false, want true`
- GREEN: `go test -json -count=1 -run '^TestStop' ./internal/codexwiring/` (output kept at
  `.moai/state/verify/m2b/budget.json`) → all three tests `pass`. No `skip` and no `fail`.
- Mutations (reverted): the AC's own, raising member 3's budget by 1 s →
  `Σ member budgets 8.2s + chain_overhead 2.8s = 11s exceeds T_stop 10s` → FAIL. Rendering the
  Stop handler at 20 s while `T_codex_max` is unmeasured →
  `T_codex_max is unmeasured, so T_stop must stay at the render constant 10s, got 20s` → FAIL.
  Setting `StopUnmeasuredCap = 1` → FAIL. Placing member 6 in-hook → FAIL.
- Declared values. **All are proposals, and none is a measurement.**
  - T_stop is 10 s, read from the rendered `hooks.json`.
  - Member budgets are (2 + 0.2) + 0.5 + 2 + 0.5 + 0.5 + 0.5 + 0.5 + 0.5 = 7.2 s, where 0.2 s is
    member 1's uncut factory step.
  - `StopChainOverhead` is 2.8 s: the full headroom, so the runner deadline equals the member sum
    and any budget raise has to be traded against another member.
  - `StopTimeoutCodexMax` is 0, meaning NOT_RUN (Q5).
  - `StopUnmeasuredCap` is 3 and `StopCapStateDir` is `.moai/state/codex-stop-cap`. M2d finalizes
    both.
- Not in M2b: the timing leg (`TestStopChainMemberCostWithinBudget`, M2d). The budgets are not
  validated until it passes.

**Package runs and static checks**

- `go test -count=1 -cover ./internal/verify/` → `ok … 84.6%`. The HEAD baseline, measured the
  same way with the M2b files set aside, was `81.0%`. `receipt.go` is at 100% except
  `LoadReceipt` (87.5%). The remaining shortfall against 85% is in pre-existing files.
- `go test -count=1 -cover ./internal/codexwiring/` → `ok … 89.6%`
- `go vet ./internal/codexwiring/ ./internal/verify/` → exit 0
- `golangci-lint run ./internal/codexwiring/... ./internal/verify/...` → `0 issues.`

**Deviation and open decision for M2d: where the sync-gate receipt is stored.** Design §D3.3
says the sync gate's receipt is the existing `.moai/state/sync-quality-gate.last` line, extended
with the §D3.6 fields. M2b does not change that line. The Claude script rejects any trailing
field: at `sync-phase-quality-gate.sh:440–445` a non-empty `R_EXTRA` leaves `RECORD_OUTCOME`
empty, so the checks re-run. Extending the line in place would therefore change Claude
behaviour on every turn, which §D3.3 rules out. The receipt contract stores any gate outcome as a
verify snapshot entry instead, with `verdict` = `pass`/`fail`/`inconclusive`. That is the other
store §D3.6 names. When M2d decides the port shape (Q4), it also chooses between two options:
- (a) the out-of-hook sync-gate entry writes a verify receipt, and `.last` stays Claude-only;
- (b) change the Claude parser as well, which would need a golden proving Claude behaviour is
  unchanged.

### M2c — decision hardening (2026-09-24 close-out)

Commits: `6a3e745d5` (codexadapter: `needs_input` rendered as a fail-closed deny through the
translation table — new `translate.go` with `TranslateCodex` / `RequiredInputUserApproval`;
`MapOutput` routes a PreToolUse `ask`/`defer` through it) and `c2d06a518` (cli: a Codex hook
fault on a decision-bearing event is answered with the table's `fatal_error` deny on exit 0 —
`hook_codex_failclosed.go`). Measured on Darwin arm64, go1.26.8, against the tree at HEAD
`c2d06a518`. `git diff --stat HEAD~2 HEAD -- internal/hook` printed nothing (Q6 / HOOK-ADAPTER REQ-7
kept).

**Evidence provenance — read this first.** The agent that wrote both commits was killed by an API
rate limit after committing and before recording any evidence. Its RED-first observations — the
failing test output captured before each implementation — were lost with it and cannot be
reconstructed. What follows is **post-hoc mutation evidence** gathered by a close-out agent on
the committed code: each mutation was applied to the implementation, the covering tests were
observed red, and the mutation was reverted by re-editing (not `git checkout --`), with
`git diff --stat` printing nothing after each revert. Mutation evidence shows the tests can detect
the defect; it does not show the tests were written before the code.

**AC leg → covering test**

| AC leg | Covering test (package) |
|---|---|
| AC-HPR-006 table leg (M2a) | `TestDecisionTranslationNeverLoosens` table subtests (`./internal/codexadapter/`) |
| AC-HPR-006 real path — every deny / needs_input through `TranslateCodex` on every decision-bearing event, and Claude-shape `deny`/`ask`/`defer` through `MapOutput` | `TestDecisionTranslationNeverLoosens/the_real_Codex_output_path_never_loosens` |
| AC-HPR-006 intentional amendment (plan.md M2c row) | `TestPreToolUseAskBecomesFailClosedDeny` (inverted from `TestPreToolUseAskDropped`; covers `ask` and `defer`) |
| AC-HPR-022 PreToolUse + PermissionRequest: deny names the required input, keeps the handler reason, exactly one sink record, stderr mirror | `TestNeedsInputVisibleDeny/PreToolUse`, `/PermissionRequest` |
| AC-HPR-022 real PreToolUse path through `MapOutput` | `TestNeedsInputVisibleDeny/PreToolUse_ask_via_MapOutput` |
| AC-HPR-008 unit — timeout, exit 1 (handler error), unparseable output, exit 2 on PreToolUse and Stop through the real subcommand under `--harness codex` | `TestHookFaultInjection/{PreToolUse,Stop}/{timeout,exit_1_(handler_error),unparseable_output,exit_2}` (`./internal/cli/`) |
| AC-HPR-008 unit — PermissionRequest (row unadapted until M2e, so the fail-closed writer is called directly) | `TestHookFaultInjection/PermissionRequest/{timeout,exit_1_(handler_error),unparseable_output}` |
| `TranslateCodex` branches (fatal_error, deny/advisory shapes, decision-bearing list) | `TestTranslateCodexFatalErrorIsFailClosed`, `TestTranslateCodexDenyAndAdvisoryShapes`, `TestIsDecisionBearingMatchesTheDeclaredList` |

Gaps in M2c scope found: none; no test or implementation was added by the close-out.

**Green runs (post-revert, HEAD `c2d06a518`)**

- `go test -json -count=1 -run 'TestDecisionTranslationNeverLoosens|TestNeedsInputVisibleDeny|TestPreToolUseAskBecomesFailClosedDeny|TestTranslateCodex|TestIsDecisionBearing' ./internal/codexadapter/`
  → `jq … | sort | uniq -c` → `16 pass`. No `skip`, no `fail`.
- `go test -json -count=1 -timeout 25m -run 'TestHookFaultInjection' ./internal/cli/` → `12 pass`
  (11 subtests + parent). No `skip`, no `fail`.
- The same two runs before the mutations gave the same counts (16 pass, 12 pass).
- `go test -count=1 -cover ./internal/codexadapter/` → `ok … codexadapter 0.623s coverage: 88.4% of statements`
- `go vet ./internal/codexadapter/ ./internal/cli/` → no output (exit 0)
- `golangci-lint run ./internal/codexadapter/... ./internal/cli/` → `0 issues.`
- Output files are kept under `.moai/state/verify/m2c/` (git-ignored, local only).

**Mutation evidence (each applied, observed red, reverted; `git diff --stat` empty after each)**

1. Restore the card-t590 drop (`"ask": true, "defer": true` back in `preToolUseDropDecisions`,
   `output.go`) →
   `decision_test.go:146: PreToolUse ask: MapOutput gave {}, want a deny`;
   `output_test.go:296: ask: output = {}, want a hookSpecificOutput deny, not the no-opinion object`;
   `needs_input_test.go:132: PreToolUse: rendered the empty no-opinion object; a needs_input must never degrade to {}` → FAIL
2. Suppress the discard record (`if row.DiscardRecorded && false`, `translate.go`) →
   `output_test.go:306: ask: discards = 0, want 1`;
   `needs_input_test.go:102: open diagnostic sink: open …/.moai/logs/codex-adapter.jsonl: no such file or directory`
   (PreToolUse and PermissionRequest) → FAIL
3. Drop the required-input name from the reason (`"input required: …"`, `translate.go`) →
   `needs_input_test.go:91: deny reason "input required: Codex hooks cannot ask for it, so MoAI denied this PermissionRequest fail-closed — Critical config file: settings.json" does not name the required input "user approval"`;
   `output_test.go:303: ask: reason = "input required: …", want it to name "user approval" and keep the handler reason` → FAIL
4. A hook fault yields allow:
   - a. the fail-closed writer emits `{}` (`hook_codex_failclosed.go`; the AC-HPR-008 mutation in
     acceptance.md) → all nine fault legs fail, e.g.
     `hook_fault_injection_test.go:176: PreToolUse: stdout = "{}", want a fail-closed deny (Codex resolves an empty or {} output as allow)`
     and `hook_fault_injection_test.go:224: PermissionRequest: stdout = "{}", …` → FAIL
   - b. the dispatcher swallows a timeout / handler error (`return nil`, `hook.go`) →
     `hook_fault_injection_test.go:176: Stop: stdout = "", want a fail-closed deny …`
     (timeout and exit 1 on PreToolUse and Stop) → FAIL
   - c. unparseable output passed through as `{}` (`hook_harness_codex.go`) →
     `hook_fault_injection_test.go:176: PreToolUse: stdout = "{}", …` and the Stop leg → FAIL
   - d. exit 2 no longer propagates under codex (`&& !harnessCodex`, `hook.go`) →
     `hook_fault_injection_test.go:197: RunE = <nil>, want exit code 2 to pass through`
     (PreToolUse and Stop) → FAIL
5. `MapOutput` bypasses the table (the `ask`/`defer` arm hand-builds a deny carrying only the
   handler reason, no `TranslateCodex`, no discard) →
   `output_test.go:303: ask: reason = "confirm?", want it to name "user approval" and keep the handler reason`;
   `output_test.go:306: ask: discards = 0, want 1`;
   `needs_input_test.go:134: deny reason "Critical config file: settings.json" does not name the required input "user approval"` → FAIL.
   `TestDecisionTranslationNeverLoosens` stayed green under this mutation: its real-path leg checks
   that the output is a deny, not which route produced it. The bypass is caught by the AC-HPR-022
   tests.
6. `needs_input` falls back to `{}` inside `TranslateCodex` (AC-HPR-022's third named mutation) →
   `decision_test.go:122: codex/PermissionRequest/needs_input: real path rendered the empty object`
   (all four decision-bearing events);
   `needs_input_test.go:89: PermissionRequest: rendered the empty no-opinion object; a needs_input must never degrade to {}` → FAIL

**Not in M2c / not observed**

- Live legs (`TestLiveHookFaultOutcome`, AC-HPR-007, the host outcome per fault × event × harness)
  are not executed (operator Q5); they belong to M2g as `NOT_RUN`.
- The timeout leg is simulated: the fault registry returns `hook.ErrHookTimeout`. No test sleeps
  past `config.DefaultHookDispatcherTimeout`, and what Codex does when it kills the process at its
  `hooks.json` timeout is the live leg.
- "Exit 1" is modelled in-process as a handler error. No handler sets `HookOutput.ExitCode = 1`;
  that shape is not covered.
- PermissionRequest reaches the fail-closed writer only by a direct call until the row is adapted
  in M2e.

Residual risk (observed while mapping, outside the AC-HPR-008 fault list): when stdin is not
valid JSON, `runHookEvent` (`hook.go`, the `ReadInput` error branch) writes the default `{}` on
exit 0 on every path, including `--harness codex` on a decision-bearing event. Codex may resolve
that `{}` as allow. The AC-HPR-008 legs cover unparseable hook *output*, not unparseable *input*,
so this was left unchanged. It needs an operator decision on whether it is in scope.

### M2d — Stop chain on Codex (2026-09-24, recorded per sub-step)

Base HEAD `fabc33812`, branch `WT-dual-harness-parity-rebuild`. Measured on Darwin arm64. Every
RED below was captured before the implementation it names, as the sub-step ran.

**Operator decisions for M2d (source: operator, 09-24)**

- Sync-gate receipt store = option (a): the Codex sync check writes a verify-snapshot receipt (the
  M2b `CheckEntry` with `verdict` / `config_digest` / `tool_version`). `.moai/state/sync-quality-gate.last`
  stays Claude-only. `.claude/hooks/moai/sync-phase-quality-gate.sh` is not modified, so Claude
  behaviour stays byte-identical.
- Q4 (sync-gate port shape) is decided in this milestone per plan.md/design.md (see Q4 below).
- The malformed-stdin fail-open in `internal/cli/hook.go` (`runHookEvent`, `ReadInput` error branch)
  is out of scope and split to a separate card. It is not changed here.

**AC-HPR-001 — Stop-chain inventory over both template renders**
(`TestStopChainInventoryMatchesClaudeTemplate`, `./internal/codexwiring/`)

- Inventory: `StopChainMembers` gained `ClaudeScript`, `Conditional`, and `ReceiptProducer`. The
  checker is `CheckStopInventory(plain, optIn []string)` in `stop_inventory.go`. The test renders
  `settings.json.tmpl` with `HookOptIn.Enabled` false and true, reads each Stop array in
  registration order, and requires: each handler maps to exactly one row; each row has a class and a
  placement (a receipt row names its producer); the opt-in-only row is `Conditional`; no row is
  absent from both renders; the numbering follows the template order.
- RED (compile): `stop_inventory_test.go:70:17: undefined: CheckStopInventory` …
  `m.ClaudeScript undefined (type StopMember has no field or method ClaudeScript)`.
- RED (runtime, a stub checker that returns nothing):
  `an unlisted Stop handler was not reported; problems = []` and
  `the opt-in-only member was not reported missing; problems = []`. The top-level check passed
  vacuously on the stub, which is why the test carries the two in-memory mutations as subtests.
- GREEN: `go test -json -count=1 -run '^TestStopChainInventoryMatchesClaudeTemplate$' ./internal/codexwiring/`
  (kept at `.moai/state/verify/m2d/ac001.json`) → `3 pass` (the test and both subtests). No `skip`,
  no `fail`.
- Mutations (each reverted with Edit; `git diff --stat` on the file empty after each):
  1. A `handle-mutant-stop.sh` handler added to the template's Stop array →
     `Stop handler handle-mutant-stop.sh (render opt-in off) has no inventory row` (and the opt-in
     render), plus the order shift for members 2–8 → FAIL.
  2. Member 8 `Conditional: false` →
     `inventory row 8 (moai hook harness-observe-stop) conditional = false, but the renders say opt-in-only = true` → FAIL.
  3. (in-test) inventory without `harness-observe-stop` against the opt-in render → reported.

**Resumption note (evidence provenance).** The agent that wrote `fe4fd9d4d` stopped on the model
quota with four files uncommitted (backup kept at the primary checkout's
`.moai/reports/t1099/m2d-partial/`, byte-identical to the worktree when this agent started:
`cmp` of both new files and `git diff | cmp - tracked.patch` → `IDENTICAL`). Any RED that agent
observed for those files was lost with it. The resuming agent recorded the partial state's own
compile RED before writing any production code for it:
`go test -count=1 -run '^TestStopChainEffectParityGolden' ./internal/cli/` →
`internal/cli/codex_stop_chain_golden_test.go:188:65: undefined: stopMemberOutcome` …
`undefined: newCodexStopChain` … `undefined: renderCodexStop` … `FAIL … [build failed]`.
Where a test below was written after its code, that is said, and mutation evidence stands in.
Commits `aedeb4fd4` … `629d13eb9`; measured on Darwin arm64, go1.26.8, at those commits. The
machine was shared with other lanes throughout (load averages 37–49 at the recorded points).
`git diff --stat fe4fd9d4d 629d13eb9 -- internal/hook .claude/hooks internal/template/templates/.claude/hooks`
printed nothing: `internal/hook` (Q6) and the Claude hook scripts are unchanged.

**Decisions made in M2d**

- **Q4 — sync-gate port shape: parallel Go implementation, not a shim.** The operator's
  constraint keeps `sync-phase-quality-gate.sh` byte-identical, so the Claude script cannot become
  a shim over a Go entry. The decision core is ported to Go (`internal/cli/codex_sync_gate.go`):
  the self-gates (subject predicate, language marker, code delta) run in-hook on Codex; the checks
  run out of hook. Equivalence proof: the AC-HPR-002 goldens run the **unmodified** script and the
  Go member on the same fixtures, and two parity tests read or source the script itself
  (`TestSyncGateSubjectPredicateMatchesScript` extracts the script's case arm;
  `TestSyncGateLanguageDetectionMatchesScript` sources `detect_languages`).
- **Receipt producers (not `moai hook` subcommands).** Sync gate: `moai verify sync-gate`. Codex
  review gate (**R1 runner**): `moai verify codex-review`. Both are `moai verify` verbs, so the
  `moai hook` subcommand count (`TestHookCmd_SubcommandCount`, `TestHookCmd_PrePushSubcommandCount`)
  and the `utilitySubcmds` reverse mapping are unchanged — no amendment needed; both tests passed
  in the regression run below.
- **Cap (§D3.8): N = 3, final**, reason in `codexwiring.StopUnmeasuredCap`'s comment: N ≥ 2 keeps
  the first continuation meaningful; the third Stop gives the agent a retry after a producer run
  that left a stale receipt; further continuations repeat the same instruction; it matches
  `goal.DefaultStagnationThreshold`. **Counter directory, final:** `.moai/state/codex-stop-cap/<session-id>.json`
  (ignored by the distributed `.gitignore`, so writing it never moves the working-tree digest it is
  keyed by). Only the Stop chain writes it; the producers never do.
- **Budgets rebalanced after the timing leg** (below): members 2 and 6 from 0.5 s to 1 s, member 3
  from 2 s to 1 s. Σ member budgets stays 7.2 s and `StopChainOverhead` stays 2.8 s.
- **Per-run verdict record:** `.moai/state/codex-stop-chain/<session-id>.json`, one status per
  member (`pass` only for an evaluated gate or goal; `unverified` for a capped gate; `failed` for a
  failed or cut-off advisory member).

**Sub-step 1 — shared Go entries** (`aedeb4fd4`). `evaluateStopGoal` (runner is a parameter) and
`harnessObserveStop` extracted; Claude behaviour unchanged. Characterization run:
`go test -json -count=1 -timeout 20m -run 'StopGoal|HarnessObserve|GoalCancelled|Classify|Propose|Ledger|WireFormat|GateUniformity' ./internal/cli/`
→ `194 pass`, no `fail`, no `skip`.

**Sub-step 2 — sync-gate core + `moai verify sync-gate`** (`b82562b92`)

- RED: `go test -count=1 -run 'TestSyncGate|TestVerifySyncGate' ./internal/cli/` →
  `codex_sync_gate_test.go:156: moai verify sync-gate: unknown command "sync-gate" for "verify"` → FAIL.
  The two parity tests passed on their first run: they are characterizations of the partial's Go
  core against the script, and their detection is shown by mutation.
- GREEN: `go test -json -count=1 -run 'TestSyncGate|TestVerifySyncGate|TestVerify' ./internal/cli/` → `35 pass`.
- Mutations (reverted): drop `chore: sync` from the Go predicate →
  `codex_sync_gate_test.go:70: subject "chore: sync lockfile": Go predicate = false, script = true`;
  drop the `.vs` directory probe →
  `TestSyncGateLanguageDetectionMatchesScript/visual_studio_dir … languages: Go = [], script = [csharp]`.

**Sub-step 3 — R1 runner `moai verify codex-review`** (`910a212d4`)

- RED (compile): `undefined: codexReviewReceiptState` … `undefined: codexVersionProbe` → build failed.
- GREEN: `go test -json -count=1 -run 'TestVerifyCodexReviewRecordsReceipt|TestVerify|TestCodexReviewGate' ./internal/cli/` → `28 pass`.
  Verdicts: findings → `fail`; clean → `pass`; review call error → `inconclusive` (mirrors
  `codex_review_gate.go` step 5); codex missing → error, nothing recorded.
- Mutation (reverted): record a call error as `pass` →
  `TestVerifyCodexReviewRecordsReceipt/review_call_errors … verdict = pass, want inconclusive` → FAIL.

**Sub-step 4 — the chain runner** (`592ada346`; `internal/cli/codex_stop_chain.go`, wired into
`runHookEvent` for `--harness codex` on Stop; member 1 = the registry dispatch, so a dispatch fault
still reaches the M2c fail-closed writer)

- AC-HPR-002 (`TestStopChainEffectParityGolden`): RED is the compile RED in the resumption note.
  GREEN, final: `go test -json -count=1 -timeout 15m -run '^TestStopChainEffectParityGolden' ./internal/cli/`
  → `36 pass` (goal 4, sync gate 6, codex review gate 7, multi review gate 3, cap 2 × 4, plus
  parents). No `skip`, no `fail`.
- A flake was found and fixed while running the mutations: under load ~45 the cap golden's
  sync-gate Nth Stop read `deny/"unmeasured"` unmutated, because the member hit its 0.5 s internal
  budget and was cut off. The decision goldens now widen the budgets (`budgetFor`, test-only
  `wideStopBudget`) so they test decisions, not timing; the timing leg tests the budgets. A
  cut-off member can no longer write the cap counter after it was abandoned (`capStep`/`capReset`
  check the member context), and the cut-off path reads the tree key without blocking (`TryLock`).
- AC-HPR-002 mutations, each applied, observed red, reverted (verbatim deciding line):
  1. merge drops every member deny (Codex path returns `{}`) →
     `goal/unmet … goal-unmet Codex chain output = {}, want a Stop block`
  2. goal `unmeasured` allows → `goal/receipt_absent … Codex path decision = allow (class "unmeasured", reason ""), Claude path = deny`
  3. sync-gate `unmeasured` allows → `sync_gate/sync-phase_commit,_receipt_absent … Codex path decision = allow (class "unmeasured" …), Claude path = deny`
  4. member-7 missing result writes no discard → `multi_review_gate/result_missing … a missing result must be recorded, got []`
  5. member-7 missing result blocks → `… Codex path decision = deny (…), Claude path = allow`
  6. member 6 allows with codex installed and no receipt → `codex_installed,_no_receipt … Codex path decision = allow …, Claude path = deny` (and `stale_receipt,_HEAD_moved`)
  7. member 6 accepts a receipt recorded under another HEAD (looked up across snapshot keys) →
     `stale_receipt,_HEAD_moved … Codex path decision = allow (class "", reason ""), Claude path = deny`
  8. member 6 blocks when codex is missing → `codex_binary_missing … Codex path decision = deny (…), Claude path = allow`
  9. sync gate requires a receipt on a non-sync HEAD → `HEAD_not_a_sync-phase_commit,_receipt_absent … Codex path decision = deny (class "unmeasured" …), Claude path = allow`
  10. member 6 skips step 2 → `codex_installed,_no_receipt,_stop_hook_active_true … Codex path decision = deny (class "unmeasured" …), Claude path = allow`
  11. sync gate ignores `stop_hook_active` → `fresh_failing_receipt,_stop_hook_active_true … Codex path decision = deny (class "gate_failed" …), Claude path = allow`
  12. cap removed (budgets widened, clean run) → `cap/sync_gate/…Nth_Stop… Stop 3: got deny/"unmeasured", want allow/unverified` and the same for `cap/codex_review_gate`
  13. `unverified` discard suppressed → `… the cap must write exactly one unverified record, wrote 0` (both gates)
  14. reset on a fresh receipt removed → `cap/*/a_fresh_receipt_resets_the_count … after a fresh receipt: got allow/"unverified", want a continuation (count reset to 1)` (both gates). The golden was strengthened first: it now invalidates the receipt **without moving the tree key** (a different `go` on PATH / a different codex version, key asserted equal), so only the receipt read can reset the count.
  15. reset on a key change removed → `cap/*/a_HEAD_change_resets_the_count … got allow/"unverified", want a continuation` (both gates)
- AC-HPR-003 (`TestStopChainGPTProfileNoClaudeDependency`, gpt-profile init, `.claude/` asserted
  absent before and after): written with the chain, passed on its first run →
  `go test -json … -run 'TestStopChainGPTProfileNoClaudeDependency|…'` → `pass`. Mutation
  (reverted): member 2 resolves `.claude/hooks/moai/sync-phase-quality-gate.sh` and allows when it
  is absent → `codex_stop_chain_test.go:101: member 2 (sync-phase quality gate): got allow/"" (MUTATION: …), want deny/"gate_failed"`.
- AC-HPR-005 (`TestStopChainAdvisoryFailureRecorded`): member 4 returns an error, member 5 hangs
  past a 50 ms budget; the merged decision and reason are byte-equal to the unfailed run, and the
  record reads `failed` for both. First run failed on the test's own premise (it re-recorded the
  goal receipt, so the reason's `recorded_at` differed) — fixed in the test, then `pass`. Mutation
  (reverted): record a failed advisory member as `ok` →
  `member 4 recorded "ok", want "failed"` / `member 4 was recorded as passed ("ok") although it failed`.
- Wiring (`TestCodexStopHandlerRunsTheChain`): `moai hook stop --harness codex` with the registry
  returning `{}` and an unmet goal → stdout carries `"decision":"block"`. Mutation standing in for
  the pre-M2d handler (registry dispatch only) → `codex Stop stdout = "{}\n\n", want the goal's block`.
- Cut-off (`TestStopChainGateCutOffNeverAllows`, `629d13eb9`, test-after): goal and gates cut at
  1 ns → deny/`unmeasured` naming the budget; member 7 → fail-open allow + one record. Mutations:
  cut-off goal allows → `member 3 cut off: got allow/"unmeasured" …`; member-7 cut-off without a
  record → `member 7 cut off: got allow/"" with 0 discard(s) …`.

**Sub-step 5 — AC-HPR-016 timing leg** (`59c760ae6`; `TestStopChainMemberCostWithinBudget`,
`./internal/cli/`, 5 runs per member, fresh chain each run so each pays the tree-key computation;
fixture: sync-phase HEAD, dirty tree, both receipts, both review gates on, hook opt-in on,
`MOAI_SECURITY_COMMIT_REVIEW=1`, 200 telemetry records for member 1; member 8 has a positive
control that it did not skip)

| Member | Run 1 (old budgets, load ~48) | Final run (load ~38) | Declared budget (final) |
|---|---|---|---|
| 1 `moai hook stop` | 0 s (no telemetry yet) | 3 ms | 2 s + 0.2 s uncut |
| 2 sync gate (self-gates + compare) | **679 ms > 0.5 s** | 489 ms | **1 s** |
| 3 goal (lookup-only) | 428 ms | 381 ms | **1 s** (was 2 s) |
| 4 security-turn | 79 ms | 77 ms | 0.5 s |
| 5 security-commit (review on) | 152 ms | 153 ms | 0.5 s |
| 6 codex review gate (self-gates + compare) | **721 ms > 0.5 s** | 403 ms | **1 s** |
| 7 multi review gate | 104 ms | 121 ms | 0.5 s |
| 8 harness-observe-stop | 2 ms | 4 ms | 0.5 s |
| whole chain | 1.659 s | 1.083 s | deadline 7.2 s (T_stop 10 s − 2.8 s) |

- Run 1 → `member 2 (sync-phase quality gate) observed max 678.720125ms exceeds its declared budget 500ms`,
  `member 6 (moai hook codex-review-gate) observed max 721.074375ms exceeds its declared budget 500ms` → FAIL.
- Final: `go test -json -count=1 -timeout 15m -run '^TestStopChainMemberCostWithinBudget$' ./internal/cli/`
  → `--- PASS: TestStopChainMemberCostWithinBudget (13.67s)`.
- Mutation (reverted): the telemetry fixture inflated to 400,000 records →
  `member 1 (moai hook stop) observed max 3.632504917s exceeds its declared budget 2.2s` → FAIL
  (member 8 also rose to 1.546 s). In-test: the checker reports an observed max 1 ms over member 7's budget.
- Member 6 in production spawns `codex --version` for its `tool_version` field; the timing leg pins
  the probe. Measured separately (not a Codex session): 5 × `/usr/bin/time -p codex --version` →
  `codex-cli 0.156.1`, `real 0.01` each.
- Sum leg after the rebalance: `go test -json -count=1 ./internal/codexwiring/` → `102 pass`
  (includes `TestStopChainAggregateBudgetFitsTimeout`, `TestStopChainInventoryMatchesClaudeTemplate`).

**Regression and static checks**

- `go test -json -count=1 -timeout 20m -run 'TestHarnessCodex|TestHookFaultInjection|TestHookCmd|TestHookValidEventTypes|StopGoal|TestCodexReviewGate|TestMultiReviewGate|TestStopChain|TestCodexStop|TestSyncGate|TestVerify|TestGoalCancelled|TestNeedsInput|TestCodexBlank|TestCodexAudit' ./internal/cli/`
  → `232 pass`, no `fail`, no `skip` (M2c fault injection and the existing Codex Stop tests unchanged).
- Final batch at `59c760ae6` (`-run 'TestStopChain|TestCodexStop|TestSyncGate|TestVerifySyncGate|TestVerifyCodexReview|TestHookFaultInjection|TestHarnessCodex|TestHookCmd|TestHookValidEventTypes'`)
  → `96 pass`; each M2d test named above reads `pass`.
- Coverage of the new files from that batch (`go tool cover -func`): `run` 91.7%, `merge` 90.9%,
  `goalMember` 94.7%, `syncGateMember` 95.5%, `codexReviewMember` 95.7%, `multiReviewMember` 100%,
  `unmeasured` 100%, `produceSyncGateReceipt` 82.6%, `produceCodexReviewReceipt` 86.4%. The
  cut-off paths (`cutOffUnmeasured` 12.5%, `multiCutOff` 0%) were then covered by
  `TestStopChainGateCutOffNeverAllows`; the package total was not measured (a whole-package
  coverage run of `internal/cli` was not done).
- `golangci-lint run ./internal/cli/ ./internal/codexwiring/...` → `0 issues.`; `go vet ./internal/cli/` → no output;
  `GOOS=windows GOARCH=amd64 go build ./internal/cli/` → no output.

**Gaps and deviations**

- AC-HPR-005's Claude leg is not covered by a MoAI failure record: on Claude each advisory member
  is its own registration, so a failing one surfaces as Claude Code's own non-blocking hook error,
  not a MoAI record. Only the Codex leg is tested.
- Security guardian members 4/5: their opt-in `block` rides `hookSpecificOutput`, which a Stop
  decision does not read, so both harnesses treat it as advisory; the Codex chain carries their
  message as `systemMessage` only. Not measured live.
- The goal member is not under the §D3.8 cap (design). A goal member that is cut off on every
  Stop continues every Stop, and its own turn ceiling advances only if the abandoned evaluation
  saves; the timing leg observed it at ≤ 428 ms against 1 s. Residual risk, not closed.
- A member cut off before its self-gates finish (for example the sync gate's git walk on a very
  large tree) continues the turn (`unmeasured`, bounded by the cap) even where Claude would have
  found the gate not applicable.
- The timing figures come from a machine shared with other lanes; they bound nothing on another
  machine or in CI.
- No live Codex run (Q5): whether Codex honours the merged Stop output end-to-end stays unmeasured.

### Merge repair before M2e (2026-09-26)

HEAD `f75957505` (a clean textual merge of local develop `4dcd4d8d4`) did not compile:
`go vet ./internal/cli/` → `tailString` called with `[]byte` at `codex_audit_live_test.go:305`,
`codex_role_live_test.go:173/240/282`, `factory_live_test.go:486`. develop's
`live_harness_test.go:439` (`d31be8ae4`, card t1100) defines `tailString(b []byte, n int)`; this
branch's `codex_sync_gate.go:381` (`b82562b92`) defined `tailString(s string, n int)`. Commit
`eb23c7aa2` renames the production helper `tailOfString`; the test helper keeps its name.

- `go vet ./internal/cli/ ./internal/codexadapter/ ./internal/codexwiring/` → exit 0, no output;
  `GOOS=windows go vet ./internal/cli/` → exit 0; `GOOS=linux go vet ./internal/cli/` → exit 0.
  The live tests carry `//go:build !windows` / `darwin || linux` / `windows`, so the three GOOS
  runs cover every tag.
- `go test -count=1 ./internal/codexadapter/ ./internal/codexwiring/` →
  `ok … codexadapter 0.580s`, `ok … codexwiring 0.938s`.
- `go test -count=1 -timeout 30m ./internal/cli/` (whole package, started on the repaired tree
  before any M2e edit) → **inconclusive, not a pass**: `panic: test timed out after 30m0s`
  (load 34–80 on the shared machine), with three `--- FAIL` lines, none from the rename:
  `TestCodexSpawn_RealAssemblyThroughStubTmux` (the session's inherited
  `MOAI_KANBAN_BACKEND=claude MOAI_FACTORY_WORKER=agent-43 MOAI_FACTORY_WORKERS=0` leaked into the
  asserted tmux command — an unscrubbed environment), and
  `TestDoctorExitCode_CodexCleanStaysZero` / `…AdvisoryOnlyStaysZero` (they `go build` the live
  tree, which M2e edits were changing mid-run: `undefined: isPermissionRequestDeny`). The
  whole-package verdict is left to the final regression run and CI.

### M2e — event adaptation (2026-09-26)

Measured on Darwin arm64, go1.26.8, against the working tree on top of `eb23c7aa2`. The machine
was shared with other lanes (load 32–80). A diff-stat of `internal/hook`, `.claude/hooks`, and
`internal/template/templates/.claude/hooks` against HEAD printed nothing, and
`grep -rn EventInterrupt internal/hook/*.go` printed nothing (Q6 / HOOK-ADAPTER REQ-7 kept).

**What changed.** All twelve `EventTable` rows are adapted: PreCompact (`compact`), PostCompact
(`post-compact`), PermissionRequest (`permission-request`), and Interrupt with the new Codex-only
dispatcher arg `interrupt`. `moai hook interrupt` (`hook_codex_interrupt.go`) refuses to run
without `--harness codex`, cross-checks the payload event, refuses a missing or path-shaped
`session_id`, appends one record to `.moai/state/codex-interrupt/<session>.jsonl`
(session, transcript path, recorded-at, goal status before, goal `created_at`), and turns that
session's **armed** goal `cancelled` (any other status is recorded, never overwritten). A
PermissionRequest deny from the shared handler is rendered through `TranslateCodex`, so its reason
reaches Codex in `decision.message` (the Claude shape keeps it in `systemMessage`, a key Codex
ignores). `RenderHooks` needed no change: it derives from the table and now installs the
Interrupt handler.

**Decisions recorded**

- **D3 (given):** PreCompact, PostCompact, and PermissionRequest get **no** stderr class;
  `TestExcludedEventsHaveNoClass` is unchanged and passes.
- **PermissionRequest with no opinion stays no opinion on Codex.** The shared handler returns no
  decision unless the tool input carries the updated-input marker; on Claude that leaves Claude's
  own permission flow in front of the user. The Codex path passes the same `{}` through, which
  hands the request to Codex's own approval flow. It is not normalized to `needs_input` (which
  would deny every Codex approval request fail-closed). Basis: REQ-HPR-011 ("the same permission
  decision logic as for Claude"). How Codex resolves `{}` on PermissionRequest is unmeasured
  (research.md H1 family) — carried as residual risk, not claimed.
- **PostCompact memo delivery on Codex.** The restore runs the same handler; its
  `systemMessage` has no measured Codex delivery channel on PostCompact, so the adapter drops it
  **with** a discard record (`key=systemMessage`, content length). The golden asserts the restored
  memo equals the saved memo and that the record exists. Whether the restored text reaches the
  Codex model is a live question (AC-HPR-009 live leg, NOT_RUN).
- **Subcommand count re-measured on the merged tree:** a build of `eb23c7aa2` (exported with
  `git archive` to a scratch directory) lists 43 `moai hook` subcommands in `hook --help`, with no
  `interrupt`; the working-tree build lists 44, and the only difference between the two name
  lists is `> interrupt`. 43 → 44, as plan.md:77 predicted.

**Intentional amendments (plan.md:77, exactly the listed set)**

| Test | Amendment |
|---|---|
| `codexadapter/events_test.go` `TestAdaptedRowCount` | `wantAdapted` 8 → 12 |
| `events_test.go` `TestEventTableMapping` | the four rows flip to adapted; Interrupt carries `interrupt` |
| `events_test.go` `TestResolveRecognizedButUnadapted` | rewritten as `TestResolveFormerlyUnadaptedNowResolve` (the four names resolve) |
| `events_test.go` `TestResolveInterruptNoCounterpart` | rewritten as `TestResolveInterruptResolvesToInterrupt` |
| `codexwiring/hooks_test.go` `TestRenderHooks_InterruptNeverInstalled` | inverted as `TestRenderHooks_InterruptInstalled` (handler rendered once; a user Interrupt entry still preserved) |
| `cli/hook_harness_codex_test.go` `TestHarnessCodexUnadaptedSubcommandRejected` | rewritten as `TestHarnessCodexCompactSubcommandAccepted` — no shipped event stays unadapted, so it cannot be retargeted |
| `cli/hook_test.go` `TestHookCmd_SubcommandCount`, `cli/hook_pre_push_test.go` `TestHookCmd_PrePushSubcommandCount` | 43 → 44, comment line naming this SPEC |
| `cli/hook_e2e_test.go` `TestHookValidEventTypes_AllHaveSubcommands` | `interrupt` joins `utilitySubcmds` with a comment |

The refusal path the rewritten Resolve tests guarded is kept covered by a new
`TestResolveUnadaptedRowIsRefused` over an explicit table (`resolveIn`). Not an amendment: the
design-§D6 consumer inventory `TestGoalStatusConsumersHandleCancelled` failed as designed on the
new producer (`unlisted goal-status site internal/cli/hook_codex_interrupt.go|recordCodexInterrupt
references Goal.Status,StatusArmed,StatusCancelled`, and `…|type interruptRecord references
Status`); the two rows were added to its table — design §D6 names this producer.
`TestDispatcherArgsExist` (not amended) now checks `interrupt` and passes: the subcommand is
registered with a `{"interrupt", "Handle …"}` literal.

**RED**

- codexadapter/codexwiring, the amended tests against the pre-change `events.go` (the edited
  `events.go` set aside for the run, a `resolveIn` shim appended):
  `events_test.go:84: adapted rows = 8, want 12`;
  `events_test.go:63: PreCompact: adapted = false, want true` (and PostCompact, PermissionRequest);
  `events_test.go:60: Interrupt: dispatcher arg = "", want "interrupt"`;
  `events_test.go:151: Resolve(Interrupt) error = codex hook event recognized but not adapted: "Interrupt" (no MoAI dispatcher counterpart; …), want "interrupt"`;
  `hooks_test.go:91: Interrupt event key not rendered into hooks.json — the adapted row must be installed` → FAIL.
  Provenance: the tests were edited first, but the `events.go` edit was written before this RED
  run and set aside for it; the RED is real, the ordering is test-then-code-then-RED.
- cli, with `moai hook interrupt` registered as a no-op stub:
  `codex_event_adaptation_test.go:180: PermissionRequest: deny carries no reason; Codex rejects a blank-reason deny: {"hookSpecificOutput":{"decision":{"behavior":"deny"},"hookEventName":"PermissionRequest"}}`;
  `codex_event_adaptation_test.go:238: cancellation records = [] (err <nil>), want exactly one`;
  `codex_event_adaptation_test.go:275: moai hook interrupt must refuse the Claude harness: Interrupt has no Claude-side counterpart` → FAIL.
  `TestCodexCompactCheckpointRoundTrip` passed at this point because the rows were already
  adapted; its detection is shown by mutation 1.

**GREEN**

- `go test -json -count=1 ./internal/codexadapter/ ./internal/codexwiring/` → `72` pass records
  (codexadapter) and `157` (codexwiring); no `skip`, no `fail`; both packages `pass`.
- `go test -json -count=1 -timeout 20m -run 'TestCodexCompactCheckpointRoundTrip|TestCodexPermissionRequestDenyPreserved|TestCodexInterruptRecordsCancellation|TestHarnessCodex|TestHookCmd_SubcommandCount|TestHookCmd_PrePushSubcommandCount|TestHookValidEventTypes_AllHaveSubcommands|TestHookFaultInjection|TestGoalCancelled' ./internal/cli/`
  → `42 pass`; no `skip`, no `fail`.
- `go test -count=1 ./internal/goal/` (after the two inventory rows) → `ok … goal 3.639s`.

**Mutations (each applied, observed red, reverted)**

1. PreCompact/PostCompact rows back to `false` (AC-HPR-009 "leave the rows unadapted") →
   `codex_event_adaptation_test.go:114: moai hook compact --harness codex: codex harness: rejecting payload: codex hook event recognized but not adapted: "PreCompact" …` → FAIL.
2. PermissionRequest deny routed past the translation (`isPermissionRequestDeny(…) && false`,
   AC-HPR-010 "a pass-through that drops the handler's deny") →
   `codex_event_adaptation_test.go:180: PermissionRequest: deny carries no reason; …` → FAIL.
3. The cancellation record written empty (AC-HPR-011 "skip writing the cancellation record") →
   `codex_event_adaptation_test.go:238: cancellation records = [] (err parse interrupt record: unexpected end of JSON input), want exactly one` → FAIL.

**AC legs**: AC-HPR-009 golden (`TestCodexCompactCheckpointRoundTrip`) PASS; AC-HPR-010 golden
(`TestCodexPermissionRequestDenyPreserved`: marker deny with reason, no-opinion stays no-opinion,
a handler fault on the real subcommand is fail-closed) PASS; AC-HPR-011 unit
(`TestCodexInterruptRecordsCancellation`, incl. the `grep -rn EventInterrupt internal/hook`
invariant as a subtest) PASS. The live legs are M2g (NOT_RUN).

### M2f — goal parity (2026-09-26)

Base HEAD `1fd697bd0` (M2e). Measured on Darwin arm64, go1.26.8, on the shared machine.

**What changed.** The Codex Stop chain's goal member already ran the existing evaluator in
lookup-only mode (M2d). M2f found one defect and closed it: when the goal member allowed, it
recorded `pass` in the per-run verdict record whatever the reason — so a cancelled goal and a goal
ended by its turn ceiling both read as a pass. `goalAllowStatus` (`codex_stop_chain.go`) now reads
the goal state the evaluation just persisted: only `satisfied` records `pass`; `cancelled` records
`cancelled`; `ceiling-exit` / `unsatisfiable` record the new `terminated`; anything else
`not-applicable`. The design-§D6 consumer inventory gained that site as a row (designed-in
maintenance, not an amendment).

**RED**

- Compile: `codex_goal_parity_test.go:282:34: undefined: stopStatusTerminated` → build failed.
- Runtime (constants added, `goalAllowStatus` not yet written):
  `codex_goal_parity_test.go:185: Stop 1: a cancelled goal was recorded "pass"`;
  `codex_goal_parity_test.go:283: a goal ended by its budget was recorded "pass", want "terminated"` → FAIL.
  The same run also failed `TestCodexGoalContinueUntilMet` on the test's own defect (it searched
  the rendered JSON for a command containing `&&`, which JSON escapes as `&&`); the test
  now decodes the reason first. `TestGoalHostOverrideNotSuccess` and the clear leg passed on the
  first run — characterizations of existing behaviour, shown to detect by mutations 4 and 5.
- `TestGoalBudgetTerminationNotSuccess` (`./internal/goal/`) passed on its first run — a
  characterization of the evaluator's existing exits; detection shown by mutation 6.

**GREEN**

- `go test -json -count=1 -timeout 20m -run 'TestCodexGoalContinueUntilMet|TestGoalCancellationPrecedence|TestGoalHostOverrideNotSuccess|TestCodexGoalBudgetTerminationRecorded|TestStopChainEffectParityGolden|TestCodexInterruptRecordsCancellation' ./internal/cli/`
  → `49 pass`; no `skip`, no `fail` (the M2d goldens included, so the record change did not move
  any Claude-vs-Codex decision).
- `go test -count=1 -v -run 'TestGoalBudgetTerminationNotSuccess|TestGoalStatusConsumersHandleCancelled' ./internal/goal/`
  → `--- PASS: TestGoalBudgetTerminationNotSuccess` (turn ceiling, wall-clock bound, stagnation),
  `--- PASS: TestGoalStatusConsumersHandleCancelled`, `ok … goal 3.728s`.

**AC legs**

- AC-HPR-012 golden (`TestCodexGoalContinueUntilMet`): fresh failing receipt → Stop block whose
  reason names the failed condition, recorded `unmet`; tree moved (no matching receipt) → block
  naming `moai verify record`, recorded `unmeasured`, and the sentinel the condition would create
  is absent (nothing executed in the hook); fresh passing receipt → allow, goal `satisfied`,
  recorded `pass`. PASS.
- AC-HPR-013 precedence (`TestGoalCancellationPrecedence`): Interrupt leg — the unmet goal blocks,
  `moai hook interrupt --harness codex` runs, and two later Stops both allow with the goal
  `cancelled`, `TurnsUsed` unchanged (the loop does not resume); Clear leg — `moai goal clear`
  (`runGoalClear`) removes the state, the next Stop allows, no verdict file exists, and the member
  is not recorded `pass`. Neither leg writes a cancellation record by hand. PASS.
- AC-HPR-014 (`TestGoalBudgetTerminationNotSuccess`, `./internal/goal/`) PASS, plus the Codex-chain
  leg `TestCodexGoalBudgetTerminationRecorded` (ceiling 2: turn 1 continues, turn 2 allows, status
  `ceiling-exit`, verdict file persisted, recorded `terminated`) PASS.
- AC-HPR-015 (`TestGoalHostOverrideNotSuccess`): a `stop_hook_active: true` Stop leaves the unmet
  goal unsatisfied on the Codex path (and on the Claude evaluator), and eight blocked Stops — the
  host's consecutive-block cap — leave it unsatisfied. PASS. Measured by the added log line:
  `after 8 host-capped Stops the goal reads "ceiling-exit" after 8 evaluation(s)`.

**The uncapped goal member (M2d residual), re-examined.** The §D3.8 cap does not cover the goal
member by design. The measurement above shows what bounds it instead: with the receipt unchanged,
each evaluation has the same fingerprint, and the stagnation guard (threshold 3) ends the loop as
`ceiling-exit` — the goal stops blocking and is recorded `terminated`, never `pass`. That bound
holds only when the evaluation completes and saves. A goal member cut off by its budget on every
Stop never saves, so neither the turn count nor the fingerprint advances and it continues every
Stop — still open, and still bounded only by the host.

**Mutations (each applied, observed red, reverted; the file restored from the kept copy)**

1. The evaluator re-executes on a receipt miss (`realCmdRunner{}` in place of the lookup-only
   runner; AC-HPR-012) →
   `codex_goal_parity_test.go:130: receipt absent: Codex output = {"decision":"block","reason":"mechanical condition failed: cmd \"touch …/executed && test -f …/done.flag\" exited 1 (want 0): "}, want a continuation naming "moai verify record"` → FAIL.
2. A met goal keeps blocking (AC-HPR-012) →
   `codex_goal_parity_test.go:143: met goal: Codex output = {"decision":"block","reason":"mutant"}, want allow` → FAIL.
3. The Interrupt producer leaves the goal armed (the cancellation branch disabled; AC-HPR-013) →
   `codex_goal_parity_test.go:192: Stop 1 after the interrupt blocked: {"decision":"block","reason":"mechanical condition failed: cmd \"false\" exited 1 (want 0): …"}` → FAIL.
4. `stop_hook_active: true` marks the goal satisfied (AC-HPR-015) →
   `codex_goal_parity_test.go:241: Codex: a stop_hook_active turn marked the unmet goal "satisfied"`;
   `codex_goal_parity_test.go:265: after the host's block cap the goal reads "satisfied", want an unsatisfied status` → FAIL.
5. The goal member records `pass` for every allow (the pre-M2f behaviour) →
   `codex_goal_parity_test.go:198: Stop 1: a cancelled goal was recorded "pass"`;
   `codex_goal_parity_test.go:296: a goal ended by its budget was recorded "pass", want "terminated"` → FAIL.
6. A turn-ceiling exit writes `satisfied` (`evaluate.go`; AC-HPR-014) →
   `budget_termination_test.go:62: a turn ceiling exit wrote status "satisfied"` → FAIL. After the
   revert, a diff of `internal/goal/evaluate.go` against HEAD printed nothing.

### M2g — live legs, built but not run (2026-09-26)

Base HEAD `c60af4998` (M2f). **No live Codex or Claude run was made** (Q5).

**What was built** (`internal/cli/parity_live_test.go`, nine tests, gated by `MOAI_PARITY_LIVE=1`):
`TestLiveStopChainGoalContinuation` (AC-HPR-004), `TestLiveCodexNeedsInputOutcome` (007),
`TestLiveHookFaultOutcome` (008 live), `TestLiveCodexCompactFires` (009 live),
`TestLiveCodexPermissionRequestFires` (010 live), `TestLiveCodexInterruptFires` (011 live),
`TestLiveCodexGoalContinueUntilMet` (012 live), `TestLiveHarnessIsolation` (020),
`TestLiveCodexStopTimeoutCeiling` (021). Shared harness: the gate (switch, `codex` / `claude`
binaries, a login at `~/.codex/auth.json`); a temporary `CODEX_HOME` holding a copy of the login
(`isolatedCodexHome`, reused from the t1100 live harness); a scratch project under the OS temp dir
with `.codex/hooks.json` rendered by `RenderHooks` and a `moai` built from this tree
(`buildLiveMoai`) first on `PATH`; the operator's `~/.codex/config.toml` and `hooks.json` hashed
before and compared at cleanup; a budget of 10 host turns / 45 min and a 5-minute outer bound per
turn; every host process group reaped through `liveProcs`. Goal state is keyed by session, and a
Codex session id is known only after its first Stop, so the Codex goal legs run one turn, read the
session id from the Stop-chain record, arm the goal for it, and continue with `codex exec resume`;
the Claude leg arms first and passes `claude -p --session-id`.

**Rule 8 discipline.** Every refusal writes a verdict record (`NOT_RUN`, with the attempted
command, the observed output, commit, tree digest, `claude --version`, `codex --version`,
`GOOS GOARCH`) to `MOAI_PARITY_VERDICT_DIR` (or the test's temp dir) and calls `t.Skipf`. A live
leg never returns normally without its trigger: AC-HPR-007 has no in-tree handler that emits `ask`
on demand, and the AC-HPR-010 deny leg needs a tool input carrying the updated-input marker, which
a real approval request does not produce — both record `NOT_RUN` with that reason even when the
switch is on.

**Axis declaration.** `codex_live_axis_declaration_test.go` declares `parity_live_test.go`
(switches `MOAI_PARITY_LIVE`, `codexBinaryName`; 9 tests), and its live-file detector now also
matches the `"MOAI_PARITY_LIVE"` literal.

**AC-HPR-020 detector, offline.** `parityIsolationViolations` is exercised without a host by
`TestParityIsolationDetector`, including the AC's two mutations: a Codex process without the
temporary `CODEX_HOME` is flagged, and a run whose working directory is the repository root is
flagged on both counts (outside the temp dir, inside the repository).

**Run once without the switch (the NOT_RUN evidence)**

`go test -json -count=1 -timeout 20m -run 'TestLive(StopChainGoalContinuation|CodexNeedsInputOutcome|HookFaultOutcome|CodexCompactFires|CodexPermissionRequestFires|CodexInterruptFires|CodexGoalContinueUntilMet|HarnessIsolation|CodexStopTimeoutCeiling)$|TestParityIsolationDetector|TestCodexLiveAxis' ./internal/cli/`
→ `skip` for all nine live tests; `pass` for `TestParityIsolationDetector` and the three
`TestCodexLiveAxis_*` guards; no `fail`. Verbatim skip line (AC-HPR-008, the others differ only in
the AC id and the attempted command):
``parity_live_test.go:392: NOT_RUN (AC-HPR-008): attempted: codex exec asking for `touch marker-<fault>` with a PreToolUse handler that faults (timeout, exit 1, unparseable, exit 2); observed: MOAI_PARITY_LIVE is not set to 1 (operator decision Q5: no live run in this SPEC); record: …/TestLiveHookFaultOutcome.json``

Under rule P every live leg is **NOT_RUN** — not PASS. The live bodies have never executed, so
their logic (trigger attempts, evidence reads, the resume flow, the timeout ladder) is compiled
(`go vet` on darwin and `GOOS=windows`) but unverified.

### M2h — coverage and verdict aggregation (2026-09-26)

Base HEAD `92a4ac964` (M2g). Measured on Darwin arm64, go1.26.8.

**What was built.**

- `internal/template/obligations.yaml`, embedded (`//go:embed`) and loaded by
  `LoadObligationRegistry`: the whole catalog (Q1) — **43 rows**: 33 M2 obligations (the eight
  Stop members, the inventory, no-`.claude/` dependency, the Stop-chain live leg, decision
  preservation and the four adapted events with their live legs, goal
  continuation/cancellation/budget/override, receipts, budgets, the timeout ceiling, isolation),
  the 4 non-Stop multi-handler chains as `unverified:` (spec.md §F), and 6 M1 standing-policy rows
  as `blocked:M1` (design report §06/§19: standing, scoped, workflow references, host isolation,
  enforced boundary, render determinism). The Claude interrupt source is `UNSUPPORTED:` on the
  two Interrupt rows. Every row's `check` names a test that exists; the M1 rows and the chain rows
  point at `TestStandingPolicyParity` / `TestNonStopChainsRecordedUnverified`, which assert those
  rows stay blocked / unverified and aggregate below PASS until the owning work replaces them.
- The production path resolver (`SourceResolver`, `IndexCLICommands`): a `moai …` path resolves
  when every command word before the first flag is a cobra `Use:` word or a table-driven hook
  subcommand in the non-test `internal/cli` sources; any other path must name a file or directory
  in the template tree. Static lookup only — nothing is executed.
- `parity_verdict.go`: `AggregateParityVerdict` (per obligation, the weakest of the path markers,
  the check's go-test action, and the verdict record; PASS only with action `pass`, record `PASS`,
  all five attribution fields, and evidence level `effect-verified`; an empty registry is not
  PASS) and `ReadGoTestActions` (rule P over a `go test -json` stream: a skip or fail in a test or
  any subtest outranks its pass; a test with no record is absent, i.e. an empty run).

**RED**

- Compile: `obligation_registry_test.go:19:37: undefined: SourceResolver` …
  `undefined: IndexCLICommands` … `undefined: LoadObligationRegistry` …
  `undefined: AggregateParityVerdict` … `undefined: VerdictUnverified` → build failed. The
  registry/resolver code written ahead of this test was removed (`obligations.go` restored from
  HEAD) before the test was written, then put back.
- Runtime (aggregate and reader as stubs that return PASS / nothing):
  `obligation_registry_test.go:211: skip: aggregate read PASS` (and the other 14 injections);
  `obligation_registry_test.go:227: the whole-catalog registry aggregated PASS with no evidence`;
  `obligation_registry_test.go:254: TestC = "", want "fail"` (and TestA/B/D) → FAIL.

**GREEN**

`go test -json -count=1 -run 'TestObligationCoverage|TestNonStopChainsRecordedUnverified|TestStandingPolicyParity|TestParityVerdictAggregate|TestReadGoTestActions' ./internal/template/`
→ `43 pass`; no `skip`, no `fail` (the M2a schema tests included).

**AC legs**

- AC-HPR-018 (`TestObligationCoverage*`, `./internal/template/`): the embedded registry has no
  coverage violation against the real command set, template tree, and test index; in-test
  mutations on the first row — remove its Codex path, remove its check, point the check at a
  missing test, point the Codex path at an unregistered command — each yield exactly one violation
  naming `stop-chain-inventory`. PASS.
- AC-HPR-019 (`TestParityVerdictAggregate`): a fully effect-verified, attributed registry reads
  PASS; each injection keeps obligation `a` and the aggregate below PASS — skip → NOT_RUN, empty
  run → NOT_RUN, go-test fail → FAIL, NOT_RUN record → NOT_RUN, a `pass` action paired with a
  NOT_RUN record → NOT_RUN, UNSUPPORTED record or path → UNSUPPORTED, `blocked:` path → BLOCKED,
  `unverified:` path → UNVERIFIED, a Stop gate the §D3.8 cap recorded `unverified` → UNVERIFIED,
  a missing attribution field or no record → UNATTRIBUTED, config-existence-only (`registered`)
  or `fired` evidence → INSUFFICIENT_EVIDENCE, obligation absent from the evidence → NOT_RUN. The
  embedded registry with no evidence is not PASS. PASS.

**Mutations (source, each applied, observed red, reverted from the kept copy)**

1. The aggregate treats a NOT_RUN record as a pass (`case VerdictPass, VerdictNotRun:`) →
   `obligation_registry_test.go:233: NOT_RUN record: aggregate read PASS`;
   `obligation_registry_test.go:233: pass action with a NOT_RUN record: aggregate read PASS` → FAIL.
2. The resolver accepts any command word (`if false && !r.Commands[w]`) →
   `obligation_registry_test.go:88: want one violation for stop-chain-inventory mentioning "does not resolve", got []` → FAIL.

**The aggregate for this SPEC is not PASS, and cannot be from this SPEC alone**: 6 rows are
`blocked:M1`, 4 are `unverified`, 2 carry a Claude `UNSUPPORTED` marker, the 9 live-leg checks
skip (NOT_RUN), and no row yet has an effect-verified, attributed verdict record.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
