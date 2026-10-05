# SPEC-CODEX-BODY-NEUTRALITY-001 — 구현 계획

> 마일스톤은 **되돌리기 어려움 순**으로 배치했다. M1·M2 는 판단이 바뀔 여지가 가장 큰 결정(결속표 행 파생, 84건 분류)이고, M4 는 기계적 재생성이다. 시간 추정은 쓰지 않는다 — 우선순위와 단계 순서만 쓴다.

---

## §A. 맥락

`.codex/agents/moai/*.toml` 11본은 `internal/template/agentemit/writer.go` `renderTOML` 이 중립 소스 `.claude/agents/moai/*.md` 본문을 축자로 실어 만든 생성물이다. 개정은 소스에 들어가고 골든 재생성으로 TOML 에 반영된다. 기준선·정정 3건·설계 판정은 `spec.md` §A·§B.

**이 카드가 다루는 것은 본문 축 하나뿐이다.** `tools:` 프론트매터 축과 MCP/effort/sandbox 축은 매니페스트가 이미 처분했다(`spec.md` §A.4 / §D).

---

## §B. 알려진 문제

- **B-1 기준선 합계 착오.** `.moai/reports/t497/measurement.md` 의 `= 74` 는 덧셈 착오다. 파일별 값은 정확히 재현되며 합은 **84**(발생 수). 서로 다른 줄은 **81**. 이 카드의 모든 판정은 **발생 84** 를 단위로 한다.
- **B-2 결속표 3행 vs 문서 4행 — 해소됨.** 실린 표는 3행, 완결 SPEC 기록은 4행이었다. **스테일한 쪽은 문서다.** 근거는 `spec.md` §A.2 정정 2 의 세 갈래: 후보 표 `.moai/reports/t196/csn003-table-4row.txt` 의 4번째 행은 `cross-session-messaging` 이고, 그 클래스의 rationale 은 "The Codex counterpart rides the moai MCP broker" 라 대응물의 **존재**를 말하며, `tool_classes` 11개 전수 대조도 부재 3건(`task-list` · `design-sync` · `question-channel`)에서 멈춘다. **코덱스 프로브는 필요 없다** — M1 은 이 파생을 기록하고 REQ-CSN-003 문면을 정정하는 문서 작업이다.
- **B-3 `.md` 전수 grep 의 과다계상.** 같은 패턴을 `.claude/agents/moai/*.md` 에 돌리면 `Task*` 가 4 가 아니라 **48** 이다. 차이는 프론트매터 `tools:` CSV 이고 emitter 는 프론트매터를 싣지 않는다. 계수는 반드시 TOML 본문에서.
- **B-4 골든 반경 오염.** 재생성은 이 카드 밖의 `.md` 변경까지 함께 실어 나를 수 있다. 커밋 전 `git status --short` 필수.

---

## §C. 사전 점검

```
git rev-parse --show-toplevel   # → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t497
git branch --show-current       # → WT-codex-neutrality
git status --short              # 착수 시 추적 파일 수정 0 이어야 반경 귀속이 성립
```

---

## §D. 제약

- 검증은 손댄 패키지만: `go test ./internal/template/agentemit/...`. 로컬 전체 스위트 금지(CLAUDE.local.md §4).
- Template-First: 템플릿 사본이 소스다. `AGENTS.md` 는 루트/템플릿 두 사본을 **함께** 고친다.
- TOML 손편집 금지. 골든 재생성만이 인가된 이동 경로.
- 시간 추정 금지. 우선순위 라벨과 단계 순서만.

---

## §E. 미해결 질문 — 없음

**미해결 마커 0건.** 앞선 라운드의 2건은 다음과 같이 처분됐고, 어느 쪽도 마커로 남지 않는다.

- **① 코덱스 능력 부재 측정 가능 여부 — 철회.** 물음 자체가 사라졌다. 부재 판정은 코덱스 프로브가 아니라 `agents-codex.yaml` rationale 을 판별식으로 삼는 문면 대조로 답해지며, 그 대조는 이미 끝났다(§B-2 · `spec.md` §A.2 정정 2). 운영자가 결정할 것이 남아 있지 않으므로 마커가 아니라 **결론**이다.
- **② M5 미러 스킬 77파일 착수 여부 — 운영자가 결정했고 착지했다(2026-09-07).** 트리 증거로 결정되지 않는 진짜 범위 결정이었고, 리드 세션의 `AskUserQuestion` 라운드에서 **분리**로 확정됐다. 기록은 `spec.md` §D 마지막 절이며, 후속 카드는 리드가 발행한다. **상신할 것이 남아 있지 않으므로 레인은 이 건을 Implementation Kickoff Approval 에 올리지 않는다.**

