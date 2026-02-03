# SPEC-TOOL-001: 수락 기준

---
spec_id: SPEC-TOOL-001
title: AST-Grep & Performance Ranking Integration - Acceptance Criteria
created: 2026-02-03
tags: [ast-grep, rank, acceptance, gherkin]
---

## 1. AST-Grep 모듈 수락 기준

### AC-AST-001: sg CLI 가용성 감지

**Given** sg CLI가 시스템 PATH에 설치되어 있을 때
**When** Analyzer가 초기화되면
**Then** `isSGAvailable()`은 true를 반환해야 한다

**Given** sg CLI가 시스템에 설치되어 있지 않을 때
**When** Analyzer가 초기화되면
**Then** `isSGAvailable()`은 false를 반환해야 한다
**And** 에러가 발생하지 않아야 한다

**Given** sg CLI 버전 확인이 5초 이상 걸릴 때
**When** `isSGAvailable()`이 호출되면
**Then** 타임아웃 후 false를 반환해야 한다

**Given** `isSGAvailable()`이 이미 호출된 적이 있을 때
**When** 다시 호출되면
**Then** 캐싱된 결과를 반환하고, sg CLI를 재실행하지 않아야 한다

---

### AC-AST-002: 파일 확장자 기반 언어 감지

**Given** ".py" 확장자를 가진 파일 경로가 주어질 때
**When** 언어 감지를 수행하면
**Then** "python"을 반환해야 한다

**Given** ".tsx" 확장자를 가진 파일 경로가 주어질 때
**When** 언어 감지를 수행하면
**Then** "typescriptreact"를 반환해야 한다

**Given** ".unknown" 확장자를 가진 파일 경로가 주어질 때
**When** 언어 감지를 수행하면
**Then** "text"를 반환해야 한다

**Given** 22개 지원 확장자 각각에 대해
**When** 언어 감지를 수행하면
**Then** Python 참조 구현의 EXTENSION_TO_LANGUAGE 맵과 동일한 결과를 반환해야 한다

---

### AC-AST-003: 단일 파일 스캔

**Given** 유효한 Python 파일 경로와 sg CLI가 설치되어 있을 때
**When** `scanFile()`을 호출하면
**Then** ScanResult를 반환해야 한다
**And** ScanResult에는 파일 경로, 매치 목록, 스캔 시간(ms), 언어 정보가 포함되어야 한다

**Given** 존재하지 않는 파일 경로가 주어질 때
**When** `scanFile()`을 호출하면
**Then** 에러를 반환해야 한다

**Given** sg CLI가 설치되어 있지 않을 때
**When** `scanFile()`을 호출하면
**Then** 빈 매치 목록이 포함된 ScanResult를 반환해야 한다
**And** 에러가 발생하지 않아야 한다

**Given** 커스텀 ScanConfig가 제공될 때
**When** `scanFile()`을 호출하면
**Then** 제공된 설정의 규칙 경로와 언어 옵션이 sg 명령에 전달되어야 한다

---

### AC-AST-004: 프로젝트 전체 스캔

**Given** 여러 언어의 소스 파일이 포함된 프로젝트 디렉토리가 주어질 때
**When** `scanProject()`를 호출하면
**Then** ProjectScanResult를 반환해야 한다
**And** 지원되는 확장자를 가진 모든 파일이 스캔되어야 한다
**And** 심각도별 집계(error, warning, info, hint)가 포함되어야 한다

**Given** ScanConfig에 exclude_patterns로 "node_modules"가 설정되어 있을 때
**When** `scanProject()`를 호출하면
**Then** node_modules 디렉토리 내 파일은 스캔에서 제외되어야 한다

**Given** ScanConfig에 include_patterns로 "*.go"가 설정되어 있을 때
**When** `scanProject()`를 호출하면
**Then** .go 파일만 스캔 대상에 포함되어야 한다

**Given** 존재하지 않는 프로젝트 경로가 주어질 때
**When** `scanProject()`를 호출하면
**Then** 에러를 반환해야 한다

---

### AC-AST-005: 패턴 검색

**Given** AST 패턴 `func $NAME($PARAMS)`, 언어 "go", 유효한 경로가 주어질 때
**When** `patternSearch()`를 호출하면
**Then** 매칭된 함수 선언의 Match 목록을 반환해야 한다
**And** 각 Match의 rule_id는 "pattern:" 접두사로 시작해야 한다

**Given** sg CLI가 설치되어 있지 않을 때
**When** `patternSearch()`를 호출하면
**Then** 빈 매치 목록을 반환해야 한다
**And** 에러가 발생하지 않아야 한다

