# progress.md — SPEC-TODO-STALE-STORE-001 (카드 t1307)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-TODO-STALE-STORE-001
card: t1307
tier: M
status: draft
artifacts:
  - .moai/specs/SPEC-TODO-STALE-STORE-001/spec.md
  - .moai/specs/SPEC-TODO-STALE-STORE-001/plan.md
  - .moai/specs/SPEC-TODO-STALE-STORE-001/acceptance.md
  - .moai/specs/SPEC-TODO-STALE-STORE-001/progress.md
spec_id_check: "PASS (Bash regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ — 충돌 없음 확인)"
code_changes_this_phase: none
open_clarifications:
  - "M3 삭제 승인 주체(운영자 직답 vs 리드 대행) — 실행 시점 확인 게이트에서 확정"
next_phase: run
```

plan-phase 조사 결과 요약:

- 큐 저장소 계층: `internal/kanban/state_dir.go`(홈 DB·레거시 경로 해석),
  `internal/homestate/paths.go`(project-key), `internal/kanban/backlog_sqlite.go`
  (`meta.last_seq` 키).
- 고지 표면: `internal/cli/todo_disclosure.go`(State D 전용 — 스테일 SQLite 미범위).
- doctor 등록: `internal/cli/doctor.go:187` `runGroupedChecksObserved`; binary_lag 쌍:
  `internal/cli/binary_lag_test.go` `namesAddedAfterBaseline` +
  `TestBinaryLag_AllowlistKeysAreLiveNames`; 골든: `doctor_golden_test.go`
  (`UPDATE_GOLDEN=1`).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
