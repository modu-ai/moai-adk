---
id: SPEC-UI-001
title: Terminal UI Framework
version: 0.1.0
status: Completed
created: 2026-02-03
updated: 2026-02-03
author: MoAI
priority: medium
phase: "Phase 4 - UI and Integration"
module: "internal/ui/"
dependencies:
  - SPEC-CONFIG-001
adr_references:
  - ADR-010
resolves:
  - "#268"
  - "#249"
  - "#286"
tags: bubbletea, lipgloss, tui, elm-architecture, charmbracelet
---

# SPEC-UI-001: Terminal UI Framework

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|----------|
| 0.1.0 | 2026-02-03 | MoAI | 초안 작성 |

---

## 1. Environment (환경)

### 1.1 프로젝트 컨텍스트

MoAI-ADK (Go Edition)는 Python 기반 MoAI-ADK(~73,000 LOC)의 완전한 Go 재작성 프로젝트이다. 기존 Python 구현에서 Rich (패널, 테이블, 프로그레스 바)와 InquirerPy (인터랙티브 프롬프트)가 담당하던 터미널 UI를 Charmbracelet 생태계(bubbletea + lipgloss + bubbles + huh)로 대체한다.

### 1.2 기술 환경

| 항목 | 값 |
|------|-----|
| 언어 | Go 1.22+ |
| TUI Framework | github.com/charmbracelet/bubbletea v1.2+ |
| Styling | github.com/charmbracelet/lipgloss v1.0+ |
| Components | github.com/charmbracelet/bubbles v0.20+ |
| Form Framework | github.com/charmbracelet/huh v0.6+ |
| 모듈 경로 | github.com/modu-ai/moai-adk-go |
| 대상 디렉터리 | internal/ui/ |
| 예상 LOC | ~1,200 |

### 1.3 플랫폼 지원

- macOS (arm64, amd64)
- Linux (arm64, amd64)
- Windows (amd64, arm64)
- CGO_ENABLED=0 (Pure Go, C 의존성 없음)

### 1.4 의존 모듈

| 모듈 | 의존 방향 | 용도 |
|------|----------|------|
| internal/config/ | UI -> Config | 테마 설정 읽기 (SPEC-CONFIG-001) |
| pkg/models/ | UI -> Models | SelectItem, WizardResult 등 공유 모델 |
| internal/cli/ | CLI -> UI | Cobra 명령에서 UI 컴포넌트 호출 |
| internal/update/ | Update -> UI | 업데이트 진행률 표시 |
| internal/core/project/ | Project -> UI | init wizard 호출 |

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

- **A1**: 대상 터미널은 ANSI escape code를 지원한다 (대부분의 현대 터미널 에뮬레이터).
- **A2**: bubbletea의 Elm Architecture(Model-Update-View)가 Go의 명시적 상태 관리와 자연스럽게 호환된다.
- **A3**: lipgloss의 스타일링이 macOS, Linux, Windows 터미널에서 일관되게 렌더링된다.
- **A4**: SPEC-CONFIG-001이 테마 설정(color scheme, dark/light mode)을 제공한다.
- **A5**: `huh` 라이브러리가 multi-step form wizard에 적합하다.

### 2.2 사용자 가정

- **A6**: 사용자는 키보드를 통해 TUI와 상호작용한다 (마우스 입력은 선택적 지원).
- **A7**: CI/CD 환경에서는 TTY가 없으므로 headless/non-interactive 모드가 필요하다.
- **A8**: 최소 80 컬럼 x 24 행 이상의 터미널 크기를 전제한다.

### 2.3 리스크 가정

- **A9**: bubbletea 학습 곡선이 존재하나, Elm Architecture 패턴이 잘 문서화되어 있어 관리 가능하다.
- **A10**: Windows Terminal에서의 유니코드 렌더링 차이가 있을 수 있으나, lipgloss가 플랫폼 차이를 추상화한다.

---

## 3. Requirements (요구사항)

### REQ-1: Theme System (테마 시스템)

> **theme.go** -- 모든 컴포넌트의 기반

**REQ-1.1** [Ubiquitous]
시스템은 **항상** lipgloss 기반의 일관된 색상 테마를 모든 UI 컴포넌트에 적용해야 한다.

**REQ-1.2** [State-Driven]
**IF** 사용자가 config에서 dark mode를 설정한 경우 **THEN** 어두운 배경에 최적화된 색상 팔레트를 적용해야 한다.

