# SPEC-INIT-TUX-I18N-001 — 조사

## §0 출처와 재확인 범위

- 측정 트리: 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, HEAD `18144b7aca714ea8924363b1eab4640cf101c6d0`(착수 시 `git rev-parse HEAD` 로 확인).
- 이전 조사 트리 `e7a7d4bb3` 와의 관계: `git diff --stat e7a7d4bb3 HEAD -- internal cmd pkg go.mod go.sum` 출력 없음(종료 0). 사이의 커밋 두 개(`d0ec7921f`, `18144b7ac`)는 SPEC 문서와 감사 기록뿐이다. 따라서 이전 조사 좌표는 이 트리에서도 유효하다.
- 1회차 감사(`.moai/reports/t586/plan-audit.md`)가 인용한 좌표 가운데 이 SPEC 이 쓰는 것은 이 트리에서 다시 쟀다. 다시 재지 않은 것은 §13 에 적는다.
- 명령은 모두 zsh 에서 실행했다. 글롭은 `git grep` 경로 인자로만 넘겨 셸 글롭 확장 실패(`no matches found`)를 피했다.

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
