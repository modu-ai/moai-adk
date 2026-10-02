# SPEC-FACTORY-STALE-RUN-HEAL-001 — Progress

## Plan Phase (2026-10-02, card t1345)

- Status: draft — spec.md, plan.md, acceptance.md authored by manager-spec (Tier M, 3 artifacts plus this record and the behavioural probe `probe/hook-probe.sh`).
- Research: read-only; mechanism read at tree `802a72235536958ada5b7cd5876a168e4b8c325f` and observed through the probe at `cda6913d127c959cee93b54254e6c7241f8b2032` (Go identical).
- Scope boundary (leader): Factory Mode only — Kanban Mode is being removed by card t1399; no kanban-only surface is touched or depended on.
- SPEC ID self-check: `SPEC-FACTORY-STALE-RUN-HEAL-001` — regex check executed as Bash, output `PASS`; unique in `.moai/specs/`.

## Plan-audit iteration 1 (FAIL 0.65, `.moai/reports/t1345/plan-audit-iter1.md`, audited `cda6913d1`) and iteration 2 revision (spec/plan/acceptance 0.2.0)

Counts after revision: 16 requirements (Tier M ceiling 16), 16 acceptance criteria (ceiling 16); 9 Release-blocking (each with an attached RED-now cell), 7 High regression-guards (reason per row).

Lane decisions applied as given: Codex option (a); Codex per-card loop out of scope; the hook rebind heals current-vocabulary sessions only; SPEC-STALE-RUN-LABEL-001 not amended; additive read of the retired run's broker dropped; single-measurement healthy path; the existing admission-checked claim path evaluated for lane admission.

