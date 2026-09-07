# SPEC-CODEX-BODY-NEUTRALITY-001 — 진행 기록

card: t497 · tree: `.claude/worktrees/t497` · branch `WT-codex-neutrality` · base `ace1c5440`

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M)
- 기준선: `.moai/reports/t497/measurement.md`(`c9b226b22`) + 이 트리 재측정. 정정 3건은 `spec.md` §A.2.
- SPEC ID 정규식 실행 결과: `PASS`
- 미해결 마커 **0건** (`plan.md` §E): ① 코덱스 능력 부재 측정 가능 여부 → **철회**(트리 안 증거로 해소, `spec.md` §A.2 정정 2) · ② M5 착수 여부 → **운영자 결정 기록으로 전환**(`spec.md` §D)
- status: `draft` → `in-progress` (M1 커밋 시점, manager-develop 소관 유일 전환)

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

### M1 — 파생 근거 명문화 + REQ-CSN-003 문면 정정

측정 트리: `.claude/worktrees/t497` · 착수 HEAD `76263a02e` · base `ace1c5440`.

| # | 명령 | 기대 | 실측 | 판정 |
|---|---|---|---|---|
| ① | `grep -cE '\|[[:space:]](absent\|present)[[:space:]]\|[[:space:]]*$' .moai/reports/t497/capability-absence.md` | 11 | `11` | PASS |
| ② | `grep -cE '\|[[:space:]]absent[[:space:]]\|[[:space:]]*$' .moai/reports/t497/capability-absence.md` | 3 | `3` | PASS |
| ③ | `sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md \| grep -c '^\| [a-z]'` | 3, ② 와 같은 값 | `3` | PASS (② == ③) |
| ④ | `grep -c '현재 측정값 4행' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md` | 0 | `0` (rc 1) | PASS |
| ⑤ | `grep -c '현재 측정값 3행' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md` | 1 | `1` | PASS |
| ⑥ | `diff <(sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md) <(sed -n … internal/template/templates/AGENTS.md)` | 무출력 rc 0 | 무출력, rc 0 | PASS |
| ⑦ | `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` | PASS + headroom 양수 | `always-loaded surface = 74535 tokens (budget 77600, headroom 3065, 17 entries)` / `--- PASS` — 셀렉터 **1매치** | PASS |
| 대조 | `grep -cE '\|[[:space:]](absent\|present)[[:space:]]\|[[:space:]]*$' AGENTS.md` | 0 | `0` | PASS (다른 표를 우연히 세지 않음) |

- 착수 전 실측(RED): ①② 파일 없음 rc 2 · ④ **1** · ⑤ **0**. ③⑥⑦ 은 불변 대조.
- 결속표 행 집합은 **바뀌지 않았다**(3행 유지) — 부재 3건이 이미 실려 있으므로 행 0개 추가가 파생이 지켜진 결과다.
- ④ 의 자기참조 함정 1건 자체 적발: 정정 HISTORY 항목 초안이 「현재 측정값 4행」을 축자 인용해 ④ 가 1, ⑤ 가 2 로 나왔다. 인용을 서술형(「실측 수치 문면을 4행에서 3행으로」)으로 바꿔 셀렉터가 자기 정정 기록을 세지 않게 했다 — plan-audit N1 과 같은 형태를 이 실행에서 재발시킬 뻔한 자리다.

### M2 — 84건 전수 분류

_<pending>_

### M3 — 중립 소스 본문 개정

_<pending>_

### M4 — 골든 재생성과 반경 확인

_<pending>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
