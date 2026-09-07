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

편집 대상은 `internal/template/templates/.claude/agents/moai/*.md` 뿐이다(REQ-CBN-008). 저장소 루트 사본 `.claude/agents/moai/*.md` 는 열지 않았고, 그 경로가 변경 목록에 나타나지 않는 것이 그 증거다.

| 소스 좌표 | 처분 |
|---|---|
| `manager-develop.md:103,128` | `task-list` 능력 이름 + 부재 시 산문 보고로 고쳐 씀 |
| `e2e-tester.md:146` | 같음 |
| `manager-design.md:115` | 사다리 문면 **유지**, 직후에 `design-sync` 부재 시 행동 1문단 추가 |
| `manager-lead.md:36,44,64,66,179,200,268` | M2 가 directive 로 판정한 7줄 전부 — `subagent-spawn` 능력 이름을 부르게 고쳐 씀. **줄 병합 없음**(7줄 그대로 7줄). `:44` 뒤에 harness note 1문단 추가 |
| `AGENTS.md` 두 사본 | 결속표 문단에 `Skill("<name>")` 덮개 1문단. 두 사본 **바이트 동일** |
| `invoke Skill(` 41줄 | **손대지 않음** |

소스↔TOML 대응(위치 짝짓기, 이 실행 실측): `36↔29 · 44↔37 · 64↔57 · 66↔59 · 179↔172 · 200↔193 · 268↔261`.

**개정 중 자체 적발 1건.** 첫 초안이 `task-list` 줄에 "(Claude harness: `TaskUpdate`)" 라는 대응 표기를 남겼는데, 그 표기가 그대로 TOML 에 실려 AC-CBN-005 의 `Task* → 0` 을 깨뜨린다. 능력 이름만 남기고 도구 토큰을 뺐다. `subagent-spawn` 쪽은 `Agent(` 에 0-기대값이 없으므로 "(Claude harness: `Agent(...)`)" 표기를 유지했다 — 두 부류의 처분이 갈리는 이유는 AC 가 다르기 때문이지 원칙이 달라서가 아니다.

M4 재생성 **후**의 TOML 실측:

| # | 명령 | 기대 | 실측 | 판정 |
|---|---|---|---|---|
| ① | `grep -rhoE 'Task(Create\|Update\|List\|Get)' …/*.toml \| wc -l` | 0 (발생) | `0` | PASS |
| ② | `grep -rn 'task-list' …/*.toml \| wc -l` | ≥3 (줄) | `3` — `e2e-tester.toml:139` · `manager-develop.toml:90` · `manager-develop.toml:115` | PASS |
| ③ | `grep -c 'subagent-spawn' …/manager-lead.toml` | ≥ N (=7) | `8` (7 지시 줄 + harness note) | PASS (M ≥ N) |
| ④ | `grep -oE '(^\|[^/])design-sync' …/manager-design.toml \| wc -l` | ≥1 (발생) | `3` | PASS |
| ⑤ | `grep -c 'default = DesignSync tool push' …/manager-design.toml` | 1 불변 | `1` | PASS |
| ⑥ | `grep -c '\.agents/skills' AGENTS.md` | ≥1 | `2` | PASS |
| ⑦ | `grep -c '\.agents/skills' internal/template/templates/AGENTS.md` | ⑥ 과 같은 값 | `2` | PASS |
| ⑧ | `grep -rhoE 'invoke Skill\(' …/*.toml \| wc -l` | 41 불변 | `41` | PASS |
| ⑨ | `grep -rhoE 'AskUserQuestion' …/*.toml \| wc -l` | 4 불변 | `4` | PASS |

착수 전 실측(RED): ① 4 · ② 0 · ③ 0 · ④ 0 · ⑥ 0 · ⑦ 0. ⑤⑧⑨ 는 전면 치환을 막는 불변 대조.

