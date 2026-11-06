auditor-model: claude-opus-5-5

# t1702 sync 감사 iter8 (카드 전체 main...HEAD + 수리 델타 bee9cfb001..cf141b5409)

verdict: FAIL
audited_sha: cf141b5409dd38b9a95c9948a3d1836771db7651

## 요약

판정은 FAIL이다. 이번에는 게이트가 아니라 코드 쪽 결함 때문이다. 필수 codex 게이트는 이번 호출로 채워졌다(영수증 `rcpt-690d37c6f81b6c1daa5c7050`). 그런데 Codex가 지적한 문제를 이 트리에서 직접 재현해 보니 실제로 실패했다. 카드가 `internal/cli/update.go`의 삭제 호출을 `removeConfigRootMoaiHooks`로 옮기면서, 파괴적 작업 레지스트리 행은 이제 삭제 호출이 없는 `ensureSettingsEnvInConfigRoot`를 가리키게 됐다. 그 결과 드리프트 가드 `TestDestructiveTargetRegistry_CoversAllSites`가 HEAD에서 실패한다. 이 테스트는 `internal/cli` 전체 스위트에 들어 있어 CI가 빨갛게 된다. 카드 diff 안에서 재현된 결함이므로 BLOCKING이다.

iter7의 MINOR 세 건(N1 테스트, N2 문서, N4 핸들 동일성)은 닫혔다. 변이 세 개로 확인했고, N4 수리가 정상 파일을 체계적으로 거부하지 않는다는 것도 함께 확인했다. 새로 생긴 결함은 MINOR 두 건뿐이다. 이번 Codex 지적(R1)은 iter5 P3 수리(8476be4d3f)에서 생겼고 iter6과 iter7 감사가 모두 놓친 것이다. 이번 델타에서 생긴 결함은 아니다.

점수: 74/100 — Functionality 70 FAIL(카드 diff로 필수 CI 테스트가 빨갛게 됨) · Security 90 · Craft 72 · Consistency 70. Functionality가 반드시 통과해야 하는 차원이라서 전체 판정은 FAIL이다.

## Codex 호출 기록

- 호출: `mcp__moai__codex_audit` 한 번만 했다(mode `native`, target `baseBranch`, project_root `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702`, delta_base 없음). 리뷰 범위는 서버가 정했다: `main (merge base 9b632c1eb3544cef5f275e502922a375e48d20aa)`. 서버 빌드는 `04d8c27a7e`, 응답 모델은 `gpt-6.1-sol`이다.
- 결과: 리뷰가 실제로 실행됐고(거부 아님) verdict `fail`이 나왔다. 재호출은 하지 않았다.
- 영수증: **`rcpt-690d37c6f81b6c1daa5c7050`**. 이전 영수증은 재사용하지 않았다.
- Codex 지적(원문 요약): [P2] `internal/cli/update_destructive_registry.go:154` — 삭제 호출이 `removeConfigRootMoaiHooks`로 옮겨졌는데 등록은 `ensureSettingsEnvInConfigRoot`에 남아 있다. 그래서 `TestDestructiveTargetRegistry_CoversAllSites`가 stale row로 실패할 것이다. 등록 함수를 바꾸고, 검사기가 `root.OpenRoot("hooks")`로 얻은 `hooksRoot.RemoveAll`도 잡도록 고쳐야 한다. Codex 자신은 테스트를 돌리지 않은 정적 검토라고 밝혔다. 나는 아래 E4에서 실행해 재현했다.
- Codex가 낸 지적은 이 하나뿐이다.

## B2(이전 쓰기 경로 충돌)를 어떻게 판정했나

`moai decision read --scope card:t1702`의 기록 d-20261106T055331Z-c28f(digest e01040700705)를 직접 읽었다. 원문: "운영자 결정 2026-11-06 (리더 AskUserQuestion 원문 답 "설치는 쓰고 판정만 읽기 (권장)"): 설치·update가 프로필 루트에 쓰는 기존 설계는 유지. 프로젝트 사본 제거 가능 판정(SafeToRemoveProjectCopy/CheckProfileReach)과 doctor 프로필 도달 판정만 읽기 전용(os.Root 읽기, 쓰기·디렉터리 생성 없음)."

