---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "acceptance.md — 인수 기준"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# acceptance.md — AC-MH-001..015

> 이 파일은 검증층이고 시나리오는 Given-When-Then이다. 요구(GEARS)는 `spec.md` §C 의 REQ-MH-001..016 이 운반한다. 모든 AC는 이진 판정이 가능하다.
>
> **명령 규약.** 검증 명령은 단일 호출형이고 `-run` 패턴은 앵커형(`^…$`)이다. 부하 규율상 `go test ./internal/cli` 전체는 로컬에서 돌리지 않는다. **레인 세션에서 `go test` 는 환경 정리 복합형으로 돌린다**: `unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test …` (실측: 기준 트리에서 정리 없이 돌린 스코프 명령은 `TestManagedSwitchDoesNotReachCodexLaneLoop` 가 레인 환경 때문에 `lane boundary` 거부로 가짜 적색이었고, 정리 후 단독 PASS 와 스코프 전체 `ok` 가 나왔다). 아래 표의 명령은 그 접두 없이 적는다.
>
> **빈 스윕 규율.** `-run` 명령은 "N개 하위 테스트가 PASS"를 기대 관측에 적는다. `no tests to run`·`[no test files]` 는 통과가 아니라 미측정이다.
>
> **측정 기준 트리**: `7109e0900`(카드 기준 develop, 코드 변경 없음). 아래 "측정" 표기는 이 트리에서 이 plan 실행이 관측한 것이다.

## §1. AC 매트릭스

