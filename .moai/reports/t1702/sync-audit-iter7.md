auditor-model: claude-opus-5-5

# t1702 sync 감사 iter7 (재감사, 수리 델타 732fce152d..bee9cfb001)

verdict: FAIL
audited_sha: bee9cfb001c92121b5f7f96af6258e7c2baf1345

## 요약

판정은 FAIL이고, 원인은 코드 결함이 아니라 게이트다. 이 트리의 `workflow.audit.gates.codex`는 `required`인데, 이번 감사에서 허용된 한 번의 `codex_audit` 호출이 리뷰가 시작되기도 전에 `stale_server`로 거부됐다. 판정은 나오지 않았고 영수증도 발급되지 않았다. 필수 게이트가 비어 있으면 PASS를 낼 수 없다(sync-auditor 규칙: 필수 게이트 미충족은 PASS를 막는 Gap).

코드만 보면 BLOCKING 결함은 찾지 못했다. iter6의 B1·S1·S2·M1·M2는 모두 닫혔고, B2는 운영자 결정이 반영된 항목으로 처리했다. 새로 찾은 것은 MINOR(선택) 4건뿐이다. 따라서 코드는 바꿀 필요가 없다. MCP 서버를 다시 연결한 뒤 Codex 호출 한 번으로 게이트만 채우면 된다. 이 호출은 상시 규칙 d-20261106T080039Z-305c("리뷰 시작 전 stale_server 등으로 거부돼 영수증이 생기지 않은 호출은 횟수에 넣지 않는다")에 따라 한도에 들어가지 않는다고 읽힌다. 다만 이번 지시가 재호출을 금지했으므로 나는 다시 호출하지 않았다.

점수(코드 기준): 89/100 — Functionality 90 · Security 90 · Craft 85 · Consistency 88. 반드시 통과해야 하는 두 차원(Functionality, Security)은 코드 기준으로 통과다. 그래도 전체 판정은 게이트 때문에 FAIL이다.

## Codex 호출 기록

- 호출: `mcp__moai__codex_audit` 한 번(mode `native`, target `baseBranch`, delta_base `732fce152ddb9d2a78903c63e17eaf5d3750c46f`, project_root `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702`). 트리를 바꾸기 전에 먼저 호출했다.
- 결과(원문 그대로):
  `{"_server":{"commit":"646955327e","version":"v3.2.0-rc.42"},"error":"codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves","stale_server":true}`
- 영수증: 없음. 이전 영수증(rcpt-6fed68e2ded5ad1a86a50d1e 등)은 재사용하지 않았다.
- Codex 지적 사항: 없음. 리뷰가 실행되지 않았다.
- 재호출: 하지 않았다(지시대로).

## B2를 어떻게 판정했나

`moai decision read --scope card:t1702`의 기록 d-20261106T055331Z-c28f를 직접 읽었다. 원문은 "설치·update가 프로필 루트에 쓰는 기존 설계는 유지. 프로젝트 사본 제거 가능 판정(SafeToRemoveProjectCopy/CheckProfileReach)과 doctor 프로필 도달 판정만 읽기 전용(os.Root 읽기, 쓰기·디렉터리 생성 없음)."이다. 그래서 기존의 쓰기 경로(install, prune, update 설정 정리, hooks/moai 제거)가 남아 있는 것은 결함이 아니라 결정된 상태로 봤다. 확인한 것은 두 가지다. 하나는 이 브랜치의 새 판정이 읽기 전용인지, 다른 하나는 델타가 프로필 루트에 새 쓰기를 더하지 않는지다.

- 델타에서 쓰기 계열 기본 연산이 걸린 비테스트 코드는 둘이다. `profile_reach.go:194`의 `root.OpenFile(rel, os.O_RDONLY|syscall.O_NONBLOCK, 0)`는 읽기 전용 열기이고 O_CREAT가 없다. `update.go:1180`의 `hooksRoot.RemoveAll("moai")`는 기존 `root.RemoveAll("hooks/moai")`를 같은 대상으로 좁힌 것이다. 새 호출 위치도 없다(`removeConfigRootMoaiHooks(` 호출부는 732fce152d와 bee9cfb001 모두 update.go 한 곳).
- 새 판정 경로(`CheckProfileReach`, `readProfileCopy`, `openRegularForRead`)는 `os.ReadDir`, `os.Lstat`, `os.Stat`, `os.OpenRoot`, `root.Lstat`, 읽기 전용 열기만 쓴다. 쓰기도 디렉터리 생성도 없다.
- 결론: B2는 닫혔다(결정 반영). iter6과 이전 회차 요약의 "프로필 루트에는 아무것도 쓰지 않는다"는 표현은 결정보다 넓었으므로 기준으로 쓰지 않았다.

