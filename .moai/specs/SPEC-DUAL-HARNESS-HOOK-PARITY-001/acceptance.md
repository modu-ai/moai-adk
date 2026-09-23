# Acceptance — SPEC-DUAL-HARNESS-HOOK-PARITY-001

## §A Scope of Verification

Twenty-two acceptance criteria, grouped under the four design acceptance families (design §19):
AC-HOOK-01 → AC-HPR-001..005, AC-HOOK-02 → AC-HPR-006..011, AC-GOAL-01 → AC-HPR-012..015,
receipts and timeouts → AC-HPR-016, 017, 021, the visible needs_input deny → AC-HPR-022, AC-POL-01 → AC-HPR-018..019, isolation → AC-HPR-020.

Test names below are the run-phase deliverables; they do not exist at HEAD `530d8cc06`.
The command column is the exact command the verifier runs. Each AC with a live leg names two
commands: a golden/unit command, and a live command gated by `MOAI_PARITY_LIVE=1`.

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
6. **Every AC names a mutation.** The mutation changes the thing under test (code, fixture, or
   environment) so that a correct check turns red. It is applied inside the test or on a scratch
   copy, never committed. A second positive case is not a mutation.
7. **Attribution.** Every recorded verdict carries commit, working-tree digest
   (`moai verify check --key-current` or the `internal/verify` `Key` value), `claude --version`,
   `codex --version`, and `uname -sm` (REQ-HPR-024).
8. **A live test never returns normally without its trigger.** When a live test cannot achieve
   its trigger (no compaction, no approval request, SIGINT not delivered, a policy that could not
   be set, host binary or credentials missing), it MUST call `t.Skipf` with the attempted command
   and the observed output, so rule P reads `NOT_RUN`. It also writes a verdict record. The
   aggregate reads **both** the go-test action and the record, and takes the weaker: a `pass`
   action paired with a `NOT_RUN` record is `NOT_RUN`. AC-HPR-019 carries this as a mutation.

## §C AC Matrix

