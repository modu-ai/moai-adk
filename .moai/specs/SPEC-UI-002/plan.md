# SPEC-UI-002: Statusline Rendering - 구현 계획

---
spec_id: SPEC-UI-002
document: plan
created: 2026-02-03
tags: statusline, ui, git, metrics, theme, claude-code
---

## 1. 구현 전략 개요

### 접근 방식

DDD(ANALYZE-PRESERVE-IMPROVE) 사이클을 적용하여 `internal/statusline/` 모듈을 구현한다. Python 원본의 동작 특성을 분석(ANALYZE)하고, Go 인터페이스 기반으로 재구현하되 Claude Code statusline 프로토콜과의 호환성을 보존(PRESERVE)하며, Go의 동시성과 타입 안전성을 활용하여 개선(IMPROVE)한다.

### 아키텍처 원칙

- **인터페이스 기반 경계**: 모든 Collector와 Builder는 Go 인터페이스로 정의하여 테스트 용이성 확보
- **병렬 데이터 수집**: errgroup.Group으로 4개 Collector를 동시 실행
- **Graceful Degradation**: 개별 Collector 실패 시 해당 섹션만 생략하고 나머지 정상 출력
- **Zero Side Effect**: statusline은 읽기 전용이며 파일 시스템 쓰기 금지

---

## 2. 마일스톤

### Milestone 1: 데이터 Collector 구현 (Primary Goal)

핵심 데이터 수집 레이어를 구현한다.

**구현 항목:**

- [ ] `memory.go` - MemoryCollector 구현
  - stdin JSON에서 `context_window.used`와 `context_window.total` 파싱
  - 토큰 사용률 백분율 계산
  - 필드 누락 시 기본값 처리
- [ ] `metrics.go` - MetricsCollector 구현
  - stdin JSON에서 `cost.total_usd`, `cost.input_tokens`, `cost.output_tokens` 파싱
  - `model` 필드 추출 및 약어 처리
  - 필드 누락 시 기본값 처리
- [ ] `git.go` - GitCollector 구현
  - SPEC-GIT-001의 Repository 인터페이스를 통한 Git 상태 수집
  - 브랜치명, modified/staged/untracked 파일 수, ahead/behind 카운트
  - Git 저장소 미존재 시 빈 GitStatusData 반환 (에러 아님)
- [ ] `update.go` - UpdateChecker 구현
  - `pkg/version`에서 현재 버전 조회
  - 캐시 기반 최신 버전 비교 (기본 캐시 TTL: 1시간)
  - 네트워크 오류 시 캐시 데이터 반환 또는 "unknown" 처리

**관련 요구사항:** REQ-E-002, REQ-E-003, REQ-E-004, REQ-E-005

**테스트 전략:**
- 각 Collector별 table-driven 단위 테스트
- Mock 인터페이스를 통한 의존성 격리
- 에러 케이스 및 필드 누락 케이스 검증

---

### Milestone 2: Builder 및 통합 Collector (Primary Goal)

수집된 데이터를 조합하는 Builder 레이어를 구현한다.

**구현 항목:**

- [ ] `builder.go` - defaultBuilder 구현
  - `Builder` 인터페이스 구현 (Build, SetMode)
  - stdin JSON 파싱 및 `StdinData` 구조체 정의
  - errgroup.Group으로 4개 Collector 병렬 실행
  - context.WithTimeout (1초) 전체 수집 타임아웃
  - Collector 개별 실패 시 partial data로 진행
  - StatuslineMode별 출력 섹션 결정 로직
- [ ] `Collector` 인터페이스 구현 (통합 Collector)
  - Builder 내부에서 개별 Collector들을 조합하여 `StatusData` 생성
  - StatusData 유효성 검증

**관련 요구사항:** REQ-U-001, REQ-U-002, REQ-U-003, REQ-U-005, REQ-E-001, REQ-S-004~006

**테스트 전략:**
- Builder 통합 테스트 (Mock Collector 주입)
- 병렬 실행 테스트 (`-race` 플래그)
- 타임아웃 시나리오 테스트
- 모드별 출력 검증

---

### Milestone 3: Renderer 및 테마 시스템 (Secondary Goal)

터미널 출력 렌더링과 테마 시스템을 구현한다.

**구현 항목:**

- [ ] `renderer.go` - Renderer 구현
  - lipgloss 기반 스타일 적용
  - StatusData를 단일 줄 문자열로 포맷팅
  - 컨텍스트 사용률에 따른 색상 분기 (녹색/황색/적색)
  - MOAI_NO_COLOR 환경변수 감지 시 plain text 모드
  - 모드별(minimal/default/verbose) 섹션 구성
  - 구분자(separator) 처리
  - 업데이트 가능 알림 렌더링

