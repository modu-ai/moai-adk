# Acceptance — SPEC-DUAL-HARNESS-HOOK-PARITY-001

## §A Scope of Verification

Twenty acceptance criteria, grouped under the four design acceptance families (design §19):
AC-HOOK-01 → AC-HPR-001..005, AC-HOOK-02 → AC-HPR-006..011, AC-GOAL-01 → AC-HPR-012..015,
receipts → AC-HPR-016..017, AC-POL-01 → AC-HPR-018..019, isolation → AC-HPR-020.

Test names below are the run-phase deliverables; they do not exist at HEAD `530d8cc06`.
The command column is the exact command the verifier runs.

## §B HARD evidence rules (bind every AC)

1. **PASS rule (P).** A Go-test-backed AC passes only when `go test -json` emits a record with
   `"Action":"pass"` for the named test AND no record with `"Action":"skip"` for that test or any
   of its subtests. Canonical check:
   ```bash
   go test -json -count=1 -run '<regex>' <pkg> > "$OUT"; \
     jq -r 'select(.Test!=null and (.Test|test("<regex-root>"))) | .Action' "$OUT" | sort -u
   ```
   Output must contain `pass`, must not contain `skip` or `fail`. An exit code of 0 alone is not
   evidence (a fully skipped run also exits 0).
2. **Empty run is not PASS.** `ok … [no tests to run]` or a test2json stream with zero records for
   the named test is `NOT_RUN`.
3. **Config existence is not PASS.** A rendered `.codex/hooks.json` or `.claude/settings.json`
   entry proves registration only. Firing and effect are separate evidence levels
   (registered → fired → effect-verified); only effect-verified satisfies a live AC.
4. **Live NOT_RUN is never PASS.** A live AC whose host, credentials, or trigger was unavailable
   records `NOT_RUN` with the attempted command and observed output (REQ-HPR-013).
5. **UNSUPPORTED is not PASS.** Where a host cannot express a requirement, the AC records
   `UNSUPPORTED` with evidence; the aggregate verdict (AC-HPR-019) is then FAIL for that
   obligation.
6. **Every AC names a negative/mutation case.** The mutation must be applied on a scratch copy or
   inside the test (never committed) and must turn the AC's check red.
7. **Attribution.** Every recorded verdict carries commit, working-tree digest
   (`moai verify check --key-current` or the `internal/verify` `Key` value), `claude --version`,
   `codex --version`, and `uname -sm` (REQ-HPR-024).

## §C AC Matrix

