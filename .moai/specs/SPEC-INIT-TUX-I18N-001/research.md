# SPEC-INIT-TUX-I18N-001 — 조사

## §0 출처와 재확인 범위

- 측정 트리: 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, HEAD `18144b7aca714ea8924363b1eab4640cf101c6d0`(착수 시 `git rev-parse HEAD` 로 확인).
- 이전 조사 트리 `e7a7d4bb3` 와의 관계: `git diff --stat e7a7d4bb3 HEAD -- internal cmd pkg go.mod go.sum` 출력 없음(종료 0). 사이의 커밋 두 개(`d0ec7921f`, `18144b7ac`)는 SPEC 문서와 감사 기록뿐이다. 따라서 이전 조사 좌표는 이 트리에서도 유효하다.
- 1회차 감사(`.moai/reports/t586/plan-audit.md`)가 인용한 좌표 가운데 이 SPEC 이 쓰는 것은 이 트리에서 다시 쟀다. 다시 재지 않은 것은 §13 에 적는다.
- 명령은 모두 zsh 에서 실행했다. 글롭은 `git grep` 경로 인자로만 넘겨 셸 글롭 확장 실패(`no matches found`)를 피했다.
- 3회차 개정 측정 트리: HEAD `538b56f1923c7b72e8dcb8379d55d05e4fadb1c5`(로컬 develop 흡수 뒤). `git diff --stat 18144b7aca714ea8924363b1eab4640cf101c6d0 HEAD -- internal cmd pkg go.mod go.sum` 출력은 파일 56개다. 이 SPEC 이 좌표를 인용하는 제품 파일 가운데 바뀐 것은 `internal/cli/update.go` 이고, 인용한 확인창 구간 `update.go:172-192`(제목 `:179`, `runProfileSetup` 호출 `:187`)와 `init.go:651`·`:659` 는 이 트리에서 다시 읽어 `acceptance.md` §D.3 L2·L4 원문과 같음을 확인했다. `update_wizard.go`·`update_tux.go` 도 바뀌었으나 이 SPEC 이 두 파일의 좌표를 인용하지 않는다. 측정 당시 `grep -n -E 'update_wizard|update_tux' .moai/specs/SPEC-INIT-TUX-I18N-001/*.md` 는 종료 1 이었다. 이 문장이 두 이름을 담은 뒤로는 이 줄 자체가 적중하므로(v0.2.3 개정 트리 `268cffe2c` 에서 이 한 줄, 종료 0), 다시 잴 때는 이 파일을 뺀 다섯 문서(`spec.md`·`plan.md`·`acceptance.md`·`design.md`·`progress.md`)를 대상으로 한다. 같은 트리에서 그 형태는 출력 없이 종료 1 이고, 같은 파일 묶음에 `grep -c 'REQ-ITI-017'` 을 돌린 대조군은 네 파일에서 1 이상이었다. `go.mod` 은 바뀌지 않았다. §8.1, §14~§16 은 이 트리에서 쟀다.

## §1 v1 표면

```
$ git grep -l '"github.com/charmbracelet/huh"' -- '*.go' ':!*_test.go'
internal/cli/huh_theme.go
internal/cli/init.go
internal/cli/profile_setup.go
internal/cli/update.go
internal/cli/update_version.go
exit=0
$ git grep -l '"charm.land/huh/v2"' -- '*.go' ':!*_test.go'
internal/cli/wizard/wizard.go
exit=0
```

```
$ git grep -n 'runProfileSetup(' -- internal/cli/init.go internal/cli/update.go internal/cli/profile.go internal/cli/profile_setup.go
internal/cli/init.go:659:			if err := runProfileSetup(cmd, nil); err != nil {
internal/cli/profile.go:62:		return runProfileSetup(cmd, args)
internal/cli/profile_setup.go:240:func runProfileSetup(cmd *cobra.Command, args []string) (err error) {
internal/cli/update.go:187:				if err := runProfileSetup(cmd, nil); err != nil {
```

`profile_setup.go:214` 는 `RunE: runProfileSetup`(괄호 없음) 등록이다.

`huh.` 사용 줄(`grep -n 'huh\.' internal/cli/{profile_setup,init,update,update_version}.go`): `profile_setup.go` 는 `:124,126,129,133`(`schemaSelectOptions`), `:320-329`(언어 폼), `:338`, `:354-436`(본 폼), `:445`; `update_version.go:327,331`; `init.go:650,657`; `update.go:178,185`.

v1 을 import 하는 테스트는 `huh_theme_test.go` 하나다(`git grep -l '"github.com/charmbracelet/huh"' -- '*_test.go'`).

## §2 소스 스캔 가드

```
$ grep -rn 'ReadFile("profile_setup.go")' --include='*_test.go' internal/cli
internal/cli/profile_setup_projectconfig_test.go:145
internal/cli/profile_setup_model_policy_test.go:38
internal/cli/profile_setup_removed_questions_test.go:61
internal/cli/profile_setup_removed_questions_test.go:87
internal/cli/profile_setup_removed_questions_test.go:116
internal/cli/schema_bridge_test.go:121
internal/cli/profile_setup_nested_test.go:27
internal/cli/profile_setup_nested_test.go:80
internal/cli/profile_setup_nested_test.go:117
```

9건. 각 테스트 본문을 읽어 단정하는 성질을 `spec.md` §A.5 표로 옮겼다. 세 음성 가드(S7·S8)와 S2 는 `nonCommentLines` 로 주석 줄을 뺀 본문을 본다(`profile_setup_nested_test.go:46-56`). S7 의 표식 목록 `removedWidgetMarkers` 는 `profile_setup_removed_questions_test.go:39-54` 에 있다.

## §3 v1 타입 노출과 취소 판별

