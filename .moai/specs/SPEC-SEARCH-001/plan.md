---
id: SPEC-SEARCH-001
type: plan
---

# SPEC-SEARCH-001: 구현 계획

## 개발 방법론

TDD (Red-Green-Refactor) - `quality.yaml` 설정에 따름

## 구현 순서

### Phase 1: Foundation (의존성 없음)

**Priority: High**

**1. 상수 정의**
- `internal/defs/dirs.go`에 `DirMoaiSearch = ".moai/search"` 추가
- `internal/defs/files.go`에 `FileSearchDB = ".moai/search/sessions.db"` 추가

**2. DB 레이어 (RED-GREEN-REFACTOR)**
- RED: `internal/search/db_test.go` 작성
  - TestOpenDB_CreatesDirectory: DB 디렉터리 자동 생성 확인
  - TestOpenDB_WALMode: WAL 모드 활성화 확인
  - TestOpenDB_ReturnsExisting: 기존 DB 재사용 확인
- GREEN: `internal/search/db.go` 구현
  - `OpenDB(dbPath string) (*sql.DB, error)`: 디렉터리 생성 + WAL 모드 설정
- REFACTOR: 에러 래핑 정리

**3. 스키마 레이어 (RED-GREEN-REFACTOR)**
- RED: `internal/search/db_test.go`에 스키마 테스트 추가
  - TestCreateTables_SessionsTable: sessions 테이블 생성 확인
  - TestCreateTables_FTS5Table: messages FTS5 가상 테이블 생성 확인
  - TestCreateTables_Idempotent: 중복 호출 안전성 확인
- GREEN: `internal/search/schema.go` 구현
  - `CreateTables(db *sql.DB) error`: DDL 실행
- REFACTOR: DDL 상수 분리

**4. 파서 레이어 (RED-GREEN-REFACTOR)**
- RED: `internal/search/parser_test.go` 작성
  - TestParseJSONL_ValidRecords: 정상 레코드 파싱
  - TestParseJSONL_NoiseFiltering: XML 태그, 짧은 텍스트 필터링
  - TestParseJSONL_InvalidJSON: 잘못된 JSON 라인 건너뜀
  - TestParseJSONL_EmptyFile: 빈 파일 처리
  - TestParseJSONL_RoleFiltering: user/assistant만 추출
- GREEN: `internal/search/parser.go` 구현
  - `ParseJSONL(path, gitBranch, projectPath string) ([]Message, error)`
- REFACTOR: 노이즈 필터 로직 분리

### Phase 2: Core (Phase 1 의존)

**Priority: High**

**5. 인덱서 레이어 (RED-GREEN-REFACTOR)**
- RED: `internal/search/indexer_test.go` 작성
  - TestIsIndexed_NotIndexed: 미인덱싱 세션 false 반환
  - TestIsIndexed_AlreadyIndexed: 인덱싱 세션 true 반환
  - TestIndexSession_NewSession: 새 세션 인덱싱
  - TestIndexSession_Idempotent: 중복 인덱싱 안전성
  - TestIndexSession_WithMessages: 메시지 FTS5 삽입 확인
- GREEN: `internal/search/indexer.go` 구현
  - `IsIndexed(db *sql.DB, sessionID string) (bool, error)`
  - `IndexSession(db *sql.DB, sessionID, filePath, gitBranch, projectPath string) error`
- REFACTOR: 트랜잭션 처리 최적화

**6. 검색 레이어 (RED-GREEN-REFACTOR)**
- RED: `internal/search/searcher_test.go` 작성
  - TestSearch_BasicQuery: 기본 텍스트 검색
  - TestSearch_BM25Ranking: BM25 점수 정렬 확인
  - TestSearch_BranchFilter: --branch 필터
  - TestSearch_RoleFilter: --role 필터
  - TestSearch_DateFilter: --since, --until 필터
  - TestSearch_LimitResults: --limit 적용
  - TestSearch_NoResults: 결과 없을 때 빈 슬라이스 반환
  - TestSearch_CJKText: 한국어 trigram 매칭
