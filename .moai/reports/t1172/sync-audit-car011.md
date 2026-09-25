# t1172 AC-CAR-011 독립 사전 감사

판정: **FAIL** (기능 60/100). 현재 기준 커밋: `e0079d62a`. 감사 범위는 AC-CAR-011 실행 준비와 AC-CPP-004·013·014 연결이다. 실제 모델 LIVE는 실행하지 않았다.

## Claim

결정적 경계 테스트와 Codex 0.157.0 비모델 시작 검사는 통과했다. 첫 감사에서 발견한 버전 불일치 시 `stop` 누락은 수리됐다. 하지만 시작 검사 뒤 픽스처 초기화 중 `git status`가 실패하면 공용 장부에 `stop` 행을 남기지 않는 별도 우회가 재현됐다. `plan.md` §D의 앞 단계 실패 시 `stop` 기록 계약을 어긴다. LIVE 결과 자체는 미검증이다.

## Evidence

원본 장부의 임시 사본과 `--version`만 `codex-cli 0.158.0`으로 답하고 나머지는 실제 Codex에 위임하는 래퍼를 사용했다. 래퍼의 `exec`는 접속 불가 제공자로 시작 검사만 실행했다. 명령과 관찰 출력:

```text
$ go test ./internal/cli -run '^TestCodexAuditLaunchLiveReadOnlyRoles$' -count=1 -v -timeout=90s
exit= 1
=== RUN   TestCodexAuditLaunchLiveReadOnlyRoles
    codex_preapproval_startup_test.go:171: car011 startup exit=-1 moai=1 decoy=1 items=0
    codex_preapproval_car011_test.go:118: Codex version "0.158.0" differs from 0.157.0
--- FAIL: TestCodexAuditLaunchLiveReadOnlyRoles (25.42s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	26.053s
FAIL
ledger_counts= {"live": 2, "startup": 9, "stop": 0}
last_kind= startup
orig_counts= {"live": 2, "startup": 8, "stop": 0}
```

검사 대상 소스 `internal/cli/codex_preapproval_car011_test.go:104-129`에서 `preApprovalRunCar011Startup`가 장부에 신규 행을 쓴 뒤, 반출 읽기·장부 읽기·버전·repo·build 검사 실패는 `t.Fatal`로 종료하고, `stop` 작성 클로저는 127행 이후에야 정의된다. 위 실행은 그중 버전 분기를 실제로 재현했다.

결정적 검사:

```text
$ go test ./internal/cli -run '^TestCodexPreApproval(Car011PreparationBoundary|Car011ExportIntegrity|GateRefusals|StartupLedgerRetention)$' -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/cli	1.059s	coverage: 5.3% of statements
$ go vet ./internal/cli
(출력 없음, exit 0)
$ gofmt -l internal/cli/codex_preapproval_car011_test.go internal/cli/codex_preapproval_startup_test.go internal/cli/codex_audit_live_test.go
(출력 없음, exit 0)
```

보안 관련 정적 확인은 로그인 내용을 읽거나 출력하지 않았다. `rg -n 'authPath|auth_sha256|CODEX_HOME|os\.WriteFile.*auth' internal/cli/codex_preapproval_car011_test.go internal/cli/codex_preapproval_startup_test.go` 결과에서 car011 장부는 `auth_sha256_before`·`auth_sha256_after`만 적고, 비모델 시작 검사는 임시 `CODEX_HOME`을 썼다. 이는 전체 인증 정보 비노출의 증명이 아니다.

## Baseline-attribution

