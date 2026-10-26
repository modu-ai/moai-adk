---
id: SPEC-FACTORY-COMPLETION-RECOVERY-001
title: "Factory completion-judgment and recovery repair: leader receipt gate, auto-done recheck, expired-lease reaper, safe reassign, stalled-wait watchdog, CI completion reader, dead-owner run consolidation, near-dup disclosure, relation disposition ledger"
version: "0.2.0"
status: draft
created: 2026-10-06
updated: 2026-10-26
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "factory,todo,auto-done,receipt-gate,lease-reaper,reassign,watchdog,ci-reader,run-lifecycle,dead-owner-run,near-dup,gtd-relations"
tier: L
---

## §A — History

- **2026-10-06** — plan-phase v0.1.0 authored from card t1538 ("팩토리 완료판정·회수 로직 수리", Class C). 근거는 배차 감사 리포트(`.moai/reports/dispatch-audit-2026-10-06-codex.md`, 코드 기준 main@ec13872f3)와 lane-1의 스코프 노트(`.moai/reports/t1538/plan-scope.md`). 본 트리 base는 local develop `a158b4b5f`이며 main은 앞서 있음(t1513 슬롯 술어 수리 `de388878e`·t1528 게이트 스킵 `ec13872f3` 미포함). 카드 항목 (5)는 스코프 노트에 따라 재범위됨 — 무소유자 행은 t1513이 이미 소멸시키므로 본 SPEC은 소유자가 있으나 신호가 없는 행만 다룬다. Tier M, 요구 16건, 기준 16건.
- **2026-10-26** — v0.2.0, 카드 t1963(Class C). M2에 **팩토리 run 수명주기 재설계(M2-B — 죽은 소유자 run "좀비" 누적 제거)**를 추가했다. 근거는 리더의 읽기 전용 분석(`.moai/reports/t1963/zombie-analysis.md`, 미추적 카드 증거 — 권위 아님)이고, 인용은 전부 본 트리 HEAD `c4784db4c2`에서 다시 읽어 확인했다(§B.5, 교정 목록은 research.md §2). 요구 7건(REQ-FCR-017..023)·기준 9건(AC-FCR-017..025)을 더해 요구 23건·기준 25건이 되어 Tier M 천장(16/16)을 넘으므로 **Tier L로 상향**하고 design.md·research.md를 추가했다(분할은 카드 지시로 금지 — 새 SPEC·새 장부 금지). `interview.decision_gate: on`이라 결정 색인 `decision-index.md`를 함께 둔다. M1은 이미 run-phase에서 착지 중이며(progress.md round-1~25) 이 개정은 M1의 요구·기준 본문을 바꾸지 않는다 — REQ-FCR-014에 한 문장(운영자 표면 정의)만 보탰다. 이 개정으로 plan 아티팩트 해시가 바뀌므로 plan-audit 재판정이 필요하다.

## §B — Problem

### B.1 — 근거와 측정 기준

배차 감사 리포트(codex thread `01a10d8e-6444-78a3-9f22-ef008cee503b`)는 정상 진행과 자동 복구가 끊기는 조합 8건을 제안했다. 리포트의 행 번호는 main@ec13872f3 기준이고 본 트리는 그 이전 base(`a158b4b5f`)에서 출발하므로, 모든 앵커를 본 트리에서 재확인했다(B.2). 제공된 리더 DB 실측(assigned=28·completed=10·picked=5·leased=4·plan=1, dead_letters=0, tmdcqc 만료 임대 4건, t1453 무임대 picked)은 리포트 인용값이며 본 SPEC은 이를 재측정하지 않았다 — 그 값들은 동기부여이고, 요구의 근거는 본 트리에서 관측한 코드 상태다.

### B.2 — Anchor verification (본 트리 a158b4b5f 기준)

리포트의 행 번호를 그대로 쓰지 않고 본 트리에서 전수 재배치했다. 14개 앵커 중 13개가 리포트와 동일 선상에서 확인됐고 1개(todo_analysis.go:74)는 소재가 부분 이동했다.