- GREEN: `internal/search/searcher.go` 구현
  - `Search(db *sql.DB, opts SearchOptions) ([]SearchResult, error)`
- REFACTOR: 쿼리 빌더 패턴 적용

### Phase 3: CLI (Phase 2 의존)

**Priority: Medium**

**7. Cobra 서브커맨드 (RED-GREEN-REFACTOR)**
- RED: `internal/cli/search_test.go` 작성
  - TestSearchCmd_BasicUsage: 기본 검색 실행
  - TestSearchCmd_WithFlags: 플래그 파싱 확인
  - TestSearchCmd_IndexSession: --index-session 모드
  - TestSearchCmd_NoQuery: 쿼리 없을 때 에러
- GREEN: `internal/cli/search.go` 구현
  - `newSearchCmd() *cobra.Command`: 커맨드 정의
  - `runSearch(cmd *cobra.Command, args []string) error`: 실행 로직
- REFACTOR: 출력 포맷팅 (lipgloss renderCard 패턴)

**8. 루트 커맨드 등록**
- `internal/cli/root.go`에 `newSearchCmd()` 등록

### Phase 4: Integration (Phase 3 의존)

**Priority: Medium**

**9. SessionEnd 훅 통합**
- `internal/hook/session_end.go`에 `triggerSessionIndex()` 함수 추가
- `Handle()` 마지막에 비동기 서브프로세스 실행
- fire-and-forget 패턴 (cmd.Start() 후 Wait() 없음)

**10. 의존성 추가**
- `go.mod`에 `modernc.org/sqlite v1.35.0` 추가
- `go mod tidy` 실행

### Phase 5: Skill (Phase 4와 병렬)

**Priority: Low**

**11. 스킬 정의 파일 생성**
- `internal/template/templates/.claude/skills/moai/workflows/search.md`
- MoAI-ADK에서 `moai search` 사용 가이드

## 기술 접근

### SQLite FTS5 선택 근거
- 전문 검색 엔진 내장 (별도 인프라 불필요)
- BM25 랭킹 기본 지원
- trigram 토크나이저로 CJK 언어 지원
- WAL 모드로 읽기/쓰기 동시성 지원

### modernc.org/sqlite 선택 근거
- 순수 Go 구현 (CGO 불필요)
- FTS5 기본 활성화
- 크로스 플랫폼 호환성 (macOS, Linux, Windows)
- 활발한 유지보수 (2025년 기준)

### 비동기 인덱싱 패턴 선택 근거
- `SessionEnd` 훅 타임아웃(5초) 내 완료 보장
- `cmd.Start()` fire-and-forget으로 훅 블로킹 방지
- 고루틴 대신 서브프로세스: 훅 프로세스 종료 후에도 인덱싱 지속
- 큐 대신 직접 실행: v1 복잡도 최소화

### 노이즈 필터링 전략
- XML 시스템 태그 제거 (Claude Code 내부 메시지)
- 20자 미만 텍스트 건너뜀 (의미 없는 짧은 응답)
- `user`/`assistant` 역할만 인덱싱 (시스템 메시지 제외)

## 위험 및 대응

| 위험 | 영향 | 대응 |
|------|------|------|
| modernc.org/sqlite 빌드 시간 증가 | 빌드 속도 저하 | 캐시된 빌드로 완화, 초기 빌드만 영향 |
| 대용량 JSONL (>100MB) 인덱싱 지연 | 인덱싱 시간 초과 | v1에서는 문서화만, v2에서 스트리밍 처리 |
| 훅 서브프로세스 PATH 문제 | 인덱싱 실패 | `moai` 바이너리 절대 경로 사용 고려 |
| FTS5 trigram 검색 정확도 | 한국어 검색 노이즈 | trigram은 부분 문자열 매칭이므로 수용 가능 |

## 추적성

- SPEC: SPEC-SEARCH-001
- Spec: `.moai/specs/SPEC-SEARCH-001/spec.md`
- Acceptance: `.moai/specs/SPEC-SEARCH-001/acceptance.md`
