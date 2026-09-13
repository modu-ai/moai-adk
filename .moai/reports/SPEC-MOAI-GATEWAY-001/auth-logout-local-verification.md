# AUTH B.5 로그아웃 snapshot broker — 로컬 구현 검증

## Claim

`Store.LogoutWithBroker`와 선택적 `LogoutBroker.Logout(ctx, home)` 계약을 추가했다. 기존 `Store.Logout(ctx) (uint64,error)`는 로컬 전용 API로 유지한다. 두 API는 짧은 state lock 아래 tombstone과 증가 세대를 먼저 게시하는 공통 함수를 사용한다. broker 작업은 operation lock을 기다리지 않는다.

로컬 게시에 성공한 경우에만 게시 전 MoAI canonical snapshot을 별도 `logout-*` private scratch에 복사한다. broker에 canonical 경로나 다른 사용자의 CODEX_HOME을 넘기지 않으며, broker의 종료 후 scratch를 정리한다. remote 실패·취소·완료 뒤 canonical credential을 다시 쓰는 코드는 없다. 이전 snapshot broker가 끝나기 전에 새 로그인이 완료돼도 새 canonical 상태는 보존한다.

`CodexBroker.Logout`은 기존 process isolation·timeout·stdout join·Wait/reaping 처리를 재사용하여 initialize → initialized → account/logout을 수행한다. RPC ID와 빈 object result를 확인하며 오류·잘못된 ID·null result를 거절한다. Store가 전달하는 broker deadline은 최대 30초다. 주입된 LogoutBroker 구현 역시 context 취소를 따르고 종료가 확인된 뒤 반환해야 한다는 계약을 명시했다.

CLI는 `local logout completed`, `broker logout completed/failed/unavailable/not_needed`, `remote revocation unknown/not_attempted`를 분리한다. 원격 철회 성공으로 표시하는 경로는 없다. broker 오류 본문이나 token 값은 출력하지 않는다. scratch 준비·정리 같은 로컬 후처리 오류가 나도 이미 완료된 로컬 폐기를 표시하고 해당 오류를 반환한다.

## Evidence

### 공개 소스 확인

부모가 지정한 감사자 `/root/gateway_auth_supervisor_audit`의 소스 결과를 받은 뒤 같은 공개 checkout `/tmp/openai-codex-audit.zbvNXB`의 실제 본문을 sed/rg로 읽었다. 기준은 `5a9eb14`이다.

- `codex-rs/app-server/src/request_processors/account_processor.rs:965`: `logout_with_revoke()` 호출.
- `codex-rs/login/src/auth/manager.rs:2914–2931`: revoke_auth_tokens 오류는 warn으로 남긴 뒤 logout_all_stores 및 reload를 수행하고 로컬 결과를 반환한다.
- `codex-rs/login/src/auth/revoke.rs:55–95`: managed Chatgpt 자격에 대해 refresh token을 우선하고 access token으로 대체하는 철회 시도.
- 같은 파일 110–121행: JSON POST 및 2xx 판정. 23행 timeout 10초.
- `manager.rs:201`: `https://auth.openai.com/oauth/revoke`.
- `account_processor.rs:1005–1006`: 성공 RPC 응답은 `LogoutAccountResponse {}`로, 원격 철회 결과를 담지 않는다.

```text
$ rg -n 'REVOKE_HTTP_TIMEOUT|DEFAULT.*REVOKE|oauth/revoke' /tmp/openai-codex-audit.zbvNXB/codex-rs/login/src/auth/revoke.rs /tmp/openai-codex-audit.zbvNXB/codex-rs/login/src/auth/manager.rs
[관련 출력]
/tmp/openai-codex-audit.zbvNXB/codex-rs/login/src/auth/revoke.rs:23:const REVOKE_HTTP_TIMEOUT: Duration = Duration::from_secs(10);
/tmp/openai-codex-audit.zbvNXB/codex-rs/login/src/auth/manager.rs:201:pub(super) const REVOKE_TOKEN_URL: &str = "https://auth.openai.com/oauth/revoke";
```

따라서 공식 broker는 원격 철회를 **시도**하지만 RPC 완료만으로 원격 성공·실패를 구별할 수 없다. unknown이 맞으며 unsupported라고 단정하지 않는다. MoAI에 별도 endpoint/client 등록이나 자체 revoke HTTP 구현을 추가하지 않았다.

