# SPEC-CODEX-GATE-SCOPING-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03
tier: M
cycle_type: tdd
artifacts: spec.md, plan.md, acceptance.md, decision-index.md (본 파일 포함 5본)
evidence_copy: .moai/reports/t1404/gate-block-evidence.md (primary 원본 t1395 처분 기록 9건의 본 카드 필요분 사본)
plan_audit: iteration-1 FAIL(2026-10-03, .moai/reports/t1404/plan-audit.md — MP-8 RED 미관측 + D2·D3·D4 blocking) 수리 완료(v0.1.1) — 재판정 대기(run 진입 전 plan-auditor 판정 필요). 수리 요지: R 5건(AC-001·005·007·008·010) RED를 plan 단계에서 실측(acceptance.md §D.0 장부 — 단일 호출 명령·원문 stdout·exit 코드·트리 SHA 2de0a2cb6), G 4건(AC-002·003·009·012) regression-guard 재분류(§D.1), primary_scope 판독 처분표(spec §F.2 — D2), M4 델타 삼팔 pathspec 확장(plan M4 — D3), AC-012 항목 멤버십 기준(§D.0/plan §C — D4), §A.4 좌표 :938/:954 정정(D5).
red_tests: internal/cli/codex_review_gate_primary_scope_red_test.go(AC-001·005·007·008) · internal/template/hook_gate_reports_exclude_test.go(AC-010) — plan 단계 저작, 현 트리 관측 적색(acceptance.md §D.0). manager-develop가 M2/M3/M4에서 GREEN 전환(새로 저작하지 않음).

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소관>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