그래서 install·prune·update 설정 정리·hooks/moai 제거 같은 기존 쓰기 경로가 남아 있는 것은 결함이 아니라 결정된 상태로 봤다. 판정 대상은 두 가지였다. 판정 함수가 읽기 전용인지, 그리고 이번 델타가 프로필 루트에 새 쓰기를 더하는지다. 델타의 비테스트 변경은 `profile_reach.go`뿐이다. 여기에 들어간 것은 빈 seam 호출, 핸들 `Stat`, `os.SameFile`, 문서 수정이고 쓰기는 없다(E1). B2는 닫힌 것으로 판정한다.

## 결함 목록

| id | 심각도 | 분류 | 위치 | 요약 |
|---|---|---|---|---|
| R1 | BLOCKING | in-diff, 재현, Codex 출처. iter5 P3 수리(8476be4d3f)에서 생김 | `internal/cli/update_destructive_registry.go:154` (삭제 호출 `internal/cli/update.go:1180`, 검사기 `internal/cli/update_destructive_registry_test.go:196-223`) | 레지스트리 행이 삭제 호출이 없는 `ensureSettingsEnvInConfigRoot`를 가리켜 `TestDestructiveTargetRegistry_CoversAllSites`가 실패한다. 실제 삭제(`hooksRoot.RemoveAll("moai")`)는 핸들을 `root.OpenRoot`로 얻었기 때문에 검사기의 `os.OpenRoot` 전용 탐지에 보이지 않는다 |
| N5 | MINOR | optional, in-diff | `internal/userassets/profile_reach.go:71` 대 `:116` | 새 주석은 `Root`가 "읽지 않은 항목이면 프로필 디렉터리 자체"라고 쓴다. 그런데 "profile directory changed since it was listed" 행(`:116`)은 읽지 않은 항목인데도 skills·agents 루트 `dir`을 넣는다. N2 정정이 한 분기에서 아직 맞지 않는다 |
| N6 | MINOR | optional, in-diff | `internal/userassets/profile_reach_judgment_t1702_test.go:297-306` | N4 테스트의 seam은 `os.Symlink`가 실패하면 `t.Errorf`로 테스트를 실패시킨다. 이 파일의 다른 링크 테스트는 `t.Skipf`로 건너뛴다. 링크를 만들 수 없는 Windows 환경에서는 거짓 FAIL이 난다. release multi-OS 게이트가 windows-latest에서 전체 스위트를 돌린다(`release-pr-multi-os.yml:316`). GitHub 러너는 관리자 권한이라 실제로는 통과할 가능성이 높지만, 실행해서 관측하지는 않았다 |

iter7 항목 처리:

| iter7 | 상태 | 근거 |
|---|---|---|
| G1(게이트) | 닫힘 | 영수증 `rcpt-690d37c6f81b6c1daa5c7050`. 다만 결과가 fail이고, 그 지적은 R1로 재현됐다 |
| N1 | 닫힘 | `TestCheckProfileReach_AStoreEntryThatCannotBeInspectedIsUnknown` 통과. 분기를 지운 변이에서 RED(E3) |
| N2 | 대부분 닫힘 | 머리말이 c28f 원문과 맞고, godoc이 저장소 나열과 Unknown 행을 설명한다. 남은 한 분기는 N5 |
| N3 | 수용 유지 | 끊긴 링크는 런처의 `os.Stat`도 실패하므로 시작할 수 없는 항목이다(iter7 판단 유지) |
| N4 | 닫힘 | `os.SameFile(fi, f.Stat())`. 이 비교를 지운 변이에서 N4 테스트가 RED(E3). 비교를 뒤집어 항상 거부하게 한 변이에서는 Same 기대 테스트 8개가 RED라서, 정상 파일을 체계적으로 거부하는 퇴행은 darwin에서 테스트가 잡는다(E3) |

### 델타에 대해 물은 것

