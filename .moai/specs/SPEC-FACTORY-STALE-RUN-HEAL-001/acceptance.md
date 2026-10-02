# SPEC-FACTORY-STALE-RUN-HEAL-001 — Acceptance Criteria

Every AC is mechanically verifiable without a live factory run. The binary-level ACs run a build of the card tree invoked by path (never the installed `moai`); the rest are unit or integration tests on a `t.TempDir()` factory database. Go tests run env-scrubbed in ONE compound invocation, because a lane-stamped environment falsifies env-reading guard tests locally:

```bash
unset MOAI_FACTORY_WORKERS MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_NAME MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test <packages> -run '^(<TestName1>|<TestName2>)$' -count=1 -v
```

**Swept-count rule (verification-completeness §1.1, §1.3).** A Go-test AC passes only when its `-v` output shows one `--- PASS` line per named test; a `[no tests to run]` line, a PASS count below the named count, or a missing name is a FAIL, never a pass. The `-run` pattern is anchored at every alternation branch (`^(A|B)$`) so it cannot also select longer names. Each AC names the packages it touches (column "Packages"); re-measurement is scoped to those packages, not the full suite.

**Load attribution.** The hook gate paths measure run state inside a 200 ms budget (`factoryGateBudget`), so these tests appear load-sensitive (E5: different tests failed on different full-set runs and passed in isolation; the cause is inferred, not observed). Record the `uptime` load average beside every Go-test result; a failure at load well above the machine's quiet baseline is re-run in isolation, and the failing and passing outputs are both reported, never one of them.

**Two-cell discipline (verification-completeness §2).** Each release-blocking AC carries a RED-now cell and a green-path cell naming the milestone that flips it. The binary-level and grep-level RED cells are measured and pinned in the evidence ledger below. The Go-test RED cells are acquired at the owning milestone's RED step — the tests do not exist on the plan-phase tree, and writing them is outside the plan phase. An AC whose RED is not observed at its milestone loses release-blocking eligibility, is reclassified as a regression-guard, and is not recorded as a pass.

## D. AC Matrix

