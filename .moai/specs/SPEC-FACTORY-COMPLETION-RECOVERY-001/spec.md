---
id: SPEC-FACTORY-COMPLETION-RECOVERY-001
title: "Factory completion-judgment and recovery repair: leader receipt gate, auto-done recheck, expired-lease reaper, safe reassign, stalled-wait watchdog, CI completion reader, near-dup disclosure, relation disposition ledger"
version: "0.1.0"
status: draft
created: 2026-10-06
updated: 2026-10-06
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "factory,todo,auto-done,receipt-gate,lease-reaper,reassign,watchdog,ci-reader,near-dup,gtd-relations"
tier: M
---

## §A — History

- **2026-10-06** — plan-phase v0.1.0 authored from card t1538 ("팩토리 완료판정·회수 로직 수리", Class C). 근거는 배차 감사 리포트(`.moai/reports/dispatch-audit-2026-10-06-codex.md`, 코드 기준 main@ec13872f3)와 lane-1의 스코프 노트(`.moai/reports/t1538/plan-scope.md`). 본 트리 base는 local develop `a158b4b5f`이며 main은 앞서 있음(t1513 슬롯 술어 수리 `de388878e`·t1528 게이트 스킵 `ec13872f3` 미포함). 카드 항목 (5)는 스코프 노트에 따라 재범위됨 — 무소유자 행은 t1513이 이미 소멸시키므로 본 SPEC은 소유자가 있으나 신호가 없는 행만 다룬다. Tier M, 요구 16건, 기준 16건.

## §B — Problem

### B.1 — 근거와 측정 기준

배차 감사 리포트(codex thread `01a10d8e-6444-78a3-9f22-ef008cee503b`)는 정상 진행과 자동 복구가 끊기는 조합 8건을 제안했다. 리포트의 행 번호는 main@ec13872f3 기준이고 본 트리는 그 이전 base(`a158b4b5f`)에서 출발하므로, 모든 앵커를 본 트리에서 재확인했다(B.2). 제공된 리더 DB 실측(assigned=28·completed=10·picked=5·leased=4·plan=1, dead_letters=0, tmdcqc 만료 임대 4건, t1453 무임대 picked)은 리포트 인용값이며 본 SPEC은 이를 재측정하지 않았다 — 그 값들은 동기부여이고, 요구의 근거는 본 트리에서 관측한 코드 상태다.

### B.2 — Anchor verification (본 트리 a158b4b5f 기준)

리포트의 행 번호를 그대로 쓰지 않고 본 트리에서 전수 재배치했다. 14개 앵커 중 13개가 리포트와 동일 선상에서 확인됐고 1개(todo_analysis.go:74)는 소재가 부분 이동했다.

