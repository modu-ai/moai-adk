# SPEC-FACTORY-STALE-RUN-HEAL-001 — Implementation Plan

## A. Context

Card t1345 (Tier M, Class C), Factory Mode only. A lane session's run identity lives in its launch environment, which a hook cannot change and `/clear` does not re-read. When the run it names dies or is replaced, the lane's hook keeps claiming from the dead run's broker while the new leader writes to the new run's broker. SPEC-FACTORY-LANE-JOIN-SOCKET-001 (t1330, completed) repaired the launcher entry; SPEC-STALE-RUN-LABEL-001 (t1373, completed) stopped the repeated prescription and carved this card out by name. This plan closes the remaining three behaviours: automatic lane re-registration, an executable notice, and the `moai factory relaunch` verb.

Root cause and the code map are in `spec.md` §C. Key files (line numbers at tree `802a72235`):

| File | Role |
|---|---|
| `internal/hook/factory_messages.go:54-130` | `registerFactoryHookPeer` — peer registration; the legacy branch (68-70) is gated, the current-vocabulary path reaches `ValidateActiveRun` (90-92) and degrades every turn |
| `internal/hook/factory_messages.go:132-190` | `factoryHookBatch` — inbox claim, keyed on the environment run id (135); the additive read of REQ-SRH-008 is a sibling path, the existing sequence is untouched |
| `internal/hook/stale_run_gate.go` | `staleRunPrescriptionGate`, the notice marker (`factoryNoticeMarker`), `unbindFactoryHookNotice`; the marker gains a rebound kind, the gate gains the current-vocabulary caller |
| `internal/hook/session_stale_run.go:74-127` | `staleRunMessages` — `roleValueRetire`, `laneLabelRetire`, `laneLabelUnbindRebind` in four locales; `roleValueRelaunch` is the kanban prose and stays byte-identical |
| `internal/hook/session_start_factory.go:51-68`, `session_start.go:466,511` | SessionStart wiring: the bootstrap notice and the session-start peer registration (REQ-SRH-005) |
| `internal/hook/user_prompt_submit.go:152-166` | UserPromptSubmit wiring: peer registration then `factoryHookBatch` |
| `internal/factorymsg/run_state.go` | `ProbeRunStateAt`, `ActiveRunExistsAt` — the shared run-state accessors; gains a read-only active-run listing |
| `internal/cli/factory_lane_relaunch.go:40-81` | the `--clear-policy relaunch` loop; reads the run id once (56) — REQ-SRH-011 |
| `internal/cli/factory.go:421-541` | `enterSelectedFactoryRun`, `enterFactoryLaneRun` (505) — the shared join gate the verb and the loop reuse |
| `internal/cli/factory_handoff_recover.go:19-62` | the `moai factory` command tree; the verb registers here next to `runs` |
| `internal/cli/cc.go:215-285`, `glm.go` | the launcher lane branch that calls the join gate and diverts to the loop |
| `internal/homestate/factory_run_retire.go:362` | `RetireRunIfDead` — the existing proof-gated predicate REQ-SRH-013 reuses unchanged |
| `internal/kanban/bootstrap.go:267-375` | lane-label helpers (`FactoryLaneLabel`, `SplitFactoryLaneLabel`, `IsLegacyFactoryRoleValue`); the natural home of the shared command builder both the hook and the CLI import |

## B. Known Issues