| Defect | Disposition | Where |
|---|---|---|
| D1 codex launch line refused | fixed (decided: option a) — verb prints and launches only `moai codex -f lane`; `--lane`/`--run` refused with `--provider codex`; A5, DP9, §C item 6, REQ-SRH-012/016, AC-SRH-001/002 corrected; classifier-acceptance assertion added; RED E1 + E2/E3 (`moai codex -f lane-3` exit 1) | spec §C.6, §D.5, §G; acceptance AC-001/002, E2/E3 |
| D2 hook rebind does not repair the measured legacy incident | fixed — §B and §H state that the rebind heals current-vocabulary sessions only and the legacy lanes (worker-69/72) get the printed command for a new session; the causal sentence is rewritten; no legacy rebinding added (REQ-RNC-009) | spec §B, §H |
| D3 SPEC-STALE-RUN-LABEL-001 reconciliation asserted, not shown | fixed (decided: no amendment) — per-requirement table (REQ-SRL-001..009, AC-SRL-005(b): preserved / superseded / amended) and an explicit supersession clause | spec §D.9, DP7 |
| D4 "final" unbind vs later rebound | fixed — notice state machine (unbound is final for legacy only; not final for current-vocabulary), REQ-SRH-003/004 amended, AC-SRH-009 added (unbind-then-active-run) | spec REQ-SRH-003, §D.8; acceptance AC-009, E5 |
| D5 claim path cannot know "rebound"; budget statement inconsistent; no timing AC | fixed — one resolution + one `ProbeRunStateAt` replaces `ValidateActiveRun` (count unchanged); registration returns the rebound run to the same invocation's claim; Stop does not rebind (stated, with its consequence in §H); budgets named per step; injectable seams; timing/fail-open AC | plan §B, §F M2; spec REQ-SRH-008..010; acceptance AC-014 |
| D6 AC-SRH-009 vacuous | fixed — REQ-SRH-005 states silence for any current-vocabulary session on a not-active run; the AC asserts absence of the degraded string; RED observed (E8); old AC-009 is now AC-012 | spec REQ-SRH-005; acceptance AC-012, E8 |
| D7 ten Release-blocking ACs without RED-now | fixed — Release-blocking ACs 001, 002, 004, 005, 006, 008, 009, 010, 012 carry plan-phase RED cells (verb absence and the behavioural probe, E1-E10); AC-003, 007, 011, 013, 014, 015, 016 reclassified High (regression-guard) with the reason per row; none is Release-blocking with a deferred RED | acceptance matrix, D.1 |
| D8 grep proxies as RED; guard ACs marked Release-blocking | fixed — E3/E4 demoted to hints H1-H3, none gates an AC; the behavioural cells replace them; guard ACs reclassified; the "gate called once outside the loop" mutant added to AC-SRH-015's probe row | acceptance D.1 hints, D.2 |
| D9 notice line under-specified; AC-005 shallow | fixed — exact-line table (vocabulary × run state × active-run count), `unknown` backend maps to `cc`, wrong-run-id mutant with exact-line assertion, current-vocabulary rows moved to M2 ACs (008-011) | spec §D.7; acceptance AC-005, D.2 |
| D10 new notice catalogue unspecified; D7 misdescribed current behaviour | fixed — message catalogue N1-N6 (purpose, surface/locales, protocol tokens, cadence carrier); the current-vocabulary account corrected (it receives the per-prompt degraded string, observed E4/E8) | spec §D.8, §C.2 |
| D11 hook registration bypasses lane admission | decided — `ClaimFactoryLaneWithin` infeasible (the slot is held by the session's own live pid → "already occupied"); declared bypass as a tested exclusion with its consequence for the free-slot view (unchanged: pid-keyed) and for declared capacity (not enforced; parity with cc/glm joins) | spec DP11, REQ-SRH-007; acceptance AC-008 |
| D12 Codex loop and loop plumbing | out-of-scope-with-reason (Codex per-card loop, residual risk named) + plumbing listed (signature gains `explicit`, `leadTarget`; the two call sites) | spec §F; plan §B, M3 |
| D13 additive read of the retired run's broker | fixed (decided: dropped) — a rebound session reads only the rebound run's broker; REQ-SRL-007 sentence 1/2 reconciled in the table | spec REQ-SRH-008, DP5, §D.9 |
| D14 agent-facing executable line invites running it | adopted (cheap) — the notice frames the line as the operator's, to run from a terminal after ending the session | spec REQ-SRH-001, §H |
| D15 load-dependent hook ACs | adopted — injectable budget/resolver/listing seams; the hook ACs do not depend on wall-clock; one elapsed bound two orders above measured cost | plan §B; acceptance preamble, AC-014 |
| D16 AC-015 pathspec, `$(...)`, t1399 overlap, two uncovered tests | adopted — pathspec widened to the kanban files and package with the new builder files excluded; command rewritten without `$(...)` against a pinned SHA; t1399 overlap named; the two literal-assertion tests added to AC-SRH-007 | acceptance AC-016, AC-007; plan §B |
| D17 REQ-002/012 bundle many behaviours | partly adopted — REQ-002 split (grammar vs provider/Codex, REQ-SRH-016); REQ-012 kept whole to stay at the 16-requirement ceiling | spec §D.1, §D.5 |
| D18 A4 unmeasured | adopted (state the consequence) — A4 stays an unmeasured assumption; §H states that an autonomous lane without operator prompts stays unrebound | spec §G A4, §H |

- 2026-10-02 — plan-audit iteration 2 (PASS-WITH-DEBT 0.85) debt F1 repaired, nothing else changed: the seven probe-cited RED cells (AC-SRH-004, 005, 006, 008, 009, 010, 012) are re-pinned to `e48d22fc4a14b1f3105ad3129c1c9f910c4afd2a` (the commit that contains `probe/hook-probe.sh`; Go code equals `802a72235`) and written as literal single-invocation commands; `probe/hook-probe.sh` gained an optional binary argument defaulting to `./bin/moai-t1345` (scenarios unchanged). Evidence ledger `acceptance.md` D.1.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-02
plan_audit_iteration: 2 (final, ceiling 2)

- Artifacts: spec.md, plan.md, acceptance.md, progress.md, probe/hook-probe.sh under `.moai/specs/SPEC-FACTORY-STALE-RUN-HEAL-001/`.
- Frontmatter: 12 canonical fields plus `tier: M`, `card`, `depends_on`, `related_specs`; `status: draft`; version 0.2.0.
- Out of Scope: eight H3 topics including the Kanban Mode and Codex per-card loop exclusions.
- Open items: the High ACs' REDs are acquired at M1/M2/M3 RED (reasons per row); A4 (UserPromptSubmit firing on autonomous lanes) is unmeasured; decisions DP1-DP12 await operator or auditor confirmation at the Kickoff.

## §E.2 Run-phase Evidence

### M1 — command grammar, the verb, legacy notice text (commit 97115dd2c on 0463c01e4; this run, this tree; cycle tdd; draft -> in-progress in the same commit)

Pre-flight (this tree, `0463c01e4`): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; golangci-lint v2.1.6 on `./internal/kanban/... ./internal/hook/... ./internal/factorymsg/... ./internal/cli/` = `0 issues.` (baseline); `probe control-healthy` = `VERDICT: PASS`. Plan §C: item 1 `retireRun` writes only `status='retired'` (`UPDATE runs SET status='retired' ... AND status='active'`); item 3 `roleValueRetire` is reachable only through `staleRunNotice` with run id plus the factory env, which production reaches from no caller (tests only) — it was updated with `laneLabelRetire` (R6) so no divergent string remains; item 4 the verb re-executes `moai <provider> -f ...` as a child through a launch seam, no provider-specific code (codex line is the bare `moai codex -f lane`).

RED (verbatim first lines, stubs compiled so each reached its intended assertion):

- builder (`go test ./internal/kanban/ -run '^(TestRelaunchProviderForBackend|TestRelaunchCommandLineAndLaunchLine|TestRelaunchNoticeTable)$'`, exit 1, 21 FAIL lines): `RelaunchProviderForBackend("gpt") = "", want "codex"`; `Line() = "", want "moai factory relaunch --provider cc"`.
- verb (7 tests, exit 1): `TestFactoryRelaunchDryRunMatrix/cc_pinned_lane: stdout = "", want "moai cc -f lane-3\n"`; `TestFactoryRelaunchRefusesLegacyAndCodexPins: positive control failed: stdout="" stderr="" err=<nil>`; `TestFactoryRelaunchHelpDistinguishesClearPolicy: help does not distinguish the verb from --clear-policy relaunch`; `TestRelaunchCommandRoundTrip: verb on "moai factory relaunch --provider cc" printed "", want "moai cc -f lane\n"`.
- listing (`TestActiveRunIDsAt`, exit 1): `ActiveRunIDsAt = [], want ["run-a" "run-b"]`; `ActiveRunIDsAt on a corrupt database = ([], nil), want an error`.
- hook (new test plus the five updated literal sites, exit 1): `stale_run_relaunch_test.go:86: peer/en: command lines = [], want ["moai factory relaunch --provider cc --from-run runX"]` with the old prose `... retire the run with 'moai factory runs --retire runX', then relaunch`; `TestStaleRunNoticeFiresWhenRunActive`, `TestStaleRunNoticeOncePerSession`, `TestUnbindNoticeRebindLinePresence`, `TestStaleRunNoticeFactoryLegacyLabel`, `TestRoleNamingM3StaleRunNoticeNamesRetireStep` FAIL on the literal-text lines; `TestStaleRunNoticeSilentWhenRunRetired` and `TestUnbindNoticeThenSilence` (cadence) stayed green.

GREEN (this run, tree `97115dd2c`): kanban 3/3 PASS; cli 9/9 PASS (`TestFactoryRelaunchDryRunMatrix`, `...RefusesLegacyAndCodexPins`, `...FromRunRetiresDeadOwnerOnly`, `...DoesNotMutateRunRecords`, `...HelpDistinguishesClearPolicy`, `TestRelaunchCommandRoundTrip` (24 combinations), `...LaunchesProviderEntry`, `...SurfacesLaunchFailure`, `...FromRunUnavailableStateLeavesRunUntouched`); factorymsg `TestActiveRunIDsAt`, `TestValidRunID` PASS; hook 15/15 PASS including the seven AC-SRH-007 tests and `TestStaleNoticeCarriesExecutableRelaunch` (16 sub-cases: R6-R9 x backends claude/glm/gpt/empty, five surfaces each); wider hook sweep `-run 'Stale|Unbind|FactoryBootstrap|RoleNaming|FactoryMessage|FactoryHook'` ok 85.082s. Probes on a binary built from the tree: `legacy-lines`, `roundtrip`, `dry-run-nonmutation`, `control-healthy` = `VERDICT: PASS`; the M2-owned `rebind`, `unbind-then-rebind`, `ambiguous`, `session-start-silent` still FAIL with their pre-change reasons (`prompt 1 still says degraded`, `SessionStart still says 'factory messaging degraded'`). `go vet` on the four packages exit 0; golangci-lint = `0 issues.`; both builds exit 0.

Mutants of `acceptance.md` D.2 executed, each failure observed and the file restored (11): 001 (`-f lane` always; codex prints `lane-3`) killed by `TestFactoryRelaunchDryRunMatrix` and `TestRelaunchCommandLineAndLaunchLine`/`TestRelaunchCommandRoundTrip`; 002 (legacy mapped to a lane; codex accepts `--lane`) killed by `TestFactoryRelaunchRefusesLegacyAndCodexPins`; 003 (retire whenever `--from-run`; never retire) killed by `TestFactoryRelaunchFromRunRetiresDeadOwnerOnly` (`runLive status = "retired", want "active"`; `runDead status = "active", want "retired"`); 004 (`--dry-run` also retires) killed by `TestFactoryRelaunchDoesNotMutateRunRecords`; 005 (ja `lane-<n>` kept; R6 `--run X --from-run X`; wrong run id on R9) killed by `TestStaleNoticeCarriesExecutableRelaunch` and `TestRelaunchNoticeTable`; 006 (builder prints a flag the verb lacks) killed by `TestRelaunchCommandRoundTrip` (`unknown flag: --retire-run`). The first execution of mutant 005 (ja `roleValueRetire`) SURVIVED: the new test did not drive the `staleRunNotice` locale field; the test now adds `staleRunNotice` x four locales on the R6 row and the mutant is killed.

Deviations and findings: (1) `factoryGateBudget` was converted to a variable in M1 (plan §B lists the seam under M2): at load averages 28-47 the 200 ms gate budget answered `factory messaging degraded: context deadline exceeded` for a healthy database (observed in probe `legacy-lines` once and in one run of the new hook test, both PASS on re-run); the gate tests (`srlGateEnv`) and the new test now pin 30 s. `factoryHookInspectionDeadline` and the other seams stay for M2. (2) `roleValueRetire` unreachable from production (above). (3) The pre-existing gofmt finding `internal/cli/mcp_claude.go` is untouched. No template mirror exists for the touched text (`internal/template/templates` has no match for `stale run:` or `factory runs --retire`).

Gaps: package-wide coverage was not measured (a full-package run is left to CI); scoped coverage of the new code from the named tests — kanban builder functions 100%, `buildRelaunchCommand` 94.7%, `newFactoryRelaunchCommand` 93.5%, `relaunchRetireFromRun` 81.0% (the working-directory, path-resolution and open-failure branches are not driven), `ActiveRunIDsAt` 90.0%. RED-before-GREEN ordering is attested by the saved outputs, not witnessed by the commit graph (tests and implementation share one commit). A measurement command using a pipe with `${PIPESTATUS[0]}` was refused by the worktree guard and re-run as redirected single commands.

### M2 — hook rebind, single measurement, notice state machine (commit adc4a51c0 on ce5877450; this run, this tree; cycle tdd)

Pre-flight (this tree `ce5877450`): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; golangci-lint v2.1.6 on the four packages `0 issues.`; `probe control-healthy` `VERDICT: PASS`. Plan §C item 2/6 by code reading (`store.go` RegisterPeer): a slot absent in a second run's broker inserts at generation 1; a same-session same-pid row keeps its generation; a different session on the same pid and a dead old owner bump the generation; only a live other owner is refused ("factory logical lane has a live owner" / "factory session UUID has a live owner"). Both refusals are now the typed `liveOwnerError` (messages unchanged).

RED (verbatim first assertion lines, tests compiled against behaviour-preserving seams so each reached its intended assertion; the same 11 tests, 7 FAIL groups): `TestLaneRebindsIntoSoleActiveRun: prompt 1 lacks "factory lane rebound:": "factory messaging degraded: NO_ACTIVE_FACTORY"`; `TestUnboundThenRebound: prompt 1 lacks "factory lane unbound:" ...degraded...`; `TestRebindAmbiguousActiveRuns: command lines = [], want [... --lane lane-3 --run runY ... runZ]` (three subtests); `TestRebindEligibleSessionStartSilent: SessionStart said "factory messaging degraded: NO_ACTIVE_FACTORY"` (three subtests); `TestLegacyLabelNeverRebindsAndLiveOwnerNotDisplaced/live owner: no refusal notice: "factory messaging degraded: NO_ACTIVE_FACTORY"`; `TestReboundClaimReadsRebindRunOnly: precondition: the first prompt did not rebind`; `TestHealthyLanePathUnchangedAndFailOpen: prompt path counts = {resolve:0 probe:0 list:0}, want one resolution, one query, no listing` (+ two fail-open subtests: degraded string carried NO_ACTIVE_FACTORY instead of the measured error); `TestRebindPathStaysInsideBindBudget/production bind budget: the rebind did not complete inside the production bind budget: "factory messaging degraded: NO_ACTIVE_FACTORY" (elapsed 67ms)`. Guard tests green at RED by design: `TestHealthyLaneNeverRebinds`, `TestLeaderOnNotActiveRunKeepsDegradedString`, `TestFactoryHookBatchForRunOpensTheNamedRun`, the legacy half of AC-SRH-011, the corrupt-database and run-record-untouched subtests. Pre-change probes: `rebind`, `unbind-then-rebind`, `ambiguous`, `session-start-silent` FAIL as in E4-E6, E8.

GREEN (this run, tree `adc4a51c0`): the 11 M2 tests, the 7 AC-SRH-007 tests, `TestInboundClaimIndependentOfEnvLabel`, `TestCurrentVocabularyBindPathUnchanged` and `TestStaleNoticeCarriesExecutableRelaunch` = 21 of 21 `--- PASS` (`ok internal/hook 97.137s`, `uptime` load averages 76.97 60.05 54.10 right after); wider hook sweep `-run '^(TestFactory|TestStale|TestUnbind|TestRoleNaming|TestInbound|TestCurrentVocabulary|TestPrescription|TestLane|TestHealthy|TestRebind|TestLeader|TestLegacy|TestReboundClaim|TestUnbound|TestSRL|TestSrl)'` `ok 253.615s`; `internal/factorymsg` `-run '^Test'` `ok 101.822s`. Probes on a binary built from the tree: `rebind`, `unbind-then-rebind`, `ambiguous`, `session-start-silent`, `control-healthy`, `legacy-lines`, `roundtrip`, `dry-run-nonmutation` all `VERDICT: PASS`. `go vet` on the four packages exit 0; golangci-lint `0 issues.`; both builds exit 0.

Healthy-path measurement counts (counting seams, `TestHealthyLanePathUnchangedAndFailOpen`): UserPromptSubmit, repeat prompt and SessionStart each = path resolutions 1, run-state queries 1, active-run listings 0; the output is the unchanged `factory messaging bound: run=runX slot=lane-3 generation=1; ...` / empty. RED was `{resolve:0 probe:0 list:0}`.

Rebind-path timing (`TestRebindPathStaysInsideBindBudget`, production `factoryBindBudget` 2 s): elapsed 322 ms (load 35) and 536 ms (load 77) of the 2 s budget; the one-nanosecond case returns silence or the degraded string, registers nothing and writes no marker.

Mutants of `acceptance.md` D.2 executed, each failure observed and the files restored (`git status --short` empty afterwards) (13 plus the DP11 mutant): 008 (registers into the retired run; no not-active guard) killed by `TestLaneRebindsIntoSoleActiveRun` (`runY peers = ""`), `TestHealthyLaneNeverRebinds` (`healthy lane did not bind into its own run`); 009 (unbound treated as final; notice re-emitted every prompt) killed by `TestUnboundThenRebound` (`prompt 3 did not rebind ... : ""`) and the prompt-2 repeat assertions; 010 (picks the first of two active runs) killed by `TestRebindAmbiguousActiveRuns` (`command lines = []`); 011 (a legacy label rebound; a live owner overwritten at the store) killed by `...LegacyLabel...` (`legacy notice lines = []`; `no refusal notice: "factory lane rebound: ... generation 2"`); 012 (binds at SessionStart; emits the degraded string there) killed by `TestRebindEligibleSessionStartSilent` (both); 013 (the claim reads the environment run after a rebind; Stop rebinds) killed by `TestReboundClaimReadsRebindRunOnly` (`the rebound claim read the environment run's broker`; `Stop did not read the environment run's broker`) — the "reads both brokers" variant was not built (see Gaps); 014 (listing on the healthy path; two measurements) killed by `TestHealthyLanePathUnchangedAndFailOpen` (`{resolve:1 probe:1 list:1}`; `{resolve:1 probe:2 list:0}`) and `TestLeaderOnNotActiveRunKeepsDegradedString`; DP11 (the rebind deletes the lane's `workers` row) killed by `TestLaneRebindsIntoSoleActiveRun` (`the rebind touched the workers registry (DP11)`).

Deviations from the plan: (1) `registerFactoryUserPromptPeer` and `registerFactoryHookPeer` keep their single-value signatures (about 25 call sites in existing tests); the two-value form is `registerFactoryHookPeerRun`, which `user_prompt_submit.go` calls — behaviour as planned, names differ. (2) `factorymsg.IsLiveOwnerRefusal` (a typed error; messages unchanged) was added so the refusal is recognised without matching text. (3) The malformed-run-id check keeps its old text (`invalid factory run id`) before the measurement; a failed query now reads `factory messaging degraded: <err>` instead of the collapsed `NO_ACTIVE_FACTORY` (REQ-SRH-009 mapping). (4) The notice marker helper `markFactoryNotice` now delegates to `updateFactoryNoticeMarker`, which the state field also uses. No template mirror exists for any touched file (`internal/template/templates` and `.claude/hooks` hold no match for the touched symbols or text); `AskUserQuestion` grep on the changed non-test files is empty.

Gaps: package-wide coverage not measured; scoped coverage of the new code from the named tests alone — `rebindFactoryLane` 100.0%, `emitReboundState` 100.0%, `emitFactoryLaneState` 100.0%, `registerReboundLane` 75.0% (the open-failure, peer-lookup-failure and handoff-notice branches are not driven), `registerFactoryHookPeerRun` 64.6% (the remaining branches are the existing non-rebind paths, driven by other existing tests), `factoryHookBatchForRun` 75.0%. A mutant that reads both brokers was not built (the claim is a single parameter; the single-run assertions kill the leak variants). RED-before-GREEN ordering is attested by the saved outputs, not witnessed by the commit graph (tests and implementation share one commit). The A4 assumption (UserPromptSubmit firing on autonomous lanes) remains unmeasured (spec §H).

### M3 — the cc/glm relaunch loop follows the run (commit 16b650d7e on 63d573a57; this run, this tree; cycle tdd)

Pre-flight (this tree `63d573a57`, clean): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0. Plan §C item 5 (the loop's RED) was measured at the pre-change loop by the RED below. Observed behaviour of `enterFactoryLaneRun` called per iteration: it prints nothing and its only side effects are the `MOAI_KANBAN_ID` stamp (restored by the returned func), the `ResolveActiveRun` reconcile the launcher's own call already performs, and on the discovery path the leader-name stamp (also restored) — none breaks the loop, so no blocker.

RED (scaffold first: the new signature `runFactoryLaneRelaunch(cmd, label, claudeArgs, explicit, leadTarget)`, the two call sites, and the `factoryLaneRunGateFn` / `factoryLaneLeaseFn` seams, with the loop's behaviour unchanged, so the test compiled and reached its assertions; `go test ./internal/cli/ -run '^(TestRelaunchLoopReResolvesRun)$' -count=1 -v`, exit 1):

```text
    factory_lane_relaunch_rerun_test.go:120: the join gate was called 0 times, want once per iteration (3: two cards plus the no-card iteration)
    factory_lane_relaunch_rerun_test.go:123: leases were taken from runs [X X X], want [X Y Y] (iteration 2 must lease from the NEW run)
    factory_lane_relaunch_rerun_test.go:126: children saw MOAI_KANBAN_ID=[X X], want [X Y] (iteration 2's child must carry the new run)
    factory_lane_relaunch_rerun_test.go:136: loop error = <nil>, want one carrying the gate's NO_ACTIVE_FACTORY text
    factory_lane_relaunch_rerun_test.go:149: loop error = <nil>, want the explicit selection's refusal
--- FAIL: TestRelaunchLoopReResolvesRun (2.09s)
```

GREEN: each iteration is `factoryLaneRelaunchIteration`: it re-enters the gate (`factoryLaneRunGateFn(launchProjectRoot(), explicit, leadTarget, nil)`), defers the returned restore for the pass (the child launched inside the pass inherits the stamp the gate just set), reads the stamped run id, leases from it, and on a gate refusal returns `factory lane: re-join the factory run: <gate text>` leasing nothing. `TestRelaunchLoopReResolvesRun` 1 test, 3 subtests, `--- PASS`; the real-gate loop tests `TestSD_AC020_ClearPolicies` (relaunch subtests: one session per card, binary missing refused, failed child continues) PASS unchanged in the same run (this proves the unseamed gate path across iterations).

Mutants of `acceptance.md` D.2 for AC-SRH-015 executed, each failure observed and the file restored (`cmp` against a saved copy empty, `git status --short` clean): (a) gate called once outside the loop, not per iteration killed by the iteration-2 assertions (`gate called 1 times`; `leases were taken from runs [X X X], want [X Y Y]`; `children saw MOAI_KANBAN_ID=[X X], want [X Y]`); (b) the gate's refusal ignored and the lease taken anyway killed (`loop error = <nil>`, plus `join gate called 3..6 times, the scenario scripts only 2 answers`); (c) the explicit selector dropped (`""`) killed (`gate calls = [{explicit: leadTarget:lead-9} ...], want [{explicit:X leadTarget:lead-9} ...]`); (d) the run id read before the gate (stale) killed (`leases were taken from runs [X X X], want [X Y Y]`).

Deviations: (1) the gate is called with `launchProjectRoot()` (the launcher's own root: `resolveProjectDir()`), not the `factoryCardRoot()` the loop leases against, so every iteration resolves exactly the way the launcher's first call did; the plan names a bare `root`. (2) The refusal is wrapped (`factory lane: re-join the factory run: %w`), so the gate's text is carried, not replaced. (3) The re-stamp of `MOAI_KANBAN_ID` is performed by the gate itself (it stamps the resolved id), not by a second `os.Setenv` in the loop. No template mirror exists for `factory_lane_relaunch.go`, `cc.go` or `glm.go`.

Gaps: scoped coverage from the two loop tests alone: `runFactoryLaneRelaunch` 93.8%, `factoryLaneRelaunchIteration` 88.2% (the lease-error and worktree-ensure-error branches are not driven), `launchFactoryLaneCardSession` 100.0%. The Codex loop `runCodexFactoryLane` is unchanged by decision (spec §F), so a Codex lane still reads its run once. RED-before-GREEN ordering is attested by the saved output, not witnessed by the commit graph (the test and the implementation share one commit).

### M4 — preservation and mechanical sweeps (commits 585a61136 and 773bc3953 on 16b650d7e; this run, this tree; guards, green on arrival)

Tests added (both guards, green at arrival, each with an observed mutant): `TestKanbanRelaunchProseUnchanged` (`internal/hook/kanban_relaunch_prose_test.go`: the four locales' `roleValueRelaunch` equal their pre-change bytes of tree `802a72235`, plus the rendered kanban notice; mutant: a full stop appended to the English prose FAILS with `roleValueRelaunch[en] changed`, file restored) and `TestStaleRunLocalesProtocolTokenParity` (`internal/hook/stale_run_i18n_parity_test.go`: per field the sorted format verbs with their explicit indices and the quoted commands, plus the interior newline count, equal across en/ko/ja/zh; mutant: the ja `laneLabelRetire` without its `\n%[3]s` command line FAILS with `protocol tokens = [%[1]q %[2]s], want [%[1]q %[2]s %[3]s]`, file restored). The golden is also confirmed against the base file: `git diff 802a72235 HEAD -- internal/hook/session_stale_run.go` has no `+`/`-` line touching the `roleValueRelaunch` strings.

AC-SRH-016 diff command exactly as written in `acceptance.md`, run at HEAD `773bc3953`: stdout empty, exit 0. Positive controls: the same range with the same pathspec but without the two exclusions lists `internal/kanban/factory_relaunch_cmd.go` and `internal/kanban/factory_relaunch_cmd_test.go`; the same range without a pathspec lists 29 files (pre-merge reading only).

Characterization and AC re-runs on a binary built from this tree (`./bin/moai-t1345 version` = `v3.1.3   773bc3953   built unknown`; `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=..." -o ./bin/moai-t1345 ./cmd/moai`): hook package, 19 named tests, 19 `--- PASS`, `ok internal/hook 86.358s` (the eight M2 AC tests, the seven AC-SRH-007 tests, `TestInboundClaimIndependentOfEnvLabel`, `TestCurrentVocabularyBindPathUnchanged`, `TestStaleNoticeCarriesExecutableRelaunch`, `TestKanbanRelaunchProseUnchanged` — the list in the E1 table); cli package, 11 named tests, 11 `--- PASS`, `ok internal/cli 31.998s` (the nine M1 tests, `TestRelaunchLoopReResolvesRun`, `TestSD_AC020_ClearPolicies`); kanban 3/3 PASS. All eight probe scenarios on that binary end `VERDICT: PASS`: `control-healthy`, `rebind`, `unbind-then-rebind`, `ambiguous`, `legacy-lines`, `session-start-silent`, `roundtrip`, `dry-run-nonmutation` (each PASS branch has now executed). Binary commands: `moai factory relaunch --dry-run` prints `moai cc -f lane-3`, `moai glm -f lane-3`, `moai cc -f lane-3 --factory-run runA`, `moai cc -f lane`, `moai codex -f lane` (exit 0); `--lane worker-3 --provider cc`, `--provider codex --lane lane-3`, `--provider codex --run runA` each exit 1 with the legacy/Codex-limitation text.

Quality gates (this run, this tree `773bc3953`, binary commit the same): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/kanban/... ./internal/hook/... ./internal/factorymsg/... ./internal/cli/...` exit 0, empty output; `golangci-lint run --timeout=10m` (v2.1.6) on the same four package trees `0 issues.` (the M1 baseline was `0 issues.`, so no issue was added); `moai spec lint SPEC-FACTORY-STALE-RUN-HEAL-001 --strict` on the tree-built binary (judging build `773bc3953`, tree HEAD `773bc3953`) `✓ No findings — all SPEC documents are valid`, exit 0; `AskUserQuestion` / `mcp__askuser` grep on the 13 changed non-test Go files empty (positive control: `grep -c 'package '` on three of them returns 1 each); no `"MOAI_*"` environment-name literal in the added non-test lines (positive control: `config.Env...` constants appear in 5 added lines); `git diff --name-only 802a72235 HEAD -- internal/template` empty and `internal/template/templates` holds no file named like any touched source (positive control: the same `find` reaches `CLAUDE.md` there). `git diff --stat 802a72235 HEAD` lists 28 files, all under `internal/cli`, `internal/factorymsg`, `internal/hook`, `internal/kanban` and this SPEC directory.

Gaps: the whole-package suites of `internal/cli` and `internal/hook` were not run (scoped, per the lane rule); the pre-existing sensitivity of `TestSD_AC020_ClearPolicies` to a lane-stamped environment (it fails with `lane boundary: a lane session cannot mutate the queue` unless the full `unset MOAI_FACTORY_*` form is used) is unchanged and unrelated. Baseline lint for the later milestones was not re-measured on `63d573a57`; the current tree has 0 issues, so none is new. The A4 assumption (UserPromptSubmit firing on autonomous lanes) remains unmeasured (spec §H). RED-before-GREEN ordering is attested by saved outputs, not witnessed by the commit graph.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-02
run_head: the commit carrying this section (code and tests complete at 773bc3953)
cycle_type: tdd (M1 97115dd2c, M2 adc4a51c0, M3 16b650d7e, M4 585a61136 and 773bc3953; evidence commits ce5877450, 63d573a57 and this one)

AC matrix (this run, this tree; every Go-test AC shows one `--- PASS` per named test; every probe ends `VERDICT: PASS`):

| AC | Result | Instrument and observed output |
|---|---|---|
| AC-SRH-001 | PASS | five `moai factory relaunch --dry-run` lines equal the five required strings; `TestFactoryRelaunchDryRunMatrix` `--- PASS` (incl. launcher-classifier acceptance) |
| AC-SRH-002 | PASS | `--lane worker-3`, `--provider codex --lane lane-3`, `--provider codex --run runA` exit 1; `TestFactoryRelaunchRefusesLegacyAndCodexPins` `--- PASS` (positive control inside) |
| AC-SRH-003 | PASS | `TestFactoryRelaunchFromRunRetiresDeadOwnerOnly` `--- PASS` (plus `...UnavailableStateLeavesRunUntouched`) |
| AC-SRH-004 | PASS | probe `dry-run-nonmutation` `VERDICT: PASS`; `TestFactoryRelaunchDoesNotMutateRunRecords`, `TestFactoryRelaunchHelpDistinguishesClearPolicy` `--- PASS` |
| AC-SRH-005 | PASS | probe `legacy-lines` `VERDICT: PASS`; `TestStaleNoticeCarriesExecutableRelaunch` `--- PASS` |
| AC-SRH-006 | PASS | probe `roundtrip` `VERDICT: PASS`; `TestRelaunchCommandRoundTrip` `--- PASS` |
| AC-SRH-007 | PASS | the seven tests `--- PASS` (listed in the hook run: 19 of 19) |
| AC-SRH-008 | PASS | probe `rebind` `VERDICT: PASS`; `TestLaneRebindsIntoSoleActiveRun` `--- PASS` |
| AC-SRH-009 | PASS | probe `unbind-then-rebind` `VERDICT: PASS`; `TestUnboundThenRebound` `--- PASS` |
| AC-SRH-010 | PASS | probe `ambiguous` `VERDICT: PASS`; `TestRebindAmbiguousActiveRuns` `--- PASS` |
| AC-SRH-011 | PASS | `TestLegacyLabelNeverRebindsAndLiveOwnerNotDisplaced` `--- PASS` |
| AC-SRH-012 | PASS | probe `session-start-silent` `VERDICT: PASS`; `TestRebindEligibleSessionStartSilent` `--- PASS` |
| AC-SRH-013 | PASS | `TestReboundClaimReadsRebindRunOnly`, `TestInboundClaimIndependentOfEnvLabel` `--- PASS` |
| AC-SRH-014 | PASS | `TestHealthyLanePathUnchangedAndFailOpen`, `TestRebindPathStaysInsideBindBudget`, `TestCurrentVocabularyBindPathUnchanged` `--- PASS`; probe `control-healthy` `VERDICT: PASS` |
| AC-SRH-015 | PASS | `TestRelaunchLoopReResolvesRun` `--- PASS` (RED observed first, four mutants killed) |
| AC-SRH-016 | PASS | the diff command prints nothing, exit 0, with positive controls; `TestKanbanRelaunchProseUnchanged` `--- PASS` (mutant killed) |

Result: 16 of 16 PASS, 0 FAIL. Release-blocking ACs (001, 002, 004, 005, 006, 008, 009, 010, 012) all PASS with their probe or binary cell. Plan §E E1-E6 satisfied (E4 separation review: the claim sequence for non-rebound sessions and Stop is untouched, asserted by `TestInboundClaimIndependentOfEnvLabel` and `TestReboundClaimReadsRebindRunOnly`'s Stop half).

Decision points DP1-DP12: none altered; the implementation deviations are listed per milestone in §E.2 (M1: gate budget seam moved to M1; M2: two-value register helper name, typed live-owner error; M3: gate called with the launcher's root, wrapped refusal, re-stamp performed by the gate).

Open items for audit and sync: A4 (UserPromptSubmit firing on an autonomous lane) is unmeasured — spec §H states the consequence (such a lane stays unrebound); the Codex per-card loop reads its run once by decision (spec §F); plan-audit optional debt F2-F8 untouched; the "reads both brokers" mutant of AC-SRH-013 was not built (single-parameter claim, killed by the single-run assertions); RED-before-GREEN is attested by saved outputs, not by the commit graph (tests and implementation share a commit per milestone); the whole-package suites are left to CI. The `spec.md` frontmatter `status` was set to `in-progress` by M1 and is untouched by M3/M4 (manager-docs owns `implemented` and `completed` at sync).

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-02
sync_commit_sha: 162d32aeae45238bea923c3161d14eafea7fe2d4
cycle: 3-phase close (plan, run, sync); the sync commit carries `in-progress -> implemented -> completed`

- b12_self_test_a (pre-emission grep): `grep -c 'SPEC-FACTORY-STALE-RUN-HEAL-001' CHANGELOG.md` read 0 before the entry was written (this run, tree 48c9ca6e0).
- b12_self_test_b (AC count): the B12 counter on `acceptance.md` (tier M, `ac_source=.moai/specs/SPEC-FACTORY-STALE-RUN-HEAL-001/acceptance.md`) printed `live=16 excluded=0 ambiguous=0`; the CHANGELOG entry states 16.
- b12_self_test_c (paths): every implementation path cited in the CHANGELOG entry exists (`ls` of the ten files, exit 0).
- changelog_entry_position: first entry under `### Added` of the first `## [Unreleased]` block of `CHANGELOG.md`.
- frontmatter_status_transitions: `spec.md` `status: in-progress -> completed` in this commit (`updated: 2026-10-02` already current); `plan.md` and `acceptance.md` carry no status field (stateless artifacts) and are untouched.
- docs changed: `CHANGELOG.md`, `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` (one passage each, 4-locale same commit; the passage extends the lane paragraph that already documents `moai factory runs --retire`). No docs-site page names the touched behaviour (search over `docs-site`, the READMEs and `.moai/docs` for the retire, clear-policy, lane and relaunch vocabulary hit only the READMEs and two unrelated docs-site sentences).
- MX validation (sync sub-step, report only): `moai mx scan --dry` on `internal/hook` (106 tags) and `internal/kanban` (44 tags) completes with no warning; the new implementation files carry tags where they apply (`factory_lane_relaunch.go` 4, `factory_rebind.go` 1, `factory_relaunch_cmd.go` 1, `run_state.go` 4) and `internal/cli/factory_relaunch.go` carries none (a leaf command builder: one caller, no goroutine, low complexity). No tag was added or removed in the sync phase.
- Gaps: the whole-package suites remain unrun (CI); the 4-locale README passages were verified for presence and shared literals by reading, not by a locale-parity tool (none exists for READMEs); A4 and the Codex per-card loop stay as disclosed in §E.3.

## §F Phase 4 Mode Selection (2026-10-02, lane-11, orchestrator)

Input parameters: tier M; scope about 12 files across 4 Go packages (internal/hook, internal/factorymsg, internal/cli, internal/kanban) plus tests; domain count 3 (hook, factory messaging, launcher CLI); coding-heavy Go; concurrency benefit LOW (milestones share files and test seams).

| Mode | Decision | Rationale |
|---|---|---|
| direct | not selected | not trivial |
| serial | selected | coding-heavy, M1->M4 share the notice builder, the marker and test seams |
| fanout | not selected | no read-only multi-domain research left; writers would share a tree |
| sweep | not selected | not a uniform mechanical transform |

Decision: serial

Justification: milestones depend on each other (M1 command builder feeds M2 notices and M3 loop), one writer per working tree. The implementer is a general-purpose-typed spawn into the card tree, because a manager-develop-typed spawn materializes into its own L1 tree and the guards then refuse every write into the card worktree (measured lesson, card t1318).

## §F.1 Plan-to-run Kickoff decision

Form: operator form. The operator answered the Kickoff question "fix F1, then enter by operator approval" in the lane session on 2026-10-02. Autonomous form (auto-semantics 9.1) is not used because the plan-artifact hash changed after the last audit (F1 repair edited acceptance.md), and a third plan-audit would exceed the Tier M ceiling of 2.

decision record: decided_by=claude-lane-11+operator evidence_refs=.moai/reports/t1345/plan-audit-iter2.md(PASS-WITH-DEBT 0.85 on e48d22fc4),commit 5a9aeaa8d (F1 repair, transcription only),operator AskUserQuestion answer 2026-10-02 ladder_path=gate row plan-run Kickoff operator form

Residual: F2-F8 of plan-audit-iter2 stay open as optional debt; the 7 probe PASS branches have never executed (acquired at M1/M2 GREEN). Run Phase 1 plan-audit re-run is NOT executed (ceiling reached); the sync audit re-reads this record.