검증(이 실행에서 실행): `grep -rnE '\[NEEDS[[:space:]]CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` → **무출력, rc 1**. 셀렉터는 **마커 모양**(`[NEEDS` + 공백 + `CLARIFICATION`)을 요구하고, 여기에 적힌 형태는 그 사이가 `[[:space:]]` 라 **자기 자신에 매치하지 않는다**. 종전의 `grep -rn 'NEEDS'` 는 이 검증 문장 자신을 세어 3줄을 냈다 — §G AP-6(자기참조 수치)을 이 SPEC 의 자기검증에 그대로 저지른 꼴이었다. 수리 전 마커 실측(v0.1.0): `plan.md:45` · `plan.md:46` 2건.

---

## §F. 마일스톤

> **[HARD] 결속표 행 수의 셀렉터와 기대값은 하나뿐이다.** 이 문서·`spec.md`·`acceptance.md` 어디서나 같은 명령, 같은 값을 쓴다.
>
> ```
> sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md | grep -c '^| [a-z]'
> ```
>
> **기대 출력 = 3** (착수 전 실측 3, 두 사본 동일). 이 셀렉터는 (a) `sed` 로 결속표 구역에만 범위를 좁혀 파일 다른 곳의 `| ` 줄에 흔들리지 않고, (b) `^| [a-z]` 라 헤더(`| Capability` — 대문자 C)와 구분자(`|---`)를 둘 다 제외해 **데이터 행만** 센다. 종전의 `grep -c '^| '` 는 헤더를 포함해 3행 표에 4를 냈고, 그 한 자리가 `4`/`3`/`5` 세 기대값이 갈린 원인이다. **그 셀렉터는 이 SPEC 에서 쓰지 않는다.**

### M1 — 파생 근거 명문화와 REQ-CSN-003 문면 정정 (우선순위 High · 되돌리기 가장 어려움)

**프로브가 아니다.** `spec.md` §A.2 정정 2 가 트리 안 증거만으로 부재 3건을 확정했으므로, M1 이 하는 일은 그 파생을 기록으로 남기고 어긋난 문면을 고치는 것이다. **결속표의 행 집합은 바뀌지 않는다.**

- 산출 1: `.moai/reports/t497/capability-absence.md` — `tool_classes` 값 집합 **11개 전수**를 한 행씩. 열은 `| class | rationale 인용 | verdict |`, `verdict ∈ {absent, present}`. 모집단은 손으로 열거하지 않고 명령으로 뽑는다:
  ```
  sed -n '/^tool_classes:/,/^$/p' internal/template/agentemit/agents-codex.yaml \
    | grep -oE ': [a-z-]+$' | sed 's/^: //' | sort -u
  ```
  **[HARD] 트랩 1건 — `moai-mcp` 는 `present` 로 판정된다.** 11개 중 유일하게 rationale 이 `unavailable` 이라는 낱말을 실은 채 능력이 존재하는 클래스다(부재한 것은 서버 안 도구별 필터링). 판정은 낱말이 아니라 `spec.md` §A.2 정정 2 의 verdict 판별식으로 한다 — 낱말로 키를 잡으면 아래 ② 가 **3 이 아니라 4** 를 내고 ③ 과 갈린다.
- 산출 2: `SPEC-CODEX-SKILL-NEUTRAL-001` `spec.md` REQ-CSN-003 문면의 **"현재 측정값 4행" → "현재 측정값 3행"** 정정 + 그 SPEC HISTORY 에 Amendments 1행(정정 근거와 이 SPEC ID). 요구사항의 **의미는 바꾸지 않는다** — 파생 기준 문장은 그대로 두고 스테일한 실측 수치만 갈아쓴다. 같은 파일 `:266` 의 「4행 = 373 B」는 후보 표의 크기 측정 기록이므로 **손대지 않는다.**

검증 명령 · 기대 출력(사전 고정). **①②④⑤ 는 착수 전 트리에서 기대 출력을 내지 못한다 — 아무것도 하지 않으면 M1 은 통과할 수 없다.**

