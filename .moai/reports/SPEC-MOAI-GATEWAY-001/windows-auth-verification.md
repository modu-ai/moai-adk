# Windows AUTH 저장 primitive 후보 검증

## Claim

Windows 전용 native ACL 생성·handle 검증과 파일 flush/rename 후보를 `internal/gateway/auth/private_windows.go`에 추가했다. `private_windows_test.go`는 실제 Windows에서 실행할 fixture 시험이다. 이 작업은 **Windows 저장 기능 완료가 아니다**. OpenStore의 ErrPlatformUnsupported와 syncDirectory의 fail-closed 경계를 유지했고, common store.go/broker.go 및 기존 잠금·송신·Owns 구현은 수정하지 않았다.

생성 시 현재 프로세스 사용자 SID를 owner로 두고 protected DACL의 단일 full-access ACE를 명시한다. 디렉터리 ACE는 자식 상속을 허용하며 파일은 별도 protected ACL로 생성한다. GetSecurityInfo(handle)로 owner·보호 비트·단일 ACE·SID·권한·상속 flags를 확인하고 reparse point와 잘못된 파일 형식을 거절한다. 기존 디렉터리는 채택하거나 권한을 고치지 않는다. 후보 파일은 CREATE_NEW로만 만들고 FlushFileBuffers 뒤 닫는다. 교체 후보는 같은 디렉터리에서 MOVEFILE_REPLACE_EXISTING|MOVEFILE_WRITE_THROUGH를 사용한다. shell ACL 명령과 새 dependency는 없다.

이 함수들은 아직 production Store에서 호출되지 않는다. 런타임 증거 없이 POSIX mode 검사를 Windows ACL 검사로 대체하거나 플랫폼 gate를 제거하지 않았다.

## Evidence

모든 Go 검증은 같은 invocation에서 `unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY &&` 뒤에 실행했다.

시험 파일을 먼저 작성한 뒤 Windows cross-compile RED:

`GOOS=windows GOARCH=amd64 GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway/auth -c -o /tmp/gateway-auth-windows.test.exe`

```text
# github.com/modu-ai/moai-adk/internal/gateway/auth [github.com/modu-ai/moai-adk/internal/gateway/auth.test]
internal/gateway/auth/private_windows_test.go:16:12: undefined: createWindowsPrivateDirectory
internal/gateway/auth/private_windows_test.go:17:12: undefined: validateWindowsPrivatePath
internal/gateway/auth/private_windows_test.go:19:12: undefined: writeWindowsPrivateCandidate
internal/gateway/auth/private_windows_test.go:20:12: undefined: validateWindowsPrivatePath
internal/gateway/auth/private_windows_test.go:21:12: undefined: writeWindowsPrivateCandidate
internal/gateway/auth/private_windows_test.go:23:12: undefined: replaceWindowsWriteThrough
internal/gateway/auth/private_windows_test.go:26:12: undefined: validateWindowsPrivatePath
internal/gateway/auth/private_windows_test.go:33:12: undefined: createWindowsPrivateDirectory
internal/gateway/auth/private_windows_test.go:39:12: undefined: validateWindowsPrivatePath
internal/gateway/auth/private_windows_test.go:43:12: undefined: createWindowsPrivateDirectory
internal/gateway/auth/private_windows_test.go:43:12: too many errors
```

구현 후 같은 cross-compile 명령: exit 0, stdout/stderr 없음. 이는 런타임 RED→GREEN 증거가 아니라 컴파일 경계 RED→GREEN이다.

`GOOS=windows GOARCH=amd64 GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway/auth`: exit 0, stdout/stderr 없음.

macOS 기존 AUTH 회귀 시험:

`GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway/auth -count=1 -timeout 60s`

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.710s
```

`git rev-parse --short HEAD`:

```text
81c1d58f9
```

## Baseline-attribution

2026-09-11, WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, HEAD 위와 동일. AUTH plan B와 E6 및 REQ-GA-004/005를 읽었다. Windows runtime는 없는 macOS 환경이며 설치 클라이언트나 사용자 credential에 접근하지 않았다. native 시험에 Skip을 넣지 않았다.

Win32 근거는 이번 작업에서 Microsoft 공식 문서를 열어 확인했다.

- [File Security and Access Rights](https://learn.microsoft.com/en-us/windows/win32/fileio/file-security-and-access-rights): CreateFile/CreateDirectory에 security descriptor를 지정하며 기본 DACL은 부모에서 상속한다.
- [GetSecurityInfo](https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-getsecurityinfo): 열린 object handle에서 security descriptor를 조회한다.
- [FlushFileBuffers](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers): 파일 buffer flush와 GENERIC_WRITE 요건을 설명한다. 파일 flush를 directory transaction 내구성 전체로 확대 해석하지 않았다.
- [MoveFileExW](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw): WRITE_THROUGH 설명과 다른 volume 이동 시 security descriptor 처리 차이를 확인했다. COPY_ALLOWED는 사용하지 않는다.

## Gaps

- Windows native 실행은 전부 미관측이다. ACL 거절, current-user readback, CREATE_NEW 충돌, 넓은 Everyone ACL 거절, 플랫폼 gate 유지 시험은 compile되었으나 실행되지 않았다. coverage나 native PASS를 주장하지 않는다.
- E6 종료에는 Windows runner에서 위 시험과 별도 프로세스 LockFileEx 경합·취소·프로세스 crash 후 잠금 해제, 다른 계정의 read/write 거절 증거가 필요하다. 실제 runtime TDD RED는 다음 native 실행에서 확보해야 한다.
- 공통 Store의 directory/lock/state/scratch/auth.json 검사를 플랫폼 primitive로 연결하는 작업이 남았다. broker가 만든 파일의 ACL 상속 여부도 native 환경에서 검증해야 한다. 단순히 GOOS gate를 없애면 안 된다.
- 부모 경로의 reparse/identity를 각 open부터 publish까지 보장하는 연결, directory durability/전원 장애 복구, sharing violation 실패 주입·이전 state 보존, 원자 교체 readback 검증은 아직 없다. rename 후보의 nil은 전체 저장 트랜잭션 성공 판정이 아니다.
- plan이 지정한 internal/atomicfile.Replace 재사용과 Windows WRITE_THROUGH 보강의 최종 연결 방식은 미정이다. unrelated atomicfile을 이 작업에서 바꾸지 않았다.
- 통합 브랜치 CI의 저장소 전체 판정은 PENDING이다. git commit/push/PR/merge 없음.

## Residual-risk

ACL helper는 현재 사용자 프로세스 token을 기준으로 한다. impersonated thread와 서비스 계정 운영은 검증하지 않았다. 경로 기반 생성/rename 후보는 상위 디렉터리 교체 경합을 단독으로 해결하지 않으므로 production 연결을 금지한다. 생성 뒤 검증 실패나 write 실패의 후보 파일 cleanup도 호출자의 소유권 보장이 필요하다. 현재 OpenStore gate가 이 미완성 primitive의 실계정 사용을 차단한다.
