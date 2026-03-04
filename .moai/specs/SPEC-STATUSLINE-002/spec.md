---
id: SPEC-STATUSLINE-002
version: "1.0.0"
status: Draft
created: "2026-03-04"
updated: "2026-03-04"
author: GOOS
priority: P2-Medium
---

# SPEC-STATUSLINE-002: Statusline 토큰 사용량 표시 (5h/7d)

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-03-04 | GOOS | 최초 작성 (이슈 #464 기반) |

---

## 1. 개요

### 1.1 목적

MoAI-ADK 상태바(statusline)에 API 토큰 누적 사용량을 실시간으로 표시한다. 사용자가 5시간 및 7일 단위의 토큰 소비 현황을 한눈에 파악하고, 설정 가능한 임계값 초과 시 경고를 확인할 수 있도록 한다.

### 1.2 범위

- 모듈 경로: `internal/statusline/`, `internal/hook/`
- 신규 파일: `internal/statusline/token_history.go`, `internal/statusline/token_history_test.go`
- 설정 파일: `.moai/config/sections/statusline.yaml` (기존 파일 확장)
- 캐시 파일: `.moai/cache/token_history.json`
- 의존성: SPEC-STATUSLINE-001 (statusline 세그먼트 설정 시스템)

### 1.3 배경

Claude Code의 statusline 훅은 세션 단위 토큰 사용량(`CostData.InputTokens`, `CostData.OutputTokens`)을 JSON stdin으로 전달한다. 현재 statusline은 세션 내 비용(`$X.XX`)만 표시하며, 시간대별 누적 집계는 지원하지 않는다. 다중 세션에 걸친 집계는 로컬 캐시 파일에 세션 레코드를 축적하여 구현한다.

---

## 2. Environment (환경)

### 2.1 기술 스택

| 구성 요소 | 기술 | 버전 |
|-----------|------|------|
| 언어 | Go | 1.22+ |
| 캐시 직렬화 | `encoding/json` (stdlib) | Go 1.22+ |
| 원자적 파일 쓰기 | `os.Rename` (stdlib) | Go 1.22+ |
| 동시성 제어 | `sync.RWMutex` (stdlib) | Go 1.22+ |
| 시간 처리 | `time` (stdlib) | Go 1.22+ |
| 설정 | `gopkg.in/yaml.v3` | v3.0+ |
| 테스트 | `github.com/stretchr/testify` | v1.9+ |

### 2.2 Claude Code 데이터 소스

Claude Code가 statusline 훅에 전달하는 stdin JSON 구조:

```json
{
  "session_id": "abc-123",
  "cost": {
    "input_tokens": 12500,
    "output_tokens": 3200,
    "total_cost_usd": 0.45
  },
  "context_window": {
    "total_input_tokens": 45000,
    "total_output_tokens": 8500
  }
}
```

`session_id`를 키로 하여 세션별 토큰 데이터를 캐시 파일에 누적 저장한다.

### 2.3 캐시 파일 위치

| 파일 | 설명 |
|------|------|
| `.moai/cache/token_history.json` | 세션별 토큰 사용 이력 (로컬 전용, git 미추적) |

### 2.4 설정 파일 위치

| 파일 | 설명 |
|------|------|
| `.moai/config/sections/statusline.yaml` | statusline 세그먼트 설정 (기존 파일 확장) |

### 2.5 환경 제약

- Anthropic API는 시간대별 토큰 사용량 조회 엔드포인트를 제공하지 않으므로, 모든 집계는 로컬 캐시 기반으로 수행한다.
- 캐시 파일은 `.moai/cache/` 디렉토리에 위치하며 `.gitignore`로 제외되어야 한다.
- 캐시 파일 부재 시 세그먼트는 "데이터 없음"(`--`) 상태로 우아하게 폴백한다.

---

## 3. Assumptions (가정)

- **A1**: `session_id`는 Claude Code가 세션 기간 내 일관되게 제공하며, 세션 간 충돌이 없다.
- **A2**: 단일 캐시 JSON 파일의 크기는 실용적 사용량 기준(90일 이내 세션)에서 1MB 이하로 유지된다.
- **A3**: 기존 `statusline.yaml`의 `segments` 맵에 새 키(`token_usage_5h`, `token_usage_7d`)를 추가해도 3-way YAML 병합 시스템과 하위 호환성이 유지된다.
- **A4**: statusline 훅은 세션 중 주기적으로 호출되므로, 세션 종료 전에도 토큰 데이터가 캐시에 기록된다.
- **A5**: `.moai/cache/` 디렉토리는 `moai init` 또는 첫 번째 캐시 쓰기 시 자동 생성된다.
- **A6**: 토큰 사용량 경고 색상(ANSI 컬러)은 기존 context window 레벨 시스템(`levelOk`, `levelWarn`, `levelError`)을 재사용할 수 있다.
- **A7**: 사용자가 `statusline.yaml`에 임계값을 설정하지 않으면 경고 색상 없이 정보만 표시한다.
- **A8**: `SessionStop` 이벤트가 항상 발생한다고 보장할 수 없으므로, statusline 렌더 시마다 캐시를 업데이트한다.

---

## 4. Requirements (요구사항)

### 4.1 토큰 이력 집계 (Token History Aggregation)

**REQ-TU-001** (Ubiquitous)
시스템은 statusline 렌더 호출 시마다 현재 세션의 토큰 사용량(`session_id`, `input_tokens`, `output_tokens`, `cost_usd`, 타임스탬프)을 `.moai/cache/token_history.json`에 upsert 방식으로 기록해야 한다.

**REQ-TU-002** (Ubiquitous)
시스템은 `token_history.json`에서 현재 시각 기준 5시간 이내 세션의 토큰 합계(input + output)를 집계해야 한다.

**REQ-TU-003** (Ubiquitous)
시스템은 `token_history.json`에서 현재 시각 기준 7일 이내 세션의 토큰 합계(input + output)를 집계해야 한다.

**REQ-TU-004** (Event-Driven)
세션의 `last_update` 타임스탬프가 시간 윈도우 경계에 걸치는 경우, 해당 세션 전체 토큰을 포함한다(세션을 분할하지 않는다).

**REQ-TU-005** (Unwanted)
시스템은 90일을 초과한 세션 레코드를 자동으로 정리(purge)해야 한다. 단, 정리 주기는 매 렌더 호출이 아닌 24시간에 1회로 제한한다.

**REQ-TU-006** (Event-Driven)
`token_history.json` 파일이 존재하지 않거나 파싱에 실패하는 경우, 시스템은 빈 이력으로 시작하고 오류를 반환하지 않는다.

### 4.2 캐시 파일 관리 (Cache Management)

**REQ-TU-010** (Ubiquitous)
캐시 파일 쓰기는 임시 파일 + `os.Rename` 패턴(원자적 교체)을 사용하여 파일 손상을 방지해야 한다.

**REQ-TU-011** (Ubiquitous)
캐시 파일 읽기/쓰기는 `sync.RWMutex`로 보호되어 동시 접근에 안전해야 한다.

**REQ-TU-012** (Ubiquitous)
캐시 파일 전체 읽기+파싱+쓰기는 50ms 이내에 완료되어야 한다.

**REQ-TU-013** (Event-Driven)
`session_id`가 빈 문자열인 경우, 캐시 업데이트를 건너뛰고 현재 세션 데이터만 집계에 포함하지 않는다.

### 4.3 UI 표시 형식 (Display Format)

**REQ-TU-020** (Ubiquitous)
statusline에 두 개의 새 세그먼트를 추가한다: `token_usage_5h`(5시간 집계), `token_usage_7d`(7일 집계).

**REQ-TU-021** (Ubiquitous)
기본 표시 형식은 다음과 같다:

```
5h:45K 7d:328K
```

토큰 수 1,000 이상은 `K` 단위로 표시하며, 소수점은 표시하지 않는다.

**REQ-TU-022** (State-Driven)
데이터가 없거나(캐시 파일 없음, 집계 결과 0) 세그먼트가 비활성화된 경우 해당 세그먼트를 렌더링에서 제외한다.

**REQ-TU-023** (State-Driven)
5h 또는 7d 토큰 사용량이 설정된 임계값을 초과하는 경우, 해당 값에 경고 색상(노란색 또는 빨간색)을 적용한다.

**REQ-TU-024** (Ubiquitous)
`NO_COLOR` 또는 `MOAI_NO_COLOR` 환경 변수가 설정된 경우, 경고 색상을 적용하지 않는다.

### 4.4 경고 임계값 (Warning Thresholds)

**REQ-TU-030** (Ubiquitous)
시스템은 `statusline.yaml`에 다음 임계값 설정을 지원해야 한다:

| 설정 키 | 타입 | 기본값 | 설명 |
|---------|------|--------|------|
| `token_usage.warn_5h_tokens` | int | 0 (비활성) | 5시간 토큰 임계값 |
| `token_usage.warn_7d_tokens` | int | 0 (비활성) | 7일 토큰 임계값 |
| `token_usage.critical_5h_tokens` | int | 0 (비활성) | 5시간 위험 임계값 |
| `token_usage.critical_7d_tokens` | int | 0 (비활성) | 7일 위험 임계값 |

**REQ-TU-031** (State-Driven)
임계값이 0인 경우, 해당 임계값에 대한 경고 색상을 적용하지 않는다(비활성).

**REQ-TU-032** (State-Driven)
토큰 사용량이 `warn` 임계값 이상 `critical` 미만인 경우 노란색으로, `critical` 이상인 경우 빨간색으로 표시한다.

### 4.5 사용자 커스터마이징 (User Customization)

**REQ-TU-040** (Ubiquitous)
시스템은 `statusline.yaml`의 `segments` 맵에서 `token_usage_5h`, `token_usage_7d` 키를 인식하고, `false`로 설정 시 해당 세그먼트를 숨긴다.

**REQ-TU-041** (Ubiquitous)
시스템은 표시할 시간 윈도우를 사용자가 설정할 수 있도록 `statusline.yaml`에 다음 설정을 지원한다:

```yaml
token_usage:
  windows:
    - period: 5h    # 5시간 (기본 활성)
      enabled: true
    - period: 7d    # 7일 (기본 활성)
      enabled: true
```

**REQ-TU-042** (Ubiquitous)
지원 가능한 시간 윈도우 값: `1h`, `3h`, `5h`, `12h`, `24h`, `3d`, `7d`, `30d`.

**REQ-TU-043** (Ubiquitous)
`moai init` 및 `moai update -c` 위저드에서 토큰 사용량 표시 세그먼트의 활성화 여부를 선택할 수 있어야 한다.

**REQ-TU-044** (Ubiquitous)
기존 `statusline.yaml`에 `token_usage` 설정이 없는 경우, 기본값(5h + 7d 활성, 임계값 없음)으로 동작한다.

---

## 5. Specifications (명세)

### 5.1 캐시 파일 스키마

```json
{
  "version": "1",
  "last_purge": "2026-03-04T00:00:00Z",
  "sessions": {
    "session-id-abc123": {
      "session_id": "session-id-abc123",
      "start_time": "2026-03-04T10:00:00Z",
      "last_update": "2026-03-04T10:30:00Z",
      "input_tokens": 12500,
      "output_tokens": 3200,
      "cost_usd": 0.45
    }
  }
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `version` | string | 캐시 스키마 버전 (마이그레이션 대비) |
| `last_purge` | RFC3339 | 마지막 오래된 세션 정리 시각 |
| `sessions` | map[string]SessionRecord | session_id → 세션 레코드 |
| `SessionRecord.start_time` | RFC3339 | 세션 최초 기록 시각 |
| `SessionRecord.last_update` | RFC3339 | 마지막 업데이트 시각 (집계 기준) |
| `SessionRecord.input_tokens` | int | 누적 입력 토큰 수 |
| `SessionRecord.output_tokens` | int | 누적 출력 토큰 수 |
| `SessionRecord.cost_usd` | float64 | 누적 비용(USD) |

### 5.2 신규 Go 타입

```go
// internal/statusline/token_history.go

// TokenHistoryCache는 로컬에 영속화된 세션별 토큰 사용 이력이다.
type TokenHistoryCache struct {
    Version   string                    `json:"version"`
    LastPurge time.Time                 `json:"last_purge"`
    Sessions  map[string]*SessionRecord `json:"sessions"`
}

// SessionRecord는 단일 세션의 토큰 사용 요약이다.
type SessionRecord struct {
    SessionID    string    `json:"session_id"`
    StartTime    time.Time `json:"start_time"`
    LastUpdate   time.Time `json:"last_update"`
    InputTokens  int       `json:"input_tokens"`
    OutputTokens int       `json:"output_tokens"`
    CostUSD      float64   `json:"cost_usd"`
}

// TokenUsageData는 특정 시간 윈도우의 집계 결과이다.
type TokenUsageData struct {
    WindowLabel  string // 예: "5h", "7d"
    TotalTokens  int
    InputTokens  int
    OutputTokens int
    Available    bool
}

// TokenHistoryStore는 캐시 파일 읽기/쓰기를 추상화한다.
type TokenHistoryStore interface {
    // UpsertSession은 현재 세션 데이터를 캐시에 upsert한다.
    UpsertSession(ctx context.Context, record *SessionRecord) error

    // Aggregate는 지정된 기간 내 토큰 합계를 반환한다.
    // duration: 예) 5*time.Hour, 7*24*time.Hour
    Aggregate(ctx context.Context, duration time.Duration) (*TokenUsageData, error)

    // Purge는 retentionPeriod보다 오래된 레코드를 삭제한다.
    // 24시간에 1회만 실제 삭제가 수행된다.
    Purge(ctx context.Context, retentionPeriod time.Duration) error
}
```

### 5.3 statusline.yaml 설정 확장

```yaml
# .moai/config/sections/statusline.yaml
statusline:
  preset: "full"

  segments:
    model: true
    context: true
    output_style: true
    directory: true
    git_status: true
    claude_version: true
    moai_version: true
    git_branch: true
    token_usage_5h: true   # 신규
    token_usage_7d: true   # 신규

  # 토큰 사용량 표시 설정 (신규)
  token_usage:
    # 표시할 시간 윈도우 목록
    windows:
      - period: "5h"
        enabled: true
      - period: "7d"
        enabled: true

    # 경고 임계값 (0 = 비활성)
    warn_5h_tokens: 0
    warn_7d_tokens: 0
    critical_5h_tokens: 0
    critical_7d_tokens: 0
```

### 5.4 StatusData 확장

```go
// internal/statusline/types.go에 추가

// TokenUsageWindow는 단일 시간 윈도우의 토큰 사용량이다.
type TokenUsageWindow struct {
    Label       string // "5h", "7d" 등
    TotalTokens int
    Available   bool
    Level       contextLevel // levelOk, levelWarn, levelError
}

// StatusData에 추가할 필드
type StatusData struct {
    // ... 기존 필드 ...
    TokenUsage5h TokenUsageWindow
    TokenUsage7d TokenUsageWindow
}
```

### 5.5 표시 형식 상세

| 상태 | 표시 예시 | 설명 |
|------|-----------|------|
| 정상 | `5h:12K 7d:89K` | 임계값 미초과 |
| 경고 | `5h:⚠120K 7d:89K` | 5h가 warn 임계값 초과 |
| 위험 | `5h:🔴180K 7d:89K` | 5h가 critical 임계값 초과 |
| 데이터 없음 | *(세그먼트 숨김)* | 캐시 파일 없거나 집계 결과 0 |
| NO_COLOR | `5h:12K 7d:89K` | 색상 없이 텍스트만 |

> **참고**: 이모지 표시 여부는 `MOAI_NO_COLOR` 설정 및 터미널 지원 여부에 따라 결정한다. 색상 비활성화 시 이모지도 생략한다.

### 5.6 토큰 집계 알고리즘

```
1. stdin JSON에서 session_id, input_tokens, output_tokens, cost_usd 추출
2. session_id가 비어있으면 캐시 업데이트 스킵
3. 캐시 파일에서 기존 세션 레코드 로드 (없으면 빈 캐시로 시작)
4. session_id로 기존 레코드 조회:
   - 존재하면: last_update, input_tokens, output_tokens, cost_usd 업데이트
   - 없으면: 신규 레코드 생성 (start_time = now)
5. 변경된 캐시를 원자적으로 파일에 저장
6. 현재 시각 기준 5h, 7d 범위 내 레코드의 토큰 합산
7. 결과를 StatusData에 저장하여 렌더러에 전달
```

---

## 6. Traceability (추적성)

| 요구사항 | 구현 파일 | 테스트 파일 |
|---------|-----------|-------------|
| REQ-TU-001~006 | `internal/statusline/token_history.go` | `internal/statusline/token_history_test.go` |
| REQ-TU-010~013 | `internal/statusline/token_history.go` | `internal/statusline/token_history_test.go` |
| REQ-TU-020~024 | `internal/statusline/renderer.go` | `internal/statusline/renderer_test.go` |
| REQ-TU-030~032 | `internal/statusline/token_history.go`, `renderer.go` | `token_history_test.go`, `renderer_test.go` |
| REQ-TU-040~044 | `internal/statusline/builder.go`, `internal/cli/statusline.go` | `builder_test.go` |

---

## 7. Constraints (제약사항)

### 성능

| 항목 | 제약값 |
|------|--------|
| 캐시 파일 I/O (읽기+파싱+쓰기) | 50ms 이내 |
| statusline 전체 렌더 추가 지연 | 10ms 이내 |
| 캐시 파일 최대 크기 | 1MB 이내 (90일 보존 기준) |

### 보안

- 캐시 파일은 로컬 파일 시스템에만 저장하며, 외부 API 호출 없음
- 캐시 파일 파싱 실패 시 오류 로그를 남기되 프로세스를 중단하지 않음

### 호환성

- 기존 `statusline.yaml`에 `token_usage` 섹션이 없을 경우 기본값 사용 (하위 호환)
- 기존 8개 세그먼트 렌더링에 영향 없음 (`token_usage_5h`, `token_usage_7d`는 선택적 추가)

---

## 8. Implementation Notes (구현 참고)

### 8.1 구현 우선순위

1. **Phase 1**: `token_history.go` - 캐시 읽기/쓰기, 집계 로직 (핵심)
2. **Phase 2**: `renderer.go` - 새 세그먼트 렌더링 + 경고 색상
3. **Phase 3**: `statusline.yaml` 템플릿 업데이트 + `make build`
4. **Phase 4**: 위저드 통합 (선택적, P3 낮은 우선순위)

### 8.2 관련 파일

| 파일 | 작업 |
|------|------|
| `internal/statusline/token_history.go` | 신규: TokenHistoryStore 구현 |
| `internal/statusline/types.go` | 수정: TokenUsageWindow, StatusData 확장 |
| `internal/statusline/builder.go` | 수정: TokenHistoryStore 통합 |
| `internal/statusline/renderer.go` | 수정: token_usage_5h, token_usage_7d 세그먼트 렌더 |
| `internal/cli/statusline.go` | 수정: statusline.yaml에서 token_usage 설정 로드 |
| `internal/template/templates/.moai/config/sections/statusline.yaml` | 수정: token_usage 섹션 추가 |

### 8.3 테스트 전략

- `token_history.go`: 단위 테스트 (파일 I/O는 `t.TempDir()` 사용)
- 시간 의존 로직: `time.Now` 주입 또는 `clockwork` 패턴
- 렌더러: 기존 테이블 드리븐 테스트 패턴 확장
- 목표 커버리지: 85%+
