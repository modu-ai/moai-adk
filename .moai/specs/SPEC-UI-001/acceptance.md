---
id: SPEC-UI-001
title: Terminal UI Framework - Acceptance Criteria
version: 0.1.0
status: draft
created: 2026-02-03
updated: 2026-02-03
spec_ref: SPEC-UI-001/spec.md
---

# SPEC-UI-001: Terminal UI Framework -- Acceptance Criteria

## 1. Wizard 전체 흐름 (REQ-5)

### AC-1.1: Wizard 정상 완료

```gherkin
Given wizard가 interactive 모드에서 실행되었을 때
When 사용자가 모든 6단계를 순서대로 완료하면
  - Step 1: 프로젝트 이름에 "my-project"를 입력하고 Enter
  - Step 2: 언어 목록에서 "Go"를 선택하고 Enter
  - Step 3: 프레임워크 목록에서 "Cobra CLI"를 선택하고 Enter
  - Step 4: 기능 목록에서 "LSP", "Quality Gates"를 Space로 선택 후 Enter
  - Step 5: 사용자 이름에 "GOOS"를 입력하고 Enter
  - Step 6: 대화 언어에서 "Korean"을 선택하고 Enter
Then WizardResult가 반환된다
  And WizardResult.ProjectName == "my-project"
  And WizardResult.Language == "Go"
  And WizardResult.Framework == "Cobra CLI"
  And WizardResult.Features == ["LSP", "Quality Gates"]
  And WizardResult.UserName == "GOOS"
  And WizardResult.ConvLang == "ko"
  And error == nil
```

### AC-1.2: Wizard 이전 단계 네비게이션

```gherkin
Given wizard가 Step 3 (프레임워크 선택)에 있을 때
When 사용자가 Back 동작을 수행하면
Then Step 2 (언어 선택)로 돌아간다
  And Step 2에서 이전에 선택했던 값이 유지된다
  And Step 2를 다시 변경할 수 있다
```

### AC-1.3: Wizard Headless 모드

```gherkin
Given TTY가 연결되지 않은 환경(CI/CD)에서
  And SetDefaults()로 다음 값이 설정되었을 때:
    | key           | value      |
    | project_name  | ci-project |
    | language      | Python     |
    | framework     | FastAPI    |
    | features      | LSP        |
    | user_name     | CI-Bot     |
    | conv_lang     | en         |
When Wizard.Run()이 호출되면
Then bubbletea Program을 생성하지 않는다
  And 기본값으로 즉시 WizardResult를 반환한다
  And 실행 시간이 10ms 이내이다
```

### AC-1.4: Wizard Context 취소

```gherkin
Given wizard가 실행 중일 때
When context가 cancel 신호를 받으면
Then 현재 bubbletea Program이 즉시 종료된다
  And 터미널 상태가 원래대로 복원된다
  And context.Canceled 에러가 반환된다
  And WizardResult는 nil이다
```

---

## 2. Selector 네비게이션 (REQ-2)

### AC-2.1: Selector 기본 선택

```gherkin
Given 다음 항목이 포함된 Selector가 표시되었을 때:
  | Label   | Value   | Desc              |
  | Go      | go      | Compiled language |
  | Python  | python  | Scripting language|
  | TypeScript | ts   | Typed JavaScript  |
When 사용자가 Down 키를 한 번 누르고 Enter를 누르면
Then 선택 결과로 "python"이 반환된다
  And error == nil
```

### AC-2.2: Selector Fuzzy Search

```gherkin
Given 10개 이상의 항목이 포함된 Selector가 표시되었을 때
When 사용자가 "typ"을 입력하면
Then 목록이 필터링되어 "TypeScript"를 포함하는 항목만 표시된다
  And 필터링 응답 시간이 50ms 이내이다
```

### AC-2.3: Selector ESC 취소

```gherkin
Given Selector가 표시되었을 때
When 사용자가 ESC 키를 누르면
Then 빈 문자열("")이 반환된다
  And error가 nil이 아닌 취소 에러를 포함한다
```

### AC-2.4: Selector Headless 모드

```gherkin
Given headless 모드에서 실행되고
  And SetDefaults에 "language" = "go"가 설정되었을 때
When Selector.Select("language", items)가 호출되면
Then 즉시 "go"를 반환한다
  And 터미널에 아무것도 렌더링하지 않는다
```

---

## 3. Checkbox 토글 (REQ-3)

### AC-3.1: Checkbox 다중 선택