| AC | Family | REQ | Kind | Verification command | Negative / mutation |
|---|---|---|---|---|---|
| AC-HPR-001 | HOOK-01 | REQ-HPR-001 | unit | `go test -json -count=1 -run '^TestStopChainInventoryMatchesClaudeTemplate$' ./internal/codexwiring/` | add a handler to the template Stop array without an inventory row → fails naming it |
| AC-HPR-002 | HOOK-01 | REQ-HPR-002, REQ-HPR-005 | golden | `go test -json -count=1 -run '^TestStopChainEffectParityGolden' ./internal/cli/` | make the Codex path return `{}` for the goal-unmet golden → fails |
| AC-HPR-003 | HOOK-01 | REQ-HPR-003 | golden | `go test -json -count=1 -run '^TestStopChainGPTProfileNoClaudeDependency$' ./internal/cli/` | point one member at a `.claude/hooks/` path → fails naming the member |
| AC-HPR-004 | HOOK-01 | REQ-HPR-002, REQ-HPR-013 | live | `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveStopChainGoalContinuation' ./internal/cli/` | goal-met fixture must stop (no continuation) |
| AC-HPR-005 | HOOK-01 | REQ-HPR-004 | golden | `go test -json -count=1 -run '^TestStopChainAdvisoryFailureRecorded$' ./internal/cli/` | advisory member returns error → chain continues AND record reads failed, not passed |
| AC-HPR-006 | HOOK-02 | REQ-HPR-006, REQ-HPR-007 | unit | `go test -json -count=1 -run '^TestDecisionTranslationNeverLoosens' ./internal/codexadapter/` | restore the `ask` → `{}` drop branch → fails |
| AC-HPR-007 | HOOK-02 | REQ-HPR-007, REQ-HPR-008 | live | `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexNeedsInputOutcome' ./internal/cli/` | each supported Codex approval policy; an observed allow → `UNSUPPORTED` + mitigation row |
| AC-HPR-008 | HOOK-02 | REQ-HPR-009 | unit + live | `go test -json -count=1 -run '^TestHookFaultInjection' ./internal/cli/` and `MOAI_PARITY_LIVE=1 … -run '^TestLiveHookFaultOutcome'` | timeout, exit 1, corrupt stdout, exit 2 on PreToolUse / PermissionRequest / Stop gate |
| AC-HPR-009 | HOOK-02 | REQ-HPR-010, REQ-HPR-013 | golden + live | `go test -json -count=1 -run '^TestCodexCompactCheckpointRoundTrip$' ./internal/cli/` | corrupt the saved memo → restore mismatch detected |
| AC-HPR-010 | HOOK-02 | REQ-HPR-011, REQ-HPR-013 | golden + live | `go test -json -count=1 -run '^TestCodexPermissionRequestDenyPreserved$' ./internal/cli/` | updatedInput-marker input must deny on both harnesses |
| AC-HPR-011 | HOOK-02 | REQ-HPR-012, REQ-HPR-013 | unit + live | `go test -json -count=1 -run '^TestCodexInterruptRecordsCancellation$' ./internal/cli/` | `grep -rn 'EventInterrupt' internal/hook/*.go` must print nothing |
| AC-HPR-012 | GOAL-01 | REQ-HPR-014 | golden + live | `go test -json -count=1 -run '^TestCodexGoalContinueUntilMet$' ./internal/cli/` | goal unmet → block; flip to met → allow + `satisfied` |
| AC-HPR-013 | GOAL-01 | REQ-HPR-015 | golden | `go test -json -count=1 -run '^TestGoalCancellationPrecedence$' ./internal/cli/` | cancellation + unmet goal must NOT block |
| AC-HPR-014 | GOAL-01 | REQ-HPR-016 | golden | `go test -json -count=1 -run '^TestGoalBudgetTerminationNotSuccess$' ./internal/goal/` | ceiling/wall-clock/stagnation each → status ≠ `satisfied` |
| AC-HPR-015 | GOAL-01 | REQ-HPR-017 | golden | `go test -json -count=1 -run '^TestGoalHostOverrideNotSuccess$' ./internal/cli/` | `stop_hook_active:true` and block-cap stop → status stays non-satisfied |
| AC-HPR-016 | Receipt | REQ-HPR-018 | unit | `go test -json -count=1 -run '^TestRenderedTimeoutCoversMemberBudget$' ./internal/codexwiring/` | a member budget above its rendered timeout → fails |
| AC-HPR-017 | Receipt | REQ-HPR-019 | unit | `go test -json -count=1 -run '^TestCheckReceipt' ./internal/verify/` | mutate each of the 5 fields in turn; delete the receipt |
| AC-HPR-018 | POL-01 | REQ-HPR-020, REQ-HPR-021 | unit | `go test -json -count=1 -run '^TestObligationCoverage' ./internal/<registry-pkg>/` | remove one obligation's Codex path, then its check → each fails naming the id |
| AC-HPR-019 | POL-01 | REQ-HPR-022, REQ-HPR-023, REQ-HPR-024 | unit | `go test -json -count=1 -run '^TestParityVerdictAggregate' ./internal/<registry-pkg>/` | inject skip, NOT_RUN, UNSUPPORTED, missing attribution → aggregate ≠ PASS |
| AC-HPR-020 | Isolation | REQ-HPR-025 | live | `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexHomeIsolation$' ./internal/cli/` | hash of `~/.codex` (config + hooks) before/after must be equal |

`<registry-pkg>` is fixed by design.md §D4 (plan.md Q1 governs its scope); the verifier substitutes the decided path.

## §D Acceptance Criteria (Given-When-Then)

### AC-HOOK-01 — Stop chain effect equivalence

- **AC-HPR-001** — **Given** the distributed `settings.json.tmpl` Stop array, **When** the inventory test parses it, **Then** every registered handler appears exactly once in the Stop-chain inventory with a class (`required-gate` / `goal` / `advisory`) and a Codex path or `UNSUPPORTED` record, and the inventory has no row absent from the template.
- **AC-HPR-002** — **Given** event-input goldens for Stop (goal unmet, goal met, no goal, sync gate failing check, sync gate passing check, sync gate with the blocking opt-out, review gate block, review gate allow), **When** each golden is run through the Claude path and the Codex path with the same project configuration, **Then** the normalized decision and continuation reason class are equal for every `required-gate` and `goal` member.
- **AC-HPR-003** — **Given** a project deployed with the `gpt` profile into a temp directory, **When** the Codex Stop chain runs on the goal-unmet and gate-failing goldens, **Then** each required member executes and returns its decision, and no member resolves a path under `.claude/`.
- **AC-HPR-004** — **Given** a scratch project with an armed, unmet mechanical goal, **When** a real turn ends in Claude Code (`claude -p`) and in Codex (temporary `CODEX_HOME`), **Then** each host continues the turn with the goal's reason and the captured hook log shows the goal member fired; **And Given** the goal is met, **Then** each host stops. A host that cannot be run yields `NOT_RUN`, not PASS.
- **AC-HPR-005** — **Given** an advisory member forced to fail, **When** the Stop chain runs on either harness, **Then** the decision is unaffected, a failure record for that member exists, and no record marks it passed.

