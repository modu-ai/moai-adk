# SPEC-TODO-ANALYSIS-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (spec.md + plan.md + acceptance.md). 근거는 `plan.md §B`.
- REQ 15 / AC 15 (Tier M 상한 각 16).
- SPEC ID regex check: 실행됨, `PASS` (`internal/spec/lint.go` `specIDPattern` 대응 패턴).
- ID 충돌 검사: `.moai/specs/SPEC-TODO-*` 0건.
- Out of Scope: 6개 `### Out of Scope — <topic>` 소제목, 각 `-` 불릿 보유.
- 독트린 충돌 입장: `spec.md §B` — 가역성 단독이 아니라 **가역성 + 현저성**을 자동 변형의 허가 기준으로 세우고, 순서 정리와 카드 흡수는 현저성 미달로 기각, 정확 중복 입력 거절과 소견 기록만 자동 허용.
- 상태: plan-audit 대기.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
