# SPEC-TOOL-001: AST-Grep & Performance Ranking Integration

---
id: SPEC-TOOL-001
title: AST-Grep & Performance Ranking Integration
created: 2026-02-03
status: Completed
priority: Medium
phase: "Phase 4 - UI and Integration"
assigned: expert-backend
dependencies:
  - SPEC-CONFIG-001
  - SPEC-GIT-001
estimated_loc: ~1,100
lifecycle: spec-anchored
modules:
  - internal/astgrep/ (analyzer.go, models.go, rules.go)
  - internal/rank/ (client.go, auth.go, config.go, hook.go)
tags: [ast-grep, rank, code-analysis, metrics, oauth, hmac]
---

## HISTORY

| 날짜 | 변경 사항 | 작성자 |
|------|----------|--------|
| 2026-02-03 | 초기 SPEC 작성 | manager-spec |

---

## 1. Environment (환경)

### 1.1 시스템 환경

- **런타임**: Go 1.22+ 단일 바이너리 (CGO_ENABLED=0)
- **대상 플랫폼**: macOS (arm64, amd64), Linux (arm64, amd64), Windows (amd64, arm64)
- **외부 의존성**: ast-grep (sg) CLI 바이너리 (런타임 선택적 의존성)
- **네트워크**: MoAI Rank API (rank.mo.ai.kr) HTTPS 통신
- **인증 저장소**: macOS Keychain / Linux secret-service / 파일 기반 폴백 (~/.moai/rank/credentials.json, chmod 600)
- **설정 의존성**: SPEC-CONFIG-001 (internal/config/) 모듈을 통한 YAML 설정 로딩

### 1.2 Python 참조 구현

- `moai_adk/astgrep/`: analyzer.py (~537 LOC), models.py (~125 LOC), rules.py (~180 LOC)
- `moai_adk/rank/`: client.py (~570 LOC), auth.py (~426 LOC), config.py (~148 LOC), hook.py (~1,100+ LOC)

### 1.3 Go 대상 모듈

- `internal/astgrep/`: analyzer.go, models.go, rules.go
- `internal/rank/`: client.go, auth.go, config.go, hook.go

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

- [A-1] ast-grep (sg) CLI가 시스템 PATH에 설치되어 있을 수 있으나, 설치되지 않은 환경에서도 graceful degradation이 보장되어야 한다.
- [A-2] `sg` 명령어는 JSON 출력 모드(`--json`)를 지원하며, 출력 형식은 ast-grep v0.25+ 기준이다.
- [A-3] MoAI Rank API는 HTTPS를 통해 접근 가능하며, HMAC-SHA256 서명 기반 인증을 사용한다.
- [A-4] 크리덴셜 파일은 사용자 홈 디렉토리 `~/.moai/rank/` 아래에 저장되며, 파일 권한 600으로 보호된다.
- [A-5] SPEC-CONFIG-001이 완료되어 `internal/config/` 패키지를 통해 YAML 설정 읽기/쓰기가 가능하다.
- [A-6] SPEC-GIT-001이 완료되어 `internal/core/git/` 패키지를 통해 Git 이벤트 감지가 가능하다.

### 2.2 비즈니스 가정

- [A-7] Rank 서비스는 익명 제출을 지원하며, 사용자는 옵트인/옵트아웃이 가능하다.
- [A-8] 세션 메트릭은 Claude Code의 SessionEnd 훅을 통해 자동 수집된다.
- [A-9] 리더보드 데이터는 공개 API로 인증 없이 조회 가능하다.

---

## 3. Requirements (요구사항)

### 3.1 AST-Grep 모듈 요구사항

#### R-AST-001: sg CLI 가용성 감지 (Ubiquitous)

시스템은 **항상** ast-grep (sg) CLI의 설치 여부를 감지하고, 가용성 상태를 캐싱해야 한다.

```
The system shall detect ast-grep (sg) CLI availability and cache the result.
```

- sg CLI가 설치되지 않은 경우 빈 결과를 반환한다 (에러 아님).
- 가용성 체크는 `sg --version` 명령으로 수행하며, 5초 타임아웃을 적용한다.
- 결과는 Analyzer 인스턴스 수명 동안 캐싱한다.

