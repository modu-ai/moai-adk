# progress.md — SPEC-TODO-CLAIM-LEASE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_phase:
  card: t1342
  spec: SPEC-TODO-CLAIM-LEASE-001
  tier: M
  artifacts: [spec.md, plan.md, acceptance.md, spec-compact.md, progress.md]
  dp1_decisions: [C1, C2, C4, C5, C6]
  dp1_approved: 2026-09-29
  research_evidence:
    - .moai/state/plan-research-t1342/spec-proposal.md
    - .moai/state/plan-research-t1342/research.md
    - .moai/state/plan-research-t1342/per-lens-reports.md
  status: draft
  plan_complete_at: 2026-09-29T22:03:37+0900
  plan_status: audit-ready
  plan_audit:
    iteration1: "FAIL 0.60 — .moai/reports/plan-audit/SPEC-TODO-CLAIM-LEASE-001-review-1.md (D1-D10)"
    iteration2: "PASS 0.86 — .moai/reports/plan-audit/SPEC-TODO-CLAIM-LEASE-001-review-2.md (must-pass 6/6, D1-D10 해소 재검증)"
    threshold: "0.80 (Tier M standard)"
    review_lenses: "FO-PLAN-2 4렌스 — .moai/state/plan-research-t1342/lens-{1-frontmatter,2-testability,3-scope,4-commands}.md"
