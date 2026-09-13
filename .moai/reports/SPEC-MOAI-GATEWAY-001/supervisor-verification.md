# M4 supervisor 기반 구현 검증

## Claim

별도 gateway child의 시작·private stdin 설정 전달·stdout 포트 인계·부모 프로세스 지문 감시·수명 상한·정상 종료와 회수·child 소유 overlay 정리 기반을 구현했다. CLI 연결과 실제 Claude exec·PTY·job control 검증은 미완료다.

### 실행 단계에서 구체화한 계약

- `StartChild(ctx, StartOptions)`는 공급된 실행 파일과 고정 내부 동사 인수로 별도 프로세스를 시작한다. 설정은 1MiB 이하 JSON 한 줄로 private stdin에 쓴다. stderr는 버린다. 토큰·credential을 argv에 넣는 인수는 받는 쪽 CLI가 만들지 않아야 하며 Payload는 stdin 전용이다.
- 현재 부모 PID와 `homestate.CurrentProcessFingerprint()`를 부모가 채운다. 지문이 없으면 시작을 거절하며 PID만으로 완화하지 않는다.
- child는 `ReadChildConfig(os.Stdin)` 후 Payload로 handler를 만들고 `RunChildWithControl(ctx,cfg,handler,os.Stdout,os.Stdin)`을 호출한다. stdout은 다른 UI 출력 없이 인계 전용이다.
- child는 항상 `tcp4`, `127.0.0.1:0`에 bind한다. 성공한 listener의 실제 주소를 `ChildHandoff{Address,OverlayPath}` JSON으로 전달한다. bind 실패면 인계 출력이 없다.
- Lifetime·PollInterval·StartupTimeout은 양수를 호출자가 공급한다. 제품 기본 수명은 정하지 않았다. design §3.4의 이중 안전망 때문에 Lifetime은 필수다. 시험의 짧은 값은 제품 기본값이 아니다.
- 부모→자식 stop은 stdin의 단일 byte 1이다. EOF는 종료 명령이 아니다. POSIX exec로 launcher의 pipe가 닫혀도 같은 PID·지문이 살아 있으면 child는 유지된다. Windows 역시 stdin/stdout pipe를 사용하므로 ExtraFiles가 필요 없다.
- `ChildProcess.Stop(ctx)`는 control byte를 보내고 최대 1초의 유예 뒤 응답 없는 child를 Kill한다. pipe write와 process Wait를 모두 join한다. 취소된 ctx에서도 Kill 뒤 reaper 완료 전 반환하지 않는다. `Wait(ctx)`는 완료 코드 또는 기다림 취소를 제공한다.
- POSIX child는 별도 세션으로 분리해 Claude의 foreground terminal signal을 직접 받지 않게 한다. Windows는 새 process group이며 살아 있는 launcher가 Stop/Wait로 직접 감독한다.
- child는 `homestate.ProbeProcessIdentity`로 부모 PID·지문을 재확인한다. 사망 또는 다른 지문이면 listener와 활성 연결을 닫는다. 신원 판정이 indeterminate이면 수명 상한까지 감시를 유지한다.
- overlay는 caller의 기존 파일을 삭제하는 API가 아니다. child가 private 0700 새 디렉터리 안에 0600 파일로 만들고 handoff에 경로를 알린다. 이는 design §3.2 순서의 run 구체화이며 파일에 비밀이 없어야 한다. `os.Root`로 경로를 고정하고 파일·디렉터리 교체와 symlink를 판별해 소유가 바뀌면 삭제를 거절한다. RemoveAll은 쓰지 않는다.

## Evidence

초기 신규 타입 RED, `go test ./internal/gateway/... -run TestSupervisor -timeout 45s`, exit 1의 관측 발췌:

```text
internal/gateway/supervisor_test.go:23:11: undefined: ReadChildConfig
internal/gateway/supervisor_test.go:25:6: undefined: RunChild
internal/gateway/supervisor_test.go:29:48: undefined: StartOptions
internal/gateway/supervisor_test.go:29: undefined: ChildConfig
internal/gateway/supervisor_test.go:30:36: undefined: ChildProcess
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
```