| AC | 요구 | 시나리오 (Given-When-Then) | 검증 명령 → 기대 관측 |
|---|---|---|---|
| AC-MH-001 | REQ-MH-001, REQ-MH-002, REQ-MH-003 | Given 가짜 App Server가 턴 중에 서버 요청 10종(각각 정수 id)과 목록에 없는 method 하나를 차례로 보낼 When Codex 소유자가 처리하면 Then 각 요청에 정확히 한 번, 같은 id로, design.md D-1 정책표대로 답하고(승인류는 `decline`/`denied`/빈 `permissions`/`decline` 액션, 오류류는 `-32000`, 모르는 method는 `-32601`), 어느 답에도 accept 계열 결정이나 비어 있지 않은 권한 부여가 없다 | `go test ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v` → 상위 `--- PASS: TestManagedCodexServerRequestPolicy (…s)` 1줄과 그 아래 하위 테스트 11개(10종 + 미지 method 1)가 각각 PASS, exit 0 |
| AC-MH-002 | REQ-MH-001 | Given 가짜 App Server가 `turn/start` 응답 뒤 명령 승인 요청을 보내고 그 답을 받은 뒤에만 `turn/completed` 를 보낼 When `DeliverTurn` 을 부르면 Then 답이 서버에 도달하고 턴이 오류 없이 끝난다(턴 타임아웃을 기다리지 않는다) | `go test ./internal/cli -run '^TestManagedCodexTurnSurvivesServerRequest$' -count=1` → `ok`, 시험 시간이 턴 타임아웃 상수보다 한참 짧음 |
| AC-MH-003 | REQ-MH-004 | Given 서버 요청의 id가 대기 중인 클라이언트 id와 같거나 문자열(`"srv-1"`)일 When 소유자가 `turn/start` 응답을 기다리면 Then 서버 요청은 응답으로 오인되지 않고(진짜 응답 본문이 호출자에게 도달), 문자열 id의 답장에 같은 문자열이 되돌아가며, 읽기 루프가 계속 살아 이후 호출이 성공한다 | `go test ./internal/cli -run '^TestManagedCodexServerRequestIDCollision$' -count=1 -v` → 하위 2줄(`integer_collision`, `string_id`) PASS, exit 0 |
| AC-MH-004 | REQ-MH-005, REQ-MH-012 | Given 읽기 고루틴이 답장을 쓰는 동안 다른 고루틴이 `call()` 로 쓰고, 별도로 Close를 여러 고루틴이 동시에 부를 When 경합 검출기로 돌리면 Then 데이터 경합 보고가 0건이고 Close는 정리를 한 번만 수행하며 모든 호출자가 같은 결과를 받는다(스트림·Codex 두 소유자) | `go test -race ./internal/cli -run '^(TestManagedCodexConcurrentWrites\|TestManagedSessionCloseConcurrent)$' -count=1 -v` → 최상위 2개와 하위 케이스 PASS, `WARNING: DATA RACE` 0건, exit 0 (`-race` 는 cgo 필요) |
| AC-MH-005 | REQ-MH-006 | Given 우선 턴 성공 뒤 두 번째 턴이 에러 결과를 내고 세 번째 턴이 성공하는 스크립트 세션 When 드라이버가 돌면 Then 드라이버는 두 번째 턴 뒤에 반환하지 않고 세 번째 턴을 전달하며, stderr에 `Factory turn failed (1/` 로 시작하는 한 줄이 남고 stdout에는 쓰이지 않는다 | `go test ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$' -count=1 -v` → PASS, 시험이 세션이 받은 턴 3건과 stderr 한 줄을 단언 |
| AC-MH-006 | REQ-MH-006, REQ-MH-007 | Given 실패 원인별 표(스트림 `result.is_error`, Codex `failed`/`interrupted` 종료, 우선 턴 실패, 스트림 종료, "connection closed", 쓰기 실패, Codex 타임아웃, 표시되지 않은 일반 오류) When 소유자와 드라이버를 통과시키면 Then 앞의 두 원인만 턴 단위(드라이버 계속)이고 나머지는 전부 첫 발생에서 드라이버가 그 오류를 반환한다 | `go test ./internal/cli -run '^TestManagedTurnFailureClassification$' -count=1 -v` → 원인별 하위 8줄 PASS, exit 0 |
| AC-MH-007 | REQ-MH-008 | Given 연속 턴 단위 실패 횟수가 `config.DefaultManagedSessionMaxConsecutiveTurnFailures` 와 같을 때 When 드라이버가 그 횟수째 실패를 받으면 Then 마지막 오류를 횟수와 함께 반환한다. 상한−1번 실패 뒤 성공 턴이 오면 횟수가 0으로 돌아가 다음 실패가 1번째로 센다. 시험은 상수 자체를 읽어 횟수를 정한다(구현에 숫자를 박아도 시험이 깨진다) | `go test ./internal/cli -run '^TestManagedDriverConsecutiveFailureCeiling$' -count=1 -v` → 하위 3줄(`at_ceiling_returns`, `success_resets`, `below_ceiling_continues`) PASS · `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` → 정의 1행, exit 0 |
| AC-MH-008 | REQ-MH-009 | Given 브로커에 메시지 1건이 claim되고 그 메시지를 실은 턴이 실패할 When 드라이버가 세션을 이어 가면 Then 관리 계층은 그 행을 `claimed` 로 두고(`acknowledged` 0, claim token 불변) 해제·재주소를 하지 않는다. 재배달은 lease 만료 뒤 브로커의 `Claim` 이 새 token으로 한다 | `go test ./internal/cli -run '^TestManagedFailedTurnLeavesClaimUntouched$' -count=1` → `ok` · `go test ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -v` → 하위 `lost_receipt_redelivery` 포함 7줄 PASS (측정: 기준 트리 PASS) |
| AC-MH-009 | REQ-MH-010, REQ-MH-011 | Given 가짜 세션이 `DeliverTurn` 안에서 자기 `Close` 가 불릴 때까지 막히고(또는 드라이버가 한가할 때) When 컨텍스트가 취소되면 Then 워처가 `Close` 를 불러 `DeliverTurn` 이 풀리고, 드라이버는 중단 오류를 반환하며 그 오류를 턴 실패로 세거나 `Factory turn failed` 로 기록하지 않는다 | `go test ./internal/cli -run '^TestManagedDriverStopsOnCancel$' -count=1 -v` → 하위 2줄(`idle`, `in_flight_turn`) PASS, 각 5초 이내 |
| AC-MH-010 | REQ-MH-010, REQ-MH-011 | Given 시험 바이너리를 `os/exec` 로 다시 기동해 소유자를 돌리고(스트림은 가짜 백엔드, Codex는 가짜 App Server) When 부모 시험이 SIGTERM·SIGHUP·SIGINT를 `os.Process.Signal` 로 보내면 Then 5초 안에 소유자가 종료하고, 소유한 자식 프로세스가 더 이상 존재하지 않으며, Codex는 토큰 임시 디렉터리가 삭제되고, launch-pending 행이 0건이며, 종료 오류가 중단을 이름에 담는다. 가짜 App Server가 준비되지 않는 상태에서 보낸 SIGTERM도 같다 | `go test ./internal/cli -run '^TestManagedLauncherSignalTeardown$' -count=1 -v` → 하위 `stream_SIGTERM`, `stream_SIGHUP`, `stream_SIGINT`, `codex_SIGTERM`, `codex_SIGTERM_during_handshake` 5줄 PASS, exit 0. Windows에서는 `runtime.GOOS == "windows"` 로 건너뛴다(이유: POSIX 시그널 전달 시험이며 Windows 빌드는 AC-MH-011이 증명) |
| AC-MH-011 | REQ-MH-013 | Given 시그널 도우미와 변경 코드가 반영된 트리 When Windows 대상으로 빌드·vet하고 `syscall.` 참조를 훑으면 Then 빌드와 vet가 통과하고, `managed_*.go` 의 `syscall.` 참조는 0행이며, `launch_signals.go` 의 `syscall.` 참조는 `SIGTERM`·`SIGHUP` 상수뿐이다 | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 · `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → exit 0 · `grep -rn 'syscall\.' internal/cli/managed_*.go` → 0행, exit 1 · `grep -n 'syscall\.' internal/cli/launch_signals.go` → `syscall.SIGTERM`, `syscall.SIGHUP` 만(양성 대조: 이 형태가 같은 두 줄을 실제로 낸다), exit 0 |
| AC-MH-012 | REQ-MH-014 | Given sync 단계가 끝난 트리 When 운영 문서와 CHANGELOG를 훑으면 Then 해결된 세 한계 문장이 문서에서 사라지고, 남은 한계(독 메시지의 TTL까지 재배달, Codex 타임아웃의 세션 치명, 라이브 미관측, Windows 콘솔 이벤트 미처리)가 명시되며, CHANGELOG에 이 SPEC 엔트리가 F3·F4·F5 해결과 남은 한계를 함께 적는다 | `grep -c "런처는 시그널을 처리하지 않는다" .moai/docs/factory-managed-session.md` → `0`(exit 1) · 같은 형태로 "턴 하나가 실패하면 세션 전체가 끝난다"·"서버가 먼저 보내는 승인·질문 요청에는 답하지 않는다" 둘도 `0` · `grep -c "SPEC-FACTORY-MANAGED-HARDEN-001" CHANGELOG.md` → 1 이상 · 남은 한계 문장 각각이 문서에서 1 이상 (정확한 문구는 manager-docs가 정하고 progress.md §E.4 에 인용) |
| AC-MH-013 | REQ-MH-015 | Given run 단계의 커밋 이력 When RED 기준선 커밋과 수리 커밋의 선후를 조회하면 Then 재현 시험 세 개(F3·F4·F5)와 그 RED 출력 원문(`.moai/reports/t1409/red-baseline.md`)이 한 RED 커밋에 있고, 그 커밋은 대응하는 모든 수리 커밋의 조상이다 | `git log --reverse --format=%h:%s develop..HEAD` → RED 커밋 줄이 수리 커밋 줄들보다 앞 · `git merge-base --is-ancestor <RED> <FIX>` → exit 0, 반대 방향은 exit 1(쌍으로 확인) · `git show --stat --format=%h <RED>` → 시험 파일과 `red-baseline.md` 가 같은 커밋에 있음 (SHA는 progress.md §E.2 에 인용) |
| AC-MH-014 | REQ-MH-016 | Given 카드 브랜치 When 부모 SPEC 디렉터리와 브로커 저장소의 변경을 조회하면 Then diff가 비어 있다 | `git diff --stat develop...HEAD -- .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 internal/factorymsg/store.go` → 출력 0행, exit 0 · 양성 대조: `git diff --stat develop...HEAD -- internal/cli` 는 M2 이후 1행 이상을 낸다 |
| AC-MH-015 | REQ-MH-001..016 (게이트) | Given 변경이 끝난 트리 When 형식·정적 분석·스코프 회귀를 돌리면 Then 오류 0건이고 회귀 없음 | `gofmt -l internal/cli internal/config` → 출력 0행 · `go vet ./internal/cli ./internal/config` → exit 0 · `golangci-lint run ./internal/cli/... ./internal/config/...` (CI 판 v2.1.6; 측정: 로컬 `golangci-lint --version` = v2.1.6) → exit 0 · `go test ./internal/cli -run '^.*(Managed\|managed).*$' -count=1` → `ok` (측정 기준선: `ok … 27.048s`, 정리 접두 있음; 위 패턴은 부분 문자열 `Managed`·`managed` 를 고르는 앵커형 동치이며 `-v` 로 센 최상위 테스트 66개로 리터럴 `Managed\|managed` 와 같음을 확인) · `go test ./internal/config -count=1` → `ok` (측정 기준선 `ok … 31.1s`) · `go test ./internal/guardstate ./internal/template -count=1` → 아래 §3 기준선 대비 신규 실패 0건 |

## §2. RED-now / green path 쌍 (verification-completeness.md §2)

두 칸이 쌍으로 있어야 채택이다(`.claude/rules/moai/development/verification-completeness.md` §2, §2.1). 기준 트리 `7109e0900`.

**릴리스 차단급(재현 시험이 있는 AC)** — 시험 파일이 아직 없어서 plan 단계에서 RED-now 네 요소(명령·stdout·exit·트리 SHA)를 시험 명령으로는 낼 수 없다. 이 칸은 **M1(RED 커밋)이 채우는 것이 채택 조건**이다(plan.md §F M1). 그전까지 이 AC들은 미채택이다. 아래는 M1이 달성해야 하는 "옳은 이유의 RED"와 plan 단계에서 임시 시험으로 실측한 근거(시험은 삭제되어 재실행 불가 — 보조 근거일 뿐 RED-now 셀이 아님)다.

| AC | RED-now가 붉어야 하는 이유(명시) | green path (어느 마일스톤이 뒤집는가) | plan 단계 보조 실측(기준 트리, 재실행 불가한 임시 시험) |
|---|---|---|---|
| AC-MH-001·002 | 소유자가 서버 요청에 답장을 보내지 않아 가짜 서버가 응답을 못 받는다(대기 타임아웃으로 실패) | M2 | 서버 요청 id=7을 보낸 뒤 1초간 클라이언트 프레임 없음: `no client frame within 1s: … i/o timeout` |
| AC-MH-003 | 서버 요청 id가 대기 중인 클라이언트 id와 같으면 호출이 서버 요청을 응답으로 반환(빈 결과·오류 없음), 문자열 id면 연결 닫힘 오류 | M2 | 충돌: `call result="" err=<nil>` (진짜 응답 `{"genuine":true}` 소실) · 문자열: `err=managed codex app server connection closed` |
| AC-MH-005 | 두 번째 턴의 에러 결과로 드라이버가 반환해 세 번째 턴이 전달되지 않는다 | M3 | 소스 판독(`managed_factory_session.go:345`), 기존 시험이 같은 반환 동작을 고정 |
| AC-MH-010 | 시그널을 받은 프로세스가 `defer` 정리 없이 끝나 자식이 살아 있고 임시 디렉터리·launch-pending 행이 남는다 | M4 | 별도 작은 프로그램: SIGTERM `launcher_rc=143 child=alive tokendir=present defer=not-run`, SIGHUP `launcher_rc=129 child=alive tokendir=present defer=not-run` (SIGINT는 하네스 한계로 미관측) |

**구조 확인형 AC(지금 명령으로 RED-now를 낼 수 있는 것)**

| AC | RED-now: 명령 | stdout | exit | 트리 | green path |
|---|---|---|---|---|---|
| AC-MH-007 (상수 부분) | `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` | (빈 출력) | 1 | `7109e0900` | M3 이 상수를 정의하면 정의 1행·exit 0 |
| AC-MH-010 (도우미 존재) | `grep -n NotifyContext internal/cli/launch_signals.go` | `grep: internal/cli/launch_signals.go: No such file or directory` | 2 | `7109e0900` | M4 가 도우미를 만들면 해당 줄·exit 0 (양성 대조: `grep -n NotifyContext internal/cli/mcp_server.go` → 123행, exit 0) |
| AC-MH-012 | `grep -c "런처는 시그널을 처리하지 않는다" .moai/docs/factory-managed-session.md` | `1` | 0 | `7109e0900` | sync 가 문장을 지우면 `0`·exit 1 |
| AC-MH-012 (CHANGELOG) | `grep -c "SPEC-FACTORY-MANAGED-HARDEN-001" CHANGELOG.md` | `0` | 1 | `7109e0900` | sync 가 엔트리를 쓰면 1 이상·exit 0 |

**회귀 가드(지금 이미 초록이 정상인 것 — 릴리스 차단 RED가 아니라 불변 가드)**: AC-MH-011(Windows 빌드: `GOOS=windows GOARCH=amd64 go build ./...` 기준 트리 exit 0, 출력 없음 측정; `grep -rn 'syscall\.' internal/cli/managed_*.go` 기준 트리 0행·exit 1), AC-MH-014(기준 트리 diff 0행), AC-MH-008 의 `lost_receipt_redelivery`(기준 트리 PASS 측정), AC-MH-015. 이들은 변경이 불변을 깨는지 잡는 용도이며 RED-now 를 요구하지 않는다.

**변이 탐침(채택 전 질문: 요구를 어기면서 기준을 만족시키는 변이가 쓰일 수 있는가)**:

- AC-MH-001 은 승인류 `accept` 로 바꾼 변이를 정책표 단언이 잡고, 모든 method를 오류로 답하는 변이를 하위 11줄 각각의 정확한 응답 단언이 잡는다 — 한 방향 기준이 아니라 표 전체 일치다.
- AC-MH-007 은 구현에 `3` 을 박은 변이를 "상수를 읽어 횟수를 정하는 시험"이 상수 변경 시 깨져 잡는다는 점에서만 방어한다. 상수 값을 바꿔 보는 것은 시험이 아니라 run 단계의 변이 확인(상수를 2로 바꿔 시험이 따라오는지)이 한다 — progress.md §E.2 에 그 한 번의 변이 실행을 인용한다.
- AC-MH-010 은 `Close` 대신 `os.Exit` 로 끝내는 변이를 "토큰 디렉터리 삭제 + launch-pending 행 0건" 단언이 잡는다(`os.Exit` 는 `defer` 를 건너뛴다).
- AC-MH-013 은 RED와 수리를 한 커밋에 합치는 변이를 `merge-base --is-ancestor` 가 같은 커밋에서 정방향 0·역방향 0 이 되어 쌍 확인이 실패하도록 잡는다.

## §3. 기준선과 알려진 사전 적색 (측정)

| 대상 | 명령 | 관측 (기준 트리 `7109e0900`) |
|---|---|---|
| 스코프 회귀 | (정리 접두) `go test ./internal/cli -run '^.*(Managed\|managed).*$' -count=1` | `ok  github.com/modu-ai/moai-adk/internal/cli  27.048s` (같은 선택을 리터럴 부분 문자열 패턴으로 돌린 첫 측정은 `ok … 30.959s`; 두 패턴 모두 `-v` 최상위 66개) |
| 같은 선택, 정리 접두 없이(레인 환경) | 리터럴 부분 문자열 패턴의 스코프 명령 | `--- FAIL: TestManagedSwitchDoesNotReachCodexLaneLoop (0.48s)` — `todo add: moai add: refused — lane boundary: a lane session cannot mutate the queue …`, `FAIL … 26.413s`. 레인 환경이 원인(정리 후 PASS 13.40s). 이 카드가 만든 적색이 아니다. |
| config | `go test ./internal/config -count=1` | `ok  …/internal/config  31.134s` |
| template | `go test ./internal/template -count=1` | `ok  …/internal/template  414.986s` (분 단위 스위트 → 임대 안에서만 실행) |
| guardstate | `go test ./internal/guardstate -count=1` | **기준 트리에서 이미 적색**: `--- FAIL: TestCensus_SetDifferenceEmptyBothDirections (0.00s)` — `census_test.go:80: disk\manifest: .github/workflows/workflow-parse-guard.yaml exists on disk with no manifest entry`, `census_test.go:89: declared 19 entries against 20 workflow files`. 이 카드는 워크플로 파일을 건드리지 않으므로 이 적색은 사전 존재이며, AC-MH-015 는 "이 한 테스트 외 신규 실패 0건"으로 읽는다. |
| 템플릿 미러 | `ls internal/template/templates/.moai/docs/factory-managed-session.md` | `No such file or directory`(미러 없음) |

## §4. 에지 케이스

- 서버 요청이 턴 사이(한가한 때)에 온다 — 읽기 고루틴이 바로 답한다(D-1 이유 1). 시험: AC-MH-002 변형 하위 케이스로 포함.
- 서버 요청 두 개가 연달아 온다 — 각각 정확히 한 번 답하고 순서가 보존된다.
- 답장 쓰기가 실패한다 — 읽기 고루틴이 끝나고 진행 중 `DeliverTurn` 은 "connection closed" 로 세션 치명 오류가 된다.
- 알 수 없는 알림(id 없음)은 지금처럼 버린다. 답하지 않는다.
- 운영자 입력으로 시작된 턴이 실패해도 같은 횟수에 센다. 성공한 운영자 턴은 횟수를 0으로 되돌린다.
- 우선 턴이 `is_error` 로 끝나도 세션 치명이다(턴 단위 분류는 우선 턴 이후에만 적용).
- 시그널이 `Start()` 의 핸드셰이크 중에 온다 — 준비 대기가 시그널 컨텍스트에서 파생되므로 핸드셰이크 예산(10초)을 기다리지 않고 풀린다(AC-MH-010 `codex_SIGTERM_during_handshake`).
- 정리 중 두 번째 시그널이 온다 — 첫 시그널에서 기본 동작을 복원하므로 프로세스가 끝난다(정리가 멈춘 경우의 탈출구). 이 경로는 자동 시험으로 고정하지 않는다(Gap, progress.md에 기록).
- `Close` 가 이미 끝난 뒤 시그널이 온다 — 두 번째 `Close` 는 같은 결과를 돌려주고 아무것도 다시 하지 않는다.
- `/exit` 와 stdin EOF 경로는 기존 동작 그대로다(AC-MH-015 스코프 회귀가 고정).

## §5. 품질 게이트

- TRUST 5 — Tested: 변경한 비테스트 코드의 줄이 재현·분류·시그널 시험으로 덮인다(`go test -cover` 를 managed 슬라이스에 한정해 측정, 부모 기준선 89.3%보다 내려가지 않는다 — 부모 sync 기록의 수치이며 이 plan에서 재측정하지 않았다). Readable/Unified: `gofmt`·`go vet`·`golangci-lint` 0건. Secured: 승인류 응답에 accept 계열 0건(AC-MH-001), 토큰 정리 유지(AC-MH-010). Trackable: Conventional Commit, 모든 커밋 메시지에 카드 id `t1409`.
- 경계 grep(부모 AC 계승): `grep -rnE 'merge-window|Decider|T29b|T29c|handover' internal/cli/managed_*.go` 0행, `grep -rn 'AskUserQuestion' internal/cli/managed_*.go` 비테스트 0행.

## §6. Definition of Done

1. AC-MH-001..015 전부 PASS — 명령과 원문 출력을 progress.md §E.2 에 인용(AC-MH-012 는 sync 후 §E.4).
2. 재현 시험 세 개의 RED 원문이 `.moai/reports/t1409/red-baseline.md` 에 있고 RED 커밋이 수리 커밋의 조상이다(AC-MH-013).
3. 상수를 한 번 바꿔 시험이 따라오는지 보인 변이 실행 1건이 증거로 있다(AC-MH-007).
4. Windows 크로스 빌드·vet가 exit 0이고 `managed_*.go` 의 `syscall.` 참조가 0행이다(AC-MH-011).
5. 부모 SPEC 디렉터리와 `store.go` 의 diff가 비어 있다(AC-MH-014).
6. 운영 문서와 CHANGELOG가 해결된 세 한계와 남은 한계를 과장 없이 적는다(AC-MH-012).
