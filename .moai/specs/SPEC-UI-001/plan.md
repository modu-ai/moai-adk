---
id: SPEC-UI-001
title: Terminal UI Framework - Implementation Plan
version: 0.1.0
status: draft
created: 2026-02-03
updated: 2026-02-03
spec_ref: SPEC-UI-001/spec.md
---

# SPEC-UI-001: Terminal UI Framework -- Implementation Plan

## 1. 구현 전략

### 1.1 핵심 원칙

- **Foundation First**: 테마 시스템을 가장 먼저 구현하여 모든 컴포넌트의 시각적 일관성 기반을 확보한다.
- **Bottom-Up Composition**: 기본 컴포넌트(selector, checkbox, prompt)를 먼저 만들고, 이를 조합하여 복합 컴포넌트(wizard)를 구성한다.
- **Headless Parallel Track**: 모든 컴포넌트에 headless 모드를 동시에 구현하여 CI/CD 호환성을 보장한다.
- **Elm Architecture Consistency**: 모든 인터랙티브 컴포넌트는 bubbletea의 Model-Update-View 패턴을 따른다.

### 1.2 기술 접근

| 컴포넌트 | 기반 라이브러리 | 접근 방식 |
|---------|--------------|----------|
| Theme | lipgloss | 색상 팔레트 정의 + config 연동 |
| Selector | bubbles/list 또는 huh | fuzzy 필터링 + 키보드 네비게이션 |
| Checkbox | bubbles/list 또는 huh | 토글 상태 관리 + 다중 선택 |
| Prompt | bubbles/textinput 또는 huh | 입력 validation + placeholder |
| Wizard | huh (multi-step form) | 단계별 폼 조합 + 상태 유지 |
| Progress | bubbles/progress + bubbles/spinner | determinate bar + indeterminate spinner |

---

## 2. Milestone 계획

### Milestone 1: Theme Foundation (Primary Goal)

**우선순위**: High
**예상 LOC**: ~150

**작업 내용**:
- [ ] MoAI 기본 색상 팔레트 정의 (Primary, Secondary, Success, Warning, Error, Muted)
- [ ] Dark mode / Light mode 팔레트 분리
- [ ] lipgloss.Style 래퍼 함수 구현 (Title, Subtitle, Body, Error, Highlight, Muted 등)
- [ ] `MOAI_NO_COLOR` 환경 변수 감지 및 plain text 모드
- [ ] SPEC-CONFIG-001 연동: config에서 테마 설정 읽기
- [ ] lipgloss.HasDarkBackground() 기반 자동 테마 감지
- [ ] 테마 유닛 테스트

**의존성**: SPEC-CONFIG-001 (config에서 테마 설정 읽기)
**완료 기준**: 모든 스타일 함수가 dark/light/no-color 3가지 모드에서 올바르게 동작

### Milestone 2: Basic Components (Primary Goal)

**우선순위**: High
**예상 LOC**: ~550

**작업 내용**:

**2a. Selector (단일 선택)**
- [ ] bubbletea Model 구현 (items, cursor, filter 상태)
- [ ] Update: 키보드 입력 처리 (Up/Down, Enter, ESC, 문자 입력)
- [ ] View: lipgloss 스타일 적용 렌더링
- [ ] Fuzzy search 필터링 로직
- [ ] Headless 모드: SetDefaults 기반 즉시 반환
- [ ] 스냅샷 테스트 + 유닛 테스트

**2b. Checkbox (다중 선택)**
- [ ] bubbletea Model 구현 (items, cursor, selected map)
- [ ] Update: Space(토글), Enter(확인), ESC(취소)
- [ ] View: 체크박스 상태 렌더링 ([x] / [ ])
- [ ] 전체 선택/해제 토글 (선택적)
- [ ] Headless 모드 지원
- [ ] 스냅샷 테스트 + 유닛 테스트

**2c. Prompt (텍스트 입력)**
- [ ] 텍스트 입력 프롬프트 (bubbles/textinput 기반)
- [ ] Validation 함수 연동 + 인라인 에러 표시
- [ ] Confirm 프롬프트 (Yes/No)
- [ ] Placeholder 텍스트 지원
- [ ] Headless 모드 지원
- [ ] 유닛 테스트

**의존성**: Milestone 1 (Theme)
**완료 기준**: 각 컴포넌트가 독립적으로 동작하며, 테마가 올바르게 적용됨

### Milestone 3: Composite Components (Secondary Goal)

**우선순위**: Medium
**예상 LOC**: ~500

**작업 내용**:

**3a. Wizard (다단계 폼)**
- [ ] huh 기반 multi-step form 구현
- [ ] 6단계 흐름 구현: 프로젝트명 -> 언어 -> 프레임워크 -> 기능 -> 사용자명 -> 대화언어
- [ ] 언어 선택에 따른 프레임워크 목록 동적 필터링
- [ ] 이전 단계 네비게이션 (Back)
- [ ] context 취소 처리
- [ ] Headless 모드: SetDefaults 기반 즉시 WizardResult 반환
- [ ] 터미널 상태 복원 보장 (defer 패턴)
- [ ] 통합 테스트 + 스냅샷 테스트

**3b. Progress (진행률)**
- [ ] bubbles/progress 기반 determinate 프로그레스 바
- [ ] bubbles/spinner 기반 indeterminate 스피너
- [ ] Increment / SetTitle / Done API 구현
- [ ] Headless 모드: 로그 기반 진행률 출력
- [ ] NO_COLOR 모드: ASCII 프로그레스 바
- [ ] 유닛 테스트

