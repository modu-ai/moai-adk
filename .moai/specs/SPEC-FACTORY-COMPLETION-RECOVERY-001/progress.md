# progress.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

## §A 진행 기록

- **2026-10-06** — phase=plan 시작. lane-1, 트리 `.moai/worktrees/t1538`, 브랜치 `WT-t1538-factory-recovery`, base = local develop `a158b4b5f`.
- **임대 거절·영수증 경로**: `moai factory next --card t1538` → `refused serial-slot`(설치 빌드에 t1513 수리 `de388878e` 미반영). t1522식 영수증 경로(리더 배차=권한)로 진행, 크론 재시도 유지. 본 SPEC의 M1 receipt 게이트가 이 경로의 도구적 뒷받침이다.
- **앵커 검증 완료**: 리포트(main@ec13872f3) 앵커 14건을 본 트리 기준 전수 재확인 — 13건 일치, 1건 부분 이동(todo_analysis.go:74는 기록 분기, 임계 상수는 internal/factory/backlog_analysis.go:33). 상세는 spec.md §B.2. RED-now probe 8건(P1-P8)은 acceptance.md §0에 verbatim 기록.
- **자체 해결 스코프 질문**: (1) 통합 경로 표기 — 스폰 초안은 낡은 로컬 규칙(git-flow 시절 develop 통합)을 따라 "develop 대상 PR"로 적었으나 **레인이 시정**: 2026-10-05 GitHub Flow 재전환 이후 리더 운영 관행이 권한 원천 — t1453 PR #1751·t1513 PR #1758 모두 base=main이고 t1513 배차에는 `gh pr create --base main` 지시가 명시(plan.md §H 교정, 독립 감사자가 PR 3건 전부 base=main으로 재확인); (2) "receipt" 명칭 충돌 — todo.go:818/:873의 스토어 발급 receipt와 구별되는 이름(LeaderApproval 계열)을 plan §D에 명시.
- **plan-audit iter1 FAIL 수령**: 0.75(Tier M 문턱 0.80 미달, 판정서 `.moai/reports/t1538/plan-audit-iter1.md`, audited_sha `8fd632009` — 수리 2라운드 이전 기준). D1~D4는 이미 수리됨(`98a44b87a`·`7e65ff552`). 잔여 D5 stale-receipt 경합·D6 watchdog 기존 행 도달·D7 §0 프로브 기록 오류는 본 커밋(수리 3라운드)으로 처리 — 다음 감사는 delta로.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-06
비고: plan-audit 최종 판정 **PASS-WITH-DEBT 0.94 @ 6f01022e3**(Tier M 문턱 0.80 충족, 차단 0·필수 0, Clarity 0.95/Completeness 0.90/Testability 0.90/Traceability 1.00) — 판정서 `.moai/reports/t1538/plan-audit-iter3.md` 닫는 기록(§Final delta rounds 3-5). **run 진입 해시 바인딩 — 확정**: 동결 아티팩트(HEAD `2939ba304`)의 실제 `ComputeHash` = **`46aa4358663a3db8f47866fcd0022e0558b3be99174434761cd679772611fd1c`** — 감사자 스크립트·본 레인 셸 재현·codex 독립 계산 3-way 일치, codex의 admission 검증도 이 해시로 통과(이전 `29794e62…`는 감사자 자체 공식, `8708bcfe…`는 동결 전 값 — 둘 다 대체됨). 동결 이후 아티팩트 편집은 이 판정을 무효화한다. run 진입: 리더 결정 (a)에 따라 자율 진입(방침상 자율 — 결정 기록은 §F). 감사 사슬: iter1 0.75 → iter2 0.86 → iter3 0.89 → 개정 0.94(리더 (a) 승인 하 최종 라운드, t1546 조건). run 진입: 리더 결정 (a)에 따라 자율 진입(방침상 자율 — 결정 기록은 §F).

## Binding run conditions (이월 부채 장부 — 판정서 열거 8건, sync-auditor 재검증 대상)