처음 sandbox 실행은 homestate 지문 조회가 차단되어 아래처럼 실패했다. handoff 실패 테스트의 pipe close 누락 때문에 이어진 테스트는 외부 Go timeout으로 끝났다. 이 실행은 검증 통과가 아니다. pipe close를 보강한 뒤 실제 프로세스·loopback 시험을 sandbox 밖에서 실행했다.

```text
--- FAIL: TestSupervisorChildStartsAndStops (0.00s)
    supervisor_test.go:73: gateway child configuration invalid
--- FAIL: TestSupervisorRunnerLifetimeClosesPortAndOwnedOverlay (0.00s)
    supervisor_test.go:138: parent fingerprint unavailable
panic: test timed out after 45s
```

추가 회귀를 RED로 관측하고 수리했다.

1. stdin을 읽지 않는 child: `TestSupervisorBlockedInputStartupDeadline`은 `panic: test timed out after 8s`로 실패했다. stack에는 supervisor.go:184의 Stop pipe Write 대기가 나타났다. Stop의 쓰기를 별도 goroutine으로 두고 Kill 이후 writer와 reaper를 모두 join하도록 수리했다. 테스트 helper 자체의 대기 시간도 한정했다. 후속 `pgrep -fl 'gateway.test.*TestSupervisorChildHelper'`는 exit 1, 출력 없음이었다.
2. 디렉터리 교체: 아래 RED 뒤 root 디렉터리 identity도 비교하도록 수리했다.

```text
--- FAIL: TestSupervisorOverlayRejectsDirectoryReplacement (0.00s)
    supervisor_test.go:341: replacement directory removed as owned
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.481s
```

3. 부모가 제기한 config reader prefetch와 취소된 Stop의 reaper join 가설을 각각 시험했다.

```text
--- FAIL: TestSupervisorConfigPreservesCoalescedStopByte (0.00s)
    supervisor_test.go:443: configuration reader consumed stop byte
--- FAIL: TestSupervisorCancelledStopStillJoinsReaper (1.10s)
    supervisor_test.go:469: Stop returned before process reaper completed
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	1.561s
```

ReadChildConfig를 정확히 newline까지만 읽도록 바꾸고 Stop 취소 분기도 done을 join했다. 같은 두 시험 GREEN:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.513s
```

4. bind·Serve 실패 시험의 internal listener seam을 테스트 우선으로 만들었다. RED:

```text
internal/gateway/supervisor_test.go:486:12: undefined: runChildWithListener
internal/gateway/supervisor_test.go:496:12: undefined: runChildWithListener
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
```

이후 통과한 최종 시험에는 위 회귀가 모두 포함된다. 다른 worker의 translate 패키지 RED를 덮거나 판정을 섞지 않기 위해 최종 범위는 gateway와 auth 두 패키지로 고정했다.

측정 작업 디렉터리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway ./internal/gateway/auth -coverpkg=./internal/gateway,./internal/gateway/auth -coverprofile=/tmp/gateway-supervisor-final-cover.out -timeout 60s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	7.333s	coverage: 93.5% of statements in ./internal/gateway, ./internal/gateway/auth
	github.com/modu-ai/moai-adk/internal/gateway/auth		coverage: 0.0% of statements
```

합산 profile의 블록별 statement 수를 파일별로 모은 실제 출력:

```text
github.com/modu-ai/moai-adk/internal/gateway/supervisor.go 100 107 93.5
github.com/modu-ai/moai-adk/internal/gateway/supervisor_posix.go 1 1 100.0
github.com/modu-ai/moai-adk/internal/gateway/supervisor_runner.go 91 105 86.7
```

각 행은 covered statements, total statements, percent 순이다. Windows 파일은 이 macOS 측정의 분모에 없으며 Windows 런타임 coverage를 주장하지 않는다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway ./internal/gateway/auth -race -timeout 60s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	9.619s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