| AC | Family | REQ | Kind | Verification command(s) | Mutation |
|---|---|---|---|---|---|
| AC-HPR-001 | HOOK-01 | REQ-HPR-001 | unit | `go test -json -count=1 -run '^TestStopChainInventoryMatchesClaudeTemplate$' ./internal/codexwiring/` | add a handler to the template Stop array without an inventory row → fails naming it; render with `HookOptIn.Enabled=true` against an inventory lacking `harness-observe-stop` → fails |
| AC-HPR-002 | HOOK-01 | REQ-HPR-002, REQ-HPR-005 | golden | `go test -json -count=1 -run '^TestStopChainEffectParityGolden' ./internal/cli/` | make the Codex path return `{}` for the goal-unmet golden → fails; make the Codex goal or sync-gate `unmeasured` case allow the stop → fails; make a Codex review-gate missing-result allow write no discard record → fails; make a Codex review-gate missing-result case block → fails (Claude allows, `multi_review_gate.go:79`) |
| AC-HPR-003 | HOOK-01 | REQ-HPR-003 | golden | `go test -json -count=1 -run '^TestStopChainGPTProfileNoClaudeDependency$' ./internal/cli/` | point one member at a `.claude/hooks/` path → fails naming the member |
| AC-HPR-004 | HOOK-01 | REQ-HPR-002, REQ-HPR-013 | live | `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveStopChainGoalContinuation$' ./internal/cli/` | make the Codex chain skip the goal member → the unmet-goal turn stops and the test fails |
| AC-HPR-005 | HOOK-01 | REQ-HPR-004 | golden | `go test -json -count=1 -run '^TestStopChainAdvisoryFailureRecorded$' ./internal/cli/` | make the recorder mark a failed advisory member as passed → fails |
| AC-HPR-006 | HOOK-02 | REQ-HPR-006, REQ-HPR-007 | unit | `go test -json -count=1 -run '^TestDecisionTranslationNeverLoosens' ./internal/codexadapter/` | restore the `ask` → `{}` drop branch → fails |
| AC-HPR-007 | HOOK-02 | REQ-HPR-007, REQ-HPR-008 | live | `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexNeedsInputOutcome$' ./internal/cli/` | build the adapter with the old drop-to-`{}` translation → under a non-prompting policy the tool executes and the test fails |
| AC-HPR-008 | HOOK-02 | REQ-HPR-009 | unit + live | unit: `go test -json -count=1 -run '^TestHookFaultInjection' ./internal/cli/` · live: `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveHookFaultOutcome$' ./internal/cli/` | make the Codex output path emit `{}` on a parse error → the unit test fails |
| AC-HPR-009 | HOOK-02 | REQ-HPR-010, REQ-HPR-013 | golden + live | golden: `go test -json -count=1 -run '^TestCodexCompactCheckpointRoundTrip$' ./internal/cli/` · live: `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexCompactFires$' ./internal/cli/` | corrupt the saved memo → restore mismatch detected; leave the rows unadapted → the golden fails |
| AC-HPR-010 | HOOK-02 | REQ-HPR-011, REQ-HPR-013 | golden + live | golden: `go test -json -count=1 -run '^TestCodexPermissionRequestDenyPreserved$' ./internal/cli/` · live: `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexPermissionRequestFires$' ./internal/cli/` | route PermissionRequest through a pass-through that drops the handler's deny → fails |
| AC-HPR-011 | HOOK-02 | REQ-HPR-012, REQ-HPR-013 | unit + live | unit: `go test -json -count=1 -run '^TestCodexInterruptRecordsCancellation$' ./internal/cli/` · live: `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexInterruptFires$' ./internal/cli/` | skip writing the cancellation record → fails. Invariant check alongside: `grep -rn 'EventInterrupt' internal/hook/*.go` prints nothing |
| AC-HPR-012 | GOAL-01 | REQ-HPR-014 | golden + live | golden: `go test -json -count=1 -run '^TestCodexGoalContinueUntilMet$' ./internal/cli/` · live: `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexGoalContinueUntilMet$' ./internal/cli/` | make the evaluator re-execute on a receipt miss instead of returning `unmeasured` → fails; make a met goal keep blocking → fails |
| AC-HPR-013 | GOAL-01 | REQ-HPR-015 | golden | `go test -json -count=1 -run '^TestGoalCancellationPrecedence$' ./internal/cli/` · `go test -json -count=1 -run '^TestGoalStatusConsumersHandleCancelled$' ./internal/goal/` | disable the cancellation branch in the Stop chain → the unmet goal blocks and the test fails; remove `cancelled` from the evaluator's early-return set (`evaluate.go:294`) → fails; write an unknown status → a silent block or `satisfied` fails the test |
| AC-HPR-014 | GOAL-01 | REQ-HPR-016 | golden | `go test -json -count=1 -run '^TestGoalBudgetTerminationNotSuccess$' ./internal/goal/` | make a ceiling exit write `satisfied` → fails |
| AC-HPR-015 | GOAL-01 | REQ-HPR-017 | golden | `go test -json -count=1 -run '^TestGoalHostOverrideNotSuccess$' ./internal/cli/` | make `stop_hook_active:true` mark the goal `satisfied` → fails |
| AC-HPR-016 | Timeout | REQ-HPR-018 | unit | sum leg: `go test -json -count=1 -run '^TestStopChainAggregateBudgetFitsTimeout$' ./internal/codexwiring/` · timing leg: `go test -json -count=1 -run '^TestStopChainMemberCostWithinBudget$' ./internal/cli/` | sum leg: raise one in-hook member's declared budget by 1 s past the aggregate → fails. Timing leg: inflate one member's fixture cost past its declared budget (for example a telemetry fixture large enough to slow the member 1 prune) → fails |
| AC-HPR-017 | Receipt | REQ-HPR-019 | unit | `go test -json -count=1 -run '^TestCheckReceipt' ./internal/verify/` | mutate each of the 5 fields in turn; delete the receipt; truncate it mid-write |
| AC-HPR-018 | POL-01 | REQ-HPR-020, REQ-HPR-021 | unit | `go test -json -count=1 -run '^TestObligationCoverage' ./internal/<registry-pkg>/` | remove one obligation's Codex path, then its check → each fails naming the id |
| AC-HPR-019 | POL-01 | REQ-HPR-022, REQ-HPR-023, REQ-HPR-024 | unit | `go test -json -count=1 -run '^TestParityVerdictAggregate' ./internal/<registry-pkg>/` | inject each of: skip; empty run; `NOT_RUN`; `UNSUPPORTED`; missing attribution field; config-existence-only evidence; a `pass` action paired with a `NOT_RUN` record → each makes the aggregate not PASS |
| AC-HPR-020 | Isolation | REQ-HPR-025 | live | `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveHarnessIsolation$' ./internal/cli/` | start one Codex process without the temporary `CODEX_HOME` → the detector fails; start one Claude run with cwd at the repository root → the detector fails |
| AC-HPR-021 | Timeout | REQ-HPR-018 | live (measurement) | `MOAI_PARITY_LIVE=1 go test -json -count=1 -run '^TestLiveCodexStopTimeoutCeiling$' ./internal/cli/` | negative control: a handler that sleeps past its declared timeout must be observed killed (or the non-kill recorded as a host fact); a probe that records a ceiling without running the sleeper fails |
| AC-HPR-022 | HOOK-02 | REQ-HPR-008 | unit | `go test -json -count=1 -run '^TestNeedsInputVisibleDeny$' ./internal/codexadapter/` | suppress the discard write for the needs_input conversion → fails; drop the required-input name from the deny reason → fails; let `needs_input` fall back to `{}` → fails |

