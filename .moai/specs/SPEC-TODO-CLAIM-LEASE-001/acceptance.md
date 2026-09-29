# acceptance.md — SPEC-TODO-CLAIM-LEASE-001

> **Observable 표기 규약 (iteration-2, D1 대응)**: 각 Observable은 (a) 이 트리에 실존하는 테스트의 정확한 함수명을 인용하거나, (b) 존재하지 않으면 `[NEW]` 마커 + run phase가 생성할 정확한 함수명을 명시한다. 모든 `-run` 선택자는 매칭 테스트 함수 수 > 0을 선행 단언한다 — 0-매치(`no tests to run`, exit 0)는 실패로 간주한다(verification-completeness §1.1). 실존 확인 근거: `grep -rn "^func <이름>(" internal/` 2026-09-29 실행, 9개 이름 전부 적중.

## §D AC Matrix

### AC-TCL-001 — 클레임 성공

- **Given** `queued` 카드가 1장 이상 있는 백로그
- **When** `moai todo claim` 실행
- **Then** `seq` 오름차순 기준 가장 오래된(최소 `seq`) `queued` 카드가 `picked`로 전환되고 `picked_by`, `lease_expires_at`(≈ now+15m, DefaultFactoryLeaseDuration), `picked_at`이 스탬프되며 출력에 id + text 접두사 + 만료 시각이 포함된다. (순서 근거 컬럼 명시: `seq` — `added_at` 아님)
- **Observable**: `[NEW]` `TestTodoClaim_Success` — `go test ./internal/cli -run '^TestTodoClaim_Success$' -count=1` (스윕>0 선행 단언)

### AC-TCL-002 — 클레임 경합 (exactly-one-wins + raced byte-identity)

- **Given** M장의 `queued` 카드와 N개의 동시 claim 액터 — **N = 2×M, 즉 카드당 최소 2개 액터가 동일 카드를 경합**하는 배치(N ≤ M·1-per-card 배치는 경합을 만들지 못하므로 금지)
- **When** N개 claim이 병렬 실행
- **Then** (a) 카드마다 정확히 1개 claim만 승리하고, 패자는 store 수준 raced 표시를 받는다. (b) **부분 쓰기 검출기 = pre/post 전체 테이블 튜플 비교**(`pragma_table_info` 기반 전 컬럼·전 행 순서쌍 비교) 기준 부분 쓰기 0건. (c) **raced 경로에서 레코드가 byte-identical로 유지된다** — 패자 반환 직후 튜플이 경합 전과 동일.
- **Observable**: `[MODIFY]` 기존 `TestConcurrencyStress`(internal/kanban/backlog_concurrency_test.go:135)에 클레임 경합 케이스 확장 — `go test ./internal/kanban -run '^TestConcurrencyStress$' -count=1`

### AC-TCL-003 — non-queued 거부

- **Given** 상태가 `picked`/`hold`/`dropped`인 카드
- **When** `moai todo claim`
- **Then** 거부되고(구별되는 종료 코드/메시지), 백로그 파일은 byte-identical로 유지된다(사전/사후 파일 해시 비교). 미지의 미래 상태도 default-refuse로 거부된다.
- **Observable**: `[NEW]` `TestTodoClaim_RefusesNonQueued` — `go test ./internal/cli -run '^TestTodoClaim_RefusesNonQueued$' -count=1`

### AC-TCL-004 — 만료 반환(reclamation)

- **Given** 임대 만료가 지난 `picked` 카드
- **When** claim 계열 연산(claim/renew/reclaim)이 실행 (expiry-first)
- **Then** 카드가 `queued`로 복귀하고 **cleared-field 집합 = 임대 필드(`picked_by`, `lease_expires_at`) + pick-time `spec_id`**가 클리어되며, 출력이 id + 이전 보유자를 밝히고, 이어지는 claim이 그 카드를 클레임할 수 있다.
- **Observable**: `[NEW]` `TestBacklogClaim_ReclaimsExpired` — `go test ./internal/kanban -run '^TestBacklogClaim_ReclaimsExpired$' -count=1`

### AC-TCL-005 — 파싱 불가 만료 = EXPIRED

- **Given** `lease_expires_at`이 파싱 불가 값인 레코드
- **When** 만료 판정
- **Then** EXPIRED로 판정된다(C4 — factory `LeaseExpired` 기본값; slot-lease의 보수적 반대 기본값을 따르지 않는다).
- **Observable**: `[NEW]` `TestBacklogClaim_UnparseableExpiryExpired` — `go test ./internal/kanban -run '^TestBacklogClaim_UnparseableExpiryExpired$' -count=1`

### AC-TCL-006 — 임대 연장(renew) + 만료 후 renew

