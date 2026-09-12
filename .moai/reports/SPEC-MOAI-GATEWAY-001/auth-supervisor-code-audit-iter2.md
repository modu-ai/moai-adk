# AUTH·supervisor 독립 코드 감사 — 2차 delta

SPEC: SPEC-MOAI-GATEWAY-001 / SPEC-MOAI-GPT-AUTH-001

Overall Verdict: **PASS — delta 100/100**.

이 판정은 1차 F1~F3의 수리, 새 `Store.Owns`의 provenance 경계, A1 파일 누락 위험의 해소 및 관련 로컬 회귀만 대상으로 한다. AUTH 전체·M4 전체·CLI/PICKER·실계정·Windows 지원의 완료 판정이 아니다. 신규 Anthropic/GLM credential과 native adapter는 다른 작성자의 소유이므로 이 감사에서 제외했다.

## Claim

1차의 독립 반례 파일을 그대로 다시 실행하여 F1·F2·F3가 모두 통과하는 것을 확인했다. F3 재실행은 1차의 수리 대조 overlay가 아니라 **수정된 실제 제품 코드 + 시험 파일만 추가하는 원래 overlay**를 사용했다. 추가 독립 시험은 실패한 broker 작업 뒤 기존 canonical 바이트·세대·로그인 상태가 보존되고 scratch가 남지 않음을 확인했다. 실제로 로그인한 두 저장소의 같은 세대 credential과 wrapper 위장을 비교하여 Owns가 generation/provider 표시만으로 다른 참조를 인정하지 않음도 확인했다.

### Dimension Scores

기본 evaluator profile의 가중치와 Functionality/Security must-pass를 그대로 적용했다. 범위를 넓혀 재감사하지 않았으며 아래 점수의 분모는 이번 delta 계약이다.

| Dimension | Score | Verdict | Evidence — 실제 출력 |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS | `--- PASS: TestAuditBrokerCancellationCannotBlockOnPipe (0.16s)` / `--- PASS: TestAuditForcedBrokerExitCannotPublishSuccess (0.26s)` / `--- PASS: TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit (2.05s)` |
| Security (25%) | 100/100 | PASS | `same-generation foreign and lookalike refs rejected; owned revoked ref remained pre-send rejected` / `failure preserved existing canonical bytes and generation; scratch count=0` |
| Craft (20%) | 100/100 | PASS | `AUDITED_AUTH_TOTAL 492 559 88.0%` / `AUDITED_SUPERVISOR_TOTAL 192 213 90.1%`; 해당 `go vet` exit 0, 빈 출력 |
| Consistency (15%) | 100/100 | PASS | 최종 `gofmt -l` exit 0, 빈 출력; `git diff --numstat -- go.mod go.sum` 빈 출력 |

### Findings 및 수리 판정

| 1차 항목 | 2차 판정 | 실제 검증과 의미 |
|---|---|---|
| F1, Medium/P2, blocking | **해소** | watchdog가 protocol receive/send와 독립적으로 timeout 시 stdin Close와 Kill을 실행한다. 256KiB loginId + 읽지 않는 stdin의 원래 반례가 0.16초에 오류로 종료했다. 실패 뒤 기존 상태와 scratch cleanup도 별도 확인했다. |
| F2, Medium/P2, blocking | **해소** | 강제 종료 분기가 `ErrBroker`를 명시적으로 반환한다. 원래 forced-success 반례가 PASS하며, 기존 유효 로그인 위에서 같은 실패를 주입해도 canonical 바이트·세대가 바뀌지 않았다. |
| F3, High/P1, blocking | **해소** | `os.ErrProcessDone`이 Dead로 분류되었다. 실제 exec 뒤 살아 있는 gateway가 부모 종료 뒤 포트를 닫았으며, parent-death 영구 시험은 Lifetime 30초와 별도의 1초 실패 제한으로 수명 타이머 대체 통과를 막는다. |
| A1, Low/P3, optional | **해소** | `api_key.go`를 `credential_apikey.go`로 변경했다. 파일 바이트 SHA256은 1차와 같으며 새 경로는 `git status`에 untracked로 나타나고 `git check-ignore`는 exit 1·빈 출력이다. 실제 staging/commit까지 수행한 것은 아니다. |

새 blocking/optional 결함을 확정하지 않았다. 감사 도중 `store.go`에서 Owns 주석 앞 빈 줄 한 개가 gofmt 차이로 나왔으나 부모 작성자가 보완했다. 보완 후 그 빈 줄을 역변환한 SHA256이 직전 기준과 정확히 같아 동작 변경이 없음을 확인했고, 최종 형식 검사는 빈 출력이었다. 다른 import-group 선호를 추가 수리로 요구하지 않았다.

