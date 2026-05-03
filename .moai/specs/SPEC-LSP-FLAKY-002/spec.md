# SPEC-LSP-FLAKY-002: LSP Launcher ETXTBSY Eager Initialization Hotfix

## Status: DRAFT

## Overview

SPEC-LSP-FLAKY-001 (PR #757) 머지 후에도 Ubuntu CI 에서 동일한 ETXTBSY flake 가 PR #758 에서 재발했다:

```
launcher_test.go:37: Launch returned unexpected error: subprocess.Launch
"/tmp/moai-lsp-subprocess-test-1440849762/shared-fake-lsp": start: fork/exec
/tmp/moai-lsp-subprocess-test-1440849762/shared-fake-lsp: text file busy
```

## Root Cause Analysis (정정)

SPEC-001 의 진단이 부분적으로 틀렸다. 실제 race 는 다음과 같다:

1. SPEC-001 의 `sharedBinaryPath` 는 `sync.OnceValues` 로 **lazy** 초기화됨
2. 따라서 첫 테스트 goroutine 이 `t.Parallel()` 진입 후에 binary write 를 시작함
3. 같은 패키지의 `supervisor_test.go` 도 t.Parallel() 로 자체 fake binary 를 동시에 작성하고 cmd.Start() 를 호출함
4. supervisor goroutine 이 fork() 할 때, launcher 의 `OnceValues` 안에서 열린 writer fd (`shared-fake-lsp` 의 writer) 가 supervisor 의 fork 자식에게 inherit 됨
5. supervisor 자식이 자기 binary 를 exec 하기 전 짧은 윈도우에 launcher 의 다른 goroutine 이 `cmd.Start(shared-fake-lsp)` 를 호출
6. 커널이 inode 를 검사: shared-fake-lsp 가 누군가 (supervisor fork 자식) 에게 writer 로 열려 있음 → ETXTBSY

핵심: O_CLOEXEC 는 exec 시점에만 fd 를 닫는데, **fork 와 exec 사이의 윈도우** 는 닫지 못한다. 그 윈도우가 ETXTBSY race 의 본질이다.

## Solution Strategy (수정)

**Eager initialization in TestMain**: lazy `sync.OnceValues` 를 제거하고, `TestMain` 의 `m.Run()` 호출 *이전* 에 binary 를 작성한다.

이 시점에는:
- 어떤 `t.Parallel()` goroutine 도 시작되지 않음
- 따라서 어떤 fork() 도 발생할 수 없음
- 따라서 어떤 자식 프로세스도 writer fd 를 inherit 할 수 없음

`m.Run()` 진입 시점에는 모든 writer fd 가 이미 닫힌 상태이므로 race window 자체가 존재하지 않는다.

## Requirements

### REQ-LSP-FLAKY-002-001

[WHEN] `internal/lsp/subprocess` 패키지의 테스트 프로세스가 시작될 때
[THEN] `TestMain` 은 `m.Run()` 호출 전에 fake LSP stub binary 를 패키지 전역 경로에 1회 작성하고, 모든 writer fd 를 닫아야 한다

### REQ-LSP-FLAKY-002-002

[WHEN] `sharedFakeBinaryPath(t)` 가 임의의 t.Parallel() 테스트에서 호출될 때
[THEN] 함수는 `pkgSharedBinaryPath` 변수의 값을 즉시 반환하며 어떤 file write 도 발생하지 않아야 한다

### REQ-LSP-FLAKY-002-003

[WHEN] `m.Run()` 이 진입한 이후 어떤 테스트 goroutine 도 `shared-fake-lsp` 에 write 하지 않아야 한다 (read+exec only)

### REQ-LSP-FLAKY-002-004

[WHEN] CI Ubuntu race detector 단계가 5회 연속 실행될 때
[THEN] launcher 패키지 테스트가 모두 green 이어야 하며 어떤 ETXTBSY 실패도 발생하지 않아야 한다

## Files Affected

**수정:**
- `internal/lsp/subprocess/launcher_main_test.go` (sync.OnceValues 제거, eager init 도입)

**무수정 (검증):**
- `internal/lsp/subprocess/launcher_test.go` (호출 시그니처 동일, 변경 불요)
- `internal/lsp/subprocess/launcher.go`
- `internal/lsp/subprocess/supervisor.go`
- `internal/lsp/subprocess/supervisor_test.go`

## Acceptance Criteria

- AC-001: `launcher_main_test.go` 에서 `sync.OnceValues` 제거, `TestMain` 이 `m.Run()` 전에 binary 작성
- AC-002: `pkgSharedBinaryPath` 패키지 변수가 m.Run() 진입 시점에 절대 경로로 초기화되어 있음
- AC-003: 로컬 `go test -race -count=20 ./internal/lsp/subprocess/...` 20회 연속 PASS
- AC-004: PR CI Ubuntu race detector 5회 연속 PASS

## Out of Scope

- supervisor_test.go 의 `writeFakeBinaryContent` 자체적으로는 race 를 유발하지 않음 (각자 고유 file 작성). 단, 그들이 launcher binary 의 writer fd 를 inherit 할 수 있는 race window 만 차단하면 충분.
- `internal/lsp/subprocess/launcher.go` production 코드는 무수정.

## Why SPEC-001 was incomplete

SPEC-001 은 "shared file 사용 시 file write 가 1회만 발생하므로 race 없음" 으로 추론했으나, lazy 초기화 패턴 자체가 race window 였다. 본 SPEC-002 는 eager 초기화로 race window 를 제거한다.
