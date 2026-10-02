# SPEC-FACTORY-STALE-RUN-HEAL-001 — Implementation Plan

## A. Context

Card t1345 (Tier M, Class C), Factory Mode only. A lane session's run identity lives in its launch environment, which a hook cannot change and `/clear` does not re-read. When the run it names dies or is replaced, the lane's hook keeps claiming from the dead run's broker while the new leader writes to the new run's broker. SPEC-FACTORY-LANE-JOIN-SOCKET-001 (t1330, completed) repaired the launcher entry; SPEC-STALE-RUN-LABEL-001 (t1373, completed) stopped the repeated prescription for legacy labels and carved this card out by name. This plan closes the remaining three behaviours: automatic lane re-registration (current-vocabulary sessions, in place), an executable notice, and the `moai factory relaunch` verb. Legacy-vocabulary sessions are never re-bound (REQ-RNC-009); they receive the command for a new session.

Root cause and the code map are in `spec.md` §C. Key files (line numbers at tree `cda6913d1`, Go identical to `802a72235`):

| File | Role |
|---|---|
| `internal/hook/factory_messages.go:54-130` | `registerFactoryHookPeer` — peer registration; the legacy branch (68-70) is gated, the current-vocabulary path reaches `ValidateActiveRun` (90-92) and degrades every prompt |
| `internal/hook/factory_messages.go:132-190` | `factoryHookBatch` — inbox claim, keyed on the environment run id (135); gains a run parameter, the existing sequence is untouched |
| `internal/hook/stale_run_gate.go` | `staleRunPrescriptionGate`, `factoryNoticeMarker`, `unbindFactoryHookNotice`; the marker gains the `state` field, the gate's unbind path gains the command lines |
| `internal/hook/session_stale_run.go:74-127` | `staleRunMessages` — `roleValueRetire`, `laneLabelRetire`, `laneLabelUnbindRebind` in four locales (N5/N6); `roleValueRelaunch` is kanban prose and stays byte-identical |
| `internal/hook/session_start.go:466,511,523`, `session_start_factory.go:51-68` | SessionStart wiring: peer registration, agent copy and operator copy of the bootstrap notice |
| `internal/hook/user_prompt_submit.go:62,152-166` | `factoryBindBudget` (2 s, already a var), registration then claim; the handoff of REQ-SRH-004 lives here |
| `internal/hook/stop.go:88` | Stop reads `factoryHookBatch` and rebinds nothing (REQ-SRH-008) |
| `internal/factorymsg/run_state.go` | `ProbeRunStateAt`, `ActiveRunExistsAt`; gains a read-only active-run listing (`...IDsAt`) |
| `internal/cli/factory_lane_relaunch.go:40-81` | the cc/glm `--clear-policy relaunch` loop; reads the run id once (56); signature carries neither the explicit selector nor the leader target |
| `internal/cli/cc.go:215-285`, `glm.go:318` | the lane branch calling the join gate and diverting to the loop — the call sites that gain the two arguments |
| `internal/cli/factory.go:421-541` | `enterSelectedFactoryRun`, `enterFactoryLaneRun` (505) — the shared join gate |
| `internal/cli/codex_launcher.go` (`codexFactoryEntryClassify`, `runCodexFactoryLane`) | the Codex entry: only `-f lane`, run via `enterSelectedFactoryRun`, loop reads the run once (excluded, `spec.md` §F) |
| `internal/cli/factory_handoff_recover.go:19-62` | the `moai factory` command tree; the verb registers next to `runs` |
| `internal/homestate/factory_run_retire.go:362` | `RetireRunIfDead` — the existing predicate REQ-SRH-013 reuses unchanged |
| `internal/kanban/bootstrap.go:267-375`, `factory_slots.go` | lane-label helpers and the `workers` registry; **unchanged** (DP11) |

## B. Known Issues and Design

