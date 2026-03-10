---
id: SPEC-SEARCH-001
type: acceptance
---

# SPEC-SEARCH-001: 인수 기준

## AC-SEARCH-001: SessionEnd 훅 비동기 인덱싱

**Given** 완료된 Claude Code 세션이 있고
**When** `SessionEnd` 훅이 5초 이내에 실행되면
**Then** 서브프로세스 `moai search --index-session <id>`가 비동기적으로 시작되어야 한다
**And** 훅 실행이 서브프로세스 완료를 기다리지 않아야 한다

### 검증 방법
- `session_end.go`의 `triggerSessionIndex()` 호출 확인
- `cmd.Start()` 후 `cmd.Wait()` 없음 확인
- 훅 타임아웃(5초) 내 반환 확인

---

## AC-SEARCH-002: BM25 랭킹 검색 결과

**Given** 인덱싱된 데이터베이스가 있고
**When** `moai search "authentication"`이 실행되면
**Then** 결과는 BM25 관련성 점수 내림차순으로 정렬되어야 한다
**And** 각 결과에 session_id, role, timestamp, excerpt (최대 200자), score가 표시되어야 한다

### 검증 방법
- `searcher_test.go`의 `TestSearch_BM25Ranking` 테스트
- 결과 필드 완전성 확인
- 점수 내림차순 정렬 확인

---

## AC-SEARCH-003: 인덱싱 성능

**Given** 22MB JSONL 파일에 1677개 assistant 레코드가 있고
**When** 인덱싱이 완료되면
**Then** 표준 하드웨어에서 30초 이내에 완료되어야 한다

### 검증 방법
- 대용량 JSONL 샘플로 벤치마크 테스트
- `time moai search --index-session <id>` 실행 시간 측정

---

## AC-SEARCH-004: CJK 텍스트 검색

**Given** 한국어 텍스트 "인증 구현"이 인덱싱되어 있고
**When** `moai search "인증"`이 실행되면
**Then** 한국어 레코드가 결과에 나타나야 한다 (trigram 부분 문자열 매칭)

### 검증 방법
- `searcher_test.go`의 `TestSearch_CJKText` 테스트
- 한국어, 일본어, 중국어 샘플 데이터로 검증

---

## AC-SEARCH-005: 멱등 인덱싱

**Given** 세션이 이미 DB에 있고
**When** `moai search --index-session <same-id>`가 호출되면
**Then** 중복 레코드 없이 성공을 반환해야 한다

### 검증 방법
- `indexer_test.go`의 `TestIndexSession_Idempotent` 테스트
- 동일 세션 2회 인덱싱 후 레코드 수 확인

---

## AC-SEARCH-006: 빈 데이터베이스 검색

**Given** 검색 DB가 비어있고
**When** `moai search "anything"`이 실행되면
**Then** "No results found" 메시지와 함께 exit code 0을 반환해야 한다

### 검증 방법
- `searcher_test.go`의 `TestSearch_NoResults` 테스트
- CLI 레벨 에러 코드 확인

---

## AC-SEARCH-007: Branch 필터링

**Given** `--branch feat/auth` 플래그가 제공되고
**When** `moai search "token"`이 실행되면
**Then** `git_branch = 'feat/auth'`인 세션의 메시지만 결과에 포함되어야 한다

### 검증 방법
- `searcher_test.go`의 `TestSearch_BranchFilter` 테스트
- 다른 브랜치 레코드 미포함 확인

---

## Quality Gate

### 테스트 커버리지
- `internal/search/` 패키지: 85% 이상
- `internal/cli/search.go`: 80% 이상

### 린트
- `golangci-lint run ./internal/search/...` 경고 0건
- `go vet ./internal/search/...` 에러 0건

### Definition of Done
- [ ] 모든 REQ-SEARCH-001~006 요구사항 구현 완료
- [ ] 모든 AC-SEARCH-001~007 인수 기준 통과
- [ ] 테스트 커버리지 목표 달성
- [ ] 린트/vet 클린
- [ ] `go test -race ./internal/search/...` 통과
- [ ] CLI 도움말 텍스트 작성 완료

## 추적성

- SPEC: SPEC-SEARCH-001
- Spec: `.moai/specs/SPEC-SEARCH-001/spec.md`
- Plan: `.moai/specs/SPEC-SEARCH-001/plan.md`
