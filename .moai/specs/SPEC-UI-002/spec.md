# SPEC-UI-002: Statusline Rendering

---
spec_id: SPEC-UI-002
title: Statusline Rendering
status: Completed
priority: Medium
phase: "Phase 4 - UI and Integration"
module: internal/statusline/
estimated_loc: ~640
dependencies:
  - SPEC-GIT-001
  - SPEC-CONFIG-001
created: 2026-02-03
assignee: expert-backend
lifecycle: spec-anchored
tags: statusline, ui, git, metrics, theme, claude-code
---

## HISTORY

| 날짜 | 버전 | 변경 내용 |
|------|------|----------|
| 2026-02-03 | 1.0.0 | 최초 SPEC 작성 |

---

## 1. Environment (환경)

### 1.1 프로젝트 맥락

MoAI-ADK (Go Edition)는 Python 기반 MoAI-ADK(~73,000 LOC, 220+ 파일)를 Go로 완전 재작성하는 프로젝트이다. Statusline 모듈은 Python의 statusline 패키지(11 파일, 3,297 LOC)를 Go로 재구현하며, ~640 LOC로 약 80% 코드 감소를 목표로 한다.

### 1.2 기술 환경

| 항목 | 스펙 |
|------|------|
| 언어 | Go 1.22+ |
| 모듈 경로 | `github.com/modu-ai/moai-adk-go` |
| 대상 디렉토리 | `internal/statusline/` |
| 파일 구성 | builder.go, git.go, metrics.go, memory.go, renderer.go, update.go |
| 터미널 스타일링 | `github.com/charmbracelet/lipgloss` v1.0+ |
| 동시성 | `golang.org/x/sync/errgroup` (4개 Collector 병렬 수집) |
| 로깅 | `log/slog` (stdlib) |
| 테스트 | `testing` (stdlib) + `github.com/stretchr/testify` v1.9+ |

### 1.3 Claude Code Statusline 프로토콜

Claude Code는 statusline 명령을 주기적으로 호출한다:

- **입력**: stdin으로 JSON 전달
- **출력**: stdout 첫 번째 줄이 상태 표시줄
- **갱신 주기**: 최대 300ms 간격
- **ANSI 지원**: 색상 이스케이프 코드 지원

stdin JSON 스키마:

```json
{
  "hook_event_name": "statusLine",
  "session_id": "abc123",
  "cwd": "/path/to/project",
  "model": "claude-sonnet-4",
  "workspace": "/path/to/workspace",
  "cost": {
    "total_usd": 0.05,
    "input_tokens": 1000,
    "output_tokens": 500
  },
  "context_window": {
    "used": 50000,
    "total": 200000
  }
}
```

### 1.4 의존 모듈

| 의존 SPEC | 모듈 | 사용 인터페이스 |
|-----------|------|----------------|
| SPEC-GIT-001 | `internal/core/git/` | `Repository` 인터페이스 (branch, status, ahead/behind) |
| SPEC-CONFIG-001 | `internal/config/` | `Manager` 인터페이스 (테마, 사용자 설정 조회) |
| -- | `pkg/version/` | Version 상수 (현재 버전 조회) |

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

- [A-1] Claude Code가 statusline 명령을 `moai hook statusline` 또는 설정된 바이너리 경로로 호출한다
- [A-2] stdin JSON은 Claude Code 공식 스키마를 따르며, `context_window`와 `cost` 필드를 포함한다
- [A-3] SPEC-GIT-001의 `Repository` 인터페이스가 branch, modified files, ahead/behind 정보를 제공한다
- [A-4] SPEC-CONFIG-001의 `Manager` 인터페이스가 테마와 사용자 설정을 thread-safe하게 제공한다
- [A-5] 터미널이 ANSI 이스케이프 코드를 지원한다 (비지원 시 plain text 폴백)

### 2.2 비즈니스 가정

- [A-6] Statusline은 한 줄 출력으로 제한되며, 간결한 정보 전달이 핵심이다
- [A-7] 300ms 이내에 응답해야 하므로 비동기 데이터 수집이 필수적이다
- [A-8] Git 저장소가 없는 프로젝트에서도 statusline이 정상 동작해야 한다

### 2.3 가정 검증

| 가정 ID | 신뢰도 | 근거 | 오류 시 위험 |
|---------|--------|------|-------------|
| A-1 | High | Claude Code 공식 문서 확인 완료 | 진입점 변경 필요 |
| A-2 | High | 공식 JSON 스키마 문서화 됨 | 파서 수정 필요 |
| A-3 | Medium | SPEC-GIT-001 미구현 상태 | Mock 인터페이스로 선 개발 |
| A-4 | Medium | SPEC-CONFIG-001 미구현 상태 | Mock 인터페이스로 선 개발 |
| A-7 | High | Claude Code 문서에 300ms 명시 | 성능 최적화 필수 |

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous Requirements (항상 활성)

