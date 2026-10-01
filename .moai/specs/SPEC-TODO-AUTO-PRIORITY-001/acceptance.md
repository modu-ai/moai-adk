# SPEC-TODO-AUTO-PRIORITY-001 — 인수조건

> 각 항목은 Given-When-Then 이며 이진 판정 가능하다. 모든 AC 는 머리에 **Covers** 로
> 자기가 덮는 `REQ-TAP-XXX` 를 인용한다. 판정은 명령의 실제 실행 출력으로 한다.
> 테스트 이름은 run 페이즈가 `internal/cli` 에 두는 Go 테스트이고, 문서 항목은 파일
> 본문 grep 과 문서 테스트로 판정한다. 총 15개 (Tier M 상한 16 이내).
>
> **명령 규약.** RED-now 셀은 증거 대장의 단일 호출 평문 명령이다 — 파이프·리다이렉션·`&&`·
> `;` 체인·서브셸이 없다(verification-completeness §2.1). green 경로 셀의 Go 테스트 명령은
> 앞에 환경 정리 접두 `<SCRUB>` 를 붙인다. 이 접두는 plan.md §C 의 11개 변수를 한 번의 복합
> 호출 안에서 푼다(`unset … && go test …`, AGENTS.md §4). 접두가 `&&` 를 담으므로 §2.1 이 말하는
> RED-now 셀의 단일 호출 형태 밖이다 — 그래서 접두를 가진 대장 행(E2, E7)은 정보용이고, 어느
> AC 의 RED-now 도 그 행 하나에만 기대지 않는다. `-run` 패턴은 전부 앵커(`^…$`)를 쓴다 —
> 앵커 없는 패턴은 더 긴 이름까지 고르기 때문이다.
>
> **분류.** 각 AC 는 분류 줄을 가진다. **release-blocking** 은 단일 호출 형태의 RED-now 대장 행이
> 있고 트리 `7d8a9bdbc` 에서 재실행으로 재현되는 기준이다(14개: AC-001~011, 013~015). **regression-guard**
> 는 RED-now 가 없고 도착 시점부터 GREEN 이어야 하는 기준이다(1개: AC-012).
>
> **RED-now / green-path 규율.** 모든 AC 는 RED-now(현 트리에서 실패하는 이유)와 green 으로
> 뒤집는 마일스톤을 한 쌍으로 적는다. RED-now 측정은 아래 증거 대장에 명령 · 원문 stdout ·
> 종료 코드로 적었고, 대장 전체가 트리 `7d8a9bdbc` 에 고정된다(항목별 핀이 없는 AC 에 문서 단위
> 핀이 적용된다). 신규 행동의 RED-now 는 "그 행동의 코드가 트리에 없다"는 부재 관측이며, 각 이름
> 붙은 테스트 자신의 RED(실패 출력 원문)는 run 의 M1/M2 에서 GREEN 앞에 캡처한다(E8 — 이 문서의
> 대장 행이 아니라 manager-develop 의 자체 검증 항목 E8, RED 실패 출력). 대장 행 번호는 그와
> 겹치지 않게 E9 부터 새로 붙였다.
>
> **green 마일스톤은 plan.md 의 M1/M2 와 같은 방향으로 맞췄다.** 출력(사이클)을 읽는 Then 은
> `runAutoCycle` 배선이 들어오는 M2 에서만 뒤집힌다. 스테이지 함수를 직접 부르는 Then(AC-004~006)
> 은 M1 에서 뒤집힌다.

## 증거 대장 (RED-now 관측, 트리 `7d8a9bdbc`)