## 결함 목록

| id | 심각도 | 분류 | 위치 | 요약 |
|---|---|---|---|---|
| G1 | 게이트(PASS 차단, 코드 결함 아님) | — | `.moai/config/sections/workflow.yaml:38` | 필수 codex 게이트 미충족: 유일한 호출이 stale_server로 리뷰 전에 거부됐고 영수증이 없다 |
| N1 | MINOR | optional, in-diff | `internal/userassets/profile_reach.go:127-129` | 검사할 수 없는 저장소 항목(Stat 오류가 not-exist가 아닌 경우)이 Unknown이 되는 분기를 고정하는 테스트가 없다. 변이 B(그 분기를 지움)가 살아남았다 |
| N2 | MINOR | optional, in-diff | `internal/userassets/profile_reach.go:8-10, 69-70, 64, 123, 128` | 문서 불일치: godoc은 아직 "the shared junk-filtered enumeration"만 읽는다고 쓰고, 파일 머리말은 철회된 "nothing is written" 중계 문구를 남겼다. `ProfileReach.Root`는 "skills\|agents root"라고 적혀 있지만 열거 밖 행에는 프로필 디렉터리 자체가 들어간다 |
| N3 | MINOR | optional, in-diff | `internal/userassets/profile_reach.go:125-126` | 끊긴 링크는 행을 만들지 않는다. 마운트가 빠진 외부 볼륨을 가리키는 프로필 링크는 그 순간 시작할 수 없으니 일관된 판단이지만, 볼륨이 다시 붙으면 사본 없이 시작되는 프로필이 생길 수 있다 |
| N4 | MINOR | optional, in-diff | `internal/userassets/profile_reach.go:160-171, 194` | Lstat(정규 파일)와 열기 사이에 같은 루트 안의 다른 정규 파일을 가리키는 링크로 바뀌면 `root.OpenFile`이 따라간다. 핸들에서는 정규 파일 여부만 확인하고 동일성은 보지 않는다. 읽기 전용이고, 프로필 루트에 쓸 수 있는 행위자가 있어야 하는 경로다 |

iter6 항목의 처리:

| iter6 | 상태 | 근거 |
|---|---|---|
| B1 | 닫힘 | 저장소를 직접 나열하고, 런처가 시작할 수 있지만 공유 열거가 뺀 항목(링크 디렉터리, 공백 이름, 예약 이름)은 Unknown이 된다. 테스트가 4행(alpha same, 나머지 unknown)과 `SafeToRemoveProjectCopy == false`를 단언한다. 변이 A(분기를 끔)는 잡혔다 |
| B2 | 닫힘(결정 반영) | 위 절 |
| S1 | 닫힘 | 루트 링크, 루트 안 파일 링크, 상한 초과, FIFO 테스트가 통과한다. 변이 C(핸들 검사를 지움)는 잡혔다. 기록이 주장한 세 변이는 다시 돌려 보지 않았다(Gaps) |
| S2 | 닫힘 | `req-owr-005-revision.md:16, 28-29`가 쓰기 금지를 제거·이전 경로로 한정하고, 설치 쪽 쓰기는 c28f로 존치된다고 밝힌다 |
| M1 | 닫힘 | 루트 안 링크로 바꾸는 테스트와 대조 테스트(실제 디렉터리 제거, 링크 거부)가 모두 통과한다 |
| M2 | 닫힘 | 비차단 열기 뒤 핸들에서 Stat한다. FIFO 테스트가 통과하고, 변이 C는 잡혔다 |
| M3 | 부분 | Lstat 재확인은 들어갔지만 결정적 테스트는 없다(알려진 Gap). 재확인과 OpenRoot 사이의 창은 남는다 |
| M4 | 수용 유지 | iter6 판단을 그대로 둔다 |

