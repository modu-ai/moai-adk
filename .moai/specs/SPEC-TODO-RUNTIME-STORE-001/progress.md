# Progress — SPEC-TODO-RUNTIME-STORE-001

## §E.1 Plan-phase Audit-Ready Signal

2026-09-12 운영자 승인: 직전 root의 “내부 저장 기반 먼저 → 해당 baseline의 실패 검증 → 통과 후 외부 연결” 검증 순서 예외 질문에 사용자가 **“승인!!!”**이라고 명시 응답했다. t648 내부 저장 기반에 한해 verification-order gate를 narrowly BYPASSED 처리한다. 기존 감사 FAIL은 그대로이며 PASS로 변경하지 않는다. 상세 범위·금지·만료는 ../../reports/t648/staged-verification-approval.md를 따른다.

초안. AC001 일부 기존 RED만 인용, 전체4 AC RED ledger 미채택. 독립 감사 전.

## §E.2 Run-phase Evidence

ModeSelection: 단일 도메인 SQLite 저장 기반으로 단일 manager-develop writer가 serial TDD를 수행했다. 읽기 전용 조사와 독립 감사만 병렬이다. 기존 Tier 유지. 생산 5개 파일(todo_runtime.go 신규 + factory_runtime.go/backlog_store.go/backlog_sqlite.go/backlog_migrate.go), 테스트 3개 파일. 신규 helper는 카드 Mutate와 전용 runtime 쓰기를 분리하되 같은 잠금과 transaction을 재사용한다.

| AC | 실행 근거 | Actual Output | Status |
|---|---|---|---|
| TRS-001 | 실제 기존 API→Todo 공개 JSON, Git/SPEC·비Git/무SPEC·archive·seed·upsert·sort | 기능 범위 20개 최상위 PASS; 독립 감사 package 전체 380개 PASS | PASS |
| TRS-002 | LoadPure 빈 배열·현재 값·활성 WAL bytes 불변·동시 30회 snapshot pair | TestTodoRuntimeSafetyPureSnapshotReadsActiveWAL PASS; 독립 감사 AC-TRS-002 PASS | PASS |
| TRS-003 | 실제 assignment abort trigger·seed rollback·카드 변경 경쟁·stale DTO 비갱신 | abort rollback·경쟁·stale DTO 보존 및 독립 감사 AC-TRS-003 PASS | PASS |
| TRS-004 | 미래 core/extension·퇴역 writer·install abort·legacy 내용·실제 relocation+parity | 미래/부분 schema·퇴역 writer 거절, relocation·JSON migration 보존 및 독립 감사 AC-TRS-004 PASS | PASS |
| 보존 | 기존 Factory guard 및 다른 세션 변경 비수정, live/archive 완료권위 유지 | git diff --check 출력 없음 | SELF-CHECK PASS |
| 품질 | 변경 패키지 기준 85% | 독립 감사 B1 전체 패키지 380개 PASS, coverage 86.2%; B2 관련 회귀 119개·race PASS, golangci-lint `0 issues.` | PASS |

원문 RED/GREEN·정확한 명령·baseline·coverage 분리는 ../../reports/t648/runtime-store-implementation.md 및 runtime-store-red.md를 따른다. 검증 순서 예외는 전체 품질 기준을 면제하지 않는다.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-pass

독립 감사 최종 판정은 최초 내부 저장 기반 4 AC에 한해 PASS 100/100이다. B1에서 변경 패키지 전체 380개 PASS와 coverage 86.2%를 관측했고, 이후 storage 연관 errcheck 16건의 cleanup·테스트 Scan 오류 처리만 교정한 B2에서 관련 회귀 119개, race, 독립 SQLite fixture, vet, diffcheck, gofmt와 golangci-lint `0 issues.`를 다시 관측했다. B1의 86.2%를 B2의 재측정 수치로 소급하지 않으며, 선택 검사 47.0%·47.6%와 전체 패키지 수치를 구분한다. 원문 명령·출력·해시·범위는 ../../reports/t648/runtime-store-sync-audit.md 및 runtime-store-implementation.md를 따른다.

run_commit_sha: 미커밋; HEAD a315dad9af0d3a0e04862e6106b3993d9a3812f7 위 동결 작업

AC 결과: 4 PASS / 0 FAIL. 통합 브랜치 CI: PENDING. commit/push/PR/설치/운영 DB/카드 변경: NOT_RUN. CLI/hooks 자동 완료와 운영전환은 금지 유지. 전체23AC나 t648 완료는 주장하지 않는다. 구현 전 별도 LSP baseline은 미캡처, 구현 후 gopls clean만 관측했다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-12T11:14:40Z
sync_commit_sha: pending-backfill-sync
sync_status: PASS_LOCAL_PENDING_INTEGRATION
sync_audit:
  report: .moai/reports/t648/runtime-store-sync-audit.md
  verdict: PASS
  score: 100/100
acceptance:
  live_count: 4
  pass_count: 4
  fail_count: 0
documentation:
  changelog: CHANGELOG.md#unreleased-fixed
  locale_pages: 4
  readme: NOT_CHANGED_NO_STALE_CLAIM_FOUND
  project_docs: NOT_CHANGED_NO_STALE_CLAIM_FOUND
frontmatter_status_transitions:
  spec_md:
    run_phase_prerequisite_repair: draft_to_in-progress_by_manager-develop
    sync_phase_close: in-progress_to_completed_by_manager-docs
    phase12_backup_to_final_observed_delta: draft_to_completed
    direct_draft_to_completed_claim: false
  plan_md: NOT_APPLICABLE_NO_FRONTMATTER
  acceptance_md: NOT_APPLICABLE_NO_FRONTMATTER
  progress_md: NOT_APPLICABLE_NO_FRONTMATTER
operational_migration: NOT_RUN
remote_ci: PENDING_INTEGRATION_BRANCH_CI
git_commit: PENDING_BACKFILL
umbrella:
  spec: SPEC-TODO-UNIFIED-001
  card: t648
  state: picked_in_progress
  completed_claim: false
b12_self_test_a: PASS_CHANGELOG_REFERENCE_BASELINE_0
b12_self_test_b: PASS_LIVE_AC_COUNT_4
b12_self_test_c: PASS_ALL_CHANGELOG_PATHS_EXIST
canary_compliance_check: NOT_APPLICABLE
```

Phase 12 최초 백업은 manager-develop의 선행 lifecycle 보수보다 먼저 만들어져 백업의 `spec.md`는 `draft`, 최종 파일은 `completed`다. 실제 소유권 이력은 manager-develop이 run-phase 선행 보수로 `draft → in-progress`를 수행한 뒤 manager-docs가 sync-phase에서 `in-progress → completed`로 닫은 두 전이이며, `draft → completed` 직접 전이를 주장하지 않는다.

이 child의 완료는 runtime 저장 기반에만 해당한다. 자동 `picked → done`, completion receipt/recovery, hooks, Graph/UI와 운영 DB 이전은 포함하지 않으며 t648은 계속 `picked`·진행 중이다. Identity child와 함께 작성한 동기화 범위와 공백은 `.moai/reports/SPEC-TODO-IDENTITY-001/stacked-sync.md`에 기록한다.
