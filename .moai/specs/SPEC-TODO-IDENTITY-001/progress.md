# Progress — SPEC-TODO-IDENTITY-001

## §E.1 Plan-phase Audit-Ready Signal

Status: implementation approved / run complete. 사용자 승인으로 gate를 우회했으며 감사 PASS가 아니다. manager-develop가 run-phase의 draft → in-progress 전이를 수행했다.

- audit_verdict: BYPASSED
- audit_report: .moai/reports/SPEC-TODO-IDENTITY-001/plan-audit-iter2.md
- audit_score: 0.75
- audit_ceiling: Tier M / 2 exhausted
- bypass_reason: 'Tier M audit ceiling 2 exhausted; D10-D13 post-audit RED contract repaired; user explicitly approved override'
- user: 'GOOS 오라버니~'
- kickoff: approved
- session: 01a09337-4567-7690-b6a5-dd0c421f3179
- approval_source: root가 전달한 사용자 명시적 informed override 및 Implementation Kickoff Approval
- approval_boundary: 독립 감사 iteration-2 FAIL 보존; 모든 7 AC·85% 품질·umbrella23AC 및 운영 권한 경계 면제 없음

## § Phase 4 Mode Selection

첫 implementation spawn 전 선택 기록이다. Kickoff passed / preferences collected. 사용자 선호는 승인된 UUIDv7·명시 null·기존 SQLite·Tier M 범위이며 추가 제품 선택은 하지 않는다.

| 판단 항목 | 값 |
|---|---|
| Tier | M |
| Estimated production/test files | 6–8 |
| Domain count | 2 — backend/database, internal/kanban 내부 |
| Language | Go + Markdown |
| Concurrency benefit | LOW |
| 작업 특성 | coding-heavy, shared transaction boundaries |
| Decision | serial |
| Scale | Standard |

| 모드 | 평가 | 결정 |
|---|---|---|
| direct | 작은 단일 수정에는 적합하나 이 범위는 schema/read/write/runtime 의존 검증을 순서대로 분리할 필요가 있다 | 미선택 |
| serial | 같은 transaction 경계와 공유 파일을 한 구현자가 순차 변경하며 RED→GREEN 귀속을 보존한다 | 선택 |
| fanout | 파일·transaction 경계 중첩으로 독립 동시 수정 이점이 낮다 | 미선택 |
| sweep | 균일한 기계적 수정이 아니라 상태·실패 처리 변경이다 | 미선택 |

Agent-team은 사용자의 명시적 요청이 없어 선택하지 않았다. tasks.md의 의존 순서를 따른다. 실제 구현·검증 결과는 담당자가 §E.2/§E.3에 기록한다.

## §E.2 Run-phase Evidence

| AC / invariant | Actual Output | Status |
|---|---|---|
| AC-TID-001 | Git/비Git fixture 2개, project UUID 2, card UUID 4, 합집합 6, t1/t2와 last_seq=2 | PASS |
| AC-TID-002 | legacy JSON·SQLite·active-WAL live/archive가 key-present null이고 pure read 변경 0 | PASS |
| AC-TID-003 | entropy/identity insert/stamp fault 각 reached=1, rollback 변경 0, fault 제거 재시도 성공 | PASS |
| AC-TID-004 | Add 반환=LoadPure, mutate→archive→restore→reopen project/card UUID 유지 | PASS |
| AC-TID-005 | live/archive runtime link 일치, missing/후보2 오류와 mutation 0, ReportedState만 completed | PASS |
| AC-TID-006 | distinct handles successes=2/cards=2/UUIDs=2; future·partial UNIQUE·malformed·retired reader/writer fail-closed와 mutation 0 | PASS |
| AC-TID-007 | 역순 UUIDv7에서도 AddedAt/State/ReportedState/OwnerLabel/ProvenanceJSON 비권위 보존 | PASS |
| core/runtime version | core schema_version=1 유지; runtime_schema_version은 생성된 runtime fixture에서 1, identity-only DB에서는 absent; identity_schema_version=1 | PASS |
| pure-reader no-write | LoadPure와 `todo list`가 legacy identity를 발급·migration하지 않고 literal null 반환 | PASS |
| transaction rollback | entropy/insert/stamp/preferred mismatch/UUID UNIQUE 충돌에서 record/schema/identity 변경 0 | PASS |
| retired/future conservation | 기존 future core/runtime 및 retired writer 거절 회귀 PASS | PASS |
| concurrency/race | 두 DB handle 동시 Add PASS; race selector exit 0, DATA RACE 미관측 | PASS |
| coverage | `todo_identity.go` 180/211 statements = 85.3%; package overall 30.8% | PASS |
| static/build | go vet exit 0; host build exit 0; CGO_ENABLED=0 Windows build exit 0; golangci-lint `0 issues.` | PASS |
| MX | TID marker exact count 5 | PASS |

상세 명령·RED 원문·GREEN 출력·DB readback은 `.moai/reports/SPEC-TODO-IDENTITY-001/red-baseline.md`와 `implementation.md`에 보존한다. 저장소 전체 CI, push, PR, merge, 운영 DB migration은 이 run-phase 근거에 포함하지 않는다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-12T10:14:19Z
run_commit_sha: PENDING
run_status: PASS
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 7
l44_pre_commit_fetch: NOT_APPLICABLE_NO_COMMIT_AUTHORITY
l44_post_push_fetch: PENDING
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_host: PASS
  windows_amd64_cgo_disabled: PASS
total_run_phase_files: 15
m1_to_mN_commit_strategy: manual_no_commit_no_push
repository_wide_ci_verdict: PENDING_INTEGRATION_BRANCH_CI
plan_audit_verdict: BYPASSED_FAIL_RETAINED
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-12T11:14:40Z
sync_commit_sha: pending-backfill-sync
sync_status: PASS_LOCAL_PENDING_INTEGRATION
sync_audit:
  report: .moai/reports/SPEC-TODO-IDENTITY-001/sync-audit-iter2.md
  verdict: PASS
  score: 98/100
acceptance:
  live_count: 7
  pass_count: 7
  fail_count: 0
documentation:
  changelog: CHANGELOG.md#unreleased-fixed
  locale_pages: 4
  readme: NOT_CHANGED_NO_STALE_CLAIM_FOUND
  project_docs: NOT_CHANGED_NO_STALE_CLAIM_FOUND
frontmatter_status_transitions:
  spec_md: in-progress_to_completed
  plan_md: NOT_APPLICABLE_NO_FRONTMATTER
  acceptance_md: NOT_APPLICABLE_NO_FRONTMATTER
  progress_md: NOT_APPLICABLE_NO_FRONTMATTER
operational_migration: NOT_RUN
remote_ci: PENDING_INTEGRATION_BRANCH_CI
git_commit: PENDING_BACKFILL
plan_audit_verdict: BYPASSED_FAIL_RETAINED
umbrella:
  spec: SPEC-TODO-UNIFIED-001
  card: t648
  state: picked_in_progress
  completed_claim: false
b12_self_test_a: PASS_CHANGELOG_REFERENCE_BASELINE_0
b12_self_test_b: PASS_LIVE_AC_COUNT_7
b12_self_test_c: PASS_ALL_CHANGELOG_PATHS_EXIST
canary_compliance_check: NOT_APPLICABLE
```

Plan audit iteration 2의 FAIL 0.75와 사용자 승인 BYPASSED 이력은 변경하지 않는다. 이 child의 완료는 UUIDv7 identity 기반에만 해당하며 자동 완료, receipt/recovery, hooks, Graph/UI와 운영 DB 이전을 포함하지 않는다. t648은 계속 `picked`·진행 중이다.
