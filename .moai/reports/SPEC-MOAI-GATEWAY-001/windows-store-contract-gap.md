# Windows AUTH 저장 계약 확인 및 잠금 시험 준비

## Claim

Windows Store 활성화는 아직 완료하지 않았다. 승인된 AUTH plan의 디렉터리 내구성 조건과 `internal/atomicfile.Replace` 지정이 Windows 후보의 직접 `MoveFileEx` 경로와 어떻게 연결되는지 먼저 명시해야 한다. 부모 조정자의 지시에 따라 product 파일은 수정하지 않고 기존 native LockFileEx의 취소·별도 프로세스 경합·강제 종료 후 잠금 해제 시험 두 개를 추가했다. CI 필수 manifest는 기존 7개를 보존하고 9개로 늘렸다.

추가 파일은 `internal/gateway/auth/lock_windows_test.go`다. `TestWindowsLockCancelledBeforeAcquisition`은 취소된 context가 잠금을 취득하지 않았는지 후속 취득으로 확인한다. `TestWindowsLockContentionCancellationAndCrashRelease`는 자식 프로세스의 LOCKED 신호를 받은 뒤 부모 취득의 deadline 초과를 확인하고, 자식을 강제 종료·회수한 뒤 다시 잠금을 취득한다. 자식에는 외부 context 8초와 내부 10초 제한이 있으며 실패 경로에도 cleanup에서 kill과 Wait를 수행한다. 이는 프로세스 종료의 잠금 해제 시험이며 전원 장애 시험이 아니다.

## Evidence

이 WT에서 실행한 명령과 실제 출력:

```text
moai session current
01a08e7b-6aa0-7361-ab7e-ea8da1f02228

git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway

GOOS=windows GOARCH=amd64 GOCACHE=/tmp/gateway-windows-store-cache go test -c ./internal/gateway/auth -o /tmp/gateway-windows-lock.test.exe
[stdout/stderr empty; exit 0]

GOOS=windows GOARCH=amd64 GOCACHE=/tmp/gateway-windows-store-cache go vet ./internal/gateway/auth
[stdout/stderr empty; exit 0]

unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/gateway/auth -run '^TestCrossProcessLockCrashReleasesOwnership$' -count=1 -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.871s

bash scripts/ci-census/windows-gateway-evidence-test.sh
"PASS Windows gateway native evidence: 2 required tests"
REJECTED missing
REJECTED skipped
REJECTED failed
REJECTED pass_without_run
REJECTED truncated
REJECTED malformed
REJECTED absent_stream

gofmt -l internal/gateway/auth/lock_windows_test.go
[stdout/stderr empty; exit 0]
jq 'length' scripts/ci-census/windows-gateway-required.json
9
git diff --check -- scripts/ci-census/windows-gateway-required.json
[stdout/stderr empty; exit 0]
```

위 checker의 2개는 합성 JSON fixture 개수다. 실제 Windows 실행 개수가 아니다. 새 native 시험은 기존 구현의 characterization이며 이번 실행에서 native RED/GREEN을 관측하지 않았다.

## Baseline-attribution

2026-09-11, macOS, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, `WT-unified-gateway`, HEAD `81c1d58f9`. 다른 agent가 같은 WT의 별도 AUTH·gateway 파일을 작업하므로 전체 diff를 이 변경으로 귀속하지 않는다. 이 작업의 파일 digest:

```text
4ccfe13874b22cfd8dc94349a1b5d1e40b7b04df608cef886ed3afbb6aa321e0  internal/gateway/auth/lock_windows_test.go
eeaed40d49edf33e30c3452f488fb522a46fc3b82b37ba11c69084a3161445ff  scripts/ci-census/windows-gateway-required.json
```

## Gaps

1. AUTH `plan.md:23-24`는 write·fsync·rename과 디렉터리 내구성 확인 전에 성공을 표시하지 못하게 한다. `plan.md:40-44`는 canonical 최종 교체를 `internal/atomicfile.Replace`로 지정한다. 그런데 현재 `internal/atomicfile/replace_windows.go:36-45`는 `os.Rename` 재시도이며, AUTH 후보 `private_windows.go:151-177`은 직접 `MoveFileEx(REPLACE_EXISTING|WRITE_THROUGH)`를 호출한다. 어느 경로를 사용할지와 Windows 내구성 판정은 SPEC body 변경 없이 임의 확정하지 않았다.
2. `store.go:60-61`의 Windows `ErrPlatformUnsupported`와 `lock_windows.go:35`의 `syncDirectory` 오류는 그대로다. 따라서 Windows Store 로그인·logout·세대 보존이 구현됐다고 주장할 수 없다.
3. `private_windows.go:92-113`의 후보는 protected DACL과 정확한 ACE flags를 요구한다. 일반 broker 파일의 실제 상속 ACL이 그 조건을 충족하는지는 Windows에서 관측하지 않았다. Store 연결 시 scratch 자체는 protected current-user ACL로 생성하고, broker 결과는 실제 handle의 owner·effective DACL·reparse·부모 identity를 검증하는 계약이 필요하다. 상속 권한을 무조건 수용하거나 chmod로 대체해서는 안 된다.
4. Windows CI run ID·native test-stream·artifact·실행 coverage가 없다. 새 시험은 cross-compile 및 vet만 완료했다. Windows Store 통합 시험·다른 계정 접근 거절·상위 경로 교체 경합·write/flush/rename 실패 후 canonical 보존·전원 장애 내구성은 미관측이다.

Microsoft [MoveFileEx](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexa)는 WRITE_THROUGH의 디스크 이동과 copy/delete flush를 설명한다. [FlushFileBuffers](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers)는 파일 또는 volume flush와 GENERIC_WRITE 요건을 설명한다. 이 문서만으로 POSIX directory fsync와 같은 계약을 입증하지 않았다.

## Residual-risk

manager-spec에 넘길 최소 결정은 Windows 최종 교체 경로를 AUTH 전용 native write-through로 예외 허용할지, 기존 atomicfile에 필요한 확장을 정식 범위로 넣을지다. 동시에 같은 volume에서 file flush·atomic replace·최종 handle ACL/identity/내용 readback을 어떤 성공 조건으로 삼는지, 프로세스 crash와 전원 장애 보장을 어떻게 구분하는지 명시해야 한다. 이는 제안이며 승인된 계약 변경이 아니다. 디렉터리 내구성을 삭제하거나 플랫폼 지원을 선언하는 방식으로 닫지 않는다.

기존 CI 요구사항은 유지했다. 새 manifest 이름과 native 시험 이름이 다르면 CI가 실패한다. 저장소 전체 판정은 integration/release branch의 실제 GitHub CI run 소관이며 현재 **PENDING**이다. push·PR·workflow dispatch는 수행하지 않았다.