첫 감사는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`의 `git rev-parse --short HEAD` 결과 `517ea54f2`에서 수행했다. 후속 감사는 `e0079d62a`에서 수행했다. 두 감사 모두 임시 장부 사본에서 시작 검사 8→9, 원본 8 유지가 관찰됐다. 선행 커밋 `a4a67655d`에는 준비 전용 테스트 진입점 9줄이 없었고, 해당 진입점은 `517ea54f2`에서 추가됐다. 각 재현 당시 `git status --short`는 빈 출력이었다.

## Findings

- **F1 [High, resolved in e0079d62a, confidence high]** 기존 `internal/cli/codex_preapproval_car011_test.go:107-129`: 시작 검사 후 반출·버전·repo·build 검사 실패 시 `stop` 누락. 비모델 버전 변이 재검사에서 `stop=1`을 확인했다.
- **F2 [High, blocking, confidence high]** `internal/cli/codex_preapproval_discriminator_live_test.go:344` → `internal/cli/codex_preapproval_startup_test.go:103-105` → `internal/cli/codex_preapproval_car011_test.go:214`: `preApprovalPrepareArm` 내부 `preApprovalRootState(t,...)`가 `git status` 실패에 `t.Fatal`로 테스트를 바로 끝내서 호출자의 `stop` 경로를 우회한다. **필수 수정:** 초기화에서 루트 상태 수집 오류를 반환해 호출자의 `stop(err.Error())`로 보내고, 시작 검사 뒤 `git status` 실패 변이에서 `startup=9`, `live=2`, `stop=1`, 마지막 행 `stop`을 확인하라.

## Dimension Scores

| 차원 | 점수 | 판정 | 근거 |
|---|---:|---|---|
| Functionality (40%) | 60/100 | FAIL | Iteration 2의 `git status` 변이: 시작 검사 성공 뒤 `stop=0` |
| Security (25%) | 미산정 | UNVERIFIED | 인증 해시 경로만 정적 확인, LIVE 및 비밀 노출 검사 미실행 |
| Craft (20%) | 미산정 | UNVERIFIED | 집중 검사 통과; 5.3%는 선택한 테스트로 잰 패키지 전체 수치여서 이 변경의 전체 커버리지 판정에 쓰지 않음 |
| Consistency (15%) | 100/100 | PASS | `gofmt -l` 출력 없음; `go vet ./internal/cli` exit 0 |

## Gaps

실제 두 역할의 Codex LIVE, AC-CAR-011의 쓰기 거부 결과, AC-CPP-004·013·014 최종 증거 판정은 수행하지 않았다. 임시 사본은 테스트 종료 후 제거되어 원자료 경로는 남지 않았고, 위 관찰 출력만 이 파일에 보존했다. 운영자 인증 파일의 내용이나 해시는 보고하지 않았다.

## Residual-risk

F1을 고쳐도 첫 역할과 둘째 역할의 실제 세션 동작, 샌드박스 명령 실행 여부, 시도별 800초 창, 재시도 보존 경로는 LIVE와 장부 판정식을 통과하기 전까지 확정할 수 없다.

## Iteration history

- Iteration 1: `517ea54f2` — F1 재현, FAIL.
- Iteration 2: `e0079d62a` — F1 수리 확인, F2 재현, FAIL. 실제 모델 호출 0회.

### Iteration 2 증거

F1의 버전 변이를 같은 임시 장부 사본에서 다시 돌린 출력:

```text
$ go test ./internal/cli -run '^TestCodexAuditLaunchLiveReadOnlyRoles$' -count=1 -v -timeout=90s
exit= 1
=== RUN   TestCodexAuditLaunchLiveReadOnlyRoles
    codex_preapproval_startup_test.go:171: car011 startup exit=-1 moai=1 decoy=1 items=0
    codex_preapproval_car011_test.go:171: ABORTED: Codex version "0.158.0" differs from 0.157.0
--- FAIL: TestCodexAuditLaunchLiveReadOnlyRoles (25.13s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	25.725s
ledger_counts= {"live": 2, "startup": 9, "stop": 1}
last_kind= stop
stop_reason= Codex version "0.158.0" differs from 0.157.0
orig_counts= {"live": 2, "startup": 8, "stop": 0}
```

픽스처의 3번째 `git status`만 exit 42로 답하는 래퍼를 사용했다. 첫 두 번은 시작 검사 준비, 3번째는 그 검사 뒤 첫 역할 직전의 `preApprovalPrepareArm` 호출이다. 관찰 출력:

```text
$ go test ./internal/cli -run '^TestCodexAuditLaunchLiveReadOnlyRoles$' -count=1 -v -timeout=90s
exit= 1
=== RUN   TestCodexAuditLaunchLiveReadOnlyRoles
    codex_preapproval_startup_test.go:171: car011 startup exit=-1 moai=1 decoy=1 items=0
--- FAIL: TestCodexAuditLaunchLiveReadOnlyRoles (25.61s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	26.207s
FAIL
git_status_count= 3
ledger_counts= {"live": 2, "startup": 9, "stop": 0}
last_kind= startup
orig_counts= {"live": 2, "startup": 8, "stop": 0}
```

`go test ./internal/cli -run '^TestCodexPreApprovalCar011(PostStartupFailuresStop|PreparationBoundary|ExportIntegrity)$' -count=1 -v`는 `fake_version`, `missing_export`를 포함해 모두 PASS했다. 이 테스트가 F2의 간접 `t.Fatal` 경로를 덮지는 않는다. 장부 파일 자체가 읽히지 않는 경우에는 같은 파일에 원자적으로 `stop`을 쓸 수 없으므로 별도 잔여 위험으로 둔다.
