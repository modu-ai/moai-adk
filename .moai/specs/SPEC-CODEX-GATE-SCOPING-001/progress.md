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

## §F Phase 4 Mode Selection

- 입력 파라미터: tier=M · scope=7 소스 파일 + 7 테스트 파일(예상) · 도메인 수=2(internal/cli·internal/config Go + 훅 스크립트/템플릿) · 파일 언어 혼합=Go+shell·markdown · 병렬 이득=LOW(코딩 중심 — Anthropic 코딩 과제 병렬화 주의) · Agent Teams 전제=미충족(명시 요청 없음)
- 모드 평가: direct=미선정(단일 자리수 자리표 이상, 의미 변경 있음) · fanout=미선정(코딩 중심, 도메인 2개로 3 미달) · sweep=미선정(30파일 미달·의미 변형 작업) · agent-team=미선정(명시 요청 없음)
- **Decision: serial** — 마일스톤당 1회 서브에이전트 순차 배차
- 근거: 코딩 중심 작업은 Anthropic 지침상 serial이 기본이고, 마일스톤 M1-M6이 config→handler→scope→템플릿 순으로 의존하는 단일 사슬이라 병렬 분해 이득이 없다. RED 테스트는 plan 단계에서 이미 저작돼 있어 manager-develop는 GREEN 전환 소관(새로 저작하지 않음).
- 경계 메모: M4 착수 전 리더의 C5(pathspec 범위) 한 줄 확정 대기 — C5 답이 오기 전 M4 진행 보류. C1-C7 부채 처분(강화 vs 후속 카드)도 리더 답과 함께 run 배차 범위에 반영.
- C5 처분(2026-10-03 리더): ①②팔 모두 reports 항목만 건다(WCI_EXCLUDES 전체 아님) — M4 보류 해제. C1-C7은 run 범위 밖(후속 카드). 상세: .moai/reports/t1404/kickoff-decision.md 덧붙임.
