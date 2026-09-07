# SPEC-CODEX-SKILL-DISABLE-001 — 진행 기록

카드: t502 · 브랜치: `WT-codex-skillconfig`

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID 정규식 검사: 실행 `[[ "SPEC-CODEX-SKILL-DISABLE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`
- ID 유일성: `.moai/specs/` 에 동명 디렉터리 없음(생성 전 `ls -d .moai/specs/SPEC-CODEX-*` 로 확인)
- 산출 파일: spec.md · plan.md · acceptance.md · progress.md (Tier M)
- 미해결 `[NEEDS CLARIFICATION]`: **0건**. v0.1.0의 3건(Q-a 경로 표기, Q-b 미러 모드 교차, 후보 루트 집합)은 `.moai/reports/t502/gate-path-shape.md`(19셀, `codex-cli 0.153.4`)로 전부 닫혔다 — 발행 표기는 `<projectRoot>/.agents/skills/<skill>/SKILL.md` 로 확정.
- 선행 증거 2본: `.moai/reports/t504/skills-config-path-shape.md`(`enabled` 필수, 게이트 존재) + `.moai/reports/t502/gate-path-shape.md`(표기 확정, realpath 정규화, skipped 주장 반증).
- 철회 1건: v0.1.0의 「`MirrorModeSkipped` ⇒ 해소 실패」 주장은 측정이 반증해 철회됨(spec.md §A 각주).
- 예산: REQ 16 / AC 16 — Tier M 상한(16/16)에 정확히 맞춤(`grep -o … | sort -u | wc -l` 로 실측). 조항을 넓혀 수를 맞추지 않았다 — 두 성질을 지는 기준에는 판별 셀도 둘씩 붙였다.
- 남은 [HARD] 게이트: AC-CSD-050(발행 코드 착지 직전, 그 시점 codex 버전에서 2셀 재측정).
- 상태: `draft` — Implementation Kickoff Approval 대기

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
