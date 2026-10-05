# acceptance.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

## §0 — 판정 형식과 RED-now 셀

각 기준은 두 셀을 한 쌍으로 채택한다. **RED-now 셀** — 본 트리(base `a158b4b5f`, plan-phase 2026-10-06)에서 실측한 결함 상태의 관측(command + verbatim output + exit code + tree SHA); 새 테스트 이름은 구현 전에 존재하지 않으므로 RED 근거는 **존재 관측**(부재 probe 또는 현재 동작 판독)이고, `[no tests to run]` exit 0은 결코 통과가 아니다(verification-completeness §1.1 — empty-sweep). **green path 셀** — 어느 milestone이 뒤집는지와 통과 시 출력 형태. green 판정 명령은 빈 스윕이면 실패로 읽는다: `go test -run`이 `[no tests to run]`을 출력하면 그 기준은 PASS가 아니라 **미측정**이다. 세 부분(WHEN=소속 milestone 패키지 스위트, INPUT=red를 만드는 fixture, 도달성=go test exit code)은 각 행에 명시한다.

RED-now probe 일괄 관측 기록(모두 `a158b4b5f` 트리, 2026-10-06 plan-phase):

| probe | command | verbatim 결과 | exit |
|---|---|---|---|
| P1 receipt 부재 | `grep -rn 'Receipt\|receipt' internal/cli/todo_autodone.go internal/cli/todo.go` | `internal/cli/todo.go:818:	// REQ-TSS-001 family) — the response's issued id is a receipt for the` + `internal/cli/todo.go:873:	// the issued id is a receipt for the store that answered. Scoped to the` | 0 |
| P2 facts 필드 | `grep -n 'type AutoDoneFacts' -A 12 internal/factory/*.go` | `internal/factory/autodone_scan.go:108:type AutoDoneFacts struct {` 이하 RecordedSHA·SHAReachable·SubjectHit·SubjectKnown… — receipt 필드 없음 | 0 |
| P3 active-only | `grep -n 'classifyRuns(ctx, opts, ' internal/homestate/factory_run_retire.go` | :162 `""`(전체), :333·:350 `"active"` | 0 |
| P4 reassign 부재 | `grep -in 'reassign' internal/homestate/*.go` | (출력 없음) | 1 |
| P5 관계 disposition 부재 | `grep -n 'disposition' internal/factory/backlog_gtd_schema.go internal/factory/gtd_relation.go` | backlog_gtd_schema.go:36만(gtd_items) — gtd_relations·gtd_relation.go 0건 | 0 |
| P6 watchdog 부재 | `grep -rn 'watchdog\|waiting_since\|review_deadline\|WaitingSince' internal/homestate --include='*.go' \| grep -v _test` | (출력 없음 — 두 번째 grep이 선택한 행이 없어 exit 1) | 1 |
| P7 reserved edge | `grep -n 'CardPushed\|CardCIGreen\|reserved' internal/homestate/card_transition.go` | :173 술어, :267 거부 | 0 |
| P8 발급 출력 | `sed -n '851p' internal/cli/todo.go` | `fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", item.ID, pos)` | 0 |

## §A — M1: 완료 게이트

### AC-FCR-001 — receipt 없는 done 거부 (REQ-FCR-001/002)

- RED-now: P1 — receipt 게이트 개념 부재(todo.go:1104 `rec.ArchiveCard(id)`가 검증 없이 도달).
- green(M1): `go test ./internal/cli -run '^TestLeaderReceiptGateRejectsMissingReceipt$'` — receipt 없는 done이 거부되고 stderr에 이유, 행은 archive되지 않음. INPUT: fcFixture 루트 + factory-linked 카드. WHEN: M1 이후 cli 스위트.
- 기준: 명령 exit 0이고 `[no tests to run]` 아님.

### AC-FCR-002 — receipt 네 바인딩 불일치 거부 (REQ-FCR-002/005)

- RED-now: P1 동일.
- green(M1): `go test ./internal/cli -run '^TestLeaderReceiptGateBinding$'` — UUID 불일치·**run id 불일치(다른 run의 receipt)**·factory version stale·증거 해시 불일치 각각 거부, 네 값 모두 일치 시 통과. INPUT: 네 바인딩 값을 하나씩 틀어놓은 fixture receipt 4종 — run 변이는 동일 카드·version·증거 SHA를 가진 두 번째 run(측정: cards 스키마에서 생성 가능).

### AC-FCR-003 — 수행자 발급 receipt 거부, 리더 발급+리더 실행 허용 (REQ-FCR-005)