**Given** 유효한 패턴이지만 매칭이 없을 때
**When** `patternSearch()`를 호출하면
**Then** 빈 매치 목록을 반환해야 한다

---

### AC-AST-006: 패턴 교체

**Given** 패턴과 대체 패턴, dry_run=true가 설정되어 있을 때
**When** `patternReplace()`를 호출하면
**Then** ReplaceResult를 반환해야 한다
**And** matches_found에 매칭 수가 포함되어야 한다
**And** changes 목록에 각 변경 사항이 포함되어야 한다
**And** 실제 파일은 수정되지 않아야 한다

**Given** 패턴과 대체 패턴, dry_run=false가 설정되어 있을 때
**When** `patternReplace()`를 호출하면
**Then** sg rewrite 명령이 실제로 실행되어야 한다
**And** 파일이 수정되어야 한다

**Given** sg CLI가 설치되어 있지 않을 때
**When** `patternReplace()`를 호출하면
**Then** matches_found=0, files_modified=0인 ReplaceResult를 반환해야 한다

---

### AC-AST-007: YAML 규칙 로딩

**Given** 유효한 ast-grep 규칙 YAML 파일이 주어질 때
**When** `LoadFromFile()`을 호출하면
**Then** Rule 객체 목록을 반환해야 한다
**And** 각 Rule에는 id, language, severity, message, pattern이 포함되어야 한다

**Given** 멀티 문서 YAML 파일(--- 구분)이 주어질 때
**When** `LoadFromFile()`을 호출하면
**Then** 모든 문서에서 규칙을 파싱하여 반환해야 한다

**Given** 존재하지 않는 파일 경로가 주어질 때
**When** `LoadFromFile()`을 호출하면
**Then** FileNotFoundError를 반환해야 한다

**Given** 잘못된 YAML 형식의 파일이 주어질 때
**When** `LoadFromFile()`을 호출하면
**Then** 파싱 에러를 반환해야 한다

**Given** 규칙 파일이 포함된 디렉토리가 주어질 때
**When** `LoadFromDirectory()`를 호출하면
**Then** .yml 및 .yaml 파일에서 모든 규칙을 로드해야 한다

**Given** 여러 언어의 규칙이 로드된 상태에서
**When** `GetRulesForLanguage("go")`를 호출하면
**Then** Go 언어에 해당하는 규칙만 반환해야 한다

---

## 2. Rank 모듈 수락 기준

### AC-RANK-001: API 클라이언트 초기화

**Given** API 키가 저장되어 있을 때
**When** RankClient가 생성되면
**Then** 저장된 API 키를 자동으로 로드해야 한다
**And** User-Agent 헤더는 "moai-adk/1.0"이어야 한다
**And** Content-Type 헤더는 "application/json"이어야 한다

**Given** API 키가 명시적으로 전달될 때
**When** RankClient가 생성되면
**Then** 전달된 API 키를 사용해야 한다

**Given** MOAI_RANK_API_URL 환경 변수가 설정되어 있을 때
**When** Config가 생성되면
**Then** 환경 변수의 URL을 기본 URL로 사용해야 한다

---

### AC-RANK-002: HMAC-SHA256 서명

**Given** API 키 "test-key", 타임스탬프 "1234567890", 바디 '{"data":"value"}'가 주어질 때
**When** `computeSignature()`를 호출하면
**Then** HMAC-SHA256("test-key", "1234567890:{\"data\":\"value\"}")의 16진수 결과를 반환해야 한다

**Given** 인증이 필요한 요청에서 API 키가 설정되지 않았을 때
**When** `getAuthHeaders()`를 호출하면
**Then** AuthenticationError를 반환해야 한다

**Given** HMAC 인증 헤더가 생성될 때
**Then** X-API-Key, X-Timestamp, X-Signature 세 가지 헤더가 모두 포함되어야 한다

---

### AC-RANK-003: 사용자 랭킹 조회

**Given** 유효한 API 키로 인증된 클라이언트가 있을 때
**When** `GetUserRank()`를 호출하면
**Then** UserRank 구조체를 반환해야 한다
**And** 일간/주간/월간/전체 랭킹 정보가 포함되어야 한다
**And** 토큰 통계(총 토큰, 총 세션, 입력/출력 토큰)가 포함되어야 한다

**Given** API 키가 유효하지 않을 때
**When** `GetUserRank()`를 호출하면
**Then** AuthenticationError를 반환해야 한다

**Given** API 서버가 응답하지 않을 때
**When** `GetUserRank()`를 호출하면
**Then** 타임아웃 후 ClientError를 반환해야 한다

---

### AC-RANK-004: 리더보드 조회