- **Existing tests assert the old notice text.** `stale_run_gate_test.go` (lines 85, 100 — `runs --retire <run>`; 160-165 — `moai cc -f lane-`), `stale_run_m1_test.go:110`, and `role_naming_m3_notice_test.go:130` assert literals this SPEC replaces. They are updated at M1, not deleted, and no cadence assertion is weakened (AC-SRH-007). The `internal/cli` tests that name `runs --retire` (`factory_role_refusal_m2_test.go:377`, `factory_run_boundary_m1_test.go:65`, `codex_factory_retire_test.go:546`) test launcher refusals and are untouched.
- **Package import direction.** `internal/hook` cannot import `internal/cli`, and `factorymsg` imports `homestate`. The shared command builder therefore lives in a package both `hook` and `cli` already import (`internal/kanban`), and the round-trip test (AC-SRH-006) lives in `internal/cli`.
- **The claim path must not change.** SPEC-STALE-RUN-LABEL-001 REQ-SRL-007 forbids the environment-label judgment from altering delivery. The rebound read is an added sibling path, not a modification of `factoryHookBatch`'s existing sequence; `TestInboundClaimIndependentOfEnvLabel` is the guard.
- **SessionStart cannot register into the new run.** SessionStart binding uses launch-pending rows (`BindLaunchPending`), which exist in the run the launcher recorded, not in the new run. Registration is therefore a UserPromptSubmit act (decision D6).
- **Hook budget.** `factoryHookInspectionDeadline` is 200 ms. The healthy path adds no measurement; the rebind path adds one run-state read (already made by `ValidateActiveRun`) and one active-run listing, on a single resolved database path (the gate's own lesson: five path resolutions exhausted the budget).
- **Codex lane provider token.** The hook's `factoryLaunchEntry` maps only `glm` to `glm` and everything else to `cc`; the relaunch builder needs a lane-own mapping (`claude`→`cc`, `glm`→`glm`, `gpt`→`codex`) read from the session's backend. The leader notice's existing mapping is not touched.

## C. Pre-flight

1. Re-read `runs.status` values and confirm `retired` is the only non-active value written by the retirement path (`factory_run_retire.go:retireRun`).
2. Confirm `RegisterPeer` into a second run's broker accepts the slot (`peers.slot` is unique per broker file) and that a same-pid session-uuid change bumps the generation instead of refusing (`store.go:384-476`).
3. Inventory every call site that renders the stale-run or unbind text: `grep -rn "runs --retire" internal --include='*.go'` (quote the glob) — the hook sites are `session_stale_run.go:85-122`; the launcher sites are out of scope.
4. Confirm the launcher entry can be re-executed by the verb for each provider token (`moai cc|glm|codex -f ...`) with no provider-specific code (A5); a codex entry that cannot is a blocker report, not a scope addition.
5. Measure the relaunch loop's run handling at RED: iteration 2 after a run switch leases from the retired run (AC-SRH-014's own RED).

## D. Constraints

- Error wrapping `fmt.Errorf("operation: %w", err)`; English comments; environment names through `internal/config/envkeys.go` constants only.
- Template neutrality: Go code under `internal/hook`, `internal/factorymsg`, `internal/cli`, `internal/kanban` only. No file under `internal/template/templates/` changes; no card or SPEC-id strings enter the template tree.
- The hook fails open and writes only the peer registration and the notice marker (REQ-SRH-010). It never retires, reactivates, or creates a run.
- The verb never creates, edits, or reactivates a run record beyond REQ-SRH-013's retirement and what the join gate performs (REQ-SRH-014).
- Kanban-only surfaces are not touched (REQ-SRH-015).
- Tests use `t.TempDir()` factory databases and seams for the launcher, discovery, owner classification, and lease; no live factory run. All Go tests run in the env-scrubbed compound invocation of `acceptance.md`.
- Verification is scoped to the changed packages (`internal/hook`, `internal/factorymsg`, `internal/cli`, `internal/kanban`, plus `internal/homestate` for the retirement AC); the full suite is left to CI.

## E. Self-Verification

- E1: AC matrix PASS/FAIL with verbatim command output and the swept count per Go-test AC (`acceptance.md`).
- E2: `go vet` and `golangci-lint run` (v2.1.6) on the four changed packages, verbatim.
- E3: `GOOS=windows GOARCH=amd64 go build ./...`, verbatim.
- E4: separation diff review — no environment-label reference added inside the existing claim sequence.
- E5: every mutant probe of `acceptance.md` D.2 executed, the observed failure recorded.
- E6: `moai spec lint SPEC-FACTORY-STALE-RUN-HEAL-001 --strict` on a binary built from the tree, invoked by path, with the judging build's commit beside the tree HEAD.

