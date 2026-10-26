# acceptance.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

## §0 — 판정 형식과 RED-now 셀

각 기준은 두 셀을 한 쌍으로 채택한다. **RED-now 셀** — 본 트리(base `a158b4b5f`, plan-phase 2026-10-06)에서 실측한 결함 상태의 관측(command + verbatim output + exit code + tree SHA); 새 테스트 이름은 구현 전에 존재하지 않으므로 RED 근거는 **존재 관측**(부재 probe 또는 현재 동작 판독)이고, `[no tests to run]` exit 0은 결코 통과가 아니다(verification-completeness §1.1 — empty-sweep). **green path 셀** — 어느 milestone이 뒤집는지와 통과 시 출력 형태. green 판정 명령은 빈 스윕이면 실패로 읽는다: `go test -run`이 `[no tests to run]`을 출력하면 그 기준은 PASS가 아니라 **미측정**이다. 세 부분(WHEN=소속 milestone 패키지 스위트, INPUT=red를 만드는 fixture, 도달성=go test exit code)은 각 행에 명시한다.

RED-now probe 일괄 관측 기록(모두 `a158b4b5f` 트리, 2026-10-06 plan-phase):

| probe | command | verbatim 결과 | exit |
|---|---|---|---|
| P1 receipt 부재 | `grep -rn 'Receipt\|receipt' internal/cli/todo_autodone.go internal/cli/todo.go` | `internal/cli/todo.go:818:	// REQ-TSS-001 family) — the response's issued id is a receipt for the` + `internal/cli/todo.go:873:	// the issued id is a receipt for the store that answered. Scoped to the` | 0 |
| P2 facts 필드 | `grep -n 'type AutoDoneFacts' -A 12 internal/factory/*.go` | `internal/factory/autodone_scan.go:108:type AutoDoneFacts struct {` 이하 RecordedSHA·SHAReachable·SubjectHit·SubjectKnown… — receipt 필드 없음 | 0 |
| P3 active-only | `grep -n 'classifyRuns(ctx, opts, ' internal/homestate/factory_run_retire.go` | :162 `""`(전체), :333·:350 `"active"` (HEAD `c4784db4c2` 위치: :169·:340·:369 — research §2.3) | 0 |
| P4 reassign 부재 | `grep -in 'reassign' internal/homestate/*.go` | (출력 없음) | 1 |
| P5 관계 disposition 부재 | `grep -n 'disposition' internal/factory/backlog_gtd_schema.go internal/factory/gtd_relation.go` | backlog_gtd_schema.go:36만(gtd_items) — gtd_relations·gtd_relation.go 0건 | 0 |
| P6 watchdog 부재 | `grep -rn 'watchdog\|waiting_since\|review_deadline\|WaitingSince' internal/homestate --include='*.go' \| grep -v _test` | (출력 없음 — 두 번째 grep이 선택한 행이 없어 exit 1) | 1 |
| P7 reserved edge | `grep -n 'CardPushed\|CardCIGreen\|reserved' internal/homestate/card_transition.go` | :173 술어, :267 거부 | 0 |
| P8 발급 출력 | `sed -n '851p' internal/cli/todo.go` | `fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", item.ID, pos)` | 0 |

### §0.1 — M2-B 증거 장부 (카드 t1963, 2026-10-26)

M2-B 기준(AC-FCR-017..025)의 RED-now 셀은 아래 장부 항목을 id로 인용한다. **트리 SHA는 항목마다 `c4784db4c2`**(브랜치 `WT-run-dead-owner`)이며 위 P1-P8의 `a158b4b5f` 핀은 M2-B 기준에 적용되지 않는다. 모든 명령은 워크트리 루트에서 단일 호출로 실행했다(파이프·연쇄 없음). L1의 오버레이 시험 원본은 `probes/m2b_red_probe_test.go.txt`, 매핑은 `probes/m2b-red-overlay.json`(트리 상대 경로)이라 누구든 같은 명령으로 재실행할 수 있다 — 이 파일들은 run-phase 시험이 아니며 커밋돼도 `go` 도구가 읽지 않는다(`.txt`·`.moai/` 아래).

```text
[L1] RED — 죽은 소유자 run 통합·worker 생존·임대 중복
tree: c4784db4c2
command: go test ./internal/homestate/ -overlay .moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/probes/m2b-red-overlay.json -run '^(TestT1963RedTwoDeadRunsOneLeaderStart|TestT1963RedGhostWorkerKeepsRunAlive|TestT1963RedLeaseLeavesDeadRowNonTerminal)$' -count=1 -v
exit: 1
stdout:
=== RUN   TestT1963RedTwoDeadRunsOneLeaderStart
    zz_m2b_red_probe_test.go:26: RED adopt: adopt="" candidates=2, want one dead-owner run adopted
    zz_m2b_red_probe_test.go:33: RED ambiguity: remaining active runs=2 preserved=2, want 1
--- FAIL: TestT1963RedTwoDeadRunsOneLeaderStart (0.78s)
=== RUN   TestT1963RedGhostWorkerKeepsRunAlive
    zz_m2b_red_probe_test.go:48: RED ghost worker: status="active" preserved=1, want retired (worker pid is dead)
--- FAIL: TestT1963RedGhostWorkerKeepsRunAlive (0.63s)
=== RUN   TestT1963RedLeaseLeavesDeadRowNonTerminal
    zz_m2b_red_probe_test.go:65: RED duplicate: dead-run row state="run" after lease in run-new, want terminal or re-keyed
--- FAIL: TestT1963RedLeaseLeavesDeadRowNonTerminal (0.67s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/homestate	2.466s
FAIL
note: 세 시험 이름이 모두 `=== RUN`으로 나타나므로 스윕이 비지 않았다(3/3 선택).
```

```text
[L2] RED — 리더 진입 재채택이 resume 인자를 요구
tree: c4784db4c2
command: grep -n 'resumeSeen && explicitRun' internal/cli/factory_run_adopt.go
exit: 0
stdout:
165:	return adoptResumedLeaderRun(root, resumeSeen && explicitRun == "" && leaderRunIDWasMinted(leaderLabel, priorRunID), warn)
```

```text
[L3] RED — runs 명령의 플래그는 retire 하나뿐(재결속 표면 없음)
tree: c4784db4c2
command: go test ./internal/cli/ -run '^(TestFactoryRunsRetire.*|TestFactoryRunsHasNoForceFlag)$' -count=1 -v
exit: 0
stdout (원문 — 하위 시험 줄 포함):
=== RUN   TestFactoryRunsRetireRefusesInProgressCards
=== RUN   TestFactoryRunsRetireRefusesInProgressCards/operator
=== RUN   TestFactoryRunsRetireRefusesInProgressCards/lane
--- PASS: TestFactoryRunsRetireRefusesInProgressCards (2.25s)
    --- PASS: TestFactoryRunsRetireRefusesInProgressCards/operator (0.74s)
    --- PASS: TestFactoryRunsRetireRefusesInProgressCards/lane (0.56s)
=== RUN   TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun
=== RUN   TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun/A_retired_run_holding_cards
=== RUN   TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun/B_retired_run,_terminal_cards_only
=== RUN   TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun/C_active_card-free_run
--- PASS: TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun (3.51s)
    --- PASS: TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun/A_retired_run_holding_cards (0.86s)
    --- PASS: TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun/B_retired_run,_terminal_cards_only (0.24s)
    --- PASS: TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun/C_active_card-free_run (0.70s)
=== RUN   TestFactoryRunsHasNoForceFlag
--- PASS: TestFactoryRunsHasNoForceFlag (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	7.026s
reading: TestFactoryRunsHasNoForceFlag는 플래그 집합이 정확히 [retire]임을 단언한다(factory_runs_retire_guard_test.go:234-242) — PASS가 곧 --rebind-to 부재의 관측이다. 같은 실행이 RETIRE-GUARD cli 시험 3건의 보존 기준선(AC-FCR-025)이기도 하다.
```

