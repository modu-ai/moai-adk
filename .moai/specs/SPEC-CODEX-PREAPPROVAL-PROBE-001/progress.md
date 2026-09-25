---
id: SPEC-CODEX-PREAPPROVAL-PROBE-001
document: progress
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
card: t1172
---

# Progress — SPEC-CODEX-PREAPPROVAL-PROBE-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier M, 산출물 3개(spec.md, plan.md, acceptance.md) + progress.md. status `draft`.
- SPEC ID 정규식 검사 `PASS`, `.moai/specs/`에 같은 ID 없음.
- 0.4.0(plan-audit iter 3 FAIL 0.81 → 리드 조건부 PASS-WITH-DEBT, 2026-09-25; 추가 plan-audit 없이 기계 게이트로 확인, sync-audit가 판정식을 실제 증거에 다시 대입한다): D-N11 이월 AC 인자·프롬프트의 호출 라벨별 반출과 라벨별 대조, D-N12 AC-CPP-010 행 전체 고정, D-N13 `temp_root` 아래 픽스처 루트, D-N14 한 토큰 모델 지정 거부, D-N15 `OPERATOR-CONFIRM`은 모든 갈래에서 Kickoff·첫 LIVE 전, D-N16 호출당 사용자 턴 1개. 기계 게이트 (a) `true`, (b)(c)(d) 거짓 — `.moai/reports/t1172/verdict.md` §Plan iter-4 gate. AC 수 변화 없음(16).
- 0.3.0(plan-audit iter 2 FAIL 0.79 반영, D-N1–D-N10): 시도 번호·리드 무효 기록 기반 재실행, 픽스처별 마지막 반출·시작 검사 대조, AC-CPP-010 고정 문장, 초기화 대상 보호. AC 수 변화 없음.
- 0.2.0(plan-audit iter 1 FAIL 0.70 반영): REQ 11개(REQ-CPP-001–011), AC 14개(AC-CPP-001–014) + 이월 LIVE AC 2개(AC-CAR-010, AC-CAR-011) = 16.
- 운영자 결정: D1 채택 형태(A1 대 A2)는 판별 결과 뒤. D1-sub(opt-in = 설정 섹션 불리언 키)와 D2(doctor WARNING + 고침 안내)는 잠정 기본값이며 Kickoff에서, 첫 LIVE 전에 확정한다(`OPERATOR-CONFIRM` 줄). 바뀌면 SPEC 개정 + plan-audit 재실행(`plan.md` §B).
- plan 단계 모델 호출 0회. 증거: `.moai/reports/t1172/verdict.md` §Plan, `.moai/reports/t1172/plan-checks/`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
