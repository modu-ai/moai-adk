# plan.md — SPEC-TODO-CLAIM-LEASE-001

## §A Context

- **Card**: t1342 — Todo 카드 원자 클레임 API (CAS+임대)
- **Tier**: M (배차 타이틀 "Tier M" 기준) · **Priority**: P1 · **Phase**: v3.3.0 target
- **Evidence paths**:
  - 제안서: `.moai/state/plan-research-t1342/spec-proposal.md`
  - 리서치: `.moai/state/plan-research-t1342/research.md`
  - 렌즈 보고서: `.moai/state/plan-research-t1342/per-lens-reports.md`
  - 결함 측정: `.moai/reports/autonomy-bottleneck-proposal-20260929.md`
- **Spec 선결 조건**: depends_on SPEC-TODO-RUNTIME-STORE-001 (completed) — 소유권 CAS를 본 SPEC에 이관한 모 SPEC.

### DP1 운영자 결정 (2026-09-29 승인, 구속력)

| ID | 질문 | 결정 |
|----|------|------|
| C1 | 마이그레이션 형태 | additive `ensureColumn`, schema-stamp bump 없음 ("2" 유지). freeze 테스트 재기록은 여전히 요구 |
| C2 | 클레임 원자성 | `BacklogStore.Mutate` 콜백 안 freshly-loaded 레코드에 대한 상태 술어 (flock = CAS 원시체). version 컬럼 없음, Mutate 밖 엔진 접근 없음 |
| C4 | 파싱 불가 만료 | EXPIRED (factory 기본값) |
| C5 | 반환 감사 | 사람이 읽는 claim/reclaim 출력 라인 (id + 이전 보유자). events 테이블 없음 |
| C6 | `--lane <label>` | 운영자/리드 공급 레인 레이블. flag-form guard extension inside todoRefuseLaneMutation — REQ-SD-015 bare-refusal semantics unchanged: `--lane` 없는 레인 세션 claim과 `factoryLaneRefusal()` 참 보유 시의 `--lane`형 claim 모두 기존 거부 텍스트로 거부 (t1338 소유) |

## §B Milestones (결정 가역성 순 — 변경 가능성 높은 결정이 앞)

### M1 (High) — 스키마 컬럼 + freeze 재기록

- **[MODIFY]** `internal/kanban/backlog_sqlite.go:53, :122, :442-456, ~:536-571` — `picked_by`/`lease_expires_at` nullable TEXT 소급 추가. `backlogItemsTableColumns` 진입 금지, `dropped_at` 뒤 ensure 경로로 추가해 fresh/upgraded 물리 순서 수렴. 버전 정합 이후 소급 실행(REQ-TCL-002).
- **[MODIFY]** `internal/kanban/backlog_migrate.go` — `archived_items` INSERT 컬럼 목록에 신규 컬럼 미러링 (`writeArchive` ~:337-401, INSERT :381); items 경로는 `writeRecordArchive` :422-515 (items INSERT :482).
- **[MODIFY]** `internal/kanban/backlog_schema_freeze_test.go` — `wantItemsColumns`/`wantArchivedItemsColumns` 재기록 + freeze-header에 additive 확장 문서화.
- Reference: `.moai/state/plan-research-t1342/research.md` §4

### M2 (High) — store 계층 Claim/RenewLease/reclaimExpired

- **[MODIFY]** `internal/kanban/backlog_store.go:759-819` — Mutate 콜백형 세 연산. `BacklogItem`에 `*string` PickedBy/LeaseExpiresAt 추가 (REQ-TLE-006 pointer discipline). expiry-first 양경로. `DefaultFactoryLeaseDuration` 상수 재사용.
- **[MODIFY]** `internal/kanban/backlog_store.go:83-105` — `BacklogItem` 포인터 필드.
- 패턴 원전(참고 전용, import 아님): `internal/homestate/card_transition.go` :333/:339/:618, `card_record.go:123-134`.
- Reference: research.md §2 (C2 clobber hazard), §6

### M3 (High) — CLI verb + 거버넌스 flag-form guard extension

- **[MODIFY]** `internal/cli/todo.go:358-372` — todoRefuseLaneMutation flag-form guard extension: `factoryLaneRefusal()` 참 보유 시 `--lane <label>`형도 기존 거부 텍스트로 거부(REQ-SD-015 bare-refusal semantics unchanged — 베어 claim 거부 유지, 거부 술어는 호출자 신원 가정 없음). 레인 셀프 클레임 거버넌스는 t1338 소유.
- **[NEW]** `internal/cli/todo_claim.go` (또는 todo.go 내) — `moai todo claim [--lane <label>] [--renew <id>]` verb 등록(:303 목록), no-card 전용 exit 코드.
- **[MODIFY]** `internal/cli/todo_disclosure.go` — cross-checkout routing + `discloseStaleLocalStores` 재사용 — pick 경로 필수 편성품.
- Reference: `factoryNextNoCardExit` internal/cli/factory_card.go:503; `--lane` 인자 선례 goal.go:275/gtd.go:272

### M4 (Medium) — MCP 미러

- **[MODIFY]** `internal/cli/mcp_todo.go:57` — `todo_claim` MCP 도구, CLI와 동일한 결정 경로 + 거부 텍스트 동일성.
- Reference: research.md §6 (one-implementation-per-verb 패턴)

### M5 (Medium) — 테스트 집합