**REQ-1.3** [State-Driven]
**IF** 사용자가 config에서 light mode를 설정한 경우 **THEN** 밝은 배경에 최적화된 색상 팔레트를 적용해야 한다.

**REQ-1.4** [Event-Driven]
**WHEN** `MOAI_NO_COLOR` 환경 변수가 설정되었을 때 **THEN** 모든 색상과 스타일링을 비활성화하고 plain text로 출력해야 한다.

**REQ-1.5** [Unwanted]
시스템은 ANSI escape code를 직접 하드코딩**하지 않아야 한다**. 모든 스타일링은 lipgloss API를 통해 적용해야 한다.

### REQ-2: Selector Component (단일 선택)

> **selector.go** -- fuzzy search 지원 단일 선택

**REQ-2.1** [Event-Driven]
**WHEN** Selector.Select()가 호출될 때 **THEN** 레이블과 선택 목록을 터미널에 렌더링하고 사용자의 선택을 반환해야 한다.

**REQ-2.2** [Event-Driven]
**WHEN** 사용자가 문자를 입력할 때 **THEN** fuzzy search로 목록을 필터링하여 일치하는 항목만 표시해야 한다.

**REQ-2.3** [Event-Driven]
**WHEN** 사용자가 위/아래 화살표 키를 누를 때 **THEN** 선택 커서를 해당 방향으로 이동해야 한다.

**REQ-2.4** [Event-Driven]
**WHEN** 사용자가 Enter 키를 누를 때 **THEN** 현재 커서 위치의 항목을 선택 결과로 반환해야 한다.

**REQ-2.5** [Event-Driven]
**WHEN** 사용자가 ESC 키를 누를 때 **THEN** 선택을 취소하고 빈 문자열과 에러를 반환해야 한다.

**REQ-2.6** [State-Driven]
**IF** headless 모드(TTY 없음)인 경우 **THEN** 기본값 또는 SetDefaults()로 설정된 값을 즉시 반환해야 한다.

### REQ-3: Checkbox Component (다중 선택)

> **checkbox.go** -- 토글 기반 다중 선택

**REQ-3.1** [Event-Driven]
**WHEN** Checkbox.MultiSelect()가 호출될 때 **THEN** 체크박스 목록을 렌더링하고 사용자의 다중 선택 결과를 반환해야 한다.

**REQ-3.2** [Event-Driven]
**WHEN** 사용자가 Space 키를 누를 때 **THEN** 현재 항목의 선택 상태를 토글해야 한다.

**REQ-3.3** [Event-Driven]
**WHEN** 사용자가 Enter 키를 누를 때 **THEN** 선택된 모든 항목의 값을 문자열 슬라이스로 반환해야 한다.

**REQ-3.4** [Optional]
**가능하면** 사용자가 'a' 키로 전체 선택/해제를 토글할 수 있는 기능을 제공한다.

**REQ-3.5** [State-Driven]
**IF** headless 모드인 경우 **THEN** SetDefaults()로 설정된 선택 목록을 즉시 반환해야 한다.

### REQ-4: Prompt Component (텍스트 입력)

> **prompt.go** -- 텍스트 입력 및 확인 프롬프트

**REQ-4.1** [Event-Driven]
**WHEN** 텍스트 프롬프트가 실행될 때 **THEN** 레이블과 입력 필드를 렌더링하고 사용자 입력을 반환해야 한다.

**REQ-4.2** [Event-Driven]
**WHEN** validation 함수가 설정되고 사용자가 Enter를 누를 때 **THEN** 입력값에 대해 validation을 실행하고, 실패 시 에러 메시지를 인라인으로 표시해야 한다.

**REQ-4.3** [Event-Driven]
**WHEN** confirm 프롬프트가 실행될 때 **THEN** Yes/No 선택을 표시하고 boolean 결과를 반환해야 한다.

**REQ-4.4** [State-Driven]
**IF** placeholder 텍스트가 설정된 경우 **THEN** 입력 필드가 비어있을 때 placeholder를 회색으로 표시해야 한다.

### REQ-5: Wizard Component (다단계 폼)

> **wizard.go** -- `moai init` 대화형 프로젝트 초기화

**REQ-5.1** [Event-Driven]
**WHEN** Wizard.Run()이 호출될 때 **THEN** 다단계 폼을 순차적으로 실행하고 WizardResult를 반환해야 한다.

