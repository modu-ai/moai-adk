---
id: SPEC-FACTORY-MANAGED-CARD-CHILD-001
title: "acceptance.md — 인수 기준"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# acceptance.md — AC-CC-001..013

> 이 파일은 검증층이고 시나리오는 Given-When-Then 이다. 요구(GEARS)는 `spec.md` §C 의 REQ-CC-001..012 가 운반한다. 모든 AC 는 이진 판정이 가능하고 실제 codex 를 쓰지 않는다 — 가짜 App Server(기존 `fakeAppServerScript`, POSIX 전용)와 기존 이음새(`codexLookPath`, `codexDirectLaunchFn`, `managedFactoryCodexLaunchFunc`, `codexWorktreeAnchorLock`)만 쓴다.
>
> **명령 규약.** `-run` 패턴은 앵커형(`^…$`)이다. 파이프를 담은 명령은 §1.1 의 fenced 블록에 원문으로 적고 표에서는 AC id 로 가리킨다. 부하 규율상 `go test ./internal/cli` 전체는 로컬에서 돌리지 않는다.
>
> **레인 세션에서 `go test` 는 환경 정리 복합형으로 돌린다**: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test …`. 아래 표의 명령은 그 접두 없이 적는다. (기준 트리 측정에서 이 접두를 붙인 `TestManagedSwitchDoesNotReachCodexLaneLoop` 가 PASS.)
>
> **빈 스윕 규율.** 모든 `-run` 명령은 읽기 전에 `go test ./internal/cli -list '<같은 패턴>'` 로 선택된 이름 수를 먼저 확인한다. `[no tests to run]` 와, `-list` 가 `Test` 이름을 하나도 안 내는 `ok` 는 통과가 아니라 **미측정**이다. 이 문서의 모든 새 시험은 기준 트리에서 아직 없으므로 `-list` 가 0개다(§2 R-2).
>
> **블로킹 호출 규약.** 가짜 App Server 를 쓰는 시험은 소유자 호출을 고루틴에서 돌리고 시험 고루틴이 **절대 상한 30초**(기존 시험과 같은 값)에 `select` 로 실패시킨다. 정리에서 가짜 서버와 연결을 닫는다. stdin 은 `os.Pipe`/`io.Pipe` 로 주입하고 `/exit` 로 끝낸다.
>
> **측정 기준 트리**: `2b9e4a4d0`(카드 기준 develop). "측정" 표기는 이 plan 실행이 관측한 것이다.

## §1. AC 매트릭스

| AC | 요구 | 시나리오 (Given-When-Then) | 검증 명령 → 기대 관측 |
|---|---|---|---|
| AC-CC-001 | REQ-CC-001 | Given 옵트인 `MOAI_FACTORY_MANAGED=1` 과 레인 스탬프(`MOAI_KANBAN_ID`, `MOAI_FACTORY_WORKERS`)가 환경에 있고 운영자가 고른 카드 두 장(`t1`, `t2`)이 큐에 있으며 `managedFactoryCodexLaunchFunc` 와 `codexDirectLaunchFn` 이 호출을 세는 스텁일 때 When `moai codex -f lane` 을 돌리면 Then 관리 소유자 스텁이 **정확히 2번**(카드마다 1번) 불리고 직접 exec 스텁은 **0번** 불리며, 각 호출의 `dir` 은 그 카드의 워크트리, 환경의 `MOAI_KANBAN_CARD` 는 그 카드의 id 다(`t1`, `t2` 순). 루프는 `next` 가 카드가 없다고 답할 때 오류 없이 끝난다 | `go test ./internal/cli -run '^TestManagedCardChildLaneLoopUsesManagedOwner$' -count=1 -v` → 하위 `switch_1`, `switch_TRUE_padded`(값 ` TRUE `, 대소문자·공백) 2개 PASS, exit 0 |
| AC-CC-002 | REQ-CC-002 | Given 같은 카드 두 장과 스위치가 **꺼진** 환경 값 — 미설정, 빈 문자열, `0`, `yes` 네 가지 — 이고 두 스텁이 호출을 셀 때 When 레인 루프를 돌리면 Then 직접 exec 스텁이 카드마다 1번(합 2번), 소유자 스텁은 0번 불리고, 직접 exec 스텁이 받은 `*exec.Cmd` 는 현행 형태다: `Args` 에 `-C`·카드 워크트리 쌍, `Dir` = 카드 워크트리, 환경에 역할 표식·레인 라벨 두 운반체·백엔드 `gpt`·카드 id 가 있고 `MOAI_KANBAN_ID` 는 **없다**. 같은 변경이 기존 직접 문 시험(`TestSD_AC003_CodexRelaunchPerCard`, `TestSD_AC004_CodexOtherFactoryShapesRefused`)을 깨지 않는다 | `go test ./internal/cli -run '^TestManagedCardChildSwitchOffKeepsDirectDoor$' -count=1 -v` → 하위 4개(`unset`, `empty`, `zero`, `yes`) PASS · §1.1 의 AC-CC-002 블록(기존 직접 문 시험 2개; 정규식의 파이프를 표 밖에 원문으로 둔다) → 선행 `-list` 가 정확히 2개 이름, 이어지는 `-run` 이 2개 PASS, exit 0 |
| AC-CC-003 | REQ-CC-003 | Given 변경이 반영된 트리 When 플레인 `moai codex` 분기·spawn 문·루프 정지 시험을 돌리면 Then 전부 이전과 같이 PASS 다: 플레인 분기 `TestManagedCodexLaunchRequiresOptIn`(하위 3개)·`TestManagedCodexLaunchDivertsFactorySession`·`TestManagedCodexLaunchCarriesLaunchDir`·`TestManagedCodexLaunchSkipsLaterDebugSteps`, spawn 문 `TestManagedCodexLaunchKeepsSpawnDoor`, 루프 정지·lease `TestSD_AC003_CodexRelaunchPerCard` | §1.1 의 AC-CC-003 블록 — 선행 `-list` 가 **정확히 6개** 최상위 이름을 내야 하고, 이어지는 `-run` 이 모두 PASS, exit 0 |
| AC-CC-004 | REQ-CC-004 | Given 카드 워크트리에 `AGENTS.local.md` 가 있고 옵트인이 켜졌을 때 When 레인 루프가 소유자 스텁을 부르면 Then (i) `args[0]` 은 `codexLookPath` 가 돌려준 바이너리이고 `-C` 토큰이 **없고**, (ii) `-c` 다음 토큰이 `codexLocalDeveloperInstructionArgs(<카드 워크트리>)` 가 만든 `developer_instructions=…` 값과 같으며, (iii) `dir` 이 카드 워크트리다. 하위 케이스: 로컬 지침이 크기 상한을 넘으면 소유자 스텁은 불리지 않고 stderr 에 `codex lane: card <id> session:` 한 줄이 남으며 루프는 이어진다. 전제 고정 하위 케이스: `managedCodexOptions([]string{bin, "-C", "/x"})` 가 `-C` 를 이름에 담은 오류로 거부한다(소유자가 `-C` 를 받지 않는다는 사실이 `-C` 를 뺀 이유다) | `go test ./internal/cli -run '^TestManagedCardChildLaunchShape$' -count=1 -v` → 하위 `no_dash_C_and_pair_and_dir`, `oversize_instruction_skips_owner`, `owner_refuses_dash_C` 3개 PASS |
| AC-CC-005 | REQ-CC-005 | Given 옵트인이 켜진 레인 루프가 라벨 `lane-N` 을 claim 했고 run id 가 `fcRun` 일 때 When 소유자 스텁이 환경을 받으면 Then 환경에 `MOAI_FACTORY_ROLE=lane`(값 상수 `config.FactoryRoleLane`), `MOAI_FACTORY_WORKER=lane-N`, `MOAI_KANBAN_LABEL=lane-N`, `MOAI_KANBAN_BACKEND=gpt`, `MOAI_KANBAN_CARD=<카드 id>`, `MOAI_KANBAN_ID=<fcRun>` 이 모두 있고 `MOAI_FACTORY_WORKERS` 와 `CLAUDE_CODE_SESSION_ID` 는 **없다**. 같은 키를 직접 문 쪽은 `MOAI_KANBAN_ID` 만 없다(AC-CC-002 와 대조) | `go test ./internal/cli -run '^TestManagedCardChildEnvCarriesIdentity$' -count=1 -v` → PASS |
| AC-CC-006 | REQ-CC-006 | Given 옵트인이 켜진 레인 루프와 호출을 세는 `codexWorktreeAnchorLock` 스텁 When 두 카드를 소유자 스텁으로 돌리면 Then 앵커 락 스텁은 0번 불리고, 소유자 스텁이 불리는 **그 순간** 레인 레지스트리(`loadFactoryRegistry(factoryRegistryPath(root))[label]`)의 PID 가 `os.Getpid()` 이며 루프가 끝난 뒤에도 같다 | `go test ./internal/cli -run '^TestManagedCardChildKeepsLauncherClaim$' -count=1 -v` → PASS |
| AC-CC-007 | REQ-CC-007 | Given 실제 브로커 저장소(임시 루트)와 가짜 App Server 를 바이너리로 쓰고, 기본 소유자 이음새를 그대로 두고, 카드 두 장을 같은 런처 프로세스가 연속으로 돌릴 때 When 각 세션이 stdin `/exit` 로 끝나면 Then 루프는 오류 없이 끝나고, 브로커 roster 에 레인 endpoint 가 정확히 1개이며 `BindingBound`·세션 UUID `fake-thread-1` 이고, 가짜 서버 로그에 `method thread/start` 가 정확히 2줄이다(두 번째 카드의 등록이 첫 카드의 바인딩된 행에 막히지 않고 교체했다). 하위 케이스 `start_failure_leaves_no_pending_row`: 바이너리가 존재하지 않아 소유자 `Start` 가 실패하면 roster 에 레인 endpoint 가 **0개**이고 루프는 두 카드를 모두 시도한 뒤 끝난다. 양성 대조: 같은 roster 조회가 `registerFactoryLaunchPending` 직후에는 launch-pending 행을 1개 낸다 | `go test ./internal/cli -run '^TestManagedCardChildSecondCardRebinds$' -count=1 -v` → 하위 `sequential_cards_rebind`, `start_failure_leaves_no_pending_row`, `positive_control_pending_row_visible` 3개 PASS (POSIX; Windows 는 skip — 가짜 서버가 POSIX 셸 shim 이다) |
| AC-CC-008 | REQ-CC-008 | Given 소유자·드라이버 파일 When 변경 diff 를 조회하고 기존 소유자 시험을 돌리면 Then 두 파일의 diff 가 비어 있고 — 새 구현이 없다 — 기존 소유자 시험이 이전과 같이 PASS 다. 또한 **루프 경유 배달 증거**: Given AC-CC-007 의 실제 소유자 배선에서 레인이 바인딩된 뒤 브로커에 메시지 1건을 보내 When 드라이버가 한가할 때 claim 하면 Then 가짜 서버 로그에 `method turn/start` 가 우선 턴 포함 2줄 이상 나타난다(상한 5초 폴링), 그 뒤 `/exit` 로 끝난다 | `git diff --stat develop...HEAD -- internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go internal/factorymsg/store.go` → 출력 0행, exit 0 (양성 대조: 같은 형태를 `internal/cli/codex_launcher.go` 에 주면 1행 이상) · `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1` → ok · `go test -race ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$' -count=1` → ok · `go test ./internal/cli -run '^TestManagedCardChildDeliversInboxThroughLoop$' -count=1 -v` → PASS |
| AC-CC-009 | REQ-CC-009 | Given 옵트인이 켜진 레인 루프에서 소유자 스텁이 첫 카드에서는 오류를, 둘째 카드에서는 nil 을 돌려줄 때 When 루프를 돌리면 Then stderr 에 `codex lane: card t1 session: <오류>` 한 줄이 남고, 소유자 스텁은 2번 불리며(루프가 첫 오류에서 멈추지 않는다), 루프는 오류 없이 끝난다. 같은 조건에서 스위치가 꺼진 경로의 같은 줄 형태(`codexDirectLaunchFn` 오류)와 문구가 같다 | `go test ./internal/cli -run '^TestManagedCardChildSessionEndContinuesLoop$' -count=1 -v` → 하위 `owner_error_logged_and_loop_continues`, `same_line_form_as_direct_door` 2개 PASS |
| AC-CC-010 | REQ-CC-010 | Given 프로세스 수명 입력 펌프(소스는 `io.Pipe`)에서 어댑터 A 가 붙었다가 닫히고 어댑터 B 가 붙을 때 When A 가 읽기 중 닫히거나 A 가 닫힌 뒤 소스에 줄이 쓰이면 Then 막힌 A 의 `Read` 는 1초 안에 `io.EOF` 로 풀리고, A 가 닫힌 뒤 쓴 줄은 A 에 **절대** 가지 않고 B 가 붙은 뒤 B 에 정확히 한 번 간다. 루프 수준: Given AC-CC-007 의 실제 소유자 배선과 공유 stdin 소스에서 첫 카드 세션이 `/exit` 로 끝난 뒤 When 두 번째 `/exit` 줄을 쓰면 Then 둘째 카드 세션이 그 줄로 끝나고(30초 watchdog 안) 루프가 정상 종료한다 | §1.1 의 AC-CC-010 블록 — 선행 `-list` 가 **정확히 2개** 최상위 이름(`TestManagedOperatorInputPumpDetachesEndedSession`, `TestManagedCardChildOperatorInputReachesNextSession`)을 내야 하고, 이어지는 `-race` `-run` 이 2개 모두 PASS, `WARNING: DATA RACE` 0건, exit 0 |
| AC-CC-011 | REQ-CC-011 | Given 변경이 반영된 트리 When Windows 대상으로 빌드·vet 하고 새 파일에서 `syscall.` 참조를 훑으면 Then 빌드와 vet 가 통과하고 새 파일의 `syscall.` 참조는 0행이다 | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 · `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → exit 0 · `grep -n 'syscall\.' internal/cli/managed_operator_input.go internal/cli/managed_operator_input_test.go internal/cli/managed_card_child_test.go` → 출력 0행, exit 1 (양성 대조: `grep -n 'syscall\.' internal/cli/codex_direct_posix.go` 는 1행 이상, exit 0) |
| AC-CC-012 | REQ-CC-012 | Given sync 단계가 끝난 트리 When 운영 문서와 CHANGELOG 를 훑으면 Then 카드 자식이 관리되지 않는다는 옛 문장 둘이 사라지고, 새 범위 문장과 SPEC 포인터가 있으며, 남는 한계(헤드리스·시그널)가 t1408·t1459 포인터와 함께 남고, CHANGELOG 에 이 SPEC 엔트리가 있다 | §1.2 의 grep 목록(각 줄이 기대하는 개수·exit 코드 포함) |
| AC-CC-013 | REQ-CC-001..012 (게이트) | Given 변경이 끝난 트리 When 형식·정적 분석·스코프 회귀를 돌리면 Then 오류 0건이고 회귀 없음 | §1.1 의 AC-CC-013 블록 — 선택 수 79(기준 트리 측정) 이상 + 새 최상위 시험 이름 수만큼 증가, 뒤 `ok`, `[no tests to run]` 는 실패 · `gofmt -l internal/cli` → 출력 0행 · `go vet ./internal/cli` → exit 0 · `golangci-lint run ./internal/cli/...` (CI 판 v2.1.6) → exit 0 · `git diff --stat develop...HEAD -- .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001` → 출력 0행, exit 0 |

### §1.1 파이프·이스케이프를 담은 명령 원문 (표 밖, 바이트 단위 원문)

**AC-CC-002 (기존 직접 문 시험 2개)**

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^(TestSD_AC003_CodexRelaunchPerCard|TestSD_AC004_CodexOtherFactoryShapesRefused)$'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -run '^(TestSD_AC003_CodexRelaunchPerCard|TestSD_AC004_CodexOtherFactoryShapesRefused)$' -count=1 -v
```