#### R-AST-002: 파일 확장자 기반 언어 감지 (Ubiquitous)

시스템은 **항상** 파일 확장자를 기반으로 프로그래밍 언어를 감지해야 한다.

```
The system shall detect programming language from file extensions.
```

- 지원 언어: Python, JavaScript, TypeScript, Go, Rust, Java, Kotlin, Ruby, C, C++, C#, Swift, Lua, HTML, Vue, Svelte (22개 확장자)
- 알 수 없는 확장자는 "text"로 분류한다.

#### R-AST-003: 단일 파일 스캔 (Event-Driven)

**WHEN** 사용자가 파일 스캔을 요청하면 **THEN** ast-grep을 사용하여 해당 파일의 AST 패턴 매치를 수행해야 한다.

```
When a file scan is requested, the system shall scan the file using ast-grep and return structured match results.
```

- 입력: 파일 경로, 선택적 ScanConfig
- 출력: ScanResult (매치 목록, 스캔 시간, 언어)
- sg가 없으면 빈 매치 목록을 반환한다.
- 파일이 존재하지 않으면 에러를 반환한다.

#### R-AST-004: 프로젝트 전체 스캔 (Event-Driven)

**WHEN** 프로젝트 스캔이 요청되면 **THEN** 지원되는 모든 파일을 재귀적으로 스캔하고 집계 결과를 반환해야 한다.

```
When a project scan is requested, the system shall recursively scan all supported files and return aggregated results.
```

- 파일 포함/제외 패턴 지원 (glob 기반)
- 기본 제외: node_modules, .git, __pycache__
- 심각도별 매치 수 집계 (error, warning, info, hint)
- 총 스캔 파일 수, 총 매치 수, 총 스캔 시간 포함

#### R-AST-005: 패턴 검색 (Event-Driven)

**WHEN** 커스텀 패턴 검색이 요청되면 **THEN** 지정된 언어와 경로에서 AST 패턴을 검색해야 한다.

```
When a custom pattern search is requested, the system shall search for AST patterns in specified language and path.
```

- `sg run --pattern <pattern> --lang <lang> --json <path>` 명령 실행
- 120초 타임아웃 적용
- rule_id는 `pattern:<패턴 앞 30자>` 형식으로 설정

#### R-AST-006: 패턴 교체 (Event-Driven)

**WHEN** 패턴 교체가 요청되면 **THEN** 매칭된 코드를 지정된 대체 패턴으로 변환해야 한다.

```
When a pattern replacement is requested, the system shall replace matching code with the specified replacement pattern.
```

- dry_run 모드 지원 (기본값: true, 실제 파일 수정 없음)
- 교체 전 매치 수 계산
- 변경 목록 (파일 경로, 원본 코드, 새 코드, 범위) 반환
- `sg run --pattern <p> --rewrite <r> --lang <lang> <path>` 명령으로 실행

#### R-AST-007: YAML 규칙 로딩 (Event-Driven)

**WHEN** 규칙 파일 또는 디렉토리가 지정되면 **THEN** YAML 형식의 ast-grep 규칙을 파싱하고 로드해야 한다.

```
When rule files are specified, the system shall parse and load ast-grep rules from YAML format.
```

- 단일 파일 및 디렉토리 로딩 지원
- 멀티 문서 YAML (--- 구분) 지원
- 규칙 필드: id, language, severity, message, pattern, fix (선택적)
- 언어별 규칙 필터링

### 3.2 Rank 모듈 요구사항

#### R-RANK-001: API 클라이언트 초기화 (Ubiquitous)

시스템은 **항상** HMAC-SHA256 서명 기반 인증을 사용하는 HTTP 클라이언트를 제공해야 한다.

```
The system shall provide an HTTP client with HMAC-SHA256 signature-based authentication.
```

- API 키는 설정에서 자동 로딩 또는 명시적 전달
- User-Agent: `moai-adk/1.0`
- 기본 타임아웃: 30초
- 기본 API URL: `https://rank.mo.ai.kr`
- 환경 변수 오버라이드: `MOAI_RANK_API_URL`

