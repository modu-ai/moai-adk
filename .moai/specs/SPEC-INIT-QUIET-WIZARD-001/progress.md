# SPEC-INIT-QUIET-WIZARD-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-11

Plan 단계 산출물 작성 완료 (manager-spec, 카드 t583, Tier L): spec.md v0.1.4 (REQ 16개, GEARS 5패턴), plan.md, acceptance.md (인수 기준 17건 — AC-IQW-001~016, 007 은 a/b 분할), design.md, research.md, spec-compact.md. 워크트리 `.claude/worktrees/t583`, 브랜치 `WT-init-quiet-wizard`, HEAD `120436f58`.

plan 감사 1회차(`.moai/reports/t583/plan-audit.md`): FAIL 0.79(Tier L 기준 0.85 미달, must-pass 전부 통과). v0.1.2 에서 D1~D8 을 반영했다 — D1 은 리드 결정(셸 설정 단계 시접 + 스파이 실행 관측)으로 바꿨다. v0.1.3 에서 리드 조건 두 가지를 반영했다: 시접을 바꾸는 테스트의 병렬 금지와 `t.Cleanup` 원복을 AC-IQW-016 으로 올렸고, AC-IQW-003 과 완료 정의의 카드 범위 diff 를 리터럴 핀에서 흡수한 로컬 `develop` 과의 merge-base 기준으로 바꿨다.

plan 감사 2회차(`.moai/reports/t583/plan-audit-iter2.md`): FAIL 0.86(Tier L 기준 0.85 는 넘었으나 blocking 결함 D9·D10 이 남음, must-pass 전부 통과). v0.1.4 에서 D9~D16 을 반영했다 — D9 는 AC-IQW-005 대상 목록의 검증 시점 스윕, D10 은 AC-IQW-004 본문 보존 관측.

plan 감사 3회차(`.moai/reports/t583/plan-audit-iter3.md`, 마지막): FAIL 0.87(점수 추이 0.79 → 0.86 → 0.87, must-pass 전부 통과). D10~D16 해소, D9 부분 해소, D1~D8 회귀 없음. 남은 blocking 결함은 D17 하나다 — AC-IQW-005 의 스윕이 추가 파일 단위라, 테스트 단위로 걸리는 REQ-IQW-012 보다 좁다. 선택 결함은 D18~D20 이다.

plan 감사 처분 — PASS-with-debt (리드·운영자 결정, 2026-09-11): D17 을 부채로 안고 run 단계로 넘긴다. 조건은 둘이다. (1) run 위임문에 "새 init 실행 테스트는 새 파일에만 추가한다" 는 제약을 싣는다. (2) M6 마감 때 §E.2 에, 기존 `internal/cli` 테스트 파일에 추가된 `^+func Test` 줄 수가 0 임을 보이는 명령과 그 출력을 그대로 기록한다. D18~D20 을 run 단계에서 채택하면 그 기록은 `.moai/reports/t583/verdict.md` 에 남긴다. 재감사는 하지 않는다. Implementation Kickoff Approval 은 운영자 결정 대기 중이며, run 단계 진입은 그 승인을 기다린다.

Gap — lint 판정 빌드 좌표: 1·2회차와 이 문서의 `moai spec lint` 종료 코드 0 은 설치본 `/Users/goos/go/bin/moai`(`v3.2.0-rc.7`, `…-ged71054d3-dirty`)가 낸 것이다. 이 빌드는 워크트리 HEAD `120436f58` 의 조상이 아니며 그 역도 아니다(plan 감사 2회차 관측, 2026-09-11 재측정: `git merge-base --is-ancestor ed71054d3 HEAD` 종료 코드 1, `git merge-base --is-ancestor HEAD ed71054d3` 종료 코드 1). 따라서 lint 판정은 이 트리로 만든 빌드에 귀속되지 않는다.

확인 항목 해소 (2026-09-11): 기존 테스트 소급 범위는 리드 결정 (a) 로 정해졌다 — 코드 점검표는 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트에만 적용하고, init 실행 테스트를 돌리는 모든 run 슬롯은 실제 홈 지문 절차(spec.md §4.3, REQ-IQW-014·016, AC-IQW-015)를 따른다. plan.md 에 남은 확인 필요 표식 없음.

run 단계 안내: §E.2 는 슬롯 표(슬롯마다 한 행, 행 머리는 슬롯 번호)로 시작하고, 각 행에 지문 명령·두 지문·종료 코드를 남긴다. 형식과 판정식은 acceptance.md AC-IQW-015 에 있다.

Pre-spawn divergence origin/develop...HEAD = 17 1 at 2026-09-11 (first plan-time measurement); the 17 commits (t657 web, t545 settings form, t564 AC ids) touch none of internal/cli/wizard, internal/cli/init.go, internal/cli/update_wizard.go, internal/core/project, internal/shell (git diff --stat empty); absorb deferred to the integration window.

plan 작성 중 재측정 (같은 워크트리, 2026-09-11):

```
$ git fetch origin develop -q; git rev-list --count --left-right origin/develop...HEAD
17	1
$ git diff --stat HEAD...origin/develop -- internal/cli/wizard internal/cli/init.go internal/cli/update_wizard.go internal/core/project internal/shell | tail -2; echo "diffstat-exit=$?"
diffstat-exit=0
```

(diff --stat 출력 없음, 종료 코드 0.)

v0.1.3 직전 재측정 (같은 워크트리, HEAD `120436f58`, 2026-09-11): `origin/develop...HEAD` = `118 1`. SPEC 대상 경로 가운데 develop 이 바꾼 파일은 `internal/cli/update_wizard.go` 하나이며, 커밋은 `c4990eea7`(t587, `applyWizardConfig` 의 system.yaml 오류 처리, `update_wizard.go:312-340` 에 +13/-4)다. research.md·design.md·plan.md·spec.md 가 인용하는 `update_wizard.go` 줄(`:64`, `:133`, `:307-310`)은 `git show origin/develop:internal/cli/update_wizard.go` 에서도 같은 줄 번호에 같은 내용이다. 흡수는 여전히 통합 창으로 미룬다. AC-IQW-003 과 완료 정의의 위저드 범위 확인은 이제 흡수한 `develop` 과의 merge-base 부터 재므로, t587 을 흡수해도 그 때문에 빨개지지 않는다.

```
$ git fetch origin develop -q 2>&1; git rev-list --count --left-right origin/develop...HEAD
118	1
$ git diff --stat HEAD...origin/develop -- internal/cli/wizard internal/cli/init.go internal/cli/update_wizard.go internal/core/project internal/shell
 internal/cli/update_wizard.go | 17 +++++++++++++----
 1 file changed, 13 insertions(+), 4 deletions(-)
$ git log --oneline HEAD..origin/develop -- internal/cli/update_wizard.go
c4990eea7 fix(update): repair recovery hint, git-mode render, archive order, wizard errors (t587)
$ git merge-base develop HEAD
93182d137159c4facbf87c66dea3fd69160a6b8b
$ git diff --name-only develop...HEAD | wc -l
       0
$ git diff --quiet 120436f58 develop -- internal/cli/update_wizard.go; echo "pinned-vs-develop-exit=$?"
pinned-vs-develop-exit=1
$ git diff --quiet develop...HEAD -- internal/cli/update_wizard.go; echo "mergebase-exit=$?"
mergebase-exit=0
```

(마지막 두 줄: 옛 리터럴 핀 형식은 로컬 develop 에 든 t587 때문에 이미 종료 코드 1 이다. merge-base 형식의 0 은 카드 커밋이 아직 없어 대조군이 0 인 상태의 값이므로 "측정 불가" 이며, 통과 근거가 아니다.)

SPEC ID 자기 검사:

```
$ ID="SPEC-INIT-QUIET-WIZARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL; ls .moai/specs | grep -c QUIET
PASS
0
```

## §E.2 Run-phase Evidence

Slot table (AC-IQW-015). Every `go test` call against `./internal/cli` or `./internal/core/project/...` is declared here before it runs. Slots are granted by the lead and executed by the lane orchestrator (card t583); the fingerprint command is acceptance.md AC-IQW-015's reference form (`stat -f %m`, macOS), identical before and after except the file name. Tree for M1 slots: HEAD `1b3666cc1` plus the uncommitted M1 change (initializer.go seam, three new test files).