- **Given** 보유자 레이블이 일치하는 live 임대
- **When** `moai todo claim --renew <id>`
- **Then** `lease_expires_at`만 임대 기간만큼 연장되고 다른 필드는 불변이며, 보유자 레이블이 다른 호출은 거부된다. **제3 팔(renew-after-expiry)**: 이미 만료된 임대에 `--renew`는 연장하지 않고 expiry-first 반환(AC-TCL-004 집합)을 먼저 수행한 뒤 거부/재선택 경로로 진입한다.
- **Observable**: `[NEW]` `TestTodoClaim_Renew` — `go test ./internal/cli -run '^TestTodoClaim_Renew$' -count=1` · renew-after-expiry 팔: `[NEW]` `TestBacklogClaim_RenewAfterExpiry` — `go test ./internal/kanban -run '^TestBacklogClaim_RenewAfterExpiry$' -count=1`

### AC-TCL-007 — 스키마 freeze + 수렴

- **Given** fresh DB와 소급(upgraded) DB
- **When** freeze 테스트 + 마이그레이션 순서 테스트 실행
- **Then** 양 DB가 동일한 11컬럼 튜플로 수렴하고, 버전 정합 → 소급 순서가 보장된다(t1310 guard).
- **Observable**: 기존 테스트, 이름 불변 — `go test ./internal/kanban -run '^(TestSchemaFreezeRecordsTransitionStamps|TestBacklogV1ToV2MigrationRoundTrip|TestTodoHistoryAddsNoSchemaChange)$' -count=1` (internal/kanban/backlog_schema_freeze_test.go:49, backlog_hold_migration_test.go:268 실존 확인)

### AC-TCL-008 — golden gate byte-identity

- **Given** freeze된 golden fixture (엔진 우회 seed)
- **When** `todo list --json` 실행
- **Then** 출력이 fixture와 byte-identical — 새 컬럼은 JSON 직렬화에서 제외된다.
- **Observable**: 기존 테스트, 이름 불변 — `go test ./internal/cli -run '^TestTodoListJSON_GoldenByteIdentity$' -count=1` (internal/cli/todo_json_golden_test.go:65 실존 확인)

### AC-TCL-009 — 레인 거버넌스 (3팔)

- **Given** 레인 세션 (`MOAI_KANBAN_ID` 설정 환경)
- **When/Then 3팔**:
  1. 베어 `moai todo claim` → REQ-SD-015 거부(텍스트가 MCP 표면과 동일).
  2. 운영자/리드 세션에서 `--lane <label>` 붙은 claim → 지명 레인에 귀속된 클레임으로 허용.
  3. **`factoryLaneRefusal()`이 참을 보유한 세션에서 `--lane <label>` 붙은 claim → 기존 거부 텍스트로 거부된다**(플래그형도 레인 셀프 클레임이 아님 — 거부 술어는 호출자 신원을 가정하지 않으며, 팔 1과 동일 거부 텍스트).
- **Observable**: `[NEW]` `TestTodoClaim_LaneGovernance` — `go test ./internal/cli -run '^TestTodoClaim_LaneGovernance$' -count=1` (거부 텍스트 동일성 단언 포함; `factoryLaneRefusal()` 정의 internal/cli/factory_card.go:59, 적용 지점 internal/cli/todo.go:360 실존 확인)

### AC-TCL-010 — no-card 종료 코드

- **Given** `queued` 카드가 0장인 백로그
- **When** `moai todo claim`
- **Then** 전용 no-card 종료 코드로 종료하고 오류가 아닌 메시지를 낸다.
- **Observable**: `[NEW]` `TestTodoClaim_NoCardExit` — `go test ./internal/cli -run '^TestTodoClaim_NoCardExit$' -count=1`

### AC-TCL-011 — live-lease 가드 (REQ-TCL-007, iteration-2 신설)

- **Given** live 임대(만료 전)를 보유한 카드
- **When/Then 3팔**:
  1. 다른 claim이 그 카드를 대상으로 실행 → 거부, 레코드 byte-identical, 임대 필드 불변.
  2. live 임대 카드 A·B가 있을 때 A의 만료 반환(reclaim)이 실행 → B의 임대 필드(`picked_by`/`lease_expires_at`) 불변 — 타 보유자 임대 교차 변이 0건.
  3. **unpick carve-out**: 운영자 `unpick`은 live-leased 카드를 여전히 `queued`로 되돌린다(기존 의미론 유지).
- **Observable**: `[NEW]` `TestBacklogClaim_LiveLeaseGuard` — `go test ./internal/kanban -run '^TestBacklogClaim_LiveLeaseGuard$' -count=1` · unpick 팔은 기존 가드 인용(이름 불변): `TestTodoUnpick_RevertsPickedToQueued`(internal/cli/todo_test.go:1062), `TestTodoUnpick_RefusalsLeaveFileUntouched`(:1107) — `go test ./internal/cli -run '^(TestTodoUnpick_RevertsPickedToQueued|TestTodoUnpick_RefusalsLeaveFileUntouched)$' -count=1`