| 리포트 앵커 | 본 트리 위치 | 판정 |
|---|---|---|
| todo_autodone.go:311 — auto-done 후보 게이트 | `internal/cli/todo_autodone.go:311` — `facts := factory.AutoDoneFacts{...}` | 일치. 리더 증거검토 입력 없음(구조체 `internal/factory/autodone_scan.go:108`, receipt 필드 부재 관측) |
| todo_autodone.go:385 — ID 존재만 재확인 | `internal/cli/todo_autodone.go:385` — `rec.Items[i].ID == outcomes[k].id` 동등 비교만 하고 archive | 일치 |
| todo.go:1104 — done 게이트 | `internal/cli/todo.go:1104` — `rec.ArchiveCard(id)` | 일치. `--expect`·`requireLanded`는 있으나 리더 승인 검증 없음 |
| todo.go:851 — 발급 출력 | `internal/cli/todo.go:851` — `fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", item.ID, pos)` | 일치. stdout은 id+위치뿐, near-dup 정보 없음 |
| todo.go:1403 — unpick | `internal/cli/todo.go:1403` — `case factory.BacklogStatePicked:` | 일치 |
| factory_run_retire.go:333 — active만 reconcile | `internal/homestate/factory_run_retire.go:333` — `f.classifyRuns(ctx, opts, "active")` | 일치. 디렉터리 명시 필요: 리포트는 디렉터리 미기재였고 `internal/factorymsg/factory_run_retire.go`(142행, leader peer identity 용도)는 별개 파일 — 수리 대상이 아니다 |
| factory_run_retire.go:393 — retirement은 runs만 갱신 | `internal/homestate/factory_run_retire.go:393` — `UPDATE runs SET status='retired' ...` | 일치. 카드 행은 그대로 |
| card_transition.go:267 — T19/T20 명시적 거부 | `internal/homestate/card_transition.go:267` — `isReservedEdge` 거부, 문구가 "CI verdict reader that admits it is owned by F3" 명시 | 일치. 술어 본체 `:173` — `pushed→ci-green`, `ci-green→done` |
| card_transition.go:353 — 만료 시 owner 유지 | `internal/homestate/card_transition.go:351-353` — `applyLeaseExpiry`: holder·expiry 클리어(`:360`), `OwnerLabel` 불변 | 일치 |
| card_transition.go:451 — 동일 owner만 임대 | `internal/homestate/card_transition.go:451` — `label != cur.OwnerLabel` 거부 | 일치 |
| card_record.go:138 — LeaseExpired 빈 임대 | `internal/homestate/card_record.go:137-138` — `LeaseExpiresAt == ""`이면 만료 아님(무기한 대기) | 일치 |
| todo_analysis.go:74 — near-dup 임계 0.80 | `internal/cli/todo_analysis.go:74`는 `case factory.BacklogMatchNear:`(기록 지점). 상수는 `internal/factory/backlog_analysis.go:33`(`BacklogNearDuplicateThreshold = 0.80`), 적용점 `internal/cli/todo_analysis.go:188` | **부분 이동** — 리포트가 가리킨 행은 판정 기록 분기이고 임계값 상수는 factory 패키지에 있다 |
| backlog_gtd_schema.go:47 — gtd_relations 스키마 | `internal/factory/backlog_gtd_schema.go:47` | 일치. 관계 테이블에 disposition 열 없음(gtd_items에는 `:36`에 존재) |
| gtd_relation.go:154 — pair upsert | `internal/factory/gtd_relation.go:154` — `ON CONFLICT(subject_id,object_id,kind) DO UPDATE` | 일치 |

보충 관측(수리 설계의 입력):

- 전체 run 분류 원시가 이미 존재 — `ClassifyRuns`(exported, status="") `internal/homestate/factory_run_retire.go:161-163`. reaper는 이를 재사용한다.
- 판정문 reader 선례 — `internal/homestate/card_evidence_readers.go:93-207`(`readAuditVerdict`·`ParseAuditVerdictFile`·`admitVerdictFile`, `audited_sha` 바인딩)와 `internal/auditverdict` 패키지. CI 완료 reader가 따를 패턴이다.
- GTD item disposition 쓰기 경로 존재 — `internal/factory/gtd_clarify.go:86`. "처분 쓰기 코드 전무"는 금지된 결론이며 결여된 것은 **관계** disposition이다.
- "receipt"라는 단어는 `internal/cli/todo.go:818/:873`의 스토어 발급 receipt(전혀 다른 개념)에 이미 쓰인다. 신규 게이트는 이와 구별되는 명칭을 쓴다.

### B.3 — 무엇이 깨져 있는가

1. **완료 게이트에 리더 증거검토 입력이 없다(M1-1).** `todo done`(todo.go:1104)과 auto-done(todo_autodone.go:311)은 리더가 증거를 읽었다는 확인을 요구하지 않는다. lane 실행 제한은 있으나 판정 주체의 검토는 없어, 조건만 맞으면 카드가 리더 검토 없이 닫힌다.
2. **auto-done 잠금 재검증이 ID뿐이다(M1-2).** 스냅샷과 lock 사이에 hold/drop/edit된 행도 ID만 같으면 닫힌다(todo_autodone.go:385).
3. **만료 임대 회수가 active run만 본다(M2-3).** `ReconcileActiveRuns`(factory_run_retire.go:332-333)는 active만 분류한다. 비활성 run의 만료 행은 `--run` 직접 접근 시에만 lazy expiry가 돌아 자동 회수가 없다. 회수 불가가 아니라 **자동 회수 부재**다.
4. **사라진 owner의 재배정 경로가 없다(M2-4).** 만료는 owner를 유지하고(card_transition.go:353) 임대는 동일 owner만 받는다(:451). owner가 사라진 assigned 카드에 직접 재배정 edge가 없다.
5. **구동 중인 행의 신호 대기가 무기한이다(M2-5, 재범위).** 소유자가 있으나 다음 신호(assigned→leased 등)가 오지 않는 행은 대기 기록도 검토 기한도 없다(card_record.go:138의 빈 임대 의미론과 결합). 무소유자 행은 t1513이 이미 처리한다 — 본 항목은 아니다.
6. **예약된 CI 완료 경로가 미구현이다(M2-6).** pushed→ci-green→done은 T19/T20으로 명시적 거부 중이다(card_transition.go:267). 정상 완료가 끊긴다.
7. **near-dup 발급이 무표시다(M3-7).** 임계 0.80 이상 인접 카드는 finding으로만 기록되고(todo_analysis.go:74) 발급 출력은 id+위치뿐이다(todo.go:851). 발급자는 중복을 모른 채 진행한다.
8. **관계 처분 기록이 없다(M3-8).** gtd_relations에 처분 필드가 없고(backlog_gtd_schema.go:47) pair upsert만 있다(gtd_relation.go:154). 유지·병합·기각 결정과 근거가 남지 않는다.