**테마 구조 구현:**
- [ ] Theme 구조체 정의 (lipgloss.Style 기반)
- [ ] 기본 테마: `default` (MoAI 컬러), `minimal` (무색), `nerd` (Nerd Font)
- [ ] SPEC-CONFIG-001에서 테마 설정 로드
- [ ] 런타임 테마 전환 지원

**관련 요구사항:** REQ-S-001~003, REQ-S-007, REQ-E-006, REQ-N-001, REQ-N-002

**테스트 전략:**
- 렌더링 결과 snapshot 테스트
- 색상 분기 로직 table-driven 테스트
- MOAI_NO_COLOR 환경 변수 테스트
- 각 테마별 출력 검증

---

### Milestone 4: 통합 및 성능 최적화 (Secondary Goal)

전체 파이프라인 통합과 성능 튜닝을 수행한다.

**구현 항목:**

- [ ] 진입점 통합
  - `moai hook statusline` CLI 서브커맨드 연결 (internal/cli/hook.go 확장)
  - stdin 읽기 -> Builder.Build() -> stdout 쓰기 파이프라인
  - 에러 발생 시 안전한 기본 출력 보장
- [ ] 성능 최적화
  - Benchmark 테스트 작성 (`*_bench_test.go`)
  - JSON 파싱 최적화 (필요 시 `encoding/json` 대신 lazy decoding)
  - 메모리 할당 최소화 (strings.Builder 재사용)
  - 업데이트 확인 결과 캐싱 (sync.Once 또는 TTL 캐시)
- [ ] settings.json 연동
  - `statusLine.command` 설정에 `moai hook statusline` 등록
  - 설정 템플릿 업데이트

**관련 요구사항:** REQ-U-002, REQ-N-003~005

**테스트 전략:**
- End-to-end 통합 테스트 (stdin/stdout 파이프)
- Benchmark 테스트 (P95 < 100ms 검증)
- 메모리 프로파일링 (< 5MB 검증)
- 경쟁 조건 테스트 (`go test -race`)

---

### Milestone 5: 선택 기능 및 문서화 (Optional Goal)

선택적 기능 구현과 문서화를 완료한다.

**구현 항목:**

- [ ] [Optional] TRUST 5 품질 점수 표시 (REQ-O-001)
- [ ] [Optional] 사용자 정의 섹션 순서 (REQ-O-002)
- [ ] [Optional] 세션 경과 시간 표시 (REQ-O-003)
- [ ] godoc 주석 작성 (모든 exported 타입/함수)
- [ ] 사용자 문서: statusline 설정 가이드

---

## 3. 기술 접근 방식

### 3.1 stdin JSON 처리

```go
// StdinData represents the JSON input from Claude Code.
type StdinData struct {
    HookEventName string        `json:"hook_event_name"`
    SessionID     string        `json:"session_id"`
    CWD           string        `json:"cwd"`
    Model         string        `json:"model"`
    Workspace     string        `json:"workspace"`
    Cost          *CostData     `json:"cost"`
    ContextWindow *ContextWindow `json:"context_window"`
}

type CostData struct {
    TotalUSD     float64 `json:"total_usd"`
    InputTokens  int     `json:"input_tokens"`
    OutputTokens int     `json:"output_tokens"`
}

type ContextWindow struct {
    Used  int `json:"used"`
    Total int `json:"total"`
}
```

- 포인터 필드(`*CostData`, `*ContextWindow`)를 사용하여 필드 존재 여부 구분
- `json.Decoder`로 stdin에서 직접 디코딩 (중간 버퍼 불필요)

### 3.2 병렬 수집 패턴

```go
func (b *defaultBuilder) collectAll(ctx context.Context, input *StdinData) *StatusData {
    ctx, cancel := context.WithTimeout(ctx, time.Second)
    defer cancel()

    g, ctx := errgroup.WithContext(ctx)
    data := &StatusData{}

    g.Go(func() error {
        result, err := b.gitCollector.CollectGit(ctx)
        if err != nil {
            slog.Debug("git collection failed", "error", err)
            return nil // non-fatal
        }
        data.Git = *result
        return nil
    })
    // ... memory, metrics, update collectors ...

    _ = g.Wait() // errors are non-fatal, logged
    return data
}
```

핵심 원칙:
- 개별 Collector 에러는 non-fatal 처리 (nil 반환)
- 전체 타임아웃은 context로 관리
- 부분 데이터로도 유효한 statusline 생성 가능

### 3.3 Renderer 설계

```go
func (r *Renderer) Render(data *StatusData, mode StatuslineMode) string {
    var sections []string

    switch mode {
    case ModeMinimal:
        sections = r.renderMinimal(data)
    case ModeDefault:
        sections = r.renderDefault(data)
    case ModeVerbose:
        sections = r.renderVerbose(data)
    }

    return strings.Join(sections, r.theme.Separator)
}
```

- 각 모드별 렌더링 메서드가 `[]string` (섹션 조각)을 반환
- 빈 섹션은 자동으로 필터링
- Separator로 섹션 간 구분

