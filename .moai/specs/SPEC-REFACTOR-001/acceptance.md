# SPEC-REFACTOR-001 인수 조건

## AC-1: 빌드 성공

**Given** rank 관련 모든 코드가 제거된 상태에서
**When** `go build ./...`을 실행하면
**Then** 제로 에러로 빌드가 성공해야 한다

## AC-2: 테스트 전체 통과

**Given** rank 관련 코드와 테스트가 모두 제거된 상태에서
**When** `go test ./...`을 실행하면
**Then** 모든 테스트가 통과해야 한다 (FAIL 없음)

## AC-3: 레이스 조건 검증

**Given** rank 코드 제거 후
**When** `go test -race ./...`을 실행하면
**Then** 레이스 조건이 감지되지 않아야 한다

## AC-4: rank 패키지 부재 확인

**Given** 코드 제거가 완료된 상태에서
**When** `internal/rank/` 디렉터리 존재 여부를 확인하면
**Then** 해당 디렉터리가 존재하지 않아야 한다

## AC-5: rank import 부재 확인

**Given** 코드 제거가 완료된 상태에서
**When** 모든 `.go` 파일에서 `"github.com/modu-ai/moai-adk/internal/rank"` import를 검색하면
**Then** 일치하는 결과가 0건이어야 한다

## AC-6: CLI 명령어 제거 확인

**Given** `moai` 바이너리가 빌드된 상태에서
**When** `moai rank` 명령을 실행하면
**Then** "unknown command" 오류가 반환되어야 한다

## AC-7: Dependencies 구조체 정리 확인

**Given** `internal/cli/deps.go` 파일에서
**When** `RankClient`, `RankCredStore`, `RankBrowser`, `EnsureRank` 키워드를 검색하면
**Then** 일치하는 결과가 0건이어야 한다

## AC-8: 상수 제거 확인

**Given** `internal/defs/` 패키지에서
**When** `RankSubdir`, `CredentialsJSON` 키워드를 검색하면
**Then** 일치하는 결과가 0건이어야 한다

## 엣지 케이스

### EC-1: Hook 등록 안전성

**Given** rank 핸들러가 제거된 상태에서
**When** `InitDependencies()`가 실행되면
**Then** 나머지 모든 훅 핸들러가 정상 등록되어야 한다 (SessionStart, SessionEnd, PreTool, PostTool 등)

### EC-2: 기존 기능 회귀 없음

**Given** rank 이외의 CLI 명령어들 (init, update, hook, cc, glm, version 등)
**When** 해당 명령어를 실행하면
**Then** 기존과 동일하게 동작해야 한다

## 품질 게이트

- `go vet ./...` 통과
- `golangci-lint run` 통과 (설정 파일 존재 시)
- 영향받는 패키지 커버리지 85%+ 유지
