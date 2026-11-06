auditor-model: claude-opus-5-5

# 동기화 감사 6회차 — 카드 t1702 (델타 0887e69728..732fce152d)

verdict: FAIL
audited_sha: 732fce152ddb9d2a78903c63e17eaf5d3750c46f

Codex leg: not run (Claude-only audit)

점수: 72/100 (Functionality 60 FAIL · Security 82 · Craft 70 · Consistency 85). 판정은 아래 BLOCKING 두 건 때문에 FAIL이다.

## 요약

iter5의 두 결함은 이번 델타로 닫혔다. F2(P1)는 `foldProfileSlug`로 배포 YAML의 `user-root:claude-skills/`·`user-root:claude-agents/` 항목이 프로필 루트까지 덮게 되었고, Bash 판정(`zoneUserRootCovered`)뿐 아니라 iter5가 직접 보지 못했던 Write/Edit 경로(`checkProtectedZone`)에서도 프로필 루트의 추적 파일이 거부됨을 감사자 탐침으로 확인했다. 대소문자 변형과 `..` 표기도 거부되고, 추적되지 않은 프로필 파일은 그대로 쓸 수 있다. P3는 `os.Root` 안에서의 삭제로 외부 트리 경합이 닫혔고, 이전 방식으로 되돌린 변이를 새 테스트가 잡았다.

새로 들어온 읽기 전용 판정 `CheckProfileReach`는 쓰기를 전혀 하지 않고, 읽은 프로필에 대해서는 모든 의심 상태를 Unknown으로 돌려준다. 그러나 **실제로 실행할 수 있는 프로필을 아예 읽지 않고 넘어가는 경우가 있다**. 심링크로 된 프로필 디렉터리, 공백이 들어간 이름의 프로필은 런처(`internal/profile`)가 받아들여 세션을 띄우는데, 공유 열거(`config.UserProfileDirNames`)는 이들을 조용히 뺀다. 그 결과 다른 프로필이 모두 같은 사본을 가지고 있으면 `SafeToRemoveProjectCopy`가 true를 돌려준다. 운영자 판정("없거나 불확실하면 프로젝트 사본을 둔다")에 정면으로 어긋나며, 이 카드가 고치려는 바로 그 결함(프로필 세션에서 스킬이 보이지 않음)을 다시 만든다. 재현 증거가 있으므로 BLOCKING이다.

또 하나, 6·7라운드가 기록한 설계 충돌(브랜치의 기존 설치·정리 코드가 프로필 루트에 쓰는데 운영자 판정은 "프로필 루트에는 아무것도 쓰지 않는다")은 결정 없이 남아 있다. 이 상태로는 merge-ready를 줄 수 없다.

## 결함 목록

| id | 등급 | 위치 | 요약 |
|---|---|---|---|
| B1 | BLOCKING (in-diff, 재현) | `internal/userassets/profile_reach.go:84-97` (열거 원천 `internal/config/userroots.go:106-108`) | 런처가 띄울 수 있는 프로필(심링크 디렉터리·공백 이름·예약 이름)이 열거에서 빠져 행이 없고, 나머지 프로필이 모두 Same이면 제거 허가가 나온다 |
| B2 | BLOCKING (카드 diff, 결정 대기) | `internal/userassets/install.go:119`, `internal/userassets/remove.go:45,258`, `internal/cli/update.go:1112` | 프로필 루트에 쓰는 기존 경로와 운영자 판정 d-20261106T055331Z-c28f "NOTHING is written to a profile root"의 충돌이 결정 없이 남아 있다 |
| S1 | SHOULD-FIX (in-diff) | `internal/userassets/profile_reach_judgment_t1702_test.go:140-183` | 프로필 루트가 링크인 경우, 루트 안을 가리키는 파일 링크·FIFO, 8 MiB 상한이 테스트로 고정되지 않았다 (변이 2개 생존) |
| S2 | SHOULD-FIX (문서) | `.moai/reports/t1702/req-owr-005-revision.md:26-28` | "모든 프로필 루트에 사본을 넣는 안은 채택하지 않았다"고 쓰지만 이 브랜치가 바로 그 설치를 구현하고 있다; 제안한 "예외 없는 쓰기 금지"가 이 브랜치의 `moai update`를 위반으로 만든다는 사실을 밝히지 않는다 |
| M1 | MINOR (in-diff, 재현) | `internal/cli/update.go:1134-1157` | 검사 뒤 `hooks`를 **루트 안** 링크(`skills`)로 바꾸면 `os.Root`가 따라가 `skills/moai`를 지운다; 주석의 "루트 안에 머무는 링크도 그대로 둔다"는 검사 시점에만 참이다 |
| M2 | MINOR (in-diff) | `internal/userassets/profile_reach.go:126-141` | `Lstat`(정규 파일) 뒤 `root.Open` 사이에 FIFO로 바뀌면 열기에서 무기한 대기한다; 핸들 기준 확인이 아니다 |
| M3 | MINOR (in-diff) | `internal/userassets/profile_reach.go:94,112,121` | 열거 뒤 `<store>/<profile>` 자체가 링크로 바뀌면 `Lstat(dir)`은 마지막 요소만 보므로 외부 트리를 읽는다 (읽기 전용, 해시 비교는 유지) |
| M4 | MINOR (수용) | `internal/cli/update_home_radius_test.go` Shape 2 | iter5 F3: 판별력 없는 Shape 2를 손대지 않았다. Shape 1과 새 TOCTOU 테스트가 변이를 잡으므로 수용 가능 |
| O1 | outside (판정 무관) | `internal/config/userroots.go:107` 를 쓰는 설치·훅 경로 | 같은 열거 탓에 심링크 프로필에는 설치도, 보호도 되지 않는다. 카드 목표(모든 실제 프로필 도달)의 빈틈이며 운영자에게 보여줄 항목이다 |

