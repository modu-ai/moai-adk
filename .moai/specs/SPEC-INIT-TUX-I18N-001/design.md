# SPEC-INIT-TUX-I18N-001 — 설계

> 카드 t586 · Tier L. 요구는 `spec.md` §B, 판정은 `acceptance.md`. 이 문서는 run 단계가 따를 구조 결정과 기각한 대안을 적는다. 줄 번호는 트리 `18144b7ac` 기준이고, t583 겹침 파일의 좌표는 흡수 트리에서 다시 잰다.

## §1 목표와 제약

- huh v1 을 `internal/cli` 에서 없애고 모든 대화형 폼을 v2 위저드 엔진 하나로 모은다.
- `internal/cli` → `internal/cli/wizard` 한 방향 import 를 지킨다. 위저드는 `cli`·`settings` 라벨 해석을 모른다.
- 저장 경로와 저장값 보존 동작(REQ-ITI-005 아홉 가지)은 한 줄도 바꾸지 않는다. 흡수는 폼 실행만 바꾼다.
- huh 는 포크하지 않는다. 공개 API 로 되는 조정만 한다.

## §2 프로필 위저드 흡수 구조

### §2.1 현재 구조

`runProfileSetup`(`profile_setup.go:240`)이 세션 worktree 진입 → 기존 값 읽기 → v1 1단계 언어 폼(`:327-335`) → 로케일 텍스트 선택 → v1 본 폼 4그룹(`:354-442`) → 권한 모드 정규화 → `preferences.yaml` 쓰기 → 프로젝트 동기화·개발 방식 저장 → 요약 순으로 한 함수에서 처리한다. 값은 지역 변수에 `Value(&x)` 로 바인딩된다.

### §2.2 변경 뒤 구조

```
runProfileSetup (cli)
 ├─ 세션 worktree 진입 / defer 정리          ← 그대로 (이음새로 뺌)
 ├─ 기존 값 읽기, 정규화, 초기값 계산        ← 그대로
 ├─ 옵션 목록 만들기 (cli, §3)
 ├─ profileWizardRunner(초기값, 옵션, 로케일) ← 이음새. 기본 구현은 wizard 패키지의 프로필 실행 함수
 │    └─ wizard: 프로필 질문 세트 + buildUnifiedForm 계열 + 스테퍼
 ├─ ErrCancelled → SetupCancelled 출력, nil
 ├─ 기타 오류   → 오류 반환
 └─ 저장 이하                                  ← 그대로
```

프로필 질문 세트(새 파일, 게이트 앞 정의 가능):

| 순번 | id | 형식 | 그룹 | 초기값 출처 |
|---|---|---|---|---|
| 1 | `conversation_language` | select | profile-language | `existingPrefs.ConversationLang`, 없으면 `en` |
| 2 | `user_name` | input | profile-identity | `existingPrefs.UserName` |
| 3 | `git_commit_lang` | select | profile-languages | 저장값, 없으면 `en` |
| 4 | `code_comment_lang` | select | profile-languages | 저장값, 없으면 `en` |
| 5 | `doc_lang` | select | profile-languages | 저장값, 없으면 `en` |
| 6 | `model` | select | profile-model | `normalizeModel(existingPrefs.Model)` |
| 7 | `model_policy` | select | profile-model | 저장값 |
| 8 | `effort_level` | select | profile-model | 저장값 |
| 9 | `permission_mode` | select | profile-model | 저장값, 비었으면 `acceptEdits` |
| 10 | `development_mode` | select | profile-project | 프로젝트 `quality.yaml`, 프로젝트 밖이면 빈 값 |

보이는 질문 수 N 은 10 이다(조건부 질문 없음). 대화 언어를 첫 그룹에 혼자 두는 이유: 언어 답이 저장되는 시점(필드 Blur)이 다음 그룹이 그려지기 전이어야 REQ-ITI-007 이 성립한다. 그룹 라벨은 화면에 그려지지 않으므로 페이지 나눔만 정한다.