## Evidence

명령의 작업 디렉터리는 아래 Baseline의 WT다. 아래 provider 변수 제거는 각 시험과 같은 invocation에서 수행했다. Go 도구 출력의 package 전체 coverage와 감사 파일만의 coverage는 구분했다.

### F1·F2 원래 독립 반례와 새 영구 시험

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit/overlay.json ./internal/gateway/auth -run '^TestAudit|TestStoreOwns|TestBrokerRegression' -count=1 -v -timeout 12s
```

exit 0, 관련 실제 출력:

```text
=== RUN   TestAuditBrokerCancellationCannotBlockOnPipe
    audit_probe_test.go:52: bounded cancellation returned an error
--- PASS: TestAuditBrokerCancellationCannotBlockOnPipe (0.16s)
=== RUN   TestAuditForcedBrokerExitCannotPublishSuccess
--- PASS: TestAuditForcedBrokerExitCannotPublishSuccess (0.26s)
=== RUN   TestBrokerRegressionBrokerCancellationCannotBlockOnPipe
    broker_regression_test.go:79: bounded cancellation returned an error
--- PASS: TestBrokerRegressionBrokerCancellationCannotBlockOnPipe (0.15s)
=== RUN   TestBrokerRegressionForcedBrokerExitCannotPublishSuccess
--- PASS: TestBrokerRegressionForcedBrokerExitCannotPublishSuccess (0.26s)
--- PASS: TestStoreOwnsOnlyItsLiveReferenceType (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.248s
```

영구 시험에서 race helper에만 설정하는 `testEnv: GORACE=atexit_sleep_ms=0`을 읽었다. 이 설정은 subprocess의 race runtime 종료 대기만 제거하며 제품의 broker Timeout이나 250ms 강제 종료 경계를 완화하지 않는다. 별도로 위 원래 독립 probe는 testEnv를 지정하지 않은 일반 테스트 바이너리로도 통과했다. 실제 Codex 바이너리의 종료 호환성을 측정한 것은 아니다.

### 새 독립 provenance·기존 로그인 보존 시험

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit-iter2/overlay.json ./internal/gateway/auth -run '^TestDelta' -count=1 -v -timeout 12s
```

exit 0:

```text
=== RUN   TestDeltaOwnsResolvedProvenanceAndRevocation
    audit_delta_test.go:26: same-generation foreign and lookalike refs rejected; owned revoked ref remained pre-send rejected
--- PASS: TestDeltaOwnsResolvedProvenanceAndRevocation (0.03s)
=== RUN   TestDeltaBrokerFailuresPreserveExistingStateAndCleanScratch
=== RUN   TestDeltaBrokerFailuresPreserveExistingStateAndCleanScratch/blocked-cancel
    audit_delta_test.go:45: failure preserved existing canonical bytes and generation; scratch count=0
=== RUN   TestDeltaBrokerFailuresPreserveExistingStateAndCleanScratch/forced-success
    audit_delta_test.go:45: failure preserved existing canonical bytes and generation; scratch count=0
--- PASS: TestDeltaBrokerFailuresPreserveExistingStateAndCleanScratch (0.47s)
    --- PASS: TestDeltaBrokerFailuresPreserveExistingStateAndCleanScratch/blocked-cancel (0.17s)
    --- PASS: TestDeltaBrokerFailuresPreserveExistingStateAndCleanScratch/forced-success (0.30s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.917s
```

Owns는 provenance만 판단한다. 로그아웃 뒤에도 같은 Store에서 나온 참조의 Owns는 true지만, SendAuthorized가 이전 세대 송신을 거절하고 dial 호출 계수는 0이다. 이 분리는 시험에 명시되어 있다. Owns가 인증 유효성 검사까지 대체한다는 주장을 하지 않는다.

### AUTH 회귀·race·coverage

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test ./internal/gateway/auth -run 'TestStore|TestRefresh|TestCrossProcess|TestAuthStore|TestCodexBroker|TestSend|TestEarlyResponse|TestHeaderBoundary|TestCredentialReferences|TestBroker|TestScratch|TestExplicitAPI' -count=1 -coverprofile=/tmp/gateway-auth-audit-iter2.cover -timeout 30s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.625s	coverage: 82.7% of statements
```

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -race ./internal/gateway/auth -run 'TestStore|TestRefresh|TestCrossProcess|TestAuthStore|TestCodexBroker|TestSend|TestEarlyResponse|TestHeaderBoundary|TestCredentialReferences|TestBroker|TestScratch|TestExplicitAPI' -count=1 -timeout 30s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	4.596s
```

82.7%는 새로 추가된 감사 제외 파일 `credential_anthropic.go`, `credential_glm.go`도 분모에 포함한 package 수치다. 같은 profile에서 두 파일을 제외하고 각 행 `numStatements × (count > 0)`와 `numStatements`를 합산한 실제 출력:

```text
AUDITED_AUTH_TOTAL 492 559 88.0%
```

### F3 원래 실제 exec 반례

번역 작성자의 공용 gateway 파일 작성이 일단 마무리되어 부모가 검사를 허용한 뒤 실행했다. 제품 수리를 덮는 control-overlay는 사용하지 않았다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit/overlay.json ./internal/gateway -run '^TestAudit' -count=1 -v -timeout 10s
```

exit 0:

```text
=== RUN   TestAuditExecParentProcess
--- PASS: TestAuditExecParentProcess (0.00s)
=== RUN   TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit
    audit_supervisor_test.go:57: reaped parent identity=dead fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true
    audit_supervisor_test.go:61: child served after actual POSIX exec; port closed after parent exited
--- PASS: TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit (2.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.451s
```

### Supervisor 및 기존 identity 소비 경로 회귀

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test ./internal/gateway -run '^TestSupervisor' -count=1 -coverprofile=/tmp/gateway-supervisor-audit-iter2.cover -timeout 20s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	6.358s	coverage: 20.5% of statements
```

gateway package는 현재 native adapter 등 다른 작업의 코드를 포함하므로 그 전체 분모를 이 delta의 품질 수치로 쓰지 않는다. `/supervisor` 파일만 동일 profile에서 합산한 실제 출력:

```text
AUDITED_SUPERVISOR_TOTAL 192 213 90.1%
```

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -race ./internal/gateway -run '^TestSupervisorParentProcessDeath$|^TestSupervisorExecRegression' -count=1 -timeout 15s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	3.595s
```

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test ./internal/homestate -run '^TestPlatformProcessIdentityLiveDeadAndIndeterminate$|^TestProfileLeaseReconcilePIDFingerprint$' -count=1 -v -timeout 10s
```

exit 0:

```text
=== RUN   TestPlatformProcessIdentityLiveDeadAndIndeterminate
--- PASS: TestPlatformProcessIdentityLiveDeadAndIndeterminate (0.00s)
=== RUN   TestProfileLeaseReconcilePIDFingerprint
--- PASS: TestProfileLeaseReconcilePIDFingerprint (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.376s
```

### 정적 검사·수리 경계·파일 포함 여부

실행 명령:

```sh
GOCACHE=/tmp/gateway-auth-audit-cache go vet ./internal/gateway/auth
GOCACHE=/tmp/gateway-auth-audit-cache go vet ./internal/gateway ./internal/homestate
gofmt -l internal/gateway/auth/broker.go internal/gateway/auth/store.go internal/gateway/auth/credential_apikey.go internal/gateway/auth/broker_regression_test.go internal/gateway/auth/ownership_test.go
gofmt -l internal/gateway/supervisor.go internal/gateway/supervisor_runner.go internal/gateway/supervisor_test.go internal/gateway/supervisor_exec_test.go internal/homestate/pid_state_unix.go
git diff --numstat -- go.mod go.sum
git check-ignore -v internal/gateway/auth/credential_apikey.go
git status --short -- internal/gateway/auth/credential_apikey.go internal/gateway/auth/api_key.go
```

최종 vet·gofmt·manifest diff는 exit 0·빈 출력. check-ignore는 exit 1·빈 출력. status 실제 출력:

```text
?? internal/gateway/auth/credential_apikey.go
```

`rg -n 'testEnv|bounded.Done|Process.Kill|resultErr = ErrBroker|Owns' internal/gateway/auth/broker.go internal/gateway/auth/store.go`로 수리 지점을 찾고 두 함수 전체를 읽었다. 관련 실제 검색 출력:

```text
internal/gateway/auth/broker.go:99:		case <-bounded.Done():
internal/gateway/auth/broker.go:101:			_ = cmd.Process.Kill()
internal/gateway/auth/broker.go:148:			resultErr = ErrBroker
internal/gateway/auth/broker.go:149:			_ = cmd.Process.Kill()
```

이 검색 자체를 행동 증거로 세지 않고 앞의 실제 pipe·process·Store 시험으로 확인했다. 의존성 원격 CVE 조회는 하지 않았다.

## Baseline-attribution

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
FINAL_AUDITED_SOURCE_HASH_CHANGES 0 []
```

source_session_id: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`. 변경된 uncommitted tree를 측정했으며 HEAD가 같다는 이유로 1차 결과를 최신 측정으로 재사용하지 않았다.

핵심 최종 SHA256:

```text
9ba37d0a5f2c373a761930eba6eb716541ac4e68ae3e4b9f4b98067e20d8f24b internal/gateway/auth/broker.go
0e2765a42286edc3601af07e557d978c6328448dac44afa201038d7f58ba71cb internal/gateway/auth/store.go
f9244c86e1debe41becf2e2da03d6c7c8366677908e62dfac8a23fb3ffdd8c5a internal/gateway/auth/credential_apikey.go
40c865cffc31946b0461685583f5f0ae48853d1bdcfc542db03d988e1a3b5629 internal/homestate/pid_state_unix.go
57530f55942d1c998f11ad193af4f5b9e92f208ffae8f8411a6789f01c265c65 internal/gateway/supervisor_test.go
f096582496b49ff73a18018972b5d1871a2a0f3586f3eefaeae0a68c13ccacea internal/gateway/supervisor_exec_test.go
```

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit-iter2/`에 감사 범위 22개 파일의 시작·최종 hash 목록, 새 `delta_test.go`, overlay가 있다. 제품 파일 변경은 부모 작성자가 한 store.go의 빈 줄 한 개만 감사 도중 관측했고 다음 대조 출력으로 확인했다:

```text
SOURCE_CHANGE internal/gateway/auth/store.go ONLY_EXPECTED_BLANK_LINE True
FINAL_HASH_FILES 22
```

검증 도우미는 같은 WT의 ignored 증거 경로에만 만들었다. 제품·SPEC·다른 작성자의 native credential 파일을 수정하지 않았다. report를 `.moai/reports/SPEC-MOAI-GATEWAY-001/auth-supervisor-code-audit-iter2.md`에 export하고 읽어 확인한다.

### Iteration history

- 1차: F1 Medium·F2 Medium·F3 High 때문에 **FAIL 65.5/100**. 원본 반례와 당시 해시는 `auth-supervisor-code-audit-iter1.md`에 보존했다.
- 수리: 부모가 watchdog, 강제 종료 오류 전파, os.ErrProcessDone 사망 판별, 영구 회귀 시험, Owns, 파일명 변경을 적용했다.
- 2차: 같은 반례를 현재 tree에서 다시 측정하고 독립 보존/provenance 시험과 관련 회귀를 추가해 **delta PASS 100/100**.
- 감사 도중 최초 새 delta 명령의 경로에 session UUID 한 글자 오기가 있어 `go: reading overlay ... no such file or directory`로 실행 전 실패했다. 올바른 실제 경로로 고친 실행의 결과만 위 판정에 사용했다.

## Gaps

- 실제 Claude Opus 5/Sonnet 5, Codex app-server, GPT 구독 로그인·refresh·provider acceptance·remote revoke·네 GPT 모델의 왕복은 실행하지 않았다. 사용자 지정 실서비스 시험 시점과 모델 조건은 여전히 별도이다.
- Windows와 Linux 런타임을 이 auditor가 이번 delta에서 실행하지 않았다. 부모가 알린 Linux 결과를 새 독립 측정으로 표시하지 않았다. Windows OpenStore 미지원 Gap도 남는다.
- AUTH 전체 AC, CLI·PICKER·native adapter·새 Anthropic/GLM credential은 범위 밖이다. 이 보고서는 이들을 PASS 처리하지 않는다.
- private callback `OnLogin`이나 trusted transport의 모든 임의 구현에 대한 종료 보장을 새로 증명하지 않았다. 현재 확인은 1차 반례와 해당 변경의 로컬 계약이다.
- 사용자 Codex/Claude 인증 파일과 실제 keyring은 읽지 않았다. mock 격리의 성공을 설치 broker의 모든 접근 부재로 확대하지 않는다.
- 전체 repository suite·CI·교차 모델 감사·커밋·push·PR·병합·worktree 제거를 하지 않았다. 해당 결과는 미완료이다.

## Residual-risk

기존 1차 보고서에서 명시한 같은 UID의 filesystem 동시 교체, OS 강제 종료 뒤 잔여 파일, 실제 broker 및 provider 프로토콜 호환의 한계는 그대로다. 이번에는 일반 반환과 비정상 종료를 구분하고 실제 부모 사망을 알아내는 수리가 확인되었지만, 전체 사용자 흐름은 별도 통합 실행 증거가 필요하다. 수명 상한은 부모 사망 관측을 대신하지 않으며, Owns는 현재 credential 유효성이나 송신 시점 세대 검사를 대신하지 않는다.