- **SameFile이 정상 파일을 잘못 거부하나.** `root.Lstat`(nofollow fstatat)와 열린 핸들의 `fstat`은 일반 파일이면 같은 dev/ino를 돌려준다. darwin에서는 Same 기대 테스트가 전부 통과했다(E2). Windows는 `os/root_windows.go:211` `rootStat`과 `os/stat_windows.go:108` `statHandle`이 둘 다 핸들에서 `GetFileInformationByHandle`로 vol/idx를 채우고, `types_windows.go:287` `loadFileId`는 path가 비어 있으면 이미 적재된 것으로 본다. 따라서 vol/idx 비교가 성립한다(Go 1.27.3 소스를 읽은 결과이고 실행 관측은 아님). 체계적 오거부가 날 수 있는 곳은 stat과 fstat의 inode가 다를 수 있는 오래된 overlayfs(xino 이전 커널) 정도다. 그런 곳에서도 결과는 Unknown이라 사본을 남기는 보수적인 방향이다.
- **seam이 운영에서 해로운가.** 패키지 내부 `var afterReachLstat = func(dir, rel string) {}`이다. 비테스트 코드에서 대입하는 곳이 없다(E1 grep). 대입하는 테스트는 `t.Parallel()`을 쓰지 않고, `CheckProfileReach`를 부르는 병렬 테스트도 없다. 그래서 seam을 바꾸는 동안 경합은 없다. 무해하다.
- **문서 주석이 정확한가.** 머리말은 결정 원문과 일치한다. godoc의 "plain files, hidden names and dangling links give no row"는 코드(`:127-136`)와 일치한다. `Root` 필드 주석은 한 분기와 어긋난다(N5).
- **잘못된 호출자가 할 수 있는 일**(판정에는 아직 비테스트 호출자가 없다, E1). (1) `err != nil`을 "프로필 없음"으로 읽고 지우는 것. 다만 오류일 때 행이 nil이라 `SafeToRemoveProjectCopy`를 그대로 쓰면 keep이 된다. (2) `want`에 프로젝트 사본이 아닌 해시(매니페스트 기록값 등)를 넘겨 거짓 Same을 만드는 것. (3) 에이전트 자산에 skills slug를 넘기거나 install과 다른 `rel`을 써서 엉뚱한 파일을 비교하는 것. (4) 이름 있는 프로필만 보므로 기본 세션 루트(`~/.claude`)를 따로 확인하지 않는 것. (5) 판정과 제거 사이에 생긴 프로필이나 바뀐 내용을 다시 판정하지 않는 것. 시그니처는 이것들을 막지 않는다.

## 필요한 조치

- **R1(필수)**: `destructiveTargetRegistry`의 해당 행 `Function`을 `removeConfigRootMoaiHooks`로 바꾼다(Sites 1, Protection 문구 유지). 동시에 `rootHandleNames`(`update_destructive_registry_test.go:196-223`)가 이미 루트 핸들인 식별자에서 `x.OpenRoot(...)`로 얻은 변수도 루트 핸들로 인정하게 한다. 둘 중 하나만 고치면 "stale row"나 "unregistered site" 가운데 하나로 다시 실패한다. 수리 뒤 `go test ./internal/cli -run 'TestDestructiveTargetRegistry_' -count=1 -v`로 PASS를 확인하고, 이 테스트가 디스크 소스를 읽는다는 점을 감안해 별도 트리에서 반증 주입을 한 번 해 본다(파일 주석 §C.4).
- N5, N6은 선택이다. N5는 주석을 "the root that was read, or that would have been read"로 바꾸거나 `:116` 행의 Root를 `profileDir`로 맞추면 된다. N6은 seam 2단계 실패 시 `t.Skipf`로 처리하거나 링크 생성 가능 여부를 먼저 확인하면 된다.
- 재감사 범위: R1 델타와 이 테스트 하나. 이번 codex 호출로 이 감사 단계의 첫 번째 호출이 쓰였다(d-20261106T080039Z-305c 기준 1/2).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. HEAD는 cf141b5409dd38b9a95c9948a3d1836771db7651이고, 이 워크트리에서 감사했다.
2. 델타는 프로필 루트에 새 쓰기를 더하지 않는다. seam은 운영 코드에서 대입되지 않는다.
3. 판정 테스트 17개, launcher parity, config, hooks 제거 테스트, 카드가 건드린 hook 테스트가 통과한다.
4. N1·N4 수리는 변이로 고정돼 있고, 항상 거부하는 변이도 잡힌다.
5. `TestDestructiveTargetRegistry_CoversAllSites`가 HEAD에서 실패한다(R1).
6. gofmt와 go vet(darwin 네 패키지, windows 세 패키지)은 깨끗하다.
7. 필수 codex 게이트는 영수증으로 채워졌고, Codex verdict는 fail이다.