같은 환경 정리와 cache로 `go vet ./internal/gateway ./internal/gateway/auth`를 실행했다: exit 0, 출력 없음. 최종 `gopls check`는 supervisor.go, supervisor_runner.go, supervisor_posix.go 세 파일 모두 exit 0, 출력 없음이었다. 구현 전 LSP baseline을 확보하지 않았으므로 전후 차이의 주장은 하지 않는다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOCACHE=/tmp/gateway-foundation-cache go test -c ./internal/gateway -o /tmp/gateway-supervisor-windows.test.exe
```

exit 0, 출력 없음. 평범한 supervisor_test.go가 Windows 시험 바이너리에 컴파일되었다. integration 태그나 t.Skip은 없으며 symlink 생성 권한이 없는 Windows에서는 regular replacement로 같은 identity 거절 계약을 시험한다. Windows에서 실행한 것은 아니다.

### 실제 시험이 단언한 범위

- 별도 시험 바이너리 child가 실제 loopback에 bind하고 전달한 주소에서 local-child HTTP 응답을 반환했다.
- startup EOF·잘못된 handoff·시간 초과·읽지 않는 stdin을 명시 실패로 처리하고 child를 회수했다.
- 취소·수명 종료·실제 부모 프로세스 사망·지문 교체 후 이전 포트가 연결 오류를 내는지 검사했다.
- 정상 취소·수명 종료·handoff 실패·Serve 실패 후 child 소유 overlay가 지워지는지 검사했다.
- symlink·regular replacement·디렉터리 교체 시 unrelated target 또는 대체 디렉터리를 보존했다.
- 종료 시험의 handler 호출 계수는 0이며 handler 내부에서 upstream을 만들지 않았다. 실제 provider endpoint를 호출하지 않았다.

## Baseline-attribution

이번 실행에서 다시 읽은 HEAD는 `81c1d58f9`, 브랜치는 `WT-unified-gateway`다. 부모가 다른 worker에 맡긴 SPEC·translate 파일을 보존했다. 이번 소유 파일은 supervisor.go, supervisor_runner.go, supervisor_posix.go, supervisor_windows.go, supervisor_test.go와 이 보고서 및 progress Run 절이다. CLI·기존 launcher·제품 설정·사용자 credential 저장소는 수정하지 않았다. 커밋과 status 전이는 하지 않았다.

## Gaps

CLI 내부 동사 연결, 실제 Claude exec 후 PID·지문 유지, signal·PTY·job control 보존, Windows 런타임, 실제 OAuth/API credential 전달, 전체 AC-MG-006/016/017은 미완료다. M4 전체 PASS가 아니다. 저장소 전체 시험 판정은 통합 브랜치 CI 소관이며 PENDING이다. Windows 증거는 design §3.5대로 release PR의 windows-latest 이벤트 스트림에서 확인해야 한다.

## Residual-risk

overlay는 child가 만든 private 디렉터리를 다른 작성자가 동시에 바꾸지 않는 독점 소유 계약이다. 관측 가능한 파일·디렉터리 교체는 거절하지만 같은 UID의 악의적 동시 변경에 대한 원자적 compare-and-unlink를 주장하지 않는다. 강제 OS 종료 또는 Kill이 overlay 생성과 인계 사이에 발생하면 child가 자체 정리를 실행하지 못할 수 있다. 살아 있는 부모의 정상 Stop·취소·기한 종료 경로는 위 시험으로 검증했다.

프로세스 지문 품질과 probe의 OS 동작은 기존 homestate 계약에 따른다. PID-only로 완화하지 않았으며 caller가 정할 제품 lifetime·poll interval과 CLI에 공급할 최소 환경/비밀 없는 overlay 내용은 후속 연결 단계의 책임이다. stdin Payload를 읽는 것은 이 supervisor지만 실제 credential 해석과 outbound 송신은 handler/adapter의 소관이다.