### 지시에서 물은 새 결함 후보

- **시작할 수 있지만 읽을 수 없는 항목.** 열거에 든 디렉터리를 읽을 수 없으면 `readProfileCopy`가 Unknown을 돌려준다. 열거 밖의 링크가 접근 불가 위치를 가리키면 기본 분기가 Unknown 행을 만든다(ENOTDIR도 ErrNotExist가 아니므로 여기로 간다). 행이 잘못 생기는 경우는 찾지 못했다. 행이 잘못 빠질 수 있는 경우는 끊긴 링크(N3)뿐이다. 기본 분기는 테스트가 없다(N1).
- **LauncherProfileDirName과 런처의 차이.** `isValidProfileName`(`internal/profile/profile.go:207-213`)과 본문이 같다. 차이는 빈 문자열 하나이고 의도된 것이며 테스트가 고정한다. 대조 테스트는 15개 이름만 보므로, 런처 규칙이 좁아지면(새 금지 문자 등) 잡지 못한다. 그 경우 config 쪽이 더 넓어져 Unknown이 늘어나는 보수적인 방향이다. 런처 규칙이 넓어지면(점으로 시작하는 이름 허용 등) `.hidden`, `.`, `..`가 잡는다. 런처의 `default`는 디렉터리를 쓰지 않는데, 저장소에 `default`라는 디렉터리가 있으면 판정은 그것을 프로필로 읽는다. 결과가 Absent/Unknown이면 사본을 남기는 쪽이라 해롭지 않고, 이 동작은 이번 델타 이전부터 있었다.
- **Windows의 `syscall.O_NONBLOCK`.** `syscall/types_windows.go:50`에 `O_NONBLOCK = 0x00800`이 정의돼 있고, `internal/syscall/windows/at_windows.go`의 `Openat`은 이 비트를 검사하지 않으므로 무시된다(소스 기준이며 실행 관측은 아니다). `GOOS=windows go vet`의 네 패키지 출력은 없음.
- **Windows의 SameFile.** `os/root_windows.go:211-244`의 `rootStat`은 `statHandle`로 핸들에서 fileStat을 만든다. `os/types_windows.go:287-293`의 `loadFileId`는 path가 비어 있으면 이미 적재된 것으로 본다. 따라서 vol/idx로 비교되고, 정상 디렉터리에서 SameFile이 항상 거짓이 되는 퇴행은 소스상 보이지 않는다. Windows에서 실행해 관측하지는 않았다(Gaps).
- **정크 디렉터리가 모든 사본을 붙잡는 결과.** 보수적인 결과로 받아들였다. 예약 이름 `__no_such_profile__`을 만드는 비테스트 코드는 없다(`git grep`은 상수 정의 한 줄만 찾는다). 운영자 판정("없거나 불확실하면 프로젝트 사본을 둔다")과 맞는다. 배선 카드는 Unknown 사유를 사용자에게 보여 주고 정리 경로(t1643)를 안내해야 한다.
- **hooks 제거가 정상 디렉터리를 지우지 못하는 퇴행.** 대조 테스트 `TestRemoveConfigRootMoaiHooks_StillRemovesARealDirectoryAndRefusesALink`가 통과한다. `hooks/moai`가 일반 파일인 경우도 RemoveAll로 기존과 같이 지워진다.

### 잘못된 호출자가 할 수 있는 일(판정은 아직 호출자가 없다)

`git grep`으로 확인한 결과, 비테스트 호출자가 없다. 배선 카드가 지켜야 할 점은 다음과 같다.
1. `err != nil`을 "프로필 없음"으로 읽고 지우면 저장소를 읽지 못했을 때 제거가 허가된다.
2. `want`를 프로젝트 사본이 아닌 다른 내용으로 계산하면 거짓 Same이 나온다.
3. 에이전트 자산에 skills slug를 넘기거나 install과 다른 `rel` 배치를 쓰면 엉뚱한 파일을 비교한다.
4. 이 판정은 이름 있는 프로필만 본다. 기본 세션(`~/.claude`)에 사본이 닿는지는 따로 확인해야 한다.
5. 판정과 제거 사이에 새로 생긴 프로필은 판정에 들어가지 않는다(TOCTOU). 제거 직전에 다시 판정해야 한다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. HEAD는 bee9cfb001c92121b5f7f96af6258e7c2baf1345이고, 이 워크트리에서 감사했다.
2. 새 테스트와 기존 판정 테스트, hooks 제거 테스트, 런처 대조 테스트가 모두 통과한다.
3. gofmt와 go vet(darwin, windows, linux)이 깨끗하다.
4. 변이 A(B1 분기 끔)와 C(핸들 검사 삭제)는 테스트가 잡고, 변이 B(검사 불가 분기 삭제)는 살아남는다.
5. 델타는 프로필 루트에 새 쓰기를 더하지 않는다.
6. 필수 codex 게이트가 미충족이다.