| 리포트 앵커 | 본 트리 위치 | 판정 |
|---|---|---|
| todo_autodone.go:311 — auto-done 후보 게이트 | `internal/cli/todo_autodone.go:311` — `facts := factory.AutoDoneFacts{...}` | 일치. 리더 증거검토 입력 없음(구조체 `internal/factory/autodone_scan.go:108`, receipt 필드 부재 관측) |
| todo_autodone.go:385 — ID 존재만 재확인 | `internal/cli/todo_autodone.go:386` — `rec.Items[i].ID == outcomes[k].id` 동등 비교만 하고 archive | 일치(행번호 1 차이 — 본 트리 재측정치로 기재) |
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

> **HEAD `c4784db4c2` 재확인(v0.2.0)**: M1 착지로 위 표의 `factory_run_retire.go`·`card_transition.go`·`card_record.go` 앵커가 다른 행으로 옮겨졌다. 주장은 그대로 성립하며, 이동한 위치는 research.md §2.3에 있다. run-phase M2 첫 커밋 전 재확인 항목(plan.md §H)은 그 표를 입력으로 쓴다.

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

### B.5 — run 수명주기: 죽은 소유자 run이 쌓인다 (M2-B, 카드 t1963)

동기: 운영자가 `moai cc -l`에서 활성 run 5개(죽은 리더 4·산 리더 1)를 제시받았다(리더 분석 인용 — 본 SPEC은 라이브 DB를 읽지 않았고 재측정하지 않는다). 아래 구조는 본 트리 HEAD `c4784db4c2`의 소스에서 확인했다.

1. **소유권이 한 프로세스에 묶인다.** run 행은 리더의 pid와 프로세스 시작 지문을 기록하고(`internal/homestate/runtime.go:44-47` INSERT, 기록 호출 `internal/cli/factory.go:707-727` `recordFactoryRunStart`), 소유자 판정은 pid+지문을 함께 본다(`internal/homestate/factory_run_retire.go:139-155` `ClassifyOwnerWith`). 리더를 다시 띄우면 새 프로세스라 이전 run은 구조적으로 "dead"가 된다.
2. **재채택이 좁다.** 리더 진입의 재채택은 명시적 resume 인자(`internal/cli/factory_run_adopt.go:126-155` `strictResumeCarrierSeen`)·새로 발급된 id·`--factory-run` 부재를 모두 요구하고(`:163-166` `leaderEntryRunAdoption`, 게이트 문장 `:165`), 후보 탐색은 **후보가 정확히 하나이면서 자기 스탬프로 dead 판정**일 때만 채택한다(`internal/homestate/factory_run_adopt.go:72-96`, 조건 `:92`). 둘 이상이면 경고만 남기고 새 id를 발급한다(`internal/cli/factory_run_adopt.go:77-81`).
3. **카드 행이 run마다 따로 있다.** `cards`의 기본 키는 `(run_id, card_id)`다(`internal/homestate/factory.go:51-83`, 키 `:83`). `RecordPicked`는 현재 run에서만 행을 찾고 없으면 새로 넣는다(`internal/homestate/card_picked.go:130-154`, INSERT `:166`). run 간 읽기는 단계·증거·worktree를 **복사만** 한다(`card_metadata.go:284-307` `PreviousCardRow`, 호출 `internal/cli/factory_card.go:1449`; `card_worktree.go:211-228` `PreviousCardWorktree`). 그래서 이전 run의 행은 비종결 상태로 남는다.
4. **은퇴가 막힌다.** 운영자 은퇴(`moai factory runs --retire`, `internal/cli/factory_handoff_recover.go:117-166` → `RetireRunIfDead` `factory_run_retire.go:381-406` → `retireRun` `:422-462`)는 비종결 카드가 있으면 거부한다(`:438-444`, 술어 `factory_run_cards.go:71-73` — 비종결 = done·failed·abandoned 밖, 상수 `factory_run_adopt.go:21`). 합류 프롬프트의 `r<n>`도 `Dead && InProgress==0 && runs>=2`일 때만 제시된다(`internal/cli/factory_run_select.go:86`, 파싱 `:202-224`). 자동 조정(`internal/factorymsg/store.go:392-446` `ResolveActiveRun` → `ReconcileActiveRuns` `factory_run_retire.go:339-375`)은 `runHasLiveRecords`(`factory_run_adopt.go:31-39`)가 참이면 보존하며, 이 술어는 **`workers` 행이 하나라도 있으면** 참이다 — 그 행의 pid가 죽었는지는 보지 않는다. 리더 종료 시 은퇴하는 경로는 없다. 은퇴는 `runs`만 갱신한다(`factory_run_retire.go:445`, `:485`).
5. **결과.** 활성 run이 2개 이상 남으면 `ResolveActiveRun`이 `AmbiguousRunsError`(`AMBIGUOUS_FACTORY`)로 실패한다(`internal/factorymsg/store.go:432-443` — 조정 호출 `:432`, 반환 `:442`; 타입 `:453`; CLI 센티널 `internal/cli/factory.go:517`).

