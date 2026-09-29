---
id: SPEC-TODO-CLAIM-LEASE-001
title: "Todo 카드 원자 클레임 API — CAS+임대"
version: "1.0.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: GOOS
priority: P1
phase: "v3.3.0 target"
module: "internal/kanban, internal/cli"
lifecycle: spec-anchored
tags: "todo, claim, lease, cas, sqlite, concurrency"
depends_on: [SPEC-TODO-RUNTIME-STORE-001]
tier: M
---

# SPEC-TODO-CLAIM-LEASE-001 — Todo 카드 원자 클레임 API (CAS+임대)

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| v1.0.0 | 2026-09-29 | card t1342 initial draft — plan-phase Phase 10 산출 (DP1 승인안 반영) |

## §A Overview

### 배경과 동기

리드 세션의 측정 보고(`.moai/reports/autonomy-bottleneck-proposal-20260929.md`)가 결함을 기록했다: 소유자 없는 `picked` 카드 11장, t810은 19일간 picked 상태로 방치, 픽업 보호는 flock뿐. 카드를 집어 가는 주체가 기록되지 않으니 "누가, 언제까지" 작업을 보유 중인지 판정할 방법이 없고, 자율 픽업(t1338 계열)의 안전 기반이 되지 못한다.

현재 백로그 스키마에는 소유·임대 개념이 전혀 없다 — `backlogItemsTableColumns`(internal/kanban/backlog_sqlite.go:122)는 `seq, id, text, added_at, spec_id, state CHECK(queued|picked|dropped|hold), landing`에 `picked_at`/`dropped_at` 소급 컬럼을 더한 9컬럼이며, `picked_by|lease_expires|PickedBy|LeaseExpires`를 non-test `internal/kanban` 전체에서 grep하면 0건이다.

### 계보(lineage)

SPEC-TODO-RUNTIME-STORE-001(t648, completed)이 소유권 CAS를 명시적으로 후속 child로 이관했다(Exclusions: "UUID backfill, 소유권 CAS, slot/resume 이동, 자동 완료/receipt, Graph, 웹 시각화는 후속 child다"). 본 SPEC은 그 후속이며, 이미 검증된 재사용 대상은 internal/homestate의 version-CAS 전이 + 임대 하트비트 패턴(SPEC-FACTORY-RECORD-001 소유)이다.

### 핵심 설계 결정 (DP1 승인, 2026-09-29)

| 결정 | 내용 |
|------|------|
| C1 | 마이그레이션 = additive `ensureColumn`, 스탬프 bump 없음(schema "2" 유지). freeze 테스트 재기록은 still required |
| C2 | 클레임 원자성 = `BacklogStore.Mutate` 콜백 안에서 freshly-loaded 레코드에 대한 상태 술어(state predicate) — flock이 CAS 원시체. version 컬럼 없음, Mutate 밖 엔진 접근 금지 |
| C4 | 파싱 불가한 임대 만료 = EXPIRED(factory 기본값) |
| C5 | 반환 감사 = 사람이 읽는 claim/reclaim 출력 라인(id + 이전 보유자). events 테이블 없음 |
| C6 | `--lane <label>` = 운영자/리드가 공급하는 레인 레이블(레인을 지칭하는 인자). flag-form guard extension inside todoRefuseLaneMutation — REQ-SD-015 bare-refusal semantics unchanged: `--lane` 없는 레인 세션 claim뿐 아니라 `factoryLaneRefusal()`이 참을 보유한 상태의 `--lane`형 claim도 기존 거부 텍스트로 거부된다(거부 술어는 호출자 신원을 가정하지 않는다). 레인 셀프 클레임 거버넌스는 t1338 소유 |

## §C 요구사항 (GEARS)

### §C.1 스키마 (M1)