### Evidence
- `git rev-parse HEAD` → `bee9cfb001c92121b5f7f96af6258e7c2baf1345`; `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702`
- `GOCACHE=…/gocache-t1702 go test ./internal/userassets -run 'TestCheckProfileReach_|TestSafeToRemoveProjectCopy_|TestOpenRegularForRead_' -count=1 -v` → 15개 `--- PASS`(FifoInPlaceOfTheCopy, OpenRegularForRead_Fifo, LaunchableProfilesTheSharedFilterDropsAreUnknown, RootAndFileShapesThatMustStayUnknown 포함), `ok  	github.com/modu-ai/moai-adk/internal/userassets	0.706s`
- 같은 명령에 `-count=3 -race` → `ok  	github.com/modu-ai/moai-adk/internal/userassets	1.690s`
- `go test ./internal/profile -run 'TestLauncherProfileDirName' -count=1 -v` → `--- PASS: TestLauncherProfileDirName_MatchesTheLauncherRule (0.00s)` / `ok  	github.com/modu-ai/moai-adk/internal/profile	0.271s`
- `go test ./internal/cli -run 'TestRemoveConfigRootMoaiHooks_' -count=1 -v` → `--- PASS: …ASwapToALinkAfterTheCheckNeverReachesTheExternalTree`, `--- PASS: …ASwapToAnInRootLinkNeverReachesAnotherTree`, `--- PASS: …StillRemovesARealDirectoryAndRefusesALink`, `ok  	github.com/modu-ai/moai-adk/internal/cli	2.426s`
- `go test ./internal/config -run 'Profile|UserRoot' -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/config	0.213s`
- `gofmt -l internal/userassets internal/config internal/profile internal/cli` → 출력 없음
- `go vet` 네 패키지 → 출력 없음; `env GOOS=windows … go vet` 네 패키지 → 출력 없음; `env GOOS=linux … go vet ./internal/userassets ./internal/config ./internal/profile` → 출력 없음
- 변이 A(`go test -overlay …/ovA.json …`) → `--- FAIL: TestCheckProfileReach_LaunchableProfilesTheSharedFilterDropsAreUnknown … rows = [{Profile:alpha … State:same …}], want exactly map[__no_such_profile__:unknown alpha:same has space:unknown linked:unknown]` / `FAIL`
- 변이 B(`…/ovB.json`) → `ok  	github.com/modu-ai/moai-adk/internal/userassets	0.479s` (생존)
- 변이 C(`…/ovC.json`) → `--- FAIL: TestOpenRegularForRead_FifoIsRefusedWithoutBlocking … a FIFO must be refused` / `FAIL`
- `git diff 732fce152d bee9cfb001 -G "WriteFile|MkdirAll|Mkdir\(|Rename|RemoveAll|Remove\(|Symlink|Chmod|Create\(|OpenFile" --stat -- internal` → 비테스트 파일은 `internal/cli/update.go`, `internal/userassets/profile_reach.go`뿐이다. 해당 줄은 위 B2 절에 적었다.
- `git grep -n "removeConfigRootMoaiHooks(" bee9cfb001 -- internal` → 비테스트 호출 `internal/cli/update.go:1198` 하나(732fce152d에서는 `:1174` 하나)
- `git grep -n "__no_such_profile__" bee9cfb001 -- "*.go" ":!*_test.go"` → `internal/config/userroots.go:61` 한 줄
- `grep -n -A6 "gates:" .moai/config/sections/workflow.yaml` → `codex: required`
- Codex 호출 결과는 위 "Codex 호출 기록" 절에 원문으로 적었다.

