# SPEC-INIT-TUX-I18N-001 — 인수 기준

> 카드 t586 · Tier L. 이 문서는 검증 층이다. 요구(GEARS)는 `spec.md` §B 에 있고, 여기 AC 는 모두 Given-When-Then 형식의 이진 판정이다. 각 AC 머리의 `maps REQ-…` 는 요구 매핑 선언이다.

## §A 판정 방식

| 방식 | 무엇으로 판정하나 | 용도 |
|---|---|---|
| **pty** | 환경 변수로 게이트된 테스트 바이너리를 tmux pty(80×30, `TERM=xterm-256color`)에서 실행하고 `tmux capture-pane -p` 로 뜬 화면 | **수리 판정.** 렌더 결함이 실제로 고쳐졌는지. CI 기본값에서는 건너뛴다 |
| **golden** | 폼을 기존 `formDriver` 식 메시지 구동(80×40)으로 그린 `View()` 문자열(ANSI 제거)을 저장된 골든과 비교 | **회귀 가드.** CI 에서 매번 돈다 |
| **mech** | 단위 테스트(이음새 주입), 또는 단일 `git grep`/`go` 명령 | 흐름·저장·import·키 참조 같은 비렌더 성질 |

[HARD] 렌더 성질(버튼 위치, 빈 줄, 빈 카드 줄, 열 정렬, 현지화된 문구가 화면에 나오는지)은 **그려진 문자열이나 캡처 화면**으로만 판정한다. 폼 구성 코드에 특정 호출이 있다는 소스 검사는 렌더 증거가 아니다.

[HARD] 골든·단위 테스트 결과를 읽기 전에 선택된 테스트 수가 0 이 아닌지 먼저 확인한다(`go test … -v` 출력의 `=== RUN` 줄 수). `[no tests to run]` 이 섞인 초록은 판정이 아니다.

[HARD] "0 개", "모두 같다" 같은 성질을 단정하기 전에 **양의 존재 조건**을 먼저 단정한다. 검사 대상 줄(필드 제목, 옵션 줄, 버튼 줄)이 기대한 수만큼 있다는 단정이 실패하면 그 AC 는 FAIL 이다.

[HARD] `git grep` 판정은 출력이 없을 때 종료 코드 1, 있을 때 0 이다. 부재를 단정하는 AC 는 같은 명령 형태의 **양성 대조군**이 출력 1줄 이상을 내는 것을 함께 보인다.

## §B pty 공통 계약 (AC-ITI-003, 015, 016, 017 에 적용, AC-ITI-019·020 이 검증)

| # | 규칙 |
|---|---|
| P1 게이트 | 환경 변수 `MOAI_PTY_CAPTURE=1` 일 때만 실행한다. 없으면 `t.Skip`. SKIP 은 E1 표에서 PASS 로 세지 않고 Gap 으로 적는다 |
| P2 tmux 부재 | `MOAI_PTY_CAPTURE=1` 인데 `tmux` 가 PATH 에 없으면 `t.Fatal`(FAIL)이다. SKIP 이나 PASS 로 떨어지지 않는다 |
| P3 자식 프로그램 | 자식은 이 패키지를 `go test -c` 로 `t.TempDir()` 에 빌드한 테스트 바이너리이며, 두 번째 환경 변수로만 켜지는 도우미 테스트가 실제 제품 함수(명령 경로, 폼 빌더)를 실행한다. 네트워크 이음새는 도우미 안에서 막는다 |
| P4 크기와 환경 | `tmux new-session -d -x 80 -y 30`, 자식 환경 `TERM=xterm-256color`, `HOME=<t.TempDir()>`, 작업 디렉터리 `<t.TempDir()>`(사례가 프로젝트를 요구하면 그 임시 디렉터리 안에 `.moai/config/sections/` 를 시드) |
| P5 캡처 시점 | 사례마다 **기준 문자열**을 정한다. `capture-pane -p` 를 100ms 이하 간격으로 되풀이해 기준 문자열이 나타난 뒤의 화면만 판정한다. 기한(기본 10초) 안에 나타나지 않으면 마지막 캡처를 출력하고 FAIL |
| P6 양의 존재 | 판정 전에 기준 문자열 줄과 사례별 최소 줄 수(필드 제목 수, 옵션 줄 수, 버튼 줄 1)를 단정한다 |
| P7 세션 소유 | 세션 이름은 `moai-ptycap-<테스트명>-<난수>`. `new-session` 성공 직후 `t.Cleanup` 으로 `tmux kill-session -t =<정확한 이름>` 을 등록한다. `kill-server`, 접두 일괄 삭제, 이름 패턴 삭제는 쓰지 않는다 |
| P8 실제 HOME 무기록 | 실행 전후로 실제 HOME 의 감시 경로(`~/.moai/`, `~/.claude/`, `~/.config/moai/`, `~/.zshrc`, `~/.bashrc`, `~/.profile` 중 존재하는 것) 아래 파일의 경로·크기·sha256 매니페스트를 떠 비교한다. 감시 경로 목록과 파일 수를 테스트 출력에 찍는다 |
| P9 반출 | 판정에 쓴 캡처는 텍스트 파일로 `.moai/reports/t586/` 에 반출한다. 수리 전 캡처 커밋이 수리 커밋보다 앞선다 |

## §C 픽스처와 실제 질문 (렌더 AC 의 표면 출처)

