# 진행 기록 — SPEC-REVIEW-GATE-REPAIR-001

상태: draft (plan 단계). 이 파일의 §E.2 ~ §E.4 는 자리표시다. 해당 단계의 담당자가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft (재감사 대기). plan-audit 1회차 FAIL(score 0.625, MP-7). 수리 1회차(spec v0.1.1)를 적용했다. plan-audit 2회차는 FAIL(codex, 영수증 rcpt-0efbcef22b780f388f658933, 발견 F-1·F-2)이었고, 수리 2회차(spec v0.1.2)를 적용했다. 다음 plan-audit 이 통과하기 전까지 plan 은 audit-ready 가 아니다. 재감사 통과 전에는 audit-ready 로 표시하지 않는다.
- v0.2.0 적용 (분할): plan-audit 3회차 FAIL(codex, 영수증 rcpt-3e341c03b359a26ff7d01963, 발견 F-3: 활성 모드를 무시하고 manual.workflow 를 읽음, P2)에 대해 리더 판정 d-20261010T070728Z-6cdf(영실이 판단, 운영자 위임)는 F-3 을 결함으로 보아 부채로 받지 않고 REQ-ACE-005 를 분할하라는 것이었다. 기준 브랜치 선택(REQ-RG-005–008 과 그 인수 기준, AC-RG-006 포함)은 후속 카드 t1662 로 이관했고, 범위는 카드 항목 (1)·(3)·(5) 로 좁혔고, (6) 중 REQ-RG-016 은 이 SPEC 에 남는다 (121e, 6d67) (spec.md v0.2.0). 계획은 다음 계획 감사가 통과하기 전까지 audit-ready 가 아니다. [ID mapping v0.2.2: REQ-RG-016→REQ-RG-012. 이 항목의 REQ-RG-005–008 은 v0.2.0 에서 폐기된 식별자다.]
- plan_complete_at: 재감사 통과 시 기록한다. 현재 없음.
- tier: L (spec.md · plan.md · acceptance.md · design.md · research.md, 그리고 progress.md 와 decision-index.md)
- 운영자 판정: Q1–Q6 과 보드 판정 D3·D4·D11 기록 완료 (decision-index.md, 보드 기록 d-20261010T053034Z-c50f, d-20261010T053834Z-422b, d-20261010T054047Z-3fc2). 2회차: Q1 개정 (가) 및 F-2 기록 (보드 기록 d-20261010T063222Z-43e0, 보류 d-20261010T063038Z-f1a1 해소)
- 인용 재검증: research.md §3 (연구 입력 3건의 인용 전수), §9 (수리 1회차 재검증)
- 기준선 셀: acceptance.md §C (B-1 회귀 가드, B-2 결함 고정)
- 릴리스 게이트: 0건 (acceptance.md §A)
- 감사 입력 기록: `.moai/reports/t1620/plan-audit.md` §1.4 의 D1–D12 와 O1–O11 을 수리 대상으로 삼았다. 감사 보고서 파일은 수정하지 않았다.

## §E.2 Run-phase Evidence

_run-phase 진입 전. 자리표시._

## §E.3 Run-phase Audit-Ready Signal

_run-phase 진입 전. 자리표시._

## §E.4 Sync-phase Audit-Ready Signal

_sync-phase 진입 전. 자리표시._