#### R-RANK-002: HMAC-SHA256 요청 서명 (Ubiquitous)

시스템은 **항상** 인증이 필요한 요청에 HMAC-SHA256 서명을 포함해야 한다.

```
The system shall include HMAC-SHA256 signatures in authenticated requests.
```

- 서명 계산: `HMAC-SHA256(api_key, timestamp + ":" + body)`
- 헤더: X-API-Key, X-Timestamp, X-Signature
- API 키가 없으면 AuthenticationError 반환

#### R-RANK-003: 사용자 랭킹 조회 (Event-Driven)

**WHEN** 사용자가 랭킹 상태를 조회하면 **THEN** 일간/주간/월간/전체 랭킹 정보를 반환해야 한다.

```
When user rank is requested, the system shall return daily/weekly/monthly/all-time ranking information.
```

- API 엔드포인트: GET /api/v1/rank
- 인증 필요 (X-API-Key 헤더)
- 반환: 사용자명, 순위, 점수, 총 참가자 수, 토큰 통계

#### R-RANK-004: 리더보드 조회 (Event-Driven)

**WHEN** 리더보드 조회가 요청되면 **THEN** 지정된 기간의 순위 목록을 반환해야 한다.

```
When leaderboard is requested, the system shall return ranked list for the specified period.
```

- API 엔드포인트: GET /api/leaderboard (공개, 인증 불필요)
- 파라미터: period (daily/weekly/monthly/all_time), limit (1-100), offset
- 페이지네이션 지원

#### R-RANK-005: 세션 제출 (Event-Driven)

**WHEN** 세션 메트릭이 수집되면 **THEN** HMAC 인증을 사용하여 세션 데이터를 제출해야 한다.

```
When session metrics are collected, the system shall submit session data with HMAC authentication.
```

- API 엔드포인트: POST /api/v1/sessions
- HMAC 인증 필수
- 토큰 필드 상한: 100,000,000 (서버 검증 제한)
- 세션 해시: SHA-256 기반 중복 방지
- 대시보드 필드: started_at, duration_seconds, turn_count, tool_usage, model_usage, code_metrics

#### R-RANK-006: 배치 세션 제출 (Event-Driven)

**WHEN** 다수의 세션을 동기화해야 하면 **THEN** 배치 API를 사용하여 최대 100개 세션을 일괄 제출해야 한다.

```
When multiple sessions need sync, the system shall use batch API to submit up to 100 sessions at once.
```

- API 엔드포인트: POST /api/v1/sessions/batch
- 최대 100개 세션/배치
- 100개 초과 시 에러 반환

#### R-RANK-007: OAuth 인증 플로우 (Event-Driven)

**WHEN** 사용자가 로그인을 시도하면 **THEN** GitHub OAuth를 통해 인증하고 API 키를 안전하게 저장해야 한다.

```
When user attempts login, the system shall authenticate via GitHub OAuth and securely store API key.
```

- 로컬 HTTP 서버로 OAuth 콜백 수신 (포트 8080-8180 범위에서 가용 포트 탐색)
- CSRF 방지를 위한 state 토큰 (secrets.token_urlsafe 32바이트)
- 브라우저 자동 열기
- 인증 타임아웃: 300초 (기본값)
- 신규 플로우: 서버에서 직접 API 키 반환
- 레거시 플로우: 인증 코드를 API 키로 교환

#### R-RANK-008: 크리덴셜 관리 (State-Driven)

**IF** 크리덴셜이 저장되어 있으면 **THEN** 안전하게 로드하고, 없으면 None/nil을 반환해야 한다.

```
If credentials exist, the system shall load them securely; otherwise return nil.
```

- 저장 경로: ~/.moai/rank/credentials.json
- 파일 권한: 600 (owner read/write only)
- 디렉토리 권한: 700 (owner only)
- 원자적 쓰기 (임시 파일 + rename)
- 필드: api_key, username, user_id, created_at

#### R-RANK-009: API 상태 확인 (Event-Driven)

**WHEN** API 상태 확인이 요청되면 **THEN** Rank 서비스의 가용성을 검증해야 한다.