t583 뒤 실제 질문 집합에는 확인형 질문이 없다(`spec.md` §A.7). 렌더 AC 는 표면마다 출처를 아래로 고정하고, 픽스처는 **실제 빌더를 거친다는 도달성**을 따로 단정한다.

| AC | 표면 | 출처 | 도달성 단정 |
|---|---|---|---|
| AC-ITI-003 | init 첫 화면 | 실제 명령 경로(t583 흡수 트리의 `InitQuestions`) | 캡처에 init 위저드 `conversation_language` 제목과 스테퍼 `1 / 4` |
| AC-ITI-015 (a) | 다운그레이드 확인창 | 실제 확인창 생성 헬퍼 | 캡처·골든에 해석된 로케일의 제목 문자열 |
| AC-ITI-015 (b) | 위저드 확인형 필드 | 픽스처 `Question{Type: QuestionTypeConfirm}` 을 `buildUnifiedForm` 에 넣음 | `buildField` 가 돌려준 필드가 `*huh.Confirm` 이고 화면에 픽스처 제목·로케일 버튼 라벨이 있음 |
| AC-ITI-016 (a) | init 첫 페이지 | 실제 `InitQuestions` 의 Basic 그룹 | 제목 2개(`conversation_language`, `user_name`) 순서대로 존재 |
| AC-ITI-016 (b) | 프로필 위저드 그룹 전부 | 실제 프로필 질문 세트 | 그룹마다 질문 제목 전부 순서대로 존재 |
| AC-ITI-017 | 설명이 있는 선택 필드 | 실제 `conversation_language` 질문(init·프로필 양쪽) | 옵션 줄 4개(`English`, `Korean (한국어)`, `Japanese (日本語)`, `Chinese (中文)`) 존재 |
| AC-ITI-018 | init 그룹 구성 | 실제 `InitQuestions` | `buildFormGroups` 결과에 `conversation_language`·`agent_wiring` 질문이 각각 존재 |
| AC-ITI-021 | init 그룹 머리 스테퍼 | 실제 `InitQuestions` | 그룹마다 첫 줄에 `●`·`○` 문자가 1개 이상 |
| AC-ITI-022 | 그룹 라벨 비렌더 | 픽스처 질문 2개(`Group` 에 센티널 라벨)를 `buildUnifiedForm` 에 넣음, 그리고 실제 `InitQuestions` | 화면에 픽스처 질문 제목 2개와 `agent_wiring` 질문 제목이 존재 |

## §D AC 매트릭스

### 프로필 없음 진입 (REQ-ITI-001~002)

- **AC-ITI-001** (maps REQ-ITI-001; mech) **Given** 임시 HOME 에 프로필이 없고, 대화형 stdin 이음새를 (a) 참과 (b) 거짓으로, 플래그를 (i) 없음 (ii) `moai init --non-interactive` (iii) `moai update --yes` 로 조합한 경우, **When** 위저드 이음새(`runWizardFn`)와 프로필 위저드 실행 이음새를 모두 주입해 `moai init` 과 `moai update` 의 진입부를 실행하면, **Then** 모든 조합에서 프로필 위저드 이음새 호출 횟수가 0 이고, 확인창 생성 경로가 없다. **And** `git grep -n 'No profile found' -- '*.go' ':!*_test.go'` 는 출력 0줄(종료 1)이고, 같은 형태의 대조군 `git grep -n 'Initialization cancelled' -- '*.go' ':!*_test.go'` 는 1줄 이상이다. **And** `git grep -n 'runProfileSetup(' -- internal/cli/init.go internal/cli/update.go` 는 출력 0줄이고, 대조군 `git grep -n 'runProfileSetup(' -- internal/cli/profile.go` 는 1줄이다.
- **AC-ITI-002** (maps REQ-ITI-002; mech) **Given** 프로필이 없는 임시 HOME 과 대화형 stdin 이음새 참, **When** 질문 목록을 기록하는 위저드 이음새로 `moai init` 대화형 경로를 실행하면, **Then** 이음새가 받은 질문 목록에서 id `conversation_language` 가 정확히 1번, 인덱스 0 에 있고, 프로필 위저드 이음새 호출은 0 이라 실행 전체에서 대화 언어 질문이 만들어진 횟수는 1 이다. **And** init 진입부 위저드 호출 앞에 `runProfileSetup` 호출을 되살린 뮤턴트에서는 이 테스트가 "대화 언어 질문 2회"로 실패함이 관측돼 있다. **And** `moai profile setup` 과 `moai profile --setup` 은 각각 프로필 위저드 이음새를 정확히 1회 호출한다.
- **AC-ITI-003** (maps REQ-ITI-001, REQ-ITI-002; pty) **Given** §B 계약, 프로필 없는 임시 HOME, 빈 임시 작업 디렉터리, **When** 자식 도우미가 `moai init` 대화형 경로를 실행하고 기준 문자열 `Select conversation language` 가 나타난 화면을 캡처한 뒤 `Ctrl+C` 를 보내고 기준 문자열 `Initialization cancelled.` 를 기다리면, **Then** 첫 캡처에 스테퍼 줄이 끝이 `1 / 4` 로 존재하고, `No profile found` 와 `Yes`·`No`·`예`·`아니오` 만으로 된 버튼 줄이 없으며, 두 번째 기준 문자열이 기한 안에 나타나고, 실제 HOME 매니페스트가 실행 전후 같다.

### v1 흡수 (REQ-ITI-003~010)

