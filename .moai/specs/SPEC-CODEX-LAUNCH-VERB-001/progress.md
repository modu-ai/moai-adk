# SPEC-CODEX-LAUNCH-VERB-001 — 진행 기록

카드: t391 · 워크트리 `.claude/worktrees/t391` · 브랜치 `WT-codex-launch-verb`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (3-artifact set: spec.md + plan.md + acceptance.md). 파일·LOC 축은 S 로 읽히나 REQ 12건이 Tier S 상한 8을 넘고, 완결 SPEC 의 요구 하나를 대체하는 승계 검증 부담이 별도 검증 층을 요구한다 — 높은 쪽을 택했다.
- REQ: REQ-CLV-001..012 (12건 / 상한 16)
- AC: AC-CLV-001..015 (15건 / 상한 16), REQ 전수 커버
- 승계: SPEC-CODEX-LAUNCHER-001 REQ-CL-002 를 대체. `depends_on: [SPEC-CODEX-LAUNCHER-001]`
- 게이트: 미해소 명확화 마커 **0건 — 2026-08-31 운영자 판정으로 2건 모두 해소**
  - §B.3 `-w` → (가) strip-and-set-Dir. resolve 이며 create 아님. 비대칭과 그 이유를 REQ-CLV-007/008 에 [HARD] 로 고정
  - §B.4 argv → 합성 동사는 어느 경로로도 자식에 닿지 않음. 라우팅 하류에 별도 번역 표(REQ-CLV-004 정규 5행)
- 이 트리에서 빌드·테스트 0건 (`plan.md` P2)
- 기계 검증 범위: `spec.md` 만 `moai spec lint --strict` 통과. `plan.md`·`acceptance.md` 는 린터 판정 없음 (`plan.md` P6)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
