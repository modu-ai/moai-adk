# spec-compact.md — SPEC-TODO-CLAIM-LEASE-001

> 요약본: 요구사항 + 수용 기준 + 수정 대상 파일 + 제외 항목만 포함 (overview/리서치 제외). 상세는 spec.md/plan.md/acceptance.md. Iteration-2 반영본.

## 요구사항 (REQ-TCL-001..015 — GEARS 요지)

- **§C.1 스키마**: REQ-TCL-001 nullable TEXT `picked_by`/`lease_expires_at`(RFC 3339)을 ensureColumn additive로 `dropped_at` 뒤 추가, `backlogItemsTableColumns` 진입 금지, `archived_items` 미러(backlog_migrate.go `writeArchive` INSERT :381) · REQ-TCL-002 소급은 버전 정합 이후에만 · REQ-TCL-003 pre-retrofit DB는 columnExpr NULL fallback으로 읽기 허용(pure/fail-open).
- **§C.2 클레임 원자성**: REQ-TCL-004 모든 claim/renew/reclaim 쓰기는 `BacklogStore.Mutate` 경유(flock = CAS 원시체, Mutate 밖 엔진 접근 금지) · REQ-TCL-005 `moai todo claim [--lane <label>]` — `seq` 최소순 `queued`를 Mutate 콜백 내 상태 술어 CAS로 클레임, picked_by/lease_expires_at(now+15m)/picked_at 스탬프 + 출력 · REQ-TCL-006(Event-driven) 경합 패배 시 raced 표시 거부, 레코드 byte-identical · REQ-TCL-007 live 임대 아이템은 선택 불가·타 보유자 임대 불변, 운영자 unpick 유지.
- **§C.3 임대 생명주기**: REQ-TCL-008 `--renew <id>`는 기간만 연장, 보유자 불일치 거부 · REQ-TCL-009 claim-family(claim/renew/reclaim 열거) expiry-first — 만료 임대는 queued 복귀 + 임대 필드·pick-time `spec_id` 클리어(generalized unpick), 파싱 불가 만료 = EXPIRED(C4) · REQ-TCL-010 반환 감사 = 사람이 읽는 출력 라인(id + 이전 보유자), events 테이블 없음, human `list`/`history`만 새 컬럼 노출(JSON은 REQ-TCL-014대로 제외).
- **§C.4 CLI + 거버넌스**: REQ-TCL-011 긍정 열거 — picked/hold/dropped·미지 상태 거부 · REQ-TCL-012 카드 없음 → 전용 no-card exit 코드 · REQ-TCL-013 베어 claim(레인) = REQ-SD-015 거부 유지, `--lane <label>` = 운영자/리드 귀속, **`factoryLaneRefusal()` 참 보유 시 `--lane`형도 기존 거부 텍스트로 거부**(flag-form guard extension inside todoRefuseLaneMutation; REQ-SD-015 bare-refusal semantics unchanged — 레인 셀프 클레임 거버넌스는 t1338 소유), MCP 미러 동일 결정+거부 텍스트 동일 · REQ-TCL-014 `todo list --json` golden byte-identity · REQ-TCL-015 freeze 테스트 재기록 튜플 + freeze-header 통과.

## 수용 기준 (AC-TCL-001..012 — 요지)