개정 후 합집합 모집단은 84 → **82** 이며 그 차가 전부 귀속된다: `Task*` 4 → 0 (−4), `DesignSync` 6 → 7 (+1, 부재 시 행동 문단), `Agent(` 25 → 26 (+1, harness note), `Skill(` 45 불변, `AskUserQuestion` 4 불변 → 45+26+0+7+4 = 82.

### M4 — 골든 재생성과 반경 확인

- 재생성: `AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/...`
- **첫 실행은 FAIL 했고 그것이 정상이다.** `TestEmbedFSPresenceAndByteEquality` 가 4본에 대해 "embedded bytes differ from committed" 를 냈는데, 임베드 바이트는 **컴파일 시점**에 굳고 파일 재작성은 그 뒤에 일어나므로 같은 실행 안에서는 어긋날 수밖에 없다. 재컴파일되는 다음 실행에서 해소됐다 — 임베드 축과 emit 축은 서로 다른 질문에 답한다.

| # | 명령 | 기대 | 실측 | 판정 |
|---|---|---|---|---|
| ① | `go test ./internal/template/agentemit/...` (UPDATE 없이) | PASS | `ok  github.com/modu-ai/moai-adk/internal/template/agentemit` | PASS |
| ② | 변경 경로 집합 ↔ `spec.md` §C.5 11줄 `diff` | 무출력 rc 0 (집합 동일성) | 양변 **11**, diff 무출력 **rc 0** | PASS |
| ③ | ② 의 두 필터가 걸러낸 나머지 | 전부 두 증거 접두 | 4개 전부 `.moai/reports/t497/` 또는 `.moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` | PASS |

**② 의 모집단은 「미커밋 변경」이 아니라 「이 카드가 만든 변경 전체」다.** 증거를 재측정 앞에 커밋하는 규율(`acceptance.md` §F) 때문에 M1·M2 산출이 이미 커밋돼, 작업 트리 상태만 보면 10줄이 나오고 11번째(`SPEC-CODEX-SKILL-NEUTRAL-001/spec.md`)가 빠진 것처럼 읽힌다. 그래서 카드 착수 HEAD `76263a02e` 이후의 추적 파일 변경 목록과 현재 작업 트리 상태를 합집합으로 놓고 대조했다 — 11 ↔ 11, diff rc 0.

경로 추출에서 오프셋 오류를 두 번 만났다(M2 ③ 의 74 ≠ 81, M4 ② 의 21 ≠ 11). 두 번째는 상태 목록의 `XY path` 형식(공백+M+공백)을 접두 제거 정규식이 못 벗겨 생겼고, `plan.md` §F M4 ② 가 처음부터 지정한 `awk '{print $NF}'` 로 바꾸자 11줄이 됐다. **두 번 다 실패가 시끄러웠다** — 잘못된 셀렉터가 조용한 초록을 내지 않고 눈에 띄는 불일치를 냈다는 것이 기록할 만한 성질이며, 그래서 두 시도를 모두 남긴다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: pending-backfill
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run  # 레인은 통합 브랜치를 push 하지 않는다 — 통합은 창에서 리드 소관
l44_post_push_fetch: not-run   # 같은 이유
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-applicable
  note: "Go 소스 변경 0 — 문서·템플릿·생성물 변경뿐이다. GOOS 매트릭스 빌드는 판정 대상이 없어 돌리지 않았다(Gap 에 명시)."