| id | 명령 | 원문 stdout | 종료 코드 | 의미 |
|---|---|---|---|---|
| E1 | `grep -rl --exclude="*_test.go" "selection: source=" internal/cli` | (출력 없음) | 1 | 비테스트 소스 중 선택 기록을 찍는 코드가 없다. plan D-1 이 그 코드를 `todo_auto_rank.go` 에 고정하므로 green 에서는 stdout 이 `internal/cli/todo_auto_rank.go` 이고 종료 0 이다. 테스트 파일은 제외했다 — 테스트 파일만 추가해도 뒤집히는 변이를 막는다 |
| E1c | `grep -rl --exclude="*_test.go" "jev: unavailable" internal/cli` | `internal/cli/todo_auto.go` | 0 | E1 의 양성 대조. 같은 형태가 이 디렉터리의 비테스트 소스에서 실제로 적중을 낸다 — E1 의 빈 출력이 "스캔이 비었다"가 아님을 보인다 |
| E2 | `<SCRUB> go test -list 'TestAutoRank' ./internal/cli/` | `ok  	github.com/modu-ai/moai-adk/internal/cli	0.859s` (테스트 이름 0개) | 0 | 정보용(접두가 `&&` 를 담는다). 신규 테스트가 없다. 이 `ok` 는 빈 sweep 이므로 초록의 증거가 되지 못한다 |
| E3 | `grep -c "auto-scoped ranking exception" .claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/gtd.md` | `.claude/rules/moai/workflow/kanban-dispatch.md:0` · `.claude/skills/moai/workflows/gtd.md:0` | 1 | 두 정본 문서에 예외 문구가 없다 |
| E4 | `grep -n "\[보류" .claude/skills/moai/workflows/gtd.md` | (출력 없음) | 1 | 마커 한계 공개가 없다 |
| E5 | `grep -c "auto-scoped ranking exception" .claude/agents/moai/manager-todo.md` | `0` | 1 | 에이전트 문서에 예외 문구가 없다 |
| E6 | `grep -n "Mutate\|ArchiveCard" internal/cli/todo_auto_rank.go` | (stdout 없음; stderr `grep: internal/cli/todo_auto_rank.go: No such file or directory`) | 2 | 순위 파일 자체가 없다 |
| E7 | AC-TAP-012 의 명령과 동일(`<SCRUB> go test -count=1 -v -run '^(TestTodoAutoPickupSelection\|TestAutoPickTargetsRelationBlocked\|TestAutoPickTargetsReturnsAfterDone\|TestRunAutoCycleSkipsBlockedCards\|TestAutoPickTargetsRescueArmUnfiltered\|TestAutoPickTargetsNonSequencingRelation\|TestTodoAutoSerialCycle\|TestTodoAutoJevPoisonedValueCausesNoMutation\|TestTodoAutoJevDegradedNonFinding\|TestTodoAutoJevScriptPresentSignal)$' ./internal/cli/`) | 10개 이름 전부 `--- PASS`, 마지막 줄 `ok  	github.com/modu-ai/moai-adk/internal/cli	5.760s`, `--- FAIL` 없음 | 0 | 정보용(접두가 `&&` 를 담는다). 회귀 가드 AC-012 의 기준선. 이름 10개가 실제로 실행됐다(sweep 이 비지 않았다) |
| E9 | `grep -c "auto-scoped ranking exception" internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md:0` · `internal/template/templates/.claude/skills/moai/workflows/gtd.md:0` | 1 | 두 template 사본에도 예외 문구가 없다 |
| E10 | `grep -c "auto-scoped ranking exception" internal/template/templates/.claude/agents/moai/manager-todo.md` | `0` | 1 | manager-todo 의 template 사본에도 예외 문구가 없다 |
| E11 | `grep -c "selection order only" .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-todo.md` | `.claude/rules/moai/workflow/kanban-dispatch.md:0` · `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md:0` · `.claude/skills/moai/workflows/gtd.md:0` · `internal/template/templates/.claude/skills/moai/workflows/gtd.md:0` · `.claude/agents/moai/manager-todo.md:0` · `internal/template/templates/.claude/agents/moai/manager-todo.md:0` | 1 | 범위 한정 문구(spec §B.5)가 어느 사본에도 없다 |

E1·E1c·E9·E10·E11 은 이 개정판 작성 중 같은 트리에서 새로 측정했고, E3·E4·E5·E6 은 같은 트리에서 다시
실행해 같은 출력과 종료 코드를 얻었다. 종료 코드는 측정용 래퍼(`sh -c '…; echo exit=$?'`)로 읽었으며,
위 명령 칸의 명령 자체는 래퍼 없는 평문이다.

## AC-TAP-001 — 선택 기록이 첫 accept 앞에 나오고 accept 순서가 ranked 순서다

**Covers**: maps REQ-TAP-001, REQ-TAP-008

- **분류**: release-blocking (RED-now: E1, 양성 대조 E1c).
- **Given** queued 카드 3장(우선순위 `normal`·`high`·`normal`)이 있고 Jev 가 꺼져 있을 때,
  **When** `--auto` 주기를 돌리면,
  **Then** 출력에서 첫 `selection: source=` 줄이 첫 `accept ` 줄보다 앞에 있고,
  `selection: ranked` 줄의 id 순서가 이후 `accept ` 줄의 id 순서와 같다.
- **RED-now**: E1 — 기록을 찍는 비테스트 소스가 없다.
- **green** (M2): `<SCRUB> go test -count=1 -v -run '^TestAutoRankSelectionRecord$' ./internal/cli/` →
  출력에 `--- PASS: TestAutoRankSelectionRecord ` (이름 뒤 공백), 종료 0. 그리고 E1 이 stdout
  `internal/cli/todo_auto_rank.go`, 종료 0 으로 뒤집힌다.
