# SPEC-TOOL-001: 구현 계획

---
spec_id: SPEC-TOOL-001
title: AST-Grep & Performance Ranking Integration - Implementation Plan
created: 2026-02-03
tags: [ast-grep, rank, implementation]
---

## 1. 구현 전략 개요

### 접근 방식

Python 참조 구현을 기반으로 Go 관용적 패턴으로 재작성한다. 두 모듈은 독립적이므로 병렬 구현이 가능하지만, AST-Grep 모듈이 상대적으로 단순하므로 먼저 구현하여 빠른 성과를 확보한 후 Rank 모듈로 진행한다.

### 핵심 설계 원칙

1. **인터페이스 우선**: 모든 외부 의존성(sg CLI, HTTP API)은 Go 인터페이스 뒤에 래핑하여 테스트 용이성 확보
2. **Context 전파**: 모든 장시간 작업에 `context.Context` 파라미터 전달 (취소, 타임아웃)
3. **Graceful Degradation**: sg CLI 미설치 시 에러 없이 빈 결과 반환
4. **구조체 직렬화**: 모든 JSON은 `json.Marshal()`로 생성 (문자열 결합 금지)
5. **에러 래핑**: `fmt.Errorf("context: %w", err)` 패턴으로 에러 체이닝

---

## 2. 마일스톤

### Primary Goal: AST-Grep 모듈 구현

**범위**: `internal/astgrep/` 패키지 전체 (analyzer.go, models.go, rules.go)

#### Task 1: 데이터 모델 정의 (models.go)

- Match, ScanResult, ScanConfig, FileChange, ProjectScanResult, ReplaceResult 구조체 정의
- JSON 태그 부착
- 기본값 설정 (ExcludePatterns: node_modules, .git, __pycache__)

#### Task 2: 규칙 관리 (rules.go)

- Rule 구조체 정의
- RuleLoader 구현
  - YAML 파일 로딩 (단일 파일, 디렉토리)
  - 멀티 문서 YAML 파싱 (gopkg.in/yaml.v3)
  - 언어별 규칙 필터링
  - 에러 처리 (파일 없음, 잘못된 YAML)

#### Task 3: Analyzer 구현 (analyzer.go)

- Analyzer 인터페이스 정의
- MoAIASTGrepAnalyzer 구조체 구현
  - `isSGAvailable()`: sg CLI 감지 및 캐싱
  - `detectLanguage()`: 파일 확장자-언어 매핑 (22개 확장자)
  - `shouldIncludeFile()`: 포함/제외 패턴 매칭
  - `scanFile()`: 단일 파일 스캔
  - `scanProject()`: 프로젝트 전체 스캔
  - `patternSearch()`: 커스텀 패턴 검색
  - `patternReplace()`: 패턴 교체 (dry_run 지원)
- sg CLI 실행: `exec.CommandContext` 사용 (타임아웃 지원)
- JSON 출력 파싱

#### Task 4: AST-Grep 테스트

- 단위 테스트: 언어 감지, 패턴 매칭, JSON 파싱
- 통합 테스트: sg CLI 존재 시 실제 스캔 (build tag로 분리)
- 벤치마크 테스트: 파일 스캔 성능
- 테스트 데이터: testdata/ 디렉토리에 샘플 코드 파일 배치

### Secondary Goal: Rank Client 구현

**범위**: `internal/rank/client.go`

#### Task 5: 데이터 모델 및 에러 타입

- ApiStatus, RankInfo, UserRank, LeaderboardEntry 구조체 정의
- SessionSubmission 구조체 (대시보드 필드 포함)
- BatchResult 구조체
- ClientError, AuthenticationError, ApiError 커스텀 에러 타입
- MaxTokensPerField 상수 (100,000,000)

#### Task 6: HMAC-SHA256 서명 모듈

- `computeSignature(apiKey, timestamp, body string) string` 구현
- `crypto/hmac` + `crypto/sha256` 사용
- `getAuthHeaders(apiKey, body string) map[string]string` 헤더 생성

#### Task 7: HTTP 클라이언트 구현