**Given** 기간 "weekly", 제한 10, 오프셋 0이 주어질 때
**When** `GetLeaderboard()`를 호출하면
**Then** LeaderboardEntry 목록을 반환해야 한다
**And** 각 항목에는 rank, username, total_tokens, composite_score가 포함되어야 한다

**Given** 인증되지 않은 상태에서
**When** `GetLeaderboard()`를 호출하면
**Then** 정상적으로 결과를 반환해야 한다 (공개 API)

**Given** limit가 100을 초과할 때
**When** `GetLeaderboard()`를 호출하면
**Then** 자동으로 100으로 클램핑되어야 한다

---

### AC-RANK-005: 세션 제출

**Given** 유효한 SessionSubmission 데이터와 인증된 클라이언트가 있을 때
**When** `SubmitSession()`을 호출하면
**Then** HMAC 서명이 포함된 POST 요청이 /api/v1/sessions로 전송되어야 한다

**Given** input_tokens가 100,000,000을 초과할 때
**When** `SubmitSession()`을 호출하면
**Then** 토큰 값은 100,000,000으로 클램핑되어야 한다

**Given** 선택적 대시보드 필드(tool_usage, model_usage, code_metrics)가 제공될 때
**When** `SubmitSession()`을 호출하면
**Then** 해당 필드가 요청 바디에 포함되어야 한다

**Given** 선택적 필드가 비어있거나 nil일 때
**When** `SubmitSession()`을 호출하면
**Then** 해당 필드는 요청 바디에서 생략되어야 한다

---

### AC-RANK-006: 배치 세션 제출

**Given** 50개의 SessionSubmission 목록이 주어질 때
**When** `SubmitSessionsBatch()`를 호출하면
**Then** 단일 POST 요청으로 /api/v1/sessions/batch에 전송되어야 한다
**And** BatchResult에 processed, succeeded, failed 수가 포함되어야 한다

**Given** 101개의 세션이 주어질 때
**When** `SubmitSessionsBatch()`를 호출하면
**Then** 에러를 반환해야 한다 (최대 100개 제한)

**Given** 배치 내 각 세션의 토큰 필드가 상한을 초과할 때
**When** `SubmitSessionsBatch()`를 호출하면
**Then** 모든 토큰 필드가 100,000,000으로 클램핑되어야 한다

---

### AC-RANK-007: OAuth 인증 플로우 - 로그인

**Given** 사용자가 `moai rank login`을 실행할 때
**When** OAuth 플로우가 시작되면
**Then** 로컬 HTTP 서버가 가용 포트(8080-8180)에서 시작되어야 한다
**And** CSRF 방지용 state 토큰이 생성되어야 한다
**And** 브라우저에 인증 URL이 열려야 한다
**And** CLI에 인증 URL이 출력되어야 한다

**Given** GitHub OAuth 콜백이 유효한 state와 API 키를 포함할 때
**When** 콜백이 수신되면
**Then** 크리덴셜이 ~/.moai/rank/credentials.json에 저장되어야 한다
**And** 성공 HTML 페이지가 브라우저에 표시되어야 한다

**Given** OAuth 콜백의 state 토큰이 불일치할 때
**When** 콜백이 수신되면
**Then** "State mismatch" 에러가 설정되어야 한다
**And** 크리덴셜이 저장되지 않아야 한다
**And** 400 에러 페이지가 브라우저에 표시되어야 한다

**Given** 인증 타임아웃(300초)이 경과할 때
**When** 아직 콜백이 수신되지 않았으면
**Then** "Authorization timed out" 에러를 반환해야 한다
**And** 로컬 서버가 정리되어야 한다

---

### AC-RANK-008: 크리덴셜 관리 - 로그아웃

**Given** 크리덴셜이 저장되어 있을 때
**When** `Delete()`를 호출하면
**Then** credentials.json 파일이 삭제되어야 한다
**And** true를 반환해야 한다

**Given** 크리덴셜이 저장되어 있지 않을 때
**When** `Delete()`를 호출하면
**Then** false를 반환해야 한다
**And** 에러가 발생하지 않아야 한다

---

### AC-RANK-009: 크리덴셜 저장소 보안

**Given** 크리덴셜이 저장될 때
**When** Save()가 완료되면
**Then** credentials.json 파일 권한은 0600이어야 한다
**And** ~/.moai/rank/ 디렉토리 권한은 0700이어야 한다

**Given** 크리덴셜 저장 중 에러가 발생할 때
**When** 임시 파일이 생성된 상태이면
**Then** 임시 파일이 정리되어야 한다
**And** 기존 크리덴셜 파일이 손상되지 않아야 한다

