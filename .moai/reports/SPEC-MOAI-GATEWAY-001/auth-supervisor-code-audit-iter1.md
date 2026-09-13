# AUTH·supervisor 독립 코드 감사 — 1차

SPEC: SPEC-MOAI-GATEWAY-001 / SPEC-MOAI-GPT-AUTH-001

Overall Verdict: **FAIL — 65.5/100**. Functionality와 Security의 must-pass 조건을 충족하지 못했다.

감사 범위는 macOS에서 실행한 신규 `internal/gateway/auth/**`, `supervisor*.go`의 로컬 인증 트랜잭션·송신·프로세스 수명 계약이다. AUTH 전체, M4 전체, CLI·PICKER 통합 또는 실계정 완료 판정이 아니다. 제품 코드는 수정하지 않았다.

## Claim

기존 변경 범위 시험과 race 검사는 통과했으나, 독립 반례에서 차단 결함 세 건을 확인했다. 가장 큰 결함은 실제 POSIX exec 이후 부모가 종료되어도 gateway 포트가 남는 것이다. 현재 Go 실행 환경에서 기존 PID 판별기가 `os.ErrProcessDone`을 사망으로 처리하지 않아 신규 supervisor가 부모 종료를 알아내지 못했다. 두 인증 결함은 취소 RPC 쓰기의 무제한 대기와 강제 종료 오류를 버린 뒤의 성공 publish다.

### Dimension Scores

프로필은 `.moai/config/sections/harness.yaml`의 `default_profile: "default"` 및 `.moai/config/evaluator-profiles/default.md`를 읽었다. SPEC에 별도 evaluator_profile이 없어 기본 가중치와 must-pass 규칙을 적용했다. flat weighted-percentage 방식이다.

| Dimension | Score | Verdict | Evidence — 실제 출력 발췌 |
|---|---:|---|---|
| Functionality (40%) | 50/100 | FAIL | `audit_probe_test.go:57: broker timeout elapsed but cancellation remained blocked writing loginId to unread stdin; explicit cleanup kill was required` |
| Security (25%) | 50/100 | FAIL | `audit_supervisor_test.go:57: reaped parent identity=indeterminate fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true` / `audit_supervisor_test.go:64: child listener survived parent exit` |
| Craft (20%) | 90/100 | PASS | `AUDITED_TOTAL 666 757 88.0%`; `go vet` exit 0, stdout/stderr 빈 출력 |
| Consistency (15%) | 100/100 | PASS | `gofmt -l` exit 0, stdout/stderr 빈 출력; `git diff --numstat -- go.mod go.sum` 빈 출력 |

Functionality는 이번에 합의한 **로컬 감사 범위**의 실패이며 전체 SPEC AC를 모두 실행했다는 뜻이 아니다. Security는 F3의 High가 독립 must-pass 조건을 깨므로 다른 점수로 상쇄할 수 없다. 65.5 = 50×0.40 + 50×0.25 + 90×0.20 + 100×0.15.

## Findings

### F1 — [Medium / P2] [blocking] 인증 취소 RPC가 제한 시간을 벗어나 stdin 쓰기에서 멈춘다

- 위치: `internal/gateway/auth/broker.go:232`, 직접 쓰기 함수 `:134-136`, 길이 제한 없이 수락하는 loginId `:216-223`.
- 신뢰도: **높음**, 별도 OS 프로세스와 실제 pipe로 재현.
- 계약: REQ-GA-003, AC-GA-003, AUTH plan B.6의 제한 만료·취소 후 대기 작업 종료.
- 재현: 1MiB scanner 제한 안에 드는 256KiB loginId를 올바른 login/start 응답으로 반환한 fake broker가 이후 stdin을 읽지 않는다. wrapper Timeout은 150ms다. 완료 대기 제한이 끝나도 동기 `send(account/login/cancel)`가 pipe에서 막혀 defer의 Kill/Wait에 도달하지 못했다. 1.5초 뒤 감사 cleanup이 직접 Kill해야 반환했다.
- 영향: 제한 시간이 로그인 취소와 operation lock 해제를 보장하지 않는다. 이는 공급자가 이런 ID를 실제 사용한다는 주장이 아니라, 현재 wrapper가 수락한 입력과 로컬 I/O 정지 조건에서의 실패다.
- 필수 수리: timeout/cancel이 RPC pipe 쓰기와 독립적으로 child 종료를 개시하고, writer와 reaper를 회수하도록 한다. cancel RPC는 종료를 막을 수 없는 best-effort로 취급한다. loginId 길이 검증을 보강할 수 있으나 이것만으로 모든 쓰기 대기의 종료 계약을 대체하지 않는다.
- 병합 차단: **예**, 해당 로컬 인증 계약 수리 후 재감사 필요.