- **변이 시험**: 기록을 첫 accept 뒤에 찍거나 ranked 줄과 다른 순서로 accept 하는 변이가 이
  단언을 실패시킨다(출력 줄 위치 비교).

## AC-TAP-002 — Jev 가 가용하면 Jev 순서를 따르고 source=jev 로 기록한다

**Covers**: maps REQ-TAP-002

- **분류**: release-blocking (RED-now: E1).
- **Given** Jev 시임이 `Available` 과 후보마다 점수(0~4)와 신뢰도를 돌려주되 그 순서가 폴백 순서와
  **다르게** 구성돼 있을 때, **When** 주기를 돌리면, **Then** `selection: source=jev` 가 찍히고
  accept 순서가 점수 내림차순이다. 점수가 같은 후보는 신뢰도 내림차순, 그것도 같으면 폴백 순서다.
  후보가 요청 상한을 넘는 케이스에서는 초과분이 Jev 순서 뒤에 폴백 순서로 붙고 `selection: note`
  줄이 그 사실을 적는다.
- **RED-now**: E1.
- **green** (M2): `<SCRUB> go test -count=1 -v -run '^TestAutoRankJevOrdering$' ./internal/cli/` →
  `--- PASS: TestAutoRankJevOrdering `, 종료 0 (점수 동률 → 신뢰도 → 폴백 순서의 3단 타이브레이크를
  각각 가르는 서브테스트 포함).
- **변이 시험**: Jev 답을 무시하고 폴백으로 정렬하는 변이는 "Jev 순서 ≠ 폴백 순서" 구성 때문에
  실패한다. 점수 오름차순으로 정렬하는 변이와 신뢰도 타이브레이크를 빼는 변이도 각 서브테스트에서
  실패한다. 이 구성이 빠진 테스트는 변이를 통과시키므로 채택하지 않는다.

## AC-TAP-003 — Jev 불가 사유는 폐쇄 어휘로 한 번 찍히고 종료 코드는 0이다

**Covers**: maps REQ-TAP-003, REQ-TAP-007

- **분류**: release-blocking (RED-now: E1).
- **Given** Jev 시임이 `Disabled`·`NoCredential`·`Unauthorized`·`RateLimited`·`Overloaded`·
  `Unreachable`·`Oversize`·`SecretDetected`·`Malformed` 를 차례로 돌려주고, 별도로 `Available` 이면서
  답이 후보 일부를 비우는 경우가 있을 때, **When** 각각 주기를 돌리면, **Then** 케이스마다
  `selection: source=fallback` 줄에 `reason=jev-disabled`·`jev-no-credential`·`jev-unauthorized`·
  `jev-rate-limited`·`jev-overloaded`·`jev-unreachable`·`jev-oversize`·`jev-secret-detected`·
  `jev-malformed`·`jev-incomplete-answer` 가 정확히 하나씩 대응해 한 번만 나오고, 주기는 오류 없이
  끝난다. 대응은 이렇게 갈린다: 시임 결과의 가용성이 `Malformed`(질문 없는 요청 — 스테이지가 빈
  질문 목록을 보내지 않으므로 시임으로만 닿는다)이면 `jev-malformed`, `Available` 인데 답 집합이
  REQ-TAP-011 검증에 실패하면 `jev-incomplete-answer`, 응답을 읽지 못한 경우는 `Unreachable` 이므로
  `jev-unreachable` 이다.
- **RED-now**: E1.
- **green** (M2): `<SCRUB> go test -count=1 -v -run '^TestAutoRankFallbackReasons$' ./internal/cli/` →
  `--- PASS: TestAutoRankFallbackReasons ` (테이블 케이스 10개 서브테스트).
- **변이 시험**: 사유를 생략하는(조용한 폴백) 변이는 `reason=` 존재 단언에서, 사유를 두 번 찍는
  변이는 개수 단언에서, `Malformed` 와 불완전 답을 같은 사유로 합치는 변이는 두 서브테스트의
  사유 비교에서 실패한다.

## AC-TAP-004 — 폴백은 우선순위 high > normal > low 순이고 같은 우선순위에서는 큐 순서를 지킨다

**Covers**: maps REQ-TAP-004

- **분류**: release-blocking (RED-now: E1).
- **Given** 큐 순서 `A(normal) B(high) C(low) D(normal) E(high)` 이고 모두 readiness 가 깨끗할 때,
  **When** 폴백 순위 단계(스테이지 함수, 시임 주입)를 적용하면, **Then** ranked 는 `B E A D C` 다.
  분류가 없는 카드는 `normal` 로 읽힌다. 이 Then 은 스테이지 반환값을 읽으므로 사이클 배선 없이
  M1 에서 판정된다.