| # | 명령 | 기대 출력 | 착수 전 실측 |
|---|---|---|---|
| ① | `grep -cE '\|[[:space:]](absent\|present)[[:space:]]\|[[:space:]]*$' .moai/reports/t497/capability-absence.md` | **11** | 파일 없음 — rc=2, 무출력 (RED) |
| ② | `grep -cE '\|[[:space:]]absent[[:space:]]\|[[:space:]]*$' .moai/reports/t497/capability-absence.md` | **3** | 파일 없음 — rc=2, 무출력 (RED) |
| ③ | `sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md \| grep -c '^\| [a-z]'` | **3**, 그리고 ② 와 같은 값 | 3 (불변 대조) |
| ④ | `grep -c '현재 측정값 4행' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md` | **0** | **1** (RED) |
| ⑤ | `grep -c '현재 측정값 3행' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md` | **1** | **0** (RED) |
| ⑥ | `diff <(sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md) <(sed -n '/^\*\*Capability bindings/,/^---$/p' internal/template/templates/AGENTS.md)` | 무출력, rc 0 | 무출력, rc 0 (불변 대조) |
| ⑦ | `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` | `PASS` + 로그 `headroom` 양수 | PASS, `budget 77600, headroom 3065` (불변 대조) |

**①②의 셀렉터는 「verdict 열이 마지막 칸」이라는 표 모양에만 의존한다** — 헤더 대소문자나 열 이름에 기대지 않는다. `^| [a-z]` 로 데이터 행을 세면 헤더가 소문자로 시작할 때(`| class | …`) 헤더까지 세어 12가 나오는데, 그것이 §F 머리말의 `4`/`3` 착오와 같은 종류의 실수다. 산출 문서의 표는 **verdict 를 마지막 칸에 둔다**(`| … | absent |` / `| … | present |`). 대조군: 같은 셀렉터를 `AGENTS.md` 에 돌리면 0 — 다른 표를 우연히 세지 않는다(이 트리 실측).

**②와 ③은 같은 판정에서 함께 잰다.** 파생 기록의 부재 건수와 실린 표의 데이터 행 수가 일치해야 파생이 지켜진 것이고, 한쪽만 재면 둘이 갈라진 경우를 못 잡는다.

### M2 — 84건 전수 분류 (우선순위 High)

- 산출: `.moai/reports/t497/body-classification.md` — 84행 표. 열은 `file` · `line` · `token` · `verdict(directive|prose)` · `subject(this-agent|orchestrator|prohibition|n/a)` · `rationale`. `subject` 는 **세 값 + n/a** 다(`spec.md` §B.4 — 금지 서술을 두 값 중 하나로 밀어 넣으면 분류가 거짓이 된다).
- **[HARD] `token` 열은 ② 합집합 패턴의 리터럴 토큰을 그대로 싣는다** — `Agent(` · `Skill(` · `AskUserQuestion` · `DesignSync` · `TaskCreate` · `TaskUpdate` · `TaskList` · `TaskGet`. AC-CBN-013 의 N 셀렉터가 이 표기에 묶여 있다(다른 표기를 쓰면 N 이 0 이 되어 그 AC 가 먼저 깨진다).
- 판정 규칙은 `spec.md` §B.4.

| # | 명령 | 기대 출력 | 착수 전 실측 |
|---|---|---|---|
| ① | `grep -c '^\| .*\.toml \| [0-9]' .moai/reports/t497/body-classification.md` | **84** | 파일 없음 — rc=2 (RED) |
| ② | `grep -rhoE 'AskUserQuestion\|TaskCreate\|TaskUpdate\|TaskList\|TaskGet\|DesignSync\|Skill\(\|Agent\(' internal/template/templates/.codex/agents/moai/*.toml \| wc -l` | **84** (모집단 불변 대조 — ① 과 같은 판정에서) | 84 |
| ③ | 분류표 `file:line` 열 정렬본과 `grep -rnoE '<② 의 합집합 패턴>' …/*.toml \| cut -d: -f1-2 \| sort -u` 를 `diff` | 무출력, rc 0 (**81** 개 서로 다른 줄) | 우변만 존재 — 81줄 (RED) |