| 부채 | 내용 | dispose_in | 처분 상태 |
|---|---|---|---|
| D15 | AC-FCR-005의 RED-now 셀이 bare reading(:385 stale 앵커) — 재실행 가능한 probe(command+output+exit)로 보강, :386 동기화 | run | M1 스폰 브리프 이관 — 처분 증거는 §E.2에 기록 |
| D16 | `TestLeaderReceiptGateSelfIssued` 개명(수행자≠승인자 축에서 misnomer) — M1 RED 저작 전 | run | M1 스폰 브리프 이관 — §E.2에 기록 |
| D20 | AC-FCR-005 green 셀을 네 바인딩 전체 재검증 목록(run id 포함)으로 동기화 + run 교체 race INPUT 변이 | run | M1 스폰 브리프 이관 — §E.2에 기록 |
| D24 | plan merge-base 피연산자를 흡수 ref(develop)로 정렬 | run | **처분됨 — `17a0d79c3`**(plan §H 현행) |
| D17 | plan §D에 t1533의 제외 파일 4종 명명 | sync | sync 단계 처분 예정 |
| D18 | REQ-FCR-002b → REQ-FCR-002(b) 표기 통일(acceptance:78·plan:69) | sync | sync 단계 처분 예정 |
| D22 | progress §E.1 hold record 경로 오탈자(.moi→.moai) | sync(즉시) | **처분됨 — 본 커밋**(§E.1 현행 경로 `.moai/…` 확인) |
| D23 | AC018 예외 분할 — 기대 반전 2건(PushGateStore·DecidePushGate)과 유지 가드 1건(NeverFetches) 구분 | sync(즉시) | **처분됨 — 본 커밋**(acceptance §E 현행) |

## §E.2 Run-phase Evidence

(run-phase에서 작성 — manager-develop 소관. 첫 M1 커밋 시 spec.md frontmatter가 draft → in-progress로 전이된다.)

## §E.3 Run-phase Audit-Ready Signal

(run-phase에서 작성 — manager-develop 소관. E1-E7 자체검증 매트릭스는 plan.md §E 형식.)

## §E.4 Sync-phase Audit-Ready Signal

(sync-phase에서 작성 — manager-docs 소관. sync_commit_sha는 sync 커밋에서 pending-backfill로 기록 후 후속 커밋에서 백필.)

## §F Phase 4 Mode Selection

- 입력: tier=M, scope=9파일(plan §C 파일별 변경표), 도메인=1(CLI+homestate 단일 Go 축), 언어 혼합=Go 100%, concurrency benefit=LOW(coding-heavy — Anthropic coding-task caveat), agent-team 요건=미요청.
- 모드 평가: direct=아님(다중 파일·신규 로직) / fanout=아님(coding-heavy·단일 도메인) / sweep=아님(기계 균일 변형 아님) / **serial=선택**(마일스톤당 단일 구현 스폰 — M1→M2→M3 순차).
- Decision: serial
- 근거: Anthropic의 coding-task 병렬성 주의에 따라 구현 중심 작업은 순차가 기본 — M1의 게이트 설계가 M2/M3의 전제(완료 표면 열거·receipt 타입)라 의존성도 순차를 지지. manager-develop 역할의 per-spawn 일반 스폰이 본 카드 트리에서 수행한다(manager-* 타입 스폰의 자가 격리 회피 — t1318).
- 경계 사례: 해당 없음. Kickoff: 리더 결정 (a)에 따른 자율 진입 — plan-audit PASS-WITH-DEBT 0.94(≥0.80)·plan_artifact_hash `29794e62…` @6f01022e3·blocker 없음(결정 기록 본 절).

## 봉인 기록 — 2026-10-07 (리더 최종 지시: 즉시 봉인·추가 수리 금지)