### B.4 — 대가

assigned 28·picked 5 중 정상 진행이 끊긴 행이 수동 복구(abandon·unpick)로만 처리된다. 완료 판정의 증거 연쇄가 리더 검토에서 끊겨 있어, 자동 완료가 SPEC 본문과 어긋난 카드를 닫을 이론적 여지가 검토 없이 열려 있다. t1453 같은 무임대 picked 행과 tmdcqc 만료 임대 4건이 회수 자동화 부재의 실측 사례다(리포트 인용, 재측정 안 함).

## §C — Goal

완료 판정에 리더 증거검토를 묶고(리더 receipt 게이트), auto-done의 잠금 재검증을 스냅샷 전면 비교로 강화하며, 만료 임대의 자동 회수·사라진 owner의 안전 재배정·구동 중 행의 신호 대기 관리·예약된 CI 완료 경로를 구현하고, near-dup 발급 표시와 관계 처분 ledger로 발급·관계 품질을 닫는다. 어떤 회수·복구 경로도 카드를 완료 처리하지 않는다.

## §D — Requirements (GEARS)

16 requirements. REQ-FCR-001..016. 각 요구는 소속 milestone과 검증된 앵커를 이름에 담는다.

### D.1 — M1: 완료 게이트 (카드 항목 1·2)

- **REQ-FCR-001** (Ubiquitous) — The completion of a factory-linked card shall require a 리더 승인 receipt bound to the 세 바인딩 값(카드 UUID·현 factory version·증거 해시). receipt 발급은 리더 경로만이 가진다.
- **REQ-FCR-002** (Event-driven) — When a factory-linked card의 **완료 표면**이 닫히려 할 때 the done gate shall verify the receipt against the 세 바인딩 값 and refuse on absent, stale(카드의 현 factory version과 불일치), or hash-mismatched receipts; 거부는 stderr에 이유를 낸다. 완료 표면은 둘 다다: (a) backlog archive 경로(internal/cli/todo.go:1104의 `rec.ArchiveCard(id)` 앞), (b) homestate 완료 전이 — T20(`ci-green→done`, receipt가 예약 edge의 admission)과 T18(`merged-local→done`, `guardNoRemote` 포함 — 원격 없는 저장소라도 receipt 없이 완료되지 않는다). (b)의 검증은 `FactoryDB.Transition` 안에서 한다 — backlog 경로만 고치면 T18·T20이 게이트 밖으로 남는다.
- **REQ-FCR-003** (Event-driven) — When the auto-done 스캔이 후보 facts를 조립할 때(internal/cli/todo_autodone.go:311, 구조체 `internal/factory/autodone_scan.go:108`), facts shall carry the receipt state and an unverified receipt shall not close the card — skip/downgrade 처리되고 오류로 스캔 전체를 끊지 않는다.
- **REQ-FCR-004** (State-driven) — While auto-done이 lock 안에서 행을 재발견할 때(internal/cli/todo_autodone.go:385 — 현재는 ID 동등만 확인), it shall compare the snapshot's UUID·본문·state·SPEC·landing against current values and skip the close on any mismatch, downgrading with the existing inconclusive reason.
- **REQ-FCR-005** (Unwanted) — The receipt gate shall not accept a receipt whose issuer is the card's **performing owner**(수행 lane/worker가 자기 작업의 승인을 발급한 receipt — 독립 축은 수행자≠승인자다) and shall not accept a receipt whose factory version is older than the card's current version. 단일 리더 구성에서 리더가 receipt를 발급하고 같은 리더가 `todo done`·auto-done을 실행하는 것은 정상 경로다 — 발급자와 완료 실행자의 동일성은 거부 사유가 아니다.

### D.2 — M2: 회수 (카드 항목 3·4·5·6)