```text
[L4] RED — 합류 프롬프트는 r 접두만 파싱
tree: c4784db4c2
command: grep -n 'strings.HasPrefix(entry, "r")' internal/cli/factory_run_select.go
exit: 0
stdout:
207:	cleanup := strings.HasPrefix(entry, "r")
```

```text
[L5] 보존 기준선 — homestate RETIRE-GUARD 5건 + t1707 재채택·조정 시험
tree: c4784db4c2
command: go test ./internal/homestate/ -run '^(TestRetireRunIfDead.*|TestFindResumableRun.*|TestAdoptRun.*|TestReconcile.*|TestLiveRecordsTerminalSetMatchesSharedDefinition)$' -count=1 -v
exit: 0
stdout (발췌 — 최상위 `--- PASS` 줄과 마지막 줄만 옮김; `=== RUN` 줄과 하위 시험 줄은 생략. 이 항목은 보존 기준선이라 release-blocking RED 셀이 아니다):
--- PASS: TestFindResumableRunAdoptsSingleDeadOwnerWithLiveRecords (0.40s)
--- PASS: TestFindResumableRunSkipsLiveOwner (0.26s)
--- PASS: TestFindResumableRunSkipsRunWithoutLiveRecords (0.25s)
--- PASS: TestFindResumableRunReportsMultipleCandidates (0.24s)
--- PASS: TestFindResumableRunWarnsSingleIndeterminateOwner (0.22s)
--- PASS: TestFindResumableRunWarnsUnstampedPeerClassifiedRun (0.25s)
--- PASS: TestAdoptRunRestampsOwnerAndAppendsResumedEvent (0.31s)
--- PASS: TestAdoptRunRefusesChangedStamp (0.25s)
--- PASS: TestAdoptRunRefusesNonActiveRun (0.23s)
--- PASS: TestLiveRecordsTerminalSetMatchesSharedDefinition (0.00s)
--- PASS: TestReconcilePreservesDeadOwnerRunWithLiveRecords (0.23s)
--- PASS: TestReconcilePreservesDeadOwnerWithOnlyWorkers (0.27s)
--- PASS: TestReconcileRetiresDeadOwnerWithTerminalRecordsOnly (0.28s)
--- PASS: TestReconcileRetiresIdentitylessRunsThatPredateBoot (0.01s)
--- PASS: TestRetireRunIfDeadAcceptsBootProof (0.01s)
--- PASS: TestRetireRunIfDeadRefusesInProgressCards (0.41s)
--- PASS: TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers (0.59s)
--- PASS: TestRetireRunIfDeadCardReadHoldsWriteLock (0.51s)
--- PASS: TestRetireRunIfDeadFlipIsCASOnClassifiedStamp (1.24s)
--- PASS: TestRetireRunIfDeadAlreadyRetiredIsNoOp (1.25s)
--- PASS: TestReconcileRetiresDeadAndLeavesLive (0.30s)
--- PASS: TestReconcileLeavesIndeterminateActive (0.32s)
--- PASS: TestReconcileUnstampedRunUsesLeadPeerFallback (0.33s)
--- PASS: TestReconcileNeverRetiresLiveOrIndeterminate (0.47s)
--- PASS: TestRetireRunIfDeadRefusesLiveAndIndeterminate (0.62s)
ok  	github.com/modu-ai/moai-adk/internal/homestate	9.555s
```

```text
[L6] 보존 기준선 — cli t1707 재채택 시험
tree: c4784db4c2
command: go test ./internal/cli/ -run '^(TestAdoptResumedLeaderRun.*|TestStrictResumeCarrierSeen|TestLeaderRunIDWasMinted)$' -count=1 -v
exit: 0
stdout:
=== RUN   TestAdoptResumedLeaderRunAdoptsSingleCandidate
--- PASS: TestAdoptResumedLeaderRunAdoptsSingleCandidate (0.94s)
=== RUN   TestAdoptResumedLeaderRunSkipsWhenNotResumeShaped
--- PASS: TestAdoptResumedLeaderRunSkipsWhenNotResumeShaped (0.23s)
=== RUN   TestAdoptResumedLeaderRunWarnsOnMultipleCandidates
--- PASS: TestAdoptResumedLeaderRunWarnsOnMultipleCandidates (0.51s)
=== RUN   TestAdoptResumedLeaderRunMintsWhenNoCandidates
--- PASS: TestAdoptResumedLeaderRunMintsWhenNoCandidates (0.23s)
=== RUN   TestAdoptResumedLeaderRunWarnsCodexLedCandidate
--- PASS: TestAdoptResumedLeaderRunWarnsCodexLedCandidate (0.74s)
=== RUN   TestStrictResumeCarrierSeen
--- PASS: TestStrictResumeCarrierSeen (0.00s)
=== RUN   TestLeaderRunIDWasMinted
--- PASS: TestLeaderRunIDWasMinted (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.689s
```

```text
[L7] RED — 산 리더 run 옆의 죽은 run이 채택 대상으로 나온다 (AC-FCR-018 사례 iii)
tree: 41f8db15deae9483161575d811904acc580cd001 (측정 직전 `git rev-parse HEAD` 관측; 소스는 c4784db4c2와 동일 — 감사 판정서 Evidence의 `git diff --name-only c4784db4c2 41f8db15de` = SPEC 디렉터리 파일만). 탐침 원본 `probes/m2b_red_l7_probe_test.go.txt`와 매핑 `probes/m2b-red-l7-overlay.json`은 측정 시점에 작업 트리에만 있었고 이 개정과 함께 커밋된다.
fixture: 죽은 스탬프 run `run-dead`(pid 20) + `OwnerLive` 활성 run `run-live`(pid 21) + `run-dead`의 비종결 카드 `c1(run)` 1장, pid별 분류기(21 → OwnerLive, 그 밖 → OwnerDead)
command: go test ./internal/homestate/ -overlay .moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/probes/m2b-red-l7-overlay.json -run '^TestT1963RedAdoptsDeadRunBesideLiveLeader$' -count=1 -v
exit: 1
stdout:
=== RUN   TestT1963RedAdoptsDeadRunBesideLiveLeader
    zz_m2b_red_l7_probe_test.go:28: RED live-leader guard: adopt="run-dead" candidates=1 while run-live is OwnerLive, want no adoption
--- FAIL: TestT1963RedAdoptsDeadRunBesideLiveLeader (1.06s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/homestate	1.508s
FAIL
reading: 오늘의 후보 탐색(`FindResumableRun`)은 산 리더가 있어도 죽은 run을 채택 대상으로 내놓고, resume 형태 진입은 그것을 채택한다 — REQ-FCR-017의 "산 리더가 있으면 채택하지 않는다"가 현재 거짓이다. `=== RUN` 1줄로 스윕은 1/1.
```

**장부의 성격(L1·L7)**: L1과 L7은 결함 관측이다. 탐침이 단언하는 함수(`FindResumableRun`·`ReconcileActiveRuns`)는 plan·design상 바뀌지 않고 새 선택 단계가 추가되므로, **올바른 구현 뒤에도 L1·L7은 RED로 남을 수 있다** — 녹색 전환의 증인이 아니다. 전환 판정은 각 AC의 green 명령만 한다. L1의 오버레이 원본은 커밋 `41f8db15de`부터 트리에 있다(L1의 측정 핀 `c4784db4c2`와 소스는 같다).