본 트리에서 관측한 결함 상태(RED, acceptance.md §0 증거 장부 L1): 죽은 소유자 run 2개에 비종결 카드를 심은 fixture에서 후보 탐색은 `adopt="" candidates=2`, 조정 뒤 활성 run은 `remaining=2`다. 죽은 pid의 worker 행만 가진 run은 `active`로 남고, 새 run에서 임대한 카드의 이전 run 행은 `state="run"`으로 남는다.

## §C — Goal

완료 판정에 리더 증거검토를 묶고(리더 receipt 게이트), auto-done의 잠금 재검증을 스냅샷 전면 비교로 강화하며, 만료 임대의 자동 회수·사라진 owner의 안전 재배정·구동 중 행의 신호 대기 관리·예약된 CI 완료 경로를 구현하고, near-dup 발급 표시와 관계 처분 ledger로 발급·관계 품질을 닫는다. 어떤 회수·복구 경로도 카드를 완료 처리하지 않는다. 또한 run 수명주기를 고쳐(M2-B) 리더를 다시 띄워도 활성 run이 하나로 모이고, 죽은 run에 비종결 중복 행이 남지 않게 한다 — 카드 행을 지우지 않고, 모든 다중 행 이동은 한 트랜잭션 안에서 한다.

## §D — Requirements (GEARS)

23 requirements. REQ-FCR-001..023. 각 요구는 소속 milestone과 검증된 앵커를 이름에 담는다.

### D.1 — M1: 완료 게이트 (카드 항목 1·2)

