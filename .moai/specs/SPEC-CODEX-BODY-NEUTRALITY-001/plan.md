# SPEC-CODEX-BODY-NEUTRALITY-001 — 구현 계획

> 마일스톤은 **되돌리기 어려움 순**으로 배치했다. M1·M2 는 판단이 바뀔 여지가 가장 큰 결정(결속표 행 파생, 84건 분류)이고, M4 는 기계적 재생성이다. 시간 추정은 쓰지 않는다 — 우선순위와 단계 순서만 쓴다.

---

## §A. 맥락

`.codex/agents/moai/*.toml` 11본은 `internal/template/agentemit/writer.go` `renderTOML` 이 중립 소스 `.claude/agents/moai/*.md` 본문을 축자로 실어 만든 생성물이다. 개정은 소스에 들어가고 골든 재생성으로 TOML 에 반영된다. 기준선·정정 3건·설계 판정은 `spec.md` §A·§B.

**이 카드가 다루는 것은 본문 축 하나뿐이다.** `tools:` 프론트매터 축과 MCP/effort/sandbox 축은 매니페스트가 이미 처분했다(`spec.md` §A.4 / §D).

---

## §B. 알려진 문제

- **B-1 기준선 합계 착오.** `.moai/reports/t497/measurement.md` 의 `= 74` 는 덧셈 착오다. 파일별 값은 정확히 재현되며 합은 **84**(발생 수). 서로 다른 줄은 **81**. 이 카드의 모든 판정은 **발생 84** 를 단위로 한다.
- **B-2 결속표 3행 vs 문서 4행.** 실린 표는 3행, 완결 SPEC 기록은 4행. 미해결 — M1 산출물.
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

## §E. 미해결 질문

- **[NEEDS CLARIFICATION: 코덱스 능력 부재 측정 가능 여부]** — M1 은 코덱스 세션이 (a) `Skill("<name>")` 지시를 적재할 표면을 갖는지, (b) 위임 표면을 갖는지 실측해야 한다. 이 워크트리 세션은 중첩 `codex exec` 프로브를 돌릴 수 없을 수 있다. 측정 불가로 판정되면 결속표는 **손대지 않고**(행 0 추가) 41줄은 본문 1문장으로 덮는다 — 이 대체 처분을 승인할지 운영자 결정이 필요하다.
- **[NEEDS CLARIFICATION: M5 미러 스킬 77파일 착수 여부]** — 기본 처분은 후속 카드 분리(`spec.md` §B.6). 이 카드에서 착수할지 운영자 결정이 필요하다.

---

## §F. 마일스톤

### M1 — 능력 부재 측정과 결속표 행 파생 (우선순위 High · 되돌리기 가장 어려움)

바꾸는 판단이 가장 큰 단계다. `skill-loader` / `subagent-spawn` 이 코덱스에 **없는지**를 실측하고, 부재가 실측된 클래스에만 행을 준다. §B-2 의 3행-vs-4행 어긋남도 여기서 판정해 기록한다.

- 산출: `.moai/reports/t497/capability-absence.md` — 클래스별 `absent` / `present` / `unmeasurable` 판정 + 각 판정의 명령과 축자 출력.
- 부재가 실측된 클래스만 `AGENTS.md` 두 사본에 **`tool_classes` 값 집합의 이름 그대로** 한 행씩 추가. `reference-loader` 같은 새 어휘 금지.
- 검증 명령 · 기대 출력(사전 고정):
  - `grep -c '^| ' AGENTS.md` → **4 + (M1 이 `absent` 로 판정한 클래스 수)**. 이 셀렉터는 `^| `(파이프+공백)이라 헤더 1행과 본문 3행만 잡고 구분자 행 `|---|---|---|` 은 잡지 않는다 — 착수 전 실측 **4**.
  - `diff <(sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md) <(sed -n '/^\*\*Capability bindings/,/^---$/p' internal/template/templates/AGENTS.md)` → **무출력**(종료코드 0).
  - `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` → `PASS`, 로그의 `headroom` 이 양수.