- **REQ-FCR-006** (Event-driven) — When the 리더 유지관리 경로가 만료 임대 reaper를 호출할 때, the reconcile shall cover **모든** run의 만료 행(현 `internal/homestate/factory_run_retire.go:333`의 active-only 분류와 달리, 전체 분류 원시 `ClassifyRuns` :161-163 재사용), re-verify each expiry inside one transaction, and reclaim each expired row with the 기존 `applyLeaseExpiry` 의미론(card_transition.go:351-353 — assigned 복귀, mid-merge는 blocked, worktree 미접촉).
- **REQ-FCR-007** (Unwanted) — The reaper shall never delete rows(retireRun의 보존 의미론, factory_run_retire.go:393의 UPDATE-only 준수) and never select among survivors(positive-gate fail-closed 원칙 준수). `retirable`(:158)의 **owner 생존 조건(OwnerDead)은 run retirement 전용이다** — reaper는 이를 상속하지 않는다: 살아 있는 리더의 run에서 만료된 lane 임대도 `applyLeaseExpiry` 의미론(card_transition.go:351-353)으로 회수한다. `applyLeaseExpiry`에는 생존 조건이 없으며, 사망 확인이 필요한 것은 소유자 교체(REQ-FCR-008)와 run retirement뿐이다.
- **REQ-FCR-008** (State-driven) — While a card holds no valid lease and the prior owner's 종료 증거(dead process identity 또는 리더 확인) is verified, an operator-path reassign edge shall move `OwnerLabel` to a new registered owner in a version-checked transition(`updateCardRow` 기대버전 패턴, card_transition.go:364) — 현 구조의 두 가드(:353 owner 유지, :451 동일 owner 임대)를 우회하지 않고 **새 edge**로만.
- **REQ-FCR-009** (Event-driven, 재범위 항목 5) — When a driven row(owner 설정, 임대 부재 또는 유효)가 다음 신호를 기다리며 검토 기한을 넘길 때, the system shall record the waiting signal + review deadline and request 리더 재판정 — 자동 완료 금지, 무소유자 행 자동 해제 금지(t1513 소관). 상태 기록은 card_record.go:138의 빈 임대 의미론을 변경하지 않는다.
- **REQ-FCR-010** (Event-driven) — When a card in `pushed` requests `pushed→ci-green`(T19), the reserved-edge refusal(card_transition.go:267, 술어 :173) shall be admitted only by a CI verdict reader following the 기존 판정문 admission 패턴(card_evidence_readers.go:93-207, `audited_sha` 바인딩) — reader 없는 T19은 계속 거부된다. **T20(`ci-green→done`)은 reader가 열지 않는다** — reader는 `ci-green`까지만 진행하며, done 전이는 REQ-FCR-002(b)의 receipt 검증이 `FactoryDB.Transition`의 예약 edge 경로 안에서 열어준다.
- **REQ-FCR-011** (State-driven) — While the CI reader가 T19 증거를 심사할 때, it shall admit only evidence bound to the **정확히 push 시 기록된 commit SHA** — 같은 브랜치의 다른 SHA 증거는 거부한다.

### D.3 — M3: 표시·처분 (카드 항목 7·8)

- **REQ-FCR-012** (Event-driven) — When `todo add`가 발급 시 분석에서 `BacklogMatchNear`를 기록할 때(internal/cli/todo_analysis.go:74, 임계 `internal/factory/backlog_analysis.go:33` = 0.80, 적용점 :188), the issuance shall print to **stderr** the 관련 카드 id·점수·판정 안내; stdout stays the bare "id position" machine line(internal/cli/todo.go:851, stdout 규율 주석 :816-824 준수).
- **REQ-FCR-013** (State-driven) — While gtd_relations가 유지·병합·기각 판정의 대상일 때, the relation store shall persist a per-relation disposition(keep/merge/dismiss) + 근거, through the 기존 upsert 경로(internal/factory/backlog_gtd_schema.go:47 스키마 확장, gtd_relation.go:154 upsert) — item disposition 쓰기 경로(gtd_clarify.go:86)는 건드리지 않는다.

### D.4 — Cross-cutting

- **REQ-FCR-014** (Unwanted) — No lane surface shall mint a receipt, invoke the reaper, or reassign a card — receipt 발급, reaper 호출, reassign edge, 리더 재판정 기록은 전부 리더/운영자 경로다(레인의 큐 변경 금지와 동일 경계).
- **REQ-FCR-015** (Ubiquitous) — Every new multi-row reconcile shall run inside one transaction with in-transaction 재검증, and every new transition edge shall be version-checked.
- **REQ-FCR-016** (Unwanted) — No recovery path(reaper, watchdog, reassign) shall complete a card, and the CI reader는 `ci-green`까지만 진행한다 — done 전이를 여는 것은 M1 receipt 게이트뿐이다.