```

## §E.2 Run-phase Evidence

### AC matrix (manager-develop, run-phase; measured at branch WT-todo-claim-lease, HEAD c982f763d + M5 repairs, 2026-09-29/30)

| AC | Status | Command (env-scrubbed compound form) | Observed result |
|----|--------|--------------------------------------|-----------------|
| AC-TCL-001 | PASS | `go test ./internal/cli -run '^TestTodoClaim_Success$' -count=1` | ok — oldest card claimed, picked_by/lease_expires_at(≈15m)/picked_at stamped, id+text+expiry output |
| AC-TCL-002 | PASS | `go test ./internal/kanban -run '^TestConcurrencyStress$' -count=1` (also `-race`) | ok — 8 actors over 4 cards: 4 wins, raced losses ErrClaimRaced, partial writes 0, raced rows byte-identical; green under -race |
| AC-TCL-003 | PASS | `go test ./internal/cli -run '^TestTodoClaim_RefusesNonQueued$' -count=1` | ok — picked/hold/dropped/future-state arms refuse distinctly, engine artifact sha256 identical around every refusal |
| AC-TCL-004 | PASS | `go test ./internal/kanban -run '^TestBacklogClaim_ReclaimsExpired$' -count=1` | ok — lapsed lease returned to queued with the 4-field cleared set, id+prev-holder surfaced, reclaimed card re-claimed |
| AC-TCL-005 | PASS | `go test ./internal/kanban -run '^TestBacklogClaim_UnparseableExpiryExpired$' -count=1` | ok — unparseable expiry judged EXPIRED (C4), reclaimed by both ReclaimExpired and claim |
| AC-TCL-006 | PASS | `go test ./internal/cli -run '^TestTodoClaim_Renew$' -count=1` + `go test ./internal/kanban -run '^TestBacklogClaim_RenewAfterExpiry$' -count=1` | ok — renewal resets to now+15m (homestate absolute-reset shape), no other field moves, foreign holder refused, renew-after-expiry commits the return then refuses |
| AC-TCL-007 | PASS | `go test ./internal/kanban -run '^(TestSchemaFreezeRecordsTransitionStamps\|TestBacklogV1ToV2MigrationRoundTrip\|TestTodoHistoryAddsNoSchemaChange)$' -count=1` + new `TestBacklogLeaseRetrofitConvergesAfterRebuild` / `TestBacklogReadToleratesMissingLeaseColumns` | ok — 11-column items / 14-column archived_items tuples converge fresh↔upgraded, retrofit strictly after version reconciliation, pre-retrofit databases read with nil lease pointers |
| AC-TCL-008 | PASS | `go test ./internal/cli -run '^TestTodoListJSON_GoldenByteIdentity$' -count=1` | ok — golden byte-identity held with NO fixture regeneration; lease fields excluded from the JSON render via todoJSONProjection |
| AC-TCL-009 | PASS | `go test ./internal/cli -run '^TestTodoClaim_LaneGovernance$' -count=1` + `TestTodoClaimMCP_Mirror` | ok — arms 1/3 refuse with the identical todoLaneMutationRefusalText("claim") (CLI and MCP carry the same text), arm 2 attributes picked_by=lane-9, queue byte-identical on refusals |
| AC-TCL-010 | PASS | `go test ./internal/cli -run '^TestTodoClaim_NoCardExit$' -count=1` | ok — exitCodeError code 3 (factoryNextNoCardExit), non-error stdout message |
| AC-TCL-011 | PASS | `go test ./internal/kanban -run '^TestBacklogClaim_LiveLeaseGuard$' -count=1` + existing `TestTodoUnpick_RevertsPickedToQueued` / `TestTodoUnpick_RefusalsLeaveFileUntouched` | ok — raced refusal byte-identical, A's reclamation leaves live-leased B untouched, unpick guards stay green name-invariant |
| AC-TCL-012 | PASS | `go test ./internal/cli -run '^TestTodoClaim_ListHistoryExposesLeaseColumns$' -count=1` | ok — human list/history carry by=/lease= cells on lease-holding rows (text stays last, no-lease rows keep historical shape), JSON face exposes nothing |

### E8 RED evidence (verbatim pre-GREEN, captured per milestone)

- M1: freeze + convergence asserted 9/12-column tuples vs expected 11/14 (`--- FAIL: TestBacklogLeaseRetrofitConvergesAfterRebuild ... want them to end with "picked_at:TEXT:0:NULL dropped_at:TEXT:0:NULL picked_by:TEXT:0:NULL lease_expires_at:TEXT:0:NULL"`, plus both freeze failures); mutant probe on the NULL-fallback guard observed failing (`lease pointers = 0x…/<nil>, want nil`) before revert.
- M2: verbatim compile failure — `store.Claim undefined`, `undefined: ErrClaimRaced`, `store.RenewLease undefined`, `undefined: ErrLeaseExpired`, `store2.ReclaimExpired undefined`.
- M3: verbatim run failures — `unknown command "claim" for "todo"`, `unknown flag: --lane`, no-card arm not an exitCodeError.
- M4: verbatim compile failure — `undefined: handleTodoClaim`.
- M5: the full-suite cascade observed four pre-existing column/verb pins red before their additive re-records (`TestTransitionStampColumns_FreshUpgradedConverge`, `TestBacklogLanding_ItemsColumnShape`, `TestBacklogLanding_ArchivedItemsColumnShape`, `TestBacklogArchive_PerItemContractFrozen`) and two walk/parity pins (`TestSD_AC015_LaneQueueAllowlistWalk`, `TestGTDAllTodoVerbsParity`).

### Security-scan pre-disposition (run-phase, for the sync-phase Phase 8 scan)

The MoAI Security Guardian flagged "sql-injection (high): SQL built by string concatenation" on the M1 edits. Disposition: NOT a finding. The flagged interpolations are (a) `ensureColumn`'s `ALTER TABLE %s ADD COLUMN %s TEXT` — SQLite cannot parameterize DDL identifiers, and every table/column value is a compile-time constant from `backlogLeaseColumns`/`backlogTransitionStampColumns`; (b) the widened SELECT column lists in `readSnapshot`/`readArchive` — every fragment comes from `columnExpr`, which returns only a bare column name or the literal `NULL`; (c) the INSERT statements — fully static column lists, every VALUE through a `?` placeholder. No user or runtime input reaches any statement text; the pattern is the in-repo additive-DDL discipline every prior retrofit (landing, transition stamps) runs through. Disposition documented in code comments at the flagged sites (backlog_sqlite.go ensureLeaseColumns, backlog_migrate.go both SELECTs).

### Known limitations / post-implementation review

- 구버전 바이너리 downgrade: an old binary's whole-record rewrite drops lease column data (documented edge case; no runtime defense per plan §C).
- `todo unpick` keeps its exact historical semantics — it does not clear picked_by/lease_expires_at (PRESERVE constraint); a queued card can carry stale lease fields until its next claim overwrites them. Reclamation's cleared-field set covers the lease fields; noted as cosmetic residue, flagged to sync/docs.
- Full `./internal/cli` suite on this host runs past go's default 10m package timeout (~20m observed); load-induced flakes observed once (`TestGateCmd_SecondRunWaitsForFirst` — execution windows overlapping by 0ms under full-suite load; passes solo and touches none of this SPEC's code). CI is the verdict surface.
- BacklogItem's ANCHOR count in backlog_store.go is now 8 vs the mx.yaml advisory limit 3 — the file already carried 5 pre-SPEC; per scope discipline no pre-existing tags were demoted.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_phase:
  card: t1342
  spec: SPEC-TODO-CLAIM-LEASE-001
  run_complete_at: 2026-09-30T00:30:00+09:00
  run_commit_sha: "pending-backfill-run"
  run_status: audit-ready
  ac_pass_count: 12
  ac_fail_count: 0
  preserve_list_post_run_count: 5
  l44_pre_commit_fetch: n/a (worktree lane; no push per lane protocol)
  l44_post_push_fetch: n/a (lane does not push)
  new_warnings_or_lints_introduced: 0
  cross_platform_build:
    native: exit 0
    windows_amd64: exit 0
  total_run_phase_files: 12
  m1_to_mn_commit_strategy: "5 commits (M1 schema+freeze / M2 store ops / M3 CLI verb / M4 MCP mirror / M5 test cascade + repairs)"
  red_evidence: captured verbatim per milestone (see §E.2 E8)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_phase:
  card: t1342
  spec: SPEC-TODO-CLAIM-LEASE-001
  sync_complete_at: 2026-09-30T00:00:00+09:00
  sync_commit_sha: "pending-backfill-sync"
  sync_status: complete
  changelog_entry_position: "[Unreleased] › Added › first bullet (SPEC-TODO-CLAIM-LEASE-001)"
  mx_tag_validation: performed as sync sub-step — no new MX tags owed; BacklogItem ANCHOR advisory exceedance documented in §E.2 known limitations
  frontmatter_status_transitions:
    in_progress_to_implemented: sync commit
    implemented_to_completed: sync commit (merged close)
  docs_sync:
    changelog: "[Unreleased]/Added entry emitted (B12 pre-emission grep count 0 before write; AC live count 12 = acceptance.md)"
    doc_parity: "DocParity selector 0 tests; real pins TestGTDAllTodoVerbsParity + TestSD_AC015_LaneQueueAllowlistWalk PASS; claim row added to gtd.md live + template mirror in same commit"
```