### §2.3 기각한 대안

- **init 질문 정의(`DefaultQuestions`)를 재사용** — `conversation_language` 는 같지만 나머지 필드가 다르고, init 쪽 질문을 조건부로 섞으면 t583 구간과 겹친다. 대화 언어 질문의 옵션 목록(라벨·설명)만 공유한다.
- **v1 폼을 두고 테마·키맵만 v2 에 맞춤** — v1 import 가 남아 REQ-ITI-003 을 만족하지 못한다.

## §3 옵션 전달 모양

`schemaSelectOptions` 는 버전 중립 구조를 돌려준다.

```go
// internal/cli (모양만 제시, 이름은 run 단계에서 확정)
type selectOption struct {
    Label string
    Value string
}
```

- 라벨 해석(`optionLabelFor` → `schemaOptionBridge`)은 `cli` 에 남는다.
- 위저드에는 필드 id 별 `[]{Label, Value}` 를 인자로 넘기고, 위저드가 v2 `huh.Option` 으로 바꾼다. 위저드 쪽 인자 타입은 `wizard` 패키지의 기존 `Option`(`questions.go` 의 `{Label, Value, Desc}`)을 쓴다. `cli` 가 `wizard.Option` 으로 바로 만들어도 import 방향은 지켜진다.
- `model_policy` 는 스키마에 없으므로 `cli` 에서 `settings.EmptyLabelFor("model_policy")` 빈 옵션 + `template.ValidModelPolicies()` 각 값(라벨은 `t.ModelPolicyHigh` 등)으로 만든다.
- 언어 필드 네 개의 값은 `settings.FieldOptionDefs("conversation_lang")` 등에서 오고, 라벨은 init 위저드 `conversation_language` 질문과 같은 원어 이름(`English`, `Korean (한국어)`, `Japanese (日本語)`, `Chinese (中文)`)과 설명을 쓴다. 원어 이름은 번역하지 않는 현재 설계(`questions.go:59-61` 주석)를 따른다.

기각: 위저드가 `settings` 를 import 해 스스로 옵션을 푸는 안 — 라벨 브리지가 `cli` 의 `profileSetupText` 에 묶여 있어 결국 `cli` 를 필요로 한다.

## §4 결과 구조와 스테퍼

현재 `stepperNote(questions, first, result *WizardResult)`, `stepperDenominator`, `visibleQuestionIndex` 는 `*WizardResult` 에 묶여 있다(`wizard.go:236-270`). huh v2 `TitleFunc` 의 재계산 바인딩도 결과 구조체 포인터를 쓴다.

결정: 스테퍼가 결과 구조체 대신 **가시성 판정 클로저와 바인딩 대상**을 받도록 일반화한다. init 호출부는 기존 `*WizardResult` 를 넘기는 얇은 감싸개로 바꾸어 출력 문자열이 그대로다(AC-ITI-009 의 init 대조군이 이를 검사). 프로필 위저드는 자기 결과 구조체(새 파일)를 바인딩 대상으로 넘긴다. 스테퍼 문자열 자체는 두 쪽 모두 `tui.Stepper(k, N, nil)` 을 부른다.

이 편집은 `wizard.go` 의 스테퍼 구간(236·242, t583 비접촉 목록)에 닿으므로 게이트 뒤 M5 에서 한다.

기각: 프로필 위저드가 `WizardResult` 에 필드를 보태 재사용 — `types.go` 는 t583 이 필드를 지우는 파일이고, init 결과와 프로필 결과가 한 구조체에 섞이면 저장 경로 판정이 흐려진다.

## §5 취소와 오류 매핑