```gherkin
Given 다음 항목이 포함된 Checkbox가 표시되었을 때:
  | Label          | Value         |
  | LSP            | lsp           |
  | Quality Gates  | quality       |
  | Git Hooks      | hooks         |
  | Statusline     | statusline    |
When 사용자가 "LSP"에서 Space를 누르고
  And Down을 두 번 누르고 "Git Hooks"에서 Space를 누르고
  And Enter를 누르면
Then 반환 결과가 ["lsp", "hooks"]이다
  And error == nil
```

### AC-3.2: Checkbox 전체 선택 토글

```gherkin
Given 4개 항목이 포함된 Checkbox가 표시되었을 때
  And 아무것도 선택되지 않은 상태에서
When 사용자가 'a' 키를 누르면
Then 모든 항목이 선택 상태([x])로 변경된다
When 사용자가 다시 'a' 키를 누르면
Then 모든 항목이 해제 상태([ ])로 변경된다
```

### AC-3.3: Checkbox 빈 선택

```gherkin
Given Checkbox가 표시되었을 때
  And 아무것도 선택하지 않은 상태에서
When 사용자가 Enter를 누르면
Then 빈 슬라이스 []string{}가 반환된다
  And error == nil
```

---

## 4. Progress 업데이트 (REQ-6)

### AC-4.1: Determinate Progress Bar

```gherkin
Given Progress.Start("Deploying templates", 10)이 호출되었을 때
When ProgressBar.Increment(1)이 3번 호출되면
Then 프로그레스 바가 30% (3/10) 상태로 렌더링된다
  And 타이틀에 "Deploying templates"가 표시된다
When ProgressBar.Done()이 호출되면
Then 프로그레스 바가 100% 완료 상태로 렌더링된다
```

### AC-4.2: Indeterminate Spinner

```gherkin
Given Progress.Spinner("Checking for updates...")가 호출되었을 때
Then 스피너 애니메이션이 시작된다
  And 타이틀에 "Checking for updates..."가 표시된다
When Spinner.SetTitle("Downloading...")가 호출되면
Then 타이틀이 "Downloading..."으로 변경된다
When Spinner.Stop()이 호출되면
Then 스피너 애니메이션이 중지된다
```

### AC-4.3: Progress Headless 모드

```gherkin
Given headless 모드에서 실행될 때
When Progress.Start("Processing", 5)가 호출되고
  And Increment(1)이 3번 호출되면
Then 터미널에 다음과 같은 로그가 출력된다:
  "[1/5] Processing"
  "[2/5] Processing"
  "[3/5] Processing"
  And bubbletea 프로그레스 바가 아닌 plain text 로그이다
```

---

## 5. Theme 전환 (REQ-1)

### AC-5.1: Dark Mode 테마 적용

```gherkin
Given config에서 theme.mode = "dark"으로 설정되었을 때
When Theme이 로드되면
Then Primary 색상이 밝은 계열(예: #7C3AED)이다
  And Background 색상이 어두운 계열이다
  And 모든 lipgloss 스타일이 dark palette를 사용한다
```

### AC-5.2: Light Mode 테마 적용

```gherkin
Given config에서 theme.mode = "light"으로 설정되었을 때
When Theme이 로드되면
Then Primary 색상이 어두운 계열(예: #5B21B6)이다
  And Background 색상이 밝은 계열이다
  And 모든 lipgloss 스타일이 light palette를 사용한다
```

### AC-5.3: NO_COLOR 모드

```gherkin
Given MOAI_NO_COLOR 환경 변수가 "true"로 설정되었을 때
When 어떤 UI 컴포넌트가 렌더링되면
Then 출력에 ANSI escape code가 포함되지 않는다
  And 텍스트만 plain text로 출력된다
  And 프로그레스 바는 ASCII 문자 [=====>     ]로 렌더링된다
```

### AC-5.4: 자동 테마 감지

```gherkin
Given config에서 theme.mode = "auto"로 설정되었을 때
When Theme이 로드되면
Then lipgloss.HasDarkBackground()를 호출하여 터미널 배경색을 감지한다
  And 결과에 따라 dark 또는 light 팔레트를 자동 적용한다
```

---

## 6. Non-Interactive 폴백 (Edge Cases)

### AC-6.1: TTY 감지

```gherkin
Given os.Stdin이 파이프(pipe)로 연결되었을 때
When NonInteractive.IsHeadless()가 호출되면
Then true를 반환한다
  And 모든 UI 컴포넌트가 headless 모드로 동작한다
```

### AC-6.2: --non-interactive 플래그

```gherkin
Given CLI에서 --non-interactive 플래그가 전달되었을 때
  And TTY가 연결되어 있더라도
When NonInteractive.IsHeadless()가 호출되면
Then true를 반환한다 (플래그가 TTY 감지를 오버라이드)
```

