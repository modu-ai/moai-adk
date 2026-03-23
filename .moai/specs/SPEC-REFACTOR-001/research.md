# Research: moai rank 기능 완전 제거

## 1. 아키텍처 분석

### 의존성 그래프

```
internal/rank/ (leaf 패키지, 내부 의존성 없음)
  ├── 소비자: internal/cli/rank.go (CLI 명령어)
  ├── 소비자: internal/cli/deps.go (의존성 주입)
  ├── 소비자: internal/hook/rank_session.go (SessionEnd 훅)
  └── 소비자: 테스트 파일 다수
```

### 코드 인벤토리

| 카테고리 | 파일 | LOC (추정) | 작업 |
|---------|------|-----------|------|
| 핵심 패키지 | `internal/rank/` (17파일) | ~2,500 | 디렉토리 삭제 |
| CLI 명령어 | `internal/cli/rank.go` | 690 | 파일 삭제 |
| Hook 핸들러 | `internal/hook/rank_session.go` | 252 | 파일 삭제 |
| DI 참조 | `internal/cli/deps.go` | ~30줄 수정 | 편집 |
| 상수 | `internal/defs/dirs.go`, `files.go` | 2줄 | 편집 |
| 레거시 | `internal/cli/update.go:2310` | 1줄 | 편집 |
| 전용 테스트 | `rank_test.go`, `rank_nonauth_test.go`, `rank_session_test.go` | ~500 | 삭제 |
| 공유 테스트 | `mock_test.go`, `integration_test.go`, `coverage_*.go`, `hook_e2e_test.go`, `coverage_boost_test.go` | ~1,500 | 편집 |

### 격리 점수: 높음

`internal/rank/`는 표준 라이브러리만 의존하는 완전한 leaf 패키지. 다른 내부 패키지가 rank를 의존하지 않음 (cli와 hook만 소비).

## 2. 위험 분석

### 위험 요소

| 위험 | 영향 | 확률 | 완화 |
|------|------|------|------|
| deps.go 편집 시 다른 필드 손상 | 높음 | 낮음 | 정밀 편집, 빌드 검증 |
| 테스트 파일에서 rank 참조 누락 | 중간 | 중간 | `grep -r "rank\|Rank"` 최종 검증 |
| CredentialsJSON 상수가 rank 외 사용 | 중간 | 낮음 | 검증 완료: rank 전용 |
| 커버리지 하락 (테스트 대량 삭제) | 낮음 | 높음 | rank 테스트만 삭제, 비율은 유지/개선 |

### 안전한 삭제 확인

- `detectGoBinPath()`: rank.go 내부에서만 사용 (deployGlobalRankHookScript)
- `claudeSettings`/`hookGroup`/`hookEntry` 구조체: rank.go 내부에서만 사용
- `anonymizePath()`: rank_session.go 내부에서만 사용
- `submitSyncBatches()`: rank.go 내부에서만 사용

## 3. 실행 순서 권장

```
Phase 1: 파일 삭제 (leaf-first)
  1.1 internal/rank/ 디렉토리 삭제
  1.2 internal/cli/rank.go 삭제
  1.3 internal/hook/rank_session.go 삭제
  1.4 internal/cli/rank_test.go 삭제
  1.5 internal/cli/rank_nonauth_test.go 삭제
  1.6 internal/hook/rank_session_test.go 삭제

Phase 2: 참조 정리 (import 수정)
  2.1 internal/cli/deps.go 편집 (필드/메서드/import 제거)
  2.2 internal/defs/dirs.go 편집 (RankSubdir 제거)
  2.3 internal/defs/files.go 편집 (CredentialsJSON 제거)
  2.4 internal/cli/update.go 편집 (레거시 참조 제거)

→ go build ./... 통과 확인

Phase 3: 테스트 정리
  3.1 internal/cli/mock_test.go 편집 (rank 모킹 제거)
  3.2 internal/cli/integration_test.go 편집
  3.3 internal/cli/coverage_test.go 편집
  3.4 internal/cli/coverage_improvement_test.go 편집
  3.5 internal/cli/target_coverage_test.go 편집
  3.6 internal/cli/hook_e2e_test.go 편집
  3.7 internal/hook/coverage_boost_test.go 편집

→ go test ./... 통과 확인
→ go test -race ./... 통과 확인
```