### AC-HOOK-02 — decision preservation and event adaptation

- **AC-HPR-006** — **Given** the translation table for every decision-bearing event, **When** the property test feeds every normalized decision, **Then** no `deny` or `needs_input` input produces `allow`, an empty object, or any output in the table's host-resolves-as-allow set.
- **AC-HPR-007** — **Given** a PreToolUse hook returning `needs_input`, **When** Codex runs under each approval policy MoAI supports, **Then** the tool call does not execute without user input; any policy where it does is recorded `UNSUPPORTED` with the observed output and the fail-closed mitigation.
- **AC-HPR-008** — **Given** injected faults (timeout past the registered budget, exit 1, unparseable stdout, exit 2), **When** they occur on PreToolUse, PermissionRequest, and the Stop required gate, **Then** the adapter never emits an allow for them, and the observed host outcome for each fault × event × harness is recorded; any host-side allow is `UNSUPPORTED`.
- **AC-HPR-009** — **Given** the PreCompact/PostCompact rows are adapted, **When** a golden PreCompact payload is followed by a PostCompact payload on the Codex path, **Then** the restored memo equals the saved memo; the live compaction trigger result is effect-verified or `NOT_RUN`.
- **AC-HPR-010** — **Given** the PermissionRequest row is adapted, **When** a payload whose tool input carries the updated-input marker arrives on the Codex path, **Then** Codex receives a deny with a non-empty reason; the live trigger result is effect-verified or `NOT_RUN`.
- **AC-HPR-011** — **Given** the Interrupt row is adapted, **When** a Codex Interrupt payload arrives, **Then** a cancellation record bound to the session and run is written, and `internal/hook` carries no Interrupt constant; the live SIGINT result is effect-verified or `NOT_RUN`.

### AC-GOAL-01 — goal continuation, cancellation, budget

- **AC-HPR-012** — **Given** an armed goal with an unmet condition, **When** Codex Stop fires, **Then** the output requests continuation with the evaluator's reason; **When** the condition becomes satisfied, **Then** the next Stop allows and the goal state reads `satisfied`.
- **AC-HPR-013** — **Given** an armed unmet goal and a recorded cancellation, **When** Stop fires on either harness, **Then** the stop is allowed, the goal state records the cancellation (not `satisfied`), and a subsequent Stop does not resume the loop.
- **AC-HPR-014** — **Given** a goal at its turn ceiling, wall-clock bound, or stagnation limit, **When** Stop fires, **Then** the loop terminates, a verdict is persisted, and the status is not `satisfied`.
- **AC-HPR-015** — **Given** an unmet goal, **When** the host stops despite a block (`stop_hook_active:true`, or the consecutive-block cap), **Then** the goal state is not `satisfied`.

### Receipts

- **AC-HPR-016** — **Given** the rendered Codex `hooks.json`, **When** each handler's registered timeout is compared with the declared internal budget of the member it runs, **Then** no member's budget exceeds its timeout.
- **AC-HPR-017** — **Given** a receipt written after a passing out-of-hook check, **When** the Stop chain evaluates it unchanged, **Then** it is accepted; **When** HEAD, working-tree digest, configuration digest, command, or tool version differs (each mutated separately), or the receipt is absent, **Then** the check is treated as not run.

### AC-POL-01 — obligation coverage

- **AC-HPR-018** — **Given** the obligation registry, **When** the coverage check runs, **Then** it passes only if every required obligation has a Claude path, a Codex path (or an explicit `UNSUPPORTED`/`blocked` marker), and a check id that resolves to an existing test; **When** one obligation's Codex path or check is removed, **Then** it fails naming that obligation.
- **AC-HPR-019** — **Given** per-obligation verdicts, **When** the aggregate is computed, **Then** it is PASS only if every required obligation is effect-verified with full attribution; a skip, an empty run, a `NOT_RUN`, an `UNSUPPORTED`, a `blocked`, or a missing attribution field makes it not PASS.

### Isolation

- **AC-HPR-020** — **Given** the live suite, **When** it runs, **Then** every Codex process sees a temporary `CODEX_HOME`, and the content hash of the operator's `~/.codex` config and hooks is identical before and after.

## §E Edge Cases