`<registry-pkg>` is fixed by design.md §D4 (plan.md Q1 governs its scope); the verifier substitutes the decided path.

## §D Acceptance Criteria (Given-When-Then)

### AC-HOOK-01 — Stop chain effect equivalence

- **AC-HPR-001** — **Given** the distributed `settings.json.tmpl`, rendered twice (`HookOptIn.Enabled` false and true), **When** the inventory test reads each rendered Stop array, **Then** every registered handler appears exactly once in the Stop-chain inventory with a class (`required-gate` / `fail-open-on-missing` / `goal` / `advisory`) and a Codex placement or `UNSUPPORTED` record; the opt-in-only member is flagged as conditional; and the inventory has no row absent from both renders.
- **AC-HPR-002** — **Given** event-input goldens for Stop (goal unmet, goal met, goal receipt absent, no goal, sync gate failing check, sync gate passing check, sync gate receipt absent, sync gate with the blocking opt-out, review gate block, review gate allow, **codex review gate result missing**, **multi review gate result missing**), **When** each golden is run through the Claude path and the Codex path with the same project configuration, **Then** the normalized decision is equal for every `required-gate`, `fail-open-on-missing`, and `goal` member, and the reason class is identical or paired by the design.md §D3.4 mapping. For a missing review-gate result both harnesses allow (Claude: `multi_review_gate.go:47, :79`, `codex_review_gate.go:59`), and the Codex output also carries a discard record and reason text naming the missing result.
- **AC-HPR-003** — **Given** a project deployed with the `gpt` profile into a temp directory, **When** the Codex Stop chain runs on the goal-unmet and gate-failing goldens, **Then** each required member executes and returns its decision, and no member resolves a path under `.claude/`.
- **AC-HPR-004** — **Given** a scratch project with an armed, unmet mechanical goal whose receipt the agent records during the turn, **When** a real turn ends in Claude Code (`claude -p`) and in Codex (temporary `CODEX_HOME`), **Then** each host continues the turn with the goal's reason and the captured hook log shows the goal member fired; **And Given** the goal is met, **Then** each host stops. A host that cannot be run makes the test `t.Skipf` (rule 8).
- **AC-HPR-005** — **Given** an advisory member forced to fail or to exceed its internal budget, **When** the Stop chain runs on either harness, **Then** the decision is unaffected, a failure record for that member exists, and no record marks it passed.

