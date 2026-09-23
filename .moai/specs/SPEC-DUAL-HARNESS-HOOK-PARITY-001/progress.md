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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