- **행이 0개 추가되는 결과도 정당한 M1 완료다.** 그때 검증은 `N = 3` 과 `diff` 무출력이며, 판정 근거가 산출 문서에 남는다.

### M2 — 84건 전수 분류 (우선순위 High)

- 산출: `.moai/reports/t497/body-classification.md` — 84행 표, 열은 `file` · `line` · `token` · `verdict(directive|prose)` · `subject(this-agent|orchestrator|n/a)` · `rationale`.
- 판정 규칙은 `spec.md` §B.4. 주어가 오케스트레이터인 줄은 `prose` + `subject=orchestrator`.
- 검증 명령 · 기대 출력(사전 고정):
  - `grep -c '^| .*\.toml | [0-9]' .moai/reports/t497/body-classification.md` → **84**.
  - `grep -rhoE 'AskUserQuestion|TaskCreate|TaskUpdate|TaskList|TaskGet|DesignSync|Skill\(|Agent\(' internal/template/templates/.codex/agents/moai/*.toml | wc -l` → **84** (분류 대상 모집단이 변하지 않았음의 대조).
  - `grep -c 'subject=orchestrator\|orchestrator' ...` 는 쓰지 않는다 — 산문 본문에도 그 낱말이 나오므로 공허하다. 대신 열 값으로 센다.

### M3 — 중립 소스 본문 개정 (우선순위 High)

M2 가 `directive` 로 판정한 항목만 고친다. 착수 시점의 예상 대상:

| 대상 | 처분 |
|---|---|
| `manager-develop.md` 의 Task\* 지시 2줄 | `task-list` 능력 이름으로 고쳐 씀 |
| `e2e-tester.md` 의 Task\* 지시 1줄 | 같음 |
| `manager-design.md` 우선순위 사다리 | `design-sync` 부재 시 행동 1문단 추가 |
| `manager-lead.md` 자기 스폰 줄 | `subagent-spawn` 결속 참조 1건(M1 이 행을 만들었을 때만) |
| `invoke Skill(` 41줄 | **손대지 않음** |

- 검증 명령 · 기대 출력(사전 고정):
  - `grep -rhoE 'Task(Create|Update|List|Get)' internal/template/templates/.codex/agents/moai/*.toml | wc -l` → M4 재생성 후 **0**(3줄 전부 클래스 이름으로 바뀌었을 때). 값이 0 이 아니면 남은 자리를 인용해 보고한다.
  - `grep -rhoE 'invoke Skill\(' internal/template/templates/.codex/agents/moai/*.toml | wc -l` → **41 불변**(REQ-CBN-011 의 무손상 대조).
  - `grep -rhoE 'AskUserQuestion' internal/template/templates/.codex/agents/moai/*.toml | wc -l` → **4 불변**(산문 4건은 손대지 않는다는 대조).

### M4 — 골든 재생성과 반경 확인 (우선순위 Medium · 기계적)

- `AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/...` → 재생성.
- 검증 명령 · 기대 출력(사전 고정):
  - `go test ./internal/template/agentemit/...` (UPDATE 없이) → **PASS**(골든 드리프트 0).
  - `git status --short` → 변경 경로가 `{.claude/agents/moai/*.md, internal/template/templates/.claude/agents/moai/*.md, internal/template/templates/.codex/agents/moai/*.toml, AGENTS.md, internal/template/templates/AGENTS.md, .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/*, .moai/reports/t497/*}` **밖으로 나가지 않는다**. 밖의 경로가 있으면 커밋 전에 보고.
  - `git diff --stat` → 변경 파일 수가 M2·M3 이 명시한 대상 수와 일치.

### M5 — 미러 스킬 77파일 (우선순위 Low · 기본 처분 = 후속 카드 분리)

`spec.md` §B.6 · §D. 운영자가 이 카드에서 착수를 지시하지 않는 한 착지 조건에 들어가지 않는다. 착수 시 검증은 M2·M3 과 같은 형태를 77파일 모집단에 적용한다(`grep -rlE ... | wc -l` → **77** 이 사전 고정 모집단).

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
