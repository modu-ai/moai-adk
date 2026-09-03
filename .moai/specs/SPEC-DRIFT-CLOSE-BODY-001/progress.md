# SPEC-DRIFT-CLOSE-BODY-001 — 진행 기록

plan_status: draft

## §E.1 Plan-phase Audit-Ready Signal

- 카드: t410 (Class C) · 브랜치 `WT-drift-false-positive` · 트리 `.claude/worktrees/t410`
- Tier: S (spec.md + plan.md; AC는 spec.md §3 인라인). 근거는 `plan.md` §A Tier S 판정 근거
- REQ 7 / AC 7 — Tier S 상한(각 8) 안
- SPEC-ID 정규식 자체 점검: `[[ "SPEC-DRIFT-CLOSE-BODY-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`. `.moai/specs/` 충돌 없음
- 조사 원장: `.moai/reports/t410/discovery.md` + `r1-walker-trace.log` · `r2-blast-radius.log` · `r3-drift-row.log`
- 수리 후보 판정: **B 채택**(close 선언 커밋에 한해 본문 조회), A 기각(측정된 위험), C 기각·D 범위 밖 — `spec.md` §5
- 미확정으로 남긴 것: 파급 건수. 조사의 LOOSE 33 / TIGHT 15는 경계값이며 TIGHT에 알려진 오탐 1건 포함. 실제 수는 run-phase M3 전수 대조의 **결과**로 정해진다(AC-DCB-005)
- plan-audit: _\<pending\>_

## §E.2 Run-phase Evidence

_\<pending run-phase\>_

## §E.3 Run-phase Audit-Ready Signal

_\<pending run-phase\>_

## §E.4 Sync-phase Audit-Ready Signal

_\<pending sync-phase\>_