**REQ-5.2** [Ubiquitous]
시스템은 **항상** 다음 단계를 포함해야 한다:
1. 프로젝트 이름 입력 (텍스트 프롬프트)
2. 프로그래밍 언어 선택 (Selector)
3. 프레임워크 선택 (Selector, 언어에 따라 동적 필터링)
4. 기능 선택 (Checkbox)
5. 사용자 이름 입력 (텍스트 프롬프트)
6. 대화 언어 선택 (Selector)

**REQ-5.3** [Event-Driven]
**WHEN** 사용자가 이전 단계로 돌아가려 할 때 **THEN** 이전 단계의 상태를 유지한 채 해당 단계로 네비게이션해야 한다.

**REQ-5.4** [Event-Driven]
**WHEN** context가 취소(cancel)될 때 **THEN** 현재 진행 상태를 정리하고 context.Canceled 에러를 반환해야 한다.

**REQ-5.5** [State-Driven]
**IF** headless 모드인 경우 **THEN** SetDefaults()로 제공된 기본값을 사용하여 WizardResult를 즉시 반환해야 한다.

**REQ-5.6** [Unwanted]
시스템은 wizard 실행 중 터미널 상태를 오염시키**지 않아야 한다**. 종료 시(정상/비정상 모두) 터미널을 원래 상태로 복원해야 한다.

### REQ-6: Progress Component (진행률)

> **progress.go** -- 프로그레스 바와 스피너

**REQ-6.1** [Event-Driven]
**WHEN** Progress.Start()가 호출될 때 **THEN** 지정된 total 값을 가진 determinate 프로그레스 바를 생성하고 ProgressBar 인터페이스를 반환해야 한다.

**REQ-6.2** [Event-Driven]
**WHEN** ProgressBar.Increment()가 호출될 때 **THEN** 프로그레스 바의 현재 진행률을 n만큼 증가시키고 터미널을 업데이트해야 한다.

**REQ-6.3** [Event-Driven]
**WHEN** ProgressBar.Done()이 호출될 때 **THEN** 프로그레스 바를 100%로 완료하고 완료 상태로 렌더링해야 한다.

**REQ-6.4** [Event-Driven]
**WHEN** Progress.Spinner()가 호출될 때 **THEN** indeterminate 스피너 애니메이션을 시작하고 Spinner 인터페이스를 반환해야 한다.

**REQ-6.5** [State-Driven]
**IF** headless 모드인 경우 **THEN** 프로그레스 바 대신 단순 로그 출력(예: "[3/10] Processing...")을 사용해야 한다.

**REQ-6.6** [State-Driven]
**IF** `MOAI_NO_COLOR`가 설정된 경우 **THEN** 프로그레스 바를 ASCII 문자(`[=====>     ]`)로 렌더링해야 한다.

---

## 4. Specifications (사양)

### 4.1 인터페이스 설계

```go
package ui

import "context"

// Wizard는 대화형 프로젝트 초기화 흐름을 실행한다.
type Wizard interface {
    Run(ctx context.Context) (*WizardResult, error)
}

// WizardResult는 init wizard의 사용자 선택을 보유한다.
type WizardResult struct {
    ProjectName string
    Language    string
    Framework   string
    Features    []string
    UserName    string
    ConvLang    string
}

// Selector는 fuzzy search를 지원하는 단일 선택을 제공한다.
type Selector interface {
    Select(label string, items []SelectItem) (string, error)
}

// Checkbox는 검색을 지원하는 다중 선택을 제공한다.
type Checkbox interface {
    MultiSelect(label string, items []SelectItem) ([]string, error)
}

// SelectItem은 선택 가능한 옵션을 나타낸다.
type SelectItem struct {
    Label string
    Value string
    Desc  string
}

// Progress는 프로그레스 바와 스피너를 제공한다.
type Progress interface {
    Start(title string, total int) ProgressBar
    Spinner(title string) Spinner
}

// ProgressBar는 determinate 진행률 표시기이다.
type ProgressBar interface {
    Increment(n int)
    SetTitle(title string)
    Done()
}

// Spinner는 indeterminate 진행률 표시기이다.
type Spinner interface {
    SetTitle(title string)
    Stop()
}

// NonInteractive는 headless 모드 지원을 제공한다.
type NonInteractive interface {
    SetDefaults(defaults map[string]string)
    IsHeadless() bool
}
```

### 4.2 파일 구조