- v2 폼 오류는 `mapFormErr`(`wizard.go:135-140`)가 `huh.ErrUserAborted` → `wizard.ErrCancelled` 로 바꾼다. 프로필 실행 함수도 같은 매핑을 쓴다.
- `runProfileSetup` 은 `errors.Is(err, wizard.ErrCancelled)` 로 취소를 판별해 `t.SetupCancelled`(선택된 로케일)를 stdout 에 쓰고 nil 을 돌려준다. 1단계 취소 시 영어 고정 `Setup cancelled.`(`profile_setup.go:339`)를 쓰던 경로는 없어진다 — 취소 시점의 로케일은 그때까지 답한 언어, 답하기 전이면 저장값 또는 `en` 이다.
- 명령 수준 테스트를 위해 위저드 실행과 세션 worktree 진입·정리(`enterSessionWorktree`, `cleanupSessionWorktree`)를 패키지 함수 변수 이음새로 뺀다. 기존 `runWizardFn`(`init_update_notice.go:69`)과 같은 방식이다.

## §6 다운그레이드 확인창

- `wizard` 패키지에 확인창 하나짜리 폼을 만드는 헬퍼(새 파일)를 둔다. 인자: 로케일, 현재 버전, 대상 태그, 바인딩 대상. 제목·설명 문구는 위저드 번역에 4개 로케일로 둔다(게이트 전에는 새 파일, M7 에서 `translations.go` 로 합침). 버튼 라벨은 기존 `ConfirmYes`/`ConfirmNo`, 도움말은 §7 키맵.
- 언어 해석은 `cli` 쪽 함수로 둔다:

```
resolveDowngradeLocale(cwd):
  if dir(cwd/.moai) exists:
     v := wizard.ReadLocaleFromProject(cwd)   // config_helpers.go:18
     if v != "" → return v
  v := profile.ReadPreferences(profile.GetCurrentName()).ConversationLang
  if v != "" → return v
  return "en"
```

- `update_version.go` 는 이 로케일로 헬퍼를 부르고, 폼 오류 처리(`downgrade confirmation: %w`)와 거절 시 `Downgrade aborted` 알약 출력은 그대로 둔다.

기각: 프로필 우선 — 리드 판정 Q3 이 프로젝트 우선을 정했다. 시스템 로케일 감지 — 저장소에 감지 코드가 없고(판정 문서) 새 규칙을 만드는 일이라 범위 밖이다.

## §7 도움말 줄 키별 라벨

huh v2 도움말 줄은 활성 필드 키맵 바인딩의 `WithHelp(키, 동작)` 에서 동작 문자열을 모아 그린다(`keymap.go:111-183`). 로케일별 `huh.NewDefaultKeyMap()` 을 만들어 표면에 나타나는 바인딩마다 `SetHelp(키, 현지화 동작)` 을 적용하고 `Form.WithKeyMap` 으로 건다. 키 표기(`enter`, `↑` 등)는 바꾸지 않는다.

| 동작 (en) | ko | ja | zh |
|---|---|---|---|
| next | 다음 | 次へ | 下一步 |
| submit | 제출 | 送信 | 提交 |
| back | 이전 | 戻る | 返回 |
| select | 선택 | 選択 | 选择 |
| up | 위 | 上 | 上 |
| down | 아래 | 下 | 下 |
| filter | 검색 | 絞り込み | 筛选 |
| set filter | 검색 적용 | 絞り込み確定 | 应用筛选 |
| clear filter | 검색 해제 | 絞り込み解除 | 清除筛选 |
| toggle | 전환 | 切替 | 切换 |
| complete | 자동 완성 | 補完 | 补全 |
| Yes / No (확인창 `y`/`n` 동작) | `ConfirmYes` / `ConfirmNo` 값 | 같음 | 같음 |

- 이 표가 번역 키의 원천이다. run 단계에서 네 표면을 그려 실제로 나타나는 동작 라벨을 모두 모으고, 표에 없는 라벨이 나오면 표를 채운 뒤 골든을 뜬다(AC-ITI-013 은 표에 없는 라벨을 실패로 친다).
- 문장형 `HelpSelect`/`HelpInput` 키와 값, 그 테스트는 지운다(Q2).
- 문구는 로케일 원어 표현 기준으로 run 단계에서 다듬을 수 있으나, 다듬은 결과가 이 표를 대체하는 새 원천이 되고 골든과 함께 커밋한다.