total_run_phase_files: 15   # 산출 11 + 증거 4
m1_to_mN_commit_strategy: "마일스톤별 분리 커밋 (M1 / M2 / M3+M4). 증거는 재측정 앞에 커밋."
```

### AC PASS/FAIL 행렬 (AC-CBN-001..014)

| AC | 판정 | 검증 명령 | 축자 출력 |
|---|---|---|---|
| AC-CBN-001 | PASS | 분류 행 수 / 모집단 / 좌표 diff | `84` / `84` / diff 무출력 rc 0 (양변 81) |
| AC-CBN-002 | PASS | 분류표 세 좌표 행 판독 + 모집단 재측정 | `sync-auditor:131` prose·`prohibition` · `plan-auditor:146` prose·`orchestrator` **2행** · `super-advisor:62` prose·`orchestrator`. 모집단 `3` 줄 / `4` 발생 |
| AC-CBN-003 | PASS | `grep -rhoE 'invoke Skill\(' … \| wc -l` | `41` |
| AC-CBN-004 | PASS | `grep -rhoE 'AskUserQuestion' … \| wc -l` | `4` |
| AC-CBN-005 | PASS | `Task*` 계수 + `task-list` 좌표 | `0` / 좌표 `3` 개 |
| AC-CBN-006 | PASS | (a)(b)(c)(d) 를 한 판정에서 | `11` / `3` / `3` / `0` — (b)==(c) 성립 |
| AC-CBN-007 | PASS | 두 사본 표 구역 `diff` | 무출력, rc 0 (파일 바이트도 양쪽 15415) |
| AC-CBN-008 | PASS | 표 첫 칸 ↔ `tool_classes` 값 집합 | `question-channel` · `task-list` · `design-sync` — 셋 다 원소, 새 어휘 0 |
| AC-CBN-009 | PASS | `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` | `always-loaded surface = 74695 tokens (budget 77600, headroom 2905, 17 entries)` / `--- PASS`. **셀렉터 1매치**. 덮개 비용 160 tokens, 여유 양수 유지 |
| AC-CBN-010 | PASS | `go test ./internal/template/agentemit/...` | `ok  github.com/modu-ai/moai-adk/internal/template/agentemit` |
| AC-CBN-011 | PASS | 반경 집합 동일성 | 11 ↔ 11, diff 무출력 rc 0. 루트 사본 `.claude/agents/moai/*.md` 는 목록에 없음 |
| AC-CBN-012 | PASS | 실행한 명령 목록 판독 | `./internal/template/agentemit/...` 와 `./internal/config/` **두 범위만**. 전체 스위트 형태 **0회** |
| AC-CBN-013 | PASS | N / M 을 한 판정에서 | `N=7` (`29 37 57 59 172 193 261`, 경계 표본 `37 57 59 193` 전부 포함) · `M=8` ≥ N · 착수 전 `M=0` |
| AC-CBN-014 | PASS | 부재 시 행동 / 사다리 불변 | `3` (착수 전 0) / `1` |

부가 가드(같은 `./internal/config/` 범위, 셀렉터 1매치): `TestCodexContractByteCeiling` → `contract document AGENTS.md = 15415 bytes (ceiling 24576, headroom 9161)`, 템플릿 사본 동일값 → PASS. plan-audit A4 는 「이 가드가 고정 검증 범위에서 돌지 않는다」고 적었으나 **그 전제가 틀렸다** — 가드는 `internal/config` 안에 있어 고정 범위가 도달한다. 다만 `-run 'TestAlwaysLoadedTokenBudget$'` 셀렉터로는 선택되지 않으므로 명시적으로 함께 돌렸다.

기타: 미해결 마커 `grep -rnE '\[NEEDS[[:space:]]CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` → 무출력 **rc 1**. `gofmt -l internal/` → 무출력 rc 0 (Go 소스 변경 0).

### Gaps

- 코덱스 런타임 실거동을 이 트리에서 프로브하지 않았다. M1 의 부재 판정은 매니페스트 rationale **문면** 판정이다.
- 크로스 플랫폼 빌드(`GOOS=windows`)를 돌리지 않았다 — Go 소스 변경이 0 이라 판정 대상이 없다.
- 전체 스위트를 돌리지 않았다(REQ-CBN-014 의 요구). 전 패키지 판정은 CI 몫이다.
- M5(미러 스킬 77파일)는 착수하지 않았다 — 운영자가 후속 카드 분리로 확정.
- plan-audit 잔여 부채 A3·A5·A6·A7·A8·A9·A10 은 손대지 않았다(전부 optional, 감사가 FAIL 근거가 아니라고 명시). 그중 A8(형제 SPEC 의 세 번째 스테일 「4행」)은 v0.3.1 HISTORY 항목 안에 있어 **이력 재작성이 되므로 의도적으로 두었다**.

### Residual-risk

- `manager-lead.toml:172` 를 directive 로 판정한 것은 수동태 문장의 행위자 귀속 판단이다. 틀렸다면 산문이 능력 이름을 부르게 된 것이고(무해하지만 불필요), 반대 방향의 실패는 코덱스에 Claude 스폰을 지시하는 줄이 남는 것이다 — 불확실할 때 directive 로 기우는 쪽이 보수적이다.
- 덮개 문장이 always-loaded 예산을 160 tokens 소비했다(여유 3065 → 2905). 예산은 여전히 양수지만 이 표면의 추가 확장은 여유를 다시 재고 시작해야 한다.
- `.agents/skills` 미러 도달 주장은 `skill_mirror.go` 의 자기 서술("makes every skill this run deployed reachable from BOTH paths")과 `agents-codex.yaml` 의 skill-roots 측정 기록에 근거한다. 이 트리에서 배포를 실행해 미러를 눈으로 확인하지는 않았다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: pending-backfill-sync
sync_status: complete
b12_self_test_a: pass   # grep -c 'SPEC-CODEX-BODY-NEUTRALITY-001' CHANGELOG.md → 0 (착수 전, 중복 없음)
b12_self_test_b: pass   # acceptance.md 고유 AC 식별자 14 == CHANGELOG 기재 14 (0 이 아님을 확인)
b12_self_test_c: pass   # CHANGELOG 가 이름 붙인 경로 전수 ls 확인 — 9/9 존재
changelog_entry_position: "[Unreleased] › ### Added — 말미 append (다른 레인 항목 무손상)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (updated: 2026-09-07)"
  plan_md: "n/a — 프론트매터 없음(status 축 stateless)"
  acceptance_md: "n/a — 프론트매터 없음(status 축 stateless)"
  progress_md: "n/a — 상태를 본문 절로 기록"
canary_compliance_check:
  applicable: false
  note: "이 SPEC 은 자기 sync 가 시험할 전향 정책을 정의하지 않는다."
```

### 이 커밋이 하는 일

`spec.md` 프론트매터의 `in-progress → completed` 전환(`status` + `updated` 두 필드만) · 이 §E.4 신호 · `CHANGELOG.md` `[Unreleased]` 항목 1건. `plan.md` · `acceptance.md` 는 프론트매터를 갖지 않아 status 축에서 stateless 이므로(`spec-frontmatter-schema.md` § Artifact Statelessness) 전환 대상이 아니다. MX Tag 검증은 별도 단계가 아니라 이 sync 안의 하위 단계다 — Go 소스 변경이 0 이라 새로 붙일 `@MX` 주석이 없고, 개정된 4본은 전부 마크다운 본문이다.

### 14개 AC 재검증 (sync-phase, 이 트리 · HEAD `321111fe5` · 무캐시)

run-phase 판정을 인용하지 않고 sync 시점에 다시 쟀다. 명령과 축자 출력:

| # | 명령 | 실측 | 대응 AC |
|---|---|---|---|
| 1 | `go test -count=1 ./internal/template/agentemit/...` | `ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.296s` rc 0 | AC-CBN-010 |
| 2 | `go test -count=1 ./internal/config/` | `ok  github.com/modu-ai/moai-adk/internal/config  1.457s` rc 0 | AC-CBN-009 |
| 3 | `go test -count=1 -run TestAlwaysLoadedTokenBudget -v ./internal/config/` | `always-loaded surface = 74695 tokens (budget 77600, headroom 2905, 17 entries)` / `PASS` | AC-CBN-009 |
| 4 | `go test -count=1 -run TestCodexContractByteCeiling -v ./internal/config/` | `AGENTS.md = 15415 bytes (ceiling 24576, headroom 9161)`, 템플릿 사본 동일값 / `--- PASS` | 부가 가드 |
| 5 | `grep -rhoE 'invoke Skill\(' …/*.toml \| wc -l` | `41` (불변) | AC-CBN-003 |
| 6 | `grep -rhoE 'AskUserQuestion' …/*.toml \| wc -l` | `4` (불변) | AC-CBN-004 |
| 7 | `grep -rhoE 'Task(Create\|Update\|List\|Get)' …/*.toml \| wc -l` | `0` | AC-CBN-005 |
| 8 | `grep -rn 'task-list' …/*.toml \| wc -l` | `3` | AC-CBN-005 |
| 9 | `grep -c 'subagent-spawn' …/manager-lead.toml` | `8` (≥ N=7) | AC-CBN-013 |
| 10 | `grep -oE '(^\|[^/])design-sync' …/manager-design.toml \| wc -l` | `3` (착수 전 0) | AC-CBN-014 |
| 11 | `grep -c 'default = DesignSync tool push' …/manager-design.toml` | `1` (불변) | AC-CBN-014 |
| 12 | `grep -cE '\|[[:space:]](absent\|present)[[:space:]]\|[[:space:]]*$' capability-absence.md` | `11` | AC-CBN-006 (a) |
| 13 | `grep -cE '\|[[:space:]]absent[[:space:]]\|[[:space:]]*$' capability-absence.md` | `3` | AC-CBN-006 (b) |
| 14 | `sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md \| grep -c '^\| [a-z]'` | `3` — (b) == (c) 성립 | AC-CBN-006 (c) |
| 15 | `grep -c '현재 측정값 4행' …/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md` | `0` | AC-CBN-006 (d) |
| 16 | 두 `AGENTS.md` 사본 결속표 구역 `diff` | 무출력, rc 0 | AC-CBN-007 |
| 17 | `grep -c '^\| .*\.toml \| [0-9]' body-classification.md` | `84` | AC-CBN-001 (a) |
| 18 | 개정 후 합집합 모집단 `grep -rhoE '<합집합 패턴>' …\| wc -l` | `82` (§E.2 M3 의 귀속과 일치: 45+26+0+7+4) | 모집단 대조 |
| 19 | `grep -rnE '\[NEEDS[[:space:]]CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` | 무출력, rc 1 | 완료 정의 |

AC-CBN-002 · 008 · 011 · 012 는 명령 재실행이 아니라 판독으로 닫힌다 — 각각 분류표 세 좌표의 행 판독(§E.2 M2), 결속표 첫 칸 값이 `tool_classes` 원소인지의 대조, 반경 집합 동일성(§E.2 M4 ②), 실행 명령 목록의 판독이다. sync-phase 에서 실행한 범위는 `./internal/template/agentemit/...` 와 `./internal/config/` **둘뿐**이며 `go test ./...` 형태는 0회다(AC-CBN-012 유지).

### 좌표 인용 규율 (이 절이 새로 쓰는 인용)

이 카드는 좌표 오류를 3회 냈고(그중 1건은 감사자 적발), 마지막 라운드는 좌표 대신 절 번호를 인용해 4번째를 막았다. 그래서 이 절은 **움직일 수 있는 `file:line` 을 새로 만들지 않는다** — 심볼·절 번호로 건다.

- 알려진 잔존 1건: `.moai/reports/t497/measurement.md` 가 미러 도달 경로를 `internal/template/skill_mirror.go:4-5,52,171` 로 인용한다. 이 트리 실측에서 `:171` 은 `mirrorSkills` 의 `@MX:ANCHOR` 주석 줄이고, `func (d *deployer) mirrorSkills` 자체는 `:178` 이다. develop 흡수 후 좌표는 밀린다(t503 이 앞쪽에 줄을 넣었고, 그 커밋 자신이 미러 거동은 불변이라 적는다) — **실질 전제는 그대로다.** 그 파일은 plan-phase 증거 기록이므로 여기서 다시 쓰지 않고, 심볼 앵커(`skill_mirror.go` 의 `mirrorSkills` 와 그 `@MX:ANCHOR` 주석 · 파일 머리말 4-5행의 자기 서술 · `mirrorSkillsRelDir` 상수)를 이 줄에 남겨 다음 독자가 좌표 없이 찾게 한다.

### docs-site 판정 — 페이지를 만들지 않는다

`docs-site/content/{ko,en,ja,zh}` 4로케일이 이 트리에 있으나 **이 카드는 그 어느 것도 건드리지 않는다.** 근거 셋: (1) 사용자에게 보이는 거동 변화가 0 이다 — 바뀐 것은 에이전트 본문의 문면과 그 생성물 TOML 이고, CLI 표면·설정 키·명령 결과 중 무엇도 달라지지 않는다. (2) 반경이 전부 하네스 내부 계약(`AGENTS.md` 결속표와 에이전트 정의)이며, docs-site 는 그 계약을 페이지로 싣고 있지 않다. (3) 4로케일 의무는 페이지를 **만들기로 정했을 때** 발생하는 후속 의무이지 페이지를 만들 사유가 아니다. 따라서 `hugo build` 도 돌리지 않았고, 돌렸다고 주장하지 않는다.

### Gaps

- **`§E.3` 의 `run_commit_sha: pending-backfill` 은 이 커밋이 채우지 않는다.** 그 필드는 run-phase 소유 면(`§E.3`)이고 manager-docs 의 금지 반경 안이다(`spec-frontmatter-schema.md` § SHA placeholder backfill exemption 은 **해당 phase 를 소유한 에이전트**에게만 backfill 을 허용한다). 값 자체는 이미 알려져 있다 — run 커밋은 `7b4ba4491`(M1) · `c3ea4670e`(M2) · `321111fe5`(M3+M4) 셋이다. 소관 밖이라 적지 않을 뿐이므로, 리드가 이 사실과 함께 backfill 을 배차하면 한 줄로 닫힌다.
- `sync_commit_sha` 는 이 커밋 안에서 자기 해시를 인용할 수 없어 canonical placeholder 로 남는다. 해소값은 완료 보고로 리드에게 전달한다.
- 코덱스 런타임 실거동은 sync-phase 에서도 프로브하지 않았다(§E.2 Gaps 와 같다). 능력 부재 판정은 여전히 매니페스트 rationale **문면** 판정이다.
- 크로스 플랫폼 빌드 미실행 — Go 소스 변경 0.
- 전체 스위트 미실행 — REQ-CBN-014 의 요구이며 전 패키지 판정은 CI 몫이다.
- M5(미러 스킬 77파일)는 이 카드에 들어 있지 않다. 후속 카드 `t523` 소관이며, `spec.md` §B.6 · §D 는 그 카드가 차갑게 읽을 대상이라 이 sync 에서 손대지 않았다.

### Residual-risk

- 덮개 문장이 always-loaded 예산을 160 tokens 썼다(여유 3065 → 2905, 이 실행 재측정으로 확인). 예산은 양수지만 이 표면의 다음 확장은 여유를 다시 재고 시작해야 한다.
- `manager-lead` 의 지시 줄 개정은 「(Claude harness: `Agent(...)`)」 대응 표기를 남긴다. 코덱스 독자에게는 정보이지 지시가 아니지만, `Agent(` 토큰이 본문에 남는다는 사실 자체는 향후 같은 축의 카드가 0-기대값을 세울 때 걸림돌이 될 수 있다 — `Task*` 쪽이 0-기대값 때문에 대응 표기를 뺀 것과 갈리는 지점이며, 갈린 이유는 원칙이 아니라 AC 다(§E.2 M3).
- `.agents/skills` 미러 도달 주장은 이 트리에서 배포를 실행해 눈으로 확인한 것이 아니다(§E.2 Residual-risk 와 같다).