```
$ grep -rn 'schemaSelectOptions' --include='*_test.go' internal/cli
internal/cli/profile_setup_projectconfig_test.go:160
internal/cli/profile_setup_schema_options_test.go:49
internal/cli/profile_setup_schema_options_test.go:68
internal/cli/profile_setup_schema_options_test.go:92
internal/cli/profile_setup_schema_options_test.go:108
internal/cli/profile_setup_nested_test.go:107
```

(`profile_setup_projectconfig_test.go:154` 는 주석.) 호출 6곳, 파일 3개. 1회차 감사는 "4개 파일"이라 적었으나 이 트리에서 센 값은 3개다.

```
$ grep -rn 'ErrUserAborted' --include='*.go' internal/cli
internal/cli/profile_setup.go:338
internal/cli/profile_setup.go:445
internal/cli/wizard/unified_form_test.go:269
internal/cli/wizard/unified_form_test.go:270
internal/cli/wizard/wizard.go:136
```

`wizard.go:135-140` `mapFormErr` 가 v2 `huh.ErrUserAborted` 를 `ErrCancelled` 로 바꾼다.

인라인 언어 옵션: `profile_setup.go:320-325`(`English`, `Korean (한국어)`, `Japanese (日本語)`, `Chinese (中文)`, 설명 없음). 스키마 언어 옵션: `internal/settings/schema.go:214-221` `languageOptions()` 값 `en`·`ko`·`ja`·`zh`, 네 언어 필드 공통(`:325-344`).

## §4 저장 경로와 보존 동작

`profile_setup.go` 에서 읽은 동작(좌표는 `grep -n` 으로 확인):

| 동작 | 위치 |
|---|---|
| 빈 권한 모드 → `defaultPermissionMode` 초기값 | `:299-301` |
| 개발 방식 초기값을 `quality.yaml` 에서 | `:310-318` |
| acceptEdits → 빈 값 정규화와 확인 줄 | `:456-459` |
| `StatuslineSegments: existingPrefs.StatuslineSegments` | `:479` |
| `StatuslineTheme` 0 값 유지(주석) | `:480-486` |
| `profile.WritePreferences` | `:489` |
| 프로젝트 안일 때 `syncPrefs.StatuslineSegments = nil` 후 `SyncToProjectConfig` | `:508-509` |
| `persistProjectConfig(cwd, developmentMode, "")` | `:520` |
| 저장 문구와 `printProfileSummary` | `:526-531` |

`internal/profile/sync.go:17-77` `SyncToProjectConfig` 는 `user` 섹션(이름), `language` 섹션(`ConversationLanguage`, `ConversationLanguageName`, `GitCommitMessages`, `CodeComments`, `Documentation`)을 설정 관리자로 저장하고, 테마가 비어 있지 않거나 세그먼트 맵이 nil 이 아닐 때만 `statusline.yaml` 을 쓴다.

## §5 huh v2 공개 API 와 키맵 기본 라벨

모듈 캐시 `charm.land/huh/v2@v2.0.3` 에서 확인. API 좌표는 `plan.md` §B. 키맵 기본 도움말(`keymap.go`, `WithHelp(키, 동작)`):

- Input `:111-114` — `complete`(ctrl+e), `back`, `next`, `submit`
- Select `:140-153` — `back`, `select`(enter), `submit`, `up`, `down`, `left`, `right`, `filter`, `set filter`, `clear filter`, `½ page up`, `½ page down`, `go to start`, `go to end`
- Confirm `:178-183` — `back`, `next`, `submit`, `toggle`(←/→), `Yes`(y), `No`(n)
- 비활성 기본값(`WithDisabled`)인 바인딩은 도움말에 나오지 않는다.

판정 문서 캡처에 나온 도움말: v2 확인 그룹 `←/→ toggle • enter next • y 예 • n 아니오`, v1 확인창 `←/→ toggle • enter submit • y Yes • n No`.

## §6 질문 집합과 그룹

```
$ grep -n '^func ' internal/cli/wizard/questions.go
48:func DefaultQuestions(projectRoot string) []Question {
162:func GitQuestions() []Question {
268:func ReconfigureQuestions(projectRoot string) []Question {
296:func InitQuestions(projectRoot string) []Question {
307:func FilteredQuestions(questions []Question, result *WizardResult) []Question {
318:func TotalVisibleQuestions(questions []Question, result *WizardResult) int {
329:func QuestionByID(questions []Question, id string) *Question {
349:func Page3Questions(projectRoot string) []Question {
```

```
$ grep -n 'func saveAnswer\|"project_mode"\|"project_continuation"\|"audit_model"\|func saveBoolAnswer\|func buildConfirmField\|func stepperDenominator\|func visibleQuestionIndex' internal/cli/wizard/wizard.go
236:func stepperDenominator
242:func visibleQuestionIndex
397:func saveAnswer
434:	case "project_mode":
440:	case "project_continuation":
442:	case "audit_model":
467:func saveBoolAnswer
487:func buildConfirmField
```

t583 lane-1 확정 범위(`spec.md` §A.7)의 시작 좌표와 모두 일치한다. `wizard/translations.go` 의 `var uiStrings` 는 540 줄이다(t583 보고값 548 과 다름 — 표 안의 어느 항목을 가리키는지 확인하지 않았다).

`Page3Questions` 의 id·형식(`awk` 로 `ID:`·`Type:` 줄 추출): `project_mode` select, `worktree_auto_create` confirm, `todo_enabled` confirm, `feedback_auto_submit` confirm, `project_continuation` select, `audit_model` select, `audit_gate_claude`·`audit_gate_codex`·`audit_gate_glm` select, `codex_audit_enabled` confirm, `agent_wiring` select, `mcp_provision` confirm, `autonomy_tier` select. 확인형 5개가 모두 t583 삭제 목록 11개에 들어 있다. `DefaultQuestions`·`GitQuestions` 에는 확인형이 없다(select·input 만).