**Given** 크리덴셜 파일이 손상된(잘못된 JSON) 상태일 때
**When** `Load()`를 호출하면
**Then** nil을 반환해야 한다
**And** 에러가 발생하지 않아야 한다 (graceful)

---

### AC-RANK-010: 세션 해시 생성

**Given** 동일한 세션 데이터가 두 번 주어질 때
**When** `ComputeSessionHash()`를 각각 호출하면
**Then** 서로 다른 해시를 반환해야 한다 (암호학적 랜덤 요소 포함)

**Given** 세션 데이터가 주어질 때
**When** `ComputeSessionHash()`를 호출하면
**Then** 64자 16진수 SHA-256 해시를 반환해야 한다

---

### AC-RANK-011: API 상태 확인

**Given** Rank 서비스가 정상 동작 중일 때
**When** `CheckStatus()`를 호출하면
**Then** ApiStatus를 반환해야 한다
**And** status, version, timestamp 필드가 포함되어야 한다

**Given** Rank 서비스가 다운되어 있을 때
**When** `CheckStatus()`를 호출하면
**Then** ClientError를 반환해야 한다

---

## 3. 보안 수락 기준

### AC-SEC-001: API 키 로깅 방지

**Given** Rank API와 통신하는 모든 코드에서
**When** 로그를 출력할 때
**Then** API 키의 평문이 로그에 포함되지 않아야 한다

### AC-SEC-002: JSON 생성 안전성

**Given** Rank 클라이언트가 JSON 요청 바디를 생성할 때
**When** 바디가 생성되면
**Then** `json.Marshal()` 또는 `json.MarshalIndent()`를 통해 생성되어야 한다
**And** 문자열 결합(`fmt.Sprintf`, `+`)을 통한 JSON 생성이 코드에 존재하지 않아야 한다

### AC-SEC-003: CSRF 방지

**Given** OAuth 콜백에서 state 파라미터가 없거나 불일치할 때
**When** 콜백을 처리하면
**Then** 인증이 진행되지 않아야 한다
**And** 에러 응답이 반환되어야 한다

---

## 4. 성능 수락 기준

### AC-PERF-001: 스캔 타임아웃

**Given** sg CLI가 설치되어 있을 때
**When** 단일 파일 스캔이 60초를 초과하면
**Then** Context 취소에 의해 스캔이 중단되어야 한다

### AC-PERF-002: API 요청 타임아웃

**Given** Rank API 서버가 응답하지 않을 때
**When** 30초가 경과하면
**Then** 요청이 타임아웃되어야 한다
**And** ClientError가 반환되어야 한다

---

## 5. 호환성 수락 기준

### AC-COMPAT-001: Python 참조 구현 호환성

**Given** Python 참조 구현의 EXTENSION_TO_LANGUAGE 맵이 있을 때
**When** Go 구현의 언어 매핑과 비교하면
**Then** 22개 확장자에 대해 동일한 언어 문자열을 반환해야 한다

### AC-COMPAT-002: API 응답 형식 호환성

**Given** Rank API의 응답 JSON이 주어질 때
**When** Go 구조체로 역직렬화하면
**Then** Python 참조 구현과 동일한 필드를 올바르게 파싱해야 한다
**And** camelCase (API) -> snake_case/Go naming 변환이 올바르게 처리되어야 한다

### AC-COMPAT-003: 크리덴셜 파일 호환성

**Given** Python ADK가 생성한 credentials.json이 있을 때
**When** Go 구현이 `Load()`를 호출하면
**Then** 기존 크리덴셜을 정상적으로 로드해야 한다

---

## 6. Definition of Done (완료 정의)

### 필수 항목

- [ ] 모든 Go 인터페이스가 정의되고 구현됨
- [ ] 단위 테스트 커버리지 85% 이상
- [ ] 모든 수락 기준 시나리오에 대한 테스트가 작성됨
- [ ] `go vet` 경고 없음
- [ ] `golangci-lint` 에러 없음
- [ ] `go test -race` 통과
- [ ] 에러 메시지에 API 키가 노출되지 않음
- [ ] 모든 JSON은 Go 구조체 직렬화로 생성됨
- [ ] context.Context가 모든 장시간 작업에 전파됨
- [ ] 인터페이스 기반 설계로 모킹 테스트 가능

### 권장 항목

- [ ] 벤치마크 테스트 작성 (파일 스캔, HMAC 계산)
- [ ] sg CLI 통합 테스트 (빌드 태그 분리)
- [ ] godoc 코멘트 작성 (모든 exported 심볼)
- [ ] 크로스 플랫폼 크리덴셜 저장 테스트 (macOS, Linux)