- **Existing tests assert the superseded text.** Five literal-assertion sites: `stale_run_gate_test.go` (lines 85, 100 — `runs --retire <run>`; 154, 162 — `moai cc -f lane-`), `stale_run_m1_test.go:110`, `role_naming_m3_notice_test.go:130`. They are updated at M1 and all seven tests of AC-SRH-007 stay green; no cadence assertion is weakened. The `internal/cli` tests naming `runs --retire` (`factory_role_refusal_m2_test.go:377`, `factory_run_boundary_m1_test.go:65`, `codex_factory_retire_test.go:546`) test launcher refusals and are untouched.
- **One measurement on the healthy path (REQ-SRH-009).** `ValidateActiveRun` (`factorymsg/store.go:152`) resolves its own `FactoryDBPath` (git subprocesses) and collapses every scan error into `NO_ACTIVE_FACTORY`, so it cannot supply the tri-state. The current-vocabulary path is changed to: resolve the database path once (through a package-variable resolver, the seam of the counting test), call `ProbeRunStateAt` once. That is the same count as today — one resolution, one query — and the gate's own lesson stands: never resolve the path twice (`stale_run_gate.go:169-172`). Mapping: active → the existing path unchanged; unavailable → `factory messaging degraded: <err>`; not active → the new branch. Only the not-active branch adds work: one read-only active-run listing on the same resolved path.
- **Hand-off within one invocation, and Stop (REQ-SRH-004, -008).** `registerFactoryUserPromptPeer` returns `(notice, reboundRun)`; the rebound run is non-empty only when the registration succeeded or the peer already matched in the sole active run. `user_prompt_submit.go` passes it to a run-parameterised claim (`factoryHookBatchForRun`); `factoryHookBatch` stays as the wrapper with an empty parameter (the environment run) and is what Stop calls. No persisted run carrier exists, so a Stop event cannot know the rebound run — by decision it does not rebind and reads the environment run's broker as before (`spec.md` §H records the consequence).
- **Budgets (REQ-SRH-010).** The registration steps — the state probe, the listing, the broker open, the peer lookup and registration, the marker read/write — run under the bind budget (`factoryBindBudget`, 2 s, wrapping the caller's context). The legacy gate keeps its own 200 ms gate budget. The claim runs under the inspection deadline (200 ms) and opens one broker, as today, so its cost is unchanged. The broker open takes no context and carries its own 5 s limit, as the bind path already does (`user_prompt_submit.go:35-60`).
- **Injectable seams (D15).** `factoryBindBudget` is already a variable. Plan: convert `factoryHookInspectionDeadline` and `factoryGateBudget` from constants to variables (production never assigns them), and add variables for the database-path resolver and the active-run listing. Tests set the budgets generously (so no AC depends on load) or to one nanosecond (the exhausted-budget path), and count resolutions and listings.
- **Marker `state` field (DP12).** `factoryNoticeMarker` gains `state string` (json `state,omitempty`) holding the last emitted state key — `rebound:Y`, `unbound:X`, `ambiguous:<ids>`, `refused:Y`. A current-vocabulary notice is emitted only when the computed key differs from the stored one. The legacy `prescription` and `unbind` fields are untouched; the legacy unbind stays final.
- **No `workers` writes (DP11).** `ClaimFactoryLaneWithin` cannot be reused: a bounded explicit-slot claim refuses a slot the session's own live pid already holds (`factory lane "lane-3" is already occupied in run Y`). The rebind writes the broker `peers` row and the marker only. AC-SRH-008 asserts the `workers` rows and `FactoryFreeSlots` are identical before and after.
- **Shared command builder and the t1399 overlap.** The builder lives in a new factory-named file `internal/kanban/factory_relaunch_cmd.go` (the package already holds the factory lane-label helpers `bootstrap.go` and the `workers` registry `factory_slots.go`). Card t1399 removes Kanban Mode; because the file is new and additive, the only overlap is a merge conflict if t1399 relocates the factory helpers — in which case the builder moves with them. `internal/hook` and `internal/cli` both already import `internal/kanban`, and the round-trip test lives in `internal/cli` (`hook` cannot import `cli`).
- **Codex launch line.** `codexFactoryEntryClassify` accepts only the bare `-f lane`; `-f lane-3` and `--factory-run` exit 1 (E2/E3). The verb's codex launch is `moai codex -f lane`; the builder never emits `--lane`/`--run` for `codex`; the verb refuses them with `--provider codex` (REQ-SRH-016). The Codex entry resolves its run through `enterSelectedFactoryRun` (single active run, no discovery), so the verb's "same join gate" statement is true of cc/glm only.
- **Loop plumbing (REQ-SRH-011).** `runFactoryLaneRelaunch(cmd, label, claudeArgs)` becomes `runFactoryLaneRelaunch(cmd, label, claudeArgs, explicit, leadTarget)`; `cc.go:280` and `glm.go:318` pass `entry.FactoryRun` and `entry.FactoryLead`. Each iteration calls `enterFactoryLaneRun(root, explicit, leadTarget, nil)` (its returned restore is deferred per iteration), re-stamps `MOAI_KANBAN_ID`, and passes the resolved id to `factoryNextLeaseOnce`. The Codex loop is out of scope (`spec.md` §F).
- **Probe in step with the spec.** `probe/hook-probe.sh` encodes the protocol prefixes and command lines of `spec.md` §D.7/§D.8. A change to either changes the probe in the same commit.

## C. Pre-flight

1. Re-read `runs.status` values and confirm `retired` is the only non-active value the retirement path writes (`factory_run_retire.go:retireRun`).
2. Confirm `RegisterPeer` into a second run's broker accepts the slot and that a same-pid session-uuid change bumps the generation instead of refusing (`store.go:384-476`).
3. Inventory the render sites of the stale-run and unbind text: `grep -rn "runs --retire" internal --include='*.go'` (quote the glob). Confirm whether `roleValueRetire` is reachable outside its own test; if only the test reaches it, update it with `laneLabelRetire` for consistency (row R6) rather than leave a divergent string.
4. Confirm `moai cc|glm -f ...` can be re-executed by the verb through a launch seam with no provider-specific code, and that `moai codex -f lane` is the whole Codex line (A5, E2, E3). A provider whose entry cannot be re-executed is a blocker report, not a scope addition.
5. Measure the relaunch loop at RED: iteration 2 after a run switch leases from the retired run (AC-SRH-015's own RED); measure `UserPromptSubmit` firing on a self-dispatch lane if a lane transcript is available (A4) — otherwise leave the §H consequence stated.
6. Run `probe/hook-probe.sh control-healthy` on the pre-change tree (PASS) before touching the hook.

## D. Constraints

- Error wrapping `fmt.Errorf("operation: %w", err)`; English comments; environment names through `internal/config/envkeys.go` constants only.
- Template neutrality: Go code under `internal/hook`, `internal/factorymsg`, `internal/cli`, `internal/kanban` only. No file under `internal/template/templates/`; no card or SPEC-id strings enter the template tree.
- The hook fails open and writes only the peer registration and the notice marker (REQ-SRH-010); it never retires, reactivates, or creates a run, and never touches the `workers` registry.
- The verb never creates, edits, or reactivates a run record beyond REQ-SRH-013's retirement and what the join gate performs (REQ-SRH-014).
- Kanban-only surfaces are untouched (REQ-SRH-015).
- Tests use `t.TempDir()` factory databases and seams (launch, join gate, discovery, owner classification, lease, budgets, path resolver, active-run listing); no live factory run. Go tests run in the env-scrubbed single-package form of `acceptance.md`.
- Verification is scoped to the changed packages; the full suite is left to CI.

## E. Self-Verification

- E1: AC matrix PASS/FAIL with verbatim output and the swept count per Go-test AC (`acceptance.md`); every probe scenario PASSES and `control-healthy` still PASSES.
- E2: `go vet` and `golangci-lint run` (v2.1.6) on the four changed packages, verbatim.
- E3: `GOOS=windows GOARCH=amd64 go build ./...`, verbatim.
- E4: separation review — the claim sequence for non-rebound sessions and Stop is untouched.
- E5: every mutant probe of `acceptance.md` D.2 executed, the observed failure recorded.
- E6: `moai spec lint SPEC-FACTORY-STALE-RUN-HEAL-001 --strict` on a binary built from the tree, invoked by path, with the judging build's commit beside the tree HEAD.

## F. Milestones

Priority order. M1 and M2 carry the decisions most likely to change in review — the command grammar and verb contract (new user-facing interface) and the rebind eligibility, hand-off, and state machine (new run-state interface); the later milestones are plumbing and re-measurement.

### M1 (High) — Command grammar, the verb, and the legacy notice text

1. **Shared command builder** (REQ-SRH-002, REQ-SRH-016) in `internal/kanban/factory_relaunch_cmd.go`: one definition of the line `moai factory relaunch --provider P [--lane S] [--run R] [--from-run F]` and of the flag names the verb parses; the provider token mapping (`glm`→glm, `gpt`→codex, else cc); the codex rule (no `--lane`/`--run`); the table of `spec.md` §D.7 as a pure function of (vocabulary, run state, active runs). Tests: `TestRelaunchCommandRoundTrip` (in `internal/cli`), the table tests.
2. **The verb** (REQ-SRH-012..014) in `internal/cli/factory_relaunch.go`, registered in `newFactoryCommand` next to `runs`: argument validation (legacy spelling and codex pins refused), the optional retirement through `RetireRunIfDead` with `ReconcileOptionsFor`, then the launch seam re-executing `moai <provider> -f lane[-<n>] [--factory-run <id>]`; `--dry-run` prints the line and writes nothing; help text states the distinction from `--clear-policy relaunch`. Tests: `TestFactoryRelaunchDryRunMatrix` (including classifier acceptance), `TestFactoryRelaunchRefusesLegacyAndCodexPins`, `TestFactoryRelaunchFromRunRetiresDeadOwnerOnly`, `TestFactoryRelaunchDoesNotMutateRunRecords`, `TestFactoryRelaunchHelpDistinguishesClearPolicy`.
3. **Legacy notice text** (REQ-SRH-001, REQ-SRH-003) in `session_stale_run.go` and `stale_run_gate.go`: N5 and N6 take their command lines as interpolated protocol tokens (rows R6-R9), computed from the measured run state and a new read-only active-run listing in `factorymsg/run_state.go`; the unbind line stays conditional on an active run; the operator-only framing is added. Test: `TestStaleNoticeCarriesExecutableRelaunch`.
4. **Literal-text test updates** (AC-SRH-007): the five sites of §B change to the new lines; every cadence assertion is retained verbatim.
5. RED first for 1-3: E1, E7, E9, E10 are pinned; the Go-test REDs are recorded here. Mutants of `acceptance.md` D.2 for AC-SRH-001..006.

### M2 (High) — Hook rebind: single measurement, registration, state machine

1. **Single-measurement current-vocabulary path** (REQ-SRH-009): resolver variable + one `ProbeRunStateAt`, replacing `ValidateActiveRun`; counting seams; the budget and listing variables of §B. Test: `TestHealthyLanePathUnchangedAndFailOpen`; `probe control-healthy` PASS before and after.
2. **Not-active branch** (REQ-SRH-004..007): SessionStart returns nothing; UserPromptSubmit lists active runs; one → register the peer in the sole active run under the same slot, return the rebound run; zero → N2; several → N3 (three lines and a count); live-owner refusal → N4; legacy labels never reach the registration. The `state` marker field gates emission. Tests: `TestLaneRebindsIntoSoleActiveRun`, `TestUnboundThenRebound`, `TestRebindAmbiguousActiveRuns`, `TestRebindEligibleSessionStartSilent`, `TestLegacyLabelNeverRebindsAndLiveOwnerNotDisplaced`.
3. **Claim hand-off** (REQ-SRH-008): `factoryHookBatchForRun`; Stop unchanged. Tests: `TestReboundClaimReadsRebindRunOnly`, `TestInboundClaimIndependentOfEnvLabel` unchanged.
4. **Rebind-path timing** (REQ-SRH-010): `TestRebindPathStaysInsideBindBudget` under the production bind budget, plus the one-nanosecond fail-open case.
5. RED first: E4, E5, E6, E8 pinned; mutants for AC-SRH-008..014. The `workers`-untouched assertion (DP11) lands with the rebind test.

### M3 (Medium) — The cc/glm relaunch loop follows the run

1. Add `explicit` and `leadTarget` to `runFactoryLaneRelaunch`; the two call sites pass `entry.FactoryRun` and `entry.FactoryLead`; each iteration re-resolves through `enterFactoryLaneRun` before leasing, re-stamps the environment, and stops with the gate's refusal text and no lease when it refuses (REQ-SRH-011).
2. RED first: `TestRelaunchLoopReResolvesRun` with launch, join-gate, and lease seams; mutants include "gate called once outside the loop".
3. The loop's lease semantics, stop condition on "no card", and child launch shape are untouched; the Codex loop is not changed.

### M4 (Medium) — Preservation and mechanical sweeps

1. `TestKanbanRelaunchProseUnchanged` (golden bytes, four locales) and the kanban-file diff command of AC-SRH-016.
2. Characterization re-runs: the seven tests of AC-SRH-007, `TestInboundClaimIndependentOfEnvLabel`, `TestCurrentVocabularyBindPathUnchanged`; every probe scenario.
3. Lint, vet, the Windows cross-build, i18n parity of N5/N6, the full AC matrix, E1-E6.

## G. Anti-Patterns

- Do NOT change the existing claim sequence for non-rebound sessions or Stop; the rebound claim is a parameter of the invocation.
- Do NOT call `ResolveActiveRun` or `ReconcileActiveRuns` from a hook: both write. The hook reads a listing.
- Do NOT classify run owners inside the hook (DP3).
- Do NOT resolve the factory database path twice on one path (the gate's measured budget lesson).
- Do NOT print a command containing `<` or `>`, and never truncate one; a placeholder or a cut command is the defect this SPEC removes.
- Do NOT map a legacy label to a lane number anywhere (REQ-RNC-009), including inside the verb and the notices.
- Do NOT register at SessionStart, and do NOT emit the degraded string for a current-vocabulary session on a not-active run.
- Do NOT write the `workers` registry from the hook, and do NOT reuse `ClaimFactoryLaneWithin` for a rebind.
- Do NOT print `--lane` or `--run` for `codex`, and do NOT add provider-specific launch code to the verb.
- Do NOT persist a rebind run carrier; the only new persisted state is the marker's `state` field.
- Do NOT touch the launcher refusal texts, the kanban prose, any kanban-only file, or the Codex per-card loop.

## H. Cross-References

- `spec.md` §C root cause, §D requirements and tables (§D.7 lines, §D.8 catalogue, §D.9 supersession), §G decision points
- `acceptance.md` AC matrix, evidence ledger, mutant probes; `probe/hook-probe.sh` the behavioural instrument
- SPEC-FACTORY-LANE-JOIN-SOCKET-001 (the join gate, REQ-010 one shared gate), SPEC-STALE-RUN-LABEL-001 (cadence, the gate; superseded rows in §D.9), SPEC-FACTORY-RUN-RETIRE-001 (the retirement predicate, REQ-014), SPEC-ROLE-NAMING-CODE-001 (legacy vocabulary, REQ-RNC-009), SPEC-FACTORY-SELF-DISPATCH-001 (the clear policies and the relaunch loop)
- `.claude/rules/moai/development/verification-completeness.md` §1.1-§2.1 (two-cell discipline, evidence ledger, mutant probes)
- `.moai/reports/t1345/plan-audit-iter1.md` (iteration 1 verdict; dispositions in `progress.md`)