- **REQ-FCR-001** (Ubiquitous) — The completion of a factory-linked card shall require a 리더 승인 receipt bound to the **네 바인딩 값**(카드 UUID·**receipt가 승인한 factory run id**·현 factory version·증거 해시). receipt 발급은 리더 경로만이 가진다. run 바인딩이 없으면 같은 카드가 새 run에서 다시 수행될 때 UUID는 유지되고 version은 별도 행에서 새로 시작해, 동일 (카드·version·증거 SHA) 조합이 run 간 재현된다 — 이전 run의 승인이 새 run에서 재사용된다(측정: cards 스키마에 동일 조합 2-run 생성 가능, run 간 카드 재사용 기존 테스트 통과).
- **REQ-FCR-002** (Event-driven) — When a **factory-linked** card의 **완료 표면**이 닫히려 할 때(factory run 기록이 존재하는 카드만 — REQ-FCR-001과 동일 대상) the done gate shall verify the receipt against the 네 바인딩 값 and refuse on absent, stale(카드의 현 factory version과 불일치), hash-mismatched, or **cross-run**(receipt의 run id가 카드의 현 run과 불일치) receipts; 거부는 stderr에 이유를 낸다. **비팩토리 카드는 대상 밖이다** — factory run 기록이 없는 일반 카드의 기존 완료 동작(검증 없는 close)은 불변이며, 이는 미검증 상태가 아니라 적용 대상 아님을 구분한 것이다. 검증은 **닫히는 시점**에 한다 — 발급·스캔 시점의 검증만으로는 그 사이의 factory 전이가 version을 올려 stale receipt가 통과한다. 완료 표면은 셋 다다: (a) backlog archive 경로 — 수동 done(internal/cli/todo.go:1104의 `rec.ArchiveCard(id)` 앞 — 수동 done에도 auto-done과 동일한 직전 재검증·직렬화가 적용된다) **및 `todo --auto` 사이클(internal/cli/todo_auto.go:331의 `r.ArchiveCard(card.ID)` — REQ-THS-012의 picked 긍정 나열 가드는 불변)**, (b) homestate 완료 전이 — T20(`ci-green→done`, receipt가 예약 edge의 admission)과 T18(`merged-local→done`, `guardNoRemote` 포함 — 원격 없는 저장소라도 receipt 없이 완료되지 않는다). (b)의 검증은 `FactoryDB.Transition` 안에서 한다 — backlog 경로만 고치면 T18·T20이 게이트 밖으로 남는다.
- **REQ-FCR-003** (Event-driven) — When the auto-done 스캔이 **factory-linked** 후보의 facts를 조립할 때(internal/cli/todo_autodone.go:311, 구조체 `internal/factory/autodone_scan.go:108`), facts shall carry the receipt state and an unverified receipt shall not close the card — skip/downgrade 처리되고 오류로 스캔 전체를 끊지 않는다. factory-linked가 아닌 후보는 receipt 상태 자체를 조립하지 않는다(기존 완료 동작 불변 — `TestTodoAutoDone_LiveFilter` 계열이 지키는 동작).
- **REQ-FCR-004** (State-driven) — While auto-done이 lock 안에서 행을 재발견할 때(internal/cli/todo_autodone.go:386 — 현재는 ID 동등만 확인), it shall compare the snapshot's UUID·본문·state·SPEC·landing against current values and skip the close on any mismatch, downgrading with the existing inconclusive reason. 5항목이 모두 동일해도 **archive 직전에 receipt의 run id·factory version·증거 해시를 같은 lock 안에서 최종 재검증**한다 — 스캔 승인과 닫힘 사이의 factory 전이는 backlog 행과 별도로 version을 올릴 수 있어, 직전 재검증 없이는 stale receipt가 통과한다. 재검증은 동시 factory 전이와 직렬화된다.
- **REQ-FCR-005** (Unwanted) — The receipt gate shall not accept a receipt whose issuer is the card's **performing owner**(수행 lane/worker가 자기 작업의 승인을 발급한 receipt — 독립 축은 수행자≠승인자다) and shall not accept a receipt whose factory version is older than the card's current version. 단일 리더 구성에서 리더가 receipt를 발급하고 같은 리더가 `todo done`·auto-done을 실행하는 것은 정상 경로다 — 발급자와 완료 실행자의 동일성은 거부 사유가 아니다.

### D.2 — M2: 회수 (카드 항목 3·4·5·6)