### 3.4 업데이트 확인 캐싱

```go
type UpdateChecker struct {
    currentVersion string
    cacheTTL       time.Duration
    mu             sync.RWMutex
    cachedResult   *VersionData
    cachedAt       time.Time
}

func (u *UpdateChecker) CheckUpdate(ctx context.Context) (*VersionData, error) {
    u.mu.RLock()
    if time.Since(u.cachedAt) < u.cacheTTL {
        defer u.mu.RUnlock()
        return u.cachedResult, nil
    }
    u.mu.RUnlock()
    // fetch latest version...
}
```

- 300ms마다 호출되므로 캐싱 필수
- 기본 TTL: 1시간 (설정 가능)
- 캐시 미스 시 비동기 갱신, 이전 캐시 반환

### 3.5 의존성 주입 구조

```go
// New creates a new statusline Builder with injected dependencies.
func New(opts Options) Builder {
    return &defaultBuilder{
        gitCollector: opts.GitCollector,
        config:       opts.Config,
        renderer:     newRenderer(opts.Theme),
        updateChecker: newUpdateChecker(opts.CurrentVersion, opts.CacheTTL),
        mode:         opts.Mode,
    }
}

type Options struct {
    GitCollector   GitCollectorInterface
    Config         config.Reader
    Theme          *Theme
    CurrentVersion string
    CacheTTL       time.Duration
    Mode           StatuslineMode
}
```

---

## 4. 위험 및 대응 계획

| 위험 | 영향 | 가능성 | 대응 |
|------|------|--------|------|
| SPEC-GIT-001 미완성 시 Git 데이터 수집 불가 | 높음 | 중간 | Mock GitCollector로 선 개발, 실제 연동은 SPEC-GIT-001 완료 후 |
| SPEC-CONFIG-001 미완성 시 테마 로드 불가 | 중간 | 중간 | 하드코딩 기본 테마로 선 개발, Config 연동은 후순위 |
| 300ms 응답 시간 초과 | 높음 | 낮음 | errgroup 타임아웃 1초, 개별 Collector 50ms 타임아웃, 캐싱 적극 활용 |
| lipgloss 터미널 호환성 문제 | 중간 | 낮음 | MOAI_NO_COLOR 폴백, CI 환경에서 자동 감지 |
| stdin JSON 스키마 변경 | 중간 | 낮음 | 포인터 필드로 유연한 파싱, 알 수 없는 필드 무시 |

---

## 5. 테스트 전략

### 5.1 단위 테스트 (파일별)

| 테스트 파일 | 대상 | 커버리지 목표 |
|-----------|------|-------------|
| `builder_test.go` | Builder.Build, SetMode, 병렬 수집 | 90% |
| `git_test.go` | GitCollector, Git 없는 환경 처리 | 85% |
| `metrics_test.go` | MetricsCollector, 필드 누락 처리 | 90% |
| `memory_test.go` | MemoryCollector, 비율 계산 | 90% |
| `renderer_test.go` | Render, 테마 적용, 색상 분기 | 90% |
| `update_test.go` | UpdateChecker, 캐싱 로직 | 85% |

### 5.2 벤치마크 테스트

| 벤치마크 | 대상 | 목표 |
|---------|------|------|
| BenchmarkBuild | 전체 Build 파이프라인 | < 100ms (P95) |
| BenchmarkParseStdin | JSON 파싱 | < 1ms |
| BenchmarkRender | Renderer.Render | < 5ms |
| BenchmarkGitCollect | GitCollector (mock) | < 10ms |

### 5.3 통합 테스트

- stdin/stdout 파이프 E2E 테스트
- 실제 Git 저장소에서의 GitCollector 테스트 (testdata/ 활용)
- 경쟁 조건 검증 (`go test -race ./internal/statusline/...`)

---

## 6. 구현 순서 요약

```
Phase 1: 데이터 수집 레이어
  memory.go -> metrics.go -> git.go -> update.go

Phase 2: 조합 레이어
  builder.go (StdinData 파싱 + errgroup 병렬 수집)

Phase 3: 출력 레이어
  renderer.go (테마 + ANSI + 모드별 포맷팅)

Phase 4: 통합
  CLI 연결 + settings.json + 벤치마크 + 최적화
```

---

## 7. 참조 문서

| 문서 | 경로 | 관련 내용 |
|------|------|----------|
| Product Document | `.moai/project/product.md` | Statusline 기능 정의 (Section 6) |
| Architecture Document | `.moai/project/structure.md` | internal/statusline/ 디렉토리 구조 |
| System Design | `.moai/project/design.md` | Statusline 인터페이스 (Section 3.13) |
| Technology Stack | `.moai/project/tech.md` | Charmbracelet, errgroup 기술 선택 |
| Claude Code Statusline | `.claude/skills/.../claude-code-statusline-official.md` | 프로토콜 스펙 |
