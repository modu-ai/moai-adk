# SPEC-LAUNCHER-ENTRY-FLAGS-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02
spec_version: "0.8.0"   # plan-audit iteration 2 (FAIL 0.79) findings D29-D45 addressed on top of v0.7.0 (D1-D28); Q20-Q22 and Q24 open and non-gating; Tier L
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, decision-index.md]
operator_verdicts_recorded: 20  # decision-index Q1-Q6, Q8-Q15, Q17-Q19, Q23 (Q1 superseded by Q10; Q9 count sentence superseded by Q17); Q7 closed as moot; Q23 (role-declaration carrier) answered after plan-audit iteration 2
orchestrator_rulings_recorded: 1   # Q16 (OD-15): raised as a ruling, CONFIRMED by the operator
open_questions: 4               # Q20-Q22 raised by the plan audit and Q24 raised by the compile proof, all non-gating (smallest-footprint reading written); Q18 (OD-17) settled earlier as Option X
author_choices_listed: 12       # spec.md §D, decision-index Q19 — all accepted (two steps)
requirements: 25                # Tier L ceiling 25 (at the ceiling; REQ-019 split into REQ-019 and REQ-025 in v0.8.0)
criteria: 25                    # Tier L ceiling 25 (at the ceiling)
milestones: 13                  # M0-M11 with M5 split into M5a/M5b; integration units: M0-M1, M2+M3+M4, then M5a, M5b, M6, M7, M8, M9, M10, M11 each alone
tree_measured: a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2
audited_head: 5445e296caa48e6ab9821afe808eac2e7385e897   # plan-audit iteration 2; content outside this SPEC directory equals tree_measured (PV-73); iteration 1 audited b9242da00ce489c4f26efb5f6d447ccafef08c54 (PV-71)
branch: WT-launcher-entry-flags
```

Plan-phase signal: plan-audit iteration 1 returned FAIL (0.72, audited at `b9242da00ce489c4f26efb5f6d447ccafef08c54`; report `.moai/reports/t1399/plan-audit-iter1.md`) and iteration 2 returned FAIL (0.79, audited at `5445e296caa48e6ab9821afe808eac2e7385e897`; report `.moai/reports/t1399/plan-audit-iter2.md`; both local and gitignored); v0.8.0 addresses findings D29-D45 and replaces the caller-grep deletion-order table with a committed cumulative compile proof (PV-73 to PV-90); iteration 3 is the last audit allowed and has not run. Every operator verdict is recorded and confirmed (Q23, the role-declaration carrier, was added after iteration 2); Q20-Q22 (raised by iteration 1) and Q24 (raised by the compile proof) are open and non-gating.

## §E.2 Run-phase Evidence

_<pending run-phase>_

### M1 evidence

Recorded by the run-phase implementation worker (cycle_type tdd, tests only) for milestone M1 of card t1399. The earlier line `_<pending run-phase>_` above is superseded by this subsection and left as written. Every output below was produced in this run; the saved outputs live in the worker scratchpad and are quoted here without editing (long lines are not shortened).

#### Claim

- M1 added tests only: no production file changed (diff against the start HEAD, below). The factory net, the enterable-pair matrix, the AC-009 characterization, the AC-016 frozen-value tests and the AC-017 tolerance tests are green on the unmodified production tree, and each of the eight AC-015 mutants (a)-(h) was observed red against the net before the tests were committed. The AC-004 golden was committed alone, ahead of every other change.
- Commits (explicit-path staging, English subjects, trailer `Authored-By-Agent: manager-develop`): `0472b060c` (the AC-004 golden alone, plus the spec.md frontmatter `status: draft` to `status: in-progress`); `1379abc9a` (the nine test files); the commit carrying this subsection.

#### Baseline-attribution

- Start state, re-read at the start of the run: `git rev-parse --short HEAD` printed `29bf84c33`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. HEAD tree `6e67e965586b440b6eec145c2fef20a8c0325d72`. `git diff --stat a6d3e6fd4 29bf84c33 -- internal cmd pkg` printed nothing (exit 0), so the Go sources equal tree `a6d3e6fd4`, the tree every SPEC measurement is pinned to.
- Golden capture tree: HEAD `29bf84c33` (tree `6e67e965586b440b6eec145c2fef20a8c0325d72`), production sources unmodified, the nine test files untracked in the working tree. Mutant runs: HEAD `0472b060c` (tree `012cbe5dc4d0a8ca03f98db34bc2a5d1f9ab2cb6`) plus the uncommitted test files that were committed unchanged as `1379abc9a`. Green and AC results below: HEAD `1379abc9a` (tree `c65b5971fa23a4f900b68de9949436229e34a48f`), measured in this run.
- The judging build is the Go toolchain `go1.26.8 darwin/arm64` compiling this tree (every command is `go test`, `go vet` or `go build` on the tree; no installed `moai` binary was used as a measuring instrument).
- Env hygiene: every test run was one compound invocation `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND && go test ...`; the new tests additionally scrub every `MOAI_FACTORY*` and `MOAI_KANBAN*` variable by prefix inside the test process.

#### The net and its swept counts (AC-015)

Selector, internal/cli (nine names; exactly one of the two settings-test names exists today, so it sweeps 8): `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestPrepareKanbanSettingsWritesTransientFile|TestPrepareFactorySettingsWritesTransientFile|TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch|TestFactoryNetBlockCap|TestFactoryEntryMatrix)$' -v -count=1` at HEAD `1379abc9a`, exit 0, swept 8 (eight `--- PASS` lines at top level):

```text
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (3.12s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (4.40s)
--- PASS: TestFactoryNetLeaderLaunch (2.10s)
    --- PASS: TestFactoryNetLeaderLaunch/cc (0.74s)
    --- PASS: TestFactoryNetLeaderLaunch/glm (1.35s)
--- PASS: TestFactoryNetLaneLaunch (5.49s)
    --- PASS: TestFactoryNetLaneLaunch/cc (3.10s)
    --- PASS: TestFactoryNetLaneLaunch/glm (2.39s)
--- PASS: TestFactoryNetBlockCap (1.67s)
    --- PASS: TestFactoryNetBlockCap/leader (0.71s)
    --- PASS: TestFactoryNetBlockCap/lane (0.97s)
--- PASS: TestFactoryEntryMatrix (5.79s)
    --- PASS: TestFactoryEntryMatrix/claude_leader (0.45s)
    --- PASS: TestFactoryEntryMatrix/glm_leader (0.54s)
    --- PASS: TestFactoryEntryMatrix/claude_lane (0.95s)
    --- PASS: TestFactoryEntryMatrix/glm_lane (0.79s)
    --- PASS: TestFactoryEntryMatrix/codex_lane (2.68s)
    --- PASS: TestFactoryEntryMatrix/codex_leader_refused (0.38s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (0.46s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	24.021s
```

Selector, internal/hook: `go test ./internal/hook -run '^(TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices)$' -v -count=1`, exit 0, swept 2:

```text
--- PASS: TestFactoryNetSessionRecord (0.01s)
    --- PASS: TestFactoryNetSessionRecord/lane (0.00s)
    --- PASS: TestFactoryNetSessionRecord/leader_signalled_by_the_factory_variable_alone (0.00s)
    --- PASS: TestFactoryNetSessionRecord/ordinary_session_writes_nothing (0.00s)
--- PASS: TestFactoryNetSessionStartNotices (1.57s)
    --- PASS: TestFactoryNetSessionStartNotices/leader (1.04s)
    --- PASS: TestFactoryNetSessionStartNotices/lane (0.29s)
    --- PASS: TestFactoryNetSessionStartNotices/ordinary_session_gets_no_factory_notice (0.24s)
ok  	github.com/modu-ai/moai-adk/internal/hook	2.257s
```

Selector, internal/discovery: `go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1`, exit 0, swept 2:

```text
--- PASS: TestDiscoverLeaderVerifiesLiveLeader (0.14s)
--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID (0.09s)
ok  	github.com/modu-ai/moai-adk/internal/discovery	0.560s
```

The net's lane launches and the matrix use today's `-f lane`; M2 (cc, glm) and M3 (codex) re-pin them to `-l`. New test files: `internal/cli/factory_net_m1_test.go` (leader launch, lane launch, block cap, matrix), `internal/hook/factory_net_m1_test.go` (session record, SessionStart notices, hook tolerance test).

#### The eight mutants (AC-015) and the AC-017 mutants: verbatim reds

Each mutant is a scratch edit to ONE production file, applied by a script that asserted the exact original text, built (`go build` of the package, exit 0), run against the net selector of the package it targets, and REVERTED immediately by the inverse script; after each revert `git diff --stat -- <file>` printed nothing. Exit code of every red run below: 1. The reds quote the saved output filtered to failing-test names, the assertion lines (`_test.go:N:`) and the package verdict.

**(a) settings injection removed** - edit: `internal/cli/cc.go`: the three `settingsFlag, settingsCleanup := prepareKanbanSettings(profileName, filteredArgs)` calls in the factory leader and factory lane branches replaced by `var settingsFlag []string; settingsCleanup := func() {}`. Run: the internal/cli selector above, exit 1.

```text
    factory_net_m1_test.go:287: the launch argv carries no --settings pair: [--name leader]
    factory_net_m1_test.go:287: the injected settings payload = map[], want crossSessionInbound=accept
    factory_net_m1_test.go:287: MOAI_KANBAN_SETTINGS_INJECTED at launch = "", want 1 (the SessionStart hook reads it)
--- FAIL: TestFactoryNetLeaderLaunch (3.16s)
    --- FAIL: TestFactoryNetLeaderLaunch/cc (1.54s)
    factory_net_m1_test.go:338: the launch argv carries no --settings pair: [--name lane-1]
    factory_net_m1_test.go:338: the injected settings payload = map[], want crossSessionInbound=accept
    factory_net_m1_test.go:338: MOAI_KANBAN_SETTINGS_INJECTED at launch = "", want 1 (the SessionStart hook reads it)
--- FAIL: TestFactoryNetLaneLaunch (5.64s)
    --- FAIL: TestFactoryNetLaneLaunch/cc (3.88s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	33.374s
FAIL
```

**(b) block-cap factory clause removed** - edit: `internal/cli/launcher_blockcap_infinite.go`: the `|| os.Getenv(config.EnvMoaiFactoryWorkers) != ""` clause removed from the unconditional-raise condition. Run: the internal/cli selector above, exit 1.

```text
    factory_net_m1_test.go:398: leader: injectStopHookBlockCapForGoal = [PATH=/usr/bin HOME=/tmp], want "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200" (the factory clause)
    factory_net_m1_test.go:405: lane: injectStopHookBlockCapForGoal = [PATH=/usr/bin HOME=/tmp], want "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200" (the factory clause)
--- FAIL: TestFactoryNetBlockCap (1.36s)
    --- FAIL: TestFactoryNetBlockCap/leader (0.51s)
    --- FAIL: TestFactoryNetBlockCap/lane (0.85s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	25.451s
FAIL
```

**(c) a factory marker publish removed** - edit: `internal/cli/factory.go`: `_ = os.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)` removed from `enterFactoryLaneMode`. Run: the internal/cli selector above, exit 1.

```text
    factory_net_m1_test.go:332: MOAI_FACTORY_ROLE at launch = "", want "lane"
    factory_net_m1_test.go:332: MOAI_FACTORY_ROLE at launch = "", want "lane"
--- FAIL: TestFactoryNetLaneLaunch (2.81s)
    --- FAIL: TestFactoryNetLaneLaunch/cc (1.41s)
    --- FAIL: TestFactoryNetLaneLaunch/glm (1.40s)
    factory_net_m1_test.go:462: claude lane did not launch as a lane: role="" worker="lane-1"
    factory_net_m1_test.go:462: glm lane did not launch as a lane: role="" worker="lane-1"
    factory_net_m1_test.go:157: stage t1 run: factory stage: refused — not a lane session: set MOAI_FACTORY_ROLE=lane in a lane session (the launcher stamps it)
--- FAIL: TestFactoryEntryMatrix (6.96s)
    --- FAIL: TestFactoryEntryMatrix/claude_lane (1.07s)
    --- FAIL: TestFactoryEntryMatrix/glm_lane (1.11s)
    --- FAIL: TestFactoryEntryMatrix/codex_lane (2.21s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	27.343s
FAIL
```

**(d) discovery read removed** - edit: `internal/discovery/factory_discovery.go`: `runID := strings.TrimSpace(env[config.EnvMoaiKanbanID])` replaced by a read of a key that does not exist (`env["MUTANT_D_NO_SUCH_KEY"]`, with `_ = config.EnvMoaiKanbanID` keeping the import used; the first form, `runID := ""`, failed to BUILD on an unused import and is not counted as a red). Run: the internal/discovery selector above, exit 1.

```text
    factory_discovery_test.go:106: verified = 0, want 1 ([])
--- FAIL: TestDiscoverLeaderVerifiesLiveLeader (0.15s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/discovery	0.655s
FAIL
```

**(e) session-record role reader returns no role for a factory leader or lane** - edit: `internal/hook/session_start_record.go`: the factory-lane branch and the factory-leader branch of `kanbanRoleFromEnv` guarded with `&& false`. Run: the internal/hook selector above, exit 1.

```text
    factory_net_m1_test.go:68: the lane wrote no session record: read kanban record: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestFactoryNetSessionRecordlane690889000/001/.moai/state/todo/net-lane-sess.json: no such file or directory
    factory_net_m1_test.go:91: the factory leader wrote no session record: read kanban record: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestFactoryNetSessionRecordleader_signalled_by_the_factory_vari1855199715/001/.moai/state/todo/net-leader-sess.json: no such file or directory
--- FAIL: TestFactoryNetSessionRecord (0.00s)
    --- FAIL: TestFactoryNetSessionRecord/lane (0.00s)
    --- FAIL: TestFactoryNetSessionRecord/leader_signalled_by_the_factory_variable_alone (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.595s
FAIL
```

**(f) factory SessionStart notice block removed** - edit: `internal/hook/session_start.go`: the call `factoryBootstrapNoticeForSource(input.Source, factoryRoot, input.SessionID, langEnglish)` replaced by `""` (`if notice := ""; notice != "" {`). Run: the internal/hook selector above, exit 1.

```text
    factory_net_m1_test.go:153: leader notice on additionalContext lacks "netrun01":
    factory_net_m1_test.go:153: leader notice on additionalContext lacks "/tmp/moai-socket-factory/netrun01":
    factory_net_m1_test.go:153: leader notice on systemMessage lacks "netrun01":
    factory_net_m1_test.go:153: leader notice on systemMessage lacks "/tmp/moai-socket-factory/netrun01":
    factory_net_m1_test.go:157: leader notice on systemMessage never names the leader role:
    factory_net_m1_test.go:173: lane notice on additionalContext does not name the lane label "lane-2":
    factory_net_m1_test.go:173: lane notice on systemMessage does not name the lane label "lane-2":
--- FAIL: TestFactoryNetSessionStartNotices (0.86s)
    --- FAIL: TestFactoryNetSessionStartNotices/leader (0.30s)
    --- FAIL: TestFactoryNetSessionStartNotices/lane (0.32s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.739s
FAIL
```

**(g) lane claim call removed (today's `-f lane` path)** - edit: `internal/cli/cc.go`: `finalLabel, claimErr := resolveFactoryLaneName(...)` replaced by `finalLabel, claimErr := factoryLabel, error(nil)`. Run: the internal/cli selector above, exit 1.

```text
    factory_net_m1_test.go:343: registry[lane-1] = ({PID:0 RegisteredAt:}, false), want a live claim under pid 8187
    factory_net_m1_test.go:352: second lane MOAI_FACTORY_WORKER = "lane-1", want "lane-2" (the first claim is live)
    factory_net_m1_test.go:356: registry lost the second claim: map[]
--- FAIL: TestFactoryNetLaneLaunch (2.88s)
    --- FAIL: TestFactoryNetLaneLaunch/cc (1.26s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	21.714s
FAIL
```

**(h) `MOAI_KANBAN_BACKEND` publication removed (today's `-f lane` and `-f` paths)** - edit: `internal/cli/kanban.go`: `_ = os.Setenv(config.EnvMoaiKanbanBackend, backend)` replaced by `_ = backend` in `exportKanbanLaunchFacts`. Run: the internal/cli selector above, exit 1.

```text
    factory_net_m1_test.go:269: MOAI_KANBAN_BACKEND at launch = "", want "claude"
    factory_net_m1_test.go:269: MOAI_KANBAN_BACKEND at launch = "", want "glm"
--- FAIL: TestFactoryNetLeaderLaunch (1.88s)
    --- FAIL: TestFactoryNetLeaderLaunch/cc (0.84s)
    --- FAIL: TestFactoryNetLeaderLaunch/glm (1.04s)
    factory_net_m1_test.go:332: MOAI_KANBAN_BACKEND at launch = "", want "claude"
    factory_net_m1_test.go:332: MOAI_KANBAN_BACKEND at launch = "", want "glm"
--- FAIL: TestFactoryNetLaneLaunch (5.03s)
    --- FAIL: TestFactoryNetLaneLaunch/cc (2.56s)
    --- FAIL: TestFactoryNetLaneLaunch/glm (2.47s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	23.660s
FAIL
```

**Bonus, glm path (a)+(g)** - edit: in `internal/cli/glm.go` the three factory-branch `prepareKanbanSettings` calls and the `resolveFactoryLaneName` call replaced exactly as in (a) and (g). Run: `go test ./internal/cli -run '^(TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch)$' -v -count=1`, exit 1 (the cc subtests stay PASS, the glm subtests go red, so the rows are independent):

```text
    factory_net_m1_test.go:287: the launch argv carries no --settings pair: [--name leader]
    factory_net_m1_test.go:287: the injected settings payload = map[], want crossSessionInbound=accept
    factory_net_m1_test.go:287: MOAI_KANBAN_SETTINGS_INJECTED at launch = "", want 1 (the SessionStart hook reads it)
--- FAIL: TestFactoryNetLeaderLaunch (2.28s)
    --- PASS: TestFactoryNetLeaderLaunch/cc (1.16s)
    --- FAIL: TestFactoryNetLeaderLaunch/glm (1.12s)
    factory_net_m1_test.go:338: the launch argv carries no --settings pair: [--name lane-1]
    factory_net_m1_test.go:338: the injected settings payload = map[], want crossSessionInbound=accept
    factory_net_m1_test.go:338: MOAI_KANBAN_SETTINGS_INJECTED at launch = "", want 1 (the SessionStart hook reads it)
    factory_net_m1_test.go:343: registry[lane-1] = ({PID:0 RegisteredAt:}, false), want a live claim under pid 30449
    factory_net_m1_test.go:352: second lane MOAI_FACTORY_WORKER = "lane-1", want "lane-2" (the first claim is live)
    factory_net_m1_test.go:356: registry lost the second claim: map[]
--- FAIL: TestFactoryNetLaneLaunch (3.69s)
    --- PASS: TestFactoryNetLaneLaunch/cc (2.13s)
    --- FAIL: TestFactoryNetLaneLaunch/glm (1.56s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	6.939s
FAIL
```

**AC-017 mutants, one per package.** Command (all four packages): `go test ./internal/cli ./internal/hook ./internal/web ./internal/statusline -run '^TestPreexistingKanbanArtifactsTolerated$' -v -count=1`.

- cli: `internal/cli/doctor_factory_run.go`, a `default:` case in the role switch of `checkFactoryRun` returning `uikit.CheckFail` for an unknown role (reader fails on an unknown role). Run: the same command, exit 1:

```text
    launcher_characterization_m1_test.go:191: doctor check "Factory Run" failed on pre-existing kanban artifacts: {Name:Factory Run Status:fail Message:unknown role Detail:}
--- FAIL: TestPreexistingKanbanArtifactsTolerated (0.12s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.110s
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.53s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.368s
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.57s)
ok  	github.com/modu-ai/moai-adk/internal/web	1.351s
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.432s
FAIL
```

- hook, web, statusline (applied together in one second run, each in its own file): hook `internal/hook/session_start_record.go` early return on an existing record removed (the pre-existing record is rewritten); web `internal/web/viewmodel_ops.go` `buildKanban` returns an error when `.moai/state/kanban-board` exists; statusline `internal/statusline/backlog.go` `resolveBacklogCounts` panics when `.moai/state/kanban-board` exists. The cli package stays PASS in this second run because the doctor mutant had already been reverted. Run: the same command, exit 1:

```text
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.16s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.768s
    factory_net_m1_test.go:236: the pre-existing record was rewritten:
--- FAIL: TestPreexistingKanbanArtifactsTolerated (0.61s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.612s
    preexisting_kanban_artifacts_m1_test.go:59: buildKanban failed on pre-existing kanban artifacts: invalid argument
--- FAIL: TestPreexistingKanbanArtifactsTolerated (0.17s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/web	1.082s
--- FAIL: TestPreexistingKanbanArtifactsTolerated (0.00s)
panic: unknown artifact [recovered, repanicked]
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.844s
FAIL
```

Revert proof: after the last revert, `git status --short` (filtered to tracked changes) printed nothing and `git diff --stat` printed nothing, with HEAD `0472b060c`, before the tests were committed.

#### The AC-004 golden

- Golden: `internal/cli/testdata/lane_entry_env_golden.json`, three rows (cc, glm, codex) of the `MOAI_FACTORY*` and `MOAI_KANBAN*` environment today's `-f lane` publishes. The codex row carries `MOAI_KANBAN_LABEL=lane-1` (the Codex lane child stamps it today); the cc and glm rows carry `MOAI_KANBAN_ID`, `MOAI_KANBAN_BACKEND`, `MOAI_KANBAN_SETTINGS_INJECTED`, `MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`.
- Capture command (run once with the capture mode of the test, output to the worker scratchpad, then copied into the tree): `unset <the eleven variables above> && MOAI_LANE_GOLDEN_WRITE=<scratchpad>/lane_entry_env_golden.json go test ./internal/cli -run '^TestLaneMarkerGoldenMatchesFLane$' -count=1 -v`, exit 0, final capture output below; tree HEAD `29bf84c33` (tree `6e67e965586b440b6eec145c2fef20a8c0325d72`). A first capture was discarded and not committed: it took its environment through an existing helper that sets the lane variables to empty strings, which wrote `MOAI_KANBAN`, `MOAI_KANBAN_CARD`, `MOAI_KANBAN_LABEL` and `MOAI_KANBAN_LEAD_ADDR` into the cc and glm rows as published empties; the final capture removes the variables by prefix (key absent, not empty) and is the committed one.
```text
    lane_entry_golden_m1_test.go:149: golden written to /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/0dcdf2d5-df5c-4da1-8870-24c2a5861303/scratchpad/lane_entry_env_golden.json
--- PASS: TestLaneMarkerGoldenMatchesFLane (6.94s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/cc (1.60s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/glm (1.67s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/codex (3.66s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	8.357s
```

- Commit order (the commit graph is the only ordering witness, verification-claim-integrity 2.3): `0472b060c` carries the golden and the frontmatter line only (`git show --stat` printed two files); the test that compares against it is in the later commit `1379abc9a`.
- M1 check, `go test ./internal/cli -run '^TestLaneMarkerGoldenMatchesFLane$' -v -count=1`, exit 0:

```text
--- PASS: TestLaneMarkerGoldenMatchesFLane (4.21s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/cc (0.82s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/glm (0.90s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/codex (2.49s)
ok  	github.com/modu-ai/moai-adk/internal/cli	4.888s
```

#### AC results

| AC | Result | Exit | Swept |
|----|--------|------|-------|
| AC-004 (M1 form) | PASS | 0 | 1 parent, 3 subtests |
| AC-009 | PASS | 0 | 5 |
| AC-015 | PASS (net green, eight mutants red) | 0 / 0 / 0 | 8 / 2 / 2 |
| AC-016 | PASS | 0 each | 1 / 2 / 1 / 1 |
| AC-017 | PASS | 0 | 1 in each of the four packages |

Commands, all this run at HEAD `1379abc9a`:

- AC-004 (golden against today's `-f lane`; the `-l` parity test is M2's): `go test ./internal/cli -run '^TestLaneMarkerGoldenMatchesFLane$' -v -count=1`
- AC-009: `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestLeaderEntryUnchanged|TestCGRetiredEntryAndModeHaveZeroEffects|TestCGRetirementCompleteEntryShapesAndCounters|TestBareMoaiPrintsBannerAndHelp)$' -v -count=1`
- AC-015: the three selectors in the net section above
- AC-016: `go test ./internal/config -run '^TestFactoryMarkerValuesFrozen$' -v -count=1`; the discovery selector of AC-015; `go test ./internal/codexwiring -run '^TestMCPServerEnvVarsKeepFactoryMarkers$' -v -count=1`; `go test ./internal/kanban -run '^TestLegacyStateDirStillRead$' -v -count=1` (the AC names ./internal/factory; the package is ./internal/kanban through M7)
- AC-017: `go test ./internal/cli ./internal/hook ./internal/web ./internal/statusline -run '^TestPreexistingKanbanArtifactsTolerated$' -v -count=1`

AC-009 output:

```text
--- PASS: TestCGRetiredEntryAndModeHaveZeroEffects (0.00s)
--- PASS: TestCGRetirementCompleteEntryShapesAndCounters (2.45s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (1.30s)
--- PASS: TestLeaderEntryUnchanged (3.68s)
    --- PASS: TestLeaderEntryUnchanged/cc_-f (1.11s)
    --- PASS: TestLeaderEntryUnchanged/cc_--factory (0.86s)
    --- PASS: TestLeaderEntryUnchanged/glm_-f (0.75s)
    --- PASS: TestLeaderEntryUnchanged/glm_--factory (0.68s)
    --- PASS: TestLeaderEntryUnchanged/codex_-f_has_no_leader_entry (0.28s)
    --- PASS: TestLeaderEntryUnchanged/gpt_is_an_unknown_command (0.00s)
--- PASS: TestBareMoaiPrintsBannerAndHelp (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	8.167s
```

AC-016 outputs:

```text
--- PASS: TestFactoryMarkerValuesFrozen (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.317s
--- PASS: TestMCPServerEnvVarsKeepFactoryMarkers (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.670s
--- PASS: TestLegacyStateDirStillRead (0.10s)
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.527s
```

AC-017 output:

```text
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.09s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.756s
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.32s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.730s
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.35s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.734s
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.194s
```

#### Hygiene and scope

- `gofmt -l` over the nine test files: no output (exit 0). `go vet ./internal/cli ./internal/hook ./internal/discovery ./internal/config ./internal/codexwiring ./internal/kanban ./internal/statusline ./internal/web`: exit 0. `GOOS=windows GOARCH=amd64 go vet` over the same eight packages: exit 0 (the windows vet type-checks the new test files). `go build ./...`: exit 0. `GOOS=windows GOARCH=amd64 go build ./...`: exit 0.
- Scope: `git diff --stat 29bf84c33 HEAD` at `1379abc9a` lists exactly eleven files: the spec.md frontmatter line, `internal/cli/testdata/lane_entry_env_golden.json`, and nine `*_test.go` files (1379 insertions, 1 deletion). `git diff --stat 29bf84c33 HEAD -- ':!*_test.go'` lists only `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/spec.md` (`2 +-`) and the golden (`31 +`). No production file changed; no template file changed.
- spec.md: `git diff` showed one changed line, `status: draft` to `status: in-progress`. The `updated:` field already read `2026-10-02`, the date of this run, so it was left as it was.

#### Gaps

- AC-004's `-l` parity (`TestLaneEntryEnvParity`) is M2's and M3's; at M1 only today's `-f lane` rows are compared with the golden.
- AC-017 clause "the hook emits no kanban notice" does NOT hold today. Measured in this run (scratch probe, removed): with `MOAI_KANBAN=1` and `MOAI_KANBAN_LABEL=plan` the real SessionStart handler returns `Kanban Mode: joined the kanban run as plan.` on both the agent context and the operator message. The AC text calls the test green today and the notice is deleted at M5b, so the M1 hook test asserts tolerance only (no error, the pre-existing record left byte-identical and readable); the absence of the notice is pinned by AC-014's `TestSessionStartEmitsNoKanbanNotice` at M5b. This is a contradiction inside AC-017, returned for the orchestrator.
- AC-017: the behavior of a live old session is unexercised (research.md R14 gap 8). The web test covers the two builders `buildOverview` and `buildKanban`; the statusline test covers the queue-count read, the only statusline reader of that state tree (the statusline has no session-record reader, measured by grep over `internal/statusline`); the doctor test covers `checkFactoryRun` and `checkOwnerLabelDrift`, not the whole `moai doctor` run.
- Mutants were applied to `cc.go` and `glm.go` for (a) and (g), and to the lane role marker for (c); not mutated: the leader's `MOAI_KANBAN_LEAD_ADDR` publish, the codex child markers in `codexCardLaunchEnv`, the codex lane claim `ClaimFactoryLaneWithin`, and the three platform process readers of `MOAI_KANBAN_ID` in `internal/discovery/leader_readers_*.go` (the discovery tests inject fakes; no test exercises a real process read). The net's codex lane row asserts the child markers and the matrix launch, not the registry claim.
- Mutants (g) and (h) are observed on today's `-f lane` path only; the AC requires them to be re-observed on the `-l` path after the M2 re-pin.
- The notice assertions of `TestFactoryNetSessionStartNotices` are limited to facts that survive the M4 rewrite (the run id, the leader socket, the lane label, the word leader); the lane-guidance wording is M4's and AC-010's.
- Only the selectors named above were run. The whole-package suites of the eight touched packages and the race detector were not run (scope discipline; no `moai slot` lease was taken because nothing wider than a selector ran). A test I did not run is not asserted green.
- plan-audit obligation J.5 item 1 (size the AC-018 exact-four-files grep sweep behind the audit's measured 123 files / 502 lines at M1, and record the measured edit list) was not part of this delegation and was not done; it remains open for the orchestrator.
- Worktree-guard refusals, each re-issued as a plain command with no measurement replaced by source reading: (1) a `sed` over `acceptance.md` through a shell variable was refused, and the file was read with the Read tool instead; (2) a `go test` compound that used a shell variable for the scratchpad path was refused, and the identical run was re-issued with literal paths; (3) a compound that read `${PIPESTATUS[0]}` was refused, and the runs were re-issued with the output redirected to a file and `echo $?` after it.
- Test hygiene note: the new tests use `t.Setenv` and `t.Chdir` and are therefore not parallel; no OTEL variable is set; no real network and no real engine binary is started (the engine launch and the codex child launch are substituted); the cc and glm launches write transient settings files under the OS temp directory, which the launcher removes on return.
- The golden was captured with the test file that is committed later (capture mode behind `MOAI_LANE_GOLDEN_WRITE`); there is no separate generator program, so the golden's provenance is this record plus the commit order, not a re-runnable tool outside that test.

#### Residual-risk

- The net proves the factory still publishes today's markers and records through today's seams; it does not prove a real `claude`, `glm` or `codex` session behaves the same, because every engine launch is substituted.
- The golden fixes the lane marker set for the three backends from one fixture each (run id `run-cli`, label `lane-1`, one picked card for codex). A marker that appears only under another option (`--clear-policy`, `--no-auto-dispatch`, `--factory-run`, the discovery path's `MOAI_KANBAN_LEAD_NAME`) is covered by AC-005's composition rows at M2, not by this golden.
- Mutant strength is bounded by the eight edits chosen; a defect shaped differently from them (for example a marker published with a wrong value rather than removed) is caught only where an assertion names the value.

### M2 evidence

Recorded by the run-phase implementation worker (cycle_type tdd) for milestone M2 of card t1399, on the branch `WT-launcher-entry-flags`. Start state, re-read before any change: `git rev-parse --short HEAD` printed `2fd2da666`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Every output block below is the verbatim output of a command run in this run on this tree; text outside a block states what it shows. Excerpts keep the verbatim lines and drop only repeated subtest rows, the count of the dropped rows being stated.

#### Claim

- M2 adds the cc/glm lane entry `-l` / `--lane` (a branch in `parseFactoryFlag` that selects the lane role and feeds the existing desugar; constants `laneFlagShort` and `laneFlagLong`), the `-l`/`--lane` refusals, the `-f <value>` refusal (every `-f lane*` form and every other value), the refusal of a lane-shaped `--name` beside `-f`, the `--leader` selector refusal, the reworded `-f <N>` line, gates re-keyed to `-l`, and updated help text and comments. `-l`'s slot comes from the unchanged shared claim (`NextFactoryLaneNumber`, then `resolveFactoryLaneName`); no slot logic was written.
- Production files changed, all in `internal/cli`: `factory.go` (grammar, refusals, constants), `cc.go` and `glm.go` (help text, `Use` line, one comment each), `kanban.go` (two comments), `factory_lane_relaunch.go` (one comment). `leadFlagShort` is kept (its remaining reader is `codex_factory.go:42-51`); `factory.go` no longer reads it. `-k` still parses.
- Tests: new `internal/cli/lane_entry_m2_test.go` (seven test functions, AC-001, 003, 004, 005, 006, 008); the M1 net lane launches and the enterable-pair matrix rows for Claude and GLM, and the golden capture's cc/glm rows, re-pinned from `-f lane` to `-l`; thirteen existing test files that drove the removed forms re-pinned (list under Gaps item 6).
- Commit 1 `6d9793aa6` (grammar, refusals, tests, net re-pin; build, vet and the net green at that commit); commit 2 is this progress record.

#### Evidence

First act, the RED-now of the explicit-name rows, the `--lane` rows and the `--leader` selector rows on this tree (ledger RED-11, RED-12, RED-14; the probe maps its test file in with `go test -overlay`, so the checkout is not written), env unset in the same invocation, exit 0 of the runner:

`go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/entry_probe.go -root .`

```text
=== RUN   TestZZT1399EntryProbe
CCPARSE args=["-f" "--name" "lane-2"] factoryEnabled=true rest=["--name" "lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["-f" "-n" "lane-2"] factoryEnabled=true rest=["-n" "lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["--factory" "--name=lane-2"] factoryEnabled=true rest=["--name=lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["-f" "--name" "leader-r7"] factoryEnabled=true rest=["--name" "leader-r7"] laneLabel="" isLane=false branch=1
CCPARSE args=["--lane"] factoryEnabled=false rest=["--lane"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "lane-2"] factoryEnabled=false rest=["--lane" "lane-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane=lane-2"] factoryEnabled=false rest=["--lane=lane-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "3"] factoryEnabled=false rest=["--lane" "3"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "leader-2"] factoryEnabled=false rest=["--lane" "leader-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "-f"] factoryEnabled=true rest=["--lane"] laneLabel="" isLane=false branch=1
CCPARSE args=["-l"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "lane-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "3"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l=lane-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "-f"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "-k"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "--name" "lane-2"] parseErr="--leader requires a leader label"
CCPARSE args=["-f" "-l"] parseErr="--leader requires a leader label"
CCPARSE args=["--lane" "-k"] factoryEnabled=false rest=["--lane"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "--name" "lane-2"] factoryEnabled=false rest=["--lane" "--name" "lane-2"] laneLabel="lane-2" isLane=true branch=0
CCPARSE args=["--leader" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["--leader=leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-f" "--leader" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-f" "--leader=leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CODEX args=["--lane"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane=lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "3"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "leader-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "-f"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--lane" "-k"] err= exitCode=1 stderr="KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead\n" stdoutLen=0
CODEX args=["--lane" "--name" "lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["-f" "-l"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--leader" "leader-2"] err=--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target exitCode=-1 stderr="" stdoutLen=0
CODEX args=["-f" "--leader" "leader-2"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
--- PASS: TestZZT1399EntryProbe (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.009s
exit 0
```

Read: `-f --name lane-2`, `-f -n lane-2` and `--factory --name=lane-2` route to the lane branch (`branch=2`, `isLane=true`), `-f --name leader-r7` to the leader (`branch=1`); `--lane` and every `--lane <x>` shape are accepted with the tokens left in the passthrough rest (`--lane --name lane-2` parses as a named lane-shaped session); `-l <anything>` fails with removed-form text (`--leader requires a leader label`, `applies to a factory lane join (-f lane / -f lane-<n>)`); the `--leader` selector without a lane entry is refused with the removed-form line. The codex rows are for M3.

RED of the new tests, written first and run before any production change (`go test ./internal/cli -run '^(TestLaneEntryJoinsNextFreeSlot|TestLaneEntryRefusals|TestLaneEntryEnvParity|TestLaneEntryComposesWithLaneOptions|TestLeaderSelectorRefusedWithoutLaneEntry|TestFactoryLaneSpellingsRefused|TestFactoryCountShapeStillRefused)$' -v -count=1`, exit 1). The seven parent tests fail; 66 subtests fail (10 + 30 + 8 + 16 + 2) and 2 pass (the two `named_leader_launches` rows, which assert behavior that is correct today). Top-level lines, then the first assertion lines (the full output was kept in the scratchpad, not in the tree):

```text
--- FAIL: TestLaneEntryJoinsNextFreeSlot (5.06s)
--- FAIL: TestLaneEntryRefusals (25.32s)
--- FAIL: TestLaneEntryEnvParity (1.35s)
--- FAIL: TestLaneEntryComposesWithLaneOptions (0.78s)
--- FAIL: TestLeaderSelectorRefusedWithoutLaneEntry (8.41s)
--- FAIL: TestFactoryLaneSpellingsRefused (21.53s)
--- FAIL: TestFactoryCountShapeStillRefused (1.67s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	65.137s
FAIL
    lane_entry_m2_test.go:128: cc -l: launched=false err=--leader requires a leader label
    lane_entry_m2_test.go:128: cc -l: launched=false err=--leader requires a leader label
    lane_entry_m2_test.go:128: cc -l: launched=false err=--leader requires a leader label
    lane_entry_m2_test.go:128: cc -l: launched=false err=--leader requires a leader label
    lane_entry_m2_test.go:192: [-l lane-2] refusal "--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target" names the removed form "-f lane"
    lane_entry_m2_test.go:198: [-l lane-2] refusal "--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target" lacks "no argument"
    lane_entry_m2_test.go:198: [-l lane-2] refusal "--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target" lacks "--leader <name>"
    lane_entry_m2_test.go:200: [-l -f] refusal "--leader requires a leader label" lacks "entry token"
    lane_entry_m2_test.go:192: [--lane lane-2] launched a session; a refusal launches nothing
    lane_entry_m2_test.go:192: [--lane lane-2] was accepted, want a refusal
    lane_entry_m2_test.go:192: [-l lane-2] refusal "--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target" names the removed form "-f lane"
    lane_entry_m2_test.go:198: [-l lane-2] refusal "--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target" lacks "no argument"
    lane_entry_m2_test.go:198: [-l lane-2] refusal "--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target" lacks "--leader <name>"
    lane_entry_m2_test.go:200: [-l -f] refusal "--leader requires a leader label" lacks "entry token"
    lane_entry_m2_test.go:358: [-f lane] launched a session; a refusal launches nothing
    lane_entry_m2_test.go:358: [-f lane] was accepted, want a refusal
    lane_entry_m2_test.go:358: [-f --name lane-2] launched a session; a refusal launches nothing
    lane_entry_m2_test.go:358: [-f --name lane-2] was accepted, want a refusal
    lane_entry_m2_test.go:358: [-f lane] launched a session; a refusal launches nothing
exit status of the run: 1
```

Production change, then GREEN. Commands (each ran as one compound invocation `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test ...`; exit code 0 on each, shown as `RC=0` by the invoking tool call):

```text
$ go test ./internal/cli -run '^TestLaneEntryJoinsNextFreeSlot$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestLaneEntryJoinsNextFreeSlot (16.20s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F1_cc_-l (1.36s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F1_cc_--lane (1.60s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F1_glm_-l (1.76s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F1_glm_--lane (1.45s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F2_cc_-l (1.56s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F2_glm_-l (1.55s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F3_cc_-l (1.33s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F3_glm_-l (1.94s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F4_cc_-l (2.23s)
    --- PASS: TestLaneEntryJoinsNextFreeSlot/F4_glm_-l (1.42s)
ok  	github.com/modu-ai/moai-adk/internal/cli	17.236s
  swept: parent PASS=1, PASS lines incl. parent=11, FAIL=0, SKIP=0

$ go test ./internal/cli -run '^TestLaneEntryRefusals$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestLaneEntryRefusals (74.02s)
    --- PASS: TestLaneEntryRefusals/cc_-l_lane-2 (1.23s)
    --- PASS: TestLaneEntryRefusals/cc_-l_3 (2.24s)
    --- PASS: TestLaneEntryRefusals/cc_-l=lane-2 (2.31s)
    --- PASS: TestLaneEntryRefusals/cc_-l_leader-2 (2.17s)
    --- PASS: TestLaneEntryRefusals/cc_-l_-f (3.14s)
    --- PASS: TestLaneEntryRefusals/cc_-f_-l (2.23s)
    --- PASS: TestLaneEntryRefusals/cc_-l_-k (1.87s)
    --- PASS: TestLaneEntryRefusals/cc_-l_--name_lane-2 (2.70s)
    --- PASS: TestLaneEntryRefusals/cc_--lane_lane-2 (2.50s)
    --- PASS: TestLaneEntryRefusals/cc_--lane=lane-2 (3.04s)
    --- PASS: TestLaneEntryRefusals/cc_--lane_3 (3.20s)
    --- PASS: TestLaneEntryRefusals/cc_--lane_leader-2 (3.16s)
    --- PASS: TestLaneEntryRefusals/cc_--lane_-f (1.58s)
    --- PASS: TestLaneEntryRefusals/cc_--lane_-k (1.53s)
    --- PASS: TestLaneEntryRefusals/cc_--lane_--name_lane-2 (1.47s)
    --- PASS: TestLaneEntryRefusals/glm_-l_lane-2 (1.76s)
    --- PASS: TestLaneEntryRefusals/glm_-l_3 (2.27s)
    --- PASS: TestLaneEntryRefusals/glm_-l=lane-2 (2.50s)
    --- PASS: TestLaneEntryRefusals/glm_-l_leader-2 (2.76s)
    --- PASS: TestLaneEntryRefusals/glm_-l_-f (2.94s)
    --- PASS: TestLaneEntryRefusals/glm_-f_-l (2.89s)
    --- PASS: TestLaneEntryRefusals/glm_-l_-k (2.39s)
    --- PASS: TestLaneEntryRefusals/glm_-l_--name_lane-2 (2.04s)
    --- PASS: TestLaneEntryRefusals/glm_--lane_lane-2 (2.09s)
    --- PASS: TestLaneEntryRefusals/glm_--lane=lane-2 (2.28s)
    --- PASS: TestLaneEntryRefusals/glm_--lane_3 (4.91s)
    --- PASS: TestLaneEntryRefusals/glm_--lane_leader-2 (3.05s)
    --- PASS: TestLaneEntryRefusals/glm_--lane_-f (3.12s)
    --- PASS: TestLaneEntryRefusals/glm_--lane_-k (2.58s)
    --- PASS: TestLaneEntryRefusals/glm_--lane_--name_lane-2 (2.03s)
ok  	github.com/modu-ai/moai-adk/internal/cli	75.094s
  swept: parent PASS=1, PASS lines incl. parent=31, FAIL=0, SKIP=0

$ go test ./internal/cli -run '^TestLaneEntryEnvParity$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestLaneEntryEnvParity (7.80s)
ok  	github.com/modu-ai/moai-adk/internal/cli	9.137s
  swept: parent PASS=1, PASS lines incl. parent=1, FAIL=0, SKIP=0

$ go test ./internal/cli -run '^TestLaneEntryComposesWithLaneOptions$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestLaneEntryComposesWithLaneOptions (9.13s)
ok  	github.com/modu-ai/moai-adk/internal/cli	10.405s
  swept: parent PASS=1, PASS lines incl. parent=1, FAIL=0, SKIP=0

$ go test ./internal/cli -run '^TestLeaderSelectorRefusedWithoutLaneEntry$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestLeaderSelectorRefusedWithoutLaneEntry (12.75s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/cc_--leader_leader-2 (1.60s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/cc_--leader=leader-2 (1.30s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/cc_-f_--leader_leader-2 (1.62s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/cc_-f_--leader=leader-2 (1.34s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/glm_--leader_leader-2 (1.07s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/glm_--leader=leader-2 (1.41s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/glm_-f_--leader_leader-2 (1.72s)
    --- PASS: TestLeaderSelectorRefusedWithoutLaneEntry/glm_-f_--leader=leader-2 (2.70s)
ok  	github.com/modu-ai/moai-adk/internal/cli	14.016s
  swept: parent PASS=1, PASS lines incl. parent=9, FAIL=0, SKIP=0

$ go test ./internal/cli -run '^TestFactoryLaneJoinLeadTargeting$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestFactoryLaneJoinLeadTargeting (18.15s)
    --- PASS: TestFactoryLaneJoinLeadTargeting/default_targets_the_canonical_leader_label (7.28s)
    --- PASS: TestFactoryLaneJoinLeadTargeting/explicit_--leader_targets_that_leader (7.67s)
    --- PASS: TestFactoryLaneJoinLeadTargeting/unmatched_label_refuses_(AC-011) (3.19s)
ok  	github.com/modu-ai/moai-adk/internal/cli	19.564s
  swept: parent PASS=1, PASS lines incl. parent=4, FAIL=0, SKIP=0

$ go test ./internal/cli -run '^TestFactoryLaneSpellingsRefused$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestFactoryLaneSpellingsRefused (40.30s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_-f_lane (2.11s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_-f_lane-2 (2.10s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_--factory_lane (2.03s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_-f=lane (2.08s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_--factory=lane-2 (2.96s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_-f_--name_lane-2 (1.98s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_-f_-n_lane-2 (2.03s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_--factory_--name=lane-2 (2.29s)
    --- PASS: TestFactoryLaneSpellingsRefused/cc_named_leader_launches (1.37s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_-f_lane (1.78s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_-f_lane-2 (1.79s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_--factory_lane (1.44s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_-f=lane (2.29s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_--factory=lane-2 (2.59s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_-f_--name_lane-2 (2.43s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_-f_-n_lane-2 (2.84s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_--factory_--name=lane-2 (2.36s)
    --- PASS: TestFactoryLaneSpellingsRefused/glm_named_leader_launches (3.84s)
ok  	github.com/modu-ai/moai-adk/internal/cli	41.623s
  swept: parent PASS=1, PASS lines incl. parent=19, FAIL=0, SKIP=0

$ go test ./internal/cli -run '^TestFactoryCountShapeStillRefused$' -v -count=1   (env unset in the same invocation; exit 0)
--- PASS: TestFactoryCountShapeStillRefused (5.96s)
    --- PASS: TestFactoryCountShapeStillRefused/cc (3.31s)
    --- PASS: TestFactoryCountShapeStillRefused/glm (2.65s)
ok  	github.com/modu-ai/moai-adk/internal/cli	7.515s
  swept: parent PASS=1, PASS lines incl. parent=3, FAIL=0, SKIP=0
```

AC matrix (HEAD `6d9793aa6` for the code, this tree):

| AC | Result | Command (exit 0) | Swept |
|----|--------|------------------|-------|
| AC-001 | PASS | `go test ./internal/cli -run '^TestLaneEntryJoinsNextFreeSlot$' -v -count=1` | 1 parent, 10 subtests (F1 cc/glm with `-l` and `--lane`, F2 to F4 cc/glm), 0 FAIL, 0 SKIP |
| AC-003 (cc, glm) | PASS | `go test ./internal/cli -run '^TestLaneEntryRefusals$' -v -count=1` | 1 parent, 30 subtests (15 shapes x cc, glm), 0 FAIL, 0 SKIP; the 15 codex subtests (45 total) are authored at M3 |
| AC-004 (cc, glm) | PASS | `go test ./internal/cli -run '^TestLaneEntryEnvParity$' -v -count=1` | 1 PASS result (no subtests): cc and glm, `-l` and `--lane`, equal the committed golden label aside, every key a constant of `envkeys.go`; the codex row joins at M3 |
| AC-005 | PASS (cc, glm) | `...TestLaneEntryComposesWithLaneOptions` (1 PASS); `...TestLeaderSelectorRefusedWithoutLaneEntry` (1 parent, 8 subtests; the 2 codex subtests of the AC's 10 are for M3); `...TestFactoryLaneJoinLeadTargeting` (1 parent, 3 subtests) | `grep -c '"-l", "leader' internal/cli/factory_join_discovery_test.go` printed 0 |
| AC-006 | PASS | `go test ./internal/cli -run '^TestFactoryLaneSpellingsRefused$' -v -count=1` | 1 parent, 16 refusal subtests, 2 launch subtests, 0 FAIL, 0 SKIP |
| AC-008 | PASS | `go test ./internal/cli -run '^TestFactoryCountShapeStillRefused$' -v -count=1` | 1 parent, 2 subtests |
| AC-004 cc/glm golden row | PASS | the M1 golden test `TestLaneMarkerGoldenMatchesFLane`, cc and glm now through `-l`, ran in the 115-test targeted set below | its codex row still uses `-f lane` until M3 |

The M1 net after the re-pin (AC-015 selectors, swept 8 cli, 2 hook, 2 discovery):

```text
$ go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestPrepareKanbanSettingsWritesTransientFile|TestPrepareFactorySettingsWritesTransientFile|TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch|TestFactoryNetBlockCap|TestFactoryEntryMatrix)$' -v -count=1   (exit 0)
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (14.44s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (8.16s)
--- PASS: TestFactoryNetLeaderLaunch (4.32s)
--- PASS: TestFactoryNetLaneLaunch (5.51s)
--- PASS: TestFactoryNetBlockCap (2.80s)
--- PASS: TestFactoryEntryMatrix (27.21s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (5.76s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/cli	70.177s
  swept: top-level PASS=8, FAIL=0, SKIP=0

$ go test ./internal/hook -run '^(TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices)$' -v -count=1   (exit 0)
--- PASS: TestFactoryNetSessionRecord (0.05s)
--- PASS: TestFactoryNetSessionStartNotices (4.26s)
ok  	github.com/modu-ai/moai-adk/internal/hook	5.324s
  swept: top-level PASS=2, FAIL=0

$ go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1   (exit 0)
--- PASS: TestDiscoverLeaderVerifiesLiveLeader (0.81s)
--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID (0.25s)
ok  	github.com/modu-ai/moai-adk/internal/discovery	1.473s
  swept: top-level PASS=2, FAIL=0
```

The two mutants re-observed on the `-l` path after the re-pin (AC-015 (g) and (h)). The mutated copies of `cc.go`, `glm.go` and `kanban.go` live in the scratchpad and are applied with `go test -overlay`, so the tree was never edited; `grep -rn MUTANT internal/` finds only pre-existing comments in unrelated packages, and `git status --short` showed no change to those files beyond the intended M2 edits.

```text
$ go test -overlay <scratch>/m2/overlay_g.json ./internal/cli -run '<the 9-name net selector>' -v -count=1   (mutant g: the resolveFactoryLaneName call replaced by 'finalLabel, claimErr := factoryLabel, error(nil)' in cc.go and glm.go, mutated copies kept OUTSIDE the tree; exit 1)
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (12.69s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (11.97s)
--- PASS: TestFactoryNetLeaderLaunch (6.16s)
--- FAIL: TestFactoryNetLaneLaunch (10.68s)
    --- FAIL: TestFactoryNetLaneLaunch/cc (4.53s)
    --- FAIL: TestFactoryNetLaneLaunch/glm (6.15s)
--- PASS: TestFactoryNetBlockCap (5.51s)
--- PASS: TestFactoryEntryMatrix (30.64s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (2.73s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	81.980s
FAIL
    factory_net_m1_test.go:343: registry[lane-1] = ({PID:0 RegisteredAt:}, false), want a live claim under pid 76136
    factory_net_m1_test.go:352: second lane MOAI_FACTORY_WORKER = "lane-1", want "lane-2" (the first claim is live)
    factory_net_m1_test.go:356: registry lost the second claim: map[]
    factory_net_m1_test.go:343: registry[lane-1] = ({PID:0 RegisteredAt:}, false), want a live claim under pid 76136
    factory_net_m1_test.go:352: second lane MOAI_FACTORY_WORKER = "lane-1", want "lane-2" (the first claim is live)
    factory_net_m1_test.go:356: registry lost the second claim: map[]

$ go test -overlay <scratch>/m2/overlay_h.json ./internal/cli -run '<the 9-name net selector>' -v -count=1   (mutant h: the MOAI_KANBAN_BACKEND Setenv removed from exportKanbanLaunchFacts in kanban.go, mutated copy OUTSIDE the tree; exit 1)
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (18.10s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (13.15s)
--- FAIL: TestFactoryNetLeaderLaunch (3.93s)
    --- FAIL: TestFactoryNetLeaderLaunch/cc (1.72s)
    --- FAIL: TestFactoryNetLeaderLaunch/glm (2.21s)
--- FAIL: TestFactoryNetLaneLaunch (8.49s)
    --- FAIL: TestFactoryNetLaneLaunch/cc (4.27s)
    --- FAIL: TestFactoryNetLaneLaunch/glm (4.22s)
--- PASS: TestFactoryNetBlockCap (2.77s)
--- PASS: TestFactoryEntryMatrix (23.01s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (2.72s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	73.707s
FAIL
    factory_net_m1_test.go:269: MOAI_KANBAN_BACKEND at launch = "", want "claude"
    factory_net_m1_test.go:269: MOAI_KANBAN_BACKEND at launch = "", want "glm"
    factory_net_m1_test.go:332: MOAI_KANBAN_BACKEND at launch = "", want "claude"
    factory_net_m1_test.go:332: MOAI_KANBAN_BACKEND at launch = "", want "glm"
```

Targeted regression set (115 test functions drawn from every test file the change touched, plus the kanban help, vocabulary-guard, cg-retirement and role-pin files): `go test ./internal/cli -count=1 -timeout 40m -run '^(...115 names...)$'` ran once in full: exit 1, 402.354s, exactly one failing test, `TestACFB019_HelpDocumentsCompanionEntry` (cc and glm), whose assertion `kanban_help_test.go:72: glm help does not state the companion name collision bump` needs the literal sentence "bumped to the next free number" that my help edit had removed from the factory section. The sentence was restored to the `-k <N> / -k <N> --name lane-<i>` help item (where it is true for the still-valid `-k` lane form) in both launchers, and the family `TestACFB018_HelpDocumentsLeadEntry|TestACFB019_HelpDocumentsCompanionEntry|TestFactoryGenealogyInHelp|TestLauncherHelpLaneVocabulary|TestProductionStringLiteralsUseLeaderLaneVocabulary|TestCGRetirementHelpDoesNotOfferCG|TestCGRetirementLiveHelpHasNoLaunchRecommendation` then passed (7 of 7, exit 0). The AC-009 selector `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestLeaderEntryUnchanged|TestCGRetiredEntryAndModeHaveZeroEffects|TestCGRetirementCompleteEntryShapesAndCounters|TestBareMoaiPrintsBannerAndHelp)$' -v -count=1` passed 5 of 5 (exit 0), and `TestCGEntryGuardRunsBeforeLaunchAndSpawn` (the one test in an untouched file that passes `-f`) passed (exit 0).

Build and static checks, after the last edit (tree HEAD `6d9793aa6` plus only this progress file):

```text
$ go build ./...                               -> no output, exit 0 (echo BUILD_OK printed)
$ GOOS=windows GOARCH=amd64 go build ./...     -> no output, exit 0 (echo WINBUILD_OK printed)
$ go vet ./internal/cli/                       -> no output, exit 0 (echo VET_OK printed)
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/   -> no output, exit 0 (echo WINVET_OK printed)
$ gofmt -l <the 19 files of commit 1>          -> no output
```

`gofmt -l internal/cli/` over the whole directory prints one pre-existing file, `internal/cli/mcp_claude.go`, which this milestone did not touch.

#### Baseline-attribution

- Tree: HEAD `2fd2da666` at the start (RED-now and RED measurements, before the production change), `6d9793aa6` for the GREEN, net, mutant and build measurements. The judging build is the Go toolchain compiling this tree; no installed `moai` binary was used as a measuring instrument (the `moai slot` commands used only the lease).
- The two mutants were measured against the same tree as the green net (HEAD `6d9793aa6`), through an overlay of one or two files, so the only difference between the green and red runs of the same selector is the mutated file.
- The heavy runs ran under `moai slot acquire --resource t1399-run --max-duration 60m` and the lease was released (`moai slot release --resource t1399-run` printed `released`) after the last run, before commit 1.

#### Gaps

1. The whole `internal/cli` suite was NOT completed: one `go test ./internal/cli -count=1 -timeout 50m` run was started in the background and stopped by the harness's background time limit with no output (load average 16 to 38 on the machine); it measures nothing and is not cited. The 115-test targeted set above, run in full once, is the widest measurement; a regression in a test file outside it and outside the nine named selectors is unobserved.
2. AC-003 (45 subtests), AC-005's selector test (10 subtests) and AC-004's codex row are staged: M2 sweeps the cc and glm halves (30, 8, and the cc/glm rows), the codex rows are authored at M3 where the codex parse changes. The AC text lists the full counts as flipped by M2 plus M3.
3. The RED output of the seven new tests is kept as the excerpt above; the full RED file lives in the scratchpad (outside the tree).
4. Worktree-guard refusals, each re-issued as a plain command with the same measurement: (a) a `cat > file <<EOF` heredoc followed by a second command, refused as "too complex to verify"; (b) two `python3 <<EOF` heredoc edits, refused, while three earlier ones of the same shape had been accepted, with no visible difference between them; (c) `go test ... ; echo rc=$?` forms, refused, re-issued as `go test ... && echo RC=0`; (d) a pipeline with `go test ... ; grep ... | head`, refused, re-issued as separate commands. Where a heredoc was refused the edit was made with the edit tool instead; no measurement was replaced by reading source.
5. `git -C <own absolute path> status --short` was run once (permitted by the guard, forbidden by the brief's wording); every other git command was plain `git`.
6. Existing tests re-pinned (the plan names only the net): `factory_test.go`, `factory_join_discovery_test.go`, `factory_autodispatch_test.go`, `factory_legacy_collision_test.go`, `factory_mixed_test.go`, `factory_worker_naming_test.go`, `factory_role_refusal_m2_test.go`, `factory_m7_test.go` (launcher-parse rows only; the codex rows are untouched), `factory_m4_test.go`, `factory_m6_test.go`, `codex_factory_retire_test.go` (cc/glm join rows only), `factory_net_m1_test.go`, `lane_entry_golden_m1_test.go`. They drove `-f lane`, `-f lane-<n>` or `-l <leader>`, which are refused after this change, so leaving them would put the suite red.
7. Comment-only update targets outside `internal/cli` (`internal/config/defaults.go:689,700`, `internal/config/envkeys.go:278,357`, `internal/kanban/bootstrap.go:370`) were NOT edited: the brief limits M2 to `internal/cli`.
8. Notice strings (`internal/hook`), the factory card errors (`factory_card.go:95,98,101`) and the Codex files still teach `-f lane`; they are M3 and M4 and merge with M2 as one integration unit.

#### Residual-risk

- Until M3 and M4 land, the leader SessionStart notice, the stale-run hint and the factory card rejoin errors still print `moai cc -f lane-<i>` / `-f lane`, which this milestone now refuses; M2 must not be merged without M3 and M4 (plan §D integration unit).
- `-k N --name lane-<i>` still reaches the lane branch (it is the kanban parse, removed at M5a) and still takes an operator-typed number; the `-f` side of the dispatch table no longer has a lane trigger.
- The pass-through fix (below) changes a launch shape that the old code mishandled; no test pinned the old behavior, but a script relying on `-f lane -- <args>` becoming a leader would change (it was refused anyway after this milestone).
- Mutant strength is bounded by the two edits chosen; a lane claim that records the wrong pid, or a marker published with a wrong value, is caught only where `TestFactoryNetLaneLaunch` and `TestLaneEntryJoinsNextFreeSlot` assert it.

#### Findings (statements the tree contradicts, no SPEC file edited)

- `acceptance.md` AC-005 (the `-- --print` passthrough row) says the lane composes "as it composed with `-f lane`". On this tree it did not: `parseLauncherEntry` appended the desugared `--name lane-<n>` after the `--` marker, where `parseFactoryLaneLabel` stops reading, so `-f lane -- --print` resolved to `factoryBranchLeader` (measured: `entry rest=[-- --print --name lane-5] ... branch=1`). M2 inserts the desugared `--name` ahead of the marker (`insertBeforePassthrough`, `factory.go`) and `TestLaneEntryComposesWithLaneOptions` asserts a lane launch with `-l -- --print` on cc and glm. This is a deviation from "as before" that the plan did not list.
- `plan.md` M2 says the `-f <N>` message is only reworded. The constant `factoryFlagUsageError` is also read by the Codex parse (`codex_factory.go:97`), so the Codex `-f <value>` usage line now carries the new text too; M3 owns the Codex refusal line.
- `plan.md` §B.4 lists `session_stale_run.go`, `session_start_factory*.go` and `factory_card.go` as prints to update: none was changed here (M4).

### M3 evidence

Recorded by the run-phase implementation worker (cycle_type tdd) for milestone M3 of card t1399 (the Codex launcher entry), on the branch `WT-launcher-entry-flags`. Start state, re-read before any change: `git rev-parse --short HEAD` printed `c0e8ce07e`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Every output block is the verbatim output of a command run in this run on this tree; excerpts say what they drop. Full outputs of the larger runs were kept in the session scratchpad (outside the tree).

#### Claim

- `moai codex -l` / `--lane` is the trigger of the supervising relaunch lane that `-f lane` started; every `-f` shape on `moai codex` is refused with the new one line (`FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex has no -f entry; the lane entry is 'moai codex -l'; use 'moai cc -f' or 'moai glm -f' for the factory leader`); `-l`/`--lane` refuses an argument, a second entry token, and an operator `--name`; `--leader` composes with the lane entry only; the Codex leader stays refused; `leadFlagShort` is deleted; the Codex usage line and help text name `-l`.
- Production files changed (all `internal/cli`): `codex_launcher.go` (classification `codexFactoryEntryClassify` rewritten, refusal constants, usage and help text, comments), `codex_factory.go` (the Codex parse now consumes `-l`/`--lane`, `--leader` and `--factory-run`; the `-f` arms, the `-l` short of `--leader` and `codexHeadTokenIsVerb` — orphaned by that change, its only reader was the `-f` value arm — are gone), `factory.go` (`leadFlagShort` deleted; two comments). Test files: new `lane_entry_m3_test.go`; codex rows added to `lane_entry_m2_test.go`; `factory_net_m1_test.go` (`netCodexLaneChildFor`, codex lane row re-pinned from `-f lane` to `-l`), `factory_m4_test.go`, `factory_m5_test.go`, `factory_m7_test.go`, `factory_join_discovery_test.go`, `codex_factory_retire_test.go`, `lane_entry_golden_m1_test.go`, `codex_debug_composition_test.go` (comment) re-pinned.
- Commits: `69588367a` (grammar, refusals, tests, net re-pin; build, vet and the net green at that commit), `1c3ab47ef` (a comment-only test edit), then this progress record.

#### Evidence

First act, the codex rows of RED-13 and RED-14 on this tree (`go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/entry_probe.go -root .`, env unset in the same invocation, runner exit 0; the cc rows of the same output are M2's and are dropped here; where a row repeats the generic usage line it is abbreviated "(same usage line)" in this record only):

```text
CODEX args=["--lane"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "lane-2"] err= exitCode=1 stderr=(same usage line) stdoutLen=0
CODEX args=["--lane=lane-2"] err= exitCode=1 stderr=(same usage line) stdoutLen=0
CODEX args=["--lane" "3"] err= exitCode=1 stderr=(same usage line) stdoutLen=0
CODEX args=["--lane" "leader-2"] err= exitCode=1 stderr=(same usage line) stdoutLen=0
CODEX args=["--lane" "-f"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--lane" "-k"] err= exitCode=1 stderr="KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead\n" stdoutLen=0
CODEX args=["--lane" "--name" "lane-2"] err= exitCode=1 stderr=(same usage line) stdoutLen=0
CODEX args=["-f" "-l"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--leader" "leader-2"] err=--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target exitCode=-1 stderr="" stdoutLen=0
CODEX args=["-f" "--leader" "leader-2"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
--- PASS: TestZZT1399EntryProbe (0.81s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.215s
exit 0
```

Read: on codex the `--lane` shapes die on the generic usage line, `--lane -f` and `-f -l` on the removed-form `-f lane` line, `--lane -k` on the kanban line, `--leader` without a lane entry returns the removed-form `-f lane / -f lane-<n>` error, and `-f --leader` the removed-form `-f lane` line — the RED-13 and RED-14 codex rows, red for the stated reason (none names `-l` as taking no argument or the lane entry).

RED of the new and re-pinned tests, written first and run before any production change (`go test ./internal/cli -run '^(TestCodexLaneEntryStartsRelaunchLane|TestCodexFactoryFlagRefusals|TestLaneEntryRefusals|TestLaneEntryEnvParity|TestLeaderSelectorRefusedWithoutLaneEntry|TestCodexFactoryEntryParsingUsesLaneOnly|TestCodexFactoryLegacyEntryIsRefused|TestSD_AC004_CodexOtherFactoryShapesRefused|TestCodexFactoryLeadFlagSurface|TestSD_AC021_LegacySpellingsRefused)$' -v -count=1`, env unset in the same invocation, exit 1): ten parent tests FAIL and 23 subtests FAIL (1 + 15 + 2 + 5), 0 SKIP. Top-level lines, the failing subtest set summarised, then a selection of assertion lines (long lines shortened with "..."):

```text
--- FAIL: TestCodexFactoryLegacyEntryIsRefused (0.00s)
--- FAIL: TestCodexFactoryEntryParsingUsesLaneOnly (0.00s)
--- FAIL: TestCodexFactoryLeadFlagSurface (0.00s)
--- FAIL: TestSD_AC004_CodexOtherFactoryShapesRefused (0.35s)
--- FAIL: TestSD_AC021_LegacySpellingsRefused (0.02s)
--- FAIL: TestLaneEntryRefusals (67.13s)
--- FAIL: TestLaneEntryEnvParity (7.75s)
--- FAIL: TestLeaderSelectorRefusedWithoutLaneEntry (14.31s)
--- FAIL: TestCodexLaneEntryStartsRelaunchLane (0.81s)
--- FAIL: TestCodexFactoryFlagRefusals (10.88s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	102.516s
(failing subtests: TestSD_AC021_LegacySpellingsRefused/moai_codex; TestLaneEntryRefusals/codex_* all 15; TestLeaderSelectorRefusedWithoutLaneEntry/codex_* both; TestCodexFactoryFlagRefusals/* all 5)
    codex_factory_retire_test.go:167: parse ["-l" "cli"]: rest=[] entry={Enabled:false ... Lead:cli ...} err=--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target
    factory_m5_test.go:210: the refusal line does not name "moai codex -l": "FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader"
    factory_m5_test.go:214: the refusal line names the removed form "-f lane": "FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only ..."
    factory_m5_test.go:218: shape [-f lane]: exit code = (0, false), want (1, true); err=factory next: refused — this verb runs from the parent checkout ..., not from .../internal/cli
    factory_m7_test.go:103: -f worker: stderr "\"worker\" is the legacy role token; 'moai codex -f lane' is the only Codex factory entry\n" does not name "moai codex -l"
    lane_entry_m2_test.go:219: codex [-l lane-2]: exit code = (0, false), want (1, true); err=--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target
    lane_entry_m2_test.go:219: codex [-l -f]: refusal "FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; ..." lacks "entry token"
    lane_entry_m2_test.go:219: [-l -k] refusal "KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead" lacks "entry token"
    lane_entry_m2_test.go:219: [--lane lane-2] refusal "unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app" lacks "-l"
exit status of the run: 1
```

(The full RED file has 33 FAIL lines. The `factory_m5_test.go:218` line shows today's `-f lane` starting the relaunch loop, which the re-pinned shape list refuses.)

Production change, then GREEN (same selector, exit 0, 0 FAIL, 0 SKIP):

```text
--- PASS: TestCodexFactoryLegacyEntryIsRefused (0.00s)
--- PASS: TestCodexFactoryEntryParsingUsesLaneOnly (0.00s)
--- PASS: TestCodexFactoryLeadFlagSurface (0.00s)
--- PASS: TestSD_AC004_CodexOtherFactoryShapesRefused (0.00s)
--- PASS: TestSD_AC021_LegacySpellingsRefused (0.02s)
--- PASS: TestLaneEntryRefusals (67.45s)
--- PASS: TestLaneEntryEnvParity (14.95s)
--- PASS: TestLeaderSelectorRefusedWithoutLaneEntry (16.90s)
--- PASS: TestCodexLaneEntryStartsRelaunchLane (11.38s)
--- PASS: TestCodexFactoryFlagRefusals (7.20s)
ok  	github.com/modu-ai/moai-adk/internal/cli	119.220s
```

Subtest counts from the same run: `grep -c "    --- PASS: TestLaneEntryRefusals/codex"` printed 15 and `grep -c "    --- PASS: TestLaneEntryRefusals/"` printed 45 (15 shapes x cc, glm, codex); `TestLeaderSelectorRefusedWithoutLaneEntry/codex` 2; `TestCodexFactoryFlagRefusals/` 5.

Per-AC commands, each its own run in this run (env unset in the same invocation, exit 0):

```text
$ go test ./internal/cli -run '^TestCodexLaneEntryStartsRelaunchLane$' -v -count=1
--- PASS: TestCodexLaneEntryStartsRelaunchLane (11.39s)
ok  	github.com/modu-ai/moai-adk/internal/cli	12.737s
$ go test ./internal/cli -run '^TestLaneEntryEnvParity$' -v -count=1
--- PASS: TestLaneEntryEnvParity (15.36s)
ok  	github.com/modu-ai/moai-adk/internal/cli	16.383s
$ go test ./internal/cli -run '^TestLaneMarkerGoldenMatchesFLane$' -v -count=1
--- PASS: TestLaneMarkerGoldenMatchesFLane (5.32s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/cc (0.99s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/glm (0.98s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/codex (3.35s)
ok  	github.com/modu-ai/moai-adk/internal/cli	6.111s
$ go test ./internal/cli -run '^TestLeaderSelectorRefusedWithoutLaneEntry$' -v -count=1
--- PASS: TestLeaderSelectorRefusedWithoutLaneEntry (10.64s)   (10 subtests PASS: 4 cc, 4 glm, 2 codex)
ok  	github.com/modu-ai/moai-adk/internal/cli	11.351s
$ go test ./internal/cli -run '^TestCodexFactoryFlagRefusals$' -v -count=1
--- PASS: TestCodexFactoryFlagRefusals (6.58s)   (5 subtests PASS: -f, -f lane, -f lane-2, -f 3, --factory lane)
ok  	github.com/modu-ai/moai-adk/internal/cli	7.751s
```

The AC-004 golden test compares the codex row now through `-l`; the golden was captured with the codex row's `MOAI_KANBAN_LABEL` and `TestLaneMarkerGoldenMatchesFLane` still compares it whole (the label is removed at M5a); `TestLaneEntryEnvParity`'s codex rows (`-l` and `--lane`) exclude exactly that key.

The three byte-pinned refusal tests, anchored selector, swept 3 (exit 0):

```text
$ go test ./internal/cli -run '^(TestCodexFactoryEntryParsingUsesLaneOnly|TestCodexFactoryLegacyEntryIsRefused|TestSD_AC004_CodexOtherFactoryShapesRefused)$' -v -count=1
--- PASS: TestCodexFactoryLegacyEntryIsRefused (0.00s)
--- PASS: TestCodexFactoryEntryParsingUsesLaneOnly (0.00s)
--- PASS: TestSD_AC004_CodexOtherFactoryShapesRefused (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.448s
```

The net after the re-pin (AC-015 selectors, on HEAD `69588367a`, exit 0 each): 8 cli, 2 hook, 2 discovery.

```text
$ go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestPrepareKanbanSettingsWritesTransientFile|TestPrepareFactorySettingsWritesTransientFile|TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch|TestFactoryNetBlockCap|TestFactoryEntryMatrix)$' -v -count=1
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (8.35s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (5.36s)
--- PASS: TestFactoryNetLeaderLaunch (2.57s)
--- PASS: TestFactoryNetLaneLaunch (5.48s)
--- PASS: TestFactoryNetBlockCap (3.20s)
--- PASS: TestFactoryEntryMatrix (13.99s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (1.31s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	41.428s
$ go test ./internal/hook -run '^(TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices)$' -v -count=1
--- PASS: TestFactoryNetSessionRecord (0.01s)
--- PASS: TestFactoryNetSessionStartNotices (3.63s)
ok  	github.com/modu-ai/moai-adk/internal/hook	4.575s
$ go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1
--- PASS: TestDiscoverLeaderVerifiesLiveLeader (0.38s)
--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID (0.24s)
ok  	github.com/modu-ai/moai-adk/internal/discovery	1.018s
```

Wider regression sets, each run once in full on this tree (env unset, exit 0, 0 FAIL): `go test ./internal/cli -run '^TestCodex' -count=1 -timeout 9m -v` — 388 top-level PASS, 0 FAIL, 8 SKIP (`ok ... 226.904s`; the skips are in tests this milestone did not touch); and `go test ./internal/cli -run '^(TestACFB018_HelpDocumentsLeadEntry|TestACFB019_HelpDocumentsCompanionEntry|TestFactoryGenealogyInHelp|TestLauncherHelpLaneVocabulary|TestProductionStringLiteralsUseLeaderLaneVocabulary|TestCGRetire.*|TestFactoryLaneJoin.*|TestLaneEntryComposesWithLaneOptions|TestFactoryLaneSpellingsRefused|TestFactoryCountShapeStillRefused|TestSD_.*|TestFactoryJoin.*|TestLeaderEntryUnchanged|TestFactoryRolePin.*)$' -count=1 -timeout 9m -v` — 59 top-level PASS, 0 FAIL, 0 SKIP. `grep -c '"-l", "leader' internal/cli/factory_join_discovery_test.go` printed 0.

Codex lane reachability, measured (the question: is the interactive Codex lane branch — `enterCodexFactory`, reached from `runCodexLaunch`, plus `codexFactoryEnv` and the claim-stamping seams — reachable once classification routes `-l` to the relaunch loop first?). Method: a scratch copy of `codex_factory.go` whose `enterCodexFactory` first appends the caller's file:line and the entry's `Enabled`/`LaneRole` to a log file named by an environment variable, applied through `go test -overlay` (the tree was not edited). Test set: every test that drives `runCodex` for a launch or a refusal plus the codex launcher and debug families (`^(TestCodexLaneEntryStartsRelaunchLane|TestCodexFactoryFlagRefusals|TestSD_AC003_CodexRelaunchPerCard|TestSD_AC004_CodexOtherFactoryShapesRefused|TestSD_AC021_LegacySpellingsRefused|TestCodexFactoryLegacyEntryIsRefused|TestCodexKanbanEntryIsRefused|TestCodexEntryRefusalHasNoStateEffect|TestCodexEntryTokensAfterDashDashPassThrough|TestCodexFactoryLaneJoinsDiscoveredLeader|TestFactoryEntryMatrix|TestCodexLaunch.*|TestCodexDebug.*|TestCodexSpawn.*|TestCodexDirect.*)$`, exit 0, 54 top-level PASS, 0 FAIL). Observed log, `sort | uniq -c`:

```text
   3 enterCodexFactory caller=codex_launcher.go:1205 enabled=true lane=true
   1 enterCodexFactory caller=factory_join_discovery_test.go:404 enabled=true lane=true
```

Read: the production call site `codex_launcher.go:1205` (inside `runCodexLaunch`, `if factoryEntry.Enabled`) was reached three times, and the only caller of `runCodexLaunch` that hands it an enabled lane entry is `codex_debug_composition_test.go:56` (`grep -rn "runCodexLaunch(" internal cmd --include='*.go'` prints that line and the one in `runCodex`; the test builds the entry itself with `stagedLaneEntry()`); no run through `runCodex` reached it, including every `-l` and `--lane` launch of `TestCodexLaneEntryStartsRelaunchLane`, `TestSD_AC003_CodexRelaunchPerCard` and the matrix. By construction too: the Codex parse sets `Enabled` only for `-l`/`--lane`, the classification returns the lane entry for exactly that case, and `runCodex` returns into `runCodexFactoryLane` before `runCodexLaunch` for it. Answer: the interactive Codex lane branch is unreachable from `moai codex` after M3 and live only through tests that call `runCodexLaunch` or `enterCodexFactory` directly. Nothing was deleted (not in M3's list); the branch and its fields (`factoryFlagParse.LaneNumber`/`LaneLabel`, which no entry sets now) are left for a later cleanup decision.

Re-measure of `factoryFlagUsageError` (the Codex parse read it at `codex_factory.go:97`): `grep -rn "factoryFlagUsageError" internal --include='*.go'` now prints `factory.go:87` (comment), `factory.go:91` (definition) and `factory.go:264` (the cc/glm `-f <value>` refusal) — the Codex parse no longer reads it, and the Codex `-f <value>` line is the new refusal line from the classification. `grep -rn "leadFlagShort" internal cmd --include='*.go'` prints nothing.

Build and static checks, after the last edit (HEAD `1c3ab47ef` plus only this progress file):

```text
$ gofmt -l <the 13 touched Go files>                      -> no output
$ go build ./...                                          -> no output, exit 0 (BUILD_OK printed)
$ GOOS=windows GOARCH=amd64 go build ./...                -> no output, exit 0 (WINBUILD_OK printed)
$ go vet ./internal/cli/                                  -> no output, exit 0 (VET_OK printed)
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/        -> no output, exit 0 (WINVET_OK printed)
$ golangci-lint run ./internal/cli/   (v2.1.6)            -> 0 issues.
```

(The gofmt, build and vet block ran at HEAD `c0e8ce07e` plus the uncommitted edits that became `69588367a`; the only later source change is the comment in `codex_debug_composition_test.go`, after which `go vet ./internal/cli/` was re-run once, exit 0.)

#### Baseline-attribution

- Tree: HEAD `c0e8ce07e` for the RED measurements (probe and the ten-test selector, before the production change), `69588367a` content for the GREEN, net, regression and build measurements (measured on the working tree before it was committed as `69588367a`; the commit added no other change), `1c3ab47ef` for the last vet. The judging build is the Go toolchain compiling this tree; `golangci-lint` is the installed v2.1.6, used as a check. No installed `moai` binary measured anything except the `moai slot` lease commands.
- The reachability measurement ran against the tree through an overlay of one file; the instrumented copy and its log are in the scratchpad.
- The heavy runs ran under `moai slot acquire --resource t1399-run --max-duration 60m`; `moai slot release --resource t1399-run` printed `slot t1399-run released` after the last heavy run, before commit 1.

#### Gaps

1. The whole `internal/cli` suite was NOT run. The widest sets are the `^TestCodex` family (388 PASS), the 59-test factory/help/vocabulary set, the nine net selectors and the AC selectors above; a regression in a test outside them is unobserved. A first reachability run with the broader selector `Codex|^TestSD_|^TestFactoryNet|...` exceeded the tool's foreground timeout and was moved to the background by the harness; it was stopped with TaskStop, produced no result, and is not cited; the narrower selector above replaced it.
2. No mutant probe was run for M3: a classification that reads `--lane <x>` as `--lane` plus a forwarded value, or refuses `-l -f` but not `-f -l`, was not applied to the codex path. The codex subtests of `TestLaneEntryRefusals` carry both orders and all fifteen shapes, but their discriminating power is argued, not measured.
3. AC-003's codex half asserts one stderr line, exit 1, nothing on stdout, no launch, no lane claim, no new run row and no temp file; a refusal path that wrote some other state is not asserted against (the codex refusal returns before the first read, by construction, read not measured).
4. The composition rows of AC-005 (`--clear-policy`, `--no-auto-dispatch`, `--factory-run`, `-p`, `-w`, `--`) are asserted for cc and glm only (the AC text lists them for cc); on codex `--factory-run` stays refused and `--clear-policy`/`--no-auto-dispatch` are not codex flags (usage failure). That the relaunch loop ignores a `--leader` it was given is read from `runCodexFactoryLane` (it passes an empty leader to `enterSelectedFactoryRun`), not measured.
5. Worktree-guard refusals, each re-issued as a plain command with the same measurement: (a) a `python3` heredoc edit followed by a `grep` in one command, refused as too complex to verify — the edits were made with the edit tool instead; (b) a `go test ... > file 2>&1 && echo RC=0 || echo RC=$?` chain, refused — re-issued without the `|| echo` clause (the tool reported exit 1 for the RED run and `RC=0` for the green ones); (c) an edit through the shared-checkout path of this file was refused by the edit tool and redone on the worktree path. No measurement was replaced by reading source, except the item-4 `--leader` statement.
6. Plain `git` only; no `git -C` was used.

#### Residual-risk

- Until M4 lands, the leader SessionStart notice, the stale-run hint and the factory card rejoin errors still print `moai cc -f lane-<i>` / `-f lane` (M2's residual, unchanged): M2 and M3 must not merge without M4.
- Behavior change on non-lane combinations: `moai codex -f -k` (either order) now prints the kanban line; before, the first token in argument order decided. The lane combinations (`-l -f`, `-l -k`) print the one-entry-token line as AC-003 requires.
- `moai codex -l --factory-run <id>` is still refused (with the `-f` refusal line, as `-f lane --factory-run` was before); the line names the codex lane entry, but its first clause says "no -f entry" for a token the operator did not type.
- The interactive Codex lane branch and its seams (`enterCodexFactory`, `codexFactoryEnv`, `stampCodexLaneClaim`, `codexExplicitFactoryEnv`) are unreachable from the launcher and still compile and test; dropping them is a decision for a later milestone.

#### Findings (statements the tree contradicts, no SPEC file edited)

- `plan.md` M3 and the brief place the interactive lane branch at `codex_factory.go:128-229`; on this tree it lives in `enterCodexFactory` and its siblings (`codex_factory.go`), with the call site in `runCodexLaunch` (`codex_launcher.go:1205` after this milestone's edits). The reachability answer does not depend on the line numbers.
- The brief lists an empty value as a shape for AC-007; `acceptance.md` AC-007 counts exactly five subtests. The empty value (`-f=`, `--factory=`) is asserted in `TestSD_AC004_CodexOtherFactoryShapesRefused`'s shape list (a loop inside one test, not subtests), so the AC's count of five holds.

### M4 evidence

Recorded by the run-phase implementation worker (cycle_type tdd) for milestone M4 of card t1399 (leader notice and notice strings), branch `WT-launcher-entry-flags`. Start state, re-read before any change: `git rev-parse --short HEAD` printed `2ee6aa770`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Output blocks are verbatim excerpts of commands run in this run on this tree; each says what it drops. Full outputs were kept in the session scratchpad (outside the tree).

#### Claim

- The factory leader SessionStart notice states one lane-start sentence (`moai cc -l`, or `moai glm -l` / `moai codex -l`, in a new terminal) in en, ko, ja, zh, carries no lane count, no per-lane launch line, no numbered lane label, no free-slot line; its entry guide names `-f` only (the `-l` entries sit in the lane-start sentence). The notice builder no longer calls `FactoryFreeSlots`; `leaderFreeSlots` and `leaderSlotsNone` are gone from the struct and the four tables; `FactoryFreeSlots` stays in `internal/kanban/factory_slots.go`, caller-less in production (its own two test files still call it).
- The foreman sentence names the factory foreman in four locales (`factory foreman loop`, `팩토리 포어맨`, `ファクトリーフォアマン`, `工厂工头`) and points at no path. The stale-run rebind hint (four locales) names `moai cc -l`. The factory card legacy-label errors (`internal/cli/factory_card.go`) end `(rejoin with -l)`.
- Production files changed: `internal/hook/session_start_factory.go`, `session_start_factory_i18n.go`, `session_stale_run.go`; `internal/cli/factory_card.go`; comment-only: `internal/config/defaults.go`, `internal/config/envkeys.go`, `internal/kanban/bootstrap.go`. Beyond the plan's list, two items orphaned by the removal went with it in the same file: the helper `factoryLaunchEntry` (its only caller was the per-lane launch-line loop) and the `root` parameter of `factoryLeaderNotice` (its only reader was the free-slot call; the two test callers were updated).
- Commit: `8d59545a4` (production, tests, comments; build, vet, gofmt, lint and the net green at that commit), then this record.

#### Evidence

RED first (E8), before any production edit, on HEAD `2ee6aa770` plus the uncommitted test edits. Env unset in the same invocation. Hook tests (10 names: the 5 AC-010 selector tests, the new `TestStaleRunRebindHintNamesLaneEntry`, and four re-pinned neighbours), tool exit 1:

```text
$ unset <MOAI_KANBAN* and MOAI_FACTORY_* axes> && go test ./internal/hook -run '^(TestFactoryLeadNoticePrintsLaneCommandOnce|TestFactoryLeadNoticeIsLaneCountIndependent|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide|TestFactoryGuideNamesWorkerJoinInEveryLocale|TestUnbindNoticeRebindLinePresence|TestStaleRunRebindHintNamesLaneEntry|TestFactoryGuideTeachesLaneFormsInEveryLocale|TestFactoryLeadProviderLaneGuidance|TestFactoryLeadNoticeCarriesDispatchDiscipline|TestFactoryLeadNoticeDispatchDisciplineKorean)$' -count=1
--- FAIL: TestFactoryLeadNoticePrintsLaneCommandOnce (1.84s)
--- FAIL: TestFactoryLeadNoticeIsLaneCountIndependent (27.61s)
--- FAIL: TestStaleRunRebindHintNamesLaneEntry (0.00s)
--- FAIL: TestFactoryLeadProviderLaneGuidance (22.54s)
--- FAIL: TestFactoryGuideNamesWorkerJoinInEveryLocale (0.00s)
--- FAIL: TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide (0.53s)
--- FAIL: TestFactoryLeadNoticeCarriesDispatchDiscipline (0.44s)
--- FAIL: TestFactoryLeadNoticeDispatchDisciplineKorean (0.47s)
--- FAIL: TestUnbindNoticeRebindLinePresence (1.23s)
--- FAIL: TestFactoryGuideTeachesLaneFormsInEveryLocale (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	55.576s
```

Assertion lines of that run, per-locale counts (4 locales each), from `session_start_factory_lane_entry_test.go`: line 65 `"moai cc -l" appears 0 times, want exactly 1` (and the same for `moai glm -l`, `moai codex -l`), line 70 `notice still teaches the removed form "-f lane"` / `"lane-<n>"` / `"lane-1..lane-"`, line 74 `notice carries a numbered lane label`, line 78 `notice carries a per-lane launch line "moai cc -f lane-1"` (also `lane-2`, `lane-3`), line 110 `notice at 8 lanes differs from the notice at 1 lane`, line 117 `notice under launch provider "glm" differs from the default`. The rendered en notice in the failure message carried `This session dispatches cards to 3 lanes over cross-session messages.` and `Lanes are named lane-1..lane-3, and a lane that joins with `-f lane` takes the next free lane-<n>`.

Card errors (cli), tool exit 1:

```text
$ go test ./internal/cli -run '^TestFactoryCardLegacyLabelErrorsNameLaneEntry$' -v -count=1
    factory_card_lane_entry_test.go:29: error does not name the -l rejoin: factory stage: "worker-2" is the legacy lane label; use "lane-2" (rejoin with -f lane)
    factory_card_lane_entry_test.go:35: error still teaches the removed -f lane form: factory stage: "worker-2" is the legacy lane label; use "lane-2" (rejoin with -f lane)
    (same two lines for agent-3, worker, agent)
--- FAIL: TestFactoryCardLegacyLabelErrorsNameLaneEntry (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.368s
```

RED-7 / RED-7b re-observed on the pre-change tree (`grep -n 'leaderFreeSlots\|FactoryFreeSlots' internal/hook/session_start_factory.go` printed `229:` the `FactoryFreeSlots` call and `240:` the `leaderFreeSlots` use; `grep -n 'launch = append' ...` printed `207:`).

GREEN, AC-010 command, on the tree whose content became `8d59545a4`, tool exit 0, swept 5 top-level tests:

```text
$ go test ./internal/hook -run '^(TestFactoryLeadNoticePrintsLaneCommandOnce|TestFactoryLeadNoticeIsLaneCountIndependent|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide|TestFactoryGuideNamesWorkerJoinInEveryLocale|TestUnbindNoticeRebindLinePresence)$' -v -count=1
--- PASS: TestFactoryLeadNoticePrintsLaneCommandOnce (0.00s)
--- PASS: TestFactoryLeadNoticeIsLaneCountIndependent (0.00s)
--- PASS: TestFactoryGuideNamesWorkerJoinInEveryLocale (0.00s)
--- PASS: TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide (0.00s)
--- PASS: TestUnbindNoticeRebindLinePresence (2.41s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	3.346s
```

Notice/stale-run/factory family (selector `Factory|StaleRun|Unbind|RoleNaming|Prescription|Preexisting|LaneSpawn|GatewayGuard`, `-v`, under lease): 52 `--- PASS`, 0 `--- FAIL`, 1 `--- SKIP` (`TestFactoryHookBenchmarkBudget`), `ok github.com/modu-ai/moai-adk/internal/hook 182.338s`.

Net re-run (AC-015 selectors), env unset, tool exit 0 each:

```text
$ go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestPrepareKanbanSettingsWritesTransientFile|TestPrepareFactorySettingsWritesTransientFile|TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch|TestFactoryNetBlockCap|TestFactoryEntryMatrix|TestFactoryCardLegacyLabelErrorsNameLaneEntry)$' -v -count=1
--- PASS: TestFactoryCardLegacyLabelErrorsNameLaneEntry (0.00s)
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (6.64s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (5.53s)
--- PASS: TestFactoryNetLeaderLaunch (2.92s)
--- PASS: TestFactoryNetLaneLaunch (5.84s)
--- PASS: TestFactoryNetBlockCap (2.67s)
--- PASS: TestFactoryEntryMatrix (12.15s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (0.96s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	38.157s
$ go test ./internal/hook -run '^(TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices)$' -v -count=1
--- PASS: TestFactoryNetSessionRecord (0.01s)
--- PASS: TestFactoryNetSessionStartNotices (1.34s)
ok  	github.com/modu-ai/moai-adk/internal/hook	2.094s
$ go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1
--- PASS: TestDiscoverLeaderVerifiesLiveLeader (0.38s)
--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID (0.19s)
ok  	github.com/modu-ai/moai-adk/internal/discovery	0.995s
```

Swept: 8 net tests in cli (the nine PASS lines include the new card-error test; `TestPrepareFactorySettingsWritesTransientFile` is the M7 name and does not exist yet, so the alternation sweeps the old name only), 2 in hook, 2 in discovery. The two hook net tests were not edited (their leader assertions are run id, socket, and the word leader, all still produced).

Removed-field grep, after the change: `grep -rnE 'leaderFreeSlots|leaderSlotsNone' internal cmd --include='*.go' | wc -l` printed `0`; `grep -c 'FactoryFreeSlots' internal/hook/session_start_factory.go` printed `0` (the function remains at `internal/kanban/factory_slots.go:360`, called only from its two test files).

Removed-form sweep of non-test Go (`grep -rnE -e '-f lane|-f N\b|-l <leader>|-f <N>|-f worker|lane-<|-f lane-' internal cmd --include='*.go'`, test files excluded): no line in `internal/hook` teaches a removed form; no `-l <leader>` or `-f <N>` or `-f N` hit anywhere. Remaining hits are `internal/cli/cc.go:87-96` and `glm.go:97-106` (help lines for the `-k` kanban forms, removed at M5a), `internal/cli/kanban.go:53,84` (kanban comments, M5a), and comments that state the refusal or describe the canonical label (`codex_launcher.go:29,770,829`, `factory.go`, `factory_card.go:98,101` where `lane-<n>` names the canonical label, `internal/factorymsg/store.go:399`, `internal/kanban/*`).

Build, vet, format, lint (tool exit 0, empty output): `gofmt -l` over the 14 touched files; `go vet ./internal/hook/ ./internal/cli/ ./internal/config/ ./internal/kanban/`; `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `golangci-lint run ./internal/hook/ ./internal/config/ ./internal/kanban/` printed `0 issues.` (v2.1.6; `internal/cli` was not linted, one file of it changed).

`git diff --stat 2ee6aa770 8d59545a4`: 14 files, 340 insertions, 267 deletions: `internal/cli/factory_card.go`, `factory_card_lane_entry_test.go` (new), `internal/config/defaults.go`, `envkeys.go`, `internal/hook/role_naming_m3_notice_test.go`, `session_stale_run.go`, `session_start_factory.go`, `session_start_factory_i18n.go`, `session_start_factory_lane_entry_test.go` (new), `session_start_factory_provider_test.go`, `session_start_factory_test.go`, `session_start_factory_worker_test.go`, `stale_run_gate_test.go`, `internal/kanban/bootstrap.go`.

Native-idiom check (the `moai-domain-humanize` skill was loaded; its locale modules `korean.md`, `japanese.md`, `chinese.md` were read for their detection tables and applied BY HAND to the strings authored here; no humanize pass was run as a rewrite, and no change-rate or grade was computed). Strings checked: `leaderManual`, `entryGuide`, the foreman clause of `leaderClasses`, and the stale-run rebind hint, in ko, ja, zh. ko: no `~을 통해` / have-verb / pronoun (A), no connective-comma habit (C-11; the first draft's `~없으니,` comma was removed), verb-centric `실행해 시작하세요`, register matches the neighbouring `합니다/하세요` strings. ja: no `することができます` padding (JA-01; the first draft's `起動することはできない` became `起動できない`), no sentence-head connective (JA-03), loanword `ファクトリーフォアマン` follows the neighbouring `ファクトリー` renderings. zh: no `因此` connector (CN-A; the first draft's `因此要启动泳道，请…` became `需要在新终端中运行…来启动泳道`), no 的-nominalization chain. Terms kept as established loanwords or protocol tokens: `포어맨`, `工头`, `run`, `/loop`, the command lines. The en strings were not checked (the policy is conditional on non-English).

#### Baseline-attribution

- Tree: HEAD `2ee6aa770` for the RED measurements (tests edited, production not yet), the working tree that became `8d59545a4` for the GREEN, net, family, build, vet, lint and grep measurements; the idiom edits to the three `leaderManual` strings came after the first GREEN run, so the AC-010 command and the family run were re-run after them (the AC-010 block and the family count above are the re-runs; the net run and the lint run preceded the idiom edits, which touch string literals only and were followed by gofmt and the two re-runs). The judging build is the Go toolchain compiling this tree. No installed `moai` binary measured anything except the `moai slot` lease commands.
- Heavy runs ran under `moai slot acquire --resource t1399-run --max-duration 45m`; `moai slot release --resource t1399-run` printed `slot t1399-run released (was 0dcdf2d5-df5c-4da1-8870-24c2a5861303)` after the last heavy run.

#### Gaps

1. The whole `internal/hook` package suite was NOT completed: one foreground run exceeded the 580 s tool timeout and was moved to the background by the harness; it was stopped with TaskStop and produced no result. The family selector above (52 PASS, 0 FAIL, 182 s) replaced it; a regression in a hook test outside that family is unobserved. The whole `internal/cli` suite was not run either (named selectors only).
2. No mutant probe was run for M4: a notice that prints the count only above one lane, or that keeps the free-slot line behind a condition, is argued to fail `TestFactoryLeadNoticeIsLaneCountIndependent` and `TestFactoryLeadNoticePrintsLaneCommandOnce` (the RED run shows the old shape failing them), not measured on a mutant. The AC-015 mutant (f) (factory notice block removed) was not re-run against the two hook net tests.
3. `internal/cli` was not linted by `golangci-lint` (one line pair changed in `factory_card.go`); `go vet` and the build cover it.
4. The stagger sentence of `leaderStagger` still says `free-slot lanes` / `残りの空きスロットのレーン` / `空闲槽位的泳道` (unchanged text); with the free-slot line gone it no longer has a list to refer to, but it names no number and no removed form, and rewording it was outside the plan's M4 list.
5. A leader whose `MOAI_FACTORY_WORKERS` does not parse to a number of at least 1 still gets no notice (the `lanes < 1` guard is kept, so the count is now only a gate); that behavior is unchanged from before and was not changed.
6. Worktree-guard refusals (all re-issued as plain commands with the same measurement): (a) a `python3` heredoc rewrite of a test file followed by a `grep` in one command, refused as too complex to verify — the edits were redone with the edit tool (one earlier `python3` heredoc that stood alone was accepted and rewrote one test function); (b) the first full-package run was not refused but moved to the background by the harness (see item 1). No measurement was replaced by reading source.
7. Plain `git` only; no `git -C` was used. No `--no-verify`, no push.

#### Residual-risk

- The en `leaderManual` and the three translations state the lane-start sentence once; the three `-l` commands are named by the notice and not by any shared constant, so a future change to the entry grammar has to edit the four tables and the stale-run hint together (the tests pin all three commands in every locale and the hint).
- M2, M3 and M4 now cover the launcher grammar, the Codex entry and every notice string the plan lists; the unit stays unmerged until the lead integrates the three together. Strings that still name `-k` forms (`cc.go`/`glm.go` help, the kanban notice code) belong to M5a/M5b and are untouched.
- The Korean, Japanese and Chinese wording passed a by-hand catalogue read, not a measured humanize grade; a native reviewer may still prefer other phrasing.

#### Findings (statements the tree contradicts, no SPEC file edited)

- `plan.md` M4 (and design.md §3) say the notice is built from `leaderManual` and `entryGuide` with the count; on this tree `entryGuide` took the count as `%[1]d` and the entry token as `%[2]s`, and `leaderManual` took the count twice — both are now plain text. Design §3's line references (`session_start_factory.go:207,229`, `session_start_factory_i18n.go:139,184,229`, `session_stale_run.go:92,103,114,125`, `factory_card.go:95,98,101`) matched the tree before this milestone.
- AC-010's command text names `TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide`, `TestFactoryGuideNamesWorkerJoinInEveryLocale` as "re-pinned"; both exist with those names and were re-pinned, and the five-name selector sweeps 5 top-level tests as the criterion states.

### M5a evidence

Recorded by the run-phase implementation worker (cycle_type tdd) for milestone M5a of card t1399 (launcher removal), branch `WT-launcher-entry-flags`. Start state, re-read before any change: `git rev-parse --short HEAD` printed `81f1f125b`, `git branch --show-current` printed `WT-launcher-entry-flags` (the session started on it), `git status --short` printed nothing. Output blocks are verbatim excerpts of commands run in this run on this tree; each says what it drops. Raw outputs were kept in `/tmp/m5a-*.txt` (outside the tree).

#### RED-now and RED before the implementation (E8)

RED-K1 re-observed on this tree (AC-011's own command, env unset in the same invocation, tool exit 0 — the tests PASS, which is the red: the `-k` launch is accepted):

```text
$ unset <MOAI_KANBAN* and MOAI_FACTORY_* axes> && go test ./internal/cli -run '^(TestCC_KanbanFlagStrippedBeforeLaunch|TestGLM_KanbanFlagParity)$' -v -count=1
=== RUN   TestCC_KanbanFlagStrippedBeforeLaunch
--- PASS: TestCC_KanbanFlagStrippedBeforeLaunch (0.02s)
=== RUN   TestGLM_KanbanFlagParity
--- PASS: TestGLM_KanbanFlagParity (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.151s
```

RED-K2 re-observed (`go run ./cmd/moai codex -k`, stdout 0 bytes, tool exit 1, stderr): `KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead`.

New tests written first (`internal/cli/launcher_retired_entries_test.go`, new; `TestSD_AC003_CodexRelaunchPerCard` in `factory_m5_test.go` re-pinned from "label present" to "label absent"). RED run on HEAD `81f1f125b` plus only those test edits, env unset, tool exit 1, `TestKanbanEntryRefused` swept 21 refusal subtests + 1 passthrough (`go test ... -run '^(TestKanbanEntryRefused|TestCodexLaneChildEnvOmitsLabelMarker|TestFactoryCardVerbsResolveLaneFromWorkerMarker)$' -v -count=1`, tail of the summary):

```text
--- FAIL: TestKanbanEntryRefused (72.96s)
    --- FAIL: TestKanbanEntryRefused/cc_-k ... cc_-k=3          (7 of 7 FAIL)
    --- FAIL: TestKanbanEntryRefused/glm_-k ... glm_-k=3        (7 of 7 FAIL)
    --- FAIL: TestKanbanEntryRefused/codex_-k ... codex_-k=3    (7 of 7 FAIL)
    --- PASS: TestKanbanEntryRefused/passthrough (0.00s)
--- FAIL: TestCodexLaneChildEnvOmitsLabelMarker (10.24s)
--- PASS: TestFactoryCardVerbsResolveLaneFromWorkerMarker (12.36s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	98.003s
```

Assertion lines (cc, glm): `launcher_retired_entries_test.go:51: [-k] launched a session; a refusal launches nothing` and `[-k] was accepted, want a refusal` (the same pair for every one of the 7 shapes on cc and on glm). Assertion lines (codex): `[-k] refusal "KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead" lacks "retired"` (and lacks `-f`, `-l`, `moai codex -l`, `moai cc -f`, `moai glm -f`), repeated for the 7 shapes.

Codex lane label pair, RED run (`-run '^(TestSD_AC003_CodexRelaunchPerCard|TestCodexLaneChildEnvOmitsLabelMarker|TestFactoryCardVerbsResolveLaneFromWorkerMarker)$' -v -count=1`, tool exit 1):

```text
    factory_m5_test.go:146: invocation 0: child env carries the retired MOAI_KANBAN_LABEL="lane-1"
    factory_m5_test.go:146: invocation 1: child env carries the retired MOAI_KANBAN_LABEL="lane-1"
--- FAIL: TestSD_AC003_CodexRelaunchPerCard (13.79s)
    launcher_retired_entries_test.go:102: the Codex lane child carries MOAI_KANBAN_LABEL="lane-1"; the label marker is retired
    launcher_retired_entries_test.go:113: Codex lane child lane-family keys = [MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD MOAI_KANBAN_LABEL], want exactly [MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD]
--- FAIL: TestCodexLaneChildEnvOmitsLabelMarker (7.02s)
--- PASS: TestFactoryCardVerbsResolveLaneFromWorkerMarker (7.72s)
```

`TestFactoryCardVerbsResolveLaneFromWorkerMarker` passes before the change by design: the card verbs already read `MOAI_FACTORY_WORKER` and `MOAI_FACTORY_ROLE` (`factory_card.go:50-52,89`), so the stamp is redundant for them; the test exists to reject the mutant that removes the stamp and also stops publishing the worker marker (AC-013's mutant clause), not to be red now.

#### Claim

- `-k`, `--kanban`, `-k=…`, and every shape of the AC-011 table are refused on `moai cc`, `moai glm` and `moai codex` with one line from `internal/cli/launcher_retired_entries.go` (`retiredLongFlag`, `retiredShortFlag`, `retiredEntryRefusal`, `refuseRetiredEntry`); `parseLauncherEntry` calls it first and no longer calls the deleted parser; the Codex classifier calls it once at the top, so the retired line wins over every other token (`-l -k` included). A `-k` after `--` is forwarded untouched.
- Deleted from `internal/cli`: `parseKanbanFlag`, `enterKanbanMode`, `enterKanbanCompanionMode`, `companionRegistryPath`, `resolveCompanionName`, `parseCompanionLabel`, `rejectKanbanOnCG`, `kanbanBranch` with its three constants, `resolveKanbanBranch`, the four constants `kanbanFlagLong`, `kanbanFlagShort`, `kanbanFlagUsageError`, `kanbanUnsupportedBackendSentinel`, `codexKanbanRefusalDiag`, the kanban branches of `cc.go` and `glm.go` (with the `parseCompanionLabel` call before the dispatch), the kanban clause of `launcher_blockcap_infinite.go` (the factory clause stays), and the Codex lane `MOAI_KANBAN_LABEL` stamp, its child-env line, and the three scrub-list entries `EnvMoaiKanban`, `EnvMoaiKanbanSpec`, `EnvMoaiKanbanLabel` of `codexLaneLaunchEnvKeys`.
- Kept: `exportKanbanLaunchFacts` (signature unchanged, first parameter now `_`); it no longer captures or writes `MOAI_KANBAN_SPEC`. `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL` stay defined in `internal/config/envkeys.go` and read by the hook (M5b). `internal/hook`, `internal/config`, `internal/kanban`, `internal/web` untouched.
- Help text of `cc` and `glm` (Use line, the Kanban Mode block, the `-k <N>` factory item, two examples) no longer teaches `-k`; the genealogy paragraph keeps the history and states that `-k` is retired.
- The comments that named deleted symbols are reworded in `kanban.go`, `factory.go`, `kanban_settings.go`, `cc.go`, `glm.go`, `codex_launcher.go`, `launcher_blockcap_infinite.go`.
- Commits (HEAD at start `81f1f125b`): `7a9c52522` (docs: the RED record above, committed first so the graph orders it ahead of the change), `7afe450f3` (production and tests), then this record.

#### Evidence

AC matrix for the parts M5a flips (each row is a command run in this run on this tree; the detail follows):

| AC | Part | Status | Command | Exit | Swept |
|---|---|---|---|---|---|
| AC-011 | `-k` refused in every shape | PASS | `go test ./internal/cli -run '^TestKanbanEntryRefused$' -v -count=1` | 0 | 1 parent, 21 refusal subtests, 1 passthrough subtest |
| AC-012 | launcher identifiers (M5a share) | PASS | the symbol grep below: 8 files remain, all named by AC-012 for after M5a; `go build ./...` and the windows build | grep 0, builds 0 | 8 files |
| AC-013 | test half | PASS | `go test ./internal/cli -run '^(TestSD_AC003_CodexRelaunchPerCard\|TestCodexLaneChildEnvOmitsLabelMarker\|TestFactoryCardVerbsResolveLaneFromWorkerMarker)$' -v -count=1` | 0 | 3 |
| AC-013 | grep half | not read (M5b) | `grep -rl 'MOAI_KANBAN_LABEL' internal cmd --include='*.go' --exclude='*_test.go'` prints `internal/config/envkeys.go` only | n/a | n/a |
| AC-015 | the net, 8 / 2 / 2 | PASS | the three selectors below | 0 each | 8 cli, 2 hook, 2 discovery |

GREEN of AC-011's command, env unset in the same invocation, tool reported no error (exit 0): `go test ./internal/cli -run '^TestKanbanEntryRefused$' -v -count=1` printed `--- PASS: TestKanbanEntryRefused (34.82s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli	35.917s`; `grep -c '^    --- PASS'` of the output printed `22` (21 refusal subtests, seven shapes times cc, glm, codex, plus `TestKanbanEntryRefused/passthrough`). Refusal line, from `go run ./cmd/moai codex -k` (stdout 0 bytes, tool exit 1): `-k/--kanban is retired: Kanban Mode was removed; start a factory leader with 'moai cc -f' or 'moai glm -f', and join as a lane with -l ('moai codex -l' on Codex)`.

GREEN of AC-013's test command, same conditions: `go test ./internal/cli -run '^(TestSD_AC003_CodexRelaunchPerCard|TestCodexLaneChildEnvOmitsLabelMarker|TestFactoryCardVerbsResolveLaneFromWorkerMarker)$' -v -count=1`:

```text
--- PASS: TestSD_AC003_CodexRelaunchPerCard (7.67s)
--- PASS: TestCodexLaneChildEnvOmitsLabelMarker (4.43s)
--- PASS: TestFactoryCardVerbsResolveLaneFromWorkerMarker (4.53s)
ok  	github.com/modu-ai/moai-adk/internal/cli	17.664s
```

Swept count 3. AC-013's grep half is read at M5b; for the baseline `grep -rl 'MOAI_KANBAN_LABEL' internal cmd --include='*.go' --exclude='*_test.go'` now prints only `internal/config/envkeys.go` (`codex_launcher.go` left), and `grep -c 'EnvMoaiKanbanLabel' internal/cli/codex_launcher.go` prints `0` (RED-8 recorded 4).

AC-012 symbol grep re-measured (RED-K3 was 15 files), `grep -rlE 'enterKanbanMode|enterKanbanCompanionMode|parseKanbanFlag|rejectKanbanOnCG|EnvMoaiKanban\b|EnvMoaiKanbanSpec|CompanionRoles|SplitCompanionLabel|kanbanLeaderNotice|kanbanCompanionNotice|kanbanBootstrapNotice' internal cmd --exclude='*_test.go'` printed 8 files, exactly the list AC-012 states for after M5a: `internal/config/envkeys.go`, `internal/cli/ptycaptest/harness.go`, `internal/kanban/bootstrap.go`, `internal/kanban/role.go`, `internal/hook/session_start_factory.go`, `session_start_kanban.go`, `session_start.go`, `session_start_record.go`. The nine launcher declarations and the constants, over all of `internal cmd` including tests: `grep -rnE '\b(parseKanbanFlag|enterKanbanMode|enterKanbanCompanionMode|companionRegistryPath|resolveCompanionName|parseCompanionLabel|rejectKanbanOnCG|kanbanBranch|kanbanBranchLeader|kanbanBranchCompanion|kanbanBranchNone|resolveKanbanBranch|kanbanFlagLong|kanbanFlagShort|kanbanFlagUsageError|kanbanUnsupportedBackendSentinel|codexKanbanRefusalDiag)\b' internal cmd --include='*.go'` printed one line: `internal/config/envkeys.go:220:	// Set by the launcher when enterKanbanMode classifies a leader, read by the` (a comment in the constant's documentation, `internal/config`, outside this milestone; M5b deletes the constant and its comment).

Build, vet, format, lint, on the committed tree (tool reported no error, empty output, exit 0 each): `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `go vet ./internal/cli/`; `GOOS=windows GOARCH=amd64 go vet ./internal/cli/`; `gofmt -l` over the changed files (the only file printed in `gofmt -l internal/cli/` is `mcp_claude.go`, untouched and pre-existing); `golangci-lint run ./internal/cli/` printed `0 issues.` (v2.1.6, the CI version).

Net re-run (AC-015 selectors), env unset, tool exit 0 each: cli `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestPrepareKanbanSettingsWritesTransientFile|TestPrepareFactorySettingsWritesTransientFile|TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch|TestFactoryNetBlockCap|TestFactoryEntryMatrix)$' -v -count=1`: 8 `--- PASS` (`TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestFactoryNetLeaderLaunch`, `TestFactoryNetLaneLaunch`, `TestFactoryNetBlockCap`, `TestFactoryEntryMatrix`, `TestCCFactoryEntryRecordsFailOpenRunMetadata`, `TestPrepareKanbanSettingsWritesTransientFile`), `ok  	github.com/modu-ai/moai-adk/internal/cli	51.176s` (the M7 name `TestPrepareFactorySettingsWritesTransientFile` does not exist yet); hook `-run '^(TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices)$'`: 2 PASS, `ok … internal/hook	2.088s`; discovery `-run '^(TestDiscoverLeaderVerifiesLiveLeader|TestDiscoverLeaderDeclinesUnparseableRunID)$'`: 2 PASS, `ok … internal/discovery	0.953s`. Swept 8 / 2 / 2.

Mutants (g) and (h) once more, against the post-M5a tree, mutated copies in the scratchpad outside the tree, applied with `go test -overlay <scratch>/m5a/overlay_{g,h}.json ./internal/cli -run '<the 9-name net selector>' -v -count=1`, tool exit 1 each (`grep -rn MUTANT internal/cli` finds only pre-existing comments in unrelated files; `git status --short` listed no mutant edit):

```text
mutant g (the resolveFactoryLaneName call replaced by 'finalLabel, claimErr := factoryLabel, error(nil)' in cc.go and glm.go)
--- FAIL: TestFactoryNetLaneLaunch (8.53s)
    --- FAIL: TestFactoryNetLaneLaunch/cc (4.41s)
    --- FAIL: TestFactoryNetLaneLaunch/glm (4.12s)
    factory_net_m1_test.go:349: registry[lane-1] = ({PID:0 RegisteredAt:}, false), want a live claim under pid 34578
    factory_net_m1_test.go:358: second lane MOAI_FACTORY_WORKER = "lane-1", want "lane-2" (the first claim is live)
    factory_net_m1_test.go:362: registry lost the second claim: map[]
FAIL	github.com/modu-ai/moai-adk/internal/cli	44.383s   (the other seven net tests PASS)

mutant h (the MOAI_KANBAN_BACKEND Setenv replaced by '_ = backend' in exportKanbanLaunchFacts, kanban.go)
    factory_net_m1_test.go:275: MOAI_KANBAN_BACKEND at launch = "", want "claude"
    factory_net_m1_test.go:275: MOAI_KANBAN_BACKEND at launch = "", want "glm"
--- FAIL: TestFactoryNetLeaderLaunch (3.09s)
    factory_net_m1_test.go:338: MOAI_KANBAN_BACKEND at launch = "", want "claude"
    factory_net_m1_test.go:338: MOAI_KANBAN_BACKEND at launch = "", want "glm"
--- FAIL: TestFactoryNetLaneLaunch (8.51s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	49.000s   (the other net tests PASS)
```

Targeted regression sets run under `moai slot acquire --resource t1399-run --max-duration 45m` (released: `slot t1399-run released (was 0dcdf2d5-df5c-4da1-8870-24c2a5861303)`), env unset in each invocation; the first three rows ran before the last three test edits named in the rows below them (`cg_retirement_test.go`, `factory_m5_test.go`, `codex_debug_trace_test.go`), the later rows after:

| Selector | Result |
|---|---|
| named tests of the touched and neighbouring files, block 1 (cc, codex retire, block cap, CG, entry grammar) | 48 top-level PASS, 0 FAIL, `ok … 119.584s` |
| named tests, block 2 (factory_test.go, kanban_*_test.go, glm_test.go, launch facts) | 55 top-level PASS, 0 FAIL, `ok … 16.988s` |
| named tests, block 3 (role refusal, lane entry, SD, net, vocabulary guard, CG retirement) | 45 PASS, 2 FAIL: `TestCGRetirementCompleteEntryShapesAndCounters` (its migrated control launched `cc -k 2`, which is refused now: shape replaced by `--name board-watch` in `cg_retirement_test.go:93`) and `TestSD_AC003_CodexRelaunchPerCard` (my first re-pin asserted key presence; the test blanks the ambient variable with `t.Setenv`, so the child inherits an empty value: re-pinned to assert an empty value) |
| family `^TestCodex` + one of Launch, Debug, Spawn, Direct, Child, Factory, Lane, Entry, Harness | `TestCodexDebugTraceEnvKeysOnly` FAILED once (it pinned that an ambient `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL` sentinel never reaches the child — the scrub-list entries this milestone removes): its sentinel map dropped the three retired keys; re-run `ok … 96.151s`, 0 FAIL |
| families TestFactory[A-H], TestSD_, TestLauncher, TestLaunch | `ok … 458.060s`, 0 FAIL |
| families TestFactory[I-Z] and TestCGRetirementCompleteEntryShapesAndCounters | `ok … 217.235s`, 0 FAIL |

Test files, actual versus the plan's nine (re-measured on this tree after M2-M4): the plan's nine all change — `cc_test.go` (9 tests deleted: four `TestParseKanbanFlag_*`, `TestCC_KanbanFlagStrippedBeforeLaunch`, `TestCC_KanbanWritesNoStateRecord`, `TestCC_KanbanEnvMutationIsRestored`, two `TestEnterKanbanMode_*`), `codex_factory_retire_test.go` (scrub-list pin and `TestCodexKanbanEntryIsRefused` re-pinned to `retiredEntryRefusal`), `factory_test.go` (3 deleted: `TestParseKanbanFlagUnifiedEntry`, `TestParseKanbanFlagPassThroughBoundary`, `TestRejectKanbanOnCGLeavesFactoryForms`; merge, `rejectFactoryOnCG`, help and launch rows re-pinned), `kanban_companion_name_test.go` (deleted, 6 tests), `kanban_dispatch_test.go` (deleted, 4 tests and the helper `clearAllKanbanEnv`, whose caller moved to `clearFactoryTestEnv`), `kanban_lead_name_test.go` (re-pinned onto the factory leader: `enterFactoryLeaderMode`, `parseFactoryLaneLabel`, the lane registry), `kanban_autonomy_test.go` (re-pinned onto `enterFactoryLeaderMode` and `enterFactoryLaneMode`, the two callers of `seedAutonomyTier`), `kanban_bootstrap_test.go` (3 tests deleted, 1 re-pinned to the factory leader: `TestEnterFactoryLeaderModeSetsRunID`), `role_naming_m3_homonym_test.go` (the kanban row dropped). Beyond the plan's nine, each of which failed or would fail on this tree: `glm_test.go` (`TestGLM_KanbanFlagParity` deleted), `kanban_help_test.go` (deleted, 2 tests that assert `-k` help), `kanban_launch_facts_test.go` (SPEC subtests replaced by `TestLaunchFactsPublishNoSpecMarker`; backend subtests stay), `launcher_blockcap_infinite_test.go` (the kanban-clause tests replaced by `TestRetiredChainSignalsDoNotRaiseBlockCap` and factory re-pins), `codex_debug_trace_test.go`, `cg_retirement_test.go`, `factory_role_refusal_m2_test.go` (the `-k` rows), `lane_entry_m2_test.go` (the two `-k` rows now expect the retired line), `factory_m5_test.go` (`TestSD_AC003_CodexRelaunchPerCard`), and the new `launcher_retired_entries_test.go`.

`git diff --stat 81f1f125b HEAD` (taken before this record): 28 files changed, 753 insertions(+), 1930 deletions(-); production: `cc.go`, `glm.go`, `kanban.go`, `factory.go`, `codex_launcher.go`, `kanban_settings.go`, `launcher_blockcap_infinite.go`, new `launcher_retired_entries.go`.

#### Baseline-attribution

- Trees: HEAD `81f1f125b` plus the uncommitted new tests for the RED measurements; the working tree that became `7afe450f3` for every GREEN, grep, build, vet, lint, net, mutant and regression measurement (HEAD stayed `81f1f125b` through every measurement; the two commits `7a9c52522` and `7afe450f3` were made after the last one, and `git diff --stat` was taken after them). The judging build is the Go toolchain compiling this tree (`go1.26.8` per the lint banner); `golangci-lint` v2.1.6 built with go1.26.8; no installed `moai` binary measured any test (it provided only the `moai slot` lease commands). The committed probe's M5a patch was read as a model and not applied (`patch` was not run); where it differed from the tree, the tree won: the Codex classifier (a pre-scan, not the probe's inline case, because the M3 classifier has the `kanbanSeen` merge branch the probe tree lacks), the retired line's text, and the test set above.

#### Gaps

1. The whole `internal/cli` suite (4,635 tests) was NOT run: the named blocks and families above ran, two broader family runs hit the 10-minute test timeout without a verdict (`^(TestFactory|TestCodex(…)|TestSD_|TestLauncher|TestLaunch|TestCG…)` at the default `10m`, one moved to the background by the harness and ended `panic: test timed out`; a `9m` re-run of `^(TestFactory|TestSD_|TestLauncher|TestLaunch|…)` ended the same way), and were replaced by the narrower selectors in the table; a regression in an `internal/cli` test outside those families, and in any other package's tests, is unobserved. `internal/hook` and `internal/discovery` ran only their net selectors.
2. The probe's M5a stage (`probe.go … -to M5a`, many full-tree builds) was not re-run for comparison; the compile proof above is `go build ./...` and `go vet ./internal/cli/` on host and windows of the committed tree.
3. Mutants (g) and (h) only; (a)-(f) were not re-run after M5a (the plan asks for (g) and (h)).
4. Exit codes of passing commands were read from the tool (no error banner), not from `echo $?`: the worktree guard refuses `; echo rc=$?` chains.
5. Worktree-guard refusals, each re-issued as plain commands with the same measurement and nothing replaced by reading source: (a) `sed -n` with a shell variable as the file (`P=…`) refused ("value computed at runtime"); (b) the first RED command as `go test … | tail; echo rc=${pipestatus[1]}` refused as too complex, re-run with output redirected to a file; (c) a `for` loop over symbol names with a variable refused, replaced by one `grep -E` alternation; (d) a heredoc followed by `gofmt`/`go vet` in one command refused, the content was added with the edit tool; (e) `which golangci-lint …; for s in …` refused. One broad run was moved to the background by the harness (item 1), not by me; no polling job was started.
6. Ordering (verification-claim-integrity §2.3): the new tests and the production change share commit `7afe450f3`, so the commit graph cannot witness that the tests were written before the change; the RED output is committed ahead of it (`7a9c52522`) and was measured with the new tests present and production not yet edited, which only the session record establishes.
7. The first `go vet`/`gofmt` runs listed compile errors in test files that the plan's "nine files" did not name (`launcher_blockcap_infinite_test.go`, `glm_test.go`, `kanban_help_test.go`, and so on): fixed as listed above, not left.

#### Residual-risk

- Removing `EnvMoaiKanban`, `EnvMoaiKanbanSpec`, `EnvMoaiKanbanLabel` from `codexLaneLaunchEnvKeys` (the plan's "scrub entries", and what makes RED-8's count 0) means a Codex lane launched from a shell that still exports one of them with a value passes it to the child: `TestCodexDebugTraceEnvKeysOnly` had pinned the opposite for those three, and its sentinel map no longer names them. The hook still reads `MOAI_KANBAN` and `MOAI_KANBAN_LABEL` until M5b, so in the window between M5a and M5b a stale kanban session's exports could reach a Codex child's hook; after M5b nothing reads them. The shipped child key-set (`TestCodexLaneChildEnvOmitsLabelMarker`) holds only for an environment without a non-empty ambient label.
- `parseLauncherEntry`'s struct keeps three fields no entry sets any more (`KanbanEnabled`, `Spec`, `FactoryLanesDeclared`), the function `rejectFactoryOnCG` has no production caller (it was already uncalled once `cg` retired), and `config.DefaultFactoryLanes` lost its only reader (the deleted `-k --name lane-<n>` default); none is in the plan's deletion list and the SPEC authorizes exactly the listed deletions, so they stay for M7 (`KanbanEnabled` is named there) and for a later card (the rest). `golangci-lint` reports none of them.
- `exportKanbanLaunchFacts(_, backend)` and `exportFactoryLaunchFacts(specID, backend)` keep a parameter nothing reads, so the eight call sites keep their shape until M7 collapses the two functions.
- The retired line names `moai codex -l` and `moai cc -f` / `moai glm -f` for every verb; on cc and glm it also mentions the Codex form, which is harmless but not minimal.

#### Findings (statements the tree contradicts, no SPEC file edited)

- AC-003 lists `-l -k` and `--lane -k` as combination refusals that name the one-entry-token rule; AC-011 and plan M5a make the retired refusal run before any branch. On the post-M5a tree both shapes print the retired line (cc, glm and codex), so `lane_entry_m2_test.go`'s two rows changed from kind `tokens` to a new kind `retired`; the entry-token rule still covers `-l -f` and `-f -l`.
- Plan M5b lists `launcher_blockcap_infinite_test.go`, `codex_debug_trace_test.go`, `kanban_launch_facts_test.go` and `cc_test.go` as files that wait for M5b to retarget removed constants. On this tree some of their assertions fail or no longer compile at M5a (the block-cap kanban clause, the three ambient sentinels, the SPEC export, the deleted-symbol calls and `-k` launches in `cc_test.go`) and were re-pinned or deleted here; the M5b retarget of what remains in them is unchanged.
- The probe's M5a patch edits `codex_launcher.go` with an inline `-k` case that returns at the first token; the M3 classifier on this tree also folds `-k` into `kanbanSeen`/`other` for the entry-token conflict, so the equivalent edit is a pre-scan at the top of `codexFactoryEntryClassify`.
- `internal/config/envkeys.go:220` still documents `MOAI_KANBAN` as "set by the launcher when enterKanbanMode classifies a leader" (the function is gone); the comment belongs to the constant M5b deletes.
- Kanban-named symbols that remain in `internal/cli` after M5a, and why: `kanban.go` (file name, `kanbanEntryParse`, `exportKanbanLaunchFacts`, `leaderRunID`, `claimName`, the leader name helpers: all read by the factory), `kanban_settings.go` (`prepareKanbanSettings`, `writeTransientSettingsFile(…, "moai-kanban")`: the factory settings injection), `kanban_launch_facts_test.go`, `kanban_lead_name_test.go`, `kanban_autonomy_test.go`, `kanban_bootstrap_test.go`, `kanban_settings_test.go` (factory tests under old file names), every `config.EnvMoaiKanban*` marker use (`MOAI_KANBAN_ID`, `_LEAD_ADDR`, `_LEAD_NAME`, `_SETTINGS_INJECTED`, `_BACKEND`, `_CARD` are the frozen factory markers), the three retired constants in `ptycaptest/harness.go` and in test scrub lists (M5b), and the `kanban` package import in `cc.go`, `glm.go`, `factory.go`, `codex_launcher.go` and the todo, doctor and factory files (M6/M8). All rename work is M7.

#### Correction after verification: `TestLaneMarkerGoldenMatchesFLane` was red at `e3900c6af`

The Claim and Evidence above say the M5a ACs and the net were green at `e3900c6af`. That was not established for every test: `TestLaneMarkerGoldenMatchesFLane` (AC-004's golden, in `lane_entry_golden_m1_test.go`) FAILED there, in an env-scrubbed run `go test ./internal/cli -run '^TestLaneMarkerGoldenMatchesFLane$' -count=1 -v` (found by the lane orchestrator, exit 1). Row codex: got `… MOAI_KANBAN_BACKEND=gpt MOAI_KANBAN_CARD=t1`, want `… MOAI_KANBAN_CARD=t1 MOAI_KANBAN_LABEL=lane-1`. Cause: the planned removal of the Codex lane label stamp (REQ-012) against a golden that keeps the key as captured. I missed it because the file was in none of my selectors: the regression selectors were name lists and prefix families, and `TestLaneMarkerGoldenMatchesFLane` matched none of them (`TestLaneEntryEnvParity`, in the sibling file, already excluded the key and passed). The statements above that name the AC-015 net, AC-011 and AC-013 results stand (each was run); the statement that nothing else turned red does not.

Fix (commit `14c69db0b`): `lane_entry_golden_m1_test.go` compares each row through `withoutRetiredLabel`, which drops exactly the key `MOAI_KANBAN_LABEL` from both sides (the literal `retiredLaneLabelMarker`), with a comment naming REQ-012 and AC-004 / REQ-003; every other key, present or absent, still has to match. `git diff --stat -- internal/cli/testdata` printed nothing, so the golden is byte-unchanged. `TestLaneEntryEnvParity` already carried the same exclusion and needed no edit.

Mutant proof that the exclusion is not wider than that key: a copy of `codex_launcher.go` outside the tree (scratchpad), with the `MOAI_KANBAN_BACKEND` entry removed from `codexCardLaunchEnv`, applied with `go test -overlay <scratch>/m5a/overlay_golden.json ./internal/cli -run '^TestLaneMarkerGoldenMatchesFLane$' -count=1 -v`, tool exit 1:

```text
    lane_entry_golden_m1_test.go:167: row "codex" differs from the golden
         got: MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_FACTORY_CLEAR_POLICY= MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-1 MOAI_KANBAN_CARD=t1
        want: MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_FACTORY_CLEAR_POLICY= MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-1 MOAI_KANBAN_BACKEND=gpt MOAI_KANBAN_CARD=t1
--- FAIL: TestLaneMarkerGoldenMatchesFLane (6.45s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/cc (1.20s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/glm (1.22s)
    --- PASS: TestLaneMarkerGoldenMatchesFLane/codex (4.02s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	7.678s
```

Re-run on the fixed tree, env unset in one compound invocation each, no error banner (exit 0): `TestLaneMarkerGoldenMatchesFLane` PASS (3 subtests cc, glm, codex; `ok … 7.122s`); `TestLaneEntryEnvParity` PASS (`ok … 23.851s`); `TestKanbanEntryRefused` PASS (22 subtests, `ok … 34.507s`); `TestLaneEntryRefusals` PASS (45 subtests, 0 FAIL, `ok … 56.496s`); the AC-015 cli net selector 8 PASS (`ok … 30.817s`). Other tests that name the golden, `MOAI_KANBAN_LABEL` or the Codex lane child env were grepped and run: `launcher_characterization_m1_test.go`, `update_version_downgrade_test.go`, `todo_axisa_guard_test.go`, `lane_entry_m2_test.go`, `lane_entry_m3_test.go`, `codex_launcher_test.go` (its `TestCodexVerbRouting_*`, `TestCodexApp_*`, `TestCodexCommand_*`, `TestLaunchers_BareInvocationConvention`), `factory_net_m1_test.go`, `factory_m4_test.go` (the last two already in earlier runs): `ok … 83.534s`, no FAIL. No other red found. `gofmt -l` on the touched file prints nothing; `go vet ./internal/cli/` host and `GOOS=windows`, `go build ./...` host and `GOOS=windows GOARCH=amd64` printed nothing (exit 0).

Gaps of this correction: the whole `internal/cli` suite is still unrun (Gap 1 above stands), so another test outside every selector could still be red; the golden test's `-count=1` run was a single run, not repeated for flakiness. Worktree-guard refusals in this round: none.

### M5b evidence

Recorded by the run-phase implementation worker (cycle_type tdd) for milestone M5b of card t1399 (hook removal and the end of the three markers), branch `WT-launcher-entry-flags`. Start state, re-read before any change: `pwd` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1399`, `git rev-parse --short HEAD` printed `7e8087f40`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Output blocks are verbatim excerpts of commands run in this run on this tree; each says what it drops. Raw outputs were kept in `/tmp/t1399-m5b-*.txt` (outside the tree).

#### RED-now and RED before the implementation (E8)

RED-now of the M5b-flipping greps on tree `7e8087f40` (read before any edit; the line `exit=` is the tool exit of the command):

```text
$ grep -rl 'MOAI_KANBAN_LABEL' internal cmd --include='*.go' --exclude='*_test.go'      (AC-013 grep half; RED-8b)
internal/config/envkeys.go
exit=0
$ grep -c 'kanbanBootstrapNotice' internal/hook/session_start.go                          (AC-014; RED-K5)
2
exit=0
$ find internal/hook -name 'session_start_kanban*'                                        (AC-014)
internal/hook/session_start_kanban_test.go
internal/hook/session_start_kanban_i18n_test.go
internal/hook/session_start_kanban_todo_test.go
internal/hook/session_start_kanban.go
internal/hook/session_start_kanban_surface_test.go
internal/hook/session_start_kanban_i18n.go
$ grep -rlE 'enterKanbanMode|enterKanbanCompanionMode|parseKanbanFlag|rejectKanbanOnCG|EnvMoaiKanban\b|EnvMoaiKanbanSpec|CompanionRoles|SplitCompanionLabel|kanbanLeaderNotice|kanbanCompanionNotice|kanbanBootstrapNotice' internal cmd --exclude='*_test.go'     (AC-012 symbol grep; RED-K3)
internal/config/envkeys.go
internal/cli/ptycaptest/harness.go
internal/kanban/role.go
internal/kanban/bootstrap.go
internal/hook/session_start_factory.go
internal/hook/session_start.go
internal/hook/session_start_record.go
internal/hook/session_start_kanban.go
exit=0
```

Step 1 (the helper moves) was done and committed first, as its own commit `c10194abc` (`langEnglish` and `operatorLang` into `internal/hook/session_start_lang.go`, `clearKanbanEnv` into `internal/hook/session_start_env_helper_test.go`; text moved, nothing else changed). On that tree, env unset in one compound invocation each, tool exit 0: `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, `GOOS=windows GOARCH=amd64 go vet ./internal/hook/`, `go vet ./internal/hook/` (empty output), and the net: hook `--- PASS: TestFactoryNetSessionRecord`, `--- PASS: TestFactoryNetSessionStartNotices`; cli 8 PASS (`TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestFactoryNetLeaderLaunch`, `TestFactoryNetLaneLaunch`, `TestFactoryNetBlockCap`, `TestFactoryEntryMatrix`, `TestCCFactoryEntryRecordsFailOpenRunMetadata`, `TestPrepareKanbanSettingsWritesTransientFile`), `ok … internal/cli 32.713s`; discovery `--- PASS: TestDiscoverLeaderVerifiesLiveLeader`, `--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID`.

New and re-pinned tests written first: `internal/hook/session_start_no_kanban_notice_test.go` (new: `TestSessionStartEmitsNoKanbanNotice`, `TestSessionRecordIgnoresRetiredKanbanMarkers`) and `internal/hook/factory_net_m1_test.go` (`TestPreexistingKanbanArtifactsTolerated` gains the notice-absence and no-record assertions). RED run on `c10194abc` plus only those test edits, env unset, tool exit 1 (`go test ./internal/hook -run '^(TestSessionStartEmitsNoKanbanNotice|TestSessionRecordIgnoresRetiredKanbanMarkers|TestPreexistingKanbanArtifactsTolerated|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide)$' -v -count=1`, assertion lines and verdict lines only):

```text
    factory_net_m1_test.go:234: session old-plan received a kanban notice on additionalContext:
    factory_net_m1_test.go:234: session old-plan received a kanban notice on systemMessage:
    factory_net_m1_test.go:234: session new-session received a kanban notice on additionalContext:
    factory_net_m1_test.go:234: session new-session received a kanban notice on systemMessage:
    factory_net_m1_test.go:241: a session carrying only the retired markers wrote a session record
--- FAIL: TestPreexistingKanbanArtifactsTolerated (1.80s)
--- PASS: TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide (0.00s)
--- FAIL: TestSessionStartEmitsNoKanbanNotice (6.63s)
    --- PASS: TestSessionStartEmitsNoKanbanNotice/no_marker (1.22s)
    --- PASS: TestSessionStartEmitsNoKanbanNotice/leader_marker_alone (0.99s)
    --- FAIL: TestSessionStartEmitsNoKanbanNotice/companion_label_alone (0.99s)
    --- FAIL: TestSessionStartEmitsNoKanbanNotice/leader_marker_with_run_id,_socket,_and_SPEC (1.05s)
    --- FAIL: TestSessionStartEmitsNoKanbanNotice/both_markers (0.61s)
    --- PASS: TestSessionStartEmitsNoKanbanNotice/factory_leader_still_receives_its_notice (0.85s)
    --- PASS: TestSessionStartEmitsNoKanbanNotice/factory_lane_still_receives_its_notice (0.92s)
    session_start_no_kanban_notice_test.go:114: kanbanRoleFromEnv with map[MOAI_KANBAN:1] = ("leader", 0, true), want ok=false
    session_start_no_kanban_notice_test.go:114: kanbanRoleFromEnv with map[MOAI_KANBAN_LABEL:plan] = ("plan", 0, true), want ok=false
    session_start_no_kanban_notice_test.go:114: kanbanRoleFromEnv with map[MOAI_KANBAN_LABEL:sync-2] = ("sync", 0, true), want ok=false
    session_start_no_kanban_notice_test.go:136: the record SPEC field = "SPEC-RETIRED-001", want the empty string
--- FAIL: TestSessionRecordIgnoresRetiredKanbanMarkers (0.02s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	9.499s
```

Why the three passing subtests pass now (read, not hidden): `no_marker` is the control that the assertion does not fire on a clean session; `leader_marker_alone` (`MOAI_KANBAN=1` with no run id) passes because the leader notice builder returns the empty string for an empty run id, so the notice test cannot see that marker alone (the record test does: `kanbanRoleFromEnv` returns `("leader", 0, true)` for it); the two factory subtests are the unchanged-notice half of AC-014. A first run of the same tests reported a false red on `no_marker`: the session id contained the word the assertion greps for, and the attribution line of the context echoes the session id; the ids were renamed (`retired-marker-…`) and the RED above is the second run.

#### Claim

- SessionStart emits no kanban notice. Deleted: `internal/hook/session_start_kanban.go`, `session_start_kanban_i18n.go` and their four tests (`session_start_kanban_i18n_test.go`, `_surface_test.go`, `_test.go`, `_todo_test.go`); the notice block and the `kanban_notice` stage lap in `session_start.go` (the factory block stays, and the lap that follows the factory rule block is kept under its existing name `factory_notice`); the companion and kanban-leader branches of `kanbanRoleFromEnv` and its `MOAI_KANBAN_SPEC` read in `session_start_record.go` (the record's SPEC field is the empty string); the three constants `EnvMoaiKanban`, `EnvMoaiKanbanSpec`, `EnvMoaiKanbanLabel` from `internal/config/envkeys.go` and their entries in `kanbanVars` of `internal/cli/ptycaptest/harness.go`.
- `langEnglish` and `operatorLang` moved to `internal/hook/session_start_lang.go` and `clearKanbanEnv` to `session_start_env_helper_test.go` first, as their own commit `c10194abc`, with the build, the vet, and the net green on that tree. `TestOperatorLangFailsOpen` and its helper `configWithLang` moved to `session_start_lang_test.go` because `operatorLang` is retained.
- Frozen and untouched: the six factory-read marker constants (`MOAI_KANBAN_ID`, `_LEAD_ADDR`, `_LEAD_NAME`, `_SETTINGS_INJECTED`, `_BACKEND`, `_CARD`), the factory notice block, `internal/kanban`, `internal/web`.
- What `TestPreexistingKanbanArtifactsTolerated` (hook half of AC-017) pins now: the pre-existing `plan` record is left byte-identical and stays readable, the hook returns no error, and (new at M5b) neither channel carries the word "kanban" for the sessions `old-plan` and `new-session` under `MOAI_KANBAN=1` and `MOAI_KANBAN_LABEL=plan`, and the session with a fresh id writes NO record (before M5b it wrote a companion record, because the role reader mapped the label to the role `plan`).
- Commits (HEAD at start `7e8087f40`): `c10194abc` (helper moves), `20b33e997` (docs: RED-now and RED, committed ahead of the change), `1d0c64ef4` (deletions, constants, tests), then this record as the fourth.

#### Evidence

AC matrix, each row a command run in this run on the committed tree `1d0c64ef4` (exit is the shell exit read with `echo`, or the tool's no-error report for `go test`; env unset in one compound invocation for every `go test`):

| AC | Part | Status | Command | Exit | Swept / observed |
|---|---|---|---|---|---|
| AC-012 | M5b share: hook identifiers and the three constants | PASS | `grep -rlE 'enterKanbanMode\|enterKanbanCompanionMode\|parseKanbanFlag\|rejectKanbanOnCG\|EnvMoaiKanban\b\|EnvMoaiKanbanSpec\|CompanionRoles\|SplitCompanionLabel\|kanbanLeaderNotice\|kanbanCompanionNotice\|kanbanBootstrapNotice' internal cmd --exclude='*_test.go'` | 0 | 2 files: `internal/kanban/role.go`, `internal/kanban/bootstrap.go` (exactly the two AC-012 names for after M5b; empty only after M6) |
| AC-012 | build | PASS | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | 0, 0 | no output |
| AC-013 | grep half | PASS | `grep -rl 'MOAI_KANBAN_LABEL' internal cmd --include='*.go' --exclude='*_test.go'` | 1 | empty (RED-now listed `internal/config/envkeys.go`) |
| AC-013 | test half (re-read) | PASS | `go test ./internal/cli -run '^(TestSD_AC003_CodexRelaunchPerCard\|TestCodexLaneChildEnvOmitsLabelMarker\|TestFactoryCardVerbsResolveLaneFromWorkerMarker)$' -v -count=1` (run with `TestPreexistingKanbanArtifactsTolerated`) | 0 | 3 PASS of 3 (+1 PASS), `ok … internal/cli 12.440s` |
| AC-014 | notice call gone | PASS | `grep -c 'kanbanBootstrapNotice' internal/hook/session_start.go` | 1 | `0` (RED-now `2`) |
| AC-014 | files gone | PASS | `find internal/hook -name 'session_start_kanban*'` | 0 | empty (RED-now six paths) |
| AC-014 | tests | PASS | `go test ./internal/hook -run '^(TestSessionStartEmitsNoKanbanNotice\|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide\|TestFactoryNetSessionRecord\|TestFactoryNetSessionStartNotices\|TestSessionRecordIgnoresRetiredKanbanMarkers\|TestPreexistingKanbanArtifactsTolerated)$' -v -count=1` (the AC's two names plus four) | 0 | 6 PASS of 6; both AC-014 names among them; the new notice test passes its 7 subtests (block below) |
| AC-016 | frozen values | PASS | `go test ./internal/config -run '^TestFactoryMarkerValuesFrozen$' -v -count=1` | 0 | 1 PASS |
| AC-016 | discovery | PASS | `go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader\|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1` | 0 | 2 PASS |
| AC-016 | Codex allowlist | PASS | `go test ./internal/codexwiring -run '^TestMCPServerEnvVarsKeepFactoryMarkers$' -v -count=1` | 0 | 1 PASS |
| AC-016 | legacy state dir | PASS | `go test ./internal/kanban -run '^TestLegacyStateDirStillRead$' -v -count=1` | 0 | 1 PASS |
| AC-017 | tolerance | PASS | `go test ./internal/cli ./internal/hook ./internal/web ./internal/statusline -run '^TestPreexistingKanbanArtifactsTolerated$' -v -count=1` (cli run in one invocation, the other three in a second) | 0, 0 | 1 PASS in each of the four packages |
| AC-015 | the net | PASS | the three selectors | 0 each | 8 cli, 2 hook, 2 discovery (block below) |

Net re-run on the committed tree `1d0c64ef4`, env unset, tool exit 0 each:

```text
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (4.16s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (3.49s)
--- PASS: TestFactoryNetLeaderLaunch (1.59s)
--- PASS: TestFactoryNetLaneLaunch (3.51s)
--- PASS: TestFactoryNetBlockCap (1.54s)
--- PASS: TestFactoryEntryMatrix (5.19s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (0.52s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	21.045s            (8 of 8; the second settings-test name does not exist until M7)
--- PASS: TestFactoryNetSessionRecord (0.00s)
--- PASS: TestFactoryNetSessionStartNotices (0.71s)
--- PASS: TestPreexistingKanbanArtifactsTolerated (0.41s)
--- PASS: TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide (0.00s)
--- PASS: TestSessionStartEmitsNoKanbanNotice (1.44s)
--- PASS: TestSessionRecordIgnoresRetiredKanbanMarkers (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	3.223s              (the hook net is the first two names; the other four are AC-014/AC-017's)
--- PASS: TestDiscoverLeaderVerifiesLiveLeader (0.14s)
--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID (0.11s)
ok  	github.com/modu-ai/moai-adk/internal/discovery	0.571s         (2 of 2)
```

The net was also green at commit 1 (`c10194abc`, block in the RED section above) and, for the same three selectors, on the tree of commit 3 before its three comment-only edits (cli `ok … 21.762s`, hook `ok … 1.389s`, discovery `ok … 0.688s`).

Mutants (e) and (f), the two the net must reject because M5b edits their neighbourhoods. Mutated copies kept outside the tree (scratchpad `m5b/`), applied with `go test -overlay <scratch>/m5b/overlay_{e,f}.json ./internal/hook -run '^(TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices)$' -v -count=1`, tool exit 1 each:

```text
mutant e (kanbanRoleFromEnv returns ("", 0, false) for a lane label and for the fan-out marker, session_start_record.go)
    factory_net_m1_test.go:68: the lane wrote no session record: read kanban record: open …/001/.moai/state/todo/net-lane-sess.json
    factory_net_m1_test.go:91: the factory leader wrote no session record: read kanban record: open …
--- FAIL: TestFactoryNetSessionRecord (0.00s)
    --- FAIL: TestFactoryNetSessionRecord/lane (0.00s)
    --- FAIL: TestFactoryNetSessionRecord/leader_signalled_by_the_factory_variable_alone (0.00s)
    --- PASS: TestFactoryNetSessionRecord/ordinary_session_writes_nothing (0.00s)
--- PASS: TestFactoryNetSessionStartNotices (0.60s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.196s

mutant f (the factory notice block's condition made `false && notice != ""`, session_start.go)
--- PASS: TestFactoryNetSessionRecord (0.00s)
    factory_net_m1_test.go:153: leader notice on additionalContext lacks "netrun01":
    factory_net_m1_test.go:153: leader notice on additionalContext lacks "/tmp/moai-socket-factory/netrun01":
    factory_net_m1_test.go:173: lane notice on additionalContext does not name the lane label "lane-2":
    (and the same lines on systemMessage)
--- FAIL: TestFactoryNetSessionStartNotices (0.74s)
    --- FAIL: TestFactoryNetSessionStartNotices/leader (0.27s)
    --- FAIL: TestFactoryNetSessionStartNotices/lane (0.26s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.373s
```

Each mutant is rejected by exactly the net test that guards its neighbourhood, and the other net test stays green, so neither red is a general breakage.

GREEN of the new tests, same conditions as the RED run above (the notice test, the record test, and the AC-017 hook half; `exit=0`, `ok … internal/hook 3.939s`): `TestSessionStartEmitsNoKanbanNotice` PASS with its seven subtests (`no_marker`, `leader_marker_alone`, `companion_label_alone`, `leader_marker_with_run_id,_socket,_and_SPEC`, `both_markers`, `factory_leader_still_receives_its_notice`, `factory_lane_still_receives_its_notice`), `TestSessionRecordIgnoresRetiredKanbanMarkers` PASS with four (`leader_marker`, `companion_label`, `companion_bumped`, `factory_leader_record_carries_no_SPEC`), `TestPreexistingKanbanArtifactsTolerated` PASS.

Suites (under `moai slot acquire --resource t1399-run --max-duration 60m`, released: `slot t1399-run released (was 0dcdf2d5-df5c-4da1-8870-24c2a5861303)`), env unset in each invocation:

| Scope | Result |
|---|---|
| whole `internal/config`, `go test ./internal/config -v -count=1` | exit 0, `ok … 3.927s`, 488 top-level PASS, 0 FAIL, 0 SKIP |
| whole `internal/hook`, first attempt, `go test ./internal/hook -count=1 -timeout 540s` | exit 1: `panic: test timed out after 9m0s`, running `TestSyncGateFailState_AC007_NoPathLooserThanToday`; before the timeout 2 FAIL lines: `TestNoNewEnvLiteralsInDiff` (below) and `TestInboundClaimIndependentOfEnvLabel` (`stale_run_gate_test.go:247: claim delivery altered under a stale env label: msg="" state=degraded: context deadline exceeded`) |
| `TestInboundClaimIndependentOfEnvLabel` alone | PASS (`--- PASS … (0.66s)`); it did not fail in the second whole run either: a load-sensitive deadline, not a result of this change |
| whole `internal/hook` except the 14 `TestSyncGate*` functions, `go test ./internal/hook -skip '^TestSyncGate' -v -count=1 -timeout 560s` | exit 1, `FAIL … 356.614s`: 1253 top-level PASS, 2 FAIL, 6 SKIP; the two FAIL are `TestNoNewEnvLiteralsInDiff` and `TestHookWrapperCopiesStayIdentical` (both pre-existing, below) |
| cli families by file-derived names: the 179 distinct `Test` functions of 24 files (the 7 `internal/cli` files edited here, the 12 further files that call `clearFactoryTestEnv` and so depend on the list this milestone edited — `codex_factory_retire_test.go`, `glm_task_test.go`, `kanban_autonomy_test.go`, `kanban_bootstrap_test.go`, `factory_mixed_test.go`, `factory_join_discovery_test.go`, `launch_session_pid_exec_posix_test.go`, `codex_worktree_anchor_test.go`, `kanban_lead_name_test.go`, `factory_role_refusal_m2_test.go`, `codex_debug_composition_test.go`, `mcp_served_model_test.go` — and the plan's five others that exist, `cc_test.go`, `codex_debug_trace_test.go`, `factory_m5_test.go`, `glm_test.go`, `kanban_launch_facts_test.go`), in three `-run '^(A|B|…)$'` invocations of 60, 60, 59 names | chunk 1: exit 0, 60 PASS, `ok … 56.616s`; chunk 2: exit 0, 60 PASS, `ok … 82.823s`; chunk 3: exit 0, 58 PASS 1 SKIP (`TestSessionPIDStampExecHelper`, a helper process by design), `ok … 127.076s` |
| `go vet ./...` (host, whole module) | exit 0, 0 lines |

The two pre-existing reds, attributed:

- `TestNoNewEnvLiteralsInDiff` (REQ-SRL-009's guard, `env_literal_diff_test.go`) sweeps `git diff <merge-base with develop>` over `internal/hook`, `internal/factorymsg`, `internal/cli` non-test files and fails on any `MOAI_FACTORY*` / `MOAI_KANBAN*` literal in an added line. Its base is `a6d3e6fd4` (the base this branch was cut from), so it measures the whole branch, not M5b. Measured: `git diff a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 HEAD -- internal/hook internal/factorymsg internal/cli ':!*envkeys.go' ':!*_test.go'` at the tree of commit 2 (`20b33e997`, before any M5b production edit) lists the same ten literals with the same counts as the working tree of commit 3 (`MOAI_FACTORY_AUTO_DISPATCH` 3, `_CLEAR_POLICY` 3, `_ROLE` 3, `_WORKER` 5, `_WORKERS` 2, `MOAI_KANBAN_BACKEND` 3, `_CARD` 1, `_ID` 3, `_LABEL` 1, `_SETTINGS_INJECTED` 2). M5b added none.
- `TestHookWrapperCopiesStayIdentical`: `wrapper_copies_contract_test.go:92: internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh differs from .claude/hooks/moai/sync-phase-quality-gate.sh`. Neither file is in `git diff --stat 7e8087f40 HEAD`.

Edited, deleted, and added files against the plan's list (re-measured: `go vet -gcflags=-e` of the four packages on the tree after the deletions, before any test edit, named 16 test files with compile errors):

- Plan's nine hook files: all nine changed (`gateway_guard_test.go`, `lane_spawn_authority_test.go`, `role_naming_m3_notice_test.go`, `session_start_factory_test.go`, `session_start_factory_worker_test.go`, `session_start_prune_test.go`, `session_start_record_test.go`, `stale_run_m1_test.go`, `subdir_cwd_write_root_test.go`).
- Plan's eleven cli files: only four needed an edit at M5b (`factory_test.go`, `factory_m4_test.go`, `launcher_blockcap_infinite_test.go`, `update_version_downgrade_test.go`); `cc_test.go`, `codex_debug_trace_test.go`, `codex_factory_retire_test.go`, `factory_m5_test.go`, `glm_test.go`, `kanban_launch_facts_test.go` already compile after M5a and `kanban_dispatch_test.go` was deleted at M5a. Two cli files outside the plan's list needed one (`factory_net_m1_test.go`, `lane_entry_m2_test.go`, authored at M1 and M2) and `launcher_retired_entries_test.go` gained the constant `retiredLeaderMarker`. The measured set is 16 test files, not 20.
- Also changed, outside the plan's list: hook `session_start_env_helper_test.go` (the three keys left its list), `factory_net_m1_test.go` (the AC-017 assertions), `session_start_additional_context_test.go` and `session_start_test.go` (comments only), new `session_start_lang_test.go` and `session_start_no_kanban_notice_test.go`.
- Production: `session_start.go`, `session_start_record.go`, `session_start_factory.go` (comments), `session_start_factory_i18n.go` (comments), `session_stale_run.go` (one comment), new `session_start_lang.go`, `internal/config/envkeys.go`, `internal/cli/ptycaptest/harness.go`.
- Tests deleted by function inside retained files: `TestKanbanCompanionNoticeCarriesSpawnAuthority` (and the companion half of `TestLaneSpawnAuthorityFailOpenPreserved`), `TestKanbanNoticeSuppressedUnderFactoryEnv`, `TestKanbanNameChoicesUseLaneNotation`, the `kanban-leader` row of `TestRoleNamingM3NoticesCarryLeaderLaneTerms`, the `kanbanMessagesFor` part of the newline-hygiene test, `TestCompanionLabelResolvesToItsBareRole` (replaced by `TestMalformedLaneLabelYieldsNoRole`, the malformed-label half on the lane path).
- Retargeted: every `MOAI_KANBAN=1` trigger in the record, prune, stale-run, and subdir tests became the factory fan-out marker (`MOAI_FACTORY_WORKERS=2`, the record reader's surviving leader branch), the `run` companion in two record tests became a lane label. In the cli tests the written-out retired names (`retiredLeaderMarker`, `retiredSpecMarker`, `retiredLaneLabelMarker`, strings, since the constants are gone) stay in the ambient scrub lists, so a surviving session's value cannot reach the assertions that a factory session carries none of them.

Grep proof for the deleted symbols and constants, on the committed tree (output of the command, and exit):

```text
$ grep -rnE 'kanbanBootstrapNotice|kanbanLeaderNotice|kanbanCompanionNotice|kanbanMessagesFor|kanbanLocales|kanbanMessages\b|queuedBacklogCount|EnvMoaiKanban\b|EnvMoaiKanbanSpec|EnvMoaiKanbanLabel' internal cmd --include='*.go'
exit=1                                   (empty output, taken on the committed tree 1d0c64ef4)
$ grep -rln 'session_start_kanban' internal cmd --include='*.go'
internal/hook/session_start_env_helper_test.go
internal/hook/session_start_lang.go
internal/hook/session_start_lang_test.go
exit=0                                   (three comments naming the origin of the moved helpers; no code)
```

`git diff --stat 7e8087f40 HEAD` (taken at commit 3, before this record): `37 files changed, 465 insertions(+), 1904 deletions(-)`; it includes the 64 lines of this section's RED part from commit 2.

Kanban-named symbols that remain in `internal/hook` and `internal/config` after M5b (baseline for M6/M7), from `grep -rhoE '[A-Za-z_]*[Kk][Aa][Nn][Bb][Aa][Nn][A-Za-z_]*' internal/hook internal/config --include='*.go' --exclude='*_test.go'`: the package name `kanban` (99 uses: the `internal/kanban` import and its qualifiers, M8), the six frozen marker constants and their literals (`EnvMoaiKanbanID` 9, `EnvMoaiKanbanBackend` 6, `EnvMoaiKanbanLeadName` 4, `EnvMoaiKanbanLeadAddr` 4, `EnvMoaiKanbanCard` 4, `EnvMoaiKanbanSettingsInjected` 3; M7 renames the Go names, AC-016 freezes the values), `writeKanbanSessionRecord` (3) and `kanbanRoleFromEnv` (3) in `session_start_record.go` (both factory-path functions, renamed at M7), the stage lap name `kanban_record` in `session_start.go`, and comment mentions (`session_start_kanban_i18n.go` named as the origin of the moved helpers in `session_start_lang.go`).

#### Baseline-attribution

- Trees: `7e8087f40` for the RED-now greps (before any edit); `c10194abc` for the step-1 measurements; `c10194abc` plus only the new and re-pinned tests for the RED run; the working tree that became `1d0c64ef4` for the GREEN runs, the mutants, the suites, and the cli families (HEAD stayed `20b33e997` through every one of those measurements); the committed tree `1d0c64ef4` for the final net, the AC-014 tests, build, windows build, vet and gofmt. Re-read before the last commit: `git rev-parse --short HEAD` printed `20b33e997`, `git branch --show-current` printed `WT-launcher-entry-flags`.
- Judging build: every `go test`, `go build`, `go vet`, and `gofmt` ran from the Go toolchain on this tree; no installed `moai` binary produced any measurement cited here, except `moai slot acquire` and `moai slot release` (lease lines only).
- The whole-hook, whole-config, `go vet ./...` and cli-family runs preceded the three comment-only edits of commit 3 (`session_start_additional_context_test.go`, `session_start_test.go`, `session_start_factory_i18n.go`); the build, vet (host and windows), gofmt, net, AC-014 and AC-017 runs were repeated after them.

#### Gaps

1. The whole `internal/cli` suite was NOT run (4,635 tests; it cannot finish in the foreground): the 179 named tests of 25 files above ran, nothing broader.
2. The 14 `TestSyncGate*` functions of `internal/hook` were skipped (`-skip '^TestSyncGate'`) after the first whole-package attempt hit the 9-minute test timeout inside `TestSyncGateFailState_AC007_NoPathLooserThanToday`; they exercise the sync-gate shell script and read none of the files M5b edited, but they were not observed.
3. `TestInboundClaimIndependentOfEnvLabel` failed once under whole-suite load (`context deadline exceeded`) and passed alone and in the second whole run; it is a flake candidate, not attributed.
4. `TestNoNewEnvLiteralsInDiff` and `TestHookWrapperCopiesStayIdentical` are red on this tree for reasons outside M5b (above); they were neither fixed nor suppressed.
5. `go vet ./...` on windows over the whole module was not run (baseline PV-73 records a pre-existing failure outside this SPEC); windows vet ran on the four touched packages (exit 0), `GOOS=windows GOARCH=amd64 go build ./...` ran whole-module (exit 0).
6. `gofmt -l` over the files touched here printed nothing; over whole directories it also lists `internal/config/slice.go` and `internal/cli/mcp_claude.go`, which are untouched by this card.
7. The probe runner `probe.go` was not re-run for comparison, and the probe's `M5b.patch` was used as a model only (divergences below).
8. Mutants (e) and (f) were re-observed against the hook net only; the cli net does not read the two files they mutate.
9. The behavior change in Findings 3 below is read from the call sites (`grep` over non-test Go), not observed at runtime on the pre-M5b tree.
10. Ordering (verification-claim-integrity §2.3): the new tests and the production change share commit `1d0c64ef4`; the RED output is committed ahead of it (`20b33e997`) and was measured on `c10194abc` plus the test edits.
11. Worktree-guard refusal, one: a `for` loop whose body ran `git show HEAD:<computed file>` ("names git in a form too complex to verify"); re-issued as four plain `git show` commands, same measurement, nothing replaced by reading source. No other command was refused; a scratchpad edit that failed did so on my mistyped path, not on the guard.

#### Residual-risk

- Removing the three names from `kanbanVars` means a pty capture child launched from a shell that still exports one of them now receives the value; no reader of the three names remains in the repository, so nothing branches on it.
- A session that outlived the upgrade and still carries `MOAI_KANBAN` or `MOAI_KANBAN_LABEL` is a plain session for the hook: no notice, no record. `TestPreexistingKanbanArtifactsTolerated` and `TestSessionStartEmitsNoKanbanNotice` pin that.
- The long rationale comment that sat on the deleted kanban block (why the launcher cannot deliver the notice, why two channels, why startup-only) is now carried, shortened, by the comment on the factory block in `session_start.go`; the startup-only allowlist reasoning (a new source stays silent by default, an empty source reads as startup) is carried by `factoryBootstrapNoticeForSource`.

#### Findings (statements the tree contradicts, no SPEC file edited)

1. acceptance.md AC-014's command names `TestSessionStartEmitsNoKanbanNotice`, which did not exist on the tree before this milestone (the hook half of AC-017 in `factory_net_m1_test.go` says so in its own comment); it is authored here (`session_start_no_kanban_notice_test.go`). Its sub-case `leader_marker_alone` (`MOAI_KANBAN=1`, no run id) is green before the change by design: the leader notice builder returned the empty string for an empty run id, so only the record test sees that marker.
2. plan.md M5b lists 20 retained test files from the probe; on this tree 16 test files needed an edit (above), because the plan was measured before M5a retargeted `cc_test.go`, `codex_debug_trace_test.go`, `codex_factory_retire_test.go`, `factory_m5_test.go`, `glm_test.go`, and `kanban_launch_facts_test.go`, and before M1 and M2 authored `factory_net_m1_test.go` and `lane_entry_m2_test.go`.
3. The deleted kanban notice builder was also the SessionStart carrier of the stale-run notice for a session that is not a factory session: `kanbanBootstrapNotice` called `staleRunNoticeFor` before it read any kanban marker, so a surviving session whose launch label is a legacy leader spelling (`MOAI_KANBAN_LEAD_NAME=lead`, no `MOAI_FACTORY_WORKERS`) and a session whose existing record carries a legacy role got the stale-run relaunch notice there. After M5b the only non-test caller of `staleRunNoticeFor` is `factoryBootstrapNotice` (`session_start_factory.go:61`, which needs `MOAI_FACTORY_WORKERS`), so that session gets no stale-run notice at SessionStart (the factory message hook still calls `staleRunNotice` through `legacyFactoryHookNotice`). AC-014 says the stale-run notices are "emitted as before" for factory sessions, which holds (the factory stale-run tests pass); a non-factory legacy session is outside AC-014's wording and plan.md M5b's list, so this is a behavior change the SPEC does not name. `roleValueRelaunch` ("the kanban branch" in `session_stale_run.go`) stays reachable through the factory hook with an empty run id.
4. Divergences from the probe's `M5b.patch`: the probe removed both `clock.lap("factory_notice")` and `clock.lap("kanban_notice")`; this milestone keeps one lap, `factory_notice`, in the place before the lineage banner, because a lap times the work since the previous lap and the factory rule block would otherwise fall into the `chain_banner` lap (no test asserts lap names: `grep -rn 'kanban_notice\|factory_notice' internal --include='*.go'` finds `clock.lap("factory_notice")` at `session_start.go:566` and a file-name comment in the new test). The probe retargeted the removed constants to other factory markers as a compile model; this milestone keeps the retired names as written-out strings in the cli scrub lists and uses factory markers only where a test needs a trigger. The probe's `envkeys.go` hunk rewrote one comment (the `MOAI_FACTORY_WORKERS` documentation, plan `:284`); this milestone also rewords the `EnvMoaiFactoryWorker` documentation (plan `:301`, which named the deleted label as its counterpart) and the `EnvMoaiKanbanLeadAddr` comment that names `enterKanbanMode` (the M5a leftover at `:220`).

### M6 evidence

Recorded by the run-phase implementation worker (cycle_type ddd: characterization first, the milestone is behavior-preserving relocation plus deletion) for milestone M6 of card t1399 (board family, role carrier, and companion symbols, two commits), branch `WT-launcher-entry-flags`. Start state, re-read before any change: `pwd` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1399`, `git rev-parse --short HEAD` printed `5aa03a393`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Output blocks are verbatim excerpts of commands run in this run on this tree; each says what it drops.

#### RED-now and baseline before any change (E8)

RED-now of AC-012 on tree `5aa03a393` (read before any edit):

```text
$ find internal -name 'board_lock*.go'
internal/kanban/board_lock_clear_unix.go
internal/kanban/board_lock_cross_test.go
internal/kanban/board_lock_join_test.go
internal/kanban/board_lock_windows.go
internal/kanban/board_lock_clear_windows_test.go
internal/kanban/board_lock_errno_test.go
internal/kanban/board_lock_clear_windows.go
internal/kanban/board_lock_unix.go
internal/kanban/board_lock_wait_test.go
internal/kanban/board_lock_test.go
internal/kanban/board_lock.go
$ find internal -name 'board_store*.go' -o -name 'board_lock*.go' -o -name 'board_recover*.go' -o -name board.go -o -name column.go -o -name reconcile.go     (AC-012 third command: 18 paths, as the AC says)
internal/kanban/board_lock_clear_unix.go          ... 11 board_lock paths as above, plus
internal/kanban/board_recover_test.go
internal/kanban/board_recover.go
internal/kanban/board_store_test.go
internal/kanban/board_store.go
internal/kanban/board.go
internal/kanban/column.go
internal/kanban/reconcile.go
$ grep -rlE '\b(LoadBoard|WriteBoardState|AcquireBoardLock|RecoverBoard|ParseColumn|TransitionIntoRun|BoardState|BoardDir|DeclareRole|ResolveDeclaredRole|RoleDeclaration)\b' internal cmd --include='*.go'     (27 paths by a separate `wc -l`, all under internal/kanban; the set)
admission_test.go board_coverage_test.go board_lock_clear_unix.go board_lock_clear_windows_test.go board_lock_cross_test.go board_lock_errno_test.go board_lock_join_test.go board_lock_test.go board_lock.go board_recover_test.go board_recover.go board_store_test.go board_store.go board_test.go board.go column_test.go column.go f1_traversal_test.go fix2_probe_test.go fix3_wedge_test.go integration_lock_mutation.go kanban_helper_test.go reconcile_test.go role_naming_m1_test.go role_test.go role.go status_read_test.go
$ grep -rlE 'enterKanbanMode|enterKanbanCompanionMode|parseKanbanFlag|rejectKanbanOnCG|EnvMoaiKanban\b|EnvMoaiKanbanSpec|CompanionRoles|SplitCompanionLabel|kanbanLeaderNotice|kanbanCompanionNotice|kanbanBootstrapNotice' internal cmd --exclude='*_test.go'
internal/kanban/role.go
internal/kanban/bootstrap.go
exit=0
$ find internal -name 'factory_slots.go' -o -name 'backlog_store.go' -o -name 'integration_lock.go' -o -name 'slot_lease.go' -o -name 'state_lock.go' -o ...      (the positive control: four today, ten after M6)
internal/kanban/backlog_store.go
internal/kanban/factory_slots.go
internal/kanban/integration_lock.go
internal/kanban/slot_lease.go
```

The `find` exit and the symbol-grep exit were 0 (files listed); AC-012 asserts empty output on both after M6. The `state_lock*` files do not exist yet, which is why the control lists four.

Baseline of the characterization (whole `internal/kanban` package, before any change), under `moai slot acquire --resource t1399-run --max-duration 60m` (`slot t1399-run acquired by 0dcdf2d5-df5c-4da1-8870-24c2a5861303 until 2026-10-02T15:46:08Z`), env unset in one compound invocation, `go test ./internal/kanban -count=1 -v` redirected to a file:

```text
exit=0
ok  	github.com/modu-ai/moai-adk/internal/kanban	214.414s
top-level `--- PASS` lines: 582; `--- FAIL`: 0; `--- SKIP`: 0; all `--- PASS` lines including subtests: 857; `=== RUN` lines: 857
```

The sorted list of the 582 top-level results was kept outside the tree (scratchpad `baseline_top.txt`) so the after-run can be diffed by name modulo the renames.

#### Claim

- Step 1 (commit `d624b1cfc`, the board still present): the file-lock substrate is re-homed. `git mv` of `board_lock{,_unix,_windows,_clear_unix,_clear_windows}.go` to `state_lock{...}.go` and of the six board lock test files to `state_lock_*_test.go`; the wait-budget block (the seven constants and `boardLockRetryWait`, lines 72-170 of `board_store.go`) moved verbatim into the new `state_lock_wait.go` (a header comment added above it); the shared identifiers renamed exactly as plan.md lists (`BoardLock` to `StateLock`, `boardLockImpl`, `acquireBoardLockImpl`, `IsBoardLockHeld`, `ErrBoardLockHeld`, `BoardLockOwner`, `ErrBoardLockChangedHands`, `IsBoardLockChangedHands`, the seven budget constants, `boardLockRetryWait`, `flockBoardLock`, `atomicFileBoardLock`, `classifyBoardFlockErr`); the three wait-budget tests became `TestStateLock*`. No behavior change: whole `internal/kanban` suite 582 top-level PASS before and after, identical modulo the three renamed names.
- Step 2 (commit `09bd28037`): deleted `board.go`, `board_store.go`, `board_recover.go`, `column.go`, `reconcile.go`; deleted `AcquireBoardLock`, `boardLockPath`, `boardLockFileName` (`state_lock.go`) and both `ClearStaleBoardLock` (the Unix no-op and the Windows wrapper); deleted the role-declaration carrier in `role.go` (the file stays); deleted `CompanionRoles`, `companionLaunchers`, `CompanionLauncher`, `CompanionLabel`, `CompanionNumberLabel`, `SplitCompanionLabel`, `isCompanionRole`, `LeaderSocketPath`, `kanbanSocketDir` from `bootstrap.go`; removed the companion clause of `Record.WithRole`. 12 whole test files and 3 lock test files deleted; the 4 errno tests and the cross-process exclusion test re-pointed to `acquireStateLockImpl`; 6 retained test files lost the cases that call a deleted symbol.
- Frozen and untouched: the todo queue, the integration lock, the slot lease, landing, the factory slot code (`factory_slots.go`, `factory_runtime.go`, `factory_alive_*.go`), `FactoryFreeSlots`, `SplitLeaderLabel`, `LeaderLabel`, `LeaderNumberLabel`, `RoleLeader`, `RoleLane`, `IsLegacyLeaderSpelling`, `legacyLeaderSpelling`, the legacy state-directory read, the on-disk `.moai/state/kanban-board/roles` files, and every file outside `internal/kanban` plus this record.

#### Evidence

AC-012 (M6 part), each row a command run in this run on the committed tree `09bd28037` (`git status --short` empty), env unset in one compound invocation for every `go test`; exit is the shell exit read with `echo`:

| Row | Status | Command | Exit | Observed |
|---|---|---|---|---|
| board files gone | PASS | `find internal -name 'board_store*.go' -o -name 'board_lock*.go' -o -name 'board_recover*.go' -o -name board.go -o -name column.go -o -name reconcile.go` | 0 | empty (RED-now 18 paths) |
| board API and carrier symbols gone | PASS | `grep -rlE '\b(LoadBoard\|WriteBoardState\|AcquireBoardLock\|RecoverBoard\|ParseColumn\|TransitionIntoRun\|BoardState\|BoardDir\|DeclareRole\|ResolveDeclaredRole\|RoleDeclaration)\b' internal cmd --include='*.go'` | 1 | empty (RED-now 27 paths) |
| companion and notice symbols gone (RED-K3) | PASS | `grep -rlE 'enterKanbanMode\|enterKanbanCompanionMode\|parseKanbanFlag\|rejectKanbanOnCG\|EnvMoaiKanban\b\|EnvMoaiKanbanSpec\|CompanionRoles\|SplitCompanionLabel\|kanbanLeaderNotice\|kanbanCompanionNotice\|kanbanBootstrapNotice' internal cmd --exclude='*_test.go'` | 1 | empty (RED-now `role.go`, `bootstrap.go`) |
| substrate stays (positive control) | PASS | the ten-name `find` of AC-012, piped to `wc -l` | 0 | `10` (RED-now 4) |
| queue, integration lock, slot lease acquire through the re-homed budget | PASS | `grep -c 'stateLockWaitBudget' internal/kanban/backlog_store.go internal/kanban/integration_lock_mutation.go internal/kanban/slot_lease.go` | 0 | `2`, `2`, `3` |
| `role.go` keeps the symbols the factory reads | PASS | `grep -c 'RoleLeader' internal/kanban/role.go` | 0 | `3` |
| build | PASS | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | 0, 0 | no output |
| substrate selector, floor four | PASS | `go test ./internal/kanban -run '^(TestStateLockWaitBudgetDerivedFromNamedInputs\|TestStateLockWaitBudgetCoversSerializedMutations\|TestStateLockRetryWaitIsNotLockstep\|TestBacklogLockStuckHolderSurfacesBoundedNamedError)$' -v -count=1` | 0 | 4 PASS of 4 (`ok … 3.505s`); the same four passed on the step-1 tree (`ok … 3.752s`) |
| re-pointed substrate tests (added to the floor) | PASS | `go test ./internal/kanban -run '^(TestStateFlockErrnoContentionRemainsHeld\|TestStateFlockErrnoNonContentionIsNotHeld\|TestStateFlockErrnoPreservesErrnoAndPath\|TestStateFlockErrnoFailurePathClosesDescriptor\|TestStateLock_ExcludesAcrossProcesses\|…the four above)$' -v -count=1` (run on the pre-deletion tree, board still present) | 0 | 9 PASS of 9 (`ok … 4.004s`) |
| REQ-SRL-009 env-literal guard | PASS | `go test ./internal/hook -run '^TestNoNewEnvLiteralsInDiff$' -count=1 -v` | 0 | `env-literal sweep: 661 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2, distinct literals=0` |
| AC-015 net | PASS | the three selectors | 0 each | 8 cli, 2 hook, 2 discovery (block below) |

Whole `internal/kanban` suite, under `moai slot acquire --resource t1399-run --max-duration 60m`, released (`slot t1399-run released (was 0dcdf2d5-df5c-4da1-8870-24c2a5861303)`), `go test ./internal/kanban -count=1 -v` to a file, exit 0 each:

| Tree | Result | Top-level PASS / FAIL / SKIP | Including subtests |
|---|---|---|---|
| `5aa03a393` (baseline) | `ok … 214.414s` | 582 / 0 / 0 | 857 |
| `d624b1cfc` (step 1) | `ok … 219.324s` | 582 / 0 / 0 | 857; diff of the sorted name lists: exactly the three renamed names (`TestBoardLockRetryWaitIsNotLockstep`, `TestBoardLockWaitBudgetCoversSerializedMutations`, `TestBoardLockWaitBudgetDerivedFromNamedInputs` to the `TestStateLock*` names) |
| `09bd28037` (step 2) | `ok … 205.838s` | 514 / 0 / 0 | 742; diff against step 1: 74 names removed, 6 added (582 - 74 + 6 = 514) |

The 6 added names: `TestStateFlockErrnoContentionRemainsHeld`, `TestStateFlockErrnoNonContentionIsNotHeld`, `TestStateFlockErrnoPreservesErrnoAndPath`, `TestStateFlockErrnoFailurePathClosesDescriptor`, `TestStateLock_ExcludesAcrossProcesses` (the re-pointed substrate tests) and `TestFactoryLaneLabelNeverLeaderShape` (the lane-versus-leader half of the old `TestFactoryLaneLabelNeverKanbanShape`). The 74 removed names are the board, role-carrier, companion, reconcile, admission, column, and board-lock cases, plus the five board-entry originals of the re-pointed tests; the sorted lists are in the scratchpad (`step1_top.txt`, `step2b_top.txt`).

Net re-run on the committed tree, tool exit 0 each (step 1 and step 2 both green, identical selection):

```text
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (3.17s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (3.47s)
--- PASS: TestFactoryNetLeaderLaunch (1.42s)
--- PASS: TestFactoryNetLaneLaunch (2.84s)
--- PASS: TestFactoryNetBlockCap (1.20s)
--- PASS: TestFactoryEntryMatrix (5.05s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (0.48s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	18.453s            (8 of 8)
--- PASS: TestFactoryNetSessionRecord (0.00s)
--- PASS: TestFactoryNetSessionStartNotices (0.65s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.245s             (2 of 2)
--- PASS: TestDiscoverLeaderVerifiesLiveLeader (0.13s)
--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID (0.08s)
ok  	github.com/modu-ai/moai-adk/internal/discovery	0.541s         (2 of 2)
```

Compile proof, each exit 0 with no output, on the step-1 tree and again on the step-2 tree: `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `go vet ./internal/kanban/ ./internal/cli/ ./internal/hook/`; the same vet with `GOOS=windows GOARCH=amd64`. On the step-2 tree additionally `go vet` and the windows vet of `./cmd/t657-merge/ ./internal/cli/ ./internal/discovery/ ./internal/escalation/ ./internal/factorymsg/ ./internal/graph/ ./internal/homestate/ ./internal/hook/ ./internal/mission/ ./internal/statusline/ ./internal/web/ ./internal/kanban/` (the twelve packages that import `internal/kanban` or are it), and `golangci-lint run --timeout=5m ./internal/kanban/...` (v2.1.6, `0 issues.`). `gofmt -l internal/kanban` printed nothing at both steps.

Substrate mutants (observed-failure completion, verification-completeness §1.1). Mutated copies of `state_lock_unix.go` kept outside the tree (scratchpad `m6/mutA.go`, `mutB.go`, `mutC.go`, with `overlay_{A,B,C}.json`), run with `go test -overlay <overlay> ./internal/kanban -run '^(TestStateFlockErrno…|TestStateLock_ExcludesAcrossProcesses)$' -v -count=1` on the pre-deletion tree; each printed `FAIL` for the package (verdict lines and assertion lines only):

```text
mutant A (acquireStateLockImpl never calls flock, so nothing is excluded)
    state_lock_errno_test.go:55: second acquireStateLockImpl: expected contention, got nil error
--- FAIL: TestStateFlockErrnoContentionRemainsHeld (0.00s)
    state_lock_errno_test.go:164: attempt 0: expected contention, got nil error
--- FAIL: TestStateFlockErrnoFailurePathClosesDescriptor (0.00s)
    state_lock_test.go:90: second process output = "ACQUIRED", want HELD — the lock excluded nothing across processes
--- FAIL: TestStateLock_ExcludesAcrossProcesses (0.02s)
mutant B (classifyStateFlockErr reports every failure as contention)
    state_lock_errno_test.go:80: IsStateLockHeld(kanban board lock held) = true, want false   (x4, one per errno)
--- FAIL: TestStateFlockErrnoNonContentionIsNotHeld (0.00s)
    state_lock_errno_test.go:100: errors.Is(kanban board lock held, no locks available) = false, want true   (x4)
--- FAIL: TestStateFlockErrnoPreservesErrnoAndPath (0.00s)
mutant C (the flock-failure path leaks the descriptor: the Close is removed)
    state_lock_errno_test.go:178: descriptor leak: probe fd 6 before, 206 after 200 failed acquisitions (slack 16)
--- FAIL: TestStateFlockErrnoFailurePathClosesDescriptor (0.01s)
```

Each mutant is rejected by the tests that guard its neighbourhood and the others stay green (A leaves the classification tests green; B leaves the exclusion and descriptor tests green; C leaves everything but the descriptor test green). All three are the substrate properties the four re-pointed tests carry, observed red on `acquireStateLockImpl` and `classifyStateFlockErr` directly, not through any board entry.

Actual lists against the plan's lists, re-measured on `5aa03a393` before editing:

- Whole test files deleted: the plan's 12 (`board_coverage_test.go`, `board_recover_test.go`, `board_store_test.go`, `board_test.go`, `column_test.go`, `reconcile_test.go`, `admission_test.go`, `f1_traversal_test.go`, `f2_unresolved_test.go`, `fix2_probe_test.go`, `fix3_wedge_test.go`, `role_test.go`) and the three lock tests the plan names (`state_lock_cross_test.go`, `state_lock_join_test.go`, `state_lock_clear_windows_test.go`); the plan's `state_lock_errno_test.go` and `state_lock_test.go` were kept and re-pointed (the other two of the five).
- Retained test files that needed an edit, from `go test -gcflags=-e -run '^NoSuchTestZZZ$' ./internal/kanban` after the deletions: exactly the plan's six (`backlog_store_test.go`, `bootstrap_test.go`, `factory_label_test.go`, `kanban_helper_test.go`, `role_naming_m1_test.go`, `status_read_test.go`). The compile of the retained tests needed `runtimeIsWindows`, `runGitAt`, `deadPID` (host) and `deadPIDWin` (windows), as the plan lists; `readFileBytes` was not moved (its only retained caller was the deleted-with-the-board test in `status_read_test.go`).
- Retained-file cases removed: `TestCompanionRolesAreTheThreePhases`, `TestSplitCompanionLabel`, `TestCompanionLabelRoundTrips`, `TestCompanionNumberLabelRoundTrips`, `TestLeadNumberLabelIsNotACompanion`, `TestSplitLeadLabelAndCompanionAreDisjoint` (`bootstrap_test.go`); `TestBoardGuardRefusesLegacyLeadDeclaration`, `TestBoardGuardAdmitsLeaderDeclaration` (`role_naming_m1_test.go`); `TestUnresolvedCard_OutcomeDistinctAndByteUnchanged` (`status_read_test.go`); the helper-process operations `resolve-role`, `reacquire-lock`, `reader-loop`, `transition-run` (`kanban_helper_test.go`). Reduced rather than deleted: `TestLeadLabelIsBareRole` (its companion assertion dropped), `TestBacklogStore_NoLeadRoleGuard` (`backlog_store_test.go`: its `IsNotSoleWriter` assertion dropped, the no-role-guard write assertion kept), `TestFactoryLaneLabelNeverKanbanShape` (renamed `TestFactoryLaneLabelNeverLeaderShape`, the leader half kept).
- Comments reworded: in `bootstrap.go` twelve comment edits against the plan's six (the header, three in `LeaderLabel`, `LeaderNumberLabel`, `SplitLeaderLabel`, `factoryLaneRole`, `FactoryLaneLabel`, `SplitFactoryLaneLabel`, `isRunIDShape`, the socket-root block, and `FactoryLeaderSocketPath`: each named a deleted symbol or the companion chain), the `RoleLeader`, `RoleLane`, and header comments in `role.go`, the `ErrStateLockHeld` and `ErrStateLockChangedHands` documentation and the header in `state_lock.go`, `integration_lock_mutation.go` (the plan's `:24` and the `board_store.go` path mention at `:21` and `:91`), the headers of `state_lock_unix.go`, `state_lock_windows.go`, `state_lock_clear_unix.go`, the wrapper and core documentation in `state_lock_clear_windows.go`, and the `board_store.go` path mentions in `backlog_store.go` (two), `integration_lock_cross_test.go`, `state_lock_wait_test.go`.

Grep proof for the deleted symbols, on the committed tree `09bd28037` (output of the command, and exit):

```text
$ grep -rnE '\b(CompanionLauncher|CompanionLabel|CompanionNumberLabel|SplitCompanionLabel|isCompanionRole|LeaderSocketPath|kanbanSocketDir|companionLaunchers|ClearStaleBoardLock|acquireBoardLockSerialized|boardLockPath|boardLockFileName|requireLeaderRole|IsNotSoleWriter|ErrNotSoleWriter|IsWipLimitExceeded|BoardOptions|ReconcileCard|BoardPath|joinBoardReleaseErr|writeBoardAtomic)\b' internal cmd --include='*.go'
internal/kanban/status_read.go:95:// @MX:REASON: expected fan_in >= 3 (ReconcileCard, the review column's verify gate, any board rendering); selecting the primary while a worktree is live ...
```

The one hit is a comment in `status_read.go` that M6 did not edit (see Finding 2).

`git diff --stat 5aa03a393 HEAD` (taken on `09bd28037`, before this record): `56 files changed, 776 insertions(+), 4636 deletions(-)`; it includes the 55 lines of this section's RED-now and baseline part from commit `f3862640b`. Commits: `f3862640b` (docs: RED-now and baseline, committed ahead of the change), `d624b1cfc` (step 1), `09bd28037` (step 2), then this record as the fourth.

Kanban- and board-named symbols that remain in `internal/kanban` after M6 (baseline for M7 and M8), from `grep -rhoE '[A-Za-z_]*[Kk][Aa][Nn][Bb][Aa][Nn][A-Za-z_]*' internal/kanban --include='*.go' --exclude='*_test.go'`: the package name `kanban` (128 uses: the package clause and prose; the import path is M8's), 11 `KANBAN` (the marker literal `MOAI_KANBAN_ID` and comments), 4 `Kanban` (comments in `record.go`), `EnvMoaiKanbanLeadAddr` (2, comments in `bootstrap.go`), `MOAI_KANBAN_ID` (1), `loadKanbanRecords` (1, a comment in `record_prune.go`). No Go identifier with the word survives in `internal/kanban` non-test code. The word `board` survives in 43 prose lines of 14 non-test files (comments and error texts such as `kanban board lock held`, `open board lock %s`, which are behavior-preserved) and in the two windows-only constants `boardLockTransientRetries` and `boardLockTransientDelay` in `state_lock_windows.go` (not in plan.md's rename list, left as they were).

#### Baseline-attribution

- Trees: `5aa03a393` for the RED-now greps and the baseline suite (before any edit); the working tree that became `d624b1cfc` for the step-1 suite, build, vet, net, and the four-test selector (HEAD stayed `f3862640b` through them); the working tree that became `09bd28037` for the substrate tests, the mutants (pre-deletion, HEAD `d624b1cfc`), the step-2 suites, the net, builds, vets, lint, and the env-literal guard (HEAD stayed `d624b1cfc` through them); the committed tree `09bd28037` for the final AC-012 greps and the exit codes (`git status --short` empty, `git rev-parse --short HEAD` printed `09bd28037`). Re-read before every commit: HEAD, branch (`WT-launcher-entry-flags`), `git status --short`.
- Judging build: every `go test`, `go build`, `go vet`, `gofmt`, and `golangci-lint` ran from the Go toolchain and the installed `golangci-lint` v2.1.6 on this tree; no installed `moai` binary produced any measurement cited here, except `moai slot acquire` and `moai slot release` (lease lines only).
- The status-test relocation (two `TestReadCardStatus_*` tests) came after the first step-2 whole-suite run (512); the suite was run again on the final tree (514) and that run is the one cited.

#### Gaps

1. The whole `internal/cli` and `internal/hook` suites were NOT run; for those packages the cited runs are the factory net, `TestNoNewEnvLiteralsInDiff`, and the vet of both (host and windows). `TestHookWrapperCopiesStayIdentical` (red on the base for an unrelated reason, M5b record) was not run at all, so nothing is claimed about it.
2. `go vet ./...` over the whole module and `GOOS=windows go vet ./...` whole-module were not run; the windows vet ran over the twelve packages above (exit 0). The pre-existing windows vet failure outside this SPEC (`internal/cli/worktree/sweep_test.go:1687`, `parseLsofCWDs`) is in a package this milestone does not import or touch and was not exercised.
3. Nothing ran on Windows: the Windows lock substrate, the Windows clear core, and the deleted Windows clear suite (`TestClearStaleBoardLock_*`, five tests incl. the re-acquire race) were checked only by compilation (`GOOS=windows go build`, `go vet`). The plan deletes that suite with the board, so the dead-owner-cleared, live-owner-refused, and re-acquire-race observations of `clearStaleLockAtPath` now have no Windows test of their own; `integration_lock_mutation_windows_test.go` (retained) still exercises the same core through the integration mutation lock.
4. The mutant runs piped the `go test` output through `grep` and `head`, so the tool exit code of each mutant run was not read separately; each run printed `FAIL` for the package and the failing test names above.
5. The probe runner `probe.go` was not re-run for comparison; `M6a.patch`, `M6a.rm`, `M6b.patch`, `M6b.rm` were used as a model only (divergences below), and `git apply` could not apply them (the patches use `/dev/null` sources that this git rejects), so the edits were made by hand.
6. The `internal/kanban` suite was run once per tree, not repeated; no flake was observed in the four whole-suite runs (baseline, step 1, step 2 twice) or the targeted ones.
7. Worktree-guard refusals, four, each re-issued as a plain form measuring the same thing, none replaced by reading source: (a) `cd internal/kanban && git mv … && cd ../..` ("changes directory to a location computed at runtime before running git"), re-issued with repo-relative `git mv` paths; (b) a `for` loop over helper names that contained the word `runGitAt` ("names git in a form too complex to verify"), re-issued as one `grep -E` alternation; (c) `perl` with a path built from a shell variable (`$S`) to make the mutated copies, replaced by files written with the Write tool and literal-path overlay JSON; (d) a heredoc that wrote a perl script followed by other commands, replaced by the Write tool for the script and a separate `perl -0pi <script> <file>` call.
8. Ordering (verification-claim-integrity §2.3): the re-pointed substrate tests and the substrate-touching production edits share commit `09bd28037` with the deletions; the baseline and RED-now are committed ahead of it (`f3862640b`), and the substrate tests were run green and the three mutants run red on the pre-deletion tree (HEAD `d624b1cfc`, the board still present), which is a session record, not a commit-graph fact.

#### Residual-risk

- Behavior preserved on purpose, which still reads wrong: `ErrStateLockHeld` keeps the text `kanban board lock held`, `ErrStateLockChangedHands` keeps its text, and the acquire errors keep `open board lock %s` and `lock board lock %s`; anything that matches those strings (none found in `internal cmd` by grep during step 1, but not searched outside Go) is unaffected. Rewording them is a behavior change this milestone does not make.
- A past board left `.moai/state/kanban-board/roles/*.json` and `board.json` on disk; after M6 nothing reads or writes them. The plan and the operator verdict leave them in place.
- `state_lock_clear_unix.go` now carries only its header comment and `package kanban` (the function it held was the board-bound no-op). It stays because AC-012's positive-control `find` lists it and because `integration_lock_mutation_unix.go` points at its gate rationale.
- `Record.WithRole` now discards the three companion role values; a rewrite through `WithRole` of an old record that carries `plan`, `run`, or `sync` would no longer store that role. Whether any reader renders such a stored value, and whether any non-test caller reaches `WithRole` with one, was not examined beyond the compile.

#### Findings (statements the tree contradicts, no SPEC file edited)

1. plan.md M6 step 2 lists `f1_traversal_test.go` among the six that "exercise only the board or the carrier"; on this tree it also carried `TestReadCardStatus_RejectsTraversalSpecID` and `TestReadCardStatus_AcceptsCanonicalSpecID`, which call the retained `ReadCardStatus` (`status_read.go`). The file was deleted as the plan says, and the two tests were relocated verbatim to `status_read_test.go` so the traversal guard of a retained function keeps its test (net effect on the plan's count: 12 whole files go, plus 2 tests move).
2. After M6, `ReadCardStatus` (`internal/kanban/status_read.go:100`) has no non-test caller: `grep -rn 'ReadCardStatus\|ReadPrimarySpecStatus' internal cmd --include='*.go' --exclude='*_test.go'` finds callers only for `ReadPrimarySpecStatus` (`internal/cli/todo_autodone.go:306`, `internal/cli/todo_landed.go:337`). Its `@MX:REASON` at `:95` still names `ReconcileCard` (deleted) as an expected caller. Not edited and not deleted: it is outside the plan's lists and AGENTS.md §5 forbids deleting seemingly-unused code without approval; it is a candidate for the operator's next decision.
3. Comment-only mentions of the renamed lock files outside `internal/kanban`, found by `grep -rn 'board_lock' internal cmd --include='*.go'` after step 1 and left alone (the stage data shows no edit outside `internal/kanban`): `internal/cli/todo_test.go:564`, `internal/cli/gate_lock_unix.go:6`, `internal/cli/gate_lock_windows.go:7` and `:26`, `internal/cli/gate_lock.go:14`. They now name files that do not exist.
4. acceptance.md AC-012's third `find` command lists "18 paths" today and "empty output, exit 0" after: both hold (18 measured at `5aa03a393`, empty at `09bd28037`).

### M7 evidence

Recorded by the run-phase implementation worker (cycle_type ddd: the milestone is a behavior-preserving mechanical rename; characterization is the factory net and the whole-package suites) for milestone M7 of card t1399 (Go identifiers and Go file names outside `internal/web`), branch `WT-launcher-entry-flags`. Start state, re-read before any change: `pwd` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1399`, `git rev-parse --short HEAD` printed `d206491f3`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Output blocks are verbatim excerpts of commands run in this run on this tree; each says what it drops. The sections below grow in four commits: this sizing and baseline record first (commit A), then the re-runnable rename program, the rename itself, and the comment rewording, and the closing evidence last.

#### Sizing of AC-018's sweeps on this tree (J.5 item 1, measured before any change)

The auditor's 123 files / 502 lines were measured on the modeled FINAL tree; these are the real counts on the tree `d206491f3`, before M7 (RED-N1, RED-N2, RED-N3 style baselines):

```text
$ grep -rlP '(?i)(?<!moai_)kanban|moai_kanban(?!_(id|lead_addr|lead_name|settings_injected|backend|card)\b)|칸반|かんばん|カンバン|看板' internal cmd --include='*.go' --include='*.templ' --include='*.js' --exclude='*_test.go' --exclude-dir=testdata --exclude-dir=node_modules      (the AC-018 word grep, files)
exit=0 files=183
$ the same pattern with grep -rhP (matching lines)
1207
$ the same pattern restricted to --include='*.go' --exclude-dir=web (files, then lines)
167 files, 998 lines
files per directory (the 183): internal/cli 65, internal/kanban 59, internal/hook 16, internal/web 14, internal/statusline 5, internal/config 4, internal/discovery 3, internal/web/assets 2, internal/spec 2, internal/graph 2, internal/factorymsg 2, and one each in internal/stateanchor, internal/session, internal/homestate, internal/feedback, internal/factorylane, internal/core/git, internal/cli/ptycaptest, internal/cli/agentlint, cmd/t657-merge
$ find internal cmd -iname '*kanban*'      (RED-N2: 15 names; the four template paths and the web test file are not Go-source M7 names)
internal/cli/kanban_autonomy_test.go
internal/cli/kanban_bootstrap_test.go
internal/cli/kanban_launch_facts_test.go
internal/cli/kanban_lead_name_test.go
internal/cli/kanban_settings_test.go
internal/cli/kanban_settings.go
internal/cli/kanban.go
internal/hook/session_start_no_kanban_notice_test.go
internal/kanban
internal/kanban/kanban_helper_test.go
internal/statusline/preexisting_kanban_artifacts_m1_test.go
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
internal/template/templates/.claude/skills/moai-kanban-foreman
internal/web/preexisting_kanban_artifacts_m1_test.go
$ grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go' | wc -l      (RED-N3, tests included)
183
```

Identifier occurrences outside `internal/web`, counted from the syntax tree (a scratch program kept outside the tree: every identifier token whose lower-case text contains `kanban`, files under `internal` and `cmd` outside `internal/web`, `testdata`, and the template tree): 2,527 occurrences of 40 distinct names, of which 1,980 are the bare package qualifier `kanban` (M8's) and 547 are not. The 547 are the six marker constants (EnvMoaiKanbanBackend 118, EnvMoaiKanbanCard 31, EnvMoaiKanbanID 143, EnvMoaiKanbanLeadAddr 34, EnvMoaiKanbanLeadName 31, EnvMoaiKanbanSettingsInjected 25: 382), the entry-parse type (`kanbanEntryParse` 14, field `KanbanEnabled` 1), the launch-facts function (`exportKanbanLaunchFacts` 7), `prepareKanbanSettings` 20, `kanbanRoleFromEnv` 8, `scrubKanbanEnv` 23, `writeKanbanSessionRecord` 31, `clearKanbanEnv` 13, `clearKanbanLauncherEnv` 6, `downgradeKanbanVars` 2, `kanbanVars` 9, `kanbanDirDB` 4, `kanbanTemp` 3, `seedKanbanRecord` 3, and 19 distinct test function names (21 occurrences: `TestPreexistingKanbanArtifactsTolerated` is declared in three packages). The plan's 515 was measured on the modeled tree (PV-79); this count is on the real tree. Four of the test names are named by acceptance commands and so stay (decision recorded under Findings below): `TestKanbanEntryRefused` (AC-011), `TestSessionStartEmitsNoKanbanNotice` (AC-014), `TestPreexistingKanbanArtifactsTolerated` (AC-017, three packages here), `TestLegacyKanbanRouteRedirects` (AC-019, `internal/web`, M9).

String literals outside `internal/web` that carry the word, by kind (scratch program, same scope; 401 distinct file-and-literal lines in its output): import paths and `./internal/kanban` path strings (M8); the six frozen marker values and the Codex allowlist literal (frozen); the transient prefix `moai-kanban`; error texts in `internal/kanban` (`kanban backlog …` 7 and `kanban: …` 22 in the five files design 4.7 names, `read/write/prune kanban record(s)` 10 and `kanban board lock …` 2 in files it does not name); two launcher diagnostics in `internal/cli/kanban.go` and one in `doctor_factory_run.go`; the timing lap `kanban_record`; help texts in `todo.go` (2), `gtd.go`, `mcp_todo.go`, `tokens.go`; the retired-entry refusal (`launcher_retired_entries.go`, an allowed file); and test fixtures.

#### Baseline before any change (characterization)

Factory net on `d206491f3`, env unset in one compound invocation (AC-015 selectors, `-v -count=1`, output to a file):

```text
--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (3.56s)
--- PASS: TestGLMFactoryLaneJoinsDiscoveredLeader (3.35s)
--- PASS: TestFactoryNetLeaderLaunch (1.41s)
--- PASS: TestFactoryNetLaneLaunch (3.04s)
--- PASS: TestFactoryNetBlockCap (1.16s)
--- PASS: TestFactoryEntryMatrix (5.78s)
--- PASS: TestCCFactoryEntryRecordsFailOpenRunMetadata (0.54s)
--- PASS: TestPrepareKanbanSettingsWritesTransientFile (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	19.599s            (8 of 8)
--- PASS: TestFactoryNetSessionRecord (0.00s)
--- PASS: TestFactoryNetSessionStartNotices (0.57s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.159s             (2 of 2)
--- PASS: TestDiscoverLeaderVerifiesLiveLeader (0.13s)
--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID (0.09s)
ok  	github.com/modu-ai/moai-adk/internal/discovery	0.539s         (2 of 2)
```

Whole-package suites, under `moai slot acquire --resource t1399-run --max-duration 60m` (`slot t1399-run acquired by 0dcdf2d5-df5c-4da1-8870-24c2a5861303 until 2026-10-02T16:25:19Z`), `go test -p 2 -v -count=1 ./internal/config ./internal/kanban ./internal/hook ./internal/discovery ./internal/codexwiring ./internal/statusline ./internal/factorymsg ./internal/factorylane` to a file (tool exit 1, the one failure named below), per-package result lines and top-level / subtest counts (`awk` over the file):

```text
ok  	github.com/modu-ai/moai-adk/internal/config	3.957s
ok  	github.com/modu-ai/moai-adk/internal/kanban	217.653s
--- FAIL: TestHookWrapperCopiesStayIdentical (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	435.972s
ok  	github.com/modu-ai/moai-adk/internal/discovery	3.056s
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	1.005s
ok  	github.com/modu-ai/moai-adk/internal/statusline	28.883s
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	71.796s
ok  	github.com/modu-ai/moai-adk/internal/factorylane	2.466s
internal/config top: pass=488 fail=0 skip=0 sub=412
internal/kanban top: pass=514 fail=0 skip=0 sub=228
internal/hook top: pass=1269 fail=1 skip=6 sub=1934
internal/discovery top: pass=14 fail=0 skip=1 sub=0
internal/codexwiring top: pass=91 fail=0 skip=0 sub=71
internal/statusline top: pass=333 fail=0 skip=2 sub=460
internal/factorymsg top: pass=70 fail=0 skip=1 sub=122
internal/factorylane top: pass=53 fail=0 skip=0 sub=8
```

The one failure, `TestHookWrapperCopiesStayIdentical`, is the known red on the base for a reason outside this SPEC (named in the delegation); it is the only test ignored anywhere in this record.

#### The rename program (committed, re-runnable)

Path: `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m7_rename.go` (`//go:build ignore`, standard library only; a copy sits in the gitignored `.moai/reports/t1399/rename/`). Run command, from the tree root:

```text
go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m7_rename.go -root <git tree> [-dry-run]
```

What it does and refuses is its header comment. Summary: step 1 deletes the `KanbanEnabled` field of the entry-parse struct; step 2 deletes the `exportFactoryLaunchFacts` wrapper where `exportKanbanLaunchFacts` is declared; step 3 renames every Go identifier that carries the word (the six marker constants by the design 4.7 map, `kanbanEntryParse` to `launcherEntryParse`, every other identifier by replacing the word), except the bare package name `kanban`, package clauses, and the four test names acceptance commands pin; step 4 rewrites tokens that spell a renamed identifier inside comments and string literals, and (comments only) file names it renames; step 5 applies the design 4.7 literal rewrites; step 6 renames files (`git mv` for tracked files). It never touches `internal/web`, non-Go files, `testdata`, `node_modules`, `vendor`, `.git`, `.moai`, `.claude`, or `internal/template/templates`, and it leaves the six marker string values alone (only identifiers, and tokens that spell an identifier, are rewritten). It refuses (exit 2) on a tree with tracked modifications, and refuses (exit 1, nothing written) on a package-level name collision, a file-rename collision, or a rewrite that does not parse.

Observed results, each a command run in this run:

| What | Command, tree | Observed |
|---|---|---|
| dry run | `-dry-run` on the real tree `d206491f3` plus the program | `m7_rename: identifiers renamed: 540 (35 distinct names, 111 files); comment/string tokens rewritten: 61; literal rewrites: 38; KanbanEnabled fields deleted: 1; launch-facts wrappers deleted: 1; files renamed: 10 (dry run, nothing written)` |
| per-package identifier counts of that run | same | `internal/cli` 294, `internal/cli/ptycaptest` 17, `internal/config` 12, `internal/discovery` 7, `internal/homestate` 3, `internal/hook` 202, `internal/kanban` 5 (sum 540) |
| the real run, commit `602aa8c93` | the run command on tree `48365b722` | the same summary line without the dry-run suffix |
| run 2, idempotence | the run command on `602aa8c93` | `m7_rename: identifiers renamed: 0 (0 distinct names, 0 files); comment/string tokens rewritten: 0; literal rewrites: 0; KanbanEnabled fields deleted: 0; launch-facts wrappers deleted: 0; files renamed: 0`, then `git status --short` printed nothing |
| program extended, then run on `5513923a6` | the run command | `… identifiers renamed: 0 (0 distinct names, 8 files); comment/string tokens rewritten: 9; …; files renamed: 0` (nine comment mentions of renamed files, commit `e41310611`) |
| run 3, idempotence on the final tree `e41310611` | the run command | all zeros again, `git status --short` printed nothing |
| refuses a dirty tree | the run command on a scratch clone with a staged rename | `m7_rename: the tree has tracked modifications; commit or revert them first:` and `R  internal/cli/cc.go -> internal/cli/cc_moved.go`; `exit status 2` |
| converts a planted late reference | the run command on a scratch clone of `602aa8c93` holding an untracked `internal/cli/kanban_planted_test.go` that uses `config.EnvMoaiKanbanBackend`, `exportKanbanLaunchFacts`, and a comment and message naming `prepareKanbanSettings` | `identifiers renamed: 2`, `text token (comment or string): prepareKanbanSettings -> prepareFactorySettings`, `file: internal/cli/kanban_planted_test.go -> internal/cli/factory_planted_test.go`; the converted file read back uses `config.EnvFactoryBackend`, `exportFactoryLaunchFacts`, and `prepareFactorySettings`; `go vet ./internal/cli/` of that scratch clone exit 0 |
| one command reproduces the commit | the final program on a scratch clone checked out at `48365b722` (plus the new guard test file), then `diff -rq` of `internal` and `cmd` against the committed tree `e41310611` | summary `identifiers renamed: 540 (35 distinct names, 112 files); comment/string tokens rewritten: 70; literal rewrites: 38; … files renamed: 10`; `diff -rq` listed exactly 14 files, the 13 hand-edited comment files of commit `f4a809fcf` and `internal/cli/ptycaptest/harness.go` (the hand edit in `602aa8c93`) |

Not covered by the program (hand edits, so a post-absorption re-run leaves them to a manual grep): the 11 doctrine-path comment lines and the other comment rewordings of commit `f4a809fcf` (13 files), and the one comment of `internal/cli/ptycaptest/harness.go` in `602aa8c93`, which would otherwise have turned the REQ-SRL-009 env-literal guard red in that commit (the rename modifies a line that spells `MOAI_KANBAN*`, and a modified line is an added line).

Actual versus the plan: 540 identifier occurrences renamed against the plan's 515 (PV-79, modeled tree), plus one field and one wrapper deleted; the guard test below saw 542 identifier tokens on the pre-rename tree (the 540, the deleted field, and the identifier inside the deleted wrapper).

#### Claim

- Every Go identifier outside `internal/web` that carried the word is renamed; the bare package qualifier `kanban` (1,980 occurrences, M8's), the package clauses (M8), and five occurrences of three pinned test names remain. The six marker constants carry the design 4.7 names and their string values are unchanged. `kanbanEntryParse` is `launcherEntryParse` with no `KanbanEnabled` field. There is one launch-facts function, `exportFactoryLaunchFacts`. The transient settings prefix is `moai-factory`, the timing lap is `factory_record`, the five landing and backlog files carry `todo queue …` and `factory: …` error prefixes, and the todo, gtd, MCP, and tokens help sentences no longer carry the word.
- Ten Go files were renamed (`kanban.go` to `factory_launch_helpers.go`, `kanban_settings.go` to `factory_settings.go`, and eight test files by the word replacement). `find internal cmd -iname '*kanban*'` now lists six names, all owned by later milestones.
- Behavior preserved: the factory net is 8 / 2 / 2 on the final tree and the whole-package test name lists of the eight packages are identical to the baseline modulo seven renamed test names.

#### Evidence

AC rows, each a command run on the committed tree `e41310611` (`git status --short` empty), env unset in one compound invocation for every `go test`; the tool exit was 0 for every row unless stated:

| AC and row | Status | Command | Observed |
|---|---|---|---|
| AC-015 net, cli | PASS | the AC-015 `./internal/cli` selector, `-v -count=1` | 8 `--- PASS` lines, including `TestPrepareFactorySettingsWritesTransientFile` (the new name exists, the old name does not): `TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestFactoryNetLeaderLaunch`, `TestFactoryNetLaneLaunch`, `TestFactoryNetBlockCap`, `TestFactoryEntryMatrix`, `TestPrepareFactorySettingsWritesTransientFile`, `TestCCFactoryEntryRecordsFailOpenRunMetadata`; `ok  	github.com/modu-ai/moai-adk/internal/cli	34.130s` |
| AC-015 net, hook | PASS | the AC-015 `./internal/hook` selector | 2 PASS (`TestFactoryNetSessionRecord`, `TestFactoryNetSessionStartNotices`), `ok … 1.882s` |
| AC-015 net, discovery | PASS | the AC-015 `./internal/discovery` selector | 2 PASS, `ok … 0.727s` |
| AC-016 frozen values | PASS | `go test ./internal/config -run '^TestFactoryMarkerValuesFrozen$' -v -count=1` | `--- PASS: TestFactoryMarkerValuesFrozen (0.00s)`; the test file is unchanged in assertion (its six literals are written in the file; the compiler retargeted the six constant references and the text rule the six label strings) |
| AC-016 allowlist | PASS | `go test ./internal/codexwiring -run '^TestMCPServerEnvVarsKeepFactoryMarkers$' -v -count=1` | 1 PASS |
| AC-016 legacy state dir | PASS | `go test ./internal/kanban -run '^TestLegacyStateDirStillRead$' -v -count=1` | 1 PASS |
| AC-016 discovery | PASS | the two discovery tests (net row) | 2 PASS |
| AC-017 | PASS | `go test ./internal/cli ./internal/hook ./internal/web ./internal/statusline -run '^TestPreexistingKanbanArtifactsTolerated$' -v -count=1` | 1 PASS in each of the four packages |
| AC-011 | PASS | `go test ./internal/cli -run '^TestKanbanEntryRefused$' -v -count=1` | `--- PASS: TestKanbanEntryRefused (30.47s)` (name pinned, unchanged) |
| AC-014 (tests) | PASS | `go test ./internal/hook -run '^(TestSessionStartEmitsNoKanbanNotice\|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide)$' -v -count=1` | 2 PASS |
| AC-024 (hook tests) | PASS | `go test ./internal/hook -run '^(TestUnbindNoticeRebindLinePresence\|TestFactoryGuideNamesWorkerJoinInEveryLocale)$' -v -count=1` | 2 PASS |
| AC-024 (cli help test) | GAP | `go test ./internal/cli -run '^TestLauncherHelpDocumentsLaneEntry$' -v -count=1` | the selector swept 0 tests: no commit of this repository contains that name (Findings 1) |
| AC-018, identifier half (new guard) | PASS | `go test ./internal/cli -run '^(TestRetiredWordIdentifierScanHasTeeth\|TestNoRetiredWordIdentifiersOutsideWeb)$' -v -count=1` | both PASS (`--- PASS: TestNoRetiredWordIdentifiersOutsideWeb (2.10s)`); RED before: the same two tests on a scratch clone of `48365b722` with the test file added printed `--- PASS: TestRetiredWordIdentifierScanHasTeeth` and `retired_word_identifiers_m7_test.go:112: 542 Go identifier(s) outside internal/web still carry the retired mode word`, `--- FAIL: TestNoRetiredWordIdentifiersOutsideWeb (0.64s)`, tool exit 1; mutant: a planted `func helperKanbanThing() {}` in a scratch clone printed `1 Go identifier(s) … ../../internal/cli/zz_planted_old_name_test.go:4 helperKanbanThing`, `--- FAIL`, exit 1 |
| AC-018, file names | PARTIAL (M7 part PASS) | `find internal cmd -iname '*kanban*'` (RED-N2 15 names) | 6 names: `internal/kanban` (M8), `internal/web/preexisting_kanban_artifacts_m1_test.go` (M9), the skill directory `internal/template/templates/.claude/skills/moai-kanban-foreman` and the three `kanban-dispatch*.md` template rules (M10); no `internal/cli`, `internal/hook`, `internal/kanban`, `internal/statusline` file name remains |
| AC-018, word grep (RED-N1) | PARTIAL | the AC-018 grep, files then lines | 169 files, 1,024 lines (RED-N1 183 / 1,207); excluding `internal/web`: 153 files, 815 lines, of which 67 are import paths, 59 are `package kanban` clauses and 512 are code lines using the package qualifier (all M8), 157 are comment lines (22 `SPEC-KANBAN-*` citations, 57 naming the package or its path, 78 other prose; no milestone owns them), and 20 are other code lines (string literals and the allowed files, listed under the remaining-word table) |
| AC-018, importers (RED-N3) | unchanged (M8) | `grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go' \| wc -l` | 183 |
| AC-018 / REQ-024 build | PASS | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | exit 0, no output, each |
| vet | PASS | `go vet` of `./internal/cli/... ./internal/hook/... ./internal/config/... ./internal/kanban/... ./internal/discovery/... ./internal/homestate/... ./internal/statusline/... ./internal/factorymsg/... ./internal/factorylane/... ./internal/codexwiring/... ./internal/web/... ./internal/escalation/... ./internal/graph/... ./internal/mission/... ./cmd/...`, host, and the same with `GOOS=windows GOARCH=amd64` (`./internal/cli/` and `./internal/cli/ptycaptest/` in place of `./internal/cli/...`) | exit 0, no output, both; the whole-module `go vet ./...` and the pre-existing windows failure `internal/cli/worktree/sweep_test.go:1687` were not exercised (Gap 2) |
| gofmt | PASS for touched files | `gofmt -l internal cmd .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe` | prints `internal/cli/mcp_claude.go`, `internal/config/slice.go`, `internal/web/codex_panel_test.go`; none of the three is in this diff (`git status --short` listed none of them), so they are unformatted on the base |
| REQ-SRL-009 env-literal guard | PASS | `go test ./internal/hook -run '^TestNoNewEnvLiteralsInDiff$' -v -count=1` | `env-literal sweep: 1167 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2, distinct literals=0`; the first full-suite run after the program printed `distinct literals=1` `[MOAI_KANBAN]` (the harness.go comment), which is why that comment was reworded in the same commit |

Whole-package suites, before (tree `d206491f3`, above) and after (tree `e41310611`; the non-hook packages in one `go test -p 2 -v -count=1` run, the `internal/hook` package in two runs by test-name range `^Test[A-L]` and `^Test[M-Z]` to stay under the foreground limit), top-level / subtest counts by the same `awk`:

```text
                         before                                    after (e41310611)
internal/config          488/0/0 sub 412  ok                       488/0/0 sub 412  ok  5.906s
internal/kanban          514/0/0 sub 228  ok                       514/0/0 sub 228  ok  251.952s
internal/hook            1269 pass, 1 fail, 6 skip, sub 1934       671+597 pass, 1 fail, 3+3 skip, sub 1137+797=1934
internal/discovery       14/0/1 sub 0     ok                       14/0/1 sub 0     ok  3.764s
internal/codexwiring     91/0/0 sub 71    ok                       91/0/0 sub 71    ok  1.178s
internal/statusline      333/0/2 sub 460  ok                       333/0/2 sub 460  ok  32.087s
internal/factorymsg      70/0/1 sub 122   ok                       70/0/1 sub 122   ok  93.504s
internal/factorylane     53/0/0 sub 8     ok                       53/0/0 sub 8     ok  4.109s
```

The hook failure is `TestHookWrapperCopiesStayIdentical` in both (the known base red). The hook top-level pass count reads 1268 against 1269 only because one `--- PASS` line is glued to a log line in the file; counting every `--- PASS|FAIL|SKIP: <name>` occurrence, glued ones included, gives 6,079 results before and 6,079 after, and `diff` of the two sorted name lists shows exactly the seven renames: `TestKanbanHelperProcess`, `TestKanbanRoleFromEnvLegacyLabelsNotRecognized`, `TestKanbanRoleFromEnvNewVocabulary`, `TestKanbanRoleFromEnvReadsOnlyLaneLabels`, `TestKanbanSessionRecord_NoStrayTreeInSubdirCWD`, `TestNonKanbanSessionWritesNoRecord` (to the `TestFactory…`/`TestNonFactory…` names) and `TestSessionRecordIgnoresRetiredKanbanMarkers` (five results, to `…RetiredFactoryMarkers`).

`internal/homestate` and `internal/cli/ptycaptest`, whole: `ok  	github.com/modu-ai/moai-adk/internal/homestate	98.156s`, `ok  	github.com/modu-ai/moai-adk/internal/cli/ptycaptest	4.425s`. `internal/cli`, by file-derived name list: 86 tests from the files the rename changed that carry behavior text (settings, launch effort, cross-session settings, launch facts, lead name, autonomy, bootstrap, the new guard, retired entries, characterization, gtd, tokens, MCP factory messages, role naming doctor, goal readers, block-cap) all PASS (`ok  	github.com/modu-ai/moai-adk/internal/cli	120.884s`, 86 `--- PASS`, 0 skip), run on the tree before the comment commits.

A wider `internal/cli` list (1,894 names from the launcher, factory, todo, hook, MCP, and settings test files) timed out at the default 10 minute package limit on a machine whose load average read 72 (`uptime`): 498 tests ran, 4 failed, 1,396 did not run. Those 4 on the renamed tree: `TestStopChainEffectParityGolden`, `TestStopChainMemberCostWithinBudget`, `TestSyncGateLanguageDetectionMatchesScript`, `TestCodexTaskBackgroundHandshakeHonorsTaskBound`. Re-run alone on the renamed tree and on a scratch clone of the pre-rename base `48365b722`: the same three fail on both (`TestStopChainEffectParityGolden` `Claude path decision = allow, want deny`; `TestStopChainMemberCostWithinBudget` member budgets exceeded under load; `TestSyncGateLanguageDetectionMatchesScript` `languages: Go = [kotlin], script = [kotlin java]`) and `TestCodexTaskBackgroundHandshakeHonorsTaskBound` passes alone on both. They are not caused by this milestone.

#### Remaining occurrences of the word outside `internal/web`, and their owners (baseline for M8-M10)

```text
$ identifiers (syntax-tree count, scratch program): 1985 occurrences of 4 names
1980 kanban (the package qualifier)                                          -> M8
   3 TestPreexistingKanbanArtifactsTolerated, 1 TestKanbanEntryRefused,
   1 TestSessionStartEmitsNoKanbanNotice                                     -> kept on purpose (acceptance commands name them)
$ non-test string literals that still carry the word (scratch program; import paths and the six frozen marker values excluded)
internal/cli/home_state_coverage.go (3), internal/cli/migrate_home_state.go (4)      "./internal/kanban" path strings and the key "kanban" -> M8
internal/cli/launcher_retired_entries.go ("--kanban", "-k/--kanban is retired: Kanban Mode was removed; ")  -> allowed file (REQ-017)
internal/kanban/state_dir.go ("kanban")                                      -> allowed file, frozen legacy state directory
internal/cli/doctor_factory_run.go ("no session records — no factory or kanban run declared")     -> no owner
internal/cli/factory_launch_helpers.go (two launcher diagnostics beginning "kanban: ")           -> no owner
internal/kanban/record.go (9), record_prune.go (1) ("read/write/prune kanban record(s)")           -> no owner (not in design 4.7's table)
internal/kanban/state_lock.go ("kanban board lock held", "kanban board lock changed hands …")      -> no owner (behavior preserved since M6)
$ symbols and constants left by design: boardLockTransientRetries and boardLockTransientDelay (windows-only, state_lock_windows.go), not in the plan's list; unchanged.
```

Comment lines (157 after M7 in non-test Go outside `internal/web`): 22 cite `SPEC-KANBAN-*` identifiers of other SPECs (a SPEC id cannot be reworded), 57 name the package or a path under it (`internal/kanban/…`, `kanban.X`; M8's program edits code, so these remain after it unless M8 also rewrites comments), 78 are other prose. Estimated from the same line filter, M8 would clear import paths, package clauses and code qualifier lines and leave about 186 lines in 84 non-test files (an estimate, not a measurement of M8). The plan assigns the 13 doctrine-path lines to M7 and nothing else of the comment prose to any milestone, while AC-018 reads zero outside four files.

#### Baseline-attribution

- Trees and builds: the sizing, the baseline suites, and the net were measured on `d206491f3`; the rename ran on `48365b722` (HEAD after the sizing commit `30b2eab3c` and the program commit); the after-suites ran on the working tree that became `602aa8c93` (the harness comment edit made after the first hook run, and the env-literal guard re-run after it) and again on the committed tree `e41310611`; the AC rows above, the greps, the builds, the vets, and the idempotence runs are on `e41310611` with `git status --short` empty. HEAD, branch (`WT-launcher-entry-flags`) and `git status --short` were re-read before each commit.
- Judging build: every measurement is by the Go toolchain (`go1.26.8`) and shell tools run from this tree; no installed `moai` build produced any cited measurement, except `moai slot acquire` (lease line) and `moai slot release`.
- Commits (card t1399): `30b2eab3c` (sizing and baseline, docs), `48365b722` (the program), `602aa8c93` (the rename output plus the one hand comment and the guard test), `f4a809fcf` (hand comment rewording), `5513923a6` (program extended to follow file renames in comments), `e41310611` (its output), then this record. `git diff --stat d206491f3 HEAD` (on `e41310611`): `126 files changed, 1551 insertions(+), 662 deletions(-)`; the only non-`.go` file in it is this `progress.md`; `internal/web` has 0 changed files.

#### Gaps

1. `internal/cli` whole suite not run (about 5,100 tests; machine load average 72 at the time). Run: the factory net, the 86-test focused list (PASS), a 1,894-name list that timed out after 498 tests (above), `homestate` and `ptycaptest` whole. The rename is type-checked on both OSes (build and vet), and the behavior-visible edits (38 literal rewrites, 70 comment or string tokens) are covered only by those runs and by the five non-`cli` suites.
2. `go vet ./...` for the whole module, `GOOS=windows go vet ./...` for the whole module, `golangci-lint`, and any run on Windows were not done; vet ran over the 15 package patterns named above on both OSes.
3. AC-024's `TestLauncherHelpDocumentsLaneEntry` cannot be run (Findings 1); the two hook tests of that row pass.
4. The RED of the new guard (542 identifiers) was observed on a scratch clone of `48365b722` with the test file copied in, not on a commit: the guard test shares commit `602aa8c93` with the rename it measures, so the order is a session record and not a commit-graph fact (verification-claim-integrity section 2.3). The baseline and sizing are committed ahead (`30b2eab3c`).
5. The whole-suite hook counts are split across two runs by name range; the sum of the sub-test counts equals the baseline (1,934) and the name lists are identical modulo the renames, but a single-run figure was not taken.
6. The program was proven on this tree and on scratch clones of it; it was not run on an absorbed develop tree (the lane does that), so what other lanes added since the base is untested against it.
7. Worktree-guard refusals, two, each re-issued as plain commands measuring the same thing: (a) a heredoc that wrote the scratch Go program followed by other commands in one call, replaced by the Write tool and a separate `go run` of a literal path; (b) a `for` loop over file names that ran `awk` and named `suites_final_*` files in a computed form, replaced by separate `awk` calls. In addition one foreground call (the 1,894-name `internal/cli` list) exceeded the 600 second tool limit and the runtime moved it to the background; its output file was read once after the completion notice and nothing was polled; the later runs were sized to fit the limit.
8. Scratch clones used for the dry comparisons (`git clone --local` into the scratchpad and a scratchpad helper that runs one git command there) are outside the tree; the guard's own refusal of `-C` was not triggered because the helper runs git through its own process.

#### Residual-risk

- Test names that describe retired behavior now read oddly because the rule is a word replacement: `TestSessionRecordIgnoresRetiredFactoryMarkers`, `TestNonFactorySessionWritesNoRecord`, `TestCodexFactoryEntryIsRefused` (it asserts the retired `-k` entry is refused), `TestFactoryRoleFromEnvLegacyLabelsNotRecognized`. Only the four names acceptance commands select stay unchanged. The files `session_start_no_factory_notice_test.go` and `preexisting_factory_artifacts_m1_test.go` (hook, statusline) are named by the same rule while their tests assert the absence of retired kanban artifacts.
- The comment pass rewrites any camel-case token that spells an old identifier, including a comment that mentions a symbol that no longer exists anywhere (it now names a `factory…` symbol that does not exist either); the program prints its 21 distinct token pairs (read in the dry run before the real run), and none of them rewrites a marker value.
- Behavior-visible text changes in `internal/kanban` error prefixes (`todo queue …`, `factory: …`) were matched against tests by the whole `internal/kanban` suite only; a consumer outside Go that parses those messages was not searched.
- The new guard test scans `internal` and `cmd` outside `internal/web`; M9 must drop its web exclusion and M8 may drop the `kanban` qualifier allowance, otherwise it keeps passing without covering them.

#### Findings (statements the tree contradicts, no SPEC file edited)

1. acceptance.md AC-024 names `go test ./internal/cli -run '^TestLauncherHelpDocumentsLaneEntry$' -v -count=1` (swept count 1). `grep -rn TestLauncherHelpDocumentsLaneEntry internal --include='*.go'` found nothing, and `git log -S 'TestLauncherHelpDocumentsLaneEntry' --oneline -- internal` printed nothing: no commit on the base or this branch ever carried the test, so M2-M6 did not author it and the criterion's help-test half sweeps zero tests.
2. design.md section 4.7 counts "13 comment lines in 11 non-test Go files" that name `kanban-dispatch.md` or `moai-kanban-foreman`; on this tree outside `internal/web` there were 11 lines in 9 files (`token_budget_guard.go` 3, `factory_card.go`, `codex_review_scope.go`, `todo_edit_move.go`, `todo_auto.go`, `todo_drop.go`, `integration.go`, `kanban/integration_lock.go`, `hook/lane_spawn_authority.go`), plus the one in `internal/web/viewmodel_ops.go:46` (M9, deleted with `ChainRoles`); `session_start_factory.go` no longer carries one.
3. AC-018 reads zero outside four files, but after M7 the non-web residue that no milestone is assigned is 157 comment lines and 20 string lines (section above), and AC-018's `SPEC-KANBAN-*` citations (22 lines here) are ids of other SPECs. The plan's M7-M10 text does not say who rewords them.
4. design.md section 4.7 says the statusline label "Kanban backlog" is a label at `internal/statusline/types.go:252,350`; both are comments, not rendered text, so the change is two comment edits (done, `f4a809fcf`).
5. `ReadCardStatus`, `FactoryFreeSlots`, `rejectFactoryOnCG`, and `config.DefaultFactoryLanes` were not deleted (not in the plan's lists).
6. plan.md M7 and the probe rename every identifier that carries the word, while four acceptance commands select tests by names that carry it: AC-011 `TestKanbanEntryRefused`, AC-014 `TestSessionStartEmitsNoKanbanNotice`, AC-017 `TestPreexistingKanbanArtifactsTolerated` (three packages), AC-019 `TestLegacyKanbanRouteRedirects` (`internal/web`, M9). A mechanical rename of them would leave those commands sweeping zero tests (the probe checks compilation only, so it did not show this). The program keeps the four names; acceptance.md was not edited. The same reasoning applies to AC-018's `find`, which covers test file names: the files that hold those tests were renamed (the cli test lives in `launcher_retired_entries_test.go` and `launcher_characterization_m1_test.go`, the hook and statusline files carry `factory` in their names now).

### M8 evidence

Recorded by the run-phase implementation worker (cycle_type ddd: the milestone is a behavior-preserving mechanical rename of the Go package `internal/kanban` to `internal/factory`; characterization is the factory net and the whole-package suites) for milestone M8 of card t1399, branch `WT-launcher-entry-flags`. Start state, re-read before any change: `pwd` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1399`, `git rev-parse --short HEAD` printed `4e7a60232`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Every output block below is a command run in this run on this tree (HEAD `4e7a60232`), with what it drops stated. The sections grow in the commits of this milestone: this commit records the sizing and the baseline; later commits add the program, its results, the hand fixups, and the self-verification.

#### Sizing of AC-018's package and import sweeps on this tree (measured before any change)

```text
$ grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'      (RED-N3 command; files counted by a separate | wc -l)
183 files, exit 0
$ grep -rn 'kanban\.' internal cmd --include='*.go' | wc -l      (text lines that contain "kanban."; comments and strings included, so an upper bound of the code qualifiers)
1695
$ grep -rlP '(?i)(?<!moai_)kanban|moai_kanban(?!_(id|lead_addr|lead_name|settings_injected|backend|card)\b)|칸반|かんばん|カンバン|看板' internal cmd --include='*.go' --include='*.templ' --include='*.js' --exclude='*_test.go' --exclude-dir=testdata --exclude-dir=node_modules | wc -l      (RED-N1 command)
169
$ grep -rIn "internal/kanban" internal cmd --include='*.go' | wc -l      (every line that spells the path: 183 import lines plus 69 others)
252
$ find internal cmd -iname '*kanban*'      (RED-N2 command, M7 cleared the 18 names inside cli/hook/kanban except the directory)
internal/kanban
internal/web/preexisting_kanban_artifacts_m1_test.go
internal/template/templates/.claude/skills/moai-kanban-foreman
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md
$ ls internal/kanban/*.go | wc -l ; grep -L '^package kanban$' internal/kanban/*.go ; grep -l '^package kanban_test' internal/kanban/*.go | wc -l
159 files, no file with another clause, 0 external-test-package files
$ grep -rn "package kanban" internal cmd --include='*.go' | grep -v '^internal/kanban/'
(no output: no package clause outside the directory)
```

Importers by shape: no aliased import (`grep -rn '^\s*[a-zA-Z_]* "github.com/modu-ai/moai-adk/internal/kanban"'` printed nothing), no dot import, no `.templ` or `.js` source imports the path; the only non-Go mentions of `internal/kanban` outside `internal/` and `cmd/` are `CHANGELOG.md`, `.claude/rules/local/gitflow-lane-protocol.md`, the generated codemaps under `.moai/project/codemaps/` (5 files), and `docs-site/content/*/advanced/kanban-mode.md` (the docs belong to M11; the codemaps stay stale until `/moai codemaps`, plan.md section E).

Syntax-tree measurement by the M8 rename program in dry-run mode (program text not yet committed at this point; the same program is committed in the next commit):

```text
$ go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m8_rename.go -root . -dry-run
m8_rename: package clauses: 159; import lines: 183; qualifiers: 1894; shadowing locals renamed: 4; string literals: 35; comment mentions: 41; files rewritten: 361; files moved: 159 (dry run, nothing written)
```

Against the plan's modeled numbers (probe stage M8 on the pinned base: 156 package clauses, 1,830 qualified references, 4 shadowing locals): 159 clauses (three package files of the live tree are newer than the model), 1,894 qualifiers (this count covers `internal/web` as well; the M7 residue count of 1,980 bare identifiers `kanban` was taken outside `internal/web`, so the two are not the same population and were not reconciled line by line; `grep -rn 'kanban\.[A-Za-z]' internal/web --include='*.go' | wc -l` printed 113 text lines, comments included), 4 shadowing locals as modeled. `-v` on the same dry run lists every string-literal and comment rewrite (35 and 41); the four shadowing occurrences are the local `factory` at `internal/hook/stale_run_m1_test.go` lines 118, 119 (two) and 120 (the plan's site; `grep -n '\bfactory\b' internal/hook/stale_run_m1_test.go` shows them).

#### Baseline before any change (characterization)

Selected by whole package, one compound invocation each, with the environment scrub (`unset MOAI_KANBAN … MOAI_FACTORY_WORKERS && go test … -count=1 -v`), output to a scratch file, counts by `grep -c '^--- PASS'` (top-level test lines; sub-tests are indented and not counted):

```text
internal/kanban   (whole)  ok  232.294s   top-level PASS 514, FAIL 0, SKIP 0 ; === RUN lines 742
internal/config            ok   6.008s    PASS 488   (small-suite run: config discovery codexwiring statusline factorymsg factorylane homestate)
internal/discovery         ok             PASS 14, SKIP 1
internal/codexwiring       ok             PASS 91
internal/statusline        ok  34.221s    PASS 333, SKIP 2
internal/factorymsg        ok  85.045s    PASS 70, SKIP 1
internal/factorylane       ok   4.363s    PASS 53
internal/homestate         ok  78.543s    PASS 133                                (the seven sum to 1182 PASS lines)
internal/web      (whole)  ok  40.883s    PASS 536, FAIL 0, SKIP 7
internal/hook     (whole)  FAIL 477.812s PASS 1269, FAIL 1, SKIP 6   the one failure is TestHookWrapperCopiesStayIdentical (known base failure, unrelated)
TestNoNewEnvLiteralsInDiff PASS: "env-literal sweep: 1167 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2, distinct literals=0"
```

The factory net (AC-015) on this tree, three commands, exit 0 each: `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestPrepareKanbanSettingsWritesTransientFile|TestPrepareFactorySettingsWritesTransientFile|TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch|TestFactoryNetBlockCap|TestFactoryEntryMatrix)$' -v -count=1` printed eight `--- PASS` lines (`TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestFactoryNetLeaderLaunch`, `TestFactoryNetLaneLaunch`, `TestFactoryNetBlockCap`, `TestFactoryEntryMatrix`, `TestPrepareFactorySettingsWritesTransientFile`, `TestCCFactoryEntryRecordsFailOpenRunMetadata`) and `ok ... internal/cli 28.068s` (swept 8); the hook command printed `--- PASS: TestFactoryNetSessionRecord`, `--- PASS: TestFactoryNetSessionStartNotices` and `ok ... internal/hook 1.698s` (swept 2); the discovery command printed `--- PASS: TestDiscoverLeaderVerifiesLiveLeader`, `--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID` and `ok ... internal/discovery 0.833s` (swept 2).

AC-016's four commands on this tree, each exit 0: `TestFactoryMarkerValuesFrozen` (internal/config, PASS), the two discovery tests above, `TestMCPServerEnvVarsKeepFactoryMarkers` (internal/codexwiring, PASS), `TestLegacyStateDirStillRead` (`./internal/kanban`, PASS, swept 1).

Resource lease: `moai slot acquire --resource t1399-run --max-duration 60m` printed `slot t1399-run acquired by 0dcdf2d5-df5c-4da1-8870-24c2a5861303 until 2026-10-02T17:56:46Z`; it is released before the milestone's last commit.

#### The rename program (committed, re-runnable)

Path: `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m8_rename.go` (`//go:build ignore`, standard library only; an identical copy sits in the gitignored `.moai/reports/t1399/rename/`; the committed file is the source of truth). Run command, from the tree root, on a tree whose tracked files are committed:

```text
go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m8_rename.go -root <git tree> [-dry-run] [-v]
```

What it does, in order (the header comment of the file is the full statement): (1) moves every file of `internal/kanban` to `internal/factory` with `git mv` for tracked files and a plain rename for untracked ones, refusing when a target already exists, and removes the emptied directory; (2) in every Go file that is inside the moved package or imports the old path, by syntax-tree position: rewrites the package clause (`kanban` and `kanban_test`), the import path (an aliased import keeps its alias), every qualifier `kanban.X` of an unaliased import (an identifier that resolves to a local declaration is left alone), and a local variable, constant, or parameter named `factory` to `factoryRun` in a file that imports the old path (the shadow rule); in string literals and comments, the package path spellings `internal/kanban` and `../kanban` (not when preceded by a colon, a git revision spelled `rev:path`) and the package-qualified exported name `kanban.Upper...` (so a reflect type name or the text a source-scanning test searches for stays true; an i18n key such as `kanban.noSession` and a file name such as `kanban.go` have a lower-case continuation and stay); in string literals only, the two adjacent arguments `"internal", "kanban"` of a path join and the home-state coverage key `"kanban"` of `internal/cli/home_state_coverage.go`; in comments only, the package doc line `Package kanban`; (3) gofmts every file it rewrote (go/format, which re-sorts the import specs inside a block). Refusals: exit 2 on a tree with tracked modifications (the message names the first ten), exit 1 with nothing written on a move collision, a rewrite that does not parse, a second import named `factory` next to the old path, a package-level `factory` declaration in an importing package, or an existing `factoryRun` in a file that needs the shadow rename.

It never touches: non-Go files, `testdata`, `node_modules`, `vendor`, `.git`, `.moai`, `.claude`, `internal/template/templates`; the six marker string values and the legacy state-directory name (`.moai/state/kanban`, a data path whose literal is `"kanban"` and does not follow `"internal"`); the string literals and comments of `internal/graph/codemaps_fold_guard_test.go` (its unit paths name entries of the stale generated codemaps record, a data file this program does not own: rewriting them would turn `TestCodemapsFoldGuardFloorCoverage`'s floor red against the unchanged record); the other prose that says the word (comments such as `kanban's own`).

Observed results, each a command run in this run:

| What | Command, tree | Observed |
|---|---|---|
| dry run | `-dry-run` on `4e7a60232` plus the uncommitted program | `m8_rename: package clauses: 159; import lines: 183; qualifiers: 1894; shadowing locals renamed: 4; string literals: 35; comment mentions: 41; files rewritten: 361; files moved: 159 (dry run, nothing written)` |
| rehearsal on a scratch clone (`git clone --local` into the scratchpad, program run with `-root`) | clone of `ac61e0b06` | the same counts without the suffix; `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and `go vet ./...` of the clone exit 0 with no output |
| the real run, commit `8a7d60ba2` | the run command on `60a0fd5e8` | `m8_rename: package clauses: 159; import lines: 183; qualifiers: 1894; shadowing locals renamed: 4; string literals: 35; comment mentions: 41; files rewritten: 361; files moved: 159`; then `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` printed nothing and exited 0, `go vet ./...` produced a 0-byte output file |
| the first whole `internal/factory` run on that tree | `go test ./internal/factory -count=1 -v` | one red: `--- FAIL: TestBacklogArchive_PerItemContractFrozen`, `backlog_archive_test.go:120: declared addition "Landing" has type *factory.LandingEvidence, want *kanban.LandingEvidence` (the test compares reflect type names spelled as strings) — the program was extended (commit `7896aae99`) with the `kanban.Upper...` rule |
| second run, commit `03c6b6404` | the run command on `7896aae99` | `package clauses: 0; import lines: 0; qualifiers: 0; shadowing locals renamed: 0; string literals: 18; comment mentions: 41; files rewritten: 33; files moved: 0` |
| idempotence on the final tree | the run command on `304692b8d`, then `git status --short \| wc -l` | all zeros (`… files rewritten: 0; files moved: 0`) and `0` |
| refuses a dirty tree | the run command on a scratch clone of `304692b8d` holding a staged `git mv internal/cli/cc.go internal/cli/cc_moved.go` | `m8_rename: the tree has 1 tracked modification(s); commit or revert them first (first 1 shown):` / `R  internal/cli/cc.go -> internal/cli/cc_moved.go` / `exit status 2` |
| converts planted late references | scratch clone of `304692b8d` with two untracked files: `internal/web/planted_late_test.go` (imports the old path, calls `kanban.RoleLeader` and `kanban.RoleLane`, declares a local `factory`) and `internal/kanban/planted_late_test.go` (`package kanban`, uses `RoleLeader`) | before the run `go vet ./internal/web ./internal/kanban` printed `github.com/modu-ai/moai-adk/internal/kanban: no non-test Go files in …` and `vet: internal/kanban/planted_late_test.go:7:5: undefined: RoleLeader`; the run printed `package clauses: 1; import lines: 1; qualifiers: 2; shadowing locals renamed: 2; string literals: 0; comment mentions: 0; files rewritten: 2; files moved: 1`; the web file read back with `"github.com/modu-ai/moai-adk/internal/factory"`, `factory.RoleLeader`, `factoryRun := factory.RoleLane`; the second file is at `internal/factory/planted_late_test.go` with `package factory`; `go build ./...` of the clone printed `CLONE_BUILD_OK`; `go vet ./internal/web ./internal/factory` exit 0; `go test ./internal/web ./internal/factory -run '^(TestPlantedLateReference\|TestPlantedLateInPackage)$' -v -count=1` printed `--- PASS: TestPlantedLateReference` and `--- PASS: TestPlantedLateInPackage` |

Plan versus actual: 159 package clauses against the model's 156 (the live tree carries three more package files); 1,894 qualifiers against 1,830 (the live tree is larger, and this count covers `internal/web`); 4 shadowing locals as modeled (`stale_run_m1_test.go` lines 118-120); the plan names path strings in `home_state_coverage.go` and `migrate_home_state.go`: they are in the 35 literals (3 and 4 rewrites; `-v` lists each with file and line).

Not covered by the program (hand edits, commit `304692b8d`, and what a post-absorption re-run leaves to a manual read): the identifier guard `internal/cli/retired_word_identifiers_m7_test.go` (the M7 allowance of the bare package name `kanban` is dropped, its positive control now plants a stale bare qualifier and expects two reports; with the allowance dropped the scan reported `internal/hook/role_naming_m3_notice_test.go:134,135,136,138,139 kanban`, a local variable of that name, which the M7 allowance had hidden — renamed `relaunch` by hand); the comment of `internal/factory/legacy_state_dir_m1_test.go` lines 5-7 (the path rewrite turned "through M7 the package path is `./internal/kanban`" into a sentence that named the new path twice; reworded). The RED of the dropped allowance (5 reports) was observed on the working tree before the fix; the guard edit and the fix share commit `304692b8d`, so that order is a session record and not a commit-graph fact (verification-claim-integrity section 2.3); the baseline and the program are committed ahead (`ac61e0b06`, `60a0fd5e8`).

Verification after absorbing develop (what the lane runs; the program has no knowledge of which files other lanes added): (1) merge develop into the branch and commit the merge, so tracked files are clean; (2) `go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m8_rename.go -root .` and commit its output; a non-empty summary means other lanes added references, and `-v` lists the literal and comment rewrites to read; (3) `grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'` must print nothing (exit 1) and `ls internal/kanban` must fail; (4) `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` must pass (exit 0, no output); (5) `go vet` of the packages the merge touched on both OSes, and `go test ./internal/cli -run '^TestNoRetiredWordIdentifiersOutsideWeb$'` (a late file that spells a bare `kanban` identifier fails it); (6) run the program a second time: all zeros and `git status --short` empty; (7) the factory net (8 cli, 2 hook, 2 discovery). Cases the program does not see and the build does: a late test that scans source text for a spelling other than `kanban.Upper...`, and a local named `factory` used as a composite-literal key in a file that imports the old path.

#### Claim

Milestone M8 of card t1399 is implemented on branch `WT-launcher-entry-flags` at `304692b8d`: the Go package `internal/kanban` is `internal/factory` (directory, `package factory`, import path of 183 files, 1,894 qualifiers, 4 shadowing locals renamed), the run-time package-path strings and the home-state coverage key follow, no file imports the old path, and the build, the windows build, and the vet pass. The factory net, the AC-016 commands, and every whole-package suite measured before the change give the same counts after it, except where stated below. Commits (card t1399): `ac61e0b06` (sizing and baseline, docs), `60a0fd5e8` (the program), `8a7d60ba2` (its output on the tree), `7896aae99` (program extension), `03c6b6404` (its second output), `304692b8d` (hand fixups), then this record. `git diff --stat 4e7a60232 HEAD | tail -3` at `304692b8d`: `369 files changed, 2744 insertions(+), 2106 deletions(-)` (the last two file lines were `internal/web/todo_section_test.go | 16 +-` and `internal/web/viewmodel_ops.go | 8 +-`).

#### Evidence

AC matrix, each row a command run on tree `304692b8d` (or the tree named):

| Criterion | Command | Observed |
|---|---|---|
| AC-018 importer grep | `grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'; echo "grep_exit=$?"` | no file, `grep_exit=1` (RED-N3: 183 files before) |
| AC-018 build | `go build ./...` | printed `BUILD_HOST_OK` after `&& echo` (exit 0, no other output) |
| AC-018 windows build (REQ-024) | `GOOS=windows GOARCH=amd64 go build ./...` | printed `BUILD_WINDOWS_OK` (exit 0, no other output) |
| AC-018 vet | `go vet ./...` redirected to a file | the file is 0 bytes, no error status |
| windows vet of the touched packages | `GOOS=windows GOARCH=amd64 go vet ./cmd/t657-merge ./internal/cli ./internal/config ./internal/discovery ./internal/escalation ./internal/factory/... ./internal/factorymsg ./internal/graph ./internal/homestate ./internal/hook ./internal/mission ./internal/session ./internal/spec ./internal/stateanchor ./internal/statusline ./internal/template ./internal/web ./internal/factorylane ./internal/codexwiring` piped to `tail -5` | no output (the pre-existing failure of `internal/cli/worktree/sweep_test.go:1687` is outside this list) |
| AC-018 find (RED-N2) | `find internal cmd -iname '*kanban*'` | 5 names, the directory `internal/kanban` is gone: `internal/web/preexisting_kanban_artifacts_m1_test.go` (M9) and the four template paths (M10) |
| AC-018 word grep (RED-N1, by file) | the AC-018 pattern over non-test `.go`, `.templ`, `.js` | 83 files, 16 in `internal/web`, 67 elsewhere (169 before); not flipped, owners in the next section |
| AC-018 per-file bound | `grep -c -i kanban internal/cli/launcher_retired_entries.go internal/factory/state_dir.go` | 4 and 1 (`state_dir.go` was 2 in `internal/kanban/state_dir.go` at the base: `git show 4e7a60232:internal/kanban/state_dir.go \| grep -c -i kanban` printed 2) |
| AC-016 value test | `go test ./internal/config -run '^TestFactoryMarkerValuesFrozen$' -v -count=1` | `--- PASS: TestFactoryMarkerValuesFrozen`, `ok … internal/config` |
| AC-016 discovery | the two discovery tests, `-v -count=1` | `--- PASS: TestDiscoverLeaderVerifiesLiveLeader`, `--- PASS: TestDiscoverLeaderDeclinesUnparseableRunID`, `ok … internal/discovery` (swept 2) |
| AC-016 allowlist | `go test ./internal/codexwiring -run '^TestMCPServerEnvVarsKeepFactoryMarkers$' -v -count=1` | `--- PASS: TestMCPServerEnvVarsKeepFactoryMarkers`, `ok … internal/codexwiring` |
| AC-016 legacy dir (new path) | `go test ./internal/factory -run '^TestLegacyStateDirStillRead$' -v -count=1` | `--- PASS: TestLegacyStateDirStillRead (0.05s)`, `ok … internal/factory` (swept 1; the baseline run on `./internal/kanban` printed the same) |
| AC-015 net, cli | the 9-name alternation of AC-015, `-v -count=1` | eight `--- PASS` lines (`TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestFactoryNetLeaderLaunch`, `TestFactoryNetLaneLaunch`, `TestFactoryNetBlockCap`, `TestFactoryEntryMatrix`, `TestPrepareFactorySettingsWritesTransientFile`, `TestCCFactoryEntryRecordsFailOpenRunMetadata`), `ok … internal/cli 20.984s` (swept 8) |
| AC-015 net, hook | `TestFactoryNetSessionRecord`, `TestFactoryNetSessionStartNotices` | both `--- PASS`, `ok … internal/hook` (swept 2) |
| AC-015 net, discovery | the AC-016 discovery command above | swept 2 |
| env-literal guard | `go test ./internal/hook -run '^TestNoNewEnvLiteralsInDiff$' -v -count=1` | `env-literal sweep: 1732 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2, distinct literals=0`, `--- PASS` (baseline 1167 lines, 0) |
| identifier guard | `go test ./internal/cli -run '^(TestRetiredWordIdentifierScanHasTeeth\|TestNoRetiredWordIdentifiersOutsideWeb)$' -v -count=1` | both `--- PASS` |

Whole-package suites, before (tree `4e7a60232`) and after (tree `03c6b6404`, the program's second output; the hand fixups of `304692b8d` touched three test files and are covered by the rows above), top-level `--- PASS` lines by `grep -c`/`awk`:

```text
package            before                                        after
internal/kanban -> internal/factory  PASS 514 FAIL 0 SKIP 0 (742 RUN lines, ok 232.294s)   PASS 514 FAIL 0 SKIP 0 (742 RUN lines, ok 219.852s)
internal/config        PASS 488                                  PASS 488
internal/discovery     PASS 14 SKIP 1                            PASS 14 SKIP 1
internal/codexwiring   PASS 91                                   PASS 91
internal/statusline    PASS 333 SKIP 2                           PASS 333 SKIP 2
internal/factorymsg    PASS 70 SKIP 1                            PASS 70 SKIP 1
internal/factorylane   PASS 53                                   PASS 53
internal/homestate     PASS 133                                  PASS 133
internal/web           PASS 536 FAIL 0 SKIP 7                    PASS 536 FAIL 0 SKIP 7
internal/hook          PASS 1269 FAIL 1 SKIP 6 (477.812s)        PASS 1269 FAIL 1 SKIP 6 (413.066s)   the one failure is TestHookWrapperCopiesStayIdentical, unrelated
internal/graph         base clone: PASS 141 FAIL 2               PASS 141 FAIL 2   TestCodemapsFoldPreservationGuard and TestCodemapsFoldGuardFixtures, red on the base commit too (below)
internal/escalation    base clone: PASS 56                       PASS 56
internal/session       base clone: PASS 161 SKIP 1               PASS 161 SKIP 1
internal/stateanchor   base clone: PASS 13                       PASS 13
internal/spec          base clone: PASS 592 SKIP 5               PASS 593 SKIP 4   (the base was measured in a scratch clone, not this worktree; one test that skips there ran here)
internal/mission       base clone: PASS 44                       PASS 44
```

The `internal/graph` failures were reproduced on the base commit: `go test ./internal/graph -count=1 -run 'TestCodemapsFold'` in a scratch clone checked out at `4e7a60232` printed `--- FAIL: TestCodemapsFoldPreservationGuard (0.00s)`, `--- FAIL: TestCodemapsFoldGuardFixtures (0.02s)`, `FAIL … internal/graph`; their text (`fold unit "internal/hook/cwd_changed_relocate.go" appears in overview.md:8`) is about other units of the codemaps record, not this package.

`internal/cli` (the whole suite does not finish in the foreground): the test functions of the 92 cli test files this milestone changed (652 `func Test…` names, derived from `git diff --name-only 4e7a60232 HEAD -- internal/cli`), run in six anchored-name invocations, each `go test ./internal/cli -count=1 -v -timeout 560s -run '^(…)$'`: factory family 139 names `ok … 249.989s PASS=140`; todo family 299 names in two invocations `ok … 160.297s PASS=127 SKIP=1` and `ok … 373.690s PASS=172`; the rest in three invocations `ok … 127.404s PASS=97`, `ok … 79.420s PASS=34` (the home-state tests) and `ok … 145.123s PASS=80`; no `--- FAIL` in any. The lists were checked against the 652 names with `comm -23`: the only differences were four names at list boundaries that my own `cat` of the three files joined without a separator (`TestWriteExportAtomic` next to `TestAdopt…`, `TestTodoNextHelpLeaderAndLanePromotion` next to `TestAC_CLOSURE_001`), all four present in their lists. The PASS counts exceed the name counts by one in two lists (140 for 139 names; 299 PASS plus one skip for 299 names) and were not reconciled test by test.

#### Remaining after M8: what still carries the word outside `internal/web`, and the owner (baseline for M9 and M10)

Measured on `304692b8d` by a scratch syntax-tree program (not committed; its output is in the gitignored scratchpad):

```text
identifier occurrences: 5   (1 TestKanbanEntryRefused, 3 TestPreexistingKanbanArtifactsTolerated, 1 TestSessionStartEmitsNoKanbanNotice)   -> kept on purpose (acceptance commands name them)
string literals carrying the word: 173 = 29 in non-test files + 144 in test files
non-test, 29:  frozen marker values 11 (envkeys.go 6, codexwiring/configtoml.go 1, hook/session_start_factory_i18n.go 4 naming MOAI_KANBAN_CARD)   -> frozen (AC-016)
               allowed files 3 (cli/launcher_retired_entries.go 2, factory/state_dir.go 1)                                                         -> allowed (REQ-015, REQ-017)
               no owner 15: cli/doctor_factory_run.go:50 (1), cli/factory_launch_helpers.go:212,257 (2), factory/record.go (9), factory/record_prune.go (1), factory/state_lock.go:24,35 (2)
test files, 144: marker names, the legacy kanban-board directory, fixtures that quote history ("fix(kanban): …"), the codemaps fold-guard unit paths, the skill id moai-kanban-foreman, messages that say "kanban"
comment lines carrying the word: 266 in 148 files (36 of them name SPEC-KANBAN-* identifiers of other SPECs); test files included, so not comparable to M7's 157 non-test lines
file names (find): internal/web/preexisting_kanban_artifacts_m1_test.go (M9); the foreman skill directory and three rule files under internal/template/templates (M10)
AC-018 word grep by file (non-test .go .templ .js): 83 files = 16 in internal/web (M9) + 67 elsewhere (frozen values, allowed files, the 15 no-owner strings, and comment-only files)
```

Owners: the web package, its template sources, assets, i18n keys, route, SSE key, and the web test name are M9; the template mirror paths and the one `update_archive.go` line are M10; docs are M11. The 15 unowned non-test string lines (error-message prefixes `write kanban record:`, `read kanban record(s):`, `prune kanban records:`, `kanban board lock …`, two launcher diagnostics beginning `kanban: `, and the doctor line `no factory or kanban run declared`) and the unowned comment prose have no milestone in plan.md M7-M10 (the M7 section recorded the same list; the counts are unchanged except that the `home_state_coverage.go` and `migrate_home_state.go` strings, M8's, are gone). The generated codemaps under `.moai/project/codemaps/` still name the old path and stay stale until `/moai codemaps` (plan.md section E); `internal/graph/codemaps_fold_guard_test.go` keeps its old-path literals for that reason.

#### Baseline-attribution

- Trees and builds: the sizing and the baseline suites were measured on `4e7a60232` (this worktree, HEAD at the start); the rename on `60a0fd5e8` and `7896aae99`; the whole-package suites after on the tree of `03c6b6404`; the AC rows, greps, builds, vets, the net, and the idempotence run on `304692b8d` with `git status --short` empty. HEAD, branch (`WT-launcher-entry-flags`), and `git status --short` were re-read before each commit.
- Judging build: every measurement is by the Go toolchain and shell tools run from this tree; `moai slot acquire` and `moai slot release` were the only invocations of the installed `moai` binary.
- The base counts of `internal/graph`, `escalation`, `session`, `stateanchor`, `spec`, and `mission` were taken in a scratch clone of this worktree checked out at `4e7a60232` (a different directory from the after counts).

#### Gaps

1. `internal/cli` whole suite not run; the changed test files' 652 test functions ran in the six invocations above (families by file-derived name lists), the other cli tests did not. `internal/template` ran only `TestContractModeChangeSetAllowlist` (`--- PASS`; its `tree` subtest `--- SKIP`, needs `MOAI_GR_BASE`), not the whole package; `go vet ./...` host and the windows vets named above were run, `golangci-lint` was not, and no test was run on Windows.
2. The whole `internal/hook` and `internal/factory` suites ran before the three hand edits of `304692b8d` and were not re-run after them; the edited tests, the net, AC-016, the env-literal guard, and the identifier guard were.
3. `internal/spec` differs by one test (592 PASS 5 SKIP at the base clone, 593 PASS 4 SKIP here): the two measurements were taken in different directories and the skip was not investigated.
4. Worktree-guard refusals, two, each re-issued as plain commands measuring the same thing: (a) `go -C $S/clone1 vet ./...` with a shell variable for the scratch path, refused with "runs go with a value computed at runtime" — re-issued with the literal path; (b) the 299-name todo-family `go test -run` (about 11 KB) refused as "too complex to verify" — split into two invocations of the same names. A command that carried `; echo "grep_exit=$?"` was accepted.
5. Scratch clones (`git clone --local` into the scratchpad) and a scratchpad helper that runs one git command in a clone were used for the rehearsal, the dirty-tree refusal, the planted-reference proof, and the base measurements; none is a worktree, and no `git worktree add` was run.
6. The RED of the dropped identifier allowance (5 reports) and the first red of `TestBacklogArchive_PerItemContractFrozen` were observed on working trees, not on commits (see above).
7. `-count` PASS figures of the cli lists are not reconciled to the name counts by one (above).
8. The vacuity argument for three source-scanning tests (`internal/cli/todo_hold_predicate_test.go`, `todo_hold_test.go`, `internal/web/todo_queue_read_test.go`) is an inference from reading them: they search source text for `kanban.…` spellings, so after the rename they would search for text that no longer exists; the tests were not run before the program learned the rule.

#### Residual-risk

- Free text outside Go sources that spells the old package path was not touched (CHANGELOG, local rules, docs-site, codemaps); a script or a human reading them finds a path that no longer exists until M10/M11 or `/moai codemaps`.
- The program rewrites by rule; a late reference of a shape the rules do not know (a source-text scan for a spelling other than `kanban.Upper...`, a `factory` used as a composite-literal key) compiles or fails at the build and the suites, not in the program.
- The comment rule rewrites `kanban.Upper...` in comments of `internal/web` too, so M9's web diff will not show those comment lines as its own.
- Test and error-message text that says `kanban.X` in strings is now `factory.X`; anything outside Go that parses these messages was not searched.

#### Findings (statements the tree contradicts, no SPEC file edited)

1. design.md section 4.7 names `queue_path_seam_scan_test.go:41-44` for the run-time path strings; on this tree the split literal is `filepath.Join("internal", "kanban", "state_dir.go")` at line 34 and the message with the path at line 89 (the program rewrote both).
2. Neither design 4.7 nor plan M8 names three kinds of spelling the move breaks: sibling-relative paths (`"../kanban"` at `internal/cli/factory_m7_test.go:315`, `vocabulary_guard_test.go:37,50,51,52`, `integration_settings_drift_exitcode_test.go:28,29,64,65`), reflect type names spelled as strings (`internal/factory/backlog_archive_test.go:64,73,92`, red on the first run), and source-text scans for `kanban.X` (Gap 8). The program covers all three by rule.
3. plan.md M8 and AC-018 count the importers at 179 and the references at 1,689/1,830; on `4e7a60232` the measured numbers are 183 importers and 1,894 qualifiers (the plan's models predate later merges).
4. AC-018's `find` reads 5 names after M8, not the 4 template paths alone: `internal/web/preexisting_kanban_artifacts_m1_test.go` carries the word in its file name and is not in plan.md M9's rename list (M9 renames the web identifiers and sources; this test file name is not mentioned).

### M9 evidence

Recorded by the run-phase implementation worker (cycle_type ddd: the milestone is a behavior-preserving rename of the web console surface plus the decided removals, the chain session board and its view model, and the decided route behavior, `GET /kanban` redirects to `/factory`; characterization is the whole `internal/web` suite and the factory net) for milestone M9 of card t1399, branch `WT-launcher-entry-flags`. Start state, re-read before any change: `pwd` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1399`, `git rev-parse --short HEAD` printed `8d50c79d6`, `git branch --show-current` printed `WT-launcher-entry-flags`, `git status --short` printed nothing. Every output block below is a command run in this run on this tree (HEAD `8d50c79d6` until the first commit of this milestone), with what it drops stated. The sections grow in the commits of this milestone: this commit records the sizing, the baseline, and the RED-now of AC-019; later commits add the test, the program, its results, the hand fixups, and the self-verification.

#### Sizing on this tree (measured before any change)

```text
$ grep -rlP '(?i)(?<!moai_)kanban|moai_kanban(?!_(id|lead_addr|lead_name|settings_injected|backend|card)\b)|칸반|かんばん|カンバン|看板' internal/web      (every file, tests included)
34 files (counted by a separate | wc -l), 328 lines (grep -rIP ... | wc -l)
$ the AC-018 command restricted to internal/web (--include='*.go' --include='*.templ' --include='*.js' --exclude='*_test.go' --exclude-dir=testdata --exclude-dir=node_modules)
internal/web/todo_queue_read.go
internal/web/shell.templ
internal/web/shell_templ.go
internal/web/screens_templ.go
internal/web/factory_lanes.go
internal/web/events.go
internal/web/widgets.templ
internal/web/app.go
internal/web/widgets_templ.go
internal/web/screens.templ
internal/web/viewmodel_ops.go
internal/web/icons.templ
internal/web/icons_templ.go
internal/web/screens.go
internal/web/assets/i18n.js
internal/web/assets/app.js
$ find internal -iname '*kanban*'
internal/web/preexisting_kanban_artifacts_m1_test.go
internal/template/templates/.claude/skills/moai-kanban-foreman
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md
$ grep -c 'kanban' internal/web/assets/i18n.js ; grep -c '"kanban\.' internal/web/assets/i18n.js
84 ; 72        (the key family; 86 lines by case-insensitive count, the other two carry the word in a value: "Kanban chain", "Open in Kanban")
$ grep -c 'data-live="kanban"' internal/web/screens.templ ; grep -c '"/kanban"' internal/web/app.go ; grep -n '"kanban"' internal/web/events.go internal/web/assets/app.js internal/web/icons.templ internal/web/screens.go
4 ; 1 ; events.go:35 (watch map key), events.go:223 and :225 (two path registrations), assets/app.js:614 (EVENTS list), icons.templ:72 (case "kanban"), screens.go:141 (area id)
Go and templ identifiers carrying the word: KanbanRecord 25, writeKanbanRecord 16, loadKanbanRecords 16, Kanban 13, KanbanVM 12, kanbanBodyFor 11, buildKanban 4, handleKanban 2, and six test names (TestTodoSectionCarriesExistingKanbanMarker, TestKanbanRoleWithNoTelemetryRecord, TestKanbanPipelineColumns, TestKanbanNoteBannerCorrected, TestKanbanLaneStates, TestKanbanChainRoleStates, two occurrences each) plus the pinned TestPreexistingKanbanArtifactsTolerated (1); these counts are text occurrences from `grep -o` over *.go and *.templ including the generated *_templ.go files, so they are an upper bound of the identifier count the syntax-tree program reports.
```

None of the five tests AC-019 selects exists on this tree: `grep -rn 'TestLegacyKanbanRouteRedirects\|TestFactoryScreenOmitsChainBoard\|TestFactoryScreenStillShowsFactoryLanes\|TestTodoScreenStillLive\|TestWebLiveKeyContract' internal` finds only the pinned-name entry of `internal/cli/retired_word_identifiers_m7_test.go:35`. AC-019's RED-now commands on this tree (the ledger's RED-W1 and RED-W2):

```text
$ grep -c 'Chain session board' internal/web/screens.templ
1                                    (exit 0)     RED-W1: the chain board panel exists
$ grep -c '"/kanban"' internal/web/app.go
1                                    (exit 0)     RED-W2: the route is registered as a screen
```

The recipes the program must keep working are in the section the program commit adds. The tools: the generated files are produced by the Makefile target `templ-generate`, whose command is `go run github.com/a-h/templ/cmd/templ generate -path ./internal/web` (Makefile line 32; `go.mod` pins `github.com/a-h/templ v0.3.1020` as a `tool` directive and the module is in the module cache, so no network is needed); run on this clean tree it printed `(✓) Complete [ updates=0 duration=123.138583ms ]` and `git status --short` printed nothing afterwards (the base's generated files are in sync with their sources). `golangci-lint` is `v2.1.6` (the CI version) and `golangci-lint run --timeout=5m ./internal/web/...` printed `0 issues.` on this tree.

#### Baseline before any change (characterization)

One compound invocation, environment scrubbed, output to a scratch file, counts by `grep -c '^--- PASS'` (top-level test lines; sub-tests are indented and not counted):

```text
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test ./internal/web -count=1 -v -timeout 560s
internal/web (whole)   ok  35.302s   top-level PASS 536, FAIL 0, SKIP 7
```

The sorted list of the 536 passing top-level names is kept in the gitignored scratch area of the session and is compared with the after list in the self-verification. Resource lease: `moai slot acquire --resource t1399-run --max-duration 60m` printed `slot t1399-run acquired by 0dcdf2d5-df5c-4da1-8870-24c2a5861303 until 2026-10-02T19:19:18Z`; it is released before the milestone's last commit.

#### The AC-019 tests, observed RED before the rename (commit `e590f0e1d`)

The five tests AC-019 selects by name were authored first, in `internal/web/factory_screen_m9_test.go` (four tests) and `internal/web/legacy_routes_test.go` (`TestLegacyKanbanRouteRedirects`, four subtests), and run on the pre-M9 tree. They use only identifiers that exist there, so the red commit builds (`go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and `GOOS=windows GOARCH=amd64 go vet ./internal/web` printed `BUILD_VET_OK`).

```text
$ go test ./internal/web -run '^(TestFactoryScreenOmitsChainBoard|TestFactoryScreenStillShowsFactoryLanes|TestLegacyKanbanRouteRedirects|TestTodoScreenStillLive|TestWebLiveKeyContract)$' -v -count=1      (exit 1)
--- FAIL: TestFactoryScreenOmitsChainBoard
--- FAIL: TestFactoryScreenStillShowsFactoryLanes
--- FAIL: TestWebLiveKeyContract
--- FAIL: TestTodoScreenStillLive
--- FAIL: TestLegacyKanbanRouteRedirects
    --- FAIL: TestLegacyKanbanRouteRedirects/GET_redirects_to_/factory_with_no_page_body
    --- FAIL: TestLegacyKanbanRouteRedirects/the_query_string_is_carried_to_the_target
    --- PASS: TestLegacyKanbanRouteRedirects/other_methods_are_refused
    --- FAIL: TestLegacyKanbanRouteRedirects/the_target_serves_the_screen
FAIL	github.com/modu-ai/moai-adk/internal/web
the reasons, from the first run of the same tests (the failure lines of the test files):
factory_screen_m9_test.go:43: screens.templ still carries the chain session board panel
factory_screen_m9_test.go:52: GET /factory status = 404, want 200
factory_screen_m9_test.go:121: the live key set [config goal kanban session spec verify] carries no "factory" area
factory_screen_m9_test.go:144: GET /todo carries no data-live="factory" area for the todo queue
factory_screen_m9_test.go:153: the todo queue directory resolves to event "kanban", want "factory"
legacy_routes_test.go:27: GET /kanban status = 200, want a 3xx redirect
legacy_routes_test.go:46: GET /kanban?profile=work Location = "", want /factory?profile=work
legacy_routes_test.go:66: GET /factory status = 404, want 200
```

The one subtest that passes on the old tree is the refused-method guard: the first run printed `legacy_routes_test.go:56: POST /kanban status = 403, want 405` because the cross-site check answers 403 before the method check when the request carries no `Sec-Fetch-Site`; the test now sets `Sec-Fetch-Site: same-origin` (the way `servePost` does), after which the old route also answers 405. It is a guard, not a flipped criterion, and is not claimed as RED. The test file of the redirect keeps the word in its strings (the retired route is its subject) and is exempt from the program's string and comment rules.

The identifier guard of `internal/cli` (M7's `TestNoRetiredWordIdentifiersOutsideWeb`) was observed red on a scratch clone of the pre-M9 commit with its `internal/web` exclusion dropped, before the guard was changed in the tree:

```text
$ go test ./internal/cli -run '^(TestRetiredWordIdentifierScanHasTeeth|TestNoRetiredWordIdentifiersOutsideWeb)$' -v -count=1      (clone of e590f0e1d, the web exclusion removed)
--- PASS: TestRetiredWordIdentifierScanHasTeeth (0.00s)
    retired_word_identifiers_m7_test.go:114: 95 Go identifier(s) outside internal/web still carry the retired mode word (run probe/rename/m7_rename.go):
        ../../internal/web/app.go:163 handleKanban
        ../../internal/web/factory_lane_identity_test.go:20 KanbanRecord
        ... (95 in all: the program's 93 plus the two on line 549 of the generated screens_templ.go, `func Kanban(vm ShellVM, k KanbanVM)`, which the regeneration rewrites)
--- FAIL: TestNoRetiredWordIdentifiersOutsideWeb (0.62s)
```

#### The rename program (committed, re-runnable)

Path: `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m9_rename.go` (`//go:build ignore`, standard library only; an identical copy sits in the gitignored `.moai/reports/t1399/rename/`; the committed file is the source of truth). Run command, from the tree root, on a tree whose tracked files are committed:

```text
go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m9_rename.go -root <git tree> [-dry-run] [-v]
```

What it does, in order (the header comment of the file is the full statement): (1) every identifier of a non-generated Go file of `internal/web` that carries the word is renamed by syntax tree (`KanbanVM` to `FactoryVM`, `buildKanban` to `buildFactory`, `handleKanban` to `handleFactory`, `loadKanbanRecords`, `KanbanRecord`, the test helpers and test names), except the two pinned test names `TestPreexistingKanbanArtifactsTolerated` and `TestLegacyKanbanRouteRedirects`; a new name already declared at package level refuses; (2) string literals, comments, the `.templ` sources, and `assets/app.js`, `assets/i18n.js`, `assets/console.css`: every standalone or camel-case token that is the word becomes `factory`/`Factory` (the route `/kanban`, `data-live="kanban"`, the SSE key, the `kanban.*` key family, the icon `case "kanban"`, the area id, `GET /kanban` in messages), a hyphenated or underscored compound is left, and in comments the Korean spelling becomes `팩토리`; a literal `"kanban"` after a `"state"` path argument (the frozen legacy state directory) is left, and the legacy-route file, its test, and the pre-existing-artifacts test keep the word in strings and comments; (3) the `nav.kanban` and `screen.kanban` values per locale become the loanwords the same file already uses for the factory (`Factory`, `팩토리`, `ファクトリー`, `工厂`) and the two Chinese strings of the SPEC board page that use the word for "board" (`board.title`, `board.subtitle`, English values Board and dashboard) become `面板` and `仪표板`; (4) the chain session board panel of `screens.templ`, the nine i18n keys only it used in each locale (`chain.title`, `chain.open`, `chain.stopped`, `chain.sessionNotStarted`, `kanban.viewA`, `kanban.card`, `kanban.model`, `kanban.noSession`, `kanban.noStart`), the view model (`ChainVM`, `RoleVM`, `ChainRoles`, `buildChain`, `chainRoleRecords`, `chainCardID`, the `Chain` field of `OverviewVM`, the `CardID`, `IdleRole`, and `Roles` fields of the screen model, the chain parameter of `buildAttention`, the chain-stopped attention row per decision Q24's smallest-footprint reading, and the builder statements that fed the chain), every test declaration that references them (one test, `TestBuildAttentionOrderAndCap`, is rewritten without the chain row because it also pins the order and the cap of the must-fix rows), then the declarations whose only readers went with them (`roleOf`, `readTelemetry`, `legacyLeaderRole`, `legacyLeaderLabel`, the test helper `writeTelemetry`, the `Role` field of `AttentionVM`) when nothing in the package reads them any more, files left without declarations, and the imports left unused; (5) `git mv` of a file whose name carries the word; (6) the redirect-only `legacy_routes.go` (created when absent) and its registration in `app.go`; (7) `go run github.com/a-h/templ/cmd/templ generate -path ./internal/web`, the command of the Makefile target `templ-generate`. Refusals: exit 2 on a tracked modification, exit 1 with nothing written on a collision, a source that does not parse or gofmt, or a structure it removes that is not in the expected shape.

Observed results, each a command run in this run:

| What | Command, tree | Observed |
|---|---|---|
| attribution of the plan's number | `-dry-run` on a scratch clone at `b6fea574c` (before this milestone's tests) | `identifiers: 91 (14 distinct)`: the plan's probe figure, exact; the real run's 93 is that plus the two `writeKanbanRecord` uses of the new `factory_screen_m9_test.go` |
| dry run on this tree | `-dry-run` at `b7851385b` | `identifiers: 93 (14 distinct); string literals: 62; comment mentions: 31; templ spellings: 23; script spellings: 66; i18n keys removed: 36; panel blocks removed: 1; view-model removals: 26; test declarations removed or rewritten: 15; imports dropped: 10; files rewritten: 29; files removed: 2; files renamed: 1; legacy route: created and registered; templ generate: skipped (dry run)` |
| rehearsal on a scratch clone | the real run on a clone of `b7851385b`, then `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, `go vet ./internal/web` host and windows, `golangci-lint run --timeout=5m ./internal/web/...` | the same counts; `HOST_BUILD_OK`, `WIN_BUILD_OK`, `HOST_VET_OK`, `WIN_VET_OK`, `0 issues.`; an earlier rehearsal of the program before it dropped orphans printed `5 issues` (`writeTelemetry`, `legacyLeaderRole`, `legacyLeaderLabel`, `roleOf`, `readTelemetry` unused), which is why the orphan rule exists |
| the real run, commit `529a294d5` | the run command on `b7851385b` | the dry-run counts without the suffix, `templ generate: exit 0`, `renamed internal/web/preexisting_kanban_artifacts_m1_test.go -> internal/web/preexisting_factory_artifacts_m1_test.go`, `removed internal/web/role_naming_m3_legacy_leader_test.go`, `removed internal/web/session_telemetry_cells_test.go`; then `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` printed `HOST_BUILD_OK` and `WIN_BUILD_OK`, `go vet ./internal/web ./internal/cli` and the windows vet of `./internal/web` passed |
| idempotence on the final tree | the run command at `bdfc8cd86`, then `git status --short \| wc -l` | `identifiers: 0 (0 distinct); string literals: 0; comment mentions: 0; templ spellings: 0; script spellings: 0; i18n keys removed: 0; panel blocks removed: 0; view-model removals: 0; test declarations removed or rewritten: 0; imports dropped: 0; files rewritten: 0; files removed: 0; files renamed: 0; legacy route: present; templ generate: exit 0` and `0` |
| refuses a dirty tree | the run command on a scratch clone of `bdfc8cd86` holding one tracked edit | `m9_rename: the tree has 1 tracked modification(s); commit or revert them first (first 1 shown):` / `M internal/web/events.go` / `exit status 2` |
| converts a planted late reference | scratch clone of `bdfc8cd86` with one untracked file `internal/web/planted_late_m9_test.go` that calls `writeKanbanRecord`, `KanbanRecord`, `loadKanbanRecords`, GETs `"/kanban"`, and asserts `data-live="kanban"` | before: `go vet ./internal/web` printed `vet: internal/web/planted_late_m9_test.go:12:2: undefined: writeKanbanRecord`; the run printed `literal ... "/kanban" -> "/factory"`, `literal ... data-live="kanban" -> data-live="factory"` and `identifiers: 3 (3 distinct); string literals: 2; ... files rewritten: 1`; after: the file reads `writeFactoryRecord`, `FactoryRecord`, `loadFactoryRecords`, `"/factory"`; `go vet ./internal/web` and `go build ./...` of the clone printed `CLONE_BUILD_VET_OK` and `go test ./internal/web -run '^TestPlantedLateReferenceM9$' -v -count=1` printed `--- PASS: TestPlantedLateReferenceM9` and `ok` |

Not covered by the program (hand edits, commit `bdfc8cd86`, and what a post-absorption re-run leaves to a manual read): (a) `TestFactoryNoteBannerCorrected` read the not-recorded marker from the blank cells of the removed board, so it fails once the board is gone (`factory_lane_section_test.go:189: the rendered marker carries no translated hover text`); it now renders a lane whose record names neither a card nor a SPEC and reads the marker there; (b) `TestLaneSectionAddsNoTransportSurface` asserted that `data-live="factory"` is absent, which after the rename is the page's own area (`factory_lane_section_test.go:260: the lane section declared a new live area`); it now asserts only that no `data-live="lane"` area is added, and the live-area count equality beside it stays; (c) prose the rules cannot judge: the comment of the watch map key in `events.go` (it said the key stays), the screen crumb `"chain + pipeline"` in `screens.go` (now `"lanes + pipeline"`), the `factory_lanes.go` and `viewmodel_ops.go` comments that explained lanes against the chain, the Korean header block above `templ Factory` in `screens.templ` (the templ output was regenerated after it), the `StateIdle` comment, a test comment, the first line of the pre-existing-artifacts test that names its own file, and `todo_queue_read.go:27`, which cited another SPEC's id that carries the word (reworded to name the `moai todo` command); (d) the identifier guard of `internal/cli` (its web exclusion dropped, renamed `TestNoRetiredWordIdentifiers`: the old name would say "outside web" about a scan that now covers the web package). The recipe a post-absorption run follows: merge develop and commit the merge; run the program and commit its output; run it again (all zeros, `git status --short` empty); `go test ./internal/cli -run '^(TestRetiredWordIdentifierScanHasTeeth|TestNoRetiredWordIdentifiers)$'`; `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`; `go vet ./internal/web`; `golangci-lint run --timeout=5m ./internal/web/...`; the five AC-019 tests; the whole `internal/web` suite; read the program's last line (what still carries the word).

#### Claim

Milestone M9 of card t1399 is implemented on branch `WT-launcher-entry-flags` at `bdfc8cd86`: `internal/web` carries no spelling of the retired word except `legacy_routes.go` (non-test Go, `.templ`, and `.js` sources, and the css), the chain session board and its view model are removed (the Overview chain-stopped attention row with it, decision Q24), the factory screen is served at `/factory` with `GET /kanban` kept as a redirect, and the live area key, `data-live` markers, i18n key family, icon id, shell area id, view-model, handler, and test names carry factory names. Commits (card t1399): `b6fea574c` (sizing and baseline, docs), `e590f0e1d` (the AC-019 tests, red), `b7851385b` (the program), `529a294d5` (its output), `bdfc8cd86` (hand fixups and the guard), then this record. `git diff --stat 8d50c79d6 HEAD | tail -3` at `bdfc8cd86`: `internal/web/widgets.templ | 4 +-`, `internal/web/widgets_templ.go | 8 +-`, `40 files changed, 2422 insertions(+), 1786 deletions(-)`.

#### Evidence

AC matrix, each row a command run on tree `bdfc8cd86`:

| Criterion | Command | Observed |
|---|---|---|
| AC-019 panel | `grep -c 'Chain session board' internal/web/screens.templ; echo "exit=$?"` | `0` and `exit=1` (RED-W1: 1 before) |
| AC-019 tests | `go test ./internal/web -run '^(TestFactoryScreenOmitsChainBoard\|TestFactoryScreenStillShowsFactoryLanes\|TestLegacyKanbanRouteRedirects\|TestTodoScreenStillLive\|TestWebLiveKeyContract)$' -v -count=1` | five top-level `--- PASS` (and the four redirect subtests), `ok … internal/web 2.824s` (swept 5; `TestWebLiveKeyContract` asserts `events.go` and the `app.js` EVENTS list agree on the key set); the same command redirected to a file and not piped printed `go_test_exit=0`, five `--- PASS`, `ok … internal/web 3.284s` |
| AC-019 route registration | `grep -c '"/kanban"' internal/web/app.go; echo "exit=$?"`, and `grep -rn '/kanban' internal/web` over the non-test `.go`, `.templ`, `.js` | `0` and `exit=1` (RED-W2: 1 before); the only non-test hit is `internal/web/legacy_routes.go:9`, the pattern `GET /kanban`; `TestLegacyKanbanRouteRedirects` pins the behavior: a 3xx with `Location: /factory`, a body without `<html`, `<body`, `data-live`, under 200 bytes |
| AC-019 build and generated files | `go build ./internal/web`; the program's step 7 on the final tree, then `git status --short \| wc -l` | exit 0; `templ generate: exit 0` and `0` (the Makefile target `templ-generate` is this same `go run github.com/a-h/templ/cmd/templ generate -path ./internal/web`; the target itself was not run) |
| AC-018 web sources | the AC-018 word grep restricted to `internal/web` (non-test `.go`, `.templ`, `.js`) | `internal/web/legacy_routes.go` only, exit 0 (16 files before) |
| AC-018 whole tree | the AC-018 word grep over `internal cmd` | 68 files, 1 of them in `internal/web` (83 and 16 before) |
| AC-018 find | `find internal cmd -iname '*kanban*'` | 4 names, all template paths owned by M10: `internal/template/templates/.claude/skills/moai-kanban-foreman` and the three `kanban-dispatch*.md` rules (5 before: `internal/web/preexisting_kanban_artifacts_m1_test.go` is renamed) |
| AC-018 importers | `grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'; echo "grep_exit=$?"` | no file, `grep_exit=1` |
| AC-018 builds | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `go vet ./...` redirected to a file | `BUILD_HOST_OK`; `BUILD_WINDOWS_OK`; `vet_exit=0` and a 0-byte file (the lease was held) |
| windows vet of the touched packages | `GOOS=windows GOARCH=amd64 go vet ./internal/web ./internal/cli ./internal/hook ./internal/config ./internal/factory/... ./internal/statusline ./internal/codexwiring ./internal/discovery` | `win_vet_exit=0`, 0 bytes (the pre-existing `internal/cli/worktree/sweep_test.go:1687` failure is outside this list) |
| AC-018 per-file bound | `grep -c -i kanban internal/cli/launcher_retired_entries.go internal/cli/update_archive.go internal/factory/state_dir.go internal/web/legacy_routes.go` | `4`, `0`, `1`, `1`; the M9 value for `legacy_routes.go` is 1, recorded for the re-runs at M11 and at sync |
| AC-015 net | the 9-name cli alternation of AC-015, `-v -count=1`; `TestFactoryNetSessionRecord` and `TestFactoryNetSessionStartNotices` in `./internal/hook`; the two discovery tests in `./internal/discovery` | eight `--- PASS` (`TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestFactoryNetLeaderLaunch`, `TestFactoryNetLaneLaunch`, `TestFactoryNetBlockCap`, `TestFactoryEntryMatrix`, `TestPrepareFactorySettingsWritesTransientFile`, `TestCCFactoryEntryRecordsFailOpenRunMetadata`), `ok … internal/cli 19.716s` at the final code state (also after the program's output); two and two `--- PASS`: 8 + 2 + 2 |
| AC-016 | `TestFactoryMarkerValuesFrozen` (`./internal/config`), `TestMCPServerEnvVarsKeepFactoryMarkers` (`./internal/codexwiring`), `TestLegacyStateDirStillRead` (`./internal/factory`), and the two discovery tests | each `--- PASS`, `ok` |
| AC-017 | `go test ./internal/web ./internal/hook ./internal/statusline -run '^TestPreexistingKanbanArtifactsTolerated$' -v -count=1` | three `--- PASS`, three `ok` (swept 1 per package; the web test lives in the renamed file) |
| env-literal guard | `go test ./internal/hook -run '^TestNoNewEnvLiteralsInDiff$' -v -count=1` | `env-literal sweep: 1732 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2, distinct literals=0`, `--- PASS` |
| identifier guard | `go test ./internal/cli -run '^(TestRetiredWordIdentifierScanHasTeeth\|TestNoRetiredWordIdentifiers)$' -v -count=1` | both `--- PASS` (the first is the positive control; the red of the web-covering form is above, 95 offenders) |
| lint | `golangci-lint run --timeout=5m ./internal/web/... ./internal/cli` (v2.1.6, the CI version) | `0 issues.` (the base printed `0 issues.` for `./internal/web/...`) |
| gofmt | `gofmt -l internal/web internal/cli/retired_word_identifiers_m7_test.go .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m9_rename.go` | prints `internal/web/codex_panel_test.go` only: a pre-existing finding, the file is unchanged since `445317711` (`git diff --stat 8d50c79d6 HEAD -- internal/web/codex_panel_test.go` printed nothing) |

Whole `internal/web`, before (tree `8d50c79d6`) and after (tree `bdfc8cd86`), top-level `--- PASS` lines by `grep -c`:

```text
before   ok  35.302s   PASS 536  FAIL 0  SKIP 7
after    ok  43.355s   PASS 529  FAIL 0  SKIP 7
536 - 12 removed + 5 added = 529; the sorted name lists differ in 16 names before and 9 after: the 12 removed, four renamed (TestKanbanLaneStates, TestKanbanNoteBannerCorrected, TestKanbanPipelineColumns, TestTodoSectionCarriesExistingKanbanMarker became TestFactory... and ...ExistingFactoryMarker), and the 5 added (the AC-019 tests; TestLegacyKanbanRouteRedirects among them)
```

The 12 removed top-level tests, the files they were in, and why (each exercised only the removed chain view model, `chainRoleRecords`, `buildChain`, or `chainCardID`, or built a `RoleVM`/`ChainVM`): `factory_lane_identity_test.go` `TestChainIsAbsentWhenOnlyFactoryLanesHaveRecords`; `role_naming_m3_legacy_leader_test.go` (the file is gone) `TestLeaderRecordRendersPresentLeaderSlot`, `TestLegacyLeadRecordRendersRelaunchLabel`, `TestLegacyLeadRecordKeepsLaneRecordsOutOfTheChain`; `screen_states_test.go` `TestKanbanChainRoleStates`, `TestKanbanRoleWithNoTelemetryRecord` and the helper `tg2PopulatedRole`; `session_telemetry_cells_test.go` (the file is gone) `TestChainCellsCarryTelemetryValues`, `TestChainCellsDoNotBorrowAnotherSessionsValues`, `TestChainCellsTolerateSchemaV1Record`, `TestChainCellsUnreadableTelemetryStaysBlank`, `TestChainCellsClampContextPercentage` and the helpers `roleByName` and `writeTelemetry`; `viewmodel_ops_edges_test.go` `TestChainCardIDAndRoleFilter`. These are the plan's five files exactly. `TestBuildAttentionOrderAndCap` stays, rewritten: the chain row and the chain parameter are gone, and the order, the cap, the non-must skip, and the quiet case keep their assertions. The lane, pipeline, todo, specs, and monitor tests of the same files are unchanged except for the names.

#### Remaining after M9: what still carries the word, and the owner (baseline for M10 and M11)

Go sources and web assets: none in `internal/web` except `internal/web/legacy_routes.go` (count 1, the `GET /kanban` redirect pattern; allowed by AC-018 and held to a redirect by `TestLegacyKanbanRouteRedirects`). `assets/app.js`, `assets/i18n.js`, `assets/console.css`, and the four `.templ` sources carry none. Web test files: `legacy_routes_test.go` (11 occurrences in strings and comments) and `preexisting_factory_artifacts_m1_test.go` (10; the retired artifacts it plants are its subject), plus the two pinned test names `TestLegacyKanbanRouteRedirects` and `TestPreexistingKanbanArtifactsTolerated` (AC-017/AC-019 select them by name). Outside `internal/web` nothing changed in M9: the AC-018 word grep lists 67 files elsewhere (83 − 16 at M8), with the owners recorded in the M8 section above (the frozen marker values, the allowed `launcher_retired_entries.go` and `state_dir.go`, the 15 unowned string lines, and comments; no milestone in plan.md M7-M10 rewords the unowned ones). The four template paths (`moai-kanban-foreman`, the three `kanban-dispatch*.md`) and the one `update_archive.go` line belong to M10; docs, READMEs, and rule and skill text belong to M10 and M11.

#### Baseline-attribution

- Trees and builds: the sizing and the baseline were measured on `8d50c79d6` (this worktree, HEAD at the start; the baseline suite and the lint were run before the first commit); the RED on the tree of `b6fea574c` plus the uncommitted test files; the program on `b7851385b`; the AC rows, the builds, the vets, the net, AC-016/017, the guard, and the idempotence run on `bdfc8cd86` (or the working tree identical to it, with `git status --short` empty); the whole after-suite on the tree of the same content. HEAD, branch (`WT-launcher-entry-flags`), and `git status --short` were re-read before each commit.
- Judging build: every measurement is by the Go toolchain and shell tools run from this tree; `moai slot acquire` and `moai slot release` were the only invocations of the installed `moai` binary (the release printed `slot t1399-run released (was 0dcdf2d5-df5c-4da1-8870-24c2a5861303)`).

#### Gaps

1. No browser or end-to-end run: the console's behavior in a browser (the rendered `/factory` page, the `data-live` refresh over SSE after a queue write, the language switch for the renamed keys) was not observed; the Go suite includes the app.js and rendering tests (7 skips, as at the base, not examined) and the five new tests.
2. `internal/cli` whole suite not run: the net, the identifier guard, and `go vet ./internal/cli` were; the other `internal/web` neighbours (`internal/hook`, `internal/statusline`, `internal/factory`) ran only the named tests above, not their whole suites (M9 changed no file there). `internal/template` was not run.
3. `make templ-generate` itself was not run; the identical `go run github.com/a-h/templ/cmd/templ generate -path ./internal/web` was (by the program, and by hand after the header comment edit: `updates=1`, and the regenerated `screens_templ.go` is in the same commit as the comment). No test was run on Windows (builds and vets only). `golangci-lint` ran over `./internal/web/...` and `./internal/cli`, not the whole tree.
4. Non-English strings changed: `nav.factory` and `screen.factory` take `Factory`, `팩토리`, `ファクトリー`, `工厂`, the loanwords the same files already use for the factory (`팩토리 레인`, `ファクトリーレーン`, `工厂通道`), so no sentence was authored; the two Chinese board strings now read `面板` and `仪表板`. The native-idiom hazard list (`native-idiom-and-register-detail.md`) and the humanize skill were not loaded; the check was by hand: every changed value is a one-word label or a loanword already in use in that file, with no English syntax or figurative stock carried over.
5. Worktree-guard refusals, three, each re-issued as plain commands measuring the same thing: (a) one compound of `grep`, `which templ`, `go version`, and `ls $(go env GOMODCACHE)/…` refused as naming git in a form too complex to verify — split into single commands; (b) one compound of `sh <helper> add` and `commit` with `$S` path variables plus a `go run` refused ("what it reads or is handed as shell text cannot be shown not to run git") — re-issued as separate plain commands with literal paths; (c) `git clone … && sh <helper> checkout …` refused as too complex — split in two.
6. Scratch clones (`git clone --local --no-hardlinks` into the scratchpad, never a worktree) and a scratchpad helper script that runs one git command inside a scratch clone (to stage, commit, and check out there) were used for the rehearsals, the idempotence run, the dirty-tree refusal, the planted-reference proof, the commit `b6fea574c` attribution, and the guard's RED; no `git worktree add` was run and the tree's own git was run directly.
7. The 95-offender RED of the guard was observed on a scratch clone with a hand edit of the guard; the edit in the tree and the program's output are later commits, so the order is a session record for that RED and a commit-graph fact for the five AC-019 tests (`e590f0e1d` precedes `529a294d5`) and for the program (`b7851385b` precedes its output).
8. The program deletes any web test declaration that references a removed name. This run printed the 15 it touched (listed above); a late test that mixes removed and kept assertions would be deleted whole, which `-v` shows and the summary counts.

#### Residual-risk

- The note banner of the factory screen (`factory.note`, four locales) still says that a blank model, effort, or context cell means the session has no telemetry record yet. Those cells belonged to the removed board; the lane rows draw none of them, so half of the sentence now describes nothing. It was kept because `TestFactoryNoteBannerCorrected` (AC-WC15-052, a different SPEC's guard that this SPEC's file list does not name) asserts that the banner survives and carries its key; the probe's panel cut would have removed it. Rewording four locales is a decision for the leader.
- `telemetryCells` and the `gauge` templ have no production reader any more; their own tests (`TestTelemetryCellsAbsence`, the gauge states in `shell_chrome_states_test.go`) keep them live for the linter. They are not in the plan's list and were not deleted (AGENTS.md section 5). The css rules `.chainbar*` (already unused before M9) and the `.roles--chain` selectors are left.
- `GET /kanban` answers `301 Moved Permanently` (the probe's choice): a browser caches it, which is the intent for a retired path. A query string is carried to the target, which AC-019's text does not state; a query-less request gets exactly `Location: /factory`.
- The program rewrites by rule; a late reference of a shape the rules do not know (a hyphenated or underscored compound, prose that the word replacement makes false) is reported only on the last informational line and by `-v`, and a wrong rewrite shows at the build and the suites.
- Free text outside Go sources that spells the old route or key names (docs, the dispatch rule's mention of the web console, the generated codemaps) was not touched; M10 and M11 own it.

#### Findings (statements the tree contradicts, no SPEC file edited)

1. plan.md M9 and the probe cut the panel through the note banner (`"kanban.note"`), but `TestKanbanNoteBannerCorrected` (`factory_lane_section_test.go`, not among the plan's five test files) requires the banner to survive. The program cuts up to the banner and keeps it (residual risk above).
2. AC-018's word pattern also matches the Chinese word for "board" (`看板`), so two strings of the SPEC board page (`board.title` and `board.subtitle` in the Chinese block of `i18n.js`, English `Board` and `Read-only SPEC lifecycle dashboard …`) carry it without being about the retired mode; they were reworded (`面板`, `仪表板`) to meet the grep.
3. The chain removal leaves five declarations that only it read (`roleOf`, `readTelemetry`, `legacyLeaderRole`, `legacyLeaderLabel`, the test helper `writeTelemetry`) and the `Role` field of `AttentionVM`; plan.md M9 does not list them, the compile proof passes without removing them, and `golangci-lint` v2.1.6 reports five `unused` issues if they stay. They are removed, by the program, when nothing reads them.
4. AC-018 cannot read "exactly four files" over the web package while a comment cites an id that carries the word: `todo_queue_read.go:27` cited `SPEC-KANBAN-TODO-CLI-001` (an id of another SPEC); it was reworded. The same holds for the other packages' citations (M8 finding 3, unchanged).
5. plan.md and design.md say the redirect handler is "in `legacy_routes.go`" and that `/kanban` is "kept as a GET redirect"; neither says the HTTP status. The program uses `301`.
6. The M8 verification recipe (section above) names `TestNoRetiredWordIdentifiersOutsideWeb`; M9 renamed the test `TestNoRetiredWordIdentifiers`, so a command that still names the old test sweeps zero tests.

### Develop absorption, residue sweep, M10, M11 (recorded by the lane orchestrator after M9)

Evidence files (gitignored, local): `.moai/reports/t1399/absorb-remeasure.md` (commands and verbatim outputs for the absorption), `merge-absorb-develop.md` (the conflict-resolution worker's table; its build/test claims are superseded by the lane's re-measurement), `residue-sweep.md`, `m10-rename.md`, `m10-factory-skill.md`, `m10-always-loaded.md`, `m11-docs-site.md`, `m11-readme-instr.md`. Where this section contradicts an earlier line of this file (the M3 section says codex `--factory-run` stays refused; M2 to M9 evidence names `a6d3e6fd4` as the base), this section is the later record and the earlier lines are left as written history.

#### Claim

The card branch absorbed develop `1e2151a380a5dd0d76efd8f1740a21f32c682d2f` (merge `be179549a`), the three rename programs were re-run in order on the absorbed tree (second runs all zeros), the residue of the retired mode's name in non-test sources was swept, and M10 (rules, skill, catalog, constitution slots, the factory skill rewrite and its behavior test) and M11 (docs-site, READMEs, instruction files, maintainer docs) are implemented. Commits after `903fb8c92`: `be179549a` merge; `6085e3233` m7, `85ca2b253` m8, `27573396c` m9 (plus 6 hand-edited i18n lines); `d64b1f461` qualifier fixes; `52b984fc2` residue sweep and the `moai factory relaunch --lane` note; `7b3445a98` M10 commit 1 (update fixture, committed red); `859477bc7` M10 commit 2 (rename); `7a28a85b9` M10 factory skill and test; `129127ad1`, `cf4e8d6d2`, `3ade55fbb`, `b7ad5adc2`, `5bcc77857` docs-site (page units, deletions and redirects, remaining pages); `d12eb9027` READMEs; `d7158b34e` AGENTS.md and AGENTS.local.md; `cd692c429` maintainer docs; `7758c9c51` plan corrections (v0.9.0, 46 rows in spec.md section A.4). The last commit that changes Go source is `7a28a85b9`; every later commit is documentation or SPEC text.

#### Evidence

Every row was run by the lane in this run (the worker reports were not taken as measurements).

| What | Command, tree | Observed |
|---|---|---|
| absorption base | `git fetch origin develop`, `git rev-list --count --left-right origin/develop...develop` | `0 0`; develop `1e2151a38`; merge-base with the card base `a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2`; the simulated merge reported 15 conflicted files (35 hunks), 4 modify/delete, 2 added-in-renamed-dir |
| rename programs, second runs | m7, m8, m9 on `d64b1f461`-era tree | `identifiers renamed: 0`; `import lines: 0; qualifiers: 0`; m9 all zeros, `templ generate: exit 0`; `git status --short \| wc -l` printed `0`; `grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'` exit 1 |
| builds | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | exit 0, exit 0 (after the five qualifier fixes of `d64b1f461`; before them: `undefined: kanban` at `stale_run_gate.go:236`, `codex_launcher.go:985`, `glm.go:269`) |
| vet | `go vet ./...` (lease `t1399-run` held); windows vet of the touched packages; again after the sweep and M10 | exit 0 each, empty output |
| factory net and guards | the 8 cli names of AC-015 (`-count=1`, env scrubbed); hook `TestFactoryNetSessionRecord`, `TestFactoryNetSessionStartNotices`, `TestPreexistingKanbanArtifactsTolerated`, `TestNoNewEnvLiteralsInDiff`; AC-016/017 over six packages; `TestRetiredWordIdentifierScanHasTeeth`, `TestNoRetiredWordIdentifiers` | 8 PASS (`ok internal/cli 64.716s`); 4 PASS with `env-literal sweep: 1812 added lines swept ... base=1e2151a380a5dd0d76efd8f1740a21f32c682d2f, distinct literals=0`; 7 PASS; 2 PASS |
| whole suites | `go test ./internal/hook ./internal/factory ./internal/web ./internal/factorymsg ./internal/config -count=1` on `d64b1f461` | all `ok` (hook 1086.624s, factory 461.643s, web 128.234s, factorymsg 226.578s, config 19.187s) |
| whole suites after the sweep | `go test ./internal/factory ./internal/factorymsg ./internal/statusline ./internal/homestate ./internal/spec -count=1` | factorymsg, statusline, homestate, spec `ok`; `internal/factory` FAIL on `TestForemanQueueWatch_FiresOnMutation`, `_FiresWithStaleJSONPresent`, `_SeesWALDeferredCommit` ("no change event within 16s"); the same names alone: `-run ^TestForemanQueueWatch -count=1` exit 0, 7 PASS, `ok 124.794s` — load-dependent timing tests, not cited as evidence |
| residue sweep | AC-018 word grep (exact acceptance command) | 70 files / 134 lines before, 24 files / 29 lines at the final tree: 22 `SPEC-KANBAN-*` citation lines (ids of other SPECs; an id cannot be reworded) plus `legacy_routes.go:9`, `launcher_retired_entries.go:3,4,21,29`, `update_archive.go:70`, `state_dir.go:23`; `find internal cmd -iname '*kanban*'` empty; AC-024 grep exit 1 (empty) |
| relaunch note | `TestFactoryRelaunchLaneIsIgnoredWithNote` | worker-observed red (`stderr carries the --lane note 0 times, want exactly once`) then green; test and code share commit `52b984fc2`, so the order is a session record, not a commit-graph fact |
| M10 fixture | `go test ./internal/cli -run '^TestUpdateRemovesRetiredRuleFilesWithBackup$' -v -count=1` | on `7b3445a98`'s tree: unmodified and user-modified subtests FAIL ("has no copy in the pre-clean backup"), absent PASS; positive control on a scratch clone whose embedded template lacks the three paths: 4 PASS; after `859477bc7`: 4 PASS. The red commit `7b3445a98` precedes the rename commit in the commit graph |
| M10 guards | template guard and catalog tests (8 names); `env MOAI_GR_BASE=1e2151a380a5dd0d76efd8f1740a21f32c682d2f go test ./internal/template -run '^(TestContractModeConstitutionDriftNotIncreased\|TestContractModeAlwaysLoadedBudget)$'`; `moai constitution validate` (tree-built binary) | 8 PASS; 2 PASS, 0 SKIP; `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)` before the first edit, after each of the four slot edits (worker), and on the final tree (lane) |
| M10 greps | catalog and archive counts; rule paths; zone registry | `moai-kanban-foreman` in catalog 0, `moai-factory-foreman` 2, in `update_archive.go` 1; six new rule paths present, old two absent; `git diff --quiet 1e2151a38 HEAD -- .claude/rules/moai/core/zone-registry.md` exit 0; `[HARD]` counts 38/8/8 unchanged across the three rules |
| always-loaded surface | `m10-always-loaded.md` method | 15 files / 191433 bytes before, 15 files / 191028 bytes after (-405) |
| AC-021 | `go test ./internal/cli -run '^TestFactorySkillAssertionsMatchBehavior$' -v -count=1` | PASS (one parent test, 16.38s); worker-observed red on the old skill text (73 lines, `m10-red.txt`), nine mutant texts each reported; greps: `factory_chain` 0 and 0, `chain head` 0 and 0, `.moai/state/factory` 0, `factory contract\|factory chain` over the dependents exit 1 |
| M11 docs-site | Hugo build to scratch (`hugo --source docs-site --minify --gc`) | exit 0, `grep -c -e WARN -e ERROR` 0, sitemap present, pages 187/185/185/185; vercel.json: factory-mode sources 0, removed-page sources 6, destinations naming a removed page 0, JSON parses; four `origin-trail-chain` pages present, twelve removed pages absent |
| M11 docs gates | AC-023 word grep, removed-forms grep, `find -iname '*kanban*'` outside internal/cmd | all empty at `cd692c429`; README H2 counts 12/12/12/12 |
| AC-025 one commit per page | `git log --name-only 1e2151a38..HEAD -- <the four locale files>` per page | origin-trail-chain `129127ad1` (4 paths), factory-mode `cf4e8d6d2` (4 paths), launchers `3ade55fbb` (4 paths): one commit each |
| SPEC lint | tree-built `moai spec lint --strict .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001` | `No findings — all SPEC documents are valid`, exit 0 |

#### Decisions taken during the absorption (read by the lane in the diff or taken from the worker's table)

1. codex `-l --factory-run <id>` is ACCEPTED (develop card t1444 ②); without a lane entry it is refused as `--factory-run requires -l/--lane`; every `-f` shape stays refused. This supersedes the M3 section's statement that it stays refused (lines in the M3 findings and residual-risk).
2. `moai factory relaunch` (develop card t1345) launches `<provider> -l [--factory-run <id>]`; a supplied `--lane` cannot reach the launcher, so the verb prints one stderr line and joins the next free lane (leader ruling, 2026-10-03).
3. Develop's re-added free-slot lines in the leader notice were dropped (REQ-009); the `gateSummary` line (SPEC-AUTONOMY-BATCH-GATE-001) was kept.
4. A leader-shaped operator `--name` is claimed like the default (develop card t1444 ④), ported into `appendLeaderName` of `factory_launch_helpers.go` because `kanban.go` stays deleted.
5. Develop tests that taught removed forms were re-pinned to `-l`; two were narrowed because they exercised removed behavior (`session_start_leader_gate_notice_test.go` Kanban kinds, `codex_review_ownership_test.go` matrix rows).

#### Baseline-attribution

The judging build for every Go measurement is the Go toolchain run from this tree; `moai constitution validate` and `moai spec lint` used binaries built from this tree into the scratchpad (`go build -o <scratchpad>/moai-c2 ./cmd/moai` at `859477bc7`, `moai-spec2` at `cd692c429`). The env-literal guard's base moved with the absorption (`base=1e2151a380a5dd0d76efd8f1740a21f32c682d2f`). Hugo, grep, find, and the Hugo/page-count figures were taken on the working tree equal to the named commit.

#### Gaps

1. The whole `internal/cli` suite was never run (about 5,100 tests, does not finish in the foreground; the repository rule leaves the full-package verdict to CI). Run for cli: the factory net, the identifier guard, the relaunch selection (15 PASS), doctor goldens, the M10 and AC-021 tests, and `go vet` of the package on both OSes.
2. `internal/factory` has no clean whole-package run after the sweep (three timing tests failed under a five-package parallel run and pass alone); the whole package passed once on `d64b1f461`.
3. `TestStaleRunNoticeFactoryLegacyLabel` (`internal/hook`) fails 6 of 6 on a scratch clone of develop `1e2151a38` and 4 of 6 on this branch (`factory messaging degraded: context deadline exceeded`); it is not introduced here and is not cited. Listed in acceptance.md section H.
4. `golangci-lint` was not run on the tree after the absorption; no test was run on Windows (builds and vets only); the docs were not rendered in a browser; the ko/ja/zh prose has no native-speaker review beyond the humanize pass; the English text was not humanize-passed.
5. The deleted Kanban notice builders were the only SessionStart stale-run carrier for a non-factory session that carries only a legacy leader label; those sessions now get no notice (REQ-013 and AC-014 do not cover it). Needs an operator notice.
6. `assets/images/moai-web-overview.png` (embedded by all four READMEs) is a stale screenshot that still shows a retired rail entry; it needs a maintainer re-capture. The README headings say v3.2 while the badge and `hugo.toml` say v3.1.3; the release sync reconciles them.
7. Unresolved debt for an operator call (not removed, not in the plan's lists): `ReadCardStatus`, `FactoryFreeSlots`, `rejectFactoryOnCG`, `config.DefaultFactoryLanes`, the unreachable interactive Codex lane branch.
8. Editing the plan artifacts (`7758c9c51`) changed the plan-artifact hash, so the cached plan-audit PASS (0.88 at `7e1ae0d80`) no longer matches; the narrowed AC-018 and the other re-scoped criteria are judged at sync-audit.

#### Residual-risk

- A late develop merge that adds a `kanban.` qualifier or an import of the old package path merges textually clean and breaks the build; the integration window must re-measure `grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'` and `go build ./...` on the merged tree (plan.md M7 coordination).
- The `moai factory relaunch --lane` note and the codex `--factory-run` acceptance are behavior decisions taken on develop's code; a reviewer may prefer dropping `--lane` from the verb grammar (about 12 develop assertions).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-03
audited_head: 7758c9c51 (Go source last changed at 7a28a85b9; later commits are documentation and SPEC text)
milestones: M1 to M11 implemented and re-measured by the lane; M10 in three commits (fixture red, rename, factory skill), M11 in eight commits
open items for the audit: Gaps 1 to 8 of the section above; the AC-018 and other criteria narrowed in spec.md section A.4

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Plan-phase Premise Verification

Every row is a measurement taken in this plan run against tree `a6d3e6fd4` (branch `WT-launcher-entry-flags`; `git rev-parse --short HEAD` re-read before the final revision: `a6d3e6fd4`; `git status --short` showed only this SPEC directory untracked). Binary rows used a binary built from this tree (`go build ./cmd/moai`, exit 0; scratchpad path `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/0dcdf2d5-df5c-4da1-8870-24c2a5861303/scratchpad/moai-t1399`, built from `a6d3e6fd4`) or `go run ./cmd/moai`; the judging build and the tree measured are the same HEAD. Rows PV-1..PV-37 were taken in earlier revisions of this plan run against the same tree; the tree has not moved since. Rows marked "v0.1.0" measured the superseded verb-less design and stay true.

| ID | Command | Observed | Exit |
|----|---------|----------|------|
| PV-1 | `moai` (no arguments; tree-built binary) | banner + help text printed | 0 |
| PV-2 (v0.1.0) | `go run ./cmd/moai -f` | stdout empty; stderr `Unknown shorthand flag: 'f' in -f.` | 1 |
| PV-3 (v0.1.0) | `go run ./cmd/moai -l` | stdout empty; stderr `Unknown shorthand flag: 'l' in -l.` | 1 |
| PV-4 | `go test ./internal/cli -run '^(TestParseLauncherEntryMarksAutoAssignedNumbers\|TestCCFactoryLaneJoinsDiscoveredLeader\|TestGLMFactoryLaneJoinsDiscoveredLeader\|TestCodexFactoryEntryParsingUsesLaneOnly\|TestCodexFactoryLegacyEntryIsRefused\|TestCGRetiredEntryAndModeHaveZeroEffects\|TestCGRetirementCompleteEntryShapesAndCounters)$' -v -count=1` | seven PASS results (swept count 7), `ok github.com/modu-ai/moai-adk/internal/cli 12.521s` | 0 |
| PV-5 | `moai gpt`; `moai cg` (tree-built binary) | `Unknown command "gpt" for "moai"`, exit 1; `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`, exit 1 | 1 / 1 |
| PV-6 (v0.1.0) | `go run ./cmd/moai -f lane`; `go run ./cmd/moai -l lane-2` | stdout empty; unknown-shorthand diagnostic (`f` / `l`) | 1 / 1 |
| PV-7 (v0.1.0) | `grep -c -e '--lane' internal/cli/root.go`; `grep -c -e '--factory' internal/cli/root.go` | `0`; `0` | 1 / 1 |
| PV-8 | `grep -rl -e '--lane' internal/cli` | only todo/gtd command files and their tests (`--lane <label>` is a subcommand-local flag there) | 0 |
| PV-9 | bounded search for a default-backend setting (`grep -rIn 'default_backend\|DefaultBackend\|launch\.yaml' --include='*.go' internal`, tests filtered with a pipe) | only the last-used-profile ledger hits; no backend key. Moot: the verb carries the backend. | 0 |
| PV-10 | local-versus-template markdown pairs compared with `cmp -s` | `kanban-dispatch-mechanics.md` and `kanban-dispatch-detail.md` identical; `kanban-dispatch.md`, `orchestration-mode-selection.md`, `cross-session-messaging-detail.md`, `manager-lead.md` differ | n/a |
| PV-11 | `go build ./cmd/moai` | exit 0 (build log empty) | 0 |
| PV-12 | `go test ./internal/hook -run '^(TestUnbindNoticeRebindLinePresence\|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide\|TestFactoryGuideNamesWorkerJoinInEveryLocale)$' -v -count=1` | three PASS results (swept count 3), `ok github.com/modu-ai/moai-adk/internal/hook 2.303s` | 0 |
| PV-13 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` (tree-built binary) | see PV-46 for the v0.4.0 run | 0 |
| PV-14 | mutation probe on a scratchpad COPY of the SPEC directory: rewrite every `REQ-005` reference in acceptance.md to `REQ-0XX`, then lint | still `No findings` — the lint does not move when a requirement loses all coverage (measured on v0.1.0, v0.2.0, v0.3.0; re-measured in PV-46) | 0 |
| PV-15 | by hand: `grep -c "REQ-<n>" acceptance.md` per requirement | see PV-47 for the v0.4.0 count | 0 |
| PV-16 | `go run ./cmd/moai cc -f 3`; `moai glm -f 3`; `moai codex -f 3`; `moai codex -f lane-2`; `moai codex -f` | `-f <N>` is ALREADY refused on all three verbs: exit 1, stdout empty, one line, nothing launched. cc/glm line: `-F/--Factory takes no argument (the factory leader), the role token -f lane, which joins this session to a running factory as the next free lane, or a lane label (e.g. -f lane-2) that launches exactly that one lane, got "3".` codex line (all three shapes): `FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader` | 1 each |
| PV-17 | `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata\|TestGLM_FactoryLeadRunIsJoinableByLane)$' -v -count=1` | two PASS results, `ok ... 2.134s`; the first launches `-f` (leader) only; the second launches `-f` then `-f lane-3`, so it is re-pinned | 0 |
| PV-18 | `grep -rlE` of removed-lane-form patterns over AGENTS.md, READMEs, docs-site/content, .claude/agents, .claude/rules, internal/template/templates | 29 paths (the v0.2.0 inventory; superseded by PV-45 for the Tier L scope) | 0 |
| PV-19 | `go run ./cmd/moai cc -l`; `moai codex -l`; `moai cc -l lane-2`; `moai cc -f -l` | stdout empty, exit 1: `--Leader requires a leader label.` (cc, codex, and `cc -f -l`); `--Leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target.` (`cc -l lane-2`) | 1 each |
| PV-20 | `grep -c -e '-f lane'` over `cc.go`, `glm.go`, `codex_launcher.go`; over `session_start_factory.go`, `session_start_factory_i18n.go`, `factory.go`, `factory_card.go`; and `grep -c -e 'cc -f lane' internal/hook/session_stale_run.go` | cc 9, glm 8, codex_launcher 10; session_start_factory 3, session_start_factory_i18n 12, factory 21, factory_card 2; stale_run 4 | 0 |
| PV-21 | `grep -rIlE` of the removed-lane-form pattern over the whole tree excluding `.git`, worktrees, reports, node_modules, specs, testdata | 136 paths: `.moai/specs` 68, docs/instruction 29, Go 14, tests 18, frozen fixtures 3, CHANGELOG and generated codemaps 4 | 0 |
| PV-22 | `go test ./internal/cli -run '^(TestCC_KanbanFlagStrippedBeforeLaunch\|TestGLM_KanbanFlagParity\|TestPrepareKanbanSettingsWritesTransientFile)$' -v -count=1` | three PASS results, `ok ... 0.921s` | 0 |
| PV-23 | `go test ./internal/template -run '^(TestWorkflowRulePathsPinned\|TestContractModeAlwaysLoadedBudget\|TestContractModeConstitutionDriftNotIncreased\|TestContractModeLocalTemplateParity\|TestRuleTemplateMirrorDrift\|TestDeclaredRuleMirrorForks\|TestAllSkillsInCatalog\|TestCatalogHashCoversSkillSubfiles)$' -v -count=1` | six PASS, TWO SKIP (`TestContractModeConstitutionDriftNotIncreased`, `TestContractModeAlwaysLoadedBudget`; reason unobserved — a skip is not a pass), `ok ... 0.254s` | 0 |
| PV-24 | `grep -rlE` per marker constant over `internal` and `cmd`, non-test files | `MOAI_KANBAN_ID` 22; `_LEAD_ADDR` 9; `_LEAD_NAME` 8; `_SETTINGS_INJECTED` 6; `_BACKEND` 12; `_CARD` 7; `MOAI_KANBAN` 8; `_SPEC` 6; `_LABEL` 8. Union of the six factory-read markers by name or constant: 95 files (34 non-test) — SUPERSEDED in v0.8.0 by PV-87: 92 files (31 production, 61 test). `internal/discovery` reads `MOAI_KANBAN_ID` from a live leader process's environment. | 0 |
| PV-25 | `grep -rln` of the board-family symbols outside `internal/kanban`, non-test | only `todo_autodone.go:306` and `todo_landed.go:337`, both `ReadPrimarySpecStatus` (not board); the board state store has no non-test caller outside the package | 0 |
| PV-26 | `moai constitution validate` (tree-built binary) | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`; `4 retired entry/entries skipped` | 0 |
| PV-27 | `grep -c` over `zone-registry.md` for `orchestrator-class\|kanban companion\|Selection Decision Tree\|manager-lead` and `grep -n -i 'companion\|lane\|factory\|foreman\|kanban'`; `grep -c 'ZONE:Frozen'` over the three kanban rules | 0 hits in the registry for each; 0 `[ZONE:Frozen]` in `kanban-dispatch.md`, `-detail.md`, `-mechanics.md` | 0 |
| PV-28 | `moai constitution guard --help`; `moai constitution amend --help` | `guard` takes `--violations` (rule IDs), not a file diff; `amend` takes `--rule CONST-V3R2-NNN` (required), `--before`, `--after`, `--evidence`, `--dry-run` | 0 |
| PV-29 | `grep -rlE -i 'kanban (mode\|companion\|chain\|board\|foreman)\|moai (cc\|glm) -k\|-k --name' .claude internal/template/templates AGENTS.md CLAUDE.md` | 25 paths (v0.3.0 inventory; superseded by PV-45) | 0 |
| PV-30 | `grep -rlE -i 'kanban mode\|moai (cc\|glm) -k\|(-f\|--factory)[ =]lane\|-l, --leader\|codex +-f'` over the READMEs, `docs-site/{content,data,i18n,layouts}`, `AGENTS.md` | 50 paths (v0.3.0 inventory; superseded by PV-45) | 0 |
| PV-31 | `grep -rlE -e '-k --name\|moai (cc\|glm) -k\|(-f\|--factory)[ =]lane\|-l, --leader' internal cmd --include='*.go' --exclude='*_test.go' --exclude-dir=testdata` | 16 paths: `internal/config/{envkeys,defaults}.go`; `internal/cli/{factory_card,glm,cc,factory,factory_lane_relaunch,codex_launcher,codex_factory,kanban}.go`; `internal/kanban/bootstrap.go`; `internal/hook/{session_start_factory_i18n,session_start_factory,session_stale_run,session_start_kanban,session_start_kanban_i18n}.go` — identical when widened from four packages to all of `internal` and `cmd` | 0 |
| PV-32 | `grep -rlE` of the kanban chain and companion identifiers over `internal` and `cmd`, `--exclude='*_test.go'` | 15 paths (listed in acceptance.md RED-K3) | 0 |
| PV-33 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict`, v0.3.0 text | `No findings`; mutant (every `REQ-005` in acceptance.md rewritten) also `No findings` — superseded by PV-46 | 0 / 0 |
| PV-34 | by hand, v0.3.0 | superseded by PV-47 | 0 |
| PV-35 | counts: `grep -rlE '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go' \| wc -l` (and the same with `--include='*_test.go'`); `grep -rIn 'kanban\.' internal cmd --include='*.go' \| wc -l` | 179 importing files, of which 113 test (so 66 production, one under `cmd`); 1,689 qualified references | 0 |
| PV-36 | `go run ./cmd/moai codex -k` | stdout empty, exit 1, stderr `KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead` | 1 |
| PV-37 | `grep -c`/`ls` single-cell checks: `grep -c 'Chain session board' internal/web/screens.templ`; `grep -c 'moai-kanban-foreman' internal/template/catalog.yaml`; `ls .claude/loop.md .claude/skills/moai-kanban-foreman/SKILL.md`; `grep -n 'kanbanBootstrapNotice' internal/hook/session_start.go` | `1`; `2`; both files exist (1254 and 11181 bytes); lines `587` and `599` | 0 |
| PV-38 | `moai constitution validate` re-run; `grep -n -i kanban CLAUDE.md .claude/rules/moai/core/moai-constitution.md .claude/rules/moai/core/agent-common-protocol.md`; `grep -n 'agent-common-protocol' .claude/rules/moai/core/zone-registry.md`; read of `zone-registry.md:370-469` | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`, 4 retired skipped, exit 0. The four sentences: `CLAUDE.md:61`, `moai-constitution.md:11`, `agent-common-protocol.md:27` and `:75` (line 27 also cites `kanban-dispatch-mechanics.md` and `kanban-dispatch.md`). Registered `agent-common-protocol.md` entries: `CONST-V3R2-036..038` (Frozen, anchor `#user-interaction-boundary`, three short clauses), `-039` (Evolvable, `#language-handling`, clause = the header sentence "All agents receive and respond in user's configured conversation_language."), `-040..-046`. Line 27 is in the `#user-interaction-boundary` section; line 75 in the `#language-handling` body. None of the four sentences' text is a registered clause string. | 0 |
| PV-39 | `go test ./internal/hook -run '^(TestFactoryLeadNoticeWorkerCountDrivesLineCount\|TestFactoryLeadNoticeAllSlotsClaimed)$' -v -count=1` | two PASS results (swept count 2), `ok github.com/modu-ai/moai-adk/internal/hook 1.639s` — the two tests the leader-notice change replaces or deletes | 0 |
| PV-40 | `grep -c 'EnvMoaiKanbanLabel'` over `codex_launcher.go factory_m5_test.go codex_factory_retire_test.go codex_debug_trace_test.go factory_m4_test.go envkeys.go`; read of `factory_m5_test.go:51-158` and `factory_card.go:50-100` | launcher 4, `factory_m5_test.go` 2, `codex_factory_retire_test.go` 1, `codex_debug_trace_test.go` 1, `factory_m4_test.go` 1, `envkeys.go` 4. `TestSD_AC003_CodexRelaunchPerCard` asserts the label PRESENT (`:145-147`) under a comment that the card verbs read it; `factory_card.go:62,88-91` read `MOAI_FACTORY_WORKER`. No production reader of the label found. | 0 |
| PV-41 | `grep -rnE 'VerifyRung\|DeepScanDir\|VerifyReentries\|\.Rung\b' internal --include='*.go' --exclude='*_test.go'`; `grep -rn 'verify exit gate\|VerifyExitGate\|Verify Exit Gate' internal --include='*.go'`; `grep -rn 'factory_chain' .claude internal/template/templates internal docs-site/content` (md/go/yaml); `grep -c factory_chain`, `grep -c '.moai/state/factory'`, `grep -c 'chain head'`; `grep -rn -A12 'func RecordPath' internal/kanban` | chain fields: only `internal/kanban/record.go` lines (declarations and comments), no writer or reader elsewhere; verify-gate in Go: no output; `factory_chain`: `factory.md:64,66` in the local copy and the template copy only (2 lines per file); `factory.md` `.moai/state/factory` count 1; `moai.md` `chain head` count 1; `RecordPath` = `filepath.Join(RuntimeStateDirForRoot(projectRoot), sessionID+".json")` (`record.go:189-191`) | 0 / 1 / 0 |
| PV-42 | `grep -c "'factory\.\|\"factory\.\|factory\.[a-zA-Z]*:" internal/web/assets/i18n.js`; `grep -n 'case "factory"\|case "kanban"' internal/web/icons_templ.go internal/web/icons.templ` | `0` existing factory i18n keys (so the key-family rename cannot collide); icon switch has only `case "kanban"` (`icons_templ.go:128`, `icons.templ:72`) | 1 / 0 |
| PV-43 | `go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader\|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1` | two PASS results (swept count 2), `ok github.com/modu-ai/moai-adk/internal/discovery 0.555s` | 0 |
| PV-44 | `grep -rlE '\b(LoadBoard\|WriteBoardState\|AcquireBoardLock\|RecoverBoard\|ParseColumn\|TransitionIntoRun\|BoardState)\b' internal cmd --include='*.go'`; the same symbols in `integration_lock_mutation.go`, `status_read_test.go`, `admission_test.go`; `find internal -name 'board_store*.go' -o -name 'board_lock*.go' -o -name 'board_recover*.go' -o -name board.go -o -name column.go -o -name reconcile.go` | 24 paths, all under `internal/kanban` (verbatim in acceptance.md RED-K4): the 10 board files, the board tests, six further tests (`admission_test.go`, `status_read_test.go:324`, `fix2_probe_test.go`, `fix3_wedge_test.go`, `f1_traversal_test.go`, `kanban_helper_test.go`), and a comment at `integration_lock_mutation.go:24`; the `find` lists 18 paths (10 production, 8 tests) | 0 |
| PV-45 | counts, each one command piped to `wc -l`: non-test Go/templ/js carrying the word (plain `grep -rliE kanban ...`): 190; same under the AC-018 pattern: 189; test files under that pattern: 308 (310 plain); `find internal cmd -iname '*kanban*'`: 22; docs-scope files under the plain Latin word: 151; under the combined pattern with `claude-code` excluded: 156; Korean/Japanese/Chinese spellings in non-test source: 8 files (`internal/web/{app.go,screens.templ,screens_templ.go,viewmodel_ops.go,screens.go,assets/i18n.js}`, `internal/hook/{session_start_factory_i18n,session_start_kanban_i18n}.go`) | counts as listed; the mirrored `claude-code/` docs contain no Latin occurrence and two generic native-language uses (a Chinese "monitoring dashboard", a Korean Agent Teams analogy) | 0 |
| PV-46 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` on the v0.4.0 text (tree-built binary, scratchpad); then on a scratchpad COPY (`scratchpad/probe4/SPEC-LAUNCHER-ENTRY-FLAGS-001`) with every `REQ-005` in acceptance.md rewritten to `REQ-0XX` (`grep -c 'REQ-005'` on the copy: 0) | original: `✓ No findings — all SPEC documents are valid`, exit 0; mutant: the same `✓ No findings`, exit 0. The lint does NOT move when a requirement loses all coverage, so its green is not coverage evidence (PV-14, now measured on four revisions). Also run: `[[ "SPEC-LAUNCHER-ENTRY-FLAGS-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS \|\| echo FAIL` → `PASS`; spec frontmatter `id`, `version: "0.4.0"`, `status: draft`, `phase: "v3.2.0 target"`, `tier: L`; the four sibling artifacts carry no `status:` field (counts 0 each) | 0 / 0 |
| PV-46b | v0.5.0 re-run after the Q18 revision: `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict`; then a scratchpad copy (`scratchpad/probe5-copy`) with every `REQ-019` in acceptance.md rewritten to `REQ-0XX`, linted with `--strict` | original: `✓ No findings — all SPEC documents are valid`, exit 0; mutant (the requirement that carries the Q18 change loses all coverage): the same `✓ No findings`, exit 0 — the lint is again not coverage evidence. Hand count re-run (PV-47): 25 criteria, 25 `Verifies` lines, 25 `Command` lines, 24 requirement lines, every REQ-001..REQ-024 on a `Verifies` line (REQ-001 twice, REQ-022 with REQ-023, REQ-023 again). Ceilings respected: 24 of 25 requirements, 25 of 25 criteria — the Q18 change folded into REQ-019 and AC-020 and the AC-022 stop condition, adding no REQ or AC | 0 / 0 |
| PV-47 | by hand: `grep -c '^### AC-' acceptance.md`; `grep -c '^- Verifies' acceptance.md`; `grep -c '^- Command:' acceptance.md`; `grep -o '^- Verifies REQ-[0-9]*\(, REQ-[0-9]*\)\?' acceptance.md`; `grep -c '^- REQ-' spec.md` | 25 criteria; 25 `Verifies` lines; 25 `Command` lines; 24 requirement lines. Every requirement REQ-001..REQ-024 appears on a `Verifies` line: REQ-001 twice (AC-001 cc/glm, AC-002 codex), REQ-022 with REQ-023 on AC-023, REQ-023 again on AC-024, every other requirement once on AC-003..AC-021, AC-022, AC-025 (AC→REQ map in `spec.md` §C matches). Ceilings: 24 of 25 requirements, 25 of 25 criteria. | 0 |
| PV-48 | `grep -n 'Env.* = "MOAI_KANBAN' internal/config/envkeys.go`; `grep -rn '"MOAI_KANBAN' internal cmd --include='*.go' --exclude='*_test.go'` | literals at `envkeys.go:182` `MOAI_KANBAN`, `:187` `_SPEC`, `:195` `_ID`, `:205` `_LABEL`, `:216` `_SETTINGS_INJECTED`, `:225` `_LEAD_ADDR`, `:238` `_BACKEND`, `:253` `_CARD`, `:273` `_LEAD_NAME`; and `internal/codexwiring/configtoml.go:21` (the MCP env allowlist naming `MOAI_KANBAN_ID` and `MOAI_KANBAN_BACKEND`); no other non-test string literal carries a marker name | 0 |
| PV-49 | `grep -n -i 'kanban\|factory-mode' docs-site/vercel.json`; read of `vercel.json:170-212`; `grep -n -i kanban docs-site/data/menu/main.yaml docs-site/layouts/index.html docs-site/i18n/en.yaml`; `grep -rn -i kanban docs-site/content/en/{advanced,multi-llm,core-concepts}/_meta.yaml`; `ls docs-site/content/*/advanced/factory-mode.md` | rules `:183-201` redirect `/advanced/factory-mode` and `/multi-llm/factory-mode` (locale and bare forms) to the kanban pages; `:202-211` `manager-kanban` → `manager-lead`; menu `main.yaml:159-162,493-496,725-728`; banner `layouts/index.html:56-76` and `i18n/en.yaml:83-85`; `_meta.yaml` listings; `advanced/factory-mode.md` exists in en (11,173 B), ko, ja, zh — so the factory-mode page is shadowed by the redirect today | 0 |
| PV-50 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0, no output | 0 |
| PV-51 | `grep -n '/loop' internal/hook/session_start_factory_i18n.go`; `grep -c` of notice string-table fields | the queue sentence at `:89` (en: "the kanban foreman loop (bare `/loop`)"), `:139` (ko), `:184` (ja), `:229` (zh); 34 references to `leaderFreeSlots`, `leaderSlotsNone`, `laneJoin`, `laneJoinNoCount`, `leaderManual`, `entryGuide` (this row first read 28; that figure did not reproduce and PV-68 measured 34); `session_start_factory.go:229` builds `FactoryFreeSlots`, `:240` `leaderFreeSlots` | 0 |
| PV-52 | `wc -c` of the rule files; read of `update_archive.go:25-70`; bounded search `grep -rniE 'legacyRule\|obsolete\|removedFiles\|deprecatedFiles\|staleRule\|retiredRule\|legacyRuleFiles' internal/cli internal/template --include='*.go' --exclude='*_test.go' -l` | `kanban-dispatch.md` 26,352 B local and 26,030 B template, `-detail.md` 39,965 B, `-mechanics.md` 18,909 B; `legacySkillIDs` archives retired skills to `.moai/archive/skills/v2.16/` and its guard test `TestLegacySkillIDsNotEmbedded` keeps the list disjoint from the embedded skills; the rule-file search found only `internal/cli/memory.go` (unrelated) — **SUPERSEDED in v0.5.0 by PV-54..PV-59: that search used the wrong terms and missed the managed-root clean, so the conclusion "no stale-rule cleanup exists" was wrong** | 0 |
| PV-53 | `grep -rn '"kanban' internal cmd --include='*.go' --exclude='*_test.go'`; `find internal cmd docs-site/content -iname '*kanban*'`; `grep -rnwE 'factory( :=\|,\| =)' internal cmd --include='*.go'` | the string-literal lines beginning with "kanban" (the generated web templates, the web events and screens, error texts in `internal/kanban`, two hook timing laps, a coverage key — summarized in research.md §R3; not counted); the file/directory names listed in RED-N2 plus 12 docs pages; four non-LSP/TUI local identifiers named `factory` (`factory_handoff_recover.go:20`, `role_naming_m3_notice_test.go:129`, `stale_run_m1_test.go:118`, `runtime_census_test.go:75`) | 0 |

| PV-54 | `grep -rliE 'namespace.protect\|user-owned namespace\|protected namespace\|namespace protection' internal .claude/rules .moai/docs CLAUDE.md AGENTS.md --include='*.go' --include='*.md'`; read of `update_namespace_protect.go`, `update_destructive_registry.go`, `update_cleanup.go:1-452`, `update_residue_cleanup.go:1-160`, `v2_detection.go:1-160`, `deploy.go:60-210`, `update_template_sync.go:360-470`, `update_archive.go:25-140,285-350`, `plan.go:152-259`; `grep -rn 'archiveLegacySkills(\|removeDeprecatedFile\|scanDeprecatedPaths(' internal/cli` | the namespace-protection contract (`SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001`): `IsUserOwnedNamespace` classifies harness and user skills/agents; `backupUserOwnedNamespace` to `.moai/backups/update-<ISO>/`; `assertNoUserOwnedNamespaceTouch` sentinel `UPDATE_USER_NAMESPACE_VIOLATION`. The managed targets include `.claude/rules/moai` (`deploy.go:75-78`); the clean is a template-sync stage (`update_template_sync.go:388-421`); `legacySkillIDs` archive runs inside that stage before the removal and its failure only warns (`:402-407`). `defs.DeprecatedPaths` is acted on only by clean-reinstall and the v3 residue sweep | 0 |
| PV-55 | scratch probe (a `_test.go` in `internal/cli` created for the run and removed after it; `git status --short` afterwards: only the SPEC directory untracked): `t.TempDir()` project with `.claude/rules/moai/workflow/kanban-dispatch.md` ("USER MODIFIED carried") and `.claude/rules/moai/workflow/zz-retired-probe.md` ("USER MODIFIED not carried"); `deploy.CleanMoaiManagedPaths(root, &out, template.EmbeddedTemplates())`; `go test ./internal/cli -run '^TestZZT1399Probe$' -v -count=1` | PASS; live tree: both files gone; backup files: `[.moai-backups/20261002_121511/pre-clean/.claude/rules/moai/workflow/zz-retired-probe.md]` only; output: `✓ Removed .claude/rules/moai (backed up 1 unmanaged file(s))`; so a user-modified template-carried rule is overwritten without a backup today, while a file the template does not carry is backed up. Probe essentials (to recreate it): write the two files, build `tmplFS` with `template.EmbeddedTemplates()`, call the clean, log `os.Stat` of both paths and the files under `.moai-backups` | 0 |
| PV-56 | `go test ./internal/cli/update/deploy -run '^(TestCleanMoaiManagedPaths_BackupsUnmanagedFiles\|TestCleanMoaiManagedPaths_BackupFailureAbortsRemoval)$' -v -count=1`; read of `deploy_preclean_backup_test.go:40-101` | two PASS results (swept count 2), `ok .../internal/cli/update/deploy 0.549s`; the first asserts the unmanaged file reaches the backup byte-identical, the managed file does not, the root is removed, and the output contains "backed up 1 unmanaged file" | 0 |
| PV-57 | read of `update.go:480-500`, `update_residue_cleanup.go:50-160`, `v2_detection.go:82-147`, `defs/dirs.go:25-68`; `grep -n 'DeprecatedPaths'` | the v3 residue branch `return cleanupErr` aborts the update; its sweep deletes after backup regardless of modification (`os.RemoveAll`); a registered path that exists sets `V2DetectedViaDeprecatedPath`, and `IsV2 = !V3VersionConfirmed && (...)`; the registry is a 40-entry slice guarded by `dirs_test.go` and `TestDeprecatedPaths_NoTemplateCollision` | 0 |
| PV-58 | `git log --oneline -- internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md \| wc -l`; `git tag --list 'v3*' \| wc -l`; `wc -c` of the rule files; `grep -c -e 'Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.' -e 'preload `AskUserQuestion` via `ToolSearch`' .claude/rules/moai/core/agent-common-protocol.md` | 65 revisions; 18 v3 tags; sizes in PV-52; the registered Frozen clause strings (`Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.` and ``preload `AskUserQuestion` via `ToolSearch` ``) match 2 lines (CONST-V3R2-038's string is a substring of the first) | 0 |
| PV-59 | `/private/tmp/.../moai-t1399 constitution validate` re-run in the final revision; `git status --short` | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`, exit 0; only the SPEC directory untracked | 0 |

Rows PV-60 to PV-72 are the v0.7.0 re-measurements of the plan-audit iteration 1 findings. They were taken in this revision against HEAD `b9242da00ce489c4f26efb5f6d447ccafef08c54`, whose content outside this SPEC directory equals tree `a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2` (PV-71). The binary rows used `scratchpad/moai-t1399`, built from `a6d3e6fd4` earlier in this plan run (its `version` prints no commit; the tree content it was built from is the code content of HEAD, PV-71); no measurement here was replaced by reading source because of a refusal — one compound shell command was refused by the worktree guard and re-run as plain commands (Gaps).

| ID | Command | Observed | Exit |
|----|---------|----------|------|
| PV-60 | `go test -overlay <scratch>/probe7/overlay.json ./internal/cli -run '^TestZZT1399EntryProbe$' -v -count=1` — a scratch test (source essentials: for each of 18 cc/glm argument lists call `parseLauncherEntry(args)`, then `parseFactoryLaneLabel(entry.Rest)` and `resolveFactoryBranch(entry.FactoryEnabled, isLane)` and print one `CCPARSE` line; for 6 codex lists call `runCodex(&cobra.Command{Use: "codex"}, args)` with captured stderr and print one `CODEX` line; the overlay maps `internal/cli/zz_t1399_probe_test.go` to the scratch file, so nothing is written into the tree; `git status --short` empty afterwards) | `-f --name lane-2`, `-f -n lane-2`, `--factory --name=lane-2` → `factoryEnabled=true laneLabel="lane-2" isLane=true branch=2`; `-f --name leader-r7` → `isLane=false branch=1`; `--lane`, `--lane lane-2`, `--lane=lane-2`, `--lane 3`, `--lane leader-2` → `factoryEnabled=false`, the tokens left in `rest`, `branch=0`, no error; `--lane -f` → `factoryEnabled=true rest=["--lane"] branch=1`; `-l` → parse error `--leader requires a leader label`; `-l lane-2`, `-l 3`, `-l=lane-2`, `-l leader-2` → `--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target`; `-l -f`, `-l -k`, `-l --name lane-2` → `--leader requires a leader label`; codex `--lane`, `--lane lane-2`, `--lane=lane-2`, `--lane 3`, `--lane leader-2` → exit 1, stderr `unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane\|lane-<n>]] [--factory-run <id>] [-- codex-args...] \| moai codex status \| moai codex app`; codex `--lane -f` → exit 1, `FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; …`; `ok … 0.857s` | 0 |
| PV-61 | `grep -n '^func \|^type \|^const \|^var ' internal/cli/kanban.go`; `grep -rn 'exportKanbanLaunchFacts\|exportFactoryLaunchFacts' internal cmd --include='*.go'`; `grep -rnE '\b(parseKanbanFlag\|enterKanbanMode\|…)\b'` over the nine kanban-only declarations and the three constants (non-test); the same for the factory-shared helpers; read of `kanban.go:500-575`, `factory.go:325-410`, `glm.go` and `cc.go` branch structure | 25 funcs, 2 types, 2 single constants, 3 constant blocks: 27 func/type declarations, 9 kanban-only (`parseKanbanFlag`, `enterKanbanMode`, `enterKanbanCompanionMode`, `companionRegistryPath`, `resolveCompanionName`, `parseCompanionLabel`, `rejectKanbanOnCG`, `kanbanBranch`, `resolveKanbanBranch`) and 18 factory-shared. `exportKanbanLaunchFacts` callers: `glm.go:271` (factory leader branch, `case factoryBranchLeader` at :256), `glm.go:304` (factory lane branch, :282), `glm.go:334,352` and `cc.go:300,320` (kanban branches), `cc.go:222,260` and `codex_launcher.go:951`, `codex_factory.go:144` through `exportFactoryLaunchFacts` (`kanban.go:565-567`), tests `kanban_launch_facts_test.go:48,70,97`. `parseLauncherEntry` refuses `-k` with `-f` (`factory.go:363-366`), so `entry.Spec` is empty on every factory path. The factory-shared helpers have factory callers directly (`factory.go:311,315,379,391,653,677,680,732,733`; `cc.go:211,224,225`; `glm.go:257,273,274`) or through `appendLeaderName` and `resolveLeaderName` (`resolveLeaderName`, `noteLegacyLeaderRegistryEntries`, `claimName`, `leaderRegistryPath`, `leaderNameArgs`, `allDigits`) | 0 |
| PV-62 | `grep -rn 'EnvMoaiKanbanLabel\|EnvMoaiKanbanSpec\|EnvMoaiKanban\b' internal cmd --include='*.go' --exclude='*_test.go'`; the same with `--include='*_test.go'`; `grep -rlE` with `\| wc -l` for both | non-test readers (lines in research.md §R16): `envkeys.go` definitions and comments, `launcher_blockcap_infinite.go:57`, `codex_launcher.go:317,319,320,959,960,1037`, `kanban.go:194-204,321,334,343,545,553`, `ptycaptest/harness.go:67,68,70`, `session_start_kanban.go:56,59,183`, `session_start_record.go:110,191,198`, comments `factory.go:665,699`; counts: 8 non-test files and 24 test files (13 `internal/cli`, 11 `internal/hook`, names in research.md §R16) | 0 |
| PV-63 | `grep -rnE '\b(CompanionRoles\|CompanionLauncher\|companionLaunchers\|CompanionLabel\|CompanionNumberLabel\|SplitCompanionLabel\|isCompanionRole\|LeaderSocketPath\|LeaderLabel\|SplitLeaderLabel)\b' internal cmd --include='*.go' --exclude='*_test.go'` minus `bootstrap.go`; `grep -rnE '\b(kanbanLeaderNotice\|kanbanCompanionNotice\|kanbanBootstrapNotice\|kanbanBootstrapNoticeForSource\|kanbanRoleFromEnv)\b'`; `grep -rhoE 'kanban\.[A-Za-z]+' internal/web internal/statusline internal/cli/doctor*.go` (non-test); `grep -rn 'kanban-dispatch\|moai-kanban-foreman' internal cmd --include='*.go' --exclude='*_test.go'`; `grep -rnoE` of the RED-K3 identifiers per file; read of `record.go:125-175`, `bootstrap.go:236-395`, `role.go` | companion-symbol readers: `cli/kanban.go:211,380-385,590`, `hook/session_start_kanban.go:131-151,233`, `hook/session_start_record.go:192`, `kanban/record.go:151` (`WithRole` → `isCompanionRole`), comment `role.go:61`; leader-label readers kept: `factory_discovery.go:167`, `factory.go:177,313,320,515`, `codex_factory.go:60`, `session_start_factory.go:196`, `kanban.go:240,422`; hook notice builders: `session_start.go:587,599` and comments; web/statusline/doctor use `RecordPath`, `Record`, `ReadAll`, `LoadFactoryRegistry`, backlog and landing symbols only; 13 comment lines in 11 non-test Go files name `kanban-dispatch.md` or `moai-kanban-foreman`; the RED-K3 per-file token map in research.md §R16 | 0 |
| PV-64 | `git ls-files AGENTS.local.md CLAUDE.local.md AGENTS.md CLAUDE.md`; `git check-ignore -v CLAUDE.local.md`; `grep -n -i 'kanban\|-f lane\|moai cc -k\|moai cc -f' AGENTS.md AGENTS.local.md`; `grep -n 'AGENTS.local' CLAUDE.md`; `grep -rn -i 'kanban\|-f lane' internal/template/templates/AGENTS.md.tmpl`; the three RED-D5, RED-D7, and RED-D1 commands (acceptance.md ledger) | `AGENTS.local.md`, `AGENTS.md`, `CLAUDE.md` tracked; `CLAUDE.local.md` ignored (`.gitignore:276`) and absent from the worktree; `AGENTS.md:125` "**Codex factory lanes (`-f lane`)**" (no word); `AGENTS.local.md:217` "Kanban(`moai cc -k`) / Factory(`moai cc -f N`) …"; `CLAUDE.md:167,172` import `@AGENTS.local.md`; `AGENTS.md.tmpl` no match (exit 1); removed-form grep 46 files (AGENTS.md and AGENTS.local.md among them); word grep with `AGENTS.local.md` added 157 files (`AGENTS.local.md` listed once); `find` outside `internal` and `cmd`: 19 paths (3 image files, 1 skill directory, 12 docs pages, 3 rules) | 0 |
| PV-65 | `env MOAI_GR_BASE=a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 go test ./internal/template -run '^TestContract' -v -count=1` (output to a scratch file); read of `contract_mode_guided_test.go:1-60,640-740` | `TestContractModeConstitutionDriftNotIncreased` PASS (6.21 s, base and current non-OK pairs `[]`), `TestContractModeAlwaysLoadedBudget` PASS (0.51 s, +0 characters); `TestContractModeGuidedPreservation` FAIL (text outside the contract blocks differs from the base in several skill documents; the first mismatch printed is `internal/template/templates/.claude/skills/moai/SKILL.md` at line 169; "stripped 40 blocks across 26 copies"), `TestContractModeChangeSetAllowlist` FAIL (`changed path outside the allowlist: .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/acceptance.md`), `TestContractModeEmitterSites` FAIL (unclassified Kickoff documents at the base: `moai-mcp-tools-catalogue.md`, `auto-semantics.md`, `contract-autonomy.md` and their template copies); the file header says a test skips when `MOAI_GR_BASE` is unset and an acceptance run must set it; `grKickoffClasses` (`:652-691`) is keyed by base-ref paths and `kanban-dispatch.md` is at `:679` | 1 (the three FAILs; unrelated to this SPEC) |
| PV-66 | `hugo --source docs-site --minify --gc --destination <scratch>/public` (output to a scratch file); `sh <scratch>/probe7/parity.sh` (the `.claude/skills/hns-oss-docs-verify/SKILL.md` §4 ratchet with scratch outputs); `grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md`; `git status --short` | Hugo exit 0, summary table ends `Total in 3099 ms`, page counts 189, 187, 187, 187, `grep -c -i -e WARN -e ERROR` 0; ratchet: 51 divergent pages, 51 baseline lines, `comm -23` empty, `comm -13` empty, 155 ko pages compared; README H2 counts 12, 12, 12, 12; the tree unchanged | 0 |
| PV-67 | `grep -n '^status:\|^id:'` over `SPEC-KANBAN-BOOTSTRAP-001`, `SPEC-KANBAN-WORKTREE-001`, `SPEC-WEB-TODO-QUEUE-001`, `SPEC-KANBAN-BOARD-001`, `SPEC-KANBAN-RENAME-001`; `grep -rn 'REQ-TOSQ-018' .moai/specs/SPEC-TODO-SQLITE-001`; `grep -rln 'TOSQ-018\|literal-cleanliness' internal --include='*.go'`; read of `state_dir.go:15-25`, `events.go:30-36`, `backlog_sqlite_test.go` | BOOTSTRAP-001 and WORKTREE-001 `draft`; WEB-TODO-QUEUE-001, KANBAN-BOARD-001, KANBAN-RENAME-001 `completed`; REQ-TOSQ-018 = "no active template source under … `state/kanban`" with an allowlist for the intentional old-layout fallback reader; no Go test enforces the sweep (the only Go hits are the comment in `state_dir.go:19-22` and an unrelated `AC-TOSQ-018` label in `backlog_sqlite_test.go`); `events.go:33-34` "SSE event KEY stays "kanban" — it is a frontend-visible contract" | 0 |
| PV-68 | `git diff --quiet a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 HEAD -- .claude/rules/moai/core/zone-registry.md`; the same over `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/spec.md`; `grep -cE '\b(leaderFreeSlots\|leaderSlotsNone\|laneJoin\|laneJoinNoCount\|leaderManual\|entryGuide)\b' internal/hook/session_start_factory_i18n.go`; `grep -n -B1 -A4 '^paths:'` over the three rules; read of `moai-constitution.md:9-13`, `zone-registry.md:281-300`; `moai constitution validate --help` (audit report E-9) | zone-registry exit 0; spec.md control exit 1; i18n references 34; the three `paths:` globs name `**/kanban-dispatch*.md` at `kanban-dispatch-detail.md:3`, `kanban-dispatch-mechanics.md:3`, `cross-session-messaging-detail.md:3`; `moai-constitution.md:11` is the kanban sentence and `:12` the Frozen `[ZONE:Frozen] [HARD] AskUserQuestion is the sole user-facing question channel` bullet, registered as `CONST-V3R2-025` (`#moai-orchestrator`, `zone-registry.md:283-289`) | 0 / 1 |
| PV-69 | `grep -n -e '-f lane' internal/hook/session_start_factory.go internal/hook/session_stale_run.go internal/cli/factory_card.go`; `grep -c -e '-f lane' internal/hook/session_start_factory_i18n.go`; `grep -c -i kanban internal/cli/update_archive.go internal/kanban/state_dir.go`; the widened AC-024 pattern over non-test Go (`-k --name\|moai (cc\|glm) -k\|(-f\|--factory)[ =]lane\|-l, --leader\|(-f\|--factory)[ =](<N>\|N\b\|[0-9])`) | stale-run hint `session_stale_run.go:92,103,114,125` (`'moai cc -f lane-<n>'` in four locales), card errors `factory_card.go:95,98` (`rejoin with -f lane`), notice comments `session_start_factory.go:162,170,258` and the per-lane line at `:207` (RED-7), 12 matches in `session_start_factory_i18n.go`; word-bearing lines: `update_archive.go` 0, `state_dir.go` 2; the widened pattern lists the same 16 files as RED-S1 (the extra `-f <N>` hits are `envkeys.go:278`, `factory.go:620,621,630,631,710,860`, all in files already listed) | 0 |
| PV-70 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` (scratch binary); the plan-auditor traceability verb (`.claude/agents/moai/plan-auditor.md` Group 4, copied verbatim to `<scratch>/probe7/trace.sh`) on `spec.md` and `acceptance.md`; the same two on a scratch COPY (`<scratch>/probe7/mut1`) in which every `REQ-019` outside its definition line is rewritten to `REQ-0XX` in all of `acceptance.md` and `spec.md` (`grep -c 'REQ-019'`: acceptance 0, spec 1); hand counts `grep -c '^### AC-'`, `'^- Verifies'`, `'^- Command:'` in acceptance.md and `grep -c '^- \*\*REQ-'` in spec.md; `grep -o '^- Verifies REQ-…'` | lint original `✓ No findings`, exit 0; verb on the original `COLLECTED: 24 REQ definitions (acceptance input: read)` with no `UNCOVERED` and no `ORPHAN` line; verb on the mutant `COLLECTED: 24`, `ORPHAN: REQ-0` (the rewritten token), `UNCOVERED: REQ-019` — the verb moves when a requirement loses coverage, so its silence on the original is measured, not blind; lint on the mutant still `✓ No findings`, exit 0 — the lint remains silent on lost coverage, so its green is not coverage evidence. A first mutant that rewrote `REQ-019` only in acceptance.md and the §C table left the verb silent, because history lines of spec.md name `REQ-019` beside `AC-020` and the verb counts any line that names both — a reminder that its "mapped" is generous and the hand count stays the evidence. Hand counts: 25 criteria, 25 `Verifies` lines, 25 `Command` lines, 24 requirement lines; `Verifies` lines cover REQ-001 (twice), 002 to 016, 017 with 024, 018, 019, 020, 021, 022 with 023, 023, 022 — every REQ-001..REQ-024 at least once; ceilings 24 of 25 requirements, 25 of 25 criteria | 0 / 0 / 0 |
| PV-71 | `git rev-parse a6d3e6fd4`; `git rev-parse HEAD`; `git diff --quiet a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 HEAD -- . ':!.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001'` | `a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2`; `b9242da00ce489c4f26efb5f6d447ccafef08c54`; exit 0 — no content differs outside this SPEC directory | 0 |
| PV-72 | read of `internal/kanban/bootstrap.go:236-395` (`NextFactoryLaneNumber` at `:370-384`), `internal/kanban/factory_slots.go:100-210`, `internal/cli/factory.go:374-395`; `grep -n 'never backfills' README.md` | the lane label is one past the highest LIVE canonical claim after `PruneFactoryDeadClaims`, 1 when nothing is claimed; the `-f lane` parse computes it at `factory.go:386` and the branch claims it in one IMMEDIATE SQLite transaction (`ClaimFactoryLane`, dead claims removed first); `README.md:80` states the same rule (a dead lane's claim no longer blocks its number; auto-assignment never backfills a gap) | 0 |


Rows PV-73 to PV-90 are the v0.8.0 evidence for plan-audit iteration 2 (FAIL 0.79; findings D29-D45). Every measurement was taken in this revision on HEAD `5445e296caa48e6ab9821afe808eac2e7385e897` (the commit iteration 2 audited), whose content outside this SPEC directory equals tree `a6d3e6fd4` (PV-73), with Go `go1.26.8` (the toolchain the module selects). The compile proof builds only a SCRATCH COPY of the Go tree; `git status --short` after the first probe run showed only the new untracked `probe/` directory. Where a command was refused rather than run, the Gaps paragraph says so.

| ID | Command | Observed | Exit |
|----|---------|----------|------|
| PV-73 | `git rev-parse HEAD`; `git diff --quiet a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 HEAD -- . ':!.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001'`; `go version` in the tree; `diff -rq` of the pristine scratch copy P0 (the tree the stage data was derived from) against the checkout for `internal`, `cmd`, `e2e`, `pkg`, `scripts`, `test`, and `diff -q` for `go.mod` and `go.sum`; baselines on P0: `go build -gcflags=-e ./...`, `GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...`, `go vet ./...`, `GOOS=windows GOARCH=amd64 go vet ./...` | HEAD `5445e296caa48e6ab9821afe808eac2e7385e897`; `git diff --quiet` exit 0 (nothing differs outside this SPEC directory); `go version go1.26.8 darwin/arm64`; every `diff` empty; baselines: host build exit 0 and empty, windows build exit 0 and empty, host vet exit 0 and empty, windows vet exit 1 with exactly three lines — `# github.com/modu-ai/moai-adk/internal/cli/worktree`, `# [github.com/modu-ai/moai-adk/internal/cli/worktree]`, `vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs` — a failure that predates this SPEC and is the baseline of every windows vet block below | 0 / 0 / 0 / 0 / 0 / 0 / 1 (baseline) |
| PV-74 | the committed runner, one invocation: `go run probe.go -src <checkout> -work <scratch>/replay-final -data .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe -from M0 -to M11` (run from the probe directory; vet on) — it copies the Go tree to the scratch directory, applies M2, M3, M4, M5a, M5b, M6a, M6b from `probe/patches/` and M7, M8, M9, M10 programmatically, and after every stage runs host and windows `go build -gcflags=-e ./...` and `go vet ./...` | every stage prints four blocks that are empty with exit 0, except the windows vet block, which prints exactly the baseline of PV-73 (exit 1); the stage table and the verbatim output follow this table. The header comment of `probe.go` was reworded after this run started (no code line changed). | 0 (the runner) |
| PV-75 | negative control for the lock substrate: on the M5b tree of the committed stage data, delete the ten files the v0.7.0 plan listed as the board family (`board.go`, `board_store.go`, `board_recover.go`, `board_lock.go`, `board_lock_unix.go`, `board_lock_windows.go`, `board_lock_clear_unix.go`, `board_lock_clear_windows.go`, `column.go`, `reconcile.go`) and run `go build -gcflags=-e ./...` on the host and on `GOOS=windows` | 25 compile errors on each OS (listings below the table): `backlog_store.go` 6, `integration_lock_mutation.go` 7, `slot_lease.go` 7, `role.go` 1 (`BoardDir`), and 4 in the per-OS mutation files — `integration_lock_mutation_unix.go` and `slot_lease_mutation_unix.go` (`ClearStaleReport`) on the host, `integration_lock_mutation_windows.go` and `slot_lease_mutation_windows.go` (`ClearStaleReport` and `clearStaleLockAtPath`) on windows | 1 / 1 |
| PV-76 | negative control for the locale helpers: on the M5b tree, delete `internal/hook/session_start_lang.go` (the file that holds `langEnglish` and `operatorLang`) and run `go build -gcflags=-e ./internal/hook/` | 8 compile errors (listing below the table): `session_stale_run.go`, `session_start_factory_i18n.go`, `factory_messages.go`, `session_start.go` | 1 |
| PV-77 | negative control for the `-l` short: on the tree of stage M2 (pristine P0 plus `probe/patches/M2.patch`), delete the constant `leadFlagShort` and run `go build -gcflags=-e ./...`; the alias and socket-directory collisions of rows 5 and 6 of design.md §7.1 were observed as redeclaration errors in an interim replay of the same stage data before the corrections entered it (`internal/cli/codex_launcher.go:769:2: codexFactoryRefusalDiag redeclared in this block`; `internal/kanban/bootstrap.go:309:2: factorySocketDir redeclared in this block`) | 4 compile errors, all `internal/cli/codex_factory.go` lines 42, 43, 50, 51: `undefined: leadFlagShort` (the Codex entry parse still reads it until M3); the two interim redeclaration errors as quoted; deleting the nine `kanban.go` declarations alone (interim tree, before the retired-entry file existed) failed with `internal/cli/codex_launcher.go:771:27: undefined: kanbanUnsupportedBackendSentinel` and `:834:17`, `:834:45`, `:835:29`, `:835:78` `undefined: kanbanFlagShort` / `kanbanFlagLong`, plus the callers in `cc.go`, `glm.go`, and `factory.go` (the tokens therefore move to the retired-entry file in M5a) | 1 |
| PV-78 | negative controls for the test helpers: on the M5b tree, delete `internal/hook/session_start_env_helper_test.go` and run `go vet ./internal/hook/`; on the M6b tree, delete `internal/kanban/test_helpers_test.go` and `test_helpers_windows_test.go` and run `go test -c -o /dev/null -gcflags=-e ./internal/kanban/` on the host and on `GOOS=windows` (the test compile lists every error; `go vet` stops at the first) | `clearKanbanEnv` is called by `session_start_additional_context_test.go`, `session_start_factory_provider_test.go`, `session_start_factory_test.go`, and more (first ten errors listed below); `runtimeIsWindows` by `f3_f4_probe_test.go`, `integration_lock_cross_test.go`, `slot_lease_cross_test.go`, `status_read_test.go`; `runGitAt` by `f3_f4_probe_test.go`; `deadPID` by `integration_lock_rotation_test.go`; windows only, `deadPIDWin` by `integration_lock_mutation_windows_test.go:52`; `readFileBytes` is called by no retained test once the board cases are gone, so it is not needed | 1 / 1 / 1 |
| PV-79 | the M7 stage of PV-74; the interim collision list from the replay of the same data before the corrections: `bootstrap.go:309:2: factorySocketDir redeclared` (`kanbanSocketDir` left alive by M6), `codex_launcher.go:769:2: codexFactoryRefusalDiag redeclared` (alias `codexKanbanRefusalDiag` left alive by M5a), file-name collision `kanban_dispatch_test.go` to `factory_dispatch_test.go` (already exists) | the M7 runner note in the stage table; after the corrections no collision remains and the stage is clean | 0 |
| PV-80 | negative control for `role.go`: on the M6b tree, delete the whole file `internal/kanban/role.go` and run `go build -gcflags=-e ./...` | 10 compile errors in the first failing package (listing below the table): `bootstrap.go` 4 uses of `RoleLeader`, `record.go:151` `RoleLeader` and `RoleLane`, `todo_owner_label.go` `IsLegacyLeaderSpelling`, `RoleLeader`, `legacyLeaderSpelling`, `RoleLane` — the four symbols that stay | 1 |
| PV-81 | the M8 stage of PV-74 | the M8 runner note in the stage table: package clauses renamed, qualified references renamed, shadowing local occurrences renamed; the single shadowing site is `internal/hook/stale_run_m1_test.go:118` (`factoryRun`, four occurrences); the three other design.md §4.7 sites are in files that do not import the package, so the Gap of design.md §4.7 is closed | 0 |
| PV-82 | the M9 stage of PV-74; interim prototype of the same edits on the M9 tree before the view-model edits entered the stage: delete `ChainRoles`, then delete `ChainVM`, `RoleVM`, `buildChain`, `chainRoleRecords`, `chainCardID`, and `go build -gcflags=-e ./internal/web/` | deleting `ChainRoles` alone: `viewmodel_ops.go:262:43`, `:263:23` (in `chainRoleRecords`) and `:330:23` (in `buildChain`) undefined; deleting the model: `viewmodel_ops.go:110:13` (`OverviewVM.Chain`), `:119:13` (the screen model's `Roles`), `:615:13` and `:677:11` (`buildChain` in the Overview and factory builders), `:615:45` and `:677:48` (`chainCardID`), `:628:78` (`buildAttention`'s parameter), `:676:18` (`chainRoleRecords`) — the Overview builder and `buildAttention` are the second reader; after the view-model edits the stage is clean and five web test files lose the cases that use the model | 1 |
| PV-83 | `go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/entry_probe.go -root .` (the committed program; one invocation from the repository root; it writes the committed probe test into a temporary directory outside the tree and runs `go test -overlay`) | the verbatim output is below the table (the same rows as RED-11 to RED-14); `git status --short` after the first run listed only the new untracked `probe/` directory, so the probe wrote nothing into the tree | 0 |
| PV-84 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict` (scratch binary `moai-v8`, built by `go build -o ... ./cmd/moai` from this tree; its `version` prints `moai-adk v3.1.3` and no commit, so its judging build is identified by the tree it was built from, HEAD `5445e296caa48e6ab9821afe808eac2e7385e897`, per verification-claim-integrity §2.2); the plan-auditor traceability verb (`plan-auditor.md` Group 4, saved verbatim as `scratchpad/probe7/trace.sh`) on `spec.md` and `acceptance.md`; hand counts; the plan-auditor CN-4 ordering verb (Group 6, extracted verbatim from `plan-auditor.md` into a scratch script) on `plan.md` and `acceptance.md` | lint: `✓ No findings — all SPEC documents are valid`, exit 0; collection verb: `COLLECTED: 25 REQ definitions (acceptance input: read)`, no `UNCOVERED` and no `ORPHAN` line, exit 0; positive control on a coverage-scrubbed scratch copy (every `REQ-025` outside its definition line rewritten to `REQ-0XX` in `spec.md` and `acceptance.md`): `COLLECTED: 25 REQ definitions`, `ORPHAN: REQ-0`, `UNCOVERED: REQ-025`, so the silence on the real files is measured and not blind; hand counts: 25 criteria, 25 `Verifies` lines, 25 `Command` lines, 25 requirement lines (ceilings 25 and 25, both at the ceiling); `Verifies` lines naming each requirement: REQ-001:2 REQ-002:1 REQ-003:1 REQ-004:1 REQ-005:1 REQ-006:1 REQ-007:1 REQ-008:1 REQ-009:1 REQ-010:1 REQ-011:1 REQ-012:1 REQ-013:1 REQ-014:1 REQ-015:1 REQ-016:1 REQ-017:1 REQ-018:1 REQ-019:1 REQ-020:1 REQ-021:1 REQ-022:2 REQ-023:2 REQ-024:1 REQ-025:1 (every REQ-001..REQ-025 at least once; REQ-001, REQ-022, REQ-023 twice by design); CN-4 verb: `COLLECTED: 12 milestones in plan order (M0 M1 M2 M3 M4 M5 M6 M7 M8 M9 M10 M11), 35 exit bindings, 46 ordering candidates` — no `CONFLICT`, no `GAP` (the verb reads M5a and M5b as the one label M5, so it collects 12 labels for the 13 milestones; its 46 candidates were read one by one and none orders an acceptance criterion across the plan order, PV-88) | 0 |
| PV-85 | the RED-now commands of every criterion this revision changed (AC-003, AC-004, AC-005, AC-012, AC-018, AC-020, AC-023), re-run on this tree: the launcher cells with the tree binary, the grep and `find` cells through a script that prints exit codes and line counts (the shell's `grep` is a function that supports `-P`; the two PCRE cells were re-run in that shell) | listing below the table: every cell reproduces its ledger value (RED-1, RED-2, RED-3, RED-5, RED-6, RED-10, RED-K2 exit 1 with the ledger stderr; RED-K3 15; RED-K4 27 (new pattern); the RED-K4 board-family `find` 18 and its positive control 4; RED-N1 189; RED-N2 22; RED-N3 179; RED-C1 exit 1 three lines; RED-C2 two lines; RED-C3 39965, 18909, 26030; RED-C4 three counts of 1; RED-D1 157; RED-D2 4; RED-D3 0 exit 1; RED-D5 46 with the new pattern and 46 with the old (unchanged union); RED-D8 16; RED-D7 19; RED-S1 16) | 1 (RED-1, RED-2, RED-3, RED-5, RED-6, RED-10, RED-K2) / 0 (the grep and `find` cells) / 1 (RED-C1, RED-D3) |
| PV-86 | baselines on the pre-change tree: `go test ./internal/kanban -run '^(TestBoardLockWaitBudgetDerivedFromNamedInputs\|TestBoardLockWaitBudgetCoversSerializedMutations\|TestBoardLockRetryWaitIsNotLockstep\|TestBacklogLockStuckHolderSurfacesBoundedNamedError)$' -v -count=1`; `go test ./internal/kanban -run 'Foreman' -v -count=1`; `go test ./internal/template -run '^TestBacklogJSONDisclosure_' -v -count=1` | the four wait-budget tests PASS (0.00 s, 0.00 s, 0.00 s, 3.32 s, `ok ... 4.318s`); the seven foreman tests PASS (`TestForemanQueueWatchResolvesCanonicalStateDir`, `TestForemanQueueWatch_FiresOnMutation`, `_ShippedJSONTargetIsSilent`, `_FiresWithStaleJSONPresent`, `_WatchTargetsAgree`, `_DBOnlyTargetMissesWALDeferral`, `_SeesWALDeferredCommit`; `ok ... 129.305s`); the two template tests PASS (`ok ... 0.989s`); the tests that read the foreman skill by path are `foreman_queue_watch_test.go:66-67`, `foreman_queue_statement_test.go:36`, and `backlog_json_disclosure_mirror_test.go:26`; the three catalog tests name the id in comments only | 0 / 0 / 0 |
| PV-87 | `grep -rlE 'EnvMoaiKanban(ID\|LeadAddr\|LeadName\|SettingsInjected\|Backend\|Card)\b\|MOAI_KANBAN_(ID\|LEAD_ADDR\|LEAD_NAME\|SETTINGS_INJECTED\|BACKEND\|CARD)\b' internal cmd --include='*.go'` (each count by a separate `wc -l`, tests by `grep -c '_test.go'`); the constants-only variant; the literals-only variant | union 92 files, 61 test, 31 production; constants only 85 files, 56 test, 29 production; literals only 27 files; so seven files (two production, five test) spell only the literal | 0 |
| PV-88 | the cross-artifact ordering check by hand (table below): every ordering obligation in the acceptance surface against the milestone order of plan.md §D | every obligation is on the correct side; the CN-4 verb prints no `CONFLICT` | 0 |
| PV-89 | stale-label sweep over the seven artifacts (commands and results below the table) | swept all seven artifacts and the probe directory for the old counts (24 requirements, 95/34 files), the old M5a/M5b/M6 deletion lists (ten board files, 13 tests, the four-constant budget), and the stale milestone flips (RED-N2 at M7, AC-018 flips at M7 and M8); results in the list below the table; the historical evidence rows PV-24, PV-44, PV-46b, PV-47, PV-51, PV-70 keep their original figures and carry a superseded note where a figure changed | 0 |
| PV-90 | green-path observations of the acceptance commands that name deletions and renames, run on the scratch tree the committed runner leaves after stage M11 (`replay-final/tree`; the paths read `internal/factory` there): AC-012's identifier grep, symbol grep, board file-name `find`, kept-file `find`, and substrate consumers; AC-018's file-name `find`, old-import grep, and package directory; AC-019's panel string and legacy-route file; AC-020's renamed rule paths, old rule path, catalog counts, archive-list count, and skill directories | listing below the table: the board file names are absent (0 lines), the ten kept files are present (10 lines), the substrate consumers carry the re-homed names (`backlog_store.go` 2, `integration_lock_mutation.go` 2, `slot_lease.go` 3), `role.go` keeps `RoleLeader`, no file name carries the word, the old import path and `internal/kanban` are gone, `Chain session board` is gone from `screens.templ` and `legacy_routes.go` exists, the three renamed rule files exist and the old one does not, the catalog has 0 old and 2 new entries, the archive list carries the old id once, and the skill directory is renamed; two greps are NOT empty on this tree because the probe rewords no comment: AC-012's identifier grep still lists 7 files and its symbol grep 2 files, every match a comment (26 comment lines in `envkeys.go`, `factory_launch_helpers.go`, `factory.go`, `factory_settings.go`, `session_start_factory.go`, `factory/role.go`, `factory/bootstrap.go`, plus `state_lock.go:22` and `integration_lock_mutation.go:24`), which the plan assigns to M5a, M5b, and M6 | 0 |

### PV-74 — per-stage result and verbatim runner output

Stage table (read from the runner output below by `build_progress.py`-style parsing: a cell says `clean` only for an empty block with exit 0; the windows vet column says `baseline failure only` where the block equals the three baseline lines of PV-73 and nothing else):

| Stage | host build | windows build | host vet | windows vet | runner note |
|---|---|---|---|---|---|
| M0 | clean | clean | clean | baseline failure only | no Go change (printed by the runner) |
| M1 | clean | clean | clean | baseline failure only | no Go change (printed by the runner) |
| M2 | clean | clean | clean | baseline failure only |  |
| M3 | clean | clean | clean | baseline failure only |  |
| M4 | clean | clean | clean | baseline failure only |  |
| M5a | clean | clean | clean | baseline failure only |  |
| M5b | clean | clean | clean | baseline failure only |  |
| M6a | clean | clean | clean | baseline failure only |  |
| M6b | clean | clean | clean | baseline failure only |  |
| M7 | clean | clean | clean | baseline failure only | identifiers renamed: 515 |
| M8 | clean | clean | clean | baseline failure only | package clauses renamed: 156, qualifiers renamed: 1830, shadowing locals renamed: 4 |
| M9 | clean | clean | clean | baseline failure only | web identifiers renamed: 91 (14 distinct) |
| M10 | clean | clean | clean | baseline failure only |  |
| M11 | clean | clean | clean | baseline failure only | no Go change (printed by the runner) |

The windows vet block of every stage is identical to the baseline of PV-73 (a failure in `internal/cli/worktree`, a package this SPEC does not touch); no block names a package this SPEC changes. Stages M0, M1, and M11 print no Go change by design. The runner's own output, verbatim:

```
##### stage M0
(no Go change: no code, baseline only)
=== [M0] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M0] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M0] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M0] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M1
(no Go change: adds test files only; the additions are not authored by this probe)
=== [M1] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M1] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M1] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M1] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M2
=== [M2] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M2] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M2] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M2] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M3
=== [M3] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M3] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M3] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M3] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M4
=== [M4] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M4] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M4] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M4] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M5a
=== [M5a] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M5a] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M5a] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M5a] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M5b
=== [M5b] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M5b] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M5b] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M5b] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M6a
=== [M6a] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M6a] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M6a] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M6a] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M6b
=== [M6b] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M6b] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M6b] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M6b] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M7
identifiers renamed: 515
=== [M7] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M7] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M7] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M7] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M8
package clauses renamed: 156, qualifiers renamed: 1830, shadowing locals renamed: 4
=== [M8] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M8] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M8] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M8] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M9
web identifiers renamed: 91 (14 distinct)
templ generate exit 0
(✓) Post-generation event received, processing... [ updates=0 needsRestart=true needsBrowserReload=true ]
(✓) Post-generation event received, processing... [ updates=1 needsRestart=false needsBrowserReload=false ]
(✓) Complete [ updates=1 duration=111.639958ms ]
=== [M9] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M9] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M9] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M9] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M10
=== [M10] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M10] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M10] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M10] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
##### stage M11
(no Go change: documentation only)
=== [M11] go build -gcflags=-e ./...  (darwin/host)

--- exit 0
=== [M11] GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...

--- exit 0
=== [M11] go vet ./...  (darwin/host; typechecks every test file)

--- exit 0
=== [M11] GOOS=windows GOARCH=amd64 go vet ./...
# github.com/modu-ai/moai-adk/internal/cli/worktree
# [github.com/modu-ai/moai-adk/internal/cli/worktree]
vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs
--- exit 1
runner-exit=0
```

### PV-75 — 25 compile errors when the ten files are deleted without the re-home

Host build:

```
# github.com/modu-ai/moai-adk/internal/kanban
internal/kanban/backlog_store.go:1254:40: undefined: BoardLock
internal/kanban/backlog_store.go:1259:29: undefined: boardLockWaitBudget
internal/kanban/backlog_store.go:1261:16: undefined: acquireBoardLockImpl
internal/kanban/backlog_store.go:1263:12: undefined: BoardLock
internal/kanban/backlog_store.go:1265:7: undefined: IsBoardLockHeld
internal/kanban/backlog_store.go:1272:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation.go:95:51: undefined: boardLockImpl
internal/kanban/integration_lock_mutation.go:97:29: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:99:16: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:103:7: undefined: IsBoardLockHeld
internal/kanban/integration_lock_mutation.go:119:26: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:126:73: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:128:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation_unix.go:18:53: undefined: ClearStaleReport
internal/kanban/integration_lock_mutation_unix.go:19:10: undefined: ClearStaleReport
internal/kanban/slot_lease.go:463:49: undefined: boardLockImpl
internal/kanban/slot_lease.go:465:29: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:467:16: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:471:7: undefined: IsBoardLockHeld
internal/kanban/slot_lease.go:477:26: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:482:67: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:484:14: undefined: boardLockRetryWait
internal/kanban/slot_lease_mutation_unix.go:9:51: undefined: ClearStaleReport
internal/kanban/slot_lease_mutation_unix.go:10:10: undefined: ClearStaleReport
internal/kanban/role.go:86:23: undefined: BoardDir
```

`GOOS=windows GOARCH=amd64` build:

```
# github.com/modu-ai/moai-adk/internal/kanban
internal/kanban/backlog_store.go:1254:40: undefined: BoardLock
internal/kanban/backlog_store.go:1259:29: undefined: boardLockWaitBudget
internal/kanban/backlog_store.go:1261:16: undefined: acquireBoardLockImpl
internal/kanban/backlog_store.go:1263:12: undefined: BoardLock
internal/kanban/backlog_store.go:1265:7: undefined: IsBoardLockHeld
internal/kanban/backlog_store.go:1272:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation.go:95:51: undefined: boardLockImpl
internal/kanban/integration_lock_mutation.go:97:29: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:99:16: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:103:7: undefined: IsBoardLockHeld
internal/kanban/integration_lock_mutation.go:119:26: undefined: acquireBoardLockImpl
internal/kanban/integration_lock_mutation.go:126:73: undefined: boardLockWaitBudget
internal/kanban/integration_lock_mutation.go:128:14: undefined: boardLockRetryWait
internal/kanban/integration_lock_mutation_windows.go:24:56: undefined: ClearStaleReport
internal/kanban/integration_lock_mutation_windows.go:25:9: undefined: clearStaleLockAtPath
internal/kanban/slot_lease.go:463:49: undefined: boardLockImpl
internal/kanban/slot_lease.go:465:29: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:467:16: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:471:7: undefined: IsBoardLockHeld
internal/kanban/slot_lease.go:477:26: undefined: acquireBoardLockImpl
internal/kanban/slot_lease.go:482:67: undefined: boardLockWaitBudget
internal/kanban/slot_lease.go:484:14: undefined: boardLockRetryWait
internal/kanban/slot_lease_mutation_windows.go:14:54: undefined: ClearStaleReport
internal/kanban/slot_lease_mutation_windows.go:15:9: undefined: clearStaleLockAtPath
internal/kanban/role.go:86:23: undefined: BoardDir
```

### PV-76 — 8 compile errors when `langEnglish` and `operatorLang` are not moved first

```
# github.com/modu-ai/moai-adk/internal/hook
internal/hook/session_stale_run.go:83:2: undefined: langEnglish
internal/hook/session_stale_run.go:136:25: undefined: langEnglish
internal/hook/session_start_factory_i18n.go:72:2: undefined: langEnglish
internal/hook/session_start_factory_i18n.go:258:24: undefined: langEnglish
internal/hook/factory_messages.go:69:78: undefined: langEnglish
internal/hook/session_start.go:511:91: undefined: langEnglish
internal/hook/session_start.go:523:74: undefined: operatorLang
internal/hook/session_start.go:542:52: undefined: operatorLang
```

### PV-78 — test-helper controls

`internal/hook` without `session_start_env_helper_test.go` (`go vet ./internal/hook/`, first errors):

```
# github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
internal/hook/session_start_additional_context_test.go:147:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_provider_test.go:24:5: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:32:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:46:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:89:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:103:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:120:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:157:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:200:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:244:2: undefined: clearKanbanEnv
internal/hook/session_start_factory_test.go:244:2: too many errors
```

`internal/kanban` without `test_helpers_test.go` and `test_helpers_windows_test.go`, host test compile:

```
# github.com/modu-ai/moai-adk/internal/kanban [github.com/modu-ai/moai-adk/internal/kanban.test]
internal/kanban/f3_f4_probe_test.go:18:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:22:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:23:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:24:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:29:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:30:21: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:63:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:67:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:68:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:69:2: undefined: runGitAt
internal/kanban/integration_lock_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_cross_test.go:311:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_rotation_test.go:113:10: undefined: deadPID
internal/kanban/slot_lease_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:109:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:139:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:162:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:196:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:224:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:288:5: undefined: runtimeIsWindows
```

the same on `GOOS=windows` (the extra error is `deadPIDWin`):

```
# github.com/modu-ai/moai-adk/internal/kanban [github.com/modu-ai/moai-adk/internal/kanban.test]
internal/kanban/f3_f4_probe_test.go:18:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:22:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:23:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:24:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:29:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:30:21: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:63:5: undefined: runtimeIsWindows
internal/kanban/f3_f4_probe_test.go:67:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:68:2: undefined: runGitAt
internal/kanban/f3_f4_probe_test.go:69:2: undefined: runGitAt
internal/kanban/integration_lock_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_cross_test.go:311:5: undefined: runtimeIsWindows
internal/kanban/integration_lock_mutation_windows_test.go:52:41: undefined: deadPIDWin
internal/kanban/integration_lock_rotation_test.go:113:10: undefined: deadPID
internal/kanban/slot_lease_cross_test.go:174:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:109:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:139:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:162:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:196:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:224:5: undefined: runtimeIsWindows
internal/kanban/status_read_test.go:288:5: undefined: runtimeIsWindows
```

### PV-80 — deleting the whole file `role.go` on the M6b tree

```
# github.com/modu-ai/moai-adk/internal/kanban
internal/kanban/bootstrap.go:89:9: undefined: RoleLeader
internal/kanban/bootstrap.go:97:9: undefined: RoleLeader
internal/kanban/bootstrap.go:119:14: undefined: RoleLeader
internal/kanban/bootstrap.go:123:23: undefined: RoleLeader
internal/kanban/record.go:151:13: undefined: RoleLeader
internal/kanban/record.go:151:35: undefined: RoleLane
internal/kanban/todo_owner_label.go:35:7: undefined: IsLegacyLeaderSpelling
internal/kanban/todo_owner_label.go:38:10: undefined: RoleLeader
internal/kanban/todo_owner_label.go:38:49: undefined: legacyLeaderSpelling
internal/kanban/todo_owner_label.go:47:10: undefined: RoleLane
```

### PV-83 — the entry-parse probe, verbatim

```
=== RUN   TestZZT1399EntryProbe
CCPARSE args=["-f" "--name" "lane-2"] factoryEnabled=true rest=["--name" "lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["-f" "-n" "lane-2"] factoryEnabled=true rest=["-n" "lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["--factory" "--name=lane-2"] factoryEnabled=true rest=["--name=lane-2"] laneLabel="lane-2" isLane=true branch=2
CCPARSE args=["-f" "--name" "leader-r7"] factoryEnabled=true rest=["--name" "leader-r7"] laneLabel="" isLane=false branch=1
CCPARSE args=["--lane"] factoryEnabled=false rest=["--lane"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "lane-2"] factoryEnabled=false rest=["--lane" "lane-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane=lane-2"] factoryEnabled=false rest=["--lane=lane-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "3"] factoryEnabled=false rest=["--lane" "3"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "leader-2"] factoryEnabled=false rest=["--lane" "leader-2"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "-f"] factoryEnabled=true rest=["--lane"] laneLabel="" isLane=false branch=1
CCPARSE args=["-l"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "lane-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "3"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l=lane-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-l" "-f"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "-k"] parseErr="--leader requires a leader label"
CCPARSE args=["-l" "--name" "lane-2"] parseErr="--leader requires a leader label"
CCPARSE args=["-f" "-l"] parseErr="--leader requires a leader label"
CCPARSE args=["--lane" "-k"] factoryEnabled=false rest=["--lane"] laneLabel="" isLane=false branch=0
CCPARSE args=["--lane" "--name" "lane-2"] factoryEnabled=false rest=["--lane" "--name" "lane-2"] laneLabel="lane-2" isLane=true branch=0
CCPARSE args=["--leader" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["--leader=leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-f" "--leader" "leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CCPARSE args=["-f" "--leader=leader-2"] parseErr="--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target"
CODEX args=["--lane"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane=lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "3"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "leader-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["--lane" "-f"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--lane" "-k"] err= exitCode=1 stderr="KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead\n" stdoutLen=0
CODEX args=["--lane" "--name" "lane-2"] err= exitCode=1 stderr="unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app\n" stdoutLen=0
CODEX args=["-f" "-l"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
CODEX args=["--leader" "leader-2"] err=--leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target exitCode=-1 stderr="" stdoutLen=0
CODEX args=["-f" "--leader" "leader-2"] err= exitCode=1 stderr="FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader\n" stdoutLen=0
--- PASS: TestZZT1399EntryProbe (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.196s
exit 0
```

### PV-85 — RED-now re-runs of the changed criteria, verbatim

Launcher cells (tree binary):

```
RED-1 moai cc -l: exit=1 stdout-bytes=0 stderr=ERROR --Leader requires a leader label.
RED-2 moai codex -l: exit=1 stdout-bytes=0 stderr=ERROR --Leader requires a leader label.
RED-3 moai cc -l lane-2: exit=1 stdout-bytes=0 stderr=ERROR --Leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target.
RED-5 moai codex -f 3: exit=1 stdout-bytes=0 stderr=FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader
RED-6 moai cc -f 3: exit=1 stdout-bytes=0 stderr=ERROR -F/--Factory takes no argument (the factory leader), the role token -f lane, which joins this session to a running factory as the next free lane, or a lane label (e.g. -f lane-2) that launches exactly that one lane, got "3".
RED-10 moai cc -f -l: exit=1 stdout-bytes=0 stderr=ERROR --Leader requires a leader label.
RED-K2 moai codex -k: exit=1 stdout-bytes=0 stderr=KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead
```

Grep and `find` cells (script output: exit code and stdout line count; the two PCRE cells RED-N1 and RED-D1 printed `exit=0` with 189 and 157 lines in the interactive shell):

```
RED-K3: exit=0 stdout-lines=15 
RED-K4: exit=0 stdout-lines=27 
RED-K4-find: exit=0 stdout-lines=18 
AC-012-positive-control-find(today): exit=0 stdout-lines=4 
RED-N1: exit=2 stdout-lines=0 
RED-N2: exit=0 stdout-lines=22 
RED-N3: exit=0 stdout-lines=179 
RED-C1: exit=1 stdout-lines=3 
RED-C2: exit=0 stdout-lines=2 internal/template/catalog.yaml:41:            - name: moai-kanban-foreman | internal/template/catalog.yaml:43:              path: templates/.claude/skills/moai-kanban-foreman/
RED-C3: exit=0 stdout-lines=3 39965 | 18909 | 26030
RED-C4: exit=0 stdout-lines=3 .claude/rules/moai/workflow/kanban-dispatch-detail.md:1 | .claude/rules/moai/workflow/kanban-dispatch-mechanics.md:1 | .claude/rules/moai/workflow/cross-session-messaging-detail.md:1
RED-D1: exit=2 stdout-lines=0 
RED-D2: exit=0 stdout-lines=1 4
RED-D3: exit=1 stdout-lines=1 0
RED-D5: exit=0 stdout-lines=46 
RED-D5-old-pattern: exit=0 stdout-lines=46 
RED-D8: exit=0 stdout-lines=16 
RED-D7: exit=0 stdout-lines=19 
RED-S1: exit=0 stdout-lines=16 
docs-line-cites: exit=0 stdout-lines=2 38:| `-l, --lead <name>` | With `-f lane` / `-f lane-<n>`: which leader session the record-absence verification targets (default `leader`; the former spelling `lead` is refused). When the run's record is missing or retired while a live leader exists, the join verifies that leader (pid + process-start) and restores its run | | 80:Grow a run one lane at a time with `moai cc -f lane` (auto-join the next free number) or `moai cc -f lane-<n>` (that number exactly). Both forms already name the lane, so passing `--name`/`-n` alongside them is an error. An explicitly-picked number that collides with a live lane bumps to the next free number. A number is otherwise skipped only while a live session holds it — a dead lane's claim no longer blocks its number (an explicit pick reuses it right away), but `-f lane` auto-assignment always takes one past the highest live number and never backfills a gap. Lane ownership is recorded in `~/.moai/db/<project-key>/factory/factory.db` — or, when the launch directory is a temporary one (no absolute `MOAI_HOME` override), project-local under `<base>/.moai/db/<project-key>/factory/`, the same exception the backlog queue follows; a legacy `.moai/state/factory/workers.json` is imported once and retained only as rollback evidence. A lane runs up to 10 concurrent `Agent()` subagents, and write-capable spawns are isolated in their own worktree. Never bring every lane up at once — start the first, confirm it is actually producing output, then activate the rest. Cards are never split across lanes. `-k` still drives the three-role kanban chain; one launch takes one entry token, so `-k` with `-f` is an error. CG is retired; use `moai migrate cg` to preview explicit migration choices. A factory run now records the process identity of the session that owns it, so a run whose leader has died is retired automatically the next time a lane joins instead of leaving that join stuck on `AMBIGUOUS_FACTORY`. The inverse is covered too: when a lane joins and the run's record is missing or retired while a live leader session exists, the join verifies that leader (pid plus process-start fingerprint, targeted with `-l/--lead`, default `leader`), restores its run record, and joins anyway — two or more verified leaders fail closed naming each candidate. `moai factory runs` lists every run with its owner's liveness, and `moai factory runs --retire <run-id>` retires one by hand — refused unless that run's owner is actually dead.
```

### PV-90 — green-path observations on the modeled final tree, verbatim

```
AC-012 chain/companion identifiers (non-test): exit=0 stdout-lines=7
AC-012 board API and carrier symbols: exit=0 stdout-lines=2
AC-012 board file names: exit=0 stdout-lines=0
AC-012 kept files: exit=0 stdout-lines=10
AC-012 substrate consumers: exit=0 stdout-lines=3 | internal/factory/backlog_store.go:2 | internal/factory/integration_lock_mutation.go:2 | internal/factory/slot_lease.go:3
AC-012 role.go keeps RoleLeader: exit=0 stdout-lines=1 | 2
AC-018 file names: exit=0 stdout-lines=0
AC-018 old import path: exit=1 stdout-lines=0
AC-018 package directory: exit=1 stdout-lines=2 | ls: internal/kanban: No such file or directory | internal/factory
AC-019 panel string: exit=1 stdout-lines=1 | 0
AC-019 legacy route file: exit=0 stdout-lines=1 | internal/web/legacy_routes.go
AC-020 renamed rules (template): exit=0 stdout-lines=3
AC-020 old rule path: exit=1 stdout-lines=1 | ls: internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md: No such file or directory
AC-020 catalog old/new: exit=0 stdout-lines=2 | 0 | 2
AC-020 archive list: exit=0 stdout-lines=1 | 1
AC-020 skill dir: exit=1 stdout-lines=2 | ls: internal/template/templates/.claude/skills/moai-kanban-foreman: No such file or directory | internal/template/templates/.claude/skills/moai-factory-foreman
```

### PV-88 — cross-artifact ordering check, by hand

Every clause that orders work in the acceptance surface (the Definition of Done, and each criterion's Given, When, and Then), read against the order of plan.md §D and, where the clause concerns a deletion or a rename, against the compile proof of PV-74. The CN-4 verb (PV-84) flags candidate records mechanically; this table is the by-hand reading of each obligation, including the ones the verb cannot see (a clause spread across two sentences, or a milestone bound by prose).

| # | Obligation (where it is stated) | Plan binding | Satisfied |
|---|---------------------------------|--------------|-----------|
| 1 | The AC-004 golden is committed alone BEFORE any change (AC-004 Given; Definition of Done 6) | M1 commits the golden alone; M2 is the first milestone that changes a parse | yes |
| 2 | AC-015's eight mutant reds (a)-(h) and AC-017's mutant red are recorded BEFORE M5a starts (AC-015, AC-017, Definition of Done 2) | M1 authors the net and the tolerance test and records the reds; M5a is the first removal | yes |
| 3 | Mutants (g) and (h) are re-observed red AFTER the M2 re-pin of the lane tests (AC-015, Definition of Done 2) | M2 states the re-observation as its closing step | yes |
| 4 | RED-11, RED-12, RED-14 are re-observed on the milestone tree BEFORE the cc/glm parse changes; RED-13 and the codex rows of RED-14 BEFORE the codex parse changes (AC-006, AC-003, AC-005, Definition of Done 1) | M2 and M3 each open with that observation as their first act (the M3 first act was added in v0.8.0) | yes |
| 5 | The update fixture of AC-020 is observed RED on the pre-rename template BEFORE the rename commit (AC-020, Definition of Done 3) | M10 commit 1 is the fixture test alone, commit 2 the rename; the commit graph is the witness | yes |
| 6 | `moai constitution validate` runs BEFORE the first constitution edit and AFTER each of the four (AC-022) | M10 commit 2 | yes |
| 7 | The three constants `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL` are deleted only AFTER their last readers (AC-012, AC-013, Definition of Done 7) | M5a removes the launcher readers, M5b removes the hook readers and then the constants; compile proof PV-74 stages M5a and M5b | yes |
| 8 | `leadFlagShort` is deleted only AFTER the Codex parse stops reading it (Definition of Done 7) | M3 deletes it with its last reader; PV-77 is the negative control for M2 | yes |
| 9 | The shared lock substrate is re-homed BEFORE the board is deleted (AC-012, Definition of Done 7) | M6 step 1 (stage M6a), then step 2 (stage M6b); PV-75 is the negative control | yes |
| 10 | `langEnglish` and `operatorLang` move BEFORE the kanban notice file is deleted; the test helpers move BEFORE the test files that define them (AC-012, AC-014) | M5b step 1; M5b and M6 helper moves; PV-76 and PV-78 are the negative controls | yes |
| 11 | Renames come AFTER removals so nothing is renamed and then deleted (design.md §1) | M7 to M10 follow M6; `internal/web` is renamed at M9, after the Go-wide M7 and M8, and its view model is deleted in the same stage; no symbol renamed at M7 is deleted later (the chain view model is not renamed at M7) | yes |
| 12 | The package path is `internal/kanban` through M7 and `internal/factory` from M8 on (AC-012, AC-016, AC-020) | M8 renames the package; M9 imports it | yes |
| 13 | The web sources carry the word until M9 lands, the template mirror paths until M10 lands, so AC-018's full command closes at M10 (AC-018) | M7, M8, M9, M10 each clear their part; AC-018 is bound to M10 | yes |
| 14 | AC-013's grep half is read AFTER the hook readers are gone (AC-013) | M5a test half, M5b grep half | yes |
| 15 | The foreman path-pinned tests are re-pointed in the SAME commit as the skill rename (AC-020) | M10 commit 2 | yes |
| 16 | The net runs at every milestone from M2 through M9 and the windows build at every Go milestone (AC-015, AC-018) | the verification gate of plan.md §D | yes |
| 17 | The four locales and each page land in ONE commit; AC-023 and AC-025 close at sync-audit (AC-023, AC-025) | M11 | yes |
| 18 | M2, M3, and M4 merge as one integration unit; M6 and M10 hold two commits in the stated order (spec.md §E, plan.md §D) | plan.md §D integration units | yes |

Sweep results:

- old counts and labels: `grep -rnE '24 requirements\|24 of 25\|REQ-001\.\.REQ-024\|M11 scope\|95 files\|34 production\|M7 flips RED-N2\|flips at M7, M8\|28 references' *.md \| grep -v '^progress.md:\(66\\|94\\|90\\|89\\|125\)' \| cut -c1-80` -> progress.md:119:| PV-70 | `moai spec lint SPEC-LAUNCHER-ENTRY-FLAGS-001 --strict / spec.md:31:- 2026-10-02 (v0.7.0): Plan-audit iteration 1 (FAIL, 0.72; findings D
- old deletion lists: `grep -rnE '10 board files and 13 tests\|13 board tests\|four-constant budget\|board family \(10 files\)' *.md \| cut -c1-80` -> no output (exit 0)
- requirement count lines: `grep -n 'this SPEC carries' acceptance.md spec.md \| cut -c1-150` -> acceptance.md:3:All criteria are mechanically verifiable. Tier L ceiling: 25 requirements and 25 criteria; this SPEC carries 25 requirements and 25 cr / spec.md:156:Full Given-When-Then scenarios, the RED-now evidence ledger, and the Definition of Done live in `acceptance.md`. Tier L ceiling: 25 requir

Reading of the sweep: the two hits of the first command are the v0.7.0 history line of `spec.md` (line 31) and the historical row PV-70, which record the v0.7.0 counts and stay as history; the second command is empty (the old M5a/M5b/M6 deletion lists are gone); the third shows both ceiling statements at 25 requirements and 25 criteria.

Re-run of the RED-now cells of the criteria this revision changed, on HEAD `b9242da00ce489c4f26efb5f6d447ccafef08c54` (this run): RED-1, RED-3, RED-10 (`cc -l`, `cc -l lane-2`, `cc -f -l`: exit 1, the same stderr lines as the ledger), RED-2, RED-5 (`codex -l` exit 1 `--Leader requires a leader label.`; `codex -f 3` exit 1 with the `FACTORY_MODE_UNSUPPORTED_BACKEND` line), RED-6, RED-4 (`--- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (2.92s)`), RED-K3 (the same 15 files), RED-K4 (24), RED-8 (4), RED-8b (`envkeys.go`, `codex_launcher.go`), RED-N1 (189), RED-N2 (22), RED-N3 (179), RED-C1 (three `No such file`, exit 1), RED-C2 (`catalog.yaml:41,43`), RED-C3 (39965, 18909, 26030 bytes), RED-D1 (157), RED-D2 (4), RED-D3 (0, exit 1), RED-D4 (four zeros, exit 1), RED-D5 (46), RED-D6 (four `No such file`, exit 1), RED-D7 (19), RED-C4 (three counts of 1, exit 0), RED-11 to RED-13 (PV-60), RED-S1 (the same 16 files), BASE-4 (windows build exit 0), BASE-5 (2). The AC-015 cli selector with both settings-test names swept 4 tests today (`TestCCFactoryEntryRecordsFailOpenRunMetadata`, `TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestPrepareKanbanSettingsWritesTransientFile`) and sweeps 8 once the four net tests of M1 exist. The cells of AC-009, AC-010, AC-011, AC-014, AC-019, AC-021 were not changed in this revision and were not re-run.

Reading of PV-46/PV-47 (and PV-70, which supersedes them for v0.7.0): REQ-to-AC coverage is established by PV-47, not by the lint. Counts at authoring: 24 requirements (Tier L ceiling 25) and 25 criteria (Tier L ceiling 25 — at the ceiling; no further criterion without merging or splitting the SPEC).

Reading of PV-84 (supersedes PV-70 for v0.8.0): REQ-to-AC coverage is again established by the hand count, not by the lint; the lint reports `No findings` and the collection verb reports 25 definitions with no `UNCOVERED` line, and the verb's `mapped` is generous (it counts any line that names both an AC and the REQ), so the per-requirement `Verifies` counts of PV-84 are the evidence. Counts at authoring: 25 requirements (Tier L ceiling 25, at the ceiling after the REQ-019 split) and 25 criteria (ceiling 25, at the ceiling; the new `-l` and `--leader` shapes were folded into AC-003 and AC-005).

Gaps: see `plan.md` §G (sixteen plan-level items) and `research.md` §R14 (twenty-one items) — none is asserted as fact in `spec.md`. Gaps recorded by the v0.8.0 revision specifically: (1) the compile proof models each milestone's DELETIONS and the edits that keep the build, not the added code, and prints M0, M1, and M11 as stages with no Go change; (2) the semantic re-pin of every test the probe retargets (the removed constants are mapped to surviving factory markers as a compile model) and the per-function classification of the cases the probe deleted are the run phase's; (3) the five lock-coupled tests of `internal/kanban` are modeled by deletion although two of them (the errno classification and the cross-process exclusion) are to be re-pointed, and whether retained tests already cover the substrate was not measured; (4) the runtime effects of the M10 renames (embedded template paths, catalog hash, foreman path pins) are not compile effects and are gated by AC-018 to AC-022, not by the probe; (5) PV-77 and PV-79 quote two redeclaration errors and one file-name collision from an interim replay of the stage data before the corrections entered it, and PV-82 quotes the view-model reader errors from the M9 prototype tree, so those rows are attributable to the same edit set but not to the committed stage data end to end; (6) the committed runner's own run (PV-74) began before the header comment of `probe.go` was reworded; no code line changed, and the file was deliberately left as run (it is not gofmt-aligned in its `var` block) so that the recorded run stays attributable to it; (7) the plan-audit reports `.moai/reports/t1399/plan-audit-iter1.md` and `plan-audit-iter2.md` are local gitignored files; both were read and neither was edited; (8) the worktree guard refused several compound shell commands and the runs were re-issued as plain commands or scripts written with the Write tool, with no measurement replaced by source reading; (9) the Overview attention row (Q24) and the file `role.go` (verdict 20) are findings of the compile proof that the verdicts did not state — they are written to the smallest-footprint reading and returned for acknowledgement.

## §J Plan-audit iteration 3 outcome and carried debt

Appended after plan-audit iteration 3 (the last audit allowed). This section is append-only: no earlier line of this file was edited, and the five plan-artifact files (`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`) were not touched, because the audit verdict is bound to their hash. Where this section contradicts an earlier line (the §E.1 prose says iteration 3 "has not run"; Gap (6) of §G says `probe.go` was left unformatted), this section is the later record and the earlier line is left as written history.

### J.1 Audit iterations

| Iteration | Verdict | Score | Audited at | Report (local, gitignored) |
|-----------|---------|-------|------------|----------------------------|
| 1 | FAIL | 0.72 | `b9242da00ce489c4f26efb5f6d447ccafef08c54` | `.moai/reports/t1399/plan-audit-iter1.md` |
| 2 | FAIL | 0.79 | `5445e296caa48e6ab9821afe808eac2e7385e897` | `.moai/reports/t1399/plan-audit-iter2.md` |
| 3 | PASS-WITH-DEBT | 0.88 (Tier L threshold 0.85) | `7e1ae0d808b180f266c71761290ab6ed96ba335f` | `.moai/reports/t1399/plan-audit-iter3.md` |

Iteration 3 reports MUST-FIX 0, SHOULD-FIX 8, ADVISORY 6 (14 findings, all Class optional). The audit-ready signal of §E.1 (`plan_status: audit-ready`, `plan_complete_at: 2026-10-02`) was not edited: it carries no `audited_sha` field, its `audited_head` field names the iteration 2 commit, and editing an existing line would break the append-only constraint of this section. The iteration 3 audited commit is recorded here instead.

### J.2 What the auditor re-ran, as the report states it

Source: `.moai/reports/t1399/plan-audit-iter3.md` § 4.2 and § 4.3 (read in this run; the report is the auditor's measurement, not mine).

- The committed compile-proof runner was re-run over all 14 stages (M0..M11 with M6a and M6b), four checks per stage. Every stage printed empty build and vet blocks with exit 0 on the host and for windows, except the windows vet block, which printed exactly the pre-existing three-line baseline (`internal/cli/worktree/sweep_test.go:1687`, `undefined: parseLsofCWDs`). The output equals the recorded PV-74 block except the `templ` generator's own timing lines at M9.
- The auditor made the proof go red on purpose: one line (`internal/kanban/state_lock_unix.go`) appended to a scratch copy of `M6b.rm`, so a re-homed lock file is deleted with no replacement. The host build failed with five `undefined: acquireStateLockImpl` errors; the windows build stayed clean (the windows file survived).
- A pristine-tree overlay control with the ten v0.7.0 "board family" files mapped to empty printed 25 `undefined:` lines, equal to PV-75.

### J.3 Carried debt

Owners are as the report assigns them. Where the report names no owner the row says `unassigned`; this section does not invent one. None of the plan-text fixes below can be applied now, because the five hash-subject files are frozen; every such fix is carried as run-phase or Kickoff work.

| ID | Class | One-line description | Owner (as the report gives it) | The report's minimal fix |
|----|-------|----------------------|--------------------------------|--------------------------|
| D-A1 | SHOULD-FIX (major) | `probe/probe.go` was not gofmt-clean (blank line missing before the M9 and M10 banners), and the CI format gate runs over tracked `.go` files | the orchestrator, before the card branch is merged or pushed | `gofmt -w probe/probe.go`, note it in progress.md Gap (6), re-run `gofmt -l` and the probe's M0 stage — EXECUTED in this run except the M0 re-run, see J.4 |
| D-A2 | SHOULD-FIX (major) | AC-018's exact-four grep still returns 123 files / 502 lines on the modeled final tree (22 files cite `SPEC-KANBAN-*`); the comment and citation sweep is unscheduled | M7 (comments), M10 (final sweep) | add a comment-and-citation sweep step to M7 with the measured count and a stated rewrite rule for completed-SPEC citations, or narrow REQ-017 and AC-018 to non-comment tokens |
| D-A3 | SHOULD-FIX (major) | run-time package-path strings in tests are ungated: `TestBacklogJSONLiteralStaysSeamScoped` fails on the modeled tree and `migrate_home_state_test.go:68,992,1059` carry `"./internal/kanban"` and are named in no artifact; Gap 13's claim is false | M8 | add the three `migrate_home_state_test.go` sites to design §4.7, add scoped lease-guarded test selectors to AC-018's command, correct Gap 13 |
| D-A4 | SHOULD-FIX (minor) | `-l -k` and `--lane -k`: REQ-002 and REQ-010 prescribe different one-line content with no stated precedence; AC-003's two rows go red at M5a unless re-pinned | M5a | state that `-k` is refused first and the line names both facts; add AC-003 to M5a's re-pin list |
| D-A5 | SHOULD-FIX (minor) | AC-003 covers `--name lane-2` but not `-n lane-2` or `--name=lane-2` beside `-l` or `--lane` (a shallow-mutant gap) | M2 | add the `-n` and `--name=` shapes and their `--lane` twins to AC-003's shape list and subtest count |
| D-A6 | SHOULD-FIX (minor) | AC-012 floors only the four wait-budget tests; the errno classification and cross-process exclusion tests could be deleted and AC-012 would still pass | M6 | add a scoped, lease-guarded selector over the retained consumer cross-process tests to AC-012's command, with a swept count |
| D-A7 | SHOULD-FIX (minor) | AC-017's M1-authored fixture may name `MOAI_KANBAN` and `MOAI_KANBAN_LABEL` as constants that M5b deletes; Gap 12 says the design reads none | M1 | state that M1-authored tests name those markers only as string literals; say so in Gap 12 |
| D-A8 | SHOULD-FIX (minor) | `graph-freshness.yml` runs `moai graph check` on every push to `develop`; the tracked codemaps cite `internal/kanban` 41 times, so the check is expected (inferred, not measured) to go red after the rename | M11 / sync | schedule `/moai codemaps` plus the stamp in the same batch as M8, or record that a red graph-freshness check is accepted until the debt card lands |
| D-A9 | ADVISORY | stale labels: `design.md:1` says v0.7.0; `design.md:96` and `research.md:146` print a closed gap as open; `design.md:47` and `research.md` §R4 say "two `-k` constants" where `plan.md:168` lists four; `decision-index.md:25` Q18 row maps to REQ-019 which REQ-025 now carries | unassigned (the report gives none) | the report names no fix beyond the label corrections it lists; not applicable now (frozen files, and `decision-index.md` is not among the files this task may touch) |
| D-A10 | ADVISORY | REQ-025 is printed between REQ-019 and REQ-020 (`spec.md:147`) | unassigned (the report gives none) | move it to the end of §B.4 or renumber |
| D-A11 | ADVISORY | AC-023's added alternative misses a page that writes the retired short as separate code spans ("`-l`, `--lead <name>`") | unassigned (the report gives none) | the report states the gap and gives no fix |
| D-A12 | ADVISORY | the proof and AC-018 cover darwin and windows only; linux-only files (`internal/discovery/leader_readers_linux.go`, `internal/session/proc_info_linux.go`) were measured clean only at the M11 tree | unassigned (the report's owner label is absent; its fix names M7) | add `GOOS=linux GOARCH=amd64 go build ./...` to the per-milestone gate at M7 (the marker-constant rename) |
| D-A13 | ADVISORY | Q20-Q22 and Q24 stay open and non-gating; REQ-018 states the Q24 reading normatively, so an answer of B at Kickoff changes REQ-018 and M9 | Kickoff | record the closures at Kickoff or sync |
| D-A14 | ADVISORY | REQ-019 still bundles several concerns (regression rows D12 and D45, partially resolved) | unassigned (the report gives none) | the report states the bundle and gives no fix |

Row count: 14 (D-A1..D-A8 SHOULD-FIX, D-A9..D-A14 ADVISORY). Owner column: nine rows carry an owner the report names (D-A1, D-A2, D-A3, D-A4, D-A5, D-A6, D-A7, D-A8, D-A13); five carry `unassigned` (D-A9, D-A10, D-A11, D-A12, D-A14), of which D-A12's fix names M7.

### J.4 Fix applied in this run

D-A1 only: `gofmt -w` on `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/probe.go`, run in this worktree at HEAD `7e1ae0d808b180f266c71761290ab6ed96ba335f` (this run, this tree). Observed:

- before: `gofmt -l .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe` printed `.moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/probe.go`;
- after: the same command printed nothing, exit 0;
- `go vet` over `probe/probe.go` and over `probe/entry_probe.go` (each a `//go:build ignore` file, vetted by path): exit 0 for both;
- the diff of the probe directory is exactly two inserted blank lines, one before the `// ---- M9: web console` banner and one before the `// ---- M10: rules, skills, catalog` banner; no code line changed. `entry_probe.go` was not listed by `gofmt -l` and was not changed; `entry_probe_test.go.txt` is not a Go file and was not touched.

This supersedes the "left as run" statement of Gap (6) in §G for the formatting only: the file's formatting changed by those two blank lines, its code did not, and the recorded PV-74 run was taken before them. The `var` block alignment that Gap (6) mentions needed no change (the formatter produced no diff there).

### J.5 Run-phase obligations (not requirement changes)

The following bind the run phase without changing a requirement or a criterion:

1. **AC-018 sweep sizing (D-A2).** At M1, size the exact-four-files grep sweep behind the audit's measured 123 files / 502 lines (41 files with a non-comment match; 22 files carrying `SPEC-KANBAN-*` citations) and record the measured edit list the plan already promises; the gofmt/format gate for tracked `.go` files (`make fmt-check`, which runs `gofmt -l` over `git ls-files '*.go'`) stays green in every milestone.
2. **Path-string tests (D-A3).** The run-time path-string tests the report names are added to the milestone that renames their path strings (M8): `TestBacklogJSONLiteralStaysSeamScoped` in `internal/factory`; the three `internal/cli` tests the auditor saw go red once the non-test path strings were renamed (`TestHomeStateVerifiedLiveGateCannotBypassOrReplay`, `TestHomeStateVerdictEvidenceValidator`, `TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases`), caused by `internal/cli/migrate_home_state_test.go:68,992,1059`; the fourth failure the auditor saw (`TestHomeStateValidationCommandWrappersAndHelperFailures`, `head="" err=exit status 128`) failed identically before the rename on a scratch tree without a repository, so it is an environment artifact of that scratch tree and not a SPEC obligation.

### J.6 Gaps of this section

- The M0 sanity re-run of the probe that D-A1's fix suggests was not executed in this run (the instruction scoped this run to formatting plus this section); the `go vet` exit 0 over both probe files is the check that was observed. The auditor's own full 14-stage re-run (J.2) was taken against the pre-format file and the diff is two blank lines, so the program is unchanged, but that is a reading of the diff, not a re-measurement.
- The report's counts (123 files, 502 lines, 41 codemaps citations, the four red tests) are the auditor's measurements; none was re-measured in this run.
- The plan-text fixes D-A2..D-A7, D-A9..D-A12 and D-A14 were not applied, by instruction (hash-bound files); `decision-index.md` (D-A9's Q18 row) was outside the two files this run was permitted to touch.


## §F Phase 4 Mode Selection

Decision (summary): serial. Reasoning and the two recorded deviations follow below.

Recorded by the lane orchestrator (lane-3) before the first run-phase delegation, 2026-10-02.

### Kickoff gate (plan to run): met in its operator form

- Plan-audit: iteration 3 PASS-WITH-DEBT, score 0.88, audited_sha 7e1ae0d808b180f266c71761290ab6ed96ba335f (iterations 0.72 FAIL, 0.79 FAIL; see section J). The five plan-artifact hash subjects (spec.md, plan.md, acceptance.md, design.md, research.md) are byte-identical between the audited commit and the run-entry commit 21f912ca7 (`git diff --quiet` over those five paths exits 0).
- Operator answers, given directly in the lane session through AskUserQuestion on 2026-10-02: (1) enter the run phase for the whole SPEC, M0 through M11; (2) the role.go reading is confirmed (delete the role-declaration carrier only, the file stays); (3) Q20, Q21, Q22 and Q24 proceed on their smallest-footprint readings; (4) progression mode: autonomous.
- The Kickoff gate is the operator's own answer here (kept in the lane session by the lane rule); it is not the autonomous audit-cross form, and no goal is armed by this record.

### Input parameters

- tier: L (13 integration units, M0..M11 with M5a, M5b, M6a, M6b).
- scope: hundreds of files (SPEC research: 12 whole kanban-only Go files, 515 identifier occurrences, 1,830 qualified references at the package rename, 25 to 50 doc and instruction files per surface, 24 test files naming the three marker constants).
- domain count: 7 or more (launcher Go, hook Go, web templ, docs-site four locales, README four locales, rules and skills and template mirrors, config).
- file language mix: Go, templ, Markdown, YAML, JSON, shell.
- concurrency benefit: LOW. The work is coding-heavy, strictly ordered, and each milestone builds on a compile-proved predecessor; one writer per working tree.
- Agent Teams prerequisites: not requested (no explicit --team).

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | not selected | not a trivial change |
| serial | selected | coding-heavy and strictly sequenced; one implementation worker per milestone or integration unit |
| fanout | not selected | only read-only fan-out is safe in one tree; used inside a milestone for read-only measurement if needed, never for writes |
| sweep | not selected | M7 to M9 are mechanical renames but are programmatic (probe/probe.go rules) and need per-milestone compile proof; not a dynamic workflow |
| agent-team | not selected | not explicitly requested |

Decision: serial

### Justification and boundary cases

Serial is the default for coding-heavy work (the coding-task parallelism caveat). Two deviations from plan.md section D's wording are recorded here so the audit trail shows them:

1. plan.md says the run phase is led by manager-lead (Tier L, CLAUDE.md section 4 item 7). manager-lead is not spawned: this session is a factory lane, whose standing spawn authority is depth-1 only (spawned agents are leaf workers and never spawn further agents), and manager-lead exists to spawn leaf workers of its own. The lane orchestrator therefore sequences the milestones itself, one worker at a time, exactly the serial envelope manager-lead would use; coordination duties plan.md assigns to the run-phase coordinator (the M7 to M9 hold on other lanes' Go-touching merges, the integration-window re-measure) are carried by the lane and requested from the factory leader.
2. The Status Transition Ownership Matrix names manager-develop for run-phase implementation. A manager-develop typed spawn is auto-isolated into its own L1 tree and its writes to the card worktree are refused (measured on card t1318, recorded in the lane memory), and a non-isolated spawn is pinned to the spawning session's tree, which here is the card worktree. Implementation workers are therefore general-purpose spawns carrying the full manager-develop role instructions in the prompt (TDD cycle, ownership boundaries, commit trailers, the Authored-By-Agent: manager-develop trailer on commits that carry the draft to in-progress transition). This is an ownership exception recorded here, not a change of owner.

Base: the card branch WT-launcher-entry-flags is based at a6d3e6fd4 (the tree every measured fact and RED-now cell is pinned to); local develop is 69 commits ahead with 79 changed files, of which only internal/kanban/classification.go and its test overlap this SPEC's edit areas (no launcher file). The run proceeds on the pinned base; develop is absorbed once, before the integration window, and every pinned measurement is re-taken on the absorbed tree at that point (merge-base discipline for scope claims).
