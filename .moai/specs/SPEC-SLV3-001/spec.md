# SPEC-SLV3-001: Statusline v3 Upgrade

| 항목 | 내용 |
|------|------|
| SPEC ID | SPEC-SLV3-001 |
| 제목 | Statusline v3 Upgrade |
| 버전 | 1.0.0 |
| 상태 | Completed |
| 우선순위 | High |
| 관련 SPEC | SPEC-SLE-001 (Statusline Engine), SPEC-STATUSLINE-001 |
| 태그 | `statusline`, `v3`, `gradient`, `usage-monitoring` |

---

## 1. 개요 (Overview)

moai-adk-go의 statusline 시스템을 v3로 업그레이드한다. 모드 이름 변경(minimal/verbose -> compact/full), 모드별 라인 레이아웃 재설계, RGB 연속 그라디언트 프로그레스 바, 세션 시간 표시, Git ahead/behind 렌더링, API 사용량 모니터링(5H/7D) 기능을 추가한다.

디자인 영감: [awesome-claude-plugins](https://github.com/nicobailon/awesome-claude-plugins) 커뮤니티의 statusline 구현 사례들을 참고하여, 더 풍부한 정보를 제공하면서도 모드별 밀도를 최적화한 레이아웃을 설계했다.

---

## 2. 배경 (Background)

### 현재 상태

- **모드**: `ModeMinimal="minimal"`, `ModeDefault="default"`, `ModeVerbose="verbose"` (types.go:15-24)
- **프로그레스 바**: `buildBar()` (renderer.go:319) - 12자 너비, 4단계 그라디언트 (`BarGradient()` in theme.go)
- **Git ahead/behind**: `git.go:50-51`에서 이미 수집 중이나, `renderGitStatus()` (renderer.go:344-366)에서 렌더링하지 않음
- **세션 시간**: `CostData.TotalDurationMS` (types.go:65)로 입력 수신하지만 `MetricsData`로 추출되지 않음
- **Provider 패턴**: `GitDataProvider`, `UpdateProvider` 인터페이스 기반, `defaultBuilder`에서 병렬 실행
- **세그먼트 구성**: `segmentConfig map[string]bool`로 세그먼트 on/off 제어
- **테마 시스템**: `Theme` 인터페이스, `BarGradient(percentage float64) lipgloss.Color`

### 동기

1. 모드 이름이 직관적이지 않음 (minimal/verbose vs compact/full)
2. 프로그레스 바가 12자로 작고 4단계 색상만 지원하여 시각적으로 부족
3. 이미 수집된 Git ahead/behind 데이터가 사용자에게 노출되지 않음
4. 세션 시간 정보가 있으나 표시되지 않음
5. API 사용량(5시간/7일) 모니터링 부재로 사용 패턴 파악 불가

---

## 3. 요구사항 (Requirements)

### GAP-1: 모드 이름 변경 (Mode Rename)

> minimal -> compact, verbose -> full, 하위 호환성 유지

**REQ-V3-MODE-001** (State-Driven):
**IF** mode="minimal" **THEN** "compact"로 처리해야 한다. 기존 설정 파일과의 하위 호환성을 위해 "minimal" 값을 수신하면 내부적으로 "compact"로 매핑한다.

**REQ-V3-MODE-002** (State-Driven):
**IF** mode="verbose" **THEN** "full"로 처리해야 한다. 기존 설정 파일과의 하위 호환성을 위해 "verbose" 값을 수신하면 내부적으로 "full"로 매핑한다.

**REQ-V3-MODE-003** (Event-Driven):
**WHEN** 프로파일 위저드가 모드 선택을 표시할 때 **THEN** compact/default/full 이름으로 표시해야 한다.

**REQ-V3-MODE-004** (Ubiquitous):
시스템은 **항상** 상수를 `ModeCompact`, `ModeDefault`, `ModeFull`로 사용해야 한다. 기존 `ModeMinimal`, `ModeVerbose`는 deprecated 주석과 함께 alias로 유지한다.

### GAP-2: 레이아웃 변경 (Layout Change)

> compact 2L, default 2L, full 5L

**REQ-V3-LAYOUT-001** (State-Driven):
**IF** mode="compact" **THEN** 2줄 레이아웃으로 렌더링해야 한다.

**REQ-V3-LAYOUT-002** (State-Driven):
**IF** mode="default" **THEN** 4줄 레이아웃으로 렌더링해야 한다.

**REQ-V3-LAYOUT-003** (State-Driven):
**IF** mode="full" **THEN** 6줄 레이아웃으로 렌더링해야 한다.

**REQ-V3-LAYOUT-004** (State-Driven):
**IF** 특정 줄의 모든 세그먼트가 비어 있으면 **THEN** 해당 줄을 생략해야 한다. 빈 줄 없이 밀착 렌더링한다.

### GAP-3: RGB 그라디언트 프로그레스 바 (RGB Gradient Progress Bar)

> 40블록 연속 보간, ANSI 24-bit 색상

**REQ-V3-BAR-001** (Ubiquitous):
시스템은 **항상** 프로그레스 바를 모드에 따라 렌더링해야 한다. full 모드는 40블록(각 바 독립 줄), default/compact 모드는 10블록. 모든 바에 `CW:`, `5H:`, `7D:` 라벨을 앞에 표시한다.

**REQ-V3-BAR-002** (Ubiquitous):
시스템은 **항상** 연속 RGB 보간 그라디언트를 적용해야 한다. 블록별로 개별 색상을 계산한다.
- 경로: Green(0,255,0) -> Yellow(255,255,0) -> Red(255,0,0)
- 0-50%: Green -> Yellow 선형 보간
- 50-100%: Yellow -> Red 선형 보간

**REQ-V3-BAR-003** (State-Driven):
**IF** NO_COLOR 환경변수가 설정되어 있거나 noColor=true **THEN** 색상 없이 Unicode 블록 문자만 사용해야 한다.

**REQ-V3-BAR-004** (State-Driven):
**IF** 터미널 너비가 바 너비보다 작으면 **THEN** 비례적으로 축소하여 렌더링해야 한다.

### GAP-4: 세션 시간 추적 (Session Time Tracking)

> CostData.TotalDurationMS 활용

**REQ-V3-TIME-001** (Event-Driven):
**WHEN** `CostData.TotalDurationMS`가 수신되면 **THEN** `MetricsData.SessionDurationMS` 필드로 추출해야 한다.

**REQ-V3-TIME-002** (Ubiquitous):
시스템은 **항상** 세션 시간을 `⏳ Xh Ym` 형식으로 표시해야 한다 (예: `⏳ 1h 23m`). 1시간 미만이면 `⏳ Xm` (예: `⏳ 45m`). 아이콘은 모래시계(⏳)를 사용한다.

**REQ-V3-TIME-003** (Ubiquitous):
시스템은 **항상** 새로운 세그먼트 상수 `SegmentSessionTime = "session_time"`을 정의해야 한다.

**REQ-V3-TIME-004** (State-Driven):
**IF** `TotalDurationMS == 0` **THEN** 세션 시간 세그먼트를 생략해야 한다.

**REQ-V3-TIME-005** (Ubiquitous):
시스템은 **항상** 세션 비용(cost/$) 세그먼트를 렌더링하지 않아야 한다. cost 표시는 v3에서 제거되었다.

**REQ-V3-TIME-006** (State-Driven):
**IF** mode="compact" **THEN** 세션 시간 세그먼트를 생략해야 한다. default/full 모드에서만 표시.

### GAP-5: Git Ahead/Behind 렌더링

> 이미 수집된 데이터의 시각적 노출

**REQ-V3-GIT-001** (Event-Driven):
**WHEN** `GitStatusData.Ahead > 0` **THEN** 브랜치명 뒤에 `up_arrow_N` 형식으로 표시해야 한다 (예: `main up_arrow_3`).

**REQ-V3-GIT-002** (Event-Driven):
**WHEN** `GitStatusData.Behind > 0` **THEN** 브랜치명 뒤에 `down_arrow_N` 형식으로 표시해야 한다 (예: `main down_arrow_2`).

**REQ-V3-GIT-003** (State-Driven):
**IF** Ahead와 Behind 모두 존재 **THEN** `branch up_arrow_N down_arrow_M` 형식으로 표시해야 한다.

**REQ-V3-GIT-004** (State-Driven):
**IF** Ahead==0 **AND** Behind==0 **THEN** 화살표를 표시하지 않고 브랜치명만 표시해야 한다.

> 참고: 위 요구사항에서 `up_arrow`는 유니코드 화살표 문자 U+2191 (&#8593;), `down_arrow`는 U+2193 (&#8595;)을 의미한다.

### GAP-6: API 사용량 모니터링 (API Usage Monitoring)

> 5시간/7일 사용량 바, 파일 캐시, 배터리 아이콘

**REQ-V3-API-001** (Ubiquitous):
시스템은 **항상** `UsageProvider` 인터페이스를 정의해야 한다.

```go
type UsageProvider interface {
    CollectUsage(ctx context.Context) (*UsageData, error)
}
```

**REQ-V3-API-002** (Ubiquitous):
시스템은 **항상** 사용량 데이터를 `~/.moai/cache/usage.json`에 파일 캐시해야 한다. TTL은 5분.

**REQ-V3-API-003** (Ubiquitous):
시스템은 **항상** 사용량 수집 타임아웃을 300ms로 제한해야 한다.

**REQ-V3-API-004** (Event-Driven):
**WHEN** 캐시가 유효하면 (TTL 내) **THEN** 네트워크 요청 없이 캐시된 데이터를 반환해야 한다.

**REQ-V3-API-005** (Event-Driven):
**WHEN** 캐시가 만료되었으면 **THEN** API에서 사용량을 조회하고 캐시를 갱신해야 한다.

**REQ-V3-API-006** (State-Driven):
**IF** 사용량 비율이 70% 이하 **THEN** 배터리 아이콘을 표시해야 한다.

**REQ-V3-API-007** (State-Driven):
**IF** 사용량 비율이 70% 초과 **THEN** 저배터리 아이콘을 표시해야 한다.

> 참고: 배터리 아이콘은 구현 시 적절한 유니코드 문자 또는 ASCII fallback을 사용한다.

**REQ-V3-API-008** (Unwanted):
시스템은 API 토큰/인증 정보를 로그에 출력**하지 않아야 한다**.

**REQ-V3-API-009** (State-Driven):
**IF** API 조회 실패 **THEN** 사용량 바를 생략하고 나머지 세그먼트는 정상 렌더링해야 한다. 오류를 상위로 전파하지 않는다.

**REQ-V3-API-010** (Ubiquitous):
시스템은 **항상** OAuth 토큰을 다음 순서로 조회해야 한다:
1. macOS Keychain (`security find-generic-password`)
2. `~/.claude/credentials.json` fallback

**REQ-V3-API-011** (State-Driven):
**IF** mode="compact" **THEN** 5H/7D 사용량 바를 생략해야 한다. default 모드에서는 CW와 함께 한줄에 10블록으로, full 모드에서는 각각 독립 줄에 40블록으로 표시.

---

## 4. 모드 레이아웃 (Mode Layouts)

### full 모드 (6줄) — 40블록 바, 각 바 독립 줄

```
🤖 Opus 4.6 │ 🔅 Claude v2.1.50 │ 🗿 MoAI v2.8.0 │ ⏳ 2h 34m
CW: 🪫 ████████████████████████████████████░░░░ 88%
5H: 🔋 ██████████████████░░░░░░░░░░░░░░░░░░░░░░ 45%
7D: 🪫 ████████████████████████████████░░░░░░░░ 82%
📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
💬 MoAI │ 📋 run SPEC-SLV3-001 improve
```

- L1: 모델명, Claude 버전, MoAI 버전, 세션 시간(⏳)
- L2: 컨텍스트 윈도우 사용량 바 (40블록, `CW:` 라벨)
- L3: 5시간 API 사용량 바 (40블록, `5H:` 라벨)
- L4: 7일 API 사용량 바 (40블록, `7D:` 라벨)
- L5: 디렉터리명, 브랜치 + ahead/behind(↑N↓M), git 변경 사항
- L6: 출력 스타일, 현재 태스크

### default 모드 (4줄) — 10블록 바, 3개 바 한줄 통합

```
🤖 Opus 4.6 │ 🔅 v2.1.50 │ 🗿 v2.8.0 │ ⏳ 2h 34m
CW: 🪫 ██████████ 88% │ 5H: 🔋 ██████████ 45% │ 7D: 🪫 ██████████ 82%
📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
💬 MoAI │ 📋 run SPEC-SLV3-001 improve
```

- L1: 모델명, Claude 버전, MoAI 버전, 세션 시간(⏳)
- L2: CW/5H/7D 바 3개 한줄 통합 (각 10블록)
- L3: 디렉터리명, 브랜치 + ahead/behind(↑N↓M), git 변경 사항
- L4: 출력 스타일, 현재 태스크

### compact 모드 (2줄) — CW만, 5H/7D/버전/스타일/태스크 생략

```
🤖 Opus 4.6 │ CW: 🪫 ██████████ 88% │ ⏳ 45m
🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
```

- L1: 모델명, CW 바(10블록), 세션 시간(⏳)
- L2: 브랜치명 + ahead/behind(↑N↓M), git 변경 사항

### 아이콘 규칙

| 사용률 | 아이콘 | 적용 대상 |
|--------|--------|-----------|
| 0-70% | 🔋 | CW, 5H, 7D 동일 |
| 71-100% | 🪫 | CW, 5H, 7D 동일 |

### 빈 줄 규칙

- REQ-V3-LAYOUT-004에 따라, 해당 줄의 모든 세그먼트가 빈 문자열이면 줄 자체를 생략
- 예: full 모드에서 API 토큰 없으면 L3/L4(5H/7D) 생략 → 실제 출력은 4줄
- 예: 태스크 없으면 L6(full)/L4(default) 생략

---

## 5. 기술 설계 (Technical Design)

### 5.1 아키텍처 결정

#### AD-1: RGB 그라디언트 알고리즘

기존 `BarGradient(percentage float64) lipgloss.Color`는 4단계 이산 색상을 반환한다. v3에서는 블록별 연속 보간으로 교체한다.

**알고리즘**:
```
// 블록 i (0 <= i < filled)에 대해:
// blockPct = i / (filled - 1)  (0.0 ~ 1.0)
//
// 0.0-0.5: Green(0,255,0) -> Yellow(255,255,0)
//   R = lerp(0, 255, blockPct * 2)
//   G = 255
//   B = 0
//
// 0.5-1.0: Yellow(255,255,0) -> Red(255,0,0)
//   R = 255
//   G = lerp(255, 0, (blockPct - 0.5) * 2)
//   B = 0
//
// 출력: ANSI 24-bit  \033[38;2;R;G;Bm
```

**결정 근거**: 그라디언트는 테마 독립적이다. 녹-황-적은 보편적 의미(안전-경고-위험)를 가지므로 모든 테마에서 동일하게 적용한다.

#### AD-2: UsageProvider 파일 캐시

- 캐시 경로: `~/.moai/cache/usage.json`
- TTL: 5분 (300초)
- 교차 프로세스 안전: `os.WriteFile` + atomic rename 패턴
- 구조:

```go
type UsageCacheFile struct {
    CachedAt    time.Time  `json:"cached_at"`
    Usage5H     *UsageData `json:"usage_5h"`
    Usage7D     *UsageData `json:"usage_7d"`
}

type UsageData struct {
    UsedTokens  int64   `json:"used_tokens"`
    LimitTokens int64   `json:"limit_tokens"`
    Percentage  float64 `json:"percentage"`  // 0-100
}
```

#### AD-3: OAuth 토큰 조회

1차 시도: macOS Keychain
```bash
security find-generic-password -s "claude.ai" -w
```

2차 시도 (fallback): `~/.claude/credentials.json` 파일에서 토큰 읽기

**보안**: 토큰은 메모리에만 보유, 로그/stderr에 절대 출력하지 않음 (REQ-V3-API-008).

#### AD-4: 모드 마이그레이션 (Lazy Migration)

기존 설정에서 `minimal`/`verbose` 값이 들어오면 런타임에 `compact`/`full`로 매핑한다. 설정 파일 자체를 수정하지 않는다 (lazy migration). 이를 통해 v2 -> v3 전환 시 사용자 설정 파일이 깨지지 않는다.

```go
func NormalizeMode(mode StatuslineMode) StatuslineMode {
    switch mode {
    case "minimal":
        return ModeCompact
    case "verbose":
        return ModeFull
    default:
        return mode
    }
}
```

### 5.2 인터페이스 계약 (Interface Contracts)

#### UsageProvider 인터페이스

```go
// UsageProvider 는 API 사용량 데이터를 수집한다.
type UsageProvider interface {
    // CollectUsage 는 5H/7D 사용량을 반환한다.
    // 캐시 유효 시 캐시 반환, 만료 시 API 조회.
    // 실패 시 nil, nil 반환 (graceful degradation).
    CollectUsage(ctx context.Context) (*UsageResult, error)
}

type UsageResult struct {
    Usage5H  *UsageData // 5시간 사용량
    Usage7D  *UsageData // 7일 사용량
}
```

#### MetricsData 확장

```go
type MetricsData struct {
    Model            string
    CostUSD          float64
    SessionDurationMS int     // NEW: CostData.TotalDurationMS 에서 추출
    Available        bool
}
```

#### 새 세그먼트 상수

```go
const (
    SegmentSessionTime = "session_time"
    SegmentUsage5H     = "usage_5h"
    SegmentUsage7D     = "usage_7d"
    SegmentTask        = "task"
)
```

### 5.3 그라디언트 구현 세부

**ANSI 24-bit 색상**: `lipgloss.Color` 대신 `lipgloss.AdaptiveColor` 또는 직접 ANSI escape 시퀀스를 사용한다.

lipgloss는 hex color를 지원하므로 RGB 값을 hex로 변환하여 사용:

```go
func rgbToHex(r, g, b int) string {
    return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
```

각 블록에 개별 `lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render("block_char")` 적용.

---

## 6. 파일 영향 분석 (File Impact Matrix)

### 새 파일 (4개)

| 파일 | 설명 |
|------|------|
| `internal/statusline/usage.go` | `UsageProvider` 인터페이스, `usageCollector` 구현, 파일 캐시 |
| `internal/statusline/usage_test.go` | 사용량 수집/캐시 테스트 |
| `internal/statusline/gradient.go` | RGB 연속 보간 그라디언트 함수 |
| `internal/statusline/gradient_test.go` | 그라디언트 계산 테스트 |

### 수정 파일 (9개)

| 파일 | 변경 내용 |
|------|-----------|
| `internal/statusline/types.go` | `ModeCompact`/`ModeFull` 상수, `MetricsData.SessionDurationMS`, `UsageData` 타입, 새 세그먼트 상수 |
| `internal/statusline/renderer.go` | 모드별 레이아웃 (2L/2L/5L), `buildGradientBar()`, `renderSessionTime()`, `renderGitBranch()` 수정 (ahead/behind 포함), `renderUsageBar()` |
| `internal/statusline/renderer_test.go` | 모드별 레이아웃 테스트, 그라디언트 바 테스트, 세션 시간 테스트, ahead/behind 테스트 |
| `internal/statusline/builder.go` | `UsageProvider` 통합, `MetricsData.SessionDurationMS` 추출, `NormalizeMode()` 호출 |
| `internal/statusline/builder_test.go` | 빌더 통합 테스트 추가 |
| `internal/statusline/theme.go` | `BarGradient()` -> deprecated, 새 `ContinuousGradient()` 메서드 추가 (또는 gradient.go에 독립 함수) |
| `internal/statusline/git.go` | 변경 없음 (이미 Ahead/Behind 수집 중) |
| `internal/statusline/config.go` (있는 경우) | 새 세그먼트 키 등록, preset 매핑 업데이트 |
| `internal/statusline/update.go` (있는 경우) | TTL 캐싱 패턴 참고 (usage.go에서 동일 패턴 적용) |

---

## 7. 구현 순서 (Implementation Order)

> TDD 접근법: RED -> GREEN -> REFACTOR

### Phase 1: 모드 이름 변경 (Priority: High)

- `types.go`: `ModeCompact`, `ModeFull` 상수 추가, `ModeMinimal`/`ModeVerbose` deprecated alias
- `NormalizeMode()` 함수 구현
- `builder.go`: `SetMode()` 에서 `NormalizeMode()` 호출
- 테스트: 하위 호환성 검증

### Phase 2: 세션 시간 + Git Ahead/Behind (Priority: High)

- `types.go`: `MetricsData.SessionDurationMS`, `SegmentSessionTime` 추가
- `builder.go`: `CostData.TotalDurationMS` -> `MetricsData.SessionDurationMS` 추출
- `renderer.go`: `renderSessionTime()`, `renderGitBranch()` 수정 (화살표 N/화살표 M)
- 테스트: 형식 검증, 0값 생략 검증

### Phase 3: RGB 그라디언트 바 (Priority: High)

- `gradient.go`: `ContinuousGradientBar()` 함수 구현
- 블록별 RGB 보간, hex 변환, lipgloss 스타일 적용
- `renderer.go`: `buildBar()` -> `buildGradientBar()` 전환
- 테스트: 색상 보간 정확성, noColor fallback, 너비 변환

### Phase 4: 모드별 레이아웃 (Priority: High)

- `renderer.go`: `Render()` 리팩토링 - compact(2L), default(2L), full(5L)
- 빈 줄 생략 로직
- 세그먼트 배치 구현
- 테스트: 모드별 줄 수 검증, 빈 줄 생략 검증

### Phase 5: UsageProvider (Priority: Medium)

- `usage.go`: `UsageProvider` 인터페이스, `usageCollector` 구현
- 파일 캐시 (5분 TTL), OAuth 토큰 조회
- `builder.go`: 병렬 provider에 `UsageProvider` 통합
- 테스트: 캐시 TTL, 타임아웃, 인증 실패 graceful degradation

### Phase 6: 통합 테스트 + 리팩토링 (Priority: Medium)

- 전체 모드별 E2E 렌더링 테스트
- 성능 벤치마크 (500ms SLA)
- 코드 정리, MX 태그 추가

---

## 8. 마이그레이션 전략 (Migration Strategy)

### v2 -> v3 모드 매핑

| v2 값 | v3 값 | 처리 방식 |
|-------|-------|-----------|
| `"minimal"` | `"compact"` | `NormalizeMode()` 런타임 매핑 |
| `"default"` | `"default"` | 변경 없음 |
| `"verbose"` | `"full"` | `NormalizeMode()` 런타임 매핑 |

### 마이그레이션 원칙

1. **Lazy Migration**: 설정 파일을 자동 수정하지 않음. 런타임에서만 매핑
2. **하위 호환성**: 기존 `ModeMinimal`, `ModeVerbose` 상수를 deprecated alias로 유지
3. **점진적 전환**: 사용자가 설정을 업데이트할 때 새 이름 사용 유도 (위저드, 문서)
4. **Zero-downtime**: v2 설정으로 v3 바이너리가 정상 동작

### 기존 API 호환성

- `Builder.SetMode(mode StatuslineMode)` 시그니처 변경 없음
- `Renderer.Render(data *StatusData, mode StatuslineMode)` 시그니처 변경 없음
- 새 세그먼트 키는 기존 `segmentConfig`와 호환 (미등록 키는 기본 표시)

---

## 9. 인수 조건 (Acceptance Criteria)

### AC-V3-01: 모드 이름 하위 호환성
- **Given** config에 `mode: "minimal"`이 설정되어 있을 때
- **When** statusline을 렌더링하면
- **Then** compact 모드(2줄)로 정상 렌더링되어야 한다

### AC-V3-02: 모드 이름 하위 호환성 (verbose)
- **Given** config에 `mode: "verbose"`이 설정되어 있을 때
- **When** statusline을 렌더링하면
- **Then** full 모드(6줄)로 정상 렌더링되어야 한다

### AC-V3-03: compact 레이아웃 줄 수
- **Given** mode가 "compact"일 때
- **When** 모든 데이터가 유효하면
- **Then** 정확히 2줄로 렌더링되어야 한다

### AC-V3-04: default 레이아웃 줄 수
- **Given** mode가 "default"일 때
- **When** 모든 데이터가 유효하면
- **Then** 정확히 4줄로 렌더링되어야 한다

### AC-V3-05: full 레이아웃 줄 수
- **Given** mode가 "full"일 때
- **When** 모든 데이터가 유효하면
- **Then** 정확히 6줄로 렌더링되어야 한다

### AC-V3-06: 빈 줄 생략 (full 모드 API 미사용)
- **Given** mode가 "full"이고 UsageProvider가 실패했을 때
- **When** statusline을 렌더링하면
- **Then** L3/L4(5H/7D 바)가 생략되어 4줄로 렌더링되어야 한다

### AC-V3-06b: 빈 줄 생략 (default 모드 API 미사용)
- **Given** mode가 "default"이고 UsageProvider가 실패했을 때
- **When** statusline을 렌더링하면
- **Then** L2에서 5H/7D 부분이 생략되고 CW 바만 표시되어야 한다

### AC-V3-07: RGB 그라디언트 바 블록 수
- **Given** 컨텍스트 사용률이 60%일 때
- **When** full 모드에서 바를 렌더링하면
- **Then** 40블록 중 24블록이 채워지고 각 블록이 개별 RGB 색상을 가져야 한다

### AC-V3-08: 세션 시간 형식
- **Given** TotalDurationMS가 4980000 (83분)일 때
- **When** 세션 시간을 렌더링하면
- **Then** "⏳ 1h 23m"으로 표시되어야 한다 (모래시계 아이콘 사용)

### AC-V3-08b: cost 미표시
- **Given** CostData.TotalCostUSD가 존재할 때
- **When** 어떤 모드에서든 렌더링하면
- **Then** cost($) 세그먼트가 출력에 포함되지 않아야 한다

### AC-V3-09: Git Ahead/Behind 표시
- **Given** Ahead=3, Behind=2일 때
- **When** 브랜치를 렌더링하면
- **Then** "main (up_arrow)3 (down_arrow)2" 형식으로 표시되어야 한다

### AC-V3-10: 사용량 캐시 TTL
- **Given** 캐시가 4분 전에 생성되었을 때
- **When** CollectUsage를 호출하면
- **Then** API 호출 없이 캐시 데이터를 반환해야 한다

### AC-V3-11: 사용량 조회 타임아웃
- **Given** API가 500ms 이상 응답하지 않을 때
- **When** CollectUsage를 호출하면
- **Then** 300ms 이내에 nil을 반환하고 에러를 전파하지 않아야 한다

### AC-V3-12: noColor 모드 바
- **Given** NO_COLOR=1이 설정되어 있을 때
- **When** 프로그레스 바를 렌더링하면
- **Then** ANSI escape 시퀀스 없이 블록 문자만 출력해야 한다

### AC-V3-13: 배터리 아이콘 전환
- **Given** 5H 사용률이 75%일 때
- **When** 사용량을 렌더링하면
- **Then** 저배터리 아이콘이 표시되어야 한다

---

## 10. 엣지 케이스 (Edge Cases)

### EC-01: TotalDurationMS가 0
세션 시작 직후 TotalDurationMS=0이면 세션 시간 세그먼트를 생략한다.

### EC-02: 매우 긴 세션 (24시간+)
TotalDurationMS가 86,400,000 이상이면 `Xd Yh` 형식으로 전환한다 (예: `1d 2h`).

### EC-03: Keychain 접근 불가
macOS Keychain 접근이 실패하면 (권한 없음, Linux 등) credentials.json fallback으로 전환한다. 둘 다 실패하면 사용량 바 생략.

### EC-04: usage.json 파일 손상
캐시 파일 JSON 파싱 실패 시 파일 삭제 후 API 재조회. 재조회도 실패하면 사용량 바 생략.

### EC-05: 동시 프로세스 캐시 접근
여러 statusline 프로세스가 동시에 캐시를 읽고 쓸 수 있다. atomic write (temp file + rename) 패턴으로 corruption 방지.

### EC-06: Ahead/Behind 모두 0
화살표 없이 브랜치명만 표시. 기존 동작과 동일.

### EC-07: 컨텍스트 사용률 100%
40블록 전체가 빨간색(255,0,0)으로 채워진다. 빈 블록 없음.

### EC-08: 컨텍스트 사용률 0%
채워진 블록 없이 40블록 전체가 빈 블록으로 표시.

### EC-09: 터미널 너비 < 40
compact 모드 기본값(10블록)으로 fallback. full 모드의 40블록 바는 터미널 너비에 맞게 축소.

### EC-10: API Rate Limit 도달
API가 429 응답을 반환하면 마지막 캐시 데이터를 사용하고 (TTL 무시), 다음 주기에 재시도.

---

## 11. 리스크 (Risks)

### R1: API 사용량 엔드포인트 가용성
- **설명**: Anthropic API 사용량 엔드포인트가 공개되지 않았거나 변경될 수 있음
- **영향**: 높음 - GAP-6 전체 구현 불가
- **완화**: 사용량 기능을 플러그인화하여 엔드포인트 없이도 나머지 기능 독립 동작. Phase 5를 마지막에 구현하여 나머지 기능에 영향 없음

### R2: macOS Keychain 접근 지연
- **설명**: `security` 명령 실행이 느릴 수 있음
- **영향**: 중간 - statusline 렌더링 지연
- **완화**: 300ms 타임아웃 적용, 캐시 우선 사용

### R3: 멀티라인 레이아웃 iTerm2 호환성
- **설명**: Claude Code statusline이 멀티라인을 어떻게 처리하는지 확인 필요
- **영향**: 높음 - 레이아웃 렌더링 깨짐 가능
- **완화**: 기존 verbose 모드(3줄)가 동작하므로 5줄도 동일 메커니즘 사용. 조기에 iTerm2/Terminal.app에서 검증

### R4: 그라디언트 바 성능
- **설명**: 40블록 x 개별 lipgloss 스타일 적용은 기존 12블록 대비 약 3배 연산
- **영향**: 낮음 - 마이크로초 단위 차이
- **완화**: 벤치마크로 확인, 필요 시 pre-render 캐싱

### R5: 하위 호환성 회귀
- **설명**: 모드 이름 변경 시 기존 스크립트가 깨질 수 있음
- **영향**: 중간 - 사용자 경험 저하
- **완화**: `NormalizeMode()` lazy migration으로 기존 값 100% 호환, 기존 상수 deprecated 유지

### R6: credentials.json 형식 변경
- **설명**: Claude Code가 credentials.json 형식을 변경할 수 있음
- **영향**: 중간 - OAuth 토큰 조회 실패
- **완화**: 방어적 JSON 파싱, 필드 누락 시 graceful degradation

### R7: 테마 독립 그라디언트와 기존 테마 충돌
- **설명**: 기존 `BarGradient()`를 사용하는 코드와의 충돌
- **영향**: 낮음 - 테마별 그라디언트가 무시됨
- **완화**: 기존 `BarGradient()`를 deprecated 처리하고 호출부를 점진적으로 마이그레이션. 새 `ContinuousGradient()` 또는 `gradient.go` 독립 함수 사용

---

## 12. 비기능 요구사항 (Non-functional Requirements)

### NF-001: 전체 statusline 렌더링 시간
statusline 전체 렌더링(데이터 수집 + 렌더링)은 **500ms** 이내에 완료되어야 한다.

### NF-002: 사용량 수집 타임아웃
UsageProvider의 API 호출은 **300ms** 이내에 완료되어야 한다. 초과 시 타임아웃 처리.

### NF-003: 캐시 파일 크기
`~/.moai/cache/usage.json`은 **10KB** 이하여야 한다.

### NF-004: 메모리 사용량
statusline 프로세스의 추가 메모리 사용량은 **5MB** 이하여야 한다.

### NF-005: 하위 호환성
v2 설정 파일로 v3 바이너리가 100% 정상 동작해야 한다. 기능 저하 없음.

### NF-006: Graceful Degradation
개별 Provider(Usage, Git, Update) 실패 시 해당 세그먼트만 생략하고 나머지는 정상 표시.

### NF-007: NO_COLOR 준수
`NO_COLOR` 환경변수 또는 `noColor=true` 설정 시 모든 ANSI escape 시퀀스를 제거해야 한다.

### NF-008: 교차 프로세스 안전
캐시 파일 읽기/쓰기는 atomic write 패턴을 사용하여 교차 프로세스 corruption을 방지해야 한다.

### NF-009: 테스트 커버리지
새로 추가되는 모든 파일은 85% 이상의 테스트 커버리지를 유지해야 한다.

### NF-010: 로깅
API 토큰, credentials 경로 등 민감 정보는 로그에 포함되지 않아야 한다. 오류 로그는 stderr로 출력하되 사용자에게 표시되는 statusline 출력에는 포함하지 않는다.

---

## 추적성 (Traceability)

| 요구사항 | 인수 조건 | 구현 Phase |
|----------|----------|------------|
| REQ-V3-MODE-001~004 | AC-V3-01, AC-V3-02 | Phase 1 |
| REQ-V3-LAYOUT-001~004 | AC-V3-03~06, AC-V3-06b | Phase 4 |
| REQ-V3-BAR-001~004 | AC-V3-07, AC-V3-12 | Phase 3 |
| REQ-V3-TIME-001~006 | AC-V3-08, AC-V3-08b | Phase 2 |
| REQ-V3-GIT-001~004 | AC-V3-09 | Phase 2 |
| REQ-V3-API-001~011 | AC-V3-10, AC-V3-11, AC-V3-13 | Phase 5 |

### 변경 이력

| 버전 | 날짜 | 변경 내용 |
|------|------|-----------|
| 1.0.0 | 2026-03-05 | 초기 SPEC 작성 |
| 1.1.0 | 2026-03-06 | 레이아웃 재설계: full 5줄→6줄, default 2줄→4줄. cost 제거. 시간 아이콘 ⏱→⏳. CW/5H/7D 라벨 추가. |