### AC-6.3: 기본값 없는 Headless 모드

```gherkin
Given headless 모드에서 실행되고
  And SetDefaults()가 호출되지 않았을 때
When Wizard.Run()이 호출되면
Then 에러를 반환한다: "headless mode requires defaults for all wizard fields"
  And WizardResult는 nil이다
```

---

## 7. Edge Cases

### AC-7.1: 터미널 리사이즈 대응

```gherkin
Given Selector가 80x24 터미널에서 표시 중일 때
When 터미널 크기가 120x40으로 변경되면
Then bubbletea WindowSizeMsg를 수신한다
  And 목록 렌더링이 새 크기에 맞게 재조정된다
  And 현재 선택 상태가 유지된다
```

### AC-7.2: 키보드 인터럽트 (Ctrl+C)

```gherkin
Given 어떤 UI 컴포넌트가 실행 중일 때
When 사용자가 Ctrl+C를 누르면
Then bubbletea Program이 즉시 종료된다
  And 터미널이 원래 상태로 복원된다 (raw mode 해제, 커서 표시)
  And 에러가 반환된다 (interrupt 또는 cancel)
```

### AC-7.3: 빈 항목 목록

```gherkin
Given Selector.Select("language", []SelectItem{})가 호출될 때 (빈 목록)
Then 에러를 반환한다: "no items to select from"
  And 빈 문자열("")이 반환된다
```

### AC-7.4: 매우 긴 항목 목록

```gherkin
Given 1000개 항목이 포함된 Selector가 표시되었을 때
When 사용자가 fuzzy search로 "typ"을 입력하면
Then 필터링이 100ms 이내에 완료된다
  And 스크롤 가능한 목록으로 결과가 표시된다
```

### AC-7.5: 특수 문자 입력

```gherkin
Given 텍스트 프롬프트가 표시되었을 때
When 사용자가 유니코드 문자(한글, 일본어, 이모지)를 입력하면
Then 입력이 올바르게 표시된다
  And WizardResult에 유니코드 문자가 정확히 저장된다
```

### AC-7.6: Validation 실패 반복

```gherkin
Given 프로젝트 이름 프롬프트에 validation (비어있지 않음, 특수문자 금지)이 설정되었을 때
When 사용자가 빈 문자열을 입력하고 Enter를 누르면
Then 인라인 에러 메시지가 표시된다: "Project name cannot be empty"
  And 프롬프트가 유지되어 재입력이 가능하다
When 사용자가 "my@project"를 입력하고 Enter를 누르면
Then 인라인 에러 메시지가 표시된다: "Project name contains invalid characters"
When 사용자가 "my-project"를 입력하고 Enter를 누르면
Then validation이 통과하고 다음 단계로 진행한다
```

---

## 8. Quality Gate

### 8.1 Definition of Done

- [ ] 모든 6개 파일이 `internal/ui/` 디렉터리에 존재한다
- [ ] 모든 공개 인터페이스(Wizard, Selector, Checkbox, Progress, NonInteractive)가 구현되었다
- [ ] 테스트 커버리지 >= 70% (ui 패키지 전체)
- [ ] Headless 경로 테스트 커버리지 = 100%
- [ ] bubbletea 스냅샷 테스트가 핵심 시나리오를 커버한다
- [ ] `golangci-lint run ./internal/ui/...` 에러 0건
- [ ] `go vet ./internal/ui/...` 에러 0건
- [ ] godoc 주석이 모든 공개 타입과 함수에 작성되었다
- [ ] dark mode, light mode, NO_COLOR 3가지 모드에서 정상 동작 확인
- [ ] macOS, Linux, Windows 최소 1개 터미널에서 수동 검증 완료
- [ ] `moai init` 명령에서 wizard가 end-to-end로 동작한다
- [ ] CI/CD 환경(headless)에서 `moai init --non-interactive`가 정상 동작한다

### 8.2 검증 방법

| 검증 항목 | 방법 | 도구 |
|----------|------|------|
| 인터페이스 준수 | 컴파일 타임 검증 | go build |
| 상태 변경 정확성 | 유닛 테스트 | go test |
| 렌더링 정확성 | 스냅샷 테스트 | bubbletea/teatest |
| Headless 모드 | 자동화 테스트 | go test (TTY mock) |
| 테마 적용 | 시각적 수동 검증 + 스냅샷 | manual + golden files |
| 크로스 플랫폼 | CI 매트릭스 | GitHub Actions (darwin, linux, windows) |
| 성능 | 벤치마크 테스트 | go test -bench |
