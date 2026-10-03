# SPEC-CODEX-GATE-SCOPING-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03
tier: M
cycle_type: tdd
artifacts: spec.md, plan.md, acceptance.md, decision-index.md (본 파일 포함 5본)
evidence_copy: .moai/reports/t1404/gate-block-evidence.md (primary 원본 t1395 처분 기록 9건의 본 카드 필요분 사본)
plan_audit: iteration-1 FAIL(D1-D5) 수리 v0.1.1 → iteration-2 FAIL 0.8625(D6-D10, `.moai/reports/t1404/plan-audit-iter2.md`) → 리더 판정 범위 축소 후 3회차 수리 완료(v0.1.2, `.moai/reports/t1404/plan-decision-iter2.md` 지시 범위 한정) — 재판정 대기. v0.1.2 수리 요지: AC-010 대조군 재정의(비-reports 파손 변경 게이트 — D6), AC-012 레이아웃 독립 분할+AC-011 이관(D7), AC-001 행 내용·AC-008 재분류 행 단언·②팔 실입력·AC-007 settings.local.json 삭제(D8), 삼클래스 문언 정렬(D9), §D.0 핀 물리 상태 보충(D10). R 5건 전부 재관측(HEAD 9ef1cbedc + 개정 테스트 파일).
red_tests: internal/cli/codex_review_gate_primary_scope_red_test.go(AC-001·005·007·008) · internal/template/hook_gate_reports_exclude_test.go(AC-010) — plan 단계 저작, iteration-2 D6·D8 개정 후 전부 재관측 적색(acceptance.md §D.0 — 개정 blob 5779dd07·1aa630c3). manager-develop가 M2/M3/M4에서 GREEN 전환(새로 저작하지 않음).

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소관>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