| Slot | Declared | Command | go test exit | before.out lines | .err bytes (before/after) | Result |
|---|---|---|---|---|---|---|
| SLOT-1 | 2026-09-11 | `go test ./internal/core/project/... -run '^(TestInitializer_ShellConfigSeamGate\|TestConfigureShellEnvFn_DefaultIsProductionFunc)$' -count=1 -v` | 0 (`--- PASS:` top-level 2, both subtests PASS, `no tests to run` 0; `.moai/state/verify/t583/ac004-gate.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-2 | 2026-09-11 | same selector, `-count=2 -v` | 0 (`--- PASS:` top-level 4, `--- FAIL` 0, `no tests to run` 0; `.moai/state/verify/t583/ac016-project.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-3 | 2026-09-11 | `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v` | 1 — TOOL_FAILURE (tdd-result-contract): `[build failed]`, `fileSHA256 redeclared in this block` (`internal/cli/update_preserve_my_harness_test.go:20:6` vs new `internal/cli/init_home_guard_test.go:130:6`); no test ran (`.moai/state/verify/t583/ac004-primary.txt`, 25 lines) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-4 | 2026-09-11 | `go test ./internal/cli -run '^(TestHomeGuard_RejectsPathInsideRealHome\|TestHomeGuard_AcceptsTempDir)$' -count=1 -timeout 600s -v` | not run — same package build failure as SLOT-3; slot returned to the lead, re-declared after the fix | — | — | — |
| SLOT-5 | 2026-09-11 | `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v` | not run — same reason as SLOT-4 | — | — | — |
| SLOT-6 | 2026-09-11 | SLOT-3 rerun after the rename: `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v` | 0 (`--- PASS:` 1, `--- FAIL` 0, `no tests to run` 0, `ok … 1.629s`; `.moai/state/verify/t583/ac004-primary.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-7 | 2026-09-11 | SLOT-4 command: `go test ./internal/cli -run '^(TestHomeGuard_RejectsPathInsideRealHome\|TestHomeGuard_AcceptsTempDir)$' -count=1 -timeout 600s -v` | 0 (`--- PASS:` 2, `--- FAIL` 0, `no tests to run` 0, `ok … 0.869s`; `.moai/state/verify/t583/ac005.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-8 | 2026-09-11 | SLOT-5 command: `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v` | 0 (`--- PASS:` 2, `--- FAIL` 0, `no tests to run` 0, `ok … 2.259s`; `.moai/state/verify/t583/ac016-cli.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-9 | 2026-09-12 | mutant A applied (`initializer.go:333` `if opts.SkipShellConfig {`), tree HEAD `9fc4bded0`: `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v` | 1 — expected RED (`--- FAIL:` 1, `init_shell_seam_test.go:59: shell-config seam calls = 0, want exactly 1`; `ac004-mutant-a.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-10 | 2026-09-12 | mutant B applied (`configureShellEnv` body `return ConfigureShellEnvFn(i.logger)` → `return nil, nil`), same command | 1 — expected RED (`--- FAIL:` 1, same "calls = 0" message; `ac004-mutant-b.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-11 | 2026-09-12 | reverted (backup `cp`, `cmp` exit 0, `git diff --quiet` → `reverted-exit=0`), same command | 0 (`--- PASS:` 1, `no tests to run` 0, `ok … 1.442s`; `ac004-revert.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-12 | 2026-09-12 | mutant C1 as written in AC-IQW-016 (delete `initializer_shell_seam_test.go:62` `t.Cleanup(...)`): `go test ./internal/core/project/... -run '^(TestInitializer_ShellConfigSeamGate\|TestConfigureShellEnvFn_DefaultIsProductionFunc)$' -count=2 -v` | 1 — TOOL_FAILURE, not an execution RED: `declared and not used: origSeam`, `[build failed]`; no test ran. Not counted as evidence | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-12R | 2026-09-12 | mutant C1 compiling form: line 62 → `_ = origSeam // mutant C1: restore removed` (`go vet` exit 0; swap-line count 2 → 1), same command | 1 — expected RED: round 1 both PASS, round 2 `--- FAIL: TestConfigureShellEnvFn_DefaultIsProductionFunc` 1 (`default points at … want defaultConfigureShellEnv … a previous test may not have restored the seam`; `ac016-mutant-c1.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-13 | 2026-09-12 | mutant C2 compiling form: `init_home_guard_test.go:202` → `_ = origSeam // mutant C2: restore removed` (`go vet` exit 0; swap-line count 2 → 1): `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v` | 1 — expected RED: round 1 PASS, round 2 `--- FAIL:` 1 (`home guard: shell-config seam is … on entry, want the original … (a previous test did not restore it)`; `ac016-mutant-c2.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-14 | 2026-09-12 | mutant D: `t.Parallel() // mutant D` as the first line of `TestRunInit_ShellConfigStepReachedViaSeam` (`go vet` exit 0): `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v` | 1 — expected panic: `testing: test using t.Setenv, t.Chdir, or cryptotest.SetGlobalRandom can not use t.Parallel`; `grep -c 'can not use t.Parallel'` → 1 (`ac016-mutant-d.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-15 | 2026-09-12 | all mutants reverted (four files `git diff --quiet` → `all-reverted-exit=0`): core/project command `-count=2 -v` | 0 (`--- PASS:` top-level 4, `--- FAIL` 0, `no tests to run` 0, `ok … 0.416s`; `ac016-project-revert.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-16 | 2026-09-12 | all mutants reverted: internal/cli command `-count=2 -timeout 600s -v` | 0 (`--- PASS:` 2, `--- FAIL` 0, `no tests to run` 0, `ok … 1.708s`; `ac016-cli-revert.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-17 | 2026-09-12 | M2 RED run, tree HEAD `87ed4cb7e` + uncommitted `internal/cli/init_quiet_wizard_test.go` (a foreign `go test ./internal/kanban/` was running at grant time; the lane waited until `ps` showed 0 go test/build processes): `go test ./internal/cli -run '^(TestRunInit_QuietWizardUnsetResolvesToDefaults\|TestRunInit_QuietWizardObserverDetectsNonDefault\|TestRunInit_QuietWizardSectionFilesMatchNonInteractive\|TestRunInit_QuietWizardProvisionsMCPByDefault\|TestRunInit_QuietWizardFlagsStillPersist)$' -count=1 -timeout 600s -v` | 1 — matches the prediction: top-level `--- PASS:` 4 (AC-IQW-006, 007a, 008, 011), `--- FAIL: TestRunInit_QuietWizardProvisionsMCPByDefault` 1 with `/claude` FAIL (`init_quiet_wizard_test.go:489: harness claude: provisioning announcement present = false, want true`) and `/codex`, `/both` PASS; `no tests to run` 0, `build failed` 0 (`m2-red.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-18 | 2026-09-12 | M3+M4 GREEN check, tree HEAD `82fc81b69` + uncommitted M3/M4 change (16 modified, 7 deleted); pre-run `ps` 0 go processes, load 6.63: `go test ./internal/cli -run '^(TestRunInit_QuietWizardUnsetResolvesToDefaults\|TestRunInit_QuietWizardObserverDetectsNonDefault\|TestRunInit_QuietWizardSectionFilesMatchNonInteractive\|TestRunInit_QuietWizardProvisionsMCPByDefault\|TestRunInit_QuietWizardFlagsStillPersist\|TestRunInit_ShellConfigStepReachedViaSeam\|TestHomeGuard_RejectsPathInsideRealHome\|TestHomeGuard_AcceptsTempDir\|TestFlagBeatsWizard_Page3Settings\|TestFlagBeatsWizard_MatchesProfilePrecedence\|TestEnableLSPDefault_MatchesWizardDefault\|TestSeedMirrorsProductionLSPSeed\|TestRunInit_WorktreeAutoCreateFlagBeatsWizard\|TestRunInit_WorkflowToggleFlagsAbsentByteIdentical\|TestRunInit_WorkflowToggleFlagsPersist\|TestRunInit_WizardCodexReachesBothConsumers\|TestRunInit_WizardCodexDeclinesMCPProvisioning\|TestRunInit_FlagClaudeBeatsWizardCodex\|TestRunInit_FlagBothBeatsWizardCodexAndForcesProvisioning\|TestRunInit_WizardBothForcesProvisioningOverDecline\|TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence\|TestRunWizardFn_ZeroInvocationsNonInteractive\|TestMCPPrecedenceComment_StatesHarnessRule\|TestRunInit_CallsMCPProvisioning\|TestRunInit_ThenDoctorCodexWiringHealthy\|TestRunInit_ClaudeOnlyThenDoctorStaysSilent\|TestInitWizardIdentityPersisted\|TestInitGitFlagOverridesDetection\|TestInitGitDetectionFillsConfig\|TestInitNoNetworkBeforeWizard)$' -count=1 -timeout 600s -v` (30 names) | 0 (top-level `--- PASS:` 30, `--- SKIP` 0, `--- FAIL` 0, `no tests to run` 0, `ok … 10.587s`; `TestRunInit_QuietWizardProvisionsMCPByDefault/claude` flipped RED → GREEN, `/codex` and `/both` PASS; `m4-green-selected.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-19 | 2026-09-12 | whole package, lead-approved exception (WizardResult field removal is package-wide); same tree as SLOT-18; pre-run `ps` 0, load 5.36, no other go command during the run: `go test ./internal/cli -count=1 -timeout 1500s -v` | 1 — `--- PASS:` 3578, `--- SKIP:` 24, `--- FAIL:` 4, `FAIL … 858.367s` (`m4-cli-full.txt`). Failing tests: `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile`, `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite`, `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition` ("audited production file changed after coverage tip: internal/cli/launcher.go" / `internal/hook/session_end.go`), `TestAuditLagUsesBinlagSeam` (line-pinned sweep: "baseline hit home_state_coverage.go:251/243 MISSING … NEW ancestry hit :245/:253"). Attribution (comparison, not a re-measurement): the uncommitted diff touches none of `home_state_coverage*`, `launcher.go`, `session_end.go`, `mcp_build_identity*`, `binlag`; `git log HEAD..develop` on those files lists `5b7927b15` (t600, "stop home-state coverage tests from reading live history") and `92494400f` (t606) — develop fixes for exactly this family, absent from this branch. Treated as base-tree known-red pending the merge-tree re-measure in the integration window; not measured on the develop tree. Lead accepted this attribution (2026-09-12) and waived a develop-tree measurement in favour of the merge-tree re-measure | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-20 | 2026-09-12 | M5 verification, tree HEAD `17b52894d` + uncommitted M5 change (2 modified, 1 production + 5 test files deleted); the lead waived the slot for this package (it does not link the root `internal/cli`) but kept the fingerprint duty; pre-run `ps` 0 go processes, load 4.38: `go test ./internal/core/project/... -count=1 -v` | 0 (`--- PASS:` 105, `--- FAIL`/`--- SKIP` 0, `no tests to run` 0, `no test files` 0, `ok … 1.616s`; `m5-project-full.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-21 | 2026-09-12 | M5 cli-side check, same tree; pre-run `ps` 0, load 6.85: `go test ./internal/cli -run '^(TestRunInit_WorkflowToggleFlagsAbsentByteIdentical\|TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence\|TestRunInit_WorktreeAutoCreateFlagBeatsWizard\|TestRunInit_CallsMCPProvisioning)$' -count=1 -timeout 600s -v` | 0 (`--- PASS:` 4, `--- FAIL`/`--- SKIP` 0, `no tests to run` 0, `ok … 2.410s`; `m5-cli-selected.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0 |
| SLOT-22 | 2026-09-12 | M6 AC-IQW-007b mutant A, tree HEAD `e334bd1c0` + `opts.ProjectMode = "team"` inserted in the interactive block of `init.go`; pre-run `ps` 0, load 5.42: `go test ./internal/cli -run '^TestRunInit_QuietWizardUnsetResolvesToDefaults$' -count=1 -timeout 600s -v` | 1 — expected RED: `--- FAIL:` 1, `init_quiet_wizard_test.go:367: removed keys did not resolve to their defaults (1 findings)` (`ac007b-a.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-23 | 2026-09-12 | M6 mutant B: `opts.WorktreeAutoCreate, opts.WorktreeAutoCreateSet = true, true` in the same block (mutant A reverted first, `cmp` exit 0), same command | 1 — expected RED: `--- FAIL:` 1, same defaults message (`ac007b-b.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-24 | 2026-09-12 | both mutants reverted by backup `cp` (`cmp` exit 0, `git diff --quiet -- internal/cli/init.go` → `reverted-exit=0`), same command | 0 (`--- PASS:` 1, `no tests to run` 0; `ac007b-revert.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-25 | 2026-09-12 | AC-IQW-016 closing two-run observation: `go test ./internal/core/project/... -run '^(TestInitializer_ShellConfigSeamGate\|TestConfigureShellEnvFn_DefaultIsProductionFunc)$' -count=2 -v` | 0 (`--- PASS:` 4, `no tests to run` 0; `ac016-project-final.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-26 | 2026-09-12 | AC-IQW-016: `go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v` | 0 (`--- PASS:` 2, `no tests to run` 0; `ac016-cli-final.txt`) | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-27 | 2026-09-12 | M6 mutant C (lead-approved, measures the lead's check item): the `case agentWiringBoth: mcpDeclined = false` arm deleted from `init.go` (`go vet` exit 0): `go test ./internal/cli -run '^(TestRunInit_FlagBothBeatsWizardCodexAndForcesProvisioning\|TestInitAgentFlagBothWiresCodexArtifacts\|TestRunInit_WizardBothForcesProvisioningOverDecline\|TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence\|TestRunInit_CallsMCPProvisioning)$' -count=1 -timeout 600s -v` | 0 — mutant SURVIVED: `--- PASS:` 4, `--- FAIL` 0, `no tests to run` 0 (`mutant-c.txt`). Selector defect found and corrected below: `TestInitAgentFlagBothWiresCodexArtifacts` does not exist, and a non-existent name is silently ignored (only an all-miss prints `no tests to run`), so this run swept 4 of the 5 intended names | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
| SLOT-27b | 2026-09-12 | mutant C still applied, selector corrected to the names that exist (`git grep '^func Test'` on `init_agent_flag_test.go`): `go test ./internal/cli -run '^(TestRunInit_AgentBothWiresBothSides\|TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning\|TestRunInit_CodexProvisioningDeclineIsolated\|TestRunInit_AgentAbsentLeavesNoCodexFiles\|TestRunInit_AgentClaudeLeavesNoCodexFiles)$' -count=1 -timeout 600s -v` | 0 — mutant SURVIVED again: `--- PASS:` 5, `--- FAIL` 0, `no tests to run` 0 (`mutant-c2.txt`). Reverted by backup `cp`, `cmp` exit 0, `git diff --quiet` → `reverted-exit=0` | sha 7, mtime 6 | 0/0 (sha), 0/0 (mtime) | home-diff-exit=0 |
SLOT-6 is the first slot that runs an init execution test through the new home-safety helper; its home-diff-exit=0 is the first runtime confirmation that the helper plus seam spy leave the real home untouched on this machine. Behavioral RED for M1 (mutants A, B, C1, C2, D) is still owed and needs further internal/cli and internal/core/project slots.

Fingerprint form deviation (recorded, not hidden). The worktree-isolation guard refused AC-IQW-015's reference fingerprint command as a single invocation ("too complex to verify that it stays inside the worktree"), so each fingerprint is taken as three plain commands, byte-identical before and after except `before`/`after` in the file names: (A) `shasum -a 256 "$HOME/.claude/settings.json" "$HOME/.zshenv" "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.profile" "$HOME/.bashrc" "$HOME/.bash_profile" > home-SLOT-<n>-<phase>-sha.out 2> …-sha.err`; (B) `stat -f '%N mtime=%m' "$HOME/.zshenv" "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.profile" "$HOME/.bashrc" "$HOME/.bash_profile" > …-mtime.out 2> …-mtime.err`; (C) `ls -d "$HOME/.claude/hooks/moai" > …-hooks.out 2> …-hooks.err`. The 8 AC items are covered: settings.json sha256 (A), hooks/moai presence (C), six rc files sha256 (A) and mtime (B). All seven files existed and `hooks/moai` was absent at SLOT-1 (checked with `ls -ld` first), so (C) exits 1 by design with a fixed "No such file" message on stderr; for (C) the comparison is `diff` of the two `.err` files instead of the zero-byte rule. Expected line counts in this form: sha 7, mtime 6. The per-slot `home-diff-exit` above is 0 only when all three diffs are 0.

VET (not a slot, no test execution): `go vet ./internal/core/project/...` → `vet-exit=0` (`.moai/state/verify/t583/m1-vet-project.txt`).

SLOT-3 repair (2026-09-11): the new helper `fileSHA256` was renamed to `homeGuardFileSHA256` in `internal/cli/init_home_guard_test.go` only. manager-develop swept all 11 top-level identifiers of the two new internal/cli test files: each has exactly one declaration hit and zero word hits outside the new files (the sweep uses `([^A-Za-z0-9_]|$)` because `\b` is not a word boundary in `git grep -E` — a control on the old name found `update_preserve_my_harness_test.go:20` only with the corrected form). Compile check allowed without a slot by the lead: `go vet ./internal/cli/` → `vet-exit=0`, output 0 bytes (`.moai/state/verify/t583/m1-vet-cli.txt`). SLOT-4/5 and the SLOT-3 rerun are re-declared when the next internal/cli slot is granted.

M1 non-slot observation — AC-IQW-004 body preservation (lane, 2026-09-11, tree above): `git show 120436f58:internal/core/project/initializer.go > .moai/state/verify/t583/ac004-initializer-base.go` → `show-exit=0`; base and head `ConfigOptions` extractions both 4 lines; `diff ac004-body-base.txt ac004-body-head.txt` → `body-diff-exit=0`.

M1 commit: status draft → in-progress landed with the M1 commit (manager-develop, 2026-09-12); mutant RED slots (A, B, revert, C1, C2, D) and mutant E are still owed against this committed tree.

Mutant E (AC-IQW-004 body preservation, no `go test`, not a slot; lane, 2026-09-12, tree HEAD `9fc4bded0` clean for `initializer.go`): pre-check `git diff --quiet -- internal/core/project/initializer.go` → `pre-clean-exit=0`; backup `cp` to `.moai/state/verify/t583/mutant-e-backup.go`; deleted the `PreferLoginShell:        true,` line (initializer.go:694) from `defaultConfigureShellEnv`; extracted with the AC-IQW-004 command into `ac004-body-mutant-e.txt`; `diff ac004-body-base.txt ac004-body-mutant-e.txt` → `body-diff-exit=1`, output `4d3` / `< 		PreferLoginShell:        true,` (`ac004-mutant-e-diff.txt`); base extract `wc -l` → `4`. Revert by backup `cp` (no `git restore`): `cmp` against the backup → `cmp-exit=0`, `cmp` against `git show HEAD:internal/core/project/initializer.go` → `cmp-head-exit=0`, `git diff --quiet` → `reverted-exit=0`. Mutant slots SLOT-9..SLOT-16 requested; lead queued lane-1 after lane-2 → lane-9 → lane-7 → lane-6 (2026-09-12).

M1 behavioral RED closed (2026-09-12, slots granted by the lead and returned after SLOT-16): mutants A, B, C1 (compiling form), C2, D and E each turned their observation red for the stated reason, and the reverted tree passed again (SLOT-11, SLOT-15, SLOT-16). Every mutant was reverted by backup `cp` and confirmed with `cmp` exit 0; the four M1 files `git diff --quiet` against `9fc4bded0` → 0 after SLOT-14. Lead verdict on the slot results: accepted (2026-09-12).

Sync obligation (lead, 2026-09-12 — close in THIS card's sync, not a later revision): acceptance.md AC-IQW-016 describes mutants C1 and C2 as "delete the restoring assignment". Applied literally, Go rejects the file (`declared and not used: origSeam`, SLOT-12) so no execution RED can be observed. The observation was taken with the compiling form (`_ = origSeam` in place of the `t.Cleanup` line, SLOT-12R and SLOT-13). manager-spec must reword the C1 and C2 mutant descriptions in AC-IQW-016 to that compiling form during sync. **DISCHARGED in commit `1c16e4227`** (manager-spec, 2026-09-12): acceptance.md:451-453 now prescribes replacing the restoring line with `_ = origSeam` and states why, with SLOT-12 vs SLOT-12R as the evidence; every existing expectation survived verbatim. This sentence is kept rather than deleted because the sync agent read this obligation, did not open acceptance.md, and reported the debt as still open — the lane corrected that in verdict.md §10.3. Check a debt against its target file, not against the document that recorded it.

M2 RED disposition differs from the plan prediction (lead ruling, 2026-09-12). plan.md §M2 predicted AC-IQW-006 and AC-IQW-008 red on the current tree. That red appears only when the real wizard fills the fields M3 deletes (`AuditModel`, `TodoEnabled`, and the rest) with default answers; a compiling test cannot reproduce it once injected results carry zero values, and filling those fields to force a failure would be a fake RED. AC-IQW-006 and AC-IQW-008 therefore enter without a RED-now observation. Their ability to fail rests on AC-IQW-007a (the in-file negative control) and AC-IQW-007b (the M6 mutant), which M6 MUST execute and record. Lead accepted this substitution. M2 test file: `internal/cli/init_quiet_wizard_test.go` (new, 5 tests; no existing `internal/cli` test file changed — D17). Author prediction from code reading, not yet observed: only `TestRunInit_QuietWizardProvisionsMCPByDefault/claude` red (AC-IQW-009, `MCPProvision` absent → announcement missing); SLOT-17 decides.

Shell-rc hypothesis measured (SLOT-20, 2026-09-12). The plan-phase hypothesis recorded in verdict §7.2.2 — that `internal/core/project` tests call `Init()` without `SkipShellConfig` or a home redirect, so Step 6 could append to the real `~/.zshenv` — is now measured for the first time, not merely read from code: the whole package ran (105 PASS) with the six rc files' mtime and sha256 identical before and after, and `~/.claude/settings.json` unchanged. On this machine the package does not write the real home. This measures the tree after M5 removed the dead writers; it does not by itself establish what the pre-M5 tree did.

M6 check item (lead, 2026-09-12): with the interactive MCP default now true, `init_agent_wizard_test.go:139` and `:159` ("both beats a decline") no longer discriminate on the interactive path — a decline can no longer be expressed there, so the `case agentWiringBoth` forcing branch in `init.go` is only meaningful on the non-interactive `--llm both` path. M6 must find (or record the absence of) a test that discriminates that branch on the non-interactive path, e.g. a mutant that removes the forcing and observes a red.

M6 closing evidence, no slot required (lane, 2026-09-12, tree HEAD `e334bd1c0`, card base `git merge-base develop HEAD` = `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`).

AC-IQW-001/002/003, measured before the M6 slots:

```
$ go test ./internal/cli/wizard/... -run '^TestInitQuestions_QuietSet$' -count=1 -v   → test-exit=0, PASS 1, no tests to run 0   (ac001.txt)
$ go test ./internal/cli/wizard/... -run '^(TestRemovedQuestionsAbsentFromInitSet|TestRemovedQuestionsHaveNoOrphanTranslations|TestRemovedQuestionsHaveNoCaptureBranch|TestSharedQuestionsRetainedForReconfigure)$' -count=1 -v   → test-exit=0, PASS 4, no tests to run 0   (ac002.txt)
$ git grep -nE '"(project_mode|…|mcp_provision)"' -- internal/cli/wizard ':!*_test.go'   → exit=1 (0 hits)
$ git grep -cE '"(project_mode|…|mcp_provision)"' 120436f58 -- internal/cli/wizard ':!*_test.go'   → questions.go:11, translations.go:33, wizard.go:11 (control, 55 hits)
$ go test ./internal/cli/wizard/... -run '^(TestReconfigureQuestionsOrder|TestQuestionOrder)$' -count=1 -v   → test-exit=0, PASS 2, no tests to run 0   (ac003.txt)
$ diff reconf-base.txt reconf-head.txt   → diff-exit=0 (base body 52 lines)
$ git diff --name-only develop...HEAD | wc -l   → 49 (card-scope control)
$ git diff --quiet develop...HEAD -- internal/cli/update_wizard.go   → 0
$ git diff --quiet HEAD -- internal/cli/update_wizard.go   → 0
```

AC-IQW-016 closing text sweep (same tree):

```
$ git grep --untracked -lE 'ConfigureShellEnvFn[[:space:]]*=[^=]' -- 'internal/cli/*_test.go' 'internal/core/project/*_test.go'
internal/cli/init_home_guard_test.go
internal/core/project/initializer_shell_seam_test.go        → sweep-exit=0, 2 files (cli 1, core/project 1)
$ … --all-match -e '<swap>' -e 't\.Parallel\('   → parallel-exit=1 (no swapping file calls t.Parallel)
$ … --all-match -e '<swap>' -e 't\.Cleanup\(' | wc -l   → 2
$ … --all-match -e '<swap>' -e 't\.Setenv\(' | wc -l   → 2
$ git grep --untracked -cE '<swap>' -- …   → init_home_guard_test.go:2, initializer_shell_seam_test.go:2
```

D17 debt evidence (the plan-audit PASS-with-debt condition — no new init execution test in a pre-existing `internal/cli` test file):

```
$ git diff develop...HEAD -- 'internal/cli/*_test.go' ':!internal/cli/init_home_guard_test.go' ':!internal/cli/init_quiet_wizard_test.go' ':!internal/cli/init_shell_seam_test.go' ':!internal/cli/wizard/*' > d17-existing-cli-tests.txt   (562 lines)
$ command grep -cE '^\+func Test' d17-existing-cli-tests.txt   → 0 (d17-added-exit=1)
$ command grep -cE '^\+func Test' d17-cli-test-diff.txt        → 11 (control: the same pattern over the unrestricted cli test diff)
$ git diff develop...HEAD --diff-filter=A --name-only -- 'internal/cli/*_test.go'
internal/cli/init_home_guard_test.go
internal/cli/init_quiet_wizard_test.go
internal/cli/init_shell_seam_test.go   (the three files this card added; every new test lives in one of them)
```

AC-IQW-014 quality gate (same tree):

```
$ go vet ./...                              → vet-exit=0, output 0 bytes      (m6-vet.txt)
$ go build ./...                            → build-exit=0                    (m6-build.txt)
$ GOOS=windows GOARCH=amd64 go build ./...  → winbuild-exit=0                 (m6-winbuild.txt)
$ golangci-lint run ./internal/cli/... ./internal/core/project/...  → lint-exit=0, "0 issues."  (m6-lint.txt)
```

AC-IQW-007b verdict: mutants A and B each turned `TestRunInit_QuietWizardUnsetResolvesToDefaults` red for the stated reason and the reverted tree passed again (SLOT-22/23/24). Together with AC-IQW-007a's in-file negative control, these are the failure evidence standing in for the RED-now cells AC-IQW-006 and AC-IQW-008 could not have (lead ruling, above).

Lead check item answered by measurement, not by reading (SLOT-27/27b): mutant C deleted the `case agentWiringBoth: mcpDeclined = false` arm and **no test failed** — 4 PASS on the first selector and 5 PASS on the corrected one. No test discriminates the non-interactive `--llm both` forcing branch. Per the lead's instruction this card writes no new test for it; it is recorded as a follow-up candidate and carried into the sync Residual-risk.

Instrument defect found and corrected in the same run: the first mutant-C selector named `TestInitAgentFlagBothWiresCodexArtifacts`, which does not exist. `go test -run` silently ignores a non-existent name when other names in the alternation match, and prints `no tests to run` only when nothing matches at all — so the miss was invisible in the output and was caught by comparing the PASS count (4) with the selector count (5). The real names were resolved with `git grep '^func Test' -- internal/cli/init_agent_flag_test.go` and re-run as SLOT-27b. Future mutant selectors verify name existence before the run.

M6 addendum — the AC checks §E.3 first reported as Gaps, now measured (lane, 2026-09-12, tree HEAD `4079087ab`, card base `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`). The §E.3 matrix rows for AC-IQW-005, 012 and 013 are superseded by this block.

AC-IQW-012 — worktree wiring prose consolidated:

```
$ git grep -nE 'worktree_auto_create|WorktreeAutoCreate' -- internal/cli/wizard internal/cli/init.go ':!*_test.go'   → ac012-exit=1 (0 hits)
$ git grep -cE 'worktree_auto_create|WorktreeAutoCreate' 120436f58 -- internal/cli/wizard internal/cli/init.go ':!*_test.go'   → control: init.go:2, questions.go:2, translations.go:3, types.go:1, wizard.go:2 (10 hits)
$ sed -n '/Worktree advisory/,/WorktreeAutoCreate bool/p' internal/core/project/initializer.go | awk 'tolower($0) ~ /wizard/ {c++} END {print "wizard-mentions=" c+0}'   → wizard-mentions=0 (range 9 lines, so the sed range is not empty)
$ git show 120436f58:… | same awk   → control-wizard-mentions=2
```

AC-IQW-013 — dead writers gone, retained items present:

```
$ git grep -nE 'writeWorkflowAuditYAML|writeWorkflowTodoYAML|writeFeedbackAutoSubmitYAML|writeWorkflowProjectContinuationYAML|AuditConfigSet' -- internal ':!*_test.go'   → deleted-exit=1 (0 hits)
$ git grep -cE 'writeWorkflowAuditYAML|AuditConfigSet' 120436f58 -- internal/core/project/initializer.go   → control 5
$ git grep -nE 'func writeProjectModeYAML|func WriteWorkflowTogglesYAML|func provisionMCPEntryUnlessDeclined|opts\.MCPProvision' -- internal ':!*_test.go'   → kept-exit=0, 7 lines: provisionMCPEntryUnlessDeclined (init.go:254), the interactive write (init.go:721), the read (init.go:968), two comment lines, writeProjectModeYAML (initializer_expansion.go:54), WriteWorkflowTogglesYAML (initializer_workflow_toggles.go:38)
$ go vet ./internal/cli/... ./internal/core/project/...   → 0 (also covered by the AC-IQW-014 module-wide vet above)
```

AC-IQW-005 — home-safety checklist sweep (the guard tests themselves ran at SLOT-7: `test-exit=0`, PASS 2, `no tests to run` 0, `ac005.txt`):

```
$ git diff --name-only --diff-filter=A develop...HEAD -- 'internal/cli/*_test.go'   → added-committed-exit=0; init_home_guard_test.go, init_quiet_wizard_test.go, init_shell_seam_test.go
$ git ls-files --others --exclude-standard -- 'internal/cli/*_test.go'   → added-uncommitted-exit=0, 0 lines (all three are committed now)
$ git grep --untracked -lE 'runInit[A-Za-z]*\(|prepareSafeInitHome\(' -- 'internal/cli/*_test.go' | sort   → 17 caller files
$ wc -l < ac005-targets.txt   → 3 (the three planned files; ≥ 3 satisfied)
$ comm -12 ac005-targets.txt ac005-reach.txt | wc -l   → 3 — equals the target count, so the searcher reads every target and the next line's 0 is a verdict rather than a vacuous pass
$ comm -12 ac005-targets.txt ac005-forbidden.txt   → no output; | wc -l → 0   (forbidden-file control: 180 files in internal/cli carry one of the tokens, so the pattern matches)
$ git grep -c 't.Setenv("HOME"' 120436f58 -- internal/cli/init_agent_wizard_test.go   → 7 (pattern control)
$ git grep --untracked -c 'bash_profile' -- internal/cli/init_home_guard_test.go   → 1 (the sixth rc file is in the comparison list)
```

Mutant F (AC-IQW-005's required record — a fourth init execution test file outside the three planned ones, sweep-only, never compiled or run, deleted after the sweep, never committed):

- **First attempt was too weak and is recorded as such.** The file called nothing: it carried `_ = runInit` without parentheses, so the discriminant `runInit[A-Za-z]*\(` did not match it. The sweep was unchanged — targets 3, forbidden intersection 0 — which looks identical to a clean tree. This is the documented limitation of the discriminant (acceptance.md AC-IQW-005, third note: the discriminant is the NAME OF AN INIT-EXECUTING CALL, so a file that executes init through some other spelling escapes the sweep), observed here rather than merely read.
- **Corrected mutant** (`internal/cli/init_quiet_wizard_mcp_test.go`, `t.Setenv("HOME", t.TempDir())` plus a real `runInit(nil, nil)` call) turned the sweep RED as specified: uncommitted-added list gained the file, the added∩caller intersection became `init_home_guard_test.go, init_quiet_wizard_mcp_test.go, init_quiet_wizard_test.go, init_shell_seam_test.go`, the target count went `3 → 4`, and the forbidden intersection went `0 → 1`, printing `internal/cli/init_quiet_wizard_mcp_test.go`.
- After `rm`, `git status --short` shows only the foreign ` M .moai/config/sections/workflow.yaml` and the clean sweep's forbidden intersection is 0 again.

AC-IQW-010 body-preservation half (the execution half ran at SLOT-18 and SLOT-21):

```
$ diff codex-absence-base.txt codex-absence-head.txt      → codex-absence-diff-exit=0
$ diff byte-identical-base.txt byte-identical-head.txt    → byte-identical-diff-exit=0
$ wc -l codex-absence-base.txt byte-identical-base.txt    → 20 and 12 (both base extracts non-empty)
```

AC-IQW-011 — the cli half ran inside SLOT-18 (`TestRunInit_QuietWizardFlagsStillPersist` and `TestRunInit_WorkflowToggleFlagsPersist`, both PASS there, on the pre-M5 tree). The core half, run here on the current tree (SLOT-28; the lead waived the slot for `internal/core/project` because it does not link the root `internal/cli`, keeping the fingerprint duty):

```
$ go test ./internal/core/project/... -run 'TestWriteProjectModeYAML|TestWriteWorkflowTogglesYAML' -count=1 -v   → test-exit=0, `--- PASS:` 8, `no tests to run` 0   (ac011-core.txt)
   fingerprints: sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0   (home-SLOT-28-*)
```

AC-IQW-014 remainder measured after §E.3 was written:

```
$ go test ./internal/cli/wizard/... -count=1   → test-exit=0, `ok … 3.104s`   (m6-wizard-full.txt)
$ go test ./internal/core/project/... -cover -count=1   → cover-exit=0, `coverage: 88.8% of statements` (target 85%)   (m6-cover-project.txt)
```

SLOT-29 — `internal/cli` coverage (lead-granted; pre-run `ps` 0 go processes and load 6.20 after waiting out a foreign `go test ./internal/hook/` run and a load spike to 12.33; no other go command during the run; tree HEAD `4079087ab`):

```
$ go test ./internal/cli -cover -count=1 -timeout 1500s   → cover-exit=1, 955.796s   (m6-cover-cli.txt)
  coverage: 82.4% of statements
  --- FAIL: 4 — TestHomeStateChangedSurfaceCoverageConsumesFreshProfile, TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite, TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition, TestAuditLagUsesBinlagSeam
  fingerprints: sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0   (home-SLOT-29-*)
```

Two findings, neither resolved by this card:

1. **`internal/cli` coverage is 82.4%, below the 85% target** (`.moai/config/sections/quality.yaml` `test_coverage_target: 85`). This is a package-wide figure for a package far larger than this card's scope, and no baseline for the same package on the card base tree was measured, so this run does not establish whether the card moved the number in either direction. Reported to the lead as a number, not as a disposition; the lead owns what happens next.
2. **The same four tests fail as in SLOT-19**, byte-for-byte the same family (home-state coverage + binlag). The attribution recorded at SLOT-19 stands: the card's diff touches none of their files, and develop's `5b7927b15` (t600) and `92494400f` (t606) fix exactly this family and are not yet absorbed. The merge-tree re-measure in the integration window is what settles it. The nested `coverage: 14.7% of statements in ./internal/cli, ./internal/homestate, ./internal/hook/handoff, ./internal/hook, ./internal/kanban` line in the output belongs to the bounded sub-suite those failing tests run themselves, not to this invocation.

SLOT-30 — the base-tree control the lead ordered so the 82.4% figure can be attributed (card base `ee99507fbe3b4a22c6a0a74815723d222dfdc04d` extracted with `git archive | tar -x` into the session scratchpad; pre-run `ps` 0, load 4.91; no other go command during the run):

```
$ (in the extracted base tree) go test ./internal/cli -cover -count=1 -timeout 1500s   → base-cover-exit=1, 949.992s   (m6-cover-cli-base.txt)
  coverage: 82.3% of statements
  --- FAIL: 10
  fingerprints: sha-diff-exit=0, mtime-diff-exit=0, hooks-diff-exit=0 → home-diff-exit=0   (home-SLOT-30-*)
```

**Attribution verdict: the card did not lower coverage.** base 82.3% → card 82.4%, a change of +0.1pp. The 85% target is missed on both trees, so the miss is a pre-existing package-level state rather than something this card introduced; raising it is out of this card's scope and belongs to a follow-up (lead ruling, 2026-09-12).

Two limits of this control, stated rather than smoothed over:

1. **The base tree is not a git repository.** `git archive` extracts a tree without `.git`, so every test that reads repository history fails there for an environmental reason. That is why the base run shows 10 failures against the card tree's 4: the extra six (`TestBuildIdentity_VersionDerivationUnchanged`, `TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation`, `TestPreCommitLegacyNoRecord`, `TestHomeStateValidationCommandWrappersAndHelperFailures`, `TestTodoHistoryNeverPrompts`, `TestVersionStampRegistry`) are artifacts of the missing repository, not base-tree defects. The failure sets are therefore NOT directly comparable, and the coverage figures carry whatever small difference those six failing-vs-passing tests make.
2. **What the control does establish**: all four of the card tree's failures — `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile`, `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite`, `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition`, `TestAuditLagUsesBinlagSeam` — also fail on the base tree. The SLOT-19/SLOT-29 known-red attribution, until now a reading of which files the diff touches, is now a measurement: those four are red without any of this card's changes present.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-09-12
run_commit_sha: 4079087ab                 # backfilled in the sync commit; the M6 commit could not cite its own hash
run_head_before_m6: e334bd1c0
card: t583
branch: WT-init-quiet-wizard
card_base: ee99507fbe3b4a22c6a0a74815723d222dfdc04d   # git merge-base develop HEAD
ac_total: 17                              # AC-IQW-001..016, 007 split a/b
ac_pass_count: 8
ac_pass_with_debt_count: 7
ac_gap_count: 2                           # AC-IQW-012, AC-IQW-013 — not measured in run phase
ac_fail_count: 0
cross_platform_build:
  darwin_native: exit 0                   # go build ./... (m6-build.txt)
  windows_amd64: exit 0                   # GOOS=windows GOARCH=amd64 go build ./... (m6-winbuild.txt)
  vet: exit 0                             # go vet ./... (m6-vet.txt, 0 bytes)
new_warnings_or_lints_introduced: 0       # golangci-lint ./internal/cli/... ./internal/core/project/... → "0 issues." (m6-lint.txt)
known_red: SLOT-19 FAIL 4                 # base-tree known-red; see § Known red below
slot_declarations: 29                     # rows beginning "| SLOT-"
slot_fingerprint_sets: 27                 # home-SLOT-*-after-*.out sets
slot_home_diff_exit_0_rows: 27
slot_home_diff_exit_nonzero_rows: 0
l44_pre_commit_fetch: not performed       # no push in this card; integration-window duty
l44_post_push_fetch: not performed        # lane does not push (lead batch-push)
preserve_list_post_run_count: not measured
total_run_phase_files: 49                 # git diff --name-only develop...HEAD | wc -l, measured at HEAD e334bd1c0
m1_to_mN_commit_strategy: one commit per milestone (M1 seam + M1 mutant evidence + M2 + M3/M4 + M5 + M6 docs), no amend, no force-push, unpushed
```

### Per-AC verdict matrix

Every row names the command that decided it and the evidence file under `.moai/state/verify/t583/` (slot rows in §E.2 carry the verbatim output). Rows marked **Gap** were not measured in the run phase and are NOT reported as passes.

| AC | Verdict | Deciding command | Evidence |
|---|---|---|---|
| AC-IQW-001 | PASS | `go test ./internal/cli/wizard/... -run '^TestInitQuestions_QuietSet$' -count=1 -v` → exit 0, PASS 1, `no tests to run` 0 | `ac001.txt` |
| AC-IQW-002 | PASS | `go test ./internal/cli/wizard/... -run '^(TestRemovedQuestionsAbsentFromInitSet\|…)$' -count=1 -v` → exit 0, PASS 4; `git grep -nE '"(project_mode\|…\|mcp_provision)"' -- internal/cli/wizard ':!*_test.go'` → exit 1 (0 hits) with the `120436f58` control at 55 hits | `ac002.txt` |
| AC-IQW-003 | PASS | `go test … -run '^(TestReconfigureQuestionsOrder\|TestQuestionOrder)$'` → exit 0, PASS 2; `diff reconf-base.txt reconf-head.txt` → 0 (base 52 lines); card-scope control 49 files; `git diff --quiet develop...HEAD -- internal/cli/update_wizard.go` → 0 and `git diff --quiet HEAD -- …` → 0 | `ac003.txt`, `reconf-base.txt`, `reconf-head.txt` |
| AC-IQW-004 | PASS | Primary SLOT-6 `-run '^TestRunInit_ShellConfigStepReachedViaSeam$'` → exit 0, PASS 1; gate SLOT-1 `./internal/core/project/...` → exit 0, PASS 2; `diff ac004-body-base.txt ac004-body-head.txt` → 0 (base 4 lines); mutants A/B RED (SLOT-9/10) and revert PASS (SLOT-11); mutant E `body-diff-exit=1` | `ac004-primary.txt`, `ac004-gate.txt`, `ac004-mutant-a.txt`, `ac004-mutant-b.txt`, `ac004-revert.txt`, `ac004-mutant-e-diff.txt` |
| AC-IQW-005 | PASS-WITH-DEBT | Guard half measured: SLOT-7 `-run '^(TestHomeGuard_RejectsPathInsideRealHome\|TestHomeGuard_AcceptsTempDir)$'` → exit 0, PASS 2. **Gap**: the target-list sweep (`ac005-targets.txt` / reach control / forbidden-token intersection) and mutant F were not run in the run phase; no `ac005-*` sweep artifacts exist | `ac005.txt` (guard half only) |
| AC-IQW-006 | PASS-WITH-DEBT | **No RED-now cell** — lead ruling recorded in §E.2 (a compiling test cannot reproduce the predicted red once injected results carry zero values; forcing it would be a fake RED). Substitute failure evidence: AC-IQW-007a in-file negative control + AC-IQW-007b mutants A/B. GREEN observed at SLOT-18 (`TestRunInit_QuietWizardUnsetResolvesToDefaults` among 30 top-level PASS) and again at SLOT-24 | `m4-green-selected.txt`, `ac007b-revert.txt` |
| AC-IQW-007a | PASS | SLOT-17 and SLOT-18 both PASS `TestRunInit_QuietWizardObserverDetectsNonDefault`; the observer is the in-file negative control standing in for AC-IQW-006/008 | `m2-red.txt`, `m4-green-selected.txt` |
| AC-IQW-007b | PASS | Mutant A (`opts.ProjectMode = "team"`) SLOT-22 → exit 1, FAIL 1 (`removed keys did not resolve to their defaults (1 findings)`); mutant B (`WorktreeAutoCreate…= true, true`) SLOT-23 → exit 1, FAIL 1; revert SLOT-24 → exit 0, PASS 1, `reverted-exit=0` | `ac007b-a.txt`, `ac007b-b.txt`, `ac007b-revert.txt` |
| AC-IQW-008 | PASS-WITH-DEBT | **No RED-now cell** — same lead ruling as AC-IQW-006; substitute evidence is AC-IQW-007a (in-file negative control) and AC-IQW-007b (M6 mutants). GREEN at SLOT-17 and SLOT-18 for `TestRunInit_QuietWizardSectionFilesMatchNonInteractive` | `m2-red.txt`, `m4-green-selected.txt` |
| AC-IQW-009 | PASS | Genuine RED→GREEN: SLOT-17 `TestRunInit_QuietWizardProvisionsMCPByDefault/claude` FAIL (`harness claude: provisioning announcement present = false, want true`) with `/codex` and `/both` PASS; SLOT-18 all three subtests PASS | `m2-red.txt`, `m4-green-selected.txt` |
| AC-IQW-010 | PASS-WITH-DEBT | Test half measured: SLOT-18 and SLOT-21 both PASS `TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence` and `TestRunInit_WorkflowToggleFlagsAbsentByteIdentical`. **Gap**: the two body-extraction diffs (`codex-absence-*`, `byte-identical-*`) were not run; no such artifacts exist | `m4-green-selected.txt`, `m5-cli-selected.txt` |
| AC-IQW-011 | PASS-WITH-DEBT | cli half measured: SLOT-18 PASS for `TestRunInit_QuietWizardFlagsStillPersist` and `TestRunInit_WorkflowToggleFlagsPersist`. **Gap**: the `./internal/core/project/...` half (`ac011-core.txt`, `TestWriteProjectModeYAML\|TestWriteWorkflowTogglesYAML`) was not run as its own slot | `m4-green-selected.txt` |
| AC-IQW-012 | **Gap** | Not measured. The AC's four commands (`git grep -nE 'worktree_auto_create\|WorktreeAutoCreate' …`, its `120436f58` control, and the two `Worktree advisory` comment sweeps) were not run in the run phase and no output is recorded in §E.2 | none |
| AC-IQW-013 | **Gap** (partial) | Not measured except the static check: `go vet ./...` → exit 0 (`m6-vet.txt`). The deleted-symbol grep, its control, and the retained-symbol grep were not run and no output is recorded in §E.2 | `m6-vet.txt` (vet only) |
| AC-IQW-014 | PASS-WITH-DEBT | `go vet ./...` exit 0; `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run ./internal/cli/... ./internal/core/project/...` exit 0, "0 issues."; `./internal/core/project/...` full package SLOT-20 → exit 0, 105 PASS, 0 SKIP, 0 `no test files`. Debt: `./internal/cli` full package SLOT-19 → exit 1 with FAIL 4 (§ Known red below). **Gap**: `./internal/cli/wizard/...` was never run as a whole package | `m6-vet.txt`, `m6-build.txt`, `m6-winbuild.txt`, `m6-lint.txt`, `m5-project-full.txt`, `m4-cli-full.txt` |
| AC-IQW-015 | PASS-WITH-DEBT | Per slot: all 27 executed slots record the three-command fingerprint with `sha` 7 lines, `mtime` 6 lines, `.err` 0/0, and `home-diff-exit=0`; no slot records a non-zero value. Closing tally: declarations 29, fingerprint sets 27, `home-diff-exit=0` rows 27, non-zero rows 0. The 29-vs-27 difference is SLOT-4 and SLOT-5, declared then not executed (the SLOT-3 package build failure) and re-declared as SLOT-7 and SLOT-8 — so the literal three-way equality the AC states does not hold, and the AC passes only with that stated deviation | `home-SLOT-*-{before,after}-{sha,mtime,hooks}.{out,err}` |
| AC-IQW-016 | PASS | Text sweep: 2 swapping files (cli 1, core/project 1), `parallel-exit=1`, `t.Setenv` intersection 2, `t.Cleanup` intersection 2, per-file assignment count 2 each. Execution: SLOT-25 `./internal/core/project/... -count=2` → exit 0, PASS 4; SLOT-26 `./internal/cli -count=2` → exit 0, PASS 2. Mutants: C1 (SLOT-12R) RED on round 2, C2 (SLOT-13) RED on round 2, D (SLOT-14) panic `can not use t.Parallel`; reverts SLOT-15/16 PASS | `ac016-project-final.txt`, `ac016-cli-final.txt`, `ac016-mutant-c1.txt`, `ac016-mutant-c2.txt`, `ac016-mutant-d.txt` |

Fingerprint-form deviation (three plain commands instead of the AC's single reference invocation, forced by the worktree-isolation guard) is recorded in §E.2 and applies to every slot row above.

### Milestone summary

| Milestone | Commit | Content |
|---|---|---|
| M1 | `9fc4bded0` | shell-config seam, spy observation, home-safety helper; `draft → in-progress` landed here |
| M1 (evidence) | `87ed4cb7e` | mutant RED slot records SLOT-9..SLOT-16 |
| M2 | `82fc81b69` | RED init-execution tests for the quiet wizard (`internal/cli/init_quiet_wizard_test.go`, 5 tests) |
| M3+M4 | `17b52894d` | init wizard cut to four questions; interactive MCP provisioning default-on |
| M5 | `e334bd1c0` | removal of the config writers the quiet wizard left without a producer |
| M6 | this commit | closing evidence (SLOT-22..27b, D17 debt evidence, quality gate) and this §E.3 signal |

### Known red (SLOT-19, base-tree)

`go test ./internal/cli -count=1 -timeout 1500s -v` at the M3/M4 tree → exit 1, `--- PASS:` 3578, `--- SKIP:` 24, `--- FAIL:` 4 (`m4-cli-full.txt`). The four are `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile`, `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite`, `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition`, and `TestAuditLagUsesBinlagSeam` — the home-state-coverage plus binlag family.

Attribution (a comparison, not a re-measurement): the card's diff touches none of `home_state_coverage*`, `launcher.go`, `session_end.go`, `mcp_build_identity*`, or `binlag`; `git log HEAD..develop` on those files lists `5b7927b15` (t600, "stop home-state coverage tests from reading live history") and `92494400f` (t606) — develop fixes for exactly that family, absent from this branch. The lead accepted this attribution on 2026-09-12 and waived a develop-tree measurement. **What settles it is the merge-tree re-measure in the integration window**, not this record: until those four run green on the tree that absorbs `develop`, the red is attributed but not resolved.

### Gaps (explicitly unobserved in the run phase)

1. **AC-IQW-012** — not measured at all (four commands, zero output recorded).
2. **AC-IQW-013** — only `go vet` measured; the deleted-symbol grep, its `120436f58` control, and the retained-symbol grep were not run.
3. **AC-IQW-005 sweep half** — target-list construction, reach control, forbidden-token intersection, and mutant F were not run; only the two guard tests were.
4. **AC-IQW-010 body-preservation half** — the two `git show 120436f58:… | sed` extractions and their diffs were not run.
5. **AC-IQW-011 core half** — `./internal/core/project/... -run 'TestWriteProjectModeYAML|TestWriteWorkflowTogglesYAML'` was never declared as a slot.
6. **AC-IQW-014 wizard package** — `go test ./internal/cli/wizard/... -count=1` (whole package) was never run; only `-run`-selected subsets were.
7. **AC-IQW-015 closing tally** — the AC's three-way equality does not hold literally (29 / 27 / 27); the difference is explained above but the AC text has no clause for a declared-then-unexecuted slot.
8. **Coverage** — no `go test -cover` measurement was taken for `internal/cli/wizard` or `internal/core/project`, so acceptance.md §4's 85% target is unverified for this card.
9. **SLOT-19 on the develop tree** — not measured (lead waiver); the merge-tree re-measure is owed.
10. **`moai spec lint` build coordinate** — the §E.1 gap stands: the exit-0 lint verdict came from the installed `v3.2.0-rc.7` build, which is neither an ancestor nor a descendant of this tree. No lint verdict in this card is attributed to a build made from this tree.

### Residual risk (what could still be wrong despite what was observed)

- **(a) Mutant C survived.** Deleting the `case agentWiringBoth: mcpDeclined = false` arm from `init.go` produced no failure — 4 PASS on the first selector (SLOT-27) and 5 PASS on the corrected one (SLOT-27b). No test discriminates the non-interactive `--llm both` forcing branch. Per the lead's instruction this card wrote no test for it; it is carried as a follow-up candidate.
- **(b) Interactive-path discrimination loss.** With the interactive MCP default now true, `init_agent_wizard_test.go:139` and `:159` ("both beats a decline") no longer discriminate on the interactive path — a decline can no longer be expressed there. Those two assertions still pass but no longer test what their names claim.
- **(c) AC-IQW-016 wording debt (owed in THIS card's sync).** acceptance.md AC-IQW-016 describes mutants C1 and C2 as "delete the restoring assignment"; applied literally Go rejects the file (`declared and not used: origSeam`, SLOT-12) and no execution RED is observable. The observations were taken with the compiling form (`_ = origSeam`). manager-spec must reword C1 and C2 to the compiling form during sync.
- **(d) Instrument fragility in mutant selectors.** `go test -run` silently ignores a non-existent name when other names in the alternation match; the first mutant-C selector named a test that does not exist and the miss was invisible in the output, caught only by comparing PASS count against selector count. Any future mutant run that does not verify name existence can report a silent partial sweep.
- **(e) Text-observation blind spots in AC-IQW-016.** The `t.Cleanup(` text check is satisfied by any `t.Cleanup` in the same file, and `-count=2` observes that restoration happened, not that it was registered via `t.Cleanup`. A test-function-tail restore would pass both and skip restoration on `t.Fatal`. Code review backs this.
- **(f) The shell-rc measurement is post-M5 only.** SLOT-20's unchanged rc fingerprints measure the tree *after* M5 removed the dead writers; it does not establish what the pre-M5 tree did on a machine whose rc files lack the configured lines.
- **(g) Already-configured machine.** `internal/shell/config.go:93,:167` skip writing when the line already exists, so an unchanged home fingerprint is weak evidence of no leak on this machine; the seam's effect rests on AC-IQW-004's spy call counts, not on the fingerprints.
- **(h) Unmeasured Gaps above.** Items 1-8 of the Gaps list are unobserved, not passing: AC-IQW-012 and AC-IQW-013 in particular assert removals this card performed, so a regression there would be invisible to every check that did run.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-09-12
sync_commit_sha: pending-backfill         # this sync commit cannot cite its own hash
sync_head_before: 1c16e4227
card: t583
branch: WT-init-quiet-wizard
b12_self_test_a: pass                     # grep -c 'SPEC-INIT-QUIET-WIZARD-001' CHANGELOG.md → 0 before emission (no duplicate entry)
b12_self_test_b: pass                     # acceptance.md distinct AC identifiers → 16 (non-zero); CHANGELOG states 16 and notes §E.3's 17 counts AC-IQW-007 split a/b
b12_self_test_c: pass                     # every path named in the CHANGELOG entry verified with ls: questions.go, types.go, wizard.go, translations.go, init.go, initializer.go, update_wizard.go present; initializer_audit.go absent as claimed (deleted by M5)
changelog_entry_position: "[Unreleased] → ### Changed (new subsection, first entry)"
frontmatter_status_transitions:
  spec_md: in-progress → completed        # this commit; updated: → 2026-09-12
  plan_md: not applicable                 # stateless on the status axis (spec-frontmatter-schema.md § Artifact Statelessness)
  acceptance_md: not applicable           # same
  design_md: not applicable               # same
  research_md: not applicable             # same
docs_sync:
  readme_4_locale: not performed          # README{,.ko,.ja,.zh}.md:288 falsified (model policy no longer asked at init) — handed to a follow-up docs card
  docs_site_4_locale: not performed       # docs-site/content/{en,ko,ja,zh}/getting-started/init-wizard.md falsified (fixed 3-page flow) — same follow-up
  rationale: ".moai/reports/t583/verdict.md §10.5 — pre-existing drift entanglement, non-parallel locale structure, unverified docs-site Vercel binding on develop (CLAUDE.local.md §4.1)"
mx_tag_validation: performed as a sync sub-step
  # @MX:NOTE + @MX:SPEC annotations landed with the implementation (e.g. questions.go InitQuestions);
  # no missing-annotation repair was needed in this commit and none was added.
spec_body_modified: false                 # spec.md §A-§H, plan.md, acceptance.md, design.md, research.md bodies untouched
verification_re_executed: none            # every figure cited in §10 of the verdict is an §E.2/§E.3 quotation; no lead-granted test slot was held by this sync step
open_debt:
  - "acceptance.md AC-IQW-016 mutant C1/C2 wording still reads 'delete the restoring assignment'; the compiling form (`_ = origSeam`) is what was observed. Reword is manager-spec's, not manager-docs'. Evidence: verdict.md §10.3, SLOT-12 vs SLOT-12R."
  - "User documentation (8 files, 4 locales × 2 surfaces) still describes the previous question set. Evidence: verdict.md §10.5."
  - "Four known-red internal/cli tests are attributed to develop fixes not yet absorbed; the merge-tree re-measure in the integration window settles them. Evidence: §E.3 § Known red, verdict.md §10.6(c)."
```

## §F Phase 4 Mode Selection

Recorded by the lane orchestrator (card t583) before the first run-phase `Agent()` spawn, after Implementation Kickoff Approval (operator, relayed by the lead, 2026-09-11) and the develop absorb (`2723447be`, HEAD^2 `4c99d973e`).

Input parameters:
- tier: L (spec.md; 6 milestones M1-M6)
- scope: about 15 files across `internal/cli/wizard`, `internal/cli`, `internal/core/project` (plan.md §F)
- domain count: 1 (Go CLI init path and its tests); no template, docs, or hook work
- file language mix: Go source and Go tests
- concurrency benefit: LOW — coding-heavy, milestones depend in order (M1 seam lands before M2 tests), and every `internal/cli` / `internal/core/project` compile or test run is a lead-granted slot with a real-home fingerprint (AC-IQW-015)
- Agent Teams: not requested

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | semantic multi-file change, not a trivial edit |
| serial | **yes** | one `manager-develop` (cycle_type=tdd) per milestone, lane runs slot-gated verification between steps |
| fanout | no | single domain, coding-heavy; parallel writers would race on `init.go` and the wizard files |
| sweep | no | not a uniform mechanical transform |
| manager-lead | no | entry predicate needs cross-domain fan-out; this card is one domain with serial dependencies |

Decision: serial

Justification: the milestones are ordered by a hard dependency (REQ-IQW-011 seam before any init execution test), and verification cannot run inside the writer because test slots are granted by the lead one at a time. A single sequential writer with lane-run slots keeps one writer on the tree and keeps every `go test` call declared and fingerprinted.

Boundary case: none of the numeric thresholds is at ±1; the deciding factor is the slot-gated verification, not file count.