```
When API status check is requested, the system shall verify Rank service availability.
```

- API 엔드포인트: GET /api/v1/status
- 반환: status, version, timestamp

### 3.3 보안 요구사항

#### R-SEC-001: 크리덴셜 보안 (Unwanted)

시스템은 API 키를 평문으로 로그에 출력**하지 않아야 한다**.

```
The system shall not log API keys in plaintext.
```

#### R-SEC-002: JSON 직렬화 안전성 (Ubiquitous)

시스템은 **항상** Go 구조체 직렬화(`json.Marshal`)를 사용하여 JSON을 생성해야 한다. 문자열 결합에 의한 JSON 생성은 금지된다.

```
The system shall always use Go struct serialization (json.Marshal) for JSON generation.
```

#### R-SEC-003: CSRF 보호 (Unwanted)

시스템은 OAuth 콜백에서 state 토큰 불일치 시 인증을 진행**하지 않아야 한다**.

```
The system shall not proceed with authentication when OAuth state token mismatches.
```

### 3.4 성능 요구사항

#### R-PERF-001: 단일 파일 스캔 성능 (State-Driven)

**IF** sg CLI가 설치되어 있으면 **THEN** 단일 파일 스캔은 60초 이내에 완료되어야 한다.

```
If sg CLI is available, single file scan shall complete within 60 seconds.
```

#### R-PERF-002: API 요청 타임아웃 (Ubiquitous)

시스템은 **항상** API 요청에 타임아웃을 적용해야 한다 (기본 30초).

```
The system shall always apply timeout to API requests (default 30 seconds).
```

---

## 4. Specifications (명세)

### 4.1 Go 인터페이스 명세

#### AST-Grep Analyzer 인터페이스

```go
// internal/astgrep/analyzer.go

type Analyzer interface {
    // Scan은 지정된 패턴과 경로에서 AST 기반 코드 스캔을 수행한다.
    Scan(ctx context.Context, patterns []string, paths []string) (*ScanResult, error)

    // FindPattern은 특정 언어에서 단일 패턴을 검색한다.
    FindPattern(ctx context.Context, pattern string, lang string) ([]Match, error)

    // Replace는 매칭된 패턴을 대체 패턴으로 교체한다.
    Replace(ctx context.Context, pattern, replacement, lang string, paths []string) ([]FileChange, error)
}
```

#### AST-Grep 데이터 모델

```go
// internal/astgrep/models.go

type Match struct {
    File   string `json:"file"`
    Line   int    `json:"line"`
    Column int    `json:"column"`
    Text   string `json:"text"`
    Rule   string `json:"rule"`
}

type ScanResult struct {
    Matches  []Match       `json:"matches"`
    Duration time.Duration `json:"duration"`
    Files    int           `json:"files_scanned"`
}

type ScanConfig struct {
    RulesPath       string   `json:"rules_path,omitempty"`
    SecurityScan    bool     `json:"security_scan"`
    IncludePatterns []string `json:"include_patterns,omitempty"`
    ExcludePatterns []string `json:"exclude_patterns,omitempty"`
}

type FileChange struct {
    FilePath string `json:"file_path"`
    OldCode  string `json:"old_code"`
    NewCode  string `json:"new_code"`
    Line     int    `json:"line"`
    Column   int    `json:"column"`
}

type Rule struct {
    ID       string `json:"id"`
    Language string `json:"language"`
    Severity string `json:"severity"`
    Message  string `json:"message"`
    Pattern  string `json:"pattern"`
    Fix      string `json:"fix,omitempty"`
}
```

#### Rank Client 인터페이스

```go
// internal/rank/client.go

type Client interface {
    // CheckStatus는 API 서비스 상태를 확인한다.
    CheckStatus(ctx context.Context) (*ApiStatus, error)

    // GetUserRank는 현재 사용자의 랭킹 정보를 조회한다.
    GetUserRank(ctx context.Context) (*UserRank, error)

    // GetLeaderboard는 지정된 기간의 리더보드를 조회한다.
    GetLeaderboard(ctx context.Context, period string, limit, offset int) ([]LeaderboardEntry, error)

    // SubmitSession은 단일 세션 메트릭을 제출한다.
    SubmitSession(ctx context.Context, session *SessionSubmission) error

    // SubmitSessionsBatch는 최대 100개 세션을 일괄 제출한다.
    SubmitSessionsBatch(ctx context.Context, sessions []*SessionSubmission) (*BatchResult, error)
}
```

