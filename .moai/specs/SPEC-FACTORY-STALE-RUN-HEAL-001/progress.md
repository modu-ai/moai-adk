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

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

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