③ 이 좌표 대응 보강이다. ① 은 「84행이 있다」만 말하므로 84행이 모두 같은 좌표를 가리켜도 통과한다. ③ 은 분류표가 **실제 모집단의 그 줄들을** 덮었는지를 본다. 발생 84 vs 줄 81 의 차는 두 토큰을 싣는 3줄이며, 그 3줄은 표에서 2행씩 차지하되 좌표 집합으로는 1개다.

`grep -c 'orchestrator' …` 형태는 쓰지 않는다 — 산문 본문에도 그 낱말이 나오므로 공허하다. `subject` 는 열 값으로 센다.

### M3 — 중립 소스 본문 개정 (우선순위 High)

M2 가 `directive` 로 판정한 항목만 고친다. **편집 대상은 `internal/template/templates/.claude/agents/moai/*.md` 뿐이다**(REQ-CBN-008 — 저장소 루트 사본이 아니다).

| 소스 좌표 | 처분 |
|---|---|
| `manager-develop.md:103,128` (Task\*) | `task-list` 능력 이름으로 고쳐 씀 |
| `e2e-tester.md:146` (Task\*) | 같음 |
| `manager-design.md:115` 우선순위 사다리 | 사다리 **문면 유지** + `design-sync` 부재 시 행동 1문단 추가 |
| `manager-lead.md` 에서 **M2 가 `directive` 로 판정한 `Agent(` 줄 전부** | `subagent-spawn` 능력 이름을 부르도록 고쳐 씀. 좌표가 아니라 **성질**로 정의된 부류다(`spec.md` §B.4 · REQ-CBN-009). §B.4 경계 표본 = 소스 `44,64,66,200`(TOML `37,57,59,193`); M2 가 반드시 판정할 추가 후보 = 소스 `179,268`(TOML `172,261`). **줄을 병합하지 않는다**(AC-CBN-013 (b)). **결속행 참조는 만들지 않는다** — `subagent-spawn` 은 능력 존재이므로 행이 없다 |
| `AGENTS.md` 두 사본 결속표 문단 | `invoke Skill(` 41줄의 덮개 1문장 추가(`.agents/skills/<name>/SKILL.md`). 두 사본을 **같은 내용으로** |
| `invoke Skill(` 41줄 자체 | **손대지 않음** |

검증 대상은 M4 재생성 **후**의 TOML 이다(생성물이 소스 개정을 반영했는지가 판정 대상이므로).

| # | 명령 | 기대 출력 | 착수 전 실측 |
|---|---|---|---|
| ① | `grep -rhoE 'Task(Create\|Update\|List\|Get)' internal/template/templates/.codex/agents/moai/*.toml \| wc -l` | **0** (발생 단위) | 4 발생 / 3 줄 (RED) |
| ② | `grep -rn 'task-list' internal/template/templates/.codex/agents/moai/*.toml \| wc -l` | **3 이상** (줄 단위) | **0** (RED) |
| ③ | `grep -c 'subagent-spawn' internal/template/templates/.codex/agents/moai/manager-lead.toml` | **N 이상** (줄 단위 — N = M2 분류표의 `manager-lead.toml` · `Agent(` · `directive` 행 수, N ≥ 4. 상수 4 가 아닌 이유는 AC-CBN-013) | **0** (RED) |
| ④ | `grep -oE '(^\|[^/])design-sync' internal/template/templates/.codex/agents/moai/manager-design.toml \| wc -l` | **1 이상** (발생 단위) | **0** — 기존 4건은 전부 `/design-sync` 슬래시 커맨드 (RED). `(^\|` 갈래는 줄 머리 발생을 놓쳐 생기는 거짓 RED 를 막는다(AC-CBN-014) |
| ⑤ | `grep -c 'default = DesignSync tool push' internal/template/templates/.codex/agents/moai/manager-design.toml` | **1 불변** | 1 |
| ⑥ | `grep -c '\.agents/skills' AGENTS.md` | **1 이상** | **0** (RED) |
| ⑦ | `grep -c '\.agents/skills' internal/template/templates/AGENTS.md` | ⑥ 과 **같은 값** | 0 (RED) |
| ⑧ | `grep -rhoE 'invoke Skill\(' internal/template/templates/.codex/agents/moai/*.toml \| wc -l` | **41 불변** (발생 단위) | 41 |
| ⑨ | `grep -rhoE 'AskUserQuestion' internal/template/templates/.codex/agents/moai/*.toml \| wc -l` | **4 불변** (발생 단위 — 줄 단위로는 3) | 4 발생 / 3 줄 |