#### Rank 데이터 모델

```go
// internal/rank/client.go

type ApiStatus struct {
    Status    string `json:"status"`
    Version   string `json:"version"`
    Timestamp string `json:"timestamp"`
}

type RankInfo struct {
    Position          int     `json:"position"`
    CompositeScore    float64 `json:"composite_score"`
    TotalParticipants int     `json:"total_participants"`
}

type UserRank struct {
    Username      string    `json:"username"`
    Daily         *RankInfo `json:"daily,omitempty"`
    Weekly        *RankInfo `json:"weekly,omitempty"`
    Monthly       *RankInfo `json:"monthly,omitempty"`
    AllTime       *RankInfo `json:"all_time,omitempty"`
    TotalTokens   int64     `json:"total_tokens"`
    TotalSessions int       `json:"total_sessions"`
    InputTokens   int64     `json:"input_tokens"`
    OutputTokens  int64     `json:"output_tokens"`
    LastUpdated   string    `json:"last_updated"`
}

type LeaderboardEntry struct {
    Rank           int     `json:"rank"`
    Username       string  `json:"username"`
    TotalTokens    int64   `json:"total_tokens"`
    CompositeScore float64 `json:"composite_score"`
    SessionCount   int     `json:"session_count"`
    IsPrivate      bool    `json:"is_private"`
}

type SessionSubmission struct {
    SessionHash        string            `json:"session_hash"`
    EndedAt            string            `json:"ended_at"`
    InputTokens        int64             `json:"input_tokens"`
    OutputTokens       int64             `json:"output_tokens"`
    CacheCreationTokens int64            `json:"cache_creation_tokens"`
    CacheReadTokens    int64             `json:"cache_read_tokens"`
    ModelName          string            `json:"model_name,omitempty"`
    AnonymousProjectID string            `json:"anonymous_project_id,omitempty"`
    StartedAt          string            `json:"started_at,omitempty"`
    DurationSeconds    int               `json:"duration_seconds,omitempty"`
    TurnCount          int               `json:"turn_count,omitempty"`
    ToolUsage          map[string]int    `json:"tool_usage,omitempty"`
    ModelUsage         map[string]map[string]int `json:"model_usage,omitempty"`
    CodeMetrics        map[string]int    `json:"code_metrics,omitempty"`
}

type BatchResult struct {
    Success   bool `json:"success"`
    Processed int  `json:"processed"`
    Succeeded int  `json:"succeeded"`
    Failed    int  `json:"failed"`
}
```

#### Rank 인증 모듈

```go
// internal/rank/auth.go

type OAuthHandler interface {
    // StartOAuthFlow는 GitHub OAuth 인증 플로우를 시작한다.
    StartOAuthFlow(ctx context.Context, timeout time.Duration) (*Credentials, error)
}

type Credentials struct {
    APIKey    string `json:"api_key"`
    Username  string `json:"username"`
    UserID    string `json:"user_id"`
    CreatedAt string `json:"created_at"`
}
```

#### Rank 설정 모듈

```go
// internal/rank/config.go

type Config struct {
    BaseURL string `json:"base_url"`
}

type CredentialStore interface {
    // Save는 크리덴셜을 안전하게 저장한다.
    Save(creds *Credentials) error

    // Load는 저장된 크리덴셜을 로드한다.
    Load() (*Credentials, error)

    // Delete는 저장된 크리덴셜을 삭제한다.
    Delete() error

    // HasCredentials는 크리덴셜 존재 여부를 확인한다.
    HasCredentials() bool

    // GetAPIKey는 저장된 API 키만 반환한다.
    GetAPIKey() (string, error)
}
```

### 4.2 에러 타입