- **REQ-FCR-006** (Event-driven) — When the 리더 유지관리 경로가 만료 임대 reaper를 호출할 때, the reconcile shall cover **모든** run의 만료 행(현 `internal/homestate/factory_run_retire.go:333`의 active-only 분류와 달리, 전체 분류 원시 `ClassifyRuns` :161-163 재사용), re-verify each expiry inside one transaction, and reclaim each expired row with the 기존 `applyLeaseExpiry` 의미론(card_transition.go:351-353 — assigned 복귀, mid-merge는 blocked, worktree 미접촉).
- **REQ-FCR-007** (Unwanted) — The reaper shall never delete rows(retireRun의 보존 의미론, factory_run_retire.go:393의 UPDATE-only 준수) and never select among survivors(positive-gate fail-closed 원칙 준수). `retirable`(:158)의 **owner 생존 조건(OwnerDead)은 run retirement 전용이다** — reaper는 이를 상속하지 않는다: 살아 있는 리더의 run에서 만료된 lane 임대도 `applyLeaseExpiry` 의미론(card_transition.go:351-353)으로 회수한다. `applyLeaseExpiry`에는 생존 조건이 없으며, 사망 확인이 필요한 것은 소유자 교체(REQ-FCR-008)와 run retirement뿐이다.
- **REQ-FCR-008** (State-driven) — While a card holds no valid lease and the prior owner's 종료 증거(dead process identity 또는 리더 확인) is verified, an operator-path reassign edge shall move `OwnerLabel` to a new registered owner in a version-checked transition(`updateCardRow` 기대버전 패턴, card_transition.go:364) — 현 구조의 두 가드(:353 owner 유지, :451 동일 owner 임대)를 우회하지 않고 **새 edge**로만.
- **REQ-FCR-009** (Event-driven, 재범위 항목 5) — When a driven row(owner 설정, 임대 부재 또는 유효)가 다음 신호를 기다리며 검토 기한을 넘길 때, the system shall record the waiting signal + review deadline and request 리더 재판정 — 자동 완료 금지, 무소유자 행 자동 해제 금지(t1513 소관). 상태 기록은 card_record.go:138의 빈 임대 의미론을 변경하지 않는다. watchdog이 행을 **처음 관측하는 시점**에 `waiting_since`를 소급 기록한다 — 백필이 없으면 이미 정체 중인 기존 행은 기한이 성숙하지 않아 이 요구의 동기 대상에 영원히 도달하지 못한다.
- **REQ-FCR-010** (Event-driven) — When a card in `pushed` requests `pushed→ci-green`(T19), the reserved-edge refusal(card_transition.go:267, 술어 :173) shall be admitted only by a CI verdict reader following the 기존 판정문 admission 패턴(card_evidence_readers.go:93-207, `audited_sha` 바인딩) — reader 없는 T19은 계속 거부된다. **T20(`ci-green→done`)은 reader가 열지 않는다** — reader는 `ci-green`까지만 진행하며, done 전이는 REQ-FCR-002(b)의 receipt 검증이 `FactoryDB.Transition`의 예약 edge 경로 안에서 열어준다.
- **REQ-FCR-011** (State-driven) — While the CI reader가 T19 증거를 심사할 때, it shall admit only evidence bound to the **정확히 push 시 기록된 commit SHA** — T17(`merged-local→pushed`)이 기록하는 pushed tip SHA다(카드의 MergeSHA가 아니며, 일괄 push에서 둘은 다르다). 같은 브랜치의 다른 SHA 증거는 거부하고, push 뒤 remote ref가 이동해도 인정 대상은 저장된 tip SHA로 변하지 않는다.

### D.2b — M2-B: run 수명주기 — 죽은 소유자 run 통합 (카드 t1963)

용어: **죽은 스탬프 run** = 상태 `active`이고 소유자가 자기 스탬프로 `OwnerDead` 판정된 run(`Basis == BasisStamp`). **산 worker** = REQ-FCR-022 기준으로 살아 있는 `workers` 행. **통합 연산** = REQ-FCR-020의 재결속·은퇴 연산.

