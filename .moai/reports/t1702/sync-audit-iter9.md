auditor-model: claude-opus-5-5

# t1702 sync 감사 iter9 (수리 델타 cf141b5409..a29bc40be5 + 카드 전체 가드 테스트 탐색)

verdict: FAIL
audited_sha: a29bc40be5ba143fed1543c84cfd4c146d1ba370

## 요약

판정은 FAIL이다. 이번 수리(R1, N5, N6)는 모두 닫혔다. 레지스트리 가드는 HEAD에서 통과하고, 검사기 수리는 변이로 고정돼 있다. Codex 재감사(수리 델타 한정)도 pass로 나왔다(영수증 `rcpt-2973cf866890155ca03162e5`).

FAIL의 원인은 지시받은 가드 테스트 탐색에서 새로 찾은 결함 하나다. 저장소 전체 홈 접근 가드 `TestHomeGuard_RepositoryHasNoNewHomeAccess`(internal/testguard)가 HEAD에서 실패한다. 실패 6건 중 4건이 이 카드가 새로 만든 테스트 파일 두 개의 `~/` 문자열이다. R1과 같은 부류다. 이름으로 테스트를 고르다 보니, 카드가 건드린 파일을 스캔하는 가드를 iter6~iter8 모두 돌리지 않았다. 이번 수리 델타에서 생긴 결함은 아니다. 처음 들어온 커밋은 8f600fca00과 b590f29594다.

다만 감안할 점이 있다. 나머지 2건은 카드와 무관한 파일(main에도 있는 줄)이다. 그래서 이 테스트는 main 기준으로도 이미 빨간 상태일 가능성이 높다(main 트리에서 직접 돌리지는 않았다, Gaps 참고). 그래도 카드의 4건만으로도 가드는 독립적으로 실패한다. main 쪽 2건을 고친 뒤에 이 카드를 착지하면 가드가 다시 빨갛게 된다. 고치는 방법은 기계적이다. 네 줄에 `// home-guard:allow <reason>`를 달거나 `~/` 표기를 바꾸면 된다.

점수: 78/100. Functionality 72 FAIL(카드 diff가 저장소 가드에 새 위반을 더함), Security 90, Craft 78, Consistency 75. Functionality가 반드시 통과해야 하는 차원이라 전체 판정은 FAIL이다.

## Codex 호출 기록

- 호출: `mcp__moai__codex_audit`를 정확히 한 번 했다. mode `native`, target `baseBranch`, project_root `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702`, delta_base `cf141b5409dd38b9a95c9948a3d1836771db7651`(d-20261106T080039Z-305c의 "수리 델타 한정"에 맞췄다). 서버가 답한 리뷰 범위는 `delta_base (merge base cf141b5409dd38b9a95c9948a3d1836771db7651)`, 서버 빌드는 `04d8c27a7e`, 응답 모델은 `gpt-6.1-sol`이다.
- 결과: 리뷰가 실제로 실행됐다(거부 아님). verdict `pass`, findings `[]`이다. Codex는 정적 검토만 했고 테스트는 돌리지 않았다고 밝혔다. 재호출은 하지 않았다.
- 영수증: **`rcpt-2973cf866890155ca03162e5`**. 이전 영수증 `rcpt-690d37c6f81b6c1daa5c7050`은 재사용하지 않았다. 이 sync 감사 단계에서 두 번째로 센 호출이다(2/2).
- 결과 요약(digest): backend codex, gate required, verdict pass, receipt rcpt-2973cf866890155ca03162e5. 필수 codex 게이트는 충족됐다.
- Codex는 델타만 봤기 때문에 아래 H1(카드 앞 커밋의 테스트 파일)을 볼 수 없었다. H1은 이 감사관이 실행해서 찾은 결함이다.

## B2(이전 쓰기 경로 충돌)에 대한 판단