### 필수 수정

- **B1**: `CheckProfileReach`가 저장소를 직접 나열해, 런처가 쓸 수 있는 항목(`os.Stat` 기준 디렉터리 + `profile.IsValidProfileName` 규칙)인데 공유 필터가 뺀 것마다 `ReachUnknown` 행을 내도록 한다(가져오기 방향이 막히면 같은 규칙을 config 쪽 공유 함수로 둔다). 심링크 프로필 디렉터리와 공백 이름 프로필 각각에 대해 `SafeToRemoveProjectCopy == false`를 단언하는 테스트를 추가한다.
- **B2**: 운영자 결정을 받아 진행 기록에 결정 id로 남긴다. (b)를 택하면 프로필 루트 쓰기를 제거하거나 스위치 뒤로 보낸다. (a)/(c)를 택하면 그 결정을 근거로 S2의 제안 문구를 함께 고친다.
- **S1**: 판별 사례 세 개를 추가한다 — 프로필 `skills`가 링크인 경우, 루트 안 파일 링크, FIFO 또는 8 MiB 초과 파일. 각각 Unknown을 단언한다.
- **S2**: "채택하지 않은 안"을 사실대로 고치고, 이 브랜치의 기존 쓰기 경로와 제안 REQ의 관계를 명시한다.
- **M1(선택)**: `root.OpenRoot("hooks")`로 연 핸들을 `Stat`하고 앞선 `Lstat` 결과와 `os.SameFile`로 비교한 뒤 그 하위 루트에서 `RemoveAll("moai")`를 한다. 고치지 않을 경우 주석을 실제 보장 범위에 맞게 바꾼다.
- **M2(선택)**: 먼저 연 뒤 `f.Stat()`으로 정규 파일 여부를 핸들에서 확인한다(비차단 열기).

## 운영자·리더 질문별 판정