- RED-now: P1 동일.
- green(M1): `go test ./internal/cli -run '^TestLeaderReceiptGateSelfIssued$'` — 카드의 수행 owner(lane)가 발급자로 표식된 receipt는 거부; **리더가 발급하고 같은 리더가 done을 실행하는 경로는 통과**(단일 리더 정상 흐름 — REQ-FCR-005의 독립 축은 수행자≠승인자). INPUT: 발급자==수행자 fixture + 발급자==리더 fixture 2종.

### AC-FCR-004 — auto-done이 receipt 없이 닫지 않음 (REQ-FCR-003)

- RED-now: P2 — `AutoDoneFacts`(autodone_scan.go:108)에 receipt 필드 없음.
- green(M1): `go test ./internal/cli -run '^TestAutoDoneReceiptSkip$'` — receipt 미검증 후보는 close 대상에서 skip/downgrade, 스캔 전체는 계속. INPUT: 조건은 모두 충족하되 receipt 없는 카드. **경합 변이 포함**: 스캔 승인 뒤 factory 전이로 카드 version이 증가한 행 — archive 직전 네 바인딩 재검증(factory version·run id 포함)이 stale receipt를 거부하고 close를 skip한다(REQ-FCR-004의 직전 재검증).

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
- green(M2): `go test ./internal/homestate -run '^TestStalledWaitWatchdog$'` — owner 있음+임대 부재 행이 기한 초과 시 재판정 요구 기록; 자동 완료·자동 해제 0. INPUT: 구동 중 정지 행 fixture + 시간 진행.

### AC-FCR-011 — CI reader가 정확한 SHA만 인정 (REQ-FCR-010/011)

- RED-now: P7 — T19/T20이 :267에서 예약 거부 중("CI verdict reader ... owned by F3").
- green(M2): `go test ./internal/homestate -run '^TestCICompletionReader$'` — push 시 기록된 SHA의 CI 증거는 **T19(pushed→ci-green)을 통과**, 다른 SHA 증거·증거 부재는 T19도 계속 거부; **T20과 T18은 receipt 게이트만 연다**(`FactoryDB.Transition` 안에서) — receipt 없는 T20·T18(원격 없는 저장소 포함)은 계속 거부, 올바른 receipt와 함께 성공(REQ-FCR-002b·016과 모순 없음). INPUT: frWriteVerdict 선례(fr_fixture_test.go:108)의 sha 바인딩 변이 + receipt 변이.

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
- green(M2): `go test ./internal/homestate -run '^(TestReaperTransactionalReverify|TestOperatorReassignEdge)$'` — 만료 재검증이 transaction 안에서 행을 다시 읽는지(도중 변경 fixture), reassign이 version 충돌 시 거부하는지. INPUT: transaction 도중 행 변형 fixture.

### AC-FCR-016 — 어떤 회수 경로도 완료 처리하지 않음 (REQ-FCR-016)

- RED-now: P3/P6(신규 경로 — 처음 적용).
- green(M2): `go test ./internal/homestate -run '^TestRecoveryNeverCompletes$'` — reaper·watchdog·reassign·CI reader 실행 후 어떤 카드도 done이 아님. INPUT: 회수 대상 행 + 완료 조건 충족 행 공존 fixture.

## §E — 커버리지·회귀 하한

- §E의 `-run` 패턴은 접두사군 선택이 **의도**다(예: `TestTodo|TestAutoDone|TestFactory` = 각 계열 전체) — §0의 앵커 규율은 AC별 단일 신규 테스트에 적용되고, 회귀 하한의 가족 패턴에 일부러 적용하지 않는다(빈 선택 함정 — wrapped-anchor 제안을 블라인드 적용 금지). 단 스윕 카운트 확인은 동일: `[no tests to run]`은 미측정.
- `go test -cover ./internal/cli/... ./internal/homestate/... ./internal/factory/...` — 소관 패키지 85% 유지(신규 파일 포함).
- 기존 회귀 하한(변경 전 green이어야 하고 M1-M3 후에도 green): `go test ./internal/cli -run 'TestTodo|TestAutoDone|TestFactory'`, `go test ./internal/homestate -run 'TestLease|TestTransition|TestReconcile|TestVerdict'`, `go test ./internal/factory -run 'TestBacklog|TestGTD'`.

## 매핑 요약

| AC | REQ | milestone | 패키지 |
|---|---|---|---|
| AC-FCR-001..005 | FCR-001..005 | M1 | internal/cli (+internal/factory) |
| AC-FCR-006..011 | FCR-006..011 | M2 | internal/homestate |
| AC-FCR-012..013 | FCR-012..013 | M3 | internal/cli + internal/factory |
| AC-FCR-014..016 | FCR-014..016 | M1-M3 | internal/homestate + internal/cli |