①②는 「그냥 지웠다」를 막는 짝이고, ③④는 나머지 두 지시 부류를 각각 **양성으로** 잡는다(둘 다 착수 전 0이므로 우연히 통과할 수 없다). ⑤⑧⑨는 전면 치환을 막는 불변 대조다.

### M4 — 골든 재생성과 반경 확인 (우선순위 Medium · 기계적)

- `AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/...` → 재생성.

| # | 명령 | 기대 출력 |
|---|---|---|
| ① | `go test ./internal/template/agentemit/...` (UPDATE 없이) | **PASS** (골든 드리프트 0) |
| ② | `git status --short \| awk '{print $NF}' \| grep -v '^\.moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/' \| grep -v '^\.moai/reports/t497/' \| sort` 를 `spec.md` §C.5 산출물 반경 11줄(정렬본)과 `diff` | 무출력, rc 0 (**집합 동일성**) |
| ③ | ② 의 두 필터가 걸러낸 나머지 경로 | 전부 `.moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` 또는 `.moai/reports/t497/` 접두 |

②가 반경 고정이다. 종전의 「변경 파일 수가 M2·M3 이 명시한 대상 수와 일치」는 비교할 대상 수가 어디에도 없었고, 설령 있었더라도 **개수 비교는 한 파일이 빠지고 다른 파일이 들어온 경우를 통과시킨다.** 지금은 `spec.md` §C.5 의 **문자 그대로의 11파일**과 집합을 맞추므로, 다른 카드의 `.md` 하나가 끼어들면 `diff` 가 그 줄을 낸다.

### M5 — 미러 스킬 77파일 (우선순위 Low · 후속 카드 분리로 확정)

`spec.md` §B.6 · §D. 운영자가 2026-09-07 에 **분리**로 확정했으므로(§E ②) **이 카드는 M5 를 착수하지 않고, M5 는 착지 조건에 들어가지 않는다.** 후속 카드는 리드가 발행하며, 그 카드가 상위 3개 디렉터리를 범위에서 빼려면 **표본 확인이 선행 조건**이다(`spec.md` §D · §B.6). 그 카드의 검증은 M2·M3 과 같은 형태를 77파일 모집단에 적용한다(`grep -rlE ... | wc -l` → **77** 이 사전 고정 모집단).

---

## §G. 안티패턴

- **AP-1 전면 치환.** `AskUserQuestion` / `Agent(` / `Skill(` 을 일괄 치환하면 산문 설명이 망가진다. AC 는 이 방식으로 통과할 수 없게 쓰여 있다(불변 대조 3건: 41 / 4 / 산문행 문면).
- **AP-2 `.md` 전수 grep 을 근거로 쓰기.** `Task*` 가 48 로 나온다(§B-3). TOML 본문이 유일한 모집단.
- **AP-3 단위 혼동.** 발생 84 와 줄 81 을 섞어 쓰면 판정이 어긋난다. 명령과 함께 단위를 적는다.
- **AP-4 결속표를 능력 인벤토리로 만들기.** 행은 부재를 채우는 것이지 능력 목록을 복제하는 것이 아니다(REQ-CSN-003 교리).
- **AP-5 새 중립 어휘 신설.** `reference-loader` 는 두 번째 어휘다. `tool_classes` 값 집합에서만 이름을 가져온다.
- **AP-6 자기참조 수치.** 이 SPEC 자신의 산출물을 주어로 삼는 계수(예: "이 문서의 표는 84행")는 문서를 편집하는 행위가 곧 무효화한다. 판정 시점에 같은 명령을 다시 돌린다.
- **AP-7 TOML 손편집.** 다음 골든 실행에서 되돌아간다.

---

## §H. 교차 참조

- `spec.md` §A(기준선·정정 3건) · §B(설계 판정) · §D(범위 밖)
- `acceptance.md` §D(AC 행렬)
- `.moai/reports/t497/measurement.md` — 재측정 기준선(`c9b226b22`)
- SPEC-CODEX-SKILL-NEUTRAL-001 REQ-CSN-002 / REQ-CSN-003 / §B.D7 — 어휘 단일성과 행 파생 기준의 정본
- `internal/template/agentemit/agents-codex.yaml` — `tool_classes` · `classes:` · `documented_drops:`
- `internal/template/skill_mirror.go` — `.agents/skills` 미러 도달 경로