### Evidence

**E0 — 트리**
```
git rev-parse HEAD                → cf141b5409dd38b9a95c9948a3d1836771db7651
git rev-parse --show-toplevel     → /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702
moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702
  → "config_status": "ok", "model": "codex", codex gate "required" (config.gates), claude "off", glm "off",
    "cross_model_active": true, "enforced_required": ["codex"]
```

**E1 — 델타와 seam**
```
git diff bee9cfb001 cf141b5409 -- internal
  → 비테스트 변경은 internal/userassets/profile_reach.go뿐: 머리말 문구, `var afterReachLstat = func(dir, rel string) {}`,
    Root 필드 주석, godoc, `afterReachLstat(dir, rel)`,
    `if hs, herr := f.Stat(); herr != nil || !os.SameFile(fi, hs) { return ReachUnknown, "the opened file is not the one that was inspected" }`
git grep -n "afterReachLstat\|CheckProfileReach(\|SafeToRemoveProjectCopy(" cf141b5409 -- internal
  → 비테스트: profile_reach.go:58,62,87,181,229 (정의·호출부 자기 자신뿐). 대입은 profile_reach_judgment_t1702_test.go:296-307뿐
```

**E2 — 대상 테스트(실제 코드)**
```
GOCACHE=…/gocache-t1702 go test ./internal/userassets -run 'TestCheckProfileReach_|TestSafeToRemoveProjectCopy_|TestOpenRegularForRead_' -count=1 -v
  → --- PASS ×17 (…AStoreEntryThatCannotBeInspectedIsUnknown, …ACopySwappedForAnInRootLinkAfterTheLstatIsUnknown 포함)
    ok  	github.com/modu-ai/moai-adk/internal/userassets	0.621s
GOCACHE=… go test ./internal/config -run 'Profile|UserRoot' -count=1          → ok  	github.com/modu-ai/moai-adk/internal/config	0.449s
GOCACHE=… go test ./internal/profile -run 'TestLauncherProfileDirName' -count=1 -v
  → --- PASS: TestLauncherProfileDirName_MatchesTheLauncherRule (0.00s) / ok … internal/profile	0.817s
GOCACHE=… go test ./internal/cli -run 'TestRemoveConfigRootMoaiHooks_' -count=1 -v
  → --- PASS ×3 (…ASwapToALinkAfterTheCheckNeverReachesTheExternalTree, …ASwapToAnInRootLinkNeverReachesAnotherTree,
    …StillRemovesARealDirectoryAndRefusesALink) / ok … internal/cli	4.658s
GOCACHE=… go test ./internal/hook -run 'TestUserRootSlugDirsMatchUserassetsRoots|TestUserRootZoneFormsCoverProfileRoots|TestUserRootFormsEmitNamespacedForm|TestUserRootFormTrackedScoping|TestZoneUserRootCovered_' -count=1
  → ok  	github.com/modu-ai/moai-adk/internal/hook	1.185s
```

