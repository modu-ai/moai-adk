# AUTH·supervisor 독립 감사 결함 수리

## Claim

부모 작성자가 1차 감사 F1·F2·F3를 수리했다. 원래 독립 반례를 이번 실행에서 다시 실패시킨 뒤 같은 반례의 통과를 확인했다. AUTH macOS race 및 Linux arm64의 관련 인증/PID 시험도 통과했다. 전체 제품 활성화나 실제 계정 시험 완료는 아니다. 후속 독립 감사의 `auth-supervisor-code-audit-iter2.md`는 delta PASS 100/100이며 부모가 보고서의 범위와 실행 근거를 읽었다.

- F1: broker 제한 시간 감시가 protocol loop와 독립적으로 stdin을 닫고 child를 종료한다. 동기 JSON 쓰기가 막혀도 teardown과 Wait에 도달한다. 감시 goroutine은 반환 전에 회수한다.
- F2: 강제 종료 분기는 항상 ErrBroker다. 비정상 broker 결과가 canonical credential로 저장되지 않는다.
- F3: 기존 homestate의 Unix PID 판정에서 os.ErrProcessDone을 사망으로 분류한다. 기존 supervisor 부모 종료 시험은 lifetime 30초에 대해 1초 안에 종료를 판정하며 수명 타이머의 대체 통과를 막는다. 실제 exec 회귀 시험도 추가했다.
- A1: 무시되던 api_key.go를 credential_apikey.go로 이름만 바꿨다. gitignore는 변경하지 않았다.
- Store.Owns는 private storeRef의 소유 Store 포인터만 판정한다. nil·typed nil·다른 Store·snapshot은 거절한다. 세대·로그아웃 판단은 SendAuthorized가 맡는다.

## Evidence

아래 명령은 이 WT에서 실행했다. `<S>`는 한 invocation 앞에 붙인 `unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-parent-home GOCACHE=/tmp/gateway-auth-audit-cache`다. 독립 probe는 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit/overlay.json`으로 공급했다.

`<S> go test -overlay <위 overlay 경로> ./internal/gateway/auth -run '^TestAudit' -count=1 -v -timeout 12s`

수리 전 exit 1:

```text
broker timeout elapsed but cancellation remained blocked writing loginId to unread stdin; explicit cleanup kill was required
--- FAIL: TestAuditBrokerCancellationCannotBlockOnPipe (1.50s)
broker killed after missing graceful exit was accepted and its credential published
--- FAIL: TestAuditForcedBrokerExitCannotPublishSuccess (0.28s)
FAIL github.com/modu-ai/moai-adk/internal/gateway/auth 2.158s
```

수리 후 같은 명령 exit 0:

```text
bounded cancellation returned an error
--- PASS: TestAuditBrokerCancellationCannotBlockOnPipe (0.15s)
--- PASS: TestAuditForcedBrokerExitCannotPublishSuccess (0.26s)
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.800s
```

`<S> go test -overlay <위 overlay 경로> ./internal/gateway -run '^TestAudit' -count=1 -v -timeout 10s`

권한 제한 안의 첫 실행은 `missing child handoff`로 실패했으며 이를 F3 재현으로 세지 않았다. 프로세스 식별 조회 권한을 얻어 같은 시험을 재실행했다. 수리 전 exit 1:

```text
reaped parent identity=indeterminate fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true
child listener survived parent exit
--- FAIL: TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit (3.04s)
FAIL github.com/modu-ai/moai-adk/internal/gateway 3.438s
```

수리 후 exit 0:

```text
reaped parent identity=dead fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true
child served after actual POSIX exec; port closed after parent exited
--- PASS: TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit (2.05s)
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.453s
```

`go test ./internal/gateway/auth -run TestStoreOwns -count=1`은 Owns undefined RED 후 다음 GREEN을 냈다(exit 0).

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.415s
```

상시 회귀 시험을 추가한 후 최초 race에서 정상 fake broker login/refresh가 실패했다. race로 빌드한 `/tmp/gateway-auth-race-helper.test -test.run=^TestCodexBrokerProcess$ -- app-server`를 빈 stdin 및 임시 HOME/CODEX_HOME에서 직접 실행한 비교 출력:

```text
GORACE='' exit=0 elapsed=1.493
GORACE='atexit_sleep_ms=0' exit=0 elapsed=0.009
```

시험 전용 비공개 testEnv에 해당 옵션을 주고 제품 Timeout/종료 유예는 늘리지 않았다. 정상 broker를 강제 종료해도 성공으로 넘기던 옛 결함이 race 종료 지연도 가렸던 상황이다. 이 비교는 모의 Go broker만 대상으로 한다.

`<S> go test -race ./internal/gateway/auth -run 'TestBroker|TestCodexBroker|TestStore|TestRefresh|TestCrossProcess|TestAuthStore|TestSend|TestEarlyResponse|TestHeaderBoundary|TestCredential|TestScratch|TestExplicitAPI' -count=1 -timeout 60s`

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	4.797s
```

`git check-ignore -v internal/gateway/auth/credential_apikey.go`: exit 1, 출력 없음.

Linux 검증 명령:

```sh
docker run --rm --network none --read-only --cap-drop ALL --security-opt no-new-privileges --tmpfs /tmp:rw,exec,nosuid,size=1g --mount type=bind,src=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified,dst=/src,readonly --mount type=bind,src=/Users/goos/go/pkg/mod,dst=/gomod,readonly --workdir /src --env GOTOOLCHAIN=local --env GOPROXY=off --env GOSUMDB=off --env GOMODCACHE=/gomod --env GOCACHE=/tmp/gocache --env HOME=/tmp/home --env MOAI_HOME=/tmp/moai golang:1.26.8 go test ./internal/gateway/auth ./internal/homestate -run 'TestBroker|TestCodexBroker|TestStore|TestRefresh|TestCrossProcess|TestAuthStore|TestSend|TestEarlyResponse|TestHeaderBoundary|TestCredential|TestScratch|TestExplicitAPI|TestPlatformProcessIdentity' -count=1 -timeout 60s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.873s
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.007s
```

같은 이미지의 `go version` 출력은 `go version go1.26.8 linux/arm64`다. 이 Linux 실행은 race가 아니다.

같은 Docker 옵션·읽기 전용 마운트·환경으로 마지막 Go 인수만 `go test ./internal/gateway -run 'TestSupervisor' -count=1 -timeout 45s`로 바꾸어 실행했다. exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	5.883s
```

상시 실제 exec 회귀도 이 TestSupervisor 선택 범위에 포함된다. macOS 독립 재감사 결과는 별도 `auth-supervisor-code-audit-iter2.md`에서 판정한다.

## Baseline-attribution

이번 부모 수리 시작 직전 fetch origin main exit 0, `git rev-parse --short HEAD; git branch --show-current; git rev-list --count --left-right origin/main...HEAD` 출력은 `81c1d58f9`, `WT-unified-gateway`, `0 2879`였다. 미커밋 변경이 함께 존재하는 해당 WT에서 측정했다. 다른 작성자는 CG CLI/config와 M6 adapter를 맡고 있다. 신규 M6 adapter의 undefined RED 작성 중 root gateway를 함께 시험한 첫 batch는 build failed였으므로 해당 batch를 통과로 기록하지 않는다.

## Gaps

상시 supervisor의 Linux 실행은 위와 같이 추가 확인했다. 실제 Claude·Codex broker·공급자 요청·OAuth refresh·Windows 런타임은 실행하지 않았다. 19시 전 외부 시험 금지는 유지했다. 임의 OnLogin callback 자체가 영원히 반환하지 않는 경우까지 종료할 수 있다는 계약은 검증하지 않았다. 새 Anthropic/GLM credential은 별도 작성·검증 범위다.

## Residual-risk

독립 delta 감사가 통과했어도 Go 시험 helper의 exit 지연 제거는 실제 installed Codex broker의 종료 호환성을 증명하지 않는다. Linux 결과는 선택한 시험 범위와 컨테이너 환경에 한정된다. 제품 gateway factory와 실제 모델 선택 게이트는 여전히 별도 작업이다.