## §F.1 Plan-phase 기록 (Plan-Phase Record)

### 리서치 팬아웃 출처 (Research Fan-out Provenance)

- **팬아웃**: FO-PLAN-1 — 4개 읽기 전용 렌즈 탐색 + 1개 신디사이저. 렌즈: `codebase-precedent`, `constraints-risks`, `prior-SPEC-memory`, `queue-consumer-inventory`. 4렌즈 + 신디사이저 전부 정상 반환.
- **run wf id**: 세션 상태 파일에서 회수 불가(기록 없음) — 증거는 디스크 아티팩트 3건(`.moai/state/plan-research-t1342/{research.md, per-lens-reports.md, spec-proposal.md}`, 2026-09-29 20:55–21:03 작성)으로 대신한다. 미회수 항목을 Gaps로 명시.
- **manager-spec 제안**: spec-proposal.md (Phase 8 산출, Phase 10 이전) — REQ-TCL-001..015, AC 스케치 AC-TCL-001..010, M1-M5 계획, MX 계획 입력 포함.

### DP1 운영자 결정 (타임스탬프 포함)

- **승인 일자**: 2026-09-29 (구속력 있는 운영자 결정)
- C1 — 마이그레이션: additive `ensureColumn`, schema-stamp bump 없음("2" 유지); freeze 테스트 재기록 필요
- C2 — 클레임 원자성: Mutate 콜백 내 freshly-loaded 레코드 상태 술어 (flock = CAS 원시체); version 컬럼·Mutate 밖 엔진 접근 없음
- C4 — 파싱 불가 만료: EXPIRED (factory 기본값)
- C5 — 반환 감사: 사람이 읽는 claim/reclaim 출력 라인(id + 이전 보유자); events 테이블 없음
- C6 — `--lane <label>`: 운영자/리드 공급 레인 레이블; REQ-SD-015 거부 의미론 불변 — 레인 셀프 클레임 승인(carve)은 t1338 소유. (iteration-2 정정: 본 SPEC의 가드 변경은 flag-form guard extension inside todoRefuseLaneMutation이며 `factoryLaneRefusal()` 참 보유 시 `--lane`형도 기존 거부 텍스트로 거부)

### Phase 4 모드

- **serial** — 팬아웃 리서치는 별도 수행, SPEC 저작은 단일 시리얼 경로.