### RED → GREEN

모든 명령의 작업 디렉터리는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`이다.

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/auth -run '^TestLogout' -count=1
# github.com/modu-ai/moai-adk/internal/gateway/auth [github.com/modu-ai/moai-adk/internal/gateway/auth.test]
internal/gateway/auth/logout_test.go:22:11: s.LogoutWithBroker undefined (type *Store has no field or method LogoutWithBroker)
internal/gateway/auth/logout_test.go:37:11: s.LogoutWithBroker undefined (type *Store has no field or method LogoutWithBroker)
internal/gateway/auth/logout_test.go:53:15: undefined: LogoutBroker
internal/gateway/auth/logout_test.go:56:13: s.LogoutWithBroker undefined (type *Store has no field or method LogoutWithBroker)
FAIL	github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]
FAIL

$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/auth -run '^TestCodexBrokerLogout' -count=1
# github.com/modu-ai/moai-adk/internal/gateway/auth [github.com/modu-ai/moai-adk/internal/gateway/auth.test]
internal/gateway/auth/logout_broker_test.go:50:9: b.Logout undefined (type CodexBroker has no field or method Logout)
FAIL	github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]
FAIL

$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/cli -run '^TestGPTLogout|^TestGPTInstalledLogout' -count=1
--- FAIL: TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote (0.09s)
    --- FAIL: TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote/completed (0.07s)
        gpt_auth_logout_test.go:40: 0 GPT local logout completed; remote revocation was not requested

    --- FAIL: TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote/failed (0.03s)
        gpt_auth_logout_test.go:40: 0 GPT local logout completed; remote revocation was not requested

FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.151s
FAIL

$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/auth -run '^TestLogout|^TestCodexBroker' -count=1 -timeout 20s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.001s

$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/cli -run '^TestGPTLogout|^TestGPTInstalledLogout|^TestGPTLogin|^TestGPTStatus' -count=1 -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/cli	1.245s
```

CLI fixture는 macOS `/var` → `/private/var` canonicalization 때문에 최초 경로 assertion이 실패하여 테스트 기대 경로를 EvalSymlinks로 정규화했다. 기존 제품의 symlink 정책은 바꾸지 않았다. fake broker subprocess에는 기존 시험처럼 `GORACE=atexit_sleep_ms=0`을 넣었다. race detector의 종료 대기만 제거하며 race 검출은 유지한다.

최종 검증:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway/auth -run '^TestLogout|^TestCodexBroker|^TestStore|^TestRefresh|^TestBrokerRegression' -count=1 -coverprofile=/tmp/gateway-logout-auth-cover.out -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	2.616s	coverage: 52.3% of statements

$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/cli -run '^TestGPTLogout|^TestGPTInstalledLogout|^TestGPTLogin|^TestGPTStatus' -count=1 -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/cli	2.550s

$ GOCACHE=/tmp/gateway-translation-cache go vet ./internal/gateway/auth ./internal/cli
[exit 0, stdout/stderr 비어 있음]