| ID | Requirement | Scenario (Given / When / Then) | Packages | Severity | Test and owning milestone |
|---|---|---|---|---|---|
| AC-SRH-001 | REQ-SRH-012 | **Given** a build of the card tree and any working directory **When** `moai factory relaunch --dry-run --lane lane-3 --provider cc` runs, and likewise for `--provider glm`, `--provider codex`, with `--run runA`, and with `--lane` omitted **Then** stdout is exactly `moai cc -f lane-3` (resp. `moai glm -f lane-3`, `moai codex -f lane-3`, `moai cc -f lane-3 --factory-run runA`, `moai cc -f lane`), stderr is empty, exit 0 | internal/cli | Release-blocking | Binary command + `TestFactoryRelaunchDryRunMatrix` (M1). RED: E1 |
| AC-SRH-002 | REQ-SRH-012 | **Given** the same build **When** `moai factory relaunch --dry-run --lane worker-3 --provider cc` runs **Then** exit is non-zero, stdout is empty, stderr names the canonical `lane-3` and the word `legacy`, and in the same test the invocation with `--lane lane-3` exits 0 (positive control) | internal/cli | Release-blocking | `TestFactoryRelaunchRefusesLegacyLane` (M1). RED: E1 (red for the stated reason — the verb is absent; note a bare non-zero exit would be vacuously green today, so the stderr content is the verdict) |
| AC-SRH-003 | REQ-SRH-013 | **Given** a factory database with run `runX` recorded active, and an owner classifier injected as dead / live / indeterminate, and a run `runR` already retired **When** the verb runs with `--from-run runX` (resp. `runR`) and the launch seam stubbed **Then** dead: `runX` reads `retired`, exactly one `run.retired` event carries the classification and basis, the seam receives the launch argv; live and indeterminate: `runX` still reads `active`, no event is added, the seam still receives the launch argv, and the output names the outcome; `runR`: no event is added | internal/cli, internal/homestate | Release-blocking | `TestFactoryRelaunchFromRunRetiresDeadOwnerOnly` (M1) |
| AC-SRH-004 | REQ-SRH-014 | **Given** a factory database with runs, peers, and cards populated **When** the verb runs under `--dry-run`, and again without `--dry-run` and without `--from-run` (launch seam stubbed, discovery seam returning no leader) **Then** the row counts of `runs`, `events`, `peers`, and `cards` are identical before and after both runs, the lane clear-policy environment value is untouched, and `moai factory relaunch --help` stdout contains the text `--clear-policy relaunch` | internal/cli, internal/homestate | Release-blocking | `TestFactoryRelaunchDoesNotMutateRunRecords` + `TestFactoryRelaunchHelpDistinguishesClearPolicy` (M1) |
| AC-SRH-005 | REQ-SRH-001, REQ-SRH-002 | **Given** a `t.TempDir()` factory database and the session cases {legacy label with the named run active; legacy label with the named run retired and another run active; current-vocabulary `lane-3` with the named run retired and exactly one other active; the same with two others active}, in each locale en, ko, ja, zh **When** the hook surfaces (bootstrap, peer registration) render the notice **Then** every line whose trimmed text begins with `moai factory relaunch ` contains no `<` and no `>`, the notice contains no `runs --retire`-then-`relaunch` prose and no `lane-<n>`, the lane label is pinned only for the current-vocabulary case, the provider token matches the session's backend (`claude`→`cc`, `glm`→`glm`, `gpt`→`codex`), and the command line is byte-identical across the four locales | internal/hook, internal/kanban | Release-blocking | `TestStaleNoticeCarriesExecutableRelaunch` (M1). RED: E2, E3 |
| AC-SRH-006 | REQ-SRH-002 | **Given** the shared command builder and the verb's parser **When** for every provider × {pinned lane, omitted lane} × {with, without `--from-run`} × {with, without `--run`} the built line is split into argv and fed to the verb with `--dry-run` appended **Then** every one exits 0 and prints the launch line the builder's arguments imply; a builder output the verb rejects fails the test | internal/cli, internal/kanban | Release-blocking | `TestRelaunchCommandRoundTrip` (M1) |
| AC-SRH-007 | REQ-SRH-003 | **Given** the five cadence tests of SPEC-STALE-RUN-LABEL-001 — `TestStaleRunNoticeSilentWhenRunRetired`, `TestStaleRunNoticeFiresWhenRunActive`, `TestStaleRunNoticeOncePerSession`, `TestUnbindNoticeThenSilence`, `TestUnbindNoticeRebindLinePresence` — with their literal-text assertions updated to the new command line and no cadence assertion weakened **When** they run **Then** all five pass | internal/hook | Release-blocking | Existing five tests, updated (M1). Baseline: E5 (green in one measurement and in an isolated re-run on the pre-change tree, flaky in two full-set runs at load averages ~57-63 — this AC is a regression-guard on cadence, not a RED→GREEN flip) |
| AC-SRH-008 | REQ-SRH-004 | **Given** run `X` recorded retired, exactly one other run `Y` active, an environment naming `X` with label `lane-3`, and a UserPromptSubmit event for session `S` **When** the peer registration path runs **Then** the broker of `Y` holds a peer for slot `lane-3` with session `S`, the returned notice names `X`, `Y`, `lane-3`, and the generation, the broker of `X` was never created or opened, and a second prompt of the same session returns silence | internal/hook, internal/factorymsg | Release-blocking | `TestLaneRebindsIntoSoleActiveRun` (M2) |
| AC-SRH-009 | REQ-SRH-005 | **Given** the AC-SRH-008 setup and a SessionStart event (source `clear` and source `startup`) **When** the session-start registration and bootstrap paths run **Then** no peer is registered in any run, no broker is created for `X` or `Y`, and no unbind notice is emitted | internal/hook | Release-blocking | `TestRebindEligibleSessionStartBindsNothing` (M2) |
| AC-SRH-010 | REQ-SRH-006 | **Given** a current-vocabulary `lane-3` session whose run `X` is retired, once with zero active runs and once with two **When** three consecutive prompts run **Then** zero active: exactly one unbind notice across the three prompts, no `factory messaging degraded` string, no peer registered; two active: no peer registered in either run, exactly one ambiguity notice listing both candidates each with its own `moai factory relaunch ... --run <id>` line (resolved ids), silence on the other two prompts, and no candidate selected | internal/hook | Release-blocking | `TestRebindIneligibleZeroOrManyActiveRuns` (M2) |
| AC-SRH-011 | REQ-SRH-007 | **Given** the AC-SRH-008 setup with the label `worker-69` (legacy), and separately the label `lane-3` with a live session already holding slot `lane-3` in `Y` **When** the prompt path runs **Then** legacy: no peer registered in `Y`, the notice's command line omits `--lane`; live owner: the existing peer row in `Y` is byte-identical afterwards, no write occurred, and one refusal notice with the command line is returned | internal/hook, internal/factorymsg | Release-blocking | `TestLegacyLabelNeverRebindsAndLiveOwnerNotDisplaced` (M2) |
| AC-SRH-012 | REQ-SRH-008 | **Given** a rebound session and one pending message in each of `X`'s and `Y`'s brokers addressed to its slot **When** the inbox claim runs **Then** both messages surface; and for the SPEC-STALE-RUN-LABEL-001 worker-70 case (live peer in the environment run's broker, no rebind) the claim output is unchanged (`TestInboundClaimIndependentOfEnvLabel` stays green) | internal/hook | Release-blocking | `TestReboundClaimReadsBothBrokers` + existing `TestInboundClaimIndependentOfEnvLabel` (M2) |
| AC-SRH-013 | REQ-SRH-009, REQ-SRH-010 | **Given** a current-vocabulary `lane-3` session whose run is active **When** the prompt and claim paths run with a counting seam on the run-state accessors **Then** the output equals the pre-change fixture byte-for-byte and the active-run listing is never invoked; **Given** a corrupt factory database or a spent inspection deadline **When** the rebind path runs **Then** the answer is the existing `factory messaging degraded:` string or silence, no error escapes the hook; and **in every case** the row counts of `runs` and `events` are unchanged by the hook | internal/hook, internal/factorymsg | Release-blocking | `TestHealthyLanePathUnchangedAndFailOpen` + existing `TestCurrentVocabularyBindPathUnchanged` (M2) |
| AC-SRH-014 | REQ-SRH-011 | **Given** the relaunch loop with launch, discovery, and lease seams, run `X` active at iteration 1 **When** `X` is retired and `Y` activated between iterations, and separately when no run is active at iteration 2 **Then** iteration 2's child receives `MOAI_KANBAN_ID=Y` and the lease is taken from `Y`'s record; with no active run the loop stops with the gate's `NO_ACTIVE_FACTORY` text and the lease seam is never called; with an explicit `--factory-run X` that retires, the loop stops (explicit stays explicit) | internal/cli | Release-blocking | `TestRelaunchLoopReResolvesRun` (M3). RED: E4 (observation) and the test's own RED at M3 |
| AC-SRH-015 | REQ-SRH-015 | **Given** the card's implementation diff **When** `git diff --name-only $(git merge-base develop HEAD)..HEAD -- internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go` runs (merge-base re-derived at read time, pre-merge only — `gitflow-lane-protocol.md` §8) **Then** stdout is empty (positive control: the same range with no pathspec lists at least one file); and `TestKanbanRelaunchProseUnchanged` asserts the four locales' `roleValueRelaunch` strings equal their pre-change golden bytes | internal/hook | High | Diff command + `TestKanbanRelaunchProseUnchanged` (M4) |

## D.1 Evidence Ledger (RED-now cells, measured this run)

All entries measured at tree `802a72235536958ada5b7cd5876a168e4b8c325f` (card branch `WT-stale-run-healing`, clean). The binary of E1 was built from this tree by `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=802a72235" -o <scratch>/moai-t1345 ./cmd/moai` and invoked by path; its `version` prints `v3.1.3   802a72235   built unknown`. The scratch path is machine-local and is not a citation target; the command and output are.

```text
E1  command : moai factory relaunch --dry-run --lane lane-3 --provider cc
    stdout  : (empty)
    stderr  : ERROR / Unknown flag: --dry-run. / Try --help for usage.   (whitespace-collapsed)
    exit    : 1
    why red : the verb does not exist; the factory parent command consumes the unknown
              argument "relaunch" and rejects the flag. (Note: `moai factory relaunch --help`
              exits 0 today and prints the factory parent's help — a vacuous green, which is
              why the AC verdict is the dry-run stdout and the help-text content, not an exit code.)

E2  command : grep -c "lane-<n>" internal/hook/session_stale_run.go
    stdout  : 4
    exit    : 0
    why red : the unbind re-bind line prints the placeholder in all four locales
              (lines 92, 103, 114, 125). Green: stdout 0, exit 1.

E3  command : grep -c "factory relaunch" internal/hook/session_stale_run.go
    stdout  : 0
    exit    : 1
    why red : no notice names the verb. Green: stdout >= 4 (the four locales), exit 0.

E4  command : grep -c "enterFactoryLaneRun(" internal/cli/factory_lane_relaunch.go
    stdout  : 0
    exit    : 1
    control : grep -c "enterFactoryLaneRun(" internal/cli/cc.go -> stdout 1, exit 0
    why red : the relaunch loop never calls the shared join gate; it reads the run id from the
              environment once (factory_lane_relaunch.go:56). Green: stdout >= 1, exit 0.

E5  command : unset <factory/kanban vars> && go test ./internal/hook -run '^(TestRoleNamingM3StaleRunNoticeNamesRetireStep|TestStaleRunNoticeSilentWhenRunRetired|TestStaleRunNoticeFiresWhenRunActive|TestStaleRunNoticeOncePerSession|TestUnbindNoticeThenSilence|TestUnbindNoticeRebindLinePresence|TestPrescriptionGateUnavailableFailsOpen|TestInboundClaimIndependentOfEnvLabel|TestCurrentVocabularyBindPathUnchanged|TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord|TestStaleRunNoticeFactoryLegacyLabel)$' -count=1 -v
    observed: three measurements, all on this tree, none of them a single clean verdict:
              (1) substring pattern `StaleRun|Unbind|Prescription|CurrentVocabulary|InboundClaim`
                  (selects the same 12 names): 12 `--- PASS`,
                  `ok  github.com/modu-ai/moai-adk/internal/hook  7.242s`, exit 0;
              (2) the anchored pattern above: 3 FAIL (TestStaleRunNoticeOncePerSession,
                  TestUnbindNoticeThenSilence, TestUnbindNoticeRebindLinePresence), 9 PASS, 15.083s, exit 1;
              (3) those three failing tests alone: 3 PASS, `ok  ...  5.554s`, exit 0;
                  `uptime` read right after: load averages 57.76 60.51 48.48;
              (4) the anchored pattern again: 2 FAIL (TestStaleRunNoticeOncePerSession,
                  TestUnbindNoticeRebindLinePresence), 10 PASS, 16.317s, exit 1;
                  `uptime` read right after: load averages 63.26 61.58 49.35.
              No load reading exists for (1) and (2).
    why     : the failing set differs run to run and passes in isolation — consistent with the
              gate's 200 ms measurement budget (`factoryGateBudget`) being exhausted by machine
              load. This is inferred from that pattern (different tests each run, green in
              isolation, load averages ~57-63 right after) — no failure message was captured,
              so the cause is not observed and no pre-existing defect is claimed.
    role    : baseline for AC-SRH-007 / AC-SRH-012 / AC-SRH-013 (characterization; the literal-text
              assertions in five of these tests change at M1, the cadence and independence
              assertions do not). A clean verdict needs a quiet machine or an isolated re-run.
```

AC-SRH-008..011 and AC-SRH-014's behavioural RED is observed at M2 / M3 RED: today a current-vocabulary lane with a retired run returns `factory messaging degraded: NO_ACTIVE_FACTORY` (`factory_messages.go:90-92`) and registers nothing, which is the stated reason each test is red before the work.

## D.2 Mutant Probes (verification-completeness §2)

Before an AC is adopted, a mutant that satisfies it while violating its requirement is sought. Each row names the mutant the AC must kill; the run phase executes the mutant at the owning milestone and records the observed failure.

| AC | Mutant (satisfies a shallow reading, violates the requirement) | Killed by |
|---|---|---|
| AC-SRH-001 | prints `-f lane` always, ignoring `--lane` | the `lane-3` row's exact-stdout assertion |
| AC-SRH-002 | accepts `worker-3` and maps it to a number | the stderr-content assertion plus the same-test positive control |
| AC-SRH-003 | retires whenever `--from-run` is given | the live and indeterminate rows (`runX` still `active`) |
| AC-SRH-003 | never retires | the dead row (`runX` reads `retired`, one `run.retired` event) |
| AC-SRH-004 | `--dry-run` also performs the retire pre-step | the before/after row-count equality |
| AC-SRH-005 | one locale keeps `lane-<n>` | the per-locale scan |
| AC-SRH-006 | the builder emits a flag the verb does not define | the round-trip exit-0 assertion |
| AC-SRH-008 | registers into the retired run `X` instead of `Y` | the peer-in-`Y` and no-broker-for-`X` assertions |
| AC-SRH-009 | binds at SessionStart | the no-broker-created assertion |
| AC-SRH-010 | picks the first of two active runs | the no-peer-in-either-run assertion |
| AC-SRH-011 | rebinds a legacy label, or overwrites a live slot owner | the no-peer-in-`Y` and byte-identical-peer-row assertions |
| AC-SRH-012 | reads only `Y` for a rebound session | the `X`-message-surfaces assertion |
| AC-SRH-013 | lists active runs on the healthy path | the counting-seam assertion |
| AC-SRH-014 | resolves the run once, before the loop | the iteration-2 `MOAI_KANBAN_ID=Y` assertion |
| AC-SRH-015 | edits `roleValueRelaunch` | the golden-bytes test |

## D.3 Severity Scale

Release-blocking: the defect class this SPEC exists to kill — a lane locked out after a run switch, a non-executable notice, an absent or lossy verb, a loop that leases from a dead run. High: a scope guard.

## D.4 Edge Cases

- `--lane lane` (the bare role token) is not a lane label: the verb treats it as invalid; `-f lane` is selected by omitting `--lane`.
- `--run` naming a non-active run: the launcher gate answers `NO_ACTIVE_FACTORY` (an explicit selection is a decision, not an absence); the verb surfaces that refusal and launches nothing.
- `--from-run` naming an unknown run id: treated as "run not active" (no retirement, outcome named), never an error that blocks the launch.
- A session whose label is current-vocabulary but whose environment names no run (`MOAI_KANBAN_ID` empty): not a stale session; unchanged (the hook already returns silence).
- `/clear` after a rebind: the session identity changes, the process does not; the first prompt of the new identity registers it in `Y` again (a generation bump) — the one rebound notice is correct again for a new identity.
- Two active runs where one has a dead owner: the hook does not classify (D3); the ambiguity notice names both, and the verb's `--run` or `--from-run` resolves it.
- The ambiguity notice with more than three active runs: three command lines, a count, and the `moai factory runs` pointer (D10).
- Hook context budget: the rebound and ambiguity notices stay within the existing notice size; no truncation of the command line (a truncated command is a non-executable command).

## D.5 Quality Gates

- `go vet` and `golangci-lint run` (CI pin v2.1.6) clean on `internal/hook`, `internal/factorymsg`, `internal/cli`, `internal/kanban`.
- `GOOS=windows GOARCH=amd64 go build ./...` passes (the verb re-executes a launcher; no POSIX-only primitive outside the existing `_posix.go` / `_windows.go` split).
- Package coverage per `quality.yaml` `test_coverage_target` on the changed packages.
- Separation diff review: no environment-label reference is added inside the existing claim path's call tree; the additive read is a separate code path (REQ-SRH-008).
- Hardcoding prevention: no new `MOAI_` environment-name string literal outside `internal/config/envkeys.go` and tests (reviewed by a diff-scoped grep at sync).
- i18n parity: the four locales of every new notice string exist and carry identical protocol tokens.
- `moai spec lint SPEC-FACTORY-STALE-RUN-HEAL-001 --strict` exits 0 on a binary built from the tree.
- Tests that assert the old notice text are updated, not deleted: `stale_run_gate_test.go` (`runs --retire` assertions, the `moai cc -f lane-` re-bind assertion), `stale_run_m1_test.go`, `role_naming_m3_notice_test.go`. The launcher-refusal tests in `internal/cli` that name `runs --retire` are out of scope and untouched.

## D.6 Definition of Done

1. All release-blocking ACs PASS with verbatim command output recorded, each Go-test AC with its swept count (one `--- PASS` per named test).
2. RED-now cells observed and pinned for every release-blocking AC (E1-E4 now; the Go-test cells at M1/M2/M3 RED), or the AC reclassified as a regression-guard with the reason recorded.
3. Every mutant probe of D.2 executed and its failure observed.
4. The SPEC-STALE-RUN-LABEL-001 cadence and independence tests are green with their literal assertions updated and no cadence assertion weakened (AC-SRH-007, AC-SRH-012).
5. No kanban-only file changes (AC-SRH-015); no change under `internal/template/templates/`.
6. Mid-run decisions D1-D10 either confirmed or returned as a blocker report; none silently altered.