- **AC-ITI-004** (maps REQ-ITI-003; mech) **Given** 수리된 트리, **When** 다음을 실행하면, **Then** 모두 기대대로다.
  1. `git grep -l '"github.com/charmbracelet/huh"' -- '*.go' ':!*_test.go'` → 출력 0줄(종료 1). 대조군 `git grep -l '"charm.land/huh/v2"' -- '*.go' ':!*_test.go'` → 1줄 이상. (추적 비테스트 Go 파일 전체가 범위다. `internal`·`cmd`·`pkg` 밖 12개 파일 포함)
  2. `git grep -l '"github.com/charmbracelet/huh"' -- '*_test.go'` → 출력 0줄.
  3. `grep -c 'github.com/charmbracelet/huh ' go.mod` → `0`.
  4. `go mod tidy && git diff --exit-code go.mod go.sum` → 종료 0.
  5. `GOOS=windows GOARCH=amd64 go build ./...` → 종료 0.
  6. 단위 테스트가 `moai profile setup` 과 `moai profile --setup` 두 경로 모두 같은 v2 위저드 실행 이음새에 도달함을 보인다.
- **AC-ITI-005** (maps REQ-ITI-004; mech) **Given** `t.TempDir()` 프로젝트(`.moai/config/sections/` 에 `user.yaml`·`language.yaml`·`quality.yaml` 시드)와 임시 프로필 저장 경로, 모든 필드에 시드와 다른 답을 주입한 위저드 결과, **When** 저장 단계를 실행하면, **Then** 다음이 모두 참이다.
  1. `preferences.yaml` 의 이름·네 언어·모델·모델 정책·추론 강도·권한 모드가 주입값(권한 모드는 REQ-ITI-005 (2) 정규화 적용)이다.
  2. `user.yaml` 이름, `language.yaml` 의 `conversation_language`·`conversation_language_name`·`git_commit_messages`·`code_comments`·`documentation` 이 주입값이다.
  3. `quality.yaml` 의 `development_mode` 가 주입값이다.
  4. 선택 필드별 옵션 값 집합: `model`·`effort_level`·`development_mode` 는 `settings.FieldOptionDefs` 값 + 빈 값, `permission_mode` 는 `settings.FieldOptionDefs` 값(빈 값 없음), 네 언어 필드는 `settings.FieldOptionDefs("conversation_lang")` 값과 같고, `model_policy` 는 `template.ValidModelPolicies()` + 빈 값이다.
  5. 프로젝트 밖(`.moai` 없음) 작업 디렉터리에서는 `preferences.yaml` 만 쓰이고 임시 디렉터리에 `.moai` 가 생기지 않는다.
- **AC-ITI-006** (maps REQ-ITI-005; mech) **Given** 다음 아홉 사례를 담은 표 테스트(각 사례는 저장 전 파일을 바이트로 보관), **When** 흡수된 위저드의 바인딩과 저장을 실행하면, **Then** 아홉 사례 모두 기대대로다.
  1. 저장값이 퇴역한 모델 id 이면 정규화된 별칭이 사전 선택된다(빈 값이 아니다).
  2. `acceptEdits` 를 고르면 저장되는 권한 모드가 빈 문자열이고, 출력에 `acceptEditsConfirmationLine` 이 정확히 1번 나온다.
  3. 개발 방식 선택 초기값이 프로젝트 `quality.yaml` 의 값이다.
  4. 저장 전 `statusline_segments` 가 저장 뒤 `preferences.yaml` 에 같은 맵으로 남는다.
  5. 나머지 선택 필드(네 언어·모델 정책·추론 강도·권한 모드)는 저장된 값이 사전 선택된다.
  6. `.moai/config/sections/statusline.yaml` 이 저장 전후 바이트 단위로 같다.
  7. `.moai/config/sections/git-convention.yaml` 이 저장 전후 바이트 단위로 같다.
  8. 저장 뒤 `preferences.yaml` 에 `statusline_theme` 키가 없다.
  9. 저장된 권한 모드가 빈 문자열이면 권한 모드 선택의 사전 선택값이 `acceptEdits` 다.
- **AC-ITI-007** (maps REQ-ITI-006; mech) **Given** 위저드 실행 이음새와 세션 worktree 진입·정리 이음새를 주입한 명령 수준 테스트, 임시 HOME, **When** `moai profile setup` 과 `moai profile --setup` 각각을 위저드 결과 (a) 사용자 취소 (b) 오류 (c) 성공 으로 실행하고, (c) 는 이름 인자 없음과 `work` 두 경우로 실행하면, **Then** (a) 는 명령이 nil 을 돌려주고(종료 0), 선택된 로케일의 `SetupCancelled` 문구가 출력되며, 프로필 파일이 생기지 않는다. (b) 는 non-nil 오류를 돌려주고 프로필 파일이 생기지 않는다. (c) 는 이름 없음이면 `default`, `work` 면 `work` 프로필 경로에 `preferences.yaml` 이 생기고 저장 문구와 요약이 출력된다. **And** 모든 실행에서 정리 이음새가 정확히 1번 불리고, 그 clean-exit 인자가 (a)·(c) 에서 참, (b) 에서 거짓이며, 진입 이음새가 프로필 읽기보다 먼저 불린다.
- **AC-ITI-008** (maps REQ-ITI-007; golden) **Given** 흡수된 프로필 위저드 폼, **When** 첫 질문에서 `ko`·`ja`·`zh` 를 각각 고르고 다음 그룹들로 이동하며 `View()` 를 그리면, **Then** 각 로케일 골든과 일치하고, 영어 문자열 집합 E(프로필 위저드가 그리는 질문 제목·설명·현지화되는 옵션 라벨·도움말 동작 라벨의 `en` 값) 가운데 예외 목록 X 에 없는 문자열이 그려진 화면에 하나도 없다. 예외 목록 X 는 다음으로 닫혀 있다(계산으로 넓히지 않는다).
  - 언어 옵션 라벨: `English`, `Korean (한국어)`, `Japanese (日本語)`, `Chinese (中文)`
  - 현지화 라벨이 없는 옵션에서 라벨 자리에 그려지는 값: `template.ModelAliasPickerValues()` 가 돌려주는 문자열 전부, `low`, `medium`, `high`, `xhigh`, `max`, `acceptEdits`, `auto`, `default`, `plan`, `bypassPermissions`, `dontAsk`, `models.ValidDevelopmentModes()` 값 전부
  - 도움말 줄의 키 표기: `enter`, `tab`, `shift+tab`, `esc`, `↑`, `↓`, `←/→`, `/`, `x`, `y`, `n`, `ctrl+e`, `ctrl+u`, `ctrl+d`, `ctrl+a`, `g/home`, `G/end`
  - 고유명사: `MoAI`, `Claude`
