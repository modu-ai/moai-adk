---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "acceptance.md — 인수 기준"
version: "0.3.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# acceptance.md — AC-MH-001..016

> 이 파일은 검증층이고 시나리오는 Given-When-Then이다. 요구(GEARS)는 `spec.md` §C 의 REQ-MH-001..016 이 운반한다. 모든 AC는 이진 판정이 가능하다.
>
> **순서는 design.md D-3.1 의 O번호와 시험 훅 단계명으로만 가리킨다.** 이 문서는 수명 순서를 다시 적지 않으며, 각 순서 단계마다 시험이 단언하는 관측(Start 결과, Close 결과, 자식·토큰 디렉터리·연결·launch-pending 행의 유무, 정리 횟수)을 하위 케이스 표(§1.3–§1.5)에 적는다.
>
> **명령 규약.** 검증 명령은 단일 호출형이고 `-run` 패턴은 앵커형(`^…$`)이다. 부하 규율상 `go test ./internal/cli` 전체는 로컬에서 돌리지 않는다. **파이프 문자를 담은 명령은 표 밖 §1.1 의 fenced 블록에 원문 그대로(바이트 단위로 파이프) 적고 표에서는 AC id로 가리킨다** — 표 안의 이스케이프(백슬래시 파이프)는 Go 정규식에서 리터럴 파이프라 테스트를 0개 고르고 `ok` 를 내기 때문이다(측정은 §3). 표의 명령에는 파이프가 없다.
>
> **레인 세션에서 `go test` 는 환경 정리 복합형으로 돌린다**: `unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test …` (실측: 기준 트리에서 정리 없이 돌린 스코프 명령은 `TestManagedSwitchDoesNotReachCodexLaneLoop` 가 레인 환경 때문에 `lane boundary` 거부로 가짜 적색이었고, 정리 후 PASS). 아래 명령은 그 접두 없이 적는다.
>
> **빈 스윕 규율.** 모든 `-run` 명령은 읽기 전에 선택된 테스트 수를 먼저 확인한다(`go test … -list '<같은 패턴>'`). `[no tests to run]`·`[no test files]` 는 통과가 아니라 **실패(미측정)** 이고, `-list` 가 이름을 하나도 안 내는 `ok` 도 같다(측정: 기준 트리에서 `-list` 로 AC-MH-004 패턴을 고르면 출력이 `ok  github.com/modu-ai/moai-adk/internal/cli  0.874s` 한 줄뿐이고 `Test` 이름이 0개다 — 시험이 아직 없기 때문이며, 이 상태에서 `-run` 이 `ok` 를 내는 것이 곧 공허 통과다).
>
> **시간 상한 단언.** 시험 안의 시간 상한은 전부 벽시계 경과 시간 단언이며 POSIX 시그널에 의존하지 않으므로 Windows에서도 성립한다. 시그널을 보내는 AC-MH-010 만 `runtime.GOOS == "windows"` 에서 건너뛴다.
>
> **측정 기준 트리**: `7109e0900`(카드 기준 develop; 계획 커밋 이후도 소스는 동일). "측정" 표기는 이 plan 실행이 관측한 것이다.

## §1. AC 매트릭스