- **RED-now**: E1 — 순위 단계 자체가 없어 현재 순서는 저장된 큐 순서 `A B C D E` 다.
- **green** (M1): `<SCRUB> go test -count=1 -v -run '^TestAutoRankFallbackOrder$' ./internal/cli/` →
  `--- PASS: TestAutoRankFallbackOrder `. 우선순위 비교는 `kanban` 의 노출 접근자를 쓴다 —
  `<SCRUB> go test -count=1 -v -run '^TestPriorityRank$' ./internal/kanban/` → `--- PASS: TestPriorityRank `.
- **변이 시험**: 큐 순서만 쓰는 변이, 오름차순 변이, 같은 우선순위 안에서 순서를 섞는 변이가
  모두 위 단일 입력에서 서로 다른 출력을 내 실패한다.

## AC-TAP-005 — readiness 불량 세 신호는 각각 후보를 뒤로 보내되 버리지 않는다

**Covers**: maps REQ-TAP-004, REQ-TAP-005

- **분류**: release-blocking (RED-now: E1).
- **Given** 후보 4장: `H`(본문이 `[보류` 로 시작), `L`(PR 상태 `landed`), `N`(near-duplicate 발견
  기록이 이름을 댄다), `C`(깨끗, 우선순위 `low`) 가 있고 앞의 셋은 우선순위가 `high` 일 때,
  **When** 폴백 순위 단계(스테이지 함수와 기록 렌더러, 시임 주입)를 적용하면, **Then** `C` 가
  먼저, 불량 셋이 뒤에 큐 순서대로 오고, 불량 셋은 `selection: flagged` 줄에 신호명(`hold-marker`·
  `landed`·`near-duplicate`)과 함께 렌더되며 모두 대상 목록에 남아 있다. 신호가 둘인 카드는 한 줄에
  두 신호명이 같이 찍힌다.
- **RED-now**: E1.
- **green** (M1): `<SCRUB> go test -count=1 -v -run '^TestAutoRankDemotion$' ./internal/cli/` →
  `--- PASS: TestAutoRankDemotion ` (신호별 서브테스트 3개 + 복합 1개).
- **변이 시험**: 세 신호 중 하나만 구현한 변이는 그 신호의 서브테스트가 실패한다. 강등 대신
  제외하는 변이는 "대상 목록에 남아 있다" 단언에서 실패한다.

## AC-TAP-006 — 측정하지 못한 신호는 불량으로 읽지 않고 note 로 밝힌다

**Covers**: maps REQ-TAP-005

- **분류**: release-blocking (RED-now: E1).
- **Given** landed 조회가 `unknown` 을 돌려주거나 실패하도록 시임을 구성했을 때, **When** 폴백 순위
  단계(스테이지 함수와 기록 렌더러)를 적용하면, **Then** 어떤 카드도 `landed` 로 강등되지 않고
  `selection: note` 줄이 미측정 신호를 명시한다. 본문 중간에만 `[보류` 가 있는 카드나 `[보류` 앞에
  다른 문자가 먼저 오는 카드는 강등되지 않는다.
- **RED-now**: E1.
- **green** (M1): `<SCRUB> go test -count=1 -v -run '^TestAutoRankUnmeasuredSignal$' ./internal/cli/` →
  `--- PASS: TestAutoRankUnmeasuredSignal `.
- **변이 시험**: `unknown` 을 `landed` 로 취급하는 변이, 본문 어디서나 마커를 찾는 변이가 실패한다.

## AC-TAP-007 — blocked 카드는 두 소스 모두에서 제외되고 상태는 그대로다

**Covers**: maps REQ-TAP-006

- **분류**: release-blocking (RED-now: E1).
- **Given** 분류가 `blocked=true` 인 queued 카드가 있을 때, **When** 폴백으로, 그리고 별도로
  Jev 가 그 카드를 1순위로 매기는 시임으로 주기를 돌리면, **Then** 두 경우 모두 그 카드는 accept
  되지 않고 `selection: excluded <id> (blocked)` 줄이 한 번 찍히며 큐에서의 카드 상태는 `queued`
  그대로다. 후보가 전부 제외되면 `ranked` 줄은 없고 기존 no-eligible 문구로 주기가 끝난다.
- **RED-now**: E1 — 현재 `autoPickTargets` 는 `blocked` 를 보지 않는다(`todo_auto.go:162-187`).
- **green** (M2): `<SCRUB> go test -count=1 -v -run '^TestAutoRankBlockedExcluded$' ./internal/cli/` →
  `--- PASS: TestAutoRankBlockedExcluded `. 폴백 소스 서브테스트는 M1 에서 먼저 초록이 되지만, 이 AC 는
  사이클 출력을 읽는 Jev 소스 서브테스트까지 통과하는 M2 에서만 초록이다.