| 파일 | 역할 | 예상 LOC | Python 대응 |
|------|------|----------|------------|
| theme.go | MoAI 색상 테마 (lipgloss) | ~150 | cli/ui/theme.py |
| selector.go | Fuzzy 단일 선택 | ~200 | InquirerPy fuzzy_select() |
| checkbox.go | 다중 선택 + 검색 | ~200 | InquirerPy fuzzy_checkbox() |
| prompt.go | 텍스트 입력 / 확인 프롬프트 | ~150 | cli/ui/prompts.py |
| wizard.go | Init wizard (bubbletea Elm model) | ~300 | cli/prompts/init_prompts.py |
| progress.go | 프로그레스 바 + 스피너 (bubbles) | ~200 | Rich Progress() |

### 4.3 의존성 그래프

```
internal/ui/
    theme.go      <-- 모든 컴포넌트가 의존
    selector.go   <-- theme.go
    checkbox.go   <-- theme.go
    prompt.go     <-- theme.go
    wizard.go     <-- selector.go, checkbox.go, prompt.go, theme.go
    progress.go   <-- theme.go
```

### 4.4 Elm Architecture 패턴 (bubbletea)

각 컴포넌트는 bubbletea의 Elm Architecture를 따른다:

```
Model (상태) --> Update (메시지 처리) --> View (렌더링)
     ^                                        |
     +----------------------------------------+
```

- **Model**: 컴포넌트의 현재 상태 (선택 인덱스, 필터 텍스트, 진행률 등)
- **Update**: KeyMsg, WindowSizeMsg 등의 메시지를 처리하여 Model을 갱신
- **View**: 현재 Model을 기반으로 lipgloss 스타일이 적용된 문자열 반환

### 4.5 성능 요구사항

| 지표 | 목표 |
|------|------|
| 컴포넌트 초기화 | < 10ms |
| 키 입력 응답 | < 16ms (60fps) |
| fuzzy search 필터링 | < 50ms (100개 항목 기준) |
| 테마 로드 | < 5ms |

### 4.6 호환성 매트릭스

| 터미널 | macOS | Linux | Windows |
|--------|-------|-------|---------|
| iTerm2 | Full | -- | -- |
| Terminal.app | Full | -- | -- |
| GNOME Terminal | -- | Full | -- |
| Alacritty | Full | Full | Full |
| Windows Terminal | -- | -- | Full |
| cmd.exe | -- | -- | Degraded (NO_COLOR) |
| CI/CD (no TTY) | Headless | Headless | Headless |

---

## 5. Traceability (추적성)

| 요구사항 | 파일 | 관련 이슈 | ADR |
|---------|------|----------|-----|
| REQ-1.x (Theme) | theme.go | -- | ADR-010 |
| REQ-2.x (Selector) | selector.go | #268 | ADR-010 |
| REQ-3.x (Checkbox) | checkbox.go | #268 | ADR-010 |
| REQ-4.x (Prompt) | prompt.go | -- | ADR-010 |
| REQ-5.x (Wizard) | wizard.go | #249, #286 | ADR-010 |
| REQ-6.x (Progress) | progress.go | #268 | ADR-010 |
| NonInteractive | 전체 | CI/CD 지원 | -- |
| SPEC-CONFIG-001 | theme.go | 테마 설정 의존 | -- |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 96.8%

### Summary

Terminal UI framework implemented using Charmbracelet bubbletea with Elm Architecture (Model-Update-View). Includes themed selector, checkbox group, text prompt, multi-step wizard, and progress bar components. Full headless mode support for CI/CD environments with non-interactive defaults. Theme system integrates with SPEC-CONFIG-001 for customizable color palettes and styling.

### Files Created

- `internal/ui/checkbox.go`
- `internal/ui/checkbox_test.go`
- `internal/ui/edge_case_test.go`
- `internal/ui/headless.go`
- `internal/ui/headless_test.go`
- `internal/ui/model_update_test.go`
- `internal/ui/progress.go`
- `internal/ui/progress_test.go`
- `internal/ui/prompt.go`
- `internal/ui/prompt_test.go`
- `internal/ui/runner.go`
- `internal/ui/runner_test.go`
- `internal/ui/selector.go`
- `internal/ui/selector_test.go`
- `internal/ui/theme.go`
- `internal/ui/theme_test.go`
- `internal/ui/ui.go`
- `internal/ui/ui_test.go`
- `internal/ui/wizard.go`
- `internal/ui/wizard_test.go`