- **REQ-TCL-001** (Ubiquitous) — The backlog store shall carry nullable TEXT `picked_by` and `lease_expires_at` (RFC 3339) columns on the `items` table. 컬럼은 `ensureColumn` additive 경로로 `dropped_at` 뒤에 추가되며, `backlogItemsTableColumns`에는 절대 진입하지 않고, `archived_items` INSERT에도 동일하게 미러링된다.
  - Reference: internal/kanban/backlog_sqlite.go:122, ~:536-571 · AC: AC-TCL-007
- **REQ-TCL-002** (Event-driven) — When a database is opened, `ensureSchema` shall run the additive lease retrofit only after version reconciliation(t1310 역순 위험 — rebuild가 소급 컬럼을 조용히 삭제하는 사례).
  - Reference: internal/kanban/backlog_sqlite.go:442-456 · AC: AC-TCL-007
- **REQ-TCL-003** (State-driven) — While a pre-retrofit database lacks the columns, every read consumer shall tolerate their absence via the `columnExpr` NULL fallback — 읽기 경로는 pure/fail-open을 유지한다.
  - Reference: REQ-TLE-006 pointer discipline (`BacklogItem` 필드는 `*string` + `omitempty`, absence = nil/SQL NULL) · AC: AC-TCL-007

### §C.2 클레임 원자성 (M2)

- **REQ-TCL-004** (Ubiquitous) — Every claim/renew/reclaim write shall route through `BacklogStore.Mutate`(flock 직렬화 chokepoint). Mutate 밖의 엔진 접근은 금지된다.
  - Reference: internal/kanban/backlog_store.go:759-819; clobber hazard vs `writeRecordArchive` internal/kanban/backlog_migrate.go:422-515 (Mutate 밖 row-level UPDATE는 delete-and-rewrite와 양방향 silent lost update) · AC: AC-TCL-002
- **REQ-TCL-005** (Event-driven) — When `moai todo claim [--lane <label>]` runs, the CLI shall claim the oldest `state=='queued'` item as a single compare-and-set inside the Mutate callback(콜백 안의 freshly-loaded 레코드에 대한 상태 술어가 곧 CAS이며, flock이 원자성을 공급한다) — `picked_by`, `lease_expires_at`(now+DefaultFactoryLeaseDuration), `picked_at`을 스탬프하고 id + text 접두사 + 만료 시각을 출력한다.
  - Reference: internal/config/defaults.go:55 (DefaultFactoryLeaseDuration = 15m) · AC: AC-TCL-001
- **REQ-TCL-006** (Event-driven) — When a claim loses a race(state가 더 이상 queued가 아님), the store shall refuse with a distinct raced indication and keep the record byte-identical — factory arm-(c) 의미론.
  - Reference: internal/cli/factory_card.go:358-379 · AC: AC-TCL-002, AC-TCL-003
- **REQ-TCL-007** (State-driven) — While an item's lease is live, the item shall not be claim-selectable and no claim-family operation shall mutate another holder's lease. 운영자 `unpick`의 수동 회수는 유지된다.
  - AC: AC-TCL-004

### §C.3 임대 생명주기 (M3)

- **REQ-TCL-008** (Event-driven) — When `claim --renew <id>` runs, the store shall extend `lease_expires_at` by the lease duration and change no other field; a holder-label mismatch shall be refused(RenewLease 패턴 — ErrLeaseHolder, expiry-first, version-checked extend의 의미론 이식).
  - Reference: internal/homestate/card_transition.go:364-410 (패턴 원전) · AC: AC-TCL-006
- **REQ-TCL-009** (Event-driven) — When a claim-family operation(본 SPEC에서 claim/renew/reclaim 셋으로 열거) runs, it shall run expiry-first; a lapsed lease shall return the item to `queued` with its lease fields and pick-time `spec_id` cleared(generalized `unpick` — cleared-field 집합은 임대 필드 + pick-time `spec_id`). 파싱 불가한 만료는 EXPIRED로 판정한다(C4 — factory 기본값).
  - Reference: applyLeaseExpiry internal/homestate/card_transition.go:339; unparseable→EXPIRED internal/homestate/card_record.go:123-134 · AC: AC-TCL-004, AC-TCL-005
