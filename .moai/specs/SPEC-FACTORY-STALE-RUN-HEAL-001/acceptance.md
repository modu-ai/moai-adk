# SPEC-FACTORY-STALE-RUN-HEAL-001 — Acceptance Criteria

Every AC is mechanically verifiable without a live factory run. There are three instruments:

- **Binary commands** — a build of the card tree invoked by path (never the installed `moai`), e.g. `moai factory relaunch --dry-run ...`.
- **The behavioural probe** — `bash .moai/specs/SPEC-FACTORY-STALE-RUN-HEAL-001/probe/hook-probe.sh <scenario> <moai-binary>`: one invocation that builds a throw-away project and factory database in a temp directory outside the repository (isolated `HOME`/`MOAI_HOME`), seeds run rows, drives the real `moai hook user-prompt-submit` / `session-start` subcommands with a lane environment, prints each turn's output, and ends in `VERDICT: PASS` (exit 0) or `VERDICT: FAIL <reason>` (exit 1). It writes nothing in the repository; it needs `bash`, `git`, `jq`, `sqlite3`. The scenario `control-healthy` is its positive control: it PASSES on the pre-change tree, proving the fixture and driver can reach a PASS, so the FAIL verdicts are not vacuous.
- **Go tests** on `t.TempDir()` factory databases, env-scrubbed in ONE compound invocation (a lane-stamped environment falsifies env-reading guard tests locally):

```bash
unset MOAI_FACTORY_WORKERS MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_NAME MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test <one package> -run '^(<TestName1>|<TestName2>)$' -count=1 -v
```

**Swept-count rule (verification-completeness §1.1, §1.3).** A Go-test AC passes only when its `-v` output shows one `--- PASS` line per named test; `[no tests to run]`, a PASS count below the named count, or a missing name is a FAIL. The `-run` pattern is anchored at every alternation branch so it cannot select longer names. Each AC names the packages it touches; re-measurement is scoped to them.

**No wall-clock dependence (D15).** The hook ACs use the injectable budget and path-resolution seams of `plan.md` §B (the bind budget, the inspection deadline, the gate budget, the factory-database path resolver, and the active-run listing are package variables the tests set), so no AC verdict depends on machine load. The only wall-clock assertion is AC-SRH-014's elapsed-time bound for the rebind path, set two orders of magnitude above the measured cost (about 75 ms for a validated open, `user_prompt_submit.go`) against the 2 s bind budget.

**Two-cell discipline (verification-completeness §2).** A criterion is **Release-blocking** only when its RED-now cell is attached here — one read-only invocation, its verbatim output, its exit code, the pinned SHA — and a green-path cell names the milestone that flips it. A criterion whose RED cannot be observed by one read-only command at plan time is classified **High (regression-guard)** with the reason stated, and its RED is acquired at the owning milestone; no criterion is Release-blocking with a deferred cell.

## D. AC Matrix