- **AC-ITI-009** (maps REQ-ITI-008; golden; 선결 V-c) **Given** 흡수된 프로필 위저드 폼과 init 위저드 폼(대조군), **When** 각 그룹을 차례로 `View()` 로 그리면, **Then** 프로필 위저드 각 그룹의 첫 줄(ANSI 제거)에서 `●`·`○` 문자 수의 합이 N 과 같고, 줄이 `<k> / <N>` 으로 끝나며, N 은 보이는 프로필 질문 수(`design.md` §2 의 10)이고 k 는 그 그룹 첫 질문의 순번이다. **And** 같은 규칙을 init 위저드 각 그룹 첫 줄에 적용해도 참이다(N 은 보이는 init 질문 수).
- **AC-ITI-010** (maps REQ-ITI-010; mech) **Given** 재조준된 가드 9건(`spec.md` §A.5 S1~S9, 대상은 `design.md` §10), **When** 다음을 실행하면, **Then** 모두 기대대로다.
  1. `go test ./internal/cli/ -run '<S1~S9 의 테스트 함수 이름 9개 또는 대체 테스트 이름>' -count=1 -v` 에서 `=== RUN` 최상위 줄이 9개 이상이고 모두 `--- PASS`.
  2. 스캔형으로 남은 가드는 읽는 파일에 기준 문자열(해당 질문 정의의 `ID: "<id>"` 또는 저장 함수 이름)이 있음을 먼저 단정한다. 빈 파일이나 무관한 파일을 가리키게 한 뮤턴트에서 실패함이 관측돼 있다.
  3. 음성 가드 뮤턴트: 프로필 질문 세트에 `ID: "statusline_theme"` 질문을 v2 정의 형태로 추가하면 S7(또는 대체 테스트)이 실패하고, 저장 경로 파일에 `yaml.Marshal` 호출을 추가하면 S2 가 실패하며, 저장 구조체에 `StatuslineTheme:` 대입을 추가하면 S8 이 실패함이 각각 관측돼 있다.
  4. 양성 가드 뮤턴트: 프로필 질문 세트에서 `model_policy` 질문을 지우면 S1 이, `development_mode` 질문을 지우면 S5 가, 프로젝트 동기화 앞의 nil 세그먼트 대입을 지우면 S9 가 실패함이 각각 관측돼 있다.
- **AC-ITI-011** (maps REQ-ITI-009; mech + golden) **Given** 수리된 트리, **When** 다음을 실행하면, **Then** 모두 기대대로다.
  1. `git ls-files internal/cli/huh_theme.go internal/cli/huh_theme_test.go` → 출력 0줄.
  2. `git grep -n -E 'moaiHuhTheme|moaiHuhStyles|huhThemeIsDark' -- '*.go'` → 출력 0줄. 대조군 `git grep -n 'var wizardIsDark' -- internal/cli/wizard/wizard.go` → 1줄.
  3. 다운그레이드 확인창과 프로필 위저드 첫 그룹을 `wizardIsDark` 를 참으로 강제해 그린 골든과 거짓으로 강제해 그린 골든이 ANSI 포함 문자열로 서로 다르고, 각각 저장된 골든과 일치한다. 강제는 패키지 변수로만 하고 환경 변수를 바꾸지 않는다.

### 잔여 영어 표면 (REQ-ITI-011~013)

- **AC-ITI-012** (maps REQ-ITI-011; golden; 선결 V-b) **Given** 다운그레이드 확인창과, 작업 디렉터리·활성 프로필을 조합한 네 사례, **When** 각 사례로 언어를 해석해 `View()` 를 그리면, **Then** 제목·설명·두 버튼·도움말 동작 라벨이 기대 로케일로 그려지고 사례별 골든과 일치한다.
  - (a) 작업 디렉터리에 `.moai` 가 있고 `language.yaml` 이 `ja`, 활성 프로필 `ko` → `ja`
  - (b) 작업 디렉터리에 `.moai` 가 없음(프로젝트 밖), 활성 프로필 `ko` → `ko`
  - (c) 프로젝트 밖, 프로필 없음 → `en`
  - (d) 작업 디렉터리에 `.moai` 가 있으나 `language.yaml` 에 `conversation_language` 가 없음, 활성 프로필 `zh` → `zh`
  **And** 프로필을 프로젝트보다 먼저 보는 순서로 바꾼 뮤턴트에서 (a) 가 실패함이 관측돼 있다.
