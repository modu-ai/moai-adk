# SPEC-CODEX-BODY-NEUTRALITY-001 — 진행 기록

card: t497 · tree: `.claude/worktrees/t497` · branch `WT-codex-neutrality` · base `ace1c5440`

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M)
- 기준선: `.moai/reports/t497/measurement.md`(`c9b226b22`) + 이 트리 재측정. 정정 3건은 `spec.md` §A.2.
- SPEC ID 정규식 실행 결과: `PASS`
- 미해결 마커 **0건** (`plan.md` §E): ① 코덱스 능력 부재 측정 가능 여부 → **철회**(트리 안 증거로 해소, `spec.md` §A.2 정정 2) · ② M5 착수 여부 → **운영자 결정 기록으로 전환**(`spec.md` §D)
- status: `draft`

### 수리 라운드 (iteration 2/2, 트리 `845dd65af`)

`.moai/reports/t497/plan-audit.md`(FAIL 0.63, MP-7 FAIL)에 대한 수리. D1-D7 전부 + 선택 D8-D10 반영. 좌표·기대값·경로는 이 트리에서 재측정했다.

- D1 — M1 을 프로브에서 문서 정정으로 재범위화, 마커 ① 철회 (`spec.md` §A.2 정정 2 · §B.1, `plan.md` §B-2 · §F M1)
- D2 — M1 에 착수 전 통과 불가능한 검증 4건 추가 (AC-CBN-006 (a)(b)(d))
- D3 — `AskUserQuestion` 좌표 3줄/발생 4로 정정, `manager-develop` 은 `Agent(` 축으로 이동 (`spec.md` §B.4, AC-CBN-002)
- D4 — 결속표 행 셀렉터를 구역 한정 `^| [a-z]` 로 단일화, 기대값 **3** 하나 (`plan.md` §F 머리말, `acceptance.md` §A · §D)
- D5 — 변경 반경을 문자 그대로의 11파일 집합으로 (`spec.md` §C.5, AC-CBN-011, `plan.md` §F M4)
- D6 — 지시 부류 셋 전부에 AC (AC-CBN-005 / 013 / 014), §D 의 건너뛰기 허용 제거
- D7 — REQ-CBN-008 경로를 `internal/template/templates/.claude/agents/moai/*.md` 로 명시, 루트 사본을 반경에서 제외
- D8 — 미러 스킬 표본 미확인 전제를 `spec.md` §B.6 · §D 로 반입
- D9 — AC-CBN-001 에 좌표 집합 대조(81줄) 추가
- D10 — 덮개 문장 검사를 실행 가능한 형태로 고쳐 AC-CBN-003 인접 경계 사례 + `plan.md` §F M3 ⑥⑦ 로 승격
- AC 수 12 → 14, REQ 수 15 → 16 (Tier M 상한 16 이내)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