> **§6.1 로 대체됨.** 아래 문단의 판독은 POSIX ERE 에서 단어 경계가 아닌 표기를 쓴 grep 결과에 기대므로 근거로 쓰지 않는다. 같은 결론을 경계 없는 패턴과 대조군으로 다시 잰 기록은 §6.1 이다.

그룹 라벨 사용: `git grep`/`grep -rn '\.Group\b'` 결과 위저드 비테스트 코드에서 `q.Group` 을 읽는 곳은 `wizard.go:183`, `:186`(묶기 비교)뿐이고, `t.Group.Title` 등은 테마 스타일 필드다. 그룹 라벨을 번역하는 표도 없다(`grep -n 'Quality & Workflow' internal/cli/wizard/translations.go` 0건). `agent_wiring`(`questions.go:490-501`)과 `autonomy_tier`(`:513-525`)에는 `Condition` 이 없다.

### §6.1 그룹 라벨 렌더 경로 재측정 (2회차 개정)

측정 트리 HEAD `d8ebb39298b206ec8cc4183e728f27f48994df86`. `git diff --stat 18144b7aca714ea8924363b1eab4640cf101c6d0 HEAD -- internal cmd pkg go.mod go.sum` 출력 없음(종료 0)이라 §0 의 좌표가 그대로 유효하다. 바로 위 문단의 `\.Group\b` 형태는 POSIX ERE 에서 `\b` 가 단어 경계가 아니어서 빈 출력이 공허할 수 있다(이번에 같은 형태를 다시 돌렸더니 `q.Group` 이 있는 트리에서 종료 1 로 비었다). 그래서 경계 없이 대조군과 함께 다시 쟀다.

```
$ git grep -n -E '\.Group([^A-Za-z0-9_]|$)' -- '*.go' ':!*_test.go'
internal/cli/huh_theme.go:107:	t.Group.Title = t.Focused.Title
internal/cli/huh_theme.go:108:	t.Group.Description = t.Focused.Description
internal/cli/model.go:148:			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", e.Agent, e.Group, e.Model, e.Effort, e.GLMModel, e.GLMReasoning)
internal/cli/model.go:153:			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", e.Agent, e.Group, e.Model, e.Effort)
internal/cli/root.go:121:		&cobra.Group{ID: "launch", Title: "Launch Commands:"},
internal/cli/root.go:122:		&cobra.Group{ID: "project", Title: "Project Commands:"},
internal/cli/root.go:123:		&cobra.Group{ID: "tools", Title: "Tools:"},
internal/cli/wizard/wizard.go:158:func buildFormGroups(questions []Question, result *WizardResult, locale *string) []*huh.Group {
internal/cli/wizard/wizard.go:159:	var groups []*huh.Group
internal/cli/wizard/wizard.go:183:		if len(pending) > 0 && q.Group != pendingLabel {
internal/cli/wizard/wizard.go:186:		pendingLabel = q.Group
internal/cli/wizard/wizard.go:195:func buildConditionalGroup(questions []Question, q *Question, result *WizardResult, locale *string) *huh.Group {
internal/cli/wizard/wizard.go:204:// buildQuestionGroup creates a huh.Group for a single question.
internal/cli/wizard/wizard.go:206:func buildQuestionGroup(q *Question, result *WizardResult, locale *string) *huh.Group {
internal/cli/wizard/wizard.go:587:	t.Group.Title = t.Focused.Title
internal/cli/wizard/wizard.go:588:	t.Group.Description = t.Focused.Description
internal/lsp/aggregator/aggregator.go:47:	sf           singleflight.Group
internal/lsp/core/manager.go:43:	// @MX:NOTE: [AUTO] sf — singleflight.Group; prevents duplicate clientFactory+Start calls for the same language (REQ-UTIL-003-004, REQ-UTIL-003-005)
internal/lsp/core/manager.go:44:	sf singleflight.Group
exit=0
$ git grep -n -E 'q\.Group([^A-Za-z0-9_]|$)' -- 'internal/cli/wizard/wizard.go'   # 대조군
internal/cli/wizard/wizard.go:183:		if len(pending) > 0 && q.Group != pendingLabel {
internal/cli/wizard/wizard.go:186:		pendingLabel = q.Group
exit=0
$ git grep -n -F 'Quality & Workflow' -- internal/cli/wizard/translations.go
exit=1
$ git grep -c -F 'ConfirmYes' -- internal/cli/wizard/translations.go   # 대조군
internal/cli/wizard/translations.go:6
exit=0
$ grep -n 'huh.NewGroup(fields' internal/cli/wizard/wizard.go
172:		groups = append(groups, huh.NewGroup(fields...))
```

판독: 질문의 `Group` 필드를 읽는 비테스트 줄은 `wizard.go:183`(묶기 비교)과 `:186`(대입) 둘뿐이다. 나머지 적중은 다른 타입이다 — `t.Group.Title`·`t.Group.Description`(huh 테마 스타일 필드), `e.Group`(`model.go`, 에이전트 모델 표), `cobra.Group`(`root.go`), `singleflight.Group`(`internal/lsp`), `huh.Group` 타입 이름. 라벨 문자열 `Quality & Workflow` 는 `questions.go` 의 `Group:` 값과 주석(`questions.go:26`, `:338`, `types.go:41`, `wizard.go:36`)에만 있고 번역 표에는 없다. `buildFormGroups` 는 `huh.NewGroup(fields...)` 에 제목을 붙이지 않는다(`:172`). 따라서 그룹 라벨을 그리는 경로가 없고, 새 라벨 `Agents & Autonomy` 에는 번역 키가 필요 없다. 이 판단은 AC-ITI-022 가 렌더와 grep 으로 고정한다.

