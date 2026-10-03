---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "acceptance.md — 인수 기준"
version: "0.5.2"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# acceptance.md — AC-MH-001..013

> 이 파일은 검증층이고 시나리오는 Given-When-Then이다. 요구(GEARS)는 `spec.md` §C 의 REQ-MH-001..012 가 운반한다. 모든 AC는 이진 판정이 가능하다. 0.4.0 은 범위 분할 개정이다: F5(시그널·`Start`/`Close` 수명주기)의 AC 와 변이는 카드 t1459 로 나갔고 이 문서에 남지 않는다(이전 텍스트는 git 이력 `95dfd85c8`·`820eff47f`·`951f2bfb6`). 0.5.0 은 축소 범위 plan-audit 1차(FAIL 0.75, `bca1e0629`)의 개정이다 — 변이표를 **한 단계 편집으로 적용 가능하고 실제로 붉어지는 것만** 남기고 나머지는 §2.5 "불변 가드, 변이 미채택" 목록으로 옮겼다.
>
> **명령 규약.** 검증 명령은 단일 호출형이고 `-run` 패턴은 앵커형(`^…$`)이다. 부하 규율상 `go test ./internal/cli` 전체는 로컬에서 돌리지 않는다. **파이프 문자를 담은 명령은 표 밖 §1.1 의 fenced 블록에 원문 그대로(바이트 단위로 파이프) 적고 표에서는 AC id로 가리킨다** — 표 안의 이스케이프(백슬래시 파이프)는 Go 정규식에서 리터럴 파이프라 테스트를 0개 고르고 `ok` 를 내기 때문이다(측정은 §3). 표의 명령에는 파이프가 없다.
>
> **레인 세션에서 `go test` 는 환경 정리 복합형으로 돌린다**: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test …` (실측: 기준 트리에서 정리 없이 돌린 스코프 명령은 `TestManagedSwitchDoesNotReachCodexLaneLoop` 가 레인 환경 때문에 `lane boundary` 거부로 가짜 적색이었고, 정리 후 PASS). 아래 명령은 그 접두 없이 적는다.
>
> **빈 스윕 규율.** 모든 `-run` 명령은 읽기 전에 선택된 테스트 수를 먼저 확인한다(`go test … -list '<같은 패턴>'`). `[no tests to run]`·`[no test files]` 는 통과가 아니라 **실패(미측정)** 이고, `-list` 가 이름을 하나도 안 내는 `ok` 도 같다(측정: 기준 트리에서 AC-MH-004·006 의 패턴을 `-list` 로 고르면 출력이 `ok  github.com/modu-ai/moai-adk/internal/cli  1.352s` 한 줄뿐이고 `Test` 이름이 0개다 — 시험이 아직 없기 때문이며, 이 상태에서 `-run` 이 `ok` 를 내는 것이 곧 공허 통과다).
>
> **로그 줄 관측 규약.** 로그 줄(`Factory server request answered: …`, `Factory turn failed (…)`)은 design.md D-1 의 로그 출력 이음새(패키지 비공개 원자 포인터, 기본 `os.Stderr`)로 나가고, 시험은 그 쓰기 대상을 바꿔 잡는다. `os.Stderr` 전역을 교체하지 않는다. 이 이음새가 이 SPEC의 유일한 시험 이음새다. **원자 포인터만으로는 `-race` 안전하지 않다**(design.md D-1) — 그래서 (1) 시험 sink 는 쓰기와 스냅숏 조회가 한 뮤텍스 아래 있는 타입이고 시험은 원시 버퍼의 `String()` 을 직접 부르지 않는다, (2) 클라이언트를 시작한 시험은 정리에서 `shutdown()` 을 부르고 읽기 고루틴이 끝나기(`events` 채널이 닫히기)까지 기다린 **뒤에** 포인터를 복원한다(등록 순서는 포인터 복원 먼저, 클라이언트 종료 나중), (3) 로그 줄을 기다리는 하위 케이스를 도는 AC-MH-001·005·006 의 명령은 `-race` 를 포함한다. 로그 줄은 답장을 쓰기 **전에** 쓰이지만(design.md D-1) 모든 로그 단언은 그 순서에 기대지 않고 **상한 5초·간격 10ms 의 폴링**으로 줄을 기다린다. "정확히 한 번"(REQ-MH-001)은 가짜 서버가 첫 답장을 받은 뒤 **500ms 의 조용한 구간** 동안 같은 id 의 둘째 답장이 없음을 확인해 보인다 — 500ms 는 저장소 자신의 한 폴링 틱(`DefaultManagedSessionPollInterval` = 500ms, `defaults.go`)과 같은 값이라, 같은 깨어남에서 나가는 중복(마이크로초 안)과 다음 틱에서 나가는 중복을 모두 덮는다(가정: loopback 왕복은 이보다 훨씬 짧다 — 측정하지 않았다).
>
> **블로킹 호출 규약 (시간 상한).** 블로킹 호출 — `DeliverTurn`, `call()`, `startTurn`, `driveManagedFactorySession` — 은 **고루틴에서** 돌리고 시험 고루틴이 **절대 상한 5초**에 `select` 로 그 호출을 실패시킨다(상한 값은 시험 코드의 상수 하나). 정리에서 가짜 서버와 연결을 닫아 막힌 호출을 풀어 주어 고루틴이 새지 않게 한다. 드라이버 시험은 stdin 에 `/exit` 종결자를 둬 반환해야 할 곳에서 변이로 계속하더라도 시험이 멈추지 않고 "반환값이 기대한 오류가 아님"으로 붉어지게 한다. AC-MH-002 의 RED 는 **이 5초 watchdog 상한**이 깨지는 것이다 — 10분 `DefaultManagedCodexTurnTimeout` 도, `go test` 패키지 기본 타임아웃(10분)도 아니다(둘은 wrong-reason RED).
>
> **측정 기준 트리**: `7109e0900`(카드 기준 develop; 계획 커밋 이후도 소스는 동일). "측정" 표기는 이 plan 실행이 관측한 것이다.

## §1. AC 매트릭스

| AC | 요구 | 시나리오 (Given-When-Then) | 검증 명령 → 기대 관측 |
|---|---|---|---|
| AC-MH-001 | REQ-MH-001, REQ-MH-002, REQ-MH-003 | Given 가짜 App Server가 턴 중에 서버 요청 10종(각각 정수 id)과 목록에 없는 method 하나를 차례로 보낼 When Codex 소유자가 처리하면 Then 각 요청에 정확히 한 번, 같은 id로, design.md D-1 정책표대로 답하고(명령·파일 승인은 `decline`, 레거시 `applyPatchApproval`·`execCommandApproval` 은 스키마 객체 `{"decision":{"denied":{"rejection":"<고정 문구>"}}}`(문자열 `"denied"` 가 아니다), 권한 승인은 빈 `permissions`, elicitation 은 `decline` 액션, 동적 도구 호출 `item/tool/call` 은 `success:false` 결과, 오류류 3종은 `-32000`, 모르는 method는 `-32601`), 정확히 한 번은 첫 답장 뒤 500ms 조용한 구간에 중복이 없음으로 보이고, 어느 답에도 accept 계열 결정이나 비어 있지 않은 권한 부여가 없으며, 답한 요청마다 주입한 로그 대상에 method를 담은 한 줄이 남고 elicitation 줄은 `serverName`·`turn=<id 또는 none>`·`broker_declined=<k>` 도 담는다. 또한 **스키마 적합 가드**: 정책표의 모든 답(10종과 미지 method 폴백)의 실제 와이어 본문이 `internal/cli/testdata/codex-0.160.0/` 에 vendoring 된 codex 0.160.0 응답 스키마(8개 파일 + 생성법을 적은 README; 오류류 3종과 미지 method 폴백은 `JSONRPCError.json`)에 대해 `santhosh-tekuri/jsonschema/v6`(기존 직접 의존)로 검증을 통과한다 — `codex` 바이너리도 네트워크도 필요 없다. 이 가드는 **모양만** 본다: 스키마상 유효하지만 틀린 값(`abort`, accept 계열)은 위의 정확한 응답 단언이 잡는다. 이 가드는 새 AC 가 아니라 AC-MH-001 의 증거이며 REQ-MH-002 가 정한 답(design.md D-1 정책표)의 모양을 고정한다 | `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v` → 상위 `--- PASS: TestManagedCodexServerRequestPolicy (…s)` 1줄과 그 아래 하위 테스트 11개(10종 + 미지 method 1)가 각각 PASS(각 하위 케이스가 응답 본문과 로그 줄을 단언), exit 0 · `go test ./internal/cli -run '^TestManagedServerRequestPolicyMatchesCodexSchema$' -count=1 -v` → 상위 PASS 1줄과 하위 12개(정책 11종 + 검증기 자체 점검 `validator_catches_the_bare_denied_string` 1개; 이 개정 트리에서 측정, `ok … 2.408s`)가 각각 PASS, exit 0. 두 시험을 한 명령으로 돌려도 같다(같은 측정에서 하위 11개와 12개 PASS) |
| AC-MH-002 | REQ-MH-001 | Given 가짜 App Server가 `turn/start` 응답 뒤 명령 승인 요청을 보내고 그 답을 받은 뒤에만 `turn/completed` 를 보낼 When `DeliverTurn` 을 부르면 Then 답이 서버에 도달하고 `DeliverTurn` 이 nil 을 반환하며, `DeliverTurn` 을 고루틴에서 돌리는 시험의 **watchdog 절대 상한 5초 안에** 돌아온다(헤더의 블로킹 호출 규약; 턴 타임아웃 상수는 10분이므로 답을 못 받으면 watchdog 이 먼저 시험을 실패시키고, 정리가 가짜 서버·연결을 닫아 막힌 호출을 푼다). 같은 시험의 하위 케이스: 요청이 턴 사이(한가한 때)에 와도 답한다 | `go test ./internal/cli -run '^TestManagedCodexTurnSurvivesServerRequest$' -count=1 -v` → 하위 `during_turn`, `between_turns` 2개 PASS(각각 경과 5초 미만 단언), exit 0 |
| AC-MH-003 | REQ-MH-004 | Given 서버 요청의 id가 대기 중인 클라이언트 id와 같거나 문자열(`"srv-1"`)일 When 소유자가 `turn/start` 응답을 기다리면 Then 서버 요청은 응답으로 오인되지 않고(진짜 응답 본문이 호출자에게 도달), 문자열 id의 답장에 같은 문자열이 되돌아가며, 읽기 루프가 계속 살아 이후 호출이 성공한다 | `go test ./internal/cli -run '^TestManagedCodexServerRequestIDCollision$' -count=1 -v` → 하위 `integer_collision`, `string_id` 2개 PASS, exit 0 |
| AC-MH-004 | REQ-MH-005 | Given 읽기 고루틴이 서버 요청에 답장을 쓰는 동안 다른 고루틴이 `call()` 로 쓰는 상황을 200회 반복(반복마다 서버 요청 5건을 묶음으로 보내며 클라이언트 `call()` 1건) When 경합 검출기로 돌리면 Then 데이터 경합 보고가 0건이고 gorilla 의 동시 쓰기 패닉이 없으며 모든 답장과 응답이 도달한다. 반복 수 근거: 경합은 스케줄 의존이라 단일 실행으로는 놓칠 수 있어 반복으로 확률을 올린다(변이 mu5 를 적용해 붉어지는지 run 이 확인; 5회 재시도해도 초록이면 반복 수를 올린다). 모든 `call()`·답장 대기는 헤더의 블로킹 호출 규약(5초 watchdog)을 따른다 | `go test -race ./internal/cli -run '^TestManagedCodexConcurrentWrites$' -count=1 -v` → 선행 `-list` 로 **정확히 1개** 최상위 `Test` 이름(`TestManagedCodexConcurrentWrites`) 확인 뒤 PASS, `WARNING: DATA RACE` 0건, `[no tests to run]` 없음, exit 0 (`-race` 는 cgo 필요: 측정 `CGO_ENABLED=1`) |
| AC-MH-005 | REQ-MH-006 | Given 우선 턴 성공 뒤 두 번째 턴이 에러 결과를 내고 세 번째 턴이 성공하는 스크립트 세션 When 드라이버가 돌면 Then 드라이버는 두 번째 턴 뒤에 반환하지 않고 세 번째 턴을 전달하며, 주입한 로그 대상에 `Factory turn failed (1/` 로 시작하는 한 줄이 남고 stdout에는 쓰이지 않는다 | `go test -race ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$' -count=1 -v` → PASS, 시험이 세션이 받은 턴 3건과 로그 한 줄(상한 있는 폴링)을 단언하고 `/exit` 로 드라이버를 끝낸다 |
| AC-MH-006 | REQ-MH-006, REQ-MH-007 | Given §1.3 의 하위 케이스 17개(실패 원인별 10행과 elicitation 귀속 7행)와, 소유 App Server 승인 인수의 서버 이름 접두가 브로커 서버 이름 상수와 같다는 고정 시험 When 소유자와 드라이버를 통과시키면 Then §1.3 표의 각 기대가 성립한다 | §1.1 의 AC-MH-006 블록 — 선행 `-list` 가 **정확히 2개** 최상위 `Test` 이름을 내야 하고, 이어지는 `-run` 이 최상위 2개와 `TestManagedTurnFailureClassification` 의 하위 17개 PASS(명령에 `-race` 포함), exit 0 |
| AC-MH-007 | REQ-MH-008 | Given 우선 턴 성공 뒤 표식이 붙은 턴 단위 실패를 차례로 돌려주는 스크립트 세션(`fakeManagedSession`; stdin 은 턴마다 운영자 줄 하나와 끝의 `/exit`; 시험은 헤더의 5초 watchdog)과 상한 N = `config.DefaultManagedSessionMaxConsecutiveTurnFailures` When 드라이버가 돌면 Then 세 하위 케이스가 성립한다. `at_ceiling_returns`: 연속 N번 실패하면 N번째 실패 뒤 드라이버가 **그 마지막 오류를 횟수 N 과 함께 반환**하고 이후 턴(`/exit` 포함)을 처리하지 않는다. `below_ceiling_continues`: 연속 N−1번 실패 뒤 드라이버가 **반환하지 않고** 다음(성공) 턴을 전달한 다음 `/exit` 로 nil 을 반환한다(세션이 받은 턴 = 우선 턴 1 + 실패 N−1 + 성공 1). `success_resets`: N−1번 실패, 성공 1번, 실패 1번 순서에서 성공이 횟수를 0으로 되돌려 마지막 실패가 `1/N` 번째로 세어지고 드라이버가 반환하지 않는다(`/exit` 의 nil). 시험은 상수 자체를 읽어 N 을 정한다 | `go test ./internal/cli -run '^TestManagedDriverConsecutiveFailureCeiling$' -count=1 -v` → 하위 3개(`at_ceiling_returns`, `success_resets`, `below_ceiling_continues`) PASS · `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` → 정의 1행, exit 0 · `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/cli/managed_factory_session.go` → 드라이버가 상수를 참조하는 줄 1행 이상, exit 0 |
| AC-MH-008 | REQ-MH-009 | Given 브로커에 메시지 1건이 claim되고 그 메시지를 실은 턴이 실패할 When 드라이버가 세션을 이어 가면 Then 관리 계층은 그 행을 `claimed` 로 두고(`acknowledged` 0, claim token 불변) 해제·재주소를 하지 않는다. 재배달은 lease 만료 뒤 브로커의 `Claim` 이 새 token으로 한다. **이 AC 는 불변 가드이고 변이를 채택하지 않았다**(§2.5 G1: 드라이버에 store 핸들이 없어 한 단계 변이로 붉게 만들 수 없고 기준 트리에서도 초록이다). 재배달 증거는 기존 브로커 시험이 운반한다 | `go test ./internal/cli -run '^TestManagedFailedTurnLeavesClaimUntouched$' -count=1 -v` → PASS · `go test ./internal/factorymsg -run '^TestDispatchResultExactlyOnce$' -count=1 -v` → 하위 `lost_receipt_redelivery` 포함 7개 PASS (**이 plan 실행에서 다시 돌렸다**: 기준 트리에서 7개 PASS, `ok  …/internal/factorymsg  2.167s`) |
| AC-MH-009 | REQ-MH-012 | Given 변경이 반영된 트리 When Windows 대상으로 빌드·vet하고 `syscall.` 참조를 훑으면 Then 빌드와 vet가 통과하고 `managed_*.go` 의 `syscall.` 참조는 0행이다 | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 · `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → exit 0 · `grep -rn 'syscall\.' internal/cli/managed_*.go` → 0행, exit 1 (양성 대조: `grep -n 'syscall\.' internal/cli/mcp_server.go` 는 1행 이상, exit 0 — 같은 형태가 실제로 줄을 낸다) |
| AC-MH-010 | REQ-MH-010 | Given sync 단계가 끝난 트리 When 운영 문서와 CHANGELOG를 훑으면 Then 해결된 F3·F4 한계 문장과 "후속 카드 t1409 대상" 단락이 문서에서 사라지고, 시그널 처리 공백은 해결이 아니라 **카드 t1459 를 가리키는 안내**로 남으며, 남은 한계가 고정 앵커 문자열로 문서에 있고, CHANGELOG에 이 SPEC 엔트리가 F3·F4 해결과 남은 한계와 t1459 포인터를 함께 적는다 | 아래 §1.2 의 grep 목록(각 줄이 기대하는 개수·exit 코드 포함) |
| AC-MH-011 | REQ-MH-011 | Given run 단계의 커밋 이력 When RED 기준선 커밋과 수리 커밋의 선후를 조회하면 Then 재현 시험 6개(F3 4개·F4 2개)와 그 RED 출력 원문인 **추적되는** `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 가 한 RED 커밋에 있고, 그 커밋은 대응하는 모든 수리 커밋의 조상이다. `.moai/reports/t1409/` 는 로컬 사본일 뿐 근거로 인용하지 않는다 | `git log --reverse --format=%h:%s develop..HEAD` → RED 커밋 줄이 수리 커밋 줄들보다 앞 · `git merge-base --is-ancestor <RED> <FIX>` → exit 0, 반대 방향은 exit 1(쌍으로 확인) · `git show --stat --format=%h <RED>` → 시험 파일들과 `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 가 같은 커밋에 있음 · `git ls-files .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` → 그 경로 1행 · `git check-ignore -v .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` → 출력 0행, exit 1 (양성 대조: 같은 명령을 `.moai/reports/t1409/red-baseline.md` 에 주면 `.gitignore:235:.moai/reports/*` 를 내고 exit 0 — 측정) (SHA는 progress.md §E.2 에 인용) |
| AC-MH-012 | REQ-MH-012 | Given 카드 브랜치 When 부모 SPEC 디렉터리와 브로커 저장소의 변경을 조회하면 Then diff가 비어 있다 | `git diff --stat develop...HEAD -- .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 internal/factorymsg/store.go` → 출력 0행, exit 0 · 양성 대조: `git diff --stat develop...HEAD -- internal/cli` 는 M2 이후 1행 이상을 낸다 |
| AC-MH-013 | REQ-MH-001..012 (게이트) | Given 변경이 끝난 트리 When 형식·정적 분석·스코프 회귀를 돌리면 Then 오류 0건이고 회귀 없음 | §1.1 의 AC-MH-013 블록 — 선택 수 확인(기준 66개 최상위, 줄어들면 실패; 새 최상위 시험 이름이 모두 `Managed` 를 포함하므로 M1 뒤에는 66+6 이상, M2 뒤 66+7 이상, M3 뒤에는 66+11 이상: M1 6개(`ServerRequestPolicy`, `TurnSurvivesServerRequest`, `ServerRequestIDCollision`, `DeclinedBrokerElicitationFailsTurn`, `DriverIsolatesTurnFailure`, `CodexNonCompletedTurnIsolated`), M2 +1(`ConcurrentWrites`), M3 +4(`TurnFailureClassification`, `BrokerNameMatchesApprovalArgs`, `DriverConsecutiveFailureCeiling`, `FailedTurnLeavesClaimUntouched`); 개정 0.5.2 의 스키마 가드 `TestManagedServerRequestPolicyMatchesCodexSchema` +1 이라 계획된 목록은 66+12 이상. 위 하한은 마일스톤별 하한이고, 개정 0.5.2 트리의 실측은 **79개**(66+13)다 — 계획 목록 밖에 run 단계가 더한 `TestManagedCodexCompletionEventCarriesBrokerVerdict` 1개가 더 있다) 뒤 `ok`, `[no tests to run]` 는 실패 · `gofmt -l internal/cli internal/config` → 출력 0행 · `go vet ./internal/cli ./internal/config` → exit 0 · `golangci-lint run ./internal/cli/... ./internal/config/...` (CI 판 v2.1.6; 측정: 로컬 `golangci-lint --version` = v2.1.6) → exit 0 · `go test ./internal/config -count=1` → `ok` (측정 기준선 `ok … 31.1s`) · `go test ./internal/guardstate -count=1` → §3 기준선 대비 신규 실패 0건(사전 적색 1건 제외) · `go test ./internal/template -count=1` → `ok`, **임대 안에서만**(`moai slot acquire --resource internal-template-suite --max-duration 15m` → 실행 → `moai slot release --resource internal-template-suite`; 측정 414.986s). template 스위트는 "영향 범위만" 규칙에 따라 이 변경이 직접 닿지 않는다는 근거(plan.md §C)가 있어 **M4 에서 한 번만 도는 스모크**이고 마일스톤별 회귀 명령이 아니다 |

### §1.1 파이프를 담은 명령 원문 (표 밖, 바이트 단위 원문)

**AC-MH-006**

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^(TestManagedTurnFailureClassification|TestManagedBrokerNameMatchesApprovalArgs)$'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test -race ./internal/cli -run '^(TestManagedTurnFailureClassification|TestManagedBrokerNameMatchesApprovalArgs)$' -count=1 -v
```

**AC-MH-013** (스코프 회귀 — `Managed`·`managed` 를 이름에 가진 최상위 테스트)

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^.*(Managed|managed).*$' | grep -c '^Test'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -run '^.*(Managed|managed).*$' -count=1
```

기대: 첫 줄이 66 이상(측정: 이 개정에서 재측정한 기준 트리 66), 둘째 줄이 `ok` 이고 `[no tests to run]` 가 없다.

### §1.2 AC-MH-010 의 grep 목록 (원문, sync 후 트리에서)

| 명령 | 기대 |
|---|---|
| `grep -c "턴 하나가 실패하면 세션 전체가 끝난다" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1` — 측정) |
| `grep -c "서버가 먼저 보내는 승인·질문 요청에는 답하지 않는다" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1` — 측정) |
| `grep -c "후속 카드 t1409 대상" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1`, 71행 — 측정) |
| `grep -c "런처는 시그널을 처리하지 않는다" .moai/docs/factory-managed-session.md` | 1 이상 — **F5 는 열려 있으므로 이 문장은 남아야 한다**(기준 트리 `1`·exit 0 — 측정; 같은 줄 또는 바로 아래 문장이 t1459 를 가리킨다) |
| `grep -c "t1459" .moai/docs/factory-managed-session.md` | 1 이상(기준 트리 `0` — 새 포인터) |
| `grep -c "관리 계층은 .syscall. 을 쓰지 않는다" .moai/docs/factory-managed-session.md` (점 두 개가 원문의 백틱을 대신한다) | 1 (기준 트리 `1`, 62행 — 측정; F5 가 빠져 이 문장은 계속 참이므로 **바뀌면 안 된다**) |
| 남은 한계 앵커 4개 각각 `grep -c "<앵커>" .moai/docs/factory-managed-session.md` | 각 1 이상: `TTL까지 재배달`, `턴 타임아웃은 세션을 끝낸다`, `거부된 elicitation`, `쓰기 데드라인` |
| 라이브 미관측 문장: `grep -c "실제 codex 세션에서는 관측하지 않았다" .moai/docs/factory-managed-session.md` | 1 이상 |
| `grep -c "SPEC-FACTORY-MANAGED-HARDEN-001" CHANGELOG.md` | 1 이상(기준 트리 `0`, exit 1 — 측정) |
| `grep -c "t1459" CHANGELOG.md` | 1 이상(CHANGELOG 엔트리가 F5 를 해결이 아니라 t1459 몫으로 적음) |
| `grep -c "후속 카드 t1409 대상" CHANGELOG.md` | `0`, exit 1 (기준 트리 `1` — 부모 CHANGELOG 엔트리의 문구, 이 plan 실행에서 측정; sync 가 부모 엔트리의 해당 문구를 F3·F4 해결과 F5 의 t1459 포인터로 정정한다) |

(앵커 문자열은 문서 문장이 반드시 포함해야 하는 고정 부분 문자열이다. "쓰기 데드라인" 앵커의 문장은 막힌 쓰기가 연결이 죽거나 세션이 닫힐 때까지 풀리지 않는다는 한계를 말해야 하며 턴 타임아웃이 풀어 준다고 쓰면 안 된다(design.md 공시한 한계). 문장의 나머지는 manager-docs가 쓰되 과장 없이 쓴다. 문서의 "알려진 한계" 에서 F3·F4 문장만 지우거나 고치고 F5 문장은 t1459 포인터와 함께 남긴다.)

### §1.3 AC-MH-006 하위 케이스 (`TestManagedTurnFailureClassification` 17개)

시험은 두 층으로 쓴다. 실패 원인 행(#1–#10)은 소유자·드라이버를 통과시키고, elicitation 귀속 행(#11–#17)은 `armTurn()` 을 직접 부르고 `read()` 고루틴을 띄운 **클라이언트 수준**에서 가짜 서버가 프레임 묶음을 쓰며, 소비자(`waitTurn`)를 시험이 원하는 시점까지 **호출하지 않는 것으로** 읽기 고루틴이 소비자보다 앞서게 한다. 읽기 고루틴이 요청까지 처리했음은 주입한 로그 대상에 그 요청의 줄이 보이는 것으로 안다(별도 이음새 없음).

| # | 하위 케이스 | 입력 | 기대 |
|---|---|---|---|
| 1 | `stream_is_error_after_priming` | 우선 턴 성공 뒤 `result.is_error` | 턴 단위: 드라이버 계속, 연속 횟수 +1 |
| 2 | `codex_failed_or_interrupted` | `turn/completed` 상태 `failed` 와 `interrupted`(두 값을 **각각** 단언 — 한쪽만 표식하는 변이도 잡는다) | 턴 단위. 채택: M1 `TestManagedCodexNonCompletedTurnIsolated`(RED 구성)와 mu15 |
| 3 | `codex_moai_elicitation_turn_completed` | 턴 중 `serverName=moai` elicitation(`turnId` 현재 턴), 이어서 상태 `completed` | 턴 단위(`DeliverTurn` 이 표식 오류), 로그 줄 `turn=<id>` `broker_declined=1` |
| 4 | `priming_is_error` | 우선 턴이 `is_error` | 세션 치명: 드라이버가 그 오류를 반환. **불변 가드 G2(변이 미채택)** — 우선 턴은 루프 앞에서 무조건 반환한다 |
| 5 | `stream_closed` | `errManagedStreamClosed` | 세션 치명(드라이버 통과 행: 아래 문단) |
| 6 | `codex_connection_closed` | "connection closed" | 세션 치명(드라이버 통과 행) |
| 7 | `write_failure` | 자식 stdin 쓰기 실패 | 세션 치명(드라이버 통과 행) |
| 8 | `codex_timeout` | `startTurn` 에 50ms 마감 컨텍스트를 주고 `turn/completed` 를 보내지 않음(이음새 없음, design.md D-2; **드라이버를 거치지 않는** `startTurn` 수준 행) | `context.DeadlineExceeded` 가 표식 오류가 **아님**(`errors.Is` 거짓). 채택: mu16 |
| 9 | `unclassified_error` | 표식 없는 일반 오류 | 세션 치명(드라이버 통과 행) |
| 10 | `other_server_elicitation` | 턴 중 `serverName` 이 브로커가 아닌 elicitation, 이어서 `completed` | 실패로 세지 않음: `waitTurn` nil, 로그 줄 `broker_declined=0` |
| 11 | `between_turns_reader_ahead` (a) | `armTurn()` 뒤 서버가 **한 묶음**으로 `turn/started(T1)`, `turn/completed(T1)`, `moai` elicitation(`turnId` null)을 씀. 시험은 로그 대상에 `turn=none` 줄이 보일 때까지(상한 5초) `waitTurn` 을 호출하지 않음 | 그 `turn=none` 줄이 `broker_declined=0` 을 담고(후행 요청은 어느 턴의 거부 수에도 들지 않음), 그 뒤 `waitTurn(ctx, "T1")` 이 **nil**(완료 프레임 뒤의 요청은 어느 턴도 실패시키지 않음). 채택: 줄이 `turn=none` 이 아니면(mu19: 창이 열린 채라 후행 요청이 T1 에 귀속돼 `turn=T1`) 5초 폴링이 붉고, 줄이 `turn=none` 이어도 `broker_declined` 가 0 이 아니면(mu20) 값 단언이 붉다 — 두 변이 모두에서 `waitTurn` 의 nil 단언은 **초록**이다(판정이 후행 요청보다 먼저 완료 이벤트에 실렸다). 소비자 쪽 계수기 설계로 되돌리면 nil 단언이 붉어진다(G7, DoD 아님) |
| 12 | `normal_turn_after_declined_turn` (b) | `moai` elicitation 으로 실패한 턴 T1 다음에 `armTurn()` 하고 정상 턴 T2 | T2 가 nil(초기화는 `armTurn`) |
| 13 | `elicitation_after_turn_start_response` (c) | `armTurn()` 뒤 `turn/started` 이전에 `turnId` null 인 `moai` elicitation, 이어서 `turn/started(T2)`·`completed` | T2 를 실패시킴(턴 단위) |
| 14 | `two_requests_in_one_turn` (d) | 한 턴 안에 `moai` elicitation 2건 | **정확히 한 번**의 턴 단위 실패(`waitTurn` 오류 1개). 둘째 요청의 로그 줄이 `broker_declined=2` 를 보임(관측 가능한 값으로 판정이 불리언 `> 0` 인지 `== 1` 인지 가른다: mu8) |
| 15 | `late_previous_turn_id_before_turn_started` | T1 완료 뒤 `armTurn()`; `turn/started(T2)` **이전에** `turnId=T1` 인 `moai` elicitation; 이어서 `turn/started(T2)`·`completed(T2)` | T2 가 **nil**(직전 완료 턴 id 를 단 요청은 어느 턴도 실패시키지 않음), 로그 줄 `broker_declined=0` |
| 16 | `turn_id_mismatch_in_open_window` | `turn/started(T2)` 로 현재 턴이 확정된 뒤 `turnId=T9`(T2 도 직전 턴도 아님)인 `moai` elicitation, `completed(T2)` | T2 가 nil(확정된 현재 턴과 다른 id 는 세지 않음) |
| 17 | `string_jsonrpc_id_request_counts` | 열린 창에서 JSON-RPC `id` 가 **문자열**(`"srv-9"`)이고 `turnId` 가 현재 턴 T2 인 `moai` elicitation | T2 를 실패시킴(JSON-RPC id 형태와 귀속은 무관), 답장 id 가 `"srv-9"` 로 되돌아감 |

**드라이버 통과 행(#5, #6, #7, #9)**: 이 네 행은 `fakeManagedSession` 이 우선 턴 뒤 두 번째 턴에서 그 오류를 돌려주게 해 **`driveManagedFactorySession` 을 실제로 통과**시킨다. 시험은 stdin 에 `/exit` 를 둔다. 단언: 드라이버가 watchdog 안에 **그 오류를 반환**하고(변이로 계속했다면 `/exit` 로 nil 이 반환돼 "반환값이 기대한 오류가 아님"으로 붉어진다), 로그 대상에 `Factory turn failed` 줄이 **없고**, 그 뒤 턴이 전달되지 않았다. 이 행들이 mu14 의 지목 대상이다. 나머지 행의 층: #1 은 M1 의 실제 `pumpManagedStreamTurn` 드라이버 시험이 채택, #2·#3·#8·#10–#17 은 Codex 클라이언트 수준(`waitTurn`·`startTurn`), #4 는 가드 G2.

`TestManagedBrokerNameMatchesApprovalArgs` 는 `factoryMoAIMCPApprovalArgs()` 의 모든 서버 이름 접두가 `mcp_servers.` + `moaiMCPServerKey` + `.` 와 같음을 단언한다(두 이름이 한 값임을 고정; 승인 인수 코드는 수정하지 않는다).

## §2. RED-now / green path 쌍과 채택 상태 (verification-completeness.md §2)

`.claude/rules/moai/development/verification-completeness.md` §2, §2.1 의 두 칸 채택 규율을 따른다. 기준 트리 `7109e0900`.

### §2.1 시험 기반 릴리스 차단급 AC — **plan 시점에는 미채택, M1의 RED 기준선이 채택한다**

AC-MH-001, 002, 003, 004, 005, 006(및 그 하위 케이스 전부)은 시험이 아직 없어서 plan 단계에서 RED-now 네 요소(명령·stdout·exit·트리 SHA)를 시험 명령으로 낼 수 없다. 이 AC들은 **plan 시점에 미채택이며 M1의 RED 기준선(`red-baseline.md`)이 채택하는 것**으로 run에 넘긴다(plan-audit PASS-with-debt 경로). 계획 단계의 임시 시험 실측은 삭제되어 재실행할 수 없으므로 아래 "보조 실측"은 RED-now 셀이 아니다.

**선결 조건**: M1의 산출물이 착지할 수 있어야 한다. 산출물 경로는 추적되는 SPEC 디렉터리 안 `red-baseline.md` 다(측정: `git check-ignore -v .moai/reports/t1409/red-baseline.md` → `.gitignore:235:.moai/reports/*`, exit 0 — 착지 불가; `git check-ignore -v .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` → 출력 없음, exit 1; `internal/runtime/audit_cache.go` 의 `planArtifactNames` 는 `acceptance.md`·`design.md`·`plan.md`·`research.md`·`spec.md`·`tasks.md` 뿐이라 이 파일은 감사 캐시 해시에 들지 않는다). `red-baseline.md` 는 run 단계 증거 파일이며(manager-develop이 쓴다) SPEC 본문이 아니다.

**"옳은 이유의 RED" 규정 (한 방향)**: 컴파일 오류, `go test` 패키지 타임아웃(기본 10분), 10분 `DefaultManagedCodexTurnTimeout` 대기, 시험 인프라 실패로 붉은 것은 wrong-reason RED 이고 RED 기준선으로 인용하지 않는다. 반면 헤더의 블로킹 호출 규약이 정한 **시험 자신의 5초 watchdog 단언**이 깨져서 붉은 것(AC-MH-002 의 RED 가 바로 이것)은 그 상한이 곧 단언 대상 관측이므로 **옳은 이유의 RED** 다. 그래서 AC-MH-002·003·004 의 시험은 watchdog 없이 쓰면 안 된다(기준 트리에서 RED 가 10분 대기가 된다). `red-baseline.md` 의 "붉은 이유" 줄에 어느 쪽인지 적는다.

**M1 재현 시험 6개와 옳은 이유의 RED (기준 트리 API에 대해 컴파일되는 형태)**

| 시험 | AC | RED가 붉어야 하는 이유(명시) | green path |
|---|---|---|---|
| `TestManagedCodexServerRequestPolicy` (가짜 서버가 직접 서버 요청 11종을 보내고 응답 수신을 단언 — 새 심볼 없이) | AC-MH-001 | 소유자가 답장을 보내지 않아 가짜 서버가 유한 대기 뒤 응답 없음을 기록 | M2 |
| `TestManagedCodexTurnSurvivesServerRequest` | AC-MH-002 | 같은 이유(턴이 답 없이 진행, 5초 상한 단언 실패 — 시험 소유 상한이라 옳은 이유) | M2 |
| `TestManagedCodexServerRequestIDCollision` (프로세스 내 클라이언트 + 임시 서버) | AC-MH-003 | id 충돌 시 호출이 서버 요청을 응답으로 반환(빈 결과·오류 없음), 문자열 id면 연결 닫힘 오류 | M2 |
| `TestManagedCodexDeclinedBrokerElicitationFailsTurn` (가짜 서버가 `serverName=moai` elicitation을 `turnId` 없이 보내고 턴을 `completed` 로 끝냄, 같은 턴에 둘째 요청도 보내는 변형; 단언은 `DeliverTurn` 이 **nil 이 아닌 오류**를 돌려주는지뿐 — 표식 심볼 없이. §1.3 #13·#14 의 컴파일 가능 형태) | AC-MH-006 | 기준 트리의 `DeliverTurn` 이 nil 을 반환 | M3 |
| `TestManagedDriverIsolatesTurnFailure` (실제 `pumpManagedStreamTurn` 을 쓰는 스크립트 세션) | AC-MH-005 | 두 번째 턴의 에러 결과로 드라이버가 반환해 세 번째 턴이 전달되지 않음 | M3 |
| `TestManagedCodexNonCompletedTurnIsolated` (기존 가짜 App Server 에 **상태 순서**(예: `failed,interrupted,completed`)를 주는 환경 변수 모드를 더해 실제 Codex 소유자를 `driveManagedFactorySession` 으로 돌림: 우선 턴 뒤 `failed` 턴, `interrupted` 턴, 정상 턴; 단언은 정상 턴이 **전달됐는지**뿐 — 표식 심볼 없이) | AC-MH-006 #2 | 기준 트리의 드라이버가 첫 비완료 턴(`failed`)에서 반환해 이후 턴이 전달되지 않음 | M3 |

**RED-first 면제(이유 명시)**: (1) AC-MH-004 의 `TestManagedCodexConcurrentWrites` 는 기준 트리의 읽기 고루틴이 쓰지 않아 쓰기 경합이 존재하지 않는다(경합 검출기가 구조상 초록) — 변이 mu5 로 채택한다. (2) AC-MH-006 의 하위 17개는 한 최상위 시험 `TestManagedTurnFailureClassification` 안에 있고 이 시험은 **M3 에서 쓴다** — 새 심볼을 부르므로(#11–#17 은 `armTurn()`, #8 은 표식 오류 심볼, #5–#7·#9·#10 은 로그 sink; 설계 기준 판독) 기준 트리에서는 컴파일되지 않아 어느 행도 기준 트리 초록으로는 채택되지 않는다. 17행의 채택 경로 전수: M1 재현 시험이 RED 를 구성하는 행은 #1(`TestManagedDriverIsolatesTurnFailure`), #2(`…NonCompletedTurnIsolated`), #3·#13·#14(`…DeclinedBrokerElicitationFailsTurn`)이고, 변이(§2.4)로 채택되는 행은 #1 mu13 · #2 mu15 · #3 mu12 · #5–#7·#9 mu14 · #8 mu16 · #10 mu9 · #11 mu19·mu20 · #12 mu6 · #13 mu7·mu12 · #14 mu8·mu12 · #15 mu10 · #16 mu11 · #17 mu12 다. 변이도 RED 도 없는 가드는 #4(§2.5 G2) 하나뿐이다. (3) AC-MH-007 의 시험은 새 상수가 필요해 기준 트리에서 컴파일되지 않는다 — §2.2 의 구조 확인(상수 부재 exit 1)과 §6 의 섭동 확인으로 채택한다. (4) AC-MH-008 은 불변 가드다(§2.5 G1).

**보조 실측(기준 트리, 임시 시험은 삭제되어 재실행 불가 — RED-now 셀이 아님)**: 서버 요청 id=7을 보낸 뒤 1초간 클라이언트 프레임 없음(`no client frame within 1s: … i/o timeout`) · id 충돌: `call result="" err=<nil>`(진짜 응답 `{"genuine":true}` 소실) · 문자열 id: `err=managed codex app server connection closed`.

### §2.2 구조 확인형 AC (지금 명령으로 RED-now를 낼 수 있는 것, 재실행 확인됨)

| AC | RED-now: 명령 | stdout | exit | 트리 | green path |
|---|---|---|---|---|---|
| AC-MH-007 (상수 부분) | `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` | (빈 출력) | 1 | `7109e0900` | M3 이 상수를 정의하면 정의 1행·exit 0 |
| AC-MH-010 (문서) | `grep -c "턴 하나가 실패하면 세션 전체가 끝난다" .moai/docs/factory-managed-session.md` | `1` | 0 | `7109e0900` | sync 가 문장을 고치면 `0`·exit 1 |
| AC-MH-010 (CHANGELOG) | `grep -c "SPEC-FACTORY-MANAGED-HARDEN-001" CHANGELOG.md` | `0` | 1 | `7109e0900` | sync 가 엔트리를 쓰면 1 이상·exit 0 |

### §2.3 회귀 가드 (지금 이미 초록이 정상 — 릴리스 차단 RED가 아니라 불변 가드)

AC-MH-009(Windows 빌드: 기준 트리 `GOOS=windows GOARCH=amd64 go build ./...` 출력 없음·성공, `grep -rn 'syscall\.' internal/cli/managed_*.go` 0행·exit 1), AC-MH-012(기준 트리 diff 0행), AC-MH-008(기준 트리에서도 초록인 부정 요구; §2.5 G1), AC-MH-013.

### §2.4 변이 탐침 — 한 단계 편집으로 적용 가능하고 실제로 붉어지는 것만

변이는 지정한 마일스톤의 GREEN 직후에 **하나씩 임시로 적용**해 지목한 케이스가 붉어지는지 보이고(원문을 progress.md §E.2 에 인용) 되돌린 뒤 `git diff --stat` 빈 출력을 확인한다. 아래 mu1–mu20 은 전부 실제 코드의 **한 곳을 고치는** 편집이고, "붉어지는 케이스" 는 그 편집을 기준 코드(및 design.md 의 설계)에 걸어 따라가 본 것에 한정했다(plan 단계의 문서 대조이며 실행이 아니다 — 실제로 붉어지는지는 run 이 보인다). 한 단계 편집으로 적용되지 않거나 지목한 케이스가 실제로 붉어지지 않는 항목은 표에 없고 §2.5 에 있다.

| 변이 | 편집 | 마일스톤 | 반드시 붉어지는 케이스 |
|---|---|---|---|
| mu1 | 결정을 담는 다섯 종류(`item/commandExecution`, `item/fileChange`, `applyPatchApproval`, `execCommandApproval`, `mcpServer/elicitation`)의 응답 값을 accept 계열(`accept`/`approved`)로 | M2 | AC-MH-001 하위 5개(그 다섯 종류). 나머지는 초록 |
| mu2 | 모든 method 를 오류 `-32000` 으로 답함 | M2 | AC-MH-001 하위 **8개**: 결과 응답 7종(`item/commandExecution`, `item/fileChange`, `item/permissions`, `mcpServer/elicitation`, `item/tool/call`, `applyPatchApproval`, `execCommandApproval`)과 미지 method(기대 `-32601`). 오류 3종(`item/tool/requestUserInput`, `account/chatgptAuthTokens/refresh`, `attestation/generate`)은 이미 `-32000` 이라 초록 |
| mu3 | 오류 코드 교환: `-32000` ↔ `-32601` | M2 | AC-MH-001 하위 **4개**: `item/tool/requestUserInput`, `account/chatgptAuthTokens/refresh`, `attestation/generate`, 미지 method |
| mu4 | 응답 로그 줄을 쓰지 않음 | M2 | AC-MH-001 하위 11개 전부(로그 단언) |
| mu5 | 연결 쓰기 뮤텍스 제거 | M2 | AC-MH-004 (`-race` 경합 보고 또는 gorilla 동시 쓰기 패닉; 확률적 — 5회 재시도해도 초록이면 반복 수를 올린다) |
| mu6 | `armTurn` 의 `brokerDeclined` 초기화 제거 | M3 | AC-MH-006 #12 |
| mu7 | 귀속 조건에 "`turnID` 가 이미 알려짐"을 더함(`turn/started` 이후에만 셈) | M3 | AC-MH-006 #13 |
| mu8 | 판정 `brokerDeclined > 0` 을 `== 1` 로 | M3 | AC-MH-006 #14 |
| mu9 | `serverName` 비교를 빼 모든 서버의 거부를 셈 | M3 | AC-MH-006 #10 |
| mu10 | `prevTurnID` 비교 제거 | M3 | AC-MH-006 #15 |
| mu11 | `turnID` 확정 뒤 id 불일치 검사 제거 | M3 | AC-MH-006 #16 |
| mu12 | `waitTurn` 이 완료 이벤트가 실어 온 판정을 무시 | M3 | AC-MH-006 #3, #13, #14, #17 (#11·#12·#15·#16 은 nil 이 기대라 초록 — #11 의 채택은 mu19·mu20) |
| mu13 | 우선 턴 이후의 `is_error` 도 세션 치명으로 둠(드라이버가 표식을 무시하고 첫 오류에서 반환 — 편집 결과는 기준 코드 `managed_factory_session.go:345-347` 의 `if err := s.DeliverTurn(turn.prompt); err != nil { return err }` 와 같다) | M3 | AC-MH-005, AC-MH-006 #1, AC-MH-007 `below_ceiling_continues`(N−1번 실패 시나리오의 **첫** 실패에서 반환해 `/exit` 의 nil 대신 그 오류가 나오고 세션이 받은 턴이 모자란다). `at_ceiling_returns`·`success_resets` 도 같은 이유로 붉을 수 있으나 그 지목은 mu18·mu17 의 몫이라 여기서는 요구하지 않는다 |
| mu14 | 드라이버가 표식 없는 오류도 로그 후 계속함 | M3 | AC-MH-006 #5, #6, #7, #9 (드라이버 통과 행: `/exit` 로 nil 이 반환돼 "기대한 오류가 아님"으로 붉어짐) |
| mu15 | `waitTurn` 이 비완료 종료 오류에 표식을 붙이지 않음 | M3 | M1 `TestManagedCodexNonCompletedTurnIsolated`(`failed`·`interrupted` 각각), AC-MH-006 #2 |
| mu16 | `waitTurn` 의 `ctx.Done()` 분기가 표식 오류를 반환 | M3 | AC-MH-006 #8 (`errors.Is` 가 거짓이어야 함) |
| mu17 | 성공 턴이 연속 횟수를 0으로 되돌리지 않음 | M3 | AC-MH-007 `success_resets` |
| mu18 | 상한 비교를 `>=` 에서 `>` 로(off-by-one) | M3 | AC-MH-007 `at_ceiling_returns` **하나**: 연속 실패가 정확히 상한과 같을 때만 `>=` 와 `>` 가 갈려 N번째 실패 뒤 드라이버가 계속하고 `/exit` 의 nil 이 마지막 오류 반환 대신 나온다. N−1 이하에서는 두 비교가 모두 거짓이라 `below_ceiling_continues`(mu13 지목)와 `success_resets`(mu17 지목)는 mu18 에서 초록이다 |
| mu19 | `turn/completed` 처리가 턴 창을 닫지 않음(design.md D-1 결정 3 의 완료 처리에서 창을 닫는 한 줄만 뺌; 판정 싣기와 `prevTurnID` 기록은 그대로) | M3 | AC-MH-006 #11: 완료 뒤 후행 `moai` elicitation(`turnId` null)이 열린 창에서 규칙 (3)으로 T1 에 귀속돼 줄이 `turn=T1` 로 나가고 시험의 `turn=none` 줄 대기가 5초 상한에서 붉어진다(시험 자신의 상한이라 옳은 이유의 RED). 이 변이에서 `waitTurn` 의 nil 단언은 초록이다 — 판정이 후행 요청보다 먼저 완료 이벤트에 실렸다 |
| mu20 | 닫힌 창 분기(design.md D-1 결정 3 귀속 규칙 (1))도 거부 수를 올리고 줄에 그 값을 씀(귀속 턴은 여전히 `none`) | M3 | AC-MH-006 #11 의 `broker_declined=0` 단언: 줄이 `turn=none … broker_declined=1` 로 나간다. mu19 와 달리 `turn=none` 대기는 통과하므로 이 변이를 잡는 것은 값 단언뿐이다 |

**편집이 아닌 점검 둘(DoD 3·4 의 일부, 변이 번호를 주지 않음)**: C1 — RED와 수리를 한 커밋에 합치지 않았는지 `merge-base --is-ancestor` 쌍으로 확인(AC-MH-011; M4 점검). P1 — 섭동 확인: `defaults.go` 의 연속 실패 상한을 일시적으로 2 로 바꿔 AC-MH-007 시험이 **여전히 PASS** 하는지(드라이버가 숫자 3 을 박았다면 붉어짐)와 AC-MH-007 의 `grep` 참조 확인.

### §2.5 불변 가드 — 변이 미채택 (이유와 함께). 이 목록 밖의 항목은 AC-MH-001..006 의 시험 기반 AC(M1 RED 또는 §2.4 의 mu1–mu20; AC-MH-006 17행의 전수는 §2.1 면제 (2)), AC-MH-007(§2.2 의 상수 구조 확인과 mu13·mu17·mu18·P1), AC-MH-010(§2.2·§1.2 의 grep), AC-MH-011(C1)로 채택되고, 이 절은 변이도 RED 도 없는 항목만 적는다

- **G1 — AC-MH-008 `TestManagedFailedTurnLeavesClaimUntouched`**: 기준 트리에서도 초록이고(관리 계층이 아무것도 하지 않는 부정 요구), 드라이버에는 store 핸들이 없으며(`driveManagedFactorySession` 시그니처 `managed_factory_session.go:300` — claim·prompt 클로저뿐) claim 경로(`claimManagedFactoryInbox`)는 턴 실패를 알지 못한다. 한 단계 편집으로 붉게 만들 변이가 없다. 재배달 증거는 기존 `TestDispatchResultExactlyOnce/lost_receipt_redelivery` 가 운반한다(이 plan 실행에서 다시 돌려 7개 PASS 확인, `ok …/internal/factorymsg 2.167s`).
- **G2 — AC-MH-006 #4 `priming_is_error`**: 우선 턴은 루프 앞에서 표식과 무관하게 무조건 반환한다(`managed_factory_session.go:302-304`). 드라이버 분류를 바꾸는 변이(mu14)가 닿지 못한다. 기존 시험 `TestManagedDriverFailureBranches` 의 "a priming turn that fails ends the session before any poll" 가 이미 이 동작을 고정한다.
- **G3 — `TestManagedBrokerNameMatchesApprovalArgs`**: 두 상수가 같은 값임을 고정하는 데이터 동등성 가드다. 어느 한 상수를 바꾸면 붉어지는 것은 자명하며 변이로 세지 않는다.
- **G4 — AC-MH-009, AC-MH-012, AC-MH-013**: 회귀·불변 가드(기준 트리에서도 초록이 정상).
- **G5 — 직전 턴의 중복·지연 `turn/completed`(`X == prevTurnID`) 방어(design.md D-1)**: codex 가 한 턴에 완료 프레임을 한 번만 보내는지 관측하지 못했다 — 방어만 두고 시험 행도 변이도 두지 않는다.
- **G6 — 동치 변이 "요청마다 실패로 세기"**: 한 턴이 `DeliverTurn` 오류 하나만 내므로 관측 가능한 차이가 없다. 가를 수 있는 형태로 mu8 을 택했다(둘째 요청 로그의 `broker_declined=2` 와 실패 1회가 그 관측).
- **G7 — 설계 대체 탐침(소비자 쪽 계수기)**: design.md D-1 의 기각한 대안 (e)로 되돌리는 것은 한 단계 편집이 아니라 설계 교체다. §1.3 #11 의 `waitTurn` nil 단언은 그 설계에서 붉어지도록 짰지만(후행 요청이 소비자 읽기보다 먼저 계수기에 닿는다) DoD 가 그것을 적용하라고 요구하지 않는다 — #11 자신의 채택은 한 단계 변이 mu19·mu20 이 운반한다(위 §2.4). 굳이 설계 교체를 적용한다면 가장 작은 형태: 읽기 고루틴이 거부된 브로커 elicitation 마다 원자 계수기를 올리고, `startTurn` 이 0 으로 되돌리고, `waitTurn` 이 완료 이벤트 소비 시점에 계수기를 읽도록 바꾼다(판정 이벤트 필드는 무시).
- **G8 — "구현에 상수 값 `3` 을 박기"**: 한 단계 변이로는 시험이 상수를 읽어 3 이 되므로 동치다. 섭동 확인 P1 과 `grep` 참조 확인으로 방어한다.

## §3. 기준선과 알려진 사전 적색 (측정)

| 대상 | 관측 (기준 트리 `7109e0900`) |
|---|---|
| 스코프 회귀(정리 접두) | `ok  github.com/modu-ai/moai-adk/internal/cli  27.048s` (앵커형), 리터럴 부분 문자열 패턴으로 돌린 첫 측정은 `ok … 30.959s`. `-v` 최상위 66개, 두 패턴이 같다 |
| 같은 선택, 정리 접두 없이(레인 환경) | `--- FAIL: TestManagedSwitchDoesNotReachCodexLaneLoop (0.48s)` — `todo add: moai add: refused — lane boundary: a lane session cannot mutate the queue …`, `FAIL … 26.413s`. 레인 환경이 원인(정리 후 PASS 13.40s). 이 카드가 만든 적색이 아니다 |
| 선택 수: 이스케이프 없는 형태 대 표 이스케이프 형태 | 명령은 아래 블록. 이스케이프 없는 형태 66(이 개정에서 재측정), 백슬래시 파이프 형태 0 (plan-audit 측정) |
| AC-MH-004·006 패턴의 선택 수(시험 부재) | `-list` 로 `TestManagedCodexConcurrentWrites`·`TestManagedTurnFailureClassification`·`TestManagedBrokerNameMatchesApprovalArgs` 를 고른 출력이 `ok  github.com/modu-ai/moai-adk/internal/cli  1.352s` 한 줄 — `Test` 이름 0개(시험 부재 → 공허 `ok`의 실례) |
| config | `ok  …/internal/config  31.134s` |
| template | `ok  …/internal/template  414.986s` (분 단위 스위트 → 임대 안에서만 실행) |
| guardstate | **기준 트리에서 이미 적색**: `--- FAIL: TestCensus_SetDifferenceEmptyBothDirections (0.00s)` — `census_test.go:80: disk\manifest: .github/workflows/workflow-parse-guard.yaml exists on disk with no manifest entry`, `census_test.go:89: declared 19 entries against 20 workflow files`. 이 카드는 워크플로 파일을 건드리지 않으므로 사전 존재이며, AC-MH-013 은 "이 한 테스트 외 신규 실패 0건"으로 읽는다 |
| 템플릿 미러 | `ls internal/template/templates/.moai/docs/factory-managed-session.md` → `No such file or directory`(미러 없음) |

선택 수 측정 명령 원문(두 번째 줄의 백슬래시 파이프는 일부러 이스케이프한 반증 형태):

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^.*(Managed|managed).*$' | grep -c '^Test'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^.*(Managed\|managed).*$' | grep -c '^Test'
```

## §4. 에지 케이스

- 서버 요청이 턴 사이(한가한 때)에 온다 — 읽기 고루틴이 바로 답한다(design.md D-1 이유 1; AC-MH-002 `between_turns`). `moai` elicitation이 그 때 오면 어느 턴에도 세지 않는다(AC-MH-006 #11).
- 서버 요청 두 개가 연달아 온다 — 각각 정확히 한 번 답하고 순서가 보존된다.
- 답장 쓰기가 실패한다 — 읽기 고루틴이 끝나고 진행 중 `DeliverTurn` 은 "connection closed" 로 세션 치명 오류가 된다.
- 알 수 없는 알림(id 없음)은 지금처럼 버린다. 답하지 않는다. `id` 가 없거나 원문이 JSON `null` 인 프레임은 "id 없음"으로 본다 — `method` 가 있으면 알림, 없으면 버린다(design.md D-1; 이 경로는 변이도 시험 행도 두지 않았다).
- 직전 턴의 중복·지연 `turn/completed` 는 창을 닫지 않는다(방어만 있음; §2.5 G5). 막힌 쓰기가 풀리는 길은 연결 사망과 세션 닫힘뿐이고 그 상한은 문서가 "상한 없음"으로 적는다.
- 직전에 완료된 턴 id 를 단 늦은 요청, 확정된 현재 턴과 다른 id 를 단 요청은 세지 않는다(AC-MH-006 #15·#16 — design.md D-1 의 약속 문장과 같은 문장).
- 운영자 입력으로 시작된 턴이 실패해도 같은 횟수에 센다. 성공한 운영자 턴은 횟수를 0으로 되돌린다.
- 우선 턴이 `is_error` 로 끝나도 세션 치명이다(턴 단위 분류는 우선 턴 이후에만 적용: REQ-MH-006).
- 연결 쓰기에 데드라인이 없다 — 서버가 읽지 않으면 쓰기가 막히고 턴 타임아웃도 그것을 풀지 못한다. 연결이 죽거나 세션이 닫힐 때까지 상한이 없다(공시한 한계).
- `/exit` 와 stdin EOF 경로는 기존 동작 그대로다(AC-MH-013 스코프 회귀가 고정).

## §5. 품질 게이트

- TRUST 5 — Tested: 변경한 비테스트 코드의 줄이 재현·분류·귀속 시험으로 덮인다(`go test -cover` 를 managed 슬라이스에 한정해 측정, 부모 기준선 89.3%보다 내려가지 않는다 — 부모 sync 기록의 수치이며 이 plan에서 재측정하지 않았다). Readable/Unified: `gofmt`·`go vet`·`golangci-lint` 0건. Secured: 승인류 응답에 accept 계열 0건(AC-MH-001). Trackable: Conventional Commit, 모든 커밋 메시지에 카드 id `t1409`.
- 경계 grep(부모 AC 계승): `grep -rnE 'merge-window|Decider|T29b|T29c|handover' internal/cli/managed_*.go` 0행, `grep -rn 'AskUserQuestion' internal/cli/managed_*.go` 비테스트 0행.

## §6. Definition of Done

1. AC-MH-001..013 전부 PASS — 명령과 원문 출력을 progress.md §E.2 에 인용(AC-MH-010 은 sync 후 §E.4).
2. 재현 시험 6개의 RED 원문이 **추적되는** `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 에 있고 RED 커밋이 수리 커밋의 조상이다(AC-MH-011; 점검 C1).
3. §2.4 의 변이 **mu1–mu20**(적용 가능하고 지목 케이스가 실제로 붉어지는 것만 남긴 20개)을 **하나씩** 적용해 지목한 케이스가 붉어지는 것을 보인 원문이 증거로 있다. §2.5 의 가드 G1–G8 은 이 범위에 없다.
4. **섭동 확인 P1**: `defaults.go` 의 연속 실패 상한을 일시적으로 2 로 바꿔 AC-MH-007 시험이 여전히 PASS 하는 것(드라이버가 숫자를 박았다면 붉어짐)을 인용하고 되돌린 뒤 `git diff --stat` 빈 출력을 확인한다.
5. Windows 크로스 빌드·vet가 exit 0이고 `managed_*.go` 의 `syscall.` 참조가 0행이다(AC-MH-009).
6. 부모 SPEC 디렉터리와 `store.go` 의 diff가 비어 있다(AC-MH-012).
7. 운영 문서와 CHANGELOG가 F3·F4 해결과 남은 한계를 과장 없이 적고 F5 를 t1459 로 가리키며, t1409 단락 삭제가 grep으로 확인된다(AC-MH-010).