`moai decision read --scope card:t1702`의 출력을 스크래치패드에 저장한 뒤 grep으로 확인했다. 기록 d-20261106T055331Z-c28f(digest e01040700705)의 원문은 다음과 같다. "운영자 결정 2026-11-06 (리더 AskUserQuestion 원문 답 "설치는 쓰고 판정만 읽기 (권장)"): 설치·update가 프로필 루트에 쓰는 기존 설계는 유지. 프로젝트 사본 제거 가능 판정(SafeToRemoveProjectCopy/CheckProfileReach)과 doctor 프로필 도달 판정만 읽기 전용(os.Root 읽기, 쓰기·디렉터리 생성 없음)."

이번 델타의 비테스트 변경은 두 가지뿐이다. `profile_reach.go`의 주석 한 줄, 그리고 `update_destructive_registry.go`의 레지스트리 행 이름과 주석이다. 프로필 루트에 대한 새 쓰기는 없다. B2는 닫힌 상태로 유지한다.

## 결함 목록

| id | 심각도 | 분류 | 위치 | 요약 |
|---|---|---|---|---|
| H1 | BLOCKING | in-diff(카드 전체), 재현, 감사관 발견. 8f600fca00과 b590f29594에서 생김, 수리에서 생긴 것 아님 | `internal/cli/update_profile_reach_test.go:110`, `:259`, `:282`; `internal/userassets/profile_reach_test.go:589` | 카드가 새로 만든 두 테스트 파일의 `~/` 문자열 4곳이 저장소 홈 접근 가드 `TestHomeGuard_RepositoryHasNoNewHomeAccess`에 "NEW home access"로 잡혀 테스트가 실패한다. seed에도 없고 `home-guard:allow` 표시도 없다 |
| N7 | MINOR | optional, in-diff | `internal/cli/update_destructive_registry_test.go:179-182` | `rootHandleNames` 문서 주석이 아직 "*os.Root 필드·매개변수와 os.OpenRoot로 받은 변수"만 적는다. 이번에 추가한 하위 루트(`<root>.OpenRoot`) 추적은 빠져 있다 |
| N8 | MINOR | optional, touched-file(diff 밖, main부터 있던 문제) | `internal/cli/update_destructive_registry.go:45-46` | 머리말은 "17 (file, function) rows covering 27 call sites"라고 하지만, 실제 행은 16개이고 Sites 합은 26이다. main도 16행이라 이 카드에서 생긴 문제는 아니다 |

H1 필수 조치: 네 줄 각각에(또는 바로 윗줄에) `// home-guard:allow <reason>`를 단다. 사유 예시는 "assertion message text, no home access"다. 아니면 메시지에서 `~/` 표기를 빼도 된다. 그 뒤 `go test ./internal/testguard -run TestHomeGuard_RepositoryHasNoNewHomeAccess -count=1 -v`로 이 카드 파일 4건이 목록에서 사라졌는지 확인한다. 남는 2건(`internal/agenticfactory/core/sanitize_test.go:95`, `internal/cli/factory_fixture_home_test.go:36`)은 카드와 무관한 main의 문제이므로 별도 카드로 넘긴다. seed 재생성(`TESTGUARD_WRITE_SEED=1`)으로 덮는 것은 가드 주석상 금지된 방법이다("no line is ever added by hand", seed는 이관용).

이전 회차 항목의 처리 상태:

| 항목 | 상태 | 근거 |
|---|---|---|
| R1 | 닫힘 | 레지스트리 행이 `removeConfigRootMoaiHooks`(Sites 1)를 가리키고 `TestDestructiveTargetRegistry_CoversAllSites`가 PASS다(E1). 검사기의 fixpoint를 끈 변이에서는 두 테스트가 모두 RED였다(E2) |
| N5 | 닫힘 | `Root` 주석이 코드와 일치한다. 목록에 있는 프로필은 `dir`(`profile_reach.go:116`, `:121`), 목록 밖 항목은 `profileDir`(`:133`, `:138`)이다 |
| N6 | 닫힘 | 링크를 만들 수 없으면 `t.Skip`으로 건너뛴다. 실패로 처리하지 않는다 |
| B2 | 닫힘 유지 | 위 절 참고 |

### R1 수리에 대해 물은 것