### AC-TCL-012 — human 표면 컬럼 노출 (REQ-TCL-010, iteration-2 신설)

- **Given** 클레임/반환 이력이 있는 백로그 (임대 컬럼 값 존재)
- **When** `todo list` / `todo history` (human 출력) 실행
- **Then** human 표면이 `picked_by`/`lease_expires_at`을 노출한다 — JSON 직렬화는 REQ-TCL-014(AC-TCL-008)대로 계속 제외된다. human-output 노출 vs JSON-exclusion은 표면이 다른 상호 재조정 조항이다(REQ-TCL-010 제2절 ↔ REQ-TCL-014).
- **Observable**: `[NEW]` `TestTodoClaim_ListHistoryExposesLeaseColumns` — `go test ./internal/cli -run '^TestTodoClaim_ListHistoryExposesLeaseColumns$' -count=1`

## Traceability (AC ↔ REQ mapping)

- AC-TCL-001 maps REQ-TCL-005
- AC-TCL-002 maps REQ-TCL-004, REQ-TCL-006
- AC-TCL-003 maps REQ-TCL-007, REQ-TCL-011 (partial — 상태 거부 경로가 live-leased 거부를 부분 운반)
- AC-TCL-004 maps REQ-TCL-007, REQ-TCL-009, REQ-TCL-010
- AC-TCL-005 maps REQ-TCL-009
- AC-TCL-006 maps REQ-TCL-007, REQ-TCL-008 (partial — 보유자 불일치 거부가 교차 보유자 불변을 부분 운반)
- AC-TCL-007 maps REQ-TCL-001, REQ-TCL-002, REQ-TCL-003, REQ-TCL-015
- AC-TCL-008 maps REQ-TCL-014
- AC-TCL-009 maps REQ-TCL-013
- AC-TCL-010 maps REQ-TCL-012
- AC-TCL-011 maps REQ-TCL-007
- AC-TCL-012 maps REQ-TCL-010

## §D.x Edge Cases

- **v1-without-stamps DB 읽기 허용**: `picked_at`조차 없는 필드 DB에서 열기/읽기가 실패하지 않는다 (`columnExpr` NULL fallback, REQ-TCL-003; AC-TCL-007의 마이그레이션 스위트가 간접 검증).
- **ErrBacklogBusy 경로**: SQLITE_BUSY → `ErrBacklogBusy` 매핑 경로에서 recovery가 DB를 삭제/덮어쓰지 않는다 (refuse-to-operate).
- **구버전 바이너리 downgrade**: 구 바이너리의 첫 큐 쓰기가 임대 컬럼 데이터를 유실함을 문서화 — 런타임 방어가 아니라 known limitation으로 기록 (research.md §4 파생).
- **만료 후 renew**: AC-TCL-006 제3팔 — expiry-first 반환 선행, Observable `TestBacklogClaim_RenewAfterExpiry`.
- **unpick 보존**: 반환(reclamation)은 **임대 필드 + pick-time `spec_id`**를 클리어하는 기존 `unpick` 의미론을 일반화하며 — 운영자 수동 `unpick`은 그대로 유지된다 (REQ-TCL-007; AC-TCL-011 제3팔, 기존 `TestTodoUnpick_*` 가드).

## Quality Gates

1. **스키마 freeze 재기록 통과** — 재기록된 튜플 + freeze-header 문서화 항목으로 AC-TCL-007 green (기존 테스트 이름 불변).
2. **golden byte-identity** — `todo list --json` fixture와 byte-identical (AC-TCL-008, 기존 `TestTodoListJSON_GoldenByteIdentity`).
3. **경합 exactly-one-wins** — 카드당 N≥2 stress에서 승자 1명/카드, pre/post 전체 튜플 비교 부분 쓰기 0, raced 경로 byte-identity (AC-TCL-002).
4. **레인 거부 텍스트 동일성** — CLI와 MCP 표면의 거부 텍스트가 동일, 베어형과 `--lane`형(factoryLaneRefusal 참 보유 시) 모두 (AC-TCL-009).
5. **LSP/빌드** — `go build ./...` 0 error, `go vet` 0 finding (run-phase LSP 게이트와 정합).

## Definition of Done

- REQ-TCL-001..015 전건이 AC-TCL-001..012 + edge case로 추적 가능(Traceability)하게 검증됨 — REQ-TCL-007은 AC-TCL-003(부분)/006(부분)/011(전담), REQ-TCL-010은 AC-TCL-004(출력 라인)/012(휴먼 표면)로 이원 커버.
- §E Verification Commands 전체가 green (plan.md §E) — [NEW] 테스트는 run phase가 정확한 함수명으로 생성 후.
- M1-M5 마일스톤 산출물 커밋 + `@MX:ANCHOR`/`@MX:WARN [TID:TX]`/`@MX:NOTE` 주입 완료 (plan.md §D).