- Client 인터페이스 정의
- RankClient 구조체 구현
  - `makeRequest()`: 공통 HTTP 요청 래퍼 (GET/POST)
  - `CheckStatus()`: GET /api/v1/status
  - `GetUserRank()`: GET /api/v1/rank (인증)
  - `GetLeaderboard()`: GET /api/leaderboard (비인증)
  - `SubmitSession()`: POST /api/v1/sessions (HMAC 인증)
  - `SubmitSessionsBatch()`: POST /api/v1/sessions/batch (HMAC 인증)
  - `ComputeSessionHash()`: SHA-256 세션 해시 생성
- net/http 표준 라이브러리 사용 (외부 HTTP 클라이언트 라이브러리 없음)
- 응답 코드별 에러 처리 (401: AuthenticationError, 400+: ApiError)

#### Task 8: Rank Client 테스트

- 모킹: httptest.NewServer를 사용한 API 모킹
- HMAC 서명 검증 테스트
- 에러 시나리오 테스트 (타임아웃, 인증 실패, 서버 에러)
- 토큰 상한 적용 테스트 (MaxTokensPerField)
- 배치 100개 제한 테스트

### Tertiary Goal: Rank Auth & Config 구현

**범위**: `internal/rank/auth.go`, `internal/rank/config.go`

#### Task 9: 크리덴셜 저장소 (config.go)

- Config 구조체 (BaseURL, APIVersion)
- CredentialStore 인터페이스 및 FileCredentialStore 구현
  - Save: 임시 파일 + rename 원자적 쓰기
  - Load: JSON 파싱
  - Delete: 파일 삭제
  - HasCredentials: 파일 존재 확인
  - GetAPIKey: API 키만 반환
- 파일 권한: os.Chmod (0600, 0700)
- 경로: ~/.moai/rank/credentials.json

#### Task 10: OAuth 핸들러 (auth.go)

- OAuthHandler 구현
  - 로컬 HTTP 서버 (net/http)
  - 가용 포트 탐색 (8080-8180)
  - state 토큰 생성 (crypto/rand)
  - OAuth 콜백 처리 (state 검증, API 키 추출)
  - 브라우저 열기 (os/exec로 플랫폼별 open 명령)
  - 타임아웃 관리 (context.WithTimeout)
- 신규 플로우: 서버 직접 API 키 반환
- 레거시 플로우: 코드-키 교환 (POST /api/auth/cli/token)

#### Task 11: Auth & Config 테스트

- 크리덴셜 CRUD 테스트 (임시 디렉토리 사용)
- 파일 권한 검증 테스트
- OAuth state 검증 테스트
- 포트 탐색 테스트

### Optional Goal: Rank Hook 구현

**범위**: `internal/rank/hook.go`

#### Task 12: 세션 훅 통합

- SessionEnd 이벤트 처리
- JSONL 트랜스크립트 파싱 (토큰 사용량 추출)
- 모델별 비용 계산 (MODEL_PRICING 맵)
- 중복 감지 로직
- 프로젝트별 옵트아웃 설정 지원
- anonymous_project_id 생성 (SHA-256 해시)

---

## 3. 기술적 접근

### 3.1 AST-Grep CLI 래핑 패턴

```go
// exec.CommandContext를 사용하여 타임아웃과 취소를 자연스럽게 지원
func (a *analyzer) runSG(ctx context.Context, args ...string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, "sg", args...)
    cmd.Dir = a.workDir
    output, err := cmd.Output()
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            // sg는 매치가 없을 때도 non-zero exit code를 반환할 수 있음
            return output, nil
        }
        return nil, fmt.Errorf("sg execution failed: %w", err)
    }
    return output, nil
}
```

### 3.2 HMAC-SHA256 인증 패턴

```go
func computeSignature(apiKey, timestamp, body string) string {
    message := timestamp + ":" + body
    mac := hmac.New(sha256.New, []byte(apiKey))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}
```

### 3.3 안전한 크리덴셜 저장 패턴

```go
func (s *fileStore) Save(creds *Credentials) error {
    // 디렉토리 생성 (0700)
    if err := os.MkdirAll(s.dir, 0700); err != nil {
        return fmt.Errorf("create config dir: %w", err)
    }

    // JSON 직렬화
    data, err := json.MarshalIndent(creds, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal credentials: %w", err)
    }

    // 임시 파일에 쓰기 + 권한 설정 + 원자적 rename
    tmpFile := s.credPath + ".tmp"
    if err := os.WriteFile(tmpFile, data, 0600); err != nil {
        return fmt.Errorf("write temp file: %w", err)
    }

    return os.Rename(tmpFile, s.credPath)
}
```