- **다른 행의 Sites가 바뀌었나.** 아니다. 레지스트리 diff에서 바뀐 것은 행 이름 하나와 주석뿐이고(Sites 1→1), 검사 결과와 레지스트리 키·개수가 완전히 일치해서 테스트가 PASS한다. fixpoint를 끈 변이의 요약 줄은 "scanned 15 … registry has 16"이었다. 그러니 HEAD의 스캔은 16쌍이고, 새로 잡힌 것은 `removeConfigRootMoaiHooks` 하나다. 다른 함수의 개수가 하위 루트 추적 때문에 늘었다면 "site count drift"나 "unregistered"로 실패했을 것이다.
- **거짓 양성·거짓 음성.** fixpoint는 RHS가 하나인 `AssignStmt`만 따라간다. 그래서 `var sub, _ = root.OpenRoot(...)` 형태(DeclStmt)는 놓친다. 그런데 기존 `os.OpenRoot` 탐지도 같은 한계가 있었으니 새 거짓 음성은 아니다. 이름 기반이라 같은 함수 안에서 이름을 가리면 과대 계수가 될 수 있지만, 이것도 기존 설계의 한계다. 현재 스캔 범위에서는 결과가 레지스트리와 일치한다.
- **Protection 문구가 새 함수에 맞나.** 맞는다. 이 함수는 `<config-root>/hooks/moai` 하나만 지운다(`update.go:1145-1181`). "the loop replays the SAME per-root radius"에서 말하는 루프는 호출자 `ensureGlobalSettingsEnv`(`update.go:1120`)의 루프를 가리키고, 그 내용도 사실이다. 문구는 정확하다.

## 가드 테스트 탐색 (지시 항목)

**어떻게 찾았나.**
1. `git diff main...a29bc40be5 --name-only`로 카드가 바꾼 파일 36개를 뽑았다(cli 16, userassets 9, config 3, hook 4, profile 1, reports 7 등).
2. `git grep -l '"go/parser"\|"go/ast"\|"go/build"\|"go/types"\|golang.org/x/tools/go/packages' a29bc40be5 -- 'internal/**/*_test.go'` → 50개 파일.
3. 그중 cli·config·hook·userassets·profile·testguard·guardliveness·homestate 테스트에서 `ReadFile(…\.go")|ParseFile(|ParseDir(|packages.Load(`을 grep했다. 또 저장소 전역 스캔 테스트에서는 `WalkDir|filepath.Walk|os.ReadDir|filepath.Glob|packages.Load`를 grep해서 스캔 루트를 확인했다.
4. 스캔 범위가 카드 파일(`internal/cli/*.go` 디렉터리, `update*.go`, `doctor.go`, `internal/config`, `internal/hook`, 저장소 전체 `internal/`·`cmd/`)에 닿는 테스트를 골랐다. 자기 패키지 디렉터리만 보는 것(agenticfactory/effect·claim·ledger·monitor·review, jev, jevledger, statusline, escalation, factory, web, harnessrun, homestate, guardliveness의 고정 파일 목록, cli/update/skillresidue)은 제외했다. 어느 것도 카드가 건드린 디렉터리가 아니다.

**실행한 것(전부 `-run` 지정, 패키지 전체 실행 없음).**

