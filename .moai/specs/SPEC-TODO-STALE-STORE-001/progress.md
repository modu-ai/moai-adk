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
open_clarifications: []
resolved_clarifications:
  - "M3 삭제 승인 주체 — 해결됨(2026-09-29, plan-audit iter1 D1): 운영자·리드 양쪽 확인으로 확정. 카드 t1307 본문 「파기 전 운영자·리드 확인」 근거. 확인 기록은 §E.2에 경로·sha256·확인 주체와 함께 기록."
next_phase: run
```

plan-audit iter1 수리(2026-09-29, manager-spec): D1(M3 게이트 확정 + 미해결 질의 마커
제거)·D2(존재하지 않는 AC 오인용 제거)·D3/D4(`-run` 패턴 비공허화 — anchored 패턴은
0개 적중)·D5(고지 진입점 2건 명시 — `todo_history.go:157` 직접 호출 경로 포함)·D6
(`state_dir.go` 인용 범위 :62-86 확대) 반영.

codex cross-audit 수리(2026-09-29, manager-spec, v0.1.2): REQ-TSS-001 5동사 고지의 AC
커버리지 결함 수리(GLM iter2 PASS 0.95를 codex OVERTURN 0.76으로 뒤집은 단일 차단
결함) — AC-TSS-001을 동사별 시나리오(AC-TSS-001a..e)로 분해, AC-TSS-002를 동사별
stdout 바이트 동일성·무고지로 확장, plan.md M1 판정 주체 표기 정합(AC-TSS-003 →
AC-TSS-001a..e).

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

## §G Gate Disposition Log

- 2026-09-29 — served_model_gate 해제(이 트리 한정): 운영자 승인(리드 전달 2026-09-29) — 게이트 해제 + GLM iter2 PASS 0.95 채택, 단 codex 재감사 OVERTURN 시 채택 제외. plan-auditor 정의 선언 모델(opus)과 실제 서빙 모델(glm-5.3-flash) 불일치로 served-kind 거절이 서서 페이즈 진입 스폰이 차단되던 것에 대한 처분.
- 2026-09-29 — 제외 조항 소진: codex(GPT-6) 재감사 OVERTURN 0.76(REQ-TSS-001 5동사 고지 vs AC 커버리지 결함)으로 GLM 채택은 제외됨. GLM plan-audit 반복 상한 2/2 소진 — 수리 후 codex 델타 재감사가 유일 재심 경로이며, RECONFIRM 시 run 진입.
- 거부 영수증 파일(`.moai/state/audit-receipts/rejections/plan-auditor--unknown-spec--served.json`)은 지시에 따라 보존(삭제하지 않음). opus 클리어 감시 루프는 경로 폐쇄 확정으로 종료.
