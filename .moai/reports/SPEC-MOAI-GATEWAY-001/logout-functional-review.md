# AUTH B.5 logout 변경분 독립 기능 검증

## Claim

**Scoped functional verdict: PASS.** 이번 일반 logout 트랜잭션·RPC 종료·CLI 결과 표기 범위에서 새 blocking finding은 관측하지 않았다. 전체 AUTH 품질·보안·실제 원격 revoke 또는 SPEC 완료 판정이 아니다.

작성자의 제품 동결 신호와 `auth-logout-local-verification.md`를 받은 뒤 기준 파일 7개의 해시를 채집하고 다음을 직접 시험했다.

- operation.lock을 이미 보유한 상태에서도 tombstone·증가 세대를 먼저 게시하고 broker를 호출한다. 실패한 broker 결과가 인증을 복원하지 않는다.
- 독립 probe에서 두 Store handle을 열었다. 첫 handle의 broker를 정지시킨 동안 두 번째 handle이 tombstone을 관측하고 새 로그인을 완료했다. 늦은 broker 성공·실패 모두 새 세대 3의 canonical bytes를 바꾸지 않았다. 각 scratch는 제거됐다.
- 정상 `account/logout`의 빈 `{}` 결과는 수락하고 RPC 오류·잘못된 ID·null은 거절한다. 독립 합성 subprocess가 `{}`를 보낸 뒤 exit 7 또는 hang하면 `ErrBroker`로 끝난다. 영구 시험은 timeout 뒤 process가 살아 있지 않음을 확인한다.
- 취소 후 tombstone과 scratch cleanup을 유지한다. 추가 독립 probe에서는 scratch 디렉터리 mode를 000으로 바꾸어 실제 RemoveAll 실패를 유발했다. `LocalCompleted=true`, `BrokerOutcome=cleanup_failed`, `RemoteOutcome=unknown`, `ErrAuthState`가 반환되고 tombstone은 유지됐다. 테스트 cleanup에서 권한을 복구하고 남은 경로를 제거했다.
- CLI는 local completed·broker completed/failed·remote unknown을 분리한다. installed broker 조회는 바이너리 부재를 오류로 반환한다. 원격 성공으로 바꾸어 표시하지 않는다.

## Evidence

모든 명령의 작업 디렉터리는 아래 WT다. 합성 자격과 임시 디렉터리·Go test helper subprocess만 사용했다. 실제 Codex/Claude/provider·외부 네트워크를 실행하지 않았다.

독립 probe와 관련 영구 시험:

```text
$ GOCACHE=/tmp/gateway-auth-audit-cache go test -race -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/logout-functional-review/overlay.json ./internal/gateway/auth -run '^TestLogoutFunctional|^TestLogoutOwned|^TestLogoutOld|^TestLogoutLocal|^TestLogoutCanceled|^TestCodexBrokerLogoutRPC' -count=1 -v -coverprofile=/tmp/gateway-logout-functional.cover -timeout 20s
=== RUN   TestCodexBrokerLogoutRPCAndReaping
=== RUN   TestCodexBrokerLogoutRPCAndReaping/completed
=== RUN   TestCodexBrokerLogoutRPCAndReaping/error
=== RUN   TestCodexBrokerLogoutRPCAndReaping/wrong
=== RUN   TestCodexBrokerLogoutRPCAndReaping/null
=== RUN   TestCodexBrokerLogoutRPCAndReaping/timeout
--- PASS: TestCodexBrokerLogoutRPCAndReaping (0.34s)
    --- PASS: TestCodexBrokerLogoutRPCAndReaping/completed (0.03s)
    --- PASS: TestCodexBrokerLogoutRPCAndReaping/error (0.04s)
    --- PASS: TestCodexBrokerLogoutRPCAndReaping/wrong (0.03s)
    --- PASS: TestCodexBrokerLogoutRPCAndReaping/null (0.02s)
    --- PASS: TestCodexBrokerLogoutRPCAndReaping/timeout (0.21s)
=== RUN   TestLogoutFunctionalConcurrentSecondHandlePreservesNewCanonical
=== RUN   TestLogoutFunctionalConcurrentSecondHandlePreservesNewCanonical/false
    logout_functional_test.go:15: second handle saw tombstone, new generation 3 survives delayed broker completed byte-for-byte; scratch removed
=== RUN   TestLogoutFunctionalConcurrentSecondHandlePreservesNewCanonical/true
    logout_functional_test.go:15: second handle saw tombstone, new generation 3 survives delayed broker failed byte-for-byte; scratch removed
--- PASS: TestLogoutFunctionalConcurrentSecondHandlePreservesNewCanonical (0.13s)
    --- PASS: TestLogoutFunctionalConcurrentSecondHandlePreservesNewCanonical/false (0.07s)
    --- PASS: TestLogoutFunctionalConcurrentSecondHandlePreservesNewCanonical/true (0.06s)
=== RUN   TestLogoutFunctionalExitHelper
--- PASS: TestLogoutFunctionalExitHelper (0.00s)
=== RUN   TestLogoutFunctionalEmptySuccessStillRequiresCleanExit
=== RUN   TestLogoutFunctionalEmptySuccessStillRequiresCleanExit/exit7
    logout_functional_test.go:18: empty RPC success followed by exit7 returned ErrBroker
=== RUN   TestLogoutFunctionalEmptySuccessStillRequiresCleanExit/hang
    logout_functional_test.go:18: empty RPC success followed by hang returned ErrBroker
--- PASS: TestLogoutFunctionalEmptySuccessStillRequiresCleanExit (0.31s)
    --- PASS: TestLogoutFunctionalEmptySuccessStillRequiresCleanExit/exit7 (0.03s)
    --- PASS: TestLogoutFunctionalEmptySuccessStillRequiresCleanExit/hang (0.27s)
=== RUN   TestLogoutOwnedSnapshotAfterLocalCommitWithoutOperationLock
--- PASS: TestLogoutOwnedSnapshotAfterLocalCommitWithoutOperationLock (0.04s)
=== RUN   TestLogoutOldSnapshotCannotOverwriteNewLogin
--- PASS: TestLogoutOldSnapshotCannotOverwriteNewLogin (0.04s)
=== RUN   TestLogoutLocalFailureAndAbsentBroker
=== RUN   TestLogoutLocalFailureAndAbsentBroker/local_failure
=== RUN   TestLogoutLocalFailureAndAbsentBroker/no_credential
=== RUN   TestLogoutLocalFailureAndAbsentBroker/no_broker
--- PASS: TestLogoutLocalFailureAndAbsentBroker (0.07s)
    --- PASS: TestLogoutLocalFailureAndAbsentBroker/local_failure (0.02s)
    --- PASS: TestLogoutLocalFailureAndAbsentBroker/no_credential (0.02s)
    --- PASS: TestLogoutLocalFailureAndAbsentBroker/no_broker (0.03s)
=== RUN   TestLogoutCanceledBrokerKeepsTombstoneAndCleansScratch
--- PASS: TestLogoutCanceledBrokerKeepsTombstoneAndCleansScratch (0.05s)
PASS
coverage: 37.3% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	2.396s	coverage: 37.3% of statements
```

Exit 0. 독립 경합 probe는 2초 context와 defer cancel/join을 사용한다. 실제 client 대신 시험 executable을 실행한 helper는 기존 CodexBroker의 kill/Wait/join을 통과한다. helper에만 `GORACE=atexit_sleep_ms=0`을 지정해 race 검출을 유지하면서 종료 지연만 제거했다.

위 묶음 이후 새로 추가한 정리 실패 probe:

```text
$ GOCACHE=/tmp/gateway-auth-audit-cache go test -race -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/logout-functional-review/overlay.json ./internal/gateway/auth -run '^TestLogoutFunctionalCleanupFailure' -count=1 -v -timeout 10s
=== RUN   TestLogoutFunctionalCleanupFailureDoesNotClaimFullCompletion
    logout_functional_test.go:25: actual scratch removal failure returned ErrAuthState with local completed, broker cleanup_failed, remote unknown; tombstone retained
--- PASS: TestLogoutFunctionalCleanupFailureDoesNotClaimFullCompletion (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.430s
```

Exit 0, 플랫폼 skip 없이 실제 실패 분기를 관측했다.

CLI 별도 race 대조:

```text
$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-logout-functional-home GOCACHE=/tmp/gateway-auth-audit-cache go test -race ./internal/cli -run '^TestGPTLogout|^TestGPTInstalledLogout' -count=1 -v -timeout 20s
=== RUN   TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote
=== RUN   TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote/completed
=== RUN   TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote/failed
--- PASS: TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote (0.07s)
    --- PASS: TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote/completed (0.04s)
    --- PASS: TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote/failed (0.03s)
=== RUN   TestGPTInstalledLogoutMissingBinary
--- PASS: TestGPTInstalledLogoutMissingBinary (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.248s
```

Exit 0. 첫 AUTH 묶음과 독립 병렬 실행했다.

```text
$ GOCACHE=/tmp/gateway-auth-audit-cache go vet ./internal/gateway/auth ./internal/cli
$ gofmt -l internal/gateway/auth/logout.go internal/gateway/auth/broker.go internal/gateway/auth/store.go internal/cli/gpt_auth.go
```