### F2 — [Medium / P2] [blocking] 강제 종료한 broker의 오류를 버려 정상 로그인으로 publish한다

- 위치: `internal/gateway/auth/broker.go:124-128`, 결과를 publish하는 `internal/gateway/auth/store.go:268-327`.
- 신뢰도: **높음**, 실제 subprocess Kill 후 `Store.Status().LoggedIn == true` 확인.
- 계약: AUTH plan B.3·B.6의 broker 종료 확인과 강제 종료 관측, AC-GA-004의 실패 시 성공 표시 금지. 구현 보고서도 종료 코드 확인 뒤에만 저장한다고 명시한다.
- 재현: fake broker가 유효 합성 auth.json과 성공 notification을 출력한 뒤 stdin EOF에도 종료하지 않는다. wrapper는 250ms 뒤 Kill하고 `<-waitDone`의 오류를 버린다. Run은 nil이고 Store.Login은 credential을 canonical로 publish했다.
- 영향: 정상 종료의 오류 경로에서는 거절하는 결과가 강제 종료 분기에서는 성공으로 바뀐다. 강제 종료가 호출자에게 관측되지 않는다. 이 시험은 secret 유출이나 부분 파일 채택을 관측한 것이 아니다.
- 필수 수리: 강제 종료·비정상 Wait 결과가 일반 nil 성공으로 숨지 않도록 반환한다. 현재 로컬 계약에 맞추어 해당 candidate의 publish를 거절하고 기존 세대·상태를 보존한다. 정상 종료 대조군과 강제 종료 대조군을 함께 검증한다.
- 병합 차단: **예**, 실패·성공 구분 계약 수리 후 재감사 필요.

### F3 — [High / P1] [blocking] 실제 부모 종료가 indeterminate가 되어 gateway가 남는다

- 신규 소비 위치: `internal/gateway/supervisor_runner.go:100-101`.
- 원인인 기존 primitive: `internal/homestate/pid_state_unix.go:19-26`, 특히 `:23`. 이 파일은 감사 중 수정하지 않았으며 기존 HEAD 대비 diff도 없다. 새로 도입된 helper 결함이라고 귀속하지 않는다.
- 기존 시험의 누락: `internal/gateway/supervisor_test.go:236`의 Lifetime=3초가 `:249`의 시험 제한=4초보다 짧아, 부모 사망 대신 lifetime 종료로도 해당 시험이 통과한다.
- 신뢰도: **높음**, 실제 exec 및 한 조건만 바꾼 대조 overlay로 인과 확인.
- 계약: REQ-MG-008, core design §3.4의 lead 세션 종료 시 gateway 종료·포트 폐쇄. lifetime은 별도 안전망이다.
- 재현: 모의 launcher가 StartChild 후 `/bin/sleep 2`로 실제 `syscall.Exec`한다. gateway는 exec 후 정상 HTTP 응답을 제공한다. 부모를 Wait로 회수한 뒤 `Signal(0)`은 `os: process already finished`이며 `errors.Is(err, os.ErrProcessDone)==true`, ESRCH는 false였다. 현재 `ProbeProcessIdentity`는 indeterminate를 반환했다. 20ms poll·5초 lifetime에서 부모 종료 뒤 추가 1초 동안 기존 포트가 계속 응답했다. 감사 cleanup으로 child를 종료했다.
- 대조: 원본 파일을 수정하지 않고 overlay에서 `errors.Is(err, os.ErrProcessDone)`을 Dead 조건에 더한 동일 시험은 PASS했다. exec 후 정상 응답이 유지되고 부모 종료 뒤 포트가 닫혔다.
- 영향: 부모 종료를 credential을 가진 gateway의 종료 조건으로 쓰는 계약이 깨진다. 사라진 세션의 loopback gateway가 수명 상한까지 남을 수 있다. 이 감사가 실제 provider 요청이나 정보 유출을 발생시킨 것은 아니다.
- 필수 수리: 실제 종료 오류 `os.ErrProcessDone`을 사망으로 분류하는 좁은 primitive 수리 또는 동등한 검증된 gateway 전용 판별을 적용한다. PID-only 생존으로 약화하지 않는다. 부모 종료 시험은 수명 상한보다 훨씬 짧은 검증 제한을 두고 실제 exec 이후의 종료도 확인하여 lifetime 대체 통과를 막는다.
- 병합 차단: **예**, High 및 세션 수명 계약 실패.

