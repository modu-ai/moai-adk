# SPEC-CLI-TUX-V3-001 — Design

## §A 아키텍처 — 3계층 출력 모델

```
[토큰]      internal/tui          — 색·글리프·테마의 유일한 원천 (문자열 토큰 SSOT)
   ↓ (plain string hex / glyph)
[컴포넌트]  internal/cli/uikit    — 배너·카드·에러 박스 (lipgloss v1 잔류, M1 무접촉)
            internal/tui 렌더 함수 — Box/Table/Pill/Stepper/ProgressLine (lipgloss v2로 포팅)
   ↓ (rendered string)
[채널]      internal/cli/printer  — 신규. 모든 커맨드의 유일한 출력 관문
            stdout = Data только / stderr = Info·Warn·Error·Success·Step
```

원칙(계획 보고서 §3 승계): 토큰 SSOT 유지(AC-CLI-TUI-013), 3계층 출력, stdout/stderr 규율 기계화, graceful degradation(TTY→rich, non-TTY/CI→plain, NO_COLOR/MOAI_REDUCED_MOTION 존중).

## §B 핵심 결정 — huh v1 / lipgloss v2 interop 경계: "문자열 토큰 경계" (String-Token Boundary)

### 문제

huh v1.0.0은 lipgloss v1에 의존한다(go.mod 실측). `internal/tui`를 lipgloss v2로 포팅할 때, tui가 lipgloss **타입**을 공개 API로 노출하면 huh v1 테마를 만드는 `internal/cli/wizard/wizard.go` `newMoAIWizardTheme`(lipgloss v1 Style 조립)와 statusline(lipgloss v1)이 컴파일 불가가 된다.

### 발견 (research.md §E)

실측 결과 **경계는 이미 문자열이다**:

- `internal/tui/theme.go`: "Colour tokens ... as plain strings; lipgloss.Color interprets hex sub-strings automatically" — Theme 구조체 필드는 전부 `string`.
- `internal/statusline/theme.go`: `lipgloss.Color(tuipkg.CatppuccinMochaPrimary)` — tui의 **문자열 상수**를 받아 statusline이 스스로 v1 타입으로 감싼다.
- `internal/cli/wizard`: `wizardColors()`가 tui Light/Dark 테마 토큰에서 색을 얻어 huh v1 스타일을 조립 (wizard.go:342-351).
- `internal/cli/uikit`: AC-CLI-TUI-013 준수 — hex 없음, tui 토큰 소비.

### 결정 (D-1)

**tui의 공개 계약을 "버전 중립 문자열 토큰"으로 동결하고, lipgloss는 tui의 내부 구현 세부로 강등한다.**

- tui 비테스트 소스의 lipgloss import는 `charm.land/lipgloss/v2`로 전환 (내부 렌더링 전용).
- Theme 구조체의 string 필드·`Catppuccin*` 상수·exported 함수 시그니처(문자열 in/문자열 out) 절대 불변 — 별도 `tui/compat` shim 패키지 **불필요** (문자열이 곧 shim).
- lipgloss v1 소비자(huh v1 경유 wizard/profile/update 폼, statusline, uikit)는 지금처럼 문자열을 각자 v1 타입으로 감싼다 — 소스 무수정 (AC-CTX-010).
- 금지: tui가 `lipgloss.Style`/`lipgloss.Color`(어느 메이저든)를 반환·인자로 받는 신규 export를 추가하는 것 (plan.md §G anti-pattern).

Go의 semantic import versioning 덕에 `github.com/charmbracelet/lipgloss`(v1)과 `charm.land/lipgloss/v2`는 서로 다른 모듈로 공존한다 — 충돌 없음 (AC-CTX-002).

### 파급

- statusline: 무접촉·무수정 (병렬 세션 보호와도 정합).
- wizard: huh v1 테마 경로 그대로 동작 — M2에서 huh v2 스파이크 시 이 경계 위에서 자연 승격.
- 후속 SPEC들이 v1 소비자를 하나씩 v2로 올려도 tui는 재수정 불필요.

## §C internal/cli/printer 설계

### 인터페이스 (계약 스케치 — 시그니처 확정은 run-phase TDD RED에서)

