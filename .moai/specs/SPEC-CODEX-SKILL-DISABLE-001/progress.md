# SPEC-CODEX-SKILL-DISABLE-001 — 진행 기록

카드: t502 · 브랜치: `WT-codex-skillconfig`

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID 정규식 검사: 실행 `[[ "SPEC-CODEX-SKILL-DISABLE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`
- ID 유일성: `.moai/specs/` 에 동명 디렉터리 없음(생성 전 `ls -d .moai/specs/SPEC-CODEX-*` 로 확인)
- 산출 파일: spec.md · plan.md · acceptance.md · progress.md (Tier M)
- 미해결 `[NEEDS CLARIFICATION]`: 3건 — Q-a(게이트가 문자 그대로의 경로에 묶이는가 해석된 경로에 묶이는가), Q-b(미러 모드 교차에서도 같은 엔트리가 묶이는가), 이름 해석 후보 루트 집합(앞 둘에 종속). AC-CSD-050이 차단 게이트로 셋을 운반한다.
- 예산: REQ 16 / AC 16 — Tier M 상한(16/16)에 정확히 맞춤(`grep -c` 로 실측).
- 상태: `draft` — Implementation Kickoff Approval 대기

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
