# GPT AUTH 로컬 구현 검증

## Claim

AUTH 계획 감사 PASS 1.00 범위에서 macOS/Linux용 MoAI 소유 인증 저장소, Codex app-server 인증 wrapper,
세대 기반 credential 참조 및 구독 송신 경계를 구현했다. 이 보고서는 **로컬 wrapper 검증**이며 실제 로그인,
실제 refresh, 네 GPT 모델의 Claude Code 통합, Windows 제품 지원 또는 AUTH 전체 완료 판정이 아니다.

구현 파일은 `internal/gateway/auth/`에만 있다. 기존 `CredentialRef` 네 메서드 계약을 유지했다. 기존 사용자 Codex
인증 파일이나 Claude 전역 설정을 읽지 않았으며 설치 Codex를 실행하지 않았다. fake broker는 이 패키지의 Go test
실행 파일을 별도 OS 프로세스로 실행한다. 테스트의 모든 token과 계정 ID는 합성 표지다.

## 구현 범위와 API

- `OpenStore(privateDir)`, `Store.Close`, `Status(ctx)`, `Resolve`: 절대 경로의 private 저장소만 허용한다.
  symlink 조상·공개 권한·손상 상태를 거절한다. 상태 조회는 비밀 없는 세대·로그인 여부·만료·방식만 반환한다.
- `Store.Login(ctx, Broker)`, `Refresh(ctx, expectedGeneration, Broker, verify)`: operation lock과 짧은 state lock을
  구분한다. broker 종료 후 private 파일·chatgpt 방식·필수 token·JWT 만료를 검증하고 시작 세대가 같은 경우에만
  원자 publish한다. API 키 fallback seed를 거절한다. token이나 account를 오류 문자열로 반환하지 않는다.
- `Store.Logout(ctx)`: operation lock을 기다리지 않고 tombstone과 증가 세대를 먼저 기록한다. **로컬 폐기 API**다.
  이 API는 remote revoke를 실행하거나 성공했다고 표시하지 않는다.
- `CodexBroker{Executable, Timeout, OnLogin}`: 새 scratch CODEX_HOME/CWD/HOME/XDG 경로와 제한된 환경을 구성하고
  file credential backend를 명시한다. initialize → chatgpt login → 정확한 loginId 완료 notification을 검사한다.
  refresh는 account/read 응답만 반환하며, 그 응답만으로 갱신 성공을 주장하지 않는다. stderr는 폐기한다.
  종료 코드·종료 대기와 scratch cleanup 성공을 확인한 뒤에만 저장할 수 있다.
- `Refresh`의 verifier는 후보 `CredentialRef`에 대한 공급자 수용 확인을 담당한다. 접근 token 변경, 만료 연장,
  동일 계정과 verifier 성공이 모두 필요하다. 현재 시험의 verifier는 mock이며 실제 provider verifier는 연결 전이다.
- `SendAuthorized(ctx, expected, request, *http.Transport, SendOptions)`: 정확한 구독 endpoint의 POST만 허용한다.
  입력은 GetBody/ContentLength가 있는 유한 body이며 MaxBodyBytes, WriteTimeout, PollInterval은 양수여야 한다.
  HTTP/1.1 전용 연결에서 TLS 완료 뒤 plaintext `net.Conn.Write`가 실제로 수락한 바이트의 CRLFCRLF를 확인할 때
  state lock을 해제한다. request header write deadline을 적용하며, 연결 실패/timeout에서는 연결을 닫고 writer를
  기다린다. response SSE 전체에는 lock을 유지하지 않는다. 세대 polling으로 다른 프로세스 logout 뒤 in-flight
  요청을 취소한다. 응답 body Close/EOF에서 watcher를 종료·회수한다.
- HTTP/2·proxy·redirect는 이 송신 API에서 사용하지 않는다. TLS 검증 기본값을 유지하며 허용 endpoint는
  `https://chatgpt.com/backend-api/codex/responses`의 정확 일치다. 자체 Transport clone과 새 연결로 이전 요청의
  header 완료를 현재 요청의 증거로 오인하지 않는다.
- `NewAPIKey(explicitKey)`는 별도 `openai:api-key` 참조이며 API endpoint에만 적용된다. 환경을 읽거나 구독 실패의
  fallback으로 동작하지 않는다. API 키 로그인 저장·명령·실서비스 호출이 구현되었다는 의미는 아니다.
- Windows `LockFileEx` 구현은 컴파일되지만 ACL·내구성 저장 구현은 아직 없다. Windows `OpenStore`는
  `ErrPlatformUnsupported`로 처음부터 거절한다. 이는 안전한 임시 미지원이며 Windows 동작 PASS가 아니다.

## Evidence

### 실제 RED → GREEN 기록

첫 저장소 시험을 구현 전에 실행했다.

```text
$ go test ./internal/gateway/auth -count=1
internal/gateway/auth/store_test.go:24:35: undefined: Store
internal/gateway/auth/store_test.go:31:9: undefined: OpenStore
FAIL github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]
```