**공통 green 판정 규칙(M2-B)**: 각 green 명령은 단일 호출이고 `-count=1 -v`로 돈다. 통과는 exit 0 **그리고** 출력에 지정 시험 이름의 `--- PASS` 줄이 있는 것이다 — `[no tests to run]` 또는 이름 줄 부재는 미측정이다(§0). 모든 fixture는 샌드박스 루트(`openSandboxFactory`·`adoptSandbox`·`runOwnerSandbox` 형태)에서만 만든다 — 라이브 factory DB를 여는 시험은 기준 위반이다.

## §A — M1: 완료 게이트

### AC-FCR-001 — receipt 없는 done 거부 (REQ-FCR-001/002)

- RED-now: P1 — receipt 게이트 개념 부재(todo.go:1104 `rec.ArchiveCard(id)`가 검증 없이 도달).
- green(M1): `go test ./internal/cli -run '^TestLeaderReceiptGateRejectsMissingReceipt$'` — receipt 없는 done이 거부되고 stderr에 이유, 행은 archive되지 않음. INPUT: fcFixture 루트 + factory-linked 카드. WHEN: M1 이후 cli 스위트.
- 기준: 명령 exit 0이고 `[no tests to run]` 아님.

### AC-FCR-002 — receipt 네 바인딩 불일치 거부 (REQ-FCR-002/005)

- RED-now: P1 동일.
- green(M1): `go test ./internal/cli -run '^TestLeaderReceiptGateBinding$'` — UUID 불일치·**run id 불일치(다른 run의 receipt)**·factory version stale·증거 해시 불일치 각각 거부, 네 값 모두 일치 시 통과. INPUT: 네 바인딩 값을 하나씩 틀어놓은 fixture receipt 4종 — run 변이는 동일 카드·version·증거 SHA를 가진 두 번째 run(측정: cards 스키마에서 생성 가능). **경합 변이(표면 a)**: 수동 done 실행 중 factory 전이로 version이 증가한 행 — archive 직전 재검증이 거부한다(REQ-FCR-002a). **세 표면 커버**: 게이트가 수동 done(todo.go:1104)·`todo --auto`(todo_auto.go:331)·homestate T18/T20 네 곳 모두에 적용됨을 테스트가 확인한다.

### AC-FCR-003 — 수행자 발급 receipt 거부, 리더 발급+리더 실행 허용 (REQ-FCR-005)

- RED-now: P1 동일.
- green(M1): `go test ./internal/cli -run '^TestLeaderReceiptGateSelfIssued$'` — 카드의 수행 owner(lane)가 발급자로 표식된 receipt는 거부; **리더가 발급하고 같은 리더가 done을 실행하는 경로는 통과**(단일 리더 정상 흐름 — REQ-FCR-005의 독립 축은 수행자≠승인자). INPUT: 발급자==수행자 fixture + 발급자==리더 fixture 2종.

### AC-FCR-004 — auto-done이 receipt 없이 닫지 않음 (REQ-FCR-003)

- RED-now: P2 — `AutoDoneFacts`(autodone_scan.go:108)에 receipt 필드 없음.
- green(M1): `go test ./internal/cli -run '^TestAutoDoneReceiptSkip$'` — receipt 미검증 **factory-linked** 후보는 close 대상에서 skip/downgrade, 스캔 전체는 계속. INPUT: 조건은 모두 충족하되 receipt 없는 카드. **경합 변이 포함**: 스캔 승인 뒤 factory 전이로 카드 version이 증가한 행 — archive 직전 네 바인딩 재검증(factory version·run id 포함)이 stale receipt를 거부하고 close를 skip한다(REQ-FCR-004의 직전 재검증). **비팩토리 보존 변이**: factory run 기록이 없는 일반 카드는 receipt 없이도 기존대로 완료된다(적용 대상 아님 — `TestTodoAutoDone_LiveFilter` 계열 동작 불변, REQ-FCR-002의 스코프 문장).

### AC-FCR-005 — 잠금 재검증 전면 비교 (REQ-FCR-004)

- RED-now: 본 트리 판독 — todo_autodone.go:385가 ID 동등만 확인하고 :397 archive까지 진행(스캔-스냅샷과 lock 사이 hold/drop/edit 행도 닫힘).
- green(M1): `go test ./internal/cli -run '^TestAutoDoneRecheckStaleRow$'` — UUID·본문·state·SPEC·landing 중 하나라도 현재값과 다르면 close skip + inconclusive downgrade, 5항목 모두 동일하더라도 **archive 직전에 receipt의 factory version·증거 해시를 lock 안에서 최종 재검증**해 stale receipt는 거부하고, 재검증은 동시 factory 전이와 직렬화된다. INPUT: 스냅샷 후 변형된 행 5종 변이 + 승인 뒤 version 증가 변이.

## §B — M2: 회수

### AC-FCR-006 — reaper가 전체 run 만료 행 회수 (REQ-FCR-006)

- RED-now: P3 — `classifyRuns(ctx, opts, "active")` :333·:350, active 아닌 run의 만료 행은 reconcile 밖.
- green(M2): `go test ./internal/homestate -run '^TestExpiredLeaseReaperAllRuns$'` — active·비활성 run 모두의 만료 임대 행이 transaction 내 재검증 후 assigned 복귀(mid-merge는 blocked), holder·expiry 클리어, owner 유지. INPUT: frLeaseUntil로 과거 만료를 심은 run 2종(active+inactive).

### AC-FCR-007 — reaper 무삭제·fail-closed (REQ-FCR-007)

- RED-now: P3 + retireRun 보존 의미론 판독(:393 UPDATE-only — 이 의미론의 reaper 계승).
- green(M2): `go test ./internal/homestate -run '^TestExpiredLeaseReaperPreservesRows$'` — 회수 후 행 수 불변, 미확실 분류 행은 회수하지 않음, **살아 있는 리더의 run에서 만료 임대도 회수**(owner 생존 조건은 run retirement 전용 — reaper 비상속). INPUT: indeterminate 분류 행 + live-leader run의 만료 행 포함 fixture.

### AC-FCR-008 — 운영자 reassign edge (REQ-FCR-008)

- RED-now: P4 — homestate 전체에 reassign 0건(exit 1).
- green(M2): `go test ./internal/homestate -run '^TestOperatorReassignEdge$'` — 유효 lease 부재 + 기존 owner 종료 증거 확인 시 owner 이동 성공; lease 존재·증거 부재·decider 비운영자·version 충돌 각각 거부. INPUT: frFixtureCard + frPlace 변이 4종.

### AC-FCR-009 — 기존 가드 불변 (REQ-FCR-008의 부정 검증)

- RED-now: P7 인접 판독 — :353 owner 유지·:451 동일 owner 임대 가드 현행.
- green(M2): `go test ./internal/homestate -run '^TestTransitionGuardsUnchanged$'` — 만료 회수는 owner를 유지하고, 임대 획득은 등록 동일 owner만 받는 기존 거부가 그대로 통과(기존 테스트 회귀 없음). INPUT: 기존 fr_transition_test 케이스 재실행.

### AC-FCR-010 — 대기 신호·검토 기한·리더 재판정 (REQ-FCR-009)

- RED-now: P6 — watchdog/waiting_since/review_deadline 개념 0건.
- green(M2): `go test ./internal/homestate -run '^TestStalledWaitWatchdog$'` — owner 있음+임대 부재 행이 기한 초과 시 재판정 요구 기록; 자동 완료·자동 해제 0. **백필 핀(D6)**: 기존 정체 행도 watchdog 첫 관측에서 `waiting_since`가 소급 기록되어 검토 기한이 성숙한다 — 백필 없는 행은 대상화되지 않는다. INPUT: 구동 중 정지 행 fixture + 시간 진행 + **스키마 이전 시점부터 정체 중이던 기존 행**.

### AC-FCR-011 — CI reader가 정확한 SHA만 인정 (REQ-FCR-010/011)