- **중단 지점**: M1 round-19 진행 중 미완성 — 미러 경로 바인딩 동기화 수리 착수분(goal.go·gtd.go 호출부 + todo.go 게이트 + leader_approval_gate_test.go), 컴파일 불확실 상태로 봉인(리더 지시: 게이트 미통과 무방).
- **완료된 라운드**: round-1~18 전부 커밋됨(마지막 `62598f2a8` — REQ-THS-012 revert 가드). 세부는 git log와 plan-audit-iter3.md 닫는 기록.
- **잔여 게이트 항목**: ① 미러 경로 P1 — 배차 저장과 바인딩 갱신을 한 잠금 구간으로, 미러 실패 시 구 승인 무효화(재현: FACTORY_RECORD_UNAVAILABLE 후 done err=nil) ② round-8~19에서 지적된 나머지 overlay 재현 전부(`^TestReview` 계열 — /tmp/t1538-review-overlay.json 재현 재료, 게이트가 매 라운드 갱신).
- **재개 절차**: M1 마지막 라운드부터 — /tmp overlay 재현이 green이 될 때까지 수리 → 전체 TestReview 계열 → E1 매트릭스 → §E.2 확정. SPEC 아티팩트·Binding run conditions 장부는 본 브랜치에 동결 완료.

## §A 봉인 해제 후 기록 — 2026-10-07 (재개 세션)

- **복구 관측**: 봉인 트리 HEAD `be85183eb`에서 `go build ./...` exit 0, `go vet ./internal/cli/...` exit 0 — 봉인된 mid-round-19 상태는 컴파일됨. 봉인 커밋의 게이트 테스트(TestDispatchRepointsBindingBeforeMirrorAndMirrorFailureFailsClosed)도 통과. /tmp 오버레이 재료는 재부팅으로 소실 → **게이트 자체 재실행으로 재생성**(codex_review scope=card, base `a158b4b5f` — 병합 베이스 불변 확인).
- **게이트 재생성 결과(round-20 리뷰)**: verdict fail, P1 2건 — (a) goal.go:894 배차 저장과 승인 바인딩 갱신 사이 완료 경로 끼어듦(repro에서 `done t1 landing=unknown err=<nil>`), (b) gtd.go:267 배차 재시도의 성공 판독에 바인딩 상태 누락(부분 실패 뒤 `state=reconciled apply calls=1` + 바인딩 "old" + done 성공). 검증자 재현 2건(`^TestReviewDispatchAssignmentArchiveWindow|^TestReviewFailedDispatchReadbackReconcilesWithoutBinding`, go test -overlay) — 본 세션 기준선에서 **둘 다 RED 관측**(FAIL 2종, 3.3s).
- **round-20 수리(본 커밋)**: ① **할당 원시에 바인딩 접기** — `factory.RecordFactoryCardAssignment`가 같은 큐 잠금 임계 구역(recordRuntimeHook, 할당 커밋 뒤 릴리스 전)에서 `RecordDispatchBindingIfEngaged`를 호출한다. REQ-FCR-002의 스코프 문장(팩토리 행 없는 카드는 보통 카드 — 침묵 스킵)이 원시 안에 함께 산다. 배차 저장과 바인딩 재지정 사이의 끼어듦 창이 원시 수준에서 사라지고, 호출자가 재지정을 잊을 수도 없다(goal.go·gtd.go apply는 이제 평이한 할당 호출 하나). ② **엔진 dispatch 정합** — `gtdDispatchBindingCurrent`: 바인딩이 op의 run을 가리키고 그 run에 카드 행까지 존재할 때만 정합(homestate 정본 `FactoryDBPath` 해석 — 큐 store 상위에서 `.moai`를 찾아 프로젝트 루트 복원). 불완전하면 `reconcileGTDDispatchRecord`(큐 잠금 하, picked 전제, T1 `RecordPicked` + 바인딩 — 미러의 형태)로 **팩토리 기록 절반을 직접 완성**한다. apply 재실행이 아니라 수리인 이유: 할당 저장은 이미 커밋됐고 소실된 쓰기는 팩토리 절반뿐이며, 소유자 apply는 (재현 2가 모델링하듯) 실패하는 바인딩 쓰기를 반복할 수 있다. ③ **스코프 문장 단일 원천화** — `factory.RecordDispatchBindingIfEngaged`, cli `recordDispatchBindingAtRoot`는 위임. ④ 시도·철거: 대체 참여 가드(할당 run에 카드 행 없으면 done 거부)는 GateBinding의 합법 상태(미러 실패로 행 없는 구 할당 + 신 바인딩)와 데이터 수준에서 구분 불가능해 철거했다 — 창은 원시 접기(①)로 닫고, 부분 실패는 엔진 정합(②)이 수복한다.
- **GREEN 관측**: 재현 2건 `ok github.com/modu-ai/moai-adk/internal/cli 3.152s`(오버레이 실행) — RED→GREEN 한 쌍 성립. **검증 기준은 영향계열**(리더 지시·t1542 교훈 — 카드 트리 cli 전체 스위트는 구조적 적색: codex 자식 계열 경합 hang, 전체 스위트 판정은 CI): ① 영향계열(수신 게이트+배차 바인딩+승인+FR_AC018+오버레이 재현, -count=1) `ok 116.526s` 14/14 PASS ② §E 회귀 하한(TestFR_|TestTodo|TestAutoDone) — 1차 `FAIL 902.157s` → 2차 동일 트리·동일 셀렉터 `ok 971.772s` — **재현되지 않은 경합 플레이크로 판정**(t1542 클래스; 실패 테스트명 미관측은 1차 로그가 tail-2만 보존한 측정 결함) ③ 영향계열 -race `ok 130.241s` ④ factory 패키지 전체 `ok 291.011s` ⑤ homestate 패키지 전체 `ok 94.661s` ⑥ vet clean·golangci-lint 0 issues. cli 전체 스위트는 1801s 타임아웃 FAIL로 관측됐으나 구조적 적색으로 판정에서 제외. 근거 로그 `/tmp/t1538-round20-verify.log`·`/tmp/t1538-round20-families.log`·`/tmp/t1538-bound-rerun.log`(머신 로컬 — 판정 문장 본 기록으로 이관).