**AC-CC-003**

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^(TestManagedCodexLaunchRequiresOptIn|TestManagedCodexLaunchDivertsFactorySession|TestManagedCodexLaunchCarriesLaunchDir|TestManagedCodexLaunchSkipsLaterDebugSteps|TestManagedCodexLaunchKeepsSpawnDoor|TestSD_AC003_CodexRelaunchPerCard)$'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -run '^(TestManagedCodexLaunchRequiresOptIn|TestManagedCodexLaunchDivertsFactorySession|TestManagedCodexLaunchCarriesLaunchDir|TestManagedCodexLaunchSkipsLaterDebugSteps|TestManagedCodexLaunchKeepsSpawnDoor|TestSD_AC003_CodexRelaunchPerCard)$' -count=1 -v
```

**AC-CC-010**

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^(TestManagedOperatorInputPumpDetachesEndedSession|TestManagedCardChildOperatorInputReachesNextSession)$'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test -race ./internal/cli -run '^(TestManagedOperatorInputPumpDetachesEndedSession|TestManagedCardChildOperatorInputReachesNextSession)$' -count=1 -v
```

**AC-CC-013** (스코프 회귀 — 이름에 `Managed`·`managed` 를 가진 최상위 테스트)

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^.*(Managed|managed).*$' | grep -c '^Test'
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ./internal/cli -run '^.*(Managed|managed).*$' -count=1
```

기대: 첫 줄이 **89 이상**(기준 79 + 새 최상위 시험 10개), 둘째 줄이 `ok`, `[no tests to run]` 가 없다. 새 최상위 시험 10개: `TestManagedCardChildLaneLoopUsesManagedOwner`, `TestManagedCardChildSwitchOffKeepsDirectDoor`, `TestManagedCardChildLaunchShape`, `TestManagedCardChildEnvCarriesIdentity`, `TestManagedCardChildKeepsLauncherClaim`, `TestManagedCardChildSecondCardRebinds`, `TestManagedCardChildDeliversInboxThroughLoop`, `TestManagedCardChildSessionEndContinuesLoop`, `TestManagedCardChildOperatorInputReachesNextSession`, `TestManagedOperatorInputPumpDetachesEndedSession`. 기존 `TestManagedSwitchDoesNotReachCodexLaneLoop` 는 AC-CC-001 의 새 시험으로 **개명·단언 반전되어 대체**되므로 개수는 +0 이다(대체 뒤 그 이름은 선택에 없어야 한다 — `-list` 로 확인). (기준 측정: 79, 이 plan 실행, 트리 `2b9e4a4d0`.)

### §1.2 AC-CC-012 의 grep 목록 (원문, sync 후 트리에서)

| 명령 | 기대 |
|---|---|
| `grep -c "카드 자식 세션은 스위치 값과 관계없이 관리되지 않고" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1` — 측정) |
| `grep -c "Codex 레인 카드 자식과 디버그 추적에 관리 계층을 연결하는 일은 아직 카드가 없는 후속 과제다" .moai/docs/factory-managed-session.md` | `0`, exit 1 (기준 트리 `1` — 측정) |
| `grep -c "SPEC-FACTORY-MANAGED-CARD-CHILD-001" .moai/docs/factory-managed-session.md` | 1 이상 (기준 트리 `0` — 측정) |
| `grep -c "Codex 관리 세션은 화면에 아무것도 보여 주지 않는다" .moai/docs/factory-managed-session.md` | 1 이상 — 헤드리스 한계는 남아야 한다(기준 `1` — 측정) |
| `grep -c "런처는 시그널을 처리하지 않는다" .moai/docs/factory-managed-session.md` | 1 이상 — 시그널 한계는 남아야 한다(기준 `1` — 측정) |
| `grep -c "t1408" .moai/docs/factory-managed-session.md` · `grep -c "t1459" .moai/docs/factory-managed-session.md` | 각 1 이상(기준 각 `1` — 측정; 두 포인터는 지워지면 안 된다) |
| `grep -c "SPEC-FACTORY-MANAGED-CARD-CHILD-001" CHANGELOG.md` | 1 이상 (기준 트리 `0` — 측정) |

문서가 새로 말해야 하는 내용(앵커 문자열은 manager-docs 가 쓰고 과장 없이): (1) 옵트인을 켜면 `moai codex -f lane` 의 카드 자식이 관리 Codex 소유자로 뜬다, 꺼 두면 지금과 같은 직접 exec 이다; (2) 관리 카드 자식도 헤드리스다 — 레인 터미널에 모델 출력이 없다; (3) 카드 자식 환경이 직접 문과 다른 두 곳(run id 가 실린다, `MOAI_SESSION_PID` 가 실린다); (4) 입력이 끝난 `/exit` 와 다음 카드 사이의 stdin 규칙(알려진 비대칭 포함); (5) stdin EOF 인 비대칭 레인은 `/exit`·`/quit`·치명 오류 없이는 세션이 끝나지 않는다; (6) 개발자 지침 쌍이 스레드에 적용되는지는 미관측이다.

## §2. RED-now 원장 (verification-completeness §2·§2.1)

모든 측정은 트리 `2b9e4a4d0`(카드 기준 develop; SPEC 아티팩트만 미커밋, 소스 동일), 워크트리 `.claude/worktrees/t1440`, 이 plan 실행에서 얻었다.

| id | 명령 | stdout (원문) | exit | RED 이유 |
|---|---|---|---|---|
| R-1 | `go test ./internal/cli -run '^TestManagedSwitchDoesNotReachCodexLaneLoop$' -count=1 -v` (정리 접두 포함) | `=== RUN   TestManagedSwitchDoesNotReachCodexLaneLoop` … `--- PASS: TestManagedSwitchDoesNotReachCodexLaneLoop (10.99s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	12.381s` | 0 | 이 시험은 스위치+스탬프 환경에서 **소유자 0번·직접 exec 2번**을 단언하고 PASS 한다(`managed_optin_test.go:217-224`). AC-CC-001 의 새 시험은 같은 픽스처의 단언을 **뒤집은** 것(소유자 2번·직접 0번)이므로 이 트리에서 "소유자 호출 0 ≠ 2" 로 붉다 — 올바른 이유다(관리 경로 미연결). 이 시험은 run 단계에서 개명·반전으로 대체된다 |
| R-2 | 아래 R-2 블록의 `-list` (앵커 패턴 `^(TestManagedCardChild.*` 또는 `TestCardChild.*)$`) | `ok  	github.com/modu-ai/moai-adk/internal/cli	1.131s` (`Test` 이름 0개) | 0 | 새 시험이 아직 없다 — 빈 스윕이며 **미측정**으로 읽는다(통과 아님). `-list` 가 새 이름을 내야 비로소 선택이 성립한다 |
| R-3 | §1.1 의 AC-CC-013 블록 첫 줄(`-list` 뒤 `grep -c '^Test'`) | `79` | 0 | 회귀 하한 기준값(AC-CC-013) |
| R-4 | §1.2 의 첫 두 `grep -c` | `1`, `1` | 0 (매치 있음) | 옛 문장이 아직 문서에 있다 |
| R-5 | `grep -c "SPEC-FACTORY-MANAGED-CARD-CHILD-001" .moai/docs/factory-managed-session.md` · `… CHANGELOG.md` | `0`, `0` | 1 | 새 포인터가 아직 없다 |
| R-6 | `grep -n 'syscall\.' internal/cli/managed_operator_input.go` (새 파일 3개 중 하나로 측정) | `ugrep: warning: internal/cli/managed_operator_input.go: No such file or directory` | 2 | 새 파일이 아직 없다(AC-CC-011 의 정상 RED; 파일 생성 뒤 exit 1 로 뒤집힌다) |

R-2 블록 (표 밖 원문):

```
go test ./internal/cli -list '^(TestManagedCardChild.*|TestCardChild.*)$'
```

**green 경로 (쌍의 둘째 칸)**: AC-CC-001·004·005·006·009 는 M2(분기·형태)가, AC-CC-007·008 의 배선·AC-CC-010 은 M3(입력 펌프)가, AC-CC-012 는 sync 단계가 뒤집는다. AC-CC-002·003 은 **처음부터 초록이어야 하는 불변 가드**다(기준 트리에서 현행 직접 문 시험이 PASS) — 그래서 변이로 붉어지는지를 아래 변이 표로 보인다.

## §2.1 변이 표 (한 단계 편집으로 AC 를 붉히는 것)

| id | 변이 | 붉어지는 AC |
|---|---|---|
| mu1 | 분기를 무시하고 레인 루프가 항상 관리 소유자로 보낸다(스위치 확인 삭제) | AC-CC-002 (소유자 0번 단언) |
| mu2 | 관리 분기에서 `-C <wt>` 를 인수에 그대로 넣는다 | AC-CC-004 `no_dash_C_and_pair_and_dir` |
| mu3 | 관리 환경에서 run id 를 더하지 않는다 | AC-CC-005, AC-CC-007(소유자가 `requires a factory run id` 로 거부) |
| mu4 | 소유자 오류에서 루프가 `return` 한다 | AC-CC-009 |
| mu5 | 관리 분기에 `codexWorktreeAnchorLock` 호출을 더한다 | AC-CC-006 |
| mu6 | 관리 분기가 소유자와 직접 exec 문을 **둘 다** 부른다 | AC-CC-001 (직접 0번 단언) |
| mu7 | 기본 소유자 이음새가 어댑터 대신 `os.Stdin` 을 그대로 넘긴다 | AC-CC-010 루프 수준(둘째 세션이 입력을 못 받아 30초 watchdog 에서 붉다) |
| mu8 | 펌프가 닫힌 어댑터에도 조각을 전달한다 | AC-CC-010 단위 시험 |

(한 단계 편집으로 적용 불가능한 불변 보장 — 예: "소유자·드라이버 파일 무수정" — 은 변이 대신 AC-CC-008 의 diff 가드가 운반한다.)

## §3. 에지 케이스

1. 카드 워크트리 생성 실패(`factoryEnsureCardWorktree` 오류): 현행과 같이 루프가 `codex lane: …` 로 반환한다. 관리 분기와 무관(분기 앞 단계).
2. 스위치 값 `TRUE ` (대문자·공백): `factoryManagedRequested` 가 소문자·trim 해서 켜짐으로 읽는다(`factory_launch_pending.go:29-35`). AC-CC-001 이 이 값 하나를 더 돈다.
3. 비대화형 stdin(EOF): 관리 카드 세션은 `/exit`·`/quit`·치명 오류 전에는 끝나지 않는다(D-5). 문서에 적는다; 이 SPEC 은 바꾸지 않는다(열린 질문 Q2).
4. 운영자가 `/exit` 를 연달아 두 번: 두 번째가 다음 카드 세션을 즉시 끝낼 수 있다(D-6 알려진 비대칭).
5. 소유자 `Start` 실패(바이너리에 `app-server` 가 없음, 핸드셰이크 10초 시간 초과 등): launch-pending 행이 남지 않고(AC-CC-007) 루프는 다음 카드를 시도한다. 연속 실패의 상한은 이 SPEC 이 만들지 않는다(열린 질문 Q3).
6. Windows: 가짜 서버가 POSIX 셸 shim 이라 AC-CC-007·008·010(루프 수준)은 skip 하고 크로스 빌드(AC-CC-011)가 증거다(부모 SPEC 의 선례).

## §4. 품질 게이트와 DoD

- **TRUST 5**: Tested — 위 AC 전부, 변이 8개, 소유자 스위트 회귀 / Readable — 기존 파일의 스타일(영어 주석, 줄 폭) / Unified — `gofmt`·golangci-lint v2.1.6 / Secured — 새 환경 변수·네트워크 면 없음, 소유자의 승인 스코핑 불변 / Trackable — 모든 커밋에 `t1440`, `Authored-By-Agent` 트레일러.
- **DoD**: (1) AC-CC-001..013 의 명령을 이 트리에서 실제로 돌려 원문 출력을 progress.md §E.2 에 인용, (2) 변이 mu1..mu8 중 적용 가능한 것을 실제로 적용해 붉어짐을 관측, (3) 기본 꺼짐 AC-CC-002·003 이 변경 전후 모두 초록, (4) `-list` 선택 수 확인, (5) 문서·CHANGELOG 는 sync 단계(manager-docs).