| ID | Requirement | Scenario (Given / When / Then) | Packages | Severity | Test (owning milestone) and RED-now cell |
|---|---|---|---|---|---|
| AC-SRH-001 | REQ-SRH-012, REQ-SRH-016 | **Given** a build of the card tree **When** `moai factory relaunch --dry-run --lane lane-3 --provider cc` runs, and likewise `--provider glm`, `--run runA`, `--lane` omitted, and `--provider codex` **Then** stdout is exactly `moai cc -f lane-3`, `moai glm -f lane-3`, `moai cc -f lane-3 --factory-run runA`, `moai cc -f lane`, and `moai codex -f lane`, stderr empty, exit 0; and in a Go test every printed line's argv is accepted by the target launcher's own argument classifier (the cc/glm entry parse yields a lane join with no error; the Codex classifier yields the lane entry) | internal/cli | Release-blocking | Binary command + `TestFactoryRelaunchDryRunMatrix` (M1). RED: E1 (verb absent). E2/E3 record why the Codex line is not `-f lane-3`: that line is refused today |
| AC-SRH-002 | REQ-SRH-014, REQ-SRH-016 | **Given** the same build **When** `--lane worker-3 --provider cc`, `--provider codex --lane lane-3`, and `--provider codex --run runA` each run with `--dry-run` **Then** each exits non-zero with empty stdout; the first names the canonical `lane-3` and the word `legacy`, the other two name the Codex limitation; and in the same test `--lane lane-3 --provider cc` exits 0 (positive control) | internal/cli | Release-blocking | `TestFactoryRelaunchRefusesLegacyAndCodexPins` (M1). RED: E1 — the stderr content, not the exit code, is the verdict (a bare non-zero exit is already true today) |
| AC-SRH-003 | REQ-SRH-013 | **Given** a factory database with `runX` recorded active, an owner classifier injected as dead / live / indeterminate, `runR` already retired, and the launch seam stubbed **When** the verb runs with `--from-run runX` (resp. `runR`, an unknown id) **Then** dead: `runX` reads `retired` with exactly one `run.retired` event carrying classification and basis, the seam receives the launch argv; live, indeterminate, retired, unknown: no event is added, `runX` still reads `active` where it was, the seam still receives the argv, the output names the outcome | internal/cli, internal/homestate | High (regression-guard) | `TestFactoryRelaunchFromRunRetiresDeadOwnerOnly` (M1). No plan-phase RED: a retirement is observable only by launching a session, and the binary has no launch seam; RED is acquired at M1 |
| AC-SRH-004 | REQ-SRH-014 | **Given** a factory database with `runX` active and a dead owner **When** `moai factory relaunch --dry-run --provider cc --from-run runX` runs **Then** exit 0, stdout exactly `moai cc -f lane`, and the `runs` and `events` tables are identical before and after; and `moai factory relaunch --help` stdout contains `--clear-policy relaunch` and the lane clear-policy environment value is untouched | internal/cli, internal/homestate | Release-blocking | `probe dry-run-nonmutation` + `TestFactoryRelaunchDoesNotMutateRunRecords` + `TestFactoryRelaunchHelpDistinguishesClearPolicy` (M1). RED: E10 |
| AC-SRH-005 | REQ-SRH-001, REQ-SRH-002, REQ-SRH-016 | **Given** a legacy-label (`worker-69`) lane and the table rows R6 (run active), R7 (retired, none active), R8 (retired, one active), R9 (retired, two active) of `spec.md` §D.7 **When** the hook renders the notice on each surface, in each locale of that surface, for backends `claude`, `glm`, `gpt`, and an empty backend **Then** the command lines equal the table's exact lines — e.g. R6 `moai factory relaunch --provider cc --from-run runX`, R8 `moai factory relaunch --provider cc`, R9 one `--run <id>` line per candidate, codex `moai factory relaunch --provider codex` — each on its own line, no `<` or `>` anywhere in the notice, no `runs --retire`-then-`relaunch` prose, an empty backend maps to `cc`, the lines byte-identical across locales, and R7 prints no command | internal/hook, internal/kanban | Release-blocking | `probe legacy-lines` + `TestStaleNoticeCarriesExecutableRelaunch` (M1). RED: E7. The current-vocabulary rows R2-R5 are AC-SRH-008..011 (M2) |
| AC-SRH-006 | REQ-SRH-002 | **Given** the shared command builder and the verb's parser **When** the legacy line the hook prints is split into argv and fed to the verb with `--dry-run` appended, and, in a Go test, for every provider × {pinned lane, omitted lane} × {with, without `--from-run`} × {with, without `--run`} the built line is fed likewise **Then** every one exits 0 and prints the launch line its arguments imply; a builder output the verb rejects fails | internal/cli, internal/kanban | Release-blocking | `probe roundtrip` + `TestRelaunchCommandRoundTrip` (M1). RED: E9 |
| AC-SRH-007 | REQ-SRH-003 | **Given** the seven existing tests whose notice text or cadence this SPEC touches — `TestStaleRunNoticeSilentWhenRunRetired`, `TestStaleRunNoticeFiresWhenRunActive`, `TestStaleRunNoticeOncePerSession`, `TestUnbindNoticeThenSilence`, `TestUnbindNoticeRebindLinePresence`, `TestRoleNamingM3StaleRunNoticeNamesRetireStep` (`role_naming_m3_notice_test.go`), `TestStaleRunNoticeFactoryLegacyLabel` (`stale_run_m1_test.go`) — with their literal-text assertions updated to the new command lines and no cadence assertion weakened **When** they run **Then** all seven pass | internal/hook | High (regression-guard) | The seven tests, updated (M1). Baseline E11: 12 of 12 PASS on the pre-change tree in this run |
| AC-SRH-008 | REQ-SRH-004, REQ-SRH-007 | **Given** `runX` retired, exactly one other run `runY` active, an environment naming `runX` with label `lane-3`, and a UserPromptSubmit for session `s1` **When** the hook runs, twice **Then** the first output carries `factory lane rebound:` naming `runX`, `runY`, `lane-3`, and the generation, never `degraded`; `runY`'s broker holds the peer `lane-3|s1`; `runX`'s broker was never created; the second output carries no notice; and the `workers` registry rows and the leader's free-slot view (`FactoryFreeSlots`) are identical before and after (DP11) | internal/hook, internal/factorymsg, internal/kanban | Release-blocking | `probe rebind` + `TestLaneRebindsIntoSoleActiveRun` (M2). RED: E4 |
| AC-SRH-009 | REQ-SRH-003, REQ-SRH-006 | **Given** `runX` retired and no active run, then, after two prompts, `runY` becoming active **When** the hook runs three prompts of session `s1` **Then** prompt 1 carries exactly one `factory lane unbound:` notice naming `lane-3` and `runX` with no command line and no `degraded`; prompt 2 carries none; prompt 3 carries the rebound notice naming `runY` and `runY`'s broker holds `lane-3|s1` — the unbound state is not final | internal/hook | Release-blocking | `probe unbind-then-rebind` + `TestUnboundThenRebound` (M2). RED: E5 |
| AC-SRH-010 | REQ-SRH-006 | **Given** `runX` retired and two other runs `runY`, `runZ` active **When** the hook runs two prompts of `lane-3` **Then** prompt 1 prints, each on its own line, `moai factory relaunch --provider cc --lane lane-3 --run runY` and `... --run runZ`, no `degraded`, and no peer is registered in either run; prompt 2 carries no notice; with four active runs three lines and a count are printed; for backend `gpt` the one line `moai factory relaunch --provider codex` is printed | internal/hook | Release-blocking | `probe ambiguous` + `TestRebindAmbiguousActiveRuns` (M2). RED: E6 |
| AC-SRH-011 | REQ-SRH-007, REQ-SRH-003 | **Given** the AC-SRH-008 setup with the label `worker-69`, and separately `lane-3` with a different live session already holding slot `lane-3` in `runY` **When** the prompt runs **Then** legacy: no peer is registered in `runY` and the notice carries the R8 line; live owner: the existing peer row in `runY` is byte-identical afterwards, no write occurred, and one refusal notice (`factory lane rebind refused:`) with the line `moai factory relaunch --provider cc --run runY` is returned (row R3) | internal/hook, internal/factorymsg | High (regression-guard) | `TestLegacyLabelNeverRebindsAndLiveOwnerNotDisplaced` (M2). The legacy half is already true today (nothing registers); the refusal half needs the new path — RED acquired at M2 |
| AC-SRH-012 | REQ-SRH-005 | **Given** `runX` retired, `runY` active, a `lane-3` environment naming `runX` **When** the SessionStart hook runs (source `clear`, and `startup`) **Then** its output contains no `factory messaging degraded`, no peer is registered, and no broker is created for `runX` or `runY` | internal/hook | Release-blocking | `probe session-start-silent` + `TestRebindEligibleSessionStartSilent` (M2). RED: E8 — the assertion that differs is the absence of the degraded string; "no peer, no broker" is already true today |
| AC-SRH-013 | REQ-SRH-008 | **Given** a rebound session with one pending message addressed to its slot in `runY`'s broker and one in `runX`'s **When** the same invocation's inbox claim runs, and separately a Stop event runs **Then** the claim names `runY` and lists only `runY`'s message; the Stop event reads `runX`'s broker and registers nothing; and for a session that is not rebound (the worker-70 case: a live peer in the environment run's broker, legacy label) the claim output is unchanged | internal/hook | High (regression-guard) | `TestReboundClaimReadsRebindRunOnly` + existing `TestInboundClaimIndependentOfEnvLabel` (M2). No plan-phase RED: seeding a message needs the rebound run's broker, which only the implemented path creates |
| AC-SRH-014 | REQ-SRH-009, REQ-SRH-010 | **Given** a current-vocabulary `lane-3` session whose run is active **When** the prompt and SessionStart paths run with counting seams on the path resolver and the run-state accessors **Then** the output equals the pre-change fixture byte-for-byte and the per-invocation count is one path resolution and one query, the active-run listing never invoked; **Given** a corrupt factory database, or a budget seam set to one nanosecond **When** the rebind path runs **Then** the answer is the existing degraded string or silence, no error escapes, and no peer is registered; **and** in every case the `runs` and `events` row counts are unchanged by the hook, and on the rebind path the elapsed time is below the 2 s bind budget | internal/hook, internal/factorymsg | High (regression-guard) | `TestHealthyLanePathUnchangedAndFailOpen` + `TestRebindPathStaysInsideBindBudget` + existing `TestCurrentVocabularyBindPathUnchanged` (M2); `probe control-healthy` PASSES before and after (E12) |
| AC-SRH-015 | REQ-SRH-011 | **Given** the cc/glm relaunch loop with launch, join-gate, and lease seams, run `X` active at iteration 1 **When** `X` is retired and `Y` activated between iterations; separately when no run is active at iteration 2; separately with an explicit `--factory-run X` that retires **Then** iteration 2's child receives `MOAI_KANBAN_ID=Y` and the lease is taken from `Y`'s record; with no active run the loop stops with the gate's `NO_ACTIVE_FACTORY` text and the lease seam is never called; the explicit selection stops the loop; and the join gate is called once per iteration, not once before the loop | internal/cli | High (regression-guard) | `TestRelaunchLoopReResolvesRun` (M3). No plan-phase RED: the loop launches an interactive `claude` child, which no read-only single command can drive; the observation is read-only (`grep -c "enterFactoryLaneRun(" internal/cli/factory_lane_relaunch.go` reads 0 today — a hint, satisfiable by a mutant, never a RED cell) |
| AC-SRH-016 | REQ-SRH-015 | **Given** the card's implementation diff **When** `git diff --name-only cda6913d127c959cee93b54254e6c7241f8b2032..HEAD -- internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go internal/cli/kanban.go internal/kanban ':!internal/kanban/factory_relaunch_cmd.go' ':!internal/kanban/factory_relaunch_cmd_test.go'` runs **Then** stdout is empty (positive control: the same range without a pathspec lists at least one file; pre-merge only — after a merge the range holds other cards' commits); and `TestKanbanRelaunchProseUnchanged` asserts the four locales' `roleValueRelaunch` strings equal their pre-change golden bytes | internal/hook | High (regression-guard) | The diff command + `TestKanbanRelaunchProseUnchanged` (M4) |

