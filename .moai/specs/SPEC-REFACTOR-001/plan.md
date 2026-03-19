# SPEC-REFACTOR-001 구현 계획

## 개발 방법론

TDD (quality.yaml: development_mode: tdd)
- RED: 제거 대상이 존재하지 않음을 확인하는 검증 테스트 작성
- GREEN: 코드 제거 실행
- REFACTOR: 전체 빌드/테스트 검증

## 실행 계획

### Phase 1: 파일 삭제 (leaf-first 전략)

순서가 중요: leaf 패키지부터 삭제해야 중간 컴파일 오류를 최소화.

| 순서 | 파일/디렉토리 | 작업 | 근거 |
|------|-------------|------|------|
| 1.1 | `internal/rank/` (17파일) | 디렉토리 삭제 | Leaf 패키지, 내부 의존 없음 |
| 1.2 | `internal/cli/rank.go` | 파일 삭제 | CLI 명령어 + hook 관리 코드 |
| 1.3 | `internal/hook/rank_session.go` | 파일 삭제 | SessionEnd 훅 핸들러 |
| 1.4 | `internal/cli/rank_test.go` | 파일 삭제 | 전용 테스트 |
| 1.5 | `internal/cli/rank_nonauth_test.go` | 파일 삭제 | 전용 테스트 |
| 1.6 | `internal/hook/rank_session_test.go` | 파일 삭제 | 전용 테스트 |

### Phase 2: 참조 정리

| 순서 | 파일 | 수정 내용 |
|------|------|----------|
| 2.1 | `internal/cli/deps.go` | `rank` import 제거, `RankClient`/`RankCredStore`/`RankBrowser` 필드 제거, `EnsureRank()` 메서드 삭제, `InitDependencies()`에서 rank 초기화 코드 제거 |
| 2.2 | `internal/defs/dirs.go` | `RankSubdir = "rank"` 상수 제거 |
| 2.3 | `internal/defs/files.go` | `CredentialsJSON = "credentials.json"` 상수 제거 |
| 2.4 | `internal/cli/update.go:2310` | `"session_end__rank_submit"` 항목 제거 |

**검증 포인트**: `go build ./...` 통과

### Phase 3: 테스트 정리

| 순서 | 파일 | 수정 내용 |
|------|------|----------|
| 3.1 | `internal/cli/mock_test.go` | `mockRankClient`, `mockCredentialStore`, `mockBrowser` 구조체 제거, `rank` import 제거 |
| 3.2 | `internal/cli/integration_test.go` | `"rank"` 문자열 제거, rank 테스트 함수 7개 제거 |
| 3.3 | `internal/cli/coverage_test.go` | rank 관련 테스트 함수 ~10개 제거, `rank` import 제거 |
| 3.4 | `internal/cli/coverage_improvement_test.go` | rank 관련 테스트 함수 ~20개 제거 |
| 3.5 | `internal/cli/target_coverage_test.go` | rank 관련 테스트 함수 ~15개 제거 |
| 3.6 | `internal/cli/hook_e2e_test.go` | rank 관련 주석/참조 정리 |
| 3.7 | `internal/hook/coverage_boost_test.go` | rank 관련 테스트 함수 ~10개 제거, `rank` import 제거 |

**검증 포인트**: `go test ./...` 통과, `go test -race ./...` 통과

### Phase 4: 최종 검증

```bash
# 빌드 검증
go build ./...

# 테스트 검증
go test ./...

# 레이스 검증
go test -race ./...

# rank 잔여 참조 확인 (Go 소스만)
grep -r "internal/rank" --include="*.go" .
grep -r "RankClient\|RankCredStore\|RankBrowser\|EnsureRank" --include="*.go" .
```

## 위험 완화

| 위험 | 완화 전략 |
|------|----------|
| deps.go 편집 시 구조 손상 | Read → 정밀 Edit, 빌드 즉시 검증 |
| 테스트 함수 경계 오판 | 함수 단위로 정확한 범위 확인 후 삭제 |
| 커버리지 하락 | rank 코드와 테스트가 동시에 제거되므로 비율 유지 예상 |

## 실행 모드 권고

이 작업은 단일 도메인 (리팩토링) 작업으로 **순차 실행 (sub-agent)** 이 가장 적합하나, 사용자 요청에 따라 team 모드 사용 가능.

Team 모드 시 권장 분배:
- Teammate 1: Phase 1 (파일 삭제) + Phase 2 (참조 정리)
- Teammate 2: Phase 3 (테스트 정리)
- 단, Phase 2 완료 후 Phase 3 진행 필수 (빌드 통과 선행 조건)