- **AC-ITI-013** (maps REQ-ITI-012; golden + mech; 선결 V-b) **Given** 네 표면(init 위저드, `moai update -c` 재구성 위저드, 프로필 위저드, 다운그레이드 확인창), **When** `en`·`ko`·`ja`·`zh` 각각으로 `View()` 를 그리면, **Then** 16개 골든과 모두 일치하고, 각 화면의 도움말 줄에 나온 동작 라벨이 모두 `design.md` §7 표의 해당 로케일 값이며, 표에 없는 동작 라벨이 나오면 실패한다. `ko`·`ja`·`zh` 화면의 도움말 줄에는 `next`, `submit`, `back`, `select`, `up`, `down`, `filter`, `toggle` 이 없다. **And** `git grep -n -E 'HelpSelect|HelpInput' -- '*.go'` → 출력 0줄이고, 같은 형태의 대조군 `git grep -n -E 'ConfirmYes|ConfirmNo' -- 'internal/cli/wizard/*.go'` → 1줄 이상이다.
- **AC-ITI-014** (maps REQ-ITI-013; mech) **Given** 수리된 트리, **When** 키 참조 스윕 테스트와 로케일 동등성 테스트를 실행하면, **Then** 위저드 번역 테이블과 `profileSetupText` 의 모든 키가 비테스트 코드에서 1회 이상 참조되고, `ko`·`ja`·`zh` 의 키 집합이 `en` 과 같다. **And** 참조되지 않는 키를 하나 심은 뮤턴트와 `ja` 에서 키 하나를 뺀 뮤턴트에서 각각 실패함이 관측돼 있다.

### 레이아웃 (REQ-ITI-014~017)

- **AC-ITI-015** (maps REQ-ITI-014; pty + golden) **Given** §C 의 두 표면 (a) 다운그레이드 확인창 (b) 위저드 확인형 픽스처, **When** 각각 pty 캡처와 `View()` 를 뜨면, **Then** 두 결과 모두에서 제목 줄·설명 줄·버튼 줄이 각각 1개 이상 있음을 먼저 단정한 뒤, 버튼 줄의 첫 버튼 라벨 시작 표시 열이 설명 줄 첫 글자의 표시 열과 같다(수리 전 값: v2 7칸, v1 22칸 들여쓰기). 골든이 이를 고정한다.
- **AC-ITI-016** (maps REQ-ITI-015; pty + golden) **Given** §C 의 표면 (a) init 첫 페이지 (b) 프로필 위저드의 모든 그룹, 그리고 골든에 한해 (c) 재구성 위저드 첫 그룹, **When** pty 캡처와 `View()` 를 뜨면, **Then** 그룹마다 질문 제목이 모두 순서대로 있고 선택 필드의 옵션 줄 수가 옵션 수와 같음을 먼저 단정한 뒤, 연속한 두 필드 사이의 빈 줄이 0 이고 선택 필드 마지막 옵션 줄 바로 아래에 내용 없는 카드 줄(`┃` 뒤 공백뿐인 줄)이 0 이다(수리 전 값: 필드 사이 1줄, 첫 페이지 빈 카드 줄 4줄).
- **AC-ITI-017** (maps REQ-ITI-016; pty + golden) **Given** §C 의 실제 `conversation_language` 선택 필드(init 과 프로필 양쪽, 라벨에 한글·한자·가나와 영문이 섞임), **When** pty 캡처와 `View()` 를 뜨고 옵션 줄 4개를 찾은 뒤 각 줄에서 설명 시작 위치를 표시 폭(동아시아 전각 문자 2칸)으로 계산하면, **Then** 옵션 줄 4개가 모두 있고, 네 줄의 설명 시작 표시 열이 같다. **And** 룬 수로 열을 맞춘 뮤턴트와 바이트 수로 열을 맞춘 뮤턴트에서 각각 이 검사가 실패함이 관측돼 있다.
- **AC-ITI-018** (maps REQ-ITI-017; mech — 페이지 수) **Given** t583 을 흡수한 트리의 `InitQuestions`, **When** `buildFormGroups` 로 그룹을 만들면, **Then** 그룹이 정확히 2개이고, `conversation_language`·`user_name` 은 첫 그룹에, `agent_wiring`·`autonomy_tier` 는 마지막 그룹에 이 순서로 함께 있으며, 두 질문의 `Group` 값이 모두 `Agents & Autonomy` 이고, 어떤 init 질문의 `Group` 도 `Quality & Workflow` 가 아니다. 이 AC 는 스테퍼 분모를 단정하지 않는다. **And** 페이지 수 뮤턴트 — `autonomy_tier` 의 `Group` 을 `Autonomy` 로 되돌려 그룹을 다시 나눈 트리(질문 수 4 는 그대로) — 에서 이 AC 는 "그룹 3개"로 실패하고, 같은 뮤턴트에서 AC-ITI-021 은 PASS 임이 함께 관측돼 있다.
- **AC-ITI-021** (maps REQ-ITI-017; mech — 스테퍼 분모) **Given** t583 을 흡수한 트리의 `InitQuestions`, **When** `buildFormGroups` 의 그룹을 차례로 `View()` 로 그리면(ANSI 제거), **Then** 그룹이 1개 이상 그려졌음을 먼저 단정한 뒤, 각 그룹 첫 줄에서 `●`·`○` 문자 수의 합이 4 이고, 줄이 `<k> / 4` 로 끝나며, k 는 그 그룹 첫 질문의 보이는 순번이다. 이 AC 는 그룹 수를 단정하지 않는다. **And** 분모 뮤턴트 — init 질문 세트의 `user_name` 바로 뒤에 조건 없는 입력형 질문 하나를 `Group: "Basic"` 으로 더한 트리(그룹 수 2 는 그대로) — 에서 이 AC 는 "분모 5"로 실패하고, 같은 뮤턴트에서 AC-ITI-018 은 PASS 임이 함께 관측돼 있다. 두 AC 와 별도로 수리 트리의 두 번째 그룹 골든(두 질문 제목과 끝이 `3 / 4` 인 스테퍼 줄)을 회귀 가드로 커밋하되, 이 골든은 두 뮤턴트 모두에서 달라지므로 어느 한 성질의 증거로 세지 않는다.
- **AC-ITI-022** (maps REQ-ITI-017; golden + mech — 그룹 라벨 비렌더, 번역 키 없음) **Given** 제목이 서로 다른 조건 없는 픽스처 질문 2개의 `Group` 에 센티널 라벨 `ZZ-GROUP-LABEL-SENTINEL` 을 넣은 질문 목록과, t583 을 흡수한 트리의 실제 `InitQuestions`, **When** 픽스처 목록으로 `buildUnifiedForm` 을 만들어 모든 그룹을 `View()` 로 그리고(ANSI 제거) 실제 init 폼도 모든 그룹을 그리면, **Then** 픽스처 화면에 두 질문 제목이 모두 있음을 먼저 단정한 뒤 센티널 문자열이 0회이고, 실제 init 화면에 `agent_wiring` 질문 제목이 있음을 먼저 단정한 뒤 `Agents & Autonomy` 가 0회다. **And** `git grep -n -F 'Agents & Autonomy' -- internal/cli/wizard/translations.go` 는 출력 0줄(종료 1)이고, 같은 형태의 대조군 `git grep -n -F 'ConfirmYes' -- internal/cli/wizard/translations.go` 는 1줄 이상이다. **And** `git grep -n -E '\.Group([^A-Za-z0-9_.]|$)' -- 'internal/cli/wizard/*.go' ':!*_test.go'` 의 출력에 `buildFormGroups` 안의 `q.Group` 줄이 1줄 이상 있고(대조군 — 이 패턴이 필드 읽기에 닿음), 출력의 모든 줄이 `huh.Group` 타입 이름이거나 `buildFormGroups` 안의 `q.Group` 묶기 비교·대입이다. 그 밖의 줄이 하나라도 있으면(새 파일의 렌더 경로 포함) FAIL 이다. **And** `buildFormGroups` 의 `huh.NewGroup(fields...)` 에 `.Title(pending[0].Group)` 을 붙인 뮤턴트에서 이 AC 가 센티널 1회 이상과 grep 규칙 위반으로 실패함이 관측돼 있다. 판단 근거(그룹 라벨을 읽는 렌더 경로가 없으므로 `en`·`ko`·`ja`·`zh` 번역 표에 그룹 라벨 키를 두지 않음)는 `research.md` §6.1 의 측정이다.