- **REQ-FCR-017** (Event-driven) — When a leader start(`moai cc -f`·`moai glm -f`) publishes a freshly minted run id without `--factory-run`(`leaderRunIDWasMinted`, `internal/cli/factory_run_adopt.go:34-42` 불변), and at least one 죽은 스탬프 run carries live records, and no active run classifies `OwnerLive`, the launcher shall adopt the newest such run(`created_at` 내림차순, 같으면 `run_id` 내림차순) through the existing `AdoptRun` compare-and-set(`internal/homestate/factory_run_adopt.go:112-145`) instead of minting — resume 인자가 없어도 된다. `OwnerIndeterminate` 또는 `BasisPeer` 후보는 채택하지 않고 기존 경고 줄을 유지한다. codex 리더 run 거부(`:87-93`)와 산 레거시 피어 거부(`:100-105`)는 그대로다.
- **REQ-FCR-018** (Event-driven) — When REQ-FCR-017의 채택이 성공하면, the launcher shall run the 통합 연산 once for every other 죽은 스탬프 run that holds no 산 worker, with the adopted run as the target; 통합 연산이 거부한 원천 run은 활성으로 남고 그 run id와 거부 사유가 stderr에 한 줄씩 남는다. 채택의 compare-and-set에서 진 리더는 통합을 하지 않고 변경 전 동작(발급+경고)으로 돌아간다.
- **REQ-FCR-019** (Event-driven) — When `moai factory next`가 현재 run에 카드를 기록할 때 같은 `card_id`의 비종결 행이 다른 run(활성 또는 은퇴)에 있고 그 run의 소유자가 `OwnerDead`로 판정되면, the lease transaction shall, in the same transaction as `RecordPicked` and after the carry-over read(`factoryPriorCarryCandidate`, `internal/cli/factory_card.go:1448-1458`), move that row to `abandoned` and append one `card.rebound` event carrying `from_run`·`to_run`·이전 state·version·stage·evidence_sha. 소유자가 `OwnerLive`·`OwnerIndeterminate`인 run의 행은 건드리지 않는다.
- **REQ-FCR-020** (State-driven) — While the 통합 연산 runs, it shall execute as one write-locked transaction(factory DSN의 BEGIN IMMEDIATE — `retireRun`과 같은 형태, `factory_run_retire.go:408-418`) that (1) re-reads the source run and proceeds only while it is still `active` under the classified owner stamp, (2) proceeds only while the target is `active`, its owner classifies `OwnerLive`, and its stamp is unchanged since classification, (3) refuses when the source holds a 산 worker, (4) for each non-terminal source card: refuses the whole operation when any third active run whose owner is not `OwnerDead` holds a non-terminal row for that `card_id`; otherwise re-keys the row to the target(`run_id`만 바뀌고 나머지 열은 그대로) when the target holds no row for that `card_id`, or moves the source row to `abandoned` with a `card.rebound` pointer event when the target already holds one, (5) repoints every `card_dispatch` binding that names the source(쓰기 지점 `internal/homestate/leader_approval.go:291`) to the target, and (6) flips the source to `retired` with one `run.retired` event carrying the classification, the basis, `rebound_to`, and the moved/abandoned counts. 어느 단계의 거부든 트랜잭션 전체를 되돌린다.
- **REQ-FCR-021** (Event-driven) — When the operator runs `moai factory runs --retire <id> --rebind-to <live-run>`(`--rebind-to`는 `runs` 명령의 플래그 — decision-index Q1 판정 (a)) or chooses `m<n>` at the lane join prompt, the CLI shall invoke the 통합 연산 with that source and target and print the moved and abandoned card ids. `m<n>`은 해당 행이 `Dead && InProgress>0`이고 `OwnerLive`인 활성 run이 정확히 하나일 때만 제시되며 그 run이 대상이다. `--rebind-to` 없는 `--retire`와 `r<n>`의 의미론은 SPEC-FACTORY-RETIRE-GUARD-001 그대로다(비종결 카드가 있으면 거부, 출력의 기존 줄 불변). 레인 세션(`MOAI_FACTORY_WORKER` 비어 있지 않음)에서 `--rebind-to`는 상태를 바꾸지 않고 거부된다(REQ-FCR-014).
- **REQ-FCR-022** (State-driven) — While `runHasLiveRecords`(`internal/homestate/factory_run_adopt.go:31-39`) evaluates a run, a `workers` row shall count as live unless its pid probes positively dead; a probe that cannot decide counts as live. `workers`에는 프로세스 시작 지문 열이 없으므로(`internal/homestate/factory.go:30-38`) 판정은 pid만으로 한다 — pid 재사용은 죽은 worker를 살아 보이게 할 뿐이라 안전한 방향이다.
- **REQ-FCR-023** (Unwanted) — No run-lifecycle path(REQ-FCR-017..022) shall delete a `cards`, `runs`, or `events` row, move a card out of a run whose owner is not `OwnerDead`, write without the compare-and-set on the stamp it classified, or migrate `leader_approvals` rows — a re-keyed card needs a fresh approval under the target run, because the REQ-FCR-002 cross-run refusal stays in force.

### D.3 — M3: 표시·처분 (카드 항목 7·8)