## D.1 Evidence Ledger (RED-now cells, measured this run)

All entries measured at tree `cda6913d127c959cee93b54254e6c7241f8b2032` (card branch `WT-stale-run-healing`). The commit is a SPEC-only commit: `git diff --name-only 802a72235 HEAD` lists the four SPEC files, so every Go-code measurement is identical to `802a72235536958ada5b7cd5876a168e4b8c325f`. The binary was built from the card working tree by `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=cda6913d1" -o <scratch>/moai-t1345 ./cmd/moai` and invoked by path; its `version` prints `v3.1.3   cda6913d1   built unknown`. The scratch path is machine-local and is not a citation target; the commands and outputs are. Probe output is shown whitespace-collapsed; the prompt/step blocks are the probe's own lines.

```text
E1  command : moai factory relaunch --dry-run --lane lane-3 --provider cc
    stdout  : (empty)
    stderr  : ERROR / Unknown flag: --dry-run. / Try --help for usage.
    exit    : 1
    why red : the verb does not exist; the factory parent consumes the unknown argument
              "relaunch" and rejects the flag. (`moai factory relaunch --help` exits 0 today and
              prints the factory parent's help — a vacuous green, so no AC uses --help as RED.)

E2  command : moai codex -f lane-3
    stdout  : (empty)
    stderr  : FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader
    exit    : 1
    why     : the Codex launcher refuses a pinned lane — the line the 0.1.0 draft expected the verb to print.

E3  command : moai codex -f lane --factory-run runA
    stderr  : FACTORY_MODE_UNSUPPORTED_BACKEND: (same text as E2)
    exit    : 1

E4  command : bash .moai/specs/SPEC-FACTORY-STALE-RUN-HEAL-001/probe/hook-probe.sh rebind <binary>
    stdout  : prompt 1: factory messaging degraded: NO_ACTIVE_FACTORY
              prompt 2: factory messaging degraded: NO_ACTIVE_FACTORY
              peers in runY: (empty)
              VERDICT: FAIL prompt 1 still says degraded
    exit    : 1
    why red : a current-vocabulary lane on a retired run never rebinds and repeats the degraded string every prompt.

E5  command : bash .../probe/hook-probe.sh unbind-then-rebind <binary>
    stdout  : prompt 1/2/3: factory messaging degraded: NO_ACTIVE_FACTORY   (all three, including after runY is active)
              VERDICT: FAIL prompt 1 says degraded instead of the one-time unbound notice
    exit    : 1

E6  command : bash .../probe/hook-probe.sh ambiguous <binary>
    stdout  : prompt 1/2: factory messaging degraded: NO_ACTIVE_FACTORY
              VERDICT: FAIL prompt 1 says degraded
    exit    : 1

E7  command : bash .../probe/hook-probe.sh legacy-lines <binary>
    stdout  : legacy, run active: stale run: lane label "worker-69" is legacy vocabulary from a binary before the leader/lane rename — end this session, retire the run with 'moai factory runs --retire runX', then relaunch
              VERDICT: FAIL row R6: no exact line '--from-run runX'
    exit    : 1
    note    : the probe stops at its first failing row. A separate hook run on the same tree for row R8
              (`worker-69`, `runX` retired, `runY` active) printed the unbind text followed by
              `an active factory run exists in this project — to rejoin its slot set, end this session and relaunch with 'moai cc -f lane-<n>'`
              — the placeholder line the table replaces.

E8  command : bash .../probe/hook-probe.sh session-start-silent <binary>
    stdout  : SessionStart (source clear): moai session attribution: source_session_id=s1 / Use 'moai session current' ... (first three lines)
              VERDICT: FAIL SessionStart still says 'factory messaging degraded'
    exit    : 1
    why red : the SessionStart additionalContext carries `factory messaging degraded: NO_ACTIVE_FACTORY`
              (observed in full output); registering nothing is already true, the string is the verdict.

E9  command : bash .../probe/hook-probe.sh roundtrip <binary>
    stdout  : printed line: (empty)
              VERDICT: FAIL the notice prints no 'moai factory relaunch' line to feed the verb
    exit    : 1

E10 command : bash .../probe/hook-probe.sh dry-run-nonmutation <binary>
    stdout  : verb --dry-run --from-run runX: ERROR / Unknown flag: --dry-run. / Try --help for usage. (exit 1)
              VERDICT: FAIL the verb exited 1
    exit    : 1

E11 command : unset <factory/kanban vars> && go test ./internal/hook -run '^(TestRoleNamingM3StaleRunNoticeNamesRetireStep|TestStaleRunNoticeSilentWhenRunRetired|TestStaleRunNoticeFiresWhenRunActive|TestStaleRunNoticeOncePerSession|TestUnbindNoticeThenSilence|TestUnbindNoticeRebindLinePresence|TestPrescriptionGateUnavailableFailsOpen|TestInboundClaimIndependentOfEnvLabel|TestCurrentVocabularyBindPathUnchanged|TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord|TestStaleRunNoticeFactoryLegacyLabel)$' -count=1 -v
    stdout  : 12 `--- PASS` lines, then `ok  github.com/modu-ai/moai-adk/internal/hook  9.212s`
    exit    : 0 (`uptime` right after: load averages 41.81 48.47 46.92)
    note    : in an earlier session on this machine, at load averages ~57-63, the same anchored set failed
              2-3 tests per full-set run (different ones each time) and passed in isolation; the cause was
              not observed (no failure text captured) and is inferred to be the 200 ms gate budget. The seams
              of `plan.md` §B exist so that no hook AC depends on it.

E12 command : bash .../probe/hook-probe.sh control-healthy <binary>
    stdout  : prompt 1: factory messaging bound: run=runX slot=lane-3 generation=1; messages arrive at turn boundaries, not idle wake
              VERDICT: PASS
    exit    : 0
    role    : positive control — the fixture and driver reach a PASS; a healthy lane's behaviour (REQ-SRH-009).

Hints (NOT RED cells — each is satisfiable by prose or a comment, so none gates any AC):
H1  grep -c "lane-<n>" internal/hook/session_stale_run.go            -> 4   (the unbind re-bind line, four locales)
H2  grep -c "factory relaunch" internal/hook/session_stale_run.go    -> 0
H3  grep -c "enterFactoryLaneRun(" internal/cli/factory_lane_relaunch.go -> 0   (control internal/cli/cc.go -> 1)
```