### Baseline-attribution
모든 측정은 이 실행에서 워크트리 트리 bee9cfb001에 대해 했다. Go 1.26.8(`go env GOVERSION`)과 비공개 GOCACHE(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/671835d7-2501-4df7-8d56-57dfa4e7b2d0/scratchpad/gocache-t1702`)를 썼다. 변이는 스크래치패드의 `mut-t1702/`에 둔 overlay 파일로만 적용했고 트리는 바꾸지 않았다. 이번 측정에는 moai 바이너리가 쓰인 것이 없다. 결정 기록은 설치된 `moai decision read`로 읽었는데, 이는 측정이 아니라 기록 열람이다.

### Gaps
- 필수 codex 게이트: 리뷰가 실행되지 않았다(G1). 영수증이 없다.
- `moai verify audit-plan`은 실행하지 않았다. 허용 명령 목록에 없어서 `workflow.yaml`의 gates 블록을 직접 읽는 것으로 대신했다. 따라서 계획 표면에서 확인한 게이트 해석은 없다.
- M1 변이(SameFile 검사 삭제)는 직접 돌리지 않았다. update.go 전체 사본이 필요해서다. RED-before 출력은 progress-r8.md의 기록에만 있다. 테스트 구조상 링크를 따라가 `skills/moai`를 지우는 경로를 막는지는 읽어서 확인했다.
- progress-r8.md가 주장한 S1 변이 세 개(root IsDir, IsRegular, 크기 상한)는 다시 돌리지 않았다. 내가 돌린 변이는 A, B, C다.
- M3(열거 뒤 링크로 바뀌는 경우)와 Lstat→OpenRoot 경쟁은 결정적 테스트가 없다.
- Windows에서 O_NONBLOCK이 무시되는지, SameFile이 맞게 동작하는지는 Go 표준 라이브러리 소스를 읽은 것이고, 실행해 관측하지 않았다. FIFO 테스트는 `!windows`이고, 링크 테스트는 링크를 만들 수 없는 환경에서 건너뛴다.
- 전체 `internal/cli`, `internal/hook` 묶음과 golangci-lint, 커버리지 측정은 실행하지 않았다(지시 범위 밖).
- 공유 빌드 캐시가 지워지는 문제: 비공개 캐시를 써서 이번 실행에는 캐시 오류가 없었다.

### Residual-risk
- Codex 리뷰가 없으므로 Claude 단일 시각의 맹점이 남는다. 이것이 G1이 PASS를 막는 이유다.
- 판정에 호출자가 없어서 실제 제거 경로의 안전성은 배선 카드에 달려 있다(위 다섯 가지).
- 정크 디렉터리가 있는 환경에서는 프로젝트 사본이 무기한 남는다. 사용성 문제이고 안전 문제는 아니다.
- N3/N4 경로는 드물지만 이론상 남는다. 둘 다 읽기 전용 판정의 오판 가능성이지 쓰기 사고가 아니다.

## 필요한 조치

- **G1**(필수): MCP 서버를 다시 연결해(호스트 세션 재시작) 설치본이 서비스되게 한 뒤, 같은 HEAD(bee9cfb001)에서 `codex_audit`를 한 번 호출해 영수증을 받는다. 코드 변경은 필요 없다. Codex가 BLOCKING을 내지 않으면 이 감사의 코드 판단(BLOCKING 없음)과 합쳐 PASS로 볼 수 있다. 호출 횟수 해석(305c에 따라 이번 거부는 횟수에 들지 않음)은 리더가 판단한다.
- N1~N4는 선택이다. 배선 카드로 넘겨도 된다. N2의 문서 정정은 한 줄짜리 수정이다.

## 반복 이력

- iter6: FAIL @732fce152d — B1(실행 가능한 프로필 누락 → 제거 허가), B2(쓰기 경로 결정 미정), S1, S2, M1~M4.
- iter7(이번): FAIL @bee9cfb001 — 코드 BLOCKING 없음, iter6 항목 전부 닫힘(M3는 부분). FAIL의 유일한 원인은 필수 codex 게이트 미충족(stale_server 거부, 영수증 없음). 수리에서 생긴 BLOCKING(repair-born)도 없다.