- **[MODIFY]** `internal/kanban/backlog_concurrency_test.go` — TestConcurrencyStress(기존, :135)에 클레임 경합 케이스 확장: 카드당 N≥2 액터(N = 2×M), exactly-one-wins, pre/post 전체 테이블 튜플 비교로 부분 쓰기 0 검출, raced 경로 byte-identity.
- **[MODIFY]** `internal/kanban/backlog_hold_migration_test.go` — 마이그레이션 순서/수렴 케이스 추가 (t1310 guard; 기존 `TestBacklogV1ToV2MigrationRoundTrip` 스위트 확장, 이름 불변).
- **[MODIFY]** `internal/cli/todo_json_golden_test.go` — golden gate 유지 확인 (엔진 우회 fixture; 기존 `TestTodoListJSON_GoldenByteIdentity`, 이름 불변).
- **[MODIFY]** 기존 pin 테스트 재확인(이름 불변, [NEW] 아님): `TestTodoVerbGuardRefusesWidenedMistypes`(internal/cli/todo_verb_leak_test.go), `TestTodoVerbsUnaffectedByFlag`(internal/cli/todo_flag_independence_test.go).
- **[NEW]** claim 테스트 집합 — internal/kanban: `TestBacklogClaim_ReclaimsExpired`, `TestBacklogClaim_LiveLeaseGuard`, `TestBacklogClaim_UnparseableExpiryExpired`, `TestBacklogClaim_RenewAfterExpiry` · internal/cli: `TestTodoClaim_Success`, `TestTodoClaim_RefusesNonQueued`, `TestTodoClaim_Renew`, `TestTodoClaim_LaneGovernance`, `TestTodoClaim_NoCardExit`, `TestTodoClaim_ListHistoryExposesLeaseColumns`.

## §C Risks & Mitigations

| 위험 | 완화 |
|------|------|
| 경합 클레임의 부분 쓰기(partial write) | Mutate 콜백 에러는 byte-identical abort가 보장됨(backlog_store.go:759-819) — 모든 클레임 경로를 Mutate 내부로 강제(REQ-TCL-004), AC-TCL-002 stress로 검증 |
| Mutate 밖 엔진 접근 (silent lost update 양방향) | REQ-TCL-004로 금지 + `@MX:WARN [TID:TX]`로 `e.db` 재진입 차단 경고 |
| 구버전 바이너리 쓰기가 임대 컬럼 데이터 유실 | 문서화된 downgrade 한계 — 구 바이너리의 `writeRecordArchive`는 자기 컬럼만 재 INSERT(research.md §4 파생, 미실행). docs에 명기, 런타임 방어는 하지 않음 |
| 라이브 플리트 v1-without-stamps (연구 발견 C3 — research.md §9, DP1 결정 아님) | 소급은 버전 정합 이후 순서 보장(REQ-TCL-002) + 읽기 NULL-tolerant(REQ-TCL-003) — 버전 정합 후 소급 순서로 retrofit |
| freeze 테스트 드리프트 | 재기록을 게이트로 강제 — M1에서 `wantItemsColumns`/`wantArchivedItemsColumns` 재기록 + freeze-header 항목이 AC-TCL-007 통과 조건 |

## §D mx_plan

- **@MX:ANCHOR 후보** — store `Claim`, `RenewLease`, `reclaimExpired` (fan_in ≥ 3 예상: CLI + MCP + todo auto + reclamation 경로)
- **@MX:WARN [TID:TX]** — claim 계열의 엔진 접근 경로에 트랜잭션 재진입 금지 표시
- **@MX:NOTE** — `todoRefuseLaneMutation` flag-form guard extension(C6 경계 — REQ-SD-015 bare-refusal semantics unchanged)에 거버넌스 소유자(t1338) 명기

## §E Verification Commands

```bash
# 스키마 freeze + 마이그레이션 순서/수렴 (AC-TCL-007) — 기존 테스트, 이름 불변
go test ./internal/kanban -run '^(TestSchemaFreezeRecordsTransitionStamps|TestBacklogV1ToV2MigrationRoundTrip|TestTodoHistoryAddsNoSchemaChange)$' -count=1

# 클레임 경합 — exactly-one-wins, 부분 쓰기 0, raced byte-identity (AC-TCL-002) — 기존 TestConcurrencyStress 확장
go test ./internal/kanban -run '^TestConcurrencyStress$' -count=1

# claim 신규 테스트 (AC-TCL-001/003~006, 009~012) — [NEW], run phase가 생성
go test ./internal/kanban -run '^(TestBacklogClaim_ReclaimsExpired|TestBacklogClaim_LiveLeaseGuard|TestBacklogClaim_UnparseableExpiryExpired|TestBacklogClaim_RenewAfterExpiry)$' -count=1
go test ./internal/cli -run '^(TestTodoClaim_Success|TestTodoClaim_RefusesNonQueued|TestTodoClaim_Renew|TestTodoClaim_LaneGovernance|TestTodoClaim_NoCardExit|TestTodoClaim_ListHistoryExposesLeaseColumns)$' -count=1

# golden + 기존 verb 표면 pin (AC-TCL-008) — 기존 테스트, 이름 불변
go test ./internal/cli -run '^(TestTodoListJSON_GoldenByteIdentity|TestTodoVerbGuardRefusesWidenedMistypes|TestTodoVerbsUnaffectedByFlag)$' -count=1

# 패키지 전체 회귀
go test ./internal/kanban/... ./internal/cli/... -count=1

# 빌드/정적
go build ./... && go vet ./internal/kanban/... ./internal/cli/...
```

> 스윕 가드: 각 `-run` 선택자는 매칭되는 테스트 함수 수 > 0을 선행 단언한다 — 0-매치(`no tests to run`, exit 0)는 실패로 간주한다(verification-completeness §1.1). [NEW] 테스트는 run phase가 위 정확한 함수명으로 생성해야 한다.