| 패키지 | 고른 테스트 | 결과 |
|---|---|---|
| internal/cli | TestDestructiveTargetRegistry_, TestCountDestructiveCalls_, TestUpdateFlow_NoDirectHomeCalls, TestUpdate_NoAskUserQuestion, TestInventoryGuard_, TestStdinFailClosed_, TestProductionStringLiteralsUseLeaderLaneVocabulary, TestVocabularyGuardControls, TestNoRetiredWordIdentifiers, TestRetiredWordIdentifierScanHasTeeth, TestLaunchViewSingleParse, TestNoBareGLMEnvVarLiteralsInCLIProduction, TestGLMEnvSetParity, TestJevCallPath_, TestEveryRegisteredHookHandlerHasBugreportName, TestConfigPathSeamIsNotBypassedByACopiedBody, TestTreeScopePolicy_SourceReadsNoEnvironment, TestPaths_SourceHasNoOpeningCalls, TestSeamDefault, TestCleanReinstallOptionsLiteralsInjectErrOut, TestBinaryLag_, TestBuildIdentity_, TestDoctorGolden_, TestSD_AC022_SchemaStatementsFrozen, TestQWR_AC007_GateDisabledReadsNothing, TestQAS_AC014_, TestPreCommitWarningWriterWiring + 카드 테스트(TestDoctorUserInstall, TestDoctorProjectVsLock, TestEnsureGlobalSettingsEnv_, TestUpdateSettingsEnvHooksSymlinkRefused, TestRemoveConfigRootMoaiHooks_, TestRunUserAssetUpdatePhase_, TestCheckProfileSkillReach, TestDoctorSkillReach, TestUpdateGlobalSettingsEnv_, TestUserAssetRootsLine) | `ok … internal/cli 61.136s` |
| internal/hook | TestProtectedZone(BaselineCovered·Liveness 하위 테스트 포함), TestHMPSourceGuard, TestHMPTestIsolation, TestDeferredScanOptOut_, TestIrreversibleActionTrips, TestDetectorNeverAltersToolCall, TestUserRoot, TestZoneUserRootCovered_ | `ok … internal/hook 71.316s` |
| internal/config | TestNoBareAnthropicEnvVarLiteralsInProduction, TestAnthropicBannedSetCoversAllNames, TestShippedConfigKeysHaveReaders, TestResolveAuditPlan_SourceIsPure, TestShippedUserRootEntries_, TestZoneUserRootMatch_, Profile, UserRoot | `ok … internal/config 81.796s` |
| internal/testguard | TestHomeGuard_, TestSessionWorktreeTests_AllUseRedirectHome | **FAIL**(H1, E3) |
| internal/bugreport | TestEveryRecoverSiteReportsOrIsAllowlisted, TestToolFailureHandlersDoNotImportCapture | `ok … 6.649s` |
| internal/agenticfactory/retirement | TestRetirement_(internal/cli 스캔) | `ok … 108.609s` |
| internal/goal, internal/template | TestGoalStatusConsumersHandleCancelled(cli·hook 패키지 로드), TestWorkflowWorktreeKeyHonesty$ | `ok … 19.510s`, `ok … 44.640s` |
| internal/userassets, internal/profile | 판정·enumeration·manifest 테스트 전체 이름 + TestLauncherProfileDirName | `ok … 3.978s`, `ok … 0.537s` |

`-list`로 cli 정규식이 실제로 고르는 가드 이름도 확인했다. TestBinaryLag_ 4개, TestDoctorGolden_ 4개, TestStdinFailClosed_ 14개, TestInventoryGuard_ 2개 등 35개가 나왔다(E4).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. HEAD는 a29bc40be5ba143fed1543c84cfd4c146d1ba370이고, 이 워크트리(작업 트리 깨끗함)에서 감사했다.
2. R1은 닫혔다. 레지스트리 가드와 새 검사기 테스트가 PASS이고, 검사기 수리를 끄면 둘 다 RED가 된다.
3. 수리 델타는 프로필 루트에 새 쓰기를 더하지 않는다.
4. 카드 diff는 저장소 홈 접근 가드에 새 위반 4건을 더한다(H1).
5. 탐색에서 고른 나머지 가드와 카드 테스트는 모두 통과한다.
6. gofmt와 go vet(네 패키지)은 깨끗하다.
7. 필수 codex 게이트는 영수증 rcpt-2973cf866890155ca03162e5로 충족됐고, Codex verdict는 pass다.

### Evidence

**E0 — 트리와 계획**
```
git rev-parse HEAD              → a29bc40be5ba143fed1543c84cfd4c146d1ba370
git rev-parse --show-toplevel   → /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702
git status --short              → (출력 없음)
moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702
  → "config_status": "ok", "model": "codex", codex gate "required" (config.gates), claude "off", glm "off",
    "cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
git diff cf141b5409 a29bc40be5 --stat
  → update_destructive_registry.go 5 +-, update_destructive_registry_test.go 65 ++, profile_reach.go 2 +-,
    profile_reach_judgment_t1702_test.go 8 +-, reports 2개
```