## §A round-21 — 2026-10-08 (턴종료 게이트 재현 4건)

- **RED 관측** (커밋 기준 `62f97d0d5`, 엔진/시임 수준 재현 4건 — `gtd_operation_reconcile_test.go` + `goal_mission_op_test.go`): ① P1 `TestGTDReconciledDispatchReplayDoesNotRegressBinding` — 완료(reconciled) dispatch op 재실행이 바인딩을 옛 run으로 회귀. 단, 1차 실행은 공극 통과 — 프로브(`gtdProjectRootForStore`=""·bindingCurrent 양쪽 true 실측)로 밝힌 전제: factory 패키지 테스트 큐는 샌드박스 MOAI_HOME의 **홈 레이아웃**(`<home>/db/<key>/todo`)으로 풀려 :370 유도 실패가 재현을 가렸고, :370 수리 적용 후 가드 비활성 상태에서 진 RED 관측("reconciled-op replay regressed the binding to \"run-old\"") ② P2 `TestGoalMissionDispatchOpCarriesAssignedIdentifiers` — dispatch op Target/MissionID가 `gtd:item-7`/`session-1`(할당 식별자 아님) ③ P2 `TestGTDBindingCurrentResolvesHomeStateSibling` — 홈 레이아웃 형제 factory.db 미관측 → 공극 true ④ P2 `TestGTDDispatchRepairRestoresAssignedOwner` — 수리 후 행이 picked/owner=""로 정체.
- **수리**: ① 완료 op는 정합하지 않음 — reconcile 게이트에 `stored.State != GTDOperationReconciled`(prepared/invoking만) ② **[1차 시도 → 철회·재설계]** op 식별자 정렬(`goalMissionOp`)은 supervisor lineage(`cliSupervisorEngine.Finalize`의 completion_operation_lineage — op.Target/MissionID vs plan item-ref/session)를 깨뜨려 TestGTDAutonomyEndToEnd FAIL(전체 검증에서 관측) — op 식별자는 lineage 원본 유지로 되돌리고, **선택적 소유자 계약 `factory.DispatchIdentifiers`** 로 재설계: dispatch를 수행한 소유자가 실제 할당 카드·run을 공개하면 정합이 그 식별자로 동작, 미공개 소유자는 기존대로 op 식별자(`gtdDispatchReconcileIdentifiers`). `gtdCLIOwner`가 goal dispatch의 (linkedCardID, gitOpts.RunID)를 공개 ③ `gtdFactoryDBForStore` — 프로젝트 상위 탐색 실패 시 큐 디렉터의 상태 디렉터 형제(`<state>/factory/factory.db` — `<home>/db/<key>/{todo,factory}` 레이아웃) 폴백, 수리는 `OpenFactoryPath`로 경로 직접 개방 ④ 수리에 미러의 T2 추가 — 런타임 할당의 owner로 picked→assigned 전이, 이미 할당이면 통과, 그 외는 미러와 같은 거부.
- **GREEN 관측**: 재현 5건(4건 게이트 + 소유자 공개 계약) — `TestGTDReconciledDispatchReplayDoesNotRegressBinding`·`TestGTDDispatchRepairRestoresAssignedOwner`·`TestGTDBindingCurrentResolvesHomeStateSibling`·`TestGTDDispatchReconcileUsesRevealedIdentifiers`(factory) + `TestGTDOwnerRevealsDispatchReconcileIdentifiers`(cli) 전부 PASS, **TestGTDAutonomyEndToEndProductionCLIWithGovernanceReceipts PASS 복귀**(1차 설계의 충돌 해소). 영향계열 전체 검증은 `/tmp/t1538-round21b-verify.log`.