```go
// Printer is the single output gateway for all CLI commands.
type Printer interface {
    Info(msg string, args ...any)     // stderr
    Warn(msg string, args ...any)     // stderr
    Error(msg string, args ...any)    // stderr
    Success(msg string, args ...any)  // stderr
    Step(name string) StepHandle      // stderr; StepHandle: Update/Complete/Fail
    Data(v any) error                 // stdout; tty/plain=텍스트, json=구조화 라인
    Spinner(label string) SpinnerHandle // stderr; non-TTY/REDUCED_MOTION → 정적 라인
    Progress(label string, total int) ProgressHandle
}
```

### 3모드 (구성 시점 1회 선택)

| 모드 | 선택 조건 | 렌더 |
|---|---|---|
| tty-rich | stderr가 TTY && !NO_COLOR | tui 토큰 색 + 글리프(✓ ✗ ! ● ○ ◆ 화이트리스트) + bubbles v2 스피너(가능 시) |
| plain | non-TTY 또는 NO_COLOR 또는 CI | 무색·무escape 접두 라인 (`ok:`/`warn:`/`error:` 계열), 스피너→정적 단계 라인 |
| json | 미래 `--json`/`--plain` 전역 플래그 | 1-line JSON 이벤트 (`{"level":"warn","msg":...}`), Data는 구조화 페이로드 |

- 채널 규율은 모드와 독립: **어느 모드든 Data만 stdout** (REQ-CTX-012).
- ANSI 부재 단언은 plain/json 모드 공통 테스트 (REQ-CTX-014).
- `MOAI_REDUCED_MOTION`은 tty-rich 내 스피너/프로그레스만 정적으로 강등 (tui 기존 시맨틱 위임, REQ-CTX-007).

### ConsoleReporter 배선 (D-2)

`internal/core/project` 패키지의 `ProgressReporter` 인터페이스(StepStart/StepUpdate/StepComplete/StepError)는 유지하고, **Printer를 주입받는 어댑터 구현**으로 교체한다. 이유: ProgressReporter는 PhaseExecutor 등 core 계층의 계약이므로 인터페이스 자체 변경은 파급이 크다(M2에서 라이브 진행률로 승격 시 재방문). reporter.go의 hex `AdaptiveColor` 3개와 `fmt.Printf` stdout 직행은 이 어댑터로 흡수 (REQ-CTX-008/015). import 방향 주의: `internal/core/project`가 `internal/cli/printer`를 import하면 계층 역전 — 어댑터는 cli 쪽에 두고 core에는 인터페이스 구현체 주입.

### 대표 콜사이트 (M1 한정)

init.go 성공/경고 경로(경고 8곳 stdout 유출 실측) + ConsoleReporter. 잔여 46곳 중 나머지는 ratchet으로만 관리 (spec.md §C Out of Scope).

## §D internal/merge/confirm.go 결정 — **DEFER (bubbletea v1 잔류)** (D-3)

| 옵션 | 내용 | 평가 |
|---|---|---|
| A. v2 포팅 | 모델(43-224) + tea.NewProgram을 bubbletea v2 API로 재작성 | 비용: v2 API 학습 + 879 LOC 파일의 키 핸들링/렌더 재검증. **M3(SPEC-CLI-TUX-V3-003)가 동일 UI를 bubbles v2 list로 전면 승격 예정이라 곧바로 폐기될 작업** |
| B. v1 잔류 + 문서화된 연기 | @MX:DEBT 주석 + design 기록 + M3 트리거 명시 | 비용 0. lipgloss v1은 huh v1 때문에 어차피 잔류하므로 bubbletea v1 유지의 추가 모듈 부담 없음 |

**결정: B (연기).** 근거: (1) 계획 보고서 §4 M3 — "기존 confirm.go 체크박스 UI를 bubbles v2 list로 승격" — 이 재작성이 확정 로드맵이므로 M1의 직행 포팅은 이중 작업; (2) v1 모듈 그래프는 huh v1로 인해 M1 종료 후에도 존재 — bubbletea v1 잔류가 신규 부담을 만들지 않음; (3) M1 리스크 예산은 tui 포팅 + printer + fang에 집중하는 편이 우월.