### AC-HOOK-02 — decision preservation and event adaptation

- **AC-HPR-006** — **Given** the translation table for every decision-bearing event, **When** the property test feeds every normalized decision, **Then** no `deny` or `needs_input` input produces `allow`, an empty object, or any output in the table's host-resolves-as-allow set.
- **AC-HPR-007** — **Given** a PreToolUse hook returning `needs_input`, **When** Codex runs under each approval policy MoAI supports, **Then** the tool call does not execute without user input; any policy where it does is recorded `UNSUPPORTED` with the observed output and the fail-closed mitigation.
- **AC-HPR-008** — **Given** injected faults (timeout past the registered budget, exit 1, unparseable stdout, exit 2), **When** they occur on PreToolUse, PermissionRequest, and the Stop required gate, **Then** the adapter never emits an allow for them, and the observed host outcome for each fault × event × harness is recorded; any host-side allow is `UNSUPPORTED`.
- **AC-HPR-009** — **Given** the PreCompact/PostCompact rows are adapted, **When** a golden PreCompact payload is followed by a PostCompact payload on the Codex path, **Then** the restored memo equals the saved memo; **And When** the live test drives Codex to a compaction, **Then** both events are observed firing with the memo round-trip, or the test `t.Skipf`s with the attempted trigger.
- **AC-HPR-010** — **Given** the PermissionRequest row is adapted, **When** a payload whose tool input carries the updated-input marker arrives on the Codex path, **Then** Codex receives a deny with a non-empty reason; **And When** the live test drives Codex to an approval request, **Then** the deny is observed taking effect, or the test `t.Skipf`s with the attempted trigger.
- **AC-HPR-011** — **Given** the Interrupt row is adapted, **When** a Codex Interrupt payload arrives, **Then** a cancellation record bound to the session and run is written, and `internal/hook` carries no Interrupt constant; **And When** the live test delivers SIGINT to a Codex turn, **Then** the record is observed, or the test `t.Skipf`s with the attempted delivery.

### AC-GOAL-01 — goal continuation, cancellation, budget

- **AC-HPR-012** — **Given** an armed goal with an unmet condition and a fresh failing receipt, **When** Codex Stop fires, **Then** the output requests continuation with the evaluator's reason; **When** the receipt is absent, **Then** the output requests continuation naming the command to run (`unmeasured`), and nothing is re-executed in the hook; **When** a fresh passing receipt exists, **Then** the next Stop allows and the goal state reads `satisfied`.
- **AC-HPR-013** — **Given** an armed unmet goal, **When** the cancellation is produced by a real producer, **Then**:
  - Interrupt leg: after the Codex Interrupt golden payload is fed to `moai hook interrupt --harness codex` and Stop then fires, the stop is allowed, the goal state reads `cancelled`, and a subsequent Stop does not resume the loop.
  - Clear leg: after `moai goal clear`, which removes the goal state (`internal/goal/state.go:120–127`), Stop does not block and no record reads `satisfied`.
  - Neither leg injects a pre-written cancellation record.
  - Every non-test reader of goal status handles `cancelled` explicitly. The consumer inventory (design.md §D6, research.md §R1.10) is enumerated by `TestGoalStatusConsumersHandleCancelled`. The test scans the non-test files of `internal/goal`, `internal/cli`, and `internal/hook`, and fails when a site that reads or writes `goal.Status`, or calls `goal.ClearGoal`, is not listed.
  - A goal file carrying an unrecognised status yields a visible diagnostic. It never yields a silent block and never reads as `satisfied`.