### A1 — [Low / P3] [optional] API 키 구현 파일이 ignore 패턴에 걸린다

- 위치: `.gitignore:195`의 `*_key.*`와 신규 `internal/gateway/auth/api_key.go`.
- 신뢰도: **높음**, `git check-ignore -v` 실제 출력으로 확인.
- 현재는 커밋 전이므로 구현 누락이 이미 원격에 배포되었다는 결함으로 세지 않는다. 커밋을 만들 때 명시적 포함 또는 파일명 조정으로 이 제품 파일이 누락되지 않게 확인해야 한다. 포괄적인 ignore 규칙 개편은 요구하지 않는다.
- 병합 차단: 현재 **아니오**. 제출 결과에서 파일 누락이 실제 발생하면 별도 빌드 결함이 된다.

## Evidence

모든 Go 명령의 작업 디렉터리는 아래 Baseline의 WT다. provider 환경을 같은 invocation에서 제거했다. 로컬 mock 외 Claude·Codex·provider 실서비스를 실행하지 않았다.

### 기존 시험의 실제 재실행

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test ./internal/gateway/auth ./internal/gateway -run 'TestSupervisor|TestStore|TestRefresh|TestCrossProcess|TestAuthStore|TestCodexBroker|TestSend|TestEarlyResponse|TestHeaderBoundary|TestCredential|TestBroker|TestScratch|TestExplicitAPI' -count=1 -coverprofile=/tmp/gateway-auth-supervisor-audit.cover -timeout 45s
```

정상 권한 재실행 exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.330s	coverage: 87.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	7.737s	coverage: 66.4% of statements
```

gateway의 66.4%는 이번에 실행하지 않은 catalog/ingress 등의 패키지 코드도 분모에 포함한다. 감사 대상 파일만 동일 coverprofile에서 statement 수로 재집계한 결과:

```text
github.com/modu-ai/moai-adk/internal/gateway/auth/api_key.go 12 13 92.3%
github.com/modu-ai/moai-adk/internal/gateway/auth/broker.go 115 135 85.2%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential.go 4 4 100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/lock_posix.go 14 16 87.5%
github.com/modu-ai/moai-adk/internal/gateway/auth/send.go 130 150 86.7%
github.com/modu-ai/moai-adk/internal/gateway/auth/store.go 199 226 88.1%
github.com/modu-ai/moai-adk/internal/gateway/supervisor.go 100 107 93.5%
github.com/modu-ai/moai-adk/internal/gateway/supervisor_posix.go 1 1 100.0%
github.com/modu-ai/moai-adk/internal/gateway/supervisor_runner.go 91 105 86.7%
AUDITED_TOTAL 666 757 88.0%
```

집계 명령은 Python으로 coverprofile의 각 `location numStatements count` 행을 읽고 `/auth/` 또는 `/supervisor` 파일만 골라 `numStatements × (count > 0)`와 `numStatements`를 합산했다. Windows 파일은 현재 macOS 분모에 없다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -race ./internal/gateway/auth ./internal/gateway -run 'TestSupervisor|TestStore|TestRefresh|TestCrossProcess|TestAuthStore|TestCodexBroker|TestSend|TestEarlyResponse|TestHeaderBoundary|TestCredential|TestBroker|TestScratch|TestExplicitAPI' -count=1 -timeout 45s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	7.853s
ok  	github.com/modu-ai/moai-adk/internal/gateway	9.514s
```

### 독립 인증 반례

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit/overlay.json ./internal/gateway/auth -run '^TestAudit' -count=1 -v -timeout 12s
```

exit 1, 실제 반례 출력:

```text
    audit_probe_test.go:57: broker timeout elapsed but cancellation remained blocked writing loginId to unread stdin; explicit cleanup kill was required
--- FAIL: TestAuditBrokerCancellationCannotBlockOnPipe (1.50s)
=== RUN   TestAuditForcedBrokerExitCannotPublishSuccess
    audit_probe_test.go:70: broker killed after missing graceful exit was accepted and its credential published
--- FAIL: TestAuditForcedBrokerExitCannotPublishSuccess (0.27s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/auth	2.138s
FAIL
```

### 실제 exec 반례와 좁은 대조

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit/overlay.json ./internal/gateway -run '^TestAudit' -count=1 -v -timeout 10s
```

원본 제품 코드, exit 1:

```text
=== RUN   TestAuditExecParentProcess
--- PASS: TestAuditExecParentProcess (0.00s)
=== RUN   TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit
    audit_supervisor_test.go:57: reaped parent identity=indeterminate fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true
    audit_supervisor_test.go:64: child listener survived parent exit