## §8 레이아웃 조정

| 결함 | 조정 | 위치 |
|---|---|---|
| 확인 버튼 중앙정렬 | `(*Confirm).WithButtonAlignment(lipgloss.Left)` — `buildConfirmField` 와 다운그레이드 헬퍼 | `wizard.go:487~`, 새 헬퍼 |
| 필드 사이 빈 줄 | 위저드 테마 `Styles.FieldSeparator = "\n"` | `moaiWizardStyles`(테마 구간) |
| 선택 필드 아래 빈 카드 줄 | `(*Select).Height(n)`, n 은 옵션 수와 터미널 높이 중 작은 값에 제목·설명 줄을 더한 값 | `buildSelectField` |
| 설명 열 불일치 | 옵션마다 `runewidth.StringWidth(label)` 로 표시 폭을 재고, 최대 폭(터미널 폭을 넘지 않게 자름)에 맞춰 공백을 채운 뒤 설명을 붙인다. `key = opt.Label + " - " + opt.Desc`(`wizard.go:290`) 대체 | `buildSelectField` |

확인창 안쪽 빈 줄(`field_confirm.go:261-264`)은 손대지 않는다(D3).

## §9 t583 뒤 그룹 재구성

t583 뒤 init 질문 4개는 라벨 기준으로 huh 그룹 3개(Basic 2 / Quality & Workflow 1 / Autonomy 1)가 된다. 그룹 라벨은 화면에 그려지지 않는다(`wizard.go:183-186`).

| 안 | 사용자가 보는 결과 | 판단 |
|---|---|---|
| 라벨만 바꿈 | 3페이지, 2·3페이지에 질문 하나씩. 화면은 지금과 같음 | 기각 — 코드 주석·테스트만 정리되고 한 문항짜리 페이지가 둘 남는다 |
| 그대로 둠 | 같음 | 기각 — 라벨이 내용과 어긋나는 상태를 방치 |
| **`agent_wiring` 과 `autonomy_tier` 를 한 그룹으로** | 2페이지. 첫 페이지 대화 언어·이름, 둘째 페이지 하네스·자율성 | **채택** |

결정: 두 질문의 `Group` 을 `Agents & Autonomy` 로 바꾼다. 두 질문 모두 조건이 없어 `buildFormGroups` 가 한 그룹으로 묶는다. 스테퍼 분모는 질문 수 기준이라 4 로 그대로다. 영향: `questions.go` 두 리터럴의 `Group` 값(t583 이 "건드리지 않음"으로 남긴 부분이라 게이트 뒤 편집), `agent_wiring_question_test.go:87-88` 의 그룹 단정, 패키지 주석의 페이지 설명(`questions.go:24-26`, `wizard.go:36`). 이 결정은 리드 판정 Q5 로 확정됐다(`plan.md` §H). 판정은 셋으로 나눈다: 페이지 수 2 는 AC-ITI-018, 스테퍼 분모 4 는 AC-ITI-021, 라벨이 화면에 나오지 않음은 AC-ITI-022. 번역 키: 비테스트 코드에서 질문의 `Group` 을 읽는 곳은 묶기 비교와 대입 두 줄(`wizard.go:183`, `:186`)뿐이고 `buildFormGroups` 는 그룹에 제목을 붙이지 않아(`wizard.go:172`) 라벨을 그리는 경로가 없다(`research.md` §6.1). 따라서 네 로케일 번역 표에 그룹 라벨 키를 두지 않는다.

## §10 소스 스캔 가드 재조준

폼이 옮겨 가면 값 바인딩은 `Value(&x)` 가 아니라 질문 정의와 결과 필드 대입으로 표현된다. 가드마다 성질을 보존하는 쪽으로 옮긴다.