## D.2 Mutant Probes (verification-completeness §2)

Each row names a mutant that satisfies a shallow reading while violating the requirement, and the assertion that kills it. The run phase executes every mutant at the owning milestone and records the observed failure.

| AC | Mutant | Killed by |
|---|---|---|
| AC-SRH-001 | prints `-f lane` always, ignoring `--lane`; or prints `moai codex -f lane-3` | exact-stdout rows; the classifier-acceptance assertion |
| AC-SRH-002 | accepts `worker-3` and maps it to a number; accepts `--lane` with codex | the stderr-content assertions plus the same-test positive control |
| AC-SRH-003 | retires whenever `--from-run` is given; or never retires | the live/indeterminate rows (`runX` still `active`); the dead row (`runX` retired, one event) |
| AC-SRH-004 | `--dry-run` also performs the retire pre-step | the before/after table equality (probe) |
| AC-SRH-005 | one locale keeps `lane-<n>`; or R6 prints `--run X --from-run X`; or any row prints the **wrong run id** (the dead run in `--run`, or `Y` where `X` belongs) | the per-locale scan; the **exact-line** assertion per row of §D.7 (a wrong id is a different line) |
| AC-SRH-006 | the builder emits a flag the verb does not define | the round-trip exit-0 assertion |
| AC-SRH-008 | registers into the retired run `runX`; or registers in `runY` without the not-active guard | the peer-in-`runY` and no-broker-for-`runX` assertions; a healthy-lane variant of the test |
| AC-SRH-009 | treats the unbind marker as final (never rebinds); or re-emits the unbound notice every prompt | prompt 3 rebound assertion; prompt 2 silence |
| AC-SRH-010 | picks the first of two active runs | the no-peer-in-either-run assertion |
| AC-SRH-011 | rebinds a legacy label; or overwrites a live slot owner | no-peer-in-`runY`; byte-identical peer row |
| AC-SRH-012 | binds at SessionStart; or emits the degraded string | the no-degraded assertion; no-broker assertion |
| AC-SRH-013 | reads both brokers; or reads only `runX` after a rebind; or lets Stop rebind | the claim-names-`runY`-only assertion; the Stop-registers-nothing assertion |
| AC-SRH-014 | lists active runs on the healthy path; or measures twice | the counting-seam assertion |
| AC-SRH-015 | resolves the run once, before the loop; or **calls the gate once outside the loop** | the iteration-2 `MOAI_KANBAN_ID=Y` assertion and the calls-per-iteration count |
| AC-SRH-016 | edits `roleValueRelaunch` | the golden-bytes test |