- **REQ-FCR-012** (Event-driven) — When `todo add`가 발급 시 분석에서 `BacklogMatchNear`를 기록할 때(internal/cli/todo_analysis.go:74, 임계 `internal/factory/backlog_analysis.go:33` = 0.80, 적용점 :188), the issuance shall print to **stderr** the 관련 카드 id·점수·판정 안내; stdout stays the bare "id position" machine line(internal/cli/todo.go:851, stdout 규율 주석 :816-824 준수).
- **REQ-FCR-013** (State-driven) — While gtd_relations가 유지·병합·기각 판정의 대상일 때, the relation store shall persist a per-relation disposition(keep/merge/dismiss) + 근거, through the 기존 upsert 경로(internal/factory/backlog_gtd_schema.go:47 스키마 확장, gtd_relation.go:154 upsert) — item disposition 쓰기 경로(gtd_clarify.go:86)는 건드리지 않는다.

### D.4 — Cross-cutting

- **REQ-FCR-014** (Unwanted) — No lane surface shall mint a receipt, invoke the reaper, or reassign a card — receipt 발급, reaper 호출, reassign edge, 리더 재판정 기록은 전부 리더/운영자 경로다(레인의 큐 변경 금지와 동일 경계). M2-B의 통합 연산(REQ-FCR-020)도 같은 경계 안에 있다 — 리더 진입(REQ-FCR-018)과 운영자 CLI·합류 프롬프트의 운영자 입력(REQ-FCR-021)만 부를 수 있고, 레인 세션의 `--rebind-to`는 거부된다. 합류 프롬프트는 레인 세션이 시작되기 전 운영자가 직접 선택을 입력하는 런처 표면이라 운영자 표면으로 본다.
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

### E.3 — M2-B 가드 (카드 t1963 지시)

- **fixture 전용.** 모든 검증은 테스트 샌드박스(`openSandboxFactory`·`adoptSandbox`처럼 저장소 밖 임시 루트와 홈 변수 정리)에서만 한다. 라이브 factory DB를 읽거나 쓰는 명령·시험은 금지다.
- **설치 바이너리 불변.** `~/go/bin/moai` 등 설치된 moai 바이너리를 교체하지 않는다. 검증은 트리에서 빌드한 코드를 `go test`로만 돌린다.
- **새 장부·새 테이블 없음.** 통합은 기존 `runs`·`cards`·`events`·`card_dispatch` 행 갱신과 기존 `events` 종류 추가(`card.rebound`)만으로 한다. 새 테이블·새 장부 파일을 만들지 않는다.
- **SPEC-FACTORY-RETIRE-GUARD-001 의미론 보존.** `--rebind-to` 없는 은퇴는 비종결 카드가 있으면 계속 거부하고, 그 거부 출력의 기존 줄은 바뀌지 않는다. 그 SPEC의 시험 8건 중 7건은 손대지 않고 통과해야 한다. 허용된 유일한 기존 시험 수정은 `TestFactoryRunsHasNoForceFlag`(`internal/cli/factory_runs_retire_guard_test.go:234-242`)의 허용 플래그 집합 단언을 `[retire]`에서 `[rebind-to retire]`로 넓히는 것이다(decision-index Q1 판정 (a)). 그 시험의 나머지 의도 — 우회 플래그(`--force` 등)가 없다 — 는 그대로 단언된다.
- **동시성·원자성.** 두 리더의 채택 경합은 `AdoptRun`의 스탬프 compare-and-set이 가르고, 진 쪽은 통합하지 않는다. 다른 산 run이 이미 쥔 카드는 옮기지 않는다(연산 전체 거부). 이동 도중 중단은 한 트랜잭션이라 전부 되돌려진다.

## §F — Verification sample

`go test ./internal/cli -run '^(TestLeaderReceiptGate.*)$'`와 `go test ./internal/homestate -run '^TestExpiredLeaseReaperAllRuns$'`가 각각 M1·M2의 대표 판정이고, M2-B의 대표 판정은 `go test ./internal/cli -run '^TestLeaderStartConsolidatesDeadOwnerRuns$'`다. 전체 기준은 acceptance.md — `[no tests to run]`은 통과가 아니라 미측정이다(§0).

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