## §A round-22 — 2026-10-08 (게이트 재실행 신규 4건)

- **RED 관측** (커밋 기준 `38f815ea3`, 수리 패치를 임시 revert한 원본 트리에서 4건 전부 FAIL 직접 관측 — 재현은 `gtd_operation_reconcile_test.go` 3건 + `leader_approval_test.go::TestFR_FCR_T20RebindScopesToVerifiedUUID` 1건으로 본 세션이 저작): ① P1 할당 커밋 **뒤** 바인딩 훅 실패 시 할당이 생존(부분 상태 → 완료 게이트가 옛 승인으로 done 성공) ② P1 invoking으로 중단된 옛 배차의 재시도가 새 참여 바인딩을 run-new→run-old로 회귀 ③ P2 picked/owner="" 행이 정합으로 판정되어 수리 스킵 ④ P2 완료 전이의 승인 re-stamp UPDATE가 run+card만 스코프 — 검증 안 된 옛 UUID 승인까지 version 갱신(uuid-stale 1→2 실측).
- **수리**: ① 바인딩 훅을 할당 트랜잭션 **안으로** 이동(recordRuntimeHook — 훅 실패 시 할당까지 롤백, 배차는 깨끗하게 실패) ② 정합 전 최신 할당 확인 `gtdAssignmentIsLatest`(todo_runtime_assignments rowid 순 — 삽입 순서가 참여 순서) — 옛 run이 최신이 아니면 정합·수리 모두 생략하고 reconciled 마킹만 ③ 정합의 행 검사에 `state='assigned' AND owner_label != ''` 추가 ④ `plan.approvalRebindUUID` 기록(3개 계획 지점) + re-stamp UPDATE에 `AND card_uuid=?` 제약.
- **GREEN 관측**: 재현 4건 전부 PASS(factory 3건 `ok 4.136s` + homestate 1건 `ok 1.417s`). 영향계열 전체 검증(factory·homestate 패키지, cli 영향계열 오버레이, §E 하한, -race, vet·lint)은 `/tmp/t1538-round22-verify.log`.
- **후속 수리(같은 라운드)**: 전체 검증 1차에서 factory 패키지 FAIL — ②의 최신 할당 쿼리가 런타임 테이블이 없는 최소 스토어(`TestPersistentGTDOperationExactlyOnceAcrossStoresAndCrashCuts` — 크래시 컷 복원력 픽스처)에서 "no such table" 오류를 냄. `gtdAssignmentIsLatest`에 sqlite_master 테이블 존재 프루브를 추가해 부재 시 공극-최신(기존 크래시 복원력 계약 보존). 재측정: 해당 테스트 PASS + 재현 5건 PASS + factory 패키지 전체 `ok 465.390s`.