- **변이 시험**: 폴백에서만 제외하는 변이는 Jev 케이스에서 실패한다.

## AC-TAP-008 — dead-owner picked 카드는 어떤 순위 소스에서도 가장 앞이다

**Covers**: maps REQ-TAP-009

- **분류**: release-blocking (RED-now: E1).
- **Given** 소유자가 죽은 `picked` 카드 `P1 P2` 와 queued 카드들이 있고, Jev 시임이 queued 카드를
  모두 `P1`·`P2` 보다 높게 매기도록 구성돼 있을 때, **When** 주기를 돌리면, **Then** targets 순서는
  `P1 P2` 뒤에 순위가 매겨진 queued 카드들이다. 구조 근거: 순위 단계는 queued 부분에만 적용된다.
- **RED-now**: E1 — 순위 단계가 없다. 현재 순서 보장은 `TestTodoAutoPickupSelection`(`:116`)이
  queue 순서만으로 고정한다.
- **green** (M2): `<SCRUB> go test -count=1 -v -run '^TestAutoRankRescueFirst$' ./internal/cli/` →
  `--- PASS: TestAutoRankRescueFirst `.
- **변이 시험**: 전체 targets 를 순위화하는 변이는 Jev 시임 구성 때문에 `P1 P2` 가 뒤로 밀려 실패한다.

## AC-TAP-009 — 순위 단계는 큐를 쓰지 않는다

**Covers**: maps REQ-TAP-010

- **분류**: release-blocking (RED-now: E6).
- **Given** 순위 결과가 큐 저장 순서와 다르게 나오는 큐와, 증거가 도착하지 않는 주기(데드라인
  1ms)가 있을 때, **When** 주기가 끝나면, **Then** 큐 레코드의 `Items` 의 id 순서·본문·분류가 주기
  전과 같고 모든 카드가 `queued` 로 돌아와 있으며(pick→unpick 외 변화 없음), 호출 중 큐 쓰기는
  `Mutate` 로 들어온 pick/unpick/done 뿐이다. 정적 가드: 순위 파일은 큐 쓰기 호출을 포함하지 않는다.
- **RED-now**: E6 — 순위 파일이 없다(`grep` 종료 2).
- **green** (M2): `<SCRUB> go test -count=1 -v -run '^(TestAutoRankQueueUnchanged|TestAutoRankNoQueueWriteGuard)$' ./internal/cli/`
  → 두 테스트 `--- PASS`. 그리고 `grep -n "Mutate\|ArchiveCard" internal/cli/todo_auto_rank.go` → 출력 없음,
  종료 1; 양성 대조 `grep -c "^func " internal/cli/todo_auto_rank.go` → 1 이상, 종료 0(파일이
  존재하고 스캔됨을 보인다. 파일이 없으면 종료 2 이므로 위 "출력 없음"과 구분된다).
- **변이 시험**: 정렬 결과를 큐에 되쓰는 변이는 `Items` id 순서 비교에서 실패한다.

## AC-TAP-010 — Jev 의 오염되거나 불완전한 답은 부분 적용 없이 폴백으로 간다

**Covers**: maps REQ-TAP-011

- **분류**: release-blocking (RED-now: E1).
- **Given** Jev 시임이 `Available` 이면서 (i) 보내지 않은 카드 id 를 가리키는 답, (ii) 후보 하나가 빠진
  답, (iii) 점수가 유한한 수가 아니거나(NaN, ±Inf) 닫힌 구간 0~4 밖인 답(예: -1, 4.5, 5)을 차례로
  돌려줄 때, **When** 주기를 돌리면, **Then** 세 경우 모두
  `selection: source=fallback reason=jev-incomplete-answer` 가 찍히고 ranked 는 폴백 순서이며,
  Jev 답이 카드 상태·본문·위치나 어떤 큐 동사에도 닿지 않는다(큐 레코드 동일). 양성 대조: 점수가
  구간 경계 0 과 4 인 답, 그리고 구간 안의 비정수(예: 2.5)는 유효로 받아들여져 `source=jev` 가 된다.
- **RED-now**: E1. (기존 `TestTodoAutoJevPoisonedValueCausesNoMutation` 은 표시 줄만 다루며
  순위 소비를 덮지 않는다.)
- **green** (M2): `<SCRUB> go test -count=1 -v -run '^TestAutoRankJevMalformedAnswer$' ./internal/cli/` →
  `--- PASS: TestAutoRankJevMalformedAnswer `. 테스트 이름은 sweep 집합에서 가져온 것이며
  `jev-malformed` 가 아니라 `jev-incomplete-answer` 를 검증한다.
