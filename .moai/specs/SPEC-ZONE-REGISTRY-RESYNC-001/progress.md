# Progress — SPEC-ZONE-REGISTRY-RESYNC-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (근거: `plan.md` §A Tier 판정 — REQ 15/16, AC 14/16, 압축 없음)
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- 범위: 3축 전부 (clause 재동기화 + anchor 탐지·수리 + 기계적 가드)
- REQ: 15건 (REQ-ZRR-001..015) · AC: 14건 (AC-ZRR-001..014, BLOCKING 9)
- 마일스톤 순서 구속: M1 수리 → M2 가드 → M3 검증 (의존성, 권고 아님 — `plan.md` §F)
- 근거 보고서: `.moai/reports/t232/findings.md` + `validate-repro.txt` + `analysis-repro.json`
- RED 기준값이 모든 AC에 명시됨 (`acceptance.md` §D 매트릭스), 측정 트리 `294b4b6ab`
- plan-audit iter1: **FAIL 0.75** (Tier M 임계 0.80; must-pass 7/7 PASS, 4개 차원 전부 0.75). 보고서: `.moai/reports/t232/plan-audit-iter1.md`
- iter1 반영 (v0.3.0): blocking 6건(D-1 자기참조 `file:` / D-2 빈 clause / D-3 `|| true` / D-4+D-8 평가 엔트리 수·변이 대상 고정 / D-9 paths-filter / D-5 파일 수 오기) + optional 4건(D-6 HISTORY / D-7 REQ-015 분리·재배치 / D-10 slug 열거 5개 / D-11 §7 갱신) 전부 적용. 새 AC·티어 변경 없음 — 기존 AC 의 Then 절 강화로 흡수
- status: draft — plan-audit iter2(델타 한정) 대기

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