## §E — Constraints

### E.1 — 경계

- t1533(같은 파일군 후속, hub 대기 결함 4건)은 본 카드에서 제외 — 파일 접촉을 최소화해 직렬 인계를 깨끗하게 유지한다(스코프 노트).
- `internal/cli/factory_card.go`의 슬롯 술어는 t1513 소관 — 본 SPEC은 항목 (5)에 필요한 범위를 넘어 그 의미론을 다시 만지지 않는다.
- "receipt" 명칭 충돌: todo.go:818/:873의 스토어 발급 receipt와 구별되는 이름을 쓴다(예: leader approval receipt → `LeaderApproval` 계열).

### E.2 — Non-functional

- 트랜잭션 안전(REQ-FCR-015), 버전 검사 전이(REQ-FCR-015), 레인 경계(REQ-FCR-014), 무자동완료(REQ-FCR-016)는 모든 milestone에 걸리는 하드 제약이다.
- 스키마 확장(gtd_relations disposition, watchdog 대기 기록)은 기존 행과의 하위호환을 유지한다 — 기존 행의 빈 disposition은 "미판정"으로 읽힌다.

## §F — Verification sample

`go test ./internal/cli -run 'TestLeaderReceiptGate'`와 `go test ./internal/homestate -run 'TestExpiredLeaseReaperAllRuns'`가 각각 M1·M2의 대표 판정이다. 전체 기준은 acceptance.md — `[no tests to run]`은 통과가 아니라 미측정이다(§0).

## §G — Out of Scope

### Out of Scope — t1533의 hub 대기 결함 4건

- 같은 파일군의 후속 카드가 소유한다. 본 SPEC은 접촉 파일을 그 경계 안에 둔다.

### Out of Scope — 자동 완료 경로

- 감사 리포트가 "자동 완료 처리로 풀면 안 된다"고 명시한 대로, watchdog·reaper·reassign이 카드를 닫는 설계는 금지다(REQ-FCR-016).

### Out of Scope — 무소유자·무임대 행의 슬롯 해제

- t1513(`de388878e`)이 `factorySerialSlotHeld`에 소유자 조건을 넣어 소멸시켰다(스코프 노트 §t1513 경계). 본 카드의 항목 (5)는 재범위되어 소유자가 있으나 신호가 없는 행만 다룬다.

### Out of Scope — `factory_card.go` 슬롯 술어의 의미론 변경

- 술어는 t1513의 착지물이다. 본 SPEC은 그 의미론을 다시 쓰지 않는다.

### Out of Scope — GTD item disposition 쓰기 경로

- `gtd_clarify.go:86`의 item 처분 쓰기는 이미 존재하며 건드리지 않는다. 신규 대상은 **관계** disposition뿐이다.

### Out of Scope — 리포트의 "전이 불가" 표현 수용

- 리포트 스스로 과표현이라 인정한 부분(운영자 abandon 탈출구 존재)을 본 SPEC은 결함으로 취급하지 않는다 — 수리 대상은 정상 진행·자동 복구가 끊기는 조합이다.

## §H — Dependencies and Risks

- **main 흡수**: 본 트리는 local develop 기준이며 main이 앞서 있다 — 통합 전 흡수로 `de388878e`(t1513 수리)·`ec13872f3`(t1528)이 본 카드 트리에 들어온다. t1513은 슬롯 술어 주변 테스트를 바꿨을 수 있어 병합 창에서 술어 인접 테스트 충돌 확인이 필수다(plan.md §H).
- **직렬 슬롯 임대 지연**: lane-1은 `factory next --card t1538`이 `refused serial-slot`(설치 빌드에 t1513 수리 미반영)으로 거부됐다 — 영수증 경로(t1522식, 리더 배차=권한)로 진행 중. 본 SPEC의 receipt 게이트(M1)가 이 경로의 도구적 뒷받침이다.
- **리포트 기준 차이**: 감사 리포트는 main@ec13872f3 기준 — 본 SPEC의 모든 앵커는 B.2에서 본 트리 기준으로 재확인됐다. 흡수 후 앵커 재확인은 run-phase 첫 커밋 전 수행 항목이다.