첫 sender와 broker 시험에서도 구현 부재를 확인했다.

```text
$ go test ./internal/gateway/auth -run TestSend -count=1
internal/gateway/auth/send_test.go:33:14: s.SendAuthorized undefined (type *Store has no field or method SendAuthorized)
internal/gateway/auth/send_test.go:33:65: undefined: SendOptions
FAIL github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]

$ go test ./internal/gateway/auth -run TestCodexBroker -count=1
internal/gateway/auth/broker_test.go:74:9: undefined: CodexBroker
internal/gateway/auth/broker_test.go:74:94: undefined: LoginPrompt
FAIL github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]
```

위 발췌의 줄 번호는 각 RED 당시 시험 파일 기준이다. 컴파일 오류 전체 출력에는 같은 미정의 심볼을 참조한
추가 줄도 있었으며, 최종 파일의 줄 번호를 뜻하지 않는다.

httptrace.WroteRequest를 실제 write 경계로 쓰는 초기 구현은 실제 net.Pipe 쓰기를 막은 시험에서 실패했다.

```text
$ go test ./internal/gateway/auth -run TestSend -count=1 -timeout=15s
--- FAIL: TestSendWriteBarrierBlocksLogoutButSSEDoesNot (0.03s)
    send_test.go:65: logout crossed blocked socket write: <nil>
panic: test timed out after 15s
    running tests:
        TestSendTimeoutCancelsBlockedSocketAndReleasesLock (15s)
```

스택은 `net/http.(*persistConn).writeLoop`의 `bufio.Writer.Flush`에서 실제 Write가 막혀 있음을 보였다.
이를 근거로 callback 방식을 버리고 실제 plaintext Write 반환 바이트를 관측하도록 수정했다. 이 수리 직후 출력:

```text
$ go test ./internal/gateway/auth -run TestSend -count=1 -timeout=15s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.628s
```

다음 음성 시험도 각각 실패를 먼저 관측한 뒤 수리했다.

```text
--- FAIL: TestCodexBrokerIsolatedLoginCompletionAndTermination/exit-failed (0.01s)
    broker_test.go:96: bad completion accepted: <nil>

--- FAIL: TestBrokerCannotReplaceScratchHomeWithSymlink (0.02s)
    boundary_test.go:217: replaced scratch accepted: <nil>

--- FAIL: TestBrokerCannotSeedAPIKeyFallback (0.00s)
    boundary_test.go:238: API fallback seed accepted: <nil>

--- FAIL: TestCodexBrokerRejectsSymlinkedEnvironmentDirectories (0.01s)
    broker_test.go:158: symlink config escape accepted

--- FAIL: TestScratchCleanupFailureDoesNotPublish (0.07s)
    boundary_test.go:334: cleanup failure reported login success
```

해당 마지막 수리는 private scratch cleanup을 canonical publish 전에 수행하여 실패 시 기존 세대를 보존한다.
로그인 취소·부분 파일·옛 세대 refresh·API endpoint 혼용·잘못된 RPC/loginId·소유하지 않은 scratch를 거절하는 시험이
최종 패키지 시험에 함께 들어 있다. 기존 `credential.go`의 네 메서드는 변경하지 않았다.

### 최종 변경 범위 검사

최종 소스에서 다음을 실행했다. 저장소 전체 suite는 실행하지 않았다.

```text
$ go test ./internal/gateway/auth -count=1 -coverprofile=/tmp/gateway-auth-cover.out -timeout=20s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	2.470s	coverage: 87.1% of statements

$ go test ./internal/gateway/auth -race -count=1 -timeout=30s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	7.984s

$ go vet ./internal/gateway/auth

$ /Users/goos/go/bin/gopls check internal/gateway/auth/credential.go internal/gateway/auth/store.go internal/gateway/auth/send.go internal/gateway/auth/broker.go internal/gateway/auth/api_key.go internal/gateway/auth/lock_posix.go

$ GOOS=windows GOARCH=amd64 go test ./internal/gateway/auth -c -o /tmp/gateway-auth-windows.test.exe
```

모두 exit 0. 마지막 세 명령의 stdout/stderr는 빈 출력이었다. Windows 항목은 test binary 컴파일만 의미한다.
Go 캐시 접근이 sandbox에서 거절된 vet/LSP/coverage/Windows compile 호출은 승인된 재실행에서 확인했다. 처음 LSP가
exit 0과 함께 packages.Load permission 오류를 출력한 결과는 PASS로 세지 않았다. 중간에 관측한 empty critical
section과 Scanner.Err 미확인 LSP 진단은 수정했으며 위 최종 LSP 출력은 비어 있다.

### 시험이 관측한 동시성 및 종료 경계

- 별도 OS 프로세스 두 개가 같은 세대 Refresh를 호출하고 파일에 기록한 token mock 교환 수가 정확히 1이었다.
- 실제 operation lock을 가진 별도 프로세스를 Kill/Wait한 뒤 다른 프로세스가 잠금을 획득했다.
- 별도 프로세스의 지연 refresh 중 logout이 state lock으로 tombstone을 먼저 기록했고, 늦은 refresh 결과는
  ErrCredentialChanged로 거절되어 미로그인 상태가 유지되었다.
