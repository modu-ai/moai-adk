# Plan — SPEC-DUAL-HARNESS-HOOK-PARITY-001

## §A Context

Card t1099, Class C, Tier L. M2 of the dual-harness design: hook chain, approval-decision
preservation, goal continuation, and obligation coverage. Sibling card t1100 owns M3/M4/M5 and MCP.
Worktree: `.claude/worktrees/dual-harness-parity-rebuild`, branch `WT-dual-harness-parity-rebuild`.

## §B Known issues at plan time (measured, research.md §R1)

- Codex Stop runs one handler that returns allow except for factory continuation (R1.3).
- 4 of 12 Codex events unadapted (R1.1).
- PreToolUse `ask` dropped to no-opinion on Codex (R1.4). Whether that resolves to allow is unmeasured.
- Codex handler timeout 10 s vs goal-condition budget 90 s (R1.5).
- No obligation registry (R1.6); installed codex-cli 0.155.1 differs from the 0.153.4 basis (R1.7).

## §C Decision record (clarifications resolved 2026-09-23)

The iter-1 plan listed six open clarifications (Q1–Q6). All six are resolved; none remains open.

| # | Question | Decision | Source |
|---|---|---|---|
| Q1 | AC-POL-01 closure scope | Whole catalog. Every obligation is registered, including M1 standing-policy obligations. M1 rows are marked `blocked:M1` and reported honestly as FAIL until M1 lands (design.md §D4). | Lead via Jev, standing operator delegation 2026-09-23 (confidence 0.62) |
| Q2 | `needs_input` on Codex | Fail-closed deny, surfaced visibly: the deny reason reaches the model and the conversion is written to the adapter's discard record. No silent allow (design.md §D5). | Lead via Jev, standing operator delegation (confidence 0.78) |
| Q3 | Goal cancellation status | Add a new `cancelled` status; do not reuse `cleared`. Every non-test reader of goal status handles it explicitly; an unrecognised status surfaces a diagnostic (design.md §D6, AC-HPR-013). | Operator, directly (2026-09-23) |
| Q4 | Sync-gate port shape (shim over a Go entry vs parallel Go implementation) | Run-phase decision inside M2d, recorded in progress.md with the AC-HPR-002 goldens as the equivalence proof (design.md §D2). Not a pre-kickoff question (plan-audit iter-1 D9). | Coordinator instruction following plan-audit iter-1 D9 |
| Q5 | Live-run budget and credentials | No live runs in this SPEC. Live tests are built, opt-in, and unexecuted; every live leg records `NOT_RUN`, which is not PASS. The SPEC closes as partial (live-uncertified); live certification goes to a separate follow-up card (spec.md §E). | Operator, directly (2026-09-23) |
| Q6 | SPEC-CODEX-HOOK-ADAPTER-001 REQ-7 (`internal/hook` invariance) | Keep. Nothing under `internal/hook` changes (design.md §D7). | Lead via Jev, standing operator delegation (confidence 0.79) |

The Jev-sourced decisions (Q1, Q2, Q6) carry the delegation's confidence as recorded. They are
decisions, not measurements. The operator can reopen any of them at Implementation Kickoff
Approval.

Q1, Q2, and Q6 are to be confirmed by the operator at Implementation Kickoff Approval (plan-audit
iter-2 N7). Q2 in particular changes user-visible Codex behaviour.

Plan-audit iter-3 (FAIL 0.90) left four blocking findings. The operator authorized one delta
re-audit (iteration 4) scoped to them and decided each:

| # | Finding | Decision | Source |
|---|---|---|---|
| R1 | Codex review gate (Stop member 6) was classed `fail-open-on-missing`, but Claude runs the review in-hook and blocks on FAIL when codex is installed | Match Claude. Member 6 becomes a `required-gate` on the receipt method: the review runs out of hook through a `moai` CLI runner (named in M2d) that the working agent runs after its last edit and whenever the Stop reason names it, writing a receipt bound to HEAD + uncommitted-diff digest + config digest + command + `codex --version` + verdict. At Stop: codex installed + reviewable change + receipt missing, stale, or FAIL → no allow (fail-closed); codex binary missing → allow + visible discard, as on Claude. `fail-open-on-missing` stays only for member 7 (design.md §D3.3, §D3.4, §D3.5) | Operator, 09-23 |
| R2 | The Codex sync gate would demand a receipt on every Stop, while Claude's gate triggers only when HEAD is a sync-phase commit | Match Claude. The Codex member evaluates the Claude script's self-gate in-hook first (`sync-phase-quality-gate.sh:207–218` subject predicate, then `:227–231` language marker, `:233–253` code delta). Not a sync-phase commit → allow; sync-phase commit → the receipt is required, and missing, stale, or FAIL does not allow (design.md §D3.3 "Self-gates before receipts") | Operator, 09-23 |
| R3 | plan.md claimed every other existing test stays green | Replace the claim with an explicit intentional-amendment list (M2e below; research.md §R1.13); record REQ-CEV-001 and REQ-CEV-003 as reversed alongside REQ-CEV-004 (spec.md §D) | Operator, 09-23 |
| R4 | The sync close was required to write a `tags` token and a HISTORY line, which manager-docs may not do | Option (b): manager-spec writes the `live-uncertified` tag and the "closure mode" HISTORY line now, in plan phase (the partial closure is already decided by Q5). The sync close keeps only manager-docs surfaces: a progress.md §E.4 statement and a CHANGELOG entry. All four markers are in acceptance.md §G. The ownership matrix is not changed | Operator, 09-23 (option chosen by manager-spec: the marker content does not depend on run results, and (a) alone would leave spec.md itself with no partial-state marker) |

A later operator decision closes a gap left after the iter-4 repair:

| # | Finding | Decision | Source |
|---|---|---|---|
| C1 | Codex continuation loop left unbounded: the sync gate's missing/stale-receipt continuation ignores `stop_hook_active`, and member 6's only other bound (step 2) is live-unmeasured on Codex | Consecutive-count cap, Codex-only. After N consecutive `unmeasured` continuations for the same gate on the same HEAD + working-tree digest, the Stop chain allows the stop and records the gate `unverified` (discard record through `RecordDiscards` + reason text); no verdict or registry reads it as PASS. N is a declared constant finalized in M2d (proposed default 3, a proposal). The counter resets on a fresh receipt or a HEAD/digest change, lives at session-scoped `.moai/state/codex-stop-cap/<session-id>.json`, and is never written by the receipt producer. Declared parity deviation: Claude has no counterpart because it runs these checks in-hook and its `stop_hook_active` bounds member 6 (design.md §D3.8) | Operator, 09-23 |

## §D Constraints

- No push, no branch switch, and no edit of any other worktree. Integration follows the local develop chain via the lead.
- Local verification is scoped to changed packages (`internal/codexadapter`, `internal/codexwiring`, `internal/cli` with `-run` filters, `internal/goal`, `internal/verify`, registry package). CI gives the full-suite verdict.
- Template-first: any change under `internal/template/templates/` is followed by `make build`; agent definitions are not touched.
- No long check inside a hook (REQ-HPR-018).

## §E Self-verification (plan phase)

- SPEC ID regex: PASS (executed; see progress.md §E.1).
- 25 REQs (Tier L ceiling 25) and 22 ACs (ceiling 25) as list items (`- **REQ-HPR-NNN**`) so the lint collector sees them.
- Each REQ maps to ≥1 AC (acceptance.md §C). Each AC names a command and a mutation.

## §F Milestones

Ordered by decision reversibility. Data-model and interface decisions come first, mechanical work
last. Priorities are labels only.