**E3 — 변이(`-overlay`, 원본은 스크래치패드 mut8-t1702/, 트리 미수정)**
```
m_nosame   (`!os.SameFile(fi, hs)` → `hs == nil`)
  --- FAIL: TestCheckProfileReach_ACopySwappedForAnInRootLinkAfterTheLstatIsUnknown
      profile_reach_judgment_t1702_test.go:313: … want Unknown, got [{Profile:alpha … State:same …}]   → 잡힘
m_non1     (검사 불가 항목의 default 분기에서 행 추가 삭제)
  --- FAIL: TestCheckProfileReach_AStoreEntryThatCannotBeInspectedIsUnknown
      profile_reach_judgment_t1702_test.go:279: want alpha Same and loop Unknown, got [{Profile:alpha … State:same …}]   → 잡힘
m_alwaysreject (`!os.SameFile` → `os.SameFile`, 정상 파일 항상 거부)
  --- FAIL ×8 (SameCopyInEveryProfileIsSafeToRemove, AbsentKeepsTheProjectCopy, DifferingContentKeepsTheProjectCopy,
      LaunchableProfilesTheSharedFilterDropsAreUnknown, AStoreEntryThatCannotBeInspectedIsUnknown,
      ACopySwappedForAnInRootLinkAfterTheLstatIsUnknown, RootAndFileShapesThatMustStayUnknown, AgentsRootAndFixedSlugs)   → 잡힘
```
(m_non1 출력에서 loop 항목의 실제 사유: `stat …/loop: too many levels of symbolic links`. ELOOP는 ErrNotExist가 아니므로 Unknown이 된다.)

**E4 — R1 재현(실제 코드)**
```
GOCACHE=… go test ./internal/cli -run 'TestDestructiveTargetRegistry_' -count=1 -v
=== RUN   TestDestructiveTargetRegistry_CoversAllSites
    update_destructive_registry_test.go:75: stale registry row: internal/cli/update.go ensureSettingsEnvInConfigRoot is registered but the source has no destructive call site there
    update_destructive_registry_test.go:94: scanned 15 (file, function) pair(s) across internal/cli/update and internal/cli/update*.go; registry has 16 row(s)
--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.12s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.584s
```
카드가 건드린 cli 테스트를 이름으로 골라 한 번에 돌린 결과(`-run 'TestBinaryLag_|TestBuildIdentity_|TestDoctorUserInstall|TestDoctorProjectVsLock|TestEnsureGlobalSettingsEnv_|TestUpdateSettingsEnvHooksSymlinkRefused|TestRunUserAssetUpdatePhase_|TestCheckProfileSkillReach|TestDoctorSkillReach|TestUpdateGlobalSettingsEnv_|TestUserAssetRootsLine|TestDoctor.*Golden|TestDestructive'`)에서 FAIL은 이 하나뿐이었다(`FAIL … internal/cli	41.186s`).
출처 추적:
```
git log --oneline -S "func removeConfigRootMoaiHooks" main..cf141b5409 -- internal/cli/update.go   → 8476be4d3f … (audit iter5 P3)
git log --oneline -S "ensureSettingsEnvInConfigRoot\", Sites" main..cf141b5409 -- …registry.go   → 8f600fca00 …
git log --oneline -S "hooksRoot.RemoveAll" main..cf141b5409 -- internal/cli/update.go            → bee9cfb001 …
```
검사기 `rootHandleNames`는 `os.OpenRoot` 호출(`pkg.Name == "os"`)과 `*os.Root` 필드만 루트 핸들로 인정한다(`update_destructive_registry_test.go:196-223`). 그래서 `hooksRoot, err := root.OpenRoot("hooks")`(`update.go:1168`)의 `hooksRoot.RemoveAll`(`:1180`)은 세지 않는다. 결과의 "scanned 15"에 `removeConfigRootMoaiHooks` 미등록 지적이 없는 것도 이와 맞는다.

**E5 — 정적 검사**
```
gofmt -l internal/userassets internal/config internal/profile internal/cli                → 출력 없음
GOCACHE=… go vet ./internal/userassets ./internal/config ./internal/profile ./internal/cli → 출력 없음
env GOOS=windows GOCACHE=… go vet ./internal/userassets ./internal/config ./internal/profile → 출력 없음
```

**E6 — 결정 원문**: `moai decision read --scope card:t1702 > …/scratchpad/decisions-t1702-iter8.txt` 뒤 grep으로 확인했다. 47-49행에 위 B2 절의 원문이 있다.