### 판정 하네스 (REQ-ITI-018)

- **AC-ITI-019** (maps REQ-ITI-018; mech) **Given** pty 캡처 테스트 전체, **When** (a) `MOAI_PTY_CAPTURE` 없이 실행하고 (b) `MOAI_PTY_CAPTURE=1` 에 `tmux` 가 없는 PATH 로 실행하면, **Then** (a) 는 캡처 테스트가 모두 `--- SKIP` 으로 보고되고, 실행 전후 `tmux list-sessions -F '#{session_name}'` 출력(tmux 가 있을 때)이 같다. (b) 는 캡처 테스트가 `--- FAIL` 로 보고되고 `--- PASS`·`--- SKIP` 이 없다.
- **AC-ITI-020** (maps REQ-ITI-018; mech; 선결 V-d — (2)(3) 절) **Given** `MOAI_PTY_CAPTURE=1` 과 tmux, 실행 전에 검증자가 만든 센티널 세션 `moai-ptycap-sentinel-<난수>`, **When** 다음 세 실행을 하면, **Then** 모두 기대대로다.
  1. 정상 실행: 캡처 파일이 생기고 기준 문자열을 담는다. 실행 뒤 `moai-ptycap-` 로 시작하는 세션은 센티널 하나뿐이다.
  2. 강제 실패 실행: 세션을 연 뒤 `t.Fatal` 하는 자기 검증 하위 테스트를 자식 `go test` 로 돌리면, 그 하위 테스트는 FAIL 이고 실행 뒤 그 테스트가 연 세션이 남지 않으며 센티널은 살아 있다.
  3. 강제 기한 초과 실행: 나타나지 않는 기준 문자열을 기다리는 자기 검증 하위 테스트는 기한 초과 메시지와 함께 FAIL 이고, 세션이 남지 않는다.
  **And** 세 실행 모두 테스트 출력에 감시 경로 목록과 파일 수가 찍히고(파일 수 0 이면 이 절은 Gap), 실제 HOME 매니페스트가 실행 전후 같다. **And** 자식 프로세스의 HOME·작업 디렉터리가 실제 HOME·저장소 경로와 다름을 출력으로 보인다.