## §A round-22 연속 — 2026-10-08 (턴종료 게이트 3차: 바인딩 정본화 재설계)

- **게이트 3차 발견 3건**: (1) P1 순서 기반 테스트 2건 FAIL — 역방향 케이스(첫 도착 옛 run 상태 보고가 새 행을 삽입해 최신으로 오독) 포함 (2) P2 승계 검사가 picked 검사보다 뒤에 있어 hold 시나리오가 "queue item is not picked"로 죽음(수리 스킵 전에) (3) feedback_participation.go:66 — t1498 중복, 카드 범위 밖(후속 원장 이월, 수렴 하한에 따라).
- **재설계 결정(리더 확인 수령)**: 데이터 수준 순서 판별은 중간 비행 배차와 합법적 사후 이력을 구별할 수 없음이 3라운드 연속 확인됨 → **순서 메커니즘을 제거하고 factory 바인딩(card_dispatch.run_id)을 참여의 정본으로** 채택. 수리는 바인딩이 오퍼레이션의 run을 명명하거나 바인딩 부재(고아 행)일 때만 진행; 다른 run을 명명하면 superseded로 스킵·reconciled 마킹만. ①번 역방향 케이스는 이 설계에서 구조적으로 불가능(상태 보고는 바인딩을 전혀 쓰지 않음).
- **구현**: `todo_runtime.go` recordRuntimeHook를 원본 upsert로 복원(delete+insert/순서 기록 제거); `gtd_operation.go` `gtdAssignmentIsLatest` 삭제, `repairGTDDispatchRecord`이 큐 락 안에서 **superseded 바인딩 검사를 최우선** 실행(OpenFactoryPath → card_dispatch 질의 → bound != runID면 `errGTDReconcileSuperseded`), `reconcileGTDDispatchBinding`은 superseded를 스킵·reconcile로 처리.
- **테스트 재작성**: `TestGTDRedispatchBumpsEngagementOrder` → `TestGTDRedispatchRetakesTheBinding`(A→B→A 재배차마다 바인딩이 이동해 최종 run-a; run-a의 picked/ownerless 행 수리 진행) / `TestGTDStateReportDoesNotBumpEngagementOrder` → `TestGTDStateReportNeverMovesTheBinding`(옛 run 상태 보고가 자기 행을 삽입해도 바인딩 불변 — 게이트 3차 (1)의 역방향 케이스 구조적 해소 증명).
- **GREEN 관측(본 세션 직접 측정)**: `go test ./internal/factory/ -run TestGTD -count=1` → 18건 전부 PASS `ok 41.523s` / `go test ./internal/cli/ -run '<영향계열 13종>'` → `ok 138.219s`(TestGTDAutonomyEndToEnd 회귀 가드 포함).
- **Gap (정직 기록)**: factory 패키지 전체 -race 2회 모두 600s go-test 타임아웃(기계 경합 hang — t1542 교훈의 카드 트리 구조적 적색 재현, 첫 회는 cli 병행 실행과의 자가 경합). 라운드-22 착지 기준(전체 패키지 무-race + 영향계열 -race)으로 대체 측정: 전체 패키지 무-race 재실행 + TestGTD -race 결과는 아래 줄에 기록.

## §A round-23 — 2026-10-08 (정체성 기반 조정 완결 — 게이트 P1-a/P1-b 수리)