| Milestone | Priority | Content | ACs | Depends on |
|---|---|---|---|---|
| **M2a — decision & state models** | High | Normalized decision type and per-event translation table (design.md §D1). New goal status `cancelled` (Q3) and its handling in every reader of goal status listed in design.md §D6: `schema.go` set, `evaluate.go:294` early return, the writers at `evaluate.go:306/325/340/404/435`, `launcher_blockcap_infinite.go:68`, `handoff.go:85`, `goal.go:1282/1331`, `dashboard.go:141`, `goal.go:1313`, `hook_stop_goal.go` emission. The `internal/hook` sites (`session_start_compact.go:88`, `handoff_inject.go:185`, `stop_failure.go:111`) are asserted unchanged (REQ-7). `TestGoalStatusConsumersHandleCancelled` scans `internal/goal`, `internal/cli`, and `internal/hook`. Unknown-status diagnostic. Obligation registry schema and loader (design.md §D4, Q1) | AC-HPR-006 (table), AC-HPR-013 (status consumers), AC-HPR-018 (schema) | — |
| **M2b — receipt contract and Stop budget** | High | Extend the `internal/verify` snapshot record and the sync gate's outcome record with config digest, command, and tool version; comparison predicate; not-run on any mismatch (design.md §D3.6). Declare per-member internal budgets and the `chain_overhead` constant; `T_stop` stays 10 s (Q5) | AC-HPR-016, AC-HPR-017 | M2a |
| **M2c — decision hardening** | High | Replace the `ask`/`defer` drop with the fail-closed visible deny (Q2). Fault-injection tests for timeout, exit 1, corrupt output, exit 2. Unit test for the visible deny: reason names the required input, and a `RecordDiscards` entry is written. **Intentional characterization amendment:** `TestPreToolUseAskDropped` (`output_test.go:278`) is inverted, not kept green | AC-HPR-006, AC-HPR-008 (unit), AC-HPR-022 | M2a |
| **M2d — Stop chain on Codex** | High | Inventory over both template renders (REQ-HPR-001). Go chain runner with the merge rule (design.md §D2). Lookup-only goal evaluation and receipt placement per design.md §D3.3. **Decide the sync-gate port shape here (Q4)** and record it in progress.md. Goldens for Claude-vs-Codex effect, including the `unmeasured` mapping, the member-6 receipt rows (codex installed with no receipt, fresh PASS, FAIL, stale HEAD; codex missing), the member-6 `stop_hook_active: true` row (the only bound on the step-7 continuation shared by both harnesses, design.md §D3.3), the sync-gate non-sync-commit and sync-commit rows, the sync-gate fresh-`fail`-receipt `stop_hook_active: true` row (allow, matching `sync-phase-quality-gate.sh:484–486`), the member-7 `fail-open-on-missing` row, and the Codex-only consecutive-`unmeasured` cap rows for the sync gate and member 6 (design.md §D3.8: N−1 continue, Nth allow + `unverified` record, reset on a fresh receipt, reset on a HEAD change). **Finalize the cap constant N** (proposed default 3) and the counter directory, and record both in progress.md §E.2. Name the out-of-hook codex review runner (R1) here, alongside the sync-gate port shape. If either entry lands as a new `moai hook` subcommand, it amends the same subcommand-count and reverse-mapping tests listed under M2e (count and `utilitySubcmds` entry), recorded in progress.md §E.2. **Timing step:** run each in-hook member on the golden fixtures, record its observed maximum cost in progress.md §E.2, and adjust the declared budgets or move a member to receipt placement if any member exceeds its budget. Member 1's factory-continuation path is exempt from the cut-off (design.md §D3.3) | AC-HPR-001, 002, 003, 005, 016 (timing leg) | M2a, M2b |
| **M2e — event adaptation** | High | Adapt PreCompact, PostCompact, PermissionRequest; add the Interrupt path in `internal/cli` (design.md §D7, Q6); update the `events.go` census comment. **Intentional characterization amendments** (none is kept green; each is rewritten to the new truth, research.md §R1.13): `internal/codexadapter/events_test.go:67` `TestAdaptedRowCount` (`wantAdapted = 8` → `12`); `events_test.go:24–42` `TestEventTableMapping` (the four rows at `:37–39` and `:42` flip to adapted, and Interrupt gets the `interrupt` dispatcher arg — REQ-CEV-001); `events_test.go:111` `TestResolveRecognizedButUnadapted` (the four names at `:114` now resolve — REQ-CEV-003); `events_test.go:134` `TestResolveInterruptNoCounterpart` (Interrupt now resolves to `interrupt` — REQ-CEV-003); `internal/codexwiring/hooks_test.go:73` `TestRenderHooks_InterruptNeverInstalled` (inverted to assert the Interrupt handler is rendered — REQ-CEV-004 / AC-CEV-005); `internal/cli/hook_harness_codex_test.go:305` `TestHarnessCodexUnadaptedSubcommandRejected` (uses `compact` as its unadapted example; retargeted to a still-unadapted event or rewritten to assert `compact` is accepted); `internal/cli/hook_test.go:61–75` `TestHookCmd_SubcommandCount` and `internal/cli/hook_pre_push_test.go:370–387` `TestHookCmd_PrePushSubcommandCount` (both pin `len(hookCmd.Commands()) == 43`; the new `moai hook interrupt` makes 44 — each gets the new count plus a comment line naming this SPEC); `internal/cli/hook_e2e_test.go:289` `TestHookValidEventTypes_AllHaveSubcommands` (its reverse check, `:352–387`, requires every subcommand to map to an `EventType` or appear in `utilitySubcmds` `:359`; `internal/hook` has no Interrupt event type — AC-HPR-011 keeps it that way — so `interrupt` joins `utilitySubcmds` with a comment naming this SPEC). Conditional: `internal/codexadapter/stderr_test.go:78` `TestExcludedEventsHaveNoClass` pins no stderr classification for PreCompact, PostCompact, and PermissionRequest; it is amended only if M2e gives those events a stderr class, and the choice is recorded in progress.md §E.2 | AC-HPR-009, 010, 011 (golden/unit legs) | M2a |
| **M2f — goal parity** | High | Codex Stop → existing evaluator in lookup-only mode; cancellation precedence through the real producers; budget termination; host override ≠ satisfied | AC-HPR-012 (golden), 013, 014, 015 | M2a, M2d, M2e |
| **M2g — live legs, built but not run** | Medium | `MOAI_PARITY_LIVE` axis declared in `codex_live_axis_declaration_test.go`; temp `CODEX_HOME`; scratch-project `claude -p`; `t.Skipf` discipline (acceptance.md §B rule 8); the Stop-timeout ceiling probe. **Not executed** (Q5): each live leg is run once without the switch and its `NOT_RUN` skip output is recorded | AC-HPR-004, 007, live legs of 008–012, 020, 021 — all `NOT_RUN` | M2c–M2f |
| **M2h — coverage & verdict aggregation** | Medium | Populate the registry with the whole catalog (Q1): M2 obligations, M1 rows `blocked:M1`, the Claude-interrupt row `UNSUPPORTED`, and the non-Stop chains `unverified`. Coverage check; aggregate verdict reading both the go-test action and the verdict record | AC-HPR-018, 019 | M2a–M2g |
| **M2i — mechanical tail** | Low | `RenderHooks` table update, `make build`, doc comments, codemaps refresh. The tests amended in M2c and M2e are the intentional-amendment list (research.md §R1.13), complete as of the M2e subcommand set (one new `moai hook` subcommand, `interrupt`) and the greps recorded there; a further `moai hook` entry decided in M2d adds its own count and `utilitySubcmds` amendments (M2d). Any other existing test that turns red is a regression, not an amendment | regression of the unamended tests | M2d, M2e |