## F. Milestones

Priority order. M1 and M2 carry the decisions most likely to change in review — the command grammar and verb contract (the new user-facing interface) and the rebind eligibility and claim shape (new run-state interface); the later milestones are mechanical wiring and re-measurement.

### M1 (High) — Command grammar, the verb, and the notice text

The user-facing decisions live here: the grammar of the line the operator pastes, and what the verb does with it.

1. **Shared command builder** (REQ-SRH-002) in `internal/kanban`: one definition of the line (`moai factory relaunch --provider <tok> [--lane <label>] [--run <id>] [--from-run <id>]`) used to print and to be parsed against. The provider token is derived from the session's backend (`claude`→`cc`, `glm`→`glm`, `gpt`→`codex`); a legacy label yields no `--lane`. `TestRelaunchCommandRoundTrip` (AC-SRH-006) and the notice scan (AC-SRH-005) both consume it.
2. **The verb** (REQ-SRH-012..014) in `internal/cli`, registered in `newFactoryCommand` next to `runs`: argument validation (legacy spelling refused naming the canonical form), the optional retirement pre-step through `RetireRunIfDead` with the existing reconcile options, then a re-execution of `moai <provider> -f lane[-<n>] [--factory-run <id>]` through a launch seam; `--dry-run` prints the line and writes nothing. Help text states the distinction from `--clear-policy relaunch`. Tests: `TestFactoryRelaunchDryRunMatrix`, `TestFactoryRelaunchRefusesLegacyLane`, `TestFactoryRelaunchFromRunRetiresDeadOwnerOnly`, `TestFactoryRelaunchDoesNotMutateRunRecords`, `TestFactoryRelaunchHelpDistinguishesClearPolicy`.
3. **Notice replacement** (REQ-SRH-001, REQ-SRH-003) in `session_stale_run.go`: `roleValueRetire`, `laneLabelRetire`, `laneLabelUnbindRebind` in en/ko/ja/zh take the command line as an interpolated protocol token; the unbind notice's command line stays conditional on an active run existing. The kanban prose `roleValueRelaunch` is not touched. Test: `TestStaleNoticeCarriesExecutableRelaunch`.
4. **Literal-text test updates** (AC-SRH-007): the four assertion sites of §B change to the new line; every cadence assertion is retained verbatim.
5. RED first for 1-3: E1-E3 are already pinned; the Go-test REDs are recorded here per `verification-completeness.md` §2.1.

### M2 (High) — Hook rebind: eligibility, registration, claim

1. **Read-only active-run listing** in `internal/factorymsg/run_state.go` (a sibling of `ActiveRunExistsAt`): returns the active run ids for one resolved database path; measurement failure is an error the caller fails open on. The hook never calls the reconciling `ResolveActiveRun`.
2. **Eligibility predicate** (REQ-SRH-004/006/007): current-vocabulary lane label, environment run measures not active, then the listing: one run → eligible; zero or several → ineligible; a legacy label is never eligible. Evaluated only when `ValidateActiveRun` has already failed, so a healthy lane pays nothing (REQ-SRH-009).
3. **Registration** at the UserPromptSubmit path only (REQ-SRH-004, D6): `RegisterPeer` into the sole active run's broker under the lane's slot; a live-owner refusal surfaces once with the command (REQ-SRH-007); the `rebound` marker kind joins `prescription` and `unbind` in `factoryNoticeMarker` (once per session identity).
4. **SessionStart** (REQ-SRH-005): an eligible session binds nothing and emits no unbind notice.
5. **Ineligible current-vocabulary sessions** (REQ-SRH-006): the gate gains the current-vocabulary caller, replacing the per-prompt `factory messaging degraded: NO_ACTIVE_FACTORY` with the once-only unbind or ambiguity notice (ambiguity: one command line per candidate, three at most plus a count — D10).
6. **Additive claim** (REQ-SRH-008): a rebound session's `factoryHookBatch` also reads the rebound run's broker, in a separate path after the existing sequence, inside the same deadline.
7. RED first: `TestLaneRebindsIntoSoleActiveRun`, `TestRebindEligibleSessionStartBindsNothing`, `TestRebindIneligibleZeroOrManyActiveRuns`, `TestLegacyLabelNeverRebindsAndLiveOwnerNotDisplaced`, `TestReboundClaimReadsBothBrokers`, `TestHealthyLanePathUnchangedAndFailOpen`; mutant probes of `acceptance.md` D.2 for AC-SRH-008..013.