- 실제 net.Pipe Write가 멈춰 있을 때 logout은 context deadline으로 실패했다. header 송신 뒤 끝나지 않는 SSE를
  열어 둔 대조군은 logout을 막지 않았고 다음 세대 확인에서 stream이 취소되었다.
- header delimiter를 여러 Write에 나누어 보냈을 때 마지막 바이트 수락 후 barrier가 한 번만 열렸으며 저장한 tail은
  비었다. request header보다 먼저 도착한 응답은 성공한 송신으로 반환되지 않았다.
- fake broker는 제한된 HOME/CODEX_HOME과 API/Claude 인증 환경 부재를 자체 검사했다. 정상 완료 후 subprocess 종료
  표지가 존재하는지 확인했다. 종료 오류·잘못된 loginId·실패/손상 응답은 publish 성공으로 처리하지 않았다.

## Baseline-attribution

모든 위 명령은 다음 WT에서 실행했다.

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
$ git rev-parse --short HEAD
81c1d58f9
$ git branch --show-current
WT-unified-gateway
```

AUTH 0.1.0과 `auth-plan-audit-iter1.md`의 제한된 PASS가 착수 기준이다. `internal/gateway/auth/`는 작업 중인
untracked 구현이며 commit·push·PR·병합을 수행하지 않았다. 부모와 다른 worker의 CLI/translator 파일을 수정하지 않았다.
설치 Go 테스트 toolchain은 timeout 스택에서 go1.26.8.darwin-arm64로 관측되었다. 전체 저장소 테스트 판정의 소유자는
통합 브랜치 CI이며, 이번에는 CI를 실행하지 않았으므로 해당 판정은 **PENDING**이다.

## Gaps

- 설치 Codex 0.154.0의 실제 app-server/RPC/file backend/관리 정책/키링 fallback 적용은 미실행이다. 공개 source
  snapshot과 설치본이 동일한 빌드라는 증거는 없다. wrapper mock은 Codex의 PKCE/state/callback 내부 시험이 아니다.
- 실제 새 구독 로그인·token refresh·provider 수용 verifier·원격 revoke는 미실행 또는 미연결이다. Logout은 로컬
  tombstone만 보장한다. late token 교환에 대한 서버 측 revoke를 수행했다고 주장하지 않는다.
- API 키 참조의 endpoint 구분은 로컬 요청 객체 시험이다. 두 endpoint의 실제 필수 header·지원 필드·과금 경로는
  실계정 preflight 대상이며 max_output_tokens/nonstream/구조 출력 호환을 가정하지 않는다.
- 실제 Claude Code, core adapter, CLI/PICKER와 네 GPT 모델의 대화·도구 왕복을 이 worker가 검증하지 않았다.
  사용자 지정 2026-09-11 19:00 KST 이전에 provider/Claude 실서비스 시험을 실행하지 않았다.
- Windows ACL 및 durable canonical store 구현과 Windows 런타임 시험이 없다. Linux에서는 실행하지 않았으며 이번
  POSIX 실행 증거는 macOS다. Windows test binary 컴파일을 기능 완료로 해석해서는 안 된다.
- 이전 세대의 in-flight 요청은 generation polling으로 취소를 시도한다. 송신된 바이트를 소급 미송신으로 만들지 않는다.
  프로세스 자체가 죽은 경우의 socket cleanup은 OS에 의존한다. 악의적인 비협조적 custom net.Conn까지 종료를 보장하는
  범용 transport wrapper가 아니며 실제 클라이언트는 trusted transport 설정을 넘겨야 한다.
- 구현 첫 파일을 쓰기 전의 전체 패키지 LSP baseline은 별도로 보존하지 못했다. 위 기록은 최종 진단 0의 증거이며
  미측정 사전 baseline 대비 회귀 0이라는 주장은 아니다.

## Residual-risk

실제 broker가 현재 후보와 다른 token 형식·RPC 순서·URL을 사용하면 명시 실패할 수 있다. 같은 사용자 권한을 가진
악의적인 프로세스의 모든 filesystem TOCTOU 공격에 대한 원자적 방어를 증명하지 않았다. 비정상적으로 이름을 바꾼
scratch는 canonical로 채택하지 않지만 이동된 디렉터리의 임의 위치를 추적하여 삭제하지는 않는다. 일반 cleanup 실패는
성공 publish 전에 거절한다. 디렉터리 fsync가 rename 뒤 실패하면 오류를 반환하며 디스크 상태를 다시 읽어야 한다.
`Store.Resolve`/`CredentialRef.Apply`만 호출하는 integration은 logout 송신 경계를 충족하지 않는다. 실제 core adapter가
반드시 SendAuthorized를 호출한 통합 시험이 끝나기 전 AUTH 전체 완료로 표시할 수 없다.