## §G Risks

| Risk | Mitigation |
|---|---|
| Codex cannot trigger compaction or permission requests non-interactively (as on 0.153.4) | REQ-HPR-013: record `NOT_RUN` with the trigger attempted; aggregate stays FAIL; no "does not fire" claim |
| Host resolves hook faults as allow (H2) | Record `UNSUPPORTED`; mitigate at pre-tool with fail-closed default where MoAI controls output; surface in verdict |
| Sync-gate port changes Claude behaviour | AC-HPR-002 goldens run both paths on the same inputs; Claude regression tests stay |
| Receipt digest misses an input the check reads | config_digest scope is declared per check; mutation per field in AC-HPR-017 |
| Live behaviour differs from the goldens | Q5: no live run here, so this risk is carried open; the SPEC closes as partial (live-uncertified) and the follow-up card runs the live legs |
| Scope creep into t1100 (permissions, kanban, rollback) | spec.md §F exclusions; plan audit checks the diff against them |

## §H Anti-patterns to avoid

- Counting `ok` / exit 0 as PASS when the named test skipped.
- Declaring an event "does not fire" because a trigger was not achieved.
- Running a long check inside a Codex hook and raising the timeout to fit it.
- Claiming full parity from this SPEC while t1100 and M1/M5 remain open.

## §I Cross-references

- `reports/moai-dual-harness-full-design-20260922.md` §07–§09, §17–§19
- `reports/moai-dual-harness-implementation-status-20260923.md`
- `reports/moai-dual-harness-handoff-20260923.md`
- SPEC-CODEX-HOOK-ADAPTER-001, SPEC-CODEX-EVENT-COVERAGE-001, SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-WIRING-001