- Two handlers on one Codex event both returning a decision: the merge rule is measured, not assumed (design.md §D2).
- Stop fired while a receipt is being written: a partially written receipt is absent, not valid.
- Goal cancellation recorded by a different session: must not cancel this session's goal.
- `needs_input` on an event with an additionalContext channel (UserPromptSubmit): still fail-closed; additionalContext is not an approval channel.
- codex-cli version differs from the measured basis: verdicts are re-measured, not carried forward.

## §F Quality Gates

- `go vet` and `golangci-lint run` clean on changed packages.
- Changed-package coverage ≥ 85%.
- `moai spec lint` on this SPEC reports zero errors, and the REQ collector sees all 25 REQs (verify the count, not only the green).
- CI full suite green on the integration head.

## §G Definition of Done

- AC-HPR-001..020 each carry a recorded verdict with attribution in progress.md §E.2.
- Every unit/golden AC is PASS under rule P.
- Every live AC is effect-verified, or is `NOT_RUN`/`UNSUPPORTED` with evidence — and in that case the SPEC closes with the aggregate verdict reported as FAIL for the affected obligations, not as full parity.
- No claim of "full dual-harness parity" is made from this SPEC alone; sibling card t1100 and M1/M5 remain.

## §H RED-now Evidence Ledger (plan phase)

Document-level pin: every entry below was measured on tree `530d8cc067765a3cf6ac49a76954d34a2193c7d0`
(worktree `dual-harness-parity-rebuild`) on 2026-09-23. Each entry records the command, its
verbatim stdout, and its exit code as separate fields. The green-path column names the milestone
expected to flip the observation.

| Ledger id | Serves AC | Command | Verbatim stdout | Exit | Why this is RED | Green path |
|---|---|---|---|---|---|---|
| L-01 | AC-HPR-002, 003, 012 | `grep -n stop-goal internal/codexwiring/hooks.go` | (empty) | 1 | the goal member is not on the Codex Stop path at all | M2d/M2f: the Codex Stop chain invokes the goal evaluator; the AC test asserts it |
| L-02 | AC-HPR-006 | `grep -n '"ask": *true' internal/codexadapter/output.go` | `40:	"ask":   true,` | 0 | `ask` is in the drop set, so it degrades to `{}` | M2c: `ask` leaves the drop set; the property test fails if it returns |
| L-03 | AC-HPR-009, 010, 011 | `grep -n "EventPreCompact\|EventPostCompact\|EventPermissionRequest\|CodexEventInterrupt, " internal/codexadapter/events.go` | `79: {hook.EventPreCompact, "compact", false},` · `80: {hook.EventPostCompact, "post-compact", false},` · `81: {hook.EventPermissionRequest, "permission-request", false},` · `83: {CodexEventInterrupt, "", false},` | 0 | all four rows unadapted (`false`) | M2e: rows adapted; round-trip, deny, and cancellation tests assert the effect |
| L-04 | AC-HPR-013 | `grep -n cancelled internal/goal/schema.go` | (empty) | 1 | no cancellation status exists | M2a (Q3): status added; precedence test asserts it |
| L-05 | AC-HPR-016 | `grep -n "defaultHandlerTimeout = " internal/codexwiring/codexwiring.go` ; `grep -n "stopGoalHookTimeout = " internal/cli/hook_stop_goal.go` | `60:	defaultHandlerTimeout = 10` · `20:const stopGoalHookTimeout = 90 * time.Second` | 0 · 0 | a 90 s member budget sits under a 10 s handler timeout | M2b/M2d: the budget test fails on any member above its timeout |
| L-06 | AC-HPR-017 | `grep -rln ToolVersion internal/verify` | (empty) | 1 | the snapshot record carries no tool-version field | M2b: receipt carries all 5 fields; per-field mutation test |
| L-07 | AC-HPR-018 | `ls internal/template/obligations.yaml` | `ls: internal/template/obligations.yaml: No such file or directory` | 1 | no registry exists | M2a/M2h: registry + coverage check |
| L-08 | rule P (all Go ACs) | `go test -count=1 -run '^TestStopChainInventoryMatchesClaudeTemplate$' ./internal/codexwiring/` | `ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.479s [no tests to run]` | 0 | exit 0 on an empty sweep: the named test does not exist yet | the test lands in M2d; until then this reads `NOT_RUN`, never PASS (§B.2) |

L-03 stdout is shown with the leading tab collapsed to one space for table rendering; the raw
output carries a tab after the colon.

The live ACs (AC-HPR-004, 007, the live legs of 008–011, and 020) have no RED-now cell that can be
re-executed at plan time without a live host run. They are classified **regression-guard /
NOT_RUN** until M2g records a first observation (verification-completeness §2.1, undecidable
disposition).