각 exit 0, 빈 출력. 첫 묶음의 coverage만 별도로 확인했다.

```text
$ GOCACHE=/tmp/gateway-auth-audit-cache go tool cover -func=/tmp/gateway-logout-functional.cover | rg 'logout.go|logoutSnapshot|broker.go.*run'
github.com/modu-ai/moai-adk/internal/gateway/auth/broker.go:56:			run			56.4%
github.com/modu-ai/moai-adk/internal/gateway/auth/logout.go:25:			LogoutWithBroker	75.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/store.go:340:			logoutSnapshot		76.9%
```

이 수치에는 뒤에 추가한 cleanup 실패 probe가 포함되지 않는다. 이번 선택 시험의 coverage를 전체 패키지 품질 기준 통과로 사용하지 않는다.

## Baseline-attribution

```text
$ git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
```

WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. 제품 기준 SHA-256:

```text
d745c35a81616285dc19af51c4d9f5aa90ba6dffef4702e8a417527b49629836  internal/gateway/auth/logout.go
1a5bb5507ec1a732dc50f2a2b2aabc048d9bdf9ab2631fc915f499c6ab5ea1c2  internal/gateway/auth/store.go
37c0963550cdeb389dd60e21c7639a8753859423a426ae932fa1b3319c563481  internal/gateway/auth/broker.go
b3b16d81ed69b9833ee6a60a3cd4ea2554cfde9412fa105a16f1aea522ec7463  internal/gateway/auth/logout_test.go
edbafa7009e533a02d3d124d4e3bcc97629e4a9422de1d5f5634668cf1d021f4  internal/gateway/auth/logout_broker_test.go
955d40662bfd95b89b86bcb58612441dbf9106728cf0ab13d17518bfa7c439c3  internal/cli/gpt_auth.go
9fecb300acedf4302666b206138813b394a3d0776dd7b114b7b2e6b497cbb7d4  internal/cli/gpt_auth_logout_test.go
```

`baseline-sha256.txt`, `independent_test.go`, `overlay.json`은 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/logout-functional-review/`에 있다. overlay는 추가 시험 파일만 매핑하며 제품 파일을 바꾸지 않는다. 마지막 Python hashlib 비교 출력은 `baseline files: 7 changed: []`였다.

공식 broker의 원격 결과 의미는 앞선 읽기 전용 상담에서 직접 확인한 `/tmp/openai-codex-audit.zbvNXB` HEAD `5a9eb14`를 재사용했다. `account_processor.rs:965`는 `logout_with_revoke`, `login/src/auth/manager.rs:2919`는 revoke 오류를 경고로 흡수하고 로컬 삭제를 계속한다. `account_processor.rs:1005`의 빈 RPC 결과에는 원격 결과가 없다. 이번 새 시험이 실제 그 broker를 실행했다고 주장하지 않는다.

## Gaps

- 실제 설치된 Codex의 account/logout 호환성, 원격 revoke 요청·수용·token 무효화는 관측하지 않았다.
- 두 Store handle 경합은 동일 Go process 안에서 수행했다. 별도 OS process의 login/logout 경합은 이번 delta에서 추가 실행하지 않았다.
- 기존 송신 lease watcher와 실제 in-flight HTTP 취소 경로는 앞선 AUTH 감사 범위이며 이번에 반복 검증하지 않았다.
- scratch 생성 실패·seed write 실패·로컬 게시 직후 취소의 모든 분기를 fault injection하지 않았다. 첫 묶음의 LogoutWithBroker coverage 75%는 전체 품질 목표 85% 충족 근거가 아니다. 뒤 cleanup probe를 포함한 coverage는 다시 측정하지 않았다.
- Windows 실행, 전체 AUTH 회귀, broad security 감사, 모든 CLI 후처리 오류 출력은 이번 범위 밖이다. 전체 차원 점수나 전체 SPEC 완료를 표시하지 않는다.

## Residual-risk

늦은 logout이 새로운 로컬 canonical 파일을 보존한다는 관측은 공급자 token-family revoke가 새 로그인 토큰에 미치는 영향의 보장이 아니다. 정상 빈 RPC 응답에도 원격 실패가 포함될 수 있으므로 `unknown`은 유지되어야 한다. 주입한 LogoutBroker가 cancellation/종료 계약을 어기면 Store 자체가 임의 구현을 강제 종료할 수는 없다. 현재 production 경로의 CodexBroker는 합성 subprocess에 대해 timeout·비정상 exit를 오류로 반환했으나 실제 설치본의 종료 동작은 별도 preflight 대상이다.