- RED-now: P7 — T19/T20이 :267에서 예약 거부 중("CI verdict reader ... owned by F3").
- green(M2): `go test ./internal/homestate -run '^TestCICompletionReader$'` — push 시 기록된 SHA의 CI 증거는 **T19(pushed→ci-green)을 통과**, 다른 SHA 증거·증거 부재는 T19도 계속 거부; **T20과 T18은 receipt 게이트만 연다**(`FactoryDB.Transition` 안에서) — receipt 없는 T20·T18(원격 없는 저장소 포함)은 계속 거부, 올바른 receipt와 함께 성공(REQ-FCR-002b·016과 모순 없음). **tip 영속화 변이**: 일괄 push 시나리오(카드 MergeSHA≠push tip)에서 reader는 T17이 저장한 tip SHA만 인정하고, push 뒤 remote ref 이동은 인정을 바꾸지 않는다. INPUT: frWriteVerdict 선례(fr_fixture_test.go:108)의 sha 바인딩 변이 + receipt 변이 + 일괄 push tip 변이.

## §B2 — M2-B: run 수명주기 (카드 t1963)

각 기준: Given-When-Then · RED-now(장부 인용과 RED인 이유) · green(뒤집는 milestone·단일 명령·통과 형태) · 변이 탐침(기준을 통과하면서 요구를 어기는 구현을 쓸 수 있는지 — 쓸 수 있으면 그 변이를 잡는 단언을 기준에 넣었다).

### AC-FCR-017 — 리더 진입 한 번에 활성 run이 하나로 모인다 (REQ-FCR-017/018/020/023) [release-blocking]

- **Given** 샌드박스 루트에 죽은 스탬프 run 2개(`run-old`·`run-new`, 생성 시각 `run-old` < `run-new`, 두 스탬프 모두 결코 살아날 수 없는 pid·지문 — `seedDeadRunWithLiveRecords` 형태)가 있고, 비종결 카드가 `run-old`에 `c1(run)`·`c2(picked)`·`c4(needs-decision)`, `run-new`에 `c3(assigned)`·`c4(picked)`로 있으며(`c4`는 두 run에 걸친 중복), 종결 카드 `c5(done)`가 `run-old`에 있고, 어떤 활성 run도 `OwnerLive`·`OwnerIndeterminate`가 아니며, `card_dispatch`가 `c1→run-old`를 가리키며, `leader_approvals`에 `c1`의 `run-old` 승인 행 1개가 있다. 두 run의 `created_at`은 픽스처가 명시적으로 다르게 심는다(같은 시각이면 `run_id` 내림차순으로 `run-old`가 뽑힌다 — 동률 규칙은 변이 탐침 (e)가 따로 단언한다)
- **When** resume 인자 없이, `--factory-run` 없이, 새로 발급된 id로 리더 진입 채택 경로를 한 번 실행한다
- **Then** (1) 활성 run은 정확히 1개이고 그것은 `run-new`이며 스탬프가 현재 프로세스로 바뀌었고 `run.resumed` 이벤트가 1개 붙었다; (2) `run-old`는 `retired`이고 `run.retired` 이벤트 1개가 `rebound_to=run-new`·`moved=2`·`abandoned=1`과 카드 id 목록(옮김 `c1`·`c2`, 포기 `c4`)을 담으며, 리더 진입의 stderr에 같은 카드 id가 나온다; (3) `c1`·`c2`는 `run-new` 아래 행으로 재키됐고 run_id 외 모든 열(state·version·stage·updated_at 포함)이 진입 전 값과 같다; (4) `run-old`의 `c4`는 `abandoned`이고 `card.rebound` 이벤트가 이전 state `needs-decision`과 version을 담으며, `run-new`의 `c4`는 바뀌지 않았다; (5) `cards` 전체 행 수가 진입 전과 **정확히 같고**(삭제 0·복사 0), `events` 행 수는 줄지 않았다; (6) `card_dispatch`의 `c1`은 `run-new`를 가리킨다; (7) `ResolveActiveRun(root, "")`이 `run-new`를 반환하고 오류가 없다 — `AMBIGUOUS_FACTORY` 0회; (8) `c1`의 `run-old` 승인 행이 `run_id=run-old`로 바이트 단위로 그대로이고 `leader_approvals` 행 수도 같다(승인 행을 대상 run으로 옮기지 않는다 — REQ-FCR-023); (9) `run-new` 아래에서 `c1`에 유효한 새 승인을 발급한 뒤, 큐 기록이 `run-old`를 가리키는 상태로 완료 게이트를 부르면 거부되고(완료 0 — 닫힌 쪽 실패, design.md §5), 거부 사유가 바인딩/기록 불일치를 이름으로 든다(승인 부재가 아닌 불일치로 거부됨을 사유 문자열로 구분한다)
- **RED-now**: L1 `TestT1963RedTwoDeadRunsOneLeaderStart` — `adopt="" candidates=2`, `remaining active runs=2`. RED인 이유: 후보가 둘이면 채택하지 않고(`internal/homestate/factory_run_adopt.go:92`), 비종결 카드가 있는 run은 조정이 보존한다(`runHasLiveRecords`). 또한 L2 — 진입 채택이 resume 인자를 요구한다(`:165`). 이 작업이 바꾸지 않는 파일 때문의 RED가 아니다.
- **green (M2-B)**: `go test ./internal/cli -run '^TestLeaderStartConsolidatesDeadOwnerRuns$' -count=1 -v` → exit 0, `--- PASS: TestLeaderStartConsolidatesDeadOwnerRuns `(이름 뒤 공백).
- **변이 탐침**: (a) 옛 run을 카드 이동 없이 은퇴시키는 구현 → (3)·(6) 실패; (b) 옮기지 않고 복사(INSERT)하는 구현 → (5)의 행 수가 늘고 `run-old`가 은퇴 가드에 막혀 (1)·(2) 실패 — 그래서 `cards` 행 수는 "줄지 않았다"가 아니라 "정확히 같다"로 단언한다; (c) 중복 `c4`를 DELETE하는 구현 → (4)·(5) 실패; (d) 최신이 아닌 run을 채택 → (1) 실패; (e) 동률 규칙 — 같은 시험 안의 하위 단언으로, `created_at`이 같은 두 죽은 스탬프 run을 심으면 `run_id` 내림차순으로 큰 쪽이 채택된다(`run_id` 오름차순이나 삽입 순서로 고르는 구현 → 실패); (f) 승인 행의 `run_id`를 대상 run으로 UPDATE하는 구현 → (8) 실패; (g) 큐 기록을 보지 않는 완료 게이트 → (9)에서 완료가 일어나 실패.

### AC-FCR-018 — 자동 채택의 가드와 통합 원천 필터 (REQ-FCR-017/018/020) [release-blocking]