**E1 — R1 가드(실제 코드)**
```
GOCACHE=…/gocache-t1702 go test ./internal/cli -run 'TestDestructiveTargetRegistry_|TestCountDestructiveCalls_|TestRemoveConfigRootMoaiHooks_' -count=1 -v
--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.08s)
--- PASS: TestCountDestructiveCalls_ASubRootOfARootHandleIsARootHandle (0.00s)
--- PASS: TestRemoveConfigRootMoaiHooks_ASwapToALinkAfterTheCheckNeverReachesTheExternalTree (0.01s)
--- PASS: TestRemoveConfigRootMoaiHooks_ASwapToAnInRootLinkNeverReachesAnotherTree (0.01s)
--- PASS: TestRemoveConfigRootMoaiHooks_StillRemovesARealDirectoryAndRefusesALink (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.073s
```
(PASS일 때는 요약 줄을 찍지 않는다. 스캔 쌍 수는 아래 변이 출력에서 얻었다.)

**E2 — 검사기 변이(`-overlay`, 변이 파일은 스크래치패드 mut9-t1702/에 두었고 트리는 바꾸지 않음)**
변이 `m_nofixpoint`: `for changed := true; changed;` → `for changed := false; changed;`(하위 루트 추적을 끔).
```
GOCACHE=… go test ./internal/cli -overlay …/mut9-t1702/overlay_nofixpoint.json -run 'TestDestructiveTargetRegistry_|TestCountDestructiveCalls_' -count=1 -v
    update_destructive_registry_test.go:49: stale registry row: internal/cli/update.go removeConfigRootMoaiHooks is registered but the source has no destructive call site there
    update_destructive_registry_test.go:65: scanned 15 (file, function) pair(s) across internal/cli/update and internal/cli/update*.go; registry has 16 row(s)
--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.22s)
    update_destructive_registry_test.go:240: f: 0 destructive call site(s), want 1 (all: map[f:0 g:1 unrelated:0])
--- FAIL: TestCountDestructiveCalls_ASubRootOfARootHandleIsARootHandle (0.00s)
```
→ 잡혔다. 반대 방향(행 이름만 옛 이름으로 되돌림)은 iter8 E4의 cf141b5409 실측이 그 상태다("stale registry row … ensureSettingsEnvInConfigRoot").

**E3 — H1 재현(실제 코드)**
```
GOCACHE=… go test ./internal/config ./internal/testguard ./internal/bugreport ./internal/agenticfactory/retirement -count=1 -run '…|TestHomeGuard_|TestSessionWorktreeTests_AllUseRedirectHome|…'
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (17.35s)
    repo_guard_test.go:87: NEW home access internal/agenticfactory/core/sanitize_test.go:95: ~/ literal: "~/work/notes.md", // a tilde path names no user
    repo_guard_test.go:87: NEW home access internal/cli/factory_fixture_home_test.go:36: ~/ literal: t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
    repo_guard_test.go:87: NEW home access internal/cli/update_profile_reach_test.go:110: ~/ literal: t.Errorf("~/.claude/settings.json should not be created, stat err = %v", err)
    repo_guard_test.go:87: NEW home access internal/cli/update_profile_reach_test.go:259: ~/ literal: t.Fatal("ensureGlobalSettingsEnv must propagate an unreadable profile store, not clean ~/.claude only")
    repo_guard_test.go:87: NEW home access internal/cli/update_profile_reach_test.go:282: ~/ literal: if !strings.Contains(check.Message, "claude-profiles") || !strings.Contains(check.Message, "~/.claude") {
    repo_guard_test.go:87: NEW home access internal/userassets/profile_reach_test.go:589: ~/ literal: t.Error("ClaudeConfigRoots must error on an unreadable store, not return the bare ~/.claude")
    repo_guard_test.go:96: scanned 3931 test files: 166 violations, seed 184 entries, 6 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
FAIL	github.com/modu-ai/moai-adk/internal/testguard	17.950s
ok  	github.com/modu-ai/moai-adk/internal/config	81.796s
ok  	github.com/modu-ai/moai-adk/internal/bugreport	6.649s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/retirement	108.609s
```
(각 위반 줄 아래에 "Use testguard.RedirectHome(t) or testguard.RealHome(), or add `// home-guard:allow <reason>` …" 안내가 붙는다. 여기서는 생략했다.)
출처 추적:
```
git cat-file -e main:internal/cli/update_profile_reach_test.go   → fatal: … exists on disk, but not in 'main'   (카드가 새로 만든 파일)
git cat-file -e main:internal/userassets/profile_reach_test.go   → fatal: … exists on disk, but not in 'main'   (카드가 새로 만든 파일)
git log --oneline main..a29bc40be5 -S "should not be created, stat err" -- internal/cli/update_profile_reach_test.go → 8f600fca00 …
git log --oneline main..a29bc40be5 -S "not return the bare ~/.claude" -- internal/userassets/profile_reach_test.go → b590f29594 …
git grep -n -E "sanitize_test|factory_fixture_home_test|update_profile_reach_test|userassets/profile_reach_test" main -- internal/testguard/home_access_seed.txt → (출력 없음)
git grep -c … 9b632c1eb3 -- internal/agenticfactory/core/sanitize_test.go internal/cli/factory_fixture_home_test.go → 각 1  (카드와 무관한 2건은 merge base에도 있음)
git log --oneline -2 9b632c1eb3 -- internal/testguard/repo_guard_test.go internal/testguard/home_access_seed.txt → cda2be5b46 … (card t1801), 가드는 merge base 이전에 들어옴
```