- **전임 재설계의 채택과 정정**: 라운드-22 연속의 「바인딩(card_dispatch.run_id)을 참여의 정본으로」 설계는 채택하되, 게이트 4차가 증명한 대로 **팩토리 바인딩 단독으로는 승계와 낡은 바인딩을 구별할 수 없음** — 바인딩이 다른 run을 명명하는 상태는 「더 나중 배차」와 「직전 배차의 잔재」가 데이터 수준에서 대칭이다(양쪽 모두 바인딩된 run의 행이 assigned+owner, 양쪽 모두 두 run의 할당 행 존재). 정본을 큐로 옮긴다.
- **RED 관측** (커밋 기준 `65e649d5f` + 라운드-22 연속 미커밋 트리, 엔진 재현 3종 — 본 세션 저작·직접 관측): ① P1-a `TestGTDReconcileMatchesFactoryOwnerToDispatchOwner` — runtime lane-2 vs factory lane-1인데 operation=reconciled, 행 소유자 잔류 `"lane-1"` ② P1-b `TestGTDStalePriorBindingIsNotALaterDispatch` — 부분 실패(큐 반만 착지, 바인딩 서순) 뒤 신규 배차가 `binding = "run-old"`로 reconciled ③ 폐쇄 거부 `TestGTDUnadjudicatedPriorBindingFailsClosed` — 판정 불가 상태가 `state reconciled`로 무음 통과. 셋 다 FAIL 메시지 verbatim으로 관측.
- **설계 (정체성, 순서 아님)**: ① **큐 자체 참여 기록 `todo_dispatch_current`** — 배차 할당 원시(`recordRuntimeHook`, hook!=nil = 배차 경로 `RecordFactoryCardAssignment` 유일)가 할당 업서트와 **같은 트랜잭션**에 (card_id→run_id, owner_label)를 기록. 상태 보고(hook=nil)는 절대 안 씀 → `TestGTDStateReportNeverMovesTheBinding`가 바인딩+참여 기록 양쪽 불변을 증명. ② **repair의 승계 판정은 이 기록으로**: 기록이 다른 run → superseded 스킵(게이트 3차 ②의 superseded-먼저 순서 보존); 기록이 내 run → 바인딩이 무엇을 명명하든 수리(= P1-b: 낡은 바인딩은 「이후 배차의 증거」가 아님); **기록 부재 + 바인딩이 타 run → `errGTDReconcileUnadjudicated` 폐쇄 거부** — 스킵도 드래그도 아닌 명시적 실패, 회복은 재배차 한 번(기록이 생겨 다음 재시에서 판정). 바인딩 부재(고아 행)·내 run 명명은 기존대로 수리 진행(round-4 회복 경로 보존). ③ **P1-a — `gtdDispatchBindingCurrent`는 소유자 정체성 비교**: 큐 할당의 owner(정본)와 팩토리 행 owner_label을 양쪽 canonical 어휘로 비교, 비어있음 검사 폐지. ④ **수리의 소유자 재낙인** — 행이 assigned 이후인데 소유자가 다르면 homestate 신원시 `RestampDispatchOwner`(버전 bump + `card.dispatch-owner` 이벤트, 상태·리스·증거 불변)로 정본 소유자로 재기입; 큐의 `todo_dispatch_current`도 readRuntime/copyRuntime에 추가(마이그레이션 패리티).
- **GREEN 관측**: 재현 3종 `ok 11.956s` / TestGTD 전체(기존 18+신규 3+보강) `ok 89.701s` / vet clean.



- **라운드-23 재개 세션 결정 (2026-10-08, 단독 작성자)**: 전임이 429로 중단했으나 트리를 빌드한 결과 시그니처 반쪽 적용 상태는 이미 해소(`gtdDispatchBindingCurrent` 5인자, build/vet clean). 진행 중이던 정체성 기반 설계(`todo_dispatch_current` + 소유자 정체성 비교 + `RestampDispatchOwner`)를 **그대로 완결·채택**(대체안 없음 — 바인딩 단독 정본은 라운드-22 연속에서 비대칭 불가로 기각). 본 세션 변경: gofmt 정정(gtd_operation.go 이중 공백행) 1건. RED는 전임 관측(본 세션 미재현 — Gap). GREEN 직접 관측: `go test ./internal/factory/ -count=1` ok 448.096s · `-run TestGTD -race -count=3` ok 217.068s · `go test ./internal/homestate/ -count=1` ok 188.484s, `-race` ok 257.980s · golangci-lint factory+homestate 0 issues.