- **Given** 채택 0 사례 여덟: (i) 미확정 소유자 run 하나뿐 (ii) `BasisPeer` run 하나뿐 (iii) 죽은 스탬프 run 하나 + `OwnerLive` 활성 run 하나 (iv) `--factory-run` 명시 (v-a) 죽은 스탬프 run 2개 — **최신이 codex 리더 run**이고 더 오래된 것이 claude run (v-b) 죽은 스탬프 run 2개 — **최신이 산 레거시 피어 기록을 가진 run**이고 더 오래된 것이 claude run (vi) 죽은 스탬프 run 2개 + `OwnerIndeterminate` 활성 run 1개 (vii) (iii)과 같은 fixture에서 `--resume <id>`를 준다. 그리고 통합 원천 필터 사례 하나: (viii) 채택 가능한 죽은 스탬프 claude run 1개(최신) + 더 오래된 죽은 codex 리더 run 1개 + 더 오래된 산 레거시 피어 기록을 가진 죽은 스탬프 run 1개 + 더 오래된 `BasisPeer`로만 dead인 run 1개(각각 비종결 카드 1장)
- **When** (i)~(vi)·(viii)은 resume 인자 없이, (vii)은 `--resume <id>`를 주고 리더 진입 채택 경로를 실행한다
- **Then** (i)~(vii) 모두 채택 0·통합 0·`runs`와 `cards` 덤프 불변이며 새 id가 발급된다. (v-a)(v-b)는 더 오래된 claude 후보로 넘어가지 않는다(채택 0). (i)(ii)(v-a)(v-b)(vi)(vii)은 해당 run id를 이름으로 든 경고 줄이 stderr에 나온다. (viii)은 최신 claude run이 채택되고, codex run·레거시 피어 run·`BasisPeer` run 셋은 통합되지 않고 `active`로 남으며 그 카드 행이 진입 전 그대로이고, stderr에 세 run 각각의 run id와 제외 사유가 한 줄씩 나온다
- **RED-now**: L7 — resume 형태 진입은 산 리더가 있어도 죽은 run을 채택한다(사례 iii RED: `adopt="run-dead" candidates=1 while run-live is OwnerLive`, exit 1, tree `41f8db15de`). (vii)은 같은 L7 결함을 resume 형태 진입 경로에서 직접 겨눈다(L7이 관측한 것이 바로 resume 경로의 채택 후보다). (i)(ii)(iv)(v-a)(v-b)(vi)(viii)은 AC-FCR-017과 짝으로 채택한다: 017이 RED→green으로 뒤집혀야 그 사례들의 green이 자명하지 않다(§2 짝 규칙 — 채택만 하는 기준과 거부만 하는 기준을 함께 둔다).
- **green (M2-B)**: `go test ./internal/cli -run '^TestLeaderStartAutoAdoptGuards$' -count=1 -v` → exit 0, `--- PASS: TestLeaderStartAutoAdoptGuards `(이름 뒤 공백).
- **변이 탐침**: 분류를 보지 않고 최신 활성 run을 채택하는 구현은 AC-FCR-017을 통과하지만 (i)(ii)(iii)(vi)에서 실패한다. 거부된 최신 후보 대신 다음 후보로 넘어가는 구현은 (v-a)(v-b)에서 각각 실패한다 — 한쪽 거부만 구현한 경우도 다른 쪽에서 잡힌다. `OwnerLive`만 보고 미확정 활성 run을 무시하는 구현은 (vi)에서 실패한다. 가드를 무-resume 경로에만 두고 resume 경로는 그대로 둔 구현은 (vii)에서 실패한다. 채택 대상에만 codex·레거시 피어 거부를 적용하고 통합 원천에는 적용하지 않는 구현, 또는 리더 진입 통합에서 `BasisStamp` 조건 없이 `OwnerDead`만 보는 구현은 (viii)에서 실패한다. 경고 없이 조용히 넘기는 구현은 경고 단언에서 실패한다.

### AC-FCR-019 — 두 리더 경합 (REQ-FCR-018/020/023) [release-blocking]

- **Given** 죽은 스탬프 run 2개와 비종결 카드
- **When** (a) 두 채택자가 같은 후보를 분류한 뒤 한쪽이 먼저 `AdoptRun`을 커밋하고 다른 쪽이 이어서 시도한다; (b) 통합 연산의 트랜잭션 진입 직전(분류 뒤)에 대상 run의 스탬프가 다른 프로세스로 바뀐다; (c) 같은 원천을 두 통합 연산이 차례로 시도한다; (d) 통합 연산의 분류 뒤·트랜잭션 진입 전에 **원천** run의 스탬프가 다른 프로세스로 바뀐다(다른 리더가 그 원천을 채택한 상황)
- **Then** (a) 정확히 하나만 채택하고 이긴 쪽의 스탬프는 덮이지 않는다. 진 리더(런처 경로, 분류와 `AdoptRun` 사이 시임에서 스탬프를 교체해 재현)는 **exit 0**으로 새로 발급된 id를 쓰고, 경합 run id를 든 경고 한 줄을 stderr에 남기며, 통합을 0회 실행한다 — 현행 코드의 런치 실패(`re-adopt factory run`)와 다른 새 동작이다. CAS 패배가 아닌 `AdoptRun` 오류(예: 닫힌 데이터베이스)는 여전히 런치를 실패시킨다(같은 cli 시험의 하위 단언 — 오류를 삼키지 않는다); (b) 트랜잭션이 거부되고 카드·`card_dispatch`·`runs` 덤프가 연산 전과 같다; (c) 두 번째 시도는 거부되고 이벤트가 중복되지 않는다(`run.retired` 1개); (d) 트랜잭션이 거부되고 카드·`card_dispatch`·`runs` 덤프가 연산 전과 같다 — 새 소유자의 카드는 옮겨지지 않는다
- **RED-now**: L1 — 통합 연산이 존재하지 않는다(`remaining active runs=2`). 경합 의미론은 새 연산에 처음 적용된다. (a)의 현행 동작(런치 실패)은 `internal/cli/factory_run_adopt.go:106-108`이 `AdoptRun` 오류를 반환하고 `internal/cli/cc.go:227-229`가 `re-adopt factory run: %w`로 실패하는 경로다(코드 판독 — 감사 판정서 D3 인용).
- **green (M2-B)**: 두 명령을 **각각 따로** 실행한다 — (a)는 런처에 닿아야 하므로 cli 시험: `go test ./internal/cli -run '^TestLeaderStartAdoptCASLoserMints$' -count=1 -v` → exit 0, `--- PASS: TestLeaderStartAdoptCASLoserMints `(이름 뒤 공백); (b)(c)(d)는 `go test ./internal/homestate -run '^TestRunConsolidationConcurrentLeaders$' -count=1 -v` → exit 0, `--- PASS: TestRunConsolidationConcurrentLeaders `(이름 뒤 공백).
- **변이 탐침**: WHERE에 스탬프 조건이 없는 UPDATE → (a)에서 둘째가 첫째의 스탬프를 덮어 실패; CAS 오류를 그대로 전파하는 구현 → (a)의 exit 0·새 id 단언에서 실패; 모든 `AdoptRun` 오류를 삼키고 새 id를 발급하는 구현 → (a)의 닫힌 데이터베이스 하위 단언에서 실패; 대상 스탬프를 트랜잭션 밖에서만 확인 → (b)의 시임(분류와 트랜잭션 사이 스탬프 교체)에서 실패; 원천의 `status`만 보고 스탬프는 보지 않는 구현 → (d)에서 실패.

### AC-FCR-020 — 이동 도중 실패는 전부 되돌린다 (REQ-FCR-020) [release-blocking]

- **Given** 원천 run에 비종결 카드 3장, 대상 run은 산 소유자
- **When** 주입점 셋에서 각각 실패로 중단된다(시험 시임 — `retireAfterCardRead`와 같은 형태): (a) 통합 연산이 두 번째 카드를 처리한 직후 (b) 카드 이동을 모두 마친 뒤, `card_dispatch` 재지정 전 (c) 모든 쓰기(카드·바인딩·은퇴·이벤트)를 마친 뒤, 커밋 직전
- **Then** 세 주입점 모두에서 원천은 `active`, 세 카드 모두 원천 아래 원래 열 그대로, 대상의 카드 집합·`card_dispatch`·`events`·`runs`가 연산 전 덤프와 바이트 단위로 같다
- **RED-now**: L1 — 연산 부재. 원자성은 새 연산에 처음 적용된다.
- **green (M2-B)**: `go test ./internal/homestate -run '^TestRebindRetireIsAtomic$' -count=1 -v` → exit 0, `--- PASS: TestRebindRetireIsAtomic `(이름 뒤 공백).
- **변이 탐침**: 카드마다 트랜잭션을 여는 구현은 정상 경로(AC-FCR-017)를 통과하지만 (a)의 덤프 동일성에서 실패한다. 카드 이동+은퇴를 tx1로 커밋하고 `card_dispatch` 재지정·`run.retired` 이벤트를 tx2로 나누는 구현은 (a)를 통과하지만 (b)에서 실패한다. 카드 이동+`card_dispatch` 재지정을 tx1로 커밋하고 은퇴 전환+`run.retired`를 tx2로 나누는 구현은 (a)(b)를 모두 통과하지만 (c)에서 tx1의 변경이 덤프 차이로 드러나 실패한다 — (c)는 어떤 분할이든 먼저 커밋된 부분을 드러낸다.

