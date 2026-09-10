# t586 1차 판정 — init TUX 렌더 결함 재현과 i18n 대상 목록

- 카드: t586 (Tier M · Class C, 재현 먼저)
- 워크트리: `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`
- 기준 트리: `110320ed274d3a18319f9fbd16d8001afcc214b8` (origin/develop `d060e0d13` 위에 로컬 develop `647ad0157` 을 `--no-ff` 병합, HEAD^2 = `647ad0157`)
- 이번 단계 범위: 재현, i18n 대상 목록, huh 쪽 수리 가능 범위 판단까지. 제품 코드 변경 없음. 실제 홈에 쓰는 `moai init` 은 실행하지 않았다.

## Claim

1. F11 렌더 결함 네 가지가 뷰 문자열과 실제 pty 캡처 양쪽에서 재현된다. 확인 버튼 중앙정렬, 필드 사이와 확인창 안쪽의 빈 줄, 선택 필드 아래 빈 카드 줄, 옵션 설명 열 불일치다.
2. F10 영어 고정 표면은 카드에 적힌 것보다 넓다. 카드에 없던 대상 두 가지를 찾았다. 버전 다운그레이드 확인창과 huh 도움말 줄(v1·v2 공통)이다.
3. 확인창 안쪽 빈 줄 하나를 뺀 나머지 레이아웃 조정은 huh 를 고치지 않고 MoAI 코드에서 공개 API 로 할 수 있다.
4. 위저드(v2) 쪽 수리 지점은 모두 `internal/cli/wizard/wizard.go` 에 있어 t583 병합 뒤로 미룬다.
5. 카드 본문의 줄 번호 일부는 현재 트리와 맞지 않고, `huh_theme.go:95-100` 은 버튼 문구가 아니라 버튼 색 설정이다.

## Evidence

### 기준 트리와 t583 겹침

```
git rev-parse HEAD HEAD^1 HEAD^2   → 110320ed2 / d060e0d13 / 647ad0157
git branch -m WT-init-tux-i18n     → RENAME_EXIT=0
git rev-list --count 110320ed2..WT-init-quiet-wizard   → 1
git log --format='%h %p %s' 110320ed2..WT-init-quiet-wizard
  → 120436f58 d060e0d13 93182d137 Merge local develop tip 93182d137 into t583 worktree (t583)
git diff --stat 110320ed2...WT-init-quiet-wizard       → 출력 없음
```

t583 브랜치의 커밋은 develop 흡수 병합 하나뿐이고, merge-base 기준 변경이 없다. lane-1 의 위저드 수정은 아직 커밋되지 않았으므로 이 트리에서는 보이지 않는다. 겹침 범위는 커밋 기준으로 판단할 수 없다.

### 카드 좌표와 현재 트리 대조

| 카드 인용 | 현재 트리 | 비고 |
|---|---|---|
| `wizard.go:288-292` 옵션 폭 | `wizard.go:290` `key = opt.Label + " - " + opt.Desc` | 줄 이동 |
| `wizard.go:549` 제목 MarginBottom | `wizard.go:557` `NoteTitle ... MarginBottom(1)` | 줄 이동 |
| `init.go:601-602` 프로필 확인 | `init.go:650-651` | 줄 이동, 같은 문자열이 `update.go:178-179` 에도 있음 |
| `huh_theme.go:95-100` Yes/No 고정 | 버튼 **색** 설정 | 문구 출처는 huh v1 기본값 `field_confirm.go:51-52` (`affirmative: "Yes"`, `negative: "No"`) |
| huh `field_confirm.go:64,267-277` | v2.0.3 `field_confirm.go:53` 기본 `buttonAlignment: lipgloss.Center`, `:261-264` 헤더 뒤 `"\n"` 두 번, `:284-291` 정렬 적용 | 버전별 줄 차이 |

사용 중인 모듈: `go.mod` 의 `charm.land/huh/v2 v2.0.3`(위저드), `github.com/charmbracelet/huh v1.0.0`(위저드 밖 표면).

### 재현 1 — 뷰 문자열 (ANSI 제거, 공백을 `·` 로 표시)

프로브는 `buildConfirmField`·`buildSelectField`·`newMoAIWizardTheme`·`buildUnifiedForm` 을 그대로 호출하고, 기존 `formDriver`(80×40)로 구동했다. 소스는 `probe-src-wizard-view.go.txt`, `probe-src-cli-v1.go.txt` 로 보존했다.

