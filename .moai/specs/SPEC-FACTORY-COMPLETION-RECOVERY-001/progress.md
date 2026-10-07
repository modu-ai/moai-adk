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