### Baseline-attribution
모든 측정은 이 실행에서 워크트리 트리 cf141b5409에 대해 했다. Go는 go1.27.3(`go env GOVERSION`)이고, 비공개 `GOCACHE=/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/671835d7-2501-4df7-8d56-57dfa4e7b2d0/scratchpad/gocache-t1702`를 썼다. 이번 실행에서 캐시 오류는 없었다. 테스트 바이너리는 전부 이 트리 소스로 빌드했다. 설치된 moai는 측정에 쓰지 않았고 기록 열람(`decision read`)과 계획 조회(`verify audit-plan`)에만 썼다. Codex 리뷰는 서버 빌드 04d8c27a7e가 수행했고, 리뷰 범위는 merge base 9b632c1eb3 기준 이 트리였다.

### Gaps
- `mcp__moai__audit_multi`와 그 결과 대조(`--result-file`)는 실행하지 않았다(지시: codex_audit 한 번만). 계획의 `enforced_required: [codex]`는 단독 `codex_audit` 영수증으로 대신했고, 계획 표면의 convergence_check는 관측하지 않았다. 디지스트: backend codex, gate required, verdict fail, receipt rcpt-690d37c6f81b6c1daa5c7050.
- `internal/cli`, `internal/hook` 전체 스위트와 `go test ./...`는 실행하지 않았다(지시). 그래서 카드 diff가 다른 드리프트 가드를 더 깨뜨리는지는 이름으로 고른 테스트 범위에서만 확인했다. R1도 이전 회차들이 같은 이유로 놓친 것이다.
- golangci-lint, 커버리지 측정, `/verify` 게이트는 실행하지 않았다.
- Linux·Windows 실행은 하지 않았다. SameFile과 O_NONBLOCK의 Windows 동작, overlayfs 동작은 소스를 읽어 판단했다. N6의 Windows 거짓 FAIL 가능성도 관측하지 않았다.
- R1을 고친 뒤의 반증(검사기 확장 후 미등록 사이트 주입)은 하지 않았다. 검사기가 디스크 소스를 읽어서 overlay로는 주입할 수 없다.
- M3(열거 뒤 링크로 바뀌는 경우)는 여전히 결정적 테스트가 없다(iter7과 같음).
- main 트리에서 같은 테스트가 통과하는지는 직접 돌리지 않았다. R1을 카드 탓으로 본 근거는 위 커밋 추적과 diff다.

### Residual-risk
- 판정 함수에 호출자가 없어서, 실제 제거 경로의 안전성은 배선 카드가 위 다섯 가지를 지키는지에 달려 있다.
- 오래된 overlayfs 같은 환경에서는 N4 수리 때문에 정상 Same이 Unknown이 될 수 있다. 보수적인 방향이지만, 그런 환경에서는 프로젝트 사본이 계속 남는다.
- 정크 디렉터리가 있는 환경에서는 프로젝트 사본이 무기한 남는다(iter7과 같음). 사용성 문제이고 안전 문제는 아니다.
- 변이 원본과 결정 기록 사본은 세션 스크래치패드에만 있다. 결정을 바꾼 출력 줄은 이 파일에 옮겼다.

## 반복 이력

- iter6: FAIL @732fce152d — B1, B2, S1, S2, M1~M4.
- iter7: FAIL @bee9cfb001 — 코드 BLOCKING 없음. 필수 codex 게이트 미충족(stale_server, 영수증 없음).
- iter8(이번): FAIL @cf141b5409 — 게이트 충족(rcpt-690d37c6f81b6c1daa5c7050). Codex P2를 재현해 BLOCKING R1로 올렸다. R1은 iter5 P3 수리(8476be4d3f)에서 생겼고 iter6·iter7이 놓쳤다. 이번 델타에서 생긴 것은 아니다. 직전 회차(iter7)에 수리에서 생긴 BLOCKING이 없었으므로 수리 중단 규칙(연속 2회)은 걸리지 않는다. N1·N4는 닫혔고, 새 MINOR는 N5·N6이다.