$ GOCACHE=/tmp/gateway-translation-cache go tool cover -func=/tmp/gateway-logout-auth-cover.out | rg 'logout.go|logoutSnapshot|broker.go.*run'
github.com/modu-ai/moai-adk/internal/gateway/auth/broker.go:56:			run			85.3%
github.com/modu-ai/moai-adk/internal/gateway/auth/logout.go:25:			LogoutWithBroker	75.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/store.go:340:			logoutSnapshot		92.3%
```

시험은 operation lock을 이미 잡은 상태에서도 local tombstone 게시 후 broker가 시작됨, local 실패 시 broker 0회, 이전 snapshot만 전달됨, private directory/file 권한, broker 실패·취소 후 tombstone 유지 및 scratch 제거, 오래된 logout 작업 중 새 login 세대 보존, RPC 오류/null/잘못된 ID 거절을 확인한다. fake subprocess timeout 후 Signal(0)으로 살아 있는 자식이 없음을 확인한다. CLI 시험은 기존 사용자 CODEX_HOME 파일이 보존되고 private 문자열이 출력되지 않는 것도 확인한다.

## Baseline-attribution

```text
$ git rev-parse --short HEAD
81c1d58f9
$ git branch --show-current
WT-unified-gateway
$ shasum -a 256 [아래 일곱 파일]
d745c35a81616285dc19af51c4d9f5aa90ba6dffef4702e8a417527b49629836  internal/gateway/auth/logout.go
b3b16d81ed69b9833ee6a60a3cd4ea2554cfde9412fa105a16f1aea522ec7463  internal/gateway/auth/logout_test.go
edbafa7009e533a02d3d124d4e3bcc97629e4a9422de1d5f5634668cf1d021f4  internal/gateway/auth/logout_broker_test.go
1a5bb5507ec1a732dc50f2a2b2aabc048d9bdf9ab2631fc915f499c6ab5ea1c2  internal/gateway/auth/store.go
37c0963550cdeb389dd60e21c7639a8753859423a426ae932fa1b3319c563481  internal/gateway/auth/broker.go
955d40662bfd95b89b86bcb58612441dbf9106728cf0ab13d17518bfa7c439c3  internal/cli/gpt_auth.go
9fecb300acedf4302666b206138813b394a3d0776dd7b114b7b2e6b497cbb7d4  internal/cli/gpt_auth_logout_test.go
```

위 파일과 본 보고서만 이번 위임 범위에서 수정했다. Windows helper, 다른 AUTH writer의 resolve 파일, CI, launcher, 실제 AUTH live harness는 수정하지 않았다. 실제 Codex/broker/provider를 실행하지 않았으며 모든 자격은 합성 fixture다. commit/push는 수행하지 않았다.

## Gaps

설치본이 공개 checkout과 동일하게 원격 철회를 수행하는지, 실제 토큰이 서버에서 철회됐는지는 미검증이다. revoked token을 재사용해 확인하는 live 시험도 하지 않았다. 로컬 새 login 보존 시험은 canonical 파일과 세대의 격리를 확인할 뿐 제공자의 token-family 철회 영향까지 보장하지 않는다.

첫 측정의 LogoutWithBroker coverage는 75%였고 아래 후속 영구 회귀를 포함한 재측정은 81.2%다. 85% 목표에는 여전히 미달하며, scratch 생성/seed 쓰기 실패 및 local 게시 직후 취소 분기의 fault injection 증거는 없다. scratch 정리 실패는 아래 두 경우에서 실제로 관측했다. 따라서 전체 품질 완료로 주장하지 않는다. 공개 broker process timeout/reap은 합성 subprocess에서만 확인했다. Windows 실행은 이번 범위 밖이다. LSP baseline은 없으며 go vet만 실행했다. develop 통합 브랜치 CI가 저장소 전체 시험 판정의 소유자이며 보고 시점 PENDING이다.

## Residual-risk

원격 철회를 숨기는 빈 RPC 응답 때문에 broker 완료와 remote unknown을 계속 분리해야 한다. 호출자가 broker outcome만 보고 원격 철회 성공으로 확대하면 안 된다. 로컬 tombstone 게시 뒤 scratch 준비나 cleanup이 실패하면 자격을 복원하지 않으며 후처리 실패를 반환한다. 주입한 임의 LogoutBroker가 취소/종료 계약을 어기면 Store가 그 구현을 강제로 종료할 수는 없다. 실제 제품 경로는 deadline·kill·Wait/join을 구현한 CodexBroker를 사용한다.

## 후속: scratch 정리 실패 영구 회귀

### Claim

독립 감사의 `logout-functional-review.md`와 ignored `logout-functional-review/independent_test.go`의 실제 probe를 읽고 `internal/gateway/auth/logout_fault_test.go`로 옮겼다. 시험이 소유한 scratch 디렉터리만 mode 000으로 변경하여 RemoveAll 실패를 유도하고, `t.Cleanup`에서 mode 복구·제거를 반드시 수행한다. tombstone 상태와 별도 Store handle의 새 login 상태 두 경우 모두 실제 실패를 관측했다. 제품 파일은 수정하지 않았다.

### Evidence

```text
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway/auth -run '^TestLogoutScratchCleanupFailure' -count=1 -v -timeout 10s
=== RUN   TestLogoutScratchCleanupFailurePreservesCanonical
=== RUN   TestLogoutScratchCleanupFailurePreservesCanonical/tombstone
    logout_fault_test.go:98: actual scratch cleanup failure returned local completed, cleanup_failed, remote unknown; canonical bytes preserved