- **변이 시험**: 유효한 답만 골라 적용하는 부분 적용 변이는 (ii) 에서 실패한다. 구간을 1~4 로
  좁히거나 0~5 로 넓히는 변이는 경계 양성 대조(0, 4)나 (iii) 의 5 에서 실패한다.

## AC-TAP-011 — 정본 독트린 두 곳이 `--auto` 한정 예외를 말하고 다른 표면의 금지는 그대로다

**Covers**: maps REQ-TAP-012

- **분류**: release-blocking (RED-now: E3, E9, E11).
- **Given** M3 이후의 `kanban-dispatch.md` 와 `workflows/gtd.md` (각각 live 와 template 사본)가 있을 때,
  **When** 문서 테스트와 grep 으로 읽으면, **Then** 네 파일 모두에 고정 문구
  `auto-scoped ranking exception` 과 `selection order only` 가 같은 문단에 함께 1회 이상 있고,
  `kanban-dispatch.md` 의 승격 조항은 리더의 선택·`gtd next`·분석기에 대한 금지 문장을 그대로
  유지한다.
- **RED-now**: E3 (live 두 파일), E9 (template 두 파일), E11 (범위 한정 문구) — 모두 0.
- **green** (M3): `grep -c "auto-scoped ranking exception" .claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/skills/moai/workflows/gtd.md`
  → 네 줄 모두 `:1` 이상, 종료 0. `grep -c "selection order only" ` 에 같은 네 파일을 주면 마찬가지로
  네 줄 모두 `:1` 이상, 종료 0. 그리고
  `<SCRUB> go test -count=1 -v -run '^TestAutoRankDoctrineAmendment$' ./internal/cli/` →
  `--- PASS: TestAutoRankDoctrineAmendment `(두 문구의 같은 문단 공존과 금지 문장 유지 단언 포함).
- **변이 시험**: 금지 문장 전체를 지우고 예외만 남기는 변이는 "금지 유지" 단언에서 실패한다. 범위 한정
  문구 없이 예외 문구만 쓰는 변이는 같은 문단 공존 단언에서 실패한다.

## AC-TAP-012 — 회귀 가드: 기존 pickup·관계 필터·hold·Jev 표시 테스트는 그대로 초록이다

**Covers**: maps REQ-TAP-001, REQ-TAP-009, REQ-TAP-010 (기존 필터·hold·`jev:` 표시 줄 불변 · rescue 우선 · 큐 불변)

- **분류**: regression-guard (RED-now 없음 — 도착 시점부터 GREEN).
- **Given** M1~M4 가 반영된 트리에서 §Findings (f) 의 특성화 테스트 10개는 수정되지 않았고
  (`todo_relation_filter_test.go` 는 파일 전체가 손대지 않았으며, `todo_auto_test.go` 에는 감사 지적 D-N1
  이 요구한 의도적 변경이 정확히 하나 있다 — Definition of Done 3번), 새 시임은 nil 기본값이 비활성이므로
  — plan.md D-5 — 기존 `autoOptions` 구성이 그대로 돈다. **When**
  `<SCRUB> go test -count=1 -v -run '^(TestTodoAutoPickupSelection|TestAutoPickTargetsRelationBlocked|TestAutoPickTargetsReturnsAfterDone|TestRunAutoCycleSkipsBlockedCards|TestAutoPickTargetsRescueArmUnfiltered|TestAutoPickTargetsNonSequencingRelation|TestTodoAutoSerialCycle|TestTodoAutoJevPoisonedValueCausesNoMutation|TestTodoAutoJevDegradedNonFinding|TestTodoAutoJevScriptPresentSignal|TestTodoAutoEntryPointFlag|TestJevCallPath_HasExactlyTheDeclaredConsumers)$' ./internal/cli/`
  를 돌리면, **Then** 이름 12개(특성화 10개 + 아래 가드 2개)가 모두 `--- PASS` 이고 `--- FAIL` 이
  없으며 종료 0 이다.
- **가드 집합에 더해진 두 이름 (run 단계 발견, AC 를 추가하지 않는다)**: `TestTodoAutoEntryPointFlag`
  (D-N1 에 따라 hermetic 하게 바뀐 진입점 테스트 — 생산 배선이 landed·Jev 두 시임에 실제로 닿는지와
  `selection: source=fallback reason=jev-disabled` 기록을 단언한다)와
  `TestJevCallPath_HasExactlyTheDeclaredConsumers`(`internal/jev` 를 import 하는 소비자 집합 가드 —
  M1 트리에서 새 소비자 `todo_auto_rank.go` 가 선언되지 않아 M2 시작 시점에 적색이었고, 소비자를
  선언해 초록이 됐다).