| # | 테스트 | 처분 | 새 대상 |
|---|---|---|---|
| S1 | `TestProfileSetup_ModelPolicySelectPresent` | 행동 테스트로 대체 | 프로필 질문 세트에 `model_policy` 질문이 있고 옵션 값이 정책 3개 + 빈 값, 저장 결과 `preferences.yaml` 에 `model_policy` 가 기록됨(AC-ITI-005 와 공유 가능) |
| S2 | `TestTUINestedConfigNoParallelWriter` | 스캔 유지 | 저장 경로가 남는 `profile_setup.go` 와 프로필 위저드 새 파일 모두. 기준 문자열 `persistProjectConfig` 는 `profile_setup.go` 에서 |
| S3 | `TestPermissionModeNormalizeAcceptEdits` | 스캔 유지 + 행동 보강 | `profile_setup.go`(정규화는 저장 쪽에 남음). AC-ITI-006 (2) 가 행동을 판정 |
| S4 | `TestTUIEmptyLabelsSchemaSourced` | 스캔 대상 이동 | `model_policy` 옵션을 만드는 `cli` 파일(§3). 앞 절의 옵션 목록 검사는 `.Label` 로 |
| S5 | `TestProfileSetupConstructsProjectSelects` | 행동 테스트로 대체 | 프로필 질문 세트에 `development_mode` 질문, 옵션 값에 `ddd`·`tdd` |
| S6 | `TestTUIRendersSchemaFieldSet` (b) | 행동 테스트로 대체 | 프로필 질문 id 집합이 §2.2 표 10개와 같음. (a) 절은 그대로 |
| S7 | `TestWizardOmitsRemovedQuestions` | 스캔 대상 확장 + 표식 추가 | `profile_setup.go` 와 프로필 위저드 새 파일. 기존 표식에 v2 형태 표식(`ID: "statusline_theme"`, `ID: "statusline_segments"`, `ID: "git_convention"`, 중첩 필드 id)을 더한다 |
| S8 | `TestWizardWritesNoStatuslineTheme` | 스캔 유지 | `profile_setup.go`(저장 구조체가 남음) + 프로필 위저드 새 파일 |
| S9 | `TestWizardCarriesStoredSegmentsIntoPrefs` | 스캔 유지 | `profile_setup.go`(저장·동기화가 남음). AC-ITI-006 (4)(6)(7) 이 행동을 판정 |

모든 스캔형 가드는 파일을 읽은 직후 그 파일의 기준 문자열 존재를 단정한다(AC-ITI-010 (2)). 테스트 함수 이름은 가능한 한 유지해 선택자가 9개를 계속 잡게 한다.

## §11 pty 하네스