### 3.4 테스트 전략

| 대상 | 패턴 | 도구 |
|------|------|------|
| AST-Grep JSON 파싱 | 테이블 기반 테스트 | testdata/*.json |
| sg CLI 통합 | 빌드 태그 분리 (`//go:build integration`) | 실제 sg CLI |
| Rank API 클라이언트 | HTTP 모킹 | httptest.NewServer |
| HMAC 서명 | 검증 벡터 | 알려진 입출력 쌍 |
| 크리덴셜 저장 | 임시 디렉토리 | t.TempDir() |
| OAuth 플로우 | 모킹 서버 | httptest + 콜백 시뮬레이션 |

---

## 4. 아키텍처 설계 방향

### 4.1 패키지 의존성

```
internal/astgrep/
    depends on: (없음 - 독립 패키지)
    used by: internal/hook/post_tool.go, internal/core/quality/

internal/rank/
    depends on: internal/config/ (SPEC-CONFIG-001)
    used by: internal/hook/session_end.go, internal/cli/rank.go
```

### 4.2 인터페이스 경계

- `Analyzer` 인터페이스: quality 게이트 모듈이 AST 분석에 의존할 때 사용
- `Client` 인터페이스: CLI rank 서브커맨드와 session_end 훅에서 사용
- `CredentialStore` 인터페이스: auth 모듈과 config 모듈의 크리덴셜 접근 추상화
- `OAuthHandler` 인터페이스: CLI rank login 커맨드에서 사용

### 4.3 동시성 모델

- AST-Grep: 프로젝트 스캔 시 파일별 병렬 스캔 (goroutine + sync.WaitGroup)
- Rank: HTTP 요청은 순차적 (서버 부하 방지), 배치 제출은 단일 요청

---

## 5. 리스크 및 대응

| 리스크 | 영향 | 대응 전략 |
|--------|------|----------|
| sg CLI 미설치 환경 | AST 분석 불가 | Graceful degradation - 빈 결과 반환, doctor 명령에서 경고 |
| Rank API 서버 다운 | 세션 제출 실패 | 로컬 큐잉 + 재시도 로직, 오프라인 모드 지원 |
| OAuth 브라우저 열기 실패 | 로그인 불가 | CLI에 인증 URL 출력, 수동 복사/붙여넣기 안내 |
| 크리덴셜 파일 권한 문제 (Windows) | 보안 취약 | Windows에서는 파일 시스템 권한 대신 DPAPI 사용 고려 |
| sg JSON 출력 형식 변경 | 파싱 실패 | 방어적 파싱 (nil 체크, 타입 어설션), 버전 체크 |
| SPEC-CONFIG-001 지연 | Rank 설정 로딩 불가 | 독립적 설정 로딩 폴백 (viper 직접 사용) |

---

## 6. 의존성 관리

### 외부 패키지

| 패키지 | 용도 | 모듈 |
|--------|------|------|
| `gopkg.in/yaml.v3` | 규칙 YAML 파싱 | astgrep |
| `net/http` (stdlib) | Rank API 통신 | rank |
| `crypto/hmac`, `crypto/sha256` (stdlib) | HMAC 서명 | rank |
| `encoding/json` (stdlib) | JSON 직렬화/역직렬화 | 양쪽 모두 |
| `os/exec` (stdlib) | sg CLI 실행 | astgrep |
| `context` (stdlib) | 타임아웃/취소 | 양쪽 모두 |

### 내부 의존성

| 의존 대상 | SPEC | 용도 |
|-----------|------|------|
| internal/config/ | SPEC-CONFIG-001 | Rank 설정 로딩 |
| internal/core/git/ | SPEC-GIT-001 | Git 이벤트 감지 (훅 연동) |

---

## 7. 다음 단계

- SPEC-TOOL-001 승인 후: `/moai run SPEC-TOOL-001` 실행
- AST-Grep 구현 완료 후: quality 게이트 연동 검토
- Rank 구현 완료 후: CLI rank 서브커맨드 연동 (SPEC-CLI-001)
- Hook 연동 완료 후: session_end 훅에서 자동 세션 제출 테스트