```
go test ./internal/cli/wizard/ -run 'TestProbeT586' -v -count=1   → PROBE_V2_EXIT=0 (probe-v2-wizard.txt)
go test ./internal/cli/ -run 'TestProbeT586HuhV1Surfaces' -v -count=1 → PROBE_V1_EXIT=0 (probe-v1-surfaces.txt)
```

v2 확인 그룹(ko):

```
00|┃·Enable·A?
01|┃·First·confirm·description.
02|┃
03|┃·······예·····아니오
04|
05|··Enable·B?
...
12|··▸·Short·-·x
13|····Medium·(Recommended)·-·a·much·longer·description·string
21|←/→·toggle·•·enter·next·•·y·예·•·n·아니오
```

v1 프로필 확인창(init·update 공통 구성):

```
00|┃·No·profile·found.·Set·up·profile·preferences·now?
01|┃·Configure·your·name,·language,·and·model·preferences.
02|┃
03|┃······················Yes·····No
05|←/→·toggle·•·enter·submit·•·y·Yes·•·n·No
```

### 재현 2 — 실제 pty 캡처

테스트 바이너리를 `go test -c` 로 만들어 tmux 3.6a(`TERM=xterm-256color`, 80×30) pty 에서 `form.Run()` 을 돌리고 `tmux capture-pane -p` 로 떴다. 테스트는 환경 변수 `MOAI_T586_TTY=1` 이 없으면 건너뛰고, 바이너리에 `-test.timeout 120s` 를 걸었다. 캡처 뒤 세션 셋을 모두 종료했다(`KILL1..3=0`, 잔존 세션 0). 홈 디렉터리에는 쓰지 않는다.

`tty-wizard-confirm-group-ko.txt`:

```
1 ┃ Enable A?
2 ┃ First confirm description.
3 ┃
4 ┃       예     아니오
5
6   Enable B?
...
13  ▸ Short - x
14    Medium (Recommended) - a much longer description string
15~21 (빈 줄)
22 ←/→ toggle • enter next • y 예 • n 아니오
```

`tty-v1-profile-confirm.txt`:

```
1 ┃ No profile found. Set up profile preferences now?
2 ┃ Configure your name, language, and model preferences.
3 ┃
4 ┃                      Yes     No
6 ←/→ toggle • enter submit • y Yes • n No
```

`tty-wizard-firstpage-ko.txt`: 1~2행이 비어 있고 3행부터 `┃ 대화 언어 선택` 이 나온다. 선택 필드 옵션 넷 아래로 빈 `┃` 줄이 4개(9~12행) 이어진다.

### 결함별 재현 결과

| 결함 | 뷰 문자열 | pty 캡처 |
|---|---|---|
| 확인 버튼 중앙정렬 | v2 들여쓰기 7칸, v1 22칸 | 동일 |
| 확인창 안쪽 빈 줄 | 설명과 버튼 사이 1줄 | 동일 |
| 필드 사이 빈 줄 | 1줄 (`FieldSeparator` 기본값 `"\n\n"`, v2 `theme.go:104`) | 동일 |
| 선택 필드 아래 빈 카드 줄 | 첫 페이지 4줄 | 동일 |
| 옵션 설명 열 불일치 | `Label - Desc` 단순 연결, 설명 시작 열이 줄마다 다름 | 동일 |
| 단계 표시 줄 | 80×40·80×30 뷰 모두 0행에 `● ○ … 1 / 18` | **80×30 pty 에서 보이지 않음** (원인 미확립) |

### i18n 대상 목록 (huh 필드 생성자 스윕)

범위: `internal`·`cmd`·`pkg` 의 비테스트 `.go` 중 `internal/cli/wizard/` 밖에서 `huh.NewConfirm|NewSelect|NewInput|NewNote|NewMultiSelect` 를 쓰는 파일. 결과는 `init.go` 1, `profile_setup.go` 10, `update.go` 1, `update_version.go` 1 이다.