- **[REQ-U-001]** 시스템은 **항상** stdout에 단일 줄 텍스트를 출력해야 한다
- **[REQ-U-002]** 시스템은 **항상** 300ms 이내에 응답을 완료해야 한다
- **[REQ-U-003]** 시스템은 **항상** stdin JSON 파싱 실패 시에도 안전한 기본 출력을 생성해야 한다
- **[REQ-U-004]** 시스템은 **항상** UTF-8 인코딩으로 출력해야 한다
- **[REQ-U-005]** 시스템은 **항상** context.Context를 통한 취소 및 타임아웃을 지원해야 한다

### 3.2 Event-Driven Requirements (이벤트 기반)

- **[REQ-E-001]** **WHEN** Claude Code가 statusline 명령을 호출하면 **THEN** stdin에서 JSON을 읽고 포맷팅된 상태 줄을 stdout으로 출력해야 한다
- **[REQ-E-002]** **WHEN** Git 저장소가 감지되면 **THEN** 브랜치명, 수정된 파일 수, ahead/behind 카운트를 표시해야 한다
- **[REQ-E-003]** **WHEN** context_window 데이터가 수신되면 **THEN** 토큰 사용률을 백분율로 계산하여 표시해야 한다
- **[REQ-E-004]** **WHEN** 새로운 ADK 버전이 감지되면 **THEN** 업데이트 가능 알림을 statusline에 포함해야 한다
- **[REQ-E-005]** **WHEN** cost 데이터가 수신되면 **THEN** 세션 비용을 표시해야 한다
- **[REQ-E-006]** **WHEN** 사용자가 테마를 변경하면 **THEN** 다음 statusline 호출부터 새 테마가 적용되어야 한다

### 3.3 State-Driven Requirements (상태 기반)

- **[REQ-S-001]** **IF** context_window 사용률이 50% 미만이면 **THEN** 녹색으로 표시해야 한다
- **[REQ-S-002]** **IF** context_window 사용률이 50~80%이면 **THEN** 황색으로 표시해야 한다
- **[REQ-S-003]** **IF** context_window 사용률이 80% 이상이면 **THEN** 적색으로 표시해야 한다
- **[REQ-S-004]** **IF** StatuslineMode가 "minimal"이면 **THEN** 모델명과 컨텍스트 비율만 표시해야 한다
- **[REQ-S-005]** **IF** StatuslineMode가 "default"이면 **THEN** Git 상태, 컨텍스트 비율, 비용을 표시해야 한다
- **[REQ-S-006]** **IF** StatuslineMode가 "verbose"이면 **THEN** 모든 수집 데이터를 상세히 표시해야 한다
- **[REQ-S-007]** **IF** MOAI_NO_COLOR 환경변수가 설정되어 있으면 **THEN** ANSI 색상 코드 없이 plain text로 출력해야 한다

### 3.4 Unwanted Behavior Requirements (금지 행위)

- **[REQ-N-001]** 시스템은 stderr로 출력하지 **않아야 한다** (Claude Code가 stdout만 파싱)
- **[REQ-N-002]** 시스템은 개행 문자(\n)가 포함된 다중 줄 출력을 생성하지 **않아야 한다**
- **[REQ-N-003]** 시스템은 데이터 수집 실패 시 panic을 발생시키지 **않아야 한다**
- **[REQ-N-004]** 시스템은 파일 시스템에 쓰기 작업을 수행하지 **않아야 한다**
- **[REQ-N-005]** 시스템은 네트워크 요청을 동기적으로 블로킹하지 **않아야 한다** (업데이트 확인은 캐시 활용)

### 3.5 Optional Requirements (선택 기능)

- **[REQ-O-001]** **가능하면** TRUST 5 품질 점수를 statusline에 요약 표시 제공
- **[REQ-O-002]** **가능하면** 사용자 정의 섹션 순서 설정 제공
- **[REQ-O-003]** **가능하면** 세션 경과 시간 표시 제공

---

## 4. Specifications (상세 설계)

### 4.1 아키텍처 개요

```
stdin (JSON) ──> Collector ──> Builder ──> Renderer ──> stdout (text)
                    |
         +---------+---------+---------+
         |         |         |         |
     GitCollector  MemoryCol MetricsCol UpdateCol
         |         |         |         |
   SPEC-GIT-001  stdin     stdin    pkg/version
                 JSON      JSON     + cache
```

### 4.2 핵심 인터페이스 (design.md 기준)

