# SPEC-UTIL-005 Research Notes

## §1. 배경 (v2.14 Review Finding)

v2.14.0 Phase 3.2 완료 후 `manager-quality` 멀티관점 코드 리뷰에서 다음 항목이 **Warning — PERFORMANCE** 등급으로 제기되었다.

- **Location**: `internal/hook/quality/astgrep_gate.go:29-41` (walkSourceFiles), 호출부 `RunAstGrepGateV2`
- **Category**: 성능 — 불필요한 I/O 및 CPU 반복
- **Severity**: Warning (v2.14 blocking은 아니며 v2.15 backlog로 defer)
- **Reviewer Note**: "`walkSourceFiles(projectDir)`가 `RunAstGrepGateV2` 호출마다 프로젝트 전체 디렉토리 트리를 재귀 탐색. Post-tool-use 훅은 파일 저장 시마다 트리거되므로 대형 모노레포(~10k 파일)에서 매 저장마다 전체 스캔 비용 누적."
- **Approach Hint**: "git diff 기반 changed-files only 또는 결과 캐싱. hook payload에 changedFiles 제공 시 활용."

## §2. Baseline (현재 비용 구조)

### 2.1 현재 동작

- `filepath.WalkDir(projectDir, ...)` 전체 트리 재귀
- 각 entry마다 exclusion check (vendor, node_modules, .git, ... — SPEC-UTIL-002 §3 규칙)
- source-file 필터(확장자 기반) + suppression pairing 수집

### 2.2 경험적 비용 (측정 필요, 현 시점 가정치)

| 파일 수 | 추정 소요시간 (SSD, macOS) |
|---------|-----------------------|
| 500 | ~2ms |
| 5,000 | ~20ms |
| 10,000 | ~40-80ms |
| 50,000 (대형 monorepo) | ~200-500ms |

- Post-tool-use 훅이 활발한 개발 세션(분당 5-10회 저장)에서 10k-file 기준 누적 ~200-800ms/분
- 실제 변경된 파일은 통상 1-5개 → 99%의 파일은 매번 무의미하게 스캔됨

### 2.3 Hook payload 구조 확인 포인트

`internal/hook/payload.go`(혹은 유사 파일)에 `PostToolUsePayload { ChangedFiles []string; ... }` 형태로 필드가 존재하는지 Plan 단계에서 재확인 필요. 존재하지 않는 경우 Options 섹션의 Option B로 fallback.

## §3. Options

### Option A — hook payload의 changedFiles 활용 (**preferred**)

- hook runtime이 이미 변경 파일 리스트를 알고 있음 → `RunAstGrepGateV2(ctx, projectDir, changedFiles)` 형태로 전달
- full scan 대비 99% I/O 제거
- Risk: payload의 정확성에 의존. Claude Code 측에서 누락 발생 시 탐지 불가 → `changedFiles == nil` 가드로 폴백

### Option B — git diff --name-only HEAD (fallback)

- `git diff --name-only HEAD` + `git ls-files --others --exclude-standard` 조합으로 changed + untracked 산출
- 장점: hook payload 독립 — 어떤 훅 트리거에서도 동작
- 단점: 외부 프로세스 exec 비용 (~10-30ms), stash/partial index 상태 처리 복잡
- Windows/Linux git 경로 normalization 추가 필요

### Option C — 파일 mtime 기반 결과 캐싱 (**rejected**)

- staleness 위험: 외부 에디터/포매터가 파일 수정 시 캐시 무효화 신호 없음
- 동일 파일이 여러 훅 호출 간 변경되어도 mtime 해상도 한계(초 단위)로 놓칠 가능성

### 3.1 결정: Option A + Option B 복합

- Primary: hook payload의 changedFiles 사용
- Secondary fallback: payload nil/empty 시 기존 full scan 유지 (backward-compat)
- Git diff는 v3.0 이슈로 defer (현재 SPEC 범위 밖)

## §4. References

- v2.14 `manager-quality` 멀티관점 리뷰 보고서 (Warning — PERFORMANCE 섹션)
- SPEC-UTIL-002 §3 (walkSourceFiles 원본 + exclusion 규칙)
- SPEC-UTIL-003 (Phase 3.3 quality gate 통합 테스트 — incremental path 회귀 가드 재사용 가능)
- `internal/hook/payload.go` 혹은 대응 훅 paylaod 정의 (Plan 단계에서 재확인)