- **RED-now**: 해당 없음 — 회귀 가드. 도착 시점부터 GREEN 이어야 하며 E7 이 그 기준선이다
  (같은 명령의 10개 이름 판을 `7d8a9bdbc` 에서 실행해 10개 PASS·종료 0 을 관측했다; E7 은 정보용 행이며
  덧붙인 두 이름은 포함하지 않는다).
- **green 이유 확인(변이)**: run M1 에서 관계 필터를 임시로 제거하면
  `TestAutoPickTargetsRelationBlocked` 가 적색이 되고 복원하면 초록이 됨을 한 번 관측해 가드가
  실제로 무는지 확인한다(plan 시점에는 실행하지 않았다 — plan.md G-5).

## AC-TAP-013 — 미러가 중립이고 개정 구절이 live 와 일치한다

**Covers**: maps REQ-TAP-012

- **분류**: release-blocking (RED-now: E3, E9).
- **Given** 두 template 사본(`gtd.md`, `kanban-dispatch.md`)이 있을 때, **When** 중립성·패리티 테스트를
  돌리면, **Then** template `gtd.md` 에 SPEC id·REQ 토큰·ISO 날짜·9자 이상 16진 연속이 없고,
  개정된 `--auto` 구절이 live 와 template 에서 일치하며, `kanban-dispatch.md` 는 토큰 정규화 후
  구조 드리프트가 없다.
- **RED-now**: E3 + E9 — 개정 구절이 live 에도 template 에도 없으므로 패리티 단언이 비교할 대상이 없다.
- **green** (M3·M4): `<SCRUB> go test -count=1 -v -run '^(TestAutoRankMirrorParity|TestTodoSkillDocumentsClassification)$' ./internal/cli/`
  → 두 이름 `--- PASS`, 그리고 `<SCRUB> go test -count=1 -run '^TestSanitizedPairParity$' ./internal/template/` → `ok`.
- **변이 시험**: template 사본에 SPEC id 를 한 번 넣으면 기존 중립성 정규식이 적색이 된다(기존 가드의
  동작이며, 이 AC 의 양성 대조다).

## AC-TAP-014 — 마커 한계가 gtd.md 와 `--auto` 도움말에 공개된다

**Covers**: maps REQ-TAP-013

- **분류**: release-blocking (RED-now: E4).
- **Given** M3 이후의 `workflows/gtd.md`(live·template)와 `--auto` 플래그 도움말이 있을 때,
  **When** 문서 테스트가 읽으면, **Then** 세 곳 모두 `[보류` 로 시작하는 카드만 강등된다는 것, 마커
  없는 산문 보류는 강등되지 않는다는 것, 구조적 보류는 `moai todo hold` 라는 것을 적는다.
- **RED-now**: E4.
- **green** (M3): `<SCRUB> go test -count=1 -v -run '^TestAutoRankMarkerDisclosure$' ./internal/cli/` →
  `--- PASS: TestAutoRankMarkerDisclosure `.
- **변이 시험**: 도움말에서만 지우거나 문서에서만 지우는 변이가 각각 실패한다(세 곳을 따로 단언).

## AC-TAP-015 — manager-todo 에이전트 문서가 같은 예외를 말한다

**Covers**: maps REQ-TAP-014

- **분류**: release-blocking (RED-now: E5, E10, E11).
- **Given** M4 이후의 `manager-todo.md`(live·template)가 있을 때, **When** grep 과 테스트로 읽으면,
  **Then** 두 사본 모두 serial-cycle 계약과 Jev 결정 경계 절에 `auto-scoped ranking exception` 과
  `selection order only` 가 같은 문단에 함께 있고 "strict queue order 만 소비한다"는 단정이 더는
  남지 않는다.
- **RED-now**: E5 (live), E10 (template), E11 (범위 한정 문구) — 모두 0.
- **green** (M4): `grep -c "auto-scoped ranking exception" .claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-todo.md`
  → 두 줄 `:2` 이상, 종료 0. `grep -c "selection order only" ` 에 같은 두 파일을 주면 마찬가지로 두 줄
  `:2` 이상, 종료 0. 그리고
  `<SCRUB> go test -count=1 -v -run '^TestAutoRankAgentDoctrine$' ./internal/cli/` →
  `--- PASS: TestAutoRankAgentDoctrine `.
- **변이 시험**: 한 사본에만 반영하는 변이는 두 사본 단언에서 실패한다. 두 절 중 한 곳에만 쓰는 변이는
  `:2` 이상 단언에서 실패한다.

## Sweep 대조 (모든 `TestAutoRank*` green 의 선행 조건)