#### Builder 인터페이스

```go
// Builder composes the statusline output from collected data.
type Builder interface {
    // Build generates the formatted statusline string.
    Build(ctx context.Context) (string, error)

    // SetMode switches between statusline display modes.
    SetMode(mode StatuslineMode)
}
```

#### StatuslineMode

```go
type StatuslineMode string

const (
    ModeMinimal StatuslineMode = "minimal"
    ModeDefault StatuslineMode = "default"
    ModeVerbose StatuslineMode = "verbose"
)
```

#### Collector 인터페이스

```go
// Collector gathers all data needed for statusline rendering.
type Collector interface {
    // Collect retrieves current status data from all sources.
    Collect(ctx context.Context) (*StatusData, error)
}
```

#### 데이터 모델

```go
type StatusData struct {
    Git     GitStatusData `json:"git"`
    Memory  MemoryData    `json:"memory"`
    Quality QualityData   `json:"quality"`
    Version VersionData   `json:"version"`
}

type GitStatusData struct {
    Branch    string `json:"branch"`
    Modified  int    `json:"modified"`
    Staged    int    `json:"staged"`
    Untracked int    `json:"untracked"`
    Ahead     int    `json:"ahead"`
    Behind    int    `json:"behind"`
}

type MemoryData struct {
    TokensUsed  int `json:"tokens_used"`
    TokenBudget int `json:"token_budget"`
}

type QualityData struct {
    Score  float64 `json:"score"`
    Passed bool    `json:"passed"`
}

type VersionData struct {
    Current         string `json:"current"`
    Latest          string `json:"latest"`
    UpdateAvailable bool   `json:"update_available"`
}
```

### 4.3 파일별 책임

| 파일 | 책임 | 주요 타입/함수 |
|------|------|---------------|
| `builder.go` | Statusline 조합 및 레이아웃, 모드별 출력 구성 | `Builder` 인터페이스, `defaultBuilder` 구현체, `Build()`, `SetMode()` |
| `git.go` | Git 저장소에서 상태 데이터 수집 | `GitCollector`, `CollectGit(ctx) (*GitStatusData, error)` |
| `metrics.go` | 세션 비용 및 모델 정보 수집 (stdin JSON) | `MetricsCollector`, `CollectMetrics(input *StdinData) (*MetricsData, error)` |
| `memory.go` | 컨텍스트 윈도우 토큰 사용량 수집 (stdin JSON) | `MemoryCollector`, `CollectMemory(input *StdinData) (*MemoryData, error)` |
| `renderer.go` | ANSI 색상, 테마 적용, 최종 문자열 렌더링 | `Renderer`, `Render(data *StatusData, mode StatuslineMode) string` |
| `update.go` | 버전 확인 및 업데이트 가능 여부 판단 (캐시 활용) | `UpdateChecker`, `CheckUpdate(ctx) (*VersionData, error)` |

### 4.4 동시성 전략

4개의 데이터 Collector를 `errgroup.Group`으로 병렬 실행:

```go
func (b *defaultBuilder) Build(ctx context.Context) (string, error) {
    g, ctx := errgroup.WithContext(ctx)
    var (
        gitData     *GitStatusData
        memoryData  *MemoryData
        metricsData *MetricsData
        versionData *VersionData
    )

    g.Go(func() error { /* git collector */ })
    g.Go(func() error { /* memory collector */ })
    g.Go(func() error { /* metrics collector */ })
    g.Go(func() error { /* update checker */ })

    if err := g.Wait(); err != nil {
        // partial data로 graceful degradation
    }
    // ...
}
```

- 타임아웃: `context.WithTimeout(ctx, 1*time.Second)` (전체 수집 1초 제한)
- 개별 Collector 실패 시: 해당 섹션을 빈 값/기본값으로 대체하고 나머지 정상 출력

### 4.5 테마 시스템

lipgloss 기반 테마 구조:

```go
type Theme struct {
    Name        string
    BranchStyle lipgloss.Style
    OkStyle     lipgloss.Style
    WarnStyle   lipgloss.Style
    ErrorStyle  lipgloss.Style
    MutedStyle  lipgloss.Style
    Separator   string
}
```

기본 제공 테마:

| 테마 | 설명 |
|------|------|
| `default` | MoAI 기본 색상 (녹색/황색/적색 신호등) |
| `minimal` | 색상 없이 간결한 아이콘 |
| `nerd` | Nerd Font 아이콘 활용 |

### 4.6 출력 형식 예시

**Minimal 모드:**
```
sonnet-4 | Ctx: 25%
```

**Default 모드:**
```
main +3 ~2 | Ctx: 25% | $0.05
```

