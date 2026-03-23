---
id: SPEC-REFACTOR-001
version: "1.0.0"
status: completed
created: "2026-03-19"
updated: "2026-03-20"
author: GOOS
priority: high
issue_number: 0
---

## HISTORY

| 버전 | 날짜 | 변경 내용 |
|------|------|----------|
| 1.0.0 | 2026-03-19 | 초안 작성 |
| 1.1.0 | 2026-03-20 | 구현 완료, SPEC 상태 completed |

---

# SPEC-REFACTOR-001: moai rank 기능 완전 제거

## 1. 개요

MoAI-ADK-Go에서 `moai rank` 리더보드 기능을 완전히 제거한다. 이 기능은 MoAI Cloud 서비스와 연동하여 세션 메트릭을 수집/제출하는 기능이었으나, 서비스 방향 변경으로 더 이상 필요하지 않다.

## 2. 요구사항 (EARS 포맷)

### REQ-REFACTOR-001: 패키지 제거 (Ubiquitous)

시스템은 `internal/rank/` 패키지를 포함하지 않아야 한다.

- `internal/rank/` 디렉터리와 모든 하위 파일이 삭제되어야 한다
- 어떤 Go 소스 파일도 `internal/rank` 패키지를 import하지 않아야 한다

### REQ-REFACTOR-002: CLI 명령어 제거 (Event-Driven)

사용자가 `moai rank` 명령어를 실행하면, 시스템은 "unknown command" 오류를 반환해야 한다.

- `rankCmd` cobra 명령어와 7개 서브커맨드 (login, status, logout, sync, exclude, include, register) 모두 제거
- `internal/cli/rank.go` 파일 삭제

### REQ-REFACTOR-003: Hook 핸들러 제거 (State-Driven)

시스템이 실행 중인 동안, rank 관련 SessionEnd 훅 핸들러가 등록되지 않아야 한다.

- `internal/hook/rank_session.go` 파일 삭제
- `InitDependencies()`에서 `EnsureRankSessionHandler()` 호출 제거
- `MOAI_RANK_ENABLED`, `MOAI_RANK_TIMEOUT` 환경 변수 처리 코드 제거

### REQ-REFACTOR-004: 의존성 주입 정리 (Complex)

`Dependencies` 구조체에서 rank 관련 필드와 메서드가 모두 제거되어야 하며, 나머지 의존성 기능에 영향이 없어야 한다.

- `RankClient rank.Client` 필드 제거
- `RankCredStore rank.CredentialStore` 필드 제거
- `RankBrowser rank.BrowserOpener` 필드 제거
- `EnsureRank()` 메서드 제거
- `InitDependencies()`에서 `RankCredStore` 초기화 코드 제거

### REQ-REFACTOR-005: 상수 및 레거시 참조 정리 (Ubiquitous)

시스템은 rank 관련 상수와 레거시 참조를 포함하지 않아야 한다.

- `internal/defs/dirs.go`에서 `RankSubdir` 상수 제거
- `internal/defs/files.go`에서 `CredentialsJSON` 상수 제거
- `internal/cli/update.go`에서 `"session_end__rank_submit"` 레거시 참조 제거

## 3. 범위

### 포함 (In Scope)

- `internal/rank/` 패키지 전체 삭제 (17개 파일)
- `internal/cli/rank.go` 삭제
- `internal/hook/rank_session.go` 삭제
- `internal/cli/deps.go` rank 참조 제거
- `internal/defs/` 상수 제거
- 관련 테스트 파일 삭제/수정 (~10개 파일)

### 제외 (Out of Scope)

- `.moai/` 프로젝트 문서 업데이트 (sync 단계에서 처리)
- `~/.claude/hooks/rank-submit.sh` 전역 스크립트 삭제 (사용자 시스템 파일)
- `~/.moai/rank/` 사용자 데이터 삭제 (사용자 판단)

## 4. 기술 제약사항

- Go 1.23+ 호환성 유지
- `go build ./...` 제로 에러
- `go test ./...` 전체 통과
- `go test -race ./...` 레이스 조건 없음
- 영향받는 패키지 커버리지 85%+ 유지

## 5. 의존성

- 이 SPEC은 다른 SPEC에 의존하지 않음
- `internal/rank/`는 leaf 패키지로 다른 내부 패키지에 의존하지 않음
- 제거 후 `go.mod`에 불필요한 의존성 발생 가능성 없음 (표준 라이브러리만 사용)