- **AC-HPR-014** — **Given** a goal at its turn ceiling, wall-clock bound, or stagnation limit, **When** Stop fires, **Then** the loop terminates, a verdict is persisted, and the status is not `satisfied`.
- **AC-HPR-015** — **Given** an unmet goal, **When** the host stops despite a block (`stop_hook_active:true`, or the consecutive-block cap), **Then** the goal state is not `satisfied`.

### Receipts and timeouts

- **AC-HPR-016** — **Given** the rendered Codex `hooks.json` and the declared per-member internal budgets (design.md §D3.3), **When** the test evaluates `Σ(internal budgets of in-hook members) + chain_overhead` for the single Codex Stop handler, **Then** the sum is ≤ the rendered Stop timeout `T_stop`, receipt members count only their compare budget, and `T_stop` is ≤ the recorded `T_codex_max` whenever AC-HPR-021 has produced one. **Timing leg:** **When** each in-hook member runs on the golden fixtures and its wall-clock cost is measured over repeated runs, **Then** the observed maximum per member is ≤ its declared internal budget. The observed figures are recorded in progress.md §E.2 as the M2d measurement, and the declared budgets are not treated as validated until this leg passes.
- **AC-HPR-017** — **Given** a receipt written after a passing out-of-hook check, **When** the Stop chain evaluates it unchanged, **Then** it is accepted; **When** HEAD, working-tree digest, configuration digest, command, or tool version differs (each mutated separately), or the receipt is absent or truncated, **Then** the check is treated as not run.
- **AC-HPR-021** — **Given** a scratch project under a temporary `CODEX_HOME`, **When** the probe renders a Stop handler at an ascending ladder of timeouts, each running a sleeper just under its timeout plus one sleeper past it, **Then** it records, for each rung, whether Codex accepted the configuration, waited, and honoured the handler's decision. The largest honoured rung is recorded as `T_codex_max` with its evidence. Until the probe has run, `T_codex_max` is `NOT_RUN` and `T_stop` stays at the current render constant.

- **AC-HPR-022** — **Given** a handler result normalized to `needs_input` on each decision-bearing event (PreToolUse, PermissionRequest), **When** it passes through the Codex translation and the adapter's discard sink (`RecordDiscards`, `internal/codexadapter/diagnostics.go:24`) with a temp project root, **Then** the Codex output is a deny whose reason names the required user input, and exactly one discard record for that conversion is written to the sink. No live host is involved; this is the non-live leg of REQ-HPR-008 (plan-audit iter-2 N6).

### AC-POL-01 — obligation coverage

- **AC-HPR-018** — **Given** the obligation registry, **When** the coverage check runs, **Then** it passes only if every required obligation has a Claude path, a Codex path (or an explicit `UNSUPPORTED`/`blocked`/`unverified` marker), and a check id that resolves to an existing test; **When** one obligation's Codex path or check is removed, **Then** it fails naming that obligation.
- **AC-HPR-019** — **Given** per-obligation verdicts, **When** the aggregate is computed from both the go-test action and the verdict record, **Then** it is PASS only if every required obligation is effect-verified with full attribution; a skip, an empty run, a `NOT_RUN`, an `UNSUPPORTED`, a `blocked`, an `unverified`, a missing attribution field, config-existence-only evidence, or a `pass` action paired with a `NOT_RUN` record makes it not PASS.

### Isolation

- **AC-HPR-020** — **Given** the live suite, **When** it runs, **Then** every Codex process sees a temporary `CODEX_HOME` and the content hash of the operator's `~/.codex` config and hooks is identical before and after; and every Claude run's working directory is under the OS temp dir and outside the repository root.

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

- AC-HPR-001..022 each carry a recorded verdict with attribution in progress.md §E.2.
- Every unit/golden AC and every unit/golden leg is PASS under rule P.
- Operator decision Q5 (2026-09-23): no live leg is run in this SPEC. The live tests are built and
  opt-in, and each live leg (AC-HPR-004, 007, the live legs of 008–012, 020, 021) is recorded
  `NOT_RUN` by running it without `MOAI_PARITY_LIVE`, with its `t.Skipf` output quoted.