| 항목 | 판정 | 근거 |
|---|---|---|
| (1) 배포 YAML로 Write/Edit/Bash가 프로필 루트에서 막히는가 | 충족 | E3 탐침: Write·Edit 모두 deny, Bash 판정 함수 테스트 PASS |
| (1) 비슷한 slug·잡동사니 이름이 잘못 접히는가 | 아니오 | `claude-skills-evil@work`, `agents-skills@work`, `@work`, `claude-skills@` 모두 미보호 (E1); 접기를 모든 base로 넓힌 변이는 테스트가 잡음 (E5) |
| (1) 쓸 수 있어야 할 것을 보호하는가 | 아니오 | 추적되지 않은 프로필 파일·프로필 `settings.json` 쓰기 허용 (E3) |
| (1) 우회(대소문자, 끝 슬래시, `..`, 심링크 프로필) | 대소문자·`..`·끝 슬래시 막힘. 심링크 프로필은 루트로 열거되지 않아 보호 대상이 아님(O1) | E1, E3, E4 |
| (2) TOCTOU 창이 닫혔는가 | 외부 링크는 닫힘, 루트 안 링크는 열려 있음(M1) | E5 변이 사살, E6 탐침 |
| (2) F1 거부 유지 | 충족 | `TestUpdateSettingsEnvHooksSymlinkRefused` PASS (E2) |
| (2) 시험용 seam이 운영에서 해로운가 | 아니오 | 패키지 내부 변수, 운영 코드에서 대입하는 곳 없음, 기본값은 빈 함수 |
| (3) 쓰기·생성·외부 링크 추적이 있는가 | 쓰기·생성 없음. 읽기 시점 경합으로만 외부를 읽을 수 있음(M3) | 코드 확인, `WritesNothingAndCreatesNothing` PASS |
| (3) 모든 불확실 상태가 keep인가 | **아니오** | B1 재현 (E4) |
| (3) 행 0개·알 수 없는 상태 | keep | `len==0`·`!= ReachSame` 처리, 해당 테스트 PASS |
| (3) rel 검증 | 충족 | `ValidateRelPath`: 빈 값·절대·`..`·`.` 거부 |
| (3) 8 MiB 상한 | 동작 충족, 테스트 미고정(S1) | E4 탐침 p4 |
| (3) 스킬 루트 자체가 링크 | 동작 Unknown, 테스트 미고정(S1) | E4 탐침 p1, Stat 변이 생존 |
| (3) 열거와 읽기 사이 프로필 교체 | 외부를 읽을 수 있음(M3) | 코드 확인 (재현 안 함) |
| (3) 내용이 다른데 ReachSame | 아니오(해시 충돌 제외). 루트 안 중간 링크를 거친 읽기는 Same이 될 수 있으나 세션도 같은 링크를 따라가므로 실제 도달과 일치 | E4 탐침 |
| (4) F3-iter5 미처리 | 수용 가능(M4) | iter5 E5 M2 사살 + 이번 E5 |
| (4) 기존 쓰기 경로 존치 | merge-ready 불가(B2) | 결정 미기록 (progress-r7 "Conflict") |
| (5) REQ-OWR-005 제안의 코드 관련 주장 | 대부분 정확, 한 곳 부정확(S2) | 아래 E7 |

## 증거 (Claim / Evidence / Baseline-attribution)

측정 기준: 트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702`, `git rev-parse HEAD` → `732fce152ddb9d2a78903c63e17eaf5d3750c46f`, 감사 전후 `git status --short` 빈 출력. 모든 테스트 바이너리는 이 트리에서 `go test`로 빌드했다(설치된 moai 바이너리 미사용, §2.2 해당 없음). 탐침·변이는 `-overlay`로 주입했고 원본은 세션 스크래치패드에만 있다. 모든 테스트는 `t.TempDir()` HOME을 쓴다.

**E1 — 델타 테스트 (실제 코드)**
```
go test ./internal/userassets -run "TestCheckProfileReach_|TestSafeToRemoveProjectCopy_" -count=1 -v
--- PASS ×12 (SameCopyInEveryProfileIsSafeToRemove ... UnreadableStoreIsAnErrorNotAnEmptyAnswer)
ok  	github.com/modu-ai/moai-adk/internal/userassets	0.585s