### AC-FCR-021 — 다른 산 run이 쥔 카드와 죽지 않은 원천은 옮기지 않는다 (REQ-FCR-020/023) [release-blocking]

- **Given** 하위 사례 다섯: (a) 원천의 카드 `cX`가 제3의 활성 run(소유자 `OwnerLive`)에도 비종결 행으로 있다 (b) 같은 상황에서 제3 run 소유자가 `OwnerIndeterminate` (c) 원천에 산 worker(현재 프로세스 pid) 행이 있다 (d) 원천의 소유자가 `OwnerLive` (e) 원천의 소유자가 `OwnerIndeterminate`
- **When** 통합 연산을 실행한다 — (d)(e)는 세 표면 모두에서: `runs --retire <원천> --rebind-to <대상>`, 합류 프롬프트 `m<n>`, 리더 진입 통합
- **Then** 모든 사례·표면에서 연산 전체가 거부되고(부분 이동 없음), 오류가 (a)(b)는 `cX`와 제3 run id를, (c)는 worker label을, (d)(e)는 원천 run id와 그 분류를 이름으로 들며, 덤프가 연산 전과 같다
- **RED-now**: L1 — 연산 부재.
- **green (M2-B)**: `go test ./internal/homestate -run '^TestRebindRefusesCardHeldByAnotherLiveRun$' -count=1 -v` → exit 0, `--- PASS: TestRebindRefusesCardHeldByAnotherLiveRun `(이름 뒤 공백). 세 표면의 (d)(e) 거부는 같은 시험이 연산 진입점별로 단언하며, CLI 표면 쪽은 AC-FCR-022의 `TestFactoryRunsRetireRebindTo`·`TestLaneChoiceMergeOffer`가 (h)로 함께 단언한다.
- **변이 탐침**: 충돌 카드만 건너뛰고 나머지를 옮기는 구현 → 덤프 동일성에서 실패. `OwnerLive`만 막고 미확정은 통과시키는 구현 → (b)(e)에서 실패. 원천 분류를 보지 않는 `--retire <산 run> --rebind-to <산 run2>` → (d)에서 실패 — Q1 플래그 확대가 숨길 수 있는 우회로를 닫는 사례다.

### AC-FCR-022 — 운영자 표면: `--rebind-to`와 `m<n>` (REQ-FCR-021/020/014) [release-blocking; 플래그 모양 = decision-index Q1 판정 (a)]

- **Given** 죽은 스탬프 원천 run(비종결 카드 2장)과 `OwnerLive` 대상 run
- **When** (a) 운영자 세션에서 `runs --retire <원천> --rebind-to <대상>` (b) 레인 세션(`MOAI_FACTORY_WORKER=lane-1`)에서 같은 명령 (c) 대상이 `OwnerLive`가 아닌 run (d) `--rebind-to` 없이 `--retire <원천>` (e) 합류 프롬프트 행 계산과 `m<n>` 파싱 (f) 플래그 집합 확인 (g) `--rebind-to`만 주고 `--retire` 없음 (h) 원천이 `OwnerLive`·`OwnerIndeterminate`인 `--rebind-to`와 `m<n>`(AC-FCR-021 (d)(e)의 CLI 표면)
- **Then** (a) 원천 `retired`, 두 카드가 대상 아래, 출력에 두 카드 id; (b)(c) 거부, 상태 덤프 불변; (d) SPEC-FACTORY-RETIRE-GUARD-001 그대로 — `ErrRunHasInProgressCards`, 거부 출력의 기존 줄 전부 포함, 상태 불변; (e) `m<n>`은 `Dead && InProgress>0 && OwnerLive 활성 run 정확히 1`인 행에만 제시되고, 조건 밖 행의 `m<n>` 입력은 거부되며, `r<n>`의 제시 조건과 동작은 그대로다. 조건을 만족하는 행에서 `m<n>`을 선택하면 통합 연산이 실행된다 — 원천 `retired`, 카드가 대상 아래, 출력에 옮긴/포기한 카드 id(`TestLaneChoiceMergeOffer`가 함께 단언); (f) `runs` 명령의 플래그 집합은 정확히 `[rebind-to retire]`이다 — 이것이 기존 시험 `TestFactoryRunsHasNoForceFlag`에 허용된 유일한 기대 변경이며(Q1 판정 (a)), `--force` 같은 우회 플래그는 여전히 없어야 한다(판정 명령은 AC-FCR-025의 셋째 명령); (g) 상태 변경 없이 거부; (h) 거부, 상태 덤프 불변
- **RED-now**: L3 — runs 명령 플래그가 정확히 `[retire]`(재결속 표면 없음). L4 — 합류 프롬프트는 `r` 접두만 파싱한다(`:207`).
- **green (M2-B)**: 두 명령을 **각각 따로** 실행한다(OR 셀렉터 한 번은 한쪽만 있어도 통과해 빈 스윕을 숨긴다): `go test ./internal/cli -run '^TestFactoryRunsRetireRebindTo$' -count=1 -v` → `--- PASS: TestFactoryRunsRetireRebindTo `(이름 뒤 공백); `go test ./internal/cli -run '^TestLaneChoiceMergeOffer$' -count=1 -v` → `--- PASS: TestLaneChoiceMergeOffer `(이름 뒤 공백). 둘 다 exit 0.
- **변이 탐침**: 레인 세션 확인을 빠뜨린 구현 → (b)에서 실패. `--rebind-to` 없는 은퇴까지 카드를 옮기게 바꾼 구현 → (d)와 AC-FCR-025에서 실패. `--rebind-to`를 단독으로도 받아들이는 구현 → (g)에서 실패. 원천 분류를 보지 않는 구현 → (h)에서 실패. `m<n>`을 파싱만 하고 통합 연산을 부르지 않는 구현 → (e)의 정상 선택 단언에서 실패.

### AC-FCR-023 — 임대가 죽은 run의 중복 행을 남기지 않는다 (REQ-FCR-019/023) [release-blocking]