### Out of Scope — run 소유권 모델 자체의 교체

- 리더 프로세스에 묶인 소유권(pid+지문 스탬프)을 세션·호스트 단위 소유권으로 바꾸는 일은 하지 않는다. M2-B는 기존 스탬프와 `AdoptRun` compare-and-set 위에서 죽은 run을 모으기만 한다.
- 리더 종료 훅에서 run을 은퇴시키는 경로는 만들지 않는다 — 비정상 종료에는 훅이 돌지 않으므로 다음 리더 진입의 통합이 유일한 수렴점이다.

### Out of Scope — 라이브 데이터 정리와 설치 바이너리

- 운영자 머신의 실제 factory DB에 남은 좀비 run(분석이 인용한 tn1p4e·tn564n·tn8jnh·tn8rpj)을 이 카드에서 정리하지 않는다. 착지 뒤 운영자가 새 빌드로 리더를 다시 띄우면 REQ-FCR-017/018이 수렴시킨다.
- 설치된 moai 바이너리를 교체하거나 rc 빌드를 만들지 않는다.

### Out of Scope — 미확정 소유자 run의 자동 통합

- `OwnerIndeterminate`·`BasisPeer` run은 자동으로 채택·통합하지 않는다(REQ-003b 방향 유지). 운영자 경로(`--factory-run`, `--retire`)만 남는다.

### Out of Scope — 리포트의 "전이 불가" 표현 수용

- 리포트 스스로 과표현이라 인정한 부분(운영자 abandon 탈출구 존재)을 본 SPEC은 결함으로 취급하지 않는다 — 수리 대상은 정상 진행·자동 복구가 끊기는 조합이다.

## §H — Dependencies and Risks

- **main 흡수**: 본 트리는 local develop 기준이며 main이 앞서 있다 — 통합 전 흡수로 `de388878e`(t1513 수리)·`ec13872f3`(t1528)이 본 카드 트리에 들어온다. t1513은 슬롯 술어 주변 테스트를 바꿨을 수 있어 병합 창에서 술어 인접 테스트 충돌 확인이 필수다(plan.md §H).
- **직렬 슬롯 임대 지연**: lane-1은 `factory next --card t1538`이 `refused serial-slot`(설치 빌드에 t1513 수리 미반영)으로 거부됐다 — 영수증 경로(t1522식, 리더 배차=권한)로 진행 중. 본 SPEC의 receipt 게이트(M1)가 이 경로의 도구적 뒷받침이다.
- **M2-B 기대 반전(기존 시험)**: `TestAdoptResumedLeaderRunWarnsOnMultipleCandidates`(`internal/cli/factory_run_adopt_test.go:134`, 카드 t1707)는 죽은 스탬프 후보 2개에서 "채택 없음"을 단언한다 — REQ-FCR-017 하에서 기대가 반전된다(최신 run 채택 + 나머지 통합). `TestFactoryRunsHasNoForceFlag`(`internal/cli/factory_runs_retire_guard_test.go:234`)는 허용 플래그 집합 단언만 `[retire]` → `[rebind-to retire]`로 넓어진다(Q1 판정 (a)) — SPEC-FACTORY-RETIRE-GUARD-001 시험 중 유일한 명시적 기대 변경이다. 그 밖의 t1707 시험(`TestFindResumableRun*`, `TestAdoptRun*`, `TestReconcile*`)은 새 선택 로직을 별도 함수로 두면 그대로 통과한다(plan.md §F M2-B).
- **M2-B 근거 문서의 지위**: 리더 분석(`.moai/reports/t1963/zombie-analysis.md`)은 미추적 카드 증거라 권위가 아니다. 인용은 research.md §2에서 소스로 다시 확인했고, 틀리거나 낡은 인용은 그곳에 교정했다.
- **리포트 기준 차이**: 감사 리포트는 main@ec13872f3 기준 — 본 SPEC의 모든 앵커는 B.2에서 본 트리 기준으로 재확인됐다. 흡수 후 앵커 재확인은 run-phase 첫 커밋 전 수행 항목이다.