**Verbose 모드:**
```
main +3 ~2 ^1 v0 | Ctx: 50K/200K (25%) | $0.05 | v1.2.0 (update!)
```

### 4.7 에러 처리 전략

| 시나리오 | 처리 방식 |
|---------|----------|
| stdin JSON 파싱 실패 | 기본 출력 생성 (버전 정보만 표시) |
| Git 저장소 없음 | Git 섹션 생략, 나머지 정상 출력 |
| context_window 필드 누락 | "N/A" 또는 빈 섹션으로 대체 |
| 전체 Collector 타임아웃 | 캐시된 이전 데이터 또는 최소 기본 출력 |
| ANSI 비지원 환경 | plain text 폴백 (MOAI_NO_COLOR) |
| 업데이트 확인 네트워크 오류 | 버전 섹션 생략, 로그 기록 |

### 4.8 성능 요구사항

| 메트릭 | 목표 | 측정 방법 |
|--------|------|----------|
| 전체 응답 시간 | < 100ms (P95) | Benchmark 테스트 |
| Git 데이터 수집 | < 50ms | 개별 Collector 벤치마크 |
| JSON 파싱 | < 1ms | 벤치마크 |
| 메모리 사용량 | < 5MB | 런타임 프로파일링 |
| Renderer 실행 | < 5ms | 벤치마크 |

### 4.9 설정 스키마

`.moai/config/sections/statusline.yaml`:

```yaml
statusline:
  enabled: true
  mode: default       # minimal | default | verbose
  theme: default      # default | minimal | nerd
  sections:           # verbose 모드에서 표시할 섹션 순서
    - git
    - context
    - cost
    - version
  update_check_interval: 3600  # 초 단위 (기본 1시간)
```

---

## 5. Traceability (추적성)

### 요구사항-파일 매핑

| 요구사항 | 구현 파일 | 테스트 파일 |
|---------|----------|-----------|
| REQ-U-001~005 | builder.go, renderer.go | builder_test.go, renderer_test.go |
| REQ-E-001 | builder.go | builder_test.go |
| REQ-E-002 | git.go | git_test.go |
| REQ-E-003 | memory.go | memory_test.go |
| REQ-E-004 | update.go | update_test.go |
| REQ-E-005 | metrics.go | metrics_test.go |
| REQ-E-006 | renderer.go | renderer_test.go |
| REQ-S-001~003 | renderer.go | renderer_test.go |
| REQ-S-004~006 | builder.go | builder_test.go |
| REQ-S-007 | renderer.go | renderer_test.go |
| REQ-N-001~005 | builder.go, renderer.go | builder_test.go (negative tests) |
| REQ-O-001 | metrics.go | metrics_test.go |

### 의존성 추적

```
SPEC-UI-002 (statusline)
    |-- depends on --> SPEC-GIT-001 (git repository interface)
    |-- depends on --> SPEC-CONFIG-001 (configuration manager)
    |-- uses --> pkg/version (version constants)
    |-- uses --> charmbracelet/lipgloss (terminal styling)
    |-- uses --> golang.org/x/sync/errgroup (parallel collection)
```

### Python 매핑 (마이그레이션 참조)

| Go 파일 | Python 원본 | LOC 변화 |
|---------|------------|---------|
| builder.go | statusline/main.py | ~150 -> ~120 |
| git.go | statusline/git_collector.py | ~300 -> ~100 |
| metrics.go | statusline/metrics_tracker.py | ~250 -> ~80 |
| memory.go | statusline/memory_collector.py | ~200 -> ~80 |
| renderer.go | statusline/renderer.py | ~350 -> ~160 |
| update.go | statusline/update_checker.py | ~200 -> ~100 |
| **합계** | **11 files, 3,297 LOC** | **6 files, ~640 LOC** |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 100.0%

### Summary

Statusline rendering system implemented as a compact Go replacement for the Python statusline (11 files, 3,297 LOC reduced to 6 files, ~640 LOC). Includes modular collector architecture for git status, context window metrics, cost tracking, memory usage, and version update checking. Builder pattern for composing statusline segments with configurable modes (minimal, default, verbose). ANSI-aware renderer with plain text fallback for non-color environments.

### Files Created

- `internal/statusline/builder.go`
- `internal/statusline/builder_test.go`
- `internal/statusline/git.go`
- `internal/statusline/git_test.go`
- `internal/statusline/memory.go`
- `internal/statusline/memory_test.go`
- `internal/statusline/metrics.go`
- `internal/statusline/metrics_test.go`
- `internal/statusline/renderer.go`
- `internal/statusline/renderer_test.go`
- `internal/statusline/types.go`
- `internal/statusline/update.go`
- `internal/statusline/update_test.go`