### Clarity-skip 사유

- 디스패치 및 카드 텍스트에 ≥5개 기술 키워드(`BacklogStore.Mutate`, `ensureColumn`, CAS, flock, `lease_expires_at`, REQ-SD-015) — Context-First Discovery Socratic 인터뷰 예외(기술 지시 완비) 적용, 클래리피케이션 라운드 스킵.

### Tier 산출 근거

- **dispatch 타이틀 "Tier M"** 기준 — Tier M 아티팩트 세트(spec.md + plan.md + acceptance.md + progress.md)에 dispatch가 명시 지정한 `spec-compact.md`를 추가 산출.

### 미회수/미검증 항목 (Gaps)

- FO-PLAN-1 run wf id — 상태 파일에서 회수 불가 (위 기술).
- 구버전 바이너리 쓰기의 임대 데이터 유실 및 Mutate 밖 CAS clobber는 코드 파생 판단으로, 런타임 재현 미실행 (research.md §8 고지 분).

### Iteration-2 수정 기록 (plan-audit review-1 대응)

- plan-audit iteration 1: FAIL 0.60 (기준 0.80) — `.moai/reports/plan-audit/SPEC-TODO-CLAIM-LEASE-001-review-1.md`.
- D1-D10 전건 수정. 결정적 2건의 구속 방향 반영: D2+D3 — REQ-TCL-013에 `factoryLaneRefusal()` 참 보유 시 `--lane`형도 기존 거부 텍스트로 거부하는 절 추가 + "carve" 용어를 "flag-form guard extension inside todoRefuseLaneMutation; REQ-SD-015 bare-refusal semantics unchanged"로 전면 치환(C6·REQ-013·Out of Scope·spec §E·plan M3 상호 일치). D1 — 모든 AC observable을 실존 테스트(9개 이름 트리 검증 완료) 또는 [NEW]+정확한 함수명으로 재고정.
- 기타: D4(신규 AC-TCL-011 live-lease 가드 + unpick carve-out 기존 가드 인용), D5(신규 AC-TCL-012 human 표면 + REQ-010 human-vs-JSON 재조정), D6(경합 AC에 카드당 N≥2·pre/post 전체 튜플 비교 검출기·raced byte-identity 고정), D7(claim-family 열거·cleared-field 집합·seq 순서·renew-expired observable), D8(C3을 DP1 결정이 아닌 연구 발견으로 재표기), D9(spec §E delta 마커가 plan [MODIFY] 표면 전체 커버 — todo_disclosure.go·golden/concurrency/migration 테스트 포함), D10(backlog_migrate.go 인용을 writeArchive ~:337-401/:381과 writeRecordArchive :422-515/:482 이원 범위로 정정, REQ-006 Event-driven 재표기, todo.go :303 정정).

## §F Phase 4 Mode Selection (run-phase)

- **Input parameters**: tier M · scope ~12 파일 (kanban 스키마/스토어 + cli verb/MCP + 테스트 6파일) · 도메인 2 (internal/kanban, internal/cli) · 언어 혼합 Go+markdown 없음(순 Go) · 병렬 이익 LOW (coding-heavy) · Agent Teams 사전 요건 미요청.
- **Mode evaluation**: direct — 아님(다중 파일 의미 변경) / serial — **selected** / fanout — 아님(coding-heavy, Anthropic 병렬화 경고) / sweep — 아님(기계적 균일 변환 아님) / agent-team — 미요청.
- **Decision**: serial
- **Justification**: 단일 도메인 밀착 코딩 작업(스키마→스토어→CLI→MCP→테스트가 순차 의존)이라 단일 manager-develop 스폰이 마일스톤 M1-M5를 순서로 수행하는 것이 병렬 팬아웃보다 낫다(Anthropic coding-task caveat). 팬아웃 리서치는 plan 단계에서 이미 별도 수행됐다.
- **Run-gate skip 기록 (Phase 1)**: plan-audit PASS 0.86 ≥ Tier M 임계 0.80 + plan-artifact 해시 불변(review-2 이후 spec/plan/acceptance 무변경 — progress.md는 해시 대상 아님) + depends_on SPEC-TODO-RUNTIME-STORE-001 completed → 3조건 충족으로 Phase 1 재실행 스킵.
- **진행 축**: 운영자가 Kickoff에서 자율(골 무장) 선택 — 기계 종료 조건 무장됨(세션 ed6e9938…, max_turns 30).
