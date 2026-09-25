# t1172 AC-CAR-011 독립 사전 감사

## 현재 판정 — Iteration 3 (`877806354`)

**판정: PASS (실모델 호출 전 경계만, 기능 100/100).** 앞선 F1·F2를 실제 Codex 0.157.0의 비모델 시작 검사와 격리 장부에서 다시 확인했다. AC-CAR-011의 역할별 LIVE와 최종 AC 판정은 아직 미측정이다.

### Claim

세 번째 픽스처 `git status` 실패, Codex 버전 불일치, 시작 검사 뒤 반출 매니페스트 삭제가 모두 신규 LIVE 호출 전에 종료되고 사유가 있는 `stop` 행을 남긴다. 반출 준비 중 `codex mcp list` 실패는 시작 검사나 LIVE를 시작하지 않았지만 `stop` 행도 남기지 않았다(F3). F3는 `plan.md` §D3의 일반 순서 문구와 맞지 않는 관측이다. 다만 `spec.md` REQ-CPP-001·003·004와 `acceptance.md` AC-CPP-012·014는 **호출 전 MCP 목록 조회 실패**의 `stop` 행을 명시적 PASS 조건으로 삼지 않으므로, 이번 사전 LIVE 게이트에서는 선택 개선으로 분류한다.

### Evidence

기준 원본 `.moai/reports/t1172/ledger.json`의 `startup=8, live=2, stop=0`과 M1-a 첫 네 행·원자료 12개를 각각 독립 임시 디렉터리에 복사했다. `MOAI_CODEX_BIN` 안전 래퍼는 `codex mcp list --json`과 접속 불가 제공자를 쓰는 `startup check`만 실제 Codex 0.157.0으로 넘기고, 다른 `codex exec`는 종료 코드 97로 막았다. 변이별로 `MOAI_CODEX_ROLE_LIVE=1`, `MOAI_T1143_EVIDENCE_DIR=<임시 사본>`, `MOAI_CODEX_BIN=<안전 래퍼>`를 지정해 아래 명령을 실행했다. 실패를 의도한 테스트이므로 명령 종료 코드 1이 기대값이다. 원본 로그인 파일의 내용은 출력하지 않았다.

```text
$ go test ./internal/cli -run '^TestCodexAuditLaunchLiveReadOnlyRoles$' -count=1 -v -timeout=90s
# 임시 픽스처의 세 번째 git status --porcelain --ignored만 exit 42
CASE git_status_3 exit 1 counts {'startup': 9, 'live': 2, 'stop': 1} last_kind stop git_status_count 3 unexpected_model False
STOP_REASON fixture git status: exit status 42
TEST_OUTPUT     codex_preapproval_startup_test.go:179: car011 startup exit=-1 moai=1 decoy=1 items=0
TEST_OUTPUT     codex_preapproval_car011_test.go:175: ABORTED: fixture git status: exit status 42
ORIGINAL_SHA_UNCHANGED True COPY_INITIAL_ROWS_UNCHANGED True

# 래퍼의 --version만 codex-cli 0.158.0으로 반환
CASE fake_version exit 1 counts {'startup': 9, 'live': 2, 'stop': 1} last_kind stop unexpected_model False
STOP_REASON Codex version "0.158.0" differs from 0.157.0
TEST_OUTPUT     codex_preapproval_startup_test.go:179: car011 startup exit=-1 moai=1 decoy=1 items=0
TEST_OUTPUT     codex_preapproval_car011_test.go:171: ABORTED: Codex version "0.158.0" differs from 0.157.0
ORIGINAL_SHA_UNCHANGED True COPY_INITIAL_ROWS_UNCHANGED True

# 시작 검사 뒤 --version 경계에서 임시 사본의 car011/export-manifest.json만 삭제
CASE missing_export exit 1 counts {'startup': 9, 'live': 2, 'stop': 1} last_kind stop unexpected_model False
STOP_REASON car011 export reference differs
TEST_OUTPUT     codex_preapproval_startup_test.go:179: car011 startup exit=-1 moai=1 decoy=1 items=0
TEST_OUTPUT     codex_preapproval_car011_test.go:175: ABORTED: car011 export reference differs
ORIGINAL_SHA_UNCHANGED True COPY_INITIAL_ROWS_UNCHANGED True

# 반출 준비의 codex mcp list --json만 exit 42; 모든 모델 호출 차단
EXIT 1 COUNTS {'startup': 8, 'live': 2, 'stop': 0} LAST live MODEL_ATTEMPT False COPY_LEDGER_UNCHANGED True ORIGINAL_SHA_UNCHANGED True
    codex_preapproval_startup_test.go:179: car011 mission-governor argv: mcp list: exit status 42: INJECTED_MCP_LIST_EXIT_42
--- FAIL: TestCodexAuditLaunchLiveReadOnlyRoles (4.78s)
```

선택한 결정적 테스트와 언어 도구 확인:

```text
$ unset MOAI_CODEX_ROLE_LIVE MOAI_CODEX_PREAPPROVAL_LIVE MOAI_T1143_EVIDENCE_DIR MOAI_T1172_EVIDENCE_DIR && go test ./internal/cli -run '^TestCodexPreApproval(Car011PostStartupFailuresStop|Car011GitStatusFailureStops|Car011PreparationBoundary|Car011ExportIntegrity|StartupLedgerRetention|GateRefusals)$' -count=1 -v -cover
--- PASS: TestCodexPreApprovalCar011PreparationBoundary (0.00s)
--- PASS: TestCodexPreApprovalCar011ExportIntegrity (0.00s)
--- PASS: TestCodexPreApprovalCar011PostStartupFailuresStop (0.00s)
--- PASS: TestCodexPreApprovalCar011GitStatusFailureStops (0.00s)
--- PASS: TestCodexPreApprovalGateRefusals (0.25s)
--- PASS: TestCodexPreApprovalStartupLedgerRetention (0.00s)
PASS
coverage: 5.3% of statements
ok  github.com/modu-ai/moai-adk/internal/cli  1.032s  coverage: 5.3% of statements
$ go vet ./internal/cli
(출력 없음, exit 0)
$ gofmt -l internal/cli/codex_preapproval_car011_test.go internal/cli/codex_preapproval_startup_test.go internal/cli/codex_preapproval_discriminator_live_test.go
(출력 없음, exit 0)
$ go mod verify
all modules verified
$ jq -cS '.[0:4][]' .moai/reports/t1172/ledger.json | shasum -a 256
4e9b04acb14ed10b7122f1740fdcfaec147a91698bdebe2867a43123d6cb73ef  -
$ shasum -a 256 <M1-a 원자료 12개를 경로 순서대로 열거> | shasum -a 256
c8653794d64014487c9f05d3e54f07e4b5fb61cf783f5e18f6460bb0e6a4b937  -
```

### Baseline-attribution

모든 측정은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`의 `git rev-parse --short HEAD` 출력 `877806354`, 빈 `git status --short`, Codex CLI 출력 `codex-cli 0.157.0`에서 수행했다. 원본 장부의 현재 sha256은 `16ed6ca3f54c12cade178d62e5d9064754a39600292720dd05e5dd612ca2aa62`였다. 각 변이는 독립 임시 사본에만 기록됐다. 최초 안전 래퍼는 비모델 `mcp list`까지 차단해 F1의 해당 시도를 판정하지 못했다. 위에 기록한 F1 결과는 이를 허용하도록 고친 별도 재실행의 출력이다.

### Dimension Scores

| 차원 | 점수 | 판정 | 근거 |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS (사전 경계) | 세 번째 git 실패, 버전·반출 실패가 모두 `stop=1`, 신규 LIVE 0 |
| Security (25%) | 100/100 | PASS (측정 범위) | 안전 래퍼가 모델 호출을 차단; 원본 장부 불변, 비모델 시작 검사만 위임; `go mod verify` 통과 |
| Craft (20%) | 미산정 | UNVERIFIED | 집중 검사 통과. `5.3%`는 선택한 테스트의 큰 패키지 전체 피복률이므로 변경 코드 85% 판정에 쓰지 않음 |
| Consistency (15%) | 100/100 | PASS | `go vet` exit 0, `gofmt -l` 출력 없음 |

### Findings

- **F1 [High, resolved, confidence high]** `codex_preapproval_car011_test.go`의 시작 검사 뒤 버전·반출 실패는 현재 `stop`을 기록한다. 두 독립 변이 모두 신규 LIVE 0회였다.
- **F2 [High, resolved, confidence high]** `codex_preapproval_discriminator_live_test.go:344`의 루트 상태 수집 오류는 반환되어 호출자의 `stop`으로 이어졌다. 세 번째 `git status` exit 42에서 재현했다.
- **F3 [Medium, optional, confidence high]** `codex_preapproval_startup_test.go:542-545`: 반출 준비 중 `codex mcp list`가 실패하면 `t.Fatalf`가 장부에 `stop` 없이 종료한다. 시작 검사와 LIVE는 모두 새로 시작하지 않았다. `plan.md:80` 일반 순서 설명과의 차이를 없애려면 오류를 반환해 장부에 사유를 기록하라. REQ-CPP-001·003·004와 AC-CPP-012·014에는 이 사전 조회 실패의 `stop` 의무가 직접 명시되지 않아 현재 LIVE 게이트 차단으로 세지 않는다.

### Gaps

두 역할의 실제 Codex LIVE, 쓰기 거부 결과, AC-CAR-011 최종 판정식, Windows 실행, 시작 검사 뒤 장부 자체가 읽히지 않거나 쓸 수 없는 경우는 관찰하지 않았다. 일부 `preApprovalRunStartupFor` 초기화 단계의 다른 `t.Fatal`은 정적으로만 찾았고 별도 오류 주입을 하지 않았으므로 확인된 결함으로 세지 않는다.

### Residual-risk

F3 때문에 시작 검사 전에 실패한 준비 작업의 원인을 장부만으로 추적할 수 없고, 재실행에서 같은 준비를 반복할 수 있다. 실모델 경로에서 역할별 샌드박스·도구 승인·쓰기 거부가 예상대로 작동하는지는 이 사전 감사로 보증할 수 없다.

### Iteration history

- Iteration 1: `517ea54f2` — F1 재현, FAIL.
- Iteration 2: `e0079d62a` — F1 수리 확인, F2 재현, FAIL.
- Iteration 3: `877806354` — F2 및 F1의 버전·반출 변이 수리 확인, 사전 경계 PASS. F3는 범위 밖 준비 진단 개선으로 기록. 실제 모델 호출 0회.

## 이전 Iteration 2 판정 및 원자료 기록

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