| # | 위치 | 영어 고정 내용 | 번역 자원 |
|---|---|---|---|
| I1 | `init.go:650-653` | 프로필 확인 제목·설명 | 없음 |
| I2 | `update.go:178-181` | I1 과 같은 문자열 | 없음 |
| I3 | `init.go`·`update.go` v1 확인창 버튼 | huh v1 기본 `Yes`/`No` | 위저드 `ConfirmYes/No` 4개 로케일 있음(v2 전용) |
| I4 | `profile_setup.go:330-334` | 1단계 `Select your language`·설명·그룹 제목 `Language` | `profileSetupText.LangSelectTitle` 등 **있으나 쓰이지 않음** |
| I5 | `profile_setup.go:339` | `Setup cancelled.` | 없음 |
| I6 | `update_version.go:327-330` | 다운그레이드 확인 제목·설명·버튼 (**카드에 없음**) | 없음 |
| I7 | v1·v2 전 표면 도움말 줄 | `toggle`·`next`·`submit`·`up`·`down`·`filter`·`select` | 위저드 `HelpSelect`/`HelpInput` 4개 로케일 있음, `internal` 의 비테스트 코드 중 `translations.go` 밖 참조 0건 (같은 패턴이 `translations.go` 에서 5건 잡혀 검색식은 작동함) |

I1·I2 는 프로필이 없을 때 뜨는 창이라 이 시점에는 대화 언어를 알 수 없다. `internal`·`pkg` 에서 `LANG`·`LC_ALL`·`LANGUAGE` 를 읽거나 이름에 `SystemLocale`·`detect…Locale` 이 들어간 함수를 찾았으나 0건이다.

### huh 수리 가능 범위 (공개 API 로 되는가)

프로브: `probe-src-wizard-feasibility.go.txt`, 결과: `probe-v2-feasibility.txt` (`FEAS_EXIT=0`, 6개 PASS). 조정 하나씩만 바꿔 뷰를 비교했다.

| 조정 | API | 결과 | v1 에도 있는가 |
|---|---|---|---|
| 버튼 좌측 정렬 | `(*Confirm).WithButtonAlignment(lipgloss.Left)` | 들여쓰기 7칸 → 3칸(카드 테두리 뒤 바로) | 있음 (`field_confirm.go:372`) |
| 필드 사이 빈 줄 제거 | 테마 `Styles.FieldSeparator = "\n"` | 필드 사이 빈 줄 0 | 있음 (`theme.go:14`) |
| 선택 필드 빈 카드 줄 제거 | `(*Select).Height(n)` | 옵션 아래 빈 `┃` 줄 0 | 있음 |
| 도움말 줄 현지화 | `NewDefaultKeyMap()` + `SetHelp` + `Form.WithKeyMap` | `←/→ 전환` 으로 바뀜 | 있음 (`form.go:286`, `ConfirmKeyMap.Toggle`) |
| 확인창 안쪽 빈 줄 | 없음. `field_confirm.go:261-264` 에 `"\n"` 두 번이 박혀 있음 | `Inline(true)` 는 빈 줄을 없애지만 제목과 설명이 붙어 `Enable A?First confirm…` 로 깨짐 | 같은 구조 |
| 버튼 문구 | `Affirmative`/`Negative` | 위저드는 이미 적용 중 | 있음 |

huh 를 포크하거나 `replace` 로 교체할 필요는 없다(`go.mod` 에 `replace` 없음, `vendor/` 없음).

## 설계 판단이 필요한 항목 (리드·운영자)