**E4 — 탐색 실행(위 표의 명령)**
```
go test ./internal/cli -count=1 -run '<위 표의 cli 정규식>'                → ok  	github.com/modu-ai/moai-adk/internal/cli	61.136s
go test ./internal/hook -count=1 -run 'TestProtectedZone|TestHMPSourceGuard|TestHMPTestIsolation|TestDeferredScanOptOut_|TestIrreversibleActionTrips|TestDetectorNeverAltersToolCall|TestUserRoot|TestZoneUserRootCovered_'
                                                                         → ok  	github.com/modu-ai/moai-adk/internal/hook	71.316s
go test ./internal/goal ./internal/template -count=1 -run 'TestGoalStatusConsumersHandleCancelled|TestWorkflowWorktreeKeyHonesty$'
                                                                         → ok … internal/goal 19.510s / ok … internal/template 44.640s
go test ./internal/userassets ./internal/profile -count=1 -run '<판정·enumeration·manifest·launcher 이름>'
                                                                         → ok … internal/userassets 3.978s / ok … internal/profile 0.537s
go test ./internal/cli -count=1 -list '<가드 부분 정규식>'                 → 35개 이름(TestBinaryLag_ ×4, TestDoctorGolden_ ×4, TestStdinFailClosed_ ×14,
                                                                           TestInventoryGuard_ ×2, TestDestructiveTargetRegistry_CoversAllSites, …) / ok
```

**E5 — 정적 검사**
```
gofmt -l internal/userassets internal/config internal/profile internal/cli                → 출력 없음
GOCACHE=… go vet ./internal/userassets ./internal/config ./internal/profile ./internal/cli → 출력 없음
```

**E6 — 결정 원문**: `moai decision read --scope card:t1702 > …/scratchpad/decisions-t1702-iter9.txt` 다음에 grep. 56-58행에 c28f(위 B2 절 원문), 27-30행에 305c("감사 단계마다 최대 2회 … 리뷰 시작 전 … 거부돼 영수증이 생기지 않은 호출은 횟수에 넣지 않는다")가 있다.

**E7 — Codex**: 위 "Codex 호출 기록" 절. `{"verdict":"pass", "findings":[], "audit_receipt":"rcpt-2973cf866890155ca03162e5", "build_commit":"04d8c27a7e", "review_base":"delta_base (merge base cf141b5409dd38b9a95c9948a3d1836771db7651)"}`.

### Baseline-attribution
모든 테스트는 이 실행에서 워크트리 트리 a29bc40be5에 대해 돌렸다. 비공개 `GOCACHE=/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/671835d7-2501-4df7-8d56-57dfa4e7b2d0/scratchpad/gocache-t1702`를 썼고, 이번 실행에서 캐시 오류는 없었다. 테스트 바이너리는 전부 이 트리 소스로 빌드했다. 설치된 moai는 측정에 쓰지 않았다. 기록 열람(`decision read`)과 계획 조회(`verify audit-plan`)에만 썼다. Codex 리뷰는 서버 빌드 04d8c27a7e가 수행했고, 범위는 cf141b5409..HEAD 델타다.