t583 설계(`SPEC-INIT-QUIET-WIZARD-001/design.md` §7)는 남는 4문항 그룹을 Basic / Quality & Workflow / Autonomy 로 두고 "라벨 재구성은 렌더링 결정이라 t586 에 맡긴다"고 적었다.

## §7 언어 해석 도구

- `wizard.ReadLocaleFromProject(projectRoot)` — `config_helpers.go:18`, `.moai/config/sections/language.yaml` 의 `language.conversation_language`, 실패 시 빈 문자열.
- `profile.GetCurrentName()` — `internal/profile/profile.go:77`. `GetCurrentNameForProject` 는 `:94`.
- `profile.ReadPreferences(name)` — `internal/profile/preferences.go:117`. `IsSetup` 은 `:100`.
- 다운그레이드 확인창 호출부 `update_version.go:324-338` 는 `--yes` 거나 비대화형이면 확인을 건너뛴다(조건 `:325`).

## §8 모듈 그래프

```
$ grep -n 'huh\|bubbletea\|bubbles\|catppuccin\|lipgloss' go.mod
6:	charm.land/bubbles/v2 v2.2.1
7:	charm.land/bubbletea/v2 v2.0.9
9:	charm.land/huh/v2 v2.0.3
10:	charm.land/lipgloss/v2 v2.0.6
14:	github.com/charmbracelet/huh v1.0.0
15:	github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834
44:	github.com/catppuccin/go v0.3.0 // indirect
46:	github.com/charmbracelet/bubbles v1.0.0 // indirect
47:	github.com/charmbracelet/bubbletea v1.3.10 // indirect
```

추적 비테스트 Go 파일 가운데 `internal`·`cmd`·`pkg` 밖에 있는 것 12개(`git ls-files '*.go' | grep -v '_test\.go$' | grep -v -E '^(internal|cmd|pkg)/'`): `.moai/reports/` 아래 5개, `.moai/scripts/` 2개, `scripts/` 5개. AC-ITI-004 의 검색 범위에 포함된다.

### §8.1 tidy 뒤 남는 간접 의존성 (3회차 개정)

```
$ go mod why -m github.com/catppuccin/go github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles
# github.com/catppuccin/go
github.com/modu-ai/moai-adk/internal/cli
github.com/charmbracelet/huh
github.com/catppuccin/go

# github.com/charmbracelet/bubbletea
github.com/modu-ai/moai-adk/internal/cli
github.com/charmbracelet/huh
github.com/charmbracelet/bubbletea

# github.com/charmbracelet/bubbles
github.com/modu-ai/moai-adk/internal/cli
github.com/charmbracelet/huh
github.com/charmbracelet/bubbles/filepicker
why_exit=0
$ go mod graph | grep -E '(github.com/catppuccin/go|github.com/charmbracelet/bubbletea@v1|github.com/charmbracelet/bubbles@v1)'
github.com/modu-ai/moai-adk github.com/catppuccin/go@v0.3.0
github.com/modu-ai/moai-adk github.com/charmbracelet/bubbles@v1.0.0
github.com/modu-ai/moai-adk github.com/charmbracelet/bubbletea@v1.3.10
charm.land/huh/v2@v2.0.3 github.com/catppuccin/go@v0.2.0
github.com/charmbracelet/bubbles@v1.0.0 github.com/charmbracelet/bubbletea@v1.3.10
github.com/charmbracelet/huh@v1.0.0 github.com/catppuccin/go@v0.3.0
github.com/charmbracelet/huh@v1.0.0 github.com/charmbracelet/bubbletea@v1.3.6
graph_exit=0
```

(그래프 출력에서 왼쪽이 `bubbles@v1.0.0`·`bubbletea@v1.3.10` 자신인 의존 줄은 뺐다. 오른쪽에 세 모듈이 오는 줄은 위 7줄이 전부다.)

판독: 오른쪽에 bubbletea v1·bubbles v1 을 두는 모듈은 루트, huh v1, bubbles v1 뿐이라 huh v1 이 빠지면 둘도 빠진다. `go mod why` 는 가장 짧은 경로 하나만 보여 catppuccin 도 huh v1 경로로 나오지만, 그래프에는 `charm.land/huh/v2@v2.0.3` 이 catppuccin 을 요구하는 줄이 있다. 2회차 감사가 모듈 캐시에서 huh v2 의 import 를 확인했다(`charm.land/huh/v2@v2.0.3/theme.go:6`, 감사 E-3). 따라서 catppuccin 은 tidy 뒤에도 남는다. `go mod tidy` 는 작업 트리 `go.mod` 를 바꾸므로 실행하지 않았고, 남는 줄의 선택 버전은 확인하지 않았다.

## §9 교차 SPEC

```
$ grep -n 'REQ-TUIM-04[0-9]' .moai/specs/SPEC-CLI-TUI-MODERNIZE-001/spec.md
117: REQ-TUIM-040 … two separate factories; the library version boundary shall not be merged into a single factory.
118: REQ-TUIM-041 … Both huh theme factories shall apply the same internal/tui token-to-role assignment …
122: REQ-TUIM-045 … Both factories shall continue to resolve the light/dark axis through their existing package-level indirection variables …
$ grep -n 'AC-TUIM-02[6-9]' .moai/specs/SPEC-CLI-TUI-MODERNIZE-001/acceptance.md
116: AC-TUIM-026 … grep -n "func moaiHuhStyles" internal/cli/huh_theme.go → 1 hit AND grep -n "func moaiWizardStyles" internal/cli/wizard/wizard.go → 1 hit
119: AC-TUIM-029 … grep -n "var huhThemeIsDark" internal/cli/huh_theme.go → 1 hit AND grep -n "var wizardIsDark" internal/cli/wizard/wizard.go → 1 hit
```