- **게이트 변수**: `MOAI_PTY_CAPTURE`(테스트 전용 상수로 한 곳에 둔다). 판정 문서 프로브의 `MOAI_T586_TTY` 는 카드 id 가 이름에 들어가 제품 테스트에 쓰지 않는다.
- **자식**: `go test -c -o <t.TempDir()>/cli.test ./internal/cli` 로 빌드. 도우미 테스트는 `MOAI_PTY_CAPTURE_CHILD=<사례 이름>` 일 때만 실행되어 사례별 제품 경로를 부른다. 네트워크 이음새(`init_update_notice.go` 의 이음새들)는 도우미 안에서 막는다.
- **세션**: `tmux new-session -d -s <이름> -x 80 -y 30 -c <임시 디렉터리> -e HOME=<임시>/home -e MOAI_HOME=<임시>/moai-home -e CLAUDE_CONFIG_DIR= -e MOAI_KANBAN…=(접두 변수 9개 각각 빈 값) -e TERM=xterm-256color -e MOAI_PTY_ENV_CANARY=<난수> -e MOAI_PTY_ENV_OUT=<임시>/child-env.txt '<바이너리> -test.run …'`. `-e` 인자는 `acceptance.md` §B 자식 환경 정리 목록 한 곳에서 만든다. `-e` 를 쓰는 이유: tmux 는 서버를 띄울 때의 환경을 전역 환경으로 복사하고, 새 창의 초기 환경은 세션 환경과 전역 환경을 합친 것이다(tmux 3.6a man, GLOBAL AND SESSION ENVIRONMENT, `research.md` §15). 이미 떠 있는 서버에서는 테스트 프로세스의 환경이 자식에 닿지 않는다. 이름은 `moai-ptycap-<테스트명>-<난수>` 이고, 세션 이름은 이름 생성 함수 하나만 만든다. 생성 직후 `t.Cleanup(kill-session -t =<이름>)`.
- **대기**: `capture-pane -p -t =<이름>` 을 100ms 간격으로 되풀이, 기준 문자열 발견 시 그 캡처를 판정 대상으로 쓴다. 기한 10초, 초과 시 마지막 캡처를 `t.Fatalf` 메시지에 담는다.
- **키 입력**: `send-keys -t =<이름> <키>`. 입력 뒤 다음 기준 문자열을 다시 기다린다.
- **HOME 감시**: `acceptance.md` §B P8 감시 목록(W1~W6)만 비교한다. 목록은 테스트 전용 상수 한 곳에 두고 항목마다 쓰는 제품 함수를 주석으로 단다. 트리 전체를 `filepath.WalkDir` 로 해시하지 않는다 — 2회차 감사 측정에서 실제 `~/.moai` 는 파일 43,918개였고 5분 사이 63개가 다른 세션 때문에 바뀌어, 전후 비교가 수리와 무관하게 달라진다. 비교 함수는 루트 경로를 인자로 받아 AC-ITI-020 (4) 양성 대조군이 가짜 HOME 에서 같은 함수를 돈다.
- **실효 환경 기록**: 자식 도우미가 제품 경로를 부르기 전에 정리 목록 변수를 `MOAI_PTY_ENV_OUT` 파일에 쓰고, 부모가 캡처를 판정하기 전에 읽어 `acceptance.md` §B 실효 환경 관측의 네 단정을 한다.
- **자기 검증**: 강제 실패·강제 기한 초과 하위 테스트는 `MOAI_PTY_CAPTURE_SELFTEST=<fail|timeout>` 일 때만 실행되어, 부모 테스트가 자식 `go test` 로 돌려 결과와 잔존 세션을 검사한다(AC-ITI-020).

## §12 결정 요약

| 결정 | 선택 | 핵심 이유 |
|---|---|---|
| init·update 의 프로필 경로 | 확인창·호출 모두 제거 | 리드 판정 Q1, 대화 언어 1회 |
| 흡수 방식 | 프로필 전용 질문 세트 + 기존 v2 엔진 | v1 import 0, t583 구간 비접촉 |
| 옵션 전달 | `cli` 에서 버전 중립 목록 → 위저드 인자 | import 방향, 라벨 브리지 위치 |
| 스테퍼 | 가시성 클로저로 일반화, 같은 `tui.Stepper` | 리드 판정 Q4, init 출력 불변 |
| 취소 | `wizard.ErrCancelled` | v1 센티널 제거 |
| 다운그레이드 언어 | 프로젝트 → 프로필 → 영어 | 리드 판정 Q3 |
| 도움말 | 키별 짧은 라벨 표, 문장형 삭제 | 리드 판정 Q2 |
| 테마 | v2 팩토리 하나 | v1 소비자 0, SPEC-CLI-TUI-MODERNIZE-001 절 인수 |
| 그룹 | `agent_wiring`·`autonomy_tier` 를 `Agents & Autonomy` 한 그룹으로(리드 판정 Q5) | 라벨은 그려지지 않음, 한 문항 페이지 제거 |
| 가드 | 성질 보존 재조준 + 뮤턴트 | 공허 방지 |
| 판정 | pty(수리) + 골든(회귀) | D6 |