## §D.1 판정 방식, 중요도, RED 를 뜰 마일스톤

| AC | 방식 | 중요도 | RED 마일스톤 | 선결 |
|---|---|---|---|---|
| AC-ITI-001, 002 | mech | MUST | M4 | — |
| AC-ITI-003 | pty | MUST | 게이트 직후 | — |
| AC-ITI-004 | mech | MUST | plan (§D.3 L1) / M6 | — |
| AC-ITI-005, 006 | mech | MUST | M1 / M5 | — |
| AC-ITI-007 | mech | MUST | M5 | — |
| AC-ITI-008 | golden | MUST | M5 | — |
| AC-ITI-009 | golden | MUST | M5 | V-c |
| AC-ITI-010 | mech | MUST | M5 | — |
| AC-ITI-011 | mech + golden | MUST | M6 | — |
| AC-ITI-012 | golden | MUST | M2 | V-b |
| AC-ITI-013 | golden + mech | MUST | plan (§D.3 L3) / M7 | V-b |
| AC-ITI-014 | mech | SHOULD | M8 | — |
| AC-ITI-015 | pty + golden | MUST | M3 (a) / 게이트 직후 (b) | — |
| AC-ITI-016, 017 | pty + golden | MUST | 게이트 직후 | — |
| AC-ITI-018 | mech | MUST | 게이트 직후 | — |
| AC-ITI-019 | mech | MUST | M3 | — |
| AC-ITI-020 | mech | MUST | M3 | V-d ((2)(3) 절) |
| AC-ITI-021 | mech | MUST | 게이트 직후 | — |
| AC-ITI-022 | golden + mech | MUST | 게이트 직후 | — |

「선결」 열의 V-a~V-d 는 `plan.md` M1 착수 검증 항목이다. 그 항목이 참으로 닫히기 전에는 해당 AC 를 판정하지 않고, 거짓으로 닫히면 판정 대신 리드에게 올린다. V-a 에 기대는 AC 는 없다.

방식별 개수(AC 22개, 각 AC 를 한 번씩 셈): pty 만으로 판정하는 AC 1개(003). 골든을 포함하는 AC 9개(008, 009, 011, 012, 013, 015, 016, 017, 022; 015~017 은 pty, 011·013·022 는 mech 와 겹침). 기계 검사만으로 판정하는 AC 10개(001, 002, 004, 005, 006, 007, 010, 014, 018, 021; 021 은 `View()` 를 그리지만 저장 골든이 아니라 규칙으로 판정). 하네스 AC 2개(019, 020). 수리 판정 pty AC 는 003·015·016·017 의 4개이고, 020 의 1절은 하네스 자체를 pty 로 돈다.

## §D.2 추적성

REQ-ITI-001→AC-ITI-001/003 · REQ-ITI-002→AC-ITI-002/003 · REQ-ITI-003→AC-ITI-004 · REQ-ITI-004→AC-ITI-005 · REQ-ITI-005→AC-ITI-006 · REQ-ITI-006→AC-ITI-007 · REQ-ITI-007→AC-ITI-008 · REQ-ITI-008→AC-ITI-009 · REQ-ITI-009→AC-ITI-011 · REQ-ITI-010→AC-ITI-010 · REQ-ITI-011→AC-ITI-012 · REQ-ITI-012→AC-ITI-013 · REQ-ITI-013→AC-ITI-014 · REQ-ITI-014→AC-ITI-015 · REQ-ITI-015→AC-ITI-016 · REQ-ITI-016→AC-ITI-017 · REQ-ITI-017→AC-ITI-018/021/022 · REQ-ITI-018→AC-ITI-019/020. 모든 REQ 에 AC 가 1개 이상 있고, 매핑 없는 AC 는 없다.

## §D.3 RED 기준선

plan 단계에서 실제로 실행한 명령만 적는다. 트리는 워크트리 `.claude/worktrees/t586`, HEAD `18144b7aca714ea8924363b1eab4640cf101c6d0`. 판정 규칙은 모두 "출력 0줄 = GREEN" 이다. 렌더 AC 의 공식 RED 는 t583 흡수 트리에서 뜬다(§D.1).

| 원장 | 명령 | 종료 코드 | 연결 AC |
|---|---|---|---|
| L1 | `git grep -l '"github.com/charmbracelet/huh"' -- '*.go' ':!*_test.go'` | 0 | AC-ITI-004 |
| L2 | `git grep -n 'No profile found' -- '*.go' ':!*_test.go'` | 0 | AC-ITI-001 |
| L3 | `git grep -n -E 'HelpSelect\|HelpInput' -- '*.go'` (셸에서는 `'HelpSelect|HelpInput'`) | 0 | AC-ITI-013 |
| L4 | `git grep -n 'runProfileSetup(' -- internal/cli/init.go internal/cli/update.go` | 0 | AC-ITI-001 |

L1 출력(원문):

```
internal/cli/huh_theme.go
internal/cli/init.go
internal/cli/profile_setup.go
internal/cli/update.go
internal/cli/update_version.go
```

대조군 `git grep -l '"charm.land/huh/v2"' -- '*.go' ':!*_test.go'` 출력: `internal/cli/wizard/wizard.go` (종료 0).

L2 출력(원문):