1. 클레임 성공 — `seq` 최소순 카드, picked + picked_by + lease≈now+15m + picked_at + 출력 (`[NEW]` TestTodoClaim_Success)
2. N-동시 경합 — **카드당 N≥2(N=2×M)**, exactly-one-wins, **pre/post 전체 튜플 비교** 부분 쓰기 0, raced 경로 byte-identity ([MODIFY] TestConcurrencyStress)
3. non-queued 거부 — 파일 byte-identical(해시 비교), 구별되는 종료 (`[NEW]` TestTodoClaim_RefusesNonQueued)
4. 만료 반환 — expiry-first queued 복귀, **임대 필드 + pick-time `spec_id`** 클리어, id+이전 보유자 출력, 재클레임 가능 (`[NEW]` TestBacklogClaim_ReclaimsExpired)
5. 파싱 불가 만료 = EXPIRED (`[NEW]` TestBacklogClaim_UnparseableExpiryExpired)
6. renew — 보유자만 +15m 연장, 타 필드 불변, 비보유자 거부 + 만료 후 renew는 expiry-first 반환 선행 (`[NEW]` TestTodoClaim_Renew · TestBacklogClaim_RenewAfterExpiry)
7. freeze + 수렴 — fresh/upgraded DB 동일 튜플, 마이그레이션 순서(t1310 guard) (기존 TestSchemaFreezeRecordsTransitionStamps · TestBacklogV1ToV2MigrationRoundTrip · TestTodoHistoryAddsNoSchemaChange)
8. golden byte-identity — `todo list --json` fixture 동일 (기존 TestTodoListJSON_GoldenByteIdentity)
9. 레인 거버넌스 3팔 — 베어 claim 거부(텍스트 == MCP), `--lane`형 허용(운영자/리드), **factoryLaneRefusal() 참 보유 시 `--lane`형도 동일 텍스트 거부** (`[NEW]` TestTodoClaim_LaneGovernance)
10. no-card 전용 exit 코드 (`[NEW]` TestTodoClaim_NoCardExit)
11. live-lease 가드 — live 임대 카드 claim 거부+byte-identical, 교차 보유자 임대 불변(A 반환 시 B 불변), unpick carve-out 유지 (`[NEW]` TestBacklogClaim_LiveLeaseGuard · 기존 TestTodoUnpick_RevertsPickedToQueued · TestTodoUnpick_RefusalsLeaveFileUntouched)
12. human 표면 노출 — list/history가 picked_by/lease_expires_at 노출, JSON은 제외 유지 (`[NEW]` TestTodoClaim_ListHistoryExposesLeaseColumns)

## 수정 대상 파일

- **[MODIFY]** `internal/kanban/backlog_sqlite.go` (ensureSchema/ensureColumn 소급), `internal/kanban/backlog_store.go` (BacklogItem 포인터 필드 + Claim/RenewLease/reclaimExpired), `internal/kanban/backlog_migrate.go` (writeArchive ~:337-401 INSERT :381 + writeRecordArchive :422-515 미러), `internal/cli/todo.go` (claim verb + todoRefuseLaneMutation flag-form guard extension + list/history 렌더링), `internal/cli/todo_disclosure.go` (cross-checkout routing + discloseStaleLocalStores 편성), `internal/cli/mcp_todo.go` (todo_claim 미러), 테스트: `backlog_schema_freeze_test.go` (재기록, 이름 불변) · `backlog_concurrency_test.go` (TestConcurrencyStress 확장) · `backlog_hold_migration_test.go` (스위트 확장) · `todo_json_golden_test.go` (gate 유지)
- **[NEW]** store claim 연산군, `moai todo claim [--lane <label>] [--renew <id>]` verb, claim 테스트 집합 — kanban `TestBacklogClaim_{ReclaimsExpired,LiveLeaseGuard,UnparseableExpiryExpired,RenewAfterExpiry}` · cli `TestTodoClaim_{Success,RefusesNonQueued,Renew,LaneGovernance,NoCardExit,ListHistoryExposesLeaseColumns}`

## 제외 항목 (Out of Scope)

- 우선순위/병렬 메타데이터 → t1338(P2)
- 레인 셀프 클레임 승인 / REQ-SD-015 개정 → t1338 (`factoryLaneRefusal()` 참 보유 시 `--lane`형도 거부되므로 본 SPEC이 권한을 부여하지 않음)
- 백로그 events 테이블 (freeze 변경 동반 — Tier L 결정)
- 다섯 번째 `leased` state (v3 rebuild 강제)
- slot-lease 어휘 통합 (두 임대 시스템 독립 유지)
- 백로그 version 컬럼 (flock과 중복, Tier M 범위 초과)