green 판정 전에 swept 집합을 센다. M4 이후
`<SCRUB> go test -list 'TestAutoRank' ./internal/cli/` 는 정확히 다음 15개 이름과 `ok` 줄을 낸다:
`TestAutoRankSelectionRecord`, `TestAutoRankJevOrdering`, `TestAutoRankFallbackReasons`,
`TestAutoRankFallbackOrder`, `TestAutoRankDemotion`, `TestAutoRankUnmeasuredSignal`,
`TestAutoRankBlockedExcluded`, `TestAutoRankRescueFirst`, `TestAutoRankQueueUnchanged`,
`TestAutoRankNoQueueWriteGuard`, `TestAutoRankJevMalformedAnswer`, `TestAutoRankDoctrineAmendment`,
`TestAutoRankMirrorParity`, `TestAutoRankMarkerDisclosure`, `TestAutoRankAgentDoctrine`.
RED-now 는 E2 (이름 0개). 그 뒤 위 15개 이름을 앵커 alternation `^(…|…)$` 으로 묶은
`-run` 한 번의 출력(파일로 돌린 뒤 `grep -c "^--- PASS: "` 로 센다 — 서브테스트 줄은 들여쓰기가
있어 세지 않는다)에서 최상위 PASS 줄이 15개이고 `--- FAIL` 이 0개여야 한다.

## 경계 사례

- 순위 대상(관계 필터를 통과한 queued)이 비어 있으면 `selection:` 줄이 없고 Jev 호출도 없다.
  rescue 대상만 있어도 같다.
- 후보 전부가 blocked 이면 `excluded` 줄들과 기존 no-eligible 문구만 나오고 `ranked` 줄은 없다.
- 후보가 요청 상한(이름 붙은 상수)을 넘으면 초과분은 Jev 순서 뒤에 폴백 순서로 붙고 note 가
  이를 밝힌다. Jev 호출은 큐 잠금 밖에서 일어난다.
- near-duplicate 쌍은 양쪽 카드가 모두 flagged 이며 둘의 상대 순서는 유지된다.
- 본문 앞 공백 뒤에 `[보류` 가 오면 강등되고, 본문 중간의 `[보류` 는 강등되지 않는다.
- Jev 시임이 꺼진 기본 설치(`workflow.jev.enabled: false`)에서는 요청이 만들어지지 않고 사유는
  `jev-disabled` 다.
- 새 시임이 nil 인 `autoOptions`(기존 10개 테스트의 구성)는 Jev 요청도 `gh`·git 도 만들지 않는다 —
  Jev 는 `jev-disabled`, landed 신호는 미측정 note 다(plan.md D-5). 이 기본값 위에서 AC-012 의
  10개 테스트가 돈다.

## 품질 게이트

- 변경 범위 테스트(plan.md §F)가 모두 통과하고, `go vet ./internal/cli/ ./internal/kanban/` 가 종료 0 이다.
- `GOOS=windows GOARCH=amd64 go build ./...` 가 종료 0 이다.
- `internal/jev` 의 표준 라이브러리 전용 import 가드(`TestPackageImports_AreStandardLibraryOnly`)가 통과한다.
- 새 파일의 상수는 이름 붙은 상수이고 인라인 리터럴이 없다.
- `AskUserQuestion` 을 CLI 코드에서 호출하지 않는다(정적 가드).

## Definition of Done

1. AC-TAP-001~015 가 위 green 명령의 실제 출력으로 PASS 이고, Sweep 대조의 15개 이름이 모두 실행됐다.
2. 각 행동 AC 의 테스트가 GREEN 앞에 RED(실패 출력 원문)를 남겼다(E8).
3. AC-TAP-012 의 회귀 가드가 변이 시험으로 실제로 무는 것을 한 번 관측했다. §Findings (f) 의 특성화
   테스트 10개는 수정되지 않았다. `todo_auto_test.go` 에는 감사 지적 D-N1 이 요구한 의도적 변경이
   정확히 하나 있다 — `TestTodoAutoEntryPointFlag` 안의 hermetic 시임 교체와 호출 횟수 단언, 그리고
   `internal/jev` import 한 줄. `todo_relation_filter_test.go` 는 손대지 않았다(run 의 merge-base 대비
   `git diff --stat` 에 나타나지 않는다).
4. 두 정본 문서와 template 사본, `manager-todo.md` 사본이 일치하고 template 사본이 중립이며, 여섯 파일
   모두 두 고정 문구를 같은 문단에 담는다.
5. 완료 보고에 이 SPEC 이 Jev 측 표면의 연동 개정과 Jev 순서 정확도를 주장하지 않는다는 문장이 있고,
   후속 카드가 plan.md §B 의 "Deferred follow-up" 목록 그대로 발행됐거나 발행 요청이 리더에게
   전달됐다. 두 운영자 결정(2026-10-02)은 열린 항목이 아니라 확정이다.
6. 큐 저장소와 라이브 큐는 이 카드의 어느 단계에서도 변경되지 않았다.