### M3 (Medium) — The relaunch loop follows the run

1. In `runFactoryLaneRelaunch`, before each lease, re-resolve the run through the same selection the launch used (REQ-SRH-011): explicit `--factory-run` stays explicit; otherwise the shared join gate. The resolved id replaces the loop's `runID`, is re-stamped into the child's environment, and a refusal stops the loop with the gate's text and no lease.
2. RED first: `TestRelaunchLoopReResolvesRun` with the launch, discovery, and lease seams; mutant probe: resolve once.
3. The loop's own card-lease semantics, stop condition on "no card", and child launch shape are untouched.

### M4 (Medium) — Preservation and mechanical sweeps

1. `TestKanbanRelaunchProseUnchanged` (golden bytes for `roleValueRelaunch` in four locales) and the kanban-file diff command (AC-SRH-015).
2. Characterization re-runs: the E5 command set, `TestInboundClaimIndependentOfEnvLabel`, `TestCurrentVocabularyBindPathUnchanged`.
3. Lint, vet, the Windows cross-build, i18n parity check, the full AC matrix re-run, E1-E6.

## G. Anti-Patterns

- Do NOT gate or modify the existing claim sequence on the environment label — the rebound read is an added path (REQ-SRH-008, the worker-70 invariant).
- Do NOT call `ResolveActiveRun` or `ReconcileActiveRuns` from a hook: both write (retirement). The hook reads a listing.
- Do NOT classify run owners inside the hook (D3); the retirement machinery owns liveness judgement, and a process probe on a 200 ms budget is the cost the decision declines.
- Do NOT print a command containing `<` or `>`; a placeholder is the defect this SPEC removes. Truncating a command to fit a budget is the same defect.
- Do NOT map a legacy label to a lane number anywhere (REQ-RNC-009), including inside the verb and the notice.
- Do NOT register at SessionStart (D6), and do NOT emit the final unbind notice for a rebind-eligible session — it would contradict the rebind one prompt later.
- Do NOT add provider-specific launch code to the verb (A5); a provider whose entry cannot be re-executed is a blocker report.
- Do NOT persist a rebind carrier: the effective run is derived from measured state each turn; the only new persisted state is the notice marker kind.
- Do NOT touch the launcher refusal texts, the kanban prose, or any kanban-only file.

## H. Cross-References

- `spec.md` §C root cause, §D requirements, §G decision points and assumptions
- `acceptance.md` AC matrix, evidence ledger, mutant probes
- SPEC-FACTORY-LANE-JOIN-SOCKET-001 (the join gate and discovery this SPEC reuses — REQ-010 one shared gate), SPEC-STALE-RUN-LABEL-001 (cadence, claim independence, the gate), SPEC-FACTORY-RUN-RETIRE-001 (the retirement predicate, REQ-014), SPEC-ROLE-NAMING-CODE-001 (legacy vocabulary, REQ-RNC-009), SPEC-FACTORY-SELF-DISPATCH-001 (the clear policies, the relaunch loop)
- `.claude/rules/moai/development/verification-completeness.md` §1.1-§2.1 (two-cell discipline, evidence ledger, mutant probes)
