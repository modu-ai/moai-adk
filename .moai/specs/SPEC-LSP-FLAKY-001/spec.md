# SPEC-LSP-FLAKY-001: Ubuntu LSP Launcher ETXTBSY Flake Stabilization

## Status: DRAFT

## Overview

Ubuntu CI에서 `internal/lsp/subprocess` 패키지의 launcher 테스트가 간헐적으로 `text file busy` (ETXTBSY) 에러로 실패한다. 최근 20+회 CI 중 다수에서 다음 3개 테스트가 flake로 확인되었다:

- `TestLauncher_Launch_HappyPath`
- `TestLauncher_Launch_StdioPipesNonNil`
- `TestLauncher_Launch_WithArgs`

## Root Cause

Linux fork-exec ETXTBSY race. `t.Parallel()` 환경에서 한 테스트가 fake binary를 `os.Create → Write → Sync → Close → Chmod 0755` 시퀀스로 작성한 직후 `cmd.Start()` (fork+exec)을 호출할 때, 다른 병렬 테스트의 fork-in-progress 자식 프로세스가 그 binary의 writer fd를 잠시 inherit한 상태에서 exec()이 발생하면 커널이 ETXTBSY를 반환한다.

기존 mitigation (`writeFakeBinary` 의 Create→Write→Sync→Close 순서 + `O_CLOEXEC`)는 fork↔exec 사이의 짧은 race window를 닫지 못한다. CLOEXEC는 exec 시점에만 fd를 닫고, ETXTBSY 검사는 exec 직전 inode write refcount 확인 단계에서 일어나기 때문이다.

## Solution Strategy

**Shared pre-built fake binary**: `t.Parallel()` 시작 *이전*에 fake binary를 1회만 생성하고, 모든 launcher 테스트가 동일 path를 재사용한다. 병렬 실행 중 어떤 테스트도 새로운 file write를 수행하지 않으므로 ETXTBSY race가 근본 제거된다.

- Production code (`internal/lsp/subprocess/launcher.go`) **무수정**
- `t.Parallel()` **유지** (parallelism 손실 없음)
- 테스트 일부(`TestLauncher_Launch_StartFails` 등 고유 binary 필요)는 기존 패턴 유지

## Requirements

### REQ-LSP-FLAKY-001-001

[WHEN] launcher_test.go 의 fake binary 가 필요한 테스트들이 동시에 실행될 때
[THEN] 모든 테스트는 동일한 사전 작성(pre-built)된 fake binary path를 공유해야 하며, 테스트 실행 중 어떤 fake binary 파일에 대한 동시 file write도 발생하지 않아야 한다

### REQ-LSP-FLAKY-001-002

[WHEN] 공유 fake binary 가 처음 요청될 때
[THEN] `sync.OnceValues` 또는 `TestMain` 을 사용하여 패키지 전역에서 정확히 1회만 작성되어야 한다 (Create → Write → Sync → Close → Chmod 0755 sequence 유지)

### REQ-LSP-FLAKY-001-003

[WHEN] 테스트가 고유한 binary content 를 요구할 때 (예: `TestLauncher_Launch_StartFails` 의 non-executable file)
[THEN] 해당 테스트는 기존 패턴(`t.TempDir()` + `os.WriteFile`)을 유지하되, 결과 파일이 다른 병렬 테스트와 inode 충돌을 일으키지 않도록 격리되어야 한다

### REQ-LSP-FLAKY-001-004

[WHEN] `go test -race -count=20 ./internal/lsp/subprocess/...` 가 실행될 때
[THEN] 20회 연속 실행 중 단 1회의 ETXTBSY 실패도 발생해서는 안 된다

### REQ-LSP-FLAKY-001-005

[WHEN] CI Ubuntu runner 에서 PR 의 race detector 단계가 실행될 때
[THEN] 5회 연속 CI 실행에서 launcher 패키지 테스트가 모두 green 이어야 한다

## Acceptance Criteria

- AC-001: `writeFakeBinary` 가 sync.OnceValues 기반 helper로 대체되어 동일 content 의 binary 가 패키지당 1회만 작성된다
- AC-002: HappyPath / StdioPipesNonNil / WithArgs 3개 테스트가 공유 binary path 를 사용한다
- AC-003: 로컬 `go test -race -count=20 ./internal/lsp/subprocess/...` 20회 연속 PASS
- AC-004: PR CI 에서 Ubuntu race detector 단계가 5회 연속 PASS (PR 머지 직전 까지 history)
- AC-005: production `internal/lsp/subprocess/launcher.go` 변경 없음 (diff stat 으로 검증)

## Files Affected

**수정:**
- `internal/lsp/subprocess/launcher_test.go` (writeFakeBinary helper 리팩토링, 공유 fixture 도입)

**무수정 (검증):**
- `internal/lsp/subprocess/launcher.go`
- `internal/lsp/subprocess/supervisor.go`
- `internal/lsp/subprocess/supervisor_test.go`

## Out of Scope

- `TestGetOrSpawn_SingleflightBarrier_*` 의 hardening — 최근 20+회 CI 에서 실패 증거 없음. 향후 별도 관찰 후 필요 시 별 SPEC.
- `internal/lsp/subprocess/launcher.go` 의 production retry 로직 — test issue를 production code 로 leak 하지 않음.
- `t.Parallel()` 제거 — race 근본 제거가 가능하므로 parallelism 희생 불필요.

## Methodology

TDD (RED-GREEN-REFACTOR) — manager-cycle 위임.

- **RED**: 현재 helper 호출이 race 를 유발하는지 회귀를 잡을 빠른 단위 검증 (예: 동시 16-goroutine fake binary 생성+exec 시뮬레이션) 추가하여 변경 전후 차이 확인
- **GREEN**: sync.OnceValues 기반 shared fixture 도입, 3개 테스트가 공유 path 사용
- **REFACTOR**: writeFakeBinary 의 ETXTBSY mitigation 주석 갱신, 향후 회귀 방지 가이드 명시

## Files Affected (구현 대상)

| File | Change | Estimated LOC |
|------|--------|---------------|
| `internal/lsp/subprocess/launcher_test.go` | sync.OnceValues 기반 sharedFakeBinary helper 추가, 3개 테스트가 공유 path 사용, writeFakeBinary 는 isolation 필요한 테스트 전용으로 retain | ~30-50 |

총 영향 LOC: ~30-50 (test only, production zero).

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| 공유 binary 가 한 테스트에 의해 오염될 가능성 | Low | binary 는 read+exec 만 사용, write 없음. shared OnceValues 결과는 path string 만 반환 |
| TestStartFails 가 여전히 race 로 ETXTBSY | Very Low | 해당 테스트는 0o644 (non-exec) 파일을 작성하므로 exec 자체가 발생하지 않아 ETXTBSY 영향 없음 |
| Windows / macOS regression | Very Low | shared helper 도 기존 `runtime.GOOS == "windows"` skip 로직 유지. macOS 는 ETXTBSY 가 발생하지 않으므로 동등 동작 |