(출력은 줄 앞부분을 줄였다.) 인수 내용은 `spec.md` §A.6.1.

## §10 명령 계약 경로

`runProfileSetup`(`profile_setup.go:240-265`): `loadSessionWorktreeConfig` → `enterSessionWorktree(swCfg, "profile", stderr)` → 경로가 있으면 `emitProfileScopeNotice` 와 `os.Chdir` → `defer cleanupSessionWorktree(swCfg, wtPath, err == nil, stderr)` → 이름 기본값 `default` → `profile.ReadPreferences`. 기존 테스트 가운데 이 함수를 명령 수준에서 몰아가는 것은 없다(`profile_worktree_test.go` 는 헬퍼 직접 호출, `profile_setup_summary_test.go` 는 `printProfileSummary` 만 — 1회차 감사 D7 인용, 이번에 파일 존재만 재확인).

기존 이음새: `isInteractiveStdin`(`init_update_notice.go:62`), `runWizardFn`(`:69`). `update.go` 의 확인창 조건은 `isatty.IsTerminal(os.Stdin.Fd())` 를 직접 부른다(`update.go:174`).

## §11 init 흐름

`init.go:647-663` 확인창 → `:665-686` 프로필 값을 `opts` 기본값으로 → `:696` 대화형이면 배너 → `:704` `runWizardFn(rootFlag, opts.ConvLang, opts.UserName)` → 취소 시 `:707` `Initialization cancelled.`(stderr) → `:719-726` 위저드 답이 프로필 값보다 우선. 프로필이 없어도 init 위저드가 `conversation_language` 를 첫 질문으로 묻는다(`questions.go:55-75`, 주석 `:56-61`).

## §12 테스트 파일 목록 (영향 후보)

`internal/cli`: `huh_theme_test.go`, `profile_setup_acceptEdits_test.go`, `profile_setup_model_policy_test.go`, `profile_setup_nested_test.go`, `profile_setup_normalize_test.go`, `profile_setup_projectconfig_test.go`, `profile_setup_removed_questions_test.go`, `profile_setup_schema_options_test.go`, `profile_setup_summary_test.go`, `profile_setup_translations_test.go`, `profile_worktree_test.go`, `schema_bridge_test.go`, `update_version_test.go`. `internal/cli/wizard`: `wizard_test.go`(`HelpSelect` 참조), `agent_wiring_question_test.go`(그룹 단정 `:87-88`), `expansion_test.go`(t583 소관 갱신 대상).

## §13 확신도와 공백

이 절은 기록으로 남긴다. 아래 항목 가운데 실행으로 확인해야 하는 네 가지는 run 단계 첫 마일스톤의 착수 검증으로 옮겨 그곳에서 추적한다(`plan.md` §F M1, V-a~V-d). 해당 줄 끝에 옮긴 곳을 적었다.

- t583 미커밋 변경은 이 트리에 없다. 범위는 lane-1 이 plan.md §F.1 로 확정한 목록이며, 시작 좌표만 이 트리에서 대조했다. 병합된 실제 diff 는 게이트에서 다시 본다.
- `uiStrings` 좌표 548(t583)과 540(이 트리)의 차이는 원인을 확인하지 않았다. → `plan.md` M1 V-a
- huh v2 에서 로케일별 `KeyMap` 을 폼마다 거는 방식이 확인형 `y`/`n` 도움말 라벨(`Accept`/`Reject` 바인딩)까지 바꾸는지는 실행으로 확인하지 않았다. 판정 문서의 가능성 프로브는 `Toggle` 한 바인딩만 확인했다. → `plan.md` M1 V-b
- 스테퍼 일반화(`design.md` §4)가 huh v2 `TitleFunc` 재계산 바인딩과 호환되는지는 실행으로 확인하지 않았다. → `plan.md` M1 V-c
- `template.ModelAliasPickerValues()` 의 현재 값 목록은 읽지 않았다. AC-ITI-008 예외 목록은 값 대신 함수 이름으로 닫았다.
- pty 하네스 설계(`design.md` §11)의 강제 실패 자식 실행 방식은 이 트리에서 시험하지 않았다. → `plan.md` M1 V-d
- `SPEC-INIT-QUIET-WIZARD-001` 은 이 트리의 `.moai/specs/` 에 없다. HEAD `538b56f19` 에서 `ls -d .moai/specs/SPEC-INIT-QUIET-WIZARD-001` 은 `No such file or directory` 로 종료 1, 대조군 `ls -d .moai/specs/SPEC-INIT-TUX-I18N-001` 은 종료 0. t583 이 커밋되지 않았으므로 예상된 상태다. 흡수 게이트에서 병합된 경로를 다시 확인한다. → `plan.md` §C 5

## §14 실제 HOME 감시 목록 도출 (3회차 개정)

측정 트리 HEAD `538b56f19`. 2회차 감사는 실제 HOME 트리 전체 매니페스트가 이진 판정이 아니라고 봤다(`~/.moai` 파일 43,918개·8.4G, 5분 사이 63개 변동, 감사 E-9). 감시 대상을 이 SPEC 의 흐름이 실제 HOME 에서 쓸 수 있는 파일로 좁히려고, 흐름 파일에서 홈 쪽 쓰기 함수 호출을 찾고 호출 대상의 경로 계산을 읽었다.

1단계 — 흐름 파일의 홈 쓰기 호출:

```
$ git grep -n -E 'profile\.WritePreferences\(|homestate\.EnsureProjectLayout\(|[Aa]pplyAutonomyTierBundle(Fn)?\(|ensureGlobalSettingsEnv\(|globalMoaiHooksDir\(|runShellEnvConfig\(|configureShellEnv\(|RecordLastUsedProfile\(|paths\.(UserSettingsFile|UserConfigSectionsDir|ProfilesDir|StateDir|CacheDir|ReleasesDir|WorktreesDir|GlmEnvFile)\(' -- internal/cli/init.go internal/cli/update.go internal/cli/update_version.go internal/cli/profile_setup.go internal/cli/profile.go internal/core/project/initializer.go ':!*_test.go'
internal/cli/init.go:877:	if err := homestate.EnsureProjectLayout(opts.ProjectRoot); err != nil {
internal/cli/init.go:890:		if tierErr := project.ApplyAutonomyTierBundle(
internal/cli/init.go:964:	if err := ensureGlobalSettingsEnv(); err != nil {
internal/cli/profile_setup.go:489:	if err := profile.WritePreferences(profileName, prefs); err != nil {
internal/cli/update.go:205:		return runShellEnvConfig(cmd)
internal/cli/update.go:790:func runShellEnvConfig(cmd *cobra.Command) error {
internal/cli/update.go:923:func globalMoaiHooksDir(homeDir string) string {
internal/cli/update.go:931:func ensureGlobalSettingsEnv() error {
internal/cli/update.go:940:	globalHooksDir := globalMoaiHooksDir(homeDir)
internal/core/project/initializer.go:334:		if shellResult, err := i.configureShellEnv(); err != nil {
internal/core/project/initializer.go:677:func (i *projectInitializer) configureShellEnv() (*shell.ConfigResult, error) {
derive_exit=0
```

패턴에 넣은 `RecordLastUsedProfile(` 와 `paths.UserSettingsFile(` 계열은 이 여섯 파일에서 0건이다. 같은 패턴의 다른 항목이 11줄을 내므로 이 0건은 패턴이 파일을 읽지 못해서 생긴 것이 아니다.

(sync 정정, 2026-09-13, 트리 HEAD `e561b162e`) t656 (`b5b5883e9`) 이 호출을 테스트 시접 변수 `applyAutonomyTierBundleFn` 뒤로 옮겨 위 패턴의 대문자 A 리터럴 `ApplyAutonomyTierBundle\(` 가 그 호출에 더 이상 맞지 않게 됐다 — 시접 호출은 소문자 a 로 시작하고 `Bundle` 뒤에 `Fn` 이 끼어 있다. 그래서 패턴의 해당 항목을 `[Aa]pplyAutonomyTierBundle(Fn)?\(` 로 정정했고, 이 트리에서 다시 돌리면 위 레코드(538b56f19 시점)와 달리 시접 형태 `internal/cli/init.go:858` 한 줄을 포함해 같은 11줄이 나온다. 재측정 원문: `.moai/reports/t586/sync-research14-recheck.txt`. 생산자와 대상 파일은 변하지 않았다 — 같은 호출이 같은 `~/.claude/settings.json` 을 쓰므로 W1~W6 와 AC-ITI-019 는 고칠 것이 없다(`.moai/reports/t586/absorb-t583/slot/c-recheck.md`).

2단계 — 호출 대상의 경로 계산(비테스트 코드를 읽은 좌표):

| 호출 | 경로 계산 | 감시 항목 |
|---|---|---|
| `profile_setup.go:489` `profile.WritePreferences` | `internal/profile/preferences.go:90-96` `GetPreferencesPath` → `GetBaseDir()`(`internal/profile/profile.go:55-65`, `os.UserHomeDir` 기준 `~/.moai/claude-profiles`, `MOAI_HOME` 을 보지 않음). 이름 있는 프로필은 `<기준>/<이름>/preferences.yaml`. `preferences.go:144-156` `migrateOldFile` 이 옛 `.preferences.yaml` 을 이름 바꿔 옮긴다 | W1, W2 |
| `init.go:877` `homestate.EnsureProjectLayout` | `internal/homestate/paths.go:164-195`: `MOAI_HOME` 이 절대 경로이거나 프로젝트가 `os.TempDir()` 밖이면 `EnsureHomeLayout`(`:131-160`)과 `db/<키>`·`cache/search/<키>`·`run/<키>`(`:55-65`, `:112-126`). 프로젝트가 `os.TempDir()` 안이고 `MOAI_HOME` 이 없으면 `<프로젝트>/.moai/db/<키>`(`:57-58`). 키는 `ProjectKey`(`:18-34`, 정규화한 루트의 sha256 앞 4바이트) | W6 |
| `init.go:889-897` `project.ApplyAutonomyTierBundle(…, filepath.Join(homeDir, ".claude", "settings.json"), …)` | `userHomeDirFn` → `paths.Home`(HOME 우선, `internal/paths/paths.go:50-58`) | W3 |
| `init.go:964`, `update.go:931-958` `ensureGlobalSettingsEnv` | `update.go:940-943` `~/.claude/hooks/moai` 삭제, `:945` `~/.claude/settings.json` 을 읽고 필요하면 다시 씀 | W3, W4 |
| `initializer.go:333-334` `configureShellEnv`, `update.go:205`·`:790-829` `runShellEnvConfig` | `internal/shell/detect.go:128-182` `selectConfigFile`: zsh `.zshenv`·`.zshrc`, bash `.profile`·`.bash_profile`·`.bashrc`, fish `.config/fish/config.fish`, 그 밖 `.profile`, Windows PowerShell 프로필 | W5 |

뺀 것과 이유:

- `profile.RecordLastUsedProfile`(`internal/profile/profile.go:526`) → `saveLaunchLedger`(`:658-690`)가 `claude-profiles/launch.yaml` 을 쓰지만 1단계 출력에 호출이 없다. 런처가 쓰는 기록이라 다른 세션이 쓴다.
- `internal/config/resolver.go:313-334` `loadUserTier` 는 `~/.moai/settings.json`·`~/.moai/config/sections/` 를 읽기만 한다.
- `internal/profile/sync.go:140-143` 은 프로젝트의 `sectionsDir` 에 쓴다. 실제 HOME 이 아니다.
- `EnsureHomeLayout` 이 만드는 최상위 디렉터리 13개의 권한은 다른 세션의 같은 호출도 바꿀 수 있어 비교하지 않는다. 이 함수가 실제 HOME 에 닿으면 같은 호출이 곧이어 W6 키 항목을 만들므로 W6 로 잡는다.
- Windows PowerShell 프로필은 pty 판정이 Windows 에서 건너뛰므로 뺐다.
- (v0.2.3 개정, 3회차 감사 F3) `internal/cli/migrate_agency.go:188-194` `checkpointPath` 가 `<홈>/.moai/.migrate-tx-<id>.json`(비어 있지 않은 절대 `MOAI_HOME` 이면 그 아래)을 쓴다. 호출 사슬은 `update_residue_cleanup.go:85` → `update.go:858` `runAgencyMigrationAdapter` 이고 체크포인트 쓰기는 `migrate_agency.go:251`·`:367` 에서 경로를 받는다(v0.2.3 트리 `268cffe2c` 에서 `grep -n 'checkpointPath\|runAgencyMigrationAdapter(' internal/cli/*.go` 로 확인). 1단계 패턴에 `paths.Home(` 과 이 함수가 없어 빠졌던 update 흐름의 홈 쓰기다. pty 사례가 이 경로를 실행하지 않으므로 감시하지 않고, 그 사유를 `acceptance.md` §B 비교 제외 문단에 적었다.

W2 는 이름마다 경로가 달라 `claude-profiles/*/` 한 단계 글롭으로 둔다(하위 트리는 걷지 않음). 운영자가 판정 도중 프로필을 저장하면 W2 가 달라져 FAIL 이 난다. 제품이 쓸 수 있는 파일이 실제로 바뀐 경우이므로 목록에서 빼지 않고, 실패 출력이 경로를 댄다.

t583 대조: t583 워크트리의 커밋되지 않은 `SPEC-INIT-QUIET-WIZARD-001/plan.md`(M1 `init_home_guard_test.go` 항목, 이 조사에서 읽었을 때 62번째 줄)가 같은 방식의 8항목 목록(`~/.claude/settings.json` sha256, `~/.claude/hooks/moai` 존재, 셸 설정 파일 6개의 mtime·sha256)을 쓴다. 이 목록은 그 방식을 따르되 코드에서 다시 도출했다. 차이: `~/.zprofile` 은 `selectConfigFile` 이 돌려주지 않아 넣지 않았고, fish 설정·프로필 `preferences.yaml`·W6 키 항목을 더했다. t583 문서는 커밋되지 않았으므로 그 줄 번호를 검증된 좌표로 쓰지 않는다.

## §15 자식 환경 정리 목록의 근거 (3회차 개정)

```
$ grep -n -E 'MOAI_KANBAN|MOAI_HOME|CLAUDE_CONFIG_DIR' internal/config/envkeys.go
22:	EnvHome = "MOAI_HOME"
182:	EnvMoaiKanban = "MOAI_KANBAN"
187:	EnvMoaiKanbanSpec = "MOAI_KANBAN_SPEC"
195:	EnvMoaiKanbanID = "MOAI_KANBAN_ID"
205:	EnvMoaiKanbanLabel = "MOAI_KANBAN_LABEL"
216:	EnvMoaiKanbanSettingsInjected = "MOAI_KANBAN_SETTINGS_INJECTED"
222:	EnvMoaiKanbanLeadAddr = "MOAI_KANBAN_LEAD_ADDR"
235:	EnvMoaiKanbanBackend = "MOAI_KANBAN_BACKEND"
246:	EnvMoaiKanbanCard = "MOAI_KANBAN_CARD"
263:	EnvMoaiKanbanLeadName = "MOAI_KANBAN_LEAD_NAME"
365:	EnvClaudeConfigDir = "CLAUDE_CONFIG_DIR"
$ grep -c -E '"MOAI_KANBAN' internal/config/envkeys.go
9
```

게이트 재측정 명령은 위 `grep -c` 다(`plan.md` §C 6).

빈 값을 읽는 방식: `MOAI_HOME` 은 비어 있지 않은 절대 경로일 때만 쓰인다(`internal/paths/paths.go:69`, `internal/homestate/paths.go:67-70`). `CLAUDE_CONFIG_DIR` 이 빈 값이면 원장 해석으로 넘어가고, 원장이 없으면 `default` 다(`internal/profile/profile.go:95-104`). `MOAI_KANBAN` 접두 변수를 읽는 비테스트 줄(`git grep -n -E 'Getenv\((config\.)?EnvMoaiKanban[A-Za-z]*\)' -- 'internal/*.go' ':!*_test.go'`, 20줄. 줄 수는 `grep -rn -E '<같은 패턴>' internal --include='*.go' | grep -v -c '_test\.go:'` 로 셌고 출력은 `20`)은 모두 빈 값 비교(`!= ""`, `== ""`, `== "1"`, `strings.TrimSpace`)이거나 값을 그대로 기록한다. 따라서 `-e VAR=`(빈 값)과 변수 없음은 이 코드에서 같게 읽히고, 실효 환경 관측은 `os.Getenv` 결과가 빈 문자열인지로 판정한다.