- **REQ-TCL-010** (Ubiquitous) — Every reclamation shall be surfaced as human-visible claim/reclaim output lines(id + 이전 보유자 — C5 확정 감사 표면). events 테이블은 만들지 않는다. The human `todo list`/`history` output shall expose the new columns — JSON 직렬화는 REQ-TCL-014대로 계속 제외된다(human-output 노출과 JSON-exclusion은 표면이 다른 상호 재조정 조항).
  - AC: AC-TCL-004

### §C.4 CLI 표면 + 거버넌스 (M4)

- **REQ-TCL-011** (Ubiquitous) — The `claim` verb shall use positive enumeration — `picked`/`hold`/`dropped`는 거부하고, 미래의 미지 상태도 거부한다(default: refuse).
  - Reference: `todo next` pick-gate 패턴, REQ-THS-011/012 (internal/cli/todo.go:1034-1145) · AC: AC-TCL-003
- **REQ-TCL-012** (Event-driven) — When no eligible card exists, `claim` shall exit with a dedicated no-card exit code and emit a non-error message.
  - Reference: `factoryNextNoCardExit` internal/cli/factory_card.go:503 · AC: AC-TCL-010
- **REQ-TCL-013** (Event-driven) — When a lane session claims without `--lane`, the existing REQ-SD-015 refusal shall apply unchanged; when the claim carries `--lane <label>`, it shall be attributed to the named lane as an operator/lead-side claim; **when `factoryLaneRefusal()` holds, the `--lane <label>` form shall ALSO be refused with the existing governance refusal text** — 플래그형은 운영자/리드 측 전용이며 거부 술어는 호출자 신원을 가정하지 않는다(flag-form guard extension inside todoRefuseLaneMutation; REQ-SD-015 bare-refusal semantics unchanged — 레인 셀프 클레임 거버넌스는 t1338 소유). MCP 미러는 동일한 결정을 수행하며 거부 텍스트가 CLI 표면과 동일해야 한다.
  - Reference: internal/cli/todo.go:358-372 (todoRefuseLaneMutation); internal/cli/mcp_todo.go:57; `--lane` 인자 선례 internal/cli/goal.go:275, gtd.go:272 · AC: AC-TCL-009
- **REQ-TCL-014** (Ubiquitous) — The `todo list --json` output shall stay byte-identical to the frozen golden fixture — 새 컬럼은 JSON 직렬화에서 제외된다.
  - Reference: AC-TST-012, internal/cli/todo_json_golden_test.go · AC: AC-TCL-008
- **REQ-TCL-015** (Event-driven) — When the schema freeze test runs, it shall pass with re-recorded tuples and a freeze-header entry documenting the additive lease extension.
  - Reference: internal/kanban/backlog_schema_freeze_test.go · AC: AC-TCL-007

## §D Out of Scope

### Out of Scope — 우선순위/병렬 메타데이터

- 카드 우선순위·병렬 실행 메타데이터 컬럼 및 이를 활용한 선택 정책은 t1338(P2)이 소유한다. 본 SPEC은 임대에 필요한 최소 컬럼만 추가한다.

### Out of Scope — 레인 셀프 클레임 / REQ-SD-015 개정

- 레인 세션 스스로 todo 큐를 변이할 권한 부여(REQ-SD-015 개정·carve)는 t1338의 거버넌스 질문이다. 본 SPEC은 `--lane <label>` 인자형(operator/lead 공급 레이블)만 추가하며, `factoryLaneRefusal()` 참 보유 시의 `--lane`형도 REQ-TCL-013 제3절에 따라 기존 거부 텍스트로 거부되므로 레인 셀프 클레임 권한 부여가 아니다.

### Out of Scope — 백로그 events 테이블

- 반환/전이 이력을 위한 events 테이블 도입은 freeze 테이블 집합 변경 + 세 번째 감사 어휘 수반 — Tier L 결정으로, t1338이 필요 시 deliberately 추가한다.

