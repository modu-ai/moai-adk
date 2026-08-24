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
- plan-audit iter2: **PASS-WITH-DEBT 0.925** (임계 0.80) — run-phase 진입 승인. blocking 6건 중 5건 완전 마감, AC 약화 0건, REQ 15 / AC 14 유지
- iter2 후속 (v0.4.0, 재감사 없음): 부채 #1 마감 — 빈 clause 금지를 **유일 적중**(정확히 1회) 요구로 강화(공백 한 칸·짧은 토큰 우회 차단; RED 1회 적중 24 / 2회 이상 1건 `CONST-V3R2-002`). N-1 §D8 인용 정정, N-2 REQ 전역 오름차순 재배치
- **잔여 부채 2건 — 수정 대상 아님, 판정자 지정됨** (`plan.md` §H): ① `CONST-V3R2-004` 근접 오답(`NOTICE.md` 로 이동 시 기계 통과 — sync 리뷰어가 이름으로 거부) ② 평가 엔트리 수 이중 카운트(clause/anchor 각각 101 — sync-auditor 가 §E.2 인용에서 수가 둘인지 확인)
- status: draft — run-phase 판단 대기 (kanban lead)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