## D.3 Severity Scale

Release-blocking: the defect class this SPEC exists to kill, with a RED-now cell attached — a running current-vocabulary lane locked out after a run switch, a non-executable notice, an absent verb. High (regression-guard): preservation, scope, and guards whose RED is acquired at the milestone for the reason stated per row.

## D.4 Edge Cases

- `--lane lane` (the bare role token) is not a lane label: the verb treats it as invalid; `-f lane` is selected by omitting `--lane`.
- `--run` naming a non-active run: the launcher gate answers `NO_ACTIVE_FACTORY` (an explicit selection is a decision, not an absence); the verb surfaces that refusal and launches nothing.
- `--from-run` naming an unknown id: treated as "not active" (no retirement, outcome named), never an error that blocks the launch.
- A session whose label is current-vocabulary but whose environment names no run: not a stale session; unchanged.
- `/clear` after a rebind: the session identity changes, the process does not; the first prompt of the new identity registers it in `Y` again (a generation bump) and the rebound notice is correct again for a new identity.
- Two active runs where one has a dead owner: the hook does not classify (DP3); the ambiguity notice names both, and the verb's `--run` or `--from-run` resolves it.
- More than three active runs: three command lines, a count, and the `moai factory runs` pointer (DP10).
- Hook context budget: the N1-N4 notices are short; no command line is ever truncated (a truncated command is a non-executable command).
- A Y that retires after the rebind: the next prompt computes a different state key (`unbound:X` or `rebound:Z`), so a new notice is correct (DP12).