- The SPEC closes as **partial (live-uncertified)**, with the aggregate verdict reported as not-PASS
  for every obligation that needs a live leg. Live certification belongs to a separate follow-up
  card.
- No claim of "full dual-harness parity" is made from this SPEC alone; sibling card t1100 and M1/M5 remain.

## §H RED-now Evidence Ledger (plan phase)

Document-level pin: every entry below was measured on tree `530d8cc067765a3cf6ac49a76954d34a2193c7d0`
(worktree `dual-harness-parity-rebuild`) on 2026-09-23. Each entry records the command, its
verbatim stdout, and its exit code as separate fields. The green-path column names the milestone
expected to flip the observation.

| Ledger id | Serves AC | Command | Verbatim stdout | Exit | Why this is RED | Green path |
|---|---|---|---|---|---|---|
| L-01 | AC-HPR-002, 003, 012 | `grep -n stop-goal internal/codexwiring/hooks.go` | (empty) | 1 | the goal member is not on the Codex Stop path at all | M2d/M2f: the Codex Stop chain invokes the goal evaluator; the AC test asserts it |
| L-02 | AC-HPR-006, 022 | `grep -n '"ask": *true' internal/codexadapter/output.go` | `40:	"ask":   true,` | 0 | `ask` is in the drop set, so it degrades to `{}` | M2c: `ask` leaves the drop set; the property test fails if it returns |
| L-03 | AC-HPR-009, 010, 011 | `grep -n "EventPreCompact\|EventPostCompact\|EventPermissionRequest\|CodexEventInterrupt, " internal/codexadapter/events.go` | `79: {hook.EventPreCompact, "compact", false},` · `80: {hook.EventPostCompact, "post-compact", false},` · `81: {hook.EventPermissionRequest, "permission-request", false},` · `83: {CodexEventInterrupt, "", false},` | 0 | all four rows unadapted (`false`) | M2e: rows adapted; round-trip, deny, and cancellation tests assert the effect |
| L-04 | AC-HPR-013 | `grep -n cancelled internal/goal/schema.go` | (empty) | 1 | no cancellation status exists | M2a (Q3): status added; precedence test asserts it |
| L-05 | AC-HPR-016 | `grep -n "defaultHandlerTimeout = " internal/codexwiring/codexwiring.go` ; `grep -n "stopGoalHookTimeout = " internal/cli/hook_stop_goal.go` | `60:	defaultHandlerTimeout = 10` · `20:const stopGoalHookTimeout = 90 * time.Second` | 0 · 0 | a 90 s member budget sits under a 10 s handler timeout | M2b/M2d: the budget test fails on any member above its timeout |
| L-06 | AC-HPR-017 | `grep -rln ToolVersion internal/verify` | (empty) | 1 | the snapshot record carries no tool-version field | M2b: receipt carries all 5 fields; per-field mutation test |
| L-07 | AC-HPR-018 | `ls internal/template/obligations.yaml` | `ls: internal/template/obligations.yaml: No such file or directory` | 1 | no registry exists | M2a/M2h: registry + coverage check |
| L-08 | rule P (all Go ACs) | `go test -count=1 -run '^TestStopChainInventoryMatchesClaudeTemplate$' ./internal/codexwiring/` | `ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.479s [no tests to run]` | 0 | exit 0 on an empty sweep: the named test does not exist yet | the test lands in M2d; until then this reads `NOT_RUN`, never PASS (§B.2) |

L-03 stdout is shown with the leading tab collapsed to one space for table rendering; the raw
output carries a tab after the colon.

The live ACs (AC-HPR-004, 007, the live legs of 008–012, 020, 021) have no RED-now cell that can be
re-executed at plan time without a live host run. Under operator decision Q5 they stay **NOT_RUN**
for this SPEC and are classified **regression-guard** (verification-completeness §2.1, undecidable
disposition) until the follow-up live-certification card records a first observation.