=== RUN   TestLogoutScratchCleanupFailurePreservesCanonical/newer_login
    logout_fault_test.go:98: actual scratch cleanup failure returned local completed, cleanup_failed, remote unknown; canonical bytes preserved
--- PASS: TestLogoutScratchCleanupFailurePreservesCanonical (0.07s)
    --- PASS: TestLogoutScratchCleanupFailurePreservesCanonical/tombstone (0.04s)
    --- PASS: TestLogoutScratchCleanupFailurePreservesCanonical/newer_login (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.468s

$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway/auth -run '^TestLogout|^TestCodexBroker|^TestStore|^TestRefresh|^TestBrokerRegression' -count=1 -coverprofile=/tmp/gateway-logout-fault-cover.out -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	2.644s	coverage: 52.6% of statements

$ GOCACHE=/tmp/gateway-translation-cache go vet ./internal/gateway/auth
$ gofmt -l internal/gateway/auth/logout_fault_test.go
[각 exit 0, 출력 없음]

$ GOCACHE=/tmp/gateway-translation-cache go tool cover -func=/tmp/gateway-logout-fault-cover.out | rg 'logout.go|logoutSnapshot|broker.go.*run'
github.com/modu-ai/moai-adk/internal/gateway/auth/broker.go:56:			run			85.3%
github.com/modu-ai/moai-adk/internal/gateway/auth/logout.go:25:			LogoutWithBroker	81.2%
github.com/modu-ai/moai-adk/internal/gateway/auth/store.go:340:			logoutSnapshot		92.3%
```

둘 다 플랫폼 skip 없이 수행했다. 반환은 `ErrAuthState`, `LocalCompleted=true`, `BrokerOutcome=cleanup_failed`, `RemoteOutcome=unknown`이다. state.json의 실패 전후 bytes와 세대·LoggedIn 상태를 대조해 tombstone 또는 새 login이 보존됨을 확인했다. 실패한 scratch 경로가 남았음을 확인한 뒤 테스트 cleanup으로 제거했다.

### Baseline-attribution

```text
$ git rev-parse --short HEAD
81c1d58f9
$ shasum -a 256 internal/gateway/auth/logout.go internal/gateway/auth/store.go internal/gateway/auth/broker.go internal/gateway/auth/logout_fault_test.go
d745c35a81616285dc19af51c4d9f5aa90ba6dffef4702e8a417527b49629836  internal/gateway/auth/logout.go
1a5bb5507ec1a732dc50f2a2b2aabc048d9bdf9ab2631fc915f499c6ab5ea1c2  internal/gateway/auth/store.go
37c0963550cdeb389dd60e21c7639a8753859423a426ae932fa1b3319c563481  internal/gateway/auth/broker.go
725f08bc6e09fd320b5d6252222f2758425c01204f3aa70bb7dd56a77c619305  internal/gateway/auth/logout_fault_test.go
```

세 제품 파일 해시는 독립 감사의 동결 기준과 같다. 후속 변경은 새 시험 파일과 본 보고서뿐이다.

### Gaps

scratch 생성과 auth.json seed 쓰기 실패는 주입하지 않았다. 코드에서 로컬 원자 게시와 scratch 생성이 같은 Store 경로를 사용하므로 미리 부모 경로의 쓰기 권한을 제거하면 원하는 후처리 실패 전에 로컬 게시 자체가 실패한다. seed 쓰기는 MkdirTemp 직후 broker callback 이전에 실행된다. 이 둘을 결정적으로 겨냥하려면 제품 hook이나 타이밍에 의존하는 파일시스템 경합이 필요하며, 이번 위임은 그런 장치를 추가하지 않도록 한정했다. 이 판단은 실패 경로가 검증됐다는 주장이 아니다.

### Residual-risk

81.2%는 함수의 현재 실측 coverage이며 85% 목표 PASS로 바꾸지 않는다. mode 000에도 제거가 가능한 특권 플랫폼에서는 이 fault 시험이 skip하며, skip 결과는 cleanup 실패를 관측한 근거가 아니다. 실제 broker·원격 제공자·Windows는 이번 후속에서도 실행하지 않았다.
