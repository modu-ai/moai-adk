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

산출: `.moai/reports/t497/body-classification.md`. 측정 트리 `.claude/worktrees/t497`, HEAD `7b4ba4491`.

| # | 명령 | 기대 | 실측 | 판정 |
|---|---|---|---|---|
| ① | `grep -c '^\| .*\.toml \| [0-9]' .moai/reports/t497/body-classification.md` | 84 | `84` | PASS |
| ② | `grep -rhoE '<합집합 패턴>' …/*.toml \| wc -l` | 84 (모집단 불변 대조) | `84` | PASS |
| ③ | 분류표 `file:line` 정렬본 ↔ `grep -rnoE '<합집합 패턴>' …/*.toml \| cut -d: -f1-2 \| sort -u` 의 `diff` | 무출력 rc 0, 양변 81 | 양변 `81`, diff 무출력 **rc 0** | PASS |
| 부 | `grep -rn 'AskUserQuestion' …/*.toml \| wc -l` / `grep -rho … \| wc -l` | 3 줄 / 4 발생 | `3` / `4` | PASS (모집단 불변) |

③ 의 표 좌표 추출 명령(축자):
`grep '^| .*\.toml | [0-9]' <표> | awk -F' \\| ' '{sub(/^\| /,"",$1); print $1":"$2}' | sort -u`
— 첫 시도에서 `$2":"$3` 로 잘못 잘라 74줄이 나왔다. 표가 아니라 **추출 명령의 필드 오프셋 오류**였고(선행 `| ` 때문에 `$1` 이 `| <path>`), 고친 뒤 81 ↔ 81 로 정확히 맞물렸다. 잘못된 셀렉터의 빈/짧은 출력을 표의 결함으로 읽지 않도록 두 시도를 함께 기록한다.

집계: verdict **directive 57 / prose 27** · subject **this-agent 57 · orchestrator 17 · n/a 9 · prohibition 1**.
directive 토큰별: `Skill(` 45 · `Agent(` 7 · `TaskUpdate` 3 · `TaskCreate` 1 · `DesignSync` 1.

**AC-CBN-002 세 좌표** — `sync-auditor.toml:131` = prose/`prohibition` · `plan-auditor.toml:146` = prose/`orchestrator` **2행**(한 줄 두 발생) · `super-advisor.toml:62` = prose/`orchestrator`. 각 행에 왜 이 에이전트의 행위가 아닌지를 한 문장으로 적었다.

**`manager-lead.toml` `Agent(` 10줄 전수 판정 (AC-CBN-013 의 N).** N = **7** (`29 37 57 59 172 193 261`), prose 3 (`7 23 130`). `spec.md` §B.4 의 경계 표본 `37 57 59 193` 은 전부 directive 집합 안에 있다. 셀렉터 실측: `grep -cE '^\| [^|]*manager-lead\.toml \| [0-9]+ \| Agent\( \| directive \|' <표>` → `7`.

- **`:172` → directive.** `:57` 이 능력 목록으로 적은 행위를 절차로 다시 적은 줄. 수동태지만 재실행 스폰의 주체는 peer cross-validation 을 오케스트레이션하는 이 에이전트이고, 오케스트레이터가 주어가 아니므로 REQ-CBN-002 상 this-agent 다. `:57` directive / `:172` prose 는 같은 행위에 두 판정을 주는 것이 된다.
- **`:261` → directive.** 이 에이전트의 위임 라우팅 목록 항목이며 화살표 오른쪽이 자기 행동이다. 닮은 `manager-develop.toml:64-66` 이 prose 인 것과 갈리는 축은 **스폰 주체** — manager-develop 은 `Agent` 도구가 없어 그 표가 오케스트레이터 라우팅의 기술이지만, manager-lead 는 카탈로그 유일 Agent-carrier 라 같은 문장이 자기 행위 지시다.
- SPEC 이 남긴 나머지 넷도 닫았다: `:7`(보드 기제 서술, 행위자는 plan 레인) · `:23`(대조표 정의 칸) · `:130`(리프 워커 정의 + CI 가드 사실) = prose, `:29`(리드 자세 문단의 배경 스폰 지시, `:59` 와 같은 행위) = directive.

### M3 — 중립 소스 본문 개정

_<pending>_

### M4 — 골든 재생성과 반경 확인

_<pending>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