기록 형식 (REQ-CTX-023, AC-CTX-023):

```go
// @MX:DEBT: bubbletea v1 program retained during Charm v2 migration (M1 deferral)
// @MX:CEILING: valid while huh v1 keeps lipgloss/bubbletea v1 in the module graph
// @MX:UPGRADE: SPEC-CLI-TUX-V3-003 promotes this UI to a charm.land/bubbles/v2 list component
```

## §E Fang 통합 설계 (D-4)

```go
// cmd/moai/main.go → internal/cli.Execute()
func Execute() error {
    initConsole()                    // Windows VT — fang보다 먼저 (acceptance §C edge case)
    if isTrivialCommand(os.Args[1:]) {
        initLightDeps()              // REQ-PERF-003-A 보존
    } else {
        InitDependencies()
    }
    return fang.Execute(context.Background(), rootCmd, /* fang options: tui 토큰 테마, version 단일 소스 */)
}
```

- **lazy-init 분기는 fang 바깥에 유지** — fang은 `rootCmd.Execute()` 자리만 대체 (REQ-CTX-019).
- **테마 주입**: fang의 색 구성(ColorScheme류 옵션)은 `internal/tui` 토큰 문자열에서 파생 — 신규 hex 금지 (REQ-CTX-022). 옵션 API의 정확한 시그니처는 run-phase 스파이크에서 확정 (research.md §F).
- **exit code**: `cmd/moai/main.go`의 ExitCoder 언래핑이 동작하려면 fang이 원본 error를 보존·반환해야 함 — characterization test 선행 (plan.md M1c-1). fang이 에러를 삼키는 형태라면 fang 옵션으로 에러 반환 모드를 선택하거나 rootCmd 레벨 래핑으로 우회 (run-phase 결정 지점, 결정 결과 progress.md 기록).
- **--version 단일 소스**: 기존 `Version: version.GetVersion()` 유지 — fang 자동 --version과 이중화 금지 (acceptance §C edge case).

## §F 리스크 매트릭스

| 리스크 | 영향 | 완화 |
|---|---|---|
| lipgloss v1→v2 렌더 차이 (배경 감지·다운샘플링·width) | tui 골든 대량 변동 | characterization 앵커 + 파일별 golden 사유 문서화 (REQ-CTX-009) |
| fang의 에러/exit code 가공 | worktree verify 0/1/2/3 파손 | M1c-1 characterization 선행 + ExitCoder 보존 확인 (REQ-CTX-020) |
| v2 API 상세 미확정 (본 세션 미실측) | 착수 지연·재작업 | run-phase 스파이크로 API 확정 후 포팅 규칙표 작성 (plan.md M1a-3) |
| trivial fast-path 회귀 | `moai --version` 기동 지연 | LazyInit 테스트 (REQ-CTX-019) |
| 병렬 세션 statusline 충돌 | 커밋 오염 | PRESERVE 목록 + specific-path add (plan.md §D) |
| charm.land 프록시 가용성 | CI 불안정 | GOPROXY 표준 경로 + go.sum 커밋 |

## §G 검토 후 기각한 대안

- **tui/compat shim 패키지 신설**: 문자열 토큰 경계가 이미 존재하므로 불필요한 간접층 — 기각 (Simplicity First).
- **uikit 동시 v2 포팅**: uikit은 문자열 토큰만 소비하므로 v1 잔류 무해; M1 범위 팽창 방지 위해 기각 (M4에서 배너 경량화와 함께 자연 승격).
- **ProgressReporter 인터페이스 즉시 교체**: core 계층 계약 변경은 M2(라이브 진행률)에서 UI 요구와 함께 — 기각, 어댑터로 한정 (§C D-2).
- **printer를 internal/tui 하위에 배치**: tui는 순수 렌더 커널(채널 무지)로 유지 — 채널 규율은 cli 계층 관심사이므로 기각.
- **pterm/survey/tview 대체 스택**: 계획 보고서 §2에서 기각 완료 (스타일 통일성·유지보수·과설계) — 재론 불요.