| AC | 요구 | 시나리오 (Given-When-Then) | 검증 명령 → 기대 관측 |
|---|---|---|---|
| AC-MH-001 | REQ-MH-001, REQ-MH-002, REQ-MH-003 | Given 가짜 App Server가 턴 중에 서버 요청 10종(각각 정수 id)과 목록에 없는 method 하나를 차례로 보낼 When Codex 소유자가 처리하면 Then 각 요청에 정확히 한 번, 같은 id로, design.md D-1 정책표대로 답하고(승인류는 `decline`/`denied`/빈 `permissions`/`decline` 액션, 오류류는 `-32000`, 모르는 method는 `-32601`), 어느 답에도 accept 계열 결정이나 비어 있지 않은 권한 부여가 없으며, 답한 요청마다 stderr에 method를 담은 한 줄이 남고 elicitation 줄은 `serverName` 과 귀속 턴(`turn=<id>` 또는 `turn=none`)도 담는다 | `go test ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v` → 상위 `--- PASS: TestManagedCodexServerRequestPolicy (…s)` 1줄과 그 아래 하위 테스트 11개(10종 + 미지 method 1)가 각각 PASS(각 하위 케이스가 응답 본문과 stderr 줄을 단언), exit 0 |
| AC-MH-002 | REQ-MH-001 | Given 가짜 App Server가 `turn/start` 응답 뒤 명령 승인 요청을 보내고 그 답을 받은 뒤에만 `turn/completed` 를 보낼 When `DeliverTurn` 을 부르면 Then 답이 서버에 도달하고 `DeliverTurn` 이 nil 을 반환하며, 반환까지 걸린 시간이 시험이 정한 **절대 상한 5초 미만**이다(턴 타임아웃 상수는 10분이므로 답을 못 받으면 이 상한을 넘는다). 같은 시험의 하위 케이스: 요청이 턴 사이(한가한 때)에 와도 답한다 | `go test ./internal/cli -run '^TestManagedCodexTurnSurvivesServerRequest$' -count=1 -v` → 하위 `during_turn`, `between_turns` 2개 PASS(각각 경과 5초 미만 단언), exit 0 |
| AC-MH-003 | REQ-MH-004 | Given 서버 요청의 id가 대기 중인 클라이언트 id와 같거나 문자열(`"srv-1"`)일 When 소유자가 `turn/start` 응답을 기다리면 Then 서버 요청은 응답으로 오인되지 않고(진짜 응답 본문이 호출자에게 도달), 문자열 id의 답장에 같은 문자열이 되돌아가며, 읽기 루프가 계속 살아 이후 호출이 성공한다 | `go test ./internal/cli -run '^TestManagedCodexServerRequestIDCollision$' -count=1 -v` → 하위 `integer_collision`, `string_id` 2개 PASS, exit 0 |
| AC-MH-004 | REQ-MH-005, REQ-MH-012 | Given 읽기 고루틴이 답장을 쓰는 동안 다른 고루틴이 `call()` 로 쓰고, 별도로 `Close` 를 여러 고루틴이 동시에 부를 When 경합 검출기로 돌리면 Then 데이터 경합 보고가 0건이고 `Close` 는 정리를 한 번만 수행하며(`teardown` 훅 횟수 1) 모든 호출자가 첫 정리가 끝난 뒤에 같은 결과를 받는다(스트림·Codex 두 소유자) | §1.1 의 AC-MH-004 블록 — 선행 `-list` 가 **정확히 2개** 최상위 `Test` 이름(`TestManagedCodexConcurrentWrites`, `TestManagedSessionCloseConcurrent`)을 내야 하고, 이어지는 `-race` 실행이 두 최상위와 하위 케이스 PASS, `WARNING: DATA RACE` 0건, `[no tests to run]` 없음, exit 0 (`-race` 는 cgo 필요: 측정 `CGO_ENABLED=1`) |
| AC-MH-005 | REQ-MH-006 | Given 우선 턴 성공 뒤 두 번째 턴이 에러 결과를 내고 세 번째 턴이 성공하는 스크립트 세션 When 드라이버가 돌면 Then 드라이버는 두 번째 턴 뒤에 반환하지 않고 세 번째 턴을 전달하며, stderr에 `Factory turn failed (1/` 로 시작하는 한 줄이 남고 stdout에는 쓰이지 않는다 | `go test ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$' -count=1 -v` → PASS, 시험이 세션이 받은 턴 3건과 stderr 한 줄을 단언 |
| AC-MH-006 | REQ-MH-006, REQ-MH-007 | Given §1.4 의 하위 케이스 14개(실패 원인별 10행과 elicitation 귀속 4행)와, 소유 App Server 승인 인수의 서버 이름 접두가 브로커 서버 이름 상수와 같다는 고정 시험 When 소유자와 드라이버를 통과시키면 Then §1.4 표의 각 기대가 성립한다 | §1.1 의 AC-MH-006 블록 — 선행 `-list` 가 **정확히 2개** 최상위 `Test` 이름을 내야 하고, 이어지는 `-run` 이 최상위 2개와 `TestManagedTurnFailureClassification` 의 하위 14개 PASS, exit 0 |
| AC-MH-007 | REQ-MH-008 | Given 연속 턴 단위 실패 횟수가 `config.DefaultManagedSessionMaxConsecutiveTurnFailures` 와 같을 때 When 드라이버가 그 횟수째 실패를 받으면 Then 마지막 오류를 횟수와 함께 반환한다. 상한−1번 실패 뒤 성공 턴이 오면 횟수가 0으로 돌아가 다음 실패가 1번째로 센다. 시험은 상수 자체를 읽어 횟수를 정한다(구현에 숫자를 박아도 시험이 깨진다) | `go test ./internal/cli -run '^TestManagedDriverConsecutiveFailureCeiling$' -count=1 -v` → 하위 3개(`at_ceiling_returns`, `success_resets`, `below_ceiling_continues`) PASS · `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` → 정의 1행, exit 0 |
| AC-MH-008 | REQ-MH-009 | Given 브로커에 메시지 1건이 claim되고 그 메시지를 실은 턴이 실패할 When 드라이버가 세션을 이어 가면 Then 관리 계층은 그 행을 `claimed` 로 두고(`acknowledged` 0, claim token 불변) 해제·재주소를 하지 않는다. 재배달은 lease 만료 뒤 브로커의 `Claim` 이 새 token으로 한다 | `go test ./internal/cli -run '^TestManagedFailedTurnLeavesClaimUntouched$' -count=1 -v` → PASS · `go test ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -v` → 하위 `lost_receipt_redelivery` 포함 7개 PASS (측정: 기준 트리 PASS) |
| AC-MH-009 | REQ-MH-010, REQ-MH-011 | Given 가짜 세션이 `DeliverTurn` 안에서 자기 `Close` 가 불릴 때까지 막히고(또는 드라이버가 한가할 때) When 컨텍스트가 취소되면 Then 워처가 `Close` 를 불러 `DeliverTurn` 이 풀리고, 드라이버는 중단 오류를 반환하며 그 오류를 턴 실패로 세거나 `Factory turn failed` 로 기록하지 않는다 | `go test ./internal/cli -run '^TestManagedDriverStopsOnCancel$' -count=1 -v` → 하위 `idle`, `in_flight_turn` 2개 PASS, 각 5초 이내 |
| AC-MH-010 | REQ-MH-010, REQ-MH-011 | Given 시험 바이너리를 `os/exec` 로 다시 기동해 소유자를 돌리고(스트림은 가짜 백엔드, Codex는 가짜 App Server), 재실행 도우미가 훅 단계 도달을 stdout 한 줄 `ready <단계명>` 으로 알릴 When 부모 시험이 **그 줄을 본 뒤에만** SIGTERM·SIGHUP·SIGINT를 `os.Process.Signal` 로 보내면 Then §1.3 표의 하위 케이스별 기대가 성립한다(공통: 5초 안 종료, 종료 오류 문자열이 `interrupted` 를 담음, launch-pending 행 0건) | `go test ./internal/cli -run '^TestManagedLauncherSignalTeardown$' -count=1 -v` → 하위 6개(`stream_SIGTERM`, `stream_SIGHUP`, `stream_SIGINT`, `stream_SIGTERM_after_registration`, `codex_SIGTERM`, `codex_SIGTERM_during_handshake`) PASS, exit 0. Windows에서는 `runtime.GOOS == "windows"` 로 건너뛴다(이유: POSIX 시그널 전달 시험이며 Windows 빌드는 AC-MH-011이 증명) |
| AC-MH-011 | REQ-MH-013 | Given 시그널 도우미와 변경 코드가 반영된 트리 When Windows 대상으로 빌드·vet하고 `syscall.` 참조를 훑으면 Then 빌드와 vet가 통과하고, `managed_*.go` 의 `syscall.` 참조는 0행이며, `launch_signals.go` 의 `syscall.` 참조는 `SIGTERM`·`SIGHUP` 상수뿐이다 | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 · `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → exit 0 · `grep -rn 'syscall\.' internal/cli/managed_*.go` → 0행, exit 1 · `grep -n 'syscall\.' internal/cli/launch_signals.go` → `syscall.SIGTERM`, `syscall.SIGHUP` 만(양성 대조: 이 형태가 같은 두 줄을 실제로 낸다), exit 0 |
| AC-MH-012 | REQ-MH-014 | Given sync 단계가 끝난 트리 When 운영 문서와 CHANGELOG를 훑으면 Then 해결된 세 한계 문장과 "관리 계층은 `syscall` 을 쓰지 않는다" 문장과 "후속 카드 t1409 대상" 단락이 문서에서 사라지고(정정된 `syscall` 문장은 시그널 상수 한정을 `launch_signals.go` 와 함께 말한다), 남은 한계가 고정 앵커 문자열로 문서에 있으며, CHANGELOG에 이 SPEC 엔트리가 F3·F4·F5 해결과 남은 한계를 함께 적는다 | 아래 §1.2 의 grep 목록(각 줄이 기대하는 개수·exit 코드 포함) |
| AC-MH-013 | REQ-MH-015 | Given run 단계의 커밋 이력 When RED 기준선 커밋과 수리 커밋의 선후를 조회하면 Then 재현 시험 8개(F3 4개·F4 1개·F5 3개)와 그 RED 출력 원문인 **추적되는** `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 가 한 RED 커밋에 있고, 그 커밋은 대응하는 모든 수리 커밋의 조상이다. `.moai/reports/t1409/` 는 로컬 사본일 뿐 근거로 인용하지 않는다 | `git log --reverse --format=%h:%s develop..HEAD` → RED 커밋 줄이 수리 커밋 줄들보다 앞 · `git merge-base --is-ancestor <RED> <FIX>` → exit 0, 반대 방향은 exit 1(쌍으로 확인) · `git show --stat --format=%h <RED>` → 시험 파일들과 `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 가 같은 커밋에 있음 · `git ls-files .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` → 그 경로 1행 · `git check-ignore -v .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` → 출력 0행, exit 1 (양성 대조: 같은 명령을 `.moai/reports/t1409/red-baseline.md` 에 주면 `.gitignore:235:.moai/reports/*` 를 내고 exit 0 — 측정) (SHA는 progress.md §E.2 에 인용) |
| AC-MH-014 | REQ-MH-016 | Given 카드 브랜치 When 부모 SPEC 디렉터리와 브로커 저장소의 변경을 조회하면 Then diff가 비어 있다 | `git diff --stat develop...HEAD -- .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 internal/factorymsg/store.go` → 출력 0행, exit 0 · 양성 대조: `git diff --stat develop...HEAD -- internal/cli` 는 M2 이후 1행 이상을 낸다 |
| AC-MH-015 | REQ-MH-001..016 (게이트) | Given 변경이 끝난 트리 When 형식·정적 분석·스코프 회귀를 돌리면 Then 오류 0건이고 회귀 없음 | §1.1 의 AC-MH-015 블록 — 선택 수 확인(기준 66개 최상위, 줄어들면 실패, 새 시험 이름이 모두 `Managed` 를 포함하므로 M1 뒤에는 66+8 이상) 뒤 `ok`, `[no tests to run]` 는 실패 · `gofmt -l internal/cli internal/config` → 출력 0행 · `go vet ./internal/cli ./internal/config` → exit 0 · `golangci-lint run ./internal/cli/... ./internal/config/...` (CI 판 v2.1.6; 측정: 로컬 `golangci-lint --version` = v2.1.6) → exit 0 · `go test ./internal/config -count=1` → `ok` (측정 기준선 `ok … 31.1s`) · `go test ./internal/guardstate ./internal/template -count=1` → §3 기준선 대비 신규 실패 0건 |
| AC-MH-016 | REQ-MH-012 | Given 두 소유자와 `Close` 가 design.md D-3.1 의 O5 이전, O6–O8 도중, O10(준비 대기) 중, O11 다이얼 성공 직후·O12 이전, O12 직후, 그리고 `Start` 와 경합하는 시점에 올 때 When 경합 검출기로 돌리면 Then §1.5 표의 하위 케이스별 기대 결과(Start 결과·Close 결과)와 공통 사후 조건(소유한 자식이 reaped, 토큰 임시 디렉터리 부재, 열린 연결 부재, 정리 본문 실행 정확히 1회, 데이터 경합 0건)이 성립한다 | `go test -race ./internal/cli -run '^TestManagedSessionStartCloseLifecycle$' -count=1 -v` → 하위 10개(`codex_before_start`, `codex_token_written`, `codex_child_spawned`, `codex_during_handshake`, `codex_dialed`, `codex_conn_recorded`, `codex_racing_start`, `stream_before_start`, `stream_child_spawned`, `stream_racing_start`) PASS, `WARNING: DATA RACE` 0건, `[no tests to run]` 없음, exit 0 (선행 `-list` 가 최상위 1개를 내야 함) |

### §1.1 파이프를 담은 명령 원문 (표 밖, 바이트 단위 원문)

**AC-MH-004**

```
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -list '^(TestManagedCodexConcurrentWrites|TestManagedSessionCloseConcurrent)$'
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test -race ./internal/cli -run '^(TestManagedCodexConcurrentWrites|TestManagedSessionCloseConcurrent)$' -count=1 -v
```

**AC-MH-006**

```
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -list '^(TestManagedTurnFailureClassification|TestManagedBrokerNameMatchesApprovalArgs)$'
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^(TestManagedTurnFailureClassification|TestManagedBrokerNameMatchesApprovalArgs)$' -count=1 -v
```

**AC-MH-015** (스코프 회귀 — `Managed`·`managed` 를 이름에 가진 최상위 테스트)

```
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -list '^.*(Managed|managed).*$' | grep -c '^Test'
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^.*(Managed|managed).*$' -count=1
```

기대: 첫 줄이 66 이상(측정: 기준 트리 66), 둘째 줄이 `ok` 이고 `[no tests to run]` 가 없다.

### §1.2 AC-MH-012 의 grep 목록 (원문, sync 후 트리에서)

| 명령 | 기대 |
|---|---|
| `grep -c "런처는 시그널을 처리하지 않는다" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1`·exit 0 — 측정) |
| `grep -c "턴 하나가 실패하면 세션 전체가 끝난다" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1` — 측정) |
| `grep -c "서버가 먼저 보내는 승인·질문 요청에는 답하지 않는다" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1` — 측정) |
| `grep -c "관리 계층은 .syscall. 을 쓰지 않는다" .moai/docs/factory-managed-session.md` (점 두 개가 원문의 백틱을 대신한다) | `0`, exit 1 (기준 트리 `1`, 62행 — 측정) |
| `grep -c "후속 카드 t1409 대상" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1`, 71행 — 측정) |
| `grep -c "launch_signals.go" .moai/docs/factory-managed-session.md` | 1 이상(정정된 문장이 시그널 상수 한정 예외를 명시) |
| 남은 한계 앵커 7개 각각 `grep -c "<앵커>" .moai/docs/factory-managed-session.md` | 각 1 이상: `TTL까지 재배달`, `턴 타임아웃은 세션을 끝낸다`, `거부된 elicitation`, `쓰기 데드라인`, `두 번째 시그널`, `후손 프로세스`, `Windows 콘솔` |
| `grep -c "SPEC-FACTORY-MANAGED-HARDEN-001" CHANGELOG.md` | 1 이상(기준 트리 `0`, exit 1 — 측정) |
| 라이브 미관측 문장: `grep -c "실제 codex 세션에서는 관측하지 않았다" .moai/docs/factory-managed-session.md` | 1 이상 |

(앵커 문자열은 문서 문장이 반드시 포함해야 하는 고정 부분 문자열이다. "쓰기 데드라인" 앵커의 문장은 막힌 쓰기가 `Close` 나 연결 종료까지 풀리지 않는다는 한계를 말해야 하며 턴 타임아웃이 풀어 준다고 쓰면 안 된다(design.md 공시한 한계). 문장의 나머지는 manager-docs가 쓰되 과장 없이 쓴다. CHANGELOG 엔트리는 같은 앵커의 의미를 영어 또는 한국어로 옮겨 F3·F4·F5 "resolved"와 위 한계를 함께 적는다.)

### §1.3 AC-MH-010 하위 케이스 (design.md D-3.1 의 O번호·훅 단계명)

공통 기대: 시그널을 보낸 뒤 5초 안에 소유자 프로세스가 종료하고, 종료 오류 문자열(stderr)이 `interrupted` 를 담으며, 부모 시험이 조회한 launch-pending 행이 0건이다. 도우미는 `ready <단계명>` 줄을 stdout 에 쓰고 부모는 그 줄이 보일 때까지만 기다린다(상한 10초, 넘으면 시험 실패).

| 하위 케이스 | 시그널을 보내는 시점(훅 단계명, O번호) | 단계별 관측 기대(공통 기대에 더해) |
|---|---|---|
| `stream_SIGTERM` | `driver-running` (O16) | 가짜 백엔드 자식 프로세스 부재(reaped) |
| `stream_SIGHUP` | `driver-running` (O16) | 같음 |
| `stream_SIGINT` | `driver-running` (O16) | 같음 |
| `stream_SIGTERM_after_registration` | `registered` (O3) — 훅이 `ctx.Done()` 까지 대기해 등록 직후·`Start` 이전에 신호가 들어간다 | 가짜 백엔드가 **한 번도 기동되지 않음**(PID 파일 부재), launch-pending 행 0건(롤백이 돌았다), 종료 오류에 `interrupted` |
| `codex_SIGTERM` | `driver-running` (O16) | 가짜 App Server 자식 부재, 토큰 임시 디렉터리 부재 |
| `codex_SIGTERM_during_handshake` | `start-published` (O8 직후) — 가짜 App Server가 `/readyz` 에 영영 답하지 않음 | 10초 준비 예산을 기다리지 않고 5초 안에 종료(취소 연결 O10), 자식 부재, 토큰 임시 디렉터리 부재 |

### §1.4 AC-MH-006 하위 케이스 (`TestManagedTurnFailureClassification` 14개)

| # | 하위 케이스 | 입력 | 기대 |
|---|---|---|---|
| 1 | `stream_is_error_after_priming` | 우선 턴 성공 뒤 `result.is_error` | 턴 단위: 드라이버 계속, 연속 횟수 +1 |
| 2 | `codex_failed_or_interrupted` | `turn/completed` 상태 `failed` 와 `interrupted`(두 값을 돌려 단언) | 턴 단위 |
| 3 | `codex_moai_elicitation_turn_completed` | 턴 중 `serverName=moai` elicitation, 이어서 상태 `completed` | 턴 단위(`DeliverTurn` 이 표식 오류), 로그 줄 `turn=<id>` |
| 4 | `priming_is_error` | 우선 턴이 `is_error` | 세션 치명: 드라이버가 그 오류를 반환 |
| 5 | `stream_closed` | `errManagedStreamClosed` | 세션 치명 |
| 6 | `codex_connection_closed` | "connection closed" | 세션 치명 |
| 7 | `write_failure` | 자식 stdin 쓰기 실패 | 세션 치명 |
| 8 | `codex_timeout` | 턴 컨텍스트 마감 초과 | 세션 치명 |
| 9 | `unclassified_error` | 표식 없는 일반 오류 | 세션 치명 |
| 10 | `other_server_elicitation` | 턴 중 `serverName` 이 브로커가 아닌 elicitation, 이어서 `completed` | 실패로 세지 않음: `DeliverTurn` nil, 로그만 |
| 11 | `between_turns_then_normal_turn` (a) | 턴 사이(창이 닫힌 때)에 `moai` elicitation, 이어서 정상 턴 | 정상 턴 `DeliverTurn` nil, 실패 0, 로그 줄 `turn=none` |
| 12 | `normal_turn_after_declined_turn` (b) | `moai` elicitation 이 있었던 턴 다음의 정상 턴 | nil(초기화는 `armTurn`) |
| 13 | `elicitation_after_turn_start_response` (c) | `turn/start` 응답 직후, `turn/started` 이전에 `turnId` 없는(null) `moai` elicitation | 그 턴을 실패시킴(턴 단위) |
| 14 | `two_requests_in_one_turn` (d) | 한 턴 안에 `moai` elicitation 2건 | 정확히 한 번의 턴 단위 실패(`DeliverTurn` 오류 1개, 드라이버 연속 횟수 +1) |

`TestManagedBrokerNameMatchesApprovalArgs` 는 `factoryMoAIMCPApprovalArgs()` 의 모든 서버 이름 접두가 `mcp_servers.` + `moaiMCPServerKey` + `.` 와 같음을 단언한다(두 이름이 한 값임을 고정; 승인 인수 코드는 수정하지 않는다).

### §1.5 AC-MH-016 하위 케이스 (`TestManagedSessionStartCloseLifecycle` 10개)

공통 사후 조건(모든 하위 케이스): 소유한 자식 프로세스가 **reaped**(`Wait` 완료 + `os.Process.Signal` 이 `os.ErrProcessDone`), 토큰 임시 디렉터리 부재(`moai-factory-app-` 접두 디렉터리 집합이 시작 전과 같음; Codex), 열린 연결 부재, `teardown` 훅 횟수 정확히 1, `-race` 경합 0건. 훅 단계명은 design.md D-3.1 의 열이다.

| 하위 케이스 | 정지 지점(훅, O번호)과 `Close` 도착 | Start 기대 | Close 기대 | 이 케이스 고유 관측 |
|---|---|---|---|---|
| `codex_before_start` | `Start` 를 부르기 전에 `Close` 완료 (O5 이전) | 닫힘 오류(`errors.Is` 로 `errManagedSessionClosed`) | nil | 가짜 App Server 가 한 번도 기동되지 않음(로그에 `token-loaded` 없음), 토큰 디렉터리가 생성되지 않음 |
| `codex_token_written` | `start-token-written`(O6)에서 정지, 그 사이 `Close` 시작 | 훅 해제 뒤 닫힘 오류 | nil | **`Close` 는 훅이 해제되기 전에 반환하지 않는다**(L 대기; 임계구역 O6–O8 의 관측), 해제 뒤 게시된 자식·토큰 디렉터리를 치움 |
| `codex_child_spawned` | `start-child-spawned`(O7)에서 정지, 그 사이 `Close` 시작 | 훅 해제 뒤 닫힘 오류 | nil | 위와 같은 L 대기 관측. `start-published`(O8) 훅은 시험이 `Close` 반환을 확인할 때까지 `Start` 를 붙잡는다(결정성) |
| `codex_during_handshake` | `start-published` 이후 가짜 서버가 `/readyz` 에 답하지 않음, 200ms 뒤 `Close` (O10) | 닫힘 오류, **`Close` 반환 뒤 2초 이내**에 반환 | nil | 시험 안 벽시계 상한 2초(10초 준비 예산보다 한참 작음; Windows 포함 성립) |
| `codex_dialed` | `start-dialed`(O11)에서 정지, `Close` 완료 | 훅 해제 뒤 닫힘 오류 | nil(연결은 아직 대상 아님) | 훅이 받은 `*websocket.Conn` 에 해제 뒤 쓰기를 시도하면 **실패**(Start 가 자기 연결을 닫았다) |
| `codex_conn_recorded` | `start-conn-recorded`(O12)에서 정지, `Close` 완료 | 훅 해제 뒤 닫힘 오류(핸드셰이크 RPC 의 "connection closed" 가 O14 정규화로 바뀜) | nil | 오류가 "connection closed" 가 아니라 닫힘 오류 |
| `codex_racing_start` | `Start` 와 `Close` 를 별개 고루틴에서 **N=40회** 반복, 반복 i 의 `Close` 지연 i×1ms(0–39ms) | 반복마다 결과가 **훅 이벤트 순서(시험이 단조 증가 번호로 기록)로 결정**된다: `close-snapshot`(O18)이 `Start` 의 반환보다 먼저면 닫힘 오류, 나중이면 nil. 두 결과 모두 같은 사후 조건 | 반복마다 nil | 반복마다 사후 조건 전부. `Start` 가 nil 이었다면 이후 `DeliverTurn` 은 오류. 반복 수 근거: 창이 수 ms 폭이고 스케줄 의존이라 결정적 훅으로 닿지 않는 틈을 지연 격자로 훑는다(Codex 자식 기동 비용 때문에 40) |
| `stream_before_start` | `Start` 전에 `Close` 완료 | 닫힘 오류 | nil | 백엔드 미기동(PID 파일 부재), 생성자가 만든 stdin 파이프에 쓰기가 실패(Close 가 파이프를 닫았다) |
| `stream_child_spawned` | `start-child-spawned`(O7)에서 정지, 그 사이 `Close` 시작 | **nil**(게시 후 반환; 이후 닫힘은 Start 결과가 아님) | nil | **`Close` 는 훅 해제 전에 반환하지 않는다**, 해제 뒤 자식 kill+`Wait`, 이후 `DeliverTurn` 은 오류 |
| `stream_racing_start` | 위와 같은 경합 반복 **N=200회**, 반복 i 의 `Close` 지연 i×100µs(0–20ms) | 반복마다 결과가 **훅 이벤트 순서로 결정**된다: `close-snapshot`(O18)이 `start-published`(O8)보다 먼저면 닫힘 오류(게시되지 않음), 나중이면 nil(게시 뒤 닫힘은 `Start` 결과가 아니다) | 반복마다 nil | 반복마다 사후 조건 전부. 반복 수 근거: 백엔드가 가벼운 스크립트라 200회가 수 초 안이고 창이 마이크로초 폭이다 |

## §2. RED-now / green path 쌍과 채택 상태 (verification-completeness.md §2)

`.claude/rules/moai/development/verification-completeness.md` §2, §2.1 의 두 칸 채택 규율을 따른다. 기준 트리 `7109e0900`.

### §2.1 시험 기반 릴리스 차단급 AC — **plan 시점에는 미채택, M1의 RED 기준선이 채택한다**

AC-MH-001, 002, 003, 004, 005, 006, 010, 016(및 그 하위 케이스 전부)은 시험이 아직 없어서 plan 단계에서 RED-now 네 요소(명령·stdout·exit·트리 SHA)를 시험 명령으로 낼 수 없다. 이 AC들은 **plan 시점에 미채택이며 M1의 RED 기준선(`red-baseline.md`)이 채택하는 것**으로 run에 넘긴다(plan-audit PASS-with-debt 경로). 계획 단계의 임시 시험 실측은 삭제되어 재실행할 수 없으므로 아래 "보조 실측"은 RED-now 셀이 아니다.

**선결 조건**: M1의 산출물이 착지할 수 있어야 한다. 산출물 경로는 추적되는 SPEC 디렉터리 안 `red-baseline.md` 다(측정: `git check-ignore -v .moai/reports/t1409/red-baseline.md` → `.gitignore:235:.moai/reports/*`, exit 0 — 착지 불가; `git check-ignore -v .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` → 출력 없음, exit 1; `internal/runtime/audit_cache.go` 의 `planArtifactNames` 는 `acceptance.md`·`design.md`·`plan.md`·`research.md`·`spec.md`·`tasks.md` 뿐이라 이 파일은 감사 캐시 해시에 들지 않는다). `red-baseline.md` 는 run 단계 증거 파일이며(manager-develop이 쓴다) SPEC 본문이 아니다.

**"옳은 이유의 RED" 규정 (한 방향으로)**: 컴파일 오류, 패키지·하네스 수준 타임아웃, 시험 인프라 실패로 붉은 것은 wrong-reason RED 이고 RED 기준선으로 인용하지 않는다. 반면 **시험 자신이 정한 시간 상한 단언**(예: `codex_during_handshake` 의 "`Close` 뒤 2초 안에 `Start` 가 반환해야 한다")이 깨져서 붉은 것은 그 상한이 곧 단언 대상 관측이므로 **옳은 이유의 RED** 다. 두 경우는 `red-baseline.md` 의 "붉은 이유" 줄에 어느 쪽인지 적는다.

**M1 재현 시험 8개와 옳은 이유의 RED (기준 트리 API에 대해 컴파일되는 형태)**

| 시험 | AC | RED가 붉어야 하는 이유(명시) | green path |
|---|---|---|---|
| `TestManagedCodexServerRequestPolicy` (가짜 서버가 직접 서버 요청 11종을 보내고 응답 수신을 단언 — 새 심볼 없이) | AC-MH-001 | 소유자가 답장을 보내지 않아 가짜 서버가 유한 대기 뒤 응답 없음을 기록 | M2 |
| `TestManagedCodexTurnSurvivesServerRequest` | AC-MH-002 | 같은 이유(턴이 답 없이 진행, 5초 상한 단언 실패) | M2 |
| `TestManagedCodexServerRequestIDCollision` (프로세스 내 클라이언트 + 임시 서버) | AC-MH-003 | id 충돌 시 호출이 서버 요청을 응답으로 반환(빈 결과·오류 없음), 문자열 id면 연결 닫힘 오류 | M2 |
| `TestManagedCodexDeclinedBrokerElicitationFailsTurn` (가짜 서버가 `serverName=moai` elicitation을 `turnId` 없이 보내고 턴을 `completed` 로 끝냄, 같은 턴에 둘째 요청도 보내는 변형; 단언은 `DeliverTurn` 이 **nil 이 아닌 오류**를 돌려주는지뿐 — 표식 심볼 없이. §1.4 (c)·(d) 의 컴파일 가능 형태) | AC-MH-006 | 기준 트리의 `DeliverTurn` 이 nil 을 반환 | M3 |
| `TestManagedDriverIsolatesTurnFailure` (실제 `pumpManagedStreamTurn` 을 쓰는 스크립트 세션) | AC-MH-005 | 두 번째 턴의 에러 결과로 드라이버가 반환해 세 번째 턴이 전달되지 않음 | M3 |
| `TestManagedLauncherSignalTeardown` 의 기준 컴파일 가능 하위 케이스 전부(§1.3; 훅 없이 `driver-running` 대신 "자식이 떴음"을 PID 파일로 확인하는 형태로 쓴다) | AC-MH-010 | 시그널을 받은 프로세스가 `defer` 정리 없이 끝나 자식이 살아 있고 임시 디렉터리·launch-pending 행이 남음 | M4 |
| `TestManagedSessionCloseConcurrent` (`-race`) | AC-MH-004 (Close 부분) | 보호되지 않은 `closed` 와 이중 `cmd.Wait` — 경합 보고 또는 `Wait was already called` | M4 |
| `TestManagedSessionStartCloseLifecycle` 의 훅 없이 쓸 수 있는 하위 5개(`codex_before_start`, `codex_during_handshake`, `codex_racing_start`, `stream_before_start`, `stream_racing_start`) | AC-MH-016 | `Close` 이후의 `Start` 가 성공하며 자식·토큰 디렉터리가 남고, 핸드셰이크 대기가 `Close` 로 풀리지 않아 **시험의 2초 상한 단언이 깨지고**, 경합 보고 | M4 |

**RED-first 면제(이유 명시)**: (1) AC-MH-004 의 쓰기 경합 부분(`TestManagedCodexConcurrentWrites`)은 기준 트리의 읽기 고루틴이 쓰지 않아 쓰기 경합이 존재하지 않는다(경합 검출기가 구조상 초록). (2) AC-MH-016 의 `codex_token_written`, `codex_child_spawned`, `codex_dialed`, `codex_conn_recorded`, `stream_child_spawned` 와 AC-MH-010 의 훅 기반 하위 케이스(`stream_SIGTERM_after_registration`, `codex_SIGTERM_during_handshake` 의 단계 정지)는 시험 훅(design.md D-3.1 "시험 이음새")이라는 새 이음새가 필요해 기준 트리에서 구성할 수 없다. (3) AC-MH-006 의 §1.4 (a)·(b) 와 다른 서버 이름 행은 "과잉 계수하지 않는다"는 부정 요구라 기준 트리(계수 자체가 없음)에서 초록이다 — 불변 가드다. 이 셋의 채택은 **변이 확인**(§2.4)으로 대신하고 원문을 progress.md §E.2 에 인용한다. (4) AC-MH-009 는 취소 컨텍스트라는 새 인자가 필요해 기준 트리에서 컴파일되지 않는다 — M4에서 GREEN으로만 채택하고, 같은 결함은 AC-MH-010 의 RED가 구성한다. (5) AC-MH-008 은 "관리 계층이 아무것도 하지 않는다"는 부정 요구라 기준 트리에서도 초록이다 — 불변 가드이며(§2.3), 재배달 증거는 기존 브로커 시험이 맡는다. (6) AC-MH-007 의 시험은 새 상수가 필요해 기준 트리에서 컴파일되지 않는다 — §2.2 의 구조 확인(상수 부재 exit 1)과 변이 확인으로 채택한다.

**보조 실측(기준 트리, 임시 시험은 삭제되어 재실행 불가 — RED-now 셀이 아님)**: 서버 요청 id=7을 보낸 뒤 1초간 클라이언트 프레임 없음(`no client frame within 1s: … i/o timeout`) · id 충돌: `call result="" err=<nil>`(진짜 응답 `{"genuine":true}` 소실) · 문자열 id: `err=managed codex app server connection closed` · 별도 작은 프로그램: SIGTERM `launcher_rc=143 child=alive tokendir=present defer=not-run`, SIGHUP `launcher_rc=129 …`(SIGINT 미관측).

### §2.2 구조 확인형 AC (지금 명령으로 RED-now를 낼 수 있는 것, 재실행 확인됨)

| AC | RED-now: 명령 | stdout | exit | 트리 | green path |
|---|---|---|---|---|---|
| AC-MH-007 (상수 부분) | `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` | (빈 출력) | 1 | `7109e0900` | M3 이 상수를 정의하면 정의 1행·exit 0 |
| AC-MH-010 (도우미 존재) | `grep -n NotifyContext internal/cli/launch_signals.go` | `grep: internal/cli/launch_signals.go: No such file or directory`(이 머신에서는 `ugrep:` 접두 — 별칭 차이) | 2 | `7109e0900` | M4 가 도우미를 만들면 해당 줄·exit 0 (양성 대조: `grep -n NotifyContext internal/cli/mcp_server.go` → 123행, exit 0) |
| AC-MH-012 | `grep -c "런처는 시그널을 처리하지 않는다" .moai/docs/factory-managed-session.md` | `1` | 0 | `7109e0900` | sync 가 문장을 지우면 `0`·exit 1 |
| AC-MH-012 (CHANGELOG) | `grep -c "SPEC-FACTORY-MANAGED-HARDEN-001" CHANGELOG.md` | `0` | 1 | `7109e0900` | sync 가 엔트리를 쓰면 1 이상·exit 0 |

### §2.3 회귀 가드 (지금 이미 초록이 정상 — 릴리스 차단 RED가 아니라 불변 가드)

AC-MH-011(Windows 빌드: 기준 트리 `GOOS=windows GOARCH=amd64 go build ./...` 출력 없음·성공, `grep -rn 'syscall\.' internal/cli/managed_*.go` 0행·exit 1), AC-MH-014(기준 트리 diff 0행), AC-MH-008(기준 트리에서도 초록인 부정 요구 + `lost_receipt_redelivery` 기준 PASS 측정), AC-MH-015.

### §2.4 변이 탐침 — 구체적 변이와 반드시 붉어져야 하는 하위 케이스

변이는 run 단계의 M4(및 M2·M3)에서 **하나씩 임시로 적용**해 지목한 하위 케이스가 붉어지는지 보이고(원문을 progress.md §E.2 에 인용) 되돌린 뒤 `git diff --stat` 빈 출력을 확인한다. 위치는 design.md D-3.1 의 O번호다.

| 변이 | 위치 | 반드시 붉어지는 하위 케이스(이유) |
|---|---|---|
| m1 `Start` 진입의 `closed` 확인 제거(`MkdirTemp` 이전) | O5 | `codex_before_start`, `stream_before_start` (`Start` 가 성공하며 자식·토큰 디렉터리를 만든다) |
| m2 임계구역 해체: `cmd.Start`/게시를 L 밖으로, 또는 게시를 L 해제 뒤로 | O6–O8 | `codex_token_written`, `codex_child_spawned`, `stream_child_spawned` (`Close` 가 훅 해제 전에 반환하거나, 기록되지 않은 자식·토큰 디렉터리가 샌다) |
| m3 연결 기록 시점의 `closed` 재확인 제거(다이얼과 기록 사이) | O12 | `codex_dialed` (닫힌 뒤에도 `Start` 가 연결을 기록·사용한다 → 훅이 받은 연결에 쓰기가 성공) |
| m4 `Start` 가 닫혔을 때 자기 연결을 닫지 않고 오류만 반환 | O12 | `codex_dialed` (훅이 받은 연결에 쓰기가 성공) |
| m5 `Close` 가 토큰 디렉터리를 치우지 않음 | O19 | `codex_token_written`, `codex_child_spawned`, `codex_during_handshake`, `codex_dialed`, `codex_conn_recorded`, `codex_racing_start`, AC-MH-010 `codex_SIGTERM`·`codex_SIGTERM_during_handshake` |
| m6 `Close` 가 자식 kill+`Wait` 를 하지 않음 | O19 | `codex_child_spawned`, `codex_during_handshake`, `codex_dialed`, `stream_child_spawned`, 두 `*_racing_start` (reaped 단언 실패) |
| m7 준비 취소 함수 연결 제거 | O10, O19 | `codex_during_handshake` (2초 상한 단언 실패), `codex_token_written`, `codex_child_spawned` |
| m8 `Close` 가 호출마다 정리(이중 정리) | O17 | 두 `*_racing_start` (`teardown` 횟수 2 이상), AC-MH-004 동시 `Close` |
| m9 `Close` 가 첫 정리 완료를 기다리지 않고 반환 | O17 | AC-MH-004 동시 `Close`, 두 `*_racing_start` (사후 조건 시점에 자식이 아직 살아 있음) |
| m10 `Start` 반환 정규화 제거 | O14 | `codex_conn_recorded` (오류가 닫힘 오류가 아니라 "connection closed") |
| m11 `Close` 가 게시되지 않은 stream 파이프를 닫지 않음 | O19 | `stream_before_start` (stdin 파이프 쓰기가 성공) |
| m12 시그널 컨텍스트 생성을 launch-pending 등록 뒤로 옮김 | O1 대 O3 | AC-MH-010 `stream_SIGTERM_after_registration` (핸들러가 없어 기본 동작으로 종료 → launch-pending 행 잔존) |
| m13 소유자 진입의 오류 → 중단 매핑 제거 | O20 ④ | AC-MH-010 모든 하위 케이스, 특히 `codex_SIGTERM_during_handshake` (종료 오류에 `interrupted` 없음) |
| m14 elicitation 판정을 소비자 쪽 계수기·`startTurn` 리셋으로 되돌림 | D-1 결정 3 | AC-MH-006 #11 (a) (정상 턴이 실패로 집계) |
| m15 창 초기화(`armTurn`) 제거 | D-1 결정 3 | AC-MH-006 #12 (b) |
| m16 열린 창 대신 `turn/started` 이후에만 귀속 | D-1 결정 3 | AC-MH-006 #13 (c) |
| m17 요청마다 실패로 세기(불리언이 아니라 횟수) | D-1 결정 3 | AC-MH-006 #14 (d) |
| m18 모든 `serverName` 의 거부를 센다 | D-1 결정 3 | AC-MH-006 #10 |
| m19 구현에 상수 값 `3` 을 박음(대신 상수를 2로 바꾸면 시험이 따라오는지 확인) | D-2 결정 2 | AC-MH-007 (상수를 읽는 시험이 상수 변경 시 깨져야 함) |
| m20 쓰기 뮤텍스 제거 | D-1 결정 1 | AC-MH-004 `TestManagedCodexConcurrentWrites` (`-race` 에서 경합 보고) |
| m21 승인류 응답을 `accept` 로 변경 | D-1 결정 2 | AC-MH-001 해당 하위 케이스 |
| m22 `Close` 대신 `os.Exit` 로 끝냄 | D-3 구성 | AC-MH-010 (토큰 디렉터리·launch-pending 행 잔존) |
| m23 RED와 수리를 한 커밋에 합침 | — | AC-MH-013 (`merge-base --is-ancestor` 가 정·역방향 모두 0) |

## §3. 기준선과 알려진 사전 적색 (측정)

| 대상 | 관측 (기준 트리 `7109e0900`) |
|---|---|
| 스코프 회귀(정리 접두) | `ok  github.com/modu-ai/moai-adk/internal/cli  27.048s` (앵커형), 리터럴 부분 문자열 패턴으로 돌린 첫 측정은 `ok … 30.959s`. `-v` 최상위 66개, 두 패턴이 같다 |
| 같은 선택, 정리 접두 없이(레인 환경) | `--- FAIL: TestManagedSwitchDoesNotReachCodexLaneLoop (0.48s)` — `todo add: moai add: refused — lane boundary: a lane session cannot mutate the queue …`, `FAIL … 26.413s`. 레인 환경이 원인(정리 후 PASS 13.40s). 이 카드가 만든 적색이 아니다 |
| 선택 수: 이스케이프 없는 형태 대 표 이스케이프 형태 | 명령은 아래 블록. 이스케이프 없는 형태 66, 백슬래시 파이프 형태 0 (plan-audit 측정을 재현) |
| AC-MH-004 패턴의 선택 수(시험 부재) | §1.1 AC-MH-004 블록의 첫 명령(`-list`)을 기준 트리에서 돌린 출력이 `ok  github.com/modu-ai/moai-adk/internal/cli  0.874s` 한 줄 — `Test` 이름 0개(시험 부재 → 공허 `ok`의 실례) |
| config | `ok  …/internal/config  31.134s` |
| template | `ok  …/internal/template  414.986s` (분 단위 스위트 → 임대 안에서만 실행) |
| guardstate | **기준 트리에서 이미 적색**: `--- FAIL: TestCensus_SetDifferenceEmptyBothDirections (0.00s)` — `census_test.go:80: disk\manifest: .github/workflows/workflow-parse-guard.yaml exists on disk with no manifest entry`, `census_test.go:89: declared 19 entries against 20 workflow files`. 이 카드는 워크플로 파일을 건드리지 않으므로 사전 존재이며, AC-MH-015 는 "이 한 테스트 외 신규 실패 0건"으로 읽는다 |
| 템플릿 미러 | `ls internal/template/templates/.moai/docs/factory-managed-session.md` → `No such file or directory`(미러 없음) |

선택 수 측정 명령 원문(두 번째 줄의 백슬래시 파이프는 일부러 이스케이프한 반증 형태):

```
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -list '^.*(Managed|managed).*$' | grep -c '^Test'
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -list '^.*(Managed\|managed).*$' | grep -c '^Test'
```

## §4. 에지 케이스

- 서버 요청이 턴 사이(한가한 때)에 온다 — 읽기 고루틴이 바로 답한다(design.md D-1 이유 1; AC-MH-002 `between_turns`). `moai` elicitation이 그 때 오면 어느 턴에도 세지 않는다(AC-MH-006 #11).
- 서버 요청 두 개가 연달아 온다 — 각각 정확히 한 번 답하고 순서가 보존된다.
- 답장 쓰기가 실패한다 — 읽기 고루틴이 끝나고 진행 중 `DeliverTurn` 은 "connection closed" 로 세션 치명 오류가 된다.
- 알 수 없는 알림(id 없음)은 지금처럼 버린다. 답하지 않는다. `id: null` 프레임은 분류되지 않아 버려진다(공시한 한계).
- 늦게 도착한 이전 턴의 elicitation(`turnId` 가 현재 턴과 다름)은 세지 않는다.
- 운영자 입력으로 시작된 턴이 실패해도 같은 횟수에 센다. 성공한 운영자 턴은 횟수를 0으로 되돌린다.
- 우선 턴이 `is_error` 로 끝나도 세션 치명이다(턴 단위 분류는 우선 턴 이후에만 적용: REQ-MH-006).
- 시그널이 등록 직후·`Start` 이전, `Start()` 의 핸드셰이크 중, 드라이버 루프 중 어디서 와도 launch-pending 행이 남지 않는다(AC-MH-010 하위 케이스; 순서는 design.md D-3.1).
- 정리 중 두 번째 시그널이 온다 — 첫 시그널에서 기본 동작을 복원하므로 프로세스가 끝난다(탈출구; 자동 시험으로 고정하지 않음 — Gap).
- `Close` 가 이미 끝난 뒤 시그널이 온다 — 두 번째 `Close` 는 같은 결과를 돌려주고 아무것도 다시 하지 않는다.
- 스트림 소유자의 `Kill` 은 직접 자식만 죽인다. 후손은 남을 수 있다(공시한 한계; AC-MH-010 은 소유한 자식만 단언).
- 연결 쓰기에 데드라인이 없다 — 서버가 읽지 않으면 쓰기가 막히고 턴 타임아웃도 그것을 풀지 못한다. `Close`(시그널) 또는 연결 종료까지 상한이 없다(공시한 한계; 시그널 경로는 영향 없음).
- 세션이 열려 있는데 `Start` 가 `cmd.Start` 성공 전에 실패하는 분기는 지금 동작 그대로다(F8, t1410 몫). 이 카드의 시험은 그 분기의 토큰 디렉터리 잔존을 단언하지도 고치지도 않는다.
- `/exit` 와 stdin EOF 경로는 기존 동작 그대로다(AC-MH-015 스코프 회귀가 고정).

## §5. 품질 게이트

- TRUST 5 — Tested: 변경한 비테스트 코드의 줄이 재현·분류·시그널·수명주기 시험으로 덮인다(`go test -cover` 를 managed 슬라이스에 한정해 측정, 부모 기준선 89.3%보다 내려가지 않는다 — 부모 sync 기록의 수치이며 이 plan에서 재측정하지 않았다). Readable/Unified: `gofmt`·`go vet`·`golangci-lint` 0건. Secured: 승인류 응답에 accept 계열 0건(AC-MH-001), 토큰 정리 유지(AC-MH-010, 016). Trackable: Conventional Commit, 모든 커밋 메시지에 카드 id `t1409`.
- 경계 grep(부모 AC 계승): `grep -rnE 'merge-window|Decider|T29b|T29c|handover' internal/cli/managed_*.go` 0행, `grep -rn 'AskUserQuestion' internal/cli/managed_*.go` 비테스트 0행.

## §6. Definition of Done

1. AC-MH-001..016 전부 PASS — 명령과 원문 출력을 progress.md §E.2 에 인용(AC-MH-012 는 sync 후 §E.4).
2. 재현 시험 8개의 RED 원문이 **추적되는** `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 에 있고 RED 커밋이 수리 커밋의 조상이다(AC-MH-013).
3. §2.4 의 변이 m1–m22 를 **하나씩** 적용해 지목한 하위 케이스가 붉어지는 것을 보인 원문이 증거로 있다(m23 은 AC-MH-013 의 쌍 확인으로 대신). 특히 닫힘 확인 지점마다 1건씩(m1 O5, m2 O6–O8, m3·m4 O12), 토큰 디렉터리 정리(m5), 자식 kill+`Wait`(m6), 준비 취소 연결(m7), 이중 정리(m8)가 빠지지 않는다.
4. Windows 크로스 빌드·vet가 exit 0이고 `managed_*.go` 의 `syscall.` 참조가 0행이다(AC-MH-011).
5. 부모 SPEC 디렉터리와 `store.go` 의 diff가 비어 있다(AC-MH-014).
6. 운영 문서와 CHANGELOG가 해결된 세 한계와 남은 한계를 과장 없이 적고, 정정된 `syscall` 문장과 t1409 단락 삭제가 grep으로 확인된다(AC-MH-012).