### Out of Scope — 다섯 번째 `leased` 상태

- `state` CHECK에 `leased` 추가는 v3 rebuild(REQ-THS 계열 전면 재건)을 강제한다. 본 SPEC은 임대 상태를 별도 컬럼으로 표현하며 state enum을 확장하지 않는다.

### Out of Scope — slot-lease 어휘 통합

- `slot_lease.go`의 세션 앵커드 슬롯 임대와 카드 임대의 어휘/의미 통합은 본 SPEC 밖이다. 두 시스템은 독립적으로 유지된다.

### Out of Scope — 백로그 version 컬럼

- per-row version 컬럼 도입은 C2에서 기각되었다 — flock이 이미 원자성을 공급하며, Tier M 범위를 초과하고 flock과 중복된다.

## §E Delta Markers (brownfield 표면)

### [EXISTING]

- `internal/kanban/backlog_store.go` — `BacklogStore.Mutate` flock chokepoint (:759-819), `LoadPure` lock-free 읽기 (:663-694)
- `internal/cli/todo.go` — todo pick-gate (`todo next` :1034-1145), `todoRefuseLaneMutation` (:358-372), verb 등록 목록 (:304)
- `internal/cli/mcp_todo.go` — MCP 미러의 lane 변이 거부 재적용 (:57)

### [MODIFY]

- `internal/kanban/backlog_sqlite.go` — `ensureSchema`/`ensureColumn` 경로에 임대 컬럼 소급 추가 (:442-456, ~:536-571)
- `internal/kanban/backlog_store.go` — `BacklogItem`에 `*string` PickedBy/LeaseExpiresAt 포인터 필드 + claim/renew/reclaim store 연산
- `internal/kanban/backlog_migrate.go` — `archived_items` INSERT 컬럼 목록 미러링 (`writeArchive` ~:337-401, INSERT :381) + items 경로 (`writeRecordArchive` :422-515, items INSERT :482)
- `internal/cli/todo.go` — `claim` verb 등록 + todoRefuseLaneMutation flag-form guard extension(`factoryLaneRefusal()` 참 보유 시 `--lane`형도 기존 거부 텍스트로 거부; REQ-SD-015 bare-refusal semantics unchanged) + `list`/`history` human 표면 렌더링(REQ-TCL-010)
- `internal/cli/todo_disclosure.go` — pick 경로에 cross-checkout routing + `discloseStaleLocalStores` 재사용 편성
- `internal/cli/mcp_todo.go` — MCP `todo_claim` 미러 (동일 결정 + 거부 텍스트 동일성)
- `internal/kanban/backlog_schema_freeze_test.go` — 컬럼 튜플 재기록 + freeze-header 항목 (기존 테스트 이름 불변)
- `internal/kanban/backlog_concurrency_test.go` — TestConcurrencyStress에 클레임 경합 케이스 확장
- `internal/kanban/backlog_hold_migration_test.go` — 마이그레이션 순서/수렴 케이스 추가 (기존 스위트 확장, 이름 불변)
- `internal/cli/todo_json_golden_test.go` — golden gate 유지 확인 (기존 테스트 이름 불변, 신규 컬럼 JSON 제외 재단언)

### [NEW]

- store 계층 Claim/RenewLease/reclaimExpired 연산 (Mutate 콜백형)
- `moai todo claim [--lane <label>] [--renew <id>]` CLI verb
- claim 테스트 집합(run phase가 생성) — internal/kanban: `TestBacklogClaim_ReclaimsExpired`, `TestBacklogClaim_LiveLeaseGuard`, `TestBacklogClaim_UnparseableExpiryExpired`, `TestBacklogClaim_RenewAfterExpiry` · internal/cli: `TestTodoClaim_Success`, `TestTodoClaim_RefusesNonQueued`, `TestTodoClaim_Renew`, `TestTodoClaim_LaneGovernance`, `TestTodoClaim_NoCardExit`, `TestTodoClaim_ListHistoryExposesLeaseColumns`