- **D1 — 프로필 없음 확인창(I1·I2)의 언어 출처.** 대화 언어를 아직 모르는 시점이고 시스템 로케일 감지 코드가 없다. 후보는 셋이다. 환경 변수로 로케일을 감지하는 방식, 영어를 유지하되 버튼·도움말만 맞추는 방식, 이 확인창을 없애고 init 위저드 첫 질문(대화 언어)으로 대신하는 방식이다. 마지막 안은 t583(질문 16→4)의 설계와 맞물린다.
- **D2 — 구형 v1 profile 위저드.** 흡수할지 최소화할지를 착수 시 정하기로 돼 있다. 판단 근거: v1 필드 생성자 13개가 파일 4개에 흩어져 있고, `profile_setup.go` 는 1단계를 빼면 이미 `profileSetupText` 로 번역돼 있다. 흡수하면 v1 테마·키맵을 따로 맞출 일이 사라지지만 공사가 크다.
- **D3 — 확인창 안쪽 빈 줄.** huh 에 박혀 있어 공개 API 로는 못 뺀다. 카드의 「빈 공간 3겹 축소」 가운데 필드 구분자와 선택 필드 높이만 줄이고 이 한 줄은 남기는 쪽을 권한다. 없애려면 포크나 확인창 자체 구현이 필요하다.
- **D4 — 단계 표시 줄 소실.** 80×30 뷰 문자열에는 보이는데 같은 크기의 pty 캡처에서는 사라진다. 원인은 확인하지 못했다. 운영자가 본 「빈 공간」과 관련 있을 수 있고, t583 뒤에는 질문 수가 바뀌어 증상도 달라질 수 있어 병합 뒤 다시 잰다.
- **D5 — 「선택 항목 폭 정규화」의 뜻.** 설명 열을 맞추는 것인지 줄 길이를 맞추는 것인지 카드만으로는 정할 수 없다. 운영자 스크린샷 두 장은 저장소에서 찾지 못했다(감사 보고서 HTML 에 이미지 참조 없음).
- **D6 — 판정 방법.** 카드는 실제 TTY 캡처로 판정하라고 한다. 이번에 쓴 tmux pty 하네스(테스트 바이너리 + `capture-pane`, 환경 변수로 게이트, 홈 쓰기 없음)를 수리 판정 방법으로 제안한다. 회귀 가드는 뷰 문자열 골든 테스트로 두고, 수리 전후 판정만 pty 캡처로 한다.

## t583 병합 뒤로 미루는 일

위저드 쪽 조정은 모두 `wizard.go` 안에서 일어난다. `buildConfirmField` 의 버튼 정렬, `moaiWizardStyles` 의 `FieldSeparator`, `buildSelectField` 의 높이와 옵션 폭, `buildUnifiedForm` 의 키맵이 여기에 해당한다. 병합 전에 손댈 수 있는 곳은 `huh_theme.go`(v1 테마)와 v1 표면 네 파일이다. 다만 `init.go` 는 t583 이 F1·F4 수리로 건드릴 가능성이 있어, 병합 전 수정은 `huh_theme.go`·`update_version.go`·`profile_setup.go` 로 한정하는 편이 충돌이 적다.

## Baseline-attribution

모든 측정은 이번 실행에서 워크트리 `.claude/worktrees/t586`, HEAD `110320ed274d3a18319f9fbd16d8001afcc214b8`, 브랜치 `WT-init-tux-i18n` 위에서 했다. Go 는 `go1.26.8 darwin/arm64`, huh 는 모듈 캐시의 `charm.land/huh/v2@v2.0.3` 과 `github.com/charmbracelet/huh@v1.0.0` 소스를 읽었다. pty 캡처는 tmux 3.6a, 80×30 에서 떴다. 프로브 테스트 파일은 소스를 증거 폴더로 복사한 뒤 트리에서 지웠고, 커밋 대상은 `.moai/reports/t586/` 뿐이다.

## Gaps

- 운영자가 실제로 쓰는 터미널 앱에서는 캡처하지 않았다. tmux 3.6a 한 가지, 80×30 한 가지 크기만 쟀다.
- pty 캡처는 원래 모습에만 했다. 수리 가능성 조정은 뷰 문자열로만 확인했다.
- 로케일은 `ko`·`en` 만 렌더했다. `ja`·`zh` 는 번역 테이블 존재만 확인했다.
- `update_wizard.go` 의 `moai update -c` 위저드 흐름은 렌더하지 않았다.
- 위저드 질문 번역에 영어가 남았는지는 스윕하지 않았다. 질문 집합을 t583 이 바꾸고 있어 병합 뒤에 잰다.
- lane-1 의 미커밋 `wizard.go` 변경은 보지 못했다.
- D4 원인, D5 의미는 확립하지 못했다.

## Residual-risk

- 프로브의 확인·선택 질문은 합성 질문이다. 실제 init 질문의 제목·설명 길이에서는 버튼 들여쓰기 칸 수가 달라진다. 중앙정렬이라는 성질은 같다.
- v1 확인창 프로브는 `init.go` 의 구성을 복사해 만든 것이라 제품 함수를 직접 부르지 않는다. 제품 쪽 구성이 바뀌면 프로브는 따라가지 않는다. 수리할 때 확인창 구성을 함수로 떼어 내면 이 간극이 사라진다.
- tmux 는 실제 터미널 에뮬레이터와 줄바꿈·너비 처리가 다를 수 있다. 특히 한글·한자 폭이 그렇다.