```
internal/cli/init.go:651:			Title("No profile found. Set up profile preferences now?").
internal/cli/update.go:179:				Title("No profile found. Set up profile preferences now?").
```

대조군 `git grep -n 'Initialization cancelled' -- '*.go' ':!*_test.go'` 출력: `internal/cli/init.go:707:				_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.")` (종료 0).

L3 출력(원문, 앞 10줄):

```
internal/cli/wizard/translations.go:18:	HelpSelect    string
internal/cli/wizard/translations.go:19:	HelpInput     string
internal/cli/wizard/translations.go:542:		HelpSelect:    "Use arrow keys to navigate, Enter to select, Esc to cancel",
internal/cli/wizard/translations.go:543:		HelpInput:     "Type your answer, Enter to confirm, Esc to cancel",
internal/cli/wizard/translations.go:549:		HelpSelect:    "방향키로 이동, Enter로 선택, Esc로 취소",
internal/cli/wizard/translations.go:550:		HelpInput:     "답변 입력 후 Enter로 확인, Esc로 취소",
internal/cli/wizard/translations.go:556:		HelpSelect:    "矢印キーで移動、Enterで選択、Escでキャンセル",
internal/cli/wizard/translations.go:557:		HelpInput:     "入力してEnterで確定、Escでキャンセル",
internal/cli/wizard/translations.go:563:		HelpSelect:    "使用方向键导航，Enter选择，Esc取消",
internal/cli/wizard/translations.go:564:		HelpInput:     "输入答案，Enter确认，Esc取消",
```

(이어서 `internal/cli/wizard/wizard_test.go:283`, `:284`, `:289`, `:290`, `:298`, `:1256`, `:1257`, `:1259`, `:1260` 9줄.) 대조군 `git grep -n -E 'ConfirmYes|ConfirmNo' -- 'internal/cli/wizard/*.go' ':!*_test.go'` 첫 줄: `internal/cli/wizard/translations.go:21:	// ConfirmYes / ConfirmNo localize the huh Confirm affirmative/negative` (종료 0).

L4 출력(원문):

```
internal/cli/init.go:659:			if err := runProfileSetup(cmd, nil); err != nil {
internal/cli/update.go:187:				if err := runProfileSetup(cmd, nil); err != nil {
```

RED 이유: L1 은 v1 importer 5개, L2 는 확인창 문자열 2곳, L3 은 문장형 도움말 키와 그 테스트, L4 는 init·update 의 프로필 위저드 호출 2곳이 남아 있어 빨갛다.

## §D.4 경계 사례

- 프로필 위저드 도중 대화 언어를 바꾼 뒤 이전 그룹으로 돌아가면 이전 그룹도 새 언어로 다시 그려진다. 확인형 버튼 라벨은 huh v2 에 동적 설정자가 없어 폼 구성 시점 언어로 고정되지만, 프로필 위저드에는 확인형 질문이 없다.
- 모델 저장값이 알 수 없는 문자열이면 정규화가 빈 값으로 떨어뜨리는 현재 동작을 유지한다(`normalizeModelLegacy1M`).
- 프로젝트 밖에서 `moai profile setup` 을 실행하면 프로젝트 동기화와 개발 방식 저장은 건너뛰고 요약에 프로젝트 동기화 없음이 나온다(AC-ITI-005 (5)).
- 옵션 라벨이 터미널 폭보다 길면 정렬 열 계산이 폭을 넘지 않도록 줄인다. 골든에 긴 라벨 사례를 하나 둔다.
- Windows 에서는 pty 캡처 테스트가 SKIP 이며, 그 SKIP 은 PASS 로 세지 않는다.
- 다운그레이드 확인창에서 `language.yaml` 이 읽기·파싱에 실패하면 값이 없는 것과 같게 보고 프로필로 넘어간다(`wizard.ReadLocaleFromProject` 의 빈 문자열 반환).

## §D.5 품질 게이트

- 변경 패키지 테스트: `go test ./internal/cli/wizard/... -count=1`, `go test ./internal/cli/ -run '<변경 영역 선택자>' -count=1 -v` — 선택 수 0 금지.
- 크로스 빌드: `GOOS=windows GOARCH=amd64 go build ./...`.
- 정적 검사: `go vet ./internal/cli/...`, `golangci-lint run` 신규 지적 0.
- 전체 스위트 판정은 develop push 뒤 CI.

## §D.6 완료 정의

- MUST AC 21개가 모두 PASS 이고, 각 판정에 명령과 원문 출력이 붙어 있다. pty SKIP 은 PASS 가 아니다.
- 뮤턴트를 요구한 AC(002, 010, 012, 014, 017, 018, 021, 022)의 뮤턴트 실패 원문이 progress 기록에 있다. 018·021 은 상대 AC 가 같은 뮤턴트에서 PASS 인 원문도 함께 있다.
- `plan.md` M1 착수 검증 V-a~V-d 가 progress 기록에서 참·거짓으로 닫혀 있고, 거짓으로 닫힌 항목에 기대는 AC 의 리드 처분이 기록돼 있다.
- pty 판정 캡처(수리 전·후)가 `.moai/reports/t586/` 에 반출돼 있고, 수리 전 캡처 커밋이 수리 커밋보다 앞선다.
- 골든 파일이 커밋돼 있고 CI 에서 돈다.
- `spec.md` §D 의 D3 한계 목록과 D4 연기 기록이 progress 기록에 남아 있다.
- 템플릿(`internal/template/templates/**`) diff 가 0 이다.