--- FAIL: TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit (3.03s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	3.427s
FAIL
```

같은 명령에서 overlay 경로만 `control-overlay.json`으로 바꿨다. 그 overlay는 `pid_state_unix.go`의 사망 조건에 `|| errors.Is(err, os.ErrProcessDone)`만 추가한 별도 사본이다. 제품 수리가 적용되었다는 뜻이 아니다. 대조 exit 0:

```text
=== RUN   TestAuditExecParentProcess
--- PASS: TestAuditExecParentProcess (0.00s)
=== RUN   TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit
    audit_supervisor_test.go:57: reaped parent identity=dead fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true
    audit_supervisor_test.go:61: child served after actual POSIX exec; port closed after parent exited
--- PASS: TestAuditSupervisorSurvivesActualExecAndStopsAfterParentExit (2.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.499s
```

### 보안 경계·의존성·형식 검사

```sh
GOCACHE=/tmp/gateway-auth-audit-cache go vet ./internal/gateway/auth ./internal/gateway && gofmt -l internal/gateway/auth internal/gateway/supervisor.go internal/gateway/supervisor_runner.go internal/gateway/supervisor_test.go internal/gateway/supervisor_posix.go internal/gateway/supervisor_windows.go
```

exit 0, stdout/stderr 빈 출력. 추가 `rg`로 프로세스 실행, 인증 헤더, proxy/TLS, 삭제, 원자 교체, flock 지점을 찾고 해당 함수 전체를 읽었다. text match만으로 보안 결함을 확정하지 않았다. 실행한 검색:

```sh
rg -n 'exec.Command|os.Environ|os.Getenv|Authorization|InsecureSkipVerify|Proxy|DialTLS|RemoveAll|atomicfile.Replace|Flock' internal/gateway/auth/*.go internal/gateway/supervisor*.go | head -70
git diff --numstat -- go.mod go.sum
rg -n '^go |golang.org/x/sys|^toolchain' go.mod
git check-ignore -v internal/gateway/auth/api_key.go .moai/reports/SPEC-MOAI-GATEWAY-001/auth-supervisor-code-audit-iter1.md
```

검색 출력의 관련 원문 발췌와 manifest/ignore 출력:

```text
internal/gateway/auth/lock_posix.go:18:		e := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
internal/gateway/auth/send.go:83:	if tr.Proxy != nil {
internal/gateway/auth/store.go:156:	if e = atomicfile.Replace(name, filepath.Join(s.dir, "state.json")); e != nil {
internal/gateway/auth/broker.go:70:	cmd := exec.Command(b.Executable, args...)
3:go 1.26.8
30:	golang.org/x/sys v0.47.0
.gitignore:195:*_key.*	internal/gateway/auth/api_key.go
```

manifest diff는 빈 출력이다. 새 dependency 도입은 이 변경 범위에서 관측되지 않았다. 원격 취약점 데이터베이스 감사는 하지 않았다.

### 검증된 비결함과 관측 경계

기존 시험을 직접 재실행하고 assertion을 읽어 다음 로컬 계약을 확인했다: MoAI scratch를 쓰는 fake broker의 환경 격리, 다른 loginId/실패 응답 거절, 손상·공개 권한·symlink 거절, cross-process refresh 단일 교환, 늦은 refresh의 logout tombstone 덮어쓰기 거절, endpoint/세대 불일치의 pre-send 거절, 실제 net.Pipe header write 동안 logout 대기, header 이후 SSE가 logout을 막지 않음, write timeout 후 lock 해제, overlay 교체 파일 보존, 시작 실패/stop reaper 회수. 이 결과는 설치 Codex 내부 동작이나 모든 filesystem 경쟁을 증명하지 않는다.

## Baseline-attribution

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
AUDITED_SOURCE_HASH_CHANGES 0 []
```

HEAD·브랜치는 감사 시작과 최종 증거 정리 시 다시 읽었다. AUTH·supervisor 18개 Go 파일의 시작 SHA256을 기록한 뒤 재비교했다. 원인 확인 과정에서 읽은 기존 homestate 파일 SHA256도 같은 manifest에 추가했다. 제품 파일 19개 해시는 다음 파일에 있으며 이 보고서 작성 후에도 다시 비교한다.

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-supervisor-audit/baseline-sha256.txt`

핵심 판정 대상의 SHA256(감사 원본):

```text
aa7cd27545fe32b28157497e81badb1f73ecd5d5977c18cb64e837a346af3c4e  internal/gateway/auth/broker.go
c50df460ab39a85f04faf7ae549fe6d35db392efbd02dfe3f6e45062f376ed05  internal/gateway/supervisor_runner.go
ee8b350c5e1f85f8ed5be87843df152cbc34f5c8cc27f32c4e7814b268e4b8ef  internal/homestate/pid_state_unix.go
```

같은 디렉터리에 `audit_probe_test.go`, `audit_supervisor_test.go`, `overlay.json`, `pid_state_unix_control.go`, `control-overlay.json`이 있다. 시험 파일은 제품 디렉터리에 쓰지 않고 Go overlay로 주입했다. fake child는 Test.Cleanup과 외부 Go timeout으로 종료 범위를 정했으며 고아 부하를 남기지 않았다. 이 보고서 자체는 `.moai/reports/SPEC-MOAI-GATEWAY-001/`에 export한다.

source_session_id는 전달된 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이다. 이 subagent의 `moai session current` 실행은 runtime UUID를 재조회하지 못하고 다음을 반환했으므로 재조회 성공으로 세지 않았다:

```text
hint: session.id not available from the runtime; emitted canonical fallback. Run 'moai session doctor' to diagnose.
source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>
```

### Iteration history

- 1차: 기존 scoped 시험/race PASS, 독립 F1·F2·F3 반례 FAIL, F3 단일 조건 대조 PASS, 최종 **FAIL 65.5/100**.
- 최초 supervisor 기존시험은 sandbox에서 지문 조회가 막혀 FAIL이었다. 정상 권한 재실행 결과를 위에 따로 기록했다.
- 최초 exec probe는 auditor가 만든 최소 env에서 PATH를 빠뜨려 handoff 전에 실패했다. `ps` 기반 지문 조회를 위해 `/usr/bin:/bin`을 probe 환경에 보완했다. 이 시험 도우미 오류는 제품 결함으로 세지 않았다. 보완 뒤의 F3 원본 FAIL·대조 PASS만 판정 근거로 썼다.
- 후속 감사는 F1~F3의 수정 delta와 연관된 회귀에 한정한다. A1은 선택 사항이며 추가 추상화 작업을 자동 지시하지 않는다.

## Gaps

- 실제 Claude Opus 5/Sonnet 5, 설치 Codex app-server, 구독 로그인/refresh/provider 수용, API 키 HTTP와 네 GPT 모델의 실제 대화는 실행하지 않았다. 사용자 지정 19:00 KST 이후 시험을 대체하지 않는다.
- Windows는 실행하지 않았다. OpenStore가 명시적으로 미지원을 반환하는 현 상태를 0600/ACL 지원으로 세지 않았다. Linux 런타임도 이 감사의 증거가 아니다.
- remote revoke 미지원은 이미 명시된 범위이며, 성공했다고 요구하거나 기록하지 않았다.
- 기존 사용자 Codex/Claude credential 파일은 읽지 않았다. fake 환경 표지 검사만으로 실제 broker의 모든 시스템 설정·키링 접근 부재가 증명되는 것은 아니다.
- sender의 custom transport는 trusted 입력이다. 임의 악성 net.Conn·관리되지 않는 transport 구현까지 보장한다고 해석하지 않았다. early-response rejection은 기존 net.Pipe 시험에서 통과했으며 추가 독립 반례를 확정하지 않았다.
- 전체 repository suite, CI, cross-backend audit, 실제 terminal job control 및 Claude exec를 실행하지 않았다. 실제 exec 시험은 `/bin/sleep`을 쓰는 POSIX 수명 검증이다.
- 커밋·push·PR·병합·worktree 제거를 하지 않았다. 제출 파일 누락 여부는 실제 staging/commit 단계에 별도 확인해야 한다.

## Residual-risk

같은 UID의 악의적 동시 파일 교체에 대한 완전한 원자 방어, OS 강제 종료 뒤 잔여 파일 전부 회수, 설치 broker 형식·필수 upstream header 호환은 입증하지 않았다. native PID 판별 수리는 현재 macOS Go 1.26.8에서 재현한 오류 형태를 충족해야 하며 다른 플랫폼의 기존 신원 계약을 유지해야 한다. 정상 종료와 수명 상한은 서로 다른 조건이므로 하나의 시험 성공으로 다른 하나를 대신할 수 없다. F1~F3가 수정되어도 실제 사용자 목표의 인증·모델 선택·대화·도구 왕복은 후속 통합 증거가 필요하다.