- **Given** 죽은 스탬프 run `run-dead`에 카드 `c9(run, stage 있음)`, 현재 산 run `run-now`; 대조로 `OwnerLive` run `run-live`에 카드 `c8(run)`, 은퇴했지만 `c7(assigned)`을 남긴 `run-ret`(소유자 dead), 그리고 리더는 죽었지만 레인이 살아 있을 수 있는 대조 행 둘 — 죽은 스탬프 run `run-dead2`에 **`c6a`(run, 유효 lease, worker 행 없음)**, 죽은 스탬프 run `run-dead3`에 **`c6b`(run, 만료된 lease, 산 worker pid=현재 프로세스)**
- **When** `run-now`에서 `moai factory next --card c9`, `--card c7`, `--card c8`, `--card c6a`, `--card c6b` 경로로 각각 기록한다. 추가로 (주입) `c9` 경로에서 T1 기록 뒤·포기 전에 실패를 주입한다
- **Then** `run-dead`의 `c9`와 `run-ret`의 `c7`은 `abandoned`이고 각각 `card.rebound` 이벤트(이전 state·version·stage·evidence_sha·`to_run=run-now`) 1개; `run-now`의 `c9`는 기존 이어받기 규칙 그대로(이어받기 읽기가 포기보다 먼저); `run-live`의 `c8`은 바뀌지 않았다; **`c6a`와 `c6b`의 두 원천 행 모두 바뀌지 않았다**(기존 동작 — 새 행 생성, 원천 행 불변); 주입 실패 시 `run-now`의 `c9` 행과 `run-dead`의 `c9` 행이 모두 진입 전 그대로다(한 트랜잭션); `cards` 행 수는 새로 생긴 `run-now` 행만큼만 늘었다(삭제 0)
- **RED-now**: L1 `TestT1963RedLeaseLeavesDeadRowNonTerminal` — `dead-run row state="run" after lease in run-new`. RED인 이유: 임대는 현재 run에만 쓰고 이전 run 행을 읽기만 한다(`card_picked.go:159-176`, `card_metadata.go:284`).
- **green (M2-B)**: `go test ./internal/cli -run '^TestFactoryNextAbandonsDeadRunDuplicate$' -count=1 -v` → exit 0, `--- PASS: TestFactoryNextAbandonsDeadRunDuplicate `(이름 뒤 공백).
- **변이 탐침**: 소유자 판정 없이 다른 run의 행을 모두 포기하는 구현 → `c8` 대조에서 실패. 리더 판정만 보고 lease·산 worker를 보지 않는 구현 → `c6a`·`c6b` 대조에서 실패. lease만 보는 구현 → `c6b`에서, worker만 보는 구현 → `c6a`에서 실패. 포기를 이어받기 읽기보다 먼저 하는 구현 → `run-now`의 `c9` stage 이어받기 단언에서 실패. 포기를 별도 트랜잭션으로 하는 구현 → 주입 실패 단언에서 실패.

### AC-FCR-024 — 죽은 pid의 worker는 run을 붙잡지 않는다 (REQ-FCR-022) [release-blocking]

- **Given** 소유자 dead인 활성 run 셋, 각각 worker 행 하나만: (a) pid 탐침이 `Dead`를 내는 pid (b) 현재 프로세스 pid(`Live`) (c) 탐침이 판정 불가를 내도록 주입한 pid; 그리고 (d) 죽은 스탬프 run 하나 — 비종결 카드 없음, 죽은 pid worker 행만
- **When** (a)~(c)는 `ReconcileActiveRuns`를, (d)는 리더 진입의 채택 후보 탐색을 실행한다
- **Then** (a)만 `retired`, (b)(c)는 `active`로 보존된다; (d)의 run은 채택 후보가 아니다(같은 술어가 후보 경로에도 적용됨)
- **RED-now**: L1 `TestT1963RedGhostWorkerKeepsRunAlive` — `status="active" preserved=1`. RED인 이유: `runHasLiveRecords`가 worker 행 존재만 본다(`factory_run_adopt.go:35`).
- **green (M2-B)**: `go test ./internal/homestate -run '^TestRunHasLiveRecordsIgnoresDeadWorkers$' -count=1 -v` → exit 0, `--- PASS: TestRunHasLiveRecordsIgnoresDeadWorkers `(이름 뒤 공백).
- **변이 탐침**: worker 절을 통째로 지운 구현 → (b)에서 실패. 판정 불가를 죽음으로 읽는 구현 → (c)에서 실패. 조정 경로에만 죽은 worker 필터를 넣고 채택 후보 경로는 빠뜨린 구현 → (d)에서 실패.

### AC-FCR-025 — 기존 보존 기준 (REQ-FCR-021/022/023) [regression-guard]

- **Given / When**: M2-B 착지 뒤 트리에서 아래 명령을 각각 따로 실행한다.
- **Then**: 모두 exit 0이고 열거한 시험 이름마다 `--- PASS` 줄이 있다.
  - `go test ./internal/homestate -run '^(TestRetireRunIfDead.*|TestFindResumableRun.*|TestAdoptRun.*|TestReconcile.*|TestLiveRecordsTerminalSetMatchesSharedDefinition)$' -count=1 -v` — L5의 25개 최상위 시험 전부, **시험 파일 변경 없이** — 단 **열거된 픽스처 예외 1건**: `mustInsertWorker`(`internal/homestate/factory_run_adopt_test.go:20-26`)의 worker pid를 `1`에서 `os.Getpid()`로 바꾸는 수정만 허용한다(시험 의도 — 산 worker는 run을 보존한다 — 불변). 이유: REQ-FCR-022 하에서 pid 1은 Windows 탐침(`internal/homestate/pid_state_windows.go:11-14`, `ERROR_INVALID_PARAMETER` → Dead)에서 dead로 읽힐 수 있어 `TestReconcilePreservesDeadOwnerWithOnlyWorkers`가 windows 러너(`.github/workflows/ci.yml`의 windows-latest)에서 거짓이 될 수 있다 — 코드 판독에 근거한 추론이며 Windows에서 측정하지 않았다. 이 헬퍼는 `TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers`도 쓰지만 그 시험은 worker 생존을 보지 않는 `retireRun` 경로라 기대가 바뀌지 않는다.
  - `go test ./internal/cli -run '^(TestFactoryRunsRetireRefusesInProgressCards|TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun)$' -count=1 -v` — SPEC-FACTORY-RETIRE-GUARD-001 cli 시험 2건, 변경 없이.
  - `go test ./internal/cli -run '^TestFactoryRunsHasNoForceFlag$' -count=1 -v` — **SPEC-FACTORY-RETIRE-GUARD-001 시험 중 유일한 명시적 기대 변경**(Q1 판정 (a)): 허용 플래그 집합 단언이 `[retire]`에서 `[rebind-to retire]`로 넓어진 상태로 통과한다. 단언의 형태(정렬된 전체 집합의 정확 일치)는 그대로라 `--force` 같은 우회 플래그가 생기면 여전히 실패한다. 그 SPEC의 나머지 7건(homestate 5건은 위 첫 명령, cli 2건은 위 둘째 명령)은 시험 파일 변경 없이 통과한다.
  - `go test ./internal/cli -run '^(TestAdoptResumedLeaderRun.*|TestStrictResumeCarrierSeen|TestLeaderRunIDWasMinted)$' -count=1 -v` — L6의 7건. **열거된 예외 1건**: `TestAdoptResumedLeaderRunWarnsOnMultipleCandidates`는 REQ-FCR-017 하에서 기대가 반전돼(최신 채택 + 나머지 통합) 개정된 기대로 통과한다. `strictResumeCarrierSeen`이 제거되면 `TestStrictResumeCarrierSeen`도 함께 제거되고, 그 사실을 run-phase 증거에 적는다.
- **RED-now**: 해당 없음 — 보존 기준이다(L5·L6이 착지 전 green 기준선). release-blocking이 아니라 regression-guard로 분류한다.
- **변이 탐침**: 은퇴 가드를 "옮길 수 있으면 옮기고 은퇴"로 일반화한 구현은 `TestRetireRunIfDeadRefusesInProgressCards`에서 실패한다.

## §C — M3: 표시·처분

### AC-FCR-012 — near-dup 발급 stderr 표시 (REQ-FCR-012)

- RED-now: P8 — 발급 stdout이 `"id position"`뿐, near-dup 정보 없음.
- green(M3): `go test ./internal/cli -run '^TestTodoAddNearDupDisclosure$'` — 0.80 이상 인접 발급 시 stderr에 관련 카드 id·점수·판정 안내; stdout은 여전히 정확히 `"id pos"` 한 줄. INPUT: backlog_analysis_test.go:37의 0.80 근접 점수 fixture 스타일.

### AC-FCR-013 — 관계 disposition 영속화 (REQ-FCR-013)

- RED-now: P5 — gtd_relations 스키마에 disposition 0건(gtd_items :36만 존재).
- green(M3): `go test ./internal/factory -run '^TestGTDRelationDisposition$'` — keep/merge/dismiss + 근거가 upsert로 영속화, 재upsert 시 갱신, 기존 행은 "미판정"으로 읽힘; item 쓰기 경로(gtd_clarify.go:86) 회귀 없음. INPUT: relation upsert fixture + 기존 clarify 테스트 재실행.