**의존성**: Milestone 2 (Basic Components)
**완료 기준**: `moai init`에서 wizard가 완전히 동작하며, `moai update`에서 progress bar가 표시됨

### Milestone 4: NonInteractive & Polish (Final Goal)

**우선순위**: Low
**예상 LOC**: 기존 코드에 통합

**작업 내용**:
- [ ] NonInteractive 인터페이스 통합 구현
- [ ] TTY 감지 로직 (`os.Stdin` + `term.IsTerminal()`)
- [ ] `--non-interactive` CLI 플래그 연동
- [ ] 터미널 리사이즈 대응 (WindowSizeMsg 처리)
- [ ] 키보드 인터럽트(Ctrl+C) 안전 처리
- [ ] 에러 발생 시 터미널 복원 보장
- [ ] 전체 컴포넌트 통합 테스트
- [ ] godoc 문서 작성

**의존성**: Milestone 3
**완료 기준**: 모든 컴포넌트가 interactive/headless 양 모드에서 안정적으로 동작

---

## 3. 아키텍처 설계 방향

### 3.1 패키지 구조

```
internal/ui/
    theme.go        -- 색상 팔레트, 스타일 정의
    selector.go     -- selectorModel (bubbletea.Model)
    checkbox.go     -- checkboxModel (bubbletea.Model)
    prompt.go       -- promptModel, confirmModel
    wizard.go       -- wizardModel (huh.Form 기반)
    progress.go     -- progressModel, spinnerModel
    headless.go     -- NonInteractive 구현, TTY 감지
    ui.go           -- 공개 인터페이스 및 생성자 함수
```

### 3.2 의존성 주입 패턴

```go
// ui.go -- 팩토리 패턴
func NewWizard(theme *Theme, opts ...WizardOption) Wizard
func NewSelector(theme *Theme) Selector
func NewCheckbox(theme *Theme) Checkbox
func NewProgress(theme *Theme) Progress
func NewTheme(cfg config.ThemeConfig) *Theme
```

모든 컴포넌트는 Theme을 주입받아 시각적 일관성을 보장한다.

### 3.3 Headless 모드 전략

```
TTY 감지 -> Interactive(bubbletea)
         -> Headless(defaults 즉시 반환 + log 출력)
```

- `term.IsTerminal(os.Stdin.Fd())` 로 TTY 유무 판별
- CLI `--non-interactive` 플래그로 강제 headless 모드
- Headless 모드에서는 bubbletea Program을 생성하지 않음

---

## 4. 리스크 및 대응

### 4.1 기술 리스크

| 리스크 | 가능성 | 영향 | 대응 전략 |
|--------|-------|------|----------|
| bubbletea 학습 곡선 | 중 | 중 | Charmbracelet 공식 예제 기반 학습, huh 라이브러리로 wizard 구현 단순화 |
| 터미널 호환성 차이 | 중 | 중 | lipgloss의 플랫폼 추상화에 의존, cmd.exe는 NO_COLOR 모드로 폴백 |
| Windows 유니코드 문제 | 중 | 하 | lipgloss v1.0+의 Windows Terminal 지원에 의존, 비지원 터미널은 ASCII 폴백 |
| huh 라이브러리 안정성 | 하 | 중 | huh v0.6+는 안정 버전이며, 필요 시 bubbletea 직접 구현으로 대체 가능 |

### 4.2 의존성 리스크

| 의존성 | 리스크 | 대응 |
|--------|-------|------|
| SPEC-CONFIG-001 미완성 | 테마 설정을 읽을 수 없음 | 기본 테마를 하드코딩하고, config 연동은 후속 적용 |
| Charmbracelet 메이저 업데이트 | API 변경 가능성 | go.mod에서 버전 고정, 메이저 업데이트는 별도 SPEC으로 대응 |

### 4.3 통합 리스크

| 리스크 | 대응 |
|--------|------|
| CLI 명령과의 통합 인터페이스 불일치 | Go interface 기반 DDD 경계 설계로 계약 보장 |
| 다른 모듈의 UI 호출 패턴 불일치 | ui.go에 명확한 팩토리 함수 노출 |

---

## 5. 테스트 전략

### 5.1 테스트 카테고리

| 카테고리 | 대상 | 도구 | 커버리지 목표 |
|---------|------|------|-------------|
| 유닛 테스트 | Theme 팔레트, Model 상태 변경 | go test | 70% |
| 스냅샷 테스트 | View 출력 (bubbletea teatest) | bubbletea/teatest | 핵심 시나리오 |
| 통합 테스트 | Wizard 전체 흐름 | go test + teatest | 핵심 경로 |
| 헤드리스 테스트 | NonInteractive 경로 | go test | 100% |

### 5.2 테스트 패턴

```go
// bubbletea 스냅샷 테스트 예시
func TestSelectorView(t *testing.T) {
    m := newSelectorModel(testTheme, testItems)
    // 초기 상태 스냅샷
    got := m.View()
    golden.RequireEqual(t, got)

    // 키 입력 후 상태 변경
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
    got = m.View()
    golden.RequireEqual(t, got)
}
```

---

## 6. 관련 참조

| 참조 | 설명 |
|------|------|
| ADR-010 | Charmbracelet for Terminal UI 결정 근거 |
| design.md Section 3.12 | UI Module 인터페이스 정의 |
| structure.md internal/ui/ | 파일 구조 및 Python 대응 테이블 |
| #268 | ESC 키 freeze 이슈 (InquirerPy) |
| #249 | 인코딩 이슈 (cp949) |
| #286 | Windows 이모지 렌더링 이슈 |