## D.5 Quality Gates

- `go vet` and `golangci-lint run` (CI pin v2.1.6) clean on `internal/hook`, `internal/factorymsg`, `internal/cli`, `internal/kanban`.
- `GOOS=windows GOARCH=amd64 go build ./...` passes.
- Package coverage per `quality.yaml` `test_coverage_target` on the changed packages.
- Separation review: the claim sequence for a non-rebound session and for Stop is untouched; the rebound claim is a parameter of the invocation, not a read of the environment label (REQ-SRH-008).
- Hardcoding prevention: no new `MOAI_` environment-name literal outside `internal/config/envkeys.go` and tests (diff-scoped grep at sync).
- i18n parity: the four locales of N5 and N6 exist and carry identical protocol tokens.
- `moai spec lint SPEC-FACTORY-STALE-RUN-HEAL-001 --strict` exits 0 on a binary built from the tree.
- Tests that assert the superseded text are updated, not deleted (the seven of AC-SRH-007); the launcher-refusal tests in `internal/cli` that name `runs --retire` are out of scope and untouched.
- The probe script is kept in step with `spec.md` §D.7/§D.8: any change to a protocol prefix or a command line changes the probe in the same commit.

## D.6 Definition of Done

1. All Release-blocking ACs PASS with verbatim output, each Go-test AC with its swept count; every probe scenario PASSES (`control-healthy` still PASSES). At plan time only `control-healthy` has reached its PASS branch; the PASS branches of the other seven scenarios have never executed (no implementation exists), so a probe defect that makes a scenario unpassable is possible — the run phase observes each PASS, and a scenario that cannot pass under an implementation that meets `spec.md` §D.7/§D.8 is a blocker report against the probe, not a reason to weaken the SPEC.
2. The RED-now cells E1-E10 are the pre-change observations; for each High AC the RED is observed at its milestone's RED step, or the AC stays a regression-guard with the observation recorded.
3. Every mutant probe of D.2 executed and its failure observed.
4. The seven SPEC-STALE-RUN-LABEL-001 tests of AC-SRH-007 are green with literal assertions updated and no cadence assertion weakened (§D.9 of `spec.md` is the supersession record).
5. No kanban-only file changes (AC-SRH-016); no change under `internal/template/templates/`.
6. Decision points DP1-DP12 either confirmed or returned as a blocker report; none silently altered.