## §D — Cross-cutting

### AC-FCR-014 — 레인 경계: 발급·회수·재배정 표면 비노출 (REQ-FCR-014)

- RED-now: P1/P4(개념 부재 — 신규 표면이 생기며 처음 적용).
- green(M2/M3): `grep -rn 'LeaderApproval\|ApprovalReceipt' internal/cli --include='*.go' \| grep -v _test` 발급 함수가 리더 경로 진입만 노출하는지 코드 판독 + `go test ./internal/cli -run '^TestLeaderReceiptGateIssuerOnly$'` — 레인 권한 actor의 발급 호출 거부. 도달성: cli 스위트 exit code.

### AC-FCR-015 — transaction 내 재검증 + version-checked 전이 (REQ-FCR-015)

- RED-now: P3·P4(신규 경로 — 처음 적용).
- green(M2): 두 시험을 **각각 따로** 실행한다 — OR 패턴 한 번은 한쪽 시험만 존재해도 exit 0으로 `[no tests to run]` 없이 통과하므로(게이트 실측) 빈 스윕이 은폐된다: `go test ./internal/homestate -run '^TestReaperTransactionalReverify$'`(만료 재검증이 transaction 안에서 행을 다시 읽는지 — 도중 변경 fixture)와 `go test ./internal/homestate -run '^TestOperatorReassignEdge$'`(reassign이 version 충돌 시 거부) — 각각 exit 0 + `[no tests to run]` 아님 둘 다 충족. INPUT: transaction 도중 행 변형 fixture.

### AC-FCR-016 — 어떤 회수 경로도 완료 처리하지 않음 (REQ-FCR-016)

- RED-now: P3/P6(신규 경로 — 처음 적용).
- green(M2): `go test ./internal/homestate -run '^TestRecoveryNeverCompletes$'` — reaper·watchdog·reassign·CI reader 실행 후 어떤 카드도 done이 아님. INPUT: 회수 대상 행 + 완료 조건 충족 행 공존 fixture.

## §E — 커버리지·회귀 하한

- §E의 `-run` 패턴은 접두사군 선택이 **의도**다(예: `TestTodo|TestAutoDone|TestFactory` = 각 계열 전체) — §0의 앵커 규율은 AC별 단일 신규 테스트에 적용되고, 회귀 하한의 가족 패턴에 일부러 적용하지 않는다(빈 선택 함정 — wrapped-anchor 제안을 블라인드 적용 금지). 단 스윕 카운트 확인은 동일: `[no tests to run]`은 미측정.
- `go test -cover ./internal/cli/... ./internal/homestate/... ./internal/factory/...` — 소관 패키지 85% 유지(신규 파일 포함).
- 기존 회귀 하한(변경 전 green이어야 하고 M1-M3 후에도 green): `go test ./internal/cli -run '^(TestFR_.*|TestTodo.*|TestAutoDone.*|TestFactory.*)$'`, `go test ./internal/homestate -run '^(TestFR_.*|TestLease.*|TestTransition.*|TestReconcile.*|TestVerdict.*)$'`(TestFR_ 계열 = 전이 원자성·만료 회수·예약 edge 등 변경 대상의 핵심 가드 — `TestFR_AC003_TransitionIsAtomic`·`TestFR_AC011_ExpiredLeaseReturnsToAssigned`·`TestFR_AC019_ReservedCIEdgesRefused` 포함; **CLI 패키지에도 TestFR_이 있다** — `TestFR_AC018_DecidePushGate`가 그 증명), `go test ./internal/factory -run '^(TestBacklog.*|TestGTD.*)$'`. 기준선 측정 시 `| tail` 파이프에는 `set -o pipefail`을 앞세킨다 — tail의 exit 0이 테스트 FAIL을 가리지 않게. **열거된 예외 — 기대 반전 2건 + 유지 가드 1건**: REQ-FCR-002(b) 하에서 기대가 반전되는 것은 `TestFR_AC018_PushGateStore`(homestate:191)·`TestFR_AC018_DecidePushGate`(cli:456) 둘뿐이다(완료 전이 성공 단언 → receipt 없으면 거부로) — 하한의 "변경 전 green → M1-M3 후에도 green"은 이 둘에 한해 "M1 수정 뒤 green(수정된 기대)"으로 읽는다. `TestFR_AC018_PushGateNeverFetches`(cli:20)는 반전 대상이 아니라 **유지되는 회귀 가드다**(no-fetch 단언은 receipt 게이트와 무관하게 계속 성공 — codex round-3 재측정: 3건 전부 as-is PASS, 반전은 done 전이 2건뿐).

## 매핑 요약

| AC | REQ | milestone | 패키지 |
|---|---|---|---|
| AC-FCR-001 | REQ-FCR-001, REQ-FCR-002 | M1 | internal/cli (+internal/factory) |
| AC-FCR-002 | REQ-FCR-002, REQ-FCR-005 | M1 | internal/cli (+internal/factory) |
| AC-FCR-003 | REQ-FCR-005 | M1 | internal/cli |
| AC-FCR-004 | REQ-FCR-003 | M1 | internal/cli |
| AC-FCR-005 | REQ-FCR-004 | M1 | internal/cli |
| AC-FCR-006 | REQ-FCR-006 | M2 | internal/homestate |
| AC-FCR-007 | REQ-FCR-007 | M2 | internal/homestate |
| AC-FCR-008 | REQ-FCR-008 | M2 | internal/homestate |
| AC-FCR-009 | REQ-FCR-008 | M2 | internal/homestate |
| AC-FCR-010 | REQ-FCR-009 | M2 | internal/homestate |
| AC-FCR-011 | REQ-FCR-010, REQ-FCR-011 | M2 | internal/homestate |
| AC-FCR-012 | REQ-FCR-012 | M3 | internal/cli + internal/factory |
| AC-FCR-013 | REQ-FCR-013 | M3 | internal/factory |
| AC-FCR-014 | REQ-FCR-014 | M2-M3 | internal/homestate + internal/cli |
| AC-FCR-015 | REQ-FCR-015 | M2 | internal/homestate |
| AC-FCR-016 | REQ-FCR-016 | M1-M3 | internal/homestate + internal/cli |
| AC-FCR-017 | REQ-FCR-017, REQ-FCR-018, REQ-FCR-020, REQ-FCR-023 | M2-B | internal/cli + internal/homestate |
| AC-FCR-018 | REQ-FCR-017, REQ-FCR-018, REQ-FCR-020 | M2-B | internal/cli |
| AC-FCR-019 | REQ-FCR-018, REQ-FCR-020, REQ-FCR-023 | M2-B | internal/cli + internal/homestate |
| AC-FCR-020 | REQ-FCR-020 | M2-B | internal/homestate |
| AC-FCR-021 | REQ-FCR-020, REQ-FCR-023 | M2-B | internal/homestate |
| AC-FCR-022 | REQ-FCR-021, REQ-FCR-020, REQ-FCR-014 | M2-B | internal/cli |
| AC-FCR-023 | REQ-FCR-019, REQ-FCR-023 | M2-B | internal/cli + internal/homestate |
| AC-FCR-024 | REQ-FCR-022 | M2-B | internal/homestate |
| AC-FCR-025 | REQ-FCR-021, REQ-FCR-022, REQ-FCR-023 | M2-B | internal/homestate + internal/cli |

전체 REQ-FCR-001..023이 위 표에서 최소 1개 AC와 연결된다(coverage reader가 읽는 전체 ID 형태 — 축약 `FCR-001..005`는 reader 미인약, 게이트 실측으로 16건 전부 CoverageIncomplete로 보고된 바 있음).