### Gaps
- **거부된 도구**: 변이 파일을 만들려고 쓴 `sed -n 's/…/…/;w <scratchpad>'` 두 번이 워크트리 격리 가드에 거부됐다("cannot be shown not to be git"). 대신 검사기 변이는 Write 도구로 스크래치패드에 전체 파일을 써서 만들었다(E2). 레지스트리 행만 되돌리는 변이는 실행하지 않았다. 그 방향은 iter8 E4(cf141b5409) 실측으로 대신했다.
- H1의 기준선: 이 테스트가 main이나 merge base 트리에서 이미 실패하는지는 직접 돌리지 않았다. 카드와 무관한 2건이 merge base에 있고 seed에 없다는 것은 git grep으로 확인했다. 그래서 base에서도 RED라는 것은 추론이다. 이 테스트가 CI 필수 체크에 들어 있는지도 관측하지 않았다.
- `internal/cli`, `internal/hook` 전체 스위트와 `go test ./...`는 실행하지 않았다(지시). 탐색은 소스를 스캔하는 테스트를 grep 패턴으로 고른 것이다. 문자열 상수로 경로를 만드는 가드처럼 grep 패턴에 걸리지 않는 형태는 빠졌을 수 있다.
- 탐색에서 뺀 패키지(위 "어떻게 찾았나" 4번)는 실행하지 않았다.
- `mcp__moai__audit_multi`와 `--result-file` 대조는 실행하지 않았다(지시: codex_audit만 사용). 계획의 `enforced_required: [codex]`는 단독 `codex_audit` 영수증으로 채웠고, convergence_check는 관측하지 않았다.
- golangci-lint, 커버리지, `/verify`, Windows·Linux 실행, `GOOS=windows go vet`은 이번에 실행하지 않았다(iter8에서 실행했고, 이번 델타의 비테스트 변경은 주석과 문자열뿐이다).

### Residual-risk
- H1을 고친 뒤에도 카드와 무관한 2건 때문에 `TestHomeGuard_RepositoryHasNoNewHomeAccess`는 main이 고쳐질 때까지 계속 RED일 수 있다. 착지 판단에서 이 테스트의 RED를 카드 탓과 main 탓으로 나눠 읽어야 한다.
- 검사기는 여전히 문법 기반이다. `var x, _ = root.OpenRoot(...)` 형태나 핸들 별칭은 따라가지 못한다. 앞으로 그런 형태로 삭제 호출을 옮기면 R1과 같은 누락이 다시 생길 수 있다.
- 판정 함수에 호출자가 없다는 점, 오래된 overlayfs에서 Unknown이 과하게 나올 수 있다는 점, 정크 디렉터리 때문에 사본이 남는다는 점은 iter8과 같다.
- 변이 원본과 결정 기록 사본은 세션 스크래치패드에만 있다. 결정에 쓴 출력 줄은 이 파일에 옮겨 두었다.

## 반복 이력

- iter6: FAIL @732fce152d — B1, B2, S1, S2, M1~M4.
- iter7: FAIL @bee9cfb001 — 코드 BLOCKING 없음, 필수 codex 게이트 미충족(stale_server).
- iter8: FAIL @cf141b5409 — Codex P2를 재현해 BLOCKING R1로 판정(iter5 P3 수리에서 생김). 영수증 rcpt-690d37c6f81b6c1daa5c7050.
- iter9(이번): FAIL @a29bc40be5 — R1, N5, N6 닫힘, Codex pass(rcpt-2973cf866890155ca03162e5). 가드 탐색에서 BLOCKING H1(카드 초기 커밋 8f600fca00, b590f29594에서 생김)을 찾았다. 수리에서 생긴 BLOCKING은 이번 회차에도, 직전 회차에도 없어서 수리 중단 규칙(연속 2회)은 걸리지 않는다. 이번 sync 감사 단계의 codex 호출 한도(2회)는 다 썼다. H1 수리 뒤 재감사를 하려면 새 승인이나 다른 감사 경로가 필요하다.