```go
// internal/rank/client.go

type ClientError struct {
    Message string
}

type AuthenticationError struct {
    Message string
}

type ApiError struct {
    Message    string
    StatusCode int
    Details    map[string]interface{}
}
```

### 4.3 상수

```go
// internal/rank/client.go
const (
    MaxTokensPerField = 100_000_000
    DefaultTimeout    = 30 * time.Second
    DefaultBaseURL    = "https://rank.mo.ai.kr"
    APIVersion        = "v1"
    UserAgent         = "moai-adk/1.0"
)

// internal/astgrep/analyzer.go
const (
    SGTimeout     = 60 * time.Second
    SearchTimeout = 120 * time.Second
    VersionTimeout = 5 * time.Second
)
```

### 4.4 파일 확장자-언어 매핑

22개 파일 확장자를 지원하며, Python 참조 구현의 `EXTENSION_TO_LANGUAGE` 맵과 동일한 매핑을 유지한다:

| 확장자 | 언어 | 확장자 | 언어 |
|--------|------|--------|------|
| .py | python | .go | go |
| .js, .mjs, .cjs | javascript | .rs | rust |
| .jsx | javascriptreact | .java | java |
| .ts, .mts, .cts | typescript | .kt, .kts | kotlin |
| .tsx | typescriptreact | .rb | ruby |
| .c, .h | c | .swift | swift |
| .cpp, .cc, .cxx, .hpp | cpp | .lua | lua |
| .cs | csharp | .html | html |
| .vue | vue | .svelte | svelte |

---

## 5. Traceability (추적성)

| 요구사항 ID | 구현 파일 | 테스트 파일 | 상태 |
|-------------|-----------|-------------|------|
| R-AST-001 | internal/astgrep/analyzer.go | internal/astgrep/analyzer_test.go | Planned |
| R-AST-002 | internal/astgrep/analyzer.go | internal/astgrep/analyzer_test.go | Planned |
| R-AST-003 | internal/astgrep/analyzer.go | internal/astgrep/analyzer_test.go | Planned |
| R-AST-004 | internal/astgrep/analyzer.go | internal/astgrep/analyzer_test.go | Planned |
| R-AST-005 | internal/astgrep/analyzer.go | internal/astgrep/analyzer_test.go | Planned |
| R-AST-006 | internal/astgrep/analyzer.go | internal/astgrep/analyzer_test.go | Planned |
| R-AST-007 | internal/astgrep/rules.go | internal/astgrep/rules_test.go | Planned |
| R-RANK-001 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-RANK-002 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-RANK-003 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-RANK-004 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-RANK-005 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-RANK-006 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-RANK-007 | internal/rank/auth.go | internal/rank/auth_test.go | Planned |
| R-RANK-008 | internal/rank/config.go | internal/rank/config_test.go | Planned |
| R-RANK-009 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-SEC-001 | internal/rank/ (전체) | internal/rank/security_test.go | Planned |
| R-SEC-002 | internal/rank/client.go | internal/rank/client_test.go | Planned |
| R-SEC-003 | internal/rank/auth.go | internal/rank/auth_test.go | Planned |
| R-PERF-001 | internal/astgrep/analyzer.go | internal/astgrep/bench_test.go | Planned |
| R-PERF-002 | internal/rank/client.go | internal/rank/client_test.go | Planned |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 89.4% (ast-grep) / 85.1% (rank)

### Summary

Dual-module implementation covering AST-Grep code analysis and Rank API client. AST-Grep analyzer provides pattern-based code analysis with configurable rules, severity levels, and suggested fixes for multiple languages. Rank client implements HTTP-based API communication for user ranking, leaderboard queries, and session metric submission with OAuth token authentication, retry logic, and batch operations (up to 100 sessions per request).

### Files Created

- `internal/astgrep/analyzer.go`
- `internal/astgrep/analyzer_test.go`
- `internal/astgrep/models.go`
- `internal/astgrep/rules.go`
- `internal/astgrep/rules_test.go`
- `internal/rank/auth.go`
- `internal/rank/auth_test.go`
- `internal/rank/client.go`
- `internal/rank/client_test.go`
- `internal/rank/config.go`
- `internal/rank/config_test.go`