```
$ command -v tmux; tmux -V
/opt/homebrew/bin/tmux
tmux 3.6a
$ man tmux | col -b   (new-session 절 발췌)
     -e takes the form ‘VARIABLE=value’ and sets an environment
     variable for the newly created session; it may be specified
     multiple times.
$ man tmux | col -b   (GLOBAL AND SESSION ENVIRONMENT 발췌)
     When the server is started, tmux copies the environment into the global
     environment; in addition, each session has a session environment.  When a
     window is created, the session and global environments are merged.  If a
     variable exists in both, the value from the session environment is used.
     The result is the initial environment passed to the new process.
```

판독: 이미 떠 있는 tmux 서버의 전역 환경은 그 서버를 띄운 프로세스의 것이다. 테스트 프로세스가 자기 환경을 바꿔도 새 창의 자식에는 닿지 않고, 세션 환경(`-e`)으로 넘긴 값만 전역 값을 이긴다. 그래서 정리 목록은 변수마다 `-e` 로 넘긴다.

공백: tmux 자식의 실효 환경을 실행으로 재지는 않았다(man 문서 판독). 그래서 AC 는 이 판독에 기대지 않고, 자식이 스스로 기록한 실효 환경과 카나리 값으로 판정한다(`acceptance.md` §B 실효 환경 관측).

## §16 부재 단정의 대조군 측정 (3회차 개정)

```
$ git grep -l '"charm.land/huh/v2"' -- '*_test.go'
internal/cli/wizard/unified_form_test.go
c1_exit=0
$ grep -c 'charm.land/huh/v2 ' go.mod
1
c2_exit=0
$ git ls-files internal/cli/wizard/wizard.go
internal/cli/wizard/wizard.go
c3_exit=0
$ git ls-files internal/cli/no_such_file_xyz.go
(출력 없음)
c3neg_exit=0
```

마지막 명령은 없는 경로에도 `git ls-files` 가 출력 0줄·종료 0 임을 보인다. AC-ITI-011 (1) 이 대조군 경로를 같은 호출에 넣는 이유다.

## §17 S2 양성 절 기준: 호출 줄과 정의 줄 (v0.2.3 개정)

측정 트리 HEAD `268cffe2cb47107c8fcf720301df974c04e383e5`. 3회차 감사 F1 은 옛 S2 기준(주석 아닌 줄에 `persistProjectConfig` 문자열 존재)이 정의 줄 `profile_setup.go:160` 때문에 호출을 지워도 참이라는 것을 테스트 바이너리로 관측했다(`.moai/reports/t586/plan-audit-iter3.md` E-4). 새 기준은 "주석 아닌 줄 가운데 `persistProjectConfig(` 를 담고 `func persistProjectConfig(` 가 아닌 줄이 1개 이상"이다. 같은 기준을 셸 명령으로 옮겨 현재 파일(양성)과 호출만 지운 사본(음성)에 돌렸다. 사본은 저장소 밖 스크래치 디렉터리에 두었고 제품 트리는 건드리지 않았다. 정규식에는 단어 경계 표기를 쓰지 않았다(POSIX ERE 에서 경계가 아니다, §6.1).

```
$ grep -n -F 'persistProjectConfig(' internal/cli/profile_setup.go
160:func persistProjectConfig(projectRoot, devMode, convention string) error {
520:			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
exit=0
$ diff internal/cli/profile_setup.go <scratch>/profile_setup.go
520c520
< 			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
---
> 			if err := error(nil); err != nil {
diff_exit=1
== 양성 (현재 파일)
$ grep -n -F 'persistProjectConfig(' internal/cli/profile_setup.go | grep -v -E '^[0-9]+:[[:space:]]*//' | grep -v -F 'func persistProjectConfig('
520:			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
exit=0
== 음성 대조 (호출만 지우고 정의는 남긴 사본)
$ grep -n -F 'persistProjectConfig(' <scratch>/profile_setup.go | grep -v -E '^[0-9]+:[[:space:]]*//' | grep -v -F 'func persistProjectConfig('
(출력 없음)
exit=1
== 대조: 사본에 정의 줄은 남아 있다
$ grep -n -F 'func persistProjectConfig(' <scratch>/profile_setup.go
160:func persistProjectConfig(projectRoot, devMode, convention string) error {
exit=0
```

판독: 걸러내기 전 첫 단계가 두 줄(정의 `:160`, 호출 `:520`)을 읽으므로 빈 결과는 파일을 못 읽어서 생긴 것이 아니다. 새 기준은 현재 파일에서 호출 줄 `:520` 하나만 남기고, 정의 줄만 남은 사본에서는 비어(종료 1) 거짓이 된다. 사본에 정의 줄이 있다는 대조가 이 음성 결과를 옛 기준으로는 참이던 바로 그 상태로 묶는다. 주석 줄 걸러내기는 `// if err := persistProjectConfig(…)` 처럼 호출을 주석으로 막은 뮤턴트도 거짓으로 만든다(현재 파일의 주석 언급 `:152`·`:227`·`:243`·`:518` 은 괄호가 붙지 않아 첫 단계에서 이미 빠진다).

대상 파일: M4 는 `init.go`·`update.go` 의 확인창과 `runProfileSetup` 호출만 지우고(`plan.md` M4), M5 는 폼을 wizard 패키지로 옮기되 저장 이하는 `runProfileSetup` 에 그대로 둔다(`design.md` §2.2 "저장 이하 ← 그대로"). 따라서 흡수 뒤에도 호출 줄은 `profile_setup.go` 에 있고, S2 양성 절의 대상 파일은 바뀌지 않는다.

공백: 이 측정은 셸 명령이다. run 단계에서 가드 테스트 본문을 이 기준으로 바꾼 뒤 호출만 지운 사본에서 `TestTUINestedConfigNoParallelWriter` 가 `--- FAIL` 하는지는 AC-ITI-010 (4) 가 판정한다. 이 개정에서는 테스트를 실행하지 않았다.