go test ./internal/config -run "Profile|UserRoot" -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	0.384s
```

**E2 — cli 집중 선택**
```
go test ./internal/cli -run "ZZIter6|RemoveConfigRootMoaiHooks|TestUpdateSettingsEnvHooksSymlinkRefused" -count=1 -v -overlay .../overlay_cli.json
--- PASS: TestUpdateSettingsEnvHooksSymlinkRefused (0.04s)
--- PASS: TestRemoveConfigRootMoaiHooks_ASwapToALinkAfterTheCheckNeverReachesTheExternalTree (0.00s)
--- PASS: TestRemoveConfigRootMoaiHooks_StillRemovesARealDirectoryAndRefusesALink (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.937s
```

**E3 — hook 집중 선택 + Write/Edit 탐침 (배포 YAML, 실제 코드)**
```
go test ./internal/hook -run "T1702|ProfileRoot|UserRoot|ZZIter6" -count=1 -v -overlay .../overlay_hook.json
--- PASS: TestZoneUserRootCovered_ProfileRootTrackedFilesAreRefusedLikeFixedRoots (0.02s)
--- PASS: TestZoneUserRootCovered_AJunkProfileNameIsNotARoot (0.00s)
PROBE Write tracked profile skill      denied=true want=true
PROBE Edit tracked profile skill       denied=true want=true
PROBE Write tracked profile agent      denied=true want=true
PROBE Write untracked profile skill    denied=false want=false
PROBE Write profile settings.json      denied=false want=false
PROBE Write case-variant spelling      denied=true want=true
PROBE Write dotdot spelling            denied=true want=true
--- PASS: TestZZIter6ProbeWriteEditProfileRoot (0.02s)
PROBE UserProfileDirNames=[real] err=<nil>      (store: real/ + linked -> 외부 디렉터리)
ok  	github.com/modu-ai/moai-adk/internal/hook	6.989s
```
탐침은 `checkProtectedZone(harnessLearnerIdentity, tool, path)`를 `zoneLoader`로 배포 `protected-zone.yaml` 파싱 결과에 묶어 호출했다(iter5가 Bash 함수만 직접 보았던 빈틈을 메움).

**E4 — B1 재현 및 판정 형태 탐침 (실제 코드)**
```
go test ./internal/userassets -run "ZZIter6" -count=1 -v -overlay .../overlay_ua.json
PROBE row profile=a state=same reason="content hash equals the project copy"
PROBE rows=1 err=<nil> SafeToRemoveProjectCopy=true          ← 프로필 b(심링크, 사본 없음)는 행조차 없음
PROBE intermediate-link row profile=a state=same reason="content hash equals the project copy"
PROBE escaping-link row profile=a state=unknown reason="path not readable inside the root: statat moai-y/SKILL.md: path escapes from parent"

go test ./internal/userassets -run "ZZIter6ProbeUndiscriminated" -count=1 -v -timeout 60s -overlay .../overlay_ua2.json
PROBE row profile=p1 state=unknown reason="profile root is not a plain directory (a link or a file)"
PROBE row profile=p2 state=unknown reason="path is not a regular file (L---------)"
PROBE row profile=p3 state=unknown reason="path is not a regular file (p---------)"
PROBE row profile=p4 state=unknown reason="file larger than the read bound"
PROBE err=<nil> safe=false
```
심링크 프로필이 실행 가능하다는 전제는 코드로 확인했다: `internal/profile/profile.go:439` `launchCandidateIsUsable`이 `os.Stat`(링크 추적) + `IsDir`로 판단하고, `profile.go:207-213` `isValidProfileName`은 공백·예약 이름을 거르지 않는다. 반면 `config/userroots.go:107`은 `DirEntry.IsDir()`(링크는 false)와 공백·`__no_such_profile__`를 거르는 `ValidProfileDirName`을 쓴다.

**E5 — 변이 검사 (`-overlay`, 트리 미수정)**
```
M-foldall  foldProfileSlug가 모든 base를 접음 (!isUserRootProfileBase(base) → base == "")
  go test ./internal/config -run "Profile|UserRoot" -count=1 -overlay .../ov_m_foldall.json
  protected_zone_profile_test.go:76: user-root:agents-skills@work/x.md   covered = true, want false
  protected_zone_profile_test.go:76: user-root:codex-agents@work/x.md    covered = true, want false
  FAIL                                                                    → 사살
M-oldrm    root.RemoveAll("hooks/moai") → os.RemoveAll(filepath.Join(configRoot,"hooks","moai"))
  go test ./internal/cli -run "RemoveConfigRootMoaiHooks" -count=1 -overlay .../ov_m_oldrm.json
  update_hooks_toctou_test.go:50: the removal followed the swapped link and deleted a file outside the config root: ... no such file or directory
  FAIL                                                                    → 사살
M-stat     readProfileCopy의 os.Lstat(dir) → os.Stat(dir)
  go test ./internal/userassets -run "TestCheckProfileReach_|TestSafeToRemoveProjectCopy_" -count=1 -timeout 120s -overlay .../ov_m_stat.json
  ok  	github.com/modu-ai/moai-adk/internal/userassets	0.483s          → 생존 (S1)
M-noreg    IsRegular·크기 상한 검사 제거(디렉터리만 Unknown)
  go test ... -overlay .../ov_m_noreg.json
  ok  	github.com/modu-ai/moai-adk/internal/userassets	0.401s          → 생존 (S1)
```
M-noreg가 사는 이유: 기존 "filelink" 사례가 루트 **밖**을 가리켜 `os.Root`의 탈출 거부가 먼저 Unknown을 만들기 때문에 IsRegular 검사 자체는 시험되지 않는다.

**E6 — M1 재현 (실제 코드, 루트 안 링크로 교체)**
```
go test ./internal/cli -run "ZZIter6|..." -count=1 -v -overlay .../overlay_cli.json
PROBE in-root swap: stat(skills/moai/SKILL.md) err=stat .../profile/skills/moai/SKILL.md: no such file or directory
```
seam에서 `hooks`를 `hooks.real`로 옮기고 상대 링크 `hooks -> skills`를 만들자 `root.RemoveAll("hooks/moai")`가 `skills/moai`를 지웠다.

**E7 — REQ-OWR-005 제안의 코드 주장 대조**
- 충돌 1 "이 브랜치의 `profile_reach.go`가 그 읽기를 구현" — 정확(`profile_reach.go:76-153`).
- 충돌 2 "`UserProfileDirNames` 열거, 루트 Lstat, `os.OpenRoot`, 상대 파일 Lstat, 제한 읽기" — 정확(`:84, :112, :121, :126, :142`). `root.Open`(`:137`)도 있으나 "제한 읽기"에 포함된다고 볼 수 있다.
- 충돌 3 "`protected_zone.go`에 `foldProfileSlug`, `userroots.go`를 단일 출처로" — 정확(`protected_zone.go:181-191`, `userroots.go` 전체).
- "채택하지 않은 안: 설치 단계에서 모든 프로필 루트에 사본을 넣는 안" — **이 브랜치 코드와 불일치**. `paths.go:124-143` `ResolveRoots`가 프로필별 `claude-skills@<p>`/`claude-agents@<p>` 루트를 내고 `install.go:119`가 그 루트 목록으로 설치한다(S2).

**E8 — 정적 검사**
```
gofmt -l <델타 7개 파일>                 → 출력 없음
go vet ./internal/config ./internal/userassets → 출력 없음
go vet ./internal/hook                   → 출력 없음
go vet ./internal/cli                    → 출력 없음
```

## Gaps (미관측)

- `./internal/cli`, `./internal/hook`, `./internal/userassets`, `./internal/config` 전체 스위트는 실행하지 않았다(지시). 회귀 여부는 집중 선택·vet로만 보았다.
- F2의 "수정 전 RED"는 이번에 다시 재현하지 않았다. 근거는 iter5 E4 탐침(0887e69728에서 `covered=false`)과 진행 기록의 변이 주장이다. 접기를 빼는 변이(no-fold)는 이번에 직접 돌리지 않았고, 모두 접는 변이만 돌렸다.
- M2(FIFO 교체 경합)와 M3(프로필 디렉터리 교체 경합)는 코드 읽기로만 판단했고 재현하지 않았다.
- 대소문자 무시 파일시스템 동작은 macOS 기본 볼륨에서 E3 한 사례로만 보았다. Windows 실행은 관측하지 않았다.
- golangci-lint와 `/verify` 게이트는 돌리지 않았다.
- B2의 운영자 판정 원문은 보지 못했고, 진행 기록에 옮겨진 문구(progress-r6.md 3행)에 의존했다.
- 독립 백엔드(codex)는 지시에 따라 실행하지 않았다. 저장소 설정 `audit.gates.codex: required`는 이 감사로 충족되지 않는다.
- 카드 계약 기반 second-review 기록은 남기지 않았다(지시 범위 밖).

## Residual-risk

- B1을 고쳐도 `CheckProfileReach`는 프로필 루트만 본다. 프로필 없이 실행되는 세션이 읽는 `~/.claude/skills`·`~/.claude/agents`는 판정 대상이 아니므로, 배선하는 쪽(t1642 / 3단계)이 고정 루트를 따로 확인하지 않으면 그 세션에서 자산이 사라질 수 있다.
- 잘못된 호출자가 할 수 있는 일: 오류를 무시하면 행이 nil이라 keep으로 안전하게 떨어진다. 그러나 skills 판정 결과만 보고 같은 자산의 agents 사본까지 지우거나, 다른 rel의 행을 재사용하거나, 프로젝트 사본의 해시가 아닌 값(예: 매니페스트 기록 해시)을 `want`로 넘기면 판정의 의미가 바뀐다. 시그니처가 이를 막지 않는다.
- M1의 루트 안 링크 경합은 사용자 설정 디렉터리 쓰기 권한이 있는 행위자가 필요하다. iter5 P3와 같은 위협 모델이라 비차단으로 두었다.
- 탐침·변이 원본은 세션 스크래치패드에만 있고 핵심 출력만 이 파일로 옮겼다 — 스크래치는 보존 대상이 아니다.

## 반복 이력

- iter3: FAIL
- iter4: FAIL — F1 심링크 RemoveAll, F2 훅 표 4루트, F3 doctor 집계
- iter5: FAIL @0887e6972 — F2 항목 매칭 결함(P1), P3 TOCTOU, F3 Shape 2
- iter6 (본 회차): FAIL @732fce152d — iter5 F2·P3 해소 확인. 새 판정 함수에서 B1(실행 가능한 프로필 누락 → 제거 허가), 쓰기 경로 충돌 미결정(B2). 이번 BLOCKING은 이전 수리에서 나온 것이 아니라(repair-born 아님) 새 기능과 미결 설계 문제다.
